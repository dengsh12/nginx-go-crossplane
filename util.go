/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package crossplane

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"unicode"
)

const (
	// extract single directive definition block
	// static ngx_command_t  {name}[] = {definition}
	// this regex extracts {name} and {definition}
	extractNgxDirectiveArrayRegex = "ngx_command_t\\s+(\\w+)\\[\\]\\s*=\\s*{(.*?)};"
	// extract one directive definition and attributes from extracted block
	// { ngx_string({directive_name}),
	//   {bitmask1|bitmask2|...},
	//   ... },
	// this regex extracts {directive_name} and {bitmask1|bitmask2|...}
	extractNgxSingleDirectiveRegex = "ngx_string\\(\"(.*?)\"\\).*?,(.*?),"
)

var specialBitmaskNameMatch = map[string]string{
	"HTTP":   "HTTP",
	"1MORE":  "1More",
	"2MORE":  "2More",
	"NOARGS": "NoArgs",
}

var allDirectiveContexts = map[string]interface{}{
	"ngxMainConf":       nil,
	"ngxEventConf":      nil,
	"ngxMailMainConf":   nil,
	"ngxMailSrvConf":    nil,
	"ngxStreamMainConf": nil,
	"ngxStreamSrvConf":  nil,
	"ngxStreamUpsConf":  nil,
	"ngxHTTPMainConf":   nil,
	"ngxHTTPSrvConf":    nil,
	"ngxHTTPLocConf":    nil,
	"ngxHTTPUpsConf":    nil,
	"ngxHTTPSifConf":    nil,
	"ngxHTTPLifConf":    nil,
	"ngxHTTPLmtConf":    nil,
	"ngxMgmtMainConf":   nil,
}

var directiveBlock2Context = map[string]string{
	"ngx_mgmt_block_commands": "ngxMgmtMainConf",
}

type included struct {
	directive *Directive
	err       error
}

func contains(xs []string, x string) bool {
	for _, s := range xs {
		if s == x {
			return true
		}
	}
	return false
}

func isSpace(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func isEOL(s string) bool {
	return strings.HasSuffix(s, "\n")
}

func repr(s string) string {
	q := fmt.Sprintf("%q", s)
	for _, char := range s {
		if char == '"' {
			q = strings.ReplaceAll(q, `\"`, `"`)
			q = strings.ReplaceAll(q, `'`, `\'`)
			q = `'` + q[1:len(q)-1] + `'`
			return q
		}
	}
	return q
}

func validFlag(s string) bool {
	l := strings.ToLower(s)
	return l == "on" || l == "off"
}

// validExpr ensures an expression is enclused in '(' and ')' and is not empty.
func validExpr(d *Directive) bool {
	l := len(d.Args)
	b := 0
	e := l - 1

	return l > 0 &&
		strings.HasPrefix(d.Args[b], "(") &&
		strings.HasSuffix(d.Args[e], ")") &&
		((l == 1 && len(d.Args[b]) > 2) || // empty expression single arg '()'
			(l == 2 && (len(d.Args[b]) > 1 || len(d.Args[e]) > 1)) || // empty expression two args '(', ')'
			(l > 2))
}

// prepareIfArgs removes parentheses from an `if` directive's arguments.
func prepareIfArgs(d *Directive) *Directive {
	b := 0
	e := len(d.Args) - 1
	if len(d.Args) > 0 && strings.HasPrefix(d.Args[0], "(") && strings.HasSuffix(d.Args[e], ")") {
		d.Args[0] = strings.TrimLeftFunc(strings.TrimPrefix(d.Args[0], "("), unicode.IsSpace)
		d.Args[e] = strings.TrimRightFunc(strings.TrimSuffix(d.Args[e], ")"), unicode.IsSpace)
		if len(d.Args[0]) == 0 {
			b++
		}
		if len(d.Args[e]) == 0 {
			e--
		}
		d.Args = d.Args[b : e+1]
	}
	return d
}

// combineConfigs combines config files into one by using include directives.
func combineConfigs(old *Payload) (*Payload, error) {
	if len(old.Config) < 1 {
		return old, nil
	}

	status := old.Status
	if status == "" {
		status = "ok"
	}

	errors := old.Errors
	if errors == nil {
		errors = []PayloadError{}
	}

	combined := Config{
		File:   old.Config[0].File,
		Status: "ok",
		Errors: []ConfigError{},
		Parsed: Directives{},
	}

	for _, config := range old.Config {
		combined.Errors = append(combined.Errors, config.Errors...)
		if config.Status == "failed" {
			combined.Status = "failed"
		}
	}

	for incl := range performIncludes(old, combined.File, old.Config[0].Parsed) {
		if incl.err != nil {
			return nil, incl.err
		}
		combined.Parsed = append(combined.Parsed, incl.directive)
	}

	return &Payload{
		Status: status,
		Errors: errors,
		Config: []Config{combined},
	}, nil
}

func performIncludes(old *Payload, fromfile string, block Directives) chan included {
	c := make(chan included)
	go func() {
		defer close(c)
		for _, d := range block {
			dir := *d
			if dir.IsBlock() {
				nblock := Directives{}
				for incl := range performIncludes(old, fromfile, dir.Block) {
					if incl.err != nil {
						c <- incl
						return
					}
					nblock = append(nblock, incl.directive)
				}
				dir.Block = nblock
			}
			if !dir.IsInclude() {
				c <- included{directive: &dir}
				continue
			}
			for _, idx := range dir.Includes {
				if idx >= len(old.Config) {
					c <- included{
						err: &ParseError{
							What:      fmt.Sprintf("include config with index: %d", idx),
							File:      &fromfile,
							Line:      &dir.Line,
							Statement: dir.String(),
						},
					}
					return
				}
				for incl := range performIncludes(old, old.Config[idx].File, old.Config[idx].Parsed) {
					c <- incl
				}
			}
		}
	}()
	return c
}

// extract directives from a source code through regex. Key of the return map is directive name, value of it is bitmask names
// one directive can have different bitmasks, so the value of the map is two-dimensional array
func extractDirectiveMapFromFolder(rootPath string) (map[string][][]string, error) {
	directiveMap := make(map[string][][]string, 0)
	directiveArrayExtracter := regexp.MustCompile(extractNgxDirectiveArrayRegex)
	singleDirectiveExtracter := regexp.MustCompile(extractNgxSingleDirectiveRegex)

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if the entry is a C/C++ file
		// Some dynamic modules are written in C++, like otel
		if !d.IsDir() && (strings.HasSuffix(path, ".c") || strings.HasSuffix(path, ".cpp")) {
			byteContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			strContent := string(byteContent)
			strContent = strings.ReplaceAll(strContent, "\r\n", "")
			strContent = strings.ReplaceAll(strContent, "\n", "")

			// extract directives definition code blocks, each code block contains an array of directives definition
			directiveArrays := directiveArrayExtracter.FindAllStringSubmatch(strContent, -1)
			// iterate through every code block
			for _, directiveArray := range directiveArrays {
				// the name of the directives array in source code, it may be used as the context
				directiveArrayName := directiveArray[1]
				// extract directives and their attributes in the code block, the first dimension of directiveAttributesArray
				// is index of directives, the second dimension is index of attributes
				directiveAttributesArray := singleDirectiveExtracter.FindAllStringSubmatch(directiveArray[2], -1)
				// iterate through every directive definition
				for _, directiveAttributes := range directiveAttributesArray {
					// extract attributes from the directive
					directiveName := strings.TrimSpace(directiveAttributes[1])
					diretiveBitmaskNames := strings.Split(directiveAttributes[2], "|")
					haveContext := false

					for idx, bitmaskName := range diretiveBitmaskNames {
						bitmaskGoName := ngxBitmaskName2Go(strings.TrimSpace(bitmaskName))
						diretiveBitmaskNames[idx] = bitmaskGoName
						if _, found := allDirectiveContexts[bitmaskGoName]; found {
							haveContext = true
						}
					}

					// if the directive doesn't have context in source code, maybe we still have a human-defined context for it
					// an example is directives in mgmt module, which was included in N+ R31
					if !haveContext {
						context, found := directiveBlock2Context[directiveArrayName]
						if found {
							diretiveBitmaskNames = append(diretiveBitmaskNames, context)
						}
					}

					if bitmaskNamesList, exist := directiveMap[directiveName]; exist {
						bitmaskNamesList = append(bitmaskNamesList, diretiveBitmaskNames)
						directiveMap[directiveName] = bitmaskNamesList
					} else {
						directiveMap[directiveName] = [][]string{diretiveBitmaskNames}
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(directiveMap) == 0 {
		return nil, &BasicError{
			reason: "can't find any directives in the directory and subdirectories, please check the path",
		}
	}

	return directiveMap, nil
}

// change the C style const name to Go style. An example is
// NGX_CONF_TAKE1 to ngxConfTake1
func ngxBitmaskName2Go(ngxVarName string) string {
	bitmasksNames := strings.Split(ngxVarName, "_")

	for idx, bitMaskName := range bitmasksNames {
		bitMaskName = strings.TrimSpace(bitMaskName)
		if goName, inMap := specialBitmaskNameMatch[bitMaskName]; inMap {
			bitmasksNames[idx] = goName
		} else {
			bitMaskNameRun := []rune(bitMaskName)

			for charIdx, char := range bitMaskNameRun {
				// the first charachter should be lowercase(private)
				if idx == 0 && charIdx == 0 && char >= 'A' && char <= 'Z' {
					bitMaskNameRun[charIdx] += 'a' - 'A'
				}

				// change remain part from uppercase to lowercase
				if charIdx >= 1 && char >= 'A' && char <= 'Z' {
					bitMaskNameRun[charIdx] += 'a' - 'A'
				}
			}
			bitmasksNames[idx] = string(bitMaskNameRun)
		}
	}

	return strings.Join(bitmasksNames, "")
}

func GetLineSeperator() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func GenerateSupportFileFromCode(codePath string, moduleName string, outputFilePath string) error {
	directiveMap, err := extractDirectiveMapFromFolder(codePath)
	if err != nil {
		return err
	}

	directory := filepath.Dir(outputFilePath)
	err = os.MkdirAll(directory, 0777)
	if err != nil {
		return err
	}

	file, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}

	// Annotations
	lineSeperator := GetLineSeperator()
	contents := make([]string, 0)
	contents = append(contents, "// This is generated code, don't modify it.")
	contents = append(contents, "// If you want to overwrite any directive's definition, please modify forced_directives_map.go")
	contents = append(contents, "// All the definitions are generated from the source code you provided")
	contents = append(contents, "// Each bit mask describes these behaviors:")
	contents = append(contents, "//   - how many arguments the directive can take")
	contents = append(contents, "//   - whether or not it is a block directive")
	contents = append(contents, "//   - whether this is a flag (takes one argument that's either \"on\" or \"off\")")
	contents = append(contents, "//   - which contexts it's allowed to be in")
	contents = append(contents, "")

	// package definition
	contents = append(contents, "package crossplane")
	contents = append(contents, "")

	contents = append(contents, "//nolint:gochecknoglobals")

	// make the first char in module name as uppercase, align with golang variable name conventions
	moduleNameRunes := []rune(moduleName)
	if moduleNameRunes[0] >= 'a' && moduleNameRunes[0] <= 'z' {
		moduleNameRunes[0] += 'A' - 'a'
	}
	moduleName = string(moduleNameRunes)
	mapVariableName := fmt.Sprintf("module%sDirectives", moduleName)
	contents = append(contents, fmt.Sprintf("var %s = map[string][]uint{", mapVariableName))

	// sort the directive names, just for easier search and stable output
	directiveNames := make([]string, 0, len(directiveMap))
	for name := range directiveMap {
		directiveNames = append(directiveNames, name)
	}
	sort.Strings(directiveNames)

	// generate directives map
	for _, name := range directiveNames {
		contents = append(contents, fmt.Sprintf("\t\"%s\": {", name))
		bitmaskNamesList := directiveMap[name]

		for _, bitmaskNames := range bitmaskNamesList {
			bitmaskNameNum := len(bitmaskNames)
			var builder strings.Builder
			builder.WriteString("\t\t")
			for idx, bitmaskName := range bitmaskNames {
				if idx > 0 {
					builder.WriteString("| ")
				}
				builder.WriteString(bitmaskName)
				if idx < bitmaskNameNum-1 {
					builder.WriteString(" ")
				} else {
					builder.WriteString(",")
				}
			}
			contents = append(contents, builder.String())
		}

		contents = append(contents, "\t},")
	}
	contents = append(contents, "}")

	// generate MatchFn
	contents = append(contents, "")
	contents = append(contents, fmt.Sprintf("func Match%s(directive string) ([]uint, bool) {", moduleName))
	contents = append(contents, fmt.Sprintf("\tmasks, matched := %s[directive]", mapVariableName))
	contents = append(contents, "\treturn masks, matched")
	contents = append(contents, "}")

	for _, str := range contents {
		_, err := file.WriteString(str)
		if err != nil {
			return err
		}
		_, err = file.WriteString(lineSeperator)
		if err != nil {
			return err
		}
	}
	file.Close()
	return nil
}

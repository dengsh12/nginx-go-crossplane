package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
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
	ngxMatchFnListFile             = "./ngx_matchFn_list.go"
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

// extract directives from a source code through regex. Key of the return map is directive name
// value of it is its bitmasks in string
func extractDirectiveMapFromFolder(rootPath string) (map[string][][]string, error) {
	directiveMap := make(map[string][][]string, 0)
	directivesDefineBlockExtracter := regexp.MustCompile(extractNgxDirectiveArrayRegex)
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
			directiveBlocks := directivesDefineBlockExtracter.FindAllStringSubmatch(strContent, -1)
			// iterate through every code block
			for _, directiveBlock := range directiveBlocks {
				// the name of the directives block in source code, it may be used as the context
				directiveBlockName := directiveBlock[1]
				// extract directives and their attributes in the code block, the first dimension of attributesList
				// is index of directive, the second dimension is index of attributes
				attributesList := singleDirectiveExtracter.FindAllStringSubmatch(directiveBlock[2], -1)
				// iterate through every directive
				for _, attributes := range attributesList {
					// extract attributes from the directive
					directiveName := strings.TrimSpace(attributes[1])
					diretiveBitmaskNames := strings.Split(attributes[2], "|")
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
						context, found := directiveBlock2Context[directiveBlockName]
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
		return nil, fmt.Errorf("can't find any directives in the directory and subdirectories, please check the path")
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

func getLineSeperator() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func generateSupportFileFromCode(codePath string, mapVariableName string, mathFnName string, outputFilePath string) error {
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
	lineSeperator := getLineSeperator()
	contents := make([]string, 0)
	contents = append(contents, "// Code generated by generator; DO NOT EDIT.")
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
	contents = append(contents, fmt.Sprintf("func %s(directive string) ([]uint, bool) {", mathFnName))
	contents = append(contents, fmt.Sprintf("\tmasks, matched := %s[directive]", mapVariableName))
	contents = append(contents, "\treturn masks, matched")
	contents = append(contents, "}")

	for _, line := range contents {
		_, err := file.WriteString(line)
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

func getModuleMapName(moduleName string) string {
	normalizedName := normalizeModuleName(moduleName)
	return fmt.Sprintf("module%sDirectives", normalizedName)

}

func getModuleMatchFnName(moduleName string) string {
	normalizedName := normalizeModuleName(moduleName)
	return fmt.Sprintf("Match%s", normalizedName)
}

func getModuleFileName(moduleName string) string {
	// normalizedName := normalizeModuleName(moduleName)
	return fmt.Sprintf("module_%s_directives.go", moduleName)
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func getProjectRootAbsPath() (string, error) {
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("Can't get path of generator_util.go through runtime")
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	// get the project root directory
	rootDir := filepath.Dir(absPath)
	rootDir = filepath.Dir(rootDir)

	return rootDir, nil
}

// todo: add nplus to it
func generateOssMatchFnList(matchFnList []string) error {
	contents := make([]string, 0)
	contents = append(contents, "// Code generated by generator; DO NOT EDIT.")
	contents = append(contents, "")
	contents = append(contents, "package crossplane")
	contents = append(contents, "")
	contents = append(contents, "import (")
	contents = append(contents, "\t\"reflect\"")
	contents = append(contents, ")")
	contents = append(contents, "")
	contents = append(contents, "// nginx matchFns")
	contents = append(contents, "var ngxOssMatchFuns = map[uintptr]interface{}{")
	lineSeperator := getLineSeperator()
	for _, matchFn := range matchFnList {
		contents = append(contents, fmt.Sprintf("\treflect.ValueOf(%s).Pointer(): nil,", matchFn))
	}
	contents = append(contents, "}")
	projectRootDir, err := getProjectRootAbsPath()
	if err != nil {
		return err
	}
	file, err := os.Create(path.Join(projectRootDir, ngxMatchFnListFile))
	if err != nil {
		return err
	}
	for _, line := range contents {
		_, err = file.WriteString(line + lineSeperator)
		if err != nil {
			return err
		}
	}
	file.Close()

	return nil
}

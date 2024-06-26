package generator

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	// Extract single directive definition block
	// static ngx_command_t  {name}[] = {definition}
	// this regex extracts {name} and {definition}
	extractNgxDirectiveArrayRegex = "ngx_command_t\\s+(\\w+)\\[\\]\\s*=\\s*{(.*?)};"
	// Extract one directive definition and attributes from extracted block
	// { ngx_string({directive_name}),
	//   {bitmask1|bitmask2|...},
	//   ... },
	// this regex extracts {directive_name} and {bitmask1|bitmask2|...}
	extractNgxSingleDirectiveRegex = "ngx_string\\(\"(.*?)\"\\).*?,(.*?),"
	extractSingleLineCommentRegex  = `//.*`
	extractMultiLineCommentRegex   = `/\*[\s\S]*?\*/`
	ngxMatchFnListFile             = "./ngx_matchFn_list.go"
)

// todo: delete this
var specialBitmaskNameMatch = map[string]string{
	"HTTP":   "HTTP",
	"1MORE":  "1More",
	"2MORE":  "2More",
	"NOARGS": "NoArgs",
}

var ngxBitmaskNameToGo = map[string]string{
	"NGX_MAIL_MAIN_CONF":   "ngxMailMainConf",
	"NGX_STREAM_MAIN_CONF": "ngxStreamMainConf",
	"NGX_CONF_TAKE1":       "ngxConfTake1",
	"NGX_STREAM_UPS_CONF":  "ngxStreamUpsConf",
	"NGX_HTTP_LIF_CONF":    "ngxHTTPLifConf",
	"NGX_CONF_TAKE2":       "ngxConfTake2",
	"NGX_HTTP_UPS_CONF":    "ngxHTTPUpsConf",
	"NGX_CONF_TAKE23":      "ngxConfTake23",
	"NGX_CONF_TAKE12":      "ngxConfTake12",
	"NGX_HTTP_MAIN_CONF":   "ngxHTTPMainConf",
	"NGX_HTTP_LMT_CONF":    "ngxHTTPLmtConf",
	"NGX_CONF_TAKE1234":    "ngxConfTake1234",
	"NGX_MAIL_SRV_CONF":    "ngxMailSrvConf",
	"NGX_CONF_FLAG":        "ngxConfFlag",
	"NGX_HTTP_SRV_CONF":    "ngxHTTPSrvConf",
	"NGX_CONF_1MORE":       "ngxConf1More",
	"NGX_ANY_CONF":         "ngxAnyConf",
	"NGX_CONF_TAKE123":     "ngxConfTake123",
	"NGX_MAIN_CONF":        "ngxMainConf",
	"NGX_CONF_NOARGS":      "ngxConfNoArgs",
	"NGX_CONF_2MORE":       "ngxConf2More",
	"NGX_CONF_TAKE3":       "ngxConfTake3",
	"NGX_HTTP_SIF_CONF":    "ngxHTTPSifConf",
	"NGX_EVENT_CONF":       "ngxEventConf",
	"NGX_CONF_BLOCK":       "ngxConfBlock",
	"NGX_HTTP_LOC_CONF":    "ngxHTTPLocConf",
	"NGX_STREAM_SRV_CONF":  "ngxStreamSrvConf",
	"NGX_DIRECT_CONF":      "ngxDirectConf",
	"NGX_CONF_TAKE13":      "ngxConfTake13",
	"NGX_CONF_ANY":         "ngxConfAny",
	"NGX_CONF_TAKE4":       "ngxConfTake4",
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

// Extract directives from a source code through regex. Key of the return map is directive name
// value of it is its bitmasks in string
func extractDirectiveMapFromFolder(rootPath string) (map[string][][]string, error) {
	directiveMap := make(map[string][][]string, 0)
	directivesDefineBlockExtracter := regexp.MustCompile(extractNgxDirectiveArrayRegex)
	singleDirectiveExtracter := regexp.MustCompile(extractNgxSingleDirectiveRegex)
	singleLineCommentExtracter := regexp.MustCompile(extractSingleLineCommentRegex)
	multiLineCommentExtracter := regexp.MustCompile(extractMultiLineCommentRegex)

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
			// Remove comments
			strContent = singleLineCommentExtracter.ReplaceAllString(string(strContent), "")
			strContent = multiLineCommentExtracter.ReplaceAllString(string(strContent), "")

			strContent = strings.ReplaceAll(strContent, "\r\n", "")
			strContent = strings.ReplaceAll(strContent, "\n", "")

			// Extract directives definition code blocks, each code block contains an array of directives definition
			directiveBlocks := directivesDefineBlockExtracter.FindAllStringSubmatch(strContent, -1)
			// Iterate through every code block
			for _, directiveBlock := range directiveBlocks {
				// The name of the directives block in source code, it may be used as the context
				directiveBlockName := directiveBlock[1]
				// Extract directives and their attributes in the code block, the first dimension of attributesList
				// is index of directive, the second dimension is index of attributes
				attributesList := singleDirectiveExtracter.FindAllStringSubmatch(directiveBlock[2], -1)
				// Iterate through every directive
				for _, attributes := range attributesList {
					// Extract attributes from the directive
					directiveName := strings.TrimSpace(attributes[1])
					diretiveBitmaskNames := strings.Split(attributes[2], "|")
					haveContext := false

					for idx, bitmaskName := range diretiveBitmaskNames {
						bitmaskGoName, found := ngxBitmaskNameToGo[strings.TrimSpace(bitmaskName)]
						if !found {
							return fmt.Errorf("parsing directive %s, bitmask %s in source code not found in crossplane", directiveName, bitmaskName)
						}
						diretiveBitmaskNames[idx] = bitmaskGoName
						if _, found := allDirectiveContexts[bitmaskGoName]; found {
							haveContext = true
						}
					}

					// If the directive doesn't have context in source code, maybe we still have a human-defined context for it
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

// Change the C style const name to Go style. An example is
// NGX_CONF_TAKE1 to ngxConfTake1
// todo: delete this
func ngxBitmaskName2Go(ngxVarName string) string {
	bitmasksNames := strings.Split(ngxVarName, "_")

	for idx, bitMaskName := range bitmasksNames {
		bitMaskName = strings.TrimSpace(bitMaskName)
		if goName, inMap := specialBitmaskNameMatch[bitMaskName]; inMap {
			bitmasksNames[idx] = goName
		} else {
			bitMaskNameRun := []rune(bitMaskName)

			for charIdx, char := range bitMaskNameRun {
				// The first charachter should be lowercase(private)
				if idx == 0 && charIdx == 0 && char >= 'A' && char <= 'Z' {
					bitMaskNameRun[charIdx] += 'a' - 'A'
				}

				// Change remain part from uppercase to lowercase
				if charIdx >= 1 && char >= 'A' && char <= 'Z' {
					bitMaskNameRun[charIdx] += 'a' - 'A'
				}
			}
			bitmasksNames[idx] = string(bitMaskNameRun)
		}
	}

	goName := strings.Join(bitmasksNames, "")
	if _, found := ngxBitmaskNameToGo[ngxVarName]; !found {
		ngxBitmaskNameToGo[ngxVarName] = goName
	}
	return goName
}

func getLineSeperator() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func generateSupportFileFromCode(codePath string, moduleName string, mapVariableName string, mathFnName string, outputFilePath string, filter map[string]interface{}) error {
	directiveMap, err := extractDirectiveMapFromFolder(codePath)
	if err != nil {
		return err
	}

	if filter != nil {
		directivesToDelete := []string{}
		for directve, _ := range directiveMap {
			if _, found := filter[directve]; !found {
				directivesToDelete = append(directivesToDelete, directve)
			}
		}
		for _, directive := range directivesToDelete {
			delete(directiveMap, directive)
		}
	}

	postProcFn, found := module2postProcFns[moduleName]
	if found {
		err = postProcFn(directiveMap)
		if err != nil {
			return err
		}
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
	contents = append(contents, "/**")
	contents = append(contents, " * Copyright (c) F5, Inc.")
	contents = append(contents, " *")
	contents = append(contents, " * This source code is licensed under the Apache License, Version 2.0 license found in the")
	contents = append(contents, " * LICENSE file in the root directory of this source tree.")
	contents = append(contents, " */")
	contents = append(contents, "")
	contents = append(contents, "// Code generated by generator; DO NOT EDIT.")
	contents = append(contents, "// If you want to overwrite any directive's definition, please modify priority_directives_map.go")
	contents = append(contents, "// All the definitions are generated from the source code")
	contents = append(contents, "// Each bit mask describes these behaviors:")
	contents = append(contents, "//   - how many arguments the directive can take")
	contents = append(contents, "//   - whether or not it is a block directive")
	contents = append(contents, "//   - whether this is a flag (takes one argument that's either \"on\" or \"off\")")
	contents = append(contents, "//   - which contexts it's allowed to be in")
	contents = append(contents, "")

	// Package definition
	contents = append(contents, "package crossplane")
	contents = append(contents, "")

	contents = append(contents, "//nolint:gochecknoglobals")
	contents = append(contents, fmt.Sprintf("var %s = map[string][]uint{", mapVariableName))

	// Sort the directive names, just for easier search and stable output
	directiveNames := make([]string, 0, len(directiveMap))
	for name := range directiveMap {
		directiveNames = append(directiveNames, name)
	}
	sort.Strings(directiveNames)

	// Generate directives map
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

	// Generate MatchFn
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

func uppercaseStrFirstChar(moduleName string) string {
	moduleNameRunes := []rune(moduleName)
	if moduleNameRunes[0] >= 'a' && moduleNameRunes[0] <= 'z' {
		moduleNameRunes[0] += 'A' - 'a'
	}
	return string(moduleNameRunes)
}

func lowercaseStrFirstChar(str string) string {
	strNameRunes := []rune(str)
	if strNameRunes[0] >= 'A' && strNameRunes[0] <= 'Z' {
		strNameRunes[0] += 'a' - 'A'
	}
	return string(strNameRunes)
}

func getModuleMapName(moduleName string) string {
	normalizedName := uppercaseStrFirstChar(moduleName)
	return fmt.Sprintf("module%sDirectives", normalizedName)

}

func getModuleMatchFnName(moduleName string) string {
	normalizedName := uppercaseStrFirstChar(moduleName)
	return fmt.Sprintf("%sDirectivesMatchFn", normalizedName)
}

func getModuleFileName(moduleName string) string {
	return fmt.Sprintf("analyze_%s_directives.go", moduleName)
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
		return "", fmt.Errorf("can't get path of generator_util.go through runtime")
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	// get the project root directory
	rootDir := filepath.Dir(absPath)
	rootDir = filepath.Dir(rootDir)
	rootDir = filepath.Dir(rootDir)

	return rootDir, nil
}

// todo: delete this
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

// todo: delete this
func outputMap(toOutput map[string]string) {
	for k, v := range toOutput {
		fmt.Printf("\"%s\":\"%s\",\n", k, v)
	}
}

func fetchDocumentedDirctives() (map[string]interface{}, error) {
	documentedDirectives := map[string]interface{}{}
	documentURL := "https://nginx.org/en/docs/dirindex.html"

	res, err := http.Get(documentURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Check the status code.
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document.
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// Find the #content a elements to get documented directives
	doc.Find("#content a").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		documentedDirectives[text] = nil
	})

	return documentedDirectives, nil
}

func gitClone(dir string, repoURL string, depth int) error {
	comm := exec.Command("git", "clone", "--depth", strconv.Itoa(depth), repoURL)
	comm.Dir = dir
	err := comm.Run()
	return err
}

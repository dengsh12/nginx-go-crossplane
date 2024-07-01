package generator

import (
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// todo: not sure if we need it
type source string

const (
	oSS       source = "OSS"
	lua       source = "lua"
	otel      source = "otel"
	nPlus     source = "NPlus"
	headsMore source = "headersMore"
	njs       source = "njs"
)

// regex
const (
	// Extract single directive definition block
	// static ngx_command_t  {name}[] = {definition}
	// this regex extracts {name} and {definition}
	extractNgxDirectiveArray = "ngx_command_t\\s+(\\w+)\\[\\]\\s*=\\s*{(.*?)};"
	// Extract one directive definition and attributes from extracted block
	// { ngx_string({directive_name}),
	//   {bitmask1|bitmask2|...},
	//   ... },
	// this regex extracts {directive_name} and {bitmask1|bitmask2|...}
	extractNgxSingleDirective = "ngx_string\\(\"(.*?)\"\\).*?,(.*?),"
	extractSingleLineComment  = `//.*`
	extractMultiLineComment   = `/\*[\s\S]*?\*/`
	// todo: delete it
	ngxMatchFnListFile = "./ngx_matchFn_list.go"
)

type bitDef []string

type supFileTmplStruct struct {
	SourceName        string
	Directive2bitDefs map[string][]bitDef
	MapVariableName   string
	MatchFnName       string
}

//go:embed tmpl/support_file.tmpl
var supFileTmplStr string

var supFileTmpl = template.Must(template.New("supFile").
	Funcs(template.FuncMap{"Join": strings.Join}).Parse(supFileTmplStr))

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

// Extract directives definitions from source code through regex
func getDirectiveDefFromSrc(rootPath string) (map[string][]bitDef, error) {
	directive2Defs := make(map[string][]bitDef, 0)
	directivesDefBlockExtracter := regexp.MustCompile(extractNgxDirectiveArray)
	singleDirectiveExtracter := regexp.MustCompile(extractNgxSingleDirective)
	singleLineCommentExtracter := regexp.MustCompile(extractSingleLineComment)
	multiLineCommentExtracter := regexp.MustCompile(extractMultiLineComment)

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

			// Extract directives definition code blocks, each code block contains a list of directives definition
			directiveDefBlocks := directivesDefBlockExtracter.FindAllStringSubmatch(strContent, -1)
			// Iterate through every code block
			for _, block := range directiveDefBlocks {
				// The name of the directives block in source code, it may be used as the context
				blockName := block[1]
				// Extract directives and their attributes in the code block, the first dimension of attributesList
				// is index of directive, the second dimension is index of attributes
				attributesList := singleDirectiveExtracter.FindAllStringSubmatch(block[2], -1)
				// Iterate through every directive
				for _, attributes := range attributesList {
					// Extract attributes from the directive
					directiveName := strings.TrimSpace(attributes[1])
					diretiveBitmasks := strings.Split(attributes[2], "|")
					haveContext := false

					for idx, bitmaskName := range diretiveBitmasks {
						bitmaskGoName, found := ngxBitmaskNameToGo[strings.TrimSpace(bitmaskName)]
						if !found {
							return fmt.Errorf("parsing directive %s, bitmask %s in source code not found in crossplane", directiveName, bitmaskName)
						}
						diretiveBitmasks[idx] = bitmaskGoName
						if _, found := allDirectiveContexts[bitmaskGoName]; found {
							haveContext = true
						}
					}

					// If the directive doesn't have context in source code, maybe we still have a human-defined context for it.
					// An example is directives in mgmt module, which was included in N+ R31
					if !haveContext {
						context, found := directiveBlock2Context[blockName]
						if found {
							diretiveBitmasks = append(diretiveBitmasks, context)
						}
					}

					if bitmaskDefList, exist := directive2Defs[directiveName]; exist {
						bitmaskDefList = append(bitmaskDefList, diretiveBitmasks)
						directive2Defs[directiveName] = bitmaskDefList
					} else {
						directive2Defs[directiveName] = []bitDef{diretiveBitmasks}
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(directive2Defs) == 0 {
		return nil, fmt.Errorf("can't find any directives in the directory and subdirectories, please check the path")
	}

	return directive2Defs, nil
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

func genSupFromSrcCode(codePath string, sourceName string, mapVariableName string, mathFnName string, outputFilePath string, filter map[string]interface{}) error {
	directive2BitDefs, err := getDirectiveDefFromSrc(codePath)
	if err != nil {
		return err
	}

	if filter != nil {
		directivesToDelete := []string{}
		for directve, _ := range directive2BitDefs {
			if _, found := filter[directve]; !found {
				directivesToDelete = append(directivesToDelete, directve)
			}
		}
		for _, directive := range directivesToDelete {
			delete(directive2BitDefs, directive)
		}
	}

	// For directives some sources, we have specific postprocess logics.
	postProcFn, found := source2postProcFns[sourceName]
	if found {
		err = postProcFn(directive2BitDefs)
		if err != nil {
			return err
		}
	}

	// Output the generated support file
	directory := filepath.Dir(outputFilePath)
	err = os.MkdirAll(directory, 0777)
	if err != nil {
		return err
	}

	file, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}

	err = supFileTmpl.Execute(file, supFileTmplStruct{
		SourceName:        sourceName,
		Directive2bitDefs: directive2BitDefs,
		MapVariableName:   mapVariableName,
		MatchFnName:       mathFnName,
	})
	if err != nil {
		return err
	}

	file.Close()
	return nil
}

func uppercaseStrFirstChar(str string) string {
	strRunes := []rune(str)
	if strRunes[0] >= 'a' && strRunes[0] <= 'z' {
		strRunes[0] += 'A' - 'a'
	}
	return string(strRunes)
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

// todo: delete it
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

func gitClone(dir string, repoURL string, depth int) (string, error) {
	var comm *exec.Cmd
	if depth <= 0 {
		comm = exec.Command("git", "clone", repoURL, ".")
	} else {
		comm = exec.Command("git", "clone", "--depth", strconv.Itoa(depth), repoURL, ".")
	}
	comm.Dir = dir
	output, err := comm.CombinedOutput()
	return string(output), err
}

func gitListRemoteBranch(dir string) ([]string, error) {
	lineSep := getLineSeperator()
	comm := exec.Command("git", "branch", "-r")
	comm.Dir = dir
	byteOutput, err := comm.Output()
	output := string(byteOutput)
	return strings.Split(output, lineSep), err
}

func gitCheckout(dir string, branch string) (string, error) {
	comm := exec.Command("git", "checkout", branch)
	comm.Dir = dir
	output, err := comm.CombinedOutput()
	return string(output), err
}

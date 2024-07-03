/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package generator

import (
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// regex.
const (
	// Extract single directive definition block
	// static ngx_command_t  {name}[] = {definition}
	// this regex extracts {name} and {definition}.
	extractDirectivesDefBlock = "ngx_command_t\\s+(\\w+)\\[\\]\\s*=\\s*{(.*?)};"

	// Extract one directive definition and attributes from extracted block
	// { ngx_string({directive_name}),
	//   {bitmask1|bitmask2|...},
	//   ... },
	// this regex extracts {directive_name} and {bitmask1|bitmask2|...}.
	extractSingleDirective = "ngx_string\\(\"(.*?)\"\\).*?,(.*?),"

	extractSingleLineComment = `//.*`

	extractMultiLineComment = `/\*[\s\S]*?\*/`
)

type bitDef []string

type supFileTmplStruct struct {
	Directive2bitDefs map[string][]bitDef
	MapVariableName   string
	MatchFnName       string
}

var (
	directivesDefBlockExtracter = regexp.MustCompile(extractDirectivesDefBlock)
	singleDirectiveExtracter    = regexp.MustCompile(extractSingleDirective)
	singleLineCommentExtracter  = regexp.MustCompile(extractSingleLineComment)
	multiLineCommentExtracter   = regexp.MustCompile(extractMultiLineComment)
)

//go:embed tmpl/support_file.tmpl
var supFileTmplStr string

//nolint:gochecknoglobals
var supFileTmpl = template.Must(template.New("supFile").
	Funcs(template.FuncMap{"Join": strings.Join}).Parse(supFileTmplStr))

//nolint:gochecknoglobals
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

//nolint:gochecknoglobals
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

//nolint:gochecknoglobals
var directiveBlock2Context = map[string]string{
	"ngx_mgmt_block_commands": "ngxMgmtMainConf",
}

func getDirectiveDefFromFile(path string) (map[string][]bitDef, error) {
	directive2Defs := make(map[string][]bitDef, 0)
	byteContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	strContent := string(byteContent)
	// Remove comments
	strContent = singleLineCommentExtracter.ReplaceAllString(strContent, "")
	strContent = multiLineCommentExtracter.ReplaceAllString(strContent, "")

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
					return nil, fmt.Errorf("parsing directive %s, bitmask %s in source code not found in crossplane", directiveName, bitmaskName)
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
	return directive2Defs, nil
}

// Extract directives definitions from source code through regex.
func getDirectiveDefFromSrc(srcPath string) (map[string][]bitDef, error) {
	directive2Defs := make(map[string][]bitDef, 0)

	err := filepath.WalkDir(srcPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if the entry is a C/C++ file
		// Some dynamic modules are written in C++, like otel
		if !d.IsDir() && (strings.HasSuffix(path, ".c") || strings.HasSuffix(path, ".cpp")) {
			dir2defsInFile, err := getDirectiveDefFromFile(path)
			if err != nil {
				return err
			}
			for directive, defsInFile := range dir2defsInFile {
				if _, found := directive2Defs[directive]; !found {
					directive2Defs[directive] = []bitDef{}
				}
				directive2Defs[directive] = append(directive2Defs[directive], defsInFile...)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(directive2Defs) == 0 {
		return nil, errors.New("can't find any directives in the directory and subdirectories, please check the path")
	}

	return directive2Defs, nil
}

func genSupFromSrcCode(codePath string, mapVariableName string, mathFnName string, writer io.Writer) error {
	directive2BitDefs, err := getDirectiveDefFromSrc(codePath)
	if err != nil {
		return err
	}

	err = supFileTmpl.Execute(writer, supFileTmplStruct{
		Directive2bitDefs: directive2BitDefs,
		MapVariableName:   mapVariableName,
		MatchFnName:       mathFnName,
	})
	if err != nil {
		return err
	}

	return nil
}

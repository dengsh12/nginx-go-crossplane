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

type bitDefinitions []string

type supportFileTmplStruct struct {
	Directive2Definitions map[string][]bitDefinitions
	MapVariableName       string
	MatchFnName           string
}

var (
	// Extract single directive definition block
	// static ngx_command_t  {name}[] = {definition}
	// this regex extracts {name} and {definition}.
	directivesDefBlockExtracter = regexp.MustCompile(`ngx_command_t\s+(\w+)\[\]\s*=\s*{(.*?)};`)

	// Extract one directive definition and attributes from extracted block
	// { ngx_string({directive_name}),
	//   {bitmask1|bitmask2|...},
	//   ... },
	// this regex extracts {directive_name} and {bitmask1|bitmask2|...}.
	singleDirectiveExtracter = regexp.MustCompile(`ngx_string\("(.*?)"\).*?,(.*?),`)

	singleLineCommentExtracter = regexp.MustCompile(`//.*`)

	multiLineCommentExtracter = regexp.MustCompile(`/\*[\s\S]*?\*/`)
)

// Template of support file. A support file contains a map from
// diective to its bitmask definitions, and a MatchFunc for it.
//
//go:embed tmpl/support_file.tmpl
var supportFileTmplStr string

//nolint:gochecknoglobals
var supportFileTmpl = template.Must(template.New("supFile").
	Funcs(template.FuncMap{"Join": strings.Join}).Parse(supportFileTmplStr))

//nolint:gochecknoglobals
var ngxBitmaskToGo = map[string]string{
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
var allNgxContexts = map[string]struct{}{
	"ngxMainConf":       {},
	"ngxEventConf":      {},
	"ngxMailMainConf":   {},
	"ngxMailSrvConf":    {},
	"ngxStreamMainConf": {},
	"ngxStreamSrvConf":  {},
	"ngxStreamUpsConf":  {},
	"ngxHTTPMainConf":   {},
	"ngxHTTPSrvConf":    {},
	"ngxHTTPLocConf":    {},
	"ngxHTTPUpsConf":    {},
	"ngxHTTPSifConf":    {},
	"ngxHTTPLifConf":    {},
	"ngxHTTPLmtConf":    {},
	"ngxMgmtMainConf":   {},
}

//nolint:gochecknoglobals
var directiveBlock2Context = map[string]string{
	"ngx_mgmt_block_commands": "ngxMgmtMainConf",
}

//nolint:nonamedreturns
func getDirectiveFromFile(path string) (directive2Definitions map[string][]bitDefinitions, err error) {
	directive2Definitions = make(map[string][]bitDefinitions, 0)
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
			directiveBitmasks := strings.Split(attributes[2], "|")
			haveContext := false

			// transfer C-style diretiveBitmasks to go style
			for idx, bitmask := range directiveBitmasks {
				bitmaskGoName, found := ngxBitmaskToGo[strings.TrimSpace(bitmask)]
				if !found {
					return nil, fmt.Errorf("parsing directive %s, bitmask %s in source code not found in crossplane", directiveName, bitmask)
				}
				directiveBitmasks[idx] = bitmaskGoName
				if _, found := allNgxContexts[bitmaskGoName]; found {
					haveContext = true
				}
			}

			// If the directive doesn't have context in source code, maybe we still have a human-defined context for it.
			// An example is directives in mgmt module, which was included in N+ R31, we add ngxMgmtMainConf for it
			if !haveContext {
				context, found := directiveBlock2Context[blockName]
				if found {
					directiveBitmasks = append(directiveBitmasks, context)
				}
			}

			directive2Definitions[directiveName] = append(directive2Definitions[directiveName], directiveBitmasks)
		}
	}
	return directive2Definitions, nil
}

//nolint:nonamedreturns
func getDirectivesFromFolder(path string) (directive2Definitions map[string][]bitDefinitions, err error) {
	directive2Definitions = make(map[string][]bitDefinitions, 0)

	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if the entry is a C/C++ file
		// Some dynamic modules are written in C++, like otel
		if d.IsDir() {
			return nil
		}

		if !(strings.HasSuffix(path, ".c") || strings.HasSuffix(path, ".cpp")) {
			return nil
		}
		dir2defsInFile, err := getDirectiveFromFile(path)
		if err != nil {
			return err
		}
		for directive, defsInFile := range dir2defsInFile {
			directive2Definitions[directive] = append(directive2Definitions[directive], defsInFile...)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(directive2Definitions) == 0 {
		return nil, errors.New("can't find any directives in the directory and subdirectories, please check the path")
	}

	return directive2Definitions, nil
}

func genFromSrcCode(codePath string, mapVariableName string, matchFnName string, writer io.Writer) error {
	directive2Definitions, err := getDirectivesFromFolder(codePath)
	if err != nil {
		return err
	}

	err = supportFileTmpl.Execute(writer, supportFileTmplStruct{
		Directive2Definitions: directive2Definitions,
		MapVariableName:       mapVariableName,
		MatchFnName:           matchFnName,
	})
	if err != nil {
		return err
	}

	return nil
}

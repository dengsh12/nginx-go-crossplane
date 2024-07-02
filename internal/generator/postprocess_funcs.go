/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package generator

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

var postProcessFuncs = []func(map[string][]bitDef) error{
	postProcLua,
	postProcNgxNative,
}

var argsNumBitmasks = []string{
	"ngxConfNoArgs",
	"ngxConfTake1",
	"ngxConfTake2",
	"ngxConfTake3",
	"ngxConfTake4",
	"ngxConfTake5",
	"ngxConfTake6",
	"ngxConfTake7",
}

// For lua module, we remove ngxConfBlock and add args num by 1
// See PR: https://github.com/nginxinc/nginx-go-crossplane/pull/86
func postProcLua(directives2Defs map[string][]bitDef) error {
	luaExclusive := []string{
		"access_by_lua",
		"lua_shared_dict",
		"init_by_lua",
	}
	// Judge if it is lua module, if not we don't need to process it
	if !containDirectives(directives2Defs, luaExclusive) {
		return nil
	}

	for directive, bitmaskNamesList := range directives2Defs {
		for dirIdx, bitmaskNames := range bitmaskNamesList {
			isBlock := false
			blockBitmaskIdx := 0
			for bitmaskNameIdx, bitmaskName := range bitmaskNames {
				if bitmaskName == "ngxConfBlock" {
					isBlock = true
					blockBitmaskIdx = bitmaskNameIdx
				}
			}

			if isBlock {
				bitmaskNamesList[dirIdx] = append(bitmaskNames[:blockBitmaskIdx], bitmaskNames[blockBitmaskIdx+1:]...)
				for idx, dirBitmaskName := range bitmaskNames {
					for argsNum, argsBitmaskName := range argsNumBitmasks {
						if dirBitmaskName == argsBitmaskName {
							if argsNum > len(argsNumBitmasks) {
								return fmt.Errorf("too many arguments for lua block directive %s", directive)
							}
							bitmaskNames[idx] = argsNumBitmasks[argsNum+1]
						}
					}
				}
			}
		}
	}
	return nil
}

// Postprocess for OSS and NPlus
func postProcNgxNative(directives2Defs map[string][]bitDef) error {
	ngxNativeExclusive := []string{
		"proxy_pass",
		"location",
		"index",
	}
	// If it is not OSS or NPlus, directly return
	if !containDirectives(directives2Defs, ngxNativeExclusive) {
		return nil
	}

	// We have a human defined bitmask 'ngxConfBlock' for 'if' directive, which
	// is not in OSS/NPlus source code. We append it here.
	for directive := range directives2Defs {
		if directive == "if" {
			defs := directives2Defs[directive]
			for idx, def := range defs {
				defs[idx] = append(def, "ngxConfExpr")
			}
		}
	}

	// For OSS and NPlus source code, there are many undocumented directives.
	// We don't provide supports for them
	// Supported list here: https://nginx.org/en/docs/dirindex.html
	documentedDirs, err := fetchNgxDocumentedDirctives()
	if err != nil {
		return err
	}

	for dir := range directives2Defs {
		dirToDelete := []string{}
		if _, found := documentedDirs[dir]; !found {
			dirToDelete = append(dirToDelete, dir)
		}
		for _, dir := range dirToDelete {
			delete(directives2Defs, dir)
		}
	}

	return nil
}

func fetchNgxDocumentedDirctives() (map[string]interface{}, error) {
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

func containDirectives(dirSet map[string][]bitDef, diretives []string) bool {
	for _, dir := range diretives {
		if _, found := dirSet[dir]; !found {
			return false
		}
	}
	return true
}

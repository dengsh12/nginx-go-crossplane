package generator

import "fmt"

var source2postProcFns = map[string]func(map[string][]bitDef) error{
	"lua":   postProcLuaMap,
	"OSS":   postProcNgxNativeMap,
	"NPlus": postProcNgxNativeMap,
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
func postProcLuaMap(directivesMap map[string][]bitDef) error {
	for directive, bitmaskNamesList := range directivesMap {
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

func postProcNgxNativeMap(directivesMap map[string][]bitDef) error {
	for directive, _ := range directivesMap {
		if directive == "if" {
			directivesMap[directive] = []bitDef{
				{"ngxHTTPSrvConf", "ngxHTTPLocConf", "ngxConfBlock", "ngxConfExpr", "ngxConf1More"},
			}
		}
	}
	return nil
}

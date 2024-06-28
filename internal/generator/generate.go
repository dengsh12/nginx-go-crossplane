package generator

import (
	_ "embed"
	"fmt"
	"path"
	"sort"
)

const (
	ossRepo          = "https://github.com/nginx/nginx.git"
	ossGenerateLimit = 3
)

// todo: not sure if we need it
var sourceStr2Enum = map[string]source{
	"OSS":         oSS,
	"lua":         lua,
	"headersMore": headsMore,
	"njs":         njs,
	"otel":        otel,
}

// todo: delete it
func compare2directiveMap(correct map[string][]uint, generated map[string][]uint) {
	for directiveName, bitmask := range correct {
		mBitMask, find := generated[directiveName]
		if !find {
			fmt.Println(directiveName + " no found")
			continue
		}
		sort.Slice(mBitMask, func(i, j int) bool {
			return mBitMask[i] > mBitMask[j]
		})
		sort.Slice(bitmask, func(i, j int) bool {
			return bitmask[i] > bitmask[j]
		})
		if len(bitmask) != len(mBitMask) {
			fmt.Println(directiveName + " no same len")
			continue
		}
		sameV := true
		humanV := false
		for idx, v := range mBitMask {
			if v != bitmask[idx] {
				sameV = false
			}
			if bitmask[idx]&uint(0x00020000) != 0 {
				humanV = true
			}
		}
		if !sameV && !humanV {
			fmt.Println(directiveName + " no same v")
		}
	}
}

// todo: delete it
func compare2directiveMapWithMatchFn(correct map[string][]uint, matchFn func(directive string) (masks []uint, matched bool)) {
	for directiveName, bitmask := range correct {
		mBitMask, find := matchFn(directiveName)
		if !find {
			fmt.Println(directiveName + " no found")
			continue
		}
		sort.Slice(mBitMask, func(i, j int) bool {
			return mBitMask[i] > mBitMask[j]
		})
		sort.Slice(bitmask, func(i, j int) bool {
			return bitmask[i] > bitmask[j]
		})
		if len(bitmask) != len(mBitMask) {
			fmt.Println(directiveName + " no same len")
			continue
		}
		sameV := true
		humanV := false
		for idx, v := range mBitMask {
			if v != bitmask[idx] {
				sameV = false
			}
			if bitmask[idx]&uint(0x00020000) != 0 {
				humanV = true
			}
		}
		if !sameV && !humanV {
			fmt.Println(directiveName + " no same v")
		}
	}
}

func TestRun() {

	// compare2directiveMapWithMatchFn(crossplane.AppProtectWAFv5Directives, crossplane.AppProtectWAFv5DirectivesMatchFn)
}

func Generate(function string, sourceName string, onlyDocumentedDirs bool, sourceCodePath string, outputFolder string) {
	validFunctions := []string{"code2map", "code2json", "json2map", "generate"}
	isValidFunc := false
	for _, funcName := range validFunctions {
		if function == funcName {
			isValidFunc = true
		}
	}

	if !isValidFunc {
		fmt.Println("func should be code2map, code2json, or json2map")
		return
	}

	if function == "" {
		fmt.Println("Please provide the function you need, -h or --help for help")
		return
	}

	if function == "generate" {
		generator, found := source2generator[sourceName]
		if !found {
			fmt.Printf("source %s not found, please ensure there is a generator for it", sourceName)
			return
		}
		fmt.Printf("generating for %s...", sourceName)
		fmt.Println()
		err := generator.generateFromWeb()
		if err != nil {
			fmt.Printf("generation for %s failed, reason: %s", sourceName, err.Error())
		} else {
			fmt.Printf("generation for %s success", sourceName)
		}
		fmt.Println()
	} else if function == "code2map" {
		if sourceCodePath == "" {
			fmt.Println("Please provide the path of the source code folder, -h or --help for help")
			return
		}
		if sourceName == "" {
			fmt.Println("Please provide the module name, -h or --help for help")
			return
		}
		var filter map[string]interface{}
		var err error
		if onlyDocumentedDirs {
			filter, err = fetchDocumentedDirctives()
			if err != nil {
				fmt.Println(err)
			}
		}
		err = generateSupportFileFromCode(sourceCodePath, sourceName, getModuleMapName(sourceName), getModuleMatchFnName(sourceName), path.Join(outputFolder, getModuleFileName(sourceName)), filter)
		if err != nil {
			fmt.Println("Generate failed, error:")
			fmt.Println(err.Error())
		}
	}
}

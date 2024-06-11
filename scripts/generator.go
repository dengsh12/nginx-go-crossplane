package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	crossplane "github.com/nginxinc/nginx-go-crossplane"
)

// var headersMoreDirectives = map[string][]uint{
// 	"more_set_headers": {
// 		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConf1More,
// 	},
// 	"more_clear_headers": {
// 		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConf1More,
// 	},
// 	"more_set_input_headers": {
// 		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConf1More,
// 	},
// 	"more_clear_input_headers": {
// 		ngxHTTPMainConf | ngxHTTPSrvConf | ngxHTTPLocConf | ngxHTTPLifConf | ngxConf1More,
// 	},
// }
// func MatchHeadersMore(directive string) ([]uint, bool) {
// 	masks, matched := headersMoreDirectives[directive]
// 	return masks, matched
// }

func compare2directiveMap(correct map[string][]uint, generated map[string][]uint) {
	for directiveName, bitmask := range correct {
		if !strings.HasPrefix(directiveName, "js") {
			continue
		}
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

func testRun() {
	// n := len(os.Args)
	for _, str := range os.Args {
		fmt.Println(str)
	}
}

func main() {
	var (
		function       = flag.String("func", "", "the function you need, should be code2map, code2json, or json2map (required)")
		sourceCodePath = flag.String("source_code", "", "the folder includes the source code your want to generate support from (required when func=code2map or code2json)")
		_              = flag.String("json_file", "", "the folder of the json file you want to generate support from (required when func=json2map)")
		moduleName     = flag.String("module_name", "", "the name of the module/OSS version/N+ version (required)")
		outputFolder   = flag.String("output_folder", "./tmp", "the folder at which the generated support file locates, ./tmp by default(optional)")
	)
	validFunctions := []string{"code2map", "code2json", "json2map"}
	isValidFunc := false
	for _, funcName := range validFunctions {
		if *function == funcName {
			isValidFunc = true
		}
	}
	if !isValidFunc {
		fmt.Println("func should be code2map, code2json, or json2map")
		return
	}
	flag.Parse()
	if *function == "" {
		fmt.Println("Please provide the function you need, -h or --help for help")
		return
	}
	// fmt.Println(*moduleName)
	// testRun()
	if *function == "code2map" {
		if *sourceCodePath == "" {
			fmt.Println("Please provide the path of the source code folder, -h or --help for help")
			return
		}
		if *moduleName == "" {
			fmt.Println("Please provide the module name, -h or --help for help")
			return
		}
		err := crossplane.GenerateSupportFileFromCode(*sourceCodePath, *moduleName, fmt.Sprintf("%s/module_%s_directives.go", *outputFolder, *moduleName))
		if err != nil {
			fmt.Println("Generate failed, error:")
			fmt.Println(err.Error())
		}
	}
	// directiveMap, _ := crossplane.ExtractDirectiveMapFromFolder(path)

	// compare2directiveMap(crossplane.CrossplaneDirectives, crossplane.ModuleNjsDirectives)

	// for key, value := range directiveMap {
	// 	fmt.Print(key)
	// 	fmt.Print(":")
	// 	fmt.Println()
	// 	for _, bitmasks := range value {
	// 		for _, bitmask := range bitmasks {
	// 			fmt.Print(bitmask)
	// 			fmt.Print("|")
	// 		}
	// 		fmt.Println()
	// 	}
	// }
	// payload, err := crossplane.Parse(path, &crossplane.ParseOptions{})
	// if err != nil {
	// 	panic(err)
	// }

	// b, err := json.Marshal(payload)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(string(b))
}

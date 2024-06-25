//go:generate go run generator.go generator_util.go postprocess_funcs.go generate_funcs.go --func=generate --module_name=lua
//go:generate go run generator.go generator_util.go postprocess_funcs.go generate_funcs.go --func=generate --module_name=headersMore
//go:generate go run generator.go generator_util.go postprocess_funcs.go generate_funcs.go  --func=generate --module_name=njs
//go:generate go run generator.go generator_util.go postprocess_funcs.go generate_funcs.go --func=generate --module_name=otel
//go:generate go run generator.go generator_util.go postprocess_funcs.go generate_funcs.go --func=generate --module_name=OSS

package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
)

const (
	ossRepo         = "https://github.com/nginx/nginx.git"
	ossVersionLimit = 3
	tmpRootDir      = "./generator_tmp"
)

const (
	luaModuleName          = "lua"
	heardersMoreModuleName = "headersMore"
	njsModuleName          = "njs"
	otelModuleName         = "otel"
	ossName                = "OSS"
	nPlusName              = "NPLUS"
)

var module2git = map[string]string{
	heardersMoreModuleName: "https://github.com/openresty/headers-more-nginx-module.git",
	luaModuleName:          "https://github.com/openresty/lua-nginx-module.git",
	njsModuleName:          "https://github.com/nginx/njs.git",
	otelModuleName:         "https://github.com/nginxinc/nginx-otel.git",
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

// todo: delete it
func testRun() {
	// compare2directiveMapWithMatchFn(crossplane.AppProtectWAFv5Directives, crossplane.AppProtectWAFv5DirectivesMatchFn)
}

func generateModuleFromWeb(moduleName string) error {
	repoURL, found := module2git[moduleName]
	if !found {
		return fmt.Errorf("can't find git repo for module {%s}, make sure it is in the module2git map (in ./scripts/generator_script.go)", moduleName)
	}

	moduleTmpDir := path.Join(tmpRootDir, moduleName)
	if directoryExists(moduleTmpDir) {
		err := os.RemoveAll(moduleTmpDir)
		if err != nil {
			return fmt.Errorf("removing %s failed, please remove this directory mannually", moduleTmpDir)
		}
	}
	defer os.RemoveAll(tmpRootDir)

	err := os.MkdirAll(moduleTmpDir, 0777)
	if err != nil {
		return err
	}
	// Clone the repository
	_, err = git.PlainClone(moduleTmpDir, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: nil,
		Depth:    1,
	})

	if err != nil {
		return err
	}

	projectRoot, err := getProjectRootAbsPath()
	if err != nil {
		return err
	}

	err = generateSupportFileFromCode(moduleTmpDir, moduleName, getModuleMapName(moduleName), getModuleMatchFnName(moduleName), path.Join(projectRoot, getModuleFileName(moduleName)), nil)
	if err != nil {
		return err
	}

	return nil
}

func generate(moduleName string) error {
	genFunc, found := module2genFunc[moduleName]
	if found {
		return genFunc()
	}
	return generateModuleFromWeb(moduleName)
}

func main() {
	// testRun()
	// return
	start_t := time.Now()
	var (
		function           = flag.String("func", "", "the function you need, should be code2map, code2json, generate, or json2map (required)")
		sourceCodePath     = flag.String("source_code", "", "the folder includes the source code your want to generate support from (required when func=code2map or code2json)")
		_                  = flag.String("json_file", "", "the folder of the json file you want to generate support from (required when func=json2map)")
		moduleName         = flag.String("module_name", "", "OSS, NPLUS, or the name of the module(required)")
		outputFolder       = flag.String("output_folder", "./tmp", "the folder at which the generated support file locates, ./tmp by default(optional)")
		onlyDocumentedDirs = flag.Bool("documented_only", false, "only output consider directives on https://nginx.org/en/docs/dirindex.html, optional, false by default")
	)
	flag.Parse()
	validFunctions := []string{"code2map", "code2json", "json2map", "generate"}
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

	if *function == "" {
		fmt.Println("Please provide the function you need, -h or --help for help")
		return
	}
	if *function == "generate" {
		fmt.Printf("generating for %s...", *moduleName)
		fmt.Println()
		err := generate(*moduleName)
		if err != nil {
			fmt.Printf("generation for %s failed, reason: %s", *moduleName, err.Error())
		} else {
			fmt.Printf("generation for %s success", *moduleName)
		}
		fmt.Println()
	} else if *function == "code2map" {
		if *sourceCodePath == "" {
			fmt.Println("Please provide the path of the source code folder, -h or --help for help")
			return
		}
		if *moduleName == "" {
			fmt.Println("Please provide the module name, -h or --help for help")
			return
		}
		var filter map[string]interface{}
		var err error
		if *onlyDocumentedDirs {
			filter, err = fetchDocumentedDirctives()
			if err != nil {
				fmt.Println(err)
			}
		}
		err = generateSupportFileFromCode(*sourceCodePath, *moduleName, getModuleMapName(*moduleName), getModuleMatchFnName(*moduleName), path.Join(*outputFolder, getModuleFileName(*moduleName)), filter)
		if err != nil {
			fmt.Println("Generate failed, error:")
			fmt.Println(err.Error())
		}
	}

	fmt.Println("use time:" + time.Since(start_t).String())
}

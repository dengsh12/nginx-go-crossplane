package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	crossplane "github.com/nginxinc/nginx-go-crossplane"
)

const (
	ossRepo         = "https://github.com/nginx/nginx.git"
	ossVersionLimit = 6
	tmpRootDir      = "./generator_tmp"
)

var module2git = map[string]string{
	"headersMore": "https://github.com/openresty/headers-more-nginx-module.git",
	"lua":         "https://github.com/openresty/lua-nginx-module.git",
	"njs":         "https://github.com/nginx/njs.git",
	"otel":        "https://github.com/nginxinc/nginx-otel.git",
}

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

func generateOSS() error {
	ossTmpDir := path.Join(tmpRootDir, "OSS")
	if directoryExists(ossTmpDir) {
		err := os.RemoveAll(ossTmpDir)
		if err != nil {
			return &crossplane.BasicError{
				Reason: fmt.Sprintf("Removing %s failed, please remove this directory mannually", ossTmpDir),
			}
		}
	}
	defer os.RemoveAll(tmpRootDir)

	repo, err := git.PlainClone(ossTmpDir, false, &git.CloneOptions{
		URL: ossRepo,
	})

	if err != nil {
		return err
	}

	// Fetch all remote branches
	// err = repo.Fetch(&git.FetchOptions{
	// 	RemoteName: "origin",
	// 	RefSpecs:   []config.RefSpec{"refs/heads/*:refs/remotes/origin/*"},
	// })
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	// List all references and filter branches
	refs, err := repo.References()
	if err != nil {
		return err
	}

	// Iterate through all branches
	allBranches := make([]string, 0)
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsRemote() && strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/") {
			branchName := ref.Name().Short()
			fmt.Printf("Checking branch: %s\n", branchName)
			// get "master" and "default" out here
			if strings.Contains(branchName, "-") {
				allBranches = append(allBranches, branchName)
			}
			if branchName == "origin/branches/stable-1.8" {

				worktree.Checkout(&git.CheckoutOptions{
					Branch: ref.Name(),
				})
				return nil
			}

		}
		return nil
	})
	if err != nil {
		return err
	}

	sort.Slice(allBranches, func(i, j int) bool {
		iVersionStr := strings.Split(allBranches[i], "-")[1]
		jVersionStr := strings.Split(allBranches[j], "-")[1]

		return iVersion > jVersion
	})
	// only supports several latest stable version
	wantedBranches := make(map[string]interface{}, 0)
	wantedBranches["master"] = nil
	for idx, branch := range allBranches {
		if idx >= ossVersionLimit-1 {
			break
		}
		wantedBranches[branch] = nil
		fmt.Println(branch)
	}

	return nil
}

func normalizeModuleName(moduleName string) string {
	// make the first char in module name as uppercase, align with golang variable name conventions
	moduleNameRunes := []rune(moduleName)
	if moduleNameRunes[0] >= 'a' && moduleNameRunes[0] <= 'z' {
		moduleNameRunes[0] += 'A' - 'a'
	}
	return string(moduleNameRunes)
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

func generateModuleFromWeb(moduleName string) error {
	repoURL, found := module2git[moduleName]
	if !found {
		return &crossplane.BasicError{
			Reason: fmt.Sprintf("can't find git repo for module {%s}, make sure it is in the module2git map (in ./scripts/generator_script.go)", moduleName),
		}
	}

	moduleTmpDir := path.Join(tmpRootDir, moduleName)
	if directoryExists(moduleTmpDir) {
		err := os.RemoveAll(moduleTmpDir)
		if err != nil {
			return &crossplane.BasicError{
				Reason: fmt.Sprintf("Removing %s failed, please remove this directory mannually", moduleTmpDir),
			}
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
	})

	if err != nil {
		return err
	}

	err = generateSupportFileFromCode(moduleTmpDir, getModuleMapName(moduleName), getModuleMatchFnName(moduleName), getModuleFileName(moduleName))
	if err != nil {
		return err
	}

	return nil
}

func generateFromWeb(moduleName string) error {
	if moduleName == "OSS" {
		return generateOSS()
	} else {
		return generateModuleFromWeb(moduleName)
	}
}

func main() {
	start_t := time.Now()
	var (
		function       = flag.String("func", "", "the function you need, should be code2map, code2json, generate, or json2map (required)")
		sourceCodePath = flag.String("source_code", "", "the folder includes the source code your want to generate support from (required when func=code2map or code2json)")
		_              = flag.String("json_file", "", "the folder of the json file you want to generate support from (required when func=json2map)")
		moduleName     = flag.String("module_name", "", "OSS, NPLUS, or the name of the module(required)")
		outputFolder   = flag.String("output_folder", "./tmp", "the folder at which the generated support file locates, ./tmp by default(optional)")
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
		err := generateFromWeb(*moduleName)
		if err != nil {
			fmt.Printf("generation for %s failed, reason: %s", *moduleName, err.Error())
		} else {
			fmt.Printf("generation for %s success, file:%s", *moduleName, getModuleFileName(*moduleName))
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
		err := generateSupportFileFromCode(*sourceCodePath, getModuleMapName(*moduleName), getModuleMatchFnName(*moduleName), path.Join(*outputFolder, getModuleFileName(*moduleName)))
		if err != nil {
			fmt.Println("Generate failed, error:")
			fmt.Println(err.Error())
		}
	}

	fmt.Println("use time:" + time.Since(start_t).String())
	// fmt.Println(*moduleName)
	// testRun()

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

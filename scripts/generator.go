//go:generate go run generator.go generator_util.go postprocess_funcs.go --func=generate --module_name=lua
//go:generate go run generator.go generator_util.go postprocess_funcs.go --func=generate --module_name=headersMore
//go:generate go run generator.go generator_util.go postprocess_funcs.go  --func=generate --module_name=njs
//go:generate go run generator.go generator_util.go postprocess_funcs.go  --func=generate --module_name=otel
//go:generate go run generator.go generator_util.go postprocess_funcs.go  --func=generate --module_name=OSS

package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
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
func testRun() {
	documentedDirectives, err := fetchDocumentedDirctives()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(documentedDirectives)
}

func generateOSS() error {
	ossTmpDir := path.Join(tmpRootDir, ossName)
	if directoryExists(ossTmpDir) {
		err := os.RemoveAll(ossTmpDir)
		if err != nil {
			return fmt.Errorf("removing %s failed, please remove this directory mannually", ossTmpDir)
		}
	}
	os.MkdirAll(ossTmpDir, 0777)
	defer os.RemoveAll(tmpRootDir)

	repo, err := git.PlainClone(ossTmpDir, false, &git.CloneOptions{
		URL:   ossRepo,
		Depth: 1,
	})

	if err != nil {
		return err
	}

	// Fetch all remote branches
	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/*:refs/remotes/origin/*"},
	})
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

	// Find all branches
	allBranches := make([]string, 0)
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsRemote() && strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/") {
			branchName := ref.Name().Short()
			// Kick "master" and "default" out
			if strings.Contains(branchName, "-") {
				allBranches = append(allBranches, branchName)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Only supports several latest stable version
	sort.Slice(allBranches, func(i, j int) bool {
		iVersionStr := strings.Split(allBranches[i], "-")[1]
		jVersionStr := strings.Split(allBranches[j], "-")[1]
		iVerSplit := strings.Split(iVersionStr, ".")
		jVerSplit := strings.Split(jVersionStr, ".")
		iVerIntPart, _ := strconv.Atoi(iVerSplit[0])
		jVerIntPart, _ := strconv.Atoi(jVerSplit[0])
		iVerDecimalPart, _ := strconv.Atoi(iVerSplit[1])
		jVerDecimalPart, _ := strconv.Atoi(jVerSplit[1])
		if iVerIntPart == jVerIntPart {
			return iVerDecimalPart > jVerDecimalPart
		}
		return iVerIntPart > jVerIntPart
	})
	wantedBranches := make(map[string]interface{}, 0)
	wantedBranches["origin/master"] = nil
	for idx, branch := range allBranches {
		if idx >= ossVersionLimit-1 {
			break
		}
		wantedBranches[branch] = nil
	}

	// Generate support files
	refs, err = repo.References()
	if err != nil {
		return err
	}
	matchFnList := make([]string, 0)
	projectRoot, err := getProjectRootAbsPath()
	if err != nil {
		return err
	}
	filter, err := fetchDocumentedDirctives()
	if err != nil {
		return err
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsRemote() && strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/") {
			branchName := ref.Name().Short()
			if _, found := wantedBranches[branchName]; found {
				err := worktree.Checkout(&git.CheckoutOptions{
					Branch: ref.Name(),
				})
				if err != nil {
					return err
				}
				ossVerStr := ""
				if strings.Contains(branchName, "master") {
					ossVerStr = "Latest"
				} else {
					ossVerStr = strings.Split(branchName, "-")[1]
					ossVerStr = strings.Join(strings.Split(ossVerStr, "."), "")
				}
				matchFnName := fmt.Sprintf("Oss%sDirectivesMatchFn", ossVerStr)
				fileName := fmt.Sprintf("./ngx_oss_%s_directives.go", lowercaseStrFirstChar(ossVerStr))
				generateSupportFileFromCode(ossTmpDir, ossName, fmt.Sprintf("ngxOss%sDirectives", ossVerStr), matchFnName, path.Join(projectRoot, fileName), filter)
				matchFnList = append(matchFnList, matchFnName)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
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

func generateFromWeb(moduleName string) error {
	if moduleName == "OSS" {
		return generateOSS()
	} else {
		return generateModuleFromWeb(moduleName)
	}
}

func main() {
	// testRun()
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
		err := generateFromWeb(*moduleName)
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

package generator

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	tmpDirPattern = "generator_tmp_"
)

type codeGenerator interface {
	generateFromWeb() error
}

type normalGenerator struct {
	sourceName string
	repoURL    string
}

var source2generator = map[string]codeGenerator{
	"lua": &normalGenerator{
		sourceName: "lua",
		repoURL:    "https://github.com/openresty/lua-nginx-module.git",
	},
	"headersMore": &normalGenerator{
		sourceName: "headersMore",
		repoURL:    "https://github.com/openresty/headers-more-nginx-module.git",
	},
	"njs": &normalGenerator{
		sourceName: "njs",
		repoURL:    "https://github.com/nginx/njs.git",
	},
	"otel": &normalGenerator{
		sourceName: "otel",
		repoURL:    "https://github.com/nginxinc/nginx-otel.git",
	},
	"OSS": &ossGenerator{
		repoURL: "https://github.com/nginx/nginx.git",
	},
}

func (generator *normalGenerator) generateFromWeb() error {
	sourceName := generator.sourceName
	repoURL := generator.repoURL

	tmpDir, err := os.MkdirTemp("", tmpDirPattern)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Clone the repository
	cmdOutput, err := gitClone(tmpDir, repoURL, 1)
	if err != nil {
		fmt.Println("git clone fail, cmd output:" + cmdOutput)
		return err
	}

	projectRoot, err := getProjectRootAbsPath()
	if err != nil {
		return err
	}

	err = generateSupportFileFromCode(tmpDir, sourceName, getModuleMapName(sourceName), getModuleMatchFnName(sourceName), path.Join(projectRoot, getModuleFileName(sourceName)), nil)
	if err != nil {
		return err
	}

	return nil
}

type ossGenerator struct {
	repoURL string
}

func (generator *ossGenerator) generateFromWeb() error {
	repoURL := generator.repoURL

	tmpDir, err := os.MkdirTemp("", tmpDirPattern)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	cmdOutput, err := gitClone(tmpDir, repoURL, 0)
	if err != nil {
		fmt.Println("git clone failed, cmd output: " + cmdOutput)
		return err
	}

	branches, err := gitListRemoteBranch(tmpDir)
	if err != nil {
		fmt.Println("git branch -r failed, cmd output: " + branches[0])
		return err
	}

	stableBranches := make([]string, 0)
	for _, branch := range branches {
		if strings.Contains(branch, "stable-") {
			stableBranches = append(stableBranches, strings.TrimSpace(branch))
		}
	}

	// sort the stable branches according to their versions
	sort.Slice(stableBranches, func(i, j int) bool {
		iVersionStr := strings.Split(stableBranches[i], "-")[1]
		jVersionStr := strings.Split(stableBranches[j], "-")[1]
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

	wantedBranches := []string{
		"master",
	}

	// only pick latest several versions
	for idx, branch := range stableBranches {
		if idx >= ossGenerateLimit-1 {
			break
		}
		wantedBranches = append(wantedBranches, branch)
	}

	projectRoot, err := getProjectRootAbsPath()
	if err != nil {
		return err
	}
	filter, err := fetchDocumentedDirctives()
	if err != nil {
		return err
	}

	// generate codes
	for _, branch := range wantedBranches {
		branch = strings.Replace(branch, "origin/", "", -1)
		cmdOutput, err := gitCheckout(tmpDir, branch)
		if err != nil {
			fmt.Println("git checkout failed, cmd output: " + cmdOutput)
			return fmt.Errorf(cmdOutput)
		}
		var ossVerStr string
		if strings.Contains(branch, "master") {
			ossVerStr = "Latest"
		} else {
			ossVerStr = strings.Split(branch, "-")[1]
			ossVerStr = strings.Join(strings.Split(ossVerStr, "."), "")
		}
		matchFnName := fmt.Sprintf("Oss%sDirectivesMatchFn", ossVerStr)
		fileName := fmt.Sprintf("./analyze_oss_%s_directives.go", lowercaseStrFirstChar(ossVerStr))
		generateSupportFileFromCode(tmpDir, ossName, fmt.Sprintf("ngxOss%sDirectives", ossVerStr), matchFnName, path.Join(projectRoot, fileName), filter)
	}

	return nil
}

// todo: delete it
var module2genFunc = map[string]func() error{
	ossName: generateOSS,
}

// todo: delete it
func generateOSS() error {
	ossTmpDir := path.Join(tmpDirPattern, ossName)
	if directoryExists(ossTmpDir) {
		err := os.RemoveAll(ossTmpDir)
		if err != nil {
			return fmt.Errorf("removing %s failed, please remove this directory mannually", ossTmpDir)
		}
	}
	os.MkdirAll(ossTmpDir, 0777)
	defer os.RemoveAll(tmpDirPattern)

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
		if idx >= ossGenerateLimit-1 {
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

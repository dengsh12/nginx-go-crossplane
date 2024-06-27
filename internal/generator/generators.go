package generator

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	tmpDirPattern = "generator_tmp_"
)

type codeGenerator interface {
	generateFromWeb() error
}

type normalGenerator struct {
	// sourceName should be OSS, NPLUS, or the name of a dynamic module
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

package generator

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
)

func getProjectRootAbsPath() (string, error) {
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("can't get path of generator_util_test.go through runtime")
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	// get the project root directory
	rootDir := filepath.Dir(absPath)
	rootDir = filepath.Dir(rootDir)
	rootDir = filepath.Dir(rootDir)

	return rootDir, nil
}

func compareFiles(file1, file2 *os.File) (bool, error) {
	b1, err := io.ReadAll(file1)
	if err != nil {
		return false, fmt.Errorf("failed to read file1: %w", err)
	}

	b2, err := io.ReadAll(file2)
	if err != nil {
		return false, fmt.Errorf("failed to read file2: %w", err)
	}

	return bytes.Equal(b1, b2), nil
}

func getTestSrcCodePath(sourceName string) (string, error) {
	root, err := getProjectRootAbsPath()
	if err != nil {
		return "", err
	}
	return path.Join(root, "internal", "generator", "testdata", "source_codes", sourceName), nil
}

func getExpectedFilePath(sourceName string) (string, error) {
	root, err := getProjectRootAbsPath()
	if err != nil {
		return "", err
	}
	return path.Join(root, "internal", "generator", "testdata", "expected", sourceName), nil
}

func TestGenSupFromSrcCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "lua_pass",
			path:    "lua",
			wantErr: false,
		},
		{
			name:    "selfDefined_pass",
			path:    "",
			wantErr: false,
		},
		{
			name: "undefinedBitmask_fail",
		},
		{
			name: "noDirective",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var err error
			codePath, err := getTestSrcCodePath(tc.name)
			if err != nil {
				t.Fatal(err)
			}

			outputFile, err := os.CreateTemp("", "TestGenSupFromSrcCode_")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(outputFile.Name())
			defer outputFile.Close()

			err = genSupFromSrcCode(codePath, "directives", "Match", outputFile)

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			err = outputFile.Sync()

			if err != nil {
				t.Fatal(err)
			}

			expectedFilePth, err := getExpectedFilePath(tc.name)
			if err != nil {
				t.Fatal(err)
			}

			expectedFile, err := os.Open(expectedFilePth)
			if err != nil {
				t.Fatal(err)
			}

			// Reset the file pointer to the beginning of the file, so that we can read from it
			_, err = outputFile.Seek(0, 0)
			if err != nil {
				t.Fatal(err)
			}

			res, err := compareFiles(outputFile, expectedFile)
			if err != nil {
				t.Fatal(err)
			}
			if res == false {
				t.Fatal("output not align with expectation")
			}
		})
	}
}

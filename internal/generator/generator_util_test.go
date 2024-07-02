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

func compareFiles(file1, file2 os.File) (bool, error) {
	b1, err := io.ReadAll(f1)
	if err != nil {
		return false, fmt.Errorf("failed to read file1: %w", err)
	}

	b2, err := io.ReadAll(f2)
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
	return path.Join(root, "internal", "testdata", "source_codes", sourceName), nil
}

func getExpectedFilePath(sourceName string) (string, error) {
	root, err := getProjectRootAbsPath()
	if err != nil {
		return "", err
	}
	return path.Join(root, "internal", "testdata", "source_codes", sourceName), nil
}

func TestGenSupFromSrcCode(t *testing.T) {
	t.Parallel()
	type arguments struct {
		codePath        string
		mapVariableName string
		mathFnName      string
		write           io.Writer
		filter          map[string]interface{}
	}
	tests := []struct {
		name            string
		args            arguments
		expectedFilePth string
		wantErr         bool
	}{}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			args := tc.args
			var err error
			args.codePath, err = getTestSrcCodePath(tc.name)
			if err != nil {
				t.Fatal(err)
			}

			outputFile, err := os.CreateTemp("", "TestGenSupFromSrcCode_")
			if err != nil {
				t.Fatal(err)
			}
			defer outputFile.Close()

			err = genSupFromSrcCode(args.codePath, args.mapVariableName, args.mathFnName, outputFile)
			
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

			res, err := compareFiles(, "")
		})
	}
}

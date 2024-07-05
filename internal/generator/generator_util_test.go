/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package generator

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func getProjectRootAbsPath() (string, error) {
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("can't get path of generator_util_test.go through runtime")
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

var (
	update = flag.Bool("update", false,
		`update the expected output of these tests, 
only use when the expected output is outdated and you are sure your output is correct`)
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

//nolint:funlen
func TestGenSupFromSrcCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		relativePath string
		wantErr      bool
	}{
		{
			name:         "lua_pass",
			relativePath: "lua",
			wantErr:      false,
		},
		{
			name:         "normalDirectiveDefinition_pass",
			relativePath: "normalDefinition",
			wantErr:      false,
		},
		{
			name:         "unknownBitmask_fail",
			relativePath: "unknownBitmask",
			wantErr:      true,
		},
		{
			name:         "noDirectivesDefinition_fail",
			relativePath: "noDirectives",
			wantErr:      true,
		},
		// For directives defined in ngx_mgmt_block_commands, there is not
		// context bitmask for them in source code. We added a ngxMgmtMainConf
		// to it in crossplane.
		{
			name:         "mgmtContext_pass",
			relativePath: "mgmtContext",
			wantErr:      false,
		},
		// If one directive was defined in several files, we should keep all
		// of the bitmask definitions
		{
			name:         "directiveRepeatDefine_pass",
			relativePath: "repeatDefine",
		},
		// If there are comments in definition, we should delete them
		{
			name:         "commentsInDefinition_pass",
			relativePath: "commentsInDefinition",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var err error
			codePath, err := getTestSrcCodePath(tc.relativePath)
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer

			err = genFromSrcCode(codePath, "directives", "Match", &buf)

			if !tc.wantErr && err != nil {
				t.Fatal(err)
			}

			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			// If the testcase wants an error and there is an error, skip the output file validation,
			// since there may not be an output file while error occurs in generation.
			if err != nil {
				return
			}

			expectedFilePth, err := getExpectedFilePath(tc.relativePath)
			if err != nil {
				t.Fatal(err)
			}

			expectedFile, err := os.OpenFile(expectedFilePth, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				t.Fatal(err)
			}

			if *update {
				expectedFile.WriteString(buf.String())
				expectedFile.Close()
				return
			}

			expected, err := io.ReadAll(expectedFile)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, string(expected), buf.String())
		})
	}
}

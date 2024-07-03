/**
 * Copyright (c) F5, Inc.
 *
 * This source code is licensed under the Apache License, Version 2.0 license found in the
 * LICENSE file in the root directory of this source tree.
 */

package main

import (
	"flag"
	"log"
	"os"

	"github.com/nginxinc/nginx-go-crossplane/internal/generator"
)

func main() {
	var (
		sourceCodePath = flag.String("src-directory", "", "the folder includes the source code your want to generate support from (required)")
	)
	flag.Parse()
	err := generator.Generate(*sourceCodePath, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}

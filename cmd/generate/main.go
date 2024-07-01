package main

//go:generate go run main.go --func=generate --source_name=lua
//go:generate go run main.go --func=generate --source_name=headersMore
//go:generate go run main.go  --func=generate --source_name=njs
//go:generate go run main.go --func=generate --source_name=otel
//go:generate go run main.go --func=generate --source_name=OSS

import (
	"flag"
	"fmt"
	"time"

	"github.com/nginxinc/nginx-go-crossplane/internal/generator"
)

// todo: delete this
func testRun() {
}

func main() {
	// generator.TestRun()
	// return
	start_t := time.Now()
	var (
		function           = flag.String("func", "", "the function you need, should be code2map, code2json, generate, or json2map (required)")
		sourceCodePath     = flag.String("source_code", "", "the folder includes the source code your want to generate support from (required when func=code2map or code2json)")
		_                  = flag.String("json_file", "", "the folder of the json file you want to generate support from (required when func=json2map)")
		sourceName         = flag.String("source_name", "", "OSS, NPLUS, or the name of the module(required)")
		outputFolder       = flag.String("output_folder", "./tmp", "the folder at which the generated support file locates, ./tmp by default(optional)")
		onlyDocumentedDirs = flag.Bool("documented_only", false, "only output consider directives on https://nginx.org/en/docs/dirindex.html, optional, false by default")
	)
	flag.Parse()
	generator.Generate(*function, *sourceName, *onlyDocumentedDirs, *sourceCodePath, *outputFolder)
	fmt.Println("use time:" + time.Since(start_t).String())
}

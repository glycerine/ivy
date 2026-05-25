// Command ivy2go is a thin CLI wrapper over the ivy2go package.
// Mirrors cmd/ivy2cpp/main.go's shape: parse argv into (params,
// filename), generate, optionally build, otherwise write outputs to
// disk.
//
// Usage:
//
//	ivy2go [key=value …] file.ivy
//
// Recognised params (see ivy2go.mergeParams for the full list):
//
//	target=impl|class|repl|test|gen
//	package=<go-package-name>     (alias: classname)
//	outdir=<dir>                  must be inside an existing Go module
//	main=<main-func-name>
//	test_iters=<n>
//	test_runs=<n>
//	trace=true|false
//	build=true|false              when true, run `go build` after WriteOutput
//	isolate=<isolate-name>|all
//
// ivy2go does NOT emit a go.mod for the generated package. Place
// outdir inside an existing Go module (e.g. the goivy repo itself,
// or a workspace that includes it) so `import "github.com/glycerine/ivy/goivy"`
// resolves naturally.
package main

import (
	"fmt"
	"os"

	"github.com/glycerine/ivy/goivy/ivy2go"
)

func main() {
	params, filename, err := ivy2go.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	batch, err := ivy2go.CompileAndGenerateAll(filename, params, ivy2go.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if ivy2go.BuildRequested(params) {
		for _, out := range batch.Outputs {
			if _, err := ivy2go.BuildOutput(out, params["outdir"]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		return
	}
	if err := ivy2go.WriteBatchOutput(batch, params["outdir"]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

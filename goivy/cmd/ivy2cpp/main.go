package main

import (
	"fmt"
	"os"

	"github.com/glycerine/ivy/goivy/ivy2cpp"
)

func main() {
	params, filename, err := ivy2cpp.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	batch, err := ivy2cpp.CompileAndGenerateAll(filename, params, ivy2cpp.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if ivy2cpp.BuildRequested(params) {
		for _, out := range batch.Outputs {
			if _, err := ivy2cpp.BuildOutput(out, params["outdir"]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		return
	}
	if err := ivy2cpp.WriteBatchOutput(batch, params["outdir"]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

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
	printWarnings(batch)
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

func printWarnings(batch *ivy2cpp.BatchOutput) {
	if batch == nil {
		return
	}
	seen := map[string]bool{}
	for _, out := range batch.Outputs {
		if out == nil {
			continue
		}
		for _, warning := range out.Warnings {
			if seen[warning] {
				continue
			}
			seen[warning] = true
			fmt.Fprintln(os.Stderr, warning)
		}
	}
}

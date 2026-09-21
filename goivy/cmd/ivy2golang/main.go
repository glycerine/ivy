package main

import (
	"fmt"
	"os"

	"github.com/glycerine/ivy/goivy/ivy2golang"
)

func main() {
	params, filename, err := ivy2golang.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	batch, err := ivy2golang.CompileAndGenerateAll(filename, params, ivy2golang.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printWarnings(batch)
	if ivy2golang.BuildRequested(params) {
		for _, out := range batch.Outputs {
			if _, err := ivy2golang.BuildOutput(out, params["outdir"]); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		return
	}
	if err := ivy2golang.WriteBatchOutput(batch, params["outdir"]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printWarnings(batch *ivy2golang.BatchOutput) {
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

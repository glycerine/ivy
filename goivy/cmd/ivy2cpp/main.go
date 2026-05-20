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
	out, err := ivy2cpp.CompileAndGenerate(filename, params, ivy2cpp.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := ivy2cpp.WriteOutput(out, params["outdir"]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

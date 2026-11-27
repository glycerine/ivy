package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func main() {
	destination := flag.String("o", "", "path for the generated RTLIL")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-o output.il] design.ivy\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	source := flag.Arg(0)
	target := *destination
	if target == "" {
		target = strings.TrimSuffix(source, filepath.Ext(source)) + ".il"
	}
	if err := goivy.WriteRTLIL(source, target); err != nil {
		fmt.Fprintf(os.Stderr, "ivy2rtl: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", target)
}

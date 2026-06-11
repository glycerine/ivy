package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/glycerine/ivy/goivy/ivyrepl"
)

func usage(myflags *flag.FlagSet) {
	fmt.Printf("ivy-repl command line help:\n")
	myflags.PrintDefaults()
	os.Exit(1)
}

func main() {
	cfg := ivyrepl.NewZlispConfig("ivy-repl")
	cfg.DefineFlags()
	err := cfg.Flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		usage(cfg.Flags)
	}

	if err != nil {
		panic(err)
	}
	err = cfg.ValidateConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ivy-repl command line error: '%v'\n", err)
		usage(cfg.Flags)
	}

	// the library does all the heavy lifting.
	ivyrepl.ReplMain(cfg)
}

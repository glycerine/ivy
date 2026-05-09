// Command goivy_check is the Go port of Python's ivy_check command.
// It checks Ivy specifications for correctness using inductive verification,
// model checking, or bounded model checking.
//
// Usage:
//
//	goivy_check [key=value ...] file.ivy
//
// Parameters are specified as key=value pairs before the .ivy filename:
//
//	goivy_check isolate=protocol_impl file.ivy
//	goivy_check diagnose=true trace=true file.ivy
//	goivy_check action=send trusted=false file.ivy
//
// Available parameters:
//
//	diagnose=bool          Launch diagnostic UI on failure (default false)
//	coverage=bool          Check assertion coverage (default true)
//	action=name            Check only the named action
//	trusted=bool           Skip verification, trust all isolates (default false)
//	mc=bool                Force model checking method (default false)
//	trace=bool             Print execution trace on counterexample (default false)
//	separate=bool          Check each assertion separately
//	isolate=name           Check only the named isolate
//	summary=bool           Summary mode, no actual checking (default false)
//	unprovable=bool        Check only unprovable assertions (default false)
//	unchecked_properties=file  YAML file specifying assumed/ignored properties
//	ivy_stats=bool         Print timing statistics (default false)
//	prioritize=a,b,c       Comma-separated actions to check first
//	no_check_guarantees=bool  Skip guarantee checking (default false)
//
// Corresponds to Python's ivy_check console_scripts entry point:
//
//	ivy_check=ivy.ivy_check:main
package main

import (
	"flag"
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var ProgramName string = "goivy_check"

type IvyCheckConfig struct {
	IncludePathStdlib string // -i where to find the "include/" dir of standard lib.
}

// call DefineFlags before myflags.Parse()
func (c *IvyCheckConfig) DefineFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.IncludePathStdlib, "i", "", "path to ivy standard lib, should contain include/ ; otherwise the embedded standard lib will be called")
}

// call c.ValidateConfig() after myflags.Parse()
func (c *IvyCheckConfig) ValidateConfig() error {
	if c.IncludePathStdlib != "" {
		if !dirExists(c.IncludePathStdlib) {
			return fmt.Errorf("%v command line error: -i include path for Ivy std lib: '%v': not found!", ProgramName, c.IncludePathStdlib)
		}
	}
	return nil
}

func main() {
	// Python: signal.signal(signal.SIGINT, signal.SIG_DFL)
	signal.Reset(syscall.SIGINT)

	cmdCfg := &IvyCheckConfig{}
	myflags := flag.NewFlagSet("goivy_check", flag.ExitOnError)
	cmdCfg.DefineFlags(myflags)

	err := myflags.Parse(os.Args[1:])
	if err != nil {
		panicf("%s command line flag parse error: '%s'", ProgramName, err)
	}
	err = cmdCfg.ValidateConfig()
	if err != nil {
		panicf("%s command line flag error: '%s'", ProgramName, err)
	}

	// Parse key=value parameters from command-line args.
	// Python: ivy_init.read_params() extracts key=value pairs from sys.argv.
	args := os.Args[1:]
	params := make(map[string]string)
	for len(args) > 0 && strings.Contains(args[0], "=") {
		parts := strings.SplitN(args[0], "=", 2)
		if len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
		args = args[1:]
	}

	// Validate: exactly one .ivy file argument remaining.
	if len(args) != 1 || !strings.HasSuffix(args[0], ".ivy") {
		fmt.Fprintf(os.Stderr, "usage: %s [key=value ...] file.ivy\n", os.Args[0])
		os.Exit(1)
	}

	// Build Config from parsed parameters.
	// Corresponds to Python's Parameter objects accessed via .get() throughout ivy_check.
	cfg := goivy.NewConfig()
	cfg.IncludePathStdlib = cmdCfg.IncludePathStdlib

	if err := goivy.ApplyIvyCheckParams(cfg, params); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Delegate to check.StartWithConfig which implements Python's
	// ivy_check.main() -> start() -> check_module() pipeline.
	os.Exit(goivy.MainWithConfig(args, cfg))
}

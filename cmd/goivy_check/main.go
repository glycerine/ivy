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
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/glycerine/goivy/check"
	iu "github.com/glycerine/goivy/ivyutils"
)

func main() {
	// Python: signal.signal(signal.SIGINT, signal.SIG_DFL)
	signal.Reset(syscall.SIGINT)

	// Python: ivy_init.read_params()
	// Parse key=value parameters from command-line args, leaving
	// only the positional .ivy filename.
	reg := iu.GlobalRegistry
	args := os.Args[1:]
	var remaining []string
	params := make(map[string]interface{})
	for len(args) > 0 && strings.Contains(args[0], "=") {
		parts := strings.SplitN(args[0], "=", 2)
		if len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
		args = args[1:]
	}
	remaining = args

	if len(params) > 0 {
		if err := iu.SetParameters(reg, params); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	// Validate: exactly one .ivy file argument
	if len(remaining) != 1 || !strings.HasSuffix(remaining[0], ".ivy") {
		fmt.Fprintf(os.Stderr, "usage: %s [key=value ...] file.ivy\n", os.Args[0])
		os.Exit(1)
	}

	// Delegate to check.Main which implements Python's
	// ivy_check.main() -> start() -> check_module() pipeline.
	os.Exit(check.Main(remaining))
}

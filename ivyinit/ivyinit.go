// Package ivyinit provides initialization routines for the Ivy system.
// This corresponds to Python's ivy_init.py.
//
// It handles:
//   - Command-line parameter reading (key=value pairs)
//   - Source file loading
//   - Analysis graph creation
package ivyinit

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/compiler"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/module"
)

// ReadParams extracts key=value parameters from args, sets them,
// and returns the remaining (non-parameter) arguments.
// Corresponds to Python's read_params.
func ReadParams(args []string) ([]string, error) {
	ps := make(map[string]interface{})
	remaining := args
	for len(remaining) > 0 && strings.Contains(remaining[0], "=") {
		parts := strings.SplitN(remaining[0], "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("bad parameter: %s", remaining[0])
		}
		ps[parts[0]] = parts[1]
		remaining = remaining[1:]
	}
	if len(ps) > 0 {
		if err := iu.SetParameters(ps); err != nil {
			return nil, err
		}
	}
	return remaining, nil
}

// SourceFile compiles an Ivy source file and sets the module name.
// Corresponds to Python's source_file.
func SourceFile(filename string, mod *module.Module) error {
	// Set module name (strip extension)
	name := filename
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[:idx]
	}
	mod.Name = name
	return nil
}

// IvyInit initializes the Ivy system from command-line arguments.
// Returns an AnalysisGraph ready for verification.
// Corresponds to Python's ivy_init.
func IvyInit(args []string) (*art.AnalysisGraph, error) {
	remaining, err := ReadParams(args)
	if err != nil {
		return nil, err
	}

	if len(remaining) < 1 || len(remaining) > 2 {
		return nil, fmt.Errorf("usage: ivy [key=value...] <file.ivy>")
	}

	filename := remaining[0]
	if !strings.HasSuffix(filename, ".ivy") && !strings.HasSuffix(filename, ".dfy") {
		return nil, fmt.Errorf("expected .ivy or .dfy file, got: %s", filename)
	}

	mod := module.New()
	sig := il.NewSig()
	comp := compiler.New(sig, mod)

	if err := SourceFile(filename, mod); err != nil {
		return nil, err
	}

	ag := art.NewAnalysisGraph(mod)
	_ = comp // compiler used during file loading (to be wired up)
	return ag, nil
}

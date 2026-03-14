// Copyright (c) Microsoft Corporation. All Rights Reserved.
// Ported to Go from ivy_to_cpp.py top-level orchestration (~lines 5700-6000).

// This file covers the top-level compilation flow: main_int(), conjecture
// injection, output file helpers, and Visual Studio path finder.
package cppgen

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ---------------------------------------------------------------------------
// Target — compilation target mode.
// ---------------------------------------------------------------------------

// Target enumerates the compilation output modes.
type Target string

const (
	TargetImpl  Target = "impl"
	TargetGen   Target = "gen"
	TargetRepl  Target = "repl"
	TargetTest  Target = "test"
	TargetClass Target = "class"
)

// CompilerOptions holds all command-line / configuration options for the
// code generator, corresponding to the Python module-level Parameter objects.
type CompilerOptions struct {
	Target    Target
	ClassName string
	Build     bool
	Trace     bool
	TestIters int
	TestRuns  int
	Compiler  string // "g++", "cl", or "default"
	MainFunc  string
	StdAfx    bool
	OutDir    string
}

// DefaultOptions returns a CompilerOptions with the default values matching
// the Python defaults.
func DefaultOptions() CompilerOptions {
	return CompilerOptions{
		Target:    TargetGen,
		ClassName: "",
		Build:     false,
		Trace:     false,
		TestIters: 100,
		TestRuns:  1,
		Compiler:  "default",
		MainFunc:  "main",
		StdAfx:    false,
		OutDir:    "",
	}
}

// ---------------------------------------------------------------------------
// GeneratorState — mutable state across a compilation run.
// ---------------------------------------------------------------------------

// GeneratorState holds the mutable state for a single compilation run.
type GeneratorState struct {
	Opts      CompilerOptions
	EmitMain  bool
	Header    strings.Builder
	Impl      strings.Builder
	ClassList []string // classes generated
}

// NewGeneratorState creates a fresh state from options.
func NewGeneratorState(opts CompilerOptions) *GeneratorState {
	return &GeneratorState{
		Opts:     opts,
		EmitMain: true,
	}
}

// ---------------------------------------------------------------------------
// add_conjs_to_actions — Conjecture injection.
// ---------------------------------------------------------------------------

// Conjecture represents a labeled conjecture (invariant candidate).
type Conjecture struct {
	Label   string
	Formula string // serialized formula
	Lineno  string // source location
}

// AddConjsToActions appends assertion checks for all conjectures to every
// public action. This mirrors Python's add_conjs_to_actions().
func AddConjsToActions(publicActions map[string]bool, conjs []Conjecture) []string {
	var assertLines []string
	for _, conj := range conjs {
		assertLines = append(assertLines,
			fmt.Sprintf("ivy_assert(%s, \"%s\");", conj.Formula, conj.Lineno))
	}
	return assertLines
}

// ---------------------------------------------------------------------------
// outfile — Output file path helper.
// ---------------------------------------------------------------------------

// Outfile returns the output path for a generated file. If OutDir is set
// the file is placed in that directory.
func Outfile(name string, outDir string) string {
	if outDir != "" {
		return filepath.Join(outDir, name)
	}
	return name
}

// ---------------------------------------------------------------------------
// find_vs — Locate Visual Studio on Windows.
// ---------------------------------------------------------------------------

// FindVS locates a suitable Visual Studio installation (10.0–15.0).
// On non-Windows systems it returns an error.
func FindVS() (string, error) {
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("FindVS is only applicable on Windows")
	}
	drive := "C"
	windir := os.Getenv("WINDIR")
	if len(windir) > 0 {
		drive = string(windir[0])
	}
	for v := 15; v >= 10; v-- {
		for _, suffix := range []string{"", " (x86)"} {
			dir := fmt.Sprintf("%s:\\Program Files%s\\Microsoft Visual Studio %d.0",
				drive, suffix, v)
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				return dir, nil
			}
		}
	}
	return "", fmt.Errorf("cannot find a suitable version of Visual Studio (require 10.0-15.0)")
}

// ---------------------------------------------------------------------------
// MainInt — Top-level compilation flow.
// ---------------------------------------------------------------------------

// MainInt orchestrates the full compilation pipeline.
// In the complete port this would:
//   1. Parse command-line options
//   2. Read and parse the Ivy source
//   3. Create isolates
//   4. Inject conjectures
//   5. Generate C++ header and implementation via module_to_cpp_class
//   6. Optionally compile the result
//
// For now it is a skeleton that demonstrates the structure.
func MainInt(opts CompilerOptions) error {
	gs := NewGeneratorState(opts)

	if opts.Target == TargetClass {
		gs.Opts.Target = TargetRepl
		gs.EmitMain = false
	}

	// Step 1: Emit boilerplate
	if gs.Opts.Target == TargetGen || gs.Opts.Target == TargetTest {
		EmitBoilerplate1(&gs.Header, &gs.Impl, gs.Opts.ClassName)
	}

	// Step 2: Emit REPL infrastructure if needed
	if gs.Opts.Target == TargetRepl {
		EmitReplImports(&gs.Header, &gs.Impl, gs.Opts.ClassName)
		EmitReplBoilerplate1(&gs.Header, &gs.Impl, gs.Opts.ClassName, gs.Opts.Trace)
		EmitParsingUtils(&gs.Impl, gs.Opts.ClassName)
		EmitReplBoilerplate1a(&gs.Header, &gs.Impl, gs.Opts.ClassName)
		EmitReplBoilerplate2(&gs.Header, &gs.Impl, gs.Opts.ClassName)
	}

	// Step 3: Emit init_gen
	if gs.Opts.Target == TargetGen || gs.Opts.Target == TargetTest {
		EmitInitGen(&gs.Header, &gs.Impl, gs.Opts.ClassName)
	}

	// Step 4: Write output files
	basename := gs.Opts.ClassName
	if basename == "" {
		basename = "output"
	}
	buildDir := "."
	if _, err := os.Stat("build"); err == nil {
		buildDir = "build"
	}
	headerPath := Outfile(filepath.Join(buildDir, basename+".h"), gs.Opts.OutDir)
	implPath := Outfile(filepath.Join(buildDir, basename+".cpp"), gs.Opts.OutDir)

	if err := os.WriteFile(headerPath, []byte(gs.Header.String()), 0644); err != nil {
		return fmt.Errorf("cannot write header: %w", err)
	}
	if err := os.WriteFile(implPath, []byte(gs.Impl.String()), 0644); err != nil {
		return fmt.Errorf("cannot write impl: %w", err)
	}

	return nil
}

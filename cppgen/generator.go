// Ported to Go from ivy_to_cpp.py top-level orchestration (~lines 5700-6000).

// This file covers the top-level compilation flow: main_int(), conjecture
// injection, output file helpers, and Visual Studio path finder.
package cppgen

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

// CompilerOptions holds all configuration options for the code generator.
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

// DefaultOptions returns CompilerOptions with default values.
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
	Header    CodeText
	Impl      CodeText
	ClassList []string
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
	Formula string
	Lineno  string
}

// AddConjsToActions returns assertion lines for all conjectures.
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

// Outfile returns the output path for a generated file.
func Outfile(name string, outDir string) string {
	if outDir != "" {
		return filepath.Join(outDir, name)
	}
	return name
}

// ---------------------------------------------------------------------------
// find_vs — Locate Visual Studio on Windows.
// ---------------------------------------------------------------------------

// FindVS locates a suitable Visual Studio installation (10.0-15.0).
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

// MainInt orchestrates the full compilation pipeline. Currently a skeleton.
func MainInt(opts CompilerOptions) error {
	gs := NewGeneratorState(opts)

	if opts.Target == TargetClass {
		gs.Opts.Target = TargetRepl
		gs.EmitMain = false
	}

	if gs.Opts.Target == TargetGen || gs.Opts.Target == TargetTest {
		EmitBoilerplate1(&gs.Header, &gs.Impl, gs.Opts.ClassName)
	}

	if gs.Opts.Target == TargetRepl {
		EmitReplImports(&gs.Header, &gs.Impl, gs.Opts.ClassName)
		EmitReplBoilerplate1(&gs.Header, &gs.Impl, gs.Opts.ClassName, gs.Opts.Trace)
		EmitParsingUtils(&gs.Impl, gs.Opts.ClassName)
		EmitReplBoilerplate1a(&gs.Header, &gs.Impl, gs.Opts.ClassName)
		EmitReplBoilerplate2(&gs.Header, &gs.Impl, gs.Opts.ClassName)
	}

	if gs.Opts.Target == TargetGen || gs.Opts.Target == TargetTest {
		EmitInitGen(&gs.Header, &gs.Impl, gs.Opts.ClassName)
	}

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

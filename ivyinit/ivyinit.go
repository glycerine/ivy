// Package ivyinit provides initialization routines for the Ivy system.
// This corresponds to Python's ivy_init.py.
//
// It handles:
//   - Command-line parameter reading (key=value pairs)
//   - Source file loading (read_module + ivy_compile pipeline)
//   - Analysis graph creation
//   - Version detection from #lang ivy header
package ivyinit

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/compiler"
	il "github.com/glycerine/goivy/ivylogic"
	iu "github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/lalr_full"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/parser"
	"github.com/glycerine/goivy/xtracer"
)

// globalIncluded tracks already-included module names across all parsers
// in a single compilation session. Matches Python's stack-based check:
//
//	if not any(p[3] in m.included for m in stack)
//
// This prevents double-includes when files include each other.
var globalIncluded = make(map[string]bool)

// ResetIncluded clears the global included set. Call at the start of
// each new top-level compilation session.
func ResetIncluded() {
	globalIncluded = make(map[string]bool)
}

// ReadParams extracts key=value parameters from args, sets them,
// and returns the remaining (non-parameter) arguments.
// Corresponds to Python's read_params (lines 38-52).
func ReadParams(args []string, reg *iu.ParameterRegistry) ([]string, error) {
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
		if err := iu.SetParameters(reg, ps); err != nil {
			return nil, err
		}
	}
	return remaining, nil
}

// ReadModule reads and parses an Ivy source file, detecting the version
// from the #lang ivy header.
// Corresponds to Python's read_module (lines 2267-2296).
func ReadModule(filename string, nested bool, cfg *module.Config) (*parser.ParseResult, error) {
	xtracer.Trace("init.ReadModule ENTER file=%s nested=%v", filename, nested)
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("not found: %s", filename)
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	// Read first line (header)
	header, err := reader.ReadString('\n')
	if err != nil && len(header) == 0 {
		return nil, fmt.Errorf("empty file: %s", filename)
	}
	header = strings.TrimSpace(header)

	// Read rest of file
	var sb strings.Builder
	sb.WriteByte('\n') // newline at beginning to preserve line numbers (matches Python)
	buf := make([]byte, 4096)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			sb.Write(buf[:n])
		}
		if readErr != nil {
			break
		}
	}
	s := sb.String()

	if strings.HasPrefix(header, "#lang ivy") {
		versionStr := strings.TrimSpace(header[len("#lang ivy"):])
		if versionStr != "" {
			oldVersion := iu.GetStringVersion()
			iu.SetStringVersion(versionStr)
			if versionStr != oldVersion {
				if nested {
					return nil, fmt.Errorf("#lang ivy%s expected in included file", oldVersion)
				}
			}
		}
		// Parse with detected version
		version := parseVersion(iu.GetStringVersion())

		if cfg != nil && cfg.UseLALRParser {
			// Use the LALR(1) goyacc-generated parser (faithful to Python PLY grammar)
			importer := func(name string) (*lalr_full.ParseResult, error) {
				pr, err := ImportModule(name, cfg)
				if err != nil {
					return nil, err
				}
				return &lalr_full.ParseResult{Decls: pr.Decls, Modules: pr.Modules, Included: pr.Included}, nil
			}
			opts := []lalr_full.ParseOption{
				lalr_full.WithImporter(importer),
				lalr_full.WithIncluded(globalIncluded),
				lalr_full.WithFilename(filename),
			}
			if cfg.AstCfg != nil {
				opts = append(opts, lalr_full.WithAstConfig(cfg.AstCfg))
			}
			if nested {
				opts = append(opts, lalr_full.WithNested())
			}
			lalrResult, parseErr := lalr_full.Parse(s, version, opts...)
			if parseErr != nil {
				return nil, fmt.Errorf("LALR parse error in %s: %w", filename, parseErr)
			}
			// Convert lalr_full.ParseResult to parser.ParseResult
			result := &parser.ParseResult{
				Decls:    lalrResult.Decls,
				Modules:  lalrResult.Modules,
				Included: lalrResult.Included,
			}
			xtracer.Trace("init.ReadModule EXIT file=%s decls=%d", filename, len(result.Decls))
			return result, nil
		}

		// Use the hand-rolled recursive-descent parser
		p := parser.New(s, version)
		// Set up include resolution (matches Python: ivy_parser.importer = import_module).
		// Share the global included set so nested includes prevent double-loading.
		p.Included = globalIncluded
		p.Importer = func(name string) (*parser.ParseResult, error) {
			return ImportModule(name, cfg)
		}
		result, parseErr := p.Parse()
		if parseErr != nil {
			return nil, fmt.Errorf("parse error in %s: %w", filename, parseErr)
		}
		xtracer.Trace("init.ReadModule EXIT file=%s decls=%d", filename, len(result.Decls))
		return result, nil
	}

	return nil, fmt.Errorf("file must begin with \"#lang ivyN.N\"")
}

// parseVersion converts a version string like "1.7" to a lexer.Version.
func parseVersion(v string) lexer.Version {
	parts := strings.SplitN(v, ".", 2)
	major := 1
	minor := 7
	if len(parts) >= 1 {
		if n, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
			major = n
		}
	}
	if len(parts) >= 2 {
		if n, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
			minor = n
		}
	}
	return lexer.Version{major, minor}
}

// ImportModule reads and parses a module by name, looking first in the
// current directory and then in the standard include directory.
// Corresponds to Python's import_module (lines 2298-2310).
func ImportModule(name string, cfg *module.Config) (res *parser.ParseResult, err error) {
	xtracer.Trace("init.ImportModule ENTER name=%s", name)
	defer func() { xtracer.Trace("init.ImportModule EXIT name=%s", name) }()

	fname := name + ".ivy"
	if _, err := os.Stat(fname); err != nil {
		// Try standard include directory
		stdDir := iu.GetStdIncludeDir()
		fname = filepath.Join(stdDir, fname)
		if _, err := os.Stat(fname); err != nil {
			return nil, fmt.Errorf("module %s not found in current directory or module path", name)
		}
	}
	// Python: with iu.SourceFile(fname): mod = read_module(f, nested=True)
	// WithSourceFile pushes/pops the global Filename for error reporting.
	var result *parser.ParseResult
	var resultErr error
	iu.WithSourceFile(fname, func() {
		result, resultErr = ReadModule(fname, true, cfg)
	})
	return result, resultErr
}

// SourceFile compiles an Ivy source file.
// Corresponds to Python's source_file (lines 69-78).
func SourceFile(filename string, mod *module.Module, sig *il.Sig, kwargs map[string]interface{}) error {
	xtracer.Trace("init.SourceFile ENTER file=%s", filename)
	defer func() { xtracer.Trace("init.SourceFile EXIT file=%s", filename) }()

	// Python: with iu.SourceFile(fn): ivy_load_file(f, **kwargs)
	// WithSourceFile pushes/pops the global Filename for error reporting.
	var outerErr error
	iu.WithSourceFile(filename, func() {
		ResetIncluded()
		result, err := ReadModule(filename, false, mod.Cfg)
		if err != nil {
			outerErr = err
			return
		}

		// Compile the declarations.
		// Corresponds to Python's ivy_compile(decls, **kwargs).
		// Pass create_isolate from kwargs (default true).
		// ivy_check passes create_isolate=false so that check_module can call
		// create_isolate separately for each isolate.
		createIsolate := true
		if v, ok := kwargs["create_isolate"]; ok {
			if b, ok := v.(bool); ok {
				createIsolate = b
			}
		}
		if err := compiler.IvyCompile(result.Decls, mod, createIsolate); err != nil {
			outerErr = err
			return
		}

		// Set module name from filename (strip extension)
		// Python: ivy_module.module.name = fn[:fn.rindex('.')]
		ext := filepath.Ext(filename)
		if ext != "" {
			mod.Name = filename[:len(filename)-len(ext)]
		} else {
			mod.Name = filename
		}
	})
	return outerErr
}

// IvyInit initializes the Ivy system from command-line arguments.
// Returns an AnalysisGraph ready for verification.
// Corresponds to Python's ivy_init (lines 80-113).
func IvyInit(args []string, reg *iu.ParameterRegistry) (*art.AnalysisGraph, error) {
	remaining, err := ReadParams(args, reg)
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

	if err := SourceFile(filename, mod, sig, nil); err != nil {
		return nil, err
	}

	ag := art.NewAnalysisGraph(mod)
	return ag, nil
}

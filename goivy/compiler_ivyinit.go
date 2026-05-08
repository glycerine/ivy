// Ivy initialization routines like ReadModule.
// This corresponds to Python's ivy_init.py.
//
// It handles:
//   - Command-line parameter reading (key=value pairs)
//   - Source file loading (read_module + ivy_compile pipeline)
//   - Analysis graph creation
//   - Version detection from #lang ivy header
package goivy

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// resetIncluded clears the included set on config. Call at the start of
// each new top-level compilation session.
func resetIncluded(cfg *Config) {
	cfg.GlobalIncluded = make(map[string]bool)
}

// ReadParams extracts key=value parameters from args, sets them,
// and returns the remaining (non-parameter) arguments.
// Corresponds to Python's read_params (lines 38-52).
func ReadParams(args []string, reg *ParameterRegistry) ([]string, error) {
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
		if err := SetParameters(reg, ps); err != nil {
			return nil, err
		}
	}
	return remaining, nil
}

// ReadModule reads and parses an Ivy source file, detecting the version
// from the #lang ivy header.
// Corresponds to Python's read_module (lines 2267-2296).
func ReadModule(filename string, nested bool, cfg *Config) (*ParseResult, error) {
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
			oldVersion := cfg.IuCfg.GetStringVersion()
			SetStringVersionOn(cfg.IuCfg, versionStr)
			if versionStr != oldVersion {
				if nested {
					return nil, fmt.Errorf("#lang ivy%s expected in included file", oldVersion)
				}
			}
		}
		// Parse with detected version
		version := parseIvyVersion(cfg.IuCfg.GetStringVersion())

		// Use the LALR(1) goyacc-generated parser (faithful to Python PLY grammar)
		importer := func(name string) (*ParseResult, error) {
			return ImportModule(name, cfg)
		}
		opts := []ParseOption{
			WithImporter(importer),
			WithIncluded(cfg.GlobalIncluded),
			WithFilename(filename),
		}
		if cfg != nil && cfg.AstCfg != nil {
			opts = append(opts, WithAstConfig(cfg.AstCfg))
		}
		if nested {
			opts = append(opts, WithNested())
		}
		result, parseErr := Parse(s, version, opts...)
		if parseErr != nil {
			return nil, fmt.Errorf("parse error in %s: %w", filename, parseErr)
		}
		xtracer.Trace("init.ReadModule EXIT file=%s decls=%d", filename, len(result.Decls))
		return result, nil
	}

	return nil, fmt.Errorf("file must begin with \"#lang ivyN.N\"")
}

// ReadModuleFromString parses an Ivy source string (with #lang header),
// used when compiling theory schemata from in-memory strings.
// Corresponds to Python's read_module(StringIO(source)).
func ReadModuleFromString(source string, cfg *Config) (*ParseResult, error) {
	// Python: sio = io.StringIO(theory); module = read_module(sio)
	// StringIO has no .name attribute, so Python traces file=?
	xtracer.Trace("init.ReadModule ENTER file=? nested=False")
	body, version := parseIvySource(source)

	var opts []ParseOption
	if cfg != nil && cfg.AstCfg != nil {
		opts = append(opts, WithAstConfig(cfg.AstCfg))
	}
	result, err := Parse(body, version, opts...)
	if err != nil {
		return nil, err
	}
	xtracer.Trace("init.ReadModule EXIT file=? decls=%d", len(result.Decls))
	return result, nil
}

// ImportModule reads and parses a module by name, looking first in the
// current directory and then in the standard include directory.
// Corresponds to Python's import_module (lines 2298-2310).
func ImportModule(name string, cfg *Config) (res *ParseResult, err error) {
	xtracer.Trace("init.ImportModule ENTER name=%s", name)
	defer func() { xtracer.Trace("init.ImportModule EXIT name=%s", name) }()

	fname := name + ".ivy"
	if _, err := os.Stat(fname); err != nil {
		// Try standard include directory
		stdDir := cfg.IuCfg.GetStdIncludeDir()
		fname = filepath.Join(stdDir, fname)
		if _, err := os.Stat(fname); err != nil {
			return nil, fmt.Errorf("module %s not found in current directory or module path", name)
		}
	}
	// Python: with iu.SourceFile(fname): mod = read_module(f, nested=True)
	// WithSourceFile pushes/pops the global Filename for error reporting.
	var result *ParseResult
	var resultErr error
	cfg.IuCfg.WithSourceFile(fname, func() {
		result, resultErr = ReadModule(fname, true, cfg)
	})
	return result, resultErr
}

// SourceFile compiles an Ivy source file.
// Corresponds to Python's source_file (lines 69-78).
func SourceFile(filename string, mod *Module, sig *Sig, kwargs map[string]interface{}) error {
	xtracer.Trace("init.SourceFile ENTER file=%s", filename)
	defer func() { xtracer.Trace("init.SourceFile EXIT file=%s", filename) }()

	// Python: with iu.SourceFile(fn): ivy_load_file(f, **kwargs)
	// WithSourceFile pushes/pops the global Filename for error reporting.
	var outerErr error
	mod.Cfg.IuCfg.WithSourceFile(filename, func() {
		resetIncluded(mod.Cfg)
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
		if err := IvyCompile(result.Decls, mod, createIsolate); err != nil {
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
func IvyInit(cfg *Config, args []string, reg *ParameterRegistry) (*AnalysisGraph, error) {
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

	mod := NewModule()
	sig := NewSigOn(mod.Cfg.IuCfg)

	if err := SourceFile(filename, mod, sig, nil); err != nil {
		return nil, err
	}

	ag := NewAnalysisGraph(mod)
	return ag, nil
}

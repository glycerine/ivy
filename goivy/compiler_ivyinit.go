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
	"fmt"
	"path/filepath"
	"strings"

	"github.com/glycerine/ivy/goivy/fileops"
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
		parts := strings.Split(remaining[0], "=")
		if len(parts) > 2 {
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
	return readModuleWithParent(filename, nested, cfg, nil)
}

func setConfigLanguageVersion(cfg *Config, version string) {
	version = strings.TrimSpace(version)
	if version == "" || cfg == nil {
		return
	}
	if cfg.IuCfg != nil {
		SetStringVersionOn(cfg.IuCfg, version)
	}
	if cfg.IsolateCfg != nil {
		cfg.IsolateCfg.IvyVersion = version
	}
}

func versionString(version Version) string {
	return fmt.Sprintf("%d.%d", version[0], version[1])
}

func configLanguageVersionOrDefault(cfg *Config) Version {
	if cfg != nil && cfg.IuCfg != nil {
		if version := strings.TrimSpace(cfg.IuCfg.GetStringVersion()); version != "" {
			return parseIvyVersion(version)
		}
	}
	return Version{1, 7}
}

func readModuleWithParent(filename string, nested bool, cfg *Config, parent *ivyAccum) (*ParseResult, error) {
	xtracer.Trace("init.ReadModule ENTER file=%s nested=%v", filename, nested)
	data, err := fileops.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("not found: %s", filename)
	}

	source := string(data)
	header, rest, hasRest := strings.Cut(source, "\n")
	header = strings.TrimSpace(header)
	if header == "" {
		return nil, fmt.Errorf("empty file: %s", filename)
	}

	var sb strings.Builder
	sb.WriteByte('\n') // newline at beginning to preserve line numbers (matches Python)
	if hasRest {
		sb.WriteString(rest)
	}
	s := sb.String()

	if strings.HasPrefix(header, "#lang ivy") {
		versionStr := strings.TrimSpace(header[len("#lang ivy"):])
		if versionStr != "" {
			oldVersion := cfg.IuCfg.GetStringVersion()
			setConfigLanguageVersion(cfg, versionStr)
			if versionStr != oldVersion {
				if nested {
					return nil, fmt.Errorf("#lang ivy%s expected in included file", oldVersion)
				}
			}
		}
		// Parse with detected version
		version := parseIvyVersion(cfg.IuCfg.GetStringVersion())

		// Use the LALR(1) goyacc-generated parser (faithful to Python PLY grammar)
		importer := func(name string, parent *ivyAccum) (*ParseResult, error) {
			return importModuleWithParent(name, cfg, parent)
		}
		opts := []ParseOption{
			WithImporter(importer),
			WithIncluded(cfg.GlobalIncluded),
			WithFilename(filename),
		}
		if parent != nil {
			opts = append(opts, WithParentAccum(parent))
		}
		if cfg != nil && cfg.AstCfg != nil {
			opts = append(opts, WithAstConfig(cfg.AstCfg))
		}
		if nested {
			opts = append(opts, WithNested())
		}
		result, parseErr := Parse(s, version, opts...)
		if parseErr != nil {
			return nil, parseErr
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
	body, version := parseIvySourceWithBareLangVersion(source, configLanguageVersionOrDefault(cfg))
	setConfigLanguageVersion(cfg, versionString(version))

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

// ReadModuleFromNamedString parses an Ivy source string while preserving the
// caller supplied filename for traces and diagnostics. This is the browser
// counterpart of ReadModule: goldweb sends a spec string, but conformance still
// needs stable file names in xtrace output.
func ReadModuleFromNamedString(filename, source string, nested bool, cfg *Config) (*ParseResult, error) {
	return readModuleFromNamedStringWithParent(filename, source, nested, cfg, nil)
}

func readModuleFromNamedStringWithParent(filename, source string, nested bool, cfg *Config, parent *ivyAccum) (*ParseResult, error) {
	xtracer.Trace("init.ReadModule ENTER file=%s nested=%v", filename, nested)

	header, rest, hasRest := strings.Cut(source, "\n")
	header = strings.TrimSpace(header)
	if header == "" {
		return nil, fmt.Errorf("empty file: %s", filename)
	}

	var sb strings.Builder
	sb.WriteByte('\n') // preserve line numbers, matching ReadModule.
	if hasRest {
		sb.WriteString(rest)
	}
	s := sb.String()

	if strings.HasPrefix(header, "#lang ivy") {
		versionStr := strings.TrimSpace(header[len("#lang ivy"):])
		if versionStr != "" {
			oldVersion := cfg.IuCfg.GetStringVersion()
			setConfigLanguageVersion(cfg, versionStr)
			if versionStr != oldVersion {
				if nested {
					return nil, fmt.Errorf("#lang ivy%s expected in included file", oldVersion)
				}
			}
		}
		version := parseIvyVersion(cfg.IuCfg.GetStringVersion())
		importer := func(name string, parent *ivyAccum) (*ParseResult, error) {
			return importModuleWithParent(name, cfg, parent)
		}
		opts := []ParseOption{
			WithImporter(importer),
			WithIncluded(cfg.GlobalIncluded),
			WithFilename(filename),
		}
		if parent != nil {
			opts = append(opts, WithParentAccum(parent))
		}
		if cfg != nil && cfg.AstCfg != nil {
			opts = append(opts, WithAstConfig(cfg.AstCfg))
		}
		if nested {
			opts = append(opts, WithNested())
		}
		result, parseErr := Parse(s, version, opts...)
		if parseErr != nil {
			return nil, parseErr
		}
		xtracer.Trace("init.ReadModule EXIT file=%s decls=%d", filename, len(result.Decls))
		return result, nil
	}

	return nil, fmt.Errorf("file must begin with \"#lang ivyN.N\"")
}

// ImportModule reads and parses a module by name, looking first in the
// current directory and then in the standard include directory.
// Corresponds to Python's import_module (lines 2298-2310).
func ImportModule(name string, cfg *Config) (res *ParseResult, err error) {
	return importModuleWithParent(name, cfg, nil)
}

func importModuleWithParent(name string, cfg *Config, parent *ivyAccum) (res *ParseResult, err error) {
	xtracer.Trace("init.ImportModule ENTER name=%s", name)
	defer func() { xtracer.Trace("init.ImportModule EXIT name=%s", name) }()

	fname := name + ".ivy"
	if !ivyReadableFileExists(fname) {
		if cfg != nil && cfg.StandardLibrary != nil && cfg.IuCfg != nil {
			if cachedName, source, ok := cfg.StandardLibrary.includeSource(cfg.IuCfg.GetStringVersion(), name); ok {
				var result *ParseResult
				var resultErr error
				cfg.IuCfg.WithSourceFile(cachedName, func() {
					result, resultErr = readModuleFromNamedStringWithParent(cachedName, source, true, cfg, parent)
				})
				return result, resultErr
			}
		}

		// Try standard include directory
		stdDir := cfg.IuCfg.GetStdIncludeDir()
		fname = filepath.Join(stdDir, fname)
		if !ivyReadableFileExists(fname) {
			return nil, fmt.Errorf("module %s not found in current directory or module path", name)
		}
	}
	// Python: with iu.SourceFile(fname): mod = read_module(f, nested=True)
	// WithSourceFile pushes/pops the global Filename for error reporting.
	var result *ParseResult
	var resultErr error
	cfg.IuCfg.WithSourceFile(fname, func() {
		result, resultErr = readModuleWithParent(fname, true, cfg, parent)
	})
	return result, resultErr
}

// SourceFile compiles an Ivy source file.
// Corresponds to Python's source_file (lines 69-78).
func SourceFile(filename string, mod *Module, sig *Sig, kwargs map[string]interface{}) error {
	xtracer.Trace("init.SourceFile ENTER file=%s", filename)

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
	if outerErr != nil {
		return outerErr
	}
	xtracer.Trace("init.SourceFile EXIT file=%s", filename)
	return nil
}

// SourceString compiles an Ivy source string under filename. It mirrors
// SourceFile closely enough that goivy_check conformance traces can compare a
// browser-supplied in-memory spec with Python ivy_check's file-backed run.
func SourceString(filename, source string, mod *Module, sig *Sig, kwargs map[string]interface{}) error {
	xtracer.Trace("init.SourceFile ENTER file=%s", filename)

	var outerErr error
	mod.Cfg.IuCfg.WithSourceFile(filename, func() {
		resetIncluded(mod.Cfg)
		result, err := ReadModuleFromNamedString(filename, source, false, mod.Cfg)
		if err != nil {
			outerErr = err
			return
		}

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

		ext := filepath.Ext(filename)
		if ext != "" {
			mod.Name = filename[:len(filename)-len(ext)]
		} else {
			mod.Name = filename
		}
	})
	if outerErr != nil {
		return outerErr
	}
	xtracer.Trace("init.SourceFile EXIT file=%s", filename)
	return nil
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

	mod := New()
	sig := NewSigOn(mod.Cfg.IuCfg)

	if err := SourceFile(filename, mod, sig, nil); err != nil {
		return nil, err
	}

	ag := NewAnalysisGraph(mod)
	return ag, nil
}

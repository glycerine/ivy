package main

/*
#cgo darwin CXXFLAGS: -std=c++20 -I/usr/local/include/node -DNODE_SHARED_MODE
#cgo darwin LDFLAGS: -L/usr/local/lib -lnode.141 -Wl,-rpath,/usr/local/lib
#cgo linux CXXFLAGS: -std=c++20 -I/usr/local/include/node -DNODE_SHARED_MODE
#cgo linux LDFLAGS: -L/usr/local/lib -lnode -Wl,-rpath,/usr/local/lib
#include <stdlib.h>
#include "node_bridge.h"
*/
import "C"

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unsafe"

	gojr "github.com/glycerine/ivy/gojr"
)

type nodeRuntime struct {
	ptr *C.gojr_node_runtime
}

type nodeCallMode int

const (
	nodeCallEval nodeCallMode = iota
	nodeCallEvalFiles
	nodeCallSetSheet
	nodeCallTest
	nodeCallTestFiles
	nodeCallBuild
)

type evalResult struct {
	OK           bool                 `json:"ok"`
	Incomplete   bool                 `json:"incomplete"`
	Diagnostics  []string             `json:"diagnostics"`
	Output       string               `json:"output"`
	Value        string               `json:"value"`
	ValueIsNil   bool                 `json:"valueIsNil"`
	ObservedDeps []observedDependency `json:"observedDeps"`
}

type observedDependency struct {
	Kind  string `json:"kind"`
	Sheet string `json:"sheet"`
	Cell  string `json:"cell,omitempty"`
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

type buildRequest struct {
	ImportPath         string         `json:"importPath,omitempty"`
	Files              []sourceFile   `json:"files"`
	PackageSources     packageSources `json:"packageSources,omitempty"`
	SourceRoots        []string       `json:"sourceRoots,omitempty"`
	ArtifactRoot       string         `json:"artifactRoot,omitempty"`
	PackageCacheParent string         `json:"packageCacheParent,omitempty"`
}

type cacheRequest struct {
	Action             string `json:"action"`
	PackageCacheParent string `json:"packageCacheParent,omitempty"`
	ArtifactRoot       string `json:"artifactRoot,omitempty"`
	Yes                bool   `json:"yes,omitempty"`
}

type compileRequest struct {
	Files       []sourceFile         `json:"files"`
	SheetJSON   string               `json:"sheetJSON,omitempty"`
	SheetsJSON  string               `json:"sheetsJSON,omitempty"`
	Packages    []runtimePackageSpec `json:"packages,omitempty"`
	SourceRoots []string             `json:"sourceRoots,omitempty"`
}

type evalWithPackagesRequest struct {
	ImportPath  string               `json:"importPath,omitempty"`
	PackageName string               `json:"packageName,omitempty"`
	Source      string               `json:"source,omitempty"`
	Files       []sourceFile         `json:"files,omitempty"`
	SheetJSON   string               `json:"sheetJSON,omitempty"`
	Packages    []runtimePackageSpec `json:"packages,omitempty"`
	SourceRoots []string             `json:"sourceRoots,omitempty"`
}

type runtimePackageSpec struct {
	ImportPath  string       `json:"importPath"`
	PackageName string       `json:"packageName,omitempty"`
	Files       []sourceFile `json:"files"`
}

type compileResult struct {
	OK          bool     `json:"ok"`
	Diagnostics []string `json:"diagnostics"`
	Output      string   `json:"output"`
}

type buildResult struct {
	OK          bool            `json:"ok"`
	Diagnostics []string        `json:"diagnostics"`
	Output      string          `json:"output"`
	Artifacts   []buildArtifact `json:"artifacts"`
	Built       []string        `json:"built"`
	Skipped     []string        `json:"skipped"`
}

type inspectJSResult struct {
	OK          bool            `json:"ok"`
	Diagnostics []string        `json:"diagnostics"`
	Output      string          `json:"output"`
	Artifacts   []buildArtifact `json:"artifacts"`
	Built       []string        `json:"built"`
	Skipped     []string        `json:"skipped"`
	Source      string          `json:"source"`
}

type cacheResult struct {
	OK          bool         `json:"ok"`
	Diagnostics []string     `json:"diagnostics"`
	Action      string       `json:"action"`
	Root        string       `json:"root"`
	Entries     []cacheEntry `json:"entries,omitempty"`
	Cleared     []string     `json:"cleared,omitempty"`
}

type cacheEntry struct {
	Path                string        `json:"path"`
	ImportPath          string        `json:"importPath"`
	Size                int64         `json:"size"`
	PackageName         string        `json:"packageName,omitempty"`
	CacheKey            string        `json:"cacheKey,omitempty"`
	SourceHash          string        `json:"sourceHash,omitempty"`
	LayoutVersion       string        `json:"layoutVersion,omitempty"`
	CompilerVersion     string        `json:"compilerVersion,omitempty"`
	Backend             string        `json:"backend,omitempty"`
	HostSpecVersion     string        `json:"hostSpecVersion,omitempty"`
	CapabilityPolicy    string        `json:"capabilityPolicy,omitempty"`
	Dependencies        []string      `json:"dependencies,omitempty"`
	DependencyCacheKeys []string      `json:"dependencyCacheKeys,omitempty"`
	Exports             []buildExport `json:"exports,omitempty"`
}

type fixtureResult struct {
	OK           bool                            `json:"ok"`
	Unstable     bool                            `json:"unstable"`
	Diagnostics  []string                        `json:"diagnostics"`
	Evaluated    []fixtureCellRef                `json:"evaluated"`
	Sheets       map[string]map[string]string    `json:"sheets"`
	ObservedDeps map[string][]observedDependency `json:"observedDeps"`
}

type fixtureWithPackagesRequest struct {
	FixtureJSON string               `json:"fixtureJSON"`
	Packages    []runtimePackageSpec `json:"packages,omitempty"`
	SourceRoots []string             `json:"sourceRoots,omitempty"`
}

type fixtureCellRef struct {
	Sheet string `json:"sheet"`
	Cell  string `json:"cell"`
}

type buildArtifact struct {
	ImportPath          string        `json:"importPath"`
	PackageName         string        `json:"packageName"`
	ArtifactPath        string        `json:"artifactPath"`
	Action              string        `json:"action"`
	SourceHash          string        `json:"sourceHash"`
	CacheKey            string        `json:"cacheKey"`
	Dependencies        []string      `json:"dependencies,omitempty"`
	DependencyCacheKeys []string      `json:"dependencyCacheKeys,omitempty"`
	Exports             []buildExport `json:"exports,omitempty"`
}

type buildExport struct {
	Name               string `json:"name"`
	Kind               string `json:"kind"`
	TypeText           string `json:"typeText"`
	UnderlyingTypeText string `json:"underlyingTypeText,omitempty"`
}

type embeddedModuleBundle struct {
	Modules []embeddedModule `json:"modules"`
}

type embeddedModule struct {
	Path   string `json:"path"`
	Source string `json:"source"`
}

type sourceFile struct {
	Filename string `json:"filename"`
	Source   string `json:"source"`
}

type runSourceTarget struct {
	Files       []sourceFile
	Package     bool
	ImportPath  string
	PackageName string
}

type packageSources map[string][]sourceFile

type packageFlag []string

func (flags *packageFlag) String() string {
	return strings.Join(*flags, ",")
}

func (flags *packageFlag) Set(value string) error {
	*flags = append(*flags, value)
	return nil
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help", "-h", "--help":
			printTopLevelUsage()
			return
		}
		if seed, ok, err := seedFromArgs(os.Args[2:]); err != nil {
			fatal(err)
		} else if ok {
			if err := os.Setenv("GOJR_RANDOM_SEED", seed); err != nil {
				fatal(err)
			}
		}
	}

	bootstrapSource, err := runtimeBootstrapSource()
	if err != nil {
		fatal(err)
	}
	moduleBundle, err := runtimeModuleBundle()
	if err != nil {
		fatal(err)
	}

	rt, err := newNodeRuntime(bootstrapSource, moduleBundle)
	if err != nil {
		fatal(err)
	}
	defer rt.Close()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "eval":
			ok, err := runEval(rt, os.Args[2:])
			if err != nil {
				fatal(err)
			}
			if !ok {
				os.Exit(1)
			}
			return
		case "run":
			ok, err := runSource(rt, os.Args[2:])
			if err != nil {
				fatal(err)
			}
			if !ok {
				os.Exit(1)
			}
			return
		case "compile":
			ok, err := runCompile(rt, os.Args[2:])
			if err != nil {
				fatal(err)
			}
			if !ok {
				os.Exit(1)
			}
			return
		case "test":
			ok, err := runTest(rt, os.Args[2:])
			if err != nil {
				fatal(err)
			}
			if !ok {
				os.Exit(1)
			}
			return
		case "build":
			ok, err := runBuild(rt, os.Args[2:])
			if err != nil {
				fatal(err)
			}
			if !ok {
				os.Exit(1)
			}
			return
		case "inspect-js":
			ok, err := runInspectJS(rt, os.Args[2:])
			if err != nil {
				fatal(err)
			}
			if !ok {
				os.Exit(1)
			}
			return
		case "cache":
			ok, err := runCache(rt, os.Args[2:])
			if err != nil {
				fatal(err)
			}
			if !ok {
				os.Exit(1)
			}
			return
		case "run-fixture":
			ok, err := runFixture(rt, os.Args[2:])
			if err != nil {
				fatal(err)
			}
			if !ok {
				os.Exit(1)
			}
			return
		default:
			fatal(fmt.Errorf("unknown command %q", os.Args[1]))
		}
	}

	fmt.Printf("Go-junior REPL (embedded Node/V8)\n")
	fmt.Printf("runtime: embedded dist/src/index.js\n")
	fmt.Printf("commands: .help .clear .source PATH .sheet JSON .load DIR .test PATH .quit\n")

	if err := repl(rt); err != nil {
		fatal(err)
	}
}

func runEval(rt *nodeRuntime, args []string) (bool, error) {
	flags := flag.NewFlagSet("gojr eval", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	sheetJSON := flags.String("sheet-json", "", "current sheet data as JSON")
	seed := flags.String("seed", "", "deterministic scheduler/random seed; must appear before Node starts")
	randomSeed := flags.String("random-seed", "", "deterministic scheduler/random seed; alias for --seed")
	jsonMode := flags.Bool("json", false, "print a machine-readable JSON result")
	var packageFlags packageFlag
	flags.Var(&packageFlags, "pkg", "Go-junior source package, import/path=DIR; may be repeated")
	var sourceRootFlags packageFlag
	flags.Var(&sourceRootFlags, "srcroot", "filesystem source root for resolving imported Go-junior packages; may be repeated")
	if err := flags.Parse(args); err != nil {
		return false, err
	}
	_ = seed
	_ = randomSeed
	if flags.NArg() == 0 {
		return false, fmt.Errorf("usage: gojr eval [--sheet-json JSON] [--pkg import=DIR] [--srcroot DIR] [--seed SEED] SOURCE")
	}
	packages, err := readRuntimePackageSpecs(packageFlags)
	if err != nil {
		return false, err
	}
	sourceRoots := buildSourceRoots("", "", sourceRootFlags)
	if len(packages) > 0 || len(sourceRoots) > 0 {
		result, err := rt.EvalWithPackages(evalWithPackagesRequest{
			Source:      strings.Join(flags.Args(), " "),
			SheetJSON:   strings.TrimSpace(*sheetJSON),
			Packages:    packages,
			SourceRoots: sourceRoots,
		})
		if err != nil {
			return false, err
		}
		printEvalResult(result, *jsonMode)
		return result.OK && !result.Incomplete, nil
	}
	if err := applySheetJSON(rt, strings.TrimSpace(*sheetJSON)); err != nil {
		return false, err
	}
	result, err := rt.Eval(strings.Join(flags.Args(), " "))
	if err != nil {
		return false, err
	}
	printEvalResult(result, *jsonMode)
	return result.OK && !result.Incomplete, nil
}

func runSource(rt *nodeRuntime, args []string) (bool, error) {
	flags := flag.NewFlagSet("gojr run", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	sheetJSON := flags.String("sheet-json", "", "current sheet data as JSON")
	seed := flags.String("seed", "", "deterministic scheduler/random seed; must appear before Node starts")
	randomSeed := flags.String("random-seed", "", "deterministic scheduler/random seed; alias for --seed")
	jsonMode := flags.Bool("json", false, "print a machine-readable JSON result")
	var packageFlags packageFlag
	flags.Var(&packageFlags, "pkg", "Go-junior source package, import/path=DIR; may be repeated")
	var sourceRootFlags packageFlag
	flags.Var(&sourceRootFlags, "srcroot", "filesystem source root for resolving imported Go-junior packages; may be repeated")
	if err := flags.Parse(args); err != nil {
		return false, err
	}
	_ = seed
	_ = randomSeed
	if flags.NArg() > 1 {
		return false, fmt.Errorf("usage: gojr run [--sheet-json JSON] [--pkg import=DIR] [--srcroot DIR] [--seed SEED] [FILE|DIR|-]")
	}
	sourcePath := optionalArg(flags.Args())
	target, err := readRunTarget(sourcePath)
	if err != nil {
		return false, err
	}
	packages, err := readRuntimePackageSpecs(packageFlags)
	if err != nil {
		return false, err
	}

	if target.Package {
		sourceRoots := buildSourceRoots(sourcePath, target.ImportPath, sourceRootFlags)
		result, err := rt.RunMainFilesWithPackages(evalWithPackagesRequest{
			ImportPath:  target.ImportPath,
			PackageName: target.PackageName,
			Files:       target.Files,
			SheetJSON:   strings.TrimSpace(*sheetJSON),
			Packages:    packages,
			SourceRoots: sourceRoots,
		})
		if err != nil {
			return false, err
		}
		printEvalResult(result, *jsonMode)
		return result.OK && !result.Incomplete, nil
	}

	source := target.Files[0]
	sourceRoots := buildSourceRoots(sourcePath, "", sourceRootFlags)
	if len(packages) > 0 || len(sourceRoots) > 0 {
		result, err := rt.EvalFilesWithPackages(evalWithPackagesRequest{
			Files:       []sourceFile{source},
			SheetJSON:   strings.TrimSpace(*sheetJSON),
			Packages:    packages,
			SourceRoots: sourceRoots,
		})
		if err != nil {
			return false, err
		}
		printEvalResult(result, *jsonMode)
		return result.OK && !result.Incomplete, nil
	}
	if err := applySheetJSON(rt, strings.TrimSpace(*sheetJSON)); err != nil {
		return false, err
	}
	result, err := rt.EvalFiles([]sourceFile{source})
	if err != nil {
		return false, err
	}
	printEvalResult(result, *jsonMode)
	return result.OK && !result.Incomplete, nil
}

func runCompile(rt *nodeRuntime, args []string) (bool, error) {
	flags := flag.NewFlagSet("gojr compile", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	sheetJSON := flags.String("sheet-json", "", "current sheet data as JSON")
	sheetsJSON := flags.String("sheets-json", "", "named sheet data as JSON object of objects")
	expr := flags.String("expr", "", "compile one Go-junior expression or statement list")
	jsonMode := flags.Bool("json", false, "print a machine-readable JSON result")
	var packageFlags packageFlag
	flags.Var(&packageFlags, "pkg", "Go-junior source package dependency, import/path=DIR; may be repeated")
	var sourceRootFlags packageFlag
	flags.Var(&sourceRootFlags, "srcroot", "filesystem source root for resolving imported Go-junior packages; may be repeated")
	if err := flags.Parse(args); err != nil {
		return false, err
	}
	if *expr != "" && flags.NArg() != 0 {
		return false, fmt.Errorf("usage: gojr compile [--json] [--sheet-json JSON] [--pkg import=DIR] [--srcroot DIR] [--expr SOURCE] [FILE|DIR|-]")
	}

	var files []sourceFile
	var err error
	target := optionalArg(flags.Args())
	if *expr != "" {
		files = []sourceFile{{Filename: "gojr-repl.go", Source: *expr}}
	} else {
		files, err = readCompileTarget(target)
		if err != nil {
			return false, err
		}
	}
	packages, err := readRuntimePackageSpecs(packageFlags)
	if err != nil {
		return false, err
	}

	result, err := rt.Compile(compileRequest{
		Files:       files,
		SheetJSON:   strings.TrimSpace(*sheetJSON),
		SheetsJSON:  strings.TrimSpace(*sheetsJSON),
		Packages:    packages,
		SourceRoots: buildSourceRoots(target, "", sourceRootFlags),
	})
	if err != nil {
		return false, err
	}
	printCompileResult(result, *jsonMode)
	return result.OK, nil
}

func runTest(rt *nodeRuntime, args []string) (bool, error) {
	flags := flag.NewFlagSet("gojr test", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	seed := flags.String("seed", "", "deterministic scheduler/random seed; must appear before Node starts")
	randomSeed := flags.String("random-seed", "", "deterministic scheduler/random seed; alias for --seed")
	jsonMode := flags.Bool("json", false, "print a machine-readable JSON result")
	var packageFlags packageFlag
	flags.Var(&packageFlags, "pkg", "Go-junior source package dependency, import/path=DIR; may be repeated")
	var sourceRootFlags packageFlag
	flags.Var(&sourceRootFlags, "srcroot", "filesystem source root for resolving imported Go-junior packages; may be repeated")
	if err := flags.Parse(args); err != nil {
		return false, err
	}
	_ = seed
	_ = randomSeed
	if flags.NArg() != 1 {
		return false, fmt.Errorf("usage: gojr test [--pkg import=DIR] [--srcroot DIR] [--seed SEED] PATH")
	}
	target := flags.Arg(0)
	files, err := readTestTarget(target)
	if err != nil {
		return false, err
	}
	packages, err := readRuntimePackageSpecs(packageFlags)
	if err != nil {
		return false, err
	}
	sourceRoots := buildSourceRoots(target, "", sourceRootFlags)
	var result evalResult
	if len(packages) > 0 || len(sourceRoots) > 0 {
		result, err = rt.TestFilesWithPackages(evalWithPackagesRequest{
			Files:       files,
			Packages:    packages,
			SourceRoots: sourceRoots,
		})
	} else {
		result, err = rt.TestFiles(files)
	}
	if err != nil {
		return false, err
	}
	printTestResult(result, *jsonMode)
	return result.OK, nil
}

func runBuild(rt *nodeRuntime, args []string) (bool, error) {
	flags := flag.NewFlagSet("gojr build", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	importPath := flags.String("importpath", "", "package import path for the generated artifact")
	packageCacheParent := flags.String("pkgdir", "", "package-cache parent directory; gojr_js is appended")
	artifactRoot := flags.String("artifact-root", "", "exact gojr_js artifact root directory")
	jsonMode := flags.Bool("json", false, "print a machine-readable JSON build report")
	var packageFlags packageFlag
	flags.Var(&packageFlags, "pkg", "Go-junior source package dependency, import/path=DIR; may be repeated")
	var sourceRootFlags packageFlag
	flags.Var(&sourceRootFlags, "srcroot", "filesystem source root for resolving imported Go-junior packages; may be repeated")
	if err := flags.Parse(args); err != nil {
		return false, err
	}
	if flags.NArg() != 1 {
		return false, fmt.Errorf("usage: gojr build [-importpath PATH] [--pkg import=DIR] [--srcroot DIR] [-pkgdir DIR|-artifact-root DIR] TARGET")
	}
	target := flags.Arg(0)
	files, err := readBuildTarget(target)
	if err != nil {
		return false, err
	}
	packageSpecs, err := readRuntimePackageSpecs(packageFlags)
	if err != nil {
		return false, err
	}
	resolvedImportPath := strings.TrimSpace(*importPath)
	if resolvedImportPath == "" {
		resolvedImportPath = deriveBuildImportPath(target, packageNameFromSourceFiles(files))
	}
	result, err := rt.Build(buildRequest{
		ImportPath:         resolvedImportPath,
		Files:              files,
		PackageSources:     packageSourceMap(packageSpecs),
		SourceRoots:        buildSourceRoots(target, resolvedImportPath, sourceRootFlags),
		ArtifactRoot:       strings.TrimSpace(*artifactRoot),
		PackageCacheParent: strings.TrimSpace(*packageCacheParent),
	})
	if err != nil {
		return false, err
	}
	printBuildResult(result, *jsonMode)
	return result.OK, nil
}

func runInspectJS(rt *nodeRuntime, args []string) (bool, error) {
	flags := flag.NewFlagSet("gojr inspect-js", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	importPath := flags.String("importpath", "", "package import path for the generated artifact")
	packageCacheParent := flags.String("pkgdir", "", "package-cache parent directory; gojr_js is appended")
	artifactRoot := flags.String("artifact-root", "", "exact gojr_js artifact root directory")
	jsonMode := flags.Bool("json", false, "print a machine-readable JSON inspect report")
	var packageFlags packageFlag
	flags.Var(&packageFlags, "pkg", "Go-junior source package dependency, import/path=DIR; may be repeated")
	var sourceRootFlags packageFlag
	flags.Var(&sourceRootFlags, "srcroot", "filesystem source root for resolving imported Go-junior packages; may be repeated")
	if err := flags.Parse(args); err != nil {
		return false, err
	}
	if flags.NArg() != 1 {
		return false, fmt.Errorf("usage: gojr inspect-js [--json] [-importpath PATH] [--pkg import=DIR] [--srcroot DIR] [-pkgdir DIR|-artifact-root DIR] TARGET")
	}
	target := flags.Arg(0)
	files, err := readBuildTarget(target)
	if err != nil {
		return false, err
	}
	packageSpecs, err := readRuntimePackageSpecs(packageFlags)
	if err != nil {
		return false, err
	}
	resolvedImportPath := strings.TrimSpace(*importPath)
	if resolvedImportPath == "" {
		resolvedImportPath = deriveBuildImportPath(target, packageNameFromSourceFiles(files))
	}
	result, err := rt.InspectJS(buildRequest{
		ImportPath:         resolvedImportPath,
		Files:              files,
		PackageSources:     packageSourceMap(packageSpecs),
		SourceRoots:        buildSourceRoots(target, resolvedImportPath, sourceRootFlags),
		ArtifactRoot:       strings.TrimSpace(*artifactRoot),
		PackageCacheParent: strings.TrimSpace(*packageCacheParent),
	})
	if err != nil {
		return false, err
	}
	printInspectJSResult(result, *jsonMode)
	return result.OK, nil
}

func runCache(rt *nodeRuntime, args []string) (bool, error) {
	if len(args) == 0 {
		return false, fmt.Errorf("usage: gojr cache path|list|clear [--json] [-pkgdir DIR|-artifact-root DIR] [--yes]")
	}
	action := args[0]
	flags := flag.NewFlagSet("gojr cache "+action, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	packageCacheParent := flags.String("pkgdir", "", "package-cache parent directory; gojr_js is appended")
	artifactRoot := flags.String("artifact-root", "", "exact gojr_js artifact root directory")
	jsonMode := flags.Bool("json", false, "print a machine-readable JSON cache report")
	yes := flags.Bool("yes", false, "confirm destructive cache clear")
	if err := flags.Parse(args[1:]); err != nil {
		return false, err
	}
	if flags.NArg() != 0 {
		return false, fmt.Errorf("usage: gojr cache %s [--json] [-pkgdir DIR|-artifact-root DIR] [--yes]", action)
	}

	if action != "path" && action != "list" && action != "clear" {
		return false, fmt.Errorf("unknown cache command %q", action)
	}
	result, err := rt.Cache(cacheRequest{
		Action:             action,
		PackageCacheParent: strings.TrimSpace(*packageCacheParent),
		ArtifactRoot:       strings.TrimSpace(*artifactRoot),
		Yes:                *yes,
	})
	if err != nil {
		return false, err
	}
	if !result.OK && len(result.Diagnostics) > 0 {
		return false, errors.New(strings.Join(result.Diagnostics, "\n"))
	}
	printCacheResult(result, *jsonMode)
	return result.OK, nil
}

func runFixture(rt *nodeRuntime, args []string) (bool, error) {
	flags := flag.NewFlagSet("gojr run-fixture", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	seed := flags.String("seed", "", "deterministic scheduler/random seed; must appear before Node starts")
	randomSeed := flags.String("random-seed", "", "deterministic scheduler/random seed; alias for --seed")
	jsonMode := flags.Bool("json", false, "print a machine-readable JSON fixture report")
	var packageFlags packageFlag
	flags.Var(&packageFlags, "pkg", "Go-junior source package dependency, import/path=DIR; may be repeated")
	var sourceRootFlags packageFlag
	flags.Var(&sourceRootFlags, "srcroot", "filesystem source root for resolving imported Go-junior packages; may be repeated")
	if err := flags.Parse(args); err != nil {
		return false, err
	}
	_ = seed
	_ = randomSeed
	if flags.NArg() != 1 {
		return false, fmt.Errorf("usage: gojr run-fixture [--json] [--pkg import=DIR] [--srcroot DIR] [--seed SEED] FIXTURE.json|-")
	}
	target := flags.Arg(0)
	fixtureJSON, err := readFixtureJSON(target)
	if err != nil {
		return false, err
	}
	packages, err := readRuntimePackageSpecs(packageFlags)
	if err != nil {
		return false, err
	}
	sourceRoots := buildSourceRoots(target, "", sourceRootFlags)
	var result fixtureResult
	if len(packages) > 0 || len(sourceRoots) > 0 {
		result, err = rt.RunFixtureWithPackages(fixtureWithPackagesRequest{
			FixtureJSON: fixtureJSON,
			Packages:    packages,
			SourceRoots: sourceRoots,
		})
	} else {
		result, err = rt.RunFixture(fixtureJSON)
	}
	if err != nil {
		return false, err
	}
	printFixtureResult(result, *jsonMode)
	return result.OK && !result.Unstable, nil
}

func printTopLevelUsage() {
	fmt.Println(`gojr
  start the interactive Go-junior REPL

gojr eval [--json] [--sheet-json JSON] [--pkg import=DIR] [--srcroot DIR] [--seed SEED] SOURCE
  evaluate one Go-junior expression or statement list

gojr run [--json] [--sheet-json JSON] [--pkg import=DIR] [--srcroot DIR] [--seed SEED] [FILE|DIR|-]
  run a Go-junior source file, package directory, or stdin

gojr compile [--json] [--sheet-json JSON] [--pkg import=DIR] [--srcroot DIR] [--expr SOURCE] [FILE|DIR|-]
  parse and typecheck Go-junior source without executing it

gojr test [--json] [--pkg import=DIR] [--srcroot DIR] [--seed SEED] PATH
  run Go-junior tests from a .go file or package directory

gojr build [--json] [-importpath PATH] [--pkg import=DIR] [--srcroot DIR] [-pkgdir DIR|-artifact-root DIR] TARGET
  compile a Go-junior package into the package artifact cache

gojr inspect-js [--json] [-importpath PATH] [--pkg import=DIR] [--srcroot DIR] [-pkgdir DIR|-artifact-root DIR] TARGET
  print generated JavaScript for a Go-junior package without writing the cache

gojr cache path|list|clear [--json] [-pkgdir DIR|-artifact-root DIR] [--yes]
  inspect or clear the Go-junior package artifact cache

gojr run-fixture [--json] [--pkg import=DIR] [--srcroot DIR] [--seed SEED] FIXTURE.json|-
  run a spreadsheet fixture whose formula cells contain Go-junior source

By default build artifacts are written under ~/go/pkg/gojr_js/.
JSON strings are always strings. JSON integers become exact integer values.
JSON numbers with a decimal point or exponent become float64 values.`)
}

func repl(rt *nodeRuntime) error {
	reader := bufio.NewReader(os.Stdin)
	var source strings.Builder

	for {
		prompt := "gojr> "
		if source.Len() > 0 {
			prompt = "....> "
		}
		fmt.Print(prompt)

		line, err := reader.ReadString('\n')
		if err != nil && len(line) == 0 {
			fmt.Println()
			return nil
		}
		line = strings.TrimRight(line, "\r\n")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, ".") {
			done, err := handleCommand(rt, &source, trimmed)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			if done {
				return nil
			}
			continue
		}

		source.WriteString(line)
		source.WriteByte('\n')
		result, err := rt.Eval(source.String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "embedded node error: %v\n", err)
			source.Reset()
			continue
		}
		if result.Incomplete {
			continue
		}
		printResult(result)
		source.Reset()
	}
}

func handleCommand(rt *nodeRuntime, source *strings.Builder, command string) (bool, error) {
	switch {
	case command == ".quit" || command == ".exit":
		return true, nil
	case command == ".help":
		printHelp()
	case command == ".clear":
		source.Reset()
		fmt.Println("cleared")
	case command == ".source":
		return false, fmt.Errorf("usage: .source PATH")
	case strings.HasPrefix(command, ".source "):
		path := strings.TrimSpace(strings.TrimPrefix(command, ".source "))
		file, err := readSourceFile(path)
		if err != nil {
			return false, err
		}
		return false, evalLoadedSourceFiles(rt, source, []sourceFile{file})
	case strings.HasPrefix(command, ".sheet "):
		result, err := rt.SetSheet(strings.TrimSpace(strings.TrimPrefix(command, ".sheet ")))
		if err != nil {
			return false, err
		}
		printResult(result)
	case command == ".load":
		return false, fmt.Errorf("usage: .load DIR")
	case strings.HasPrefix(command, ".load "):
		dir := strings.TrimSpace(strings.TrimPrefix(command, ".load "))
		files, err := readPackageDir(dir)
		if err != nil {
			return false, err
		}
		return false, evalLoadedSourceFiles(rt, source, files)
	case command == ".test":
		return false, fmt.Errorf("usage: .test PATH")
	case strings.HasPrefix(command, ".test "):
		target := strings.TrimSpace(strings.TrimPrefix(command, ".test "))
		files, err := readTestTarget(target)
		if err != nil {
			return false, err
		}
		return false, testLoadedSourceFiles(rt, source, files)
	default:
		return false, fmt.Errorf("unknown command %q", command)
	}
	return false, nil
}

func printHelp() {
	fmt.Println(`Commands:
  .help          show this help
  .clear         clear the accumulated Go-junior source buffer
  .source PATH   load one Go-junior .go source file into the session
  .sheet JSON    replace the current sheet, e.g. .sheet {"A1":40,"B1":2.5}
  .load DIR      load a Go-junior package directory into the session
  .test PATH     run Go-junior tests from a .go file or package directory
  .quit          exit

Normal input is evaluated eagerly. If the parser reaches EOF while expecting
more input, the line is kept as pending multi-line source and the prompt changes
to ....>. Use .clear to discard pending input.`)
}

func evalLoadedSourceFiles(rt *nodeRuntime, pending *strings.Builder, files []sourceFile) error {
	if pending.Len() > 0 {
		return fmt.Errorf("cannot load while multi-line input is pending; use .clear first")
	}
	result, err := rt.EvalFiles(files)
	if err != nil {
		return err
	}
	printLoadResult(result)
	return nil
}

func testLoadedSourceFiles(rt *nodeRuntime, pending *strings.Builder, files []sourceFile) error {
	if pending.Len() > 0 {
		return fmt.Errorf("cannot run tests while multi-line input is pending; use .clear first")
	}
	result, err := rt.TestFiles(files)
	if err != nil {
		return err
	}
	printTestResult(result, false)
	return nil
}

func applySheetJSON(rt *nodeRuntime, sheetJSON string) error {
	if sheetJSON == "" {
		return nil
	}
	result, err := rt.SetSheet(sheetJSON)
	if err != nil {
		return err
	}
	if !result.OK {
		printResult(result)
		return errors.New("sheet JSON was rejected")
	}
	return nil
}

func optionalArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

func readRunSource(path string) (sourceFile, error) {
	target, err := readRunTarget(path)
	if err != nil {
		return sourceFile{}, err
	}
	if target.Package {
		return sourceFile{}, fmt.Errorf("%s is a package main target; use gojr run through runSourceTarget", path)
	}
	return target.Files[0], nil
}

func readRunTarget(path string) (runSourceTarget, error) {
	if path == "" || path == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return runSourceTarget{}, err
		}
		source, err := stripPackageClausePreservingLines(string(data))
		if err != nil {
			return runSourceTarget{}, err
		}
		return runSourceTarget{Files: []sourceFile{{Filename: "stdin.go", Source: source}}}, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return runSourceTarget{}, err
	}
	if info.IsDir() {
		files, err := readBuildTarget(path)
		if err != nil {
			return runSourceTarget{}, err
		}
		packageName := packageNameFromSourceFiles(files)
		return runSourceTarget{
			Files:       files,
			Package:     true,
			ImportPath:  deriveBuildImportPath(path, packageName),
			PackageName: packageName,
		}, nil
	}
	if !strings.HasSuffix(path, ".go") {
		return runSourceTarget{}, fmt.Errorf("%s is not a .go file or directory", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return runSourceTarget{}, err
	}
	if name, _, _, found, err := findPackageClause(string(data)); err != nil {
		return runSourceTarget{}, err
	} else if found && name == "main" {
		file := sourceFile{Filename: filepath.Clean(path), Source: string(data)}
		return runSourceTarget{
			Files:       []sourceFile{file},
			Package:     true,
			ImportPath:  deriveBuildImportPath(path, name),
			PackageName: name,
		}, nil
	}
	source, err := stripPackageClausePreservingLines(string(data))
	if err != nil {
		return runSourceTarget{}, err
	}
	return runSourceTarget{Files: []sourceFile{{Filename: filepath.Clean(path), Source: source}}}, nil
}

func readCompileTarget(path string) ([]sourceFile, error) {
	if path == "" || path == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		return []sourceFile{{Filename: "stdin.go", Source: string(data)}}, nil
	}
	return readBuildTarget(path)
}

func readRuntimePackageSpecs(values []string) ([]runtimePackageSpec, error) {
	var packages []runtimePackageSpec
	for _, value := range values {
		importPath, target, found := strings.Cut(value, "=")
		importPath = strings.TrimSpace(importPath)
		target = strings.TrimSpace(target)
		if !found || importPath == "" || target == "" {
			return nil, fmt.Errorf("--pkg expects import/path=DIR")
		}
		files, err := readBuildTarget(target)
		if err != nil {
			return nil, err
		}
		packages = append(packages, runtimePackageSpec{
			ImportPath:  importPath,
			PackageName: packageNameFromSourceFiles(files),
			Files:       files,
		})
	}
	return packages, nil
}

func packageSourceMap(specs []runtimePackageSpec) packageSources {
	if len(specs) == 0 {
		return nil
	}
	sources := make(packageSources, len(specs))
	for _, spec := range specs {
		sources[spec.ImportPath] = spec.Files
	}
	return sources
}

func buildSourceRoots(target string, importPath string, explicitRoots []string) []string {
	if len(explicitRoots) == 0 && strings.TrimSpace(importPath) == "" && buildTargetDir(target) == "" {
		return nil
	}
	var roots []string
	seen := map[string]bool{}
	add := func(path string) {
		path = strings.TrimSpace(expandHome(path))
		if path == "" {
			return
		}
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		path = filepath.Clean(path)
		if seen[path] {
			return
		}
		seen[path] = true
		roots = append(roots, path)
	}

	for _, root := range explicitRoots {
		add(root)
	}

	if packageDir := buildTargetDir(target); packageDir != "" {
		if moduleRoot, _ := moduleRootAndPathForDir(packageDir); moduleRoot != "" {
			add(moduleRoot)
		}
		if root := sourceRootFromImportPath(packageDir, importPath); root != "" {
			add(root)
		}
	}

	for _, gopath := range candidateGOPATHs() {
		add(filepath.Join(gopath, "src"))
	}
	return roots
}

func buildTargetDir(target string) string {
	if target == "" || target == "-" {
		return ""
	}
	info, err := os.Stat(target)
	if err != nil {
		return ""
	}
	dir := target
	if !info.IsDir() {
		dir = filepath.Dir(target)
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return filepath.Clean(dir)
}

func sourceRootFromImportPath(packageDir string, importPath string) string {
	importPath = strings.Trim(filepath.ToSlash(strings.TrimSpace(importPath)), "/")
	if importPath == "" {
		return ""
	}
	parts := strings.Split(importPath, "/")
	candidate := filepath.Clean(packageDir)
	for index := len(parts) - 1; index >= 0; index-- {
		if filepath.Base(candidate) != parts[index] {
			return ""
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			return ""
		}
		candidate = parent
	}
	return candidate
}

func readFixtureJSON(path string) (string, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readSourceFile(path string) (sourceFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return sourceFile{}, err
	}
	if info.IsDir() {
		return sourceFile{}, fmt.Errorf("%s is a directory; use .load DIR for packages", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return sourceFile{}, err
	}
	source, err := stripPackageClausePreservingLines(string(data))
	if err != nil {
		return sourceFile{}, err
	}
	return sourceFile{Filename: filepath.Clean(path), Source: source}, nil
}

func readPackageDir(dir string) ([]sourceFile, error) {
	return readPackageDirMatching(dir, func(name string) bool {
		return !strings.HasSuffix(name, "_test.go")
	}, "non-test .go files")
}

func readTestTarget(target string) ([]sourceFile, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return readPackageDirMatching(target, func(string) bool {
			return true
		}, ".go files")
	}
	if !strings.HasSuffix(target, ".go") {
		return nil, fmt.Errorf("%s is not a .go file or directory", target)
	}
	dir := filepath.Dir(target)
	base := filepath.Base(target)
	if strings.HasSuffix(base, "_test.go") {
		return readPackageDirMatching(dir, func(name string) bool {
			return !strings.HasSuffix(name, "_test.go") || name == base
		}, ".go files")
	}
	return readPackageDirMatching(dir, func(string) bool {
		return true
	}, ".go files")
}

func readBuildTarget(target string) ([]sourceFile, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return readPackageDirRaw(target, func(name string) bool {
			return !strings.HasSuffix(name, "_test.go")
		}, "non-test .go files")
	}
	if !strings.HasSuffix(target, ".go") {
		return nil, fmt.Errorf("%s is not a .go file or directory", target)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return nil, err
	}
	return []sourceFile{{Filename: filepath.Clean(target), Source: string(data)}}, nil
}

func seedFromArgs(args []string) (string, bool, error) {
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--" {
			return "", false, nil
		}
		if arg == "--seed" || arg == "--random-seed" {
			if index+1 >= len(args) {
				return "", false, fmt.Errorf("%s expects a seed value", arg)
			}
			return args[index+1], true, nil
		}
		for _, prefix := range []string{"--seed=", "--random-seed="} {
			if strings.HasPrefix(arg, prefix) {
				return strings.TrimPrefix(arg, prefix), true, nil
			}
		}
	}
	return "", false, nil
}

func readPackageDirRaw(dir string, include func(name string) bool, emptyDescription string) ([]sourceFile, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".go") || !include(name) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return nil, fmt.Errorf("%s contains no %s", dir, emptyDescription)
	}

	files := make([]sourceFile, 0, len(names))
	for _, name := range names {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		files = append(files, sourceFile{Filename: filepath.Clean(path), Source: string(data)})
	}
	return files, nil
}

func packageNameFromSourceFiles(files []sourceFile) string {
	for _, file := range files {
		name, _, _, found, err := findPackageClause(file.Source)
		if err == nil && found && name != "" {
			return name
		}
	}
	return ""
}

func deriveBuildImportPath(target string, packageName string) string {
	packagePath := target
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		packagePath = filepath.Dir(target)
	}
	if abs, err := filepath.Abs(packagePath); err == nil {
		packagePath = abs
	}
	if moduleRoot, modulePath := moduleRootAndPathForDir(packagePath); moduleRoot != "" && modulePath != "" {
		if rel, err := filepath.Rel(moduleRoot, packagePath); err == nil {
			if rel == "." {
				return modulePath
			}
			if !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." {
				return strings.TrimSuffix(modulePath+"/"+filepath.ToSlash(rel), "/")
			}
		}
	}
	for _, gopath := range candidateGOPATHs() {
		srcRoot := filepath.Join(gopath, "src")
		rel, err := filepath.Rel(srcRoot, packagePath)
		if err == nil && rel != "." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." {
			return filepath.ToSlash(rel)
		}
	}
	if packageName != "" {
		return packageName
	}
	base := filepath.Base(packagePath)
	return strings.TrimSuffix(base, ".go")
}

func moduleRootAndPathForDir(dir string) (string, string) {
	if dir == "" {
		return "", ""
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	dir = filepath.Clean(dir)
	for {
		goMod := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(goMod)
		if err == nil {
			if modulePath := parseGoModModule(string(data)); modulePath != "" {
				return dir, modulePath
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ""
		}
		dir = parent
	}
}

func parseGoModModule(source string) string {
	for _, line := range strings.Split(source, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1]
		}
		return ""
	}
	return ""
}

func expandHome(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func candidateGOPATHs() []string {
	var paths []string
	if env := os.Getenv("GOPATH"); env != "" {
		paths = append(paths, filepath.SplitList(env)...)
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		paths = append(paths, filepath.Join(home, "go"))
	}
	out := paths[:0]
	seen := map[string]bool{}
	for _, path := range paths {
		if path == "" {
			continue
		}
		clean := filepath.Clean(path)
		if seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	return out
}

func readPackageDirMatching(dir string, include func(name string) bool, emptyDescription string) ([]sourceFile, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory; use .source PATH for a single file", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".go") || !include(name) {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("%s contains no %s", dir, emptyDescription)
	}

	var packageName string
	var out []sourceFile
	for _, name := range files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		source, found, nextPackageName, err := stripPackageClausePreservingLinesWithName(string(data))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if !found {
			return nil, fmt.Errorf("%s: missing package clause", path)
		}
		if packageName == "" {
			packageName = nextPackageName
		} else if nextPackageName != packageName {
			return nil, fmt.Errorf("%s: package %s does not match package %s", path, nextPackageName, packageName)
		}
		out = append(out, sourceFile{Filename: filepath.Clean(path), Source: source})
	}
	return out, nil
}

func stripPackageClausePreservingLines(source string) (string, error) {
	stripped, _, _, err := stripPackageClausePreservingLinesWithName(source)
	return stripped, err
}

func stripPackageClausePreservingLinesWithName(source string) (string, bool, string, error) {
	name, start, end, found, err := findPackageClause(source)
	if err != nil || !found {
		return source, found, name, err
	}
	var out strings.Builder
	out.Grow(len(source))
	out.WriteString(source[:start])
	out.WriteString(strings.Repeat(" ", end-start))
	out.WriteString(source[end:])
	return out.String(), true, name, nil
}

func findPackageClause(source string) (name string, start int, end int, found bool, err error) {
	i := 0
	if strings.HasPrefix(source, "\ufeff") {
		i = len("\ufeff")
	}
	for {
		i = skipSpace(source, i)
		if strings.HasPrefix(source[i:], "//") {
			i = skipLine(source, i+2)
			continue
		}
		if strings.HasPrefix(source[i:], "/*") {
			next := strings.Index(source[i+2:], "*/")
			if next < 0 {
				return "", 0, 0, false, nil
			}
			i += 2 + next + 2
			continue
		}
		break
	}
	if !strings.HasPrefix(source[i:], "package") || isIdentPart(byteAt(source, i+len("package"))) {
		return "", 0, 0, false, nil
	}
	start = i
	i += len("package")
	if !isSpace(byteAt(source, i)) {
		return "", 0, 0, true, fmt.Errorf("malformed package clause")
	}
	i = skipSpace(source, i)
	if !isIdentStart(byteAt(source, i)) {
		return "", 0, 0, true, fmt.Errorf("missing package name")
	}
	nameStart := i
	i++
	for isIdentPart(byteAt(source, i)) {
		i++
	}
	name = source[nameStart:i]
	end = skipLine(source, i)
	return name, start, end, true, nil
}

func skipSpace(source string, offset int) int {
	for offset < len(source) && isSpace(source[offset]) {
		offset++
	}
	return offset
}

func skipLine(source string, offset int) int {
	for offset < len(source) && source[offset] != '\n' {
		offset++
	}
	return offset
}

func byteAt(source string, offset int) byte {
	if offset < 0 || offset >= len(source) {
		return 0
	}
	return source[offset]
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isIdentStart(ch byte) bool {
	return ch == '_' || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}

func printResult(result evalResult) {
	for _, diagnostic := range result.Diagnostics {
		if result.OK {
			fmt.Println(diagnostic)
		} else {
			fmt.Fprintln(os.Stderr, diagnostic)
		}
	}
	if result.Output != "" {
		fmt.Print(result.Output)
	}
	if shouldPrintValue(result) {
		fmt.Println(result.Value)
	}
}

func printLoadResult(result evalResult) {
	for _, diagnostic := range result.Diagnostics {
		if result.OK {
			fmt.Println(diagnostic)
		} else {
			fmt.Fprintln(os.Stderr, diagnostic)
		}
	}
	if result.Output != "" {
		fmt.Print(result.Output)
	}
}

func printEvalResult(result evalResult, jsonMode bool) {
	if jsonMode {
		printJSON(result)
		return
	}
	printResult(result)
}

func printTestResult(result evalResult, jsonMode bool) {
	if jsonMode {
		printJSON(result)
		return
	}
	if result.Output != "" {
		fmt.Print(result.Output)
	}
	for _, diagnostic := range result.Diagnostics {
		if result.OK {
			fmt.Println(diagnostic)
		} else {
			fmt.Fprintln(os.Stderr, diagnostic)
		}
	}
}

func printCompileResult(result compileResult, jsonMode bool) {
	if jsonMode {
		printJSON(result)
		return
	}
	if result.Output != "" {
		fmt.Print(result.Output)
	}
	for _, diagnostic := range result.Diagnostics {
		if result.OK {
			fmt.Println(diagnostic)
		} else {
			fmt.Fprintln(os.Stderr, diagnostic)
		}
	}
	if result.OK && len(result.Diagnostics) == 0 {
		fmt.Println("ok")
	}
}

func printBuildResult(result buildResult, jsonMode bool) {
	if jsonMode {
		printJSON(result)
		return
	}
	if result.Output != "" {
		fmt.Print(result.Output)
	}
	for _, diagnostic := range result.Diagnostics {
		if result.OK {
			fmt.Println(diagnostic)
		} else {
			fmt.Fprintln(os.Stderr, diagnostic)
		}
	}
	for _, artifact := range result.Artifacts {
		if artifact.Action == "" {
			artifact.Action = "built"
		}
		fmt.Printf("%s %s\n", artifact.Action, artifact.ArtifactPath)
	}
}

func printInspectJSResult(result inspectJSResult, jsonMode bool) {
	if jsonMode {
		printJSON(result)
		return
	}
	for _, diagnostic := range result.Diagnostics {
		if result.OK {
			fmt.Println(diagnostic)
		} else {
			fmt.Fprintln(os.Stderr, diagnostic)
		}
	}
	if result.Output != "" {
		fmt.Print(result.Output)
	}
	if result.Source != "" {
		fmt.Print(result.Source)
		if !strings.HasSuffix(result.Source, "\n") {
			fmt.Println()
		}
	}
}

func printCacheResult(result cacheResult, jsonMode bool) {
	if jsonMode {
		printJSON(result)
		return
	}
	for _, diagnostic := range result.Diagnostics {
		if result.OK {
			fmt.Println(diagnostic)
		} else {
			fmt.Fprintln(os.Stderr, diagnostic)
		}
	}
	if len(result.Entries) > 0 {
		for _, entry := range result.Entries {
			if entry.CacheKey != "" {
				fmt.Printf("%s %s %d %s\n", entry.ImportPath, entry.CacheKey, entry.Size, entry.Path)
			} else {
				fmt.Printf("%s %d %s\n", entry.ImportPath, entry.Size, entry.Path)
			}
		}
		return
	}
	if len(result.Cleared) > 0 {
		for _, path := range result.Cleared {
			fmt.Printf("removed %s\n", path)
		}
		return
	}
	if result.Action == "path" || result.Action == "clear" {
		fmt.Println(result.Root)
	}
}

func printFixtureResult(result fixtureResult, jsonMode bool) {
	if jsonMode {
		printJSON(result)
		return
	}
	for _, diagnostic := range result.Diagnostics {
		if result.OK {
			fmt.Println(diagnostic)
		} else {
			fmt.Fprintln(os.Stderr, diagnostic)
		}
	}
	for _, sheetName := range sortedStringKeys(result.Sheets) {
		cells := result.Sheets[sheetName]
		for _, cell := range sortedStringKeys(cells) {
			fmt.Printf("%s!%s = %s\n", sheetName, cell, cells[cell])
		}
	}
}

func sortedStringKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func printJSON(value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func shouldPrintValue(result evalResult) bool {
	return result.Value != "" && !result.ValueIsNil
}

func newNodeRuntime(bootstrapSource string, moduleBundleJSON string) (*nodeRuntime, error) {
	cBootstrapSource := C.CString(bootstrapSource)
	defer C.free(unsafe.Pointer(cBootstrapSource))
	cModuleBundle := C.CString(moduleBundleJSON)
	defer C.free(unsafe.Pointer(cModuleBundle))

	var cErr *C.char
	ptr := C.gojr_node_new(cBootstrapSource, cModuleBundle, &cErr)
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return nil, errors.New(C.GoString(cErr))
	}
	if ptr == nil {
		return nil, errors.New("Node runtime initialization returned nil")
	}
	return &nodeRuntime{ptr: ptr}, nil
}

func (rt *nodeRuntime) Eval(source string) (evalResult, error) {
	return rt.call(source, nodeCallEval)
}

func (rt *nodeRuntime) EvalFiles(files []sourceFile) (evalResult, error) {
	data, err := json.Marshal(files)
	if err != nil {
		return evalResult{}, err
	}
	return rt.call(string(data), nodeCallEvalFiles)
}

func (rt *nodeRuntime) EvalWithPackages(request evalWithPackagesRequest) (evalResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return evalResult{}, err
	}
	cInput := C.CString(string(data))
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	cResult := C.gojr_node_eval_with_packages(rt.ptr, cInput, &cErr)
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return evalResult{}, errors.New(C.GoString(cErr))
	}
	if cResult == nil {
		return evalResult{}, errors.New("embedded Node package eval returned nil")
	}
	defer C.gojr_string_free(cResult)

	var result evalResult
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return evalResult{}, err
	}
	return result, nil
}

func (rt *nodeRuntime) EvalFilesWithPackages(request evalWithPackagesRequest) (evalResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return evalResult{}, err
	}
	cInput := C.CString(string(data))
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	cResult := C.gojr_node_eval_files_with_packages(rt.ptr, cInput, &cErr)
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return evalResult{}, errors.New(C.GoString(cErr))
	}
	if cResult == nil {
		return evalResult{}, errors.New("embedded Node package file eval returned nil")
	}
	defer C.gojr_string_free(cResult)

	var result evalResult
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return evalResult{}, err
	}
	return result, nil
}

func (rt *nodeRuntime) RunMainFilesWithPackages(request evalWithPackagesRequest) (evalResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return evalResult{}, err
	}
	cInput := C.CString(string(data))
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	cResult := C.gojr_node_run_main_files_with_packages(rt.ptr, cInput, &cErr)
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return evalResult{}, errors.New(C.GoString(cErr))
	}
	if cResult == nil {
		return evalResult{}, errors.New("embedded Node package main run returned nil")
	}
	defer C.gojr_string_free(cResult)

	var result evalResult
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return evalResult{}, err
	}
	return result, nil
}

func (rt *nodeRuntime) SetSheet(json string) (evalResult, error) {
	return rt.call(json, nodeCallSetSheet)
}

func (rt *nodeRuntime) Test(source string) (evalResult, error) {
	return rt.call(source, nodeCallTest)
}

func (rt *nodeRuntime) TestFiles(files []sourceFile) (evalResult, error) {
	data, err := json.Marshal(files)
	if err != nil {
		return evalResult{}, err
	}
	return rt.call(string(data), nodeCallTestFiles)
}

func (rt *nodeRuntime) TestFilesWithPackages(request evalWithPackagesRequest) (evalResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return evalResult{}, err
	}
	cInput := C.CString(string(data))
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	cResult := C.gojr_node_test_files_with_packages(rt.ptr, cInput, &cErr)
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return evalResult{}, errors.New(C.GoString(cErr))
	}
	if cResult == nil {
		return evalResult{}, errors.New("embedded Node package test returned nil")
	}
	defer C.gojr_string_free(cResult)

	var result evalResult
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return evalResult{}, err
	}
	return result, nil
}

func (rt *nodeRuntime) Compile(request compileRequest) (compileResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return compileResult{}, err
	}
	cInput := C.CString(string(data))
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	cResult := C.gojr_node_compile(rt.ptr, cInput, &cErr)
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return compileResult{}, errors.New(C.GoString(cErr))
	}
	if cResult == nil {
		return compileResult{}, errors.New("embedded Node compile returned nil")
	}
	defer C.gojr_string_free(cResult)

	var result compileResult
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return compileResult{}, err
	}
	return result, nil
}

func (rt *nodeRuntime) Build(request buildRequest) (buildResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return buildResult{}, err
	}
	cInput := C.CString(string(data))
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	cResult := C.gojr_node_build(rt.ptr, cInput, &cErr)
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return buildResult{}, errors.New(C.GoString(cErr))
	}
	if cResult == nil {
		return buildResult{}, errors.New("embedded Node build returned nil")
	}
	defer C.gojr_string_free(cResult)

	var result buildResult
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return buildResult{}, err
	}
	return result, nil
}

func (rt *nodeRuntime) InspectJS(request buildRequest) (inspectJSResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return inspectJSResult{}, err
	}
	cInput := C.CString(string(data))
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	cResult := C.gojr_node_inspect_js(rt.ptr, cInput, &cErr)
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return inspectJSResult{}, errors.New(C.GoString(cErr))
	}
	if cResult == nil {
		return inspectJSResult{}, errors.New("embedded Node inspect-js returned nil")
	}
	defer C.gojr_string_free(cResult)

	var result inspectJSResult
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return inspectJSResult{}, err
	}
	return result, nil
}

func (rt *nodeRuntime) Cache(request cacheRequest) (cacheResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return cacheResult{}, err
	}
	cInput := C.CString(string(data))
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	cResult := C.gojr_node_cache(rt.ptr, cInput, &cErr)
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return cacheResult{}, errors.New(C.GoString(cErr))
	}
	if cResult == nil {
		return cacheResult{}, errors.New("embedded Node cache returned nil")
	}
	defer C.gojr_string_free(cResult)

	var result cacheResult
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return cacheResult{}, err
	}
	return result, nil
}

func (rt *nodeRuntime) RunFixture(fixtureJSON string) (fixtureResult, error) {
	cInput := C.CString(fixtureJSON)
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	cResult := C.gojr_node_run_fixture(rt.ptr, cInput, &cErr)
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return fixtureResult{}, errors.New(C.GoString(cErr))
	}
	if cResult == nil {
		return fixtureResult{}, errors.New("embedded Node fixture run returned nil")
	}
	defer C.gojr_string_free(cResult)

	var result fixtureResult
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return fixtureResult{}, err
	}
	return result, nil
}

func (rt *nodeRuntime) RunFixtureWithPackages(request fixtureWithPackagesRequest) (fixtureResult, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return fixtureResult{}, err
	}
	cInput := C.CString(string(data))
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	cResult := C.gojr_node_run_fixture_with_packages(rt.ptr, cInput, &cErr)
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return fixtureResult{}, errors.New(C.GoString(cErr))
	}
	if cResult == nil {
		return fixtureResult{}, errors.New("embedded Node package fixture run returned nil")
	}
	defer C.gojr_string_free(cResult)

	var result fixtureResult
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return fixtureResult{}, err
	}
	return result, nil
}

func (rt *nodeRuntime) call(input string, mode nodeCallMode) (evalResult, error) {
	cInput := C.CString(input)
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	var cResult *C.char
	switch mode {
	case nodeCallEvalFiles:
		cResult = C.gojr_node_eval_files(rt.ptr, cInput, &cErr)
	case nodeCallSetSheet:
		cResult = C.gojr_node_set_sheet(rt.ptr, cInput, &cErr)
	case nodeCallTest:
		cResult = C.gojr_node_test(rt.ptr, cInput, &cErr)
	case nodeCallTestFiles:
		cResult = C.gojr_node_test_files(rt.ptr, cInput, &cErr)
	case nodeCallBuild:
		cResult = C.gojr_node_build(rt.ptr, cInput, &cErr)
	default:
		cResult = C.gojr_node_eval(rt.ptr, cInput, &cErr)
	}
	if cErr != nil {
		defer C.gojr_string_free(cErr)
		return evalResult{}, errors.New(C.GoString(cErr))
	}
	if cResult == nil {
		return evalResult{}, errors.New("embedded Node call returned nil")
	}
	defer C.gojr_string_free(cResult)

	var result evalResult
	if err := json.Unmarshal([]byte(C.GoString(cResult)), &result); err != nil {
		return evalResult{}, err
	}
	return result, nil
}

func (rt *nodeRuntime) Close() {
	if rt.ptr != nil {
		C.gojr_node_free(rt.ptr)
		rt.ptr = nil
	}
}

func runtimeModuleBundle() (string, error) {
	var modules []embeddedModule
	if err := fs.WalkDir(gojr.EmbeddedDist, "dist", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".js") || path == "dist/src/embeddedNodeBootstrap.js" {
			return nil
		}
		data, err := gojr.EmbeddedDist.ReadFile(path)
		if err != nil {
			return err
		}
		modules = append(modules, embeddedModule{
			Path:   strings.TrimPrefix(path, "dist"),
			Source: string(data),
		})
		return nil
	}); err != nil {
		return "", err
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Path < modules[j].Path
	})
	if len(modules) == 0 {
		return "", errors.New("embedded Go-junior dist is empty; run `make` from ~/ivy/gojr to rebuild")
	}
	bundle, err := json.Marshal(embeddedModuleBundle{Modules: modules})
	if err != nil {
		return "", err
	}
	return string(bundle), nil
}

func runtimeBootstrapSource() (string, error) {
	data, err := gojr.EmbeddedDist.ReadFile("dist/src/embeddedNodeBootstrap.js")
	if err != nil {
		return "", fmt.Errorf("embedded Go-junior bootstrap not found; run `make` from ~/ivy/gojr to rebuild: %w", err)
	}
	if len(data) == 0 {
		return "", errors.New("embedded Go-junior bootstrap is empty; run `make` from ~/ivy/gojr to rebuild")
	}
	return string(data), nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "gojr: %v\n", err)
	os.Exit(1)
}

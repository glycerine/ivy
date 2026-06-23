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
	OK          bool     `json:"ok"`
	Incomplete  bool     `json:"incomplete"`
	Diagnostics []string `json:"diagnostics"`
	Output      string   `json:"output"`
	Value       string   `json:"value"`
	ValueIsNil  bool     `json:"valueIsNil"`
}

type buildRequest struct {
	ImportPath         string       `json:"importPath,omitempty"`
	Files              []sourceFile `json:"files"`
	ArtifactRoot       string       `json:"artifactRoot,omitempty"`
	PackageCacheParent string       `json:"packageCacheParent,omitempty"`
}

type buildResult struct {
	OK          bool            `json:"ok"`
	Diagnostics []string        `json:"diagnostics"`
	Output      string          `json:"output"`
	Artifacts   []buildArtifact `json:"artifacts"`
	Built       []string        `json:"built"`
	Skipped     []string        `json:"skipped"`
}

type buildArtifact struct {
	ImportPath   string `json:"importPath"`
	PackageName  string `json:"packageName"`
	ArtifactPath string `json:"artifactPath"`
	Action       string `json:"action"`
	SourceHash   string `json:"sourceHash"`
	CacheKey     string `json:"cacheKey"`
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

func main() {
	moduleBundle, err := runtimeModuleBundle()
	if err != nil {
		fatal(err)
	}

	rt, err := newNodeRuntime(moduleBundle)
	if err != nil {
		fatal(err)
	}
	defer rt.Close()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "build":
			ok, err := runBuild(rt, os.Args[2:])
			if err != nil {
				fatal(err)
			}
			if !ok {
				os.Exit(1)
			}
			return
		case "help", "-h", "--help":
			printTopLevelUsage()
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

func runBuild(rt *nodeRuntime, args []string) (bool, error) {
	flags := flag.NewFlagSet("gojr build", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	importPath := flags.String("importpath", "", "package import path for the generated artifact")
	packageCacheParent := flags.String("pkgdir", "", "package-cache parent directory; gojr_js is appended")
	artifactRoot := flags.String("artifact-root", "", "exact gojr_js artifact root directory")
	if err := flags.Parse(args); err != nil {
		return false, err
	}
	if flags.NArg() != 1 {
		return false, fmt.Errorf("usage: gojr build [-importpath PATH] [-pkgdir DIR|-artifact-root DIR] TARGET")
	}
	target := flags.Arg(0)
	files, err := readBuildTarget(target)
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
		ArtifactRoot:       strings.TrimSpace(*artifactRoot),
		PackageCacheParent: strings.TrimSpace(*packageCacheParent),
	})
	if err != nil {
		return false, err
	}
	printBuildResult(result)
	return result.OK, nil
}

func printTopLevelUsage() {
	fmt.Println(`gojr
  start the interactive Go-junior REPL

gojr build [-importpath PATH] [-pkgdir DIR|-artifact-root DIR] TARGET
  compile a Go-junior package into the package artifact cache

By default build artifacts are written under ~/go/pkg/gojr_js/.`)
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
	printTestResult(result)
	return nil
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

func printTestResult(result evalResult) {
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

func printBuildResult(result buildResult) {
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

func shouldPrintValue(result evalResult) bool {
	return result.Value != "" && !result.ValueIsNil
}

func newNodeRuntime(moduleBundleJSON string) (*nodeRuntime, error) {
	cModuleBundle := C.CString(moduleBundleJSON)
	defer C.free(unsafe.Pointer(cModuleBundle))

	var cErr *C.char
	ptr := C.gojr_node_new(cModuleBundle, &cErr)
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
		if entry.IsDir() || !strings.HasSuffix(path, ".js") {
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

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "gojr: %v\n", err)
	os.Exit(1)
}

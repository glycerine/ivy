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
	nodeCallSetSheet
	nodeCallTest
)

type evalResult struct {
	OK          bool     `json:"ok"`
	Incomplete  bool     `json:"incomplete"`
	Diagnostics []string `json:"diagnostics"`
	Output      string   `json:"output"`
	Value       string   `json:"value"`
	ValueIsNil  bool     `json:"valueIsNil"`
}

type embeddedModuleBundle struct {
	Modules []embeddedModule `json:"modules"`
}

type embeddedModule struct {
	Path   string `json:"path"`
	Source string `json:"source"`
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

	fmt.Printf("Go-junior REPL (embedded Node/V8)\n")
	fmt.Printf("runtime: embedded dist/src/index.js\n")
	fmt.Printf("commands: .help .clear .source PATH .sheet JSON .load DIR .test PATH .quit\n")

	if err := repl(rt); err != nil {
		fatal(err)
	}
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
		data, err := readSourceFile(path)
		if err != nil {
			return false, err
		}
		return false, evalLoadedSource(rt, source, data)
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
		data, err := readPackageDir(dir)
		if err != nil {
			return false, err
		}
		return false, evalLoadedSource(rt, source, data)
	case command == ".test":
		return false, fmt.Errorf("usage: .test PATH")
	case strings.HasPrefix(command, ".test "):
		target := strings.TrimSpace(strings.TrimPrefix(command, ".test "))
		data, err := readTestTarget(target)
		if err != nil {
			return false, err
		}
		return false, testLoadedSource(rt, source, data)
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

func evalLoadedSource(rt *nodeRuntime, pending *strings.Builder, data string) error {
	if pending.Len() > 0 {
		return fmt.Errorf("cannot load while multi-line input is pending; use .clear first")
	}
	result, err := rt.Eval(data)
	if err != nil {
		return err
	}
	printLoadResult(result)
	return nil
}

func testLoadedSource(rt *nodeRuntime, pending *strings.Builder, data string) error {
	if pending.Len() > 0 {
		return fmt.Errorf("cannot run tests while multi-line input is pending; use .clear first")
	}
	result, err := rt.Test(data)
	if err != nil {
		return err
	}
	printTestResult(result)
	return nil
}

func readSourceFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory; use .load DIR for packages", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return stripPackageClausePreservingLines(string(data))
}

func readPackageDir(dir string) (string, error) {
	return readPackageDirMatching(dir, func(name string) bool {
		return !strings.HasSuffix(name, "_test.go")
	}, "non-test .go files")
}

func readTestTarget(target string) (string, error) {
	info, err := os.Stat(target)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return readPackageDirMatching(target, func(string) bool {
			return true
		}, ".go files")
	}
	if !strings.HasSuffix(target, ".go") {
		return "", fmt.Errorf("%s is not a .go file or directory", target)
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

func readPackageDirMatching(dir string, include func(name string) bool, emptyDescription string) (string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory; use .source PATH for a single file", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("%s contains no %s", dir, emptyDescription)
	}

	var packageName string
	var combined strings.Builder
	for _, name := range files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		source, found, nextPackageName, err := stripPackageClausePreservingLinesWithName(string(data))
		if err != nil {
			return "", fmt.Errorf("%s: %w", path, err)
		}
		if !found {
			return "", fmt.Errorf("%s: missing package clause", path)
		}
		if packageName == "" {
			packageName = nextPackageName
		} else if nextPackageName != packageName {
			return "", fmt.Errorf("%s: package %s does not match package %s", path, nextPackageName, packageName)
		}
		combined.WriteString("\n// ---- ")
		combined.WriteString(name)
		combined.WriteString(" ----\n")
		combined.WriteString(source)
		if !strings.HasSuffix(source, "\n") {
			combined.WriteByte('\n')
		}
	}
	return combined.String(), nil
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

func (rt *nodeRuntime) SetSheet(json string) (evalResult, error) {
	return rt.call(json, nodeCallSetSheet)
}

func (rt *nodeRuntime) Test(source string) (evalResult, error) {
	return rt.call(source, nodeCallTest)
}

func (rt *nodeRuntime) call(input string, mode nodeCallMode) (evalResult, error) {
	cInput := C.CString(input)
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	var cResult *C.char
	switch mode {
	case nodeCallSetSheet:
		cResult = C.gojr_node_set_sheet(rt.ptr, cInput, &cErr)
	case nodeCallTest:
		cResult = C.gojr_node_test(rt.ptr, cInput, &cErr)
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

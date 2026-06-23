package main

/*
#cgo darwin CXXFLAGS: -std=c++20 -I/usr/local/include/node -DNODE_SHARED_MODE
#cgo darwin LDFLAGS: -L/usr/local/lib -lnode.141 -Wl,-rpath,/usr/local/lib
#include <stdlib.h>
#include "node_bridge.h"
*/
import "C"

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

type nodeRuntime struct {
	ptr *C.gojr_node_runtime
}

type evalResult struct {
	OK          bool     `json:"ok"`
	Incomplete  bool     `json:"incomplete"`
	Diagnostics []string `json:"diagnostics"`
	Output      string   `json:"output"`
	Value       string   `json:"value"`
}

func main() {
	modulePath, err := findRuntimeModule()
	if err != nil {
		fatal(err)
	}

	rt, err := newNodeRuntime(modulePath)
	if err != nil {
		fatal(err)
	}
	defer rt.Close()

	fmt.Printf("Go-junior REPL (embedded Node/V8)\n")
	fmt.Printf("runtime: %s\n", modulePath)
	fmt.Printf("commands: .help .clear .source .sheet JSON .load PATH .quit\n")

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
		if source.Len() == 0 {
			fmt.Println("(empty)")
			return false, nil
		}
		fmt.Print(source.String())
	case strings.HasPrefix(command, ".sheet "):
		result, err := rt.SetSheet(strings.TrimSpace(strings.TrimPrefix(command, ".sheet ")))
		if err != nil {
			return false, err
		}
		printResult(result)
	case strings.HasPrefix(command, ".load "):
		path := strings.TrimSpace(strings.TrimPrefix(command, ".load "))
		data, err := os.ReadFile(path)
		if err != nil {
			return false, err
		}
		source.Write(data)
		if !strings.HasSuffix(source.String(), "\n") {
			source.WriteByte('\n')
		}
		result, err := rt.Eval(source.String())
		if err != nil {
			return false, err
		}
		printResult(result)
		if !result.Incomplete {
			source.Reset()
		}
	default:
		return false, fmt.Errorf("unknown command %q", command)
	}
	return false, nil
}

func printHelp() {
	fmt.Println(`Commands:
  .help          show this help
  .clear         clear the accumulated Go-junior source buffer
  .source        print the pending multi-line source buffer
  .sheet JSON    replace the current sheet, e.g. .sheet {"A1":40,"B1":2.5}
  .load PATH     append a Go-junior source file and evaluate the buffer
  .quit          exit

Normal input is evaluated eagerly. If the parser reaches EOF while expecting
more input, the line is kept as pending multi-line source and the prompt changes
to ....>. Use .clear to discard pending input.`)
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
	if result.Value != "" {
		fmt.Println(result.Value)
	}
}

func newNodeRuntime(modulePath string) (*nodeRuntime, error) {
	cModulePath := C.CString(modulePath)
	defer C.free(unsafe.Pointer(cModulePath))

	var cErr *C.char
	ptr := C.gojr_node_new(cModulePath, &cErr)
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
	return rt.call(source, false)
}

func (rt *nodeRuntime) SetSheet(json string) (evalResult, error) {
	return rt.call(json, true)
}

func (rt *nodeRuntime) call(input string, setSheet bool) (evalResult, error) {
	cInput := C.CString(input)
	defer C.free(unsafe.Pointer(cInput))

	var cErr *C.char
	var cResult *C.char
	if setSheet {
		cResult = C.gojr_node_set_sheet(rt.ptr, cInput, &cErr)
	} else {
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

func findRuntimeModule() (string, error) {
	if explicit := os.Getenv("GOJR_NODE_MODULE"); explicit != "" {
		return filepath.Abs(explicit)
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	candidates := []string{
		filepath.Join(wd, "dist", "src", "index.js"),
		filepath.Join(wd, "gojr", "dist", "src", "index.js"),
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "dist", "src", "index.js"),
			filepath.Join(exeDir, "..", "..", "dist", "src", "index.js"),
			filepath.Join(exeDir, "..", "..", "..", "dist", "src", "index.js"),
		)
	}

	for _, candidate := range candidates {
		abs, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
	}

	return "", errors.New("could not find gojr/dist/src/index.js; run `npm run build` in ~/ivy/gojr or set GOJR_NODE_MODULE")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "gojr: %v\n", err)
	os.Exit(1)
}

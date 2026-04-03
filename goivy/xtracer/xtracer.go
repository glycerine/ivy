//go:build !xtracer_off

// Package xtracer provides execution tracing for parallel conformance
// testing between Go goivy and Python ivy. Trace() prints timestamped
// messages to stdout by default, matching Python's `if __debug__:` guards.
//
// To disable at build time: go build -tags xtracer_off ./cmd/goivy_check/
// To disable at runtime:    XTRACE_OFF=1 go test ./...
//
// Both Go and Python emit the same format:
//
//	XTRACE: <category>.<function> <ENTER|EXIT> [<detail>]
//
// so that `diff` on the two outputs reveals the first divergence.
package xtracer

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// Enabled is true when xtracer is active (the default).
// Set XTRACE_OFF=1 to suppress output at runtime.
const Enabled = true

// HashVerbose causes HASH trace lines to include
// the full canonical string.
const HashVerbose bool = true

// suppressed is set by XTRACE_OFF=1 to silence output without
// changing Enabled (so `if xtracer.Enabled` guards still compile away).
var suppressed bool

// base directory of this (ivy) Go project
var repo string
var gopath string
var home string
var ivyIncludeDirs []string
var ivyExamplesDir []string

func init() {
	if os.Getenv("XTRACE_OFF") == "1" {
		suppressed = true
	}
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to get current .go test file path")
	}
	dir := filepath.Dir(filename) // ~/ivy/goivy/xtracer/
	dir = filepath.Dir(dir)       // ~/ivy/goivy
	repo = filepath.Dir(dir)      // ~/ivy
	if !dirExists(repo) {
		panic(fmt.Sprintf("could not find base directory of this repo: '%v'", repo))
	}
	gopath = os.Getenv("GOPATH")
	home = os.Getenv("HOME")
	if gopath == "" {
		gopath = filepath.Join(home, "go")
	}

	ivyIncludeDirs = []string{
		filepath.Join(repo, "/pyivy/ivy/ivy/include/"),
		filepath.Join(repo, "/ivy-lang-examples/ivy/include/"),
		filepath.Join(home, "/ivy/pyivy/ivy/ivy/include/"),
		filepath.Join(home, "/ivy/ivy-lang-examples/ivy/include/"),
		filepath.Join(gopath, "/src/github.com/glycerine/ivy/ivy-lang-examples/ivy/include/"),
		filepath.Join(gopath, "/src/github.com/glycerine/ivy/pyivy/ivy/ivy/include/"),
	}
	ivyExamplesDir = []string{
		filepath.Join(home, "/ivy/ivy-lang-examples/"),
		filepath.Join(repo, "/ivy-lang-examples/"),
		filepath.Join(gopath, "/src/github.com/glycerine/ivy/ivy-lang-examples/"),
	}
}

func Trace1(format string, args ...interface{}) {
	trace(format, args...)
}

// Trace prints an execution trace line to stdout.
// Format: "XTRACE: " + fmt.Sprintf(format, args...) + "\n"
func Trace(format string, args ...interface{}) {
	trace(format, args...)
}

// Trace prints an execution trace line to stdout.
// Format: "XTRACE: " + fmt.Sprintf(format, args...) + "\n"
func trace(format string, args ...interface{}) {
	if suppressed {
		return
	}
	// replace true/false with True/False to match python
	// and avoid spurious diffs.
	for i, a := range args {
		switch b := a.(type) {
		case bool:
			if b {
				args[i] = "True"
			} else {
				args[i] = "False"
			}
		case string:
			args[i] = NormalizeLine(b)
		}
	}

	if strings.Contains(format, "\n") {
		splt := strings.SplitN(format, "\n", 2)
		if len(splt) == 2 {
			format = splt[0] + "\n;" + fileLine(2) + ":" + splt[1]
		}
	}
	fmt.Printf("XTRACE: "+format+"\n", args...)
}

func fileLine(depth int) string {
	_, fileName, fileLine, ok := runtime.Caller(depth)
	var s string
	if ok {
		s = fmt.Sprintf("%s:%d", path.Base(fileName), fileLine)
	} else {
		s = ""
	}
	return s
}

func NormalizeLine(line string) string {
	// Strip known path prefixes for include files
	for _, prefix := range ivyIncludeDirs {
		if strings.Contains(line, prefix) {
			line = strings.ReplaceAll(line, prefix, "<IVY_INCLUDE>")
		}
	}
	for _, prefix := range ivyExamplesDir {
		if strings.Contains(line, prefix) {
			//fmt.Printf("from: '%v'\n", line)
			line = strings.ReplaceAll(line, prefix, "<IVY_EXAMPLES>")
			//fmt.Printf("to: '%v'\n", line)
		}
	}
	return line
}

func dirExists(name string) bool {
	fi, err := os.Stat(name)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return true
	}
	return false
}

//go:build !xtrace_off && !js

package xtracer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// base directory of this (ivy) Go project
var repo string
var gopath string
var home string
var ivyIncludeDirs []string
var ivyExamplesDir []string

func init() {
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
	user := os.Getenv("USER")
	users := filepath.Join("/Users", user)

	ivyIncludeDirs = []string{
		filepath.Join(repo, "/pyivy/ivy/ivy/include/"),
		filepath.Join(repo, "/ivy-lang-examples/ivy/include/"),
		filepath.Join(home, "/ivy/pyivy/ivy/ivy/include/"),
		filepath.Join(home, "/ivy/ivy-lang-examples/ivy/include/"),
		filepath.Join(users, "/ivy/pyivy/ivy/ivy/include/"),
		filepath.Join(users, "/ivy/ivy-lang-examples/ivy/include/"),
		filepath.Join(gopath, "/src/github.com/glycerine/ivy/ivy-lang-examples/ivy/include/"),
		filepath.Join(gopath, "/src/github.com/glycerine/ivy/pyivy/ivy/ivy/include/"),
	}
	ivyIncludeDirs = appendCanonicalPathAliases(ivyIncludeDirs)
	ivyExamplesDir = []string{
		filepath.Join(home, "/ivy/ivy-lang-examples/"),
		filepath.Join(repo, "/ivy-lang-examples/"),
		filepath.Join(users, "/ivy/ivy-lang-examples/"),
		filepath.Join(gopath, "/src/github.com/glycerine/ivy/ivy-lang-examples/"),
	}
	ivyExamplesDir = appendCanonicalPathAliases(ivyExamplesDir)
	sort.Slice(ivyIncludeDirs, func(i, j int) bool {
		return len(ivyIncludeDirs[i]) > len(ivyIncludeDirs[j])
	})
	sort.Slice(ivyExamplesDir, func(i, j int) bool {
		return len(ivyExamplesDir[i]) > len(ivyExamplesDir[j])
	})
}

func appendCanonicalPathAliases(paths []string) []string {
	seen := make(map[string]bool, len(paths)*2)
	var result []string
	for _, p := range paths {
		if p == "" {
			continue
		}
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
		if real, err := filepath.EvalSymlinks(p); err == nil && real != "" && !seen[real] {
			seen[real] = true
			result = append(result, real)
		}
	}
	return result
}

func NormalizeLine(line string) string {
	// Strip known path prefixes for include files.
	for _, prefix := range ivyIncludeDirs {
		if strings.Contains(line, prefix) {
			line = strings.ReplaceAll(line, prefix, "<IVY_INCLUDE>")
		}
	}
	for _, prefix := range ivyExamplesDir {
		if strings.Contains(line, prefix) {
			line = strings.ReplaceAll(line, prefix, "<IVY_EXAMPLES>")
		}
	}
	return line
}

func dirExists(name string) bool {
	fi, err := os.Stat(name)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

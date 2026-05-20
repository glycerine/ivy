package ivy2cpp

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type missingZ3ToolchainError struct {
	reason string
}

func (e *missingZ3ToolchainError) Error() string {
	return "ivy2cpp: Z3 C++ toolchain unavailable: " + e.reason
}

func isMissingZ3ToolchainError(err error) bool {
	var target *missingZ3ToolchainError
	return errors.As(err, &target)
}

func BuildOutput(out *Output, outDir string) (string, error) {
	if out == nil {
		return "", fmt.Errorf("ivy2cpp: nil output")
	}
	if err := WriteOutput(out, outDir); err != nil {
		return "", err
	}
	dir := outDir
	if dir == "" {
		dir = "."
	}
	cxx, err := cxxCompiler()
	if err != nil {
		return "", err
	}
	cppPath := filepath.Join(dir, out.BaseName+".cpp")
	exePath := filepath.Join(dir, out.BaseName)
	args := []string{"-std=c++11"}
	if includeArgs, err := supportIncludeArgs(); err == nil {
		args = append(args, includeArgs...)
	} else {
		return "", err
	}
	var linkArgs []string
	if outputUsesZ3(out) {
		includeArgs, z3LinkArgs, err := z3BuildArgs()
		if err != nil {
			return "", err
		}
		args = append(args, includeArgs...)
		linkArgs = z3LinkArgs
	}
	args = append(args, cppPath, "-o", exePath)
	args = append(args, linkArgs...)
	cmd := exec.Command(cxx, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ivy2cpp: build failed: %w\n%s", err, buf.String())
	}
	return exePath, nil
}

func supportIncludeArgs() ([]string, error) {
	goivyRoot, err := packageGoivyRoot()
	if err != nil {
		return nil, err
	}
	includeDir := filepath.Join(filepath.Dir(goivyRoot), "include2cpp")
	if _, err := os.Stat(filepath.Join(includeDir, "ivy_hash.hpp")); err != nil {
		return nil, fmt.Errorf("ivy2cpp: support include directory not found: %w", err)
	}
	return []string{"-I" + includeDir}, nil
}

func cxxCompiler() (string, error) {
	if cxx := strings.TrimSpace(os.Getenv("CXX")); cxx != "" {
		return cxx, nil
	}
	if cxx, err := exec.LookPath("c++"); err == nil {
		return cxx, nil
	}
	if cxx, err := exec.LookPath("g++"); err == nil {
		return cxx, nil
	}
	return "", &missingZ3ToolchainError{reason: "no c++ or g++ compiler found"}
}

func outputUsesZ3(out *Output) bool {
	if out == nil {
		return false
	}
	return strings.Contains(out.Header, "z3++.h") || strings.Contains(out.Impl, "z3::")
}

func z3BuildArgs() ([]string, []string, error) {
	goivyRoot, err := packageGoivyRoot()
	if err != nil {
		return nil, nil, err
	}
	var includeDirs []string
	var libDirs []string
	if z3dir := strings.TrimSpace(os.Getenv("Z3DIR")); z3dir != "" {
		includeDirs = append(includeDirs, filepath.Join(z3dir, "include"), filepath.Join(z3dir, "include", "c++"))
		libDirs = append(libDirs, filepath.Join(z3dir, "lib"), filepath.Join(z3dir, "bin"))
	}
	includeDirs = append(includeDirs,
		filepath.Join(goivyRoot, "z3vendor", "z3", "src", "api"),
		filepath.Join(goivyRoot, "z3vendor", "z3", "src", "api", "c++"),
	)
	libDirs = append(libDirs, filepath.Join(goivyRoot, "z3vendor", "native_lib"))

	includeDirs = existingIncludeDirs(includeDirs)
	libDirs = existingZ3LibDirs(libDirs)
	if len(includeDirs) == 0 {
		return nil, nil, &missingZ3ToolchainError{reason: "z3++.h include directory not found"}
	}
	if len(libDirs) == 0 {
		return nil, nil, &missingZ3ToolchainError{reason: "libz3 not found"}
	}
	var includeArgs []string
	for _, dir := range includeDirs {
		includeArgs = append(includeArgs, "-I"+dir)
	}
	var linkArgs []string
	for _, dir := range libDirs {
		linkArgs = append(linkArgs, "-L"+dir)
		if runtime.GOOS == "darwin" {
			linkArgs = append(linkArgs, "-Wl,-rpath,"+dir)
		}
	}
	linkArgs = append(linkArgs, "-lz3")
	return includeArgs, linkArgs, nil
}

func existingIncludeDirs(candidates []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, dir := range candidates {
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		if _, err := os.Stat(filepath.Join(dir, "z3++.h")); err == nil {
			out = append(out, dir)
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, "z3.h")); err == nil {
			out = append(out, dir)
		}
	}
	return out
}

func existingZ3LibDirs(candidates []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, dir := range candidates {
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		if hasZ3Lib(dir) {
			out = append(out, dir)
		}
	}
	return out
}

func hasZ3Lib(dir string) bool {
	for _, name := range []string{"libz3.a", "libz3.dylib", "libz3.so", "z3.lib"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}

func packageGoivyRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", &missingZ3ToolchainError{reason: "cannot locate ivy2cpp package"}
	}
	return filepath.Dir(filepath.Dir(file)), nil
}

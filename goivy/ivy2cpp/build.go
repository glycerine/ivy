package ivy2cpp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
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

type BuildPlan struct {
	Compiler    string
	Args        []string
	Env         []string
	WorkDir     string
	OutputPath  string
	CompileOnly bool
}

func BuildOutput(out *Output, outDir string) (string, error) {
	if out == nil {
		return "", fmt.Errorf("ivy2cpp: nil output")
	}
	if err := WriteOutput(out, outDir); err != nil {
		return "", err
	}
	cfg, _, err := normalizeConfig(out.Config)
	if err != nil {
		return "", err
	}
	compiler, err := cxxCompilerFor(cfg.Compiler)
	if err != nil {
		return "", err
	}
	plan, err := BuildPlanFor(out, outDir, cfg)
	if err != nil {
		return "", err
	}
	plan.Compiler = compiler
	cmd := exec.Command(plan.Compiler, plan.Args...)
	if plan.WorkDir != "" {
		cmd.Dir = plan.WorkDir
	}
	if len(plan.Env) != 0 {
		cmd.Env = plan.Env
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ivy2cpp: build failed: %w\n%s", err, buf.String())
	}
	return plan.OutputPath, nil
}

func BuildPlanFor(out *Output, outDir string, cfg Config) (*BuildPlan, error) {
	if out == nil {
		return nil, fmt.Errorf("ivy2cpp: nil output")
	}
	if cfg.Compiler == "" {
		cfg = out.Config
	}
	cfg, _, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Compiler == "cl" && outputUsesWideBV(out) {
		return nil, fmt.Errorf("ivy2cpp: compiler=cl cannot build bv widths greater than 64 bits; MSVC does not support unsigned __int128")
	}
	cxx, err := buildPlanCompiler(cfg.Compiler)
	if err != nil {
		return nil, err
	}
	dir := outputDirectory(outDir)
	cppPath := filepath.Join(dir, out.BaseName+".cpp")
	compileOnly := out.Target == "class" || !out.EmitMain
	outputPath := filepath.Join(dir, out.BaseName)
	if compileOnly {
		outputPath += ".o"
	}
	if cfg.Compiler == "cl" {
		return msvcBuildPlan(out, cxx, cppPath, outputPath, compileOnly)
	}
	args := []string{"-std=c++11", "-Wno-parentheses-equality", "-g"}
	if includeArgs, err := supportIncludeArgs(); err == nil {
		args = append(args, includeArgs...)
	} else {
		return nil, err
	}
	var linkArgs []string
	if outputUsesZ3(out) {
		includeArgs, z3LinkArgs, err := z3BuildArgs()
		if err != nil {
			return nil, err
		}
		args = append(args, includeArgs...)
		linkArgs = z3LinkArgs
	}
	args = append(args, includeLibSpecArgs(out)...)
	args = append(args, cppPath)
	if compileOnly {
		args = append(args, "-c")
	}
	args = append(args, "-o", outputPath)
	if !compileOnly {
		args = append(args, linkArgs...)
		args = append(args, linkLibSpecArgs(out)...)
		args = append(args, "-pthread")
	}
	return &BuildPlan{
		Compiler:    cxx,
		Args:        args,
		WorkDir:     "",
		OutputPath:  outputPath,
		CompileOnly: compileOnly,
	}, nil
}

func buildPlanCompiler(compiler string) (string, error) {
	switch compiler {
	case "", "default":
		if cxx := strings.TrimSpace(os.Getenv("CXX")); cxx != "" {
			return cxx, nil
		}
		return "c++", nil
	case "g++":
		return "g++", nil
	case "cl":
		if runtime.GOOS != "windows" {
			return "", &missingZ3ToolchainError{reason: "compiler=cl is only supported on Windows"}
		}
		return "cl", nil
	default:
		return "", fmt.Errorf("ivy2cpp: compiler %q is not supported", compiler)
	}
}

func msvcBuildPlan(out *Output, compiler, cppPath, outputPath string, compileOnly bool) (*BuildPlan, error) {
	args := []string{"/EHsc", "/Zi"}
	var env []string
	var toolchainLinkArgs []string
	if vs, err := findVS(); err == nil {
		args = append(args, msvcIncludeDirArgs(vs.IncludeDirs)...)
		toolchainLinkArgs = append(toolchainLinkArgs, msvcLibDirArgs(vs.LibDirs)...)
		env = msvcToolchainEnv(vs)
	}
	if includeArgs, err := supportIncludeArgs(); err == nil {
		args = append(args, msvcIncludeArgs(includeArgs)...)
	} else {
		return nil, err
	}
	var linkArgs []string
	if outputUsesZ3(out) {
		includeArgs, z3LinkArgs, err := z3BuildArgs()
		if err != nil {
			return nil, err
		}
		args = append(args, msvcIncludeArgs(includeArgs)...)
		linkArgs = msvcLinkArgs(z3LinkArgs)
	}
	args = append(args, msvcIncludeArgs(includeLibSpecArgs(out))...)
	if compileOnly {
		args = append(args, "/c", cppPath, "/Fo"+outputPath)
	} else {
		args = append(args, cppPath, "/Fe"+outputPath, "ws2_32.lib")
		args = append(args, toolchainLinkArgs...)
		args = append(args, linkArgs...)
		args = append(args, msvcLinkArgs(linkLibSpecArgs(out))...)
	}
	return &BuildPlan{
		Compiler:    compiler,
		Args:        args,
		Env:         env,
		OutputPath:  outputPath,
		CompileOnly: compileOnly,
	}, nil
}

func msvcIncludeDirArgs(dirs []string) []string {
	var out []string
	for _, dir := range dirs {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		out = append(out, "/I", dir)
	}
	return out
}

func msvcLibDirArgs(dirs []string) []string {
	var out []string
	for _, dir := range dirs {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		out = append(out, "/LIBPATH:"+dir)
	}
	return out
}

func msvcIncludeArgs(args []string) []string {
	var out []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-I") {
			out = append(out, "/I", strings.TrimPrefix(arg, "-I"))
			continue
		}
		out = append(out, arg)
	}
	return out
}

func msvcLinkArgs(args []string) []string {
	var out []string
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "-L"):
			out = append(out, "/LIBPATH:"+strings.TrimPrefix(arg, "-L"))
		case strings.HasPrefix(arg, "-l"):
			out = append(out, strings.TrimPrefix(arg, "-l")+".lib")
		case strings.HasPrefix(arg, "-Wl,"):
		default:
			out = append(out, arg)
		}
	}
	return out
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
	return cxxCompilerFor("default")
}

func cxxCompilerFor(compiler string) (string, error) {
	switch compiler {
	case "", "default":
	case "g++":
		if cxx, err := exec.LookPath("g++"); err == nil {
			return cxx, nil
		}
		return "", &missingZ3ToolchainError{reason: "g++ compiler not found"}
	case "cl":
		if runtime.GOOS != "windows" {
			return "", &missingZ3ToolchainError{reason: "compiler=cl is only supported on Windows"}
		}
		if cxx, err := exec.LookPath("cl"); err == nil {
			return cxx, nil
		}
		return "", &missingZ3ToolchainError{reason: "cl compiler not found"}
	default:
		return "", fmt.Errorf("ivy2cpp: compiler %q is not supported", compiler)
	}
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

func outputUsesWideBV(out *Output) bool {
	if out == nil {
		return false
	}
	return strings.Contains(out.Header, "unsigned __int128") ||
		strings.Contains(out.Header, "ivy_uint<") ||
		strings.Contains(out.Impl, "unsigned __int128") ||
		strings.Contains(out.Impl, "ivy_uint<")
}

func z3BuildArgs() ([]string, []string, error) {
	goivyRoot, err := packageGoivyRoot()
	if err != nil {
		return nil, nil, err
	}
	includeDirs := []string{
		filepath.Join(goivyRoot, "z3vendor", "z3", "src", "api"),
		filepath.Join(goivyRoot, "z3vendor", "z3", "src", "api", "c++"),
	}
	staticLib := filepath.Join(goivyRoot, "z3vendor", "native_lib", "libz3.a")

	includeDirs = existingIncludeDirs(includeDirs)
	if len(includeDirs) == 0 {
		return nil, nil, &missingZ3ToolchainError{reason: "z3++.h include directory not found"}
	}
	if _, err := os.Stat(staticLib); err != nil {
		return nil, nil, &missingZ3ToolchainError{reason: "vendored static libz3.a not found"}
	}
	var includeArgs []string
	for _, dir := range includeDirs {
		includeArgs = append(includeArgs, "-I"+dir)
	}
	linkArgs := []string{staticLib}
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

func includeLibSpecArgs(out *Output) []string {
	var args []string
	for _, spec := range combinedLibSpecs(out) {
		if strings.HasPrefix(spec, "-I") || strings.HasPrefix(spec, "/I") {
			args = append(args, spec)
		}
	}
	return args
}

func linkLibSpecArgs(out *Output) []string {
	var args []string
	for _, spec := range combinedLibSpecs(out) {
		if strings.HasPrefix(spec, "-I") || strings.HasPrefix(spec, "/I") {
			continue
		}
		if strings.HasSuffix(spec, ".lib") {
			if runtime.GOOS == "windows" {
				args = append(args, spec)
			}
			continue
		}
		if strings.HasPrefix(spec, "-") {
			args = append(args, spec)
		} else {
			args = append(args, "-l"+spec)
		}
	}
	return args
}

func combinedLibSpecs(out *Output) []string {
	seen := map[string]bool{}
	var specs []string
	for _, spec := range readSpecsFileLibs() {
		if !seen[spec] {
			seen[spec] = true
			specs = append(specs, spec)
		}
	}
	if out != nil {
		for _, spec := range out.LibSpecs {
			if !seen[spec] {
				seen[spec] = true
				specs = append(specs, spec)
			}
		}
	}
	sort.Strings(specs)
	return specs
}

func readSpecsFileLibs() []string {
	goivyRoot, err := packageGoivyRoot()
	if err != nil {
		return nil
	}
	path := filepath.Join(filepath.Dir(goivyRoot), "pyivy", "ivy", "lib", "specs")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var raw [][]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	var specs []string
	for _, entry := range raw {
		if len(entry) >= 2 {
			if dir, ok := entry[1].(string); ok && dir != "" {
				specs = append(specs, "-I"+filepath.Join(dir, "include"))
			}
		}
		if len(entry) >= 3 {
			if dir, ok := entry[2].(string); ok && dir != "" {
				specs = append(specs, "-L"+dir)
			}
		}
	}
	return specs
}

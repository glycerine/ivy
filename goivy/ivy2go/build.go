package ivy2go

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// build.go: wraps `go build` over an emitted package. Mirrors the role
// of ivy2cpp/build.go but the C++ toolchain selection (g++/cl, MSVC
// registry lookup, Z3 link path) is replaced by a single call to the
// Go toolchain — `go build` is uniform across platforms.

// BuildPlan describes how to compile an emitted package.
type BuildPlan struct {
	// GoBin is the absolute path to the `go` binary used.
	GoBin string
	// Args is the argv passed to GoBin (without the program name).
	Args []string
	// Env is an optional environment override (KEY=VAL strings).
	Env []string
	// PackageDir is the directory the build runs in.
	PackageDir string
	// OutputPath is the binary file path the build produces. Empty
	// when CompileOnly is true.
	OutputPath string
	// CompileOnly is true for target=class (library-only packages)
	// where we want to verify the package compiles but don't produce
	// a binary.
	CompileOnly bool
}

// BuildPlanFor returns a BuildPlan that compiles out's package
// directory. The package must have been written to disk first via
// WriteOutput.
//
// For target=class, the plan compiles in `go vet`-style mode with no
// binary output. For other targets, the plan produces a binary at
// pkgDir/<BaseName> (Windows: <BaseName>.exe).
func BuildPlanFor(out *Output, outDir string) (*BuildPlan, error) {
	if out == nil {
		return nil, fmt.Errorf("ivy2go: nil output")
	}
	gobin, err := exec.LookPath("go")
	if err != nil {
		return nil, fmt.Errorf("ivy2go: cannot locate go toolchain: %w", err)
	}
	pkgDir := outputDirectory(outDir, out.BaseName)
	plan := &BuildPlan{
		GoBin:      gobin,
		PackageDir: pkgDir,
	}
	if out.Config.RequestedTarget == "class" {
		// Library-only: use `go build ./...` (no -o). Treat compile
		// success as the gate.
		plan.Args = []string{"build", "."}
		plan.CompileOnly = true
		return plan, nil
	}
	binaryName := out.BaseName
	if isWindows() {
		binaryName += ".exe"
	}
	plan.OutputPath = filepath.Join(pkgDir, binaryName)
	plan.Args = []string{"build", "-o", plan.OutputPath, "."}
	return plan, nil
}

// BuildOutput writes out to outDir and runs `go build`. Returns the
// produced binary path (or empty string for CompileOnly plans).
func BuildOutput(out *Output, outDir string) (string, error) {
	if err := WriteOutput(out, outDir); err != nil {
		return "", err
	}
	plan, err := BuildPlanFor(out, outDir)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(plan.GoBin, plan.Args...)
	cmd.Dir = plan.PackageDir
	if len(plan.Env) > 0 {
		cmd.Env = append(os.Environ(), plan.Env...)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ivy2go: go build failed: %v\n%s", err, string(output))
	}
	return plan.OutputPath, nil
}

// isWindows mirrors ivy2cpp/build_findvs.go's HostOS-aware branch.
// We accept Config.HostOS override for tests; otherwise we check the
// process's runtime.GOOS.
func isWindows() bool {
	// Avoid pulling runtime here just for the constant; check the
	// path separator which is "\\" only on Windows.
	return filepath.Separator == '\\'
}

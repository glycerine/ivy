package ivy2go

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
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
	// Resolve OutputPath to an absolute path so the recorded path
	// matches where `go build` actually writes the binary. The
	// `-o` arg passed to `go build` is the bare name so it lands
	// in pkgDir (which is cmd.Dir).
	absPkgDir, err := filepath.Abs(pkgDir)
	if err != nil {
		return nil, fmt.Errorf("ivy2go: cannot resolve package dir: %w", err)
	}
	plan.OutputPath = filepath.Join(absPkgDir, binaryName)
	plan.Args = []string{"build", "-tags=xtrace_off", "-o", binaryName, "."}
	return plan, nil
}

// BuildOutput writes out to outDir and runs `go build`. Returns the
// produced binary path (or empty string for CompileOnly plans).
//
// Preflight checks (each returns a clear, actionable error):
//   - The package directory has an enclosing go.mod. ivy2go does NOT
//     emit a go.mod; place outdir inside an existing Go module.
//   - For non-class targets, the resolved PackageName must be "main"
//     so `func main()` actually lands. Without this, `go build`
//     silently produces a .a archive instead of an executable.
func BuildOutput(out *Output, outDir string) (string, error) {
	if err := WriteOutput(out, outDir); err != nil {
		return "", err
	}
	plan, err := BuildPlanFor(out, outDir)
	if err != nil {
		return "", err
	}
	if err := checkBuildContext(plan.PackageDir, out); err != nil {
		return "", err
	}
	if err := checkBuildableBinary(out); err != nil {
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

// checkBuildableBinary errors when the resolved Config would
// produce a Go library (.a archive) instead of an executable.
// `func main()` can only live in `package main`, so non-class
// targets with a non-main PackageName silently fail to produce a
// binary — surface that upfront.
func checkBuildableBinary(out *Output) error {
	if out.Config.RequestedTarget == "class" {
		// class target is library-only by design (CompileOnly plan).
		return nil
	}
	if out.PackageName == "main" {
		return nil
	}
	return fmt.Errorf(`ivy2go: cannot build executable for target=%s with package=%q.
Go's `+"`func main()`"+` only lives in package main; without it the build
produces a .a archive, not an executable. Pass package=main on the
command line (or set Config.PackageName = "main" in code).`,
		out.Config.RequestedTarget, out.PackageName)
}

// checkBuildContext verifies the emitted package is inside an
// existing Go module. On failure it returns an actionable error.
func checkBuildContext(pkgDir string, out *Output) error {
	_ = out
	if findEnclosingGoMod(pkgDir) != "" {
		return nil
	}
	return fmt.Errorf(`ivy2go: cannot build %s — the directory has no enclosing go.mod.
Set outdir=<dir> to a path inside an existing Go module whose go.mod
provides github.com/glycerine/ivy/goivy (e.g. inside the goivy repo
itself, or inside a workspace that includes it). ivy2go does not
emit a go.mod for the generated package.`, pkgDir)
}

// findEnclosingGoMod walks up from dir looking for a go.mod file.
// Returns the absolute path to the go.mod (or "" if none found).
func findEnclosingGoMod(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(abs, "go.mod")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return ""
		}
		abs = parent
	}
}

// isWindows mirrors ivy2cpp/build_findvs.go's HostOS-aware branch.
// We accept Config.HostOS override for tests; otherwise we check the
// process's runtime.GOOS.
func isWindows() bool {
	// Avoid pulling runtime here just for the constant; check the
	// path separator which is "\\" only on Windows.
	return filepath.Separator == '\\'
}

// hostOS mirrors ivy2cpp/runtime.go hostOS. Returns the configured
// HostOS override or — when empty — the actual `runtime.GOOS` of the
// process. Used by emission paths that need to branch on the build
// host (mostly a stub on the Go side; `go build` is uniform).
func (g *Generator) hostOS() string {
	if g != nil && g.Config.HostOS != "" {
		return g.Config.HostOS
	}
	return goruntime.GOOS
}

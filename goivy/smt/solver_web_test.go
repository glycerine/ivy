//go:build web && !wasip1 && (darwin || linux || windows)

package smt

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const solverBoolRoundTripWasip1Source = `//go:build wasip1

package main

import "github.com/glycerine/ivy/goivy/smt"

func main() {}

//go:wasmexport smt_solver_sat_true
func smtSolverSatTrue() int32 {
	ctx := smt.NewZ3Context()
	defer ctx.Close()

	solver := ctx.NewZ3Solver()
	solver.Assert(ctx.BoolVal(true))
	return int32(solver.Check())
}

//go:wasmexport smt_solver_unsat_true_and_not_true
func smtSolverUnsatTrueAndNotTrue() int32 {
	ctx := smt.NewZ3Context()
	defer ctx.Close()

	truth := ctx.BoolVal(true)
	solver := ctx.NewZ3Solver()
	solver.Assert(truth)
	solver.Assert(ctx.Not(truth))
	return int32(solver.Check())
}

//go:wasmexport smt_solver_uninterpreted_sort_name_ok
func smtSolverUninterpretedSortNameOK() int32 {
	ctx := smt.NewZ3Context()
	defer ctx.Close()

	sort := ctx.UninterpretedSort("Thing")
	if sort.String() != "Thing" {
		return 0
	}
	return 1
}
`

const z3ErrorBoundaryWasip1Source = `//go:build wasip1

package main

import "github.com/glycerine/ivy/goivy/smt"

func main() {}

//go:wasmexport smt_z3_invalid_bv_sort_returns_to_caller
func smtZ3InvalidBVSortReturnsToCaller() (recovered int32) {
	ctx := smt.NewZ3Context()
	defer ctx.Close()

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(*smt.ErrMsg); ok {
				recovered = 1
			}
		}
	}()

	_ = ctx.BvSort(0)
	return 0
}
`

func TestSolverBoolRoundTrip(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve smt test path")
	}
	goivyDir := filepath.Dir(filepath.Dir(file))
	webvueDir := filepath.Join(goivyDir, "webvue")
	staticDir := filepath.Join(webvueDir, "static")

	requireFile(t, filepath.Join(staticDir, "z3-471-api.js"), "run make z3-wasm-api")
	requireFile(t, filepath.Join(staticDir, "z3-471-api.wasm"), "run make z3-wasm-api")
	requireFile(t, filepath.Join(webvueDir, "src", "workers", "smtZ3Imports.js"), "missing browser Z3 import host")
	requireFile(t, filepath.Join(webvueDir, "node_modules", ".bin", playwrightBin()), "run make webvue-setup")

	roundTripWasm := buildWasip1MainSource(t, goivyDir, staticDir, "solver-bool-round-trip", solverBoolRoundTripWasip1Source)

	runCommand(t, webvueDir, []string{"SMT_SOLVER_ROUND_TRIP_WASM=" + roundTripWasm},
		filepath.Join(".", "node_modules", ".bin", playwrightBin()),
		"test", "--config", "playwright.config.mjs", "tests/smtSolverRoundTrip.browser.spec.js",
	)
}

func TestZ3ErrorCallbackBoundary(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve smt test path")
	}
	goivyDir := filepath.Dir(filepath.Dir(file))
	webvueDir := filepath.Join(goivyDir, "webvue")
	staticDir := filepath.Join(webvueDir, "static")

	requireFile(t, filepath.Join(staticDir, "z3-471-api.js"), "run make z3-wasm-api")
	requireFile(t, filepath.Join(staticDir, "z3-471-api.wasm"), "run make z3-wasm-api")
	requireFile(t, filepath.Join(webvueDir, "src", "workers", "smtZ3Imports.js"), "missing browser Z3 import host")
	requireFile(t, filepath.Join(webvueDir, "node_modules", ".bin", playwrightBin()), "run make webvue-setup")

	wasm := buildWasip1MainSource(t, goivyDir, staticDir, "z3-error-boundary", z3ErrorBoundaryWasip1Source)

	runCommand(t, webvueDir, []string{"SMT_ERROR_BOUNDARY_WASM=" + wasm},
		filepath.Join(".", "node_modules", ".bin", playwrightBin()),
		"test", "--config", "playwright.config.mjs", "tests/smtSolverRoundTrip.browser.spec.js",
	)
}

func buildWasip1MainSource(t *testing.T, goivyDir, staticDir, name, source string) string {
	sourceDir := t.TempDir()
	requireWriteFile(t, filepath.Join(sourceDir, "go.mod"), fmt.Sprintf(`module smtwasip1fixture

go 1.25.0

require github.com/glycerine/ivy/goivy v0.0.0

replace github.com/glycerine/ivy/goivy => %s
`, filepath.ToSlash(goivyDir)))
	requireWriteFile(t, filepath.Join(sourceDir, "main.go"), source)

	assetName := fmt.Sprintf("smt-%s-%d-%s.wasm", safeAssetName(name), os.Getpid(), filepath.Base(sourceDir))
	wasmPath := filepath.Join(staticDir, assetName)
	t.Cleanup(func() {
		if err := os.Remove(wasmPath); err != nil && !os.IsNotExist(err) {
			t.Logf("could not remove generated wasm %s: %v", wasmPath, err)
		}
	})

	runCommand(t, sourceDir, []string{
		"GOCACHE=/private/tmp/go-build",
		"GOOS=wasip1",
		"GOARCH=wasm",
	}, "go", "build", "-o", wasmPath, ".")

	return assetName
}

func requireWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func safeAssetName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func playwrightBin() string {
	if runtime.GOOS == "windows" {
		return "playwright.cmd"
	}
	return "playwright"
}

func requireFile(t *testing.T, path, hint string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("missing required file %s: %v; %s", path, err, hint)
	}
}

func runCommand(t *testing.T, dir string, env []string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %v failed in %s: %v\n%s", name, args, dir, err, out.String())
	}
	if out.Len() > 0 {
		t.Logf("%s %v output:\n%s", name, args, out.String())
	}
}

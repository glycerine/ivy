//go:build web && !tinygo && !wasip1 && (darwin || linux || windows)

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
import "strings"

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

//go:wasmexport smt_solver_canon_preserves_enum_quantifier_and_false
func smtSolverCanonPreservesEnumQuantifierAndFalse() int32 {
	ctx := smt.NewZ3Context()
	defer ctx.Close()

	opSort, opVals := ctx.EnumSort("op_type", []string{"nop", "write", "read"})
	lclockSort := ctx.UninterpretedSort("lclock")
	req := ctx.Function("ref.evs.req", []smt.Z3Sort{lclockSort}, opSort)
	tick := ctx.Const("T", lclockSort)

	solver := ctx.NewZ3Solver()
	solver.Assert(ctx.ForAll([]smt.Z3Expr{tick}, ctx.Eq(req.Apply(tick), opVals[2])))
	solver.Assert(ctx.BoolVal(false))

	canon := solver.CanonZ3Assertions()
	if strings.Contains(canon, "unknown_kind=") {
		return 0
	}
	if !strings.Contains(canon, "(a = (a ref.evs.req (v T)) (c read op_type))") {
		return 0
	}
	if !strings.Contains(canon, "(c false Bool)") {
		return 0
	}
	return 1
}

//go:wasmexport smt_solver_canon_preserves_enum_function_copy
func smtSolverCanonPreservesEnumFunctionCopy() int32 {
	ctx := smt.NewZ3Context()
	defer ctx.Close()

	opSort, opVals := ctx.EnumSort("op_type", []string{"nop", "write", "read", "rrsp", "wr_cmp"})
	lclockSort := ctx.UninterpretedSort("lclock")
	req := ctx.Function("ref.evs.req", []smt.Z3Sort{lclockSort}, opSort)
	oldReq := ctx.Function("__ref.evs.req", []smt.Z3Sort{lclockSort}, opSort)
	tick := ctx.Const("V0:lclock", lclockSort)

	solver := ctx.NewZ3Solver()
	solver.Assert(ctx.ForAll([]smt.Z3Expr{tick}, ctx.Eq(req.Apply(tick), oldReq.Apply(tick))))
	solver.Assert(ctx.ForAll([]smt.Z3Expr{tick}, ctx.Eq(req.Apply(tick), ctx.Ite(ctx.BoolVal(true), oldReq.Apply(tick), opVals[0]))))

	canon := solver.CanonZ3Assertions()
	if strings.Contains(canon, "unknown_kind=") {
		return 0
	}
	if !strings.Contains(canon, "(a = (a ref.evs.req (v V0:lclock)) (a __ref.evs.req (v V0:lclock)))") {
		return 0
	}
	if !strings.Contains(canon, "(a if (c true Bool) (a __ref.evs.req (v V0:lclock)) (c nop op_type))") {
		return 0
	}
	return 1
}

//go:wasmexport smt_solver_canon_preserves_multi_bound_enum_premise
func smtSolverCanonPreservesMultiBoundEnumPremise() int32 {
	ctx := smt.NewZ3Context()
	defer ctx.Close()

	opSort, opVals := ctx.EnumSort("op_type", []string{"nop", "write", "read", "rrsp", "wr_cmp"})
	locSort, locVals := ctx.EnumSort("loc_type", []string{"init_l", "cf_mem_l", "cf_pio_l", "cf_cmp_l", "if_l", "memc_l", "dramc_l", "arm_l"})
	lclockSort := ctx.UninterpretedSort("lclock")
	tarCFClockSort := ctx.UninterpretedSort("tar_cf_clock")

	req := ctx.Function("__ref.evs.req", []smt.Z3Sort{lclockSort}, opSort)
	lReq := ctx.Function("__ref.evs.l_req", []smt.Z3Sort{lclockSort}, locSort)
	arrMin := ctx.Const("cfabric.t_rd_arr_min", tarCFClockSort)
	tick := ctx.Const("T:lclock", lclockSort)
	tarr := ctx.Const("Tarr:tar_cf_clock", tarCFClockSort)

	body := ctx.Implies(
		ctx.And(
			ctx.Eq(arrMin, tarr),
			ctx.Eq(req.Apply(tick), opVals[2]),
			ctx.Eq(lReq.Apply(tick), locVals[1]),
		),
		ctx.Eq(arrMin, tarr),
	)

	solver := ctx.NewZ3Solver()
	solver.Assert(ctx.ForAll([]smt.Z3Expr{tick, tarr}, body))

	canon := solver.CanonZ3Assertions()
	if strings.Contains(canon, "unknown_kind=") {
		return 0
	}
	if !strings.Contains(canon, "(a = (a __ref.evs.req (v T:lclock)) (c read op_type))") {
		return 0
	}
	if !strings.Contains(canon, "(a = (a __ref.evs.l_req (v T:lclock)) (c cf_mem_l loc_type))") {
		return 0
	}
	return 1
}
`

const solverBoolRoundTripJSSource = `//go:build js && wasm

package main

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/glycerine/ivy/goivy/smt"
)

var callbacks []js.Func

func main() {
	run := js.FuncOf(func(this js.Value, args []js.Value) any {
		return runSolverRoundTrip()
	})
	callbacks = append(callbacks, run)
	js.Global().Set("smtSolverRoundTripJSRun", run)
	js.Global().Set("smtSolverRoundTripJSReady", true)
	select {}
}

func runSolverRoundTrip() map[string]any {
	result := map[string]any{
		"satTrue":                           0,
		"unsatTrueAndNotTrue":               0,
		"uninterpretedSortNameOK":           0,
		"canonPreservesEnumQuantifierFalse": 0,
		"canonPreservesEnumFunctionCopy":    0,
		"canonPreservesMultiBoundPremise":   0,
		"canon":                             "",
		"panic":                             "",
	}
	defer func() {
		if r := recover(); r != nil {
			result["panic"] = fmt.Sprint(r)
		}
	}()

	ctx := smt.NewZ3Context()
	defer ctx.Close()

	satSolver := ctx.NewZ3Solver()
	satSolver.Assert(ctx.BoolVal(true))
	result["satTrue"] = int(satSolver.Check())

	unsatSolver := ctx.NewZ3Solver()
	truth := ctx.BoolVal(true)
	unsatSolver.Assert(truth)
	unsatSolver.Assert(ctx.Not(truth))
	result["unsatTrueAndNotTrue"] = int(unsatSolver.Check())

	sort := ctx.UninterpretedSort("Thing")
	if sort.String() == "Thing" {
		result["uninterpretedSortNameOK"] = 1
	}

	opSort, opVals := ctx.EnumSort("op_type", []string{"nop", "write", "read"})
	lclockSort := ctx.UninterpretedSort("lclock")
	req := ctx.Function("ref.evs.req", []smt.Z3Sort{lclockSort}, opSort)
	tick := ctx.Const("T", lclockSort)
	canonSolver := ctx.NewZ3Solver()
	canonSolver.Assert(ctx.ForAll([]smt.Z3Expr{tick}, ctx.Eq(req.Apply(tick), opVals[2])))
	canonSolver.Assert(ctx.BoolVal(false))
	canon := canonSolver.CanonZ3Assertions()
	result["canon"] = canon
	if !strings.Contains(canon, "unknown_kind=") &&
		strings.Contains(canon, "(a = (a ref.evs.req (v T)) (c read op_type))") &&
		strings.Contains(canon, "(c false Bool)") {
		result["canonPreservesEnumQuantifierFalse"] = 1
	}

	opSortCopy, opValsCopy := ctx.EnumSort("op_type_copy", []string{"nop", "write", "read", "rrsp", "wr_cmp"})
	lclockSortCopy := ctx.UninterpretedSort("lclock_copy")
	reqCopy := ctx.Function("ref.evs.req.copy", []smt.Z3Sort{lclockSortCopy}, opSortCopy)
	oldReqCopy := ctx.Function("__ref.evs.req.copy", []smt.Z3Sort{lclockSortCopy}, opSortCopy)
	tickCopy := ctx.Const("V0:lclock", lclockSortCopy)
	copySolver := ctx.NewZ3Solver()
	copySolver.Assert(ctx.ForAll([]smt.Z3Expr{tickCopy}, ctx.Eq(reqCopy.Apply(tickCopy), oldReqCopy.Apply(tickCopy))))
	copySolver.Assert(ctx.ForAll([]smt.Z3Expr{tickCopy}, ctx.Eq(reqCopy.Apply(tickCopy), ctx.Ite(ctx.BoolVal(true), oldReqCopy.Apply(tickCopy), opValsCopy[0]))))
	copyCanon := copySolver.CanonZ3Assertions()
	if !strings.Contains(copyCanon, "unknown_kind=") &&
		strings.Contains(copyCanon, "(a = (a ref.evs.req.copy (v V0:lclock)) (a __ref.evs.req.copy (v V0:lclock)))") &&
		strings.Contains(copyCanon, "(a if (c true Bool) (a __ref.evs.req.copy (v V0:lclock)) (c nop op_type_copy))") {
		result["canonPreservesEnumFunctionCopy"] = 1
	}

	locSort, locVals := ctx.EnumSort("loc_type", []string{"init_l", "cf_mem_l", "cf_pio_l", "cf_cmp_l", "if_l", "memc_l", "dramc_l", "arm_l"})
	tarCFClockSort := ctx.UninterpretedSort("tar_cf_clock")
	reqPremise := ctx.Function("__ref.evs.req", []smt.Z3Sort{lclockSort}, opSort)
	lReqPremise := ctx.Function("__ref.evs.l_req", []smt.Z3Sort{lclockSort}, locSort)
	arrMin := ctx.Const("cfabric.t_rd_arr_min", tarCFClockSort)
	tickPremise := ctx.Const("T:lclock", lclockSort)
	tarr := ctx.Const("Tarr:tar_cf_clock", tarCFClockSort)
	premiseBody := ctx.Implies(
		ctx.And(
			ctx.Eq(arrMin, tarr),
			ctx.Eq(reqPremise.Apply(tickPremise), opVals[2]),
			ctx.Eq(lReqPremise.Apply(tickPremise), locVals[1]),
		),
		ctx.Eq(arrMin, tarr),
	)
	premiseSolver := ctx.NewZ3Solver()
	premiseSolver.Assert(ctx.ForAll([]smt.Z3Expr{tickPremise, tarr}, premiseBody))
	premiseCanon := premiseSolver.CanonZ3Assertions()
	if !strings.Contains(premiseCanon, "unknown_kind=") &&
		strings.Contains(premiseCanon, "(a = (a __ref.evs.req (v T:lclock)) (c read op_type))") &&
		strings.Contains(premiseCanon, "(a = (a __ref.evs.l_req (v T:lclock)) (c cf_mem_l loc_type))") {
		result["canonPreservesMultiBoundPremise"] = 1
	}

	return result
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

func TestSolverBoolRoundTripJS(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve smt test path")
	}
	goivyDir := filepath.Dir(filepath.Dir(file))
	webvueDir := filepath.Join(goivyDir, "webvue")
	staticDir := filepath.Join(webvueDir, "static")

	requireFile(t, filepath.Join(staticDir, "z3-471-api.js"), "run make z3-wasm-api")
	requireFile(t, filepath.Join(staticDir, "z3-471-api.wasm"), "run make z3-wasm-api")
	requireFile(t, filepath.Join(staticDir, "wasm_exec-go1.25.6.js"), "run make vendor-go-wasm-exec-js")
	requireFile(t, filepath.Join(webvueDir, "src", "workers", "smtZ3Imports.js"), "missing browser Z3 import host")
	requireFile(t, filepath.Join(webvueDir, "node_modules", ".bin", playwrightBin()), "run make webvue-setup")

	wasm := buildJSMainSource(t, goivyDir, staticDir, "solver-bool-round-trip-js", solverBoolRoundTripJSSource)

	runCommand(t, webvueDir, []string{"SMT_SOLVER_ROUND_TRIP_JS_WASM=" + wasm},
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

func buildJSMainSource(t *testing.T, goivyDir, staticDir, name, source string) string {
	sourceDir := t.TempDir()
	requireWriteFile(t, filepath.Join(sourceDir, "go.mod"), fmt.Sprintf(`module smtjsfixture

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
		"GOOS=js",
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

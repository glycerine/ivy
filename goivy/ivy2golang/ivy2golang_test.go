package ivy2golang

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

func compileIvySource(t *testing.T, src string) *goivy.Module {
	t.Helper()
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	sig := goivy.NewSigOn(mod.Cfg.IuCfg)
	if err := goivy.SourceString("test.ivy", src, mod, sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatalf("compile ivy source: %v", err)
	}
	return mod
}

func TestGenerateEmptyModuleProducesGoSource(t *testing.T) {
	mod := goivy.New()
	mod.Name = "empty"
	out, err := Generate(mod, Config{Target: "test"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out.BaseName != "empty" || out.ClassName != "empty" {
		t.Fatalf("unexpected names: %+v", out)
	}
	for _, want := range []string{"package main", "type empty struct", "func main()", "test_completed"} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestGenerateBasicAssignCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type small = {0..3}
individual x : small
after init {
    x := 0
}
action set = {
    x := 1
}
export set
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "basic_assign", TestIters: "2"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{"x int", "func (ivy *basic_assign) set()", "ivy.x = 1"} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGenerateForallAssignUsesMapState(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green, blue}
relation marked(C:color)
after init {
    marked(C) := false
}
action set_all = {
    marked(C) := true
}
export set_all
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "forall_assign", TestIters: "2"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type color int",
		"marked map[color]bool",
		"for _, C := range []color{red, green, blue}",
		"ivy.marked[C] = true",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGenerateActionParamsAndReturnsCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green, blue}
individual saved : color
action choose(c:color) returns(out:color) = {
    saved := c;
    out := saved
}
export choose
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "enum_dispatch", TestIters: "2"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func (ivy *enum_dispatch) choose(c color) color",
		"out := color(ivy.___ivy_choose",
		"ivy.saved = c",
		"_ = ivy.choose(color(ivy.___ivy_choose",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestOracleRangeBoundsCompilesAndRuns(t *testing.T) {
	fixture := filepath.Join("..", "ivy2cpp", "test_vec", "oracle", "range_bounds.ivy")
	out, err := CompileAndGenerate(fixture, map[string]string{"target": "test", "classname": "range_bounds"}, Config{TestIters: "2"})
	if err != nil {
		t.Fatalf("CompileAndGenerate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"saved int",
		"func (ivy *range_bounds) ext__set(i int) int",
		"ivy.saved = i",
		"(2 + ivy.___ivy_choose(6",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestOracleProgressPropertyEmitsTickCounters(t *testing.T) {
	fixture := filepath.Join("..", "ivy2cpp", "test_vec", "oracle", "progress_property.ivy")
	out, err := CompileAndGenerate(fixture, map[string]string{"target": "test", "classname": "progress_property"}, Config{TestIters: "2"})
	if err != nil {
		t.Fatalf("CompileAndGenerate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"wait int",
		"helper int",
		"if ivy.ready {",
		"ivy.wait = ivy.wait + 1",
		"if ivy.helper > __tmp0 {",
		"if __tmp0 > timeout {",
		"ivyCheckProgress(ivy.wait, __tmp0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestOracleBVArithmeticCompilesAndRuns(t *testing.T) {
	fixture := filepath.Join("..", "ivy2cpp", "test_vec", "oracle", "bv_arithmetic.ivy")
	out, err := CompileAndGenerate(fixture, map[string]string{"target": "test", "classname": "bv_arithmetic"}, Config{TestIters: "1"})
	if err != nil {
		t.Fatalf("CompileAndGenerate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"x int",
		"y int",
		"bvxor map[struct{ A0 int; A1 int }]int",
		"ivy.bvxor[struct{ A0 int; A1 int }{ivy.x, ivy.y}]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGenerateQuantifiedAssertionCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
after init {
    marked(C) := true
}
action check = {
    assert forall C:color. marked(C)
}
export check
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "quant_assert", TestIters: "2"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func() bool {",
		"return false",
		"return true",
		"ivyAssert((func() bool",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestParseArgsMatchesIvy2CppShape(t *testing.T) {
	params, file, err := ParseArgs([]string{"target=test", "classname=MyIvy", "outdir=build", "x.ivy"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if file != "x.ivy" || params["target"] != "test" || params["classname"] != "MyIvy" || params["outdir"] != "build" {
		t.Fatalf("unexpected parse result: params=%v file=%q", params, file)
	}
}

func TestCompileAndGenerateAllWritesDescriptorForTest(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "desc.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
action step = {}
export step
`), 0o644); err != nil {
		t.Fatal(err)
	}
	batch, err := CompileAndGenerateAll(spec, map[string]string{"target": "test", "classname": "Desc"}, Config{})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	if len(batch.Outputs) != 1 {
		t.Fatalf("outputs=%d", len(batch.Outputs))
	}
	if got := batch.ExtraFiles["Desc.dsc"]; !strings.Contains(got, `"test_params"`) {
		t.Fatalf("descriptor missing test params: %q", got)
	}
}

func outSource(out *Output) string {
	if out == nil {
		return "<nil>"
	}
	return out.Source
}

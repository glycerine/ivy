package ivy2golang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
	ivy2cppgen "github.com/glycerine/ivy/goivy/ivy2cpp"
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

func newGoTestGeneratorWithInterps(interps map[string]string) *Generator {
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Sig = goivy.NewSigOn(mod.Cfg.IuCfg)
	for name, interp := range interps {
		mod.Sig.Interp[name] = interp
	}
	return &Generator{Mod: mod}
}

func findFirstAssertInAction(a goivy.Action) *goivy.LogicAssertAction {
	if a == nil {
		return nil
	}
	if as, ok := a.(*goivy.LogicAssertAction); ok {
		return as
	}
	if iter, ok := a.(interface {
		IterSubactions() []goivy.ActionsAction
	}); ok {
		for _, sub := range iter.IterSubactions() {
			if as, ok := sub.(*goivy.LogicAssertAction); ok {
				return as
			}
		}
	}
	return nil
}

func bodyAfterMarker(s, marker string) string {
	i := strings.Index(s, marker)
	if i < 0 {
		return ""
	}
	open := strings.Index(s[i:], "{")
	if open < 0 {
		return ""
	}
	start := i + open
	depth := 0
	for j := start; j < len(s); j++ {
		switch s[j] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : j+1]
			}
		}
	}
	return s[start:]
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

func TestConjectureAppendsAssertToPublicActions(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
relation r
after init { r := false }
action step = { r := true }
export step
conjecture r
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "conj_actions", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	stepAct, ok := mod.Actions.Get2("step")
	if !ok {
		t.Fatalf("step action missing after prepareModuleForGo")
	}
	if findFirstAssertInAction(stepAct) == nil {
		t.Fatalf("expected conjecture assert appended to exported action step; not found in:\n%v", stepAct)
	}
	if !strings.Contains(out.Source, "ivyAssert(") {
		t.Fatalf("expected ivyAssert(...) in generated Go after conjecture insertion:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestConjectureAddsCheckInvariantsInitializer(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
relation r
after init { r := false }
action step = { r := true }
export step
conjecture r
`)
	if _, err := Generate(mod, Config{Target: "test", ClassName: "conj_init", TestIters: "1"}); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	found := false
	for _, init := range mod.Initializers {
		if init.Name == "__check_invariants" {
			found = true
			if findFirstAssertInAction(init.Action) == nil {
				t.Fatalf("__check_invariants initializer is present but contains no assert")
			}
			break
		}
	}
	if !found {
		names := make([]string, 0, len(mod.Initializers))
		for _, init := range mod.Initializers {
			names = append(names, init.Name)
		}
		t.Fatalf("__check_invariants initializer not found; have %v", names)
	}
}

func TestPropertyMovedIntoAxioms(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
relation r
after init { r := true }
property r
action step = {}
export step
`)
	beforeProps := len(mod.LabeledProps)
	beforeAxioms := len(mod.LabeledAxioms)
	if beforeProps == 0 {
		t.Fatalf("test setup: expected at least one LabeledProp before Generate, got 0")
	}
	if _, err := Generate(mod, Config{Target: "test", ClassName: "prop_axiom", TestIters: "1"}); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(mod.LabeledProps) != 0 {
		t.Fatalf("expected LabeledProps cleared after prepareModuleForGo, got %d", len(mod.LabeledProps))
	}
	if len(mod.LabeledAxioms) != beforeAxioms+beforeProps {
		t.Fatalf("expected LabeledAxioms to grow by %d; was %d, now %d",
			beforeProps, beforeAxioms, len(mod.LabeledAxioms))
	}
}

func TestConjectureLinenoPropagatesToAssert(t *testing.T) {
	src := "#lang ivy1.7\n" +
		"relation r\n" +
		"after init { r := false }\n" +
		"action step = { r := true }\n" +
		"export step\n" +
		"conjecture r\n"
	const conjLine = 6
	mod := compileIvySource(t, src)
	if len(mod.LabeledConjs) == 0 {
		t.Fatalf("test setup: expected at least one LabeledConj, got 0")
	}
	if got := mod.LabeledConjs[0].GetLineno().Line; got != conjLine {
		t.Fatalf("test setup: expected conjecture on line %d, got %d", conjLine, got)
	}
	out, err := Generate(mod, Config{Target: "test", ClassName: "conj_lineno", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	stepAct, ok := mod.Actions.Get2("step")
	if !ok {
		t.Fatalf("step action missing")
	}
	as := findFirstAssertInAction(stepAct)
	if as == nil {
		t.Fatalf("no assert appended to step")
	}
	if got := as.GetLineno().Line; got != conjLine {
		t.Fatalf("appended assert lineno: want %d, got %d", conjLine, got)
	}
	wantFragment := fmt.Sprintf("line %d", conjLine)
	if !strings.Contains(out.Source, wantFragment) {
		t.Fatalf("expected generated Go to reference conjecture %q:\n%s", wantFragment, out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestGeneratedConjectureAssertFailsAfterPublicAction(t *testing.T) {
	src := "#lang ivy1.7\n" +
		"relation r\n" +
		"after init { r := true }\n" +
		"action step = { r := false }\n" +
		"export step\n" +
		"conjecture r\n"
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "conj_runtime", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err == nil {
		t.Fatalf("expected generated run to fail conjecture\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	for _, want := range []string{"assertion failed", "line 6"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr missing %q after conjecture failure\nstdout:\n%s\nstderr:\n%s", want, stdout, stderr)
		}
	}
	if !strings.Contains(stdout, `assertion_failed("test.ivy: line 6")`) {
		t.Fatalf("stdout missing assertion_failed event after conjecture failure\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGeneratedAssumeFailureEmitsTraceEvent(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step returns(ok:bool) = {
    assume false;
    ok := true
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "assume_runtime", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, `ivyAssume(false, "test.ivy: line 3")`) {
		t.Fatalf("assume action should carry source-line label:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err == nil {
		t.Fatalf("expected generated run to fail assumption\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stdout, `assumption_failed("test.ivy: line 3")`) {
		t.Fatalf("stdout missing assumption_failed event\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	for _, want := range []string{"assumption failed", "line 3"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("stderr missing %q after assumption failure\nstdout:\n%s\nstderr:\n%s", want, stdout, stderr)
		}
	}
}

func TestSubgoalNeverConvertedToAssume(t *testing.T) {
	cfg := goivy.NewConfig()
	fmla := goivy.NewConst("p", goivy.Boolean)
	sub := goivy.NewSubgoalAction(fmla)

	kinds := map[string]bool{"assert": true, "require": true}
	out := goivy.AssertToAssume(sub, kinds, cfg.IuCfg)
	if _, ok := out.(*goivy.LogicSubgoalAction); !ok {
		t.Fatalf("AssertToAssume converted SubgoalAction to %T", out)
	}

	out = goivy.AssertToAssume(sub, map[string]bool{"subgoal": true}, cfg.IuCfg)
	if _, ok := out.(*goivy.LogicAssumeAction); !ok {
		t.Fatalf("AssertToAssume with subgoal opt-in returned %T", out)
	}
}

func TestRequiresInExternalBecomesAssume(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step(c:color) returns(out:color) = {
    require c = green;
    saved := c;
    out := saved
}
export step
`)
	stepAct, ok := mod.Actions.Get2("step")
	if !ok {
		t.Fatalf("step action missing")
	}
	extAct := goivy.AssertToAssume(stepAct, map[string]bool{"require": true}, mod.Cfg.IuCfg)
	mod.Actions.Set("ext:step", extAct)
	mod.PublicActions = goivy.NewInsMap[string, bool]()
	mod.PublicActions.Set("ext:step", true)
	mod.PublicActions.Set("step", false)

	out, err := Generate(mod, Config{Target: "test", ClassName: "reqassume", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	extBody := bodyAfterMarker(out.Source, "func (ivy *reqassume) ext__step")
	intBody := bodyAfterMarker(out.Source, "func (ivy *reqassume) step")
	if extBody == "" {
		t.Fatalf("ext:step method not emitted:\n%s", out.Source)
	}
	if intBody == "" {
		t.Fatalf("internal step method not emitted:\n%s", out.Source)
	}
	if !strings.Contains(extBody, "ivyAssume(") {
		t.Fatalf("ext:step body missing ivyAssume(...):\n%s", extBody)
	}
	if !strings.Contains(intBody, "ivyAssert(") {
		t.Fatalf("internal step body missing ivyAssert(...):\n%s", intBody)
	}
	if !strings.Contains(extBody, "return out") {
		t.Fatalf("assume early exit in returning action should return current outputs:\n%s", extBody)
	}
	compileGeneratedGo(t, out)
}

func TestTestMainBareFinalizeIsOrdinaryAction(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action _finalize = {}
export _finalize
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "fin", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if strings.Contains(out.Source, "ivy.ext___finalize()") {
		t.Fatalf("bare _finalize should not emit the ext:_finalize end-of-run hook:\n%s", out.Source)
	}
	if !strings.Contains(out.Source, "ivy._finalize()") {
		t.Fatalf("bare _finalize should remain an ordinary generated action:\n%s", out.Source)
	}
	if !strings.Contains(out.Source, `fmt.Fprintln(__ivy_out, "> _finalize")`) {
		t.Fatalf("bare _finalize should remain in the randomized action loop:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestTestMainHonorsExtFinalize(t *testing.T) {
	mod := goivy.New()
	mod.Name = "fin"
	mod.Actions.Set("ext:_finalize", goivy.NewSequence())
	mod.PublicActions.Set("ext:_finalize", true)
	mod.Actions.Set("step", goivy.NewSequence())
	mod.PublicActions.Set("step", true)
	out, err := Generate(mod, Config{Target: "test", ClassName: "fin", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.ext___finalize()") {
		t.Fatalf("expected ext:_finalize hook in test main:\n%s", out.Source)
	}
	if strings.Contains(out.Source, `fmt.Fprintln(__ivy_out, "> _finalize")`) {
		t.Fatalf("ext:_finalize should not be in the randomized action loop:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestPublicMultipleReturnActionRejected(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action split(c:color) returns (out:color, good:bool) = {
    out := c;
    good := true
}
export split
`)
	_, err := Generate(mod, Config{Target: "test", ClassName: "multi_public", TestIters: "1"})
	if err == nil {
		t.Fatal("Generate should reject exporting a multi-return action")
	}
	if !strings.Contains(err.Error(), "cannot handle multiple output in exported actions: split") {
		t.Fatalf("expected multi-output rejection error, got: %v", err)
	}
}

func TestConstructorInitializesModuleParams(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter initial : color
individual saved : color
after init {
    saved := initial
}
action step = {
    assert saved = initial
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "paramcase", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"initial color",
		"ivy.initial = initial",
		"ivy.saved = ivy.initial",
		"p__initial := red",
		`p__initial = green`,
		`if len(rest) == 2 {`,
		`os.Open(rest[1])`,
		`rest = rest[:1]`,
		`usage: paramcase initial`,
		"ivy := newParamcase(p__initial)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("parameterized constructor source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "green", "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go with positional param: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte(""), 0o644); err != nil {
		t.Fatalf("write command file: %v", err)
	}
	stdout, stderr, err = runBinary(t, bin, "green", "iters=1", "runs=1", "seed=1", commandFile)
	if err != nil {
		t.Fatalf("run generated Go with positional param and command file: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker with command file:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	stdout, stderr, err = runBinary(t, bin, "green", commandFile, commandFile, "iters=0", "runs=1")
	if err == nil {
		t.Fatalf("extra command-file arg should fail usage\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "usage: paramcase initial") {
		t.Fatalf("extra command-file arg missing usage\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestMainParsesDefaultedModuleParamOverride(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter initial : color = red
individual saved : color
after init {
    saved := initial
}
action check = {
    assert saved = green
}
export check
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "paramdefault", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"p__initial := red",
		`if __ivy_param, ok := opts["initial"]; ok {`,
		`delete(opts, "initial")`,
		"ivy := newParamdefault(p__initial)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("defaulted parameter source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "initial=green", "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go with parameter override: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	stdout, stderr, err = runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err == nil {
		t.Fatalf("default red should violate green assertion without override\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "assertion failed") {
		t.Fatalf("expected default run to fail assertion\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGeneratedTestMainHonorsSpecialRuntimeOptions(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "runtimeopts", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`"time"`,
		`sleepMs := ivyAtoiDefault(opts["delay"], 10)`,
		`finalMs := ivyAtoiDefault(opts["wait"], 0)`,
		`if modelfile := opts["modelfile"]; modelfile != "" {`,
		`time.Sleep(time.Duration(sleepMs) * time.Millisecond)`,
		`time.Sleep(time.Duration(finalMs) * time.Millisecond)`,
		`if runidx == runs-1 {`,
		`time.Sleep(50 * time.Millisecond)`,
		`os.Exit(0)`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("runtime option source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	modelFile := filepath.Join(t.TempDir(), "model.out")
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1", "delay=1", "wait=1", "modelfile="+modelFile)
	if err != nil {
		t.Fatalf("run generated Go with runtime opts: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if _, err := os.Stat(modelFile); err != nil {
		t.Fatalf("modelfile option did not create file %s: %v", modelFile, err)
	}
}

func TestGeneratedTestMainTracksGeneratingFlag(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual _generating : bool
after init {
    _generating := false
}
action step = {
    assert _generating
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "genflag", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`_ = ivy.___ivy_randomize(2, "init._generating", 0)`,
		`ivy._generating = ivy.___ivy_choose(0, "init", 0) != 0`,
		`ivy._generating = false`,
		`ivy._generating = true`,
		`ivy.step()`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("_generating source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `ivy._generating = ivy.___ivy_choose(2, "init._generating", 0) != 0`) {
		t.Fatalf("_generating initial state should use C++ choose(0, \"init\", 0) shape:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go with _generating guard: %v\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", err, stdout, stderr, out.Source)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGeneratedTestMainHandlesZeroParamCommandFileArg(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "cmdfile", TestIters: "0"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`opts, rest := parseIvyTestArgs(os.Args[1:])`,
		`if len(rest) == 1 {`,
		`os.Open(rest[0])`,
		`cannot open to read: %s`,
		`if len(rest) != 0 {`,
		`usage: cmdfile`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("zero-param command-file source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte(""), 0o644); err != nil {
		t.Fatalf("write command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, "iters=0", "runs=1", commandFile)
	if err != nil {
		t.Fatalf("run generated Go with command file arg: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	stdout, stderr, err = runBinary(t, bin, "iters=0", "runs=1", commandFile, commandFile)
	if err == nil {
		t.Fatalf("extra positional arg should fail usage\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "usage: cmdfile") {
		t.Fatalf("extra positional arg missing usage\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	stdout, stderr, err = runBinary(t, bin, "iters=0", "runs=1", filepath.Join(t.TempDir(), "missing.in"))
	if err == nil {
		t.Fatalf("missing command file should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "cannot open to read:") {
		t.Fatalf("missing command file should report open failure\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGeneratedSmokeMainHandlesZeroParamArgsLikeIvy2Cpp(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "smokerepl"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`opts, rest := parseIvyTestArgs(os.Args[1:])`,
		`if len(rest) == 1 {`,
		`os.Open(rest[0])`,
		`if len(rest) != 0 {`,
		`usage: smokerepl`,
		`for key := range opts {`,
		`unknown option: %s`,
		`ivy := newSmokerepl()`,
		`ivy.__argv = append([]string{os.Args[0]}, rest...)`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("zero-param smoke-main source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte(""), 0o644); err != nil {
		t.Fatalf("write command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err != nil {
		t.Fatalf("run generated smoke main with command file arg: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	stdout, stderr, err = runBinary(t, bin, "unexpected=1")
	if err == nil {
		t.Fatalf("unknown option should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "unknown option: unexpected") {
		t.Fatalf("unknown option missing diagnostic\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	stdout, stderr, err = runBinary(t, bin, commandFile, commandFile)
	if err == nil {
		t.Fatalf("extra positional arg should fail usage\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "usage: smokerepl") {
		t.Fatalf("extra positional arg missing usage\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGeneratedSmokeMainHonorsSpecialRuntimeOptions(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
after init {
    saved := green
}
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "smokeopts", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`seed := ivyAtoiDefault(opts["seed"], 1)`,
		`ivySetSeed(seed)`,
		`if out := opts["out"]; out != "" {`,
		`__ivy_out = f`,
		`if modelfile := opts["modelfile"]; modelfile != "" {`,
		`ivy := newSmokeopts()`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("smoke runtime option source missing %q:\n%s", want, out.Source)
		}
	}
	if seedIdx, newIdx := strings.Index(out.Source, `ivySetSeed(seed)`), strings.Index(out.Source, `ivy := newSmokeopts()`); seedIdx < 0 || newIdx < 0 || seedIdx > newIdx {
		t.Fatalf("seed setup should happen before construction:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	dir := t.TempDir()
	traceFile := filepath.Join(dir, "trace.out")
	modelFile := filepath.Join(dir, "model.out")
	stdout, stderr, err := runBinary(t, bin, "seed=2", "out="+traceFile, "modelfile="+modelFile)
	if err != nil {
		t.Fatalf("run smoke main with runtime opts: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "" {
		t.Fatalf("out= should redirect smoke trace output away from stdout, got %q", stdout)
	}
	data, err := os.ReadFile(traceFile)
	if err != nil {
		t.Fatalf("read trace output: %v", err)
	}
	if got := string(data); !strings.Contains(got, "write(saved,green)") {
		t.Fatalf("out= file missing initializer trace: %q", got)
	}
	if _, err := os.Stat(modelFile); err != nil {
		t.Fatalf("modelfile option did not create file %s: %v", modelFile, err)
	}
}

func TestTargetGenEmitsRunnableRandomizedMain(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "genrunner", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if out.Target != "gen" || out.EffectiveTarget != "gen" || !out.EmitMain {
		t.Fatalf("gen target metadata = target %q effective %q emitMain %v", out.Target, out.EffectiveTarget, out.EmitMain)
	}
	for _, want := range []string{
		`"time"`,
		`type Genrunner_init_gen struct`,
		`func (gen *Genrunner_init_gen) generate() bool`,
		`init_generator := &Genrunner_init_gen{ivy: ivy}`,
		`init_generator.generate()`,
		`type Genrunner_step_generator struct`,
		`func (gen *Genrunner_step_generator) generate() bool`,
		`func (gen *Genrunner_step_generator) execute()`,
		`step_generator := &Genrunner_step_generator{ivy: ivy}`,
		`if step_generator.generate() {`,
		`step_generator.execute()`,
		`func ivy2golangGenerate(ivy *genrunner, testIters int, sleepMs int)`,
		`ivy2golangGenerate(ivy, testIters, sleepMs)`,
		`testIters := ivyAtoiDefault(opts["iters"], 1)`,
		`for cycle := 0; cycle < testIters; cycle++ {`,
		`__choice := float64(ivyRand31()) * __choices / 2147483648.0`,
		`fmt.Fprintln(__ivy_out, "test_completed")`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("gen target source missing randomized runner fragment %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `_ = newgenrunner()`) {
		t.Fatalf("gen target should not emit the smoke-only main:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go gen target: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestCompileAndGenerateDefaultTargetGenRunsRandomizedMain(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "defaultgen.ivy")
	if err := os.WriteFile(file, []byte(`#lang ivy1.7
action step = {
}
export step
`), 0o644); err != nil {
		t.Fatalf("write ivy source: %v", err)
	}
	batch, err := CompileAndGenerateAll(file, nil, Config{TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	if batch.Config.RequestedTarget != "gen" || batch.Config.Target != "gen" || !batch.Config.EmitMain {
		t.Fatalf("default target metadata = requested %q effective %q emitMain %v", batch.Config.RequestedTarget, batch.Config.Target, batch.Config.EmitMain)
	}
	if len(batch.Outputs) != 1 {
		t.Fatalf("expected one output, got %d", len(batch.Outputs))
	}
	out := batch.Outputs[0]
	for _, want := range []string{
		`for cycle := 0; cycle < testIters; cycle++ {`,
		`fmt.Fprintln(__ivy_out, "test_completed")`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("default gen source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run default gen output: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestCompileAndGenerateAllCrossVersionSameProcess(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name      string
		className string
		source    string
	}{
		{
			name:      "ivy16",
			className: "Cross16",
			source: `#lang ivy1.6
action step = {
}
export step
`,
		},
		{
			name:      "ivy17",
			className: "Cross17",
			source: `#lang ivy1.7
action step = {
}
export step
`,
		},
		{
			name:      "ivy18",
			className: "Cross18",
			source: `#lang ivy1.8
action step = {
}
export step
`,
		},
		{
			name:      "ivy16_again",
			className: "Cross16Again",
			source: `#lang ivy1.6
action step = {
}
export step
`,
		},
	}
	for _, tc := range cases {
		path := filepath.Join(dir, tc.name+".ivy")
		if err := os.WriteFile(path, []byte(tc.source), 0o644); err != nil {
			t.Fatalf("write %s: %v", tc.name, err)
		}
		batch, err := CompileAndGenerateAll(path, map[string]string{"target": "test", "classname": tc.className}, Config{TestIters: "1", TestRuns: "1"})
		if err != nil {
			t.Fatalf("CompileAndGenerateAll %s: %v", tc.name, err)
		}
		if len(batch.Outputs) != 1 {
			t.Fatalf("%s outputs=%d, want 1", tc.name, len(batch.Outputs))
		}
		out := batch.Outputs[0]
		if out.ClassName != tc.className {
			t.Fatalf("%s class name = %q, want %q", tc.name, out.ClassName, tc.className)
		}
		bin := compileGeneratedGo(t, out)
		stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
		if err != nil {
			t.Fatalf("run %s generated Go: %v\nstdout:\n%s\nstderr:\n%s", tc.name, err, stdout, stderr)
		}
		if !strings.Contains(stdout, "test_completed") {
			t.Fatalf("%s output missing completion marker:\nstdout:\n%s\nstderr:\n%s", tc.name, stdout, stderr)
		}
	}
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
	for _, want := range []string{
		`"math/rand/v2"`,
		"rand.NewChaCha8(ivySeedBytes(1))",
		"return int(__ivy_rng.Uint64() >> 33)",
		"__choice := float64(ivyRand31()) * __choices / 2147483648.0",
		"cycle--",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing C++-parity RNG/dispatch shape %q:\n%s", want, out.Source)
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
	if !strings.Contains(stdout, "> set") {
		t.Fatalf("run output missing action trace:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if strings.Contains(out.Source, "ivy2golangGenerate") {
		t.Fatalf("target=test should not emit the target=gen helper:\n%s", out.Source)
	}
	traceFile := filepath.Join(t.TempDir(), "trace.out")
	stdout, stderr, err = runBinary(t, bin, "iters=1", "runs=1", "seed=1", "out="+traceFile)
	if err != nil {
		t.Fatalf("run generated Go with out=: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "" {
		t.Fatalf("out= should redirect trace output away from stdout, got %q", stdout)
	}
	data, err := os.ReadFile(traceFile)
	if err != nil {
		t.Fatalf("read trace output: %v", err)
	}
	if got := string(data); !strings.Contains(got, "> set") || !strings.Contains(got, "test_completed") {
		t.Fatalf("out= file missing trace/completion markers: %q", got)
	}
	stdout, stderr, err = runBinary(t, bin, "iters=0", "runs=2", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go with runs=2: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if got := strings.Count(stdout, "test_completed"); got != 2 {
		t.Fatalf("runs=2 should emit completion per run, got %d in stdout:\n%s", got, stdout)
	}
}

func TestGenerateWhileActionCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
individual flag : bool
after init {
    flag := true
}
action drain returns(out:bool) = {
    while flag {
        flag := false
    };
    out := flag
}
export drain
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "while_action", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func (ivy *while_action) drain() bool",
		"for ivy.flag {",
		"ivy.flag = false",
		"out = ivy.flag",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("while source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "unsupported action") {
		t.Fatalf("while action should not fall through to unsupported:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := strings.Repeat("> drain\n= false\n", 3) + "test_completed\n"
	if stdout != wantTrace {
		t.Fatalf("while trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

func TestGeneratedChoiceUsesIfElseChainCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green, blue}
individual saved : color
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	saved := goivy.NewConst("saved", color)
	red := goivy.NewConst("red", color)
	green := goivy.NewConst("green", color)
	blue := goivy.NewConst("blue", color)
	mod.Actions.Set("step", goivy.NewChoiceActionOn(goivy.NewActionsConfig(),
		goivy.NewAssignAction(saved, red),
		goivy.NewAssignAction(saved, green),
		goivy.NewAssignAction(saved, blue),
	))
	out, err := Generate(mod, Config{Target: "test", ClassName: "choice3", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"__ivy_branch",
		`ivy.___ivy_choose(0, "___branch"`,
		"if __ivy_branch",
		"} else if __ivy_branch",
		"} else {",
		"ivy.saved = red",
		"ivy.saved = green",
		"ivy.saved = blue",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("choice source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "switch ivy.___ivy_choose") {
		t.Fatalf("choice should use an if/else chain, not a switch:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestGeneratedVarActionUsesNondetLocalCompiles(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action step = {
    var tmp : color := green;
    saved := tmp
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "locals", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"loc__tmp := color(ivy.___ivy_choose(0, \"loc:tmp\"",
		"loc__tmp = green",
		"ivy.saved = loc__tmp",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("local source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `color(ivy.___ivy_choose(2, "loc:tmp"`) {
		t.Fatalf("local should use scalar nondet choose(0), not sort-range choose:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestGeneratedLocalFunctionUsesNondetLoopCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : bool
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	relSort, err := goivy.NewFunctionSort(color, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	marked := goivy.NewConst("marked", relSort)
	red := goivy.NewConst("red", color)
	body := goivy.NewAssignAction(goivy.NewConst("saved", goivy.Boolean), goivy.MustApply(marked, red))
	mod.Actions.Set("step", goivy.NewLocalActionOn(goivy.NewActionsConfig(), "test", marked, body))
	out, err := Generate(mod, Config{Target: "test", ClassName: "locfunc", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"marked := make(map[color]bool)",
		"for _, __i0 := range []color{red, green} {",
		`marked[__i0] = ivy.___ivy_choose(0, "marked", 0) != 0`,
		"ivy.saved = marked[red]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("local function source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `ivy.___ivy_choose(2, "marked"`) {
		t.Fatalf("local function cells should use scalar nondet choose(0), not bool-range choose:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestGeneratedLocalUnboundedFunctionUsesNondetThunkCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
individual saved : bool
action step(n:node) = {
    saved := false
}
export step
`)
	node, ok := mod.Sig.Sorts.Get2("node")
	if !ok {
		t.Fatal("missing node sort")
	}
	relSort, err := goivy.NewFunctionSort(node, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	marked := goivy.NewConst("marked", relSort)
	n := goivy.NewConst("n", node)
	app, err := goivy.NewApply(marked, n)
	if err != nil {
		t.Fatalf("NewApply marked: %v", err)
	}
	body := goivy.NewAssignAction(goivy.NewConst("saved", goivy.Boolean), app)
	step := goivy.NewLocalActionOn(goivy.NewActionsConfig(), "test", marked, body)
	step.SetFormalParams([]*goivy.Const{n})
	mod.Actions.Set("step", step)
	out, err := Generate(mod, Config{Target: "test", ClassName: "locunbounded", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"marked := newIvyThunkMap[int, bool](false)",
		"marked.base = func(__ivy_key int) bool {",
		`return ivy.___ivy_choose(0, "marked", 0) != 0`,
		"ivy.saved = marked.Get(n)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("local unbounded function source missing %q:\n%s", want, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestGeneratedLetActionSubstitutesBoundSymbolCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
relation other(C:color)
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	relSort, err := goivy.NewFunctionSort(color, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	marked := goivy.NewConst("marked", relSort)
	other := goivy.NewConst("other", relSort)
	red := goivy.NewConst("red", color)
	binding := goivy.NewEqualsNode(marked, other)
	mod.Actions.Set("step", goivy.NewLetAction(binding, goivy.NewSetAction(goivy.MustApply(marked, red))))
	out, err := Generate(mod, Config{Target: "test", ClassName: "lets", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.other[red] = true") {
		t.Fatalf("let binding did not substitute marked->other:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "ivy.marked[red] = true") {
		t.Fatalf("let body still writes original symbol:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
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
		"__ivy_tmp",
		"__ivy_tmp0[C] = true",
		"ivy.marked[C] = __ivy_tmp0[C]",
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

func TestQuantifiedAssignUsesTwoPhaseTemporary(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
relation edge(X:color,Y:color)
after init {
    edge(red,red) := false;
    edge(red,green) := true;
    edge(green,red) := false;
    edge(green,green) := false
}
action transpose = {
    edge(X,Y) := edge(Y,X);
    assert ~edge(red,green) & edge(green,red)
}
export transpose
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "transpose_assign", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"__ivy_tmp",
		"make(map[struct{ A0 color; A1 color }]bool)",
		"ivy.edge[struct{ A0 color; A1 color }{Y, X}]",
		"ivy.edge[struct{ A0 color; A1 color }{X, Y}]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("two-phase quantified assignment source missing %q:\n%s", want, out.Source)
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

func TestQuantifiedAssignUsesGuardBounds(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..5}
relation marked(I:idx)
individual limit : idx
action step = {
}
export step
`)
	idx, ok := mod.Sig.Sorts.Get2("idx")
	if !ok {
		t.Fatal("missing idx sort")
	}
	marked, err := mod.Sig.FindSymbol("marked", false)
	if err != nil {
		t.Fatalf("FindSymbol marked: %v", err)
	}
	iVar, err := goivy.NewVariable("I", idx)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	markedApp, err := goivy.NewApply(goivy.NewConst("marked", marked.CSort), iVar)
	if err != nil {
		t.Fatalf("NewApply marked: %v", err)
	}
	ltSort, err := goivy.NewFunctionSort(idx, idx, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort <: %v", err)
	}
	cond, err := goivy.NewApply(goivy.NewConst("<", ltSort), iVar, goivy.NewConst("limit", idx))
	if err != nil {
		t.Fatalf("NewApply <: %v", err)
	}
	ite, err := goivy.NewIte(cond, goivy.NewConst("true", goivy.Boolean), markedApp)
	if err != nil {
		t.Fatalf("NewIte: %v", err)
	}
	mod.Actions.Set("step", goivy.NewAssignAction(markedApp, ite))
	out, err := Generate(mod, Config{Target: "test", ClassName: "guard_bounds", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for I := 0; I < ivy.limit; I++",
		"__ivy_tmp0[I] = ivyTernary((I < ivy.limit), true, ivy.marked[I])",
		"ivy.marked[I] = __ivy_tmp0[I]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("bounded quantified assignment source missing %q:\n%s", want, out.Source)
		}
	}
	stepBody := out.Source
	if idx := strings.Index(stepBody, "func (ivy *guard_bounds) step()"); idx >= 0 {
		stepBody = stepBody[idx:]
	}
	if strings.Contains(stepBody, "range []int{0, 1, 2, 3, 4, 5}") {
		t.Fatalf("guarded assignment should use bounded integer loop, not full range enumeration:\n%s", stepBody)
	}
	compileGeneratedGo(t, out)
}

func TestGenerateUnboundedAllFalseRelationAssignClearsMap(t *testing.T) {
	src := `#lang ivy1.7
type key
relation seen(K:key)
action clear_all = {
    seen(K) := false
}
export clear_all
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "hash_thunk_assign_direct", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"seen ivyThunkMap[int, bool]",
		"ivy.seen = newIvyThunkMap[int, bool](false)",
		"func (ivy *hash_thunk_assign_direct) clear_all()",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "unsupported assignment over free variable") ||
		strings.Contains(out.Source, "cannot enumerate function domain sort") {
		t.Fatalf("unbounded all-false relation reset should not fall through to finite enumeration:\n%s", out.Source)
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

func TestUnboundedInitialStateUsesNondetThunkMap(t *testing.T) {
	src := `#lang ivy1.7
type node
relation marked(N:node)
action read(n:node) returns(out:bool) = {
    out := marked(n)
}
export read
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "unbounded_init", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"marked ivyThunkMap[int, bool]",
		"ivy.marked = newIvyThunkMap[int, bool](false)",
		"ivy.marked.base = func(__ivy_key int) bool {",
		`return ivy.___ivy_choose(0, "init", 0) != 0`,
		"m.overrides[key] = value",
		"out = ivy.marked.Get(n)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("unbounded initial thunk source missing %q:\n%s", want, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestUnboundedQuantifiedAssignUsesThunkMap(t *testing.T) {
	src := `#lang ivy1.7
type node
relation marked(N:node)
action flip(n:node) = {
    marked(n) := true;
    marked(X) := ~marked(X);
    assert ~marked(n)
}
export flip
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "unbounded_flip", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"marked ivyThunkMap[int, bool]",
		"ivy.marked.Set(n, true)",
		"__ivy_old_marked",
		"ivy.marked.base = func(",
		"return !(__ivy_old_marked",
		"ivy.marked.Get(n)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("unbounded assignment source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", err, stdout, stderr, out.Source)
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
		"out := red",
		"ivy.saved = c",
		"__arg0 := color(ivy.___ivy_randomize",
		`fmt.Fprintf(__ivy_out, "= %v\n", ivy.choose(__arg0))`,
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
	if !strings.Contains(stdout, "> choose(") || !strings.Contains(stdout, "= ") {
		t.Fatalf("run output missing action/return trace:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestTraceLHSRuntimeOutputCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action step = {
    saved := green
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "trace_simple", Trace: true, TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`fmt.Fprintf(__ivy_out, "  write(%s,%s)\n", "saved", ivyTraceValue(green, false))`,
		"ivy.saved = green",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("trace source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := "> step\n  write(saved,green)\ntest_completed\n"
	if stdout != wantTrace {
		t.Fatalf("write trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

func TestTraceLHSFunctionApplicationArgs(t *testing.T) {
	src := `#lang ivy1.7
type idx = {i0, i1}
type color = {red, green}
function slot(I:idx) : color
action step(i:idx,c:color) = {
    slot(i) := c
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "trace_func", Trace: true, TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	want := `fmt.Fprintf(__ivy_out, "  write(%s,%s)\n", "slot" + "(" + ivyTraceValue(i, false) + ")", ivyTraceValue(c, false))`
	if !strings.Contains(out.Source, want) {
		t.Fatalf("function-application trace missing %q:\n%s", want, out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestTraceLHSRespectsHexNumberFormat(t *testing.T) {
	src := `#lang ivy1.7
type idx = {0..20}
individual x : idx
action step = {
    x := 10
}
export step
`
	mod := compileIvySource(t, src)
	mod.SetAttribute("radix", "16")
	out, err := Generate(mod, Config{Target: "test", ClassName: "trace_hex", Trace: true, TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	want := `fmt.Fprintf(__ivy_out, "  write(%s,%s)\n", "x", ivyTraceValue(10, true))`
	if !strings.Contains(out.Source, want) {
		t.Fatalf("hex trace source missing %q:\n%s", want, out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := "> step\n  write(x,0xa)\ntest_completed\n"
	if stdout != wantTrace {
		t.Fatalf("hex write trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

func TestTraceLHSDestructorChain(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
type inner
type outer
destructor child(O:outer) : inner
destructor shade(I:inner) : color
individual root : outer
action step(c:color) = {
    root.child.shade := c
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "trace_field", Trace: true, TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	want := `fmt.Fprintf(__ivy_out, "  write(%s,%s)\n", "root" + ".child" + ".shade", ivyTraceValue(c, false))`
	if !strings.Contains(out.Source, want) {
		t.Fatalf("destructor-chain trace missing %q:\n%s", want, out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestTraceLHSSuppressesNamespacedLocalAndTraceOff(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action step = {
    var tmp : color := green;
    tmp := red;
    saved := tmp
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "trace_local", Trace: true, TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if strings.Contains(out.Source, `write(%s,%s)\n", "loc:tmp"`) ||
		strings.Contains(out.Source, `write(%s,%s)\n", "loc__tmp"`) {
		t.Fatalf("namespaced local assignment should not emit write trace:\n%s", out.Source)
	}
	if !strings.Contains(out.Source, `fmt.Fprintf(__ivy_out, "  write(%s,%s)\n", "saved", ivyTraceValue(loc__tmp, false))`) {
		t.Fatalf("control saved assignment trace missing:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)

	mod2 := compileIvySource(t, src)
	out2, err := Generate(mod2, Config{Target: "test", ClassName: "trace_off", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate trace off: %v\n%s", err, outSource(out2))
	}
	if strings.Contains(out2.Source, `write(%s,%s)`) {
		t.Fatalf("write trace emitted with Trace=false:\n%s", out2.Source)
	}
	compileGeneratedGo(t, out2)
}

func TestImportCallerTracePrologueInTest(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual seen : color
action callback(c:color) = {
    seen := c
}
import callback
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "traceimp", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *traceimp) callback(")
	if body == "" {
		t.Fatalf("callback method not emitted:\n%s", out.Source)
	}
	if !strings.Contains(body, `fmt.Fprint(__ivy_out, "< callback"`) {
		t.Fatalf("expected import caller trace prologue in test body:\n%s", body)
	}
	if !strings.Contains(body, `ivyTraceValue(c, false)`) {
		t.Fatalf("expected import caller trace to format callback argument:\n%s", body)
	}
	compileGeneratedGo(t, out)

	mod2 := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual seen : color
action callback(c:color) = {
    seen := c
}
import callback
`)
	out2, err := Generate(mod2, Config{Target: "repl", ClassName: "traceimp2"})
	if err != nil {
		t.Fatalf("Generate repl: %v\n%s", err, outSource(out2))
	}
	body2 := bodyAfterMarker(out2.Source, "func (ivy *traceimp2) callback(")
	if strings.Contains(body2, `fmt.Fprint(__ivy_out, "< callback"`) {
		t.Fatalf("trace prologue should be test-only; appeared in repl target:\n%s", body2)
	}
}

func TestImportCallerBodyWrappedInBracesUnderTrace(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual seen : color
action callback(c:color) = {
    seen := c
}
import callback
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "tracedimp", Trace: true, TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *tracedimp) callback(")
	if body == "" {
		t.Fatalf("callback method not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`fmt.Fprint(__ivy_out, "< callback"`,
		`fmt.Fprintln(__ivy_out, "{")`,
		`fmt.Fprintln(__ivy_out, "}")`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in traced callback body:\n%s", want, body)
		}
	}
	compileGeneratedGo(t, out)
}

func TestTargetTestImportedCallbackCallIsNoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action notify = {
    assert false
}
import notify
action trigger = {
    call notify
}
export trigger
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "importnoop", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *importnoop) trigger(")
	if body == "" {
		t.Fatalf("trigger method not emitted:\n%s", out.Source)
	}
	if strings.Contains(body, "ivy.notify(") {
		t.Fatalf("target=test trigger should not call imported callback body:\n%s", body)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("imported callback call should be a no-op in target=test\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", stdout, stderr, out.Source)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if strings.Contains(stdout, "assertion_failed") || strings.Contains(stderr, "assertion failed") {
		t.Fatalf("imported callback body appears to have executed\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestEnumDispatchTraceMatchesIvy2Cpp(t *testing.T) {
	fixture := filepath.Join("..", "ivy2cpp", "test_vec", "oracle", "enum_dispatch.ivy")
	cppOut, err := ivy2cppgen.CompileAndGenerate(fixture, map[string]string{"target": "test", "classname": "enum_dispatch"}, ivy2cppgen.Config{})
	if err != nil {
		t.Fatalf("ivy2cpp CompileAndGenerate: %v", err)
	}
	cppBin, err := ivy2cppgen.BuildOutput(cppOut, t.TempDir())
	if err != nil {
		t.Skipf("ivy2cpp build unavailable for parity check: %v", err)
	}
	goOut, err := CompileAndGenerate(fixture, map[string]string{"target": "test", "classname": "enum_dispatch"}, Config{})
	if err != nil {
		t.Fatalf("ivy2golang CompileAndGenerate: %v\n%s", err, outSource(goOut))
	}
	goBin := compileGeneratedGo(t, goOut)
	args := []string{"iters=5", "runs=1", "seed=1"}
	cppStdout, cppStderr, err := runBinary(t, cppBin, args...)
	if err != nil {
		t.Fatalf("run ivy2cpp binary: %v\nstdout:\n%s\nstderr:\n%s", err, cppStdout, cppStderr)
	}
	goStdout, goStderr, err := runBinary(t, goBin, args...)
	if err != nil {
		t.Fatalf("run ivy2golang binary: %v\nstdout:\n%s\nstderr:\n%s", err, goStdout, goStderr)
	}
	if goStdout != cppStdout {
		t.Fatalf("ivy2golang trace differs from ivy2cpp\nivy2cpp stdout:\n%s\nivy2golang stdout:\n%s", cppStdout, goStdout)
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
		"(2 + ivy.___ivy_randomize(6",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=5", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := "> set(6)\n= 6\n> set(7)\n= 7\n> set(6)\n= 6\n> set(4)\n= 4\n> set(7)\n= 7\ntest_completed\n"
	if stdout != wantTrace {
		t.Fatalf("run output differs from ivy2cpp range trace:\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
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

func TestOracleHashThunkAssignCompilesAndRuns(t *testing.T) {
	fixture := filepath.Join("..", "ivy2cpp", "test_vec", "oracle", "hash_thunk_assign.ivy")
	out, err := CompileAndGenerate(fixture, map[string]string{"target": "test", "classname": "hash_thunk_assign"}, Config{TestIters: "1"})
	if err != nil {
		t.Fatalf("CompileAndGenerate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func (ivy *hash_thunk_assign) ext__clear_all()",
		"type hash_thunk_assign struct",
		"_generating bool",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "seen map[int]bool") {
		t.Fatalf("isolated oracle output should match ivy2cpp and omit pruned relation state:\n%s", out.Source)
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

func TestOracleNativeActionsCompileAsNoops(t *testing.T) {
	for _, name := range []string{"native_block", "callback_thunk"} {
		t.Run(name, func(t *testing.T) {
			fixture := filepath.Join("..", "ivy2cpp", "test_vec", "oracle", name+".ivy")
			out, err := CompileAndGenerate(fixture, map[string]string{"target": "test", "classname": name}, Config{TestIters: "1"})
			if err != nil {
				t.Fatalf("CompileAndGenerate: %v\n%s", err, outSource(out))
			}
			if len(out.Warnings) == 0 || !strings.Contains(out.Warnings[0], "native C++ action code") {
				t.Fatalf("expected native no-op warning, got %v", out.Warnings)
			}
			if !strings.Contains(out.Source, "ivy native action omitted") {
				t.Fatalf("source missing native no-op marker:\n%s", out.Source)
			}
			bin := compileGeneratedGo(t, out)
			stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
			if err != nil {
				t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
			}
			if !strings.Contains(stdout, "test_completed") {
				t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
			}
		})
	}
}

func TestOracleDestructorRecordCompilesAndMatchesTrace(t *testing.T) {
	fixture := filepath.Join("..", "ivy2cpp", "test_vec", "oracle", "destructor_record.ivy")
	out, err := CompileAndGenerate(fixture, map[string]string{"target": "test", "classname": "destructor_record"}, Config{TestIters: "1"})
	if err != nil {
		t.Fatalf("CompileAndGenerate: %v\n%s", err, outSource(out))
	}
	if len(out.Warnings) != 0 {
		t.Fatalf("destructor record should not use non-enumerable fallback warnings, got %v", out.Warnings)
	}
	for _, want := range []string{
		"type cell struct {",
		"shade color",
		"bits int",
		"func (v cell) String() string",
		"__arg0 := cell{shade: color(ivy.___ivy_randomize(2",
		"bits: ivy.___ivy_randomize(256",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=5", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := "> store({shade:red,bits:18})\n> store({shade:green,bits:116})\n> store({shade:red,bits:139})\n> store({shade:red,bits:12})\n> store({shade:green,bits:75})\ntest_completed\n"
	if stdout != wantTrace {
		t.Fatalf("destructor record trace differs from ivy2cpp\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

func TestGeneratedAssignFieldActionCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type cell
destructor shade(C:cell) : color
individual current : cell
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	cell, ok := mod.Sig.Sorts.Get2("cell")
	if !ok {
		t.Fatal("missing cell sort")
	}
	fieldSort, err := goivy.NewFunctionSort(cell, color)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	mod.Actions.Set("step", goivy.NewAssignFieldAction(
		goivy.NewConst("shade", fieldSort),
		goivy.NewConst("current", cell),
		goivy.NewConst("green", color),
	))
	out, err := Generate(mod, Config{Target: "test", ClassName: "assign_field", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.current.shade = green") {
		t.Fatalf("assign field action did not route through field assignment:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "unsupported action") {
		t.Fatalf("assign field action should not fall through to unsupported:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestGeneratedNullFieldActionCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type cell
destructor shade(C:cell) : color
individual current : cell
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	cell, ok := mod.Sig.Sorts.Get2("cell")
	if !ok {
		t.Fatal("missing cell sort")
	}
	fieldSort, err := goivy.NewFunctionSort(cell, color)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	mod.Actions.Set("step", goivy.NewNullFieldAction(
		goivy.NewConst("shade", fieldSort),
		goivy.NewConst("current", cell),
	))
	out, err := Generate(mod, Config{Target: "test", ClassName: "null_field", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.current.shade = red") {
		t.Fatalf("null field action did not assign the field zero value:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "unsupported action") {
		t.Fatalf("null field action should not fall through to unsupported:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestGeneratedCopyFieldActionCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type cell
destructor shade(C:cell) : color
individual current : cell
individual other : cell
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	cell, ok := mod.Sig.Sorts.Get2("cell")
	if !ok {
		t.Fatal("missing cell sort")
	}
	fieldSort, err := goivy.NewFunctionSort(cell, color)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	mod.Actions.Set("step", goivy.NewCopyFieldAction(
		goivy.NewConst("current", cell),
		goivy.NewConst("shade", fieldSort),
		goivy.NewConst("other", cell),
		goivy.NewConst("shade", fieldSort),
	))
	out, err := Generate(mod, Config{Target: "test", ClassName: "copy_field", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.current.shade = ivy.other.shade") {
		t.Fatalf("copy field action did not copy through field refs:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "unsupported action") {
		t.Fatalf("copy field action should not fall through to unsupported:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestGeneratedSetActionWithFreeVariableCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	marked, err := mod.Sig.FindSymbol("marked", false)
	if err != nil {
		t.Fatalf("FindSymbol marked: %v", err)
	}
	c, err := goivy.NewVariable("C", color)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	markedApp, err := goivy.NewApply(goivy.NewConst("marked", marked.CSort), c)
	if err != nil {
		t.Fatalf("NewApply marked: %v", err)
	}
	mod.Actions.Set("step", goivy.NewSetAction(markedApp))
	out, err := Generate(mod, Config{Target: "test", ClassName: "set_free", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for _, C := range []color{red, green}",
		"ivy.marked[C] = true",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("set action source missing %q:\n%s", want, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestEmitCrashActionEmitsNothing(t *testing.T) {
	crash := goivy.NewCrashAction(goivy.NewConst("anything", goivy.TopS))
	var w goWriter
	g := &Generator{}
	g.emitAction(&w, crash)
	if got := strings.TrimSpace(w.String()); got != "" {
		t.Fatalf("crash action should emit nothing, got %q", got)
	}
	if len(g.errs) != 0 {
		t.Fatalf("crash action should not record errors, got %v", g.errs)
	}
}

func TestKnownUnsupportedActionsReturnSpecificErrors(t *testing.T) {
	flag := goivy.NewConst("flag", goivy.Boolean)
	trueExpr := goivy.NewConst("true", goivy.Boolean)
	cases := []struct {
		name string
		act  goivy.Action
		want string
	}{
		{
			name: "bindolds",
			act:  goivy.NewBindOldsAction(goivy.NewAssignAction(flag, trueExpr)),
			want: "bindolds reached emit",
		},
		{
			name: "thunk",
			act:  goivy.NewThunkAction(goivy.NewConst("x", goivy.TopS)),
			want: "thunk reached emit",
		},
		{
			name: "instantiate",
			act:  goivy.NewInstantiateAction(goivy.NewConst("schema", goivy.TopS)),
			want: "instantiate reached emit",
		},
		{
			name: "ranking",
			act:  goivy.NewRanking(goivy.NewConst("rank_rel", goivy.TopS), flag),
			want: "ranking reached emit",
		},
		{
			name: "havoc",
			act:  goivy.NewHavocAction(flag),
			want: "havoc reached emit",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mod := goivy.New()
			mod.Name = tc.name
			mod.Actions.Set("step", tc.act)
			_, err := Generate(mod, Config{Target: "test", ClassName: tc.name, TestIters: "1"})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unexpected Generate error: %v", err)
			}
		})
	}
}

func TestUnsupportedExpressionsReturnSpecificErrors(t *testing.T) {
	globally, err := goivy.NewGlobally(nil, goivy.True)
	if err != nil {
		t.Fatalf("NewGlobally: %v", err)
	}
	named, err := goivy.NewNamedBinder("custom", nil, nil, goivy.True)
	if err != nil {
		t.Fatalf("NewNamedBinder: %v", err)
	}
	cases := []struct {
		name string
		expr goivy.Expr
		want string
	}{
		{
			name: "temporal",
			expr: globally,
			want: "temporal expression",
		},
		{
			name: "named binder",
			expr: named,
			want: "named binder",
		},
		{
			name: "native",
			expr: &goivy.LogicNativeExpr{CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)}},
			want: "native C++ expression",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := (&Generator{}).emitExpr(tc.expr)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unexpected emitExpr error: %v", err)
			}
		})
	}
}

func TestEmitRangeArithmeticClamps(t *testing.T) {
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "5"}}
	fnSort, err := goivy.NewFunctionSort(rng, rng, rng)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	app := goivy.MustApply(goivy.NewConst("+", fnSort), goivy.NewConst("4", rng), goivy.NewConst("4", rng))
	got, err := (&Generator{}).emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{"__x := (4) + (4)", "if __x < 0", "if 5 < __x"} {
		if !strings.Contains(got, want) {
			t.Fatalf("range arithmetic should clamp; missing %q in:\n%s", want, got)
		}
	}
}

func TestEmitCastToRangeClamps(t *testing.T) {
	src := &goivy.UninterpretedSort{Name: "int"}
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "1"}, Ub: goivy.NumeralBound{Value: "9"}}
	castSort, err := goivy.NewFunctionSort(src, rng)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	app := goivy.MustApply(goivy.NewConst("cast", castSort), goivy.NewConst("v", src))
	got, err := (&Generator{}).emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{"if v < 1", "if 9 < v", "return v"} {
		if !strings.Contains(got, want) {
			t.Fatalf("cast to range should clamp; missing %q in:\n%s", want, got)
		}
	}
}

func TestEmitNatMinusSaturates(t *testing.T) {
	g := newGoTestGeneratorWithInterps(map[string]string{"nat": "nat"})
	nat := &goivy.UninterpretedSort{Name: "nat"}
	fnSort, err := goivy.NewFunctionSort(nat, nat, nat)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	app := goivy.MustApply(goivy.NewConst("-", fnSort), goivy.NewConst("a", nat), goivy.NewConst("b", nat))
	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{"__a := a", "__b := b", "if __a < __b", "return 0", "return __a - __b"} {
		if !strings.Contains(got, want) {
			t.Fatalf("nat subtraction should saturate; missing %q in:\n%s", want, got)
		}
	}
}

func TestGeneratedRangeArithmeticClampCompiles(t *testing.T) {
	src := `#lang ivy1.7
type idx = {0..5}
individual x : idx
after init {
    x := 4
}
action bump returns(out:idx) = {
    x := x + 4;
    out := x
}
export bump
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "range_arith", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{"__x := (ivy.x) + (4)", "if 5 < __x", "ivy.x = func() int"} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("range arithmetic generated source missing %q:\n%s", want, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestGeneratedDebugActionPrintsNamedValues(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual flag : bool
after init {
    flag := true
}
action step = {
}
export step
`)
	dbg := goivy.NewDebugAction(
		goivy.NewConst(`"tick"`, goivy.TopS),
		goivy.NewConst("flag", goivy.Boolean),
	)
	dbg.WithNames = []string{"f"}
	mod.Actions.Set("step", dbg)
	out, err := Generate(mod, Config{Target: "test", ClassName: "debug_named", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`fmt.Fprintln(__ivy_out, "{")`,
		`fmt.Fprintf(__ivy_out, "    \"event\" : %q,\n", "tick")`,
		`fmt.Fprintf(__ivy_out, "    %q : ", "f")`,
		`fmt.Fprint(__ivy_out, ivy.flag)`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("debug source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	for _, want := range []string{
		`> step`,
		`{`,
		`    "event" : "tick",`,
		`    "f" : true,`,
		`}`,
		`test_completed`,
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("debug stdout missing %q:\nstdout:\n%s\nstderr:\n%s", want, stdout, stderr)
		}
	}
}

func TestGeneratedDebugActionPrintsQuantifiedValueLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..1}
relation marked(I:idx)
action step = {
}
export step
`)
	idx, ok := mod.Sig.Sorts.Get2("idx")
	if !ok {
		t.Fatal("missing idx sort")
	}
	relSort, err := goivy.NewFunctionSort(idx, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	iVar, err := goivy.NewVariable("I", idx)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	app, err := goivy.NewApply(goivy.NewConst("marked", relSort), iVar)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	dbg := goivy.NewDebugAction(goivy.NewConst(`"scan"`, goivy.TopS), app)
	dbg.WithNames = []string{"seen"}
	mod.Actions.Set("step", dbg)
	out, err := Generate(mod, Config{Target: "test", ClassName: "debug_quant", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`fmt.Fprint(__ivy_out, "[")`,
		`range []int{0, 1}`,
		`fmt.Fprint(__ivy_out, ivy.marked[I])`,
		`fmt.Fprint(__ivy_out, "]")`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("quantified debug source missing %q:\n%s", want, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestInitialConditionEqualityEmitsAssignment(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	eq, err := goivy.NewEq(goivy.NewConst("saved", color), goivy.NewConst("green", color))
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(eq, nil)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initeq", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.saved = green") {
		t.Fatalf("initial equality did not emit assignment:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestInitialConditionRelationIffEmitsLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	marked, err := mod.Sig.FindSymbol("marked", false)
	if err != nil {
		t.Fatalf("FindSymbol marked: %v", err)
	}
	c, err := goivy.NewVariable("C", color)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	markedApp, err := goivy.NewApply(goivy.NewConst("marked", marked.CSort), c)
	if err != nil {
		t.Fatalf("NewApply marked: %v", err)
	}
	isGreen, err := goivy.NewEq(c, goivy.NewConst("green", color))
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	init, err := goivy.NewIff(markedApp, isGreen)
	if err != nil {
		t.Fatalf("NewIff: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(init, nil)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initrel", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for _, C := range []color{red, green}",
		"__ivy_tmp0[C] = (C == green)",
		"ivy.marked[C] = __ivy_tmp0[C]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("initial relation source missing %q:\n%s", want, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestInconsistentInitialConditionReturnsError(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	eqRed, err := goivy.NewEq(goivy.NewConst("saved", color), goivy.NewConst("red", color))
	if err != nil {
		t.Fatalf("NewEq red: %v", err)
	}
	eqGreen, err := goivy.NewEq(goivy.NewConst("saved", color), goivy.NewConst("green", color))
	if err != nil {
		t.Fatalf("NewEq green: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(&goivy.LogicAnd{Terms: []goivy.Expr{eqRed, eqGreen}}, nil)
	_, err = Generate(mod, Config{Target: "test", ClassName: "badinit", TestIters: "1"})
	if err == nil || !strings.Contains(err.Error(), "inconsistent") {
		t.Fatalf("unexpected Generate error: %v", err)
	}
}

func TestInitialConditionRejectsParameterDependency(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter initial : color
individual saved : color
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	eq, err := goivy.NewEq(goivy.NewConst("saved", color), goivy.NewConst("initial", color))
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(eq, nil)
	mod.LabeledInits = []*goivy.LabeledFormula{mod.Cfg.AstCfg.NewLabeledFormula(nil, eq)}
	_, err = Generate(mod, Config{Target: "test", ClassName: "parambad", TestIters: "1"})
	if err == nil || !strings.Contains(err.Error(), `stripped parameter "initial"`) {
		t.Fatalf("unexpected Generate error: %v", err)
	}
}

func TestDerivedDefinitionApplicationCompiles(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
definition is_green(C:color) = C = green
individual saved : bool
action check(c:color) = {
    saved := is_green(c)
}
export check
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "derived_expr", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.saved = (c == green)") {
		t.Fatalf("derived definition application should inline in action body:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestInitialConditionUsesRelevantDefinition(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
definition is_green(C:color) = C = green
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	if len(mod.Definitions) == 0 {
		t.Fatal("missing is_green definition")
	}
	def, ok := mod.Definitions[0].Formula.(*goivy.LogicDefinition)
	if !ok {
		t.Fatalf("unexpected definition formula %T", mod.Definitions[0].Formula)
	}
	relSort := def.Defines().NodeSort()
	init, err := goivy.NewApply(goivy.NewConst("is_green", relSort), goivy.NewConst("saved", color))
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(init, nil)
	mod.LabeledInits = []*goivy.LabeledFormula{mod.Cfg.AstCfg.NewLabeledFormula(nil, init)}
	out, err := Generate(mod, Config{Target: "test", ClassName: "initdef", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.saved = green") {
		t.Fatalf("relevant definition should constrain saved to green:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestOracleVariantStructReturnsMatchTrace(t *testing.T) {
	tests := []struct {
		name      string
		wantType  string
		wantTrace string
	}{
		{
			name:     "variant_simple",
			wantType: "type msg struct {",
			wantTrace: strings.Repeat("> make_req\n= {req:{shade:green}}\n", 5) +
				"test_completed\n",
		},
		{
			name:     "variant_recursive",
			wantType: "type tree struct {",
			wantTrace: strings.Repeat("> set_leaf\n= {leaf:{value:leaf_label}}\n", 5) +
				"test_completed\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := filepath.Join("..", "ivy2cpp", "test_vec", "oracle", tt.name+".ivy")
			out, err := CompileAndGenerate(fixture, map[string]string{"target": "test", "classname": tt.name}, Config{TestIters: "1"})
			if err != nil {
				t.Fatalf("CompileAndGenerate: %v\n%s", err, outSource(out))
			}
			if len(out.Warnings) != 0 {
				t.Fatalf("variant struct return should not use fallback warnings, got %v", out.Warnings)
			}
			for _, want := range []string{tt.wantType, "valid bool", "valid: true"} {
				if !strings.Contains(out.Source, want) {
					t.Fatalf("variant source missing %q:\n%s", want, out.Source)
				}
			}
			bin := compileGeneratedGo(t, out)
			stdout, stderr, err := runBinary(t, bin, "iters=5", "runs=1", "seed=1")
			if err != nil {
				t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
			}
			if stdout != tt.wantTrace {
				t.Fatalf("variant trace differs from ivy2cpp\nwant:\n%s\ngot:\n%s\nstderr:\n%s", tt.wantTrace, stdout, stderr)
			}
		})
	}
}

func TestRecursiveVariantActionParamCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type tree
type label = {leaf_label, node_label}
variant leaf of tree = struct {
    value : label
}
variant node of tree = struct {
    left : tree,
    right : tree
}
action touch(t:tree) = {
}
export touch
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "variant_param", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{"func() tree {", "left: tree{}", "right: tree{}"} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("recursive variant argument source missing %q:\n%s", want, out.Source)
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

func TestCallSiteVariantUpcastOnArgumentCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type t
variant a of t
variant b of t
individual v : t
action receive(x:t) = {
    v := x
}
action sender(y:a) = {
    call receive(y)
}
export sender
`
	mod := compileIvySource(t, src)
	mod.PublicActions.Set("receive", false)
	out, err := Generate(mod, Config{Target: "test", ClassName: "variant_call", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	want := "ivy.receive(t{tag: 0, value: y, valid: true})"
	if !strings.Contains(out.Source, want) {
		t.Fatalf("variant call should upcast subtype argument with %q:\n%s", want, out.Source)
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

func TestVariantRelationAssertionCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type t
variant a of t
variant b of t
individual v : t
individual av : a
after init {
    v := av
}
action check = {
    assert v *> av
}
export check
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "variant_rel", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{".valid && ivy.v.tag == 0", "ivy.v.value.(int) == ivy.av"} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("variant relation source missing %q:\n%s", want, out.Source)
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

func TestIfSomeVariantDowncastCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type t
variant a of t
variant b of t
individual v : t
individual av : a
after init {
    v := av
}
action load returns(out:a) = {
    if some (q:a) v *> q {
        out := q
    }
}
export load
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "variant_down", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"if ivy.v.valid && ivy.v.tag == 0 {",
		"loc__q := ivy.v.value.(int)",
		"out = loc__q",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("variant downcast source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestEmitExprSomeVariantRelationReturnsPayload(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t
variant a of t
variant b of t
individual v : t
`)
	tSort := mod.Sig.Sorts.Get("t")
	aSort := mod.Sig.Sorts.Get("a")
	q, err := goivy.NewVariable("Q", aSort)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	relSort := goivy.LogicRelationSort([]goivy.Sort{tSort, aSort})
	body, err := goivy.NewApply(goivy.NewConst("*>", relSort), goivy.NewConst("v", tSort), q)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	got, err := (&Generator{Mod: mod}).emitExpr(goivy.NewSome([]goivy.Expr{q}, body))
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{
		"func() int {",
		"if ivy.v.valid && ivy.v.tag == 0 {",
		"Q := ivy.v.value.(int)",
		"return Q",
		"return 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in variant some expression:\n%s", want, got)
		}
	}
}

func TestEmitExprSomeWithElseVariantRelation(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t
variant a of t
variant b of t
individual v : t
individual fallback : a
`)
	tSort := mod.Sig.Sorts.Get("t")
	aSort := mod.Sig.Sorts.Get("a")
	q, err := goivy.NewVariable("Q", aSort)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	relSort := goivy.LogicRelationSort([]goivy.Sort{tSort, aSort})
	body, err := goivy.NewApply(goivy.NewConst("*>", relSort), goivy.NewConst("v", tSort), q)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	some := goivy.NewSomeWithElse([]goivy.Expr{q}, body, q, goivy.NewConst("fallback", aSort))
	got, err := (&Generator{Mod: mod}).emitExpr(some)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{
		"func() int {",
		"if ivy.v.valid && ivy.v.tag == 0 {",
		"Q := ivy.v.value.(int)",
		"return Q",
		"return ivy.fallback",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in variant some/else expression:\n%s", want, got)
		}
	}
}

func TestIfSomeFiniteEnumCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
action pick returns(out:color) = {
    if some (c:color) c = green {
        out := c
    } else {
        out := red
    }
}
export pick
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "some_enum", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"__ivy_some",
		"for _, loc__c := range []color{red, green} {",
		"if !__ivy_some",
		"out = loc__c",
		"if !__ivy_some",
		"out = red",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("finite if-some source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := strings.Repeat("> pick\n= green\n", 3) + "test_completed\n"
	if stdout != wantTrace {
		t.Fatalf("finite if-some trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

func TestIfSomeInequalityBoundOverRangeCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type idx = {0..4}
action pick returns(out:idx) = {
    if some (x:idx) x >= 2 {
        out := x
    } else {
        out := 0
    }
}
export pick
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "some_bound", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"__ivy_some",
		"for loc__x := 2; loc__x < 5; loc__x++ {",
		"if !__ivy_some",
		"out = loc__x",
		"if !__ivy_some",
		"out = 0",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("bounded if-some source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "[]int{0, 1, 2, 3, 4}") {
		t.Fatalf("bounded if-some should not enumerate the full range:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := strings.Repeat("> pick\n= 2\n", 3) + "test_completed\n"
	if stdout != wantTrace {
		t.Fatalf("bounded if-some trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

const ifSomeExtensionalRelationSource = `#lang ivy1.7
type key
relation seen(K:key)
after init {
    seen(K) := false;
    seen(3) := true
}
action pick returns(out:key) = {
    if some (k:key) seen(k) {
        out := k
    } else {
        out := 0
    }
}
export pick
`

func TestIfSomeExtensionalRelationCompilesAndRuns(t *testing.T) {
	mod := compileIvySource(t, ifSomeExtensionalRelationSource)
	out, err := Generate(mod, Config{Target: "test", ClassName: "some_extensional", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"seen ivyThunkMap[int, bool]",
		"for __ivy_some_key",
		"range ivy.seen.overrides",
		"loc__k := __ivy_some_key",
		"if !__ivy_some",
		"out = loc__k",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("extensional if-some source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "unsupported some condition") {
		t.Fatalf("extensional if-some should not fall through to unsupported:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := strings.Repeat("> pick\n= 3\n", 3) + "test_completed\n"
	if stdout != wantTrace {
		t.Fatalf("extensional if-some trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

func TestIfSomeExtensionalRelationMatchesIvy2Cpp(t *testing.T) {
	const src = `#lang ivy1.7
type key
interpret key -> <<< int >>>
relation seen(K:key)
after init {
    seen(K) := false;
    seen(3) := true
}
action pick returns(out:bool) = {
    if some (k:key) seen(k) {
        out := true
    } else {
        out := false
    }
}
export pick
`
	cppMod := compileIvySource(t, src)
	cppOut, err := ivy2cppgen.Generate(cppMod, ivy2cppgen.Config{Target: "test", ClassName: "some_extensional"})
	if err != nil {
		t.Fatalf("ivy2cpp Generate: %v", err)
	}
	cppBin, err := ivy2cppgen.BuildOutput(cppOut, t.TempDir())
	if err != nil {
		t.Skipf("ivy2cpp build unavailable for parity check: %v", err)
	}
	goMod := compileIvySource(t, src)
	goOut, err := Generate(goMod, Config{Target: "test", ClassName: "some_extensional", TestIters: "1"})
	if err != nil {
		t.Fatalf("ivy2golang Generate: %v\n%s", err, outSource(goOut))
	}
	goBin := compileGeneratedGo(t, goOut)
	args := []string{"iters=3", "runs=1", "seed=1"}
	cppStdout, cppStderr, err := runBinary(t, cppBin, args...)
	if err != nil {
		t.Skipf("ivy2cpp generated binary unavailable for parity check: %v\nstdout:\n%s\nstderr:\n%s", err, cppStdout, cppStderr)
	}
	goStdout, goStderr, err := runBinary(t, goBin, args...)
	if err != nil {
		t.Fatalf("run ivy2golang binary: %v\nstdout:\n%s\nstderr:\n%s", err, goStdout, goStderr)
	}
	if goStdout != cppStdout {
		t.Fatalf("extensional if-some trace differs from ivy2cpp\nivy2cpp stdout:\n%s\nivy2golang stdout:\n%s", cppStdout, goStdout)
	}
}

func TestExtensionalExistsQuantifierCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type key
relation seen(K:key)
after init {
    seen(K) := false;
    seen(3) := true
}
action check = {
    assert exists K:key. seen(K)
}
export check
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "quant_extensional", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"seen ivyThunkMap[int, bool]",
		"for __ivy_quant_key",
		"range ivy.seen.overrides",
		"if !__ivy_quant_val",
		"if ivy.seen.Get(K)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("extensional quantifier source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "cannot enumerate quantified variable") {
		t.Fatalf("extensional quantifier should not fall through to finite enumeration:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := strings.Repeat("> check\n", 3) + "test_completed\n"
	if stdout != wantTrace {
		t.Fatalf("extensional quantifier trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

func TestExtensionalForallNegativeQuantifierCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type key
relation seen(K:key)
after init {
    seen(K) := false
}
action check = {
    assert forall K:key. ~seen(K)
}
export check
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "quant_extensional_forall", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"seen ivyThunkMap[int, bool]",
		"for __ivy_quant_key",
		"range ivy.seen.overrides",
		"if !__ivy_quant_val",
		"!(ivy.seen.Get(K))",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("extensional forall source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "cannot enumerate quantified variable") {
		t.Fatalf("extensional forall should not fall through to finite enumeration:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := strings.Repeat("> check\n", 3) + "test_completed\n"
	if stdout != wantTrace {
		t.Fatalf("extensional forall trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

func TestQuantifierInequalityBoundOverRange(t *testing.T) {
	src := `#lang ivy1.7
type idx = {0..4}
relation marked(X:idx)
after init {
    marked(X) := false;
    marked(0) := true;
    marked(1) := true
}
action check = {
    assert forall X:idx. X < 2 -> marked(X)
}
export check
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "quant_bound", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	checkIdx := strings.Index(out.Source, "func (ivy *quant_bound) check()")
	if checkIdx < 0 {
		t.Fatalf("missing check method:\n%s", out.Source)
	}
	checkSrc := out.Source[checkIdx:]
	if !strings.Contains(checkSrc, "for X := 0; X < 2; X++ {") {
		t.Fatalf("quantifier should use inequality-derived upper bound in check():\n%s", checkSrc)
	}
	if strings.Contains(checkSrc, "[]int{0, 1, 2, 3, 4}") {
		t.Fatalf("quantifier check should not enumerate the full range after deriving X < 2:\n%s", checkSrc)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := strings.Repeat("> check\n", 3) + "test_completed\n"
	if stdout != wantTrace {
		t.Fatalf("bounded quantifier trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

func TestIfSomeMinFiniteRangeCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type idx = {0..4}
action pick returns(out:idx) = {
    if some x:idx. x >= 2 minimizing x {
        out := x
    } else {
        out := 0
    }
}
export pick
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "some_min_range", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"var __ivy_some_idx",
		"var __ivy_some_w0_loc__x",
		"__ivy_some_cur",
		"if !__ivy_some",
		"break",
		"loc__x := __ivy_some_w0_loc__x",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("some_min source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := strings.Repeat("> pick\n= 2\n", 3) + "test_completed\n"
	if stdout != wantTrace {
		t.Fatalf("some_min trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

func TestIfSomeMaxFiniteRangeCompilesAndRuns(t *testing.T) {
	src := `#lang ivy1.7
type idx = {0..4}
action pick returns(out:idx) = {
    if some x:idx, y:idx. x = 1 maximizing y {
        out := y
    } else {
        out := 0
    }
}
export pick
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "some_max_range", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"var __ivy_some_idx",
		"var __ivy_some_w1_loc__y",
		"< __ivy_some_cur",
		"loc__y := __ivy_some_w1_loc__y",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("some_max source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := strings.Repeat("> pick\n= 4\n", 3) + "test_completed\n"
	if stdout != wantTrace {
		t.Fatalf("some_max trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
	}
}

func TestEmitExprSomeFiniteEnum(t *testing.T) {
	color := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	x, err := goivy.NewVariable("X", color)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	green := goivy.NewConst("green", color)
	body, err := goivy.NewEq(x, green)
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	got, err := (&Generator{}).emitExpr(goivy.NewSome([]goivy.Expr{x}, body))
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{
		"func() color {",
		"for _, X := range []color{red, green} {",
		"if (X == green) {",
		"return X",
		"return red",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("finite some expression missing %q:\n%s", want, got)
		}
	}
}

func TestEmitExprSomeInequalityBoundOverRange(t *testing.T) {
	idx := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "4"}}
	x, err := goivy.NewVariable("X", idx)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	geSort, err := goivy.NewFunctionSort(idx, idx, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	body, err := goivy.NewApply(goivy.NewConst(">=", geSort), x, goivy.NewConst("2", idx))
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	got, err := (&Generator{}).emitExpr(goivy.NewSome([]goivy.Expr{x}, body))
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{
		"func() int {",
		"for X := 2; X < (4)+1; X++ {",
		"if (X >= 2) {",
		"return X",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("bounded some expression missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "[]int{0, 1, 2, 3, 4}") {
		t.Fatalf("bounded some expression should not enumerate full range:\n%s", got)
	}
}

func TestEmitExprSomeWithElseFiniteEnum(t *testing.T) {
	color := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	x, err := goivy.NewVariable("X", color)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	green := goivy.NewConst("green", color)
	red := goivy.NewConst("red", color)
	body, err := goivy.NewEq(x, green)
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	got, err := (&Generator{}).emitExpr(goivy.NewSomeWithElse([]goivy.Expr{x}, body, x, red))
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{
		"func() color {",
		"for _, X := range []color{red, green} {",
		"if (X == green) {",
		"return X",
		"return red",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("finite some/else expression missing %q:\n%s", want, got)
		}
	}
}

func TestOracleRemainingCLIFixturesCompileAndRun(t *testing.T) {
	for _, name := range []string{"empty", "isolate_two_parts"} {
		t.Run(name, func(t *testing.T) {
			fixture := filepath.Join("..", "ivy2cpp", "test_vec", "oracle", name+".ivy")
			out, err := CompileAndGenerate(fixture, map[string]string{"target": "test", "classname": name}, Config{TestIters: "1"})
			if err != nil {
				t.Fatalf("CompileAndGenerate: %v\n%s", err, outSource(out))
			}
			bin := compileGeneratedGo(t, out)
			stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
			if err != nil {
				t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
			}
			if !strings.Contains(stdout, "test_completed") {
				t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
			}
		})
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
		"for _, C := range []color{red, green}",
		"__ivy_tmp0[C] = true",
		"ivy.marked[C] = __ivy_tmp0[C]",
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

func TestMergeParamsAcceptsIvy2CppDriverSurfaceAndDefaultsToGen(t *testing.T) {
	cfg, ivyParams, err := mergeParams(map[string]string{
		"classname":  "MyIvy",
		"main":       "ivy_main",
		"outdir":     "out",
		"trace":      "true",
		"stdafx":     "yes",
		"build":      "1",
		"isolate":    "iso",
		"test_iters": "7",
		"test_runs":  "3",
		"compiler":   "g++",
	}, Config{})
	if err != nil {
		t.Fatalf("mergeParams: %v", err)
	}
	if cfg.RequestedTarget != "gen" || cfg.Target != "gen" || !cfg.EmitMain {
		t.Fatalf("default target normalization = requested %q effective %q emitMain %v, want gen/gen/true", cfg.RequestedTarget, cfg.Target, cfg.EmitMain)
	}
	if cfg.ClassName != "MyIvy" || cfg.MainName != "ivy_main" || cfg.OutDir != "out" {
		t.Fatalf("string params not merged: %+v", cfg)
	}
	if cfg.TestIters != "7" || cfg.TestRuns != "3" || cfg.Compiler != "g++" {
		t.Fatalf("driver params not merged: %+v", cfg)
	}
	if !cfg.Trace || !cfg.Stdafx || !cfg.Build {
		t.Fatalf("bool params not merged: %+v", cfg)
	}
	if ivyParams["isolate"] != "iso" {
		t.Fatalf("isolate param = %q, want iso", ivyParams["isolate"])
	}
}

func TestMergeParamsRejectsUnknownTargetCompilerAndParameter(t *testing.T) {
	for _, params := range []map[string]string{
		{"target": "bad"},
		{"compiler": "clang++"},
		{"unknown": "value"},
	} {
		if _, _, err := mergeParams(params, Config{}); err == nil {
			t.Fatalf("mergeParams(%v) succeeded, want error", params)
		}
	}
}

func TestMainNameAndTestDefaultsAreEmitted(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "WithMain", MainName: "ivy_main", TestIters: "7", TestRuns: "3"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func ivy_main()",
		`testIters := ivyAtoiDefault(opts["iters"], 7)`,
		`runs := ivyAtoiDefault(opts["runs"], 3)`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("missing %q:\n%s", want, out.Source)
		}
	}
	compileGeneratedGo(t, out)
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

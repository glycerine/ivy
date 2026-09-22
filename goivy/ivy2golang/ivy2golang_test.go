package ivy2golang

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
	ivy2cppgen "github.com/glycerine/ivy/goivy/ivy2cpp"
)

type testRadixRelname struct {
	rep string
}

func (r testRadixRelname) Relname() string { return r.rep }

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

func assertNoUnsupportedGo(t *testing.T, src string) {
	t.Helper()
	for _, bad := range []string{
		"ivy2golang: unsupported",
		"cannot enumerate quantified variable",
		`panic("ivy2golang`,
	} {
		if strings.Contains(src, bad) {
			t.Fatalf("generated Go should not contain %q:\n%s", bad, src)
		}
	}
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

func TestGenerateDeterministicAcrossFreshCompiles(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green, blue}
type idx = {0..2}
relation edge(C:color,I:idx)
individual saved : color
individual count : idx
after init {
    saved := red;
    count := 0;
    edge(C,I) := false
}
action beta(i:idx) = {
    edge(saved,i) := true;
    count := i
}
action alpha(c:color) = {
    assume edge(c,count) -> count = count;
    saved := c
}
action gamma returns(out:color) = {
    out := saved
}
export gamma
export beta
export alpha
`
	var first string
	for i := 0; i < 10; i++ {
		mod := compileIvySource(t, src)
		out, err := Generate(mod, Config{Target: "test", ClassName: "deterministic", TestIters: "2", TestRuns: "1"})
		if err != nil {
			t.Fatalf("Generate pass %d: %v\n%s", i, err, outSource(out))
		}
		got := out.Source + "\n-- warnings --\n" + strings.Join(out.Warnings, "\n")
		if i == 0 {
			first = got
			continue
		}
		if got != first {
			t.Fatalf("Generate output changed on pass %d\nfirst:\n%s\n\ngot:\n%s", i, first, got)
		}
	}
}

func TestGenerateNormalizesExplicitClassNameToGoIdentifier(t *testing.T) {
	mod := goivy.New()
	mod.Name = "keyword"
	out, err := Generate(mod, Config{Target: "test", ClassName: "type.runner[0]", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out.ClassName != "type__runner__0__" || out.Config.ClassName != out.ClassName {
		t.Fatalf("class name was not normalized consistently: out=%q cfg=%q", out.ClassName, out.Config.ClassName)
	}
	for _, want := range []string{
		"type type__runner__0__ struct",
		"func newType__runner__0__(",
		"func (ivy *type__runner__0__) __initState()",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("normalized class name source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "type.runner") || strings.Contains(out.Source, "runner[0]") {
		t.Fatalf("raw class name leaked into generated source:\n%s", out.Source)
	}
}

func TestGenerateNormalizesQuotedExplicitClassNameAsIdentifier(t *testing.T) {
	mod := goivy.New()
	mod.Name = "quoted"
	out, err := Generate(mod, Config{Target: "test", ClassName: `"quoted.runner"`, TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out.ClassName != "_quoted__runner_" || out.Config.ClassName != out.ClassName {
		t.Fatalf("class name was not normalized as an identifier: out=%q cfg=%q", out.ClassName, out.Config.ClassName)
	}
	for _, bad := range []string{
		`type "quoted.runner" struct`,
		`func new"quoted.runner"`,
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("quoted class name leaked into generated source as %q:\n%s", bad, out.Source)
		}
	}
	if !strings.Contains(out.Source, "type _quoted__runner_ struct") {
		t.Fatalf("quoted class name source missing normalized type:\n%s", out.Source)
	}
}

func TestCheckGeneratedNamesRejectsActionCollision(t *testing.T) {
	mod := goivy.New()
	mod.Name = "modname"
	mod.Actions.Set("foo", goivy.NewSequence())
	_, err := Generate(mod, Config{ClassName: "foo"})
	if err == nil || !strings.Contains(err.Error(), "cannot create Go type foo with generated name foo") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckGeneratedNamesRejectsSymbolCollision(t *testing.T) {
	mod := goivy.New()
	mod.Name = "modname"
	if _, err := mod.Sig.AddSymbol("bar", goivy.Boolean); err != nil {
		t.Fatalf("AddSymbol: %v", err)
	}
	_, err := Generate(mod, Config{ClassName: "bar"})
	if err == nil || !strings.Contains(err.Error(), "cannot create Go type bar with generated name bar") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckGeneratedNamesRejectsSortCollision(t *testing.T) {
	mod := goivy.New()
	mod.Name = "modname"
	if err := mod.Sig.AddSort(&goivy.UninterpretedSort{Name: "baz"}); err != nil {
		t.Fatalf("AddSort: %v", err)
	}
	_, err := Generate(mod, Config{ClassName: "baz"})
	if err == nil || !strings.Contains(err.Error(), "cannot create Go type baz with generated name baz") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckGeneratedNamesRejectsEnumElementCollision(t *testing.T) {
	mod := goivy.New()
	mod.Name = "modname"
	if err := mod.Sig.AddSort(&goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}); err != nil {
		t.Fatalf("AddSort: %v", err)
	}
	_, err := Generate(mod, Config{ClassName: "red"})
	if err == nil || !strings.Contains(err.Error(), "cannot create Go type red with generated name red") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckGeneratedNamesAllowsDistinctClassname(t *testing.T) {
	mod := goivy.New()
	mod.Name = "modname"
	mod.Actions.Set("foo", goivy.NewSequence())
	_, err := Generate(mod, Config{ClassName: "qux"})
	if err != nil && strings.Contains(err.Error(), "cannot create Go type") {
		t.Fatalf("unexpected generated-name collision error: %v", err)
	}
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

func TestConjectureDoesNotAppendAssertToPrivateActions(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
relation r
after init { r := true }
action helper = { r := false }
action step = {
    call helper;
    r := true
}
export step
conjecture r
`)
	mod.PublicActions.Set("helper", false)
	out, err := Generate(mod, Config{Target: "test", ClassName: "conj_private", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	stepAct, ok := mod.Actions.Get2("step")
	if !ok {
		t.Fatalf("step action missing after prepareModuleForGo")
	}
	if findFirstAssertInAction(stepAct) == nil {
		t.Fatalf("expected conjecture assert appended to exported step action")
	}
	helperAct, ok := mod.Actions.Get2("helper")
	if !ok {
		t.Fatalf("helper action missing after prepareModuleForGo")
	}
	if findFirstAssertInAction(helperAct) != nil {
		t.Fatalf("private helper action should not receive conjecture assert:\n%v", helperAct)
	}
	helperBody := bodyAfterMarker(out.Source, "func (ivy *conj_private) helper(")
	if helperBody == "" {
		t.Fatalf("helper body not emitted:\n%s", out.Source)
	}
	if strings.Contains(helperBody, "ivyAssert(") {
		t.Fatalf("generated private helper body should not contain conjecture assert:\n%s", helperBody)
	}
	stepBody := bodyAfterMarker(out.Source, "func (ivy *conj_private) step(")
	if !strings.Contains(stepBody, "ivyAssert(") {
		t.Fatalf("generated exported step body should contain conjecture assert:\n%s", stepBody)
	}
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
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if strings.Contains(mainBody, "if !(false)") && strings.Contains(mainBody, "cycle--") {
		t.Fatalf("returning action assume should execute and fail, not retry forever:\n%s", mainBody)
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

func TestAssertLabelStripsTrailingColonSpace(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action check = {
    assert saved = red
}
export check
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "labelfmt", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *labelfmt) check(")
	if body == "" {
		t.Fatalf("check method body not found:\n%s", out.Source)
	}
	idx := strings.Index(body, "ivyAssert(")
	if idx < 0 {
		t.Fatalf("no ivyAssert in check body:\n%s", body)
	}
	lineEnd := strings.Index(body[idx:], "\n")
	if lineEnd < 0 {
		lineEnd = len(body) - idx
	}
	line := body[idx : idx+lineEnd]
	if !strings.Contains(line, `"test.ivy: line 5"`) {
		t.Fatalf("ivyAssert label missing source line:\n%s", line)
	}
	if strings.Contains(line, `: ")`) {
		t.Fatalf("ivyAssert label ends in trailing colon-space:\n%s", line)
	}
}

func TestLinenoStrUsesBasenameAndDenormalizesExamples(t *testing.T) {
	got := linenoStr(goivy.Location{Filename: "/tmp/specs/hermes_rmw_o3_testing.ivy", Line: 185})
	if got != "hermes_rmw_o3_testing.ivy: line 185" {
		t.Fatalf("linenoStr basename=%q", got)
	}
	t.Setenv("HOME", "/home/tester")
	got = linenoStr(goivy.Location{Filename: "<IVY_EXAMPLES>", Line: 7})
	if got != "ivy-lang-examples: line 7" {
		t.Fatalf("linenoStr examples sentinel=%q", got)
	}
}

func TestTargetTestLeadingAssumeGuardSkipsRejectedInputs(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    assume ~(c = red);
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "testguard", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Testguard_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`gen.c = color(ivy.___ivy_randomize(2, "set.fml:c", 0))`,
		`if !(!((gen.c == red))) {`,
		`return false`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("target=test assume guard generator missing %q:\n%s", want, genBody)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`set_generator := &Testguard_set_generator{ivy: ivy}`,
		`if !set_generator.generate() {`,
		`cycle--`,
		`continue`,
		`__arg0 := set_generator.c`,
		`fmt.Fprintf(__ivy_out, "> set(%s)\n", ivyTraceValue(__arg0, false))`,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("target=test assume guard source missing %q:\n%s", want, mainBody)
		}
	}
	guardIdx := strings.Index(mainBody, `if !set_generator.generate() {`)
	traceIdx := strings.Index(mainBody, `fmt.Fprintf(__ivy_out, "> set(%s)\n", ivyTraceValue(__arg0, false))`)
	if guardIdx < 0 || traceIdx < 0 || guardIdx > traceIdx {
		t.Fatalf("assume guard should run before action trace; guard=%d trace=%d\n%s", guardIdx, traceIdx, mainBody)
	}

	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=20", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("guarded target=test run should skip failed assumes\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", stdout, stderr, out.Source)
	}
	if strings.Contains(stdout, "assumption_failed") || strings.Contains(stderr, "assumption failed") {
		t.Fatalf("leading assume guard should not execute failed assumes\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if strings.Contains(stdout, "> set(red)") {
		t.Fatalf("leading assume guard should not trace rejected red inputs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stdout, "> set(green)") || !strings.Contains(stdout, "test_completed") {
		t.Fatalf("guarded run missing accepted call or completion marker\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestTargetTestNonLeadingAssumeGuardSkipsRejectedInputsFast(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
after init {
    saved := red
}
action set(c:color) = {
    assert c = c;
    assume c = green;
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "testnonguard", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Testnonguard_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not emitted:\n%s", out.Source)
	}
	guard := `if !((gen.c == green)) {`
	for _, want := range []string{
		guard,
		`return false`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("target=test non-leading assume generator missing %q:\n%s", want, genBody)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	trace := `fmt.Fprintf(__ivy_out, "> set(%s)\n", ivyTraceValue(__arg0, false))`
	for _, want := range []string{
		`if !set_generator.generate() {`,
		`cycle--`,
		`continue`,
		trace,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("target=test non-leading assume guard source missing %q:\n%s", want, mainBody)
		}
	}
	guardIdx := strings.Index(mainBody, `if !set_generator.generate() {`)
	traceIdx := strings.Index(mainBody, trace)
	if guardIdx < 0 || traceIdx < 0 || guardIdx > traceIdx {
		t.Fatalf("non-leading assume guard should run before action trace; guard=%d trace=%d\n%s", guardIdx, traceIdx, mainBody)
	}
}

func TestTargetTestAssignedStateAssumeUsesPreimageFast(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
after init {
    saved := red
}
action set(c:color) = {
    saved := c;
    assume saved = green
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "testpreimage", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Testpreimage_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not emitted:\n%s", out.Source)
	}
	guard := `if !((gen.c == green)) {`
	for _, want := range []string{
		guard,
		`return false`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("target=test assigned-state assume generator missing %q:\n%s", want, genBody)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	trace := `fmt.Fprintf(__ivy_out, "> set(%s)\n", ivyTraceValue(__arg0, false))`
	for _, want := range []string{
		`if !set_generator.generate() {`,
		`cycle--`,
		`continue`,
		trace,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("target=test assigned-state assume preimage source missing %q:\n%s", want, mainBody)
		}
	}
	guardIdx := strings.Index(mainBody, `if !set_generator.generate() {`)
	traceIdx := strings.Index(mainBody, trace)
	if guardIdx < 0 || traceIdx < 0 || guardIdx > traceIdx {
		t.Fatalf("assigned-state assume guard should run before action trace; guard=%d trace=%d\n%s", guardIdx, traceIdx, mainBody)
	}
}

func TestTargetTestUsesPerActionGeneratorFast(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    assume c = green;
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "testactiongen", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Testactiongen_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("target=test action generator body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`gen.c = color(ivy.___ivy_randomize(2, "set.fml:c", 0))`,
		`if !((gen.c == green)) {`,
		`return false`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("target=test action generator missing %q:\n%s", want, genBody)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`set_generator := &Testactiongen_set_generator{ivy: ivy}`,
		`if !set_generator.generate() {`,
		`__arg0 := set_generator.c`,
		`fmt.Fprintf(__ivy_out, "> set(%s)\n", ivyTraceValue(__arg0, false))`,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("target=test main missing generator use %q:\n%s", want, mainBody)
		}
	}
	if strings.Contains(mainBody, `color(ivy.___ivy_randomize(2, "set.fml:c", 0))`) {
		t.Fatalf("target=test main should use the per-action generator field, not direct randomization:\n%s", mainBody)
	}
}

func TestTargetTestConditionalAssumeUsesImplicationGuardFast(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
after init {
    saved := red
}
action set(c:color) = {
    if c = red {
        assume false
    };
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "testifguard", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Testifguard_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not emitted:\n%s", out.Source)
	}
	guardExpr := `(!((gen.c == red)) || (false))`
	for _, want := range []string{
		guardExpr,
		`return false`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("target=test conditional assume generator missing %q:\n%s", want, genBody)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	trace := `fmt.Fprintf(__ivy_out, "> set(%s)\n", ivyTraceValue(__arg0, false))`
	for _, want := range []string{
		`if !set_generator.generate() {`,
		`cycle--`,
		`continue`,
		trace,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("target=test conditional assume guard source missing %q:\n%s", want, mainBody)
		}
	}
	guardIdx := strings.Index(mainBody, `if !set_generator.generate() {`)
	traceIdx := strings.Index(mainBody, trace)
	if guardIdx < 0 || traceIdx < 0 || guardIdx > traceIdx {
		t.Fatalf("conditional assume guard should run before action trace; guard=%d trace=%d\n%s", guardIdx, traceIdx, mainBody)
	}
}

func TestTargetTestLocalAssumeUsesTrialGeneratorFast(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
after init {
    saved := red
}
action step = {
    var choice : color;
    assume choice = green;
    saved := choice
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "testlocalguard", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		"__ivy_trial := ivy.__ivy_clone()",
		"__ivy_assume_rejecting = true",
		"if __ivy_trial_rejected {",
		"cycle--",
		"ivy = __ivy_trial",
		`fmt.Fprintln(__ivy_out, "> step")`,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("target=test local assume trial source missing %q:\n%s", want, mainBody)
		}
	}
	trialIdx := strings.Index(mainBody, "__ivy_trial := ivy.__ivy_clone()")
	traceIdx := strings.Index(mainBody, `fmt.Fprintln(__ivy_out, "> step")`)
	if trialIdx < 0 || traceIdx < 0 || trialIdx > traceIdx {
		t.Fatalf("local-assume trial should execute before public trace; trial=%d trace=%d\n%s", trialIdx, traceIdx, mainBody)
	}
}

func TestTargetTestLeadingAssumeGuardOnZeroArgAction(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
    assume false;
    assert false
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "zeroguard", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Zeroguard_step_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("step generator body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`if !(false) {`,
		`return false`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("target=test zero-arg assume generator missing %q:\n%s", want, genBody)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`if !step_generator.generate() {`,
		`cycle--`,
		`continue`,
		`fmt.Fprintln(__ivy_out, "> step")`,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("target=test zero-arg assume guard source missing %q:\n%s", want, mainBody)
		}
	}
	guardIdx := strings.Index(mainBody, `if !step_generator.generate() {`)
	traceIdx := strings.Index(mainBody, `fmt.Fprintln(__ivy_out, "> step")`)
	if guardIdx < 0 || traceIdx < 0 || guardIdx > traceIdx {
		t.Fatalf("zero-arg assume guard should run before action trace; guard=%d trace=%d\n%s", guardIdx, traceIdx, mainBody)
	}
}

func TestTargetTestExtPreconditionGuardOnZeroArgAction(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
    assert false
}
export step
`)
	mod.ExtPreconds["step"] = goivy.False
	out, err := Generate(mod, Config{Target: "test", ClassName: "zeroextguard", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Zeroextguard_step_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("step generator body not emitted:\n%s", out.Source)
	}
	if !strings.Contains(genBody, `if !(false) {`) {
		t.Fatalf("target=test zero-arg ext-precondition generator guard missing:\n%s", genBody)
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	if !strings.Contains(mainBody, `if !step_generator.generate() {`) {
		t.Fatalf("target=test zero-arg ext-precondition guard missing:\n%s", mainBody)
	}
	guardIdx := strings.Index(mainBody, `if !step_generator.generate() {`)
	traceIdx := strings.Index(mainBody, `fmt.Fprintln(__ivy_out, "> step")`)
	if guardIdx < 0 || traceIdx < 0 || guardIdx > traceIdx {
		t.Fatalf("zero-arg ext-precondition guard should run before action trace; guard=%d trace=%d\n%s", guardIdx, traceIdx, mainBody)
	}
}

func TestTargetTestUnsupportedAssumeGuardErrorIncludesSourceLine(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	native := &goivy.LogicNativeExpr{
		CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
	}
	native.SetLineno(goivy.Location{Filename: "guards.ivy", Line: 41})
	mod.ExtPreconds["step"] = native

	_, err := Generate(mod, Config{Target: "test", ClassName: "badtestguard", TestIters: "1"})
	if err == nil {
		t.Fatal("Generate should reject unsupported target=test assume guard")
	}
	for _, want := range []string{
		"guards.ivy: line 41:",
		"unsupported target=test action generator assume guard",
		"native C++ expression is not translated to Go",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("target=test assume guard error missing %q:\n%v", want, err)
		}
	}
}

func TestTargetTestUsesBeforeExportAssumeGuard(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	action, ok := mod.Actions.Get2("set")
	if !ok {
		t.Fatalf("action set not found")
	}
	params := action.GetFormalParams()
	if len(params) != 1 {
		t.Fatalf("set params = %d, want 1", len(params))
	}
	greenEntry, ok := mod.Sig.Symbols.Get2("green")
	if !ok {
		t.Fatalf("symbol green not found")
	}
	paramGreen, err := goivy.NewEq(params[0], goivy.NewConst("green", greenEntry.Sort))
	if err != nil {
		t.Fatalf("param/green NewEq: %v", err)
	}
	actionExpr, ok := action.(goivy.Expr)
	if !ok {
		t.Fatalf("action set does not implement Expr: %T", action)
	}
	before := goivy.NewSequence(goivy.NewAssumeAction(paramGreen), actionExpr)
	before.SetLineno(action.GetLineno())
	goivy.CopyFormalsTo(action, before)
	mod.BeforeExport.Set("set", before)

	out, err := Generate(mod, Config{Target: "test", ClassName: "beforetest", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Beforetest_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`gen.c = color(ivy.___ivy_randomize(2, "set.fml:c", 0))`,
		`if !((gen.c == green)) {`,
		`return false`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("target=test before_export generator missing %q:\n%s", want, genBody)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`if !set_generator.generate() {`,
		`cycle--`,
		`continue`,
		`__arg0 := set_generator.c`,
		`__ivy_trial.set(__arg0)`,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("target=test before_export guard source missing %q:\n%s", want, mainBody)
		}
	}
	guardIdx := strings.Index(mainBody, `if !set_generator.generate() {`)
	callIdx := strings.Index(mainBody, `__ivy_trial.set(__arg0)`)
	traceIdx := strings.Index(mainBody, `fmt.Fprintf(__ivy_out, "> set(%s)\n", ivyTraceValue(__arg0, false))`)
	if guardIdx < 0 || traceIdx < 0 || guardIdx > traceIdx {
		t.Fatalf("before_export guard should run before action trace; guard=%d trace=%d\n%s", guardIdx, traceIdx, mainBody)
	}
	if callIdx < 0 || guardIdx > callIdx || callIdx > traceIdx {
		t.Fatalf("before_export guard should run before trial action and trace after commit; guard=%d call=%d trace=%d\n%s", guardIdx, callIdx, traceIdx, mainBody)
	}
}

func TestTargetTestBeforeExportNonLeadingAssumeUsesPreimageFast(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	action, ok := mod.Actions.Get2("set")
	if !ok {
		t.Fatalf("action set not found")
	}
	params := action.GetFormalParams()
	if len(params) != 1 {
		t.Fatalf("set params = %d, want 1", len(params))
	}
	savedEntry, ok := mod.Sig.Symbols.Get2("saved")
	if !ok {
		t.Fatalf("symbol saved not found")
	}
	greenEntry, ok := mod.Sig.Symbols.Get2("green")
	if !ok {
		t.Fatalf("symbol green not found")
	}
	saved := goivy.NewConst("saved", savedEntry.Sort)
	assignSaved := goivy.NewAssignAction(saved, params[0])
	savedGreen, err := goivy.NewEq(saved, goivy.NewConst("green", greenEntry.Sort))
	if err != nil {
		t.Fatalf("saved/green NewEq: %v", err)
	}
	actionExpr, ok := action.(goivy.Expr)
	if !ok {
		t.Fatalf("action set does not implement Expr: %T", action)
	}
	before := goivy.NewSequence(assignSaved, goivy.NewAssumeAction(savedGreen), actionExpr)
	before.SetLineno(action.GetLineno())
	goivy.CopyFormalsTo(action, before)
	mod.BeforeExport.Set("set", before)

	out, err := Generate(mod, Config{Target: "test", ClassName: "beforetestpreimage", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Beforetestpreimage_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`gen.c = color(ivy.___ivy_randomize(2, "set.fml:c", 0))`,
		`if !((gen.c == green)) {`,
		`return false`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("before_export preimage generator missing %q:\n%s", want, genBody)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`if !set_generator.generate() {`,
		`__arg0 := set_generator.c`,
		`__ivy_trial.set(__arg0)`,
		`fmt.Fprintf(__ivy_out, "> set(%s)\n", ivyTraceValue(__arg0, false))`,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("before_export preimage source missing %q:\n%s", want, mainBody)
		}
	}
	guardIdx := strings.Index(mainBody, `if !set_generator.generate() {`)
	callIdx := strings.Index(mainBody, `__ivy_trial.set(__arg0)`)
	traceIdx := strings.Index(mainBody, `fmt.Fprintf(__ivy_out, "> set(%s)\n", ivyTraceValue(__arg0, false))`)
	if guardIdx < 0 || callIdx < 0 || traceIdx < 0 || guardIdx > callIdx || callIdx > traceIdx {
		t.Fatalf("before_export preimage guard should precede trial call and public trace; guard=%d call=%d trace=%d\n%s", guardIdx, callIdx, traceIdx, mainBody)
	}
}

func TestTargetGenUnsupportedAssumeGuardErrorIncludesSourceLine(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	native := &goivy.LogicNativeExpr{
		CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
	}
	native.SetLineno(goivy.Location{Filename: "guards.ivy", Line: 43})
	mod.ExtPreconds["step"] = native

	_, err := Generate(mod, Config{Target: "gen", ClassName: "badgenguard", TestIters: "1"})
	if err == nil {
		t.Fatal("Generate should reject unsupported target=gen assume guard")
	}
	for _, want := range []string{
		"guards.ivy: line 43:",
		"unsupported action generator assume guard",
		"native C++ expression is not translated to Go",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("target=gen assume guard error missing %q:\n%v", want, err)
		}
	}
}

func TestDefinedInputUnsupportedErrorIncludesSourceLine(t *testing.T) {
	makeModule := func(t *testing.T) *goivy.Module {
		t.Helper()
		mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..7}
individual stored : idx
action set(x:idx) = {
    stored := x
}
export set
`)
		action, ok := mod.Actions.Get2("set")
		if !ok {
			t.Fatal("action set not found")
		}
		params := action.GetFormalParams()
		if len(params) != 1 {
			t.Fatalf("set params = %d, want 1", len(params))
		}
		native := &goivy.LogicNativeExpr{
			CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
		}
		loc := goivy.Location{Filename: "defined_inputs.ivy", Line: 47}
		native.SetLineno(loc)
		eq, err := goivy.NewEq(params[0], native)
		if err != nil {
			t.Fatalf("NewEq: %v", err)
		}
		eq.SetLineno(loc)
		if mod.ExtPreconds == nil {
			mod.ExtPreconds = map[string]goivy.Expr{}
		}
		mod.ExtPreconds["set"] = eq
		return mod
	}

	for _, tc := range []struct {
		name   string
		target string
		want   string
	}{
		{name: "target-test", target: "test", want: "unsupported target=test defined input"},
		{name: "target-gen", target: "gen", want: "unsupported action generator defined input"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Generate(makeModule(t), Config{Target: tc.target, ClassName: "baddefinput", TestIters: "1", TestRuns: "1"})
			if err == nil {
				t.Fatalf("Generate target=%s should reject unsupported defined input", tc.target)
			}
			for _, want := range []string{
				"defined_inputs.ivy: line 47: " + tc.want,
				"native C++ expression is not translated to Go",
			} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("target=%s defined input error missing %q:\n%v", tc.target, want, err)
				}
			}
		})
	}
}

func TestTargetTestLeadingAssumeDefinedInput(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..7}
individual stored : idx
action set(x:idx) = {
    assume x = 3;
    stored := x
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "testdefparam", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Testdefparam_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`gen.x = (0 + ivy.___ivy_randomize(8, "set.fml:x", 0))`,
		`gen.x = 3`,
		`if !((gen.x == 3)) {`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("target=test defined input generator missing %q:\n%s", want, genBody)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	if !strings.Contains(mainBody, `fmt.Fprintf(__ivy_out, "> set(%s)\n", ivyTraceValue(__arg0, false))`) {
		t.Fatalf("target=test defined input trace missing:\n%s", mainBody)
	}
	assignIdx := strings.Index(genBody, `gen.x = 3`)
	guardIdx := strings.Index(genBody, `if !((gen.x == 3)) {`)
	if assignIdx < 0 || guardIdx < 0 || assignIdx > guardIdx {
		t.Fatalf("defined input should be assigned before generator guard; assign=%d guard=%d\n%s", assignIdx, guardIdx, genBody)
	}
	traceIdx := strings.Index(mainBody, `fmt.Fprintf(__ivy_out, "> set(%s)\n", ivyTraceValue(__arg0, false))`)
	if traceIdx < 0 {
		t.Fatalf("defined input trace missing; trace=%d\n%s", traceIdx, mainBody)
	}

	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("defined-input target=test run failed\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", stdout, stderr, out.Source)
	}
	wantTrace := strings.Repeat("> set(3)\n", 3) + "test_completed\n"
	if stdout != wantTrace {
		t.Fatalf("defined input trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
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
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main not emitted:\n%s", out.Source)
	}
	lockIdx := strings.Index(mainBody, "ivy.__lock()")
	finalizeIdx := strings.Index(mainBody, "ivy.ext___finalize()")
	unlockIdx := strings.Index(mainBody, "ivy.__unlock()")
	if lockIdx < 0 || finalizeIdx < 0 || unlockIdx < 0 || lockIdx > finalizeIdx || finalizeIdx > unlockIdx {
		t.Fatalf("ext:_finalize should run under generated lock/unlock; lock=%d finalize=%d unlock=%d\n%s", lockIdx, finalizeIdx, unlockIdx, mainBody)
	}
	if strings.Contains(out.Source, `fmt.Fprintln(__ivy_out, "> _finalize")`) {
		t.Fatalf("ext:_finalize should not be in the randomized action loop:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestTestMainIgnoresPrivateExtFinalize(t *testing.T) {
	mod := goivy.New()
	mod.Name = "finprivate"
	mod.Actions.Set("ext:_finalize", goivy.NewSequence())
	mod.PublicActions.Set("ext:_finalize", false)
	mod.Actions.Set("step", goivy.NewSequence())
	mod.PublicActions.Set("step", true)
	out, err := Generate(mod, Config{Target: "test", ClassName: "finprivate", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main not emitted:\n%s", out.Source)
	}
	if strings.Contains(mainBody, "ivy.ext___finalize()") ||
		strings.Contains(mainBody, `fmt.Fprintln(__ivy_out, "> _finalize")`) {
		t.Fatalf("private ext:_finalize should not be treated as exported:\n%s", mainBody)
	}
}

func TestTestMainUsesIvy2CppActionWeights(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action heavy = {}
export heavy
attribute heavy.weight = "3.0"
action light = {}
export light
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "wt", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`__choices := 4.0 + 5.0`,
		`__choice := float64(ivyRand31()) * __choices / 2147483648.0`,
		`if __choice >= 4.0 {`,
		`if __choice < 3.0 {`,
		`fmt.Fprintln(__ivy_out, "> heavy")`,
		`} else if __choice < 4.0 {`,
		`fmt.Fprintln(__ivy_out, "> light")`,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("weighted test main missing %q:\n%s", want, mainBody)
		}
	}
	if strings.Contains(mainBody, `__choices := 2.0 + 5.0`) ||
		strings.Contains(mainBody, `if __choice < 1.0 {`) {
		t.Fatalf("weighted test main ignored action.weight attributes:\n%s", mainBody)
	}
}

func TestTargetTestRandomizedActionsUseSortedIvy2CppOrder(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action beta = {}
export beta
action alpha = {}
export alpha
action gamma = {}
export gamma
action internal = {}
`)
	mod.PublicActions.Set("internal", false)
	out, err := Generate(mod, Config{Target: "test", ClassName: "testorder", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main not emitted:\n%s", out.Source)
	}
	alpha := strings.Index(mainBody, `fmt.Fprintln(__ivy_out, "> alpha")`)
	beta := strings.Index(mainBody, `fmt.Fprintln(__ivy_out, "> beta")`)
	gamma := strings.Index(mainBody, `fmt.Fprintln(__ivy_out, "> gamma")`)
	if alpha < 0 || beta < 0 || gamma < 0 {
		t.Fatalf("missing randomized action traces alpha=%d beta=%d gamma=%d:\n%s", alpha, beta, gamma, mainBody)
	}
	if !(alpha < beta && beta < gamma) {
		t.Fatalf("target=test randomized actions should follow ivy2cpp sorted order alpha, beta, gamma; positions alpha=%d beta=%d gamma=%d\n%s", alpha, beta, gamma, mainBody)
	}
	for _, want := range []string{
		`__choices := 3.0 + 5.0`,
		`if __choice < 1.0 {`,
		`} else if __choice < 2.0 {`,
		`} else if __choice < 3.0 {`,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("target=test sorted dispatch ladder missing %q:\n%s", want, mainBody)
		}
	}
	if strings.Contains(mainBody, `> internal`) || strings.Contains(mainBody, `ivy.internal()`) {
		t.Fatalf("target=test randomized actions should ignore explicitly private actions:\n%s", mainBody)
	}
}

func TestTestMainUsesUnprefixedWeightForExtAction(t *testing.T) {
	mod := goivy.New()
	mod.Name = "extweight"
	mod.Actions.Set("ext:ping", goivy.NewSequence())
	mod.PublicActions.Set("ext:ping", true)
	mod.Attributes.Set("ping.weight", "7.0")

	out, err := Generate(mod, Config{Target: "test", ClassName: "extweight", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`__choices := 7.0 + 5.0`,
		`if __choice >= 7.0 {`,
		`if __choice < 7.0 {`,
		`fmt.Fprintln(__ivy_out, "> ping")`,
		`ivy.ext__ping()`,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("ext action weight main missing %q:\n%s", want, mainBody)
		}
	}
	if strings.Contains(mainBody, `__choices := 1.0 + 5.0`) {
		t.Fatalf("ext action weight should use ping.weight rather than default weight:\n%s", mainBody)
	}
}

func TestGeneratedTestMainWeightedLoopExecutesActions(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action heavy = {}
export heavy
attribute heavy.weight = "100.0"
action light = {}
export light
attribute light.weight = "100.0"
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "wtrun", TestIters: "40"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	assertNoUnsupportedGo(t, out.Source)
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=40", "runs=1", "seed=1", "delay=0", "wait=0")
	if err != nil {
		t.Fatalf("run weighted generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	for _, want := range []string{"> heavy\n", "> light\n", "test_completed\n"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("weighted generated run missing %q\nstdout:\n%s\nstderr:\n%s", want, stdout, stderr)
		}
	}
	if strings.Contains(stdout, "assertion_failed") || strings.Contains(stderr, "assertion failed") {
		t.Fatalf("weighted generated run failed unexpectedly\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
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
	out, err := Generate(mod, Config{Target: "test", ClassName: "multi_public", TestIters: "1"})
	if err == nil {
		t.Fatal("Generate should reject exporting a multi-return action")
	}
	if !strings.Contains(err.Error(), "cannot handle multiple output in exported actions: split") {
		t.Fatalf("expected multi-output rejection error, got: %v", err)
	}
	if out != nil {
		t.Fatalf("multi-output rejection should not return partial generated source:\n%s", out.Source)
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
		`fmt.Fprintf(os.Stderr, "syntax error in command argument\n")`,
		"ivy := newParamcase(p__initial)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("parameterized constructor source missing %q:\n%s", want, out.Source)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if strings.Contains(mainBody, "syntax error in parameter value initial") {
		t.Fatalf("positional parameter syntax errors should use C++ command-argument diagnostic:\n%s", mainBody)
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
	stdout, stderr, err = runBinary(t, bin, "green junk", "iters=0", "runs=1")
	if err == nil {
		t.Fatalf("trailing junk in positional parameter should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "syntax error in command argument") || strings.Contains(stderr, "syntax error in parameter value initial") {
		t.Fatalf("positional syntax diagnostic should match ivy2cpp\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestPositionalNumericModuleParamSyntaxDiagnosticLikeIvy2Cpp(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..5}
parameter initial : idx
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "posnumparam", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	for _, want := range []string{
		`fmt.Fprintf(os.Stderr, "syntax error in command argument\n")`,
		`fmt.Fprintf(os.Stderr, "parameter initial out of bounds\n")`,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("positional numeric param source missing %q:\n%s", want, mainBody)
		}
	}
	if strings.Contains(mainBody, "syntax error in parameter value initial") {
		t.Fatalf("positional numeric decode syntax should use command-argument diagnostic:\n%s", mainBody)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "abc", "iters=0", "runs=1")
	if err == nil {
		t.Fatalf("nonnumeric positional parameter should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "syntax error in command argument") || strings.Contains(stderr, "syntax error in parameter value initial") {
		t.Fatalf("positional numeric syntax diagnostic should match ivy2cpp\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
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
		`case "initial":`,
		`__ivy_param := __ivy_opt.value`,
		`delete(opts, "initial")`,
		`for _, __ivy_opt := range optArgs {`,
		"ivy := newParamdefault(p__initial)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("defaulted parameter source missing %q:\n%s", want, out.Source)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	overrideIdx := strings.Index(mainBody, `case "initial":`)
	unknownIdx := strings.Index(mainBody, `fmt.Fprintf(os.Stderr, "unknown option: %s\n", __ivy_opt.key)`)
	openIdx := strings.Index(mainBody, `os.Open(rest[0])`)
	if overrideIdx < 0 || unknownIdx < 0 || openIdx < 0 || !(overrideIdx < unknownIdx && unknownIdx < openIdx) {
		t.Fatalf("defaulted module params should be consumed before unknown options, and unknowns before command files (override=%d unknown=%d open=%d):\n%s", overrideIdx, unknownIdx, openIdx, mainBody)
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
	stdout, stderr, err = runBinary(t, bin, "initial=green", "unexpected=1", filepath.Join(t.TempDir(), "missing.in"), "iters=1", "runs=1", "seed=1")
	if err == nil {
		t.Fatalf("unknown option with defaulted param and missing command file should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "unknown option: unexpected") || strings.Contains(stderr, "cannot open to read:") {
		t.Fatalf("unknown option should win over command-file open failure after consuming defaulted param\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	stdout, stderr, err = runBinary(t, bin, "initial=blue", "initial=green", "iters=1", "runs=1", "seed=1")
	if err == nil {
		t.Fatalf("first bad duplicate defaulted param should fail before later good override\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "parameter initial out of bounds") {
		t.Fatalf("bad duplicate defaulted param diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestDefaultedModuleParamParserUsesSourceLineLikeIvy2Cpp(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter initial : color = red
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "paramlineno", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	defaultIdx := strings.Index(out.Source, `ivyParseValue("red")`)
	overrideIdx := strings.Index(out.Source, `case "initial":`)
	if defaultIdx < 0 || overrideIdx < 0 || defaultIdx > overrideIdx {
		t.Fatalf("default parser should run before argv override (default=%d override=%d):\n%s", defaultIdx, overrideIdx, out.Source)
	}
	defaultBlock := out.Source[defaultIdx:overrideIdx]
	for _, want := range []string{
		`fmt.Fprintf(os.Stderr, "test.ivy: line 3: syntax error in parameter value initial\n")`,
		`fmt.Fprintf(os.Stderr, "test.ivy: line 3: parameter initial out of bounds\n")`,
	} {
		if !strings.Contains(defaultBlock, want) {
			t.Fatalf("default parser missing line-prefixed diagnostic %q:\n%s", want, defaultBlock)
		}
	}
	overrideBlock := out.Source[overrideIdx:]
	if overrideBlock == "" {
		t.Fatalf("override parser block not found:\n%s", out.Source)
	}
	for _, want := range []string{
		`fmt.Fprintf(os.Stderr, "syntax error in parameter value initial\n")`,
		`fmt.Fprintf(os.Stderr, "parameter initial out of bounds\n")`,
	} {
		if !strings.Contains(overrideBlock, want) {
			t.Fatalf("override parser missing unprefixed diagnostic %q:\n%s", want, overrideBlock)
		}
	}
	if strings.Contains(overrideBlock, "test.ivy: line 3") {
		t.Fatalf("argv override parser should not inherit default source line:\n%s", overrideBlock)
	}
}

func TestModuleParamValueParserRejectsTrailingJunkByShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter initial : color = red
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "paramtrail", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	parser := bodyAfterMarker(out.Source, "func ivyParseValue(text string)")
	if parser == "" {
		t.Fatalf("ivyParseValue parser not found:\n%s", out.Source)
	}
	for _, want := range []string{
		"value, err := ivyParseValueAt(text, &pos)",
		"ivySkipWhite(text, &pos)",
		"if pos != len(text) {",
		"return value, ivySyntaxError(pos)",
		"return value, nil",
	} {
		if !strings.Contains(parser, want) {
			t.Fatalf("ivyParseValue should reject trailing junk; missing %q:\n%s", want, parser)
		}
	}
	if strings.Contains(parser, "return ivyParseValueAt(text, &pos)") {
		t.Fatalf("ivyParseValue should validate full consumption, not return the raw recursive parse:\n%s", parser)
	}
}

func TestGeneratedModuleParamRejectsTrailingJunk(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter initial : color = red
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "paramtrailrun", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "initial=green junk", "iters=0", "runs=1", "seed=1")
	if err == nil {
		t.Fatalf("trailing junk in parameter override should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "syntax error in parameter value initial") {
		t.Fatalf("trailing junk diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestMainParsesNatModuleParamAsNonnegative(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type budget
interpret budget -> nat
parameter initial : budget = 5
individual saved : budget
after init {
    saved := initial
}
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "natparam", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"p__initial := 5",
		`case "initial":`,
		`__ivy_param := __ivy_opt.value`,
		"if __ivy_num < 0 {",
		`fmt.Fprintf(os.Stderr, "parameter initial out of bounds\n")`,
		"p__initial = __ivy_num",
		"ivy := newNatparam(p__initial)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("nat parameter source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "__ivy_param >") {
		t.Fatalf("nat parameter should have only a lower-bound check:\n%s", out.Source)
	}
}

func TestGeneratedNatModuleParamRejectsNegativeOverride(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type budget
interpret budget -> nat
parameter initial : budget = 5
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "natparamrun", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "initial=-1", "iters=1", "runs=1", "seed=1")
	if err == nil {
		t.Fatalf("negative nat parameter override should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "parameter initial out of bounds") {
		t.Fatalf("negative nat parameter diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func newFunctionParamTestModule(t *testing.T) *goivy.Module {
	t.Helper()
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Name = "runner"
	color := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	mod.Sig.Sorts.Set("color", color)
	mod.SortOrder = append(mod.SortOrder, "color")
	slot := &goivy.UninterpretedSort{Name: "slot"}
	mod.Sig.Sorts.Set("slot", slot)
	mod.Sig.Interp["slot"] = &goivy.RangeSort{Name: "slot", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "1"}}
	mod.SortOrder = append(mod.SortOrder, "slot")
	fs, err := goivy.NewFunctionSort(slot, color)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	mod.Params = append(mod.Params, &goivy.Const{Name: "tbl", CSort: fs})
	mod.ParamDefaults = append(mod.ParamDefaults, nil)
	mod.Actions.Set("step", goivy.NewSequence())
	mod.PublicActions.Set("step", true)
	return mod
}

func newFunctionSortedActionParamTestModule(t *testing.T) *goivy.Module {
	t.Helper()
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Name = "runner"
	color := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	mod.Sig.Sorts.Set("color", color)
	mod.SortOrder = append(mod.SortOrder, "color")
	slot := &goivy.UninterpretedSort{Name: "slot"}
	mod.Sig.Sorts.Set("slot", slot)
	mod.Sig.Interp["slot"] = &goivy.RangeSort{Name: "slot", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "1"}}
	mod.SortOrder = append(mod.SortOrder, "slot")
	fs, err := goivy.NewFunctionSort(slot, color)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	tbl := goivy.NewConst("tbl", fs)
	outParam := goivy.NewConst("out", color)
	body := goivy.NewAssignAction(outParam, goivy.MustApply(tbl, goivy.NewConst("1", slot)))
	body.SetFormalParams([]*goivy.Const{tbl})
	body.SetFormalReturns([]*goivy.Const{outParam})
	mod.Actions.Set("read", body)
	mod.PublicActions.Set("read", true)
	return mod
}

func newLargeFunctionSortedActionParamTestModule(t *testing.T) *goivy.Module {
	t.Helper()
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Name = "largerandom"
	key := &goivy.UninterpretedSort{Name: "key"}
	mod.Sig.Sorts.Set("key", key)
	mod.SortOrder = append(mod.SortOrder, "key")
	fs, err := goivy.NewFunctionSort(key, key, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	tbl := goivy.NewConst("tbl", fs)
	outParam := goivy.NewConst("out", goivy.Boolean)
	body := goivy.NewAssignAction(outParam, goivy.MustApply(tbl, goivy.NewConst("0", key), goivy.NewConst("1", key)))
	body.SetFormalParams([]*goivy.Const{tbl})
	body.SetFormalReturns([]*goivy.Const{outParam})
	mod.Actions.Set("probe", body)
	mod.PublicActions.Set("probe", true)
	return mod
}

func TestFunctionSortedModuleParamParsesTupleList(t *testing.T) {
	mod := newFunctionParamTestModule(t)
	out, err := Generate(mod, Config{Target: "test", ClassName: "runner", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type ivyValue struct",
		"p__tbl := make(map[int]color)",
		"ivyParseValue(rest[0])",
		".atom != \"\"",
		".fields) != 2",
		"switch __ivy_param_item",
		`case "red":`,
		`case "green":`,
		"p__tbl[__ivy_arg",
		"ivy := newRunner(p__tbl)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("function-sorted parameter source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		"p__tbl := 0",
		"p__tbl = __ivy_param",
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("function-sorted parameter should not use scalar path %q:\n%s", bad, out.Source)
		}
	}
}

func TestGeneratedFunctionSortedModuleParamCompilesAndRuns(t *testing.T) {
	requireSlowTest(t)
	mod := newFunctionParamTestModule(t)
	out, err := Generate(mod, Config{Target: "test", ClassName: "runner", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "[[0,red],[1,green]]", "iters=0", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go with function parameter: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestStringModuleParamUsesIvyValueParser(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strlit
parameter initial : text
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "strparam", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"p__initial := \"\"",
		"ivyParseValue(rest[0])",
		".atom",
		"ivy := newStrparam(p__initial)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("strlit parameter source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "p__initial = rest[0]") {
		t.Fatalf("strlit parameter should not assign raw positional text:\n%s", out.Source)
	}
}

func TestGeneratedStringModuleParamParsesQuotedValue(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strlit
parameter initial : text
action check = {
    assert initial = "hello"
}
export check
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "strparamrun", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, `"hello"`, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go with quoted strlit parameter: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestScalarModuleParamsUseIvyValueParser(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type idx = {0..5}
parameter c : color
parameter i : idx
parameter b : bool
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "scalarparams", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"ivyParseValue(rest[0])",
		"ivyParseValue(rest[1])",
		"ivyParseValue(rest[2])",
		"switch __ivy_param",
		`case "red":`,
		`case "green":`,
		"if __ivy_num < 0 || __ivy_num > 5 {",
		"p__c = green",
		"p__i = __ivy_num",
		"p__b = true",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("scalar parameter source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "switch strings.ToLower(rest[2])") ||
		strings.Contains(out.Source, "strconv.Atoi(rest[1])") {
		t.Fatalf("scalar parameters should use ivyValue parsing, not raw text:\n%s", out.Source)
	}
}

func TestGeneratedScalarModuleParamsParseQuotedValues(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type idx = {0..5}
parameter c : color
parameter i : idx
parameter b : bool
action check = {
    assert c = green;
    assert i = 3;
    assert b
}
export check
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "scalarparamsrun", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, `"green"`, `"3"`, `"true"`, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go with quoted scalar params: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
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
		`func ivyAtoiOption(opts map[string]string, key string, def int) int`,
		`sleepMs := ivyAtoiOption(opts, "delay", 10)`,
		`finalMs := ivyAtoiOption(opts, "wait", 0)`,
		`__ivy_out_opened := false`,
		`case "modelfile":`,
		`fmt.Fprintln(__ivy_modelfile_file, "ivy2golang: modelfile solver logging is not implemented for generated Go")`,
		`time.Sleep(time.Duration(sleepMs) * time.Millisecond)`,
		`ivy.__tick(sleepMs)`,
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
	modelData, err := os.ReadFile(modelFile)
	if err != nil {
		t.Fatalf("read modelfile %s: %v", modelFile, err)
	}
	if !strings.Contains(string(modelData), "modelfile solver logging is not implemented") {
		t.Fatalf("modelfile should contain unsupported solver-log marker, got %q", modelData)
	}
}

func TestRuntimeOptionsUseIvy2CppAtoiAndPresenceSemantics(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
    assert false
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "runtimeparse", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`func ivyAtoi(text string) int`,
		`func ivyAtoiOption(opts map[string]string, key string, def int) int`,
		`testIters := ivyAtoiOption(opts, "iters", 1)`,
		`runs := ivyAtoiOption(opts, "runs", 1)`,
		`seed := ivyAtoiOption(opts, "seed", 1)`,
		`__ivy_out_opened := false`,
		`case "out":`,
		`case "modelfile":`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("runtime option source missing C++-style parse fragment %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `ivyAtoiDefault(opts["iters"]`) ||
		strings.Contains(out.Source, `if out := opts["out"]; out != ""`) {
		t.Fatalf("runtime options should distinguish absent from present empty values:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=bogus", "runs=1")
	if err != nil {
		t.Fatalf("nonnumeric iters should behave like C++ atoi and run zero iterations: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker for iters=bogus\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	stdout, stderr, err = runBinary(t, bin, "iters=1x", "runs=1", "seed=1")
	if err == nil {
		t.Fatalf("numeric-prefix iters should behave like C++ atoi and run one failing iteration\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "assertion failed") {
		t.Fatalf("numeric-prefix iters diagnostic missing assertion failure\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	stdout, stderr, err = runBinary(t, bin, "out=", "iters=0", "runs=1")
	if err == nil {
		t.Fatalf("present empty out= should attempt to open an empty path and fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "cannot open to write:") {
		t.Fatalf("empty out= diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	traceFile := filepath.Join(t.TempDir(), "trace.out")
	stdout, stderr, err = runBinary(t, bin, "out=", "out="+traceFile, "iters=0", "runs=1")
	if err == nil {
		t.Fatalf("earlier empty out= should fail before later good out= override\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "cannot open to write:") {
		t.Fatalf("ordered empty out= diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if _, statErr := os.Stat(traceFile); !os.IsNotExist(statErr) {
		t.Fatalf("later out= file should not be created after earlier empty out= failure; stat err=%v", statErr)
	}
}

func TestTargetTestMainLoopReinitializesEachRunByShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual saved : bool
after init {
    saved := false
}
action step = {
    saved := true
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "reruns", TestIters: "2", TestRuns: "3"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`runs := ivyAtoiOption(opts, "runs", 3)`,
		`ivySetSeed(seed)`,
		`for runidx := 0; runidx < runs; runidx++ {`,
		`ivy := newReruns()`,
		`ivy._generating = false`,
		`for cycle := 0; cycle < testIters; cycle++ {`,
		`fmt.Fprintln(__ivy_out, "test_completed")`,
		`if runidx == runs-1 {`,
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("target=test main loop source missing %q:\n%s", want, mainBody)
		}
	}
	seedIdx := strings.Index(mainBody, `ivySetSeed(seed)`)
	loopIdx := strings.Index(mainBody, `for runidx := 0; runidx < runs; runidx++ {`)
	newIdx := strings.Index(mainBody, `ivy := newReruns()`)
	doneIdx := strings.Index(mainBody, `fmt.Fprintln(__ivy_out, "test_completed")`)
	tailIdx := strings.Index(mainBody, `if runidx == runs-1 {`)
	if seedIdx < 0 || loopIdx < 0 || newIdx < 0 || doneIdx < 0 || tailIdx < 0 {
		t.Fatalf("target=test main loop pieces missing (seed=%d loop=%d new=%d done=%d tail=%d):\n%s", seedIdx, loopIdx, newIdx, doneIdx, tailIdx, mainBody)
	}
	if !(seedIdx < loopIdx && loopIdx < newIdx && newIdx < doneIdx && doneIdx < tailIdx) {
		t.Fatalf("target=test main should seed once, then create a fresh object and finish inside each run; seed=%d loop=%d new=%d done=%d tail=%d\n%s", seedIdx, loopIdx, newIdx, doneIdx, tailIdx, mainBody)
	}
	preLoop := mainBody[:loopIdx]
	if strings.Contains(preLoop, `newReruns()`) {
		t.Fatalf("target=test should not construct the Ivy object before the run loop:\n%s", mainBody)
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
		`opts, rest, optArgs := parseIvyTestArgs(os.Args[1:])`,
		`for _, __ivy_opt := range optArgs {`,
		`switch __ivy_opt.key {`,
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
	mainBody := bodyAfterMarker(out.Source, "func main()")
	unknownIdx := strings.Index(mainBody, `for _, __ivy_opt := range optArgs {`)
	openIdx := strings.Index(mainBody, `os.Open(rest[0])`)
	if unknownIdx < 0 || openIdx < 0 || unknownIdx > openIdx {
		t.Fatalf("unknown options should be rejected before opening command file args (unknown=%d open=%d):\n%s", unknownIdx, openIdx, mainBody)
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
	stdout, stderr, err = runBinary(t, bin, "unexpected=1", filepath.Join(t.TempDir(), "missing.in"))
	if err == nil {
		t.Fatalf("unknown option with missing command file should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "unknown option: unexpected") || strings.Contains(stderr, "cannot open to read:") {
		t.Fatalf("unknown option should win over command-file open failure\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
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
		`opts, rest, optArgs := parseIvyTestArgs(os.Args[1:])`,
		`if len(rest) == 1 {`,
		`os.Open(rest[0])`,
		`if len(rest) != 0 {`,
		`usage: smokerepl`,
		`for _, __ivy_opt := range optArgs {`,
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
	stdout, stderr, err = runBinary(t, bin, "first_bad=1", "second_bad=1")
	if err == nil {
		t.Fatalf("multiple unknown options should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "unknown option: first_bad") || strings.Contains(stderr, "unknown option: second_bad") {
		t.Fatalf("unknown option diagnostic should use first argv-order key\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
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
		`seed := ivyAtoiOption(opts, "seed", 1)`,
		`ivySetSeed(seed)`,
		`case "out":`,
		`__ivy_out = __ivy_out_file`,
		`case "modelfile":`,
		`fmt.Fprintln(__ivy_modelfile_file, "ivy2golang: modelfile solver logging is not implemented for generated Go")`,
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
	modelData, err := os.ReadFile(modelFile)
	if err != nil {
		t.Fatalf("read modelfile %s: %v", modelFile, err)
	}
	if !strings.Contains(string(modelData), "modelfile solver logging is not implemented") {
		t.Fatalf("modelfile should contain unsupported solver-log marker, got %q", modelData)
	}
}

func TestReplMainDispatchesParenthesizedExportedAction(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action echo(c:color) returns (out:color) = {
    out := c
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "gorepl"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`"bufio"`,
		"ivyParseCommand",
		"bufio.NewScanner(os.Stdin)",
		`case "echo":`,
		"if len(__ivy_args) != 1 {",
		"var __ivy_arg0 color",
		"ivy.echo(__ivy_arg0)",
		`fmt.Fprintf(__ivy_out, "= %s\n", ivyTraceValue(__ivy_result, false))`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("repl dispatch source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestReplParserRejectsWhitespaceArgDialectByShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action set(c:color) = {
}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "whitespacearg"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	parser := bodyAfterMarker(out.Source, "func ivyParseCommand(")
	if parser == "" {
		t.Fatalf("ivyParseCommand not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		"if pos == len(text) {",
		"return action, nil, nil",
		"if text[pos] != '(' {",
		"return \"\", nil, ivySyntaxError(pos)",
		"arg, err := ivyParseValueAt(text, &pos)",
	} {
		if !strings.Contains(parser, want) {
			t.Fatalf("parser should reject whitespace-separated args before value parsing; missing %q:\n%s", want, parser)
		}
	}
	setIdx := strings.Index(out.Source, `case "set":`)
	if setIdx < 0 {
		t.Fatalf("missing set dispatch case:\n%s", out.Source)
	}
	dispatch := out.Source[setIdx:]
	for _, want := range []string{
		"if len(__ivy_args) != 1 {",
		`fmt.Fprintf(os.Stderr, "action %s takes %d input parameters\n", __ivy_action, 1)`,
		"var __ivy_arg0 color",
	} {
		if !strings.Contains(dispatch, want) {
			t.Fatalf("dispatch should keep parenthesized-arity checks; missing %q:\n%s", want, dispatch)
		}
	}
	if strings.Contains(parser, "strings.Fields") || strings.Contains(parser, "FieldsFunc") {
		t.Fatalf("repl parser should not accept the old whitespace-split dialect:\n%s", parser)
	}
}

func TestReplMainDispatchParsesFunctionSortedActionParam(t *testing.T) {
	mod := newFunctionSortedActionParamTestModule(t)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`case "read":`,
		"if len(__ivy_args) != 1 {",
		"var __ivy_arg0 map[int]color",
		"__ivy_arg0 = make(map[int]color)",
		"for _, __ivy_param_item",
		".fields) != 2",
		"switch __ivy_param_item",
		`case "green":`,
		"__ivy_arg0[__ivy_arg",
		"out = tbl[1]",
		"__ivy_result := ivy.read(__ivy_arg0)",
		`fmt.Fprintf(os.Stderr, "line %d:%d: %s bad value\n", __ivy_lineno, __ivy_param_item`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("function-sorted repl arg source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "out = tbl(1)") {
		t.Fatalf("function-sorted action body should read from generated map storage, not call the map:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "unsupported repl function-sorted action parameter") {
		t.Fatalf("function-sorted repl action parameter should be decoded, not rejected:\n%s", out.Source)
	}
}

func TestTargetTestRandomizesFunctionSortedActionParam(t *testing.T) {
	mod := newFunctionSortedActionParamTestModule(t)
	out, err := Generate(mod, Config{Target: "test", ClassName: "fnargtest", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`gen.tbl = func() map[int]color {`,
		`make(map[int]color)`,
		`for __i0 := 0; __i0 < (1)+1; __i0++ {`,
		`ivy.___ivy_randomize(2, "read.tbl", 0)`,
		`__arg0 := read_generator.tbl`,
		`fmt.Fprintf(__ivy_out, "> read(%s)\n", ivyTraceValue(__arg0, false))`,
		`__ivy_result := ivy.read(__arg0)`,
		`fmt.Fprintf(__ivy_out, "= %s\n", ivyTraceValue(__ivy_result, false))`,
		`out = tbl[1]`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("target=test function-sorted action parameter source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		`__arg0 := 0`,
		`using zero value for non-enumerable sort`,
		`unsupported action generator parameter`,
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("target=test function-sorted action parameter should not fall back to %q:\n%s", bad, out.Source)
		}
	}
}

func TestTargetTestRejectsOpaqueUninterpretedActionParamLikeIvy2Cpp(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type data
action set(d:data) = {
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "opaqueparam", TestIters: "1", TestRuns: "1", Build: true})
	if err == nil {
		t.Fatalf("Generate should reject opaque uninterpreted test-generator inputs:\n%s", outSource(out))
	}
	if !strings.Contains(err.Error(), "cannot create test generator because type data is uninterpreted") {
		t.Fatalf("opaque uninterpreted diagnostic mismatch: %v", err)
	}
	if strings.Contains(err.Error(), "using zero value for non-enumerable sort") {
		t.Fatalf("opaque uninterpreted sort should hard-fail, not use zero fallback: %v", err)
	}
}

func TestUnsupportedActionParamErrorIncludesSourceLine(t *testing.T) {
	makeModule := func(t *testing.T) *goivy.Module {
		t.Helper()
		mod := compileIvySource(t, `#lang ivy1.7
type data
action set(d:data) = {
}
export set
`)
		action, ok := mod.Actions.Get2("set")
		if !ok {
			t.Fatal("action set not found")
		}
		params := action.GetFormalParams()
		if len(params) != 1 {
			t.Fatalf("set params = %d, want 1", len(params))
		}
		params[0].SetLineno(goivy.Location{Filename: "opaque_params.ivy", Line: 12})
		return mod
	}

	for _, tc := range []struct {
		name   string
		target string
		want   string
	}{
		{name: "target-test", target: "test", want: "opaque_params.ivy: line 12: unsupported target=test action generator parameter"},
		{name: "target-gen", target: "gen", want: "opaque_params.ivy: line 12: unsupported action generator parameter"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Generate(makeModule(t), Config{Target: tc.target, ClassName: "badparam", TestIters: "1", TestRuns: "1", Build: true})
			if err == nil {
				t.Fatalf("Generate target=%s should reject opaque action parameter", tc.target)
			}
			for _, want := range []string{
				tc.want,
				"cannot create test generator because type data is uninterpreted",
			} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("target=%s action parameter error missing %q:\n%v", tc.target, want, err)
				}
			}
		})
	}
}

func TestTargetTestBuildRejectsNonEnumerableZeroFallback(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type data
interpret data -> strlit
individual saved : data
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "nonenum", TestIters: "1", TestRuns: "1", Build: true})
	if err == nil {
		t.Fatalf("Generate should reject strict-build zero fallback:\n%s", outSource(out))
	}
	if !strings.Contains(err.Error(), "cannot create test generator because type data is non-enumerable") {
		t.Fatalf("non-enumerable fallback diagnostic mismatch: %v", err)
	}
	if strings.Contains(err.Error(), "using zero value for non-enumerable sort") {
		t.Fatalf("strict build should hard-fail, not warn-and-zero: %v", err)
	}
}

func TestTargetTestNonBuildKeepsNonEnumerableZeroFallbackWarning(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type data
interpret data -> strlit
individual saved : data
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "nonenumwarn", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("non-build generation should keep warning-only fallback: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, `ivy.saved = ""`) {
		t.Fatalf("non-build fallback should still emit zero value for inspection:\n%s", out.Source)
	}
	if len(out.Warnings) != 1 || !strings.Contains(out.Warnings[0], "using zero value for non-enumerable sort data") {
		t.Fatalf("non-build fallback warning mismatch: %#v", out.Warnings)
	}
}

func TestTargetTestRandomizesUnboundedIntInterpWithIvy2CppDefaultRange(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type data
interpret data -> int
individual saved : data
action set(d:data) = {
    saved := d
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "intdefault", TestIters: "1", TestRuns: "1", Build: true})
	if err != nil {
		t.Fatalf("Generate should randomize unbounded int-interpreted sorts over the C++ default range: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`ivy.saved = ivy.___ivy_choose(5, "init.saved", 0)`,
		`gen.d = ivy.___ivy_randomize(5, "set.fml:d", 0)`,
		`__arg0 := set_generator.d`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("unbounded int-interpreted randomization missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		"using zero value for non-enumerable sort data",
		"cannot create test generator because type data is non-enumerable",
	} {
		if strings.Contains(out.Source, bad) || strings.Contains(strings.Join(out.Warnings, "\n"), bad) {
			t.Fatalf("unbounded int-interpreted sort should not fall back to %q:\n%s\nwarnings:%v", bad, out.Source, out.Warnings)
		}
	}
}

func TestTargetTestBuildAllowsNativeSocketHandleSort(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type endpoint = {me}
type socket
individual sock : socket
action open(e:endpoint) returns (s:socket) = {
    <<< impure
        s = 0;
    >>>
}
after init {
    sock := open(me)
}
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "sockhandle", TestIters: "1", TestRuns: "1", Build: true})
	if err != nil {
		t.Fatalf("Generate should allow native socket handle state in strict test builds: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func (ivy *sockhandle) __initState()",
		`ivy.sock = 0`,
		"func (ivy *sockhandle) open(e endpoint) int",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("native socket handle source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestRuntimeSocketHandleSortRecognizesExtNativeFactory(t *testing.T) {
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	socket := &goivy.UninterpretedSort{Name: "client.sock.intf.socket"}
	open := goivy.NewNativeAction(mod.Cfg.AstCfg.NewNativeCode("impure\ns = 0;"))
	open.SetFormalReturns([]*goivy.Const{goivy.NewConst("s", socket)})
	mod.Actions.Set("ext:client.sock.intf.open", open)

	g := &Generator{Mod: mod}
	if !g.isRuntimeHandleSort(socket) {
		t.Fatalf("ext-prefixed native socket factory should mark %s as a runtime handle", sortName(socket))
	}
}

func TestTargetGenRejectsUnboundedInitialQuantifierLikeIvy2Cpp(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type item
relation marked(X:item)
axiom forall X:item. marked(X)
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "badinitgen"})
	if err == nil {
		t.Fatalf("Generate target=gen should reject unbounded initial quantifier:\n%s", outSource(out))
	}
	if !strings.Contains(err.Error(), "cannot enumerate quantified initial-state variable X:item") {
		t.Fatalf("target=gen initial quantifier diagnostic mismatch: %v", err)
	}
	if strings.Contains(err.Error(), "using zero value for non-enumerable sort") {
		t.Fatalf("target=gen should hard-fail initial quantifier, not fall back to zero value: %v", err)
	}
}

func TestTargetGenRandomizesFunctionSortedActionParam(t *testing.T) {
	mod := newFunctionSortedActionParamTestModule(t)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "fnarggen", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`type Fnarggen_read_generator struct {`,
		`tbl map[int]color`,
		`gen.tbl = func() map[int]color {`,
		`make(map[int]color)`,
		`for __i0 := 0; __i0 < (1)+1; __i0++ {`,
		`ivy.___ivy_randomize(2, "__fml:tbl", 0)`,
		`fmt.Fprintf(__ivy_out, "> read(%s)\n", ivyTraceValue(gen.tbl, false))`,
		`__res := ivy.read(gen.tbl)`,
		`fmt.Fprintf(__ivy_out, "= %s\n", ivyTraceValue(__res, false))`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("target=gen function-sorted action parameter source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `gen.tbl = 0`) {
		t.Fatalf("target=gen function-sorted action parameter should not use scalar zero fallback:\n%s", out.Source)
	}
}

func TestTargetTestRandomizesLargeFunctionSortedActionParam(t *testing.T) {
	mod := newLargeFunctionSortedActionParamTestModule(t)
	out, err := Generate(mod, Config{Target: "test", ClassName: "largerandom", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`gen.tbl = func() ivyThunkMap[struct{ A0 int; A1 int }, bool] {`,
		`newIvyThunkMap[struct{ A0 int; A1 int }, bool](false)`,
		`.base = func(__ivy_key struct{ A0 int; A1 int }) bool {`,
		`return ivy.___ivy_randomize(2, "probe.tbl", 0) != 0`,
		`__arg0 := probe_generator.tbl`,
		`out = tbl.Get(struct{ A0 int; A1 int }{0, 1})`,
		`__ivy_result := ivy.probe(__arg0)`,
		`fmt.Fprintf(__ivy_out, "= %s\n", ivyTraceValue(__ivy_result, false))`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("target=test large function-sorted action parameter source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `__arg0 := 0`) {
		t.Fatalf("large function-sorted action parameter should not use scalar zero fallback:\n%s", out.Source)
	}
}

func TestGeneratedTargetTestLargeFunctionSortedActionParam(t *testing.T) {
	requireSlowTest(t)
	mod := newLargeFunctionSortedActionParamTestModule(t)
	out, err := Generate(mod, Config{Target: "test", ClassName: "largerandomrun", TestIters: "2", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=2", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("run generated large function-sorted action parameter target=test: %v\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", err, stdout, stderr, out.Source)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGeneratedTargetTestFunctionSortedActionParam(t *testing.T) {
	requireSlowTest(t)
	mod := newFunctionSortedActionParamTestModule(t)
	out, err := Generate(mod, Config{Target: "test", ClassName: "fnargrun", TestIters: "2", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=2", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("run generated function-sorted action parameter target=test: %v\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", err, stdout, stderr, out.Source)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplMainDispatchParsesLargeFunctionSortedActionParam(t *testing.T) {
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Name = "largearg"
	key := &goivy.UninterpretedSort{Name: "key"}
	mod.Sig.Sorts.Set("key", key)
	mod.SortOrder = append(mod.SortOrder, "key")
	fs, err := goivy.NewFunctionSort(key, key, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	tbl := goivy.NewConst("tbl", fs)
	act := goivy.NewSequence()
	act.SetFormalParams([]*goivy.Const{tbl})
	mod.Actions.Set("probe", act)
	mod.PublicActions.Set("probe", true)

	out, err := Generate(mod, Config{Target: "repl", ClassName: "largearg"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`case "probe":`,
		"var __ivy_arg0 ivyThunkMap[struct{ A0 int; A1 int }, bool]",
		"__ivy_arg0 = newIvyThunkMap[struct{ A0 int; A1 int }, bool](false)",
		"if len(__ivy_param_item",
		".fields) != 3",
		"__ivy_arg0.Set(struct{ A0 int; A1 int }{",
		"ivy.probe(__ivy_arg0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("large function-sorted repl arg source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "__ivy_arg0[struct{ A0 int; A1 int }{") {
		t.Fatalf("large function-sorted repl arg should use thunk-map Set, not map indexing:\n%s", out.Source)
	}
}

func TestGeneratedReplFunctionSortedActionParam(t *testing.T) {
	requireSlowTest(t)
	mod := newFunctionSortedActionParamTestModule(t)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	dir := t.TempDir()
	valid := filepath.Join(dir, "valid.in")
	if err := os.WriteFile(valid, []byte("read([[0,red],[1,green]])\n"), 0o644); err != nil {
		t.Fatalf("write valid command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, valid)
	if err != nil {
		t.Fatalf("run generated repl function-sorted command: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "= green\n" {
		t.Fatalf("function-sorted repl output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}

	bad := filepath.Join(dir, "bad.in")
	if err := os.WriteFile(bad, []byte("read([[0,blue],[1,green]])\n"), 0o644); err != nil {
		t.Fatalf("write bad command file: %v", err)
	}
	stdout, stderr, err = runBinary(t, bin, bad)
	if err == nil {
		t.Fatalf("bad function value should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "line 1:") || !strings.Contains(stderr, "bad value: blue bad value") {
		t.Fatalf("bad function value diagnostic missing line/column and bad value\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplMainDispatchUsesPublicActionsOnly(t *testing.T) {
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Name = "goreplprivate"
	mod.Actions.Set("internal", goivy.NewSequence())
	mod.Actions.Set("step", goivy.NewSequence())
	mod.PublicActions.Set("internal", false)
	mod.PublicActions.Set("step", true)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "goreplprivate"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, `case "step":`) {
		t.Fatalf("repl dispatch missing public action:\n%s", out.Source)
	}
	if strings.Contains(out.Source, `case "internal":`) {
		t.Fatalf("repl dispatch should include only public actions:\n%s", out.Source)
	}
}

func TestReplMainNoPublicActionsMarksParsedArgsUsed(t *testing.T) {
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Name = "goreplnopublic"
	mod.Actions.Set("internal", goivy.NewSequence())
	out, err := Generate(mod, Config{Target: "repl", ClassName: "goreplnopublic"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "__ivy_action, __ivy_args, err := ivyParseCommand(") {
		t.Fatalf("repl main missing parsed command assignment:\n%s", out.Source)
	}
	if !strings.Contains(out.Source, "_ = __ivy_args") {
		t.Fatalf("repl main should mark parsed args used even without public actions:\n%s", out.Source)
	}
	if strings.Contains(out.Source, `case "internal":`) {
		t.Fatalf("repl dispatch should not expose private action:\n%s", out.Source)
	}
	if !strings.Contains(out.Source, "undefined action") {
		t.Fatalf("repl dispatch should still reject unknown actions:\n%s", out.Source)
	}
}

func TestGeneratedReplReadsParenthesizedCommandFile(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action echo(c:color) returns (out:color) = {
    out := c
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "goreplrun"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte("echo(green)\n"), 0o644); err != nil {
		t.Fatalf("write command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err != nil {
		t.Fatalf("run generated repl with command file: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "= green\n" {
		t.Fatalf("repl output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplMainReadsCommandsFromStdinByShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "stdinshape"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		"__ivy_repl_scanner = bufio.NewScanner(os.Stdin)",
		"for __ivy_repl_scanner.Scan() {",
		"__ivy_action, __ivy_args, err := ivyParseCommand(__ivy_repl_scanner.Text())",
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("repl stdin main source missing %q:\n%s", want, mainBody)
		}
	}
}

func TestGeneratedReplReadsCommandsFromStdin(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
after init {
    saved := red
}
action set(c:color) = {
    saved := c
}
action get returns (out:color) = {
    out := saved
}
export set
export get
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "stdinrepl"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinaryWithInput(t, bin, "set(green)\nget\n")
	if err != nil {
		t.Fatalf("run generated repl with stdin: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "= green\n" {
		t.Fatalf("stdin repl output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplParserRejectsEmptyParenthesizedCommandByShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "zerorepl"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	parser := bodyAfterMarker(out.Source, "func ivyParseCommand(")
	if parser == "" {
		t.Fatalf("ivyParseCommand not emitted:\n%s", out.Source)
	}
	preLoop := parser
	if idx := strings.Index(parser, "for {"); idx >= 0 {
		preLoop = parser[:idx]
	}
	if strings.Contains(preLoop, "text[pos] == ')'") || strings.Contains(preLoop, "return action, args, nil") {
		t.Fatalf("ivy2cpp-style parser should not accept empty parenthesized commands before parsing an arg:\n%s", parser)
	}
	if !strings.Contains(parser, "arg, err := ivyParseValueAt(text, &pos)") {
		t.Fatalf("parser should parse the first parenthesized argument immediately:\n%s", parser)
	}
}

func TestReplSyntaxErrorsIncludeLineAndPositionByShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replsyntax"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type ivySyntaxErr struct",
		"func ivySyntaxErrorPos(err error) (int, bool)",
		`fmt.Fprintf(os.Stderr, "line %d:%d: syntax error\n", __ivy_lineno, __ivy_pos)`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("syntax-error position source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestReplParserRejectsBlankLinesByShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "blankrepl"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	parser := bodyAfterMarker(out.Source, "func ivyParseCommand(")
	if parser == "" {
		t.Fatalf("ivyParseCommand not emitted:\n%s", out.Source)
	}
	if strings.Contains(parser, `return "", nil, nil`) {
		t.Fatalf("ivy2cpp-style parser should reject blank command lines, not return an empty action:\n%s", parser)
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if strings.Contains(mainBody, `if __ivy_action == ""`) {
		t.Fatalf("repl main should not skip empty parsed actions; parser rejects them:\n%s", mainBody)
	}
}

func TestGeneratedReplRejectsEmptyParenthesizedCommand(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "zeroreplrun"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte("step()\n"), 0o644); err != nil {
		t.Fatalf("write command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err == nil {
		t.Fatalf("empty parenthesized command should fail like ivy2cpp\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "line 1:5: syntax error") {
		t.Fatalf("empty parenthesized command should report syntax error\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGeneratedReplRejectsBlankCommandLine(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "blankreplrun"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte("\n"), 0o644); err != nil {
		t.Fatalf("write command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err == nil {
		t.Fatalf("blank command line should fail like ivy2cpp\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "line 1:0: syntax error") {
		t.Fatalf("blank command line should report positioned syntax error\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplTraceClosesAssertBraces(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
    assert false
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "repltracefail", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	assertBody := bodyAfterMarker(out.Source, "func ivyAssert(")
	if assertBody == "" {
		t.Fatalf("ivyAssert helper not emitted:\n%s", out.Source)
	}
	assumeBody := bodyAfterMarker(out.Source, "func ivyAssume(")
	if assumeBody == "" {
		t.Fatalf("ivyAssume helper not emitted:\n%s", out.Source)
	}
	for name, body := range map[string]string{"ivyAssert": assertBody, "ivyAssume": assumeBody} {
		if !strings.Contains(body, `fmt.Fprintln(__ivy_out, "}")`) {
			t.Fatalf("%s should close trace braces in traced repl output:\n%s", name, body)
		}
	}
	dispatch := bodyAfterMarker(out.Source, "func main()")
	if !strings.Contains(dispatch, `fmt.Fprintln(__ivy_out, "step {")`) {
		t.Fatalf("traced repl dispatch should open action trace block:\n%s", dispatch)
	}
}

func TestGeneratedReplTraceAssertFailureClosesBrace(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
    assert false
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "repltracefailrun", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte("step\n"), 0o644); err != nil {
		t.Fatalf("write command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err == nil {
		t.Fatalf("assertion failure should exit nonzero\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	want := "step {\nassertion_failed(\"test.ivy: line 3\")\n}\n"
	if stdout != want {
		t.Fatalf("traced repl assertion output differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", want, stdout, stderr)
	}
	if !strings.Contains(stderr, "test.ivy: line 3: error: assertion failed") {
		t.Fatalf("assertion failure stderr missing line label\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplMainDispatchParsesEnumBoolRangeArgs(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type idx = {0..2}
individual saved : color
individual ok : bool
individual seen : idx
action set(c:color,b:bool,i:idx) = {
    saved := c;
    ok := b;
    seen := i
}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replargs"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`case "set":`,
		"if len(__ivy_args) != 3 {",
		"var __ivy_arg0 color",
		"var __ivy_arg1 bool",
		"var __ivy_arg2 int",
		"switch __ivy_args[0].atom",
		"switch __ivy_args[1].atom",
		"if __ivy_num < 0 || __ivy_num > 2 {",
		`fmt.Fprintf(os.Stderr, "line %d:%d: %s bad value\n", __ivy_lineno, __ivy_args[0].pos, "bad value: " + __ivy_args[0].atom)`,
		`fmt.Fprintf(os.Stderr, "line %d:%d: %s bad value\n", __ivy_lineno, __ivy_args[2].pos, "argument 3")`,
		"ivy.set(__ivy_arg0, __ivy_arg1, __ivy_arg2)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("repl arg dispatch source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestReplMainDispatchParsesNamedEnumByNameAndNumericEnumByNumber(t *testing.T) {
	named := compileIvySource(t, `#lang ivy1.7
type color = {red, green, blue}
action echo(c:color) returns (out:color) = {
    out := c
}
export echo
`)
	namedOut, err := Generate(named, Config{Target: "repl", ClassName: "enumparse"})
	if err != nil {
		t.Fatalf("Generate named enum: %v\n%s", err, outSource(namedOut))
	}
	for _, want := range []string{
		"var __ivy_arg0 color",
		"switch __ivy_args[0].atom",
		`case "red":`,
		"__ivy_arg0 = red",
		`case "green":`,
		"__ivy_arg0 = green",
		`case "blue":`,
		"__ivy_arg0 = blue",
		"ivy.echo(__ivy_arg0)",
	} {
		if !strings.Contains(namedOut.Source, want) {
			t.Fatalf("named enum parser source missing %q:\n%s", want, namedOut.Source)
		}
	}

	numeric := compileIvySource(t, `#lang ivy1.7
type digit = {0, 1, 2}
action echo(d:digit) returns (out:digit) = {
    out := d
}
export echo
`)
	numericOut, err := Generate(numeric, Config{Target: "repl", ClassName: "numenum"})
	if err != nil {
		t.Fatalf("Generate numeric enum: %v\n%s", err, outSource(numericOut))
	}
	for _, want := range []string{
		"var __ivy_arg0 int",
		"__ivy_num, err := strconv.Atoi(__ivy_args[0].atom)",
		"if __ivy_num < 0 || __ivy_num > 2 {",
		"__ivy_arg0 = __ivy_num",
		"ivy.echo(__ivy_arg0)",
	} {
		if !strings.Contains(numericOut.Source, want) {
			t.Fatalf("numeric enum parser source missing %q:\n%s", want, numericOut.Source)
		}
	}
	for _, bad := range []string{
		"type digit int",
		"switch __ivy_args[0].atom",
		`case "0":`,
	} {
		if strings.Contains(numericOut.Source, bad) {
			t.Fatalf("numeric enum should route through primitive numeric parser, found %q:\n%s", bad, numericOut.Source)
		}
	}
}

func TestGeneratedReplEnumBoolRangeArgsAndDiagnostics(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type idx = {0..2}
individual saved : color
individual ok : bool
individual seen : idx
action set(c:color,b:bool,i:idx) = {
    saved := c;
    ok := b;
    seen := i
}
action check returns (out:idx) = {
    assert saved = green;
    assert ok;
    out := seen
}
export set
export check
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replargsrun"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	dir := t.TempDir()
	commandFile := filepath.Join(dir, "commands.in")
	if err := os.WriteFile(commandFile, []byte("set(green,true,2)\ncheck\n"), 0o644); err != nil {
		t.Fatalf("write command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err != nil {
		t.Fatalf("run generated repl valid command: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "= 2\n" {
		t.Fatalf("valid repl output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}

	badArity := filepath.Join(dir, "bad_arity.in")
	if err := os.WriteFile(badArity, []byte("set(green,true)\n"), 0o644); err != nil {
		t.Fatalf("write bad arity command file: %v", err)
	}
	stdout, stderr, err = runBinary(t, bin, badArity)
	if err == nil {
		t.Fatalf("bad arity should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "action set takes 3 input parameters") {
		t.Fatalf("bad arity diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}

	badEnum := filepath.Join(dir, "bad_enum.in")
	if err := os.WriteFile(badEnum, []byte("set(blue,true,1)\n"), 0o644); err != nil {
		t.Fatalf("write bad enum command file: %v", err)
	}
	stdout, stderr, err = runBinary(t, bin, badEnum)
	if err == nil {
		t.Fatalf("bad enum should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "line 1:4: bad value: blue bad value") {
		t.Fatalf("bad enum diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}

	badRange := filepath.Join(dir, "bad_range.in")
	if err := os.WriteFile(badRange, []byte("set(green,true,3)\n"), 0o644); err != nil {
		t.Fatalf("write bad range command file: %v", err)
	}
	stdout, stderr, err = runBinary(t, bin, badRange)
	if err == nil {
		t.Fatalf("bad range should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "line 1:15: argument 3 bad value") {
		t.Fatalf("bad range diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplMainDispatchParsesNatAndStrlitArgs(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..5}
type bigint
interpret bigint -> nat
type str
interpret str -> strlit
individual ok : bool
individual seen_i : idx
individual seen_n : bigint
individual seen_s : str
action poke(b:bool,i:idx,n:bigint,s:str) = {
    ok := b;
    seen_i := i;
    seen_n := n;
    seen_s := s
}
export poke
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replprim"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`case "poke":`,
		"if len(__ivy_args) != 4 {",
		"var __ivy_arg0 bool",
		"var __ivy_arg1 int",
		"var __ivy_arg2 int",
		"var __ivy_arg3 string",
		"switch __ivy_args[0].atom",
		`case "true":`,
		`case "false":`,
		"__ivy_num, err := strconv.Atoi(__ivy_args[1].atom)",
		"if __ivy_num < 0 || __ivy_num > 5 {",
		"if __ivy_num < 0 {",
		`fmt.Fprintf(os.Stderr, "line %d:%d: %s bad value\n", __ivy_lineno, __ivy_args[2].pos, "argument 3")`,
		"__ivy_arg3 = __ivy_args[3].atom",
		"ivy.poke(__ivy_arg0, __ivy_arg1, __ivy_arg2, __ivy_arg3)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("repl primitive dispatch source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "strconv.Atoi(__ivy_args[3].atom)") {
		t.Fatalf("strlit repl argument should decode as an atom string, not a number:\n%s", out.Source)
	}
}

func TestReplValueParserHandlesStringEscapesByShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strlit
action echo(s:text) returns(out:text) = {
    out := s
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replescapes"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	parser := bodyAfterMarker(out.Source, "func ivyParseValueAt(")
	if parser == "" {
		t.Fatalf("ivyParseValueAt not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		"if c == '\\\\' {",
		"case 'n':",
		"c = 10",
		"case 'r':",
		"c = 13",
		"case 't':",
		"c = 9",
		"res.atom += string(c)",
	} {
		if !strings.Contains(parser, want) {
			t.Fatalf("string escape parser missing %q:\n%s", want, parser)
		}
	}
	if !strings.Contains(out.Source, "__ivy_arg0 = __ivy_args[0].atom") {
		t.Fatalf("strlit dispatch should receive the unescaped parsed atom:\n%s", out.Source)
	}
}

func TestGeneratedReplNatStrlitArgsAndDiagnostics(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..5}
type bigint
interpret bigint -> nat
type str
interpret str -> strlit
individual ok : bool
individual seen_i : idx
individual seen_n : bigint
individual seen_s : str
action poke(b:bool,i:idx,n:bigint,s:str) = {
    ok := b;
    seen_i := i;
    seen_n := n;
    seen_s := s
}
action check returns (out:str) = {
    assert ok;
    assert seen_i = 5;
    assert seen_n = 7;
    out := seen_s
}
export poke
export check
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replprimrun"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	dir := t.TempDir()
	commandFile := filepath.Join(dir, "commands.in")
	if err := os.WriteFile(commandFile, []byte("poke(true,5,7,\"hello\")\ncheck\n"), 0o644); err != nil {
		t.Fatalf("write command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err != nil {
		t.Fatalf("run generated repl primitive command: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "= hello\n" {
		t.Fatalf("primitive repl output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}

	badNat := filepath.Join(dir, "bad_nat.in")
	if err := os.WriteFile(badNat, []byte("poke(true,5,-1,\"hello\")\n"), 0o644); err != nil {
		t.Fatalf("write bad nat command file: %v", err)
	}
	stdout, stderr, err = runBinary(t, bin, badNat)
	if err == nil {
		t.Fatalf("bad nat should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "line 1:12: argument 3 bad value") {
		t.Fatalf("bad nat diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplMainDispatchParsesUninterpretedArgs(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
individual saved : node
relation edge(X:node,Y:node)
action set(n:node) = {
    assume edge(n,7);
    saved := n
}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`case "set":`,
		"var __ivy_arg0 int",
		`__ivy_num, err := strconv.Atoi(__ivy_args[0].atom)`,
		"ivy.set(__ivy_arg0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("repl uninterpreted dispatch source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		"unsupported repl",
		"cannot create test generator because type node is uninterpreted",
		"using zero value for non-enumerable sort node",
	} {
		if strings.Contains(out.Source, bad) || strings.Contains(strings.Join(out.Warnings, "\n"), bad) {
			t.Fatalf("plain uninterpreted repl argument should not fall back to %q:\n%s\nwarnings:%v", bad, out.Source, out.Warnings)
		}
	}
}

func TestReplMainDispatchParsesBVArgWithWidthBound(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type word
interpret word -> bv[8]
action echo(w:word) returns(out:word) = {
    out := w
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replbv"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`case "echo":`,
		"var __ivy_arg0 int",
		"if __ivy_num < 0 || __ivy_num > 255 {",
		`fmt.Fprintf(os.Stderr, "line %d:%d: %s bad value\n", __ivy_lineno, __ivy_args[0].pos, "argument 1")`,
		"__ivy_result := ivy.echo(__ivy_arg0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("repl bv dispatch source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestGeneratedReplBVArgRejectsOutOfRange(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type word
interpret word -> bv[8]
action echo(w:word) returns(out:word) = {
    out := w
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replbvrun"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	dir := t.TempDir()
	valid := filepath.Join(dir, "valid.in")
	if err := os.WriteFile(valid, []byte("echo(255)\n"), 0o644); err != nil {
		t.Fatalf("write valid command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, valid)
	if err != nil {
		t.Fatalf("run generated repl valid bv command: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "= 255\n" {
		t.Fatalf("valid bv repl output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}

	bad := filepath.Join(dir, "bad.in")
	if err := os.WriteFile(bad, []byte("echo(256)\n"), 0o644); err != nil {
		t.Fatalf("write bad command file: %v", err)
	}
	stdout, stderr, err = runBinary(t, bin, bad)
	if err == nil {
		t.Fatalf("out-of-range bv should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "line 1:5: argument 1 bad value") {
		t.Fatalf("out-of-range bv diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestTargetTestIntBVActionInputUsesIvy2CppCardinality(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type small
interpret small -> intbv[10][20][4]
individual saved : small
action set(x:small) = {
    saved := x
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "intbvargs", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"saved int",
		`ivy.saved = ivy.___ivy_choose(11, "init.saved", 0)`,
		`gen.x = ivy.___ivy_randomize(11, "set.fml:x", 0)`,
		`__arg0 := set_generator.x`,
		"ivy.saved = x",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("intbv action input source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `___ivy_randomize(1024, "set.fml:x", 0)`) {
		t.Fatalf("intbv[10][20][4] should use hi-lo+1 cardinality, not the first bracket as bit width:\n%s", out.Source)
	}
}

func TestReplMainDispatchParsesIntBVArgAsNumber(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type small
interpret small -> intbv[0][7][3]
individual saved : small
action set(x:small) returns(out:small) = {
    saved := x;
    out := x
}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replintbv"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`case "set":`,
		"var __ivy_arg0 int",
		`if len(__ivy_args[0].fields) != 0 {`,
		`fmt.Fprintf(os.Stderr, "line %d:%d: %s bad value\n", __ivy_lineno, __ivy_args[0].pos, "argument 1")`,
		`__ivy_num, err := strconv.Atoi(__ivy_args[0].atom)`,
		"__ivy_arg0 = __ivy_num",
		"__ivy_result := ivy.set(__ivy_arg0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("repl intbv dispatch source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "__ivy_num < 0") || strings.Contains(out.Source, "__ivy_num > 7") {
		t.Fatalf("intbv REPL parser should mirror ivy2cpp and parse the backing atom without a range guard:\n%s", out.Source)
	}
}

func TestReplMainDispatchParsesStrBVArgAsString(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strbv[4]
individual saved : text
action set(t:text) returns(out:text) = {
    saved := t;
    out := t
}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replstrbv"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`case "set":`,
		"saved string",
		"var __ivy_arg0 string",
		"if len(__ivy_args[0].fields) != 0 {",
		"__ivy_arg0 = __ivy_args[0].atom",
		"__ivy_result := ivy.set(__ivy_arg0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("repl strbv dispatch source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "strconv.Atoi(__ivy_args[0].atom)") {
		t.Fatalf("strbv repl argument should decode as an atom string, not a number:\n%s", out.Source)
	}
}

func TestTargetTestStrBVActionInputUsesCardinalityAndStrings(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strbv[4]
individual saved : text
action set(t:text) = {
    saved := t
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "strbvargs", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"saved string",
		`ivy.saved = strconv.Itoa(ivy.___ivy_choose(16, "init.saved", 0))`,
		`gen.t = strconv.Itoa(ivy.___ivy_randomize(16, "set.fml:t", 0))`,
		`__arg0 := set_generator.t`,
		"ivy.saved = t",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("strbv action input source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `__arg0 := ivy.___ivy_randomize(16, "set.fml:t", 0)`) {
		t.Fatalf("strbv randomized action input should convert the finite choice to string:\n%s", out.Source)
	}
}

func TestStrBVFiniteQuantifierEnumeratesStringValues(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strbv[2]
relation seen(T:text)
action check = {
    assert forall T. seen(T) -> seen(T)
}
export check
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "strbvquant", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"seen map[string]bool",
		`for _, T := range []string{"0", "1", "2", "3"} {`,
		"ivy.seen[T]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("strbv finite quantifier source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "cannot enumerate quantified variable") {
		t.Fatalf("strbv quantifier should enumerate the interpreted finite domain:\n%s", out.Source)
	}
}

func TestGeneratedReplStrBVArgAndDiagnostics(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strbv[4]
individual saved : text
action set(t:text) returns(out:text) = {
    saved := t;
    out := t
}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replstrbvrun"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	dir := t.TempDir()
	commandFile := filepath.Join(dir, "commands.in")
	if err := os.WriteFile(commandFile, []byte("set(\"hello\")\n"), 0o644); err != nil {
		t.Fatalf("write strbv command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err != nil {
		t.Fatalf("run generated repl strbv command: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "= hello\n" {
		t.Fatalf("strbv repl output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}

	bad := filepath.Join(dir, "bad.in")
	if err := os.WriteFile(bad, []byte("set({bogus:true})\n"), 0o644); err != nil {
		t.Fatalf("write bad strbv command file: %v", err)
	}
	stdout, stderr, err = runBinary(t, bin, bad)
	if err == nil {
		t.Fatalf("nested strbv value should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "line 1:4: argument 1 bad value") {
		t.Fatalf("nested strbv diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplMainDispatchParsesDestructorRecordArg(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type leaf
destructor shade(L:leaf) : color
type cell
destructor child(C:cell) : leaf
destructor ok(C:cell) : bool
action echo(c:cell) returns (out:cell) = {
    out := c
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replrecord"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type cell struct {",
		"var __ivy_arg0 cell",
		"for _, __ivy_field",
		`case "child":`,
		`case "ok":`,
		`case "shade":`,
		`fmt.Fprintf(os.Stderr, "line %d:%d: %s bad value\n", __ivy_lineno, __ivy_field`,
		`"unexpected field: " + __ivy_field`,
		"ivy.echo(__ivy_arg0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("repl record dispatch source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestGeneratedReplDestructorRecordArgAndDiagnostics(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type leaf
destructor shade(L:leaf) : color
type cell
destructor child(C:cell) : leaf
destructor ok(C:cell) : bool
action echo(c:cell) returns (out:cell) = {
    out := c
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replrecordrun"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	dir := t.TempDir()
	commandFile := filepath.Join(dir, "commands.in")
	if err := os.WriteFile(commandFile, []byte("echo({child:{shade:green},ok:true})\n"), 0o644); err != nil {
		t.Fatalf("write record command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err != nil {
		t.Fatalf("run generated repl record command: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "= {child:{shade:green},ok:true}\n" {
		t.Fatalf("record repl output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}

	badField := filepath.Join(dir, "bad_field.in")
	if err := os.WriteFile(badField, []byte("echo({bogus:true})\n"), 0o644); err != nil {
		t.Fatalf("write bad field command file: %v", err)
	}
	stdout, stderr, err = runBinary(t, bin, badField)
	if err == nil {
		t.Fatalf("bad record field should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "line 1:0: unexpected field: bogus bad value") {
		t.Fatalf("bad record field diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplParsesMultiArgDestructorRecordList(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green, blue}
type idx = {0..3}
type cell
destructor shade(C:cell, I:idx) : color
action echo(c:cell) returns (out:cell) = {
    out := c
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replindexedrecord"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type cell struct {",
		"shade [4]color",
		".shade[0]",
		".shade[1]",
		".shade[2]",
		".shade[3]",
		".atom != \"\" || len(",
		".fields) != 4",
		`__b.WriteString(fmt.Sprintf("%v", v.shade[__i0]))`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("indexed record source missing %q:\n%s", want, out.Source)
		}
	}

	bin := compileGeneratedGo(t, out)
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte("echo({shade:[red,green,blue,red]})\n"), 0o644); err != nil {
		t.Fatalf("write indexed record command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err != nil {
		t.Fatalf("run generated repl indexed record command: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "= {shade:[red,green,blue,red]}\n" {
		t.Fatalf("indexed record repl output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplMainDispatchParsesVariantArg(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type msg
type color = {red, green}
variant request of msg = struct {
    shade : color
}
variant ack of msg
action echo(m:msg) returns (out:msg) = {
    out := m
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replvariant"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type msg struct {",
		"var __ivy_arg0 msg",
		`if __ivy_args[0].atom != "" {`,
		`"unexpected value for sort msg: " + __ivy_args[0].atom`,
		`if len(__ivy_args[0].fields) > 1 {`,
		`"too many fields for sort msg (expected one)"`,
		"switch __ivy_args[0].fields[0].atom",
		`case "request":`,
		`case "ack":`,
		"msg{tag: 0",
		"msg{tag: 1",
		"valid: true",
		`fmt.Fprintf(os.Stderr, "line %d:%d: %s bad value\n", __ivy_lineno, __ivy_args[0].pos, "unexpected field sort SORTNAME: " + __ivy_args[0].fields[0].atom)`,
		"ivy.echo(__ivy_arg0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("repl variant dispatch source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestGeneratedReplVariantArgAndDiagnostics(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type msg
type color = {red, green}
variant request of msg = struct {
    shade : color
}
variant ack of msg
action echo(m:msg) returns (out:msg) = {
    out := m
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replvariantrun"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	dir := t.TempDir()
	requestFile := filepath.Join(dir, "request.in")
	if err := os.WriteFile(requestFile, []byte("echo({request:{shade:green}})\n"), 0o644); err != nil {
		t.Fatalf("write request command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, requestFile)
	if err != nil {
		t.Fatalf("run generated repl variant request: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "= {request:{shade:green}}\n" {
		t.Fatalf("variant request output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}

	ackFile := filepath.Join(dir, "ack.in")
	if err := os.WriteFile(ackFile, []byte("echo({ack:7})\n"), 0o644); err != nil {
		t.Fatalf("write ack command file: %v", err)
	}
	stdout, stderr, err = runBinary(t, bin, ackFile)
	if err != nil {
		t.Fatalf("run generated repl variant ack: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "= {ack:7}\n" {
		t.Fatalf("variant ack output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}

	badTag := filepath.Join(dir, "bad_tag.in")
	if err := os.WriteFile(badTag, []byte("echo({bogus:0})\n"), 0o644); err != nil {
		t.Fatalf("write bad tag command file: %v", err)
	}
	stdout, stderr, err = runBinary(t, bin, badTag)
	if err == nil {
		t.Fatalf("bad variant tag should fail\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "line 1:5: unexpected field sort SORTNAME: bogus bad value") {
		t.Fatalf("bad variant tag diagnostic missing\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestTargetGenEmitsRunnableSingleShotMain(t *testing.T) {
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
		`func ivy2golangGenerate(ivy *genrunner)`,
		`ivy2golangGenerate(ivy)`,
		`testIters := ivyAtoiOption(opts, "iters", 1)`,
		`_ = testIters`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("gen target source missing single-shot runner fragment %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		`"time"`,
		`for cycle := 0; cycle < testIters; cycle++ {`,
		`__choice := float64(ivyRand31()) * __choices / 2147483648.0`,
		`fmt.Fprintln(__ivy_out, "test_completed")`,
		`for runidx := 0; runidx < runs; runidx++ {`,
		`_ = newgenrunner()`,
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("gen target should not emit test/smoke runner fragment %q:\n%s", bad, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go gen target: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "> step\n" {
		t.Fatalf("gen target should execute each generator once and omit completion marker\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestTargetGenActionGeneratorsStoreRandomizedInputs(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type idx = {0..2}
individual saved : color
individual seen : idx
individual ok : bool
action set(c:color,b:bool,i:idx) = {
    saved := c;
    ok := b;
    seen := i
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "genparams", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`type Genparams_set_generator struct {`,
		`ivy *genparams`,
		`c color`,
		`b bool`,
		`i int`,
		`gen.c = color(ivy.___ivy_randomize(2, "__fml:c", 0))`,
		`gen.b = ivy.___ivy_randomize(2, "__fml:b", 1) != 0`,
		`gen.i = (0 + ivy.___ivy_randomize(3, "__fml:i", 2))`,
		`return true`,
		`fmt.Fprintf(__ivy_out, "> set(%s,%s,%s)\n", ivyTraceValue(gen.c, false), ivyTraceValue(gen.b, false), ivyTraceValue(gen.i, false))`,
		`ivy.set(gen.c, gen.b, gen.i)`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("gen parameter source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `__arg0 :=`) || strings.Contains(out.Source, `ivy.set(__arg0`) {
		t.Fatalf("target=gen execute should use generator fields, not local randomized args:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=30", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("run generated Go gen target: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "> set(") || strings.Contains(stdout, "test_completed") || strings.Count(stdout, "> set(") != 1 {
		t.Fatalf("gen target should run the exported generator once without a completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestTargetGenOpaqueParamMarksIvyReceiverUsed(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type data
action accept(d:data) = {
}
export accept
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "unusedivy", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (gen *Unusedivy_accept_generator) generate() bool")
	if body == "" {
		t.Fatalf("missing accept generator body:\n%s", out.Source)
	}
	for _, want := range []string{
		"ivy := gen.ivy",
		"_ = ivy",
		"gen.d = 0",
		"return true",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("opaque-param generator body missing %q:\n%s", want, body)
		}
	}
}

func TestTargetGenRandomizesNumericAndVariantSuperParams(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node = {0..4}
type t
variant a of t
individual saved : node
individual wrapped : t
action touch(n:node,av:t) = {
    saved := n;
    wrapped := av
}
export touch
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "gennumeric", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (gen *Gennumeric_touch_generator) generate() bool")
	if body == "" {
		t.Fatalf("touch generator body not found:\n%s", out.Source)
	}
	for _, want := range []string{
		`n int`,
		`av t`,
		`gen.n = (0 + ivy.___ivy_randomize(5, "__fml:n", 0))`,
		`gen.av = func() t {`,
		`switch ivy.___ivy_randomize(1, "__fml:av", 1) {`,
		`return t{tag: 0, value: 0, valid: true}`,
		`ivy.touch(gen.n, gen.av)`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("target=gen numeric/variant source missing %q:\n%s", want, out.Source)
		}
	}
	runnerBody := bodyAfterMarker(out.Source, "func ivy2golangGenerate(ivy *gennumeric)")
	if runnerBody == "" {
		t.Fatalf("gen runner body not found:\n%s", out.Source)
	}
	for _, want := range []string{
		`ivy.wrapped = func() t {`,
		`switch ivy.___ivy_randomize(1, "randomize.wrapped", 0) {`,
		`return t{tag: 0, value: 0, valid: true}`,
	} {
		if !strings.Contains(runnerBody, want) {
			t.Fatalf("target=gen variant state randomization missing %q:\n%s", want, runnerBody)
		}
	}
	if strings.Contains(runnerBody, `ivy.wrapped = t{}`) {
		t.Fatalf("target=gen variant state should be randomized, not zeroed:\n%s", runnerBody)
	}
	if strings.Contains(body, "using zero value for non-enumerable sort") ||
		strings.Contains(body, "unsupported action generator parameter") {
		t.Fatalf("target=gen numeric/variant generator should not fall back to unsupported/zero comments:\n%s", body)
	}
	compileGeneratedGo(t, out)
}

func TestTargetTestRandomizesVariantSuperState(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t
variant a of t
variant b of t
individual wrapped : t
action observe returns(out:t) = {
    out := wrapped
}
export observe
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "testvariantstate", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	initBody := bodyAfterMarker(out.Source, "func (ivy *testvariantstate) __initState()")
	if initBody == "" {
		t.Fatalf("__initState body not found:\n%s", out.Source)
	}
	for _, want := range []string{
		`ivy.wrapped = func() t {`,
		`switch ivy.___ivy_choose(2, "init.wrapped", 0) {`,
		`return t{tag: 0, value: 0, valid: true}`,
		`return t{tag: 1, value: 0, valid: true}`,
	} {
		if !strings.Contains(initBody, want) {
			t.Fatalf("target=test variant state randomization missing %q:\n%s", want, initBody)
		}
	}
	if strings.Contains(initBody, `ivy.wrapped = t{}`) {
		t.Fatalf("target=test variant state should be randomized, not zeroed:\n%s", initBody)
	}
}

func TestTargetGenRandomizesRangeSortWithOffsetAndWidth(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type rng = {2..7}
individual r : rng
action step(x:rng) = {
    r := x
}
export step
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "genrange", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`r int`,
		`x int`,
		`gen.x = (2 + ivy.___ivy_randomize(6, "__fml:x", 0))`,
		`ivy.r = (2 + ivy.___ivy_randomize(6, "randomize.r", 0))`,
		`ivy.step(gen.x)`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("target=gen range source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		`gen.x = 0`,
		`ivy.r = 0`,
		`using zero value for non-enumerable sort rng`,
		`unsupported action generator parameter`,
	} {
		if strings.Contains(out.Source, bad) || strings.Contains(strings.Join(out.Warnings, "\n"), bad) {
			t.Fatalf("target=gen range sort should use primitive randomization, found %q:\nsource:\n%s\nwarnings:%v", bad, out.Source, out.Warnings)
		}
	}
}

func TestTargetGenActionGeneratorAssumeGuard(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    assume c = green;
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "genguard", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`gen.c = color(ivy.___ivy_randomize(2, "__fml:c", 0))`,
		`if !((gen.c == green)) {`,
		`return false`,
		`if set_generator.generate() {`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("gen assume-guard source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=20", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("run generated Go gen target: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if strings.Contains(stderr, "assumption failed") || strings.Contains(stdout, "assumption_failed") {
		t.Fatalf("generator guard should skip invalid inputs, not execute failed assumes\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if strings.Contains(stdout, "> set(red)") {
		t.Fatalf("generator guard should not execute set(red)\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if strings.Contains(stdout, "test_completed") || strings.Count(stdout, "> set(") > 1 {
		t.Fatalf("gen target should be single-shot without completion marker\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestTargetGenAssignedStateAssumeUsesPreimageFast(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c;
    assume saved = green
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "genpreimage", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Genpreimage_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not found:\n%s", out.Source)
	}
	for _, want := range []string{
		`gen.c = color(ivy.___ivy_randomize(2, "__fml:c", 0))`,
		`if !((gen.c == green)) {`,
		`return false`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("target=gen preimage guard source missing %q:\n%s", want, genBody)
		}
	}
	if strings.Contains(genBody, `ivy.saved == green`) {
		t.Fatalf("target=gen guard should use the assignment preimage, not current state:\n%s", genBody)
	}
}

func TestTargetGenActionGeneratorUsesBeforeExport(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	action, ok := mod.Actions.Get2("set")
	if !ok {
		t.Fatalf("action set not found")
	}
	params := action.GetFormalParams()
	if len(params) != 1 {
		t.Fatalf("set params = %d, want 1", len(params))
	}
	greenEntry, ok := mod.Sig.Symbols.Get2("green")
	if !ok {
		t.Fatalf("symbol green not found")
	}
	paramGreen, err := goivy.NewEq(params[0], goivy.NewConst("green", greenEntry.Sort))
	if err != nil {
		t.Fatalf("param/green NewEq: %v", err)
	}
	actionExpr, ok := action.(goivy.Expr)
	if !ok {
		t.Fatalf("action set does not implement Expr: %T", action)
	}
	before := goivy.NewSequence(goivy.NewAssumeAction(paramGreen), actionExpr)
	before.SetLineno(action.GetLineno())
	goivy.CopyFormalsTo(action, before)
	mod.BeforeExport.Set("set", before)

	out, err := Generate(mod, Config{Target: "gen", ClassName: "beforegen", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Beforegen_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not found:\n%s", out.Source)
	}
	for _, want := range []string{
		`gen.c = color(ivy.___ivy_randomize(2, "__fml:c", 0))`,
		`if !((gen.c == green)) {`,
		`return false`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("before_export guard source missing %q:\n%s", want, genBody)
		}
	}
	if strings.Contains(genBody, `ivy.saved =`) {
		t.Fatalf("before_export analysis action should only drive input generation, not inline execution:\n%s", genBody)
	}
	if !strings.Contains(out.Source, `ivy.set(gen.c)`) {
		t.Fatalf("execute should still call the public action with generated formals:\n%s", out.Source)
	}
}

func TestTargetGenRunnerRandomizesStateOnceAndSkipsUnsat(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    assume c = green;
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "genstate", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	runnerBody := bodyAfterMarker(out.Source, "func ivy2golangGenerate(ivy *genstate)")
	if runnerBody == "" {
		t.Fatalf("gen runner body not found:\n%s", out.Source)
	}
	for _, want := range []string{
		`ivy.saved = color(ivy.___ivy_randomize(2, "randomize.saved", 0))`,
		`init_generator := &Genstate_init_gen{ivy: ivy}`,
		`if set_generator.generate() {`,
		`set_generator.execute()`,
	} {
		if !strings.Contains(runnerBody, want) {
			t.Fatalf("gen state-randomization runner source missing %q:\n%s", want, runnerBody)
		}
	}
	randomizeIdx := strings.Index(runnerBody, `ivy.saved = color(ivy.___ivy_randomize(2, "randomize.saved", 0))`)
	initIdx := strings.Index(runnerBody, `init_generator := &Genstate_init_gen{ivy: ivy}`)
	if randomizeIdx < 0 || initIdx < 0 || randomizeIdx > initIdx {
		t.Fatalf("gen runner should randomize state before init generator; randomize=%d init=%d\n%s", randomizeIdx, initIdx, runnerBody)
	}
	for _, bad := range []string{
		`} else {`,
		`cycle--`,
		`continue`,
	} {
		if strings.Contains(runnerBody, bad) {
			t.Fatalf("gen single-shot runner should skip unsat generators without retry loop %q:\n%s", bad, runnerBody)
		}
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Genstate_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not found:\n%s", out.Source)
	}
	for _, bad := range []string{
		`randomize.saved`,
		`randomize._generating`,
	} {
		if strings.Contains(genBody, bad) {
			t.Fatalf("action generator should not rerandomize model state %q:\n%s", bad, genBody)
		}
	}
	if !strings.Contains(out.Source, `ivy._generating = true`) {
		t.Fatalf("gen execute should still toggle _generating around action calls:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestTargetGenInitGeneratorRunsInitAfterRandomize(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
after init {
    saved := green
}
action observe returns(out:color) = {
    out := saved
}
export observe
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "geninit", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	constructorBody := bodyAfterMarker(out.Source, "func newGeninit() *geninit")
	if constructorBody == "" {
		t.Fatalf("constructor body not found:\n%s", out.Source)
	}
	for _, bad := range []string{
		`ivy.__initState()`,
		`ivy.__init()`,
	} {
		if strings.Contains(constructorBody, bad) {
			t.Fatalf("target=gen constructor should leave initialization to init generator, found %q:\n%s", bad, constructorBody)
		}
	}
	initBody := bodyAfterMarker(out.Source, "func (gen *Geninit_init_gen) generate() bool")
	if initBody == "" {
		t.Fatalf("init generator body not found:\n%s", out.Source)
	}
	for _, want := range []string{
		`ivy := gen.ivy`,
		`ivy.__initState()`,
		`ivy.__init()`,
		`return true`,
	} {
		if !strings.Contains(initBody, want) {
			t.Fatalf("target=gen init generator missing %q:\n%s", want, initBody)
		}
	}
	afterInitBody := bodyAfterMarker(out.Source, "func (ivy *geninit) __init()")
	if !strings.Contains(afterInitBody, `ivy.saved = green`) {
		t.Fatalf("after-init body should set saved to green:\n%s", afterInitBody)
	}
	runnerBody := bodyAfterMarker(out.Source, "func ivy2golangGenerate(ivy *geninit)")
	if runnerBody == "" {
		t.Fatalf("gen runner body not found:\n%s", out.Source)
	}
	randomizeIdx := strings.Index(runnerBody, `ivy.saved = color(ivy.___ivy_randomize(2, "randomize.saved", 0))`)
	initIdx := strings.Index(runnerBody, `_ = init_generator.generate()`)
	if randomizeIdx < 0 || initIdx < 0 || randomizeIdx > initIdx {
		t.Fatalf("gen runner should randomize state before running init generator; randomize=%d init=%d\n%s", randomizeIdx, initIdx, runnerBody)
	}

	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("run generated Go gen target: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "> observe\n= green\n" {
		t.Fatalf("gen init should apply after-init assignments before action generators\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestTargetGenActionGeneratorExtractsDefinedInput(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..7}
individual stored : idx
action set(x:idx) = {
    assume x = 3;
    stored := x
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "defparam", TestIters: "3", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`gen.x = (0 + ivy.___ivy_randomize(8, "__fml:x", 0))`,
		`gen.x = 3`,
		`if !((gen.x == 3)) {`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("gen defined-input source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("run generated Go gen target: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if strings.Contains(stdout, "> set(0)") ||
		strings.Contains(stdout, "> set(1)") ||
		strings.Contains(stdout, "> set(2)") ||
		strings.Contains(stdout, "> set(4)") ||
		strings.Contains(stdout, "> set(5)") ||
		strings.Contains(stdout, "> set(6)") ||
		strings.Contains(stdout, "> set(7)") {
		t.Fatalf("defined input should force all executed calls to set(3)\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if stdout != "> set(3)\n" {
		t.Fatalf("gen target should emit the defined input once without completion marker\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestActionGeneratorsDefineFormalDestructorField(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type cell
destructor shade(C:cell) : color
destructor valid(C:cell) : bool
individual saved : color
action set(c:cell) = {
    assume shade(c) = red;
    saved := shade(c)
}
export set
`)
	genOut, err := Generate(mod, Config{Target: "gen", ClassName: "fieldgen", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate target=gen: %v\n%s", err, outSource(genOut))
	}
	genBody := bodyAfterMarker(genOut.Source, "func (gen *Fieldgen_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not found:\n%s", genOut.Source)
	}
	for _, want := range []string{
		`gen.c = cell{`,
		`gen.c.shade = red`,
		`if !((gen.c.shade == red)) {`,
		`return false`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("target=gen field-defined input missing %q:\n%s", want, genBody)
		}
	}
	assignIdx := strings.Index(genBody, `gen.c.shade = red`)
	guardIdx := strings.Index(genBody, `if !((gen.c.shade == red)) {`)
	if assignIdx < 0 || guardIdx < 0 || assignIdx > guardIdx {
		t.Fatalf("target=gen should define the formal field before checking the guard; assign=%d guard=%d\n%s", assignIdx, guardIdx, genBody)
	}
	if !strings.Contains(genOut.Source, `ivy.set(gen.c)`) {
		t.Fatalf("target=gen execute should pass the updated formal struct:\n%s", genOut.Source)
	}

	testOut, err := Generate(mod, Config{Target: "test", ClassName: "fieldtest", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate target=test: %v\n%s", err, outSource(testOut))
	}
	testGenBody := bodyAfterMarker(testOut.Source, "func (gen *Fieldtest_set_generator) generate() bool")
	for _, want := range []string{
		`gen.c = cell{`,
		`gen.c.shade = red`,
		`if !((gen.c.shade == red)) {`,
	} {
		if !strings.Contains(testGenBody, want) {
			t.Fatalf("target=test field-defined input missing %q:\n%s", want, testGenBody)
		}
	}
	testAssignIdx := strings.Index(testGenBody, `gen.c.shade = red`)
	testGuardIdx := strings.Index(testGenBody, `if !((gen.c.shade == red)) {`)
	if testAssignIdx < 0 || testGuardIdx < 0 || testAssignIdx > testGuardIdx {
		t.Fatalf("target=test should define the formal field before checking the guard; assign=%d guard=%d\n%s", testAssignIdx, testGuardIdx, testGenBody)
	}

	bin := compileGeneratedGo(t, genOut)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("run generated Go gen target: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "> set({shade:red,") || strings.Contains(stdout, "test_completed") {
		t.Fatalf("gen target should execute the field-constrained action once without completion marker\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestDestructorMultiArgFiniteFieldDeclarationAndAccess(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green, blue}
type idx = {0..3}
type cell
destructor shade(C:cell, I:idx) : color
individual a : cell
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	idx, ok := mod.Sig.Sorts.Get2("idx")
	if !ok {
		t.Fatal("missing idx sort")
	}
	cell, ok := mod.Sig.Sorts.Get2("cell")
	if !ok {
		t.Fatal("missing cell sort")
	}
	shadeSort, err := goivy.NewFunctionSort(cell, idx, color)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	shade := goivy.NewConst("shade", shadeSort)
	lhs, err := goivy.NewApply(shade, goivy.NewConst("a", cell), goivy.NewConst("0", idx))
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	mod.Actions.Set("step", goivy.NewSequence(goivy.NewAssignAction(lhs, goivy.NewConst("red", color))))
	mod.PublicActions.Set("step", true)

	out, err := Generate(mod, Config{Target: "test", ClassName: "multiarg_field", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type cell struct {",
		"shade [4]color",
		"ivy.a.shade[0] = red",
		"func (v cell) String() string",
		`__b.WriteString("[")`,
		`__b.WriteString(",")`,
		`__b.WriteString(fmt.Sprintf("%v", v.shade[__i0]))`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("multi-arg destructor source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "shade(ivy.a, 0)") || strings.Contains(out.Source, "shade(a, 0)") {
		t.Fatalf("multi-arg destructor leaked as a function call:\n%s", out.Source)
	}
	if len(out.Warnings) != 0 {
		t.Fatalf("multi-arg destructor should not fall back to non-enumerable warnings, got %v", out.Warnings)
	}
	compileGeneratedGo(t, out)
}

func TestDestructorHashThunkFieldDeclarationAndAccess(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type key
type cell
destructor shade(C:cell, K:key) : color
individual a : cell
action paint(k:key) = {
    shade(a,k) := green;
    assert shade(a,k) = green
}
export paint
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "thunkfield", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type cell struct {",
		"shade *ivyThunkMap[int, color]",
		`__b.WriteString("<hash_thunk>")`,
		"ivyThunkSet(&ivy.a.shade, k, green, red)",
		"ivyThunkGet(ivy.a.shade, k, red) == green",
		"newIvyThunkMap[int, color](red)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("hash-thunk destructor source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "shade(ivy.a,") || strings.Contains(out.Source, "shade(a,") {
		t.Fatalf("hash-thunk destructor leaked as a function call:\n%s", out.Source)
	}
	for _, warning := range out.Warnings {
		if strings.Contains(warning, "non-enumerable sort cell") {
			t.Fatalf("hash-thunk destructor should not fall back to record zero value, got warnings %v", out.Warnings)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run hash-thunk destructor generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("hash-thunk destructor run missing completion marker\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestActionGeneratorsDefineFormalHashThunkDestructorField(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type key
type cell
destructor shade(C:cell, K:key) : color
action set(c:cell, k:key) = {
    assume shade(c,k) = green
}
export set
`)
	genOut, err := Generate(mod, Config{Target: "gen", ClassName: "thunkfieldgen", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate target=gen: %v\n%s", err, outSource(genOut))
	}
	genBody := bodyAfterMarker(genOut.Source, "func (gen *Thunkfieldgen_set_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("set generator body not found:\n%s", genOut.Source)
	}
	for _, want := range []string{
		`ivyThunkSet(&gen.c.shade, gen.k, green, red)`,
		`if !((ivyThunkGet(gen.c.shade, gen.k, red) == green)) {`,
	} {
		if !strings.Contains(genBody, want) {
			t.Fatalf("target=gen hash-thunk field-defined input missing %q:\n%s", want, genBody)
		}
	}
	if strings.Contains(genBody, `ivyThunkGet(gen.c.shade, gen.k, red) = green`) {
		t.Fatalf("target=gen hash-thunk field-defined input assigned to getter:\n%s", genBody)
	}
	assignIdx := strings.Index(genBody, `ivyThunkSet(&gen.c.shade, gen.k, green, red)`)
	guardIdx := strings.Index(genBody, `if !((ivyThunkGet(gen.c.shade, gen.k, red) == green)) {`)
	if assignIdx < 0 || guardIdx < 0 || assignIdx > guardIdx {
		t.Fatalf("target=gen should define the hash-thunk field before checking the guard; assign=%d guard=%d\n%s", assignIdx, guardIdx, genBody)
	}

	testOut, err := Generate(mod, Config{Target: "test", ClassName: "thunkfieldtest", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate target=test: %v\n%s", err, outSource(testOut))
	}
	testGenBody := bodyAfterMarker(testOut.Source, "func (gen *Thunkfieldtest_set_generator) generate() bool")
	for _, want := range []string{
		`ivyThunkSet(&gen.c.shade, gen.k, green, red)`,
		`if !((ivyThunkGet(gen.c.shade, gen.k, red) == green)) {`,
	} {
		if !strings.Contains(testGenBody, want) {
			t.Fatalf("target=test hash-thunk field-defined input missing %q:\n%s", want, testGenBody)
		}
	}
	if strings.Contains(testGenBody, `ivyThunkGet(gen.c.shade, gen.k, red) = green`) {
		t.Fatalf("target=test hash-thunk field-defined input assigned to getter:\n%s", testGenBody)
	}
	testAssignIdx := strings.Index(testGenBody, `ivyThunkSet(&gen.c.shade, gen.k, green, red)`)
	testGuardIdx := strings.Index(testGenBody, `if !((ivyThunkGet(gen.c.shade, gen.k, red) == green)) {`)
	if testAssignIdx < 0 || testGuardIdx < 0 || testAssignIdx > testGuardIdx {
		t.Fatalf("target=test should define the hash-thunk field before checking the guard; assign=%d guard=%d\n%s", testAssignIdx, testGuardIdx, testGenBody)
	}

	bin := compileGeneratedGo(t, genOut)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("run generated Go gen target: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "> set({shade:<hash_thunk>},0)") || strings.Contains(stdout, "test_completed") {
		t.Fatalf("gen target should execute the hash-thunk-constrained action once without completion marker\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestDestructorHashThunkFieldsAreSkippedInRecordEquality(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type key
type cell
destructor shade(C:cell, K:key) : color
action step = {
    var c1 : cell;
    var c2 : cell;
    var k : key;
    shade(c1,k) := green;
    shade(c2,k) := red;
    assert c1 = c2
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "thunkeq"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func (v cell) Equal(other cell) bool",
		"return true",
		".Equal(loc__c2)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("hash-thunk equality source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "loc__c1 == loc__c2") {
		t.Fatalf("hash-thunk record equality should not use Go struct ==:\n%s", out.Source)
	}

	bin := compileGeneratedGo(t, out)
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte("step\n"), 0o644); err != nil {
		t.Fatalf("write step command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err != nil {
		t.Fatalf("run generated hash-thunk equality repl: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "> step\n" {
		t.Fatalf("hash-thunk equality repl output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestDestructorNestedHashThunkFieldsUseNestedRecordEquality(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type key
type cell
type wrapper
destructor shade(C:cell, K:key) : color
destructor child(W:wrapper) : cell
action step = {
    var w1 : wrapper;
    var w2 : wrapper;
    assert w1 = w2
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "nestedthunkeq"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func (v cell) Equal(other cell) bool",
		"func (v wrapper) Equal(other wrapper) bool",
		"(v.child).Equal(other.child)",
		".Equal(loc__w2)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("nested hash-thunk equality source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		"v.child == other.child",
		"loc__w1 == loc__w2",
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("nested hash-thunk equality should not use Go struct == %q:\n%s", bad, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestTargetGenActionGeneratorUsesExtPreconds(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	savedEntry, ok := mod.Sig.Symbols.Get2("saved")
	if !ok {
		t.Fatalf("symbol saved not found")
	}
	redEntry, ok := mod.Sig.Symbols.Get2("red")
	if !ok {
		t.Fatalf("symbol red not found")
	}
	action, ok := mod.Actions.Get2("set")
	if !ok {
		t.Fatalf("action set not found")
	}
	params := action.GetFormalParams()
	if len(params) != 1 {
		t.Fatalf("set params = %d, want 1", len(params))
	}
	savedRed, err := goivy.NewEq(
		goivy.NewConst("saved", savedEntry.Sort),
		goivy.NewConst("red", redEntry.Sort),
	)
	if err != nil {
		t.Fatalf("saved/red NewEq: %v", err)
	}
	paramRed, err := goivy.NewEq(params[0], goivy.NewConst("red", redEntry.Sort))
	if err != nil {
		t.Fatalf("param/red NewEq: %v", err)
	}
	pre, err := goivy.NewAnd(savedRed, paramRed)
	if err != nil {
		t.Fatalf("NewAnd: %v", err)
	}
	mod.ExtPreconds["set"] = pre

	out, err := Generate(mod, Config{Target: "gen", ClassName: "extpregen", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`ivy.saved = color(ivy.___ivy_randomize(2, "randomize.saved", 0))`,
		`gen.c = color(ivy.___ivy_randomize(2, "__fml:c", 0))`,
		`if !((((ivy.saved == red)) && ((gen.c == red)))) {`,
		`return false`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("gen ext-precondition source missing %q:\n%s", want, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestTargetGenExtPreconditionPreservesReturningActionFormals(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) returns (out:color) = {
    saved := c;
    out := c
}
export set
`)
	savedEntry, ok := mod.Sig.Symbols.Get2("saved")
	if !ok {
		t.Fatalf("symbol saved not found")
	}
	redEntry, ok := mod.Sig.Symbols.Get2("red")
	if !ok {
		t.Fatalf("symbol red not found")
	}
	action, ok := mod.Actions.Get2("set")
	if !ok {
		t.Fatalf("action set not found")
	}
	params := action.GetFormalParams()
	if len(params) != 1 {
		t.Fatalf("set params = %d, want 1", len(params))
	}
	savedRed, err := goivy.NewEq(
		goivy.NewConst("saved", savedEntry.Sort),
		goivy.NewConst("red", redEntry.Sort),
	)
	if err != nil {
		t.Fatalf("saved/red NewEq: %v", err)
	}
	paramRed, err := goivy.NewEq(params[0], goivy.NewConst("red", redEntry.Sort))
	if err != nil {
		t.Fatalf("param/red NewEq: %v", err)
	}
	pre, err := goivy.NewAnd(savedRed, paramRed)
	if err != nil {
		t.Fatalf("NewAnd: %v", err)
	}
	mod.ExtPreconds["set"] = pre

	out, err := Generate(mod, Config{Target: "gen", ClassName: "extretgen", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`type Extretgen_set_generator struct {`,
		`c color`,
		`gen.c = color(ivy.___ivy_randomize(2, "__fml:c", 0))`,
		`if !((((ivy.saved == red)) && ((gen.c == red)))) {`,
		`return false`,
		`__res := ivy.set(gen.c)`,
		`fmt.Fprintf(__ivy_out, "= %s\n", ivyTraceValue(__res, false))`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("gen ext-precondition returning source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		`ivy.set()`,
		`ivy.set(__arg0`,
		`unsupported action generator parameter`,
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("gen ext-precondition returning action lost formals or fell back to %q:\n%s", bad, out.Source)
		}
	}
}

func TestTargetGenActionGeneratorsFollowActionOrder(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual saved : bool
action beta = {
    saved := true
}
export beta
action alpha = {
    saved := false
}
export alpha
action gamma = {
    saved := true
}
export gamma
action internal = {
    saved := false
}
`)
	mod.PublicActions.Set("internal", false)
	var first string
	for i := 0; i < 20; i++ {
		out, err := Generate(mod, Config{Target: "gen", ClassName: "genorder"})
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if i == 0 {
			first = out.Source
		} else if out.Source != first {
			t.Fatalf("target=gen action generator output should be stable across runs\nfirst:\n%s\nrun %d:\n%s", first, i, out.Source)
		}
		beta := strings.Index(out.Source, "type Genorder_beta_generator struct")
		alpha := strings.Index(out.Source, "type Genorder_alpha_generator struct")
		gamma := strings.Index(out.Source, "type Genorder_gamma_generator struct")
		if beta < 0 || alpha < 0 || gamma < 0 {
			t.Fatalf("missing action generators in source:\n%s", out.Source)
		}
		if !(beta < alpha && alpha < gamma) {
			t.Fatalf("target=gen action generators should follow action order beta, alpha, gamma; positions beta=%d alpha=%d gamma=%d\n%s", beta, alpha, gamma, out.Source)
		}
		for _, bad := range []string{
			"type Genorder_internal_generator struct",
			"internal_generator := &Genorder_internal_generator",
			"ivy.internal()",
		} {
			if strings.Contains(out.Source, bad) {
				t.Fatalf("target=gen should ignore explicitly private actions, found %q:\n%s", bad, out.Source)
			}
		}
	}
}

func TestTargetGenTraceWrapsActionGeneratorExecute(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action echo(c:color) returns(out:color) = {
    out := c
}
export echo
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "gentrace", Trace: true, TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (gen *Gentrace_echo_generator) execute()")
	if body == "" {
		t.Fatalf("echo generator execute body not found:\n%s", out.Source)
	}
	for _, want := range []string{
		`fmt.Fprintf(__ivy_out, "> echo(%s)\n", ivyTraceValue(gen.c, false))`,
		`fmt.Fprintln(__ivy_out, "{")`,
		`__res := ivy.echo(gen.c)`,
		`fmt.Fprintln(__ivy_out, "}")`,
		`fmt.Fprintf(__ivy_out, "= %s\n", ivyTraceValue(__res, false))`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("traced gen execute body missing %q:\n%s", want, body)
		}
	}
	compileGeneratedGo(t, out)

	out2, err := Generate(mod, Config{Target: "gen", ClassName: "gentrace2", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate without trace: %v\n%s", err, outSource(out2))
	}
	body2 := bodyAfterMarker(out2.Source, "func (gen *Gentrace2_echo_generator) execute()")
	if body2 == "" {
		t.Fatalf("echo generator execute body not found without trace:\n%s", out2.Source)
	}
	if strings.Contains(body2, `fmt.Fprintln(__ivy_out, "{")`) ||
		strings.Contains(body2, `fmt.Fprintln(__ivy_out, "}")`) ||
		strings.Contains(body2, `__res := ivy.echo(gen.c)`) {
		t.Fatalf("untraced gen execute body should not emit trace braces/capture:\n%s", body2)
	}
}

func TestCompileAndGenerateDefaultTargetGenUsesSingleShotMain(t *testing.T) {
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
		`func ivy2golangGenerate(ivy *defaultgen) {`,
		`init_generator := &Defaultgen_init_gen{ivy: ivy}`,
		`step_generator := &Defaultgen_step_generator{ivy: ivy}`,
		`if step_generator.generate() {`,
		`step_generator.execute()`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("default gen source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		`for cycle := 0; cycle < testIters; cycle++ {`,
		`fmt.Fprintln(__ivy_out, "test_completed")`,
		`for runidx := 0; runidx < runs; runidx++ {`,
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("target=gen should be single-shot like ivy2cpp, found %q:\n%s", bad, out.Source)
		}
	}
	for _, want := range []string{
		`opts, rest, optArgs := parseIvyTestArgs(os.Args[1:])`,
		`__ivy_out_opened := false`,
		`case "out":`,
		`__ivy_out = __ivy_out_file`,
		`case "modelfile":`,
		`__ivy_modelfile_file, err := os.Create(__ivy_opt.value)`,
		`fmt.Fprintln(__ivy_modelfile_file, "ivy2golang: modelfile solver logging is not implemented for generated Go")`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("target=gen runtime option source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run default gen output: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "> step\n" {
		t.Fatalf("target=gen run output differs from ivy2cpp single-shot form\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	traceFile := filepath.Join(t.TempDir(), "gen_trace.out")
	modelFile := filepath.Join(t.TempDir(), "gen_model.out")
	stdout, stderr, err = runBinary(t, bin, "iters=1", "runs=1", "seed=1", "out="+traceFile, "modelfile="+modelFile)
	if err != nil {
		t.Fatalf("run default gen output with runtime files: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "" {
		t.Fatalf("target=gen out= should redirect trace output away from stdout, got %q", stdout)
	}
	gotTrace, err := os.ReadFile(traceFile)
	if err != nil {
		t.Fatalf("out= trace file missing: %v", err)
	}
	if string(gotTrace) != "> step\n" {
		t.Fatalf("target=gen out= trace differs: %q", gotTrace)
	}
	if _, err := os.Stat(modelFile); err != nil {
		t.Fatalf("target=gen modelfile option did not create file %s: %v", modelFile, err)
	}
	modelData, err := os.ReadFile(modelFile)
	if err != nil {
		t.Fatalf("read target=gen modelfile %s: %v", modelFile, err)
	}
	if !strings.Contains(string(modelData), "modelfile solver logging is not implemented") {
		t.Fatalf("target=gen modelfile should contain unsupported solver-log marker, got %q", modelData)
	}
}

func TestTargetGenSeedPlumbingPrecedesRandomCalls(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual c : color
action pick(x:color) = {
    c := x
}
export pick
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "seeded", EmitMain: true, TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("missing generated main:\n%s", out.Source)
	}
	declIdx := strings.Index(mainBody, `seed := ivyAtoiOption(opts, "seed", 1)`)
	setSeedIdx := strings.Index(mainBody, "ivySetSeed(seed)")
	newIdx := strings.Index(mainBody, "ivy := newSeeded()")
	genIdx := strings.Index(mainBody, "ivy2golangGenerate(ivy)")
	if declIdx < 0 || setSeedIdx < 0 || newIdx < 0 || genIdx < 0 {
		t.Fatalf("seed/gen plumbing pieces missing (decl=%d setSeed=%d new=%d gen=%d):\n%s", declIdx, setSeedIdx, newIdx, genIdx, mainBody)
	}
	if !(declIdx < setSeedIdx && setSeedIdx < newIdx && newIdx < genIdx) {
		t.Fatalf("seed/gen plumbing out of order (decl=%d setSeed=%d new=%d gen=%d):\n%s", declIdx, setSeedIdx, newIdx, genIdx, mainBody)
	}
	generateBody := bodyAfterMarker(out.Source, "func ivy2golangGenerate(ivy *seeded)")
	if generateBody == "" || !strings.Contains(generateBody, `ivy.___ivy_randomize`) {
		t.Fatalf("single-shot generator should contain randomized generation calls:\n%s", out.Source)
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
	setBody := bodyAfterMarker(out.Source, "func (ivy *basic_assign) set()")
	if strings.Contains(setBody, "__ivy_tmp") {
		t.Fatalf("simple scalar assignment should not allocate a temporary:\n%s", setBody)
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

func TestBeforeAfterMixinBodiesEmittedInAction(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mixcase.ivy")
	if err := os.WriteFile(path, []byte(`#lang ivy1.7
individual flag : bool
action step returns(out:bool) = {
    out := flag
}
before step {
    flag := true
}
after step {
    flag := false
}
export step
`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	batch, err := CompileAndGenerateAll(path, map[string]string{"target": "test", "classname": "mixcase"}, Config{TestIters: "1"})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	if len(batch.Outputs) != 1 {
		t.Fatalf("outputs=%d, want 1", len(batch.Outputs))
	}
	out := batch.Outputs[0]
	body := bodyAfterMarker(out.Source, "func (ivy *mixcase) ext__step() bool")
	if body == "" {
		t.Fatalf("ext:step method not emitted:\n%s", out.Source)
	}
	beforeIdx := strings.Index(body, "ivy.flag = true")
	stepIdx := strings.Index(body, "out = ivy.flag")
	afterIdx := strings.Index(body, "ivy.flag = false")
	if beforeIdx < 0 || stepIdx < 0 || afterIdx < 0 || beforeIdx > stepIdx || stepIdx > afterIdx {
		t.Fatalf("ext:step body should include before mixin, base body, then after mixin (before=%d step=%d after=%d):\n%s", beforeIdx, stepIdx, afterIdx, body)
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

func TestGeneratedEnvActionUsesIfElseChain(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual flag : bool
action step = {
}
export step
`)
	assign := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("true", goivy.Boolean))
	clear := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("false", goivy.Boolean))
	mod.Actions.Set("step", goivy.NewEnvActionOn(goivy.NewActionsConfig(), assign, clear))
	out, err := Generate(mod, Config{Target: "test", ClassName: "envcase", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *envcase) step()")
	if body == "" {
		t.Fatalf("step method not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		"__ivy_branch",
		`ivy.___ivy_choose(0, "___branch"`,
		"if __ivy_branch",
		"} else {",
		"ivy.flag = true",
		"ivy.flag = false",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("env action source missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "switch ivy.___ivy_choose") {
		t.Fatalf("env action should use an if/else chain, not a switch:\n%s", body)
	}
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

func TestGeneratedStructLocalNondetUsesScalarFieldChoices(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
type byte
interpret byte -> bv[8]
type cell
destructor shade(C:cell) : color
destructor bits(C:cell) : byte
individual saved_color : color
individual saved_bits : byte
action step = {
    var c : cell;
    saved_color := c.shade;
    saved_bits := c.bits
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "struct_local", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`loc__c := cell{shade: color(ivy.___ivy_choose(0, "loc:c.shade", 0)), bits: ivy.___ivy_choose(0, "loc:c.bits", 1)}`,
		"ivy.saved_color = loc__c.shade",
		"ivy.saved_bits = loc__c.bits",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("struct local source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		`color(ivy.___ivy_choose(2, "loc:c.shade"`,
		`ivy.___ivy_choose(256, "loc:c.bits"`,
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("struct local fields should use scalar nondet choose(0), found %q:\n%s", bad, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestGeneratedHelperClassLocalNondetUsesDefaultValue(t *testing.T) {
	src := `#lang ivy1.7
type text
interpret text -> strbv[4]
type small
interpret small -> intbv[10][20][4]
individual saved_text : text
individual saved_small : small
action step = {
    var t : text;
    var n : small;
    saved_text := t;
    saved_small := n
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "helper_local", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`loc__t := ""`,
		`loc__n := 0`,
		`ivy.saved_text = loc__t`,
		`ivy.saved_small = loc__n`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("helper-class local source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		`strconv.Itoa(ivy.___ivy_choose(16, "loc:t"`,
		`ivy.___ivy_choose(0, "loc:n"`,
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("helper-class local should use default value, found %q:\n%s", bad, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestVariantSuperLocalNondetMatchesIvy2CppZeroShape(t *testing.T) {
	src := `#lang ivy1.7
type t
variant a of t
variant b of t
variant c of t
action step = {
    var choice : t
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "variant_local", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"loc__choice := t{}",
		"_ = loc__choice",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("variant local source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		`switch ivy.___ivy_choose(3, "loc:choice"`,
		`switch ivy.___ivy_randomize(3, "loc:choice"`,
		`return t{tag: 0, value: a{}, valid: true}`,
		`return t{tag: 1, value: b{}, valid: true}`,
		`return t{tag: 2, value: c{}, valid: true}`,
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("variant super local nondet should match ivy2cpp default/zero shape, found %q:\n%s", bad, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestIgnoredLocalDeclarationMarkedUsed(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
action step = {
    var res : color;
    res := green
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "ignored_local", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"loc__res := color(ivy.___ivy_choose(0, \"loc:res\"",
		"_ = loc__res",
		"loc__res = green",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("ignored local source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestGeneratedIgnoredLocalDeclarationCompilesAndRuns(t *testing.T) {
	requireSlowTest(t)
	src := `#lang ivy1.7
type color = {red, green}
action step = {
    var res : color;
    res := green
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "ignored_local_run", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated ignored local target=test: %v\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", err, stdout, stderr, out.Source)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
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

func TestGeneratedLocalFunctionSmallBitvectorDomainUsesNondetLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type word
interpret word -> bv[8]
individual saved : bool
action step = {
}
export step
`)
	word, ok := mod.Sig.Sorts.Get2("word")
	if !ok {
		t.Fatal("missing word sort")
	}
	relSort, err := goivy.NewFunctionSort(word, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	marked := goivy.NewConst("marked", relSort)
	zero := goivy.NewConst("0", word)
	body := goivy.NewAssignAction(goivy.NewConst("saved", goivy.Boolean), goivy.MustApply(marked, zero))
	mod.Actions.Set("step", goivy.NewLocalActionOn(goivy.NewActionsConfig(), "test", marked, body))
	out, err := Generate(mod, Config{Target: "test", ClassName: "locfuncbv8", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"marked := make(map[int]bool)",
		"for __i0 := 0; __i0 < 256; __i0++ {",
		`marked[__i0] = ivy.___ivy_choose(0, "marked", 0) != 0`,
		"ivy.saved = marked[0]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("local bv8 function source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		"marked := newIvyThunkMap[int, bool]",
		"marked.base = func",
		"marked.Get(",
		`ivy.___ivy_choose(256, "marked"`,
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("bv8 local function should eagerly enumerate scalar nondet cells, found %q:\n%s", bad, out.Source)
		}
	}
}

func TestGeneratedLocalFunctionHelperClassRangeUsesDefaultCells(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..1}
type text
interpret text -> strbv[4]
individual saved : text
action step = {
}
export step
`)
	idx, ok := mod.Sig.Sorts.Get2("idx")
	if !ok {
		t.Fatal("missing idx sort")
	}
	text, ok := mod.Sig.Sorts.Get2("text")
	if !ok {
		t.Fatal("missing text sort")
	}
	fnSort, err := goivy.NewFunctionSort(idx, text)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	table := goivy.NewConst("table", fnSort)
	zero := goivy.NewConst("0", idx)
	body := goivy.NewAssignAction(goivy.NewConst("saved", text), goivy.MustApply(table, zero))
	mod.Actions.Set("step", goivy.NewLocalActionOn(goivy.NewActionsConfig(), "test", table, body))
	out, err := Generate(mod, Config{Target: "test", ClassName: "locfunchelper", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"table := make(map[int]string)",
		"_ = table",
		"ivy.saved = table[0]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("local function helper-range source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		`table[__i0] =`,
		`strconv.Itoa(ivy.___ivy_choose(16, "table"`,
		`ivy.___ivy_choose(0, "table"`,
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("helper-class local function range should use default cells, found %q:\n%s", bad, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestGeneratedLocalFunctionBitvectorDomainUsesNondetThunk(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type word
interpret word -> bv[32]
individual saved : bool
action step = {
}
export step
`)
	word, ok := mod.Sig.Sorts.Get2("word")
	if !ok {
		t.Fatal("missing word sort")
	}
	relSort, err := goivy.NewFunctionSort(word, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	marked := goivy.NewConst("marked", relSort)
	zero := goivy.NewConst("0", word)
	body := goivy.NewAssignAction(goivy.NewConst("saved", goivy.Boolean), goivy.MustApply(marked, zero))
	mod.Actions.Set("step", goivy.NewLocalActionOn(goivy.NewActionsConfig(), "test", marked, body))
	out, err := Generate(mod, Config{Target: "test", ClassName: "locfuncbv32", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"marked := newIvyThunkMap[int, bool](false)",
		"marked.base = func(__ivy_key int) bool {",
		`return ivy.___ivy_choose(0, "marked", 0) != 0`,
		"ivy.saved = marked.Get(0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("local bv32 function source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		"marked := make(map[int]bool)",
		"for __i0 := 0; __i0 < 4294967296; __i0++ {",
		"range []int{0, 1, 2",
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("bv32 local function should use lazy thunk storage, found %q:\n%s", bad, out.Source)
		}
	}
}

func TestLargeFiniteFunctionStateUsesThunkMap(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..32}
relation big(X:idx,Y:idx)
action step = {
    assert big(0,0) -> big(0,0)
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "largefinite", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"big ivyThunkMap[struct{ A0 int; A1 int }, bool]",
		"ivy.big = newIvyThunkMap[struct{ A0 int; A1 int }, bool](false)",
		"ivy.big.base = func(__ivy_key struct{ A0 int; A1 int }) bool {",
		`return ivy.___ivy_choose(0, "init", 0) != 0`,
		"ivy.big.Get(struct{ A0 int; A1 int }{0, 0})",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("large finite function state source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		"big map[struct{ A0 int; A1 int }]bool",
		"make(map[struct{ A0 int; A1 int }]bool)",
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("large finite function state should be thunk-backed, found %q:\n%s", bad, out.Source)
		}
	}
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

func TestQuantifiedAssignSkipsGuardBoundsWhenConditionReadsModifiedSymbol(t *testing.T) {
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
	ltLimit, err := goivy.NewApply(goivy.NewConst("<", ltSort), iVar, goivy.NewConst("limit", idx))
	if err != nil {
		t.Fatalf("NewApply <: %v", err)
	}
	cond, err := goivy.NewAnd(markedApp, ltLimit)
	if err != nil {
		t.Fatalf("NewAnd: %v", err)
	}
	ite, err := goivy.NewIte(cond, goivy.NewConst("true", goivy.Boolean), markedApp)
	if err != nil {
		t.Fatalf("NewIte: %v", err)
	}
	mod.Actions.Set("step", goivy.NewAssignAction(markedApp, ite))
	out, err := Generate(mod, Config{Target: "test", ClassName: "guard_bounds_skipped", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *guard_bounds_skipped) step()")
	if body == "" {
		t.Fatalf("step body not emitted:\n%s", out.Source)
	}
	if got := strings.Count(body, "for I := 0; I < (5)+1; I++ {"); got != 2 {
		t.Fatalf("mutable-condition assignment should use full default range twice, got %d:\n%s", got, body)
	}
	if strings.Contains(body, "for I := 0; I < ivy.limit; I++") {
		t.Fatalf("condition that reads modified symbol must not tighten assignment loop bounds:\n%s", body)
	}
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

func TestNondetThunkRecordRangeInitializesFields(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type key
type idx = {0..3}
type endpoint
destructor addr(E:endpoint): idx
destructor port(E:endpoint): idx
function src(K:key): endpoint
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "nondetrec"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *nondetrec) __initState()")
	if body == "" {
		t.Fatalf("__initState not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`ivy.src.base = func(__ivy_key int) endpoint {`,
		`return endpoint{addr: (0 + ivy.___ivy_choose(4, "init.addr", 0)), port: (0 + ivy.___ivy_choose(4, "init.port", 1))}`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("record-range nondet thunk missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "endpoint(ivy.___ivy_choose") {
		t.Fatalf("record-range nondet thunk should initialize fields instead of casting int to struct:\n%s", body)
	}
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

func TestObjectAliasLargeFunctionAssignmentUsesThunkSet(t *testing.T) {
	src := `#lang ivy1.7
object payload = {
    type t = struct {
        ok : bool
    }
}
object ref = {
    type txid
    function txs(T:txid) : payload.t
    function txres(T:txid) : payload.t
    action exec(inp:payload.t) returns (out:payload.t) = {
        out := inp
    }
    action commit(tx:txid) = {
        txres(tx) := exec(txs(tx))
    }
    interpret txid -> int
}
export ref.commit
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "object_alias_set", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"ref__txres ivyThunkMap[int, payload__t]",
		`ivy.___ivy_push("ref.exec")`,
		"__ivy_ret0 := ivy.ref__exec(ivy.ref__txs.Get(tx))",
		"ivy.___ivy_pop()",
		"ivy.ref__txres.Set(tx, __ivy_ret0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("object alias thunk assignment source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "ivy.ref__txres.Get(tx) = ") {
		t.Fatalf("object-local large function assignment must write with Set, not assign to Get:\n%s", out.Source)
	}
}

func TestGeneratedObjectAliasLargeFunctionAssignmentCompilesAndRuns(t *testing.T) {
	requireSlowTest(t)
	src := `#lang ivy1.7
object payload = {
    type t = struct {
        ok : bool
    }
}
object ref = {
    type txid
    function txs(T:txid) : payload.t
    function txres(T:txid) : payload.t
    action exec(inp:payload.t) returns (out:payload.t) = {
        out := inp
    }
    action commit(tx:txid) = {
        txres(tx) := exec(txs(tx))
    }
    interpret txid -> int
}
export ref.commit
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "object_alias_set_run", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated object alias assignment target=test: %v\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", err, stdout, stderr, out.Source)
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
		"gen.c = color(ivy.___ivy_randomize",
		"__arg0 := choose_generator.c",
		`__ivy_result := ivy.choose(__arg0)`,
		`fmt.Fprintf(__ivy_out, "= %s\n", ivyTraceValue(__ivy_result, false))`,
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

func TestPrivateMultipleReturnActionCallEmitsTupleAssignment(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
individual ok : bool
action split(c:color) returns (out:color, good:bool) = {
    out := c;
    good := true
}
action step = {
    call saved, ok := split(green)
}
export step
`)
	mod.PublicActions.Set("split", false)
	out, err := Generate(mod, Config{Target: "test", ClassName: "multi_private", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func (ivy *multi_private) split(c color) (color, bool)",
		"out := red",
		"good := false",
		"out = c",
		"good = true",
		"return out, good",
		`ivy.___ivy_push("split")`,
		"__ivy_ret0, __ivy_ret1 := ivy.split(green)",
		"ivy.___ivy_pop()",
		"ivy.saved = __ivy_ret0",
		"ivy.ok = __ivy_ret1",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("private multi-return source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `fmt.Fprintln(__ivy_out, "> split")`) {
		t.Fatalf("private split should not be directly randomized as a public action:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestPrivateMultipleReturnActionCallUsesTempsForThunkReturnStores(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation marked(N:node)
action split returns (a:bool, b:bool) = {
    a := true;
    b := false
}
action step = {
    var x : node;
    var y : node;
    call marked(x), marked(y) := split
}
export step
`)
	mod.PublicActions.Set("split", false)
	out, err := Generate(mod, Config{Target: "test", ClassName: "multi_thunk_returns", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func (ivy *multi_thunk_returns) split() (bool, bool)",
		"__ivy_ret0, __ivy_ret1 := ivy.split()",
		"ivy.marked.Set(loc__x, __ivy_ret0)",
		"ivy.marked.Set(loc__y, __ivy_ret1)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("private multi-return thunk store source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "ivy.marked.Set(loc__x, ivy.split())") {
		t.Fatalf("thunk return stores must not inline the multi-return call:\n%s", out.Source)
	}
	assertNoUnsupportedGo(t, out.Source)
}

func TestReturnNameAliasingInputDoesNotRedeclareFormal(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..5}
individual saved : idx
action same(x:idx) returns (x:idx) = {
    x := x
}
action step = {
    call saved := same(saved)
}
export step
`)
	mod.PublicActions.Set("same", false)
	out, err := Generate(mod, Config{Target: "test", ClassName: "alias_return", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *alias_return) same(")
	if body == "" {
		t.Fatalf("same body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		"func (ivy *alias_return) same(x int) int",
		"x = x",
		"return x",
		`ivy.___ivy_push("same")`,
		"__ivy_ret0 := ivy.same(ivy.saved)",
		"ivy.___ivy_pop()",
		"ivy.saved = __ivy_ret0",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("return/input alias source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(body, "x := 0") {
		t.Fatalf("return/input alias should reuse the formal, not redeclare it:\n%s", body)
	}
}

func TestCallReturnAliasInputAndOutputUsesValueReturn(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..5}
action bump(x:idx) returns (x:idx) = {
    x := x + 1
}
action step = {
    var a : idx;
    var b : idx;
    call a := bump(b)
}
export step
`)
	mod.PublicActions.Set("bump", false)
	out, err := Generate(mod, Config{Target: "test", ClassName: "alias_call", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func (ivy *alias_call) bump(x int) int",
		"__x := (x) + (1)",
		"return x",
		`ivy.___ivy_push("bump")`,
		"__ivy_ret0 := ivy.bump(loc__b)",
		"ivy.___ivy_pop()",
		"loc__a = __ivy_ret0",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("aliasing call source missing %q:\n%s", want, out.Source)
		}
	}
	stepBody := bodyAfterMarker(out.Source, "func (ivy *alias_call) step()")
	if stepBody == "" {
		t.Fatalf("step body not emitted:\n%s", out.Source)
	}
	for _, bad := range []string{
		"loc__b = ivy.bump(",
		"__tmp",
	} {
		if strings.Contains(stepBody, bad) {
			t.Fatalf("Go aliasing call should keep the input as an argument and assign only the output; found %q:\n%s", bad, stepBody)
		}
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	if strings.Contains(mainBody, `> bump`) {
		t.Fatalf("private return-alias helper should not be randomized as a public test action:\n%s", mainBody)
	}
}

func TestGeneratedReturnNameAliasingInputCompilesAndRuns(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..5}
individual saved : idx
action same(x:idx) returns (x:idx) = {
    x := x
}
action step = {
    call saved := same(saved)
}
export step
`)
	mod.PublicActions.Set("same", false)
	out, err := Generate(mod, Config{Target: "test", ClassName: "alias_return_run", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated return/input alias target=test: %v\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", err, stdout, stderr, out.Source)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
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

func TestNumberFormatHexFromRadixAttribute(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..20}
action step(i:idx) returns(out:idx) = {
    out := i
}
export step
`)
	mod.SetAttribute("radix", "16")
	out, err := Generate(mod, Config{Target: "test", ClassName: "hexfmt", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`fmt.Fprintf(__ivy_out, "> step(%s)\n", ivyTraceValue(__arg0, true))`,
		`__ivy_result := ivy.step(__arg0)`,
		`fmt.Fprintf(__ivy_out, "= %s\n", ivyTraceValue(__ivy_result, true))`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("radix=16 action/result trace source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestNumberFormatHexFromRadixRelnameAttribute(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..20}
action step(i:idx) = {
}
export step
`)
	mod.SetAttribute("radix", testRadixRelname{rep: "16"})
	out, err := Generate(mod, Config{Target: "test", ClassName: "hexrelname", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	want := `fmt.Fprintf(__ivy_out, "> step(%s)\n", ivyTraceValue(__arg0, true))`
	if !strings.Contains(out.Source, want) {
		t.Fatalf("radix Relname attribute should enable hex trace source missing %q:\n%s", want, out.Source)
	}
}

func TestReplTraceUsesRadixForArgsAndReturn(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..20}
action echo(i:idx) returns(out:idx) = {
    out := i
}
export echo
`)
	mod.SetAttribute("radix", "16")
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replhexfmt", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`fmt.Fprintf(__ivy_out, "echo(%s) {\n", ivyTraceValue(__ivy_arg0, true))`,
		`fmt.Fprintf(__ivy_out, "= %s\n", ivyTraceValue(__ivy_result, true))`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("radix=16 repl trace source missing %q:\n%s", want, out.Source)
		}
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

func TestReplImportCallbackEmitsTraceStub(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action notify(c:color) = {
    assert false
}
import notify
action trigger = {
    call notify(green)
}
export trigger
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replimp"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *replimp) notify(")
	if body == "" {
		t.Fatalf("notify method not emitted:\n%s", out.Source)
	}
	if !strings.Contains(body, `fmt.Fprintf(__ivy_out, "< notify(%s)\n", ivyTraceValue(c, false))`) {
		t.Fatalf("expected repl import callback trace stub:\n%s", body)
	}
	if strings.Contains(body, "ivyAssert") {
		t.Fatalf("repl import callback should not execute Ivy body:\n%s", body)
	}
	trigger := bodyAfterMarker(out.Source, "func (ivy *replimp) trigger(")
	if !strings.Contains(trigger, "ivy.notify(green)") {
		t.Fatalf("trigger should call notify callback stub:\n%s", trigger)
	}
}

func TestGeneratedReplImportCallbackPrintsTrace(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action notify(c:color) = {
    assert false
}
import notify
action trigger = {
    call notify(green)
}
export trigger
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replimprun"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte("trigger\n"), 0o644); err != nil {
		t.Fatalf("write command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err != nil {
		t.Fatalf("run generated repl import callback: %v\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", err, stdout, stderr, out.Source)
	}
	if stdout != "< notify(green)\n" {
		t.Fatalf("repl import callback output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestReplImportCallbackReturnUsesAskRet(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action poll returns (out:color) = {
    assert false
}
import poll
action trigger returns (out:color) = {
    out := poll
}
export trigger
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replret"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type ivyLineScanner interface",
		"func ivyAskRet(bound int) int",
		"__ivy_repl_scanner = bufio.NewScanner(os.Stdin)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("repl source missing %q:\n%s", want, out.Source)
		}
	}
	body := bodyAfterMarker(out.Source, "func (ivy *replret) poll(")
	if body == "" {
		t.Fatalf("poll method not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		`fmt.Fprintln(__ivy_out, "< poll")`,
		"return color(ivyAskRet(2))",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("return-valued repl import callback body missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "ivyAssert") {
		t.Fatalf("return-valued repl import callback should not execute Ivy body:\n%s", body)
	}
}

func TestGeneratedReplImportCallbackReturnReadsAskRet(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action poll returns (out:color) = {
    assert false
}
import poll
action trigger returns (out:color) = {
    out := poll
}
export trigger
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "replretrun"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte("trigger\n1\n"), 0o644); err != nil {
		t.Fatalf("write command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err != nil {
		t.Fatalf("run generated repl import callback return: %v\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", err, stdout, stderr, out.Source)
	}
	if stdout != "< poll\n? = green\n" {
		t.Fatalf("repl import callback return output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
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

const pingPongIvySource = `#lang ivy1.7

object intf = {
    action ping
    action pong
}

type side_t = {left,right}

specification {
    individual side : side_t
    after init {
        side := left
    }

    before intf.ping {
        require side = left;
        side := right
    }

    before intf.pong {
        require side = right;
        side := left
    }
}

implementation {
    isolate left_player = {
        individual ball : bool
        after init {
            ball := true
        }

        action hit = {
            if ball {
                call intf.ping;
                ball := false
            }
        }

        implement intf.pong {
            ball := true
        }

        invariant ball -> side = left
    } with this

    isolate right_player = {
        individual ball : bool
        after init {
            ball := false
        }

        action hit = {
            if ball {
                call intf.pong;
                ball := false
            }
        }

        implement intf.ping {
            ball := true
        }

        invariant ball -> side = right
    } with this
}

export left_player.hit
export right_player.hit
`

func TestPingPongLeftPlayerTargetTestMatchesIvy2CppRuntimeShape(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "pingpong.ivy")
	if err := os.WriteFile(spec, []byte(pingPongIvySource), 0o644); err != nil {
		t.Fatalf("write ping-pong source: %v", err)
	}

	out, err := CompileAndGenerate(spec, map[string]string{
		"target":    "test",
		"isolate":   "left_player",
		"classname": "pingpong",
	}, Config{TestIters: "30", TestRuns: "1"})
	if err != nil {
		t.Fatalf("CompileAndGenerate ping-pong left_player: %v\n%s", err, outSource(out))
	}
	assertNoUnsupportedGo(t, out.Source)

	intfPingBody := bodyAfterMarker(out.Source, "func (ivy *pingpong) intf__ping(")
	if intfPingBody == "" {
		t.Fatalf("intf__ping body not emitted:\n%s", out.Source)
	}
	if !strings.Contains(intfPingBody, `fmt.Fprint(__ivy_out, "< intf.ping"`) {
		t.Fatalf("intf__ping should trace the imported call like ivy2cpp/Python:\n%s", intfPingBody)
	}

	impPingBody := bodyAfterMarker(out.Source, "func (ivy *pingpong) imp__intf__ping(")
	if impPingBody == "" {
		t.Fatalf("imp__intf__ping body not emitted:\n%s", out.Source)
	}
	if strings.Contains(impPingBody, `fmt.Fprint(__ivy_out, "< imp__intf.ping"`) ||
		strings.Contains(impPingBody, `fmt.Fprint(__ivy_out, "< intf.ping"`) {
		t.Fatalf("implementation stub should not carry the imported-call trace:\n%s", impPingBody)
	}

	for _, marker := range []string{
		"func (ivy *pingpong) __init(",
		"func (ivy *pingpong) ext__intf__pong(",
		"func (ivy *pingpong) ext__left_player__hit(",
	} {
		body := bodyAfterMarker(out.Source, marker)
		if body == "" {
			t.Fatalf("%s body not emitted:\n%s", marker, out.Source)
		}
		if strings.Contains(body, "ivyAssert") &&
			strings.Contains(body, "left_player__ball") &&
			strings.Contains(body, "ivy.side == left") {
			t.Fatalf("%s should match ivy1.7 target=test output without appended invariant checks:\n%s", marker, body)
		}
	}
}

func TestEnumDispatchTraceMatchesIvy2Cpp(t *testing.T) {
	requireSlowTest(t)
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

func TestTickDeclaredForMinimalModule(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "tickempty"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func ivyCheckProgress(counter, max int)",
		"func (ivy *tickempty) __tick(timeout int)",
		"_ = timeout",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("minimal tick source missing %q:\n%s", want, out.Source)
		}
	}
	tickBody := bodyAfterMarker(out.Source, "func (ivy *tickempty) __tick(timeout int)")
	if tickBody == "" {
		t.Fatalf("__tick body not emitted:\n%s", out.Source)
	}
	if strings.Contains(tickBody, "ivyCheckProgress(") {
		t.Fatalf("empty tick should not call ivyCheckProgress:\n%s", tickBody)
	}
	assertNoUnsupportedGo(t, out.Source)
}

func TestProgressCounterDeclarationAndTickUpdate(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual ready : bool
relation ok(C:color)
progress wait = ready
progress waitn(C) = ok(C)
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "tickprog"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"wait int",
		"waitn map[color]int",
		"ivy.waitn = make(map[color]int)",
		"if ivy.ready {",
		"ivy.wait = 0",
		"ivy.wait = ivy.wait + 1",
		"for _, C := range []color{red, green}",
		"if ivy.ok[C] {",
		"ivy.waitn[C] = 0",
		"ivy.waitn[C] = ivy.waitn[C] + 1",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("progress source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestUnsupportedProgressConditionErrorIncludesSourceLine(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	native := &goivy.LogicNativeExpr{
		CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
	}
	def := goivy.NewDefinition(goivy.NewConst("wait", goivy.Boolean), native)
	def.SetLineno(goivy.Location{Filename: "progress.ivy", Line: 21})
	mod.Progress = []interface{}{def}

	_, err := Generate(mod, Config{Target: "test", ClassName: "badprogress", TestIters: "1"})
	if err == nil {
		t.Fatal("Generate should reject unsupported progress condition")
	}
	for _, want := range []string{
		"progress.ivy: line 21:",
		"unsupported progress condition wait",
		"native C++ expression is not translated to Go",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("progress error missing %q:\n%v", want, err)
		}
	}
}

func TestTickCallsIvyCheckProgressWithoutRely(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual ready : bool
progress wait = ready
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "ticknor"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"__tmp0 := 0",
		"if __tmp0 > timeout {",
		"ivy.wait = 0",
		"ivyCheckProgress(ivy.wait, __tmp0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("no-rely progress source missing %q:\n%s", want, out.Source)
		}
	}
	assertNoUnsupportedGo(t, out.Source)
	compileGeneratedGo(t, out)
}

func TestUnsupportedRelyExpressionErrorIncludesSourceLine(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual ready : bool
action step = {
}
export step
`)
	progress := goivy.NewDefinition(goivy.NewConst("wait", goivy.Boolean), goivy.NewConst("ready", goivy.Boolean))
	progress.SetLineno(goivy.Location{Filename: "progress.ivy", Line: 30})
	native := &goivy.LogicNativeExpr{
		CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
	}
	rely, err := goivy.NewImplies(goivy.NewConst("wait", goivy.Boolean), native)
	if err != nil {
		t.Fatalf("NewImplies: %v", err)
	}
	rely.SetLineno(goivy.Location{Filename: "rely.ivy", Line: 31})
	mod.Progress = []interface{}{progress}
	mod.Rely = []goivy.Expr{rely}

	_, err = Generate(mod, Config{Target: "test", ClassName: "badrely", TestIters: "1"})
	if err == nil {
		t.Fatal("Generate should reject unsupported rely expression")
	}
	for _, want := range []string{
		"rely.ivy: line 31:",
		"unsupported rely expression for wait",
		"native C++ expression is not translated to Go",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("rely error missing %q:\n%v", want, err)
		}
	}
}

func TestProgressBinaryTupleCounterAndTickUpdate(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type bit = {low, high}
relation edge(C:color, B:bit)
progress wait(C, B) = edge(C, B)
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "ticktuple"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"wait map[struct{ A0 color; A1 bit }]int",
		"ivy.wait = make(map[struct{ A0 color; A1 bit }]int)",
		"for _, C := range []color{red, green}",
		"for _, B := range []bit{low, high}",
		"if ivy.edge[struct{ A0 color; A1 bit }{C, B}] {",
		"ivy.wait[struct{ A0 color; A1 bit }{C, B}] = 0",
		"ivy.wait[struct{ A0 color; A1 bit }{C, B}] = ivy.wait[struct{ A0 color; A1 bit }{C, B}] + 1",
		"ivyCheckProgress(ivy.wait[struct{ A0 color; A1 bit }{C, B}], __tmp0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("tuple progress source missing %q:\n%s", want, out.Source)
		}
	}
	assertNoUnsupportedGo(t, out.Source)
	compileGeneratedGo(t, out)
}

func TestTickBareRelySkipsProgressCheckLikePython(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual ready : bool
progress wait = ready
rely wait
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "tickbarely"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"if ivy.ready {",
		"ivy.wait = 0",
		"ivy.wait = ivy.wait + 1",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("bare rely source missing progress update %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "ivyCheckProgress(ivy.wait") {
		t.Fatalf("bare rely should suppress the progress check like ivy2cpp/Python:\n%s", out.Source)
	}
}

func TestTickRelyImplicationComputesMaxAndChecksProgress(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual ready : bool
individual helper_ready : bool
progress wait = ready
progress helper = helper_ready
rely wait -> helper
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "tickrely"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"if ivy.helper > __tmp0 {",
		"__tmp0 = ivy.helper",
		"if __tmp0 > timeout {",
		"ivy.wait = 0",
		"ivyCheckProgress(ivy.wait, __tmp0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("rely progress source missing %q:\n%s", want, out.Source)
		}
	}
	assertNoUnsupportedGo(t, out.Source)
	compileGeneratedGo(t, out)
}

func TestTickRelySubstitutesProgressArgs(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation ready(C:color)
relation helper_ready(C:color)
progress wait(C) = ready(C)
progress helper(X) = helper_ready(X)
rely wait(Y) -> helper(Y)
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "ticksubst"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for _, C := range []color{red, green}",
		"if ivy.ready[C] {",
		"ivy.wait[C] = 0",
		"ivy.wait[C] = ivy.wait[C] + 1",
		"if ivy.helper[C] > __tmp0 {",
		"__tmp0 = ivy.helper[C]",
		"ivyCheckProgress(ivy.wait[C], __tmp0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("substituted rely source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{"ivy.helper[Y]", "ivy.wait[Y]"} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("substituted rely source kept binder name %q:\n%s", bad, out.Source)
		}
	}
	assertNoUnsupportedGo(t, out.Source)
}

func TestTickRelyExtraFreeVariableLoops(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation ready(C:color)
relation helper_ready(C:color)
progress wait(C) = ready(C)
progress helper(D) = helper_ready(D)
rely wait(C) -> helper(D)
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "tickextravar"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for _, C := range []color{red, green}",
		"for _, D__ := range []color{red, green}",
		"if ivy.helper[D__] > __tmp0 {",
		"__tmp0 = ivy.helper[D__]",
		"ivyCheckProgress(ivy.wait[C], __tmp0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("extra-variable rely source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{"if ivy.helper[D] > __tmp0", "__tmp0 = ivy.helper[D]\n"} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("extra-variable rely source was not alpha-renamed for %q:\n%s", bad, out.Source)
		}
	}
	assertNoUnsupportedGo(t, out.Source)
	compileGeneratedGo(t, out)
}

func TestTickRelyExtraSharesProgressVarNameRenamed(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation ready(C:color)
relation helper_ready(C:color, D:color)
progress wait(D) = ready(D)
progress helper(X, Y) = helper_ready(X, Y)
rely wait(C) -> helper(C, D)
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "tickextrarename"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for _, D := range []color{red, green}",
		"for _, D__ := range []color{red, green}",
		"ivy.helper[struct{ A0 color; A1 color }{D, D__}]",
		"ivyCheckProgress(ivy.wait[D], __tmp0)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("renamed rely source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "ivy.helper[struct{ A0 color; A1 color }{D, D}]") {
		t.Fatalf("rely RHS extra variable was captured instead of alpha-renamed:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "ivy.helper[struct{ A0 color; A1 color }{D__, D__}]") {
		t.Fatalf("rely LHS alias was over-renamed with the extra variable:\n%s", out.Source)
	}
}

func TestProgressCountersNotResetInConstructorOrInit(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual ready : bool
individual helper_ready : bool
progress wait = ready
progress helper = helper_ready
rely wait -> helper
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "tickctor"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, marker := range []string{
		"func newTickctor(",
		"func (ivy *tickctor) __initStorage(",
		"func (ivy *tickctor) __initState(",
		"func (ivy *tickctor) __init(",
	} {
		body := bodyAfterMarker(out.Source, marker)
		if body == "" {
			t.Fatalf("%s body not emitted:\n%s", marker, out.Source)
		}
		for _, bad := range []string{
			"ivy.wait = 0",
			"ivy.helper = 0",
		} {
			if strings.Contains(body, bad) {
				t.Fatalf("progress counter reset %q should stay out of constructor/init body %s:\n%s", bad, marker, body)
			}
		}
	}
	tickBody := bodyAfterMarker(out.Source, "func (ivy *tickctor) __tick(timeout int)")
	if tickBody == "" {
		t.Fatalf("__tick body not emitted:\n%s", out.Source)
	}
	if !strings.Contains(tickBody, "ivy.wait = 0") ||
		!strings.Contains(tickBody, "ivy.helper = 0") ||
		!strings.Contains(tickBody, "ivyCheckProgress(ivy.wait, __tmp0)") {
		t.Fatalf("progress update/check should remain in __tick:\n%s", tickBody)
	}
	assertNoUnsupportedGo(t, out.Source)
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
		"bvxor ivyThunkMap[struct{ A0 int; A1 int }, int]",
		"ivy.bvxor.Get(struct{ A0 int; A1 int }{ivy.x, ivy.y})",
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

func generateGoBVOperatorFixture(t *testing.T, width int, op string, arity int) *Output {
	t.Helper()
	src := fmt.Sprintf(`#lang ivy1.7
type word
interpret word -> bv[%d]
individual x : word
individual y : word
individual z : word
action step = {}
export step
`, width)
	mod := compileIvySource(t, src)
	word := mod.Sig.Sorts.Get("word")
	sorts := make([]goivy.Sort, 0, arity+1)
	args := make([]goivy.Expr, 0, arity)
	for i := 0; i < arity; i++ {
		sorts = append(sorts, word)
		name := "x"
		if i == 1 {
			name = "y"
		}
		args = append(args, goivy.NewConst(name, word))
	}
	sorts = append(sorts, word)
	fnSort, err := goivy.NewFunctionSort(sorts...)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	rhs := goivy.MustApply(goivy.NewConst(op, fnSort), args...)
	mod.Actions.Set("step", goivy.NewAssignAction(goivy.NewConst("z", word), rhs))
	out, err := Generate(mod, Config{ClassName: fmt.Sprintf("gobvop%d", width), Target: "test"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	return out
}

func TestBVBuiltinsLowerToGoBitwiseExpressions(t *testing.T) {
	for _, tc := range []struct {
		name  string
		op    string
		arity int
		want  []string
	}{
		{
			name:  "xor",
			op:    "bvxor",
			arity: 2,
			want:  []string{"ivy.z = ((ivy.x ^ ivy.y) & 255)"},
		},
		{
			name:  "neg",
			op:    "bvneg",
			arity: 1,
			want:  []string{"ivy.z = ((-(ivy.x)) & 255)"},
		},
		{
			name:  "shift-left",
			op:    "bvshl",
			arity: 2,
			want: []string{
				"__s := ivyBVShiftAmount(ivy.y)",
				"if __s >= 8",
				"return ((__x << __s) & 255)",
			},
		},
		{
			name:  "logical-shift-right",
			op:    "bvlshr",
			arity: 2,
			want: []string{
				"__s := ivyBVShiftAmount(ivy.y)",
				"if __s >= 8",
				"return ((__x >> __s) & 255)",
			},
		},
		{
			name:  "arithmetic-shift-right",
			op:    "bvashr",
			arity: 2,
			want: []string{
				"__sign := (__x & 128) != 0",
				"if __s >= 8",
				"__res = (__res | ((255 << (8 - __s)) & 255))",
			},
		},
		{
			name:  "symbolic-shift-left",
			op:    "<<",
			arity: 2,
			want:  []string{"return ((__x << __s) & 255)"},
		},
		{
			name:  "symbolic-shift-right",
			op:    ">>",
			arity: 2,
			want:  []string{"return ((__x >> __s) & 255)"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := generateGoBVOperatorFixture(t, 8, tc.op, tc.arity)
			for _, want := range tc.want {
				if !strings.Contains(out.Source, want) {
					t.Fatalf("source missing %q:\n%s", want, out.Source)
				}
			}
			if strings.Contains(out.Source, "ivy."+goName(tc.op)+"[") ||
				strings.Contains(out.Source, goName(tc.op)+"(") {
				t.Fatalf("built-in BV operator %q should not be emitted as uninterpreted storage/call:\n%s", tc.op, out.Source)
			}
		})
	}
}

func TestDeclaredBVOperatorNameRemainsStateFunction(t *testing.T) {
	out := generateGoBVOperatorFixture(t, 8, "bvxor", 2)
	if strings.Contains(out.Source, "ivy.bvxor[") {
		t.Fatalf("test setup unexpectedly declared bvxor:\n%s", out.Source)
	}
	mod := compileIvySource(t, `#lang ivy1.7
type word
interpret word -> bv[8]
individual x : word
individual y : word
individual z : word
function bvxor(X:word,Y:word) : word
action step = {
    z := bvxor(x,y)
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "declaredbv", Target: "test"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"bvxor ivyThunkMap[struct{ A0 int; A1 int }, int]",
		"ivy.z = ivy.bvxor.Get(struct{ A0 int; A1 int }{ivy.x, ivy.y})",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("declared bvxor should remain user state, missing %q:\n%s", want, out.Source)
		}
	}
}

func TestUnknownBVOperatorReportsUnsupported(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type word
interpret word -> bv[8]
individual x : word
individual y : word
individual z : word
action step = {}
export step
`)
	word := mod.Sig.Sorts.Get("word")
	fnSort, err := goivy.NewFunctionSort(word, word, word)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	rhs := goivy.MustApply(goivy.NewConst("bvfoo", fnSort), goivy.NewConst("x", word), goivy.NewConst("y", word))
	mod.Actions.Set("step", goivy.NewAssignAction(goivy.NewConst("z", word), rhs))
	_, err = Generate(mod, Config{ClassName: "bvunknown", Target: "test"})
	if err == nil || !strings.Contains(err.Error(), "unknown BV operator bvfoo/2") {
		t.Fatalf("Generate error = %v, want unknown BV operator diagnostic", err)
	}
}

func TestStringOpAddEmitsPlus(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strlit
individual a : text
individual b : text
individual c : text
action step = {
    c := a + b
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "stringadd", Target: "test"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.c = (ivy.a + ivy.b)") {
		t.Fatalf("string + should remain a Go string concatenation expression:\n%s", out.Source)
	}
}

func TestBVConcatBFEAndNumeralMask(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type nibble
type word
interpret nibble -> bv[4]
interpret word -> bv[8]
individual n0 : nibble
individual n1 : nibble
individual w : word
action step = {}
export step
`)
	nibble := mod.Sig.Sorts.Get("nibble")
	word := mod.Sig.Sorts.Get("word")
	concatSort, err := goivy.NewFunctionSort(nibble, nibble, word)
	if err != nil {
		t.Fatalf("NewFunctionSort concat: %v", err)
	}
	bfeSort, err := goivy.NewFunctionSort(word, nibble)
	if err != nil {
		t.Fatalf("NewFunctionSort bfe: %v", err)
	}
	concat := goivy.MustApply(
		goivy.NewConst("concat", concatSort),
		goivy.NewConst("n0", nibble),
		goivy.NewConst("n1", nibble),
	)
	bfe := goivy.MustApply(goivy.NewConst("bfe[0][3]", bfeSort), goivy.NewConst("w", word))
	mod.Actions.Set("step", goivy.NewSequence(
		goivy.NewAssignAction(goivy.NewConst("w", word), concat),
		goivy.NewAssignAction(goivy.NewConst("n0", nibble), bfe),
		goivy.NewAssignAction(goivy.NewConst("w", word), goivy.NewConst("300", word)),
	))
	out, err := Generate(mod, Config{ClassName: "bvconcat", Target: "test"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"ivy.w = (((ivy.n0) << 4 | (ivy.n1)) & 255)",
		"ivy.n0 = ((ivy.w >> 0) & 15)",
		"ivy.w = 44",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestBVOperatorTableCoversPythonAndZ3Names(t *testing.T) {
	data, err := os.ReadFile("bv_expr.go")
	if err != nil {
		t.Fatalf("read bv_expr.go: %v", err)
	}
	src := string(data)
	for _, op := range []string{
		"concat", "bvand", "bvor", "bvxor", "bvnot", "bvneg",
		"bvshl", "bvlshr", "bvashr", "<<", ">>",
		"bvadd", "bvsub", "bvmul", "bvudiv", "bvurem",
		"cast",
	} {
		if !strings.Contains(src, fmt.Sprintf("%q", op)) {
			t.Fatalf("bv_expr.go does not mention BV operator %q", op)
		}
	}
}

func TestBVBuiltinsGeneratedGoCompilesAndRuns(t *testing.T) {
	requireSlowTest(t)
	out := generateGoBVOperatorFixture(t, 8, "bvashr", 2)
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

const sortConstructorPointSource = `#lang ivy1.7
type point
destructor x_field(P:point) : bool
destructor y_field(P:point) : bool
constructor mkpoint : point
individual saved : point
action step = {
    saved := mkpoint(true,false);
    assert x_field(saved)
}
export step
`

func TestSortConstructorEmitsGoMethodAndCall(t *testing.T) {
	mod := compileIvySource(t, sortConstructorPointSource)
	conss, ok := mod.SortConstructors.Get2("point")
	if !ok || len(conss) == 0 {
		t.Fatalf("test setup: no constructor for sort point; module: %+v", mod.SortConstructors)
	}
	out, err := Generate(mod, Config{ClassName: "mkpointgo", Target: "test"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type point struct",
		"x_field bool",
		"y_field bool",
		"func (ivy *mkpointgo) mkpoint(X0 bool, X1 bool) point",
		"val := point{}",
		"val.x_field = X0",
		"val.y_field = X1",
		"return val",
		"ivy.saved = ivy.mkpoint(true, false)",
		"ivyAssert(ivy.saved.x_field",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("constructor source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "\n\tmkpoint point\n") ||
		strings.Contains(out.Source, "\n\tmkpoint func") {
		t.Fatalf("constructor leaked into state fields:\n%s", out.Source)
	}
}

func TestSortConstructorEmitsWithoutActions(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type point
destructor x_field(P:point) : bool
constructor mkpoint : point
`)
	out, err := Generate(mod, Config{ClassName: "mkonly", Target: "test"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type point struct",
		"func (ivy *mkonly) mkpoint(X0 bool) point",
		"val.x_field = X0",
		"return val",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("constructor-only source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "\n\tmkpoint point\n") ||
		strings.Contains(out.Source, "\n\tmkpoint func") {
		t.Fatalf("constructor leaked into state fields:\n%s", out.Source)
	}
}

func TestSortConstructorGeneratedGoCompilesAndRuns(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, sortConstructorPointSource)
	out, err := Generate(mod, Config{ClassName: "mkpointgo", Target: "test"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
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

func TestOracleNativeActionsRejectBehaviorWeakening(t *testing.T) {
	for _, name := range []string{"native_block", "callback_thunk"} {
		t.Run(name, func(t *testing.T) {
			fixture := filepath.Join("..", "ivy2cpp", "test_vec", "oracle", name+".ivy")
			out, err := CompileAndGenerate(fixture, map[string]string{"target": "test", "classname": name}, Config{TestIters: "1"})
			if err == nil {
				t.Fatalf("CompileAndGenerate should reject native action weakening:\n%s", outSource(out))
			}
			for _, want := range []string{
				"native C++ action code is not translated to Go",
				"action",
			} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("native action rejection missing %q:\n%v", want, err)
				}
			}
		})
	}
}

func TestTopLevelNativeBlocksWarnAndEmitGoComments(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
<<< header
#include <sstream>
>>>
<<< member
int native_counter;
>>>
<<< init
native_counter = 7;
>>>
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "native_blocks", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(strings.Join(out.Warnings, "\n"), "top-level native C++ blocks") {
		t.Fatalf("expected top-level native warning, got %v", out.Warnings)
	}
	if got := strings.Count(out.Source, "ivy top-level native block omitted"); got != 3 {
		t.Fatalf("top-level native marker count=%d, want 3:\n%s", got, out.Source)
	}
	for _, want := range []string{
		"// native: header",
		"// native: #include <sstream>",
		"// native: member",
		"// native: int native_counter;",
		"// native: init",
		"// native: native_counter = 7;",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("top-level native source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestNativeTypeInterpretRejectsIntPlaceholderWeakening(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type handle
interpret handle -> <<< primitive int >>>
individual h : handle
action step(x:handle) = {
    h := x
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "native_type", TestIters: "1", Build: true})
	if err == nil {
		t.Fatalf("Generate should reject native type placeholder weakening:\n%s", outSource(out))
	}
	for _, want := range []string{
		"native C++ type interpretation is not translated to Go",
		"handle",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("native type rejection missing %q:\n%v", want, err)
		}
	}
}

const nativeDefinitionSource = `#lang ivy1.6
type idx = {0..3}
function size(A:idx) : idx
definition size(A:idx) = <<< 0 >>>
<<< native ` + "`size`" + ` >>>
individual saved : idx
action step(x:idx) = {
    saved := size(x)
}
export step
`

func TestNativeDefinitionRejectsZeroValueStubWeakening(t *testing.T) {
	mod := compileIvySource(t, nativeDefinitionSource)
	out, err := Generate(mod, Config{Target: "test", ClassName: "native_def", TestIters: "1"})
	if err == nil {
		t.Fatalf("Generate should reject native definition zero-value weakening:\n%s", outSource(out))
	}
	for _, want := range []string{
		"native C++ expression definition is not translated to Go",
		"size",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("native definition rejection missing %q:\n%v", want, err)
		}
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
		"gen.c = cell{shade: color(ivy.___ivy_randomize(2",
		"__arg0 := store_generator.c",
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

func TestGeneratedSetActionWithUnboundedFreeVariableUsesThunkAssign(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type item
relation marked(I:item)
action step = {
}
export step
`)
	item, ok := mod.Sig.Sorts.Get2("item")
	if !ok {
		t.Fatal("missing item sort")
	}
	marked, err := mod.Sig.FindSymbol("marked", false)
	if err != nil {
		t.Fatalf("FindSymbol marked: %v", err)
	}
	i, err := goivy.NewVariable("I", item)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	markedApp, err := goivy.NewApply(goivy.NewConst("marked", marked.CSort), i)
	if err != nil {
		t.Fatalf("NewApply marked: %v", err)
	}
	mod.Actions.Set("step", goivy.NewSetAction(markedApp))
	out, err := Generate(mod, Config{Target: "test", ClassName: "set_unbounded", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"marked ivyThunkMap[int, bool]",
		"__ivy_old_marked",
		"ivy.marked = newIvyThunkMap[int, bool](false)",
		"ivy.marked.base = func(__ivy_key",
		"return true",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("unbounded set action source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "unsupported set over free variable") {
		t.Fatalf("unbounded set action should use thunk assignment fallback:\n%s", out.Source)
	}
}

func TestUnboundedSetActionWithRepeatedVariableConstrainsThunkKey(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type item
relation edge(I:item,J:item)
action step = {
}
export step
`)
	item, ok := mod.Sig.Sorts.Get2("item")
	if !ok {
		t.Fatal("missing item sort")
	}
	edge, err := mod.Sig.FindSymbol("edge", false)
	if err != nil {
		t.Fatalf("FindSymbol edge: %v", err)
	}
	i, err := goivy.NewVariable("I", item)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	edgeApp, err := goivy.NewApply(goivy.NewConst("edge", edge.CSort), i, i)
	if err != nil {
		t.Fatalf("NewApply edge: %v", err)
	}
	mod.Actions.Set("step", goivy.NewSetAction(edgeApp))
	out, err := Generate(mod, Config{Target: "test", ClassName: "set_diag", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *set_diag) step()")
	if body == "" {
		t.Fatalf("step body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		"edge ivyThunkMap[struct{ A0 int; A1 int }, bool]",
		"I := __ivy_key",
		"X1 := __ivy_key",
		"return ivyTernary((I == X1), true, __ivy_old_edge",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("repeated-variable set source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Count(body, "I :=") != 1 {
		t.Fatalf("repeated-variable set should bind I once in thunk body:\n%s", body)
	}
}

func TestIssue60SetActionOverSymbolicRangeUsesRuntimeBoundLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.8
type client_id

module iterable = {
    interpret this -> {0..max}
}

global {
    instance client_id : iterable
}

relation marked(C:client_id)
action step = {
}
export step
`)
	clientID, ok := mod.Sig.Sorts.Get2("client_id")
	if !ok {
		t.Fatal("missing client_id sort")
	}
	marked, err := mod.Sig.FindSymbol("marked", false)
	if err != nil {
		t.Fatalf("FindSymbol marked: %v", err)
	}
	c, err := goivy.NewVariable("C", clientID)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	markedApp, err := goivy.NewApply(goivy.NewConst("marked", marked.CSort), c)
	if err != nil {
		t.Fatalf("NewApply marked: %v", err)
	}
	mod.Actions.Set("step", goivy.NewSetAction(markedApp))
	out, err := Generate(mod, Config{Target: "test", ClassName: "issue60set", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for C := 0; C < (ivy.client_id__max + 1); C++",
		"ivy.marked.Set(C, true)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("symbolic range set source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "[]int{") {
		t.Fatalf("symbolic range set should not expand to a literal slice:\n%s", out.Source)
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

func TestEmitActionMatrixCoverage(t *testing.T) {
	const (
		emits  = "emits"
		noop   = "noop"
		errors = "errors"
	)
	acfg := goivy.NewAstConfig()
	cfg := goivy.NewActionsConfig()
	flag := goivy.NewConst("flag", goivy.Boolean)
	trueExpr := goivy.NewConst("true", goivy.Boolean)
	enum := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	node := &goivy.UninterpretedSort{Name: "node"}
	fldSort, err := goivy.NewFunctionSort(node, enum)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	shade := goivy.NewConst("shade", fldSort)
	obj := goivy.NewConst("obj", node)
	srcObj := goivy.NewConst("src", node)
	green := goivy.NewConst("green", enum)
	cases := []struct {
		name   string
		build  func() goivy.Action
		want   string
		errSub string
	}{
		{name: "Sequence", build: func() goivy.Action {
			return goivy.NewSequence(goivy.NewAssignAction(flag, trueExpr))
		}, want: emits},
		{name: "AssertAction", build: func() goivy.Action {
			return goivy.NewAssertAction(trueExpr)
		}, want: emits},
		{name: "EnsuresAction", build: func() goivy.Action {
			return goivy.NewEnsuresAction(trueExpr)
		}, want: emits},
		{name: "RequiresAction", build: func() goivy.Action {
			return goivy.NewRequiresAction(trueExpr)
		}, want: emits},
		{name: "SubgoalAction", build: func() goivy.Action {
			return goivy.NewSubgoalAction(trueExpr)
		}, want: emits},
		{name: "AssumeAction", build: func() goivy.Action {
			return goivy.NewAssumeAction(trueExpr)
		}, want: emits},
		{name: "AssignAction", build: func() goivy.Action {
			return goivy.NewAssignAction(flag, trueExpr)
		}, want: emits},
		{name: "SetAction", build: func() goivy.Action {
			return goivy.NewSetAction(flag)
		}, want: emits},
		{name: "AssignFieldAction", build: func() goivy.Action {
			return goivy.NewAssignFieldAction(shade, obj, green)
		}, want: emits},
		{name: "NullFieldAction", build: func() goivy.Action {
			return goivy.NewNullFieldAction(shade, obj)
		}, want: emits},
		{name: "CopyFieldAction", build: func() goivy.Action {
			return goivy.NewCopyFieldAction(obj, shade, srcObj, shade)
		}, want: emits},
		{name: "ChoiceAction", build: func() goivy.Action {
			return goivy.NewChoiceActionOn(cfg,
				goivy.NewAssignAction(flag, trueExpr),
				goivy.NewAssignAction(flag, trueExpr))
		}, want: emits},
		{name: "EnvAction", build: func() goivy.Action {
			return goivy.NewEnvActionOn(cfg,
				goivy.NewAssignAction(flag, trueExpr),
				goivy.NewAssignAction(flag, trueExpr))
		}, want: emits},
		{name: "IfAction", build: func() goivy.Action {
			return goivy.NewIfAction(flag, goivy.NewAssignAction(flag, trueExpr))
		}, want: emits},
		{name: "WhileAction", build: func() goivy.Action {
			return goivy.NewWhileAction(flag, goivy.NewAssignAction(flag, trueExpr))
		}, want: emits},
		{name: "LocalAction", build: func() goivy.Action {
			return goivy.NewLocalActionOn(cfg, "test", flag, goivy.NewAssignAction(flag, trueExpr))
		}, want: emits},
		{name: "LetAction", build: func() goivy.Action {
			return goivy.NewLetAction(goivy.NewAssignAction(flag, trueExpr))
		}, want: emits},
		{name: "DebugAction", build: func() goivy.Action {
			return goivy.NewDebugAction(goivy.NewConst(`"tick"`, goivy.TopS))
		}, want: emits},
		{name: "NativeAction", build: func() goivy.Action {
			return goivy.NewNativeAction(acfg.NewNativeCode("// noop"))
		}, want: errors, errSub: "native C++ action code is not translated to Go"},
		{name: "ReturnAction", build: func() goivy.Action {
			return goivy.NewReturnAction()
		}, want: emits},
		{name: "IgnoreAction", build: func() goivy.Action {
			return goivy.NewIgnoreAction()
		}, want: noop},
		{name: "CrashAction", build: func() goivy.Action {
			return goivy.NewCrashAction(goivy.NewConst("anything", goivy.TopS))
		}, want: noop},
		{name: "HavocAction", build: func() goivy.Action {
			return goivy.NewHavocAction(flag)
		}, want: errors, errSub: "havoc reached emit"},
		{name: "BindOldsAction", build: func() goivy.Action {
			return goivy.NewBindOldsAction(goivy.NewAssignAction(flag, trueExpr))
		}, want: errors, errSub: "bindolds reached emit"},
		{name: "ThunkAction", build: func() goivy.Action {
			return goivy.NewThunkAction(goivy.NewConst("x", goivy.TopS))
		}, want: errors, errSub: "thunk reached emit"},
		{name: "InstantiateAction", build: func() goivy.Action {
			return goivy.NewInstantiateAction(goivy.NewConst("schema", goivy.TopS))
		}, want: errors, errSub: "instantiate reached emit"},
		{name: "Ranking", build: func() goivy.Action {
			return goivy.NewRanking(goivy.NewConst("r", goivy.TopS))
		}, want: errors, errSub: "ranking reached emit"},
	}
	mod := goivy.New()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var w goWriter
			g := &Generator{Mod: mod, currentReturns: []*goivy.Const{flag}}
			g.emitAction(&w, tc.build())
			out := strings.TrimSpace(w.String())
			switch tc.want {
			case emits:
				if out == "" {
					t.Fatalf("%s: expected non-empty Go output", tc.name)
				}
				if len(g.errs) != 0 {
					t.Fatalf("%s: expected no errors for has-emit class, got %v", tc.name, g.errs)
				}
			case noop:
				if out != "" {
					t.Fatalf("%s: expected empty output, got %q", tc.name, out)
				}
				if len(g.errs) != 0 {
					t.Fatalf("%s: expected no errors, got %v", tc.name, g.errs)
				}
			case errors:
				if len(g.errs) == 0 {
					t.Fatalf("%s: expected an error identifying the Python class, got none (output=%q)", tc.name, out)
				}
				gotErr := g.errs[0].Error()
				if !strings.Contains(gotErr, tc.errSub) {
					t.Fatalf("%s: error %q missing substring %q", tc.name, gotErr, tc.errSub)
				}
			}
		})
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
			tc.act.SetLineno(goivy.Location{Filename: "unsupported_actions.ivy", Line: 31})
			mod := goivy.New()
			mod.Name = tc.name
			mod.Actions.Set("step", tc.act)
			_, err := Generate(mod, Config{Target: "test", ClassName: tc.name, TestIters: "1"})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unexpected Generate error: %v", err)
			}
			wantPrefix := "ivy2golang: unsupported_actions.ivy: line 31: " + tc.want
			if !strings.HasPrefix(err.Error(), wantPrefix) {
				t.Fatalf("unsupported action error should start with %q:\n%v", wantPrefix, err)
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

func TestUnsupportedAssignmentExpressionErrorIncludesSourceLine(t *testing.T) {
	flag := goivy.NewConst("flag", goivy.Boolean)
	assign := goivy.NewAssignAction(flag, &goivy.LogicNativeExpr{
		CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
	})
	assign.SetLineno(goivy.Location{Filename: "test.ivy", Line: 7})

	mod := goivy.New()
	mod.Name = "assignloc"
	mod.Actions.Set("step", assign)
	mod.PublicActions.Set("step", true)

	_, err := Generate(mod, Config{Target: "impl", ClassName: "assignloc"})
	if err == nil {
		t.Fatal("Generate should reject unsupported assignment expression")
	}
	for _, want := range []string{
		"test.ivy: line 7:",
		"unsupported assignment rhs",
		"native C++ expression is not translated to Go",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("unsupported assignment error missing %q:\n%v", want, err)
		}
	}
}

func TestUnsupportedCallArgumentErrorIncludesSourceLine(t *testing.T) {
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Actions.Set("callee", goivy.NewSequence())

	nativeArg := &goivy.LogicNativeExpr{
		CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
	}
	callee := goivy.NewConst("callee", goivy.TopS)
	call := goivy.NewCallActionOn(mod.Cfg.ActCfg, goivy.NewApplyUnchecked(callee, nativeArg))
	call.AstCallee = mod.Cfg.AstCfg.NewAtom("callee")
	call.SetLineno(goivy.Location{Filename: "calls.ivy", Line: 11})
	mod.Actions.Set("step", call)
	mod.PublicActions.Set("step", true)

	_, err := Generate(mod, Config{Target: "impl", ClassName: "callloc"})
	if err == nil {
		t.Fatal("Generate should reject unsupported call argument")
	}
	for _, want := range []string{
		"calls.ivy: line 11:",
		"unsupported call argument",
		"native C++ expression is not translated to Go",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("unsupported call error missing %q:\n%v", want, err)
		}
	}
}

func TestUnsupportedDebugExpressionErrorIncludesSourceLine(t *testing.T) {
	nativeExpr := &goivy.LogicNativeExpr{
		CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
	}
	debug := goivy.NewDebugAction(goivy.NewConst(`"tick"`, goivy.TopS), nativeExpr)
	debug.SetLineno(goivy.Location{Filename: "debug.ivy", Line: 13})

	mod := goivy.New()
	mod.Name = "debugloc"
	mod.Actions.Set("step", debug)
	mod.PublicActions.Set("step", true)

	_, err := Generate(mod, Config{Target: "impl", ClassName: "debugloc"})
	if err == nil {
		t.Fatal("Generate should reject unsupported debug expression")
	}
	for _, want := range []string{
		"debug.ivy: line 13:",
		"unsupported debug print expression",
		"native C++ expression is not translated to Go",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("unsupported debug error missing %q:\n%v", want, err)
		}
	}
}

func TestUnsupportedLetActionBodyErrorIncludesSourceLine(t *testing.T) {
	letAct := goivy.NewLetAction(goivy.NewDefinition(goivy.NewConst("x", goivy.Boolean), goivy.NewConst("true", goivy.Boolean)))
	letAct.SetLineno(goivy.Location{Filename: "letact.ivy", Line: 17})

	mod := goivy.New()
	mod.Name = "letloc"
	mod.Actions.Set("step", letAct)
	mod.PublicActions.Set("step", true)

	_, err := Generate(mod, Config{Target: "impl", ClassName: "letloc"})
	if err == nil {
		t.Fatal("Generate should reject unsupported let action body")
	}
	for _, want := range []string{
		"letact.ivy: line 17:",
		"unsupported let action body",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("unsupported let action error missing %q:\n%v", want, err)
		}
	}
}

func TestUnsupportedLocalDeclarationErrorIncludesSourceLine(t *testing.T) {
	nativeLocal := &goivy.LogicNativeExpr{
		CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
	}
	localAct := goivy.NewLocalActionOn(goivy.NewActionsConfig(), "test", nativeLocal, goivy.NewSequence())
	localAct.SetLineno(goivy.Location{Filename: "locals.ivy", Line: 19})

	mod := goivy.New()
	mod.Name = "localdeclloc"
	mod.Actions.Set("step", localAct)
	mod.PublicActions.Set("step", true)

	_, err := Generate(mod, Config{Target: "impl", ClassName: "localdeclloc"})
	if err == nil {
		t.Fatal("Generate should reject unsupported local declaration")
	}
	for _, want := range []string{
		"locals.ivy: line 19:",
		"unsupported local declaration",
		"native(1)",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("unsupported local declaration error missing %q:\n%v", want, err)
		}
	}
}

func TestUnsupportedLocalFunctionRangeErrorIncludesSourceLine(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action step = {
}
export step
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	tableSort, err := goivy.NewFunctionSort(color, goivy.TopS)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	localAct := goivy.NewLocalActionOn(goivy.NewActionsConfig(), "test", goivy.NewConst("table", tableSort), goivy.NewSequence())
	localAct.SetLineno(goivy.Location{Filename: "locals.ivy", Line: 23})
	mod.Actions.Set("step", localAct)

	_, err = Generate(mod, Config{Target: "test", ClassName: "localrangeloc", TestIters: "1"})
	if err == nil {
		t.Fatal("Generate should reject unsupported local function range")
	}
	for _, want := range []string{
		"locals.ivy: line 23:",
		"unsupported local function range",
		"unsupported local nondet sort TopSort",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("unsupported local function range error missing %q:\n%v", want, err)
		}
	}
}

func TestUnsupportedDoesNotDoubleIvy2GolangPrefix(t *testing.T) {
	var w goWriter
	g := &Generator{}
	g.unsupported(&w, "%s", "ivy2golang: cannot create test generator because type data is non-enumerable")
	if len(g.errs) != 1 {
		t.Fatalf("expected one error, got %d", len(g.errs))
	}
	if got := g.errs[0].Error(); got != "ivy2golang: cannot create test generator because type data is non-enumerable" {
		t.Fatalf("unexpected normalized error: %q", got)
	}
	if strings.Contains(w.String(), "ivy2golang: ivy2golang:") {
		t.Fatalf("panic text should not double-prefix:\n%s", w.String())
	}
}

func TestJoinUniqueErrorsDeduplicatesMessages(t *testing.T) {
	err := joinUniqueErrors([]error{
		fmt.Errorf("ivy2golang: duplicate"),
		fmt.Errorf("ivy2golang: duplicate"),
		fmt.Errorf("ivy2golang: distinct"),
	})
	if err == nil {
		t.Fatal("expected joined error")
	}
	if got := err.Error(); got != "ivy2golang: duplicate\nivy2golang: distinct" {
		t.Fatalf("unexpected joined error:\n%s", got)
	}
}

func TestEmitStringInterpConstantZero(t *testing.T) {
	g := newGoTestGeneratorWithInterps(map[string]string{"str": "strlit"})
	strSort := &goivy.UninterpretedSort{Name: "str"}
	zero := goivy.NewConst("0", strSort)

	got, err := g.emitExpr(zero)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != `""` {
		t.Fatalf("expected empty string literal for strlit 0, got: %s", got)
	}
}

func TestEmitStringInterpConstantNonzeroErrors(t *testing.T) {
	g := newGoTestGeneratorWithInterps(map[string]string{"str": "strlit"})
	strSort := &goivy.UninterpretedSort{Name: "str"}
	five := goivy.NewConst("5", strSort)

	_, err := g.emitExpr(five)
	if err == nil || !strings.Contains(err.Error(), "string sort") {
		t.Fatalf("expected string-sort numeral error, got: %v", err)
	}
}

func TestGeneratedStringInterpZeroConstant(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strlit
individual saved : text
action step = {
    saved := 0
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "strzero", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, `ivy.saved = ""`) {
		t.Fatalf("strlit zero should emit empty string assignment:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "ivy.saved = 0") {
		t.Fatalf("strlit zero must not emit numeric assignment:\n%s", out.Source)
	}
}

func TestGeneratedStringInterpZeroConstantCompilesAndRuns(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strlit
individual saved : text
action step = {
    saved := 0
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "strzero", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
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

func TestTargetTestUnboundedQuantifierSoftDiagnosticsIncludeLines(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..1}
type version
interpret version -> nat

individual cur : idx
individual witness : version
relation ts_version(I:idx,V:version)

after init {
    cur := 0;
}

action check(p:version) = {
    if (exists V. ts_version(cur,V)) {
        witness := p
    };
    assume forall UV. ts_version(cur,UV) -> cur = cur;
}

export check
`)
	out, err := Generate(mod, Config{ClassName: "natsoft", Target: "test"})
	if err != nil {
		t.Fatalf("Generate target=test should keep Python/C++-style unsupported expressions soft, got: %v", err)
	}
	for _, want := range []string{
		"// ivy2golang: unsupported if condition: test.ivy: line 15: error: cannot enumerate quantified variable V:version",
		"// ivy2golang: unsupported assumption expression: test.ivy: line 18: error: cannot enumerate quantified variable UV:version",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("soft unsupported diagnostic missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, `panic("ivy2golang: unsupported if condition`) ||
		strings.Contains(out.Source, `panic("ivy2golang: unsupported assumption expression`) {
		t.Fatalf("soft expression fallback should emit comments, not panics:\n%s", out.Source)
	}
	checkBody := bodyAfterMarker(out.Source, "func (ivy *natsoft) check(")
	if checkBody == "" {
		t.Fatalf("check body not emitted:\n%s", out.Source)
	}
	for _, bad := range []string{
		"for V :=",
		"for UV :=",
		"ivy.ts_version[",
	} {
		if strings.Contains(checkBody, bad) {
			t.Fatalf("soft unsupported quantified expression should discard partial loop/code %q:\n%s", bad, checkBody)
		}
	}
}

func TestUnsupportedExportedAssumeAndIfConditionsAreFatal(t *testing.T) {
	for _, tc := range []struct {
		name      string
		build     func(goivy.Expr) goivy.Action
		wantKind  string
		wantLine  int
		className string
	}{
		{
			name: "assume",
			build: func(native goivy.Expr) goivy.Action {
				act := goivy.NewAssumeAction(native)
				act.SetLineno(goivy.Location{Filename: "soft.ivy", Line: 31})
				return act
			},
			wantKind:  "unsupported target=test action generator assume guard",
			wantLine:  31,
			className: "fatal_assume",
		},
		{
			name: "if",
			build: func(native goivy.Expr) goivy.Action {
				act := goivy.NewIfAction(native, goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("true", goivy.Boolean)))
				act.SetLineno(goivy.Location{Filename: "soft.ivy", Line: 37})
				return act
			},
			wantKind:  "unsupported if condition",
			wantLine:  37,
			className: "fatal_if",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mod := compileIvySource(t, `#lang ivy1.7
individual flag : bool
action step = {
}
export step
`)
			native := &goivy.LogicNativeExpr{
				CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
			}
			native.SetLineno(goivy.Location{Filename: "soft.ivy", Line: tc.wantLine})
			mod.Actions.Set("step", tc.build(native))

			out, err := Generate(mod, Config{ClassName: tc.className, Target: "test", TestIters: "1"})
			if err == nil {
				t.Fatalf("Generate should reject non-compatible soft unsupported path:\n%s", outSource(out))
			}
			for _, want := range []string{
				fmt.Sprintf("soft.ivy: line %d:", tc.wantLine),
				tc.wantKind,
				"native C++ expression is not translated to Go",
			} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("fatal soft-unsupported error missing %q:\n%v", want, err)
				}
			}
		})
	}
}

func TestQuantifierMixedFiniteThenBoundedNatUsesPerVariableBounds(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type ts
interpret ts -> bv[2]
type version
interpret version -> nat

relation ts_version(T:ts,V:version)
relation lt(T:ts,U:ts)

action check(new_ver:version,t:ts) = {
    assume forall U,UV. ts_version(U,UV) & UV < new_ver -> lt(U,t)
}

export check
`)
	out, err := Generate(mod, Config{ClassName: "mixedbounds", Target: "test", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for U := 0; U < 4; U++ {",
		"for UV := 0; UV < new_ver; UV++ {",
		"ivy.ts_version.Get(struct{ A0 int; A1 int }{U, UV})",
		"ivy.lt[struct{ A0 int; A1 int }{U, t}]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("mixed quantifier source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "cannot enumerate quantified variable UV:version") {
		t.Fatalf("bounded later quantified variable should not fall through to finite enumeration:\n%s", out.Source)
	}
}

func TestSparseUnboundedRelationQuantifierUsesThunkSupport(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type ts = {0..3}
type version
interpret version -> nat

individual init_ts : ts
relation ts_version(T:ts,V:version)
individual found : bool

after init {
    init_ts := 0;
    ts_version(T,V) := T = init_ts & V = 0
}

action add(t:ts,v:version) = {
    ts_version(T,V) := ts_version(T,V) | (T = t & V = v)
}

action check(t:ts,new_ver:version) = {
    if (exists BV. ts_version(t,BV)) {
        found := true
    };
    assume ts_version(U,UV) & new_ver < UV -> t = U;
}

export add
export check
`)
	out, err := Generate(mod, Config{ClassName: "sparse_support", Target: "test", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"ts_version ivyThunkMap[struct{ A0 int; A1 int }, bool]",
		"ivy.ts_version.Set(struct{ A0 int; A1 int }{ivy.init_ts, 0}, true)",
		"for __ivy_quant_key",
		"range ivy.ts_version.overrides",
		"BV := __ivy_quant_key",
		".A1",
		"ivy.found = true",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("sparse support source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		"unsupported if condition",
		"unsupported assumption expression",
		"cannot enumerate quantified variable",
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("sparse support quantifier should not fall through to %q:\n%s", bad, out.Source)
		}
	}
}

func TestHermesStyleUnboundedVersionQuantifierUsesRelationSupport(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type ts
type version
interpret version -> nat

individual init_ts : ts
relation ts_version(T:ts,V:version)
relation lt(T:ts,U:ts)
individual found : bool

after init {
    ts_version(T,V) := T = init_ts & V = 0
}

action check(t:ts,new_ver:version) = {
    if (exists BV. ts_version(t,BV)) {
        found := true
    };
    assume forall U,UV. ts_version(U,UV) & new_ver < UV -> lt(t,U)
}

export check
`)
	out, err := Generate(mod, Config{ClassName: "hermes_quant_support", Target: "test", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"ts_version ivyThunkMap[struct{ A0 int; A1 int }, bool]",
		"ivy.ts_version.Set(struct{ A0 int; A1 int }{ivy.init_ts, 0}, true)",
		"range ivy.ts_version.overrides",
		"U := __ivy_quant_key",
		"UV := __ivy_quant_key",
		".A0",
		".A1",
		"ivy.lt.Get(struct{ A0 int; A1 int }{t, U})",
		"ivy.found = true",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("Hermes-style support source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		"unsupported if condition",
		"unsupported assumption expression",
		"cannot enumerate quantified variable",
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("Hermes-style support should not fall through to %q:\n%s", bad, out.Source)
		}
	}
}

func TestEmitLetExpression(t *testing.T) {
	sort := &goivy.UninterpretedSort{Name: "int"}
	x := goivy.NewConst("x", sort)
	seven := goivy.NewConst("7", sort)
	addSort, err := goivy.NewFunctionSort(sort, sort, sort)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	plus := goivy.NewConst("+", addSort)
	body := goivy.MustApply(plus, x, x)
	def := goivy.NewDefinition(x, seven)
	let := goivy.NewLet([]goivy.Expr{def}, body)

	got, err := newGoTestGeneratorWithInterps(nil).emitExpr(let)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if strings.Contains(got, "let ") {
		t.Fatalf("let was not expanded: %s", got)
	}
	if !strings.Contains(got, "7 + 7") {
		t.Fatalf("let substitution missing 7 + 7: %s", got)
	}
}

func TestEmitParametricLetExpression(t *testing.T) {
	sort := &goivy.UninterpretedSort{Name: "int"}
	x := &goivy.LogicVariable{Name: "X", VSort: sort}
	two := goivy.NewConst("2", sort)
	three := goivy.NewConst("3", sort)
	five := goivy.NewConst("5", sort)
	addSort, err := goivy.NewFunctionSort(sort, sort, sort)
	if err != nil {
		t.Fatalf("NewFunctionSort add: %v", err)
	}
	plus := goivy.NewConst("+", addSort)
	fnSort, err := goivy.NewFunctionSort(sort, sort)
	if err != nil {
		t.Fatalf("NewFunctionSort f: %v", err)
	}
	f := goivy.NewConst("f", fnSort)
	def := goivy.NewDefinition(goivy.MustApply(f, x), goivy.MustApply(plus, x, three))
	let := goivy.NewLet([]goivy.Expr{def}, goivy.MustApply(plus, goivy.MustApply(f, two), five))

	got, err := newGoTestGeneratorWithInterps(nil).emitExpr(let)
	if err != nil {
		t.Fatalf("emitExpr parametric let: %v", err)
	}
	for _, want := range []string{"2 + 3", "+ 5"} {
		if !strings.Contains(got, want) {
			t.Fatalf("parametric let substitution missing %q in %s", want, got)
		}
	}
	if strings.Contains(got, "f(") || strings.Contains(got, "parametric let definition not yet supported") {
		t.Fatalf("parametric let should be expanded, got %s", got)
	}
}

func TestGeneratedParametricLetExpression(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..10}
individual saved : idx
action step = {
}
export step
`)
	idx, ok := mod.Sig.Sorts.Get2("idx")
	if !ok {
		t.Fatal("missing idx sort")
	}
	x := &goivy.LogicVariable{Name: "X", VSort: idx}
	two := goivy.NewConst("2", idx)
	three := goivy.NewConst("3", idx)
	five := goivy.NewConst("5", idx)
	addSort, err := goivy.NewFunctionSort(idx, idx, idx)
	if err != nil {
		t.Fatalf("NewFunctionSort add: %v", err)
	}
	plus := goivy.NewConst("+", addSort)
	fnSort, err := goivy.NewFunctionSort(idx, idx)
	if err != nil {
		t.Fatalf("NewFunctionSort f: %v", err)
	}
	f := goivy.NewConst("f", fnSort)
	def := goivy.NewDefinition(goivy.MustApply(f, x), goivy.MustApply(plus, x, three))
	let := goivy.NewLet([]goivy.Expr{def}, goivy.MustApply(plus, goivy.MustApply(f, two), five))
	mod.Actions.Set("step", goivy.NewAssignAction(goivy.NewConst("saved", idx), let))

	out, err := Generate(mod, Config{Target: "test", ClassName: "paramlet", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	assertNoUnsupportedGo(t, out.Source)
	stepBody := bodyAfterMarker(out.Source, "func (ivy *paramlet) step()")
	if stepBody == "" {
		t.Fatalf("step body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{"ivy.saved =", "2", "3", "5"} {
		if !strings.Contains(stepBody, want) {
			t.Fatalf("generated parametric let body missing %q:\n%s", want, stepBody)
		}
	}
	if strings.Contains(stepBody, "f(") {
		t.Fatalf("generated parametric let should not leave f(...) call:\n%s", stepBody)
	}
}

func TestEmitMacroExpansionInApply(t *testing.T) {
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Cfg.IuCfg.UsePolymorphicMacros = true
	mod.Sig = goivy.NewSigOn(mod.Cfg.IuCfg)

	sort := &goivy.UninterpretedSort{Name: "nat"}
	leSort, err := goivy.NewFunctionSort(sort, sort, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	leSym := goivy.NewConst("<=", leSort)
	a := goivy.NewConst("a", sort)
	b := goivy.NewConst("b", sort)
	app := goivy.MustApply(leSym, a, b)

	if !goivy.IsMacro(app, mod.Cfg.IuCfg) {
		t.Fatalf("fixture <= should be a macro")
	}
	got, err := (&Generator{Mod: mod}).emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if !strings.Contains(got, "<") || !strings.Contains(got, "||") {
		t.Fatalf("expected expanded < / || expression, got: %s", got)
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

func TestEmitCastToNatClamps(t *testing.T) {
	g := newGoTestGeneratorWithInterps(map[string]string{"nat": "nat"})
	src := &goivy.UninterpretedSort{Name: "int"}
	nat := &goivy.UninterpretedSort{Name: "nat"}
	castSort, err := goivy.NewFunctionSort(src, nat)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	app := goivy.MustApply(goivy.NewConst("cast", castSort), goivy.NewConst("v", src))
	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{"__x := v", "if __x < 0", "return 0", "return __x"} {
		if !strings.Contains(got, want) {
			t.Fatalf("cast to nat should saturate; missing %q in:\n%s", want, got)
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

func TestGeneratedCastToNatCompilesAndRuns(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, `#lang ivy1.7
type signed
interpret signed -> int
type budget
interpret budget -> nat
parameter initial : signed
individual saved : budget
action check = {
    saved := cast(initial);
    assert saved = 0
}
export check
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "nat_cast", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{"__x := ivy.initial", "if __x < 0", "ivy.saved = func() int"} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("cast-to-nat generated source missing %q:\n%s", want, out.Source)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "-3", "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated nat cast: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
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
		`for I := 0; I < (1)+1; I++ {`,
		`fmt.Fprint(__ivy_out, ivy.marked[I])`,
		`fmt.Fprint(__ivy_out, "]")`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("quantified debug source missing %q:\n%s", want, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestIssue60DebugPrintOverSymbolicRangeUsesRuntimeBoundLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.8
type client_id

module iterable = {
    interpret this -> {0..max}
}

global {
    instance client_id : iterable
}

relation marked(C:client_id)
action step = {
}
export step
`)
	clientID, ok := mod.Sig.Sorts.Get2("client_id")
	if !ok {
		t.Fatal("missing client_id sort")
	}
	marked, err := mod.Sig.FindSymbol("marked", false)
	if err != nil {
		t.Fatalf("FindSymbol marked: %v", err)
	}
	c, err := goivy.NewVariable("C", clientID)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	markedApp, err := goivy.NewApply(goivy.NewConst("marked", marked.CSort), c)
	if err != nil {
		t.Fatalf("NewApply marked: %v", err)
	}
	dbg := goivy.NewDebugAction(goivy.NewConst(`"scan"`, goivy.TopS), markedApp)
	dbg.WithNames = []string{"seen"}
	mod.Actions.Set("step", dbg)
	out, err := Generate(mod, Config{Target: "test", ClassName: "issue60debug", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`fmt.Fprint(__ivy_out, "[")`,
		"__ivy_debug_idx0 := 0",
		"for C := 0; C < (ivy.client_id__max + 1); C++",
		"if __ivy_debug_idx0 > 0",
		"fmt.Fprint(__ivy_out, ivy.marked.Get(C))",
		"__ivy_debug_idx0++",
		`fmt.Fprint(__ivy_out, "]")`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("symbolic range debug source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "[]int{") {
		t.Fatalf("symbolic range debug should not expand to a literal slice:\n%s", out.Source)
	}
	compileGeneratedGo(t, out)
}

func TestIssue60ParameterizedObjectWithNestedModuleInstanceDoesNotPanic(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.8
type client_id
type msg_t

module net(msg_t) = {
    module socket = {
        action ping = {
        }
    }
}

object client(self:client_id) = {
    common {
        instance net_inst: net(msg_t)
        instance sock: net_inst.socket
    }
}

extract executable_runner = client
`)
	if mod == nil {
		t.Fatal("compileIvySource returned nil module")
	}
	out, err := Generate(mod, Config{Target: "test", ClassName: "issue60nested", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
}

func TestIssue60ParameterizedRangeBoundVisibleToTheoryProperty(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.8
type client_id

module iterable = {
    interpret this -> {0..max}
    property 0 <= X:this & X <= max
}

global {
    instance client_id : iterable
}
`)
	if mod == nil {
		t.Fatal("compileIvySource returned nil module")
	}
	found := false
	for _, p := range mod.Params {
		if p != nil && p.Name == "client_id.max" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("symbolic range bound parameter not retained: %#v", mod.Params)
	}
}

func TestIssue60NestedModuleFieldReferenceKeepsOuterArity(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "issue60_head.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.8
	type client_id
	type tag = {zero}

	object tcp = {
    type endpoint

    module net(msg_t) = {
		specification {
			relation sent(S:tcp.endpoint,D:tcp.endpoint,T:tag,M:msg_t)
			function head(S:tcp.endpoint,D:tcp.endpoint): tag
		}

        module socket = {
            parameter id:tcp.endpoint
            action send(dst:tcp.endpoint, msg:msg_t)

			specification {
				before send {
					net.sent(id,dst,net.head(id,dst),msg) := true;
				}
			}
        }
    }
	}

	object client(self:client_id) = {
		common {
			class pkt = {
				field kind: tag
			}

			instance net: tcp.net(pkt)
		}

		instance sock: net.socket
	}

extract executable_runner = client
`), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	if _, err := CompileAndGenerateAll(spec, map[string]string{"target": "test", "classname": "issue60head"}, Config{}); err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
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

func TestInitialConditionEqualitySwapsStateOnRight(t *testing.T) {
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
	eq, err := goivy.NewEq(goivy.NewConst("green", color), goivy.NewConst("saved", color))
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(eq, nil)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initright", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.saved = green") {
		t.Fatalf("swapped initial equality did not emit assignment:\n%s", out.Source)
	}
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

func TestInitialConditionRelationIffSwapsStateOnRight(t *testing.T) {
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
	init, err := goivy.NewIff(isGreen, markedApp)
	if err != nil {
		t.Fatalf("NewIff: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(init, nil)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initrightrel", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for _, C := range []color{red, green}",
		"__ivy_tmp0[C] = (C == green)",
		"ivy.marked[C] = __ivy_tmp0[C]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("swapped initial relation source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestInitialAxiomEqualityEmitsAssignment(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
axiom saved = green
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initaxiom", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.saved = green") {
		t.Fatalf("initial state axiom should constrain saved:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "unsupported initial condition") {
		t.Fatalf("initial axiom should not fall through to unsupported:\n%s", out.Source)
	}
}

func TestInitialStateRejectsParameterDependentInit(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter initial : color
individual saved : color
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
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitialAxiomRejectsParameterDependentFormula(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter initial : color
individual saved : color
axiom saved = initial
action step = {
}
export step
`)

	_, err := Generate(mod, Config{Target: "test", ClassName: "paramaxiombad", TestIters: "1"})
	if err == nil {
		t.Fatal("Generate should reject initial-state axiom depending on stripped parameter")
	}
	for _, want := range []string{
		`axiom`,
		`stripped parameter "initial"`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("initial axiom parameter error missing %q:\n%v", want, err)
		}
	}
}

func TestUnsupportedInitialConditionErrorIncludesSourceLine(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual saved : bool
action step = {
}
export step
`)
	native := &goivy.LogicNativeExpr{
		CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
	}
	native.SetLineno(goivy.Location{Filename: "init.ivy", Line: 27})
	mod.InitCond = goivy.FormulaToClauses(native, nil)

	_, err := Generate(mod, Config{Target: "test", ClassName: "badinit", TestIters: "1"})
	if err == nil {
		t.Fatal("Generate should reject unsupported initial condition")
	}
	for _, want := range []string{
		"init.ivy: line 27:",
		"unsupported initial condition",
		"unsupported initial formula *goivy.LogicNativeExpr",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("unsupported initial condition error missing %q:\n%v", want, err)
		}
	}
}

func TestInitialFiniteAxiomConstructsNestedExistentialWitness(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
type quorum
interpret node -> bv[2]
interpret quorum -> bv[2]

relation member(N:node,Q:quorum)
axiom forall Q1:quorum, Q2:quorum. exists N:node. member(N,Q1) & member(N,Q2)

action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initretry", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *initretry) __initState()")
	if body == "" {
		t.Fatalf("missing __initState:\n%s", out.Source)
	}
	for _, want := range []string{
		"for Q1 := 0; Q1 < 4; Q1++ {",
		"ivy.member[struct{ A0 int; A1 int }{0, Q1}] = true",
		"for Q2 := 0; Q2 < 4; Q2++ {",
		"ivy.member[struct{ A0 int; A1 int }{0, Q2}] = true",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("nested existential initial axiom source missing %q:\n%s", want, body)
		}
	}
	for _, bad := range []string{
		"for __ivy_init_attempt := 0; ; __ivy_init_attempt++ {",
		`ivyAssume(false, "initial condition")`,
		"unsupported initial axiom retry condition",
		"unsupported initial condition",
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("nested existential initial axiom should construct a witness, found %q:\n%s", bad, out.Source)
		}
	}
}

func TestInitialExistentialAxiomConstructsFiniteWitness(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx
interpret idx -> bv[5]

relation rel(X:idx,Y:idx)
axiom exists X:idx. forall Y:idx. rel(X,Y)

action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initexists", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *initexists) __initState()")
	if body == "" {
		t.Fatalf("missing __initState:\n%s", out.Source)
	}
	for _, want := range []string{
		"for Y := 0; Y < 32; Y++ {",
		"ivy.rel[struct{ A0 int; A1 int }{0, Y}] = true",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("finite existential initial axiom source missing %q:\n%s", want, body)
		}
	}
	for _, bad := range []string{
		"for __ivy_init_attempt := 0; ; __ivy_init_attempt++ {",
		`ivyAssume(false, "initial condition")`,
	} {
		if strings.Contains(body, bad) {
			t.Fatalf("finite existential initial axiom should construct a witness, found %q:\n%s", bad, body)
		}
	}
}

func TestInitialTotalOrderAxiomsConstructFiniteOrder(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type ts
interpret ts -> bv[4]

relation le(X:ts,Y:ts)
individual init_ts : ts
axiom le(X,X)
axiom le(X,Y) & le(Y,Z) -> le(X,Z)
axiom le(X,Y) & le(Y,X) -> X = Y
axiom le(X,Y) | le(Y,X)
axiom le(init_ts,T)

action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "inittotalorder", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *inittotalorder) __initState()")
	if body == "" {
		t.Fatalf("missing __initState:\n%s", out.Source)
	}
	for _, want := range []string{
		"ivy.init_ts = 0",
		"for X := 0; X < 16; X++ {",
		"for Y := 0; Y < 16; Y++ {",
		"__ivy_tmp0[struct{ A0 int; A1 int }{X, Y}] = (((X < Y)) || ((X == Y)))",
		"ivy.le[struct{ A0 int; A1 int }{X, Y}] = __ivy_tmp0[struct{ A0 int; A1 int }{X, Y}]",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("finite total-order initial axiom source missing %q:\n%s", want, body)
		}
	}
	for _, bad := range []string{
		"for __ivy_init_attempt := 0; ; __ivy_init_attempt++ {",
		`ivyAssume(false, "initial condition")`,
	} {
		if strings.Contains(body, bad) {
			t.Fatalf("finite total-order initial axioms should not retain retry guard %q:\n%s", bad, body)
		}
	}
}

func TestUnsupportedInitialAxiomRetryConditionErrorIncludesSourceLine(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	native := &goivy.LogicNativeExpr{
		CompiledChildren: []goivy.Expr{goivy.NewConst(`"x"`, goivy.TopS)},
	}
	native.SetLineno(goivy.Location{Filename: "axioms.ivy", Line: 33})
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.Cfg.AstCfg.NewLabeledFormula(nil, native))

	_, err := Generate(mod, Config{Target: "test", ClassName: "badretry", TestIters: "1"})
	if err == nil {
		t.Fatal("Generate should reject unsupported initial axiom retry condition")
	}
	for _, want := range []string{
		"axioms.ivy: line 33:",
		"unsupported initial axiom retry condition",
		"native C++ expression is not translated to Go",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("unsupported initial retry error missing %q:\n%v", want, err)
		}
	}
}

func TestInitialAxiomRelationIffEmitsLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
axiom forall C:color. marked(C) <-> C = green
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomrel", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for _, C := range []color{red, green}",
		"__ivy_tmp0[C] = (C == green)",
		"ivy.marked[C] = __ivy_tmp0[C]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("initial relation axiom source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "unsupported initial condition") {
		t.Fatalf("initial relation axiom should not fall through to unsupported:\n%s", out.Source)
	}
}

func TestInitialAxiomImplicationEmitsGuardedAssignment(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
axiom forall C:color. C = red -> marked(C)
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomimpl", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for _, C := range []color{red, green}",
		"__ivy_tmp0[C] = ivyTernary((C == red), true, ivy.marked[C])",
		"ivy.marked[C] = __ivy_tmp0[C]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("initial implication axiom source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "unsupported initial condition") {
		t.Fatalf("initial implication axiom should not fall through to unsupported:\n%s", out.Source)
	}
}

func TestInitialAxiomDisjunctionEmitsGuardedAssignment(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
axiom forall C:color. ~(C = red) | marked(C)
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomor", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for _, C := range []color{red, green}",
		"__ivy_tmp0[C] = ivyTernary((C == red), true, ivy.marked[C])",
		"ivy.marked[C] = __ivy_tmp0[C]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("initial disjunction axiom source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "unsupported initial condition") {
		t.Fatalf("initial disjunction axiom should not fall through to unsupported:\n%s", out.Source)
	}
}

func TestInitialAxiomPositiveDisjunctionEmitsGuardedAssignment(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
axiom forall C:color. C = red | marked(C)
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomposor", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for _, C := range []color{red, green}",
		"__ivy_tmp0[C] = ivyTernary(!((C == red)), true, ivy.marked[C])",
		"ivy.marked[C] = __ivy_tmp0[C]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("positive initial disjunction axiom source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "unsupported initial condition") {
		t.Fatalf("positive initial disjunction axiom should not fall through to unsupported:\n%s", out.Source)
	}
}

func TestInitialAxiomNaryDisjunctionEmitsGuardedAssignment(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green, blue}
relation marked(C:color)
axiom forall C:color. C = red | C = green | marked(C)
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomnaryor", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for _, C := range []color{red, green, blue}",
		"__ivy_tmp0[C] = ivyTernary(((!((C == red))) && (!((C == green)))), true, ivy.marked[C])",
		"ivy.marked[C] = __ivy_tmp0[C]",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("n-ary initial disjunction axiom source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "unsupported initial condition") {
		t.Fatalf("n-ary initial disjunction axiom should not fall through to unsupported:\n%s", out.Source)
	}
}

func TestInitialAxiomUsesDerivedDefinition(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
definition is_green(C:color) = C = green
axiom is_green(saved)
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomdef", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.saved = green") {
		t.Fatalf("derived initial axiom should constrain saved:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "unsupported initial condition") {
		t.Fatalf("derived initial axiom should not fall through to unsupported:\n%s", out.Source)
	}
}

func TestInitialAxiomEquivalentDerivedAssignmentDoesNotConflict(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
definition initial_color = green
axiom saved = initial_color
axiom saved = green
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomdefeq", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.saved = green") {
		t.Fatalf("equivalent derived initial axiom should constrain saved to green:\n%s", out.Source)
	}
}

func TestInitialAxiomConflictingDerivedAssignmentStillErrors(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
definition initial_color = green
axiom saved = initial_color
axiom saved = red
action step = {
}
export step
`)
	_, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomdefconflict", TestIters: "1"})
	if err == nil || !strings.Contains(err.Error(), "inconsistent") {
		t.Fatalf("expected inconsistent derived initial axiom error, got %v", err)
	}
}

func TestInitialAxiomTwoValueDisequalityEmitsForcedAssignment(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
axiom ~(saved = red)
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomneq", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.saved = green") {
		t.Fatalf("two-value initial axiom disequality should force saved to green:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "unsupported initial condition") {
		t.Fatalf("two-value initial axiom disequality should not fall through to unsupported:\n%s", out.Source)
	}
}

func TestInitialAxiomTwoValueRangeDisequalityEmitsForcedAssignment(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type bit = {0..1}
individual saved : bit
axiom ~(saved = 0)
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomrange", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "ivy.saved = 1") {
		t.Fatalf("two-value range disequality should force saved to 1:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "unsupported initial condition") {
		t.Fatalf("two-value range disequality should not fall through to unsupported:\n%s", out.Source)
	}
}

func TestInitialAxiomThreeValueDisequalityUsesSolverModelFast(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green, blue}
individual saved : color
axiom ~(saved = red)
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomneq3", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *initaxiomneq3) __initState()")
	if body == "" {
		t.Fatalf("missing __initState:\n%s", out.Source)
	}
	if !(strings.Contains(body, "ivy.saved = green") || strings.Contains(body, "ivy.saved = blue")) {
		t.Fatalf("three-value disequality should be solved to a concrete non-red model value:\n%s", body)
	}
	for _, bad := range []string{
		"ivy.saved = red",
		"for __ivy_init_attempt := 0; ; __ivy_init_attempt++ {",
		`ivyAssume(false, "initial condition")`,
		"unsupported initial condition",
	} {
		if strings.Contains(body, bad) {
			t.Fatalf("solver-backed initial disequality should not emit %q:\n%s", bad, body)
		}
	}
	compileGeneratedGo(t, out)
}

func TestTargetGenInitialAxiomThreeValueDisequalityUsesSolverModelFast(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green, blue}
individual saved : color
axiom ~(saved = red)
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "geninitneq3", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (ivy *geninitneq3) __initState()")
	if body == "" {
		t.Fatalf("missing __initState:\n%s", out.Source)
	}
	if !(strings.Contains(body, "ivy.saved = green") || strings.Contains(body, "ivy.saved = blue")) {
		t.Fatalf("target=gen three-value disequality should be solved to a concrete non-red model value:\n%s", body)
	}
	if strings.Contains(body, `ivy.saved = color(ivy.___ivy_choose`) || strings.Contains(body, "ivy.saved = red") {
		t.Fatalf("target=gen solver-backed initial disequality should not be random/red:\n%s", body)
	}
	compileGeneratedGo(t, out)
}

func TestInitialAxiomDisequalityConflictsWithEquality(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
axiom saved = red
axiom ~(saved = red)
action step = {
}
export step
`)
	_, err := Generate(mod, Config{Target: "test", ClassName: "initaxiomconflict", TestIters: "1"})
	if err == nil || !strings.Contains(err.Error(), "inconsistent") {
		t.Fatalf("expected inconsistent initial axiom error, got %v", err)
	}
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

func TestIssue73UnboundedInitializerUsesAssignmentFallback(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type key
relation pending(K:key)
`)
	key, ok := mod.Sig.Sorts.Get2("key")
	if !ok {
		t.Fatal("missing key sort")
	}
	pendingEntry, ok := mod.Sig.Symbols.Get2("pending")
	if !ok {
		t.Fatal("missing pending symbol")
	}
	pending := goivy.NewConst("pending", pendingEntry.Sort)
	k, err := goivy.NewVariable("K", key)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	assign := goivy.NewAssignAction(goivy.MustApply(pending, k), goivy.NewConst("false", goivy.Boolean))
	assign.SetFormalParams([]*goivy.Const{goivy.NewConst("K", key)})

	var w goWriter
	(&Generator{Mod: mod, ClassName: "issue73", Config: Config{Target: "test"}}).emitInitializerAction(&w, assign)
	got := w.String()
	if strings.Contains(got, "cannot enumerate initializer parameter") {
		t.Fatalf("unbounded initializer should fall through to assignment fallback:\n%s", got)
	}
	for _, want := range []string{
		"ivy.pending = newIvyThunkMap",
		"ivy.pending.base = func(",
		"return false",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("unbounded initializer fallback missing %q:\n%s", want, got)
		}
	}
}

func TestIssue73UnboundedInitializerConcreteFormalStaysUnsupported(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type key
relation pending(K:key)
`)
	key, ok := mod.Sig.Sorts.Get2("key")
	if !ok {
		t.Fatal("missing key sort")
	}
	pendingEntry, ok := mod.Sig.Symbols.Get2("pending")
	if !ok {
		t.Fatal("missing pending symbol")
	}
	pending := goivy.NewConst("pending", pendingEntry.Sort)
	formal := goivy.NewConst("prm:M", key)
	assign := goivy.NewAssignAction(goivy.MustApply(pending, formal), goivy.NewConst("false", goivy.Boolean))
	assign.SetFormalParams([]*goivy.Const{formal})
	assign.SetLineno(goivy.Location{Filename: "issue73.ivy", Line: 21})

	var w goWriter
	g := &Generator{Mod: mod, ClassName: "issue73", Config: Config{Target: "test"}}
	g.emitInitializerAction(&w, assign)
	if len(g.errs) == 0 {
		t.Fatalf("concrete unbounded initializer formal should stay unsupported:\n%s", w.String())
	}
	for _, want := range []string{
		"issue73.ivy: line 21:",
		"cannot enumerate initializer parameter prm:M",
	} {
		if !strings.Contains(g.errs[0].Error(), want) {
			t.Fatalf("concrete unbounded initializer error missing %q:\n%v", want, g.errs[0])
		}
	}
	if strings.Contains(w.String(), "ivy.pending.Set(prm__M") {
		t.Fatalf("concrete unbounded initializer formal must not be emitted unbound:\n%s", w.String())
	}
}

func TestTargetTestInitSkipsTemporalAxioms(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
temporal axiom [always_marked] globally marked(red)
after init {
    marked(red) := true
}
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "temporalinit", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate target=test with temporal axiom: %v", err)
	}
	if strings.Contains(out.Source, "LogicGlobally") ||
		strings.Contains(out.Source, "always_marked") ||
		strings.Contains(out.Source, "unsupported initial condition") {
		t.Fatalf("target=test init should not emit temporal axiom as an initial constraint:\n%s", out.Source)
	}
	if !strings.Contains(out.Source, "ivy.marked[red] = true") {
		t.Fatalf("after-init assignment should still be emitted:\n%s", out.Source)
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

func TestZeroArgDerivedDefinitionInlinesAsConst(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
relation flag
definition flag = true
individual saved : bool
action step = {
    saved := flag
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "derived_zero", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"ivy.saved = true",
		"func (ivy *derived_zero) step()",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("zero-arg derived definition source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		"flag bool",
		"ivy.flag",
		"func (ivy *derived_zero) flag()",
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("zero-arg derived definition should inline, found %q:\n%s", bad, out.Source)
		}
	}
}

const derivedSomeDefinitionSource = `#lang ivy1.7
type key = {k0, k1}
type idx = {i0, i1}
type value = {zero, one}
relation key_at(I:idx,K:key)
function value_at(I:idx) : value
function lookup(K:key) = some I. key_at(I,K) in value_at(I) else zero
individual saved : value
action step(k:key) = {
    saved := lookup(k)
}
export step
`

func TestDerivedDefinitionWithSomeElseSubstitutes(t *testing.T) {
	mod := compileIvySource(t, derivedSomeDefinitionSource)
	out, err := Generate(mod, Config{Target: "test", ClassName: "derived_some", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func() value {",
		"for _, I := range []idx{i0, i1} {",
		"if ivy.key_at[struct{ A0 idx; A1 key }{I, k}] {",
		"return ivy.value_at[I]",
		"return zero",
		"ivy.saved = func() value {",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("derived some definition source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestSomeWithElseUsesDerivedRelationBoundsForUninterpretedIndex(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx
type table
type key = {k0, k1}
type value = {zero, one}
function end(T:table) : idx
function value_at(T:table,I:idx) : value
relation key_at(T:table,I:idx,K:key) = 0 <= I & I < end(T)
function lookup(T:table,K:key) = some I. key_at(T,I,K) in value_at(T,I) else zero
individual saved : value
action step(t:table,k:key) = {
    saved := lookup(t,k)
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "some_index_bound", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for I := 0; I < ivy.end.Get(t); I++ {",
		"I < ivy.end.Get(t)",
		"return ivy.value_at.Get(struct{ A0 int; A1 int }{t, I})",
		"return zero",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("bounded some source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "cannot enumerate some variable I:idx") {
		t.Fatalf("bounded some should not fall through to finite enumeration:\n%s", out.Source)
	}
}

func TestGeneratedDerivedDefinitionWithSomeElseCompilesAndRuns(t *testing.T) {
	requireSlowTest(t)
	mod := compileIvySource(t, derivedSomeDefinitionSource)
	out, err := Generate(mod, Config{Target: "test", ClassName: "derived_some_run", TestIters: "1", TestRuns: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated derived some target=test: %v\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", err, stdout, stderr, out.Source)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("run output missing completion marker:\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
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

func TestIssue57InitRequireClosesFreeVariables(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type my_type_1
type my_type_2 = { val1, val2 }
type my_type_3

interpret my_type_1 -> bv[2]
interpret my_type_3 -> bv[8]

object node(type_1:my_type_1) = {
    relation voted(MY_TYPE_2:my_type_2, MY_TYPE_3:my_type_3)
    after init {
        require voted(MY_TYPE_2, MY_TYPE_3);
    }
}

extract executable_runner = node
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "issue57", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	initBody := bodyAfterMarker(out.Source, "func (ivy *issue57) __init()")
	if initBody == "" {
		t.Fatalf("__init body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		"for prm__V0 := 0; prm__V0 < 4; prm__V0++",
		"for _, MY_TYPE_2 := range []my_type_2{val1, val2}",
		"for MY_TYPE_3 := 0; MY_TYPE_3 < 256; MY_TYPE_3++",
		"ivyAssert((func() bool {",
		"if !(ivy.node__voted.Get(struct{ A0 int; A1 my_type_2; A2 int }{prm__V0, MY_TYPE_2, MY_TYPE_3})) {",
	} {
		if !strings.Contains(initBody, want) {
			t.Fatalf("issue57 init require source missing %q:\n%s", want, initBody)
		}
	}
	compileGeneratedGo(t, out)
}

func TestIssue60InterpretedRangeInitializerParamLoopUsesSymbolicBound(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.8
type client_id

module iterable = {
    interpret this -> {0..max}
}

global {
    instance client_id : iterable
}

object client(self:client_id) = {
    var ready: bool
    after init {
        ready := false;
    }
}
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "issue60init", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	initBody := bodyAfterMarker(out.Source, "func (ivy *issue60init) __init()")
	if initBody == "" {
		t.Fatalf("__init body not emitted:\n%s", out.Source)
	}
	if !strings.Contains(initBody, "for prm__V0 := 0; prm__V0 < (ivy.client_id__max + 1); prm__V0++") {
		t.Fatalf("interpreted range initializer loop missing symbolic bound:\n%s", initBody)
	}
	for _, want := range []string{
		"ivy.client__ready.Set(prm__V0, false)",
	} {
		if !strings.Contains(initBody, want) {
			t.Fatalf("interpreted range initializer source missing %q:\n%s", want, initBody)
		}
	}
	if strings.Contains(initBody, "[]int{") {
		t.Fatalf("symbolic interpreted range should not expand to a literal slice:\n%s", initBody)
	}
	compileGeneratedGo(t, out)
}

func TestParameterizedInitializerCallReturnUsesFormalLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.8
type client_id
type socket

module iterable = {
    interpret this -> {0..max}
}

global {
    instance client_id : iterable
}

object client(self:client_id) = {
    var sock: socket
    action open returns (s:socket) = {
        <<< impure
            s = 0;
        >>>
    }
    after init {
        sock := open
    }
}
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "paraminitcall", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	initBody := bodyAfterMarker(out.Source, "func (ivy *paraminitcall) __init()")
	if initBody == "" {
		t.Fatalf("__init body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		"for prm__V0 := 0; prm__V0 < (ivy.client_id__max + 1); prm__V0++",
		`ivy.___ivy_push("client.open")`,
		"__ivy_ret0 := ivy.client__open(prm__V0)",
		"ivy.___ivy_pop()",
		"ivy.client__sock.Set(prm__V0, __ivy_ret0)",
	} {
		if !strings.Contains(initBody, want) {
			t.Fatalf("parameterized initializer call source missing %q:\n%s", want, initBody)
		}
	}
}

func TestIssue60SymbolicRangeRandomizationUsesRuntimeBound(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.8
type client_id

module iterable = {
    interpret this -> {0..max}
}

global {
    instance client_id : iterable
}

individual last: client_id

action touch(c:client_id) = {
    last := c
}
export touch
`)
	testOut, err := Generate(mod, Config{Target: "test", ClassName: "issue60randomtest", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate target=test: %v\n%s", err, outSource(testOut))
	}
	for _, want := range []string{
		`ivy.last = (0 + ivy.___ivy_choose((ivy.client_id__max + 1), "init.last", 0))`,
		`gen.c = (0 + ivy.___ivy_randomize((ivy.client_id__max + 1), "touch.fml:c", 0))`,
		`__arg0 := touch_generator.c`,
	} {
		if !strings.Contains(testOut.Source, want) {
			t.Fatalf("target=test symbolic range randomization missing %q:\n%s", want, testOut.Source)
		}
	}

	genOut, err := Generate(mod, Config{Target: "gen", ClassName: "issue60randomgen", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate target=gen: %v\n%s", err, outSource(genOut))
	}
	for _, want := range []string{
		`ivy.last = (0 + ivy.___ivy_randomize((ivy.client_id__max + 1), "randomize.last", 0))`,
		`gen.c = (0 + ivy.___ivy_randomize((ivy.client_id__max + 1), "__fml:c", 0))`,
	} {
		if !strings.Contains(genOut.Source, want) {
			t.Fatalf("target=gen symbolic range randomization missing %q:\n%s", want, genOut.Source)
		}
	}

	testBin := compileGeneratedGo(t, testOut)
	stdout, stderr, err := runBinary(t, testBin, "2", "iters=1", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated target=test Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("generated target=test run did not complete\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	compileGeneratedGo(t, genOut)
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

func TestReplWritesVariantSupertypeReturn(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t
variant a of t
action make returns(out:t) = {
    var q:a;
    out := q
}
export make
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "variantout"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type t struct {",
		"func (v t) String() string {",
		`return fmt.Sprintf("{a:%v}", v.value.(int))`,
		"out = t{tag: 0, value: loc__q, valid: true}",
		"__ivy_result := ivy.make()",
		`fmt.Fprintf(__ivy_out, "= %s\n", ivyTraceValue(__ivy_result, false))`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("variant supertype return source missing %q:\n%s", want, out.Source)
		}
	}
}

func TestReplWritesStructVariantSupertypeReturn(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t
type color = {red, green}
variant req of t = struct {
    shade : color
}
action make returns(out:t) = {
    var r:req;
    r.shade := green;
    out := r
}
export make
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "variantstructout"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"type req struct {",
		"func (v req) String() string {",
		`return fmt.Sprintf("{shade:%v}", v.shade)`,
		"func (v t) String() string {",
		`return fmt.Sprintf("{req:%v}", v.value.(req))`,
		"loc__r.shade = green",
		"out = t{tag: 0, value: loc__r, valid: true}",
		`fmt.Fprintf(__ivy_out, "= %s\n", ivyTraceValue(__ivy_result, false))`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("struct variant supertype return source missing %q:\n%s", want, out.Source)
		}
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

func TestVariantEqualityUsesPayloadRecordEquality(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
type key
type cell
destructor shade(C:cell, K:key) : color
type msg
variant wrap of msg = struct {
    item : cell
}
individual saved : msg
individual saved2 : msg
action step = {
    var c1 : cell;
    var c2 : cell;
    var k : key;
    var w1 : wrap;
    var w2 : wrap;
    shade(c1,k) := green;
    shade(c2,k) := red;
    w1.item := c1;
    w2.item := c2;
    saved := w1;
    saved2 := w2;
    assert saved = saved2;
    assert saved *> w2
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "variant_record_eq"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"func (v cell) Equal(other cell) bool",
		"func (v wrap) Equal(other wrap) bool",
		"func (v msg) Equal(other msg) bool",
		"(v.item).Equal(other.item)",
		".Equal(ivy.saved2)",
		".Equal(loc__w2)",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("variant record equality source missing %q:\n%s", want, out.Source)
		}
	}
	for _, bad := range []string{
		"ivy.saved == ivy.saved2",
		"ivy.saved.value.(wrap) == loc__w2",
		"v.item == other.item",
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("variant record equality should not use Go struct == %q:\n%s", bad, out.Source)
		}
	}

	bin := compileGeneratedGo(t, out)
	commandFile := filepath.Join(t.TempDir(), "commands.in")
	if err := os.WriteFile(commandFile, []byte("step\n"), 0o644); err != nil {
		t.Fatalf("write step command file: %v", err)
	}
	stdout, stderr, err := runBinary(t, bin, commandFile)
	if err != nil {
		t.Fatalf("run generated variant record equality repl: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	if stdout != "> step\n" {
		t.Fatalf("variant record equality repl output differs\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
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

func TestIfSomeMixedFiniteThenBoundedNatUsesPerVariableBounds(t *testing.T) {
	src := `#lang ivy1.7
type ts
interpret ts -> bv[2]
type version
interpret version -> nat

relation ts_version(T:ts,V:version)
individual saved : ts

action pick(limit:version) = {
    if some (t:ts,v:version) ts_version(t,v) & v < limit {
        saved := t
    }
}
export pick
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "some_mixed_bounds", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"for loc__t := 0; loc__t < 4; loc__t++ {",
		"for loc__v := 0; loc__v < limit; loc__v++ {",
		"ivy.ts_version.Get(struct{ A0 int; A1 int }{loc__t, loc__v})",
		"saved = loc__t",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("mixed if-some source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "cannot enumerate some variable v:version") ||
		strings.Contains(out.Source, "unsupported if condition") {
		t.Fatalf("bounded later some variable should not fall through to unsupported:\n%s", out.Source)
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
	requireSlowTest(t)
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

func TestLocalWitnessFromLeadingAssumeUsesExtensionalRelation(t *testing.T) {
	src := `#lang ivy1.7
type node = {0..0}
type version = {0..2}
relation ts_version(N:node,V:version)
after init {
    ts_version(N,V) := false;
    ts_version(0,1) := true
}
action inner(n:node) = {
    var base:version;
    assume ts_version(n,base)
}
action step = {
    if exists BV:version. ts_version(0,BV) {
        call inner(0)
    }
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "local_witness", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	innerBody := bodyAfterMarker(out.Source, "func (ivy *local_witness) inner(")
	if innerBody == "" {
		t.Fatalf("inner method not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		"for __ivy_witness_key, __ivy_witness_val := range ivy.ts_version {",
		"if !__ivy_witness_val {",
		"(__ivy_witness_key.A0 == n)",
		"loc__base = __ivy_witness_key.A1",
		"ivyAssume(ivy.ts_version[struct{ A0 int; A1 int }{n, loc__base}]",
	} {
		if !strings.Contains(innerBody, want) {
			t.Fatalf("local witness source missing %q:\n%s", want, innerBody)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("local witness run failed\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", stdout, stderr, out.Source)
	}
	if strings.Contains(stdout, "assumption_failed") || strings.Contains(stderr, "assumption failed") {
		t.Fatalf("local witness should satisfy the leading assume\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestTargetTestCallActionAssumeRejectsCandidateWithoutFailure(t *testing.T) {
	src := `#lang ivy1.7
type idx = {0..1}
relation allowed(I:idx)
after init {
    allowed(I) := false;
    allowed(1) := true
}
action inner(i:idx) = {
    assume allowed(i)
}
action step(i:idx) = {
    call inner(i)
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "trial_call", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	mainBody := bodyAfterMarker(out.Source, "func main()")
	if mainBody == "" {
		t.Fatalf("main body not emitted:\n%s", out.Source)
	}
	for _, want := range []string{
		"__ivy_trial := ivy.__ivy_clone()",
		"__ivy_assume_rejecting = true",
		"if __ivy_trial_rejected {",
		"cycle--",
		"ivy = __ivy_trial",
	} {
		if !strings.Contains(mainBody, want) {
			t.Fatalf("trial call source missing %q:\n%s", want, mainBody)
		}
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "iters=3", "runs=1", "seed=1", "delay=0")
	if err != nil {
		t.Fatalf("trial call run failed\nstdout:\n%s\nstderr:\n%s\nsource:\n%s", stdout, stderr, out.Source)
	}
	if strings.Contains(stdout, "assumption_failed") || strings.Contains(stderr, "assumption failed") {
		t.Fatalf("trial call should reject bad candidates without assumption failure\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stdout, "test_completed") {
		t.Fatalf("trial call run missing completion marker\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestGeneratedChoicesUseCallStackQualifiedLabelsFast(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action helper = {
    var c : color;
    saved := c
}
action alpha = {
    call helper
}
action beta = {
    call helper
}
export alpha
export beta
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "stackchoice", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"__ivy_stack []string",
		"func (ivy *stackchoice) ___ivy_choice_label(name string, id int) string",
		"return strings.Join(parts, \".\")",
		"label := ivy.___ivy_choice_label(name, id)",
		"___ivy_push(\"alpha\")",
		"___ivy_push(\"beta\")",
		"ivy.___ivy_push(\"helper\")",
		"ivy.___ivy_pop()",
		`loc__c := color(ivy.___ivy_choose(0, "loc:c"`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("call-stack choice source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "_ = name") || strings.Contains(out.Source, "_ = id") {
		t.Fatalf("choice/randomize labels should not be discarded:\n%s", out.Source)
	}
}

func TestExtensionalRelationDetectedFromInitializer(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation r(X:node, Y:node)
individual c : node
after init {
    r(X, Y) := false;
}
action check = {
    assert forall X:node. r(X, c) -> false
}
`)
	ext := (&Generator{Mod: mod}).extensionalRelations()
	if !ext["r"] {
		t.Fatalf("expected r to be extensional, got %v", ext)
	}
	out, err := Generate(mod, Config{Target: "test", ClassName: "extinit", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "range ivy.r.overrides") {
		t.Fatalf("missing r overrides iteration in extensional quantifier:\n%s", out.Source)
	}
}

func TestNonExtensionalRelationDueToBadUpdate(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green, blue}
relation r(X:color)
after init {
    r(X) := false;
}
action set_all = {
    r(X) := true
}
export set_all
`)
	ext := (&Generator{Mod: mod}).extensionalRelations()
	if ext["r"] {
		t.Fatalf("expected r to not be extensional after variable-arg true update, got %v", ext)
	}
}

func TestUninitializedRelationNotExtensional(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation r(X:node)
individual c : node
action point_update = {
    r(c) := true
}
export point_update
`)
	ext := (&Generator{Mod: mod}).extensionalRelations()
	if ext["r"] {
		t.Fatalf("expected r to not be extensional without all-false init, got %v", ext)
	}
}

func TestExtensionalThroughDerivedDefinition(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation r(X:node)
relation s(X:node)
definition q(X:node) = r(X) & s(X)
after init {
    r(X) := false;
    s(X) := false;
}
action check = {
    assert exists X:node. q(X)
}
export check
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "extderived", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "range ivy.r.overrides") &&
		!strings.Contains(out.Source, "range ivy.s.overrides") {
		t.Fatalf("missing extensional overrides iteration after unfolding derived q:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "cannot enumerate quantified variable") {
		t.Fatalf("derived extensional quantifier should not fall through to finite enumeration:\n%s", out.Source)
	}
}

func TestExtensionalQuantifierMultipleVariables(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation r(X:node, Y:node)
relation p(X:node)
after init {
    r(X, Y) := false;
    p(X) := false;
}
action check = {
    assert exists X:node, Y:node. r(X, Y) & p(X) & p(Y)
}
export check
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "extmulti", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"range ivy.r.overrides",
		"X := __ivy_quant_key",
		".A0",
		"Y := __ivy_quant_key",
		".A1",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("multi-variable extensional quantifier source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "cannot enumerate quantified variable") {
		t.Fatalf("multi-variable extensional quantifier should not fall through to finite enumeration:\n%s", out.Source)
	}
}

func TestExtensionalRelationViaPolarityNegation(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation r(X:node)
relation p(X:node)
after init {
    r(X) := false;
    p(X) := false;
}
action check = {
    assert forall X:node. ~r(X) | p(X)
}
export check
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "extpol", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if !strings.Contains(out.Source, "range ivy.r.overrides") {
		t.Fatalf("missing extensional iteration for polarity-flipped forall:\n%s", out.Source)
	}
	if strings.Contains(out.Source, "cannot enumerate quantified variable") {
		t.Fatalf("polarity-flipped extensional quantifier should not fall through to finite enumeration:\n%s", out.Source)
	}
}

func TestExtensionalIfSomeBoundedByDerivedDefinition(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation r(X:node)
definition q(X:node) = r(X)
individual saved : node
after init {
    r(X) := false;
}
action pick = {
    if some x:node. q(x) {
        saved := x
    }
}
export pick
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "extifsomederived", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"range ivy.r.overrides",
		"loc__x := __ivy_some_key",
		"ivy.saved = loc__x",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("if-some derived extensional source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "unsupported if condition") {
		t.Fatalf("if-some derived extensional should not fall through to unsupported:\n%s", out.Source)
	}
}

func TestIssue60ExtensionalQuantifierRemainingSymbolicRangeUsesRuntimeBoundLoop(t *testing.T) {
	src := `#lang ivy1.8
type key
type client_id

module iterable = {
    interpret this -> {0..max}
}

global {
    instance client_id : iterable
}

relation seen(K:key)
after init {
    seen(K) := false;
    seen(3) := true
}
action check = {
}
export check
`
	mod := compileIvySource(t, src)
	keySort, ok := mod.Sig.Sorts.Get2("key")
	if !ok {
		t.Fatal("missing key sort")
	}
	clientID, ok := mod.Sig.Sorts.Get2("client_id")
	if !ok {
		t.Fatal("missing client_id sort")
	}
	seen, err := mod.Sig.FindSymbol("seen", false)
	if err != nil {
		t.Fatalf("FindSymbol seen: %v", err)
	}
	k, err := goivy.NewVariable("K", keySort)
	if err != nil {
		t.Fatalf("NewVariable K: %v", err)
	}
	c, err := goivy.NewVariable("C", clientID)
	if err != nil {
		t.Fatalf("NewVariable C: %v", err)
	}
	seenK, err := goivy.NewApply(goivy.NewConst("seen", seen.CSort), k)
	if err != nil {
		t.Fatalf("NewApply seen: %v", err)
	}
	cEq, err := goivy.NewEq(c, c)
	if err != nil {
		t.Fatalf("NewEq C: %v", err)
	}
	body, err := goivy.NewAnd(seenK, cEq)
	if err != nil {
		t.Fatalf("NewAnd: %v", err)
	}
	mod.Actions.Set("check", goivy.NewAssertAction(&goivy.LogicExists{
		Variables: []*goivy.LogicVariable{k, c},
		Body:      body,
	}))
	out, err := Generate(mod, Config{Target: "test", ClassName: "issue60quant_ext", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		"seen ivyThunkMap[int, bool]",
		"for __ivy_quant_key",
		"range ivy.seen.overrides",
		"K := __ivy_quant_key",
		"for C := 0; C < (ivy.client_id__max + 1); C++",
		"if ((ivy.seen.Get(K)) && ((C == C)))",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("extensional symbolic range quantifier source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "cannot enumerate quantified variable") || strings.Contains(out.Source, "[]int{") {
		t.Fatalf("extensional symbolic range quantifier should use bounded loops:\n%s", out.Source)
	}
	bin := compileGeneratedGo(t, out)
	stdout, stderr, err := runBinary(t, bin, "2", "iters=3", "runs=1", "seed=1")
	if err != nil {
		t.Fatalf("run generated Go: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}
	wantTrace := strings.Repeat("> check\n", 3) + "test_completed\n"
	if stdout != wantTrace {
		t.Fatalf("extensional symbolic range quantifier trace differs\nwant:\n%s\ngot:\n%s\nstderr:\n%s", wantTrace, stdout, stderr)
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

func TestQuantifierMultiVarInequalityFiltersSiblingBound(t *testing.T) {
	src := `#lang ivy1.7
type idx = {0..7}
relation marked(X:idx, Y:idx)
after init {
    marked(X,Y) := false
}
action check = {
    assert forall X:idx, Y:idx. X < Y -> marked(X,Y)
}
export check
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "quant_multi_bound", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	checkIdx := strings.Index(out.Source, "func (ivy *quant_multi_bound) check()")
	if checkIdx < 0 {
		t.Fatalf("missing check method:\n%s", out.Source)
	}
	checkSrc := out.Source[checkIdx:]
	for _, want := range []string{
		"for X := 0; X < 8; X++ {",
		"for Y := (X)+1; Y < 8; Y++ {",
	} {
		if !strings.Contains(checkSrc, want) {
			t.Fatalf("multi-variable quantifier should use ivy2cpp sibling-filtered bound %q:\n%s", want, checkSrc)
		}
	}
	if strings.Contains(checkSrc, "for X := 0; X < Y; X++") {
		t.Fatalf("outer quantified variable must not use quantified sibling Y as its upper bound:\n%s", checkSrc)
	}
}

func TestIfSomeMinUsesInequalityBound(t *testing.T) {
	src := `#lang ivy1.7
type idx = {0..10}
relation marked(X:idx)
individual saved : idx
after init {
    marked(X) := false;
    saved := 0
}
action pick = {
    if some x:idx. x < 5 & marked(x) minimizing x {
        saved := x
    }
}
export pick
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "some_min_bound", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	pickBody := bodyAfterMarker(out.Source, "func (ivy *some_min_bound) pick()")
	if pickBody == "" {
		t.Fatalf("missing generated pick body:\n%s", out.Source)
	}
	for _, want := range []string{
		"for loc__x := 0; loc__x < 5; loc__x++ {",
		"loc__x < 5",
		"ivy.marked[loc__x]",
		"break",
	} {
		if !strings.Contains(pickBody, want) {
			t.Fatalf("some_min inequality-bound source missing %q:\n%s", want, pickBody)
		}
	}
	if strings.Contains(pickBody, "for loc__x := 0; loc__x < 11; loc__x++") ||
		strings.Contains(pickBody, "[]int{0, 1, 2, 3, 4, 5") {
		t.Fatalf("some_min should use inequality-derived upper bound rather than full range:\n%s", pickBody)
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
	pickBody := bodyAfterMarker(out.Source, "func (ivy *some_min_range) pick()")
	if pickBody == "" {
		t.Fatalf("missing generated pick body:\n%s", out.Source)
	}
	if !strings.Contains(pickBody, "break") {
		t.Fatalf("first-param minimizing should emit the ivy2cpp early break:\n%s", pickBody)
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
	pickBody := bodyAfterMarker(out.Source, "func (ivy *some_max_range) pick()")
	if pickBody == "" {
		t.Fatalf("missing generated pick body:\n%s", out.Source)
	}
	if strings.Contains(pickBody, "break") {
		t.Fatalf("non-first-param maximizing should not emit the ivy2cpp early break:\n%s", pickBody)
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

func TestEmitExprSomeMixedFiniteThenBoundedNat(t *testing.T) {
	ts := &goivy.LogicEnumeratedSort{Name: "ts", Extension: []string{"t0", "t1"}}
	version := &goivy.UninterpretedSort{Name: "version"}
	tVar, err := goivy.NewVariable("T", ts)
	if err != nil {
		t.Fatalf("NewVariable T: %v", err)
	}
	vVar, err := goivy.NewVariable("V", version)
	if err != nil {
		t.Fatalf("NewVariable V: %v", err)
	}
	selfEq, err := goivy.NewEq(tVar, tVar)
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	ltSort, err := goivy.NewFunctionSort(version, version, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	bound, err := goivy.NewApply(goivy.NewConst("<", ltSort), vVar, goivy.NewConst("limit", version))
	if err != nil {
		t.Fatalf("NewApply <: %v", err)
	}
	body, err := goivy.NewAnd(selfEq, bound)
	if err != nil {
		t.Fatalf("NewAnd: %v", err)
	}
	got, err := newGoTestGeneratorWithInterps(map[string]string{"version": "nat"}).emitExpr(goivy.NewSome([]goivy.Expr{tVar, vVar}, body))
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{
		"func() ts {",
		"for _, T := range []ts{t0, t1} {",
		"for V := 0; V < limit; V++ {",
		"V < limit",
		"return T",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("mixed some expression missing %q:\n%s", want, got)
		}
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
			fixture := goOracleFixturePath(name)
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

var goOracleFixtureNames = []string{
	"empty",
	"basic_assign",
	"forall_assign",
	"bv_arithmetic",
	"enum_dispatch",
	"range_bounds",
	"destructor_record",
	"variant_simple",
	"variant_recursive",
	"hash_thunk_assign",
	"native_block",
	"callback_thunk",
	"progress_property",
	"isolate_two_parts",
}

func goOracleFixturePath(name string) string {
	return filepath.Join("..", "ivy2cpp", "test_vec", "oracle", name+".ivy")
}

func goOracleFixtureRejectsNative(name string) bool {
	return name == "native_block" || name == "callback_thunk"
}

func assertNativeFixtureGenerationRejection(t *testing.T, name string, err error) {
	t.Helper()
	if !goOracleFixtureRejectsNative(name) {
		return
	}
	if err == nil {
		t.Fatalf("expected native fixture %s to reject native C++ weakening", name)
	}
	if !strings.Contains(err.Error(), "native C++ action code is not translated to Go") {
		t.Fatalf("native fixture %s rejected with unexpected error: %v", name, err)
	}
}

func goOracleTesterArgsPath(name string) string {
	return filepath.Join("..", "ivy2cpp", "test_vec", "oracle", name+".test.args")
}

func readGoOracleTesterArgs(t *testing.T, name string) ([]string, bool) {
	t.Helper()
	data, err := os.ReadFile(goOracleTesterArgsPath(name))
	if err != nil {
		t.Fatalf("read oracle tester args for %s: %v", name, err)
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		t.Fatalf("oracle tester args for %s are empty", name)
	}
	fields := strings.Fields(text)
	if len(fields) > 0 && fields[0] == "SKIP" {
		return nil, true
	}
	return fields, false
}

func goOracleArgsRequestRun(args []string) bool {
	for _, arg := range args {
		if arg == "runs=0" {
			return false
		}
	}
	return true
}

func TestOracleTesterArgsFixtureCoverage(t *testing.T) {
	seen := map[string]bool{}
	for _, name := range goOracleFixtureNames {
		seen[name] = true
		if _, err := os.Stat(goOracleFixturePath(name)); err != nil {
			t.Fatalf("missing oracle fixture %s: %v", name, err)
		}
		_, _ = readGoOracleTesterArgs(t, name)
	}
	matches, err := filepath.Glob(filepath.Join("..", "ivy2cpp", "test_vec", "oracle", "*.test.args"))
	if err != nil {
		t.Fatalf("glob oracle tester args: %v", err)
	}
	for _, match := range matches {
		name := strings.TrimSuffix(filepath.Base(match), ".test.args")
		if !seen[name] {
			t.Fatalf("oracle tester args %s is not covered by goOracleFixtureNames", match)
		}
	}
}

func TestOracleGeneratedTesterArgsCompileAndRunSlow(t *testing.T) {
	requireSlowTest(t)
	for _, name := range goOracleFixtureNames {
		args, skip := readGoOracleTesterArgs(t, name)
		if skip {
			continue
		}
		t.Run(name, func(t *testing.T) {
			out, err := CompileAndGenerate(goOracleFixturePath(name), map[string]string{"target": "test", "classname": name}, Config{TestIters: "1"})
			if goOracleFixtureRejectsNative(name) {
				assertNativeFixtureGenerationRejection(t, name, err)
				return
			}
			if err != nil {
				t.Fatalf("CompileAndGenerate: %v\n%s", err, outSource(out))
			}
			assertNoUnsupportedGo(t, out.Source)
			bin := compileGeneratedGo(t, out)
			stdout, stderr, err := runBinary(t, bin, args...)
			if err != nil {
				t.Fatalf("run generated Go with oracle tester args %v: %v\nstdout:\n%s\nstderr:\n%s", args, err, stdout, stderr)
			}
			if goOracleArgsRequestRun(args) && !strings.Contains(stdout, "test_completed") {
				t.Fatalf("run output missing completion marker\nargs: %v\nstdout:\n%s\nstderr:\n%s", args, stdout, stderr)
			}
		})
	}
}

func TestOracleGeneratedTargetGenCompilesSlow(t *testing.T) {
	requireSlowTest(t)
	for _, name := range goOracleFixtureNames {
		t.Run(name, func(t *testing.T) {
			out, err := CompileAndGenerate(goOracleFixturePath(name), map[string]string{"target": "gen", "classname": name}, Config{TestIters: "1"})
			if goOracleFixtureRejectsNative(name) {
				assertNativeFixtureGenerationRejection(t, name, err)
				return
			}
			if err != nil {
				t.Fatalf("CompileAndGenerate target=gen: %v\n%s", err, outSource(out))
			}
			assertNoUnsupportedGo(t, out.Source)
			compileGeneratedGo(t, out)
		})
	}
}

func TestOracleGeneratedTargetReplCompilesSlow(t *testing.T) {
	requireSlowTest(t)
	for _, name := range goOracleFixtureNames {
		t.Run(name, func(t *testing.T) {
			out, err := CompileAndGenerate(goOracleFixturePath(name), map[string]string{"target": "repl", "classname": name}, Config{TestIters: "1"})
			if goOracleFixtureRejectsNative(name) {
				assertNativeFixtureGenerationRejection(t, name, err)
				return
			}
			if err != nil {
				t.Fatalf("CompileAndGenerate target=repl: %v\n%s", err, outSource(out))
			}
			assertNoUnsupportedGo(t, out.Source)
			compileGeneratedGo(t, out)
		})
	}
}

func TestOracleFixturesGenerateWithoutUnsupportedComments(t *testing.T) {
	for _, name := range goOracleFixtureNames {
		t.Run(name, func(t *testing.T) {
			fixture := goOracleFixturePath(name)
			batch, err := CompileAndGenerateAll(fixture, map[string]string{"target": "test", "classname": name}, Config{TestIters: "1"})
			if goOracleFixtureRejectsNative(name) {
				assertNativeFixtureGenerationRejection(t, name, err)
				return
			}
			if err != nil {
				t.Fatalf("CompileAndGenerateAll: %v", err)
			}
			if len(batch.Outputs) == 0 {
				t.Fatalf("expected at least one generated output for %s", name)
			}
			for _, out := range batch.Outputs {
				assertNoUnsupportedGo(t, out.Source)
			}
		})
	}
}

func TestOracleFixturesGenerateTargetGenWithoutUnsupportedComments(t *testing.T) {
	for _, name := range goOracleFixtureNames {
		t.Run(name, func(t *testing.T) {
			fixture := goOracleFixturePath(name)
			batch, err := CompileAndGenerateAll(fixture, map[string]string{"target": "gen", "classname": name}, Config{TestIters: "1"})
			if goOracleFixtureRejectsNative(name) {
				assertNativeFixtureGenerationRejection(t, name, err)
				return
			}
			if err != nil {
				t.Fatalf("CompileAndGenerateAll target=gen: %v", err)
			}
			if len(batch.Outputs) == 0 {
				t.Fatalf("expected at least one generated output for %s", name)
			}
			for _, out := range batch.Outputs {
				assertNoUnsupportedGo(t, out.Source)
			}
		})
	}
}

func TestOracleFixturesGenerateTargetReplWithoutUnsupportedComments(t *testing.T) {
	for _, name := range goOracleFixtureNames {
		t.Run(name, func(t *testing.T) {
			fixture := goOracleFixturePath(name)
			batch, err := CompileAndGenerateAll(fixture, map[string]string{"target": "repl", "classname": name}, Config{TestIters: "1"})
			if goOracleFixtureRejectsNative(name) {
				assertNativeFixtureGenerationRejection(t, name, err)
				return
			}
			if err != nil {
				t.Fatalf("CompileAndGenerateAll target=repl: %v", err)
			}
			if len(batch.Outputs) == 0 {
				t.Fatalf("expected at least one generated output for %s", name)
			}
			for _, out := range batch.Outputs {
				assertNoUnsupportedGo(t, out.Source)
			}
		})
	}
}

func TestLocalPublicFixturesGenerateTargetTest(t *testing.T) {
	for _, fixture := range []string{
		filepath.Join("..", "examples", "client_server_example.ivy"),
		filepath.Join("..", "end2end", "data", "client_server.ivy"),
		filepath.Join("..", "end2end", "data", "enum_types.ivy"),
		filepath.Join("..", "end2end", "data", "relation_sort.ivy"),
		filepath.Join("..", "end2end", "data", "trivial_pass.ivy"),
		filepath.Join("..", "test_vectors", "bmc_method.ivy"),
		filepath.Join("..", "test_vectors", "bmc_minimal.ivy"),
	} {
		fixture := fixture
		name := strings.TrimSuffix(filepath.Base(fixture), ".ivy")
		t.Run(name, func(t *testing.T) {
			batch, err := CompileAndGenerateAll(fixture, map[string]string{"target": "test", "classname": name}, Config{TestIters: "1"})
			if err != nil {
				t.Fatalf("CompileAndGenerateAll: %v", err)
			}
			if len(batch.Outputs) == 0 {
				t.Fatalf("expected at least one generated output for %s", fixture)
			}
			for _, out := range batch.Outputs {
				assertNoUnsupportedGo(t, out.Source)
				for _, warning := range out.Warnings {
					if !strings.Contains(warning, "using zero value for non-enumerable sort") {
						t.Fatalf("unexpected warning for %s: %s", fixture, warning)
					}
				}
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

func TestParseArgsEdgeCasesMatchIvy2Cpp(t *testing.T) {
	cases := [][]string{
		{"target=repl", "x.ivy"},
		{"target=test", "target=gen", "x.ivy"},
		{"target=", "x.ivy"},
		{"=bad", "x.ivy"},
		{"target", "x.ivy"},
		{"target=test", "x.txt"},
		{"target=test", "x.ivy", "other.ivy"},
		{"x.ivy", "target=test"},
		nil,
	}
	normalizeErr := func(err error) string {
		if err == nil {
			return ""
		}
		msg := strings.ReplaceAll(err.Error(), "ivy2golang", "ivy2tool")
		msg = strings.ReplaceAll(msg, "ivy2cpp", "ivy2tool")
		return msg
	}
	for _, args := range cases {
		args := args
		t.Run(fmt.Sprintf("%q", args), func(t *testing.T) {
			goParams, goFile, goErr := ParseArgs(args)
			cppParams, cppFile, cppErr := ivy2cppgen.ParseArgs(args)
			if normalizeErr(goErr) != normalizeErr(cppErr) {
				t.Fatalf("ParseArgs(%q) error mismatch\ngolang: %v\ncpp:    %v", args, goErr, cppErr)
			}
			if goErr != nil || cppErr != nil {
				return
			}
			if goFile != cppFile {
				t.Fatalf("ParseArgs(%q) file=%q, want ivy2cpp %q", args, goFile, cppFile)
			}
			if !reflect.DeepEqual(goParams, cppParams) {
				t.Fatalf("ParseArgs(%q) params=%v, want ivy2cpp %v", args, goParams, cppParams)
			}
		})
	}
}

func TestBuildRequestedBoolSurfaceMatchesIvy2Cpp(t *testing.T) {
	if BuildRequested(nil) != ivy2cppgen.BuildRequested(nil) {
		t.Fatalf("nil BuildRequested mismatch")
	}
	for _, value := range []string{
		"",
		"true",
		"TRUE",
		"1",
		"yes",
		" yes ",
		"false",
		"0",
		"no",
		"on",
	} {
		params := map[string]string{"build": value}
		got := BuildRequested(params)
		want := ivy2cppgen.BuildRequested(params)
		if got != want {
			t.Fatalf("BuildRequested(build=%q)=%v, want ivy2cpp %v", value, got, want)
		}
		cfg, _, err := mergeParams(params, Config{})
		if err != nil {
			t.Fatalf("mergeParams(build=%q): %v", value, err)
		}
		if cfg.Build != want {
			t.Fatalf("mergeParams(build=%q).Build=%v, want %v", value, cfg.Build, want)
		}
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

func TestCompileAndGenerateAllCarriesBuildFlag(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "buildflag.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
action step = {}
export step
`), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		value string
		want  bool
	}{
		{"true", true},
		{"1", true},
		{"yes", true},
		{"false", false},
		{"0", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			batch, err := CompileAndGenerateAll(spec, map[string]string{"target": "test", "classname": "BuildFlag", "build": tc.value}, Config{})
			if err != nil {
				t.Fatalf("CompileAndGenerateAll(build=%q): %v", tc.value, err)
			}
			if batch.Config.Build != tc.want {
				t.Fatalf("batch.Config.Build for build=%q = %v, want %v", tc.value, batch.Config.Build, tc.want)
			}
			if len(batch.Outputs) != 1 {
				t.Fatalf("outputs=%d, want 1", len(batch.Outputs))
			}
		})
	}
}

func TestWriteBatchOutputUsesBuildSubdirLikeIvy2Cpp(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.Mkdir("build", 0o755); err != nil {
		t.Fatalf("create cwd build dir: %v", err)
	}
	spec := filepath.Join(dir, "layout.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
action step = {}
export step
`), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(dir, "out")
	batch, err := CompileAndGenerateAll(spec, map[string]string{"target": "test", "classname": "Layout"}, Config{})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	if err := WriteBatchOutput(batch, outDir); err != nil {
		t.Fatalf("WriteBatchOutput: %v", err)
	}
	sourcePath := filepath.Join(outDir, "build", "Layout.go")
	if _, err := os.Stat(sourcePath); err != nil {
		t.Fatalf("generated Go should be written under outdir/build when cwd/build exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "Layout.go")); !os.IsNotExist(err) {
		t.Fatalf("generated Go should not be written directly under outdir when cwd/build exists; stat err=%v", err)
	}
	descPath := filepath.Join(outDir, "Layout.dsc")
	desc, err := os.ReadFile(descPath)
	if err != nil {
		t.Fatalf("descriptor should still be written at outdir root: %v", err)
	}
	if !strings.Contains(string(desc), `"test_params":["iters","runs","seed","delay","wait","modelfile"]`) {
		t.Fatalf("descriptor missing test params:\n%s", desc)
	}
}

func TestMergeParamsRejectsUnknownTargetCompilerAndParameter(t *testing.T) {
	for _, params := range []map[string]string{
		{"target": "bad"},
		{"compiler": "go"},
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
		`testIters := ivyAtoiOption(opts, "iters", 7)`,
		`runs := ivyAtoiOption(opts, "runs", 3)`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("missing %q:\n%s", want, out.Source)
		}
	}
	compileGeneratedGo(t, out)
}

func TestMainNameNormalizesToGoIdentifier(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "WithMain", MainName: "type.runner[0]", TestIters: "1"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if out.Config.MainName != "type__runner__0__" {
		t.Fatalf("main name was not normalized: %q", out.Config.MainName)
	}
	for _, want := range []string{
		"func type__runner__0__()",
		"func main() {",
		"type__runner__0__()",
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("normalized main source missing %q:\n%s", want, out.Source)
		}
	}
	if strings.Contains(out.Source, "type.runner") || strings.Contains(out.Source, "runner[0]") {
		t.Fatalf("raw main name leaked into generated source:\n%s", out.Source)
	}
}

func TestCompileAndGenerateAllWritesDescriptorForReplAndTest(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "desc.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
action step = {}
export step
`), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"repl", "test"} {
		t.Run(target, func(t *testing.T) {
			batch, err := CompileAndGenerateAll(spec, map[string]string{"target": target, "classname": "Desc"}, Config{})
			if err != nil {
				t.Fatalf("CompileAndGenerateAll: %v", err)
			}
			if len(batch.Outputs) != 1 {
				t.Fatalf("outputs=%d", len(batch.Outputs))
			}
			raw := batch.ExtraFiles["Desc.dsc"]
			if raw == "" {
				t.Fatalf("missing descriptor in %#v", batch.ExtraFiles)
			}
			var desc map[string]interface{}
			if err := json.Unmarshal([]byte(raw), &desc); err != nil {
				t.Fatalf("descriptor is not JSON: %v\n%s", err, raw)
			}
			if _, ok := desc["processes"].([]interface{}); !ok {
				t.Fatalf("descriptor missing processes: %#v", desc)
			}
			if target == "test" {
				if _, ok := desc["test_params"].([]interface{}); !ok {
					t.Fatalf("test descriptor missing test_params: %#v", desc)
				}
			} else if _, ok := desc["test_params"]; ok {
				t.Fatalf("repl descriptor unexpectedly has test_params: %#v", desc)
			}
		})
	}
}

func TestDescriptorProcessParamsMatchIvy2Cpp(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "desc_params.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
type color = {red, green}
type idx = {0..3}
parameter pick : color
parameter limit : idx
action step = {}
export step
`), 0o644); err != nil {
		t.Fatal(err)
	}
	params := map[string]string{"target": "test", "classname": "ParamDesc"}
	goBatch, err := CompileAndGenerateAll(spec, params, Config{})
	if err != nil {
		t.Fatalf("ivy2golang CompileAndGenerateAll: %v", err)
	}
	cppBatch, err := ivy2cppgen.CompileAndGenerateAll(spec, params, ivy2cppgen.Config{})
	if err != nil {
		t.Fatalf("ivy2cpp CompileAndGenerateAll: %v", err)
	}
	goRaw := goBatch.ExtraFiles["ParamDesc.dsc"]
	cppRaw := cppBatch.ExtraFiles["ParamDesc.dsc"]
	if goRaw == "" || cppRaw == "" {
		t.Fatalf("missing descriptor\ngo extra: %#v\ncpp extra: %#v", goBatch.ExtraFiles, cppBatch.ExtraFiles)
	}
	var goDesc any
	if err := json.Unmarshal([]byte(goRaw), &goDesc); err != nil {
		t.Fatalf("ivy2golang descriptor is not JSON: %v\n%s", err, goRaw)
	}
	var cppDesc any
	if err := json.Unmarshal([]byte(cppRaw), &cppDesc); err != nil {
		t.Fatalf("ivy2cpp descriptor is not JSON: %v\n%s", err, cppRaw)
	}
	goNorm, err := json.Marshal(goDesc)
	if err != nil {
		t.Fatalf("normalize ivy2golang descriptor: %v", err)
	}
	cppNorm, err := json.Marshal(cppDesc)
	if err != nil {
		t.Fatalf("normalize ivy2cpp descriptor: %v", err)
	}
	if string(goNorm) != string(cppNorm) {
		t.Fatalf("descriptor differs from ivy2cpp\ngolang: %s\ncpp:    %s", goNorm, cppNorm)
	}
}

func TestCompileAndGenerateAllIsolateAllEmitsSortedOutputsAndDescriptor(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "isoparts.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
type value = {zero, one}
object client = {
    individual last : value
    action send(v:value) = {
        last := v
    }
}
object server = {
    individual seen : value
    action recv(v:value) = {
        seen := v
    }
}
export client.send
export server.recv
isolate iso_client = client with server
isolate iso_server = server with client
extract iso_impl = client,server
`), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	batch, err := CompileAndGenerateAll(spec, map[string]string{"target": "test", "isolate": "all", "classname": "isoparts"}, Config{TestIters: "1"})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	wantBases := []string{"isoparts_iso_client", "isoparts_iso_impl", "isoparts_iso_server", "isoparts_this"}
	if len(batch.Outputs) != len(wantBases) {
		t.Fatalf("outputs=%d, want %d", len(batch.Outputs), len(wantBases))
	}
	for i, want := range wantBases {
		if got := batch.Outputs[i].BaseName; got != want {
			t.Fatalf("output[%d].BaseName=%q, want %q", i, got, want)
		}
	}
	raw := batch.ExtraFiles["isoparts.dsc"]
	if raw == "" {
		t.Fatalf("missing descriptor in %#v", batch.ExtraFiles)
	}
	var desc struct {
		Processes []struct {
			Binary string `json:"binary"`
			Name   string `json:"name"`
		} `json:"processes"`
		TestParams []string `json:"test_params"`
	}
	if err := json.Unmarshal([]byte(raw), &desc); err != nil {
		t.Fatalf("descriptor is not JSON: %v\n%s", err, raw)
	}
	if len(desc.Processes) != len(wantBases) {
		t.Fatalf("descriptor processes=%d, want %d: %s", len(desc.Processes), len(wantBases), raw)
	}
	for i, want := range []struct {
		binary string
		name   string
	}{
		{"isoparts_iso_client", "iso_client"},
		{"isoparts_iso_impl", "iso_impl"},
		{"isoparts_iso_server", "iso_server"},
		{"isoparts_this", "this"},
	} {
		if desc.Processes[i].Binary != want.binary || desc.Processes[i].Name != want.name {
			t.Fatalf("process[%d]=%+v, want %+v in %s", i, desc.Processes[i], want, raw)
		}
	}
	if len(desc.TestParams) == 0 {
		t.Fatalf("test descriptor missing test_params: %s", raw)
	}
}

func TestCompileAndGenerateAllPrunesFilteredStateStores(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "pruned.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
type small = {0..3}
individual x : small
after init {
    x := 0
}
action set = {
    x := 1
}
export set
`), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	batch, err := CompileAndGenerateAll(spec, map[string]string{"target": "impl", "classname": "Pruned"}, Config{})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	if len(batch.Outputs) != 1 {
		t.Fatalf("outputs=%d, want 1", len(batch.Outputs))
	}
	out := batch.Outputs[0]
	structBody := bodyAfterMarker(out.Source, "type Pruned struct")
	if structBody == "" {
		t.Fatalf("Pruned struct not emitted:\n%s", out.Source)
	}
	for _, bad := range []string{
		"\tx ",
	} {
		if strings.Contains(structBody, bad) {
			t.Fatalf("filtered state x should not be resurrected in Pruned struct; found %q:\n%s", bad, structBody)
		}
	}
	for _, bad := range []string{
		"ivy.x =",
		"> x",
		`"init.x"`,
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("filtered state x should not be resurrected after CreateIsolate; found %q:\n%s", bad, out.Source)
		}
	}
}

func TestRestoreSortForGoInterfaceInitializesInterpMap(t *testing.T) {
	sort := &goivy.UninterpretedSort{Name: "idx"}
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Sig = goivy.NewSigOn(mod.Cfg.IuCfg)
	mod.Sig.Interp = nil
	mod.SortDestructors = nil
	mod.DestructorSorts = nil
	field := goivy.NewConst("field", goivy.Boolean)
	snap := goInterfaceSnapshot{
		Sorts:           map[string]goivy.Sort{"idx": sort},
		Interps:         map[string]interface{}{"idx": "int"},
		SortDestructors: map[string][]*goivy.Const{"idx": {field}},
		DestructorSorts: map[string]goivy.Sort{"idx": sort},
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("restoreSortForGoInterface panicked with nil Interp map: %v", r)
		}
	}()
	restoreSortForGoInterface(mod, snap, sort, map[string]bool{})
	if got := mod.Sig.Interp["idx"]; got != "int" {
		t.Fatalf("restored interpretation = %#v, want int", got)
	}
	if got, ok := mod.SortDestructors.Get2("idx"); !ok || len(got) != 1 || got[0] != field {
		t.Fatalf("restored destructors = %#v, %v; want field", got, ok)
	}
	if got := mod.DestructorSorts["idx"]; got != sort {
		t.Fatalf("restored destructor sort = %#v, want idx sort", got)
	}
}

func TestTargetImplDoesNotEmitGeneratedMain(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "impl", ClassName: "ImplOnly"})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	if out.Target != "impl" || out.EffectiveTarget != "impl" || !out.EmitMain {
		t.Fatalf("impl target metadata = target %q effective %q emitMain %v", out.Target, out.EffectiveTarget, out.EmitMain)
	}
	if !strings.Contains(out.Source, "package main") {
		t.Fatalf("impl target should keep main package shape:\n%s", out.Source)
	}
	for _, bad := range []string{
		"func main()",
		"func ivy_main()",
		"func main() {",
	} {
		if strings.Contains(out.Source, bad) {
			t.Fatalf("impl target should not emit generated driver %q:\n%s", bad, out.Source)
		}
	}
}

func TestCompileAndGenerateAllClassDoesNotWriteDescriptor(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "class.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
action step = {
}
export step
`), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	batch, err := CompileAndGenerateAll(spec, map[string]string{"target": "class", "classname": "Classy"}, Config{})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	if len(batch.ExtraFiles) != 0 {
		t.Fatalf("class target should not write descriptors: %#v", batch.ExtraFiles)
	}
	if len(batch.Outputs) != 1 || batch.Outputs[0].Target != "class" || batch.Outputs[0].EmitMain {
		t.Fatalf("unexpected class output metadata: %#v", batch.Outputs)
	}
}

func outSource(out *Output) string {
	if out == nil {
		return "<nil>"
	}
	return out.Source
}

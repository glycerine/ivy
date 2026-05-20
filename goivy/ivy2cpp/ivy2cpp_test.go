package ivy2cpp

import (
	"os"
	"os/exec"
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

func normalizeCPP(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func runPythonIvyToCpp(t *testing.T, src string, params ...string) (string, string) {
	t.Helper()
	root := repoRoot(t)
	pyivyRoot := filepath.Join(root, "pyivy", "ivy")
	runner := filepath.Join(root, "pyivy", "goivy-venv", "bin", "ivy_to_cpp")
	args := append([]string{}, params...)
	if _, err := os.Stat(runner); err != nil {
		python := filepath.Join(root, "pyivy", "goivy-venv", "bin", "python3")
		if _, statErr := os.Stat(python); statErr != nil {
			found, lookErr := exec.LookPath("python3")
			if lookErr != nil {
				t.Skip("python3 not available")
			}
			python = found
		}
		runner = python
		args = append([]string{"-c", "from ivy.ivy_to_cpp import main; main()"}, params...)
	}
	dir := t.TempDir()
	spec := filepath.Join(dir, "oracle.ivy")
	if err := os.WriteFile(spec, []byte(src), 0o644); err != nil {
		t.Fatalf("write oracle spec: %v", err)
	}
	args = append(args, filepath.Base(spec))
	cmd := exec.Command(runner, args...)
	cmd.Dir = dir
	pythonPath := pyivyRoot
	if existing := os.Getenv("PYTHONPATH"); existing != "" {
		pythonPath += string(os.PathListSeparator) + existing
	}
	cmd.Env = append(os.Environ(), "IVY_HOME="+pyivyRoot, "PYTHONPATH="+pythonPath, "XTRACE_OFF=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("python ivy_to_cpp failed: %v\n%s", err, out)
	}
	h, err := os.ReadFile(filepath.Join(dir, "oracle.h"))
	if err != nil {
		t.Fatalf("read python header: %v", err)
	}
	cpp, err := os.ReadFile(filepath.Join(dir, "oracle.cpp"))
	if err != nil {
		t.Fatalf("read python impl: %v", err)
	}
	return string(h), string(cpp)
}

func compileGeneratedCPP(t *testing.T, out *Output) {
	t.Helper()
	cxx, err := cxxCompiler()
	if err != nil {
		if isMissingZ3ToolchainError(err) {
			t.Skip(err.Error())
		}
		t.Fatalf("C++ compiler lookup: %v", err)
	}
	dir := t.TempDir()
	if err := WriteOutput(out, dir); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	args := []string{"-std=c++11"}
	if outputUsesZ3(out) {
		includeArgs, _, err := z3BuildArgs()
		if err != nil {
			if isMissingZ3ToolchainError(err) {
				t.Skip(err.Error())
			}
			t.Fatalf("Z3 build args: %v", err)
		}
		args = append(args, includeArgs...)
	}
	args = append(args, "-c", filepath.Join(dir, out.BaseName+".cpp"), "-o", filepath.Join(dir, out.BaseName+".o"))
	cmd := exec.Command(cxx, args...)
	if buf, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile generated C++: %v\n%s\nheader:\n%s\nimpl:\n%s", err, buf, out.Header, out.Impl)
	}
}

func assertNoUnsupportedCPP(t *testing.T, out *Output) {
	t.Helper()
	if strings.Contains(out.Header, "unsupported") || strings.Contains(out.Impl, "unsupported") {
		t.Fatalf("generated output contains unsupported marker:\nheader:\n%s\nimpl:\n%s", out.Header, out.Impl)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

func TestGenerateEmptyModuleProducesHeaderAndImpl(t *testing.T) {
	mod := goivy.New()
	mod.Name = "empty"
	out, err := Generate(mod, Config{})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out.BaseName != "empty" || out.ClassName != "empty" {
		t.Fatalf("unexpected names: %+v", out)
	}
	for _, want := range []string{"#pragma once", "class empty", "void __init();"} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("header missing %q:\n%s", want, out.Header)
		}
	}
	if !strings.Contains(out.Impl, `#include "empty.h"`) {
		t.Fatalf("impl missing include:\n%s", out.Impl)
	}
}

func TestPythonOracleMinimalReplShape(t *testing.T) {
	src := `#lang ivy1.7
action step = {
}
export step
`
	pyHeader, pyImpl := runPythonIvyToCpp(t, src, "target=repl")
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"class oracle", "__init", "step"} {
		if !strings.Contains(pyHeader+pyImpl, want) {
			t.Fatalf("python oracle missing %q:\nheader:\n%s\nimpl:\n%s", want, pyHeader, pyImpl)
		}
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("go output missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
}

func TestPythonOracleGoPortSharedShape(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`
	fixtures := []struct {
		name   string
		target string
		want   []string
	}{
		{
			name:   "repl",
			target: "repl",
			want:   []string{"class oracle", "enum color", "__init", "set", "main"},
		},
		{
			name:   "test",
			target: "test",
			want:   []string{`#include "z3++.h"`, "class gen", "__from_solver", "randomize", "enum color"},
		},
		{
			name:   "gen",
			target: "gen",
			want:   []string{`#include "z3++.h"`, "class gen", "__from_solver", "randomize", "enum color"},
		},
	}
	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			pyHeader, pyImpl := runPythonIvyToCpp(t, src, "target="+tc.target)
			mod := compileIvySource(t, src)
			out, err := Generate(mod, Config{Target: tc.target, ClassName: "oracle"})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			py := normalizeCPP(pyHeader + "\n" + pyImpl)
			goOut := normalizeCPP(out.Header + "\n" + out.Impl)
			for _, want := range tc.want {
				if !strings.Contains(py, want) {
					t.Fatalf("python output missing %q:\nheader:\n%s\nimpl:\n%s", want, pyHeader, pyImpl)
				}
				if !strings.Contains(goOut, want) {
					t.Fatalf("go output missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
				}
			}
		})
	}
}

func TestPythonAndGoGeneratedFixturesCompile(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`
	for _, target := range []string{"repl", "test"} {
		t.Run(target, func(t *testing.T) {
			pyHeader, pyImpl := runPythonIvyToCpp(t, src, "target="+target)
			compileGeneratedCPP(t, &Output{
				Header:    pyHeader,
				Impl:      pyImpl,
				BaseName:  "oracle",
				ClassName: "oracle",
			})

			mod := compileIvySource(t, src)
			goOut, err := Generate(mod, Config{Target: target, ClassName: "oracle"})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			compileGeneratedCPP(t, goOut)
		})
	}
}

func TestGeneratedOutputContainsNoUnsupportedComments(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "clean"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertNoUnsupportedCPP(t, out)
}

func TestCompileSmokeFixturesTable(t *testing.T) {
	fixtures := []struct {
		name   string
		target string
		src    string
		want   []string
	}{
		{
			name:   "empty",
			target: "impl",
			src: `#lang ivy1.7
action step = {
}
`,
			want: []string{"class empty", "void step()"},
		},
		{
			name:   "enum relation init",
			target: "impl",
			src: `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
after init {
    marked(C) := false
}
`,
			want: []string{"enum color", "std::map<color,bool> marked;", "marked[C] = false;"},
		},
		{
			name:   "small repl",
			target: "repl",
			src: `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`,
			want: []string{`if (action == "set")`, "small_repl ivy;"},
		},
	}
	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			mod := compileIvySource(t, tc.src)
			className := strings.ReplaceAll(tc.name, " ", "_")
			out, err := Generate(mod, Config{Target: tc.target, ClassName: className})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			assertNoUnsupportedCPP(t, out)
			for _, want := range tc.want {
				if !strings.Contains(out.Header+out.Impl, want) {
					t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
				}
			}
			compileGeneratedCPP(t, out)
		})
	}
}

func TestCLIParamParsingAcceptsIvyStyleKeyValues(t *testing.T) {
	params, file, err := ParseArgs([]string{"target=repl", "classname=MyIvy", "outdir=build", "x.ivy"})
	if err != nil {
		t.Fatalf("ParseArgs: %v", err)
	}
	if file != "x.ivy" || params["target"] != "repl" || params["classname"] != "MyIvy" || params["outdir"] != "build" {
		t.Fatalf("unexpected parse result params=%v file=%s", params, file)
	}
}

func TestVarNameMatchesPythonCases(t *testing.T) {
	cases := map[string]string{
		"<":              "__lt",
		"ext:step":       "ext__step",
		"loc:x":          "loc__x",
		"prm:y":          "prm__y",
		"fml:z":          "z",
		"a.b[c]":         "a__b__c__",
		"foo:bar":        "foo__COLON__bar",
		"thing@@member":  "thing.member",
		"___branch:case": "__branch__case",
	}
	for in, want := range cases {
		if got := varName(in); got != want {
			t.Fatalf("varName(%q)=%q want %q", in, got, want)
		}
	}
}

func TestCTypeBoolEnumRangeFunction(t *testing.T) {
	enum := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "3"}}
	fn, err := goivy.NewFunctionSort(enum, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	if got := cppType(goivy.Boolean); got != "bool" {
		t.Fatalf("bool cppType=%s", got)
	}
	if got := cppType(enum); got != "color" {
		t.Fatalf("enum cppType=%s", got)
	}
	if got := cppType(rng); got != "long long" {
		t.Fatalf("range cppType=%s", got)
	}
	if got := cppType(fn); got != "std::map<color,bool>" {
		t.Fatalf("function cppType=%s", got)
	}
}

func TestTupleTypeForBinaryRelation(t *testing.T) {
	s := &goivy.UninterpretedSort{Name: "node"}
	fn, err := goivy.NewFunctionSort(s, s, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	if got := cppType(fn); got != "std::map<std::tuple<node,node>,bool>" {
		t.Fatalf("binary cppType=%s", got)
	}
}

func TestStateSymbolDeclarationsRelationAndFunction(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation link(X:node,Y:node)
action step = {
}
`)
	out, err := Generate(mod, Config{ClassName: "net"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Header, "std::map<std::tuple<node,node>,bool> link;") {
		t.Fatalf("missing relation declaration:\n%s", out.Header)
	}
}

func TestImplOneConstantFunctionRelationEnumRange(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type idx = {0..2}
individual saved : color
function owner(I:idx) : color
relation marked(C:color)
action step = {
    saved := owner(0);
    marked(saved) := true
}
`)
	out, err := Generate(mod, Config{ClassName: "mixed"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"enum color { red, green };",
		"typedef long long idx;",
		"color saved;",
		"std::map<idx,color> owner;",
		"std::map<color,bool> marked;",
		"saved = owner[0];",
		"marked[saved] = true;",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestImplDerivedDefinitionEmitsMethodNotState(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
derived is_red(C:color) = C = red
action step = {
    assert is_red(red)
}
`)
	out, err := Generate(mod, Config{ClassName: "deriveds"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"bool is_red(color C);", "bool deriveds::is_red(deriveds::color C)", "return (C == red);", "ivy_assert(is_red(red)"} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	if strings.Contains(out.Header, "std::map<color,bool> is_red;") {
		t.Fatalf("derived definition should not be emitted as mutable state:\n%s", out.Header)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestImplBeforeAfterMixinShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual flag : bool
action step = {
}
before step {
    flag := true
}
after step {
    flag := false
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "mixcase"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"flag = true;", "flag = false;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestDestructorStructDeclaration(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type cell
destructor shade(C:cell) : color
`)
	out, err := Generate(mod, Config{ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"struct cell {", "color shade;", "bool operator==(const cell &other) const", "bool operator<(const cell &other) const"} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q in header:\n%s", want, out.Header)
		}
	}
	if strings.Contains(out.Header, "typedef long long cell;") {
		t.Fatalf("destructor-backed sort should not also be typedef'd:\n%s", out.Header)
	}
	compileGeneratedCPP(t, out)
}

func TestEmitExprBooleanAndEquality(t *testing.T) {
	x := goivy.NewConst("x", goivy.Boolean)
	y := goivy.NewConst("y", goivy.Boolean)
	eq, err := goivy.NewEq(x, y)
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	g := &Generator{}
	got, err := g.emitExpr(eq)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != "(x == y)" {
		t.Fatalf("emitExpr=%q", got)
	}
}

func TestEmitExprLiteral(t *testing.T) {
	flag := goivy.NewConst("flag", goivy.Boolean)
	got, err := (&Generator{}).emitExpr(goivy.NewLiteral(0, flag))
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != "!(flag)" {
		t.Fatalf("literal code=%q", got)
	}
}

func TestEmitExprDisequality(t *testing.T) {
	x := goivy.NewConst("x", goivy.Boolean)
	y := goivy.NewConst("y", goivy.Boolean)
	eq, err := goivy.NewEq(x, y)
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	neq, err := goivy.NewNot(eq)
	if err != nil {
		t.Fatalf("NewNot: %v", err)
	}
	got, err := (&Generator{}).emitExpr(neq)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != "!((x == y))" {
		t.Fatalf("disequality code=%q", got)
	}
}

func TestEmitExprQuantifierFiniteEnumLoop(t *testing.T) {
	color := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	x, err := goivy.NewVariable("X", color)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	body, err := goivy.NewEq(x, x)
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	fa, err := goivy.NewForAll([]*goivy.LogicVariable{x}, body)
	if err != nil {
		t.Fatalf("NewForAll: %v", err)
	}
	g := &Generator{}
	got, err := g.emitExpr(fa)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if !strings.Contains(got, "for (color X : {red, green})") || !strings.Contains(got, "return true;") {
		t.Fatalf("unexpected quantifier code:\n%s", got)
	}
}

func TestEmitExprSomeFiniteEnumLoop(t *testing.T) {
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
	some := goivy.NewSome([]goivy.Expr{x}, body)
	got, err := (&Generator{}).emitExpr(some)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{"for (color X : {red, green})", "if ((X == green)) return X;", "return red;"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in some expression:\n%s", want, got)
		}
	}
}

func TestEmitExprTemporalUnsupportedIsActionable(t *testing.T) {
	g, err := goivy.NewGlobally(nil, goivy.True)
	if err != nil {
		t.Fatalf("NewGlobally: %v", err)
	}
	_, err = (&Generator{}).emitExpr(g)
	if err == nil || !strings.Contains(err.Error(), "temporal expression") || !strings.Contains(err.Error(), "*goivy.LogicGlobally") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmitExprUnsupportedIsActionable(t *testing.T) {
	lambda, err := goivy.NewLambda(nil, goivy.True)
	if err != nil {
		t.Fatalf("NewLambda: %v", err)
	}
	_, err = (&Generator{}).emitExpr(lambda)
	if err == nil || !strings.Contains(err.Error(), "unsupported expression") || !strings.Contains(err.Error(), "*goivy.Lambda") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateUnsupportedActionReturnsError(t *testing.T) {
	mod := goivy.New()
	mod.Name = "unsupported"
	mod.Actions.Set("step", goivy.NewThunkAction(goivy.NewConst("x", goivy.TopS)))
	_, err := Generate(mod, Config{ClassName: "unsupported"})
	if err == nil || !strings.Contains(err.Error(), "unsupported action") || !strings.Contains(err.Error(), "*goivy.LogicThunkAction") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAllKnownUnsupportedActionsReturnErrors(t *testing.T) {
	cases := []struct {
		name string
		act  goivy.Action
		want string
	}{
		{
			name: "thunk",
			act:  goivy.NewThunkAction(goivy.NewConst("x", goivy.TopS)),
			want: "*goivy.LogicThunkAction",
		},
		{
			name: "instantiate",
			act:  goivy.NewInstantiateAction(goivy.NewConst("schema", goivy.TopS)),
			want: "*goivy.LogicInstantiateAction",
		},
		{
			name: "bad sequence child",
			act:  goivy.NewSequence(goivy.NewConst("not_action", goivy.TopS)),
			want: "unsupported sequence child",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mod := goivy.New()
			mod.Name = tc.name
			mod.Actions.Set("step", tc.act)
			_, err := Generate(mod, Config{ClassName: "bad"})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEmitAssignNullaryAndIndexed(t *testing.T) {
	s := &goivy.UninterpretedSort{Name: "node"}
	fn, err := goivy.NewFunctionSort(s, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	r := goivy.NewConst("r", fn)
	x := goivy.NewConst("x", s)
	lhs := goivy.MustApply(r, x)
	assign := goivy.NewAssignAction(lhs, goivy.NewConst("true", goivy.Boolean))
	var w cppWriter
	(&Generator{}).emitAction(&w, assign)
	if got := normalizeCPP(w.String()); got != "r[x] = true;" {
		t.Fatalf("assign code=%q", got)
	}
}

func TestEmitIfWhileChoice(t *testing.T) {
	assign := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("true", goivy.Boolean))
	ifAct := goivy.NewIfAction(goivy.NewConst("cond", goivy.Boolean), assign)
	whileAct := goivy.NewWhileAction(goivy.NewConst("cond", goivy.Boolean), ifAct)
	choice := goivy.NewChoiceActionOn(goivy.NewActionsConfig(), whileAct, assign)
	var w cppWriter
	(&Generator{}).emitAction(&w, choice)
	got := w.String()
	for _, want := range []string{"switch (___ivy_choose", "while (cond)", "if (cond)", "flag = true;"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestGeneratedChoiceActionCompiles(t *testing.T) {
	assign := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("true", goivy.Boolean))
	choice := goivy.NewChoiceActionOn(goivy.NewActionsConfig(), assign)
	mod := goivy.New()
	mod.Name = "choice"
	mod.Actions.Set("step", choice)
	mod.Relations.Set("flag", goivy.Boolean)
	out, err := Generate(mod, Config{ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"int ___ivy_choose(int rng, const char *name, int id);", "int runner::___ivy_choose", "return 0;"} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestGeneratedIfSomeActionCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step = {
    if some c:color. c = green {
        saved := c
    } else {
        saved := red
    }
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "someact"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"bool __ivy_some", "for (color loc__c : {red, green})", "if (!__ivy_some", "saved = loc__c;", "saved = red;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestReturnAndIgnoreActionsCompileInVoidMethod(t *testing.T) {
	mod := goivy.New()
	mod.Name = "markers"
	mod.Actions.Set("step", goivy.NewSequence(goivy.NewIgnoreAction(), goivy.NewReturnAction()))
	out, err := Generate(mod, Config{ClassName: "markers"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "return;") {
		t.Fatalf("return marker was not emitted:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestReturnActionInSingleReturnMethodEmitsValueReturn(t *testing.T) {
	color := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	ret := goivy.NewConst("out", color)
	var w cppWriter
	(&Generator{currentReturns: []*goivy.Const{ret}}).emitAction(&w, goivy.NewReturnAction())
	if got := normalizeCPP(w.String()); got != "return out;" {
		t.Fatalf("return marker code=%q", got)
	}
}

func TestEmitAfterInitEnumLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
after init {
    marked(C) := false
}
`)
	out, err := Generate(mod, Config{ClassName: "paint"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"void paint::__init()", "for (color C : {red, green})", "marked[C] = false;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestEmitAfterInitRangeLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..2}
relation marked(I:idx)
after init {
    marked(I) := false
}
`)
	out, err := Generate(mod, Config{ClassName: "ranges"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"typedef long long idx;", "for (idx I = 0; I <= 2; I++)", "marked[I] = false;"} {
		if !strings.Contains(out.Header, want) && !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestEmitExprQuantifierFiniteRangeLoop(t *testing.T) {
	idx := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "1"}, Ub: goivy.NumeralBound{Value: "3"}}
	i, err := goivy.NewVariable("I", idx)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	body, err := goivy.NewEq(i, i)
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	fa, err := goivy.NewForAll([]*goivy.LogicVariable{i}, body)
	if err != nil {
		t.Fatalf("NewForAll: %v", err)
	}
	got, err := (&Generator{}).emitExpr(fa)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if !strings.Contains(got, "for (long long I = 1; I <= 3; I++)") || !strings.Contains(got, "return true;") {
		t.Fatalf("unexpected quantifier code:\n%s", got)
	}
}

func TestGeneratedSingleReturnActionCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action echo(c:color) returns (out:color) = {
    out := c
}
action step = {
    call saved := echo(green)
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "calls"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"color echo(color c)", "color out = red;", "return out;", "saved = echo(green);"} {
		if !strings.Contains(out.Impl, want) && !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestGeneratedMultipleReturnActionCompiles(t *testing.T) {
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
	out, err := Generate(mod, Config{ClassName: "multi"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"void split(color c, color &out, bool &good)", "out = c;", "good = true;", "split(green, saved, ok);"} {
		if !strings.Contains(out.Impl, want) && !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestGeneratedVarActionCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step = {
    var tmp : color := green;
    saved := tmp
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "locals"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"{", "color loc__tmp = red;", "loc__tmp = green;", "saved = loc__tmp;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestGeneratedHavocActionCompiles(t *testing.T) {
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
	mod.Actions.Set("step", goivy.NewHavocAction(goivy.NewConst("saved", color)))
	out, err := Generate(mod, Config{ClassName: "havocs"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"saved = red;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestGeneratedSetActionCompiles(t *testing.T) {
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
	relSort, err := goivy.NewFunctionSort(color, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	marked := goivy.NewConst("marked", relSort)
	red := goivy.NewConst("red", color)
	green := goivy.NewConst("green", color)
	setRed := goivy.NewSetAction(goivy.MustApply(marked, red))
	clearGreen := goivy.NewSetAction(goivy.NewLiteral(0, goivy.MustApply(marked, green)))
	mod.Actions.Set("step", goivy.NewSequence(setRed, clearGreen))
	out, err := Generate(mod, Config{ClassName: "sets"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"marked[red] = true;", "marked[green] = false;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestGeneratedDebugActionCompiles(t *testing.T) {
	mod := goivy.New()
	mod.Name = "debugcase"
	mod.Relations.Set("flag", goivy.Boolean)
	mod.Actions.Set("step", goivy.NewDebugAction(goivy.NewConst(`"tick"`, goivy.TopS), goivy.NewConst("flag", goivy.Boolean)))
	out, err := Generate(mod, Config{ClassName: "debugcase"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{`"event\" : \"tick\"`, `"value0\" : " << (flag)`} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestGeneratedLetActionSubstitutesBoundSymbol(t *testing.T) {
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
	out, err := Generate(mod, Config{ClassName: "lets"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "other[red] = true;") {
		t.Fatalf("let binding did not substitute marked->other:\n%s", out.Impl)
	}
	if strings.Contains(out.Impl, "marked[red] = true;") {
		t.Fatalf("let body still writes original symbol:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestGeneratedEnvAndBindOldsActionsCompile(t *testing.T) {
	mod := goivy.New()
	mod.Name = "envcase"
	mod.Relations.Set("flag", goivy.Boolean)
	assign := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("true", goivy.Boolean))
	mod.Actions.Set("step", goivy.NewEnvActionOn(goivy.NewActionsConfig(), goivy.NewBindOldsAction(assign)))
	out, err := Generate(mod, Config{ClassName: "envcase"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"switch (___ivy_choose", "flag = true;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestGeneratedFieldActionsCompile(t *testing.T) {
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
	shadeSort, err := goivy.NewFunctionSort(cell, color)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	shade := goivy.NewConst("shade", shadeSort)
	current := goivy.NewConst("current", cell)
	other := goivy.NewConst("other", cell)
	green := goivy.NewConst("green", color)
	mod.Actions.Set("step", goivy.NewSequence(
		goivy.NewAssignFieldAction(shade, current, green),
		goivy.NewCopyFieldAction(other, shade, current, shade),
		goivy.NewNullFieldAction(shade, current),
	))
	out, err := Generate(mod, Config{ClassName: "fields"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"current.shade = green;", "other.shade = current.shade;", "current.shade = red;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestConstructorInitializesParams(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter initial : color
individual saved : color
after init {
    saved := initial
}
`)
	out, err := Generate(mod, Config{ClassName: "paramcase"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"color initial;", "paramcase(color initial);"} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q in header:\n%s", want, out.Header)
		}
	}
	for _, want := range []string{"paramcase::paramcase(paramcase::color initial)", "this->initial = initial;", "saved = initial;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	if !strings.Contains(out.Impl, "__init();") {
		t.Fatalf("constructor should run __init:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestReplWithModuleParameterCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter initial : color
individual saved : color
after init {
    saved := initial
}
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "paramrepl"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "paramrepl ivy{paramrepl::red};") {
		t.Fatalf("missing parameterized repl construction:\n%s", out.Impl)
	}
	if strings.Contains(out.Impl, "ivy.__init();") {
		t.Fatalf("repl should rely on constructor initialization:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestUnsupportedModelInitialStateReturnsError(t *testing.T) {
	mod := goivy.New()
	mod.Name = "badinit"
	mod.InitCond = goivy.FormulaToClauses(goivy.NewConst("flag", goivy.Boolean), nil)
	_, err := Generate(mod, Config{ClassName: "badinit"})
	if err == nil || !strings.Contains(err.Error(), "initial constraints") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitEqualityConstraintEmitsAssignment(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step = {
}
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
	out, err := Generate(mod, Config{ClassName: "initeq"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "saved = green;") {
		t.Fatalf("missing init assignment:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestInitRelationIffConstraintEmitsLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
action step = {
}
`)
	color, ok := mod.Sig.Sorts.Get2("color")
	if !ok {
		t.Fatal("missing color sort")
	}
	relSort, ok := mod.Relations.Get2("marked")
	if !ok {
		relSort, ok = mod.Functions.Get2("marked")
	}
	if !ok {
		t.Fatal("missing marked relation")
	}
	c, err := goivy.NewVariable("C", color)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	marked, err := goivy.NewApply(goivy.NewConst("marked", relSort), c)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	isGreen, err := goivy.NewEq(c, goivy.NewConst("green", color))
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	init, err := goivy.NewIff(marked, isGreen)
	if err != nil {
		t.Fatalf("NewIff: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(init, nil)
	out, err := Generate(mod, Config{ClassName: "initrel"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"for (color C : {red, green})", "marked[C] = (C == green);"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestEmitNativeActionAntiquotes(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set_native(c:color) = {
    <<<
        `+"`saved`"+` = `+"`c`"+`;
    >>>
}
export set_native
`)
	out, err := Generate(mod, Config{ClassName: "nativecase"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "saved = c;") {
		t.Fatalf("missing antiquoted native assignment:\n%s", out.Impl)
	}
	if strings.Contains(out.Impl, "native action omitted") {
		t.Fatalf("native action was not emitted:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestNativePrimitiveTypeDeclarationFromInterpret(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type handle
interpret handle -> <<< primitive int >>>
individual h : handle
action step = {
    h := h
}
`)
	out, err := Generate(mod, Config{ClassName: "nativeint"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"typedef int handle;", "handle h;"} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q in header:\n%s", want, out.Header)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestNativeTypeAntiquoteReferencesSort(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..3}
type vec
interpret vec -> <<< primitive std::vector<`+"`idx`"+`> >>>
individual xs : vec
action step = {
}
`)
	out, err := Generate(mod, Config{ClassName: "nativevec"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Header, "typedef std::vector<idx> vec;") {
		t.Fatalf("missing rendered native vector type:\n%s", out.Header)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestNativeClassParameterDefaultCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..3}
type vec
interpret vec -> <<< std::vector<`+"`idx`"+`> >>>
parameter initial : vec
individual xs : vec
after init {
    xs := initial
}
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "nativeparam"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "nativeparam ivy{nativeparam::vec()};") {
		t.Fatalf("native parameter default should value-initialize the class type:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestNativeDefinitionEmitsTemplateMethod(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx
interpret idx -> <<< int >>>
relation lt(X:idx,Y:idx)
definition lt(x:idx,y:idx) = <<< `+"`x`"+` < `+"`y`"+` >>>
action step(x:idx,y:idx) = {
    assert lt(x,y)
}
`)
	out, err := Generate(mod, Config{ClassName: "nativedef"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"bool lt(idx x, idx y);", "bool nativedef::lt(nativedef::idx x, nativedef::idx y)", "return x < y;", "ivy_assert(lt(x, y)"} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	if strings.Contains(out.Header, "std::map<std::tuple<idx,idx>,bool> lt;") {
		t.Fatalf("native definition should not be emitted as mutable state:\n%s", out.Header)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTopLevelNativeBlocksEmitHeaderMemberAndInit(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
<<< header
#include <sstream>
>>>
<<< member
int `+"`native_counter`"+`;
>>>
<<< init
`+"`native_counter`"+` = 7;
>>>
action tick_native = {
    <<<
        `+"`native_counter`"+`++;
    >>>
}
export tick_native
`)
	out, err := Generate(mod, Config{ClassName: "nativeblocks"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"#include <sstream>", "int native_counter;"} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q in header:\n%s", want, out.Header)
		}
	}
	for _, want := range []string{"nativeblocks::nativeblocks() {", "native_counter = 7;", "native_counter++;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestDuplicateOnceNativeHeaderEmitsOnce(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
<<< once
#include <deque>
>>>
<<< once
#include <deque>
>>>
action step = {
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "dedupe"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got := strings.Count(out.Header, "#include <deque>"); got != 1 {
		t.Fatalf("native once header count=%d header:\n%s", got, out.Header)
	}
}

func TestReplDispatchForExportedAction(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, `if (action == "step")`) {
		t.Fatalf("missing repl dispatch:\n%s", out.Impl)
	}
}

func TestReplMainReadsCommandsFromStdin(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"std::string action;", "while (std::cin >> action)", "ivy2cpp_dispatch(ivy, action);"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl main:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestReplIgnoresInternalAction(t *testing.T) {
	mod := goivy.New()
	mod.Name = "replprivate"
	mod.Actions.Set("internal", goivy.NewSequence())
	mod.Actions.Set("step", goivy.NewSequence())
	mod.PublicActions.Set("step", true)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, `if (action == "step")`) {
		t.Fatalf("missing exported dispatch:\n%s", out.Impl)
	}
	if strings.Contains(out.Impl, `if (action == "internal")`) {
		t.Fatalf("internal action should not be dispatchable:\n%s", out.Impl)
	}
}

func TestReplDispatchForParameterizedExportCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, `if (action == "set") { ivy.set(runner::red); return; }`) {
		t.Fatalf("missing parameterized repl dispatch:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestTargetTestEmitsZ3Includes(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step = {
    saved := green
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "testrunner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{`#include "z3++.h"`, "class gen", "z3::context ctx", "__from_solver", "ivy2cpp_randomize"} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
}

func TestTargetGenEmitsGeneratorHooks(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
parameter enabled : bool
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "genrunner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"static void ivy2cpp_generate(genrunner &ivy)", "gen g;", "ivy2cpp_setup(g);", "ivy2cpp_randomize(g, ivy);", "ivy2cpp_progress"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
}

func TestRandomizeFiniteEnumRelation(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
individual saved : color
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "randomizer"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		`g.mk_enum("color", {"red", "green"});`,
		`g.mk_decl("marked", {"color"}, "bool");`,
		"for (randomizer::color __ivy_arg0 : {randomizer::red, randomizer::green})",
		`g.randomize("marked", static_cast<int>(__ivy_arg0), "bool");`,
		"ivy.marked[__ivy_arg0] = g.random_bool();",
		"ivy.saved = ivy2cpp_random_color(g);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
}

func TestGeneratedTestFixtureCompilesWhenZ3Available(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
after init {
    marked(C) := false
}
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "z3fixture"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestBuildTrueCompilesWhenToolchainAvailable(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "buildme.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
type color = {red, green}
relation marked(C:color)
action step = {
}
export step
`), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	out, err := CompileAndGenerate(spec, map[string]string{"target": "test", "build": "true"}, Config{ClassName: "buildme"})
	if err != nil {
		t.Fatalf("CompileAndGenerate: %v", err)
	}
	exe, err := BuildOutput(out, dir)
	if err != nil {
		if isMissingZ3ToolchainError(err) {
			t.Skipf("Z3 C++ toolchain unavailable: %v", err)
		}
		t.Fatalf("BuildOutput: %v", err)
	}
	if _, err := os.Stat(exe); err != nil {
		t.Fatalf("missing built executable %s: %v", exe, err)
	}
}

func TestCommandWritesHeaderAndImpl(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go run integration in short mode")
	}
	dir := t.TempDir()
	spec := filepath.Join(dir, "tiny.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
action step = {
}
export step
`), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	cmd := exec.Command("go", "run", "./cmd/ivy2cpp", "target=repl", "classname=Tiny", "outdir="+dir, spec)
	cmd.Dir = filepath.Join(repoRoot(t), "goivy")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run ivy2cpp: %v\n%s", err, out)
	}
	headerPath := filepath.Join(dir, "tiny.h")
	if _, err := os.Stat(headerPath); err != nil {
		t.Fatalf("missing tiny.h: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "tiny.cpp")); err != nil {
		t.Fatalf("missing tiny.cpp: %v", err)
	}
	header, err := os.ReadFile(headerPath)
	if err != nil {
		t.Fatalf("read tiny.h: %v", err)
	}
	if !strings.Contains(string(header), "class Tiny") {
		t.Fatalf("classname parameter not reflected in header:\n%s", header)
	}
}

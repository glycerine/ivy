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
	compileGeneratedCPPWithPrefix(t, out, "")
}

func compileGeneratedCPPWithPrefix(t *testing.T, out *Output, prefix string) {
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
	if prefix != "" {
		prefixPath := filepath.Join(dir, "ivy2cpp_prefix.h")
		if err := os.WriteFile(prefixPath, []byte(prefix), 0o644); err != nil {
			t.Fatalf("write C++ prefix: %v", err)
		}
		args = append(args, "-include", prefixPath)
	}
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

func buildGeneratedExecutable(t *testing.T, out *Output) string {
	t.Helper()
	dir := t.TempDir()
	exe, err := BuildOutput(out, dir)
	if err != nil {
		if isMissingZ3ToolchainError(err) {
			t.Skip(err.Error())
		}
		t.Fatalf("BuildOutput: %v\nheader:\n%s\nimpl:\n%s", err, out.Header, out.Impl)
	}
	return exe
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

func TestPythonAndGoGeneratedGenFixtureCompilesWithStubs(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`
	pyHeader, pyImpl := runPythonIvyToCpp(t, src, "target=gen")
	prefix := "#include <fstream>\nextern std::ofstream __ivy_modelfile;\n"
	compileGeneratedCPPWithPrefix(t, &Output{
		Header:    pyHeader,
		Impl:      pyImpl,
		BaseName:  "oracle",
		ClassName: "oracle",
	}, prefix)

	mod := compileIvySource(t, src)
	goOut, err := Generate(mod, Config{Target: "gen", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	compileGeneratedCPP(t, goOut)
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
	if got := cppType(rng); got != "idx" {
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
		"saved = owner[(0 < 0 ? 0 : 2 < 0 ? 2 : 0)];",
		"marked[saved] = true;",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestNamedRangeTypedefUsedInStateAndMethods(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..2}
individual seen : idx
action set(i:idx) = {
    seen := i
}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "rangetypes"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"typedef long long idx;",
		"idx seen;",
		"void set(idx i);",
		"void rangetypes::set(rangetypes::idx i)",
		"rangetypes::idx i = ivy2cpp_parse_idx",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	if strings.Contains(out.Header, "long long seen;") || strings.Contains(out.Header, "void set(long long i);") {
		t.Fatalf("named range was erased to long long:\n%s", out.Header)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestRangeNumeralConstantsAreClamped(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..2}
individual saved : idx
action high = {
    saved := 5
}
`)
	out, err := Generate(mod, Config{ClassName: "rangeclamp"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"saved = (5 < 0 ? 0 : 2 < 5 ? 2 : 5);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in range clamp output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestNegativeRangeNumeralConstantIsClamped(t *testing.T) {
	idx := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "2"}}
	got, err := (&Generator{}).emitExpr(goivy.NewConst("-1", idx))
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != "(-1 < 0 ? 0 : 2 < -1 ? 2 : -1)" {
		t.Fatalf("negative range clamp=%q", got)
	}
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

func TestDestructorFunctionNotEmittedAsMutableState(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type cell
destructor shade(C:cell) : color
`)
	out, err := Generate(mod, Config{ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if strings.Contains(out.Header, "std::map<cell,color> shade;") {
		t.Fatalf("destructor function should be a struct field, not mutable state:\n%s", out.Header)
	}
	compileGeneratedCPP(t, out)
}

func TestVariantStructDeclarationAndNoDestructorState(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type msg
type color = {red, green}
variant request of msg = struct {
    shade : color
}
individual saved : msg
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "variants"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"struct request {", "color shade;"} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q in variant header:\n%s", want, out.Header)
		}
	}
	if strings.Contains(out.Header, "std::map<request,color> shade;") {
		t.Fatalf("variant destructor field should not be emitted as mutable state:\n%s", out.Header)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestVariantSupertypeAssignmentCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type msg
type color = {red, green}
variant request of msg = struct {
    shade : color
}
individual saved : msg
after init {
    var tmp : request;
    tmp.shade := green;
    saved := tmp
}
action step = {
    assert saved = saved
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "variants"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"struct request {",
		"struct msg {",
		"int __tag;",
		"request __request;",
		"msg(const request &value)",
		"saved = loc__tmp;",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	if strings.Contains(out.Header, "typedef long long msg;") {
		t.Fatalf("variant supertype must not be emitted as long long:\n%s", out.Header)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestVariantSomeDowncastCompiles(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t
variant a of t
variant b of t
individual v : t
action save(inp:a) = {
    v := inp
}
action load returns(out:a) = {
    if some (q:a) v *> q {
        out := q
    }
}
export save
export load
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "variantdown"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"struct t {",
		"a __a;",
		"t(const a &value) : __tag(0), __a(value) {}",
		"if (v.__tag == 0)",
		"static variantdown::a ivy2cpp_parse_a(const std::string &s)",
		`variantdown::a inp = ivy2cpp_parse_a(ivy2cpp_read_arg(input, "inp"));`,
		"a loc__q = v.__a;",
		"out = loc__q;",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
	exe := buildGeneratedExecutable(t, out)
	cmd := exec.Command(exe)
	cmd.Stdin = strings.NewReader("save 7\nload\n")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated variant repl failed: %v\n%s\nimpl:\n%s", err, buf, out.Impl)
	}
	if strings.TrimSpace(string(buf)) != "7" {
		t.Fatalf("unexpected variant repl output %q\nimpl:\n%s", buf, out.Impl)
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
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"friend std::ostream &operator<<(std::ostream &out, const a &value)",
		"friend std::ostream &operator<<(std::ostream &out, const t &value)",
		"case 0: out << value.__a; return out;",
		"variantout::t __ivy_result = ivy.make();",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	exe := buildGeneratedExecutable(t, out)
	cmd := exec.Command(exe)
	cmd.Stdin = strings.NewReader("make\n")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated variant-return repl failed: %v\n%s\nimpl:\n%s", err, buf, out.Impl)
	}
	if strings.TrimSpace(string(buf)) != "0" {
		t.Fatalf("unexpected variant-return repl output %q\nimpl:\n%s", buf, out.Impl)
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
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"friend std::ostream &operator<<(std::ostream &out, const req &value)",
		`out << "shade:" << value.shade;`,
		"friend std::ostream &operator<<(std::ostream &out, const t &value)",
		"case 0: out << value.__req; return out;",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	exe := buildGeneratedExecutable(t, out)
	cmd := exec.Command(exe)
	cmd.Stdin = strings.NewReader("make\n")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated struct-variant-return repl failed: %v\n%s\nimpl:\n%s", err, buf, out.Impl)
	}
	if !strings.Contains(strings.TrimSpace(string(buf)), "shade:") {
		t.Fatalf("unexpected struct variant output %q\nimpl:\n%s", buf, out.Impl)
	}
}

func TestEmitExprVariantRelation(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t
variant a of t
individual v : t
individual av : a
action step = {
    assert v *> av
}
`)
	out, err := Generate(mod, Config{ClassName: "variantrel"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "ivy_assert((v.__tag == 0 && v.__a == av)") {
		t.Fatalf("missing variant relation assertion:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestEmitExprExistsVariantRelation(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t
variant a of t
variant b of t
individual v : t
action step = {
    assert exists Q:a. v *> Q
}
`)
	out, err := Generate(mod, Config{ClassName: "variantexists"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "ivy_assert((v.__tag == 0)") {
		t.Fatalf("missing optimized variant-exists assertion:\n%s", out.Impl)
	}
	if strings.Contains(out.Impl, "cannot emit bounded loop") || strings.Contains(out.Impl, "for (a Q") {
		t.Fatalf("variant-exists should not loop over unbounded variant leaf sort:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
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

func TestEmitExprExistsUsesExtensionalRelationMap(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation marked(N:node)
action check = {
    assert exists X:node. marked(X)
}
`)
	out, err := Generate(mod, Config{ClassName: "extq"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"for (auto it = marked.begin(), en = marked.end(); it != en; ++it)",
		"if (!it->second) continue;",
		"node X = it->first;",
		"if (marked[X]) return true;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in extensional quantifier:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestEmitExprExistsUsesBinaryExtensionalRelationMap(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
individual dst : node
relation edge(X:node,Y:node)
action check = {
    assert exists X:node. edge(X,dst)
}
`)
	out, err := Generate(mod, Config{ClassName: "extq2"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"for (auto it = edge.begin(), en = edge.end(); it != en; ++it)",
		"node X = std::get<0>(it->first);",
		"if (edge[std::make_tuple(X, dst)]) return true;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in binary extensional quantifier:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestEmitExprForallUsesExtensionalRelationAntecedent(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation marked(N:node)
relation ok(N:node)
action check = {
    assert forall X:node. marked(X) -> ok(X)
}
`)
	out, err := Generate(mod, Config{ClassName: "extfa"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"for (auto it = marked.begin(), en = marked.end(); it != en; ++it)",
		"if (!it->second) continue;",
		"node X = it->first;",
		"if (!((!(marked[X]) || (ok[X])))) return false;",
		"return true;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in forall extensional quantifier:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
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
		"if (v.__tag == 0)",
		"a Q = v.__a;",
		"return Q;",
		"return a();",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in variant some expression:\n%s", want, got)
		}
	}
	if strings.Contains(got, "for (a Q") {
		t.Fatalf("variant some should not loop over unbounded variant leaf sort:\n%s", got)
	}
}

func TestEmitExprSomeWithElseFiniteEnumLoop(t *testing.T) {
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
	some := goivy.NewSomeWithElse([]goivy.Expr{x}, body, x, red)
	got, err := (&Generator{}).emitExpr(some)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	for _, want := range []string{"for (color X : {red, green})", "if ((X == green)) return X;", "return red;"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in some/else expression:\n%s", want, got)
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
		"if (v.__tag == 0)",
		"a Q = v.__a;",
		"return Q;",
		"return fallback;",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in variant some/else expression:\n%s", want, got)
		}
	}
	if strings.Contains(got, "for (a Q") {
		t.Fatalf("variant some/else should not loop over unbounded variant leaf sort:\n%s", got)
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

func TestGeneratedIfSomeUsesExtensionalRelationMap(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation marked(N:node)
individual saved : node
action pick = {
    if some x:node. marked(x) {
        saved := x
    }
}
export pick
`)
	out, err := Generate(mod, Config{ClassName: "someext"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"for (auto it = marked.begin(), en = marked.end(); it != en; ++it)",
		"if (!it->second) continue;",
		"node loc__x = it->first;",
		"if (!__ivy_some1 && (marked[loc__x]))",
		"saved = loc__x;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in extensional if some:\n%s", want, out.Impl)
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
	if !strings.Contains(got, "for (idx I = 1; I <= 3; I++)") || !strings.Contains(got, "return true;") {
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

func TestInitEqualityConstraintSwapsStateOnRight(t *testing.T) {
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
	eq, err := goivy.NewEq(goivy.NewConst("green", color), goivy.NewConst("saved", color))
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(eq, nil)
	out, err := Generate(mod, Config{ClassName: "initright"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "saved = green;") {
		t.Fatalf("missing swapped init assignment:\n%s", out.Impl)
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

func TestInitRelationIffConstraintSwapsStateOnRight(t *testing.T) {
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
	init, err := goivy.NewIff(isGreen, marked)
	if err != nil {
		t.Fatalf("NewIff: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(init, nil)
	out, err := Generate(mod, Config{ClassName: "initrightrel"})
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

func TestInitForAllRelationIffConstraintEmitsLoop(t *testing.T) {
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
	isRed, err := goivy.NewEq(c, goivy.NewConst("red", color))
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	init, err := goivy.NewIff(marked, isRed)
	if err != nil {
		t.Fatalf("NewIff: %v", err)
	}
	forall, err := goivy.NewForAll([]*goivy.LogicVariable{c}, init)
	if err != nil {
		t.Fatalf("NewForAll: %v", err)
	}
	mod.InitCond = goivy.FormulaToClauses(forall, nil)
	out, err := Generate(mod, Config{ClassName: "initforall"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"for (color C : {red, green})", "marked[C] = (C == red);"} {
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
	for _, want := range []string{"std::string action;", "while (std::cin >> action)", "ivy2cpp_dispatch(ivy, action, std::cin);"} {
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
	for _, want := range []string{`if (action == "set")`, `runner::color c = ivy2cpp_parse_color(ivy2cpp_read_arg(input, "c"));`, `ivy.set(c);`} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl dispatch:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestReplDispatchParsesEnumBoolRangeArgs(t *testing.T) {
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
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"static std::string ivy2cpp_read_arg(std::istream &input, const char *name)",
		"static bool ivy2cpp_parse_bool(const std::string &s)",
		"static runner::color ivy2cpp_parse_color(const std::string &s)",
		"static runner::idx ivy2cpp_parse_idx(const std::string &s)",
		`runner::color c = ivy2cpp_parse_color(ivy2cpp_read_arg(input, "c"));`,
		`bool b = ivy2cpp_parse_bool(ivy2cpp_read_arg(input, "b"));`,
		`runner::idx i = ivy2cpp_parse_idx(ivy2cpp_read_arg(input, "i"));`,
		"ivy.set(c, b, i);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestReplDispatchParsesUninterpretedArgs(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
individual saved : node
action set(n:node) = {
    saved := n;
    assert n = 7
}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"static runner::node ivy2cpp_parse_node(const std::string &s)",
		"long long value = std::stoll(s);",
		"return static_cast<runner::node>(value);",
		`runner::node n = ivy2cpp_parse_node(ivy2cpp_read_arg(input, "n"));`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	exe := buildGeneratedExecutable(t, out)
	cmd := exec.Command(exe)
	cmd.Stdin = strings.NewReader("set 7\n")
	if buf, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated repl did not parse uninterpreted arg: %v\n%s\nimpl:\n%s", err, buf, out.Impl)
	}
}

func TestGeneratedReplExecutableParsesEnumArg(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action set(c:color) = {
    assert c = green
}
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	exe := buildGeneratedExecutable(t, out)
	cmd := exec.Command(exe)
	cmd.Stdin = strings.NewReader("set green\n")
	if buf, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated repl did not parse enum arg: %v\n%s\nimpl:\n%s", err, buf, out.Impl)
	}
}

func TestReplDispatchWritesSingleReturn(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action echo(c:color) returns (out:color) = {
    out := c
}
export echo
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"static void ivy2cpp_write_value(std::ostream &out, runner::color value)",
		"runner::color __ivy_result = ivy.echo(c);",
		"ivy2cpp_write_value(std::cout, __ivy_result);",
		"std::cout << std::endl;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl output:\n%s", want, out.Impl)
		}
	}
	exe := buildGeneratedExecutable(t, out)
	cmd := exec.Command(exe)
	cmd.Stdin = strings.NewReader("echo green\n")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated repl single return failed: %v\n%s\nimpl:\n%s", err, buf, out.Impl)
	}
	if strings.TrimSpace(string(buf)) != "green" {
		t.Fatalf("unexpected repl output %q\nimpl:\n%s", buf, out.Impl)
	}
}

func TestReplDispatchWritesMultipleReturns(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action split(c:color) returns (out:color, good:bool) = {
    out := c;
    good := true
}
export split
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"runner::color out = runner::red;",
		"bool good = false;",
		"ivy.split(c, out, good);",
		"ivy2cpp_write_value(std::cout, out);",
		`std::cout << " ";`,
		"ivy2cpp_write_value(std::cout, good);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl output:\n%s", want, out.Impl)
		}
	}
	exe := buildGeneratedExecutable(t, out)
	cmd := exec.Command(exe)
	cmd.Stdin = strings.NewReader("split green\n")
	buf, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated repl multiple return failed: %v\n%s\nimpl:\n%s", err, buf, out.Impl)
	}
	if strings.TrimSpace(string(buf)) != "green true" {
		t.Fatalf("unexpected repl output %q\nimpl:\n%s", buf, out.Impl)
	}
}

func TestPythonOracleReplParameterizedActionShape(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`
	pyHeader, pyImpl := runPythonIvyToCpp(t, src, "target=repl")
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"set", "color", "green"} {
		if !strings.Contains(pyHeader+pyImpl, want) {
			t.Fatalf("python repl output missing %q:\nheader:\n%s\nimpl:\n%s", want, pyHeader, pyImpl)
		}
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("go repl output missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	if !strings.Contains(out.Impl, "ivy2cpp_parse_color") {
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

func TestTargetGenGeneratesInitAndActionGeneratorClasses(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "genclasses"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"class ivy2cpp_action_gen",
		"class init_gen : public gen",
		"bool init_gen::generate(genclasses &obj)",
		"class set_gen : public gen",
		"bool set_gen::generate(genclasses &obj)",
		"void set_gen::execute(genclasses &obj)",
		"init_gen my_init_gen(ivy);",
		"my_init_gen.generate(ivy);",
		"set_gen set_generator(ivy);",
		"if (set_generator.generate(ivy))",
		"set_generator.execute(ivy);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in gen output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTargetGenRandomizesActionParamsAndExecutes(t *testing.T) {
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
	out, err := Generate(mod, Config{Target: "gen", ClassName: "genparams"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"genparams::color c;",
		"bool b;",
		"genparams::idx i;",
		"this->c = ivy2cpp_random_color(*this);",
		"this->b = this->random_bool();",
		"this->i = ivy2cpp_random_idx(*this);",
		"obj.set(this->c, this->b, this->i);",
		"if (set_generator.generate(ivy))",
		"set_generator.execute(ivy);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in gen output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTargetGenRandomizesNumericAndVariantLeafParams(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
type t
variant a of t
individual saved : node
individual wrapped : t
action touch(n:node,av:a) = {
    saved := n;
    wrapped := av
}
export touch
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "gennumeric"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"static gennumeric::node ivy2cpp_random_node(gen &g)",
		"static gennumeric::a ivy2cpp_random_a(gen &g)",
		"return static_cast<gennumeric::node>(g.random_index(0, 4));",
		"return static_cast<gennumeric::a>(g.random_index(0, 4));",
		"this->n = ivy2cpp_random_node(*this);",
		"this->av = ivy2cpp_random_a(*this);",
		"obj.touch(this->n, this->av);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in gen output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTargetGenEmitsSolverConversionsForFiniteSorts(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "solverconv"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"template <typename T> z3::expr __to_solver(gen &g, const char *sort_name, const T &value)",
		"return g.int_to_z3(sort_name, static_cast<long long>(value));",
		"static void __from_solver(gen &g, const z3::expr &expr, solverconv::color &out)",
		"std::string text = g.eval_expr(expr).to_string();",
		`if (text == "color_0" || text == "red")`,
		"out = solverconv::red;",
		"static z3::expr __to_solver(gen &g, const char *sort_name, solverconv::color value)",
		"case solverconv::green: return g.int_to_z3(sort_name, 1);",
		"template <typename T> z3::expr __to_solver(gen &g, const z3::expr &expr, const T &value)",
		"return expr == g.int_to_z3(expr.get_sort(), static_cast<long long>(value));",
		"static z3::expr __to_solver(gen &g, const z3::expr &expr, solverconv::color value)",
		"case solverconv::green: return expr == g.int_to_z3(expr.get_sort(), 1);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in gen output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTargetGenEmitsBoolRangeSolverConversionsAndProgressState(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..2}
individual ok : bool
individual seen : idx
action step(i:idx) = {
    seen := i;
    ok := true
}
export step
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "solverconv2"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"z3::model model;",
		"model = slvr.get_model();",
		"z3::expr eval_expr(const z3::expr &expr)",
		"long long eval(const z3::expr &expr)",
		"template <> void __from_solver<bool>(gen &g, const z3::expr &expr, bool &out)",
		"z3::expr solver_value = g.eval_expr(expr);",
		"Z3_lbool value = solver_value.bool_value();",
		"template <> void __from_solver<long long>(gen &g, const z3::expr &expr, long long &out)",
		"out = g.eval(expr);",
		"static void ivy2cpp_progress(gen &g, const std::string &label)",
		"g.progress.push_back(label);",
		`ivy2cpp_progress(g, "randomize");`,
		`ivy2cpp_progress(*this, "step_gen");`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in gen output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTargetTestUsesInitAndActionGeneratorClasses(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "testclasses"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"class init_gen : public gen",
		"class set_gen : public gen",
		"init_gen my_init_gen(ivy);",
		"my_init_gen.generate(ivy);",
		"set_gen set_generator(ivy);",
		"if (set_generator.generate(ivy))",
		"set_generator.execute(ivy);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in test output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTargetGenChecksSolverBeforeExecute(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "checkgen"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"bool check()",
		"model = slvr.get_model();",
		"bool init_gen::generate(checkgen &obj)",
		"bool set_gen::generate(checkgen &obj)",
		"if (!check())",
		"return false;",
		"if (set_generator.generate(ivy))",
		"set_generator.execute(ivy);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in gen output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
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

func TestRandomizeBinaryFiniteRelationUsesTupleKey(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type bit = {low, high}
relation edge(C:color,B:bit)
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "randomizer2"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		`g.mk_decl("edge", {"color", "bit"}, "bool");`,
		"for (randomizer2::color __ivy_arg0 : {randomizer2::red, randomizer2::green})",
		"for (randomizer2::bit __ivy_arg1 : {randomizer2::low, randomizer2::high})",
		`g.randomize("edge", {static_cast<int>(__ivy_arg0), static_cast<int>(__ivy_arg1)}, "bool");`,
		"ivy.edge[std::make_tuple(__ivy_arg0, __ivy_arg1)] = g.random_bool();",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTargetGenRandomizeAddsZ3Constraints(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "z3random"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"z3::expr mk_apply_expr(const char *decl_name, const std::vector<int> &args)",
		"expr_args.push_back(int_to_z3(decl.domain(i), args[i]));",
		"void add_alit(const z3::expr &pred)",
		"slvr.add(pred);",
		"randomize(mk_apply_expr(decl_name, args_vec), range);",
		"z3::expr pred = expr == int_to_z3(expr.get_sort(), value);",
		"add_alit(pred);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTargetGenRandomizeUsesFiniteSortBounds(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type idx = {2..4}
individual saved : color
individual seen : idx
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "z3bounds"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"std::map<std::string, long long> sort_los;",
		"std::map<std::string, long long> sort_his;",
		"sort_his[std::string(name)] = values.size() == 0 ? 0 : static_cast<long long>(values.size()) - 1;",
		`g.mk_enum("color", {"red", "green"});`,
		`g.mk_int("idx", 2, 4);`,
		"std::map<std::string, long long>::const_iterator lo_it = sort_los.find(range);",
		"value = random_index(static_cast<int>(lo), static_cast<int>(hi));",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
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

func TestGeneratedGenFixtureCompilesWhenZ3Available(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type idx = {0..2}
relation marked(C:color,I:idx)
individual saved : color
action set(c:color,i:idx) = {
    saved := c;
    marked(c,i) := true
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "z3genfixture"})
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

func TestCommandBuildTrueProducesExecutable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go run integration in short mode")
	}
	dir := t.TempDir()
	spec := filepath.Join(dir, "tinybuild.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
type color = {red, green}
action echo(c:color) returns (out:color) = {
    out := c
}
export echo
`), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	cmd := exec.Command("go", "run", "./cmd/ivy2cpp", "target=repl", "build=true", "classname=TinyBuild", "outdir="+dir, spec)
	cmd.Dir = filepath.Join(repoRoot(t), "goivy")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run ivy2cpp build=true: %v\n%s", err, out)
	}
	exe := filepath.Join(dir, "tinybuild")
	if _, err := os.Stat(exe); err != nil {
		t.Fatalf("missing tinybuild executable: %v", err)
	}
	run := exec.Command(exe)
	run.Stdin = strings.NewReader("echo green\n")
	buf, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("built ivy2cpp executable failed: %v\n%s", err, buf)
	}
	if strings.TrimSpace(string(buf)) != "green" {
		t.Fatalf("unexpected executable output %q", buf)
	}
}

func TestCommandBuildTrueTargetTestProducesExecutableWhenZ3Available(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go run integration in short mode")
	}
	if _, _, err := z3BuildArgs(); err != nil {
		if isMissingZ3ToolchainError(err) {
			t.Skipf("Z3 C++ toolchain unavailable: %v", err)
		}
		t.Fatalf("Z3 build args: %v", err)
	}
	dir := t.TempDir()
	spec := filepath.Join(dir, "tinyz3.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
type color = {red, green}
relation marked(C:color)
action step = {
}
export step
`), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	cmd := exec.Command("go", "run", "./cmd/ivy2cpp", "target=test", "build=true", "classname=TinyZ3", "outdir="+dir, spec)
	cmd.Dir = filepath.Join(repoRoot(t), "goivy")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go run ivy2cpp target=test build=true: %v\n%s", err, out)
	}
	exe := filepath.Join(dir, "tinyz3")
	if _, err := os.Stat(exe); err != nil {
		t.Fatalf("missing tinyz3 executable: %v", err)
	}
}

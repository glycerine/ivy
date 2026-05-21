package ivy2cpp

// Test policy for ivy2cpp:
// - Keep this package's tests fast enough for the inner development loop.
// - Do not shell out from these tests. Stay in-process and call the ivy2cpp
//   library API directly.
// - Do not invoke the generated C++ toolchain from these tests. Validate the
//   generated C++ shape with focused string/metadata assertions, and leave
//   compiler/toolchain coverage to a separate explicit slow/integration suite.
// - CLI behavior should be tested by parsing params and calling
//   CompileAndGenerateAll(filename, params, Config{...}) directly.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

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

func normalizeOracleCPPForComparison(s string) string {
	return normalizeCPP(s)
}

type comparableFeature struct {
	name  string
	terms []string
}

func assertComparableFeatures(t *testing.T, label, text string, features []comparableFeature) {
	t.Helper()
	for _, f := range features {
		if !hasLineWithAllTerms(text, f.terms...) {
			t.Fatalf("%s missing comparable feature %q terms %v:\n%s", label, f.name, f.terms, text)
		}
	}
}

func hasLineWithAllTerms(s string, terms ...string) bool {
	for _, line := range strings.Split(s, "\n") {
		ok := true
		for _, term := range terms {
			if !strings.Contains(line, term) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func compileGeneratedCPP(t *testing.T, out *Output) {
	t.Helper()
	compileGeneratedCPPWithPrefix(t, out, "")
}

// Historical helper name: this must stay in-process and must not invoke a C++
// compiler. The unit test contract is generated-output shape validation only.
func compileGeneratedCPPWithPrefix(t *testing.T, out *Output, prefix string) {
	t.Helper()
	if out == nil {
		t.Fatal("nil generated output")
	}
	if out.Header == "" || out.Impl == "" {
		t.Fatalf("generated output should contain header and impl: %+v", out)
	}
	assertNoUnsupportedCPP(t, out)
	if prefix != "" {
		_ = prefix
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

func readSupportHeader(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "include2cpp", name))
	if err != nil {
		t.Fatalf("read support header %s: %v", name, err)
	}
	return string(b)
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

func TestMinimalReplShape(t *testing.T) {
	src := `#lang ivy1.7
action step = {
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"class oracle", "__init", "__tick", "step"} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("go output missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
}

func TestGoOutputUsesSharedRuntimeIncludes(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	raw := out.Header + "\n" + out.Impl
	for _, want := range []string{`#include "ivy_hash.hpp"`, `#include "ivy_go_repl.hpp"`} {
		if !strings.Contains(raw, want) {
			t.Fatalf("go output missing shared runtime support %q:\n%s", want, raw)
		}
	}
	if strings.Contains(raw, "class reader {\npublic:") ||
		strings.Contains(raw, "class timer {\npublic:") {
		t.Fatalf("go output should not embed reader/timer runtime:\n%s", raw)
	}
	for _, unwanted := range []string{
		"struct ivy_value {",
		"int ask_ret(long long bound)",
		"void parse_command(const std::string &cmd",
		"class stdin_reader: public reader",
	} {
		if strings.Contains(raw, unwanted) {
			t.Fatalf("go output should not embed runtime support %q:\n%s", unwanted, raw)
		}
	}
	stripped := normalizeCPP(raw)
	for _, want := range []string{"class oracle", "void __tick(int timeout);", "void oracle::__tick"} {
		if !strings.Contains(stripped, want) {
			t.Fatalf("stripped output lost %q:\n%s", want, stripped)
		}
	}
}

func TestGoOutputUsesSharedSupportIncludes(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{`#include "ivy_hash.hpp"`, `#include "ivy_go_repl.hpp"`} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("go output missing support include %q:\n%s", want, out.Header)
		}
	}
	if strings.Contains(out.Impl, "static void ivy2cpp_parse_command") {
		t.Fatalf("go output should include repl support instead of embedding parser:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestGoOutputUsesSharedZ3RuntimeIncludes(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`
	for _, target := range []string{"test", "gen"} {
		t.Run(target, func(t *testing.T) {
			mod := compileIvySource(t, src)
			out, err := Generate(mod, Config{Target: target, ClassName: "oracle"})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			raw := out.Header + "\n" + out.Impl
			for _, want := range []string{`#include "z3++.h"`, `#include "ivy_go_z3.hpp"`} {
				if !strings.Contains(raw, want) {
					t.Fatalf("expected Go %s output to include shared Z3 runtime %q:\n%s", target, want, raw)
				}
			}
			for _, unwanted := range []string{
				"class gen : public ivy_gen",
				"template <class T> void __from_solver( gen",
				"class z3_thunk : public thunk",
			} {
				if strings.Contains(raw, unwanted) {
					t.Fatalf("Go %s output should not embed Z3 runtime support %q:\n%s", target, unwanted, raw)
				}
			}
		})
	}
}

func TestGoComparableCoreHooksAndActionShape(t *testing.T) {
	src := `#lang ivy1.7
var x : bool
after init {
    x := true
}
action step = {
    x := ~x
}
export step
extract iso = this
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	goOut := normalizeCPP(out.Header + "\n" + out.Impl)
	features := []comparableFeature{
		{name: "class", terms: []string{"class", "oracle"}},
		{name: "init hook", terms: []string{"void", "__init"}},
		{name: "tick hook", terms: []string{"void", "__tick", "int"}},
		{name: "progress hook", terms: []string{"ivy_check_progress", "int"}},
		{name: "choice hook", terms: []string{"___ivy_choose", "int"}},
		{name: "x state", terms: []string{"bool", "x"}},
		{name: "step action", terms: []string{"step"}},
		{name: "init assignment", terms: []string{"x", "=", "true"}},
		{name: "step assignment", terms: []string{"x", "=", "!"}},
	}
	assertComparableFeatures(t, "go ivy2cpp", goOut, features)
}

func TestGoComparableProgressTickShape(t *testing.T) {
	src := `#lang ivy1.7
individual ready : bool
individual helper_ready : bool
progress wait = ready
progress helper = helper_ready
rely wait -> helper
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	goOut := normalizeCPP(out.Header + "\n" + out.Impl)
	features := []comparableFeature{
		{name: "tick method", terms: []string{"__tick", "__timeout"}},
		{name: "wait update", terms: []string{"wait", "ready", "?", "0", "+ 1"}},
		{name: "helper update", terms: []string{"helper", "helper_ready", "?", "0", "+ 1"}},
		{name: "rely max", terms: []string{"std::max", "helper"}},
		{name: "progress reset", terms: []string{"wait", "=", "0"}},
		{name: "progress check", terms: []string{"ivy_check_progress", "wait"}},
	}
	assertComparableFeatures(t, "go ivy2cpp", goOut, features)
}

func TestTickDeclaredForMinimalModule(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "tickempty"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"virtual void ivy_check_progress(int guarantee_ticks, int assume_ticks) {}",
		"void __tick(int timeout);",
		"void tickempty::__tick(int __timeout)",
		"(void)__timeout;",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	if strings.Contains(out.Impl, "ivy_check_progress(wait") {
		t.Fatalf("empty tick should not call ivy_check_progress:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestProgressCounterDeclarationAndTickUpdate(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual ready : bool
relation ok(C:color)
progress wait = ready
progress waitn(C) = ok(C)
`)
	out, err := Generate(mod, Config{ClassName: "tickprog"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"long long wait;",
		"std::map<color,long long> waitn;",
		"wait = ready ? 0 : wait + 1;",
		"for (color C : {red, green})",
		"waitn[C] = ok[C] ? 0 : waitn[C] + 1;",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTickCallsIvyCheckProgressWithoutRely(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual ready : bool
progress wait = ready
`)
	out, err := Generate(mod, Config{ClassName: "ticknor"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"long long __ivy_maxt",
		"__ivy_maxt",
		"= 0;",
		"ivy_check_progress(wait, __ivy_maxt",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestProgressBinaryTupleCounterAndTickUpdate(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type bit = {low, high}
relation edge(C:color,B:bit)
progress wait(C,B) = edge(C,B)
`)
	out, err := Generate(mod, Config{ClassName: "ticktuple"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"std::map<std::tuple<color,bit>,long long> wait;",
		"for (color C : {red, green})",
		"for (bit B : {low, high})",
		"wait[std::make_tuple(C, B)] = edge[std::make_tuple(C, B)] ? 0 : wait[std::make_tuple(C, B)] + 1;",
		"ivy_check_progress(wait[std::make_tuple(C, B)], __ivy_maxt",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTickRelyImplicationComputesMaxAndChecksProgress(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual ready : bool
individual helper_ready : bool
progress wait = ready
progress helper = helper_ready
rely wait -> helper
`)
	if len(mod.Rely) == 0 {
		t.Fatalf("expected parsed rely declarations to populate module")
	}
	out, err := Generate(mod, Config{ClassName: "tickrely"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"#include <algorithm>",
		"long long __ivy_maxt",
		"__ivy_maxt",
		"= std::max(",
		"helper",
		"if (__ivy_maxt",
		"> __timeout)",
		"wait = 0;",
		"ivy_check_progress(wait, __ivy_maxt",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
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
	out, err := Generate(mod, Config{ClassName: "ticksubst"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"for (color C : {red, green})",
		"wait[C] = ready[C] ? 0 : wait[C] + 1;",
		"__ivy_maxt",
		"= std::max(",
		"helper[C]",
		"ivy_check_progress(wait[C], __ivy_maxt",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
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
	out, err := Generate(mod, Config{ClassName: "tickextravar"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"for (color C : {red, green})",
		"for (color D : {red, green})",
		"__ivy_maxt",
		"= std::max(",
		"helper[D]",
		"ivy_check_progress(wait[C], __ivy_maxt",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTickUnconditionalRelySkipsProgressCheckLikePython(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual ready : bool
progress wait = ready
rely wait
`)
	if len(mod.Rely) == 0 {
		t.Fatalf("expected parsed rely declarations to populate module")
	}
	out, err := Generate(mod, Config{ClassName: "tickbarely"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "wait = ready ? 0 : wait + 1;") {
		t.Fatalf("missing progress update:\n%s", out.Impl)
	}
	if strings.Contains(out.Impl, "ivy_check_progress(wait") {
		t.Fatalf("bare rely should skip progress check like Python:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestGoPortSharedShape(t *testing.T) {
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
			want:   []string{`#include "z3++.h"`, "enum color"},
		},
		{
			name:   "gen",
			target: "gen",
			want:   []string{`#include "z3++.h"`, "enum color"},
		},
	}
	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			mod := compileIvySource(t, src)
			out, err := Generate(mod, Config{Target: tc.target, ClassName: "oracle"})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			goOut := normalizeCPP(out.Header + "\n" + out.Impl)
			for _, want := range tc.want {
				if !strings.Contains(goOut, want) {
					t.Fatalf("go output missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
				}
			}
		})
	}
}

func TestGeneratedFixturesHaveExpectedShape(t *testing.T) {
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
			mod := compileIvySource(t, src)
			goOut, err := Generate(mod, Config{Target: target, ClassName: "oracle"})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			compileGeneratedCPP(t, goOut)
		})
	}
}

func TestGeneratedGenFixtureHasExpectedShapeWithStubs(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`
	prefix := "#include <fstream>\nextern std::ofstream __ivy_modelfile;\n"
	mod := compileIvySource(t, src)
	goOut, err := Generate(mod, Config{Target: "gen", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	compileGeneratedCPPWithPrefix(t, goOut, prefix)
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
			want: []string{`if (action == "set")`, "small_repl_repl ivy;"},
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

func TestRuntimeSkeletonAcrossTargets(t *testing.T) {
	src := `#lang ivy1.7
action step = {
}
export step
`
	for _, target := range []string{"impl", "class", "repl", "test", "gen"} {
		t.Run(target, func(t *testing.T) {
			mod := compileIvySource(t, src)
			out, err := Generate(mod, Config{Target: target, ClassName: "runtime_" + target})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			raw := out.Header + out.Impl
			for _, want := range []string{
				"typedef std::string __strlit;",
				"extern std::ofstream __ivy_out;",
				"void __ivy_exit(int);",
				"std::vector<std::string> __argv;",
				"void __lock();",
				"void __unlock();",
				"void install_reader(reader *);",
				"void install_thread(reader *);",
				"void install_timer(timer *);",
				"std::vector<int> ___ivy_stack;",
				"std::ofstream __ivy_modelfile;",
				"pthread_mutex_init(&mutex, NULL);",
				"pthread_cancel(thread_ids[i]);",
				"pthread_join(thread_ids[i], NULL);",
			} {
				if !strings.Contains(raw, want) {
					t.Fatalf("%s missing runtime skeleton %q:\nheader:\n%s\nimpl:\n%s", target, want, out.Header, out.Impl)
				}
			}
			if target == "gen" || target == "test" {
				for _, want := range []string{
					"struct ivy_gen",
					"ivy_gen *___ivy_gen;",
					"ss << name << ':' << id;",
					"___ivy_gen->choose(rng, ss.str().c_str());",
				} {
					if !strings.Contains(raw, want) {
						t.Fatalf("%s missing generator skeleton %q:\nheader:\n%s\nimpl:\n%s", target, want, out.Header, out.Impl)
					}
				}
			} else if !strings.Contains(raw, "return 0;") {
				t.Fatalf("%s non-generator choose should return 0:\n%s", target, out.Impl)
			}
			if target == "gen" {
				if !strings.Contains(out.Header, "extern void ivy_assert(bool, const char *);") {
					t.Fatalf("gen target should use external assert hook:\n%s", out.Header)
				}
				if strings.Contains(out.Header, "virtual void ivy_assert(bool truth") {
					t.Fatalf("gen target should not emit base assert member:\n%s", out.Header)
				}
			} else if !strings.Contains(out.Header, "virtual void ivy_assert(bool truth, const char *msg) {}") {
				t.Fatalf("%s target should emit no-op base assert member:\n%s", target, out.Header)
			}
		})
	}
}

func TestRuntimeReplAndTestSubclassGlue(t *testing.T) {
	src := `#lang ivy1.7
action step = {
    assert false
}
export step
`
	cases := []struct {
		target string
		want   []string
	}{
		{
			target: "repl",
			want: []string{
				"class runtime_repl_repl : public runtime_repl",
				`__ivy_out << "assertion_failed(\"" << msg << "\")" << std::endl;`,
				`std::cerr << msg << ": error: assertion failed\n";`,
				"runtime_repl_repl ivy;",
				"ivy.__unlock();",
				"ivy.__lock();",
				"ivy2cpp_dispatch(ivy, action, args);",
				`__ivy_out << "> ";`,
			},
		},
		{
			target: "test",
			want: []string{
				"class runtime_test_repl : public runtime_test",
				`__ivy_out << "assumption_failed(\"" << msg << "\")" << std::endl;`,
				"std::vector<reader *> readers;",
				"std::vector<timer *> timers;",
				"bool initializing = false;",
				"initializing = true;",
				"readers[rdridx]->bind();",
				`__ivy_out << "test_completed" << std::endl;`,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.target, func(t *testing.T) {
			mod := compileIvySource(t, src)
			out, err := Generate(mod, Config{Target: tc.target, ClassName: "runtime_" + tc.target})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			raw := out.Header + out.Impl
			for _, want := range tc.want {
				if !strings.Contains(raw, want) {
					t.Fatalf("%s missing %q:\nheader:\n%s\nimpl:\n%s", tc.target, want, out.Header, out.Impl)
				}
			}
		})
	}
}

func TestRuntimeChoiceStackAndGeneratorPlumbing(t *testing.T) {
	src := `#lang ivy1.7
action helper = {
}
action step = {
    call helper
}
export step
`
	for _, target := range []string{"gen", "test"} {
		t.Run(target, func(t *testing.T) {
			mod := compileIvySource(t, src)
			out, err := Generate(mod, Config{Target: target, ClassName: "stacky"})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			for _, want := range []string{
				"obj.___ivy_gen = this;",
				"___ivy_stack.push_back(",
				"helper();",
				"___ivy_stack.pop_back();",
				"return ___ivy_gen->choose(rng, ss.str().c_str());",
			} {
				if !strings.Contains(out.Header+out.Impl, want) {
					t.Fatalf("%s missing generator plumbing %q:\nheader:\n%s\nimpl:\n%s", target, want, out.Header, out.Impl)
				}
			}
		})
	}
	support := readSupportHeader(t, "ivy_go_z3.hpp")
	for _, want := range []string{
		"struct ivy_gen",
		"class gen : public ivy_gen",
		"int choose(int rng, const char *name)",
	} {
		if !strings.Contains(support, want) {
			t.Fatalf("support header missing %q:\n%s", want, support)
		}
	}
}

func TestRuntimeNativeReaderTimerSkeletonShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
<<< impl
class demo_reader : public reader {
public:
    int fdes() { return -1; }
    void read() {}
};
class demo_timer : public timer {
public:
    int ms_delay() { return 1; }
    void timeout(int) {}
};
>>>
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "native_runtime"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	raw := out.Header + out.Impl
	for _, want := range []string{
		`#include "ivy_threads.hpp"`,
		"class demo_reader : public reader",
		"class demo_timer : public timer",
		"void native_runtime::install_reader(reader *r)",
		"void native_runtime::install_thread(reader *r)",
		"void native_runtime::install_timer(timer *r)",
		"ReaderThreadFunction",
		"TimerThreadFunction",
	} {
		if !strings.Contains(raw, want) {
			t.Fatalf("missing native runtime shape %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
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

func TestMergeParamsAcceptsPythonDriverSurfaceAndDefaultsToGen(t *testing.T) {
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

func TestMergeParamsRejectsUnknownTargetAndCompiler(t *testing.T) {
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

func TestGenerateTargetClassEmitsNoMainButKeepsReplSupport(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "class", ClassName: "OnlyClass"})
	if err != nil {
		t.Fatalf("Generate target=class: %v", err)
	}
	if out.Target != "class" || out.EffectiveTarget != "repl" || out.EmitMain {
		t.Fatalf("class target metadata = target %q effective %q emitMain %v", out.Target, out.EffectiveTarget, out.EmitMain)
	}
	raw := out.Header + "\n" + out.Impl
	for _, want := range []string{"class OnlyClass", "ivy2cpp_dispatch", `#include "ivy_go_repl.hpp"`} {
		if !strings.Contains(raw, want) {
			t.Fatalf("class output missing %q:\n%s", want, raw)
		}
	}
	if strings.Contains(raw, "int main(") || strings.Contains(raw, "int custom_main(") || strings.Contains(raw, "ivy2cpp_generate(") {
		t.Fatalf("class output should not emit a main/gen harness:\n%s", raw)
	}
	compileGeneratedCPP(t, out)
}

func TestMainNameAndTestDefaultsAreEmitted(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "WithMain", MainName: "ivy_main", TestIters: "7", TestRuns: "3"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"int ivy_main(int argc, char **argv)", "int test_iters = 7;", "int runs = 3;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestCompileAndGenerateAllWritesDescriptorForReplAndTest(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "desc.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
action step = {
}
export step
`), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	for _, target := range []string{"repl", "test"} {
		t.Run(target, func(t *testing.T) {
			batch, err := CompileAndGenerateAll(spec, map[string]string{"target": target, "classname": "Desc"}, Config{})
			if err != nil {
				t.Fatalf("CompileAndGenerateAll: %v", err)
			}
			if len(batch.Outputs) != 1 {
				t.Fatalf("outputs len = %d, want 1", len(batch.Outputs))
			}
			var raw string
			for name, text := range batch.ExtraFiles {
				if strings.HasSuffix(name, ".dsc") {
					raw = text
				}
			}
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
			}
		})
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

func TestBuildPlanForClassIsCompileOnly(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "class", ClassName: "ClassBuild"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	plan, err := BuildPlanFor(out, t.TempDir(), out.Config)
	if err != nil {
		if isMissingZ3ToolchainError(err) {
			t.Skip(err.Error())
		}
		t.Fatalf("BuildPlanFor: %v", err)
	}
	if !plan.CompileOnly || !strings.HasSuffix(plan.OutputPath, ".o") {
		t.Fatalf("class build plan = compileOnly %v output %q", plan.CompileOnly, plan.OutputPath)
	}
	if !sliceContains(plan.Args, "-c") || sliceContains(plan.Args, "-pthread") {
		t.Fatalf("class build args should compile only without pthread link arg: %v", plan.Args)
	}
}

func TestBuildPlanRejectsCLOnNonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("compiler=cl is meaningful on Windows")
	}
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
`)
	out, err := Generate(mod, Config{Target: "impl", ClassName: "CLNope", Compiler: "cl"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := BuildPlanFor(out, t.TempDir(), Config{Compiler: "cl", Target: "impl"}); err == nil {
		t.Fatalf("BuildPlanFor compiler=cl succeeded on non-Windows")
	}
}

func sliceContains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
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

func TestVariantSomeDowncastReplShape(t *testing.T) {
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
		`variantdown::a inp = ivy2cpp_parse_a(ivy2cpp_read_arg(args, 0, "inp"));`,
		"a loc__q = v.__a;",
		"out = loc__q;",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
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
	compileGeneratedCPP(t, out)
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
	compileGeneratedCPP(t, out)
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
	if !strings.Contains(out.Impl, "paramrepl_repl ivy{paramrepl::red};") {
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
	for _, want := range []string{"std::string line;", "while (std::getline(std::cin, line))", "ivy2cpp_parse_command(line, action, args);", "ivy2cpp_dispatch(ivy, action, args);"} {
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
	for _, want := range []string{`if (action == "set")`, `runner::color c = ivy2cpp_parse_color(ivy2cpp_read_arg(args, 0, "c"));`, `ivy.set(c);`} {
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
	text := out.Header + out.Impl
	for _, want := range []string{
		`#include "ivy_go_repl.hpp"`,
		"static bool ivy2cpp_parse_bool(const std::string &s)",
		"static runner::color ivy2cpp_parse_color(const std::string &s)",
		"static runner::idx ivy2cpp_parse_idx(const std::string &s)",
		`runner::color c = ivy2cpp_parse_color(ivy2cpp_read_arg(args, 0, "c"));`,
		`bool b = ivy2cpp_parse_bool(ivy2cpp_read_arg(args, 1, "b"));`,
		`runner::idx i = ivy2cpp_parse_idx(ivy2cpp_read_arg(args, 2, "i"));`,
		"ivy.set(c, b, i);",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in repl output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
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
		`runner::node n = ivy2cpp_parse_node(ivy2cpp_read_arg(args, 0, "n"));`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestGeneratedReplParsesPythonEnumArgByShape(t *testing.T) {
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
	for _, want := range []string{
		"static runner::color ivy2cpp_parse_color(const std::string &s)",
		`if (s == "green") return runner::green;`,
		`runner::color c = ivy2cpp_parse_color(ivy2cpp_read_arg(args, 0, "c"));`,
		"ivy.set(c);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl output:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestGeneratedReplRejectsWhitespaceArgDialectByShape(t *testing.T) {
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
	support := readSupportHeader(t, "ivy_go_repl.hpp")
	for _, want := range []string{
		"static void ivy2cpp_parse_command(const std::string &line, std::string &action, std::vector<std::string> &args)",
		`if (line[pos] != '(')`,
		`throw std::runtime_error("expected '(' after action");`,
		`throw std::runtime_error("trailing text after command");`,
	} {
		if !strings.Contains(support, want) {
			t.Fatalf("shared repl support missing %q:\n%s", want, support)
		}
	}
	if !strings.Contains(out.Impl, "ivy2cpp_check_arity(args, 1, action);") {
		t.Fatalf("generated dispatch should keep arity checking for parenthesized args:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestGeneratedReplAcceptsPythonParenthesizedCommandByShape(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
action echo(c:color) returns (out:color) = {
    out := c
}
export echo
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Header + out.Impl
	for _, want := range []string{
		`#include "ivy_go_repl.hpp"`,
		"ivy2cpp_parse_command",
		`runner::color c = ivy2cpp_parse_color(ivy2cpp_read_arg(args, 0, "c"));`,
		"runner::color __ivy_result = ivy.echo(c);",
		"ivy2cpp_write_value(std::cout, __ivy_result);",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in repl output:\n%s", want, out.Impl)
		}
	}
	if strings.Contains(out.Impl, "static void ivy2cpp_parse_command") {
		t.Fatalf("go repl output should include support parser, not embed it:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
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
	compileGeneratedCPP(t, out)
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
	compileGeneratedCPP(t, out)
}

func TestReplParameterizedActionShape(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"set", "color", "green"} {
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
	for _, want := range []string{`#include "z3++.h"`, `#include "ivy_go_z3.hpp"`, "__from_solver", "ivy2cpp_randomize"} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	for _, unwanted := range []string{"class gen {", "z3::context ctx;"} {
		if strings.Contains(out.Impl, unwanted) {
			t.Fatalf("go output still embeds Z3 runtime %q instead of including it:\n%s", unwanted, out.Impl)
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
	support := readSupportHeader(t, "ivy_go_z3.hpp")
	for _, want := range []string{
		"z3::model model;",
		"model = slvr.get_model();",
		"z3::expr eval_expr(const z3::expr &expr)",
		"long long eval(const z3::expr &expr)",
		"static void ivy2cpp_progress(gen &g, const std::string &label)",
		"g.progress.push_back(label);",
	} {
		if !strings.Contains(support, want) {
			t.Fatalf("missing %q in shared Go Z3 support:\n%s", want, support)
		}
	}
	for _, want := range []string{
		`#include "ivy_go_z3.hpp"`,
		"template <> void __from_solver<bool>(gen &g, const z3::expr &expr, bool &out)",
		"z3::expr solver_value = g.eval_expr(expr);",
		"Z3_lbool value = solver_value.bool_value();",
		"template <> void __from_solver<long long>(gen &g, const z3::expr &expr, long long &out)",
		"out = g.eval(expr);",
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
	support := readSupportHeader(t, "ivy_go_z3.hpp")
	for _, want := range []string{
		"bool check()",
		"model = slvr.get_model();",
	} {
		if !strings.Contains(support, want) {
			t.Fatalf("missing %q in shared Go Z3 support:\n%s", want, support)
		}
	}
	for _, want := range []string{
		`#include "ivy_go_z3.hpp"`,
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
	t0 := time.Now()
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
	vv("elap = %v", time.Since(t0))
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
	support := readSupportHeader(t, "ivy_go_z3.hpp")
	for _, want := range []string{
		"z3::expr mk_apply_expr(const char *decl_name, const std::vector<int> &args)",
		"expr_args.push_back(int_to_z3(decl.domain(i), args[i]));",
		"void add_alit(const z3::expr &pred)",
		"slvr.add(pred);",
		"randomize(mk_apply_expr(decl_name, args_vec), range);",
		"z3::expr pred = expr == int_to_z3(expr.get_sort(), value);",
		"add_alit(pred);",
	} {
		if !strings.Contains(support, want) {
			t.Fatalf("missing %q in shared Go Z3 support:\n%s", want, support)
		}
	}
	for _, want := range []string{
		`#include "ivy_go_z3.hpp"`,
		`g.randomize("marked", static_cast<int>(__ivy_arg0), "bool");`,
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
	support := readSupportHeader(t, "ivy_go_z3.hpp")
	for _, want := range []string{
		"std::map<std::string, long long> sort_los;",
		"std::map<std::string, long long> sort_his;",
		"sort_his[std::string(name)] = values.size() == 0 ? 0 : static_cast<long long>(values.size()) - 1;",
		"std::map<std::string, long long>::const_iterator lo_it = sort_los.find(range);",
		"value = random_index(static_cast<int>(lo), static_cast<int>(hi));",
	} {
		if !strings.Contains(support, want) {
			t.Fatalf("missing %q in shared Go Z3 support:\n%s", want, support)
		}
	}
	for _, want := range []string{
		`#include "ivy_go_z3.hpp"`,
		`g.mk_enum("color", {"red", "green"});`,
		`g.mk_int("idx", 2, 4);`,
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

func TestGeneratedGenFixtureHasExpectedShape(t *testing.T) {
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

func TestBuildTruePlansBuildWithoutToolchain(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "buildme.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
type color = {red, green}
action echo(c:color) returns(out:color) = {
    out := c
}
export echo
`), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	batch, err := CompileAndGenerateAll(spec, map[string]string{"target": "repl", "build": "true"}, Config{ClassName: "buildme", OutDir: dir})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	if !batch.Config.Build {
		t.Fatalf("build=true parameter not reflected in batch config: %#v", batch.Config)
	}
	if len(batch.Outputs) != 1 {
		t.Fatalf("expected one output, got %#v", batch.Outputs)
	}
	out := batch.Outputs[0]
	plan, err := BuildPlanFor(out, dir, out.Config)
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	if plan.CompileOnly {
		t.Fatalf("repl build plan should link an executable: %#v", plan)
	}
	if !strings.HasSuffix(plan.OutputPath, "buildme") || sliceContains(plan.Args, "-c") {
		t.Fatalf("unexpected repl build plan: %#v", plan)
	}
}

func TestCommandPathWritesHeaderAndImplInProcess(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "tiny.ivy")
	if err := os.WriteFile(spec, []byte(`#lang ivy1.7
action step = {
}
export step
`), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	params := map[string]string{"target": "repl", "classname": "Tiny", "outdir": dir}
	batch, err := CompileAndGenerateAll(spec, params, Config{})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	if err := WriteBatchOutput(batch, batch.Config.OutDir); err != nil {
		t.Fatalf("WriteBatchOutput: %v", err)
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

func TestCommandBuildTrueProducesBuildPlanInProcess(t *testing.T) {
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
	params := map[string]string{"target": "repl", "build": "true", "classname": "TinyBuild", "outdir": dir}
	batch, err := CompileAndGenerateAll(spec, params, Config{})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	if !batch.Config.Build {
		t.Fatalf("build=true parameter not reflected in batch config: %#v", batch.Config)
	}
	if err := WriteBatchOutput(batch, batch.Config.OutDir); err != nil {
		t.Fatalf("WriteBatchOutput: %v", err)
	}
	if len(batch.Outputs) != 1 {
		t.Fatalf("expected one output, got %#v", batch.Outputs)
	}
	plan, err := BuildPlanFor(batch.Outputs[0], batch.Config.OutDir, batch.Outputs[0].Config)
	if err != nil {
		t.Fatalf("BuildPlanFor: %v", err)
	}
	if plan.CompileOnly || !strings.HasSuffix(plan.OutputPath, "tinybuild") {
		t.Fatalf("unexpected build plan: %#v", plan)
	}
	if !strings.Contains(batch.Outputs[0].Impl, "runner") && !strings.Contains(batch.Outputs[0].Impl, "TinyBuild") {
		t.Fatalf("generated repl impl missing expected classname:\n%s", batch.Outputs[0].Impl)
	}
}

func TestCommandBuildTrueTargetTestGeneratesDescriptorAndZ3ShapeInProcess(t *testing.T) {
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
	params := map[string]string{"target": "test", "build": "true", "classname": "TinyZ3", "outdir": dir}
	batch, err := CompileAndGenerateAll(spec, params, Config{})
	if err != nil {
		t.Fatalf("CompileAndGenerateAll: %v", err)
	}
	if !batch.Config.Build || batch.Config.Target != "test" {
		t.Fatalf("unexpected batch config: %#v", batch.Config)
	}
	if len(batch.Outputs) != 1 {
		t.Fatalf("expected one output, got %#v", batch.Outputs)
	}
	out := batch.Outputs[0]
	text := out.Header + out.Impl
	for _, want := range []string{
		`#include "z3++.h"`,
		`#include "ivy_go_z3.hpp"`,
		"init_gen",
		"ext__step_gen",
		"test_params",
	} {
		if !strings.Contains(text+batch.ExtraFiles["TinyZ3.dsc"], want) {
			t.Fatalf("missing %q in target=test output:\nheader:\n%s\nimpl:\n%s\ndsc:\n%s", want, out.Header, out.Impl, batch.ExtraFiles["TinyZ3.dsc"])
		}
	}
	if err := WriteBatchOutput(batch, batch.Config.OutDir); err != nil {
		t.Fatalf("WriteBatchOutput: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "TinyZ3.dsc")); err != nil {
		t.Fatalf("missing descriptor: %v", err)
	}
}

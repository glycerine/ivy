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
//
// The one exception would be when the env var SLOW_CPP_TEST is defined or the package var SlowCppTest bool is true.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/glycerine/ivy/goivy"
)

// SlowCppTest true means we actually compile the C++, which
// means the tests will take several minutes rather than
// a second or less. Useful only occassionally to validate
// that the C++ output is compilable.
var SlowCppTest bool

func init() {
	_, SlowCppTest = os.LookupEnv("SLOW_CPP_TEST")
}

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

// Historical helper name: the normal unit test contract is generated-output
// shape validation only. Set SLOW_CPP_TEST or SlowCppTest to additionally run
// the generated C++ through the configured compiler.
func compileGeneratedCPPWithPrefix(t *testing.T, out *Output, prefix string) {
	t.Helper()
	if out == nil {
		t.Fatal("nil generated output")
	}
	if out.Header == "" || out.Impl == "" {
		t.Fatalf("generated output should contain header and impl: %+v", out)
	}
	assertNoUnsupportedCPP(t, out)
	if SlowCppTest {
		compileGeneratedCPPSlow(t, out, prefix)
	}
}

func compileGeneratedCPPSlow(t *testing.T, out *Output, prefix string) {
	t.Helper()
	compiled := *out
	// This helper validates that the generated translation unit compiles. It
	// intentionally stops before linking so impl/class shape fixtures do not
	// need to provide a generated main.
	compiled.EmitMain = false
	if prefix != "" {
		compiled.Impl = prefix + compiled.Impl
	}
	outputPath, err := BuildOutput(&compiled, t.TempDir())
	if err != nil {
		if isMissingZ3ToolchainError(err) {
			t.Skip(err.Error())
		}
		t.Fatalf("compile generated C++: %v", err)
	}
	if outputPath == "" {
		t.Fatalf("compile generated C++ produced an empty output path")
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
	for _, want := range []string{`#include "ivy_hash.hpp"`, `#include "ivy_threads.hpp"`, `#include "ivy_repl.hpp"`, `#include "ivy_value.hpp"`} {
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
		"inline int ask_ret(long long bound)",
		"inline void parse_command(const std::string &cmd",
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
	// Header: hash and threads. Impl: ivy_value + ivy_repl (Python-style).
	for _, want := range []string{`#include "ivy_hash.hpp"`, `#include "ivy_threads.hpp"`} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("go output missing header support include %q:\n%s", want, out.Header)
		}
	}
	for _, want := range []string{`#include "ivy_value.hpp"`, `#include "ivy_repl.hpp"`} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("go output missing impl support include %q:\n%s", want, out.Impl)
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
			// `ivy_go_z3.hpp` must live in the impl, mirroring Python
			// `ivy_to_cpp.py:2210-2211` (`ivy_z3_helpers.hpp`).
			if !strings.Contains(out.Impl, `#include "ivy_go_z3.hpp"`) {
				t.Fatalf("Go %s output: ivy_go_z3.hpp must be in impl, not header-only:\nimpl=\n%s", target, out.Impl)
			}
			if strings.Contains(out.Header, `#include "ivy_go_z3.hpp"`) {
				t.Fatalf("Go %s output: ivy_go_z3.hpp should not appear in header:\nheader=\n%s", target, out.Header)
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
		"int wait;",
		"int waitn[2];",
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
		"int __ivy_maxt",
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
		"int wait[2][2];",
		"for (color C : {red, green})",
		"for (bit B : {low, high})",
		"wait[C][B] = edge[C][B] ? 0 : wait[C][B] + 1;",
		"ivy_check_progress(wait[C][B], __ivy_maxt",
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
		"int __ivy_maxt",
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
	// Python ivy_to_cpp.py:1798-1802 alpha-renames extra rely-RHS
	// variables by appending "__" to their name; here D is an extra
	// (in the rely RHS but not in the rely LHS), so the inner loop and
	// the helper subscript both use D__.
	for _, want := range []string{
		"for (color C : {red, green})",
		"for (color D__ : {red, green})",
		"__ivy_maxt",
		"= std::max(",
		"helper[D__]",
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

// TestTickRelyExtraSharesProgressVarNameRenamed exercises the
// alpha-rename path in emitRelyMax. The rely RHS extra variable D
// shares its name with the outer progress LHS variable D; without
// alpha-renaming the inner loop would shadow the outer and the
// helper[D][D] subscript would collapse to a single coordinate. Python
// (ivy_to_cpp.py:1798-1802) renames the extra to D__.
func TestTickRelyExtraSharesProgressVarNameRenamed(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation ready(C:color)
relation helper_ready(C:color, D:color)
progress wait(D) = ready(D)
progress helper(X, Y) = helper_ready(X, Y)
rely wait(C) -> helper(C, D)
`)
	if len(mod.Rely) == 0 {
		t.Fatalf("expected parsed rely declarations to populate module")
	}
	out, err := Generate(mod, Config{ClassName: "tickextrarename"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"for (color D : {red, green})",
		"for (color D__ : {red, green})",
		"helper[D][D__]",
		"ivy_check_progress(wait[D], __ivy_maxt",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q:\n%s", want, out.Impl)
		}
	}
	// The bug we're guarding against: without alpha-renaming the inner
	// extras loop would shadow the outer progress loop and helper[D][D]
	// would collapse both subscripts to the inner D.
	if strings.Contains(out.Impl, "helper[D][D]") {
		t.Fatalf("variable capture: expected alpha-rename to avoid helper[D][D]:\n%s", out.Impl)
	}
	if strings.Contains(out.Impl, "helper[D__][D__]") {
		t.Fatalf("over-substitution: progress var also renamed to D__:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

// TestProgressCounterEmittedAsInt verifies the declared counter type
// is `int`, matching Python ivy_to_cpp.py:2332-2333 which passes
// c_type='int' to declare_symbol. ivy_check_progress's signature is
// (int,int), so the counter must be int to avoid narrowing.
func TestProgressCounterEmittedAsInt(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual ready : bool
relation ok(C:color)
progress wait = ready
progress waitn(C) = ok(C)
`)
	out, err := Generate(mod, Config{ClassName: "tickintctr"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"int wait;",
		"int waitn[2];",
	} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q in header:\n%s", want, out.Header)
		}
	}
	if strings.Contains(out.Header, "long long wait") {
		t.Fatalf("progress counters must be int, not long long:\n%s", out.Header)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

// TestProgressCountersNotResetInConstructor verifies that the class
// constructor does NOT reset progress counters to 0. Python only
// clears progress in init_gen (ivy_to_cpp.py:964); the constructor
// (ivy_to_cpp.py:2345-2377) does not. The only `wait = 0;` should be
// inside __tick's `if (__ivy_maxt > __timeout)` block.
func TestProgressCountersNotResetInConstructor(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
individual ready : bool
individual helper_ready : bool
progress wait = ready
progress helper = helper_ready
rely wait -> helper
`)
	out, err := Generate(mod, Config{ClassName: "tickctor"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Slice out the constructor body. Python's constructor is
	//   <classname>::<classname>(...) { ... }
	// Find that and assert no `wait = 0;` inside.
	impl := out.Impl
	ctorMarker := "tickctor::tickctor("
	start := strings.Index(impl, ctorMarker)
	if start < 0 {
		t.Fatalf("could not locate constructor in impl:\n%s", impl)
	}
	open := strings.Index(impl[start:], "{")
	if open < 0 {
		t.Fatalf("could not locate constructor body opening brace:\n%s", impl[start:])
	}
	open += start
	depth := 0
	end := -1
	for i := open; i < len(impl); i++ {
		switch impl[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i
			}
		}
		if end >= 0 {
			break
		}
	}
	if end < 0 {
		t.Fatalf("could not locate end of constructor body:\n%s", impl[open:])
	}
	body := impl[open : end+1]
	if strings.Contains(body, "wait = 0;") || strings.Contains(body, "helper = 0;") {
		t.Fatalf("progress counters must not be reset in constructor; body:\n%s", body)
	}
	// Sanity: the reset should still appear in __tick where Python emits it.
	tickIdx := strings.Index(impl, "__tick(int __timeout)")
	if tickIdx < 0 || !strings.Contains(impl[tickIdx:], "wait = 0;") {
		t.Fatalf("expected wait = 0; inside __tick body:\n%s", impl)
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
			want: []string{
				"enum color",
				"bool marked[2];",
				// Two-phase quantified assignment per Python emit_assign.
				"__ivy_tmp1[C] = false;",
				"marked[C] = __ivy_tmp1[C];",
			},
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
				// New: cmd_reader-based dispatch (Python style).
				"class runtime_repl_cmd_reader : public stdin_reader",
				"parse_command(cmd, action, args);",
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
	for _, want := range []string{
		"class OnlyClass",
		"class OnlyClass_cmd_reader : public stdin_reader",
		`#include "ivy_repl.hpp"`,
	} {
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
	if got := cppType(rng); got != "unsigned" {
		t.Fatalf("range cppType=%s", got)
	}
	if got := cppType(fn); got != "bool[2]" {
		t.Fatalf("function cppType=%s", got)
	}
}

func TestTupleTypeForBinaryRelation(t *testing.T) {
	s := &goivy.UninterpretedSort{Name: "node"}
	fn, err := goivy.NewFunctionSort(s, s, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	if got := cppType(fn); got != "hash_thunk<__tup__int__int,bool>" {
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
	if !strings.Contains(out.Header, "struct __tup__int__int {") {
		t.Fatalf("missing tuple key declaration:\n%s", out.Header)
	}
	if !strings.Contains(out.Header, "hash_thunk<__tup__int__int,bool> link;") {
		t.Fatalf("missing relation declaration:\n%s", out.Header)
	}
}

func TestLargeFiniteFunctionUsesHashThunkAndCTuple(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..32}
relation big(X:idx,Y:idx)
action step = {
}
`)
	out, err := Generate(mod, Config{ClassName: "largefun"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"struct __tup__unsigned__unsigned {",
		"unsigned arg0;",
		"unsigned arg1;",
		"size_t __hash() const { return hash_space::hash<unsigned>()(arg0) + hash_space::hash<unsigned>()(arg1); }",
		"hash_thunk<__tup__unsigned__unsigned,bool> big;",
		"hash_space::hash_map<D,R,HashFun> memo;",
		"bool operator==(const hash_thunk<D,R,HashFun> &other) const",
	} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q in header:\n%s", want, out.Header)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestRangeArrayDimensionUsesUpperBoundIndexSpace(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {2..4}
relation marked(I:idx)
after init {
    marked(I) := true
}
`)
	out, err := Generate(mod, Config{ClassName: "rangedim"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Two-phase quantified assignment per Python emit_assign.
	for _, want := range []string{
		"typedef unsigned idx;",
		"bool marked[5];",
		"for (unsigned I = 2; I <= 4; I++)",
		"__ivy_tmp1[I] = true;",
		"marked[I] = __ivy_tmp1[I];",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
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
		"typedef unsigned idx;",
		"color saved;",
		"color owner[3];",
		"bool marked[2];",
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
		"typedef unsigned idx;",
		"unsigned seen;",
		"void set(unsigned i);",
		"void rangetypes::set(unsigned i)",
		// Range sorts typedef to unsigned, so _arg<unsigned> from
		// ivy_value.hpp handles the parse with the carc bound.
		`ivy.set(_arg<unsigned>(args, 0, 3));`,
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	if strings.Contains(out.Header, "long long seen;") || strings.Contains(out.Header, "void set(long long i);") {
		t.Fatalf("named range was erased to long long instead of the Python unsigned cardinal type:\n%s", out.Header)
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
	// Python emit_some_action body shape (ivy_to_cpp.py:1592-1625):
	//   declare primary-return local, run AssignAction(retval, rhs),
	//   trailing `return retval;`. Mirrored by TODO 010 port.
	for _, want := range []string{"bool is_red(color C);", "bool deriveds::is_red(deriveds::color C)", "val = (C == red);", "return val;", "ivy_assert(is_red(red)"} {
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
		"struct wrap {",
		"template <typename T> struct twrap : public wrap",
		"int tag;",
		"wrap *ptr;",
		"msg(int tag, wrap *ptr) : tag(tag), ptr(ptr) {}",
		"static int temp_counter;",
		"template <typename T> static const T &unwrap(const msg &x)",
		"int variants::msg::temp_counter = 0;",
		"bool operator==(const variants::msg &s, const variants::msg &t)",
		"std::ostream &operator<<(std::ostream &s, const variants::msg &t)",
		"template <> variants::msg _arg<variants::msg>",
		"template <> void __ser<variants::msg>",
		"template <> void __deser<variants::msg>",
		`g.ctx.function("*>:msg:request", g.sort("msg"), g.sort("request"), g.ctx.bool_sort())`,
		"saved = msg(0, new msg::twrap<request>(loc__tmp));",
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
		"struct wrap {",
		"template <typename T> struct twrap : public wrap",
		"if (v.tag == 0)",
		// Variant `a` already has its `_arg<>` specialization emitted by
		// variant.go; the dispatcher calls it directly. The bound for
		// uninterpreted variant subtypes is 0 (Python csortcard fallback).
		`variantdown::a _arg<variantdown::a>(std::vector<ivy_value> &args, unsigned idx, long long bound)`,
		`ivy.save(_arg<variantdown::a>(args, 0, 0));`,
		"v = t(0, new t::twrap<a>(inp));",
		"a loc__q = t::unwrap< a >(v);",
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
		"friend std::ostream &operator<<(std::ostream &s, const t &t);",
		"std::ostream &operator<<(std::ostream &s, const variantout::t &t)",
		`case 0: s << "a:" << variantout::t::unwrap< variantout::a >(t); break;`,
		"out = t(0, new t::twrap<a>(loc__q));",
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
		`out << "shade:";`,
		`out << value.shade;`,
		"friend std::ostream &operator<<(std::ostream &s, const t &t);",
		`case 0: s << "req:" << variantstructout::t::unwrap< variantstructout::req >(t); break;`,
		"out = t(0, new t::twrap<req>(loc__r));",
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
	if !strings.Contains(out.Impl, "ivy_assert((v.tag == 0 && t::unwrap< a >(v) == av)") {
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
	if !strings.Contains(out.Impl, "ivy_assert((v.tag == 0)") {
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
	// Python emit_quant emits the cast loop form for EnumeratedSort
	// (ivy_to_cpp.py:1706-1708): `for (T X = (T)0; (int) X < N; X = ...)`.
	if !strings.Contains(got, "for (color X = (color)0; (int) X < 2; X = (color)(((int)X) + 1))") || !strings.Contains(got, "return true;") {
		t.Fatalf("unexpected quantifier code:\n%s", got)
	}
}

func TestEmitExprExistsUsesExtensionalRelationMap(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
relation marked(N:node)
after init {
    marked(X) := false;
}
action check = {
    assert exists X:node. marked(X)
}
`)
	out, err := Generate(mod, Config{ClassName: "extq"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"for (auto it = marked.memo.begin(), en = marked.memo.end(); it != en; ++it)",
		"if (!it->second) continue;",
		"int X = it->first;",
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
after init {
    edge(X, Y) := false;
}
action check = {
    assert exists X:node. edge(X,dst)
}
`)
	out, err := Generate(mod, Config{ClassName: "extq2"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"for (auto it = edge.memo.begin(), en = edge.memo.end(); it != en; ++it)",
		"int X = it->first.arg0;",
		"if (edge[__tup__int__int(X, dst)]) return true;",
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
after init {
    marked(X) := false;
    ok(X) := false;
}
action check = {
    assert forall X:node. marked(X) -> ok(X)
}
`)
	out, err := Generate(mod, Config{ClassName: "extfa"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"for (auto it = marked.memo.begin(), en = marked.memo.end(); it != en; ++it)",
		"if (!it->second) continue;",
		"int X = it->first;",
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

// TODO 008 tests: verify Python-faithful extensional-relation analysis.

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
	g := &Generator{Mod: mod}
	ext := g.extensionalRelations()
	if !ext["r"] {
		t.Fatalf("expected r to be in extensionalRelations, got %v", ext)
	}
	out, err := Generate(mod, Config{ClassName: "extinit"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "for (auto it = r.memo.begin(), en = r.memo.end(); it != en; ++it)") {
		t.Fatalf("missing r.memo iteration in:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
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
	g := &Generator{Mod: mod}
	ext := g.extensionalRelations()
	if ext["r"] {
		t.Fatalf("expected r to NOT be extensional (variable-arg true update), got %v", ext)
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
	g := &Generator{Mod: mod}
	ext := g.extensionalRelations()
	if ext["r"] {
		t.Fatalf("expected r to NOT be extensional (no init-to-false), got %v", ext)
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
`)
	out, err := Generate(mod, Config{ClassName: "extderived"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Either r or s should drive the memo loop. The derived definition
	// q is unfolded by matchExtensionalBoundExprs and the matcher takes
	// the first extensional atom it finds (r appears first in q's RHS).
	if !strings.Contains(out.Impl, "for (auto it = r.memo.begin(), en = r.memo.end(); it != en; ++it)") &&
		!strings.Contains(out.Impl, "for (auto it = s.memo.begin(), en = s.memo.end(); it != en; ++it)") {
		t.Fatalf("missing extensional memo loop for r or s after unfolding q:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
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
`)
	out, err := Generate(mod, Config{ClassName: "extmulti"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Single memo loop binds both X and Y from the pair key.
	for _, want := range []string{
		"for (auto it = r.memo.begin(), en = r.memo.end(); it != en; ++it)",
		"int X = it->first.arg0;",
		"int Y = it->first.arg1;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in multi-variable extensional quantifier:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
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
`)
	out, err := Generate(mod, Config{ClassName: "extpol"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Forall with `~r(X) | p(X)` : matcher should descend through Or
	// (with !exists) and Not (flipping to exists=true), find r(X) as
	// an extensional bound, and iterate r.memo.
	if !strings.Contains(out.Impl, "for (auto it = r.memo.begin(), en = r.memo.end(); it != en; ++it)") {
		t.Fatalf("missing r.memo iteration for polarity-flipped forall:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
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
	out, err := Generate(mod, Config{ClassName: "extifsomederived"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Unfolding q(x) should reveal r(x); if-some should iterate r.memo.
	if !strings.Contains(out.Impl, "for (auto it = r.memo.begin(), en = r.memo.end(); it != en; ++it)") {
		t.Fatalf("missing r.memo iteration in if-some-via-derived:\n%s", out.Impl)
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
	for _, want := range []string{"for (color X = (color)0; (int) X < 2; X = (color)(((int)X) + 1))", "if ((X == green)) return X;", "return red;"} {
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
		"if (v.tag == 0)",
		"a Q = t::unwrap< a >(v);",
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
	for _, want := range []string{"for (color X = (color)0; (int) X < 2; X = (color)(((int)X) + 1))", "if ((X == green)) return X;", "return red;"} {
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
		"if (v.tag == 0)",
		"a Q = t::unwrap< a >(v);",
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
	if err == nil || !strings.Contains(err.Error(), "thunk reached emit") {
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
			want: "thunk reached emit",
		},
		{
			name: "instantiate",
			act:  goivy.NewInstantiateAction(goivy.NewConst("schema", goivy.TopS)),
			want: "instantiate reached emit",
		},
		{
			name: "ranking",
			act:  goivy.NewRanking(goivy.NewConst("r", goivy.TopS)),
			want: "ranking reached emit",
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

// TestEmitCrashEmitsNothing locks down TODO 026's Crash fix. Python
// ivy_to_cpp.py:3888-3891 binds CrashAction.emit to `pass`, so reaching
// emit must produce zero C++ output and zero errors. The Go branch used
// to emit `std::abort();`, which diverged from Python.
func TestEmitCrashEmitsNothing(t *testing.T) {
	crash := goivy.NewCrashAction(goivy.NewConst("anything", goivy.TopS))
	var w cppWriter
	g := &Generator{}
	g.emitAction(&w, crash)
	if got := strings.TrimSpace(w.String()); got != "" {
		t.Fatalf("emit_crash should produce no output (Python pass), got %q", got)
	}
	if len(g.errs) != 0 {
		t.Fatalf("emit_crash should not record errors, got %v", g.errs)
	}
}

// TestEmitActionMatrixCoverage is TODO 026's Python action-class matrix
// in test form: every Python Action subclass that may reach emit is
// constructed and routed through (*Generator).emitAction to confirm we
// landed on the right branch — either a successful emit, a deliberate
// no-op, or an `unsupported`-style error that names the Python class.
//
// This is a coverage probe, not a code-shape probe. Per-class shape
// tests (TestEmitAssignNullaryAndIndexed, TestEmitIfWhileChoice,
// TestEmitNativeActionAntiquotes, ...) lock down the C++ output for
// the `emits` rows; this matrix proves no class falls through Go's
// default branch.
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
	green := goivy.NewConst("green", enum)
	cases := []struct {
		name   string
		build  func() goivy.Action
		want   string // emits, noop, errors
		errSub string // for `errors`: required substring in error message
	}{
		// has-emit Python rows: Go must produce non-empty output and zero errors.
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
		}, want: emits},
		{name: "ReturnAction", build: func() goivy.Action {
			return goivy.NewReturnAction()
		}, want: emits},
		// has-emit but no actual output for this fixture:
		{name: "IgnoreAction", build: func() goivy.Action {
			return goivy.NewIgnoreAction()
		}, want: noop},
		// Crash mirrors Python `pass` — no output, no errors.
		{name: "CrashAction", build: func() goivy.Action {
			return goivy.NewCrashAction(goivy.NewConst("anything", goivy.TopS))
		}, want: noop},
		// no-Python-emit rows: must error and name the Python class.
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
			act := tc.build()
			var w cppWriter
			g := &Generator{Mod: mod, currentReturns: []*goivy.Const{flag}}
			g.emitAction(&w, act)
			out := strings.TrimSpace(w.String())
			switch tc.want {
			case emits:
				if out == "" {
					t.Fatalf("%s: expected non-empty C++ output, got empty", tc.name)
				}
				if len(g.errs) != 0 {
					t.Fatalf("%s: expected no errors for has-emit class, got %v", tc.name, g.errs)
				}
				if strings.Contains(out, "/* ivy2cpp:") {
					t.Fatalf("%s: emit output looks like an unsupported-comment fallthrough:\n%s", tc.name, out)
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

// TODO 024: mirror Python check_member_names (ivy_to_cpp.py:1830-1834).
// The Go check lives in (*Generator).checkMemberNames and rejects modules
// whose generated C++ class name would collide with a varName-lowered
// signature symbol, sort, or action.

func TestCheckMemberNamesRejectsActionCollision(t *testing.T) {
	mod := goivy.New()
	mod.Name = "modname"
	mod.Actions.Set("foo", goivy.NewSequence())
	_, err := Generate(mod, Config{ClassName: "foo"})
	if err == nil || !strings.Contains(err.Error(), "cannot create C++ class foo with member foo") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckMemberNamesRejectsSymbolCollision(t *testing.T) {
	mod := goivy.New()
	mod.Name = "modname"
	if _, err := mod.Sig.AddSymbol("bar", goivy.Boolean); err != nil {
		t.Fatalf("AddSymbol: %v", err)
	}
	_, err := Generate(mod, Config{ClassName: "bar"})
	if err == nil || !strings.Contains(err.Error(), "cannot create C++ class bar with member bar") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckMemberNamesRejectsSortCollision(t *testing.T) {
	mod := goivy.New()
	mod.Name = "modname"
	if err := mod.Sig.AddSort(&goivy.UninterpretedSort{Name: "baz"}); err != nil {
		t.Fatalf("AddSort: %v", err)
	}
	_, err := Generate(mod, Config{ClassName: "baz"})
	if err == nil || !strings.Contains(err.Error(), "cannot create C++ class baz with member baz") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckMemberNamesAllowsDistinctClassname(t *testing.T) {
	mod := goivy.New()
	mod.Name = "modname"
	mod.Actions.Set("foo", goivy.NewSequence())
	_, err := Generate(mod, Config{ClassName: "qux"})
	if err != nil && strings.Contains(err.Error(), "cannot create C++ class") {
		t.Fatalf("unexpected member-name collision error: %v", err)
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
	// Mirrors Python emit_choice (ivy_to_cpp.py:3976-3994): one int temp
	// populated by mk_nondet, then an if/else chain (no `switch`).
	// Python puts `}` and `else` on separate lines, which the Go writer
	// also does.
	for _, want := range []string{"int __ivy_branch", `___ivy_choose(0, "___branch"`, "if (__ivy_branch", "else {", "while (cond)", "if (cond)", "flag = true;"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "switch (___ivy_choose") {
		t.Fatalf("emit_choice should no longer use switch (Python uses if/else chain):\n%s", got)
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
	for _, want := range []string{"bool __ivy_some", "for (color loc__c = (color)0; (int) loc__c < 2; loc__c = (color)(((int)loc__c) + 1))", "if (!__ivy_some", "saved = loc__c;", "saved = red;"} {
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
after init {
    marked(X) := false;
}
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
		"for (auto it = marked.memo.begin(), en = marked.memo.end(); it != en; ++it)",
		"if (!it->second) continue;",
		"int loc__x = it->first;",
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
	// Python emit_assign uses a two-phase temporary for quantified
	// assignments (ivy_to_cpp.py:3725-3764) so RHS reads see the
	// pre-assignment value. Match that structure.
	for _, want := range []string{
		"void paint::__init()",
		"bool __ivy_tmp1[2];",
		"for (color C : {red, green})",
		"__ivy_tmp1[C] = false;",
		"marked[C] = __ivy_tmp1[C];",
	} {
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
	// Two-phase quantified assignment per Python emit_assign.
	for _, want := range []string{
		"typedef unsigned idx;",
		"for (unsigned I = 0; I <= 2; I++)",
		"__ivy_tmp1[I] = false;",
		"marked[I] = __ivy_tmp1[I];",
	} {
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
	// Python `get_bounds` pushes the cardinality-derived bound first
	// (ivy_to_cpp.py:3322-3323): los="0", his=card. For RangeSort with
	// goivy's storage-dimension card (`ub+1`), the bound is `< 4` here.
	if !strings.Contains(got, "for (unsigned I = 0; I < 4; I++)") || !strings.Contains(got, "return true;") {
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
	// compileIvySource runs with create_isolate=false so the parser
	// leaves every action public. Trim to mirror what create_isolate
	// would do (ivy_isolate.py:1674-1678) — only the exported action
	// stays public. Without this, split (multi-return) gets the public
	// ValueType-only annotation, which Python rejects with IvyError
	// (ivy_to_cpp.py:1565-1567).
	mod.PublicActions.Set("split", false)
	out, err := Generate(mod, Config{ClassName: "multi"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Python annotation for private `split(c) returns (out, good)`:
	//   ptypes=[ValueType], rtypes=[ValueType, ReturnRefType{Pos:1}]
	// → signature returns `out` by value, `good` as trailing ref.
	// → call site: saved = split(green, ok);
	for _, want := range []string{"color split(color c, bool& good)", "out = c;", "good = true;", "return out;", "saved = split(green, ok);"} {
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
	// Mirrors Python local_start + mk_nondet_sym (ivy_to_cpp.py:3893-3917):
	// declare uninitialized, then ___ivy_choose-initialize, then run the body.
	// The parser prefixes the local's name with "loc:" so the choose label
	// is "loc:tmp" — this preserves the original symbol identity in nondet
	// gen/test traces.
	for _, want := range []string{"{", "color loc__tmp;", `loc__tmp = (color)___ivy_choose(0, "loc:tmp"`, "loc__tmp = green;", "saved = loc__tmp;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	if strings.Contains(out.Impl, "color loc__tmp = red;") {
		t.Fatalf("local should not be zero-initialized — Python uses nondet init:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

// TestGeneratedHavocReportsUnsupported asserts that a HavocAction
// reaching emit fails generation. Mirrors Python emit_havoc
// (ivy_to_cpp.py:3768-3773) which `assert False`s — havoc should be
// eliminated upstream and never reach C++ emission.
func TestGeneratedHavocReportsUnsupported(t *testing.T) {
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
	_, err := Generate(mod, Config{ClassName: "havocs"})
	if err == nil {
		t.Fatalf("Generate should fail when HavocAction reaches emit")
	}
	if !strings.Contains(err.Error(), "havoc") {
		t.Fatalf("Generate error should mention havoc, got: %v", err)
	}
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
	// Post-TODO 021: the JSON key is taken from the with-clause name
	// (here `flag` because the synthetic DebugAction has no parallel
	// WithNames slot and we fall back to ExprName). The value print is
	// emitted on its own line by emitPrintExpr so quantified values can
	// open loops around it.
	for _, want := range []string{`"event\" : \"tick\"`, `"flag\" : `, `std::cout << (flag);`} {
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

// TestGeneratedEnvActionEmitsChoice asserts that an EnvAction with
// multiple branches lowers to the if/else chain over `___ivy_choose`,
// matching Python emit_choice. EnvAction shares LogicChoiceAction's
// dispatch, so the surface form is identical.
func TestGeneratedEnvActionEmitsChoice(t *testing.T) {
	mod := goivy.New()
	mod.Name = "envcase"
	mod.Relations.Set("flag", goivy.Boolean)
	assign := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("true", goivy.Boolean))
	clear := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("false", goivy.Boolean))
	mod.Actions.Set("step", goivy.NewEnvActionOn(goivy.NewActionsConfig(), assign, clear))
	out, err := Generate(mod, Config{ClassName: "envcase"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{"int __ivy_branch", `___ivy_choose(0, "___branch"`, "if (__ivy_branch", "flag = true;", "flag = false;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	if strings.Contains(out.Impl, "switch (___ivy_choose") {
		t.Fatalf("emit_choice should no longer use switch:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

// TestGeneratedBindOldsReportsUnsupported asserts that a
// BindOldsAction reaching emit fails generation. Python defines no
// `emit_bind_olds` (it is eliminated upstream by
// `ivy_transrel.bind_olds_action`), so reaching emit is a bug.
func TestGeneratedBindOldsReportsUnsupported(t *testing.T) {
	mod := goivy.New()
	mod.Name = "bindcase"
	mod.Relations.Set("flag", goivy.Boolean)
	assign := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("true", goivy.Boolean))
	mod.Actions.Set("step", goivy.NewBindOldsAction(assign))
	_, err := Generate(mod, Config{ClassName: "bindcase"})
	if err == nil {
		t.Fatalf("Generate should fail when BindOldsAction reaches emit")
	}
	if !strings.Contains(err.Error(), "bindolds") {
		t.Fatalf("Generate error should mention bindolds, got: %v", err)
	}
}

// TestGeneratedChoiceUsesIfElseChain exercises the three-branch
// LogicChoiceAction path. Mirrors Python emit_choice
// (ivy_to_cpp.py:3976-3994): one int temp populated by mk_nondet, then
// `if (tmp == 0) {...} else if (tmp == 1) {...} else {...}`.
func TestGeneratedChoiceUsesIfElseChain(t *testing.T) {
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
	red := goivy.NewConst("red", color)
	green := goivy.NewConst("green", color)
	blue := goivy.NewConst("blue", color)
	saved := goivy.NewConst("saved", color)
	branchRed := goivy.NewAssignAction(saved, red)
	branchGreen := goivy.NewAssignAction(saved, green)
	branchBlue := goivy.NewAssignAction(saved, blue)
	mod.Actions.Set("step", goivy.NewChoiceActionOn(goivy.NewActionsConfig(), branchRed, branchGreen, branchBlue))
	out, err := Generate(mod, Config{ClassName: "choice3"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"int __ivy_branch",
		`___ivy_choose(0, "___branch"`,
		"if (__ivy_branch",
		"saved = red;",
		"saved = green;",
		"saved = blue;",
		"else {",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	if strings.Contains(out.Impl, "switch (___ivy_choose") {
		t.Fatalf("emit_choice should no longer use switch (Python uses if/else chain):\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

// TestGeneratedLocalActionUsesNondet exercises scalar local nondet
// initialization. Mirrors Python local_start + mk_nondet_sym
// (ivy_to_cpp.py:3893-3917 + 196-214) for the empty-domain case.
func TestGeneratedLocalActionUsesNondet(t *testing.T) {
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
	cfg := goivy.NewActionsConfig()
	local := goivy.NewConst("x", color)
	body := goivy.NewAssignAction(goivy.NewConst("saved", color), local)
	mod.Actions.Set("step", goivy.NewLocalActionOn(cfg, "test", local, body))
	out, err := Generate(mod, Config{ClassName: "locact"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"color x;",
		`x = (color)___ivy_choose(0, "x"`,
		"saved = x;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	if strings.Contains(out.Impl, "color x = red;") {
		t.Fatalf("local should not be zero-initialized:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

// TestGeneratedLocalFunctionUsesNondetLoop exercises mk_nondet_sym's
// bounded-array branch: a function-typed local with a small enum
// domain. Each cell is nondet-initialized inside a per-domain loop.
func TestGeneratedLocalFunctionUsesNondetLoop(t *testing.T) {
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
	relSort, err := goivy.NewFunctionSort(color, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	cfg := goivy.NewActionsConfig()
	local := goivy.NewConst("marked", relSort)
	body := goivy.NewSequence()
	mod.Actions.Set("step", goivy.NewLocalActionOn(cfg, "test", local, body))
	out, err := Generate(mod, Config{ClassName: "locfunc"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"bool marked[2];",
		"for (color X0 : {red, green}) {",
		`marked[X0] = (bool)___ivy_choose(0, "marked"`,
	} {
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
	for _, want := range []string{"paramcase::paramcase(paramcase::color initial)", "this->initial = initial;", "void paramcase::__init()", "saved = initial;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	if strings.Contains(out.Impl, "initial = red;") {
		t.Fatalf("constructor should preserve parameter assignment:\n%s", out.Impl)
	}
	if strings.Contains(out.Impl, "paramcase::paramcase(paramcase::color initial) {\n    __init();") {
		t.Fatalf("constructor should not run __init directly:\n%s", out.Impl)
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
	// Param has no default — declared as positional; main constructs ivy
	// with the parsed local `p__initial`.
	for _, want := range []string{
		"paramrepl::color p__initial",
		`arg_values[0] = parse_value(args[0], pos);`,
		`p__initial = _arg<paramrepl::color>(arg_values, 0, 2);`,
		"paramrepl_repl ivy(p__initial);",
		"ivy.__init();",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestInconsistentModelInitialStateReturnsError(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
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
	_, err = Generate(mod, Config{ClassName: "badinit"})
	if err == nil || !strings.Contains(err.Error(), "inconsistent") {
		t.Fatalf("unexpected error: %v", err)
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
	_, err = Generate(mod, Config{ClassName: "parambad"})
	if err == nil || !strings.Contains(err.Error(), `stripped parameter "initial"`) {
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
	for _, want := range []string{"marked[red] = false;", "marked[green] = true;"} {
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
	for _, want := range []string{"marked[red] = false;", "marked[green] = true;"} {
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
	for _, want := range []string{"marked[red] = true;", "marked[green] = false;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestInitialStateUsesRelevantDefinitions(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
relation is_green(C:color)
definition is_green(C:color) = C = green
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
	out, err := Generate(mod, Config{ClassName: "initdef"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "saved = green;") {
		t.Fatalf("relevant definition should constrain saved to green:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestTargetGenAndTestEmitInitialStateSolverPath(t *testing.T) {
	for _, target := range []string{"gen", "test"} {
		t.Run(target, func(t *testing.T) {
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
			mod.LabeledInits = []*goivy.LabeledFormula{mod.Cfg.AstCfg.NewLabeledFormula(nil, eq)}
			out, err := Generate(mod, Config{Target: target, ClassName: "initgenpath"})
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			for _, want := range []string{
				`add((mk_apply_expr("saved", {}) == int_to_z3("color", 1)));`,
				`obj.saved = (initgenpath::color)eval_apply("saved");`,
				"obj.___ivy_gen = this;",
				"obj.__init();",
			} {
				if !strings.Contains(out.Impl, want) {
					t.Fatalf("%s missing %q in impl:\n%s", target, want, out.Impl)
				}
			}
			assertNoUnsupportedCPP(t, out)
			compileGeneratedCPP(t, out)
		})
	}
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
	if !strings.Contains(out.Header, "typedef std::vector<unsigned> vec;") {
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
	// Parameter has no default — declared as positional. Native vec
	// requires no special handling beyond _arg<>.
	for _, want := range []string{
		"nativeparam::vec p__initial",
		"nativeparam ivy(p__initial);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
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
	// Python emit_some_action wraps the body in `val = rhs; return val;`
	// even when rhs is a native expression (ivy_to_cpp.py:1592-1625).
	// `idx` is registered in NativeTypes via `interpret idx -> <<< int >>>`
	// (compiler_decl.go:1047), so the derived definition's ptype policy
	// gives ConstRefType for its parameters (annotateAction, ptype.go:103).
	for _, want := range []string{"bool lt(const idx& x, const idx& y);", "bool nativedef::lt(const nativedef::idx& x, const nativedef::idx& y)", "val = x < y;", "return val;", "ivy_assert(lt(x, y)"} {
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

func TestNativeInlineTagEmitsAfterClass(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
<<< inline
inline int inlined_helper() { return 7; }
>>>
action step = {
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "inlinecase"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	classEnd := strings.Index(out.Header, "};")
	helperAt := strings.Index(out.Header, "inline int inlined_helper()")
	if classEnd < 0 {
		t.Fatalf("missing class closing brace in header:\n%s", out.Header)
	}
	if helperAt < 0 {
		t.Fatalf("inline native body missing from header:\n%s", out.Header)
	}
	if helperAt < classEnd {
		t.Fatalf("inline native should appear AFTER closing class brace.\nhelper@%d class}@%d\n%s", helperAt, classEnd, out.Header)
	}
	compileGeneratedCPP(t, out)
}

func TestNativeUnknownTagReportsError(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
<<< bogus_tag
int unused_member_for_validation;
>>>
action step = {
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "badtag"})
	if err == nil {
		t.Fatalf("expected error for unknown native tag, got nil; header:\n%s\nimpl:\n%s", out.Header, out.Impl)
	}
	if !strings.Contains(err.Error(), "syntax error at token bogus_tag") {
		t.Fatalf("error should mention the offending tag, got: %v", err)
	}
}

func TestNativeEncodeTagEmitsImplBody(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type mysort
<<< encode mysort
/* encode body for mysort */
>>>
action step = {
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "enc"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got := strings.Count(out.Impl, "/* encode body for mysort */"); got != 1 {
		t.Fatalf("encode body count=%d, want 1; impl:\n%s", got, out.Impl)
	}
}

func TestNativeCallbackThunkEmitted(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..3}
action handler(i:idx) = {
}
<<< member
int storage_for_`+"`handler`"+`;
>>>
action step = {
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "cbcase"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	wants := []string{
		"struct thunk__handler {",
		"cbcase *__ivy;",
		"thunk__handler(cbcase *__ivy)",
		"void operator()(unsigned i) const {",
		"__ivy->handler(i);",
	}
	for _, w := range wants {
		if !strings.Contains(out.Impl, w) {
			t.Fatalf("missing %q in impl:\n%s", w, out.Impl)
		}
	}
}

func TestNativeActionCallsThunkConstructor(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..3}
action handler(i:idx) = {
}
action install = {
    <<<
        register(`+"`handler`"+`);
    >>>
}
export install
`)
	out, err := Generate(mod, Config{ClassName: "cbact"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "register(thunk__handler(this));") {
		t.Fatalf("native action body should call thunk constructor:\n%s", out.Impl)
	}
	if !strings.Contains(out.Impl, "struct thunk__handler {") {
		t.Fatalf("thunk struct should be emitted for native-action callback:\n%s", out.Impl)
	}
}

func TestNativeInlineDedup(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
<<< inline
inline int dup_inline() { return 1; }
>>>
<<< inline
inline int dup_inline() { return 1; }
>>>
action step = {
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "indedup"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got := strings.Count(out.Header, "inline int dup_inline()"); got != 1 {
		t.Fatalf("inline native dedup expected 1, got %d; header:\n%s", got, out.Header)
	}
}

func TestNativeDuplicateEncodeErrors(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type mysort
<<< encode mysort
/* first encoding */
>>>
<<< encode mysort
/* second encoding */
>>>
action step = {
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "encdup"})
	if err == nil {
		t.Fatalf("expected duplicate encode error; header:\n%s\nimpl:\n%s", out.Header, out.Impl)
	}
	if !strings.Contains(err.Error(), "duplicate encoding for sort mysort") {
		t.Fatalf("error should mention duplicate encoding, got: %v", err)
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
	// Python-style cmd_reader loop: runner_cmd_reader.read() in a while
	// loop dispatches each line through parse_command + _arg<T>.
	for _, want := range []string{
		"class runner_cmd_reader : public stdin_reader",
		"runner_cmd_reader *cr = new runner_cmd_reader(ivy);",
		"while (!cr->eof())",
		"cr->read();",
		"parse_command(cmd, action, args);",
	} {
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
	for _, want := range []string{
		`if (action == "set")`,
		`check_arity(args, 1, action);`,
		`ivy.set(_arg<runner::color>(args, 0, 2));`,
	} {
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
	// Python-style dispatch via _arg<T>(args, idx, csortcard).
	// `idx` has bound 3 (range 0..2).
	for _, want := range []string{
		`#include "ivy_repl.hpp"`,
		`#include "ivy_value.hpp"`,
		`ivy.set(_arg<runner::color>(args, 0, 2), _arg<bool>(args, 1, 2), _arg<unsigned>(args, 2, 3));`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in repl output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

// Phase 1: per-enum operator<<, _arg<T>, __ser<T>, __deser<T>
// specializations alongside the existing ad-hoc `ivy2cpp_parse_TYPE`
// family. Mirrors Python ivy_to_cpp.py:2497-2510 and 2634-2652.
func TestEmitsArgSpecForEnum(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green, blue}
individual saved : color
action set(c:color) = { saved := c }
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		// Forward decls in the impl preamble.
		`std::ostream &operator<<(std::ostream &s, const runner::color &t);`,
		`runner::color _arg<runner::color>(std::vector<ivy_value> &args, unsigned idx, long long bound);`,
		`void __ser<runner::color>(ivy_ser &res, const runner::color &);`,
		`void __deser<runner::color>(ivy_deser &inp, runner::color &res);`,
		// Definitions.
		`std::ostream &operator<<(std::ostream &s, const runner::color &t) {`,
		`if (t == runner::red) s << "red";`,
		`if (t == runner::green) s << "green";`,
		`if (t == runner::blue) s << "blue";`,
		`void __ser<runner::color>(ivy_ser &res, const runner::color &t) {`,
		`__ser(res, (int)t);`,
		`runner::color _arg<runner::color>(std::vector<ivy_value> &args, unsigned idx, long long bound) {`,
		`ivy_value &arg = args[idx];`,
		`if (arg.atom.size() == 0 || arg.fields.size() != 0) throw out_of_bounds(idx, arg.pos);`,
		`if (arg.atom == "red") return runner::red;`,
		`if (arg.atom == "green") return runner::green;`,
		`if (arg.atom == "blue") return runner::blue;`,
		`throw out_of_bounds("bad value: " + arg.atom, arg.pos);`,
		`void __deser<runner::color>(ivy_deser &inp, runner::color &res) {`,
		`__deser(inp, __res);`,
		`res = (runner::color)__res;`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl output:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
}

// Numeric enums (extension is integer literals) are mapped to `int`
// and must NOT get a template specialization — that would collide with
// `_arg<int>` from ivy_value.hpp. Mirrors Python's `is_numeric_range`
// gate in ctype_remaining_cases.
func TestSkipsArgSpecForNumericEnum(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..2}
individual seen : idx
action set(i:idx) = { seen := i }
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// `idx` is a numeric range/enum; no per-sort _arg<runner::idx>
	// template should be emitted (it would conflict with _arg<unsigned>
	// from ivy_value.hpp via the typedef).
	for _, banned := range []string{
		`_arg<runner::idx>`,
		`__ser<runner::idx>`,
		`__deser<runner::idx>`,
	} {
		if strings.Contains(out.Impl, banned) {
			t.Fatalf("unexpected %q in impl output:\n%s", banned, out.Impl)
		}
	}
}

// Phase 2: per-module cmd_reader subclass. Mirrors Python
// emit_repl_boilerplate1a + emit_repl_boilerplate2.
func TestReplEmitsCmdReader(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action set(c:color) = { assert c = green }
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"class runner_cmd_reader : public stdin_reader {",
		"runner_repl &ivy;",
		"runner_cmd_reader(runner_repl &_ivy) : ivy(_ivy) {",
		"virtual void process(const std::string &cmd) {",
		"parse_command(cmd, action, args);",
		"ivy.__lock();",
		`if (action == "set") {`,
		"check_arity(args, 1, action);",
		`ivy.set(_arg<runner::color>(args, 0, 2));`,
		"ivy.__unlock();",
		`std::cerr << "undefined action: " << action << std::endl;`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in cmd_reader impl:\n%s", want, out.Impl)
		}
	}
}

// Phase 2: cmd_reader catches syntax_error / out_of_bounds / bad_arity.
// Python emit_repl_boilerplate2 (ivy_to_cpp.py:4160-4187).
func TestReplCatchesSyntaxOutOfBoundsBadArity(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"catch (syntax_error &err)",
		`std::cerr << "line " << lineno << ":" << err.pos << ": syntax error"`,
		"catch (out_of_bounds &err)",
		`std::cerr << "line " << lineno << ":" << err.pos << ": " << err.txt << " bad value"`,
		"catch (bad_arity &err)",
		`std::cerr << "action " << err.action << " takes " << err.num << " input parameters"`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in cmd_reader catches:\n%s", want, out.Impl)
		}
	}
}

// Phase 2: REPL server mode when no public actions exist. Python
// emit_repl_boilerplate3server (ivy_to_cpp.py:4243-4263). Build a
// module with no exports directly so the init mixin isn't auto-promoted.
func TestReplServerModeWhenNoPublicActions(t *testing.T) {
	mod := goivy.New()
	mod.Name = "srv"
	out, err := Generate(mod, Config{Target: "repl", ClassName: "srv"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"for (unsigned i = 0; true; i++) {",
		"if (i >= ivy.thread_ids.size()) {",
		"pthread_join(tid, NULL);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl server-mode main:\n%s", want, out.Impl)
		}
	}
	if strings.Contains(out.Impl, "while (!cr->eof())") {
		t.Fatalf("server mode should not instantiate cmd_reader loop:\n%s", out.Impl)
	}
}

// Phase 2: Config.Trace wraps action dispatch with trace prelude/postlude.
// Python ivy_to_cpp.py:2685-2690.
func TestReplTraceWrapsActionCalls(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action set(c:color) = { assert c = green }
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		`__ivy_out << "set(" << _arg<runner::color>(args, 0, 2) << ") {" << std::endl;`,
		`ivy.set(_arg<runner::color>(args, 0, 2));`,
		`__ivy_out << "}" << std::endl;`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in traced repl dispatch:\n%s", want, out.Impl)
		}
	}
}

// Phase 6: the legacy `ivy_go_repl.hpp` shim has been removed. No
// emitted output should still reference it, and the include directory
// should no longer contain it.
func TestReplGoReplShimRemoved(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {}
export step
`)
	for _, target := range []string{"repl", "test", "gen", "class"} {
		out, err := Generate(mod, Config{Target: target, ClassName: "shim"})
		if err != nil {
			t.Fatalf("Generate(%s): %v", target, err)
		}
		if strings.Contains(out.Header+out.Impl, "ivy_go_repl.hpp") {
			t.Fatalf("%s target still references ivy_go_repl.hpp:\n%s", target, out.Impl)
		}
		if strings.Contains(out.Impl, "ivy2cpp_parse_command") || strings.Contains(out.Impl, "ivy2cpp_dispatch") {
			t.Fatalf("%s target still uses ivy2cpp_parse_command / ivy2cpp_dispatch helpers:\n%s", target, out.Impl)
		}
	}
}

// Phase 6: when Config.Trace is set, the assert/assume overrides emit
// the closing `}` so traced output stays balanced. Python ivy_to_cpp.py:
// emit_repl_boilerplate1 CLOSE_TRACE substitution (4099-4101).
func TestReplTraceClosesAssertBraces(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = { assert false }
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "tr", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Both ivy_assert and ivy_assume should emit the closing `}` when
	// trace is on, after the assertion_failed/assumption_failed lines.
	count := strings.Count(out.Impl, `__ivy_out << "}" << std::endl;`)
	if count < 2 {
		t.Fatalf("expected closing `}` in both ivy_assert and ivy_assume under Trace, got %d:\n%s", count, out.Impl)
	}
}

// Phase 5: emit_repl_boilerplate3test multi-run + weighted random
// loop shape. Python ivy_to_cpp.py:4265-4467.
func TestTestMainWeightedGeneratorLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = { saved := c }
export set
attribute set.weight = "3.0"
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "wt"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		// Multi-run outer loop.
		"for (int runidx = 0; runidx < runs; runidx++)",
		"initializing = true;",
		"ivy._generating = false;",
		"initializing = false;",
		// Init generator + per-action generators.
		"init_gen my_init_gen(ivy);",
		"my_init_gen.generate(ivy);",
		"std::vector<double> weights;",
		"set_gen set_generator(ivy);",
		"weights.push_back(3);",
		"double totalweight = 3;",
		"int num_gens = 1;",
		// Random choice + do_over.
		"double frnd = 0.0;",
		"bool do_over = false;",
		"for (int cycle = 0; cycle < test_iters; cycle++)",
		"if (frnd < totalweight) {",
		"ivy._generating = true;",
		// Per-index dispatch instead of virtual gen *.
		"case 0: sat = set_generator.generate(ivy); break;",
		"case 0: set_generator.execute(ivy); break;",
		// select() branch for readers/timers.
		"FD_ZERO(&rdfds);",
		"int foo = select(maxfds + 1, &rdfds, 0, 0, &timeout);",
		// Run epilogue.
		`__ivy_out << "test_completed" << std::endl;`,
		"if (runidx == runs - 1)",
		"nanosleep(&ts, NULL);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in test main:\n%s", want, out.Impl)
		}
	}
}

// Phase 5: ext:_finalize gets invoked at end of the loop if exported.
// Python ivy_to_cpp.py:4303 (FINALIZE substitution).
func TestTestMainHonorsFinalize(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action _finalize = {}
export _finalize
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "fin"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "ivy.__lock(); ivy.ext___finalize(); ivy.__unlock();") {
		t.Fatalf("expected finalize hook in test main:\n%s", out.Impl)
	}
	// _finalize is filtered out of the generator list (Python 4282-4284).
	if strings.Contains(out.Impl, "_finalize_gen") {
		t.Fatalf("_finalize should not be in the action generator list:\n%s", out.Impl)
	}
}

// Phase 4: a parameter with a default value is applied via the
// emit_value_parser block and overridable from argv via `param=value`.
// Mirrors Python ivy_to_cpp.py:2706-2735.
func TestReplMainParsesParamWithDefault(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter pick : color = red
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"runner::color p__pick",
		`arg_values[0] = parse_value("red", pos);`,
		`p__pick = _arg<runner::color>(arg_values, 0, 2);`,
		`if (param == "pick") {`,
		`arg_values[0] = parse_value(value, pos);`,
		"runner_repl ivy(p__pick);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in main:\n%s", want, out.Impl)
		}
	}
}

// Phase 4: a positional parameter (no default) consumes a positional
// argv slot via parse_value + _arg<T>. Python ivy_to_cpp.py:2779-2832.
func TestReplMainParsesPositionalParam(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter pick : color
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"runner::color p__pick",
		"if (argc == 3) {",
		"if (argc != 2) {",
		`std::cerr << "usage: runner pick\n";`,
		"std::vector<ivy_value> arg_values(1);",
		`arg_values[0] = parse_value(args[0], pos);`,
		`p__pick = _arg<runner::color>(arg_values, 0, 2);`,
		"runner_repl ivy(p__pick);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in main:\n%s", want, out.Impl)
		}
	}
}

// Phase 4: special argv keys (out/iters/runs/seed/delay/wait/modelfile).
// Mirrors Python ivy_to_cpp.py:2738-2770.
func TestReplMainAcceptsSpecialOptions(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		`if (param == "out") {`,
		"__ivy_out.open(value.c_str());",
		`else if (param == "iters") { test_iters = atoi(value.c_str()); }`,
		`else if (param == "runs") { runs = atoi(value.c_str()); }`,
		`else if (param == "seed") { seed = atoi(value.c_str()); }`,
		`else if (param == "delay") { sleep_ms = atoi(value.c_str()); }`,
		`else if (param == "wait") { final_ms = atoi(value.c_str()); }`,
		`else if (param == "modelfile") {`,
		"__ivy_modelfile.open(value.c_str());",
		`std::cerr << "unknown option: " << param << std::endl;`,
		"srand(seed);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in main:\n%s", want, out.Impl)
		}
	}
}

// Phase 4: function-sorted positional parameter. Mirrors Python
// ivy_to_cpp.py:2803-2826 (make_function_app over arg.fields).
//
// The ivy1.7 surface syntax does not allow `parameter` for function
// sorts directly — instead a "parameter" of relational/function sort is
// expressed via attribute. Build the module manually so we can exercise
// the make_function_app emission path.
func TestReplMainFunctionSortedParam(t *testing.T) {
	mod := goivy.New()
	mod.Name = "runner"
	color := &goivy.LogicEnumeratedSort{Name: "color", Extension: []string{"red", "green"}}
	mod.Sig.Sorts.Set("color", color)
	mod.SortOrder = append(mod.SortOrder, "color")
	slot := &goivy.UninterpretedSort{Name: "slot"}
	mod.Sig.Sorts.Set("slot", slot)
	mod.Sig.Interp["slot"] = &goivy.RangeSort{Name: "slot", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "1"}}
	mod.SortOrder = append(mod.SortOrder, "slot")
	// Function-sorted param: tbl : slot -> color
	fs, err := goivy.NewFunctionSort(slot, color)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	p := &goivy.Const{Name: "tbl", CSort: fs}
	mod.Params = append(mod.Params, p)
	mod.ParamDefaults = append(mod.ParamDefaults, nil)
	mod.Actions.Set("step", goivy.NewSequence())
	mod.PublicActions.Set("step", true)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"p__tbl",
		"ivy_value &arg = arg_values[0];",
		"if (arg.atom.size()) {",
		"if (arg.fields[i].fields.size() != 2) {",
		"p__tbl[_arg<unsigned>(arg.fields[i].fields, 0, 0)] = _arg<runner::color>(arg.fields[i].fields, 1, 0);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in function-sort main:\n%s", want, out.Impl)
		}
	}
}

// Phase 4: Winsock init block from emit_winsock_init.
// Python ivy_to_cpp.py:4189-4225.
func TestReplMainEmitsWinsockInit(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"#ifdef _WIN32",
		"WORD wVersionRequested;",
		"wVersionRequested = MAKEWORD(2, 2);",
		"err = WSAStartup(wVersionRequested, &wsaData);",
		"if (LOBYTE(wsaData.wVersion) != 2 || HIBYTE(wsaData.wVersion) != 2) {",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in main:\n%s", want, out.Impl)
		}
	}
}

// Phase 3: imports show up as callback methods on ClassName_repl that
// trace `< action(args)` to __ivy_out. Mirrors Python
// emit_repl_boilerplate1 (ivy_to_cpp.py:4107-4128).
func TestReplEmitsImportCallback(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action notify(c:color) = {}
import notify
action trigger = {
    call notify(red)
}
export trigger
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"class runner_repl : public runner",
		// Override of `notify` inside the repl subclass. Class-body
		// emission uses the unqualified type since the class is in
		// runner's scope (matches Python emit_method_decl).
		"virtual void notify(color c)",
		`__ivy_out << "< notify" << "(" << c << ")" << std::endl;`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl impl:\n%s", want, out.Impl)
		}
	}
}

// Phase 3: import callbacks that return a value prompt the user via
// ask_ret(__CARD__<sort>). Python ivy_to_cpp.py:4126-4127.
func TestReplImportCallbackReturnUsesAskRet(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action poll returns (out:color) = {}
import poll
action trigger returns (out:color) = {
    out := poll
}
export trigger
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"virtual color poll(",
		`__ivy_out << "< poll" << std::endl;`,
		"return ask_ret(__CARD__color);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl impl:\n%s", want, out.Impl)
		}
	}
	// __CARD__color is declared (set in cardinality initializers).
	if !strings.Contains(out.Header, "long long __CARD__color;") {
		t.Fatalf("missing __CARD__ declaration in header:\n%s", out.Header)
	}
}

// Phase 3: in test target, imported callbacks have empty bodies (no
// console prompt). Python ivy_to_cpp.py:4112-4114.
func TestReplImportCallbackEmptyForTest(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action notify = {}
import notify
action trigger = {
    call notify
}
export trigger
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "virtual void notify() {}") {
		t.Fatalf("test target should emit empty notify body:\n%s", out.Impl)
	}
}

// Plain uninterpreted sorts already get coverage via the primitive
// `_arg<int>` / `_arg<unsigned>` etc. specializations in ivy_value.hpp
// once the typedef alias is in place. We must NOT emit a duplicate
// per-sort template since C++ forbids redundant specializations on the
// same underlying type.
func TestSkipsArgSpecForPlainUninterpreted(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
individual saved : node
action set(n:node) = { saved := n; assert n = 7 }
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, banned := range []string{
		`_arg<runner::node>`,
		`__ser<runner::node>`,
		`__deser<runner::node>`,
	} {
		if strings.Contains(out.Impl, banned) {
			t.Fatalf("unexpected %q in impl output:\n%s", banned, out.Impl)
		}
	}
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
	// Plain uninterpreted sort `node` maps to `int` (no typedef emitted
	// since cppType returns "int" directly). The dispatcher uses the
	// stock `_arg<int>` template from ivy_value.hpp with bound 0
	// (unbounded — node has no defined cardinality).
	for _, want := range []string{
		`ivy.set(_arg<int>(args, 0, 0));`,
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
	// Python emits `_arg<runner::color>` specialization (in repl.go) and
	// dispatches via `_arg<runner::color>(args, 0, 2)`.
	for _, want := range []string{
		`runner::color _arg<runner::color>(std::vector<ivy_value> &args, unsigned idx, long long bound) {`,
		`if (arg.atom == "green") return runner::green;`,
		`ivy.set(_arg<runner::color>(args, 0, 2));`,
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
	// Python's ivy_repl.hpp provides parse_command + check_arity, which
	// must reject whitespace-separated args by requiring parentheses.
	support := readSupportHeader(t, "ivy_repl.hpp")
	for _, want := range []string{
		"inline void parse_command(const std::string &cmd, std::string &action, std::vector<ivy_value> &args)",
		"cmd[pos] == '('",
		"throw_syntax(pos);",
		"inline void check_arity(std::vector<ivy_value> &args, unsigned num, std::string &action)",
	} {
		if !strings.Contains(support, want) {
			t.Fatalf("shared repl support missing %q:\n%s", want, support)
		}
	}
	if !strings.Contains(out.Impl, "check_arity(args, 1, action);") {
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
		`#include "ivy_repl.hpp"`,
		"parse_command(cmd, action, args);",
		`runner::color __ivy_result = ivy.echo(_arg<runner::color>(args, 0, 2));`,
		`__ivy_out << "= " << __ivy_result << std::endl;`,
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
		// `operator<<` for the enum (from Phase 1) handles return printing.
		"std::ostream &operator<<(std::ostream &s, const runner::color &t) {",
		`runner::color __ivy_result = ivy.echo(_arg<runner::color>(args, 0, 2));`,
		`__ivy_out << "= " << __ivy_result << std::endl;`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in repl output:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestReplDispatchWritesMultipleReturns(t *testing.T) {
	// Mirrors Python ivy_to_cpp.py:1565-1567: exporting a multi-return
	// action is rejected with "cannot handle multiple output in exported
	// actions". The previous Go behavior accepted it; TODO 009 brings the
	// port back in line with Python.
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
action split(c:color) returns (out:color, good:bool) = {
    out := c;
    good := true
}
export split
`)
	_, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err == nil {
		t.Fatal("Generate should reject exporting a multi-return action")
	}
	if !strings.Contains(err.Error(), "cannot handle multiple output in exported actions: split") {
		t.Fatalf("expected multi-output rejection error, got: %v", err)
	}
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
	if !strings.Contains(out.Impl, `_arg<oracle::color>(args, 0, 2)`) {
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
	// After TODO 017 milestone 4 the gen path randomizes formal-param
	// inputs through the solver: each input gets g.randomize() in
	// generate() and __from_solver() reads its value back after solve().
	// The member declarations still use the Go-side var name (e.g. `c`,
	// `b`, `i`); the underlying Z3 symbol carries the `__fml:` prefix.
	for _, want := range []string{
		"genparams::color c;",
		"bool b;",
		"unsigned i;",
		`randomize("__fml:c", "color");`,
		`randomize("__fml:b", "bool");`,
		`randomize("__fml:i", "idx");`,
		`c = (genparams::color)eval_apply("__fml:c");`,
		`b = (bool)eval_apply("__fml:b");`,
		`i = (unsigned)eval_apply("__fml:i");`,
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
	// After TODO 017 milestone 4 the action_gen randomizes inputs via the
	// solver and reads them back via __from_solver. The free
	// ivy2cpp_random_node / ivy2cpp_random_a helpers remain emitted for
	// init_gen's assign_array_from_model branch; we still assert they
	// exist.
	for _, want := range []string{
		"static int ivy2cpp_random_node(gen &g)",
		"static gennumeric::a ivy2cpp_random_a(gen &g)",
		"return static_cast<int>(g.random_index(0, 4));",
		"return static_cast<gennumeric::a>(g.random_index(0, 4));",
		`randomize("__fml:n", "node");`,
		`randomize("__fml:av", "a");`,
		`n = (int)eval_apply("__fml:n");`,
		`av = (gennumeric::a)eval_apply("__fml:av");`,
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
	// Per-sort enum specializations are still emitted into the impl.
	for _, want := range []string{
		"template <>",
		"void __from_solver<solverconv::color>(gen &g, const z3::expr &v, solverconv::color &res)",
		"__from_solver<int>(g, v, temp);",
		"res = (solverconv::color)temp;",
		"z3::expr __to_solver<solverconv::color>(gen &g, const z3::expr &v, const solverconv::color &val)",
		"int thing = val;",
		"return __to_solver<int>(g, v, thing);",
		"void __randomize<solverconv::color>(gen &g, const z3::expr &v, const std::string &sort_name)",
		"__randomize<int>(g, v, sort_name);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in gen output:\n%s", want, out.Impl)
		}
	}
	// The primary __from_solver / __to_solver / __randomize templates
	// now live in ivy_go_z3.hpp so the impl's forward declarations of
	// per-sort specializations see them before use.
	support := readSupportHeader(t, "ivy_go_z3.hpp")
	for _, want := range []string{
		"template <typename T> void __from_solver(gen &, const z3::expr &, T &out)",
		"template <typename T> z3::expr __to_solver(gen &g, const char *sort_name, const T &value)",
		"template <typename T> z3::expr __to_solver(gen &g, const z3::expr &expr, const T &value)",
		"template <typename T> void __randomize(gen &g, const z3::expr &expr, const std::string &range)",
	} {
		if !strings.Contains(support, want) {
			t.Fatalf("missing primary template %q in ivy_go_z3.hpp:\n%s", want, support)
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
		// New weighted random loop pattern, with stack-allocated
		// generators dispatched by index switch (ivy_go_z3.hpp's `gen`
		// lacks virtual generate/execute).
		"set_gen set_generator(ivy);",
		"for (int cycle = 0; cycle < test_iters; cycle++)",
		"case 0: sat = set_generator.generate(ivy); break;",
		"case 0: set_generator.execute(ivy); break;",
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
	// After TODO 017 milestones 2 and 4, init_gen and action_gen both use
	// solve() (with alits as assumption literals) rather than check(),
	// and action_gen wraps generate() in push()/pop().
	for _, want := range []string{
		`#include "ivy_go_z3.hpp"`,
		"bool init_gen::generate(checkgen &obj)",
		"bool set_gen::generate(checkgen &obj)",
		"bool __res = solve();",
		"push();",
		"pop();",
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

// TestInitGenBodyMatchesPythonStructure asserts the init_gen::generate
// body emits the Python emit_init_gen control flow (ivy_to_cpp.py:935-976):
//   - cpptype_prepare(*this) at entry
//   - alits.clear() before randomization
//   - solve() (not check()) to compute the model
//   - eval+__init guarded by the solve() success
//   - cpptype_cleanup(*this) at exit
//
// Added as part of TODO 017 milestone 2.
func TestInitGenBodyMatchesPythonStructure(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "initstruct"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	idx := strings.Index(out.Impl, "bool init_gen::generate")
	if idx < 0 {
		t.Fatalf("init_gen::generate not emitted:\n%s", out.Impl)
	}
	end := strings.Index(out.Impl[idx:], "void init_gen::execute")
	if end < 0 {
		end = len(out.Impl) - idx
	}
	body := out.Impl[idx : idx+end]
	for _, want := range []string{
		"cpptype_prepare(*this);",
		"alits.clear();",
		"bool __res = solve();",
		"if (__res) {",
		"cpptype_cleanup(*this);",
		"obj.___ivy_gen = this;",
		"return __res;",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("init_gen body missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "if (!check())") {
		t.Fatalf("init_gen body should use solve(), not check():\n%s", body)
	}
}

// TestInitGenSkipsRandomizeForPrimitiveUnusedRanges verifies that the
// per-symbol dispatch inside init_gen follows Python emit_init_gen's
// branching at ivy_to_cpp.py:940-955. State symbols that appear in no
// constraint and have a primitive range (bool here, treated as "not
// is_primitive_sort" in Python's stub semantics) should be assigned
// directly via mk_rand rather than via the solver's randomize().
//
// This fixture has `marked` (relation) and `saved` (individual); neither
// is referenced by axioms or initial conditions, so Python emits
// assign_array_from_model with mk_rand. The Go port should match.
func TestInitGenSkipsRandomizeForUnusedSymbols(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
relation marked(C:color)
individual saved : color
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "useddisp"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// The init_gen body should contain direct C++ random assignments to
	// state symbols (assign_array_from_model branch), not g.randomize().
	idx := strings.Index(out.Impl, "bool init_gen::generate")
	if idx < 0 {
		t.Fatalf("init_gen::generate not emitted:\n%s", out.Impl)
	}
	end := strings.Index(out.Impl[idx:], "void init_gen::execute")
	if end < 0 {
		end = len(out.Impl) - idx
	}
	body := out.Impl[idx : idx+end]
	if strings.Contains(body, `randomize("marked"`) {
		t.Fatalf("init_gen body should not call randomize() for unused state %q:\n%s", "marked", body)
	}
	if !strings.Contains(body, "obj.marked[__ivy_arg0] =") {
		t.Fatalf("init_gen body should assign marked via mk_rand:\n%s", body)
	}
	if !strings.Contains(body, "obj.saved = ivy2cpp_random_color((*this));") {
		t.Fatalf("init_gen body should assign saved via ivy2cpp_random_color:\n%s", body)
	}
}

// TestActionGenEmitsReverseImagePrecondition asserts that the per-action
// generator emits its precondition as an SMT-LIB `add("(assert ...)")`
// string (Python ivy_to_cpp.py:1277, fed through the new
// `add(const std::string&)` runtime overload).
//
// Added as part of TODO 017 milestone 4.
func TestActionGenEmitsReverseImagePrecondition(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "ripre"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	idx := strings.Index(out.Impl, "set_gen::set_gen")
	if idx < 0 {
		t.Fatalf("set_gen constructor not emitted:\n%s", out.Impl)
	}
	end := strings.Index(out.Impl[idx:], "bool set_gen::generate")
	if end < 0 {
		end = len(out.Impl) - idx
	}
	body := out.Impl[idx : idx+end]
	for _, want := range []string{
		`mk_decl("__fml:c", {}, "color");`,
		`add(std::string("(assert `,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("set_gen constructor missing %q:\n%s", want, body)
		}
	}
}

// TestActionGenEmitsEmitEvalAfterSolve asserts that the generate() body
// of an action_gen reads each input back from the solver model after
// solve() succeeds (Python ivy_to_cpp.py:1307-1310).
//
// Added as part of TODO 017 milestone 4.
func TestActionGenEmitsEmitEvalAfterSolve(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "rieval"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	idx := strings.Index(out.Impl, "bool set_gen::generate")
	if idx < 0 {
		t.Fatalf("set_gen::generate not emitted:\n%s", out.Impl)
	}
	end := strings.Index(out.Impl[idx:], "void set_gen::execute")
	if end < 0 {
		end = len(out.Impl) - idx
	}
	body := out.Impl[idx : idx+end]
	for _, want := range []string{
		"push();",
		"alits.clear();",
		"bool __res = solve();",
		"if (__res) {",
		`c = (rieval::color)eval_apply("__fml:c");`,
		"pop();",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("set_gen::generate missing %q:\n%s", want, body)
		}
	}
}

// TestActionGenExtractDefinedParameters asserts that an `assume x = expr`
// in an action body causes extractDefinedParameters to remove the
// equation from the asserted precondition, and that emitDefinedInputs
// then computes the input value directly in the post-solve branch.
// Python ivy_to_cpp.py:1027-1046 + :1316.
//
// Added as part of TODO 017 milestone 5.
func TestActionGenExtractDefinedParameters(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..7}
individual stored : idx
action set(x:idx) = {
    assume x = 3;
    stored := x
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "defparam"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	idx := strings.Index(out.Impl, "bool set_gen::generate")
	if idx < 0 {
		t.Fatalf("set_gen::generate not emitted:\n%s", out.Impl)
	}
	end := strings.Index(out.Impl[idx:], "void set_gen::execute")
	if end < 0 {
		end = len(out.Impl) - idx
	}
	body := out.Impl[idx : idx+end]
	if !strings.Contains(body, "this->x = 3;") {
		t.Fatalf("expected emit_defined_inputs to emit `this->x = 3;` in set_gen body:\n%s", body)
	}
	// The defined input MUST NOT also be read back from the solver model.
	if strings.Contains(body, `x = (`) && strings.Contains(body, `)eval_apply("__fml:x");`) {
		t.Fatalf("defined input x should not be read from solver model:\n%s", body)
	}
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
		"ivy.edge[__ivy_arg0][__ivy_arg1] = g.random_bool();",
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

func TestBitvectorTypesAndExpressionsShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type byte
type nibble
interpret byte -> bv[8]
interpret nibble -> bv[4]
individual x : byte
individual y : byte
individual n : nibble
function concat(X:nibble,Y:nibble) : byte
action step = {
    x := 300;
    x := bvand(x,y);
    x := bvor(x,y);
    x := bvnot(x);
    x := cast(n);
    x := concat(n,n);
    n := bfe[0][3](x)
}
export step
`)
	out, err := Generate(mod, Config{ClassName: "bvshape"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Header + out.Impl
	for _, want := range []string{
		"typedef unsigned byte;",
		"typedef unsigned nibble;",
		"unsigned x;",
		"unsigned y;",
		"unsigned n;",
		"x = (300 & 255);",
		"x = (((x & y)) & 255);",
		"x = (((x | y)) & 255);",
		"x = (((~x)) & 255);",
		"x = ((static_cast<unsigned>(n)) & 255);",
		"x = ((((static_cast<unsigned>(n)) << 4 | (static_cast<unsigned>(n)))) & 255);",
		"n = ((static_cast<unsigned>((x >> 0))) & 15);",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in bitvector output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

func TestStringBitvectorHelperReplAndZ3Shape(t *testing.T) {
	src := `#lang ivy1.7
type text
interpret text -> strbv[4]
individual saved : text
action set(t:text) returns(out:text) = {
    saved := t;
    out := t
}
export set
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "strshape"})
	if err != nil {
		t.Fatalf("Generate repl: %v", err)
	}
	text := out.Header + out.Impl
	for _, want := range []string{
		"class text : public std::string {",
		"text(const std::string &s) : std::string(s) {}",
		"text(const char *s) : std::string(s) {}",
		"size_t __hash() const { return hash_space::hash<std::string>()(*this); }",
		"static hash_space::hash_map<std::string,int> x_to_bv_hash;",
		"static hash_space::hash_map<int,std::string> bv_to_x_hash;",
		"std::ostream &operator<<(std::ostream &s, const strshape::text &t)",
		// cpp_types.go emits the _arg<T> specialization; the dispatcher
		// calls it with the carc bound (strbv[4] has 2^4 = 16 elements).
		`template <> strshape::text _arg<strshape::text>(std::vector<ivy_value> &args, unsigned idx, long long bound)`,
		"return args[idx].atom;",
		`strshape::text __ivy_result = ivy.set(_arg<strshape::text>(args, 0, 16));`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in strbv repl output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)

	mod = compileIvySource(t, src)
	genOut, err := Generate(mod, Config{Target: "gen", ClassName: "strgen"})
	if err != nil {
		t.Fatalf("Generate gen: %v", err)
	}
	genText := genOut.Header + genOut.Impl
	for _, want := range []string{
		`g.mk_bv("text", 4);`,
		"template <> void __from_solver<strgen::text>(gen &g, const z3::expr &v, strgen::text &res)",
		"template <> z3::expr __to_solver<strgen::text>(gen &g, const z3::expr &v, const strgen::text &val)",
		"template <> void __randomize<strgen::text>(gen &g, const z3::expr &apply_expr, const std::string &sort_name)",
		"static strgen::text ivy2cpp_random_text(gen &g)",
		"return strgen::text::bv_to_x(g.random_index(0, 15));",
	} {
		if !strings.Contains(genText, want) {
			t.Fatalf("missing %q in strbv gen output:\nheader:\n%s\nimpl:\n%s", want, genOut.Header, genOut.Impl)
		}
	}
	assertNoUnsupportedCPP(t, genOut)
	compileGeneratedCPP(t, genOut)
}

func TestIntBitvectorHelperReplAndZ3Shape(t *testing.T) {
	src := `#lang ivy1.7
type small
interpret small -> intbv[10][20][4]
individual saved : small
action set(x:small) returns(out:small) = {
    saved := x;
    out := x
}
export set
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "intshape"})
	if err != nil {
		t.Fatalf("Generate repl: %v", err)
	}
	text := out.Header + out.Impl
	for _, want := range []string{
		"struct IntClass {",
		"class small : public IntClass {",
		"small(long long v) : IntClass(v) {}",
		"long long val;",
		"size_t __hash() const { return hash_space::hash<IntClass>()(*this); }",
		// intbv _arg<> from cpp_types.go (else branch at line 344).
		// intbv[10][20][4]: bound is hi-lo+1 = 11.
		`template <> intshape::small _arg<intshape::small>(std::vector<ivy_value> &args, unsigned idx, long long bound)`,
		"s >> res.val;",
		"std::ostream &operator<<(std::ostream &s, const intshape::small &t)",
		"s << t.val;",
		`intshape::small __ivy_result = ivy.set(_arg<intshape::small>(args, 0, 11));`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in intbv repl output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)

	mod = compileIvySource(t, src)
	genOut, err := Generate(mod, Config{Target: "gen", ClassName: "intgen"})
	if err != nil {
		t.Fatalf("Generate gen: %v", err)
	}
	genText := genOut.Header + genOut.Impl
	for _, want := range []string{
		`g.mk_bv("small", 4);`,
		"template <> void __from_solver<intgen::small>(gen &g, const z3::expr &v, intgen::small &res)",
		"template <> z3::expr __to_solver<intgen::small>(gen &g, const z3::expr &v, const intgen::small &val)",
		"template <> void __randomize<intgen::small>(gen &g, const z3::expr &apply_expr, const std::string &sort_name)",
		"static intgen::small ivy2cpp_random_small(gen &g)",
		"return intgen::small(g.random_index(10, 20));",
	} {
		if !strings.Contains(genText, want) {
			t.Fatalf("missing %q in intbv gen output:\nheader:\n%s\nimpl:\n%s", want, genOut.Header, genOut.Impl)
		}
	}
	assertNoUnsupportedCPP(t, genOut)
	compileGeneratedCPP(t, genOut)
}

func TestBitvectorBackedSortsInGenSetupAndStorageShape(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type word
type text
type small
interpret word -> bv[8]
interpret text -> strbv[4]
interpret small -> intbv[10][20][4]
individual w : word
individual t : text
individual i : small
action set(x:word,y:text,z:small) = {
    w := x;
    t := y;
    i := z
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "genbits"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Header + out.Impl
	// After TODO 028 the action_gen randomizes cpptype-range inputs via
	// the per-type __randomize<class::T> template (matching Python's
	// emit_randomize at ivy_to_cpp.py:996); init_gen still calls the
	// ivy2cpp_random_* helpers for state symbols not referenced by any
	// constraint.
	for _, want := range []string{
		`g.mk_bv("word", 8);`,
		`g.mk_bv("text", 4);`,
		`g.mk_bv("small", 4);`,
		"obj.w = ivy2cpp_random_word((*this));",
		"obj.t = ivy2cpp_random_text((*this));",
		"obj.i = ivy2cpp_random_small((*this));",
		"genbits::text y;",
		"genbits::small z;",
		`__randomize<genbits::word>(*this, apply("__fml:x"), "word");`,
		`__randomize<genbits::text>(*this, apply("__fml:y"), "text");`,
		`__randomize<genbits::small>(*this, apply("__fml:z"), "small");`,
		`__from_solver<genbits::word>(*this, apply("__fml:x"), x);`,
		`__from_solver<genbits::text>(*this, apply("__fml:y"), y);`,
		`__from_solver<genbits::small>(*this, apply("__fml:z"), z);`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in gen bitvector-backed output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)

	storageMod := compileIvySource(t, `#lang ivy1.7
type word
type text
interpret word -> bv[8]
interpret text -> strbv[4]
function owner(W:word) : text
relation seen(T:text)
`)
	storageOut, err := Generate(storageMod, Config{ClassName: "bitstore"})
	if err != nil {
		t.Fatalf("Generate storage: %v", err)
	}
	for _, want := range []string{
		"typedef unsigned word;",
		"class text : public std::string {",
		"text owner[256];",
		"hash_thunk<text,bool> seen;",
	} {
		if !strings.Contains(storageOut.Header, want) {
			t.Fatalf("missing %q in storage header:\n%s", want, storageOut.Header)
		}
	}
	assertNoUnsupportedCPP(t, storageOut)
	compileGeneratedCPP(t, storageOut)
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

// --- TODO 007: full destructor / struct support ---

const destructorMultiArgIvySource = `#lang ivy1.7
type color = {red, green, blue}
type idx = {0..3}
type cell
destructor shade(C:cell, I:idx) : color
`

func TestDestructorMultiArgFieldDeclaration(t *testing.T) {
	mod := compileIvySource(t, destructorMultiArgIvySource)
	out, err := Generate(mod, Config{ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"struct cell {",
		"color shade[4];",
		"size_t __hash() const {",
		"size_t hv = 0;",
		// cppHashType collapses enums to int for hashing; shade is an enum.
		"hv += hash_space::hash<int>()(shade[X__0]);",
		"bool operator==(const cell &other) const {",
		"if (!(shade[X__0] == other.shade[X__0])) return false;",
		"bool operator<(const cell &other) const {",
	} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q in header:\n%s", want, out.Header)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestDestructorHashThunkField(t *testing.T) {
	// `key` is uninterpreted with no cardinality, so the destructor field
	// must lower to hash_thunk storage. The struct __hash skips it per
	// Python is_large_destr; equality and stream fall back to scalar form.
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type key
type cell
destructor shade(C:cell, K:key) : color
`)
	out, err := Generate(mod, Config{ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"struct cell {",
		"hash_thunk<",
		",color> shade;",
		"size_t __hash() const {",
	} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q in header:\n%s", want, out.Header)
		}
	}
	if strings.Contains(out.Header, "hv += hash_space::hash<color>()(shade") {
		t.Fatalf("hash should skip hash_thunk-storage destructor field:\n%s", out.Header)
	}
	if !strings.Contains(out.Header, "if (!(shade == other.shade)) return false;") {
		t.Fatalf("equality should fall back to scalar compare for hash_thunk field:\n%s", out.Header)
	}
	compileGeneratedCPP(t, out)
}

func TestDestructorStructStreamMultiArg(t *testing.T) {
	mod := compileIvySource(t, destructorMultiArgIvySource)
	out, err := Generate(mod, Config{ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		`out << "shade:";`,
		`out << "[";`,
		"for (int X__0 = 0; X__0 < 4; X__0++) {",
		`if (X__0) out << ",";`,
		"out << value.shade[X__0];",
		`out << "]";`,
	} {
		if !strings.Contains(out.Header, want) {
			t.Fatalf("missing %q in header writer:\n%s", want, out.Header)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestDestructorMultiArgFieldRoundTrip(t *testing.T) {
	// Construct multi-arg field actions manually (matching the pattern in
	// TestGeneratedFieldActionsCompile). For multi-arg destructor LHS the
	// emitter routes through Assign(App(shade, c, i), v) and must use
	// cppDestructorFieldAccess for indexing.
	mod := compileIvySource(t, destructorMultiArgIvySource+`
individual a : cell
action step = {}
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
	a := goivy.NewConst("a", cell)
	zero := goivy.NewConst("0", idx)
	red := goivy.NewConst("red", color)
	// Build: a.shade(0) := red
	lhs, err := goivy.NewApply(shade, a, zero)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	assign := goivy.NewAssignAction(lhs, red)
	mod.Actions.Set("step", goivy.NewSequence(assign))
	mod.PublicActions.Set("step", true)
	out, err := Generate(mod, Config{ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// The index expression is bounds-clamped by emitExpr for range targets,
	// so we look for "a.shade[" and "] = red;" with red as the RHS rather
	// than a literal "a.shade[0] = red;".
	if !strings.Contains(out.Impl, "a.shade[") || !strings.Contains(out.Impl, "] = red;") {
		t.Fatalf("missing array-indexed destructor write in impl:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

func TestDestructorSerDeserShape(t *testing.T) {
	mod := compileIvySource(t, destructorMultiArgIvySource)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"template <> void __ser<heap::cell>(ivy_ser &res, const heap::cell &t) {",
		"res.open_struct();",
		`res.open_field("shade");`,
		"__ser<heap::color>(res, t.shade[X__0]);",
		"res.close_field();",
		"res.close_struct();",
		"template <> void __deser<heap::cell>(ivy_deser &inp, heap::cell &res) {",
		"inp.open_struct();",
		`inp.open_field("shade");`,
		"inp.open_list();",
		"__deser(inp, res.shade[X__0]);",
		"inp.close_list();",
		"inp.close_field();",
		"inp.close_struct();",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestDestructorArgShape(t *testing.T) {
	mod := compileIvySource(t, destructorMultiArgIvySource)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"template <> heap::cell _arg<heap::cell>(std::vector<ivy_value> &args, unsigned idx, long long bound) {",
		"ivy_value &arg = args[idx];",
		"std::vector<ivy_value> tmp_args(1);",
		"for (unsigned i = 0; i < arg.fields.size(); i++) {",
		"if (arg.fields[i].is_member()) {",
		`if (arg.fields[i].atom == "shade") {`,
		"if (tmp.atom.size() || tmp.fields.size() != 4) throw out_of_bounds(idx, tmp.pos);",
		"for (int X__0 = 0; X__0 < 4; X__0++) {",
		"res.shade[X__0] = _arg<heap::color>(tmp_args, 0, 3);",
		`throw out_of_bounds("in field shade: " + err.txt, err.pos);`,
		`throw out_of_bounds("unexpected field: " + arg.fields[i].atom, arg.fields[i].pos);`,
		`throw out_of_bounds("expected struct", args[idx].pos);`,
		"return res;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestDestructorZ3ImplShape(t *testing.T) {
	mod := compileIvySource(t, destructorMultiArgIvySource+`
individual a : cell
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"#ifdef Z3PP_H_",
		"template <> void __from_solver<heap::cell>(gen &g, const z3::expr &v, heap::cell &res) {",
		`g.apply("shade", v, g.int_to_z3(g.sort("idx"), X__0))`,
		"res.shade[X__0]",
		"template <> z3::expr __to_solver<heap::cell>(gen &g, const z3::expr &v, const heap::cell &val) {",
		"std::string fname = g.fresh_name();",
		`z3::expr tmp = g.ctx.constant(fname.c_str(), g.sort("cell"));`,
		`g.slvr.add(__to_solver(g, g.apply("shade", tmp, g.int_to_z3(g.sort("idx"), X__0)), val.shade[X__0]));`,
		"return v == tmp;",
		"template <> void __randomize<heap::cell>(gen &g, const z3::expr &v, const std::string &sort_name) {",
		`__randomize<heap::color>(g, g.apply("shade", v, g.int_to_z3(g.sort("idx"), X__0)), "color");`,
		`g.mk_decl("shade", {"cell", "idx"}, "color");`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

func TestDestructorRandomizeSkipsUninterpretedRange(t *testing.T) {
	// `pos` is an uninterpreted sort with no cppInterp, no native type, and
	// not a destructor sort — so it satisfies isReallyUninterpretedRange and
	// must be skipped by the destructor __randomize body.
	mod := compileIvySource(t, `#lang ivy1.7
type pos
type cell
destructor here(C:cell) : pos
individual a : cell
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// The __randomize template must still be emitted (empty-bodied) — but
	// must not contain a __randomize<...>(...) call for the `here` field.
	idx := strings.Index(out.Impl, "template <> void __randomize<heap::cell>")
	if idx < 0 {
		t.Fatalf("missing __randomize<heap::cell> template:\n%s", out.Impl)
	}
	tail := out.Impl[idx:]
	endIdx := strings.Index(tail, "#endif")
	if endIdx < 0 {
		t.Fatalf("missing #endif after __randomize template:\n%s", tail)
	}
	body := tail[:endIdx]
	if strings.Contains(body, `__randomize<`) && strings.Contains(body, `g.apply("here"`) {
		t.Fatalf("__randomize should skip uninterpreted-range destructor field:\n%s", body)
	}
	compileGeneratedCPP(t, out)
}

// TestDestructorForwardDeclsInImplPreamble verifies destructor sort
// forward declarations are emitted right after the include block. Mirrors
// Python ivy_to_cpp.py:2232-2243.
func TestDestructorForwardDeclsInImplPreamble(t *testing.T) {
	mod := compileIvySource(t, destructorMultiArgIvySource)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"std::ostream &operator<<(std::ostream &s, const heap::cell &t);",
		"heap::cell _arg<heap::cell>(std::vector<ivy_value> &args, unsigned idx, long long bound);",
		"void __ser<heap::cell>(ivy_ser &res, const heap::cell &);",
		"void __deser<heap::cell>(ivy_deser &inp, heap::cell &res);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	declIdx := strings.Index(out.Impl, "void __ser<heap::cell>(ivy_ser &res, const heap::cell &);")
	defIdx := strings.Index(out.Impl, "template <> void __ser<heap::cell>(ivy_ser &res, const heap::cell &t) {")
	if declIdx < 0 || defIdx < 0 || declIdx >= defIdx {
		t.Fatalf("destructor forward decl must precede its definition; declIdx=%d defIdx=%d", declIdx, defIdx)
	}
	compileGeneratedCPP(t, out)
}

// TestDestructorZ3ForwardDeclsInImplPreamble verifies destructor Z3
// template forward decls are emitted (under #ifdef Z3PP_H_) for test
// targets. Mirrors Python ivy_to_cpp.py:2246-2254.
func TestDestructorZ3ForwardDeclsInImplPreamble(t *testing.T) {
	mod := compileIvySource(t, destructorMultiArgIvySource+`
individual a : cell
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"void __from_solver<heap::cell>(gen &g, const z3::expr &v, heap::cell &res);",
		"z3::expr __to_solver<heap::cell>(gen &g, const z3::expr &v, const heap::cell &val);",
		"void __randomize<heap::cell>(gen &g, const z3::expr &v, const std::string &sort_name);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

// TestDestructorArgZeroInitsPrimitiveFields verifies that a destructor
// _arg<T> body zero-initializes every primitive field before parsing,
// so partial input atoms leave omitted fields well-defined. Mirrors
// Python ivy_to_cpp.py:2529 (`assign_zero_symbol`).
func TestDestructorArgZeroInitsPrimitiveFields(t *testing.T) {
	mod := compileIvySource(t, destructorMultiArgIvySource)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	declIdx := strings.Index(out.Impl, "heap::cell res;")
	if declIdx < 0 {
		t.Fatalf("missing `heap::cell res;` in impl:\n%s", out.Impl)
	}
	tail := out.Impl[declIdx:]
	parseStart := strings.Index(tail, "ivy_value &arg = args[idx];")
	if parseStart < 0 {
		t.Fatalf("missing parse-loop marker in impl:\n%s", tail)
	}
	preamble := tail[:parseStart]
	for _, want := range []string{
		"for (int X__0 = 0; X__0 < 4; X__0++) {",
		"res.shade[X__0] = (heap::color)0;",
	} {
		if !strings.Contains(preamble, want) {
			t.Fatalf("missing %q in _arg zero-init preamble:\n%s", want, preamble)
		}
	}
	compileGeneratedCPP(t, out)
}

// TestDestructorZeroInitSkipsStringField verifies that strlit-typed
// destructor fields are NOT assigned `(__strlit)0` (which would not
// compile). Mirrors Python's `ctype(v.sort) == '__strlit'` guard in
// assign_zero_symbol (ivy_to_cpp.py:216-219).
func TestDestructorZeroInitSkipsStringField(t *testing.T) {
	src := `#lang ivy1.7
type cell
type str
interpret str -> strlit
destructor label(C:cell) : str
individual a : cell
action step = {}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "heap"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	declIdx := strings.Index(out.Impl, "heap::cell res;")
	if declIdx < 0 {
		t.Fatalf("missing `heap::cell res;` in impl:\n%s", out.Impl)
	}
	tail := out.Impl[declIdx:]
	parseStart := strings.Index(tail, "ivy_value &arg = args[idx];")
	preamble := tail[:parseStart]
	if strings.Contains(preamble, "(__strlit)0") || strings.Contains(preamble, "(std::string)0") {
		t.Fatalf("_arg zero-init must skip strlit destructor fields; preamble:\n%s", preamble)
	}
	compileGeneratedCPP(t, out)
}

// TestHashThunkToSolverSpecializationEmittedForTestTarget verifies a
// single-arg hash_thunk domain produces a `to_solver_class<hash_thunk<D,R>>`
// specialization in test mode. Mirrors Python ivy_to_cpp.py:2673 →
// emit_all_ctuples_to_solver.
func TestHashThunkToSolverSpecializationEmittedForTestTarget(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type key
relation seen(K:key)
after init { seen(K) := false }
action mark(k:key) = { seen(k) := true }
export mark
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Unbounded uninterpreted `key` lowers to `int` per cppScalarTypeWith,
	// matching Python's ctype dispatcher (ivy_to_cpp.py:414-432).
	for _, want := range []string{
		"template<typename R> class to_solver_class<hash_thunk<int,R> > {",
		"z3::expr cond = __to_solver(g, v.arg(0), it->first);",
		"res = res && (disj || bg);",
		"dynamic_cast<z3_thunk<int,R> *>(val.fun)->to_z3(g, v)",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

// TestHashThunkToSolverNotEmittedForReplTarget verifies the
// specialization is suppressed for repl targets — emit is gated on
// usesZ3 per generator.go.
func TestHashThunkToSolverNotEmittedForReplTarget(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type key
relation seen(K:key)
after init { seen(K) := false }
action mark(k:key) = { seen(k) := true }
export mark
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "runner"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if strings.Contains(out.Impl, "to_solver_class<hash_thunk<") {
		t.Fatalf("repl target must not emit to_solver_class<hash_thunk<...>>:\n%s", out.Impl)
	}
	compileGeneratedCPP(t, out)
}

// TestEmitSetSolverDestructorRecordRange asserts that emit_set on a state
// symbol with a destructor record range emits the per-field
// add(__to_solver(*this, apply("<destr>", apply("<sym>"), <z3-idx>...),
// <obj>.<sym>[<idx>...].<field>)) constraint, mirroring Python
// emit_set_field at ivy_to_cpp.py:806-824. The Python emission appends the
// loop indices BEFORE the field name in the rhs lvalue (a literal port of
// the Python recursion).
//
// White-box: invokes emitSetSolver directly with a constructed Generator
// so the test exercises the destructor branch without depending on the
// action_gen reverse-image path (which currently strips destructor-record
// receivers via expandFieldReferences).
//
// Added as part of TODO 018 milestone 1.
func TestEmitSetSolverDestructorRecordRange(t *testing.T) {
	mod := compileIvySource(t, destructorMultiArgIvySource+`
individual saved : cell
`)
	g := &Generator{Mod: mod, ClassName: "heap"}
	g.Config = Config{Target: "test", ClassName: "heap"}
	// Find the `saved` symbol from stateSymbols so emitSetSolver gets the
	// real Sort wiring.
	var saved stateSymbol
	for _, s := range g.stateSymbols() {
		if s.Name == "saved" {
			saved = s
			break
		}
	}
	if saved.Name == "" {
		t.Fatalf("saved not found in stateSymbols")
	}
	w := &cppWriter{}
	g.emitSetSolver(w, saved, "obj")
	body := w.String()
	// New behavior: a per-destructor loop opens X__0 over idx (4 elements)
	// and emits the recursive add(__to_solver(...)) call. With the literal
	// Python port, the rhs is `obj.saved[X__0].shade` (indices precede the
	// dot-field per Python emit_set_field line 816).
	for _, want := range []string{
		"X__0 = 0; X__0 <= 3; X__0++",
		`add(__to_solver(*this, apply("shade", apply("saved"), int_to_z3(sort("idx"), static_cast<long long>(X__0))), obj.saved[X__0].shade));`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in emitSetSolver output:\n%s", want, body)
		}
	}
	// The old stub marker must be gone.
	if strings.Contains(body, "emitSetField unsupported") {
		t.Fatalf("emit_set_field stub marker still present:\n%s", body)
	}
}

// TestEmitEvalBranchesByRangeKind asserts that emit_eval emits the
// Python-faithful branched form (ivy_to_cpp.py:789-796):
//
//   - destructor / native / cpp-interp range → `__from_solver<class::T>
//     (*this, apply("name", ...), lvalue);`
//   - primitive range → `lvalue = (ctype)eval_apply("name", ...);`
//
// Added as part of TODO 018 milestone 4.b.
func TestEmitEvalBranchesByRangeKind(t *testing.T) {
	// Primitive range (enum): direct cast via eval_apply.
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = { saved := c }
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "evalbr"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, `c = (evalbr::color)eval_apply("__fml:c");`) {
		t.Fatalf("expected primitive-range eval_apply cast in gen:\n%s", out.Impl)
	}
	if strings.Contains(out.Impl, `__from_solver(*this, mk_apply_expr("__fml:c", {})`) {
		t.Fatalf("legacy unified __from_solver form should be gone:\n%s", out.Impl)
	}

	// Cpp-interp range (bv): __from_solver<class::T> via apply, NOT
	// __from_solver<unsigned> (which would skip the type's specialization).
	mod2 := compileIvySource(t, `#lang ivy1.7
type word
interpret word -> bv[8]
individual w : word
action set(x:word) = { w := x }
export set
`)
	out2, err := Generate(mod2, Config{Target: "gen", ClassName: "evalbr2"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out2.Impl, `__from_solver<evalbr2::word>(*this, apply("__fml:x"), x);`) {
		t.Fatalf("expected cpp-typed range to route to class-qualified __from_solver:\n%s", out2.Impl)
	}
}

// TestEnumSortSolverSpecsAreTemplateSpecializations asserts that the
// per-enum __from_solver / __to_solver / __randomize emissions for
// gen/test targets are template specializations (Python form,
// ivy_to_cpp.py:2654-2668) rather than static overloads, with matching
// forward declarations in the impl preamble (Python lines 2224-2230).
//
// Added as part of TODO 018 milestone 3.
func TestEnumSortSolverSpecsAreTemplateSpecializations(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green, blue}
individual saved : color
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "enumspec"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Forward decls — gated under #ifdef Z3PP_H_ for gen/test targets.
	for _, want := range []string{
		"#ifdef Z3PP_H_",
		"template <>",
		"void __from_solver<enumspec::color>(gen &g, const z3::expr &v, enumspec::color &res);",
		"z3::expr __to_solver<enumspec::color>(gen &g, const z3::expr &v, const enumspec::color &val);",
		"void __randomize<enumspec::color>(gen &g, const z3::expr &v, const std::string &sort_name);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	// Bodies — must be template specializations, not static overloads.
	for _, want := range []string{
		"void __from_solver<enumspec::color>(gen &g, const z3::expr &v, enumspec::color &res) {",
		"__from_solver<int>(g, v, temp);",
		"res = (enumspec::color)temp;",
		"z3::expr __to_solver<enumspec::color>(gen &g, const z3::expr &v, const enumspec::color &val) {",
		"int thing = val;",
		"return __to_solver<int>(g, v, thing);",
		"void __randomize<enumspec::color>(gen &g, const z3::expr &v, const std::string &sort_name) {",
		"__randomize<int>(g, v, sort_name);",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	// The old static-overload form must NOT be present.
	for _, bad := range []string{
		"static void __from_solver(gen &g, const z3::expr &expr, enumspec::color",
		"static z3::expr __to_solver(gen &g, const char *sort_name, enumspec::color",
		"static z3::expr __to_solver(gen &g, const z3::expr &expr, enumspec::color",
	} {
		if strings.Contains(out.Impl, bad) {
			t.Fatalf("unexpected legacy static-overload form %q in impl", bad)
		}
	}
}

// TestEmitSetSolverLargeTypeForall asserts that emit_set on a state
// symbol whose function-sort domain is "large" (Python is_large_type at
// ivy_to_cpp.py:445-449) emits a forall-quantified __to_solver
// constraint (Python lines 844-852) instead of unrolling the domain.
//
// Added as part of TODO 018 milestone 2.
func TestEmitSetSolverLargeTypeForall(t *testing.T) {
	// `key` is uninterpreted with no cardinality, so the domain of
	// `valueof : (key -> bool)` is non-integer/unsized and triggers the
	// large-type branch.
	mod := compileIvySource(t, `#lang ivy1.7
type key
relation valueof(K:key)
`)
	g := &Generator{Mod: mod, ClassName: "kv"}
	g.Config = Config{Target: "test", ClassName: "kv"}
	var sym stateSymbol
	for _, s := range g.stateSymbols() {
		if s.Name == "valueof" {
			sym = s
			break
		}
	}
	if sym.Name == "" {
		t.Fatalf("valueof not found in stateSymbols")
	}
	if !g.isLargeType(sym.Sort) {
		t.Fatalf("expected isLargeType=true for valueof : (key -> bool)")
	}
	w := &cppWriter{}
	g.emitSetSolver(w, sym, "obj")
	body := w.String()
	for _, want := range []string{
		"std::vector<z3::expr> __quants;",
		`__quants.push_back(ctx.constant("X__0", sort("key")));`,
		`add(forall(__quants, __to_solver(*this, apply("valueof", ctx.constant("X__0", sort("key"))), obj.valueof)));`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in emitSetSolver output:\n%s", want, body)
		}
	}
	// Must not fall through to the default per-element loop.
	if strings.Contains(body, `int_to_z3(sort("key"), static_cast<long long>`) {
		t.Fatalf("large-type branch should not unroll into int_to_z3 calls:\n%s", body)
	}
}

// --- TODO 009: method signatures, parameter passing, return handling ---

// TestPrivateActionStructParamUsesConstRef verifies that a private action
// taking a struct (destructor-sort) parameter it does not modify is emitted
// with a const-reference parameter, per Python annotate_action
// (ivy_to_cpp.py:1493-1495): not modified and is_struct → ConstRefType.
func TestPrivateActionStructParamUsesConstRef(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type cell
destructor shade(C:cell) : color
individual saved : color
action read_shade(c:cell) returns (out:color) = {
    out := shade(c)
}
action step = {
    var k : cell;
    call saved := read_shade(k)
}
export step
`)
	// Trim PublicActions to mirror create_isolate.
	mod.PublicActions.Set("read_shade", false)
	out, err := Generate(mod, Config{ClassName: "constref"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := "color read_shade(const cell& c)"
	if !strings.Contains(out.Header+out.Impl, want) {
		t.Fatalf("expected const-ref signature %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
	}
	compileGeneratedCPP(t, out)
}

// TestPrivateActionStructParamModifiedUsesValue verifies that a private
// action that modifies its struct parameter is emitted with a by-value
// parameter, per Python annotate_action's `action_assigns(p) → ValueType`.
func TestPrivateActionStructParamModifiedUsesValue(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type cell
destructor shade(C:cell) : color
action recolor(c:cell) = {
    c.shade := red
}
action step = {
    var k : cell;
    call recolor(k)
}
export step
`)
	mod.PublicActions.Set("recolor", false)
	out, err := Generate(mod, Config{ClassName: "byval"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := "void recolor(cell c)"
	if !strings.Contains(out.Header+out.Impl, want) {
		t.Fatalf("expected by-value signature %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
	}
	bad := "const cell&"
	if strings.Contains(out.Header+out.Impl, bad) {
		t.Fatalf("modified struct param must not be const-ref %q:\nheader:\n%s\nimpl:\n%s", bad, out.Header, out.Impl)
	}
	compileGeneratedCPP(t, out)
}

// TestReturnAliasingInputUsesRefType verifies that when a return name matches
// an input parameter name, the input is passed as a non-const reference and
// there is no synthetic local or `return` statement in the body — the
// parameter IS the return storage. Mirrors Python's RefType branch (input
// in returns) at ivy_to_cpp.py:1493 and the body-emission skip when
// rtypes[0] is ReturnRefType.
func TestReturnAliasingInputUsesRefType(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx
interpret idx -> <<< int >>>
action bump(x:idx) returns (x:idx) = {
    x := x + 1
}
action step = {
    var v : idx;
    call v := bump(v)
}
export step
`)
	mod.PublicActions.Set("bump", false)
	out, err := Generate(mod, Config{ClassName: "alias"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := "void bump(idx& x)"
	if !strings.Contains(out.Header+out.Impl, want) {
		t.Fatalf("expected ref signature %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
	}
	// Body must NOT contain `return x;` and must NOT pre-initialize x.
	bodyStart := strings.Index(out.Impl, "alias::bump(")
	if bodyStart < 0 {
		t.Fatalf("missing alias::bump body:\n%s", out.Impl)
	}
	body := out.Impl[bodyStart:]
	if end := strings.Index(body, "\n        }"); end >= 0 {
		body = body[:end]
	}
	if strings.Contains(body, "return x;") {
		t.Fatalf("ReturnRefType body must not emit return:\n%s", body)
	}
	if strings.Contains(body, "idx x = ") {
		t.Fatalf("ReturnRefType body must not synthesize x local:\n%s", body)
	}
	compileGeneratedCPP(t, out)
}

// TestCallAliasSafetySwapInputAndOutput verifies that when an output target
// at slot pos differs from the input expression at that slot, the input is
// saved to a temporary BEFORE the call and a post-call copy-back restores
// the output. Mirrors Python emit_call (ivy_to_cpp.py:3845-3861, 3877-3880).
func TestCallAliasSafetySwapInputAndOutput(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx
interpret idx -> <<< int >>>
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
	out, err := Generate(mod, Config{ClassName: "aliassafe"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// rtypes[0] = ReturnRefType{Pos:0}; iparg=b, rv=a; names differ →
	// pre-call save to a temp + post-call copy back.
	for _, want := range []string{"= loc__b;", "bump(__tmp", "loc__a = __tmp"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("alias-safety: expected snippet %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

// TestCallNoAliasNoTempEmitted verifies that an ordinary call with distinct,
// non-aliasing input and output expressions does not introduce an alias-
// safety temporary. The matched-input optimization in Python emit_call
// (ivy_to_cpp.py:3852: `iparg != rv`) keeps the call clean.
func TestCallNoAliasNoTempEmitted(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx
interpret idx -> <<< int >>>
action ident(a:idx) returns (b:idx) = {
    b := a
}
action step = {
    var x : idx;
    var y : idx;
    call y := ident(x)
}
export step
`)
	mod.PublicActions.Set("ident", false)
	out, err := Generate(mod, Config{ClassName: "noalias"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// rtypes[0] = ValueType (b not in formals, pos=0) → primary
	// assigned via `lhs = ident(args);`, no trailing ref, no temp.
	want := "loc__y = ident(loc__x);"
	if !strings.Contains(out.Impl, want) {
		t.Fatalf("expected clean assigning call %q:\n%s", want, out.Impl)
	}
	if strings.Contains(out.Impl, "__tmp") {
		// Some __tmpN identifiers exist for unrelated reasons in the
		// runtime skeleton. Restrict the check to the step body.
		bodyStart := strings.Index(out.Impl, "noalias::step(")
		if bodyStart >= 0 {
			body := out.Impl[bodyStart:]
			if end := strings.Index(body, "\n        }"); end >= 0 {
				body = body[:end]
			}
			if strings.Contains(body, "__tmp") {
				t.Fatalf("clean call should not emit alias temps in step body:\n%s", body)
			}
		}
	}
	compileGeneratedCPP(t, out)
}

// TestPublicActionAllValueTypeSignature verifies that an exported action
// gets the all-ValueType signature regardless of its parameter sorts, per
// Python annotate_action (ivy_to_cpp.py:1481-1484): public branch forces
// ValueType for every parameter and return.
func TestPublicActionAllValueTypeSignature(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
type cell
destructor shade(C:cell) : color
individual seen : color
action observe(c:cell) = {
    seen := shade(c)
}
export observe
`)
	// observe is exported and STAYS public; compileIvySource already
	// puts it there.
	out, err := Generate(mod, Config{ClassName: "publicval"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want := "void observe(cell c)"
	if !strings.Contains(out.Header+out.Impl, want) {
		t.Fatalf("public action must take cell by value, got:\nheader:\n%s\nimpl:\n%s", out.Header, out.Impl)
	}
	if strings.Contains(out.Header, "const cell&") {
		t.Fatalf("public action must NOT use const-ref for struct param:\nheader:\n%s", out.Header)
	}
	compileGeneratedCPP(t, out)
}

// TestVirtualKeywordOmittedForGenTarget verifies the `virtual ` prefix on
// header method declarations is present for non-gen targets and absent for
// target=gen, mirroring Python emit_method_decl (ivy_to_cpp.py:1559-1560).
func TestVirtualKeywordOmittedForGenTarget(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`
	implOut, err := Generate(compileIvySource(t, src), Config{Target: "impl", ClassName: "vk"})
	if err != nil {
		t.Fatalf("Generate impl: %v", err)
	}
	if !strings.Contains(implOut.Header, "virtual void set(") {
		t.Fatalf("impl target should emit 'virtual void set(':\n%s", implOut.Header)
	}
	genOut, err := Generate(compileIvySource(t, src), Config{Target: "gen", ClassName: "vk"})
	if err != nil {
		t.Fatalf("Generate gen: %v", err)
	}
	if strings.Contains(genOut.Header, "virtual void set(") {
		t.Fatalf("gen target should NOT emit 'virtual ' on set:\n%s", genOut.Header)
	}
}

// TestCallSiteVariantUpcastOnArgumentOnly verifies that variant upcast is
// applied to the call argument expression (matching Python ivy_to_cpp.py:
// 3870-3874) and is NOT applied around the call result (the removed
// result-side upcast at old action.go:459).
func TestCallSiteVariantUpcastOnArgumentOnly(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
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
`)
	mod.PublicActions.Set("receive", false)
	out, err := Generate(mod, Config{ClassName: "upcast"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// At the call site inside sender, the argument y:a is upcast to t.
	// Python emits: receive(t(0, new t::twrap<a>(y)));
	want := "receive(t(0, new t::twrap<a>(y)));"
	if !strings.Contains(out.Impl, want) {
		t.Fatalf("expected argument-side upcast %q in impl:\n%s", want, out.Impl)
	}
	compileGeneratedCPP(t, out)
}

// --- TODO 010 tests: derived definitions and sort constructors ---

// TestDerivedDefinitionEmitsMethod verifies that a derived definition
// of a relation is emitted as a C++ method with the same body shape as
// Python emit_some_action: declare a primary-return local, assign rhs
// to it, return it. Mirrors Python emit_derived (ivy_to_cpp.py:1351-1362)
// composed with emit_some_action (ivy_to_cpp.py:1592-1625).
func TestDerivedDefinitionEmitsMethod(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t = {a, b, c}
relation eq(X:t, Y:t)
definition eq(X:t, Y:t) = X = Y
action step(p:t, q:t) = {
    assert eq(p, q)
}
`)
	out, err := Generate(mod, Config{ClassName: "der"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Header: forward declaration.
	if want := "bool eq(t X, t Y);"; !strings.Contains(out.Header, want) {
		t.Fatalf("missing %q in header:\n%s", want, out.Header)
	}
	// Impl: signature + body shape.
	for _, want := range []string{
		"bool der::eq(der::t X, der::t Y)",
		"bool val = false;",
		"val = (X == Y);",
		"return val;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	// Relation must NOT be emitted as mutable state — derived defs are methods.
	if strings.Contains(out.Header, "std::map") && strings.Contains(out.Header, "eq;") {
		t.Fatalf("derived relation should not be emitted as state:\n%s", out.Header)
	}
	compileGeneratedCPP(t, out)
}

// TestZeroArgDerivedDefinitionEmitsMethod verifies that a derived
// definition with no parameters emits a no-arg C++ method. The synthetic
// AssignAction's formal_params is empty.
func TestZeroArgDerivedDefinitionEmitsMethod(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
relation flag
definition flag = true
action step = {
    assert flag
}
`)
	out, err := Generate(mod, Config{ClassName: "zd"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if want := "bool flag();"; !strings.Contains(out.Header, want) {
		t.Fatalf("missing %q in header:\n%s", want, out.Header)
	}
	for _, want := range []string{
		"bool zd::flag()",
		"val = true;",
		"return val;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

// TestDerivedDefinitionStructParamUsesConstRef verifies that derived
// definitions go through the same ptype annotation pipeline as ordinary
// actions: an uninterpreted sort registered in NativeTypes via
// `interpret S -> <<< ... >>>` produces ConstRefType for derived-def
// parameters (Python emit_method_decl, ivy_to_cpp.py:1564 + 1489-1491).
func TestDerivedDefinitionStructParamUsesConstRef(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx
interpret idx -> <<< int >>>
relation lt(X:idx, Y:idx)
definition lt(X:idx, Y:idx) = X < Y
`)
	out, err := Generate(mod, Config{ClassName: "crd"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// idx is registered in NativeTypes by `interpret ... -> <<< int >>>`
	// (compiler_decl.go:1047). isStructSort returns true → ConstRefType.
	for _, want := range []string{
		"bool lt(const idx& X, const idx& Y);",
		"bool crd::lt(const crd::idx& X, const crd::idx& Y)",
	} {
		if !strings.Contains(out.Header+out.Impl, want) {
			t.Fatalf("missing %q:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

// TestSortConstructorEmitsMethod verifies that an explicit constructor
// declaration over a sort with destructors emits a C++ method that
// assigns each destructor field from the corresponding formal. Mirrors
// Python emit_constructor (ivy_to_cpp.py:1364-1375). FixConstructors
// (compiler_ivy_compile.go:1052) rebuilds zero-arg constructors with
// the destructor range sorts as the constructor's domain.
func TestSortConstructorEmitsMethod(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type point
destructor x_field(P:point) : bool
destructor y_field(P:point) : bool
constructor mkpoint : point
`)
	out, err := Generate(mod, Config{ClassName: "mk"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Confirm the test setup actually produced a constructor in the module.
	conss, ok := mod.SortConstructors.Get2("point")
	if !ok || len(conss) == 0 {
		t.Fatalf("test setup: no constructor for sort point; module: %+v", mod.SortConstructors)
	}
	cname := conss[0].Name
	// Header: forward declaration. Two bool destructor fields → two bool args.
	wantHdr := fmt.Sprintf("point %s(bool X0, bool X1);", cname)
	if !strings.Contains(out.Header, wantHdr) {
		t.Fatalf("missing %q in header:\n%s", wantHdr, out.Header)
	}
	// Impl: signature + per-destructor assignments + return.
	wantSig := fmt.Sprintf("point mk::%s(bool X0, bool X1)", cname)
	for _, want := range []string{
		wantSig,
		"val.x_field = X0;",
		"val.y_field = X1;",
		"return val;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

// TestConstructorExcludedFromStateSymbols verifies that constructors,
// even though they are emitted as methods, do not also appear in the
// class's state-variable declarations. allStateSymbols filters by
// g.Mod.Sig.Constructors (generator.go:559 + Python ivy_to_cpp.py:35).
func TestConstructorExcludedFromStateSymbols(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type point
destructor x_field(P:point) : bool
constructor mkpoint : point
`)
	out, err := Generate(mod, Config{ClassName: "mk2"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	conss, ok := mod.SortConstructors.Get2("point")
	if !ok || len(conss) == 0 {
		t.Fatalf("test setup: no constructor for sort point")
	}
	cname := conss[0].Name
	// The constructor name must not appear as a state field — only as a method.
	// State fields would look like `point mkpoint;` (no parens).
	stateDecl := fmt.Sprintf("    %s;\n", cname)
	if strings.Contains(out.Header, stateDecl) {
		t.Fatalf("constructor %s leaked into state declarations:\n%s", cname, out.Header)
	}
	// And it should appear as a method.
	methodForm := fmt.Sprintf("%s(", cname)
	if !strings.Contains(out.Header, methodForm) {
		t.Fatalf("constructor %s missing as method:\n%s", cname, out.Header)
	}
}

// --- TODO 011 tests: complete expression emission ---

func newTestGeneratorWithInterps(interps map[string]string) *Generator {
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Sig = goivy.NewSigOn(mod.Cfg.IuCfg)
	for sortName, interp := range interps {
		mod.Sig.Interp[sortName] = interp
	}
	return &Generator{Mod: mod}
}

func TestEmitLetExpression(t *testing.T) {
	// let x := 7 in (x + x), built as Apply(+, x, x) with a LogicDefinition x := 7.
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

	got, err := newTestGeneratorWithInterps(nil).emitExpr(let)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if strings.Contains(got, "let ") {
		t.Fatalf("let was not expanded: %s", got)
	}
	if !strings.Contains(got, "7 + 7") {
		t.Fatalf("let substitution missing 7+7: %s", got)
	}
}

func TestEmitMacroExpansionInApply(t *testing.T) {
	// a <= b should expand to (a < b) || (a == b).
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Cfg.IuCfg.UsePolymorphicMacros = true
	mod.Sig = goivy.NewSigOn(mod.Cfg.IuCfg)

	s := &goivy.UninterpretedSort{Name: "nat"}
	leSort, err := goivy.NewFunctionSort(s, s, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	leSym := goivy.NewConst("<=", leSort)
	a := goivy.NewConst("a", s)
	b := goivy.NewConst("b", s)
	app := goivy.MustApply(leSym, a, b)

	if !goivy.IsMacro(app, mod.Cfg.IuCfg) {
		t.Fatalf("fixture <= should be a macro")
	}
	got, err := (&Generator{Mod: mod}).emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if !strings.Contains(got, "<") || !strings.Contains(got, "||") {
		t.Fatalf("expected expanded `<` / `||`, got: %s", got)
	}
}

func TestEmitNatSaturationOnMinus(t *testing.T) {
	g := newTestGeneratorWithInterps(map[string]string{"nat": "nat"})
	natSort := &goivy.UninterpretedSort{Name: "nat"}
	fnSort, err := goivy.NewFunctionSort(natSort, natSort, natSort)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	minus := goivy.NewConst("-", fnSort)
	a := goivy.NewConst("a", natSort)
	b := goivy.NewConst("b", natSort)
	app := goivy.MustApply(minus, a, b)

	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if !strings.Contains(got, "__a < __b ? 0") {
		t.Fatalf("expected nat saturation, got: %s", got)
	}
}

func TestEmitRangeSaturationOnArithmetic(t *testing.T) {
	g := newTestGeneratorWithInterps(nil)
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "5"}}
	fnSort, err := goivy.NewFunctionSort(rng, rng, rng)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	plus := goivy.NewConst("+", fnSort)
	a := goivy.NewConst("a", rng)
	b := goivy.NewConst("b", rng)
	app := goivy.MustApply(plus, a, b)

	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if !strings.Contains(got, "__x < 0") || !strings.Contains(got, "5 < __x") {
		t.Fatalf("expected range clamp, got: %s", got)
	}
}

func TestEmitCastToRange(t *testing.T) {
	g := newTestGeneratorWithInterps(nil)
	src := &goivy.UninterpretedSort{Name: "int"}
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "1"}, Ub: goivy.NumeralBound{Value: "9"}}
	castSort, err := goivy.NewFunctionSort(src, rng)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	castSym := goivy.NewConst("cast", castSort)
	v := goivy.NewConst("v", src)
	app := goivy.MustApply(castSym, v)

	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if !strings.Contains(got, "v < 1 ? 1") || !strings.Contains(got, "9 < v ? 9") {
		t.Fatalf("expected range cast clamp, got: %s", got)
	}
}

func TestEmitCastToNat(t *testing.T) {
	g := newTestGeneratorWithInterps(map[string]string{"nat": "nat"})
	src := &goivy.UninterpretedSort{Name: "int"}
	natSort := &goivy.UninterpretedSort{Name: "nat"}
	castSort, err := goivy.NewFunctionSort(src, natSort)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	castSym := goivy.NewConst("cast", castSort)
	v := goivy.NewConst("v", src)
	app := goivy.MustApply(castSym, v)

	got, err := g.emitExpr(app)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if !strings.Contains(got, "__x < 0 ? 0") {
		t.Fatalf("expected nat cast saturation, got: %s", got)
	}
}

func TestEmitStringInterpConstantZero(t *testing.T) {
	g := newTestGeneratorWithInterps(map[string]string{"str": "strlit"})
	strSort := &goivy.UninterpretedSort{Name: "str"}
	zero := goivy.NewConst("0", strSort)

	got, err := g.emitExpr(zero)
	if err != nil {
		t.Fatalf("emitExpr: %v", err)
	}
	if got != `""` {
		t.Fatalf("expected \"\" for strlit 0, got: %s", got)
	}
}

func TestEmitStringInterpConstantNonzeroErrors(t *testing.T) {
	g := newTestGeneratorWithInterps(map[string]string{"str": "strlit"})
	strSort := &goivy.UninterpretedSort{Name: "str"}
	five := goivy.NewConst("5", strSort)

	_, err := g.emitExpr(five)
	if err == nil || !strings.Contains(err.Error(), "string sort") {
		t.Fatalf("expected string-sort numeral error, got: %v", err)
	}
}

func TestEmitNamedBinderUnsupportedIsActionable(t *testing.T) {
	nb, err := goivy.NewNamedBinder("custom", nil, nil, goivy.True)
	if err != nil {
		t.Fatalf("NewNamedBinder: %v", err)
	}
	_, err = (&Generator{}).emitExpr(nb)
	if err == nil || !strings.Contains(err.Error(), "named binder") {
		t.Fatalf("expected named-binder error, got: %v", err)
	}
}

func TestEmitIfSomeMinimizing(t *testing.T) {
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "4"}}
	p := goivy.NewConst("v", rng)
	some := &goivy.SomeCondition{
		Params: []*goivy.Const{p},
		Fmla:   goivy.NewConst("true", goivy.Boolean),
		Kind:   "some_min",
		Index:  p,
	}
	thenAct := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("true", goivy.Boolean))
	ifAct := goivy.NewIfAction(nil, thenAct)
	ifAct.Cond = some

	var w cppWriter
	(&Generator{}).emitAction(&w, ifAct)
	got := w.String()
	if strings.Contains(got, "not supported yet") {
		t.Fatalf("some_min still hits unsupported sentinel:\n%s", got)
	}
	for _, want := range []string{"bool __ivy_some", "__ivy_some_idx", "for (unsigned v = 0; v < 5; v++)"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in some_min lowering:\n%s", want, got)
		}
	}
}

func TestEmitIfSomeMaximizing(t *testing.T) {
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "4"}}
	p := goivy.NewConst("v", rng)
	some := &goivy.SomeCondition{
		Params: []*goivy.Const{p},
		Fmla:   goivy.NewConst("true", goivy.Boolean),
		Kind:   "some_max",
		Index:  p,
	}
	thenAct := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("true", goivy.Boolean))
	ifAct := goivy.NewIfAction(nil, thenAct)
	ifAct.Cond = some

	var w cppWriter
	(&Generator{}).emitAction(&w, ifAct)
	got := w.String()
	if strings.Contains(got, "not supported yet") {
		t.Fatalf("some_max still hits unsupported sentinel:\n%s", got)
	}
	// Min and max differ only in the comparison direction; assert the
	// reversed compare and the witness assignment are present.
	if !strings.Contains(got, "__ivy_some_idx") || !strings.Contains(got, "__ivy_some_cur") {
		t.Fatalf("missing min/max bookkeeping vars:\n%s", got)
	}
}

// TestDerivedAndConstructorCompile (SLOW_CPP_TEST) — end-to-end compile
// of a fixture combining a derived definition and a sort constructor.
func TestDerivedAndConstructorCompile(t *testing.T) {
	if !SlowCppTest {
		t.Skip("set SLOW_CPP_TEST=1 to run this test")
	}
	mod := compileIvySource(t, `#lang ivy1.7
type t = {a, b}
relation eq(X:t, Y:t)
definition eq(X:t, Y:t) = X = Y
type point
destructor x_field(P:point) : bool
constructor mkpoint : point
`)
	out, err := Generate(mod, Config{ClassName: "both"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	compileGeneratedCPP(t, out)
}

// TODO 012 tests — exercise inequality-derived bounds, cardinality
// attribute bounds, multi-variable bound filtering, and the
// first-param-is-index break in `if some X. ... minimizing X`.

// TestEmitQuantInequalityBoundOverRange checks that `forall X:t. X < N -> p(X)`
// uses the inequality `X < 5` as upper bound rather than the type's
// full range. Mirrors Python emit_quant via get_bounds.
func TestEmitQuantInequalityBoundOverRange(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..10}
relation p(X:idx)
after init {
    p(X) := false
}
action check = {
    assert forall X:idx. X < 5 -> p(X)
}
`)
	out, err := Generate(mod, Config{ClassName: "qbound"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// Upper bound from inequality should be `5` (range-clamped via the
	// same ternary Python emits in emit_constant:3010-3014). The init
	// block still uses `X <= 10`; only the `check` quantifier loop must
	// use the inequality.
	if !strings.Contains(out.Impl, "X < (5 < 0 ? 0 : 10 < 5 ? 10 : 5)") {
		t.Fatalf("expected inequality-derived upper bound from literal `5`:\n%s", out.Impl)
	}
	checkIdx := strings.Index(out.Impl, "::check()")
	if checkIdx < 0 {
		t.Fatalf("missing check() function:\n%s", out.Impl)
	}
	tail := out.Impl[checkIdx:]
	if strings.Contains(tail, "X <= 10") {
		t.Fatalf("inequality bound should override range bound in check():\n%s", tail)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

// TestEmitSomeMinMaxBreakWhenIndexIsFirstParam asserts that the
// generated `if some` minimizing the loop variable adds the early `break`
// (Python emit_some:3539-3540).
func TestEmitSomeMinMaxBreakWhenIndexIsFirstParam(t *testing.T) {
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "4"}}
	p := goivy.NewConst("v", rng)
	some := &goivy.SomeCondition{
		Params: []*goivy.Const{p},
		Fmla:   goivy.NewConst("true", goivy.Boolean),
		Kind:   "some_min",
		Index:  p,
	}
	thenAct := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("true", goivy.Boolean))
	ifAct := goivy.NewIfAction(nil, thenAct)
	ifAct.Cond = some

	var w cppWriter
	(&Generator{}).emitAction(&w, ifAct)
	got := w.String()
	if !strings.Contains(got, "break;") {
		t.Fatalf("first-param-is-index minimizing should emit `break;`:\n%s", got)
	}
}

// TestEmitSomeMinMaxNoBreakWhenIndexIsExpression asserts that the
// break-early optimization is suppressed when the minimizer is a
// non-trivial expression (the first param does not equal the index).
func TestEmitSomeMinMaxNoBreakWhenIndexIsExpression(t *testing.T) {
	rng := &goivy.RangeSort{Name: "idx", Lb: goivy.NumeralBound{Value: "0"}, Ub: goivy.NumeralBound{Value: "4"}}
	p := goivy.NewConst("v", rng)
	// Index is a *different* variable — must not break early.
	other := goivy.NewConst("w", rng)
	some := &goivy.SomeCondition{
		Params: []*goivy.Const{p},
		Fmla:   goivy.NewConst("true", goivy.Boolean),
		Kind:   "some_min",
		Index:  other,
	}
	thenAct := goivy.NewAssignAction(goivy.NewConst("flag", goivy.Boolean), goivy.NewConst("true", goivy.Boolean))
	ifAct := goivy.NewIfAction(nil, thenAct)
	ifAct.Cond = some

	var w cppWriter
	(&Generator{}).emitAction(&w, ifAct)
	got := w.String()
	if strings.Contains(got, "break;") {
		t.Fatalf("break-early should not fire when index differs from first param:\n%s", got)
	}
}

// TestEmitIfSomeUsesInequalityBound exercises the bound path inside
// emitIfSomeMinMax: `if some X:t. X < n & p(X) minimizing X` should
// constrain the loop to `X < n` rather than scanning the whole range.
func TestEmitIfSomeUsesInequalityBound(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..10}
relation p(X:idx)
individual saved : idx
after init {
    p(X) := false;
    saved := 0
}
action pick = {
    if some x:idx. x < 5 & p(x) minimizing x {
        saved := x
    }
}
export pick
`)
	out, err := Generate(mod, Config{ClassName: "qsomemin"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, "loc__x < (5 < 0 ? 0 : 10 < 5 ? 10 : 5)") {
		t.Fatalf("expected inequality-derived bound (range-clamped `5`):\n%s", out.Impl)
	}
	if !strings.Contains(out.Impl, "break;") {
		t.Fatalf("first-param-is-index minimizing should emit `break;`:\n%s", out.Impl)
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

// TestEmitQuantMultiVarInequalityFiltersSibling asserts that a
// quantified sibling variable cannot be used as the bound for the first
// quantified variable. Mirrors Python get_all_bounds's `variables[1:]`
// slicing.
func TestEmitQuantMultiVarInequalityFiltersSibling(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..7}
relation r(X:idx, Y:idx)
after init {
    r(X, Y) := false
}
action check = {
    assert forall X:idx, Y:idx. X < Y -> r(X, Y)
}
`)
	out, err := Generate(mod, Config{ClassName: "qmulti"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// X's upper bound must not be Y (Y is also quantified) — it should
	// fall back to the type's cardinality. Y's lower bound *can* mention
	// X since Y is the inner var (others = []).
	for _, want := range []string{"X = 0; X < 8;", "Y = (X)+1; Y < 8;"} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("expected %q in multi-var bound:\n%s", want, out.Impl)
		}
	}
	assertNoUnsupportedCPP(t, out)
	compileGeneratedCPP(t, out)
}

// --- TODO 020 tests: requires/ensures/subgoals and exported action semantics ---

// TestActionGenExtPrecondsPreservesFormals verifies that when an action_gen
// plan wraps an action via ExtPreconds (Python ivy_to_cpp.py:1213-1217), the
// resulting Sequence keeps the original action's lineno, formal_params and
// formal_returns — not the empty defaults a freshly-constructed Sequence
// would have.
//
// Added as part of TODO 020.
func TestActionGenExtPrecondsPreservesFormals(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) returns (r:color) = {
    saved := c;
    r := c
}
export set
`)
	pre := goivy.NewConst("saved", goivy.Boolean)
	if mod.ExtPreconds == nil {
		mod.ExtPreconds = map[string]goivy.Expr{}
	}
	mod.ExtPreconds["set"] = pre

	g := &Generator{Mod: mod, ClassName: "extpre", Config: Config{ClassName: "extpre"}}
	act, ok := mod.Actions.Get2("set")
	if !ok {
		t.Fatalf("action `set` missing")
	}
	wantParams := act.GetFormalParams()
	wantReturns := act.GetFormalReturns()
	plan := g.buildActionGenPlan("set", act)
	gotParams := plan.act.GetFormalParams()
	gotReturns := plan.act.GetFormalReturns()
	if len(gotParams) != len(wantParams) {
		t.Fatalf("plan.act.GetFormalParams() = %d entries, want %d", len(gotParams), len(wantParams))
	}
	for i := range wantParams {
		if gotParams[i].Name != wantParams[i].Name {
			t.Fatalf("plan.act formal param[%d] = %q, want %q", i, gotParams[i].Name, wantParams[i].Name)
		}
	}
	if len(gotReturns) != len(wantReturns) {
		t.Fatalf("plan.act.GetFormalReturns() = %d entries, want %d", len(gotReturns), len(wantReturns))
	}
	for i := range wantReturns {
		if gotReturns[i].Name != wantReturns[i].Name {
			t.Fatalf("plan.act formal return[%d] = %q, want %q", i, gotReturns[i].Name, wantReturns[i].Name)
		}
	}
}

// TestAssertLabelStripsTrailingColonSpace verifies that the C++ label
// passed to ivy_assert mirrors Python's iu.lineno_str (ivy_utils.py:285-
// 291), which strips the trailing ": " from Location.String(). Without
// the fix, labels emit as `"test.ivy: line 5: "` (trailing colon-space);
// Python emits `"test.ivy: line 5"`.
//
// Added as part of TODO 020.
func TestAssertLabelStripsTrailingColonSpace(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action check = {
    assert saved = red
}
export check
`)
	out, err := Generate(mod, Config{ClassName: "labelfmt"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	idx := strings.Index(out.Impl, "ivy_assert(")
	if idx < 0 {
		t.Fatalf("no ivy_assert in impl:\n%s", out.Impl)
	}
	lineEnd := strings.Index(out.Impl[idx:], "\n")
	if lineEnd < 0 {
		lineEnd = len(out.Impl) - idx
	}
	line := out.Impl[idx : idx+lineEnd]
	if strings.Contains(line, `: ");`) {
		t.Fatalf("ivy_assert label ends in trailing colon-space (Python lineno_str strips it):\n%s", line)
	}
}

// TestSubgoalNeverConvertedToAssume verifies that goivy.AssertToAssume
// does NOT downgrade SubgoalAction even when "assert" is in kinds.
// Python ivy_actions.py:396-403 compares type(self) to the kinds list;
// SubgoalAction is its own class so passing [AssertAction] leaves it
// untouched (ivy_actions.py:416-421).
//
// Added as part of TODO 020.
func TestSubgoalNeverConvertedToAssume(t *testing.T) {
	cfg := goivy.NewConfig()
	fmla := goivy.NewConst("p", goivy.Boolean)
	sub := goivy.NewSubgoalAction(fmla)

	kinds := map[string]bool{"assert": true, "require": true}
	out := goivy.AssertToAssume(sub, kinds, cfg.IuCfg)

	if _, ok := out.(*goivy.LogicSubgoalAction); !ok {
		t.Fatalf("AssertToAssume converted SubgoalAction to %T (want *LogicSubgoalAction)", out)
	}

	// Opt-in case: caller passes "subgoal" — now we DO convert.
	kindsOpt := map[string]bool{"subgoal": true}
	out2 := goivy.AssertToAssume(sub, kindsOpt, cfg.IuCfg)
	if _, ok := out2.(*goivy.LogicAssumeAction); !ok {
		t.Fatalf("AssertToAssume with kinds={subgoal} returned %T (want *LogicAssumeAction)", out2)
	}
}

// TestRequiresInExternalBecomesAssume verifies that an exported action's
// external wrapper (`ext:foo`) has its RequiresAction converted to
// AssumeAction by isolation, producing an `ivy_assume(...)` call in the
// C++ for the ext: method while the internal wrapper still emits
// `ivy_assert(...)`. Mirrors Python ivy_isolate.py:1101 +
// ivy_actions.py:413 + ivy_to_cpp.py:3788-3808.
//
// Added as part of TODO 020.
func TestRequiresInExternalBecomesAssume(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step(c:color) = {
    require c = green;
    saved := c
}
export step
`)
	// Emulate CreateIsolate's classification (isolate.go:567-606) for a
	// verified exported action: external wrapper has requires→assume,
	// internal version is unchanged.
	stepAct, ok := mod.Actions.Get2("step")
	if !ok {
		t.Fatalf("step action missing")
	}
	extAct := goivy.AssertToAssume(stepAct, map[string]bool{"require": true}, mod.Cfg.IuCfg)
	mod.Actions.Set("ext:step", extAct)
	mod.PublicActions = goivy.NewInsMap[string, bool]()
	mod.PublicActions.Set("ext:step", true)
	mod.PublicActions.Set("step", false)

	out, err := Generate(mod, Config{ClassName: "reqassume"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	extBody := bodyAfterMarker(out.Impl, "reqassume::ext__step")
	intBody := bodyAfterMarker(out.Impl, "reqassume::step(")
	if extBody == "" {
		t.Fatalf("ext:step method not emitted:\n%s", out.Impl)
	}
	if intBody == "" {
		t.Fatalf("internal step method not emitted:\n%s", out.Impl)
	}
	if !strings.Contains(extBody, "ivy_assume(") {
		t.Fatalf("ext:step body missing ivy_assume(...):\n%s", extBody)
	}
	if !strings.Contains(intBody, "ivy_assert(") {
		t.Fatalf("internal step body missing ivy_assert(...):\n%s", intBody)
	}
}

// bodyAfterMarker returns the substring from the first '{' after the
// marker up to the matching closing brace. Used to extract one method
// body from emitted C++ for the TODO 020 tests.
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

// TestImportCallerTracePrologueInTest verifies that, for target=test,
// the body of an unscoped imported action whose name resolves to a known
// action begins with `__ivy_out << "< name(args)"`. Mirrors Python
// emit_some_action (ivy_to_cpp.py:1604-1607) + find_import_callers
// (ivy_to_cpp.py:1888-1897).
//
// Added as part of TODO 020.
func TestImportCallerTracePrologueInTest(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual seen : color
action callback(c:color) = {
    seen := c
}
import callback
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "traceimp"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	body := bodyAfterMarker(out.Impl, "traceimp::callback(")
	if body == "" {
		t.Fatalf("callback method not emitted:\n%s", out.Impl)
	}
	want := `__ivy_out << "< callback"`
	if !strings.Contains(body, want) {
		t.Fatalf("expected trace prologue %q in callback body:\n%s", want, body)
	}

	// The same module at target=repl must NOT emit the prologue
	// (Python find_import_callers gates on target=="test"; repl is the
	// closest non-test interactive target).
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
		t.Fatalf("Generate (repl): %v", err)
	}
	body2 := bodyAfterMarker(out2.Impl, "traceimp2::callback(")
	if strings.Contains(body2, `__ivy_out << "< callback"`) {
		t.Fatalf("trace prologue should be test-only; appeared in repl target:\n%s", body2)
	}
}

// TestExtPrecondsAppearsInActionGenSMT verifies that an ExtPreconds
// formula on an exported action is propagated into the action_gen's
// SMT-LIB precondition via reverse_image(Sequence(Assume(pre), action)).
// Python ivy_to_cpp.py:1213-1239.
//
// Added as part of TODO 020.
func TestExtPrecondsAppearsInActionGenSMT(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	redEntry, ok := mod.Sig.Symbols.Get2("red")
	if !ok {
		t.Fatalf("symbol red not found")
	}
	savedEntry, ok := mod.Sig.Symbols.Get2("saved")
	if !ok {
		t.Fatalf("symbol saved not found")
	}
	if mod.ExtPreconds == nil {
		mod.ExtPreconds = map[string]goivy.Expr{}
	}
	eq, err := goivy.NewEq(
		goivy.NewConst("saved", savedEntry.Sort),
		goivy.NewConst("red", redEntry.Sort),
	)
	if err != nil {
		t.Fatalf("NewEq: %v", err)
	}
	mod.ExtPreconds["set"] = eq

	out, err := Generate(mod, Config{Target: "gen", ClassName: "extpregen"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	idx := strings.Index(out.Impl, "set_gen::set_gen")
	if idx < 0 {
		t.Fatalf("set_gen constructor not emitted:\n%s", out.Impl)
	}
	end := strings.Index(out.Impl[idx:], "bool set_gen::generate")
	if end < 0 {
		end = len(out.Impl) - idx
	}
	body := out.Impl[idx : idx+end]
	if !strings.Contains(body, `add(std::string("(assert `) {
		t.Fatalf("set_gen constructor missing SMT-LIB add:\n%s", body)
	}
	if !strings.Contains(body, "saved") || !strings.Contains(body, "red") {
		t.Fatalf("set_gen constructor SMT-LIB missing references to ext_preconds saved/red:\n%s", body)
	}
}

// --- TODO 021 tests: debug/trace output semantics ---

// TestDebugActionEmitsNamedValues confirms that the WithNames slice on
// LogicDebugAction reaches emitDebug, producing `"<name>" : <value>`
// lines instead of the legacy `"value%d"` keys. Python
// ivy_to_cpp.py:4011-4020. The `debug` keyword isn't reserved in ivy1.7
// (lexer.go:204-209), so we synthesize the action directly.
//
// Added as part of TODO 021.
func TestDebugActionEmitsNamedValues(t *testing.T) {
	mod := goivy.New()
	mod.Name = "named"
	mod.Relations.Set("flag", goivy.Boolean)
	mod.Relations.Set("saved", goivy.Boolean)
	dbg := goivy.NewDebugAction(
		goivy.NewConst(`"tick"`, goivy.TopS),
		goivy.NewConst("saved", goivy.Boolean),
		goivy.NewConst("flag", goivy.Boolean),
	)
	dbg.WithNames = []string{"c", "f"}
	mod.Actions.Set("step", dbg)
	out, err := Generate(mod, Config{ClassName: "named"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		`"event\" : \"tick\"`,
		`"c\" : `,
		`"f\" : `,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	if strings.Contains(out.Impl, `"value0\"`) || strings.Contains(out.Impl, `"value1\"`) {
		t.Fatalf("debug values still using legacy value%%d keys:\n%s", out.Impl)
	}
}

// TestDebugActionEmitsQuantifiedLoop verifies that a debug value
// expression carrying a free variable opens a loop over that variable
// and wraps the print with `std::cout << "[" ... "]"`. Python
// emit_print_expr (ivy_to_cpp.py:3989-3996). Synthesized directly
// because `debug` is not reserved in ivy1.7.
//
// Added as part of TODO 021.
func TestDebugActionEmitsQuantifiedLoop(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..3}
relation marked(I:idx)
`)
	idxSort, ok := mod.Sig.Sorts.Get2("idx")
	if !ok {
		t.Fatal("idx sort missing")
	}
	relSort, err := goivy.NewFunctionSort(idxSort, goivy.Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	markedFn := goivy.NewConst("marked", relSort)
	iVar, err := goivy.NewVariable("I", idxSort)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	app, err := goivy.NewApply(markedFn, iVar)
	if err != nil {
		t.Fatalf("NewApply: %v", err)
	}
	dbg := goivy.NewDebugAction(
		goivy.NewConst(`"scan"`, goivy.TopS),
		app,
	)
	dbg.WithNames = []string{"seen"}
	mod.Actions.Set("step", dbg)

	out, err := Generate(mod, Config{ClassName: "quant"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		`std::cout << "[";`,
		`std::cout << "]";`,
		`std::cout << (marked[I]);`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
}

// TestNumberFormatHexFromRadixAttribute synthesizes a module with
// `attribute radix = "16"` and confirms that the trace prologue + REPL
// assert override inject `<< std::hex << std::showbase`. Python
// ivy_to_cpp.py:1935-1938.
//
// Added as part of TODO 021.
func TestNumberFormatHexFromRadixAttribute(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action step(c:color) = {
    saved := c
}
import step
`)
	mod.Attributes["radix"] = "16"
	out, err := Generate(mod, Config{Target: "test", ClassName: "hexfmt"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, `__ivy_out << std::hex << std::showbase << "< step"`) {
		t.Fatalf("expected number_format prefix on trace prologue:\n%s", out.Impl)
	}
}

// TestActionGenExecuteWrapsTraceBraces verifies that under target=gen
// with Trace=true, the action_gen execute body opens with `< name(args)`,
// then `{`, then the call, then `}`. Python ivy_to_cpp.py:1331-1346.
//
// Added as part of TODO 021.
func TestActionGenExecuteWrapsTraceBraces(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "gentrace", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	body := bodyAfterMarker(out.Impl, "void set_gen::execute")
	if body == "" {
		t.Fatalf("set_gen::execute not found in impl:\n%s", out.Impl)
	}
	for _, want := range []string{
		`__ivy_out << "> set("`,
		`__ivy_out << "{" << std::endl;`,
		`obj.set(`,
		`__ivy_out << "}" << std::endl;`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in set_gen::execute body:\n%s", want, body)
		}
	}

	// Without Trace=true, only the `>` line should be present; no `{`/`}`.
	out2, err := Generate(mod, Config{Target: "gen", ClassName: "gentrace2"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	body2 := bodyAfterMarker(out2.Impl, "void set_gen::execute")
	if strings.Contains(body2, `__ivy_out << "{" << std::endl;`) {
		t.Fatalf("trace `{` brace should be Trace-gated:\n%s", body2)
	}
}

// TestImportCallerBodyWrappedInBracesUnderTrace verifies that the
// imported-action method body, in target=test + Trace=true, opens with
// the prologue and a `{` and closes with `}` — matching Python
// ivy_to_cpp.py:1604-1619.
//
// Added as part of TODO 021.
func TestImportCallerBodyWrappedInBracesUnderTrace(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual seen : color
action callback(c:color) = {
    seen := c
}
import callback
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "tracedimp", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	body := bodyAfterMarker(out.Impl, "tracedimp::callback(")
	if body == "" {
		t.Fatalf("callback method body not found:\n%s", out.Impl)
	}
	for _, want := range []string{
		`__ivy_out << "< callback"`,
		`__ivy_out << "{" << std::endl;`,
		`__ivy_out << "}" << std::endl;`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in traced callback body:\n%s", want, body)
		}
	}
}

// TestAssignSimpleEmitsWriteTraceUnderTrace verifies that an Ivy
// assignment `x := y` emits an extra `__ivy_out << "  write(x," << (y)
// << ")"` line above the C++ assignment when Config.Trace is set, and
// only then. Python ivy_to_cpp.py:3627-3650.
//
// Added as part of TODO 021.
func TestAssignSimpleEmitsWriteTraceUnderTrace(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action step(c:color) = {
    saved := c
}
export step
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "writetrace", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Impl, `__ivy_out << "  write(`) {
		t.Fatalf("expected write-trace line under Trace=true:\n%s", out.Impl)
	}

	// Without Trace, no write line.
	mod2 := compileIvySource(t, src)
	out2, err := Generate(mod2, Config{Target: "repl", ClassName: "writetrace2"})
	if err != nil {
		t.Fatalf("Generate (no trace): %v", err)
	}
	if strings.Contains(out2.Impl, `__ivy_out << "  write(`) {
		t.Fatalf("write-trace line should be Trace-gated:\n%s", out2.Impl)
	}
}

// TODO 023 — Python `ivy_to_cpp.py:1952` emits this define unconditionally
// at the top of the header. Used to suppress MSVC's iterator debug overhead.
func TestHeaderPreambleEmitsIteratorDebuggingDefine(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Header, "#define _HAS_ITERATOR_DEBUGGING 0") {
		t.Fatalf("expected _HAS_ITERATOR_DEBUGGING define in header:\n%s", out.Header)
	}
}

// TODO 023 — Python `ivy_to_cpp.py:1949-1951` emits these only when the
// codegen host is Windows. Go uses Config.HostOS to expose the gate.
func TestHeaderPreambleEmitsWindowsHostIncludesWhenHostOSIsWindows(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle", HostOS: "windows"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	wantDefine := "#define WIN32_LEAN_AND_MEAN"
	wantInclude := "#include <windows.h>"
	if !strings.Contains(out.Header, wantDefine) {
		t.Fatalf("expected %q in header:\n%s", wantDefine, out.Header)
	}
	if !strings.Contains(out.Header, wantInclude) {
		t.Fatalf("expected %q in header:\n%s", wantInclude, out.Header)
	}
	defIdx := strings.Index(out.Header, wantDefine)
	incIdx := strings.Index(out.Header, wantInclude)
	algIdx := strings.Index(out.Header, "#include <algorithm>")
	if defIdx < 0 || incIdx < 0 || algIdx < 0 {
		t.Fatalf("expected define, windows.h, and algorithm in header; got defIdx=%d incIdx=%d algIdx=%d", defIdx, incIdx, algIdx)
	}
	if !(defIdx < incIdx && incIdx < algIdx) {
		t.Fatalf("expected order: WIN32_LEAN_AND_MEAN(%d) < <windows.h>(%d) < <algorithm>(%d)\n%s",
			defIdx, incIdx, algIdx, out.Header)
	}
}

// TODO 023 — Python `ivy_to_cpp.py:1948` only emits Windows preamble when
// the codegen host is Windows. Non-Windows hosts skip these lines.
func TestHeaderPreambleSkipsWindowsHostIncludesByDefaultOnNonWindowsHost(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle", HostOS: "linux"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, unwanted := range []string{"#define WIN32_LEAN_AND_MEAN", "#include <windows.h>"} {
		if strings.Contains(out.Header, unwanted) {
			t.Fatalf("did not expect %q in header on non-Windows host:\n%s", unwanted, out.Header)
		}
	}
}

// TODO 023 — Python `ivy_to_cpp.py:2210-2211` emits `ivy_z3_helpers.hpp`
// in the impl preamble, immediately after `ivy_value.hpp`/`ivy_repl.hpp`,
// before any per-sort Z3 template specializations. Go's `ivy_go_z3.hpp`
// occupies that slot.
func TestImplPreambleEmitsZ3HelperIncludeAfterReplInclude(t *testing.T) {
	src := `#lang ivy1.7
type color = {red, green}
individual saved : color
action set(c:color) = {
    saved := c
}
export set
`
	mod := compileIvySource(t, src)
	out, err := Generate(mod, Config{Target: "test", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	const z3Inc = `#include "ivy_go_z3.hpp"`
	const replInc = `#include "ivy_repl.hpp"`
	const valueInc = `#include "ivy_value.hpp"`
	z3Idx := strings.Index(out.Impl, z3Inc)
	replIdx := strings.Index(out.Impl, replInc)
	valueIdx := strings.Index(out.Impl, valueInc)
	if z3Idx < 0 {
		t.Fatalf("expected %q in impl for test target:\n%s", z3Inc, out.Impl)
	}
	if replIdx < 0 || valueIdx < 0 {
		t.Fatalf("expected value/repl includes in impl: valueIdx=%d replIdx=%d\n%s", valueIdx, replIdx, out.Impl)
	}
	if !(valueIdx < replIdx && replIdx < z3Idx) {
		t.Fatalf("expected ivy_value.hpp(%d) < ivy_repl.hpp(%d) < ivy_go_z3.hpp(%d):\n%s",
			valueIdx, replIdx, z3Idx, out.Impl)
	}
	// Z3 helper include must precede the inline `__from_solver` template
	// definitions emitted by emitZ3SolverTemplates.
	if tmplIdx := strings.Index(out.Impl, "void __from_solver"); tmplIdx >= 0 && tmplIdx < z3Idx {
		t.Fatalf("ivy_go_z3.hpp(%d) must precede __from_solver template definitions(%d):\n%s",
			z3Idx, tmplIdx, out.Impl)
	}
}

// TODO 023 — Repl-only target does not need the Z3 runtime helper.
func TestImplPreambleOmitsZ3HelperIncludeForReplTarget(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
action step = {
}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if strings.Contains(out.Impl, `#include "ivy_go_z3.hpp"`) {
		t.Fatalf("ivy_go_z3.hpp must not be emitted for repl target:\n%s", out.Impl)
	}
}

// TODO 025 — Conjecture/property/isolate integration tests.
//
// findFirstAssertInAction walks the action tree of `a` and returns the
// first *LogicAssertAction it encounters. Uses IterSubactions which
// returns a flat list (self plus all descendants), so we walk it
// without recursing.
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

// TestConjectureAppendsAssertToPublicActions mirrors Python
// ivy_to_cpp.py:4495-4503 add_conjs_to_actions. After
// prepareModuleForCPP runs on a module with a conjecture and one
// exported action, the action's body must end with a sequence that
// contains an ivy_assert(...) for the conjecture.
func TestConjectureAppendsAssertToPublicActions(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
relation r
after init { r := false }
action step = { r := true }
export step
conjecture r
`)
	out, err := Generate(mod, Config{Target: "impl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	stepAct, ok := mod.Actions.Get2("step")
	if !ok {
		t.Fatalf("step action missing after prepareModuleForCPP")
	}
	if findFirstAssertInAction(stepAct) == nil {
		t.Fatalf("expected conjecture assert appended to exported action step; not found in:\n%v", stepAct)
	}
	if !strings.Contains(out.Impl, "ivy_assert(") {
		t.Fatalf("expected ivy_assert(...) in generated impl after conjecture insertion:\n%s", out.Impl)
	}
}

// TestConjectureAddsCheckInvariantsInitializer mirrors Python
// ivy_to_cpp.py:4500 — the conjecture sequence is appended to
// im.module.initializers as "__check_invariants".
func TestConjectureAddsCheckInvariantsInitializer(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
relation r
after init { r := false }
action step = { r := true }
export step
conjecture r
`)
	if _, err := Generate(mod, Config{Target: "impl", ClassName: "oracle"}); err != nil {
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
		for _, i := range mod.Initializers {
			names = append(names, i.Name)
		}
		t.Fatalf("__check_invariants initializer not found; have %v", names)
	}
}

// TestPropertyMovedIntoAxioms mirrors Python ivy_to_cpp.py:4635-4636 —
// labeled_props are appended to labeled_axioms and then cleared.
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
	if _, err := Generate(mod, Config{Target: "impl", ClassName: "oracle"}); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(mod.LabeledProps) != 0 {
		t.Fatalf("expected LabeledProps cleared after prepareModuleForCPP, got %d", len(mod.LabeledProps))
	}
	if len(mod.LabeledAxioms) != beforeAxioms+beforeProps {
		t.Fatalf("expected LabeledAxioms to grow by %d (properties moved in); was %d, now %d",
			beforeProps, beforeAxioms, len(mod.LabeledAxioms))
	}
}

// TestConjectureLinenoPropagatesToAssert verifies Python
// ivy_to_cpp.py:4496 `set_lineno(conj.lineno)` is honored — the assert
// appended for the conjecture carries the same line as the
// conjecture's source location, and the generated C++ assert label
// references that line.
func TestConjectureLinenoPropagatesToAssert(t *testing.T) {
	// The conjecture is on a known line of the input; record its line
	// number relative to the source string so the test is robust to
	// edits above.
	src := "#lang ivy1.7\n" +
		"relation r\n" +
		"after init { r := false }\n" +
		"action step = { r := true }\n" +
		"export step\n" +
		"conjecture r\n" // ← this is line 6
	const conjLine = 6
	mod := compileIvySource(t, src)
	if len(mod.LabeledConjs) == 0 {
		t.Fatalf("test setup: expected at least one LabeledConj, got 0")
	}
	if got := mod.LabeledConjs[0].GetLineno().Line; got != conjLine {
		t.Fatalf("test setup: expected conjecture on line %d, got %d", conjLine, got)
	}
	out, err := Generate(mod, Config{Target: "impl", ClassName: "oracle"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
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
	// The emitted ivy_assert label embeds the line via linenoStr,
	// which produces "<file>: line <N>" (see action.go:200-207).
	wantFragment := fmt.Sprintf("line %d", conjLine)
	if !strings.Contains(out.Impl, wantFragment) {
		t.Fatalf("expected generated impl to reference conjecture %q; impl:\n%s", wantFragment, out.Impl)
	}
}

// TODO 027: per-sort REPL-value parsing tests. Each test locks down both
// the good-value accept path and the bad-value error throw path inside
// each emitted `_arg<T>` body, for the sort categories not already
// covered by TestEmitsArgSpecForEnum and TestDestructorArgShape.

// TestArgSpecVariantBadAndGoodValuePaths locks down the variant
// supertype `_arg<>` body shape, mirroring Python ivy_to_cpp.py via
// ivy_cpp_types.py:374-388 (the VariantType.emit_templates `_arg<>`
// branch).
func TestArgSpecVariantBadAndGoodValuePaths(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type t
variant a of t
variant b of t
individual v : t
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "vrun"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		// _arg<vrun::t> opener.
		"template <> vrun::t _arg<vrun::t>(std::vector<ivy_value> &args, unsigned idx, long long bound) {",
		// Bad value: atom rejected.
		`throw out_of_bounds("unexpected value for sort t: " + args[idx].atom, args[idx].pos);`,
		// Bad value: multiple fields rejected.
		`throw out_of_bounds("too many fields for sort t (expected one)", args[idx].pos);`,
		// Good value paths: one per subtype, with upcast.
		`if (args[idx].fields[0].atom == "a") return vrun::t(0, new vrun::t::twrap<vrun::a>(_arg<vrun::a>(args[idx].fields[0].fields, 0, 0)));`,
		`if (args[idx].fields[0].atom == "b") return vrun::t(1, new vrun::t::twrap<vrun::b>(_arg<vrun::b>(args[idx].fields[0].fields, 0, 0)));`,
		// Bad value: unknown subtype rejected.
		`throw out_of_bounds("unexpected field sort t: " + args[idx].fields[0].atom, args[idx].pos);`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

// TestArgSpecStrBVBadAndGoodValuePaths locks down the strbv `_arg<>`
// body shape, mirroring Python ivy_cpp_types.py:147-152 (StrBV.emit_templates).
func TestArgSpecStrBVBadAndGoodValuePaths(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type text
interpret text -> strbv[4]
individual saved : text
action set(t:text) = { saved := t }
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "srun"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"template <> srun::text _arg<srun::text>(std::vector<ivy_value> &args, unsigned idx, long long bound) {",
		// Bad value: nested fields rejected.
		"if (args[idx].fields.size()) {",
		"throw out_of_bounds(idx);",
		// Good value: atom returned as-is.
		"return args[idx].atom;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

// TestArgSpecIntBVBadAndGoodValuePaths locks down the intbv `_arg<>`
// body shape, mirroring Python ivy_cpp_types.py:211-223 (IntBV.emit_templates).
func TestArgSpecIntBVBadAndGoodValuePaths(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type small
interpret small -> intbv[0][7][3]
individual saved : small
action set(x:small) = { saved := x }
export set
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "irun"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		"template <> irun::small _arg<irun::small>(std::vector<ivy_value> &args, unsigned idx, long long bound) {",
		// Bad value: nested fields rejected.
		"if (args[idx].fields.size()) {",
		"throw out_of_bounds(idx);",
		// Good value: atom parsed via istringstream into res.val.
		"std::istringstream s(args[idx].atom.c_str());",
		"s >> res.val;",
		"return res;",
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing %q in impl:\n%s", want, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

// TestReplDispatchUsesPrimitiveArgForBoolRangeNatStrlit verifies the
// dispatch chain in emitDispatchArgExprs (repl.go:282-291) picks the
// right runtime `_arg<>` primitive specialization for each primitive
// sort kind. The matching `_arg<bool>`, `_arg<unsigned>`, `_arg<unsigned
// long long>`, and `_arg<__strlit>` template specializations live in
// include2cpp/ivy_value.hpp.
func TestReplDispatchUsesPrimitiveArgForBoolRangeNatStrlit(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type rng = {0..5}
type bigint
interpret bigint -> nat
type str
interpret str -> strlit
action poke(b:bool, r:rng, n:bigint, s:str) = {}
export poke
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "prim"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// rng cardinality 6 → bound 6. bool bound 2. nat/strlit are
	// unbounded — Python passes csortcard which is 0 in those cases.
	wantDispatch := `ivy.poke(_arg<bool>(args, 0, 2), _arg<unsigned>(args, 1, 6), _arg<unsigned long long>(args, 2, 0), _arg<__strlit>(args, 3, 0));`
	if !strings.Contains(out.Impl, wantDispatch) {
		t.Fatalf("missing dispatch call %q in impl:\n%s", wantDispatch, out.Impl)
	}
	// No per-sort _arg<> spec should have been emitted for these
	// primitive-dispatched sorts.
	for _, banned := range []string{
		"_arg<prim::rng>",
		"_arg<prim::bigint>",
		"_arg<prim::str>",
	} {
		if strings.Contains(out.Impl, banned) {
			t.Fatalf("unexpected per-sort _arg<> %q in impl:\n%s", banned, out.Impl)
		}
	}
	compileGeneratedCPP(t, out)
}

// TestValueParserPrefixesLinenoOnError mirrors Python's
// emit_value_parser lineno interpolation at ivy_to_cpp.py:2865-2868:
// when the param default's AST node carries a non-zero source line, the
// generated error message starts with "line <N>: " ahead of the
// "parameter ..." / "syntax error in parameter value ..." text.
func TestValueParserPrefixesLinenoOnError(t *testing.T) {
	// `parameter pick : color = red` lives on line 3 of the source; the
	// generator should propagate that lineno into the emit_value_parser
	// error strings.
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
parameter pick : color = red
action step = {}
export step
`)
	out, err := Generate(mod, Config{Target: "repl", ClassName: "lrun"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, want := range []string{
		// The default-application block uses the lineno prefix.
		`std::cerr << "test.ivy: line 3: parameter pick out of bounds\n";`,
		`std::cerr << "test.ivy: line 3: syntax error in parameter value pick\n";`,
	} {
		if !strings.Contains(out.Impl, want) {
			t.Fatalf("missing lineno-prefixed message %q in impl:\n%s", want, out.Impl)
		}
	}
	// The argv `param=value` branch passes no lineno (Python omits the
	// lineno= keyword at ivy_to_cpp.py:2733), so the second occurrence
	// inside `if (param == "pick") { ... }` must have an empty prefix.
	idx := strings.Index(out.Impl, `if (param == "pick") {`)
	if idx < 0 {
		t.Fatalf("missing argv dispatch for `pick`:\n%s", out.Impl)
	}
	tail := out.Impl[idx:]
	end := strings.Index(tail, `continue;`)
	if end < 0 {
		t.Fatalf("missing terminator for argv pick branch:\n%s", tail)
	}
	argvBlock := tail[:end]
	for _, want := range []string{
		`std::cerr << "parameter pick out of bounds\n";`,
		`std::cerr << "syntax error in parameter value pick\n";`,
	} {
		if !strings.Contains(argvBlock, want) {
			t.Fatalf("argv branch missing unprefixed message %q; block:\n%s", want, argvBlock)
		}
	}
	if strings.Contains(argvBlock, "line ") {
		t.Fatalf("argv branch should not include a lineno prefix; block:\n%s", argvBlock)
	}
	compileGeneratedCPP(t, out)
}

// TestRandomizeRangeSortUsesPrimitivePath confirms that a top-level
// state symbol with a range-sort range (typedef'd int with [lo,hi]
// bounds) routes through the primitive g.randomize() path — not
// __randomize<T>. Range sorts are not in destructors/native/cpptype,
// so Python's emit_randomize (ivy_to_cpp.py:996-1004) hits the else
// branch and emits randomize(name, "<rng>").
func TestRandomizeRangeSortUsesPrimitivePath(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type rng = {2..7}
individual r : rng
action step(x:rng) = {
    r := x
}
export step
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "ranger"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Header + out.Impl
	for _, want := range []string{
		`g.mk_int("rng", 2, 7);`,
		`randomize("__fml:x", "rng");`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in range-sort gen output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	if strings.Contains(text, `__randomize<ranger::rng>(*this, apply("__fml:x")`) {
		t.Fatalf("range sort must not route through __randomize<T>:\n%s", text)
	}
	assertNoUnsupportedCPP(t, out)
}

// TestRandomizeNativeRangeUsesRandomizeTemplate confirms that a state
// symbol whose function range is a native-interpreted sort routes
// through the per-type __randomize<classname::T> template, matching
// Python emit_randomize at ivy_to_cpp.py:996.
func TestRandomizeNativeRangeUsesRandomizeTemplate(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type idx = {0..3}
type vec
interpret vec -> <<< primitive std::vector<`+"`idx`"+`> >>>
function owner(I:idx) : vec
action step(i:idx, v:vec) = {
    owner(i) := v
}
export step
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "natg"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Header + out.Impl
	// The action formal `v` has range `vec` (native). Python emits
	// `__randomize<natg::vec>(*this, apply("__fml:v"), "vec");` for
	// this input. The generic __randomize<T> template (z3.go:62)
	// services this at runtime via g.randomize(expr, range).
	for _, want := range []string{
		`__randomize<natg::vec>(*this, apply("__fml:v"), "vec");`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in native-range gen output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	if strings.Contains(text, `randomize("__fml:v", "vec")`) {
		t.Fatalf("native-range input must not route through primitive randomize():\n%s", text)
	}
	assertNoUnsupportedCPP(t, out)
}

// TestRandomizeUninterpretedRangeUsesDefaultBounds locks in the Go
// divergence from Python at emit_randomize: Python (ivy_to_cpp.py:1001)
// raises IvyError when the symbol's range is uninterpreted; Go falls
// back to g.randomize() against the runtime's default mk_sort [0,4]
// bounds. The divergence is intentional (solver_emit.go:295-299) so
// gen-target fixtures over uninterpreted sorts work end-to-end.
func TestRandomizeUninterpretedRangeUsesDefaultBounds(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type node
individual root : node
action pick(n:node) = {
    root := n
}
export pick
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "ung"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Header + out.Impl
	for _, want := range []string{
		`g.mk_sort("node");`,
		`randomize("__fml:n", "node");`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in uninterpreted-range gen output:\nheader:\n%s\nimpl:\n%s", want, out.Header, out.Impl)
		}
	}
	if strings.Contains(text, `__randomize<ung::node>`) {
		t.Fatalf("uninterpreted range must not pick the __randomize<T> dispatch:\n%s", text)
	}
	assertNoUnsupportedCPP(t, out)
}

// TestRandomizeSeedPlumbingPrecedesRandomCalls confirms the
// `seed=<int>` argv plumbing matches Python emit_main: a default
// `int seed = 1;` declaration, the `else if (param == "seed")` argv
// branch, and `srand(seed);` after argv parsing — in that order,
// before randomize() is ever called by the generator. Cross-binary
// reproducibility between Python-emitted and Go-emitted C++ under the
// same `seed=N` depends on this exact shape.
func TestRandomizeSeedPlumbingPrecedesRandomCalls(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
type color = {red, green}
individual c : color
action pick(x:color) = {
    c := x
}
export pick
`)
	out, err := Generate(mod, Config{Target: "gen", ClassName: "seeded", EmitMain: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	mainIdx := strings.LastIndex(out.Impl, "int main(int argc, char **argv)")
	if mainIdx < 0 {
		t.Fatalf("missing generated main() in gen output:\n%s", out.Impl)
	}
	mainBody := out.Impl[mainIdx:]
	declIdx := strings.Index(mainBody, "int seed = 1;")
	keyIdx := strings.Index(mainBody, `else if (param == "seed") { seed = atoi(value.c_str()); }`)
	srandIdx := strings.Index(mainBody, "srand(seed);")
	if declIdx < 0 || keyIdx < 0 || srandIdx < 0 {
		t.Fatalf("seed plumbing pieces missing (decl=%d key=%d srand=%d):\n%s", declIdx, keyIdx, srandIdx, mainBody)
	}
	if !(declIdx < keyIdx && keyIdx < srandIdx) {
		t.Fatalf("seed plumbing out of order (decl=%d key=%d srand=%d); want decl < key < srand", declIdx, keyIdx, srandIdx)
	}
	// The first ivy2cpp_randomize / __randomize call in the program
	// flow must come after srand(seed); otherwise the seed wouldn't
	// affect it.
	randCallIdx := -1
	for _, marker := range []string{"ivy2cpp_randomize(", "__randomize<", "g.randomize("} {
		if i := strings.Index(out.Impl[srandIdx:], marker); i >= 0 {
			if randCallIdx < 0 || i < randCallIdx {
				randCallIdx = i
			}
		}
	}
	if randCallIdx < 0 {
		t.Fatalf("no randomize call found anywhere after srand(seed) in:\n%s", out.Impl[srandIdx:])
	}
	assertNoUnsupportedCPP(t, out)
}

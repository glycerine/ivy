package goivy

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/xtracer"
)

func captureParserTrace(t *testing.T, fn func()) string {
	t.Helper()

	oldSuppressed := xtracer.Suppressed
	xtracer.Suppressed = false
	defer func() {
		xtracer.Suppressed = oldSuppressed
	}()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	var buf bytes.Buffer
	done := make(chan error, 1)
	go func() {
		_, copyErr := io.Copy(&buf, r)
		done <- copyErr
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("closing trace writer: %v", err)
	}
	os.Stdout = oldStdout
	if err := <-done; err != nil {
		t.Fatalf("reading trace output: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("closing trace reader: %v", err)
	}
	return buf.String()
}

func requireParserTraceEnabled(t *testing.T) {
	t.Helper()
	if !xtracer.Enabled {
		t.Skip("requires xtrace; disabled by xtrace_off build tag")
	}
}

func TestParseV16AtermCallReducesCalleeBeforeArguments(t *testing.T) {
	requireParserTraceEnabled(t)

	src := `#lang ivy1.6
action foo = {
    store(K) := 0
}`

	var parseErr error
	trace := captureParserTrace(t, func() {
		_, parseErr = Parse(src, Version{1, 6}, WithFilename("v16_aterm.ivy"))
	})
	if parseErr != nil {
		t.Fatalf("Parse ivy1.6: %v", parseErr)
	}

	callee := strings.Index(trace, "parser.p_aterm_symbol ENTER (aterm)")
	arg := strings.Index(trace, "parser.p_var_variable ENTER (var)")
	if callee < 0 {
		t.Fatalf("missing v1.6 callee aterm reduction in trace:\n%s", trace)
	}
	if arg < 0 {
		t.Fatalf("missing argument variable reduction in trace:\n%s", trace)
	}
	if callee > arg {
		t.Fatalf("callee should reduce before argument for Ivy 1.6 call terms; trace:\n%s", trace)
	}
}

func TestParseV16Repstore1(t *testing.T) {
	path := filepath.Join("..", "ivy-lang-examples", "doc", "examples", "MSV", "repstore1.ivy")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	oldSuppressed := xtracer.Suppressed
	xtracer.Suppressed = true
	defer func() {
		xtracer.Suppressed = oldSuppressed
	}()

	result, err := Parse(string(src), Version{1, 6}, WithFilename(path))
	if err != nil {
		t.Fatalf("Parse repstore1 ivy1.6: %v", err)
	}
	if len(result.Decls) == 0 {
		t.Fatal("expected declarations from repstore1")
	}
}

func TestParseV16CorpusSmoke(t *testing.T) {
	paths := []string{
		filepath.Join("..", "ivy-lang-examples", "test", "schema1.ivy"),
		filepath.Join("..", "ivy-lang-examples", "test", "asgn_call1.ivy"),
		filepath.Join("..", "ivy-lang-examples", "test", "old1.ivy"),
		filepath.Join("..", "ivy-lang-examples", "test", "somemin.ivy"),
		filepath.Join("..", "ivy-lang-examples", "test", "theory1.ivy"),
		filepath.Join("..", "ivy-lang-examples", "doc", "examples", "paraminit.ivy"),
		filepath.Join("..", "ivy-lang-examples", "doc", "examples", "MSV", "hello1.ivy"),
	}

	oldSuppressed := xtracer.Suppressed
	xtracer.Suppressed = true
	defer func() {
		xtracer.Suppressed = oldSuppressed
	}()

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading %s: %v", path, err)
			}
			result, err := Parse(string(src), Version{1, 6}, WithFilename(path))
			if err != nil {
				t.Fatalf("Parse %s as ivy1.6: %v", path, err)
			}
			if len(result.Decls) == 0 {
				t.Fatalf("expected declarations from %s", path)
			}
		})
	}
}

func TestListIsolatesV16ParameterizedObjectExtract(t *testing.T) {
	path := filepath.Join("..", "ivy-lang-examples", "doc", "examples", "interference2.ivy")

	isolates, err := ListIsolates(path)
	if err != nil {
		t.Fatalf("ListIsolates(%s): %v", path, err)
	}
	if len(isolates) != 1 || isolates[0] != "iso_foo" {
		t.Fatalf("ListIsolates(%s) = %v, want [iso_foo]", path, isolates)
	}
}

func TestReadModuleFromStringInheritsV16Version(t *testing.T) {
	requireParserTraceEnabled(t)

	cfg := NewConfig()
	SetStringVersionOn(cfg.IuCfg, "1.6")
	src := `#lang ivy
action foo = {
    store(K) := 0
}`

	var parseErr error
	trace := captureParserTrace(t, func() {
		_, parseErr = ReadModuleFromString(src, cfg)
	})
	if parseErr != nil {
		t.Fatalf("ReadModuleFromString ivy1.6: %v", parseErr)
	}

	callee := strings.Index(trace, "parser.p_aterm_symbol ENTER (aterm)")
	arg := strings.Index(trace, "parser.p_var_variable ENTER (var)")
	if callee < 0 || arg < 0 || callee > arg {
		t.Fatalf("expected no-header string parse to inherit Ivy 1.6 aterm order; trace:\n%s", trace)
	}
}

func TestARGSetupInitKeepsCompiledLabeledFormula(t *testing.T) {
	requireParserTraceEnabled(t)

	cfg := NewConfig()
	src := `#lang ivy1.6
individual bit : bool
init bit`

	var compileErr error
	trace := captureParserTrace(t, func() {
		pr, err := ReadModuleFromString(src, cfg)
		if err != nil {
			compileErr = err
			return
		}
		mod := New()
		mod.Cfg = cfg
		mod.Sig = NewSigOn(cfg.IuCfg)
		compileErr = IvyCompile(pr.Decls, mod, false)
	})
	if compileErr != nil {
		t.Fatalf("compile init source: %v", compileErr)
	}

	compiled := strings.Index(trace, "compiler.ARGSetup.init compiled sort=")
	clauses := strings.Index(trace, "compiler.ARGSetup.init clauses fmlas=")
	if compiled < 0 || clauses < 0 || compiled > clauses {
		t.Fatalf("missing init trace markers:\n%s", trace)
	}
	if between := trace[compiled:clauses]; strings.Contains(between, "ast.LF.__init__") {
		t.Fatalf("init allocated a wrapper LabeledFormula after compile:\n%s", between)
	}
}

func TestParseV16ConjunctionReducesBeforeArrow(t *testing.T) {
	requireParserTraceEnabled(t)

	src := `#lang ivy1.6
type t
function n : t
function p(X:t) : bool
property forall B. (B >= n & p(B)-> p(B))`

	var parseErr error
	trace := captureParserTrace(t, func() {
		_, parseErr = Parse(src, Version{1, 6}, WithFilename("v16_arrow_precedence.ivy"))
	})
	if parseErr != nil {
		t.Fatalf("Parse ivy1.6: %v", parseErr)
	}

	andReduce := strings.Index(trace, "parser.p_fmla_fmla_and_fmla ENTER (fmla)")
	arrowReduce := strings.Index(trace, "parser.p_fmla_fmla_arrow_fmla ENTER (fmla)")
	if andReduce < 0 {
		t.Fatalf("missing conjunction reduction in trace:\n%s", trace)
	}
	if arrowReduce < 0 {
		t.Fatalf("missing arrow reduction in trace:\n%s", trace)
	}
	if andReduce > arrowReduce {
		t.Fatalf("Ivy 1.6 should reduce conjunction before arrow; trace:\n%s", trace)
	}
}

func TestParseV16ConjunctionPreservesPythonNesting(t *testing.T) {
	src := `#lang ivy1.6
type t
function s(X:t) : bool
function m(X:t,Y:t) : bool
conjecture s(K) & m(K,L) & L ~= K -> s(L)`

	result, err := Parse(src, Version{1, 6}, WithFilename("v16_and_nesting.ivy"))
	if err != nil {
		t.Fatalf("Parse ivy1.6 conjunction: %v", err)
	}

	impl, ok := v16ConjectureFormula(t, result).(*Implies)
	if !ok {
		t.Fatalf("conjecture formula = %T, want *Implies", v16ConjectureFormula(t, result))
	}
	outer, ok := impl.T1.(*And)
	if !ok {
		t.Fatalf("implication antecedent = %T, want *And", impl.T1)
	}
	if len(outer.Terms) != 2 {
		t.Fatalf("outer And has %d terms, want 2", len(outer.Terms))
	}
	inner, ok := outer.Terms[0].(*And)
	if !ok {
		t.Fatalf("outer And first term = %T, want nested *And", outer.Terms[0])
	}
	if len(inner.Terms) != 2 {
		t.Fatalf("inner And has %d terms, want 2", len(inner.Terms))
	}
}

func TestParseV16DisjunctionPreservesPythonNesting(t *testing.T) {
	src := `#lang ivy1.6
type t
function s(X:t) : bool
function m(X:t,Y:t) : bool
conjecture s(K) | m(K,L) | L = K`

	result, err := Parse(src, Version{1, 6}, WithFilename("v16_or_nesting.ivy"))
	if err != nil {
		t.Fatalf("Parse ivy1.6 disjunction: %v", err)
	}

	outer, ok := v16ConjectureFormula(t, result).(*Or)
	if !ok {
		t.Fatalf("conjecture formula = %T, want *Or", v16ConjectureFormula(t, result))
	}
	if len(outer.Terms) != 2 {
		t.Fatalf("outer Or has %d terms, want 2", len(outer.Terms))
	}
	inner, ok := outer.Terms[0].(*Or)
	if !ok {
		t.Fatalf("outer Or first term = %T, want nested *Or", outer.Terms[0])
	}
	if len(inner.Terms) != 2 {
		t.Fatalf("inner Or has %d terms, want 2", len(inner.Terms))
	}
}

func v16ConjectureFormula(t *testing.T, result *ParseResult) Node {
	t.Helper()
	for _, decl := range result.Decls {
		conj, ok := decl.(*ConjectureDecl)
		if !ok {
			continue
		}
		if len(conj.DeclArgs) != 1 {
			t.Fatalf("conjecture has %d args, want 1", len(conj.DeclArgs))
		}
		lf, ok := conj.DeclArgs[0].(*LabeledFormula)
		if !ok {
			t.Fatalf("conjecture arg = %T, want *LabeledFormula", conj.DeclArgs[0])
		}
		return lf.Formula
	}
	t.Fatal("missing conjecture declaration")
	return nil
}

func TestParseV16NativeQuoteTypeUsesNativeCode(t *testing.T) {
	src := `#lang ivy1.6
type t
interpret t -> <<< std::vector<` + "`t`" + `> >>>`

	result, err := Parse(src, Version{1, 6}, WithFilename("v16_nativequote.ivy"))
	if err != nil {
		t.Fatalf("Parse ivy1.6 native quote: %v", err)
	}

	var nativeType *NativeType
	var walk func(Node)
	walk = func(n Node) {
		if n == nil || nativeType != nil {
			return
		}
		if nt, ok := n.(*NativeType); ok {
			nativeType = nt
			return
		}
		for _, arg := range n.Args() {
			walk(arg)
		}
	}
	for _, decl := range result.Decls {
		walk(decl)
	}
	if nativeType == nil {
		t.Fatal("expected parsed interpret declaration to contain NativeType")
	}
	if len(nativeType.Elems) == 0 {
		t.Fatal("expected NativeType to contain native code element")
	}
	if _, ok := nativeType.Elems[0].(*NativeCode); !ok {
		t.Fatalf("NativeType first element = %T, want *NativeCode", nativeType.Elems[0])
	}
}

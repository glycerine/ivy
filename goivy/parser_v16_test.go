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

func TestParseV16AtermCallReducesCalleeBeforeArguments(t *testing.T) {
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

func TestReadModuleFromStringInheritsV16Version(t *testing.T) {
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

func TestParseV16ConjunctionReducesBeforeArrow(t *testing.T) {
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

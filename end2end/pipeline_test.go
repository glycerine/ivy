// Package end2end provides integration tests for the full Ivy verification
// pipeline: parse → compile → isolate → check → Z3.
//
// These tests use external .ivy files from data/ and exercise the end-to-end
// flow, cutting across all packages.
package end2end

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/glycerine/goivy/compiler"
	il "github.com/glycerine/goivy/ivylogic"
	"github.com/glycerine/goivy/lalr_full"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// dataDir returns the path to the data directory.
func dataDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "data")
}

// compileIvySource parses and compiles an Ivy source string into a module.
func compileIvySource(t *testing.T, src string) *module.Module {
	t.Helper()
	version := lexer.Version{1, 7}
	// Try LALR parser first, fall back to hand-rolled for debugging
	result, err := lalr_full.Parse(src, version)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	mod := module.New()
	mod.Sig = il.NewSig()
	err = compiler.IvyCompile(result.Decls, mod, true)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	return mod
}



// compileIvyFile parses and compiles an Ivy file into a module.
func compileIvyFile(t *testing.T, filename string) *module.Module {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dataDir(), filename))
	if err != nil {
		t.Fatalf("read file %v: %v", filename, err)
	}
	return compileIvySource(t, string(data))
}

// -----------------------------------------------------------------------
// Category A: Parse + Compile (no Z3)
// -----------------------------------------------------------------------

func TestParseCompile_RelationSort(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7

type t

relation r(X:t)

after init {
    r(X) := false
}

action step(x:t) = {
    r(x) := true
}

export step

conjecture r(X) | ~r(X)
`)
	// Verify the module compiled
	if mod == nil {
		t.Fatal("module is nil")
	}
	if mod.Sig == nil {
		t.Fatal("sig is nil")
	}

	// Check that sort "t" exists
	_, ok := mod.Sig.Sorts["t"]
	if !ok {
		t.Error("sort 't' not found in signature")
	}

	// Check that "r" is a relation with Boolean range
	entry, err := mod.Sig.FindSymbol("r", false)
	if err != nil {
		t.Fatalf("symbol 'r' not found: %v", err)
	}
	fs, ok := entry.CSort.(*logic.FunctionSort)
	if !ok {
		t.Fatalf("r sort should be FunctionSort, got %T", entry.CSort)
	}
	if !logic.SortEqual(fs.Range(), logic.Boolean) {
		t.Errorf("r range should be Boolean, got %s", fs.Range())
	}

	// Check we have at least one conjecture
	if len(mod.LabeledConjs) == 0 {
		t.Error("expected at least one conjecture")
	}
}

func TestParseCompile_EnumTypes(t *testing.T) {
	mod := compileIvyFile(t, "enum_types.ivy")
	if mod == nil {
		t.Fatal("module is nil")
	}

	// Check that "color" is an EnumeratedSort
	sort, ok := mod.Sig.Sorts["color"]
	if !ok {
		t.Fatal("sort 'color' not found")
	}
	es, ok := sort.(*logic.EnumeratedSort)
	if !ok {
		t.Fatalf("expected EnumeratedSort for color, got %T", sort)
	}
	if len(es.Extension) != 3 {
		t.Errorf("expected 3 color values, got %d", len(es.Extension))
	}
	expected := map[string]bool{"red": true, "green": true, "blue": true}
	for _, e := range es.Extension {
		if !expected[e] {
			t.Errorf("unexpected enum value: %s", e)
		}
	}
}

func TestParseCompile_ClientServer(t *testing.T) {
	mod := compileIvyFile(t, "client_server.ivy")
	if mod == nil {
		t.Fatal("module is nil")
	}

	// Check sorts
	if _, ok := mod.Sig.Sorts["client"]; !ok {
		t.Error("sort 'client' not found")
	}
	if _, ok := mod.Sig.Sorts["server"]; !ok {
		t.Error("sort 'server' not found")
	}

	// Check relations
	for _, relName := range []string{"link", "semaphore"} {
		entry, err := mod.Sig.FindSymbol(relName, false)
		if err != nil {
			t.Errorf("relation '%s' not found: %v", relName, err)
			continue
		}
		fs, ok := entry.CSort.(*logic.FunctionSort)
		if !ok {
			t.Errorf("%s sort should be FunctionSort, got %T", relName, entry.CSort)
			continue
		}
		if !logic.SortEqual(fs.Range(), logic.Boolean) {
			t.Errorf("%s range should be Boolean, got %s", relName, fs.Range())
		}
	}

	// Check link arity = 2 (client, server)
	linkEntry, _ := mod.Sig.FindSymbol("link", false)
	linkFS := linkEntry.CSort.(*logic.FunctionSort)
	if linkFS.Arity() != 2 {
		t.Errorf("link arity should be 2, got %d", linkFS.Arity())
	}

	// Check semaphore arity = 1 (server)
	semEntry, _ := mod.Sig.FindSymbol("semaphore", false)
	semFS := semEntry.CSort.(*logic.FunctionSort)
	if semFS.Arity() != 1 {
		t.Errorf("semaphore arity should be 1, got %d", semFS.Arity())
	}

	// Check actions exist (after isolate processing, actions have "ext:" prefix)
	if _, ok := mod.Actions["ext:connect"]; !ok {
		t.Error("action 'ext:connect' not found")
	}
	if _, ok := mod.Actions["ext:disconnect"]; !ok {
		t.Error("action 'ext:disconnect' not found")
	}

	// Check conjectures
	if len(mod.LabeledConjs) < 2 {
		t.Errorf("expected at least 2 conjectures, got %d", len(mod.LabeledConjs))
	}
}

func TestParseCompile_MultipleConjectures(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7

type t

relation p(X:t)
relation q(X:t)

after init {
    p(X) := false;
    q(X) := false
}

action set_p(x:t) = {
    p(x) := true
}

action set_q(x:t) = {
    q(x) := true
}

export set_p
export set_q

conjecture p(X) | ~p(X)
conjecture q(X) | ~q(X)
conjecture ~p(X) | ~q(X) | p(X) & q(X)
`)
	if len(mod.LabeledConjs) != 3 {
		t.Errorf("expected 3 conjectures, got %d", len(mod.LabeledConjs))
	}
	for i, lc := range mod.LabeledConjs {
		if lc.Formula == nil {
			t.Errorf("conjecture %d has nil formula", i)
		}
	}
}

// TestEnum_InitPostState_BooleanSort verifies that compiling
// "after init { c := red }" for an enum type produces init clauses
// with Boolean-sorted formulas (equalities), not bare color-sorted symbols.
//
// Bug: the init post-state was And(c, red) — a conjunction of two
// color-sorted symbols — instead of the correct Eq(c, red).
// This caused a Z3 sort mismatch: "supplied sort is color" where Bool
// was expected as an argument to And.
func TestEnum_InitPostState_BooleanSort(t *testing.T) {
	src := `
type color = {red, green, blue}
individual c : color
after init {
    c := red
}
action set_color(x:color) = {
    c := x
}
export set_color
conjecture c = red | c = green | c = blue
`
	mod := compileIvySource(t, src)

	// The init condition clauses must have Boolean-sorted formulas.
	// If the assign action "c := red" is compiled correctly, the init
	// condition should contain Eq(c, red) (Boolean), not bare And(c, red).
	if mod.InitCond == nil {
		t.Fatal("expected non-nil InitCond after 'after init { c := red }'")
	}
	// The init condition clauses must have Boolean-sorted formulas.
	if mod.InitCond != nil {
		for i, fmla := range mod.InitCond.Fmlas {
			sort := fmla.NodeSort()
			sortName := il.SortName(sort)
			if sortName != "bool" {
				t.Errorf("InitCond.Fmlas[%d] has sort %q (%T = %v), want bool",
					i, sortName, fmla, fmla)
			}
		}
	}

	// Verify the full pipeline doesn't panic with a Z3 sort mismatch.
	// The bug: init post-state was And(c, red) — bare color-sorted symbols —
	// instead of Eq(c, red). This caused Z3 to reject the And() call.
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("verification panicked (Z3 sort mismatch bug): %v", r)
			}
		}()
		if !verifyInitInvariant(t, mod) {
			t.Error("init should establish enum exhaustiveness invariant")
		}
	}()
}

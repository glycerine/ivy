package ivy2go

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

// --- OPEN 052: quantified-LHS assignment tests ----------------------

func TestEmitAssign_QuantifiedLHSEmitsTwoPhaseLoop(t *testing.T) {
	// `link(X, Y) := false` over node = {0..3} should two-phase
	// loop over X and Y.
	g := newExprGen(t, `
type node = {0..3}
relation link(N1: node, N2: node)
`)
	nodeSort := g.Mod.Sig.Sorts.Get("node")
	linkSort, _ := goivy.NewFunctionSort(nodeSort, nodeSort, goivy.Boolean)
	link := goivy.NewConst("link", linkSort)
	x, _ := goivy.NewVariable("X", nodeSort)
	y, _ := goivy.NewVariable("Y", nodeSort)
	lhs, _ := goivy.NewApply(link, x, y)
	rhs := &goivy.Const{Name: "false", CSort: goivy.Boolean}
	a := &goivy.LogicAssignAction{LHS: lhs, RHS: rhs}

	w := newGoWriter(NewGoText())
	g.emitAction(&w, a)
	got := w.String()

	// Phase 1: a temp gets the RHS via nested loops.
	if !strings.Contains(got, "__ivy_tmp") {
		t.Errorf("two-phase should declare __ivy_tmp, got:\n%s", got)
	}
	// Two loops over Node.
	if strings.Count(got, "X := Node(0)") < 2 || strings.Count(got, "Y := Node(0)") < 2 {
		t.Errorf("expected two passes of X and Y loops (one per phase), got:\n%s", got)
	}
}

func TestEmitAssign_UnboundedQuantifiedLHSWrapsInThunk(t *testing.T) {
	// An uninterpreted sort with no known cardinality has no
	// derivable loop bounds — should route through emitAssignLarge
	// and emit a thunk construction.
	g := newExprGen(t, `
type node
relation slot(N: node)
`)
	nodeSort, _ := g.Mod.Sig.Sorts.Get2("node")
	slotSort, _ := goivy.NewFunctionSort(nodeSort, goivy.Boolean)
	slot := goivy.NewConst("slot", slotSort)
	x, _ := goivy.NewVariable("X", nodeSort)
	lhs, _ := goivy.NewApply(slot, x)
	rhs := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	a := &goivy.LogicAssignAction{LHS: lhs, RHS: rhs}

	w := newGoWriter(NewGoText())
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "OPEN 061:") {
		t.Errorf("thunk-fallback path should mark OPEN 061, got:\n%s", got)
	}
	if !strings.Contains(got, "__thunk :=") {
		t.Errorf("thunk-fallback should construct a thunk, got:\n%s", got)
	}
}

func TestEmitAssign_QuantifiedLHSSingleVar(t *testing.T) {
	g := newExprGen(t, `
type idx = {0..7}
relation slot(I: idx)
`)
	idxSort := g.Mod.Sig.Sorts.Get("idx")
	slotSort, _ := goivy.NewFunctionSort(idxSort, goivy.Boolean)
	slot := goivy.NewConst("slot", slotSort)
	x, _ := goivy.NewVariable("X", idxSort)
	lhs, _ := goivy.NewApply(slot, x)
	rhs := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	a := &goivy.LogicAssignAction{LHS: lhs, RHS: rhs}

	w := newGoWriter(NewGoText())
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "X := Idx(0)") {
		t.Errorf("single-var loop missing, got:\n%s", got)
	}
}

// --- OPEN 058: trace-LHS emission tests -----------------------------

func TestEmitTrace_FmtCallEmittedWhenTraceOn(t *testing.T) {
	g := newExprGen(t, `relation flag`)
	g.Config.Trace = true
	flag := goivy.NewConst("flag", goivy.Boolean)
	rhs := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	a := &goivy.LogicAssignAction{LHS: flag, RHS: rhs}

	w := newGoWriter(NewGoText())
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "ivyTraceOut") {
		t.Errorf("trace mode should emit ivyTraceOut call, got:\n%s", got)
	}
	if !strings.Contains(got, "write(flag") {
		t.Errorf("trace should describe LHS write, got:\n%s", got)
	}
}

func TestEmitTrace_NamespacedNameSuppressed(t *testing.T) {
	g := newExprGen(t, "")
	g.Config.Trace = true
	loc := &goivy.Const{Name: "loc:tmp", CSort: goivy.Boolean}
	rhs := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	a := &goivy.LogicAssignAction{LHS: loc, RHS: rhs}

	w := newGoWriter(NewGoText())
	g.emitAction(&w, a)
	got := w.String()
	if strings.Contains(got, "ivyTraceOut") {
		t.Errorf("loc:-prefixed symbol should NOT emit trace, got:\n%s", got)
	}
}

func TestEmitTrace_OffByDefault(t *testing.T) {
	g := newExprGen(t, `relation flag`)
	flag := goivy.NewConst("flag", goivy.Boolean)
	rhs := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	a := &goivy.LogicAssignAction{LHS: flag, RHS: rhs}

	w := newGoWriter(NewGoText())
	g.emitAction(&w, a)
	got := w.String()
	if strings.Contains(got, "ivyTraceOut") {
		t.Errorf("Config.Trace=false should not emit trace, got:\n%s", got)
	}
}

func TestEmitTrace_RuntimeIvyTraceOutDeclared(t *testing.T) {
	mod := compileIvySource(t, `relation flag`)
	out, err := Generate(mod, Config{Target: "impl", PackageName: "p", Trace: true})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	text := out.Files["runtime.go"]
	if !strings.Contains(text, "ivyTraceOut io.Writer") {
		t.Errorf("Trace=true should declare ivyTraceOut, got:\n%s", text)
	}
}

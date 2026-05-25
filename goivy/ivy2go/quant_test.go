package ivy2go

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

// --- OPEN 050: quantifier emission tests ----------------------------

func TestEmitQuant_ForallBoolEmitsIIFEWithEarlyExit(t *testing.T) {
	g := newExprGen(t, "")
	x, err := goivy.NewVariable("X", goivy.Boolean)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	body := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	got, err := g.emitQuant([]*goivy.LogicVariable{x}, body, true)
	if err != nil {
		t.Fatalf("emitQuant forall: %v", err)
	}
	if !strings.Contains(got, "func() bool {") {
		t.Errorf("forall should emit IIFE, got:\n%s", got)
	}
	if !strings.Contains(got, "for _, X := range [2]bool{false, true}") {
		t.Errorf("forall over bool should iterate [false,true], got:\n%s", got)
	}
	if !strings.Contains(got, "return false") {
		t.Errorf("forall should return false on violation, got:\n%s", got)
	}
}

func TestEmitQuant_ExistsBoolEmitsIIFEWithEarlyTrue(t *testing.T) {
	g := newExprGen(t, "")
	x, err := goivy.NewVariable("X", goivy.Boolean)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	body := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	got, err := g.emitQuant([]*goivy.LogicVariable{x}, body, false)
	if err != nil {
		t.Fatalf("emitQuant exists: %v", err)
	}
	if !strings.Contains(got, "return true") {
		t.Errorf("exists should return true on hit, got:\n%s", got)
	}
}

func TestEmitQuant_EnumSortIteratesCardinality(t *testing.T) {
	g := newExprGen(t, "type color = {red, green, blue}")
	x, err := goivy.NewVariable("X", g.Mod.Sig.Sorts.Get("color"))
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	body := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	got, err := g.emitQuant([]*goivy.LogicVariable{x}, body, true)
	if err != nil {
		t.Fatalf("emitQuant: %v", err)
	}
	if !strings.Contains(got, "__i < 3") {
		t.Errorf("forall over Color (card 3) should iterate 0..3, got:\n%s", got)
	}
	if !strings.Contains(got, "X := Color(__i)") {
		t.Errorf("loop should cast to Color named type, got:\n%s", got)
	}
}

func TestEmitQuant_RangeSortUsesBoundsLoop(t *testing.T) {
	g := newExprGen(t, "type idx = {0..7}")
	x, err := goivy.NewVariable("X", g.Mod.Sig.Sorts.Get("idx"))
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	body := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	got, err := g.emitQuant([]*goivy.LogicVariable{x}, body, true)
	if err != nil {
		t.Fatalf("emitQuant: %v", err)
	}
	if !strings.Contains(got, "X := Idx(0)") || !strings.Contains(got, "Idx(7)") {
		t.Errorf("range loop should use bounds 0..7, got:\n%s", got)
	}
}

func TestEmitQuant_MultipleVariablesNestLoops(t *testing.T) {
	g := newExprGen(t, "")
	x, _ := goivy.NewVariable("X", goivy.Boolean)
	y, _ := goivy.NewVariable("Y", goivy.Boolean)
	body := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	got, err := g.emitQuant([]*goivy.LogicVariable{x, y}, body, true)
	if err != nil {
		t.Fatalf("emitQuant: %v", err)
	}
	if strings.Count(got, "for _,") != 2 {
		t.Errorf("two-variable quant should emit two loops, got:\n%s", got)
	}
}

func TestEmitQuant_EmptyVarsFallsThroughToBody(t *testing.T) {
	g := newExprGen(t, "")
	body := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	got, err := g.emitQuant(nil, body, true)
	if err != nil {
		t.Fatalf("emitQuant: %v", err)
	}
	if got != "true" {
		t.Errorf("empty-vars quant should just emit body, got %q", got)
	}
}

// --- OPEN 051: if-some lowering tests --------------------------------

func TestEmitIfSome_EmitsScanAndFoundFlag(t *testing.T) {
	g := newExprGen(t, "")
	param := &goivy.Const{Name: "X", CSort: goivy.Boolean}
	fmla := param
	some := &goivy.SomeCondition{Params: []*goivy.Const{param}, Fmla: fmla, Kind: "some"}
	then := &goivy.LogicAssertAction{Formula: &goivy.Const{Name: "true", CSort: goivy.Boolean}}
	a := &goivy.LogicIfAction{Cond: some, ThenBody: then}
	w := newGoWriter(NewGoText())
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "__found := false") {
		t.Errorf("if-some should declare __found, got:\n%s", got)
	}
	if !strings.Contains(got, "for _, __some_X := range") {
		t.Errorf("if-some should iterate the param sort, got:\n%s", got)
	}
	if !strings.Contains(got, "if !__found && (__some_X)") {
		t.Errorf("if-some should check found-flag + cond, got:\n%s", got)
	}
}

func TestEmitIfSome_WithElseBranch(t *testing.T) {
	g := newExprGen(t, "")
	param := &goivy.Const{Name: "X", CSort: goivy.Boolean}
	some := &goivy.SomeCondition{Params: []*goivy.Const{param}, Fmla: param, Kind: "some"}
	then := &goivy.LogicAssertAction{Formula: &goivy.Const{Name: "true", CSort: goivy.Boolean}}
	els := &goivy.LogicAssumeAction{Formula: &goivy.Const{Name: "false", CSort: goivy.Boolean}}
	a := &goivy.LogicIfAction{Cond: some, ThenBody: then, ElseBody: els}
	w := newGoWriter(NewGoText())
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "if !__found {") {
		t.Errorf("if-some else branch should be gated on !__found, got:\n%s", got)
	}
	if !strings.Contains(got, "ivyAssume(false,") {
		t.Errorf("else body should be emitted, got:\n%s", got)
	}
}

func TestEmitIfSome_MinMaxScansAndTracksBest(t *testing.T) {
	g := newExprGen(t, "")
	// some X:Bool. X minimizing X (silly, but exercises the shape).
	param := &goivy.Const{Name: "X", CSort: goivy.Boolean}
	some := &goivy.SomeCondition{
		Params: []*goivy.Const{param},
		Fmla:   param,
		Kind:   "some_min",
		Index:  param,
	}
	then := &goivy.LogicAssertAction{Formula: &goivy.Const{Name: "true", CSort: goivy.Boolean}}
	a := &goivy.LogicIfAction{Cond: some, ThenBody: then}
	w := newGoWriter(NewGoText())
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "__best_idx") {
		t.Errorf("some_min should declare __best_idx, got:\n%s", got)
	}
	if !strings.Contains(got, "__cur_idx < __best_idx") {
		t.Errorf("some_min should use < for best comparison, got:\n%s", got)
	}
	if !strings.Contains(got, "if __found {") {
		t.Errorf("some_min should dispatch THEN on __found, got:\n%s", got)
	}
}

func TestEmitIfSome_MaxUsesGreater(t *testing.T) {
	g := newExprGen(t, "")
	param := &goivy.Const{Name: "X", CSort: goivy.Boolean}
	some := &goivy.SomeCondition{
		Params: []*goivy.Const{param},
		Fmla:   param,
		Kind:   "some_max",
		Index:  param,
	}
	then := &goivy.LogicAssertAction{Formula: &goivy.Const{Name: "true", CSort: goivy.Boolean}}
	a := &goivy.LogicIfAction{Cond: some, ThenBody: then}
	w := newGoWriter(NewGoText())
	g.emitAction(&w, a)
	got := w.String()
	if !strings.Contains(got, "__cur_idx > __best_idx") {
		t.Errorf("some_max should use > for best comparison, got:\n%s", got)
	}
}

func TestEmitQuant_ThroughEmitExpr(t *testing.T) {
	// Smoke: emitting a ForAll node via the top-level dispatch
	// should also produce a working IIFE.
	g := newExprGen(t, "")
	x, _ := goivy.NewVariable("X", goivy.Boolean)
	body := &goivy.Const{Name: "true", CSort: goivy.Boolean}
	forall := &goivy.ForAll{Variables: []*goivy.LogicVariable{x}, Body: body}
	got, err := g.emitExpr(forall)
	if err != nil {
		t.Fatalf("emitExpr forall: %v", err)
	}
	if !strings.Contains(got, "func() bool {") {
		t.Errorf("forall via emitExpr should produce IIFE, got:\n%s", got)
	}
}

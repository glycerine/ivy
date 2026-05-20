package goivy

import (
	"fmt"
	"strings"
	"testing"
)

// stubAction is a minimal Action implementation for canon tests.
type stubAction struct {
	name    string
	formula Expr
}

func (s *stubAction) ActionClone(args []Expr) Action { return s }
func (s *stubAction) ActionArgs() []Expr             { return []Expr{s.formula} }
func (s *stubAction) IterCalls() []string            { return nil }
func (s *stubAction) IterSubactions() []Action       { return nil }
func (s *stubAction) GetFormalParams() []*Const      { return nil }
func (s *stubAction) GetFormalReturns() []*Const     { return nil }
func (s *stubAction) SetFormalParams([]*Const)       {}
func (s *stubAction) SetFormalReturns([]*Const)      {}
func (s *stubAction) Name() string                   { return s.name }
func (s *stubAction) Decompose() [][]Action          { return nil }
func (s *stubAction) Args() []Node                   { return nil }
func (s *stubAction) Clone(args []Node) Node         { return s }
func (s *stubAction) GetLineno() Location            { return Location{} }
func (s *stubAction) SetLineno(Location)             {}
func (s *stubAction) HasLineno() bool                { return false }
func (s *stubAction) String() string                 { return s.name }
func (s *stubAction) Canon() Canonical               { return Canonical(s.Sexp()) }
func (s *stubAction) GetAstConfig() *AstConfig       { return nil }
func (s *stubAction) NodeSort() Sort                 { return ActionS }
func (s *stubAction) Children() []Expr               { return nil }
func (s *stubAction) Equal(other Expr) bool          { return s.Sexp() == other.Sexp() }
func (s *stubAction) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(%s formula:%s)", s.name, string(s.formula.Sexp())))
}

func mkSortRefinement(old, new_ Sort) map[NodeKey]*SortRefinement {
	sr := &SortRefinement{Old: old, New: new_}
	return map[NodeKey]*SortRefinement{SortKey(old): sr}
}

func TestResortSortUnrefined(t *testing.T) {
	s := &UninterpretedSort{Name: "t"}
	rn := map[NodeKey]*SortRefinement{}
	result := ResortSort(s, rn)
	if !SortEqual(result, s) {
		t.Error("unrefinement should not change sort")
	}
}

func TestResortSortRefined(t *testing.T) {
	old := &UninterpretedSort{Name: "abstract_t"}
	new_ := &UninterpretedSort{Name: "concrete_t"}
	rn := mkSortRefinement(old, new_)

	result := ResortSort(old, rn)
	if !SortEqual(result, new_) {
		t.Errorf("expected concrete_t, got %s", result)
	}
}

func TestResortSortFunctionSort(t *testing.T) {
	old := &UninterpretedSort{Name: "abstract_t"}
	new_ := &UninterpretedSort{Name: "concrete_t"}
	other := &UninterpretedSort{Name: "u"}
	fs, _ := NewFunctionSort(old, other)
	rn := mkSortRefinement(old, new_)

	result := ResortSort(fs, rn)
	rfs, ok := result.(*LogicFunctionSort)
	if !ok {
		t.Fatal("expected FunctionSort")
	}
	if !SortEqual(rfs.Domain()[0], new_) {
		t.Errorf("expected domain[0] = concrete_t, got %s", rfs.Domain()[0])
	}
	if !SortEqual(rfs.Range(), other) {
		t.Errorf("expected range = u, got %s", rfs.Range())
	}
}

func TestResortAST(t *testing.T) {
	old := &UninterpretedSort{Name: "abstract_t"}
	new_ := &UninterpretedSort{Name: "concrete_t"}
	rn := mkSortRefinement(old, new_)

	// Test variable resort
	v, _ := NewVariable("X", old)
	result := ResortAST(v, rn)
	rv, ok := result.(*LogicVariable)
	if !ok {
		t.Fatal("expected Var")
	}
	if !SortEqual(rv.VSort, new_) {
		t.Errorf("variable sort should be concrete_t, got %s", rv.VSort)
	}
	if rv.Name != "X" {
		t.Errorf("variable name should be X, got %s", rv.Name)
	}
}

func TestResortASTConst(t *testing.T) {
	old := &UninterpretedSort{Name: "abstract_t"}
	new_ := &UninterpretedSort{Name: "concrete_t"}
	rn := mkSortRefinement(old, new_)

	c := NewConst("f", old)
	result := ResortAST(c, rn)
	rc, ok := result.(*Const)
	if !ok {
		t.Fatal("expected Const")
	}
	if !SortEqual(rc.CSort, new_) {
		t.Errorf("const sort should be concrete_t, got %s", rc.CSort)
	}
}

func TestResortSymbol(t *testing.T) {
	old := &UninterpretedSort{Name: "abstract_t"}
	new_ := &UninterpretedSort{Name: "concrete_t"}
	rn := mkSortRefinement(old, new_)

	c := NewConst("f", old)
	result := ResortSymbol(c, rn)
	if !SortEqual(result.CSort, new_) {
		t.Errorf("expected concrete_t, got %s", result.CSort)
	}
	if result.Name != "f" {
		t.Errorf("name should be preserved, got %s", result.Name)
	}
}

func TestCanonizeTypesNoOp(t *testing.T) {
	m := New()
	f := &LogicAnd{Terms: []Expr{True}}
	m.LabeledAxioms = []*LabeledFormula{{Formula: f}}

	// Empty refinement should be a no-op.
	m.CanonizeTypes(nil)
	if len(m.LabeledAxioms) != 1 {
		t.Error("axioms should be unchanged")
	}
}

func TestCanonizeTypesApplied(t *testing.T) {
	m := New()
	old := &UninterpretedSort{Name: "abstract_t"}
	new_ := &UninterpretedSort{Name: "concrete_t"}

	v, _ := NewVariable("X", old)
	c := NewConst("a", old)
	eq := &Eq{T1: v, T2: c}

	m.LabeledAxioms = []*LabeledFormula{
		m.Cfg.AstCfg.NewLabeledFormula(nil, eq),
	}
	m.GhostSorts["abstract_t"] = true
	m.SortOrder = []string{"abstract_t", "other"}

	refinement := []SortRefinement{{Old: old, New: new_}}
	m.CanonizeTypes(refinement)

	// Check that the axiom formula was resorted.
	resortedEq, ok := m.LabeledAxioms[0].Formula.(*Eq)
	if !ok {
		t.Fatal("expected Eq formula after canonize")
	}
	if !SortEqual(resortedEq.T1.NodeSort(), new_) {
		t.Errorf("T1 sort should be concrete_t, got %s", resortedEq.T1.NodeSort())
	}

	// Check that ghost sorts were updated.
	if m.GhostSorts["abstract_t"] {
		t.Error("abstract_t should have been removed from ghost sorts")
	}

	// Check that sort order was updated.
	if len(m.SortOrder) != 1 || m.SortOrder[0] != "other" {
		t.Errorf("sort order should be [other], got %v", m.SortOrder)
	}
}

// --- canonActionMap tests ---

func TestCanonActionMapNil(t *testing.T) {
	got := canonActionMap(nil)
	if got != "(insMap)" {
		t.Errorf("nil InsMap: expected (insMap), got %s", got)
	}
}

func TestCanonActionMapEmpty(t *testing.T) {
	m := NewInsMap[string, Action]()
	got := canonActionMap(m)
	if got != "(insMap)" {
		t.Errorf("empty InsMap: expected (insMap), got %s", got)
	}
}

func TestCanonActionMapWithActions(t *testing.T) {
	m := NewInsMap[string, Action]()
	a1 := &stubAction{name: "assume", formula: True}
	a2 := &stubAction{name: "assert", formula: False}
	m.Set("act1", a1)
	m.Set("act2", a2)

	got := canonActionMap(m)

	// Verify overall shape
	if !strings.HasPrefix(got, "(insMap ") || !strings.HasSuffix(got, ")") {
		t.Fatalf("wrong shape: %s", got)
	}

	// Verify insertion order: act1 before act2
	i1 := strings.Index(got, "act1:")
	i2 := strings.Index(got, "act2:")
	if i1 < 0 || i2 < 0 {
		t.Fatalf("missing keys in output: %s", got)
	}
	if i1 >= i2 {
		t.Errorf("expected act1 before act2 (insertion order), got: %s", got)
	}

	// Verify each entry contains the action sexp
	// lg.True = LogicAnd{} with empty terms, lg.False = LogicOr{} with empty terms
	exp1 := fmt.Sprintf("act1:(assume formula:%s)", string(True.Sexp()))
	exp2 := fmt.Sprintf("act2:(assert formula:%s)", string(False.Sexp()))
	if !strings.Contains(got, exp1) {
		t.Errorf("act1 sexp mismatch in: %s\nexpected to contain: %s", got, exp1)
	}
	if !strings.Contains(got, exp2) {
		t.Errorf("act2 sexp mismatch in: %s\nexpected to contain: %s", got, exp2)
	}
}

func TestCanonModuleFunctionsPreservesPolymorphicEntries(t *testing.T) {
	mod := New()
	idSort := &UninterpretedSort{Name: "id.t"}
	nodeSort := &UninterpretedSort{Name: "node.t"}
	idCmp, err := NewFunctionSort(idSort, idSort, Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort id: %v", err)
	}
	nodeCmp, err := NewFunctionSort(nodeSort, nodeSort, Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort node: %v", err)
	}
	pending, err := NewFunctionSort(idSort, nodeSort, Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort pending: %v", err)
	}

	mod.AddFunction("<", idCmp)
	mod.AddFunction("node.head", nodeSort)
	mod.AddFunction("node.tail", nodeSort)
	mod.AddFunction("<", nodeCmp)
	mod.AddFunction("trans.pending", pending)
	mod.AddFunction("<", nodeCmp)

	got := canonModuleFunctions(mod)
	want := "(insMap <:2 node.head:0 node.tail:0 <:2 trans.pending:2)"
	if got != want {
		t.Fatalf("canonModuleFunctions = %s, want %s", got, want)
	}

	if mod.Functions.Len() != 4 {
		t.Fatalf("Functions lookup map has %d entries, want 4", mod.Functions.Len())
	}
}

// --- Module.Canon() tests ---

func TestModuleCanonEmpty(t *testing.T) {
	m := New()
	got := string(m.Canon())

	// Empty module should have all empty slices/maps
	if !strings.HasPrefix(got, "(module ") {
		t.Fatalf("wrong prefix: %s", got)
	}
	if !strings.Contains(got, "axioms:[]") {
		t.Errorf("missing axioms:[] in: %s", got)
	}
	if !strings.Contains(got, "actions:(insMap)") {
		t.Errorf("missing actions:(insMap) in: %s", got)
	}
	if !strings.Contains(got, "schemata:(hash)") {
		t.Errorf("missing schemata:(hash) in: %s", got)
	}
}

func TestModuleCanonWithActions(t *testing.T) {
	m := New()
	a := &stubAction{name: "assume", formula: True}
	m.Actions.Set("my_action", a)

	got := string(m.Canon())

	expActions := fmt.Sprintf("actions:(insMap my_action:(assume formula:%s))", string(True.Sexp()))
	if !strings.Contains(got, expActions) {
		t.Errorf("actions field mismatch in: %s\nexpected to contain: %s", got, expActions)
	}
}

func TestModuleCanonFieldOrder(t *testing.T) {
	m := New()
	got := string(m.Canon())

	// Verify field order matches Python: axioms, defs, props, inits, conjs, schemata, actions
	fields := []string{"axioms:", "defs:", "props:", "inits:", "conjs:", "schemata:", "actions:"}
	prev := -1
	for _, f := range fields {
		idx := strings.Index(got, f)
		if idx < 0 {
			t.Errorf("missing field %s in: %s", f, got)
			continue
		}
		if idx <= prev {
			t.Errorf("field %s out of order in: %s", f, got)
		}
		prev = idx
	}
}

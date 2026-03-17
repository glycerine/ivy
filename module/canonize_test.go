package module

import (
	"testing"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
)

func TestResortSortUnrefined(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "t"}
	rn := map[string]lg.Sort{}
	result := ResortSort(s, rn)
	if !lg.SortEqual(result, s) {
		t.Error("unrefinement should not change sort")
	}
}

func TestResortSortRefined(t *testing.T) {
	old := &lg.UninterpretedSort{Name: "abstract_t"}
	new_ := &lg.UninterpretedSort{Name: "concrete_t"}
	rn := map[string]lg.Sort{"abstract_t": new_}

	result := ResortSort(old, rn)
	if !lg.SortEqual(result, new_) {
		t.Errorf("expected concrete_t, got %s", result)
	}
}

func TestResortSortFunctionSort(t *testing.T) {
	old := &lg.UninterpretedSort{Name: "abstract_t"}
	new_ := &lg.UninterpretedSort{Name: "concrete_t"}
	other := &lg.UninterpretedSort{Name: "u"}
	fs, _ := lg.NewFunctionSort(old, other)
	rn := map[string]lg.Sort{"abstract_t": new_}

	result := ResortSort(fs, rn)
	rfs, ok := result.(*lg.FunctionSort)
	if !ok {
		t.Fatal("expected FunctionSort")
	}
	if !lg.SortEqual(rfs.Domain()[0], new_) {
		t.Errorf("expected domain[0] = concrete_t, got %s", rfs.Domain()[0])
	}
	if !lg.SortEqual(rfs.Range(), other) {
		t.Errorf("expected range = u, got %s", rfs.Range())
	}
}

func TestResortAST(t *testing.T) {
	old := &lg.UninterpretedSort{Name: "abstract_t"}
	new_ := &lg.UninterpretedSort{Name: "concrete_t"}
	rn := map[string]lg.Sort{"abstract_t": new_}

	// Test variable resort
	v, _ := lg.NewVar("X", old)
	result := ResortAST(v, rn)
	rv, ok := result.(*lg.Var)
	if !ok {
		t.Fatal("expected Var")
	}
	if !lg.SortEqual(rv.VSort, new_) {
		t.Errorf("variable sort should be concrete_t, got %s", rv.VSort)
	}
	if rv.Name != "X" {
		t.Errorf("variable name should be X, got %s", rv.Name)
	}
}

func TestResortASTConst(t *testing.T) {
	old := &lg.UninterpretedSort{Name: "abstract_t"}
	new_ := &lg.UninterpretedSort{Name: "concrete_t"}
	rn := map[string]lg.Sort{"abstract_t": new_}

	c := lg.NewConst("f", old)
	result := ResortAST(c, rn)
	rc, ok := result.(*lg.Const)
	if !ok {
		t.Fatal("expected Const")
	}
	if !lg.SortEqual(rc.CSort, new_) {
		t.Errorf("const sort should be concrete_t, got %s", rc.CSort)
	}
}

func TestResortSymbol(t *testing.T) {
	old := &lg.UninterpretedSort{Name: "abstract_t"}
	new_ := &lg.UninterpretedSort{Name: "concrete_t"}
	rn := map[string]lg.Sort{"abstract_t": new_}

	c := lg.NewConst("f", old)
	result := ResortSymbol(c, rn)
	if !lg.SortEqual(result.CSort, new_) {
		t.Errorf("expected concrete_t, got %s", result.CSort)
	}
	if result.Name != "f" {
		t.Errorf("name should be preserved, got %s", result.Name)
	}
}

func TestCanonizeTypesNoOp(t *testing.T) {
	m := New()
	f := &lg.And{Terms: []lg.Node{lg.True}}
	m.LabeledAxioms = []*LabeledFormula{{Formula: f}}

	// Empty refinement should be a no-op.
	m.CanonizeTypes(nil)
	if len(m.LabeledAxioms) != 1 {
		t.Error("axioms should be unchanged")
	}
}

func TestCanonizeTypesApplied(t *testing.T) {
	m := New()
	old := &lg.UninterpretedSort{Name: "abstract_t"}
	new_ := &lg.UninterpretedSort{Name: "concrete_t"}

	v, _ := lg.NewVar("X", old)
	c := lg.NewConst("a", old)
	eq := &lg.Eq{T1: v, T2: c}

	m.LabeledAxioms = []*LabeledFormula{
		{Formula: eq},
	}
	m.GhostSorts["abstract_t"] = true
	m.SortOrder = []string{"abstract_t", "other"}

	refinement := []SortRefinement{{Old: old, New: new_}}
	m.CanonizeTypes(refinement)

	// Check that the axiom formula was resorted.
	resortedEq, ok := m.LabeledAxioms[0].Formula.(*lg.Eq)
	if !ok {
		t.Fatal("expected Eq formula after canonize")
	}
	if !lg.SortEqual(resortedEq.T1.NodeSort(), new_) {
		t.Errorf("T1 sort should be concrete_t, got %s", resortedEq.T1.NodeSort())
	}

	// Ghost sorts should no longer contain abstract_t.
	if m.GhostSorts["abstract_t"] {
		t.Error("abstract_t should have been removed from GhostSorts")
	}

	// SortOrder should no longer contain abstract_t.
	for _, s := range m.SortOrder {
		if s == "abstract_t" {
			t.Error("abstract_t should have been removed from SortOrder")
		}
	}
}

func TestResortLabeledFormulas(t *testing.T) {
	old := &lg.UninterpretedSort{Name: "abstract_t"}
	new_ := &lg.UninterpretedSort{Name: "concrete_t"}
	rn := map[string]lg.Sort{"abstract_t": new_}

	v, _ := lg.NewVar("X", old)
	lfs := []*LabeledFormula{
		{Formula: v, Lineno: 5, Temporal: true},
	}

	result := resortLabeledFormulas(lfs, rn)
	if len(result) != 1 {
		t.Fatal("expected 1 result")
	}
	if result[0].Lineno != 5 {
		t.Error("lineno should be preserved")
	}
	if !result[0].Temporal {
		t.Error("temporal should be preserved")
	}
	rv, ok := result[0].Formula.(*lg.Var)
	if !ok {
		t.Fatal("expected Var")
	}
	if !lg.SortEqual(rv.VSort, new_) {
		t.Error("variable sort should be refined")
	}
}

func TestResortASTDefinition(t *testing.T) {
	old := &lg.UninterpretedSort{Name: "abstract_t"}
	new_ := &lg.UninterpretedSort{Name: "concrete_t"}
	rn := map[string]lg.Sort{"abstract_t": new_}

	lhs := lg.NewConst("f", old)
	rhs := lg.NewConst("g", old)
	def := il.NewDefinition(lhs, rhs)

	result := ResortAST(def, rn)
	rd, ok := result.(*il.Definition)
	if !ok {
		t.Fatal("expected Definition")
	}
	if !lg.SortEqual(rd.Lhs.NodeSort(), new_) {
		t.Errorf("LHS sort should be concrete_t, got %s", rd.Lhs.NodeSort())
	}
}

func TestResortASTNil(t *testing.T) {
	result := ResortAST(nil, map[string]lg.Sort{"x": lg.Boolean})
	if result != nil {
		t.Error("ResortAST(nil) should return nil")
	}
}

package module

import (
	"testing"

	lg "github.com/glycerine/goivy/logic"
)

func mkSortRefinement(old, new_ lg.Sort) map[lg.NodeKey]*SortRefinement {
	sr := &SortRefinement{Old: old, New: new_}
	return map[lg.NodeKey]*SortRefinement{lg.SortKey(old): sr}
}

func TestResortSortUnrefined(t *testing.T) {
	s := &lg.UninterpretedSort{Name: "t"}
	rn := map[lg.NodeKey]*SortRefinement{}
	result := ResortSort(s, rn)
	if !lg.SortEqual(result, s) {
		t.Error("unrefinement should not change sort")
	}
}

func TestResortSortRefined(t *testing.T) {
	old := &lg.UninterpretedSort{Name: "abstract_t"}
	new_ := &lg.UninterpretedSort{Name: "concrete_t"}
	rn := mkSortRefinement(old, new_)

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
	rn := mkSortRefinement(old, new_)

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
	rn := mkSortRefinement(old, new_)

	// Test variable resort
	v, _ := lg.NewVariable("X", old)
	result := ResortAST(v, rn)
	rv, ok := result.(*lg.Variable)
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
	rn := mkSortRefinement(old, new_)

	c := lg.NewSymbol("f", old)
	result := ResortAST(c, rn)
	rc, ok := result.(*lg.Symbol)
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
	rn := mkSortRefinement(old, new_)

	c := lg.NewSymbol("f", old)
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

	v, _ := lg.NewVariable("X", old)
	c := lg.NewSymbol("a", old)
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

	// Check that ghost sorts were updated.
	if m.GhostSorts["abstract_t"] {
		t.Error("abstract_t should have been removed from ghost sorts")
	}

	// Check that sort order was updated.
	if len(m.SortOrder) != 1 || m.SortOrder[0] != "other" {
		t.Errorf("sort order should be [other], got %v", m.SortOrder)
	}
}

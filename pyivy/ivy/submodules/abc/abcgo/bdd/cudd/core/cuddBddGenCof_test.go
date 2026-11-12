package core

import "testing"

func bddValueAt(n *DdNode, assignment int) bool {
	if Cudd_IsConstant(n) {
		return n == n.manager.One
	}
	r := Cudd_Regular(n)
	child := r.E
	if assignment&(1<<r.Index) != 0 {
		child = r.T
	}
	return bddValueAt(Cudd_NotCond(child, Cudd_IsComplement(n)), assignment)
}

func bddSupportSet(root *DdNode) map[uint32]bool {
	support := make(map[uint32]bool)
	seen := make(map[*DdNode]bool)
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		n = Cudd_Regular(n)
		if n == nil || Cudd_IsConstant(n) || seen[n] {
			return
		}
		seen[n] = true
		support[n.Index] = true
		visit(n.T)
		visit(n.E)
	}
	visit(root)
	return support
}

func TestBDDConstrainAndRestrictExhaustiveThreeVariables(t *testing.T) {
	for _, order := range [][]int{{0, 1, 2}, {2, 0, 1}} {
		dd := NewBDDManager(3)
		if Cudd_ShuffleHeap(dd, order) == 0 {
			t.Fatalf("failed to set variable order %v", order)
		}
		functions := make([]*DdNode, 256)
		for truth := range functions {
			functions[truth] = bddFromTruth(dd, uint8(truth), 3)
		}
		for fTruth, f := range functions {
			for cTruth, c := range functions {
				constrained := Cudd_bddConstrain(dd, f, c)
				restricted := Cudd_bddRestrict(dd, f, c)
				if constrained == nil || restricted == nil {
					t.Fatalf("nil result for f=%02x c=%02x", fTruth, cTruth)
				}
				fSupport := bddSupportSet(f)
				for index := range bddSupportSet(restricted) {
					if !fSupport[index] {
						t.Fatalf("restrict added support variable %d for f=%02x c=%02x", index, fTruth, cTruth)
					}
				}
				if c != dd.LogicZero && Cudd_bddConstrain(dd, Cudd_Not(f), c) != Cudd_Not(constrained) {
					t.Fatalf("constrain lost complement canonicality for f=%02x c=%02x", fTruth, cTruth)
				}
				for assignment := 0; assignment < 8; assignment++ {
					if bddValueAt(c, assignment) {
						want := bddValueAt(f, assignment)
						if bddValueAt(constrained, assignment) != want {
							t.Fatalf("constrain mismatch f=%02x c=%02x assignment=%03b", fTruth, cTruth, assignment)
						}
						if bddValueAt(restricted, assignment) != want {
							t.Fatalf("restrict mismatch f=%02x c=%02x assignment=%03b", fTruth, cTruth, assignment)
						}
					}
				}
			}
		}
	}
}

func TestBDDGeneralizedCofactorRejectsForeignNodes(t *testing.T) {
	a, b := NewBDDManager(1), NewBDDManager(1)
	if Cudd_bddConstrain(a, Cudd_bddIthVar(a, 0), Cudd_bddIthVar(b, 0)) != nil {
		t.Fatal("constrain accepted a foreign care set")
	}
	if Cudd_bddRestrict(a, Cudd_bddIthVar(a, 0), Cudd_bddIthVar(b, 0)) != nil {
		t.Fatal("restrict accepted a foreign care set")
	}
}

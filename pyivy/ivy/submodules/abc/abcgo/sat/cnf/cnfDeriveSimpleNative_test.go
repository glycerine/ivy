package cnf

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// TestCnfDeriveSimpleMatchesNativeNumbering checks Cnf_DeriveSimple against
// the clauses and variable numbers that C's Cnf_DeriveSimple (cnfWrite.c)
// produces: variable 0 is unused, all COs are numbered first (when
// nOutputs != 0), then AND nodes, then CIs, then the constant; clauses are
// emitted per AND node (positive phase first), then the constant unit, then
// one unit per non-free CO and two equivalence clauses per free CO.
func TestCnfDeriveSimpleMatchesNativeNumbering(t *testing.T) {
	g := aig.Aig_ManStart(10)
	a := aig.Aig_ObjCreateCi(g)
	b := aig.Aig_ObjCreateCi(g)
	n := aig.Aig_And(g, a, b.Not())
	co0 := aig.Aig_ObjCreateCo(g, n)
	co1 := aig.Aig_ObjCreateCo(g, a.Not())
	p := Cnf_DeriveSimple(g, 1)
	if p.nVars != 7 {
		t.Fatalf("nVars = %d, want 7", p.nVars)
	}
	wantVars := map[aig.Edge]int32{co0: 1, co1: 2, n: 3, a: 4, b: 5, aig.Aig_ManConst1(g): 6}
	for e, v := range wantVars {
		if got := p.pVarNums[e.NodeID()]; got != v {
			t.Fatalf("var of object %d = %d, want %d", e.NodeID(), got, v)
		}
	}
	want := [][]int32{{6, 9, 10}, {7, 8}, {7, 11}, {12}, {6}, {4, 8}, {5, 9}}
	got := make([][]int32, p.nClauses)
	for i := range got {
		got[i] = Cnf_DataClause(p, int32(i))
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("clauses = %v, want %v", got, want)
	}
	if p.nLiterals != 13 {
		t.Fatalf("nLiterals = %d, want 13", p.nLiterals)
	}
}

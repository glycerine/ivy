package core

import (
	"fmt"
	"strings"
	"testing"
)

// mapOrderSignature records every node ID reachable from root plus the order
// of the BDD unique table, so a handle numbering or table layout that depends
// on Go's randomized map iteration shows up as a differing string.
func mapOrderSignature(dd *DdManager, root *DdNode) string {
	var b strings.Builder
	seen := map[*DdNode]bool{}
	var walk func(*DdNode)
	walk = func(e *DdNode) {
		n := Cudd_Regular(e)
		if seen[n] {
			return
		}
		seen[n] = true
		if Cudd_IsConstant(n) {
			fmt.Fprintf(&b, "c%d=%g ", n.ID, n.Value)
			return
		}
		fmt.Fprintf(&b, "n%d[%d %d %d %v] ", n.ID, n.Index, Cudd_Regular(n.T).ID, Cudd_Regular(n.E).ID, Cudd_IsComplement(n.E))
		walk(n.T)
		walk(n.E)
	}
	walk(root)
	b.WriteString("| ")
	for key, n := range dd.unique.entries {
		fmt.Fprintf(&b, "%d:%d ", key.index, n.ID)
	}
	return b.String()
}

func mapOrderFunction(dd *DdManager) *DdNode {
	// A comparator-like function whose size changes a lot with the order.
	n := len(dd.Vars) / 2
	f := Cudd_ReadOne(dd)
	Cudd_Ref(f)
	for i := 0; i < n; i++ {
		eq := Cudd_bddXnor(dd, dd.Vars[i], dd.Vars[n+i])
		Cudd_Ref(eq)
		g := Cudd_bddOr(dd, Cudd_bddAnd(dd, dd.Vars[i], Cudd_Not(dd.Vars[n+i])), Cudd_bddAnd(dd, eq, f))
		Cudd_Ref(g)
		Cudd_RecursiveDeref(dd, eq)
		Cudd_RecursiveDeref(dd, f)
		f = g
	}
	return f
}

// The scratch-rebuild reordering used to hand out new node IDs in map
// order, so the same reordering produced differently numbered diagrams.
func TestReorderAssignsNodeIDsDeterministically(t *testing.T) {
	var want string
	for run := 0; run < 20; run++ {
		dd := NewBDDManager(12)
		f := mapOrderFunction(dd)
		order := []int{11, 0, 10, 1, 9, 2, 8, 3, 7, 4, 6, 5}
		if Cudd_ShuffleHeap(dd, order) != 1 {
			t.Fatal("shuffle failed")
		}
		if Cudd_ReduceHeap(dd, CUDD_REORDER_SIFT, 0) != 1 {
			t.Fatal("sifting failed")
		}
		got := mapOrderSignature(dd, f)
		if run == 0 {
			want = got
		} else if got != want {
			t.Fatalf("run %d: reordered diagram differs:\n got %s\nwant %s", run, got, want)
		}
	}
}

func TestLinearTransformAssignsNodeIDsDeterministically(t *testing.T) {
	var want string
	for run := 0; run < 20; run++ {
		dd := NewBDDManager(10)
		f := mapOrderFunction(dd)
		for _, pair := range [][2]int{{0, 1}, {2, 3}, {4, 5}, {1, 2}} {
			x, y := dd.InvPerm[pair[0]], dd.InvPerm[pair[1]]
			if !Cudd_bddLinearTransformInPlace(dd, x, y) {
				t.Fatalf("linear transform %v failed", pair)
			}
		}
		got := mapOrderSignature(dd, f)
		if run == 0 {
			want = got
		} else if got != want {
			t.Fatalf("run %d: transformed diagram differs:\n got %s\nwant %s", run, got, want)
		}
	}
}

// Two constants within epsilon of a requested value: the one returned must
// not depend on map iteration order (the earliest constant wins).
func TestConstantEpsilonMatchIsDeterministic(t *testing.T) {
	for run := 0; run < 20; run++ {
		dd := NewBDDManager(1)
		eps := Cudd_ReadEpsilon(dd)
		var first *DdNode
		for i := 0; i < 8; i++ {
			c := Cudd_addConst(dd, 1+1.5*eps*float64(i))
			if i == 0 {
				first = c
			}
		}
		second := Cudd_addConst(dd, 1+1.5*eps)
		got := Cudd_addConst(dd, 1+0.75*eps)
		if got != first {
			t.Fatalf("run %d: got constant %d (value %v), want earliest %d (second is %d)", run, got.ID, got.Value, first.ID, second.ID)
		}
	}
}

package core

import (
	"math/big"
	"reflect"
	"testing"
)

func TestZddLinearSiftingPreservesFamilyAndValidOrder(t *testing.T) {
	dd := NewBDDManager(4)
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to initialize ZDD variables")
	}
	f := Cudd_zddUnion(dd,
		Cudd_zddChange(dd, dd.One, 0),
		Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 2), Cudd_zddChange(dd, dd.One, 3)))
	if f == nil {
		t.Fatal("failed to construct family")
	}
	Cudd_Ref(f)
	wantCount, err := dd.CountZDD(f)
	if err != nil {
		t.Fatal(err)
	}
	wantOrder := append([]int(nil), dd.InvPermZ...)
	if Cudd_zddLinearSifting(dd, 0, dd.ZddSize-1) != 1 {
		t.Fatal("linear ZDD sifting failed")
	}
	gotCount, err := dd.CountZDD(f)
	if err != nil || gotCount.Cmp(wantCount) != 0 {
		t.Fatalf("linear sifting changed family cardinality: got %v, want %v (err %v)", gotCount, wantCount, err)
	}
	if !validZddOrder(dd) || len(dd.InvPermZ) != len(wantOrder) {
		t.Fatalf("invalid ZDD order after sifting: %v", dd.InvPermZ)
	}
	if gotSize := zddHeapSize(dd); gotSize < 0 {
		t.Fatalf("negative ZDD heap size: %d", gotSize)
	}
}

func TestZddLinearAdjacentTransformMatchesSourceCofactorMapping(t *testing.T) {
	dd := NewBDDManager(2)
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to initialize ZDD variables")
	}
	// Family {{0}} has cardinality one; the linear transform changes the
	// coordinate of variable 0 by XOR with variable 1, preserving cardinality.
	f := Cudd_zddChange(dd, dd.One, 0)
	Cudd_Ref(f)
	truth := make([]bool, 1<<dd.ZddSize)
	for mask := range truth {
		assignment := []bool{mask&1 != 0, mask&2 != 0}
		truth[mask] = zddTestContains(dd, f, assignment)
	}
	before, err := dd.CountZDD(f)
	if err != nil || before.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("initial family count = %v, err %v", before, err)
	}
	if !linearZddAdjacentInPlace(dd, 0, 1) {
		t.Fatal("adjacent linear transform failed")
	}
	after, err := dd.CountZDD(f)
	if err != nil || after.Cmp(before) != 0 {
		t.Fatalf("transform changed family cardinality: got %v, want %v (err %v)", after, before, err)
	}
	for mask := range truth {
		assignment := []bool{mask&1 != 0, mask&2 != 0}
		oldAssignment := append([]bool(nil), assignment...)
		oldAssignment[0] = assignment[0] != assignment[1]
		oldMask := 0
		if oldAssignment[0] {
			oldMask |= 1
		}
		if oldAssignment[1] {
			oldMask |= 2
		}
		got := zddTestContains(dd, f, assignment)
		if got != truth[oldMask] {
			t.Fatalf("mask %02b: transformed membership=%v, want source cofactor mapping=%v", mask, got, truth[oldMask])
		}
	}
	if !reflect.DeepEqual(dd.InvPermZ, []int{0, 1}) {
		t.Fatalf("linear transform changed level order: %v", dd.InvPermZ)
	}
	for key, node := range dd.zddUnique {
		if key.index != node.Index || key.t != node.T || key.e != node.E {
			t.Fatal("linear transform left a stale unique-table key")
		}
	}
}

func zddTestContains(dd *DdManager, node *DdNode, assignment []bool) bool {
	level := 0
	for node != dd.One && node != dd.Zero {
		if int(node.Index) >= len(assignment) {
			return false
		}
		nodeLevel := dd.PermZ[node.Index]
		for level < nodeLevel {
			if assignment[dd.InvPermZ[level]] {
				return false
			}
			level++
		}
		if assignment[node.Index] {
			node = node.T
		} else {
			node = node.E
		}
		level++
	}
	if node != dd.One {
		return false
	}
	for ; level < dd.ZddSize; level++ {
		if assignment[dd.InvPermZ[level]] {
			return false
		}
	}
	return true
}

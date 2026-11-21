package core

import "testing"

// Build one node for each of the sixteen families over x and y. A set with
// bit x and/or bit y is represented by bit (x + 2*y) of the family mask.
func allTwoVarZDDFamilies(dd *DdManager, xi, yi uint32, terminal *DdNode) [16]*DdNode {
	var families [16]*DdNode
	for family := range families {
		leaf := func(bit int) *DdNode {
			if family&(1<<bit) != 0 {
				return terminal
			}
			return dd.Zero
		}
		y1 := dd.zddNode(yi, leaf(3), leaf(1))
		y0 := dd.zddNode(yi, leaf(2), leaf(0))
		families[family] = dd.zddNode(xi, y1, y0)
	}
	return families
}

func TestLinearZDDTransformChangesAllPinnedCoordinateFamilies(t *testing.T) {
	dd := Cudd_Init(0, 4, 32, 128, 0)
	defer Cudd_Quit(dd)
	tail := dd.zddNode(3, dd.One, dd.Zero)
	roots := allTwoVarZDDFamilies(dd, 1, 2, tail)
	pins := make(map[*DdNode]uint32)
	for _, f := range roots {
		Cudd_Ref(f)
		pins[f]++
	}
	// An ancestor above x may bypass x and point straight to a y node.
	upper := dd.zddNode(0, roots[4], dd.One)
	Cudd_Ref(upper)
	pins[upper]++
	for pass := 1; pass <= 2; pass++ {
		if !linearZddAdjacentInPlace(dd, 1, 2) {
			t.Fatal("transform")
		}
		for family, f := range roots {
			for mask := range 16 {
				assignment := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				bits := (mask >> 1) & 3
				if pass == 1 {
					bits ^= (bits >> 1) & 1
				}
				want := mask&9 == 8 && family&(1<<bits) != 0
				if got := zddTestContains(dd, f, assignment); got != want {
					t.Fatalf("pass %d family %04b set %04b = %t, want %t", pass, family, mask, got, want)
				}
				if family == 4 && zddTestContains(dd, upper, assignment) != (mask == 0 || mask&9 == 9 && family&(1<<bits) != 0) {
					t.Fatalf("ancestor pass %d set %04b", pass, mask)
				}
			}
		}
		live := 0
		for _, n := range dd.zddUnique.entries {
			if !n.childrenReleased {
				live++
			}
			if got := Cudd_ExternalRefCount(dd, n); got != pins[n] {
				t.Fatalf("pass %d handle %d has %d external references, want %d", pass, n.ID, got, pins[n])
			}
			if n.inverse == nil || n.inverse.Index != n.Index || dd.zddUnique.get(bddUniqueKey{n.Index, n.T, n.E}) != n {
				t.Fatalf("noncanonical handle %d", n.ID)
			}
		}
		if zddHeapSize(dd) != live {
			t.Fatalf("heap score %d, counted %d live nodes", zddHeapSize(dd), live)
		}
	}
	for _, f := range roots {
		Cudd_RecursiveDeref(dd, f)
	}
	Cudd_RecursiveDeref(dd, upper)
	Cudd_GarbageCollect(dd, true)
	if dd.zddUnique.count != 4 {
		t.Fatalf("after releasing all families got %d keys, want 4", dd.zddUnique.count)
	}
}

func BenchmarkLinearZDDAllPairs(b *testing.B) {
	dd := Cudd_Init(0, 4, 32, 128, 0)
	defer Cudd_Quit(dd)
	roots := allTwoVarZDDFamilies(dd, 1, 2, dd.zddNode(3, dd.One, dd.Zero))
	for _, f := range roots {
		Cudd_Ref(f)
	}
	b.ReportAllocs()
	for b.Loop() {
		if !linearZddAdjacentInPlace(dd, 1, 2) || !linearZddAdjacentInPlace(dd, 1, 2) {
			b.Fatal("transform")
		}
	}
}

func TestLinearZDDRepeatedTransformsReuseSubtableStorage(t *testing.T) {
	dd := Cudd_Init(0, 4, 32, 128, 0)
	defer Cudd_Quit(dd)
	roots := allTwoVarZDDFamilies(dd, 1, 2, dd.zddNode(3, dd.One, dd.Zero))
	for _, f := range roots {
		Cudd_Ref(f)
	}
	allocs := testing.AllocsPerRun(20, func() {
		if !linearZddAdjacentInPlace(dd, 1, 2) || !linearZddAdjacentInPlace(dd, 1, 2) {
			t.Fatal("transform")
		}
	})
	if allocs != 0 {
		t.Fatalf("repeated transform pair used %.0f allocations, want zero", allocs)
	}
}

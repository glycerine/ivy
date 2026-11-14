package core

import "testing"

func TestSwapZddAdjacentInPlacePreservesTruthTablesAndHandles(t *testing.T) {
	for table := 0; table < 256; table++ {
		if table == 0 {
			continue
		}
		dd := NewBDDManager(3)
		if Cudd_zddIthVar(dd, 2) == nil {
			t.Fatal("ZDD initialization failed")
		}
		var build func(int, int) *DdNode
		build = func(prefix, level int) *DdNode {
			if level == 3 {
				if table&(1<<prefix) != 0 {
					return dd.One
				}
				return dd.Zero
			}
			t := build(prefix|(1<<level), level+1)
			e := build(prefix, level+1)
			return dd.zddNode(uint32(level), t, e)
		}
		f := build(0, 0)
		if f == nil {
			t.Fatalf("convert table %02x", table)
		}
		// Also retain an explicit nonterminal handle found in the graph.
		var retained *DdNode
		for _, n := range dd.zddUnique {
			retained = n
			break
		}
		if !swapZddAdjacentInPlace(dd, 0, 1) {
			t.Fatal("swap failed")
		}
		if retained != nil && retained.manager != dd {
			t.Fatal("node handle lost")
		}
		for a := 0; a < 8; a++ {
			n := f
			for level, index := range dd.InvPermZ {
				if n == dd.Zero {
					break
				}
				if Cudd_IsConstant(n) || dd.PermZ[n.Index] > level {
					if a&(1<<index) != 0 {
						n = dd.Zero
					}
					continue
				}
				if a&(1<<index) != 0 {
					n = n.T
				} else {
					n = n.E
				}
			}
			if (n == dd.One) != (table&(1<<a) != 0) {
				t.Fatalf("table %02x assignment %d", table, a)
			}
		}
		if err := checkUniqueTable(dd, dd.zddUnique, dd.PermZ, "ZDD"); err != nil {
			t.Fatalf("table %02x: %v", table, err)
		}
	}
}

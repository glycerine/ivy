package core

import "testing"

func TestZddPortAllThreeVariableFunctions(t *testing.T) {
	dd := NewBDDManager(3)
	if Cudd_zddIthVar(dd, 2) == nil {
		t.Fatal("ZDD initialization failed")
	}
	var build func(uint8, int) *DdNode
	build = func(table uint8, level int) *DdNode {
		if level == 3 {
			if table&1 != 0 {
				return dd.One
			}
			return dd.LogicZero
		}
		var a, b uint8
		for j := 0; j < 1<<(3-level-1); j++ {
			a |= ((table >> (2 * j)) & 1) << j
			b |= ((table >> (2*j + 1)) & 1) << j
		}
		return Cudd_bddIte(dd, dd.Vars[level], build(b, level+1), build(a, level+1))
	}
	for table := 0; table < 256; table++ {
		f := build(uint8(table), 0)
		z := Cudd_zddPortFromBdd(dd, f)
		if z == nil {
			t.Fatalf("from BDD %02x", table)
		}
		for assignment := 0; assignment < 8; assignment++ {
			n := z
			for level := 0; level < 3; level++ {
				bit := assignment&(1<<dd.InvPermZ[level]) != 0
				if n == dd.Zero {
					break
				}
				if Cudd_IsConstant(n) || dd.PermZ[n.Index] > level {
					if bit {
						n = dd.Zero
					}
					continue
				}
				if bit {
					n = n.T
				} else {
					n = n.E
				}
			}
			if (n == dd.One) != ((table & (1 << assignment)) != 0) {
				t.Fatalf("table %02x assignment %d", table, assignment)
			}
		}
		back := Cudd_zddPortToBdd(dd, z)
		if back != f {
			t.Fatalf("round trip %02x got %p want %p", table, back, f)
		}
	}
}

func TestZddPortPreservesReorderedAssignments(t *testing.T) {
	dd := NewBDDManager(3)
	if Cudd_ShuffleHeap(dd, []int{2, 0, 1}) != 1 || Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to align reordered BDD and ZDD variables")
	}
	x, y, z := dd.Vars[0], dd.Vars[1], dd.Vars[2]
	for _, f := range []*DdNode{dd.One, dd.LogicZero, x, Cudd_Not(y), Cudd_bddIte(dd, z, x, Cudd_Not(y))} {
		converted := Cudd_zddPortFromBdd(dd, f)
		if converted == nil {
			t.Fatal("BDD to ZDD failed")
		}
		restored := Cudd_zddPortToBdd(dd, converted)
		for mask := 0; mask < 8; mask++ {
			assignment := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			want, err := dd.Evaluate(f, assignment)
			if err != nil {
				t.Fatal(err)
			}
			zdd := converted
			for level, index := range dd.InvPermZ {
				if zdd == dd.Zero {
					break
				}
				if Cudd_IsConstant(zdd) || dd.PermZ[zdd.Index] > level {
					if assignment[index] {
						zdd = dd.Zero
					}
					continue
				}
				if assignment[index] {
					zdd = zdd.T
				} else {
					zdd = zdd.E
				}
			}
			if (zdd == dd.One) != want {
				t.Fatalf("mask %d: ZDD does not represent input function", mask)
			}
			got, err := dd.Evaluate(restored, assignment)
			if err != nil || got != want {
				t.Fatalf("mask %d: got %v want %v, error %v", mask, got, want, err)
			}
		}
	}
}

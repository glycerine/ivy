package core

import (
	"bytes"
	"strings"
	"testing"
)

func TestZddSymmetricSiftingPreservesFamilyAndRecordsRing(t *testing.T) {
	dd := NewBDDManager(2)
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to create ZDD variables")
	}
	root := dd.Univ[1]
	want, err := dd.CountZDD(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := cuddZddSymmSifting(dd, 0, dd.ZddSize-1, false); got != 3 {
		t.Fatalf("internal symm-sifting result = %d, want 3 (1 + two symmetric variables)", got)
	}
	if len(dd.symmetryGroups) != 1 || len(dd.symmetryGroups[0]) != 2 {
		t.Fatalf("symmetry groups = %v, want one group containing both variables", dd.symmetryGroups)
	}
	if !validZddOrder(dd) {
		t.Fatal("symmetric sifting corrupted the ZDD permutation")
	}
	got, err := dd.CountZDD(root)
	if err != nil || got.Cmp(want) != 0 {
		t.Fatalf("sifting changed the represented family: got %v, want %v (err %v)", got, want, err)
	}
	if public := Cudd_zddReduceHeap(dd, CUDD_REORDER_SYMM_SIFT, 0); public != 1 {
		t.Fatalf("public SYMM_SIFT status = %d, want success", public)
	}
}

func TestZddSymmProfileStartsAtRequestedLowerBound(t *testing.T) {
	dd := NewBDDManager(2)
	dd.ZddSize = 2
	dd.PermZ, dd.InvPermZ = []int{0, 1}, []int{0, 1}
	dd.symmetryGroups = [][]int{{0, 1}}
	var out bytes.Buffer
	Cudd_zddSymmProfileTo(dd, 1, 2, &out)
	if got := out.String(); !strings.Contains(got, "Group:  1  0\nTotal Symmetric = 2\nTotal Groups = 1\n") {
		t.Fatalf("profile did not rotate the group at lower bound: %q", got)
	}
}

func TestZddSymmetricSiftingPreservesThreeVariableFamilies(t *testing.T) {
	for table := 0; table < 256; table++ {
		for _, method := range []Cudd_ReorderingType{CUDD_REORDER_SYMM_SIFT, CUDD_REORDER_SYMM_SIFT_CONV} {
			dd := NewBDDManager(3)
			if Cudd_zddIthVar(dd, 2) == nil {
				t.Fatal("failed to initialize ZDD order")
			}
			var build func(int, int) *DdNode
			build = func(prefix, level int) *DdNode {
				if level == 3 {
					if table&(1<<prefix) != 0 {
						return dd.One
					}
					return dd.Zero
				}
				then := build(prefix|(1<<level), level+1)
				elseNode := build(prefix, level+1)
				return dd.zddNode(uint32(level), then, elseNode)
			}
			root := build(0, 0)
			if root == nil {
				t.Fatalf("table %02x could not be built", table)
			}
			if Cudd_zddReduceHeap(dd, method, 0) != 1 {
				t.Fatalf("symmetric sifting method %d failed for table %02x", method, table)
			}
			if !validZddOrder(dd) {
				t.Fatalf("table %02x: invalid order", table)
			}
			for assignment := 0; assignment < 8; assignment++ {
				node := root
				for level, variable := range dd.InvPermZ {
					if node == dd.Zero {
						break
					}
					if Cudd_IsConstant(node) || dd.PermZ[node.Index] > level {
						if assignment&(1<<variable) != 0 {
							node = dd.Zero
						}
						continue
					}
					if assignment&(1<<variable) != 0 {
						node = node.T
					} else {
						node = node.E
					}
				}
				if (node == dd.One) != (table&(1<<assignment) != 0) {
					t.Fatalf("table %02x assignment %d changed", table, assignment)
				}
			}
			if err := checkUniqueTable(dd, dd.zddUnique, dd.PermZ, "ZDD"); err != nil {
				t.Fatalf("table %02x: %v", table, err)
			}
		}
	}
}

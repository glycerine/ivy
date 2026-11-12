package cascade

import (
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"hash/fnv"
	"testing"
)

func TestNonStrictFourColumnsDirectInputs(t *testing.T) {
	dd := cudd.NewBDDManager(4)
	columns := make([]*DdNode, 4)
	for i := range columns {
		columns[i] = Extra_bddBitsToCube(dd, i, 2, dd.Vars[:2], true)
	}
	vars := cudd.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	r, simple := Extra_bddEncodingNonStrict(dd, columns, vars, dd.Vars[2:])
	if simple != 2 {
		t.Fatalf("direct bits=%d", simple)
	}
	want := cudd.Cudd_bddAnd(dd, cudd.Cudd_bddXnor(dd, dd.Vars[0], dd.Vars[2]), cudd.Cudd_bddXnor(dd, dd.Vars[1], dd.Vars[3]))
	if r != want {
		t.Fatal("direct encoding relation")
	}
}
func TestNonStrictAllPartitions(t *testing.T) {
	hash := fnv.New64a()
	// Every surjection from eight input assignments onto up to four columns.
	// Verify total, exclusive codes and preservation of each column identity.
	for partition := 0; partition < 65536; partition++ {
		if partition%17 != 0 {
			continue
		}
		dd := cudd.NewBDDManager(5)
		cols := make([]*DdNode, 4)
		for i := range cols {
			cols[i] = dd.LogicZero
		}
		used := map[int]bool{}
		for a := 0; a < 8; a++ {
			c := (partition >> uint(2*a)) & 3
			used[c] = true
			cols[c] = cudd.Cudd_bddOr(dd, cols[c], Extra_bddBitsToCube(dd, a, 3, dd.Vars[:3], true))
		}
		compact := []*DdNode{}
		for i, c := range cols {
			if used[i] {
				compact = append(compact, c)
			}
		}
		cube := Extra_bddBitsToCube(dd, 7, 3, dd.Vars[:3], true)
		r, simple := Extra_bddEncodingNonStrict(dd, compact, cube, dd.Vars[3:])
		if r == nil || simple > 2 {
			t.Fatal(partition, "encoding failed")
		}
		hash.Write([]byte{byte(simple)})
		reverse := map[int]int{}
		for a := 0; a < 8; a++ {
			found := -1
			for code := 0; code < 4; code++ {
				assignment := []bool{a&4 != 0, a&2 != 0, a&1 != 0, code&2 != 0, code&1 != 0}
				ok, err := dd.Evaluate(r, assignment)
				if err != nil {
					t.Fatal(err)
				}
				if ok {
					if found != -1 {
						t.Fatal("multiple codes", partition, a)
					}
					found = code
				}
			}
			if found < 0 {
				t.Fatal("missing code", partition, a)
			}
			hash.Write([]byte{1})
			// Non-strict encodings may give one column multiple codes in different
			// direct-input cofactors, but no code may identify different columns.
			column := (partition >> uint(2*a)) & 3
			if prev, ok := reverse[found]; ok && prev != column {
				t.Fatal("code aliases columns", partition, a)
			}
			reverse[found] = column
		}
	}
	if got := hash.Sum64(); got != 0x6b850130b3287765 {
		t.Fatalf("original C oracle hash: got %016x", got)
	}
}
func TestPathsAndWidthAcrossSkippedLevels(t *testing.T) {
	dd := cudd.NewBDDManager(4)
	f := cudd.Cudd_bddIte(dd, dd.Vars[0], dd.Vars[3], dd.Vars[2])
	a := cudd.Cudd_BddToAdd(dd, f)
	profile := Extra_ProfileWidth(dd, a)
	want := []int{1, 2, 2, 3, 2}
	for i, v := range want {
		if profile[i] != v {
			t.Fatal(profile)
		}
	}
	paths := Extra_bddNodePathsUnderCutArray(dd, []*DdNode{a}, []*DdNode{dd.One}, 2)
	sum := dd.LogicZero
	for i, n := range paths.Nodes {
		pattern := cudd.Cudd_addBddPattern(dd, n)
		sum = cudd.Cudd_bddOr(dd, sum, cudd.Cudd_bddAnd(dd, paths.Columns[i], pattern))
	}
	if sum != f {
		t.Fatal("cut reconstruction")
	}
}

func TestProfileZDDLevelsAndInvalidOwnership(t *testing.T) {
	dd := cudd.NewBDDManager(0)
	dd.ZddSize = 3
	dd.InvPermZ = []int{2, 0, 1}
	dd.PermZ = []int{1, 2, 0}
	if cudd.InitializeZDDUniverse(dd) != 1 {
		t.Fatal("ZDD initialization")
	}
	f := cudd.Cudd_zddUnion(dd, cudd.Cudd_zddChange(dd, dd.One, 0), cudd.Cudd_zddChange(dd, dd.One, 2))
	profile := Extra_ProfileWidth(dd, f)
	if len(profile) != 4 {
		t.Fatal(profile)
	}
	for cut := 0; cut <= 3; cut++ {
		frontier := map[*DdNode]bool{}
		var visit func(*DdNode)
		visit = func(n *DdNode) {
			if cudd.Cudd_IsConstant(n) || dd.PermZ[n.Index] >= cut {
				frontier[n] = true
				return
			}
			visit(n.T)
			visit(n.E)
		}
		visit(f)
		if profile[cut] != len(frontier) {
			t.Fatal("ZDD profile", cut, profile, len(frontier))
		}
	}
	other := cudd.NewBDDManager(1)
	if Extra_ProfileWidth(dd, other.Vars[0]) != nil || Extra_ProfileWidth(nil, f) != nil {
		t.Fatal("invalid profile domain accepted")
	}
}

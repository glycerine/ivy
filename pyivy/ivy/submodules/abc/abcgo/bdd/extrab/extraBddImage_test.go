package extrab

import (
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func imageDirect(dd *DdManager, care *DdNode, parts []*DdNode, keep []int) *DdNode {
	relation := care
	for _, p := range parts {
		relation = cudd.Cudd_bddAnd(dd, relation, p)
	}
	retained := map[int]bool{}
	for _, i := range keep {
		retained[i] = true
	}
	quantify := map[int]bool{}
	for i := range dd.Vars {
		if !retained[i] {
			quantify[i] = true
		}
	}
	return cudd.Cudd_bddExistAbstract(dd, relation, imageCube(dd, quantify))
}
func TestImageTreeAndMonolithicAllCareSets(t *testing.T) {
	seed := uint64(17)
	for sample := 0; sample < 128; sample++ {
		dd := cudd.NewBDDManager(3)
		parts := []*DdNode{}
		count := sample % 6
		for i := 0; i < count; i++ {
			seed ^= seed << 13
			seed ^= seed >> 7
			seed ^= seed << 17
			parts = append(parts, autoTruth(dd, seed&255, 3))
		}
		initial := cudd.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[1])
		keep := []*DdNode{dd.Vars[2]}
		tree := Extra_bddImageStart(dd, initial, count, parts, 1, keep, false)
		mono := Extra_bddImageStart2(dd, initial, count, parts, 1, keep, false)
		if tree == nil || mono == nil {
			t.Fatal("image initialization")
		}
		want := imageDirect(dd, initial, parts, []int{2})
		if Extra_bddImageRead(tree) != want || Extra_bddImageRead2(mono) != want {
			t.Fatal("initial image", sample)
		}
		for careTable := uint64(0); careTable < 16; careTable++ {
			care := autoTruth(dd, careTable, 2)
			want = imageDirect(dd, care, parts, []int{2})
			if Extra_bddImageCompute(tree, care) != want || Extra_bddImageCompute2(mono, care) != want {
				t.Fatal("image recompute", sample, careTable)
			}
		}
		if len(dd.Vars) != 3 {
			t.Fatal("scheduler allocated fictitious partition variables")
		}
		Extra_bddImageTreeDelete(tree)
		Extra_bddImageTreeDelete2(mono)
		Extra_bddImageTreeDelete(tree)
		Extra_bddImageTreeDelete2(mono)
		if Extra_bddImageRead(tree) != nil || Extra_bddImageRead2(mono) != nil || Extra_bddImageCompute(tree, dd.One) != nil {
			t.Fatal("deleted tree remained active")
		}
	}
}
func TestImageCareSupportAndInvalidInputs(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	parts := []*DdNode{dd.Vars[1]}
	keep := []*DdNode{dd.Vars[2]}
	tree := Extra_bddImageStart(dd, dd.Vars[0], 1, parts, 1, keep, false)
	mono := Extra_bddImageStart2(dd, dd.Vars[0], 1, parts, 1, keep, false)
	old := Extra_bddImageRead(tree)
	if Extra_bddImageCompute(tree, dd.Vars[1]) != nil || Extra_bddImageRead(tree) != old {
		t.Fatal("care support expansion accepted or destroyed old result")
	}
	if Extra_bddImageCompute2(mono, dd.Vars[1]) != imageDirect(dd, dd.Vars[1], parts, []int{2}) {
		t.Fatal("monolithic support expansion")
	}
	other := cudd.NewBDDManager(3)
	if Extra_bddImageStart(dd, other.Vars[0], 0, nil, 0, nil, false) != nil || Extra_bddImageStart(dd, dd.One, 1, nil, 0, nil, false) != nil || Extra_bddImageStart(dd, dd.One, 0, nil, 1, []*DdNode{cudd.Cudd_Not(dd.Vars[0])}, false) != nil || Extra_bddImageCompute2(mono, cudd.Cudd_addIthVar(dd, 0)) != nil {
		t.Fatal("invalid image input accepted")
	}
	Extra_bddImageTreeDelete(tree)
	Extra_bddImageTreeDelete2(mono)
}
func TestImageTreeSurvivesManagerReordering(t *testing.T) {
	dd := cudd.NewBDDManager(4)
	parts := []*DdNode{cudd.Cudd_bddXnor(dd, dd.Vars[0], dd.Vars[2]), cudd.Cudd_bddXor(dd, dd.Vars[1], dd.Vars[3])}
	for _, p := range parts {
		cudd.Cudd_Ref(p)
	}
	care := cudd.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[1])
	tree := Extra_bddImageStart(dd, care, 2, parts, 2, dd.Vars[2:], false)
	if cudd.Cudd_ShuffleHeap(dd, []int{2, 0, 3, 1}) == 0 {
		t.Fatal("shuffle")
	}
	care = cudd.Cudd_bddOr(dd, dd.Vars[0], dd.Vars[1])
	if Extra_bddImageCompute(tree, care) != imageDirect(dd, care, parts, []int{2, 3}) {
		t.Fatal("image after reorder")
	}
	Extra_bddImageTreeDelete(tree)
	for _, p := range parts {
		cudd.Cudd_RecursiveDeref(dd, p)
	}
}

func TestOriginalCImageFixture(t *testing.T) {
	var hash uint64 = 14695981039346656037
	byteHash := func(v byte) { hash ^= uint64(v); hash *= 1099511628211 }
	seed := uint64(17)
	for sample := 0; sample < 128; sample++ {
		dd := cudd.NewBDDManager(3)
		count := sample % 3
		parts := make([]*DdNode, count)
		for i := range parts {
			seed ^= seed << 13
			seed ^= seed >> 7
			seed ^= seed << 17
			parts[i] = autoTruth(dd, seed&255, 3)
		}
		initial := cudd.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[1])
		keep := []*DdNode{dd.Vars[2]}
		tree := Extra_bddImageStart(dd, initial, count, parts, 1, keep, false)
		mono := Extra_bddImageStart2(dd, initial, count, parts, 1, keep, false)
		byteHash(byte(autoTruthBits(t, dd, Extra_bddImageRead(tree), 3)))
		byteHash(byte(autoTruthBits(t, dd, Extra_bddImageRead2(mono), 3)))
		for table := uint64(0); table < 16; table++ {
			care := autoTruth(dd, table, 2)
			byteHash(byte(autoTruthBits(t, dd, Extra_bddImageCompute(tree, care), 3)))
			byteHash(byte(autoTruthBits(t, dd, Extra_bddImageCompute2(mono, care), 3)))
		}
		Extra_bddImageTreeDelete(tree)
		Extra_bddImageTreeDelete2(mono)
	}
	if hash != 0x2c5b1f2faff964eb {
		t.Fatalf("original C image fixture: %016x", hash)
	}
}

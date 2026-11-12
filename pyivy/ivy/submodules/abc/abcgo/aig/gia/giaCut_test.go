package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCutMergeAndDominance(t *testing.T) {
	a := &Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{1, 5}}
	a.Sign = Gia_CutGetSign(a)
	b := &Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{2, 5}}
	b.Sign = Gia_CutGetSign(b)
	var c Gia_Cut_t
	if Gia_CutMergeOrder(a, b, &c, 3) != 1 || c.nLeaves != 3 || c.pLeaves[0] != 1 || c.pLeaves[1] != 2 || c.pLeaves[2] != 5 {
		t.Fatalf("bad union %+v", c)
	}
	if Gia_CutMergeOrder(a, b, &c, 2) != 0 {
		t.Fatal("oversized union accepted")
	}
	if Gia_CutCheck(&c, a) != 1 {
		t.Fatal("containment missed")
	}
	cuts := []*Gia_Cut_t{&c, a, b}
	if Gia_CutSetLastCutContains(cuts, 2) != 1 || cuts[0] != a || cuts[1] != b {
		t.Fatal("dominance handling")
	}
	if cuts[2] == cuts[0] || cuts[2] == cuts[1] {
		t.Fatal("candidate storage alias")
	}
}

func TestCutEnumerationTruth(t *testing.T) {
	for _, size := range []int32{4, 6, 8} {
		p := Gia_ManStart(20)
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a, b)
		y := Gia_ManAppendAnd(p, x, c^1)
		Gia_ManAppendCo(p, y)
		store := Gia_StoAlloc(p, size, 10, 1, 1, 0)
		for id := int32(0); id < p.nObjs; id++ {
			Gia_StoRefObj(store, id)
		}
		Gia_StoComputeCutsConst0(store, 0)
		for i := int32(0); i < Gia_ManCiNum(p); i++ {
			Gia_StoComputeCutsCi(store, Gia_ManCi(p, i).id)
		}
		Gia_StoComputeCutsNode(store, x>>1)
		Gia_StoComputeCutsNode(store, y>>1)
		count := Gia_StoPrepareSet(store, y>>1, 0)
		found := false
		for i := int32(0); i < count; i++ {
			cut := &store.pCuts[0][i]
			if cut.nLeaves != 3 {
				continue
			}
			found = true
			truth := Gia_CutTruth(store, cut)
			for mask := 0; mask < 8; mask++ {
				got := (truth[mask>>6] >> uint(mask&63) & 1) != 0
				if cut.iFunc&1 != 0 {
					got = !got
				}
				want := mask&1 != 0 && mask&2 != 0 && mask&4 == 0
				if got != want {
					t.Fatalf("size %d mask %d cut %+v", size, mask, cut)
				}
			}
		}
		if !found {
			t.Fatalf("size %d no 3-leaf cut: %v", size, vec.Vec_IntArray(vec.Vec_WecEntry(store.vCuts, int(y>>1))))
		}
	}
}

func TestCutTruthMinimizesIrrelevantLeaf(t *testing.T) {
	for _, size := range []int32{4, 8} {
		p := Gia_ManStart(12)
		store := Gia_StoAlloc(p, size, 10, 1, 1, 0)
		// (a & b) & (a & !b) is false even though the merged cut has two leaves.
		left := &Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{1, 2}}
		right := &Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{1, 2}}
		left.iFunc = int32(vec.Vec_MemHashInsert(store.vTtMem, func() []uint64 {
			n := max(1, 1<<uint(max(0, int(size)-6)))
			words := make([]uint64, n)
			for i := range words {
				words[i] = 0x8888888888888888
			}
			return words
		}()) << 1)
		right.iFunc = int32(vec.Vec_MemHashInsert(store.vTtMem, func() []uint64 {
			n := max(1, 1<<uint(max(0, int(size)-6)))
			words := make([]uint64, n)
			for i := range words {
				words[i] = 0x2222222222222222
			}
			return words
		}()) << 1)
		result := &Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{1, 2}}
		if Gia_CutComputeTruth(store, left, right, 0, 0, result, 0) != 1 || result.nLeaves != 0 || result.iFunc != 0 {
			t.Fatalf("size %d result %+v", size, result)
		}
	}
}

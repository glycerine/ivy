package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestKfHashAndSelect(t *testing.T) {
	p := &Kf_Set_t{nLutSize: 4, nCutNum: 3, TableMask: 255}
	for i := range p.pList {
		p.pList[i] = -1
	}
	for i, leaves := range [][]int32{{1, 2, 3}, {1, 4}, {2, 5}} {
		c := &p.pCutsR[i]
		c.nLeaves = int32(len(leaves))
		copy(c.pLeaves[:], leaves)
		c.Sign = Kf_SetCutGetSign(c)
		c.Delay = int32(3 - i)
		c.Area = float32(i + 1)
		Kf_SetAddToList(p, c, 0)
	}
	Kf_HashPopulate(p, &p.pCutsR[0])
	if p.nTEntries != 3 || Kf_HashLookup(p, 2) != -1 || Kf_HashFindOrAdd(p, 8) != 0 || Kf_HashFindOrAdd(p, 9) != 1 {
		t.Fatal("hash behavior")
	}
	Kf_HashCleanup(p, 3)
	if Kf_HashLookup(p, 8) < 0 || Kf_HashLookup(p, 2) != -1 {
		t.Fatal("partial cleanup")
	}
	Kf_HashCleanup(p, 0)
	Kf_SetSelectBest(p, 0, 1)
	if p.nCuts != 2 || p.pCutBest != &p.pCutsR[2] || p.ppCuts[0].nLeaves < p.ppCuts[1].nLeaves {
		t.Fatalf("selection %d %v", p.nCuts, p.pCutBest)
	}
}

func TestKfMergeAndDominance(t *testing.T) {
	makeList := func(ids ...int32) []int32 {
		v := vec.Vec_IntAlloc(0)
		Kf_ManStoreStart(v, 0)
		for _, id := range ids {
			Kf_ManStoreAddUnit(v, id, 0, 0)
		}
		return append([]int32(nil), vec.Vec_IntArray(v)...)
	}
	for _, ordered := range []bool{false, true} {
		p := &Kf_Set_t{nLutSize: 4, nCutNum: 4, TableMask: 255}
		a, b := makeList(1, 2), makeList(2, 3)
		// Unit cut pairs with identical leaves are not valid AND fanins in KF;
		// use disjoint fanin lists to exercise all merged pairs.
		b = makeList(3, 4)
		if ordered {
			Kf_SetMergeOrder(p, a, b, 0, 0)
		} else {
			Kf_SetMergeSimple(p, a, b, 0, 0)
		}
		if p.nCuts != 3 || p.pCutBest == nil || p.CutCount[2] != 4 || p.CutCount[3] != 3 {
			t.Fatalf("ordered %v count %d statistics %v", ordered, p.nCuts, p.CutCount)
		}
		if !Kf_CheckCuts(p) {
			t.Fatal("invalid cut set")
		}
	}
}

func TestKfHashMerge(t *testing.T) {
	makeList := func(ids ...int32) []int32 {
		v := vec.Vec_IntAlloc(0)
		Kf_ManStoreStart(v, 0)
		for _, id := range ids {
			Kf_ManStoreAddUnit(v, id, 0, 0)
		}
		return append([]int32(nil), vec.Vec_IntArray(v)...)
	}
	p := &Kf_Set_t{nLutSize: 4, nCutNum: 4, TableMask: 255}
	Kf_SetMerge(p, makeList(1, 2), makeList(3, 4), 0, 0)
	if p.nCuts != 3 || p.CutCount[0] != 4 || p.CutCount[1] != 4 || p.CutCount[2] != 4 || p.CutCount[3] != 3 || p.nTEntries != 0 || !Kf_CheckCuts(p) {
		t.Fatalf("cut set %d stats %v entries %d", p.nCuts, p.CutCount, p.nTEntries)
	}
}

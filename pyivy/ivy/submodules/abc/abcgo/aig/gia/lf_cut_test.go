package gia

import (
	"reflect"
	"testing"
)

func TestLfCutMergeAndCompare(t *testing.T) {
	a, b, c := &Lf_Cut_t{nLeaves: 2, Delay: 1, Flow: 2}, &Lf_Cut_t{nLeaves: 2, Delay: 2, Flow: 1}, &Lf_Cut_t{nLeaves: 1}
	a.pLeaves[0], a.pLeaves[1] = 1, 4
	b.pLeaves[0], b.pLeaves[1] = 4, 8
	c.pLeaves[0] = 9
	a.Sign, b.Sign, c.Sign = Lf_CutGetSign(a), Lf_CutGetSign(b), Lf_CutGetSign(c)
	var out Lf_Cut_t
	if Lf_CutMergeOrder(a, b, &out, 3) != 1 || !reflect.DeepEqual(out.pLeaves[:out.nLeaves], []int32{1, 4, 8}) || out.Sign != a.Sign|b.Sign || !Lf_SetCutIsContainedOrder(&out, a) {
		t.Fatalf("merge %+v", out)
	}
	if Lf_CutMergeOrder(a, b, &out, 2) != 0 || Lf_CutMergeOrderMux(a, b, c, &out, 4) != 1 || out.nLeaves != 4 {
		t.Fatal("limit or mux merge")
	}
	if Lf_CutCompareDelay(a, b) >= 0 || Lf_CutCompareArea(a, b) <= 0 {
		t.Fatal("cut order")
	}
}
func TestLfBestAndUnit(t *testing.T) {
	var cut Lf_Cut_t
	Lf_CutCreateUnit(&cut, 65)
	if cut.nLeaves != 1 || cut.Sign != 2 || cut.iFunc != 2 {
		t.Fatal(cut)
	}
	var best Lf_Bst_t
	if Lf_BestCutIndex(&best) != 2 {
		t.Fatal("no best")
	}
	best.Cut[1] = Lf_Plc_t{fUsed: 1, Handle: 4}
	if Lf_BestCutIndex(&best) != 1 || Lf_BestDiffCuts(&best) != 1 || Lf_BestIsMapped(&best) != 1 {
		t.Fatal("best state")
	}
}

func TestLfCutStoreDominance(t *testing.T) {
	mk := func(delay int32, flow float32, leaves ...int32) *Lf_Cut_t {
		c := &Lf_Cut_t{Delay: delay, Flow: flow, nLeaves: uint32(len(leaves))}
		copy(c.pLeaves[:], leaves)
		c.Sign = Lf_CutGetSign(c)
		return c
	}
	a, b, c := mk(2, 2, 1, 2), mk(3, 3, 3, 4, 5), mk(1, 1, 3, 4)
	store := []*Lf_Cut_t{a, b, c, nil}
	n := Lf_SetAddCut(store, 2, 4)
	if n != 2 || store[0] != c || store[1] != a {
		t.Fatalf("store n=%d best=%p second=%p", n, store[0], store[1])
	}
	if Lf_SetCheckArray(store, n) != 1 {
		t.Fatal("store check")
	}
	store[n] = mk(4, 4, 1, 2, 6)
	if Lf_SetLastCutIsContained(store, n) != 1 {
		t.Fatal("missed containment")
	}
}

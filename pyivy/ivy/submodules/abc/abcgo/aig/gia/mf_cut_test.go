package gia

import "testing"

func TestMfCutMergeAndContainment(t *testing.T) {
	makeCut := func(ids ...int32) *Mf_Cut_t {
		c := &Mf_Cut_t{nLeaves: uint32(len(ids))}
		copy(c.pLeaves[:], ids)
		c.Sign = Mf_CutGetSign(ids, int32(len(ids)))
		return c
	}
	a, b, c := makeCut(1, 3, 65), makeCut(2, 3), makeCut(1, 4)
	var out Mf_Cut_t
	if Mf_CutMergeOrderMux(a, b, c, &out, 5) != 1 || out.nLeaves != 5 {
		t.Fatalf("merge: %+v", out)
	}
	for i, want := range []int32{1, 2, 3, 4, 65} {
		if out.pLeaves[i] != want {
			t.Fatalf("leaf %d: %d", i, out.pLeaves[i])
		}
	}
	if out.Sign != a.Sign|b.Sign|c.Sign || out.iFunc != MF_NO_FUNC {
		t.Fatal("invalid merged metadata")
	}
	if Mf_SetCutIsContainedOrder(&out, b) != 1 || Mf_CutCheck(b, &out) != 0 {
		t.Fatal("incorrect containment")
	}
	if Mf_CutMergeOrder(a, b, &out, 3) != 0 {
		t.Fatal("accepted oversized cut")
	}
	var unit Mf_Cut_t
	Mf_CutCreateUnit(&unit, 65)
	if unit.Sign != 2 || Mf_CutCountBits(unit.Sign) != 1 || unit.iFunc != 2 {
		t.Fatal("invalid unit cut")
	}
}

func TestMfCutSelection(t *testing.T) {
	makeCut := func(flow float32, ids ...int32) *Mf_Cut_t {
		c := &Mf_Cut_t{Flow: flow, nLeaves: uint32(len(ids))}
		copy(c.pLeaves[:], ids)
		c.Sign = Mf_CutGetSign(ids, int32(len(ids)))
		return c
	}
	a, b, c := makeCut(3, 1, 2, 3), makeCut(2, 4, 5), makeCut(1, 1, 2)
	cuts := []*Mf_Cut_t{a, b, c}
	if Mf_SetLastCutIsContained(cuts, 2) != 0 {
		t.Fatal("candidate incorrectly contained")
	}
	if n := Mf_SetAddCut(cuts, 2, 4); n != 2 || cuts[0] != c || cuts[1] != b || cuts[2] != a {
		t.Fatalf("selection: n=%d cuts=%v", n, cuts)
	}
	if a.nLeaves != 31 {
		t.Fatal("dominated cut not invalidated")
	}
	d := makeCut(0, 1, 2, 6)
	cuts = []*Mf_Cut_t{c, b, d}
	if Mf_SetLastCutIsContained(cuts, 2) != 1 {
		t.Fatal("missed contained candidate")
	}
}

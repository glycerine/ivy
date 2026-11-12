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

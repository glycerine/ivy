package gia

import "testing"

func TestGiaManPackingBorrowAndReplaceLifecycle(t *testing.T) {
	p := Gia_ManStart(2)
	ci := Gia_ManAppendCi(p)
	if got := Gia_ManPacking(p); got != nil {
		t.Fatal("new manager unexpectedly has packing")
	}

	packing := Vec_IntAlloc(3)
	for _, value := range []int32{1, 1, ci >> 1} {
		Vec_IntPush(packing, value)
	}
	if err := p.ReplacePacking(packing); err != nil {
		t.Fatal(err)
	}
	if got := Gia_ManPacking(p); got != packing {
		t.Fatal("packing getter did not return the manager-owned vector")
	}
	// The getter is borrowed, matching direct C field access; manager and caller
	// see the same attached vector and its subsequent mutations.
	Vec_IntWriteEntry(Gia_ManPacking(p), 2, ci>>1)
	if Vec_IntEntry(p.VPacking, 2) != ci>>1 {
		t.Fatal("borrowed packing vector did not refer to attached state")
	}

	invalid := Vec_IntAlloc(1)
	Vec_IntPush(invalid, 1)
	if err := p.ReplacePacking(invalid); err == nil {
		t.Fatal("accepted invalid replacement packing")
	}
	if Gia_ManPacking(p) != packing {
		t.Fatal("invalid replacement changed attached packing")
	}
}

package cec

import "testing"

func TestCec4PatternPackingCompatibleAndConflicting(t *testing.T) {
	p := newCec4PatternPacker(4, 1)
	if !p.try(1, []int32{2, 5}) || !p.try(1, []int32{2, 7}) {
		t.Fatal("compatible assignments rejected")
	}
	if p.try(1, []int32{3, 4}) {
		t.Fatal("conflicting assignment accepted")
	}
	if p.bit(p.values, 1, 1) || !p.bit(p.values, 2, 1) || !p.bit(p.values, 3, 1) {
		t.Fatal("failed attempt changed packed assignment")
	}
	if n := p.add([]int32{3, 4}, true); n != 2 {
		t.Fatalf("expected one conflict before a free slot, visited %d", n)
	}
	if !p.bit(p.values, 1, p.cursor) || p.bit(p.values, 2, p.cursor) {
		t.Fatal("packed value mismatch")
	}
}

func TestCec4PatternPackingFillsAllUsableSlots(t *testing.T) {
	p := newCec4PatternPacker(7, 1)
	for pattern := 0; pattern < 63; pattern++ {
		literals := make([]int32, 6)
		for bit := 0; bit < 6; bit++ {
			literals[bit] = int32((bit+1)<<1 | ((pattern >> bit) & 1))
		}
		probes := p.add(literals, false)
		if pattern < 62 && probes != 1 {
			t.Fatalf("pattern %d: got %d probes", pattern, probes)
		}
		if pattern == 62 && probes != 64 {
			t.Fatalf("last slot: got %d probes", probes)
		}
		for _, lit := range literals {
			if !p.bit(p.reserved, lit>>1, p.cursor) || p.bit(p.values, lit>>1, p.cursor) != (lit&1 != 0) {
				t.Fatalf("pattern %d not stored in slot %d", pattern, p.cursor)
			}
		}
	}
}

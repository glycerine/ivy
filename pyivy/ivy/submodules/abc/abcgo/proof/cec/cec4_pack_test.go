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

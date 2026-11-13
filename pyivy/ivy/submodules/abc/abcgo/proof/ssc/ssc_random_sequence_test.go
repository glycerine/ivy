package ssc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// Values are from the C Gia_ManRandom and Ssc_Random{,1,2} formulas,
// starting after the reset call made by Ssc_PerformSweepingInt.
func TestRandomPiPatternsMatchCSequence(t *testing.T) {
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCi(g)
	s := NewSimulation(g, 3)
	if got := gia.Gia_ManRandom(1); got != 0x37681c96 {
		t.Fatalf("C reset step: got %08x", got)
	}
	Ssc_GiaRandomPiPattern(s, 3, nil)
	want := [2][3]uint64{
		{0xd896000bd3c8352a, 0xdf6c6323e1669096, 0xabbb4e1a7b4aabd6},
		{0x2eda4c53a942ec8a, 0xa586db96f018d645, 0x48ac0fdf34e92a8a},
	}
	for i := range want {
		for w := range want[i] {
			if got := s.Inputs[i][w]; got != want[i][w] {
				t.Fatalf("input %d word %d: got %016x want %016x", i, w, got, want[i][w])
			}
		}
	}
	// A second call continues the same stream, using Ssc_Random1 for the
	// first word and an unshifted Ssc_Random for the remaining word.
	Ssc_GiaRandomPiPattern(s, 2, []bool{false, true})
	pivot := [2][2]uint64{
		{0x4d500b7d332714de, 0xbd569fd720f9e335},
		{0xc88daa77854d8e8f, 0xf706df019f672d0b},
	}
	for i := range pivot {
		for w := range pivot[i] {
			if got := s.Inputs[i][w]; got != pivot[i][w] {
				t.Fatalf("pivot %d word %d: got %016x want %016x", i, w, got, pivot[i][w])
			}
		}
	}
}

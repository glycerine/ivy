package cba

import "testing"

func TestCbaBlastTableTruthUsesSourceIndexing(t *testing.T) {
	// In cbaBlast.c, with two fanins, minterm m reads pTable bits
	// 2*m and 2*m+1 and sets truth[m] if either bit is set.
	var table uint64 = (1 << 1) | (1 << 4)
	got := cbaBlastTableTruth([]uint64{table}, 2)
	want := []int32{1, 0, 1, 0}
	if len(got) != len(want) {
		t.Fatalf("truth length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("truth[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

package gia

import "testing"

func TestOfCutMergeAndDominance(t *testing.T) {
	a, b, c := ofUnitCut(1), ofUnitCut(3), ofUnitCut(5)
	var merged ofCut
	if !ofMergeCuts(&merged, 3, &a, &b, &c) || merged.size != 3 || merged.leaves[1] != 3 {
		t.Fatal("incorrect merged cut")
	}
	old := merged
	if ofMergeCuts(&merged, 2, &a, &b, &c) || merged != old {
		t.Fatal("overflow changed cut")
	}
	small := a
	small.delay = 1
	merged.delay = 2
	cuts := []*ofCut{&merged, &small, new(ofCut)}
	if n := ofAddCut(cuts, 1, 3); n != 1 || cuts[0] != &small {
		t.Fatal("failed to remove dominated cut")
	}
	*cuts[1] = old
	if !ofLastContained(cuts, 1) {
		t.Fatal("failed to detect contained cut")
	}
}

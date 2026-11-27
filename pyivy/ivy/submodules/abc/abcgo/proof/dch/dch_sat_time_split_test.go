package dch

import "testing"

// TestSATTimeSplitsLikeNative: Dch_NodesAreEquiv adds each solve's time to
// timeSat and to timeSatUnsat, timeSatSat or timeSatUndec by its outcome,
// which Dch_ManPrintStats reports; the port kept only the total and printed
// zeros for the split.
func TestSATTimeSplitsLikeNative(t *testing.T) {
	variants, _, _ := associativeSnapshots()
	_, stats, err := ComputeChoicesWithStats(Dch_DeriveTotalAig(variants), nil)
	if err != nil {
		t.Fatal(err)
	}
	if stats.SATCalls == 0 || stats.SolvingUnsat <= 0 {
		t.Fatalf("no unsat time: %+v", stats)
	}
	if split := stats.SolvingSat + stats.SolvingUnsat + stats.SolvingUndecided; split > stats.Solving {
		t.Fatalf("split %v exceeds total %v", split, stats.Solving)
	}
}

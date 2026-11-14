package saig

import (
	"math"
	"testing"
)

// saigSimFast.c passes (frames-prefix)/2 as an integer word count to
// Faig_ManComputeProbOne when estimating transition activity.
func TestFastTransitionOddFrameDenominatorMatchesSource(t *testing.T) {
	g := simulationFixture()
	state := g.Inputs()[1].NodeID()
	activity, err := SwitchingProbabilities(g, 3, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if activity[state] != 1 {
		t.Fatalf("three frames: activity=%v, want 1", activity[state])
	}
	activity, err = SwitchingProbabilities(g, 1, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(float64(activity[state])) {
		t.Fatalf("one frame: activity=%v, want C 0/0 NaN", activity[state])
	}
}

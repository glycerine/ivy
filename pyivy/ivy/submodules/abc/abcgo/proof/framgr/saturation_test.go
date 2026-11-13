package framgr

import "testing"

func TestSimulationSaturationWithNoNonconstantClasses(t *testing.T) {
	// ABC continues if constant candidates were refined even when there are
	// no nonconstant classes, and stops when there were no refinements.
	for _, tc := range []struct {
		changes int32
		classes int
		limit   float64
		want    bool
	}{
		{1, 0, 0.005, true},
		{0, 0, 0.005, false},
		{1, 100, 0.005, true},
		{1, 200, 0.005, false},
	} {
		if got := fraSimulationContinue(tc.changes, tc.classes, tc.limit); got != tc.want {
			t.Fatalf("changes=%d classes=%d: got %v, want %v", tc.changes, tc.classes, got, tc.want)
		}
	}
}

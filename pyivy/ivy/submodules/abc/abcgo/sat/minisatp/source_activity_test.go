package minisatp

import "testing"

func TestLearnedClauseActivityUsesSourceFloat32(t *testing.T) {
	s := newSolver(2)
	s.claInc = 1 + 1.0/(1<<25)
	s.learned([]int32{0, 2}, 0)
	if got := float64(s.learnts[0].activity); got != 1 {
		t.Fatalf("upstream stores activity as float32: got %.12g", got)
	}
}
func TestNewLearnedClauseIsNotRescaledBeforeInsertion(t *testing.T) {
	s := newSolver(2)
	s.claInc = 1e21
	s.learned([]int32{0, 2}, 0)
	if got, want := float64(s.learnts[0].activity), float64(float32(1e21)); got != want {
		t.Fatalf("upstream scales existing learnts only: got %g want %g", got, want)
	}
	if s.claInc != 10 {
		t.Fatalf("scaled increment = %g", s.claInc)
	}
}

// The upstream C++ probe shows that float32(1e20) is slightly above the
// double-precision 1e20 threshold, so this bump triggers a rescale.
func TestClauseActivityFloat32ThresholdTriggersRescale(t *testing.T) {
	s := newSolver(2)
	s.claInc = 1e20
	fresh := &clause{learnt: true}
	s.bumpClauseActivity(fresh)
	if got, want := fresh.activity, float32(1e20); got != want {
		t.Fatalf("new clause activity = %g, want %g", got, want)
	}
	if s.claInc != 1 {
		t.Fatalf("upstream rescales increment to 1, got %g", s.claInc)
	}
}

func TestClauseRescaleChangesExistingButNotNewClause(t *testing.T) {
	s := newSolver(2)
	existing := &clause{learnt: true, activity: float32(1e20)}
	s.learnts = append(s.learnts, existing)
	s.claInc = 1e20
	fresh := &clause{learnt: true}
	s.bumpClauseActivity(fresh)
	if existing.activity != 1 {
		t.Fatalf("upstream existing activity = 1, got %g", existing.activity)
	}
	if got, want := fresh.activity, float32(1e20); got != want {
		t.Fatalf("uninserted new activity = %g, want %g", got, want)
	}
	if s.claInc != 1 {
		t.Fatalf("rescaled increment = %g", s.claInc)
	}
}

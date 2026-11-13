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

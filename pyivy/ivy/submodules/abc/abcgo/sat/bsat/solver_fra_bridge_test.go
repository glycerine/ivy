package bsat

import "testing"

func TestSetLearntReductionPreservesUnspecifiedSettings(t *testing.T) {
	s := NewSolver()
	oldDelta, oldRatio := s.nLearntDelta, s.nLearntRatio
	s.SetLearntReduction(42, 0, 0)
	if s.nLearntStart != 42 || s.nLearntMax != 42 || s.nLearntDelta != oldDelta || s.nLearntRatio != oldRatio {
		t.Fatal("unexpected learned-reduction settings")
	}
	s.SetLearntReduction(0, 7, 30)
	if s.nLearntStart != 42 || s.nLearntMax != 42 || s.nLearntDelta != 7 || s.nLearntRatio != 30 {
		t.Fatal("independent learned-reduction settings were not retained")
	}
}

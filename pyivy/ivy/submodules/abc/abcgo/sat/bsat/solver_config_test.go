package bsat

import "testing"

func TestSetLearnedPolicy(t *testing.T) {
	s := NewSolver()
	s.SetLearnedPolicy(11, 7, 83)
	if s.nLearntStart != 11 || s.nLearntMax != 11 || s.nLearntDelta != 7 || s.nLearntRatio != 83 {
		t.Fatalf("learned policy = start %d max %d delta %d ratio %d", s.nLearntStart, s.nLearntMax, s.nLearntDelta, s.nLearntRatio)
	}
	s.SetLearnedPolicy(0, 0, 0)
	if s.nLearntStart != 0 || s.nLearntMax != 0 || s.nLearntDelta != 0 || s.nLearntRatio != 0 {
		t.Fatalf("zero learned policy = start %d max %d delta %d ratio %d", s.nLearntStart, s.nLearntMax, s.nLearntDelta, s.nLearntRatio)
	}
}

func TestSetLearnedPolicyRejectsNegativeValues(t *testing.T) {
	for _, policy := range [][3]int32{{-1, 0, 0}, {0, -1, 0}, {0, 0, -1}} {
		s := NewSolver()
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("SetLearnedPolicy%v did not panic", policy)
				}
			}()
			s.SetLearnedPolicy(policy[0], policy[1], policy[2])
		}()
	}
}

func TestSetNoRestarts(t *testing.T) {
	s := NewSolver()
	s.SetNoRestarts(true)
	if s.fNoRestarts != 1 {
		t.Fatalf("disabled restarts flag = %d, want 1", s.fNoRestarts)
	}
	s.SetNoRestarts(false)
	if s.fNoRestarts != 0 {
		t.Fatalf("enabled restarts flag = %d, want 0", s.fNoRestarts)
	}
}

package bsat

import "testing"

func TestRunStatsReturnsReadOnlySearchSnapshot(t *testing.T) {
	s := NewSolver()
	defer s.Close()
	s.stats = stats_t{starts: 7, conflicts: 11, decisions: 13, propagations: 17}
	got := s.RunStats()
	if got != (RunStatistics{Starts: 7, Conflicts: 11, Decisions: 13, Propagations: 17}) {
		t.Fatalf("RunStats() = %+v", got)
	}
	got.Starts = 100
	if s.stats.starts != 7 {
		t.Fatal("snapshot exposed mutable solver stats")
	}
}

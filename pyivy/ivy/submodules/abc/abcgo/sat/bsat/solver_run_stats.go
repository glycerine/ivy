package bsat

// RunStatistics is a read-only snapshot of the cumulative search counters used
// by the legacy ABC SAT diagnostics.
type RunStatistics struct {
	Starts       uint32
	Clauses      uint32
	Conflicts    int64
	Decisions    int64
	Propagations int64
}

// RunStats returns the current cumulative search counters without exposing the
// solver's mutable internal statistics structure.
func (s *sat_solver_t) RunStats() RunStatistics {
	return RunStatistics{
		Starts:       s.stats.starts,
		Clauses:      s.stats.clauses,
		Conflicts:    s.stats.conflicts,
		Decisions:    s.stats.decisions,
		Propagations: s.stats.propagations,
	}
}

package bmc

import "github.com/glycerine/abcgo/sat/bsat"

// transferBmcLearnedUnits mirrors the default SAT backend's post-UNSAT
// maintenance in Saig_ManBmcScalable: copy queued learned unit literals back
// into the solver, consume the queue, then compress to propagate them.
// SATOKO and Glucose do not expose this ABC-specific unit queue.
func transferBmcLearnedUnits(s *bsat.Solver) bool {
	for _, unit := range s.TakeUnitLiterals() {
		if !s.AddClause(unit) {
			return false
		}
	}
	return s.Compress()
}

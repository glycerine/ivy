package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

// cuddAnnealing reorders the selected variable-level interval with the core's
// simulated-annealing search. Candidate orders are built off-manager and the
// best one is committed once, preserving existing node handles.
func cuddAnnealing(dd *DdManager, lower, upper int32) int32 {
	return int32(core.Cudd_AnnealRange(dd, int(lower), int(upper), 0))
}

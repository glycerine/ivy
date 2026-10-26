package abc

// Owned checkpoint lifecycle from sat/bsat/satSolver.h:sat_solver_bookmark
// and satSolver.c:sat_solver_rollback.

import "slices"

type bsatBookmark struct{ state SatSolver }

// cloneState copies all mutable solver storage. The checkpoint itself is
// excluded so repeated rollback cannot alias or recursively copy bookmarks.
func (s *SatSolver) cloneState() SatSolver {
	r := *s
	r.bookmark = nil
	r.clauses = make([]bsatClause, len(s.clauses))
	for i, c := range s.clauses {
		r.clauses[i] = c
		r.clauses[i].lits = slices.Clone(c.lits)
	}
	r.watches = make([][]int, len(s.watches))
	for i, w := range s.watches {
		r.watches[i] = slices.Clone(w)
	}
	r.units = slices.Clone(s.units)
	r.assigns = slices.Clone(s.assigns)
	r.levels = slices.Clone(s.levels)
	r.reasons = slices.Clone(s.reasons)
	r.polarity = slices.Clone(s.polarity)
	r.used = slices.Clone(s.used)
	r.enqueueMask = slices.Clone(s.enqueueMask)
	r.activity = slices.Clone(s.activity)
	r.activityFactors = slices.Clone(s.activityFactors)
	r.activityFactorVariables = slices.Clone(s.activityFactorVariables)
	r.globalActivityVariables = slices.Clone(s.globalActivityVariables)
	r.order = slices.Clone(s.order)
	r.orderpos = slices.Clone(s.orderpos)
	r.seen = slices.Clone(s.seen)
	r.trail = slices.Clone(s.trail)
	r.trailLim = slices.Clone(s.trailLim)
	r.model = slices.Clone(s.model)
	r.lastConflict = slices.Clone(s.lastConflict)
	r.proofRootConflict = slices.Clone(s.proofRootConflict)
	r.proof = s.proof.clone()
	r.resolutionProof = s.resolutionProof.clone()
	return r
}

// Bookmark replaces the single active checkpoint after root propagation.
// Rollback restores its variables, clauses, learned prefix, watches, units and
// heuristic state. Unlike C's memory-arena checkpoint this uses owned snapshots
// so later learned-clause deletion cannot destroy saved clauses. Root UNSAT is
// representable. It is invalid to call either method from a solve callback.
func (s *SatSolver) Bookmark() error {
	if s == nil {
		return ErrDimension
	}
	if s.solving || len(s.clauses) == 0 {
		return ErrInvariant
	}
	s.prepareRoot(nil)
	s.bookmark = &bsatBookmark{state: s.cloneState()}
	return nil
}

// Rollback retains the bookmark for repeated use. Cumulative statistics remain
// cumulative for a nonempty variable prefix; an empty prefix resets statistics
// and the random seed, as native rollback does. Reset discards the bookmark.
func (s *SatSolver) Rollback() error {
	if s == nil {
		return ErrDimension
	}
	if s.solving || s.bookmark == nil {
		return ErrInvariant
	}
	b := s.bookmark
	stats := s.stats
	r := b.state.cloneState()
	r.bookmark = b
	if len(r.assigns) > 0 {
		r.stats = stats
	} else {
		r.stats = SatStats{}
		r.randomSeed = 91648253
	}
	*s = r
	return nil
}

// Ported to Go from ivy_core.py.

// Package core implements SAT-based core extraction: biased core selection
// and core minimization for unsatisfiable subsets of assumptions.
package core

// SatResult represents the result of a satisfiability check.
type SatResult int

const (
	Sat SatResult = iota
	Unsat
	Unknown
)

func (r SatResult) String() string {
	switch r {
	case Sat:
		return "sat"
	case Unsat:
		return "unsat"
	default:
		return "unknown"
	}
}

// Assumption is an opaque assumption literal used for incremental solving.
// Implementations should provide a way to identify assumptions (e.g., by ID).
type Assumption interface {
	// ID returns a unique identifier for this assumption.
	ID() int
}

// Solver is the interface for an incremental SAT/SMT solver that supports
// assumption-based checking and unsatisfiable core extraction.
type Solver interface {
	// Check checks satisfiability under the given assumptions.
	Check(assumptions []Assumption) SatResult

	// UnsatCore returns the unsatisfiable core from the last unsat Check.
	UnsatCore() []Assumption
}

// GetID returns the unique ID of an assumption.
func GetID(a Assumption) int {
	return a.ID()
}

// BiasedCore tries to produce a minimal unsatisfiable subset of alits,
// using as few of the assumptions in unlikely as possible.
// It first tries removing each unlikely assumption; if the result is still
// unsat, that assumption is dropped. Then it minimizes the remaining core.
func BiasedCore(s Solver, alits []Assumption, unlikely []Assumption) []Assumption {
	core := make([]Assumption, len(alits))
	copy(core, alits)

	unlikelyIDs := make(map[int]bool, len(unlikely))
	for _, lit := range unlikely {
		unlikelyIDs[GetID(lit)] = true
	}

	for _, lit := range unlikely {
		litID := GetID(lit)
		test := make([]Assumption, 0, len(core))
		for _, c := range core {
			if GetID(c) != litID {
				test = append(test, c)
			}
		}
		if s.Check(test) == Unsat {
			core = test
		}
	}

	result := s.Check(core)
	if result != Unsat {
		// Shouldn't happen if alits is truly unsat, but be defensive.
		return core
	}
	return MinimizeCore(s)
}

// MinimizeCore extracts and minimizes the unsatisfiable core from the solver.
// The solver must have just returned Unsat from a Check call.
func MinimizeCore(s Solver) []Assumption {
	core := s.UnsatCore()
	cp := make([]Assumption, len(core))
	copy(cp, core)
	return minimizeCoreAux(s, cp)
}

// minimizeCoreAux implements the iterative MUS (Minimal Unsatisfiable Subset)
// extraction algorithm. It tries removing each assumption in turn; if the
// remainder is still unsat, the assumption is redundant. If sat, the
// assumption is necessary and kept.
func minimizeCoreAux(s Solver, core []Assumption) []Assumption {
	var mus []Assumption
	ids := make(map[int]bool)

	for len(core) > 0 {
		c := core[0]
		// Try without c
		newCore := make([]Assumption, 0, len(mus)+len(core)-1)
		newCore = append(newCore, mus...)
		newCore = append(newCore, core[1:]...)

		result := s.Check(newCore)
		if result == Sat {
			// c is necessary
			mus = append(mus, c)
			ids[GetID(c)] = true
			core = core[1:]
		} else {
			// c is redundant; use the new unsat core (filtering out already-known MUS members)
			unsatCore := s.UnsatCore()
			core = make([]Assumption, 0, len(unsatCore))
			for _, a := range unsatCore {
				if !ids[GetID(a)] {
					core = append(core, a)
				}
			}
		}
	}
	return mus
}

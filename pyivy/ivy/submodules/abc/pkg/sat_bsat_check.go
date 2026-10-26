package abc

// Native invariant checks for the port of src/sat/bsat/satSolver.c.

import (
	"fmt"
	"slices"
)

// Check validates owned clauses, watches, assignment trail, reasons and the
// variable-order heap. It is diagnostic and does not certify learned proofs.
func (s *SatSolver) Check() error {
	fail := func(message string) error { return fmt.Errorf("%w: SAT %s", ErrInvariant, message) }
	n := len(s.assigns)
	if len(s.clauses) == 0 || len(s.watches) != 2*n || len(s.levels) != n || len(s.reasons) != n || len(s.polarity) != n || len(s.used) != n || len(s.activity) != n || len(s.orderpos) != n || len(s.seen) != n {
		return fail("array dimensions")
	}
	if s.qhead < 0 || s.qhead > len(s.trail) {
		return fail("propagation queue")
	}
	if (s.lastStatus == SatSatisfiable && (s.model == nil || len(s.model) != n)) || (s.lastStatus != SatSatisfiable && s.model != nil) || (s.lastStatus != SatUnsatisfiable && len(s.lastConflict) != 0) {
		return fail("retained result")
	}
	for _, literal := range s.lastConflict {
		if !literal.Valid() || uint64(literal.Var()) >= uint64(n) {
			return fail("final conflict literal")
		}
	}
	last := 0
	for _, bound := range s.trailLim {
		if bound < last || bound > len(s.trail) {
			return fail("decision boundaries")
		}
		last = bound
	}
	assigned := make([]bool, n)
	for _, l := range s.trail {
		if !l.Valid() || uint64(l.Var()) >= uint64(n) || assigned[l.Var()] || s.literalValue(l) != 1 {
			return fail("assignment trail")
		}
		assigned[l.Var()] = true
	}
	for v, value := range s.assigns {
		if value < -1 || value > 1 || (value >= 0) != assigned[v] || s.levels[v] < 0 || s.levels[v] > len(s.trailLim) || s.seen[v] {
			return fail("assignment metadata")
		}
		if value < 0 && (s.reasons[v] != 0 || s.orderpos[v] < 0) {
			return fail("unassigned variable")
		}
		if reason := s.reasons[v]; reason != 0 {
			if value < 0 || reason < 0 || reason >= len(s.clauses) || s.clauses[reason].deleted {
				return fail("reason handle")
			}
			contains := false
			for _, l := range s.clauses[reason].lits {
				if int(l.Var()) == v && s.literalValue(l) == 1 {
					contains = true
				}
			}
			if !contains {
				return fail("reason literal")
			}
		}
	}
	watchCount := make([][2]int, len(s.clauses))
	for l, bucket := range s.watches {
		for _, id := range bucket {
			if id <= 0 || id >= len(s.clauses) {
				return fail("watch handle")
			}
			c := s.clauses[id]
			if c.deleted {
				continue
			}
			if len(c.lits) < 2 {
				return fail("unit watch")
			}
			if c.lits[0].Not() == SatLit(l) {
				watchCount[id][0]++
			} else if c.lits[1].Not() == SatLit(l) {
				watchCount[id][1]++
			} else {
				return fail("watch literal")
			}
		}
	}
	learnt := 0
	units := make([]bool, len(s.clauses))
	for _, id := range s.units {
		if id <= 0 || id >= len(s.clauses) || len(s.clauses[id].lits) != 1 || units[id] {
			return fail("unit index")
		}
		units[id] = true
	}
	for id, c := range s.clauses {
		if id == 0 || c.deleted {
			continue
		}
		if len(c.lits) == 0 {
			return fail("empty arena clause")
		}
		if c.learnt {
			learnt++
		}
		if len(c.lits) == 1 {
			if !units[id] {
				return fail("missing unit index")
			}
		} else if c.unwatched {
			if watchCount[id] != [2]int{} {
				return fail("pruned watch multiplicity")
			}
			satisfied := false
			for _, literal := range c.lits {
				if s.literalValue(literal) == 1 && s.levels[literal.Var()] == 0 {
					satisfied = true
					break
				}
			}
			if !satisfied {
				return fail("pruned clause is not root-satisfied")
			}
		} else if watchCount[id] != [2]int{1, 1} {
			return fail("watch multiplicity")
		}
		sorted := slices.Clone(c.lits)
		slices.Sort(sorted)
		for i, l := range sorted {
			if !l.Valid() || uint64(l.Var()) >= uint64(n) {
				return fail("clause literal")
			}
			if i > 0 && l.Var() == sorted[i-1].Var() {
				return fail("clause repeats variable")
			}
		}
	}
	if learnt != s.learntCount {
		return fail("learned count")
	}
	if s.proof != nil {
		if s.proof.manager == nil || s.proof.first != 0 || s.proof.next != learnt || s.proof.final < -1 {
			return fail("proof state")
		}
		records := s.proof.next
		if s.proof.final >= 0 {
			if s.proof.final != s.proof.next {
				return fail("final proof ID")
			}
			records++
		}
		if stats := s.proof.manager.Stats(); stats.FirstClause != s.proof.first || stats.Records != records || stats.Active {
			return fail("proof records")
		}
		proofIDs := make([]bool, learnt)
		for id := 1; id < len(s.clauses); id++ {
			clause := s.clauses[id]
			if clause.deleted || !clause.learnt {
				continue
			}
			if clause.proofID < 0 || clause.proofID >= learnt || proofIDs[clause.proofID] {
				return fail("learned proof ID")
			}
			proofIDs[clause.proofID] = true
		}
	}
	if s.resolutionProof != nil {
		proof := s.resolutionProof
		if proof.dag == nil || len(proof.roots) == 0 || len(proof.rootPartA) != len(proof.roots) || len(proof.variablePart) != n || proof.final < -1 {
			return fail("resolution proof state")
		}
		if err := proof.dag.Check(); err != nil {
			return fail("resolution proof DAG")
		}
		if proof.final >= proof.dag.Stats().Nodes {
			return fail("resolution final proof")
		}
		for id := 1; id < len(s.clauses); id++ {
			clause := s.clauses[id]
			if clause.deleted || !clause.learnt {
				continue
			}
			if clause.proofNode < 0 || clause.proofNode >= proof.dag.Stats().Nodes {
				return fail("learned resolution proof")
			}
		}
	}
	present := make([]bool, n)
	for index, v := range s.order {
		if v < 0 || v >= n || present[v] || s.orderpos[v] != index {
			return fail("heap positions")
		}
		present[v] = true
		if index > 0 && s.activity[s.order[(index-1)/2]] < s.activity[v] {
			return fail("heap activity order")
		}
	}
	for v, pos := range s.orderpos {
		if (pos >= 0) != present[v] {
			return fail("heap membership")
		}
	}
	return nil
}

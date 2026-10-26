package abc

// Proof-information attachment for src/sat/bsat/satSolver2.c pPrf2 mode.

import "slices"

type satSolverProofInfo struct {
	manager         *SatProofInfoManager
	first, next     int
	final           int
	lastAssumptions []SatLit
}

func (p *satSolverProofInfo) clone() *satSolverProofInfo {
	if p == nil {
		return nil
	}
	clone := *p
	clone.manager = p.manager.clone()
	clone.lastAssumptions = slices.Clone(p.lastAssumptions)
	return &clone
}

func (s *SatSolver) attachProofInfo(idToProof []int, width int, limits SatProofInfoLimits) error {
	if s == nil || s.solving || s.proof != nil || s.learntCount != 0 || width <= 0 {
		return ErrInvariant
	}
	manager, err := NewSatProofInfoManager(limits)
	if err != nil {
		return err
	}
	if err = manager.Restart(idToProof, 0, width); err != nil {
		return err
	}
	s.proof = &satSolverProofInfo{manager: manager, first: 0, final: -1}
	return nil
}

func (s *SatSolver) proofBeginSolve(assumptions []SatLit) error {
	if s.proof == nil {
		return nil
	}
	p := s.proof
	if p.manager == nil || p.final < -1 {
		return ErrInvariant
	}
	if p.final >= 0 {
		if err := p.manager.Shrink(p.final); err != nil {
			return err
		}
		p.final = -1
	}
	p.lastAssumptions = slices.Clone(assumptions)
	return nil
}

// proofClauseReferences conservatively collects every clause in the current
// implication subgraph. Extra antecedents only enlarge the root core; they do
// not weaken its independently checked UNSAT guarantee.
func (s *SatSolver) proofClauseReferences(initial []int) ([]SatProofClauseRef, error) {
	seen := make([]bool, len(s.clauses))
	stack := slices.Clone(initial)
	references := make([]SatProofClauseRef, 0, len(initial))
	for len(stack) > 0 {
		last := len(stack) - 1
		id := stack[last]
		stack = stack[:last]
		if id == 0 {
			continue
		}
		if id < 0 || id >= len(s.clauses) || s.clauses[id].deleted {
			return nil, ErrInvariant
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		clause := s.clauses[id]
		if clause.learnt {
			if clause.proofID < s.proof.first || clause.proofID >= s.proof.next {
				return nil, ErrInvariant
			}
			references = append(references, SatProofClauseRef{Learned: true, ID: clause.proofID})
		} else {
			if clause.proofRoot < 0 {
				return nil, ErrInvariant
			}
			references = append(references, SatProofClauseRef{ID: clause.proofRoot})
		}
		for index := len(clause.lits) - 1; index >= 0; index-- {
			variable := clause.lits[index].Var()
			if uint64(variable) >= uint64(len(s.reasons)) {
				return nil, ErrInvariant
			}
			if reason := s.reasons[variable]; reason != 0 && reason != id {
				stack = append(stack, reason)
			}
		}
	}
	return references, nil
}

func (s *SatSolver) proofRecord(initial []int, final bool) (int, error) {
	if s.proof == nil {
		return -1, nil
	}
	references, err := s.proofClauseReferences(initial)
	if err != nil {
		return -1, err
	}
	managerBefore := s.proof.manager.clone()
	reference := SatProofClauseRef{ID: -1}
	if len(references) > 0 {
		reference = references[0]
	}
	if err = s.proof.manager.ChainStart(reference); err == nil {
		for index := 1; index < len(references); index++ {
			next := references[index]
			if err = s.proof.manager.ChainResolve(next); err != nil {
				break
			}
		}
	}
	if err == nil {
		_, err = s.proof.manager.ChainStop()
	}
	if err != nil {
		s.proof.manager = managerBefore
		return -1, err
	}
	id := s.proof.next
	if size := s.proof.manager.Size(); size != id-s.proof.first+1 {
		s.proof.manager = managerBefore
		return -1, ErrInvariant
	}
	if final {
		s.proof.final = id
	} else {
		s.proof.next++
	}
	return id, nil
}

func (s *SatSolver) proofRecordLearned(conflict int) (int, error) {
	return s.proofRecord([]int{conflict}, false)
}

func (s *SatSolver) proofRecordFinal(conflicts ...int) error {
	if s.proof == nil {
		return nil
	}
	_, err := s.proofRecord(conflicts, true)
	return err
}

func (s *SatSolver) proofCompactLearned(remove []bool, oldToNew []int, clauses []bsatClause) error {
	if s.proof == nil {
		return nil
	}
	if s.proof.final >= 0 || len(remove) != len(s.clauses) || len(oldToNew) != len(s.clauses) {
		return ErrInvariant
	}
	managerBefore := s.proof.manager.clone()
	next := 0
	for oldID := 1; oldID < len(s.clauses); oldID++ {
		clause := s.clauses[oldID]
		if !clause.learnt || clause.deleted || remove[oldID] {
			continue
		}
		if clause.proofID < s.proof.first || clause.proofID >= s.proof.next || oldToNew[oldID] <= 0 || oldToNew[oldID] >= len(clauses) {
			return ErrInvariant
		}
		if err := s.proof.manager.AddSaved(clause.proofID, next); err != nil {
			s.proof.manager = managerBefore
			return err
		}
		clauses[oldToNew[oldID]].proofID = next
		next++
	}
	if err := s.proof.manager.Compact(0); err != nil {
		s.proof.manager = managerBefore
		return err
	}
	s.proof.first, s.proof.next = 0, next
	return nil
}

// EnableUnsatCoreTracking attaches the concrete proof-information manager.
// The complete formula must already be loaded, and no solve may have started.
func (s *SatSolver2) EnableUnsatCoreTracking(limits SatProofInfoLimits) error {
	if !s.usable() || s.called || s.core.proof != nil || s.bookmark != nil || s.core.bookmark != nil || len(s.clauses) <= 1 {
		return ErrInvariant
	}
	mapping := make([]int, len(s.clauses))
	mapping[0] = -1
	for handle := 1; handle < len(mapping); handle++ {
		mapping[handle] = handle - 1
	}
	return s.core.attachProofInfo(mapping, len(s.clauses)-1, limits)
}

// UnsatCore returns facade clause handles in the proof manager's descending
// order. It refuses to expose a core unless a fresh Go solver independently
// proves it UNSAT under the assumptions from the last solve.
func (s *SatSolver2) UnsatCore() ([]int, error) {
	if !s.usable() || s.core.lastStatus != SatUnsatisfiable || s.core.proof == nil || s.core.proof.final < 0 {
		return nil, ErrInvariant
	}
	core, err := s.core.proof.manager.UnsatCore()
	if err != nil {
		return nil, err
	}
	if err = s.verifyClauseHandleCore(core, s.core.proof.lastAssumptions); err != nil {
		return nil, err
	}
	return core, nil
}

func (s *SatSolver2) verifyClauseHandleCore(core []int, assumptions []SatLit) error {
	verifier, err := NewSatSolverWithLimit(max(1, s.core.VariableCount()))
	if err != nil {
		return err
	}
	if err = verifier.SetVariableCount(s.core.VariableCount()); err != nil {
		return err
	}
	for _, handle := range core {
		if handle <= 0 || handle >= len(s.clauses) {
			return ErrInvariant
		}
		if _, err = verifier.AddClause(s.clauses[handle].Literals); err != nil {
			return err
		}
	}
	result, err := verifier.Solve(assumptions, SatLimits{})
	if err != nil || result.Status != SatUnsatisfiable {
		return ErrInvariant
	}
	return nil
}

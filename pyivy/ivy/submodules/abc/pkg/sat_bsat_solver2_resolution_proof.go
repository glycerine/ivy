package abc

// Resolution-proof attachment for src/sat/bsat/satSolver2.c pPrf1 mode.

import "slices"

type satSolverResolutionProof struct {
	dag             *SatProofDAG
	roots           [][]SatLit
	rootPartA       []bool
	variablePart    []bool
	final           int
	lastAssumptions []SatLit
}

// SatResolutionCoreClause is one checked learned clause in the reachable
// resolution cone. Node is the stable proof-DAG handle for this solve. Final
// distinguishes the terminal conflict clause from ordinary learned clauses.
type SatResolutionCoreClause struct {
	Node     int
	Literals []SatLit
	Final    bool
}

func (p *satSolverResolutionProof) clone() *satSolverResolutionProof {
	if p == nil {
		return nil
	}
	clone := *p
	clone.dag = p.dag.clone()
	clone.roots = make([][]SatLit, len(p.roots))
	for index := range p.roots {
		clone.roots[index] = slices.Clone(p.roots[index])
	}
	clone.variablePart = slices.Clone(p.variablePart)
	clone.rootPartA = slices.Clone(p.rootPartA)
	clone.lastAssumptions = slices.Clone(p.lastAssumptions)
	return &clone
}

func canonicalSatClause(clause []SatLit) []SatLit {
	result := slices.Clone(clause)
	slices.Sort(result)
	return result
}

func equalSatProofClauses(first, second []SatLit) bool {
	return slices.Equal(canonicalSatClause(first), canonicalSatClause(second))
}

func resolveSatLiteralClauses(first, second []SatLit) ([]SatLit, error) {
	left := make([]int, len(first))
	right := make([]int, len(second))
	for index, literal := range first {
		left[index] = literal.DIMACS()
	}
	for index, literal := range second {
		right[index] = literal.DIMACS()
	}
	resolved, err := ResolveSatProofClauses(left, right)
	if err != nil {
		return nil, err
	}
	result := make([]SatLit, len(resolved))
	for index, value := range resolved {
		result[index], err = ParseSatDIMACSLiteral(value)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *SatSolver) resolutionProofClause(id int) ([]SatLit, SatProofDAGRef, error) {
	if s.resolutionProof == nil || id <= 0 || id >= len(s.clauses) || s.clauses[id].deleted {
		return nil, SatProofDAGRef{}, ErrInvariant
	}
	clause := s.clauses[id]
	if clause.learnt {
		if clause.proofNode < 0 || clause.proofNode >= s.resolutionProof.dag.Stats().Nodes {
			return nil, SatProofDAGRef{}, ErrInvariant
		}
		reference, err := MakeSatProofDAGNodeRef(clause.proofNode, false)
		return slices.Clone(clause.lits), reference, err
	}
	root := clause.proofRoot - 1
	if root < 0 || root >= len(s.resolutionProof.roots) {
		return nil, SatProofDAGRef{}, ErrInvariant
	}
	reference, err := MakeSatProofDAGClauseRef(root, false)
	return slices.Clone(s.resolutionProof.roots[root]), reference, err
}

func (s *SatSolver) resolutionProofBeginSolve(assumptions []SatLit) error {
	if s.resolutionProof == nil {
		return nil
	}
	if s.resolutionProof.dag == nil || s.resolutionProof.dag.stopped {
		return ErrInvariant
	}
	s.resolutionProof.final = -1
	s.resolutionProof.lastAssumptions = slices.Clone(assumptions)
	return nil
}

// resolutionProofRecord reconstructs and checks an exact resolution chain
// from the live implication graph. It commits a DAG node only after the chain
// reaches exactly target.
func (s *SatSolver) resolutionProofRecord(target []SatLit, conflict int, final bool) (int, error) {
	if s.resolutionProof == nil {
		return -1, nil
	}
	if conflict == 0 {
		if final {
			s.resolutionProof.final = -1
		}
		return -1, nil
	}
	current, first, err := s.resolutionProofClause(conflict)
	if err != nil {
		return -1, err
	}
	references := []SatProofDAGRef{first}
	trailPosition := make([]int, len(s.assigns))
	for index := range trailPosition {
		trailPosition[index] = -1
	}
	for index, literal := range s.trail {
		trailPosition[literal.Var()] = index
	}
	targetSet := make(map[SatLit]struct{}, len(target))
	for _, literal := range target {
		targetSet[literal] = struct{}{}
	}
	resolvedVariables := make([]bool, len(s.assigns))
	for !equalSatProofClauses(current, target) {
		pivotVariable := -1
		pivotPosition := -1
		for _, literal := range current {
			if _, retained := targetSet[literal]; retained {
				continue
			}
			variable := int(literal.Var())
			if variable < 0 || variable >= len(s.reasons) || s.reasons[variable] == 0 || resolvedVariables[variable] {
				continue
			}
			if trailPosition[variable] > pivotPosition {
				pivotVariable, pivotPosition = variable, trailPosition[variable]
			}
		}
		if pivotVariable < 0 {
			return -1, ErrInvariant
		}
		resolvedVariables[pivotVariable] = true
		reason := s.reasons[pivotVariable]
		reasonClause, reference, reasonErr := s.resolutionProofClause(reason)
		if reasonErr != nil {
			return -1, reasonErr
		}
		current, err = resolveSatLiteralClauses(current, reasonClause)
		if err != nil {
			return -1, err
		}
		reference.PivotA = pivotVariable < len(s.resolutionProof.variablePart) && s.resolutionProof.variablePart[pivotVariable]
		references = append(references, reference)
	}
	node, err := s.resolutionProof.dag.AddNode(SatProofDAGNode{Learned: !final, Antecedents: references})
	if err != nil {
		return -1, err
	}
	if final {
		s.resolutionProof.final = node
	}
	return node, nil
}

func (s *SatSolver) resolutionProofCompact(remove []bool, oldToNew []int, clauses []bsatClause) error {
	if s.resolutionProof == nil {
		return nil
	}
	p := s.resolutionProof
	if p.final >= 0 || len(remove) != len(s.clauses) || len(oldToNew) != len(s.clauses) {
		return ErrInvariant
	}
	roots := make([]int, 0, s.learntCount)
	oldClauseIDs := make([]int, 0, s.learntCount)
	for oldID := 1; oldID < len(s.clauses); oldID++ {
		clause := s.clauses[oldID]
		if !clause.learnt || clause.deleted || remove[oldID] {
			continue
		}
		if clause.proofNode < 0 || oldToNew[oldID] <= 0 || oldToNew[oldID] >= len(clauses) {
			return ErrInvariant
		}
		roots = append(roots, clause.proofNode)
		oldClauseIDs = append(oldClauseIDs, oldID)
	}
	if len(roots) == 0 {
		p.dag.nodes = nil
		p.dag.antecedents = 0
		return nil
	}
	reduction, err := p.dag.Reduce(roots, roots[0])
	if err != nil {
		return err
	}
	for index, oldID := range oldClauseIDs {
		clauses[oldToNew[oldID]].proofNode = reduction.Roots[index]
	}
	return nil
}

// EnableResolutionProofTracking attaches exact checked resolution logging.
// The complete facade formula must be loaded and no solve may have started.
func (s *SatSolver2) EnableResolutionProofTracking(limits SatProofDAGLimits) error {
	if !s.usable() || s.called || s.core.resolutionProof != nil || s.bookmark != nil || s.core.bookmark != nil || len(s.clauses) <= 1 {
		return ErrInvariant
	}
	dag, err := NewSatProofDAG(limits)
	if err != nil {
		return err
	}
	if len(s.clauses)-2 > dag.limits.MaxClauseID {
		return ErrLimit
	}
	roots := make([][]SatLit, len(s.clauses)-1)
	rootPartA := make([]bool, len(s.clauses)-1)
	for handle := 1; handle < len(s.clauses); handle++ {
		roots[handle-1] = slices.Clone(s.clauses[handle].Literals)
		rootPartA[handle-1] = s.clauses[handle].PartA
	}
	s.core.resolutionProof = &satSolverResolutionProof{
		dag:          dag,
		roots:        roots,
		rootPartA:    rootPartA,
		variablePart: slices.Clone(s.variablePart),
		final:        -1,
	}
	return nil
}

func (s *SatSolver2) resolutionProofStore() (*SatProofStore, error) {
	final, err := s.VerifyResolutionProof()
	if err != nil || len(final) != 0 || len(s.core.resolutionProof.lastAssumptions) != 0 {
		return nil, ErrInvariant
	}
	p := s.core.resolutionProof
	if len(p.rootPartA) != len(p.roots) {
		return nil, ErrInvariant
	}
	used, err := p.dag.CollectUsedSorted([]int{p.final})
	if err != nil {
		return nil, err
	}
	core, err := p.dag.CoreFromUsed(used)
	if err != nil {
		return nil, err
	}
	usedRoot := make([]bool, len(p.roots))
	for _, root := range core {
		if root < 0 || root >= len(usedRoot) {
			return nil, ErrInvariant
		}
		usedRoot[root] = true
	}
	store, err := NewSatProofStore(SatProofStoreLimits{})
	if err != nil {
		return nil, err
	}
	for root := range p.roots {
		if usedRoot[root] && p.rootPartA[root] {
			if _, err = store.AddClause(p.roots[root]); err != nil {
				return nil, err
			}
		}
	}
	if err = store.MarkClausesA(); err != nil {
		return nil, err
	}
	for root := range p.roots {
		if usedRoot[root] && !p.rootPartA[root] {
			if _, err = store.AddClause(p.roots[root]); err != nil {
				return nil, err
			}
		}
	}
	if err = store.MarkRoots(); err != nil {
		return nil, err
	}
	resolved, err := p.dag.Verify(used, p.roots)
	if err != nil || len(resolved) != len(used) {
		return nil, ErrInvariant
	}
	for _, clause := range resolved {
		if _, err = store.AddClause(clause); err != nil {
			return nil, err
		}
	}
	if stats := store.Stats(); stats.EmptyClause != stats.Clauses-1 {
		return nil, ErrInvariant
	}
	return store, nil
}

// ProofAigInterpolant derives a checked Craig interpolant from the attached
// exact resolution proof. Assumption proofs are rejected because assumptions
// are not partition roots.
func (s *SatSolver2) ProofAigInterpolant(options SatAigInterpolationOptions) (SatAigInterpolation, error) {
	store, err := s.resolutionProofStore()
	if err != nil {
		return SatAigInterpolation{}, err
	}
	interpolator, err := NewSatAigInterpolator(options)
	if err != nil {
		return SatAigInterpolation{}, err
	}
	return interpolator.Interpolate(store)
}

// ProofTruthInterpolant derives the corresponding bounded truth table.
func (s *SatSolver2) ProofTruthInterpolant(options SatTruthInterpolationOptions) (SatTruthInterpolation, error) {
	store, err := s.resolutionProofStore()
	if err != nil {
		return SatTruthInterpolation{}, err
	}
	interpolator, err := NewSatTruthInterpolator(options)
	if err != nil {
		return SatTruthInterpolation{}, err
	}
	return interpolator.Interpolate(store)
}

func (s *SatSolver2) ResolutionProofStats() (SatProofDAGStats, error) {
	if !s.usable() || s.core.resolutionProof == nil {
		return SatProofDAGStats{}, ErrInvariant
	}
	return s.core.resolutionProof.dag.Stats(), nil
}

// VerifyResolutionProof replays the complete reachable DAG using only copied
// facade roots and returns the final derived clause.
func (s *SatSolver2) VerifyResolutionProof() ([]SatLit, error) {
	if !s.usable() || s.core.lastStatus != SatUnsatisfiable || s.core.resolutionProof == nil || s.core.resolutionProof.final < 0 {
		return nil, ErrInvariant
	}
	p := s.core.resolutionProof
	resolved, err := p.dag.Verify([]int{p.final}, p.roots)
	if err != nil || len(resolved) != 1 || !equalSatProofClauses(resolved[0], s.core.lastConflict) {
		return nil, ErrInvariant
	}
	return slices.Clone(resolved[0]), nil
}

// ResolutionUnsatCore derives exact reachable problem roots from the checked
// resolution DAG and independently proves the resulting facade subset UNSAT.
func (s *SatSolver2) ResolutionUnsatCore() ([]int, error) {
	if _, err := s.VerifyResolutionProof(); err != nil {
		return nil, err
	}
	p := s.core.resolutionProof
	zeroBased, err := p.dag.DeriveCore(p.final)
	if err != nil {
		return nil, err
	}
	handles := make([]int, len(zeroBased))
	for index, root := range zeroBased {
		handles[index] = root + 1
	}
	if err = s.verifyClauseHandleCore(handles, p.lastAssumptions); err != nil {
		return nil, err
	}
	return handles, nil
}

// ResolutionLearnedCore returns every learned clause reachable from the final
// conflict, ordered by proof-DAG handle. Each returned clause is reconstructed
// by the stand-alone resolution checker before it is exposed. The final
// conflict is included, matching the learned-core convention used by
// Intp_ManUnsatCore with fLearned set.
func (s *SatSolver2) ResolutionLearnedCore() ([]SatResolutionCoreClause, error) {
	if _, err := s.VerifyResolutionProof(); err != nil {
		return nil, err
	}
	p := s.core.resolutionProof
	used, err := p.dag.CollectUsedSorted([]int{p.final})
	if err != nil {
		return nil, err
	}
	resolved, err := p.dag.Verify(used, p.roots)
	if err != nil || len(resolved) != len(used) {
		return nil, ErrInvariant
	}
	result := make([]SatResolutionCoreClause, len(used))
	for index, handle := range used {
		node, nodeErr := p.dag.Node(handle)
		if nodeErr != nil {
			return nil, nodeErr
		}
		result[index] = SatResolutionCoreClause{
			Node:     handle,
			Literals: slices.Clone(resolved[index]),
			Final:    handle == p.final,
		}
		if result[index].Final == node.Learned {
			return nil, ErrInvariant
		}
	}
	return result, nil
}

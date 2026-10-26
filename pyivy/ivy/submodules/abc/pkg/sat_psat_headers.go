package abc

// Owned compatibility surface for src/sat/psat/m114p.h and m114p_types.h.

import (
	"slices"
	"strconv"
	"strings"
	"time"
)

// M114PSolverOptions bounds the concrete replacement for the integer solver
// handle declared by M114p_Solver_t. Zero limits select package defaults.
type M114PSolverOptions struct {
	MaxVariables, MaxClauses           int
	MaxProofNodes, MaxProofAntecedents int
	Stop                               func() bool
}

// M114PProofChain is one reachable resolution chain. ClauseIDs are stable,
// zero-based IDs in the accompanying snapshot: roots first, then chains.
// Variables contains one pivot for every resolution after the first clause.
type M114PProofChain struct {
	ClauseIDs []int
	Variables []SatVar
	Clause    []SatLit
}

// M114PProofSnapshot is an owned replacement for the four pointer-returning
// proof traversal functions and the filename-based proof saver.
type M114PProofSnapshot struct {
	Roots       [][]SatLit
	Chains      []M114PProofChain
	FinalClause int
}

// M114PSolver owns one concrete solver and must not be copied or used
// concurrently. Close replaces M114p_SolverDelete.
type M114PSolver struct {
	solver        *SatSolver2
	options       M114PSolverOptions
	recordProof   bool
	proofAttached bool
	closed        bool
	forcedUnsat   bool
	rootCursor    int
	chainCursor   int
	proof         *M114PProofSnapshot
}

func normalizeM114POptions(options M114PSolverOptions) (M114PSolverOptions, error) {
	if options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxProofNodes < 0 || options.MaxProofAntecedents < 0 {
		return options, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 16 << 20
	}
	if options.MaxProofNodes == 0 {
		options.MaxProofNodes = 16 << 20
	}
	if options.MaxProofAntecedents == 0 {
		options.MaxProofAntecedents = 64 << 20
	}
	return options, nil
}

func NewM114PSolver(recordProof bool, options M114PSolverOptions) (*M114PSolver, error) {
	options, err := normalizeM114POptions(options)
	if err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	solver, err := NewSatSolver2(SatSolver2Options{MaxVariables: options.MaxVariables, MaxClauses: options.MaxClauses})
	if err != nil {
		return nil, err
	}
	return &M114PSolver{solver: solver, options: options, recordProof: recordProof}, nil
}

func (s *M114PSolver) usable() bool {
	return s != nil && !s.closed && s.solver != nil && s.solver.usable()
}

func (s *M114PSolver) Close() {
	if s == nil || s.closed {
		return
	}
	s.closed = true
	if s.solver != nil {
		s.solver.Stop()
	}
	s.solver = nil
	s.proof = nil
}

func (s *M114PSolver) SetVariableCount(count int) error {
	if !s.usable() || s.proofAttached {
		return ErrInvariant
	}
	if count < s.solver.VariableCount() {
		return ErrDimension
	}
	return s.solver.SetVariableCount(count)
}

// AddClause preserves the integer API's consistency result while returning a
// typed error for malformed input. Input storage is copied by the solver.
func (s *M114PSolver) AddClause(literals []SatLit) (bool, error) {
	if !s.usable() || s.proofAttached {
		return false, ErrInvariant
	}
	if s.options.Stop != nil && s.options.Stop() {
		return false, ErrCancelled
	}
	if len(literals) == 0 {
		s.forcedUnsat = true
		s.proof = nil
		return false, nil
	}
	handle, err := s.solver.AddClause(literals, s.solver.ClauseCount()+1)
	if err != nil {
		// SatSolver2 commits atomically. At this point ErrInvariant means the
		// new root clause made the accumulated formula inconsistent.
		if err == ErrInvariant {
			s.forcedUnsat = true
			return false, nil
		}
		return false, err
	}
	return handle > 0, nil
}

func (s *M114PSolver) Simplify() (bool, error) {
	if !s.usable() || s.proofAttached {
		return false, ErrInvariant
	}
	if s.forcedUnsat {
		return false, nil
	}
	if s.options.Stop != nil && s.options.Stop() {
		return false, ErrCancelled
	}
	return s.solver.Simplify()
}

func (s *M114PSolver) attachProof() error {
	if !s.recordProof || s.proofAttached {
		return nil
	}
	if s.solver.ClauseCount() == 0 {
		return ErrInvariant
	}
	err := s.solver.EnableResolutionProofTracking(SatProofDAGLimits{
		MaxNodes:       s.options.MaxProofNodes,
		MaxAntecedents: s.options.MaxProofAntecedents,
		MaxClauseID:    s.options.MaxClauses,
		Stop:           s.options.Stop,
	})
	if err == nil {
		s.proofAttached = true
	}
	return err
}

// Solve maps the source conflict limit to the concrete solver. A negative
// limit, like zero, means unlimited; other negative values are rejected.
func (s *M114PSolver) Solve(assumptions []SatLit, conflictLimit int) (SatResult, error) {
	if !s.usable() {
		return SatResult{}, ErrInvariant
	}
	if conflictLimit < -1 {
		return SatResult{}, ErrDimension
	}
	s.proof = nil
	s.rootCursor, s.chainCursor = 0, 0
	if s.forcedUnsat {
		return SatResult{Status: SatUnsatisfiable, Reason: SatFinished}, nil
	}
	if s.recordProof && s.solver.ClauseCount() > 0 {
		if err := s.attachProof(); err != nil {
			return SatResult{}, err
		}
	}
	limits := SatLimits{Stop: s.options.Stop}
	if conflictLimit > 0 {
		limits.Conflicts = uint64(conflictLimit)
	}
	result, err := s.solver.Solve(assumptions, limits)
	if err != nil {
		return result, err
	}
	if result.Status == SatUnsatisfiable && s.recordProof && len(assumptions) == 0 {
		proof, proofErr := s.buildProofSnapshot()
		if proofErr != nil {
			return result, proofErr
		}
		s.proof = proof
	}
	return result, nil
}

func (s *M114PSolver) ConflictCount() (uint64, error) {
	if !s.usable() {
		return 0, ErrInvariant
	}
	return s.solver.ConflictCount(), nil
}

func (s *M114PSolver) StatisticsText(elapsed time.Duration) (string, error) {
	if !s.usable() || elapsed < 0 {
		return "", ErrDimension
	}
	var text strings.Builder
	text.WriteString(FormatSatSolverStatistics(s.solver.Statistics(), false))
	text.WriteString("elapsed: ")
	text.WriteString(strconv.FormatFloat(elapsed.Seconds(), 'f', 6, 64))
	text.WriteString(" s\n")
	return text.String(), nil
}

func m114PCloneRoots(roots [][]SatLit) [][]SatLit {
	result := make([][]SatLit, len(roots))
	for index := range roots {
		result[index] = slices.Clone(roots[index])
	}
	return result
}

func cloneM114PProof(proof *M114PProofSnapshot) *M114PProofSnapshot {
	if proof == nil {
		return nil
	}
	clone := &M114PProofSnapshot{Roots: m114PCloneRoots(proof.Roots), FinalClause: proof.FinalClause}
	clone.Chains = make([]M114PProofChain, len(proof.Chains))
	for index, chain := range proof.Chains {
		clone.Chains[index] = M114PProofChain{
			ClauseIDs: slices.Clone(chain.ClauseIDs),
			Variables: slices.Clone(chain.Variables),
			Clause:    slices.Clone(chain.Clause),
		}
	}
	return clone
}

func m114PReferenceClause(reference SatProofDAGRef, roots [][]SatLit, nodes map[int][]SatLit) ([]SatLit, error) {
	if reference.Clause >= 0 && reference.Node == -1 && reference.Clause < len(roots) {
		return roots[reference.Clause], nil
	}
	if reference.Node >= 0 && reference.Clause == -1 {
		if clause, found := nodes[reference.Node]; found {
			return clause, nil
		}
	}
	return nil, ErrInvariant
}

func m114PResolutionPivot(first, second []SatLit) (SatVar, error) {
	found := false
	var pivot SatVar
	for _, literal := range first {
		index, present := slices.BinarySearch(second, literal.Not())
		if !present || index < 0 {
			continue
		}
		if found && pivot != literal.Var() {
			return 0, ErrInvariant
		}
		found, pivot = true, literal.Var()
	}
	if !found {
		return 0, ErrInvariant
	}
	return pivot, nil
}

func (s *M114PSolver) buildProofSnapshot() (*M114PProofSnapshot, error) {
	if !s.usable() || !s.proofAttached || s.solver.core.resolutionProof == nil {
		return nil, ErrInvariant
	}
	if final, err := s.solver.VerifyResolutionProof(); err != nil || len(final) != 0 {
		return nil, ErrInvariant
	}
	p := s.solver.core.resolutionProof
	used, err := p.dag.CollectUsedSorted([]int{p.final})
	if err != nil {
		return nil, err
	}
	// The source traversal exposes every root in insertion order. ABC's
	// interpolation clients index a separately built root-partition map with
	// that traversal index, so compacting roots to an UNSAT core would silently
	// change the meaning of the proof even though the resolution still replays.
	rootMap := make(map[int]int, len(p.roots))
	proof := &M114PProofSnapshot{Roots: m114PCloneRoots(p.roots), FinalClause: -1}
	for root := range p.roots {
		rootMap[root] = root
	}
	nodeMap := make(map[int]int, len(used))
	nodeClauses := make(map[int][]SatLit, len(used))
	for index, handle := range used {
		nodeMap[handle] = len(proof.Roots) + index
		node, nodeErr := p.dag.Node(handle)
		if nodeErr != nil || len(node.Antecedents) == 0 {
			return nil, ErrInvariant
		}
		chain := M114PProofChain{ClauseIDs: make([]int, len(node.Antecedents))}
		current, clauseErr := m114PReferenceClause(node.Antecedents[0], p.roots, nodeClauses)
		if clauseErr != nil {
			return nil, clauseErr
		}
		current = slices.Clone(current)
		for antecedentIndex, reference := range node.Antecedents {
			if reference.Clause >= 0 {
				mapped, found := rootMap[reference.Clause]
				if !found {
					return nil, ErrInvariant
				}
				chain.ClauseIDs[antecedentIndex] = mapped
			} else {
				mapped, found := nodeMap[reference.Node]
				if !found || reference.Node >= handle {
					return nil, ErrInvariant
				}
				chain.ClauseIDs[antecedentIndex] = mapped
			}
			if antecedentIndex == 0 {
				continue
			}
			next, nextErr := m114PReferenceClause(reference, p.roots, nodeClauses)
			if nextErr != nil {
				return nil, nextErr
			}
			pivot, pivotErr := m114PResolutionPivot(current, next)
			if pivotErr != nil {
				return nil, pivotErr
			}
			chain.Variables = append(chain.Variables, pivot)
			current, nextErr = resolveSatLiteralClauses(current, next)
			if nextErr != nil {
				return nil, nextErr
			}
		}
		chain.Clause = slices.Clone(current)
		nodeClauses[handle] = slices.Clone(current)
		proof.Chains = append(proof.Chains, chain)
	}
	final, found := nodeMap[p.final]
	if !found || len(proof.Chains) == 0 || len(proof.Chains[len(proof.Chains)-1].Clause) != 0 {
		return nil, ErrInvariant
	}
	proof.FinalClause = final
	return proof, nil
}

func (s *M114PSolver) ProofReady() bool {
	return s != nil && !s.closed && s.proof != nil
}

func (s *M114PSolver) ProofClauseCount() (int, error) {
	if !s.ProofReady() {
		return 0, ErrInvariant
	}
	return len(s.proof.Roots) + len(s.proof.Chains), nil
}

func (s *M114PSolver) ProofSnapshot() (*M114PProofSnapshot, error) {
	if !s.ProofReady() {
		return nil, ErrInvariant
	}
	return cloneM114PProof(s.proof), nil
}

func (s *M114PSolver) FirstRoot() ([]SatLit, bool, error) {
	if !s.ProofReady() {
		return nil, false, ErrInvariant
	}
	s.rootCursor = 0
	return s.NextRoot()
}

func (s *M114PSolver) NextRoot() ([]SatLit, bool, error) {
	if !s.ProofReady() {
		return nil, false, ErrInvariant
	}
	if s.rootCursor >= len(s.proof.Roots) {
		return nil, false, nil
	}
	root := slices.Clone(s.proof.Roots[s.rootCursor])
	s.rootCursor++
	return root, true, nil
}

func (s *M114PSolver) FirstChain() (M114PProofChain, bool, error) {
	if !s.ProofReady() {
		return M114PProofChain{}, false, ErrInvariant
	}
	s.chainCursor = 0
	return s.NextChain()
}

func (s *M114PSolver) NextChain() (M114PProofChain, bool, error) {
	if !s.ProofReady() {
		return M114PProofChain{}, false, ErrInvariant
	}
	if s.chainCursor >= len(s.proof.Chains) {
		return M114PProofChain{}, false, nil
	}
	chain := s.proof.Chains[s.chainCursor]
	s.chainCursor++
	return M114PProofChain{
		ClauseIDs: slices.Clone(chain.ClauseIDs),
		Variables: slices.Clone(chain.Variables),
		Clause:    slices.Clone(chain.Clause),
	}, true, nil
}

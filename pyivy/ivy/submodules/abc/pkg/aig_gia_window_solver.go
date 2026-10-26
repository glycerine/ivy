package abc

// SAT-window assumptions and solving from aig/gia/giaSatLut.c:Sbl_ManTestSat.

// GiaWindowSolver owns one fixed structural window. Repeated bound queries
// retain learned clauses. It must not be copied or used concurrently.
// Candidate Boolean validity and timing feedback are separate layers.
type GiaWindowSolver struct {
	baseClauses, baseLiterals  int
	solver                     *SatSolver
	outputs                    []SatVar
	roots                      []int
	nodes, inputs, first, cuts int
}

type GiaWindowSolveResult struct {
	SAT           SatResult
	SelectedCuts  []int // ascending original cut indices, including any true unused records
	SelectedNodes []int
}

func NewGiaWindowSolver(logN, nodeCount int, cuts []GiaWindowCut, ranges []GiaWindowCutRange, roots []int, options GiaWindowCNFOptions) (*GiaWindowSolver, error) {
	if nodeCount < 0 || logN < 0 || logN > 20 || nodeCount > 1<<uint(logN) {
		return nil, ErrDimension
	}
	if options.MaxVariables < 0 || options.MaxClauses < 0 || options.MaxLiterals < 0 {
		return nil, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 1 << 20
	}
	if options.MaxClauses == 0 {
		options.MaxClauses = 1 << 20
	}
	if options.MaxLiterals == 0 {
		options.MaxLiterals = 4 << 20
	}
	if len(roots) > options.MaxLiterals {
		return nil, ErrLimit
	}
	for _, root := range roots {
		if root < 0 || root >= nodeCount {
			return nil, ErrDimension
		}
	}
	cardinality, e := BuildGiaCardinalityCNF(logN, options)
	if e != nil {
		return nil, e
	}
	remaining := options
	remaining.MaxClauses -= len(cardinality.Clauses)
	for _, c := range cardinality.Clauses {
		remaining.MaxLiterals -= len(c)
	}
	// Zero denotes a default in the builders; reject exhausted aggregate budgets
	// if structural clauses are needed, then use explicit positive residual caps.
	if nodeCount > 0 && (remaining.MaxClauses == 0 || remaining.MaxLiterals == 0) {
		return nil, ErrLimit
	}
	if remaining.MaxClauses == 0 {
		remaining.MaxClauses = 1
	}
	if remaining.MaxLiterals == 0 {
		remaining.MaxLiterals = 1
	}
	structure, e := BuildGiaWindowCNF(nodeCount, cardinality.Variables, cuts, ranges, remaining)
	if e != nil {
		return nil, e
	}
	s, e := NewSatSolverWithLimit(options.MaxVariables)
	if e != nil {
		return nil, e
	}
	if e = s.SetVariableCount(cardinality.Variables); e != nil {
		return nil, e
	}
	for index, group := range [][][]SatLit{cardinality.Clauses, structure.Clauses} {
		for _, c := range group {
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			if _, e = s.AddClause(c); e != nil {
				return nil, e
			}
		}
		if index == 0 {
			if e = s.Bookmark(); e != nil {
				return nil, e
			}
			if e = s.SetVariableCount(structure.Variables); e != nil {
				return nil, e
			}
		}
	}
	baseLiterals := 0
	for _, c := range cardinality.Clauses {
		baseLiterals += len(c)
	}
	return &GiaWindowSolver{baseClauses: len(cardinality.Clauses), baseLiterals: baseLiterals, solver: s, outputs: cardinality.Outputs, roots: append([]int(nil), roots...), nodes: nodeCount, inputs: cardinality.Inputs, first: cardinality.Variables, cuts: len(cuts)}, nil
}

// Assumptions returns the native order: negative cardinality threshold,
// negative unused padded node variables, then positive roots (duplicates kept).
// A bound equal to the input capacity needs no threshold literal.
func (s *GiaWindowSolver) Assumptions(bound int) ([]SatLit, error) {
	if s == nil || s.solver == nil || bound < 0 || bound > s.inputs {
		return nil, ErrDimension
	}
	a := make([]SatLit, 0, 1+s.inputs-s.nodes+len(s.roots))
	add := func(v SatVar, neg bool) { l, _ := MakeSatLit(v, neg); a = append(a, l) }
	if bound < s.inputs {
		add(s.outputs[bound], true)
	}
	for i := s.nodes; i < s.inputs; i++ {
		add(SatVar(i), true)
	}
	for _, root := range s.roots {
		add(SatVar(root), false)
	}
	return a, nil
}

func (s *GiaWindowSolver) Solve(bound int, limits SatLimits) (GiaWindowSolveResult, error) {
	a, e := s.Assumptions(bound)
	if e != nil {
		return GiaWindowSolveResult{}, e
	}
	answer, e := s.solver.Solve(a, limits)
	if e != nil {
		return GiaWindowSolveResult{}, e
	}
	r := GiaWindowSolveResult{SAT: answer}
	if answer.Status == SatSatisfiable {
		for i := 0; i < s.nodes; i++ {
			if answer.Model[i] {
				r.SelectedNodes = append(r.SelectedNodes, i)
			}
		}
		for i := 0; i < s.cuts; i++ {
			if answer.Model[s.first+i] {
				r.SelectedCuts = append(r.SelectedCuts, i)
			}
		}
	}
	return r, nil
}

// ResetWindow replaces window clauses and learned clauses while retaining the
// compiled cardinality prefix. Invalid input/cancellation leaves the current
// window usable. Its capacity (the original power-of-two input count) is fixed.
func (s *GiaWindowSolver) ResetWindow(nodeCount int, cuts []GiaWindowCut, ranges []GiaWindowCutRange, roots []int, o GiaWindowCNFOptions) error {
	if s == nil || s.solver == nil || s.solver.bookmark == nil || nodeCount < 0 || nodeCount > s.inputs || o.MaxVariables < 0 || o.MaxClauses < 0 || o.MaxLiterals < 0 {
		return ErrDimension
	}
	if s.solver.solving {
		return ErrInvariant
	}
	if o.MaxVariables == 0 {
		o.MaxVariables = 1 << 20
	}
	if o.MaxClauses == 0 {
		o.MaxClauses = 1 << 20
	}
	if o.MaxLiterals == 0 {
		o.MaxLiterals = 4 << 20
	}
	if len(roots) > o.MaxLiterals {
		return ErrLimit
	}
	for _, root := range roots {
		if root < 0 || root >= nodeCount {
			return ErrDimension
		}
	}
	remaining := o
	remaining.MaxClauses -= s.baseClauses
	remaining.MaxLiterals -= s.baseLiterals
	if remaining.MaxClauses < 0 || remaining.MaxLiterals < 0 || nodeCount > 0 && (remaining.MaxClauses == 0 || remaining.MaxLiterals == 0) {
		return ErrLimit
	}
	if remaining.MaxClauses == 0 {
		remaining.MaxClauses = 1
	}
	if remaining.MaxLiterals == 0 {
		remaining.MaxLiterals = 1
	}
	structure, e := BuildGiaWindowCNF(nodeCount, s.first, cuts, ranges, remaining)
	if e != nil {
		return e
	}
	base := s.solver.bookmark
	candidate := base.state.cloneState()
	candidate.bookmark = base
	candidate.stats = s.solver.stats
	if e = candidate.Rollback(); e != nil {
		return e
	}
	candidate.maxVariables = o.MaxVariables
	if e = candidate.SetVariableCount(structure.Variables); e != nil {
		return e
	}
	for _, c := range structure.Clauses {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		if _, e = candidate.AddClause(c); e != nil {
			return e
		}
	}
	if o.Stop != nil && o.Stop() {
		return ErrCancelled
	}
	s.solver = &candidate
	s.nodes = nodeCount
	s.cuts = len(cuts)
	s.roots = append([]int(nil), roots...)
	return nil
}

package abc

// Graph-free CNF and literal-position transforms from sat/cnf/cnfMan.c.

import "slices"

type CNFFormula struct {
	variables int
	clauses   [][]SatLit
}

func (f *CNFFormula) VariableCount() int { return f.variables }
func (f *CNFFormula) Clauses() [][]SatLit {
	r := make([][]SatLit, len(f.clauses))
	for i, c := range f.clauses {
		r[i] = slices.Clone(c)
	}
	return r
}
func checkCNFFormula(f *CNFFormula, o CNFTransformOptions) error {
	if f == nil || f.variables < 0 || o.MaxVariables < 0 || o.MaxClauses < 0 || o.MaxLiterals < 0 {
		return ErrDimension
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
	if o.Stop != nil && o.Stop() {
		return ErrCancelled
	}
	if f.variables > o.MaxVariables || uint64(f.variables) > uint64(MaxLiteralID)+1 || len(f.clauses) > o.MaxClauses {
		return ErrLimit
	}
	count := 0
	for _, clause := range f.clauses {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		if len(clause) > o.MaxLiterals-count {
			return ErrLimit
		}
		count += len(clause)
		for _, l := range clause {
			if uint64(l.Var()) >= uint64(f.variables) {
				return ErrInvalidLiteral
			}
		}
	}
	return nil
}

// NewCNFFormula copies a checked formula, including empty clauses and a zero-
// variable empty formula. No AIG metadata or custom solver interface is needed.
func NewCNFFormula(variables int, clauses [][]SatLit, o CNFTransformOptions) (*CNFFormula, error) {
	f := &CNFFormula{variables: variables, clauses: clauses}
	if e := checkCNFFormula(f, o); e != nil {
		return nil, e
	}
	r := &CNFFormula{variables: variables}
	for _, c := range clauses {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		r.clauses = append(r.clauses, slices.Clone(c))
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return r, nil
}

// CollectFlipPositions returns all matching flattened literal positions in
// order. A variable absent from the formula yields an empty list.
func (f *CNFFormula) CollectFlipPositions(variable SatVar, o CNFTransformOptions) ([]int, error) {
	if e := checkCNFFormula(f, o); e != nil {
		return nil, e
	}
	if uint64(variable) > uint64(MaxLiteralID) {
		return nil, ErrInvalidLiteral
	}
	var positions []int
	index := 0
	for _, clause := range f.clauses {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		for _, l := range clause {
			if l.Var() == variable {
				positions = append(positions, index)
			}
			index++
		}
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return positions, nil
}

// LiftAndFlipPositions flips then shifts only the selected literal occurrences.
// Positions refer to the flattened source clause stream. Duplicates apply in
// order, repeatedly shifting the same occurrence. Other literals are unchanged.
// Unlike native nVars, the owned result's span expands as needed. MaxLiterals
// bounds both source literal count and number of requested operations.
func (f *CNFFormula) LiftAndFlipPositions(base SatVar, positions []int, o CNFTransformOptions) (*CNFFormula, error) {
	if e := checkCNFFormula(f, o); e != nil {
		return nil, e
	}
	maxLiterals, maxVars := o.MaxLiterals, o.MaxVariables
	if maxLiterals == 0 {
		maxLiterals = 4 << 20
	}
	if maxVars == 0 {
		maxVars = 1 << 20
	}
	if len(positions) > maxLiterals {
		return nil, ErrLimit
	}
	count := 0
	for _, c := range f.clauses {
		count += len(c)
	}
	for _, p := range positions {
		if p < 0 || p >= count {
			return nil, ErrDimension
		}
	}
	flat := make([]SatLit, 0, count)
	for _, c := range f.clauses {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		flat = append(flat, c...)
	}
	variables := f.variables
	for _, p := range positions {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		v := uint64(flat[p].Var()) + uint64(base)
		if v > uint64(MaxLiteralID) || v >= uint64(maxVars) {
			return nil, ErrLimit
		}
		flat[p] = SatLit(v)<<1 | (flat[p]^1)&1
		variables = max(variables, int(v)+1)
	}
	r := &CNFFormula{variables: variables}
	offset := 0
	for _, c := range f.clauses {
		r.clauses = append(r.clauses, flat[offset:offset+len(c):offset+len(c)])
		offset += len(c)
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return r, nil
}
func (f *CNFFormula) LoadInto(s *SatSolver, base SatVar) (bool, error) {
	if f == nil {
		return false, ErrDimension
	}
	if f.variables > 0 {
		return (&AigCNF{variables: f.variables, clauses: f.clauses}).LoadInto(s, base)
	}
	if s == nil {
		return false, ErrDimension
	}
	if uint64(base) > uint64(s.maxVariables) || uint64(base) > uint64(MaxLiteralID)+1 {
		return false, ErrLimit
	}
	if e := s.SetVariableCount(int(base)); e != nil {
		return false, e
	}
	for _, c := range f.clauses {
		ok, e := s.AddClause(c)
		if e != nil || !ok {
			return ok, e
		}
	}
	return true, nil
}

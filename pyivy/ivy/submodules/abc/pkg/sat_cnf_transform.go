package abc

// Owned CNF operations from sat/cnf/cnfMan.c.

import "slices"

type CNFTransformOptions struct {
	MaxVariables, MaxClauses, MaxLiterals int // zero: 1M, 1M, 4M
	Stop                                  func() bool
}

// Clone owns every clause and mapping slice. A nil receiver yields nil.
func (c *AigMappedCNF) Clone() *AigMappedCNF {
	if c == nil {
		return nil
	}
	r := &AigMappedCNF{variables: c.variables, objectVariables: slices.Clone(c.objectVariables), inputs: slices.Clone(c.inputs), outputs: slices.Clone(c.outputs)}
	for _, clause := range c.clauses {
		r.clauses = append(r.clauses, slices.Clone(clause))
	}
	return r
}
func checkCNFTransform(c *AigMappedCNF, o CNFTransformOptions) error {
	if c == nil || c.variables < 1 || o.MaxVariables < 0 || o.MaxClauses < 0 || o.MaxLiterals < 0 {
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
	if c.variables > o.MaxVariables || len(c.clauses) > o.MaxClauses {
		return ErrLimit
	}
	count := 0
	for _, clause := range c.clauses {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		if len(clause) > o.MaxLiterals-count {
			return ErrLimit
		}
		count += len(clause)
		for _, l := range clause {
			if uint64(l.Var()) >= uint64(c.variables) {
				return ErrInvalidLiteral
			}
		}
	}
	for _, v := range c.objectVariables {
		if v < -1 || v >= c.variables {
			return ErrInvalidLiteral
		}
	}
	for _, ls := range [][]SatLit{c.inputs, c.outputs} {
		for _, l := range ls {
			if uint64(l.Var()) >= uint64(c.variables) {
				return ErrInvalidLiteral
			}
		}
	}
	return nil
}

// Shifted returns an independently owned relation with every literal and
// assigned object variable shifted by base. Unlike native Cnf_DataLift, the
// returned variable count includes the offset, making it directly solver-ready.
func (c *AigMappedCNF) Shifted(base SatVar, o CNFTransformOptions) (*AigMappedCNF, error) {
	if e := checkCNFTransform(c, o); e != nil {
		return nil, e
	}
	limit := o.MaxVariables
	if limit == 0 {
		limit = 1 << 20
	}
	end := uint64(base) + uint64(c.variables)
	if end > uint64(limit) || end > uint64(MaxLiteralID)+1 {
		return nil, ErrLimit
	}
	r := c.Clone()
	r.variables = int(end)
	offset := SatLit(base) << 1
	for _, clause := range r.clauses {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		for i := range clause {
			clause[i] += offset
		}
	}
	for i, v := range r.objectVariables {
		if v >= 0 {
			r.objectVariables[i] = v + int(base)
		}
	}
	for _, ls := range [][]SatLit{r.inputs, r.outputs} {
		for i := range ls {
			ls[i] += offset
		}
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return r, nil
}

// TransformPolarity changes each assigned variable by its AIG zero-input phase,
// optionally including COs, matching Cnf_DataTranformPolarity. Returned input/
// output literals also track that phase so their original signal meanings are
// preserved. Source clauses, graph and object-variable numbers are unchanged.
func (c *AigMappedCNF) TransformPolarity(g *AigManager, includeOutputs bool, o CNFTransformOptions) (*AigMappedCNF, error) {
	if e := checkCNFTransform(c, o); e != nil {
		return nil, e
	}
	if g == nil || len(g.objects) != len(c.objectVariables) {
		return nil, ErrDimension
	}
	if e := g.Check(); e != nil {
		return nil, e
	}
	phases := make([]bool, c.variables)
	for id, obj := range g.objects {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		if obj.Type == AigNone || obj.Type == AigCO && !includeOutputs {
			continue
		}
		if v := c.objectVariables[id]; v >= 0 {
			phases[v] = obj.Phase
		}
	}
	r := c.Clone()
	for _, clause := range r.clauses {
		if o.Stop != nil && o.Stop() {
			return nil, ErrCancelled
		}
		for i, l := range clause {
			if phases[l.Var()] {
				clause[i] = l.Not()
			}
		}
	}
	for _, ls := range [][]SatLit{r.inputs, r.outputs} {
		for i, l := range ls {
			if phases[l.Var()] {
				ls[i] = l.Not()
			}
		}
	}
	if o.Stop != nil && o.Stop() {
		return nil, ErrCancelled
	}
	return r, nil
}

// LoadInto adds this relation to an incremental solver at base. Assumptions
// from the stored mappings must receive the same offset. A false result means
// an added clause proves inconsistency. Solver mutation follows AddClause.
func (c *AigMappedCNF) LoadInto(s *SatSolver, base SatVar) (bool, error) {
	if c == nil {
		return false, ErrDimension
	}
	return (&AigCNF{variables: c.variables, clauses: c.clauses}).LoadInto(s, base)
}

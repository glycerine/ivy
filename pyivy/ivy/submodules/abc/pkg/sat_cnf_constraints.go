package abc

// Output and XOR clauses from sat/cnf/cnfMan.c.

import "slices"

// OutputORClause requires assigned CO variables and returns one clause over all
// COs, including latch inputs. Signal phases and base offsets are preserved.
// With no COs the empty clause is returned; Go reports inconsistency instead
// of triggering the native solver assertion on an empty clause.
func (c *AigMappedCNF) OutputORClause(g *AigManager, base SatVar) ([]SatLit, error) {
	if c == nil || g == nil || len(c.objectVariables) != len(g.objects) || len(c.outputs) != len(g.cos) {
		return nil, ErrDimension
	}
	if e := g.Check(); e != nil {
		return nil, e
	}
	if uint64(base)+uint64(c.variables) > uint64(MaxLiteralID)+1 {
		return nil, ErrLimit
	}
	result := make([]SatLit, len(g.cos))
	for i, id := range g.cos {
		v := c.objectVariables[id]
		if v < 0 || v >= c.variables || uint64(c.outputs[i].Var()) != uint64(v) {
			return nil, ErrDimension
		}
		result[i] = c.outputs[i] + (SatLit(base) << 1)
	}
	return result, nil
}
func (c *AigMappedCNF) OutputANDClauses(g *AigManager, base SatVar) ([][]SatLit, error) {
	outputs, e := c.OutputORClause(g, base)
	if e != nil {
		return nil, e
	}
	clauses := make([][]SatLit, len(outputs))
	for i, l := range outputs {
		clauses[i] = []SatLit{l}
	}
	return clauses, nil
}
func cnfConstraintRoom(s *SatSolver, variables int, base SatVar) error {
	if s == nil || variables < 1 {
		return ErrDimension
	}
	end := uint64(base) + uint64(variables)
	if end > uint64(MaxLiteralID)+1 || end > uint64(s.maxVariables) {
		return ErrLimit
	}
	return s.SetVariableCount(int(end))
}
func (c *AigMappedCNF) AddOutputOR(s *SatSolver, g *AigManager, base SatVar) (bool, error) {
	clause, e := c.OutputORClause(g, base)
	if e != nil {
		return false, e
	}
	if e = cnfConstraintRoom(s, c.variables, base); e != nil {
		return false, e
	}
	return s.AddClause(clause)
}

// AddOutputORSolver2 is the proof-capable counterpart of AddOutputOR. The
// clause and variable growth commit atomically and retain the caller's ID.
func (c *AigMappedCNF) AddOutputORSolver2(s *SatSolver2, g *AigManager, base SatVar, id int) (bool, error) {
	clause, err := c.OutputORClause(g, base)
	if err != nil {
		return false, err
	}
	if s == nil || id == 0 || len(clause) == 0 {
		return false, ErrDimension
	}
	if !s.usable() || s.core.proof != nil || s.core.resolutionProof != nil {
		return false, ErrInvariant
	}
	end := uint64(base) + uint64(c.variables)
	if end > uint64(MaxLiteralID)+1 || end > uint64(s.core.maxVariables) {
		return false, ErrLimit
	}
	trial := *s
	core := s.core.cloneState()
	core.bookmark = s.core.bookmark
	trial.core = &core
	trial.clauses = cloneSatSolver2Clauses(s.clauses)
	trial.variablePart = slices.Clone(s.variablePart)
	if err = trial.SetVariableCount(int(end)); err != nil {
		return false, err
	}
	if _, err = trial.AddClause(clause, id); err != nil {
		if err == ErrInvariant {
			return false, nil
		}
		return false, err
	}
	*s = trial
	return true, nil
}
func (c *AigMappedCNF) AddOutputAND(s *SatSolver, g *AigManager, base SatVar) (bool, error) {
	clauses, e := c.OutputANDClauses(g, base)
	if e != nil {
		return false, e
	}
	if e = cnfConstraintRoom(s, c.variables, base); e != nil {
		return false, e
	}
	for _, clause := range clauses {
		ok, e := s.AddClause(clause)
		if e != nil || !ok {
			return ok, e
		}
	}
	return true, nil
}

// CNFXORClauses encodes result = a XOR b in native clause order. Variable IDs
// must be positive; aliases are allowed. Each returned clause owns its storage.
func CNFXORClauses(a, b, result SatVar) ([][]SatLit, error) {
	for _, v := range []SatVar{a, b, result} {
		if v == 0 {
			return nil, ErrDimension
		}
		if uint64(v) > uint64(MaxLiteralID) {
			return nil, ErrInvalidLiteral
		}
	}
	x, y, z := SatLit(a)<<1, SatLit(b)<<1, SatLit(result)<<1
	return [][]SatLit{{x.Not(), y.Not(), z.Not()}, {x.Not(), y, z}, {x, y.Not(), z}, {x, y, z.Not()}}, nil
}
func AddCNFXOR(s *SatSolver, a, b, result SatVar) (bool, error) {
	clauses, e := CNFXORClauses(a, b, result)
	if e != nil {
		return false, e
	}
	if e = cnfConstraintRoom(s, int(max(a, b, result))+1, 0); e != nil {
		return false, e
	}
	for _, clause := range clauses {
		ok, e := s.AddClause(clause)
		if e != nil || !ok {
			return ok, e
		}
	}
	return true, nil
}

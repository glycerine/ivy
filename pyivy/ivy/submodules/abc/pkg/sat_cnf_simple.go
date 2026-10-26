package abc

// Reference Tseitin CNF for the AIG port. This is the independent simple
// encoder called for by the port plan, not yet ABC's optimized DAR-cut encoder
// in src/sat/cnf/cnfCore.c and cnfWrite.c.

import (
	"math"
	"slices"
)

const InvalidSatLit = SatLit(math.MaxUint32)

// AigCNF represents a full Boolean relation: no outputs are asserted by the
// encoder. The constant-one variable is explicitly constrained. Buffers and
// COs alias their drivers; every live AND and CI has its own SAT variable.
// A sequential AIG is encoded as one combinational transition relation.
type AigCNF struct {
	variables       int
	clauses         [][]SatLit
	objectLiterals  []SatLit
	inputs, outputs []SatLit
}

func (c *AigCNF) VariableCount() int { return c.variables }
func (c *AigCNF) Clauses() [][]SatLit {
	result := make([][]SatLit, len(c.clauses))
	for i, lits := range c.clauses {
		result[i] = slices.Clone(lits)
	}
	return result
}
func (c *AigCNF) InputLiterals() []SatLit  { return slices.Clone(c.inputs) }
func (c *AigCNF) OutputLiterals() []SatLit { return slices.Clone(c.outputs) }
func (c *AigCNF) ObjectLiteral(id AigID) (SatLit, error) {
	if uint64(id) >= uint64(len(c.objectLiterals)) || c.objectLiterals[id] == InvalidSatLit {
		return InvalidSatLit, ErrInvalidLiteral
	}
	return c.objectLiterals[id], nil
}

// EncodeAigCNF preserves both positive and complemented graph signals; its
// result does not depend on a SAT solver or any optimization/cut data.
func EncodeAigCNF(g *AigManager) (*AigCNF, error) {
	if g == nil {
		return nil, ErrDimension
	}
	if err := g.Check(); err != nil {
		return nil, err
	}
	c := &AigCNF{variables: 1, objectLiterals: make([]SatLit, len(g.objects))}
	for i := range c.objectLiterals {
		c.objectLiterals[i] = InvalidSatLit
	}
	c.objectLiterals[0] = 0
	c.clauses = append(c.clauses, []SatLit{0})
	for _, id := range g.cis {
		l := SatLit(c.variables) << 1
		c.variables++
		c.objectLiterals[id] = l
		c.inputs = append(c.inputs, l)
	}
	convert := func(l AigLit) SatLit { return c.objectLiterals[l.ID()] ^ SatLit(l&1) }
	for i, obj := range g.objects {
		switch obj.Type {
		case AigBuffer, AigCO:
			c.objectLiterals[i] = convert(obj.Fanin0)
		case AigAND:
			a, b := convert(obj.Fanin0), convert(obj.Fanin1)
			z := SatLit(c.variables) << 1
			c.variables++
			c.objectLiterals[i] = z
			c.clauses = append(c.clauses, []SatLit{z.Not(), a}, []SatLit{z.Not(), b}, []SatLit{z, a.Not(), b.Not()})
		}
	}
	for _, id := range g.cos {
		c.outputs = append(c.outputs, c.objectLiterals[id])
	}
	return c, nil
}

// LoadInto adds this relation to an incremental solver with a variable offset.
// Literal mappings must be shifted by 2*base when preparing assumptions. Using
// different bases creates independent copies, useful for timeframe unrolling.
func (c *AigCNF) LoadInto(s *SatSolver, base SatVar) (bool, error) {
	return c.loadIntoWithStop(s, base, nil)
}
func (c *AigCNF) loadIntoWithStop(s *SatSolver, base SatVar, stop func() bool) (bool, error) {
	if c == nil || s == nil || c.variables < 1 {
		return false, ErrDimension
	}
	if stop != nil && stop() {
		return false, ErrCancelled
	}
	end := uint64(base) + uint64(c.variables)
	if end > uint64(MaxLiteralID)+1 || end > uint64(s.maxVariables) {
		return false, ErrLimit
	}
	if err := s.SetVariableCount(int(end)); err != nil {
		return false, err
	}
	offset := SatLit(base) << 1
	for _, clause := range c.clauses {
		if stop != nil && stop() {
			return false, ErrCancelled
		}
		mapped := make([]SatLit, len(clause))
		for i, l := range clause {
			mapped[i] = l + offset
		}
		ok, err := s.AddClause(mapped)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

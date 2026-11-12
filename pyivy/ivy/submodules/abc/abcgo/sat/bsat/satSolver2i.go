package bsat

import "github.com/glycerine/abcgo/aig/aig"

/**CFile****************************************************************

  FileName    [satSolver2i.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver.]

  Synopsis    [Records the trace of SAT solving in the CNF form.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 2, 2013.]

  Revision    [$Id: satSolver2i.c,v 1.4 2013/09/02 00:00:00 casem Exp $]

***********************************************************************/

// The graph representation uses the shared AIG package. Clause interpolation
// follows the original on-line OR/AND resolution algorithm and literal handles.
type Int2_Man_t = Int2_Man_t_
type Int2_Man_t_ struct {
	pSat     *sat_solver2
	vGloVars []int32
	vVar2Glo map[int32]int32
	pAig     *aig.Graph
	roots    map[*clause]int32
}

func Int2_ManStart(s *sat_solver2, globals []int32, n int32) *Int2_Man_t {
	p := &Int2_Man_t{pSat: s, vGloVars: append([]int32(nil), globals[:n]...), vVar2Glo: map[int32]int32{}, pAig: aig.NewGraph(), roots: map[*clause]int32{}}
	p.pAig.Name = "interpolant"
	for i, v := range p.vGloVars {
		if v < 0 || v >= s.size {
			panic("interpolation variable outside solver")
		}
		if _, exists := p.vVar2Glo[v]; exists {
			panic("duplicate interpolation variable")
		}
		p.vVar2Glo[v] = int32(i)
		p.pAig.NewInput()
	}
	return p
}
func Int2_ManStop(p *Int2_Man_t) {
	if p != nil {
		*p = Int2_Man_t{}
	}
}
func Int2_ManReadInterpolant(s *sat_solver2) *aig.Graph {
	if s.hProofLast == -1 {
		return nil
	}
	p, ok := s.pInt2.(*Int2_Man_t)
	if !ok || p.pAig == nil {
		return nil
	}
	result := p.pAig
	p.pAig = nil
	result.AddOutput(aig.Edge(s.hProofLast))
	s.hProofLast = -1
	return result
}
func Int2_ManChainStart(p *Int2_Man_t, c *clause) int32 { return p.chainStart(c) }
func (p *Int2_Man_t_) chainStart(c *clause) int32 {
	if c.lrn != 0 {
		return p.pSat.claProofs.ptr[clause_id(c)]
	}
	if c.partA == 0 {
		return int32(p.pAig.Const1())
	}
	if cached, ok := p.roots[c]; ok {
		return cached
	}
	value := p.pAig.Const0()
	for _, l := range c.lits[:c.size] {
		if position, shared := p.vVar2Glo[lit_var(l)]; shared {
			variable := aig.Aig_IthVar(p.pAig, position)
			if lit_sign(l) != 0 {
				variable = variable.Not()
			}
			value = p.pAig.OR(value, variable)
		}
	}
	p.roots[c] = int32(value)
	return int32(value)
}
func Int2_ManChainResolve(p *Int2_Man_t, c *clause, current, varA int32) int32 {
	return p.chainResolve(c, current, varA)
}
func (p *Int2_Man_t_) chainResolve(c *clause, current, varA int32) int32 {
	other := aig.Edge(p.chainStart(c))
	value := aig.Edge(current)
	if varA != 0 {
		return int32(p.pAig.OR(value, other))
	}
	return int32(p.pAig.AND(value, other))
}

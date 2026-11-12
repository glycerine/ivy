package bsat2

/************************************************************************************[SimpSolver.h]
Copyright (c) 2006,      Niklas Een, Niklas Sorensson
Copyright (c) 2007-2010, Niklas Sorensson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

type SimpSolver struct {
	*Solver
	Grow, ClauseLimit, SubsumptionLimit               int
	SimplificationGarbageFraction                     float64
	UseAsymmetric, UseRedundancyCheck, UseElimination bool
	Merges, AsymmetricLiterals, EliminatedVariables   uint64
	simplifying                                       bool
	eliminated, frozen, touched                       []bool
	occurrences                                       *OccLists[CRef]
	counts                                            []int
	eliminationHeap                                   *Heap
	queue                                             []CRef
	backwardAssignments                               int
	dummy                                             CRef
	eliminationClauses                                [][]Lit
}

func NewSimpSolver() *SimpSolver {
	p := &SimpSolver{Solver: NewSolver(), ClauseLimit: 20, SubsumptionLimit: 1000, SimplificationGarbageFraction: 0.5, UseElimination: true, simplifying: true}
	p.ca.ExtraClauseField = true
	p.dummy = p.ca.Alloc([]Lit{LitUndef}, false)
	p.removeSatisfiedOriginal = false
	p.occurrences = NewOccLists(p.clauseDeleted)
	p.eliminationHeap = NewHeap(p.eliminationLess)
	p.relocateExtra = p.relocateSimplification
	return p
}
func (p *SimpSolver) clauseDeleted(h CRef) bool { return p.ca.At(h).Mark == 1 }
func (p *SimpSolver) eliminationLess(a, b int32) bool {
	return int64(p.counts[2*a])*int64(p.counts[2*a+1]) < int64(p.counts[2*b])*int64(p.counts[2*b+1])
}
func (p *SimpSolver) NewVar(negative, decision bool) Var {
	v := p.Solver.NewVar(negative, decision)
	p.frozen = append(p.frozen, false)
	p.eliminated = append(p.eliminated, false)
	if p.simplifying {
		p.counts = append(p.counts, 0, 0)
		p.touched = append(p.touched, false)
		p.occurrences.Init(int(v))
		p.eliminationHeap.Insert(int32(v))
	}
	return v
}
func (p *SimpSolver) NewVariable() Var        { return p.NewVar(true, true) }
func (p *SimpSolver) IsEliminated(v Var) bool { return p.eliminated[v] }
func (p *SimpSolver) SetFrozen(v Var, frozen bool) {
	p.frozen[v] = frozen
	if p.simplifying && !frozen {
		p.updateElimination(v)
	}
}
func (p *SimpSolver) updateElimination(v Var) {
	if p.eliminationHeap.Contains(int32(v)) || !p.frozen[v] && !p.eliminated[v] && p.assigns[v] == Undef {
		p.eliminationHeap.Update(int32(v))
	}
}

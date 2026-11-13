package glucose

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

// SimpSolver adds MiniSat/Glucose's bounded variable elimination, backward
// subsumption, self-subsuming resolution and optional asymmetric branching.
type SimpSolver struct {
	*Solver
	Grow, ClauseLimit, SubsumptionLimit                  int
	UseAsymm, UseRCheck, UseElim                         bool
	Parsing                                              bool
	Merges, AsymmLits, EliminatedVars, EliminatedClauses int
	frozen, eliminated, touched                          []bool
	occurs                                               [][]*Clause
	literalOccurrences                                   []int
	elimHeap                                             *Heap
	subsumptionQueue                                     []*Clause
	backwardAssignments                                  int
	useSimplification                                    bool
	elimination                                          []eliminationRecord
}
type eliminationRecord struct {
	variable       Var
	clauses        [][]Lit
	defaultLit     Lit
	substitute     Lit
	isSubstitution bool
}

func NewSimpSolver() *SimpSolver {
	s := &SimpSolver{Solver: NewSolver(), ClauseLimit: 20, SubsumptionLimit: 1000, UseElim: true, useSimplification: true}
	s.initializeSimplification()
	return s
}
func (s *SimpSolver) initializeSimplification() {
	s.allocator.ExtraClauseField = true
	s.removeSatisfied = false
	s.elimHeap = NewHeap(func(a, b int) bool {
		pa, na := uint64(s.literalOccurrences[2*a]), uint64(s.literalOccurrences[2*a+1])
		pb, nb := uint64(s.literalOccurrences[2*b]), uint64(s.literalOccurrences[2*b+1])
		return pa*na < pb*nb
	})
}
func (s *SimpSolver) NewVar(sign, decision bool) Var {
	v := s.Solver.NewVar(sign, decision)
	s.frozen = append(s.frozen, false)
	s.eliminated = append(s.eliminated, false)
	if s.useSimplification {
		s.literalOccurrences = append(s.literalOccurrences, 0, 0)
		s.occurs = append(s.occurs, nil)
		s.touched = append(s.touched, false)
		s.elimHeap.Insert(int(v))
	}
	return v
}
func (s *SimpSolver) AddVar(v Var) {
	for s.NVars() <= int(v) {
		s.NewVar(true, true)
	}
}
func (s *SimpSolver) SetFrozen(v Var, b bool) {
	s.frozen[v] = b
	if s.useSimplification {
		s.updateElimHeap(v)
	}
}
func (s *SimpSolver) IsEliminated(v Var) bool { return s.eliminated[v] }
func (s *SimpSolver) updateElimHeap(v Var) {
	if s.elimHeap.InHeap(int(v)) || !s.frozen[v] && !s.eliminated[v] && s.ValueVar(v).IsUndef() {
		s.elimHeap.Update(int(v))
	}
}
func (s *SimpSolver) AddClause(ps []Lit) bool {
	for _, p := range ps {
		if p < 0 || int(p.Var()) >= s.NVars() {
			panic("clause variable out of range")
		}
		if s.eliminated[p.Var()] {
			panic("adding clause containing eliminated variable")
		}
	}
	if s.UseRCheck && s.implied(ps) {
		return true
	}
	before := len(s.clauses)
	if !s.Solver.AddClause(ps) {
		return false
	}
	if !s.Parsing {
		s.proofClause(ps, false)
	}
	if s.useSimplification && len(s.clauses) == before+1 {
		c := s.clauses[len(s.clauses)-1]
		s.subsumptionQueue = append(s.subsumptionQueue, c)
		for _, p := range c.Lits {
			s.occurs[p.Var()] = append(s.occurs[p.Var()], c)
			s.literalOccurrences[p]++
			s.touched[p.Var()] = true
			if s.elimHeap.InHeap(int(p.Var())) {
				s.elimHeap.Increase(int(p.Var()))
			}
		}
	}
	return true
}
func (s *SimpSolver) AddEmptyClause() bool { return s.AddClause(nil) }
func (s *SimpSolver) removeClause(c *Clause) {
	if c.freed {
		return
	}
	if s.useSimplification {
		for _, p := range c.Lits {
			s.literalOccurrences[p]--
			s.updateElimHeap(p.Var())
		}
	}
	s.Solver.removeClause(c)
}
func (s *SimpSolver) cleanOccurrences(v Var) []*Clause {
	cs := s.occurs[v]
	dst := cs[:0]
	for _, c := range cs {
		if !c.freed {
			dst = append(dst, c)
		}
	}
	clear(cs[len(dst):])
	s.occurs[v] = dst
	return dst
}
func (s *SimpSolver) strengthenClause(c *Clause, l Lit) bool {
	s.subsumptionQueue = append(s.subsumptionQueue, c)
	next := make([]Lit, 0, len(c.Lits)-1)
	for _, p := range c.Lits {
		if p != l {
			next = append(next, p)
		}
	}
	s.proofClause(next, false)
	if len(c.Lits) == 2 {
		s.removeClause(c)
		c.Strengthen(l)
	} else {
		s.proofClause(c.Lits, true)
		s.detachClause(c)
		c.Strengthen(l)
		s.attachClause(c)
		cs := s.occurs[l.Var()]
		for i, r := range cs {
			if r == c {
				copy(cs[i:], cs[i+1:])
				cs[len(cs)-1] = nil
				s.occurs[l.Var()] = cs[:len(cs)-1]
				break
			}
		}
		s.literalOccurrences[l]--
		s.updateElimHeap(l.Var())
	}
	if len(c.Lits) == 1 {
		return s.enqueue(c.Lits[0], nil) && s.propagate() == nil
	}
	return true
}
func (s *SimpSolver) Reset() {
	s.Solver.Reset()
	s.Grow = 0
	s.AsymmLits = 0
	s.EliminatedVars = 0
	s.EliminatedClauses = 0
	s.Merges = 0
	s.backwardAssignments = 0
	s.elimination = nil
	s.frozen = nil
	s.eliminated = nil
	s.touched = nil
	s.occurs = nil
	s.literalOccurrences = nil
	s.subsumptionQueue = nil
	s.useSimplification = true
	s.initializeSimplification()
}
func (s *SimpSolver) GarbageCollect() { s.cleanUpClauses(); s.Solver.GarbageCollect() }
func (s *SimpSolver) Solve(assumptions []Lit) bool {
	s.BudgetOff()
	return s.SolveLimited(assumptions, true, false) == LTrue
}

// SolveLimited optionally runs preprocessing and can permanently turn it off.
func (s *SimpSolver) SolveLimited(assumptions []Lit, options ...bool) LBool {
	for _, p := range assumptions {
		if p < 0 || int(p.Var()) >= s.NVars() {
			panic("assumption variable out of range")
		}
		if s.eliminated[p.Var()] {
			panic("assumption variable already eliminated")
		}
	}
	doSimp, turnOff := true, false
	if len(options) > 0 {
		doSimp = options[0]
	}
	if len(options) > 1 {
		turnOff = options[1]
	}
	doSimp = doSimp && s.useSimplification
	var extraFrozen []Var
	if doSimp {
		for _, p := range assumptions {
			v := p.Var()
			if !s.frozen[v] {
				s.SetFrozen(v, true)
				extraFrozen = append(extraFrozen, v)
			}
		}
	}
	result := LFalse
	if !doSimp || s.Eliminate(turnOff) {
		result = s.Solver.SolveLimited(assumptions)
		if result == LTrue {
			s.extendModel()
		}
	}
	for _, v := range extraFrozen {
		s.SetFrozen(v, false)
	}
	return result
}

package glucose

/***********************************************************************************[SimpSolver.cc]
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

func (s *SimpSolver) merge(a, b *Clause, v Var) ([]Lit, bool) {
	s.Merges++
	if len(a.Lits) < len(b.Lits) {
		a, b = b, a
	}
	out := make([]Lit, 0, len(a.Lits)+len(b.Lits)-2)
	for _, p := range b.Lits {
		if p.Var() == v {
			continue
		}
		found := false
		for _, q := range a.Lits {
			if q.Var() == p.Var() {
				if q == p.Not() {
					return nil, false
				}
				found = true
				break
			}
		}
		if !found {
			out = append(out, p)
		}
	}
	for _, p := range a.Lits {
		if p.Var() != v {
			out = append(out, p)
		}
	}
	return out, true
}
func (s *SimpSolver) gatherTouchedClauses() {
	queued := make(map[*Clause]bool, len(s.subsumptionQueue))
	for _, c := range s.subsumptionQueue {
		queued[c] = true
	}
	for v, touched := range s.touched {
		if !touched {
			continue
		}
		for _, c := range s.cleanOccurrences(Var(v)) {
			if !queued[c] {
				s.subsumptionQueue = append(s.subsumptionQueue, c)
				queued[c] = true
			}
		}
		s.touched[v] = false
	}
}
func (s *SimpSolver) implied(c []Lit) bool {
	s.newDecisionLevel()
	for _, p := range c {
		v := s.ValueLit(p)
		if v == LTrue {
			s.CancelUntil(0)
			return false
		}
		if v != LFalse {
			s.uncheckedEnqueue(p.Not(), nil)
		}
	}
	result := s.propagate() != nil
	s.CancelUntil(0)
	return result
}
func (s *SimpSolver) backwardSubsumptionCheck() bool {
	for len(s.subsumptionQueue) > 0 || s.backwardAssignments < len(s.trail) {
		if s.interrupted.Load() {
			s.subsumptionQueue = nil
			s.backwardAssignments = len(s.trail)
			break
		}
		if len(s.subsumptionQueue) == 0 {
			p := s.trail[s.backwardAssignments]
			s.backwardAssignments++
			s.subsumptionQueue = append(s.subsumptionQueue, NewClause([]Lit{p}, false, true))
		}
		c := s.subsumptionQueue[0]
		s.subsumptionQueue[0] = nil
		s.subsumptionQueue = s.subsumptionQueue[1:]
		if c.freed {
			continue
		}
		best := c.Lits[0].Var()
		for _, p := range c.Lits[1:] {
			if len(s.occurs[p.Var()]) < len(s.occurs[best]) {
				best = p.Var()
			}
		}
		// Strengthening can remove a candidate from the very occurrence list being
		// scanned; a snapshot ensures every original candidate is visited once.
		candidates := append([]*Clause(nil), s.cleanOccurrences(best)...)
		for _, other := range candidates {
			if c.freed {
				break
			}
			if other.freed || other == c || s.SubsumptionLimit != -1 && len(other.Lits) >= s.SubsumptionLimit {
				continue
			}
			l := c.Subsumes(other)
			if l == LitUndef {
				s.removeClause(other)
			} else if l != LitError {
				if !s.strengthenClause(other, l.Not()) {
					return false
				}
			}
		}
	}
	return true
}
func (s *SimpSolver) asymm(v Var, c *Clause) bool {
	if c.freed || s.satisfied(c) {
		return true
	}
	s.newDecisionLevel()
	l := LitUndef
	for _, p := range c.Lits {
		if p.Var() != v && s.ValueLit(p) != LFalse {
			s.uncheckedEnqueue(p.Not(), nil)
		} else {
			l = p
		}
	}
	conflict := s.propagate() != nil
	s.CancelUntil(0)
	if conflict {
		s.AsymmLits++
		return s.strengthenClause(c, l)
	}
	return true
}
func (s *SimpSolver) asymmVar(v Var) bool {
	cs := append([]*Clause(nil), s.cleanOccurrences(v)...)
	if !s.ValueVar(v).IsUndef() || len(cs) == 0 {
		return true
	}
	for _, c := range cs {
		if !s.asymm(v, c) {
			return false
		}
	}
	return s.backwardSubsumptionCheck()
}
func (s *SimpSolver) eliminateVar(v Var) bool {
	cs := append([]*Clause(nil), s.cleanOccurrences(v)...)
	var pos, neg []*Clause
	for _, c := range cs {
		positive := false
		for _, p := range c.Lits {
			if p == MkLit(v, false) {
				positive = true
				break
			}
		}
		if positive {
			pos = append(pos, c)
		} else {
			neg = append(neg, c)
		}
	}
	count := 0
	for _, a := range pos {
		for _, b := range neg {
			r, ok := s.merge(a, b, v)
			if ok {
				count++
				if count > len(cs)+s.Grow || s.ClauseLimit != -1 && len(r) > s.ClauseLimit {
					return true
				}
			}
		}
	}
	s.eliminated[v] = true
	s.SetDecisionVar(v, false)
	s.EliminatedVars++
	record := eliminationRecord{variable: v, defaultLit: MkLit(v, true)}
	store := pos
	if len(pos) > len(neg) {
		store = neg
		record.defaultLit = MkLit(v, false)
	}
	for _, c := range store {
		row := append([]Lit(nil), c.Lits...)
		for i, p := range row {
			if p.Var() == v {
				row[0], row[i] = row[i], row[0]
				break
			}
		}
		record.clauses = append(record.clauses, row)
	}
	s.EliminatedClauses += len(store)
	s.elimination = append(s.elimination, record)
	for _, a := range pos {
		for _, b := range neg {
			if r, ok := s.merge(a, b, v); ok && !s.addDerivedClause(r) {
				return false
			}
		}
	}
	for _, c := range cs {
		s.removeClause(c)
	}
	s.occurs[v] = nil
	return s.backwardSubsumptionCheck()
}
func (s *SimpSolver) Substitute(v Var, x Lit) bool {
	if !s.ok {
		return false
	}
	if !s.useSimplification || v < 0 || int(v) >= s.NVars() || x < 0 || int(x.Var()) >= s.NVars() || s.eliminated[x.Var()] {
		panic("invalid substitution")
	}
	if s.frozen[v] || s.eliminated[v] || !s.ValueVar(v).IsUndef() || x.Var() == v {
		panic("invalid substitution")
	}
	cs := append([]*Clause(nil), s.cleanOccurrences(v)...)
	s.eliminated[v] = true
	s.SetDecisionVar(v, false)
	for _, c := range cs {
		row := append([]Lit(nil), c.Lits...)
		for i, p := range row {
			if p.Var() == v {
				row[i] = x.Xor(p.Sign())
			}
		}
		if !s.addDerivedClause(row) {
			s.ok = false
			return false
		}
		s.removeClause(c)
	}
	s.elimination = append(s.elimination, eliminationRecord{variable: v, substitute: x, isSubstitution: true})
	return true
}
func (s *SimpSolver) extendModel() {
	for i := len(s.elimination) - 1; i >= 0; i-- {
		r := s.elimination[i]
		if r.isSubstitution {
			s.Model[r.variable] = s.ModelValue(r.substitute)
			continue
		}
		s.Model[r.variable] = LTrue.Xor(r.defaultLit.Sign())
		for j := len(r.clauses) - 1; j >= 0; j-- {
			row := r.clauses[j]
			sat := false
			for _, p := range row[1:] {
				if s.ModelValue(p) != LFalse {
					sat = true
					break
				}
			}
			if !sat {
				s.Model[r.variable] = LTrue.Xor(row[0].Sign())
			}
		}
	}
}
func (s *SimpSolver) cleanUpClauses() {
	if s.useSimplification {
		for v := range s.occurs {
			s.cleanOccurrences(Var(v))
		}
	}
	dst := s.clauses[:0]
	for _, c := range s.clauses {
		if !c.freed {
			dst = append(dst, c)
		}
	}
	clear(s.clauses[len(dst):])
	s.clauses = dst
}
func (s *SimpSolver) Eliminate(turnOff bool) bool {
	if !s.Simplify() {
		return false
	}
	if !s.useSimplification {
		return true
	}
	if len(s.clauses) <= 4800000 {
		for {
			s.gatherTouchedClauses()
			if !s.backwardSubsumptionCheck() {
				s.ok = false
				break
			}
			if s.interrupted.Load() {
				s.elimHeap.Clear()
				break
			}
			for !s.elimHeap.Empty() {
				v := Var(s.elimHeap.RemoveMin())
				if s.interrupted.Load() {
					break
				}
				if s.eliminated[v] || !s.ValueVar(v).IsUndef() {
					continue
				}
				if s.UseAsymm {
					was := s.frozen[v]
					s.frozen[v] = true
					ok := s.asymmVar(v)
					s.frozen[v] = was
					if !ok {
						s.ok = false
						break
					}
				}
				if s.UseElim && s.ValueVar(v).IsUndef() && !s.frozen[v] && !s.eliminateVar(v) {
					s.ok = false
					break
				}
			}
			if !s.ok {
				break
			}
			touched := false
			for _, b := range s.touched {
				if b {
					touched = true
					break
				}
			}
			if !touched && s.backwardAssignments >= len(s.trail) && s.elimHeap.Empty() {
				break
			}
		}
	}
	s.cleanUpClauses()
	if turnOff {
		s.touched = nil
		s.occurs = nil
		s.literalOccurrences = nil
		s.elimHeap.Clear()
		s.subsumptionQueue = nil
		s.useSimplification = false
		s.removeSatisfied = true
		s.allocator.ExtraClauseField = false
	}
	s.GarbageCollect()
	return s.ok
}

// A resolvent is implied by its parents, which are about to be removed. Running
// the redundant-clause check against those parents would drop every resolvent.
func (s *SimpSolver) addDerivedClause(ps []Lit) bool {
	check := s.UseRCheck
	s.UseRCheck = false
	ok := s.AddClause(ps)
	s.UseRCheck = check
	return ok
}

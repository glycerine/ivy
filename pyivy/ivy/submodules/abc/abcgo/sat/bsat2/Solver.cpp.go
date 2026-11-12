package bsat2

import "slices"

/***************************************************************************************[Solver.cc]
Copyright (c) 2003-2006, Niklas Een, Niklas Sorensson
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

func (s *Solver) AddClause(input ...Lit) bool {
	if s.level() != 0 {
		panic("add clause above level zero")
	}
	if !s.ok {
		return false
	}
	data := append([]Lit(nil), input...)
	slices.Sort(data)
	for _, l := range data {
		if l < 0 {
			panic("negative encoded literal")
		}
		for int(l.Var()) >= s.NumVars() {
			s.NewVariable()
		}
	}
	kept := 0
	previous := LitUndef
	for _, l := range data {
		if s.Value(l) == True || l == previous.Not() {
			return true
		}
		if s.Value(l) != False && l != previous {
			data[kept] = l
			kept++
			previous = l
		}
	}
	data = data[:kept]
	if kept == 0 {
		s.ok = false
	} else if kept == 1 {
		s.uncheckedEnqueue(data[0], CRefUndef)
		s.ok = s.propagate() == CRefUndef
	} else {
		h := s.ca.Alloc(data, false)
		s.clauses = append(s.clauses, h)
		s.attach(h)
	}
	return s.ok
}
func (s *Solver) attach(h CRef) {
	c := s.ca.At(h)
	s.watches.Lists[c.Lits[0].Not()] = append(s.watches.Lists[c.Lits[0].Not()], watcher{h, c.Lits[1]})
	s.watches.Lists[c.Lits[1].Not()] = append(s.watches.Lists[c.Lits[1].Not()], watcher{h, c.Lits[0]})
	if c.Learnt {
		s.LearntLiterals += uint64(len(c.Lits))
	} else {
		s.ClauseLiterals += uint64(len(c.Lits))
	}
}
func (s *Solver) detach(h CRef, strict bool) {
	c := s.ca.At(h)
	for _, l := range c.Lits[:2] {
		index := int(l.Not())
		if strict {
			ws := s.watches.Lists[index]
			for i, w := range ws {
				if w.ref == h {
					copy(ws[i:], ws[i+1:])
					s.watches.Lists[index] = ws[:len(ws)-1]
					break
				}
			}
		} else {
			s.watches.Smudge(index)
		}
	}
	if c.Learnt {
		s.LearntLiterals -= uint64(len(c.Lits))
	} else {
		s.ClauseLiterals -= uint64(len(c.Lits))
	}
}
func (s *Solver) locked(h CRef) bool {
	c := s.ca.At(h)
	return len(c.Lits) > 0 && s.Value(c.Lits[0]) == True && s.vardata[c.Lits[0].Var()].reason == h
}
func (s *Solver) removeClause(h CRef) {
	c := s.ca.At(h)
	s.detach(h, false)
	if s.locked(h) {
		s.vardata[c.Lits[0].Var()].reason = CRefUndef
	}
	c.Mark = 1
	s.ca.Free(h)
}
func (s *Solver) satisfied(c *Clause) bool {
	for _, l := range c.Lits {
		if s.Value(l) == True {
			return true
		}
	}
	return false
}
func (s *Solver) uncheckedEnqueue(l Lit, reason CRef) {
	v := l.Var()
	if s.assigns[v] != Undef {
		panic("enqueue assigned variable")
	}
	s.assigns[v] = True.Xor(l.Sign())
	s.vardata[v] = variableData{reason, s.level()}
	s.trail = append(s.trail, l)
}
func (s *Solver) enqueue(l Lit, reason CRef) bool {
	if s.Value(l) != Undef {
		return s.Value(l) != False
	}
	s.uncheckedEnqueue(l, reason)
	return true
}
func (s *Solver) cancelUntil(level int) {
	if s.level() <= level {
		return
	}
	bound := s.trailLimits[level]
	last := s.trailLimits[len(s.trailLimits)-1]
	for i := len(s.trail) - 1; i >= bound; i-- {
		l := s.trail[i]
		v := l.Var()
		s.assigns[v] = Undef
		if s.PhaseSaving > 1 || s.PhaseSaving == 1 && i > last {
			s.polarity[v] = l.Sign()
		}
		s.insertOrder(v)
	}
	s.qhead = bound
	s.trail = s.trail[:bound]
	s.trailLimits = s.trailLimits[:level]
}
func (s *Solver) pickBranch() Lit {
	next := VarUndef
	if s.random() < s.RandomFrequency && s.order.Len() > 0 {
		next = Var(s.order.At(int(s.random() * float64(s.order.Len()))))
		if s.assigns[next] == Undef && s.decision[next] {
			s.RandomDecisions++
		}
	}
	for next == VarUndef || s.assigns[next] != Undef || !s.decision[next] {
		if s.order.Len() == 0 {
			return LitUndef
		}
		next = Var(s.order.RemoveMin())
	}
	negative := s.polarity[next]
	if s.RandomPolarity {
		negative = s.random() < 0.5
	}
	return MakeLit(next, negative)
}
func (s *Solver) propagate() CRef {
	conflict := CRefUndef
	props := 0
	s.watches.CleanAll()
	for s.qhead < len(s.trail) {
		literal := s.trail[s.qhead]
		s.qhead++
		props++
		ws := s.watches.Lists[literal]
		kept := 0
		for i := 0; i < len(ws); i++ {
			w := ws[i]
			if s.Value(w.blocker) == True {
				ws[kept] = w
				kept++
				continue
			}
			c := s.ca.At(w.ref)
			falseLiteral := literal.Not()
			if c.Lits[0] == falseLiteral {
				c.Lits[0], c.Lits[1] = c.Lits[1], c.Lits[0]
			}
			if c.Lits[1] != falseLiteral {
				panic("watch invariant")
			}
			first := c.Lits[0]
			updated := watcher{w.ref, first}
			if first != w.blocker && s.Value(first) == True {
				ws[kept] = updated
				kept++
				continue
			}
			moved := false
			for k := 2; k < len(c.Lits); k++ {
				if s.Value(c.Lits[k]) != False {
					c.Lits[1], c.Lits[k] = c.Lits[k], falseLiteral
					index := c.Lits[1].Not()
					s.watches.Lists[index] = append(s.watches.Lists[index], updated)
					moved = true
					break
				}
			}
			if moved {
				continue
			}
			ws[kept] = updated
			kept++
			if s.Value(first) == False {
				conflict = w.ref
				s.qhead = len(s.trail)
				kept += copy(ws[kept:], ws[i+1:])
				break
			}
			s.uncheckedEnqueue(first, w.ref)
		}
		s.watches.Lists[literal] = ws[:kept]
	}
	s.Propagations += uint64(props)
	s.simplifyProps -= int64(props)
	return conflict
}

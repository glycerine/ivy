package bsat2

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

func (p *SimpSolver) AddClause(literals ...Lit) bool {
	for _, l := range literals {
		if l < 0 {
			panic("negative encoded literal")
		}
		for int(l.Var()) >= p.NumVars() {
			p.NewVariable()
		}
		if p.eliminated[l.Var()] {
			panic("clause references eliminated variable")
		}
	}
	if p.UseRedundancyCheck && p.implied(literals) {
		return true
	}
	count := len(p.clauses)
	if !p.Solver.AddClause(literals...) {
		return false
	}
	if p.simplifying && len(p.clauses) == count+1 {
		h := p.clauses[len(p.clauses)-1]
		p.queue = append(p.queue, h)
		for _, l := range p.ca.At(h).Lits {
			v := l.Var()
			p.occurrences.Lists[v] = append(p.occurrences.Lists[v], h)
			p.counts[l]++
			p.touched[v] = true
			if p.eliminationHeap.Contains(int32(v)) {
				p.eliminationHeap.Increase(int32(v))
			}
		}
	}
	return true
}
func (p *SimpSolver) removeOriginal(h CRef) {
	c := p.ca.At(h)
	if p.simplifying {
		for _, l := range c.Lits {
			p.counts[l]--
			p.updateElimination(l.Var())
			p.occurrences.Smudge(int(l.Var()))
		}
	}
	p.Solver.removeClause(h)
}
func removeReference(items []CRef, h CRef) []CRef {
	for i, x := range items {
		if x == h {
			copy(items[i:], items[i+1:])
			return items[:len(items)-1]
		}
	}
	return items
}
func (p *SimpSolver) strengthen(h CRef, l Lit) bool {
	c := p.ca.At(h)
	p.queue = append(p.queue, h)
	if len(c.Lits) == 2 {
		p.removeOriginal(h)
		c.Strengthen(l)
	} else {
		p.detach(h, true)
		c.Strengthen(l)
		p.attach(h)
		v := l.Var()
		p.occurrences.Lists[v] = removeReference(p.occurrences.Lists[v], h)
		p.counts[l]--
		p.updateElimination(v)
	}
	if len(c.Lits) == 1 {
		return p.enqueue(c.Lits[0], CRefUndef) && p.propagate() == CRefUndef
	}
	return true
}
func (p *SimpSolver) implied(literals []Lit) bool {
	p.trailLimits = append(p.trailLimits, len(p.trail))
	for _, l := range literals {
		if p.Value(l) == True {
			p.cancelUntil(0)
			return false
		}
		if p.Value(l) == Undef {
			p.uncheckedEnqueue(l.Not(), CRefUndef)
		}
	}
	result := p.propagate() != CRefUndef
	p.cancelUntil(0)
	return result
}
func (p *SimpSolver) gatherTouched() {
	queued := make(map[CRef]bool, len(p.queue))
	for _, h := range p.queue {
		queued[h] = true
	}
	for v, touched := range p.touched {
		if !touched {
			continue
		}
		for _, h := range p.occurrences.Lookup(v) {
			if !queued[h] {
				p.queue = append(p.queue, h)
				queued[h] = true
			}
		}
		p.touched[v] = false
	}
}
func (p *SimpSolver) backwardSubsumption() bool {
	for len(p.queue) > 0 || p.backwardAssignments < len(p.trail) {
		if p.interrupted.Load() {
			p.queue = nil
			p.backwardAssignments = len(p.trail)
			return true
		}
		if len(p.queue) == 0 {
			c := p.ca.At(p.dummy)
			c.Lits[0] = p.trail[p.backwardAssignments]
			p.backwardAssignments++
			c.CalcAbstraction()
			p.queue = append(p.queue, p.dummy)
		}
		h := p.queue[0]
		p.queue = p.queue[1:]
		c := p.ca.At(h)
		if c.Mark != 0 {
			continue
		}
		best := c.Lits[0].Var()
		for _, l := range c.Lits[1:] {
			if len(p.occurrences.Lists[l.Var()]) < len(p.occurrences.Lists[best]) {
				best = l.Var()
			}
		}
		candidates := append([]CRef(nil), p.occurrences.Lookup(int(best))...)
		for _, other := range candidates {
			if c.Mark != 0 {
				break
			}
			d := p.ca.At(other)
			if d.Mark != 0 || other == h || p.SubsumptionLimit != -1 && len(d.Lits) >= p.SubsumptionLimit {
				continue
			}
			l := c.Subsumes(d)
			if l == LitUndef {
				p.removeOriginal(other)
			} else if l != LitError && !p.strengthen(other, l.Not()) {
				return false
			}
		}
	}
	return true
}
func (p *SimpSolver) asymmetric(v Var, h CRef) bool {
	c := p.ca.At(h)
	if c.Mark != 0 || p.satisfied(c) {
		return true
	}
	p.trailLimits = append(p.trailLimits, len(p.trail))
	target := LitUndef
	for _, l := range c.Lits {
		if l.Var() != v && p.Value(l) != False {
			p.uncheckedEnqueue(l.Not(), CRefUndef)
		} else {
			target = l
		}
	}
	conflict := p.propagate() != CRefUndef
	p.cancelUntil(0)
	if conflict {
		p.AsymmetricLiterals++
		return p.strengthen(h, target)
	}
	return true
}
func (p *SimpSolver) asymmetricVariable(v Var) bool {
	if p.assigns[v] != Undef {
		return true
	}
	for _, h := range append([]CRef(nil), p.occurrences.Lookup(int(v))...) {
		if !p.asymmetric(v, h) {
			return false
		}
	}
	return p.backwardSubsumption()
}
func (p *SimpSolver) merge(a, b *Clause, v Var) ([]Lit, bool) {
	p.Merges++
	result := []Lit{}
	seen := map[Lit]bool{}
	for _, c := range []*Clause{a, b} {
		for _, l := range c.Lits {
			if l.Var() == v {
				continue
			}
			if seen[l.Not()] {
				return nil, false
			}
			if !seen[l] {
				result = append(result, l)
				seen[l] = true
			}
		}
	}
	return result, true
}
func (p *SimpSolver) rememberEliminated(v Var, c *Clause) {
	literals := append([]Lit(nil), c.Lits...)
	for i, l := range literals {
		if l.Var() == v {
			literals[0], literals[i] = literals[i], literals[0]
			p.eliminationClauses = append(p.eliminationClauses, literals)
			return
		}
	}
	panic("elimination pivot missing")
}
func (p *SimpSolver) eliminateVariable(v Var) bool {
	occurrences := append([]CRef(nil), p.occurrences.Lookup(int(v))...)
	positive, negative := []CRef{}, []CRef{}
	for _, h := range occurrences {
		isPositive := false
		for _, l := range p.ca.At(h).Lits {
			if l == MakeLit(v, false) {
				isPositive = true
				break
			}
		}
		if isPositive {
			positive = append(positive, h)
		} else {
			negative = append(negative, h)
		}
	}
	count := 0
	for _, a := range positive {
		for _, b := range negative {
			if resolvent, valid := p.merge(p.ca.At(a), p.ca.At(b), v); valid {
				count++
				if count > len(occurrences)+p.Grow || p.ClauseLimit != -1 && len(resolvent) > p.ClauseLimit {
					return true
				}
			}
		}
	}
	p.eliminated[v] = true
	p.SetDecisionVar(v, false)
	p.EliminatedVariables++
	saved := positive
	defaultLiteral := MakeLit(v, true)
	if len(positive) > len(negative) {
		saved = negative
		defaultLiteral = MakeLit(v, false)
	}
	for _, h := range saved {
		p.rememberEliminated(v, p.ca.At(h))
	}
	p.eliminationClauses = append(p.eliminationClauses, []Lit{defaultLiteral})
	for _, h := range occurrences {
		p.removeOriginal(h)
	}
	for _, a := range positive {
		for _, b := range negative {
			if literals, valid := p.merge(p.ca.At(a), p.ca.At(b), v); valid && !p.AddClause(literals...) {
				return false
			}
		}
	}
	p.occurrences.Lists[v] = nil
	return p.backwardSubsumption()
}
func (p *SimpSolver) extendModel() {
	for i := len(p.eliminationClauses) - 1; i >= 0; i-- {
		literals := p.eliminationClauses[i]
		satisfied := false
		for _, l := range literals[1:] {
			if p.ModelValue(l) != False {
				satisfied = true
				break
			}
		}
		if !satisfied {
			l := literals[0]
			p.Model[l.Var()] = True.Xor(l.Sign())
		}
	}
}
func (p *SimpSolver) cleanOriginals() {
	p.occurrences.CleanAll()
	n := 0
	for _, h := range p.clauses {
		if p.ca.At(h).Mark == 0 {
			p.clauses[n] = h
			n++
		}
	}
	p.clauses = p.clauses[:n]
}
func (p *SimpSolver) relocateSimplification(to *ClauseAllocator) {
	if !p.simplifying {
		return
	}
	p.cleanOriginals()
	for i := range p.occurrences.Lists {
		for j := range p.occurrences.Lists[i] {
			p.ca.Relocate(&p.occurrences.Lists[i][j], to)
		}
	}
	n := 0
	for _, h := range p.queue {
		if p.ca.At(h).Mark == 0 {
			p.queue[n] = h
			n++
		}
	}
	p.queue = p.queue[:n]
	for i := range p.queue {
		p.ca.Relocate(&p.queue[i], to)
	}
	p.ca.Relocate(&p.dummy, to)
}
func (p *SimpSolver) Eliminate(turnOff bool) bool {
	if !p.Simplify() {
		return false
	}
	if !p.simplifying {
		return true
	}
	for {
		p.gatherTouched()
		if !p.backwardSubsumption() {
			p.ok = false
			break
		}
		if p.interrupted.Load() {
			p.eliminationHeap.Clear(false)
			break
		}
		if p.eliminationHeap.Len() == 0 {
			break
		}
		for p.eliminationHeap.Len() > 0 {
			v := Var(p.eliminationHeap.RemoveMin())
			if p.interrupted.Load() {
				break
			}
			if p.eliminated[v] || p.assigns[v] != Undef {
				continue
			}
			if p.UseAsymmetric {
				frozen := p.frozen[v]
				p.frozen[v] = true
				valid := p.asymmetricVariable(v)
				p.frozen[v] = frozen
				if !valid {
					p.ok = false
					break
				}
			}
			if p.UseElimination && p.assigns[v] == Undef && !p.frozen[v] && !p.eliminateVariable(v) {
				p.ok = false
				break
			}
			if float64(p.ca.Wasted()) > float64(p.ca.Size())*p.SimplificationGarbageFraction {
				p.GarbageCollect()
			}
		}
		if !p.ok {
			break
		}
	}
	if turnOff {
		p.touched = nil
		p.occurrences.Clear()
		p.counts = nil
		p.eliminationHeap.Clear(true)
		p.queue = nil
		p.simplifying = false
		p.removeSatisfiedOriginal = true
		p.ca.ExtraClauseField = false
		p.rebuildOrder()
		p.GarbageCollect()
	} else {
		p.cleanOriginals()
		p.checkGarbage()
	}
	return p.ok
}
func (p *SimpSolver) SolveSimplified(assumptions []Lit, doSimplify, turnOff bool) LBool {
	for _, l := range assumptions {
		if l < 0 || int(l.Var()) >= p.NumVars() {
			panic("assumption outside solver")
		}
		if p.eliminated[l.Var()] {
			panic("assumption references eliminated variable")
		}
	}
	p.Model = nil
	p.Conflict = nil
	extraFrozen := []Var{}
	if doSimplify && p.simplifying {
		for _, l := range assumptions {
			v := l.Var()
			if !p.frozen[v] {
				p.SetFrozen(v, true)
				extraFrozen = append(extraFrozen, v)
			}
		}
		if !p.Eliminate(turnOff) {
			for _, v := range extraFrozen {
				p.SetFrozen(v, false)
			}
			return False
		}
	}
	status := p.Solver.SolveLimited(assumptions...)
	if status == True {
		p.extendModel()
	}
	for _, v := range extraFrozen {
		p.SetFrozen(v, false)
	}
	return status
}
func (p *SimpSolver) Solve(assumptions ...Lit) LBool {
	p.BudgetOff()
	return p.SolveSimplified(assumptions, true, false)
}
func (p *SimpSolver) SolveLimited(assumptions ...Lit) LBool {
	return p.SolveSimplified(assumptions, true, false)
}

// Substitute removes v by equating it to replacement, preserving reconstruction
// records so the returned model still assigns the original variable.
func (p *SimpSolver) Substitute(v Var, replacement Lit) bool {
	if !p.simplifying || p.frozen[v] || p.eliminated[v] || p.assigns[v] != Undef || replacement.Var() == v {
		panic("invalid variable substitution")
	}
	if replacement < 0 || int(replacement.Var()) >= p.NumVars() || p.eliminated[replacement.Var()] {
		panic("invalid substitution literal")
	}
	if !p.ok {
		return false
	}
	p.eliminated[v] = true
	p.SetDecisionVar(v, false)
	positive := MakeLit(v, false)
	p.eliminationClauses = append(p.eliminationClauses, []Lit{positive, replacement.Not()}, []Lit{positive.Not(), replacement})
	for _, h := range append([]CRef(nil), p.occurrences.Lookup(int(v))...) {
		literals := append([]Lit(nil), p.ca.At(h).Lits...)
		for i, l := range literals {
			if l.Var() == v {
				literals[i] = replacement
				if l.Sign() {
					literals[i] = literals[i].Not()
				}
			}
		}
		p.removeOriginal(h)
		if !p.AddClause(literals...) {
			p.ok = false
			return false
		}
	}
	return true
}

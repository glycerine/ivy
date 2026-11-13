package glucose

/***************************************************************************************[Solver.cc]
 Glucose -- Copyright (c) 2013, Gilles Audemard, Laurent Simon
                                CRIL - Univ. Artois, France
                                LRI  - Univ. Paris Sud, France

Glucose sources are based on MiniSat (see below MiniSat copyrights). Permissions and copyrights of
Glucose are exactly the same as Minisat on which it is based on. (see below).

---------------

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

import (
	"fmt"
	"io"
	"os"
	"sort"
)

func (s *Solver) proofClause(c []Lit, delete bool) {
	if !s.CertifiedUNSAT || s.CertifiedOutput == nil {
		return
	}
	if delete {
		fmt.Fprint(s.CertifiedOutput, "d ")
	}
	for _, p := range c {
		n := int(p.Var()) + 1
		if p.Sign() {
			n = -n
		}
		fmt.Fprintf(s.CertifiedOutput, "%d ", n)
	}
	fmt.Fprintln(s.CertifiedOutput, "0")
}
func (s *Solver) AddClause(input []Lit) bool {
	if s.DecisionLevel() != 0 {
		panic("adding clause above root level")
	}
	if !s.ok {
		return false
	}
	ps := append([]Lit(nil), input...)
	sort.Slice(ps, func(i, j int) bool { return ps[i] < ps[j] })
	out := ps[:0]
	previous := LitUndef
	for _, p := range ps {
		if p < 0 || int(p.Var()) >= s.NVars() {
			panic("clause variable out of range")
		}
		if s.ValueLit(p) == LTrue || p == previous.Not() {
			return true
		}
		if p != previous && s.ValueLit(p) != LFalse {
			out = append(out, p)
		}
		previous = p
	}
	if len(out) != len(input) {
		s.proofClause(out, false)
		s.proofClause(input, true)
	}
	switch len(out) {
	case 0:
		s.ok = false
	case 1:
		s.uncheckedEnqueue(out[0], nil)
		s.ok = s.propagate() == nil
	default:
		c := s.allocator.Alloc(out, false)
		s.clauses = append(s.clauses, c)
		s.attachClause(c)
	}
	return s.ok
}
func (s *Solver) attachClause(c *Clause) {
	a, b := c.Lits[0], c.Lits[1]
	if len(c.Lits) == 2 {
		s.watchesBin[a.Not()] = append(s.watchesBin[a.Not()], Watcher{c, b})
		s.watchesBin[b.Not()] = append(s.watchesBin[b.Not()], Watcher{c, a})
	} else {
		s.watches[a.Not()] = append(s.watches[a.Not()], Watcher{c, b})
		s.watches[b.Not()] = append(s.watches[b.Not()], Watcher{c, a})
	}
}
func (s *Solver) propagate() *Clause {
	var conflict *Clause
	props := 0
	for s.qhead < len(s.trail) {
		p := s.trail[s.qhead]
		s.qhead++
		props++

		for _, w := range s.watchesBin[p] {
			if w.Clause.freed {
				continue
			}
			v := s.ValueLit(w.Blocker)
			if v == LFalse {
				return w.Clause
			}
			if v.IsUndef() {
				s.uncheckedEnqueue(w.Blocker, w.Clause)
			}
		}
		ws := s.watches[p]
		kept := ws[:0]
		for i, w := range ws {
			c := w.Clause
			if c.freed {
				continue
			}
			if s.ValueLit(w.Blocker) == LTrue {
				kept = append(kept, w)
				continue
			}
			falseLit := p.Not()
			if c.Lits[0] == falseLit {
				c.Lits[0], c.Lits[1] = c.Lits[1], c.Lits[0]
			}
			if c.Lits[1] != falseLit {
				panic("invalid clause watch")
			}
			first := c.Lits[0]
			nw := Watcher{c, first}
			if first != w.Blocker && s.ValueLit(first) == LTrue {
				kept = append(kept, nw)
				continue
			}
			chosen := -1
			for k := 2; k < len(c.Lits); k++ {
				if s.ValueLit(c.Lits[k]) != LFalse {
					chosen = k
					if !s.incremental || s.DecisionLevel() > len(s.assumptions) || s.ValueLit(c.Lits[k]) == LTrue || !s.isSelector(c.Lits[k].Var()) {
						break
					}
				}
			}
			if chosen >= 0 {
				c.Lits[1], c.Lits[chosen] = c.Lits[chosen], falseLit
				s.watches[c.Lits[1].Not()] = append(s.watches[c.Lits[1].Not()], nw)
				continue
			}
			kept = append(kept, nw)
			if s.ValueLit(first) == LFalse {
				conflict = c
				s.qhead = len(s.trail)
				kept = append(kept, ws[i+1:]...)
				break
			}
			s.uncheckedEnqueue(first, c)
		}
		s.watches[p] = kept
	}
	s.Propagations += int64(props)
	s.simpProps -= props
	return conflict
}
func (s *Solver) computeLBD(lits []Lit) int {
	levels := make(map[int]struct{}, len(lits))
	for _, p := range lits {
		if !s.isSelector(p.Var()) {
			levels[s.levels[p.Var()]] = struct{}{}
		}
	}
	return len(levels)
}
func (s *Solver) litRedundant(p Lit, abstract uint32) bool {
	stack := []Lit{p}
	top := len(s.analyzeToClear)
	for len(stack) > 0 {
		p = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		c := s.reasons[p.Var()]
		if len(c.Lits) == 2 && s.ValueLit(c.Lits[0]) == LFalse {
			c.Lits[0], c.Lits[1] = c.Lits[1], c.Lits[0]
		}
		for _, q := range c.Lits[1:] {
			v := q.Var()
			if !s.seen[v] && s.levels[v] > 0 {
				if s.reasons[v] != nil && (uint32(1)<<uint(s.levels[v]&31))&abstract != 0 {
					s.seen[v] = true
					stack = append(stack, q)
					s.analyzeToClear = append(s.analyzeToClear, q)
				} else {
					for _, r := range s.analyzeToClear[top:] {
						s.seen[r.Var()] = false
					}
					s.analyzeToClear = s.analyzeToClear[:top]
					return false
				}
			}
		}
	}
	return true
}
func (s *Solver) minimizeBinary(out []Lit) []Lit {
	if s.computeLBD(out) > s.LBLBDMinimizingClause {
		return out
	}
	marked := make(map[Var]bool, len(out))
	for _, p := range out[1:] {
		marked[p.Var()] = true
	}
	for _, w := range s.watchesBin[out[0].Not()] {
		if !w.Clause.freed && s.ValueLit(w.Blocker) == LTrue {
			delete(marked, w.Blocker.Var())
		}
	}
	removed := 0
	for _, p := range out[1:] {
		if !marked[p.Var()] {
			removed++
		}
	}
	if removed == 0 {
		return out
	}
	s.ReducedClauses++
	last := len(out) - 1
	for i := 1; i < len(out)-removed; i++ {
		if !marked[out[i].Var()] {
			out[i], out[last] = out[last], out[i]
			last--
			i--
		}
	}
	return out[:len(out)-removed]
}
func (s *Solver) analyze(conflict *Clause) (out []Lit, backtrack, lbd, withoutSelectors int) {
	out = append(out, LitUndef)
	var selectors, lastLevel []Lit
	path := 0
	p := LitUndef
	index := len(s.trail) - 1
	for {
		c := conflict
		if p != LitUndef {
			c = s.reasons[p.Var()]
		}
		if p != LitUndef && len(c.Lits) == 2 && s.ValueLit(c.Lits[0]) == LFalse {
			c.Lits[0], c.Lits[1] = c.Lits[1], c.Lits[0]
		}
		if c.learnt {
			s.bumpClause(c)
			if c.lbd > 2 {
				n := s.computeLBD(c.Lits)
				if n+1 < c.lbd {
					if c.lbd <= s.LBLBDFrozenClause {
						c.canBeDel = false
					}
					c.lbd = n
				}
			}
		}
		start := 0
		if p != LitUndef {
			start = 1
		}
		for _, q := range c.Lits[start:] {
			v := q.Var()
			if !s.seen[v] && s.levels[v] > 0 {
				if !s.isSelector(v) {
					s.bumpVar(v)
				}
				s.seen[v] = true
				if s.levels[v] >= s.DecisionLevel() {
					path++
					if !s.isSelector(v) && s.reasons[v] != nil && s.reasons[v].learnt {
						lastLevel = append(lastLevel, q)
					}
				} else if s.isSelector(v) {
					selectors = append(selectors, q)
				} else {
					out = append(out, q)
				}
			}
		}
		for !s.seen[s.trail[index].Var()] {
			index--
		}
		p = s.trail[index]
		index--
		conflict = s.reasons[p.Var()]
		s.seen[p.Var()] = false
		path--
		if path == 0 {
			break
		}
	}
	out[0] = p.Not()
	out = append(out, selectors...)
	s.analyzeToClear = append(s.analyzeToClear[:0], out...)
	if s.CCMinMode == 2 {
		var abstract uint32
		for _, q := range out[1:] {
			abstract |= uint32(1) << uint(s.levels[q.Var()]&31)
		}
		dst := out[:1]
		for _, q := range out[1:] {
			if s.reasons[q.Var()] == nil || !s.litRedundant(q, abstract) {
				dst = append(dst, q)
			}
		}
		out = dst
	} else if s.CCMinMode == 1 {
		dst := out[:1]
		for _, q := range out[1:] {
			c := s.reasons[q.Var()]
			keep := c == nil
			if c != nil {
				start := 1
				if len(c.Lits) == 2 {
					start = 0
				}
				for _, r := range c.Lits[start:] {
					if !s.seen[r.Var()] && s.levels[r.Var()] > 0 {
						keep = true
						break
					}
				}
			}
			if keep {
				dst = append(dst, q)
			}
		}
		out = dst
	}
	if !s.incremental && len(out) <= s.LBSizeMinimizingClause {
		out = s.minimizeBinary(out)
	}
	if len(out) > 1 {
		max := 1
		for i := 2; i < len(out); i++ {
			if s.levels[out[i].Var()] > s.levels[out[max].Var()] {
				max = i
			}
		}
		out[1], out[max] = out[max], out[1]
		backtrack = s.levels[out[1].Var()]
	}
	withoutSelectors = len(out)
	if s.incremental {
		withoutSelectors = 0
		for i, q := range out {
			if !s.isSelector(q.Var()) {
				withoutSelectors++
			} else if i > 0 {
				break
			}
		}
	}
	lbd = s.computeLBD(out)
	for _, q := range lastLevel {
		if s.reasons[q.Var()].lbd < lbd {
			s.bumpVar(q.Var())
		}
	}
	for _, q := range s.analyzeToClear {
		s.seen[q.Var()] = false
	}
	for _, q := range selectors {
		s.seen[q.Var()] = false
	}
	return
}

func (s *Solver) analyzeFinal(p Lit) {
	s.Conflict = append(s.Conflict[:0], p)
	if s.DecisionLevel() == 0 {
		return
	}
	s.seen[p.Var()] = true
	for i := len(s.trail) - 1; i >= s.trailLim[0]; i-- {
		v := s.trail[i].Var()
		if !s.seen[v] {
			continue
		}
		c := s.reasons[v]
		if c == nil {
			s.Conflict = append(s.Conflict, s.trail[i].Not())
		} else {
			start := 1
			if len(c.Lits) == 2 {
				start = 0
			}
			for _, q := range c.Lits[start:] {
				if s.levels[q.Var()] > 0 {
					s.seen[q.Var()] = true
				}
			}
		}
		s.seen[v] = false
	}
	s.seen[p.Var()] = false
}
func (s *Solver) locked(c *Clause) bool {
	for _, p := range c.Lits {
		if s.ValueLit(p) == LTrue && s.reasons[p.Var()] == c {
			return true
		}
	}
	return false
}
func (s *Solver) removeClause(c *Clause) {
	s.proofClause(c.Lits, true)
	s.detachClause(c)
	for _, p := range c.Lits {
		if s.reasons[p.Var()] == c {
			s.reasons[p.Var()] = nil
		}
	}
	s.allocator.Free(c)
}
func (s *Solver) satisfied(c *Clause) bool {
	if s.incremental {
		return s.ValueLit(c.Lits[0]) == LTrue || s.ValueLit(c.Lits[1]) == LTrue
	}
	for _, p := range c.Lits {
		if s.ValueLit(p) == LTrue {
			return true
		}
	}
	return false
}

func (s *Solver) Simplify() bool {
	if s.DecisionLevel() != 0 {
		panic("simplifying above root")
	}
	if !s.ok || s.propagate() != nil {
		s.ok = false
		return false
	}
	if s.simpAssigns == len(s.trail) || s.simpProps > 0 {
		return true
	}
	clean := func(cs []*Clause) []*Clause {
		dst := cs[:0]
		for _, c := range cs {
			if s.satisfied(c) {
				s.removeClause(c)
			} else {
				dst = append(dst, c)
			}
		}
		return dst
	}
	s.learnts = clean(s.learnts)
	if s.removeSatisfied {
		s.clauses = clean(s.clauses)
	}
	var candidates []int
	for v := range s.assigns {
		if s.decision[v] && s.assigns[v].IsUndef() {
			candidates = append(candidates, v)
		}
	}
	s.orderHeap.Build(candidates)
	s.simpAssigns = len(s.trail)
	s.simpProps = 0
	for _, c := range s.clauses {
		s.simpProps += len(c.Lits)
	}
	for _, c := range s.learnts {
		s.simpProps += len(c.Lits)
	}
	return true
}
func (s *Solver) reduceDB() {
	s.ReduceDBs++
	sort.Slice(s.learnts, func(i, j int) bool {
		a, b := s.learnts[i], s.learnts[j]
		if len(a.Lits) == 2 {
			return false
		}
		if len(b.Lits) == 2 {
			return true
		}
		if a.lbd != b.lbd {
			return a.lbd > b.lbd
		}
		return a.activity < b.activity
	})
	if len(s.learnts) == 0 {
		return
	}
	if s.learnts[len(s.learnts)/2].lbd <= 3 {
		s.nextReduce += s.SpecialIncReduceDB
	}
	if s.learnts[len(s.learnts)-1].lbd <= 5 {
		s.nextReduce += s.SpecialIncReduceDB
	}
	limit := len(s.learnts) / 2
	dst := s.learnts[:0]
	for i, c := range s.learnts {
		if c.lbd > 2 && len(c.Lits) > 2 && c.canBeDel && !s.locked(c) && i < limit {
			s.removeClause(c)
			s.RemovedClauses++
		} else {
			if !c.canBeDel {
				limit++
			}
			c.canBeDel = true
			dst = append(dst, c)
		}
	}
	s.learnts = dst
}
func (s *Solver) pickBranchLit() Lit {
	next := -1
	if s.RandomVarFreq > 0 && s.random() < s.RandomVarFreq && !s.orderHeap.Empty() {
		v := s.orderHeap.At(int(s.random() * float64(s.orderHeap.Size())))
		if s.assigns[v].IsUndef() && s.decision[v] {
			next = v
			s.RandomDecisions++
		}
	}
	for next < 0 || !s.assigns[next].IsUndef() || !s.decision[next] {
		if s.orderHeap.Empty() {
			return LitUndef
		}
		next = s.orderHeap.RemoveMin()
	}
	sign := s.polarity[next]
	if s.RandomPolarity {
		sign = s.random() < .5
	}
	return MkLit(Var(next), sign)
}
func (s *Solver) search() LBool {
	s.Starts++
	for {
		conflict := s.propagate()
		if conflict != nil {
			s.Conflicts++
			s.ConflictsRestarts++
			if s.Conflicts%5000 == 0 && s.VarDecay < .95 {
				s.VarDecay += .01
			}
			if s.DecisionLevel() == 0 {
				return LFalse
			}
			s.trailQueue.Push(uint64(len(s.trail)))
			if s.ConflictsRestarts > 10000 && s.lbdQueue.Valid() && float64(len(s.trail)) > s.R*float64(s.trailQueue.Avg()) {
				s.lbdQueue.FastClear()
			}
			learnt, bt, lbd, without := s.analyze(conflict)
			s.lbdQueue.Push(uint64(lbd))
			s.sumLBD += float64(lbd)
			s.CancelUntil(bt)
			s.proofClause(learnt, false)
			if len(learnt) == 1 {
				s.uncheckedEnqueue(learnt[0], nil)
			} else {
				c := s.allocator.Alloc(learnt, true)
				c.lbd = lbd
				c.sizeWithoutSelectors = without
				s.learnts = append(s.learnts, c)
				s.attachClause(c)
				s.bumpClause(c)
				s.uncheckedEnqueue(learnt[0], c)
			}
			s.varInc /= s.VarDecay
			s.clauseInc /= s.ClauseDecay
		} else {
			if (s.ConflictsRestarts > 0 && s.lbdQueue.Valid() && float64(s.lbdQueue.Avg())*s.K > s.sumLBD/float64(s.ConflictsRestarts)) || (s.Stop != nil && s.Stop()) {
				s.lbdQueue.FastClear()
				bt := 0
				if s.incremental {
					bt = s.DecisionLevel()
					if bt > len(s.assumptions) {
						bt = len(s.assumptions)
					}
				}
				s.CancelUntil(bt)
				return LUndef
			}
			if s.DecisionLevel() == 0 && !s.Simplify() {
				return LFalse
			}
			if s.nextReduce > 0 && s.Conflicts >= int64(s.curRestart*s.nextReduce) && len(s.learnts) > 0 {
				s.curRestart = int(s.Conflicts)/s.nextReduce + 1
				s.reduceDB()
				s.nextReduce += s.IncReduceDB
			}
			next := LitUndef
			for s.DecisionLevel() < len(s.assumptions) {
				p := s.assumptions[s.DecisionLevel()]
				v := s.ValueLit(p)
				if v == LTrue {
					s.newDecisionLevel()
				} else if v == LFalse {
					s.analyzeFinal(p.Not())
					return LFalse
				} else {
					next = p
					break
				}
			}

			if next == LitUndef {
				s.Decisions++
				next = s.pickBranchLit()
				if next == LitUndef {
					return LTrue
				}
			}
			s.newDecisionLevel()
			s.uncheckedEnqueue(next, nil)
		}
	}
}
func (s *Solver) SolveLimited(assumptions []Lit) LBool {
	if s.incremental && s.CertifiedUNSAT {
		panic("incremental solving and certified UNSAT cannot be combined")
	}

	s.Solves++

	s.Model = nil
	s.Conflict = nil
	if !s.ok {
		return LFalse
	}
	s.assumptions = append(s.assumptions[:0], assumptions...)
	for _, p := range assumptions {
		if p < 0 || int(p.Var()) >= s.NVars() {
			panic("assumption variable out of range")
		}
	}
	status := LUndef
	for status.IsUndef() {
		status = s.search()
		if !s.WithinBudget() {
			break
		}
	}
	if status == LTrue {

		s.Model = append([]LBool(nil), s.assigns...)
	} else if status == LFalse && len(s.Conflict) == 0 {
		s.ok = false
		s.proofClause(nil, false)
	}
	s.CancelUntil(0)
	if s.CnfCallback != nil {
		if !s.TerminateSearchEarly {
			message := -1
			var model []int
			if status == LTrue {
				message = 1
				model = s.GetCex()
			} else if status == LFalse {
				message = 0
			}
			if s.CnfCallback(s.CnfManager, message, model) != 0 {
				panic("solver callback returned nonzero status")
			}
		} else {
			s.TerminateSearchEarly = false
		}
	}
	return status
}
func (s *Solver) Solve(assumptions []Lit) bool {
	s.BudgetOff()
	return s.SolveLimited(assumptions) == LTrue
}

// ToDimacs writes the root-simplified formula plus the supplied assumptions.
// Variables are compacted in first-clause occurrence order, as in Glucose.
func (s *Solver) ToDimacs(w io.Writer, assumptions []Lit) error {
	for _, p := range assumptions {
		if p < 0 || int(p.Var()) >= s.NVars() {
			return fmt.Errorf("DIMACS assumption variable out of range: %d", p)
		}
	}
	for _, p := range assumptions {
		if s.ValueLit(p) == LFalse {
			_, err := io.WriteString(w, "p cnf 1 2\n1 0\n-1 0\n")
			return err
		}
	}
	if !s.ok {
		_, err := io.WriteString(w, "p cnf 1 2\n1 0\n-1 0\n")
		return err
	}
	mapped := make(map[Var]int)
	next := 1
	mapVar := func(v Var) int {
		if n, ok := mapped[v]; ok {
			return n
		}
		n := next
		next++
		mapped[v] = n
		return n
	}
	var clauses [][]Lit
	for _, c := range s.clauses {
		if s.satisfied(c) {
			continue
		}
		var row []Lit
		for _, p := range c.Lits {
			if s.ValueLit(p) != LFalse {
				mapVar(p.Var())
				row = append(row, p)
			}
		}
		clauses = append(clauses, row)
	}
	for _, p := range assumptions {
		mapVar(p.Var())
	}
	if _, err := fmt.Fprintf(w, "p cnf %d %d\n", len(mapped), len(clauses)+len(assumptions)); err != nil {
		return err
	}
	writeClause := func(row []Lit) error {
		for _, p := range row {
			n := mapped[p.Var()]
			if p.Sign() {
				n = -n
			}
			if _, err := fmt.Fprintf(w, "%d ", n); err != nil {
				return err
			}
		}
		_, err := io.WriteString(w, "0\n")
		return err
	}
	for _, p := range assumptions {
		if err := writeClause([]Lit{p}); err != nil {
			return err
		}
	}
	for _, c := range clauses {
		if err := writeClause(c); err != nil {
			return err
		}
	}
	return nil
}
func (s *Solver) ToDimacsFile(name string, assumptions []Lit) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	err = s.ToDimacs(f, assumptions)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}

// detachClause removes both current watches before an in-place strengthening.
// Go uses eager removal: unlike arena offsets, clause pointers must not retain
// obsolete watch membership when a preprocessor reuses the same clause object.
func (s *Solver) detachClause(c *Clause) {
	remove := func(ws []Watcher) []Watcher {
		for i, w := range ws {
			if w.Clause == c {
				copy(ws[i:], ws[i+1:])
				ws[len(ws)-1] = Watcher{}
				return ws[:len(ws)-1]
			}
		}
		return ws
	}
	a, b := c.Lits[0].Not(), c.Lits[1].Not()
	if len(c.Lits) == 2 {
		s.watchesBin[a] = remove(s.watchesBin[a])
		s.watchesBin[b] = remove(s.watchesBin[b])
	} else {
		s.watches[a] = remove(s.watches[a])
		s.watches[b] = remove(s.watches[b])
	}
}

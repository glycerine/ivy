package bsat2

import "slices"

func (s *Solver) reduceDB() {
	if len(s.learnts) == 0 {
		return
	}
	limit := s.clauseInc / float64(len(s.learnts))
	slices.SortFunc(s.learnts, s.compareLearned)
	n := 0
	for i, h := range s.learnts {
		c := s.ca.At(h)
		if len(c.Lits) > 2 && !s.locked(h) && (i < len(s.learnts)/2 || float64(c.Activity) < limit) {
			s.removeClause(h)
		} else {
			s.learnts[n] = h
			n++
		}
	}
	s.learnts = s.learnts[:n]
	s.checkGarbage()
}
func (s *Solver) compareLearned(a, b CRef) int {
	x, y := s.ca.At(a), s.ca.At(b)
	if len(x.Lits) == 2 {
		if len(y.Lits) == 2 {
			return 0
		}
		return 1
	}
	if len(y.Lits) == 2 {
		return -1
	}
	if x.Activity < y.Activity {
		return -1
	}
	if x.Activity > y.Activity {
		return 1
	}
	return 0
}
func (s *Solver) removeSatisfied(clauses []CRef) []CRef {
	n := 0
	for _, h := range clauses {
		if s.satisfied(s.ca.At(h)) {
			s.removeClause(h)
		} else {
			clauses[n] = h
			n++
		}
	}
	return clauses[:n]
}
func (s *Solver) rebuildOrder() {
	variables := []int32{}
	for v := range s.assigns {
		if s.decision[v] && s.assigns[v] == Undef {
			variables = append(variables, int32(v))
		}
	}
	s.order.Build(variables)
}
func (s *Solver) Simplify() bool {
	if s.level() != 0 {
		panic("simplify above level zero")
	}
	if !s.ok || s.propagate() != CRefUndef {
		s.ok = false
		return false
	}
	if len(s.trail) == s.simplifyAssigns || s.simplifyProps > 0 {
		return true
	}
	s.learnts = s.removeSatisfied(s.learnts)
	if s.removeSatisfiedOriginal {
		s.clauses = s.removeSatisfied(s.clauses)
	}
	s.checkGarbage()
	s.rebuildOrder()
	s.simplifyAssigns = len(s.trail)
	s.simplifyProps = int64(s.ClauseLiterals + s.LearntLiterals)
	return true
}
func (s *Solver) checkGarbage() {
	if float64(s.ca.Wasted()) > float64(s.ca.Size())*s.GarbageFraction {
		s.GarbageCollect()
	}
}
func (s *Solver) GarbageCollect() {
	to := NewClauseAllocator()
	to.ExtraClauseField = s.ca.ExtraClauseField
	s.watches.CleanAll()
	if s.relocateExtra != nil {
		s.relocateExtra(to)
	}
	for i := range s.watches.Lists {
		for j := range s.watches.Lists[i] {
			s.ca.Relocate(&s.watches.Lists[i][j].ref, to)
		}
	}
	for _, l := range s.trail {
		v := l.Var()
		h := s.vardata[v].reason
		if h != CRefUndef && (s.ca.At(h).relocated || s.locked(h)) {
			s.ca.Relocate(&h, to)
			s.vardata[v].reason = h
		}
	}
	for i := range s.learnts {
		s.ca.Relocate(&s.learnts[i], to)
	}
	for i := range s.clauses {
		s.ca.Relocate(&s.clauses[i], to)
	}
	s.ca = to
}

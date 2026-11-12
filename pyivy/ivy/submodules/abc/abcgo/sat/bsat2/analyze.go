package bsat2

// analyze performs MiniSAT's first-UIP traversal, then the configured basic or
// deep clause minimization. Every temporary seen bit is cleared before return.
func (s *Solver) analyze(conflict CRef) ([]Lit, int) {
	learned := []Lit{LitUndef}
	pending := 0
	pivot := LitUndef
	index := len(s.trail) - 1
	for {
		if conflict == CRefUndef {
			panic("first-UIP reason missing")
		}
		c := s.ca.At(conflict)
		if c.Learnt {
			s.bumpClause(c)
		}
		start := 0
		if pivot != LitUndef {
			start = 1
		}
		for _, l := range c.Lits[start:] {
			v := l.Var()
			if s.seen[v] || s.vardata[v].level == 0 {
				continue
			}
			s.bumpVariable(v)
			s.seen[v] = true
			if s.vardata[v].level >= s.level() {
				pending++
			} else {
				learned = append(learned, l)
			}
		}
		for index >= 0 && !s.seen[s.trail[index].Var()] {
			index--
		}
		if index < 0 {
			panic("first-UIP pivot missing")
		}
		pivot = s.trail[index]
		index--
		conflict = s.vardata[pivot.Var()].reason
		s.seen[pivot.Var()] = false
		pending--
		if pending == 0 {
			break
		}
	}
	learned[0] = pivot.Not()
	s.analyzeClear = append(s.analyzeClear[:0], learned...)
	originalSize := len(learned)
	kept := 1
	switch s.Minimization {
	case 2:
		abstract := uint32(0)
		for _, l := range learned[1:] {
			abstract |= 1 << uint(s.vardata[l.Var()].level&31)
		}
		for _, l := range learned[1:] {
			if s.vardata[l.Var()].reason == CRefUndef || !s.literalRedundant(l, abstract) {
				learned[kept] = l
				kept++
			}
		}
	case 1:
		for _, l := range learned[1:] {
			reason := s.vardata[l.Var()].reason
			retain := reason == CRefUndef
			if !retain {
				for _, q := range s.ca.At(reason).Lits[1:] {
					if !s.seen[q.Var()] && s.vardata[q.Var()].level > 0 {
						retain = true
						break
					}
				}
			}
			if retain {
				learned[kept] = l
				kept++
			}
		}
	default:
		kept = len(learned)
	}
	s.MaxLiterals += uint64(originalSize)
	learned = learned[:kept]
	s.TotalLiterals += uint64(kept)
	backtrack := 0
	if kept > 1 {
		highest := 1
		for i := 2; i < kept; i++ {
			if s.vardata[learned[i].Var()].level > s.vardata[learned[highest].Var()].level {
				highest = i
			}
		}
		learned[1], learned[highest] = learned[highest], learned[1]
		backtrack = s.vardata[learned[1].Var()].level
	}
	for _, l := range s.analyzeClear {
		s.seen[l.Var()] = false
	}
	return learned, backtrack
}
func (s *Solver) literalRedundant(literal Lit, abstract uint32) bool {
	s.analyzeStack = append(s.analyzeStack[:0], literal)
	top := len(s.analyzeClear)
	for len(s.analyzeStack) > 0 {
		last := len(s.analyzeStack) - 1
		l := s.analyzeStack[last]
		s.analyzeStack = s.analyzeStack[:last]
		reason := s.vardata[l.Var()].reason
		if reason == CRefUndef {
			panic("redundancy reason missing")
		}
		for _, q := range s.ca.At(reason).Lits[1:] {
			v := q.Var()
			if s.seen[v] || s.vardata[v].level == 0 {
				continue
			}
			if s.vardata[v].reason != CRefUndef && abstract&(1<<uint(s.vardata[v].level&31)) != 0 {
				s.seen[v] = true
				s.analyzeStack = append(s.analyzeStack, q)
				s.analyzeClear = append(s.analyzeClear, q)
			} else {
				for _, added := range s.analyzeClear[top:] {
					s.seen[added.Var()] = false
				}
				s.analyzeClear = s.analyzeClear[:top]
				return false
			}
		}
	}
	return true
}
func (s *Solver) analyzeFinal(literal Lit) {
	s.Conflict = append(s.Conflict[:0], literal)
	if s.level() == 0 {
		return
	}
	s.seen[literal.Var()] = true
	for i := len(s.trail) - 1; i >= s.trailLimits[0]; i-- {
		v := s.trail[i].Var()
		if !s.seen[v] {
			continue
		}
		reason := s.vardata[v].reason
		if reason == CRefUndef {
			s.Conflict = append(s.Conflict, s.trail[i].Not())
		} else {
			for _, q := range s.ca.At(reason).Lits[1:] {
				if s.vardata[q.Var()].level > 0 {
					s.seen[q.Var()] = true
				}
			}
		}
		s.seen[v] = false
	}
	s.seen[literal.Var()] = false
}

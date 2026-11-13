package minisatp

import "sort"

// analyze follows Solver::analyze's first-UIP clause and expensive recursive
// conflict-clause minimization, recording every resolution in proof order.
func (s *solver) analyze(conflict *clause) ([]int32, int) {
	s.proof.beginChain(conflict.id)
	learnt := []int32{0}
	pathCount, backtrack := 0, 0
	index := len(s.trail) - 1
	p := int32(-1)
	for {
		if s.expired() {
			return nil, 0
		}
		if conflict.learnt {
			conflict.activity += s.claInc
			if conflict.activity > 1e20 {
				for _, c := range s.learnts {
					c.activity *= 1e-20
				}
				s.claInc *= 1e-20
			}
		}
		start := 0
		if p >= 0 {
			start = 1
		}
		for _, q := range conflict.lits[start:] {
			v := q >> 1
			if s.seen[v] {
				continue
			}
			if s.level[v] > 0 {
				s.activity[v] += s.varInc
				if s.activity[v] > 1e100 {
					for i := range s.activity {
						s.activity[i] *= 1e-100
					}
					s.varInc *= 1e-100
				}
				if s.order.index[v] != 0 {
					s.order.up(s.order.index[v])
				}
				s.seen[v] = true
				if s.level[v] == len(s.limits) {
					pathCount++
				} else {
					learnt = append(learnt, q)
					if s.level[v] > backtrack {
						backtrack = s.level[v]
					}
				}
			} else {
				s.proof.resolve(s.unitID[v], v)
			}
		}
		for !s.seen[s.trail[index]>>1] {
			index--
		}
		p = s.trail[index]
		index--
		conflict = s.reason[p>>1]
		s.seen[p>>1] = false
		pathCount--
		if pathCount == 0 {
			break
		}
		s.proof.resolve(conflict.id, p>>1)
	}
	learnt[0] = p ^ 1
	var levels uint32
	for _, lit := range learnt[1:] {
		if s.expired() {
			return nil, 0
		}
		levels |= 1 << (uint(s.level[lit>>1]) & 31)
	}
	toClear := make([]int32, 0)
	kept := learnt[:1]
	for _, lit := range learnt[1:] {
		if s.reason[lit>>1] == nil || !s.removable(lit, levels, &toClear) {
			kept = append(kept, lit)
		}
	}
	sort.Slice(toClear, func(i, j int) bool { return s.trailPos[toClear[i]>>1] > s.trailPos[toClear[j]>>1] })
	for _, lit := range toClear {
		v := lit >> 1
		c := s.reason[v]
		s.proof.resolve(c.id, v)
		for _, q := range c.lits[1:] {
			if s.level[q>>1] == 0 {
				s.proof.resolve(s.unitID[q>>1], q>>1)
			}
		}
	}
	s.proof.endChain()
	for _, lit := range learnt {
		s.seen[lit>>1] = false
	}
	for _, lit := range toClear {
		s.seen[lit>>1] = false
	}
	return kept, backtrack
}

func (s *solver) removable(lit int32, levels uint32, clear *[]int32) bool {
	stack := []int32{lit}
	top := len(*clear)
	for len(stack) != 0 {
		if s.expired() {
			return false
		}
		last := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, q := range s.reason[last>>1].lits[1:] {
			v := q >> 1
			if s.seen[v] || s.level[v] == 0 {
				continue
			}
			if s.reason[v] != nil && levels&(1<<(uint(s.level[v])&31)) != 0 {
				s.seen[v] = true
				stack = append(stack, q)
				*clear = append(*clear, q)
			} else {
				for _, x := range (*clear)[top:] {
					s.seen[x>>1] = false
				}
				*clear = (*clear)[:top]
				return false
			}
		}
	}
	*clear = append(*clear, lit)
	return true
}

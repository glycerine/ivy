// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

func (s *Solver) clauseLBD(lits []int) uint32 {
	levels := map[int]bool{}
	for _, lit := range lits {
		levels[s.levels[lit>>1]] = true
	}
	return uint32(len(levels))
}

// minimizeLearnt follows reason edges to remove literals whose explanations
// are already covered by the learnt clause, then applies binary resolution.
func (s *Solver) minimizeLearnt(lits []int) []int {
	present := make([]bool, len(s.values))
	var mask uint32
	for _, lit := range lits {
		present[lit>>1] = true
	}
	for _, lit := range lits[1:] {
		mask |= 1 << uint(s.levels[lit>>1]&31)
	}
	var removable func(int, map[int]bool) bool
	removable = func(v int, visiting map[int]bool) bool {
		reason := s.reasons[v]
		if reason < 0 {
			return false
		}
		if visiting[v] {
			return false
		}
		visiting[v] = true
		defer delete(visiting, v)
		for _, q := range s.clauses[reason].lits {
			u := q >> 1
			if u == v || s.levels[u] == 0 || present[u] {
				continue
			}
			if mask&(1<<uint(s.levels[u]&31)) == 0 || !removable(u, visiting) {
				return false
			}
		}
		return true
	}
	out := lits[:1]
	for _, lit := range lits[1:] {
		if !removable(lit>>1, map[int]bool{}) {
			out = append(out, lit)
		}
	}
	if uint32(len(out)) > s.opts.BinaryResolutionSize || s.clauseLBD(out) > s.opts.BinaryResolutionLBD {
		return out
	}
	redundant := map[int]bool{}
	asserting := out[0]
	for _, id := range s.watches[asserting^1] {
		c := s.clauses[id]
		if c.deleted || len(c.lits) != 2 {
			continue
		}
		other := -1
		if c.lits[0] == asserting {
			other = c.lits[1]
		} else if c.lits[1] == asserting {
			other = c.lits[0]
		}
		if other >= 0 && s.literalValue(other) == 1 {
			redundant[other^1] = true
		}
	}
	result := out[:1]
	for _, lit := range out[1:] {
		if !redundant[lit] {
			result = append(result, lit)
		}
	}
	return result
}

// analyzeFinal walks reasons back to assumption decisions rather than returning
// an unrelated superset of the assumption vector.
func (s *Solver) analyzeFinal(lit int) {
	s.final = []int{lit}
	if len(s.limits) == 0 {
		return
	}
	seen := make([]bool, len(s.values))
	seen[lit>>1] = true
	for i := len(s.trail) - 1; i >= s.limits[0]; i-- {
		p := s.trail[i]
		v := p >> 1
		if !seen[v] {
			continue
		}
		reason := s.reasons[v]
		if reason < 0 {
			s.final = append(s.final, p^1)
		} else {
			for _, q := range s.clauses[reason].lits {
				u := q >> 1
				if u != v && s.levels[u] > 0 {
					seen[u] = true
				}
			}
		}
		seen[v] = false
	}
}

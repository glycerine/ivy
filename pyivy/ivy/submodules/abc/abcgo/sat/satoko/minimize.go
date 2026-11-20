// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

func (s *Solver) nextStamp(size int) uint32 {
	if len(s.stamps) < size {
		s.stamps = append(s.stamps, make([]uint32, size-len(s.stamps))...)
	}
	s.stamp++
	if s.stamp == 0 {
		clear(s.stamps)
		s.stamp = 1
	}
	return s.stamp
}

func (s *Solver) clauseLBD(lits []int) uint32 {
	stamp := s.nextStamp(s.decisionLevel() + 1)
	var count uint32
	for _, lit := range lits {
		level := s.levels[lit>>1]
		if s.stamps[level] != stamp {
			s.stamps[level] = stamp
			count++
		}
	}
	return count
}

// minimizeLearnt follows reason edges to remove literals whose explanations
// are already covered by the learnt clause, then applies binary resolution.
func (s *Solver) minimizeLearnt(lits []int) []int {
	scratch := &s.analysis
	// Non-asserting clause variables are still marked by analyze. Keep
	// successful proofs marked for subsequent literals in this clause.
	tagged := scratch.tagged[:0]
	var mask uint32
	for _, lit := range lits[1:] {
		tagged = append(tagged, lit>>1)
		mask |= 1 << uint(s.levels[lit>>1]&31)
	}
	scratch.tagged = tagged
	out := lits[:1]
	for _, lit := range lits[1:] {
		if s.reasons[lit>>1] < 0 || !s.reasonRemovable(lit>>1, mask) {
			out = append(out, lit)
		}
	}
	for _, v := range scratch.tagged {
		scratch.seen[v] = false
	}
	if uint32(len(out)) > s.opts.BinaryResolutionSize || s.clauseLBD(out) > s.opts.BinaryResolutionLBD {
		return out
	}
	stamp := s.nextStamp(len(s.values))
	for _, lit := range out {
		s.stamps[lit>>1] = stamp
	}
	list := s.watches[out[0]^1]
	for _, w := range list.entries[:list.binaries] {
		if s.literalValue(w.blocker) == 1 {
			s.stamps[w.blocker>>1] = 0
		}
	}
	end := len(out)
	for i := 1; i < end; {
		if s.stamps[out[i]>>1] != stamp {
			end--
			out[i], out[end] = out[end], out[i]
		} else {
			i++
		}
	}
	return out[:end]
}

// reasonRemovable proves redundancy by walking reasons to variables already
// in the clause or proved redundant. Discard tentative marks on failure.
func (s *Solver) reasonRemovable(v int, mask uint32) bool {
	scratch := &s.analysis
	top := len(scratch.tagged)
	pending := append(scratch.pending[:0], v)
	for len(pending) != 0 {
		v = pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		lits := s.clauses[s.reasons[v]].lits
		if len(lits) == 2 && s.literalValue(lits[0]) == 0 {
			lits[0], lits[1] = lits[1], lits[0]
		}
		for _, lit := range lits[1:] {
			u := lit >> 1
			if scratch.seen[u] || s.levels[u] == 0 {
				continue
			}
			if s.reasons[u] < 0 || mask&(1<<uint(s.levels[u]&31)) == 0 {
				for _, newlySeen := range scratch.tagged[top:] {
					scratch.seen[newlySeen] = false
				}
				scratch.tagged = scratch.tagged[:top]
				scratch.pending = pending
				return false
			}
			scratch.seen[u] = true
			scratch.tagged = append(scratch.tagged, u)
			pending = append(pending, u)
		}
	}
	scratch.pending = pending
	return true
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

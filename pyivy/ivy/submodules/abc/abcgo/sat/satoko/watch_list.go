// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

// propagate maintains two watched literals, inspecting only clauses watching
// the complement of each newly assigned literal.
func (s *Solver) propagate() int {
	var propagated int64
	finish := func() {
		s.stats.Propagations += propagated
		s.stats.AllPropagations += propagated
		s.simplifyProps -= propagated
	}
	for s.qhead < len(s.trail) {
		lit := s.trail[s.qhead]
		s.qhead++
		propagated++
		list := s.watches[lit]
		// Binary implications run before long clauses and do not reorder literals.
		for _, id := range list {
			c := s.clauses[id]
			if c.deleted || len(c.lits) != 2 {
				continue
			}
			other := c.lits[0]
			if other == lit^1 {
				other = c.lits[1]
			}
			if s.coneActive && !s.marks[other>>1] {
				continue
			}
			if s.literalValue(other) == 0 {
				return id
			}
			s.enqueue(other, id)
		}
		kept := list[:0]
		for pos, id := range list {
			c := s.clauses[id]
			if c.deleted {
				continue
			}
			if len(c.lits) == 2 {
				kept = append(kept, id)
				continue
			}
			s.stats.Inspections++
			blocker := s.watchBlockers[[2]int{lit, id}]
			if s.literalValue(blocker) == 1 {
				kept = append(kept, id)
				continue
			}
			falseLit := lit ^ 1
			if c.lits[0] == falseLit {
				c.lits[0], c.lits[1] = c.lits[1], c.lits[0]
			}
			first := c.lits[0]
			s.watchBlockers[[2]int{lit, id}] = first
			if s.coneActive && !s.marks[first>>1] {
				kept = append(kept, id)
				continue
			}
			if s.literalValue(first) == 1 {
				kept = append(kept, id)
				continue
			}
			replacement := -1
			for k := 2; k < len(c.lits); k++ {
				if s.literalValue(c.lits[k]) != 0 {
					replacement = k
					break
				}
			}
			if replacement >= 0 {
				c.lits[1], c.lits[replacement] = c.lits[replacement], falseLit
				other := c.lits[1] ^ 1
				s.watches[other] = append(s.watches[other], id)
				s.watchBlockers[[2]int{other, id}] = first
				continue
			}
			kept = append(kept, id)
			if s.literalValue(first) == 0 || !s.enqueue(first, id) {
				kept = append(kept, list[pos+1:]...)
				s.watches[lit] = kept
				s.qhead = len(s.trail)
				finish()
				return id
			}
		}
		s.watches[lit] = kept
	}
	finish()
	return -1
}

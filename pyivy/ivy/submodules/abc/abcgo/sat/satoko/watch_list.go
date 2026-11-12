// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

// propagate maintains two watched literals, inspecting only clauses watching
// the complement of each newly assigned literal.
func (s *Solver) propagate() int {
	for s.qhead < len(s.trail) {
		lit := s.trail[s.qhead]
		s.qhead++
		s.stats.Propagations++
		s.stats.AllPropagations++
		list := s.watches[lit]
		kept := list[:0]
		for pos, id := range list {
			c := s.clauses[id]
			if c.deleted {
				continue
			}
			s.stats.Inspections++
			falseLit := lit ^ 1
			if c.lits[0] == falseLit {
				c.lits[0], c.lits[1] = c.lits[1], c.lits[0]
			}
			first := c.lits[0]
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
				continue
			}
			kept = append(kept, id)
			if s.literalValue(first) == 0 || !s.enqueue(first, id) {
				kept = append(kept, list[pos+1:]...)
				s.watches[lit] = kept
				return id
			}
		}
		s.watches[lit] = kept
	}
	return -1
}

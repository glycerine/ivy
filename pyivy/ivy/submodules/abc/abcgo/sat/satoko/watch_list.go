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
		list := &s.watches[lit]
		for _, w := range list.entries[:list.binaries] {
			if s.coneActive && !s.marks[w.blocker>>1] {
				continue
			}
			if !s.enqueue(w.blocker, w.clause) {
				return w.clause
			}
		}
		kept := list.entries[:list.binaries]
		long := list.entries[list.binaries:]
		for pos, w := range long {
			c := s.clauses[w.clause]
			if c.deleted {
				continue
			}
			if s.coneActive && !s.marks[w.blocker>>1] || s.literalValue(w.blocker) == 1 {
				kept = append(kept, w)
				continue
			}
			falseLit := lit ^ 1
			if c.lits[0] == falseLit {
				c.lits[0], c.lits[1] = c.lits[1], c.lits[0]
			}
			first := c.lits[0]
			w.blocker = first
			if s.literalValue(first) == 1 {
				kept = append(kept, w)
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
				s.watches[other].entries = append(s.watches[other].entries, w)
				continue
			}
			kept = append(kept, w)
			if !s.enqueue(first, w.clause) {
				kept = append(kept, long[pos+1:]...)
				list.entries = kept
				s.stats.Inspections += int64(len(kept))
				s.qhead = len(s.trail)
				finish()
				return w.clause
			}
		}
		list.entries = kept
		s.stats.Inspections += int64(len(kept))
	}
	finish()
	return -1
}

// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

func (s *Solver) bumpClause(c *clause) {
	if !c.learnt {
		return
	}
	c.activity += s.clauseInc
	if c.activity&0x80000000 != 0 {
		for _, other := range s.clauses {
			if other.learnt {
				other.activity >>= 10
			}
		}
		s.clauseInc >>= 10
		if s.clauseInc < 1<<11 {
			s.clauseInc = 1 << 11
		}
	}
}
func (s *Solver) decayClauses() { s.clauseInc += s.clauseInc >> 10 }

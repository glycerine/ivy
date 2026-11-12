// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

// Clause IDs remain stable while Go's garbage collector owns the storage.
func (s *Solver) addStored(lits []int, learnt bool, lbd uint32) int {
	id := len(s.clauses)
	c := &clause{lits: append([]int(nil), lits...), learnt: learnt, lbd: lbd}
	s.clauses = append(s.clauses, c)
	if len(lits) > 1 {
		s.watches[lits[0]^1] = append(s.watches[lits[0]^1], id)
		s.watches[lits[1]^1] = append(s.watches[lits[1]^1], id)
	}
	if learnt {
		s.stats.LearntLiterals += int64(len(lits))
	} else {
		s.stats.OriginalLiterals += int64(len(lits))
	}
	return id
}
func (s *Solver) rebuildWatches() {
	s.watches = make([][]int, 2*len(s.values))
	for id, c := range s.clauses {
		if !c.deleted && len(c.lits) > 1 {
			s.watches[c.lits[0]^1] = append(s.watches[c.lits[0]^1], id)
			s.watches[c.lits[1]^1] = append(s.watches[c.lits[1]^1], id)
		}
	}
}

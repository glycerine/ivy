// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

// Clause IDs remain stable while Go's garbage collector owns the storage.
func (s *Solver) addStored(lits []int, learnt bool, lbd uint32) int {
	id := len(s.clauses)
	c := &clause{lits: append([]int(nil), lits...), learnt: learnt, lbd: lbd}
	s.clauses = append(s.clauses, c)
	if len(lits) > 1 {
		s.addWatch(lits[0]^1, id)
		s.addWatch(lits[1]^1, id)
	}
	if learnt && len(lits) > 1 {
		s.stats.LearntLiterals += int64(len(lits))
	} else if !learnt && len(lits) > 1 {
		s.stats.OriginalLiterals += int64(len(lits))
	}
	return id
}
func (s *Solver) rebuildWatches() {
	s.watches = make([][]int, 2*len(s.values))
	s.watchBlockers = nil
	for id, c := range s.clauses {
		if !c.deleted && len(c.lits) > 1 {
			s.addWatch(c.lits[0]^1, id)
			s.addWatch(c.lits[1]^1, id)
		}
	}
}

// Binary watches precede long watches, matching Satoko's propagation priority.
func (s *Solver) addWatch(lit, id int) {
	if s.watchBlockers == nil {
		s.watchBlockers = make(map[[2]int]int)
	}
	c := s.clauses[id]
	blocker := c.lits[0]
	if blocker == lit^1 {
		blocker = c.lits[1]
	}
	s.watchBlockers[[2]int{lit, id}] = blocker
	list := s.watches[lit]
	n := 0
	if len(s.clauses[id].lits) == 2 {
		for n < len(list) && len(s.clauses[list[n]].lits) == 2 {
			n++
		}
	} else {
		n = len(list)
	}
	list = append(list, id)
	list[n], list[len(list)-1] = list[len(list)-1], list[n]
	s.watches[lit] = list
}

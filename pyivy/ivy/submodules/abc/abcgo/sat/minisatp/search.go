package minisatp

// cancelUntil follows Solver::cancelUntil, reinserting unassigned variables
// into the original activity heap.
func (s *solver) cancelUntil(level int) {
	if len(s.limits) <= level {
		return
	}
	for i := len(s.trail) - 1; i >= s.limits[level]; i-- {
		v := int(s.trail[i] >> 1)
		s.assign[v] = 0
		s.reason[v] = nil
		if s.order.index[v] == 0 {
			s.order.insert(v)
		}
	}
	s.trail = s.trail[:s.limits[level]]
	s.limits = s.limits[:level]
	s.qhead = len(s.trail)
}

func (s *solver) learned(lits []int32, id ClauseID) {
	if len(lits) == 1 {
		s.unitID[lits[0]>>1] = id
		s.enqueue(lits[0], nil)
		return
	}
	q := append([]int32(nil), lits...)
	max := 1
	for i := 2; i < len(q); i++ {
		if s.level[q[i]>>1] > s.level[q[max]>>1] {
			max = i
		}
	}
	q[1], q[max] = q[max], q[1]
	c := &clause{lits: q, id: id, learnt: true, activity: s.claInc}
	if c.activity > 1e20 {
		for _, other := range s.learnts {
			other.activity *= 1e-20
		}
		c.activity *= 1e-20
		s.claInc *= 1e-20
	}
	s.enqueue(q[0], c)
	s.watches[q[0]^1] = append(s.watches[q[0]^1], c)
	s.watches[q[1]^1] = append(s.watches[q[1]^1], c)
	s.learnts = append(s.learnts, c)
}

func (s *solver) remove(c *clause) {
	for _, lit := range c.lits[:2] {
		w := s.watches[lit^1]
		for i, p := range w {
			if p == c {
				s.watches[lit^1] = append(w[:i], w[i+1:]...)
				break
			}
		}
	}
	s.proof.deleted(c.id)
}

func (s *solver) locked(c *clause) bool { return s.reason[c.lits[0]>>1] == c }

func sortLearntClauses(clauses []*clause, seed *float64) {
	less := func(x, y *clause) bool {
		return len(x.lits) > 2 && (len(y.lits) == 2 || x.activity < y.activity)
	}
	if len(clauses) <= 15 {
		for i := 0; i < len(clauses)-1; i++ {
			best := i
			for j := i + 1; j < len(clauses); j++ {
				if less(clauses[j], clauses[best]) {
					best = j
				}
			}
			clauses[i], clauses[best] = clauses[best], clauses[i]
		}
		return
	}
	*seed *= 1389796
	*seed -= float64(int(*seed/2147483647)) * 2147483647
	pivot := clauses[int(*seed/2147483647*float64(len(clauses)))]
	i, j := -1, len(clauses)
	for {
		for {
			i++
			if !less(clauses[i], pivot) {
				break
			}
		}
		for {
			j--
			if !less(pivot, clauses[j]) {
				break
			}
		}
		if i >= j {
			break
		}
		clauses[i], clauses[j] = clauses[j], clauses[i]
	}
	sortLearntClauses(clauses[:i], seed)
	sortLearntClauses(clauses[i:], seed)
}

func (s *solver) reduceDB() {
	extra := s.claInc / float64(len(s.learnts))
	seed := 91648253.0
	sortLearntClauses(s.learnts, &seed)
	kept := s.learnts[:0]
	for i, c := range s.learnts {
		if len(c.lits) > 2 && !s.locked(c) && (i < len(s.learnts)/2 || c.activity < extra) {
			s.remove(c)
		} else {
			kept = append(kept, c)
		}
	}
	s.learnts = kept
}

// simplifyDB performs the original root propagation and removes satisfied,
// unlocked clauses at the same assignment and propagation thresholds.
func (s *solver) simplifyDB() {
	if !s.ok || s.expired() {
		return
	}
	if s.propagate() != nil {
		s.ok = false
		return
	}
	if len(s.trail) == s.simpAssigns || s.simpProps > 0 {
		return
	}
	for _, lit := range s.trail[s.simpAssigns:] {
		s.watches[lit] = nil
		s.watches[lit^1] = nil
	}
	for _, db := range []*[]*clause{&s.clauses, &s.learnts} {
		kept := (*db)[:0]
		for _, c := range *db {
			if s.expired() {
				return
			}
			satisfied := false
			for _, lit := range c.lits {
				if s.value(lit) > 0 {
					satisfied = true
					break
				}
			}
			if !s.locked(c) && satisfied {
				s.remove(c)
			} else {
				kept = append(kept, c)
			}
		}
		*db = kept
	}
	s.simpAssigns = len(s.trail)
	s.simpProps = 0
	for _, db := range [][]*clause{s.clauses, s.learnts} {
		for _, c := range db {
			s.simpProps += int64(len(c.lits))
		}
	}
}

// search is the no-assumptions portion of Solver::search. Additional source
// controls and exported proof handling are still under development.
func (s *solver) search(limit, maxLearnts int) int {
	s.model = nil
	count := 0
	for {
		if s.expired() {
			return -1
		}
		if conflict := s.propagate(); conflict != nil {
			s.conflicts++
			count++
			if s.conflictLimit > 0 && s.conflicts >= s.conflictLimit {
				s.interrupted = true
				return -1
			}
			if len(s.limits) == 0 {
				return 0
			}
			learnt, level := s.analyze(conflict)
			if s.interrupted {
				return -1
			}
			id := s.proof.next - 1
			s.cancelUntil(level)
			s.learned(learnt, id)
			s.varInc /= 0.95
			s.claInc /= 0.999
		} else {
			if s.interrupted {
				return -1
			}
			if limit >= 0 && count >= limit {
				s.cancelUntil(0)
				return -1
			}
			if len(s.limits) == 0 {
				s.simplifyDB()
			}
			if maxLearnts >= 0 && len(s.learnts)-len(s.trail) >= maxLearnts {
				s.reduceDB()
			}
			v := s.order.selectVar()
			if v < 0 {
				s.model = append(s.model[:0], s.assign...)
				s.cancelUntil(0)
				return 1
			}
			s.limits = append(s.limits, len(s.trail))
			s.enqueue(int32(v*2+1), nil)
		}
	}
}

func (s *solver) solve() int {
	if s.expired() {
		return -1
	}
	if !s.ok {
		return 0
	}
	s.simplifyDB()
	if s.interrupted {
		return -1
	}
	if !s.ok {
		return 0
	}
	conflicts, learnts := 100.0, float64(len(s.clauses)/3)
	for {
		result := s.search(int(conflicts), int(learnts))
		if s.interrupted {
			return -1
		}
		if result >= 0 {
			return result
		}
		conflicts *= 1.5
		learnts *= 1.1
	}
}

// Copyright Satoko contributors. Distributed under the BSD 2-Clause License.
package satoko

import "sort"
import "github.com/glycerine/abcgo/sat/satoko/utils"

// analyze resolves the conflicting implication graph to its first UIP. The
// resulting asserting clause is added permanently, including across solves.
func (s *Solver) analyze(conflict int) ([]int, int, uint32) {
	seen := make([]bool, len(s.values))
	learnt := []int{0}
	path := 0
	var lastLevel []int
	p := -1
	index := len(s.trail) - 1
	for {
		c := s.clauses[conflict]
		s.bumpClause(c)
		if c.learnt && c.lbd > 2 {
			fresh := s.clauseLBD(c.lits)
			if fresh+1 < c.lbd {
				if c.lbd <= s.opts.FreezeLBD {
					c.protected = true
				}
				c.lbd = fresh
			}
		}
		for _, q := range c.lits {
			v := q >> 1
			if p >= 0 && v == p>>1 {
				continue
			}
			if seen[v] || s.levels[v] == 0 {
				continue
			}
			seen[v] = true
			s.bumpVariable(v)
			if s.levels[v] == s.decisionLevel() {
				path++
				if r := s.reasons[v]; r >= 0 && s.clauses[r].learnt {
					lastLevel = append(lastLevel, v)
				}
			} else {
				learnt = append(learnt, q)
			}
		}
		for index >= 0 && !seen[s.trail[index]>>1] {
			index--
		}
		if index < 0 {
			panic("invalid SAT implication graph")
		}
		p = s.trail[index]
		index--
		seen[p>>1] = false
		path--
		if path == 0 {
			break
		}
		conflict = s.reasons[p>>1]
		if conflict < 0 {
			panic("missing SAT implication reason")
		}
	}
	learnt[0] = p ^ 1
	learnt = s.minimizeLearnt(learnt)
	back := 0
	if len(learnt) > 1 {
		best := 1
		for i := 2; i < len(learnt); i++ {
			if s.levels[learnt[i]>>1] > s.levels[learnt[best]>>1] {
				best = i
			}
		}
		learnt[1], learnt[best] = learnt[best], learnt[1]
		back = s.levels[learnt[1]>>1]
	}
	levels := map[int]bool{}
	for _, lit := range learnt {
		levels[s.levels[lit>>1]] = true
	}
	for _, v := range lastLevel {
		if s.clauses[s.reasons[v]].lbd < uint32(len(levels)) {
			s.bumpVariable(v)
		}
	}
	return learnt, back, uint32(len(levels))
}
func (s *Solver) reduce() uint32 {
	locked := make(map[int]bool)
	for v, r := range s.reasons {
		if s.values[v] >= 0 && r >= 0 {
			locked[r] = true
		}
	}
	var ids []int
	for id, c := range s.clauses {
		if c.learnt && !c.deleted {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return 0
	}
	sort.Slice(ids, func(i, j int) bool { return clause_compare(s.clauses[ids[i]], s.clauses[ids[j]]) })
	var special uint32
	if s.clauses[ids[len(ids)/2]].lbd <= 3 {
		special += s.opts.SpecialReduceIncrement
	}
	if s.clauses[ids[len(ids)-1]].lbd <= 6 {
		special += s.opts.SpecialReduceIncrement
	}
	limit := int(float32(len(ids)) * s.opts.LearntRatio)
	deleted := 0
	for i, id := range ids {
		c := s.clauses[id]
		if !c.protected && c.lbd > 2 && len(c.lits) > 2 && !locked[id] && i < limit {
			c.deleted = true
			s.stats.LearntLiterals -= int64(len(c.lits))
			c.lits = nil
			deleted++
		} else {
			if c.protected {
				limit++
			}
			c.protected = false
		}
	}
	s.stats.Reductions++
	if float32(deleted) > float32(len(s.clauses))*s.opts.GarbageRatio {
		s.rebuildWatches()
	}
	return special
}
func (s *Solver) prepare() bool {
	s.trail = nil
	s.limits = nil
	s.qhead = 0
	s.order.Clear()
	for v := range s.values {
		s.values[v] = -1
		s.reasons[v] = -1
		s.levels[v] = 0
		s.order.Insert(uint32(v))
	}
	for id, c := range s.clauses {
		if c.deleted {
			continue
		}
		if len(c.lits) == 0 {
			return false
		}
		if len(c.lits) == 1 && !s.enqueue(c.lits[0], id) {
			return false
		}
	}
	return s.propagate() < 0
}
func (s *Solver) Solve() int {
	s.final = nil
	s.model = nil
	allConflicts, allPropagations := s.stats.AllConflicts, s.stats.AllPropagations
	s.stats = Stats{AllConflicts: allConflicts, AllPropagations: allPropagations}
	if !s.ok {
		return SATOKO_UNSAT
	}
	s.cancel(0)
	if s.propagate() >= 0 {
		s.ok = false
		return SATOKO_UNSAT
	}
	if !s.opts.NoSimplify {
		s.simplifyOrder()
	}
	restart := uint32(0)
	window := s.opts.RestartWindow
	if window == 0 {
		window = 50
	}
	recent := make([]uint32, 0, window)
	trailWindow := s.opts.TrailWindow
	if trailWindow == 0 {
		trailWindow = 5000
	}
	trailQueue := utils.NewBoundedQueue(trailWindow)
	var sum uint64
	var totalLBD uint64
	reduceAt := s.opts.FirstReduce
	if reduceAt == 0 {
		reduceAt = 2000
	}
	rc2 := reduceAt
	s.stats.Starts++
	for {
		if s.stopped() || s.opts.ConflictLimit > 0 && s.stats.Conflicts >= s.opts.ConflictLimit || s.opts.PropagationLimit > 0 && s.stats.Propagations >= s.opts.PropagationLimit {
			s.cancel(0)
			return SATOKO_UNDEC
		}
		conflict := s.propagate()
		if conflict < 0 && s.opts.PropagationLimit > 0 && s.stats.Propagations >= s.opts.PropagationLimit {
			s.cancel(0)
			return SATOKO_UNDEC
		}
		if conflict >= 0 {
			s.stats.Conflicts++
			s.stats.AllConflicts++
			restart++
			if s.decisionLevel() == 0 {
				s.ok = false
				return SATOKO_UNSAT
			}
			trailQueue.Push(uint32(len(s.trail)))
			if uint32(s.stats.Conflicts) > s.opts.FirstBlockRestart && uint32(len(recent)) == window && float64(len(s.trail)) > s.opts.BlockFactor*float64(trailQueue.Average()) {
				recent = recent[:0]
				sum = 0
			}
			learnt, back, lbd := s.analyze(conflict)
			s.cancel(back)
			id := s.addStored(learnt, true, lbd)
			if !s.enqueue(learnt[0], id) {
				panic("nonasserting learned clause")
			}
			s.decayVariables()
			s.decayClauses()
			if uint32(len(recent)) == window {
				sum -= uint64(recent[0])
				copy(recent, recent[1:])
				recent = recent[:len(recent)-1]
			}
			recent = append(recent, lbd)
			sum += uint64(lbd)
			totalLBD += uint64(lbd)
			if s.opts.LearntRatio > 0 && s.LearntCount() > 100 && uint32(s.stats.Conflicts) >= reduceAt {
				rc1 := uint32(s.stats.Conflicts)/rc2 + 1
				rc2 += s.reduce() + s.opts.ReduceIncrement
				reduceAt = rc1 * rc2
			}
			if uint32(len(recent)) == window && restart >= window && float64(sum)/float64(window)*s.opts.RestartFactor > float64(totalLBD)/float64(s.stats.Conflicts) {
				s.cancel(0)
				recent = recent[:0]
				sum = 0
				restart = 0
				s.stats.Starts++
			}
			continue
		}
		if !s.opts.NoSimplify && s.decisionLevel() == 0 {
			s.simplifyOrder()
		}
		decision := -1
		for s.decisionLevel() < len(s.assumptions) {
			lit := s.assumptions[s.decisionLevel()]
			value := s.literalValue(lit)
			if value == 0 {
				s.analyzeFinal(lit ^ 1)
				s.cancel(0)
				return SATOKO_UNSAT
			}
			s.limits = append(s.limits, len(s.trail))
			if value < 0 {
				decision = lit
				break
			}
		}
		if decision < 0 {
			for s.order.Len() > 0 {
				v := int(s.order.Pop())
				if s.values[v] < 0 && s.decision[v] && (!s.coneActive || s.marks[v]) {
					decision = v*2 + int(s.phases[v])
					break
				}
			}
			if decision < 0 {
				s.model = make([]int, len(s.values))
				for v, value := range s.values {
					s.model[v] = int(value)
				}
				s.cancel(0)
				return SATOKO_SAT
			}
			s.stats.Decisions++
			s.limits = append(s.limits, len(s.trail))
		}
		if !s.enqueue(decision, -1) {
			panic("conflicting decision")
		}
	}
}

// Root simplification rebuilds decision order only when its source budget expires.
func (s *Solver) simplifyOrder() {
	if s.simplifyAssigned == len(s.trail) || s.simplifyProps > 0 {
		return
	}
	var entries []uint32
	for v, value := range s.values {
		if value < 0 {
			entries = append(entries, uint32(v))
		}
	}
	s.order.Build(entries)
	s.simplifyAssigned = len(s.trail)

	for _, c := range s.clauses {
		if c.deleted || c.learnt || c.simplified || len(c.lits) < 2 {
			continue
		}
		for _, lit := range c.lits {
			if s.literalValue(lit) == 1 {
				c.simplified = true
				s.stats.OriginalLiterals -= int64(len(c.lits))
				break
			}
		}
	}
	s.simplifyProps = s.stats.OriginalLiterals + s.stats.LearntLiterals
}

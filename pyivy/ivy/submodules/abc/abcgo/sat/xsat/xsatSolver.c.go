// Go port of xsatSolver.c by Bruno Schmitt (2016).
package xsat

import (
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
	"time"
)

func xSAT_SolverRebuildOrderHeap(s *Solver) {
	vars := make([]int, 0, len(s.assigns))
	for v, value := range s.assigns {
		if value == VarX {
			vars = append(vars, v)
		}
	}
	xSAT_HeapBuild(s.order, vars)
}
func xSAT_SolverVarActRescale(s *Solver) {
	for i := range s.activity {
		s.activity[i] >>= 19
	}
	s.varInc >>= 19
	if s.varInc < 1<<5 {
		s.varInc = 1 << 5
	}
}
func xSAT_SolverVarActBump(s *Solver, v int) {
	s.activity[v] += s.varInc
	if s.activity[v]&0x80000000 != 0 {
		xSAT_SolverVarActRescale(s)
	}
	if xSAT_HeapInHeap(s.order, v) != 0 {
		xSAT_HeapDecrease(s.order, v)
	}
}
func xSAT_SolverVarActDecay(s *Solver) { s.varInc += s.varInc >> 4 }
func xSAT_SolverClaActRescale(s *Solver) {
	for _, ref := range s.learnts {
		s.memory.clauses[ref].activity >>= 14
	}
	s.claInc >>= 14
	if s.claInc < 1<<10 {
		s.claInc = 1 << 10
	}
}
func xSAT_SolverClaActBump(s *Solver, c *xSAT_Clause_t) {
	c.activity += s.claInc
	if c.activity&0x80000000 != 0 {
		xSAT_SolverClaActRescale(s)
	}
}
func xSAT_SolverClaActDecay(s *Solver) { s.claInc += s.claInc >> 10 }
func (s *Solver) lbd(lits []int) uint32 {
	levels := map[int]bool{}
	for _, lit := range lits {
		levels[s.levels[lit>>1]] = true
	}
	return uint32(len(levels))
}
func xSAT_SolverClaCalcLBD(s *Solver, c *xSAT_Clause_t) int { return int(s.lbd(c.lits)) }
func xSAT_SolverClaCalcLBD2(s *Solver, lits []int) int      { return int(s.lbd(lits)) }
func (s *Solver) newClause(lits []int, learnt bool) uint32 {
	if len(lits) < 2 {
		panic("stored clauses require at least two literals")
	}
	for _, lit := range lits {
		if lit < 0 || lit>>1 >= len(s.assigns) {
			panic("stored clause variable out of range")
		}
	}
	ref := xSAT_MemAppend(s.memory, len(lits)+2)
	c := s.memory.clauses[ref]
	c.lits = append(c.lits, lits...)
	c.learnt = learnt
	if learnt {
		c.lbd = s.lbd(lits)
		s.learnts = append(s.learnts, ref)
		s.Stats.LearntLiterals += int64(len(lits))
		xSAT_SolverClaActBump(s, c)
	} else {
		s.clauses = append(s.clauses, ref)
		s.Stats.ClauseLiterals += int64(len(lits))
	}
	watches := s.watches
	if len(lits) == 2 {
		watches = s.binary
	}
	xSAT_WatchListPush(watches.data[lits[0]^1], xSAT_Watcher_t{Blocker: lits[1], CRef: ref})
	xSAT_WatchListPush(watches.data[lits[1]^1], xSAT_Watcher_t{Blocker: lits[0], CRef: ref})
	return ref
}
func xSAT_SolverClaNew(s *Solver, lits *vec.Vec_Int_t, learnt int) uint32 {
	array := vec.Vec_IntArray(lits)
	native := make([]int, len(array))
	for i, value := range array {
		native[i] = int(value)
	}
	return s.newClause(native, learnt != 0)
}
func (s *Solver) enqueue(lit int, reason uint32) bool {
	value := s.litValue(lit)
	if value != LBoolUndef {
		return value == LBoolTrue
	}
	v := lit >> 1
	s.assigns[v] = byte(lit & 1)
	s.levels[v] = len(s.limits)
	s.reasons[v] = reason
	s.trail = append(s.trail, lit)
	return true
}
func xSAT_SolverEnqueue(s *Solver, lit int, reason uint32) int {
	return boolInt(s.enqueue(lit, reason))
}
func xSAT_SolverCancelUntil(s *Solver, level int) {
	if level < 0 {
		panic("negative decision level")
	}
	if len(s.limits) <= level {
		return
	}
	at := s.limits[level]
	for _, lit := range s.trail[at:] {
		v := lit >> 1
		s.polarity[v] = byte(lit & 1)
		s.assigns[v] = VarX
		s.reasons[v] = CRefUndef
		xSAT_HeapUpdate(s.order, v)
	}
	s.trail = s.trail[:at]
	s.limits = s.limits[:level]
	if s.qhead > at {
		s.qhead = at
	}
}
func (s *Solver) propagate() uint32 {
	for s.qhead < len(s.trail) {
		lit := s.trail[s.qhead]
		s.qhead++
		s.Stats.Propagations++
		s.simplifyRemaining--
		for _, w := range s.binary.data[lit].data {
			s.Stats.Inspections++
			if !s.enqueue(w.Blocker, w.CRef) {
				return w.CRef
			}
		}
		list := s.watches.data[lit].data
		kept := list[:0]
		for i, w := range list {
			s.Stats.Inspections++
			if s.litValue(w.Blocker) == LBoolTrue {
				kept = append(kept, w)
				continue
			}
			c := s.memory.clauses[w.CRef]
			falseLit := lit ^ 1
			if c.lits[0] == falseLit {
				c.lits[0], c.lits[1] = c.lits[1], c.lits[0]
			}
			first := c.lits[0]
			replacement := xSAT_Watcher_t{Blocker: first, CRef: w.CRef}
			if s.litValue(first) == LBoolTrue {
				kept = append(kept, replacement)
				continue
			}
			found := -1
			for k := 2; k < len(c.lits); k++ {
				if s.litValue(c.lits[k]) != LBoolFalse {
					found = k
					break
				}
			}
			if found >= 0 {
				c.lits[1], c.lits[found] = c.lits[found], falseLit
				xSAT_WatchListPush(s.watches.data[c.lits[1]^1], replacement)
				continue
			}
			kept = append(kept, replacement)
			if !s.enqueue(first, w.CRef) {
				kept = append(kept, list[i+1:]...)
				s.watches.data[lit].data = kept
				s.qhead = len(s.trail)
				return w.CRef
			}
		}
		s.watches.data[lit].data = kept
	}
	return CRefUndef
}
func xSAT_SolverPropagate(s *Solver) uint32 { return s.propagate() }
func (s *Solver) minimize(lits []int) []int {
	present := make([]bool, len(s.assigns))
	var mask uint32
	for _, lit := range lits {
		present[lit>>1] = true
	}
	for _, lit := range lits[1:] {
		mask |= 1 << uint(s.levels[lit>>1]&31)
	}
	var removable func(int, map[int]bool) bool
	removable = func(v int, visiting map[int]bool) bool {
		ref := s.reasons[v]
		if ref == CRefUndef || visiting[v] {
			return false
		}
		visiting[v] = true
		defer delete(visiting, v)
		for _, q := range s.memory.clauses[ref].lits {
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
	if len(out) > 30 || s.lbd(out) > 6 {
		return out
	}
	redundant := map[int]bool{}
	for _, w := range s.binary.data[out[0]^1].data {
		if s.litValue(w.Blocker) == LBoolTrue {
			redundant[w.Blocker^1] = true
		}
	}
	kept := out[:1]
	for _, lit := range out[1:] {
		if !redundant[lit] {
			kept = append(kept, lit)
		}
	}
	return kept
}
func (s *Solver) analyze(ref uint32) ([]int, int, uint32) {
	seen := make([]bool, len(s.assigns))
	learnt := []int{0}
	path := 0
	lit := LitUndef
	index := len(s.trail) - 1
	lastLevel := []int{}
	for {
		c := s.memory.clauses[ref]
		if c.learnt {
			xSAT_SolverClaActBump(s, c)
			if c.lbd > 2 {
				fresh := s.lbd(c.lits)
				if fresh+1 < c.lbd {
					if c.lbd <= s.Config.FrozenLBD {
						c.protected = true
					}
					c.lbd = fresh
				}
			}
		}
		for _, q := range c.lits {
			v := q >> 1
			if lit >= 0 && v == lit>>1 {
				continue
			}
			if seen[v] || s.levels[v] == 0 {
				continue
			}
			seen[v] = true
			xSAT_SolverVarActBump(s, v)
			if s.levels[v] == len(s.limits) {
				path++
				reason := s.reasons[v]
				if reason != CRefUndef && s.memory.clauses[reason].learnt {
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
			panic("invalid implication trail")
		}
		lit = s.trail[index]
		index--
		seen[lit>>1] = false
		path--
		if path == 0 {
			break
		}
		ref = s.reasons[lit>>1]
		if ref == CRefUndef {
			panic("missing implication reason")
		}
	}
	learnt[0] = lit ^ 1
	learnt = s.minimize(learnt)
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
	lbd := s.lbd(learnt)
	for _, v := range lastLevel {
		if s.memory.clauses[s.reasons[v]].lbd < lbd {
			xSAT_SolverVarActBump(s, v)
		}
	}
	return learnt, back, lbd
}
func (s *Solver) unwatch(ref uint32) {
	c := s.memory.clauses[ref]
	if len(c.lits) < 2 {
		return
	}
	watches := s.watches
	if len(c.lits) == 2 {
		watches = s.binary
	}
	xSAT_WatchListRemove(watches.data[c.lits[0]^1], ref)
	xSAT_WatchListRemove(watches.data[c.lits[1]^1], ref)
}
func xSAT_SolverReduceDB(s *Solver) {
	if len(s.learnts) == 0 {
		return
	}
	sort.Slice(s.learnts, func(i, j int) bool {
		return xSAT_ClauseCompare(s.memory.clauses[s.learnts[i]], s.memory.clauses[s.learnts[j]])
	})
	if s.memory.clauses[s.learnts[len(s.learnts)/2]].lbd <= 3 {
		s.rc2 += int64(s.Config.SpecialReduceIncrement)
	}
	if s.memory.clauses[s.learnts[len(s.learnts)-1]].lbd <= 5 {
		s.rc2 += int64(s.Config.SpecialReduceIncrement)
	}
	limit := len(s.learnts) / 2
	kept := s.learnts[:0]
	for i, ref := range s.learnts {
		c := s.memory.clauses[ref]
		locked := false
		for v, r := range s.reasons {
			if r == ref && s.assigns[v] != VarX {
				locked = true
				break
			}
		}
		if !c.protected && c.lbd > 2 && len(c.lits) > 2 && !locked && i < limit {
			c.marked = true
			s.Stats.LearntLiterals -= int64(len(c.lits))
			s.unwatch(ref)
		} else {
			if c.protected {
				limit++
			}
			c.protected = false
			kept = append(kept, ref)
		}
	}
	s.learnts = kept
	s.Stats.Reductions++
	if float64(xSAT_MemWastedCap(s.memory)) > float64(len(s.memory.clauses))*0.3 {
		xSAT_SolverGarbageCollect(s)
	}
}
func (s *Solver) overLimit() bool {
	return s.Config.ConflictLimit > 0 && s.Stats.Conflicts >= s.Config.ConflictLimit || s.Config.PropagationLimit > 0 && s.Stats.Propagations >= s.Config.PropagationLimit || !s.Config.Deadline.IsZero() && time.Now().After(s.Config.Deadline)
}
func xSAT_SolverSearch(s *Solver) int {
	s.Stats.Starts++
	for {
		if s.overLimit() {
			xSAT_SolverCancelUntil(s, 0)
			return LBoolUndef
		}
		conflict := s.propagate()
		if conflict != CRefUndef {
			s.Stats.Conflicts++
			if len(s.limits) == 0 {
				s.okay = false
				return LBoolFalse
			}
			xSAT_BQueuePush(s.trailQueue, uint32(len(s.trail)))
			if s.Stats.Conflicts > int64(s.Config.FirstBlockRestart) && xSAT_BQueueIsValid(s.lbdQueue) != 0 && float64(len(s.trail)) > s.Config.R*float64(xSAT_BQueueAvg(s.trailQueue)) {
				xSAT_BQueueClean(s.lbdQueue)
			}
			learnt, back, lbd := s.analyze(conflict)
			s.sumLBD += float64(lbd)
			xSAT_BQueuePush(s.lbdQueue, lbd)
			xSAT_SolverCancelUntil(s, back)
			ref := CRefUndef
			if len(learnt) > 1 {
				ref = s.newClause(learnt, true)
			}
			if !s.enqueue(learnt[0], ref) {
				panic("nonasserting learnt clause")
			}
			xSAT_SolverVarActDecay(s)
			xSAT_SolverClaActDecay(s)
			continue
		}
		if s.overLimit() {
			xSAT_SolverCancelUntil(s, 0)
			return LBoolUndef
		}
		if xSAT_BQueueIsValid(s.lbdQueue) != 0 && s.Stats.Conflicts > 0 && float64(xSAT_BQueueAvg(s.lbdQueue))*s.Config.K > s.sumLBD/float64(s.Stats.Conflicts) {
			xSAT_BQueueClean(s.lbdQueue)
			xSAT_SolverCancelUntil(s, 0)
			return LBoolUndef
		}
		if len(s.limits) == 0 && !s.Simplify() {
			return LBoolFalse
		}
		if s.Stats.Conflicts >= s.reduceAt {
			if s.rc2 <= 0 {
				s.rc2 = 2000
			}
			rc1 := s.Stats.Conflicts/s.rc2 + 1
			xSAT_SolverReduceDB(s)
			inc := s.Config.ReduceIncrement
			if inc <= 0 {
				inc = 300
			}
			s.rc2 += int64(inc)
			s.reduceAt = rc1 * s.rc2
		}
		variable := VarUndef
		for len(s.order.data) > 0 {
			v := xSAT_HeapRemoveMin(s.order)
			if s.assigns[v] == VarX {
				variable = v
				break
			}
		}
		if variable == VarUndef {
			return LBoolTrue
		}
		s.Stats.Decisions++
		s.limits = append(s.limits, len(s.trail))
		if !s.enqueue(2*variable+int(s.polarity[variable]), CRefUndef) {
			panic("conflicting decision")
		}
	}
}
func (s *Solver) Solve() int {
	s.model = nil
	if s.Config.Verbose {
		defer xSAT_SolverPrintStats(s)
	}
	if !s.okay {
		return LBoolFalse
	}
	defer xSAT_SolverCancelUntil(s, 0)
	for {
		status := xSAT_SolverSearch(s)
		if status == LBoolTrue {
			s.model = make([]int, len(s.assigns))
			for v, value := range s.assigns {
				if value == VarX {
					s.model[v] = -1
				} else {
					s.model[v] = 1 - int(value)
				}
			}
		}
		if status != LBoolUndef {
			return status
		}
		if s.overLimit() {
			return status
		}
	}
}
func xSAT_SolverClaRealloc(dest, src *xSAT_Mem_t, ref *uint32) {
	if *ref == CRefUndef {
		return
	}
	c := src.clauses[*ref]
	for i, existing := range dest.clauses {
		if existing == c {
			*ref = uint32(i)
			return
		}
	}
	*ref = uint32(len(dest.clauses))
	dest.clauses = append(dest.clauses, c)
}
func xSAT_SolverGarbageCollect(s *Solver) {
	dest := xSAT_MemAlloc(len(s.clauses) + len(s.learnts))
	mapping := map[uint32]uint32{}
	copyRef := func(ref uint32) uint32 {
		if mapped, ok := mapping[ref]; ok {
			return mapped
		}
		mapped := uint32(len(dest.clauses))
		dest.clauses = append(dest.clauses, s.memory.clauses[ref])
		mapping[ref] = mapped
		return mapped
	}
	for _, list := range [][]uint32{s.clauses, s.learnts} {
		for i, ref := range list {
			list[i] = copyRef(ref)
		}
	}
	for v, ref := range s.reasons {
		if ref != CRefUndef {
			s.reasons[v] = copyRef(ref)
		}
	}
	for _, vec := range []*xSAT_VecWatchList_t{s.watches, s.binary} {
		for _, list := range vec.data {
			for i, w := range list.data {
				list.data[i].CRef = copyRef(w.CRef)
			}
		}
	}
	s.memory = dest
}

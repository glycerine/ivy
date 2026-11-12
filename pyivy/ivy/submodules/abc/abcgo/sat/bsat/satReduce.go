package bsat

import (
	"fmt"
	"slices"
	"sync/atomic"
	"time"
)

var satReductionNanoseconds atomic.Int64

func sat_solver_simplify(s *sat_solver) int32 {
	if sat_solver_dl(s) != 0 {
		panic("simplify above root level")
	}
	if sat_solver_propagate(s) != 0 {
		return 0
	}
	return 1
}

// sat_solver_reducedb preserves locked/binary/recent/high-scoring learned
// clauses, then remaps reasons and watches before moving compacted clauses.
func sat_solver_reducedb(s *sat_solver) {
	started := time.Now()
	m := &s.Mem
	old := s.act_clas.size
	if s.nLearntMax <= 0 || old != m.nEntries[1] || uint32(old) != s.stats.learnts {
		panic("inconsistent learned database")
	}
	s.nDBreduces++
	s.nLearntMax = s.nLearntStart + s.nLearntDelta*s.nDBreduces
	if old == 0 {
		return
	}
	handles := m.handles(1)
	scores := make([]int32, old)
	for _, h := range handles {
		c := Sat_MemClauseHand(m, h)
		id := clause_id(c)
		score := (7 - min(c.lbd, 7)) << 28
		if s.ClaActType == 0 {
			score |= s.act_clas.ptr[id] >> 4
		}
		scores[id] = score
	}
	ordered := append([]int32(nil), scores...)
	slices.Sort(ordered)
	selected := old * s.nLearntRatio / 100
	selected = min(max(selected, 0), old)
	cutoff := int32(0x7fffffff)
	if selected > 0 {
		cutoff = ordered[old-selected]
	}
	recentStart := old - s.nLearntMax/20
	write := int32(0)
	for index, h := range handles {
		c := Sat_MemClauseHand(m, h)
		id := clause_id(c)
		if c.mark != 0 {
			panic("premarked learned clause")
		}
		if int32(index) > recentStart || c.size < 3 || scores[id] > cutoff || s.reasons[lit_var(c.lits[0])] == h {
			s.act_clas.ptr[write] = s.act_clas.ptr[id]
			write++
		} else {
			c.mark = 1
			s.stats.learnts--
			s.stats.learnts_literals -= int64(c.size)
		}
	}
	veci_resize(&s.act_clas, write)
	if Sat_MemCompactLearned(m, 0) != write {
		panic("learned compaction count mismatch")
	}
	for v := int32(0); v < s.size; v++ {
		h := s.reasons[v]
		if h == 0 || clause_is_lit(h) != 0 || clause_learnt_h(m, h) == 0 {
			continue
		}
		c := Sat_MemClauseHand(m, h)
		if c.mark != 0 {
			panic("removed locked clause")
		}
		s.reasons[v] = clause_id(c)
	}
	for i := range s.wlists {
		w := &s.wlists[i]
		kept := int32(0)
		for _, h := range veci_begin(w) {
			if clause_is_lit(h) != 0 || clause_learnt_h(m, h) == 0 {
				w.ptr[kept] = h
				kept++
				continue
			}
			c := Sat_MemClauseHand(m, h)
			if c.mark == 0 {
				w.ptr[kept] = clause_id(c)
				kept++
			}
		}
		veci_resize(w, kept)
	}
	if Sat_MemCompactLearned(m, 1) != write {
		panic("learned move count mismatch")
	}
	elapsed := satReductionNanoseconds.Add(time.Since(started).Nanoseconds())
	if s.fVerbose != 0 {
		fmt.Printf("reduceDB: Keeping %7d out of %7d clauses (%5.2f %%)  Time: %.2f sec\n", s.stats.learnts, old, 100*float64(s.stats.learnts)/float64(old), float64(elapsed)/1e9)
	}
}

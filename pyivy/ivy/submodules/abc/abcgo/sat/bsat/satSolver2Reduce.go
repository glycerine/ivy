package bsat

import "slices"

func sat_solver2_reducedb(s *sat_solver2) {
	old := s.act_clas.size
	if s.nLearntMax <= 0 {
		panic("disabled solver2 reduction")
	}
	s.nDBreduces++
	s.nLearntMax = s.nLearntStart + s.nLearntDelta*s.nDBreduces
	if old == 0 {
		return
	}
	handles := s.Mem.handles(1)
	scores := make([]int32, old)
	for _, h := range handles {
		c := clause2_read(s, h)
		id := clause_id(c)
		scores[id] = (7-min(c.lbd, 7))<<28 | int32(uint32(s.act_clas.ptr[id])>>4)
	}
	sorted := append([]int32(nil), scores...)
	slices.Sort(sorted)
	selected := min(max(old*s.nLearntRatio/100, 0), old)
	cutoff := int32(0x7fffffff)
	if selected > 0 {
		cutoff = sorted[old-selected]
	}
	recent := old - s.nLearntMax/20
	keep := make([]bool, len(handles))
	removed := int32(0)
	for i, h := range handles {
		c := clause2_read(s, h)
		keep[i] = int32(i) > recent || c.size < 2 || scores[clause_id(c)] >= cutoff || s.reasons[lit_var(c.lits[0])] == h
		if !keep[i] {
			removed++
		}
	}
	if removed < old/6 {
		return
	}
	write := int32(0)
	for i, h := range handles {
		c := clause2_read(s, h)
		id := clause_id(c)
		if !keep[i] {
			c.mark = 1
			s.stats.learnts--
			s.stats.learnts_literals -= int64(c.size)
			continue
		}
		s.act_clas.ptr[write] = s.act_clas.ptr[id]
		if s.claProofs.size != 0 {
			s.claProofs.ptr[write] = s.claProofs.ptr[id]
		}
		if s.pPrf2 != nil {
			Prf_ManAddSaved(s.pPrf2, id, write)
		}
		write++
	}
	if s.pPrf2 != nil {
		Prf_ManCompact(s.pPrf2, write)
	}
	veci_resize(&s.act_clas, write)
	if s.claProofs.size != 0 {
		veci_resize(&s.claProofs, write)
	}
	Sat_MemCompactLearned(&s.Mem, 0)
	for v := int32(0); v < s.size; v++ {
		if h := s.reasons[v]; h != 0 && clause_learnt_h(&s.Mem, h) != 0 {
			c := clause2_read(s, h)
			if c.mark != 0 {
				panic("removed locked solver2 reason")
			}
			s.reasons[v] = clause_id(c)
		}
		if h := s.units[v]; h != 0 && clause_learnt_h(&s.Mem, h) != 0 {
			c := clause2_read(s, h)
			if c.mark != 0 {
				panic("removed solver2 unit proof")
			}
			s.units[v] = clause_id(c)
		}
	}
	for i := range s.wlists {
		w := &s.wlists[i]
		n := int32(0)
		for _, h := range veci_begin(w) {
			if clause_learnt_h(&s.Mem, h) == 0 {
				w.ptr[n] = h
				n++
				continue
			}
			c := clause2_read(s, h)
			if c.mark == 0 {
				w.ptr[n] = clause_id(c)
				n++
			}
		}
		veci_resize(w, n)
	}
	Sat_MemCompactLearned(&s.Mem, 1)
	if s.pPrf1 != nil {
		roots := append([]int32(nil), veci_begin(&s.claProofs)...)
		finalIndex := -1
		if s.hProofLast > 0 {
			finalIndex = len(roots)
			roots = append(roots, s.hProofLast)
		}
		s.hProofPivot = Sat_ProofReduce(s.pPrf1, roots, s.hProofPivot)
		copy(s.claProofs.ptr[:s.claProofs.size], roots)
		if finalIndex >= 0 {
			s.hProofLast = roots[finalIndex]
		}
	}
}

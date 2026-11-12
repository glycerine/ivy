package bsat

// satClauseLBD preserves ABC's 32-bit decision-level signature estimate.
func satClauseLBD(a *satAssignments, c *clause) int32 {
	var mask uint32
	var count int32
	for _, l := range clause_begin(c) {
		bit := uint32(1) << uint(a.levels[lit_var(l)]&31)
		if mask&bit == 0 {
			mask |= bit
			count++
		}
	}
	return count
}

// satPropagate implements ABC's binary and two-watched-literal propagation.
// Watch entries are either tagged literals or stable clause handles. Compaction
// retains unprocessed watches on conflict and moves replacements to their list.
func satPropagate(a *satAssignments, mem *Sat_Mem_t, watches []veci, binaryHandle cla, binary *clause, frequencies []uint8, stats *stats_t) cla {
	for a.qhead < a.qtail {
		p := a.trail[a.qhead]
		a.qhead++
		ws := &watches[p]
		read := veci_begin(ws)
		write := 0
		stats.propagations++
		var conflict cla
		for i := 0; i < len(read); i++ {
			h := read[i]
			if clause_is_lit(h) != 0 {
				implied := clause_read_lit(h)
				ws.ptr[write] = h
				write++
				if a.assigns[lit_var(implied)] == lit_sign(implied) {
					continue
				}
				if !a.enqueueTracked(implied, clause_from_lit(p), frequencies) {
					conflict = binaryHandle
					binary.lits[0] = implied
					binary.lits[1] = lit_neg(p)
				}
			} else {
				c := Sat_MemClauseHand(mem, h)
				lits := clause_begin(c)
				if len(lits) < 2 {
					panic("watched unit clause")
				}
				falseLit := lit_neg(p)
				if lits[0] == falseLit {
					lits[0], lits[1] = lits[1], lits[0]
				}
				if lits[1] != falseLit {
					panic("watch does not match clause")
				}
				if a.assigns[lit_var(lits[0])] == lit_sign(lits[0]) {
					ws.ptr[write] = h
					write++
					continue
				}
				moved := false
				for k := 2; k < len(lits); k++ {
					candidate := lits[k]
					if a.assigns[lit_var(candidate)] != 1-lit_sign(candidate) {
						lits[1], lits[k] = candidate, falseLit
						veci_push(&watches[lit_neg(candidate)], h)
						moved = true
						break
					}
				}
				if moved {
					continue
				}
				ws.ptr[write] = h
				write++
				if c.lrn != 0 {
					c.lbd = satClauseLBD(a, c)
				}
				if !a.enqueueTracked(lits[0], h, frequencies) {
					conflict = h
				}
			}
			if conflict != 0 {
				write += copy(ws.ptr[write:], read[i+1:])
				break
			}
		}
		stats.inspects += int64(write)
		veci_resize(ws, int32(write))
		if conflict != 0 {
			return conflict
		}
	}
	return 0
}

func (a *satAssignments) enqueueTracked(l lit, reason cla, frequencies []uint8) bool {
	if frequencies != nil {
		frequencies[lit_var(l)] = 1
	}
	return a.enqueue(l, reason)
}

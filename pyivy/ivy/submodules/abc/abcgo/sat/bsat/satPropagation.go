package bsat

import "math/bits"

// satClauseLBD preserves ABC's 32-bit decision-level signature estimate.
func satClauseLBD(a *satAssignments, c *clause) uint8 {
	var mask uint32
	for _, l := range c.lits {
		mask |= uint32(1) << uint(a.levels[lit_var(l)]&31)
	}
	return uint8(bits.OnesCount32(mask))
}

// satPropagate implements ABC's binary and two-watched-literal propagation.
// Watch entries are either tagged literals or stable clause handles. Compaction
// retains unprocessed watches on conflict and moves replacements to their list.
func satPropagate(a *satAssignments, mem *Sat_Mem_t, watches []veci, binaryHandle cla, binary *clause, frequencies []uint8, stats *stats_t) cla {
	assigns := a.assigns
	for a.qhead < a.qtail {
		p := a.trail[a.qhead]
		a.qhead++
		ws := &watches[p]
		read := veci_begin(ws)
		write := 0
		stats.propagations++
		var conflict cla
		// A replacement literal cannot be the false literal for this list,
		// so pushing its watch cannot reallocate the slice being compacted.
	nextWatch:
		for i, h := range read {
			if clause_is_lit(h) != 0 {
				implied := clause_read_lit(h)
				read[write] = h
				write++
				if assigns[lit_var(implied)] == uint8(lit_sign(implied)) {
					continue
				}
				if !a.enqueueTracked(implied, clause_from_lit(p), frequencies) {
					conflict = binaryHandle
					binary.lits[0] = implied
					binary.lits[1] = lit_neg(p)
				}
			} else {
				c := mem.clauseAt(h)
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
				if assigns[lit_var(lits[0])] == uint8(lit_sign(lits[0])) {
					read[write] = h
					write++
					continue
				}
				for k := 2; k < len(lits); k++ {
					candidate := lits[k]
					if assigns[lit_var(candidate)] != uint8(1-lit_sign(candidate)) {
						lits[1], lits[k] = candidate, falseLit
						veci_push(&watches[lit_neg(candidate)], h)
						continue nextWatch
					}
				}
				read[write] = h
				write++
				if c.lrn != 0 {
					c.lbd = satClauseLBD(a, c)
				}
				if !a.enqueueTracked(lits[0], h, frequencies) {
					conflict = h
				}
			}
			if conflict != 0 {
				write += copy(read[write:], read[i+1:])
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
	if frequencies != nil && frequencies[lit_var(l)] == 0 {
		frequencies[lit_var(l)] = 1
	}
	return a.enqueue(l, reason)
}

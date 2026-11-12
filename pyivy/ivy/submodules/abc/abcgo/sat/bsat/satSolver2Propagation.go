package bsat

// Solver2 retains handles for binary clauses and materializes proof-producing
// level-zero unit clauses. This is distinct from the primary binary-tag path.
func solver2_propagate(s *sat_solver2) *clause {
	var conflict *clause
	for conflict == nil && s.qhead < s.qtail {
		literal := s.trail[s.qhead]
		s.qhead++
		watches := &s.wlists[literal]
		old := veci_begin(watches)
		kept := 0
		s.stats.propagations++
		for index := 0; index < len(old); index++ {
			handle := old[index]
			c := clause2_read(s, handle)
			falseLiteral := lit_neg(literal)
			if c.lits[0] == falseLiteral {
				c.lits[0], c.lits[1] = c.lits[1], c.lits[0]
			}
			if c.lits[1] != falseLiteral {
				panic("solver2 watch invariant")
			}
			if s.assigns[lit_var(c.lits[0])] == lit_sign(c.lits[0]) {
				old[kept] = handle
				kept++
				continue
			}
			moved := false
			for k := int32(2); k < c.size; k++ {
				l := c.lits[k]
				if s.assigns[lit_var(l)] != 1-lit_sign(l) {
					c.lits[1], c.lits[k] = l, falseLiteral
					veci_push(&s.wlists[lit_neg(l)], handle)
					moved = true
					break
				}
			}
			if moved {
				continue
			}
			implied := c.lits[0]
			if s.fProofLogging != 0 && solver2_dlevel(s) == 0 {
				variable := lit_var(implied)
				proof_chain_start(s, c)
				for _, l := range c.lits[1:c.size] {
					proof_chain_resolve(s, nil, lit_var(l))
				}
				proof := proof_chain_stop(s)
				unit := clause2_create_new(s, []lit{implied}, 1, proof)
				if s.units[variable] == 0 {
					s.units[variable] = unit
				} else {
					if s.assigns[variable] != 1-lit_sign(implied) {
						panic("duplicate same-polarity solver2 unit proof")
					}
					proof_chain_start(s, clause2_read(s, unit))
					proof_chain_resolve(s, nil, variable)
					s.hProofLast = proof_chain_stop(s)
				}
			}
			old[kept] = handle
			kept++
			if c.lrn != 0 {
				c.lbd = satClauseLBD(&s.satAssignments, c)
			}
			if solver2_enqueue(s, implied, handle) == 0 {
				conflict = c
				// Preserve every remaining watch; original pointer loop's index update was
				// incidental, while the retained watcher set is the semantic requirement.
				kept += copy(old[kept:], old[index+1:])
				break
			}
		}
		s.stats.inspects += int64(kept)
		veci_resize(watches, int32(kept))
	}
	return conflict
}

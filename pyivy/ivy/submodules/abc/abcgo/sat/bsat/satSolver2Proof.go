package bsat

func proof_chain_start(s *sat_solver2, c *clause) {
	if s.fProofLogging == 0 {
		return
	}
	if s.pInt2 != nil {
		s.tempInter = s.pInt2.chainStart(c)
	}
	if s.pPrf2 != nil {
		Prf_ManChainStart(s.pPrf2, c)
	}
	if s.pPrf1 != nil {
		id := clause2_proofid(s, c, 0)
		if id>>2 <= 0 {
			panic("invalid proof antecedent")
		}
		veci_resize(&s.temp_proof, 0)
		veci_push(&s.temp_proof, id)
	}
}
func proof_chain_resolve(s *sat_solver2, c *clause, variable int32) {
	if s.fProofLogging == 0 {
		return
	}
	if c == nil {
		c = clause2_read(s, s.units[variable])
	}
	if c == nil {
		panic("missing unit proof reason")
	}
	if s.pInt2 != nil {
		s.tempInter = s.pInt2.chainResolve(c, s.tempInter, s.variableA[variable])
	}
	if s.pPrf2 != nil {
		Prf_ManChainResolve(s.pPrf2, c)
	}
	if s.pPrf1 != nil {
		id := clause2_proofid(s, c, s.variableA[variable])
		if id>>2 <= 0 {
			panic("invalid proof antecedent")
		}
		veci_push(&s.temp_proof, id)
	}
}
func proof_chain_stop(s *sat_solver2) int32 {
	if s.fProofLogging == 0 {
		return 0
	}
	if s.pInt2 != nil {
		id := s.tempInter
		s.tempInter = -1
		return id
	}
	if s.pPrf2 != nil {
		Prf_ManChainStop(s.pPrf2)
	}
	if s.pPrf1 != nil {
		return s.pPrf1.Append(veci_begin(&s.temp_proof))
	}
	return 0
}
func clause2_create_new(s *sat_solver2, literals []lit, learnt, proof int32) cla {
	for _, l := range literals {
		if lit_check(l, s.size) == 0 {
			panic("solver2 clause literal outside solver")
		}
	}
	h := Sat_MemAppend(&s.Mem, literals, int32(len(literals)), learnt, 1)
	c := clause2_read(s, h)
	if learnt != 0 {
		if s.pPrf1 != nil && proof == 0 {
			panic("missing learned proof")
		}
		c.lbd = satClauseLBD(&s.satAssignments, c)
		if clause_id(c) != s.act_clas.size {
			panic("solver2 learned activity mismatch")
		}
		if s.pPrf1 != nil || s.pInt2 != nil {
			veci_push(&s.claProofs, proof)
		}
		veci_push(&s.act_clas, 0)
		if len(literals) > 2 {
			act_clause2_bump(s, c)
		}
		s.stats.learnts++
		s.stats.learnts_literals += int64(len(literals))
		s.hLearntLast = h
	} else {
		if clause_id(c) != int32(s.stats.clauses) {
			panic("solver2 problem clause ID mismatch")
		}
		s.stats.clauses++
		s.stats.clauses_literals += int64(len(literals))
	}
	if len(literals) > 1 {
		veci_push(&s.wlists[lit_neg(literals[0])], h)
		veci_push(&s.wlists[lit_neg(literals[1])], h)
	}
	return h
}

// Solver2 uses unsigned 32-bit integer activity, including its wrap and rescale
// thresholds; the common heap stores widened values without changing ordering.
func act_var2_bump(s *sat_solver2, v int32) {
	s.activity[v] = uint64(uint32(s.activity[v]) + uint32(s.var_inc))
	if s.activity[v]&0x80000000 != 0 {
		for i := range s.activity {
			s.activity[i] >>= 19
		}
		s.var_inc = max(s.var_inc>>19, 1<<4)
	}
	if s.orderpos[v] >= 0 {
		order_update(&s.sat_solver_t, v)
	}
}
func act_var2_decay(s *sat_solver2) { s.var_inc = uint64(uint32(s.var_inc) + (uint32(s.var_inc) >> 4)) }
func act_clause2_bump(s *sat_solver2, c *clause) {
	id := clause_id(c)
	value := uint32(s.act_clas.ptr[id]) + s.cla_inc
	s.act_clas.ptr[id] = int32(value)
	if value&0x80000000 != 0 {
		for i := range s.act_clas.ptr[:s.act_clas.size] {
			s.act_clas.ptr[i] = int32(uint32(s.act_clas.ptr[i]) >> 14)
		}
		s.cla_inc = max(s.cla_inc>>14, 1<<10)
	}
}
func act_clause2_decay(s *sat_solver2) { s.cla_inc += s.cla_inc >> 10 }

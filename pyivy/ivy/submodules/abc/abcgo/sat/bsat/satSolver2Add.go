package bsat

import "slices"

func sat_solver2_addclause(s *sat_solver2, input []lit, id int32) cla {
	if solver2_dlevel(s) != 0 || len(input) == 0 || id == 0 {
		panic("invalid solver2 problem clause")
	}
	literals := append([]lit(nil), input...)
	slices.Sort(literals)
	if literals[0] < 0 {
		panic("negative encoded literal")
	}
	sat_solver2_setnvars(s, lit_var(literals[len(literals)-1])+1)
	unique := 0
	tautology := false
	for _, l := range literals {
		if unique > 0 && literals[unique-1] == l {
			continue
		}
		if unique > 0 && literals[unique-1] == lit_neg(l) {
			tautology = true
		}
		literals[unique] = l
		unique++
	}
	literals = literals[:unique]
	satisfied := tautology
	free := -1
	falseCount := 0
	for i, l := range literals {
		value := s.assigns[lit_var(l)]
		if value == lit_sign(l) {
			satisfied = true
		}
		if value == satUnassigned {
			free = i
		} else {
			falseCount++
		}
	}
	if !satisfied && free < 0 {
		panic("solver2 input clause already contradicted")
	}
	if !satisfied {
		literals[0], literals[free] = literals[free], literals[0]
	}
	handle := clause2_create_new(s, literals, 0, 0)
	clause2_set_id(s, handle, id)
	if !satisfied && falseCount+1 == len(literals) && s.fProofLogging != 0 {
		if falseCount == 0 {
			s.units[lit_var(literals[0])] = handle
			solver2_enqueue(s, literals[0], 0)
		} else {
			proof_chain_start(s, clause2_read(s, handle))
			for _, l := range literals[1:] {
				proof_chain_resolve(s, nil, lit_var(l))
			}
			proof := proof_chain_stop(s)
			unit := clause2_create_new(s, literals[:1], 1, proof)
			s.units[lit_var(literals[0])] = unit
			if solver2_enqueue(s, literals[0], handle) == 0 {
				panic("inconsistent solver2 unit")
			}
		}
	}
	return handle
}

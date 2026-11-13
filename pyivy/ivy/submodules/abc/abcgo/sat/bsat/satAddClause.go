package bsat

import (
	"fmt"
	"slices"
)

func sat_solver_addclause(s *sat_solver, input []lit) int32 {
	if s.fPrintClause != 0 {
		for _, l := range input {
			sign := ""
			if lit_sign(l) != 0 {
				sign = "!"
			}
			fmt.Printf("%s%d ", sign, lit_var(l))
		}
		fmt.Println()
	}
	data := append([]lit(nil), input...)
	slices.Sort(data)
	if len(data) > 0 {
		if data[0] < 0 {
			panic("negative encoded literal")
		}
		sat_solver_setnvars(s, lit_var(data[len(data)-1])+1)
	}
	// Proof stores require normalized clauses. Deduplicate before logging and
	// omit tautologies, which impose no constraint on the formula.
	data = slices.Compact(data)
	for i := 1; i < len(data); i++ {
		if data[i] == lit_neg(data[i-1]) {
			return 1
		}
	}
	if p := solverClauseStore(s); p != nil {
		if Sto_ManAddClause(p, data) == 0 {
			panic("invalid clause for proof store")
		}
	}
	write := 0
	for _, l := range data {
		if s.assigns[lit_var(l)] == lit_sign(l) {
			return 1
		}
		if s.assigns[lit_var(l)] == satUnassigned {
			data[write] = l
			write++
		}
	}
	data = data[:write]
	if len(data) == 0 {
		return 0
	}
	if len(data) == 1 {
		return sat_solver_enqueue(s, data[0], 0)
	}
	sat_solver_clause_new(s, data, 0)
	return 1
}
func sat_solver_record(s *sat_solver, learnt *veci) {
	literals := veci_begin(learnt)
	if len(literals) == 0 {
		panic("cannot record empty learned clause")
	}
	var h cla
	if len(literals) > 1 {
		h = sat_solver_clause_new(s, literals, 1)
	}
	if sat_solver_enqueue(s, literals[0], h) == 0 {
		panic("learned clause not asserting")
	}
	if h == 0 {
		veci_push(&s.unit_lits, literals[0])
	}
	if p := solverClauseStore(s); p != nil {
		if Sto_ManAddClause(p, literals) == 0 {
			panic("invalid recorded proof clause")
		}
	}
}

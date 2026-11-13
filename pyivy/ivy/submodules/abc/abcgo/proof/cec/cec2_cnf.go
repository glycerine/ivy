package cec

import "github.com/glycerine/abcgo/sat/satoko"

// cec2AddMuxClauses encodes the same four ITE clauses and two strengthening
// clauses as Cec2_AddClausesMux. Input and output values are SAT literals.
func cec2AddMuxClauses(s *satoko.Solver, out, control, then, otherwise int) bool {
	clauses := [][]int{
		{control ^ 1, then ^ 1, out},
		{control ^ 1, then, out ^ 1},
		{control, otherwise ^ 1, out},
		{control, otherwise, out ^ 1},
	}
	if then>>1 != otherwise>>1 {
		clauses = append(clauses, []int{then, otherwise, out ^ 1}, []int{then ^ 1, otherwise ^ 1, out})
	}
	for _, clause := range clauses {
		if !s.AddClause(clause) {
			return false
		}
	}
	return true
}

// cec2AddSuperClauses encodes out = AND(fanins) using the source's one
// implication per input and its single reverse implication.
func cec2AddSuperClauses(s *satoko.Solver, out int, fanins []int) bool {
	reverse := make([]int, len(fanins)+1)
	for i, fanin := range fanins {
		if !s.AddClause([]int{fanin, out ^ 1}) {
			return false
		}
		reverse[i] = fanin ^ 1
	}
	reverse[len(fanins)] = out
	return s.AddClause(reverse)
}

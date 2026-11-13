package abs

import "github.com/glycerine/abcgo/sat/bsat"

// These gate clauses match satSolver2's literal encoding and preserve the
// VTA object ID on every clause in a gate's proof root.
func vtaLit(variable, complement int32) bsat.Literal {
	return bsat.Literal(2*variable + complement)
}
func vtaSolverValue(s *bsat.LabeledSolver2, variable int32) bool {
	value, ok := s.Value(variable)
	return ok && value
}
func vtaSolverAddConst(s *bsat.LabeledSolver2, v, complement, mark, id int32) {
	s.AddClause(id, vtaLit(v, complement))
}
func vtaSolverAddBuffer(s *bsat.LabeledSolver2, a, b, complement, mark, id int32) {
	s.AddClause(id, vtaLit(a, 0), vtaLit(b, 1-complement))
	s.AddClause(id, vtaLit(a, 1), vtaLit(b, complement))
}
func vtaSolverAddAnd(s *bsat.LabeledSolver2, v, a, b, ca, cb, mark, id int32) {
	s.AddClause(id, vtaLit(v, 1), vtaLit(a, ca))
	s.AddClause(id, vtaLit(v, 1), vtaLit(b, cb))
	s.AddClause(id, vtaLit(v, 0), vtaLit(a, 1-ca), vtaLit(b, 1-cb))
}
func vtaSolverAddConstraint(s *bsat.LabeledSolver2, v, w, complement, mark, id int32) {
	s.AddClause(id, vtaLit(v, complement), vtaLit(w, 0))
}

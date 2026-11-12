package bsat

func solver2_addGate(s *sat_solver2, clauses [][]lit, mark, id int32) int32 {
	for _, literals := range clauses {
		handle := sat_solver2_addclause(s, literals, id)
		if mark != 0 {
			clause2_set_partA(s, handle, 1)
		}
	}
	return int32(len(clauses))
}
func sat_solver2_add_const(s *sat_solver2, v, complement, mark, id int32) int32 {
	return solver2_addGate(s, [][]lit{{toLitCond(v, complement)}}, mark, id)
}
func sat_solver2_add_buffer(s *sat_solver2, a, b, complement, mark, id int32) int32 {
	return solver2_addGate(s, [][]lit{{toLit(a), toLitCond(b, 1-complement)}, {toLitCond(a, 1), toLitCond(b, complement)}}, mark, id)
}
func sat_solver2_add_and(s *sat_solver2, v, a, b, ca, cb, mark, id int32) int32 {
	return solver2_addGate(s, [][]lit{{toLitCond(v, 1), toLitCond(a, ca)}, {toLitCond(v, 1), toLitCond(b, cb)}, {toLit(v), toLitCond(a, 1-ca), toLitCond(b, 1-cb)}}, mark, id)
}
func sat_solver2_add_xor(s *sat_solver2, a, b, c, complement, mark, id int32) int32 {
	return solver2_addGate(s, [][]lit{{toLitCond(a, 1-complement), toLitCond(b, 1), toLitCond(c, 1)}, {toLitCond(a, 1-complement), toLit(b), toLit(c)}, {toLitCond(a, complement), toLitCond(b, 1), toLit(c)}, {toLitCond(a, complement), toLit(b), toLitCond(c, 1)}}, mark, id)
}
func sat_solver2_add_constraint(s *sat_solver2, v, w, complement, mark, id int32) int32 {
	return solver2_addGate(s, [][]lit{{toLitCond(v, complement), toLit(w)}}, mark, id)
}

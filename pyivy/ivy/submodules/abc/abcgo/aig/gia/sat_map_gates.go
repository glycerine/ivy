package gia

import "github.com/glycerine/abcgo/sat/bsat"

func sat_solver_add_and1(s *bsat.Solver, v, a, b, ca, cb, cv int32) int32 {
	lit := func(v int32, neg bool) int32 {
		if neg {
			return 2*v + 1
		}
		return 2 * v
	}
	if !s.AddClause(lit(v, cv == 0), lit(a, ca != 0)) || !s.AddClause(lit(v, cv == 0), lit(b, cb != 0)) {
		panic("gia: inconsistent half AND")
	}
	return 3
}
func sat_solver_add_and2(s *bsat.Solver, v, a, b, ca, cb, cv int32) int32 {
	lit := func(v int32, neg bool) int32 {
		if neg {
			return 2*v + 1
		}
		return 2 * v
	}
	if !s.AddClause(lit(v, cv != 0), lit(a, ca == 0), lit(b, cb == 0)) {
		panic("gia: inconsistent half AND")
	}
	return 3
}

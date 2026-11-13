package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Gia_ManTestOnePair_rec(s *bsat.Solver, p *gia.Gia_Man_t, id int32, vars []int32) int32 {
	if vars[id] >= 0 {
		return vars[id]
	}
	v := s.NewVariable()
	vars[id] = v
	obj := gia.Gia_ManObj(p, id)
	if gia.Gia_ObjIsAnd(obj) != 0 {
		a := Gia_ManTestOnePair_rec(s, p, gia.Gia_ObjFaninId0(obj, id), vars)
		b := Gia_ManTestOnePair_rec(s, p, gia.Gia_ObjFaninId1(obj, id), vars)
		l0 := bsat.Literal(a<<1 | gia.Gia_ObjFaninC0(obj))
		l1 := bsat.Literal(b<<1 | gia.Gia_ObjFaninC1(obj))
		s.AddClause(bsat.Negative(v), l0)
		s.AddClause(bsat.Negative(v), l1)
		s.AddClause(bsat.Positive(v), l0^1, l1^1)
	}
	return v
}

// Gia_ManTestOnePair returns the SAT solver status for a difference between
// the two nodes, allowing the second node to be complemented by phase.
func Gia_ManTestOnePair(p *gia.Gia_Man_t, first, second, phase int32) int32 {
	s := bsat.NewSolver()
	vars := make([]int32, gia.Gia_ManObjNum(p))
	for i := range vars {
		vars[i] = -1
	}
	vars[0] = s.NewVariable()
	s.AddClause(bsat.Negative(vars[0]))
	a := Gia_ManTestOnePair_rec(s, p, first, vars)
	b := Gia_ManTestOnePair_rec(s, p, second, vars)
	l0, l1 := bsat.Negative(a), bsat.Literal(b<<1|phase)
	status := s.Solve(l0, l1)
	if status == bsat.Unsatisfiable {
		status = s.Solve(l0^1, l1^1)
	}
	return int32(status)
}

package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Sbm_AddSorter(s *bsat.Solver, vars []int32, i, k int32, next *int32) {
	a, b := *next, *next+1
	*next += 2
	add := func(c ...int32) {
		if !s.AddClause(c...) {
			panic("gia: inconsistent cardinality sorter")
		}
	}
	add(2*a, 2*vars[i]+1)
	add(2*a, 2*vars[k]+1)
	add(2*b, 2*vars[i]+1, 2*vars[k]+1)
	vars[i], vars[k] = a, b
}
func Sbm_AddCardinConstrMerge(s *bsat.Solver, vars []int32, lo, hi, r int32, next *int32) {
	step := 2 * r
	if step < hi-lo {
		Sbm_AddCardinConstrMerge(s, vars, lo, hi-r, step, next)
		Sbm_AddCardinConstrMerge(s, vars, lo+r, hi, step, next)
		for i := lo + r; i < hi-r; i += step {
			Sbm_AddSorter(s, vars, i, i+r, next)
		}
		for i := lo + r; i < hi-r-1; i += r {
			if !s.AddClause(2*vars[i], 2*vars[i+r]+1) {
				panic("gia: inconsistent cardinality merge")
			}
		}
	}
}
func Sbm_AddCardinConstrRange(s *bsat.Solver, vars []int32, lo, hi int32, next *int32) {
	if hi-lo >= 1 {
		mid := lo + (hi-lo)/2
		for i := lo; i <= mid; i++ {
			Sbm_AddSorter(s, vars, i, i+(hi-lo+1)/2, next)
		}
		Sbm_AddCardinConstrRange(s, vars, lo, mid, next)
		Sbm_AddCardinConstrRange(s, vars, mid+1, hi, next)
		Sbm_AddCardinConstrMerge(s, vars, lo, hi, 1, next)
	}
}
func Sbm_AddCardinConstrPairWise(s *bsat.Solver, vars *Vec_Int_t, _ int32) int32 {
	next := int32(vec.Vec_IntSize(vars))
	Sbm_AddCardinConstrRange(s, vec.Vec_IntArray(vars), 0, next-1, &next)
	return next
}
func Sbm_AddCardinSolver(logN int32, output **Vec_Int_t) *bsat.Solver {
	if logN < 0 || logN > 20 {
		panic("gia: invalid cardinality size")
	}
	n := int32(1) << logN
	vars := vec.Vec_IntAlloc(int(n))
	for i := int32(0); i < n; i++ {
		vec.Vec_IntPush(vars, i)
	}
	solver := bsat.NewSolver()
	for solver.NumVars() < n {
		solver.NewVariable()
	}
	Sbm_AddCardinConstrPairWise(solver, vars, 2)
	*output = vars
	return solver
}

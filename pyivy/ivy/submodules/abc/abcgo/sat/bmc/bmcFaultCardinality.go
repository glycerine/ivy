package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"io"
	"os"
)

func cardClause(s *bsat.Solver, lits ...bsat.Literal) {
	if !s.AddClause(lits...) {
		panic("bmc: inconsistent cardinality constraint")
	}
}

// Cnf_AddSorder adds a two-input descending sorting comparator.
func Cnf_AddSorder(s *bsat.Solver, vars []int32, i, k int32, next *int32) {
	if s == nil || next == nil || i < 0 || k < 0 || int(i) >= len(vars) || int(k) >= len(vars) || *next != s.NumVars() {
		panic("bmc: invalid sorter")
	}
	a, b := vars[i], vars[k]
	high, low := s.NewVariable(), s.NewVariable()
	*next = s.NumVars()
	cardClause(s, bsat.Negative(a), bsat.Positive(high))
	cardClause(s, bsat.Negative(b), bsat.Positive(high))
	cardClause(s, bsat.Positive(a), bsat.Positive(b), bsat.Negative(high))
	cardClause(s, bsat.Positive(a), bsat.Negative(low))
	cardClause(s, bsat.Positive(b), bsat.Negative(low))
	cardClause(s, bsat.Negative(a), bsat.Negative(b), bsat.Positive(low))
	vars[i], vars[k] = high, low
}
func Cnf_AddCardinConstrMerge(s *bsat.Solver, vars []int32, lo, hi, stride int32, next *int32) {
	step := stride * 2
	if step < hi-lo {
		Cnf_AddCardinConstrMerge(s, vars, lo, hi-stride, step, next)
		Cnf_AddCardinConstrMerge(s, vars, lo+stride, hi, step, next)
		for i := lo + stride; i < hi-stride; i += step {
			Cnf_AddSorder(s, vars, i, i+stride, next)
		}
	}
}
func Cnf_AddCardinConstrRange(s *bsat.Solver, vars []int32, lo, hi int32, next *int32) {
	if hi-lo >= 1 {
		mid := lo + (hi-lo)/2
		for i := lo; i <= mid; i++ {
			Cnf_AddSorder(s, vars, i, i+(hi-lo+1)/2, next)
		}
		Cnf_AddCardinConstrRange(s, vars, lo, mid, next)
		Cnf_AddCardinConstrRange(s, vars, mid+1, hi, next)
		Cnf_AddCardinConstrMerge(s, vars, lo, hi, 1, next)
	}
}
func cardSort(s *bsat.Solver, input []int32, k, strict int32) {
	if s == nil || len(input) < 2 || k <= 0 || int(k) >= len(input) {
		panic("bmc: invalid cardinality constraint")
	}
	for _, v := range input {
		if v < 0 || v >= s.NumVars() {
			panic("bmc: invalid cardinality variable")
		}
	}
	size := 1
	for size < len(input) {
		size *= 2
	}
	work := append([]int32(nil), input...)
	if size > len(work) {
		zero := s.NewVariable()
		cardClause(s, bsat.Negative(zero))
		for len(work) < size {
			work = append(work, zero)
		}
	}
	next := s.NumVars()
	Cnf_AddCardinConstrRange(s, work, 0, int32(size-1), &next)
	cardClause(s, bsat.Negative(work[k]))
	if strict != 0 {
		cardClause(s, bsat.Positive(work[k-1]))
	}
}
func Cnf_AddCardinConstrPairWise(s *bsat.Solver, vars *vec.Vec_Int_t, k, strict int32) {
	cardSort(s, vec.Vec_IntArray(vars), k, strict)
	vec.Vec_IntClear(vars)
}
func Cnf_AddCardinVar(level, base int32, vars *vec.Vec_Int_t, k int32) int32 {
	if level != 0 {
		return base + k
	}
	return vec.Vec_IntEntry(vars, int(k))
}
func Cnf_AddCardinConstrGeneral(s *bsat.Solver, vars *vec.Vec_Int_t, k, strict int32) {
	cardSort(s, vec.Vec_IntArray(vars), k, strict)
}

// Cnf_AddCardinConstr constrains at most one input to be true and leaves
// their OR in vars. Each pair requires one aggregate variable.
func Cnf_AddCardinConstr(s *bsat.Solver, vars *vec.Vec_Int_t) int32 {
	for vec.Vec_IntSize(vars) > 1 {
		n := vec.Vec_IntSize(vars)
		write := 0
		for i := 0; i+1 < n; i += 2 {
			a, b := vec.Vec_IntEntry(vars, i), vec.Vec_IntEntry(vars, i+1)
			out := s.NewVariable()
			cardClause(s, bsat.Negative(a), bsat.Negative(b))
			cardClause(s, bsat.Negative(a), bsat.Positive(out))
			cardClause(s, bsat.Negative(b), bsat.Positive(out))
			cardClause(s, bsat.Positive(a), bsat.Positive(b), bsat.Negative(out))
			vec.Vec_IntWriteEntry(vars, write, out)
			write++
		}
		if n&1 != 0 {
			vec.Vec_IntWriteEntry(vars, write, vec.Vec_IntEntry(vars, n-1))
			write++
		}
		vec.Vec_IntShrink(vars, write)
	}
	return s.NumVars()
}

// Cnf_AddCardinConstrTest prints all eight-bit patterns with exactly two bits set.
func Cnf_AddCardinConstrTest() { printCardinalityPatterns(os.Stdout) }

func printCardinalityPatterns(out io.Writer) int {
	s := bsat.NewSolver()
	vars := vec.Vec_IntAlloc(8)
	for i := 0; i < 8; i++ {
		vec.Vec_IntPush(vars, s.NewVariable())
	}
	Cnf_AddCardinConstrPairWise(s, vars, 2, 1)
	count := 0
	for s.Solve() == bsat.Satisfiable {
		count++
		fmt.Fprintf(out, "%3d : ", count)
		block := make([]bsat.Literal, 8)
		for i := range block {
			value, ok := s.Value(int32(i))
			if !ok {
				panic("bmc: missing solver model")
			}
			if value {
				fmt.Fprint(out, 1)
				block[i] = bsat.Negative(int32(i))
			} else {
				fmt.Fprint(out, 0)
				block[i] = bsat.Positive(int32(i))
			}
		}
		fmt.Fprintln(out)
		if !s.AddClause(block...) {
			break
		}
	}
	return count
}

package bsat

import (
	"fmt"
	"reflect"
)

/*
*************************************************************************************************
MiniSat -- Copyright (c) 2005, Niklas Sorensson
http://www.cs.chalmers.se/Cs/Research/FormalMethods/MiniSat/

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*************************************************************************************************
*/
func check(expr int32) {
	if expr == 0 {
		panic("solver check failed")
	}
}
func printlits(literals []lit) {
	for _, l := range literals {
		sign := ""
		if lit_sign(l) != 0 {
			sign = "~"
		}
		fmt.Printf("%sx%d ", sign, lit_var(l))
	}
}
func selectionsort(array []any, size int32, compare func(any, any) int32) {
	array = array[:size]
	for i := 0; i < len(array)-1; i++ {
		best := i
		for j := i + 1; j < len(array); j++ {
			if compare(array[j], array[best]) < 0 {
				best = j
			}
		}
		array[i], array[best] = array[best], array[i]
	}
}
func sortrnd(array []any, size int32, compare func(any, any) int32, seed *float64) {
	array = array[:size]
	if size <= 15 {
		selectionsort(array, size, compare)
		return
	}
	pivot := array[irand(seed, size)]
	i, j := int32(-1), size
	for {
		for {
			i++
			if compare(array[i], pivot) >= 0 {
				break
			}
		}
		for {
			j--
			if compare(pivot, array[j]) >= 0 {
				break
			}
		}
		if i >= j {
			break
		}
		array[i], array[j] = array[j], array[i]
	}
	sortrnd(array[:i], i, compare, seed)
	sortrnd(array[i:], size-i, compare, seed)
}
func sat_clause_compute_lbd(s *sat_solver, c *clause) int32 {
	return satClauseLBD(&s.satAssignments, c)
}

// Memory is an owned-payload estimate, excluding runtime/map bookkeeping.
func sat_solver_memory(s *sat_solver) float64 {
	total := uint64(reflect.TypeOf(*s).Size())
	for _, a := range [][]int32{s.levels, s.assigns, s.polarity, s.tags, s.loads, s.orderpos, s.reasons, s.trail, s.model} {
		total += uint64(cap(a)) * 4
	}
	total += uint64(cap(s.activity)+cap(s.activity2)+cap(s.factors))*8 + uint64(cap(s.pFreqs)) + uint64(cap(s.wlists))*uint64(reflect.TypeOf(veci{}).Size())
	for _, w := range s.wlists {
		total += uint64(cap(w.ptr)) * 4
	}
	for _, v := range []*veci{&s.order, &s.trail_lim, &s.tagged, &s.stack, &s.act_vars, &s.unit_lits, &s.act_clas, &s.temp_clause, &s.conf_final, &s.user_vars, &s.user_values, &s.pivot_vars} {
		total += uint64(cap(v.ptr)) * 4
	}
	for _, c := range s.Mem.clauses {
		total += uint64(reflect.TypeOf(*c).Size()) + uint64(cap(c.lits))*4
	}
	if p := solverClauseStore(s); p != nil {
		total += uint64(Sto_ManMemoryReport(p))
	}
	return float64(total)
}
func sat_solver_nvars(s *sat_solver) int32      { return s.size }
func sat_solver_nclauses(s *sat_solver) int32   { return int32(s.stats.clauses) }
func sat_solver_nconflicts(s *sat_solver) int32 { return int32(s.stats.conflicts) }

// array.go provides Z3 array theory operations: ArraySort, Select, Store,
// ConstArray, and array sort introspection.
// Ported from Python ivy_solver.py z3.Select/z3.Update/z3.K/z3.ArraySort.
package z3bridge

/*
#include <z3.h>
*/
import "C"
import "runtime"

// ArraySort creates a Z3 array sort: Array(domain, range).
// Wraps Z3_mk_array_sort.
// Corresponds to Python's z3.ArraySort(domain, range).
func (ctx *Z3Context) ArraySort(domain, rng Z3Sort) Z3Sort {
	var s Z3Sort
	ctx.do(func() {
		s = ctx.newSort(C.Z3_mk_array_sort(ctx.c, domain.c, rng.c))
	})
	runtime.KeepAlive(domain)
	runtime.KeepAlive(rng)
	return s
}

// Select reads an element from an array: array[index].
// Wraps Z3_mk_select.
// Corresponds to Python's z3.Select(array, index).
func (ctx *Z3Context) Select(array, index Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_select(ctx.c, array.c, index.c))
	})
	runtime.KeepAlive(array)
	runtime.KeepAlive(index)
	return r
}

// Store writes an element to an array: array[index] = value.
// Wraps Z3_mk_store.
// Corresponds to Python's z3.Update(array, index, value).
func (ctx *Z3Context) Store(array, index, value Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_store(ctx.c, array.c, index.c, value.c))
	})
	runtime.KeepAlive(array)
	runtime.KeepAlive(index)
	runtime.KeepAlive(value)
	return r
}

// ConstArray creates a constant array where every index maps to the same value.
// Wraps Z3_mk_const_array.
// Corresponds to Python's z3.K(domain, value).
func (ctx *Z3Context) ConstArray(domain Z3Sort, value Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_const_array(ctx.c, domain.c, value.c))
	})
	runtime.KeepAlive(domain)
	runtime.KeepAlive(value)
	return r
}

// ArrayDomain returns the domain sort of an array sort.
// Wraps Z3_get_array_sort_domain.
func (s Z3Sort) ArrayDomain() Z3Sort {
	var r Z3Sort
	s.ctx.do(func() {
		r = s.ctx.newSort(C.Z3_get_array_sort_domain(s.ctx.c, s.c))
	})
	runtime.KeepAlive(s)
	return r
}

// ArrayRange returns the range sort of an array sort.
// Wraps Z3_get_array_sort_range.
func (s Z3Sort) ArrayRange() Z3Sort {
	var r Z3Sort
	s.ctx.do(func() {
		r = s.ctx.newSort(C.Z3_get_array_sort_range(s.ctx.c, s.c))
	})
	runtime.KeepAlive(s)
	return r
}

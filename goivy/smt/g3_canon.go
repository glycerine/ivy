//go:build !js && !classic_z3

// canon_z3.go provides a hand-rolled deterministic s-expression printer
// for Z3 ASTs. The output is byte-identical between Go and Python (pyivy)
// for the same logical formula, regardless of construction order or Z3
// internal state. This is the only smt2-fingerprint format that aligns
// across Go's CGO bindings and pyivy's z3py bindings — Z3's own
// Z3_benchmark_to_smtlib_string and Z3_solver_to_string both emit
// $xN/?xN alias names derived from `expr->get_id()`, which is a
// per-context internal counter that differs between bindings.
//
// Two specific Python-vs-Go non-determinisms must be neutralized:
//
//  1. Quantifier bound-variable order. Python's logic.ForAll stores
//     bound variables as a frozenset (logic.py:380). Iterating a
//     frozenset gives hash-randomized order, so successive Python runs
//     can build the same logical forall with bound vars in different
//     positions in the Z3 AST. Even within a single run, this differs
//     from Go's deterministic alphabetical order.
//
//     Fix: when emitting a quantifier, sort the bound-variable list
//     alphabetically by name. This produces a stable binder list.
//
//  2. Bound-variable references in the body. Z3 substitutes constants
//     with de Bruijn indices based on their POSITION in the binder
//     list. Different binder orders → different indices in the body
//     for the same logical variable.
//
//     Fix: when emitting a `(var idx ...)` reference, look up the
//     ORIGINAL name (from the AST's binder list at that index) and
//     emit it instead of the index. The lookup uses the convention
//     that de Bruijn 0 is the LAST entry in the binder list, so
//     the AST-order names[len-1-idx] is the original name.
//
// With these two fixes, the canon string is purely a function of
// the logical formula, independent of Z3's bound-variable ordering
// choices.
//
// Format:
//
//	Top-level (returned by CanonZ3Assertions):
//	  "(asserts <expr1> <expr2> ... <exprN>)"
//
//	Expressions:
//	  (n <numeral_string>)             — numeric literal
//	  (v <bound_var_name>)             — bound variable, looked up by name
//	  (c <name> <sort>)                — 0-ary application (constant)
//	  (a <name> <arg1> <arg2> ...)     — n-ary application
//	  (forall weight:<w> ((<bn1> <s1>) ...) <body>) — bound vars sorted by name
//	  (exists weight:<w> ((<bn1> <s1>) ...) <body>) — bound vars sorted by name
//
//	Sorts:
//	  <sort_name>                       — bare symbol from Z3_get_sort_name
//
// The Python mirror lives in pyivy/ivy/ivy/ivy_solver.py
// (_canon_z3_assertions, _canon_z3_expr, _canon_z3_sort) and produces
// the EXACT same byte string for the same logical solver state.
package smt

import (
	"fmt"
	"sort"
	"strings"
)

func (s *Z3Solver) CanonZ3Assertions() string {
	var sb strings.Builder
	sb.WriteString("(asserts")
	for _, e := range s.assertions() {
		sb.WriteByte(' ')
		canonG3Expr(e, nil, &sb)
	}
	sb.WriteByte(')')
	return sb.String()
}
func canonG3Expr(e Z3Expr, binders [][]string, sb *strings.Builder) {
	switch {
	case e.IsNumeral():
		sb.WriteString("(n ")
		sb.WriteString(textResult(e.ctx.c.NumeralString(e.c)))
		sb.WriteByte(')')
	case e.IsVar():
		sb.WriteString("(v ")
		sb.WriteString(lookupBinderName(binders, e.VarIndex()))
		sb.WriteByte(')')
	case e.IsApp():
		fd := e.Decl()
		name := fd.Name()
		n := e.NumArgs()
		if n == 0 {
			sb.WriteString("(c ")
			sb.WriteString(name)
			sb.WriteByte(' ')
			sb.WriteString(textResult(e.ctx.c.SortName(e.ExprSort().c)))
			sb.WriteByte(')')
		} else {
			sb.WriteString("(a ")
			sb.WriteString(name)
			for i := 0; i < n; i++ {
				sb.WriteByte(' ')
				canonG3Expr(e.Arg(i), binders, sb)
			}
			sb.WriteByte(')')
		}
	case e.IsQuantifier():
		n := e.QuantNumVars()
		names := make([]string, n)
		type pair struct{ name, sort string }
		pairs := make([]pair, n)
		for i := range names {
			names[i] = e.QuantVarName(i)
			pairs[i] = pair{names[i], textResult(e.ctx.c.SortName(e.QuantVarSort(i).c))}
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].name < pairs[j].name })
		w, err := e.ctx.c.QuantifierWeight(e.c)
		fail(err)
		if e.IsForAll() {
			fmt.Fprintf(sb, "(forall weight:%d (", w)
		} else {
			fmt.Fprintf(sb, "(exists weight:%d (", w)
		}
		for i, p := range pairs {
			if i > 0 {
				sb.WriteByte(' ')
			}
			fmt.Fprintf(sb, "(%s %s)", p.name, p.sort)
		}
		sb.WriteString(") ")
		canonG3Expr(e.QuantBody(), append(binders, names), sb)
		sb.WriteByte(')')
	default:
		sb.WriteString("(unknown)")
	}
}
func lookupBinderName(binders [][]string, idx int) string {
	for i := len(binders) - 1; i >= 0; i-- {
		scope := binders[i]
		n := len(scope)
		if idx < n {
			return scope[n-1-idx]
		}
		idx -= n
	}
	return "??unbound"
}

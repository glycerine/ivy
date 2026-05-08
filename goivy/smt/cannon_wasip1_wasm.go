//go:build wasip1

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
//	  (forall ((<bn1> <s1>) ...) <body>) — bound vars sorted by name
//	  (exists ((<bn1> <s1>) ...) <body>) — bound vars sorted by name
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
	"runtime"
	"sort"
	"strings"
)

// CanonZ3Assertions returns a deterministic s-expression representation
// of the solver's current assertions, suitable for cross-language hash
// comparison between goivy and pyivy. See the file header comment for
// the format spec.
func (s *Z3Solver) CanonZ3Assertions() string {
	var res string
	s.ctx.do(func() {
		// Get assertions as an AST vector.
		vec := z3Solver_get_assertions(s.ctx.c, s.c)
		z3ASTVector_inc_ref(s.ctx.c, vec)
		defer z3ASTVector_dec_ref(s.ctx.c, vec)

		n := int(z3ASTVector_size(s.ctx.c, vec))
		var sb strings.Builder
		sb.WriteString("(asserts")
		for i := 0; i < n; i++ {
			ast := z3ASTVector_get(s.ctx.c, vec, uint32(i))
			sb.WriteByte(' ')
			canonZ3ExprUnlocked(s.ctx.c, ast, nil, &sb)
		}
		sb.WriteByte(')')
		res = sb.String()
	})
	runtime.KeepAlive(s)
	return res
}

// canonZ3ExprUnlocked recursively walks a Z3 expression AST and writes
// a canonical s-expression to sb. MUST be called with the Z3 context
// lock held (i.e., from inside an s.ctx.do(...) block). Uses raw CGO
// calls instead of the Expr/Sort wrappers to avoid re-entering ctx.do.
//
// binders is a stack of bound-variable name lists, one entry per
// enclosing quantifier (innermost LAST). Each entry is the AST-order
// list of names for that quantifier (NOT the sorted-canon order); de
// Bruijn lookups must use the AST order.
func canonZ3ExprUnlocked(c z3Context, ast z3AST, binders [][]string, sb *strings.Builder) {
	kind := z3_get_ast_kind(c, ast)
	switch kind {
	case z3_NUMERAL_AST:
		sb.WriteString("(n ")
		cstr := z3_get_numeral_string(c, ast)
		sb.WriteString(z3String(cstr))
		sb.WriteByte(')')

	case z3_VAR_AST:
		idx := int(z3_get_index_value(c, ast))
		name := lookupBinderName(binders, idx)
		sb.WriteString("(v ")
		sb.WriteString(name)
		sb.WriteByte(')')

	case z3_APP_AST:
		app := z3_to_app(c, ast)
		decl := z3_get_app_decl(c, app)
		sym := z3_get_decl_name(c, decl)
		name := z3String(z3_get_symbol_string(c, sym))
		nargs := int(z3_get_app_num_args(c, app))
		if nargs == 0 {
			sb.WriteString("(c ")
			sb.WriteString(name)
			sb.WriteByte(' ')
			srt := z3_get_sort(c, ast)
			canonZ3SortUnlocked(c, srt, sb)
			sb.WriteByte(')')
		} else {
			sb.WriteString("(a ")
			sb.WriteString(name)
			for i := 0; i < nargs; i++ {
				sb.WriteByte(' ')
				arg := z3_get_app_arg(c, app, uint32(i))
				canonZ3ExprUnlocked(c, arg, binders, sb)
			}
			sb.WriteByte(')')
		}

	case z3_QUANTIFIER_AST:
		nbound := int(z3_get_quantifier_num_bound(c, ast))

		// Collect AST-order names and sort names. The AST-order names
		// are needed to push onto the binder stack for body lookups
		// (de Bruijn indices reference these positions). The canon
		// output below sorts the (name, sort) pairs alphabetically by
		// name to neutralize Python's frozenset-randomized order.
		astNames := make([]string, nbound)
		type bvPair struct{ name, sort string }
		pairs := make([]bvPair, nbound)
		for i := 0; i < nbound; i++ {
			nameSym := z3_get_quantifier_bound_name(c, ast, uint32(i))
			astNames[i] = z3String(z3_get_symbol_string(c, nameSym))
			bsort := z3_get_quantifier_bound_sort(c, ast, uint32(i))
			sortNameSym := z3_get_sort_name(c, bsort)
			pairs[i] = bvPair{astNames[i], z3String(z3_get_symbol_string(c, sortNameSym))}
		}
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].name < pairs[j].name
		})

		if z3_is_quantifier_forall(c, ast) != 0 {
			sb.WriteString("(forall (")
		} else {
			sb.WriteString("(exists (")
		}
		for i, p := range pairs {
			if i > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteByte('(')
			sb.WriteString(p.name)
			sb.WriteByte(' ')
			sb.WriteString(p.sort)
			sb.WriteByte(')')
		}
		sb.WriteString(") ")

		// Push the AST-order names so body var lookups resolve correctly.
		body := z3_get_quantifier_body(c, ast)
		newBinders := append(binders, astNames)
		canonZ3ExprUnlocked(c, body, newBinders, sb)
		sb.WriteByte(')')

	default:
		fmt.Fprintf(sb, "(unknown_kind=%d)", int(kind))
	}
}

// canonZ3SortUnlocked writes the canonical name of a Z3 sort to sb.
// MUST be called with the Z3 context lock held.
func canonZ3SortUnlocked(c z3Context, s z3Sort, sb *strings.Builder) {
	sym := z3_get_sort_name(c, s)
	name := z3String(z3_get_symbol_string(c, sym))
	sb.WriteString(name)
}

// lookupBinderName resolves a de Bruijn index to a bound-variable name
// using the binder stack. The convention: de Bruijn index 0 refers to
// the innermost binder = the LAST entry of the topmost binder list.
// Indices that exceed the topmost scope size walk outward.
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

// canon_z3.go provides a hand-rolled deterministic s-expression printer
// for Z3 ASTs. The output is byte-identical between Go and Python (pyivy)
// for the same logical formula, regardless of construction order or Z3
// internal state. This is the only smt2-fingerprint format that aligns
// across Go's CGO bindings and pyivy's z3py bindings — Z3's own
// Z3_benchmark_to_smtlib_string and Z3_solver_to_string both emit
// $xN/?xN alias names derived from `expr->get_id()`, which is a
// per-context internal counter that differs between bindings.
//
// The format is intentionally compact and unambiguous:
//
//   Top-level (returned by CanonZ3Assertions):
//     "(asserts <expr1> <expr2> ... <exprN>)"
//
//   Expressions:
//     (n <numeral_string>)             — numeric literal
//     (v <de_bruijn_idx> <sort>)       — bound variable
//     (c <name> <sort>)                — 0-ary application (constant)
//     (a <name> <arg1> <arg2> ...)     — n-ary application
//     (forall ((<bn1> <s1>) ...) <body>)
//     (exists ((<bn1> <s1>) ...) <body>)
//
//   Sorts:
//     <sort_name>                       — bare symbol from Z3_get_sort_name
//
// The Python mirror lives in pyivy/ivy/ivy/ivy_solver.py
// (_canon_z3_assertions, _canon_z3_expr, _canon_z3_sort) and produces
// the EXACT same byte string for the same logical solver state.
package z3bridge

/*
#include <z3.h>
*/
import "C"

import (
	"fmt"
	"runtime"
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
		vec := C.Z3_solver_get_assertions(s.ctx.c, s.c)
		C.Z3_ast_vector_inc_ref(s.ctx.c, vec)
		defer C.Z3_ast_vector_dec_ref(s.ctx.c, vec)

		n := int(C.Z3_ast_vector_size(s.ctx.c, vec))
		var sb strings.Builder
		sb.WriteString("(asserts")
		for i := 0; i < n; i++ {
			ast := C.Z3_ast_vector_get(s.ctx.c, vec, C.uint(i))
			sb.WriteByte(' ')
			canonZ3ExprUnlocked(s.ctx.c, ast, &sb)
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
func canonZ3ExprUnlocked(c C.Z3_context, ast C.Z3_ast, sb *strings.Builder) {
	kind := C.Z3_get_ast_kind(c, ast)
	switch kind {
	case C.Z3_NUMERAL_AST:
		sb.WriteString("(n ")
		cstr := C.Z3_get_numeral_string(c, ast)
		sb.WriteString(C.GoString(cstr))
		sb.WriteByte(')')

	case C.Z3_VAR_AST:
		sb.WriteString("(v ")
		idx := C.Z3_get_index_value(c, ast)
		fmt.Fprintf(sb, "%d ", int(idx))
		srt := C.Z3_get_sort(c, ast)
		canonZ3SortUnlocked(c, srt, sb)
		sb.WriteByte(')')

	case C.Z3_APP_AST:
		app := C.Z3_to_app(c, ast)
		decl := C.Z3_get_app_decl(c, app)
		sym := C.Z3_get_decl_name(c, decl)
		name := C.GoString(C.Z3_get_symbol_string(c, sym))
		nargs := int(C.Z3_get_app_num_args(c, app))
		if nargs == 0 {
			sb.WriteString("(c ")
			sb.WriteString(name)
			sb.WriteByte(' ')
			srt := C.Z3_get_sort(c, ast)
			canonZ3SortUnlocked(c, srt, sb)
			sb.WriteByte(')')
		} else {
			sb.WriteString("(a ")
			sb.WriteString(name)
			for i := 0; i < nargs; i++ {
				sb.WriteByte(' ')
				arg := C.Z3_get_app_arg(c, app, C.uint(i))
				canonZ3ExprUnlocked(c, arg, sb)
			}
			sb.WriteByte(')')
		}

	case C.Z3_QUANTIFIER_AST:
		if bool(C.Z3_is_quantifier_forall(c, ast)) {
			sb.WriteString("(forall (")
		} else {
			sb.WriteString("(exists (")
		}
		nbound := int(C.Z3_get_quantifier_num_bound(c, ast))
		for i := 0; i < nbound; i++ {
			if i > 0 {
				sb.WriteByte(' ')
			}
			nameSym := C.Z3_get_quantifier_bound_name(c, ast, C.uint(i))
			bsort := C.Z3_get_quantifier_bound_sort(c, ast, C.uint(i))
			sb.WriteByte('(')
			sb.WriteString(C.GoString(C.Z3_get_symbol_string(c, nameSym)))
			sb.WriteByte(' ')
			canonZ3SortUnlocked(c, bsort, sb)
			sb.WriteByte(')')
		}
		sb.WriteString(") ")
		body := C.Z3_get_quantifier_body(c, ast)
		canonZ3ExprUnlocked(c, body, sb)
		sb.WriteByte(')')

	default:
		sb.WriteString("(unknown)")
	}
}

// canonZ3SortUnlocked writes the canonical name of a Z3 sort to sb.
// MUST be called with the Z3 context lock held.
func canonZ3SortUnlocked(c C.Z3_context, s C.Z3_sort, sb *strings.Builder) {
	sym := C.Z3_get_sort_name(c, s)
	name := C.GoString(C.Z3_get_symbol_string(c, sym))
	sb.WriteString(name)
}

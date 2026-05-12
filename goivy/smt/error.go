package smt

import "fmt"

// ErrMsg reports a Z3 API error observed
// at the smt boundary.
//
// Z3 reports API misuse and other external
// errors through the context error
// code and the registered error handler.
// The handler must not throw across the
// wasm/native boundary; smt checks the
// context error state after Z3 calls and
// raises this error from normal Go control
// flow when the existing API shape has
// no error return.
//
// In the browser js/wasm backend, the registered
// Z3 error handler is installed by
// webvue/src/workers/smtZ3Imports.js because Z3 wasm
// needs a function-table callback supplied by
// the JavaScript host. That callback is essential,
// but it is deliberately dumb: it records the Z3
// callback event and returns normally.
//
// The Go smt layer remains the semantic owner of
// errors by reading Z3_get_error_code/Z3_get_error_msg
// after each Z3 call and converting the
// result into ErrMsg from the Go call stack.
//
// Thus the browser version of goivy can still
// panic/recover and unwind a stack, just like
// the Go side does. It turns out that the number
// of places that Z3 can issue errors is pretty
// large. So we stuck with the pattern that
// matched the original python. Here is the
// audit of the Z3 source against the calls that
// we actually use that backs up that decision.
//
// Z3 4.7.1 + craig interpolation branch inspection
//
// Current count: smt uses 131 distinct underlying
// Z3_* C API calls. Of those, I found 65 with
// a direct external/API error path that can
// invoke the registered handler via SET_ERROR_CODE,
// check_sorts, CHECK_VALID_AST, or helper calls
// like check_numeral_sort.
//
// Representative source anchors:
//
// Handler fires from set_error_code: api_context.cpp (line 147)
//
// Sort/type construction errors come
// through check_sorts: api_context.cpp (line 284)
//
// Macro-generated AST builders all call
// check_sorts: api_util.h (line 111)
//
// Invalid BV sort width is explicit
// Z3_INVALID_ARG: api_bv.cpp (line 26)
//
// Quantifier construction has several explicit
// invalid-usage/arg paths: api_quant.cpp (line 159)
//
// The direct-error list in our current boundary is:
//
// Z3_ast_vector_get, Z3_dec_ref, Z3_get_app_arg, Z3_get_app_decl,
// Z3_get_array_sort_domain, Z3_get_array_sort_range, Z3_get_ast_kind,
// Z3_get_bv_sort_size, Z3_get_domain, Z3_get_index_value,
// Z3_get_numeral_string, Z3_get_quantifier_body,
// Z3_get_quantifier_bound_name, Z3_get_quantifier_bound_sort,
// Z3_get_quantifier_num_bound, Z3_get_range, Z3_get_sort_kind,
// Z3_is_quantifier_forall,
//
// Z3_mk_add, Z3_mk_and, Z3_mk_app, Z3_mk_bv2int, Z3_mk_bv_sort,
// Z3_mk_bvadd, Z3_mk_bvand, Z3_mk_bvashr, Z3_mk_bvlshr,
// Z3_mk_bvmul, Z3_mk_bvnot, Z3_mk_bvor, Z3_mk_bvshl,
// Z3_mk_bvsub, Z3_mk_bvudiv, Z3_mk_bvuge, Z3_mk_bvugt,
// Z3_mk_bvule, Z3_mk_bvult, Z3_mk_bvxor, Z3_mk_concat,
// Z3_mk_const_array, Z3_mk_div, Z3_mk_enumeration_sort,
// Z3_mk_eq, Z3_mk_exists_const, Z3_mk_forall_const, Z3_mk_ge,
// Z3_mk_gt, Z3_mk_iff, Z3_mk_implies, Z3_mk_int2bv,
// Z3_mk_int64, Z3_mk_interpolant, Z3_mk_le, Z3_mk_lt,
// Z3_mk_mul, Z3_mk_not, Z3_mk_or, Z3_mk_select, Z3_mk_store,
// Z3_mk_sub,
//
// Z3_model_get_sort, Z3_model_get_sort_universe,
// Z3_solver_get_model, Z3_solver_pop, Z3_substitute
//
// Another 51 calls are “catch-only”: they can
// invoke the handler if Z3 throws an internal
// z3_exception, but I did not find an external
// misuse branch in their wrapper body. The
// remaining 14 are simple config/cast/equality/message
// helpers where I saw no handler path.
//
// So the practical design answer is: no, not
// every Z3_* call has an external-error callback
// path, but enough of the central constructors/
// accessors do that adding explicit error
// handling after every use would tragically
// junk up the code base for no good reason.
type ErrMsg struct {
	Op   string
	Code int
	Msg  string
}

func (e *ErrMsg) Error() string {
	if e == nil {
		return "z3: <nil>"
	}
	if e.Op != "" {
		return fmt.Sprintf("z3 %s: %s", e.Op, e.Msg)
	}
	return fmt.Sprintf("z3: %s", e.Msg)
}

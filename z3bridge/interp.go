// interp.go provides CGo wrappers for Z3's Craig interpolation API.
// These functions correspond to Z3_mk_interpolant, Z3_mk_interpolation_context,
// and Z3_compute_interpolant from z3_interp.h.
package z3bridge

/*
#include <z3.h>

extern void goZ3BridgeErrorHandler(Z3_context c, Z3_error_code e);
*/
import "C"
import (
	"fmt"
	"runtime"
)

// MkInterpolant marks a formula for interpolation.
// The expression must have Boolean sort.
// Corresponds to Python's z3.Interpolant(a) and C's Z3_mk_interpolant.
func (ctx *Context) MkInterpolant(e Expr) Expr {
	var r Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_interpolant(ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return r
}

// NewInterpolationContext creates a Z3 context suitable for interpolation.
// This uses Z3_mk_interpolation_context instead of Z3_mk_context_rc,
// which provides a legacy solver that supports proof generation required
// for interpolation.
// Corresponds to Python's context created via Z3_mk_interpolation_context.
func NewInterpolationContext() *Context {
	cfg := C.Z3_mk_config()
	defer C.Z3_del_config(cfg)
	c := C.Z3_mk_interpolation_context(cfg)
	C.Z3_set_error_handler(c, (*C.Z3_error_handler)(C.goZ3BridgeErrorHandler))
	ctx := &Context{c: c, syms: make(map[string]C.Z3_symbol)}
	runtime.SetFinalizer(ctx, func(ctx *Context) {
		C.Z3_del_context(ctx.c)
	})
	return ctx
}

// ComputeInterpolant computes an interpolant from an interpolation pattern.
// The pattern should be a conjunction where some subformulas are marked with
// MkInterpolant. Returns the interpolant expressions if UNSAT, or an error
// if SAT/unknown.
// Corresponds to Python's z3.tree_interpolant / Z3_compute_interpolant.
func (ctx *Context) ComputeInterpolant(pattern Expr) ([]Expr, error) {
	var result []Expr
	var resErr error

	ctx.do(func() {
		params := C.Z3_mk_params(ctx.c)
		C.Z3_params_inc_ref(ctx.c, params)
		defer C.Z3_params_dec_ref(ctx.c, params)

		var interp C.Z3_ast_vector
		var model C.Z3_model

		res := C.Z3_compute_interpolant(ctx.c, pattern.c, params, &interp, &model)

		if res == C.Z3_L_FALSE {
			// UNSAT — extract interpolants from the ast_vector
			C.Z3_ast_vector_inc_ref(ctx.c, interp)
			n := int(C.Z3_ast_vector_size(ctx.c, interp))
			result = make([]Expr, n)
			for i := 0; i < n; i++ {
				ast := C.Z3_ast_vector_get(ctx.c, interp, C.uint(i))
				result[i] = ctx.newExpr(ast)
			}
			C.Z3_ast_vector_dec_ref(ctx.c, interp)
		} else if res == C.Z3_L_TRUE {
			resErr = fmt.Errorf("interpolation: formula is satisfiable")
		} else {
			resErr = fmt.Errorf("interpolation: result unknown")
		}
	})
	runtime.KeepAlive(pattern)

	if resErr != nil {
		return nil, resErr
	}
	return result, nil
}

// NewTranslatorWithInterpolation creates a Translator backed by an
// interpolation-capable Z3 context. Use this translator when you need
// to compute Craig interpolants.
func NewTranslatorWithInterpolation() *Translator {
	return &Translator{
		Ctx:    NewInterpolationContext(),
		sorts:  make(map[string]Sort),
		consts: make(map[string]Expr),
		funcs:  make(map[string]FuncDecl),
	}
}

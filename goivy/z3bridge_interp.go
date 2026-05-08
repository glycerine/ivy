//go:build !tinygo

// interp.go provides CGo wrappers for Z3's Craig interpolation API.
// These functions correspond to Z3_mk_interpolant, Z3_mk_interpolation_context,
// and Z3_compute_interpolant from z3_interp.h.
package goivy

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
func (ctx *Z3Context) MkInterpolant(e Z3Expr) Z3Expr {
	var r Z3Expr
	ctx.do(func() {
		r = ctx.newExpr(C.Z3_mk_interpolant(ctx.c, e.c))
	})
	runtime.KeepAlive(e)
	return r
}

// NewInterpolationZ3Context creates a Z3 context suitable for interpolation.
// This uses Z3_mk_interpolation_context instead of Z3_mk_context_rc,
// which provides a legacy solver that supports proof generation required
// for interpolation.
// Corresponds to Python's context created via Z3_mk_interpolation_context.
func NewInterpolationZ3Context() *Z3Context {
	return defaultZ3Backend().NewInterpolationZ3Context()
}

func newCGoInterpolationZ3Context(backend Z3Backend) *Z3Context {
	cfg := C.Z3_mk_config()
	defer C.Z3_del_config(cfg)
	c := C.Z3_mk_interpolation_context(cfg)
	C.Z3_set_error_handler(c, (*C.Z3_error_handler)(C.goZ3BridgeErrorHandler))
	ctx := &Z3Context{backend: backend, c: c, syms: make(map[string]C.Z3_symbol)}
	//runtime.SetFinalizer(ctx, func(ctx *Z3Context) {
	//	ctx.Close()
	//})
	return ctx
}

// ComputeInterpolant computes an interpolant from an interpolation pattern.
// The pattern should be a conjunction where some subformulas are marked with
// MkInterpolant. Returns the interpolant expressions if UNSAT, or an error
// if SAT/unknown.
// Corresponds to Python's z3.tree_interpolant / Z3_compute_interpolant.
func (ctx *Z3Context) ComputeInterpolant(pattern Z3Expr) ([]Z3Expr, error) {
	var result []Z3Expr
	var resErr error

	ctx.do(func() {
		defer func() {
			if r := recover(); r != nil {
				resErr = fmt.Errorf("Z3 interpolation failed: %v", r)
			}
		}()

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
			result = make([]Z3Expr, n)
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
// to compute Craig interpolants. The interpolation translator owns its
// own private cache (different Z3 context, so cannot share with the
// main solver).
func (s *Solver) NewTranslatorWithInterpolation() *Translator {
	cache := &Z3SessionCache{
		Ctx: s.z3Backend().NewInterpolationZ3Context(),
	}
	cache.resetMaps()
	return &Translator{
		s:     s,
		cache: cache,
		Ctx:   cache.Ctx,
	}
}

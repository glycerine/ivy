//go:build js && wasm

// interp.go provides CGo wrappers for Z3's Craig interpolation API.
// These functions correspond to Z3_mk_interpolant, Z3_mk_interpolation_context,
// and Z3_compute_interpolant from z3_interp.h.
package smt

import (
	"fmt"
	"runtime"
)

// MkInterpolant marks a formula for interpolation.
// The expression must have Boolean sort.
// Corresponds to Python's z3.Interpolant(a) and C's Z3_mk_interpolant.
func (ctx *Z3Context) MkInterpolant(e Z3Expr) Z3Expr {
	var r Z3Expr

	r = ctx.newExpr(z3_mk_interpolant(ctx.c, e.c))

	runtime.KeepAlive(e)
	return r
}

// NewInterpolationZ3Context creates a Z3 context suitable for interpolation.
// This uses Z3_mk_interpolation_context instead of Z3_mk_context_rc,
// which provides a legacy solver that supports proof generation required
// for interpolation.
// Corresponds to Python's context created via Z3_mk_interpolation_context.
func NewInterpolationZ3Context() *Z3Context {
	cfg := z3_mk_config()
	defer z3_del_config(cfg)
	c := z3_mk_interpolation_context(cfg)
	z3_set_error_handler(c)
	ctx := &Z3Context{c: c, syms: make(map[string]z3Symbol)}
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

	defer func() {
		if r := recover(); r != nil {
			resErr = fmt.Errorf("Z3 interpolation failed: %v", r)
		}
	}()

	params := z3_mk_params(ctx.c)
	z3Params_inc_ref(ctx.c, params)
	defer z3Params_dec_ref(ctx.c, params)

	var interp z3ASTVector

	res := z3_compute_interpolant(ctx.c, pattern.c, params)
	interp = z3ASTVector(z3_last_u32_result(0))

	if res == z3_L_FALSE {
		// UNSAT — extract interpolants from the ast_vector
		z3ASTVector_inc_ref(ctx.c, interp)
		n := int(z3ASTVector_size(ctx.c, interp))
		result = make([]Z3Expr, n)
		for i := 0; i < n; i++ {
			ast := z3ASTVector_get(ctx.c, interp, uint32(i))
			result[i] = ctx.newExpr(ast)
		}
		z3ASTVector_dec_ref(ctx.c, interp)
	} else if res == z3_L_TRUE {
		resErr = fmt.Errorf("interpolation: formula is satisfiable")
	} else {
		resErr = fmt.Errorf("interpolation: result unknown")
	}

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
/*func (s *Solver) NewTranslatorWithInterpolation() *Translator {
	cache := &Z3SessionCache{
		Ctx: NewInterpolationZ3Context(),
	}
	cache.resetMaps()
	return &Translator{
		s:     s,
		cache: cache,
		Ctx:   cache.Ctx,
	}
}
*/

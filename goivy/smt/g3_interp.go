//go:build !js && !classic_z3

package smt

import (
	"github.com/glycerine/g3/pkg/api"
	"github.com/glycerine/g3/pkg/ast"
)

func NewInterpolationZ3Context() *Z3Context { return NewZ3Context() }
func (ctx *Z3Context) MkInterpolant(e Z3Expr) Z3Expr {
	return ctx.newExpr(handle(ctx.c.Interpolant(e.c)))
}

// g3's interpolation API owns a proof-enabled context. Translate the problem
// and result explicitly so every expression exposed to Ivy retains its original
// context and declaration identity.
func (ctx *Z3Context) ComputeInterpolant(pattern Z3Expr) ([]Z3Expr, error) {
	ic, err := api.NewInterpolationContext()
	if err != nil {
		return nil, err
	}
	defer ic.Close()
	source, err := ctx.c.Manager()
	if err != nil {
		return nil, err
	}
	target, err := ic.Manager()
	if err != nil {
		return nil, err
	}
	forward, err := ast.NewASTTranslation(source, target, true)
	if err != nil {
		return nil, err
	}
	defer forward.Close()
	problem, err := forward.Translate(pattern.c)
	if err != nil {
		return nil, err
	}
	values, err := ic.ComputeInterpolant(problem)
	if err != nil {
		return nil, err
	}
	backward, err := ast.NewASTTranslation(target, source, true)
	if err != nil {
		return nil, err
	}
	defer backward.Close()
	result := make([]Z3Expr, len(values))
	for i, value := range values {
		h, e := backward.Translate(value)
		if e != nil {
			return nil, e
		}
		result[i] = ctx.newExpr(h)
	}
	return result, nil
}

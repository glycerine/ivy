package llb

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// Llb_BddComputeBad forms the disjunction of the AIG's primary outputs. The
// final register-count outputs encode next-state functions and are excluded.
func Llb_BddComputeBad(g *Aig_Man_t, dd *DdManager, timeout abctime) *DdNode {
	if g == nil || dd == nil || len(dd.Vars) < len(g.Inputs()) {
		return nil
	}
	variables := make([]*core.DdNode, len(g.Inputs()))
	for i := range variables {
		variables[i] = core.Cudd_bddIthVar(dd, i)
	}
	functions := llbBuildBDDTable(g, dd, variables)
	if functions == nil {
		return nil
	}
	propertyCount := len(g.Outputs()) - int(g.RegisterCount())
	if propertyCount < 0 {
		return nil
	}
	bad := core.Cudd_ReadLogicZero(dd)
	for i := 0; i < propertyCount; i++ {
		bad = core.Cudd_bddOr(dd, bad, llbBDDForEdge(g, functions, g.Outputs()[i]))
		if bad == nil {
			return nil
		}
	}
	return bad
}

// Llb_BddQuantifyPis existentially removes primary-input variables from f.
func Llb_BddQuantifyPis(g *Aig_Man_t, dd *DdManager, f *DdNode) *DdNode {
	if g == nil || dd == nil || f == nil {
		return nil
	}
	piCount := len(g.Inputs()) - int(g.RegisterCount())
	if piCount < 0 {
		return nil
	}
	cube := core.Cudd_ReadOne(dd)
	for i := 0; i < piCount; i++ {
		index := int(g.RegisterCount()) + i
		cube = core.Cudd_bddAnd(dd, cube, core.Cudd_bddIthVar(dd, index))
		if cube == nil {
			return nil
		}
	}
	return core.Cudd_bddExistAbstract(dd, f, cube)
}

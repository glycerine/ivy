package ivy2go

import "github.com/glycerine/ivy/goivy"

// assign.go mirrors ivy2cpp/assign.go. M4 implements only the simple
// scalar case (no free variables in the LHS). Quantified-LHS
// assignment (nested loops) and the two-phase / large-thunk paths
// are deferred to M5/M8.

// emitAssign dispatches by LHS shape. Mirrors ivy2cpp/action.go
// emitAssign, but the quantified-LHS branches all stub to
// `unsupported` for M4.
func (g *Generator) emitAssign(w *goWriter, a *goivy.LogicAssignAction) {
	vs := goivy.VariablesAstList(a.LHS)
	if len(vs) == 0 {
		g.emitAssignSimple(w, a)
		return
	}
	g.unsupported(w, "quantified-LHS assignment emission deferred (M5)")
}

// emitAssignSimple ports ivy2cpp/assign.go emitAssignSimple, minus
// the trace-LHS emission (M7 lands trace LHS via vprint.go).
func (g *Generator) emitAssignSimple(w *goWriter, a *goivy.LogicAssignAction) {
	lhs, err := g.emitExpr(a.LHS)
	if err != nil {
		g.unsupported(w, "unsupported assignment lhs: %s", err.Error())
		return
	}
	rhs, err := g.emitExpr(a.RHS)
	if err != nil {
		g.unsupported(w, "unsupported assignment rhs: %s", err.Error())
		return
	}
	w.linef("%s = %s", lhs, rhs)
}

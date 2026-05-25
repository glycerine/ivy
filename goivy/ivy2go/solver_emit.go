package ivy2go

// solver_emit.go mirrors ivy2cpp/solver_emit.go. The C++ file emits
// `add(__to_solver(...))` calls per state cell plus forall-quantified
// thunks for large functions; ivy2go's equivalent path goes through
// goivy.Translator and Solver.Assert.
//
// M9 ships a skeleton: per-state-symbol assertion emission that
// generates a comment marker. The real solver assertions land in
// later increments as fixtures exercise them.

// emitSetSolver writes the body of a setSolver method on actionGen<Name>
// that pushes the current State into a goivy.Solver. Skeleton for
// M9 — real assertions are layered in as needed.
func (g *Generator) emitSetSolver(w *goWriter) {
	if !g.usesZ3() {
		return
	}
	w.line("// pushStateIntoSolver asserts the current State into sol.")
	w.line("// M9 skeleton: per-cell assertions are added incrementally as")
	w.line("// fixtures exercise them. The hook lets future emission")
	w.line("// extend this without changing the action_gen call sites.")
	w.linef("func pushStateIntoSolver(_ *%s, _ *goivy.Solver) error {", g.StateTypeName)
	w.line("\t// TODO(M9.1): emit Solver.Assert for each state symbol.")
	w.line("\treturn nil")
	w.line("}")
	w.blank()
}

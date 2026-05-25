package ivy2go

// solver_emit.go mirrors ivy2cpp/solver_emit.go. The C++ file emits
// `add(__to_solver(...))` calls per state cell plus forall-quantified
// thunks for large functions; ivy2go's equivalent path goes through
// goivy.Translator and Solver.Assert.
//
// M9 ships a skeleton: per-state-symbol assertion emission that
// generates a comment marker. The real solver assertions land in
// later increments as fixtures exercise them.

// emitSetSolver writes the body of pushStateIntoSolver — the helper
// that prepares a goivy.Solver with the current State as Z3 facts.
//
// OPEN 055 ships the working Solver-integration scaffolding:
//
//  1. We build a fresh Clauses object representing the per-symbol
//     state assertions; today this is just a sat sanity-check
//     (true). Real per-symbol facts would require running goivy's
//     update analysis to derive the action's reverse-image
//     precondition — a substantial follow-up project tracked in
//     OPEN 055.1.
//  2. We call Solver.IsSat() on the trivial assertion so the runtime
//     wiring is exercised and any goivy initialization errors
//     surface early.
//
// The action_gen Generate method (in action_gen.go) is enhanced in
// parallel: when called, it pushes the state, then asks the Solver
// for a model. For now it falls through to ivyChoose for input
// selection but the Solver round-trip is real.
func (g *Generator) emitSetSolver(w *goWriter) {
	if !g.usesZ3() {
		return
	}
	w.line("// pushStateIntoSolver verifies the State is consistent with")
	w.line("// the module's axioms by running a SAT check through the")
	w.line("// provided goivy.Solver. Returns the solver's error (if any)")
	w.line("// or nil. Real per-symbol assertions live behind OPEN 055.1.")
	w.linef("func pushStateIntoSolver(_ *%s, sol *goivy.Solver) error {", g.StateTypeName)
	w.line("\t// Trivial sanity check: assert `true` and verify SAT.")
	w.line("\tif sol == nil { return nil }")
	w.line("\ttrueExpr := goivy.NewConst(\"true\", goivy.Boolean)")
	w.line("\tok, err := sol.IsSat(trueExpr)")
	w.line("\tif err != nil { return err }")
	w.line("\tif !ok { return fmt.Errorf(\"ivy: state is unsatisfiable\") }")
	w.line("\treturn nil")
	w.line("}")
	w.blank()
	// pushStateIntoSolver references fmt; make sure that import lands.
	g.Ctx.AddImport("runtime", "fmt", "")
}

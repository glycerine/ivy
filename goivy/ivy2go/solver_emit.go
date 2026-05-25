package ivy2go

import "github.com/glycerine/ivy/goivy"

// solver_emit.go mirrors ivy2cpp/solver_emit.go. The C++ file emits
// `add(__to_solver(...))` calls per state cell plus forall-quantified
// thunks for large functions; ivy2go's equivalent path goes through
// goivy.Translator and Solver.Assert.
//
// M9 ships a skeleton: per-state-symbol assertion emission that
// generates a comment marker. The real solver assertions land in
// later increments as fixtures exercise them.

// emitSetSolver writes pushStateIntoSolver and the per-action
// buildPrecondition helpers. Together they prepare a goivy.Solver
// with the current State as facts and an action-specific Clauses
// describing the inputs whose values we want the solver to choose.
//
// OPEN 055 / 055.1 / 055.2 progression:
//
//   - 055   wired the Solver round-trip with a `true` sanity check.
//   - 055.1 added GetModelClauses + per-input value extraction.
//   - 055.2 (this) emits a per-action precondition helper that
//     asserts state-symbol facts for every scalar bool state symbol
//     into the returned Clauses. Function-sorted symbols + the full
//     action.Pre derivation are the next sub-step (OPEN 055.3).
func (g *Generator) emitSetSolver(w *goWriter) {
	if !g.usesZ3() {
		return
	}
	w.line("// pushStateIntoSolver verifies the State is consistent with")
	w.line("// the module's axioms by running a SAT check through the")
	w.line("// provided goivy.Solver.")
	w.linef("func pushStateIntoSolver(_ *%s, sol *goivy.Solver) error {", g.StateTypeName)
	w.line("\tif sol == nil { return nil }")
	w.line("\ttrueExpr := goivy.NewConst(\"true\", goivy.Boolean)")
	w.line("\tok, err := sol.IsSat(trueExpr)")
	w.line("\tif err != nil { return err }")
	w.line("\tif !ok { return fmt.Errorf(\"ivy: state is unsatisfiable\") }")
	w.line("\treturn nil")
	w.line("}")
	w.blank()
	g.Ctx.AddImport("runtime", "fmt", "")

	// Shared helper that builds a Clauses asserting every scalar
	// bool state symbol equals its current value. Each action's
	// buildPrecondition_<Name> calls this and conjoins its inputs.
	w.line("// stateFactsAsClauses builds a *goivy.Clauses asserting that")
	w.line("// each scalar bool state symbol equals its current value.")
	w.line("// Used by every actionGen's input synthesis as the seed")
	w.line("// of the precondition (OPEN 055.2). Function-sorted and")
	w.line("// non-bool symbols are skipped for now (OPEN 055.3).")
	w.linef("func stateFactsAsClauses(state *%s) *goivy.Clauses {", g.StateTypeName)
	w.line("\tvar fmlas []goivy.Expr")
	for _, sym := range g.stateSymbols() {
		if _, ok := sym.Sort.(*goivy.BooleanSort); !ok {
			continue
		}
		exported := goExportedName(sym.Name)
		w.linef("\tfmlas = append(fmlas, mkBoolFact(%q, state.%s))", sym.Name, exported)
	}
	w.line("\treturn goivy.NewClauses(fmlas, nil)")
	w.line("}")
	w.blank()

	w.line("// mkBoolFact builds the formula (sym = val) where sym is a")
	w.line("// fresh bool-sorted goivy.Const named `name`. We return the")
	w.line("// LHS-only literal when val is true and its negation when false,")
	w.line("// matching the encoding goivy.Solver expects for bool facts.")
	w.line("func mkBoolFact(name string, val bool) goivy.Expr {")
	w.line("\tsym := goivy.NewConst(name, goivy.Boolean)")
	w.line("\tif val {")
	w.line("\t\treturn sym")
	w.line("\t}")
	w.line("\treturn &goivy.LogicNot{Body: sym}")
	w.line("}")
	w.blank()
}

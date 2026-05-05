package goivy

// RegisterFactories wires the proof package's factory functions into a
// module.Config, replacing the old init()-based global assignment.
func RegisterFactories(modCfg *Config, proofCfg *ProofConfig) {
	modCfg.ProofCfg = proofCfg
	modCfg.NewProofCheckerFn = func(mod *Module, axioms, definitions []*LabeledFormula, schemata *InsMap[string, *LabeledFormula]) ProofCheckerInterface {
		// Pass modCfg.AstCfg so the ProofChecker shares the module's LF counter,
		// matching Python's single global lf_counter.
		return NewProofChecker(proofCfg, mod, axioms, definitions, schemata, modCfg.AstCfg)
	}
	// GoalConcFn returns lg.Expr (the narrow flavor); use GoalConcExpr which
	// preserves the old "nil on non-Expr formulas" behavior. The compiler's
	// goalConcExpr (compiler/phase6.go:2207) expects lg.Expr, so this matches.
	modCfg.GoalConcFn = func(g *LabeledFormula) Expr {
		return GoalConcExpr(g)
	}
}

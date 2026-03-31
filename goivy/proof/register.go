package proof

import (
	"github.com/glycerine/ivy/goivy/ast"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
)

// RegisterFactories wires the proof package's factory functions into a
// module.Config, replacing the old init()-based global assignment.
func RegisterFactories(modCfg *module.Config, proofCfg *module.ProofConfig) {
	modCfg.ProofCfg = proofCfg
	modCfg.NewProofCheckerFn = func(mod *module.Module, axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) module.ProofCheckerInterface {
		// Pass modCfg.AstCfg so the ProofChecker shares the module's LF counter,
		// matching Python's single global lf_counter.
		return NewProofChecker(proofCfg, mod, axioms, definitions, schemata, modCfg.AstCfg)
	}
	modCfg.GoalConcFn = func(g *ast.LabeledFormula) lg.Expr {
		return GoalConc(g)
	}
}

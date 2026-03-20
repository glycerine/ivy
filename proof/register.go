package proof

import (
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// RegisterFactories wires the proof package's factory functions into a
// module.Config, replacing the old init()-based global assignment.
func RegisterFactories(modCfg *module.Config, proofCfg *Config) {
	modCfg.NewProofCheckerFn = func(axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) module.ProofCheckerInterface {
		return NewProofChecker(proofCfg, axioms, definitions, schemata)
	}
	modCfg.GoalConcFn = func(g *ast.LabeledFormula) lg.Expr {
		return GoalConc(g)
	}
}

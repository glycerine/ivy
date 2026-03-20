package proof

import (
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

func init() {
	module.NewProofCheckerFn = func(axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) module.ProofCheckerInterface {
		return NewProofChecker(axioms, definitions, schemata)
	}
	module.GoalConcFn = func(g *ast.LabeledFormula) lg.Expr {
		return GoalConc(g)
	}
}

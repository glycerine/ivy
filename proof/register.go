package proof

import (
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/compiler"
	lg "github.com/glycerine/goivy/logic"
)

func init() {
	compiler.NewProofCheckerFn = func(axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) compiler.ProofCheckerInterface {
		return NewProofChecker(axioms, definitions, schemata)
	}
	compiler.GoalConcFn = func(g *ast.LabeledFormula) lg.Expr {
		return GoalConc(g)
	}
}

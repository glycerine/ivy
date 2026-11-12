// Package-level aliases preserve the CUDD API while the remaining
// original modules are ported. Graph algorithms live only in core.
package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

var (
	Cudd_bddIte         = core.Cudd_bddIte
	Cudd_bddAnd         = core.Cudd_bddAnd
	Cudd_bddOr          = core.Cudd_bddOr
	Cudd_bddXor         = core.Cudd_bddXor
	Cudd_bddXnor        = core.Cudd_bddXnor
	Cudd_bddNand        = core.Cudd_bddNand
	Cudd_bddNor         = core.Cudd_bddNor
	Cudd_bddLeq         = core.Cudd_bddLeq
	Cudd_bddIteConstant = core.Cudd_bddIteConstant
	Cudd_bddIntersect   = core.Cudd_bddIntersect
	Cudd_bddAndLimit    = core.Cudd_bddAndLimit
)

var DD_NON_CONSTANT = core.DD_NON_CONSTANT
var cuddBddIteRecur = core.Cudd_bddIte
var cuddBddAndRecur = core.Cudd_bddAnd
var cuddBddXorRecur = core.Cudd_bddXor
var cuddBddIntersectRecur = core.Cudd_bddIntersect

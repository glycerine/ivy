// Package-level aliases preserve the CUDD API while the remaining
// original modules are ported. Graph algorithms live only in core.
package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

var (
	Cudd_bddExistAbstract    = core.Cudd_bddExistAbstract
	Cudd_bddUnivAbstract     = core.Cudd_bddUnivAbstract
	Cudd_bddXorExistAbstract = core.Cudd_bddXorExistAbstract
	Cudd_bddBooleanDiff      = core.Cudd_bddBooleanDiff
	Cudd_bddVarIsDependent   = core.Cudd_bddVarIsDependent
)

var cuddBddExistAbstractRecur = core.Cudd_bddExistAbstract
var cuddBddXorExistAbstractRecur = core.Cudd_bddXorExistAbstract

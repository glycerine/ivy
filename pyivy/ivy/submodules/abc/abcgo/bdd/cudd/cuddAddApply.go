// ADD apply algorithms are implemented only in core.
package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

var (
	Cudd_addApply          = core.Cudd_addApply
	Cudd_addMonadicApply   = core.Cudd_addMonadicApply
	Cudd_addPlus           = core.Cudd_addPlus
	Cudd_addTimes          = core.Cudd_addTimes
	Cudd_addDivide         = core.Cudd_addDivide
	Cudd_addMinus          = core.Cudd_addMinus
	Cudd_addMinimum        = core.Cudd_addMinimum
	Cudd_addMaximum        = core.Cudd_addMaximum
	Cudd_addThreshold      = core.Cudd_addThreshold
	Cudd_addSetNZ          = core.Cudd_addSetNZ
	Cudd_addOneZeroMaximum = core.Cudd_addOneZeroMaximum
	Cudd_addDiff           = core.Cudd_addDiff
	Cudd_addAgreement      = core.Cudd_addAgreement
	Cudd_addOr             = core.Cudd_addOr
	Cudd_addNand           = core.Cudd_addNand
	Cudd_addNor            = core.Cudd_addNor
	Cudd_addXor            = core.Cudd_addXor
	Cudd_addXnor           = core.Cudd_addXnor
	Cudd_addLog            = core.Cudd_addLog
)
var cuddAddApplyRecur = core.Cudd_addApply
var cuddAddMonadicApplyRecur = core.Cudd_addMonadicApply

// Original ADD ITE APIs delegate to the sole implementation in core.
package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

var (
	Cudd_addIte         = core.Cudd_addIte
	Cudd_addIteConstant = core.Cudd_addIteConstant
	Cudd_addEvalConst   = core.Cudd_addEvalConst
	Cudd_addCmpl        = core.Cudd_addCmpl
	Cudd_addLeq         = core.Cudd_addLeq
	cuddAddIteRecur     = core.Cudd_addIte
	cuddAddCmplRecur    = core.Cudd_addCmpl
)

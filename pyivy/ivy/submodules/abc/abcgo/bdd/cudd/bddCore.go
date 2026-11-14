// Package-level aliases preserve the CUDD API while the remaining
// original modules are ported. Graph algorithms live only in core.
package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

var (
	NewBDDManager     = core.NewBDDManager
	Cudd_Regular      = core.Cudd_Regular
	Cudd_IsComplement = core.Cudd_IsComplement
	Cudd_Not          = core.Cudd_Not
	Cudd_NotCond      = core.Cudd_NotCond
	Cudd_IsConstant   = core.Cudd_IsConstant
	Cudd_T            = core.Cudd_T
	Cudd_E            = core.Cudd_E
	Cudd_Support      = core.Cudd_Support
	Cudd_SupportSize  = core.Cudd_SupportSize
	Cudd_DagSize      = core.Cudd_DagSize
)

type DdNode = core.DdNode
type DdManager = core.DdManager

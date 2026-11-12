// Package-level aliases preserve the CUDD API while the remaining
// original modules are ported. Graph algorithms live only in core.
package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

var (
	BDDOperations = core.BDDOperations
	ParseFormula  = core.ParseFormula
)

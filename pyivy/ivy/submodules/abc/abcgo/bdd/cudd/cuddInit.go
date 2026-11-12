// Package-level aliases preserve the CUDD API while the remaining
// original modules are ported. Graph algorithms live only in core.
package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

var (
	Cudd_Init = core.Cudd_Init
	Cudd_Quit = core.Cudd_Quit
)

var cuddZddInitUniv = core.InitializeZDDUniverse
var cuddZddFreeUniv = core.FreeZDDUniverse

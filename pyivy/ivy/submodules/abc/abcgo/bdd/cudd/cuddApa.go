// Package-level aliases preserve the CUDD API while the remaining
// original modules are ported. Graph algorithms live only in core.
package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

var (
	Cudd_ApaNumberOfDigits   = core.Cudd_ApaNumberOfDigits
	Cudd_NewApaNumber        = core.Cudd_NewApaNumber
	Cudd_ApaCopy             = core.Cudd_ApaCopy
	Cudd_ApaAdd              = core.Cudd_ApaAdd
	Cudd_ApaSubtract         = core.Cudd_ApaSubtract
	Cudd_ApaShortDivision    = core.Cudd_ApaShortDivision
	Cudd_ApaIntDivision      = core.Cudd_ApaIntDivision
	Cudd_ApaShiftRight       = core.Cudd_ApaShiftRight
	Cudd_ApaSetToLiteral     = core.Cudd_ApaSetToLiteral
	Cudd_ApaPowerOfTwo       = core.Cudd_ApaPowerOfTwo
	Cudd_ApaCompare          = core.Cudd_ApaCompare
	Cudd_ApaCompareRatios    = core.Cudd_ApaCompareRatios
	Cudd_ApaPrintHex         = core.Cudd_ApaPrintHex
	Cudd_ApaPrintDecimal     = core.Cudd_ApaPrintDecimal
	Cudd_ApaPrintExponential = core.Cudd_ApaPrintExponential
	Cudd_ApaCountMinterm     = core.Cudd_ApaCountMinterm
	Cudd_ApaPrintMinterm     = core.Cudd_ApaPrintMinterm
	Cudd_ApaPrintMintermExp  = core.Cudd_ApaPrintMintermExp
	Cudd_ApaPrintDensity     = core.Cudd_ApaPrintDensity
)

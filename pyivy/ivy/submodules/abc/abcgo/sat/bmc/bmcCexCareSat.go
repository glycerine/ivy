package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// Bmc_CexCareSatBasedMinimize is bmcCexCare.c's: the SAT-based minimizer on
// the AIG's GIA copy (its result cannot be checked by ternary simulation).
func Bmc_CexCareSatBasedMinimize(p *aig.Graph, nRealPis int32, cex *util.Abc_Cex_t, highEffort, check, verbose int32) *util.Abc_Cex_t {
	return gia.Bmc_CexCareSatBasedMinimizeAig(gia.Gia_ManFromAigSimple(p), cex, highEffort, verbose)
}

// Bmc_CexCareVerify is bmcCexCare.c's: checks that the care bits of cexMin
// still falsify the output, and reports it.
func Bmc_CexCareVerify(p *aig.Graph, cex, cexMin *util.Abc_Cex_t, verbose int32) {
	g := gia.Gia_ManFromAigSimple(p)
	if verbose != 0 {
		fmt.Printf("Original :    ")
		Bmc_CexPrint(cex, gia.Gia_ManPiNum(g), 0)
		fmt.Printf("Minimized:    ")
		Bmc_CexPrint(cexMin, gia.Gia_ManPiNum(g), 0)
	}
	if Bmc_CexVerify(g, cex, cexMin) == 0 {
		fmt.Printf("Counter-example verification has failed.\n")
	} else {
		fmt.Printf("Counter-example verification succeeded.\n")
	}
}

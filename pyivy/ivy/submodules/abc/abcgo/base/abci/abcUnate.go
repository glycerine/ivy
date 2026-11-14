package abci

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab"
)

// Abc_NtkPrintUnate reports unate input variables for each network output.
func Abc_NtkPrintUnate(pNtk *Abc_Ntk_t, fUseBdds, fUseNaive, fVerbose int32) {
	if fUseBdds != 0 || fUseNaive != 0 {
		Abc_NtkPrintUnateBdd(pNtk, fUseNaive, fVerbose)
		return
	}
	Abc_NtkPrintUnateSat(pNtk, fVerbose)
}

// Abc_NtkPrintUnateBdd computes output unateness from shared global BDDs.
func Abc_NtkPrintUnateBdd(pNtk *Abc_Ntk_t, fUseNaive, fVerbose int32) {
	clk := time.Now()
	dd := Abc_NtkBuildGlobalBdds(pNtk, 10_000_000, 1, 1, 0, fVerbose)
	if dd == nil {
		return
	}
	clkBdd := time.Since(clk)
	fmt.Printf("Shared BDD size = %6d nodes.\n", int(core.Cudd_ReadKeys(dd)-core.Cudd_ReadDead(dd)))

	var totalSupps, totalUnate int
	if fUseNaive == 0 {
		core.Cudd_zddVarsFromBddVars(dd, 2)
	}
	for i := int32(0); i < abc.Abc_NtkCoNum(pNtk); i++ {
		output := abc.Abc_NtkCo(pNtk, i)
		function := Abc_ObjGlobalBdd(output)
		var info *extrab.Extra_UnateInfo_t
		if fUseNaive != 0 {
			info = extrab.Extra_UnateComputeSlow(dd, function)
		} else {
			info = extrab.Extra_UnateComputeFast(dd, function)
		}
		if info == nil {
			panic(fmt.Sprintf("Abc_NtkPrintUnateBdd: unateness computation failed for output %d", i))
		}
		if fVerbose != 0 {
			fmt.Printf("Out%4d : ", i)
			extrab.Extra_UnateInfoPrint(info)
		}
		totalSupps += info.NVars
		totalUnate += info.NUnate
		extrab.Extra_UnateInfoDissolve(info)
	}
	clkUnate := time.Since(clk) - clkBdd
	fmt.Printf("Ins/Outs = %4d/%4d.  Total supp = %5d.  Total unate = %5d.\n", abc.Abc_NtkCiNum(pNtk), abc.Abc_NtkCoNum(pNtk), totalSupps, totalUnate)
	fmt.Printf("Glob BDDs = %9.2f sec\n", clkBdd.Seconds())
	fmt.Printf("Unateness = %9.2f sec\n", clkUnate.Seconds())
	fmt.Printf("Total     = %9.2f sec\n", time.Since(clk).Seconds())
	Abc_NtkFreeGlobalBdds(pNtk, 1)
}

// Abc_NtkPrintUnateSat is intentionally empty, matching the C source stub.
func Abc_NtkPrintUnateSat(_ *Abc_Ntk_t, _ int32) {}

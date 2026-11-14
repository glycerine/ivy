package abci

import base_abc "github.com/glycerine/abcgo/base/abc"

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/bdd/cudd"
	"github.com/glycerine/abcgo/bdd/extrab"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
)

// Abc_NtkAutoPrint reports autosymmetry statistics for one output or all
// outputs. The BDD manager is extended with one copy of each input variable,
// as required by the space and equation computations.
func Abc_NtkAutoPrint(pNtk *Abc_Ntk_t, output, naive, verbose int32) {
	if Abc_NtkBuildGlobalBdds(pNtk, 10000000, 1, 1, 0, verbose) == nil {
		return
	}

	nInputs := Abc_NtkCiNum(pNtk)
	nOutputs := Abc_NtkCoNum(pNtk)
	dd := Abc_NtkGlobalBddMan(pNtk)
	if dd == nil {
		Abc_NtkFreeGlobalBdds(pNtk, 1)
		return
	}

	outputs := make([]*cudd.DdNode, nOutputs)
	for i := int32(0); i < nOutputs; i++ {
		outputs[i] = Abc_ObjGlobalBdd(base_abc.Abc_NtkCo(pNtk, i))
	}

	if verbose != 0 {
		fmt.Printf("Shared BDD size = %6d nodes.\n", cudd.Cudd_ReadKeys(dd)-cudd.Cudd_ReadDead(dd))
	}
	for i := int32(0); i < nInputs; i++ {
		cudd.Cudd_bddNewVar(dd)
	}
	assert(cudd.Cudd_ReadSize(dd) == 2*nInputs, "Cudd_ReadSize(dd) == 2*nInputs")
	cudd.Cudd_zddVarsFromBddVars(dd, 2)

	if output == -1 {
		Abc_NtkAutoPrintAll(dd, nInputs, outputs, nil, nil, naive)
	} else {
		Abc_NtkAutoPrintOne(dd, nInputs, outputs, output, nil, nil, naive)
	}

	Abc_NtkFreeGlobalBdds(pNtk, 1)
}

// Abc_NtkAutoPrintAll prints autosymmetry counts for every output and a
// cumulative summary. The matrix reconstruction check verifies that the ZDD
// equation cover describes the same space as the original BDD.
func Abc_NtkAutoPrintAll(dd *cudd.DdManager, nInputs int32, outputs []*cudd.DdNode, _ []string, _ []string, _ int32) {
	start := util.Abc_Clock()
	var autoOutputs, autoSum, autoMax, autoMaxSupport, maxSupport int32

	for output, function := range outputs {
		space := extrab.Extra_bddSpaceFromFunction(dd, function, function)
		cudd.Cudd_Ref(space)
		canonicalVars := extrab.Extra_bddSpaceCanonVars(dd, space)
		cudd.Cudd_Ref(canonicalVars)
		reduced := extrab.Extra_bddSpaceReduce(dd, function, canonicalVars)
		cudd.Cudd_Ref(reduced)
		equations := extrab.Extra_bddSpaceEquations(dd, space)
		cudd.Cudd_Ref(equations)

		support := int32(len(dd.Support(space)))
		minterms := cudd.Cudd_CountMinterm(dd, space, support)
		autos := extra.Extra_Base2LogDouble(minterms)
		fmt.Printf("Output #%3d: Inputs = %2d. AutoK = %2d.\n", output, support, autos)

		if autos > 0 {
			autoOutputs++
			autoSum += autos
			if autoMax < autos {
				autoMax = autos
				autoMaxSupport = support
			}
		}
		if maxSupport < support {
			maxSupport = support
		}

		matrixSpace := extrab.Extra_bddSpaceFromMatrixPos(dd, equations)
		cudd.Cudd_Ref(matrixSpace)
		if space != matrixSpace {
			fmt.Println("Spaces are NOT EQUAL!")
		}
		cudd.Cudd_RecursiveDeref(dd, space)
		cudd.Cudd_RecursiveDeref(dd, matrixSpace)
		cudd.Cudd_RecursiveDeref(dd, canonicalVars)
		cudd.Cudd_RecursiveDeref(dd, reduced)
		cudd.Cudd_RecursiveDerefZdd(dd, equations)
	}

	fmt.Println("The cumulative statistics for all outputs:")
	fmt.Printf("Ins=%3d ", nInputs)
	fmt.Printf("InMax=%3d   ", maxSupport)
	fmt.Printf("Outs=%3d ", len(outputs))
	fmt.Printf("Auto=%3d   ", autoOutputs)
	fmt.Printf("SumK=%3d ", autoSum)
	fmt.Printf("KMax=%2d ", autoMax)
	fmt.Printf("Supp=%3d   ", autoMaxSupport)
	fmt.Printf("Time=%4.2f ", time.Duration(util.Abc_Clock()-start).Seconds())
	fmt.Println()
}

// Abc_NtkAutoPrintOne prints the autosymmetry count for one output.
func Abc_NtkAutoPrintOne(dd *cudd.DdManager, _ int32, outputs []*cudd.DdNode, output int32, _ []string, _ []string, _ int32) {
	function := outputs[output]
	space := extrab.Extra_bddSpaceFromFunctionFast(dd, function)
	cudd.Cudd_Ref(space)
	canonicalVars := extrab.Extra_bddSpaceCanonVars(dd, space)
	cudd.Cudd_Ref(canonicalVars)
	reduced := extrab.Extra_bddSpaceReduce(dd, function, canonicalVars)
	cudd.Cudd_Ref(reduced)
	equations := extrab.Extra_bddSpaceEquations(dd, space)
	cudd.Cudd_Ref(equations)

	support := int32(len(dd.Support(space)))
	minterms := cudd.Cudd_CountMinterm(dd, space, support)
	autos := extra.Extra_Base2LogDouble(minterms)
	fmt.Printf("Output #%3d: Inputs = %2d. AutoK = %2d.\n", output, support, autos)

	cudd.Cudd_RecursiveDeref(dd, space)
	cudd.Cudd_RecursiveDeref(dd, canonicalVars)
	cudd.Cudd_RecursiveDeref(dd, reduced)
	cudd.Cudd_RecursiveDerefZdd(dd, equations)
}

package abci

import (
	"fmt"

	ifpkg "github.com/glycerine/abcgo/map/if"
)

/**CFile****************************************************************

  FileName    [abcIf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Interface with the FPGA mapping package.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: abcIf.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "map/if/if.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Abc_NtkFunc_sTruth = [3]int32{0xAA, 0xCC, 0xF0}

func Abc_NtkFuncCof0(t, v int32) int32 {
	mask := spStatic_Abc_NtkFunc_sTruth[v]
	return 0xff & ((t & ^mask) | ((t & ^mask) << (1 << v)))
}
func Abc_NtkFuncCof1(t, v int32) int32 {
	mask := spStatic_Abc_NtkFunc_sTruth[v]
	return 0xff & ((t & mask) | ((t & mask) >> (1 << v)))
}
func Abc_NtkFuncHasVar(t, v int32) int32 {
	mask := spStatic_Abc_NtkFunc_sTruth[v]
	if ((t & mask) >> (1 << v)) != (t & ^mask) {
		return 1
	}
	return 0
}
func Abc_NtkFuncSuppSize(t int32) int32 {
	return Abc_NtkFuncHasVar(t, 0) + Abc_NtkFuncHasVar(t, 1) + Abc_NtkFuncHasVar(t, 2)
}

// Abc_NtkCutCostMuxPrecompute prints the source-derived truth table for review.
func Abc_NtkCutCostMuxPrecompute() {
	counter := int32(0)
	for i := int32(0); i < 256; i++ {
		value := int32(0)
		if Abc_NtkFuncSuppSize(i) < 3 {
			value = 1
		} else {
			for v := int32(0); v < 3; v++ {
				if Abc_NtkFuncSuppSize(Abc_NtkFuncCof0(i, v)) == 1 && Abc_NtkFuncSuppSize(Abc_NtkFuncCof1(i, v)) == 1 {
					value = 1
					counter++
					break
				}
			}
		}
		fmt.Printf("%d, // %3d  0x%02X\n", value, i, i)
	}
	fmt.Printf("Total number of MUXes = %d.\n", counter)
}

var abcNtkCutCostMuxLookup = [256]uint8{1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1}

// Abc_NtkCutCostMux returns the built source's MUX-count estimate.
func Abc_NtkCutCostMux(p *If_Man_t, pCut *If_Cut_t) int32 {
	leaves := ifpkg.If_CutLeaveNum(pCut)
	if leaves < 3 {
		return 1
	}
	if leaves == 3 && abcNtkCutCostMuxLookup[byte(*ifpkg.If_CutTruth(p, pCut))] != 0 {
		return 1
	}
	return (1 << leaves) - 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

package lpk

import (
	"fmt"
	kit "github.com/glycerine/abcgo/bool/kit"
	ifpkg "github.com/glycerine/abcgo/map/if"
	vec "github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [lpkMux.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkMux.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "lpkInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Find the best cofactoring variable.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_MapTreeBestCofVar(p *Lpk_Man_t, pTruth []uint32, nVars int32, pCof0 []uint32, pCof1 []uint32) int32 {
	var i, iBestVar, nSuppSizeCur0, nSuppSizeCur1, nSuppSizeCur, nSuppSizeMin int32
	// iterate through variables
	iBestVar = -1
	nSuppSizeMin = kit.KIT_INFINITY
	for i = 0; i < nVars; i++ {
		// cofactor the functiona and get support sizes
		kit.Kit_TruthCofactor0New(pCof0, pTruth, nVars, i)
		kit.Kit_TruthCofactor1New(pCof1, pTruth, nVars, i)
		nSuppSizeCur0 = kit.Kit_TruthSupportSize(pCof0, nVars)
		nSuppSizeCur1 = kit.Kit_TruthSupportSize(pCof1, nVars)
		nSuppSizeCur = nSuppSizeCur0 + nSuppSizeCur1
		// skip cofactoring that goes above the limit
		if nSuppSizeCur0 > p.pPars.nLutSize || nSuppSizeCur1 > p.pPars.nLutSize {
			continue
		}
		// compare this variable with other variables
		if nSuppSizeMin > nSuppSizeCur {
			nSuppSizeMin = nSuppSizeCur
			iBestVar = i
		}
	}
	// cofactor w.r.t. this variable
	if iBestVar != -1 {
		kit.Kit_TruthCofactor0New(pCof0, pTruth, nVars, iBestVar)
		kit.Kit_TruthCofactor1New(pCof1, pTruth, nVars, iBestVar)
	}
	return iBestVar
}

/*
*Function*************************************************************

	Synopsis    [Maps the function by the best cofactoring.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_MapTreeMux_rec(p *Lpk_Man_t, pTruth []uint32, nVars int32, ppLeaves []*If_Obj_t) *If_Obj_t {
	var pCof0 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 0).([]uint32)
	var pCof1 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 1).([]uint32)
	var pObj0, pObj1 *If_Obj_t
	var ppNtks [2]*Kit_DsdNtk_t
	var iBestVar int32
	assert(nVars > 3, "nVars > 3")
	p.fCalledOnce = 1
	// cofactor w.r.t. the best variable
	iBestVar = Lpk_MapTreeBestCofVar(p, pTruth, nVars, pCof0, pCof1)
	if iBestVar == -1 {
		return nil
	}
	// decompose the functions
	ppNtks[0] = kit.Kit_DsdDecompose(pCof0, nVars)
	ppNtks[1] = kit.Kit_DsdDecompose(pCof1, nVars)
	if p.pPars.fVeryVerbose != 0 {
		fmt.Printf("Cofactoring w.r.t. var %c (%d -> %d+%d supp vars):\n", 'a'+iBestVar, nVars, kit.Kit_TruthSupportSize(pCof0, nVars), kit.Kit_TruthSupportSize(pCof1, nVars))
		kit.Kit_DsdPrintExpanded(ppNtks[0])
		kit.Kit_DsdPrintExpanded(ppNtks[1])
	}
	// map the DSD structures
	pObj0 = Lpk_MapTree_rec(p, ppNtks[0], ppLeaves[:], int32(ppNtks[0].Root), nil)
	pObj1 = Lpk_MapTree_rec(p, ppNtks[1], ppLeaves[:], int32(ppNtks[1].Root), nil)
	kit.Kit_DsdNtkFree(ppNtks[0])
	kit.Kit_DsdNtkFree(ppNtks[1])
	return ifpkg.If_ManCreateMux(p.pIfMan, pObj0, pObj1, ppLeaves[iBestVar])
}

/*
*Function*************************************************************

	Synopsis    [Implements support-reducing decomposition.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_MapSuppRedDec_rec(p *Lpk_Man_t, pTruth []uint32, nVars int32, ppLeaves []*If_Obj_t) *If_Obj_t {
	var pNtkDec *Kit_DsdNtk_t
	var pNtkComp *Kit_DsdNtk_t
	var ppNtks [2]*Kit_DsdNtk_t
	var pTemp *Kit_DsdNtk_t
	var pObjNew *If_Obj_t
	var pCof0 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 0).([]uint32)
	var pCof1 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 1).([]uint32)
	var pDec0 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 2).([]uint32)
	var pDec1 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 3).([]uint32)
	var pDec []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 4).([]uint32)
	var pCo00 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 5).([]uint32)
	var pCo01 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 6).([]uint32)
	var pCo10 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 7).([]uint32)
	var pCo11 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 8).([]uint32)
	var pCo0 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 9).([]uint32)
	var pCo1 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 10).([]uint32)
	var pCo []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 11).([]uint32)
	var TrueMint0, TrueMint1, FalseMint0, FalseMint1 int32
	var uSubsets, uSubset0, uSubset1 uint32
	var iVar, iVarReused, i int32

	// determine if supp-red decomposition exists
	uSubsets = Lpk_MapSuppRedDecSelect(p, pTruth, nVars, &iVar, &iVarReused)
	if uSubsets == 0 {
		return nil
	}
	p.nCalledSRed++

	// get the cofactors
	kit.Kit_TruthCofactor0New(pCof0, pTruth, nVars, iVar)
	kit.Kit_TruthCofactor1New(pCof1, pTruth, nVars, iVar)

	// get the bound sets
	uSubset0 = uSubsets & 0xFFFF
	uSubset1 = uSubsets >> 16

	// compute the decomposed functions
	ppNtks[0] = kit.Kit_DsdDecompose(pCof0, nVars)
	ppNtks[1] = kit.Kit_DsdDecompose(pCof1, nVars)
	pTemp = ppNtks[0]
	ppNtks[0] = kit.Kit_DsdExpand(pTemp)
	kit.Kit_DsdNtkFree(pTemp)
	pTemp = ppNtks[1]
	ppNtks[1] = kit.Kit_DsdExpand(pTemp)
	kit.Kit_DsdNtkFree(pTemp)
	kit.Kit_DsdTruthPartial(p.pDsdMan, ppNtks[0], pDec0, uSubset0)
	kit.Kit_DsdTruthPartial(p.pDsdMan, ppNtks[1], pDec1, uSubset1)
	//    Kit_DsdTruthPartialTwo( p->pDsdMan, ppNtks[0], uSubset0, iVarReused, pCo0, pDec0 );
	//    Kit_DsdTruthPartialTwo( p->pDsdMan, ppNtks[1], uSubset1, iVarReused, pCo1, pDec1 );
	kit.Kit_DsdNtkFree(ppNtks[0])
	kit.Kit_DsdNtkFree(ppNtks[1])
	//Kit_DsdPrintFromTruth( pDec0, nVars );
	//Kit_DsdPrintFromTruth( pDec1, nVars );
	// get the decomposed function
	kit.Kit_TruthMuxVar(pDec, pDec0, pDec1, nVars, iVar)

	// find any true assignments of the decomposed functions
	TrueMint0 = kit.Kit_TruthFindFirstBit(pDec0, nVars)
	TrueMint1 = kit.Kit_TruthFindFirstBit(pDec1, nVars)
	assert(TrueMint0 >= 0 && TrueMint1 >= 0, "TrueMint0 >= 0 && TrueMint1 >= 0")
	// find any false assignments of the decomposed functions
	FalseMint0 = kit.Kit_TruthFindFirstZero(pDec0, nVars)
	FalseMint1 = kit.Kit_TruthFindFirstZero(pDec1, nVars)
	assert(FalseMint0 >= 0 && FalseMint1 >= 0, "FalseMint0 >= 0 && FalseMint1 >= 0")

	// cofactor the cofactors according to these minterms
	kit.Kit_TruthCopy(pCo00, pCof0, nVars)
	kit.Kit_TruthCopy(pCo01, pCof0, nVars)
	for i = 0; i < nVars; i++ {
		if (uSubset0 & (1 << i)) != 0 {
			if (FalseMint0 & (1 << i)) != 0 {
				kit.Kit_TruthCofactor1(pCo00, nVars, i)
			} else {
				kit.Kit_TruthCofactor0(pCo00, nVars, i)
			}
			if (TrueMint0 & (1 << i)) != 0 {
				kit.Kit_TruthCofactor1(pCo01, nVars, i)
			} else {
				kit.Kit_TruthCofactor0(pCo01, nVars, i)
			}
		}
	}
	kit.Kit_TruthCopy(pCo10, pCof1, nVars)
	kit.Kit_TruthCopy(pCo11, pCof1, nVars)
	for i = 0; i < nVars; i++ {
		if (uSubset1 & (1 << i)) != 0 {
			if (FalseMint1 & (1 << i)) != 0 {
				kit.Kit_TruthCofactor1(pCo10, nVars, i)
			} else {
				kit.Kit_TruthCofactor0(pCo10, nVars, i)
			}
			if (TrueMint1 & (1 << i)) != 0 {
				kit.Kit_TruthCofactor1(pCo11, nVars, i)
			} else {
				kit.Kit_TruthCofactor0(pCo11, nVars, i)
			}
		}
	}

	// derive the functions by composing them with the new variable (iVarReused)
	kit.Kit_TruthMuxVar(pCo0, pCo00, pCo01, nVars, iVarReused)
	kit.Kit_TruthMuxVar(pCo1, pCo10, pCo11, nVars, iVarReused)
	//Kit_DsdPrintFromTruth( pCo0, nVars );
	//Kit_DsdPrintFromTruth( pCo1, nVars );

	// derive the composition function
	kit.Kit_TruthMuxVar(pCo, pCo0, pCo1, nVars, iVar)

	// process the decomposed function
	pNtkDec = kit.Kit_DsdDecompose(pDec, nVars)
	pNtkComp = kit.Kit_DsdDecompose(pCo, nVars)
	//Kit_DsdPrint( stdout, pNtkDec );
	//Kit_DsdPrint( stdout, pNtkComp );
	//printf( "cofactored variable %c\n", 'a' + iVar );
	//printf( "reused variable %c\n", 'a' + iVarReused );

	ppLeaves[iVarReused] = Lpk_MapTree_rec(p, pNtkDec, ppLeaves[:], int32(pNtkDec.Root), nil)
	pObjNew = Lpk_MapTree_rec(p, pNtkComp, ppLeaves[:], int32(pNtkComp.Root), nil)

	kit.Kit_DsdNtkFree(pNtkDec)
	kit.Kit_DsdNtkFree(pNtkComp)
	return pObjNew
}

/*
*Function*************************************************************

	Synopsis    [Evaluates 4-LUT decompositions of larger functions.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_TryDec(pTruth []word, nVars int32) {

	//int nWords = Abc_Truth6WordNum( nVars );
	assert(nVars == 8, "nVars == 8")
	fmt.Printf("%d", ifpkg.If_Dec6Perform(pTruth[0], 1) > 0)
	fmt.Printf("%d", ifpkg.If_Dec6Perform(pTruth[1], 1) > 0)
	fmt.Printf("%d", ifpkg.If_Dec6Perform(pTruth[2], 1) > 0)
	fmt.Printf("%d", ifpkg.If_Dec6Perform(pTruth[3], 1) > 0)
	fmt.Printf("\n")
}
func Lpk_TryDecTest() {
	var t0 word = word(0x4f1ead396f247a04)
	var t1 word = word(0x10bdb210c006eab5)
	var t2 word = word(0x68ab4bfa8acb7a13)
	var t3 word = word(0xb14ede67096c6eed)
	Truth := []word{t0, t1, t2, t3}
	Lpk_TryDec(Truth, 8)

}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

package lpk

import (
	"fmt"
	kit "github.com/glycerine/abcgo/bool/kit"
	util "github.com/glycerine/abcgo/misc/util"
	vec "github.com/glycerine/abcgo/misc/vec"
	"os"
)

/**CFile****************************************************************

  FileName    [lpkSets.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkSets.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "lpkInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Lpk_Set_t struct {
	iVar     byte   // the cofactoring variable
	Over     byte   // the overlap in supports
	SRed     byte   // the support reduction
	Size     byte   // the size of the boundset
	uSubset0 uint32 // the first subset (with removed)
	uSubset1 uint32 // the second subset (with removed)
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Recursively computes decomposable subsets.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_ComputeSets_rec(p *Kit_DsdNtk_t, iLit int32, vSets *Vec_Int_t) uint32 {
	var i, uSupport, uSuppCur uint32
	var iLitFanin int32
	var pObj *Kit_DsdObj_t
	// consider the case of simple gate
	pObj = kit.Kit_DsdNtkObj(p, util.Abc_Lit2Var(iLit))
	if pObj == nil {
		return (1 << util.Abc_Lit2Var(iLit))
	}
	if pObj.Type == kit.KIT_DSD_AND || pObj.Type == kit.KIT_DSD_XOR {
		var uSupps [16]uint32
		var Limit uint32
		var s uint32
		uSupport = 0
		// inlining of macro Kit_DsdObjForEachFanin
		for i = 0; i < uint32(kit.Kit_DsdObjFanNum(pObj)); i++ {
			iLitFanin = int32(kit.Kit_DsdObjFans(pObj)[i])
			uSupps[i] = Lpk_ComputeSets_rec(p, iLitFanin, vSets)
			uSupport |= uSupps[i]
		}
		// create all subsets, except empty and full
		Limit = (1 << uint32(kit.Kit_DsdObjFanNum(pObj))) - 1
		for s = 1; s < Limit; s++ {
			uSuppCur = 0
			for i = 0; i < uint32(kit.Kit_DsdObjFanNum(pObj)); i++ {
				if (s & (1 << i)) != 0 {
					uSuppCur |= uSupps[i]
				}
			}
			vec.Vec_IntPush(vSets, int32(uSuppCur))
		}
		return uSupport
	}
	assert(pObj.Type == kit.KIT_DSD_PRIME, "pObj.Type == kit.KIT_DSD_PRIME")
	// get the cumulative support of all fanins
	uSupport = 0
	// inlining of macro Kit_DsdObjForEachFanin
	for i = 0; i < uint32(kit.Kit_DsdObjFanNum(pObj)); i++ {
		iLitFanin = int32(kit.Kit_DsdObjFans(pObj)[i])
		uSuppCur = Lpk_ComputeSets_rec(p, iLitFanin, vSets)
		uSupport |= uSuppCur
		vec.Vec_IntPush(vSets, int32(uSuppCur))
	}
	return uSupport
}

/*
*Function*************************************************************

	Synopsis    [Computes the set of subsets of decomposable variables.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_ComputeSets(p *Kit_DsdNtk_t, vSets *Vec_Int_t) uint32 {
	var uSupport, Entry uint32
	var Number, i int32
	assert(kit.Kit_DsdNtkVarNum(p) <= 16, "p.nVars <= 16")
	vec.Vec_IntClear(vSets)
	vec.Vec_IntPush(vSets, 0)
	if kit.Kit_DsdNtkRoot(p).Type == kit.KIT_DSD_CONST1 {
		return 0
	}
	if kit.Kit_DsdNtkRoot(p).Type == kit.KIT_DSD_VAR {
		uSupport = (1 << util.Abc_Lit2Var(int32(kit.Kit_DsdObjFans(kit.Kit_DsdNtkRoot(p))[0])))
		vec.Vec_IntPush(vSets, int32(uSupport))
		return uSupport
	}
	uSupport = Lpk_ComputeSets_rec(p, int32(p.Root), vSets)
	assert((uSupport&0xFFFF0000) == 0, "(uSupport & 0xFFFF0000) == 0")
	vec.Vec_IntPush(vSets, int32(uSupport))
	// set the remaining variables
	for i = 0; i < int32(vec.Vec_IntSize(vSets)); i++ {
		Number = vec.Vec_IntEntry(vSets, int(i))
		Entry = uint32(Number)
		vec.Vec_IntWriteEntry(vSets, int(i), int32(Entry|((uSupport&^Entry)<<16)))
	}
	return uSupport
}

/*
*Function*************************************************************

	Synopsis    [Prints the sets of subsets.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_PrintSetOne(uSupport int32) {
	var k uint32
	for k = 0; k < 16; k++ {
		if (uSupport & (1 << k)) != 0 {
			fmt.Printf("%c", 'a'+k)
		}
	}
	fmt.Printf(" ")
}

/*
*Function*************************************************************

	Synopsis    [Prints the sets of subsets.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_PrintSets(vSets *Vec_Int_t) {
	var uSupport uint32
	var Number, i int32
	fmt.Printf("Subsets(%d): ", vec.Vec_IntSize(vSets))
	for i = 0; i < int32(vec.Vec_IntSize(vSets)); i++ {
		Number = vec.Vec_IntEntry(vSets, int(i))
		uSupport = uint32(Number)
		Lpk_PrintSetOne(int32(uSupport))
	}
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Computes maximal support reducing bound-sets.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Lpk_ComposeSets_nTravId int32
var spStatic_Lpk_ComposeSets_TravId [1 << 16]int32
var spStatic_Lpk_ComposeSets_SRed [1 << 16]byte
var spStatic_Lpk_ComposeSets_Over [1 << 16]byte
var spStatic_Lpk_ComposeSets_Parents [1 << 16]uint32

func Lpk_ComposeSets(vSets0 *Vec_Int_t, vSets1 *Vec_Int_t, nVars int32, iCofVar int32, pStore []Lpk_Set_t, pSize *int32, nSizeLimit int32) {
	// the number of the times this is visited
	// last visited
	// best support reduction
	// best overlaps
	// best set of parents
	var Used [1 << 16]uint16 // storage for used subsets
	var nSuppSize, nSuppOver, nSuppRed, nUsed, nMinOver, i, k, s int32
	var Entry, Entry0, Entry1 uint32
	var uSupp, uSupp0, uSupp1, uSuppTotal uint32
	var pEntry *Lpk_Set_t

	if spStatic_Lpk_ComposeSets_nTravId == (1 << 30) {
		clear(spStatic_Lpk_ComposeSets_TravId[:])
	}

	// collect support reducing subsets
	nUsed = 0
	spStatic_Lpk_ComposeSets_nTravId++
	uSuppTotal = kit.Kit_BitMask(nVars) & ^(1 << iCofVar)
	for i = 0; i < int32(vec.Vec_IntSize(vSets0)); i++ {
		Entry0 = uint32(vec.Vec_IntEntry(vSets0, int(i)))
		for k = 0; k < int32(vec.Vec_IntSize(vSets1)); k++ {
			Entry1 = uint32(vec.Vec_IntEntry(vSets1, int(k)))
			uSupp0 = (Entry0 & 0xFFFF)
			uSupp1 = (Entry1 & 0xFFFF)
			// skip trivial
			if uSupp0 == 0 || uSupp1 == 0 || (uSupp0|uSupp1) == uSuppTotal {
				continue
			}
			if kit.Kit_WordHasOneBit(uSupp0) != 0 && kit.Kit_WordHasOneBit(uSupp1) != 0 {
				continue
			}
			// get the entry
			Entry = Entry0 | Entry1
			uSupp = Entry & 0xFFFF
			// set the bound set size
			nSuppSize = kit.Kit_WordCountOnes(uSupp)
			// get the number of overlapping vars
			nSuppOver = kit.Kit_WordCountOnes(Entry & (Entry >> 16))
			// get the support reduction
			nSuppRed = nSuppSize - 1 - nSuppOver
			// only consider support-reducing subsets
			if nSuppRed <= 0 {
				continue
			}
			// check if this support is already used
			if spStatic_Lpk_ComposeSets_TravId[uSupp] < spStatic_Lpk_ComposeSets_nTravId {
				Used[nUsed] = uint16(uSupp)
				nUsed++

				spStatic_Lpk_ComposeSets_TravId[uSupp] = spStatic_Lpk_ComposeSets_nTravId
				spStatic_Lpk_ComposeSets_SRed[uSupp] = byte(nSuppRed)
				spStatic_Lpk_ComposeSets_Over[uSupp] = byte(nSuppOver)
				spStatic_Lpk_ComposeSets_Parents[uSupp] = uint32((k << 16) | i)

			} else if spStatic_Lpk_ComposeSets_TravId[uSupp] == spStatic_Lpk_ComposeSets_nTravId && spStatic_Lpk_ComposeSets_SRed[uSupp] < byte(nSuppRed) {
				spStatic_Lpk_ComposeSets_TravId[uSupp] = spStatic_Lpk_ComposeSets_nTravId
				spStatic_Lpk_ComposeSets_SRed[uSupp] = byte(nSuppRed)
				spStatic_Lpk_ComposeSets_Over[uSupp] = byte(nSuppOver)
				spStatic_Lpk_ComposeSets_Parents[uSupp] = uint32((k << 16) | i)
			}
		}
	}

	// find the minimum overlap
	nMinOver = 1000
	for s = 0; s < nUsed; s++ {
		if nMinOver > int32(spStatic_Lpk_ComposeSets_Over[Used[s]]) {
			nMinOver = int32(spStatic_Lpk_ComposeSets_Over[Used[s]])
		}
	}

	// collect the accumulated ones
	for s = 0; s < nUsed; s++ {
		if spStatic_Lpk_ComposeSets_Over[Used[s]] == byte(nMinOver) {
			// save the entry
			if *pSize == nSizeLimit {
				return
			}
			pEntry = &pStore[int(*pSize)]
			*pSize++

			i = int32(spStatic_Lpk_ComposeSets_Parents[Used[s]] & 0xFFFF)
			k = int32(spStatic_Lpk_ComposeSets_Parents[Used[s]] >> 16)

			pEntry.uSubset0 = uint32(vec.Vec_IntEntry(vSets0, int(i)))
			pEntry.uSubset1 = uint32(vec.Vec_IntEntry(vSets1, int(k)))
			Entry = pEntry.uSubset0 | pEntry.uSubset1

			// record the cofactoring variable
			pEntry.iVar = byte(iCofVar)
			// set the bound set size
			pEntry.Size = byte(kit.Kit_WordCountOnes(Entry & 0xFFFF))
			// get the number of overlapping vars
			pEntry.Over = byte(kit.Kit_WordCountOnes(Entry & (Entry >> 16)))
			// get the support reduction
			pEntry.SRed = pEntry.Size - 1 - pEntry.Over
			assert(pEntry.SRed > 0, "pEntry.SRed > 0")
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Prints one set.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_MapSuppPrintSet(pSet *Lpk_Set_t, i int32) {
	var Entry uint32
	Entry = pSet.uSubset0 | pSet.uSubset1
	fmt.Printf("%2d : ", i)
	fmt.Printf("Var = %c  ", 'a'+pSet.iVar)
	fmt.Printf("Size = %2d  ", pSet.Size)
	fmt.Printf("Over = %2d  ", pSet.Over)
	fmt.Printf("SRed = %2d  ", pSet.SRed)
	Lpk_PrintSetOne(int32(Entry))
	fmt.Printf("              ")
	Lpk_PrintSetOne(int32(Entry >> 16))
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Evaluates the cofactors.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Lpk_MapSuppRedDecSelect_nStoreSize int32 = 256
var spStatic_Lpk_MapSuppRedDecSelect_pStore [256]Lpk_Set_t
var spStatic_Lpk_MapSuppRedDecSelect_pSet *Lpk_Set_t
var spStatic_Lpk_MapSuppRedDecSelect_pSetBest *Lpk_Set_t

func Lpk_MapSuppRedDecSelect(p *Lpk_Man_t, pTruth []uint32, nVars int32, piVar *int32, piVarReused *int32) uint32 {

	var ppNtks [2]*Kit_DsdNtk_t
	var pTemp *Kit_DsdNtk_t
	var vSets0 *Vec_Int_t = p.vSets[0]
	var vSets1 *Vec_Int_t = p.vSets[1]
	var pCof0 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 0).([]uint32)
	var pCof1 []uint32 = vec.Vec_PtrEntry(p.vTtNodes, 1).([]uint32)
	var nSets, i, SizeMax int32
	var bestIndex int32 //, SRedMax;
	var Entry uint32
	var fVerbose int32 = p.pPars.fVeryVerbose
	//    int fVerbose = 0;

	// collect decomposable subsets for each pair of cofactors
	if fVerbose != 0 {
		fmt.Printf("\nExploring support-reducing bound-sets of function:\n")
		kit.Kit_DsdPrintFromTruth(pTruth, nVars)
	}
	nSets = 0
	for i = 0; i < nVars; i++ {
		if fVerbose != 0 {
			fmt.Printf("Evaluating variable %c:\n", 'a'+i)
		}
		// evaluate the cofactor pair
		kit.Kit_TruthCofactor0New(pCof0, pTruth, nVars, i)
		kit.Kit_TruthCofactor1New(pCof1, pTruth, nVars, i)
		// decompose and expand
		ppNtks[0] = kit.Kit_DsdDecompose(pCof0, nVars)
		ppNtks[1] = kit.Kit_DsdDecompose(pCof1, nVars)
		pTemp = ppNtks[0]
		ppNtks[0] = kit.Kit_DsdExpand(pTemp)
		kit.Kit_DsdNtkFree(pTemp)
		pTemp = ppNtks[1]
		ppNtks[1] = kit.Kit_DsdExpand(pTemp)
		kit.Kit_DsdNtkFree(pTemp)
		if fVerbose != 0 {
			kit.Kit_DsdPrint(os.Stdout, ppNtks[0])
		}
		if fVerbose != 0 {
			kit.Kit_DsdPrint(os.Stdout, ppNtks[1])
		}
		// compute subsets
		Lpk_ComputeSets(ppNtks[0], vSets0)
		Lpk_ComputeSets(ppNtks[1], vSets1)
		// print subsets
		if fVerbose != 0 {
			Lpk_PrintSets(vSets0)
		}
		if fVerbose != 0 {
			Lpk_PrintSets(vSets1)
		}
		// free the networks
		kit.Kit_DsdNtkFree(ppNtks[0])
		kit.Kit_DsdNtkFree(ppNtks[1])
		// evaluate the pair
		Lpk_ComposeSets(vSets0, vSets1, nVars, i, spStatic_Lpk_MapSuppRedDecSelect_pStore[:], &nSets, spStatic_Lpk_MapSuppRedDecSelect_nStoreSize)
	}

	// print the results
	if fVerbose != 0 {
		fmt.Printf("\n")
	}
	if fVerbose != 0 {
		for i = 0; i < nSets; i++ {
			Lpk_MapSuppPrintSet(&spStatic_Lpk_MapSuppRedDecSelect_pStore[i], i)
		}
	}

	// choose the best subset
	SizeMax = 0
	spStatic_Lpk_MapSuppRedDecSelect_pSetBest = nil
	for i = 0; i < nSets; i++ {
		spStatic_Lpk_MapSuppRedDecSelect_pSet = &spStatic_Lpk_MapSuppRedDecSelect_pStore[i]
		if int32(spStatic_Lpk_MapSuppRedDecSelect_pSet.Size) > p.pPars.nLutSize-1 {
			continue
		}
		if SizeMax < int32(spStatic_Lpk_MapSuppRedDecSelect_pSet.Size) {
			spStatic_Lpk_MapSuppRedDecSelect_pSetBest = spStatic_Lpk_MapSuppRedDecSelect_pSet
			bestIndex = i
			SizeMax = int32(spStatic_Lpk_MapSuppRedDecSelect_pSet.Size)
		}
	}
	/*
	   // if the best is not choosen, select the one with largest reduction
	   SRedMax = 0;
	   if ( pSetBest == NULL )
	   {
	       for ( i = 0; i < nSets; i++ )
	       {
	           pSet = &pStore[i];
	           if ( SRedMax < pSet->SRed )
	           {
	               pSetBest = pSet;
	               SRedMax = pSet->SRed;
	           }
	       }
	   }
	*/
	if spStatic_Lpk_MapSuppRedDecSelect_pSetBest == nil {
		if fVerbose != 0 {
			fmt.Printf("Could not select a subset.\n")
		}
		return 0

	} else {
		if fVerbose != 0 {
			fmt.Printf("Selected the following subset:\n")
		}
		if fVerbose != 0 {
			Lpk_MapSuppPrintSet(spStatic_Lpk_MapSuppRedDecSelect_pSetBest, bestIndex)
		}
	}

	// prepare the return result
	// get the remaining variables
	Entry = ((spStatic_Lpk_MapSuppRedDecSelect_pSetBest.uSubset0 >> 16) | (spStatic_Lpk_MapSuppRedDecSelect_pSetBest.uSubset1 >> 16))
	// get the variables to be removed
	Entry = kit.Kit_BitMask(nVars) & ^(1 << spStatic_Lpk_MapSuppRedDecSelect_pSetBest.iVar) & ^Entry
	// make sure there are some - otherwise it is not supp-red
	assert((Entry != 0), "Entry")
	// remember the first such variable
	*piVarReused = kit.Kit_WordFindFirstBit(Entry)
	*piVar = int32(spStatic_Lpk_MapSuppRedDecSelect_pSetBest.iVar)
	return (spStatic_Lpk_MapSuppRedDecSelect_pSetBest.uSubset1 << 16) | (spStatic_Lpk_MapSuppRedDecSelect_pSetBest.uSubset0 & 0xFFFF)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

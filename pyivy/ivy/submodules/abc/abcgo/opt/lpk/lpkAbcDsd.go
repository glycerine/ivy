package lpk

import (
	"fmt"
	kit "github.com/glycerine/abcgo/bool/kit"
	util "github.com/glycerine/abcgo/misc/util"
	vec "github.com/glycerine/abcgo/misc/vec"
	"os"
)

/**CFile****************************************************************

  FileName    [lpkAbcDsd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    [LUT-decomposition based on recursive DSD.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkAbcDsd.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

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

	Synopsis    [Cofactors TTs w.r.t. all vars and finds the best var.]

	Description [The best variable is the variable with the minimum
	sum total of the support sizes of all truth tables. This procedure
	computes and returns cofactors w.r.t. the best variable.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_FunComputeMinSuppSizeVar(p *Lpk_Fun_t, ppTruths [][]uint32, nTruths int32, ppCofs [][]uint32, uNonDecSupp uint32) int32 {
	var i, Var, VarBest, nSuppSize0, nSuppSize1 int32
	var nSuppTotalMin int32 = -1 // Suppress "might be used uninitialized"
	var nSuppTotalCur int32
	var nSuppMaxMin int32 = -1 // Suppress "might be used uninitialized"
	var nSuppMaxCur int32
	assert(nTruths > 0, "nTruths > 0")
	VarBest = -1
	// inlining of macro Lpk_SuppForEachVar
	for Var = 0; Var < 16; Var++ {
		if (p.uSupp & (1 << Var)) == 0 {
		} else {
			if (uNonDecSupp & (1 << Var)) == 0 {
				continue
			}
			nSuppMaxCur = 0
			nSuppTotalCur = 0
			for i = 0; i < nTruths; i++ {
				if nTruths == 1 {
					nSuppSize0 = kit.Kit_WordCountOnes(p.puSupps[2*Var+0])
					nSuppSize1 = kit.Kit_WordCountOnes(p.puSupps[2*Var+1])

				} else {
					kit.Kit_TruthCofactor0New(ppCofs[2*i+0], ppTruths[i], int32(p.nVars), Var)
					kit.Kit_TruthCofactor1New(ppCofs[2*i+1], ppTruths[i], int32(p.nVars), Var)
					nSuppSize0 = kit.Kit_TruthSupportSize(ppCofs[2*i+0], int32(p.nVars))
					nSuppSize1 = kit.Kit_TruthSupportSize(ppCofs[2*i+1], int32(p.nVars))
				}
				nSuppMaxCur = util.Abc_MaxInt(nSuppMaxCur, nSuppSize0)
				nSuppMaxCur = util.Abc_MaxInt(nSuppMaxCur, nSuppSize1)
				nSuppTotalCur += nSuppSize0 + nSuppSize1
			}
			if VarBest == -1 || nSuppMaxMin > nSuppMaxCur ||
				(nSuppMaxMin == nSuppMaxCur && nSuppTotalMin > nSuppTotalCur) {
				VarBest = Var
				nSuppMaxMin = nSuppMaxCur
				nSuppTotalMin = nSuppTotalCur
			}
		}
	}
	// recompute cofactors for the best var
	for i = 0; i < nTruths; i++ {
		kit.Kit_TruthCofactor0New(ppCofs[2*i+0], ppTruths[i], int32(p.nVars), VarBest)
		kit.Kit_TruthCofactor1New(ppCofs[2*i+1], ppTruths[i], int32(p.nVars), VarBest)
	}
	return VarBest
}

/*
*Function*************************************************************

	Synopsis    [Recursively computes decomposable subsets.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_ComputeBoundSets_rec(p *Kit_DsdNtk_t, iLit int32, vSets *Vec_Int_t, nSizeMax int32) uint32 {
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
			uSupps[i] = Lpk_ComputeBoundSets_rec(p, iLitFanin, vSets, nSizeMax)
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
			if kit.Kit_WordCountOnes(uSuppCur) <= nSizeMax {
				vec.Vec_IntPush(vSets, int32(uSuppCur))
			}
		}
		return uSupport
	}
	assert(pObj.Type == kit.KIT_DSD_PRIME, "pObj.Type == kit.KIT_DSD_PRIME")
	// get the cumulative support of all fanins
	uSupport = 0
	// inlining of macro Kit_DsdObjForEachFanin
	for i = 0; i < uint32(kit.Kit_DsdObjFanNum(pObj)); i++ {
		iLitFanin = int32(kit.Kit_DsdObjFans(pObj)[i])
		uSuppCur = Lpk_ComputeBoundSets_rec(p, iLitFanin, vSets, nSizeMax)
		uSupport |= uSuppCur
		if kit.Kit_WordCountOnes(uSuppCur) <= nSizeMax {
			vec.Vec_IntPush(vSets, int32(uSuppCur))
		}
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
func Lpk_ComputeBoundSets(p *Kit_DsdNtk_t, nSizeMax int32) *Vec_Int_t {
	var vSets *Vec_Int_t
	var uSupport, Entry uint32
	var Number, i int32
	assert(kit.Kit_DsdNtkVarNum(p) <= 16, "p.nVars <= 16")
	vSets = vec.Vec_IntAlloc(100)
	vec.Vec_IntPush(vSets, 0)
	if kit.Kit_DsdNtkRoot(p).Type == kit.KIT_DSD_CONST1 {
		return vSets
	}
	if kit.Kit_DsdNtkRoot(p).Type == kit.KIT_DSD_VAR {
		uSupport = (1 << util.Abc_Lit2Var(int32(kit.Kit_DsdObjFans(kit.Kit_DsdNtkRoot(p))[0])))
		if kit.Kit_WordCountOnes(uSupport) <= nSizeMax {
			vec.Vec_IntPush(vSets, int32(uSupport))
		}
		return vSets
	}
	uSupport = Lpk_ComputeBoundSets_rec(p, int32(p.Root), vSets, nSizeMax)
	assert((uSupport&0xFFFF0000) == 0, "(uSupport & 0xFFFF0000) == 0")
	// add the total support of the network
	if kit.Kit_WordCountOnes(uSupport) <= nSizeMax {
		vec.Vec_IntPush(vSets, int32(uSupport))
	}
	// set the remaining variables
	for i = 0; i < int32(vec.Vec_IntSize(vSets)); i++ {
		Number = vec.Vec_IntEntry(vSets, int(i))
		Entry = uint32(Number)
		vec.Vec_IntWriteEntry(vSets, int(i), int32(Entry|((uSupport&^Entry)<<16)))
	}
	return vSets
}

/*
*Function*************************************************************

	Synopsis    [Prints the sets of subsets.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func lpkDsdPrintSetOne(uSupport int32) {
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
func lpkDsdPrintSets(vSets *Vec_Int_t) {
	var uSupport uint32
	var Number, i int32
	fmt.Printf("Subsets(%d): ", vec.Vec_IntSize(vSets))
	for i = 0; i < int32(vec.Vec_IntSize(vSets)); i++ {
		Number = vec.Vec_IntEntry(vSets, int(i))
		uSupport = uint32(Number)
		lpkDsdPrintSetOne(int32(uSupport))
	}
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Merges two bound sets.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_MergeBoundSets(vSets0 *Vec_Int_t, vSets1 *Vec_Int_t, nSizeMax int32) *Vec_Int_t {
	var vSets *Vec_Int_t
	var Entry0, Entry1, Entry int32
	var i, k int32
	vSets = vec.Vec_IntAlloc(100)
	for i = 0; i < int32(vec.Vec_IntSize(vSets0)); i++ {
		Entry0 = vec.Vec_IntEntry(vSets0, int(i))
		for k = 0; k < int32(vec.Vec_IntSize(vSets1)); k++ {
			Entry1 = vec.Vec_IntEntry(vSets1, int(k))
			Entry = Entry0 | Entry1
			if Entry&(Entry>>16) != 0 {
				continue
			}
			if kit.Kit_WordCountOnes(uint32(Entry&0xffff)) <= nSizeMax {
				vec.Vec_IntPush(vSets, Entry)
			}
		}
	}
	return vSets
}

/*
*Function*************************************************************

	Synopsis    [Performs DSD-based decomposition of the function.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_FunCompareBoundSets(p *Lpk_Fun_t, vBSets *Vec_Int_t, nCofDepth int32, uNonDecSupp uint32, uLateArrSupp uint32, pRes *Lpk_Res_t) {
	var FVerbose int32
	var uBoundSet uint32
	var i, nVarsBS, nVarsRem, Delay, Area int32

	// compare the resulting boundsets
	*pRes = Lpk_Res_t{}
	for i = 0; i < int32(vec.Vec_IntSize(vBSets)); i++ {
		uBoundSet = uint32(vec.Vec_IntEntry(vBSets, int(i)))
		if (uBoundSet & 0xFFFF) == 0 { // skip empty boundset
			continue
		}
		if (uBoundSet & uNonDecSupp) == 0 { // skip those boundsets that are not in the domain of interest
			continue
		}
		if uBoundSet&uLateArrSupp != 0 { // skip those boundsets that are late arriving
			continue
		}
		if FVerbose != 0 {
			lpkDsdPrintSetOne(int32(uBoundSet & 0xFFFF))
			// printf( "\n" );
			// lpkDsdPrintSetOne( uBoundSet >> 16 );
			// printf( "\n" );
		}
		assert((uBoundSet&(uBoundSet>>16)) == 0, "(uBoundSet & (uBoundSet >> 16)) == 0")
		nVarsBS = kit.Kit_WordCountOnes(uBoundSet & 0xFFFF)
		if nVarsBS == 1 {
			continue
		}
		assert(nVarsBS <= int32(p.nLutK)-nCofDepth, "nVarsBS <= int32(p.nLutK) - nCofDepth")
		nVarsRem = int32(p.nVars) - nVarsBS + 1
		Area = 1 + Lpk_LutNumLuts(nVarsRem, int32(p.nLutK))
		Delay = 1 + Lpk_SuppDelay(uBoundSet&0xFFFF, p.pDelays[:])
		if FVerbose != 0 {
			fmt.Printf("area = %d limit = %d  delay = %d limit = %d\n", Area, int32(p.nAreaLim), Delay, int32(p.nDelayLim))
		}
		if Area > int32(p.nAreaLim) || Delay > int32(p.nDelayLim) {
			continue
		}
		if pRes.BSVars == 0 || pRes.nSuppSizeL > nVarsRem || (pRes.nSuppSizeL == nVarsRem && pRes.DelayEst > Delay) {
			pRes.nBSVars = nVarsBS
			pRes.BSVars = (uBoundSet & 0xFFFF)
			pRes.nSuppSizeS = nVarsBS + nCofDepth
			pRes.nSuppSizeL = nVarsRem
			pRes.DelayEst = Delay
			pRes.AreaEst = Area
		}
	}
	if FVerbose != 0 {
		if pRes.BSVars != 0 {
			fmt.Printf("Found bound set ")
			lpkDsdPrintSetOne(int32(pRes.BSVars))
			fmt.Printf("\n")

		} else {
			fmt.Printf("Did not find boundsets.\n")
		}
		fmt.Printf("\n")
	}
	if pRes.BSVars != 0 {
		assert(pRes.DelayEst <= int32(p.nDelayLim), "pRes.DelayEst <= int32(p.nDelayLim)")
		assert(pRes.AreaEst <= int32(p.nAreaLim), "pRes.AreaEst <= int32(p.nAreaLim)")
	}
}

/*
*Function*************************************************************

	Synopsis    [Finds late arriving inputs, which cannot be in the bound set.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_DsdLateArriving(p *Lpk_Fun_t) uint32 {
	var i, uLateArrSupp uint32
	// inlining of macro Lpk_SuppForEachVar
	for i = 0; i < 16; i++ {
		if (p.uSupp & (1 << i)) == 0 {
		} else if p.pDelays[i] > int32(p.nDelayLim)-2 {
			uLateArrSupp |= (1 << i)
		}
	}
	return uLateArrSupp
}

/*
*Function*************************************************************

	Synopsis    [Performs DSD-based decomposition of the function.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_DsdAnalizeOne(p *Lpk_Fun_t, ppTruths [][16][]uint32, pNtks []*Kit_DsdNtk_t, pCofVars []byte, nCofDepth int32, pRes *Lpk_Res_t) int32 {
	var FVerbose int32
	var pvBSets [4][8]*Vec_Int_t
	var uNonDecSupp, uLateArrSupp uint32
	var i, k, nNonDecSize, nNonDecSizeMax int32
	assert(nCofDepth >= 1 && nCofDepth <= 3, "nCofDepth >= 1 && nCofDepth <= 3")
	assert(nCofDepth < int32(p.nLutK)-1, "nCofDepth < int32(p.nLutK) - 1")
	assert(p.fSupports != 0, "p.fSupports")

	// find the support of the largest non-DSD block
	nNonDecSizeMax = 0
	uNonDecSupp = p.uSupp
	for i = 0; i < (1 << (nCofDepth - 1)); i++ {
		nNonDecSize = kit.Kit_DsdNonDsdSizeMax(pNtks[i])
		if nNonDecSizeMax < nNonDecSize {
			nNonDecSizeMax = nNonDecSize
			uNonDecSupp = kit.Kit_DsdNonDsdSupports(pNtks[i])

		} else if nNonDecSizeMax == nNonDecSize {
			uNonDecSupp |= kit.Kit_DsdNonDsdSupports(pNtks[i])
		}
	}

	// remove those variables that cannot be used because of delay constraints
	// if variables arrival time is more than p->DelayLim - 2, it cannot be used
	uLateArrSupp = Lpk_DsdLateArriving(p)
	if (uNonDecSupp & ^uLateArrSupp) == 0 {
		*pRes = Lpk_Res_t{}
		return 0
	}

	// find the next cofactoring variable
	pCofVars[nCofDepth-1] = byte(Lpk_FunComputeMinSuppSizeVar(p, ppTruths[nCofDepth-1][:], 1<<(nCofDepth-1), ppTruths[nCofDepth][:], uNonDecSupp&^uLateArrSupp))

	// derive decomposed networks
	for i = 0; i < (1 << nCofDepth); i++ {
		if pNtks[i] != nil {
			kit.Kit_DsdNtkFree(pNtks[i])
		}
		pNtks[i] = kit.Kit_DsdDecomposeExpand(ppTruths[nCofDepth][i], int32(p.nVars))
		if FVerbose != 0 {
			kit.Kit_DsdPrint(os.Stdout, pNtks[i])
		}
		pvBSets[nCofDepth][i] = Lpk_ComputeBoundSets(pNtks[i], int32(p.nLutK)-nCofDepth) // try restricting to those in uNonDecSupp!!!
	}

	// derive the set of feasible boundsets
	for i = nCofDepth - 1; i >= 0; i-- {
		for k = 0; k < (1 << i); k++ {
			pvBSets[i][k] = Lpk_MergeBoundSets(pvBSets[i+1][2*k+0], pvBSets[i+1][2*k+1], int32(p.nLutK)-nCofDepth)
		}
	}
	// compare bound-sets
	Lpk_FunCompareBoundSets(p, pvBSets[0][0], nCofDepth, uNonDecSupp, uLateArrSupp, pRes)
	// free the bound sets
	for i = nCofDepth; i >= 0; i-- {
		for k = 0; k < (1 << i); k++ {
			vec.Vec_IntFree(pvBSets[i][k])
		}
	}

	// copy the cofactoring variables
	if pRes.BSVars != 0 {
		pRes.nCofVars = nCofDepth
		for i = 0; i < nCofDepth; i++ {
			pRes.pCofVars[i] = pCofVars[i]
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Performs DSD-based decomposition of the function.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Lpk_DsdAnalize_Res0 Lpk_Res_t
var spStatic_Lpk_DsdAnalize_pRes0 *Lpk_Res_t = &spStatic_Lpk_DsdAnalize_Res0
var spStatic_Lpk_DsdAnalize_Res1 Lpk_Res_t
var spStatic_Lpk_DsdAnalize_pRes1 *Lpk_Res_t = &spStatic_Lpk_DsdAnalize_Res1
var spStatic_Lpk_DsdAnalize_Res2 Lpk_Res_t
var spStatic_Lpk_DsdAnalize_pRes2 *Lpk_Res_t = &spStatic_Lpk_DsdAnalize_Res2
var spStatic_Lpk_DsdAnalize_Res3 Lpk_Res_t
var spStatic_Lpk_DsdAnalize_pRes3 *Lpk_Res_t = &spStatic_Lpk_DsdAnalize_Res3

func Lpk_DsdAnalize(pMan *Lpk_Man_t, p *Lpk_Fun_t, nShared int32) *Lpk_Res_t {

	var fUseBackLooking int32 = 1
	var pRes *Lpk_Res_t = nil
	var vBSets *Vec_Int_t
	var pNtks [8]*Kit_DsdNtk_t = [8]*Kit_DsdNtk_t{nil}
	var pCofVars [5]byte
	var i int32

	assert(p.nLutK >= 3, "p.nLutK >= 3")
	assert(nShared >= 0 && nShared <= 3, "nShared >= 0 && nShared <= 3")
	assert(p.uSupp == kit.Kit_BitMask(int32(p.nVars)), "p.uSupp == kit.Kit_BitMask(int32(p.nVars))")

	// try decomposition without cofactoring
	pNtks[0] = kit.Kit_DsdDecomposeExpand(Lpk_FunTruth(p, 0), int32(p.nVars))
	if pMan.pPars.FVerbose != 0 {
		pMan.nBlocks[kit.Kit_DsdNonDsdSizeMax(pNtks[0])]++
	}
	vBSets = Lpk_ComputeBoundSets(pNtks[0], int32(p.nLutK))
	Lpk_FunCompareBoundSets(p, vBSets, 0, 0xFFFF, Lpk_DsdLateArriving(p), spStatic_Lpk_DsdAnalize_pRes0)
	vec.Vec_IntFree(vBSets)

	// check the result
	if spStatic_Lpk_DsdAnalize_pRes0.nBSVars == int32(p.nLutK) {
		pRes = spStatic_Lpk_DsdAnalize_pRes0
		goto finish
	}
	if spStatic_Lpk_DsdAnalize_pRes0.nBSVars == int32(p.nLutK)-1 {
		pRes = spStatic_Lpk_DsdAnalize_pRes0
		goto finish
	}
	if nShared == 0 {
		goto finish
	}

	// prepare storage
	kit.Kit_TruthCopy(pMan.ppTruths[0][0], Lpk_FunTruth(p, 0), int32(p.nVars))

	// cofactor 1 time
	if !(Lpk_DsdAnalizeOne(p, pMan.ppTruths[:], pNtks[:], pCofVars[:], 1, spStatic_Lpk_DsdAnalize_pRes1) != 0) {
		goto finish
	}
	assert(spStatic_Lpk_DsdAnalize_pRes1.nBSVars <= int32(p.nLutK)-1, "spStatic_Lpk_DsdAnalize_pRes1.nBSVars <= int32(p.nLutK) - 1")
	if spStatic_Lpk_DsdAnalize_pRes1.nBSVars == int32(p.nLutK)-1 {
		pRes = spStatic_Lpk_DsdAnalize_pRes1
		goto finish
	}
	if spStatic_Lpk_DsdAnalize_pRes0.nBSVars == int32(p.nLutK)-2 {
		pRes = spStatic_Lpk_DsdAnalize_pRes0
		goto finish
	}
	if spStatic_Lpk_DsdAnalize_pRes1.nBSVars == int32(p.nLutK)-2 {
		pRes = spStatic_Lpk_DsdAnalize_pRes1
		goto finish
	}
	if nShared == 1 {
		goto finish
	}

	// cofactor 2 times
	if p.nLutK >= 4 {
		if !(Lpk_DsdAnalizeOne(p, pMan.ppTruths[:], pNtks[:], pCofVars[:], 2, spStatic_Lpk_DsdAnalize_pRes2) != 0) {
			goto finish
		}
		assert(spStatic_Lpk_DsdAnalize_pRes2.nBSVars <= int32(p.nLutK)-2, "spStatic_Lpk_DsdAnalize_pRes2.nBSVars <= int32(p.nLutK) - 2")
		if spStatic_Lpk_DsdAnalize_pRes2.nBSVars == int32(p.nLutK)-2 {
			pRes = spStatic_Lpk_DsdAnalize_pRes2
			goto finish
		}
		if fUseBackLooking != 0 {
			if spStatic_Lpk_DsdAnalize_pRes0.nBSVars == int32(p.nLutK)-3 {
				pRes = spStatic_Lpk_DsdAnalize_pRes0
				goto finish
			}
			if spStatic_Lpk_DsdAnalize_pRes1.nBSVars == int32(p.nLutK)-3 {
				pRes = spStatic_Lpk_DsdAnalize_pRes1
				goto finish
			}
		}
		if spStatic_Lpk_DsdAnalize_pRes2.nBSVars == int32(p.nLutK)-3 {
			pRes = spStatic_Lpk_DsdAnalize_pRes2
			goto finish
		}
		if nShared == 2 {
			goto finish
		}
		assert(nShared == 3, "nShared == 3")
	}

	// cofactor 3 times
	if p.nLutK >= 5 {
		if !(Lpk_DsdAnalizeOne(p, pMan.ppTruths[:], pNtks[:], pCofVars[:], 3, spStatic_Lpk_DsdAnalize_pRes3) != 0) {
			goto finish
		}
		assert(spStatic_Lpk_DsdAnalize_pRes3.nBSVars <= int32(p.nLutK)-3, "spStatic_Lpk_DsdAnalize_pRes3.nBSVars <= int32(p.nLutK) - 3")
		if spStatic_Lpk_DsdAnalize_pRes3.nBSVars == int32(p.nLutK)-3 {
			pRes = spStatic_Lpk_DsdAnalize_pRes3
			goto finish
		}
		if fUseBackLooking != 0 {
			if spStatic_Lpk_DsdAnalize_pRes0.nBSVars == int32(p.nLutK)-4 {
				pRes = spStatic_Lpk_DsdAnalize_pRes0
				goto finish
			}
			if spStatic_Lpk_DsdAnalize_pRes1.nBSVars == int32(p.nLutK)-4 {
				pRes = spStatic_Lpk_DsdAnalize_pRes1
				goto finish
			}
			if spStatic_Lpk_DsdAnalize_pRes2.nBSVars == int32(p.nLutK)-4 {
				pRes = spStatic_Lpk_DsdAnalize_pRes2
				goto finish
			}
		}
		if spStatic_Lpk_DsdAnalize_pRes3.nBSVars == int32(p.nLutK)-4 {
			pRes = spStatic_Lpk_DsdAnalize_pRes3
			goto finish
		}
	}

finish:
	// free the networks
	for i = 0; i < (1 << nShared); i++ {
		if pNtks[i] != nil {
			kit.Kit_DsdNtkFree(pNtks[i])
		}
	}
	// choose the best under these conditions
	return pRes
}

/*
*Function*************************************************************

	Synopsis    [Splits the function into two subfunctions using DSD.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_DsdSplit(pMan *Lpk_Man_t, p *Lpk_Fun_t, pCofVars []byte, nCofVars int32, uBoundSet uint32) *Lpk_Fun_t {
	var pNew *Lpk_Fun_t
	var pNtkDec *Kit_DsdNtk_t
	var i, k, iVacVar, nCofs int32
	// prepare storage
	kit.Kit_TruthCopy(pMan.ppTruths[0][0], Lpk_FunTruth(p, 0), int32(p.nVars))
	// get the vacuous variable
	iVacVar = kit.Kit_WordFindFirstBit(uBoundSet)
	// compute the cofactors
	for i = 0; i < nCofVars; i++ {
		for k = 0; k < (1 << i); k++ {
			kit.Kit_TruthCofactor0New(pMan.ppTruths[i+1][2*k+0], pMan.ppTruths[i][k], int32(p.nVars), int32(pCofVars[i]))
			kit.Kit_TruthCofactor1New(pMan.ppTruths[i+1][2*k+1], pMan.ppTruths[i][k], int32(p.nVars), int32(pCofVars[i]))
		}
	}
	// decompose each cofactor w.r.t. the bound set
	nCofs = (1 << nCofVars)
	for k = 0; k < nCofs; k++ {
		pNtkDec = kit.Kit_DsdDecomposeExpand(pMan.ppTruths[nCofVars][k], int32(p.nVars))
		kit.Kit_DsdTruthPartialTwo(pMan.pDsdMan, pNtkDec, uBoundSet, iVacVar, pMan.ppTruths[nCofVars+1][k], pMan.ppTruths[nCofVars+1][nCofs+k])
		kit.Kit_DsdNtkFree(pNtkDec)
	}
	// compute the composition/decomposition functions (they will be in pMan->ppTruths[1][0]/pMan->ppTruths[1][1])
	for i = nCofVars; i >= 1; i-- {
		for k = 0; k < (1 << i); k++ {
			kit.Kit_TruthMuxVar(pMan.ppTruths[i][k], pMan.ppTruths[i+1][2*k+0], pMan.ppTruths[i+1][2*k+1], int32(p.nVars), int32(pCofVars[i-1]))
		}
	}

	// derive the new component (decomposition function)
	pNew = Lpk_FunDup(p, pMan.ppTruths[1][1])
	// update the old component (composition function)
	kit.Kit_TruthCopy(Lpk_FunTruth(p, 0), pMan.ppTruths[1][0], int32(p.nVars))
	p.uSupp = kit.Kit_TruthSupport(Lpk_FunTruth(p, 0), int32(p.nVars))
	p.pFanins[iVacVar] = byte(pNew.Id)
	p.pDelays[iVacVar] = Lpk_SuppDelay(pNew.uSupp, pNew.pDelays[:])
	// support minimize both
	p.fSupports = 0
	Lpk_FunSuppMinimize(p)
	Lpk_FunSuppMinimize(pNew)
	// update delay and area requirements
	pNew.nDelayLim = uint32(p.pDelays[iVacVar])
	pNew.nAreaLim = 1
	p.nAreaLim = p.nAreaLim - 1
	return pNew
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

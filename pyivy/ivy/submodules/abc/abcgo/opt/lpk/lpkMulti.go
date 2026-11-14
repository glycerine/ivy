package lpk

import (
	"fmt"
	kit "github.com/glycerine/abcgo/bool/kit"
	ifpkg "github.com/glycerine/abcgo/map/if"
	util "github.com/glycerine/abcgo/misc/util"
	"os"
)

/**CFile****************************************************************

  FileName    [lpkMulti.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpkMulti.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

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

	Synopsis    [Records variable order.]

	Description [Increaments Order[x][y] by 1 if x should be above y in the DSD.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_CreateVarOrder(pNtk *Kit_DsdNtk_t, pTable [][16]byte) {
	var pObj *Kit_DsdObj_t
	var uSuppFanins uint32
	var k int32
	var Above, Below [16]int32
	var nAbove, nBelow, iFaninLit, x, y int32
	// iterate through the nodes
	// inlining of macro Kit_DsdNtkForEachObj
	for _, node := range kit.Kit_DsdNtkForEachObj(pNtk) {
		pObj = node
		// collect fanin support of this node
		nAbove = 0
		uSuppFanins = 0
		// inlining of macro Kit_DsdObjForEachFanin
		for k = 0; k < kit.Kit_DsdObjFanNum(pObj); k++ {
			iFaninLit = int32(kit.Kit_DsdObjFans(pObj)[k])
			if kit.Kit_DsdLitIsLeaf(pNtk, iFaninLit) != 0 {
				Above[nAbove] = util.Abc_Lit2Var(iFaninLit)
				nAbove++
			} else {
				uSuppFanins |= kit.Kit_DsdLitSupport(pNtk, iFaninLit)
			}
		}
		// find the below variables
		nBelow = 0
		for y = 0; y < 16; y++ {
			if (uSuppFanins & (1 << y)) != 0 {
				Below[nBelow] = y
				nBelow++
			}
		}
		// create all pairs
		for x = 0; x < nAbove; x++ {
			for y = 0; y < nBelow; y++ {
				pTable[Above[x]][Below[y]]++
			}
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Creates commmon variable order.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_CreateCommonOrder(pTable [][16]byte, piCofVar []int32, nCBars int32, pPrios []int32, nVars int32, fVerbose int32) {
	var Score [16]int32
	var pPres [16]int32
	var i, y, x, iVarBest, ScoreMax, PrioCount int32

	// mark the present variables
	for i = 0; i < nVars; i++ {
		pPres[i] = 1
	}
	// remove cofactored variables
	for i = 0; i < nCBars; i++ {
		pPres[piCofVar[i]] = 0
	}

	// compute scores for each leaf
	for i = 0; i < nVars; i++ {
		if pPres[i] == 0 {
			continue
		}
		for y = 0; y < nVars; y++ {
			Score[i] += int32(pTable[i][y])
		}
		for x = 0; x < nVars; x++ {
			Score[i] -= int32(pTable[x][i])
		}
	}

	// print the scores
	if fVerbose != 0 {
		fmt.Printf("Scores: ")
		for i = 0; i < nVars; i++ {
			fmt.Printf("%c=%d ", 'a'+i, Score[i])
		}
		fmt.Printf("   ")
		fmt.Printf("Prios: ")
	}

	// derive variable priority
	// variables with equal score receive the same priority
	for i = 0; i < nVars; i++ {
		pPrios[i] = 16
	}

	// iterate until variables remain
	for PrioCount = 1; ; PrioCount++ {
		// find the present variable with the highest score
		iVarBest = -1
		ScoreMax = -100000
		for i = 0; i < nVars; i++ {
			if pPres[i] == 0 {
				continue
			}
			if ScoreMax < Score[i] {
				ScoreMax = Score[i]
				iVarBest = i
			}
		}
		if iVarBest == -1 {
			break
		}
		// give the next priority to all vars having this score
		if fVerbose != 0 {
			fmt.Printf("%d=", PrioCount)
		}
		for i = 0; i < nVars; i++ {
			if pPres[i] == 0 {
				continue
			}
			if Score[i] == ScoreMax {
				pPrios[i] = PrioCount
				pPres[i] = 0
				if fVerbose != 0 {
					fmt.Printf("%c", 'a'+i)
				}
			}
		}
		if fVerbose != 0 {
			fmt.Printf(" ")
		}
	}
	if fVerbose != 0 {
		fmt.Printf("\n")
	}
}

/*
*Function*************************************************************

	Synopsis    [Finds components with the highest priority.]

	Description [Returns the number of components selected.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_FindHighest(ppNtks []*Kit_DsdNtk_t, piLits []int32, nSize int32, pPrio []int32, pDecision []int32) int32 {
	var pObj *Kit_DsdObj_t
	var uSupps [8]uint32
	var uSuppFanin uint32
	var uSuppTotal uint32
	var uSuppLarge uint32
	var i int32
	var pTriv [8]int32
	var PrioMin int32
	var iVarMax int32
	var nComps int32
	var fOneNonTriv int32

	// find individual support and total support
	uSuppTotal = 0
	for i = 0; i < nSize; i++ {
		pTriv[i] = 1
		if piLits[i] < 0 {
			uSupps[i] = 0
		} else if kit.Kit_DsdLitIsLeaf(ppNtks[i], piLits[i]) != 0 {
			uSupps[i] = kit.Kit_DsdLitSupport(ppNtks[i], piLits[i])
		} else {
			pObj = kit.Kit_DsdNtkObj(ppNtks[i], util.Abc_Lit2Var(piLits[i]))
			if pObj.Type == kit.KIT_DSD_PRIME {
				pTriv[i] = 0
				uSuppFanin = kit.Kit_DsdLitSupport(ppNtks[i], int32(kit.Kit_DsdObjFans(pObj)[0]))

			} else {
				assert(kit.Kit_DsdObjFanNum(pObj) == 2, "kit.Kit_DsdObjFanNum(pObj) == 2")
				if kit.Kit_DsdLitIsLeaf(ppNtks[i], int32(kit.Kit_DsdObjFans(pObj)[0])) == 0 {
					pTriv[i] = 0
				}
				uSuppFanin = kit.Kit_DsdLitSupport(ppNtks[i], int32(kit.Kit_DsdObjFans(pObj)[1]))
			}
			uSupps[i] = kit.Kit_DsdLitSupport(ppNtks[i], piLits[i]) & ^uSuppFanin
		}
		assert(uSupps[i] <= 0xFFFF, "uSupps[i] <= 0xFFFF")
		uSuppTotal |= uSupps[i]
	}
	if uSuppTotal == 0 {
		return 0
	}

	// find one support variable with the highest priority
	PrioMin = util.ABC_INFINITY
	iVarMax = -1
	for i = 0; i < 16; i++ {
		if (uSuppTotal & (1 << i)) != 0 {
			if PrioMin > pPrio[i] {
				PrioMin = pPrio[i]
				iVarMax = i
			}
		}
	}
	assert(iVarMax != -1, "iVarMax != -1")

	// select components, which have this variable
	nComps = 0
	fOneNonTriv = 0
	uSuppLarge = 0
	for i = 0; i < nSize; i++ {
		if (uSupps[i] & (1 << iVarMax)) != 0 {
			if pTriv[i] != 0 || fOneNonTriv == 0 {
				if pTriv[i] == 0 {
					uSuppLarge = uSupps[i]
					fOneNonTriv = 1
				}
				pDecision[i] = 1
				nComps++

			} else {
				pDecision[i] = 0
			}

		} else {
			pDecision[i] = 0
		}
	}

	// add other non-trivial not-taken components whose support is contained in the current large component support
	if fOneNonTriv != 0 {
		for i = 0; i < nSize; i++ {
			if pTriv[i] == 0 && pDecision[i] == 0 && (uSupps[i]&^uSuppLarge) == 0 {
				pDecision[i] = 1
				nComps++
			}
		}
	}

	return nComps
}

/*
*Function*************************************************************

	Synopsis    [Prepares the mapping manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Lpk_MapTreeMulti_rec(p *Lpk_Man_t, ppNtks []*Kit_DsdNtk_t, piLits []int32, piCofVar []int32, nCBars int32, ppLeaves []*If_Obj_t, nLeaves int32, pPrio []int32) *If_Obj_t {
	var pObj *Kit_DsdObj_t
	var pObjsNew [4][8]*If_Obj_t
	var pResPrev *If_Obj_t
	var piLitsNew, pDecision [8]int32
	var i, k, nComps, nSize int32

	// find which of the variables is highest in the order
	nSize = (1 << nCBars)
	nComps = Lpk_FindHighest(ppNtks[:], piLits, nSize, pPrio, pDecision[:])
	if nComps == 0 {
		return ifpkg.If_Not(ifpkg.If_ManConst1(p.pIfMan))
	}

	// iterate over the nodes
	if p.pPars.fVeryVerbose != 0 {
		fmt.Printf("Decision: ")
	}
	for i = 0; i < nSize; i++ {
		if pDecision[i] != 0 {
			if p.pPars.fVeryVerbose != 0 {
				fmt.Printf("%d ", i)
			}
			assert(piLits[i] >= 0, "piLits[i] >= 0")
			pObj = kit.Kit_DsdNtkObj(ppNtks[i], util.Abc_Lit2Var(piLits[i]))
			if pObj == nil {
				piLitsNew[i] = -2
			} else if pObj.Type == kit.KIT_DSD_PRIME {
				piLitsNew[i] = int32(kit.Kit_DsdObjFans(pObj)[0])
			} else {
				piLitsNew[i] = int32(kit.Kit_DsdObjFans(pObj)[1])
			}

		} else {
			piLitsNew[i] = piLits[i]
		}
	}
	if p.pPars.fVeryVerbose != 0 {
		fmt.Printf("\n")
	}

	// call again
	pResPrev = Lpk_MapTreeMulti_rec(p, ppNtks[:], piLitsNew[:], piCofVar, nCBars, ppLeaves[:], nLeaves, pPrio)

	// create new set of nodes
	for i = 0; i < nSize; i++ {
		if pDecision[i] != 0 {
			pObjsNew[nCBars][i] = Lpk_MapTree_rec(p, ppNtks[i], ppLeaves[:], piLits[i], pResPrev)
		} else if piLits[i] == -1 {
			pObjsNew[nCBars][i] = ifpkg.If_ManConst1(p.pIfMan)
		} else if piLits[i] == -2 {
			pObjsNew[nCBars][i] = ifpkg.If_Not(ifpkg.If_ManConst1(p.pIfMan))
		} else {
			pObjsNew[nCBars][i] = pResPrev
		}
	}

	// create MUX using these outputs
	for k = nCBars; k > 0; k-- {
		nSize /= 2
		for i = 0; i < nSize; i++ {
			pObjsNew[k-1][i] = ifpkg.If_ManCreateMux(p.pIfMan, pObjsNew[k][2*i+0], pObjsNew[k][2*i+1], ppLeaves[piCofVar[k-1]])
		}
	}
	assert(nSize == 1, "nSize == 1")
	return pObjsNew[0][0]
}

/*
*Function*************************************************************

	Synopsis    [Prepares the mapping manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Lpk_MapTreeMulti_Counter int32

func Lpk_MapTreeMulti(p *Lpk_Man_t, pTruth []uint32, nVars int32, ppLeaves []*If_Obj_t) *If_Obj_t {

	var pResult *If_Obj_t
	var ppNtks [8]*Kit_DsdNtk_t
	var pTemp *Kit_DsdNtk_t
	var pRoot *Kit_DsdObj_t
	var piCofVar [4]int32
	var pPrios [16]int32
	var pFreqs [16]int32
	var piLits [16]int32
	var i, k, nCBars, nSize, nMemSize int32
	var ppCofs [4][8][]uint32
	var uSupport uint32
	var pTable [16][16]byte = [16][16]byte{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	var fVerbose int32 = p.pPars.fVeryVerbose

	spStatic_Lpk_MapTreeMulti_Counter++
	//    printf( "Run %d.\n", Counter );

	// allocate storage for cofactors
	nMemSize = kit.Kit_TruthWordNum(nVars)
	ppCofs[0][0] = make([]uint32, 32*int(nMemSize))
	nSize = 0
	for i = 0; i < 4; i++ {
		for k = 0; k < 8; k++ {
			ppCofs[i][k] = ppCofs[0][0][int(nMemSize*nSize):]
			nSize++
		}
	}
	assert(nSize == 32, "nSize == 32")

	// find the best cofactoring variables
	nCBars = kit.Kit_DsdCofactoring(pTruth, nVars, piCofVar[:], p.pPars.nVarsShared, 0)
	//    nCBars = 2;
	//    piCofVar[0] = 0;
	//    piCofVar[1] = 1;

	// copy the function
	kit.Kit_TruthCopy(ppCofs[0][0], pTruth, nVars)

	// decompose w.r.t. these variables
	for k = 0; k < nCBars; k++ {
		nSize = (1 << k)
		for i = 0; i < nSize; i++ {
			kit.Kit_TruthCofactor0New(ppCofs[k+1][2*i+0], ppCofs[k][i], nVars, piCofVar[k])
			kit.Kit_TruthCofactor1New(ppCofs[k+1][2*i+1], ppCofs[k][i], nVars, piCofVar[k])
		}
	}
	nSize = (1 << nCBars)
	// compute DSD networks
	for i = 0; i < nSize; i++ {
		ppNtks[i] = kit.Kit_DsdDecompose(ppCofs[nCBars][i], nVars)
		pTemp = ppNtks[i]
		ppNtks[i] = kit.Kit_DsdExpand(pTemp)
		kit.Kit_DsdNtkFree(pTemp)
		if fVerbose != 0 {
			fmt.Printf("Cof%d%d: ", nCBars, i)
			kit.Kit_DsdPrint(os.Stdout, ppNtks[i])
		}
	}

	// compute variable frequences
	for i = 0; i < nSize; i++ {
		uSupport = kit.Kit_TruthSupport(ppCofs[nCBars][i], nVars)
		for k = 0; k < nVars; k++ {
			if (uSupport & (1 << k)) != 0 {
				pFreqs[k]++
			}
		}
	}

	// find common variable order
	for i = 0; i < nSize; i++ {
		kit.Kit_DsdGetSupports(ppNtks[i])
		Lpk_CreateVarOrder(ppNtks[i], pTable[:])
	}
	Lpk_CreateCommonOrder(pTable[:], piCofVar[:], nCBars, pPrios[:], nVars, fVerbose)
	// update priorities with frequences
	for i = 0; i < nVars; i++ {
		pPrios[i] = pPrios[i]*256 + (16-pFreqs[i])*16 + i
	}

	if fVerbose != 0 {
		fmt.Printf("After restructuring with priority:\n")
	}

	// transform all networks according to the variable order
	for i = 0; i < nSize; i++ {
		pTemp = ppNtks[i]
		ppNtks[i] = kit.Kit_DsdShrink(pTemp, pPrios[:])
		kit.Kit_DsdNtkFree(pTemp)
		kit.Kit_DsdGetSupports(ppNtks[i])
		assert(kit.Kit_DsdLitSupport(ppNtks[i], kit.Kit_DsdNtkRootLit(ppNtks[i])) <= 0xFFFF, "kit.Kit_DsdLitSupport(ppNtks[i], kit.Kit_DsdNtkRootLit(ppNtks[i])) <= 0xFFFF")
		// undec nodes should be rotated in such a way that the first input has as many shared inputs as possible
		kit.Kit_DsdRotate(ppNtks[i], pFreqs[:])
		// print the resulting networks
		if fVerbose != 0 {
			fmt.Printf("Cof%d%d: ", nCBars, i)
			kit.Kit_DsdPrint(os.Stdout, ppNtks[i])
		}
	}

	for i = 0; i < nSize; i++ {
		// collect the roots
		pRoot = kit.Kit_DsdNtkRoot(ppNtks[i])
		if pRoot.Type == kit.KIT_DSD_CONST1 {
			piLits[i] = func() int32 {
				if util.Abc_LitIsCompl(int32(ppNtks[i].Root)) != 0 {
					return -2
				}
				return -1
			}()
		} else if pRoot.Type == kit.KIT_DSD_VAR {
			piLits[i] = util.Abc_LitNotCond(int32(kit.Kit_DsdObjFans(pRoot)[0]), util.Abc_LitIsCompl(int32(ppNtks[i].Root)))
		} else {
			piLits[i] = int32(ppNtks[i].Root)
		}
	}

	// recursively construct AIG for mapping
	p.fCofactoring = 1
	pResult = Lpk_MapTreeMulti_rec(p, ppNtks[:], piLits[:], piCofVar[:], nCBars, ppLeaves[:], nVars, pPrios[:])
	p.fCofactoring = 0

	if fVerbose != 0 {
		fmt.Printf("\n")
	}

	// verify the transformations
	nSize = (1 << nCBars)
	for i = 0; i < nSize; i++ {
		kit.Kit_DsdTruth(ppNtks[i], ppCofs[nCBars][i])
	}
	// mux the truth tables
	for k = nCBars - 1; k >= 0; k-- {
		nSize = (1 << k)
		for i = 0; i < nSize; i++ {
			kit.Kit_TruthMuxVar(ppCofs[k][i], ppCofs[k+1][2*i+0], ppCofs[k+1][2*i+1], nVars, piCofVar[k])
		}
	}
	if kit.Kit_TruthIsEqual(pTruth, ppCofs[0][0], nVars) == 0 {
		fmt.Printf("Verification failed.\n")
	}

	// free the networks
	for i = 0; i < 8; i++ {
		if ppNtks[i] != nil {
			kit.Kit_DsdNtkFree(ppNtks[i])
		}
	}
	// Go garbage collection releases the cofactor workspace.

	return pResult
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

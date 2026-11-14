package abci

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/cut"
)

/**CFile****************************************************************

  FileName    [abcLut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Superchoicing for K-LUTs.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcLut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "opt/cut/cut.h"

//ABC_NAMESPACE_IMPL_START

const LARGE_LEVEL = 1000000

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const SCL_LUT_MAX = 6     // the maximum LUT size
const SCL_VARS_MAX = 15   // the maximum number of variables
const SCL_NODE_MAX = 1000 // the maximum number of nodes

type Abc_ManScl_t struct {
	// paramers
	nLutSize    int32 // the LUT size
	nCutSizeMax int32 // the max number of leaves of the cone
	nNodesMax   int32 // the max number of divisors in the cone
	nWords      int32 // the number of machine words in sim info
	// structural representation of the cone
	vLeaves *Vec_Ptr_t          // leaves of the cut
	vVolume *Vec_Ptr_t          // volume of the cut
	pBSet   [SCL_VARS_MAX]int32 // bound set
	// functional representation of the cone
	uTruth []uint32 // truth table of the cone
	// representation of truth tables
	uVars [][]uint32 // elementary truth tables
	uSims [][]uint32 // truth tables of the nodes
	uCofs [][]uint32 // truth tables of the cofactors
}

var s_pLeaves *Vec_Ptr_t = nil

// static Cut_Man_t * Abc_NtkStartCutManForScl( Abc_Ntk_t * pNtk, int nLutSize );

// static Abc_ManScl_t * Abc_ManSclStart( int nLutSize, int nCutSizeMax, int nNodesMax );

// static void Abc_ManSclStop( Abc_ManScl_t * p );

// static void Abc_NodeLutMap( Cut_Man_t * pManCuts, Abc_Obj_t * pObj );

// static Abc_Obj_t * Abc_NodeSuperChoiceLut( Abc_ManScl_t * pManScl, Abc_Obj_t * pObj );

// static int Abc_NodeDecomposeStep( Abc_ManScl_t * pManScl );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Performs superchoicing for K-LUTs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkSuperChoiceLut(pNtk *Abc_Ntk_t, nLutSize int32, nCutSizeMax int32, fVerbose int32) int32 {
	var pProgress *extra.ProgressBar
	var pManCut *Abc_ManCut_t
	var pManScl *Abc_ManScl_t
	var pManCuts *Cut_Man_t
	var pObj, pFanin, pObjTop *Abc_Obj_t
	var i, LevelMax, nNodes int32
	var nNodesTried, nNodesDec, nNodesExist, nNodesUsed int32

	assert(abc.Abc_NtkIsSopLogic(pNtk) != 0, "Abc_NtkIsSopLogic(pNtk)")
	if nLutSize < 3 || nLutSize > SCL_LUT_MAX {
		fmt.Printf("LUT size (%d) does not belong to the interval: 3 <= LUT size <= %d\n", nLutSize, SCL_LUT_MAX)
		return 0
	}
	if nCutSizeMax <= nLutSize || nCutSizeMax > SCL_VARS_MAX {
		fmt.Printf("Cut size (%d) does not belong to the interval: LUT size (%d) < Cut size <= %d\n", nCutSizeMax, nLutSize, SCL_VARS_MAX)
		return 0
	}

	assert(nLutSize <= SCL_LUT_MAX, "nLutSize <= SCL_LUT_MAX")
	assert(nCutSizeMax <= SCL_VARS_MAX, "nCutSizeMax <= SCL_VARS_MAX")
	nNodesTried, nNodesDec, nNodesExist, nNodesUsed = 0, 0, 0, 0

	// set the delays of the CIs
	// inlining of macro Abc_NtkForEachCi
	for i = 0; i < Abc_NtkCiNum(pNtk); i++ {
		pObj = abc.Abc_NtkCi(pNtk, i)
		pObj.Level = 0
	}

	//Abc_NtkLevel( pNtk );

	// start the managers
	pManScl = Abc_ManSclStart(nLutSize, nCutSizeMax, 1000)
	pManCuts = Abc_NtkStartCutManForScl(pNtk, nLutSize)
	pManCut = Abc_NtkManCutStart(nCutSizeMax, 100000, 100000, 100000)
	s_pLeaves = Abc_NtkManCutReadCutSmall(pManCut)
	pManScl.vVolume = Abc_NtkManCutReadVisited(pManCut)

	// process each internal node (assuming topological order of nodes!!!)
	nNodes = abc.Abc_NtkObjNumMax(pNtk)
	pProgress = extra.Extra_ProgressBarStart(os.Stdout, nNodes)
	// inlining of macro Abc_NtkForEachObj
	for i = int32(0); i < abc.Abc_NtkObjNumMax(pNtk); i++ {
		pObj = abc.Abc_NtkObj(pNtk, i)
		if pObj == nil {
			continue
		}
		//        if ( i != nNodes-1 )
		//            continue;
		extra.Extra_ProgressBarUpdate(pProgress, i, "")
		if i >= nNodes {
			break
		}
		if abc.Abc_ObjFaninNum(pObj) != 2 {
			continue
		}
		nNodesTried++

		// map this node using regular cuts
		//        pObj->Level = 0;
		Abc_NodeLutMap(pManCuts, pObj)
		// compute the cut
		pManScl.vLeaves = Abc_NodeFindCut(pManCut, pObj, 0)
		if Vec_PtrSize(pManScl.vLeaves) <= nLutSize {
			continue
		}
		// get the volume of the cut
		if Vec_PtrSize(pManScl.vVolume) > SCL_NODE_MAX {
			continue
		}
		nNodesDec++

		// decompose the cut
		pObjTop = Abc_NodeSuperChoiceLut(pManScl, pObj)
		if pObjTop == nil {
			continue
		}
		nNodesExist++

		// if there is no delay improvement, skip; otherwise, update level
		if pObjTop.Level >= pObj.Level {
			abc.Abc_NtkDeleteObj_rec(pObjTop, 1)
			continue
		}
		pObj.Level = pObjTop.Level
		nNodesUsed++
	}
	extra.Extra_ProgressBarStop(pProgress)

	// delete the managers
	Abc_ManSclStop(pManScl)
	Abc_NtkManCutStop(pManCut)
	cut.Cut_ManStop(pManCuts)

	// get the largest arrival time
	LevelMax = 0
	// inlining of macro Abc_NtkForEachCo
	for i = 0; i < Abc_NtkCoNum(pNtk); i++ {
		pObj = abc.Abc_NtkCo(pNtk, i)
		pFanin = abc.Abc_ObjFanin0(pObj)
		// skip inv/buf
		if abc.Abc_ObjFaninNum(pFanin) == 1 {
			pFanin = abc.Abc_ObjFanin0(pFanin)
		}
		// get the new level
		LevelMax = util.Abc_MaxInt(LevelMax, int32(pFanin.Level))
	}

	if fVerbose != 0 {
		fmt.Printf("Try = %d. Dec = %d. Exist = %d. Use = %d. SUPER = %d levels of %d-LUTs.\n", nNodesTried, nNodesDec, nNodesExist, nNodesUsed, LevelMax, nLutSize)
	}
	//    if ( fVerbose )
	//    printf( "The network is superchoiced for %d levels of %d-LUTs.\n", LevelMax, nLutSize );

	// clean the data field
	// inlining of macro Abc_NtkForEachObj
	for i = int32(0); i < abc.Abc_NtkObjNumMax(pNtk); i++ {
		if pObj = abc.Abc_NtkObj(pNtk, i); pObj != nil {
			pObj.Next = nil
		}
	}

	// check
	if abc.Abc_NtkCheck(pNtk) == 0 {
		fmt.Printf("Abc_NtkSuperChoiceLut: The network check has failed.\n")
		return 0
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Performs LUT mapping of the node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeLutMap(pManCuts *Cut_Man_t, pObj *Abc_Obj_t) {
	var pCut *Cut_Cut_t
	var pFanin *Abc_Obj_t
	var i, DelayMax int32
	pCut = Abc_NodeGetCutsRecursive(pManCuts, pObj, 0, 0)
	assert(pCut != nil, "pCut != nil")
	assert(pObj.Level == 0, "pObj.Level == 0")
	// go through the cuts
	pObj.Level = LARGE_LEVEL
	for pCut = pCut.Next; pCut != nil; pCut = pCut.Next {
		DelayMax = 0
		for i = 0; i < int32(len(pCut.Leaves())); i++ {
			pFanin = abc.Abc_NtkObj(pObj.Network(), pCut.Leaves()[i])
			//            assert( Abc_ObjIsCi(pFanin) || pFanin->Level > 0 ); // should hold if node ordering is topological
			if DelayMax < int32(pFanin.Level) {
				DelayMax = int32(pFanin.Level)
			}
		}
		if int32(pObj.Level) > DelayMax {
			pObj.Level = int(DelayMax)
		}
	}
	assert(pObj.Level < LARGE_LEVEL, "pObj.Level < LARGE_LEVEL")
	pObj.Level++
	// printf( "%d(%d) ", pObj->Id, pObj->Level );
}

/*
*Function*************************************************************

	Synopsis    [Starts the cut manager for rewriting.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Abc_NtkStartCutManForScl_Params Cut_Params_t
var spStatic_Abc_NtkStartCutManForScl_pParams *Cut_Params_t = &spStatic_Abc_NtkStartCutManForScl_Params

func Abc_NtkStartCutManForScl(pNtk *Abc_Ntk_t, nLutSize int32) *Cut_Man_t {

	var pManCut *Cut_Man_t
	var pObj *Abc_Obj_t
	var i int32
	// start the cut manager
	*spStatic_Abc_NtkStartCutManForScl_pParams = Cut_Params_t{}
	spStatic_Abc_NtkStartCutManForScl_pParams.NVarsMax = nLutSize
	spStatic_Abc_NtkStartCutManForScl_pParams.NKeepMax = 500
	spStatic_Abc_NtkStartCutManForScl_pParams.FFilter = 1
	spStatic_Abc_NtkStartCutManForScl_pParams.NIdsMax = abc.Abc_NtkObjNumMax(pNtk)
	pManCut = cut.Cut_ManStart(spStatic_Abc_NtkStartCutManForScl_pParams)
	if spStatic_Abc_NtkStartCutManForScl_pParams.FDrop != 0 {
		cut.Cut_ManSetFanoutCounts(pManCut, abc.Abc_NtkFanoutCounts(pNtk).Array())
	}
	// set cuts for PIs
	// inlining of macro Abc_NtkForEachCi
	for i = 0; i < Abc_NtkCiNum(pNtk); i++ {
		pObj = abc.Abc_NtkCi(pNtk, i)
		if abc.Abc_ObjFanoutNum(pObj) > 0 {
			cut.Cut_NodeSetTriv(pManCut, int32(pObj.ID))
		}
	}
	return pManCut
}

/*
*Function*************************************************************

	Synopsis    [Starts the manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_ManSclStart(nLutSize int32, nCutSizeMax int32, nNodesMax int32) *Abc_ManScl_t {
	p := &Abc_ManScl_t{}
	if nLutSize < 0 || nLutSize > SCL_LUT_MAX || nCutSizeMax < 0 || nNodesMax < 0 {
		return nil
	}
	p.nLutSize = nLutSize
	p.nCutSizeMax = nCutSizeMax
	p.nNodesMax = nNodesMax
	p.nWords = extra.Extra_TruthWordNum(nCutSizeMax)
	words := int(p.nWords)
	p.uVars = make([][]uint32, int(nCutSizeMax))
	p.uSims = make([][]uint32, int(nNodesMax))
	p.uCofs = make([][]uint32, 2<<uint(nLutSize))
	for i := range p.uVars {
		p.uVars[i] = make([]uint32, words)
	}
	for i := range p.uSims {
		p.uSims[i] = make([]uint32, words)
	}
	for i := range p.uCofs {
		p.uCofs[i] = make([]uint32, words)
	}
	// assign elementary truth tables
	for k := int32(0); k < p.nCutSizeMax; k++ {
		for bit := int32(0); bit < p.nWords*32; bit++ {
			if bit&(1<<uint(k)) != 0 {
				p.uVars[k][bit>>5] |= uint32(1) << uint(bit&31)
			}
		}
	}
	// other data structures
	//    p->vBound = Vec_IntAlloc( nCutSizeMax );
	return p
}

/*
*Function*************************************************************

	Synopsis    [Stops the manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_ManSclStop(p *Abc_ManScl_t) {
	//    Vec_IntFree( p->vBound );
	p.uVars = nil
	p.uSims = nil
	p.uCofs = nil
	p = nil
}

/*
*Function*************************************************************

	Synopsis    [Performs superchoicing for one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeSuperChoiceTruth(pManScl *Abc_ManScl_t) []uint32 {
	var pObj *Abc_Obj_t
	var puData0, puData1, puData []uint32
	var pSop string
	var i, k int32
	sims := make(map[*Abc_Obj_t][]uint32, Vec_PtrSize(pManScl.vLeaves)+Vec_PtrSize(pManScl.vVolume))
	// set elementary truth tables
	// inlining of macro Vec_PtrForEachEntry
	for i = int32(0); i < Vec_PtrSize(pManScl.vLeaves); i++ {
		pObj = Vec_PtrEntry(pManScl.vLeaves, i).(*Abc_Obj_t)
		sims[pObj] = pManScl.uVars[i]
	}
	// compute truth tables for internal nodes
	// inlining of macro Vec_PtrForEachEntry
	for i = int32(0); i < Vec_PtrSize(pManScl.vVolume); i++ {
		pObj = Vec_PtrEntry(pManScl.vVolume, i).(*Abc_Obj_t)
		// set storage for the node's simulation info
		sims[pObj] = pManScl.uSims[i]
		// get pointer to the simulation info
		puData = pManScl.uSims[i]
		puData0 = sims[abc.Abc_ObjFanin0(pObj)]
		puData1 = sims[abc.Abc_ObjFanin1(pObj)]
		// simulate
		pSop, _ = pObj.Data.(string)
		if pSop[0] == '0' && pSop[1] == '0' {
			for k = 0; k < pManScl.nWords; k++ {
				puData[k] = ~puData0[k] & ~puData1[k]
			}
		} else if pSop[0] == '0' {
			for k = 0; k < pManScl.nWords; k++ {
				puData[k] = ~puData0[k] & puData1[k]
			}
		} else if pSop[1] == '0' {
			for k = 0; k < pManScl.nWords; k++ {
				puData[k] = puData0[k] & ~puData1[k]
			}
		} else {
			for k = 0; k < pManScl.nWords; k++ {
				puData[k] = puData0[k] & puData1[k]
			}
		}
	}
	return puData
}

/*
*Function*************************************************************

	Synopsis    [Performs superchoicing for one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeSuperChoiceCollect2_rec(pObj *Abc_Obj_t, vVolume *Vec_Ptr_t) {
	if pObj.MarkC {
		return
	}
	pObj.MarkC = true
	assert(abc.Abc_ObjFaninNum(pObj) == 2, "abc.Abc_ObjFaninNum(pObj) == 2")
	Abc_NodeSuperChoiceCollect2_rec(abc.Abc_ObjFanin0(pObj), vVolume)
	Abc_NodeSuperChoiceCollect2_rec(abc.Abc_ObjFanin1(pObj), vVolume)
	Vec_PtrPush(vVolume, pObj)
}

/*
*Function*************************************************************

	Synopsis    [Performs superchoicing for one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeSuperChoiceCollect2(pRoot *Abc_Obj_t, vLeaves *Vec_Ptr_t, vVolume *Vec_Ptr_t) {
	var pObj *Abc_Obj_t
	var i int32
	// inlining of macro Vec_PtrForEachEntry
	for i = int32(0); i < Vec_PtrSize(vLeaves); i++ {
		pObj = Vec_PtrEntry(vLeaves, i).(*Abc_Obj_t)
		pObj.MarkC = true
	}
	Vec_PtrClear(vVolume)
	Abc_NodeSuperChoiceCollect2_rec(pRoot, vVolume)
	// inlining of macro Vec_PtrForEachEntry
	for i = int32(0); i < Vec_PtrSize(vLeaves); i++ {
		pObj = Vec_PtrEntry(vLeaves, i).(*Abc_Obj_t)
		pObj.MarkC = false
	}
	// inlining of macro Vec_PtrForEachEntry
	for i = int32(0); i < Vec_PtrSize(vVolume); i++ {
		pObj = Vec_PtrEntry(vVolume, i).(*Abc_Obj_t)
		pObj.MarkC = false
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs superchoicing for one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeSuperChoiceCollect_rec(pObj *Abc_Obj_t, vLeaves *Vec_Ptr_t, vVolume *Vec_Ptr_t) {
	if pObj.MarkB {
		Vec_PtrPush(vLeaves, pObj)
		pObj.MarkB = false
	}
	if pObj.MarkC {
		return
	}
	pObj.MarkC = true
	assert(abc.Abc_ObjFaninNum(pObj) == 2, "abc.Abc_ObjFaninNum(pObj) == 2")
	Abc_NodeSuperChoiceCollect_rec(abc.Abc_ObjFanin0(pObj), vLeaves, vVolume)
	Abc_NodeSuperChoiceCollect_rec(abc.Abc_ObjFanin1(pObj), vLeaves, vVolume)
	Vec_PtrPush(vVolume, pObj)
}

/*
*Function*************************************************************

	Synopsis    [Performs superchoicing for one node.]

	Description [Orders the leaves topologically.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeSuperChoiceCollect(pRoot *Abc_Obj_t, vLeaves *Vec_Ptr_t, vVolume *Vec_Ptr_t) {
	var pObj *Abc_Obj_t
	var i, nLeaves int32
	nLeaves = Vec_PtrSize(vLeaves)
	// inlining of macro Vec_PtrForEachEntry
	for i = int32(0); i < Vec_PtrSize(vLeaves); i++ {
		pObj = Vec_PtrEntry(vLeaves, i).(*Abc_Obj_t)
		pObj.MarkB, pObj.MarkC = true, true
	}
	Vec_PtrClear(vVolume)
	Vec_PtrClear(vLeaves)
	Abc_NodeSuperChoiceCollect_rec(pRoot, vLeaves, vVolume)
	assert(Vec_PtrSize(vLeaves) == nLeaves, "Vec_PtrSize(vLeaves) == nLeaves")
	// inlining of macro Vec_PtrForEachEntry
	for i = int32(0); i < Vec_PtrSize(vLeaves); i++ {
		pObj = Vec_PtrEntry(vLeaves, i).(*Abc_Obj_t)
		pObj.MarkC = false
	}
	// inlining of macro Vec_PtrForEachEntry
	for i = int32(0); i < Vec_PtrSize(vVolume); i++ {
		pObj = Vec_PtrEntry(vVolume, i).(*Abc_Obj_t)
		pObj.MarkC = false
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs superchoicing for one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeLeavesRemove(vLeaves *Vec_Ptr_t, uPhase uint32, nVars int32) {
	var i int32
	for i := nVars - 1; i >= 0; i-- {
		if uPhase&(1<<uint(i)) != 0 {
			vec.Vec_PtrRemove(vLeaves, Vec_PtrEntry(vLeaves, i))
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs superchoicing for one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeGetLevel(pObj *Abc_Obj_t) int32 {
	var pFanin *Abc_Obj_t
	var i, Level int32
	Level = 0
	// inlining of macro Abc_ObjForEachFanin
	for i = int32(0); i < abc.Abc_ObjFaninNum(pObj); i++ {
		pFanin = abc.Abc_ObjFanin(pObj, i)
		Level = util.Abc_MaxInt(Level, int32(pFanin.Level))
	}
	return Level + 1
}

/*
*Function*************************************************************

	Synopsis    [Performs superchoicing for one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeSuperChoiceLut(p *Abc_ManScl_t, pObj *Abc_Obj_t) *Abc_Obj_t {
	var pFanin, pObjNew *Abc_Obj_t
	var i, nVars, uSupport, nSuppVars int32
	// collect the cone using DFS (excluding leaves)
	Abc_NodeSuperChoiceCollect2(pObj, p.vLeaves, p.vVolume)
	assert(Vec_PtrEntry(p.vVolume, Vec_PtrSize(p.vVolume)-1) == pObj, "Vec_PtrEntry(p.vVolume, Vec_PtrSize(p.vVolume)-1) == pObj")
	// compute the truth table
	p.uTruth = Abc_NodeSuperChoiceTruth(p)
	// get the support of this truth table
	nVars = Vec_PtrSize(p.vLeaves)
	uSupport = extra.Extra_TruthSupport(p.uTruth, nVars)
	nSuppVars = extra.Extra_WordCountOnes(uint32(uSupport))
	assert(nSuppVars <= nVars, "nSuppVars <= nVars")
	if nSuppVars == 0 {
		pObj.Level = 0
		return nil
	}
	if nSuppVars == 1 {
		// find the variable
		for i = 0; i < nVars; i++ {
			if uSupport&(1<<uint(i)) != 0 {
				break
			}
		}
		assert(i < nVars, "i < nVars")
		pFanin = Vec_PtrEntry(p.vLeaves, i).(*Abc_Obj_t)
		pObj.Level = pFanin.Level
		return nil
	}
	// support-minimize the truth table
	if nSuppVars != nVars {
		extra.Extra_TruthShrink(p.uCofs[0], p.uTruth, nSuppVars, nVars, uint32(uSupport))
		extra.Extra_TruthCopy(p.uTruth, p.uCofs[0], nVars)
		Abc_NodeLeavesRemove(p.vLeaves, ((1<<nVars)-1)&~uSupport, nVars)
	}
	//    return NULL;
	// decompose the truth table recursively
	for Vec_PtrSize(p.vLeaves) > p.nLutSize {
		if Abc_NodeDecomposeStep(p) == 0 {
			// inlining of macro Vec_PtrForEachEntry
			for i = int32(0); i < Vec_PtrSize(p.vLeaves); i++ {
				pFanin = Vec_PtrEntry(p.vLeaves, i).(*Abc_Obj_t)
				if pFanin.Kind == abc.ObjectNode && abc.Abc_ObjFanoutNum(pFanin) == 0 {
					abc.Abc_NtkDeleteObj_rec(pFanin, 1)
				}
			}
			return nil
		}
	}
	// create the topmost node
	pObjNew = abc.Abc_NtkCreateNode(pObj.Network())
	// inlining of macro Vec_PtrForEachEntry
	for i = int32(0); i < Vec_PtrSize(p.vLeaves); i++ {
		pFanin = Vec_PtrEntry(p.vLeaves, i).(*Abc_Obj_t)
		abc.Abc_ObjAddFanin(pObjNew, pFanin)
	}
	// create the function
	pObjNew.Data = abc.Abc_SopCreateFromTruth(pObj.Network().FunctionManager.(*mem.Mem_Flex_t), Vec_PtrSize(p.vLeaves), p.uTruth)
	pObjNew.Level = int(Abc_NodeGetLevel(pObjNew))
	return pObjNew
}

/*
*Function*************************************************************

	Synopsis    [Procedure used for sorting the nodes in increasing order of levels.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeCompareLevelsInc(pp1 *int32, pp2 *int32) int32 {
	var pNode1, pNode2 *Abc_Obj_t
	pNode1 = Vec_PtrEntry(s_pLeaves, *pp1).(*Abc_Obj_t)
	pNode2 = Vec_PtrEntry(s_pLeaves, *pp2).(*Abc_Obj_t)
	if pNode1.Level < pNode2.Level {
		return -1
	}
	if pNode1.Level > pNode2.Level {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Selects the earliest arriving nodes from the array.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeDecomposeSort(pLeaves []any, nVars int32, pBSet []int32, nLutSize int32) {
	var pTemp [SCL_VARS_MAX]*Abc_Obj_t
	var i, k, kBest, LevelMin int32
	assert(nLutSize < nVars, "nLutSize < nVars")
	assert(nVars <= SCL_VARS_MAX, "nVars <= SCL_VARS_MAX")
	// copy nodes into the internal storage
	//    printf( "(" );
	for i = 0; i < nVars; i++ {
		pTemp[i] = pLeaves[i].(*Abc_Obj_t)
		//        printf( " %d", pLeaves[i]->Level );
	}
	//    printf( " )\n" );
	// choose one node at a time
	for i = 0; i < nLutSize; i++ {
		kBest = -1
		LevelMin = LARGE_LEVEL
		for k = 0; k < nVars; k++ {
			if pTemp[k] && LevelMin > int32(pTemp[k].Level) {
				LevelMin = int32(pTemp[k].Level)
				kBest = k
			}
		}
		pBSet[i] = kBest
		pTemp[kBest] = nil
	}
}

/*
*Function*************************************************************

	Synopsis    [Performs superchoicing for one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Abc_NodeDecomposeStep_pCofClasses [1 << SCL_LUT_MAX][1 << SCL_LUT_MAX]byte
var spStatic_Abc_NodeDecomposeStep_nCofClasses [1 << SCL_LUT_MAX]byte

func Abc_NodeDecomposeStep(p *Abc_ManScl_t) int32 {

	var pNtk *Abc_Ntk_t
	var pObjNew *Abc_Obj_t
	var pFanin *Abc_Obj_t
	var pNodesNew [SCL_LUT_MAX]*Abc_Obj_t
	var pTruthCof, pTruthClass, pTruth []uint32
	var uPhase uint32
	var i, k, c, v, w, nVars, nVarsNew, nClasses, nCofs int32
	// set the network
	pNtk = Vec_PtrEntry(p.vLeaves, 0).(*Abc_Obj_t).Network()
	// find the earliest nodes
	nVars = Vec_PtrSize(p.vLeaves)
	assert(nVars > p.nLutSize, "nVars > p.nLutSize")
	/*
	   for ( v = 0; v < nVars; v++ )
	       p->pBSet[v] = v;
	   qsort( (void *)p->pBSet, (size_t)nVars, sizeof(int),
	           (int (*)(const void *, const void *)) Abc_NodeCompareLevelsInc );
	*/
	Abc_NodeDecomposeSort(Vec_PtrArray(p.vLeaves), Vec_PtrSize(p.vLeaves), p.pBSet[:], p.nLutSize)
	assert(Vec_PtrEntry(p.vLeaves, p.pBSet[0]).(*Abc_Obj_t).Level <= Vec_PtrEntry(p.vLeaves, p.pBSet[1]).(*Abc_Obj_t).Level, "leaves are level-sorted")
	// cofactor w.r.t. the selected variables
	extra.Extra_TruthCopy(p.uCofs[1], p.uTruth, nVars)
	c = 2
	for v = 0; v < p.nLutSize; v++ {
		for k = 0; k < (1 << v); k++ {
			extra.Extra_TruthCopy(p.uCofs[c], p.uCofs[c/2], nVars)
			extra.Extra_TruthCopy(p.uCofs[c+1], p.uCofs[c/2], nVars)
			extra.Extra_TruthCofactor0(p.uCofs[c], nVars, p.pBSet[v])
			extra.Extra_TruthCofactor1(p.uCofs[c+1], nVars, p.pBSet[v])
			c += 2
		}
	}
	assert(c == (2<<p.nLutSize), "c == (2 << p.nLutSize)")
	// count unique cofactors
	nClasses = 0
	nCofs = (1 << p.nLutSize)
	for i = 0; i < nCofs; i++ {
		pTruthCof = p.uCofs[nCofs+i]
		for k = 0; k < nClasses; k++ {
			pTruthClass = p.uCofs[nCofs+int32(spStatic_Abc_NodeDecomposeStep_pCofClasses[k][0])]
			if extra.Extra_TruthIsEqual(pTruthCof, pTruthClass, nVars) != 0 {
				spStatic_Abc_NodeDecomposeStep_pCofClasses[k][spStatic_Abc_NodeDecomposeStep_nCofClasses[k]] = byte(i)
				spStatic_Abc_NodeDecomposeStep_nCofClasses[k]++
				break
			}
		}
		if k != nClasses {
			continue
		}
		// not found
		spStatic_Abc_NodeDecomposeStep_pCofClasses[nClasses][0] = byte(i)
		spStatic_Abc_NodeDecomposeStep_nCofClasses[nClasses] = 1
		nClasses++
		if nClasses > nCofs/2 {
			return 0
		}
	}
	// the number of cofactors is acceptable
	nVarsNew = util.Abc_Base2Log(uint32(nClasses))
	assert(nVarsNew < p.nLutSize, "nVarsNew < p.nLutSize")
	// create the remainder truth table
	// for each class of cofactors, multiply cofactor truth table by its code
	extra.Extra_TruthClear(p.uTruth, nVars)
	for k = 0; k < nClasses; k++ {
		pTruthClass = p.uCofs[nCofs+int32(spStatic_Abc_NodeDecomposeStep_pCofClasses[k][0])]
		for v = 0; v < nVarsNew; v++ {
			if k&(1<<uint(v)) != 0 {
				extra.Extra_TruthAnd(pTruthClass, pTruthClass, p.uVars[p.pBSet[v]], nVars)
			} else {
				extra.Extra_TruthSharp(pTruthClass, pTruthClass, p.uVars[p.pBSet[v]], nVars)
			}
		}
		extra.Extra_TruthOr(p.uTruth, p.uTruth, pTruthClass, nVars)
	}
	// create nodes
	pTruth = p.uCofs[0]
	for v = 0; v < nVarsNew; v++ {
		extra.Extra_TruthClear(pTruth, p.nLutSize)
		for k = 0; k < nClasses; k++ {
			if k&(1<<uint(v)) != 0 {
				for i = 0; i < spStatic_Abc_NodeDecomposeStep_nCofClasses[k]; i++ {
					pTruthCof = p.uCofs[1]
					extra.Extra_TruthFill(pTruthCof, p.nLutSize)
					for w = 0; w < p.nLutSize; w++ {
						if uint32(spStatic_Abc_NodeDecomposeStep_pCofClasses[k][i])&(1<<uint(p.nLutSize-1-w)) != 0 {
							extra.Extra_TruthAnd(pTruthCof, pTruthCof, p.uVars[w], p.nLutSize)
						} else {
							extra.Extra_TruthSharp(pTruthCof, pTruthCof, p.uVars[w], p.nLutSize)
						}
					}
					extra.Extra_TruthOr(pTruth, pTruth, pTruthCof, p.nLutSize)
				}
			}
		}
		// implement the node
		pObjNew = abc.Abc_NtkCreateNode(pNtk)
		for i = 0; i < p.nLutSize; i++ {
			pFanin = Vec_PtrEntry(p.vLeaves, p.pBSet[i]).(*Abc_Obj_t)
			abc.Abc_ObjAddFanin(pObjNew, pFanin)
		}
		// create the function
		pObjNew.Data = abc.Abc_SopCreateFromTruth(pNtk.FunctionManager.(*mem.Mem_Flex_t), p.nLutSize, pTruth) // need ISOP
		pObjNew.Level = int(Abc_NodeGetLevel(pObjNew))
		pNodesNew[v] = pObjNew
	}
	// put the new nodes back into the list
	for v = 0; v < nVarsNew; v++ {
		Vec_PtrWriteEntry(p.vLeaves, p.pBSet[v], pNodesNew[v])
	}
	// compute the variables that should be removed
	uPhase = 0
	for v = nVarsNew; v < p.nLutSize; v++ {
		uPhase |= (1 << p.pBSet[v])
	}
	// remove entries from the array
	Abc_NodeLeavesRemove(p.vLeaves, uPhase, nVars)
	// update truth table
	extra.Extra_TruthShrink(p.uCofs[0], p.uTruth, nVars-p.nLutSize+nVarsNew, nVars, ((1<<uint(nVars))-1)&^uPhase)
	extra.Extra_TruthCopy(p.uTruth, p.uCofs[0], nVars)
	assert(extra.Extra_TruthVarInSupport(p.uTruth, nVars, nVars-p.nLutSize+nVarsNew) == 0, "removed variable absent from support")
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

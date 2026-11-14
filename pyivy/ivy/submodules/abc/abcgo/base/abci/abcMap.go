package abci

import aig_gia "github.com/glycerine/abcgo/aig/gia"

import base_main "github.com/glycerine/abcgo/base/main"

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/mapper"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/opt/sim"
)

type Map_Man_t = mapper.Map_Man_t
type Map_Node_t = mapper.Map_Node_t
type Map_Time_t = mapper.Map_Time_t
type Map_Cut_t = mapper.Map_Cut_t
type Map_Super_t = mapper.Map_Super_t
type Mio_Library_t = mio.Mio_Library_t
type Mio_Gate_t = mio.Mio_Gate_t
type Map_SuperLib_t = mapper.Map_SuperLib_t
type ProgressBar = extra.ProgressBar

func mapMaxFloat(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func mapCopyNode(object *Abc_Obj_t) *Map_Node_t {
	node, _ := object.CopyData.(*Map_Node_t)
	return node
}
func mapGenlib() *mio.Mio_Library_t {
	lib, _ := base_main.Abc_FrameReadLibGen().(*mio.Mio_Library_t)
	return lib
}
func mapSuperlib() *mapper.Map_SuperLib_t {
	lib, _ := base_main.Abc_FrameReadLibSuper().(*mapper.Map_SuperLib_t)
	return lib
}
func mapScllib() *scl.SC_Lib { lib, _ := base_main.Abc_FrameReadLibScl().(*scl.SC_Lib); return lib }

/**CFile****************************************************************

  FileName    [abcMap.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Interface with the SC mapping package.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcMap.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "map/mio/mio.h"
//#include "map/mapper/mapper.h"
//#include "misc/util/utilNam.h"
//#include "map/scl/sclCon.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Map_Man_t *  Abc_NtkToMap( Abc_Ntk_t * pNtk, double DelayTarget, int fRecovery, float * pSwitching, int fVerbose );

// static Abc_Ntk_t *  Abc_NtkFromMap( Map_Man_t * pMan, Abc_Ntk_t * pNtk, int fUseBuffs );

// static Abc_Obj_t *  Abc_NodeFromMap_rec( Abc_Ntk_t * pNtkNew, Map_Node_t * pNodeMap, int fPhase );

// static Abc_Obj_t *  Abc_NodeFromMapPhase_rec( Abc_Ntk_t * pNtkNew, Map_Node_t * pNodeMap, int fPhase );

// static Abc_Ntk_t *  Abc_NtkFromMapSuperChoice( Map_Man_t * pMan, Abc_Ntk_t * pNtk );

// static void         Abc_NodeSuperChoice( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pNode );

// static void         Abc_NodeFromMapCutPhase( Abc_Ntk_t * pNtkNew, Map_Cut_t * pCut, int fPhase );

// static Abc_Obj_t *  Abc_NodeFromMapSuperChoice_rec( Abc_Ntk_t * pNtkNew, Map_Super_t * pSuper, Abc_Obj_t * pNodePis[], int nNodePis );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Interface with the mapping package.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
var spStatic_Abc_NtkMap_fUseMulti int32

func Abc_NtkMap(pNtk *Abc_Ntk_t, DelayTarget float64, AreaMulti float64, DelayMulti float64, LogFan float32, Slew float32, Gain float32, nGatesMin int32, fRecovery int32, fSwitching int32, fSkipFanout int32, fUseProfile int32, fUseBuffs int32, fVerbose int32) *Abc_Ntk_t {

	var fShowSwitching int32 = 1
	var pNtkNew *Abc_Ntk_t
	var pMan *Map_Man_t
	var vSwitching []float32
	var pSwitching *float32 = nil
	var clkTotal abctime = Abc_Clock()
	pLib := mapGenlib()

	assert(baseabc.Abc_NtkIsStrash(pNtk) != 0, "baseabc.Abc_NtkIsStrash(pNtk)")
	// derive library from SCL
	// if the library is created here, it will be deleted when pSuperLib is deleted in mapper.Map_SuperLibFree()
	if mapScllib() != nil && scl.Abc_SclHasDelayInfo(mapScllib()) {
		var err error
		if pLib != nil && (mio.Mio_LibraryHasProfile(pLib) != 0) {
			pLib, err = scl.Abc_SclDeriveGenlib(mapScllib(), pLib, Slew, Gain, int(nGatesMin), fVerbose)
		} else {
			pLib, err = scl.Abc_SclDeriveGenlib(mapScllib(), nil, Slew, Gain, int(nGatesMin), fVerbose)
		}
		if err != nil {
			return nil
		}
		if mapGenlib() != nil {
			mio.Mio_LibraryTransferDelays(mapGenlib(), pLib)
			mio.Mio_LibraryTransferProfile(pLib, mapGenlib())
		}
		// remove supergate library
		mapper.Map_SuperLibFree(mapSuperlib())
		base_main.Abc_FrameSetLibSuper(nil)
	}
	// quit if there is no library
	if pLib == nil {
		fmt.Printf("The current library is not available.\n")
		return nil
	}
	if AreaMulti != 0.0 {
		spStatic_Abc_NtkMap_fUseMulti = 1
		fmt.Printf("The cell areas are multiplied by the factor: <num_fanins> ^ (%.2f).\n", AreaMulti)
	}
	if DelayMulti != 0.0 {
		spStatic_Abc_NtkMap_fUseMulti = 1
		fmt.Printf("The cell delays are multiplied by the factor: <num_fanins> ^ (%.2f).\n", DelayMulti)
	}

	// penalize large gates by increasing their area
	if AreaMulti != 0.0 {
		mio.Mio_LibraryMultiArea(pLib, AreaMulti)
	}
	if DelayMulti != 0.0 {
		mio.Mio_LibraryMultiDelay(pLib, DelayMulti)
	}

	// derive the supergate library
	if spStatic_Abc_NtkMap_fUseMulti != 0 || mapSuperlib() == nil {
		if fVerbose != 0 {
			fmt.Printf("Converting \"%s\" into supergate library \"%s\".\n", mio.Mio_LibraryReadName(pLib), extra.Extra_FileNameGenericAppend(mio.Mio_LibraryReadName(pLib), ".super"))
		}
		// compute supergate library to be used for mapping
		if mio.Mio_LibraryHasProfile(pLib) != 0 {
			fmt.Printf("Abc_NtkMap(): Genlib library has profile.\n")
		}
		mapper.Map_SuperLibDeriveFromGenlib(pLib, fVerbose)
	}

	// return the library to normal
	if AreaMulti != 0.0 {
		mio.Mio_LibraryMultiArea(mapGenlib(), -AreaMulti)
	}
	if DelayMulti != 0.0 {
		mio.Mio_LibraryMultiDelay(mapGenlib(), -DelayMulti)
	}

	// print a warning about choice nodes
	if fVerbose != 0 && baseabc.Abc_NtkGetChoiceNum(pNtk) != 0 {
		fmt.Printf("Performing mapping with choices.\n")
	}

	// compute switching activity
	fShowSwitching |= fSwitching
	if fShowSwitching != 0 {
		// Sim_NtkComputeSwitching is provided by the simulation package.
		vSwitching = sim.Sim_NtkComputeSwitching(baseabc.Abc_NtkToDar(pNtk, 0, 0), 4096)
		if len(vSwitching) > 0 {
			pSwitching = &vSwitching[0]
		}
	}

	// perform the mapping
	pMan = Abc_NtkToMap(pNtk, DelayTarget, fRecovery, pSwitching, fVerbose)
	if pMan == nil {
		return nil
	}
	mapper.Map_ManSetSwitching(pMan, fSwitching)
	mapper.Map_ManSetSkipFanout(pMan, fSkipFanout)
	if fUseProfile != 0 {
		mapper.Map_ManSetUseProfile(pMan)
	}
	if LogFan != 0 {
		mapper.Map_ManCreateNodeDelays(pMan, int32(LogFan))
	}
	if (mapper.Map_Mapping(pMan) == 0) {
		mapper.Map_ManFree(pMan)
		return nil
	}
	//    mapper.Map_ManPrintStatsToFile( pNtk->pSpec, mapper.Map_ManReadAreaFinal(pMan), mapper.Map_ManReadRequiredGlo(pMan), Abc_Clock()-clk );

	// reconstruct the network after mapping (use buffers when user requested or in the area mode)
	pNtkNew = Abc_NtkFromMap(pMan, pNtk, fUseBuffs != 0 || (DelayTarget == float64(ABC_INFINITY)))
	if mio.Mio_LibraryHasProfile(pLib) != 0 {
		mio.Mio_LibraryTransferProfile2(mapGenlib(), pLib)
	}
	mapper.Map_ManFree(pMan)
	if pNtkNew == nil {
		return nil
	}

	if pNtk.Exdc != nil {
		pNtkNew.Exdc = baseabc.Abc_NtkDup(pNtk.Exdc)
	}
	if fVerbose != 0 {
		// inlining of macro ABC_PRT
		Abc_Print(1, "%s =", ("Total runtime"))
		Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))))
	}

	// make sure that everything is okay
	if baseabc.Abc_NtkCheck(pNtkNew) == 0 {
		fmt.Printf("Abc_NtkMap: The network check has failed.\n")
		Abc_NtkDelete(pNtkNew)
		return nil
	}
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Load the network into manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMapCopyCiArrival(pNtk *Abc_Ntk_t, ppTimes *Abc_Time_t) []Map_Time_t {
	var i int32
	// inlining of macro ABC_CALLOC
	p := make([]Map_Time_t, int(Abc_NtkCiNum(pNtk)))
	for i = 0; i < Abc_NtkCiNum(pNtk); i++ {
		p[i].Fall = unsafe.Slice(ppTimes, int(Abc_NtkCiNum(pNtk)))[i].Fall
		p[i].Rise = unsafe.Slice(ppTimes, int(Abc_NtkCiNum(pNtk)))[i].Rise
		p[i].Worst = mapMaxFloat(p[i].Fall, p[i].Rise)
	}
	ppTimes = nil
	return p
}
func Abc_NtkMapCopyCoRequired(pNtk *Abc_Ntk_t, ppTimes *Abc_Time_t) []Map_Time_t {
	var i int32
	// inlining of macro ABC_CALLOC
	p := make([]Map_Time_t, int(Abc_NtkCoNum(pNtk)))
	for i = 0; i < Abc_NtkCoNum(pNtk); i++ {
		p[i].Fall = unsafe.Slice(ppTimes, int(Abc_NtkCoNum(pNtk)))[i].Fall
		p[i].Rise = unsafe.Slice(ppTimes, int(Abc_NtkCoNum(pNtk)))[i].Rise
		p[i].Worst = mapMaxFloat(p[i].Fall, p[i].Rise)
	}
	ppTimes = nil
	return p
}

/*
*Function*************************************************************

	Synopsis    [Load the network into manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkMapCopyCiArrivalCon(pNtk *Abc_Ntk_t) []Map_Time_t {
	var i int32
	// inlining of macro ABC_CALLOC
	p := make([]Map_Time_t, int(Abc_NtkCiNum(pNtk)))
	for i = 0; i < Abc_NtkCiNum(pNtk); i++ {
		p[i].Fall = scl.Scl_Int2Flt(scl.Scl_ConGetInArr(int(i)))
		p[i].Rise = p[i].Fall
		p[i].Worst = p[i].Fall
	}
	return p
}
func Abc_NtkMapCopyCoRequiredCon(pNtk *Abc_Ntk_t) []Map_Time_t {
	var i int32
	// inlining of macro ABC_CALLOC
	p := make([]Map_Time_t, int(Abc_NtkCoNum(pNtk)))
	for i = 0; i < Abc_NtkCoNum(pNtk); i++ {
		p[i].Fall = scl.Scl_Int2Flt(scl.Scl_ConGetOutReq(int(i)))
		p[i].Rise = p[i].Fall
		p[i].Worst = p[i].Fall
	}
	return p
}

/*
*Function*************************************************************

	Synopsis    [Load the network into manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkToMap(pNtk *Abc_Ntk_t, DelayTarget float64, fRecovery int32, pSwitching *float32, fVerbose int32) *Map_Man_t {
	var pMan *Map_Man_t
	var pNodeMap *Map_Node_t
	var vNodes []*Abc_Obj_t
	var pNode, pFanin, pPrev *Abc_Obj_t
	var i int32

	assert(baseabc.Abc_NtkIsStrash(pNtk) != 0, "baseabc.Abc_NtkIsStrash(pNtk)")

	// start the mapping manager and set its parameters
	pMan = mapper.Map_ManCreate(Abc_NtkPiNum(pNtk)+Abc_NtkLatchNum(pNtk)-baseabc.Abc_NtkBarBufNum(pNtk), Abc_NtkPoNum(pNtk)+Abc_NtkLatchNum(pNtk)-baseabc.Abc_NtkBarBufNum(pNtk), fVerbose)
	if pMan == nil {
		return nil
	}
	mapper.Map_ManSetAreaRecovery(pMan, fRecovery)
	mapper.Map_ManSetOutputNames(pMan, baseabc.Abc_NtkCollectCioNames(pNtk, 1))
	mapper.Map_ManSetDelayTarget(pMan, float32(DelayTarget))

	// set arrival and requireds
	if mapper.Scl_ConIsRunning() && scl.Scl_ConHasInArrs() {
		mapper.Map_ManSetInputArrivals(pMan, Abc_NtkMapCopyCiArrivalCon(pNtk))
	} else {
		mapper.Map_ManSetInputArrivals(pMan, Abc_NtkMapCopyCiArrival(pNtk, Abc_NtkGetCiArrivalTimes(pNtk)))
	}
	if mapper.Scl_ConIsRunning() && mapper.Scl_ConHasOutReqs() {
		mapper.Map_ManSetOutputRequireds(pMan, Abc_NtkMapCopyCoRequiredCon(pNtk))
	} else {
		mapper.Map_ManSetOutputRequireds(pMan, Abc_NtkMapCopyCoRequired(pNtk, Abc_NtkGetCoRequiredTimes(pNtk)))
	}

	// create PIs and remember them in the old nodes
	baseabc.Abc_NtkCleanCopy(pNtk)
	baseabc.Abc_AigConst1(pNtk).CopyData = mapper.Map_ManReadConst1(pMan)
	// inlining of macro Abc_NtkForEachCi
	for i = int32(0); i < Abc_NtkCiNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCi(pNtk, i)
		if i == Abc_NtkCiNum(pNtk)-baseabc.Abc_NtkBarBufNum(pNtk) {
			break
		}
		pNodeMap = mapper.Map_ManReadInputs(pMan)[i]
		pNode.CopyData = pNodeMap
		if pSwitching != nil {
			mapper.Map_NodeSetSwitching(pNodeMap, pSwitching[pNode.ID])
		}
	}

	// load the AIG into the mapper
	vNodes = baseabc.Abc_AigDfsMap(pNtk)
	// inlining of macro Vec_PtrForEachEntry
	for i := 0; i < len(vNodes); i++ {
		pNode := vNodes[i]
		if baseabc.Abc_ObjIsLatch(pNode) != 0 {
			pFanin = baseabc.Abc_ObjFanin0(pNode)
			// inlining of macro mapper.Map_NotCond
			pNodeMap = mapper.Map_NodeBuf(pMan, mapper.Map_NotCond(mapCopyNode(baseabc.Abc_ObjFanin0(pFanin)), baseabc.Abc_ObjFaninC0(pFanin) != 0))
			baseabc.Abc_ObjFanout0(pNode).CopyData = pNodeMap
			continue
		}
		assert(baseabc.Abc_ObjIsNode(pNode) != 0, "baseabc.Abc_ObjIsNode(pNode)")
		// add the node to the mapper
		pNodeMap = mapper.Map_NodeAnd(pMan,
			// inlining of macro mapper.Map_NotCond
			mapper.Map_NotCond(mapCopyNode(baseabc.Abc_ObjFanin0(pNode)), baseabc.Abc_ObjFaninC0(pNode) != 0),
			// inlining of macro mapper.Map_NotCond
			mapper.Map_NotCond(mapCopyNode(baseabc.Abc_ObjFanin1(pNode)), baseabc.Abc_ObjFaninC1(pNode) != 0))
		assert(pNode.CopyData == nil, "pNode.CopyData == nil")
		// remember the node
		pNode.CopyData = pNodeMap
		if pSwitching != nil {
			mapper.Map_NodeSetSwitching(pNodeMap, pSwitching[pNode.ID])
		}
		// set up the choice node
		if baseabc.Abc_AigNodeIsChoice(pNode) != 0 {
			for pPrev, pFanin = pNode, pNode.Choice; pFanin != nil; pPrev, pFanin = pFanin, pFanin.Choice {
				mapper.Map_NodeSetNextE(mapCopyNode(pPrev), mapCopyNode(pFanin))
				mapper.Map_NodeSetRepr(mapCopyNode(pFanin), mapCopyNode(pNode))
			}
		}
	}
	assert(mapper.Map_ManReadBufNum(pMan) == baseabc.Abc_NtkBarBufNum(pNtk), "mapper.Map_ManReadBufNum(pMan) == baseabc.Abc_NtkBarBufNum(pNtk)")

	// set the primary outputs in the required phase
	// inlining of macro Abc_NtkForEachCo
	for i = int32(0); i < Abc_NtkCoNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCo(pNtk, i)
		if i == Abc_NtkCoNum(pNtk)-baseabc.Abc_NtkBarBufNum(pNtk) {
			break
		}
		// inlining of macro mapper.Map_NotCond
		mapper.Map_ManReadOutputs(pMan)[i] = mapper.Map_NotCond(mapCopyNode(baseabc.Abc_ObjFanin0(pNode)), baseabc.Abc_ObjFaninC0(pNode) != 0)
	}
	return pMan
}

/*
*Function*************************************************************

	Synopsis    [Creates the mapped network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeFromMapSuper_rec(pNtkNew *Abc_Ntk_t, pNodeMap *Map_Node_t, pSuper *Map_Super_t, pNodePis []*Abc_Obj_t, nNodePis int32) *Abc_Obj_t {
	pLib := mapGenlib()
	var pRoot *Mio_Gate_t
	var ppFanins []*Map_Super_t
	var pNodeNew, pNodeFanin *Abc_Obj_t
	var nFanins, Number, i int32

	// get the parameters of the supergate
	pRoot = mapper.Map_SuperReadRoot(pSuper)
	if pRoot == nil {
		Number = mapper.Map_SuperReadNum(pSuper)
		if Number < nNodePis {
			return pNodePis[Number]

		} else {
			//            assert( 0 );
			/* It might happen that a super gate with 5 inputs is constructed that
			 * actually depends only on the first four variables; i.e the fifth is a
			 * don't care -- in that case we connect constant node for the fifth
			 * (since the cut only has 4 variables). An interesting question is what
			 * if the first variable (and not the fifth one is the redundant one;
			 * can that happen?) */
			return baseabc.Abc_NtkCreateNodeConst0(pNtkNew)
		}
	}
	pRoot = mio.Mio_LibraryReadGateByName(pLib, mio.Mio_GateReadName(pRoot), nil)

	// get information about the fanins of the supergate
	nFanins = mapper.Map_SuperReadFaninNum(pSuper)
	ppFanins = mapper.Map_SuperReadFanins(pSuper)
	// create a new node with these fanins
	pNodeNew = baseabc.Abc_NtkCreateNode(pNtkNew)
	for i = 0; i < nFanins; i++ {
		pNodeFanin = Abc_NodeFromMapSuper_rec(pNtkNew, pNodeMap, ppFanins[i], pNodePis, nNodePis)
		baseabc.Abc_ObjAddFanin(pNodeNew, pNodeFanin)
	}
	pNodeNew.Data = pRoot
	return pNodeNew
}
func Abc_NodeFromMapPhase_rec(pNtkNew *Abc_Ntk_t, pNodeMap *Map_Node_t, fPhase int32) *Abc_Obj_t {
	var pNodePIs [10]*Abc_Obj_t
	var pNodeNew *Abc_Obj_t
	var ppLeaves []*Map_Node_t
	var pCutBest *Map_Cut_t
	var pSuperBest *Map_Super_t
	var uPhaseBest uint32
	var i, fInvPin, nLeaves int32

	// make sure the node can be implemented in this phase
	assert(mapper.Map_NodeReadCutBest(pNodeMap, fPhase) != nil || mapper.Map_NodeIsConst(pNodeMap) != 0, "mapper.Map_NodeReadCutBest(pNodeMap, fPhase) != nil || mapper.Map_NodeIsConst(pNodeMap) != 0")
	// check if the phase is already implemented
	pNodeNew, _ = mapper.Map_NodeReadData(pNodeMap, fPhase).(*Abc_Obj_t)
	if pNodeNew != nil {
		return pNodeNew
	}

	// get the information about the best cut
	pCutBest = mapper.Map_NodeReadCutBest(pNodeMap, fPhase)
	pSuperBest = mapper.Map_CutReadSuperBest(pCutBest, fPhase)
	uPhaseBest = mapper.Map_CutReadPhaseBest(pCutBest, fPhase)
	nLeaves = mapper.Map_CutReadLeavesNum(pCutBest)
	ppLeaves = mapper.Map_CutReadLeaves(pCutBest)

	// collect the PI nodes
	for i = 0; i < nLeaves; i++ {
		fInvPin = boolToInt32((uPhaseBest & (1 << i)) > 0)
		pNodePIs[i] = Abc_NodeFromMap_rec(pNtkNew, ppLeaves[i], fInvPin^1)
		assert(pNodePIs[i] != nil, "pNodePIs[i] != nil")
	}

	// implement the supergate
	pNodeNew = Abc_NodeFromMapSuper_rec(pNtkNew, pNodeMap, pSuperBest, pNodePIs[:], nLeaves)
	mapper.Map_NodeSetData(pNodeMap, fPhase, pNodeNew)
	return pNodeNew
}
func Abc_NodeFromMap_rec(pNtkNew *Abc_Ntk_t, pNodeMap *Map_Node_t, fPhase int32) *Abc_Obj_t {
	var pNodeNew, pNodeInv *Abc_Obj_t

	// check the case of constant node
	if mapper.Map_NodeIsConst(pNodeMap) != 0 {
		if fPhase != 0 {
			pNodeNew = baseabc.Abc_NtkCreateNodeConst1(pNtkNew)
		} else {
			pNodeNew = baseabc.Abc_NtkCreateNodeConst0(pNtkNew)
		}
		if pNodeNew.Data == nil {
			fmt.Printf("Error creating mapped network: Library does not have a constant %d gate.\n", fPhase)
		}
		return pNodeNew
	}

	// check if the phase is already implemented
	pNodeNew, _ = mapper.Map_NodeReadData(pNodeMap, fPhase).(*Abc_Obj_t)
	if pNodeNew != nil {
		return pNodeNew
	}

	// implement the node if the best cut is assigned
	if mapper.Map_NodeReadCutBest(pNodeMap, fPhase) != nil {
		return Abc_NodeFromMapPhase_rec(pNtkNew, pNodeMap, fPhase)
	}

	// if the cut is not assigned, implement the node
	assert(mapper.Map_NodeReadCutBest(pNodeMap, fPhase^1) != nil || mapper.Map_NodeIsConst(pNodeMap) != 0, "mapper.Map_NodeReadCutBest(pNodeMap, !fPhase) != nil || mapper.Map_NodeIsConst(pNodeMap) != 0")
	pNodeNew = Abc_NodeFromMapPhase_rec(pNtkNew, pNodeMap, fPhase^1)

	// add the inverter
	pNodeInv = baseabc.Abc_NtkCreateNode(pNtkNew)
	baseabc.Abc_ObjAddFanin(pNodeInv, pNodeNew)
	pNodeInv.Data = mio.Mio_LibraryReadInv(mapGenlib())

	// set the inverter
	mapper.Map_NodeSetData(pNodeMap, fPhase, pNodeInv)
	return pNodeInv
}
func Abc_NtkFromMap(pMan *Map_Man_t, pNtk *Abc_Ntk_t, fUseBuffs int32) *Abc_Ntk_t {
	var pNtkNew *Abc_Ntk_t
	var pNodeMap *Map_Node_t
	var pNode, pNodeNew *Abc_Obj_t
	var i int32
	assert(mapper.Map_ManReadBufNum(pMan) == baseabc.Abc_NtkBarBufNum(pNtk), "mapper.Map_ManReadBufNum(pMan) == baseabc.Abc_NtkBarBufNum(pNtk)")
	// create the new network
	pNtkNew = baseabc.Abc_NtkStartFrom(pNtk, baseabc.ABC_NTK_LOGIC, baseabc.ABC_FUNC_MAP)
	// make the mapper point to the new network
	mapper.Map_ManCleanData(pMan)
	// inlining of macro Abc_NtkForEachCi
	for i = int32(0); i < Abc_NtkCiNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCi(pNtk, i)
		if i >= Abc_NtkCiNum(pNtk)-baseabc.Abc_NtkBarBufNum(pNtk) {
			break
		}
		mapper.Map_NodeSetData(mapper.Map_ManReadInputs(pMan)[i], 1, pNode.Copy)
	}
	// inlining of macro Abc_NtkForEachCi
	for i = int32(0); i < Abc_NtkCiNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCi(pNtk, i)
		if i < Abc_NtkCiNum(pNtk)-baseabc.Abc_NtkBarBufNum(pNtk) {
			continue
		}
		mapper.Map_NodeSetData(mapper.Map_ManReadBufs(pMan)[i-(Abc_NtkCiNum(pNtk)-baseabc.Abc_NtkBarBufNum(pNtk))], 1, pNode.Copy)
	}
	// assign the mapping of the required phase to the POs
	// inlining of macro Abc_NtkForEachCo
	for i = int32(0); i < Abc_NtkCoNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCo(pNtk, i)
		if i < Abc_NtkCoNum(pNtk)-baseabc.Abc_NtkBarBufNum(pNtk) {
			continue
		}
		pNodeMap = mapper.Map_ManReadBufDriver(pMan, i-(Abc_NtkCoNum(pNtk)-baseabc.Abc_NtkBarBufNum(pNtk)))
		// inlining of macro mapper.Map_Regular
		// inlining of macro mapper.Map_IsComplement
		pNodeNew = Abc_NodeFromMap_rec(pNtkNew, mapper.Map_Regular(pNodeMap), boolToInt32(!mapper.Map_IsComplement(pNodeMap)))
		assert(baseabc.Abc_ObjIsComplement(pNodeNew) == 0, "!baseabc.Abc_ObjIsComplement(pNodeNew)")
		baseabc.Abc_ObjAddFanin(pNode.Copy, pNodeNew)
	}
	// inlining of macro Abc_NtkForEachCo
	for i = int32(0); i < Abc_NtkCoNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCo(pNtk, i)
		if i >= Abc_NtkCoNum(pNtk)-baseabc.Abc_NtkBarBufNum(pNtk) {
			break
		}
		pNodeMap = mapper.Map_ManReadOutputs(pMan)[i]
		// inlining of macro mapper.Map_Regular
		// inlining of macro mapper.Map_IsComplement
		pNodeNew = Abc_NodeFromMap_rec(pNtkNew, mapper.Map_Regular(pNodeMap), boolToInt32(!mapper.Map_IsComplement(pNodeMap)))
		assert(baseabc.Abc_ObjIsComplement(pNodeNew) == 0, "!baseabc.Abc_ObjIsComplement(pNodeNew)")
		baseabc.Abc_ObjAddFanin(pNode.Copy, pNodeNew)
	}
	// decouple the PO driver nodes to reduce the number of levels
	baseabc.Abc_NtkLogicMakeSimpleCos(pNtkNew, boolToInt32(fUseBuffs == 0))
	//    if ( nDupGates && mapper.Map_ManReadVerbose(pMan) )
	//        printf( "Duplicated %d gates to decouple the CO drivers.\n", nDupGates );
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Interface with the mapping package.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkSuperChoice(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	var pNtkNew *Abc_Ntk_t

	var pMan *Map_Man_t

	assert(baseabc.Abc_NtkIsStrash(pNtk) != 0, "baseabc.Abc_NtkIsStrash(pNtk)")

	// check that the library is available
	if base_main.Abc_FrameReadLibGen() == nil {
		fmt.Printf("The current library is not available.\n")
		return 0
	}

	// derive the supergate library
	if base_main.Abc_FrameReadLibSuper() == nil && base_main.Abc_FrameReadLibGen() != nil {
		//        printf( "A simple supergate library is derived from gate library \"%s\".\n",
		//            mio.Mio_LibraryReadName((Mio_Library_t *)Abc_FrameReadLibGen()) );
		mapper.Map_SuperLibDeriveFromGenlib(mapGenlib(), 0)
	}

	// print a warning about choice nodes
	if baseabc.Abc_NtkGetChoiceNum(pNtk) != 0 {
		fmt.Printf("Performing mapping with choices.\n")
	}

	// perform the mapping
	pMan = Abc_NtkToMap(pNtk, -1, 1, nil, 0)
	if pMan == nil {
		return nil
	}
	if (mapper.Map_Mapping(pMan) == 0) {
		mapper.Map_ManFree(pMan)
		return nil
	}

	// reconstruct the network after mapping
	pNtkNew = Abc_NtkFromMapSuperChoice(pMan, pNtk)
	if pNtkNew == nil {
		return nil
	}
	mapper.Map_ManFree(pMan)

	// make sure that everything is okay
	if baseabc.Abc_NtkCheck(pNtkNew) == 0 {
		fmt.Printf("Abc_NtkMap: The network check has failed.\n")
		Abc_NtkDelete(pNtkNew)
		return nil
	}
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Creates the mapped network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFromMapSuperChoice(pMan *Map_Man_t, pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	var pProgress *ProgressBar
	var pNtkNew, pNtkNew2 *Abc_Ntk_t
	var pNode *Abc_Obj_t
	var i int32

	// save the pointer to the mapped nodes
	// inlining of macro Abc_NtkForEachCi
	for i = int32(0); i < Abc_NtkCiNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCi(pNtk, i)
		// Mapper node remains in CopyData; Next carries only network object links.
	}
	// inlining of macro Abc_NtkForEachPo
	for i = int32(0); i < Abc_NtkPoNum(pNtk); i++ {
		pNode = Abc_NtkPo(pNtk, i)
		// Mapper node remains in CopyData; Next carries only network object links.
	}
	// inlining of macro Abc_NtkForEachNode
	for i = int32(0); i < baseabc.Abc_NtkObjNumMax(pNtk); i++ {
		pNode = baseabc.Abc_NtkObj(pNtk, i)

	}

	// duplicate the network
	pNtkNew2 = baseabc.Abc_NtkDup(pNtk)
	pNtkNew = Abc_NtkMulti(pNtkNew2, 0, 20, 0, 0, 1, 0)
	if baseabc.Abc_NtkBddToSop(pNtkNew, -1, ABC_INFINITY) == 0 {
		fmt.Printf("Abc_NtkFromMapSuperChoice(): Converting to SOPs has failed.\n")
		return nil
	}

	// set the old network to point to the new network
	// inlining of macro Abc_NtkForEachCi
	for i = int32(0); i < Abc_NtkCiNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCi(pNtk, i)
		pNode.Copy = pNode.Copy.Copy
	}
	// inlining of macro Abc_NtkForEachPo
	for i = int32(0); i < Abc_NtkPoNum(pNtk); i++ {
		pNode = Abc_NtkPo(pNtk, i)
		pNode.Copy = pNode.Copy.Copy
	}
	// inlining of macro Abc_NtkForEachNode
	for i = int32(0); i < baseabc.Abc_NtkObjNumMax(pNtk); i++ {
		pNode = baseabc.Abc_NtkObj(pNtk, i)
		if pNode != nil && baseabc.Abc_ObjIsNode(pNode) != 0 {
			pNode.Copy = pNode.Copy.Copy
		}
	}
	Abc_NtkDelete(pNtkNew2)

	// set the pointers from the mapper to the new nodes
	// inlining of macro Abc_NtkForEachCi
	for i = int32(0); i < Abc_NtkCiNum(pNtk); i++ {
		pNode := baseabc.Abc_NtkCi(pNtk, i)
		mapper.Map_NodeSetData(mapper.Map_ManReadInputs(pMan)[i], 0, baseabc.Abc_NtkCreateNodeInv(pNtkNew, pNode.Copy))
		mapper.Map_NodeSetData(mapper.Map_ManReadInputs(pMan)[i], 1, pNode.Copy)
	}
	// inlining of macro Abc_NtkForEachNode
	for i = int32(0); i < baseabc.Abc_NtkObjNumMax(pNtk); i++ {
		pNode = baseabc.Abc_NtkObj(pNtk, i)
		if pNode == nil || baseabc.Abc_ObjIsNode(pNode) == 0 {
			continue
		}
		//        if ( baseabc.Abc_NodeIsConst(pNode) )
		//            continue;
		pMapNode := mapCopyNode(pNode)
		mapper.Map_NodeSetData(pMapNode, 0, baseabc.Abc_NtkCreateNodeInv(pNtkNew, pNode.Copy))
		mapper.Map_NodeSetData(pMapNode, 1, pNode.Copy)
	}

	// assign the mapping of the required phase to the POs
	pProgress = extra.Extra_ProgressBarStart(os.Stdout, baseabc.Abc_NtkObjNumMax(pNtk))
	// inlining of macro Abc_NtkForEachNode
	for i = int32(0); i < baseabc.Abc_NtkObjNumMax(pNtk); i++ {
		pNode = baseabc.Abc_NtkObj(pNtk, i)
		if pNode == nil || baseabc.Abc_ObjIsNode(pNode) == 0 {
			continue
		}
		extra.Extra_ProgressBarUpdate(pProgress, i, nil)
		//        if ( baseabc.Abc_NodeIsConst(pNode) )
		//            continue;
		Abc_NodeSuperChoice(pNtkNew, pNode)
	}
	extra.Extra_ProgressBarStop(pProgress)
	return pNtkNew
}

/*
*Function*************************************************************

	Synopsis    [Creates the mapped network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeSuperChoice(pNtkNew *Abc_Ntk_t, pNode *Abc_Obj_t) {
	pMapNode := mapCopyNode(pNode)
	var pCuts, pTemp *Map_Cut_t

	pCuts = mapper.Map_NodeReadCuts(pMapNode)
	for pTemp = mapper.Map_CutReadNext(pCuts); pTemp != nil; pTemp = mapper.Map_CutReadNext(pTemp) {
		Abc_NodeFromMapCutPhase(pNtkNew, pTemp, 0)
		Abc_NodeFromMapCutPhase(pNtkNew, pTemp, 1)
	}
}

/*
*Function*************************************************************

	Synopsis    [Constructs the nodes corrresponding to one node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeFromMapCutPhase(pNtkNew *Abc_Ntk_t, pCut *Map_Cut_t, fPhase int32) {
	var pNodePIs [10]*Abc_Obj_t
	var ppLeaves []*Map_Node_t
	var pSuperBest *Map_Super_t
	var uPhaseBest uint32
	var i, fInvPin, nLeaves int32

	pSuperBest = mapper.Map_CutReadSuperBest(pCut, fPhase)
	if pSuperBest == nil {
		return
	}

	// get the information about the best cut
	uPhaseBest = mapper.Map_CutReadPhaseBest(pCut, fPhase)
	nLeaves = mapper.Map_CutReadLeavesNum(pCut)
	ppLeaves = mapper.Map_CutReadLeaves(pCut)

	// collect the PI nodes
	for i = 0; i < nLeaves; i++ {
		fInvPin = boolToInt32((uPhaseBest & (1 << i)) > 0)
		pNodePIs[i], _ = mapper.Map_NodeReadData(ppLeaves[i], fInvPin^1).(*Abc_Obj_t)
		assert(pNodePIs[i] != nil, "pNodePIs[i] != nil")
	}

	// implement the supergate
	Abc_NodeFromMapSuperChoice_rec(pNtkNew, pSuperBest, pNodePIs[:], nLeaves)
}

/*
*Function*************************************************************

	Synopsis    [Constructs the nodes corrresponding to one supergate.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NodeFromMapSuperChoice_rec(pNtkNew *Abc_Ntk_t, pSuper *Map_Super_t, pNodePis []*Abc_Obj_t, nNodePis int32) *Abc_Obj_t {
	pLib := mapGenlib()
	var pRoot *Mio_Gate_t
	var ppFanins []*Map_Super_t
	var pNodeNew, pNodeFanin *Abc_Obj_t
	var nFanins, Number, i int32

	// get the parameters of the supergate
	pRoot = mapper.Map_SuperReadRoot(pSuper)
	if pRoot == nil {
		Number = mapper.Map_SuperReadNum(pSuper)
		if Number < nNodePis {
			return pNodePis[Number]

		} else {
			//            assert( 0 );
			/* It might happen that a super gate with 5 inputs is constructed that
			 * actually depends only on the first four variables; i.e the fifth is a
			 * don't care -- in that case we connect constant node for the fifth
			 * (since the cut only has 4 variables). An interesting question is what
			 * if the first variable (and not the fifth one is the redundant one;
			 * can that happen?) */
			return baseabc.Abc_NtkCreateNodeConst0(pNtkNew)
		}
	}
	pRoot = mio.Mio_LibraryReadGateByName(pLib, mio.Mio_GateReadName(pRoot), nil)

	// get information about the fanins of the supergate
	nFanins = mapper.Map_SuperReadFaninNum(pSuper)
	ppFanins = mapper.Map_SuperReadFanins(pSuper)
	// create a new node with these fanins
	pNodeNew = baseabc.Abc_NtkCreateNode(pNtkNew)
	for i = 0; i < nFanins; i++ {
		pNodeFanin = Abc_NodeFromMapSuperChoice_rec(pNtkNew, ppFanins[i], pNodePis, nNodePis)
		baseabc.Abc_ObjAddFanin(pNodeNew, pNodeFanin)
	}
	pNodeNew.Data = baseabc.Abc_SopRegister(nil, mio.Mio_GateReadSop(pRoot))
	return pNodeNew
}

/*
*Function*************************************************************

	Synopsis    [Returns the twin node if it exists.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkFetchTwinNode(pNode *Abc_Obj_t) *Abc_Obj_t {
	var pNode2 *Abc_Obj_t
	pGate, _ := pNode.Data.(*Mio_Gate_t)
	assert(baseabc.Abc_NtkHasMapping(pNode.Network()) != 0, "baseabc.Abc_NtkHasMapping(pNode.Network())")
	if pGate == nil || mio.Mio_GateReadTwin(pGate) == nil {
		return nil
	}
	// assuming the twin node is following next
	if int32(baseabc.Abc_ObjId(pNode)) == baseabc.Abc_NtkObjNumMax(pNode.Network())-1 {
		return nil
	}
	pNode2 = baseabc.Abc_NtkObj(pNode.Network(), baseabc.Abc_ObjId(pNode)+1)
	if pNode2 == nil || baseabc.Abc_ObjIsNode(pNode2) == 0 || baseabc.Abc_ObjFaninNum(pNode) != baseabc.Abc_ObjFaninNum(pNode2) {
		return nil
	}
	pGate2, _ := pNode2.Data.(*Mio_Gate_t)
	if mio.Mio_GateReadTwin(pGate) != pGate2 {
		return nil
	}
	return pNode2
}

/*
*Function*************************************************************

	Synopsis    [Dumps mapped network in the mini-mapped format.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkWriteMiniMapping(pNtk *Abc_Ntk_t) *Vec_Int_t {
	var vNodes []*Abc_Obj_t
	var vMapping *Vec_Int_t
	var vGates *Vec_Str_t
	var pObj, pFanin *Abc_Obj_t
	var i int32
	var k int32
	var nNodes int32
	var nFanins int32
	var nExtra int32
	var pArray []int32
	assert(baseabc.Abc_NtkHasMapping(pNtk) != 0, "baseabc.Abc_NtkHasMapping(pNtk)")
	// collect nodes in the DFS order
	vNodes = baseabc.Abc_NtkDfs(pNtk, 0)
	// assign unique numbers
	nNodes, nFanins = 0, 0
	// inlining of macro Abc_NtkForEachCi
	for i = int32(0); i < Abc_NtkCiNum(pNtk); i++ {
		pObj = baseabc.Abc_NtkCi(pNtk, i)
		pObj.Temp = nNodes
		nNodes++
	}
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < len(vNodes); i++ {
		pObj = vNodes[i]
		pObj.Temp = nNodes
		nNodes++
		nFanins += baseabc.Abc_ObjFaninNum(pObj)
	}
	// allocate attay to store mapping (4 counters + fanins for each node + PO drivers + gate names)
	vMapping = Vec_IntAlloc(4 + int(baseabc.Abc_NtkNodeNum(pNtk)+nFanins+Abc_NtkCoNum(pNtk)) + 10000)
	// write the numbers of CI/CO/Node/FF
	Vec_IntPush(vMapping, Abc_NtkCiNum(pNtk))
	Vec_IntPush(vMapping, Abc_NtkCoNum(pNtk))
	Vec_IntPush(vMapping, baseabc.Abc_NtkNodeNum(pNtk))
	Vec_IntPush(vMapping, Abc_NtkLatchNum(pNtk))
	// write the nodes
	vGates = Vec_StrAlloc(10000)
	// inlining of macro Vec_PtrForEachEntry
	for i = 0; i < len(vNodes); i++ {
		pObj = vNodes[i]
		Vec_IntPush(vMapping, baseabc.Abc_ObjFaninNum(pObj))
		// inlining of macro Abc_ObjForEachFanin
		for k = int32(0); k < baseabc.Abc_ObjFaninNum(pObj); k++ {
			pFanin = baseabc.Abc_ObjFanin(pObj, k)
			Vec_IntPush(vMapping, pFanin.Temp)
		}
		// remember this gate (to be added to the mapping later)
		gate, _ := pObj.Data.(*Mio_Gate_t)
		aig_gia.Vec_StrPrintStr(vGates, mio.Mio_GateReadName(gate))
		Vec_StrPush(vGates, 0)
	}
	// write the COs literals
	// inlining of macro Abc_NtkForEachCo
	for i = int32(0); i < Abc_NtkCoNum(pNtk); i++ {
		pObj = baseabc.Abc_NtkCo(pNtk, i)
		Vec_IntPush(vMapping, baseabc.Abc_ObjFanin0(pObj).Temp)
	}
	// finish off the array
	nExtra = int32(4 - Vec_StrSize(vGates)%4)
	for i = 0; i < nExtra; i++ {
		Vec_StrPush(vGates, 0)
	}
	// add gates to the array
	assert(Vec_StrSize(vGates)%4 == 0, "Vec_StrSize(vGates) % 4 == 0")
	nExtra = int32(Vec_StrSize(vGates) / 4)
	gateBytes := aig_gia.Vec_StrArray(vGates)
	for i = int32(0); i < nExtra; i++ {
		j := int(i * 4)
		packed := int32(binary.NativeEndian.Uint32(gateBytes[j : j+4]))
		Vec_IntPush(vMapping, packed)
	}
	// cleanup and return
	Vec_StrFree(vGates)
	return vMapping
}

/*
*Function*************************************************************

	Synopsis    [Prints mapped network represented in mini-mapped format.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkPrintMiniMapping(pArray []int32) {
	var nCis, nCos, nNodes, nFlops int32
	var i int32
	var k int32
	var nLeaves int32
	var Pos int32 = 4
	var pBuffer, pName string
	nCis = pArray[0]
	nCos = pArray[1]
	nNodes = pArray[2]
	nFlops = pArray[3]
	fmt.Printf("Mapped network has %d CIs, %d COs, %d gates, and %d flops.\n", nCis, nCos, nNodes, nFlops)
	fmt.Printf("The first %d object IDs (from 0 to %d) are reserved for the CIs.\n", nCis, nCis-1)
	for i = int32(0); i < nNodes; i++ {
		fmt.Printf("Node %d has fanins {", nCis+i)
		nLeaves = pArray[Pos]
		Pos++
		for k = int32(0); k < nLeaves; k++ {
			fmt.Printf(" %d", pArray[Pos])
			Pos++
		}
		fmt.Printf(" }\n")
	}
	for i = int32(0); i < nCos; i++ {
		fmt.Printf("CO %d is driven by node %d\n", i, pArray[Pos])
		Pos++
	}
	raw := make([]byte, 0, (len(pArray)-int(Pos))*4)
	for _, word := range pArray[Pos:] {
		var bytes [4]byte
		binary.NativeEndian.PutUint32(bytes[:], uint32(word))
		raw = append(raw, bytes[:]...)
	}
	pBuffer = string(raw)
	for i = int32(0); i < nNodes; i++ {
		nul := strings.IndexByte(pBuffer, 0)
		if nul < 0 {
			nul = len(pBuffer)
		}
		pName = pBuffer[:nul]
		if nul < len(pBuffer) {
			pBuffer = pBuffer[nul+1:]
		} else {
			pBuffer = ""
		}
		fmt.Printf("Node %d has gate \"%s\"\n", nCis+i, pName)
	}
}

/*
*Function*************************************************************

	Synopsis    [This procedure outputs an array representing mini-mapped network.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkOutputMiniMapping(pAbc *Abc_Frame_t) []int32 {
	//Abc_Frame_t * pAbc = (Abc_Frame_t *)pAbc0;
	var pNtk *Abc_Ntk_t
	var vMapping *Vec_Int_t
	var pArray []int32
	if pAbc == nil {
		fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n")
	}
	pNtk = Abc_FrameReadNtk(pAbc)
	if pNtk == nil {
		fmt.Printf("Current network in ABC framework is not defined.\n")
	}
	if baseabc.Abc_NtkHasMapping(pNtk) == 0 {
		fmt.Printf("Current network in ABC framework is not mapped.\n")
	}
	// derive mini-mapping
	vMapping = Abc_NtkWriteMiniMapping(pNtk)
	pArray = Vec_IntArray(vMapping)
	vMapping = nil
	// print mini-mapping (optional)
	//    Abc_NtkPrintMiniMapping( pArray );
	// return the array representation of mini-mapping
	return pArray
}

/*
*Function*************************************************************

	Synopsis    [Test for mini-mapped format.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkTestMiniMapping(p *Abc_Ntk_t) {
	var vMapping *Vec_Int_t
	vMapping = Abc_NtkWriteMiniMapping(p)
	Abc_NtkPrintMiniMapping(Vec_IntArray(vMapping))
	fmt.Printf("Array has size %d ints.\n", Vec_IntSize(vMapping))
	Vec_IntFree(vMapping)
}

/*
*Function*************************************************************

	Synopsis    [These APIs set arrival/required times of CIs/COs.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkSetCiArrivalTime(pAbc *Abc_Frame_t, iCi int32, Rise float32, Fall float32) {
	//Abc_Frame_t * pAbc = (Abc_Frame_t *)pAbc0;
	var pNtk *Abc_Ntk_t
	var pNode *Abc_Obj_t
	if pAbc == nil {
		fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n")
		return
	}
	pNtk = Abc_FrameReadNtk(pAbc)
	if pNtk == nil {
		fmt.Printf("Current network in ABC framework is not defined.\n")
		return
	}
	if iCi < 0 || iCi >= Abc_NtkCiNum(pNtk) {
		fmt.Printf("CI index is not valid.\n")
		return
	}
	pNode = baseabc.Abc_NtkCi(pNtk, iCi)
	Abc_NtkTimeSetArrival(pNtk, baseabc.Abc_ObjId(pNode), Rise, Fall)
}
func Abc_NtkSetCoRequiredTime(pAbc *Abc_Frame_t, iCo int32, Rise float32, Fall float32) {
	//Abc_Frame_t * pAbc = (Abc_Frame_t *)pAbc0;
	var pNtk *Abc_Ntk_t
	var pNode *Abc_Obj_t
	if pAbc == nil {
		fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n")
		return
	}
	pNtk = Abc_FrameReadNtk(pAbc)
	if pNtk == nil {
		fmt.Printf("Current network in ABC framework is not defined.\n")
		return
	}
	if iCo < 0 || iCo >= Abc_NtkCoNum(pNtk) {
		fmt.Printf("CO index is not valid.\n")
		return
	}
	pNode = baseabc.Abc_NtkCo(pNtk, iCo)
	Abc_NtkTimeSetRequired(pNtk, baseabc.Abc_ObjId(pNode), Rise, Fall)
}

/*
*Function*************************************************************

	Synopsis    [This APIs set AND gate delay.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Abc_NtkSetAndGateDelay(pAbc *Abc_Frame_t, Delay float32) {
	//Abc_Frame_t * pAbc = (Abc_Frame_t *)pAbc0;
	var pNtk *Abc_Ntk_t
	if pAbc == nil {
		fmt.Printf("ABC framework is not initialized by calling Abc_Start()\n")
		return
	}
	pNtk = Abc_FrameReadNtk(pAbc)
	if pNtk == nil {
		fmt.Printf("Current network in ABC framework is not defined.\n")
		return
	}
	pNtk.AndGateDelay = float64(Delay)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

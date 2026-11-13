package mapper

import (
	"fmt"
	"math"
	"time"
)

import "os"

import "github.com/glycerine/abcgo/map/mio"
import abcmain "github.com/glycerine/abcgo/base/main"
import "unsafe"
import "github.com/glycerine/abcgo/misc/extra"
import abcutil "github.com/glycerine/abcgo/misc/util"

/**CFile****************************************************************

  FileName    [mapperCreate.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperCreate.c,v 1.15 2005/02/28 05:34:26 alanmi Exp $]

***********************************************************************/

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void            Map_TableCreate( Map_Man_t * p );

// static void            Map_TableResize( Map_Man_t * p );

// hash key for the structural hash table
func Map_HashKey2(p0, p1 *Map_Node_t, tableSize int32) uint32 {
	if tableSize <= 0 {
		return 0
	}
	return uint32((uint32(p0.Num)*12582917 + uint32(p1.Num)) % uint32(tableSize))
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Reads parameters from the mapping manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_ManReadInputNum(p *Map_Man_t) int32                       { return p.nInputs }
func Map_ManReadOutputNum(p *Map_Man_t) int32                      { return p.nOutputs }
func Map_ManReadBufNum(p *Map_Man_t) int32                         { return Map_NodeVecReadSize(p.vMapBufs) }
func Map_ManReadInputs(p *Map_Man_t) []*Map_Node_t                 { return p.pInputs }
func Map_ManReadOutputs(p *Map_Man_t) []*Map_Node_t                { return p.pOutputs }
func Map_ManReadBufs(p *Map_Man_t) []*Map_Node_t                   { return Map_NodeVecReadArray(p.vMapBufs) }
func Map_ManReadBufDriver(p *Map_Man_t, i int32) *Map_Node_t       { return Map_ManReadBufs(p)[i].p1 }
func Map_ManReadConst1(p *Map_Man_t) *Map_Node_t                   { return p.pConst1 }
func Map_ManReadInputArrivals(p *Map_Man_t) []Map_Time_t           { return p.pInputArrivals }
func Map_ManReadOutputRequireds(p *Map_Man_t) []Map_Time_t         { return p.pOutputRequireds }
func Map_ManReadGenLib(p *Map_Man_t) *Mio_Library_t                { return p.pSuperLib.pGenlib }
func Map_ManReadVerbose(p *Map_Man_t) int32                        { return p.fVerbose }
func Map_ManReadAreaFinal(p *Map_Man_t) float32                    { return p.AreaFinal }
func Map_ManReadRequiredGlo(p *Map_Man_t) float32                  { return p.fRequiredGlo }
func Map_ManSetOutputNames(p *Map_Man_t, ppNames []string)         { p.ppOutputNames = ppNames }
func Map_ManSetAreaRecovery(p *Map_Man_t, fAreaRecovery int32)     { p.fAreaRecovery = fAreaRecovery }
func Map_ManSetDelayTarget(p *Map_Man_t, DelayTarget float32)      { p.DelayTarget = DelayTarget }
func Map_ManSetInputArrivals(p *Map_Man_t, pArrivals []Map_Time_t) { p.pInputArrivals = pArrivals }
func Map_ManSetOutputRequireds(p *Map_Man_t, pRequireds []Map_Time_t) {
	p.pOutputRequireds = pRequireds
}
func Map_ManSetObeyFanoutLimits(p *Map_Man_t, fObeyFanoutLimits int32) {
	p.fObeyFanoutLimits = fObeyFanoutLimits
}
func Map_ManSetNumIterations(p *Map_Man_t, nIterations int32)  { p.nIterations = nIterations }
func Map_ManReadFanoutViolations(p *Map_Man_t) int32           { return p.nFanoutViolations }
func Map_ManSetFanoutViolations(p *Map_Man_t, nVio int32)      { p.nFanoutViolations = nVio }
func Map_ManSetChoiceNodeNum(p *Map_Man_t, nChoiceNodes int32) { p.nChoiceNodes = nChoiceNodes }
func Map_ManSetChoiceNum(p *Map_Man_t, nChoices int32)         { p.nChoices = nChoices }
func Map_ManSetVerbose(p *Map_Man_t, fVerbose int32)           { p.fVerbose = fVerbose }
func Map_ManSetSwitching(p *Map_Man_t, fSwitching int32)       { p.fSwitching = fSwitching }
func Map_ManSetSkipFanout(p *Map_Man_t, fSkipFanout int32)     { p.fSkipFanout = fSkipFanout }
func Map_ManSetUseProfile(p *Map_Man_t)                        { p.fUseProfile = 1 }

/*
*Function*************************************************************

	Synopsis    [Reads parameters from the mapping node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_NodeReadMan(p *Map_Node_t) *Map_Man_t { return p.p }
func Map_NodeReadData(p *Map_Node_t, fPhase int32) any {
	if fPhase != 0 {
		return p.pData1
	} else {
		return p.pData0
	}
}
func Map_NodeReadNum(p *Map_Node_t) int32 { return p.Num }

// inlining of macro Map_Regular
func Map_NodeReadLevel(p *Map_Node_t) int32                      { return int32(Map_Regular(p).Level) }
func Map_NodeReadCuts(p *Map_Node_t) *Map_Cut_t                  { return p.pCuts }
func Map_NodeReadCutBest(p *Map_Node_t, fPhase int32) *Map_Cut_t { return p.pCutBest[fPhase] }
func Map_NodeReadOne(p *Map_Node_t) *Map_Node_t                  { return p.p1 }
func Map_NodeReadTwo(p *Map_Node_t) *Map_Node_t                  { return p.p2 }
func Map_NodeSetData(p *Map_Node_t, fPhase int32, pData any) {
	if fPhase != 0 {
		p.pData1 = pData
	} else {
		p.pData0 = pData
	}
}
func Map_NodeSetNextE(p *Map_Node_t, pNextE *Map_Node_t)    { p.pNextE = pNextE }
func Map_NodeSetRepr(p *Map_Node_t, pRepr *Map_Node_t)      { p.pRepr = pRepr }
func Map_NodeSetSwitching(p *Map_Node_t, Switching float32) { p.Switching = Switching }

/**Function*************************************************************

  Synopsis    [Checks the type of the node.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// inlining of macro Map_Regular
func Map_NodeIsConst(p *Map_Node_t) int32 {
	if Map_Regular(p).Num == -1 {
		return 1
	}
	return 0
}

// inlining of macro Map_Regular
func Map_NodeIsVar(p *Map_Node_t) int32 {
	n := Map_Regular(p)
	if n.p1 == nil && n.Num >= 0 {
		return 1
	}
	return 0
}

// inlining of macro Map_Regular
func Map_NodeIsBuf(p *Map_Node_t) int32 {
	n := Map_Regular(p)
	if n.p1 != nil && n.p2 == nil {
		return 1
	}
	return 0
}

// inlining of macro Map_Regular
func Map_NodeIsAnd(p *Map_Node_t) int32 {
	n := Map_Regular(p)
	if n.p1 != nil && n.p2 != nil {
		return 1
	}
	return 0
}

// inlining of macro Map_IsComplement
func Map_NodeComparePhase(p1, p2 *Map_Node_t) int32 { return int32(p1.fInv ^ p2.fInv) }

/*
*Function*************************************************************

	Synopsis    [Reads parameters from the cut.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_CutReadSuperBest(p *Map_Cut_t, fPhase int32) *Map_Super_t { return p.M[fPhase].pSuperBest }
func Map_CutReadSuper0(p *Map_Cut_t) *Map_Super_t                  { return p.M[0].pSuperBest }
func Map_CutReadSuper1(p *Map_Cut_t) *Map_Super_t                  { return p.M[1].pSuperBest }
func Map_CutReadLeavesNum(p *Map_Cut_t) int32                      { return p.nLeaves }
func Map_CutReadLeaves(p *Map_Cut_t) []*Map_Node_t                 { return p.ppLeaves[:] }
func Map_CutReadPhaseBest(p *Map_Cut_t, fPhase int32) uint32       { return p.M[fPhase].uPhaseBest }
func Map_CutReadPhase0(p *Map_Cut_t) uint32                        { return p.M[0].uPhaseBest }
func Map_CutReadPhase1(p *Map_Cut_t) uint32                        { return p.M[1].uPhaseBest }
func Map_CutReadNext(p *Map_Cut_t) *Map_Cut_t                      { return p.pNext }

/*
*Function*************************************************************

	Synopsis    [Reads parameters from the supergate.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_SuperReadFormula(p *Map_Super_t) string        { return p.pFormula }
func Map_SuperReadRoot(p *Map_Super_t) *Mio_Gate_t      { return p.pRoot }
func Map_SuperReadNum(p *Map_Super_t) int32             { return p.Num }
func Map_SuperReadFanins(p *Map_Super_t) []*Map_Super_t { return p.pFanins[:] }
func Map_SuperReadFaninNum(p *Map_Super_t) int32        { return int32(p.nFanins) }
func Map_SuperReadNext(p *Map_Super_t) *Map_Super_t     { return p.pNext }
func Map_SuperReadNumPhases(p *Map_Super_t) int32       { return int32(p.nPhases) }
func Map_SuperReadPhases(p *Map_Super_t) []uint8        { return p.uPhases[:] }
func Map_SuperReadFanoutLimit(p *Map_Super_t) int32     { return int32(p.nFanLimit) }

func Map_SuperLibReadGenLib(p *Map_SuperLib_t) *mio.Mio_Library_t { return p.pGenlib }
func Map_SuperLibReadAreaInv(p *Map_SuperLib_t) float32           { return p.AreaInv }
func Map_SuperLibReadDelayInv(p *Map_SuperLib_t) Map_Time_t       { return p.tDelayInv }
func Map_SuperLibReadVarsMax(p *Map_SuperLib_t) int32             { return p.nVarsMax }

/*
*Function*************************************************************

	Synopsis    [Create the mapping manager.]

	Description [The number of inputs and outputs is assumed to be
	known is advance. It is much simpler to have them fixed upfront.
	When it comes to representing the object graph in the form of
	AIG, the resulting manager is similar to the regular AIG manager,
	except that it does not use reference counting (and therefore
	does not have garbage collections). It does have table resizing.
	The data structure is more flexible to represent additional
	information needed for mapping.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_ManCreate(nInputs, nOutputs, fVerbose int32) *Map_Man_t {
	if nInputs < 0 || nOutputs < 0 {
		return nil
	}
	lib, ok := abcmain.Abc_FrameReadLibSuper().(*Map_SuperLib_t)
	if !ok || lib == nil {
		fmt.Fprintln(os.Stderr, `The supergate library is not specified. Use "read_super".`)
		return nil
	}
	p := &Map_ManStruct_t_{pSuperLib: lib, nVarsMax: lib.nVarsMax, fVerbose: fVerbose, fEpsilon: 0.001}
	if p.nVarsMax <= 0 {
		return nil
	}
	if p.nVarsMax == 5 {
		p.uCanons, p.uPhases, p.pCounters = extra.Extra_Truth4VarN(8)
	}
	Map_TableCreate(p)
	Map_MappingSetupTruthTables(p.uTruths[:])
	Map_MappingSetupTruthTablesLarge(p.uTruthsLarge[:])
	p.mmNodes = extra.Extra_MmFixedStart(int32(unsafe.Sizeof(Map_NodeStruct_t_{})))
	p.mmCuts = extra.Extra_MmFixedStart(int32(unsafe.Sizeof(Map_CutStruct_t_{})))
	p.vMapObjs, p.vMapBufs, p.vVisited = Map_NodeVecAlloc(100), Map_NodeVecAlloc(100), Map_NodeVecAlloc(100)
	p.nNodes = -1
	p.pConst1 = Map_NodeCreate(p, nil, nil)
	p.pInputs = make([]*Map_Node_t, nInputs)
	p.nInputs = nInputs
	for i := range p.pInputs {
		p.pInputs[i] = Map_NodeCreate(p, nil, nil)
	}
	p.pOutputs = make([]*Map_Node_t, nOutputs)
	p.nOutputs = nOutputs
	return p
}

func Map_ManFree(p *Map_Man_t) {
	if p == nil {
		return
	}
	Map_NodeVecFree(p.vMapObjs)
	Map_NodeVecFree(p.vMapBufs)
	Map_NodeVecFree(p.vVisited)
	extra.Extra_MmFixedStop(p.mmNodes)
	extra.Extra_MmFixedStop(p.mmCuts)
	p.pNodeDelays, p.pInputArrivals, p.pOutputRequireds = nil, nil, nil
	p.pInputs, p.pOutputs, p.pBins = nil, nil, nil
	p.ppOutputNames, p.uCanons, p.uPhases, p.pCounters = nil, nil, nil, nil
}

func Map_ManCreateNodeDelays(p *Map_Man_t, logFan int32) {
	if p.pNodeDelays != nil {
		panic("mapper: node delays already initialized")
	}
	p.pNodeDelays = make([]float32, p.vMapObjs.nSize)
	for i, node := range Map_NodeVecReadArray(p.vMapObjs) {
		if node.nRefs != 0 {
			p.pNodeDelays[i] = 0.014426 * float32(logFan) * p.pSuperLib.tDelayInv.Worst * float32(math.Log(float64(node.nRefs)))
		}
	}
}

func Map_ManPrintTimeStats(p *Map_Man_t) {
	for _, entry := range []struct {
		name  string
		value int64
	}{
		{"ToMap", p.timeToMap}, {"Cuts", p.timeCuts}, {"Truth", p.timeTruth}, {"Match", p.timeMatch},
		{"Area", p.timeArea}, {"Sweep", p.timeSweep}, {"ToNet", p.timeToNet}, {"TOTAL", p.timeTotal},
	} {
		fmt.Printf("%-8s %9.2f sec\n", entry.name, float64(entry.value)/float64(time.Second))
	}
}

func Map_ManPrintStatsToFile(name string, area, delay float32, elapsed abctime) {
	f, err := os.OpenFile("map_stats.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(f, "%s %4.2f %4.2f %4.2f\n", name, area, delay, float64(elapsed)/float64(time.Second))
	_ = f.Close()
}

func Map_NodeCreate(p *Map_Man_t, p1, p2 *Map_Node_t) *Map_Node_t {
	node := &Map_NodeStruct_t_{p: p, p1: p1, p2: p2}
	for phase := range node.tRequired {
		node.tRequired[phase] = Map_Time_t{Rise: MAP_FLOAT_LARGE, Fall: MAP_FLOAT_LARGE, Worst: MAP_FLOAT_LARGE}
	}
	node.Num = p.nNodes
	p.nNodes++
	if node.Num >= 0 {
		Map_NodeVecPush(p.vMapObjs, node)
	} else {
		node.fInv = 1
	}
	if p1 != nil {
		p1 = Map_Regular(p1)
		node.Level = p1.Level
		if p2 != nil && Map_Regular(p2).Level > node.Level {
			node.Level = Map_Regular(p2).Level
		}
		if p2 != nil {
			node.Level++
		}
		phase1, phase2 := mapNodeSimComplement(p1), false
		if p2 != nil {
			phase2 = mapNodeSimComplement(p2)
		}
		if phase1 && phase2 {
			node.fInv = 1
		}
	}
	if p1 != nil {
		Map_NodeRef(p1)
	}
	if p2 != nil {
		Map_NodeRef(p2)
	}
	node.nRefEst[0], node.nRefEst[1] = -1, -1
	return node
}

func mapNodeSimComplement(node *Map_Node_t) bool {
	if Map_IsComplement(node) {
		return Map_Regular(node).fInv == 0
	}
	return node.fInv != 0
}

func Map_TableCreate(p *Map_Man_t) {
	if p.pBins != nil {
		panic("mapper: unique table already initialized")
	}
	p.nBins = abcutil.Abc_PrimeCudd(5000)
	p.pBins = make([]*Map_Node_t, p.nBins)
	p.nNodes = 0
}

func Map_NodeAnd(p *Map_Man_t, p1, p2 *Map_Node_t) *Map_Node_t {
	if p1 == p2 {
		return p1
	}
	if Map_Not(p1) == p2 {
		return Map_Not(p.pConst1)
	}
	if Map_NodeIsConst(p1) != 0 {
		if p1 == p.pConst1 {
			return p2
		}
		return Map_Not(p.pConst1)
	}
	if Map_NodeIsConst(p2) != 0 {
		if p2 == p.pConst1 {
			return p1
		}
		return Map_Not(p.pConst1)
	}
	if Map_Regular(p1).Num > Map_Regular(p2).Num {
		p1, p2 = p2, p1
	}
	key := Map_HashKey2(p1, p2, p.nBins)
	for node := p.pBins[key]; node != nil; node = node.pNext {
		if node.p1 == p1 && node.p2 == p2 {
			return node
		}
	}
	if p.nNodes >= 2*p.nBins {
		Map_TableResize(p)
		key = Map_HashKey2(p1, p2, p.nBins)
	}
	node := Map_NodeCreate(p, p1, p2)
	node.pNext, p.pBins[key] = p.pBins[key], node
	return node
}

func Map_TableResize(p *Map_Man_t) {
	old := p.pBins
	newSize := abcutil.Abc_PrimeCudd(uint32(2 * p.nBins))
	bins := make([]*Map_Node_t, newSize)
	count := int32(0)
	for _, head := range old {
		for node := head; node != nil; {
			next := node.pNext
			key := Map_HashKey2(node.p1, node.p2, newSize)
			node.pNext, bins[key] = bins[key], node
			count++
			node = next
		}
	}
	assert(count == p.nNodes-p.nInputs, "count == p.nNodes - p.nInputs")
	p.pBins, p.nBins = bins, newSize
}

func Map_NodeBuf(p *Map_Man_t, p1 *Map_Node_t) *Map_Node_t {
	node := Map_NodeCreate(p, p1, nil)
	Map_NodeVecPush(p.vMapBufs, node)
	return node
}

func Map_NodeSetChoice(_ *Map_Man_t, oldNode, newNode *Map_Node_t) {
	newNode.pNextE, oldNode.pNextE = oldNode.pNextE, newNode
	newNode.pRepr = oldNode
}

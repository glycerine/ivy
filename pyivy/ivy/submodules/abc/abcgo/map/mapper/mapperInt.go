package mapper

import (
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/st"
	"github.com/glycerine/abcgo/misc/vec"
)

// Map_Time_t is the rise/fall arrival or required time of a signal.

type abctime = int64
type Vec_Str_t = vec.Vec_Str_t
type st__table = st.Table

type Mio_Library_t = mio.Mio_Library_t
type Mio_Gate_t = mio.Mio_Gate_t
type Mio_Pin_t = mio.Mio_Pin_t
type Extra_MmFixed_t = extra.Extra_MmFixed_t
type Extra_MmFlex_t = extra.Extra_MmFlex_t

type Map_ManStruct_t_ struct {
	pBins                                                                                                     []*Map_Node_t
	nBins                                                                                                     int32
	pInputs                                                                                                   []*Map_Node_t
	nInputs                                                                                                   int32
	pOutputs                                                                                                  []*Map_Node_t
	nOutputs, nNodes                                                                                          int32
	pConst1                                                                                                   *Map_Node_t
	vMapObjs, vMapBufs                                                                                        *Map_NodeVec_t
	pNodeDelays                                                                                               []float32
	ppOutputNames                                                                                             []string
	pInputArrivals, pOutputRequireds                                                                          []Map_Time_t
	nVarsMax, fAreaRecovery, fVerbose, fMappingMode                                                           int32
	fRequiredGlo, fEpsilon, AreaBase, AreaFinal                                                               float32
	nIterations, fObeyFanoutLimits                                                                            int32
	DelayTarget                                                                                               float32
	nTravIds, fSwitching, fSkipFanout, fUseProfile                                                            int32
	pSuperLib                                                                                                 *Map_SuperLib_t
	uTruths                                                                                                   [6][2]uint32
	uTruthsLarge                                                                                              [10][32]uint32
	nCounts, nCountsBest                                                                                      [32]int32
	vVisited                                                                                                  *Map_NodeVec_t
	mmNodes, mmCuts                                                                                           *extra.Extra_MmFixed_t
	uCanons                                                                                                   []uint16
	uPhases                                                                                                   [][]uint8
	pCounters                                                                                                 []uint8
	nChoiceNodes, nChoices, nCanons, nMatches, nPhases, nFanoutViolations                                     int32
	timeToMap, timeCuts, timeTruth, timeMatch, timeArea, timeSweep, timeToNet, timeTotal, time1, time2, time3 int64
}

type Map_SuperLibStruct_t_ struct {
	pName                                               string
	pGenlib                                             *mio.Mio_Library_t
	nVarsMax, nSupersAll, nSupersReal, nLines, fVerbose int32
	ppSupers                                            []*Map_Super_t
	tTableC, tTable                                     *Map_HashTable_t
	uTruths                                             [6][2]uint32
	uMask                                               [2]uint32
	pGateInv                                            *mio.Mio_Gate_t
	tDelayInv                                           Map_Time_t
	AreaInv, AreaBuf                                    float32
	pSuperInv                                           *Map_Super_t
	mmSupers, mmEntries                                 *extra.Extra_MmFixed_t
	mmForms                                             *extra.Extra_MmFlex_t
}

// Map_Node_t pointers remain ordinary pointers. Complemented references are
// represented by the explicit complement bit on wrapper nodes made by Map_Not.
type Map_NodeStruct_t_ struct {
	p                                    *Map_Man_t
	pNext                                *Map_Node_t
	Num, TravId, nRefs                   int32
	fMark0, fMark1, fUsed, fInv, fInvert uint32
	Level, NumTemp                       uint32
	nRefAct                              [3]int32
	nRefEst                              [3]float32
	Switching                            float32
	p1, p2, pNextE, pRepr                *Map_Node_t
	pFanPivot, pFanFanin1, pFanFanin2    *Map_Node_t
	fanouts                              []*Map_Node_t
	tArrival, tRequired                  [2]Map_Time_t
	pCutBest                             [2]*Map_Cut_t
	pCuts                                *Map_Cut_t
	pData0, pData1                       any
	complemented                         bool
	regular                              *Map_Node_t
	neg                                  *Map_Node_t
}

type Map_MatchStruct_t_ struct {
	pSupers            *Map_Super_t
	uPhase, uPhaseBest uint32
	pSuperBest         *Map_Super_t
	tArrive            Map_Time_t
	AreaFlow           float32
}

type Map_CutStruct_t_ struct {
	pNext, pOne, pTwo *Map_Cut_t
	ppLeaves          [6]*Map_Node_t
	uTruth            uint32
	nLeaves, nVolume  int32
	fMark, Phase      uint8
	M                 [2]Map_Match_t
	complemented      bool
	regular           *Map_Cut_t
	neg               *Map_Cut_t
}

type Map_SuperStruct_t_ struct {
	Num                                                            int32
	fSuper, fExclude, nFanins, nGates, nFanLimit, nSupers, nPhases uint32
	uPhases                                                        [4]uint8
	nUsed                                                          int32
	pFanins                                                        [6]*Map_Super_t
	pRoot                                                          *mio.Mio_Gate_t
	uTruth                                                         [2]uint32
	tDelaysR, tDelaysF                                             [6]Map_Time_t
	tDelayMax                                                      Map_Time_t
	Area                                                           float32
	pFormula                                                       string
	pNext                                                          *Map_Super_t
}

// The vector keeps one slice as the source of truth. nSize/nCap are retained
// because translated mapper callers currently inspect them directly.
type Map_NodeVecStruct_t_ struct {
	pArray      []*Map_Node_t
	nSize, nCap int32
}
type Map_HashTableStruct_t_ struct {
	pBins           []*Map_HashEntry_t
	nBins, nEntries int32
	mmMan           *extra.Extra_MmFixed_t
}
type Map_HashEntryStruct_t_ struct {
	uTruth [2]uint32
	uPhase uint32
	pGates *Map_Super_t
	pNext  *Map_HashEntry_t
}

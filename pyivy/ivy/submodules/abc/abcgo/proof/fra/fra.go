package fra

// FRAIG manager and parameters, ported from proof/fra/fra.h by Alan
// Mishchenko, UC Berkeley (2007). AIG objects use graph-owned edge IDs rather
// than C object pointers and their shared pData field.

import (
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/framgr"
	"github.com/glycerine/abcgo/sat/bsat"
)

type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type abctime = time.Duration

type Fra_Par_t = framgr.CombOptions

type Fra_Ssw_t struct {
	nPartSize, nOverSize, nFramesP, nFramesK, nMaxImps, nMaxLevs, nMinDomSize          int32
	fUseImps, fRewrite, fFraiging, fLatchCorr, fWriteImps, fUse1Hot, fVerbose, fSilent int32
	nIters                                                                             int32
	TimeLimit                                                                          float32
}

type Fra_Sec_t struct {
	fTryComb, fTryBmc, nFramesMax, nBTLimit, nBTLimitGlobal, nBTLimitInter    int32
	nBddVarsMax, nBddMax, nBddIterMax, nPdrTimeout                            int32
	fPhaseAbstract, fRetimeFirst, fRetimeRegs, fFraiging, fInduction          int32
	fInterpolation, fInterSeparate, fReachability, fReorderImage              int32
	fStopOnFirstFail, fUseNewProver, fUsePdr, fSilent, fVerbose, fVeryVerbose int32
	TimeLimit, fReadUnsolved, nSMnumber, fRecursive, fReportSolution          int32
}

type Fra_Cla_t = classPartition
type Fra_Sml_t = fraSimulation
type Fra_Bmc_t = Fra_Bmc_t_

type Fra_Man_t struct {
	pPars                                                                 *Fra_Par_t
	pManAig, pManFraig                                                    *aig.Graph
	nFramesAll, nSizeAlloc                                                int32
	pMemFraig                                                             []aig.Edge
	pCla                                                                  *classPartition
	pSml                                                                  *fraSimulation
	pBmc                                                                  *Fra_Bmc_t
	nPatWords                                                             int32
	pPatWords                                                             []uint32
	vCex, vOneHots                                                        *vec.Vec_Int_t
	pSat                                                                  *bsat.Solver
	nSatVars                                                              int32
	vPiVars                                                               *vec.Vec_Ptr_t
	nBTLimitGlobal, nInsLimitGlobal                                       int64
	pMemFanins                                                            [][]aig.Edge
	pMemSatNums                                                           []int32
	nMemAlloc                                                             int32
	vTimeouts                                                             *vec.Vec_Ptr_t
	nSimRounds, nNodesMiter, nLitsBeg, nLitsEnd, nNodesBeg, nNodesEnd     int32
	nRegsBeg, nRegsEnd, nSatCalls, nSatCallsSat, nSatCallsUnsat           int32
	nSatProof, nSatFails, nSatFailsReal, nSpeculs, nChoices, nChoicesFake int32
	nSatCallsRecent, nSatCallsSkipped                                     int32
	timeSim, timeTrav, timeRwr, timeSat, timeSatUnsat, timeSatSat         abctime
	timeSatFail, timeRef, timeTotal, time1, time2                         abctime
}

func Fra_ObjSim(s *Fra_Sml_t, id int32) []uint32 { return s.nodeWords(int(id)) }
func Fra_ObjRandomSim() uint32                   { return aig.Aig_ManRandom(0) }
func Fra_ObjFraig(m *Fra_Man_t, object aig.Edge, frame int32) aig.Edge {
	return m.pMemFraig[int32(object.NodeID())*m.nFramesAll+frame]
}
func Fra_ObjSetFraig(m *Fra_Man_t, object aig.Edge, frame int32, value aig.Edge) {
	m.pMemFraig[int32(object.NodeID())*m.nFramesAll+frame] = value
}
func Fra_ObjFaninVec(m *Fra_Man_t, object aig.Edge) []aig.Edge { return m.pMemFanins[object.NodeID()] }
func Fra_ObjSetFaninVec(m *Fra_Man_t, object aig.Edge, fanins []aig.Edge) {
	m.pMemFanins[object.NodeID()] = fanins
}
func Fra_ObjSatNum(m *Fra_Man_t, object aig.Edge) int32 { return m.pMemSatNums[object.NodeID()] }
func Fra_ObjSetSatNum(m *Fra_Man_t, object aig.Edge, number int32) {
	m.pMemSatNums[object.NodeID()] = number
}
func Fra_ClassObjRepr(m *Fra_Man_t, object aig.Edge) aig.Edge {
	repr := m.pCla.reprs[object.NodeID()]
	if repr < 0 {
		return aig.InvalidEdge
	}
	return aig.Edge(repr << 1)
}
func Fra_ClassObjSetRepr(m *Fra_Man_t, object, repr aig.Edge) {
	if repr == aig.InvalidEdge {
		m.pCla.reprs[object.NodeID()] = -1
	} else {
		m.pCla.reprs[object.NodeID()] = repr.NodeID()
	}
}
func Fra_ObjChild0Fra(m *Fra_Man_t, object aig.Edge, frame int32) aig.Edge {
	return fraMappedChild(m, m.pManAig.Node(object).Fanins[0], frame)
}
func Fra_ObjChild1Fra(m *Fra_Man_t, object aig.Edge, frame int32) aig.Edge {
	return fraMappedChild(m, m.pManAig.Node(object).Fanins[1], frame)
}
func fraMappedChild(m *Fra_Man_t, child aig.Edge, frame int32) aig.Edge {
	if child == aig.InvalidEdge {
		return child
	}
	result := Fra_ObjFraig(m, child, frame)
	if result != aig.InvalidEdge && child.Complemented() {
		result = result.Not()
	}
	return result
}
func Fra_ImpLeft(imp int32) int32           { return imp & 0xffff }
func Fra_ImpRight(imp int32) int32          { return imp >> 16 }
func Fra_ImpCreate(left, right int32) int32 { return right<<16 | left }

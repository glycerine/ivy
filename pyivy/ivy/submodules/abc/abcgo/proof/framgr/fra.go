package framgr

// FRAIG manager and parameters, ported from proof/fra/fra.h by Alan
// Mishchenko, UC Berkeley (2007). AIG objects use graph-owned edge IDs rather
// than C object pointers and their shared pData field.

import (
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type abctime = time.Duration

type Fra_Par_t = CombOptions

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
	PPars                                                                 *Fra_Par_t
	PManAig, PManFraig                                                    *aig.Graph
	NFramesAll, NSizeAlloc                                                int32
	PMemFraig                                                             []aig.Edge
	PCla                                                                  *classPartition
	PSml                                                                  *fraSimulation
	PBmc                                                                  *Fra_Bmc_t
	NPatWords                                                             int32
	PPatWords                                                             []uint32
	VCex, VOneHots                                                        *vec.Vec_Int_t
	PSat                                                                  *bsat.Solver
	NSatVars                                                              int32
	VPiVars                                                               *vec.Vec_Ptr_t
	NBTLimitGlobal, NInsLimitGlobal                                       int64
	PMemFanins                                                            [][]aig.Edge
	PMemSatNums                                                           []int32
	NMemAlloc                                                             int32
	VTimeouts                                                             *vec.Vec_Ptr_t
	NSimRounds, NNodesMiter, NLitsBeg, NLitsEnd, NNodesBeg, NNodesEnd     int32
	NRegsBeg, NRegsEnd, NSatCalls, NSatCallsSat, NSatCallsUnsat           int32
	NSatProof, NSatFails, NSatFailsReal, NSpeculs, NChoices, NChoicesFake int32
	NSatCallsRecent, NSatCallsSkipped                                     int32
	TimeSim, TimeTrav, TimeRwr, TimeSat, TimeSatUnsat, TimeSatSat         abctime
	TimeSatFail, TimeRef, TimeTotal, Time1, Time2                         abctime
}

func Fra_ObjSim(s *Fra_Sml_t, id int32) []uint32 { return s.NodeWords(int(id)) }
func Fra_ObjRandomSim() uint32                   { return aig.Aig_ManRandom(0) }
func Fra_ObjFraig(m *Fra_Man_t, object aig.Edge, frame int32) aig.Edge {
	return m.PMemFraig[int32(object.NodeID())*m.NFramesAll+frame]
}
func Fra_ObjSetFraig(m *Fra_Man_t, object aig.Edge, frame int32, value aig.Edge) {
	m.PMemFraig[int32(object.NodeID())*m.NFramesAll+frame] = value
}
func Fra_ObjFaninVec(m *Fra_Man_t, object aig.Edge) []aig.Edge { return m.PMemFanins[object.NodeID()] }
func Fra_ObjSetFaninVec(m *Fra_Man_t, object aig.Edge, fanins []aig.Edge) {
	m.PMemFanins[object.NodeID()] = fanins
}
func Fra_ObjSatNum(m *Fra_Man_t, object aig.Edge) int32 { return m.PMemSatNums[object.NodeID()] }
func Fra_ObjSetSatNum(m *Fra_Man_t, object aig.Edge, number int32) {
	m.PMemSatNums[object.NodeID()] = number
}
func Fra_ClassObjRepr(m *Fra_Man_t, object aig.Edge) aig.Edge {
	repr := m.PCla.Reprs[object.NodeID()]
	if repr < 0 {
		return aig.InvalidEdge
	}
	return aig.Edge(repr << 1)
}
func Fra_ClassObjSetRepr(m *Fra_Man_t, object, repr aig.Edge) {
	if repr == aig.InvalidEdge {
		m.PCla.Reprs[object.NodeID()] = -1
	} else {
		m.PCla.Reprs[object.NodeID()] = repr.NodeID()
	}
}
func Fra_ObjChild0Fra(m *Fra_Man_t, object aig.Edge, frame int32) aig.Edge {
	return fraMappedChild(m, m.PManAig.Node(object).Fanins[0], frame)
}
func Fra_ObjChild1Fra(m *Fra_Man_t, object aig.Edge, frame int32) aig.Edge {
	return fraMappedChild(m, m.PManAig.Node(object).Fanins[1], frame)
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

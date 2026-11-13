package ssw

// Optimized latch-correlation sweeping, ported from sswLcorr.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Ssw_ManSweepTransfer(p *Ssw_Man_t) {
	for _, obj := range p.pAig.Inputs() {
		mapped := Ssw_ObjFrame(p, obj, 0)
		if mapped == p.pFrames.Const0() {
			Ssw_SmlObjAssignConst(p.pSml, obj, 0, 0)
			continue
		}
		if mapped.Complemented() || !aig.Aig_ObjIsCi(p.pFrames, mapped) {
			panic("ssw: invalid latch simulation mapping")
		}
		info := vec.Vec_PtrEntry(p.vSimInfo, int(aig.Aig_ObjCioId(p.pFrames, mapped))).([]uint32)
		Ssw_SmlObjSetWord(p.pSml, obj, info[0], 0, 0)
	}
}
func Ssw_ManSweepResimulate(p *Ssw_Man_t) int32 {
	start := time.Now()
	Ssw_ManSweepTransfer(p)
	Ssw_SmlSimulateOneFrame(p.pSml)
	a, b := Ssw_ClassesRefineConst1(p.ppClasses, 1), Ssw_ClassesRefine(p.ppClasses, 1)
	vec.Vec_PtrCleanSimInfo(p.vSimInfo, 0, 1)
	p.nPatterns = 0
	p.nSimRounds++
	p.timeSimSat += time.Since(start)
	if a > 0 || b > 0 {
		return 1
	}
	return 0
}
func Ssw_SmlAddPattern(p *Ssw_Man_t, _, _ aig.Edge) {
	for _, item := range vec.Vec_PtrArray(p.pMSat.vUsedPis) {
		obj := item.(aig.Edge)
		variable := Ssw_ObjSatNum(p.pMSat, obj)
		if variable <= 0 {
			panic("ssw: missing SAT variable")
		}
		value, ok := p.pMSat.pSat.Value(variable)
		if !ok {
			panic("ssw: SAT model unavailable")
		}
		if p.pMSat.fPolarFlip != 0 && aig.Aig_ObjPhase(p.pFrames, obj) {
			value = !value
		}
		if value {
			info := vec.Vec_PtrEntry(p.vSimInfo, int(aig.Aig_ObjCioId(p.pFrames, obj))).([]uint32)
			util.Abc_InfoSetBit(info, p.nPatterns)
		}
	}
}
func Ssw_ManBuildCone_rec(p *Ssw_Man_t, obj aig.Edge) {
	if obj.Complemented() {
		panic("ssw: complemented cone node")
	}
	if Ssw_ObjFrame(p, obj, 0) != aig.InvalidEdge {
		return
	}
	n := p.pAig.Node(obj)
	if n.Kind != aig.And && n.Kind != aig.ExclusiveOr && n.Kind != aig.Buffer {
		panic("ssw: unmapped cone input")
	}
	Ssw_ManBuildCone_rec(p, n.Fanins[0]&^1)
	if n.Kind != aig.Buffer {
		Ssw_ManBuildCone_rec(p, n.Fanins[1]&^1)
	}
	Ssw_ObjSetFrame(p, obj, 0, sswMakeNode(p, p.pFrames, obj, 0))
}
func Ssw_ManSweepLatchOne(p *Ssw_Man_t, repr, obj aig.Edge) {
	if !aig.Aig_ObjIsCi(p.pAig, obj) || (!aig.Aig_ObjIsCi(p.pAig, repr) && repr != p.pAig.Const1()) {
		panic("ssw: invalid latch candidate")
	}
	if p.nCallsCount > 100 && p.nCallsUnsat < p.nCallsSat {
		p.nCallsDelta++
		if p.nCallsDelta < 0 {
			return
		}
	}
	p.nCallsDelta = 0
	start := time.Now()
	li := saig.Saig_ObjLoToLi(p.pAig, obj)
	Ssw_ManBuildCone_rec(p, p.pAig.Node(li).Fanins[0]&^1)
	candidate := Ssw_ObjChild0Fra(p, li, 0)
	representative := Ssw_ObjFrame(p, repr, 0)
	if aig.Aig_ObjIsCi(p.pAig, repr) {
		li = saig.Saig_ObjLoToLi(p.pAig, repr)
		Ssw_ManBuildCone_rec(p, p.pAig.Node(li).Fanins[0]&^1)
		representative = Ssw_ObjChild0Fra(p, li, 0)
	}
	p.timeReduce += time.Since(start)
	if candidate.NodeID() == representative.NodeID() {
		return
	}
	p.nRecycleCalls++
	p.nCallsCount++
	if (p.pAig.Node(obj).Phase == p.pAig.Node(repr).Phase) != (aig.Aig_ObjPhaseReal(p.pFrames, candidate) == aig.Aig_ObjPhaseReal(p.pFrames, representative)) {
		p.nPatterns++
		p.nStrangers++
		p.fRefined = 1
		return
	}
	switch Ssw_NodesAreEquiv(p, representative&^1, candidate&^1) {
	case 1:
		p.nCallsUnsat++
	case -1:
		Ssw_ClassesRemoveNode(p.ppClasses, obj)
		p.nCallsUnsat++
		p.fRefined = 1
	default:
		Ssw_SmlAddPattern(p, repr, obj)
		p.nPatterns++
		p.nCallsSat++
		p.fRefined = 1
	}
}
func Ssw_ManSweepLatch(p *Ssw_Man_t) int32 {
	p.pFrames = aig.NewGraph()
	p.pMSat.pAig = p.pFrames
	g := p.pAig
	inputs := g.Inputs()
	regs := int(g.RegisterCount())
	pis := len(inputs) - regs
	Ssw_ObjSetFrame(p, g.Const1(), 0, p.pFrames.Const1())
	for _, pi := range inputs[:pis] {
		Ssw_ObjSetFrame(p, pi, 0, p.pFrames.NewInput())
	}
	for _, lo := range inputs[pis:] {
		repr := aig.Aig_ObjRepr(g, lo)
		var mapped aig.Edge
		if repr == aig.InvalidEdge {
			mapped = p.pFrames.NewInput()
		} else {
			mapped = Ssw_ObjFrame(p, repr, 0)
			if g.Node(repr).Phase != g.Node(lo).Phase {
				mapped = mapped.Not()
			}
		}
		Ssw_ObjSetFrame(p, lo, 0, mapped)
	}
	if p.vSimInfo != nil {
		panic("ssw: simulation info already allocated")
	}
	p.vSimInfo = vec.Vec_PtrAllocSimInfo(p.pFrames.InputCount(), 1)
	class := vec.Vec_PtrAlloc(100)
	p.fRefined = 0
	p.nCallsCount = 0
	p.nCallsSat = 0
	p.nCallsUnsat = 0
	for _, lo := range inputs[pis:] {
		if Ssw_ObjIsConst1Cand(g, lo) != 0 {
			Ssw_ManSweepLatchOne(p, g.Const1(), lo)
		} else {
			Ssw_ClassesCollectClass(p.ppClasses, lo, class)
			for _, item := range vec.Vec_PtrArray(class) {
				candidate := item.(aig.Edge)
				if aig.Aig_ObjRepr(g, candidate) == lo {
					Ssw_ManSweepLatchOne(p, lo, candidate)
					if p.nPatterns == 32 {
						break
					}
				}
			}
		}
		if p.nPatterns == 32 {
			Ssw_ManSweepResimulate(p)
		}
		if p.pPars.nSatVarMax != 0 && p.pMSat.nSatVars > p.pPars.nSatVarMax && p.nRecycleCalls > p.pPars.nRecycleCalls {
			p.nVarsMax = max(p.nVarsMax, p.pMSat.nSatVars)
			p.nCallsMax = max(p.nCallsMax, p.pMSat.nSolverCalls)
			Ssw_SatStop(p.pMSat)
			p.pMSat = Ssw_SatStart(p.pPars.fPolarFlip)
			p.pMSat.pAig = p.pFrames
			p.nRecycles++
			p.nRecycleCalls = 0
		}
	}
	if p.nPatterns > 0 {
		Ssw_ManSweepResimulate(p)
	}
	return p.fRefined
}

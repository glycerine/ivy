package ssw

// Dynamic constraint loading and simulation, ported from sswDyn.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Ssw_ManLabelPiNodes(p *Ssw_Man_t) {
	g := p.pFrames
	aig.Aig_ObjSetMarkA(g, g.Const1())
	aig.Aig_ObjSetMarkB(g, g.Const1())
	pis := saig.Saig_ManPiNum(p.pAig)
	for f := int32(0); f < p.nFrames; f++ {
		for _, obj := range p.pAig.Inputs()[:pis] {
			mapped := Ssw_ObjFrame(p, obj, f)
			if !aig.Aig_ObjIsCi(g, mapped) || g.Node(mapped).MarkB {
				panic("ssw: invalid frame primary input")
			}
			aig.Aig_ObjSetMarkA(g, mapped)
			aig.Aig_ObjSetMarkB(g, mapped)
		}
	}
}
func Ssw_ManCollectPis_rec(g *aig.Graph, obj aig.Edge, newPis *vec.Vec_Ptr_t) {
	if obj.Complemented() {
		panic("ssw: complemented support object")
	}
	n := g.Node(obj)
	if n.MarkA {
		return
	}
	aig.Aig_ObjSetMarkA(g, obj)
	if aig.Aig_ObjIsCi(g, obj) {
		vec.Vec_PtrPush(newPis, obj)
		return
	}
	switch n.Kind {
	case aig.And, aig.ExclusiveOr:
		Ssw_ManCollectPis_rec(g, aig.Aig_Regular(n.Fanins[0]), newPis)
		Ssw_ManCollectPis_rec(g, aig.Aig_Regular(n.Fanins[1]), newPis)
	case aig.Buffer:
		Ssw_ManCollectPis_rec(g, aig.Aig_Regular(n.Fanins[0]), newPis)
	default:
		panic("ssw: invalid support object")
	}
}
func Ssw_ManCollectPos_rec(p *Ssw_Man_t, obj aig.Edge, newPos *vec.Vec_Int_t) {
	g := p.pFrames
	if obj.Complemented() {
		panic("ssw: complemented fanout object")
	}
	if g.Node(obj).MarkB {
		return
	}
	aig.Aig_ObjSetMarkB(g, obj)
	if int32(obj.NodeID()) > p.nSRMiterMaxId {
		return
	}
	if aig.Aig_ObjIsCo(g, obj) {
		id := aig.Aig_ObjCioId(g, obj)
		if id >= int32(g.OutputCount())-aig.Aig_ManRegNum(p.pAig) {
			return
		}
		vec.Vec_IntPush(newPos, id/2)
		return
	}
	for fan := aig.Aig_ObjFanout0Int(g, int32(obj.NodeID())); fan != 0; fan = aig.Aig_ObjFanoutNext(g, fan) {
		Ssw_ManCollectPos_rec(p, aig.Aig_ManObj(g, fan>>1), newPos)
	}
}
func Ssw_ManLoadSolver(p *Ssw_Man_t, repr, obj aig.Edge) {
	f := p.pPars.nFramesK
	if repr == obj {
		panic("ssw: duplicate comparison nodes")
	}
	a, b := aig.Aig_Regular(Ssw_ObjFrame(p, repr, f)), aig.Aig_Regular(Ssw_ObjFrame(p, obj, f))
	if a == b {
		panic("ssw: duplicate framed comparison nodes")
	}
	Ssw_CnfNodeAddToSolver(p.pMSat, a)
	Ssw_CnfNodeAddToSolver(p.pMSat, b)
	vec.Vec_PtrClear(p.vNewLos)
	Ssw_ManCollectPis_rec(p.pFrames, a, p.vNewLos)
	Ssw_ManCollectPis_rec(p.pFrames, b, p.vNewLos)
	vec.Vec_IntClear(p.vNewPos)
	for _, item := range vec.Vec_PtrArray(p.vNewLos) {
		Ssw_ManCollectPos_rec(p, item.(aig.Edge), p.vNewPos)
	}
	for _, id := range vec.Vec_IntArray(p.vNewPos) {
		x, y := aig.Aig_ManCo(p.pFrames, 2*id), aig.Aig_ManCo(p.pFrames, 2*id+1)
		if p.pFrames.Node(x).MarkB || p.pFrames.Node(y).MarkB {
			aig.Aig_ObjSetMarkB(p.pFrames, x)
			aig.Aig_ObjSetMarkB(p.pFrames, y)
			Ssw_NodesAreConstrained(p, aig.Aig_ObjChild0(p.pFrames, x), aig.Aig_ObjChild0(p.pFrames, y))
		}
	}
	if !p.pMSat.pSat.Simplify() {
		panic("ssw: inconsistent loaded constraints")
	}
}
func Ssw_ManSweepTransferDyn(p *Ssw_Man_t) {
	for _, obj := range p.pAig.Inputs() {
		mapped := Ssw_ObjFrame(p, obj, 0)
		if mapped == p.pFrames.Const0() {
			Ssw_SmlObjAssignConst(p.pSml, obj, 0, 0)
			continue
		}
		if mapped.Complemented() || !aig.Aig_ObjIsCi(p.pFrames, mapped) {
			panic("ssw: invalid simulation input")
		}
		info := vec.Vec_PtrEntry(p.vSimInfo, int(aig.Aig_ObjCioId(p.pFrames, mapped))).([]uint32)
		Ssw_SmlObjSetWord(p.pSml, obj, info[0], 0, 0)
	}
	pis := saig.Saig_ManPiNum(p.pAig)
	var f int32
	for f = 1; f < p.nFrames; f++ {
		for _, obj := range p.pAig.Inputs()[:pis] {
			mapped := Ssw_ObjFrame(p, obj, f)
			if mapped.Complemented() || !aig.Aig_ObjIsCi(p.pFrames, mapped) {
				panic("ssw: invalid simulation input")
			}
			info := vec.Vec_PtrEntry(p.vSimInfo, int(aig.Aig_ObjCioId(p.pFrames, mapped))).([]uint32)
			Ssw_SmlObjSetWord(p.pSml, obj, info[0], 0, f)
		}
	}
	for ; f < Ssw_SmlNumFrames(p.pSml); f++ {
		for _, obj := range p.pAig.Inputs()[:pis] {
			Ssw_SmlAssignRandomFrame(p.pSml, obj, f)
		}
	}
}
func Ssw_ManSweepResimulateDyn(p *Ssw_Man_t, f int32) int32 {
	start := time.Now()
	Ssw_ManSweepTransferDyn(p)
	Ssw_SmlSimulateOne(p.pSml)
	a := Ssw_ClassesRefineConst1(p.ppClasses, 1)
	b := Ssw_ClassesRefine(p.ppClasses, 1)
	vec.Vec_PtrCleanSimInfo(p.vSimInfo, 0, 1)
	p.nPatterns = 0
	p.nSimRounds++
	p.timeSimSat += time.Since(start)
	if a > 0 || b > 0 {
		return 1
	}
	return 0
}
func Ssw_ManSweepResimulateDynLocal(p *Ssw_Man_t, f int32) int32 {
	start := time.Now()
	p.nSimRounds++
	Ssw_ManSweepTransferDyn(p)
	vec.Vec_PtrClear(p.vResimConsts)
	vec.Vec_PtrClear(p.vResimClasses)
	aig.Aig_ManIncrementTravId(p.pAig)
	for i := p.iNodeStart; i < p.iNodeLast+p.pPars.nResimDelta && i < aig.Aig_ManObjNumMax(p.pAig); i++ {
		obj := aig.Aig_ManObj(p.pAig, i)
		if !p.pAig.ValidEdge(obj) {
			continue
		}
		if Ssw_ObjIsConst1Cand(p.pAig, obj) != 0 {
			vec.Vec_PtrPush(p.vResimConsts, obj)
			continue
		}
		repr := aig.Aig_ObjRepr(p.pAig, obj)
		if repr == aig.InvalidEdge || aig.Aig_ObjIsTravIdCurrent(p.pAig, repr) {
			continue
		}
		aig.Aig_ObjSetTravIdCurrent(p.pAig, repr)
		vec.Vec_PtrPush(p.vResimClasses, repr)
	}
	p.nVisCounter++
	for _, item := range vec.Vec_PtrArray(p.vResimConsts) {
		Ssw_SmlSimulateOneDyn_rec(p.pSml, item.(aig.Edge), p.nFrames-1, p.pVisited, p.nVisCounter)
	}
	for _, item := range vec.Vec_PtrArray(p.vResimClasses) {
		var n int32
		members := Ssw_ClassesReadClass(p.ppClasses, item.(aig.Edge), &n)
		for _, obj := range members[:n] {
			Ssw_SmlSimulateOneDyn_rec(p.pSml, obj, p.nFrames-1, p.pVisited, p.nVisCounter)
		}
	}
	a := Ssw_ClassesRefineConst1Group(p.ppClasses, p.vResimConsts, 1)
	var b int32
	for _, item := range vec.Vec_PtrArray(p.vResimClasses) {
		b += Ssw_ClassesRefineOneClass(p.ppClasses, item.(aig.Edge), 1)
	}
	vec.Vec_PtrCleanSimInfo(p.vSimInfo, 0, 1)
	p.nPatterns = 0
	p.nSimRounds++
	p.timeSimSat += time.Since(start)
	if a > 0 || b > 0 {
		return 1
	}
	return 0
}
func Ssw_ManSweepDyn(p *Ssw_Man_t) int32 {
	start := time.Now()
	p.pFrames = Ssw_FramesWithClasses(p)
	aig.Aig_ManFanoutStart(p.pFrames)
	p.nSRMiterMaxId = aig.Aig_ManObjNumMax(p.pFrames)
	f := p.pPars.nFramesK
	Ssw_ObjSetFrame(p, p.pAig.Const1(), f, p.pFrames.Const1())
	for _, obj := range p.pAig.Inputs()[:saig.Saig_ManPiNum(p.pAig)] {
		Ssw_ObjSetFrame(p, obj, f, p.pFrames.NewInput())
	}
	Ssw_ManLabelPiNodes(p)
	p.timeReduce += time.Since(start)
	if p.vSimInfo != nil {
		panic("ssw: simulation data already allocated")
	}
	p.vSimInfo = vec.Vec_PtrAllocSimInfo(p.pFrames.InputCount(), 1)
	p.fRefined = 0
	Ssw_ClassesClearRefined(p.ppClasses)
	p.iNodeStart = 0
	resim := func(i int32) {
		p.iNodeLast = i
		if p.pPars.fLocalSim != 0 {
			Ssw_ManSweepResimulateDynLocal(p, f)
		} else {
			Ssw_ManSweepResimulateDyn(p, f)
		}
		p.iNodeStart = i + 1
	}
	var i int32
	for i = 0; i < aig.Aig_ManObjNumMax(p.pAig); i++ {
		obj := aig.Aig_ManObj(p.pAig, i)
		if !p.pAig.ValidEdge(obj) {
			continue
		}
		if p.iNodeStart == 0 {
			p.iNodeStart = i
		}
		if saig.Saig_ObjIsLo(p.pAig, obj) {
			p.fRefined |= Ssw_ManSweepNode(p, obj, f, 0, nil)
		} else if aig.Aig_ObjIsNode(p.pAig, obj) {
			Ssw_ObjSetFrame(p, obj, f, sswMakeNode(p, p.pFrames, obj, f))
			p.fRefined |= Ssw_ManSweepNode(p, obj, f, 0, nil)
		}
		if p.pMSat.pSat == nil || (p.pPars.nSatVarMax2 != 0 && p.pMSat.nSatVars > p.pPars.nSatVarMax2 && p.nRecycleCalls > p.pPars.nRecycleCalls2) {
			if p.nPatterns > 0 {
				resim(i)
			}
			aig.Aig_ManCleanMarkAB(p.pFrames)
			Ssw_ManLabelPiNodes(p)
			if p.pMSat != nil {
				p.nVarsMax = max(p.nVarsMax, p.pMSat.nSatVars)
				p.nCallsMax = max(p.nCallsMax, p.pMSat.nSolverCalls)
				Ssw_SatStop(p.pMSat)
				p.nRecycles++
				p.nRecyclesTotal++
				p.nRecycleCalls = 0
			}
			p.pMSat = Ssw_SatStart(0)
			p.pMSat.pAig = p.pFrames
			if p.nPatterns != 0 {
				panic("ssw: pending simulation patterns")
			}
		}
		if p.nPatterns == 32 {
			resim(i)
		}
	}
	if p.nPatterns > 0 {
		resim(i)
	}
	return p.fRefined
}

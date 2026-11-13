package ssw

// Signal filtering with sequential simulation and SAT sweeping, from sswFilter.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
)

func sswFilterSimulateFrame(p *Ssw_Man_t, piValue func(int) bool) {
	g := p.pAig
	pis := int(saig.Saig_ManPiNum(g))
	pos := int(saig.Saig_ManPoNum(g))
	inputs, outputs := g.Inputs(), g.Outputs()
	sswSetMark(g, g.Const1(), true)
	for i, pi := range inputs[:pis] {
		sswSetMark(g, pi, piValue(i))
	}
	for i, lo := range inputs[pis:] {
		sswSetMark(g, lo, g.Node(outputs[pos+i]).MarkB)
	}
	for _, obj := range aig.Aig_ManDfsAll(g) {
		n := g.Node(obj)
		switch n.Kind {
		case aig.And:
			sswSetMark(g, obj, sswMarkValue(g, n.Fanins[0]) && sswMarkValue(g, n.Fanins[1]))
		case aig.ExclusiveOr:
			sswSetMark(g, obj, sswMarkValue(g, n.Fanins[0]) != sswMarkValue(g, n.Fanins[1]))
		case aig.Buffer:
			sswSetMark(g, obj, sswMarkValue(g, n.Fanins[0]))
		}
	}
	for _, out := range outputs {
		sswSetMark(g, out, sswMarkValue(g, g.Node(out).Fanins[0]))
	}
}
func sswFilterInitState(p *Ssw_Man_t) {
	g := p.pAig
	pis := saig.Saig_ManPiNum(g)
	pos := saig.Saig_ManPoNum(g)
	for i, li := range g.Outputs()[pos:] {
		sswSetMark(g, li, util.Abc_InfoHasBit(p.pPatWords, pis+int32(i)))
	}
}
func Ssw_ManRefineByFilterSim(p *Ssw_Man_t, frames int32) {
	if frames <= 0 {
		panic("ssw: invalid filter frame count")
	}
	sswFilterInitState(p)
	for f := int32(0); f < frames; f++ {
		sswFilterSimulateFrame(p, func(int) bool { return false })
		if f == 0 {
			for _, obj := range p.pAig.LiveEdges() {
				aig.Aig_ObjSetPhase(p.pAig, obj, p.pAig.Node(obj).MarkB)
			}
		} else {
			Ssw_ClassesRefineConst1(p.ppClasses, 0)
			Ssw_ClassesRefine(p.ppClasses, 0)
		}
	}
}
func Ssw_ManRollForward(p *Ssw_Man_t, frames int32) {
	if frames <= 0 {
		panic("ssw: invalid filter frame count")
	}
	sswFilterInitState(p)
	for f := int32(0); f < frames; f++ {
		sswFilterSimulateFrame(p, func(int) bool { return aig.Aig_ManRandom(0)&1 != 0 })
	}
	pis := saig.Saig_ManPiNum(p.pAig)
	pos := saig.Saig_ManPoNum(p.pAig)
	for i, li := range p.pAig.Outputs()[pos:] {
		idx := pis + int32(i)
		if p.pAig.Node(li).MarkB != util.Abc_InfoHasBit(p.pPatWords, idx) {
			util.Abc_InfoXorBit(p.pPatWords, idx)
		}
	}
}
func Ssw_ManFindStartingState(p *Ssw_Man_t, cex *util.Abc_Cex_t) {
	g := p.pAig
	pis := saig.Saig_ManPiNum(g)
	pos := saig.Saig_ManPoNum(g)
	for _, li := range g.Outputs()[pos:] {
		sswSetMark(g, li, false)
	}
	bit := cex.NRegs
	for f := int32(0); f <= cex.IFrame; f++ {
		sswFilterSimulateFrame(p, func(int) bool { v := util.Abc_InfoHasBit(cex.PData, bit); bit++; return v })
	}
	if bit != cex.NBits {
		panic("ssw: invalid counterexample dimensions")
	}
	for i, lo := range g.Inputs()[pis:] {
		idx := pis + int32(i)
		if g.Node(lo).MarkB != util.Abc_InfoHasBit(p.pPatWords, idx) {
			util.Abc_InfoXorBit(p.pPatWords, idx)
		}
	}
}
func Ssw_ManSweepNodeFilter(p *Ssw_Man_t, obj aig.Edge, frame int32) int32 {
	repr := aig.Aig_ObjRepr(p.pAig, obj)
	if repr == aig.InvalidEdge {
		return 0
	}
	node, rep := Ssw_ObjFrame(p, obj, frame), Ssw_ObjFrame(p, repr, frame)
	if node == aig.InvalidEdge {
		panic("ssw: missing framed candidate")
	}
	if rep == aig.InvalidEdge {
		panic("ssw: missing framed representative")
	}
	phase := aig.Aig_ObjPhase(p.pAig, obj) != aig.Aig_ObjPhase(p.pAig, repr)
	if (aig.Aig_ObjPhaseReal(p.pFrames, node) != aig.Aig_ObjPhaseReal(p.pFrames, rep)) != phase {
		panic("ssw: representative phase mismatch")
	}
	if aig.Aig_Regular(node) == aig.Aig_Regular(rep) {
		return 0
	}
	a, b := aig.Aig_Regular(rep), aig.Aig_Regular(node)
	if b == p.pFrames.Const1() {
		a, b = b, a
	}
	status := Ssw_NodesAreEquiv(p, a, b)
	if status == 1 {
		if phase {
			rep = rep.Not()
		}
		Ssw_ObjSetFrame(p, obj, frame, rep)
		return 0
	}
	if status == -1 {
		return 1
	}
	Ssw_SmlSavePatternAig(p, frame)
	Ssw_ManResimulateBit(p, obj, repr)
	if aig.Aig_ObjRepr(p.pAig, obj) == repr {
		panic("ssw: failed to refine representative")
	}
	return 0
}
func Ssw_ManSweepBmcFilter_rec(p *Ssw_Man_t, obj aig.Edge, frame int32) aig.Edge {
	if x := Ssw_ObjFrame(p, obj, frame); x != aig.InvalidEdge {
		return x
	}
	var result aig.Edge
	if saig.Saig_ObjIsLo(p.pAig, obj) {
		if frame <= 0 {
			panic("ssw: missing initial latch value")
		}
		li := saig.Saig_ObjLoToLi(p.pAig, obj)
		child := aig.Aig_ObjChild0(p.pAig, li)
		result = Ssw_ManSweepBmcFilter_rec(p, aig.Aig_Regular(child), frame-1)
		if child.Complemented() {
			result = result.Not()
		}
	} else {
		if !aig.Aig_ObjIsNode(p.pAig, obj) {
			panic("ssw: unexpected filter object")
		}
		Ssw_ManSweepBmcFilter_rec(p, aig.Aig_ObjFanin0(p.pAig, obj), frame)
		Ssw_ManSweepBmcFilter_rec(p, aig.Aig_ObjFanin1(p.pAig, obj), frame)
		result = sswMakeNode(p, p.pFrames, obj, frame)
	}
	Ssw_ObjSetFrame(p, obj, frame, result)
	return result
}
func sswFilterGrowFrames(p *Ssw_Man_t) {
	old := p.nFrames
	next := old * 2
	if next <= old {
		panic("ssw: invalid frame mapping capacity")
	}
	mapping := make([]aig.Edge, int(aig.Aig_ManObjNumMax(p.pAig))*int(next))
	for i := range mapping {
		mapping[i] = aig.InvalidEdge
	}
	for _, obj := range p.pAig.LiveEdges() {
		for f := int32(0); f < old; f++ {
			mapping[obj.NodeID()*int(next)+int(f)] = Ssw_ObjFrame(p, obj, f)
		}
	}
	p.pNodeToFrames = mapping
	p.nFrames = next
}
func Ssw_ManSweepBmcFilter(p *Ssw_Man_t, timeLimit int32) int32 {
	var started int64
	if timeLimit != 0 {
		defer sswPinClock()()
		started = sswCPUClock()
	}
	g := p.pAig
	pis := saig.Saig_ManPiNum(g)
	pos := saig.Saig_ManPoNum(g)
	p.pFrames = aig.Aig_ManStart(aig.Aig_ManObjNumMax(g) * p.pPars.nFramesK)
	p.pMSat.pAig = p.pFrames
	for i, lo := range g.Inputs()[pis:] {
		v := p.pFrames.Const0()
		if util.Abc_InfoHasBit(p.pPatWords, pis+int32(i)) {
			v = p.pFrames.Const1()
		}
		Ssw_ObjSetFrame(p, lo, 0, v)
	}
	for f := int32(0); f < p.pPars.nFramesK; f++ {
		if f == p.nFrames-1 {
			sswFilterGrowFrames(p)
		}
		Ssw_ObjSetFrame(p, g.Const1(), f, p.pFrames.Const1())
		for _, pi := range g.Inputs()[:pis] {
			Ssw_ObjSetFrame(p, pi, f, p.pFrames.NewInput())
		}
		timedOut := false
		for _, obj := range sswFrameNodes(g) {
			Ssw_ObjSetFrame(p, obj, f, sswMakeNode(p, p.pFrames, obj, f))
			if Ssw_ManSweepNodeFilter(p, obj, f) != 0 {
				timedOut = true
				break
			}
		}
		if p.pPars.fVerbose != 0 {
			fmt.Printf("Frame %4d : ", f)
			Ssw_ClassesPrint(p.ppClasses, 0)
		}
		if timedOut || f == p.pPars.nFramesK-1 || (timeLimit != 0 && float32(timeLimit) <= float32(sswCPUClock()-started)/float32(sswTicksPerSecond)) {
			break
		}
		for _, out := range g.Outputs() {
			Ssw_ObjSetFrame(p, out, f, Ssw_ObjChild0Fra(p, out, f))
		}
		for i, lo := range g.Inputs()[pis:] {
			next := Ssw_ObjFrame(p, g.Outputs()[int(pos)+i], f)
			Ssw_ObjSetFrame(p, lo, f+1, next)
			Ssw_CnfNodeAddToSolver(p.pMSat, aig.Aig_Regular(next))
		}
	}
	return 1
}

// Ssw_SignalFilter retains ABC's integer-second time budget behavior: a
// remaining fraction of a second rounds to zero, which means no per-sweep
// limit, and that CPU-time limit is not passed to individual SAT calls.
func Ssw_SignalFilter(g *aig.Graph, frames, conflicts, rounds, timeLimit, timeLimit2 int32, cex *util.Abc_Cex_t, latchOnly, verbose int32) {
	var deadline int64
	if timeLimit != 0 || timeLimit2 != 0 {
		defer sswPinClock()()
	}
	if timeLimit != 0 {
		deadline = int64(timeLimit)*sswTicksPerSecond + sswCPUClock()
	}
	if aig.Aig_ManRegNum(g) <= 0 || aig.Aig_ManConstrNum(g) != 0 {
		panic("ssw: invalid signal filter graph")
	}
	if aig.Aig_ManNodeNum(g) == 0 {
		return
	}
	aig.Aig_ManRandom(1)
	pars := &Ssw_Pars_t{}
	Ssw_ManSetDefaultParams(pars)
	pars.nFramesK = 3
	pars.nBTLimit = conflicts
	pars.TimeLimit = timeLimit
	pars.fVerbose = verbose
	p := Ssw_ManCreate(g, pars)
	pars.nFramesK = frames
	if aig.Aig_ManHasReprs(g) {
		p.ppClasses = Ssw_ClassesPrepareFromReprs(g)
	} else {
		p.ppClasses = Ssw_ClassesPrepareSimple(g, latchOnly, 0)
	}
	Ssw_ClassesSetData(p.ppClasses, g, nil, Ssw_SmlObjIsConstBit, Ssw_SmlObjsAreEqualBit)
	if p.vInits != nil {
		panic("ssw: unexpected filter initial values")
	}
	if cex != nil {
		Ssw_ManFindStartingState(p, cex)
	}
	for r := int32(0); r < rounds; r++ {
		if verbose != 0 {
			fmt.Printf("Round %3d:\n", r)
		}
		Ssw_ManRefineByFilterSim(p, frames)
		if Ssw_ClassesCand1Num(p.ppClasses) == 0 && Ssw_ClassesClassNum(p.ppClasses) == 0 {
			fmt.Println("All equivalences are refined away.")
			break
		}
		if verbose != 0 {
			fmt.Print("Initial    : ")
			Ssw_ClassesPrint(p.ppClasses, 0)
		}
		p.pMSat = Ssw_SatStart(0)
		part := sswFilterPartLimit(timeLimit, timeLimit2, deadline, sswCPUClockIfTimed(timeLimit))
		Ssw_ManSweepBmcFilter(p, part)
		Ssw_SatStop(p.pMSat)
		p.pMSat = nil
		Ssw_ManCleanup(p)
		Ssw_ManRollForward(p, frames)
		if timeLimit != 0 && sswCPUClock() > deadline {
			fmt.Printf("Reached timeout (%d seconds).\n", timeLimit)
			break
		}
	}
	aig.Aig_ManSetPhase(g)
	aig.Aig_ManCleanMarkB(g)
	pars.fVerbose = 0
	Ssw_ManStop(p)
}

func sswCPUClockIfTimed(limit int32) int64 {
	if limit == 0 {
		return 0
	}
	return sswCPUClock()
}

// sswFilterPartLimit retains C's signed integer truncation and zero handling.
func sswFilterPartLimit(limit, limit2 int32, deadline, now int64) int32 {
	part := int32(0)
	if limit != 0 {
		part = int32((deadline - now) / sswTicksPerSecond)
	}
	if limit2 != 0 {
		if part != 0 {
			part = min(part, limit2)
		} else {
			part = limit2
		}
	}
	return part
}
func Ssw_SignalFilterGia(g *gia.Gia_Man_t, frames, conflicts, rounds, timeLimit, timeLimit2 int32, cex *util.Abc_Cex_t, latchOnly, verbose int32) {
	graph := gia.Gia_ManToAigSimple(g)
	if g.HasRepresentatives() {
		gia.Gia_ManReprToAigRepr2(graph, g)
		g.ClearRepresentatives()
	}
	Ssw_SignalFilter(graph, frames, conflicts, rounds, timeLimit, timeLimit2, cex, latchOnly, verbose)
	gia.Gia_ManReprFromAigRepr(graph, g)
	aig.Aig_ManStop(graph)
}

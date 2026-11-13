package ssw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func sswConstrInitFrames(p *Ssw_Man_t) {
	p.pFrames = aig.NewGraph()
	p.pMSat.pAig = p.pFrames
	inputs := p.pAig.Inputs()
	regs := int(p.pAig.RegisterCount())
	for _, lo := range inputs[len(inputs)-regs:] {
		Ssw_ObjSetFrame(p, lo, 0, p.pFrames.Const0())
	}
}
func sswConstrFrameInputs(p *Ssw_Man_t, frame int32, index *int) {
	Ssw_ObjSetFrame(p, p.pAig.Const1(), frame, p.pFrames.Const1())
	inputs := p.pAig.Inputs()
	pis := len(inputs) - int(p.pAig.RegisterCount())
	for _, pi := range inputs[:pis] {
		mapped := p.pFrames.NewInput()
		aig.Aig_ObjSetPhase(p.pFrames, mapped, vec.Vec_IntEntry(p.vInits, *index) != 0)
		*index++
		Ssw_ObjSetFrame(p, pi, frame, mapped)
	}
}
func sswConstrFrameAssumptions(p *Ssw_Man_t, frame int32) {
	outputs := p.pAig.Outputs()
	pos := len(outputs) - int(p.pAig.RegisterCount())
	count := int(aig.Aig_ManConstrNum(p.pAig))
	for _, out := range outputs[pos-count : pos] {
		driver := p.pAig.Node(out).Fanins[0]
		mapped := Ssw_ManSweepBmcConstr_rec(p, driver&^1, frame)
		if driver.Complemented() {
			mapped = mapped.Not()
		}
		if mapped.NodeID() == 0 {
			if mapped != p.pFrames.Const0() {
				panic("ssw: constraint violates reference phase")
			}
			continue
		}
		Ssw_NodesAreConstrained(p, mapped, p.pFrames.Const0())
	}
}
func sswConstrSweepGates(p *Ssw_Man_t, frame int32) {
	for _, obj := range sswFrameNodes(p.pAig) {
		Ssw_ObjSetFrame(p, obj, frame, sswMakeNode(p, p.pFrames, obj, frame))
		if p.pAig.Node(obj).Kind != aig.Buffer {
			p.fRefined |= Ssw_ManSweepNodeConstr(p, obj, frame, 1)
		}
	}
}
func sswConstrTransfer(p *Ssw_Man_t, frame int32) {
	inputs, outputs := p.pAig.Inputs(), p.pAig.Outputs()
	regs := int(p.pAig.RegisterCount())
	pos := len(outputs) - regs
	for _, out := range outputs {
		Ssw_ObjSetFrame(p, out, frame, Ssw_ObjChild0Fra(p, out, frame))
	}
	for i, lo := range inputs[len(inputs)-regs:] {
		mapped := Ssw_ObjFrame(p, outputs[pos+i], frame)
		Ssw_ObjSetFrame(p, lo, frame+1, mapped)
		Ssw_CnfNodeAddToSolver(p.pMSat, mapped&^1)
	}
}
func sswConstrCheckInitCount(p *Ssw_Man_t, index int) {
	pis := len(p.pAig.Inputs()) - int(p.pAig.RegisterCount())
	if vec.Vec_IntSize(p.vInits) != index+pis {
		panic("ssw: incorrect initial phase count")
	}
}
func Ssw_ManSweepBmcConstr(p *Ssw_Man_t) int32 {
	start := time.Now()
	sswConstrInitFrames(p)
	p.fRefined = 0
	index := 0
	inputs := p.pAig.Inputs()
	regs := int(p.pAig.RegisterCount())
	for f := int32(0); f < p.pPars.nFramesK; f++ {
		if sswDeadlineExpired(p.pPars) {
			return p.fRefined
		}
		sswConstrFrameInputs(p, f, &index)
		sswConstrFrameAssumptions(p, f)
		for _, lo := range inputs[len(inputs)-regs:] {
			p.fRefined |= Ssw_ManSweepNodeConstr(p, lo, f, 1)
		}
		sswConstrSweepGates(p, f)
		if f+1 < p.pPars.nFramesK {
			sswConstrTransfer(p, f)
		}
	}
	sswConstrCheckInitCount(p, index)
	p.timeBmc += time.Since(start)
	return p.fRefined
}
func Ssw_ManSweepBmcConstr_old(p *Ssw_Man_t) int32 {
	start := time.Now()
	sswConstrInitFrames(p)
	index := 0
	for f := int32(0); f < p.pPars.nFramesK; f++ {
		if sswDeadlineExpired(p.pPars) {
			return p.fRefined
		}
		sswConstrFrameInputs(p, f, &index)
		sswConstrFrameAssumptions(p, f)
	}
	sswConstrCheckInitCount(p, index)
	p.fRefined = 0
	for f := int32(0); f < p.pPars.nFramesK; f++ {
		if sswDeadlineExpired(p.pPars) {
			return p.fRefined
		}
		sswConstrSweepGates(p, f)
		if f+1 < p.pPars.nFramesK {
			sswConstrTransfer(p, f)
		}
	}
	p.timeBmc += time.Since(start)
	return p.fRefined
}
func Ssw_ManSweepConstr(p *Ssw_Man_t) int32 {
	start := time.Now()
	p.pFrames = Ssw_FramesWithClasses(p)
	if p.pFrames == nil {
		return p.fRefined
	}
	p.pMSat.pAig = p.pFrames
	graph := p.pAig
	inputs, outputs := graph.Inputs(), graph.Outputs()
	regs := int(graph.RegisterCount())
	pis, pos := len(inputs)-regs, len(outputs)-regs
	frameOutputs := p.pFrames.Outputs()
	pairs := len(frameOutputs) - regs
	if pairs < 0 || pairs%2 != 0 {
		panic("ssw: malformed equivalence constraints")
	}
	for i := 0; i < pairs; i += 2 {
		Ssw_NodesAreConstrained(p, p.pFrames.Node(frameOutputs[i]).Fanins[0], p.pFrames.Node(frameOutputs[i+1]).Fanins[0])
	}
	for _, out := range frameOutputs[pairs:] {
		Ssw_CnfNodeAddToSolver(p.pMSat, p.pFrames.Node(out).Fanins[0]&^1)
	}
	f := p.pPars.nFramesK
	index := int(f) * pis
	sswConstrFrameInputs(p, f, &index)
	if vec.Vec_IntSize(p.vInits) != index {
		panic("ssw: incorrect initial phase count")
	}
	p.timeReduce += time.Since(start)
	nConstraints := int(aig.Aig_ManConstrNum(graph))
	for frame := int32(0); frame <= f; frame++ {
		if sswDeadlineExpired(p.pPars) {
			return p.fRefined
		}
		for _, out := range outputs[pos-nConstraints : pos] {
			driver := graph.Node(out).Fanins[0]
			Ssw_FramesWithClasses_rec(p, driver&^1, frame)
			mapped := Ssw_ObjChild0Fra(p, out, frame)
			if mapped == p.pFrames.Const0() {
				continue
			}
			if mapped == p.pFrames.Const1() {
				fmt.Println("Polarity violation.")
				continue
			}
			Ssw_NodesAreConstrained(p, mapped, p.pFrames.Const0())
		}
	}
	p.pMSat.pSat.Simplify()
	p.fRefined = 0
	Ssw_ClassesClearRefined(p.ppClasses)
	for _, lo := range inputs[pis:] {
		p.fRefined |= Ssw_ManSweepNodeConstr(p, lo, f, 0)
	}
	for _, obj := range sswFrameNodes(graph) {
		Ssw_ObjSetFrame(p, obj, f, sswMakeNode(p, p.pFrames, obj, f))
		if graph.Node(obj).Kind != aig.Buffer {
			p.fRefined |= Ssw_ManSweepNodeConstr(p, obj, f, 0)
		}
	}
	return p.fRefined
}

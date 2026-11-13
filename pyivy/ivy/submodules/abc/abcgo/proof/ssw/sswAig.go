package ssw

// AIG frame construction and speculative reduction, ported from sswAig.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ssw_FrmStart(graph *aig.Graph) *Ssw_Frm_t {
	p := &Ssw_Frm_t{pAig: graph, nObjs: aig.Aig_ManObjNumMax(graph), vAig2Frm: vec.Vec_PtrAlloc(0)}
	vec.Vec_PtrFill(p.vAig2Frm, int(2*p.nObjs), nil)
	return p
}
func Ssw_FrmStop(p *Ssw_Frm_t) {
	if p == nil {
		return
	}
	if p.pFrames != nil {
		aig.Aig_ManStop(p.pFrames)
	}
	*p = Ssw_Frm_t{}
}

func Ssw_FramesConstrainNode(p *Ssw_Man_t, frames, graph *aig.Graph, obj aig.Edge, frame int32, twoPos int32) {
	repr := aig.Aig_ObjRepr(graph, obj)
	if repr == aig.InvalidEdge {
		return
	}
	p.nConstrTotal++
	if repr.NodeID() >= obj.NodeID() {
		panic("ssw: representative is not earlier than object")
	}
	actual, representative := Ssw_ObjFrame(p, obj, frame), Ssw_ObjFrame(p, repr, frame)
	if actual == aig.InvalidEdge {
		panic("ssw: missing candidate frame mapping")
	}
	if representative == aig.InvalidEdge {
		// ID order in a rewired graph can precede its representative's
		// topological position; do not assume the candidate is equivalent.
		if p.ppClasses != nil {
			Ssw_ClassesRemoveNode(p.ppClasses, obj)
		} else {
			aig.Aig_ObjClearRepr(graph, obj)
		}
		return
	}
	differentPhase := aig.Aig_ObjPhase(graph, obj) != aig.Aig_ObjPhase(graph, repr)
	reduced := representative
	if differentPhase {
		reduced = reduced.Not()
	}
	if actual == reduced.Not() {
		panic("ssw: inconsistent speculative phases")
	}
	if actual == reduced {
		return
	}
	p.nConstrReduced++
	Ssw_ObjSetFrame(p, obj, frame, reduced)
	if twoPos != 0 {
		frames.AddOutput(reduced)
		frames.AddOutput(actual)
	} else {
		miter := frames.XOR(actual, reduced)
		if aig.Aig_ObjPhaseReal(frames, miter) {
			miter = miter.Not()
		}
		frames.AddOutput(miter)
	}
}

// sswFrameNodes visits all live internal objects in fanin-first order.
func sswFrameNodes(graph *aig.Graph) []aig.Edge {
	var nodes []aig.Edge
	for _, e := range aig.Aig_ManDfsAll(graph) {
		if aig.Aig_ObjIsNode(graph, e) || graph.Node(e).Kind == aig.Buffer {
			nodes = append(nodes, e)
		}
	}
	return nodes
}
func sswCheckFrames(p *Ssw_Man_t) {
	if p.pFrames != nil || aig.Aig_ManRegNum(p.pAig) <= 0 || int(aig.Aig_ManRegNum(p.pAig)) >= p.pAig.InputCount() {
		panic("ssw: invalid frame construction state")
	}
	p.nConstrTotal = 0
	p.nConstrReduced = 0
}
func sswMakeNode(p *Ssw_Man_t, frames *aig.Graph, obj aig.Edge, frame int32) aig.Edge {
	a := Ssw_ObjChild0Fra(p, obj, frame)
	if p.pAig.Node(obj).Kind == aig.Buffer {
		return a
	}
	b := Ssw_ObjChild1Fra(p, obj, frame)
	if p.pAig.Node(obj).Kind == aig.ExclusiveOr {
		return frames.NewXor(a, b)
	}
	return frames.AND(a, b)
}
func Ssw_FramesWithClasses(p *Ssw_Man_t) *aig.Graph {
	if sswDeadlineExpired(p.pPars) {
		return nil
	}
	sswCheckFrames(p)
	graph := p.pAig
	frames := aig.Aig_ManStart(aig.Aig_ManObjNumMax(graph) * p.nFrames)
	abort := func() *aig.Graph {
		aig.Aig_ManStop(frames)
		for i := range p.pNodeToFrames {
			p.pNodeToFrames[i] = aig.InvalidEdge
		}
		return nil
	}
	inputs, outputs := graph.Inputs(), graph.Outputs()
	regs := int(aig.Aig_ManRegNum(graph))
	pis := len(inputs) - regs
	pos := len(outputs) - regs
	for _, lo := range inputs[pis:] {
		Ssw_ObjSetFrame(p, lo, 0, frames.NewInput())
	}
	nodes := sswFrameNodes(graph)
	iLits := 0
	for f := int32(0); f < p.pPars.nFramesK; f++ {
		if sswDeadlineExpired(p.pPars) {
			return abort()
		}
		Ssw_ObjSetFrame(p, graph.Const1(), f, frames.Const1())
		for _, pi := range inputs[:pis] {
			next := frames.NewInput()
			if p.vInits != nil {
				aig.Aig_ObjSetPhase(frames, next, vec.Vec_IntEntry(p.vInits, iLits) != 0)
				iLits++
			}
			Ssw_ObjSetFrame(p, pi, f, next)
		}
		for _, lo := range inputs[pis:] {
			Ssw_FramesConstrainNode(p, frames, graph, lo, f, 1)
		}
		for i, obj := range nodes {
			if i&255 == 0 && sswDeadlineExpired(p.pPars) {
				return abort()
			}
			Ssw_ObjSetFrame(p, obj, f, sswMakeNode(p, frames, obj, f))
			Ssw_FramesConstrainNode(p, frames, graph, obj, f, 1)
		}
		for _, out := range outputs {
			Ssw_ObjSetFrame(p, out, f, Ssw_ObjChild0Fra(p, out, f))
		}
		for i, lo := range inputs[pis:] {
			Ssw_ObjSetFrame(p, lo, f+1, Ssw_ObjFrame(p, outputs[pos+i], f))
		}
	}
	if p.vInits != nil && vec.Vec_IntSize(p.vInits) != iLits+pis {
		panic("ssw: incorrect initial phase count")
	}
	for _, lo := range inputs[pis:] {
		frames.AddOutput(Ssw_ObjFrame(p, lo, p.pPars.nFramesK))
	}
	aig.Aig_ManCleanup(frames)
	if frames.Data != nil {
		panic("ssw: unexpected assignment on frame graph")
	}
	return frames
}
func Ssw_SpeculativeReduction(p *Ssw_Man_t) *aig.Graph {
	sswCheckFrames(p)
	graph := p.pAig
	frames := aig.Aig_ManStart(aig.Aig_ManObjNumMax(graph) * p.nFrames)
	frames.Name = graph.Name
	inputs, outputs := graph.Inputs(), graph.Outputs()
	regs := int(aig.Aig_ManRegNum(graph))
	pis := len(inputs) - regs
	pos := len(outputs) - regs
	Ssw_ObjSetFrame(p, graph.Const1(), 0, frames.Const1())
	for _, obj := range inputs {
		Ssw_ObjSetFrame(p, obj, 0, frames.NewInput())
	}
	for _, lo := range inputs[pis:] {
		Ssw_FramesConstrainNode(p, frames, graph, lo, 0, 0)
	}
	for _, obj := range sswFrameNodes(graph) {
		Ssw_ObjSetFrame(p, obj, 0, sswMakeNode(p, frames, obj, 0))
		Ssw_FramesConstrainNode(p, frames, graph, obj, 0, 0)
	}
	for _, li := range outputs[pos:] {
		frames.AddOutput(Ssw_ObjChild0Fra(p, li, 0))
	}
	aig.Aig_ManCleanup(frames)
	aig.Aig_ManSetRegNum(frames, int32(regs))
	return frames
}

package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"time"
)

func Ssw_SmlSimulateOneDyn_rec(p *Ssw_Sml_t, obj aig.Edge, frame int32, visited []int32, visitID int32) {
	index := int(p.nFrames)*obj.NodeID() + int(frame)
	if visited[index] == visitID {
		return
	}
	visited[index] = visitID
	if saig.Saig_ObjIsPi(p.pAig, obj) || obj == p.pAig.Const1() {
		return
	}
	if saig.Saig_ObjIsLo(p.pAig, obj) {
		if frame == 0 {
			return
		}
		li := saig.Saig_ObjLoToLi(p.pAig, obj)
		Ssw_SmlSimulateOneDyn_rec(p, li, frame-1, visited, visitID)
		Ssw_SmlNodeTransferNext(p, li, obj, frame-1)
		return
	}
	n := p.pAig.Node(obj)
	if n.Kind == aig.Output {
		Ssw_SmlSimulateOneDyn_rec(p, n.Fanins[0]&^1, frame, visited, visitID)
		Ssw_SmlNodeCopyFanin(p, obj, frame)
		return
	}
	if n.Kind != aig.And && n.Kind != aig.ExclusiveOr && n.Kind != aig.Buffer {
		panic("ssw: expected internal node")
	}
	Ssw_SmlSimulateOneDyn_rec(p, n.Fanins[0]&^1, frame, visited, visitID)
	if n.Kind != aig.Buffer {
		Ssw_SmlSimulateOneDyn_rec(p, n.Fanins[1]&^1, frame, visited, visitID)
	}
	Ssw_SmlNodeSimulate(p, obj, frame)
}
func Ssw_SmlSimulateOneFrame(p *Ssw_Sml_t) {
	start := time.Now()
	inputs, outputs := p.pAig.Inputs(), p.pAig.Outputs()
	regs := int(p.pAig.RegisterCount())
	for _, obj := range aig.Aig_ManDfsAll(p.pAig) {
		kind := p.pAig.Node(obj).Kind
		if kind == aig.And || kind == aig.ExclusiveOr || kind == aig.Buffer {
			Ssw_SmlNodeSimulate(p, obj, 0)
		}
	}
	for _, li := range outputs[len(outputs)-regs:] {
		Ssw_SmlNodeCopyFanin(p, li, 0)
	}
	for i := 0; i < regs; i++ {
		Ssw_SmlNodeTransferNext(p, outputs[len(outputs)-regs+i], inputs[len(inputs)-regs+i], 0)
	}
	p.timeSim += time.Since(start)
	p.nSimRounds++
}

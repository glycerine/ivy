package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

func sswConstrainedFrameRec(p *Ssw_Man_t, obj aig.Edge, frame int32) aig.Edge {
	if value := Ssw_ObjFrame(p, obj, frame); value != aig.InvalidEdge {
		return value
	}
	if obj.Complemented() || saig.Saig_ObjIsPi(p.pAig, obj) {
		panic("ssw: unmapped primary input")
	}
	var result aig.Edge
	if saig.Saig_ObjIsLo(p.pAig, obj) {
		if frame <= 0 {
			panic("ssw: unmapped initial latch")
		}
		li := saig.Saig_ObjLoToLi(p.pAig, obj)
		driver := p.pAig.Node(li).Fanins[0]
		result = sswConstrainedFrameRec(p, driver&^1, frame-1)
		if driver.Complemented() {
			result = result.Not()
		}
	} else {
		n := p.pAig.Node(obj)
		if n.Kind != aig.And && n.Kind != aig.ExclusiveOr && n.Kind != aig.Buffer {
			panic("ssw: expected internal node")
		}
		sswConstrainedFrameRec(p, n.Fanins[0]&^1, frame)
		if n.Kind != aig.Buffer {
			sswConstrainedFrameRec(p, n.Fanins[1]&^1, frame)
		}
		result = sswMakeNode(p, p.pFrames, obj, frame)
	}
	Ssw_ObjSetFrame(p, obj, frame, result)
	return result
}
func Ssw_ManSweepBmcConstr_rec(p *Ssw_Man_t, obj aig.Edge, frame int32) aig.Edge {
	return sswConstrainedFrameRec(p, obj, frame)
}
func Ssw_FramesWithClasses_rec(p *Ssw_Man_t, obj aig.Edge, frame int32) aig.Edge {
	return sswConstrainedFrameRec(p, obj, frame)
}

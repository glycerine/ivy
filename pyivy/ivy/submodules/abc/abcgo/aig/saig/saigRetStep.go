package saig

import "github.com/glycerine/abcgo/aig/aig"

// Saig_ManRetimeSteps moves up to nSteps gates across register boundaries.
// The existing register interface is retained, including registers left unused
// by a step; subsequent cleanup may remove them.
func Saig_ManRetimeSteps(p *aig.Graph, nSteps, forward, addBugs int32) int32 {
	if p == nil || nSteps <= 0 {
		return 0
	}
	aig.Aig_ManFanoutStart(p)
	defer aig.Aig_ManFanoutStop(p)
	if forward != 0 {
		Saig_ManMarkAutonomous(p)
	}
	var steps int32
	for ; steps < nSteps; steps++ {
		moved := false
		if forward != 0 {
			for _, obj := range p.LiveEdges() {
				if !p.ValidEdge(obj) || !aig.Aig_ObjIsNode(p, obj) {
					continue
				}
				var inject int32
				if addBugs != 0 && steps == 10 {
					inject = 1
				}
				replacement := Saig_ManRetimeNodeFwd(p, obj, inject)
				if replacement == aig.InvalidEdge {
					continue
				}
				aig.Aig_ObjReplace(p, obj, replacement, 0)
				moved = true
				break
			}
		} else {
			for i, count := int32(0), Saig_ManRegNum(p); i < count; i++ {
				lo := Saig_ManLo(p, i)
				if p.Node(lo).Refs == 0 {
					continue
				}
				replacement := Saig_ManRetimeNodeBwd(p, lo)
				if replacement == aig.InvalidEdge {
					continue
				}
				aig.Aig_ObjReplaceInputFanouts(p, lo, replacement)
				moved = true
				break
			}
		}
		if !moved {
			break
		}
	}
	aig.Aig_ManCleanup(p)
	return steps
}

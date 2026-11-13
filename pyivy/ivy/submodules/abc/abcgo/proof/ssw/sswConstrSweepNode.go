package ssw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

func Ssw_ManSweepNodeConstr(p *Ssw_Man_t, obj aig.Edge, frame, _ int32) int32 {
	repr := aig.Aig_ObjRepr(p.pAig, obj)
	if repr == aig.InvalidEdge {
		return 0
	}
	candidate, representative := Ssw_ObjFrame(p, obj, frame), Ssw_ObjFrame(p, repr, frame)
	if candidate == aig.InvalidEdge || representative == aig.InvalidEdge {
		panic("ssw: missing framed class node")
	}
	expectedSamePhase := p.pAig.Node(obj).Phase == p.pAig.Node(repr).Phase
	if expectedSamePhase != (aig.Aig_ObjPhaseReal(p.pFrames, candidate) == aig.Aig_ObjPhaseReal(p.pFrames, representative)) {
		panic("ssw: inconsistent framed node phases")
	}
	if candidate.NodeID() == representative.NodeID() {
		return 0
	}
	var result int32
	if candidate.NodeID() != 0 {
		result = Ssw_NodesAreEquiv(p, representative&^1, candidate&^1)
	} else {
		result = Ssw_NodesAreEquiv(p, candidate&^1, representative&^1)
	}
	if result == 1 {
		if !expectedSamePhase {
			representative = representative.Not()
		}
		Ssw_ObjSetFrame(p, obj, frame, representative)
		return 0
	}
	if result == -1 {
		Ssw_ClassesRemoveNode(p.ppClasses, obj)
		return 1
	}
	Ssw_SmlSavePatternAig(p, frame)
	Ssw_ManResimulateBit(p, obj, repr)
	if sswDeadlineExpired(p.pPars) {
		return 0
	}
	if aig.Aig_ObjRepr(p.pAig, obj) == repr {
		fmt.Println("Ssw_ManSweepNodeConstr(): Failed to refine representative.")
	}
	return 1
}

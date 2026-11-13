package cnf

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Cnf_CutAssignAreaFlow scores a cut by its leaves' current reference pressure.
func Cnf_CutAssignAreaFlow(manager *Cnf_Man_t, cut *aig.Aig_Cut_t, area []int32) int32 {
	if manager == nil || manager.pManAig == nil || cut == nil {
		return 0
	}
	var score int32
	for _, id := range aig.Aig_CutLeaves(cut) {
		edge := aig.Edge(uint32(id) << 1)
		node := manager.pManAig.Node(edge)
		score += int32(node.Refs)
		if node.Kind == aig.And || node.Kind == aig.ExclusiveOr {
			refs := int32(node.Refs)
			if refs < 1 {
				refs = 1
			}
			if int(id) < len(area) {
				score += area[id] / refs
			}
		}
	}
	return score
}

func Cnf_CutSuperAreaFlow(super *Vec_Ptr_t, area []int32, graph *aig.Graph) int32 {
	if super == nil || graph == nil {
		return 0
	}
	var score int32
	for _, item := range vec.Vec_PtrArray(super) {
		edge, ok := item.(aig.Edge)
		if !ok {
			continue
		}
		node := graph.Node(aig.Aig_Regular(edge))
		refs := int32(node.Refs)
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
			continue
		}
		if refs < 1 {
			refs = 1
		}
		id := edge.NodeID()
		if id < len(area) {
			score += area[id] / refs
		}
	}
	return score
}

func cutCNFCost(cut *aig.Aig_Cut_t) int {
	leaves := int(aig.Aig_CutLeaveNum(cut))
	truth := aig.Aig_CutTruth(cut)
	if len(truth) > 0 {
		return cutSopCost(truth, leaves)
	}
	return leaves + 1
}

// Cnf_DeriveMapping selects DAR cuts using the C area-flow cost and tie order.
func Cnf_DeriveMapping(manager *Cnf_Man_t) {
	if manager == nil || manager.pManAig == nil {
		return
	}
	if manager.pCuts != nil {
		aig.Aig_ManCutStop(manager.pCuts)
		manager.pCuts = nil
	}
	if err := cnfApplyDarMapping(manager); err != nil {
		panic(err)
	}
}

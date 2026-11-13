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

// Cnf_DeriveMapping computes native AIG cuts and picks a small support cut for
// each logic node. It stores copies in the manager, not in the graph's nodes.
func Cnf_DeriveMapping(manager *Cnf_Man_t) {
	if manager == nil || manager.pManAig == nil {
		return
	}
	if manager.pCuts != nil {
		aig.Aig_ManCutStop(manager.pCuts)
	}
	manager.pCuts = aig.Aig_ComputeCuts(manager.pManAig, 16, 4, 1, 0)
	nodes := manager.pManAig.Nodes()
	manager.selectedCuts = make([]*Cnf_Cut_t, len(nodes))
	for id, node := range nodes {
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer && node.Kind != aig.Output {
			continue
		}
		var best *aig.Aig_Cut_t
		for cut := aig.Aig_ObjCuts(manager.pCuts, aig.Edge(id)<<1); cut != nil; cut = aig.Aig_CutNext(cut) {
			leaves := aig.Aig_CutLeaves(cut)
			if len(leaves) == 0 || len(leaves) == 1 && leaves[0] == int32(id) || len(aig.Aig_CutTruth(cut)) == 0 {
				continue
			}
			if best == nil || cutCNFCost(cut) < cutCNFCost(best) {
				best = cut
			}
		}
		if best == nil {
			continue
		}
		leaves := append([]int32(nil), aig.Aig_CutLeaves(best)...)
		truth := append([]uint32(nil), aig.Aig_CutTruth(best)...)
		cnfCut := Cnf_CutFromTruth(leaves, truth)
		cnfCut.Cost = byte(cutCNFCost(best))
		manager.selectedCuts[id] = cnfCut
	}
}

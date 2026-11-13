package cnf

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

// cnfApplyDarMapping implements Cnf_DeriveMapping's area-flow score and
// reference-count tie break over the source-order DAR cut slots.
func cnfApplyDarMapping(manager *Cnf_Man_t) error {
	if manager == nil || manager.pManAig == nil {
		return fmt.Errorf("cnf: missing mapping graph")
	}
	g := manager.pManAig
	aig.Aig_ManCleanup(g)
	cuts, err := dar.Dar_ManComputeCutSlots(g, 10, false)
	if err != nil {
		return err
	}
	nodes := g.Nodes()
	manager.selectedCuts = make([]*Cnf_Cut_t, len(nodes))
	area := make([]int, len(nodes))
	sizes, _ := cnfReadMsops()
	for _, edge := range g.LiveEdges() {
		id := edge.NodeID()
		node := nodes[id]
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
			continue
		}
		bestIndex := -1
		bestScore, bestRefs := 0, 0
		for index, slot := range cuts[edge] {
			if index == 0 || !slot.Used {
				continue
			}
			c := slot.Cut
			score := 10 * (int(sizes[c.Truth]) + int(sizes[^c.Truth]))
			refs := 0
			for _, leaf := range c.Leaves {
				n := nodes[leaf.NodeID()]
				refs += n.Refs
				if n.Kind == aig.And || n.Kind == aig.ExclusiveOr {
					if n.Refs <= 0 {
						return fmt.Errorf("cnf: unreferenced mapped cut leaf %d", leaf.NodeID())
					}
					score += area[leaf.NodeID()] / n.Refs
				}
			}
			if bestIndex < 0 || score < bestScore || score == bestScore && refs > bestRefs {
				bestIndex, bestScore, bestRefs = index, score, refs
			}
		}
		if bestIndex < 0 {
			return fmt.Errorf("cnf: node %d has no nontrivial DAR cut", id)
		}
		chosen := cuts[edge][bestIndex].Cut
		area[id] = bestScore
		leaves := make([]int32, len(chosen.Leaves))
		for i, e := range chosen.Leaves {
			leaves[i] = int32(e.NodeID())
		}
		manager.selectedCuts[id] = Cnf_CutFromTruth(leaves, []uint32{uint32(chosen.Truth) | uint32(chosen.Truth)<<16})
		manager.selectedCuts[id].Cost = byte(int(sizes[chosen.Truth]) + int(sizes[^chosen.Truth]))
	}
	return nil
}

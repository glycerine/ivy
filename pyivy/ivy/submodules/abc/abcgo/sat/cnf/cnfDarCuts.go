package cnf

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

// cnfApplyDarMapping implements Cnf_DeriveMapping's area-flow score and
// reference-count tie break over the source-order DAR cut slots.
func cnfApplyDarMapping(manager *Cnf_Man_t, skipTruthMin bool) error {
	if manager == nil || manager.pManAig == nil {
		return fmt.Errorf("cnf: missing mapping graph")
	}
	g := manager.pManAig
	cuts, err := dar.NewWholeGraphCutEnumerator(g, 10, skipTruthMin)
	if err != nil {
		return err
	}
	nodes := g.Nodes()
	manager.selectedCuts = make([]*Cnf_Cut_t, len(nodes))
	area := make([]int, len(nodes))
	sizes, _ := cnfReadMsops()
	for _, edge := range g.TopologicalObjects() {
		id := edge.NodeID()
		node := nodes[id]
		if node.Kind == aig.Buffer {
			driver := node.Fanins[0]
			truth := uint32(0xaaaaaaaa)
			if driver.Complemented() {
				truth = ^truth
			}
			manager.selectedCuts[id] = Cnf_CutFromTruth([]int32{int32(driver.NodeID())}, []uint32{truth})
			manager.selectedCuts[id].Cost = 2
			continue
		}
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
			continue
		}
		var chosen dar.CutCandidate
		found := false
		bestScore, bestRefs := 0, 0
		for index, c := range cuts.Candidates(edge) {
			if index == 0 {
				continue
			}
			score := 10 * (int(sizes[c.Truth]) + int(sizes[^c.Truth]))
			refs := 0
			for _, leaf := range c.Leaves[:c.NumLeaves] {
				n := nodes[leaf.NodeID()]
				refs += n.Refs
				if n.Kind == aig.And || n.Kind == aig.ExclusiveOr {
					if n.Refs <= 0 {
						return fmt.Errorf("cnf: unreferenced mapped cut leaf %d", leaf.NodeID())
					}
					score += area[leaf.NodeID()] / n.Refs
				}
			}
			// Dar_Cut_t.Value stores the tie break in an 11-bit field.
			refs &= 0x7ff
			if !found || score < bestScore || score == bestScore && refs > bestRefs {
				chosen, bestScore, bestRefs = c, score, refs
				found = true
			}
		}
		if !found {
			return fmt.Errorf("cnf: node %d has no nontrivial DAR cut", id)
		}
		area[id] = bestScore
		cut := Cnf_CutAlloc(manager, int32(chosen.NumLeaves))
		for i, e := range chosen.Leaves[:chosen.NumLeaves] {
			cut.pFanins[i] = int32(e.NodeID())
		}
		cut.truth[0] = uint32(chosen.Truth) | uint32(chosen.Truth)<<16
		cut.Cost = byte(int(sizes[chosen.Truth]) + int(sizes[^chosen.Truth]))
		manager.selectedCuts[id] = cut
	}
	return nil
}

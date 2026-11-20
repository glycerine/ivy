package cnf

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

// Four-leaf DAR cuts always need one truth word after transfer. Reserve
// one record per gate while keeping each mutable CNF cut independent.
type smallMappedCut struct {
	cut    Cnf_Cut_t
	leaves [4]int32
	truth  [1]uint32
}

func (c *smallMappedCut) init(leaves []int32, truth uint32, cost byte) *Cnf_Cut_t {
	n := len(leaves)
	copy(c.leaves[:], leaves)
	c.truth[0] = truth
	c.cut = Cnf_Cut_t{nFanins: byte(n), nWords: 1, Cost: cost, pFanins: c.leaves[:n:n], truth: c.truth[:]}
	return &c.cut
}

// cnfApplyDarMapping implements Cnf_DeriveMapping's area-flow score and
// reference-count tie break over the source-order DAR cut slots.
func cnfApplyDarMapping(manager *Cnf_Man_t, skipTruthMin bool) error {
	if manager == nil || manager.pManAig == nil {
		return fmt.Errorf("cnf: missing mapping graph")
	}
	g := manager.pManAig
	manager.selectedCuts, manager.darBest = nil, nil
	cuts, err := dar.NewWholeGraphCutEnumerator(g, 10, skipTruthMin)
	if err != nil {
		return err
	}
	count := aigObjectCount(g)
	selected := make([]*Cnf_Cut_t, count)
	best := make([]Dar_Cut_t, count)
	refsByID := make([]int32, count)
	order := g.TopologicalObjects()
	gates := 0
	for _, edge := range order {
		node := g.Node(edge)
		refsByID[edge.NodeID()] = int32(node.Refs)
		if node.Kind == aig.And || node.Kind == aig.ExclusiveOr || node.Kind == aig.Buffer {
			gates++
		}
	}
	storage := make([]smallMappedCut, gates)
	sizes, _ := cnfReadMsops()
	for _, edge := range order {
		id := edge.NodeID()
		node := g.Node(edge)
		if node.Kind == aig.Buffer {
			driver := node.Fanins[0]
			truth := uint32(0xaaaaaaaa)
			if driver.Complemented() {
				truth = ^truth
			}
			selected[id] = storage[0].init([]int32{int32(driver.NodeID())}, truth, 2)
			storage = storage[1:]
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
				leafID := leaf.NodeID()
				leafRefs := int(refsByID[leafID])
				refs += leafRefs
				if best[leafID].fBest {
					if leafRefs <= 0 {
						return fmt.Errorf("cnf: unreferenced mapped cut leaf %d", leafID)
					}
					score += int(best[leafID].uSign) / leafRefs
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
		d := &best[id]
		*d = Dar_Cut_t{uTruth: chosen.Truth, uSign: uint32(bestScore), Value: uint16(bestRefs), nLeaves: chosen.NumLeaves, fBest: true}
		for i, e := range chosen.Leaves[:chosen.NumLeaves] {
			d.pLeaves[i] = int32(e.NodeID())
		}
		truth := uint32(d.uTruth) * 0x10001
		cost := byte(int(sizes[d.uTruth]) + int(sizes[^d.uTruth]))
		selected[id] = storage[0].init(d.pLeaves[:d.nLeaves], truth, cost)
		storage = storage[1:]
	}
	manager.selectedCuts, manager.darBest = selected, best
	return nil
}

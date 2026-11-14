package aig

import "fmt"

// Aig_ManReorderCisCos reorders the graph's interface lists without changing
// node IDs or logic. inputOrder and outputOrder must be permutations of the
// existing interfaces; registers remain the final regCount CIs/COs.
// Stored CioIDs and true-PI/PO counts are preserved for source-style retiming.
// Call Aig_ManSetCioIds explicitly when current interface indices are needed.
func Aig_ManReorderCisCos(g *Graph, inputOrder, outputOrder []Edge, regCount int32) error {
	if g == nil {
		return fmt.Errorf("aig: cannot reorder a nil graph")
	}
	if len(inputOrder) != len(g.inputs) || len(outputOrder) != len(g.outputs) || regCount < 0 || int(regCount) > len(inputOrder) || int(regCount) > len(outputOrder) {
		return fmt.Errorf("aig: invalid reordered interface dimensions")
	}
	if !sameEdgePermutation(g.inputs, inputOrder) || !sameEdgePermutation(g.outputs, outputOrder) {
		return fmt.Errorf("aig: reordered interfaces must be permutations")
	}
	oldRegCount := int(g.nRegs)
	oldPiCount := len(g.inputs) - oldRegCount
	newPiCount := len(inputOrder) - int(regCount)
	regRemap := make([]int32, oldRegCount)
	for i := range regRemap {
		regRemap[i] = -1
	}
	for newIndex, edge := range inputOrder {
		oldIndex := edgeIndex(g.inputs, edge)
		if oldIndex >= oldPiCount && newIndex >= newPiCount {
			regRemap[oldIndex-oldPiCount] = int32(newIndex - newPiCount)
		}
	}
	if len(g.InputNumbers) == len(g.inputs) {
		values := append([]int32(nil), g.InputNumbers...)
		for i, edge := range inputOrder {
			g.InputNumbers[i] = values[edgeIndex(g.inputs, edge)]
		}
	}
	remapRegisterSlice := func(values []int32) []int32 {
		if len(values) != oldRegCount {
			return nil
		}
		result := make([]int32, regCount)
		for oldID, newID := range regRemap {
			if newID >= 0 {
				result[newID] = values[oldID]
			}
		}
		return result
	}
	g.FlopNumbers = remapRegisterSlice(g.FlopNumbers)
	g.FlopReprs = remapRegisterSlice(g.FlopReprs)
	if len(g.ClockDomains) != 0 {
		domains := make([][]int32, len(g.ClockDomains))
		for domain, registers := range g.ClockDomains {
			for _, oldID := range registers {
				if oldID < 0 || int(oldID) >= len(regRemap) {
					continue
				}
				if newID := regRemap[oldID]; newID >= 0 {
					domains[domain] = append(domains[domain], newID)
				}
			}
		}
		g.ClockDomains = domains
	}
	g.inputs = append(g.inputs[:0], inputOrder...)
	g.outputs = append(g.outputs[:0], outputOrder...)
	g.nRegs = regCount
	return nil
}

func sameEdgePermutation(original, reordered []Edge) bool {
	if len(original) != len(reordered) {
		return false
	}
	seen := make(map[Edge]bool, len(original))
	for _, edge := range original {
		seen[edge] = true
	}
	for _, edge := range reordered {
		if !seen[edge] {
			return false
		}
		delete(seen, edge)
	}
	return len(seen) == 0
}

func edgeIndex(edges []Edge, target Edge) int {
	for i, edge := range edges {
		if edge == target {
			return i
		}
	}
	return -1
}

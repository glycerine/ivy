// Package mf ports the cut mapper and mapped-CNF generator from giaMf.c.
package mf

import "math/bits"

const epsilon float32 = 0.005
const maxLeaves = 10

// Cut stores a source MF cut without the packed header or paged arena.
// Function is an interned truth-table literal, including its complement bit.
type Cut struct {
	Leaves   []int32
	Function int32
	Delay    int32
	Flow     float32
}

func signature(leaves []int32) uint64 {
	var sign uint64
	for _, leaf := range leaves {
		sign |= uint64(1) << uint(leaf&63)
	}
	return sign
}
func contains(base, sub []int32) bool {
	i := 0
	for _, leaf := range sub {
		for i < len(base) && base[i] < leaf {
			i++
		}
		if i == len(base) || base[i] != leaf {
			return false
		}
		i++
	}
	return true
}
func mergeLeaves(a, b []int32, limit int) ([]int32, bool) {
	if len(a)+len(b) > limit && bits.OnesCount64(signature(a)|signature(b)) > limit {
		return nil, false
	}
	merged := make([]int32, 0, min(len(a)+len(b), limit))
	for i, j := 0, 0; i < len(a) || j < len(b); {
		var leaf int32
		switch {
		case j == len(b) || i < len(a) && a[i] < b[j]:
			leaf = a[i]
			i++
		case i == len(a) || b[j] < a[i]:
			leaf = b[j]
			j++
		default:
			leaf = a[i]
			i++
			j++
		}
		if len(merged) == limit {
			return nil, false
		}
		merged = append(merged, leaf)
	}
	return merged, true
}
func compareArea(a, b Cut) int {
	switch {
	case a.Flow < b.Flow-epsilon:
		return -1
	case a.Flow > b.Flow+epsilon:
		return 1
	case a.Delay < b.Delay:
		return -1
	case a.Delay > b.Delay:
		return 1
	case len(a.Leaves) < len(b.Leaves):
		return -1
	case len(a.Leaves) > len(b.Leaves):
		return 1
	}
	return 0
}

// addCut reproduces the MF insertion order: equal-ranked new cuts precede old
// cuts, strict supersets are removed, and one cut slot is reserved for the unit.
func addCut(cuts []Cut, cut Cut, limit int) []Cut {
	kept := cuts[:0]
	for _, old := range cuts {
		if len(cut.Leaves) < len(old.Leaves) && contains(old.Leaves, cut.Leaves) {
			continue
		}
		kept = append(kept, old)
	}
	cuts = append(kept, cut)
	for i := len(cuts) - 1; i > 0; i-- {
		if compareArea(cuts[i-1], cuts[i]) < 0 {
			break
		}
		cuts[i-1], cuts[i] = cuts[i], cuts[i-1]
	}
	if len(cuts) > limit-1 {
		cuts = cuts[:limit-1]
	}
	return cuts
}
func isContained(cuts []Cut, leaves []int32) bool {
	for _, old := range cuts {
		if len(old.Leaves) <= len(leaves) && contains(leaves, old.Leaves) {
			return true
		}
	}
	return false
}

// expandTruth aligns a source cut's variables with a merged cut. MF stores all
// truth words at the configured LUT width; smaller supports repeat their table.
func expandTruth(table []uint64, old, merged []int32, width int, complement bool) []uint64 {
	result := make([]uint64, 1<<max(0, width-6))
	positions := make([]int, len(old))
	for i, leaf := range old {
		for positions[i] < len(merged) && merged[positions[i]] != leaf {
			positions[i]++
		}
		if positions[i] == len(merged) {
			panic("mf: merged cut omitted a leaf")
		}
	}
	for assignment := 0; assignment < len(result)*64; assignment++ {
		oldAssignment := 0
		for i, position := range positions {
			oldAssignment |= ((assignment >> uint(position)) & 1) << uint(i)
		}
		value := table[oldAssignment>>6] >> uint(oldAssignment&63) & 1
		if complement {
			value ^= 1
		}
		result[assignment>>6] |= value << uint(assignment&63)
	}
	return result
}

// minimizeSupport removes truth-independent variables, preserving sorted leaves
// and the source convention of repeating the reduced table to the LUT width.
func minimizeSupport(table []uint64, leaves []int32, width int) ([]uint64, []int32) {
	supported := make([]int, 0, len(leaves))
	for variable := range leaves {
		dependent := false
		for assignment := 0; assignment < 1<<len(leaves); assignment++ {
			if assignment&(1<<uint(variable)) != 0 {
				continue
			}
			other := assignment | 1<<uint(variable)
			if (table[assignment>>6]>>uint(assignment&63)^table[other>>6]>>uint(other&63))&1 != 0 {
				dependent = true
				break
			}
		}
		if dependent {
			supported = append(supported, variable)
		}
	}
	if len(supported) == len(leaves) {
		return table, leaves
	}
	result := make([]uint64, 1<<max(0, width-6))
	reduced := make([]int32, len(supported))
	for i, variable := range supported {
		reduced[i] = leaves[variable]
	}
	for assignment := 0; assignment < len(result)*64; assignment++ {
		original := 0
		for i, variable := range supported {
			original |= ((assignment >> uint(i)) & 1) << uint(variable)
		}
		result[assignment>>6] |= (table[original>>6] >> uint(original&63) & 1) << uint(assignment&63)
	}
	return result, reduced
}

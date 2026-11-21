// Package mf ports the cut mapper and mapped-CNF generator from giaMf.c.
package mf

import (
	"math/bits"

	"github.com/glycerine/abcgo/misc/util"
)

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
func mergeLeaves(merged, a, b []int32, limit int) ([]int32, bool) {
	if len(a)+len(b) > limit && bits.OnesCount64(signature(a)|signature(b)) > limit {
		return nil, false
	}
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
	kept := 0
	for i, old := range cuts {
		if len(cut.Leaves) < len(old.Leaves) && contains(old.Leaves, cut.Leaves) {
			continue
		}
		if kept != i {
			cuts[kept] = old
		}
		kept++
	}
	pos := kept
	for pos > 0 && compareArea(cuts[pos-1], cut) >= 0 {
		pos--
	}
	if pos >= limit-1 {
		return cuts[:limit-1]
	}
	if kept < limit-1 {
		cuts = append(cuts[:kept], Cut{})
	} else {
		cuts = cuts[:kept]
	}
	copy(cuts[pos+1:], cuts[pos:])
	// A retained cut owns its leaves, including cuts seeded from a sibling.
	cut.Leaves = append([]int32(nil), cut.Leaves...)
	cuts[pos] = cut
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
func expandTruth(result, table []uint64, old, merged []int32, width int, complement bool) {
	copy(result, table)
	if complement {
		for i := range result {
			result[i] = ^result[i]
		}
	}
	util.Abc_TtExpand(result, int32(width), old, int32(len(old)), merged, int32(len(merged)))
}

// The candidate owns both arrays. Compact its leaves and truth table in place.
func minimizeSupport(table []uint64, leaves []int32, width int) ([]uint64, []int32) {
	count := util.Abc_TtMinBase(table, leaves, int32(len(leaves)), int32(width))
	return table, leaves[:count]
}

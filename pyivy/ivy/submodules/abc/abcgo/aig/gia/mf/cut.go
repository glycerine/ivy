// Package mf ports the cut mapper and mapped-CNF generator from giaMf.c.
package mf

import (
	"math/bits"
	"slices"

	"github.com/glycerine/abcgo/misc/util"
)

const epsilon float32 = 0.005
const maxLeaves = 10
const maxCuts = 16

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

// The candidate borrows the caller's array until its selected cut is saved.
func mergeLeaves(buffer *[maxLeaves]int32, a, b []int32, limit int) ([]int32, bool) {
	merged := buffer[:limit]
	n := 0
	for len(a) != 0 && len(b) != 0 {
		if n == limit {
			return nil, false
		}
		left, right := a[0], b[0]
		switch {
		case left < right:
			merged[n] = left
			a = a[1:]
		case right < left:
			merged[n] = right
			b = b[1:]
		default:
			merged[n] = left
			a, b = a[1:], b[1:]
		}
		n++
	}
	if len(b) > 0 {
		a = b
	}
	if n+len(a) > limit {
		return nil, false
	}
	n += copy(merged[n:], a)
	return merged[:n:n], true
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

type cutSlot struct {
	cut    Cut
	sign   uint64
	leaves [maxLeaves]int32
}

// Each node reuses these candidate records. Sorting moves slot indices, and
// the first unused index names the record available for the next insertion.
type cutSet struct {
	index        leafIndex
	unitLeaves   [4][1]int32
	slots        [maxCuts]cutSlot
	order        [maxCuts]uint8
	count, limit int
}

// A gate's candidate cuts often use at most 64 distinct objects between them.
// Assign bits in leaf-ID order so that a union can be decoded in sorted order.
type leafIndex struct {
	position []uint8 // One-based bit position, zero for objects outside this gate.
	ids      [64]int32
	count    int
	exact    bool
}

func (x *leafIndex) reset(fanins [4][]Cut) {
	for _, id := range x.ids[:x.count] {
		x.position[id] = 0
	}
	x.count, x.exact = 0, false
	for _, cuts := range fanins {
		for _, cut := range cuts {
			for _, id := range cut.Leaves {
				if x.position[id] != 0 {
					continue
				}
				if x.count == len(x.ids) {
					return
				}
				x.position[id] = 1
				x.ids[x.count] = id
				x.count++
			}
		}
	}
	slices.Sort(x.ids[:x.count])
	for i, id := range x.ids[:x.count] {
		x.position[id] = uint8(i + 1)
	}
	x.exact = true
}

func (x *leafIndex) mask(leaves []int32) uint64 {
	if !x.exact {
		return signature(leaves)
	}
	var mask uint64
	for _, id := range leaves {
		mask |= uint64(1) << (x.position[id] - 1)
	}
	return mask
}

func (s *cutSet) candidate(buffer *[maxLeaves]int32, a, b []int32, sign uint64, limit int) ([]int32, bool) {
	if bits.OnesCount64(sign) > limit {
		return nil, false
	}
	if s.index.exact {
		for _, id := range s.order[:s.count] {
			if s.slots[id].sign&^sign == 0 {
				return nil, false
			}
		}
		n := 0
		for mask := sign; mask != 0; mask &= mask - 1 {
			buffer[n] = s.index.ids[bits.TrailingZeros64(mask)]
			n++
		}
		return buffer[:n:n], true
	}
	leaves, ok := mergeLeaves(buffer, a, b, limit)
	return leaves, ok && !s.contains(leaves, sign)
}

func newCutSet(cutLimit int) *cutSet {
	s := &cutSet{limit: cutLimit - 1} // Reserve one entry for the unit cut.
	for i := range s.order {
		s.order[i] = uint8(i)
	}
	return s
}

// Equal-ranked new cuts precede old cuts, and strict supersets are removed.
func (s *cutSet) add(cut Cut, sign uint64) {
	kept := 0
	for i, id := range s.order[:s.count] {
		old := &s.slots[id]
		if len(cut.Leaves) < len(old.cut.Leaves) && sign&^old.sign == 0 && (s.index.exact || contains(old.cut.Leaves, cut.Leaves)) {
			continue
		}
		if kept != i {
			s.order[kept], s.order[i] = id, s.order[kept]
		}
		kept++
	}
	pos := kept
	for pos > 0 && compareArea(s.slots[s.order[pos-1]].cut, cut) >= 0 {
		pos--
	}
	if pos == s.limit {
		return
	}
	id := s.order[kept]
	slot := &s.slots[id]
	leaves := slot.leaves[:len(cut.Leaves)]
	copy(leaves, cut.Leaves)
	slot.cut = Cut{Leaves: leaves, Function: cut.Function, Delay: cut.Delay, Flow: cut.Flow}
	slot.sign = sign
	copy(s.order[pos+1:kept+1], s.order[pos:kept])
	s.order[pos] = id
	s.count = min(kept+1, s.limit)
}

func (s *cutSet) contains(leaves []int32, sign uint64) bool {
	for _, id := range s.order[:s.count] {
		old := &s.slots[id]
		if len(old.cut.Leaves) <= len(leaves) && old.sign&^sign == 0 && contains(leaves, old.cut.Leaves) {
			return true
		}
	}
	return false
}

func (s *cutSet) save() []Cut {
	order := s.order[:s.count]
	total := 0
	for _, id := range order {
		total += len(s.slots[id].cut.Leaves)
	}
	leaves := make([]int32, total)
	cuts := make([]Cut, len(order))
	for i, id := range order {
		cut := s.slots[id].cut
		n := copy(leaves, cut.Leaves)
		if n == 0 {
			cut.Leaves = nil
		} else {
			cut.Leaves = leaves[:n:n]
			leaves = leaves[n:]
		}
		cuts[i] = cut
	}
	return cuts
}

// leafPositions records a sorted subset in a merged cut of at most eight leaves.
func leafPositions(subset, merged []int32) uint8 {
	var positions uint8
	k := 0
	for i, leaf := range merged {
		if k == len(subset) {
			break
		}
		if subset[k] == leaf {
			positions |= 1 << uint(i)
			k++
		}
	}
	if k != len(subset) {
		panic("mf: merged cut is missing a leaf")
	}
	return positions
}

// expandTruth aligns a source cut's variables with a merged cut. MF stores all
// truth words at the configured LUT width; smaller supports repeat their table.
func expandTruth(result, table []uint64, count int, positions uint8, width int, complement bool) {
	copy(result, table)
	if complement {
		for i := range result {
			result[i] = ^result[i]
		}
	}
	for k := count - 1; k >= 0; k-- {
		last := bits.Len8(positions) - 1
		if k == last {
			break // All remaining positions are already aligned.
		}
		util.Abc_TtSwapVars(result, int32(width), int32(k), int32(last))
		positions &^= 1 << uint(last)
	}
}

// The candidate owns both arrays. Compact its leaves and truth table in place.
func minimizeSupport(table []uint64, leaves []int32, width int) ([]uint64, []int32) {
	count := util.Abc_TtMinBase(table, leaves, int32(len(leaves)), int32(width))
	return table, leaves[:count]
}

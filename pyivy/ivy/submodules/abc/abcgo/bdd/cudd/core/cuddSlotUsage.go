package core

import (
	"math"
	"math/bits"
)

func (table *uniqueTable) levelCount(index int) int {
	if index >= len(table.levels) {
		return 0
	}
	return table.levels[index].count
}

// This ABC build of CUDD hashes allocation IDs shifted left four bits,
// then shifted once more to make room for the complement bit.
func uniqueChildToken(n *DdNode) uint32 {
	if n.complemented {
		return uint32(n.inverse.ID+2)<<4 | 1
	}
	return uint32(n.ID+2) << 4
}

func slotHash(f, g uint32, shift int) uint32 {
	return ((f*12582917 + g) * 4256249) >> shift
}

type occupiedBins struct {
	local  [4]uint64
	words  []uint64
	sparse map[uint32]struct{}
	shift  int
}

func (bins *occupiedBins) init(width uint32, keys int) {
	bins.shift = bits.LeadingZeros32(width - 1)
	words := (uint64(width) + 63) / 64
	if words <= uint64(len(bins.local)) {
		return
	}
	if words <= 4*uint64(keys) {
		bins.words = make([]uint64, words)
	} else {
		bins.sparse = make(map[uint32]struct{}, keys)
	}
}

func (bins *occupiedBins) mark(f, g uint32) {
	b := slotHash(f, g, bins.shift)
	switch {
	case bins.words != nil:
		bins.words[b/64] |= uint64(1) << (b % 64)
	case bins.sparse != nil:
		bins.sparse[b] = struct{}{}
	default:
		bins.local[b/64] |= uint64(1) << (b % 64)
	}
}

func (bins *occupiedBins) count() int {
	if bins.sparse != nil {
		return len(bins.sparse)
	}
	words := bins.words
	if words == nil {
		words = bins.local[:]
	}
	count := 0
	for _, w := range words {
		count += bits.OnesCount64(w)
	}
	return count
}

func (table *uniqueTable) markLevel(index int, bins *occupiedBins) {
	if index >= len(table.levels) {
		return
	}
	l := &table.levels[index]
	if l.first.node != nil {
		bins.mark(uniqueChildToken(l.first.t), uniqueChildToken(l.first.e))
	}
	for _, r := range l.slots {
		if r.node != nil {
			bins.mark(uniqueChildToken(r.t), uniqueChildToken(r.e))
		}
	}
}

// Each BDD/ADD level shares its native hash buckets. ZDD levels and constants
// have their own buckets. Include resident dead keys until they are collected.
func (dd *DdManager) slotUsage(expected bool) float64 {
	total := Cudd_ReadSlots(dd)
	if total == 0 {
		return 0
	}
	var used float64
	for family, widths := range [2][]uint32{dd.bddTableSlots, dd.zddTableSlots} {
		perm, table := dd.InvPerm, &dd.unique
		if family == 1 {
			perm, table = dd.InvPermZ, &dd.zddUnique
		}
		for level, width := range widths {
			index := level
			if level < len(perm) {
				index = perm[level]
			}
			keys := table.levelCount(index)
			if family == 0 {
				keys += dd.addUnique.levelCount(index)
			}
			if keys == 0 {
				continue
			}
			if expected {
				used += float64(width) * (1 - math.Exp(-float64(keys)/float64(width)))
			} else if keys == 1 {
				used++
			} else {
				var bins occupiedBins
				bins.init(width, keys)
				table.markLevel(index, &bins)
				if family == 0 {
					dd.addUnique.markLevel(index, &bins)
				}
				used += float64(bins.count())
			}
		}
	}
	width := dd.constantTableWidth()
	if expected {
		used += float64(width) * (1 - math.Exp(-float64(len(dd.constants))/float64(width)))
	} else {
		var bins occupiedBins
		bins.init(width, len(dd.constants))
		for key := range dd.constants {
			bins.mark(uint32(key), uint32(key>>32))
		}
		used += float64(bins.count())
	}
	return used / float64(total)
}

func Cudd_ReadUsedSlots(dd *DdManager) float64     { return dd.slotUsage(false) }
func Cudd_ExpectedUsedSlots(dd *DdManager) float64 { return dd.slotUsage(true) }

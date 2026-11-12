package sfm

import "slices"

// sfmCombineDecomposition forms selector ? on : off, preserving sorted support.
// A nil result indicates that the combined support exceeds the variable limit.
func sfmCombineDecomposition(off, on []uint64, offSupport, onSupport []int32, selector int32, limit int) ([]uint64, []int32) {
	if slices.Equal(offSupport, onSupport) && slices.Equal(off, on) {
		words := 1
		if limit > 6 {
			words = 1 << uint(limit-6)
		}
		truth := make([]uint64, words)
		for i := range truth {
			truth[i] = off[i%len(off)]
		}
		return truth, append([]int32(nil), offSupport...)
	}
	support := append(append([]int32(nil), offSupport...), onSupport...)
	support = append(support, selector)
	slices.Sort(support)
	support = slices.Compact(support)
	if len(support) > limit {
		return nil, nil
	}
	positions := func(source []int32) []int {
		result := make([]int, len(source))
		for i, v := range source {
			index, ok := slices.BinarySearch(support, v)
			if !ok {
				panic("sfm: missing support")
			}
			result[i] = index
		}
		return result
	}
	a, b := positions(offSupport), positions(onSupport)
	selectPos, _ := slices.BinarySearch(support, selector)
	words := 1
	if limit > 6 {
		words = 1 << uint(limit-6)
	}
	truth := make([]uint64, words)
	for assignment := 0; assignment < words*64; assignment++ {
		source, positions := off, a
		if assignment&(1<<uint(selectPos)) != 0 {
			source, positions = on, b
		}
		index := 0
		for i, pos := range positions {
			if assignment&(1<<uint(pos)) != 0 {
				index |= 1 << uint(i)
			}
		}
		if source[index>>6]&(uint64(1)<<uint(index&63)) != 0 {
			truth[assignment>>6] |= uint64(1) << uint(assignment&63)
		}
	}
	return truth, support
}

// Package canon ports the active Abc_TtCanonicize algorithm from dauCanon.c.
// It is a leaf package so both DAU and GIA can use it without an import cycle.
package canon

import "math/bits"

var variables = [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}

func flip(table []uint64, v int) {
	if v < 6 {
		shift := uint(1) << uint(v)
		mask := variables[v]
		for i, x := range table {
			table[i] = (x&^mask)<<shift | (x&mask)>>shift
		}
		return
	}
	step := 1 << uint(v-6)
	for start := 0; start < len(table); start += 2 * step {
		for i := 0; i < step; i++ {
			table[start+i], table[start+step+i] = table[start+step+i], table[start+i]
		}
	}
}
func adjacent(table []uint64, v int) {
	if v < 5 {
		shift := uint(1) << uint(v)
		lo, hi := variables[v]&^variables[v+1], variables[v+1]&^variables[v]
		for i, x := range table {
			table[i] = x&^(lo|hi) | (x&lo)<<shift | (x&hi)>>shift
		}
		return
	}
	if v == 5 {
		for i := 0; i < len(table); i += 2 {
			a, b := table[i], table[i+1]
			table[i] = a&0xffffffff | b<<32
			table[i+1] = a>>32 | b&0xffffffff00000000
		}
		return
	}
	step := 1 << uint(v-6)
	for start := 0; start < len(table); start += 4 * step {
		for i := 0; i < step; i++ {
			table[start+step+i], table[start+2*step+i] = table[start+2*step+i], table[start+step+i]
		}
	}
}
func swap(table []uint64, a, b int) {
	if a > b {
		a, b = b, a
	}
	for i := a; i < b; i++ {
		adjacent(table, i)
	}
	for i := b - 2; i >= a; i-- {
		adjacent(table, i)
	}
}
func less(a, b []uint64) bool {
	for i := len(a) - 1; i >= 0; i-- {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}
func swapPhase(phase uint32, a, b int) uint32 {
	if phase>>uint(a)&1 != phase>>uint(b)&1 {
		phase ^= 1<<uint(a) | 1<<uint(b)
	}
	return phase
}
func negativeCount(table []uint64, v int) int {
	count := 0
	for i, x := range table {
		if v < 6 {
			count += bits.OnesCount64(x &^ variables[v])
		} else if i&(1<<uint(v-6)) == 0 {
			count += bits.OnesCount64(x)
		}
	}
	return count
}

// refine considers exactly the source's configuration order; strict comparison
// preserves its tie-breaking and the resulting phase/permutation metadata.
func refine(table, work, best []uint64, v int, swapOnly bool) int {
	copy(work, table)
	if swapOnly {
		adjacent(work, v)
		if less(work, table) {
			copy(table, work)
			return 4
		}
		return 0
	}
	copy(best, table)
	config := 0
	operations := [7]int{1, 2, 1, 4, 2, 1, 2}
	configs := [7]int{1, 3, 2, 6, 7, 5, 4}
	for i, op := range operations {
		switch op {
		case 1:
			flip(work, v)
		case 2:
			flip(work, v+1)
		case 4:
			adjacent(work, v)
		}
		if less(work, best) {
			copy(best, work)
			config = configs[i]
		}
	}
	if config != 0 {
		copy(table, best)
	}
	return config
}

// Canonicalize modifies a normalized little-endian truth table in place.
// It returns ABC's phase mask and writes new-position -> old-position into perm.
// This is the original bounded semi-canonicalization/refinement algorithm,
// including selection-sort ties and the maximum of five bidirectional passes.
func Canonicalize(truth []uint64, nVars int32, perm []byte) uint32 {
	if nVars < 0 || nVars > 16 || len(perm) < int(nVars) {
		panic("canon: invalid truth dimensions")
	}
	words := 1
	if nVars > 6 {
		words = 1 << uint(nVars-6)
	}
	if len(truth) < words {
		panic("canon: short truth table")
	}
	table := truth[:words]
	n := int(nVars)
	for i := 0; i < n; i++ {
		perm[i] = byte(i)
	}
	counts := make([]int, n+1)
	ones := 0
	for _, x := range table {
		ones += bits.OnesCount64(x)
	}
	var phase uint32
	if ones > words*32 {
		for i := range table {
			table[i] = ^table[i]
		}
		ones = words*64 - ones
		phase |= 1 << uint(n)
	}
	counts[n] = ones
	for v := 0; v < n; v++ {
		counts[v] = negativeCount(table, v)
		if counts[v] < ones-counts[v] {
			flip(table, v)
			phase |= 1 << uint(v)
			counts[v] = ones - counts[v]
		}
	}
	for i := 0; i < n-1; i++ {
		best := i + 1
		for k := i + 2; k < n; k++ {
			if counts[k] < counts[best] {
				best = k
			}
		}
		if counts[i] <= counts[best] {
			continue
		}
		counts[i], counts[best] = counts[best], counts[i]
		perm[i], perm[best] = perm[best], perm[i]
		phase = swapPhase(phase, i, best)
		swap(table, i, best)
	}
	work, best := make([]uint64, words), make([]uint64, words)
	apply := func(v int) bool {
		if counts[v] != counts[v+1] {
			return false
		}
		config := refine(table, work, best, v, counts[v] != counts[n]/2)
		if config == 0 {
			return false
		}
		if config&1 != 0 {
			phase ^= 1 << uint(v)
		}
		if config&2 != 0 {
			phase ^= 1 << uint(v+1)
		}
		if config&4 != 0 {
			phase = swapPhase(phase, v, v+1)
			perm[v], perm[v+1] = perm[v+1], perm[v]
		}
		return true
	}
	for pass := 0; pass < 5; pass++ {
		changed := false
		for v := n - 2; v >= 0; v-- {
			changed = apply(v) || changed
		}
		if !changed {
			break
		}
		changed = false
		for v := 1; v < n-1; v++ {
			changed = apply(v) || changed
		}
		if !changed {
			break
		}
	}
	return phase
}

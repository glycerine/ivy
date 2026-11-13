package fra

// clausTranspose32 is the 32x32 bit transpose in fraClaus.c.
func clausTranspose32(a *[32]uint32) {
	mask := uint32(0x0000ffff)
	for shift := uint(16); shift != 0; {
		for k := uint(0); k < 32; k = ((k | shift) + 1) &^ shift {
			t := (a[k] ^ (a[k|shift] >> shift)) & mask
			a[k] ^= t
			a[k|shift] ^= t << shift
		}
		shift >>= 1
		mask ^= mask << shift
	}
}

// clausPatternScores counts simulated minterms, in the same leaf/bit order
// as Fra_ClausProcessClausesCut2 and the general Cut3 implementation.
func clausPatternScores(leaves [][]uint32, skip int) []int32 {
	if len(leaves) < 2 || len(leaves) > 16 || skip < 0 {
		panic("invalid simulated cut")
	}
	count := len(leaves[0])
	if skip > count {
		panic("invalid simulation prefix")
	}
	for _, leaf := range leaves {
		if len(leaf) != count {
			panic("unequal simulation lengths")
		}
	}
	scores := make([]int32, 1<<uint(len(leaves)))
	for word := skip; word < count; word++ {
		for bit := uint(0); bit < 32; bit++ {
			pattern := 0
			for leaf, values := range leaves {
				if values[word]&(uint32(1)<<bit) != 0 {
					pattern |= 1 << uint(leaf)
				}
			}
			scores[pattern]++
		}
	}
	return scores
}

// clausSelectCosts preserves source ordering among equal-cost candidates.
func clausSelectCosts(costs []int32, maxClauses, simulationWords int) (int32, int) {
	if len(costs) <= maxClauses || simulationWords < 0 {
		panic("invalid clause selection")
	}
	limit := simulationWords*32 + 1
	histogram := make([]int, limit)
	for _, cost := range costs {
		if cost == -1 {
			continue
		}
		if cost <= 0 || int(cost) >= limit {
			panic("invalid clause cost")
		}
		histogram[cost]++
	}
	selected, threshold := 0, limit-1
	for ; threshold > 0; threshold-- {
		selected += histogram[threshold]
		if selected >= maxClauses {
			break
		}
	}
	selected = 0
	for i, cost := range costs {
		if cost >= int32(threshold) && selected < maxClauses {
			selected++
		} else {
			costs[i] = -1
		}
	}
	return int32(threshold), selected
}

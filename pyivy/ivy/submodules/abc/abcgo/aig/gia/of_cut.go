package gia

// ofMergeCuts merges sorted leaf sets without modifying dst on overflow.
func ofMergeCuts(dst *ofCut, limit int32, inputs ...*ofCut) bool {
	if limit < 0 || limit > 6 || len(inputs) > 3 {
		panic("gia: invalid Of cut merge")
	}
	var leaves [7]int32
	var positions [3]uint32
	count := int32(0)
	var sign uint64
	for _, c := range inputs {
		if c.size > 6 {
			panic("gia: invalid Of cut")
		}
		sign |= c.sign
	}
	for {
		var smallest int32
		found := false
		for i, c := range inputs {
			if positions[i] < c.size && (!found || c.leaves[positions[i]] < smallest) {
				smallest = c.leaves[positions[i]]
				found = true
			}
		}
		if !found {
			break
		}
		if count == limit {
			return false
		}
		leaves[count] = smallest
		count++
		for i, c := range inputs {
			if positions[i] < c.size && c.leaves[positions[i]] == smallest {
				positions[i]++
			}
		}
	}
	dst.leaves = leaves
	dst.size = uint32(count)
	dst.sign = sign
	dst.function = 0x7ffffff
	return true
}
func ofContains(base, cut *ofCut) bool {
	if cut.size > base.size {
		return false
	}
	j := uint32(0)
	for i := uint32(0); i < base.size && j < cut.size; i++ {
		if base.leaves[i] == cut.leaves[j] {
			j++
		} else if base.leaves[i] > cut.leaves[j] {
			return false
		}
	}
	return j == cut.size
}
func ofLastContained(cuts []*ofCut, n int32) bool {
	candidate := cuts[n]
	for _, c := range cuts[:n] {
		if c.size <= candidate.size && c.sign&candidate.sign == c.sign && ofContains(candidate, c) {
			return true
		}
	}
	return false
}
func ofCompare(a, b *ofCut) int {
	if a.delay != b.delay {
		if a.delay < b.delay {
			return -1
		}
		return 1
	}
	if a.flow != b.flow {
		if a.flow < b.flow {
			return -1
		}
		return 1
	}
	if a.size < b.size {
		return -1
	}
	if a.size > b.size {
		return 1
	}
	return 0
}
func ofAddCut(cuts []*ofCut, n, limit int32) int32 {
	if n == 0 {
		return 1
	}
	candidate := cuts[n]
	write := int32(0)
	for read := int32(0); read < n; read++ {
		c := cuts[read]
		if candidate.size < c.size && candidate.sign&c.sign == candidate.sign && ofContains(c, candidate) {
			c.size = 31
			continue
		}
		cuts[write], cuts[read] = cuts[read], cuts[write]
		write++
	}
	cuts[write], cuts[n] = cuts[n], cuts[write]
	n = write
	for i := n; i > 0; i-- {
		if ofCompare(cuts[i-1], cuts[i]) < 0 {
			break
		}
		cuts[i-1], cuts[i] = cuts[i], cuts[i-1]
	}
	return min(n+1, limit-1)
}

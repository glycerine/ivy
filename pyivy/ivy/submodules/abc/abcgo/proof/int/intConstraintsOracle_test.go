package int

func constrainedReachableBad(bad, forbidden, next0, next1 uint8) bool {
	seen := [4]bool{true}
	for changed := true; changed; {
		changed = false
		for state := 0; state < 4; state++ {
			if !seen[state] {
				continue
			}
			for input := 0; input < 2; input++ {
				index := input | state<<1
				if forbidden&(1<<index) != 0 {
					continue
				}
				if bad&(1<<index) != 0 {
					return true
				}
				next := int(next0>>index)&1 | (int(next1>>index)&1)<<1
				if !seen[next] {
					seen[next], changed = true, true
				}
			}
		}
	}
	return false
}

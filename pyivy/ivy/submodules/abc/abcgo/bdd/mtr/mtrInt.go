package mtr

// Internal interval checks use uint64 to avoid overflow in low+size.
func contains(n *MtrNode, low, size uint32) bool {
	return n != nil && size > 0 && low >= n.Low && uint64(low)+uint64(size) <= uint64(n.Low)+uint64(n.Size)
}
func shiftValid(n *MtrNode, shift int64) bool {
	v := int64(n.Low) + shift
	if v < 0 || v+int64(n.Size)-1 > int64(MTR_MAXHIGH) {
		return false
	}
	if n.Flags&MTR_TERMINAL == 0 {
		for c := n.Child; c != nil; c = c.Younger {
			if !shiftValid(c, shift) {
				return false
			}
		}
	}
	return true
}
func shiftTree(n *MtrNode, shift int64) {
	n.Low = uint32(int64(n.Low) + shift)
	if n.Flags&MTR_TERMINAL == 0 {
		for c := n.Child; c != nil; c = c.Younger {
			shiftTree(c, shift)
		}
	}
}
func mtrShiftHL(n *MtrNode, shift int) int {
	if n == nil || !shiftValid(n, int64(shift)) {
		return 0
	}
	shiftTree(n, int64(shift))
	return 1
}

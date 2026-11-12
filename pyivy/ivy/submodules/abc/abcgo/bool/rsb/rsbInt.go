package rsb

// Temporary storage is sized from the actual divisor/table slices rather than
// mirroring C vector length/capacity metadata.
func rsbWords(n int32) int {
	if n < 0 || n > 20 {
		panic("rsb: variable count out of range")
	}
	if n <= 6 {
		return 1
	}
	return 1 << uint(n-6)
}
func rsbBit(t []uint64, i int32) uint64 { return t[i>>6] >> uint(i&63) & 1 }
func rsbStretch(t uint64, n int32) uint64 {
	if n < 0 || n > 6 {
		panic("rsb: invalid small truth dimension")
	}
	width := uint(1) << uint(n)
	if width == 64 {
		return t
	}
	mask := uint64(1)<<width - 1
	t &= mask
	for width < 64 {
		t |= t << width
		width *= 2
	}
	return t
}

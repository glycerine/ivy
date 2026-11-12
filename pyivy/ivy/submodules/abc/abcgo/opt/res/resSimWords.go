package res

// Simulation entries are Go-owned word slices. The original C implementation
// occasionally viewed the same words as bytes; these helpers preserve that
// little-endian layout without unsafe pointer casts.
func resWords(entry any) []uint32 { return entry.([]uint32) }

func resByte(words []uint32, index int32) byte {
	return byte(words[index/4] >> (uint(index%4) * 8))
}

func setResByte(words []uint32, index int32, value byte) {
	shift := uint(index%4) * 8
	mask := uint32(0xff) << shift
	words[index/4] = (words[index/4] &^ mask) | uint32(value)<<shift
}

func clearResBytes(words []uint32, start, end int32) {
	for i := start; i < end; i++ {
		setResByte(words, i, 0)
	}
}

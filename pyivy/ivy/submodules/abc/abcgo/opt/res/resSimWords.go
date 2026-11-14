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

type resRandomStream interface{ next() uint32 }

// Abc_InfoRandomBytes mirrors the source's byte construction; its caller owns
// the random stream.
func Abc_InfoRandomBytes(p []uint32, nWords int32, rng resRandomStream) {
	for i := nWords - 1; i >= 0; i-- {
		num := rng.next()
		var word uint32
		for bit := uint(0); bit < 4; bit++ {
			word <<= 8
			if num&(1<<bit) != 0 {
				word |= 0xff
			}
		}
		p[i] = word
	}
}

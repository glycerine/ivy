package gia

// Tas_StorePatternTry merges a literal assignment into one simulation bit.
// The presence bitmap distinguishes an unspecified bit from a zero value.
func Tas_StorePatternTry(info, present [][]uint32, bit int32, literals []int32, count int32) int32 {
	if bit < 0 || count < 0 || int(count) > len(literals) {
		panic("gia: invalid pattern bounds")
	}
	word, mask := int(bit)/32, uint32(1)<<uint(bit%32)
	for _, lit := range literals[:count] {
		variable := int(lit >> 1)
		if lit < 0 || variable >= len(info) || variable >= len(present) || word >= len(info[variable]) || word >= len(present[variable]) {
			panic("gia: pattern literal out of bounds")
		}
		isSet := info[variable][word]&mask != 0
		if present[variable][word]&mask != 0 && isSet == (lit&1 != 0) {
			return 0
		}
	}
	for _, lit := range literals[:count] {
		variable := int(lit >> 1)
		present[variable][word] |= mask
		if (info[variable][word]&mask != 0) == (lit&1 != 0) {
			info[variable][word] ^= mask
		}
	}
	return 1
}

// Tas_StorePattern tries the 31 usable bits in the first simulation word.
func Tas_StorePattern(info, present [][]uint32, literals []int32) int32 {
	for bit := int32(1); bit < 32; bit++ {
		if Tas_StorePatternTry(info, present, bit, literals, int32(len(literals))) != 0 {
			return 1
		}
	}
	return 0
}

package ssw

// word matches ABC's 64-bit simulation word.
type word = uint64

func Ssw_RarBitWordNum(nBits int32) int32 {
	words := nBits >> 6
	if nBits&63 != 0 {
		words++
	}
	return words
}

// transpose32 and transpose64 preserve the bit ordering used by ABC's
// bit-matrix transposition in sswRarity.c, credited there to Henry Warren's
// Hacker's Delight.
func transpose32(a []uint32) {
	mask := uint32(0x0000ffff)
	for shift := 16; shift != 0; {
		for k := 0; k < 32; k = (k + shift + 1) &^ shift {
			t := (a[k] ^ (a[k+shift] >> shift)) & mask
			a[k] ^= t
			a[k+shift] ^= t << shift
		}
		shift >>= 1
		mask ^= mask << shift
	}
}

func transpose64(a []word) {
	mask := word(0x00000000ffffffff)
	for shift := 32; shift != 0; {
		for k := 0; k < 64; k = (k + shift + 1) &^ shift {
			t := (a[k] ^ (a[k+shift] >> shift)) & mask
			a[k] ^= t
			a[k+shift] ^= t << shift
		}
		shift >>= 1
		mask ^= mask << shift
	}
}

func transpose64Simple(a, b []word) {
	for i := 0; i < 64; i++ {
		b[i] = 0
	}
	for i := 0; i < 64; i++ {
		for k := 0; k < 64; k++ {
			if a[i]>>k&1 != 0 {
				b[63-k] |= word(1) << (63 - i)
			}
		}
	}
}

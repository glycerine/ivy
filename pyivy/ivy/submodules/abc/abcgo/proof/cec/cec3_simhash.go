package cec

// Cec3_ManSimHashKey hashes a phase-normalized 64-bit simulation signature.
// ABC treats each word as two 32-bit words and combines products modulo 2^32.
func Cec3_ManSimHashKey(sim []uint64, nSims, tableSize int32) int32 {
	if nSims <= 0 || int(nSims) > len(sim) || tableSize <= 0 {
		panic("cec: invalid simulation hash dimensions")
	}
	primes := [...]uint32{1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147}
	invert := sim[0]&1 != 0
	var hash uint32
	for i, word := range sim[:nSims] {
		lo, hi := uint32(word), uint32(word>>32)
		if invert {
			lo, hi = ^lo, ^hi
		}
		hash ^= lo * primes[(2*i)&15]
		hash ^= hi * primes[(2*i+1)&15]
	}
	return int32(hash % uint32(tableSize))
}

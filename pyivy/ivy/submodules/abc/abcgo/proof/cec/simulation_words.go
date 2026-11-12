package cec

import "math/bits"

// CEC compares signatures modulo a common complement. The first bit fixes
// each signature's phase, allowing constants and equivalence classes to share
// the same deterministic normalization used by the original simulator.
func Cec_ManSimCompareConst(p []uint32, nWords int32) int32 {
	var phase uint32
	if p[0]&1 != 0 {
		phase = ^uint32(0)
	}
	for _, w := range p[:nWords] {
		if w != phase {
			return 0
		}
	}
	return 1
}
func Cec_ManSimCompareEqual(a, b []uint32, nWords int32) int32 {
	var phase uint32
	if (a[0]^b[0])&1 != 0 {
		phase = ^uint32(0)
	}
	for i, w := range a[:nWords] {
		if w != b[i]^phase {
			return 0
		}
	}
	return 1
}
func Cec_ManSimCompareConstFirstBit(p []uint32, nWords int32) int32 {
	var phase uint32
	if p[0]&1 != 0 {
		phase = ^uint32(0)
	}
	for i, w := range p[:nWords] {
		if x := w ^ phase; x != 0 {
			return int32(i*32 + bits.TrailingZeros32(x))
		}
	}
	return -1
}
func Cec_ManSimCompareEqualFirstBit(a, b []uint32, nWords int32) int32 {
	var phase uint32
	if (a[0]^b[0])&1 != 0 {
		phase = ^uint32(0)
	}
	for i, w := range a[:nWords] {
		if x := w ^ b[i] ^ phase; x != 0 {
			return int32(i*32 + bits.TrailingZeros32(x))
		}
	}
	return -1
}
func cecSimScore(x uint32, base int, scores []int32) {
	for x != 0 {
		b := bits.TrailingZeros32(x)
		scores[base+b]++
		x &= x - 1
	}
}
func Cec_ManSimCompareConstScore(p []uint32, nWords int32, scores []int32) {
	var phase uint32
	if p[0]&1 != 0 {
		phase = ^uint32(0)
	}
	for i, w := range p[:nWords] {
		cecSimScore(w^phase, i*32, scores)
	}
}
func Cec_ManSimCompareEqualScore(a, b []uint32, nWords int32, scores []int32) {
	var phase uint32
	if (a[0]^b[0])&1 != 0 {
		phase = ^uint32(0)
	}
	for i, w := range a[:nWords] {
		cecSimScore(w^b[i]^phase, i*32, scores)
	}
}
func Cec_ManSimHashKey(sim []uint32, nWords, nTableSize int32) int32 {
	primes := [16]uint32{1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147}
	var phase, hash uint32
	if sim[0]&1 != 0 {
		phase = ^uint32(0)
	}
	for i, w := range sim[:nWords] {
		hash ^= (w ^ phase) * primes[i&15]
	}
	return int32(hash % uint32(nTableSize))
}

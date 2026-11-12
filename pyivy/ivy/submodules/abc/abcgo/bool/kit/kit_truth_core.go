package kit

import "math/bits"

func Kit_SuppIsMinBase(s int32) int32 {
	if s&(s+1) == 0 {
		return 1
	}
	return 0
}
func Kit_BitWordNum(n int32) int32 {
	if n < 0 {
		panic("negative bit count")
	}
	words := n / 32
	if n%32 != 0 {
		words++
	}
	return words
}
func Kit_TruthWordNum(n int32) int32 {
	if n < 0 || n > 30 {
		panic("truth variable count out of range")
	}
	if n <= 5 {
		return 1
	}
	return 1 << uint(n-5)
}
func Kit_BitMask(n int32) uint32 {
	if n < 0 || n > 32 {
		panic("bit mask count out of range")
	}
	if n == 32 {
		return ^uint32(0)
	}
	return (uint32(1) << uint(n)) - 1
}
func Kit_TruthSetBit(p []uint32, b int32) { p[b>>5] |= 1 << uint(b&31) }
func Kit_TruthXorBit(p []uint32, b int32) { p[b>>5] ^= 1 << uint(b&31) }
func Kit_TruthHasBit(p []uint32, b int32) int32 {
	if p[b>>5]&(1<<uint(b&31)) != 0 {
		return 1
	}
	return 0
}
func Kit_WordFindFirstBit(w uint32) int32 {
	if w == 0 {
		return -1
	}
	return int32(bits.TrailingZeros32(w))
}
func Kit_WordHasOneBit(w uint32) int32 {
	if w&(w-1) == 0 {
		return 1
	}
	return 0
}
func Kit_WordCountOnes(w uint32) int32 { return int32(bits.OnesCount32(w)) }
func Kit_TruthCountOnes(p []uint32, n int32) int32 {
	var total int32
	for _, x := range p[:Kit_TruthWordNum(n)] {
		total += Kit_WordCountOnes(x)
	}
	return total
}
func Kit_TruthFindFirstBit(p []uint32, n int32) int32 {
	for i, x := range p[:Kit_TruthWordNum(n)] {
		if x != 0 {
			return int32(i)*32 + Kit_WordFindFirstBit(x)
		}
	}
	return -1
}
func Kit_TruthFindFirstZero(p []uint32, n int32) int32 {
	for i, x := range p[:Kit_TruthWordNum(n)] {
		if ^x != 0 {
			return int32(i)*32 + Kit_WordFindFirstBit(^x)
		}
	}
	return -1
}
func Kit_TruthIsEqualWithPhase(a, b []uint32, n int32) int32 {
	if a[0]&1 == b[0]&1 {
		return Kit_TruthIsEqual(a, b, n)
	}
	return Kit_TruthIsOpposite(a, b, n)
}
func Kit_TruthCopy(out, in []uint32, n int32) {
	copy(out[:Kit_TruthWordNum(n)], in[:Kit_TruthWordNum(n)])
}
func Kit_TruthClear(out []uint32, n int32) { clear(out[:Kit_TruthWordNum(n)]) }
func Kit_TruthFill(out []uint32, n int32) {
	for i := range out[:Kit_TruthWordNum(n)] {
		out[i] = ^uint32(0)
	}
}
func Kit_TruthNot(out, in []uint32, n int32) {
	for i := range out[:Kit_TruthWordNum(n)] {
		out[i] = ^in[i]
	}
}
func Kit_TruthAndPhase(out, a, b []uint32, n, c0, c1 int32) {
	for i := range out[:Kit_TruthWordNum(n)] {
		x, y := a[i], b[i]
		if c0 != 0 {
			x = ^x
		}
		if c1 != 0 {
			y = ^y
		}
		out[i] = x & y
	}
}
func Kit_TruthOrPhase(out, a, b []uint32, n, c0, c1 int32) {
	for i := range out[:Kit_TruthWordNum(n)] {
		x, y := a[i], b[i]
		if c0 != 0 {
			x = ^x
		}
		if c1 != 0 {
			y = ^y
		}
		out[i] = x | y
	}
}
func Kit_TruthMux(out, a, b, ctrl []uint32, n int32) { Kit_TruthMuxPhase(out, a, b, ctrl, n, 0) }
func Kit_TruthMuxPhase(out, a, b, ctrl []uint32, n, compl int32) {
	for i := range out[:Kit_TruthWordNum(n)] {
		x := a[i]
		if compl != 0 {
			x = ^x
		}
		out[i] = x&^ctrl[i] | b[i]&ctrl[i]
	}
}
func Kit_TruthIthVar(out []uint32, n, v int32) {
	if v < 0 || v >= n {
		panic("variable outside truth table")
	}
	masks := [5]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	for i := range out[:Kit_TruthWordNum(n)] {
		if v < 5 {
			out[i] = masks[v]
		} else if i&(1<<uint(v-5)) != 0 {
			out[i] = ^uint32(0)
		} else {
			out[i] = 0
		}
	}
}
func Kit_TruthIsEqual(a, b []uint32, n int32) int32 {
	for i := int32(0); i < Kit_TruthWordNum(n); i++ {
		if a[i] != b[i] {
			return 0
		}
	}
	return 1
}
func Kit_TruthIsEqualWithCare(a, b, care []uint32, n int32) int32 {
	for i := int32(0); i < Kit_TruthWordNum(n); i++ {
		if (a[i] & care[i]) != (b[i] & care[i]) {
			return 0
		}
	}
	return 1
}
func Kit_TruthIsOpposite(a, b []uint32, n int32) int32 {
	for i := int32(0); i < Kit_TruthWordNum(n); i++ {
		if a[i] != ^b[i] {
			return 0
		}
	}
	return 1
}
func Kit_TruthIsConst0(a []uint32, n int32) int32 {
	for i := int32(0); i < Kit_TruthWordNum(n); i++ {
		if a[i] != 0 {
			return 0
		}
	}
	return 1
}
func Kit_TruthIsConst1(a []uint32, n int32) int32 {
	for i := int32(0); i < Kit_TruthWordNum(n); i++ {
		if a[i] != ^uint32(0) {
			return 0
		}
	}
	return 1
}
func Kit_TruthIsImply(a, b []uint32, n int32) int32 {
	for i := int32(0); i < Kit_TruthWordNum(n); i++ {
		if a[i]&^b[i] != 0 {
			return 0
		}
	}
	return 1
}
func Kit_TruthIsDisjoint(a, b []uint32, n int32) int32 {
	for i := int32(0); i < Kit_TruthWordNum(n); i++ {
		if a[i]&b[i] != 0 {
			return 0
		}
	}
	return 1
}
func Kit_TruthIsDisjoint3(a, b, c []uint32, n int32) int32 {
	for i := int32(0); i < Kit_TruthWordNum(n); i++ {
		if a[i]&b[i]&c[i] != 0 {
			return 0
		}
	}
	return 1
}
func Kit_TruthAnd(out, a, b []uint32, n int32) {
	for i := range out[:Kit_TruthWordNum(n)] {
		out[i] = a[i] & b[i]
	}
}
func Kit_TruthOr(out, a, b []uint32, n int32) {
	for i := range out[:Kit_TruthWordNum(n)] {
		out[i] = a[i] | b[i]
	}
}
func Kit_TruthXor(out, a, b []uint32, n int32) {
	for i := range out[:Kit_TruthWordNum(n)] {
		out[i] = a[i] ^ b[i]
	}
}
func Kit_TruthSharp(out, a, b []uint32, n int32) {
	for i := range out[:Kit_TruthWordNum(n)] {
		out[i] = a[i] &^ b[i]
	}
}
func Kit_TruthNand(out, a, b []uint32, n int32) {
	for i := range out[:Kit_TruthWordNum(n)] {
		out[i] = ^(a[i] & b[i])
	}
}

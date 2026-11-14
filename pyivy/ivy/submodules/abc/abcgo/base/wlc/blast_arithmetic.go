package wlc

import "github.com/glycerine/abcgo/aig/gia"

func Wlc_BlastLess2(g *gia.Gia_Man_t, a, b []int32, n int32) int32 {
	known, result := int32(0), int32(0)
	for i := n - 1; i >= 0; i-- {
		result = gia.Gia_ManHashMux(g, known, result, gia.Gia_ManHashAnd(g, a[i]^1, b[i]))
		known = gia.Gia_ManHashOr(g, known, gia.Gia_ManHashXor(g, a[i], b[i]))
		if known == 1 {
			break
		}
	}
	return result
}
func Wlc_BlastLess_rec(g *gia.Gia_Man_t, a, b []int32, n int32, yes, no *int32) {
	if n > 1 {
		y := gia.Gia_ManHashAnd(g, a[n-1]^1, b[n-1])
		no0 := gia.Gia_ManHashAnd(g, b[n-1]^1, a[n-1])
		if y == 1 || no0 == 1 {
			*yes, *no = y, no0
			return
		}
		var yr, nr int32
		Wlc_BlastLess_rec(g, a, b, n-1, &yr, &nr)
		*yes = gia.Gia_ManHashOr(g, y, gia.Gia_ManHashAnd(g, no0^1, yr))
		*no = gia.Gia_ManHashOr(g, no0, gia.Gia_ManHashAnd(g, y^1, nr))
		return
	}
	if n != 1 {
		panic("invalid comparison width")
	}
	*yes = gia.Gia_ManHashAnd(g, a[0]^1, b[0])
	*no = gia.Gia_ManHashAnd(g, b[0]^1, a[0])
}
func Wlc_BlastLess(g *gia.Gia_Man_t, a, b []int32, n int32) int32 {
	if n == 0 {
		return 0
	}
	var yes, no int32
	Wlc_BlastLess_rec(g, a, b, n, &yes, &no)
	return yes
}
func Wlc_BlastLessSigned(g *gia.Gia_Man_t, a, b []int32, n int32) int32 {
	diff := gia.Gia_ManHashXor(g, a[n-1], b[n-1])
	return gia.Gia_ManHashMux(g, diff, a[n-1], Wlc_BlastLess(g, a, b, n-1))
}
func Wlc_BlastFullAdder(g *gia.Gia_Man_t, a, b, c int32, carry, sum *int32) {
	complement := a == 1 || b == 1 || c == 1
	if complement {
		a ^= 1
		b ^= 1
		c ^= 1
	}
	and1 := gia.Gia_ManHashAnd(g, a, b)
	and1not := gia.Gia_ManHashAnd(g, a^1, b^1)
	xor := gia.Gia_ManHashAnd(g, and1^1, and1not^1)
	and2 := gia.Gia_ManHashAnd(g, c, xor)
	and2not := gia.Gia_ManHashAnd(g, c^1, xor^1)
	*sum = gia.Gia_ManHashAnd(g, and2^1, and2not^1)
	*carry = gia.Gia_ManHashOr(g, and1, and2)
	if complement {
		*sum ^= 1
		*carry ^= 1
	}
}
func Wlc_BlastAdder(g *gia.Gia_Man_t, a, b []int32, n, carry int32) {
	for i := int32(0); i < n; i++ {
		Wlc_BlastFullAdder(g, a[i], b[i], carry, &carry, &a[i])
	}
}
func Wlc_BlastSubtract(g *gia.Gia_Man_t, a, b []int32, n, carry int32) {
	for i := int32(0); i < n; i++ {
		Wlc_BlastFullAdder(g, a[i], b[i]^1, carry, &carry, &a[i])
	}
}

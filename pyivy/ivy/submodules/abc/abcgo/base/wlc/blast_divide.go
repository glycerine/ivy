package wlc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Wlc_BlastDivider follows ABC's restoring unsigned divider; equal operand
// widths are required by the original implementation.
func Wlc_BlastDivider(g *gia.Gia_Man_t, num []int32, n int32, div []int32, ndiv, quotient int32, res *vec.Vec_Int_t) {
	if n != ndiv {
		panic("divider operand widths differ")
	}
	out := Wlc_VecCopy(res, num, n)
	quo, tmp := make([]int32, n), make([]int32, n)
	for j := n - 1; j >= 0; j-- {
		known := int32(0)
		for i := n - 1; i > n-1-j; i-- {
			known = gia.Gia_ManHashOr(g, known, div[i])
			if known == 1 {
				break
			}
		}
		quo[j] = known
		for i := n - 1; i >= 0; i-- {
			if known == 1 {
				break
			}
			y := int32(0)
			if i >= j {
				y = div[i-j]
			}
			quo[j] = gia.Gia_ManHashMux(g, known, quo[j], gia.Gia_ManHashAnd(g, y, out[i]^1))
			known = gia.Gia_ManHashOr(g, known, gia.Gia_ManHashXor(g, y, out[i]))
		}
		quo[j] ^= 1
		if quo[j] == 0 {
			continue
		}
		borrow := int32(0)
		for i := int32(0); i < n; i++ {
			top := gia.Gia_ManHashMux(g, borrow, out[i]^1, out[i])
			y := int32(0)
			if i >= j {
				y = div[i-j]
			}
			borrow = gia.Gia_ManHashMux(g, out[i], gia.Gia_ManHashAnd(g, borrow, y), gia.Gia_ManHashOr(g, borrow, y))
			tmp[i] = gia.Gia_ManHashXor(g, top, y)
		}
		if quo[j] == 1 {
			out = Wlc_VecCopy(res, tmp, n)
		} else {
			for i := int32(0); i < n; i++ {
				out[i] = gia.Gia_ManHashMux(g, quo[j], tmp[i], out[i])
			}
		}
	}
	if quotient != 0 {
		Wlc_VecCopy(res, quo, n)
	}
}

func Wlc_BlastDividerSigned(g *gia.Gia_Man_t, num []int32, n int32, div []int32, ndiv, quotient int32, res *vec.Vec_Int_t) {
	vn, vd := vec.Vec_IntAlloc(int(n)), vec.Vec_IntAlloc(int(ndiv))
	r00, r01, r10, r11 := vec.Vec_IntAlloc(int(n)), vec.Vec_IntAlloc(int(n)), vec.Vec_IntAlloc(int(n)), vec.Vec_IntAlloc(int(n))
	neg := vec.Vec_IntAlloc(int(n))
	diff := gia.Gia_ManHashXor(g, num[n-1], div[ndiv-1])
	Wlc_BlastMinus(g, num, n, vn)
	Wlc_BlastMinus(g, div, ndiv, vd)
	Wlc_BlastDivider(g, num, n, div, ndiv, quotient, r00)
	Wlc_BlastDivider(g, num, n, vec.Vec_IntArray(vd), ndiv, quotient, r01)
	Wlc_BlastDivider(g, vec.Vec_IntArray(vn), n, div, ndiv, quotient, r10)
	Wlc_BlastDivider(g, vec.Vec_IntArray(vn), n, vec.Vec_IntArray(vd), ndiv, quotient, r11)
	vec.Vec_IntClear(res)
	for i := int32(0); i < n; i++ {
		d0 := gia.Gia_ManHashMux(g, div[ndiv-1], vec.Vec_IntEntry(r01, int(i)), vec.Vec_IntEntry(r00, int(i)))
		d1 := gia.Gia_ManHashMux(g, div[ndiv-1], vec.Vec_IntEntry(r11, int(i)), vec.Vec_IntEntry(r10, int(i)))
		vec.Vec_IntPush(res, gia.Gia_ManHashMux(g, num[n-1], d1, d0))
	}
	Wlc_BlastMinus(g, vec.Vec_IntArray(res), n, neg)
	phase := num[n-1]
	if quotient != 0 {
		phase = diff
	}
	for i := int32(0); i < n; i++ {
		vec.Vec_IntWriteEntry(res, int(i), gia.Gia_ManHashMux(g, phase, vec.Vec_IntEntry(neg, int(i)), vec.Vec_IntEntry(res, int(i))))
	}
}

// Wlc_BlastZeroCondition gates each result bit with the OR of divisor bits.
func Wlc_BlastZeroCondition(g *gia.Gia_Man_t, div []int32, n int32, res *vec.Vec_Int_t) {
	nonzero := int32(0)
	for i := int32(0); i < n; i++ {
		nonzero = gia.Gia_ManHashOr(g, nonzero, div[i])
	}
	for i, value := range vec.Vec_IntArray(res) {
		vec.Vec_IntWriteEntry(res, i, gia.Gia_ManHashAnd(g, nonzero, value))
	}
}

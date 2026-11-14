package wlc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Wlc_BlastMinus(g *gia.Gia_Man_t, num []int32, n int32, res *vec.Vec_Int_t) {
	out := Wlc_VecCopy(res, num, n)
	invert := int32(0)
	for i := int32(0); i < n; i++ {
		out[i] = gia.Gia_ManHashMux(g, invert, out[i]^1, out[i])
		invert = gia.Gia_ManHashOr(g, invert, num[i])
	}
}
func Wlc_BlastMultiplier2(g *gia.Gia_Man_t, a, b []int32, n int32, temp, res *vec.Vec_Int_t) {
	vec.Vec_IntFill(res, int(n), 0)
	for i := int32(0); i < n; i++ {
		vec.Vec_IntFill(temp, int(i), 0)
		for j := int32(0); vec.Vec_IntSize(temp) < int(n); j++ {
			vec.Vec_IntPush(temp, gia.Gia_ManHashAnd(g, a[j], b[i]))
		}
		Wlc_BlastAdder(g, vec.Vec_IntArray(res), vec.Vec_IntArray(temp), n, 0)
	}
}
func Wlc_BlastFullAdderCtrl(g *gia.Gia_Man_t, a, ac, b, c int32, carry, sum *int32, neg int32) {
	product := gia.Gia_ManHashAnd(g, a, ac)
	if neg != 0 {
		product ^= 1
	}
	Wlc_BlastFullAdder(g, product, b, c, carry, sum)
}
func Wlc_BlastFullAdderSubtr(g *gia.Gia_Man_t, a, b, c int32, carry, sum *int32, sub int32) {
	Wlc_BlastFullAdder(g, gia.Gia_ManHashXor(g, a, sub), b, c, carry, sum)
}
func Wlc_BlastMultiplier(g *gia.Gia_Man_t, a, b []int32, nA, nB int32, temp, res *vec.Vec_Int_t, signed int32) {
	if nA <= 0 || nB <= 0 || (signed != 0 && signed != 1) {
		panic("invalid multiplier parameters")
	}
	vec.Vec_IntFill(res, int(nA+nB), 0)
	out := vec.Vec_IntArray(res)
	vec.Vec_IntFill(temp, int(2*nA), 0)
	scratch := vec.Vec_IntArray(temp)
	carries, sums := scratch[:nA], scratch[nA:]
	for j := int32(0); j < nB; j++ {
		for i := int32(0); i < nA; i++ {
			target := &out[j]
			if i != 0 {
				target = &sums[i-1]
			}
			neg := int32(0)
			if signed != 0 && ((i+1 == nA) != (j+1 == nB)) {
				neg = 1
			}
			Wlc_BlastFullAdderCtrl(g, a[i], b[j], sums[i], carries[i], &carries[i], target, neg)
		}
	}
	sums[nA-1] = signed
	carry := signed
	for i := int32(0); i < nA; i++ {
		Wlc_BlastFullAdderCtrl(g, 1, carries[i], sums[i], carry, &carry, &out[nB+i], 0)
	}
}

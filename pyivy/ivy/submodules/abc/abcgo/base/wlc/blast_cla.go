package wlc

import "github.com/glycerine/abcgo/aig/gia"

func Wlc_BlastAdderCLA_one(g *gia.Gia_Man_t, gen, pro, car []int32, gen1, pro1, car1 *int32) {
	temp := gia.Gia_ManHashAnd(g, gen[0], pro[1])
	*pro1 = gia.Gia_ManHashAnd(g, pro[0], pro[1])
	*gen1 = gia.Gia_ManHashOr(g, gia.Gia_ManHashOr(g, gen[1], temp), gia.Gia_ManHashAnd(g, *pro1, car[0]))
	*car1 = gia.Gia_ManHashOr(g, gen[0], gia.Gia_ManHashAnd(g, pro[0], car[0]))
}
func Wlc_BlastAdderCLA_rec(g *gia.Gia_Man_t, gen, pro, car []int32, n int32, gen1, pro1 *int32) {
	if n == 2 {
		Wlc_BlastAdderCLA_one(g, gen, pro, car, gen1, pro1, &car[1])
		return
	}
	if n%2 != 0 {
		panic("CLA width must be even")
	}
	var g2, p2 [2]int32
	Wlc_BlastAdderCLA_rec(g, gen, pro, car, n/2, &g2[0], &p2[0])
	car[n/2] = g2[0]
	Wlc_BlastAdderCLA_rec(g, gen[n/2:], pro[n/2:], car[n/2:], n/2, &g2[1], &p2[1])
	Wlc_BlastAdderCLA_one(g, g2[:], p2[:], car, gen1, pro1, &car[n/2])
}
func Wlc_BlastAdderCLA_int(g *gia.Gia_Man_t, a, b []int32, n, carry int32) {
	if n == 1 {
		Wlc_BlastFullAdder(g, a[0], b[0], carry, &carry, &a[0])
		return
	}
	if n < 2 {
		panic("invalid CLA width")
	}
	gen, pro, car := make([]int32, n), make([]int32, n), make([]int32, n+1)
	car[0] = carry
	for i := int32(0); i < n; i++ {
		gen[i] = gia.Gia_ManHashAnd(g, a[i], b[i])
		pro[i] = gia.Gia_ManHashXor(g, a[i], b[i])
	}
	var go1, po1 int32
	Wlc_BlastAdderCLA_rec(g, gen, pro, car, n, &go1, &po1)
	for i := int32(0); i < n; i++ {
		a[i] = gia.Gia_ManHashXor(g, pro[i], car[i])
	}
}
func Wlc_BlastAdderCLA(g *gia.Gia_Man_t, a, b []int32, n, sign, carry int32) {
	size := int32(1)
	for size < n {
		size *= 2
	}
	aa, bb := make([]int32, size), make([]int32, size)
	copy(aa, a[:n])
	copy(bb, b[:n])
	if sign != 0 {
		for i := n; i < size; i++ {
			aa[i] = a[n-1]
			bb[i] = b[n-1]
		}
	}
	Wlc_BlastAdderCLA_int(g, aa, bb, size, carry)
	copy(a[:n], aa[:n])
}

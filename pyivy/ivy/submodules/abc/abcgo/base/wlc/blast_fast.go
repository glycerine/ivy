package wlc

import "github.com/glycerine/abcgo/aig/gia"

func Wlc_BlastAdderFast_int(g *gia.Gia_Man_t, a, b []int32, log2, carry int32) {
	n := int32(1) << log2
	gen, pro, original := make([]int32, n+1), make([]int32, n+1), make([]int32, n+1)
	if n == 1 {
		Wlc_BlastFullAdder(g, a[0], b[0], carry, &carry, &a[0])
		return
	}
	gen[0] = carry
	for i := int32(1); i <= n; i++ {
		gen[i] = gia.Gia_ManHashAnd(g, a[i-1], b[i-1])
		pro[i] = gia.Gia_ManHashXor(g, a[i-1], b[i-1])
		original[i] = pro[i]
	}
	combine := func(i, j int32) {
		g0 := gia.Gia_ManHashOr(g, gen[i], gia.Gia_ManHashAnd(g, pro[i], gen[j]))
		p0 := gia.Gia_ManHashAnd(g, pro[i], pro[j])
		gen[i], pro[i] = g0, p0
	}
	for i := int32(1); i <= n; i += 2 {
		combine(i, i-1)
	}
	for i := int32(1); i < log2-1; i++ {
		for j := 1 + 2*i; j <= n; j += 2 {
			combine(j, j-2*i)
		}
	}
	for i := n/2 + 1; i <= n; i += 2 {
		combine(i, i-n/2)
	}
	for i := int32(2); i <= n; i += 2 {
		combine(i, i-1)
	}
	for i := int32(0); i < n; i++ {
		a[i] = gia.Gia_ManHashXor(g, original[i+1], gen[i])
	}
	a[n] = gen[n]
}
func Wlc_BlastAdderFast(g *gia.Gia_Man_t, a, b []int32, n, sign, carry int32) {
	size, log := int32(1), int32(0)
	for size < n {
		size *= 2
		log++
	}
	aa, bb := make([]int32, size+1), make([]int32, size+1)
	copy(aa, a[:n])
	copy(bb, b[:n])
	if sign != 0 {
		for i := n; i < size; i++ {
			aa[i] = a[n-1]
			bb[i] = b[n-1]
		}
	}
	Wlc_BlastAdderFast_int(g, aa, bb, log, carry)
	copy(a[:n+1], aa[:n+1])
}

package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func Sle_CutSize(c []int32) int32                    { return c[0] & 15 }
func Sle_CutSign(c []int32) int32                    { return int32(uint32(c[0]) >> 4) }
func Sle_CutSetSizeSign(size, signature int32) int32 { return int32(uint32(signature)<<4) | size }
func Sle_CutLeaves(c []int32) []int32                { return c[1 : 1+Sle_CutSize(c)] }
func Sle_CutIsUsed(c []int32) int32 {
	if c[1] != 0 {
		return 1
	}
	return 0
}
func Sle_CutSetUnused(c []int32)      { c[1] = 0 }
func Sle_ListCutNum(c []int32) int32  { return c[0] }
func Sle_CutCountBits(v uint32) int32 { return int32(bits.OnesCount32(v)) }
func Sle_CutMergeOrder(a, b, out []int32, limit int32) int32 {
	x, y := Sle_CutLeaves(a), Sle_CutLeaves(b)
	i, j, n := 0, 0, 0
	for i < len(x) || j < len(y) {
		if int32(n) >= limit {
			return 0
		}
		var value int32
		if j == len(y) || (i < len(x) && x[i] < y[j]) {
			value = x[i]
			i++
		} else if i == len(x) || y[j] < x[i] {
			value = y[j]
			j++
		} else {
			value = x[i]
			i++
			j++
		}
		out[n+1] = value
		n++
	}
	out[0] = Sle_CutSetSizeSign(int32(n), Sle_CutSign(a)|Sle_CutSign(b))
	return 1
}
func Sle_SetCutIsContainedOrder(base, cut []int32) int32 {
	a, b := Sle_CutLeaves(base), Sle_CutLeaves(cut)
	i := 0
	for _, x := range b {
		for i < len(a) && a[i] < x {
			i++
		}
		if i == len(a) || a[i] != x {
			return 0
		}
		i++
	}
	return 1
}
func sleEachCut(v *Vec_Int_t, visit func([]int32) bool) {
	a := vec.Vec_IntArray(v)
	pos := 1
	for i := int32(0); i < a[0]; i++ {
		n := int(Sle_CutSize(a[pos:]))
		if visit(a[pos : pos+n+1]) {
			return
		}
		pos += n + 1
	}
}
func Sle_SetLastCutIsContained(v *Vec_Int_t, base []int32) int32 {
	found := int32(0)
	sleEachCut(v, func(c []int32) bool {
		if Sle_CutIsUsed(c) != 0 && Sle_CutSize(c) <= Sle_CutSize(base) && Sle_CutSign(c)&Sle_CutSign(base) == Sle_CutSign(c) && Sle_SetCutIsContainedOrder(base, c) != 0 {
			found = 1
			return true
		}
		return false
	})
	return found
}
func Sle_SetAddCut(v *Vec_Int_t, cut []int32) {
	sleEachCut(v, func(base []int32) bool {
		if Sle_CutIsUsed(base) != 0 && Sle_CutSize(cut) < Sle_CutSize(base) && Sle_CutSign(cut)&Sle_CutSign(base) == Sle_CutSign(cut) && Sle_SetCutIsContainedOrder(base, cut) != 0 {
			Sle_CutSetUnused(base)
		}
		return false
	})
	vec.Vec_IntPushArray(v, cut, int(Sle_CutSize(cut)+1))
	vec.Vec_IntAddToEntry(v, 0, 1)
}

package gia

import "math/bits"

// Jf cuts store size, cost, and function literal in a packed header,
// followed by the leaf literals. The order of leaves is significant.
const (
	JF_LEAF_MAX = 8
	JF_WORD_MAX = 1 << (JF_LEAF_MAX - 6)
	JF_CUT_MAX  = 16
	JF_EPSILON  = 0.005
)

type Jf_Cut_t struct {
	Sign              word
	Flow              float32
	Time, iFunc, Cost int32
	pCut              [JF_LEAF_MAX + 2]int32
}

func Jf_CutSize(c []int32) int32       { return c[0] & 15 }
func Jf_CutCost(c []int32) int32       { return (c[0] >> 4) & 15 }
func Jf_CutFunc(c []int32) int32       { return int32(uint32(c[0]) >> 8) }
func Jf_CutSetAll(f, c, s int32) int32 { return f<<8 | c<<4 | s }
func Jf_CutSetSize(c []int32, s int32) {
	if s < 0 || s >= 16 {
		panic("invalid cut size")
	}
	c[0] ^= Jf_CutSize(c) ^ s
}
func Jf_CutSetCost(c []int32, v int32) {
	if v < 0 || v >= 16 {
		panic("invalid cut cost")
	}
	c[0] ^= (Jf_CutCost(c) ^ v) << 4
}
func Jf_CutSetFunc(c []int32, f int32) {
	if f < 0 {
		panic("invalid cut function")
	}
	c[0] ^= (Jf_CutFunc(c) ^ f) << 8
}
func Jf_CutFuncClass(c []int32) int32 { return Jf_CutFunc(c) >> 1 }
func Jf_CutFuncCompl(c []int32) int32 { return Jf_CutFunc(c) & 1 }
func Jf_CutLits(c []int32) []int32    { return c[1:] }
func Jf_CutLit(c []int32, i int32) int32 {
	if i == 0 {
		panic("zero cut index")
	}
	return c[i]
}
func Jf_CutVar(c []int32, i int32) int32   { return Jf_CutLit(c, i) >> 1 }
func Jf_CutIsTriv(c []int32, i int32) bool { return Jf_CutSize(c) == 1 && Jf_CutVar(c, 1) == i }
func Jf_CutCheck(c []int32) {
	for i := int32(2); i <= Jf_CutSize(c); i++ {
		for k := int32(1); k < i; k++ {
			if c[i] == c[k] {
				panic("duplicate cut literal")
			}
		}
	}
}
func Jf_CountBitsSimple(n uint32) int32 { return int32(bits.OnesCount32(n)) }
func Jf_CountBits32(n uint32) int32     { return int32(bits.OnesCount32(n)) }
func Jf_CountBits(n word) int32         { return int32(bits.OnesCount64(uint64(n))) }
func Jf_CutGetSign32(c []int32) uint32 {
	var s uint32
	for i := int32(1); i <= Jf_CutSize(c); i++ {
		s |= 1 << (Jf_CutVar(c, i) & 31)
	}
	return s
}
func Jf_CutGetSign(c []int32) word {
	var s word
	for i := int32(1); i <= Jf_CutSize(c); i++ {
		s |= 1 << (Jf_CutVar(c, i) & 63)
	}
	return s
}

func Jf_CutIsContainedOrder(base, cut []int32) bool {
	n, m := Jf_CutSize(base), Jf_CutSize(cut)
	if m == 0 {
		return true
	}
	if n < m {
		return false
	}
	k := int32(1)
	for i := int32(1); i <= n; i++ {
		if base[i] > cut[k] {
			return false
		}
		if base[i] == cut[k] {
			if k == m {
				return true
			}
			k++
		}
	}
	return false
}
func Jf_CutMergeOrder(a, b, d []int32, limit int32) int32 {
	n, m := Jf_CutSize(a), Jf_CutSize(b)
	i, j, k := int32(1), int32(1), int32(0)
	for i <= n || j <= m {
		if k == limit {
			return 0
		}
		var v int32
		if j > m || (i <= n && a[i] < b[j]) {
			v = a[i]
			i++
		} else if i > n || b[j] < a[i] {
			v = b[j]
			j++
		} else {
			v = a[i]
			i++
			j++
		}
		k++
		d[k] = v
	}
	d[0] = k
	return 1
}
func Jf_CutFindLeaf0(c []int32, leaf int32) int32 {
	for i := int32(1); i <= Jf_CutSize(c); i++ {
		if c[i] == leaf {
			return i
		}
	}
	return Jf_CutSize(c) + 1
}
func Jf_CutFindLeaf1(c []int32, leaf int32) int32 {
	for i := int32(1); i <= Jf_CutSize(c); i++ {
		if c[i]>>1 == leaf {
			return i
		}
	}
	return Jf_CutSize(c) + 1
}
func Jf_CutIsContained0(base, cut []int32) bool {
	for i := int32(1); i <= Jf_CutSize(cut); i++ {
		if Jf_CutFindLeaf0(base, cut[i]) > base[0] {
			return false
		}
	}
	return true
}
func Jf_CutIsContained1(base, cut []int32) bool {
	for i := int32(1); i <= Jf_CutSize(cut); i++ {
		if Jf_CutFindLeaf1(base, cut[i]>>1) > base[0] {
			return false
		}
	}
	return true
}
func jfCutMerge(a, b, d []int32, limit int32, variable bool) int32 {
	n, m := Jf_CutSize(a), Jf_CutSize(b)
	d[0] = n
	for i := int32(1); i <= m; i++ {
		place := Jf_CutFindLeaf0(a, b[i])
		if variable {
			place = Jf_CutFindLeaf1(a, b[i]>>1)
		}
		if place > n {
			if d[0] == limit {
				return 0
			}
			d[0]++
			d[d[0]] = b[i]
		}
	}
	copy(d[1:], a[1:n+1])
	return 1
}
func Jf_CutMerge0(a, b, d []int32, limit int32) int32 { return jfCutMerge(a, b, d, limit, false) }
func Jf_CutMerge1(a, b, d []int32, limit int32) int32 { return jfCutMerge(a, b, d, limit, true) }
func Jf_CutMerge2(a, b, d []int32, limit int32) int32 {
	config := int32(0x3ffff)
	n, m := Jf_CutSize(a), Jf_CutSize(b)
	d[0] = n
	for i := int32(1); i <= m; i++ {
		place := Jf_CutFindLeaf1(a, b[i]>>1)
		if place > n {
			if d[0] == limit {
				return 0
			}
			d[0]++
			place = d[0]
			d[place] = b[i]
		} else if a[place] != b[i] {
			config |= 1 << (place + 17)
		}
		config ^= ((i - 1) ^ 7) << (3 * (place - 1))
	}
	copy(d[1:], a[1:n+1])
	return config
}

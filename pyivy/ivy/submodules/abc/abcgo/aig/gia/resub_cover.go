package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"sort"
)

func resubCount(words []word) int32 {
	var n int32
	for _, w := range words {
		n += int32(bits.OnesCount64(w))
	}
	return n
}
func Gia_ManDivCover(off, on, a []word, complA int32, b []word, complB, nWords int32) int32 {
	_ = off
	for i := 0; i < int(nWords); i++ {
		x, y := a[i], b[i]
		if complA != 0 {
			x = ^x
		}
		if complB != 0 {
			y = ^y
		}
		if on[i]&^x&^y != 0 {
			return 0
		}
	}
	return 1
}
func Gia_ManFindTwoUnateInt(off, on []word, divs *Vec_Ptr_t, nWords int32, lits, weights *Vec_Int_t, pairs *int32) int32 {
	values, covers := vec.Vec_IntArray(lits), vec.Vec_IntArray(weights)
	if len(values) != len(covers) {
		panic("gia: mismatched unate weights")
	}
	total := resubCount(on)
	*pairs = 0
	for i, l0 := range values {
		if 2*covers[i] < total {
			break
		}
		for j := i + 1; j < len(values); j++ {
			if covers[i]+covers[j] < total {
				break
			}
			a, b := min(l0, values[j]), max(l0, values[j])
			*pairs++
			if Gia_ManDivCover(off, on, resubDiv(divs, b>>1, nWords), b&1, resubDiv(divs, a>>1, nWords), a&1, nWords) != 0 {
				return 2*(((b^1)<<15)|(a^1)) + 1
			}
		}
	}
	return -1
}
func Gia_ManFindTwoUnate(sets [2][]word, divs *Vec_Ptr_t, nWords int32, lits, weights [2]*Vec_Int_t, verbose int32) int32 {
	if verbose != 0 {
		fmt.Print("  ")
	}
	for n := 0; n < 2; n++ {
		var count int32
		result := Gia_ManFindTwoUnateInt(sets[n], sets[n^1], divs, nWords, lits[n], weights[n], &count)
		if verbose != 0 {
			fmt.Printf("UU%d =%5d ", n, count)
		}
		if result >= 0 {
			return result ^ int32(n)
		}
	}
	return -1
}
func Gia_ManFindDivGateInt(off, on []word, divs *Vec_Ptr_t, nWords int32, lits, pairs, litWeights, pairWeights *Vec_Int_t, temp []word) int32 {
	ls, ps, lw, pw := vec.Vec_IntArray(lits), vec.Vec_IntArray(pairs), vec.Vec_IntArray(litWeights), vec.Vec_IntArray(pairWeights)
	if len(ls) != len(lw) || len(ps) != len(pw) {
		panic("gia: mismatched divisor weights")
	}
	total := resubCount(on)
	for i, lit := range ls {
		if 2*lw[i] < total {
			break
		}
		a := resubDiv(divs, lit>>1, nWords)
		for j, pair := range ps {
			if lw[i]+pw[j] < total {
				break
			}
			Gia_ManDeriveDivPair(pair, divs, nWords, temp)
			if Gia_ManDivCover(off, on, a, lit&1, temp, pair&1, nWords) != 0 {
				return 2*((int32(2*j+1)<<15)|(lit^1)) + 1
			}
		}
	}
	return -1
}
func Gia_ManFindDivGate(sets [2][]word, divs *Vec_Ptr_t, nWords int32, lits, pairs, litWeights, pairWeights [2]*Vec_Int_t, temp []word) int32 {
	for n := 0; n < 2; n++ {
		result := Gia_ManFindDivGateInt(sets[n], sets[n^1], divs, nWords, lits[n], pairs[n], litWeights[n], pairWeights[n], temp)
		if result >= 0 {
			return result ^ int32(n)
		}
	}
	return -1
}
func Gia_ManFindGateGateInt(off, on []word, divs *Vec_Ptr_t, nWords int32, pairs, weights *Vec_Int_t, tempA, tempB []word) int32 {
	ps, ws := vec.Vec_IntArray(pairs), vec.Vec_IntArray(weights)
	if len(ps) != len(ws) {
		panic("gia: mismatched pair weights")
	}
	total := resubCount(on)
	for i, a := range ps {
		if 2*ws[i] < total {
			break
		}
		Gia_ManDeriveDivPair(a, divs, nWords, tempA)
		for j := i + 1; j < len(ps); j++ {
			if ws[i]+ws[j] < total {
				break
			}
			b := ps[j]
			Gia_ManDeriveDivPair(b, divs, nWords, tempB)
			if Gia_ManDivCover(off, on, tempA, a&1, tempB, b&1, nWords) != 0 {
				return 2*((int32(2*j+1)<<15)|int32(2*i+1)) + 1
			}
		}
	}
	return -1
}
func Gia_ManFindGateGate(sets [2][]word, divs *Vec_Ptr_t, nWords int32, pairs, weights [2]*Vec_Int_t, tempA, tempB []word) int32 {
	for n := 0; n < 2; n++ {
		result := Gia_ManFindGateGateInt(sets[n], sets[n^1], divs, nWords, pairs[n], weights[n], tempA, tempB)
		if result >= 0 {
			return result ^ int32(n)
		}
	}
	return -1
}

type resubWeighted struct{ lit, weight int32 }

func resubSortWeighted(values, weights *Vec_Int_t, compute func(int32) int32, sorter *Vec_Wec_t) {
	items := make([]resubWeighted, 0, vec.Vec_IntSize(values))
	for _, lit := range vec.Vec_IntArray(values) {
		items = append(items, resubWeighted{lit, compute(lit)})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].weight > items[j].weight })
	vec.Vec_IntClear(values)
	vec.Vec_IntClear(weights)
	for _, item := range items {
		vec.Vec_IntPush(values, item.lit)
		vec.Vec_IntPush(weights, item.weight)
	}
	vec.Vec_WecClear(sorter)
}
func Gia_ManSortUnatesInt(off, on []word, divs *Vec_Ptr_t, nWords int32, lits, weights *Vec_Int_t, sorter *Vec_Wec_t) {
	_ = off
	resubSortWeighted(lits, weights, func(lit int32) int32 {
		d := resubDiv(divs, lit>>1, nWords)
		var n int32
		for i, v := range d {
			if lit&1 != 0 {
				v = ^v
			}
			n += int32(bits.OnesCount64(v & on[i]))
		}
		return n
	}, sorter)
}
func Gia_ManSortUnates(sets [2][]word, divs *Vec_Ptr_t, nWords int32, lits, weights [2]*Vec_Int_t, sorter *Vec_Wec_t) {
	for n := 0; n < 2; n++ {
		Gia_ManSortUnatesInt(sets[n], sets[n^1], divs, nWords, lits[n], weights[n], sorter)
	}
}
func Gia_ManSortPairsInt(off, on []word, divs *Vec_Ptr_t, nWords int32, pairs, weights *Vec_Int_t, sorter *Vec_Wec_t) {
	_ = off
	temp := make([]word, nWords)
	resubSortWeighted(pairs, weights, func(pair int32) int32 {
		Gia_ManDeriveDivPair(pair, divs, nWords, temp)
		var n int32
		for i, v := range temp {
			if pair&1 != 0 {
				v = ^v
			}
			n += int32(bits.OnesCount64(v & on[i]))
		}
		return n
	}, sorter)
}
func Gia_ManSortPairs(sets [2][]word, divs *Vec_Ptr_t, nWords int32, pairs, weights [2]*Vec_Int_t, sorter *Vec_Wec_t) {
	for n := 0; n < 2; n++ {
		Gia_ManSortPairsInt(sets[n], sets[n^1], divs, nWords, pairs[n], weights[n], sorter)
	}
}
func Gia_ManSortBinate(sets [2][]word, divs *Vec_Ptr_t, nWords int32, vars *Vec_Int_t, sorter *Vec_Wec_t) {
	big, small := sets[1], sets[0]
	if resubCount(sets[0]) > resubCount(sets[1]) {
		big, small = sets[0], sets[1]
	}
	nBig, nSmall := resubCount(big), resubCount(small)
	weights := vec.Vec_IntAlloc(0)
	resubSortWeighted(vars, weights, func(id int32) int32 {
		d := resubDiv(divs, id, nWords)
		var a, b int32
		for i, v := range d {
			a += int32(bits.OnesCount64(big[i] & v))
			b += int32(bits.OnesCount64(small[i] & v))
		}
		if a < nBig/2 {
			a = nBig - a
			b = nSmall - b
		}
		return max(0, a-nBig/2+nSmall/2-b)
	}, sorter)
	if vec.Vec_IntSize(vars) > 2000 {
		vec.Vec_IntShrink(vars, 2000)
	}
}

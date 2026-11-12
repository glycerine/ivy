package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func resubDiv(divs *Vec_Ptr_t, index int32, nWords int32) []word {
	data, ok := vec.Vec_PtrEntry(divs, int(index)).([]word)
	if !ok || nWords < 0 || len(data) < int(nWords) {
		panic("gia: invalid resubstitution divisor")
	}
	return data[:nWords]
}
func resubIntersects(a, b []word, complement bool) bool {
	if len(a) != len(b) {
		panic("gia: mismatched resubstitution truth sizes")
	}
	for i, v := range a {
		x := b[i]
		if complement {
			x = ^x
		}
		if v&x != 0 {
			return true
		}
	}
	return false
}
func Gia_ManFindFirstCommonLit(a, b *Vec_Int_t, verbose int32) int32 {
	_ = verbose
	x, y := vec.Vec_IntArray(a), vec.Vec_IntArray(b)
	i, j, wi, wj := 0, 0, 0, 0
	for i < len(x) && j < len(y) {
		if x[i]>>1 == y[j]>>1 {
			if x[i] != y[j] {
				return x[i]
			}
			i++
			j++
		} else if x[i] < y[j] {
			x[wi] = x[i]
			wi++
			i++
		} else {
			y[wj] = y[j]
			wj++
			j++
		}
	}
	for i < len(x) {
		x[wi] = x[i]
		wi++
		i++
	}
	for j < len(y) {
		y[wj] = y[j]
		wj++
		j++
	}
	vec.Vec_IntShrink(a, wi)
	vec.Vec_IntShrink(b, wj)
	return -1
}
func Gia_ManFindOneUnateInt(off, on []word, divs *Vec_Ptr_t, nWords int32, unate, notUnate *Vec_Int_t) {
	_ = on
	if len(off) != int(nWords) {
		panic("gia: invalid resubstitution truth size")
	}
	vec.Vec_IntClear(unate)
	vec.Vec_IntClear(notUnate)
	for i := 2; i < vec.Vec_PtrSize(divs); i++ {
		d := resubDiv(divs, int32(i), nWords)
		if !resubIntersects(off, d, false) {
			vec.Vec_IntPush(unate, int32(2*i))
		} else if !resubIntersects(off, d, true) {
			vec.Vec_IntPush(unate, int32(2*i+1))
		} else {
			vec.Vec_IntPush(notUnate, int32(i))
		}
	}
}
func Gia_ManFindOneUnate(sets [2][]word, divs *Vec_Ptr_t, nWords int32, unate, notUnate [2]*Vec_Int_t, verbose int32) int32 {
	if verbose != 0 {
		fmt.Print("  ")
	}
	for n := 0; n < 2; n++ {
		Gia_ManFindOneUnateInt(sets[n], sets[n^1], divs, nWords, unate[n], notUnate[n])
		if verbose != 0 {
			fmt.Printf("U%d =%4d ", n, vec.Vec_IntSize(unate[n]))
		}
	}
	return Gia_ManFindFirstCommonLit(unate[0], unate[1], verbose)
}
func Gia_ManFindXorInt(off, on []word, binate *Vec_Int_t, divs *Vec_Ptr_t, nWords int32, pairs *Vec_Int_t) {
	_ = on
	vars := vec.Vec_IntArray(binate)
	for i, v1 := range vars {
		for _, v0 := range vars[:i] {
			a, b := min(v0, v1), max(v0, v1)
			x, y := resubDiv(divs, a, nWords), resubDiv(divs, b, nWords)
			positive, negative := false, false
			for k := 0; k < int(nWords); k++ {
				z := x[k] ^ y[k]
				positive = positive || off[k]&z != 0
				negative = negative || off[k]&^z != 0
			}
			packed := ((2 * a) << 15) | (2 * b)
			if !positive {
				vec.Vec_IntPush(pairs, 2*packed)
			} else if !negative {
				vec.Vec_IntPush(pairs, 2*packed+1)
			}
		}
	}
}
func Gia_ManFindXor(sets [2][]word, divs *Vec_Ptr_t, nWords int32, binate *Vec_Int_t, pairs [2]*Vec_Int_t, verbose int32) int32 {
	if verbose != 0 {
		fmt.Print("  ")
	}
	for n := 0; n < 2; n++ {
		vec.Vec_IntClear(pairs[n])
		Gia_ManFindXorInt(sets[n], sets[n^1], binate, divs, nWords, pairs[n])
		if verbose != 0 {
			fmt.Printf("UX%d =%5d ", n, vec.Vec_IntSize(pairs[n]))
		}
	}
	return Gia_ManFindFirstCommonLit(pairs[0], pairs[1], verbose)
}
func Gia_ManFindUnatePairsInt(off, on []word, binate *Vec_Int_t, divs *Vec_Ptr_t, nWords int32, pairs *Vec_Int_t) {
	vars := vec.Vec_IntArray(binate)
	for i, v1 := range vars {
		for _, v0 := range vars[:i] {
			a, b := min(v0, v1), max(v0, v1)
			x, y := resubDiv(divs, a, nWords), resubDiv(divs, b, nWords)
			for phase := int32(0); phase < 4; phase++ {
				intersectsOff, intersectsOn := false, false
				for k := 0; k < int(nWords); k++ {
					u, v := x[k], y[k]
					if phase&1 != 0 {
						u = ^u
					}
					if phase&2 != 0 {
						v = ^v
					}
					z := u & v
					intersectsOff = intersectsOff || off[k]&z != 0
					intersectsOn = intersectsOn || on[k]&z != 0
				}
				if !intersectsOff && intersectsOn {
					l0, l1 := 2*a+(phase&1), 2*b+(phase>>1)
					vec.Vec_IntPush(pairs, 2*((l1<<15)|l0))
				}
			}
		}
	}
}
func Gia_ManFindUnatePairs(sets [2][]word, divs *Vec_Ptr_t, nWords int32, binate *Vec_Int_t, pairs [2]*Vec_Int_t, verbose int32) {
	for n := 0; n < 2; n++ {
		before := vec.Vec_IntSize(pairs[n])
		Gia_ManFindUnatePairsInt(sets[n], sets[n^1], binate, divs, nWords, pairs[n])
		if verbose != 0 {
			fmt.Printf("UP%d =%5d ", n, vec.Vec_IntSize(pairs[n])-before)
		}
	}
	if Gia_ManFindFirstCommonLit(pairs[0], pairs[1], verbose) != -1 {
		panic("gia: inconsistent unate pair")
	}
}
func Gia_ManDeriveDivPair(pair int32, divs *Vec_Ptr_t, nWords int32, result []word) {
	if len(result) < int(nWords) {
		panic("gia: short divisor result")
	}
	packed := pair >> 1
	a, b := packed&0x7fff, packed>>15
	x, y := resubDiv(divs, a>>1, nWords), resubDiv(divs, b>>1, nWords)
	if a < b {
		if pair&1 != 0 {
			panic("gia: invalid AND pair")
		}
		for i := 0; i < int(nWords); i++ {
			u, v := x[i], y[i]
			if a&1 != 0 {
				u = ^u
			}
			if b&1 != 0 {
				v = ^v
			}
			result[i] = u & v
		}
	} else {
		if a&1 != 0 || b&1 != 0 {
			panic("gia: invalid XOR pair")
		}
		for i := 0; i < int(nWords); i++ {
			result[i] = x[i] ^ y[i]
		}
	}
}

package gia

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManSimBitPackOne merges an assignment into the first compatible slot.
func Gia_ManSimBitPackOne(words int32, values, care *Vec_Wrd_t, patterns int32, lits []int32) int32 {
	if words <= 0 || patterns < 0 || patterns >= 64*words || vec.Vec_WrdSize(values) != vec.Vec_WrdSize(care) {
		panic("gia: invalid pattern packing dimensions")
	}
	data, mask := vec.Vec_WrdArray(values), vec.Vec_WrdArray(care)
	bit := int32(0)
	for ; bit < patterns; bit++ {
		compatible := true
		for _, lit := range lits {
			id := lit >> 1
			start := int(id * words)
			if start < 0 || start+int(words) > len(data) {
				panic("gia: invalid pattern variable")
			}
			if giaSimBit(mask[start:start+int(words)], bit) != 0 && giaSimBit(data[start:start+int(words)], bit) == uint64(lit&1) {
				compatible = false
				break
			}
		}
		if compatible {
			break
		}
	}
	for _, lit := range lits {
		start := int((lit >> 1) * words)
		v, c := data[start:start+int(words)], mask[start:start+int(words)]
		if giaSimBit(c, bit) == 0 && giaSimBit(v, bit) == uint64(lit&1) {
			giaSimFlip(v, bit)
		}
		c[bit>>6] |= uint64(1) << uint(bit&63)
		if giaSimBit(v, bit) == uint64(lit&1) {
			panic("gia: inconsistent pattern assignment")
		}
	}
	return giaBool(bit == patterns)
}
func Gia_ManSimBitPacking(p *Gia_Man_t, store *Vec_Int_t, cexes, undecided int32) *Vec_Wrd_t {
	if cexes < 0 || undecided < 0 || Gia_ManCiNum(p) <= 0 {
		panic("gia: invalid counterexample count")
	}
	nWords := util.Abc_Bit6WordNum(cexes + 1)
	values := vec.Vec_WrdStartRandom(int(Gia_ManCiNum(p) * nWords))
	care := vec.Vec_WrdStart(int(Gia_ManCiNum(p) * nWords))
	entries := vec.Vec_IntArray(store)
	cursor, patterns := 0, int32(0)
	for c := int32(0); c < cexes+undecided; c++ {
		if cursor+2 > len(entries) {
			panic("gia: truncated counterexample store")
		}
		size := entries[cursor+1]
		cursor += 2
		if size == -1 {
			continue
		}
		if size < 0 || cursor+int(size) > len(entries) {
			panic("gia: invalid counterexample size")
		}
		patterns += Gia_ManSimBitPackOne(nWords, values, care, patterns, entries[cursor:cursor+int(size)])
		cursor += int(size)
	}
	if cursor != len(entries) {
		panic("gia: trailing counterexample data")
	}
	use := util.Abc_Bit6WordNum(patterns + 1)
	result := Gia_ManSimCombine(Gia_ManCiNum(p), p.VSimsPi, values, use)
	fmt.Printf("Compressed %d CEXes into %d patterns and added %d words to available %d words.\n", cexes, patterns, use, vec.Vec_WrdSize(p.VSimsPi)/int(Gia_ManCiNum(p)))
	return result
}

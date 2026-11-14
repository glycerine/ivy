package if_

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManFromIfLogicCreateLut builds a truth function and records its mapped
// leaves when the result is an internal node. Constant and leaf results do not
// require a separate LUT record. cover is reserved for factoring scratch.
func Gia_ManFromIfLogicCreateLut(dst *Gia_Man_t, truth []word, leaves, cover, mapping, records *Vec_Int_t) int32 {
	n := Vec_IntSize(leaves)
	needed := (1 << uint(n))
	if n < 0 || n > IF_MAX_FUNC_LUTSIZE || len(truth)*64 < needed {
		panic("if: invalid LUT truth table")
	}
	words := make([]uint32, len(truth)*2)
	for i, value := range truth {
		words[2*i], words[2*i+1] = uint32(value), uint32(value>>32)
	}
	var scratch []uint32
	lit := kit.Kit_TruthToGia(dst, words, int32(n), &scratch, Vec_IntArray(leaves), 0)
	id := lit >> 1
	if id == 0 {
		return lit
	}
	for _, leaf := range Vec_IntArray(leaves) {
		if leaf>>1 == id {
			return lit
		}
	}
	vec.Vec_IntSetEntry(mapping, int(id), int32(Vec_IntSize(records)))
	Vec_IntPush(records, int32(n))
	for _, leaf := range Vec_IntArray(leaves) {
		Vec_IntPush(records, leaf>>1)
	}
	Vec_IntPush(records, id)
	return lit
}

// Gia_ManFromIfLogicCreateLutSpecial reconstructs the pair of four-input LUTs
// selected by the 4-4 decomposition and records their packing group.
func Gia_ManFromIfLogicCreateLutSpecial(dst *Gia_Man_t, truth []word, leaves, temp, cover, mapping, records, packing *Vec_Int_t) int32 {
	n := Vec_IntSize(leaves)
	if n < 5 || n > 7 || len(truth) < (1<<uint(n)+63)/64 || Vec_IntSize(packing) == 0 {
		panic("if: invalid 4-4 decomposition arguments")
	}
	truth32 := make([]uint32, len(truth)*2)
	for i, value := range truth {
		truth32[2*i], truth32[2*i+1] = uint32(value), uint32(value>>32)
	}
	z := If_CutPerformDerive07(nil, truth32, int32(n), int32(n), "")
	if z == 0 {
		panic("if: function has no 4-4 decomposition")
	}
	build := func(shift uint, bound int32) int32 {
		vec.Vec_IntClear(temp)
		for i := uint(0); i < 4; i++ {
			pin := int((z >> (shift + 16 + 4*i)) & 7)
			if pin == 6 && n == 5 {
				continue
			}
			if pin == 7 {
				if shift == 0 {
					panic("if: invalid bound LUT pin")
				}
				Vec_IntPush(temp, bound)
			} else {
				if pin >= n {
					panic("if: invalid LUT decomposition pin")
				}
				Vec_IntPush(temp, Vec_IntEntry(leaves, pin))
			}
		}
		part := word((z >> shift) & 0xffff)
		part |= part << 16
		part |= part << 32
		return Gia_ManFromIfLogicCreateLut(dst, []word{part}, temp, cover, mapping, records)
	}
	bound := build(0, 0)
	result := build(32, bound)
	Vec_IntPush(packing, 2)
	Vec_IntPush(packing, bound>>1)
	Vec_IntPush(packing, result>>1)
	vec.Vec_IntAddToEntry(packing, 0, 1)
	return result
}

// Gia_ManFromIfLogicFindLut maps a cut to one LUT or to a pair of LUTs using
// the DSD-selected bound set and the exact two-level decomposition checker.
func Gia_ManFromIfLogicFindLut(p *If_Man_t, dst *Gia_Man_t, cut *If_Cut_t, sat any, leaves, scratch, cover, mapping, records, packing *Vec_Int_t) int32 {
	if len(p.pPars.PLutStruct) == 0 || Vec_IntSize(packing) == 0 {
		panic("if: missing LUT structure or packing header")
	}
	size := int32(p.pPars.PLutStruct[0] - '0')
	if size < 2 || size > 6 || Vec_IntSize(leaves) != int(cut.nLeaves) {
		panic("if: invalid LUT structure or cut leaves")
	}
	source := If_CutTruthW(p, cut)
	if source == nil {
		panic("if: cut truth is unavailable")
	}
	if Vec_IntSize(leaves) <= int(size) {
		lit := Gia_ManFromIfLogicCreateLut(dst, []word{*source}, leaves, cover, mapping, records)
		if lit > 1 && gia.Gia_ObjIsCi(Gia_ManObj(dst, lit>>1)) == 0 {
			Vec_IntPush(packing, 1)
			Vec_IntPush(packing, lit>>1)
			vec.Vec_IntAddToEntry(packing, 0, 1)
		}
		return lit
	}
	assert(If_DsdManSuppSize(p.pIfDsdMan, If_CutDsdLit(p, cut)) == int32(cut.nLeaves), "DSD support matches cut")
	oldSet := uint32(cut.uMaskFunc)
	if p.pPars.FDelayOptLut == 0 {
		oldSet = If_DsdManCheckXY(p.pIfDsdMan, If_CutDsdLit(p, cut), size, 1, 0, 1, 0)
	}
	var set uint32
	perm := If_CutDsdPerm(p, cut)
	for k := int32(0); k < int32(cut.nLeaves); k++ {
		pin := int32(perm[k] >> 1)
		class := (oldSet >> uint(2*k)) & 3
		if pin >= int32(cut.nLeaves) || class == 2 {
			panic("if: invalid DSD cut partition")
		}
		set |= class << uint(2*pin)
	}
	var boundTruth, freeTruth word
	if ifManSatCheckXY(sat, size, p.puTempW[:p.nTruth6Words[cut.nLeaves]], int32(cut.nLeaves), set, &boundTruth, &freeTruth, scratch) == 0 {
		panic("if: cannot derive selected two-LUT partition")
	}
	var bound, shared, free []int
	for k := 0; k < int(cut.nLeaves); k++ {
		switch (set >> uint(2*k)) & 3 {
		case 0:
			free = append(free, k)
		case 1:
			bound = append(bound, k)
		case 3:
			shared = append(shared, k)
		default:
			panic("if: invalid cut partition")
		}
	}
	vec.Vec_IntClear(scratch)
	for _, pin := range shared {
		Vec_IntPush(scratch, Vec_IntEntry(leaves, pin))
	}
	for _, pin := range bound {
		Vec_IntPush(scratch, Vec_IntEntry(leaves, pin))
	}
	first := Gia_ManFromIfLogicCreateLut(dst, []word{boundTruth}, scratch, cover, mapping, records)
	vec.Vec_IntClear(scratch)
	Vec_IntPush(scratch, first)
	for _, pin := range shared {
		Vec_IntPush(scratch, Vec_IntEntry(leaves, pin))
	}
	for _, pin := range free {
		Vec_IntPush(scratch, Vec_IntEntry(leaves, pin))
	}
	result := Gia_ManFromIfLogicCreateLut(dst, []word{freeTruth}, scratch, cover, mapping, records)
	Vec_IntPush(packing, 2)
	Vec_IntPush(packing, first>>1)
	Vec_IntPush(packing, result>>1)
	vec.Vec_IntAddToEntry(packing, 0, 1)
	return result
}

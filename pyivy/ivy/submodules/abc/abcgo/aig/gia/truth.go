package gia

import "math/bits"

func Gia_ObjCutSign(id uint32) uint32 { return 1 << (id & 31) }

// Preserve ABC's predicate, which also accepts zero (at most one set bit).
func Gia_WordHasOneBit(word uint32) int32  { return giaBool(word&(word-1) == 0) }
func Gia_WordHasOnePair(word uint32) int32 { return Gia_WordHasOneBit(word & (word >> 1) & 0x55555555) }
func Gia_WordCountOnes(word uint32) int32  { return int32(bits.OnesCount32(word)) }
func Gia_WordFindFirstBit(word uint32) int32 {
	if word == 0 {
		return -1
	}
	return int32(bits.TrailingZeros32(word))
}
func giaTruthWords(nVars int32) int {
	if nVars < 0 || nVars > 30 {
		panic("gia: unsupported truth variable count")
	}
	if nVars <= 5 {
		return 1
	}
	return 1 << uint(nVars-5)
}
func Gia_ManTruthIsConst0(in []uint32, nVars int32) int32 {
	for _, word := range in[:giaTruthWords(nVars)] {
		if word != 0 {
			return 0
		}
	}
	return 1
}
func Gia_ManTruthIsConst1(in []uint32, nVars int32) int32 {
	for _, word := range in[:giaTruthWords(nVars)] {
		if word != ^uint32(0) {
			return 0
		}
	}
	return 1
}
func Gia_ManTruthCopy(out, in []uint32, nVars int32) {
	n := giaTruthWords(nVars)
	copy(out[:n], in[:n])
}
func Gia_ManTruthClear(out []uint32, nVars int32) { clear(out[:giaTruthWords(nVars)]) }
func Gia_ManTruthFill(out []uint32, nVars int32) {
	for i := range out[:giaTruthWords(nVars)] {
		out[i] = ^uint32(0)
	}
}
func Gia_ManTruthNot(out, in []uint32, nVars int32) {
	for i, word := range in[:giaTruthWords(nVars)] {
		out[i] = ^word
	}
}

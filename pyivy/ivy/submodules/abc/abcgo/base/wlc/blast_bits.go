package wlc

import "github.com/glycerine/abcgo/misc/vec"

func Wlc_NtkCountConstBits(bits []int32, nSize int32) int32 {
	var count int32
	for _, bit := range bits[:nSize] {
		if bit == 0 || bit == 1 {
			count++
		}
	}
	return count
}
func Wlc_VecCopy(out *vec.Vec_Int_t, bits []int32, nSize int32) []int32 {
	vec.Vec_IntClear(out)
	for _, bit := range bits[:nSize] {
		vec.Vec_IntPush(out, bit)
	}
	return vec.Vec_IntArray(out)
}
func Wlc_VecLoadFanins(out *vec.Vec_Int_t, fanins []int32, nFanins, nTotal, signed int32) []int32 {
	fill := int32(0)
	if signed != 0 {
		fill = fanins[nFanins-1]
	}
	vec.Vec_IntClear(out)
	if nFanins > nTotal {
		panic("fanin count exceeds width")
	}
	for i := int32(0); i < nTotal; i++ {
		bit := fill
		if i < nFanins {
			bit = fanins[i]
		}
		vec.Vec_IntPush(out, bit)
	}
	return vec.Vec_IntArray(out)
}
func Wlc_BlastGetConst(bits []int32, nBits int32) int32 {
	var result int32
	for i, bit := range bits[:nBits] {
		if bit == 1 {
			result |= int32(1) << i
		} else if bit != 0 {
			return -1
		}
	}
	return result
}

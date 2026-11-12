package gia

import "math/bits"

const (
	MF_LEAF_MAX = 10
	MF_NO_FUNC  = 134217727
)

type Mf_Cut_t struct {
	Sign    word
	Delay   int32
	Flow    float32
	iFunc   uint32
	nLeaves uint32
	pLeaves [MF_LEAF_MAX + 1]int32
}

func Mf_CutCountBits(v word) int32 { return int32(bits.OnesCount64(v)) }
func Mf_CutGetSign(leaves []int32, n int32) word {
	var sign word
	for _, leaf := range leaves[:n] {
		sign |= word(1) << (uint32(leaf) & 63)
	}
	return sign
}
func Mf_CutCreateUnit(c *Mf_Cut_t, id int32) int32 {
	c.Delay, c.Flow, c.iFunc, c.nLeaves = 0, 0, 2, 1
	c.pLeaves[0], c.Sign = id, word(1)<<(uint32(id)&63)
	return 1
}

// Mf_CutMergeOrder and Mf_CutMergeOrderMux merge sorted leaf sets.
func Mf_CutMergeOrder(a, b, out *Mf_Cut_t, limit int32) int32 {
	return mfMergeCuts(out, limit, a, b)
}
func Mf_CutMergeOrderMux(a, b, c, out *Mf_Cut_t, limit int32) int32 {
	return mfMergeCuts(out, limit, a, b, c)
}
func mfMergeCuts(out *Mf_Cut_t, limit int32, cuts ...*Mf_Cut_t) int32 {
	var positions [3]uint32
	var sign word
	for _, c := range cuts {
		sign |= c.Sign
	}
	count := int32(0)
	for {
		found := false
		var least int32
		for i, c := range cuts {
			if positions[i] < c.nLeaves {
				v := c.pLeaves[positions[i]]
				if !found || v < least {
					least, found = v, true
				}
			}
		}
		if !found {
			break
		}
		if count == limit {
			return 0
		}
		out.pLeaves[count] = least
		count++
		for i, c := range cuts {
			if positions[i] < c.nLeaves && c.pLeaves[positions[i]] == least {
				positions[i]++
			}
		}
	}
	out.nLeaves, out.iFunc, out.Sign = uint32(count), MF_NO_FUNC, sign
	return 1
}
func Mf_CutCheck(base, cut *Mf_Cut_t) int32 {
	for _, v := range cut.pLeaves[:cut.nLeaves] {
		found := false
		for _, b := range base.pLeaves[:base.nLeaves] {
			if b == v {
				found = true
				break
			}
		}
		if !found {
			return 0
		}
	}
	return 1
}
func Mf_SetCutIsContainedOrder(base, cut *Mf_Cut_t) int32 {
	i := 0
	for _, b := range base.pLeaves[:base.nLeaves] {
		if i == int(cut.nLeaves) {
			break
		}
		if b > cut.pLeaves[i] {
			return 0
		}
		if b == cut.pLeaves[i] {
			i++
		}
	}
	if i == int(cut.nLeaves) {
		return 1
	}
	return 0
}

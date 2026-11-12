package gia

import "math/bits"

const (
	NF_LEAF_MAX = 6
	NF_CUT_MAX  = 32
	NF_NO_LEAF  = 31
	NF_NO_FUNC  = 0x3ffffff
	NF_EPSILON  = 0.001
)

type Nf_Cut_t struct {
	Sign                    uint64
	Delay                   int32
	Flow                    float32
	iFunc, Useless, nLeaves uint32
	pLeaves                 [NF_LEAF_MAX + 1]int32
}
type Nf_Cfg_t struct{ fCompl, Phase, Perm uint32 }

// Nf_Cfg2Int uses the bit layout of the original packed C configuration.
func Nf_Cfg2Int(c Nf_Cfg_t) int32 {
	return int32((c.fCompl & 1) | ((c.Phase & 127) << 1) | ((c.Perm & 0xffffff) << 8))
}
func Nf_Int2Cfg(v int32) Nf_Cfg_t {
	u := uint32(v)
	return Nf_Cfg_t{fCompl: u & 1, Phase: (u >> 1) & 127, Perm: u >> 8}
}
func Nf_CfgVar(c Nf_Cfg_t, i int32) int32   { return int32((c.Perm >> uint(i*4)) & 15) }
func Nf_CfgCompl(c Nf_Cfg_t, i int32) int32 { return int32((c.Phase >> uint(i)) & 1) }
func Nf_CutCountBits(v uint64) int32        { return int32(bits.OnesCount64(v)) }
func Nf_CutGetSign(leaves []int32) uint64 {
	var sign uint64
	for _, id := range leaves {
		sign |= uint64(1) << uint(id&63)
	}
	return sign
}
func Nf_CutCreateUnit(c *Nf_Cut_t, id int32) int32 {
	*c = Nf_Cut_t{Sign: uint64(1) << uint(id&63), iFunc: 2, nLeaves: 1}
	c.pLeaves[0] = id
	return 1
}

// Nf_CutMergeOrder merges sorted cut leaves; it leaves dst unchanged on overflow.
func Nf_CutMergeOrder(a, b, dst *Nf_Cut_t, limit int32) int32 { return nfMergeCuts(dst, limit, a, b) }
func Nf_CutMergeOrderMux(a, b, c, dst *Nf_Cut_t, limit int32) int32 {
	return nfMergeCuts(dst, limit, a, b, c)
}
func nfMergeCuts(dst *Nf_Cut_t, limit int32, cuts ...*Nf_Cut_t) int32 {
	if limit < 0 || limit > NF_LEAF_MAX {
		panic("gia: invalid Nf cut limit")
	}
	var leaves [NF_LEAF_MAX + 1]int32
	var positions [3]int
	count := 0
	var sign uint64
	for _, c := range cuts {
		if c.nLeaves > NF_LEAF_MAX {
			panic("gia: invalid Nf cut")
		}
		sign |= c.Sign
	}
	for {
		value := int32(1<<31 - 1)
		found := false
		for i, c := range cuts {
			if positions[i] < int(c.nLeaves) && (!found || c.pLeaves[positions[i]] < value) {
				value = c.pLeaves[positions[i]]
				found = true
			}
		}
		if !found {
			break
		}
		if count >= int(limit) {
			return 0
		}
		leaves[count] = value
		count++
		for i, c := range cuts {
			if positions[i] < int(c.nLeaves) && c.pLeaves[positions[i]] == value {
				positions[i]++
			}
		}
	}
	dst.pLeaves = leaves
	dst.nLeaves = uint32(count)
	dst.iFunc = NF_NO_FUNC
	dst.Sign = sign
	return 1
}
func Nf_SetCutIsContainedOrder(base, cut *Nf_Cut_t) int32 {
	if cut.nLeaves > base.nLeaves {
		return 0
	}
	j := 0
	for i := 0; i < int(base.nLeaves) && j < int(cut.nLeaves); i++ {
		if base.pLeaves[i] == cut.pLeaves[j] {
			j++
		} else if base.pLeaves[i] > cut.pLeaves[j] {
			return 0
		}
	}
	return giaBool(j == int(cut.nLeaves))
}
func Nf_CutCompareArea(a, b *Nf_Cut_t) int32 {
	if a.Useless < b.Useless {
		return -1
	}
	if a.Useless > b.Useless {
		return 1
	}
	if a.Flow < b.Flow-NF_EPSILON {
		return -1
	}
	if a.Flow > b.Flow+NF_EPSILON {
		return 1
	}
	if a.Delay < b.Delay {
		return -1
	}
	if a.Delay > b.Delay {
		return 1
	}
	if a.nLeaves < b.nLeaves {
		return -1
	}
	if a.nLeaves > b.nLeaves {
		return 1
	}
	return 0
}

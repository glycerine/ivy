package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// nfTruthStore owns canonical six-input cut functions. matchCount is the
// number of functions for which a standard-cell implementation exists.
type nfTruthStore struct {
	memory     *vec.Vec_Mem_t
	matchCount int
}

func (s *nfTruthStore) compute(inputs []*Nf_Cut_t, complements []int32, out *Nf_Cut_t, op byte) int32 {
	if s == nil || s.memory == nil || len(inputs) != len(complements) || out.nLeaves > 6 {
		panic("gia: invalid Nf truth computation")
	}
	values := make([]uint64, len(inputs))
	for i, cut := range inputs {
		truth := vec.Vec_MemReadEntry(s.memory, int(cut.iFunc>>1))[0]
		if int32(cut.iFunc&1)^complements[i] != 0 {
			truth = ^truth
		}
		values[i] = util.Abc_Tt6Expand(truth, cut.pLeaves[:], int32(cut.nLeaves), out.pLeaves[:], int32(out.nLeaves))
	}
	var truth uint64
	switch op {
	case '&':
		truth = values[0] & values[1]
	case '^':
		truth = values[0] ^ values[1]
	case 'm':
		truth = (values[2] & values[1]) | (^values[2] & values[0])
	default:
		panic("gia: invalid Nf truth operation")
	}
	phase := uint32(truth & 1)
	if phase != 0 {
		truth = ^truth
	}
	old := out.nLeaves
	out.nLeaves = uint32(util.Abc_Tt6MinBase(&truth, out.pLeaves[:], int32(out.nLeaves)))
	id := vec.Vec_MemHashInsert(s.memory, []uint64{truth})
	out.iFunc = uint32(id<<1) | phase
	out.Useless = uint32(giaBool(id >= s.matchCount))
	return giaBool(out.nLeaves < old)
}
func Nf_CutComputeTruth6(s *nfTruthStore, a, b *Nf_Cut_t, ca, cb int32, out *Nf_Cut_t, xor int32) int32 {
	op := byte('&')
	if xor != 0 {
		op = '^'
	}
	return s.compute([]*Nf_Cut_t{a, b}, []int32{ca, cb}, out, op)
}
func Nf_CutComputeTruthMux6(s *nfTruthStore, a, b, c *Nf_Cut_t, ca, cb, cc int32, out *Nf_Cut_t) int32 {
	return s.compute([]*Nf_Cut_t{a, b, c}, []int32{ca, cb, cc}, out, 'm')
}

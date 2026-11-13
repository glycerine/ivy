package mpm

import (
	"math/bits"
	"slices"

	"github.com/glycerine/abcgo/misc/util"
)

// Ifd_ComputeSignature computes the sorted cofactor counts in mpmPre.c.
func Ifd_ComputeSignature(truth word, counts []int32) {
	for v := int32(0); v < 6; v++ {
		zero := util.Abc_Tt6Cofactor0(truth, v)
		one := util.Abc_Tt6Cofactor1(truth, v)
		neg, pos, xor := int32(bits.OnesCount64(zero)/2), int32(bits.OnesCount64(one)/2), int32(bits.OnesCount64(zero^one)/2)
		if pos > neg {
			pos, neg = neg, pos
		}
		counts[v] = pos<<20 | neg<<10 | xor
	}
	slices.Sort(counts[:6])
}

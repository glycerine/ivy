package bmc

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

// Maj3_ManValue gives the strict majority value of the selected input bits.
func Maj3_ManValue(minterm, nVars int32) int32 {
	if nVars < 0 || nVars > 31 {
		panic("bmc: invalid majority input count")
	}
	mask := uint32(1)<<uint(nVars) - 1
	if bits.OnesCount32(uint32(minterm)&mask) > int(nVars/2) {
		return 1
	}
	return 0
}
func Zyx_ManValue(minterm, nVars int32) int32 { return Maj3_ManValue(minterm, nVars) }

// Maj3_ManFirstAndLevel maps objects to levels, reversing the level-count
// order during the walk as in ABC, without modifying the caller's vector.
func Maj3_ManFirstAndLevel(levelCounts *vec.Vec_Int_t, firsts, levels []int32, nVars, nObjs int32) {
	counts := vec.Vec_IntArray(levelCounts)
	if nVars < 0 || nObjs < nVars || int(nObjs) > len(levels) || len(firsts) < len(counts)+1 {
		panic("bmc: invalid majority levels")
	}
	obj := int32(0)
	firsts[0] = 0
	for ; obj < nVars; obj++ {
		levels[obj] = 0
	}
	for i := range counts {
		firsts[i+1] = obj
		count := counts[len(counts)-1-i]
		if count < 0 || obj+count > nObjs {
			panic("bmc: invalid majority level count")
		}
		for j := int32(0); j < count; j++ {
			levels[obj] = int32(i + 1)
			obj++
		}
	}
	if obj != nObjs {
		panic("bmc: majority level count does not match objects")
	}
}

// Zyx_ManCreateSymVarPairs lists variable pairs whose exchange leaves truth unchanged.
func Zyx_ManCreateSymVarPairs(truth []uint64, nVars int32) *vec.Vec_Int_t {
	if nVars < 0 || nVars > 12 {
		panic("bmc: invalid truth table variable count")
	}
	n := util.Abc_TtWordNum(nVars)
	if len(truth) < int(n) {
		panic("bmc: short truth table")
	}
	result := vec.Vec_IntAlloc(16)
	c0, c1, c01, c10 := make([]uint64, n), make([]uint64, n), make([]uint64, n), make([]uint64, n)
	for i := int32(0); i < nVars; i++ {
		util.Abc_TtCofactor0p(c0, truth, n, i)
		util.Abc_TtCofactor1p(c1, truth, n, i)
		for j := i + 1; j < nVars; j++ {
			util.Abc_TtCofactor1p(c01, c0, n, j)
			util.Abc_TtCofactor0p(c10, c1, n, j)
			if util.Abc_TtEqual(c01, c10, n) != 0 {
				vec.Vec_IntPushTwo(result, i, j)
			}
		}
	}
	return result
}

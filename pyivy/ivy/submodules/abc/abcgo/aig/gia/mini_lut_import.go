package gia

import (
	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManFromMiniLut expands LUT truth tables into a structurally hashed AIG.
// When requested, copies maps MiniLUT object IDs to cleaned-up GIA literals.
func Gia_ManFromMiniLut(p *miniaig.Mini_Lut_t, out **Vec_Int_t) *Gia_Man_t {
	if p == nil || !miniaig.Mini_LutCheck(p) {
		panic("gia: invalid MiniLUT")
	}
	n := miniaig.Mini_LutNodeNum(p)
	g := Gia_ManStart(int32(3 * n))
	g.pName = "MiniLut"
	copies := vec.Vec_IntAlloc(n)
	vec.Vec_IntPush(copies, 0)
	vec.Vec_IntPush(copies, 1)
	var scratch []uint32
	Gia_ManHashAlloc(g)
	for id := 2; id < n; id++ {
		var lit int32
		switch {
		case miniaig.Mini_LutNodeIsPi(p, id):
			lit = Gia_ManAppendCi(g)
		case miniaig.Mini_LutNodeIsPo(p, id):
			lit = Gia_ManAppendCo(g, vec.Vec_IntEntry(copies, miniaig.Mini_LutNodeFanin(p, id, 0)))
		case miniaig.Mini_LutNodeIsNode(p, id):
			var leaves []int32
			for k := 0; k < miniaig.Mini_LutSize(p); k++ {
				fan := miniaig.Mini_LutNodeFanin(p, id, k)
				if fan == miniaig.MINI_LUT_NULL {
					break
				}
				leaves = append(leaves, vec.Vec_IntEntry(copies, fan))
			}
			lit = kit.Kit_TruthToGia(g, miniaig.Mini_LutNodeTruth(p, id), int32(len(leaves)), &scratch, leaves, 1)
		default:
			panic("gia: invalid MiniLUT node")
		}
		vec.Vec_IntPush(copies, lit)
	}
	Gia_ManHashStop(g)
	Gia_ManSetRegNum(g, int32(miniaig.Mini_LutRegNum(p)))
	result := Gia_ManCleanup(g)
	if out != nil {
		Gia_ManDupRemapLiterals(copies, g)
		*out = copies
	}
	Gia_ManStop(g)
	return result
}

package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// targetJustificationGraph builds a symbolic condition for preserving the
// counterexample's output using the fixed input values from each start frame.
func targetJustificationGraph(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, initial **vec.Vec_Bit_t, begin, end int32, allFrames bool) *gia.Gia_Man_t {
	if p == nil || cex == nil || initial == nil || begin < 0 || end < begin || end > cex.IFrame {
		panic("bmc: invalid target frame range")
	}
	var values, just *vec.Vec_Bit_t
	Bmc_GiaGenerateJust(p, cex, &values, &just)
	n, regs := gia.Gia_ManObjNum(p), gia.Gia_ManRegNum(p)
	*initial = vec.Vec_BitStart(int(regs))
	q := gia.Gia_ManStart(100)
	inputs := make([]int32, regs)
	for i := int32(0); i < regs; i++ {
		id := gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))
		if vec.Vec_BitEntry(values, int(begin*n+id)) != 0 {
			vec.Vec_BitWriteEntry(*initial, int(i), 1)
		}
		inputs[i] = gia.Gia_ManAppendCi(q)
	}
	gia.Gia_ManHashStart(q)
	last := end
	if allFrames {
		last = cex.IFrame
	}
	for start := end; start <= last; start++ {
		state := make([]int32, regs)
		for i := int32(0); i < regs; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))
			state[i] = inputs[i]
			if vec.Vec_BitEntry(values, int(start*n+id)) == 0 {
				state[i] ^= 1
			}
		}
		for frame := start; frame <= cex.IFrame; frame++ {
			deps := make([]int32, n)
			deps[0] = 1
			for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
				deps[gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))] = 1
			}
			for i := int32(0); i < regs; i++ {
				deps[gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))] = state[i]
			}
			for id := int32(1); id < n; id++ {
				o := gia.Gia_ManObj(p, id)
				if gia.Gia_ObjIsAnd(o) != 0 {
					a, b := gia.Gia_ObjFaninId0(o, id), gia.Gia_ObjFaninId1(o, id)
					x := vec.Vec_BitEntry(values, int(frame*n+a)) != int(gia.Gia_ObjFaninC0(o))
					y := vec.Vec_BitEntry(values, int(frame*n+b)) != int(gia.Gia_ObjFaninC1(o))
					switch {
					case gia.Gia_ObjIsXor(o) != 0 || x && y:
						deps[id] = gia.Gia_ManHashAnd(q, deps[a], deps[b])
					case !x && !y:
						deps[id] = gia.Gia_ManHashOr(q, deps[a], deps[b])
					case !x:
						deps[id] = deps[a]
					default:
						deps[id] = deps[b]
					}
				} else if gia.Gia_ObjIsCo(o) != 0 {
					deps[id] = deps[gia.Gia_ObjFaninId0(o, id)]
				}
			}
			if frame == cex.IFrame {
				gia.Gia_ManAppendCo(q, deps[gia.Gia_ObjId(p, gia.Gia_ManPo(p, cex.IPo))])
				break
			}
			for i := int32(0); i < regs; i++ {
				state[i] = deps[gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))]
			}
		}
	}
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}
func Bmc_GiaGenerateGiaAllFrames(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, initial **vec.Vec_Bit_t, begin, end int32) *gia.Gia_Man_t {
	return targetJustificationGraph(p, cex, initial, begin, end, true)
}
func Bmc_GiaGenerateGiaAllOne(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, initial **vec.Vec_Bit_t, begin, end int32) *gia.Gia_Man_t {
	return targetJustificationGraph(p, cex, initial, begin, end, false)
}

package bmc

import "github.com/glycerine/abcgo/aig/gia"

// Bmc_CexTargetEnlarge unrolls the transition relation from free state
// variables. Primary inputs are ordered from the last frame to the first.
func Bmc_CexTargetEnlarge(p *gia.Gia_Man_t, frames int32) *gia.Gia_Man_t {
	if p == nil || frames <= 0 || gia.Gia_ManPoNum(p) == 0 {
		return nil
	}
	q := gia.Gia_ManStart(max(1, gia.Gia_ManObjNum(p)))
	pis, regs, n := gia.Gia_ManPiNum(p), gia.Gia_ManRegNum(p), gia.Gia_ManObjNum(p)
	frameInputs := make([][]int32, frames)
	for f := range frameInputs {
		frameInputs[f] = make([]int32, pis)
		for i := range frameInputs[f] {
			frameInputs[f][i] = gia.Gia_ManAppendCi(q)
		}
	}
	state := make([]int32, regs)
	for i := range state {
		state[i] = gia.Gia_ManAppendCi(q)
	}
	gia.Gia_ManHashStart(q)
	var mapping []int32
	for f := int32(0); f < frames; f++ {
		mapping = make([]int32, n)
		for i := int32(0); i < pis; i++ {
			mapping[gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))] = frameInputs[frames-1-f][i]
		}
		for i := int32(0); i < regs; i++ {
			mapping[gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))] = state[i]
		}
		for id := int32(1); id < n; id++ {
			o := gia.Gia_ManObj(p, id)
			if gia.Gia_ObjIsAnd(o) != 0 {
				a := mapping[gia.Gia_ObjFaninId0(o, id)] ^ gia.Gia_ObjFaninC0(o)
				b := mapping[gia.Gia_ObjFaninId1(o, id)] ^ gia.Gia_ObjFaninC1(o)
				switch {
				case gia.Gia_ObjIsMux(p, o) != 0:
					c := mapping[gia.Gia_ObjFaninId2(p, id)] ^ gia.Gia_ObjFaninC2(p, o)
					mapping[id] = gia.Gia_ManHashMux(q, c, b, a)
				case gia.Gia_ObjIsXor(o) != 0:
					mapping[id] = gia.Gia_ManHashXor(q, a, b)
				default:
					mapping[id] = gia.Gia_ManHashAnd(q, a, b)
				}
			} else if gia.Gia_ObjIsCo(o) != 0 {
				mapping[id] = mapping[gia.Gia_ObjFaninId0(o, id)] ^ gia.Gia_ObjFaninC0(o)
			}
		}
		for i := int32(0); i < regs; i++ {
			state[i] = mapping[gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))]
		}
	}
	gia.Gia_ManAppendCo(q, mapping[gia.Gia_ObjId(p, gia.Gia_ManPo(p, 0))])
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}
func Bmc_CexTarget(p *gia.Gia_Man_t, frames int32) *gia.Gia_Man_t {
	q := Bmc_CexTargetEnlarge(p, frames)
	if q == nil {
		return nil
	}
	for i := int32(0); i < frames*gia.Gia_ManPiNum(p); i++ {
		q = gia.Gia_ManDupExist(q, i)
	}
	q = gia.Gia_ManDupLastPis(q, gia.Gia_ManRegNum(p))
	return gia.Gia_ManDupAppendCones(p, []*gia.Gia_Man_t{q}, 1, 1)
}

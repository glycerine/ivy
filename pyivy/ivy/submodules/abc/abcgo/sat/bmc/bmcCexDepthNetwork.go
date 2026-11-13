package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

type justifiedBit struct {
	value bool
	dep   int32
}

func justifiedAnd(q *gia.Gia_Man_t, a, b justifiedBit) justifiedBit {
	v := a.value && b.value
	switch {
	case v:
		return justifiedBit{true, gia.Gia_ManHashAnd(q, a.dep, b.dep)}
	case !a.value && !b.value:
		return justifiedBit{false, gia.Gia_ManHashOr(q, a.dep, b.dep)}
	case !a.value:
		return justifiedBit{false, a.dep}
	default:
		return justifiedBit{false, b.dep}
	}
}
func justifiedNot(a justifiedBit) justifiedBit { a.value = !a.value; return a }

// Bmc_CexBuildNetwork2 derives the condition on the state at start that
// preserves the failing output while replaying the CEX primary inputs.
func Bmc_CexBuildNetwork2(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, start int32) *gia.Gia_Man_t {
	if p == nil || cex == nil || start < 0 || start > cex.IFrame || cex.NRegs != 0 || cex.NPis != gia.Gia_ManCiNum(p) || cex.IPo < 0 || cex.IPo >= gia.Gia_ManPoNum(p) {
		panic("bmc: invalid state-extended counterexample")
	}
	q := gia.Gia_ManStart(100)
	regs, pis, n := gia.Gia_ManRegNum(p), gia.Gia_ManPiNum(p), gia.Gia_ManObjNum(p)
	state := make([]justifiedBit, regs)
	for i := int32(0); i < regs; i++ {
		v := util.Abc_InfoHasBit(cex.PData, start*cex.NPis+pis+i)
		lit := gia.Gia_ManAppendCi(q)
		if !v {
			lit ^= 1
		}
		state[i] = justifiedBit{v, lit}
	}
	gia.Gia_ManHashStart(q)
	var property justifiedBit
	for frame := start; frame <= cex.IFrame; frame++ {
		values := make([]justifiedBit, n)
		values[0] = justifiedBit{false, 1}
		for i := int32(0); i < pis; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))
			values[id] = justifiedBit{util.Abc_InfoHasBit(cex.PData, frame*cex.NPis+i), 1}
		}
		for i := int32(0); i < regs; i++ {
			values[gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))] = state[i]
		}
		read := func(lit int32) justifiedBit {
			v := values[lit>>1]
			if lit&1 != 0 {
				v = justifiedNot(v)
			}
			return v
		}
		for id := int32(1); id < n; id++ {
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 2:
				values[id] = justifiedAnd(q, read(a), read(b))
			case 3:
				x, y := read(a), read(b)
				values[id] = justifiedBit{x.value != y.value, gia.Gia_ManHashAnd(q, x.dep, y.dep)}
			case 4:
				control, yes, no := read(a), read(b), read(c)
				left := justifiedAnd(q, control, yes)
				right := justifiedAnd(q, justifiedNot(control), no)
				values[id] = justifiedNot(justifiedAnd(q, justifiedNot(left), justifiedNot(right)))
			case 5, 6:
				values[id] = read(a)
			}
		}
		property = values[gia.Gia_ObjId(p, gia.Gia_ManPo(p, cex.IPo))]
		for i := int32(0); i < regs; i++ {
			state[i] = values[gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))]
		}
	}
	if !property.value {
		panic("bmc: counterexample does not fail")
	}
	gia.Gia_ManAppendCo(q, property.dep)
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}
func Bmc_CexBuildNetwork2_(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, start int32) *gia.Gia_Man_t {
	return Bmc_CexBuildNetwork2(p, cex, start)
}
func Bmc_CexBuildNetwork2Test(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, frames int32) *gia.Gia_Man_t {
	if p == nil || cex == nil || frames < 0 {
		return nil
	}
	frames = min(frames, cex.IFrame)
	cones := make([]*gia.Gia_Man_t, 0, frames)
	for i := cex.IFrame; i > cex.IFrame-frames; i-- {
		cones = append(cones, Bmc_CexBuildNetwork2(p, cex, i))
	}
	return gia.Gia_ManDupAppendCones(p, cones, int32(len(cones)), 1)
}

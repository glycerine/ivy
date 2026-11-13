package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// Bmc_CexPerformUnrolling creates the selected final-frame property under
// zero initial state, with a primary input for each original input and frame.
func Bmc_CexPerformUnrolling(p *gia.Gia_Man_t, cex *util.Abc_Cex_t) *gia.Gia_Man_t {
	if p == nil || cex == nil || cex.IFrame < 0 || cex.IPo < 0 || cex.IPo >= gia.Gia_ManPoNum(p) {
		return nil
	}
	q := gia.Gia_ManStart(max(1, gia.Gia_ManObjNum(p)))
	gia.Gia_ManHashStart(q)
	pis, regs, n := gia.Gia_ManPiNum(p), gia.Gia_ManRegNum(p), gia.Gia_ManObjNum(p)
	state := make([]int32, regs)
	var values []int32
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		values = make([]int32, n)
		for i := int32(0); i < pis; i++ {
			values[gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))] = gia.Gia_ManAppendCi(q)
		}
		for i := int32(0); i < regs; i++ {
			values[gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))] = state[i]
		}
		read := func(lit int32) int32 { return values[lit>>1] ^ (lit & 1) }
		for id := int32(1); id < n; id++ {
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 2:
				values[id] = gia.Gia_ManHashAnd(q, read(a), read(b))
			case 3:
				values[id] = gia.Gia_ManHashXor(q, read(a), read(b))
			case 4:
				values[id] = gia.Gia_ManHashMux(q, read(a), read(b), read(c))
			case 5, 6:
				values[id] = read(a)
			}
		}
		for i := int32(0); i < regs; i++ {
			state[i] = values[gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))]
		}
	}
	gia.Gia_ManAppendCo(q, values[gia.Gia_ObjId(p, gia.Gia_ManPo(p, cex.IPo))])
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}

// Bmc_CexBuildNetwork constructs the unate condition on frame input bits that
// preserves the failing trace under the original zero initial state.
func Bmc_CexBuildNetwork(p *gia.Gia_Man_t, cex *util.Abc_Cex_t) *gia.Gia_Man_t {
	if p == nil || cex == nil || cex.IFrame < 0 || cex.NPis != gia.Gia_ManPiNum(p) || cex.NRegs != gia.Gia_ManRegNum(p) || cex.IPo < 0 || cex.IPo >= gia.Gia_ManPoNum(p) {
		return nil
	}
	q := gia.Gia_ManStart(100)
	gia.Gia_ManHashStart(q)
	pis, regs, n := gia.Gia_ManPiNum(p), gia.Gia_ManRegNum(p), gia.Gia_ManObjNum(p)
	state := make([]justifiedBit, regs)
	for i := range state {
		state[i] = justifiedBit{false, 1}
	}
	var property justifiedBit
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		values := make([]justifiedBit, n)
		values[0] = justifiedBit{false, 1}
		for i := int32(0); i < pis; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))
			values[id] = justifiedBit{util.Abc_InfoHasBit(cex.PData, cex.NRegs+frame*pis+i), gia.Gia_ManAppendCi(q)}
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
				ctl, yes, no := read(a), read(b), read(c)
				x := justifiedAnd(q, ctl, yes)
				y := justifiedAnd(q, justifiedNot(ctl), no)
				values[id] = justifiedNot(justifiedAnd(q, justifiedNot(x), justifiedNot(y)))
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
		return nil
	}
	gia.Gia_ManAppendCo(q, property.dep)
	gia.Gia_ManHashStop(q)
	return gia.Gia_ManCleanup(q)
}

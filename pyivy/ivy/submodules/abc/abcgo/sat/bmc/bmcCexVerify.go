package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// Bmc_CexVerify returns one only when the retained input assignments force the
// indicated output high under ternary simulation from the zero initial state.
func Bmc_CexVerify(p *gia.Gia_Man_t, cex, care *util.Abc_Cex_t) int32 {
	if p == nil || cex == nil || care == nil || cex.IFrame < 0 || cex.IPo < 0 || cex.IPo >= gia.Gia_ManPoNum(p) || cex.NPis != gia.Gia_ManPiNum(p) || cex.NRegs != gia.Gia_ManRegNum(p) || care.NPis < cex.NPis || care.IFrame < cex.IFrame {
		return 0
	}
	pis, regs, n := gia.Gia_ManPiNum(p), gia.Gia_ManRegNum(p), gia.Gia_ManObjNum(p)
	state := make([]impliedValue, regs)
	for i := range state {
		state[i] = impliedValue{false, true}
	}
	var output impliedValue
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		values := make([]impliedValue, n)
		values[0] = impliedValue{false, true}
		for i := int32(0); i < pis; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))
			known := util.Abc_InfoHasBit(care.PData, care.NRegs+frame*care.NPis+i)
			value := util.Abc_InfoHasBit(cex.PData, cex.NRegs+frame*pis+i)
			values[id] = impliedValue{value, known}
		}
		for i := int32(0); i < regs; i++ {
			values[gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))] = state[i]
		}
		read := func(lit int32) impliedValue {
			v := values[lit>>1]
			if lit&1 != 0 {
				v = impliedNot(v)
			}
			return v
		}
		for id := int32(1); id < n; id++ {
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 2:
				values[id] = impliedAnd(read(a), read(b))
			case 3:
				x, y := read(a), read(b)
				values[id] = impliedValue{x.value != y.value, x.known && y.known}
			case 4:
				ctl, yes, no := read(a), read(b), read(c)
				x := impliedAnd(ctl, yes)
				y := impliedAnd(impliedNot(ctl), no)
				values[id] = impliedNot(impliedAnd(impliedNot(x), impliedNot(y)))
			case 5, 6:
				values[id] = read(a)
			}
		}
		output = values[gia.Gia_ObjId(p, gia.Gia_ManPo(p, cex.IPo))]
		for i := int32(0); i < regs; i++ {
			state[i] = values[gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))]
		}
	}
	if output.known && output.value {
		return 1
	}
	return 0
}

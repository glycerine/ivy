package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// Bmc_CexCareBits traces sufficient input and register bits backwards from the
// failing output, excluding values already implied by the initial state.
func Bmc_CexCareBits(p *gia.Gia_Man_t, states, implied, essential *util.Abc_Cex_t, findAll, verbose int32) *util.Abc_Cex_t {
	if p == nil || states == nil || states.NRegs != 0 || states.NPis != gia.Gia_ManCiNum(p) || states.IFrame < 0 || states.IPo < 0 || states.IPo >= gia.Gia_ManPoNum(p) {
		panic("bmc: invalid expanded counterexample")
	}
	n, ci, pis, regs := gia.Gia_ManObjNum(p), gia.Gia_ManCiNum(p), gia.Gia_ManPiNum(p), gia.Gia_ManRegNum(p)
	care := util.Abc_CexAlloc(0, ci, states.IFrame+1)
	care.IFrame, care.IPo = states.IFrame, states.IPo
	for frame := states.IFrame; frame >= 0; frame-- {
		values := make([]impliedValue, n)
		values[0] = impliedValue{false, true}
		for i := int32(0); i < ci; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))
			bit := frame*ci + i
			known := (implied != nil && util.Abc_InfoHasBit(implied.PData, bit)) || (essential != nil && util.Abc_InfoHasBit(essential.PData, bit))
			values[id] = impliedValue{util.Abc_InfoHasBit(states.PData, bit), known}
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
		var mark func(int32)
		mark = func(id int32) {
			if id == 0 || values[id].known {
				return
			}
			values[id].known = true
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 1:
				return
			case 2:
				x, y := read(a), read(b)
				if values[id].value {
					mark(a >> 1)
					mark(b >> 1)
				} else if findAll != 0 {
					if !x.value {
						mark(a >> 1)
					}
					if !y.value {
						mark(b >> 1)
					}
				} else if !x.value {
					mark(a >> 1)
				} else {
					mark(b >> 1)
				}
			case 3:
				mark(a >> 1)
				mark(b >> 1)
			case 4:
				mark(a >> 1)
				if read(a).value {
					mark(b >> 1)
				} else {
					mark(c >> 1)
				}
			case 5, 6:
				mark(a >> 1)
			}
		}
		if frame == states.IFrame {
			po := gia.Gia_ObjId(p, gia.Gia_ManPo(p, states.IPo))
			if !values[po].value {
				panic("bmc: counterexample output is not asserted")
			}
			a, _, _, _ := p.Gate(po)
			mark(a >> 1)
		} else {
			for i := int32(0); i < regs; i++ {
				if util.Abc_InfoHasBit(care.PData, (frame+1)*ci+pis+i) {
					ri := gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))
					a, _, _, _ := p.Gate(ri)
					mark(a >> 1)
				}
			}
		}
		for i := int32(0); i < ci; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))
			bit := frame*ci + i
			if values[id].known && (implied == nil || !util.Abc_InfoHasBit(implied.PData, bit)) {
				util.Abc_InfoSetBit(care.PData, bit)
			}
		}
	}
	_ = verbose
	return care
}

package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

type impliedValue struct{ value, known bool }

func impliedNot(v impliedValue) impliedValue { v.value = !v.value; return v }
func impliedAnd(a, b impliedValue) impliedValue {
	if a.value && b.value {
		return impliedValue{true, a.known && b.known}
	}
	if !a.value && !b.value {
		return impliedValue{false, a.known || b.known}
	}
	if !a.value {
		return impliedValue{false, a.known}
	}
	return impliedValue{false, b.known}
}

// Bmc_CexInnerStates expands each frame into primary-input and register bits;
// implied marks identify register values forced by the zero initial state.
func Bmc_CexInnerStates(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, implied **util.Abc_Cex_t, verbose int32) *util.Abc_Cex_t {
	if p == nil || cex == nil || cex.IFrame < 0 || cex.NRegs != gia.Gia_ManRegNum(p) || cex.NPis != gia.Gia_ManPiNum(p) {
		panic("bmc: invalid counterexample dimensions")
	}
	pis, regs, n := gia.Gia_ManPiNum(p), gia.Gia_ManRegNum(p), gia.Gia_ManObjNum(p)
	states := util.Abc_CexAlloc(0, pis+regs, cex.IFrame+1)
	marks := util.Abc_CexAlloc(0, pis+regs, cex.IFrame+1)
	states.IFrame, states.IPo = cex.IFrame, cex.IPo
	marks.IFrame, marks.IPo = cex.IFrame, cex.IPo
	state := make([]impliedValue, regs)
	for i := range state {
		state[i] = impliedValue{false, true}
	}
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		values := make([]impliedValue, n)
		values[0] = impliedValue{false, true}
		for i := int32(0); i < pis; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))
			v := util.Abc_InfoHasBit(cex.PData, regs+frame*pis+i)
			values[id] = impliedValue{v, false}
			if v {
				util.Abc_InfoSetBit(states.PData, frame*(pis+regs)+i)
			}
		}
		for i := int32(0); i < regs; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManRo(p, i))
			values[id] = state[i]
			if state[i].value {
				util.Abc_InfoSetBit(states.PData, frame*(pis+regs)+pis+i)
			}
			if state[i].known {
				util.Abc_InfoSetBit(marks.PData, frame*(pis+regs)+pis+i)
			}
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
				control, yes, no := read(a), read(b), read(c)
				left := impliedAnd(control, yes)
				right := impliedAnd(impliedNot(control), no)
				values[id] = impliedNot(impliedAnd(impliedNot(left), impliedNot(right)))
			case 5, 6:
				values[id] = read(a)
			}
		}
		for i := int32(0); i < regs; i++ {
			state[i] = values[gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))]
		}
	}
	if implied != nil {
		*implied = marks
	}
	_ = verbose
	return states
}

package gia

import (
	"github.com/glycerine/abcgo/misc/util"
)

// Gia_ManCexCareVerify returns one only when the retained input assignments force the
// indicated output high under ternary simulation from the zero initial state.
func Gia_ManCexCareVerify(p *Gia_Man_t, cex, care *util.Abc_Cex_t) int32 {
	if p == nil || cex == nil || care == nil || cex.IFrame < 0 || cex.IPo < 0 || cex.IPo >= Gia_ManPoNum(p) || cex.NPis != Gia_ManPiNum(p) || cex.NRegs != Gia_ManRegNum(p) || care.NPis < cex.NPis || care.IFrame < cex.IFrame {
		return 0
	}
	pis, regs, n := Gia_ManPiNum(p), Gia_ManRegNum(p), Gia_ManObjNum(p)
	state := make([]giaCexImplied, regs)
	for i := range state {
		state[i] = giaCexImplied{false, true}
	}
	var output giaCexImplied
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		values := make([]giaCexImplied, n)
		values[0] = giaCexImplied{false, true}
		for i := int32(0); i < pis; i++ {
			id := Gia_ObjId(p, Gia_ManCi(p, i))
			known := util.Abc_InfoHasBit(care.PData, care.NRegs+frame*care.NPis+i)
			value := util.Abc_InfoHasBit(cex.PData, cex.NRegs+frame*pis+i)
			values[id] = giaCexImplied{value, known}
		}
		for i := int32(0); i < regs; i++ {
			values[Gia_ObjId(p, Gia_ManRo(p, i))] = state[i]
		}
		read := func(lit int32) giaCexImplied {
			v := values[lit>>1]
			if lit&1 != 0 {
				v = giaCexImpliedNot(v)
			}
			return v
		}
		for id := int32(1); id < n; id++ {
			a, b, c, kind := p.Gate(id)
			switch kind {
			case 2:
				values[id] = giaCexImpliedAnd(read(a), read(b))
			case 3:
				x, y := read(a), read(b)
				values[id] = giaCexImplied{x.value != y.value, x.known && y.known}
			case 4:
				ctl, yes, no := read(a), read(b), read(c)
				x := giaCexImpliedAnd(ctl, yes)
				y := giaCexImpliedAnd(giaCexImpliedNot(ctl), no)
				values[id] = giaCexImpliedNot(giaCexImpliedAnd(giaCexImpliedNot(x), giaCexImpliedNot(y)))
			case 5, 6:
				values[id] = read(a)
			}
		}
		output = values[Gia_ObjId(p, Gia_ManPo(p, cex.IPo))]
		for i := int32(0); i < regs; i++ {
			state[i] = values[Gia_ObjId(p, Gia_ManRi(p, i))]
		}
	}
	if output.known && output.value {
		return 1
	}
	return 0
}

type giaCexImplied struct{ value, known bool }

func giaCexImpliedNot(v giaCexImplied) giaCexImplied { v.value = !v.value; return v }
func giaCexImpliedAnd(a, b giaCexImplied) giaCexImplied {
	if a.value && b.value {
		return giaCexImplied{true, a.known && b.known}
	}
	if !a.value && !b.value {
		return giaCexImplied{false, a.known || b.known}
	}
	if !a.value {
		return giaCexImplied{false, a.known}
	}
	return giaCexImplied{false, b.known}
}

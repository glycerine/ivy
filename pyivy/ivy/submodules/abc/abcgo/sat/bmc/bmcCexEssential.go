package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// Bmc_CexEssentialBitOne checks the effect of removing one expanded-trace bit.
// It returns the remaining care mask, or nil if the effect stops propagating.
func Bmc_CexEssentialBitOne(p *gia.Gia_Man_t, states *util.Abc_Cex_t, bit int32, previous *util.Abc_Cex_t, equal *int32) *util.Abc_Cex_t {
	if p == nil || states == nil || states.NRegs != 0 || states.NPis != gia.Gia_ManCiNum(p) || states.NPis <= 0 || bit < 0 || bit >= states.NBits {
		panic("bmc: invalid essential-bit analysis")
	}
	if equal != nil {
		*equal = 0
	}
	mask := util.Abc_CexAllocFull(0, states.NPis, states.IFrame+1)
	mask.IFrame, mask.IPo = states.IFrame, states.IPo
	mask.PData[bit>>5] ^= 1 << uint(bit&31)
	pis, regs, n := gia.Gia_ManPiNum(p), gia.Gia_ManRegNum(p), gia.Gia_ManObjNum(p)
	for frame := bit / states.NPis; frame <= states.IFrame; frame++ {
		values := make([]impliedValue, n)
		values[0] = impliedValue{false, true}
		for i := int32(0); i < states.NPis; i++ {
			id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))
			pos := frame*states.NPis + i
			values[id] = impliedValue{util.Abc_InfoHasBit(states.PData, pos), util.Abc_InfoHasBit(mask.PData, pos)}
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
		for id := int32(0); id < n; id++ {
			if values[id].known {
				gia.Gia_ObjSetMark1(gia.Gia_ManObj(p, id), 1)
			} else {
				gia.Gia_ObjSetMark1(gia.Gia_ManObj(p, id), 0)
			}
		}
		if frame < states.IFrame {
			changed, same := false, previous != nil
			offset := (frame+1)*states.NPis + pis
			for i := int32(0); i < regs; i++ {
				known := values[gia.Gia_ObjId(p, gia.Gia_ManRi(p, i))].known
				if same && known != util.Abc_InfoHasBit(previous.PData, offset+i) {
					same = false
				}
				if !known {
					changed = true
					mask.PData[(offset+i)>>5] &^= 1 << uint((offset+i)&31)
				}
			}
			if !changed || same {
				if equal != nil && same {
					*equal = 1
				}
				return nil
			}
		}
	}
	return mask
}
func Bmc_CexEssentialBits(p *gia.Gia_Man_t, states, care *util.Abc_Cex_t, verbose int32) *util.Abc_Cex_t {
	if p == nil || states == nil || care == nil || states.NBits != care.NBits {
		panic("bmc: incompatible care mask")
	}
	result := util.Abc_CexAlloc(0, states.NPis, states.IFrame+1)
	result.IFrame, result.IPo = states.IFrame, states.IPo
	var previous *util.Abc_Cex_t
	previousEssential := false
	for bit := int32(0); bit < states.NBits; bit++ {
		if !util.Abc_InfoHasBit(care.PData, bit) {
			continue
		}
		if bit%states.NPis >= gia.Gia_ManPiNum(p) {
			util.Abc_InfoSetBit(result.PData, bit)
			continue
		}
		var equal int32
		trial := Bmc_CexEssentialBitOne(p, states, bit, previous, &equal)
		if trial == nil {
			if equal != 0 && previousEssential {
				util.Abc_InfoSetBit(result.PData, bit)
			}
			continue
		}
		previous = trial
		previousEssential = gia.Gia_ObjMark1(gia.Gia_ManPo(p, states.IPo)) == 0
		if previousEssential {
			util.Abc_InfoSetBit(result.PData, bit)
		}
	}
	_ = verbose
	return result
}

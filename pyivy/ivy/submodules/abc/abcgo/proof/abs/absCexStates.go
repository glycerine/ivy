package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// replayCexFromZero evaluates the specified suffix with registers initially zero.
func replayCexFromZero(p *gia.Gia_Man_t, c *util.Abc_Cex_t, start, capture int32) ([]int32, bool) {
	if p == nil || c == nil || start < 0 || start > c.IFrame || capture < start || capture > c.IFrame || c.NRegs != gia.Gia_ManRegNum(p) || c.NPis != gia.Gia_ManPiNum(p) || c.IPo < 0 || c.IPo >= gia.Gia_ManPoNum(p) || int64(c.NBits) != int64(c.NRegs)+(int64(c.IFrame)+1)*int64(c.NPis) || int64(c.NBits) > int64(len(c.PData))*32 {
		panic("abs: invalid counterexample dimensions")
	}
	gia.Gia_ManCleanMark0(p)
	defer gia.Gia_ManCleanMark0(p)
	inputs := make([]bool, c.NPis+c.NRegs)
	values := make([]bool, gia.Gia_ManObjNum(p))
	state := make([]int32, c.NRegs)
	for frame := start; frame <= c.IFrame; frame++ {
		if frame == capture {
			for i := range state {
				if inputs[int(c.NPis)+i] {
					state[i] = 1
				}
			}
		}
		for i := int32(0); i < c.NPis; i++ {
			inputs[i] = util.Abc_InfoHasBit(c.PData, c.NRegs+frame*c.NPis+i)
		}
		for i, input := range p.InputLiterals() {
			values[input>>1] = inputs[i]
		}
		// absOut.c iterates Gia_ManForEachAnd and uses the two physical
		// fanins as an AND, including native XOR and MUX objects.
		for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
			obj := gia.Gia_ManObj(p, id)
			if gia.Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			a := gia.Gia_ObjId(p, gia.Gia_ObjFanin0(obj))
			b := gia.Gia_ObjId(p, gia.Gia_ObjFanin1(obj))
			values[id] = (values[a] != (gia.Gia_ObjFaninC0(obj) != 0)) && (values[b] != (gia.Gia_ObjFaninC1(obj) != 0))
		}
		outputs := make([]bool, gia.Gia_ManCoNum(p))
		for i, lit := range p.OutputLiterals() {
			outputs[i] = values[lit>>1] != (lit&1 != 0)
		}
		if frame == c.IFrame {
			return state, outputs[c.IPo]
		}
		copy(inputs[c.NPis:], outputs[gia.Gia_ManPoNum(p):])
	}
	panic("abs: unreachable counterexample frame")
}

func Gia_ManGetStateAndCheckCex(p *gia.Gia_Man_t, c *util.Abc_Cex_t, frame int32) *vec.Vec_Int_t {
	state, valid := replayCexFromZero(p, c, 0, frame)
	if !valid {
		return nil
	}
	result := vec.Vec_IntAlloc(len(state))
	for _, bit := range state {
		vec.Vec_IntPush(result, bit)
	}
	return result
}

func Gia_ManCheckCex(p *gia.Gia_Man_t, c *util.Abc_Cex_t, frame int32) {
	_, valid := replayCexFromZero(p, c, frame, frame)
	if valid {
		fmt.Println("Shortened CEX holds for the abstraction of the fast-forwarded model.")
	} else {
		fmt.Println("Shortened CEX does not hold for the abstraction of the fast-forwarded model.")
	}
}

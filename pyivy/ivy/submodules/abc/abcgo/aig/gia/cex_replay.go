package gia

import "github.com/glycerine/abcgo/misc/util"

// replayCounterexample evaluates every frame, starting from either the witness
// register bits or zero. Register inputs are copied simultaneously between frames.
func replayCounterexample(p *Gia_Man_t, cex *util.Abc_Cex_t, zeroState bool, visit func(int32, []bool, []bool, []bool) bool) {
	if p == nil || cex == nil || cex.NRegs != Gia_ManRegNum(p) || cex.NPis != Gia_ManPiNum(p) || cex.IFrame < 0 {
		panic("gia: invalid counterexample dimensions")
	}
	required := int64(cex.NRegs) + (int64(cex.IFrame)+1)*int64(cex.NPis)
	if required < 0 || required != int64(cex.NBits) || required > int64(len(cex.PData))*32 {
		panic("gia: inconsistent counterexample bits")
	}
	bit := func(i int64) bool { return cex.PData[i>>5]&(1<<uint(i&31)) != 0 }
	inputs := make([]bool, Gia_ManCiNum(p))
	if !zeroState {
		for i := int32(0); i < cex.NRegs; i++ {
			inputs[cex.NPis+i] = bit(int64(i))
		}
	}
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		for i := int32(0); i < cex.NPis; i++ {
			inputs[i] = bit(int64(cex.NRegs) + int64(frame)*int64(cex.NPis) + int64(i))
		}
		values, outputs, err := p.evaluateObjects(inputs)
		if err != nil {
			panic(err)
		}
		if !visit(frame, inputs, values, outputs) {
			return
		}
		copy(inputs[cex.NPis:], outputs[Gia_ManPoNum(p):])
	}
}

// Gia_ManVerifyCex checks the indicated output in the final frame.
func Gia_ManVerifyCex(p *Gia_Man_t, cex *util.Abc_Cex_t, dual int32) int32 {
	index := cex.IPo
	if dual != 0 {
		index *= 2
	}
	if index < 0 || index >= Gia_ManPoNum(p) || dual != 0 && index+1 >= Gia_ManPoNum(p) {
		panic("gia: invalid counterexample output")
	}
	var failed bool
	replayCounterexample(p, cex, false, func(frame int32, inputs, values, outputs []bool) bool {
		if frame == cex.IFrame {
			failed = outputs[index]
			if dual != 0 {
				failed = failed != outputs[index+1]
			}
		}
		return true
	})
	if failed {
		return 1
	}
	return 0
}

// Gia_ManFindFailedPoCex finds the first failing final-frame output at or after start.
func Gia_ManFindFailedPoCex(p *Gia_Man_t, cex *util.Abc_Cex_t, start int32) int32 {
	if start < 0 {
		panic("gia: invalid output index")
	}
	result := int32(-1)
	replayCounterexample(p, cex, false, func(frame int32, inputs, values, outputs []bool) bool {
		if frame == cex.IFrame {
			for i := start; i < Gia_ManPoNum(p); i++ {
				if outputs[i] {
					result = i
					break
				}
			}
		}
		return true
	})
	return result
}

// Gia_ManSetFailedPoCex finds the first failure, assuming zero initial state.
func Gia_ManSetFailedPoCex(p *Gia_Man_t, cex *util.Abc_Cex_t) int32 {
	cex.IPo = -1
	replayCounterexample(p, cex, true, func(frame int32, inputs, values, outputs []bool) bool {
		for i := int32(0); i < Gia_ManPoNum(p); i++ {
			if outputs[i] {
				cex.IPo = i
				cex.IFrame = frame
				cex.NBits = cex.NRegs + (frame+1)*cex.NPis
				return false
			}
		}
		return true
	})
	return cex.IPo
}

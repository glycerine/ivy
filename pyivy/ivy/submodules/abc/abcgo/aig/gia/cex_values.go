package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
)

// Gia_ManCounterExampleValueStart records the value of every object in each frame.
func Gia_ManCounterExampleValueStart(p *Gia_Man_t, cex *util.Abc_Cex_t) {
	if Gia_ManRegNum(p) <= 0 || p.pData2 != nil {
		panic("gia: invalid counterexample probing state")
	}
	giaValidateReplayCounterexample(p, cex)
	for i := int32(0); i < cex.NRegs; i++ {
		if cex.PData[i>>5]&(1<<uint(i&31)) != 0 {
			panic("gia: counterexample initial state is not zero")
		}
	}
	words := make([]uint32, ((int64(cex.IFrame)+1)*int64(p.nObjs)+31)/32)
	replayCounterexample(p, cex, false, func(frame int32, _, values, outputs []bool) bool {
		for id, value := range values {
			if value {
				bit := int64(frame)*int64(p.nObjs) + int64(id)
				words[bit>>5] |= 1 << uint(bit&31)
			}
		}
		if frame == cex.IFrame && (cex.IPo < 0 || cex.IPo >= Gia_ManPoNum(p) || !outputs[cex.IPo]) {
			panic("gia: counterexample does not assert its output")
		}
		return true
	})
	p.pData2, p.cexLastFrame = words, cex.IFrame
}

func Gia_ManCounterExampleValueStop(p *Gia_Man_t) {
	if p.pData2 == nil {
		panic("gia: counterexample probing not started")
	}
	p.pData2, p.cexLastFrame = nil, 0
}

func Gia_ManCounterExampleValueLookup(p *Gia_Man_t, id, frame int32) int32 {
	if p.pData2 == nil || frame < 0 || frame > p.cexLastFrame || id < 0 || id >= Gia_ManObjNum(p) {
		panic("gia: counterexample probe out of range")
	}
	bit := int64(frame)*int64(p.nObjs) + int64(id)
	return int32(p.pData2[bit>>5] >> uint(bit&31) & 1)
}

func Gia_ManCounterExampleValueTest(p *Gia_Man_t, cex *util.Abc_Cex_t) {
	id, frame := Gia_ManObjNum(p)/2, cex.IFrame-1
	if frame < 0 {
		frame = 0
	}
	fmt.Printf("\nUsing counter-example, which asserts output %d in frame %d.\n", cex.IPo, cex.IFrame)
	Gia_ManCounterExampleValueStart(p, cex)
	defer Gia_ManCounterExampleValueStop(p)
	fmt.Printf("Value of object %d in frame %d is %d.\n", id, frame, Gia_ManCounterExampleValueLookup(p, id, frame))
}

func Gia_ManCexExtendToIncludeCurrentStates(p *Gia_Man_t, cex *util.Abc_Cex_t) *util.Abc_Cex_t {
	return giaCexExtend(p, cex, false)
}

func Gia_ManCexExtendToIncludeAllObjects(p *Gia_Man_t, cex *util.Abc_Cex_t) *util.Abc_Cex_t {
	return giaCexExtend(p, cex, true)
}

func giaCexExtend(p *Gia_Man_t, cex *util.Abc_Cex_t, all bool) *util.Abc_Cex_t {
	if cex.NRegs <= 0 {
		panic("gia: expected sequential counterexample")
	}
	width := Gia_ManCiNum(p)
	if all {
		width = Gia_ManObjNum(p)
	}
	result := util.Abc_CexAlloc(0, width, cex.IFrame+1)
	result.IPo, result.IFrame = cex.IPo, cex.IFrame
	replayCounterexample(p, cex, false, func(frame int32, inputs, values, outputs []bool) bool {
		selected := inputs
		if all {
			selected = values
		}
		for i, value := range selected {
			if value {
				bit := int64(frame)*int64(width) + int64(i)
				result.PData[bit>>5] |= 1 << uint(bit&31)
			}
		}
		if frame == cex.IFrame && (cex.IPo < 0 || cex.IPo >= Gia_ManPoNum(p) || !outputs[cex.IPo]) {
			panic("gia: counterexample does not assert its output")
		}
		return true
	})
	return result
}

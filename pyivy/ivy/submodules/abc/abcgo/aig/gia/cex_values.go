package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
)

// Gia_ManCounterExampleValueStart records the value of every object in each frame.
func Gia_ManCounterExampleValueStart(p *Gia_Man_t, cex *util.Abc_Cex_t) {
	if Gia_ManRegNum(p) <= 0 || p.cexValues != nil || p.pData2 != nil {
		panic("gia: invalid counterexample probing state")
	}
	for i := int32(0); i < cex.NRegs; i++ {
		if cex.PData[i>>5]&(1<<uint(i&31)) != 0 {
			panic("gia: counterexample initial state is not zero")
		}
	}
	var frames [][]bool
	replayCounterexample(p, cex, false, func(_ int32, _, values, _ []bool) bool { frames = append(frames, values); return true })
	if cex.IPo < 0 || cex.IPo >= Gia_ManPoNum(p) || !frames[cex.IFrame][Gia_ObjId(p, Gia_ManPo(p, cex.IPo))] {
		panic("gia: counterexample does not assert its output")
	}
	p.cexValues = frames
}

func Gia_ManCounterExampleValueStop(p *Gia_Man_t) {
	if p.cexValues == nil {
		panic("gia: counterexample probing not started")
	}
	p.cexValues = nil
}

func Gia_ManCounterExampleValueLookup(p *Gia_Man_t, id, frame int32) int32 {
	if frame < 0 || int(frame) >= len(p.cexValues) || id < 0 || id >= Gia_ManObjNum(p) {
		panic("gia: counterexample probe out of range")
	}
	if p.cexValues[frame][id] {
		return 1
	}
	return 0
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

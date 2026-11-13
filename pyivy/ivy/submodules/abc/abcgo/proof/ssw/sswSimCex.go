package ssw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"math/bits"
)

func Ssw_SmlCheckOutputSavePattern(p *Ssw_Sml_t, out aig.Edge) []int32 {
	fanin := p.pAig.Node(out).Fanins[0]
	words := Ssw_ObjSim(p, int32(fanin.NodeID()))
	pattern := -1
	for i, w := range words {
		if w != 0 {
			pattern = i*32 + bits.TrailingZeros32(w)
			break
		}
	}
	if pattern < 0 {
		panic("ssw: output has no detecting pattern")
	}
	inputs := p.pAig.Inputs()
	model := make([]int32, len(inputs)+1)
	for i, obj := range inputs {
		model[i] = sswPatternBit(Ssw_ObjSim(p, int32(obj.NodeID())), pattern)
	}
	model[len(inputs)] = int32(out.NodeID())
	return model
}
func Ssw_SmlCheckOutput(p *Ssw_Sml_t) []int32 {
	outputs := p.pAig.Outputs()
	if len(outputs) == 0 {
		panic("ssw: expected output")
	}
	first := p.pAig.Node(outputs[0]).Fanins[0]
	if p.pAig.Node(first).Phase != first.Complemented() {
		panic("ssw: reference pattern detects output")
	}
	for _, out := range outputs {
		fanin := p.pAig.Node(out).Fanins[0] &^ 1
		if Ssw_SmlObjIsConstWord(p, fanin) == 0 {
			return Ssw_SmlCheckOutputSavePattern(p, out)
		}
	}
	return nil
}
func Ssw_SmlGetCounterExample(p *Ssw_Sml_t) *util.Abc_Cex_t {
	if p.fNonConstOut == 0 {
		panic("ssw: no failing output")
	}
	inputs, outputs := p.pAig.Inputs(), p.pAig.Outputs()
	regs := int(p.pAig.RegisterCount())
	pis := len(inputs) - regs
	outputIndex, frame, bit := -1, -1, -1
	for i, out := range outputs[:len(outputs)-regs] {
		words := Ssw_ObjSim(p, int32(out.NodeID()))
		for w := p.nWordsPref; w < p.nWordsTotal; w++ {
			if words[w] != 0 {
				outputIndex = i
				frame = int(w / p.nWordsFrame)
				bit = 32*int(w%p.nWordsFrame) + bits.TrailingZeros32(words[w])
				break
			}
		}
		if outputIndex >= 0 {
			break
		}
	}
	if outputIndex < 0 {
		panic("ssw: failing simulation has no output bit")
	}
	cex := util.Abc_CexAlloc(int32(regs), int32(pis), int32(frame+1))
	cex.IPo = int32(outputIndex)
	cex.IFrame = int32(frame)
	for i, lo := range inputs[pis:] {
		if sswPatternBit(Ssw_ObjSim(p, int32(lo.NodeID())), bit) != 0 {
			util.Abc_InfoSetBit(cex.PData, int32(i))
		}
	}
	for f := 0; f <= frame; f++ {
		for i, pi := range inputs[:pis] {
			if sswPatternBit(Ssw_ObjSim(p, int32(pi.NodeID())), 32*int(p.nWordsFrame)*f+bit) != 0 {
				util.Abc_InfoSetBit(cex.PData, int32(regs+pis*f+i))
			}
		}
	}
	if saig.Saig_ManVerifyCex(p.pAig, cex) == 0 {
		fmt.Println("Ssw_SmlGetCounterExample(): Counter-example is invalid.")
		return nil
	}
	return cex
}

package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"math/bits"
)

// counterExample reconstructs the first failing simulated trace and verifies it.
func (s *fraSimulation) counterExample() *util.Abc_Cex_t {
	g := s.graph
	regs := int(aig.Aig_ManRegNum(g))
	inputs, outputs := g.Inputs(), g.Outputs()
	pis := len(inputs) - regs
	for po, output := range outputs[:len(outputs)-regs] {
		words := s.nodeWords(output.NodeID())
		for w := s.prefixWords; w < len(words); w++ {
			if words[w] == 0 {
				continue
			}
			frame := w / s.wordsPerFrame
			bit := 32*(w%s.wordsPerFrame) + bits.TrailingZeros32(words[w])
			cex := util.Abc_CexAlloc(int32(regs), int32(pis), int32(frame+1))
			cex.IPo, cex.IFrame = int32(po), int32(frame)
			for i := 0; i < regs; i++ {
				if s.nodeWords(inputs[pis+i].NodeID())[bit/32]&(uint32(1)<<uint(bit%32)) != 0 {
					util.Abc_InfoSetBit(cex.PData, int32(i))
				}
			}
			for f := 0; f <= frame; f++ {
				for i := 0; i < pis; i++ {
					if s.frameWords(inputs[i].NodeID(), f)[bit/32]&(uint32(1)<<uint(bit%32)) != 0 {
						util.Abc_InfoSetBit(cex.PData, int32(regs+f*pis+i))
					}
				}
			}
			if saig.Saig_ManVerifyCex(g, cex) == 0 {
				return nil
			}
			return cex
		}
	}
	return nil
}

// copyFrameCounterExample converts a combinational unrolling model to a trace.
func copyFrameCounterExample(g, frames *aig.Graph, model []int32) *util.Abc_Cex_t {
	if g == nil || frames == nil {
		return nil
	}
	regs := int(aig.Aig_ManRegNum(g))
	pis, pos := len(g.Inputs())-regs, len(g.Outputs())-regs
	numInputs := len(frames.Inputs())
	if regs <= 0 || aig.Aig_ManRegNum(frames) != 0 || pis <= 0 || pos <= 0 || numInputs%pis != 0 || len(frames.Outputs()) != numInputs/pis*pos || len(model) <= numInputs {
		return nil
	}
	for i, output := range frames.Outputs() {
		if int32(output.NodeID()) != model[numInputs] {
			continue
		}
		frame := i / pos
		cex := util.Abc_CexAlloc(int32(regs), int32(pis), int32(frame+1))
		cex.IPo, cex.IFrame = int32(i%pos), int32(frame)
		for j := 0; j < (frame+1)*pis; j++ {
			if model[j] != 0 {
				util.Abc_InfoSetBit(cex.PData, int32(regs+j))
			}
		}
		if saig.Saig_ManVerifyCex(g, cex) == 0 {
			return nil
		}
		return cex
	}
	return nil
}

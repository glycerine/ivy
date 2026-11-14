package gia

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/misc/util"
)

func Gia_Sim2GenerateCounter(g *Gia_Man_t, frame, output, words, pattern int32) *Abc_Cex_t {
	if words <= 0 || pattern < 0 || pattern >= 32*words {
		panic("gia: invalid simulation pattern")
	}
	cex := util.Abc_CexAlloc(Gia_ManRegNum(g), Gia_ManPiNum(g), frame+1)
	cex.IFrame, cex.IPo = frame, output
	data := make([]uint32, words)
	for f := int32(0); f <= frame; f++ {
		for i := int32(0); i < cex.NPis; i++ {
			for w := words - 1; w >= 0; w-- {
				data[w] = Gia_ManRandom(0)
			}
			if data[pattern>>5]&(1<<uint(pattern&31)) != 0 {
				util.Abc_InfoSetBit(cex.PData, cex.NRegs+f*cex.NPis+i)
			}
		}
	}
	return cex
}
func Gia_ManSimSimulateEquiv(g *Gia_Man_t, pars *Gia_ParSim_t) int32 {
	if g.pReprs == nil || g.pNexts == nil {
		panic("gia: equivalence classes required")
	}
	started := time.Now()
	g.pCexSeq = nil
	p := Gia_Sim2Create(g, pars)
	defer Gia_Sim2Delete(p)
	Gia_ManResetRandom(pars)
	for i := int32(0); i < Gia_ManRegNum(g); i++ {
		Gia_Sim2InfoZero(p, Gia_Sim2Data(p, Gia_ManRo(g, i).id))
	}
	var frame, output, pattern int32
	for frame = 0; frame < pars.NIters; frame++ {
		Gia_Sim2SimulateRound(p)
		if pars.FVerbose != 0 {
			fmt.Printf("Frame %4d out of %4d and timeout %3d sec. Lits = %4d. Time = %7.2f sec\r", frame+1, pars.NIters, pars.TimeLimit, Gia_ManEquivCountLitsAll(g), time.Since(started).Seconds())
		}
		if pars.FCheckMiter != 0 && Gia_Sim2CheckPos(p, &output, &pattern) != 0 {
			Gia_ManResetRandom(pars)
			pars.iOutFail = output
			g.pCexSeq = Gia_Sim2GenerateCounter(g, frame, output, p.nWords, pattern)
			fmt.Printf("Output %d of miter %q was asserted in frame %d.  ", output, g.pName, frame)
			if Gia_ManVerifyCex(g, g.pCexSeq, 0) == 0 {
				fmt.Print("\nGenerated counter-example is INVALID.                    ")
			}
			fmt.Printf("Time: %s\n", time.Since(started))
			return 1
		}
		Gia_Sim2InfoRefineEquivs(p)
		if pars.TimeLimit > 0 && time.Since(started) >= time.Duration(pars.TimeLimit)*time.Second {
			frame++
			break
		}
		if frame < pars.NIters-1 {
			Gia_Sim2InfoTransfer(p)
		}
	}
	fmt.Printf("No bug detected after simulating %d frames with %d words.  Time: %s\n", frame, pars.NWords, time.Since(started))
	return 0
}

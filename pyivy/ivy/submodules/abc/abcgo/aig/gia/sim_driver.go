package gia

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManResetRandom(pars *Gia_ParSim_t) {
	Gia_ManRandom(1)
	for i := int32(0); i < pars.RandSeed; i++ {
		Gia_ManRandom(0)
	}
}

// Gia_ManGenerateCounter replays the seeded random inputs for the selected bit.
func Gia_ManGenerateCounter(g *Gia_Man_t, frame, output, words, pattern int32, ciIDs *Vec_Int_t) *Abc_Cex_t {
	if words <= 0 || pattern < 0 || pattern >= 32*words {
		panic("gia: invalid simulation pattern")
	}
	cex := util.Abc_CexAlloc(Gia_ManRegNum(g), Gia_ManPiNum(g), frame+1)
	cex.IFrame, cex.IPo = frame, output
	data := make([]uint32, words)
	for f := int32(0); f <= frame; f++ {
		for i := int32(0); i < cex.NPis; i++ {
			id := vec.Vec_IntEntry(ciIDs, int(i))
			if id >= cex.NPis {
				continue
			}
			for w := words - 1; w >= 0; w-- {
				data[w] = Gia_ManRandom(0)
			}
			if data[pattern>>5]&(1<<uint(pattern&31)) != 0 {
				util.Abc_InfoSetBit(cex.PData, cex.NRegs+f*cex.NPis+id)
			}
		}
	}
	return cex
}

// giaSimSimulate runs the original random-frame driver; equivalence-class
// simulation is supplied by the separately implemented equivalence engine.
func giaSimSimulate(g *Gia_Man_t, pars *Gia_ParSim_t, equiv func(*Gia_Man_t, *Gia_ParSim_t) int32) int32 {
	started := time.Now()
	if g.pReprs != nil && g.pNexts != nil {
		return equiv(g, pars)
	}
	g.pCexSeq = nil
	sim := Gia_ManSimCreate(g, pars)
	defer Gia_ManSimDelete(sim)
	Gia_ManResetRandom(pars)
	Gia_ManSimInfoInit(sim)
	var frame, output, pattern int32
	for frame = 0; frame < pars.NIters; frame++ {
		Gia_ManSimulateRound(sim)
		if pars.FVerbose != 0 {
			fmt.Printf("Frame %4d out of %4d and timeout %3d sec. Time = %7.2f sec\r", frame+1, pars.NIters, pars.TimeLimit, time.Since(started).Seconds())
		}
		if pars.FCheckMiter != 0 && Gia_ManCheckPos(sim, &output, &pattern) != 0 {
			Gia_ManResetRandom(pars)
			pars.iOutFail = output
			g.pCexSeq = Gia_ManGenerateCounter(g, frame, output, sim.nWords, pattern, sim.vCis2Ids)
			fmt.Printf("Output %d of miter %q was asserted in frame %d.  ", output, g.pName, frame)
			if Gia_ManVerifyCex(g, g.pCexSeq, 0) == 0 {
				fmt.Print("\nGenerated counter-example is INVALID.                    ")
			}
			fmt.Printf("Time: %s\n", time.Since(started))
			return 1
		}
		if pars.TimeLimit > 0 && time.Since(started) >= time.Duration(pars.TimeLimit)*time.Second {
			frame++
			break
		}
		if frame < pars.NIters-1 {
			Gia_ManSimInfoTransfer(sim)
		}
	}
	fmt.Printf("No bug detected after simulating %d frames with %d words.  Time: %s\n", frame, pars.NWords, time.Since(started))
	return 0
}

// Gia_ManSimSimulate selects class-aware simulation when representatives exist.
func Gia_ManSimSimulate(g *Gia_Man_t, pars *Gia_ParSim_t) int32 {
	return giaSimSimulate(g, pars, Gia_ManSimSimulateEquiv)
}

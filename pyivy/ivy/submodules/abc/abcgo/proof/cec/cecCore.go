// Go port of cecCore.c, by Alan Mishchenko, UC Berkeley (2005).
package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// Cec_ManSimulationOne performs a round of candidate simulation.
func Cec_ManSimulationOne(graph *Gia_Man_t, pars *Cec_ParSim_t) int32 {
	started := util.Abc_Clock()
	sim := Cec_ManSimStart(graph, pars)
	defer Cec_ManSimStop(sim)
	var result int32
	if !graph.HasRepresentatives() {
		result = Cec_ManSimClassesPrepare(sim, -1)
	}
	if result == 0 {
		result = Cec_ManSimClassesRefine(sim)
	}
	if result != 0 {
		util.Abc_Print(1, "The number of failed outputs of the miter = %6d. (Words = %4d. Frames = %4d.)\n", sim.nOuts, pars.nWords, pars.nFrames)
	}
	if pars.fVerbose != 0 {
		util.Abc_PrintTime(1, "Time", util.Abc_Clock()-started)
	}
	return result
}

func Cec_ManSimulation(graph *Gia_Man_t, pars *Cec_ParSim_t) {
	gia.Gia_ManRandom(1)
	if pars.fSeqSimulate != 0 {
		// The C format has two verbs but passes rounds, frames, then words.
		util.Abc_Print(1, "Performing rounds of random simulation of %d frames with %d words.\n", pars.nRounds, pars.nFrames)
	}
	old := gia.Gia_ManEquivCountLits(graph)
	var round, noRefine int32
	stopped := false
	for ; round < pars.nRounds; round++ {
		if Cec_ManSimulationOne(graph, pars) != 0 {
			stopped = true
			break
		}
		next := gia.Gia_ManEquivCountLits(graph)
		if old == 0 || old > next {
			old, noRefine = next, 0
		} else {
			noRefine++
			if noRefine == pars.nNonRefines {
				round++
				break
			}
		}
		if old != next {
			panic("cec: simulation increased equivalence candidates")
		}
	}
	if round == pars.nRounds || stopped {
		util.Abc_Print(1, "Random simulation is stopped after %d rounds.\n", round)
	} else {
		util.Abc_Print(1, "Random simulation saturated after %d rounds.\n", round)
	}
	if pars.fCheckMiter != 0 {
		if n := Cec_ManCountNonConstOutputs(graph); n != 0 {
			util.Abc_Print(1, "The number of POs that are not const-0 candidates = %d.\n", n)
		}
	}
}

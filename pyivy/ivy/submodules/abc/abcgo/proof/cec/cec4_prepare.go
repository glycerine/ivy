package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// cec4RandomInputs installs the random CI patterns used by Cec4; bit zero is
// always the all-zero input so the stored phase remains a valid reference.
func cec4RandomInputs(sim *cec3Simulation, graph *gia.Gia_Man_t) {
	for i := int32(0); i < gia.Gia_ManCiNum(graph); i++ {
		id := gia.Gia_ObjId(graph, gia.Gia_ManCi(graph, i))
		words := sim.object(id)
		for j := range words {
			words[j] = gia.Gia_ManRandomW(0)
		}
		words[0] <<= 1
	}
}

// cec4PrepareSimulation implements the initial and generated-pattern rounds
// preceding the SAT sweep. A returned witness ends a miter run early.
func cec4PrepareSimulation(graph *gia.Gia_Man_t, pars *Cec_ParFra_t, simOnly bool) (*cec4PatternPacker, *util.Abc_Cex_t) {
	if pars.nWords <= 0 || pars.nRounds < 0 || pars.nGenIters < 0 || (!simOnly && pars.nGenIters > 0 && pars.nItersMax <= 0) {
		panic("cec4: invalid simulation parameters")
	}
	gia.Gia_ManRandom(1)
	gia.Gia_ManSetPhase(graph)
	if pars.nLevelMax != 0 {
		gia.Gia_ManLevelNum(graph)
	}
	if pars.fCheckMiter != 0 {
		for i := int32(0); i < gia.Gia_ManCoNum(graph); i++ {
			if gia.Gia_ObjPhase(gia.Gia_ManCo(graph, i)) != 0 {
				cex := util.Abc_CexAlloc(0, gia.Gia_ManCiNum(graph), 1)
				cex.IPo = i
				return nil, cex
			}
		}
	}
	pack := newCec4PatternPacker(int(gia.Gia_ManObjNum(graph)), int(pars.nWords))
	for round := int32(0); round <= pars.nRounds; round++ {
		cec4RandomInputs(pack.values, graph)
		cec4SimulationRound(pack.values, graph, pars.nLevelMax, round != 0)
		if pars.fCheckMiter != 0 {
			if _, cex := cec4FirstFailingOutput(pack.values, graph); cex != nil {
				return pack, cex
			}
		}
	}
	if simOnly {
		return pack, nil
	}
	candidates := newCec4Candidates(graph)
	keepGenerating := true
	for round := int32(0); keepGenerating && round < pars.nGenIters; round++ {
		cec4RandomInputs(pack.values, graph)
		_, keepGenerating = cec4GeneratePatterns(graph, pack, candidates, pars.nItersMax)
		cec4SimulationRound(pack.values, graph, pars.nLevelMax, true)
		if pars.fCheckMiter != 0 {
			if _, cex := cec4FirstFailingOutput(pack.values, graph); cex != nil {
				return pack, cex
			}
		}
	}
	return pack, nil
}

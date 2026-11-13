package cec

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
)

// cecOriginalSweep is the speculative-reduction loop from cecCore.c. It is
// deliberately separate from the public sweep entry until the full source
// interface, including original-ID transfer, is ported.
func cecOriginalSweep(graph *Gia_Man_t, pars *Cec_ParFra_t) (*Gia_Man_t, error) {
	if pars.fUseOrigIds != 0 {
		return nil, fmt.Errorf("cec: original sweep metadata transfer is not implemented")
	}
	if words := graph.SimulationWordCount(); words > 0 && len(graph.SimulationInputs()) != int(gia.Gia_ManCiNum(graph)*words) {
		return nil, fmt.Errorf("cec: invalid original sweep input simulation size")
	}
	if pars.nWords <= 0 || pars.nRounds < 0 || pars.nItersMax < 0 || pars.nBTLimit < 0 || pars.TimeLimit < 0 {
		return nil, fmt.Errorf("cec: invalid original sweep parameters")
	}
	for id := int32(0); int(id) < graph.ObjectCount(); id++ {
		_, _, _, kind := graph.Gate(id)
		if kind != 0 && kind != 1 && kind != 2 && kind != 6 {
			return nil, fmt.Errorf("cec: original sweep requires AND-inverter graph (object %d has kind %d)", id, kind)
		}
	}
	graph.ClearCombinationalCounterexample()
	pars.iOutFail = -1
	started := time.Now()
	deadline := time.Time{}
	if pars.TimeLimit > 0 {
		deadline = started.Add(time.Duration(pars.TimeLimit) * time.Second)
	}
	gia.Gia_ManRandom(1)
	initial := gia.Gia_ManDup(graph)
	if words := graph.SimulationWordCount(); words > 0 {
		gia.Gia_ManCecSimAlloc(initial, words)
		for _, word := range graph.SimulationInputs() {
			gia.Gia_ManCecSaveCIWord(initial, word)
		}
	}
	if graph.HasRepresentatives() {
		gia.Gia_ManTransferEquivs(graph, initial)
	}
	fra := Cec_ManFraStart(initial, pars)
	defer Cec_ManFraStop(fra)
	if pars.fDualOut != 0 {
		pars.fColorDiff = 1
	}
	var simPars Cec_ParSim_t
	Cec_ManSimSetDefaultParams(&simPars)
	simPars.nWords = max(pars.nWords, 2*graph.SimulationWordCount())
	simPars.nFrames = pars.nRounds
	simPars.fCheckMiter, simPars.fDualOut, simPars.fVerbose = pars.fCheckMiter, pars.fDualOut, pars.fVerbose
	simPars.deadline = deadline
	sim := Cec_ManSimStart(initial, &simPars)
	defer Cec_ManSimStop(sim)
	var satPars Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&satPars)
	satPars.nBTLimit, satPars.fVerbose, satPars.deadline = pars.nBTLimit, pars.fVeryVerbose, deadline
	pat := Cec_ManPatStart()
	defer Cec_ManPatStop(pat)
	failure := false
	if !initial.HasRepresentatives() {
		failure = Cec_ManSimClassesPrepare(sim, -1) != 0
		if !failure && graph.SimulationWordCount() == 0 {
			failure = Cec_ManSimClassesRefine(sim) != 0
		}
	}
	for i := int32(1); !failure && i <= pars.nItersMax; i++ {
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			break
		}
		if pars.fDualOut != 0 {
			gia.Gia_ManEquivSetColors(fra.pAig, pars.fVeryVerbose)
		}
		queries := Cec_ManFraSpecReduction(fra)
		if gia.Gia_ManCoNum(queries) == 0 {
			gia.Gia_ManStop(queries)
			if pars.fDualOut != 0 && gia.Gia_ManAndNum(fra.pAig) > 0 {
				if pars.fColorDiff != 0 {
					pars.fColorDiff = 0
				} else {
					pars.fDualOut = 0
				}
				continue
			}
			break
		}
		if pars.fRunCSat != 0 {
			Cec_ManSatSolveCSat(pat, queries, &satPars)
		} else {
			Cec_ManSatSolve(pat, queries, &satPars, nil, fra.vXorNodes, nil, 0)
		}
		failure = Cec_ManFraClassesUpdate(fra, sim, pat, queries) != 0
		gia.Gia_ManStop(queries)
		if failure {
			break
		}
		old := fra.pAig
		reduced := gia.Gia_ManEquivReduceAndRemap(old, 0, simPars.fDualOut)
		if reduced == nil {
			break
		}
		fra.pAig, sim.pAig = reduced, reduced
		gia.Gia_ManStop(old)
		if gia.Gia_ManAndNum(reduced) == 0 {
			break
		}
		if fra.nAllFailed > fra.nAllProved+fra.nAllDisproved {
			if satPars.nBTLimit >= 10001 || pars.fSatSweeping != 0 {
				break
			}
			satPars.nBTLimit *= 10
		}
		if pars.fDualOut != 0 && pars.fColorDiff != 0 && (gia.Gia_ManAndNum(reduced) < 100000 || fra.nAllProved+fra.nAllDisproved < 10) {
			pars.fColorDiff = 0
		} else if pars.fDualOut != 0 && (gia.Gia_ManAndNum(reduced) < 20000 || fra.nAllProved+fra.nAllDisproved < 10) {
			pars.fColorDiff, pars.fDualOut = 0, 0
		}
	}
	if failure {
		if sim.pCexComb == nil || sim.iOut < 0 {
			return nil, fmt.Errorf("cec: simulation failed without miter witness")
		}
		witness := sim.pCexComb
		values := make([]bool, gia.Gia_ManCiNum(graph))
		for i := range values {
			values[i] = witness.PData[i>>5]&(uint32(1)<<uint(i&31)) != 0
		}
		outputs, err := graph.Evaluate(values)
		if err != nil {
			return nil, fmt.Errorf("cec: replay original sweep witness: %w", err)
		}
		output := int(sim.iOut)
		if simPars.fDualOut != 0 {
			if output < 0 || 2*output+1 >= len(outputs) || outputs[2*output] == outputs[2*output+1] {
				return nil, fmt.Errorf("cec: original sweep witness does not disprove output pair %d", output)
			}
		} else if output < 0 || output >= len(outputs) || !outputs[output] {
			return nil, fmt.Errorf("cec: original sweep witness does not disprove output %d", output)
		}
		pars.iOutFail = sim.iOut
		if err := graph.SetCombinationalCounterexample(values, int(sim.iOut)); err != nil {
			return nil, err
		}
		gia.Gia_ManStop(fra.pAig)
		return nil, nil
	}
	// The C routine frees the temporary representative and next arrays before
	// returning. Gia_ManDup retains the graph but intentionally omits them.
	result := gia.Gia_ManDup(fra.pAig)
	gia.Gia_ManStop(fra.pAig)
	return result, nil
}

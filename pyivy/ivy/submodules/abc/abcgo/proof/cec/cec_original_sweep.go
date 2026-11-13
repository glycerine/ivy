package cec

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// cecOriginalSweep is the speculative-reduction loop from cecCore.c.
func cecOriginalSweep(graph *Gia_Man_t, pars *Cec_ParFra_t) (*Gia_Man_t, error) {
	return cecOriginalSweepMapped(graph, graph, nil, nil, pars)
}

func cecOriginalSweepMapped(graph, original *Gia_Man_t, sourceIDs []int32, aliases *Vec_Int_t, pars *Cec_ParFra_t) (*Gia_Man_t, error) {
	now, release := cecSweepSourceClock()
	defer release()
	if now == nil {
		if pars.TimeLimit > 0 {
			return nil, fmt.Errorf("cec: source timeout clock unavailable on this platform")
		}
		// Untimed sweeps never inspect the reading after initialization.
		now = func() time.Time { return time.Time{} }
	}
	return cecOriginalSweepMappedClock(graph, original, sourceIDs, aliases, pars, now)
}

// now is injectable so the source's iteration-boundary timeout is testable.
func cecOriginalSweepMappedClock(graph, original *Gia_Man_t, sourceIDs []int32, aliases *Vec_Int_t, pars *Cec_ParFra_t, now func() time.Time) (*Gia_Man_t, error) {
	if graph == nil {
		return nil, fmt.Errorf("cec: nil sweep graph")
	}
	if words := graph.SimulationWordCount(); words > 0 && len(graph.SimulationInputs()) != int(gia.Gia_ManCiNum(graph)*words) {
		return nil, fmt.Errorf("cec: invalid original sweep input simulation size")
	}
	if pars.nWords <= 0 || pars.nRounds < 0 || pars.nItersMax < 0 || pars.nBTLimit < 0 || pars.TimeLimit < 0 {
		return nil, fmt.Errorf("cec: invalid original sweep parameters")
	}
	if pars.fDualOut != 0 && gia.Gia_ManPoNum(graph)&1 != 0 {
		return nil, fmt.Errorf("cec: dual-output sweep requires output pairs")
	}
	for id := int32(0); int(id) < graph.ObjectCount(); id++ {
		_, _, _, kind := graph.Gate(id)
		if kind != 0 && kind != 1 && kind != 2 && kind != 6 {
			return nil, fmt.Errorf("cec: original sweep requires AND-inverter graph (object %d has kind %d)", id, kind)
		}
	}
	graph.ClearCombinationalCounterexample()
	if original != graph {
		original.ClearCombinationalCounterexample()
	}
	started := now()
	gia.Gia_ManRandom(1)
	initial := gia.Gia_ManDup(graph)
	if pars.fUseOrigIds != 0 {
		if sourceIDs == nil {
			gia.Gia_ManOrigIdsInit(initial)
		} else {
			gia.Gia_ManOrigIdsSet(initial, sourceIDs)
		}
		gia.Gia_ManOrigEquivPairsReset(original)
	}
	var candidates *Vec_Int_t
	if pars.fUseOrigIds != 0 {
		candidates = vec.Vec_IntAlloc(1000)
		if aliases != nil {
			vec.Vec_IntAppend(candidates, aliases)
		}
	}
	if words := graph.SimulationWordCount(); words > 0 {
		gia.Gia_ManCecSimAlloc(initial, words)
		for _, word := range graph.SimulationInputs() {
			gia.Gia_ManCecSaveCIWord(initial, word)
		}
	}
	if graph.HasRepresentatives() {
		gia.Gia_ManTransferEquivs(graph, initial)
		graph.ClearRepresentatives()
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
	sim := Cec_ManSimStart(initial, &simPars)
	defer Cec_ManSimStop(sim)
	var satPars Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&satPars)
	satPars.nBTLimit, satPars.fVerbose = pars.nBTLimit, pars.fVeryVerbose
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
			Cec_ManSatSolve(pat, queries, &satPars, gia.Gia_ManOrigIds(fra.pAig), fra.vXorNodes, candidates, 0)
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
		// C checks elapsed whole seconds only after reduction, before adapting
		// the conflict limit or dual-output mode.
		if pars.TimeLimit > 0 && now().Sub(started)/time.Second >= time.Duration(pars.TimeLimit) {
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
	if candidates != nil {
		vec.Vec_IntAppend(gia.Gia_ManOrigEquivPairs(original), candidates)
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
		pars.iOutFail = sim.iOut
		if err := original.SetCombinationalCounterexample(values, int(sim.iOut)); err != nil {
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

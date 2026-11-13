package cec

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
)

type cecSweepRunner func(*Gia_Man_t, *Cec_ParFra_t) (*Gia_Man_t, error)
type cecFallbackRunner func(*Gia_Man_t, int32) (int32, int32, error)

// cecVerifyPipeline follows cecCec.c: normalize output pairs, sweep, then
// invoke the FRA engine on any unresolved graph. Runner arguments make the
// unresolved branch testable without changing production solver state.
func cecVerifyPipeline(graph *Gia_Man_t, pars *Cec_ParCec_t, sweep cecSweepRunner, fallback cecFallbackRunner) int32 {
	started := time.Now()
	if graph.RegisterCount() != 0 {
		pars.Error = fmt.Errorf("cec: verification requires a combinational miter")
		return -1
	}
	work := gia.Gia_ManDup(graph)
	// The source sweep works on AND-inverter graphs. Lower native GIA nodes
	// through ABC's existing AIG conversion, preserving the CI/CO interface.
	for id := int32(0); int(id) < work.ObjectCount(); id++ {
		_, _, _, kind := work.Gate(id)
		if kind == 3 || kind == 4 || kind == 5 {
			lowered := gia.Gia_ManFromAig(gia.Gia_ManToAig(work, 0))
			gia.Gia_ManStop(work)
			work = lowered
			break
		}
	}
	gia.Gia_ManEquivFixOutputPairs(work)
	clean := gia.Gia_ManCleanup(work)
	gia.Gia_ManStop(work)
	work = clean
	if pars.fNaive != 0 {
		status := Cec_ManVerifyNaive(work, pars)
		if status == 0 && work.CombinationalCounterexample() != nil {
			gia.SwapCombinationalCounterexamples(graph, work)
		}
		gia.Gia_ManStop(work)
		return status
	}
	if words := graph.SimulationWordCount(); words > 0 {
		data := graph.SimulationInputs()
		if len(data) != int(gia.Gia_ManCiNum(graph)*words) {
			gia.Gia_ManStop(work)
			pars.Error = fmt.Errorf("cec: invalid input simulation size")
			return -1
		}
		gia.Gia_ManCecSimAlloc(work, words)
		for _, word := range data {
			gia.Gia_ManCecSaveCIWord(work, word)
		}
	}
	var sweepPars Cec_ParFra_t
	Cec_ManFraSetDefaultParams(&sweepPars)
	sweepPars.nItersMax = 1000
	sweepPars.nBTLimit = pars.nBTLimit
	sweepPars.TimeLimit = pars.TimeLimit
	sweepPars.fVerbose = pars.fVerbose
	sweepPars.fVeryVerbose = pars.fVeryVerbose
	sweepPars.fCheckMiter = 1
	sweepPars.fDualOut = 1
	reduced, err := sweep(work, &sweepPars)
	pars.iOutFail = sweepPars.iOutFail
	if err != nil {
		gia.Gia_ManStop(work)
		pars.Error = err
		return -1
	}
	if reduced == nil && work.CombinationalCounterexample() != nil {
		// The internal sweep has replayed its model on work; repeat on the
		// original graph before transferring ownership.
		w := work.CombinationalCounterexample()
		values := make([]bool, len(graph.InputLiterals()))
		for i := range values {
			values[i] = w.PData[i>>5]&(uint32(1)<<uint(i&31)) != 0
		}
		outputs, e := graph.Evaluate(values)
		index := int(pars.iOutFail)
		if e != nil || index < 0 || 2*index+1 >= len(outputs) || outputs[2*index] == outputs[2*index+1] {
			gia.Gia_ManStop(work)
			pars.Error = fmt.Errorf("cec: sweep witness fails original miter replay")
			return -1
		}
		gia.SwapCombinationalCounterexamples(graph, work)
		gia.Gia_ManStop(work)
		return 0
	}
	gia.Gia_ManStop(work)
	if reduced == nil {
		work = gia.Gia_ManDup(graph)
		gia.Gia_ManEquivFixOutputPairs(work)
		reduced = gia.Gia_ManCleanup(work)
		gia.Gia_ManStop(work)
	}
	defer gia.Gia_ManStop(reduced)
	if pars.TimeLimit > 0 && time.Since(started) >= time.Duration(pars.TimeLimit)*time.Second {
		return -1
	}
	status, output, err := fallback(reduced, pars.fVerbose)
	if err != nil {
		pars.Error = err
		return -1
	}
	pars.iOutFail = output
	if status == 0 {
		w := reduced.CombinationalCounterexample()
		if w == nil {
			pars.Error = fmt.Errorf("cec: FRA fallback omitted counterexample")
			return -1
		}
		values := make([]bool, len(graph.InputLiterals()))
		for i := range values {
			values[i] = w.PData[i>>5]&(uint32(1)<<uint(i&31)) != 0
		}
		outputs, e := graph.Evaluate(values)
		index := int(output)
		if e != nil || index < 0 || 2*index+1 >= len(outputs) || outputs[2*index] == outputs[2*index+1] {
			pars.Error = fmt.Errorf("cec: FRA witness fails original miter replay")
			return -1
		}
		gia.SwapCombinationalCounterexamples(graph, reduced)
	}
	return status
}

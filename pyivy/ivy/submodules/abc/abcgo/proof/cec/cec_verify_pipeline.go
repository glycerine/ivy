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
	if err := cecRequireAIG(graph); err != nil {
		pars.Error = err
		return -1
	}
	work := gia.Gia_ManDup(graph)
	gia.Gia_ManEquivFixOutputPairs(work)
	clean := gia.Gia_ManCleanup(work)
	gia.Gia_ManStop(work)
	work = clean
	if pars.FNaive != 0 {
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
	sweepPars.NItersMax = 1000
	sweepPars.NBTLimit = pars.NBTLimit
	sweepPars.TimeLimit = pars.TimeLimit
	sweepPars.FVerbose = pars.FVerbose
	sweepPars.FVeryVerbose = pars.FVeryVerbose
	sweepPars.FCheckMiter = 1
	sweepPars.FDualOut = 1
	reduced, err := sweep(work, &sweepPars)
	pars.iOutFail = sweepPars.iOutFail
	if err != nil {
		gia.Gia_ManStop(work)
		pars.Error = err
		return -1
	}
	// C transfers the witness slot even when it is nil.
	gia.SwapCombinationalCounterexamples(graph, work)
	gia.Gia_ManStop(work)
	if reduced == nil && graph.CombinationalCounterexample() != nil {
		if gia.Gia_ManVerifyCex(graph, graph.CombinationalCounterexample(), 1) == 0 {
			fmt.Println("Counter-example simulation has failed.")
		}
		return 0
	}
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
	status, output, err := fallback(reduced, pars.FVerbose)
	if err != nil {
		pars.Error = err
		return -1
	}
	pars.iOutFail = output
	gia.SwapCombinationalCounterexamples(graph, reduced)
	if graph.CombinationalCounterexample() != nil && gia.Gia_ManVerifyCex(graph, graph.CombinationalCounterexample(), 1) == 0 {
		fmt.Println("Counter-example simulation has failed.")
	}
	return status
}

// cecRequireAIG rejects native nodes that the source pipeline does not
// preserve and that the current sweep port cannot process faithfully.
func cecRequireAIG(graph *Gia_Man_t) error {
	if graph == nil {
		return fmt.Errorf("cec: nil graph")
	}
	for id := int32(0); int(id) < graph.ObjectCount(); id++ {
		_, _, _, kind := graph.Gate(id)
		if kind == 3 || kind == 4 || kind == 5 {
			return fmt.Errorf("cec: unsupported native gate kind %d at object %d", kind, id)
		}
	}
	return nil
}

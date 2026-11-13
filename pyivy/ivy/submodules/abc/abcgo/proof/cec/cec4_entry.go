package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/sat/glucose2"
	"time"
)

// Cec4_ManPerformSweeping simulates candidates and optionally performs the
// Glucose-backed SAT sweep, returning zero when the miter has a witness.
func Cec4_ManPerformSweeping(graph *Gia_Man_t, pars *Cec_ParFra_t, out **Gia_Man_t, simOnly int32) int32 {
	manager := Cec4_ManCreate(graph, pars)
	defer Cec4_ManDestroy(manager)
	if pars.fVerbose != 0 {
		fmt.Printf("Solver type = %d. Simulate %d words in %d rounds. SAT with %d confs. Recycle after %d SAT calls.\n", pars.jType, pars.nWords, pars.nRounds, pars.nBTLimit, pars.nCallsRecycle)
	}
	result, witness, stats := cec4Sweep(graph, pars, simOnly != 0, out != nil, func(s *glucose2.SimpSolver) { manager.pSat = s })
	manager.nSatSat = int32(stats.sat)
	manager.nSatUnsat = int32(stats.unsat)
	manager.nSatUndec = int32(stats.unknown)
	manager.nRecycles = int32(stats.recycles)
	if witness != nil {
		if err := graph.SetCombinationalCounterexampleInSequentialSlot(witness); err != nil {
			panic(err)
		}
	}
	if out != nil && result != nil {
		*out = result
	}
	if graph.SequentialCounterexample() != nil {
		return 0
	}
	return 1
}
func Cec4_ManSimulateTest(graph *Gia_Man_t, pars *Cec_ParFra_t) *Gia_Man_t {
	var result *Gia_Man_t
	Cec4_ManPerformSweeping(graph, pars, &result, 0)
	return result
}
func Cec4_ManSimulateTest2(graph *Gia_Man_t, verbose int32) {
	start := time.Now()
	var pars Cec_ParFra_t
	Cec4_ManSetParams(&pars)
	Cec4_ManPerformSweeping(graph, &pars, nil, 0)
	pars.fVerbose = verbose
	if verbose != 0 {
		fmt.Printf("New choice computation time: %s\n", time.Since(start))
	}
}
func Cec4_ManSimulateTest3(graph *Gia_Man_t, limit, verbose int32) *Gia_Man_t {
	var pars Cec_ParFra_t
	Cec4_ManSetParams(&pars)
	pars.fVerbose = verbose
	pars.nBTLimit = limit
	return Cec4_ManSimulateTest(graph, &pars)
}
func Cec4_ManSimulateOnlyTest(graph *Gia_Man_t, verbose int32) int32 {
	var pars Cec_ParFra_t
	Cec4_ManSetParams(&pars)
	pars.fVerbose = verbose
	return Cec4_ManPerformSweeping(graph, &pars, nil, 1)
}

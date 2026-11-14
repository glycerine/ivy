package cec

import (
	"fmt"
	"time"
)

// Cec4_ManPerformSweeping simulates candidates and optionally performs the
// Glucose-backed SAT sweep, returning zero when the miter has a witness.
func Cec4_ManPerformSweeping(graph *Gia_Man_t, pars *Cec_ParFra_t, out **Gia_Man_t, simOnly int32) int32 {
	unpin := cecPinClockThread()
	defer unpin()
	manager := Cec4_ManCreate(graph, pars)
	defer Cec4_ManDestroy(manager)
	if pars.FVerbose != 0 {
		fmt.Printf("Solver type = %d. Simulate %d words in %d rounds. SAT with %d confs. Recycle after %d SAT calls.\n", pars.JType, pars.NWords, pars.NRounds, pars.NBTLimit, pars.NCallsRecycle)
	}
	result, witness, _ := cec4Sweep(graph, pars, simOnly != 0, out != nil, manager)
	if pars.FVerbose != 0 {
		proofs, disproofs := manager.nSatUnsat, manager.nSatSat
		pctXor := 100 * float64(manager.nGates[1]) / float64(max(int32(1), manager.nGates[0]+manager.nGates[1]))
		fmt.Printf("SAT calls = %d:  P = %d (0=%d a=%.2f m=%d)  D = %d (0=%d a=%.2f m=%d)  F = %d   Sim = %d  Recyc = %d  Xor = %.2f %%\n",
			proofs+disproofs+manager.nSatUndec, proofs, manager.nConflicts[1][0], float64(manager.nConflicts[1][1])/float64(max(int32(1), proofs-manager.nConflicts[1][0])), manager.nConflicts[1][2],
			disproofs, manager.nConflicts[0][0], float64(manager.nConflicts[0][1])/float64(max(int32(1), disproofs-manager.nConflicts[0][0])), manager.nConflicts[0][2],
			manager.nSatUndec, manager.nSimulates, manager.nRecycles, pctXor)
	}
	if witness != nil {
		if err := graph.SetCecCombinationalWitness(witness); err != nil {
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
	unpin := cecPinClockThread()
	defer unpin()
	start := cecCPUClock()
	var pars Cec_ParFra_t
	Cec4_ManSetParams(&pars)
	Cec4_ManPerformSweeping(graph, &pars, nil, 0)
	pars.FVerbose = verbose
	if verbose != 0 {
		fmt.Printf("New choice computation time: %s\n", time.Duration(cecCPUClock()-start))
	}
}
func Cec4_ManSimulateTest3(graph *Gia_Man_t, limit, verbose int32) *Gia_Man_t {
	var pars Cec_ParFra_t
	Cec4_ManSetParams(&pars)
	pars.FVerbose = verbose
	pars.NBTLimit = limit
	return Cec4_ManSimulateTest(graph, &pars)
}
func Cec4_ManSimulateOnlyTest(graph *Gia_Man_t, verbose int32) int32 {
	var pars Cec_ParFra_t
	Cec4_ManSetParams(&pars)
	pars.FVerbose = verbose
	return Cec4_ManPerformSweeping(graph, &pars, nil, 1)
}

package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
)

// CecG_ManSatSolve uses the Glucose2 modes selected by SolverType.
func CecG_ManSatSolve(patterns *Cec_ManPat_t, graph *Gia_Man_t, pars *Cec_ParSat_t, proveZero int32) {
	graph.ClearSequentialCounterexamples()
	if pars.SolverType != 0 {
		pars.fPolarFlip = 0
	}
	cecResetPatterns(patterns)
	gia.Gia_ManSetPhase(graph)
	gia.Gia_ManLevelNum(graph)
	gia.Gia_ManIncrementTravId(graph)
	sat := newCecGeneralSolver(graph, pars)
	for i, lit := range graph.OutputLiterals() {
		var status int32
		if lit == 0 {
			status = 1
		} else if lit == 1 {
			status = 0
		} else {
			status = sat.check(lit)
		}
		cecSetOutputStatus(gia.Gia_ManCo(graph, int32(i)), status)
		if lit > 1 && proveZero != 0 && status == 1 {
			gia.Gia_ManPatchCoDriver(graph, int32(i), 0)
		}
		if lit > 1 && status == 0 && pars.FCheckMiter != 0 {
			break
		}
	}

	fmt.Printf("Recycles %d\n", sat.recycles)
	if pars.FVerbose != 0 {
		sat.printStats()
	}
}

package cec

import "github.com/glycerine/abcgo/aig/gia"

// CecG_ManSatSolve follows the general-solver output marking and zero-patching
// workflow. The shared Go proof adapter currently supplies the SAT backend for
// all SolverType values; it does not reproduce the C solver's backend tuning.
func CecG_ManSatSolve(patterns *Cec_ManPat_t, graph *Gia_Man_t, pars *Cec_ParSat_t, proveZero int32) {
	graph.ClearSequentialCounterexamples()
	if pars.SolverType != 0 {
		pars.fPolarFlip = 0
	}
	cecResetPatterns(patterns)
	gia.Gia_ManSetPhase(graph)
	gia.Gia_ManLevelNum(graph)
	gia.Gia_ManIncrementTravId(graph)
	sat := Cec_ManSatCreate(graph, pars)
	defer Cec_ManSatStop(sat)
	for i, lit := range graph.OutputLiterals() {
		var status int32
		if lit == 0 {
			status = 1
		} else if lit == 1 {
			status = 0
		} else {
			status = sat.query(lit, nil)
		}
		cecSetOutputStatus(gia.Gia_ManCo(graph, int32(i)), status)
		if lit > 1 && proveZero != 0 && status == 1 {
			gia.Gia_ManPatchCoDriver(graph, int32(i), 0)
		}
		if lit > 1 && status == 0 && pars.fCheckMiter != 0 {
			break
		}
	}
	if pars.fVerbose != 0 {
		Cec_ManSatPrintStats(sat)
	}
}

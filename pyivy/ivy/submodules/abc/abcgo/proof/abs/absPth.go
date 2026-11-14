// Go port of absPth.c by Alan Mishchenko, UC Berkeley (2005).
package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/ssw"
)

// Abs_CallBackToStop reports whether a newer prover superseded this run.
func Abs_CallBackToStop(runID int32) int32 {
	if absProver.stopped(runID) {
		return 1
	}
	return 0
}

// Gia_GlaProveAbsracted starts PDR on a private copy of the gate abstraction.
// The spelling is retained from the ABC API.
func Gia_GlaProveAbsracted(graph *Gia_Man_t, simplify, verbose int32) {
	if graph == nil || gia.Gia_ManGateClasses(graph) == nil {
		panic("abs: missing gate abstraction")
	}
	abstraction := Gia_ManDupAbsGates(graph, gia.Gia_ManGateClasses(graph))
	gia.Gia_ManCleanValue(graph)
	aig := absGiaToAigSimple(abstraction)
	if simplify != 0 {
		var pars ssw.Ssw_Pars_t
		ssw.Ssw_ManSetDefaultParams(&pars)
		ssw.Ssw_ParsSetFramesK(&pars, 4)
		aig = ssw.Ssw_SignalCorrespondence(aig, &pars)
	}
	absProver.start(func(id int32) int32 {
		return absRunPDR(id, aig, verbose)
	})
}

func Gia_GlaProveCancel(_ int32)      { absProver.cancel() }
func Gia_GlaProveCheck(_ int32) int32 { return absProver.check() }

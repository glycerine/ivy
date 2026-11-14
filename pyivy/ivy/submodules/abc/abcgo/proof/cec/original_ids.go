package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
)

// Gia_ManComputeGiaEquivs sweeps the graph, then rebuilds it from the proved
// original-object pairs. It lives in CEC to avoid a GIA-to-CEC import cycle.
func Gia_ManComputeGiaEquivs(graph *gia.Gia_Man_t, conflicts, verbose int32) *gia.Gia_Man_t {
	var pars Cec_ParFra_t
	Cec_ManFraSetDefaultParams(&pars)
	pars.NItersMax = 100
	pars.fUseOrigIds = 1
	pars.fSatSweeping = 1
	pars.NBTLimit = conflicts
	pars.FVerbose = verbose
	swept := Cec_ManSatSweeping(graph, &pars, 0)
	if swept == nil {
		return nil
	}
	gia.Gia_ManStop(swept)
	pairs := gia.Gia_ManOrigEquivPairs(graph)
	return gia.Gia_ManOrigIdsReduce(graph, pairs)
}

// InstallMiniLutNameMapping connects the frame's GIA bridge to SAT sweeping.
func InstallMiniLutNameMapping(frame *abcmain.Frame) {
	gia.InstallFrameEquivalenceComputer(frame, Gia_ManComputeGiaEquivs)
}

package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManComputeGiaEquivs sweeps the graph, then rebuilds it from the proved
// original-object pairs. It lives in CEC to avoid a GIA-to-CEC import cycle.
func Gia_ManComputeGiaEquivs(graph *gia.Gia_Man_t, conflicts, verbose int32) *gia.Gia_Man_t {
	var pars Cec_ParFra_t
	Cec_ManFraSetDefaultParams(&pars)
	pars.nItersMax = 100
	pars.fUseOrigIds = 1
	pars.fSatSweeping = 1
	pars.nBTLimit = conflicts
	pars.fVerbose = verbose
	swept := Cec_ManSatSweeping(graph, &pars, 0)
	if swept == nil {
		return nil
	}
	gia.Gia_ManStop(swept)
	pairs := vec.Vec_IntAlloc(100)
	for id := int32(0); id < gia.Gia_ManObjNum(graph); id++ {
		repr := gia.Gia_ObjRepr(graph, id)
		if repr != gia.GIA_VOID {
			vec.Vec_IntPush(pairs, repr)
			vec.Vec_IntPush(pairs, id)
		}
	}
	graph.ClearRepresentatives()
	return gia.Gia_ManOrigIdsReduce(graph, pairs)
}

// InstallMiniLutNameMapping connects the frame's GIA bridge to SAT sweeping.
func InstallMiniLutNameMapping(frame *abcmain.Frame) {
	gia.InstallFrameEquivalenceComputer(frame, Gia_ManComputeGiaEquivs)
}

package cec

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
)

func Cec_LatchCorrespondence(graph *Aig_Man_t, conflicts, useCSat int32) *Aig_Man_t {
	params := new(Cec_ParCor_t)
	Cec_ManCorSetDefaultParams(params)
	params.fLatchCorr = 1
	params.fUseCSat = useCSat
	params.nBTLimit = conflicts
	return cecCorrespondence(graph, params)
}
func Cec_SignalCorrespondence(graph *Aig_Man_t, conflicts, useCSat int32) *Aig_Man_t {
	params := new(Cec_ParCor_t)
	Cec_ManCorSetDefaultParams(params)
	params.fUseCSat = useCSat
	params.nBTLimit = conflicts
	return cecCorrespondence(graph, params)
}
func cecCorrespondence(graph *Aig_Man_t, params *Cec_ParCor_t) *Aig_Man_t {
	converted := gia.Gia_ManFromAigSimple(graph)
	defer gia.Gia_ManStop(converted)
	Cec_ManLSCorrespondenceClasses(converted, params)
	gia.Gia_ManReprToAigRepr(graph, converted)
	return aig.Aig_ManDupRepr(graph, 0)
}

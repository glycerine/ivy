package cec

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"time"
)

func Cec_LatchCorrespondence(graph *Aig_Man_t, conflicts, useCSat int32) *Aig_Man_t {
	return Cec_LatchCorrespondenceDeadline(graph, conflicts, useCSat, time.Time{})
}
func Cec_LatchCorrespondenceDeadline(graph *Aig_Man_t, conflicts, useCSat int32, deadline time.Time) *Aig_Man_t {
	params := new(Cec_ParCor_t)
	Cec_ManCorSetDefaultParams(params)
	params.fLatchCorr = 1
	params.fUseCSat = useCSat
	params.nBTLimit = conflicts
	params.deadline = deadline
	return cecCorrespondence(graph, params)
}
func Cec_SignalCorrespondence(graph *Aig_Man_t, conflicts, useCSat int32) *Aig_Man_t {
	return Cec_SignalCorrespondenceDeadline(graph, conflicts, useCSat, time.Time{})
}
func Cec_SignalCorrespondenceDeadline(graph *Aig_Man_t, conflicts, useCSat int32, deadline time.Time) *Aig_Man_t {
	params := new(Cec_ParCor_t)
	Cec_ManCorSetDefaultParams(params)
	params.fUseCSat = useCSat
	params.nBTLimit = conflicts
	params.deadline = deadline
	return cecCorrespondence(graph, params)
}
func cecCorrespondence(graph *Aig_Man_t, params *Cec_ParCor_t) *Aig_Man_t {
	if cecCorrDeadlineExpired(params) {
		return nil
	}
	converted := gia.Gia_ManFromAigSimple(graph)
	defer gia.Gia_ManStop(converted)
	if cecCorrDeadlineExpired(params) {
		return nil
	}
	if Cec_ManLSCorrespondenceClasses(converted, params) < 0 || cecCorrDeadlineExpired(params) {
		return nil
	}
	gia.Gia_ManReprToAigRepr(graph, converted)
	return aig.Aig_ManDupRepr(graph, 0)
}

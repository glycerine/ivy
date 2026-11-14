package cec

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
)

func Cec_FraigCombinational(graph *Aig_Man_t, conflicts, verbose int32) *Aig_Man_t {
	params := new(Cec_ParFra_t)
	Cec_ManFraSetDefaultParams(params)
	params.fSatSweeping = 1
	params.fUseOrigIds = 1
	params.NBTLimit = conflicts
	params.NItersMax = 20
	params.FVerbose = verbose
	converted := gia.Gia_ManFromAigSimple(graph)
	defer gia.Gia_ManStop(converted)
	reduced := Cec_ManSatSweeping(converted, params, 0)
	if reduced != nil {
		defer gia.Gia_ManStop(reduced)
	}
	if params.Error != nil {
		return nil
	}
	// The source sweep records proved pairs separately; reconstruct the input
	// representatives before transferring them to the AIG manager.
	rebuilt := gia.Gia_ManOrigIdsReduce(converted, gia.Gia_ManOrigEquivPairs(converted))
	gia.Gia_ManStop(rebuilt)
	gia.Gia_ManReprToAigRepr(graph, converted)
	return aig.Aig_ManDupRepr(graph, 0)
}

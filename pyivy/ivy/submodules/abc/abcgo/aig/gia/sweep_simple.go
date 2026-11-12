package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/dch"
)

// Gia_ManFraigSweepPerform computes DCH equivalences and imports them into p.
func Gia_ManFraigSweepPerform(p *Gia_Man_t, parameters any) {
	var params *dch.Params
	if parameters != nil {
		var ok bool
		params, ok = parameters.(*dch.Params)
		if !ok {
			panic("gia: invalid DCH parameters")
		}
	}
	graph := Gia_ManToAigSimple(p)
	defer aig.Aig_ManStop(graph)
	if err := dch.ComputeEquivalences(graph, params); err != nil {
		panic(err)
	}
	Gia_ManReprFromAigRepr(graph, p)
}

func Gia_ManFraigSweepSimple(p *Gia_Man_t, parameters any) *Gia_Man_t {
	if p.pManTime != nil && Gia_ManBoxNum(p) != 0 {
		panic("gia: simple sweep does not support boxes")
	}
	Gia_ManFraigSweepPerform(p, parameters)
	result := Gia_ManEquivReduce(p, 1, 0, 0, 0)
	if result == nil {
		result = Gia_ManDup(p)
	}
	Gia_ManTransferTiming(result, p)
	return result
}

package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/opt/dau"
	"github.com/glycerine/abcgo/proof/dch"
)

// giaAigSynch2 prepares three synthesis variants and computes their choices.
func giaAigSynch2(input *Gia_Man_t, pars *dch.Params, relax int32, dsdBalance func(*Gia_Man_t, int32, int32, int32, int32) *Gia_Man_t) *Gia_Man_t {
	if input == nil || pars == nil {
		panic("gia: choice synthesis needs a graph and parameters")
	}
	var mapping Jf_Par_t
	Lf_ManSetDefaultPars(&mapping)
	mapping.FCutMin, mapping.FCoarsen = 1, 1
	mapping.NRelaxRatio, mapping.NAreaTuner, mapping.NCutNum = relax, 5, 12
	if pars.Verbose {
		mapping.FVerbose = 1
		Gia_ManPrintStats(input, nil)
	}
	first := Gia_ManDup(input)
	if Gia_ManAndNum(first) == 0 {
		Gia_ManTransferTiming(first, input)
		return first
	}
	if first.pManTime != nil && first.vLevels == nil {
		Gia_ManLevelWithBoxes(first)
	}
	if Gia_ManHasMapping(input) != 0 {
		Gia_ManTransferMapping(first, input)
		previous := first
		first = Gia_ManFromAig(dau.Dsm_ManDeriveGia(previous, 0))
		Gia_ManStop(previous)
	}
	second := Gia_ManAreaBalance(first, 0, util.ABC_INFINITY, 0, 0)
	if pars.Verbose {
		Gia_ManPrintStats(second, nil)
	}
	previous := second
	second = Lf_ManPerformMapping(previous, &mapping)
	if pars.Verbose {
		Gia_ManPrintStats(second, nil)
	}
	if previous != second {
		Gia_ManStop(previous)
	}
	var third *Gia_Man_t
	if pars.LightSynth || Gia_ManBufNum(second) != 0 {
		third = Gia_ManAreaBalance(second, 0, util.ABC_INFINITY, 0, 0)
	} else {
		previous = second
		second = Gia_ManAreaBalance(previous, 0, util.ABC_INFINITY, 0, 0)
		if pars.Verbose {
			Gia_ManPrintStats(second, nil)
		}
		Gia_ManStop(previous)
		third = dsdBalance(second, 6, 8, 0, 0)
	}
	if pars.Verbose {
		Gia_ManPrintStats(third, nil)
	}
	barriers := Gia_ManBufNum(input)
	if barriers != 0 {
		variants := []*Gia_Man_t{first, second, third}
		for i, g := range variants {
			if Gia_ManBufNum(g) != barriers {
				panic("gia: synthesis changed the barrier count")
			}
			variants[i] = Gia_ManDupFromBarBufs(g)
			Gia_ManStop(g)
		}
		first, second, third = variants[0], variants[1], variants[2]
	}
	result := Gia_ManAigSynch2Choices(first, second, third, pars)
	Gia_ManStop(first)
	Gia_ManStop(second)
	Gia_ManStop(third)
	if barriers != 0 {
		previous = result
		result = Gia_ManDupToBarBufs(previous, barriers)
		Gia_ManStop(previous)
	}
	result.pName, result.pSpec = input.pName, input.pSpec
	Gia_ManTransferTiming(result, input)
	return result
}

func Gia_ManAigSynch2Choices(first, second, third *Gia_Man_t, pars *dch.Params) *Gia_Man_t {
	if first == nil {
		panic("gia: missing reference graph for choices")
	}
	variants := make([]*Gia_Man_t, 0, 3)
	for _, g := range []*Gia_Man_t{third, second, first} {
		if g != nil {
			variants = append(variants, g)
		}
	}
	miter := Gia_ManChoiceMiter(variants)
	graph := Gia_ManToAigSkip(miter, 3)
	choices, err := dch.Dch_ComputeChoices(graph, pars)
	if err != nil {
		panic(err)
	}
	order := make([]aig.PIO, 0, Gia_ManCiNum(first)+Gia_ManCoNum(first))
	for id := int32(0); id < first.nObjs; id++ {
		obj := Gia_ManObj(first, id)
		if Gia_ObjIsCi(obj) != 0 {
			order = append(order, aig.PIO{Input: true, Index: int(Gia_ObjCioId(obj))})
		} else if Gia_ObjIsCo(obj) != 0 {
			order = append(order, aig.PIO{Input: false, Index: int(Gia_ObjCioId(obj))})
		}
	}
	return Gia_ManFromAigChoices(aig.Aig_ManDupDfsGuided(choices, order))
}

package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/dch"
)

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

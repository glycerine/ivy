package abs

import (
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gla_ManRefinement(p *Gla_Man_t) *vec.Vec_Int_t {
	var cex *util.Abc_Cex_t
	var mapping *vec.Vec_Int_t
	Gia_GlaPrepareCexAndMap(p, &cex, &mapping)
	selected := Rnm_ManRefine(p.pRnm, cex, mapping, p.pPars.fPropFanout, p.pPars.fNewRefine, 1)
	if vec.Vec_IntSize(selected) == 0 {
		p.pGia.ClearSequentialCounterexample()
		witness := Gla_ManDeriveCex(p, mapping)
		if err := p.pGia.SetSequentialCounterexample(witness); err != nil {
			panic(err)
		}
		return nil
	}
	for i, id := range vec.Vec_IntArray(selected) {
		if id < 0 || int(id) >= len(p.pObj2Obj) {
			panic("invalid GLA refinement object")
		}
		vec.Vec_IntWriteEntry(selected, i, int32(p.pObj2Obj[id]))
	}
	p.nObjAdded += int32(vec.Vec_IntSize(selected))
	return selected
}

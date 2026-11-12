package gia

import "github.com/glycerine/abcgo/misc/vec"

func ofDefaultParameters(p *Jf_Par_t) {
	*p = Jf_Par_t{nLutSize: 4, nCutNum: 16, nRounds: 3, nRoundsEla: 4, nCoarseLimit: 3, nAreaTuner: 10, DelayTarget: -1, nDelayLut1: 10, nDelayLut2: 2, fOptEdge: 1, nLutSizeMax: 6, nCutNumMax: 32, MapDelayTarget: -1}
}
func ofMap(g *Gia_Man_t, pars *Jf_Par_t) *Gia_Man_t {
	if g == nil || pars == nil || pars.nRounds < 1 {
		panic("gia: invalid Of mapping request")
	}
	if Gia_ManHasChoices(g) != 0 {
		pars.fCoarsen = 0
		pars.fCutMin = 1
	}
	if pars.fCoarsen != 0 {
		g = Gia_ManDupMuxes(g, pars.nCoarseLimit)
	}
	g.Refs = nil
	g.vCellMapping = nil
	cuts := newOfCutManager(g, pars)
	if cuts.truth != nil {
		defer vec.Vec_MemFree(cuts.truth.memory)
		defer vec.Vec_MemHashFree(cuts.truth.memory)
	}
	cuts.computeCuts()
	mapper := newOfMapper(cuts)
	if pars.nFastEdges != 0 {
		pars.nRounds = 1
		mapper.forwardDirect()
		mapper.backwardDirect()
	} else {
		for i := int32(0); i < pars.nRounds; i++ {
			mapper.forward()
			mapper.backward()
		}
		for i := int32(0); i < pars.nRoundsEla; i++ {
			if i < pars.nRoundsEla-1 {
				mapper.forwardExact()
			} else {
				mapper.forward()
			}
			mapper.backwardExact()
		}
	}
	result := mapper.deriveMapping()
	Gia_ManMappingVerify(result)
	if result.vPacking != nil {
		Gia_ManConvertPackingToEdges(result)
	}
	return result
}

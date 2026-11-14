package gia

import "github.com/glycerine/abcgo/misc/vec"

func ofDefaultParameters(p *Jf_Par_t) {
	*p = Jf_Par_t{NLutSize: 4, NCutNum: 16, NRounds: 3, NRoundsEla: 4, NCoarseLimit: 3, NAreaTuner: 10, DelayTarget: -1, NDelayLut1: 10, NDelayLut2: 2, FOptEdge: 1, NLutSizeMax: 6, NCutNumMax: 32, MapDelayTarget: -1}
}
func ofMap(g *Gia_Man_t, pars *Jf_Par_t) *Gia_Man_t {
	if g == nil || pars == nil || pars.NRounds < 1 {
		panic("gia: invalid Of mapping request")
	}
	if Gia_ManHasChoices(g) != 0 {
		pars.FCoarsen = 0
		pars.FCutMin = 1
	}
	if pars.FCoarsen != 0 {
		g = Gia_ManDupMuxes(g, pars.NCoarseLimit)
	}
	g.Refs = nil
	g.VCellMapping = nil
	cuts := newOfCutManager(g, pars)
	if cuts.truth != nil {
		defer vec.Vec_MemFree(cuts.truth.memory)
		defer vec.Vec_MemHashFree(cuts.truth.memory)
	}
	cuts.computeCuts()
	mapper := newOfMapper(cuts)
	if pars.NFastEdges != 0 {
		pars.NRounds = 1
		mapper.forwardDirect()
		mapper.backwardDirect()
	} else {
		for i := int32(0); i < pars.NRounds; i++ {
			mapper.forward()
			mapper.backward()
		}
		for i := int32(0); i < pars.NRoundsEla; i++ {
			if i < pars.NRoundsEla-1 {
				mapper.forwardExact()
			} else {
				mapper.forward()
			}
			mapper.backwardExact()
		}
	}
	result := mapper.deriveMapping()
	Gia_ManMappingVerify(result)
	if result.VPacking != nil {
		Gia_ManConvertPackingToEdges(result)
	}
	return result
}

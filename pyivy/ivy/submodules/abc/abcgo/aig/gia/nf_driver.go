package gia

// nfDefaultParameters follows the standard-cell mapper defaults.
func nfDefaultParameters(p *Jf_Par_t) {
	*p = Jf_Par_t{NLutSize: 6, NCutNum: 16, NRounds: 4, NRoundsEla: 2, NCoarseLimit: 3, NVerbLimit: 5, DelayTarget: -1, FOptEdge: 1, FCutMin: 1, NLutSizeMax: NF_LEAF_MAX, NCutNumMax: NF_CUT_MAX}
}

// nfMap runs the native cut, timing and area stages with an already prepared
// cell library. The caller retains ownership of the library truth store.
func nfMap(g *Gia_Man_t, pars *Jf_Par_t, library *nfLibrary) *Gia_Man_t {
	if g == nil || pars == nil || library == nil || pars.NRounds < 1 {
		panic("gia: invalid Nf mapping request")
	}
	if Gia_ManHasChoices(g) != 0 {
		pars.FCoarsen = 0
	}
	if pars.FCoarsen != 0 {
		g = Gia_ManDupMuxes(g, pars.NCoarseLimit)
	}
	g.Refs = nil
	g.VCellMapping = nil
	cuts := newNfCutManager(g, pars, library)
	cuts.computeCuts()
	matcher := newNfMatcher(cuts)
	var refs *nfMappingRefs
	for round := int32(0); round < pars.NRounds; round++ {
		matcher.computeMapping()
		refs = matcher.mappingRefs(round)
	}
	for round := int32(0); round < pars.NRoundsEla; round++ {
		matcher.exactAreaRound(refs)
	}
	matcher.fixOutputDrivers(refs)
	return matcher.deriveMapping(refs)
}

package gia

// nfDefaultParameters follows the standard-cell mapper defaults.
func nfDefaultParameters(p *Jf_Par_t) {
	*p = Jf_Par_t{nLutSize: 6, nCutNum: 16, nRounds: 4, nRoundsEla: 2, nCoarseLimit: 3, nVerbLimit: 5, DelayTarget: -1, fOptEdge: 1, fCutMin: 1, nLutSizeMax: NF_LEAF_MAX, nCutNumMax: NF_CUT_MAX}
}

// nfMap runs the native cut, timing and area stages with an already prepared
// cell library. The caller retains ownership of the library truth store.
func nfMap(g *Gia_Man_t, pars *Jf_Par_t, library *nfLibrary) *Gia_Man_t {
	if g == nil || pars == nil || library == nil || pars.nRounds < 1 {
		panic("gia: invalid Nf mapping request")
	}
	if Gia_ManHasChoices(g) != 0 {
		pars.fCoarsen = 0
	}
	if pars.fCoarsen != 0 {
		g = Gia_ManDupMuxes(g, pars.nCoarseLimit)
	}
	g.Refs = nil
	g.vCellMapping = nil
	cuts := newNfCutManager(g, pars, library)
	cuts.computeCuts()
	matcher := newNfMatcher(cuts)
	var refs *nfMappingRefs
	for round := int32(0); round < pars.nRounds; round++ {
		matcher.computeMapping()
		refs = matcher.mappingRefs(round)
	}
	for round := int32(0); round < pars.nRoundsEla; round++ {
		matcher.exactAreaRound(refs)
	}
	matcher.fixOutputDrivers(refs)
	return matcher.deriveMapping(refs)
}

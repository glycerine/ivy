package ssw

// Ssw_ManSetDefaultParams restores the source signal-correspondence defaults.
func Ssw_ManSetDefaultParams(p *Ssw_Pars_t) {
	*p = Ssw_Pars_t{
		nFramesK: 1, nFramesAddSim: 2, nBTLimit: 1000,
		nBTLimitGlobal: 5000000, nMinDomSize: 100, nItersStop: -1,
		nResimDelta: 1000, nStepsMax: -1, nSatVarMax: 1000,
		nRecycleCalls: 50, nSatVarMax2: 5000, nRecycleCalls2: 250,
	}
}

// Ssw_ManSetDefaultParamsLcorr selects the optimized latch correspondence defaults.
func Ssw_ManSetDefaultParamsLcorr(p *Ssw_Pars_t) {
	Ssw_ManSetDefaultParams(p)
	p.fLatchCorrOpt = 1
	p.nBTLimit = 10000
}

// Ssw_RarSetDefaultParams restores the source rarity-simulation defaults.
func Ssw_RarSetDefaultParams(p *Ssw_RarPars_t) {
	*p = Ssw_RarPars_t{nFrames: 20, nWords: 50, nBinSize: 8}
}

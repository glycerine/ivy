package ssw

// Ssw_ManSetDefaultParams restores the source signal-correspondence defaults.
func Ssw_ManSetDefaultParams(p *Ssw_Pars_t) {
	*p = Ssw_Pars_t{
		NFramesK: 1, NFramesAddSim: 2, NBTLimit: 1000,
		nBTLimitGlobal: 5000000, nMinDomSize: 100, NItersStop: -1,
		nResimDelta: 1000, nStepsMax: -1, NSatVarMax: 1000,
		nRecycleCalls: 50, NSatVarMax2: 5000, NRecycleCalls2: 250,
	}
}

// Ssw_ManSetDefaultParamsLcorr selects the optimized latch correspondence defaults.
func Ssw_ManSetDefaultParamsLcorr(p *Ssw_Pars_t) {
	Ssw_ManSetDefaultParams(p)
	p.FLatchCorrOpt = 1
	p.NBTLimit = 10000
}

// Ssw_RarSetDefaultParams restores the source rarity-simulation defaults.
func Ssw_RarSetDefaultParams(p *Ssw_RarPars_t) {
	*p = Ssw_RarPars_t{NFrames: 20, NWords: 50, NBinSize: 8}
}

package abs

// Gia_ManAbsSetDefaultParams initializes the legacy abstraction settings.
func Gia_ManAbsSetDefaultParams(p *Gia_ParAbs_t) {
	*p = Gia_ParAbs_t{
		nFramesMax: 10, nConfMax: 10000, fDynamic: 1,
		nFramesBmc: 250, nConfMaxBmc: 5000,
		nStableMax: 1000000, nRatio: 10, nBobPar: 1000000,
		fUseStart: 1, Status: -1, nFramesDone: -1,
	}
}

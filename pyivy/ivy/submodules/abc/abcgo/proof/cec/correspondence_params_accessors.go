package cec

// Cec_ParCorSetScorr configures the source parameters and callback used by the
// legacy signal-correspondence entry point.
func Cec_ParCorSetScorr(p *Cec_ParCor_t, backtrackLimit, stepsMax, verbose int32, data, callback any) {
	p.NBTLimit = backtrackLimit
	p.nStepsMax = stepsMax
	p.FVerbose = verbose
	p.FUseCSat = 1
	p.pData = data
	p.pFunc = callback
}

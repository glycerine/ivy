package ssw

// Ssw_ParsSetScorr applies the legacy signal-correspondence options and
// callback after the caller has initialized defaults.
func Ssw_ParsSetScorr(p *Ssw_Pars_t, backtrackLimit, stepsMax, verbose int32, data, callback any) {
	p.nBTLimit = backtrackLimit
	p.nStepsMax = stepsMax
	p.fVerbose = verbose
	p.pData = data
	p.pFunc = callback
}

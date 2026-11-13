package cec

// Cec4_ManSetParams resets the SAT-sweeping parameters to ABC's defaults.
func Cec4_ManSetParams(p *Cec_ParFra_t) {
	*p = Cec_ParFra_t{jType: 2, fSatSweeping: 1, nWords: 4, nRounds: 10, nItersMax: 2000, nBTLimit: 1000000, nSatVarMax: 1000, nCallsRecycle: 500, nGenIters: 100}
}

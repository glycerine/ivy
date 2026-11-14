package cec

// Cec4_ManSetParams resets the SAT-sweeping parameters to ABC's defaults.
func Cec4_ManSetParams(p *Cec_ParFra_t) {
	*p = Cec_ParFra_t{JType: 2, fSatSweeping: 1, NWords: 4, NRounds: 10, NItersMax: 2000, NBTLimit: 1000000, nSatVarMax: 1000, NCallsRecycle: 500, NGenIters: 100}
}

package cec

// Whole-struct assignment preserves C's memset-before-default behavior: callers
// may safely reuse parameter values that previously enabled optional modes.
func Cec_ManSatSetDefaultParams(p *Cec_ParSat_t) {
	*p = Cec_ParSat_t{SolverType: -1, nBTLimit: 100, nSatVarMax: 2000, nCallsRecycle: 200, fNonChrono: 1, fPolarFlip: 1}
}
func Cec_ManSimSetDefaultParams(p *Cec_ParSim_t) {
	*p = Cec_ParSim_t{nWords: 31, nFrames: 100, nRounds: 20, nNonRefines: 3}
}
func Cec_ManSmfSetDefaultParams(p *Cec_ParSmf_t) {
	*p = Cec_ParSmf_t{nWords: 31, nRounds: 200, nFrames: 200, nNonRefines: 3, nBTLimit: 100}
}
func Cec_ManFraSetDefaultParams(p *Cec_ParFra_t) {
	*p = Cec_ParFra_t{nWords: 15, nRounds: 15, nItersMax: 10, nBTLimit: 100, nDepthMax: 1, iOutFail: -1}
}
func Cec_ManCecSetDefaultParams(p *Cec_ParCec_t) { *p = Cec_ParCec_t{nBTLimit: 1000, iOutFail: -1} }
func Cec_ParCecSetVerbose(p *Cec_ParCec_t, value int32) { p.fVerbose = value }
func Cec_ParCecVerbose(p *Cec_ParCec_t) int32 { return p.fVerbose }
func Cec_ManCorSetDefaultParams(p *Cec_ParCor_t) {
	*p = Cec_ParCor_t{nWords: 15, nRounds: 15, nFrames: 1, nBTLimit: 100, nLevelMax: -1, nStepsMax: -1, fUseRings: 1, fUseCSat: 1}
}
func Cec_ManChcSetDefaultParams(p *Cec_ParChc_t) {
	*p = Cec_ParChc_t{nWords: 15, nRounds: 15, nBTLimit: 1000, fUseRings: 1}
}

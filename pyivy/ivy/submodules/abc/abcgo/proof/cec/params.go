package cec

// Whole-struct assignment preserves C's memset-before-default behavior: callers
// may safely reuse parameter values that previously enabled optional modes.
func Cec_ManSatSetDefaultParams(p *Cec_ParSat_t) {
	*p = Cec_ParSat_t{SolverType: -1, NBTLimit: 100, NSatVarMax: 2000, NCallsRecycle: 200, FNonChrono: 1, fPolarFlip: 1}
}
func Cec_ManSimSetDefaultParams(p *Cec_ParSim_t) {
	*p = Cec_ParSim_t{NWords: 31, NFrames: 100, NRounds: 20, NNonRefines: 3}
}
func Cec_ManSmfSetDefaultParams(p *Cec_ParSmf_t) {
	*p = Cec_ParSmf_t{NWords: 31, NRounds: 200, NFrames: 200, NNonRefines: 3, NBTLimit: 100}
}
func Cec_ManFraSetDefaultParams(p *Cec_ParFra_t) {
	*p = Cec_ParFra_t{NWords: 15, NRounds: 15, NItersMax: 10, NBTLimit: 100, NDepthMax: 1, iOutFail: -1}
}
func Cec_ManCecSetDefaultParams(p *Cec_ParCec_t) { *p = Cec_ParCec_t{NBTLimit: 1000, iOutFail: -1} }
func Cec_ParCecSetVerbose(p *Cec_ParCec_t, value int32) { p.FVerbose = value }
func Cec_ParCecVerbose(p *Cec_ParCec_t) int32 { return p.FVerbose }
func Cec_ManCorSetDefaultParams(p *Cec_ParCor_t) {
	*p = Cec_ParCor_t{nWords: 15, nRounds: 15, NFrames: 1, NBTLimit: 100, nLevelMax: -1, nStepsMax: -1, FUseRings: 1, FUseCSat: 1}
}
func Cec_ManChcSetDefaultParams(p *Cec_ParChc_t) {
	*p = Cec_ParChc_t{nWords: 15, nRounds: 15, NBTLimit: 1000, fUseRings: 1}
}

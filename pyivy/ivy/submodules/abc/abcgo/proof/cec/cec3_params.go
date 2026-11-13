package cec

// Cec3_Par_t configures the original Glucose-backed Cec3 sweeping engine.
type Cec3_Par_t struct {
	nSimWords, nSimRounds, nItersMax, nConfLimit int32
	fIsMiter, fUseCones, fVeryVerbose, fVerbose  int32
}

// Cec3_SetDefaultParams also clears any earlier optional settings.
func Cec3_SetDefaultParams(p *Cec3_Par_t) {
	*p = Cec3_Par_t{nSimWords: 12, nSimRounds: 4, nItersMax: 10, nConfLimit: 1000}
}

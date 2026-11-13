package bmc

// SetInterpolationOptions configures the bounded search used to reconstruct
// an interpolation counterexample. Other options retain their current values.
func (p *Saig_ParBmc_t) SetInterpolationOptions(start, conflictLimit, verbose int32) {
	p.nStart = start
	p.nConfLimit = conflictLimit
	p.fVerbose = verbose
}

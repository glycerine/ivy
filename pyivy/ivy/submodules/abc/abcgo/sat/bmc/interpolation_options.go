package bmc

// SetInterpolationOptions configures the bounded search used to reconstruct
// an interpolation counterexample. Other options retain their current values.
func (p *Saig_ParBmc_t) SetInterpolationOptions(start, conflictLimit, verbose int32) {
	p.NStart = start
	p.NConfLimit = conflictLimit
	p.FVerbose = verbose
}

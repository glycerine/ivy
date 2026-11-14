package bmc

// SetAbstractionVerbose sets diagnostic output for abstraction refinement.
func (p *Saig_ParBmc_t) SetAbstractionVerbose(verbose int32) { p.FVerbose = verbose }

// AbstractionVerbose reports whether abstraction refinement diagnostics are enabled.
func (p *Saig_ParBmc_t) AbstractionVerbose() int32 { return p.FVerbose }

package cec

// SetConflictLimit configures the per-query SAT conflict limit. Zero leaves
// the solver unbounded; initialize other settings with Cec_ManCecSetDefaultParams.
func (p *Cec_ParCec_t) SetConflictLimit(limit int32) { p.nBTLimit = limit }

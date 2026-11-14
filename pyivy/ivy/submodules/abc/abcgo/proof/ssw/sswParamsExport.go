package ssw

// Ssw_ParsSetFramesK sets the induction depth for signal correspondence.
func Ssw_ParsSetFramesK(p *Ssw_Pars_t, frames int32) { p.NFramesK = frames }

// Ssw_ParsSetBTLimit sets the per-node conflict limit.
func Ssw_ParsSetBTLimit(p *Ssw_Pars_t, limit int32) { p.NBTLimit = limit }

// Ssw_ParsSetBTLimitGlobal sets the conflict limit across runs.
func Ssw_ParsSetBTLimitGlobal(p *Ssw_Pars_t, limit int32) { p.nBTLimitGlobal = limit }

// Ssw_ParsIterations returns the number of iterations performed.
func Ssw_ParsIterations(p *Ssw_Pars_t) int32 { return p.nIters }

// Ssw_ParsConflicts returns the number of SAT conflicts performed.
func Ssw_ParsConflicts(p *Ssw_Pars_t) int32 { return p.nConflicts }

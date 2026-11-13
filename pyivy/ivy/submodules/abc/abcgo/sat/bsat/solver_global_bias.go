package bsat

// SetGlobalVariableBias enables ABC's interpolation-specific activity bias for
// the supplied solver variables. Passing nil disables the bias. The solver owns
// its mask; invalid variables are rejected before the existing bias changes.
func (s *sat_solver_t) SetGlobalVariableBias(variables []int32) {
	for _, v := range variables {
		if v < 0 || v >= s.size {
			panic("global variable outside solver")
		}
	}
	if variables == nil {
		s.pGlobalVars = nil
		return
	}
	mask := make([]int32, s.size)
	for _, v := range variables {
		mask[v] = 1
	}
	s.pGlobalVars = mask
}

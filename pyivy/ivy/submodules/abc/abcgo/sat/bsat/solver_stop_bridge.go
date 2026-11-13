package bsat

// SetStopFunc installs the original solver's stop callback and its run ID.
// A nil callback removes it. The callback is consulted at search boundaries.
func (s *sat_solver_t) SetStopFunc(runID int32, stop func(int32) int32) {
	s.RunId = runID
	s.pFuncStop = stop
}

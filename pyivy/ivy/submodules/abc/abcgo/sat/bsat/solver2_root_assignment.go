package bsat

// RootAssigned reports whether a variable currently has a root-level value.
// Call after Solve or Simplify, when temporary decisions have been canceled.
func (p *LabeledSolver2) RootAssigned(variable int32) bool {
	if variable < 0 || variable >= p.solver.solver.size {
		panic("bsat: variable outside solver")
	}
	return p.solver.solver.assigns[variable] != satUnassigned
}

// RootValue returns a variable's root-level Boolean value when assigned.
func (p *LabeledSolver2) RootValue(variable int32) (bool, bool) {
	if !p.RootAssigned(variable) {
		return false, false
	}
	return p.solver.solver.assigns[variable] == 0, true
}

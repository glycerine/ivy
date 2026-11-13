package bsat

// RootAssigned reports whether a variable currently has a root-level value.
// Call after Solve or Simplify, when temporary decisions have been canceled.
func (p *LabeledSolver2) RootAssigned(variable int32) bool {
	if variable < 0 || variable >= p.solver.solver.size {
		panic("bsat: variable outside solver")
	}
	return p.solver.solver.assigns[variable] != satUnassigned
}

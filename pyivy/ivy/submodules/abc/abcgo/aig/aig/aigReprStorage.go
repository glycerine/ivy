package aig

// Aig_ManHasReprs reports whether representative storage has been started.
func Aig_ManHasReprs(g *Graph) bool { return g.reprs != nil }

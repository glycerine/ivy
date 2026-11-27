package aig

// Aig_ManHasEquivs reports whether an equivalence table has been allocated,
// matching the C pEquivs != NULL precondition used by choice reconstruction.
func Aig_ManHasEquivs(g *Graph) int32 {
	if g == nil || g.equivs == nil {
		return 0
	}
	return 1
}

// Aig_ManStartEquivs allocates the choice table (pEquivs), so a graph built
// for choices has one even when it ends up with no choices.
func Aig_ManStartEquivs(g *Graph) {
	if g.equivs == nil {
		g.equivs = newEdgeTable()
	}
}

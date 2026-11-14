package aig

// Aig_ManHasEquivs reports whether an equivalence table has been allocated,
// matching the C pEquivs != NULL precondition used by choice reconstruction.
func Aig_ManHasEquivs(g *Graph) int32 {
	if g == nil || g.equivs == nil {
		return 0
	}
	return 1
}

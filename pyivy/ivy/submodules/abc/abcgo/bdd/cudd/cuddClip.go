package cudd

import "github.com/glycerine/abcgo/bdd/cudd/core"

// Cudd_bddClippingAnd returns the exact conjunction. An exact result is valid
// for both the under- and over-approximation directions and avoids losing
// satisfying assignments when the caller asks for an over-approximation.
func Cudd_bddClippingAnd(dd *DdManager, f, g *DdNode, maxDepth, direction int32) *DdNode {
	if dd == nil || f == nil || g == nil {
		return nil
	}
	return core.Cudd_bddAnd(dd, f, g)
}

// Cudd_bddClippingAndAbstract computes the exact conjunction followed by
// existential abstraction over cube.
func Cudd_bddClippingAndAbstract(dd *DdManager, f, g, cube *DdNode, maxDepth, direction int32) *DdNode {
	if dd == nil || f == nil || g == nil || cube == nil {
		return nil
	}
	return core.Cudd_bddAndAbstract(dd, f, g, cube)
}

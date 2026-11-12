/**CFile***********************************************************************

  FileName    [zMaxMin.c]

  PackageName [extra]

  Synopsis    [Experimental version of some ZDD operators.]

  Description [External procedures included in this module:
				<ul>
				<li> Extra_zddMaximal();
				<li> Extra_zddMinimal();
				<li> Extra_zddMaxUnion();
				<li> Extra_zddMinUnion();
				<li> Extra_zddDotProduct();
				<li> Extra_zddCrossProduct();
				<li> Extra_zddMaxDotProduct();
				</ul>
			Internal procedures included in this module:
				<ul>
				<li> extraZddMaximal();
				<li> extraZddMinimal();
				<li> extraZddMaxUnion();
				<li> extraZddMinUnion();
				<li> extraZddDotProduct();
				<li> extraZddCrossProduct();
				<li> extraZddMaxDotProduct();
				</ul>
			StaTc procedures included in this module:
				<ul>
				</ul>

          DotProduct and MaxDotProduct were introduced
		  by O.Coudert to solve problems arising in two-level planar routing
          See O. Coudert, C.-J. R. Shi. Exact Multi-Layer Topological Planar
		  Routing. Proc. of IEEE Custom Integrated Circuit Conference '96,
		  pp. 179-182.
	      ]

  SeeAlso     []

  Author      [Alan Mishchenko]

  Copyright   []

  ReviSon    [$zMaxMin.c, v.1.2, November 26, 2000, alanmi $]

******************************************************************************/

package extrab

import "github.com/glycerine/abcgo/bdd/extrab/sets"

func Extra_zddMinimal(dd *DdManager, f *DdNode) *DdNode     { return sets.Extra_zddMinimal(dd, f) }
func extraZddMinimal(dd *DdManager, f *DdNode) *DdNode      { return Extra_zddMinimal(dd, f) }
func Extra_zddMaxUnion(dd *DdManager, x, y *DdNode) *DdNode { return sets.Extra_zddMaxUnion(dd, x, y) }
func extraZddMaxUnion(dd *DdManager, x, y *DdNode) *DdNode  { return Extra_zddMaxUnion(dd, x, y) }
func Extra_zddMinUnion(dd *DdManager, x, y *DdNode) *DdNode { return sets.Extra_zddMinUnion(dd, x, y) }
func extraZddMinUnion(dd *DdManager, x, y *DdNode) *DdNode  { return Extra_zddMinUnion(dd, x, y) }
func Extra_zddDotProduct(dd *DdManager, x, y *DdNode) *DdNode {
	return sets.Extra_zddDotProduct(dd, x, y)
}
func extraZddDotProduct(dd *DdManager, x, y *DdNode) *DdNode { return Extra_zddDotProduct(dd, x, y) }
func Extra_zddCrossProduct(dd *DdManager, x, y *DdNode) *DdNode {
	return sets.Extra_zddCrossProduct(dd, x, y)
}
func extraZddCrossProduct(dd *DdManager, x, y *DdNode) *DdNode {
	return Extra_zddCrossProduct(dd, x, y)
}
func Extra_zddMaxDotProduct(dd *DdManager, x, y *DdNode) *DdNode {
	return sets.Extra_zddMaxDotProduct(dd, x, y)
}
func extraZddMaxDotProduct(dd *DdManager, x, y *DdNode) *DdNode {
	return Extra_zddMaxDotProduct(dd, x, y)
}

// Maximal-family entrypoints are defined in extraBddSet.go and share the same
// authoritative sets engine used by the fused not-superset operation.

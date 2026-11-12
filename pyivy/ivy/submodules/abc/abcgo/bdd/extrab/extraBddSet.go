/**CFile***********************************************************************

  FileName    [zSubSet.c]

  PackageName [extra]

  Synopsis    [Experimental version of some ZDD operators.]

  Description [External procedures included in this module:
		    <ul>
		    <li> Extra_zddSubSet();
		    <li> Extra_zddSupSet();
		    <li> Extra_zddNotSubSet();
		    <li> Extra_zddNotSupSet();
		    <li> Extra_zddMaxNotSupSet();
			<li> Extra_zddEmptyBelongs();
			<li> Extra_zddIsOneSubset();
		    </ul>
	       Internal procedures included in this module:
		    <ul>
		    <li> extraZddSubSet();
		    <li> extraZddSupSet();
		    <li> extraZddNotSubSet();
		    <li> extraZddNotSupSet();
		    <li> extraZddMaxNotSupSet();
		    </ul>
	       Static procedures included in this module:
		    <ul>
		    </ul>

          SubSet, SupSet, NotSubSet, NotSupSet were introduced
		  by O.Coudert to solve problems arising in two-level SOP
		  minimization. See O. Coudert, "Two-Level Logic Minimization:
		  An Overview", Integration. Vol. 17, No. 2, pp. 97-140, Oct 1994.
	      ]

  SeeAlso     []

  Author      [Alan Mishchenko]

  Copyright   []

  Revision    [$zSubSet.c, v.1.2, November 16, 2000, alanmi $]

******************************************************************************/

package extrab

import "github.com/glycerine/abcgo/bdd/extrab/sets"

func Extra_zddSubSet(dd *DdManager, x, y *DdNode) *DdNode { return sets.Extra_zddSubSet(dd, x, y) }
func extraZddSubSet(dd *DdManager, x, y *DdNode) *DdNode  { return Extra_zddSubSet(dd, x, y) }
func Extra_zddSupSet(dd *DdManager, x, y *DdNode) *DdNode { return sets.Extra_zddSupSet(dd, x, y) }
func extraZddSupSet(dd *DdManager, x, y *DdNode) *DdNode  { return Extra_zddSupSet(dd, x, y) }
func Extra_zddNotSubSet(dd *DdManager, x, y *DdNode) *DdNode {
	return sets.Extra_zddNotSubSet(dd, x, y)
}
func extraZddNotSubSet(dd *DdManager, x, y *DdNode) *DdNode { return Extra_zddNotSubSet(dd, x, y) }
func Extra_zddNotSupSet(dd *DdManager, x, y *DdNode) *DdNode {
	return sets.Extra_zddNotSupSet(dd, x, y)
}
func extraZddNotSupSet(dd *DdManager, x, y *DdNode) *DdNode { return Extra_zddNotSupSet(dd, x, y) }
func Extra_zddMaxNotSupSet(dd *DdManager, x, y *DdNode) *DdNode {
	return sets.Extra_zddMaxNotSupSet(dd, x, y)
}
func extraZddMaxNotSupSet(dd *DdManager, x, y *DdNode) *DdNode {
	return Extra_zddMaxNotSupSet(dd, x, y)
}
func Extra_zddEmptyBelongs(dd *DdManager, f *DdNode) bool { return sets.Extra_zddEmptyBelongs(dd, f) }
func Extra_zddIsOneSubset(dd *DdManager, f *DdNode) bool  { return sets.Extra_zddIsOneSubset(dd, f) }
func Extra_zddMaximal(dd *DdManager, f *DdNode) *DdNode   { return sets.Extra_zddMaximal(dd, f) }
func extraZddMaximal(dd *DdManager, f *DdNode) *DdNode    { return Extra_zddMaximal(dd, f) }

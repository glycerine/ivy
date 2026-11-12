/**CFile****************************************************************

  FileName    [extraBdd.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Various reusable software utilities.]

  Description [This library contains a number of operators and
  traversal routines developed to extend the functionality of
  CUDD v.2.3.x, by Fabio Somenzi (http://vlsi.colorado.edu/~fabio/)
  To compile your code with the library, #include "extra.h"
  in your source files and link your project to CUDD and this
  library. Use the library at your own risk and with caution.
  Note that debugging of some operators still continues.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraBdd.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

package extrab

import cudd "github.com/glycerine/abcgo/bdd/cudd/core"

type DdManager = cudd.DdManager
type DdNode = cudd.DdNode
type Word = uint64

type Extra_UnateVar_t struct {
	IVar     int
	Pos, Neg bool
}
type Extra_UnateInfo_t struct {
	NVars, NVarsMax, NUnate int
	Vars                    []Extra_UnateVar_t
}
type Extra_SymmInfo_t struct {
	NVars, NVarsMax, NSymms, NNodes int
	Vars                            []int
	Symms                           [][]bool
}

func extraValidBDD(dd *DdManager, roots ...*DdNode) bool {
	if dd == nil {
		return false
	}
	for _, f := range roots {
		if !cudd.Cudd_bddLeq(dd, f, f) {
			return false
		}
	}
	return true
}
func extraSupportSet(dd *DdManager, f *DdNode) map[int]bool {
	support := map[int]bool{}
	for _, i := range dd.Support(f) {
		support[i] = true
	}
	return support
}

func extraBDDCofactors(f *DdNode) (lo, hi *DdNode) {
	r := cudd.Cudd_Regular(f)
	lo, hi = r.E, r.T
	if cudd.Cudd_IsComplement(f) {
		lo, hi = cudd.Cudd_Not(lo), cudd.Cudd_Not(hi)
	}
	return
}
func extraPositiveCube(dd *DdManager, cube *DdNode) ([]int, bool) {
	if !extraValidBDD(dd, cube) {
		return nil, false
	}
	vars := []int{}
	for cube != dd.One {
		if cudd.Cudd_IsConstant(cube) || cudd.Cudd_IsComplement(cube) || cube.E != dd.LogicZero {
			return nil, false
		}
		vars = append(vars, int(cube.Index))
		cube = cube.T
	}
	return vars, true
}

func extraSelectSubset(dd *DdManager, f *DdNode) *DdNode {
	cube := dd.One
	for f != dd.One {
		if f == dd.Zero {
			return dd.Zero
		}
		if f.E != dd.Zero {
			f = f.E
		} else {
			cube = cudd.Cudd_zddChange(dd, cube, int(f.Index))
			f = f.T
		}
	}
	return cube
}

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

package gia

import (
	"fmt"
)

/**CFile****************************************************************

  FileName    [giaJf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaJf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecSet.h"
//#include "misc/vec/vecMem.h"
//#include "misc/extra/extra.h"
//#include "bool/kit/kit.h"
//#include "misc/util/utilTruth.h"
//#include "opt/dau/dau.h"
//#include "sat/cnf/cnf.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

//static inline int    Jf_ObjLit( int i, int c )            { return i;                                                }

//static inline int    Jf_CutVar( int * pCut, int i )         { assert(i); return pCut[i];                               }

// Original macro:
// #define Jf_ObjForEachCut( pList, pCut, i )   for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += Jf_CutSize(pCut) + 1 )

// Original macro:
// #define Jf_CutForEachLit( pCut, Lit, i )     for ( i = 1; i <= Jf_CutSize(pCut) && (Lit = Jf_CutLit(pCut, i)); i++ )

// Original macro:
// #define Jf_CutForEachVar( pCut, Var, i )     for ( i = 1; i <= Jf_CutSize(pCut) && (Var = Jf_CutVar(pCut, i)); i++ )

// extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derives CNF for the mapped GIA.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Computing references while discounting XOR/MUX.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Manager manipulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cut functions.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

func Jf_ObjBestCutConePrint(p *Jf_Man_t, pObj *Gia_Obj_t) {
	pCut := Jf_ObjCutBest(p, Gia_ObjId(p.pGia, pObj))
	fmt.Printf("Best cut of node %d : ", Gia_ObjId(p.pGia, pObj))
	Jf_CutPrint(pCut)
	Gia_ManPrintCone(p.pGia, pObj, Jf_CutLits(pCut), Jf_CutSize(pCut), p.vTemp)
}

/**Function*************************************************************

  Synopsis    [Cut merging.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cut merging.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cut merging.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cut filtering.]

  Description [Returns the number of cuts after filtering and the last
  cut in the last entry.  If the cut is filtered, its size is set to -1.]

  SideEffects [This was found to be 15% slower.]

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Sorting cuts by size.]

  Description []

  SideEffects [Did not really help.]

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Reference counting.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

// returns 1 if MFFC size is less than limit

/**Function*************************************************************

  Synopsis    [Comparison procedures.]

  Description [Return positive value if the new cut is better than the old cut.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cut minimization.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cut enumeration.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Computing delay/area.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Mapping rounds.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Derives the result of mapping.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Derive GIA without mapping.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

package xsat

import "fmt"

/**CFile****************************************************************

  FileName    [xsatClause.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [xSAT - A SAT solver written in C.
               Read the license file for more info.]

  Synopsis    [Clause data type definition.]

  Author      [Bruno Schmitt <boschmitt@inf.ufrgs.br>]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - November 10, 2016.]

  Revision    []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
//#include "misc/util/abc_global.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
   
type xSAT_Clause_t struct {
    fLearnt uint32; /* C bitfield: 1 bits; review packing and truncation */
    fMark uint32; /* C bitfield: 1 bits; review packing and truncation */
    fReallocd uint32; /* C bitfield: 1 bits; review packing and truncation */
    fCanBeDel uint32; /* C bitfield: 1 bits; review packing and truncation */
    nLBD uint32; /* C bitfield: 28 bits; review packing and truncation */
    nSize int32;
    union {
        int Lit;
        unsigned Act;
    }
 pData[0];
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DECLARATIONS                        ///
////////////////////////////////////////////////////////////////////////
/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_ClauseCompare(p1 any, p2 any) int32 {
    var pC1 *xSAT_Clause_t  = ( xSAT_Clause_t * ) p1;
    var pC2 *xSAT_Clause_t  = ( xSAT_Clause_t * ) p2;

    if pC1.nSize > 2 && pC2.nSize == 2 {
        return 1; }
    if pC1.nSize == 2 && pC2.nSize > 2 {
        return 0; }
    if pC1.nSize == 2 && pC2.nSize == 2 {
        return 0; }

    if pC1.nLBD > pC2.nLBD {
        return 1; }
    if pC1.nLBD < pC2.nLBD {
        return 0; }

    return pC1.pData[pC1.nSize].Act < pC2.pData[pC2.nSize].Act;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_ClausePrint(pCla *xSAT_Clause_t) {
    var i int32;

    fmt.Printf("{ ");
    for  i = 0; i < pCla.nSize; i++  {
        fmt.Printf("%d ", pCla.pData[i].Lit);
    }
    fmt.Printf("}\n");
}

//ABC_NAMESPACE_HEADER_END


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

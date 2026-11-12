package gia


/**CFile****************************************************************

  FileName    [giaEquiv.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Manipulation of equivalence classes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaEquiv.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "proof/cec/cec.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Manipulating original IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManOrigIdsReduceTest(p *Gia_Man_t, vPairs *Vec_Int_t) *Gia_Man_t {
    var pTemp *Gia_Man_t;
    var pNew *Gia_Man_t = Gia_ManOrigIdsReduce( p, vPairs );
    Gia_ManPrintStats( p, nil );
    Gia_ManPrintStats( pNew, nil );
    //Gia_ManStop( pNew );
    // cleanup the resulting one
    pTemp = pNew;
    pNew = Gia_ManCleanup(pTemp);
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Compute equivalence classes of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Reduces AIG using equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Converting AIG after SAT sweeping into AIG with choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

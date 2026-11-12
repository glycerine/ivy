package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaShrink7.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Implementation of DAG-aware unmapping for 6-input cuts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaShrink6.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecHash.h"
//#include "misc/util/utilTruth.h"


//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// operation manager
   
/**Function*************************************************************

  Synopsis    [Computes information about node pairs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Collects decomposable pairs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Compute truth tables for the selected nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unm_ManWork(p *Unm_Man_t) {
    var vPairs, vPairs2 *Vec_Int_t;
    // find the duplicated pairs
    vPairs = Unm_ManComputePairs( p, 1 );
    // find the used nodes
    p.vUsed = Unm_ManFindUsedNodes( vPairs, Gia_ManObjNum(p.pGia) );
    p.vId2Used = Vec_IntInvert( p.vUsed, -1 );
    Vec_IntFillExtra( p.vId2Used, Gia_ManObjNum(p.pGia), -1 );
    // compute truth tables for used nodes
    p.vTruths = Unm_ManComputeTruths( p );
    // derive new pairs
    vPairs2 = Unm_ManCollectDecomp( p, vPairs, 1 );
    Vec_IntFreeP( &vPairs );
    Vec_IntFreeP( &vPairs2 );
}


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unm_ManTest(pGia *Gia_Man_t) *Gia_Man_t {
    var p *Unm_Man_t;
    p = Unm_ManAlloc( pGia );
    Unm_ManWork( p );

    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    return Unm_ManFree( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


package reo

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [reoUnits.c]

  PackageName [REO: A specialized DD reordering engine.]

  Synopsis    [Procedures which support internal data structures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 15, 2002.]

  Revision    [$Id: reoUnits.c,v 1.0 2002/15/10 03:00:00 alanmi Exp $]

***********************************************************************/

//#include "reo.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void reoUnitsAddToFreeUnitList( reo_man * p );


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Extract the next unit from the free unit list.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoUnitsGetNextUnit(p *reo_man) *reo_unit {
    var pUnit *reo_unit;
    // check there are stil units to extract
    if p.pUnitFreeList == nil {
        reoUnitsAddToFreeUnitList( p ); }
    // extract the next unit from the linked list
    pUnit            = p.pUnitFreeList;
    p.pUnitFreeList = pUnit.Next;
    p.nUnitsUsed++;
    return pUnit;
}

/**Function*************************************************************

  Synopsis    [Returns the unit to the free unit list.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoUnitsRecycleUnit(p *reo_man, pUnit *reo_unit) {
    pUnit.Next      = p.pUnitFreeList;
    p.pUnitFreeList = pUnit;
    p.nUnitsUsed--;
}

/**Function*************************************************************

  Synopsis    [Returns the list of units to the free unit list.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoUnitsRecycleUnitList(p *reo_man, pPlane *reo_plane) {
    var pUnit *reo_unit;
    var pTail *reo_unit  = nil; // Suppress "might be used uninitialized"

    if pPlane.pHead == nil {
        return; }

    // find the tail
    for  pUnit = pPlane.pHead; (pUnit != nil); pUnit = pUnit.Next  {
        pTail = pUnit;
    }
    pTail.Next = p.pUnitFreeList;
    p.pUnitFreeList    = pPlane.pHead;
    memset( pPlane, 0, sizeof(reo_plane) );
//    pPlane->pHead = NULL;
}

/**Function*************************************************************

  Synopsis    [Stops the unit dispenser.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoUnitsStopDispenser(p *reo_man) {
    var i int32;
    for  i = 0; i < p.nMemChunks; i++  {
        ABC_FREE( p.pMemChunks[i] );
    }
//    printf("\nThe number of chunks used is %d, each of them %d units\n", p->nMemChunks, REO_CHUNK_SIZE );
    p.nMemChunks = 0;
}

/**Function*************************************************************

  Synopsis    [Adds one unit to the list of units which constitutes the plane.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoUnitsAddUnitToPlane(pPlane *reo_plane, pUnit *reo_unit) {
    if pPlane.pHead == nil {
        pPlane.pHead = pUnit;
        pUnit.Next   = nil;
    
    } else {
        pUnit.Next   = pPlane.pHead;
        pPlane.pHead = pUnit;
    }
    pPlane.statsNodes++;
}


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoUnitsAddToFreeUnitList(p *reo_man) {
    var c int32;
    // check that we still have chunks left
    if p.nMemChunks == p.nMemChunksAlloc {
        fmt.Printf("reoUnitsAddToFreeUnitList(): Memory manager ran out of memory!\n");
        fflush( os.Stdout );
        return;
    }
    // allocate the next chunk
    assert(p.pUnitFreeList == nil, "p.pUnitFreeList == nil");
    p.pUnitFreeList = ABC_ALLOC( reo_unit, REO_CHUNK_SIZE );
    // split chunks into list-connected units
    for  c = 0; c < REO_CHUNK_SIZE-1; c++  {
        (p.pUnitFreeList + c).Next = p.pUnitFreeList + c + 1;
    }
    // set the last pointer to NULL
    (p.pUnitFreeList + REO_CHUNK_SIZE-1).Next = nil;
    // add the chunk to the array of chunks
    p.pMemChunks[p.nMemChunks++] = p.pUnitFreeList;
}


////////////////////////////////////////////////////////////////////////
///                         END OF FILE                              ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END


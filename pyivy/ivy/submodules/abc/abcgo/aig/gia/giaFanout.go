package gia

/**CFile****************************************************************

  FileName    [giaFanout.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaFanout.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// 0: first iFan
// 1: prev iFan0
// 2: prev iFan1
// 3: next iFan0
// 4: next iFan1

// these two procedures are only here for the use inside the iterator

// iterator over the fanouts
// Original macro:
// #define Gia_ObjForEachFanout( p, pObj, pFanout, iFan, i )                       \
//     for ( assert(p->pFanData), i = 0; (i < (int)(pObj)->nRefs) &&               \
//           (((iFan) = i? Gia_ObjFanoutNext(p, iFan) : Gia_ObjFanout0Int(p, Gia_ObjId(p, pObj))), 1) && \
//           (((pFanout) = Gia_ManObj(p, iFan>>1)), 1); i++ )

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Create fanout for all objects in the manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Deletes fanout for all objects in the manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Adds fanout (pFanout) of node (pObj).]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Removes fanout (pFanout) of node (pObj).]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Compute the map of all edges.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Allocates static fanout.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Compute the map of all edges.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

package gia

/**CFile****************************************************************

  FileName    [giaEmbed.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Logic network derived from AIG.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaEmbed.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <math.h>
//#include "gia.h"
//#include "aig/ioa/ioa.h"

//ABC_NAMESPACE_IMPL_START

/*
   The code is based on the paper by D. Harel and Y. Koren,
   "Graph drawing by high-dimensional embedding",
   J. Graph Algs & Apps, Vol 8(2), pp. 195-217 (2004).
   http://www.emis.de/journals/JGAA/accepted/2004/HarelKoren2004.8.2.pdf

   Iterative refinement is described in the paper: F. A. Aloul, I. L. Markov, and K. A. Sakallah.
   "FORCE: A Fast and Easy-To-Implement Variable-Ordering Heuristic", Proc. GLSVLSI_03.
   http://www.eecs.umich.edu/~imarkov/pubs/conf/glsvlsi03-force.pdf
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// objects will be placed in box [0, GIA_PLACE_SIZE] x [0, GIA_PLACE_SIZE]

// Original macro:
// #define Emb_ManForEachObj( p, pObj, i )               \
//     for ( i = 0; (i < p->nObjData) && (pObj = Emb_ManObj(p,i)); i += Emb_ObjSize(pObj) )

// Original macro:
// #define Emb_ManForEachNode( p, pObj, i )              \
//     for ( i = 0; (i < p->nObjData) && (pObj = Emb_ManObj(p,i)); i += Emb_ObjSize(pObj) ) if ( Emb_ObjIsTerm(pObj) ) {} else

// Original macro:
// #define Emb_ManForEachObjVec( vVec, p, pObj, i )                        \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Emb_ManObj(p, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Emb_ObjForEachFanin( pObj, pNext, i )         \
//     for ( i = 0; (i < (int)pObj->nFanins) && (pNext = Emb_ObjFanin(pObj,i)); i++ )

// Original macro:
// #define Emb_ObjForEachFanout( pObj, pNext, i )        \
//     for ( i = 0; (i < (int)pObj->nFanouts) && (pNext = Emb_ObjFanout(pObj,i)); i++ )

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates fanin/fanout pair.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

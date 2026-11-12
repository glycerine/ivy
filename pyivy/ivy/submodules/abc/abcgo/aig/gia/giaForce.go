package gia

/**CFile****************************************************************

  FileName    [gia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START

/*
   The code is based on the paper by F. A. Aloul, I. L. Markov, and K. A. Sakallah.
   "FORCE: A Fast and Easy-To-Implement Variable-Ordering Heuristic", Proc. GLSVLSI�03.
   http://www.eecs.umich.edu/~imarkov/pubs/conf/glsvlsi03-force.pdf
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Frc_ManForEachObj( p, pObj, i )               \
//     for ( i = 0; (i < p->nObjData) && (pObj = Frc_ManObj(p,i)); i += Frc_ObjSize(pObj) )

// Original macro:
// #define Frc_ManForEachObjVec( vVec, p, pObj, i )                        \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Frc_ManObj(p, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Frc_ManForEachNode( p, pObj, i )              \
//     for ( i = 0; (i < p->nObjData) && (pObj = Frc_ManObj(p,i)); i += Frc_ObjSize(pObj) ) if ( Frc_ObjIsTerm(pObj) ) {} else

// Original macro:
// #define Frc_ManForEachCi( p, pObj, i )                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCis)) && (pObj = Frc_ManObj(p,Vec_IntEntry(p->vCis,i))); i++ )

// Original macro:
// #define Frc_ManForEachCo( p, pObj, i )                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCos)) && (pObj = Frc_ManObj(p,Vec_IntEntry(p->vCos,i))); i++ )

// Original macro:
// #define Frc_ObjForEachFanin( pObj, pNext, i )         \
//     for ( i = 0; (i < (int)pObj->nFanins) && (pNext = Frc_ObjFanin(pObj,i)); i++ )

// Original macro:
// #define Frc_ObjForEachFaninReverse( pObj, pNext, i )  \
//     for ( i = (int)pObj->nFanins - 1; (i >= 0) && (pNext = Frc_ObjFanin(pObj,i)); i-- )

// Original macro:
// #define Frc_ObjForEachFanout( pObj, pNext, i )        \
//     for ( i = 0; (i < (int)pObj->nFanouts) && (pNext = Frc_ObjFanout(pObj,i)); i++ )

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates fanin/fanout pair.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

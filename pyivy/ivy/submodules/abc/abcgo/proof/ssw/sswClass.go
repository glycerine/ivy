package ssw

/**CFile****************************************************************

  FileName    [sswClass.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Representation of candidate equivalence classes.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswClass.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"

//ABC_NAMESPACE_IMPL_START

/*
   The candidate equivalence classes are stored as a vector of pointers
   to the array of pointers to the nodes in each class.
   The first node of the class is its representative node.
   The representative has the smallest topological order among the class nodes.
   The nodes inside each class are ordered according to their topological order.
   The classes are ordered according to the topo order of their representatives.
*/

// internal representation of candidate equivalence classes

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// iterator through the equivalence classes
// Original macro:
// #define Ssw_ManForEachClass( p, ppClass, i )                 \
//     for ( i = 0; i < Aig_ManObjNumMax(p->pAig); i++ )        \
//         if ( ((ppClass) = p->pId2Class[i]) == NULL ) {} else

// iterator through the nodes in one class
// Original macro:
// #define Ssw_ClassForEachNode( p, pRepr, pNode, i )           \
//     for ( i = 0; i < p->pClassSizes[pRepr->Id]; i++ )        \
//         if ( ((pNode) = p->pId2Class[pRepr->Id][i]) == NULL ) {} else

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates one equivalence class.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Removes one equivalence class.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Starts representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Starts representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

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

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Stop representation of equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Checks candidate equivalence classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prints simulation classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prints simulation classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prints simulation classes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Takes the set of const1 cands and rehashes them using sim info.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates initial simulation classes.]

  Description [Assumes that simulation info is assigned.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates classes from the temporary representation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates classes from the temporary representation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Iteratively refines the classes after simulation.]

  Description [Returns the number of refinements performed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Refines the classes after simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Refines the classes after simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Refine the group of constant 1 nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Refine the group of constant 1 nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

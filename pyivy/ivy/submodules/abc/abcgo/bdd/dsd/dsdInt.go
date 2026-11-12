package dsd

/**CFile****************************************************************

  FileName    [dsdInt.h]

  PackageName [DSD: Disjoint-support decomposition package.]

  Synopsis    [Internal declarations of the package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 8.0. Started - September 22, 2003.]

  Revision    [$Id: dsdInt.h,v 1.0 2002/22/09 00:00:00 alanmi Exp $]

***********************************************************************/





//#include "bdd/extrab/extraBdd.h"
//#include "dsd.h"

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                      TYPEDEF DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
typedef unsigned char byte;

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

// DSD manager
type Dsd_Manager_t_ struct {
    dd     *DdManager;         // the BDD manager
    Table      *st__table;      // the mapping of BDDs into their DEs
    nInputs            int32;    // the number of primary inputs
    nRoots            int32;     // the number of primary outputs
    nRootsAlloc            int32;// the number of primary outputs
    pInputs   **Dsd_Node_t;    // the primary input nodes
    pRoots   **Dsd_Node_t;     // the primary output nodes
    pConst1    *Dsd_Node_t;    // the constant node
    fVerbose            int32;   // the verbosity level 
}

// DSD node
type Dsd_Node_t_ struct {
    Type     Dsd_Type_t;       // decomposition type
    G        *DdNode;          // function of the node   
    S        *DdNode;          // support of this function
    pDecs   **Dsd_Node_t;      // pointer to structures for formal inputs
    Mark            int32;       // the mark used by CASE 4 of disjoint decomposition
    nDecs          int16;      // the number of formal inputs
    nVisits          int16;    // the counter of visits
}

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

const MAXINPUTS = 1000

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DECLARATIONS                        ///
////////////////////////////////////////////////////////////////////////

/*=== dsdCheck.c =======================================================*/
// extern void         Dsd_CheckCacheAllocate( int nEntries );

// extern void         Dsd_CheckCacheDeallocate();

// extern void         Dsd_CheckCacheClear();

// extern int          Dsd_CheckRootFunctionIdentity( DdManager * dd, DdNode * bF1, DdNode * bF2, DdNode * bC1, DdNode * bC2 );

/*=== dsdTree.c =======================================================*/
// extern Dsd_Node_t * Dsd_TreeNodeCreate( int Type, int nDecs, int BlockNum );

// extern void         Dsd_TreeNodeDelete( DdManager * dd, Dsd_Node_t * pNode );

// extern void         Dsd_TreeUnmark( Dsd_Manager_t * dMan );

// extern DdNode *     Dsd_TreeGetPrimeFunctionOld( DdManager * dd, Dsd_Node_t * pNode, int fRemap );




//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                           END OF FILE                            ///
////////////////////////////////////////////////////////////////////////


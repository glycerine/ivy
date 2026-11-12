package gia

/**CFile****************************************************************

  FileName    [giaMap.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Manipulation of mapping associated with the AIG.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaMap.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "aig/aig/aig.h"
//#include "map/if/if.h"
//#include "bool/kit/kit.h"
//#include "base/main/main.h"
//#include "sat/bsat/satSolver.h"

// Windows console headers are unnecessary in Go.

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );

// extern int Abc_RecToGia3( Gia_Man_t * pMan, If_Man_t * pIfMan, If_Cut_t * pCut, Vec_Int_t * vLeaves, int fHash );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Load the network into FPGA manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Assigns levels.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Assigns levels.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Calculate mapping overlap.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Calculate mapping overlap.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

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

  Synopsis    [Computes levels for AIG with choices and white boxes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Checks integrity of choice nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Converts GIA into IF manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Rebuilds GIA from mini AIG.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Converts IF into GIA manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Write mapping for LUT with given fanins.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Write mapping for LUT with given fanins.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Write the node into a file.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Recursively derives the local AIG for the cut.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Converts IF into GIA manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Converts IF into GIA manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// Gia_ManFromIfLogicFindCell is implemented in map/if with its IF types.

/**Function*************************************************************

  Synopsis    [Converts IF into GIA manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Converts IF into GIA manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Verifies mapping.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Transfers mapping from hie GIA to normalized GIA.]

  Description [Hie GIA (pGia) points to normalized GIA (p).]

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

  Synopsis    [Interface of LUT mapping package.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Interface of other mapping-based procedures.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Tests decomposition structures.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs hashing for a mapped AIG.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

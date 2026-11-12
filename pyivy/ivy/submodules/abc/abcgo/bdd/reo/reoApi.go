package reo

/**CFile****************************************************************

  FileName    [reoApi.c]

  PackageName [REO: A specialized DD reordering engine.]

  Synopsis    [Implementation of API functions.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 15, 2002.]

  Revision    [$Id: reoApi.c,v 1.0 2002/15/10 03:00:00 alanmi Exp $]

***********************************************************************/

//#include "reo.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Initializes the reordering engine.]

  Description [The first argument is the max number of variables in the
  CUDD DD manager which will be used with the reordering engine 
  (this number of should be the maximum of BDD and ZDD parts). 
  The second argument is the maximum number of BDD nodes in the BDDs 
  to be reordered. These limits are soft. Setting lower limits will later 
  cause the reordering manager to resize internal data structures. 
  However, setting the exact values will make reordering more efficient 
  because resizing will be not necessary.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReorderInit(nDdVarsMax int32, nNodesMax int32) *reo_man {
    var p *reo_man;
    // allocate and clean the data structure
    p = new(reo_man);
    memset( p, 0, sizeof(reo_man) );
    // resize the manager to meet user's needs    
    reoResizeStructures( p, nDdVarsMax, nNodesMax, 100 );
    // set the defaults
    p.fMinApl   = 0;
    p.fMinWidth = 0;
    p.fRemapUp  = 0;
    p.fVerbose  = 0;
    p.fVerify   = 0;
    p.nIters    = 1;
    return p;
}

/**Function*************************************************************

  Synopsis    [Disposes of the reordering engine.]

  Description [Removes all memory associated with the reordering engine.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReorderQuit(p *reo_man) {
    ABC_FREE( p.pTops );
    ABC_FREE( p.pSupp );
    ABC_FREE( p.pOrderInt );
    ABC_FREE( p.pWidthCofs );
    ABC_FREE( p.pMapToPlanes );
    ABC_FREE( p.pMapToDdVarsOrig );
    ABC_FREE( p.pMapToDdVarsFinal );
    ABC_FREE( p.pPlanes );
    ABC_FREE( p.pVarCosts );
    ABC_FREE( p.pLevelOrder );
    ABC_FREE( p.HTable );
    ABC_FREE( p.pRefNodes );
    reoUnitsStopDispenser( p );
    ABC_FREE( p.pMemChunks );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Sets the type of DD minimizationl that will be performed.]

  Description [Currently, three different types of minimization are supported.
  It is possible to minimize the number of BDD nodes. This is a classical type
  of minimization, which is attempting to reduce the total number of nodes in
  the (shared) BDD of the given Boolean functions. It is also possible to 
  minimize the BDD width, defined as the sum total of the number of cofactors 
  on each level in the (shared) BDD (note that the number of cofactors on the 
  given level may be larger than the number of nodes appearing on the given level).
  It is also possible to minimize the average path length in the (shared) BDD 
  defined as the sum of products, for all BDD paths from the top node to any 
  terminal node, of the number of minterms on the path by the number of nodes 
  on the path. The default reordering type is minimization for the number of 
  BDD nodes. Calling this function with REO_MINIMIZE_WIDTH or REO_MINIMIZE_APL
  as the second argument, changes the default minimization option for all the 
  reorder calls performed afterwards.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReorderSetMinimizationType(p *reo_man, fMinType reo_min_type) {
    if fMinType == REO_MINIMIZE_NODES {
        p.fMinWidth = 0;
        p.fMinApl   = 0;
    
    } else if fMinType == REO_MINIMIZE_WIDTH {
        p.fMinWidth = 1;
        p.fMinApl   = 0;
    
    } else if fMinType == REO_MINIMIZE_APL {
        p.fMinWidth = 0;
        p.fMinApl   = 1;
    
    } else {
        assert(false, "0");
    }
}

/**Function*************************************************************

  Synopsis    [Sets the type of remapping performed by the engine.]

  Description [The remapping refers to the way the resulting BDD
  is expressed using the elementary variables of the CUDD BDD manager.
  Currently, two types possibilities are supported: remapping and no
  remapping. Remapping means that the function(s) after reordering
  depend on the topmost variables in the manager. No remapping means 
  that the function(s) after reordering depend on the same variables 
  as before. Consider the following example. Suppose the initial four
  variable function depends on variables 2,4,5, and 9 on the CUDD BDD
  manager, which may be found anywhere in the current variable order.
  If remapping is set, the function after ordering depends on the 
  topmost variables in the manager, which may or may not be the same
  as the variables 2,4,5, and 9. If no remapping is set, then the 
  reordered function depend on the same variables 2,4,5, and 9, but
  the meaning of each variale has changed according to the new ordering.
  The resulting ordering is returned in the array "pOrder" filled out
  by the reordering engine in the call to Extra_Reorder(). The default
  is no remapping.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReorderSetRemapping(p *reo_man, fRemapUp int32) {
    p.fRemapUp = fRemapUp;
}

/**Function*************************************************************

  Synopsis    [Sets the number of iterations of sifting performed.]

  Description [The default is one iteration. But a higher minimization
  quality is desired, it is possible to set the number of iterations 
  to any number larger than 1. Convergence is often reached after
  several iterations, so typically it make no sense to set the number
  of iterations higher than 3.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReorderSetIterations(p *reo_man, nIters int32) {
    p.nIters = nIters;
}

/**Function*************************************************************

  Synopsis    [Sets the verification mode.]

  Description [Setting the level to 1 results in verifying the results
  of variable reordering. Verification is performed by remapping the
  resulting functions into the original variable order and comparing
  them with the original functions given by the user. Enabling verification
  typically leads to 20-30% increase in the total runtime of REO.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReorderSetVerification(p *reo_man, fVerify int32) {
    p.fVerify = fVerify;
}

/**Function*************************************************************

  Synopsis    [Sets the verbosity level.]

  Description [Setting the level to 1 results in printing statistics
  before and after the reordering.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReorderSetVerbosity(p *reo_man, fVerbose int32) {
    p.fVerbose = fVerbose;
}

/**Function*************************************************************

  Synopsis    [Performs reordering of the function.]

  Description [Returns the DD minimized by variable reordering in the REO 
  engine. Takes the CUDD decision diagram manager (dd) and the function (Func) 
  represented as a BDD or ADD (MTBDD). If the variable array (pOrder) is not NULL, 
  returns the resulting  variable permutation. The permutation is such that if the resulting 
  function is permuted by Cudd_(add,bdd)Permute() using pOrder as the permutation 
  array, the initial function (Func) results.
  Several flag set by other interface functions specify reordering options: 
  - Remappig can be set by Extra_ReorderSetRemapping(). Then the resulting DD after 
  reordering is remapped into the topmost levels of the DD manager. Otherwise, 
  the resulting DD after reordering is mapped using the same variables, on which it 
  originally depended, only (possibly) permuted as a result of reordering.
  - Minimization type can be set by Extra_ReorderSetMinimizationType(). Note
  that when the BDD is minimized for the total width of the total APL, the number
  BDD nodes can increase. The total width is defines as sum total of widths on each 
  level. The width on one level is defined as the number of distinct BDD nodes 
  pointed by the nodes situated above the given level.
  - The number of iterations of sifting can be set by Extra_ReorderSetIterations().
  The decision diagram returned by this procedure is not referenced.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_Reorder(p *reo_man, dd *DdManager, Func *DdNode, pOrder *int32) *DdNode {
    var FuncRes *DdNode;
    Extra_ReorderArray( p, dd, &Func, &FuncRes, 1, pOrder );
    Cudd_Deref( FuncRes );
    return FuncRes;
}

/**Function*************************************************************

  Synopsis    [Performs reordering of the array of functions.]

  Description [The options are similar to the procedure Extra_Reorder(), except that
  the user should also provide storage for the resulting DDs, which are returned 
  referenced.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReorderArray(p *reo_man, dd *DdManager, Funcs []*DdNode, FuncsRes []*DdNode, nFuncs int32, pOrder *int32) {
    reoReorderArray( p, dd, Funcs, FuncsRes, nFuncs, pOrder );
}


////////////////////////////////////////////////////////////////////////
///                         END OF FILE                              ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END


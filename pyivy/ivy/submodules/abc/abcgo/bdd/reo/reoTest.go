package reo

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [reoTest.c]

  PackageName [REO: A specialized DD reordering engine.]

  Synopsis    [Various testing procedures (may be outdated).]

  Author      [Alan Mishchenko <alanmi@ece.pdx.edu>]
  
  Affiliation [ECE Department. Portland State University, Portland, Oregon.]

  Date        [Ver. 1.0. Started - October 15, 2002.]

  Revision    [$Id: reoTest.c,v 1.0 2002/15/10 03:00:00 alanmi Exp $]

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

  Synopsis    [Reorders the DD using REO and CUDD.]

  Description [This function can be used to test the performance of the reordering package.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReorderTest(dd *DdManager, Func *DdNode) {
    var pReo *reo_man;
    var Temp, Temp1 *DdNode;
    var pOrder [1000]int32;

    pReo = Extra_ReorderInit( 100, 100 );

//Extra_DumpDot( dd, &Func, 1, "beforReo.dot", 0 );
    Temp  = Extra_Reorder( pReo, dd, Func, pOrder );  Cudd_Ref( Temp );
//Extra_DumpDot( dd, &Temp, 1, "afterReo.dot", 0 );

    Temp1 = Extra_ReorderCudd(dd, Func, nil );           Cudd_Ref( Temp1 );
fmt.Printf("Initial = %d. Final = %d. Cudd = %d.\n", Cudd_DagSize(Func), Cudd_DagSize(Temp), Cudd_DagSize(Temp1));
    Cudd_RecursiveDeref( dd, Temp1 );
    Cudd_RecursiveDeref( dd, Temp );
 
    Extra_ReorderQuit( pReo );
}


/**Function*************************************************************

  Synopsis    [Reorders the DD using REO and CUDD.]

  Description [This function can be used to test the performance of the reordering package.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ReorderTestArray(dd *DdManager, Funcs []*DdNode, nFuncs int32) {
    var pReo *reo_man;
    var FuncsRes [1000]*DdNode;
    var pOrder [1000]int32;
    var i int32;

    pReo = Extra_ReorderInit( 100, 100 );
    Extra_ReorderArray( pReo, dd, Funcs, FuncsRes, nFuncs, pOrder );  
    Extra_ReorderQuit( pReo );

fmt.Printf("Initial = %d. Final = %d.\n", Cudd_SharingSize(Funcs,nFuncs), Cudd_SharingSize(FuncsRes,nFuncs));

    for  i = 0; i < nFuncs; i++  {
        Cudd_RecursiveDeref( dd, FuncsRes[i] );
    }

}

/**Function*************************************************************

  Synopsis    [Reorders the DD using CUDD package.]

  Description [Transfers the DD into a temporary manager in such a way
  that the level correspondence is preserved. Reorders the manager
  and transfers the DD back into the original manager using the topmost
  levels of the manager, in such a way that the ordering of levels is
  preserved. The resulting permutation is returned in the array
  given by the user.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Extra_ReorderCudd_ddReorder *DdManager = nil;
var spStatic_Extra_ReorderCudd_Permute *int32 = nil;
var spStatic_Extra_ReorderCudd_PermuteReo1 *int32 = nil;
var spStatic_Extra_ReorderCudd_PermuteReo2 *int32 = nil;
func Extra_ReorderCudd(dd *DdManager, aFunc *DdNode, pPermuteReo []int32) *DdNode {
    
    
    
    
    var aFuncReorder, aFuncNew *DdNode;
    var lev, var int32;

    // start the reordering manager
    if spStatic_Extra_ReorderCudd_ddReorder == nil {
        spStatic_Extra_ReorderCudd_Permute       = ABC_ALLOC( int, dd.size );
        spStatic_Extra_ReorderCudd_PermuteReo1   = ABC_ALLOC( int, dd.size );
        spStatic_Extra_ReorderCudd_PermuteReo2   = ABC_ALLOC( int, dd.size );
        spStatic_Extra_ReorderCudd_ddReorder = Cudd_Init( dd.size, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
        Cudd_AutodynDisable(spStatic_Extra_ReorderCudd_ddReorder);
    }

    // determine the permutation of variable to make sure that var order in bFunc
    // will not change when this function is transfered into the new manager
    for  lev = 0; lev < dd.size; lev++  {
        spStatic_Extra_ReorderCudd_Permute[ dd.invperm[lev] ] = spStatic_Extra_ReorderCudd_ddReorder.invperm[lev];
        spStatic_Extra_ReorderCudd_PermuteReo1[ spStatic_Extra_ReorderCudd_ddReorder.invperm[lev] ] = dd.invperm[lev];
    }
    // transfer this function into the new manager in such a way that ordering of vars does not change
    aFuncReorder = Extra_TransferPermute( dd, spStatic_Extra_ReorderCudd_ddReorder, aFunc, spStatic_Extra_ReorderCudd_Permute );  Cudd_Ref( aFuncReorder );
//    assert( Cudd_DagSize(aFunc) == Cudd_DagSize(aFuncReorder)  );

    // perform the reordering
fmt.Printf("Nodes before = %d.\n", Cudd_DagSize(aFuncReorder));
    Cudd_ReduceHeap( spStatic_Extra_ReorderCudd_ddReorder, CUDD_REORDER_SYMM_SIFT, 1 );
fmt.Printf("Nodes before = %d.\n", Cudd_DagSize(aFuncReorder));

    // determine the reverse variable permutation
    for  lev = 0; lev < dd.size; lev++  {
        spStatic_Extra_ReorderCudd_Permute[ spStatic_Extra_ReorderCudd_ddReorder.invperm[lev] ] = dd.invperm[lev];
        spStatic_Extra_ReorderCudd_PermuteReo2[ dd.invperm[lev] ] = spStatic_Extra_ReorderCudd_ddReorder.invperm[lev];
    }

    // transfer this function into the new manager in such a way that ordering of vars does not change
    aFuncNew = Extra_TransferPermute( spStatic_Extra_ReorderCudd_ddReorder, dd, aFuncReorder, spStatic_Extra_ReorderCudd_Permute );  Cudd_Ref( aFuncNew );
//    assert( Cudd_DagSize(aFuncNew) == Cudd_DagSize(aFuncReorder)  );
    Cudd_RecursiveDeref( spStatic_Extra_ReorderCudd_ddReorder, aFuncReorder );

    // derive the resulting variable ordering
    if pPermuteReo {
        for  var = 0; var < dd.size; var++  {
            pPermuteReo[var] = spStatic_Extra_ReorderCudd_PermuteReo1[ spStatic_Extra_ReorderCudd_PermuteReo2[var] ];
        } }

    Cudd_Deref( aFuncNew );
    return aFuncNew;
}


/**Function*************************************************************

  Synopsis    []

  Description [Transfers the BDD into another manager minimizes it and 
  returns the min number of nodes; disposes of the BDD in the new manager.
  Useful for debugging or comparing the performance of other reordering
  procedures.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Extra_bddReorderTest_s_ddmin *DdManager;
func Extra_bddReorderTest(dd *DdManager, bF *DdNode) int32 {
    
    var bFmin *DdNode;
    var nNodes int32;
//    abctime clk1;

    if spStatic_Extra_bddReorderTest_s_ddmin == nil {
        spStatic_Extra_bddReorderTest_s_ddmin = Cudd_Init( dd.size, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0); }

//    Cudd_ShuffleHeap( s_ddmin, dd->invperm );

//    clk1 = Abc_Clock();
    bFmin = Cudd_bddTransfer( dd, spStatic_Extra_bddReorderTest_s_ddmin, bF );  Cudd_Ref( bFmin );
    Cudd_ReduceHeap(spStatic_Extra_bddReorderTest_s_ddmin,CUDD_REORDER_SIFT,1);
//    Cudd_ReduceHeap(s_ddmin,CUDD_REORDER_SYMM_SIFT,1);
    nNodes = Cudd_DagSize( bFmin );
    Cudd_RecursiveDeref( spStatic_Extra_bddReorderTest_s_ddmin, bFmin );

//    printf( "Classical variable reordering time = %.2f sec\n", (float)(Abc_Clock() - clk1)/(float)(CLOCKS_PER_SEC) );
    return nNodes;
}

/**Function*************************************************************

  Synopsis    []

  Description [Transfers the ADD into another manager minimizes it and 
  returns the min number of nodes; disposes of the BDD in the new manager.
  Useful for debugging or comparing the performance of other reordering
  procedures.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Extra_addReorderTest_s_ddmin *DdManager;
func Extra_addReorderTest(dd *DdManager, aF *DdNode) int32 {
    
    var bF *DdNode;
    var bFmin *DdNode;
    var aFmin *DdNode;
    var nNodesBeg int32;
    var nNodesEnd int32;
    var clk1 abctime;

    if spStatic_Extra_addReorderTest_s_ddmin == nil {
        spStatic_Extra_addReorderTest_s_ddmin = Cudd_Init( dd.size, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0); }

//    Cudd_ShuffleHeap( s_ddmin, dd->invperm );

    clk1 = Abc_Clock();
    bF    = Cudd_addBddPattern( dd, aF );         Cudd_Ref( bF );
    bFmin = Cudd_bddTransfer( dd, spStatic_Extra_addReorderTest_s_ddmin, bF );  Cudd_Ref( bFmin );
    Cudd_RecursiveDeref( dd, bF );
    aFmin = Cudd_BddToAdd( spStatic_Extra_addReorderTest_s_ddmin, bFmin );      Cudd_Ref( aFmin );
    Cudd_RecursiveDeref( spStatic_Extra_addReorderTest_s_ddmin, bFmin );

    nNodesBeg = Cudd_DagSize( aFmin );
    Cudd_ReduceHeap(spStatic_Extra_addReorderTest_s_ddmin,CUDD_REORDER_SIFT,1);
//    Cudd_ReduceHeap(s_ddmin,CUDD_REORDER_SYMM_SIFT,1);
    nNodesEnd = Cudd_DagSize( aFmin );
    Cudd_RecursiveDeref( spStatic_Extra_addReorderTest_s_ddmin, aFmin );

    fmt.Printf("Classical reordering of ADDs: Before = %d. After = %d.\n", nNodesBeg, nNodesEnd);
    fmt.Printf("Classical variable reordering time = %.2f sec\n", float32((Abc_Clock() - clk1))/float32((int64(time.Second))));
    return nNodesEnd;
}


////////////////////////////////////////////////////////////////////////
///                         END OF FILE                              ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END


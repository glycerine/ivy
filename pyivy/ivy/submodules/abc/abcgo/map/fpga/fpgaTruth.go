package fpga

import "fmt"

/**CFile****************************************************************

  FileName    [fpgaTruth.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Technology mapping for variable-size-LUT FPGAs.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - August 18, 2004.]

  Revision    [$Id: fpgaTruth.c,v 1.4 2005/01/23 06:59:42 alanmi Exp $]

***********************************************************************/

//#include "fpgaInt.h"
//#include "bdd/cudd/cudd.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Recursively derives the truth table for the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_TruthsCutBdd_rec(dd *DdManager, pCut *Fpga_Cut_t, vVisited *Fpga_NodeVec_t) *DdNode {
    var bFunc, bFunc0, bFunc1 *DdNode;
    // inlining of macro Fpga_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pCut) & 01)) != 0))), "!/*__starting_point_inline_macro:Fpga_IsComplement__*/((int32(((ABC_PTRUINT_T) (pCut) & 01))))");
    // if the cut is visited, return the result
    if pCut.uSign {
        return (DdNode *)(ABC_PTRUINT_T)pCut.uSign; }
    // compute the functions of the children
    // inlining of macro Fpga_CutRegular
    bFunc0 = Fpga_TruthsCutBdd_rec( dd, ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pCut.pOne) & ~01)) , vVisited );   Cudd_Ref( bFunc0 );
    // inlining of macro Cudd_NotCond
    // inlining of macro Fpga_CutIsComplement
    bFunc0 = ((DdNode *)((ptrint)(bFunc0) ^ (((int32(((ABC_PTRUINT_T)(pCut.pOne) & 01)))))));
    // inlining of macro Fpga_CutRegular
    bFunc1 = Fpga_TruthsCutBdd_rec( dd, ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pCut.pTwo) & ~01)) , vVisited );   Cudd_Ref( bFunc1 );
    // inlining of macro Cudd_NotCond
    // inlining of macro Fpga_CutIsComplement
    bFunc1 = ((DdNode *)((ptrint)(bFunc1) ^ (((int32(((ABC_PTRUINT_T)(pCut.pTwo) & 01)))))));
    // get the function of the cut
    bFunc  = Cudd_bddAnd( dd, bFunc0, bFunc1 );   Cudd_Ref( bFunc );
    // inlining of macro Cudd_NotCond
    bFunc  = ((DdNode *)((ptrint)(bFunc) ^ (pCut.Phase)));
    Cudd_RecursiveDeref( dd, bFunc0 );
    Cudd_RecursiveDeref( dd, bFunc1 );
    assert(pCut.uSign == 0, "pCut.uSign == 0");
    pCut.uSign = uint32((ABC_PTRUINT_T))bFunc;
    // add this cut to the visited list
    Fpga_NodeVecPush( vVisited, (Fpga_Node_t *)pCut );
    return bFunc;
}

/**Function*************************************************************

  Synopsis    [Derives the truth table for one cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_TruthsCutBdd(dd any, pCut *Fpga_Cut_t) any {
    var vVisited *Fpga_NodeVec_t;
    var bFunc *DdNode;
    var i int32;
    assert(pCut.nLeaves > 1, "pCut.nLeaves > 1");
    // set the leaf variables
    for  i = 0; i < pCut.nLeaves; i++  {
        pCut.ppLeaves[i].pCuts.uSign = uint32((ABC_PTRUINT_T))Cudd_bddIthVar( (DdManager *)dd, i );
    }
    // recursively compute the function
    vVisited = Fpga_NodeVecAlloc( 10 );
    bFunc = Fpga_TruthsCutBdd_rec( (DdManager *)dd, pCut, vVisited );   Cudd_Ref( bFunc );
    // clean the intermediate BDDs
    for  i = 0; i < pCut.nLeaves; i++  {
        pCut.ppLeaves[i].pCuts.uSign = 0;
    }
    for  i = 0; i < vVisited.nSize; i++  {
        pCut = (Fpga_Cut_t *)vVisited.pArray[i];
        Cudd_RecursiveDeref( (DdManager *)dd, (DdNode*)(ABC_PTRUINT_T)pCut.uSign );
        pCut.uSign = 0;
    }
//    printf( "%d ", vVisited->nSize );
    Fpga_NodeVecFree( vVisited );
    Cudd_Deref( bFunc );
    return bFunc;
}


/**Function*************************************************************

  Synopsis    [Recursively derives the truth table for the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutVolume_rec(pCut *Fpga_Cut_t, vVisited *Fpga_NodeVec_t) {
    // inlining of macro Fpga_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pCut) & 01)) != 0))), "!/*__starting_point_inline_macro:Fpga_IsComplement__*/((int32(((ABC_PTRUINT_T) (pCut) & 01))))");
    if pCut.fMark {
        return; }
    pCut.fMark = 1;
    // inlining of macro Fpga_CutRegular
    Fpga_CutVolume_rec( ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pCut.pOne) & ~01)) , vVisited );
    // inlining of macro Fpga_CutRegular
    Fpga_CutVolume_rec( ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pCut.pTwo) & ~01)) , vVisited );
    Fpga_NodeVecPush( vVisited, (Fpga_Node_t *)pCut );
}

/**Function*************************************************************

  Synopsis    [Derives the truth table for one cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutVolume(pCut *Fpga_Cut_t) int32 {
    var vVisited *Fpga_NodeVec_t;
    var Volume, i int32;
    assert(pCut.nLeaves > 1, "pCut.nLeaves > 1");
    // set the leaf variables
    for  i = 0; i < pCut.nLeaves; i++  {
        pCut.ppLeaves[i].pCuts.fMark = 1;
    }
    // recursively compute the function
    vVisited = Fpga_NodeVecAlloc( 10 );
    Fpga_CutVolume_rec( pCut, vVisited ); 
    // clean the marks
    for  i = 0; i < pCut.nLeaves; i++  {
        pCut.ppLeaves[i].pCuts.fMark = 0;
    }
    for  i = 0; i < vVisited.nSize; i++  {
        pCut = (Fpga_Cut_t *)vVisited.pArray[i];
        pCut.fMark = 0;
    }
    Volume = vVisited.nSize;
    fmt.Printf("%d ", Volume);
    Fpga_NodeVecFree( vVisited );
    return Volume;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


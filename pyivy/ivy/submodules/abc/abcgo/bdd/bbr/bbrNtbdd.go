package bbr

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [bbrNtbdd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD-based reachability analysis.]

  Synopsis    [Procedures to construct global BDDs for the network.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bbrNtbdd.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bbr.h"
//#include "bar.h"

//ABC_NAMESPACE_IMPL_START


typedef char ProgressBar;

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Aig_ObjSetGlobalBdd(pObj *Aig_Obj_t, bFunc *DdNode) { pObj.pData = bFunc; }
func Aig_ObjCleanGlobalBdd(dd *DdManager, pObj *Aig_Obj_t) { Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData ); pObj.pData = nil; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derives the global BDD for one AIG node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbr_NodeGlobalBdds_rec(dd *DdManager, pNode *Aig_Obj_t, nBddSizeMax int32, fDropInternal int32, pProgress *ProgressBar, pCounter *int32, fVerbose int32) *DdNode {
    var bFunc, bFunc0, bFunc1 *DdNode;
    assert(!Aig_IsComplement(pNode), "!Aig_IsComplement(pNode)");
    if Cudd_ReadKeys(dd)-Cudd_ReadDead(dd) > uint32(nBddSizeMax) {
//        Extra_ProgressBarStop( pProgress );
        if fVerbose != 0 {
        fmt.Printf("The number of live nodes reached %d.\n", nBddSizeMax); }
        fflush( os.Stdout );
        return nil;
    }
    // if the result is available return
    if Aig_ObjGlobalBdd(pNode) == nil {
        // compute the result for both branches
        bFunc0 = Bbr_NodeGlobalBdds_rec( dd, Aig_ObjFanin0(pNode), nBddSizeMax, fDropInternal, pProgress, pCounter, fVerbose ); 
        if bFunc0 == nil {
            return nil; }
        Cudd_Ref( bFunc0 );
        bFunc1 = Bbr_NodeGlobalBdds_rec( dd, Aig_ObjFanin1(pNode), nBddSizeMax, fDropInternal, pProgress, pCounter, fVerbose ); 
        if bFunc1 == nil {
            return nil; }
        Cudd_Ref( bFunc1 );
        // inlining of macro Cudd_NotCond
        bFunc0 = ((DdNode *)((ptrint)(bFunc0) ^ (Aig_ObjFaninC0(pNode))));
        // inlining of macro Cudd_NotCond
        bFunc1 = ((DdNode *)((ptrint)(bFunc1) ^ (Aig_ObjFaninC1(pNode))));
        // get the final result
        bFunc = Cudd_bddAnd( dd, bFunc0, bFunc1 );   Cudd_Ref( bFunc );
        Cudd_RecursiveDeref( dd, bFunc0 );
        Cudd_RecursiveDeref( dd, bFunc1 );
        // add the number of used nodes
        (*pCounter)++;
        // set the result
        assert(Aig_ObjGlobalBdd(pNode) == nil, "Aig_ObjGlobalBdd(pNode) == nil");
        Aig_ObjSetGlobalBdd( pNode, bFunc );
        // increment the progress bar
//        if ( pProgress )
//            Extra_ProgressBarUpdate( pProgress, *pCounter, NULL );
    }
    // prepare the return value
    bFunc = Aig_ObjGlobalBdd(pNode);
    // dereference BDD at the node
    if --pNode.nRefs == 0 && fDropInternal {
        Cudd_Deref( bFunc );
        Aig_ObjSetGlobalBdd( pNode, nil );
    }
    return bFunc;
}

/**Function*************************************************************

  Synopsis    [Frees the global BDDs of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFreeGlobalBdds(p *Aig_Man_t, dd *DdManager) { 
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManForEachObj( p, pObj, i )
        if ( Aig_ObjGlobalBdd(pObj) )
            Aig_ObjCleanGlobalBdd( dd, pObj );
}

/**Function*************************************************************

  Synopsis    [Returns the shared size of global BDDs of the COs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSizeOfGlobalBdds(p *Aig_Man_t) int32 {
    var vFuncsGlob *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var RetValue, i int32;
    // complement the global functions
    vFuncsGlob = Vec_PtrAlloc( Aig_ManCoNum(p) );
    Aig_ManForEachCo( p, pObj, i )
        Vec_PtrPush( vFuncsGlob, Aig_ObjGlobalBdd(pObj) );
    RetValue = Cudd_SharingSize( (DdNode **)Vec_PtrArray(vFuncsGlob), Vec_PtrSize(vFuncsGlob) );
    Vec_PtrFree( vFuncsGlob );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Recursively computes global BDDs for the AIG in the manager.]

  Description [On exit, BDDs are stored in the pNode->pData fields.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManComputeGlobalBdds(p *Aig_Man_t, nBddSizeMax int32, fDropInternal int32, fReorder int32, fVerbose int32) *DdManager {
    var pProgress *ProgressBar  = nil;
    var pObj *Aig_Obj_t;
    var dd *DdManager;
    var bFunc *DdNode;
    var i, Counter int32;
    // start the manager
    dd = Cudd_Init( Aig_ManCiNum(p), 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    // set reordering
    if fReorder != 0 {
        Cudd_AutodynEnable( dd, CUDD_REORDER_SYMM_SIFT ); }
    // prepare to construct global BDDs
    Aig_ManCleanData( p );
    // assign the constant node BDD
    Aig_ObjSetGlobalBdd( Aig_ManConst1(p), dd.one );   Cudd_Ref( dd.one );
    // set the elementary variables
    Aig_ManForEachCi( p, pObj, i )
    {
        Aig_ObjSetGlobalBdd( pObj, dd.vars[i] );  Cudd_Ref( dd.vars[i] );
    }

    // collect the global functions of the COs
    Counter = 0;
    // construct the BDDs
//    pProgress = Extra_ProgressBarStart( stdout, Aig_ManNodeNum(p) );
    Aig_ManForEachCo( p, pObj, i )
    {
        bFunc = Bbr_NodeGlobalBdds_rec( dd, Aig_ObjFanin0(pObj), nBddSizeMax, fDropInternal, pProgress, &Counter, fVerbose );
        if bFunc == nil {
            if fVerbose != 0 {
            fmt.Printf("Constructing global BDDs is aborted.\n"); }
            Aig_ManFreeGlobalBdds( p, dd );
            Cudd_Quit( dd ); 
            // reset references
            Aig_ManResetRefs( p );
            return nil;
        }
        // inlining of macro Cudd_NotCond
        bFunc = ((DdNode *)((ptrint)(bFunc) ^ (Aig_ObjFaninC0(pObj))));  Cudd_Ref( bFunc ); 
        Aig_ObjSetGlobalBdd( pObj, bFunc );
    }
//    Extra_ProgressBarStop( pProgress );
    // reset references
    Aig_ManResetRefs( p );
    // reorder one more time
    if fReorder != 0 {
        Cudd_ReduceHeap( dd, CUDD_REORDER_SYMM_SIFT, 1 );
        Cudd_AutodynDisable( dd );
    }
//    Cudd_PrintInfo( dd, stdout );
    return dd;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


package ret

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [retLvalue.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Retiming package.]

  Synopsis    [Implementation of Pan's retiming algorithm.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Oct 31, 2006.]

  Revision    [$Id: retLvalue.c,v 1.00 2006/10/31 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "retInt.h"

ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// node status after updating its arrival time
const ( ABC_RET_UPDATE_FAIL = iota
 ABC_RET_UPDATE_NO
 ABC_RET_UPDATE_YES 
)


// the internal procedures
// static Vec_Int_t * Abc_NtkRetimeGetLags( Abc_Ntk_t * pNtk, int nIterLimit, int fVerbose );

// static int         Abc_NtkRetimeSearch_rec( Abc_Ntk_t * pNtk, Vec_Ptr_t * vNodes, Vec_Ptr_t * vLatches, int FiMin, int FiMax, int nMaxIters, int fVerbose );

// static int         Abc_NtkRetimeForPeriod( Abc_Ntk_t * pNtk, Vec_Ptr_t * vNodes, Vec_Ptr_t * vLatches, int Fi, int nMaxIters, int fVerbose );

// static int         Abc_NtkRetimeUpdateLValue( Abc_Ntk_t * pNtk, Vec_Ptr_t * vNodes, Vec_Ptr_t * vLatches, int Fi );

// static int         Abc_NtkRetimePosOverLimit( Abc_Ntk_t * pNtk, int Fi );

// static Vec_Ptr_t * Abc_ManCollectLatches( Abc_Ntk_t * pNtk );

// static int         Abc_NtkRetimeUsingLags( Abc_Ntk_t * pNtk, Vec_Int_t * vLags, int fVerbose );


func Abc_NodeComputeLag(LValue int32, Fi int32) int32 { return (LValue + (1<<16)*Fi)/Fi - (1<<16) - int32((LValue % Fi == 0));     }
func Abc_NodeGetLValue(pNode *Abc_Obj_t) int32 { return int32((ABC_PTRUINT_T))pNode.pCopy;        }
func Abc_NodeSetLValue(pNode *Abc_Obj_t, Value int32) { pNode.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)Value;    }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Implements Pan's retiming algorithm.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeLValue(pNtk *Abc_Ntk_t, nIterLimit int32, fVerbose int32) int32 {
    var vLags *Vec_Int_t;
    var nLatches int32  = Abc_NtkLatchNum(pNtk);
    assert(Abc_NtkIsLogic( pNtk ), "Abc_NtkIsLogic( pNtk )");
    // get the lags
    vLags = Abc_NtkRetimeGetLags( pNtk, nIterLimit, fVerbose );
    // compute the retiming
//    Abc_NtkRetimeUsingLags( pNtk, vLags, fVerbose );
    Vec_IntFree( vLags );
    // fix the COs
//    Abc_NtkLogicMakeSimpleCos( pNtk, 0 );
    // check for correctness
    if !Abc_NtkCheck( pNtk ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkRetimeLValue(): Network check has failed.\n"); }
    // return the number of latches saved
    return nLatches - Abc_NtkLatchNum(pNtk);
}

/**Function*************************************************************

  Synopsis    [Computes the retiming lags.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeGetLags(pNtk *Abc_Ntk_t, nIterLimit int32, fVerbose int32) *Vec_Int_t {
    var vLags *Vec_Int_t;
    var vNodes, vLatches *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var i, FiMax, FiBest, RetValue int32;
    var clk, clkIter abctime;
    var NodeLag byte;

    // get the upper bound on the clock period
    FiMax = Abc_NtkLevel(pNtk);

    // make sure this clock period is feasible
    vNodes = Abc_NtkDfs( pNtk, 0 );
    vLatches = Abc_ManCollectLatches( pNtk );
    if !(Abc_NtkRetimeForPeriod( pNtk, vNodes, vLatches, FiMax, nIterLimit, fVerbose ) != 0) {
        Vec_PtrFree( vLatches );
        Vec_PtrFree( vNodes );
        fmt.Printf("Abc_NtkRetimeGetLags() error: The upper bound on the clock period cannot be computed.\n");
        return Vec_IntStart( Abc_NtkObjNumMax(pNtk) + 1 );
    }
 
    // search for the optimal clock period between 0 and nLevelMax
clk = Abc_Clock();
    FiBest = Abc_NtkRetimeSearch_rec( pNtk, vNodes, vLatches, 0, FiMax, nIterLimit, fVerbose );
clkIter = Abc_Clock() - clk;

    // recompute the best l-values
    RetValue = Abc_NtkRetimeForPeriod( pNtk, vNodes, vLatches, FiBest, nIterLimit, fVerbose );
    assert((RetValue != 0), "RetValue");

    // fix the problem with non-converged delays
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else
        if Abc_NodeGetLValue(pNode) < -ABC_INFINITY/2 {
            Abc_NodeSetLValue( pNode, 0 ); }
    }

    // write the retiming lags
    vLags = Vec_IntStart( Abc_NtkObjNumMax(pNtk) + 1 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        NodeLag = Abc_NodeComputeLag( Abc_NodeGetLValue(pNode), FiBest );
        Vec_IntWriteEntry( vLags, pNode.Id, NodeLag );
    }
    }
/*
    Abc_NtkForEachPo( pNtk, pNode, i )
        printf( "%d ", Abc_NodeGetLValue(Abc_ObjFanin0(pNode)) );
    printf( "\n" );
    Abc_NtkForEachLatch( pNtk, pNode, i )
        printf( "%d/%d ", Abc_NodeGetLValue(Abc_ObjFanout0(pNode)), Abc_NodeGetLValue(Abc_ObjFanout0(pNode)) + FiBest );
    printf( "\n" );
*/

    // print the result
//    if ( fVerbose )
    fmt.Printf("The best clock period is %3d. (Currently, network is not modified.)\n", FiBest);
/*
    {
        FILE * pTable;
        pTable = fopen( "iscas/seqmap__stats2.txt", "a+" );
        fprintf( pTable, "%d ", FiBest );
        fprintf( pTable, "\n" );
        fclose( pTable );
    }
*/
    Vec_PtrFree( vNodes );
    Vec_PtrFree( vLatches );
    return vLags;
}

/**Function*************************************************************

  Synopsis    [Performs binary search for the optimal clock period.]

  Description [Assumes that FiMin is infeasible while FiMax is feasible.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeSearch_rec(pNtk *Abc_Ntk_t, vNodes *Vec_Ptr_t, vLatches *Vec_Ptr_t, FiMin int32, FiMax int32, nMaxIters int32, fVerbose int32) int32 {
    var Median int32;
    assert(FiMin < FiMax, "FiMin < FiMax");
    if FiMin + 1 == FiMax {
        return FiMax; }
    Median = FiMin + (FiMax - FiMin)/2;
    if Abc_NtkRetimeForPeriod( pNtk, vNodes, vLatches, Median, nMaxIters, fVerbose ) != 0 {
        return Abc_NtkRetimeSearch_rec( pNtk, vNodes, vLatches, FiMin, Median, nMaxIters, fVerbose ); // Median is feasible
    } else { 
        return Abc_NtkRetimeSearch_rec( pNtk, vNodes, vLatches, Median, FiMax, nMaxIters, fVerbose ); } // Median is infeasible
}

/**Function*************************************************************

  Synopsis    [Returns 1 if retiming with this clock period is feasible.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeForPeriod(pNtk *Abc_Ntk_t, vNodes *Vec_Ptr_t, vLatches *Vec_Ptr_t, Fi int32, nMaxIters int32, fVerbose int32) int32 {
    var pObj *Abc_Obj_t;
    var c, i, fConverged int32;
    // set l-values of all nodes to be minus infinity, except PIs and constants
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else
        if Abc_ObjFaninNum(pObj) == 0 {
            Abc_NodeSetLValue( pObj, 0 );
        } else {
            Abc_NodeSetLValue( pObj, -ABC_INFINITY ); }
    }
    // update all values iteratively
    fConverged = 0;
    for  c = 1; c <= nMaxIters; c++  {
        if !(Abc_NtkRetimeUpdateLValue( pNtk, vNodes, vLatches, Fi ) != 0) {
            fConverged = 1;
            break;
        }
        if Abc_NtkRetimePosOverLimit(pNtk, Fi) != 0 {
            break; }
    }
    // report the results
    if fVerbose != 0 {
        if !(fConverged != 0) {
            fmt.Printf("Period = %3d.  Iterations = %3d.    Infeasible %s\n", Fi, c, (c > nMaxIters)? "(timeout)" : "");
        } else {
            fmt.Printf("Period = %3d.  Iterations = %3d.      Feasible\n", Fi, c); }
    }
/*
    // check if any AND gates have infinite delay
    Counter = 0;
    Abc_NtkForEachNode( pNtk, pObj, i )
        Counter += (Abc_NodeGetLValue(pObj) < -ABC_INFINITY/2);
    if ( Counter > 0 )
        printf( "Warning: %d internal nodes have wrong l-values!\n", Counter );
*/
    return fConverged;
}

/**Function*************************************************************

  Synopsis    [Performs one iteration of l-value computation for the nodes.]

  Description [Experimentally it was found that checking POs changes
  is not enough to detect the convergence of l-values in the network.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeUpdateLValue(pNtk *Abc_Ntk_t, vNodes *Vec_Ptr_t, vLatches *Vec_Ptr_t, Fi int32) int32 {
    var pObj, pFanin *Abc_Obj_t;
    var i, k, lValueNew, fChange int32;
    // go through the nodes and detect change
    fChange = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
        lValueNew = -ABC_INFINITY;
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if lValueNew < Abc_NodeGetLValue(pFanin) {
                lValueNew = Abc_NodeGetLValue(pFanin); }
        }
        lValueNew++;
        if Abc_NodeGetLValue(pObj) < lValueNew {
            Abc_NodeSetLValue( pObj, lValueNew );
            fChange = 1;
        }
    }
    // propagate values through the latches
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLatches)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLatches, i)), 1); i++  {
        Abc_NodeSetLValue( Abc_ObjFanout0(pObj), Abc_NodeGetLValue(Abc_ObjFanin0(Abc_ObjFanin0(pObj))) - Fi );
    }
    return fChange;
}

/**Function*************************************************************

  Synopsis    [Detects the case when l-values exceeded the limit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimePosOverLimit(pNtk *Abc_Ntk_t, Fi int32) int32 {
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        if Abc_NodeGetLValue(Abc_ObjFanin0(pObj)) > Fi {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Collects latches in the topological order.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Abc_ManCollectLatches_rec(pObj *Abc_Obj_t, vLatches *Vec_Ptr_t) {
    var pDriver *Abc_Obj_t;
    if !Abc_ObjIsLatch(pObj) {
        return; }
    // skip already collected latches
    if Abc_NodeIsTravIdCurrent(pObj) {
        return; }
    Abc_NodeSetTravIdCurrent(pObj);
    // get the driver node feeding into the latch
    pDriver = Abc_ObjFanin0(Abc_ObjFanin0(pObj));
    // call recursively if the driver looks like a latch output
    if Abc_ObjIsBo(pDriver) {
        Abc_ManCollectLatches_rec( Abc_ObjFanin0(pDriver), vLatches ); }
    // collect the latch
    Vec_PtrPush( vLatches, pObj );
}

/**Function*************************************************************

  Synopsis    [Collects latches in the topological order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManCollectLatches(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vLatches *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    vLatches = Vec_PtrAlloc( Abc_NtkLatchNum(pNtk) );
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        Abc_ManCollectLatches_rec( pObj, vLatches ); }
    }
    assert(Vec_PtrSize(vLatches) == Abc_NtkLatchNum(pNtk), "Vec_PtrSize(vLatches) == Abc_NtkLatchNum(pNtk)");
    return vLatches;
}

/**Function*************************************************************

  Synopsis    [Implements the retiming given as the array of retiming lags.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeUsingLags(pNtk *Abc_Ntk_t, vLags *Vec_Int_t, fVerbose int32) int32 {
    var pObj *Abc_Obj_t;
    var fChanges, fForward, nTotalMoves, Lag, Counter, i int32;
    // iterate over the nodes
    nTotalMoves = 0;
    do {
        fChanges = 0;
        // inlining of macro Abc_NtkForEachNode
        for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
            Lag = Vec_IntEntry( vLags, pObj.Id );
            if !Lag {
                continue; }
            fForward = (Lag < 0);
            if Abc_NtkRetimeNodeIsEnabled( pObj, fForward ) {
                Abc_NtkRetimeNode( pObj, fForward, 0 );
                fChanges = 1;
                nTotalMoves++;
                Vec_IntAddToEntry( vLags, pObj.Id, fForward? 1 : -1 );
            }
        }
        }
    } while ( fChanges );
    if fVerbose != 0 {
        fmt.Printf("Total latch moves = %d.\n", nTotalMoves); }
    // check if there are remaining lags
    Counter = 0;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Counter += (Vec_IntEntry( vLags, pObj.Id ) != 0); }
    }
    if Counter {
        fmt.Printf("Warning! The number of nodes with unrealized lag = %d.\n", Counter); }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


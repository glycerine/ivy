package llb

import "fmt"

/**CFile****************************************************************

  FileName    [llb2Sweep.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Non-linear quantification scheduling.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb2Sweep.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"

//ABC_NAMESPACE_IMPL_START
 

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Find good static variable ordering.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4SweepOrder_rec(pAig *Aig_Man_t, pObj *Aig_Obj_t, vOrder *Vec_Int_t, pCounter *int32, fSaveAll int32) {
    var pFanin0, pFanin1 *Aig_Obj_t;
    if Aig_ObjIsTravIdCurrent(pAig, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent( pAig, pObj );
    assert(Llb_ObjBddVar(vOrder, pObj) < 0, "Llb_ObjBddVar(vOrder, pObj) < 0");
    if Aig_ObjIsCi(pObj) {
        Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), (*pCounter)++ );
        return;
    }
    // try fanins with higher level first
    pFanin0 = Aig_ObjFanin0(pObj);
    pFanin1 = Aig_ObjFanin1(pObj);
//    if ( pFanin0->Level > pFanin1->Level || (pFanin0->Level == pFanin1->Level && pFanin0->Id < pFanin1->Id) )
    if pFanin0.Level > pFanin1.Level {
        Llb_Nonlin4SweepOrder_rec( pAig, pFanin0, vOrder, pCounter, fSaveAll );
        Llb_Nonlin4SweepOrder_rec( pAig, pFanin1, vOrder, pCounter, fSaveAll );
    
    } else {
        Llb_Nonlin4SweepOrder_rec( pAig, pFanin1, vOrder, pCounter, fSaveAll );
        Llb_Nonlin4SweepOrder_rec( pAig, pFanin0, vOrder, pCounter, fSaveAll );
    }
    if (fSaveAll != 0) || pObj.fMarkA {
        Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), (*pCounter)++ ); }
}

/**Function*************************************************************

  Synopsis    [Find good static variable ordering.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4SweepOrder(pAig *Aig_Man_t, pCounter *int32, fSaveAll int32) *Vec_Int_t {
    var vOrder *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // collect nodes in the order
    vOrder = Vec_IntStartFull( Aig_ManObjNumMax(pAig) );
    Aig_ManIncrementTravId( pAig );
    Aig_ObjSetTravIdCurrent( pAig, Aig_ManConst1(pAig) );
    Aig_ManForEachCo( pAig, pObj, i )
    {
        Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), Counter++ );
        Llb_Nonlin4SweepOrder_rec( pAig, Aig_ObjFanin0(pObj), vOrder, &Counter, fSaveAll );
    }
    Aig_ManForEachCi( pAig, pObj, i )
        if ( Llb_ObjBddVar(vOrder, pObj) < 0 )
            Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), Counter++ );
//    assert( Counter == Aig_ManObjNum(pAig) - 1 ); // no dangling nodes
    if pCounter != nil {
        *pCounter = Counter - Aig_ManCiNum(pAig) - Aig_ManCoNum(pAig); }
    return vOrder;
}


/**Function*************************************************************

  Synopsis    [Performs BDD sweep on the netlist.]

  Description [Returns AIG with internal cut points labeled with fMarkA.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb4_Nonlin4SweepCutpoints(pAig *Aig_Man_t, vOrder *Vec_Int_t, nBddLimit int32, fVerbose int32) int32 {
    var dd *DdManager;
    var bFunc0, bFunc1, bFunc *DdNode;
    var pObj *Aig_Obj_t;
    var i, Counter, Counter1 int32;
    dd = Cudd_Init( Aig_ManObjNumMax(pAig), 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    // assign elementary variables
    Aig_ManCleanData( pAig );
    Aig_ManForEachCi( pAig, pObj, i )
        pObj.pData = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) );
    // sweep internal nodes
    Aig_ManForEachNode( pAig, pObj, i )
    {
/*
        if ( pObj->nRefs >= 4 )
        {
            bFunc = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) );  Cudd_Ref( bFunc );
            pObj->pData = bFunc;
            Counter1++;
            continue;
        }
*/
        // inlining of macro Cudd_NotCond
        bFunc0 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin0(pObj).pData) ^ (Aig_ObjFaninC0(pObj))));
        // inlining of macro Cudd_NotCond
        bFunc1 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin1(pObj).pData) ^ (Aig_ObjFaninC1(pObj))));
        bFunc  = Cudd_bddAnd( dd, bFunc0, bFunc1 );  Cudd_Ref( bFunc );
        if Cudd_DagSize(bFunc) > nBddLimit {
//            if ( fVerbose )
//                printf( "Node %5d : Beg =%5d. ", i, Cudd_DagSize(bFunc) );

            // add cutpoint at a larger one
            Cudd_RecursiveDeref( dd, bFunc );
            if Cudd_DagSize(bFunc0) >= Cudd_DagSize(bFunc1) {
                Cudd_RecursiveDeref( dd, (DdNode *)Aig_ObjFanin0(pObj).pData );
                bFunc = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, Aig_ObjFanin0(pObj)) );
                Aig_ObjFanin0(pObj).pData = bFunc;  Cudd_Ref( bFunc );
                Aig_ObjFanin0(pObj).fMarkA = 1;

//                if ( fVerbose )
//                    printf( "Ref =%3d  ", Aig_ObjFanin0(pObj)->nRefs );
            
            } else {
                Cudd_RecursiveDeref( dd, (DdNode *)Aig_ObjFanin1(pObj).pData );
                bFunc = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, Aig_ObjFanin1(pObj)) );
                Aig_ObjFanin1(pObj).pData = bFunc;  Cudd_Ref( bFunc );
                Aig_ObjFanin1(pObj).fMarkA = 1;

//                if ( fVerbose )
//                    printf( "Ref =%3d  ", Aig_ObjFanin1(pObj)->nRefs );
            }
            // perform new operation
            // inlining of macro Cudd_NotCond
            bFunc0 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin0(pObj).pData) ^ (Aig_ObjFaninC0(pObj))));
            // inlining of macro Cudd_NotCond
            bFunc1 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin1(pObj).pData) ^ (Aig_ObjFaninC1(pObj))));
            bFunc  = Cudd_bddAnd( dd, bFunc0, bFunc1 );  Cudd_Ref( bFunc );
//            assert( Cudd_DagSize(bFunc) <= nBddLimit );

//            if ( fVerbose )
//                printf( "End =%5d.\n", Cudd_DagSize(bFunc) );
            Counter++;
        }
        pObj.pData = bFunc;
//printf( "%d ", Cudd_DagSize(bFunc) );
    }
//printf( "\n" );
    // clean up
    Aig_ManForEachNode( pAig, pObj, i )
        Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData );
    Extra_StopManager( dd );
//    Aig_ManCleanMarkA( pAig );
    if fVerbose != 0 {
    fmt.Printf("Added %d cut points.  Used %d high fanout points.\n", Counter, Counter1); }
    return Counter + Counter1;
}

/**Function*************************************************************

  Synopsis    [Derives BDDs for the partitions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4SweepPartitions_rec(dd *DdManager, pObj *Aig_Obj_t, vOrder *Vec_Int_t, vRoots *Vec_Ptr_t) *DdNode {
    var bBdd, bBdd0, bBdd1, bPart, vVar *DdNode;
    if Aig_ObjIsConst1(pObj) {
        return Cudd_ReadOne(dd); } 
    if Aig_ObjIsCi(pObj) {
        return Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) ); }
    if pObj.pData {
        return (DdNode *)pObj.pData; }
    if Aig_ObjIsCo(pObj) {
        // inlining of macro Cudd_NotCond
        bBdd0 = ((DdNode *)((ptrint)(Llb_Nonlin4SweepPartitions_rec(dd, Aig_ObjFanin0(pObj), vOrder, vRoots)) ^ (Aig_ObjFaninC0(pObj))));
        bPart = Cudd_bddXnor( dd, Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) ), bBdd0 );  Cudd_Ref( bPart );
        Vec_PtrPush( vRoots, bPart );
        return nil;
    }
    // inlining of macro Cudd_NotCond
    bBdd0 = ((DdNode *)((ptrint)(Llb_Nonlin4SweepPartitions_rec(dd, Aig_ObjFanin0(pObj), vOrder, vRoots)) ^ (Aig_ObjFaninC0(pObj))));
    // inlining of macro Cudd_NotCond
    bBdd1 = ((DdNode *)((ptrint)(Llb_Nonlin4SweepPartitions_rec(dd, Aig_ObjFanin1(pObj), vOrder, vRoots)) ^ (Aig_ObjFaninC1(pObj))));
    bBdd  = Cudd_bddAnd( dd, bBdd0, bBdd1 );   Cudd_Ref( bBdd );
    if Llb_ObjBddVar(vOrder, pObj) >= 0 {
        vVar  = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) );
        bPart = Cudd_bddXnor( dd, vVar, bBdd );  Cudd_Ref( bPart );
        Vec_PtrPush( vRoots, bPart );
        Cudd_RecursiveDeref( dd, bBdd );
        bBdd = vVar;  Cudd_Ref( vVar );
    }
    pObj.pData = bBdd;
    return bBdd;
}

/**Function*************************************************************

  Synopsis    [Derives BDDs for the partitions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4SweepPartitions(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t, fTransition int32) *Vec_Ptr_t {
    var vRoots *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManCleanData( pAig );
    vRoots = Vec_PtrAlloc( 100 );
    if fTransition != 0 {
        Saig_ManForEachLi( pAig, pObj, i )
            Llb_Nonlin4SweepPartitions_rec( dd, pObj, vOrder, vRoots );
    
    } else {
        Saig_ManForEachPo( pAig, pObj, i )
            Llb_Nonlin4SweepPartitions_rec( dd, pObj, vOrder, vRoots );
    }
    Aig_ManForEachNode( pAig, pObj, i )
        if ( pObj.pData )
            Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData );
    return vRoots;
}

/**Function*************************************************************

  Synopsis    [Get bad state monitor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb4_Nonlin4SweepBadMonitor(pAig *Aig_Man_t, vOrder *Vec_Int_t, dd *DdManager) *DdNode {
    var pObj *Aig_Obj_t;
    var bRes, bVar, bTemp *DdNode;
    var i int32;
    var TimeStop abctime;
    TimeStop = dd.TimeStop;  dd.TimeStop = 0;
    bRes = Cudd_ReadOne( dd );   Cudd_Ref( bRes );
    Saig_ManForEachPo( pAig, pObj, i )
    {
        bVar = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) );
        // inlining of macro Cudd_Not
        bRes = Cudd_bddAnd( dd, bTemp = bRes, ((DdNode *)((ptrint)(bVar) ^ 01)) );  Cudd_Ref( bRes );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    Cudd_Deref( bRes );
    dd.TimeStop = TimeStop;
    // inlining of macro Cudd_Not
    return ((DdNode *)((ptrint)(bRes) ^ 01));
}

/**Function*************************************************************

  Synopsis    [Creates quantifiable variables for both types of traversal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4SweepVars2Q(pAig *Aig_Man_t, vOrder *Vec_Int_t, fAddLis int32) *Vec_Int_t {
    var vVars2Q *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vVars2Q = Vec_IntAlloc( 0 );
    Vec_IntFill( vVars2Q, Aig_ManObjNumMax(pAig), 1 );
    // add flop outputs
    Saig_ManForEachLo( pAig, pObj, i )
        Vec_IntWriteEntry( vVars2Q, Llb_ObjBddVar(vOrder, pObj), 0 );
    // add flop inputs
    if fAddLis != 0 {
    Saig_ManForEachLi( pAig, pObj, i )
        Vec_IntWriteEntry( vVars2Q, Llb_ObjBddVar(vOrder, pObj), 0 ); }
    return vVars2Q;
}

/**Function*************************************************************

  Synopsis    [Multiply every partition by the cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4SweepDeref(dd *DdManager, vParts *Vec_Ptr_t) {
    var bFunc *DdNode;
    var i int32;
    Vec_PtrForEachEntry( DdNode *, vParts, bFunc, i )
        Cudd_RecursiveDeref( dd, bFunc );
    Vec_PtrFree( vParts );
}

/**Function*************************************************************

  Synopsis    [Multiply every partition by the cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4SweepPrint(vFuncs *Vec_Ptr_t) {
    var bFunc *DdNode;
    var i int32;
    fmt.Printf("(%d) ", Vec_PtrSize(vFuncs));
    Vec_PtrForEachEntry( DdNode *, vFuncs, bFunc, i )
        fmt.Printf("%d ", Cudd_DagSize(bFunc));
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Computes bad states.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb4_Nonlin4SweepBadStates(pAig *Aig_Man_t, vOrder *Vec_Int_t, nVars int32) *DdManager {
    var dd *DdManager;
    var vParts *Vec_Ptr_t;
    var vVars2Q *Vec_Int_t;
    var bMonitor, bImage *DdNode;
    // get quantifiable variables
    vVars2Q = Llb_Nonlin4SweepVars2Q( pAig, vOrder, 0 );
    // start BDD manager and create partitions 
    dd = Cudd_Init( nVars, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    vParts = Llb_Nonlin4SweepPartitions( dd, pAig, vOrder, 0 );
//printf( "Outputs: " );
//Llb_Nonlin4SweepPrint( vParts );
    // compute image of the partitions
    bMonitor = Llb4_Nonlin4SweepBadMonitor( pAig, vOrder, dd );  Cudd_Ref( bMonitor );
    Cudd_AutodynEnable( dd,  CUDD_REORDER_SYMM_SIFT );
    bImage = Llb_Nonlin4Image( dd, vParts, bMonitor, vVars2Q );  Cudd_Ref( bImage );
    Cudd_RecursiveDeref( dd, bMonitor );
    Llb_Nonlin4SweepDeref( dd, vParts );
    Vec_IntFree( vVars2Q );
    // save image and return
    dd.bFunc = bImage;
    return dd;
}

/**Function*************************************************************

  Synopsis    [Computes clusters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb4_Nonlin4SweepGroups(pAig *Aig_Man_t, vOrder *Vec_Int_t, nVars int32, pvGroups **Vec_Ptr_t, nBddLimitClp int32, fVerbose int32) *DdManager {
    var dd *DdManager;
    var vParts *Vec_Ptr_t;
    var vVars2Q *Vec_Int_t;
    // get quantifiable variables
    vVars2Q = Llb_Nonlin4SweepVars2Q( pAig, vOrder, 1 );
    // start BDD manager and create partitions 
    dd = Cudd_Init( nVars, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    vParts = Llb_Nonlin4SweepPartitions( dd, pAig, vOrder, 1 );
//printf( "Transitions: " );
//Llb_Nonlin4SweepPrint( vParts );
    // compute image of the partitions

    Cudd_AutodynEnable( dd,  CUDD_REORDER_SYMM_SIFT );
    *pvGroups = Llb_Nonlin4Group( dd, vParts, vVars2Q, nBddLimitClp );
    Llb_Nonlin4SweepDeref( dd, vParts );
//    *pvGroups = vParts;

if fVerbose != 0 {
fmt.Printf("Groups: ");
Llb_Nonlin4SweepPrint( *pvGroups );
}

    Vec_IntFree( vVars2Q );
    return dd;
}


/**Function*************************************************************

  Synopsis    [Creates quantifiable variables for both types of traversal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4SweepPrintSuppProfile(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t, vGroups *Vec_Ptr_t, fVerbose int32) {
    var pObj *Aig_Obj_t;
    var i int32;
    var pSupp *int32;
    var nSuppAll, nSuppPi, nSuppPo, nSuppLi, nSuppLo, nSuppAnd int32;

    pSupp = ABC_CALLOC( int, Cudd_ReadSize(dd) );
    Extra_VectorSupportArray( dd, (DdNode **)Vec_PtrArray(vGroups), Vec_PtrSize(vGroups), pSupp );

    Aig_ManForEachObj( pAig, pObj, i )
    {
        if Llb_ObjBddVar(vOrder, pObj) < 0 {
            continue; }
        // remove variables that do not participate
        if pSupp[Llb_ObjBddVar(vOrder, pObj)] == 0 {
            if Aig_ObjIsNode(pObj) {
                Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), -1 ); }
            continue;
        }
        nSuppAll++;
        if Saig_ObjIsPi(pAig, pObj) {
            nSuppPi++;
        } else if Saig_ObjIsLo(pAig, pObj) {
            nSuppLo++;
        } else if Saig_ObjIsPo(pAig, pObj) {
            nSuppPo++;
        } else if Saig_ObjIsLi(pAig, pObj) {
            nSuppLi++;
        } else {
            nSuppAnd++; }
    }
    ABC_FREE( pSupp );

    if fVerbose != 0 {
    fmt.Printf("Groups =%3d  ", Vec_PtrSize(vGroups));
    fmt.Printf("Variables: all =%4d ", nSuppAll);
    fmt.Printf("pi =%4d ", nSuppPi);
    fmt.Printf("po =%4d ", nSuppPo);
    fmt.Printf("lo =%4d ", nSuppLo);
    fmt.Printf("li =%4d ", nSuppLi);
    fmt.Printf("and =%4d", nSuppAnd);
    fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Performs BDD sweep on the netlist.]

  Description [Returns BDD manager, ordering, clusters, and bad states 
  inside dd->bFunc.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb4_Nonlin4Sweep(pAig *Aig_Man_t, nSweepMax int32, nClusterMax int32, pdd **DdManager, pvOrder **Vec_Int_t, pvGroups **Vec_Ptr_t, fVerbose int32) {
    var ddBad, ddWork *DdManager;
    var vGroups *Vec_Ptr_t;
    var vOrder *Vec_Int_t;
    var Counter, nCutPoints int32;

    // get the original ordering
    Aig_ManCleanMarkA( pAig );
    vOrder = Llb_Nonlin4SweepOrder( pAig, &Counter, 1 );
    assert(Counter == Aig_ManNodeNum(pAig), "Counter == Aig_ManNodeNum(pAig)");
    // mark the nodes
    nCutPoints = Llb4_Nonlin4SweepCutpoints( pAig, vOrder, nSweepMax, fVerbose );
    Vec_IntFree( vOrder );
    // get better ordering
    vOrder = Llb_Nonlin4SweepOrder( pAig, &Counter, 0 );
    assert(Counter == nCutPoints, "Counter == nCutPoints");
    Aig_ManCleanMarkA( pAig );
    // compute the BAD states
    ddBad = Llb4_Nonlin4SweepBadStates( pAig, vOrder, nCutPoints + Aig_ManCiNum(pAig) + Aig_ManCoNum(pAig) );
    // compute the clusters
    ddWork = Llb4_Nonlin4SweepGroups( pAig, vOrder, nCutPoints + Aig_ManCiNum(pAig) + Aig_ManCoNum(pAig), &vGroups, nClusterMax, fVerbose );
    // transfer the result from the Bad manager
//printf( "Bad before = %d.\n", Cudd_DagSize(ddBad->bFunc) );
    ddWork.bFunc = Cudd_bddTransfer( ddBad, ddWork, ddBad.bFunc );   Cudd_Ref( ddWork.bFunc );
    Cudd_RecursiveDeref( ddBad, ddBad.bFunc );  ddBad.bFunc = nil;
    Extra_StopManager( ddBad );
    // update ordering to exclude quantified variables
//printf( "Bad after = %d.\n", Cudd_DagSize(ddWork->bFunc) );

    Llb_Nonlin4SweepPrintSuppProfile( ddWork, pAig, vOrder, vGroups, fVerbose );

    // return the result
    *pdd = ddWork;
    *pvOrder = vOrder;
    *pvGroups = vGroups;
}

/**Function*************************************************************

  Synopsis    [Performs BDD sweep on the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb4_Nonlin4SweepExperiment(pAig *Aig_Man_t) {
    var dd *DdManager;
    var vOrder *Vec_Int_t;
    var vGroups *Vec_Ptr_t;
    Llb4_Nonlin4Sweep( pAig, 100, 500, &dd, &vOrder, &vGroups, 1 );

    Llb_Nonlin4SweepDeref( dd, vGroups );

    Cudd_RecursiveDeref( dd, dd.bFunc );
    Extra_StopManager( dd );
    Vec_IntFree( vOrder );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


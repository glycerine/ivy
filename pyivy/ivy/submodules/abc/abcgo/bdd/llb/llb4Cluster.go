package llb

import "fmt"

/**CFile****************************************************************

  FileName    [llb2Cluster.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Non-linear quantification scheduling.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb2Cluster.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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
func Llb_Nonlin4FindOrder_rec(pAig *Aig_Man_t, pObj *Aig_Obj_t, vOrder *Vec_Int_t, pCounter *int32) {
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
        Llb_Nonlin4FindOrder_rec( pAig, pFanin0, vOrder, pCounter );
        Llb_Nonlin4FindOrder_rec( pAig, pFanin1, vOrder, pCounter );
    
    } else {
        Llb_Nonlin4FindOrder_rec( pAig, pFanin1, vOrder, pCounter );
        Llb_Nonlin4FindOrder_rec( pAig, pFanin0, vOrder, pCounter );
    }
    if pObj.fMarkA {
        Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), (*pCounter)++ ); }
}

/**Function*************************************************************

  Synopsis    [Find good static variable ordering.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4FindOrder(pAig *Aig_Man_t, pCounter *int32) *Vec_Int_t {
    var vNodes *Vec_Int_t  = nil;
    var vOrder *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // mark nodes to exclude:  AND with low level and CO drivers
    Aig_ManCleanMarkA( pAig );
    Aig_ManForEachNode( pAig, pObj, i )
        if ( Aig_ObjLevel(pObj) > 3 )
            pObj.fMarkA = 1;
    Aig_ManForEachCo( pAig, pObj, i )
        Aig_ObjFanin0(pObj).fMarkA = 0;

    // collect nodes in the order
    vOrder = Vec_IntStartFull( Aig_ManObjNumMax(pAig) );
    Aig_ManIncrementTravId( pAig );
    Aig_ObjSetTravIdCurrent( pAig, Aig_ManConst1(pAig) );
//    Aig_ManForEachCo( pAig, pObj, i )
    Saig_ManForEachLi( pAig, pObj, i )
    {
        Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), Counter++ );
        Llb_Nonlin4FindOrder_rec( pAig, Aig_ObjFanin0(pObj), vOrder, &Counter );
    }
    Aig_ManForEachCi( pAig, pObj, i )
        if ( Llb_ObjBddVar(vOrder, pObj) < 0 )
            Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), Counter++ );
    Aig_ManCleanMarkA( pAig );
    Vec_IntFreeP( &vNodes );
//    assert( Counter == Aig_ManObjNum(pAig) - 1 );

/*
    Saig_ManForEachPi( pAig, pObj, i )
        printf( "pi%d ", Llb_ObjBddVar(vOrder, pObj) );
    printf( "\n" );
    Saig_ManForEachLo( pAig, pObj, i )
        printf( "lo%d ", Llb_ObjBddVar(vOrder, pObj) );
    printf( "\n" );
    Saig_ManForEachPo( pAig, pObj, i )
        printf( "po%d ", Llb_ObjBddVar(vOrder, pObj) );
    printf( "\n" );
    Saig_ManForEachLi( pAig, pObj, i )
        printf( "li%d ", Llb_ObjBddVar(vOrder, pObj) );
    printf( "\n" );
    Aig_ManForEachNode( pAig, pObj, i )
        printf( "n%d ", Llb_ObjBddVar(vOrder, pObj) );
    printf( "\n" );
*/
    if pCounter != nil {
        *pCounter = Counter; }
    return vOrder;
}

/**Function*************************************************************

  Synopsis    [Derives BDDs for the partitions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4FindPartitions_rec(dd *DdManager, pObj *Aig_Obj_t, vOrder *Vec_Int_t, vRoots *Vec_Ptr_t) *DdNode {
    var bBdd, bBdd0, bBdd1, bPart, vVar *DdNode;
    if Aig_ObjIsConst1(pObj) {
        return Cudd_ReadOne(dd); } 
    if Aig_ObjIsCi(pObj) {
        return Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) ); }
    if pObj.pData {
        return (DdNode *)pObj.pData; }
    if Aig_ObjIsCo(pObj) {
        // inlining of macro Cudd_NotCond
        bBdd0 = ((DdNode *)((ptrint)(Llb_Nonlin4FindPartitions_rec(dd, Aig_ObjFanin0(pObj), vOrder, vRoots)) ^ (Aig_ObjFaninC0(pObj))));
        bPart = Cudd_bddXnor( dd, Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) ), bBdd0 );  Cudd_Ref( bPart );
        Vec_PtrPush( vRoots, bPart );
        return nil;
    }
    // inlining of macro Cudd_NotCond
    bBdd0 = ((DdNode *)((ptrint)(Llb_Nonlin4FindPartitions_rec(dd, Aig_ObjFanin0(pObj), vOrder, vRoots)) ^ (Aig_ObjFaninC0(pObj))));
    // inlining of macro Cudd_NotCond
    bBdd1 = ((DdNode *)((ptrint)(Llb_Nonlin4FindPartitions_rec(dd, Aig_ObjFanin1(pObj), vOrder, vRoots)) ^ (Aig_ObjFaninC1(pObj))));
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
func Llb_Nonlin4FindPartitions(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t, fOutputs int32) *Vec_Ptr_t {
    var vRoots *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManCleanData( pAig );
    vRoots = Vec_PtrAlloc( 100 );
    if fOutputs != 0 {
        Saig_ManForEachPo( pAig, pObj, i )
            Llb_Nonlin4FindPartitions_rec( dd, pObj, vOrder, vRoots );
    
    } else {
        Saig_ManForEachLi( pAig, pObj, i )
            Llb_Nonlin4FindPartitions_rec( dd, pObj, vOrder, vRoots );
    }
    Aig_ManForEachNode( pAig, pObj, i )
        if ( pObj.pData )
            Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData );
    return vRoots;
}

/**Function*************************************************************

  Synopsis    [Creates quantifiable variables for both types of traversal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4FindVars2Q(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t) *Vec_Int_t {
    var vVars2Q *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vVars2Q = Vec_IntAlloc( 0 );
    Vec_IntFill( vVars2Q, Cudd_ReadSize(dd), 1 );
    Saig_ManForEachLo( pAig, pObj, i )
        Vec_IntWriteEntry( vVars2Q, Llb_ObjBddVar(vOrder, pObj), 0 );
//    Aig_ManForEachCo( pAig, pObj, i )
    Saig_ManForEachLi( pAig, pObj, i )
        Vec_IntWriteEntry( vVars2Q, Llb_ObjBddVar(vOrder, pObj), 0 );
    return vVars2Q;
}

/**Function*************************************************************

  Synopsis    [Creates quantifiable variables for both types of traversal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4CountTerms(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t, bFunc *DdNode, fCo int32, fFlop int32) int32 {
    var bSupp *DdNode;
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    bSupp = Cudd_Support( dd, bFunc );  Cudd_Ref( bSupp );
    if !(fCo != 0) && !(fFlop != 0) {
        Saig_ManForEachPi( pAig, pObj, i )
            if ( Llb_ObjBddVar(vOrder, pObj) >= 0 )
                Counter += Cudd_bddLeq( dd, bSupp, Cudd_bddIthVar(dd, Llb_ObjBddVar(vOrder, pObj)) );
    
    } else if (fCo != 0) && !(fFlop != 0) {
        Saig_ManForEachPo( pAig, pObj, i )
            if ( Llb_ObjBddVar(vOrder, pObj) >= 0 )
                Counter += Cudd_bddLeq( dd, bSupp, Cudd_bddIthVar(dd, Llb_ObjBddVar(vOrder, pObj)) );
    
    } else if !(fCo != 0) && (fFlop != 0) {
        Saig_ManForEachLo( pAig, pObj, i )
            if ( Llb_ObjBddVar(vOrder, pObj) >= 0 )
                Counter += Cudd_bddLeq( dd, bSupp, Cudd_bddIthVar(dd, Llb_ObjBddVar(vOrder, pObj)) );
    
    } else if (fCo != 0) && (fFlop != 0) {
        Saig_ManForEachLi( pAig, pObj, i )
            if ( Llb_ObjBddVar(vOrder, pObj) >= 0 )
                Counter += Cudd_bddLeq( dd, bSupp, Cudd_bddIthVar(dd, Llb_ObjBddVar(vOrder, pObj)) );
    }
    Cudd_RecursiveDeref( dd, bSupp );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Creates quantifiable variables for both types of traversal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4PrintGroups(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t, vGroups *Vec_Ptr_t) {
    var bTemp *DdNode;
    var i, nSuppAll, nSuppPi, nSuppPo, nSuppLi, nSuppLo, nSuppAnd int32;
    Vec_PtrForEachEntry( DdNode *, vGroups, bTemp, i )
    {
//Extra_bddPrintSupport(dd, bTemp);  printf("\n" );
        nSuppAll = Cudd_SupportSize(dd,bTemp);
        nSuppPi  = Llb_Nonlin4CountTerms(dd, pAig, vOrder, bTemp, 0, 0);
        nSuppPo  = Llb_Nonlin4CountTerms(dd, pAig, vOrder, bTemp, 1, 0);
        nSuppLi  = Llb_Nonlin4CountTerms(dd, pAig, vOrder, bTemp, 0, 1);
        nSuppLo  = Llb_Nonlin4CountTerms(dd, pAig, vOrder, bTemp, 1, 1);
        nSuppAnd = nSuppAll - (nSuppPi+nSuppPo+nSuppLi+nSuppLo);

        if Cudd_DagSize(bTemp) <= 10 {
            continue; }

        fmt.Printf("%4d : bdd =%6d  supp =%3d  ", i, Cudd_DagSize(bTemp), nSuppAll);
        fmt.Printf("pi =%3d ", nSuppPi);
        fmt.Printf("po =%3d ", nSuppPo);
        fmt.Printf("lo =%3d ", nSuppLo);
        fmt.Printf("li =%3d ", nSuppLi);
        fmt.Printf("and =%3d", nSuppAnd);
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Creates quantifiable variables for both types of traversal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4PrintSuppProfile(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t, vGroups *Vec_Ptr_t) {
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

    fmt.Printf("Groups =%3d  ", Vec_PtrSize(vGroups));
    fmt.Printf("Variables: all =%4d ", nSuppAll);
    fmt.Printf("pi =%4d ", nSuppPi);
    fmt.Printf("po =%4d ", nSuppPo);
    fmt.Printf("lo =%4d ", nSuppLo);
    fmt.Printf("li =%4d ", nSuppLi);
    fmt.Printf("and =%4d", nSuppAnd);
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4Cluster(pAig *Aig_Man_t, pdd **DdManager, pvOrder **Vec_Int_t, pvGroups **Vec_Ptr_t, nBddMax int32, fVerbose int32) {
    var dd *DdManager;
    var vOrder, vVars2Q *Vec_Int_t;
    var vParts, vGroups *Vec_Ptr_t;
    var bTemp *DdNode;
    var i, nVarNum int32;

    // create the BDD manager
    vOrder  = Llb_Nonlin4FindOrder( pAig, &nVarNum );
    dd      = Cudd_Init( nVarNum, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
//    Cudd_AutodynEnable( dd,  CUDD_REORDER_SYMM_SIFT );

    vVars2Q = Llb_Nonlin4FindVars2Q( dd, pAig, vOrder );
    vParts  = Llb_Nonlin4FindPartitions( dd, pAig, vOrder, 0 );

    vGroups = Llb_Nonlin4Group( dd, vParts, vVars2Q, nBddMax );
    Vec_IntFree( vVars2Q );

    Vec_PtrForEachEntry( DdNode *, vParts, bTemp, i )
        Cudd_RecursiveDeref( dd, bTemp );
    Vec_PtrFree( vParts );


//    if ( fVerbose )
    Llb_Nonlin4PrintSuppProfile( dd, pAig, vOrder, vGroups );
    if fVerbose != 0 {
    fmt.Printf("Before reordering\n"); }
    if fVerbose != 0 {
    Llb_Nonlin4PrintGroups( dd, pAig, vOrder, vGroups ); }

//    Cudd_ReduceHeap( dd, CUDD_REORDER_SYMM_SIFT, 1 );
//    printf( "After reordering\n" );
//    Llb_Nonlin4PrintGroups( dd, pAig, vOrder, vGroups );

    if pvOrder {
        *pvOrder = vOrder;
    } else {
        Vec_IntFree( vOrder ); }

    if pvGroups {
        *pvGroups = vGroups;
    } else {
        Vec_PtrForEachEntry( DdNode *, vGroups, bTemp, i )
            Cudd_RecursiveDeref( dd, bTemp );
        Vec_PtrFree( vGroups );
    }

    if pdd {
        *pdd = dd;
    } else {
        Extra_StopManager( dd ); }
//    Cudd_Quit( dd );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


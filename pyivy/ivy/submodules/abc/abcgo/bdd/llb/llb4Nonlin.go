package llb

import (
	"fmt"
	"math"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [llb2Nonlin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Non-linear quantification scheduling.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb2Nonlin.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"
//#include "base/abc/abc.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START
 

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Llb_Mnx_t struct {
    // user info
    pAig      *Aig_Man_t;           // AIG manager
    pPars   *Gia_ParLlb_t;          // parameters

    // intermediate BDDs
    dd      *DdManager;             // BDD manager
    bBad         *DdNode;           // bad states in terms of CIs
    bReached         *DdNode;       // reached states 
    bCurrent         *DdNode;       // from states
    bNext         *DdNode;          // to states
    vRings      *Vec_Ptr_t;         // onion rings in ddR
    vRoots      *Vec_Ptr_t;         // BDDs for partitions

    // structural info
    vOrder      *Vec_Int_t;         // for each object ID, its BDD variable number or -1
    vVars2Q      *Vec_Int_t;        // 1 if variable is quantifiable; 0 othervise

    timeImage         abctime;
    timeRemap         abctime;
    timeReo         abctime;
    timeOther         abctime;
    timeTotal         abctime;
}

//extern int timeBuild, timeAndEx, timeOther;
//extern int nSuppMax;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Computes bad in working manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4ComputeBad(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t) *DdNode {
    var vNodes *Vec_Ptr_t;
    var bBdd, bBdd0, bBdd1, bTemp, bResult, bCube *DdNode;
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManCleanData( pAig );
    // assign elementary variables
    Aig_ManConst1(pAig).pData = Cudd_ReadOne(dd); 
    Aig_ManForEachCi( pAig, pObj, i )
        pObj.pData = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) );
    // compute internal nodes
    vNodes = Aig_ManDfsNodes( pAig, (Aig_Obj_t **)Vec_PtrArray(pAig.vCos), Saig_ManPoNum(pAig) );
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
    {
        if !Aig_ObjIsNode(pObj) {
            continue; }
        // inlining of macro Cudd_NotCond
        bBdd0 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin0(pObj).pData) ^ (Aig_ObjFaninC0(pObj))));
        // inlining of macro Cudd_NotCond
        bBdd1 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin1(pObj).pData) ^ (Aig_ObjFaninC1(pObj))));
        bBdd  = Cudd_bddAnd( dd, bBdd0, bBdd1 );
        if bBdd == nil {
            Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
                if ( Aig_ObjIsNode(pObj) && pObj.pData != nil )
                    Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData );
            Vec_PtrFree( vNodes );
            return nil;
        }
        Cudd_Ref( bBdd );
        pObj.pData = bBdd;
    }
    // quantify PIs of each PO
    bResult = Cudd_ReadLogicZero( dd );  Cudd_Ref( bResult );
    Saig_ManForEachPo( pAig, pObj, i )
    {
        // inlining of macro Cudd_NotCond
        bBdd0   = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin0(pObj).pData) ^ (Aig_ObjFaninC0(pObj))));
        bResult = Cudd_bddOr( dd, bTemp = bResult, bBdd0 );     
        if bResult == nil {
            Cudd_RecursiveDeref( dd, bTemp );
            break;
        }
        Cudd_Ref( bResult );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    // deref
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
        if ( Aig_ObjIsNode(pObj) && pObj.pData != nil )
            Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData );
    Vec_PtrFree( vNodes );
    if bResult {
        bCube = Cudd_ReadOne(dd);  Cudd_Ref( bCube );
        Saig_ManForEachPi( pAig, pObj, i )
        {
            bCube = Cudd_bddAnd( dd, bTemp = bCube, (DdNode *)pObj.pData );  
            if bCube == nil {
                Cudd_RecursiveDeref( dd, bTemp );
                Cudd_RecursiveDeref( dd, bResult );
                bResult = nil;
                break;
            }
            Cudd_Ref( bCube );
            Cudd_RecursiveDeref( dd, bTemp );
        }
        if bResult != nil {
            bResult = Cudd_bddExistAbstract( dd, bTemp = bResult, bCube );  Cudd_Ref( bResult );
            Cudd_RecursiveDeref( dd, bTemp );
            Cudd_RecursiveDeref( dd, bCube );
            Cudd_Deref( bResult );
        }
    }
//if ( bResult )
//printf( "Bad state = %d.\n", Cudd_DagSize(bResult) );
    return bResult;
}

/**Function*************************************************************

  Synopsis    [Derives BDDs for the partitions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4DerivePartitions(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t) *Vec_Ptr_t {
    var vRoots *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var bBdd, bBdd0, bBdd1, bPart *DdNode;
    var i int32;
    Aig_ManCleanData( pAig );
    // assign elementary variables
    Aig_ManConst1(pAig).pData = Cudd_ReadOne(dd); 
    Aig_ManForEachCi( pAig, pObj, i )
        pObj.pData = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) );
    Aig_ManForEachNode( pAig, pObj, i )
        if Llb_ObjBddVar(vOrder, pObj) >= 0 {
            pObj.pData = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) );
            Cudd_Ref( (DdNode *)pObj.pData );
        }
    Saig_ManForEachLi( pAig, pObj, i )
        pObj.pData = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObj) );
    // compute intermediate BDDs
    vRoots = Vec_PtrAlloc( 100 );
    Aig_ManForEachNode( pAig, pObj, i )
    {
        // inlining of macro Cudd_NotCond
        bBdd0 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin0(pObj).pData) ^ (Aig_ObjFaninC0(pObj))));
        // inlining of macro Cudd_NotCond
        bBdd1 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin1(pObj).pData) ^ (Aig_ObjFaninC1(pObj))));
        bBdd  = Cudd_bddAnd( dd, bBdd0, bBdd1 );
        if bBdd == nil {
            goto finish; }
        Cudd_Ref( bBdd );
        if pObj.pData == nil {
            pObj.pData = bBdd;
            continue;
        }
        // create new partition
        bPart = Cudd_bddXnor( dd, (DdNode *)pObj.pData, bBdd );  
        if bPart == nil {
            goto finish; }
        Cudd_Ref( bPart );
        Cudd_RecursiveDeref( dd, bBdd );
        Vec_PtrPush( vRoots, bPart );
//printf( "%d ", Cudd_DagSize(bPart) );
    }
    // compute register output BDDs
    Saig_ManForEachLi( pAig, pObj, i )
    {
        // inlining of macro Cudd_NotCond
        bBdd0 = ((DdNode *)((ptrint)((DdNode *)Aig_ObjFanin0(pObj).pData) ^ (Aig_ObjFaninC0(pObj))));
        bPart = Cudd_bddXnor( dd, (DdNode *)pObj.pData, bBdd0 );  
        if bPart == nil {
            goto finish; }
        Cudd_Ref( bPart );
        Vec_PtrPush( vRoots, bPart );
//printf( "%d ", Cudd_DagSize(bPart) );
    }
//printf( "\n" );
    Aig_ManForEachNode( pAig, pObj, i )
        Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData );
    return vRoots;
    // early termination
finish:
    Aig_ManForEachNode( pAig, pObj, i )
        if ( pObj.pData )
            Cudd_RecursiveDeref( dd, (DdNode *)pObj.pData );
    Vec_PtrForEachEntry( DdNode *, vRoots, bPart, i )
        Cudd_RecursiveDeref( dd, bPart );
    Vec_PtrFree( vRoots );
    return nil;
}

/**Function*************************************************************

  Synopsis    [Find simple variable ordering.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4CreateOrderSimple(pAig *Aig_Man_t) *Vec_Int_t {
    var vOrder *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    vOrder = Vec_IntStartFull( Aig_ManObjNumMax(pAig) );
    Aig_ManForEachCi( pAig, pObj, i )
        Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), Counter++ );
    Saig_ManForEachLi( pAig, pObj, i )
        Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), Counter++ );
    return vOrder;
}

/**Function*************************************************************

  Synopsis    [Find good static variable ordering.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4CreateOrder_rec(pAig *Aig_Man_t, pObj *Aig_Obj_t, vOrder *Vec_Int_t, pCounter *int32) {
    var pFanin0, pFanin1 *Aig_Obj_t;
    if Aig_ObjIsTravIdCurrent(pAig, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent( pAig, pObj );
    assert(Llb_ObjBddVar(vOrder, pObj) < 0, "Llb_ObjBddVar(vOrder, pObj) < 0");
    if Aig_ObjIsCi(pObj) {
//        if ( Saig_ObjIsLo(pAig, pObj) )
//            Vec_IntWriteEntry( vOrder, Aig_ObjId(Saig_ObjLoToLi(pAig, pObj)), (*pCounter)++ );
        Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), (*pCounter)++ );
        return;
    }
    // try fanins with higher level first
    pFanin0 = Aig_ObjFanin0(pObj);
    pFanin1 = Aig_ObjFanin1(pObj);
//    if ( pFanin0->Level > pFanin1->Level || (pFanin0->Level == pFanin1->Level && pFanin0->Id < pFanin1->Id) )
    if pFanin0.Level > pFanin1.Level {
        Llb_Nonlin4CreateOrder_rec( pAig, pFanin0, vOrder, pCounter );
        Llb_Nonlin4CreateOrder_rec( pAig, pFanin1, vOrder, pCounter );
    
    } else {
        Llb_Nonlin4CreateOrder_rec( pAig, pFanin1, vOrder, pCounter );
        Llb_Nonlin4CreateOrder_rec( pAig, pFanin0, vOrder, pCounter );
    }
    if pObj.fMarkA {
        Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), (*pCounter)++ ); }
}

/**Function*************************************************************

  Synopsis    [Collect nodes with the given fanout count.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4CollectHighRefNodes(pAig *Aig_Man_t, nFans int32) *Vec_Int_t {
    var vNodes *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManCleanMarkA( pAig );
    Aig_ManForEachNode( pAig, pObj, i )
        if ( Aig_ObjRefs(pObj) >= nFans )
            pObj.fMarkA = 1;
    // unmark flop drivers
    Saig_ManForEachLi( pAig, pObj, i )
        Aig_ObjFanin0(pObj).fMarkA = 0;
    // collect mapping
    vNodes = Vec_IntAlloc( 100 );
    Aig_ManForEachNode( pAig, pObj, i )
        if ( pObj.fMarkA )
            Vec_IntPush( vNodes, Aig_ObjId(pObj) );
    Aig_ManCleanMarkA( pAig );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Find good static variable ordering.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4CreateOrder(pAig *Aig_Man_t) *Vec_Int_t {
    var vNodes *Vec_Int_t  = nil;
    var vOrder *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i, Counter int32;
/*
    // mark internal nodes to be used
    Aig_ManCleanMarkA( pAig );
    vNodes = Llb_Nonlin4CollectHighRefNodes( pAig, 4 );
    Aig_ManForEachObjVec( vNodes, pAig, pObj, i )
        pObj->fMarkA = 1;
printf( "Techmapping added %d pivots.\n", Vec_IntSize(vNodes) );
*/
    // collect nodes in the order
    vOrder = Vec_IntStartFull( Aig_ManObjNumMax(pAig) );
    Aig_ManIncrementTravId( pAig );
    Aig_ObjSetTravIdCurrent( pAig, Aig_ManConst1(pAig) );
    Saig_ManForEachLi( pAig, pObj, i )
    {
        Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), Counter++ );
        Llb_Nonlin4CreateOrder_rec( pAig, Aig_ObjFanin0(pObj), vOrder, &Counter );
    }
    Aig_ManForEachCi( pAig, pObj, i )
        if Llb_ObjBddVar(vOrder, pObj) < 0 {
//            if ( Saig_ObjIsLo(pAig, pObj) )
//                Vec_IntWriteEntry( vOrder, Aig_ObjId(Saig_ObjLoToLi(pAig, pObj)), Counter++ );
            Vec_IntWriteEntry( vOrder, Aig_ObjId(pObj), Counter++ );
        }
    assert(Counter <= Aig_ManCiNum(pAig) + Aig_ManRegNum(pAig) + (vNodes?Vec_IntSize(vNodes):0), "Counter <= Aig_ManCiNum(pAig) + Aig_ManRegNum(pAig) + (vNodes?Vec_IntSize(vNodes):0)");
    Aig_ManCleanMarkA( pAig );
    Vec_IntFreeP( &vNodes );
    return vOrder;
}


/**Function*************************************************************

  Synopsis    [Creates quantifiable varaibles for both types of traversal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4CreateVars2Q(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t, fBackward int32) *Vec_Int_t {
    var vVars2Q *Vec_Int_t;
    var pObjLi, pObjLo *Aig_Obj_t;
    var i int32;
    vVars2Q = Vec_IntAlloc( 0 );
    Vec_IntFill( vVars2Q, Cudd_ReadSize(dd), 1 );
    Saig_ManForEachLiLo( pAig, pObjLi, pObjLo, i )
        Vec_IntWriteEntry( vVars2Q, Llb_ObjBddVar(vOrder, fBackward ? pObjLo : pObjLi), 0 );
    return vVars2Q;
}

/**Function*************************************************************

  Synopsis    [Compute initial state in terms of current state variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4SetupVarMap(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t) {
    var pVarsX, pVarsY **DdNode;
    var pObjLo, pObjLi *Aig_Obj_t;
    var i int32;
    pVarsX = ABC_ALLOC( DdNode *, Cudd_ReadSize(dd) );
    pVarsY = ABC_ALLOC( DdNode *, Cudd_ReadSize(dd) );
    Saig_ManForEachLiLo( pAig, pObjLo, pObjLi, i )
    {
        assert(Llb_ObjBddVar(vOrder, pObjLo) >= 0, "Llb_ObjBddVar(vOrder, pObjLo) >= 0");
        assert(Llb_ObjBddVar(vOrder, pObjLi) >= 0, "Llb_ObjBddVar(vOrder, pObjLi) >= 0");
        pVarsX[i] = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObjLo) );
        pVarsY[i] = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObjLi) );
    }
    Cudd_SetVarMap( dd, pVarsX, pVarsY, Aig_ManRegNum(pAig) );
    ABC_FREE( pVarsX );
    ABC_FREE( pVarsY );
}

/**Function*************************************************************

  Synopsis    [Compute initial state in terms of current state variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4ComputeInitState(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t, fBackward int32) *DdNode {
    var pObjLi, pObjLo *Aig_Obj_t;
    var bRes, bVar, bTemp *DdNode;
    var i int32;
    var TimeStop abctime;
    TimeStop = dd.TimeStop;  dd.TimeStop = 0;
    bRes = Cudd_ReadOne( dd );   Cudd_Ref( bRes );
    Saig_ManForEachLiLo( pAig, pObjLi, pObjLo, i )
    {
        bVar = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, fBackward? pObjLi : pObjLo) );
        // inlining of macro Cudd_Not
        bRes = Cudd_bddAnd( dd, bTemp = bRes, ((DdNode *)((ptrint)(bVar) ^ 01)) );  Cudd_Ref( bRes );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    Cudd_Deref( bRes );
    dd.TimeStop = TimeStop;
    return bRes;
}

/**Function*************************************************************

  Synopsis    [Compute initial state in terms of current state variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4ComputeCube(dd *DdManager, pAig *Aig_Man_t, vOrder *Vec_Int_t, pValues string, Flag int32) *DdNode {
    var pObjLo, pObjLi, pObjTemp *Aig_Obj_t;
    var bRes, bVar, bTemp *DdNode;
    var i int32;
    var TimeStop abctime;
    TimeStop = dd.TimeStop;  dd.TimeStop = 0;
    bRes = Cudd_ReadOne( dd );   Cudd_Ref( bRes );
    Saig_ManForEachLiLo( pAig, pObjLi, pObjLo, i )
    {
        if Flag != 0 {
            pObjTemp = pObjLo;
pObjLo = pObjLi;
pObjLi = pObjTemp; }
        // get the correspoding flop input variable
        bVar = Cudd_bddIthVar( dd, Llb_ObjBddVar(vOrder, pObjLi) );
        if pValues[Llb_ObjBddVar(vOrder, pObjLo)] != 1 {
            // inlining of macro Cudd_Not
            bVar = ((DdNode *)((ptrint)(bVar) ^ 01)); }
        // create cube
        bRes = Cudd_bddAnd( dd, bTemp = bRes, bVar );  Cudd_Ref( bRes );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    Cudd_Deref( bRes );
    dd.TimeStop = TimeStop;
    return bRes;
}

/**Function*************************************************************

  Synopsis    [Compute initial state in terms of current state variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4RecordState(pAig *Aig_Man_t, vOrder *Vec_Int_t, pState *uint32, pValues string, fBackward int32) {
    var pObjLo, pObjLi *Aig_Obj_t;
    var i int32;
    Saig_ManForEachLiLo( pAig, pObjLi, pObjLo, i )
        if ( pValues[Llb_ObjBddVar(vOrder, fBackward? pObjLi : pObjLo)] == 1 )
            Abc_InfoSetBit( pState, i );
}

/**Function*************************************************************

  Synopsis    [Multiply every partition by the cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4Multiply(dd *DdManager, bCube *DdNode, vParts *Vec_Ptr_t) *Vec_Ptr_t {
    var vNew *Vec_Ptr_t;
    var bTemp, bFunc *DdNode;
    var i int32;
    vNew = Vec_PtrAlloc( Vec_PtrSize(vParts) );
    Vec_PtrForEachEntry( DdNode *, vParts, bFunc, i )
    {
        bTemp = Cudd_bddAnd( dd, bFunc, bCube );  Cudd_Ref( bTemp );
        Vec_PtrPush( vNew, bTemp );
    }
    return vNew;
}

/**Function*************************************************************

  Synopsis    [Multiply every partition by the cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4Deref(dd *DdManager, vParts *Vec_Ptr_t) {
    var bFunc *DdNode;
    var i int32;
    Vec_PtrForEachEntry( DdNode *, vParts, bFunc, i )
        Cudd_RecursiveDeref( dd, bFunc );
    Vec_PtrFree( vParts );
}

/**Function*************************************************************

  Synopsis    [Derives counter-example by backward reachability.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4DeriveCex(p *Llb_Mnx_t, fBackward int32, fVerbose int32) *Vec_Ptr_t {
    var vVars2Q *Vec_Int_t;
    var vStates, vRootsNew *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var bState *DdNode = nil;
    var bImage *DdNode;
    var bOneCube *DdNode;
    var bRing *DdNode;
    var i, v, RetValue int32;//, clk = Abc_Clock();
    var pValues string;
    assert(Vec_PtrSize(p.vRings) > 0, "Vec_PtrSize(p.vRings) > 0");
    // disable the timeout
    p.dd.TimeStop  = 0;

    // start the state set
    vStates = Vec_PtrAllocSimInfo( Vec_PtrSize(p.vRings), Abc_BitWordNum(Aig_ManRegNum(p.pAig)) );
    Vec_PtrCleanSimInfo( vStates, 0, Abc_BitWordNum(Aig_ManRegNum(p.pAig)) );
    if fBackward != 0 {
        Vec_PtrReverseOrder( vStates ); }

    // get the last cube
    pValues = ABC_ALLOC( char, Cudd_ReadSize(p.dd) );
    bOneCube = Cudd_bddIntersect( p.dd, (DdNode *)Vec_PtrEntryLast(p.vRings), p.bBad );  Cudd_Ref( bOneCube );
    RetValue = Cudd_bddPickOneCube( p.dd, bOneCube, pValues );
    Cudd_RecursiveDeref( p.dd, bOneCube );
    assert((RetValue != 0), "RetValue");

    // record the cube
    Llb_Nonlin4RecordState( p.pAig, p.vOrder, (unsigned *)Vec_PtrEntryLast(vStates), pValues, fBackward );

    // write state in terms of NS variables
    if Vec_PtrSize(p.vRings) > 1 {
        bState = Llb_Nonlin4ComputeCube( p.dd, p.pAig, p.vOrder, pValues, fBackward );   Cudd_Ref( bState );
    }
    // perform backward analysis
    vVars2Q = Llb_Nonlin4CreateVars2Q( p.dd, p.pAig, p.vOrder, !fBackward );
    Vec_PtrForEachEntryReverse( DdNode *, p.vRings, bRing, v )
    { 
        if v == Vec_PtrSize(p.vRings) - 1 {
            continue; }

        // preprocess partitions
        vRootsNew = Llb_Nonlin4Multiply( p.dd, bState, p.vRoots );
        Cudd_RecursiveDeref( p.dd, bState );

        // compute the next states
        bImage = Llb_Nonlin4Image( p.dd, vRootsNew, nil, vVars2Q ); Cudd_Ref( bImage );
        Llb_Nonlin4Deref( p.dd, vRootsNew );

        // intersect with the previous set
        bOneCube = Cudd_bddIntersect( p.dd, bImage, bRing );  Cudd_Ref( bOneCube );
        Cudd_RecursiveDeref( p.dd, bImage );

        // find any assignment of the BDD
        RetValue = Cudd_bddPickOneCube( p.dd, bOneCube, pValues );
        Cudd_RecursiveDeref( p.dd, bOneCube );
        assert((RetValue != 0), "RetValue");

        // record the cube
        Llb_Nonlin4RecordState( p.pAig, p.vOrder, (unsigned *)Vec_PtrEntry(vStates, v), pValues, fBackward );

        // check that we get the init state
        if v == 0 {
            Saig_ManForEachLo( p.pAig, pObj, i )
                assert((fBackward != 0) || pValues[Llb_ObjBddVar(p.vOrder, pObj)] == 0, "fBackward || pValues[Llb_ObjBddVar(p.vOrder, pObj)] == 0");
            break;
        } 

        // write state in terms of NS variables
        bState = Llb_Nonlin4ComputeCube( p.dd, p.pAig, p.vOrder, pValues, fBackward );   Cudd_Ref( bState );
    }
    Vec_IntFree( vVars2Q );
    ABC_FREE( pValues );
    if fBackward != 0 {
        Vec_PtrReverseOrder( vStates ); }
//    if ( fVerbose )
//        Abc_PrintTime( 1, "BDD-based cex generation time", Abc_Clock() - clk );
    return vStates;
}


/**Function*************************************************************

  Synopsis    [Perform reachability with hints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4Reachability(p *Llb_Mnx_t) int32 { 
    var bAux *DdNode;
    var nIters, nBddSizeFr, nBddSizeTo, nBddSizeTo2 int32;
    var clkTemp abctime;
    var clkIter abctime;
    var clk abctime = Abc_Clock();
    assert(Aig_ManRegNum(p.pAig) > 0, "Aig_ManRegNum(p.pAig) > 0");

    if p.pPars.fBackward {
        // create bad state in the ring manager
        if !p.pPars.fSkipOutCheck {
            p.bBad = Llb_Nonlin4ComputeInitState( p.dd, p.pAig, p.vOrder, p.pPars.fBackward );  Cudd_Ref( p.bBad );
        }
        // create init state
        if p.pPars.fCluster {
            p.bCurrent = p.dd.bFunc, p.dd.bFunc = nil; 
        } else {
            p.bCurrent = Llb_Nonlin4ComputeBad( p.dd, p.pAig, p.vOrder );          
            if p.bCurrent == nil {
                if !p.pPars.fSilent {
                    fmt.Printf("Reached timeout (%d seconds) during constructing the bad states.\n", p.pPars.TimeLimit); }
                p.pPars.iFrame = -1;
                return -1;
            }
            Cudd_Ref( p.bCurrent );
        }
        // remap into the next states
        p.bCurrent = Cudd_bddVarMap( p.dd, bAux = p.bCurrent );
        if p.bCurrent == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during remapping bad states.\n", p.pPars.TimeLimit); }
            Cudd_RecursiveDeref( p.dd, bAux );
            p.pPars.iFrame = -1;
            return -1;
        }
        Cudd_Ref( p.bCurrent );
        Cudd_RecursiveDeref( p.dd, bAux );
    
    } else {
        // create bad state in the ring manager
        if !p.pPars.fSkipOutCheck {
            if p.pPars.fCluster {
                p.bBad = p.dd.bFunc, p.dd.bFunc = nil; 
            } else {
                p.bBad = Llb_Nonlin4ComputeBad( p.dd, p.pAig, p.vOrder );          
                if p.bBad == nil {
                    if !p.pPars.fSilent {
                        fmt.Printf("Reached timeout (%d seconds) during constructing the bad states.\n", p.pPars.TimeLimit); }
                    p.pPars.iFrame = -1;
                    return -1;
                }
                Cudd_Ref( p.bBad );
            }
        
        } else if p.dd.bFunc {
            Cudd_RecursiveDeref( p.dd, p.dd.bFunc );
p.dd.bFunc = nil; }
        // compute the starting set of states
        p.bCurrent = Llb_Nonlin4ComputeInitState( p.dd, p.pAig, p.vOrder, p.pPars.fBackward );  Cudd_Ref( p.bCurrent );
    }
    // perform iterations
    p.bReached = p.bCurrent; Cudd_Ref( p.bReached );
    for  nIters = 0; nIters < p.pPars.nIterMax; nIters++  { 
        clkIter = Abc_Clock();
        // check the runtime limit
        if p.pPars.TimeLimit && Abc_Clock() > p.pPars.TimeTarget {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            return -1;
        }

        // save the onion ring
        Vec_PtrPush( p.vRings, p.bCurrent );   Cudd_Ref( p.bCurrent );

        // check it for bad states
        // inlining of macro Cudd_Not
        if !p.pPars.fSkipOutCheck && !Cudd_bddLeq( p.dd, p.bCurrent, ((DdNode *)((ptrint)(p.bBad) ^ 01)) ) {
            var vStates *Vec_Ptr_t;
            assert(p.pAig.pSeqModel == nil, "p.pAig.pSeqModel == nil");
            vStates = Llb_Nonlin4DeriveCex( p, p.pPars.fBackward, p.pPars.fVerbose ); 
            p.pAig.pSeqModel = Llb4_Nonlin4TransformCex( p.pAig, vStates, -1, p.pPars.fVerbose );
            Vec_PtrFreeP( &vStates );
            if !p.pPars.fSilent {
                Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.  ", p.pAig.pSeqModel.iPo, p.pAig.pName, nIters );
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            p.pPars.iFrame = nIters - 1;
            return 0;
        }

        // compute the next states
        clkTemp = Abc_Clock();
        p.bNext = Llb_Nonlin4Image( p.dd, p.vRoots, p.bCurrent, p.vVars2Q );
        if p.bNext == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation in quantification.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            return -1;
        }
        Cudd_Ref( p.bNext );
        p.timeImage += Abc_Clock() - clkTemp;

        // remap into current states
        clkTemp = Abc_Clock();
        p.bNext = Cudd_bddVarMap( p.dd, bAux = p.bNext );
        if p.bNext == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during remapping next states.\n", p.pPars.TimeLimit); }
            Cudd_RecursiveDeref( p.dd, bAux );
            p.pPars.iFrame = nIters - 1;
            return -1;
        }
        Cudd_Ref( p.bNext );
        Cudd_RecursiveDeref( p.dd, bAux );
        p.timeRemap += Abc_Clock() - clkTemp;

        // collect statistics
        if p.pPars.fVerbose {
            nBddSizeFr  = Cudd_DagSize( p.bCurrent );
            nBddSizeTo  = Cudd_DagSize( bAux );
            nBddSizeTo2 = Cudd_DagSize( p.bNext );
        }
        Cudd_RecursiveDeref( p.dd, p.bCurrent ); p.bCurrent = nil;

        // derive new states
        // inlining of macro Cudd_Not
        p.bCurrent = Cudd_bddAnd( p.dd, p.bNext, ((DdNode *)((ptrint)(p.bReached) ^ 01)) );     
        if p.bCurrent == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation in transfer 1.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            return -1;
        }
        Cudd_Ref( p.bCurrent );
        Cudd_RecursiveDeref( p.dd, p.bNext ); p.bNext = nil;
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        if  (((DdNode *)((ptruint)(p.bCurrent) & ~01))).index == CUDD_CONST_INDEX {
            break; }
/*
        // reduce BDD size using constrain // Cudd_bddRestrict
        p->bCurrent = Cudd_bddRestrict( p->dd, bAux = p->bCurrent, Cudd_Not(p->bReached) );   
        Cudd_Ref( p->bCurrent );
printf( "Before = %d.  After = %d.\n", Cudd_DagSize(bAux), Cudd_DagSize(p->bCurrent) );
        Cudd_RecursiveDeref( p->dd, bAux );
*/

        // add to the reached set
        p.bReached = Cudd_bddOr( p.dd, bAux = p.bReached, p.bCurrent );                 
        if p.bReached == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation in transfer 1.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Cudd_RecursiveDeref( p.dd, bAux );  
            return -1;
        }
        Cudd_Ref( p.bReached );
        Cudd_RecursiveDeref( p.dd, bAux );  


        // report the results
        if p.pPars.fVerbose {
            fmt.Printf("I =%5d : ", nIters);
            fmt.Printf("Fr =%7d  ", nBddSizeFr);
            fmt.Printf("ImNs =%7d  ", nBddSizeTo);
            fmt.Printf("ImCs =%7d  ", nBddSizeTo2);
            fmt.Printf("Rea =%7d   ", Cudd_DagSize(p.bReached));
            fmt.Printf("(%4d %4d)  ", Cudd_ReadReorderings(p.dd), Cudd_ReadGarbageCollections(p.dd));
            Abc_PrintTime( 1, "T", Abc_Clock() - clkIter );
        }
/*
        if ( pPars->fVerbose )
        {
            double nMints = Cudd_CountMinterm(p->dd, bReached, Saig_ManRegNum(p->pAig) );
//            Extra_bddPrint( p->dd, bReached );printf( "\n" );
            printf( "Reachable states = %.0f. (Ratio = %.4f %%)\n", nMints, 100.0*nMints/pow(2.0, Saig_ManRegNum(p->pAig)) );
            fflush( stdout ); 
        }
*/
        if nIters == p.pPars.nIterMax - 1 {
            if !p.pPars.fSilent {
                fmt.Printf("Reached limit on the number of timeframes (%d).\n", p.pPars.nIterMax); }
            p.pPars.iFrame = nIters;
            return -1;
        }
    }
    
    // report the stats
    if p.pPars.fVerbose {
        var nMints float64  = Cudd_CountMinterm(p.dd, p.bReached, Saig_ManRegNum(p.pAig) );
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        if p.bCurrent && ((((DdNode *)((ptruint)(p.bCurrent) & ~01))).index == CUDD_CONST_INDEX) {
            fmt.Printf("Reachability analysis completed after %d frames.\n", nIters);
        } else {
            fmt.Printf("Reachability analysis is stopped after %d frames.\n", nIters); }
        fmt.Printf("Reachable states = %.0f. (Ratio = %.4f %%)\n", nMints, 100.0*nMints/math.Pow(float64(2.0), float64(Saig_ManRegNum(p.pAig))));
        fflush( os.Stdout ); 
    }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if p.bCurrent == nil || !((((DdNode *)((ptruint)(p.bCurrent) & ~01))).index == CUDD_CONST_INDEX) {
        if !p.pPars.fSilent {
            fmt.Printf("Verified only for states reachable in %d frames.  ", nIters); }
        p.pPars.iFrame = p.pPars.nIterMax;
        return -1; // undecided
    }
    // report
    if !p.pPars.fSilent {
        fmt.Printf("The miter is proved unreachable after %d iterations.  ", nIters); }
    if !p.pPars.fSilent {
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk ); }
    p.pPars.iFrame = nIters - 1;
    return 1; // unreachable
}

/**Function*************************************************************

  Synopsis    [Reorders BDDs in the working manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4Reorder(dd *DdManager, fTwice int32, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    if fVerbose != 0 {
        Abc_Print( 1, "Reordering... Before =%5d. ", Cudd_ReadKeys(dd) - Cudd_ReadDead(dd) ); }
    Cudd_ReduceHeap( dd, CUDD_REORDER_SYMM_SIFT, 100 );
    if fVerbose != 0 {
        Abc_Print( 1, "After =%5d. ", Cudd_ReadKeys(dd) - Cudd_ReadDead(dd) ); }
    if fTwice != 0 {
        Cudd_ReduceHeap( dd, CUDD_REORDER_SYMM_SIFT, 100 );
        if fVerbose != 0 {
            Abc_Print( 1, "After =%5d. ", Cudd_ReadKeys(dd) - Cudd_ReadDead(dd) ); }
    }
    if fVerbose != 0 {
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk ); }
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_MnxStart(pAig *Aig_Man_t, pPars *Gia_ParLlb_t) *Llb_Mnx_t {
    var p *Llb_Mnx_t;

    p = new(Llb_Mnx_t);
    p.pAig    = pAig;
    p.pPars   = pPars;

    // compute time to stop
    if p.pPars.TimeLimit {
        p.pPars.TimeTarget = p.pPars.TimeLimit * int64(time.Second) + Abc_Clock();
    } else {
        p.pPars.TimeTarget = 0;
    }

    if pPars.fCluster {
//        Llb_Nonlin4Cluster( p->pAig, &p->dd, &p->vOrder, &p->vRoots, pPars->nBddMax, pPars->fVerbose );
//        Cudd_AutodynEnable( p->dd,  CUDD_REORDER_SYMM_SIFT );
        Llb4_Nonlin4Sweep( p.pAig, pPars.nBddMax, pPars.nClusterMax, &p.dd, &p.vOrder, &p.vRoots, pPars.fVerbose );
        // set the stop time parameter
        p.dd.TimeStop  = p.pPars.TimeTarget;
    
    } else {
//    p->vOrder  = Llb_Nonlin4CreateOrderSimple( pAig );
        p.vOrder  = Llb_Nonlin4CreateOrder( pAig );
        p.dd      = Cudd_Init( Vec_IntCountPositive(p.vOrder) + 1, 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
        Cudd_AutodynEnable( p.dd,  CUDD_REORDER_SYMM_SIFT );
        Cudd_SetMaxGrowth( p.dd, 1.05 );
        // set the stop time parameter
        p.dd.TimeStop  = p.pPars.TimeTarget;
        p.vRoots  = Llb_Nonlin4DerivePartitions( p.dd, pAig, p.vOrder );
    }

    Llb_Nonlin4SetupVarMap( p.dd, pAig, p.vOrder );
    p.vVars2Q = Llb_Nonlin4CreateVars2Q( p.dd, pAig, p.vOrder, p.pPars.fBackward );
    p.vRings  = Vec_PtrAlloc( 100 );

    if pPars.fReorder {
        Llb_Nonlin4Reorder( p.dd, 0, 1 ); }
    return p;
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_MnxStop(p *Llb_Mnx_t) {
    var bTemp *DdNode;
    var i int32;
    if p.pPars.fVerbose {
        p.timeReo = Cudd_ReadReorderingTime(p.dd);
        p.timeOther = p.timeTotal - p.timeImage - p.timeRemap;
        ABC_PRTP( "Image    ", p.timeImage, p.timeTotal );
        ABC_PRTP( "Remap    ", p.timeRemap, p.timeTotal );
        ABC_PRTP( "Other    ", p.timeOther, p.timeTotal );
        ABC_PRTP( "TOTAL    ", p.timeTotal, p.timeTotal );
        ABC_PRTP( "  reo    ", p.timeReo,   p.timeTotal );
    }
    // remove BDDs
    if p.bBad {
        Cudd_RecursiveDeref( p.dd, p.bBad ); }
    if p.bReached {
        Cudd_RecursiveDeref( p.dd, p.bReached ); }
    if p.bCurrent {
        Cudd_RecursiveDeref( p.dd, p.bCurrent ); }
    if p.bNext {
        Cudd_RecursiveDeref( p.dd, p.bNext ); }
    if p.vRings {
    Vec_PtrForEachEntry( DdNode *, p.vRings, bTemp, i )
        Cudd_RecursiveDeref( p.dd, bTemp ); }
    if p.vRoots {
    Vec_PtrForEachEntry( DdNode *, p.vRoots, bTemp, i )
        Cudd_RecursiveDeref( p.dd, bTemp ); }
    // remove arrays
    Vec_PtrFreeP( &p.vRings );
    Vec_PtrFreeP( &p.vRoots );
//Cudd_PrintInfo( p->dd, stdout );
    Extra_StopManager( p.dd );
    Vec_IntFreeP( &p.vOrder );
    Vec_IntFreeP( &p.vVars2Q );
    ABC_FREE( p );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_MnxCheckNextStateVars(p *Llb_Mnx_t) {
    var pObj *Aig_Obj_t;
    var i, Counter0, Counter1 int32;
    Saig_ManForEachLi( p.pAig, pObj, i )
        if Saig_ObjIsLo(p.pAig, Aig_ObjFanin0(pObj)) {
            if Aig_ObjFaninC0(pObj) {
                Counter0++;
            } else {
                Counter1++; }
        }
    fmt.Printf("Total = %d.  Direct LO = %d. Compl LO = %d.\n", Aig_ManRegNum(p.pAig), Counter1, Counter0);
}

/**Function*************************************************************

  Synopsis    [Finds balanced cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_Nonlin4CoreReach(pAig *Aig_Man_t, pPars *Gia_ParLlb_t) int32 {
    var pMnn *Llb_Mnx_t;
    var RetValue int32  = -1;
    if pPars.fVerbose {
    Aig_ManPrintStats( pAig ); }
    if pPars.fCluster && Aig_ManObjNum(pAig) >= (1 << 15) {
        fmt.Printf("The number of objects is more than 2^15.  Clustering cannot be used.\n");
        return RetValue;
    }
    {
        var clk abctime  = Abc_Clock();
        pMnn = Llb_MnxStart( pAig, pPars );
//Llb_MnxCheckNextStateVars( pMnn );
        if !pPars.fSkipReach {
            RetValue = Llb_Nonlin4Reachability( pMnn ); }
        pMnn.timeTotal = Abc_Clock() - clk;
        Llb_MnxStop( pMnn );
    }
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Takes an AIG and returns an AIG representing reachable states.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ReachableStates(pAig *Aig_Man_t) *Aig_Man_t {
    extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
    var vPermute *Vec_Int_t;
    var vNames *Vec_Ptr_t;
    var Pars Gia_ParLlb_t;
    var pPars *Gia_ParLlb_t = &Pars;
    var dd *DdManager;
    var bReached *DdNode;
    var pMnn *Llb_Mnx_t;
    var pNtk, pNtkMuxes *Abc_Ntk_t;
    var pObj *Aig_Obj_t;
    var i, RetValue int32;
    var clk abctime  = Abc_Clock();

    // create parameters
    Llb_ManSetDefaultParams( pPars );
    pPars.fSkipOutCheck = 1;
    pPars.fCluster      = 0;
    pPars.fReorder      = 0;
    pPars.fSilent       = 1;
    pPars.nBddMax       = 100;
    pPars.nClusterMax   = 500;

    // run reachability
    pMnn = Llb_MnxStart( pAig, pPars );
    RetValue = Llb_Nonlin4Reachability( pMnn );
    assert(RetValue == 1, "RetValue == 1");

    // print BDD
//    Extra_bddPrint( pMnn->dd, pMnn->bReached ); 
//    Extra_bddPrintSupport( pMnn->dd, pMnn->bReached ); 
//    printf( "\n" );

    // collect flop output variables
    vPermute = Vec_IntStartFull( Cudd_ReadSize(pMnn.dd) );
    Saig_ManForEachLo( pAig, pObj, i )
        Vec_IntWriteEntry( vPermute, Llb_ObjBddVar(pMnn.vOrder, pObj), i );

    // transfer the reached state BDD into the new manager
    dd = Cudd_Init( Saig_ManRegNum(pAig), 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    Cudd_AutodynEnable( dd,  CUDD_REORDER_SYMM_SIFT );
    bReached = Extra_TransferPermute( pMnn.dd, dd, pMnn.bReached, Vec_IntArray(vPermute) );  Cudd_Ref( bReached );
    Vec_IntFree( vPermute );
    assert(Cudd_ReadSize(dd) == Saig_ManRegNum(pAig), "Cudd_ReadSize(dd) == Saig_ManRegNum(pAig)");

    // quit reachability engine
    pMnn.timeTotal = Abc_Clock() - clk;
    Llb_MnxStop( pMnn );

    // derive the network
    vNames = Abc_NodeGetFakeNames( Saig_ManRegNum(pAig) );
    pNtk = Abc_NtkDeriveFromBdd( dd, bReached, "reached", vNames );
    Abc_NodeFreeNames( vNames );
    Cudd_RecursiveDeref( dd, bReached );
    Cudd_Quit( dd );

    // convert
    pNtkMuxes = Abc_NtkBddToMuxes( pNtk, 0, 1000000 );
    Abc_NtkDelete( pNtk );
    pNtk = Abc_NtkStrash( pNtkMuxes, 0, 1, 0 );
    Abc_NtkDelete( pNtkMuxes );
    pAig = Abc_NtkToDar( pNtk, 0, 0 );
    Abc_NtkDelete( pNtk );
    return pAig;
}
func Llb_ReachableStatesGia(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pAig, pReached *Aig_Man_t;
    pAig = Gia_ManToAigSimple( p );
    pReached = Llb_ReachableStates( pAig );
    Aig_ManStop( pAig );
    pNew = Gia_ManFromAigSimple( pReached );
    Aig_ManStop( pReached );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaDfs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [DFS procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaDfs.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Counts the support size of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectCis_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vSupp *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        Vec_IntPush( vSupp, Gia_ObjId(p, pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManCollectCis_rec( p, Gia_ObjFanin0(pObj), vSupp );
    Gia_ManCollectCis_rec( p, Gia_ObjFanin1(pObj), vSupp );
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectCis(p *Gia_Man_t, pNodes *int32, nNodes int32, vSupp *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i int32; 
    Vec_IntClear( vSupp );
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    for  i = 0; i < nNodes; i++  {
        pObj = Gia_ManObj( p, pNodes[i] );
        if Gia_ObjIsCo(pObj) {
            Gia_ManCollectCis_rec( p, Gia_ObjFanin0(pObj), vSupp );
        } else {
            Gia_ManCollectCis_rec( p, pObj, vSupp ); }
    }
}

/**Function*************************************************************

  Synopsis    [Counts the support size of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectAnds_rec(p *Gia_Man_t, iObj int32, vNodes *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManCollectAnds_rec( p, Gia_ObjFaninId0(pObj, iObj), vNodes );
    Gia_ManCollectAnds_rec( p, Gia_ObjFaninId1(pObj, iObj), vNodes );
    Vec_IntPush( vNodes, iObj );
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectAnds(p *Gia_Man_t, pNodes *int32, nNodes int32, vNodes *Vec_Int_t, vLeaves *Vec_Int_t) {
    var i, iLeaf int32; 
//    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrentId( p, 0 );
    if vLeaves != nil {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLeaves)) && (((iLeaf) = Vec_IntEntry(vLeaves, i)), 1); i++  {
            Gia_ObjSetTravIdCurrentId( p, iLeaf );
        } }
    Vec_IntClear( vNodes );
    for  i = 0; i < nNodes; i++  {
        var pObj *Gia_Obj_t  = Gia_ManObj( p, pNodes[i] );
        if Gia_ObjIsCo(pObj) {
            Gia_ManCollectAnds_rec( p, Gia_ObjFaninId0(pObj, pNodes[i]), vNodes );
        } else if Gia_ObjIsAnd(pObj) {
            Gia_ManCollectAnds_rec( p, pNodes[i], vNodes ); }
    }
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectAndsAll(p *Gia_Man_t) *Vec_Int_t {
    var pObj *Gia_Obj_t; var i int32;
    var vNodes *Vec_Int_t  = Vec_IntAlloc( Gia_ManAndNum(p) );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Vec_IntPush( vNodes, i ); }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Counts the support size of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectNodesCis_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vNodes *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManCollectNodesCis_rec( p, Gia_ObjFanin0(pObj), vNodes );
    Gia_ManCollectNodesCis_rec( p, Gia_ObjFanin1(pObj), vNodes );
    Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectNodesCis(p *Gia_Man_t, pNodes *int32, nNodes int32) *Vec_Int_t {
    var vNodes *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32; 
    vNodes = Vec_IntAlloc( 10000 );
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    for  i = 0; i < nNodes; i++  {
        pObj = Gia_ManObj( p, pNodes[i] );
        if Gia_ObjIsCo(pObj) {
            Gia_ManCollectNodesCis_rec( p, Gia_ObjFanin0(pObj), vNodes );
        } else {
            Gia_ManCollectNodesCis_rec( p, pObj, vNodes ); }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectTest(p *Gia_Man_t) {
    var vNodes *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, iNode int32;
    var clk abctime  = Abc_Clock();
    vNodes = Vec_IntAlloc( 100 );
    Gia_ManIncrementTravId( p );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        iNode = Gia_ObjId(p, pObj);
        Gia_ManCollectAnds( p, &iNode, 1, vNodes, nil );
    }
    Vec_IntFree( vNodes );
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("DFS from each output"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSuppSize_rec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return 0; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        return 1; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    return Gia_ManSuppSize_rec( p, Gia_ObjFanin0(pObj) ) +
        Gia_ManSuppSize_rec( p, Gia_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Computes support size of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSuppSizeOne(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    Gia_ManIncrementTravId( p );
    return Gia_ManSuppSize_rec( p, pObj );
}

/**Function*************************************************************

  Synopsis    [Computes support size of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSuppSizeTest(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    var clk abctime  = Abc_Clock();
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            Counter += (Gia_ManSuppSizeOne(p, pObj) <= 16); }
    }
    fmt.Printf("Nodes with small support %d (out of %d)\n", Counter, Gia_ManAndNum(p));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSuppSize(p *Gia_Man_t, pNodes *int32, nNodes int32) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32; 
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    for  i = 0; i < nNodes; i++  {
        pObj = Gia_ManObj( p, pNodes[i] );
        if Gia_ObjIsCo(pObj) {
            Counter += Gia_ManSuppSize_rec( p, Gia_ObjFanin0(pObj) );
        } else {
            Counter += Gia_ManSuppSize_rec( p, pObj ); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManConeSize_rec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return 0; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        return 0; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    return 1 + Gia_ManConeSize_rec( p, Gia_ObjFanin0(pObj) ) +
        Gia_ManConeSize_rec( p, Gia_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Collects support nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManConeSize(p *Gia_Man_t, pNodes *int32, nNodes int32) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32; 
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    for  i = 0; i < nNodes; i++  {
        pObj = Gia_ManObj( p, pNodes[i] );
        if Gia_ObjIsCo(pObj) {
            Counter += Gia_ManConeSize_rec( p, Gia_ObjFanin0(pObj) );
        } else {
            Counter += Gia_ManConeSize_rec( p, pObj ); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Levelizes the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManLevelize(p *Gia_Man_t) *Vec_Vec_t { 
    var pObj *Gia_Obj_t;
    var vLevels *Vec_Vec_t;
    var nLevels, Level, i int32;
    nLevels = Gia_ManLevelNum( p );
    vLevels = Vec_VecStart( nLevels + 1 );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Level = Gia_ObjLevel( p, pObj );
        assert(Level <= nLevels, "Level <= nLevels");
        Vec_VecPush( vLevels, Level, pObj );
    }
    }
    return vLevels;
}

/**Function*************************************************************

  Synopsis    [Levelizes the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManLevelizeR(p *Gia_Man_t) *Vec_Wec_t { 
    var pObj *Gia_Obj_t;
    var vLevels *Vec_Wec_t;
    var nLevels, Level, i int32;
    nLevels = Gia_ManLevelRNum( p );
    vLevels = Vec_WecStart( nLevels + 1 );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if i == 0 || (!Gia_ObjIsCo(pObj) && !Gia_ObjLevel(p, pObj)) {
            continue; }
        Level = Gia_ObjLevel( p, pObj );
        assert(Level <= nLevels, "Level <= nLevels");
        Vec_WecPush( vLevels, Level, i );
    }
    return vLevels;
}
/**Function*************************************************************

  Synopsis    [Computes reverse topological order.]

  Description [Assumes that levels are already assigned.
  The levels of CO nodes may not be assigned.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManOrderReverse(p *Gia_Man_t) *Vec_Int_t {
    var pObj *Gia_Obj_t;
    var vLevels *Vec_Vec_t;
    var vLevel *Vec_Ptr_t;
    var vResult *Vec_Int_t;
    var i, k int32;
    vLevels = Vec_VecStart( 100 );
    // make sure levels are assigned
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        assert(Gia_ObjLevel(p, pObj) > 0, "Gia_ObjLevel(p, pObj) > 0"); }
    }
    // add CO nodes based on the level of their fanin
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Vec_VecPush( vLevels, Gia_ObjLevel(p, Gia_ObjFanin0(pObj)), pObj );
    }
    // add other nodes based on their level
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if !Gia_ObjIsCo(pObj) {
            Vec_VecPush( vLevels, Gia_ObjLevel(p, pObj), pObj ); }
    }
    // put the nodes in the reverse topological order
    vResult = Vec_IntAlloc( Gia_ManObjNum(p) );
    // inlining of macro Vec_VecForEachLevelReverse
    for  i = Vec_VecSize(vLevels)-1; (i >= 0) && (((vLevel) = Vec_VecEntry(vLevels, i)), 1); i--  {
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vLevel)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vLevel, k)), 1); k++  {
            Vec_IntPush( vResult, Gia_ObjId(p, pObj) );
        }
    }
    Vec_VecFree( vLevels );
    return vResult;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectSeq_rec(p *Gia_Man_t, Id int32, vRoots *Vec_Int_t, vObjs *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId( p, Id ) {
        return; }
    Gia_ObjSetTravIdCurrentId( p, Id );
    pObj = Gia_ManObj( p, Id );
    if Gia_ObjIsAnd(pObj) {
        Gia_ManCollectSeq_rec( p, Gia_ObjFaninId0(pObj, Id), vRoots, vObjs );
        Gia_ManCollectSeq_rec( p, Gia_ObjFaninId1(pObj, Id), vRoots, vObjs );
    
    } else if Gia_ObjIsCi(pObj) {
        if Gia_ObjIsRo(p, pObj) {
            Vec_IntPush( vRoots, Gia_ObjId(p, Gia_ObjRoToRi(p, pObj)) ); }
    
    } else if Gia_ObjIsCo(pObj) {
        Gia_ManCollectSeq_rec( p, Gia_ObjFaninId0(pObj, Id), vRoots, vObjs );
    } else { assert(false, "0"); }
    Vec_IntPush( vObjs, Id );
}
func Gia_ManCollectSeq(p *Gia_Man_t, pPos *int32, nPos int32) *Vec_Int_t {
    var vObjs, vRoots *Vec_Int_t;
    var i, iRoot int32;
    // collect roots
    vRoots = Vec_IntAlloc( 100 );
    for  i = 0; i < nPos; i++  {
        Vec_IntPush( vRoots, Gia_ObjId(p, Gia_ManPo(p, pPos[i])) );
    }
    // start trav IDs
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrentId( p, 0 );
    // collect objects
    vObjs = Vec_IntAlloc( 1000 );
    Vec_IntPush( vObjs, 0 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRoots)) && (((iRoot) = Vec_IntEntry(vRoots, i)), 1); i++  {
        Gia_ManCollectSeq_rec( p, iRoot, vRoots, vObjs );
    }
    Vec_IntFree( vRoots );
    return vObjs;
}
func Gia_ManCollectSeqTest(p *Gia_Man_t) {
    var vObjs *Vec_Int_t;
    var i int32;
    var clk abctime  = Abc_Clock();
    for  i = 0; i < Gia_ManPoNum(p); i++  {
        if i % 10000 == 0 {
            fmt.Printf("%8d finished...\r", i); }

        vObjs = Gia_ManCollectSeq( p, &i, 1 );
//        printf( "%d ", Vec_IntSize(vObjs) );
        Vec_IntFree( vObjs );
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

}

/**Function*************************************************************

  Synopsis    [Collect TFI nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectTfi_rec(p *Gia_Man_t, iObj int32, vNodes *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManCollectTfi_rec( p, Gia_ObjFaninId0(pObj, iObj), vNodes );
    Gia_ManCollectTfi_rec( p, Gia_ObjFaninId1(pObj, iObj), vNodes );
    Vec_IntPush( vNodes, iObj );
}
func Gia_ManCollectTfi(p *Gia_Man_t, vRoots *Vec_Int_t, vNodes *Vec_Int_t) {
    var i, iRoot int32; 
    Vec_IntClear( vNodes );
    Gia_ManIncrementTravId( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRoots)) && (((iRoot) = Vec_IntEntry(vRoots, i)), 1); i++  {
        Gia_ManCollectTfi_rec( p, iRoot, vNodes );
    }
}

/**Function*************************************************************

  Synopsis    [Collect TFI nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectTfo_rec(p *Gia_Man_t, iObj int32, vNodes *Vec_Int_t) {
    var pObj *Gia_Obj_t; var i, iFan int32;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCo(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    // inlining of macro Gia_ObjForEachFanoutStaticId
    for  i = 0; (i < Gia_ObjFanoutNumId(p, iObj))   && (((iFan) = Gia_ObjFanoutId(p, iObj, i)), 1); i++  {
        Gia_ManCollectTfo_rec( p, iFan, vNodes );
    }
    Vec_IntPush( vNodes, iObj );
}
func Gia_ManCollectTfo(p *Gia_Man_t, vRoots *Vec_Int_t, vNodes *Vec_Int_t) {
    var i, iRoot int32; 
    Vec_IntClear( vNodes );
    Gia_ManIncrementTravId( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRoots)) && (((iRoot) = Vec_IntEntry(vRoots, i)), 1); i++  {
        Gia_ManCollectTfo_rec( p, iRoot, vNodes );
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


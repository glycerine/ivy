package saig

import "fmt"

/**CFile****************************************************************

  FileName    [saigWnd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Sequential windowing.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigWnd.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Extracts the window AIG from the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManWindowExtractNodes(p *Aig_Man_t, vNodes *Vec_Ptr_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pMatch *Aig_Obj_t;
    var vNodesPi, vNodesPo *Vec_Ptr_t;
    var i, nRegCount int32;
    Aig_ManCleanData( p ); 
    // create the new manager
    pNew = Aig_ManStart( Vec_PtrSize(vNodes) );
    pNew.pName = Abc_UtilStrsav( "wnd" );
    pNew.pSpec = nil;
    // map constant nodes
    pObj = Aig_ManConst1( p );
    pObj.pData = Aig_ManConst1( pNew );
    // create real PIs
    vNodesPi = Saig_ManWindowCollectPis( p, vNodes );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodesPi)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodesPi, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi(pNew);
    }
    Vec_PtrFree( vNodesPi );
    // create register outputs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Saig_ObjIsLo(p, pObj) {
            pObj.pData = Aig_ObjCreateCi(pNew); }
    }
    // create internal nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Aig_ObjIsNode(pObj) {
            pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // create POs
    vNodesPo = Saig_ManWindowCollectPos( p, vNodes, nil );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodesPo)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodesPo, i)), 1); i++  {
        Aig_ObjCreateCo( pNew, (Aig_Obj_t *)pObj.pData );
    }
    Vec_PtrFree( vNodesPo );
    // create register inputs
    nRegCount = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Saig_ObjIsLo(p, pObj) {
            pMatch = Saig_ObjLoToLi( p, pObj );
            Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pMatch) );
            nRegCount++;
        }
    }
    Aig_ManSetRegNum( pNew, nRegCount );
    Aig_ManCleanup( pNew );
    return pNew;
}

// static void Saig_ManWindowInsertSmall_rec( Aig_Man_t * pNew, Aig_Obj_t * pObjSmall, 
//      Vec_Ptr_t * vBigNode2SmallPo, Vec_Ptr_t * vSmallPi2BigNode );


/**Function*************************************************************

  Synopsis    [Adds nodes for the big manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManWindowInsertBig_rec(pNew *Aig_Man_t, pObjBig *Aig_Obj_t, vBigNode2SmallPo *Vec_Ptr_t, vSmallPi2BigNode *Vec_Ptr_t) {
    var pMatch *Aig_Obj_t;
    if pObjBig.pData {
        return; }
    if pMatch = (Aig_Obj_t *)Vec_PtrEntry( vBigNode2SmallPo, pObjBig.Id ) {
        Saig_ManWindowInsertSmall_rec( pNew, Aig_ObjFanin0(pMatch), vBigNode2SmallPo, vSmallPi2BigNode );
        pObjBig.pData = Aig_ObjChild0Copy(pMatch);
        return;        
    }
    assert(Aig_ObjIsNode(pObjBig), "Aig_ObjIsNode(pObjBig)");
    Saig_ManWindowInsertBig_rec( pNew, Aig_ObjFanin0(pObjBig), vBigNode2SmallPo, vSmallPi2BigNode );
    Saig_ManWindowInsertBig_rec( pNew, Aig_ObjFanin1(pObjBig), vBigNode2SmallPo, vSmallPi2BigNode );
    pObjBig.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObjBig), Aig_ObjChild1Copy(pObjBig) );
}

/**Function*************************************************************

  Synopsis    [Adds nodes for the small manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManWindowInsertSmall_rec(pNew *Aig_Man_t, pObjSmall *Aig_Obj_t, vBigNode2SmallPo *Vec_Ptr_t, vSmallPi2BigNode *Vec_Ptr_t) {
    var pMatch *Aig_Obj_t;
    if pObjSmall.pData {
        return; }
    if pMatch = (Aig_Obj_t *)Vec_PtrEntry( vSmallPi2BigNode, pObjSmall.Id ) {
        Saig_ManWindowInsertBig_rec( pNew, pMatch, vBigNode2SmallPo, vSmallPi2BigNode );
        pObjSmall.pData = pMatch.pData;
        return;        
    }
    assert(Aig_ObjIsNode(pObjSmall), "Aig_ObjIsNode(pObjSmall)");
    Saig_ManWindowInsertSmall_rec( pNew, Aig_ObjFanin0(pObjSmall), vBigNode2SmallPo, vSmallPi2BigNode );
    Saig_ManWindowInsertSmall_rec( pNew, Aig_ObjFanin1(pObjSmall), vBigNode2SmallPo, vSmallPi2BigNode );
    pObjSmall.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObjSmall), Aig_ObjChild1Copy(pObjSmall) );
}

/**Function*************************************************************

  Synopsis    [Extracts the network from the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManWindowInsertNodes(p *Aig_Man_t, vNodes *Vec_Ptr_t, pWnd *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var vBigNode2SmallPo, vSmallPi2BigNode *Vec_Ptr_t;
    var vNodesPi, vNodesPo *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;

    // set mapping of small PIs into big nodes
    vSmallPi2BigNode = Vec_PtrStart( Aig_ManObjNumMax(pWnd) ); 
    vNodesPi = Saig_ManWindowCollectPis( p, vNodes );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodesPi)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodesPi, i)), 1); i++  {
        Vec_PtrWriteEntry( vSmallPi2BigNode, Aig_ManCi(pWnd, i).Id, pObj );
    }
    assert(i == Saig_ManPiNum(pWnd), "i == Saig_ManPiNum(pWnd)");
    Vec_PtrFree( vNodesPi );

    // set mapping of big nodes into small POs
    vBigNode2SmallPo = Vec_PtrStart( Aig_ManObjNumMax(p) ); 
    vNodesPo = Saig_ManWindowCollectPos( p, vNodes, nil );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodesPo)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodesPo, i)), 1); i++  {
        Vec_PtrWriteEntry( vBigNode2SmallPo, pObj.Id, Aig_ManCo(pWnd, i) );
    }
    assert(i == Saig_ManPoNum(pWnd), "i == Saig_ManPoNum(pWnd)");
    Vec_PtrFree( vNodesPo );

    // create the new manager
    Aig_ManCleanData( p ); 
    Aig_ManCleanData( pWnd ); 
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // map constant nodes
    pObj = Aig_ManConst1( p );
    pObj.pData = Aig_ManConst1( pNew );
    pObj = Aig_ManConst1( pWnd );
    pObj.pData = Aig_ManConst1( pNew );

    // create real PIs
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        if Saig_ObjIsPi(p, pObj) || !Aig_ObjIsTravIdCurrent(p, pObj) {
            pObj.pData = Aig_ObjCreateCi(pNew); }
    }
    // create additional latch outputs
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pWnd)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pWnd.vCis, i+Saig_ManPiNum(pWnd))), 1); i++  {
        pObj.pData = Aig_ObjCreateCi(pNew);
    }

    // create internal nodes starting from the big
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        if Saig_ObjIsPo(p, pObj) || !Aig_ObjIsTravIdCurrent(p, pObj) {
            Saig_ManWindowInsertBig_rec( pNew, Aig_ObjFanin0(pObj), vBigNode2SmallPo, vSmallPi2BigNode );
            pObj.pData = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
        }
    }
    // create internal nodes starting from the small
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pWnd)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pWnd.vCos, i+Saig_ManPoNum(pWnd))), 1); i++  {
        Saig_ManWindowInsertSmall_rec( pNew, Aig_ObjFanin0(pObj), vBigNode2SmallPo, vSmallPi2BigNode );
        pObj.pData = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    Vec_PtrFree( vBigNode2SmallPo );
    Vec_PtrFree( vSmallPi2BigNode );
    // set the new number of registers
    assert(Aig_ManCiNum(pNew) - Aig_ManCiNum(p) == Aig_ManCoNum(pNew) - Aig_ManCoNum(p), "Aig_ManCiNum(pNew) - Aig_ManCiNum(p) == Aig_ManCoNum(pNew) - Aig_ManCoNum(p)");
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) + (Aig_ManCiNum(pNew) - Aig_ManCiNum(p)) );
    Aig_ManCleanup( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Find a good object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManFindPivot(p *Aig_Man_t) *Aig_Obj_t {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    if Aig_ManRegNum(p) > 0 {
        if Aig_ManRegNum(p) == 1 {
            return Saig_ManLo( p, 0 ); }
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
            if i == Aig_ManRegNum(p)/2 {
                return pObj; }
        }
    
    } else {
        Counter = 0;
        assert(Aig_ManNodeNum(p) > 1, "Aig_ManNodeNum(p) > 1");
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            if Counter++ == Aig_ManNodeNum(p)/2 {
                return pObj; }
        }
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Computes sequential window of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManWindowExtract(p *Aig_Man_t, pObj *Aig_Obj_t, nDist int32) *Aig_Man_t {
    var pWnd *Aig_Man_t;
    var vNodes *Vec_Ptr_t;
    Aig_ManFanoutStart( p );
    vNodes = Saig_ManWindowOutline( p, pObj, nDist );
    pWnd = Saig_ManWindowExtractNodes( p, vNodes );
    Vec_PtrFree( vNodes );
    Aig_ManFanoutStop( p );
    return pWnd;
}

/**Function*************************************************************

  Synopsis    [Computes sequential window of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManWindowInsert(p *Aig_Man_t, pObj *Aig_Obj_t, nDist int32, pWnd *Aig_Man_t) *Aig_Man_t {
    var pNew, pWndTest *Aig_Man_t;
    var vNodes *Vec_Ptr_t;
    Aig_ManFanoutStart( p );

    vNodes = Saig_ManWindowOutline( p, pObj, nDist );
    pWndTest = Saig_ManWindowExtractNodes( p, vNodes );
    if Saig_ManPiNum(pWndTest) != Saig_ManPiNum(pWnd) ||
         Saig_ManPoNum(pWndTest) != Saig_ManPoNum(pWnd) {
        fmt.Printf("The window cannot be reinserted because PI/PO counts do not match.\n");
        Aig_ManStop( pWndTest );
        Vec_PtrFree( vNodes );
        Aig_ManFanoutStop( p );
        return nil;
    }
    Aig_ManStop( pWndTest );
    Vec_PtrFree( vNodes );

    // insert the nodes
    Aig_ManCleanData( p ); 
    vNodes = Saig_ManWindowOutline( p, pObj, nDist );
    pNew = Saig_ManWindowInsertNodes( p, vNodes, pWnd );
    Vec_PtrFree( vNodes );
    Aig_ManFanoutStop( p );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Tests the above computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManWindowTest(p *Aig_Man_t) *Aig_Man_t {
    var nDist int32  = 3;
    var pWnd, pNew *Aig_Man_t;
    var pPivot *Aig_Obj_t;
    pPivot = Saig_ManFindPivot( p );
    assert(pPivot != nil, "pPivot != nil");
    pWnd = Saig_ManWindowExtract( p, pPivot, nDist );
    pNew = Saig_ManWindowInsert( p, pPivot, nDist, pWnd );
    Aig_ManStop( pWnd );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Collects the nodes that are not linked to each other.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCollectedDiffNodes(p0 *Aig_Man_t, p1 *Aig_Man_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj0, pObj1 *Aig_Obj_t;
    var i int32;
    // collect nodes that are not linked
    Aig_ManIncrementTravId( p0 );
    vNodes = Vec_PtrAlloc( 1000 );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p0.vObjs)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vObjs, i)), 1); i++  { if (pObj0) == nil {} else {
        pObj1 = Aig_ObjRepr( p0, pObj0 );
        if pObj1 != nil {
            assert(pObj0 == Aig_ObjRepr( p1, pObj1 ), "pObj0 == Aig_ObjRepr( p1, pObj1 )");   
            continue;
        }
        // mark and collect unmatched objects
        Aig_ObjSetTravIdCurrent( p0, pObj0 ); 
        if Aig_ObjIsNode(pObj0) || Aig_ObjIsCi(pObj0) {
            Vec_PtrPush( vNodes, pObj0 ); }
    }
    }
    // make sure LI/LO are labeled/unlabeled mutually
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(p0)) && (((pObj0) = Saig_ManLi(p0, i)), 1)            && (((pObj1)=Saig_ManLo(p0, i)), 1); i++  {
        assert(Aig_ObjIsTravIdCurrent(p0, pObj0) == 
                        Aig_ObjIsTravIdCurrent(p0, pObj1), "Aig_ObjIsTravIdCurrent(p0, pObj0) == \n                Aig_ObjIsTravIdCurrent(p0, pObj1)");
    }
    return vNodes;
}
         
/**Function*************************************************************

  Synopsis    [Creates PIs of the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManWindowCreatePis(pNew *Aig_Man_t, p0 *Aig_Man_t, p1 *Aig_Man_t, vNodes0 *Vec_Ptr_t) {
    var pObj, pMatch, pFanin *Aig_Obj_t;
    var i, Counter int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes0)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes0, i)), 1); i++  {
        if Saig_ObjIsLo(p0, pObj) {
            pMatch = Saig_ObjLoToLi( p0, pObj );
            pFanin = Aig_ObjFanin0(pMatch);
            if !Aig_ObjIsTravIdCurrent(p0, pFanin) && pFanin.pData == nil {
                pFanin.pData = Aig_ObjCreateCi(pNew);
                pMatch = Aig_ObjRepr( p0, pFanin );
                assert(pFanin == Aig_ObjRepr( p1, pMatch ), "pFanin == Aig_ObjRepr( p1, pMatch )");
                assert(pMatch != nil, "pMatch != nil");
                pMatch.pData = pFanin.pData;
                Counter++;
            }
        
        } else {
            assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
            pFanin = Aig_ObjFanin0(pObj);
            if !Aig_ObjIsTravIdCurrent(p0, pFanin) && pFanin.pData == nil {
                pFanin.pData = Aig_ObjCreateCi(pNew);
                pMatch = Aig_ObjRepr( p0, pFanin );
                assert(pFanin == Aig_ObjRepr( p1, pMatch ), "pFanin == Aig_ObjRepr( p1, pMatch )");
                assert(pMatch != nil, "pMatch != nil");
                pMatch.pData = pFanin.pData;
                Counter++;
            }
            pFanin = Aig_ObjFanin1(pObj);
            if !Aig_ObjIsTravIdCurrent(p0, pFanin) && pFanin.pData == nil {
                pFanin.pData = Aig_ObjCreateCi(pNew);
                pMatch = Aig_ObjRepr( p0, pFanin );
                assert(pFanin == Aig_ObjRepr( p1, pMatch ), "pFanin == Aig_ObjRepr( p1, pMatch )");
                assert(pMatch != nil, "pMatch != nil");
                pMatch.pData = pFanin.pData;
                Counter++;
            }
        }
    }
//    printf( "Added %d primary inputs.\n", Counter );
}
        
/**Function*************************************************************

  Synopsis    [Creates POs of the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManWindowCreatePos(pNew *Aig_Man_t, p0 *Aig_Man_t, p1 *Aig_Man_t) {
    var pObj0, pObj1, pMiter *Aig_Obj_t;
    var pFanin0, pFanin1 *Aig_Obj_t;
    var i int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p0.vObjs)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vObjs, i)), 1); i++  { if (pObj0) == nil {} else {
        if Aig_ObjIsTravIdCurrent(p0, pObj0) {
            continue; }
        if Aig_ObjIsConst1(pObj0) {
            continue; }
        if Aig_ObjIsCi(pObj0) {
            continue; }
        pObj1 = Aig_ObjRepr( p0, pObj0 );
        assert(pObj0 == Aig_ObjRepr( p1, pObj1 ), "pObj0 == Aig_ObjRepr( p1, pObj1 )");
        if Aig_ObjIsCo(pObj0) {
            pFanin0 = Aig_ObjFanin0(pObj0);
            pFanin1 = Aig_ObjFanin0(pObj1);
            assert(Aig_ObjIsTravIdCurrent(p0, pFanin0) ==
                                Aig_ObjIsTravIdCurrent(p1, pFanin1), "Aig_ObjIsTravIdCurrent(p0, pFanin0) ==\n                    Aig_ObjIsTravIdCurrent(p1, pFanin1)");
            if Aig_ObjIsTravIdCurrent(p0, pFanin0) {
                pMiter = Aig_Exor( pNew, (Aig_Obj_t *)pFanin0.pData, (Aig_Obj_t *)pFanin1.pData );
                Aig_ObjCreateCo( pNew, pMiter );
            }
        
        } else {
            assert(Aig_ObjIsNode(pObj0), "Aig_ObjIsNode(pObj0)");

            pFanin0 = Aig_ObjFanin0(pObj0);
            pFanin1 = Aig_ObjFanin0(pObj1);
            assert(Aig_ObjIsTravIdCurrent(p0, pFanin0) ==
                                Aig_ObjIsTravIdCurrent(p1, pFanin1), "Aig_ObjIsTravIdCurrent(p0, pFanin0) ==\n                    Aig_ObjIsTravIdCurrent(p1, pFanin1)");
            if Aig_ObjIsTravIdCurrent(p0, pFanin0) {
                pMiter = Aig_Exor( pNew, (Aig_Obj_t *)pFanin0.pData, (Aig_Obj_t *)pFanin1.pData );
                Aig_ObjCreateCo( pNew, pMiter );
            }

            pFanin0 = Aig_ObjFanin1(pObj0);
            pFanin1 = Aig_ObjFanin1(pObj1);
            assert(Aig_ObjIsTravIdCurrent(p0, pFanin0) ==
                                Aig_ObjIsTravIdCurrent(p1, pFanin1), "Aig_ObjIsTravIdCurrent(p0, pFanin0) ==\n                    Aig_ObjIsTravIdCurrent(p1, pFanin1)");
            if Aig_ObjIsTravIdCurrent(p0, pFanin0) {
                pMiter = Aig_Exor( pNew, (Aig_Obj_t *)pFanin0.pData, (Aig_Obj_t *)pFanin1.pData );
                Aig_ObjCreateCo( pNew, pMiter );
            }
        }
    }
    }
}


/**Function*************************************************************

  Synopsis    [Extracts the window AIG from the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManWindowExtractMiter(p0 *Aig_Man_t, p1 *Aig_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj0, pObj1, pMatch0, pMatch1 *Aig_Obj_t;
    var vNodes0, vNodes1 *Vec_Ptr_t;
    var i, nRegCount int32;
    // add matching of POs and LIs
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p0)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vCos, i)), 1); i++  {
        pObj1 = Aig_ManCo( p1, i );
        Aig_ObjSetRepr( p0, pObj0, pObj1 );
        Aig_ObjSetRepr( p1, pObj1, pObj0 );
    }
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p0)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vCos, i+Saig_ManPoNum(p0))), 1); i++  {
        pMatch0 = Saig_ObjLiToLo( p0, pObj0 );
        pMatch1 = Aig_ObjRepr( p0, pMatch0 );
        if pMatch1 == nil {
            continue; }
        assert(pMatch0 == Aig_ObjRepr( p1, pMatch1 ), "pMatch0 == Aig_ObjRepr( p1, pMatch1 )");
        pObj1 = Saig_ObjLoToLi( p1, pMatch1 );
        Aig_ObjSetRepr( p0, pObj0, pObj1 );
        Aig_ObjSetRepr( p1, pObj1, pObj0 );
    }
    // clean the markings
    Aig_ManCleanData( p0 ); 
    Aig_ManCleanData( p1 ); 
    // collect nodes that are not linked
    vNodes0 = Saig_ManCollectedDiffNodes( p0, p1 );
    vNodes1 = Saig_ManCollectedDiffNodes( p1, p0 );
    // create the new manager
    pNew = Aig_ManStart( Vec_PtrSize(vNodes0) + Vec_PtrSize(vNodes1) );
    pNew.pName = Abc_UtilStrsav( "wnd" );
    pNew.pSpec = nil;
    // map constant nodes
    pObj0 = Aig_ManConst1( p0 );
    pObj0.pData = Aig_ManConst1( pNew );
    pObj1 = Aig_ManConst1( p1 );
    pObj1.pData = Aig_ManConst1( pNew );
    // create real PIs
    Saig_ManWindowCreatePis( pNew, p0, p1, vNodes0 );
    Saig_ManWindowCreatePis( pNew, p1, p0, vNodes1 );
    // create register outputs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes0)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(vNodes0, i)), 1); i++  {
        if Saig_ObjIsLo(p0, pObj0) {
            pObj0.pData = Aig_ObjCreateCi(pNew); }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes1)) && (((pObj1) = (Aig_Obj_t *)Vec_PtrEntry(vNodes1, i)), 1); i++  {
        if Saig_ObjIsLo(p1, pObj1) {
            pObj1.pData = Aig_ObjCreateCi(pNew); }
    }
    // create internal nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes0)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(vNodes0, i)), 1); i++  {
        if Aig_ObjIsNode(pObj0) {
            pObj0.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj0), Aig_ObjChild1Copy(pObj0) ); }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes1)) && (((pObj1) = (Aig_Obj_t *)Vec_PtrEntry(vNodes1, i)), 1); i++  {
        if Aig_ObjIsNode(pObj1) {
            pObj1.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj1), Aig_ObjChild1Copy(pObj1) ); }
    }
    // create POs
    Saig_ManWindowCreatePos( pNew, p0, p1 );
//    Saig_ManWindowCreatePos( pNew, p1, p0 );
    // create register inputs
    nRegCount = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes0)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(vNodes0, i)), 1); i++  {
        if Saig_ObjIsLo(p0, pObj0) {
            pMatch0 = Saig_ObjLoToLi( p0, pObj0 );
            Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pMatch0) );
            nRegCount++;
        }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes1)) && (((pObj1) = (Aig_Obj_t *)Vec_PtrEntry(vNodes1, i)), 1); i++  {
        if Saig_ObjIsLo(p1, pObj1) {
            pMatch1 = Saig_ObjLoToLi( p1, pObj1 );
            Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pMatch1) );
            nRegCount++;
        }
    }
    Aig_ManSetRegNum( pNew, nRegCount );
    Aig_ManCleanup( pNew );
    Vec_PtrFree( vNodes0 );
    Vec_PtrFree( vNodes1 );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


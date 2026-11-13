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


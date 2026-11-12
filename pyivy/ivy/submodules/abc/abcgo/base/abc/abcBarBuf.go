package abc

import "fmt"

/**CFile****************************************************************

  FileName    [abcHie.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures to handle hierarchy.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcHie.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

#define ABC_OBJ_VOID ((Abc_Obj_t *)(ABC_PTRINT_T)1)

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checks the the hie design has no duplicated networks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCheckSingleInstance(pNtk *Abc_Ntk_t) int32 {
    var pTemp, pModel *Abc_Ntk_t;
    var pBox *Abc_Obj_t;
    var i int32;
    var k int32;
    var RetValue int32 = 1;
    if pNtk.pDesign == nil {
        return 1; }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
        pTemp.fHieVisited = 0;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
        // inlining of macro Abc_NtkForEachBox
        for  k = 0; (k < Vec_PtrSize((pTemp).vBoxes)) && (((pBox) = Abc_NtkBox(pTemp, k)), 1); k++  {
            pModel = (Abc_Ntk_t *)pBox.pData;
            if pModel == nil {
                continue; }
            if Abc_NtkLatchNum(pModel) > 0 {
                fmt.Printf("Network \"%s\" contains %d flops.\n", Abc_NtkName(pNtk), Abc_NtkLatchNum(pModel));
                RetValue = 0;
            }
            if pModel.fHieVisited {
                fmt.Printf("Network \"%s\" contains box \"%s\" whose model \"%s\" is instantiated more than once.\n", Abc_NtkName(pNtk), Abc_ObjName(pBox), Abc_NtkName(pModel));
                RetValue = 0;
            }
            pModel.fHieVisited = 1;
        }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
        pTemp.fHieVisited = 0;
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Collect PIs and POs of internal networks in the topo order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectPiPos_rec(pNet *Abc_Obj_t, vLiMaps *Vec_Ptr_t, vLoMaps *Vec_Ptr_t) int32 {
    extern int Abc_NtkCollectPiPos_int( Abc_Obj_t * pBox, Abc_Ntk_t * pNtk, Vec_Ptr_t * vLiMaps, Vec_Ptr_t * vLoMaps );
    var pObj, pFanin *Abc_Obj_t; 
    var i, Counter int32;
    assert(Abc_ObjIsNet(pNet), "Abc_ObjIsNet(pNet)");
    if Abc_NodeIsTravIdCurrent( pNet ) {
        return 0; }
    Abc_NodeSetTravIdCurrent( pNet );
    pObj = Abc_ObjFanin0(pNet);
    if Abc_ObjIsNode(pObj) {
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
            Counter += Abc_NtkCollectPiPos_rec( pFanin, vLiMaps, vLoMaps );
        } }
    if Abc_ObjIsNode(pObj) {
        return Counter; }
    if Abc_ObjIsBo(pObj) {
        pObj = Abc_ObjFanin0(pObj); }
    assert(Abc_ObjIsBox(pObj), "Abc_ObjIsBox(pObj)");
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanin) = Abc_ObjFanout(pObj, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( Abc_ObjFanout0(pFanin) );
    }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Counter += Abc_NtkCollectPiPos_rec( Abc_ObjFanin0(pFanin), vLiMaps, vLoMaps );
    }
    Counter += Abc_NtkCollectPiPos_int( pObj, Abc_ObjModel(pObj), vLiMaps, vLoMaps );
    return Counter;
}
func Abc_NtkCollectPiPos_int(pBox *Abc_Obj_t, pNtk *Abc_Ntk_t, vLiMaps *Vec_Ptr_t, vLoMaps *Vec_Ptr_t) int32 {
    var pObj *Abc_Obj_t; 
    var i, Counter int32;
    // mark primary inputs
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( Abc_ObjFanout0(pObj) );
    }
    // add primary inputs
    if pBox != nil {
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pBox)) && (((pObj) = Abc_ObjFanin(pBox, i)), 1); i++  {
            Vec_PtrPush( vLiMaps, pObj );
        }
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
            Vec_PtrPush( vLoMaps, pObj );
        }
    }
    // visit primary outputs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Counter += Abc_NtkCollectPiPos_rec( Abc_ObjFanin0(pObj), vLiMaps, vLoMaps );
    }
    // add primary outputs
    if pBox != nil {
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
            Vec_PtrPush( vLiMaps, pObj );
        }
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pBox)) && (((pObj) = Abc_ObjFanout(pBox, i)), 1); i++  {
            Vec_PtrPush( vLoMaps, pObj );
        }
        Counter++;
    }
    return Counter;
}
func Abc_NtkCollectPiPos(pNtk *Abc_Ntk_t, pvLiMaps **Vec_Ptr_t, pvLoMaps **Vec_Ptr_t) int32 {
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    *pvLiMaps = Vec_PtrAlloc( 1000 );
    *pvLoMaps = Vec_PtrAlloc( 1000 );
    return Abc_NtkCollectPiPos_int( nil, pNtk, *pvLiMaps, *pvLoMaps );
}

/**Function*************************************************************

  Synopsis    [Derives logic network with barbufs from the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkToBarBufs_rec(pNtkNew *Abc_Ntk_t, pNet *Abc_Obj_t) *Abc_Obj_t {
    var pObj, pFanin *Abc_Obj_t; 
    var i int32;
    assert(Abc_ObjIsNet(pNet), "Abc_ObjIsNet(pNet)");
    if pNet.pCopy {
        return pNet.pCopy; }
    pObj = Abc_ObjFanin0(pNet);
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    pNet.pCopy = Abc_NtkDupObj( pNtkNew, pObj, 0 );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_ObjAddFanin( pObj.pCopy, Abc_NtkToBarBufs_rec(pNtkNew, pFanin) );
    }
    return pNet.pCopy;
}
func Abc_NtkToBarBufs(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var Buffer string // was [1000]byte in C

    var vLiMaps, vLoMaps *Vec_Ptr_t;
    var pNtkNew, pTemp *Abc_Ntk_t;
    var pLatch, pObjLi, pObjLo *Abc_Obj_t;
    var pObj, pLiMap, pLoMap *Abc_Obj_t;
    var i, k, nBoxes int32;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    if !(Abc_NtkCheckSingleInstance(pNtk) != 0) {
        return nil; }
    assert(pNtk.pDesign != nil, "pNtk.pDesign != nil");
    // start the network
    pNtkNew = Abc_NtkAlloc( ABC_NTK_LOGIC, pNtk.ntkFunc, 1 );
    pNtkNew.pName = Extra_UtilStrsav(pNtk.pName);
    pNtkNew.pSpec = Extra_UtilStrsav(pNtk.pSpec);
    // clone CIs/CIs/boxes
    Abc_NtkCleanCopy_rec( pNtk );
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_ObjFanout0(pObj).pCopy = Abc_NtkDupObj( pNtkNew, pObj, 1 );
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_NtkDupObj( pNtkNew, pObj, 1 );
    }
    // create latches and transfer copy labels
    nBoxes = Abc_NtkCollectPiPos( pNtk, &vLiMaps, &vLoMaps );
    // inlining of macro Vec_PtrForEachEntryTwo
    for  i = 0; (i < Vec_PtrSize(vLiMaps)) && (((pLiMap) = (Abc_Obj_t *)Vec_PtrEntry(vLiMaps, i)), 1) && (((pLoMap) = (Abc_Obj_t *)Vec_PtrEntry(vLoMaps, i)), 1); i++  {
        pObjLi = Abc_NtkCreateBi(pNtkNew);
        pLatch = Abc_NtkCreateLatch(pNtkNew);
        pObjLo = Abc_NtkCreateBo(pNtkNew);
        Abc_ObjAddFanin( pLatch, pObjLi );
        Abc_ObjAddFanin( pObjLo, pLatch );
        pLatch.pData = (void *)ABC_INIT_ZERO;
        pTemp = nil;
        if Abc_ObjFanin0(pLiMap).pNtk != pNtk {
            pTemp = Abc_ObjFanin0(pLiMap).pNtk;
        } else if Abc_ObjFanout0(pLoMap).pNtk != pNtk {
            pTemp = Abc_ObjFanout0(pLoMap).pNtk;
        } else { assert(false, "0"); }
        Buffer = fmt.Sprintf("_%s_in", Abc_NtkName(pTemp));
        Abc_ObjAssignName( pObjLi, Abc_ObjName(Abc_ObjFanin0(pLiMap)), Buffer );
        Buffer = fmt.Sprintf("_%s_out", Abc_NtkName(pTemp));
        Abc_ObjAssignName( pObjLo, Abc_ObjName(Abc_ObjFanout0(pLoMap)), Buffer );
        pLiMap.pCopy = pObjLi;
        Abc_ObjFanout0(pLoMap).pCopy = pObjLo;
        assert(Abc_ObjIsNet(Abc_ObjFanout0(pLoMap)), "Abc_ObjIsNet(Abc_ObjFanout0(pLoMap))");
    }
    Vec_PtrFree( vLiMaps );
    Vec_PtrFree( vLoMaps );
    // rebuild networks
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
        // inlining of macro Abc_NtkForEachCo
        for  k = 0; (k < Abc_NtkCoNum(pTemp)) && (((pObj) = Abc_NtkCo(pTemp, k)), 1); k++  {
            Abc_ObjAddFanin( pObj.pCopy, Abc_NtkToBarBufs_rec(pNtkNew, Abc_ObjFanin0(pObj)) );
        }
    }
    pNtkNew.nBarBufs = Abc_NtkLatchNum(pNtkNew);
    fmt.Printf("Hierarchy reader flattened %d instances of logic boxes and introduced %d barbufs.\n", nBoxes, pNtkNew.nBarBufs);
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Converts the logic with barbufs into a hierarchical network.]

  Description [The base network is the original hierarchical network. The
  second argument is the optimized network with barbufs.  This procedure
  reconstructs the original hierarcical network which adding logic from
  the optimized network.  It is assumed that the PIs/POs of the original
  network one-to-one mapping with the flops of the optimized network.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromBarBufs_rec(pNtkNew *Abc_Ntk_t, pObj *Abc_Obj_t) *Abc_Obj_t {
    var pFanin *Abc_Obj_t; 
    var i int32;
    if pObj.pCopy {
        return pObj.pCopy; }
    Abc_NtkDupObj( pNtkNew, pObj, 0 );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_ObjAddFanin( pObj.pCopy, Abc_NtkFromBarBufs_rec(pNtkNew, pFanin) );
    }
    return pObj.pCopy;
}
func Abc_NtkFromBarBufs(pNtkBase *Abc_Ntk_t, pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew, pTemp *Abc_Ntk_t;
    var vLiMaps, vLoMaps *Vec_Ptr_t;
    var pObj, pLiMap, pLoMap *Abc_Obj_t;
    var i, k int32;
    assert(pNtkBase.pDesign != nil, "pNtkBase.pDesign != nil");
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(Abc_NtkIsNetlist(pNtkBase), "Abc_NtkIsNetlist(pNtkBase)");
    assert(Abc_NtkLatchNum(pNtkBase) == 0, "Abc_NtkLatchNum(pNtkBase) == 0");
    assert(Abc_NtkLatchNum(pNtk) == pNtk.nBarBufs, "Abc_NtkLatchNum(pNtk) == pNtk.nBarBufs");
    assert(Abc_NtkWhiteboxNum(pNtk) == 0, "Abc_NtkWhiteboxNum(pNtk) == 0");
    assert(Abc_NtkBlackboxNum(pNtk) == 0, "Abc_NtkBlackboxNum(pNtk) == 0");
    assert(Abc_NtkPiNum(pNtk) == Abc_NtkPiNum(pNtkBase), "Abc_NtkPiNum(pNtk) == Abc_NtkPiNum(pNtkBase)");
    assert(Abc_NtkPoNum(pNtk) == Abc_NtkPoNum(pNtkBase), "Abc_NtkPoNum(pNtk) == Abc_NtkPoNum(pNtkBase)");
    // start networks
    Abc_NtkCleanCopy_rec( pNtkBase );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtkBase.pDesign.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtkBase.pDesign.vModules, i)), 1); i++  {
        pTemp.pCopy = Abc_NtkStartFrom( pTemp, pNtk.ntkType, pNtk.ntkFunc );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtkBase.pDesign.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtkBase.pDesign.vModules, i)), 1); i++  {
        if pTemp.pAltView {
            pTemp.pCopy.pAltView = pTemp.pAltView.pCopy;
        } else {
            pTemp.pCopy.pAltView = nil;
        }
    }
    // update box models
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtkBase.pDesign.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtkBase.pDesign.vModules, i)), 1); i++  {
        // inlining of macro Abc_NtkForEachBox
        for  k = 0; (k < Vec_PtrSize((pTemp).vBoxes)) && (((pObj) = Abc_NtkBox(pTemp, k)), 1); k++  {
            if Abc_ObjIsWhitebox(pObj) || Abc_ObjIsBlackbox(pObj) {
                pObj.pCopy.pData = Abc_ObjModel(pObj).pCopy; }
        }
    }
    // create the design
    pNtkNew = pNtkBase.pCopy;
    pNtkNew.pDesign = Abc_DesCreate( pNtkBase.pDesign.pName );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtkBase.pDesign.vModules)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtkBase.pDesign.vModules, i)), 1); i++  {
        Abc_DesAddModel( pNtkNew.pDesign, pTemp.pCopy );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pNtkBase.pDesign.vTops)) && (((pTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtkBase.pDesign.vTops, i)), 1); i++  {
        Vec_PtrPush( pNtkNew.pDesign.vTops, pTemp.pCopy );
    }
    assert(Vec_PtrEntry(pNtkNew.pDesign.vTops, 0) == pNtkNew, "Vec_PtrEntry(pNtkNew.pDesign.vTops, 0) == pNtkNew");
    // transfer copy attributes to pNtk
    Abc_NtkCleanCopy( pNtk );
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pObj.pCopy = Abc_NtkPi(pNtkNew, i);
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pObj.pCopy = Abc_NtkPo(pNtkNew, i);
    }
    Abc_NtkCollectPiPos( pNtkBase, &vLiMaps, &vLoMaps );
    assert(Vec_PtrSize(vLiMaps) == Abc_NtkLatchNum(pNtk), "Vec_PtrSize(vLiMaps) == Abc_NtkLatchNum(pNtk)");
    assert(Vec_PtrSize(vLoMaps) == Abc_NtkLatchNum(pNtk), "Vec_PtrSize(vLoMaps) == Abc_NtkLatchNum(pNtk)");
    // inlining of macro Vec_PtrForEachEntryTwo
    for  i = 0; (i < Vec_PtrSize(vLiMaps)) && (((pLiMap) = (Abc_Obj_t *)Vec_PtrEntry(vLiMaps, i)), 1) && (((pLoMap) = (Abc_Obj_t *)Vec_PtrEntry(vLoMaps, i)), 1); i++  {
        pObj = Abc_NtkBox( pNtk, i );
        Abc_ObjFanin0(pObj).pCopy = pLiMap.pCopy; 
        Abc_ObjFanout0(pObj).pCopy = pLoMap.pCopy; 
    }
    Vec_PtrFree( vLiMaps );
    Vec_PtrFree( vLoMaps );
    // create internal nodes
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_ObjAddFanin( pObj.pCopy, Abc_NtkFromBarBufs_rec(pObj.pCopy.pNtk, Abc_ObjFanin0(pObj)) );
    }
    // transfer net names
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if Abc_ObjFanoutNum(pObj.pCopy) == 0 { // handle PI without fanout
            Abc_ObjAddFanin( Abc_NtkCreateNet(pObj.pCopy.pNtk), pObj.pCopy ); }
        Nm_ManStoreIdName( pObj.pCopy.pNtk.pManName, Abc_ObjFanout0(pObj.pCopy).Id, Abc_ObjFanout0(pObj.pCopy).Type, Abc_ObjName(Abc_ObjFanout0(pObj)), nil );
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Nm_ManStoreIdName( pObj.pCopy.pNtk.pManName, Abc_ObjFanin0(pObj.pCopy).Id, Abc_ObjFanin0(pObj.pCopy).Type, Abc_ObjName(Abc_ObjFanin0(pObj)), nil );
    }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Collect nodes in the barbuf-friendly order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkToBarBufsCollect_rec(pObj *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t; 
    var i int32;
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_NtkToBarBufsCollect_rec( pFanin, vNodes );
    }
    Vec_PtrPush( vNodes, pObj );
}
func Abc_NtkToBarBufsCollect(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    assert(pNtk.nBarBufs > 0, "pNtk.nBarBufs > 0");
    assert(pNtk.nBarBufs == Abc_NtkLatchNum(pNtk), "pNtk.nBarBufs == Abc_NtkLatchNum(pNtk)");
    vNodes = Vec_PtrAlloc( Abc_NtkObjNum(pNtk) );
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if i >= Abc_NtkCiNum(pNtk) - pNtk.nBarBufs {
            break; }
        Vec_PtrPush( vNodes, pObj );
        Abc_NodeSetTravIdCurrent( pObj );
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if i < Abc_NtkCoNum(pNtk) - pNtk.nBarBufs {
            continue; }
        Abc_NtkToBarBufsCollect_rec( Abc_ObjFanin0(pObj), vNodes );
        Vec_PtrPush( vNodes, pObj );
        Vec_PtrPush( vNodes, Abc_ObjFanout0(pObj) );
        Vec_PtrPush( vNodes, Abc_ObjFanout0(Abc_ObjFanout0(pObj)) );
        Abc_NodeSetTravIdCurrent( pObj );
        Abc_NodeSetTravIdCurrent( Abc_ObjFanout0(pObj) );
        Abc_NodeSetTravIdCurrent( Abc_ObjFanout0(Abc_ObjFanout0(pObj)) );
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if i >= Abc_NtkCoNum(pNtk) - pNtk.nBarBufs {
            break; }
        Abc_NtkToBarBufsCollect_rec( Abc_ObjFanin0(pObj), vNodes );
        Vec_PtrPush( vNodes, pObj );
        Abc_NodeSetTravIdCurrent( pObj );
    }
    assert(Vec_PtrSize(vNodes) == Abc_NtkObjNum(pNtk), "Vec_PtrSize(vNodes) == Abc_NtkObjNum(pNtk)");
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Count barrier buffers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCountBarBufs(pNtk *Abc_Ntk_t) int32 {
    var pObj *Abc_Obj_t;
    var i, Counter int32;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Counter += Abc_ObjIsBarBuf( pObj ); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Converts the network to dedicated barbufs and back.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkBarBufsToBuffers(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var vNodes *Vec_Ptr_t;
    var pNtkNew *Abc_Ntk_t;
    var pObj, pFanin *Abc_Obj_t;
    var i, k int32;
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    assert(pNtk.pDesign == nil, "pNtk.pDesign == nil");
    assert(pNtk.nBarBufs > 0, "pNtk.nBarBufs > 0");
    assert(pNtk.nBarBufs == Abc_NtkLatchNum(pNtk), "pNtk.nBarBufs == Abc_NtkLatchNum(pNtk)");
    vNodes = Abc_NtkToBarBufsCollect( pNtk );
    // start the network
    pNtkNew = Abc_NtkAlloc( ABC_NTK_LOGIC, pNtk.ntkFunc, 1 );
    pNtkNew.pName = Extra_UtilStrsav(pNtk.pName);
    pNtkNew.pSpec = Extra_UtilStrsav(pNtk.pSpec);
    // create objects
    Abc_NtkCleanCopy( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Abc_ObjIsPi(pObj) {
            Abc_NtkDupObj( pNtkNew, pObj, 1 );
        } else if Abc_ObjIsPo( pObj) {
            Abc_ObjAddFanin( Abc_NtkDupObj(pNtkNew, pObj, 1), Abc_ObjFanin0(pObj).pCopy );
        } else if Abc_ObjIsBi(pObj) || Abc_ObjIsBo(pObj) {
            pObj.pCopy = Abc_ObjFanin0(pObj).pCopy;
        } else if Abc_ObjIsLatch(pObj) {
            Abc_ObjAddFanin( (pObj.pCopy = Abc_NtkCreateNode(pNtkNew)), Abc_ObjFanin0(pObj).pCopy );
        } else if Abc_ObjIsNode(pObj) {
            Abc_NtkDupObj( pNtkNew, pObj, 1 );
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
                Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy );
            }
        
        } else { assert(false, "0"); }
    }
    Vec_PtrFree( vNodes );
    return pNtkNew;
}
func Abc_NtkBarBufsFromBuffers(pNtkBase *Abc_Ntk_t, pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pObj, pFanin, pLatch *Abc_Obj_t;
    var i, k, nBarBufs int32;
    assert(Abc_NtkIsLogic(pNtkBase), "Abc_NtkIsLogic(pNtkBase)");
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    assert(pNtkBase.nBarBufs == Abc_NtkLatchNum(pNtkBase), "pNtkBase.nBarBufs == Abc_NtkLatchNum(pNtkBase)");
    // start the network
    pNtkNew = Abc_NtkStartFrom( pNtkBase, pNtk.ntkType, pNtk.ntkFunc );
    // transfer PI pointers
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pObj.pCopy = Abc_NtkPi(pNtkNew, i);
    }
    // assuming that the order/number of barbufs remains the same
    nBarBufs = 0;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        if Abc_ObjIsBarBuf(pObj) {
            pLatch = Abc_NtkBox(pNtkNew, nBarBufs++);
            Abc_ObjAddFanin( Abc_ObjFanin0(pLatch), Abc_ObjFanin0(pObj).pCopy );
            pObj.pCopy = Abc_ObjFanout0(pLatch);
        
        } else {
            Abc_NtkDupObj( pNtkNew, pObj, 1 );
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
                Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy );
            }
        }
    }
    }
    assert(nBarBufs == pNtkBase.nBarBufs, "nBarBufs == pNtkBase.nBarBufs");
    // connect POs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Abc_ObjAddFanin( Abc_NtkPo(pNtkNew, i), Abc_ObjFanin0(pObj).pCopy );
    }
    return pNtkNew;
}
func Abc_NtkBarBufsOnOffTest(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkNew, pNtkNew2 *Abc_Ntk_t;
    pNtkNew  = Abc_NtkBarBufsToBuffers( pNtk );
    pNtkNew2 = Abc_NtkBarBufsFromBuffers( pNtk, pNtkNew );
    Abc_NtkDelete( pNtkNew );
    return pNtkNew2;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


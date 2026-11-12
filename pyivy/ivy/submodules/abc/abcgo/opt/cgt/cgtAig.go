package cgt

/**CFile****************************************************************

  FileName    [cgtAig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Clock gating package.]

  Synopsis    [Creates AIG to compute clock-gating.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cgtAig.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cgtInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes transitive fanout cone of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManDetectCandidates_rec(pAig *Aig_Man_t, vUseful *Vec_Int_t, pObj *Aig_Obj_t, nLevelMax int32, vCands *Vec_Ptr_t) {
    if Aig_ObjIsTravIdCurrent(pAig, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(pAig, pObj);
    if Aig_ObjIsNode(pObj) {
        Cgt_ManDetectCandidates_rec( pAig, vUseful, Aig_ObjFanin0(pObj), nLevelMax, vCands );
        Cgt_ManDetectCandidates_rec( pAig, vUseful, Aig_ObjFanin1(pObj), nLevelMax, vCands );
    }
    if Aig_ObjLevel(pObj) <= nLevelMax && (vUseful == nil || Vec_IntEntry(vUseful, Aig_ObjId(pObj))) {
        Vec_PtrPush( vCands, pObj ); }
}

/**Function*************************************************************

  Synopsis    [Computes transitive fanout cone of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManDetectCandidates(pAig *Aig_Man_t, vUseful *Vec_Int_t, pObj *Aig_Obj_t, nLevelMax int32, vCands *Vec_Ptr_t) {
    Vec_PtrClear( vCands );
    if !Aig_ObjIsNode(pObj) {
        return; }
    Aig_ManIncrementTravId( pAig );
    Cgt_ManDetectCandidates_rec( pAig, vUseful, pObj, nLevelMax, vCands );
}

/**Function*************************************************************

  Synopsis    [Computes transitive fanout cone of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManDetectFanout_rec(pAig *Aig_Man_t, pObj *Aig_Obj_t, nOdcMax int32, vFanout *Vec_Ptr_t) {
    var pFanout *Aig_Obj_t;
    var f int32;
    var iFanout int32 = -1;
    if Aig_ObjIsCo(pObj) || Aig_ObjLevel(pObj) > nOdcMax {
        return; }
    if Aig_ObjIsTravIdCurrent(pAig, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(pAig, pObj);
    Vec_PtrPush( vFanout, pObj );
    // inlining of macro Aig_ObjForEachFanout
    for  assert(pAig.pFanData, "pAig.pFanData"), f = 0; (f < int32((pObj).nRefs)) &&                         (((iFanout) = f? Aig_ObjFanoutNext(pAig, iFanout) : Aig_ObjFanout0Int(pAig, pObj.Id)), 1) &&           (((pFanout) = Aig_ManObj(pAig, iFanout>>1)), 1); f++  {
        Cgt_ManDetectFanout_rec( pAig, pFanout, nOdcMax, vFanout );
    }
}

/**Function*************************************************************

  Synopsis    [Computes transitive fanout cone of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManDetectFanout(pAig *Aig_Man_t, pObj *Aig_Obj_t, nOdcMax int32, vFanout *Vec_Ptr_t) {
    var pFanout *Aig_Obj_t;
    var i int32;
    var k int32;
    var f int32;
    var iFanout int32 = -1;
    // collect visited nodes
    Vec_PtrClear( vFanout );
    Aig_ManIncrementTravId( pAig );
    Cgt_ManDetectFanout_rec( pAig, pObj, nOdcMax, vFanout );
    // remove those nodes whose fanout is included
    k = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanout)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vFanout, i)), 1); i++  {
        // go through the fanouts of this node
        // inlining of macro Aig_ObjForEachFanout
        for  assert(pAig.pFanData, "pAig.pFanData"), f = 0; (f < int32((pObj).nRefs)) &&                         (((iFanout) = f? Aig_ObjFanoutNext(pAig, iFanout) : Aig_ObjFanout0Int(pAig, pObj.Id)), 1) &&           (((pFanout) = Aig_ManObj(pAig, iFanout>>1)), 1); f++  {
            if !Aig_ObjIsTravIdCurrent(pAig, pFanout) {
                break; }
        }
        if f == Aig_ObjRefs(pObj) { // all fanouts are included
            continue; }
        Vec_PtrWriteEntry( vFanout, k++, pObj );
    }
    Vec_PtrShrink( vFanout, k );
    Vec_PtrSort( vFanout, (int (*)(void))Aig_ObjCompareIdIncrease );
    assert(Vec_PtrSize(vFanout) > 0, "Vec_PtrSize(vFanout) > 0");
}

/**Function*************************************************************

  Synopsis    [Computes visited nodes in the topological order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManCollectVisited_rec(pAig *Aig_Man_t, pObj *Aig_Obj_t, vVisited *Vec_Ptr_t) {
    if Aig_ObjIsCi(pObj) {
        return; }
    if Aig_ObjIsTravIdCurrent(pAig, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(pAig, pObj);
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Cgt_ManCollectVisited_rec( pAig, Aig_ObjFanin0(pObj), vVisited );
    Cgt_ManCollectVisited_rec( pAig, Aig_ObjFanin1(pObj), vVisited );
    Vec_PtrPush( vVisited, pObj );
}

/**Function*************************************************************

  Synopsis    [Computes visited nodes in the topological order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManCollectVisited(pAig *Aig_Man_t, vFanout *Vec_Ptr_t, vVisited *Vec_Ptr_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Vec_PtrClear( vVisited );
    Aig_ManIncrementTravId( pAig );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanout)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vFanout, i)), 1); i++  {
        Cgt_ManCollectVisited_rec( pAig, pObj, vVisited );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjChild0CopyVec(vCopy *Vec_Ptr_t, pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_NotCond((Aig_Obj_t *)Vec_PtrEntry(vCopy, Aig_ObjFaninId0(pObj)), Aig_ObjFaninC0(pObj));  }
func Aig_ObjChild1CopyVec(vCopy *Vec_Ptr_t, pObj *Aig_Obj_t) *Aig_Obj_t { return Aig_NotCond((Aig_Obj_t *)Vec_PtrEntry(vCopy, Aig_ObjFaninId1(pObj)), Aig_ObjFaninC1(pObj));  }

/**Function*************************************************************

  Synopsis    [Derives miter for clock-gating.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManConstructCareCondition(p *Cgt_Man_t, pNew *Aig_Man_t, pObjLo *Aig_Obj_t, vCopy0 *Vec_Ptr_t, vCopy1 *Vec_Ptr_t) *Aig_Obj_t {
    var pMiter, pObj, pTemp *Aig_Obj_t;
    var i int32;
    assert(Aig_ObjIsCi(pObjLo), "Aig_ObjIsCi(pObjLo)");
    // detect nodes and their cone
    Cgt_ManDetectFanout( p.pAig, pObjLo, p.pPars.nOdcMax, p.vFanout );
    Cgt_ManCollectVisited( p.pAig, p.vFanout, p.vVisited );
    // add new variables if the observability condition depends on PI variables
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vVisited)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vVisited, i)), 1); i++  {
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        if Saig_ObjIsPi(p.pAig, Aig_ObjFanin0(pObj)) && Vec_PtrEntry(vCopy0, Aig_ObjFaninId0(pObj)) == nil {
            pTemp = Aig_ObjCreateCi( pNew );
            Vec_PtrWriteEntry( vCopy0, Aig_ObjFaninId0(pObj), pTemp );
            Vec_PtrWriteEntry( vCopy1, Aig_ObjFaninId0(pObj), pTemp );
        }
        if Saig_ObjIsPi(p.pAig, Aig_ObjFanin1(pObj)) && Vec_PtrEntry(vCopy0, Aig_ObjFaninId1(pObj)) == nil {
            pTemp = Aig_ObjCreateCi( pNew );
            Vec_PtrWriteEntry( vCopy0, Aig_ObjFaninId1(pObj), pTemp );
            Vec_PtrWriteEntry( vCopy1, Aig_ObjFaninId1(pObj), pTemp );
        }
    }
    // construct AIGs for the nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vVisited)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vVisited, i)), 1); i++  {
        pTemp = Aig_And( pNew, Aig_ObjChild0CopyVec(vCopy0, pObj), Aig_ObjChild1CopyVec(vCopy0, pObj) );
        Vec_PtrWriteEntry( vCopy0, Aig_ObjId(pObj), pTemp );
        pTemp = Aig_And( pNew, Aig_ObjChild0CopyVec(vCopy1, pObj), Aig_ObjChild1CopyVec(vCopy1, pObj) );
        Vec_PtrWriteEntry( vCopy1, Aig_ObjId(pObj), pTemp );
    }
    // construct the care miter
    pMiter = Aig_ManConst0( pNew );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vFanout)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vFanout, i)), 1); i++  {
        pTemp = Aig_Exor( pNew, (Aig_Obj_t *)Vec_PtrEntry(vCopy0, Aig_ObjId(pObj)), (Aig_Obj_t *)Vec_PtrEntry(vCopy1, Aig_ObjId(pObj)) );
        pMiter = Aig_Or( pNew, pMiter, pTemp );
    }
    return pMiter;
}

/**Function*************************************************************

  Synopsis    [Derives AIG for clock-gating.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManDeriveAigForGating(p *Cgt_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjLi, pObjLo, pCare, pMiter *Aig_Obj_t;
    var vCopy0, vCopy1 *Vec_Ptr_t;
    var i int32;
    assert(Aig_ManRegNum(p.pAig), "Aig_ManRegNum(p.pAig)");
    pNew = Aig_ManStart( Aig_ManObjNumMax(p.pAig) );
    pNew.pName = Abc_UtilStrsav( "CG_miter" );
    // build the first frame
    Aig_ManConst1(p.pAig).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
//    Saig_ManForEachPo( p->pAig, pObj, i )
//        pObj->pData = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    if p.pPars.nOdcMax > 0 {
        // create storage for observability conditions
        vCopy0 = Vec_PtrStart( Aig_ManObjNumMax(p.pAig) );
        vCopy1 = Vec_PtrStart( Aig_ManObjNumMax(p.pAig) );
        // initialize register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(p.pAig, i)), 1); i++  {
            Vec_PtrWriteEntry( vCopy0, Aig_ObjId(pObjLo), Aig_ObjChild0Copy(pObjLi) );
            Vec_PtrWriteEntry( vCopy1, Aig_ObjId(pObjLo), Aig_ObjChild0Copy(pObjLi) );
        }
        // compute observability condition for each latch output
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(p.pAig, i)), 1); i++  {
            // set the constants
            Vec_PtrWriteEntry( vCopy0, Aig_ObjId(pObjLo), Aig_ManConst0(pNew) );
            Vec_PtrWriteEntry( vCopy1, Aig_ObjId(pObjLo), Aig_ManConst1(pNew) );
            // compute condition
            pCare = Cgt_ManConstructCareCondition( p, pNew, pObjLo, vCopy0, vCopy1 );
            // restore the values
            Vec_PtrWriteEntry( vCopy0, Aig_ObjId(pObjLo), Aig_ObjChild0Copy(pObjLi) );
            Vec_PtrWriteEntry( vCopy1, Aig_ObjId(pObjLo), Aig_ObjChild0Copy(pObjLi) );
            // compute the miter
            pMiter = Aig_Exor( pNew, (Aig_Obj_t *)pObjLo.pData, Aig_ObjChild0Copy(pObjLi) );
            pMiter = Aig_And( pNew, pMiter, pCare );
            pObjLi.pData = Aig_ObjCreateCo( pNew, pMiter );
        }
        Vec_PtrFree( vCopy0 );
        Vec_PtrFree( vCopy1 );
    
    } else {
        // construct clock-gating miters for each register input
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(p.pAig, i)), 1); i++  {
            pMiter = Aig_Exor( pNew, (Aig_Obj_t *)pObjLo.pData, Aig_ObjChild0Copy(pObjLi) );
            pObjLi.pData = Aig_ObjCreateCo( pNew, pMiter );
        }
    }
    Aig_ManCleanup( pNew );
    Aig_ManSetCioIds( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Adds relevant constraints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManConstructCare_rec(pCare *Aig_Man_t, pObj *Aig_Obj_t, pNew *Aig_Man_t) *Aig_Obj_t {
    var pObj0, pObj1 *Aig_Obj_t;
    if Aig_ObjIsTravIdCurrent( pCare, pObj ) {
        return (Aig_Obj_t *)pObj.pData; }
    Aig_ObjSetTravIdCurrent( pCare, pObj );
    if Aig_ObjIsCi(pObj) {
        return (Aig_Obj_t *)(pObj.pData = nil); }
    pObj0 = Cgt_ManConstructCare_rec( pCare, Aig_ObjFanin0(pObj), pNew );
    if pObj0 == nil {
        return (Aig_Obj_t *)(pObj.pData = nil); }
    pObj1 = Cgt_ManConstructCare_rec( pCare, Aig_ObjFanin1(pObj), pNew );
    if pObj1 == nil {
        return (Aig_Obj_t *)(pObj.pData = nil); }
    pObj0 = Aig_NotCond( pObj0, Aig_ObjFaninC0(pObj) );
    pObj1 = Aig_NotCond( pObj1, Aig_ObjFaninC1(pObj) );
    return (Aig_Obj_t *)(pObj.pData = Aig_And( pNew, pObj0, pObj1 ));
}

/**Function*************************************************************

  Synopsis    [Builds constraints belonging to the given partition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManConstructCare(pNew *Aig_Man_t, pCare *Aig_Man_t, vSuppsInv *Vec_Vec_t, vLeaves *Vec_Ptr_t) {
    var vOuts *Vec_Int_t;
    var pLeaf, pPi, pPo, pObjAig *Aig_Obj_t;
    var i, k, iOut int32;
    // go through the PIs of the partition
    // label the corresponding PIs of the care set
    Aig_ManIncrementTravId( pCare );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pLeaf) = (Aig_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pPi = Aig_ManCi( pCare, Aig_ObjCioId(pLeaf) );
        Aig_ObjSetTravIdCurrent( pCare, pPi );
        pPi.pData = pLeaf.pData;
    }
    // construct the constraints
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pLeaf) = (Aig_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        vOuts = Vec_VecEntryInt( vSuppsInv, Aig_ObjCioId(pLeaf) );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vOuts)) && (((iOut) = Vec_IntEntry(vOuts, k)), 1); k++  {
            pPo = Aig_ManCo( pCare, iOut );
            if Aig_ObjIsTravIdCurrent( pCare, pPo ) {
                continue; }
            Aig_ObjSetTravIdCurrent( pCare, pPo );
            if Aig_ObjFanin0(pPo) == Aig_ManConst1(pCare) {
                continue; }
            pObjAig = Cgt_ManConstructCare_rec( pCare, Aig_ObjFanin0(pPo), pNew );
            if pObjAig == nil {
                continue; }
            pObjAig = Aig_NotCond( pObjAig, Aig_ObjFaninC0(pPo) );
            Aig_ObjCreateCo( pNew, pObjAig );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManDupPartition_rec(pNew *Aig_Man_t, pAig *Aig_Man_t, pObj *Aig_Obj_t, vLeaves *Vec_Ptr_t) *Aig_Obj_t {
    if Aig_ObjIsTravIdCurrent(pAig, pObj) {
        return (Aig_Obj_t *)pObj.pData; }
    Aig_ObjSetTravIdCurrent(pAig, pObj);
    if Aig_ObjIsCi(pObj) {
        pObj.pData = Aig_ObjCreateCi( pNew );
        Vec_PtrPush( vLeaves, pObj );
        return (Aig_Obj_t *)pObj.pData;
    }
    Cgt_ManDupPartition_rec( pNew, pAig, Aig_ObjFanin0(pObj), vLeaves );
    Cgt_ManDupPartition_rec( pNew, pAig, Aig_ObjFanin1(pObj), vLeaves );
    return (Aig_Obj_t *)(pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ));
}

/**Function*************************************************************

  Synopsis    [Duplicates register outputs starting from the given one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManDupPartition(pFrame *Aig_Man_t, nVarsMin int32, nFlopsMin int32, iStart int32, pCare *Aig_Man_t, vSuppsInv *Vec_Vec_t, pnOutputs *int32) *Aig_Man_t {
    var vRoots, vLeaves, vPos *Vec_Ptr_t;
    var pNew *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    assert(Aig_ManRegNum(pFrame) == 0, "Aig_ManRegNum(pFrame) == 0");
    vRoots = Vec_PtrAlloc( 100 );
    vLeaves = Vec_PtrAlloc( 100 ); 
    vPos = Vec_PtrAlloc( 100 );
    pNew = Aig_ManStart( nVarsMin );
    pNew.pName = Abc_UtilStrsav( "partition" );
    Aig_ManIncrementTravId( pFrame );
    Aig_ManConst1(pFrame).pData = Aig_ManConst1(pNew);
    Aig_ObjSetTravIdCurrent( pFrame, Aig_ManConst1(pFrame) );
    for  i = iStart; i < iStart + nFlopsMin && i < Aig_ManCoNum(pFrame); i++  {
        pObj = Aig_ManCo( pFrame, i );
        Cgt_ManDupPartition_rec( pNew, pFrame, Aig_ObjFanin0(pObj), vLeaves );
        Vec_PtrPush( vRoots, Aig_ObjChild0Copy(pObj) );
        Vec_PtrPush( vPos, pObj );
    }
    for  ; Aig_ManObjNum(pNew) < nVarsMin && i < Aig_ManCoNum(pFrame); i++  {
        pObj = Aig_ManCo( pFrame, i );
        Cgt_ManDupPartition_rec( pNew, pFrame, Aig_ObjFanin0(pObj), vLeaves );
        Vec_PtrPush( vRoots, Aig_ObjChild0Copy(pObj) );
        Vec_PtrPush( vPos, pObj );
    }
    assert(nFlopsMin >= Vec_PtrSize(vRoots) || Vec_PtrSize(vRoots) >= nFlopsMin, "nFlopsMin >= Vec_PtrSize(vRoots) || Vec_PtrSize(vRoots) >= nFlopsMin");
    // create constaints
    if pCare != nil {
        Cgt_ManConstructCare( pNew, pCare, vSuppsInv, vLeaves ); }
    // create POs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vPos, i)), 1); i++  {
        pObj.pData = (Aig_Obj_t *)Aig_ObjCreateCo( pNew, (Aig_Obj_t *)Vec_PtrEntry(vRoots, i) );
    }
    if pnOutputs != nil {
        *pnOutputs = Vec_PtrSize( vPos ); }
    Vec_PtrFree( vRoots );
    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vPos );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Implements one clock-gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManBuildClockGate(pNew *Aig_Man_t, vGates *Vec_Ptr_t) *Aig_Obj_t {
    var pGate, pTotal *Aig_Obj_t;
    var i int32;
    assert(Vec_PtrSize(vGates) > 0, "Vec_PtrSize(vGates) > 0");
    pTotal = Aig_ManConst0(pNew);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGates)) && (((pGate) = (Aig_Obj_t *)Vec_PtrEntry(vGates, i)), 1); i++  {
        if Aig_Regular(pGate).pNext {
            pGate = Aig_NotCond( Aig_Regular(pGate).pNext, Aig_IsComplement(pGate) );
        } else {
            pGate = Aig_NotCond( (Aig_Obj_t *)Aig_Regular(pGate).pData, Aig_IsComplement(pGate) ); }
        pTotal = Aig_Or( pNew, pTotal, pGate );
    }
    return pTotal;
}

/**Function*************************************************************

  Synopsis    [Derives AIG after clock-gating.]

  Description [The array contains, for each flop, its gate if present.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cgt_ManDeriveGatedAig(pAig *Aig_Man_t, vGates *Vec_Vec_t, fReduce int32, pnUsedNodes *int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pObjNew, pObjLi, pObjLo, pGateNew *Aig_Obj_t;
    var vOne *Vec_Ptr_t;
    var i, k int32;
    Aig_ManCleanNext( pAig );
    // label nodes
    // inlining of macro Vec_VecForEachEntry
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; i < Vec_VecSize(vGates); i++  {                                                        for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(vGates, i))) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(Vec_VecEntry(vGates, i), k)), 1); k++  {
        if Aig_IsComplement(pObj) {
            Aig_Regular(pObj).fMarkB = 1;
        } else {
            Aig_Regular(pObj).fMarkA = 1; }
    }
    }
    // construct AIG
    assert(Aig_ManRegNum(pAig), "Aig_ManRegNum(pAig)");
    pNew = Aig_ManStart( Aig_ManObjNumMax(pAig) );
    pNew.pName = Abc_UtilStrsav( pAig.pName );
    pNew.pSpec = Abc_UtilStrsav( pAig.pSpec );
    Aig_ManConst1(pAig).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pNew );
    }
    if fReduce != 0 {
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            assert(!(pObj.fMarkA && pObj.fMarkB), "!(pObj.fMarkA && pObj.fMarkB)");
            pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
            if pObj.fMarkA {
                pObj.pNext = (Aig_Obj_t *)pObj.pData;
                pObj.pData = Aig_ManConst0(pNew);
            
            } else if pObj.fMarkB {
                pObj.pNext = (Aig_Obj_t *)pObj.pData;
                pObj.pData = Aig_ManConst1(pNew);
            }
        }
        }
    
    } else {
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
        }
    }
    if pnUsedNodes != nil {
        *pnUsedNodes = Aig_ManNodeNum(pNew); }
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        pObj.pData = Aig_ObjCreateCo( pNew, Aig_ObjChild0Copy(pObj) );
    }
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
        vOne = Vec_VecEntry( vGates, i );
        if Vec_PtrSize(vOne) == 0 {
            pObjNew = Aig_ObjChild0Copy(pObjLi);
        } else {
//            pGateNew = Aig_NotCond( Aig_Regular(pGate)->pData, Aig_IsComplement(pGate) );
            pGateNew = Cgt_ManBuildClockGate( pNew, vOne );
            pObjNew = Aig_Mux( pNew, pGateNew, (Aig_Obj_t *)pObjLo.pData, Aig_ObjChild0Copy(pObjLi) );
        }
        pObjLi.pData = Aig_ObjCreateCo( pNew, pObjNew );
    }
    Aig_ManCleanup( pNew );
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(pAig) );
    // unlabel nodes
    Aig_ManCleanMarkAB( pAig );
    Aig_ManCleanNext( pAig );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


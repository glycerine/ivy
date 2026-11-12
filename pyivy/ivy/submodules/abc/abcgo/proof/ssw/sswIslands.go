package ssw

import "time"

/**CFile****************************************************************

  FileName    [sswIslands.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Detection of islands of difference.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswIslands.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates pair of structurally equivalent nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_CreatePair(vPairs *Vec_Int_t, pObj0 *Aig_Obj_t, pObj1 *Aig_Obj_t) {
    pObj0.pData = pObj1;
    pObj1.pData = pObj0;
    Vec_IntPush( vPairs, pObj0.Id );
    Vec_IntPush( vPairs, pObj1.Id );
}

/**Function*************************************************************

  Synopsis    [Establishes relationship between nodes using pairing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_MatchingStart(p0 *Aig_Man_t, p1 *Aig_Man_t, vPairs *Vec_Int_t) {
    var pObj0, pObj1 *Aig_Obj_t;
    var i int32;
    // create matching
    Aig_ManCleanData( p0 );
    Aig_ManCleanData( p1 );
    for  i = 0; i < Vec_IntSize(vPairs); i += 2  {
        pObj0 = Aig_ManObj( p0, Vec_IntEntry(vPairs, i) );
        pObj1 = Aig_ManObj( p1, Vec_IntEntry(vPairs, i+1) );
        assert(pObj0.pData == nil, "pObj0.pData == nil");
        assert(pObj1.pData == nil, "pObj1.pData == nil");
        pObj0.pData = pObj1;
        pObj1.pData = pObj0;
    }
    // make sure constants are matched
    pObj0 = Aig_ManConst1( p0 );
    pObj1 = Aig_ManConst1( p1 );
    assert(pObj0.pData == pObj1, "pObj0.pData == pObj1");
    assert(pObj1.pData == pObj0, "pObj1.pData == pObj0");
    // make sure PIs are matched
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p0)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vCis, i)), 1); i++  {
        pObj1 = Aig_ManCi( p1, i );
        assert(pObj0.pData == pObj1, "pObj0.pData == pObj1");
        assert(pObj1.pData == pObj0, "pObj1.pData == pObj0");
    }
    // make sure the POs are not matched
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p0.vCos)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vCos, i)), 1); i++  {
        pObj1 = Aig_ManCo( p1, i );
        assert(pObj0.pData == nil, "pObj0.pData == nil");
        assert(pObj1.pData == nil, "pObj1.pData == nil");
    }

    // check that LIs/LOs are matched in sync
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p0)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vCis, i+Saig_ManPiNum(p0))), 1); i++  {
        if pObj0.pData == nil {
            continue; }
        pObj1 = (Aig_Obj_t *)pObj0.pData;
        if !Saig_ObjIsLo(p1, pObj1) {
            Abc_Print( 1, "Mismatch between LO pairs.\n" ); }
    }
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p1)) && (((pObj1) = (Aig_Obj_t *)Vec_PtrEntry(p1.vCis, i+Saig_ManPiNum(p1))), 1); i++  {
        if pObj1.pData == nil {
            continue; }
        pObj0 = (Aig_Obj_t *)pObj1.pData;
        if !Saig_ObjIsLo(p0, pObj0) {
            Abc_Print( 1, "Mismatch between LO pairs.\n" ); }
    }
}

/**Function*************************************************************

  Synopsis    [Establishes relationship between nodes using pairing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_MatchingExtendOne(p *Aig_Man_t, vNodes *Vec_Ptr_t) {
    var pNext, pObj *Aig_Obj_t;
    var i int32;
    var k int32;
    var iFan int32 = -1;
    Vec_PtrClear( vNodes );
    Aig_ManIncrementTravId( p );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsNode(pObj) && !Aig_ObjIsCi(pObj) {
            continue; }
        if pObj.pData != nil {
            continue; }
        if Saig_ObjIsLo(p, pObj) {
            pNext = Saig_ObjLoToLi(p, pObj);
            pNext = Aig_ObjFanin0(pNext);
            if pNext.pData && !Aig_ObjIsTravIdCurrent(p, pNext) && !Aig_ObjIsConst1(pNext) {
                Aig_ObjSetTravIdCurrent(p, pNext);
                Vec_PtrPush( vNodes, pNext );
            }
        }
        if Aig_ObjIsNode(pObj) {
            pNext = Aig_ObjFanin0(pObj);
            if pNext.pData && !Aig_ObjIsTravIdCurrent(p, pNext) {
                Aig_ObjSetTravIdCurrent(p, pNext);
                Vec_PtrPush( vNodes, pNext );
            }
            pNext = Aig_ObjFanin1(pObj);
            if pNext.pData && !Aig_ObjIsTravIdCurrent(p, pNext) {
                Aig_ObjSetTravIdCurrent(p, pNext);
                Vec_PtrPush( vNodes, pNext );
            }
        }
        // inlining of macro Aig_ObjForEachFanout
        for  assert(p.pFanData, "p.pFanData"), k = 0; (k < int32((pObj).nRefs)) &&                         (((iFan) = k? Aig_ObjFanoutNext(p, iFan) : Aig_ObjFanout0Int(p, pObj.Id)), 1) &&           (((pNext) = Aig_ManObj(p, iFan>>1)), 1); k++  {
            if Saig_ObjIsPo(p, pNext) {
                continue; }
            if Saig_ObjIsLi(p, pNext) {
                pNext = Saig_ObjLiToLo(p, pNext); }
            if pNext.pData && !Aig_ObjIsTravIdCurrent(p, pNext) {
                Aig_ObjSetTravIdCurrent(p, pNext);
                Vec_PtrPush( vNodes, pNext );
            }
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Establishes relationship between nodes using pairing.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_MatchingCountUnmached(p *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsNode(pObj) && !Aig_ObjIsCi(pObj) {
            continue; }
        if pObj.pData != nil {
            continue; }
        Counter++;
    }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Establishes relationship between nodes using pairing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_MatchingExtend(p0 *Aig_Man_t, p1 *Aig_Man_t, nDist int32, fVerbose int32) {
    var vNodes0, vNodes1 *Vec_Ptr_t;
    var pNext0, pNext1 *Aig_Obj_t;
    var d, k int32;
    Aig_ManFanoutStart(p0);
    Aig_ManFanoutStart(p1);
    vNodes0 = Vec_PtrAlloc( 1000 );
    vNodes1 = Vec_PtrAlloc( 1000 );
    if fVerbose != 0 {
        var nUnmached int32  = Ssw_MatchingCountUnmached(p0);
        Abc_Print( 1, "Extending islands by %d steps:\n", nDist );
        Abc_Print( 1, "%2d : Total = %6d. Unmatched = %6d.  Ratio = %6.2f %%\n",
            0, Aig_ManCiNum(p0) + Aig_ManNodeNum(p0),
            nUnmached, 100.0 * nUnmached/(Aig_ManCiNum(p0) + Aig_ManNodeNum(p0)) );
    }
    for  d = 0; d < nDist; d++  {
        Ssw_MatchingExtendOne( p0, vNodes0 );
        Ssw_MatchingExtendOne( p1, vNodes1 );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vNodes0)) && (((pNext0) = (Aig_Obj_t *)Vec_PtrEntry(vNodes0, k)), 1); k++  {
            pNext1 = (Aig_Obj_t *)pNext0.pData;
            if pNext1 == nil {
                continue; }
            assert(pNext1.pData == pNext0, "pNext1.pData == pNext0");
            if Saig_ObjIsPi(p0, pNext1) {
                continue; }
            pNext0.pData = nil;
            pNext1.pData = nil;
        }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vNodes1)) && (((pNext0) = (Aig_Obj_t *)Vec_PtrEntry(vNodes1, k)), 1); k++  {
            pNext1 = (Aig_Obj_t *)pNext0.pData;
            if pNext1 == nil {
                continue; }
            assert(pNext1.pData == pNext0, "pNext1.pData == pNext0");
            if Saig_ObjIsPi(p1, pNext1) {
                continue; }
            pNext0.pData = nil;
            pNext1.pData = nil;
        }
        if fVerbose != 0 {
            var nUnmached int32  = Ssw_MatchingCountUnmached(p0);
            Abc_Print( 1, "%2d : Total = %6d. Unmatched = %6d.  Ratio = %6.2f %%\n",
                d+1, Aig_ManCiNum(p0) + Aig_ManNodeNum(p0),
                nUnmached, 100.0 * nUnmached/(Aig_ManCiNum(p0) + Aig_ManNodeNum(p0)) );
        }
    }
    Vec_PtrFree( vNodes0 );
    Vec_PtrFree( vNodes1 );
    Aig_ManFanoutStop(p0);
    Aig_ManFanoutStop(p1);
}

/**Function*************************************************************

  Synopsis    [Used differences in p0 to complete p1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_MatchingComplete(p0 *Aig_Man_t, p1 *Aig_Man_t) {
    var vNewLis *Vec_Ptr_t;
    var pObj0, pObj0Li, pObj1 *Aig_Obj_t;
    var i int32;
    // create register outputs in p0 that are absent in p1
    vNewLis = Vec_PtrAlloc( 100 );
    // inlining of macro Saig_ManForEachLiLo
    for  i = 0; (i < Saig_ManRegNum(p0)) && (((pObj0Li) = Saig_ManLi(p0, i)), 1)            && (((pObj0)=Saig_ManLo(p0, i)), 1); i++  {
        if pObj0.pData != nil {
            continue; }
        pObj1 = Aig_ObjCreateCi( p1 );
        pObj0.pData = pObj1;
        pObj1.pData = pObj0;
        Vec_PtrPush( vNewLis, pObj0Li );
    }
    // add missing nodes in the topological order
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p0.vObjs)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vObjs, i)), 1); i++  { if (pObj0) == nil || !Aig_ObjIsNode(pObj0) {} else {
        if pObj0.pData != nil {
            continue; }
        pObj1 = Aig_And( p1, Aig_ObjChild0Copy(pObj0), Aig_ObjChild1Copy(pObj0) );
        pObj0.pData = pObj1;
        pObj1.pData = pObj0;
    }
    }
    // create register outputs in p0 that are absent in p1
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNewLis)) && (((pObj0Li) = (Aig_Obj_t *)Vec_PtrEntry(vNewLis, i)), 1); i++  {
        Aig_ObjCreateCo( p1, Aig_ObjChild0Copy(pObj0Li) );
    }
    // increment the number of registers
    Aig_ManSetRegNum( p1, Aig_ManRegNum(p1) + Vec_PtrSize(vNewLis) );
    Vec_PtrFree( vNewLis );
}


/**Function*************************************************************

  Synopsis    [Derives matching for all pairs.]

  Description [Modifies both AIGs.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_MatchingPairs(p0 *Aig_Man_t, p1 *Aig_Man_t) *Vec_Int_t {
    var vPairsNew *Vec_Int_t;
    var pObj0, pObj1 *Aig_Obj_t;
    var i int32;
    // check correctness
    assert(Aig_ManCiNum(p0) == Aig_ManCiNum(p1), "Aig_ManCiNum(p0) == Aig_ManCiNum(p1)");
    assert(Aig_ManCoNum(p0) == Aig_ManCoNum(p1), "Aig_ManCoNum(p0) == Aig_ManCoNum(p1)");
    assert(Aig_ManRegNum(p0) == Aig_ManRegNum(p1), "Aig_ManRegNum(p0) == Aig_ManRegNum(p1)");
    assert(Aig_ManObjNum(p0) == Aig_ManObjNum(p1), "Aig_ManObjNum(p0) == Aig_ManObjNum(p1)");
    // create complete pairs
    vPairsNew = Vec_IntAlloc( 2*Aig_ManObjNum(p0) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p0.vObjs)) && (((pObj0) = (Aig_Obj_t *)Vec_PtrEntry(p0.vObjs, i)), 1); i++  { if (pObj0) == nil {} else {
        if Aig_ObjIsCo(pObj0) {
            continue; }
        pObj1 = (Aig_Obj_t *)pObj0.pData;
        Vec_IntPush( vPairsNew, pObj0.Id );
        Vec_IntPush( vPairsNew, pObj1.Id );
    }
    }
    return vPairsNew;
}





/**Function*************************************************************

  Synopsis    [Transfers the result of matching to miter.]

  Description [The array of pairs should be complete.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_MatchingMiter(pMiter *Aig_Man_t, p0 *Aig_Man_t, p1 *Aig_Man_t, vPairsAll *Vec_Int_t) *Vec_Int_t {
    var vPairsMiter *Vec_Int_t;
    var pObj0, pObj1 *Aig_Obj_t;
    var i int32;
    // create matching of nodes in the miter
    vPairsMiter = Vec_IntAlloc( 2*Aig_ManObjNum(p0) );
    for  i = 0; i < Vec_IntSize(vPairsAll); i += 2  {
        pObj0 = Aig_ManObj( p0, Vec_IntEntry(vPairsAll, i) );
        pObj1 = Aig_ManObj( p1, Vec_IntEntry(vPairsAll, i+1) );
        assert(pObj0.pData != nil, "pObj0.pData != nil");
        assert(pObj1.pData != nil, "pObj1.pData != nil");
        if pObj0.pData == pObj1.pData {
            continue; }
        if Aig_ObjIsNone((Aig_Obj_t *)pObj0.pData) || Aig_ObjIsNone((Aig_Obj_t *)pObj1.pData) {
            continue; }
        // get the miter nodes
        pObj0 = (Aig_Obj_t *)pObj0.pData;
        pObj1 = (Aig_Obj_t *)pObj1.pData;
        assert(!Aig_IsComplement(pObj0), "!Aig_IsComplement(pObj0)");
        assert(!Aig_IsComplement(pObj1), "!Aig_IsComplement(pObj1)");
        assert(Aig_ObjType(pObj0) == Aig_ObjType(pObj1), "Aig_ObjType(pObj0) == Aig_ObjType(pObj1)");
        if Aig_ObjIsCo(pObj0) {
            continue; }
        assert(Aig_ObjIsNode(pObj0) || Saig_ObjIsLo(pMiter, pObj0), "Aig_ObjIsNode(pObj0) || Saig_ObjIsLo(pMiter, pObj0)");
        assert(Aig_ObjIsNode(pObj1) || Saig_ObjIsLo(pMiter, pObj1), "Aig_ObjIsNode(pObj1) || Saig_ObjIsLo(pMiter, pObj1)");
        assert(pObj0.Id < pObj1.Id, "pObj0.Id < pObj1.Id");
        Vec_IntPush( vPairsMiter, pObj0.Id );
        Vec_IntPush( vPairsMiter, pObj1.Id );
    }
    return vPairsMiter;
}





/**Function*************************************************************

  Synopsis    [Solves SEC using structural similarity.]

  Description [Modifies both p0 and p1 by adding extra logic.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SecWithSimilaritySweep(p0 *Aig_Man_t, p1 *Aig_Man_t, vPairs *Vec_Int_t, pPars *Ssw_Pars_t) *Aig_Man_t {
    var p *Ssw_Man_t;
    var vPairsAll, vPairsMiter *Vec_Int_t;
    var pMiter, pAigNew *Aig_Man_t;
    // derive full matching
    Ssw_MatchingStart( p0, p1, vPairs );
    if pPars.nIsleDist {
        Ssw_MatchingExtend( p0, p1, pPars.nIsleDist, pPars.fVerbose ); }
    Ssw_MatchingComplete( p0, p1 );
    Ssw_MatchingComplete( p1, p0 );
    vPairsAll = Ssw_MatchingPairs( p0, p1 );
    // create miter and transfer matching
    pMiter = Saig_ManCreateMiter( p0, p1, 0 );
    vPairsMiter = Ssw_MatchingMiter( pMiter, p0, p1, vPairsAll );
    Vec_IntFree( vPairsAll );
    // start the induction manager
    p = Ssw_ManCreate( pMiter, pPars );
    // create equivalence classes using these IDs
    if p.pPars.fPartSigCorr {
        p.ppClasses = Ssw_ClassesPreparePairsSimple( pMiter, vPairsMiter );
    } else {
        p.ppClasses = Ssw_ClassesPrepare( pMiter, pPars.nFramesK, pPars.fLatchCorr, pPars.fConstCorr, pPars.fOutputCorr, pPars.nMaxLevs, pPars.fVerbose ); }
    if p.pPars.fDumpSRInit {
        if p.pPars.fPartSigCorr {
            var pSRed *Aig_Man_t  = Ssw_SpeculativeReduction( p );
            Aig_ManDumpBlif( pSRed, "srm_part.blif", nil, nil );
            Aig_ManStop( pSRed );
            Abc_Print( 1, "Speculatively reduced miter is saved in file \"%s\".\n", "srm_part.blif" );
        
        } else {
            Abc_Print( 1, "Dumping speculative miter is possible only for partial signal correspondence (switch \"-c\").\n" ); }
    }
    p.pSml = Ssw_SmlStart( pMiter, 0, 1 + p.pPars.nFramesAddSim, 1 );
    Ssw_ClassesSetData( p.ppClasses, p.pSml, (unsigned(*)(void *,Aig_Obj_t *))Ssw_SmlObjHashWord, (int(*)(void *,Aig_Obj_t *))Ssw_SmlObjIsConstWord, (int(*)(void *,Aig_Obj_t *,Aig_Obj_t *))Ssw_SmlObjsAreEqualWord );
    // perform refinement of classes
    pAigNew = Ssw_SignalCorrespondenceRefine( p );
    // cleanup
    Ssw_ManStop( p );
    Aig_ManStop( pMiter );
    Vec_IntFree( vPairsMiter );
    return pAigNew;
}

/**Function*************************************************************

  Synopsis    [Solves SEC with structural similarity.]

  Description [The first two arguments are pointers to the AIG managers.
  The third argument is the array of pairs of IDs of structurally equivalent
  nodes from the first and second managers, respectively.]
               
  SideEffects [The managers will be updated by adding "islands of difference".]

  SeeAlso     []

***********************************************************************/
func Ssw_SecWithSimilarityPairs(p0 *Aig_Man_t, p1 *Aig_Man_t, vPairs *Vec_Int_t, pPars *Ssw_Pars_t) int32 {
    var Pars Ssw_Pars_t;
    var pAigRes *Aig_Man_t;
    var RetValue int32;
    var clk abctime  = Abc_Clock();
    // derive parameters if not given
    if pPars == nil {
        Ssw_ManSetDefaultParams( pPars = &Pars ); }
    // reduce the AIG with pairs
    pAigRes = Ssw_SecWithSimilaritySweep( p0, p1, vPairs, pPars );
    // report the result of verification
    RetValue = Ssw_MiterStatus( pAigRes, 1 );
    if RetValue == 1 {
        Abc_Print( 1, "Verification successful.  " );
    } else if RetValue == 0 {
        Abc_Print( 1, "Verification failed with a counter-example.  " );
    } else {
        Abc_Print( 1, "Verification UNDECIDED. The number of remaining regs = %d (total = %d).  ",
            Aig_ManRegNum(pAigRes), Aig_ManRegNum(p0)+Aig_ManRegNum(p1) ); }
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    Aig_ManStop( pAigRes );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Dummy procedure to detect structural similarity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_StrSimPerformMatching_hack(p0 *Aig_Man_t, p1 *Aig_Man_t) *Vec_Int_t {
    var vPairs *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // create array of pairs
    vPairs = Vec_IntAlloc( 100 );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p0.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p0.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsConst1(pObj) && !Aig_ObjIsCi(pObj) && !Aig_ObjIsNode(pObj) {
            continue; }
        Vec_IntPush( vPairs, i );
        Vec_IntPush( vPairs, i );
    }
    }
    return vPairs;
}

/**Function*************************************************************

  Synopsis    [Solves SEC with structural similarity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SecWithSimilarity(p0 *Aig_Man_t, p1 *Aig_Man_t, pPars *Ssw_Pars_t) int32 {
    var vPairs *Vec_Int_t;
    var pPart0, pPart1 *Aig_Man_t;
    var RetValue int32;
    if pPars.fVerbose {
        Abc_Print( 1, "Performing sequential verification using structural similarity.\n" ); }
    // consider the case when a miter is given
    if p1 == nil {
        if pPars.fVerbose {
            Aig_ManPrintStats( p0 );
        }
        // demiter the miter
        if !Saig_ManDemiterSimpleDiff( p0, &pPart0, &pPart1 ) {
            Abc_Print( 1, "Demitering has failed.\n" );
            return -1;
        }
    
    } else {
        pPart0 = Aig_ManDupSimple( p0 );
        pPart1 = Aig_ManDupSimple( p1 );
    }
    if pPars.fVerbose {
//        Aig_ManPrintStats( pPart0 );
//        Aig_ManPrintStats( pPart1 );
        if p1 == nil {
//        Aig_ManDumpBlif( pPart0, "part0.blif", NULL, NULL );
//        Aig_ManDumpBlif( pPart1, "part1.blif", NULL, NULL );
//        Abc_Print( 1, "The result of demitering is written into files \"%s\" and \"%s\".\n", "part0.blif", "part1.blif" );
        }
    }
    assert(Aig_ManRegNum(pPart0) > 0, "Aig_ManRegNum(pPart0) > 0");
    assert(Aig_ManRegNum(pPart1) > 0, "Aig_ManRegNum(pPart1) > 0");
    assert(Saig_ManPiNum(pPart0) == Saig_ManPiNum(pPart1), "Saig_ManPiNum(pPart0) == Saig_ManPiNum(pPart1)");
    assert(Saig_ManPoNum(pPart0) == Saig_ManPoNum(pPart1), "Saig_ManPoNum(pPart0) == Saig_ManPoNum(pPart1)");
    // derive pairs
//    vPairs = Saig_StrSimPerformMatching_hack( pPart0, pPart1 );
    vPairs = Saig_StrSimPerformMatching( pPart0, pPart1, 0, pPars.fVerbose, nil );
    RetValue = Ssw_SecWithSimilarityPairs( pPart0, pPart1, vPairs, pPars );
    Aig_ManStop( pPart0 );
    Aig_ManStop( pPart1 );
    Vec_IntFree( vPairs );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

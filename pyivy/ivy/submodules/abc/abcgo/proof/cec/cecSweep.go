package cec

/**CFile****************************************************************

  FileName    [cecSweep.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [SAT sweeping manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSweep.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs limited speculative reduction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManFraSpecReduction(p *Cec_ManFra_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var pRepr *Gia_Obj_t = nil;
    var iRes0, iRes1, iRepr, iNode, iMiter int32;
    var i int32;
    var fCompl int32;
    var piCopies *int32;
    var pDepths *int32;
    Gia_ManSetPhase( p.pAig );
    Vec_IntClear( p.vXorNodes );
    if p.pPars.nLevelMax {
        Gia_ManLevelNum( p.pAig ); }
    pNew = Gia_ManStart( Gia_ManObjNum(p.pAig) );
    pNew.pName = Abc_UtilStrsav( p.pAig.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pAig.pName );
    Gia_ManHashAlloc( pNew );
    // inlining of macro ABC_FALLOC
    piCopies = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p.pAig))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p.pAig))));
    // inlining of macro ABC_CALLOC
    pDepths  = ((int *) calloc((size_t)(Gia_ManObjNum(p.pAig)), sizeof(int)));
    piCopies[0] = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.pAig.nObjs) && ((pObj) = Gia_ManObj(p.pAig, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            piCopies[i] = Gia_ManAppendCi( pNew );
            continue;
        }
        if Gia_ObjIsCo(pObj) { 
            continue; }
        if piCopies[Gia_ObjFaninId0(pObj,i)] == -1 ||
             piCopies[Gia_ObjFaninId1(pObj,i)] == -1 {
             continue; }
        iRes0 = Abc_LitNotCond( piCopies[Gia_ObjFaninId0(pObj,i)], Gia_ObjFaninC0(pObj) );
        iRes1 = Abc_LitNotCond( piCopies[Gia_ObjFaninId1(pObj,i)], Gia_ObjFaninC1(pObj) );
        iNode = piCopies[i] = Gia_ManHashAnd( pNew, iRes0, iRes1 );
        pDepths[i] = Abc_MaxInt( pDepths[Gia_ObjFaninId0(pObj,i)], pDepths[Gia_ObjFaninId1(pObj,i)] );
        if Gia_ObjRepr(p.pAig, i) == GIA_VOID || Gia_ObjFailed(p.pAig, i) {
            continue; }
        assert(Gia_ObjRepr(p.pAig, i) < i, "Gia_ObjRepr(p.pAig, i) < i");
        iRepr = piCopies[Gia_ObjRepr(p.pAig, i)];
        if iRepr == -1 {
            continue; }
        if Abc_LitRegular(iNode) == Abc_LitRegular(iRepr) {
            continue; }
        if p.pPars.nLevelMax && 
            (Gia_ObjLevelId(p.pAig, i)  > p.pPars.nLevelMax || 
             Gia_ObjLevelId(p.pAig, Abc_Lit2Var(iRepr)) > p.pPars.nLevelMax) {
            continue; }
        if p.pPars.fDualOut {
//            if ( i % 1000 == 0 && Gia_ObjRepr(p->pAig, i) )
//                Gia_ManEquivPrintOne( p->pAig, Gia_ObjRepr(p->pAig, i), 0 );
            if p.pPars.fColorDiff {
                if !Gia_ObjDiffColors( p.pAig, Gia_ObjRepr(p.pAig, i), i ) {
                    continue; }
            
            } else {
                if !Gia_ObjDiffColors2( p.pAig, Gia_ObjRepr(p.pAig, i), i ) {
                    continue; }
            }
        }
        pRepr = Gia_ManObj( p.pAig, Gia_ObjRepr(p.pAig, i) );
        fCompl = Gia_ObjPhaseReal(pObj) ^ Gia_ObjPhaseReal(pRepr);
        piCopies[i] = Abc_LitNotCond( iRepr, fCompl );
        if Gia_ObjProved(p.pAig, i) {
            continue; }
        // produce speculative miter
        iMiter = Gia_ManHashXor( pNew, iNode, piCopies[i] );
        Gia_ManAppendCo( pNew, iMiter );
        Vec_IntPush( p.vXorNodes, Gia_ObjRepr(p.pAig, i) );
        Vec_IntPush( p.vXorNodes, i );
        // add to the depth of this node
        pDepths[i] = 1 + Abc_MaxInt( pDepths[i], pDepths[Gia_ObjRepr(p.pAig, i)] );
        if p.pPars.nDepthMax && pDepths[i] >= p.pPars.nDepthMax {
            piCopies[i] = -1; }
    }
    piCopies = nil;
    pDepths = nil;
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, 0 );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManFraClassesUpdate_rec(pObj *Gia_Obj_t) int32 {
    var Result int32;
    if pObj.fMark0 {
        return 1; }
    if Gia_ObjIsCi(pObj) || Gia_ObjIsConst0(pObj) {
        return 0; }
    Result = (Cec_ManFraClassesUpdate_rec( Gia_ObjFanin0(pObj) ) |
              Cec_ManFraClassesUpdate_rec( Gia_ObjFanin1(pObj) ));
    return pObj.fMark0 = Result;
}

/**Function*************************************************************

  Synopsis    [Creates simulation info for this round.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManFraCreateInfo(p *Cec_ManSim_t, vCiInfo *Vec_Ptr_t, vInfo *Vec_Ptr_t, nSeries int32) {
    var pRes0, pRes1 *uint32;
    var i, w int32;
    for  i = 0; i < Gia_ManCiNum(p.pAig); i++  {
        pRes0 = (unsigned *)Vec_PtrEntry( vCiInfo, i );
        pRes1 = (unsigned *)Vec_PtrEntry( vInfo, i );
        pRes1 += p.nWords * nSeries;
        for  w = 0; w < p.nWords; w++  {
            pRes0[w] = pRes1[w];
        }
    }
}

/**Function*************************************************************

  Synopsis    [Updates equivalence classes using the patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManFraClassesUpdate(p *Cec_ManFra_t, pSim *Cec_ManSim_t, pPat *Cec_ManPat_t, pNew *Gia_Man_t) int32 {
    var vInfo *Vec_Ptr_t;
    var pObj, pObjOld, pReprOld *Gia_Obj_t;
    var i, k, iRepr, iNode int32;
    var clk abctime;
clk = Abc_Clock();
    vInfo = Cec_ManPatCollectPatterns( pPat, Gia_ManCiNum(p.pAig), pSim.nWords );
p.timePat += Abc_Clock() - clk;
clk = Abc_Clock();
    if vInfo != nil {
        Gia_ManCreateValueRefs( p.pAig );
        for  i = 0; i < pPat.nSeries; i++  {
            Cec_ManFraCreateInfo( pSim, pSim.vCiSimInfo, vInfo, i );
            if Cec_ManSimSimulateRound( pSim, pSim.vCiSimInfo, pSim.vCoSimInfo ) {
                Vec_PtrFree( vInfo );
                return 1;
            }
        }
        Vec_PtrFree( vInfo );
    }
p.timeSim += Abc_Clock() - clk;
    assert(Vec_IntSize(p.vXorNodes) == 2*Gia_ManCoNum(pNew), "Vec_IntSize(p.vXorNodes) == 2*Gia_ManCoNum(pNew)");
    // mark the transitive fanout of failed nodes
    if p.pPars.nDepthMax != 1 {
        Gia_ManCleanMark0( p.pAig );
        Gia_ManCleanMark1( p.pAig );
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(pNew.vCos)) && ((pObj) = Gia_ManCo(pNew, k)); k++  {
            iRepr = Vec_IntEntry( p.vXorNodes, 2*k );
            iNode = Vec_IntEntry( p.vXorNodes, 2*k+1 );
            if pObj.fMark0 == 0 && pObj.fMark1 == 1 { // proved
                continue; }
//            Gia_ManObj(p->pAig, iRepr)->fMark0 = 1;
            Gia_ManObj(p.pAig, iNode).fMark0 = 1;
        }
        // mark the nodes reachable through the failed nodes
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.pAig.nObjs) && ((pObjOld) = Gia_ManObj(p.pAig, k)); k++  {      if !Gia_ObjIsAnd(pObjOld) {} else {
            pObjOld.fMark0 |= (Gia_ObjFanin0(pObjOld).fMark0 | Gia_ObjFanin1(pObjOld).fMark0); }
        }
        // unmark the disproved nodes
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(pNew.vCos)) && ((pObj) = Gia_ManCo(pNew, k)); k++  {
            iRepr = Vec_IntEntry( p.vXorNodes, 2*k );
            iNode = Vec_IntEntry( p.vXorNodes, 2*k+1 );
            if pObj.fMark0 == 0 && pObj.fMark1 == 1 { // proved
                continue; } 
            pObjOld = Gia_ManObj(p.pAig, iNode);
            assert(pObjOld.fMark0 == 1, "pObjOld.fMark0 == 1");
            if Gia_ObjFanin0(pObjOld).fMark0 == 0 && Gia_ObjFanin1(pObjOld).fMark0 == 0 {
                pObjOld.fMark1 = 1; }
        }
        // clean marks
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.pAig.nObjs) && ((pObjOld) = Gia_ManObj(p.pAig, k)); k++  {      if !Gia_ObjIsAnd(pObjOld) {} else
            if pObjOld.fMark1 {
                pObjOld.fMark0 = 0;
                pObjOld.fMark1 = 0;
            }
        }
    }
    // set the results
    p.nAllProved = p.nAllDisproved = p.nAllFailed = 0;
    // inlining of macro Gia_ManForEachCo
    for  k = 0; (k < Vec_IntSize(pNew.vCos)) && ((pObj) = Gia_ManCo(pNew, k)); k++  {
        iRepr = Vec_IntEntry( p.vXorNodes, 2*k );
        iNode = Vec_IntEntry( p.vXorNodes, 2*k+1 );
        pReprOld = Gia_ManObj(p.pAig, iRepr);
        pObjOld = Gia_ManObj(p.pAig, iNode);
        if pObj.fMark1 { // proved
            assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
            assert(!Gia_ObjProved(p.pAig, iNode), "!Gia_ObjProved(p.pAig, iNode)");
            if pReprOld.fMark0 == 0 && pObjOld.fMark0 == 0 {
//            if ( pObjOld->fMark0 == 0 )
            
                assert(iRepr == Gia_ObjRepr(p.pAig, iNode), "iRepr == Gia_ObjRepr(p.pAig, iNode)");
                Gia_ObjSetProved( p.pAig, iNode );
                p.nAllProved++;
            }
        
        } else if pObj.fMark0 { // disproved
            assert(pObj.fMark1 == 0, "pObj.fMark1 == 0");
            if pReprOld.fMark0 == 0 && pObjOld.fMark0 == 0 {
//            if ( pObjOld->fMark0 == 0 )
            
                if iRepr == Gia_ObjRepr(p.pAig, iNode) {
                    Abc_Print( 1, "Cec_ManFraClassesUpdate(): Error! Node is not refined!\n" ); }
                p.nAllDisproved++;
            }
        
        } else { // failed
            assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
            assert(pObj.fMark1 == 0, "pObj.fMark1 == 0");
            assert(!Gia_ObjFailed(p.pAig, iNode), "!Gia_ObjFailed(p.pAig, iNode)");
            assert(!Gia_ObjProved(p.pAig, iNode), "!Gia_ObjProved(p.pAig, iNode)");
            Gia_ObjSetFailed( p.pAig, iNode );
            p.nAllFailed++;
        }
    } 
    p.nAllProvedS    += p.nAllProved;
    p.nAllDisprovedS += p.nAllDisproved;
    p.nAllFailedS    += p.nAllFailed;
    return 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


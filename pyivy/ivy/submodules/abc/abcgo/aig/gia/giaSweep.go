package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaSweep.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Sweeping of GIA manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSweep.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "giaAig.h"
//#include "proof/dch/dch.h"
//#include "misc/tim/tim.h"
//#include "proof/cec/cec.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Mark GIA nodes that feed into POs.]

  Description [Returns the array of classes of remaining registers.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupWithBoxes(p *Gia_Man_t, fSeq int32) *Gia_Man_t {
    var pManTime *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vBoxesLeft *Vec_Int_t;
    var curCi, curCo, nBoxIns, nBoxOuts int32;
    var i, k, iShift, nMarked int32;
    assert(Gia_ManBoxNum(p) > 0, "Gia_ManBoxNum(p) > 0");
    // mark useful boxes
    Gia_ManMarkSeqGiaWithBoxes( p, fSeq );
    // duplicate marked entries
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if !Gia_ObjIsTravIdCurrent(p, pObj) {
            continue; }
        if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi(pNew);
        } else if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else { assert(false, "0"); }
    }
    assert(!Gia_ManHasDangling(pNew), "!Gia_ManHasDangling(pNew)");
    // collect remaining flops
    if fSeq != 0 {
        pNew.vRegClasses = Vec_IntAlloc( Gia_ManRegBoxNum(p) );
        if p.vRegInits {
            pNew.vRegInits = Vec_IntAlloc( Gia_ManRegBoxNum(p) ); }
        iShift = Gia_ManPoNum(p) - Gia_ManRegBoxNum(p);
        for  i = 0; i < Gia_ManRegBoxNum(p); i++  {
            if Gia_ObjIsTravIdCurrent(p, Gia_ManCo(p, iShift + i)) {
                Vec_IntPush( pNew.vRegClasses, Vec_IntEntry(p.vRegClasses, i) );
                if p.vRegInits {
                    Vec_IntPush( pNew.vRegInits, Vec_IntEntry(p.vRegInits, i) ); }
            }
        }
    
    } else {
        if p.vRegClasses {
            pNew.vRegClasses = Vec_IntDup( p.vRegClasses ); }
        if p.vRegInits {
            pNew.vRegInits = Vec_IntDup( p.vRegInits ); }
    }
    // collect remaining boxes
    vBoxesLeft = Vec_IntAlloc( Gia_ManBoxNum(p) );
    curCi = Tim_ManPiNum(pManTime);
    curCo = 0;
    for  i = 0; i < Gia_ManBoxNum(p); i++  {
        nBoxIns = Tim_ManBoxInputNum(pManTime, i);
        nBoxOuts = Tim_ManBoxOutputNum(pManTime, i);
        nMarked = 0;
        for  k = 0; k < nBoxIns; k++  {
            nMarked += Gia_ObjIsTravIdCurrent( p, Gia_ManCo(p, curCo + k) );
        }
        for  k = 0; k < nBoxOuts; k++  {
            nMarked += Gia_ObjIsTravIdCurrent( p, Gia_ManCi(p, curCi + k) );
        }
        curCo += nBoxIns;
        curCi += nBoxOuts;
        // check presence
        assert(nMarked == 0 || nMarked == nBoxIns + nBoxOuts, "nMarked == 0 || nMarked == nBoxIns + nBoxOuts");
        if nMarked != 0 {
            Vec_IntPush( vBoxesLeft, i ); }
    }
    curCo += Tim_ManPoNum(pManTime);
    assert(curCi == Gia_ManCiNum(p), "curCi == Gia_ManCiNum(p)");
    assert(curCo == Gia_ManCoNum(p), "curCo == Gia_ManCoNum(p)");
    // update timing manager
    pNew.pManTime = Gia_ManUpdateTimMan2( p, vBoxesLeft, Gia_ManRegBoxNum(p) - Gia_ManRegBoxNum(pNew) );
    // update extra STG
    assert(p.pAigExtra != nil, "p.pAigExtra != nil");
    assert(pNew.pAigExtra == nil, "pNew.pAigExtra == nil");
    pNew.pAigExtra = Gia_ManUpdateExtraAig2( p.pManTime, p.pAigExtra, vBoxesLeft );
    assert(Gia_ManCiNum(pNew) == Tim_ManPiNum((Tim_Man_t*)pNew.pManTime) + Gia_ManCoNum(pNew.pAigExtra), "Gia_ManCiNum(pNew) == Tim_ManPiNum((Tim_Man_t*)pNew.pManTime) + Gia_ManCoNum(pNew.pAigExtra)");
    Vec_IntFree( vBoxesLeft );
    pNew.nAnd2Delay = p.nAnd2Delay;
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Mark GIA nodes that feed into POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFraigCheckCis(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    for  assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)"); Gia_ObjIsCi(pObj); pObj--  {
        if Gia_ObjIsTravIdCurrent(p, pObj) {
            return 1; }
    }
    return 0;
}
func Gia_ManFraigMarkCis(p *Gia_Man_t, pObj *Gia_Obj_t, fMark int32) *Gia_Obj_t {
    for  assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)"); Gia_ObjIsCi(pObj); pObj--  {
        if fMark != 0 {
            Gia_ObjSetTravIdCurrent( p, pObj ); }
    }
    return pObj;
}
func Gia_ManFraigMarkCos(p *Gia_Man_t, pObj *Gia_Obj_t, fMark int32) *Gia_Obj_t {
    for  assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)"); Gia_ObjIsCo(pObj); pObj--  {
        if fMark != 0 {
            Gia_ObjSetTravIdCurrent( p, pObj );
            Gia_ObjSetTravIdCurrent( p, Gia_ObjFanin0(pObj) );
        }
    }
    return pObj;
}
func Gia_ManFraigMarkAnd(p *Gia_Man_t, pObj *Gia_Obj_t) *Gia_Obj_t {
    for  assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)"); Gia_ObjIsAnd(pObj); pObj--  {
        if Gia_ObjIsTravIdCurrent(p, pObj) {
            Gia_ObjSetTravIdCurrent( p, Gia_ObjFanin0(pObj) );
            Gia_ObjSetTravIdCurrent( p, Gia_ObjFanin1(pObj) );
        }
    }
    return pObj;
}
func Gia_ManFraigCreateGia(p *Gia_Man_t) *Gia_Man_t {
    var vBoxPres *Vec_Int_t;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, fLabelPos int32;
    assert(p.pManTime != nil, "p.pManTime != nil");
    // start marks
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    vBoxPres = Vec_IntAlloc( 1000 );
    // mark primary outputs
    fLabelPos = 1;
    pObj = Gia_ManObj( p, Gia_ManObjNum(p) - 1 );
    assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
    for  Gia_ObjIsCo(pObj)  {
        pObj = Gia_ManFraigMarkCos( p, pObj, fLabelPos );
        if Gia_ObjIsAnd(pObj) {
            pObj = Gia_ManFraigMarkAnd( p, pObj ); }
        assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)");
        fLabelPos = Gia_ManFraigCheckCis(p, pObj);
        pObj = Gia_ManFraigMarkCis( p, pObj, fLabelPos );
        Vec_IntPush( vBoxPres, fLabelPos );
    }
    Vec_IntPop( vBoxPres );
    Vec_IntReverseOrder( vBoxPres );
    assert(Gia_ObjIsConst0(pObj), "Gia_ObjIsConst0(pObj)");
    // mark primary inputs
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            Gia_ObjSetTravIdCurrent( p, pObj );
        } else {
            break; }
    }
    // duplicate marked entries
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if !Gia_ObjIsTravIdCurrent(p, pObj) {
            continue; }
        if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi(pNew);
        } else if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else { assert(false, "0"); }
    }
    // update timing manager
    pNew.pManTime = Gia_ManUpdateTimMan( p, vBoxPres );
    // update extra STG
    assert(p.pAigExtra != nil, "p.pAigExtra != nil");
    assert(pNew.pAigExtra == nil, "pNew.pAigExtra == nil");
    pNew.pAigExtra = Gia_ManUpdateExtraAig( p.pManTime, p.pAigExtra, vBoxPres );
    Vec_IntFree( vBoxPres );
//    assert( Gia_ManPiNum(pNew) == Tim_ManCiNum(pNew->pManTime) );
//    assert( Gia_ManPoNum(pNew) == Tim_ManCoNum(pNew->pManTime) );
//    assert( Gia_ManPiNum(pNew) == Tim_ManPiNum(pNew->pManTime) + Gia_ManPoNum(pNew->pAigExtra) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjFanin0CopyRepr(p *Gia_Man_t, pObj *Gia_Obj_t, pReprs *int32) int32 {
    var fanId int32  = Gia_ObjFaninId0p( p, pObj );
    if pReprs[fanId] == -1 {
        return Gia_ObjFanin0Copy( pObj ); }
    assert(Abc_Lit2Var(pReprs[fanId]) < Gia_ObjId(p, pObj), "Abc_Lit2Var(pReprs[fanId]) < Gia_ObjId(p, pObj)");
    return Abc_LitNotCond( Gia_ObjValue(Gia_ManObj(p, Abc_Lit2Var(pReprs[fanId]))), Gia_ObjFaninC0(pObj) ^ Abc_LitIsCompl(pReprs[fanId]) );
}
func Gia_ObjFanin1CopyRepr(p *Gia_Man_t, pObj *Gia_Obj_t, pReprs *int32) int32 {
    var fanId int32  = Gia_ObjFaninId1p( p, pObj );
    if pReprs[fanId] == -1 {
        return Gia_ObjFanin1Copy( pObj ); }
    assert(Abc_Lit2Var(pReprs[fanId]) < Gia_ObjId(p, pObj), "Abc_Lit2Var(pReprs[fanId]) < Gia_ObjId(p, pObj)");
    return Abc_LitNotCond( Gia_ObjValue(Gia_ManObj(p, Abc_Lit2Var(pReprs[fanId]))), Gia_ObjFaninC1(pObj) ^ Abc_LitIsCompl(pReprs[fanId]) );
}
func Gia_ManFraigReduceGia(p *Gia_Man_t, pReprs *int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(pReprs != nil, "pReprs != nil");
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0CopyRepr(p, pObj, pReprs), Gia_ObjFanin1CopyRepr(p, pObj, pReprs) );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0CopyRepr(p, pObj, pReprs) );
        } else if Gia_ObjIsConst0(pObj) {
            pObj.Value = 0;
        } else { assert(false, "0"); }
    }
    Gia_ManHashStop( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Compute the set of CIs representing carry-outs of boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Checks integriting of complex flops and carry-chains.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Computes representatives in terms of the original objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFraigSelectReprs(p *Gia_Man_t, pClp *Gia_Man_t, fVerbose int32, pFlopTypes []int32) *int32 {
    var pObj *Gia_Obj_t;
    var vCarryOuts *Vec_Int_t;
    var pManTime *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
    // inlining of macro ABC_FALLOC
    var pReprs *int32    = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p))));
    // inlining of macro ABC_FALLOC
    var pClp2Gia *int32  = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(pClp))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(pClp))));
    var i, iLitClp, iLitClp2, iReprClp, fCompl int32;
    var nConsts, nReprs int32;
    assert(pManTime != nil, "pManTime != nil");
    // count the number of equivalent objects
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < pClp.nObjs) && ((pObj) = Gia_ManObj(pClp, i)); i++  {
        if Gia_ObjIsCo(pObj) {
            continue; }
        if i == Gia_ObjReprSelf(pClp, i) {
            continue; }
        if Gia_ObjReprSelf(pClp, i) == 0 {
            nConsts++;
        } else {
            nReprs++; }
    }
    if fVerbose != 0 {
        fmt.Printf("Computed %d const objects and %d other objects.\n", nConsts, nReprs); }
    nConsts = nReprs = 0;

    // mark flop input boxes
    Gia_ManCleanMark0( p );
    for  i = Gia_ManPoNum(p) - Gia_ManRegBoxNum(p); i < Gia_ManPoNum(p); i++  {
        pObj = Gia_ObjFanin0( Gia_ManPo(p, i) );
        assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)");
        pObj.fMark0 = 1;
    }
    // mark connects between last box inputs and first box outputs
    vCarryOuts = Gia_ManComputeCarryOuts( p );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCarryOuts)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vCarryOuts,i))); i++  {
        pObj.fMark0 = 1;
    }
    if fVerbose != 0 {
        fmt.Printf("Fixed %d flop inputs and %d box/box connections (out of %d non-flop boxes).\n", Gia_ManRegBoxNum(p), Vec_IntSize(vCarryOuts), Gia_ManNonRegBoxNum(p)); }
    Vec_IntFree( vCarryOuts );

    // collect equivalent node info
    pFlopTypes[0] = pFlopTypes[1] = pFlopTypes[2] = 0;
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(pClp)) && ((pObj) = Gia_ManCi(pClp, Gia_ManPiNum(pClp)+i)); i++  {
        var pRepr *Gia_Obj_t  = Gia_ObjReprObj(pClp, i);
        if (pRepr != nil) && pRepr != pObj {
            if pRepr == Gia_ManConst0(pClp) {
                pFlopTypes[0]++;
            } else if Gia_ObjIsRo(pClp, pRepr) {
                pFlopTypes[1]++; }
        }
    }

    // compute representatives
    pClp2Gia[0] = 0;
    Gia_ManSetPhase( pClp );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCo(pObj) {
            continue; }
        if Gia_ObjIsCi(pObj) && pObj.fMark0 { // skip CI pointed by CO
            continue; }
        assert(Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj), "Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj)");
        iLitClp = Gia_ObjValue(pObj);
        if iLitClp == -1 {
            continue; }
        iReprClp = Gia_ObjReprSelf( pClp, Abc_Lit2Var(iLitClp) );
        if pClp2Gia[iReprClp] == -1 {
            pClp2Gia[iReprClp] = i;
        } else { 
            iLitClp2 = Gia_ObjValue( Gia_ManObj(p, pClp2Gia[iReprClp]) );
            assert(Gia_ObjReprSelf(pClp, Abc_Lit2Var(iLitClp)) == Gia_ObjReprSelf(pClp, Abc_Lit2Var(iLitClp2)), "Gia_ObjReprSelf(pClp, Abc_Lit2Var(iLitClp)) == Gia_ObjReprSelf(pClp, Abc_Lit2Var(iLitClp2))");
            fCompl  = Abc_LitIsCompl(iLitClp) ^ Abc_LitIsCompl(iLitClp2);
            fCompl ^= Gia_ManObj(pClp, Abc_Lit2Var(iLitClp)).fPhase;
            fCompl ^= Gia_ManObj(pClp, Abc_Lit2Var(iLitClp2)).fPhase;
            pReprs[i] = Abc_Var2Lit( pClp2Gia[iReprClp], fCompl );
            assert(Abc_Lit2Var(pReprs[i]) < i, "Abc_Lit2Var(pReprs[i]) < i");
            if pClp2Gia[iReprClp] == 0 {
                nConsts++;
            } else {
                nReprs++; }
        }
    }
    // inlining of macro ABC_FREE
    pClp2Gia = nil;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.fMark0 = 0;
    }
    if fVerbose != 0 {
        fmt.Printf("Found %d const objects and %d other objects.\n", nConsts, nReprs); }
    return pReprs;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFraigSweepPerform(p *Gia_Man_t, pPars any) {
    var pNew *Aig_Man_t;
    pNew = Gia_ManToAigSimple( p );
    assert(Gia_ManObjNum(p) == Aig_ManObjNum(pNew), "Gia_ManObjNum(p) == Aig_ManObjNum(pNew)");
    Dch_ComputeEquivalences( pNew, (Dch_Pars_t *)pPars );
    Gia_ManReprFromAigRepr( pNew, p );
    Aig_ManStop( pNew );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFraigSweepSimple(p *Gia_Man_t, pPars any) *Gia_Man_t {
    var pNew *Gia_Man_t;
    assert(p.pManTime == nil || Gia_ManBoxNum(p) == 0, "p.pManTime == nil || Gia_ManBoxNum(p) == 0");
    Gia_ManFraigSweepPerform( p, pPars );
    pNew = Gia_ManEquivReduce( p, 1, 0, 0, 0 );
    if pNew == nil {
        pNew = Gia_ManDup(p); }
    Gia_ManTransferTiming( pNew, p );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Computes equivalences for one clock domain.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSweepComputeOneDomainEquivs(p *Gia_Man_t, vRegClasses *Vec_Int_t, iDom int32, pParsS any, fConst int32, fEquiv int32, fVerbose int32) {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vPerm *Vec_Int_t;
    var i, Class, nFlops int32;
    var nDoms int32  = Vec_IntFindMax(vRegClasses);
    assert(iDom >= 1 && iDom <= nDoms, "iDom >= 1 && iDom <= nDoms");
    assert(p.pManTime == nil, "p.pManTime == nil");
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
    // create required flop permutation
    vPerm = Vec_IntAlloc( Gia_ManRegNum(p) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRegClasses)) && (((Class) = Vec_IntEntry(vRegClasses, i)), 1); i++  {
        if Class != iDom {
            Vec_IntPush( vPerm, i ); }
    }
    nFlops = Vec_IntSize( vPerm );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRegClasses)) && (((Class) = Vec_IntEntry(vRegClasses, i)), 1); i++  {
        if Class == iDom {
            Vec_IntPush( vPerm, i ); }
    }
    nFlops = Vec_IntSize(vPerm) - nFlops;
    assert(Vec_IntSize(vPerm) == Gia_ManRegNum(p), "Vec_IntSize(vPerm) == Gia_ManRegNum(p)");
    // derive new AIG
    pNew = Gia_ManDupPermFlop( p, vPerm );
    assert(Gia_ManObjNum(pNew) == Gia_ManObjNum(p), "Gia_ManObjNum(pNew) == Gia_ManObjNum(p)");
    Vec_IntFree( vPerm );
    // perform computation of equivalences 
    pNew.nRegs = nFlops;
    if pParsS {
        Cec_ManLSCorrespondenceClasses( pNew, (Cec_ParCor_t *)pParsS );
    } else { 
        Gia_ManSeqCleanupClasses( pNew, fConst, fEquiv, fVerbose ); }
    pNew.nRegs = Gia_ManRegNum(p);
    // make new point to old
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        assert(!Abc_LitIsCompl(pObj.Value), "!Abc_LitIsCompl(pObj.Value)");
        Gia_ManObj(pNew, Abc_Lit2Var(pObj.Value)).Value = Abc_Var2Lit(i, 0);
    }
    // transfer
    Gia_ManDupRemapEquiv( p, pNew );
    Gia_ManStop( pNew );
}
func Gia_ManSweepWithBoxesAndDomains(p *Gia_Man_t, pParsS any, fConst int32, fEquiv int32, fVerbose int32, fVerbEquivs int32) *Gia_Man_t { 
    var pClp, pNew, pTemp *Gia_Man_t;
    var nDoms int32  = Vec_IntFindMax(p.vRegClasses);
    var pReprs *int32;
    var iDom int32;
    var pFlopTypes [3]int32 ;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(p.pAigExtra != nil, "p.pAigExtra != nil");
    assert(nDoms > 1, "nDoms > 1");
    // order AIG objects
    pNew = Gia_ManDupUnnormalize( p );
    if pNew == nil {
        return nil; }
    Gia_ManTransferTiming( pNew, p );
    // iterate over domains
    for  iDom = 1; iDom <= nDoms; iDom++  {
        var nFlopsNew int32;
        var nFlops int32 = Vec_IntCountEntry(pNew.vRegClasses, iDom);
        if nFlops < 2 {
            continue; }
        // find global equivalences
        pClp = Gia_ManDupCollapse( pNew, pNew.pAigExtra, nil, 1 );
        //Gia_DumpAiger( pClp, p->pSpec, iDom, 2 );
        //Gia_ManPrintStats( pClp, NULL );
        // compute equivalences
        Gia_ManSweepComputeOneDomainEquivs( pClp, pNew.vRegClasses, iDom, pParsS, fConst, fEquiv, fVerbose );
        // transfer equivalences
        pReprs = Gia_ManFraigSelectReprs( pNew, pClp, fVerbose, pFlopTypes );
        Gia_ManStop( pClp );
        // reduce AIG
        Gia_ManTransferTiming( p, pNew );
        pNew = Gia_ManFraigReduceGia( pTemp = pNew, pReprs );
        Gia_ManTransferTiming( pNew, p );
        Gia_ManStop( pTemp );
        // inlining of macro ABC_FREE
        pReprs = nil;
        // derive new AIG
        pNew = Gia_ManDupWithBoxes( pTemp = pNew, 1 );
        Gia_ManStop( pTemp );
        // report
        nFlopsNew = Vec_IntCountEntry(pNew.vRegClasses, iDom);
        pFlopTypes[2] = nFlops - nFlopsNew - (pFlopTypes[0] + pFlopTypes[1]);
        if fVerbEquivs != 0 {
            fmt.Printf("Domain %2d : %5d -> %5d :  ", iDom, nFlops, nFlopsNew);
            fmt.Printf("EqConst =%4d.  EqFlop =%4d.  Dangling =%4d.  Unused =%4d.\n", pFlopTypes[0], pFlopTypes[1], Abc_MaxInt(0, pFlopTypes[2]), Abc_MaxInt(0, -pFlopTypes[2]));
            //Gia_ManPrintStats( pNew, NULL );
        }
    }
    // normalize the result
    pNew = Gia_ManDupNormalize( pTemp = pNew, 0 );
    Gia_ManTransferTiming( pNew, pTemp );
    Gia_ManStop( pTemp );
    // check integrity
    //Gia_ManCheckIntegrityWithBoxes( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Reduces root model with scorr.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSweepWithBoxes(p *Gia_Man_t, pParsC any, pParsS any, fConst int32, fEquiv int32, fVerbose int32, fVerbEquivs int32) *Gia_Man_t { 
    var pClp, pNew, pTemp *Gia_Man_t;
    var pReprs *int32;
    var pFlopTypes [3]int32 ;
    var nFlopsNew, nFlops int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(p.pAigExtra != nil, "p.pAigExtra != nil");
    // consider seq synthesis with multiple clock domains
    if pParsC == nil && Gia_ManClockDomainNum(p) > 1 {
        return Gia_ManSweepWithBoxesAndDomains( p, pParsS, fConst, fEquiv, fVerbose, fVerbEquivs ); }
    // order AIG objects
    pNew = Gia_ManDupUnnormalize( p );
    if pNew == nil {
        return nil; }
    Gia_ManTransferTiming( pNew, p );
    nFlops = Vec_IntCountEntry(pNew.vRegClasses, 1);
    // find global equivalences
    pClp = Gia_ManDupCollapse( pNew, pNew.pAigExtra, nil, pParsC ? 0 : 1 );
    //Gia_DumpAiger( pClp, p->pSpec, 1, 1 );
    // compute equivalences
    if pParsC {
        Gia_ManFraigSweepPerform( pClp, pParsC );
    } else if pParsS {
        Cec_ManLSCorrespondenceClasses( pClp, (Cec_ParCor_t *)pParsS );
    } else { 
        Gia_ManSeqCleanupClasses( pClp, fConst, fEquiv, fVerbose ); }
    // transfer equivalences
    pReprs = Gia_ManFraigSelectReprs( pNew, pClp, fVerbose, pFlopTypes );
    Gia_ManStop( pClp );
    // reduce AIG
    Gia_ManTransferTiming( p, pNew );
    pNew = Gia_ManFraigReduceGia( pTemp = pNew, pReprs );
    Gia_ManTransferTiming( pNew, p );
    Gia_ManStop( pTemp );
    // inlining of macro ABC_FREE
    pReprs = nil;
    // derive new AIG
    pNew = Gia_ManDupWithBoxes( pTemp = pNew, pParsC ? 0 : 1 );
    Gia_ManStop( pTemp );
    // report
    nFlopsNew = Vec_IntCountEntry(pNew.vRegClasses, 1);
    pFlopTypes[2] = nFlops - nFlopsNew - (pFlopTypes[0] + pFlopTypes[1]);
    if fVerbEquivs != 0 {
        fmt.Printf("Domain %2d : %5d -> %5d :  ", 1, nFlops, nFlopsNew);
        fmt.Printf("EqConst =%4d.  EqFlop =%4d.  Dangling =%4d.  Unused =%4d.\n", pFlopTypes[0], pFlopTypes[1], Abc_MaxInt(0, pFlopTypes[2]), Abc_MaxInt(0, -pFlopTypes[2]));
    }
    // normalize the result
    pNew = Gia_ManDupNormalize( pTemp = pNew, 0 );
    Gia_ManTransferTiming( pNew, pTemp );
    Gia_ManStop( pTemp );
    // check integrity
    //Gia_ManCheckIntegrityWithBoxes( pNew );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


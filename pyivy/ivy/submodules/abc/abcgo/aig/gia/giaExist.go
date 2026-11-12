package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaExist.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Existential quantification.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaExist.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Existentially quantified several variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func Gia_ManQuantInfoId(p *Gia_Man_t, iObj int32) *word { return Vec_WrdEntryP( p.vSuppWords, p.nSuppWords * iObj ); }
func Gia_ManQuantInfo(p *Gia_Man_t, pObj *Gia_Obj_t) *word { return Gia_ManQuantInfoId( p, Gia_ObjId(p, pObj) );      }

func Gia_ObjCopyGetTwoArray(p *Gia_Man_t, iObj int32, LitsOut []int32) { 
    var pLits *int32  = Vec_IntEntryP( &p.vCopiesTwo, 2*iObj );
    LitsOut[0] = pLits[0];
    LitsOut[1] = pLits[1];
}
func Gia_ObjCopySetTwoArray(p *Gia_Man_t, iObj int32, LitsIn []int32) { 
    var pLits *int32  = Vec_IntEntryP( &p.vCopiesTwo, 2*iObj );
    pLits[0] = LitsIn[0];
    pLits[1] = LitsIn[1];
}


/**Function*************************************************************

  Synopsis    [Existentially quantified several variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManQuantVerify_rec(p *Gia_Man_t, iObj int32, CiId int32) int32 {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId( p, iObj ) {
        return 0; }
    Gia_ObjSetTravIdCurrentId( p, iObj );
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        return Gia_ObjCioId(pObj) == CiId; }
    return Gia_ManQuantVerify_rec( p, Gia_ObjFaninId0(pObj, iObj), CiId ) ||
           Gia_ManQuantVerify_rec( p, Gia_ObjFaninId1(pObj, iObj), CiId );
}
func Gia_ManQuantVerify(p *Gia_Man_t, iObj int32) {
    var pInfo *word  = Gia_ManQuantInfoId( p, iObj );  var i, CiId int32;
    assert(Gia_ObjIsAnd(Gia_ManObj(p, iObj)), "Gia_ObjIsAnd(Gia_ManObj(p, iObj))");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(&p.vSuppVars)) && (((CiId) = Vec_IntEntry(&p.vSuppVars, i)), 1); i++  {
        Gia_ManIncrementTravId( p );
        if Abc_TtGetBit(pInfo, i) != Gia_ManQuantVerify_rec(p, iObj, CiId) {
            fmt.Printf("Mismatch at node %d related to CI %d (%d).\n", iObj, CiId, Abc_TtGetBit(pInfo, i)); }
    }
}



func Gia_ManQuantSetSuppStart(p *Gia_Man_t) {
    assert(Gia_ManObjNum(p) == 1, "Gia_ManObjNum(p) == 1");
    assert(p.vSuppWords == nil, "p.vSuppWords == nil");
    assert(Vec_IntSize(&p.vSuppVars) == 0, "Vec_IntSize(&p.vSuppVars) == 0");
    p.iSuppPi    = 0;
    p.nSuppWords = 1;
    p.vSuppWords = Vec_WrdAlloc( 1000 );
    Vec_WrdPush( p.vSuppWords, 0 );
}
func Gia_ManQuantSetSuppZero(p *Gia_Man_t) {
    var w int32;
    for  w = 0; w < p.nSuppWords; w++  {
        Vec_WrdPush( p.vSuppWords, 0 );
    }
    assert(Vec_WrdSize(p.vSuppWords) == p.nSuppWords * Gia_ManObjNum(p), "Vec_WrdSize(p.vSuppWords) == p.nSuppWords * Gia_ManObjNum(p)");
}
func Gia_ManQuantSetSuppCi(p *Gia_Man_t, pObj *Gia_Obj_t) {
    assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)");
    assert(p.vSuppWords != nil, "p.vSuppWords != nil");
    if p.iSuppPi == 64 * p.nSuppWords {
        var Data word; var w int32;
var Count int32;
var Size int32 = Vec_WrdSize(p.vSuppWords);
        var vTemp *Vec_Wrd_t  = Vec_WrdAlloc( Size ? 2 * Size : 1000 ); 
        // inlining of macro Vec_WrdForEachEntry
        for  w = 0; (w < Vec_WrdSize(p.vSuppWords)) && (((Data) = Vec_WrdEntry(p.vSuppWords, w)), 1); w++  {
            Vec_WrdPush( vTemp, Data );
            if Count++; Count == p.nSuppWords {
                Vec_WrdPush( vTemp, 0 );
                Count = 0;
            }
        }
        Vec_WrdFree( p.vSuppWords );
        p.vSuppWords = vTemp;
        p.nSuppWords++;
        assert(Vec_WrdSize(p.vSuppWords) == p.nSuppWords * Gia_ManObjNum(p), "Vec_WrdSize(p.vSuppWords) == p.nSuppWords * Gia_ManObjNum(p)");
        //printf( "Resizing to %d words.\n", p->nSuppWords );
    }
    assert(p.iSuppPi == Vec_IntSize(&p.vSuppVars), "p.iSuppPi == Vec_IntSize(&p.vSuppVars)");
    Vec_IntPush( &p.vSuppVars, Gia_ObjCioId(pObj) );
    Abc_TtSetBit( Gia_ManQuantInfo(p, pObj), p.iSuppPi++ );
}

func Gia_ManQuantCheckSupp(p *Gia_Man_t, iObj int32, iSupp int32) int32 {
    return Abc_TtGetBit( Gia_ManQuantInfoId(p, iObj), iSupp );
}
func Gia_ManQuantUpdateCiSupp(p *Gia_Man_t, iObj int32) {
    if Abc_TtIsConst0( Gia_ManQuantInfoId(p, iObj), p.nSuppWords ) {
        Gia_ManQuantSetSuppCi( p, Gia_ManObj(p, iObj) ); }
    assert(!Abc_TtIsConst0( Gia_ManQuantInfoId(p, iObj), p.nSuppWords ), "!Abc_TtIsConst0( Gia_ManQuantInfoId(p, iObj), p.nSuppWords )");
}
func Gia_ManQuantCheckOverlap(p *Gia_Man_t, iObj int32) int32 {
    return Abc_TtIntersect( Gia_ManQuantInfoId(p, iObj), Gia_ManQuantInfoId(p, 0), p.nSuppWords, 0 );
}
func Gia_ManQuantMarkUsedCis(p *Gia_Man_t, pFuncCiToKeep func(any, int32) int32, pData any) {
    var i, CiId int32;
    var pInfo *word  = Gia_ManQuantInfoId( p, 0 );
    Abc_TtClear( pInfo, p.nSuppWords );
    assert(Abc_TtIsConst0(pInfo, p.nSuppWords), "Abc_TtIsConst0(pInfo, p.nSuppWords)");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(&p.vSuppVars)) && (((CiId) = Vec_IntEntry(&p.vSuppVars, i)), 1); i++  {
        if !pFuncCiToKeep( pData, CiId ) { // quant var
            Abc_TtSetBit( pInfo, i ); }
    }
}

func Gia_ManQuantCountUsed_rec(p *Gia_Man_t, iObj int32) int32 {
    var pObj *Gia_Obj_t; var Count int32  = 1;
    if Gia_ObjIsTravIdCurrentId( p, iObj ) {
        return 0; }
    Gia_ObjSetTravIdCurrentId( p, iObj );
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        return 0; }
    if Gia_ManQuantCheckSupp(p, Gia_ObjFaninId0(pObj, iObj), p.iSuppPi) != 0 {
        Count += Gia_ManQuantCountUsed_rec( p, Gia_ObjFaninId0(pObj, iObj) ); }
    if Gia_ManQuantCheckSupp(p, Gia_ObjFaninId1(pObj, iObj), p.iSuppPi) != 0 {
        Count += Gia_ManQuantCountUsed_rec( p, Gia_ObjFaninId1(pObj, iObj) ); }
    return Count;
}
func Gia_ManQuantCountUsed(p *Gia_Man_t, iObj int32) int32 {
    Gia_ManIncrementTravId( p );
    return Gia_ManQuantCountUsed_rec( p, iObj );
}

func Gia_ManQuantDupConeSupp_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, vCis *Vec_Int_t, vObjs *Vec_Int_t, pFuncCiToKeep func(any, int32) int32, pData any) {
    var iLit0 int32;
    var iLit1 int32;
    var iObj int32 = Gia_ObjId( p, pObj );
    var iLit int32  = Gia_ObjCopyArray( p, iObj );
    if iLit >= 0 {
        return; }
    if Gia_ObjIsCi(pObj) {
        var iLit int32  = Gia_ManAppendCi( pNew );
        var pObjNew *Gia_Obj_t  = Gia_ManObj( pNew, Abc_Lit2Var(iLit) );
        Gia_ManQuantSetSuppZero( pNew );
        if !pFuncCiToKeep( pData, Gia_ObjCioId(pObj) ) {
            //printf( "Collecting CI %d\n", Gia_ObjCioId(pObj)+1 );
            Gia_ManQuantSetSuppCi( pNew, pObjNew );
        }
        Gia_ObjSetCopyArray( p, iObj, iLit );
        Vec_IntPush( vCis, iObj );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManQuantDupConeSupp_rec( pNew, p, Gia_ObjFanin0(pObj), vCis, vObjs, pFuncCiToKeep, pData );
    Gia_ManQuantDupConeSupp_rec( pNew, p, Gia_ObjFanin1(pObj), vCis, vObjs, pFuncCiToKeep, pData );
    iLit0 = Gia_ObjCopyArray( p, Gia_ObjFaninId0(pObj, iObj) );
    iLit1 = Gia_ObjCopyArray( p, Gia_ObjFaninId1(pObj, iObj) );
    iLit0 = Abc_LitNotCond( iLit0, Gia_ObjFaninC0(pObj) );
    iLit1 = Abc_LitNotCond( iLit1, Gia_ObjFaninC1(pObj) );
    iLit  = Gia_ManHashAnd( pNew, iLit0, iLit1 );
    Gia_ObjSetCopyArray( p, iObj, iLit );
    Vec_IntPush( vObjs, iObj );
}
func Gia_ManQuantDupConeSupp(p *Gia_Man_t, iLit int32, pFuncCiToKeep func(any, int32) int32, pData any, pvCis **Vec_Int_t, pOutLit *int32) *Gia_Man_t {
    var pNew *Gia_Man_t; var i, iLit0, iObj int32;
    var pObj *Gia_Obj_t;
    var pRoot *Gia_Obj_t = Gia_ManObj( p, Abc_Lit2Var(iLit) );
    var vCis *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vObjs *Vec_Int_t  = Vec_IntAlloc( 1000 );
    assert(Gia_ObjIsAnd(pRoot), "Gia_ObjIsAnd(pRoot)");
    if Vec_IntSize(&p.vCopies) < Gia_ManObjNum(p) {
        Vec_IntFillExtra( &p.vCopies, Gia_ManObjNum(p), -1 ); }
    pNew = Gia_ManStart( 1000 );
    Gia_ManHashStart( pNew ); 
    Gia_ManQuantSetSuppStart( pNew );
    Gia_ManQuantDupConeSupp_rec( pNew, p, pRoot, vCis, vObjs, pFuncCiToKeep, pData );
    iLit0 = Gia_ObjCopyArray( p, Abc_Lit2Var(iLit) );
    iLit0 = Abc_LitNotCond( iLit0, Abc_LitIsCompl(iLit) );
    if pOutLit != nil { *pOutLit = iLit0; }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCis)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vCis,i))); i++  {
        Gia_ObjSetCopyArray( p, Gia_ObjId(p, pObj), -1 );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        Gia_ObjSetCopyArray( p, Gia_ObjId(p, pObj), -1 );
    }
    //assert( Vec_IntCountLarger(&p->vCopies, -1) == 0 );
    Vec_IntFree( vObjs );
    // remap into CI Ids
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCis)) && (((iObj) = Vec_IntEntry(vCis, i)), 1); i++  {
        Vec_IntWriteEntry( vCis, i, Gia_ManIdToCioId(p, iObj) );
    }
    if pvCis { *pvCis = vCis; }
    return pNew;
}
func Gia_ManQuantExist_rec(p *Gia_Man_t, iObj int32, pRes []int32) {
    var pObj *Gia_Obj_t;
    var Lits0, Lits1, pFans, fCompl [2]int32;
    if Gia_ObjIsTravIdCurrentId( p, iObj ) {
        Gia_ObjCopyGetTwoArray( p, iObj, pRes );
        return;
    }
    Gia_ObjSetTravIdCurrentId( p, iObj );
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        pRes[0] = 0; pRes[1] = 1;
        Gia_ObjCopySetTwoArray( p, iObj, pRes );
        return;
    }
    pFans[0]  = Gia_ObjFaninId0( pObj, iObj );
    pFans[1]  = Gia_ObjFaninId1( pObj, iObj );
    fCompl[0] = Gia_ObjFaninC0( pObj );
    fCompl[1] = Gia_ObjFaninC1( pObj );
    if Gia_ManQuantCheckSupp(p, pFans[0], p.iSuppPi) != 0 {
        Gia_ManQuantExist_rec( p, pFans[0], Lits0 );
    } else {
        Lits0[0] = Lits0[1] = Abc_Var2Lit( pFans[0], 0 ); }
    if Gia_ManQuantCheckSupp(p, pFans[1], p.iSuppPi) != 0 {
        Gia_ManQuantExist_rec( p, pFans[1], Lits1 );
    } else {
        Lits1[0] = Lits1[1] = Abc_Var2Lit( pFans[1], 0 ); }
    pRes[0] = Gia_ManHashAnd( p, Abc_LitNotCond(Lits0[0], fCompl[0]), Abc_LitNotCond(Lits1[0], fCompl[1]) );
    pRes[1] = Gia_ManHashAnd( p, Abc_LitNotCond(Lits0[1], fCompl[0]), Abc_LitNotCond(Lits1[1], fCompl[1]) );
    Gia_ObjCopySetTwoArray( p, iObj, pRes );
}
func Gia_ManQuantExist2(p0 *Gia_Man_t, iLit int32, pFuncCiToKeep func(any, int32) int32, pData any) int32 {
//    abctime clk = Abc_Clock();
    var pNew *Gia_Man_t;
    var vOuts, vOuts2, vCis *Vec_Int_t; 
    var pObj *Gia_Obj_t  = Gia_ManObj( p0, Abc_Lit2Var(iLit) );
    var i int32;
    var n int32;
    var Entry int32;
    var Lit int32;
    var OutLit int32 = -1;
    var pLits [2]int32;
    var nVarsQua int32;
    var nAndsOld int32;
    var nAndsNew int32; 
    if iLit < 2 { return iLit; }
    if Gia_ObjIsCi(pObj) { if pFuncCiToKeep(pData, Gia_ObjCioId(pObj)) {
        return iLit;
    } else {
        return 1;
    } }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    pNew = Gia_ManQuantDupConeSupp( p0, iLit, pFuncCiToKeep, pData, &vCis, &OutLit );
    if pNew.iSuppPi == 0 {
        Gia_ManStop( pNew );
        Vec_IntFree( vCis );
        return iLit;
    }
    assert(pNew.iSuppPi > 0 && pNew.iSuppPi <= 64 * pNew.nSuppWords, "pNew.iSuppPi > 0 && pNew.iSuppPi <= 64 * pNew.nSuppWords");
    vOuts = Vec_IntAlloc( 100 );
    vOuts2 = Vec_IntAlloc( 100 );
    assert(OutLit > 1, "OutLit > 1");
    Vec_IntPush( vOuts, OutLit );
    nVarsQua = pNew.iSuppPi;
    nAndsOld = Gia_ManAndNum(pNew);
    for  --pNew.iSuppPi >= 0  {
        Entry = Abc_Lit2Var( Vec_IntEntry(vOuts, 0) );
//        printf( "Quantifying input %d with %d affected nodes (out of %d):\n", 
//            pNew->iSuppPi, Gia_ManQuantCountUsed(pNew,Entry), Gia_ManAndNum(pNew) );
        if Vec_IntSize(&pNew.vCopiesTwo) < 2*Gia_ManObjNum(pNew) {
            Vec_IntFillExtra( &pNew.vCopiesTwo, 2*Gia_ManObjNum(pNew), -1 ); }
        assert(Vec_IntSize(vOuts) > 0, "Vec_IntSize(vOuts) > 0");
        Vec_IntClear( vOuts2 );
        Gia_ManIncrementTravId( pNew );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vOuts)) && (((Entry) = Vec_IntEntry(vOuts, i)), 1); i++  {
            Gia_ManQuantExist_rec( pNew, Abc_Lit2Var(Entry), pLits );
            for  n = 0; n < 2; n++  {
                Lit = Abc_LitNotCond( pLits[n], Abc_LitIsCompl(Entry) );
                if Lit == 0 {
                    continue; }
                if Lit == 1 {
                    Vec_IntFree( vOuts );
                    Vec_IntFree( vOuts2 );
                    Gia_ManStop( pNew );
                    Vec_IntFree( vCis );
                    return 1;
                }
                Vec_IntPushUnique( vOuts2, Lit );
            }
        }
        Vec_IntClear( vOuts );
        // inlining of macro ABC_SWAP
        { var t *Vec_Int_t  = vOuts; vOuts = vOuts2; vOuts2 = t; };
    }
//    printf( "\n" );
    //printf( "The number of diff cofactors = %d.\n", Vec_IntSize(vOuts) );
    assert(Vec_IntSize(vOuts) > 0, "Vec_IntSize(vOuts) > 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOuts)) && (((Entry) = Vec_IntEntry(vOuts, i)), 1); i++  {
        Vec_IntWriteEntry( vOuts, i, Abc_LitNot(Entry) );
    }
    OutLit = Gia_ManHashAndMulti( pNew, vOuts );
    OutLit = Abc_LitNot( OutLit );
    Vec_IntFree( vOuts );
    Vec_IntFree( vOuts2 );
    // transfer back
    Gia_ManAppendCo( pNew, OutLit );
    nAndsNew = Gia_ManAndNum(p0);
    Lit = Gia_ManDupConeBack( p0, pNew, vCis );
    nAndsNew = Gia_ManAndNum(p0) - nAndsNew;
    Gia_ManStop( pNew );
    // report the result
//    printf( "Performed quantification with %6d nodes, %3d keep-vars, %3d quant-vars, resulting in %5d new nodes. \n", 
//        nAndsOld, Vec_IntSize(vCis) - nVarsQua, nVarsQua, nAndsNew );
//    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Vec_IntFree( vCis );
    return Lit;
}


/**Function*************************************************************

  Synopsis    [Existentially quantified several variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManQuantCollect_rec(p *Gia_Man_t, iObj int32, vQuantCis *Vec_Int_t, vQuantSide *Vec_Int_t, vQuantAnds *Vec_Int_t) {
    var pObj *Gia_Obj_t; 
    if Gia_ObjIsTravIdCurrentId( p, iObj ) {
        return; }
    Gia_ObjSetTravIdCurrentId( p, iObj );
    if !(Gia_ManQuantCheckOverlap(p, iObj) != 0) {
        Vec_IntPush( vQuantSide, iObj );
        return;
    }
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        Vec_IntPush( vQuantCis, iObj );
        return;
    }
    Gia_ManQuantCollect_rec( p, Gia_ObjFaninId0(pObj, iObj), vQuantCis, vQuantSide, vQuantAnds );
    Gia_ManQuantCollect_rec( p, Gia_ObjFaninId1(pObj, iObj), vQuantCis, vQuantSide, vQuantAnds );
    Vec_IntPush( vQuantAnds, iObj );
}
func Gia_ManQuantCollect(p *Gia_Man_t, iObj int32, pFuncCiToKeep func(any, int32) int32, pData any, vQuantCis *Vec_Int_t, vQuantSide *Vec_Int_t, vQuantAnds *Vec_Int_t) {
    Gia_ManQuantMarkUsedCis( p, pFuncCiToKeep, pData );
    Gia_ManIncrementTravId( p );
    Gia_ManQuantCollect_rec( p, iObj, vQuantCis, vQuantSide, vQuantAnds );
//    printf( "\nCreated cone with %d quant-vars, %d side-inputs, and %d internal nodes.\n", 
//        Vec_IntSize(p->vQuantCis), Vec_IntSize(p->vQuantSide), Vec_IntSize(p->vQuantAnds) );
}

func Gia_ManQuantExist2Dup(p *Gia_Man_t, iLit int32, vCis *Vec_Int_t, vSide *Vec_Int_t, vAnds *Vec_Int_t, pOutLit *int32) *Gia_Man_t {
    var i, iObj, iLit0, iLit1, iLitR int32;
    var pNew *Gia_Man_t  = Gia_ManStart( Vec_IntSize(vSide) + Vec_IntSize(vCis) + 10*Vec_IntSize(vAnds) );
    Gia_ManQuantSetSuppStart( pNew );
    Gia_ManHashStart( pNew ); 
    if Vec_IntSize(&p.vCopies) < Gia_ManObjNum(p) {
        Vec_IntFillExtra( &p.vCopies, Gia_ManObjNum(p), -1 ); }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSide)) && (((iObj) = Vec_IntEntry(vSide, i)), 1); i++  {
        Gia_ObjSetCopyArray( p, iObj, Gia_ManAppendCi(pNew) );
        Gia_ManQuantSetSuppZero( pNew );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCis)) && (((iObj) = Vec_IntEntry(vCis, i)), 1); i++  {
        Gia_ObjSetCopyArray( p, iObj, (iLit0 = Gia_ManAppendCi(pNew)) );
        Gia_ManQuantSetSuppZero( pNew );
        Gia_ManQuantSetSuppCi( pNew, Gia_ManObj(pNew, Abc_Lit2Var(iLit0)) );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vAnds)) && (((iObj) = Vec_IntEntry(vAnds, i)), 1); i++  {
        var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
        iLit0 = Gia_ObjCopyArray( p, Gia_ObjFaninId0(pObj, iObj) );
        iLit1 = Gia_ObjCopyArray( p, Gia_ObjFaninId1(pObj, iObj) );
        iLit0 = Abc_LitNotCond( iLit0, Gia_ObjFaninC0(pObj) );
        iLit1 = Abc_LitNotCond( iLit1, Gia_ObjFaninC1(pObj) );
        iLitR = Gia_ManHashAnd( pNew, iLit0, iLit1 );
        Gia_ObjSetCopyArray( p, iObj, iLitR );
    }
    iLit0 = Gia_ObjCopyArray( p, Abc_Lit2Var(iLit) );
    iLit0 = Abc_LitNotCond( iLit0, Abc_LitIsCompl(iLit) );
    if pOutLit != nil { *pOutLit = iLit0; }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSide)) && (((iObj) = Vec_IntEntry(vSide, i)), 1); i++  {
        Gia_ObjSetCopyArray( p, iObj, -1 );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCis)) && (((iObj) = Vec_IntEntry(vCis, i)), 1); i++  {
        Gia_ObjSetCopyArray( p, iObj, -1 );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vAnds)) && (((iObj) = Vec_IntEntry(vAnds, i)), 1); i++  {
        Gia_ObjSetCopyArray( p, iObj, -1 );
    }
    return pNew;
}
func Gia_ManQuantExistInt(p0 *Gia_Man_t, iLit int32, vCis *Vec_Int_t, vSide *Vec_Int_t, vAnds *Vec_Int_t) int32 {
    var i int32;
    var Lit int32;
    var iOutLit int32;
    var nAndsNew int32;
    var pLits [2]int32;
    var pRes [2]int32 ;
    var pNew *Gia_Man_t;
    if iLit < 2 {
        return 0; }
    if Vec_IntSize(vCis) == 0 {
        return iLit; }
    if Vec_IntSize(vAnds) == 0 {
        assert(Gia_ObjIsCi( Gia_ManObj(p0, Abc_Lit2Var(iLit)) ), "Gia_ObjIsCi( Gia_ManObj(p0, Abc_Lit2Var(iLit)) )");
        if Vec_IntFind(vCis, Abc_Lit2Var(iLit)) == -1 {
            return iLit;
        } else {
            return 1;
        }
    }
    pNew = Gia_ManQuantExist2Dup( p0, iLit, vCis, vSide, vAnds, &iOutLit );
    if Vec_IntSize(&pNew.vCopiesTwo) < 2*Gia_ManObjNum(pNew) {
        Vec_IntFillExtra( &pNew.vCopiesTwo, 2*Gia_ManObjNum(pNew), -1 ); }
    Gia_ObjCopySetTwoArray( pNew, 0, pRes );
    for  i = 0; i < Gia_ManCiNum(pNew); i++  {
        pRes[0] = pRes[1] = Abc_Var2Lit( i+1, 0 );
        Gia_ObjCopySetTwoArray( pNew, i+1, pRes );
    }
    assert(pNew.iSuppPi == Gia_ManCiNum(pNew) - Vec_IntSize(vSide), "pNew.iSuppPi == Gia_ManCiNum(pNew) - Vec_IntSize(vSide)");
    for  i = Gia_ManCiNum(pNew) - 1; i >= Vec_IntSize(vSide); i--  {
        pRes[0] = 0; pRes[1] = 1;
        Gia_ObjCopySetTwoArray( pNew, i+1, pRes );

        pNew.iSuppPi--;
        if Vec_IntSize(&pNew.vCopiesTwo) < 2*Gia_ManObjNum(pNew) {
            Vec_IntFillExtra( &pNew.vCopiesTwo, 2*Gia_ManObjNum(pNew), -1 ); }
        Gia_ManIncrementTravId( pNew );
        Gia_ManQuantExist_rec( pNew, Abc_Lit2Var(iOutLit), pLits );
        pLits[0] = Abc_LitNotCond( pLits[0], Abc_LitIsCompl(iOutLit) );
        pLits[1] = Abc_LitNotCond( pLits[1], Abc_LitIsCompl(iOutLit) );
        iOutLit = Gia_ManHashOr( pNew, pLits[0], pLits[1] );

        pRes[0] = pRes[1] = Abc_Var2Lit( i+1, 0 );
        Gia_ObjCopySetTwoArray( pNew, i+1, pRes );
    }
    assert(pNew.iSuppPi == 0, "pNew.iSuppPi == 0");
    Vec_IntAppend( vSide, vCis );
    // transfer back
    Gia_ManAppendCo( pNew, iOutLit );
    nAndsNew = Gia_ManAndNum(p0);
    Lit = Gia_ManDupConeBackObjs( p0, pNew, vSide );
    nAndsNew = Gia_ManAndNum(p0) - nAndsNew;
    Vec_IntShrink( vSide, Vec_IntSize(vSide) - Vec_IntSize(vCis) );
//    printf( "Performed quantification with %6d -> %6d nodes, %3d side-vars, %3d quant-vars, resulting in %5d new nodes. \n", 
//        Vec_IntSize(vAnds), Gia_ManAndNum(pNew), Vec_IntSize(vSide), Vec_IntSize(vCis), nAndsNew );
    Gia_ManStop( pNew );
    return Lit;
}
func Gia_ManQuantExist(p0 *Gia_Man_t, iLit int32, pFuncCiToKeep func(any, int32) int32, pData any) int32 {
    var Res int32;
    var vQuantCis *Vec_Int_t   = Vec_IntAlloc( 100 );
    var vQuantSide *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vQuantAnds *Vec_Int_t  = Vec_IntAlloc( 100 );
    Gia_ManQuantCollect( p0, Abc_Lit2Var(iLit), pFuncCiToKeep, pData, vQuantCis, vQuantSide, vQuantAnds );
    Res = Gia_ManQuantExistInt( p0, iLit, vQuantCis, vQuantSide, vQuantAnds );
    Vec_IntFree( vQuantCis );
    Vec_IntFree( vQuantSide );
    Vec_IntFree( vQuantAnds );
    return Res;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


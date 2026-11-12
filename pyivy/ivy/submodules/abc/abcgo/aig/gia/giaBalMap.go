package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaSopb.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [SOP balancing for a window.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSopb.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManHighlight_rec(p *Gia_Man_t, iObj int32) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsAnd(pObj) {
        Gia_ManHighlight_rec( p, Gia_ObjFaninId0(pObj, iObj) ); }
    if Gia_ObjIsAnd(pObj) {
        Gia_ManHighlight_rec( p, Gia_ObjFaninId1(pObj, iObj) ); }
}
func Gia_ManPrepareWin(p *Gia_Man_t, vOuts *Vec_Int_t, pvPis **Vec_Int_t, pvPos **Vec_Int_t, pvAnds **Vec_Int_t, fPoOnly int32) {
    var pObj *Gia_Obj_t;
    var i int32;
    // mark the section
    Gia_ManIncrementTravId( p );
    // inlining of macro Gia_ManForEachCoVec
    for  i = 0; (i < Vec_IntSize(vOuts)) && ((pObj) = Gia_ManCo(p, Vec_IntEntry(vOuts,i))); i++  {
        Gia_ManHighlight_rec( p, Gia_ObjFaninId0p(p, pObj) );
    }
    // mark fanins of the outside area
    Gia_ManCleanMark0( p );
    if fPoOnly != 0 {
        // inlining of macro Gia_ManForEachCoVec
        for  i = 0; (i < Vec_IntSize(vOuts)) && ((pObj) = Gia_ManCo(p, Vec_IntEntry(vOuts,i))); i++  {
            Gia_ObjFanin0(pObj).fMark0 = 1;
        }
    
    } else {
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if Gia_ObjIsCi(pObj) {
                continue; }
            if Gia_ObjIsAnd(pObj) && !Gia_ObjIsTravIdCurrentId(p, i) {
                continue; }
            Gia_ObjFanin0(pObj).fMark0 = 1;
            if Gia_ObjIsAnd(pObj) {
                Gia_ObjFanin1(pObj).fMark0 = 1; }
        }
    }
    // collect pointed nodes
    *pvPis  = Vec_IntAlloc( 1000 );
    *pvPos  = Vec_IntAlloc( 1000 );
    *pvAnds = Vec_IntAlloc( 1000 );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if !Gia_ObjIsTravIdCurrentId(p, i) {
            continue; }
        if Gia_ObjIsCi(pObj) {
            Vec_IntPush( *pvPis, i );
        } else if pObj.fMark0 {
            Vec_IntPush( *pvPos, i ); }
        if Gia_ObjIsAnd(pObj) {
            Vec_IntPush( *pvAnds, i ); }
    }
    Gia_ManCleanMark0( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManExtractWin(p *Gia_Man_t, vOuts *Vec_Int_t, fPoOnly int32) *Gia_Man_t {
    var vPis, vPos, vAnds *Vec_Int_t;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManPrepareWin( p, vOuts, &vPis, &vPos, &vAnds, fPoOnly );
    // create AIG
    pNew = Gia_ManStart( Vec_IntSize(vPis) + Vec_IntSize(vPos) + Vec_IntSize(vAnds) + 1 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPis)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vPis,i))); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vAnds)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vAnds,i))); i++  {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPos)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vPos,i))); i++  {
        Gia_ManAppendCo( pNew, pObj.Value );
    }
    Vec_IntFree( vPis );
    Vec_IntFree( vPos );
    Vec_IntFree( vAnds );
    return pNew;
}
func Gia_ManInsertWin(p *Gia_Man_t, vOuts *Vec_Int_t, pWin *Gia_Man_t) *Gia_Man_t {
    var vPos, vPis, vAnds *Vec_Int_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManPrepareWin( p, vOuts, &vPis, &vPos, &vAnds, 0 );
    // create AIG
    pNew = Gia_ManStart( Gia_ManObjNum(p) - Vec_IntSize(vAnds) + Gia_ManAndNum(pWin) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // inputs
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    Gia_ManConst0(pWin).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pWin.vCis)) && ((pObj) = Gia_ManCi(pWin, i)); i++  {
        pObj.Value = Gia_ManObj(p, Vec_IntEntry(vPis, i)).Value;
    }
    // internal nodes
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pWin.nObjs) && ((pObj) = Gia_ManObj(pWin, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pWin.vCos)) && ((pObj) = Gia_ManCo(pWin, i)); i++  {
        Gia_ManObj( p, Vec_IntEntry(vPos, i) ).Value = Gia_ObjFanin0Copy(pObj);
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if !Gia_ObjIsTravIdCurrentId(p, i) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    // cleanup
    Vec_IntFree( vPis );
    Vec_IntFree( vPos );
    Vec_IntFree( vAnds );
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
func Gia_ManFindLatest(p *Gia_Man_t, LevelMax int32, nTimeWindow int32) *Vec_Int_t {
    var pObj *Gia_Obj_t;
    var vOuts *Vec_Int_t;
    vOuts = Vec_IntAlloc( 1000 );
    if Gia_ManHasMapping(p) {
        var i, k, iFan, nLevels int32;
        // inlining of macro ABC_CALLOC
        var pLevels *int32  = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
        // inlining of macro Gia_ManForEachLut
        for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsLut(p, i) {} else {
            // inlining of macro Gia_LutForEachFanin
            for  k = 0; k < Gia_ObjLutSize(p,i) && ((iFan = Gia_ObjLutFanins(p,i)[k]),1); k++  {
                pLevels[i] = Abc_MaxInt( pLevels[i], pLevels[iFan] );
            }
            pLevels[i]++;
            nLevels = Abc_MaxInt( nLevels, pLevels[i] );
        }
        }
        if nTimeWindow != 0 {
            LevelMax = int32(((1.0 - 0.01 * nTimeWindow) * nLevels)); }
        if nLevels < LevelMax {
            fmt.Printf("The maximum mapped level (%d) is less than the target level (%d).\n", nLevels, LevelMax); }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if pLevels[Gia_ObjFaninId0p(p, pObj)] >= LevelMax {
                Vec_IntPush( vOuts, i ); }
        }
        // inlining of macro ABC_FREE
        ((pLevels) ? (free((char *) (pLevels)), (pLevels) = 0) : 0);
    
    } else {
        var i int32;
        var nLevels int32 = Gia_ManLevelNum( p );
        if nTimeWindow != 0 {
            LevelMax = int32(((1.0 - 0.01 * nTimeWindow) * nLevels)); }
        if nLevels < LevelMax {
            fmt.Printf("The maximum AIG level (%d) is less than the target level (%d).\n", nLevels, LevelMax); }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if Gia_ObjLevel(p, pObj) >= LevelMax {
                Vec_IntPush( vOuts, i ); }
        }
    }
    return vOuts;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManExtractWindow(p *Gia_Man_t, LevelMax int32, nTimeWindow int32, fVerbose int32) *Gia_Man_t {
    var vOuts *Vec_Int_t;
    var pWin *Gia_Man_t;
    assert(!LevelMax != !nTimeWindow, "!LevelMax != !nTimeWindow");
    vOuts = Gia_ManFindLatest( p, LevelMax, nTimeWindow );
    if fVerbose != 0 {
        fmt.Printf("Collected %d outputs to extract.\n", Vec_IntSize(vOuts)); }
    if Vec_IntSize(vOuts) == 0 {
        Vec_IntFree( vOuts );
        return Gia_ManDup( p );
    }
    pWin = Gia_ManExtractWin( p, vOuts, 1 );
    Vec_IntFree( vOuts );
    return pWin;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformSopBalanceWin(p *Gia_Man_t, LevelMax int32, nTimeWindow int32, nCutNum int32, nRelaxRatio int32, fVerbose int32) *Gia_Man_t {
    var vOuts *Vec_Int_t;
    var pNew, pWin, pWinNew *Gia_Man_t;
    assert(!LevelMax != !nTimeWindow, "!LevelMax != !nTimeWindow");
    vOuts = Gia_ManFindLatest( p, LevelMax, nTimeWindow );
    if fVerbose != 0 {
        fmt.Printf("Collected %d outputs to extract.\n", Vec_IntSize(vOuts)); }
    if Vec_IntSize(vOuts) == 0 {
        Vec_IntFree( vOuts );
        return Gia_ManDup( p );
    }
    pWin = Gia_ManExtractWin( p, vOuts, 0 );
    pWinNew = Gia_ManPerformSopBalance( pWin, nCutNum, nRelaxRatio, fVerbose );
    Gia_ManStop( pWin );
    pNew = Gia_ManInsertWin( p, vOuts, pWinNew );
    Gia_ManStop( pWinNew );
    Vec_IntFree( vOuts );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformDsdBalanceWin(p *Gia_Man_t, LevelMax int32, nTimeWindow int32, nLutSize int32, nCutNum int32, nRelaxRatio int32, fVerbose int32) *Gia_Man_t {
    var vOuts *Vec_Int_t;
    var pNew, pWin, pWinNew *Gia_Man_t;
    assert(!LevelMax != !nTimeWindow, "!LevelMax != !nTimeWindow");
    vOuts = Gia_ManFindLatest( p, LevelMax, nTimeWindow );
    if fVerbose != 0 {
        fmt.Printf("Collected %d outputs to extract.\n", Vec_IntSize(vOuts)); }
    if Vec_IntSize(vOuts) == 0 {
        Vec_IntFree( vOuts );
        return Gia_ManDup( p );
    }
    pWin = Gia_ManExtractWin( p, vOuts, 0 );
    pWinNew = Gia_ManPerformDsdBalance( pWin, nLutSize, nCutNum, nRelaxRatio, fVerbose );
    Gia_ManStop( pWin );
    pNew = Gia_ManInsertWin( p, vOuts, pWinNew );
    Gia_ManStop( pWinNew );
    Vec_IntFree( vOuts );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


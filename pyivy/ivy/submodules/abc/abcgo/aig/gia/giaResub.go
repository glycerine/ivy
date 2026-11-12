package gia

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [giaResub.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Resubstitution.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaResub.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecWec.h"
//#include "misc/vec/vecQue.h"
//#include "misc/vec/vecHsh.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes MFFCs of all qualifying nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjCheckMffc_rec(p *Gia_Man_t, pObj *Gia_Obj_t, Limit int32, vNodes *Vec_Int_t) int32 {
    var iFanin int32;
    if Gia_ObjIsCi(pObj) {
        return 1; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    iFanin = Gia_ObjFaninId0p(p, pObj);
    Vec_IntPush( vNodes, iFanin );
    if !Gia_ObjRefDecId(p, iFanin) && (Vec_IntSize(vNodes) > Limit || !(Gia_ObjCheckMffc_rec(p, Gia_ObjFanin0(pObj), Limit, vNodes) != 0)) {
        return 0; }
    iFanin = Gia_ObjFaninId1p(p, pObj);
    Vec_IntPush( vNodes, iFanin );
    if !Gia_ObjRefDecId(p, iFanin) && (Vec_IntSize(vNodes) > Limit || !(Gia_ObjCheckMffc_rec(p, Gia_ObjFanin1(pObj), Limit, vNodes) != 0)) {
        return 0; }
    if !Gia_ObjIsMux(p, pObj) {
        return 1; }
    iFanin = Gia_ObjFaninId2p(p, pObj);
    Vec_IntPush( vNodes, iFanin );
    if !Gia_ObjRefDecId(p, iFanin) && (Vec_IntSize(vNodes) > Limit || !(Gia_ObjCheckMffc_rec(p, Gia_ObjFanin2(p, pObj), Limit, vNodes) != 0)) {
        return 0; }
    return 1;
}
func Gia_ObjCheckMffc(p *Gia_Man_t, pRoot *Gia_Obj_t, Limit int32, vNodes *Vec_Int_t, vLeaves *Vec_Int_t, vInners *Vec_Int_t) int32 {
    var RetValue, iObj, i int32;
    Vec_IntClear( vNodes );
    RetValue = Gia_ObjCheckMffc_rec( p, pRoot, Limit, vNodes );
    if RetValue != 0 {
        Vec_IntClear( vLeaves );
        Vec_IntClear( vInners );
        Vec_IntSort( vNodes, 0 );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vNodes)) && (((iObj) = Vec_IntEntry(vNodes, i)), 1); i++  {
            if Gia_ObjRefNumId(p, iObj) > 0 || Gia_ObjIsCi(Gia_ManObj(p, iObj)) {
                if !Vec_IntSize(vLeaves) || Vec_IntEntryLast(vLeaves) != iObj {
                    Vec_IntPush( vLeaves, iObj ); }
            
            } else {
                if !Vec_IntSize(vInners) || Vec_IntEntryLast(vInners) != iObj {
                    Vec_IntPush( vInners, iObj ); }
            }
        }
        Vec_IntPush( vInners, Gia_ObjId(p, pRoot) );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((iObj) = Vec_IntEntry(vNodes, i)), 1); i++  {
        Gia_ObjRefIncId( p, iObj );
    }
    return RetValue;
}
func Gia_ManComputeMffcs(p *Gia_Man_t, LimitMin int32, LimitMax int32, SuppMax int32, RatioBest int32) *Vec_Wec_t {
    var pObj *Gia_Obj_t;
    var vMffcs *Vec_Wec_t;
    var vNodes, vLeaves, vInners, vMffc *Vec_Int_t;
    var i, iPivot int32;
    assert(p.pMuxes, "p.pMuxes");
    vNodes  = Vec_IntAlloc( 2 * LimitMax );
    vLeaves = Vec_IntAlloc( 2 * LimitMax );
    vInners = Vec_IntAlloc( 2 * LimitMax );
    vMffcs  = Vec_WecAlloc( 1000 );
    Gia_ManCreateRefs( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !Gia_ObjRefNum(p, pObj) {
            continue; }
        if !(Gia_ObjCheckMffc(p, pObj, LimitMax, vNodes, vLeaves, vInners) != 0) {
            continue; }
        if Vec_IntSize(vInners) < LimitMin {
            continue; }
        if Vec_IntSize(vLeaves) > SuppMax {
            continue; }
        // improve cut
        // collect cut
        vMffc = Vec_WecPushLevel( vMffcs );
        Vec_IntGrow( vMffc, Vec_IntSize(vLeaves) + Vec_IntSize(vInners) + 20 );
        Vec_IntPush( vMffc, i );
        Vec_IntPush( vMffc, Vec_IntSize(vLeaves) );
        Vec_IntPush( vMffc, Vec_IntSize(vInners) );
        Vec_IntAppend( vMffc, vLeaves );
//        Vec_IntAppend( vMffc, vInners );
        // add last entry equal to the ratio
        Vec_IntPush( vMffc, 1000 * Vec_IntSize(vInners) / Vec_IntSize(vLeaves) );
    }
    }
    Vec_IntFree( vNodes );
    Vec_IntFree( vLeaves );
    Vec_IntFree( vInners );
    // sort MFFCs by their inner/leaf ratio
    Vec_WecSortByLastInt( vMffcs, 1 );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vMffcs)) && (((vMffc) = Vec_WecEntry(vMffcs, i)), 1); i++  {
        Vec_IntPop( vMffc );
    }
    // remove those whose ratio is not good
    iPivot = RatioBest * Vec_WecSize(vMffcs) / 100;
    // inlining of macro Vec_WecForEachLevelStart
    for  i = iPivot; (i < Vec_WecSize(vMffcs)) && (((vMffc) = Vec_WecEntry(vMffcs, i)), 1); i++  {
        Vec_IntErase( vMffc );
    }
    assert(iPivot <= Vec_WecSize(vMffcs), "iPivot <= Vec_WecSize(vMffcs)");
    Vec_WecShrink( vMffcs, iPivot );
    return vMffcs;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintDivStats(p *Gia_Man_t, vMffcs *Vec_Wec_t, vPivots *Vec_Wec_t) {
    var fVerbose int32;
    var vMffc *Vec_Int_t;
    var i, nDivs, nDivsAll, nDivs0 int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vMffcs)) && (((vMffc) = Vec_WecEntry(vMffcs, i)), 1); i++  {
        nDivs = Vec_IntSize(vMffc) - 3 - Vec_IntEntry(vMffc, 1) - Vec_IntEntry(vMffc, 2);
        nDivs0 += (nDivs == 0);
        nDivsAll += nDivs;
        if !(fVerbose != 0) {
            continue; }
        fmt.Printf("%6d : ", Vec_IntEntry(vMffc, 0));
        fmt.Printf("Leaf =%3d  ", Vec_IntEntry(vMffc, 1));
        fmt.Printf("Mffc =%4d  ", Vec_IntEntry(vMffc, 2));
        fmt.Printf("Divs =%4d  ", nDivs);
        fmt.Printf("\n");
    }
    fmt.Printf("Collected %d (%.1f %%) MFFCs and %d (%.1f %%) have no divisors (div ave for others is %.2f).\n", Vec_WecSize(vMffcs), 100.0 * Vec_WecSize(vMffcs) / Gia_ManAndNum(p), nDivs0, 100.0 * nDivs0 / Gia_ManAndNum(p), 1.0*nDivsAll/Abc_MaxInt(1, Vec_WecSize(vMffcs) - nDivs0));
    fmt.Printf("Using %.2f MB for MFFCs and %.2f MB for pivots.   ", Vec_WecMemory(vMffcs)/(1<<20), Vec_WecMemory(vPivots)/(1<<20));
}

/**Function*************************************************************

  Synopsis    [Compute divisors and Boolean functions for the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAddDivisors(p *Gia_Man_t, vMffcs *Vec_Wec_t) {
    var vPivots *Vec_Wec_t;
    var vMffc, vPivot, vPivot0, vPivot1 *Vec_Int_t;
    var vCommon, vCommon2, vMap *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, k, iObj, iPivot, iMffc int32;
//abctime clkStart = Abc_Clock();
    // initialize pivots (mapping of nodes into MFFCs whose leaves they are)
    vMap = Vec_IntStartFull( Gia_ManObjNum(p) );
    vPivots = Vec_WecStart( Gia_ManObjNum(p) );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vMffcs)) && (((vMffc) = Vec_WecEntry(vMffcs, i)), 1); i++  {
        assert(Vec_IntSize(vMffc) == 3 + Vec_IntEntry(vMffc, 1), "Vec_IntSize(vMffc) == 3 + Vec_IntEntry(vMffc, 1)");
        iPivot = Vec_IntEntry( vMffc, 0 );
        Vec_IntWriteEntry( vMap, iPivot, i );
        // iterate through the MFFC leaves
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 3; (k < Vec_IntSize(vMffc)) && (((iObj) = Vec_IntEntry(vMffc, k)), 1); k++  {
            vPivot = Vec_WecEntry( vPivots, iObj );
            if Vec_IntSize(vPivot) == 0 {
                Vec_IntGrow(vPivot, 4); }
            Vec_IntPush( vPivot, iPivot );            
        }
    }
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vPivots)) && (((vPivot) = Vec_WecEntry(vPivots, i)), 1); i++  {
        Vec_IntSort( vPivot, 0 );
    }
    // create pivots for internal nodes while growing MFFCs
    vCommon = Vec_IntAlloc( 100 );
    vCommon2 = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        // find commont pivots
        // the slow down happens because some PIs have very large sets of pivots
        vPivot0 = Vec_WecEntry( vPivots, Gia_ObjFaninId0(pObj, i) );
        vPivot1 = Vec_WecEntry( vPivots, Gia_ObjFaninId1(pObj, i) );
        Vec_IntTwoFindCommon( vPivot0, vPivot1, vCommon );
        if Gia_ObjIsMuxId(p, i) {
            vPivot = Vec_WecEntry( vPivots, Gia_ObjFaninId2(p, i) );
            Vec_IntTwoFindCommon( vPivot, vCommon, vCommon2 );
            // inlining of macro ABC_SWAP
            { var t *Vec_Int_t  = vCommon; vCommon = vCommon2; vCommon2 = t; };
        }
        if Vec_IntSize(vCommon) == 0 {
            continue; }
        // add new pivots (this trick increased memory used in vPivots)
        vPivot = Vec_WecEntry( vPivots, i );
        Vec_IntTwoMerge2( vPivot, vCommon, vCommon2 );
        // inlining of macro ABC_SWAP
        { var t Vec_Int_t  = *vPivot; *vPivot = *vCommon2; *vCommon2 = t; };
        // grow MFFCs
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vCommon)) && (((iObj) = Vec_IntEntry(vCommon, k)), 1); k++  {
            iMffc = Vec_IntEntry( vMap, iObj );
            assert(iMffc != -1, "iMffc != -1");
            vMffc = Vec_WecEntry( vMffcs, iMffc );
            Vec_IntPush( vMffc, i );
        }
    }
    }
//Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
    Vec_IntFree( vCommon );
    Vec_IntFree( vCommon2 );
    Vec_IntFree( vMap );
    Gia_ManPrintDivStats( p, vMffcs, vPivots );
    Vec_WecFree( vPivots );
    // returns the modified array of MFFCs
}
func Gia_ManResubTest(p *Gia_Man_t) {
    var vMffcs *Vec_Wec_t;
    var pNew *Gia_Man_t  = Gia_ManDupMuxes( p, 2 );
var clkStart abctime  = Abc_Clock();
    vMffcs = Gia_ManComputeMffcs( pNew, 4, 100, 8, 100 );
    Gia_ManAddDivisors( pNew, vMffcs );
    Vec_WecFree( vMffcs );
Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
    Gia_ManStop( pNew );
}





/**Function*************************************************************

  Synopsis    [Resubstitution data-structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
   
type Gia_ResbMan_t struct {
    nWords         int32;
    nLimit         int32;
    nDivsMax         int32;
    iChoice         int32;
    fUseXor         int32;
    fDebug         int32;
    fVerbose         int32;
    fVeryVerbose         int32;
    vDivs  *Vec_Ptr_t;
    vGates  *Vec_Int_t;
    Vec_Int_t * vUnateLits[2];
    Vec_Int_t * vNotUnateVars[2];
    Vec_Int_t * vUnatePairs[2];
    vBinateVars  *Vec_Int_t;
    Vec_Int_t * vUnateLitsW[2];
    Vec_Int_t * vUnatePairsW[2];
    vSorter  *Vec_Wec_t;
    word *      pSets[2];
    pDivA       *word;
    pDivB       *word;
    vSims  *Vec_Wrd_t;
}
func Gia_ResbAlloc(nWords int32) *Gia_ResbMan_t {
    // inlining of macro ABC_CALLOC
    var p *Gia_ResbMan_t    = &Gia_ResbMan_t{};
    p.nWords           = nWords;
    p.vUnateLits[0]    = Vec_IntAlloc( 100 );
    p.vUnateLits[1]    = Vec_IntAlloc( 100 );
    p.vNotUnateVars[0] = Vec_IntAlloc( 100 );
    p.vNotUnateVars[1] = Vec_IntAlloc( 100 );
    p.vUnatePairs[0]   = Vec_IntAlloc( 100 );
    p.vUnatePairs[1]   = Vec_IntAlloc( 100 );
    p.vUnateLitsW[0]   = Vec_IntAlloc( 100 );
    p.vUnateLitsW[1]   = Vec_IntAlloc( 100 );
    p.vUnatePairsW[0]  = Vec_IntAlloc( 100 );
    p.vUnatePairsW[1]  = Vec_IntAlloc( 100 );
    p.vSorter          = Vec_WecAlloc( nWords*64 );
    p.vBinateVars      = Vec_IntAlloc( 100 );
    p.vGates           = Vec_IntAlloc( 100 );
    p.vDivs            = Vec_PtrAlloc( 100 );
    // inlining of macro ABC_CALLOC
    p.pSets[0]         = ((word *) calloc((size_t)(nWords), sizeof(word)));
    // inlining of macro ABC_CALLOC
    p.pSets[1]         = ((word *) calloc((size_t)(nWords), sizeof(word)));
    // inlining of macro ABC_CALLOC
    p.pDivA            = ((word *) calloc((size_t)(nWords), sizeof(word)));
    // inlining of macro ABC_CALLOC
    p.pDivB            = ((word *) calloc((size_t)(nWords), sizeof(word)));
    p.vSims            = Vec_WrdAlloc( 100 );
    return p;
}
func Gia_ResbInit(p *Gia_ResbMan_t, vDivs *Vec_Ptr_t, nWords int32, nLimit int32, nDivsMax int32, iChoice int32, fUseXor int32, fDebug int32, fVerbose int32, fVeryVerbose int32) {
    assert(p.nWords == nWords, "p.nWords == nWords");
    p.nLimit       = nLimit;
    p.nDivsMax     = nDivsMax;
    p.iChoice      = iChoice;
    p.fUseXor      = fUseXor;
    p.fDebug       = fDebug;
    p.fVerbose     = fVerbose;
    p.fVeryVerbose = fVeryVerbose;
    Abc_TtCopy( p.pSets[0], (word *)Vec_PtrEntry(vDivs, 0), nWords, 0 );
    Abc_TtCopy( p.pSets[1], (word *)Vec_PtrEntry(vDivs, 1), nWords, 0 );
    Vec_PtrClear( p.vDivs );
    Vec_PtrAppend( p.vDivs, vDivs );
    Vec_IntClear( p.vGates );
    Vec_IntClear( p.vUnateLits[0]    );
    Vec_IntClear( p.vUnateLits[1]    );
    Vec_IntClear( p.vNotUnateVars[0] );
    Vec_IntClear( p.vNotUnateVars[1] );
    Vec_IntClear( p.vUnatePairs[0]   );
    Vec_IntClear( p.vUnatePairs[1]   );
    Vec_IntClear( p.vUnateLitsW[0]   );
    Vec_IntClear( p.vUnateLitsW[1]   );
    Vec_IntClear( p.vUnatePairsW[0]  );
    Vec_IntClear( p.vUnatePairsW[1]  );
    Vec_IntClear( p.vBinateVars      );
}
func Gia_ResbFree(p *Gia_ResbMan_t) {
    Vec_IntFree( p.vUnateLits[0]    );
    Vec_IntFree( p.vUnateLits[1]    );
    Vec_IntFree( p.vNotUnateVars[0] );
    Vec_IntFree( p.vNotUnateVars[1] );
    Vec_IntFree( p.vUnatePairs[0]   );
    Vec_IntFree( p.vUnatePairs[1]   );
    Vec_IntFree( p.vUnateLitsW[0]   );
    Vec_IntFree( p.vUnateLitsW[1]   );
    Vec_IntFree( p.vUnatePairsW[0]  );
    Vec_IntFree( p.vUnatePairsW[1]  );
    Vec_IntFree( p.vBinateVars      );
    Vec_IntFree( p.vGates           );
    Vec_WrdFree( p.vSims            );
    Vec_PtrFree( p.vDivs            );
    Vec_WecFree( p.vSorter          );
    // inlining of macro ABC_FREE
    ((p.pSets[0]) ? (free((char *) (p.pSets[0])), (p.pSets[0]) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSets[1]) ? (free((char *) (p.pSets[1])), (p.pSets[1]) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pDivA) ? (free((char *) (p.pDivA)), (p.pDivA) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pDivB) ? (free((char *) (p.pDivB)), (p.pDivB) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Print resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManResubPrintNode(vRes *Vec_Int_t, nVars int32, Node int32, fCompl int32) {
    extern void Gia_ManResubPrintLit( Vec_Int_t * vRes, int nVars, int iLit );
    var iLit0 int32  = Vec_IntEntry( vRes, 2*Node + 0 );
    var iLit1 int32  = Vec_IntEntry( vRes, 2*Node + 1 );
    assert(iLit0 != iLit1, "iLit0 != iLit1");
    if iLit0 > iLit1 && Abc_LitIsCompl(fCompl) { // xor
    
        fmt.Printf("~");
        fCompl = 0;
    }
    fmt.Printf("(");
    Gia_ManResubPrintLit( vRes, nVars, Abc_LitNotCond(iLit0, fCompl) );
    fmt.Printf(" %c ", iLit0 > iLit1 ? '^' : (fCompl ? '|' : '&'));
    Gia_ManResubPrintLit( vRes, nVars, Abc_LitNotCond(iLit1, fCompl) );
    fmt.Printf(")");
}
func Gia_ManResubPrintLit(vRes *Vec_Int_t, nVars int32, iLit int32) {
    if Abc_Lit2Var(iLit) < nVars {
        if nVars < 26 {
            fmt.Printf("%s%c", Abc_LitIsCompl(iLit) ? "~":"", 'a' + Abc_Lit2Var(iLit)-2);
        } else {
            fmt.Printf("%si%d", Abc_LitIsCompl(iLit) ? "~":"", Abc_Lit2Var(iLit)-2); }
    
    } else {
        Gia_ManResubPrintNode( vRes, nVars, Abc_Lit2Var(iLit) - nVars, Abc_LitIsCompl(iLit) ); }
}
func Gia_ManResubPrint(vRes *Vec_Int_t, nVars int32) int32 {
    var iTopLit int32;
    if Vec_IntSize(vRes) == 0 {
        return fmt.Printf("none"); }
    assert(Vec_IntSize(vRes) % 2 == 1, "Vec_IntSize(vRes) % 2 == 1");
    iTopLit = Vec_IntEntryLast(vRes);
    if iTopLit == 0 {
        return fmt.Printf("const0"); }
    if iTopLit == 1 {
        return fmt.Printf("const1"); }
    Gia_ManResubPrintLit( vRes, nVars, iTopLit );
    return 0;
}

/**Function*************************************************************

  Synopsis    [Verify resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManResubVerify(p *Gia_ResbMan_t, pFunc *word) int32 {
    var nVars int32  = Vec_PtrSize(p.vDivs);
    var iTopLit, RetValue int32;
    var pDivRes *word; 
    if Vec_IntSize(p.vGates) == 0 {
        return -1; }
    iTopLit = Vec_IntEntryLast(p.vGates);
    assert(iTopLit >= 0, "iTopLit >= 0");
    if iTopLit == 0 {
        if pFunc != nil { Abc_TtClear( pFunc, p.nWords ); }
        return Abc_TtIsConst0( p.pSets[1], p.nWords );
    }
    if iTopLit == 1 {
        if pFunc != nil { Abc_TtFill( pFunc, p.nWords ); }
        return Abc_TtIsConst0( p.pSets[0], p.nWords );
    }
    if Abc_Lit2Var(iTopLit) < nVars {
        assert(Vec_IntSize(p.vGates) == 1, "Vec_IntSize(p.vGates) == 1");
        pDivRes = (word *)Vec_PtrEntry( p.vDivs, Abc_Lit2Var(iTopLit) );
    
    } else {
        var i, iLit0, iLit1 int32;
        assert(Vec_IntSize(p.vGates) > 1, "Vec_IntSize(p.vGates) > 1");
        assert(Vec_IntSize(p.vGates) % 2 == 1, "Vec_IntSize(p.vGates) % 2 == 1");
        assert(Abc_Lit2Var(iTopLit)-nVars == Vec_IntSize(p.vGates)/2-1, "Abc_Lit2Var(iTopLit)-nVars == Vec_IntSize(p.vGates)/2-1");
        Vec_WrdFill( p.vSims, p.nWords * Vec_IntSize(p.vGates)/2, 0 );
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(p.vGates)) && (((iLit0) = Vec_IntEntry(p.vGates, i)), 1) && (((iLit1) = Vec_IntEntry(p.vGates, i+1)), 1); i += 2  {
            var iVar0 int32  = Abc_Lit2Var(iLit0);
            var iVar1 int32  = Abc_Lit2Var(iLit1);
            var pDiv0 *word;
            if iVar0 < nVars {
                pDiv0 = (word *)Vec_PtrEntry(p.vDivs, iVar0);
            } else {
                pDiv0 = Vec_WrdEntryP(p.vSims, p.nWords*(iVar0 - nVars));
            }
            var pDiv1 *word;
            if iVar1 < nVars {
                pDiv1 = (word *)Vec_PtrEntry(p.vDivs, iVar1);
            } else {
                pDiv1 = Vec_WrdEntryP(p.vSims, p.nWords*(iVar1 - nVars));
            }
            var pDiv *word   = Vec_WrdEntryP(p.vSims, p.nWords*i/2);
            if iVar0 < iVar1 {
                Abc_TtAndCompl( pDiv, pDiv0, Abc_LitIsCompl(iLit0), pDiv1, Abc_LitIsCompl(iLit1), p.nWords );
            } else if iVar0 > iVar1 {
                assert(!Abc_LitIsCompl(iLit0), "!Abc_LitIsCompl(iLit0)");
                assert(!Abc_LitIsCompl(iLit1), "!Abc_LitIsCompl(iLit1)");
                Abc_TtXor( pDiv, pDiv0, pDiv1, p.nWords, 0 );
            
            } else { assert(false, "0"); }
        }
        pDivRes = Vec_WrdEntryP( p.vSims, p.nWords*(Vec_IntSize(p.vGates)/2-1) );
    }
    if Abc_LitIsCompl(iTopLit) {
        RetValue = !Abc_TtIntersectOne(p.pSets[1], 0, pDivRes, 0, p.nWords) && !Abc_TtIntersectOne(p.pSets[0], 0, pDivRes, 1, p.nWords);
    } else {
        RetValue = !Abc_TtIntersectOne(p.pSets[0], 0, pDivRes, 0, p.nWords) && !Abc_TtIntersectOne(p.pSets[1], 0, pDivRes, 1, p.nWords); }
    if pFunc != nil { Abc_TtCopy( pFunc, pDivRes, p.nWords, Abc_LitIsCompl(iTopLit) ); }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Construct AIG manager from gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManConstructFromMap(pNew *Gia_Man_t, vGates *Vec_Int_t, nVars int32, vUsed *Vec_Int_t, vCopy *Vec_Int_t, fHash int32) int32 {
    var i int32;
    var iLit0 int32;
    var iLit1 int32;
    var iLitRes int32;
    var iTopLit int32 = Vec_IntEntryLast( vGates );
    assert(Vec_IntSize(vUsed) == nVars, "Vec_IntSize(vUsed) == nVars");
    assert(Vec_IntSize(vGates) > 1, "Vec_IntSize(vGates) > 1");
    assert(Vec_IntSize(vGates) % 2 == 1, "Vec_IntSize(vGates) % 2 == 1");
    assert(Abc_Lit2Var(iTopLit)-nVars == Vec_IntSize(vGates)/2-1, "Abc_Lit2Var(iTopLit)-nVars == Vec_IntSize(vGates)/2-1");
    Vec_IntClear( vCopy );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vGates)) && (((iLit0) = Vec_IntEntry(vGates, i)), 1) && (((iLit1) = Vec_IntEntry(vGates, i+1)), 1); i += 2  {
        var iVar0 int32  = Abc_Lit2Var(iLit0);
        var iVar1 int32  = Abc_Lit2Var(iLit1);
        var iRes0 int32;
        if iVar0 < nVars {
            iRes0 = Vec_IntEntry(vUsed, iVar0);
        } else {
            iRes0 = Vec_IntEntry(vCopy, iVar0 - nVars);
        }
        var iRes1 int32;
        if iVar1 < nVars {
            iRes1 = Vec_IntEntry(vUsed, iVar1);
        } else {
            iRes1 = Vec_IntEntry(vCopy, iVar1 - nVars);
        }
        if iVar0 < iVar1 {
            if fHash != 0 {
                iLitRes = Gia_ManHashAnd( pNew, Abc_LitNotCond(iRes0, Abc_LitIsCompl(iLit0)), Abc_LitNotCond(iRes1, Abc_LitIsCompl(iLit1)) );
            } else {
                iLitRes = Gia_ManAppendAnd( pNew, Abc_LitNotCond(iRes0, Abc_LitIsCompl(iLit0)), Abc_LitNotCond(iRes1, Abc_LitIsCompl(iLit1)) ); }
        
        } else if iVar0 > iVar1 {
            assert(!Abc_LitIsCompl(iLit0), "!Abc_LitIsCompl(iLit0)");
            assert(!Abc_LitIsCompl(iLit1), "!Abc_LitIsCompl(iLit1)");
            if fHash != 0 {
                iLitRes = Gia_ManHashXor( pNew, Abc_LitNotCond(iRes0, Abc_LitIsCompl(iLit0)), Abc_LitNotCond(iRes1, Abc_LitIsCompl(iLit1)) );
            } else {
                iLitRes = Gia_ManAppendXor( pNew, Abc_LitNotCond(iRes0, Abc_LitIsCompl(iLit0)), Abc_LitNotCond(iRes1, Abc_LitIsCompl(iLit1)) ); }
        
        } else { assert(false, "0"); }
        Vec_IntPush( vCopy, iLitRes );
    }
    assert(Vec_IntSize(vCopy) == Vec_IntSize(vGates)/2, "Vec_IntSize(vCopy) == Vec_IntSize(vGates)/2");
    iLitRes = Vec_IntEntry( vCopy, Vec_IntSize(vGates)/2-1 );
    return iLitRes;
}
func Gia_ManConstructFromGates(vFuncs *Vec_Wec_t, nVars int32) *Gia_Man_t {
    var vGates *Vec_Int_t; var i, k, iLit int32;
    var vCopy *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vUsed *Vec_Int_t  = Vec_IntStartFull( nVars );
    var pNew *Gia_Man_t  = Gia_ManStart( 100 );
    pNew.pName = Abc_UtilStrsav( "resub" );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vFuncs)) && (((vGates) = Vec_WecEntry(vFuncs, i)), 1); i++  {
        assert(Vec_IntSize(vGates) % 2 == 1, "Vec_IntSize(vGates) % 2 == 1");
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vGates)) && (((iLit) = Vec_IntEntry(vGates, k)), 1); k++  {
            var iVar int32  = Abc_Lit2Var(iLit);
            if iVar > 0 && iVar < nVars && Vec_IntEntry(vUsed, iVar) == -1 {
                Vec_IntWriteEntry( vUsed, iVar, Gia_ManAppendCi(pNew) ); }
        }
    }
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vFuncs)) && (((vGates) = Vec_WecEntry(vFuncs, i)), 1); i++  {
        var iLitRes int32;
        var iTopLit int32 = Vec_IntEntryLast( vGates );
        if Abc_Lit2Var(iTopLit) == 0 {
            iLitRes = 0;
        } else if Abc_Lit2Var(iTopLit) < nVars {
            iLitRes = Gia_ManAppendCi(pNew);
        } else {
            iLitRes = Gia_ManConstructFromMap( pNew, vGates, nVars, vUsed, vCopy, 0 ); }
        Gia_ManAppendCo( pNew, Abc_LitNotCond( iLitRes, Abc_LitIsCompl(iTopLit) ) );
    }
    Vec_IntFree( vCopy );
    Vec_IntFree( vUsed );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Construct AIG manager from gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManInsertOrder_rec(p *Gia_Man_t, iObj int32, vObjs *Vec_Int_t, vFuncs *Vec_Wec_t, vNodes *Vec_Int_t) {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if iObj == 0 {
        return; }
    if pObj.fPhase {
        var nVars int32  = Gia_ManObjNum(p);
        var k int32;
        var iLit int32;
        var Index int32 = Vec_IntFind( vObjs, iObj );
        var vGates *Vec_Int_t  = Vec_WecEntry( vFuncs, Index );
        assert(Gia_ObjIsCo(pObj) || Gia_ObjIsAnd(pObj), "Gia_ObjIsCo(pObj) || Gia_ObjIsAnd(pObj)");
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vGates)) && (((iLit) = Vec_IntEntry(vGates, k)), 1); k++  {
            if Abc_Lit2Var(iLit) < nVars {
                Gia_ManInsertOrder_rec( p, Abc_Lit2Var(iLit), vObjs, vFuncs, vNodes ); }
        }
    
    } else if Gia_ObjIsCo(pObj) {
        Gia_ManInsertOrder_rec( p, Gia_ObjFaninId0p(p, pObj), vObjs, vFuncs, vNodes );
    } else if Gia_ObjIsAnd(pObj) {
        Gia_ManInsertOrder_rec( p, Gia_ObjFaninId0p(p, pObj), vObjs, vFuncs, vNodes );
        Gia_ManInsertOrder_rec( p, Gia_ObjFaninId1p(p, pObj), vObjs, vFuncs, vNodes );
    
    } else { assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)"); }
    if !Gia_ObjIsCi(pObj) {
        Vec_IntPush( vNodes, iObj ); }
}
func Gia_ManInsertOrder(p *Gia_Man_t, vObjs *Vec_Int_t, vFuncs *Vec_Wec_t) *Vec_Int_t {
    var i, Id int32;
    var vNodes *Vec_Int_t  = Vec_IntAlloc( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
        Gia_ManInsertOrder_rec( p, Id, vObjs, vFuncs, vNodes );
    }
    return vNodes;
}
func Gia_ManInsertFromGates(p *Gia_Man_t, vObjs *Vec_Int_t, vFuncs *Vec_Wec_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t; 
    var pObj *Gia_Obj_t; 
    var i int32;
    var nVars int32 = Gia_ManObjNum(p);
    var vUsed *Vec_Int_t  = Vec_IntStartFull( nVars );
    var vNodes *Vec_Int_t;
    var vCopy *Vec_Int_t = Vec_IntAlloc(100);
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        pObj.fPhase = 1;
    }
    vNodes = Gia_ManInsertOrder( p, vObjs, vFuncs );
    pNew = Gia_ManStart( Gia_ManObjNum(p) + 1000 );
    Gia_ManHashStart( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
        if !pObj.fPhase {
            if Gia_ObjIsCo(pObj) {
                pObj.Value = Gia_ObjFanin0Copy(pObj);
            } else if Gia_ObjIsAnd(pObj) {            
                pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
            } else { assert(false, "0"); }
        
        } else {
            var k int32;
            var iLit int32;
            var Index int32 = Vec_IntFind( vObjs, Gia_ObjId(p, pObj) );
            var vGates *Vec_Int_t  = Vec_WecEntry( vFuncs, Index );
            var iLitRes int32;
            var iTopLit int32 = Vec_IntEntryLast( vGates );
            if Abc_Lit2Var(iTopLit) == 0 {
                iLitRes = 0;
            } else if Abc_Lit2Var(iTopLit) < nVars {
                iLitRes = Gia_ManObj(p, Abc_Lit2Var(iTopLit)).Value;
            } else {
                // inlining of macro Vec_IntForEachEntry
                for  k = 0; (k < Vec_IntSize(vGates)) && (((iLit) = Vec_IntEntry(vGates, k)), 1); k++  {
                    Vec_IntWriteEntry( vUsed, Abc_Lit2Var(iLit), Gia_ManObj(p, Abc_Lit2Var(iLit)).Value );
                }
                iLitRes = Gia_ManConstructFromMap( pNew, vGates, nVars, vUsed, vCopy, 1 );
                // inlining of macro Vec_IntForEachEntry
                for  k = 0; (k < Vec_IntSize(vGates)) && (((iLit) = Vec_IntEntry(vGates, k)), 1); k++  {
                    Vec_IntWriteEntry( vUsed, Abc_Lit2Var(iLit), -1 );
                }
            }
            pObj.Value = Abc_LitNotCond( iLitRes, Abc_LitIsCompl(iTopLit) );
        }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, pObj.Value );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        pObj.fPhase = 0;
    }
    Gia_ManHashStop( pNew );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    Vec_IntFree( vNodes );
    Vec_IntFree( vUsed );
    Vec_IntFree( vCopy );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Perform resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFindFirstCommonLit(vArr1 *Vec_Int_t, vArr2 *Vec_Int_t, fVerbose int32) int32 {
    var pBeg1 *int32  = vArr1.pArray;
    var pBeg2 *int32  = vArr2.pArray;
    var pEnd1 *int32  = vArr1.pArray + vArr1.nSize;
    var pEnd2 *int32  = vArr2.pArray + vArr2.nSize;
    var pStart1 *int32  = vArr1.pArray;
    var pStart2 *int32  = vArr2.pArray;
    var nRemoved int32;
    for  pBeg1 < pEnd1 && pBeg2 < pEnd2  {
        if Abc_Lit2Var(*pBeg1) == Abc_Lit2Var(*pBeg2) { 
            if *pBeg1 != *pBeg2 { 
                return *pBeg1; 
            } else {
                pBeg1++, pBeg2++; }
            nRemoved++;
        
        } else if *pBeg1 < *pBeg2 {
            *pStart1++ = *pBeg1++;
        } else { 
            *pStart2++ = *pBeg2++; }
    }
    for  pBeg1 < pEnd1  {
        *pStart1++ = *pBeg1++;
    }
    for  pBeg2 < pEnd2  {
        *pStart2++ = *pBeg2++;
    }
    Vec_IntShrink( vArr1, pStart1 - vArr1.pArray );
    Vec_IntShrink( vArr2, pStart2 - vArr2.pArray );
    //if ( fVerbose ) printf( "Removed %d duplicated entries.  Array1 = %d.  Array2 = %d.\n", nRemoved, Vec_IntSize(vArr1), Vec_IntSize(vArr2) );
    return -1;
}

func Gia_ManFindOneUnateInt(pOff *word, pOn *word, vDivs *Vec_Ptr_t, nWords int32, vUnateLits *Vec_Int_t, vNotUnateVars *Vec_Int_t) {
    var pDiv *word; var i int32;
    Vec_IntClear( vUnateLits );
    Vec_IntClear( vNotUnateVars );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 2; (i < Vec_PtrSize(vDivs)) && (((pDiv) = (word *)Vec_PtrEntry(vDivs, i)), 1); i++  {
        if !Abc_TtIntersectOne( pOff, 0, pDiv, 0, nWords ) {
            Vec_IntPush( vUnateLits, Abc_Var2Lit(i, 0) );
        } else if !Abc_TtIntersectOne( pOff, 0, pDiv, 1, nWords ) {
            Vec_IntPush( vUnateLits, Abc_Var2Lit(i, 1) );
        } else {
            Vec_IntPush( vNotUnateVars, i ); }
    }
}
func Gia_ManFindOneUnate(pSets []*word, vDivs *Vec_Ptr_t, nWords int32, vUnateLits []*Vec_Int_t, vNotUnateVars []*Vec_Int_t, fVerbose int32) int32 {
    var n int32;
    if fVerbose != 0 { fmt.Printf("  "); }
    for  n = 0; n < 2; n++  {
        Gia_ManFindOneUnateInt( pSets[n], pSets[!n], vDivs, nWords, vUnateLits[n], vNotUnateVars[n] );
        if fVerbose != 0 { fmt.Printf("U%d =%4d ", n, Vec_IntSize(vUnateLits[n])); }
    }
    return Gia_ManFindFirstCommonLit( vUnateLits[0], vUnateLits[1], fVerbose );
}

func Gia_ManDivCover(pOff *word, pOn *word, pDivA *word, ComplA int32, pDivB *word, ComplB int32, nWords int32) int32 {
    //assert( !Abc_TtIntersectOne(pOff, 0, pDivA, ComplA, nWords) );
    //assert( !Abc_TtIntersectOne(pOff, 0, pDivB, ComplB, nWords) );
    return !Abc_TtIntersectTwo( pOn, 0, pDivA, !ComplA, pDivB, !ComplB, nWords );
}
func Gia_ManFindTwoUnateInt(pOff *word, pOn *word, vDivs *Vec_Ptr_t, nWords int32, vUnateLits *Vec_Int_t, vUnateLitsW *Vec_Int_t, pnPairs *int32) int32 {
    var i, k, iDiv0_, iDiv1_, Cover0, Cover1 int32;
    var nTotal int32  = Abc_TtCountOnesVec( pOn, nWords );
    (*pnPairs) = 0;
    // inlining of macro Vec_IntForEachEntryTwo
    for  i = 0; (i < Vec_IntSize(vUnateLits)) && (((iDiv0_) = Vec_IntEntry(vUnateLits, i)), 1) && (((Cover0) = Vec_IntEntry(vUnateLitsW, i)), 1); i++  {
        if 2*Cover0 < nTotal {
            break; }
        // inlining of macro Vec_IntForEachEntryTwoStart
        for  k = i+1; (k < Vec_IntSize(vUnateLits)) && (((iDiv1_) = Vec_IntEntry(vUnateLits, k)), 1) && (((Cover1) = Vec_IntEntry(vUnateLitsW, k)), 1); k++  {
            var iDiv0 int32  = Abc_MinInt( iDiv0_, iDiv1_ );
            var iDiv1 int32  = Abc_MaxInt( iDiv0_, iDiv1_ );
            var pDiv0 *word  = (word *)Vec_PtrEntry(vDivs, Abc_Lit2Var(iDiv0));
            var pDiv1 *word  = (word *)Vec_PtrEntry(vDivs, Abc_Lit2Var(iDiv1));
            if Cover0 + Cover1 < nTotal {
                break; }
            (*pnPairs)++;
            if Gia_ManDivCover(pOff, pOn, pDiv1, Abc_LitIsCompl(iDiv1), pDiv0, Abc_LitIsCompl(iDiv0), nWords) != 0 {
                return Abc_Var2Lit((Abc_LitNot(iDiv1) << 15) | Abc_LitNot(iDiv0), 1); }
        }
    }
    return -1;
}
func Gia_ManFindTwoUnate(pSets []*word, vDivs *Vec_Ptr_t, nWords int32, vUnateLits []*Vec_Int_t, vUnateLitsW []*Vec_Int_t, fVerbose int32) int32 {
    var n, iLit, nPairs int32;
    if fVerbose != 0 { fmt.Printf("  "); }
    for  n = 0; n < 2; n++  {
        //int nPairsAll = Vec_IntSize(vUnateLits[n])*(Vec_IntSize(vUnateLits[n])-1)/2;
        iLit = Gia_ManFindTwoUnateInt( pSets[n], pSets[!n], vDivs, nWords, vUnateLits[n], vUnateLitsW[n], &nPairs );
        if fVerbose != 0 { fmt.Printf("UU%d =%5d ", n, nPairs); }
        if iLit >= 0 {
            return Abc_LitNotCond(iLit, n); }
    }
    return -1;
}

func Gia_ManFindXorInt(pOff *word, pOn *word, vBinate *Vec_Int_t, vDivs *Vec_Ptr_t, nWords int32, vUnatePairs *Vec_Int_t) {
    var i, k, iDiv0_, iDiv1_ int32;
    var Limit2 int32  = Vec_IntSize(vBinate);//Abc_MinInt( Vec_IntSize(vBinate), 100 );
    // inlining of macro Vec_IntForEachEntryStop
    for  i = 0; (i < Limit2) && (((iDiv1_) = Vec_IntEntry(vBinate, i)), 1); i++  {
    // inlining of macro Vec_IntForEachEntryStop
    for  k = 0; (k < i) && (((iDiv0_) = Vec_IntEntry(vBinate, k)), 1); k++  {
        var iDiv0 int32  = Abc_MinInt( iDiv0_, iDiv1_ );
        var iDiv1 int32  = Abc_MaxInt( iDiv0_, iDiv1_ );
        var pDiv0 *word  = (word *)Vec_PtrEntry(vDivs, iDiv0);
        var pDiv1 *word  = (word *)Vec_PtrEntry(vDivs, iDiv1);
        if !Abc_TtIntersectXor( pOff, 0, pDiv0, pDiv1, 0, nWords ) {
            Vec_IntPush( vUnatePairs, Abc_Var2Lit((Abc_Var2Lit(iDiv0, 0) << 15) | Abc_Var2Lit(iDiv1, 0), 0) );
        } else if !Abc_TtIntersectXor( pOff, 0, pDiv0, pDiv1, 1, nWords ) {
            Vec_IntPush( vUnatePairs, Abc_Var2Lit((Abc_Var2Lit(iDiv0, 0) << 15) | Abc_Var2Lit(iDiv1, 0), 1) ); }
    }
    }
}
func Gia_ManFindXor(pSets []*word, vDivs *Vec_Ptr_t, nWords int32, vBinateVars *Vec_Int_t, vUnatePairs []*Vec_Int_t, fVerbose int32) int32 {
    var n int32;
    if fVerbose != 0 { fmt.Printf("  "); }
    for  n = 0; n < 2; n++  {
        Vec_IntClear( vUnatePairs[n] );
        Gia_ManFindXorInt( pSets[n], pSets[!n], vBinateVars, vDivs, nWords, vUnatePairs[n] );
        if fVerbose != 0 { fmt.Printf("UX%d =%5d ", n, Vec_IntSize(vUnatePairs[n])); }
    }
    return Gia_ManFindFirstCommonLit( vUnatePairs[0], vUnatePairs[1], fVerbose );
}

func Gia_ManFindUnatePairsInt(pOff *word, pOn *word, vBinate *Vec_Int_t, vDivs *Vec_Ptr_t, nWords int32, vUnatePairs *Vec_Int_t) {
    var n, i, k, iDiv0_, iDiv1_ int32;
    var Limit2 int32  = Vec_IntSize(vBinate);//Abc_MinInt( Vec_IntSize(vBinate), 100 );
    // inlining of macro Vec_IntForEachEntryStop
    for  i = 0; (i < Limit2) && (((iDiv1_) = Vec_IntEntry(vBinate, i)), 1); i++  {
    // inlining of macro Vec_IntForEachEntryStop
    for  k = 0; (k < i) && (((iDiv0_) = Vec_IntEntry(vBinate, k)), 1); k++  {
        var iDiv0 int32  = Abc_MinInt( iDiv0_, iDiv1_ );
        var iDiv1 int32  = Abc_MaxInt( iDiv0_, iDiv1_ );
        var pDiv0 *word  = (word *)Vec_PtrEntry(vDivs, iDiv0);
        var pDiv1 *word  = (word *)Vec_PtrEntry(vDivs, iDiv1);
        for  n = 0; n < 4; n++  {
            var iLit0 int32  = Abc_Var2Lit( iDiv0, n&1 );
            var iLit1 int32  = Abc_Var2Lit( iDiv1, n>>1 );
            //if ( !Abc_TtIntersectTwo( pOff, 0, pDiv1, n>>1, pDiv0, n&1, nWords ) )
            if !Abc_TtIntersectTwo( pOff, 0, pDiv1, n>>1, pDiv0, n&1, nWords ) && Abc_TtIntersectTwo( pOn, 0, pDiv1, n>>1, pDiv0, n&1, nWords ) {
                Vec_IntPush( vUnatePairs, Abc_Var2Lit((iLit1 << 15) | iLit0, 0) ); }
        }
    }
    }
}
func Gia_ManFindUnatePairs(pSets []*word, vDivs *Vec_Ptr_t, nWords int32, vBinateVars *Vec_Int_t, vUnatePairs []*Vec_Int_t, fVerbose int32) {
    var n, RetValue int32;
    if fVerbose != 0 { fmt.Printf("  "); }
    for  n = 0; n < 2; n++  {
        var nBefore int32  = Vec_IntSize(vUnatePairs[n]);
        Gia_ManFindUnatePairsInt( pSets[n], pSets[!n], vBinateVars, vDivs, nWords, vUnatePairs[n] );
        if fVerbose != 0 { fmt.Printf("UP%d =%5d ", n, Vec_IntSize(vUnatePairs[n])-nBefore); }
    }
    RetValue = Gia_ManFindFirstCommonLit( vUnatePairs[0], vUnatePairs[1], fVerbose );
    assert(RetValue == -1, "RetValue == -1");
}

func Gia_ManDeriveDivPair(iDiv int32, vDivs *Vec_Ptr_t, nWords int32, pRes *word) {
    var fComp int32  = Abc_LitIsCompl(iDiv);
    var iDiv0 int32  = Abc_Lit2Var(iDiv) & 0x7FFF;
    var iDiv1 int32  = Abc_Lit2Var(iDiv) >> 15;
    var pDiv0 *word  = (word *)Vec_PtrEntry(vDivs, Abc_Lit2Var(iDiv0));
    var pDiv1 *word  = (word *)Vec_PtrEntry(vDivs, Abc_Lit2Var(iDiv1));
    if iDiv0 < iDiv1 {
        assert(!(fComp != 0), "!fComp");
        Abc_TtAndCompl( pRes, pDiv0, Abc_LitIsCompl(iDiv0), pDiv1, Abc_LitIsCompl(iDiv1), nWords );
    
    } else {
        assert(!Abc_LitIsCompl(iDiv0), "!Abc_LitIsCompl(iDiv0)");
        assert(!Abc_LitIsCompl(iDiv1), "!Abc_LitIsCompl(iDiv1)");
        Abc_TtXor( pRes, pDiv0, pDiv1, nWords, 0 );
    }
}
func Gia_ManFindDivGateInt(pOff *word, pOn *word, vDivs *Vec_Ptr_t, nWords int32, vUnateLits *Vec_Int_t, vUnatePairs *Vec_Int_t, vUnateLitsW *Vec_Int_t, vUnatePairsW *Vec_Int_t, pDivTemp *word) int32 {
    var i, k, iDiv0, iDiv1, Cover0, Cover1 int32;
    var nTotal int32  = Abc_TtCountOnesVec( pOn, nWords );
    // inlining of macro Vec_IntForEachEntryTwo
    for  i = 0; (i < Vec_IntSize(vUnateLits)) && (((iDiv0) = Vec_IntEntry(vUnateLits, i)), 1) && (((Cover0) = Vec_IntEntry(vUnateLitsW, i)), 1); i++  {
        var pDiv0 *word  = (word *)Vec_PtrEntry(vDivs, Abc_Lit2Var(iDiv0));
        if 2*Cover0 < nTotal {
            break; }
        // inlining of macro Vec_IntForEachEntryTwo
        for  k = 0; (k < Vec_IntSize(vUnatePairs)) && (((iDiv1) = Vec_IntEntry(vUnatePairs, k)), 1) && (((Cover1) = Vec_IntEntry(vUnatePairsW, k)), 1); k++  {
            var fComp1 int32  = Abc_LitIsCompl(iDiv1);
            if Cover0 + Cover1 < nTotal {
                break; }
            Gia_ManDeriveDivPair( iDiv1, vDivs, nWords, pDivTemp );
            if Gia_ManDivCover(pOff, pOn, pDiv0, Abc_LitIsCompl(iDiv0), pDivTemp, fComp1, nWords) != 0 {
                return Abc_Var2Lit((Abc_Var2Lit(k, 1) << 15) | Abc_LitNot(iDiv0), 1); }
        }
    }
    return -1;
}
func Gia_ManFindDivGate(pSets []*word, vDivs *Vec_Ptr_t, nWords int32, vUnateLits []*Vec_Int_t, vUnatePairs []*Vec_Int_t, vUnateLitsW []*Vec_Int_t, vUnatePairsW []*Vec_Int_t, pDivTemp *word) int32 {
    var n, iLit int32;
    for  n = 0; n < 2; n++  {
        iLit = Gia_ManFindDivGateInt( pSets[n], pSets[!n], vDivs, nWords, vUnateLits[n], vUnatePairs[n], vUnateLitsW[n], vUnatePairsW[n], pDivTemp );
        if iLit >= 0 {
            return Abc_LitNotCond( iLit, n ); }
    }
    return -1;
}

func Gia_ManFindGateGateInt(pOff *word, pOn *word, vDivs *Vec_Ptr_t, nWords int32, vUnatePairs *Vec_Int_t, vUnatePairsW *Vec_Int_t, pDivTempA *word, pDivTempB *word) int32 {
    var i, k, iDiv0, iDiv1, Cover0, Cover1 int32;
    var nTotal int32  = Abc_TtCountOnesVec( pOn, nWords );
    // inlining of macro Vec_IntForEachEntryTwo
    for  k = 0; (k < Vec_IntSize(vUnatePairs)) && (((iDiv0) = Vec_IntEntry(vUnatePairs, k)), 1) && (((Cover0) = Vec_IntEntry(vUnatePairsW, k)), 1); k++  {
        var fCompA int32  = Abc_LitIsCompl(iDiv0);
        if 2*Cover0 < nTotal {
            break; }
        Gia_ManDeriveDivPair( iDiv0, vDivs, nWords, pDivTempA );
        // inlining of macro Vec_IntForEachEntryTwoStart
        for  i = k+1; (i < Vec_IntSize(vUnatePairs)) && (((iDiv1) = Vec_IntEntry(vUnatePairs, i)), 1) && (((Cover1) = Vec_IntEntry(vUnatePairsW, i)), 1); i++  {
            var fCompB int32  = Abc_LitIsCompl(iDiv1);
            if Cover0 + Cover1 < nTotal {
                break; }
            Gia_ManDeriveDivPair( iDiv1, vDivs, nWords, pDivTempB );
            if Gia_ManDivCover(pOff, pOn, pDivTempA, fCompA, pDivTempB, fCompB, nWords) != 0 {
                return Abc_Var2Lit((Abc_Var2Lit(i, 1) << 15) | Abc_Var2Lit(k, 1), 1); }
        }
    }
    return -1;
}
func Gia_ManFindGateGate(pSets []*word, vDivs *Vec_Ptr_t, nWords int32, vUnatePairs []*Vec_Int_t, vUnatePairsW []*Vec_Int_t, pDivTempA *word, pDivTempB *word) int32 {
    var n, iLit int32;
    for  n = 0; n < 2; n++  {
        iLit = Gia_ManFindGateGateInt( pSets[n], pSets[!n], vDivs, nWords, vUnatePairs[n], vUnatePairsW[n], pDivTempA, pDivTempB );
        if iLit >= 0 {
            return Abc_LitNotCond( iLit, n ); }
    }
    return -1;
}

func Gia_ManSortUnatesInt(pOff *word, pOn *word, vDivs *Vec_Ptr_t, nWords int32, vUnateLits *Vec_Int_t, vUnateLitsW *Vec_Int_t, vSorter *Vec_Wec_t) {
    var i, k, iLit int32;
    var vLevel *Vec_Int_t;
    Vec_WecInit( vSorter, nWords*64 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vUnateLits)) && (((iLit) = Vec_IntEntry(vUnateLits, i)), 1); i++  {
        var pDiv *word  = (word *)Vec_PtrEntry(vDivs, Abc_Lit2Var(iLit));
        //assert( !Abc_TtIntersectOne( pOff, 0, pDiv, Abc_LitIsCompl(iLit), nWords ) );
        Vec_WecPush( vSorter, Abc_TtCountOnesVecMask(pDiv, pOn, nWords, Abc_LitIsCompl(iLit)), iLit );
    }
    Vec_IntClear( vUnateLits );
    Vec_IntClear( vUnateLitsW );
    // inlining of macro Vec_WecForEachLevelReverse
    for  k = Vec_WecSize(vSorter)-1; (k >= 0) && (((vLevel) = Vec_WecEntry(vSorter, k)), 1); k--  {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLevel)) && (((iLit) = Vec_IntEntry(vLevel, i)), 1); i++  {
            Vec_IntPush( vUnateLits, iLit );
            Vec_IntPush( vUnateLitsW, k );
        }
    }
    //Vec_IntPrint( Vec_WecEntry(vSorter, 0) );
    Vec_WecClear( vSorter );
}
func Gia_ManSortUnates(pSets []*word, vDivs *Vec_Ptr_t, nWords int32, vUnateLits []*Vec_Int_t, vUnateLitsW []*Vec_Int_t, vSorter *Vec_Wec_t) {
    var n int32;
    for  n = 0; n < 2; n++  {
        Gia_ManSortUnatesInt( pSets[n], pSets[!n], vDivs, nWords, vUnateLits[n], vUnateLitsW[n], vSorter );
    }
}

func Gia_ManSortPairsInt(pOff *word, pOn *word, vDivs *Vec_Ptr_t, nWords int32, vUnatePairs *Vec_Int_t, vUnatePairsW *Vec_Int_t, vSorter *Vec_Wec_t) {
    var i, k, iPair int32;
    var vLevel *Vec_Int_t;
    Vec_WecInit( vSorter, nWords*64 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vUnatePairs)) && (((iPair) = Vec_IntEntry(vUnatePairs, i)), 1); i++  {
        var fComp int32  = Abc_LitIsCompl(iPair);
        var iLit0 int32  = Abc_Lit2Var(iPair) & 0x7FFF;
        var iLit1 int32  = Abc_Lit2Var(iPair) >> 15;
        var pDiv0 *word  = (word *)Vec_PtrEntry( vDivs, Abc_Lit2Var(iLit0) );
        var pDiv1 *word  = (word *)Vec_PtrEntry( vDivs, Abc_Lit2Var(iLit1) );
        if iLit0 < iLit1 {
            assert(!(fComp != 0), "!fComp");
            //assert( !Abc_TtIntersectTwo( pOff, 0, pDiv0, Abc_LitIsCompl(iLit0), pDiv1, Abc_LitIsCompl(iLit1), nWords ) );
            Vec_WecPush( vSorter, Abc_TtCountOnesVecMask2(pDiv0, pDiv1, Abc_LitIsCompl(iLit0), Abc_LitIsCompl(iLit1), pOn, nWords), iPair );
        
        } else {
            assert(!Abc_LitIsCompl(iLit0), "!Abc_LitIsCompl(iLit0)");
            assert(!Abc_LitIsCompl(iLit1), "!Abc_LitIsCompl(iLit1)");
            //assert( !Abc_TtIntersectXor( pOff, 0, pDiv0, pDiv1, fComp, nWords ) );
            Vec_WecPush( vSorter, Abc_TtCountOnesVecXorMask(pDiv0, pDiv1, fComp, pOn, nWords), iPair );
        }
    }
    Vec_IntClear( vUnatePairs );
    Vec_IntClear( vUnatePairsW );
    // inlining of macro Vec_WecForEachLevelReverse
    for  k = Vec_WecSize(vSorter)-1; (k >= 0) && (((vLevel) = Vec_WecEntry(vSorter, k)), 1); k--  {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLevel)) && (((iPair) = Vec_IntEntry(vLevel, i)), 1); i++  {
            Vec_IntPush( vUnatePairs, iPair );
            Vec_IntPush( vUnatePairsW, k );
        }
    }
    //Vec_IntPrint( Vec_WecEntry(vSorter, 0) );
    Vec_WecClear( vSorter );

}
func Gia_ManSortPairs(pSets []*word, vDivs *Vec_Ptr_t, nWords int32, vUnateLits []*Vec_Int_t, vUnateLitsW []*Vec_Int_t, vSorter *Vec_Wec_t) {
    var n int32;
    for  n = 0; n < 2; n++  {
        Gia_ManSortPairsInt( pSets[n], pSets[!n], vDivs, nWords, vUnateLits[n], vUnateLitsW[n], vSorter );
    }
}

func Gia_ManSortBinate(pSets []*word, vDivs *Vec_Ptr_t, nWords int32, vBinateVars *Vec_Int_t, vSorter *Vec_Wec_t) {
    var vLevel *Vec_Int_t;
    var nMints [2]int32 = [2]int32{ Abc_TtCountOnesVec(pSets[0], nWords), Abc_TtCountOnesVec(pSets[1], nWords) };
    var pBig *word;
    if nMints[0] > nMints[1] {
        pBig = pSets[0];
    } else {
        pBig = pSets[1];
    }
    var pSmo *word;
    if nMints[0] > nMints[1] {
        pSmo = pSets[1];
    } else {
        pSmo = pSets[0];
    }
    var Big int32  = Abc_MaxInt( nMints[0], nMints[1] );
    var Smo int32  = Abc_MinInt( nMints[0], nMints[1] );
    var i, k, iDiv, Gain int32;

    Vec_WecInit( vSorter, nWords*64 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vBinateVars)) && (((iDiv) = Vec_IntEntry(vBinateVars, i)), 1); i++  {
        var pDiv *word  = (word *)Vec_PtrEntry( vDivs, iDiv );
        var nInter [2]int32 = [2]int32{ Abc_TtCountOnesVecMask(pBig, pDiv, nWords, 0), Abc_TtCountOnesVecMask(pSmo, pDiv, nWords, 0) };
        if nInter[0] < Big/2 { // complement the divisor
        
            nInter[0] = Big - nInter[0];
            nInter[1] = Smo - nInter[1];
        }
        assert(nInter[0] >= Big/2, "nInter[0] >= Big/2");
        Gain = Abc_MaxInt( 0, nInter[0] - Big/2 + Smo/2 - nInter[1] );
        Vec_WecPush( vSorter, Gain, iDiv );
    }

    Vec_IntClear( vBinateVars );
    // inlining of macro Vec_WecForEachLevelReverse
    for  k = Vec_WecSize(vSorter)-1; (k >= 0) && (((vLevel) = Vec_WecEntry(vSorter, k)), 1); k--  {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLevel)) && (((iDiv) = Vec_IntEntry(vLevel, i)), 1); i++  {
            Vec_IntPush( vBinateVars, iDiv );
        }
    }
    Vec_WecClear( vSorter );

    if Vec_IntSize(vBinateVars) > 2000 {
        Vec_IntShrink( vBinateVars, 2000 ); }
}

/**Function*************************************************************

  Synopsis    [Perform resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManResubFindBestBinate(p *Gia_ResbMan_t) int32 {
    var nMintsAll int32  = Abc_TtCountOnesVec(p.pSets[0], p.nWords) + Abc_TtCountOnesVec(p.pSets[1], p.nWords);
    var i int32;
    var iDiv int32;
    var iLitBest int32 = -1;
    var CostBest int32 = -1;
//Vec_IntPrint( p->vBinateVars );
//Dau_DsdPrintFromTruth( p->pSets[0], 6 );
//Dau_DsdPrintFromTruth( p->pSets[1], 6 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vBinateVars)) && (((iDiv) = Vec_IntEntry(p.vBinateVars, i)), 1); i++  {
        var pDiv *word  = (word *)Vec_PtrEntry(p.vDivs, iDiv);
        var nMints0 int32  = Abc_TtCountOnesVecMask( pDiv, p.pSets[0], p.nWords, 0 );
        var nMints1 int32  = Abc_TtCountOnesVecMask( pDiv, p.pSets[1], p.nWords, 0 );
        if CostBest < nMints0 + nMints1 {
            CostBest = nMints0 + nMints1;
            iLitBest = Abc_Var2Lit( iDiv, 0 );
        }
        if CostBest < nMintsAll - nMints0 - nMints1 {
            CostBest = nMintsAll - nMints0 - nMints1;
            iLitBest = Abc_Var2Lit( iDiv, 1 );
        }
    }
    return iLitBest;
}
func Gia_ManResubAddNode(p *Gia_ResbMan_t, iLit0 int32, iLit1 int32, Type int32) int32 {
    var iNode int32  = Vec_PtrSize(p.vDivs) + Vec_IntSize(p.vGates)/2;
    var fFlip int32  = (Type == 2) ^ (iLit0 > iLit1);
    var iFan0 int32;
    if fFlip != 0 {
        iFan0 = iLit1;
    } else {
        iFan0 = iLit0;
    }
    var iFan1 int32;
    if fFlip != 0 {
        iFan1 = iLit0;
    } else {
        iFan1 = iLit1;
    }
    assert(iLit0 != iLit1, "iLit0 != iLit1");
    if Type == 2 {
        assert(iFan0 > iFan1, "iFan0 > iFan1");
    } else {
        assert(iFan0 < iFan1, "iFan0 < iFan1"); }
    Vec_IntPushTwo( p.vGates, Abc_LitNotCond(iFan0, Type==1), Abc_LitNotCond(iFan1, Type==1) );
    return Abc_Var2Lit( iNode, Type==1 );
}
func Gia_ManResubPerformMux_rec(p *Gia_ResbMan_t, nLimit int32) int32 {
    extern int Gia_ManResubPerform_rec( Gia_ResbMan_t * p, int nLimit );
    var iDivBest, iResLit0, iResLit1, nNodes int32;
    var pDiv *word;
    var pCopy [2]*word;
    if nLimit < 3 {
        return -1; }
    iDivBest = Gia_ManResubFindBestBinate( p );
    if iDivBest == -1 {
        return -1; }
    // inlining of macro ABC_CALLOC
    pCopy[0] = ((word *) calloc((size_t)(p.nWords), sizeof(word)));
    // inlining of macro ABC_CALLOC
    pCopy[1] = ((word *) calloc((size_t)(p.nWords), sizeof(word)));
    Abc_TtCopy( pCopy[0], p.pSets[0], p.nWords, 0 );
    Abc_TtCopy( pCopy[1], p.pSets[1], p.nWords, 0 );
    pDiv = (word *)Vec_PtrEntry( p.vDivs, Abc_Lit2Var(iDivBest) );
    Abc_TtAndSharp( p.pSets[0], pCopy[0], pDiv, p.nWords, !Abc_LitIsCompl(iDivBest) );
    Abc_TtAndSharp( p.pSets[1], pCopy[1], pDiv, p.nWords, !Abc_LitIsCompl(iDivBest) );
    nNodes = Vec_IntSize(p.vGates)/2;
    iResLit0 = Gia_ManResubPerform_rec( p, nLimit-3 );
    if iResLit0 == -1 {
        // inlining of macro ABC_FREE
        ((pCopy[0]) ? (free((char *) (pCopy[0])), (pCopy[0]) = 0) : 0);
        // inlining of macro ABC_FREE
        ((pCopy[1]) ? (free((char *) (pCopy[1])), (pCopy[1]) = 0) : 0);
        return -1;
    }
    Abc_TtAndSharp( p.pSets[0], pCopy[0], pDiv, p.nWords,  Abc_LitIsCompl(iDivBest) );
    Abc_TtAndSharp( p.pSets[1], pCopy[1], pDiv, p.nWords,  Abc_LitIsCompl(iDivBest) );
    // inlining of macro ABC_FREE
    ((pCopy[0]) ? (free((char *) (pCopy[0])), (pCopy[0]) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pCopy[1]) ? (free((char *) (pCopy[1])), (pCopy[1]) = 0) : 0);
    nNodes = Vec_IntSize(p.vGates)/2 - nNodes;
    if nLimit-nNodes < 3 {
        return -1; }
    iResLit1 = Gia_ManResubPerform_rec( p, nLimit-3-nNodes );
    if iResLit1 == -1 {
        return -1;
    } else {
        var iLit0 int32  = Gia_ManResubAddNode( p, Abc_LitNot(iDivBest), iResLit0, 0 );
        var iLit1 int32  = Gia_ManResubAddNode( p,            iDivBest,  iResLit1, 0 );
        return Gia_ManResubAddNode( p, iLit0, iLit1, 1 );
    }
}

/**Function*************************************************************

  Synopsis    [Perform resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManResubPerform_rec(p *Gia_ResbMan_t, nLimit int32) int32 {
    var TopOneW [2]int32 ;
    var TopTwoW [2]int32 ;
    var Max1 int32;
    var Max2 int32;
    var iResLit int32;
    var nVars int32 = Vec_PtrSize(p.vDivs);
    if p.fVerbose {
        var nMints [2]int32 = [2]int32{ Abc_TtCountOnesVec(p.pSets[0], p.nWords), Abc_TtCountOnesVec(p.pSets[1], p.nWords) };
        fmt.Printf("      "); 
        fmt.Printf("ISF: "); 
        fmt.Printf("0=%5d (%5.2f %%) ", nMints[0], 100.0*nMints[0]/(64*p.nWords));
        fmt.Printf("1=%5d (%5.2f %%)  ", nMints[1], 100.0*nMints[1]/(64*p.nWords));
    }
    if Abc_TtIsConst0( p.pSets[1], p.nWords ) {
        return 0; }
    if Abc_TtIsConst0( p.pSets[0], p.nWords ) {
        return 1; }
    iResLit = Gia_ManFindOneUnate( p.pSets, p.vDivs, p.nWords, p.vUnateLits, p.vNotUnateVars, p.fVerbose );
    if iResLit >= 0 { // buffer
        return iResLit; }
    if nLimit == 0 {
        return -1; }
    Gia_ManSortUnates( p.pSets, p.vDivs, p.nWords, p.vUnateLits, p.vUnateLitsW, p.vSorter );
    iResLit = Gia_ManFindTwoUnate( p.pSets, p.vDivs, p.nWords, p.vUnateLits, p.vUnateLitsW, p.fVerbose );
    if iResLit >= 0 { // and
    
        var iNode int32  = nVars + Vec_IntSize(p.vGates)/2;
        var fComp int32  = Abc_LitIsCompl(iResLit);
        var iDiv0 int32  = Abc_Lit2Var(iResLit) & 0x7FFF;
        var iDiv1 int32  = Abc_Lit2Var(iResLit) >> 15;
        assert(iDiv0 < iDiv1, "iDiv0 < iDiv1");
        Vec_IntPushTwo( p.vGates, iDiv0, iDiv1 );
        return Abc_Var2Lit( iNode, fComp );
    }
    Vec_IntTwoFindCommon( p.vNotUnateVars[0], p.vNotUnateVars[1], p.vBinateVars );
    //return Gia_ManResubPerformMux_rec( p, nLimit );
    if Vec_IntSize(p.vBinateVars) > p.nDivsMax {
        Vec_IntShrink( p.vBinateVars, p.nDivsMax ); }
    if p.fVerbose { fmt.Printf("  B = %3d", Vec_IntSize(p.vBinateVars)); }
    //Gia_ManSortBinate( p->pSets, p->vDivs, p->nWords, p->vBinateVars, p->vSorter );
    if p.fUseXor {
        iResLit = Gia_ManFindXor( p.pSets, p.vDivs, p.nWords, p.vBinateVars, p.vUnatePairs, p.fVerbose );
        if iResLit >= 0 { // xor
        
            var iNode int32  = nVars + Vec_IntSize(p.vGates)/2;
            var fComp int32  = Abc_LitIsCompl(iResLit);
            var iDiv0 int32  = Abc_Lit2Var(iResLit) & 0x7FFF;
            var iDiv1 int32  = Abc_Lit2Var(iResLit) >> 15;
            assert(!Abc_LitIsCompl(iDiv0), "!Abc_LitIsCompl(iDiv0)");
            assert(!Abc_LitIsCompl(iDiv1), "!Abc_LitIsCompl(iDiv1)");
            assert(iDiv0 > iDiv1, "iDiv0 > iDiv1");
            Vec_IntPushTwo( p.vGates, iDiv0, iDiv1 );
            return Abc_Var2Lit( iNode, fComp );
        }
    }
    if nLimit == 1 {
        return -1; }
    Gia_ManFindUnatePairs( p.pSets, p.vDivs, p.nWords, p.vBinateVars, p.vUnatePairs, p.fVerbose );
    Gia_ManSortPairs( p.pSets, p.vDivs, p.nWords, p.vUnatePairs, p.vUnatePairsW, p.vSorter );
    iResLit = Gia_ManFindDivGate( p.pSets, p.vDivs, p.nWords, p.vUnateLits, p.vUnatePairs, p.vUnateLitsW, p.vUnatePairsW, p.pDivA );
    if iResLit >= 0 { // and(div,pair)
    
        var iNode int32   = nVars + Vec_IntSize(p.vGates)/2;

        var fComp int32   = Abc_LitIsCompl(iResLit);
        var iDiv0 int32   = Abc_Lit2Var(iResLit) & 0x7FFF; // div
        var iDiv1 int32   = Abc_Lit2Var(iResLit) >> 15;    // pair

        var Div1 int32    = Vec_IntEntry( p.vUnatePairs[!fComp], Abc_Lit2Var(iDiv1) );
        var fComp1 int32  = Abc_LitIsCompl(Div1) ^ Abc_LitIsCompl(iDiv1);
        var iDiv10 int32  = Abc_Lit2Var(Div1) & 0x7FFF;
        var iDiv11 int32  = Abc_Lit2Var(Div1) >> 15;   

        Vec_IntPushTwo( p.vGates, iDiv10, iDiv11 );
        Vec_IntPushTwo( p.vGates, iDiv0, Abc_Var2Lit(iNode, fComp1) );
        return Abc_Var2Lit( iNode+1, fComp );
    }
//    if ( nLimit == 2 )
//        return -1;
    if nLimit >= 3 {
        iResLit = Gia_ManFindGateGate( p.pSets, p.vDivs, p.nWords, p.vUnatePairs, p.vUnatePairsW, p.pDivA, p.pDivB );
        if iResLit >= 0 { // and(pair,pair)
        
            var iNode int32   = nVars + Vec_IntSize(p.vGates)/2;

            var fComp int32   = Abc_LitIsCompl(iResLit);
            var iDiv0 int32   = Abc_Lit2Var(iResLit) & 0x7FFF; // pair
            var iDiv1 int32   = Abc_Lit2Var(iResLit) >> 15;    // pair

            var Div0 int32    = Vec_IntEntry( p.vUnatePairs[!fComp], Abc_Lit2Var(iDiv0) );
            var fComp0 int32  = Abc_LitIsCompl(Div0) ^ Abc_LitIsCompl(iDiv0);
            var iDiv00 int32  = Abc_Lit2Var(Div0) & 0x7FFF;
            var iDiv01 int32  = Abc_Lit2Var(Div0) >> 15;   
        
            var Div1 int32    = Vec_IntEntry( p.vUnatePairs[!fComp], Abc_Lit2Var(iDiv1) );
            var fComp1 int32  = Abc_LitIsCompl(Div1) ^ Abc_LitIsCompl(iDiv1);
            var iDiv10 int32  = Abc_Lit2Var(Div1) & 0x7FFF;
            var iDiv11 int32  = Abc_Lit2Var(Div1) >> 15;   
        
            Vec_IntPushTwo( p.vGates, iDiv00, iDiv01 );
            Vec_IntPushTwo( p.vGates, iDiv10, iDiv11 );
            Vec_IntPushTwo( p.vGates, Abc_Var2Lit(iNode, fComp0), Abc_Var2Lit(iNode+1, fComp1) );
            return Abc_Var2Lit( iNode+2, fComp );
        }
    }
//    if ( nLimit == 3 )
//        return -1;
    if Vec_IntSize(p.vUnateLits[0]) + Vec_IntSize(p.vUnateLits[1]) + Vec_IntSize(p.vUnatePairs[0]) + Vec_IntSize(p.vUnatePairs[1]) == 0 {
        return -1; }

    TopOneW[0] = Vec_IntSize(p.vUnateLitsW[0]) ? Vec_IntEntry(p.vUnateLitsW[0], 0) : 0;
    TopOneW[1] = Vec_IntSize(p.vUnateLitsW[1]) ? Vec_IntEntry(p.vUnateLitsW[1], 0) : 0;

    TopTwoW[0] = Vec_IntSize(p.vUnatePairsW[0]) ? Vec_IntEntry(p.vUnatePairsW[0], 0) : 0;
    TopTwoW[1] = Vec_IntSize(p.vUnatePairsW[1]) ? Vec_IntEntry(p.vUnatePairsW[1], 0) : 0;

    Max1 = Abc_MaxInt(TopOneW[0], TopOneW[1]);
    Max2 = Abc_MaxInt(TopTwoW[0], TopTwoW[1]);
    if Abc_MaxInt(Max1, Max2) == 0 {
        return -1; }

    if Max1 > Max2/2 {
        if nLimit >= 2 && (Max1 == TopOneW[0] || Max1 == TopOneW[1]) {
            var fUseOr int32   = Max1 == TopOneW[0];
            var iDiv int32     = Vec_IntEntry( p.vUnateLits[!fUseOr], 0 );
            var fComp int32    = Abc_LitIsCompl(iDiv);
            var pDiv *word  = (word *)Vec_PtrEntry( p.vDivs, Abc_Lit2Var(iDiv) );
            Abc_TtAndSharp( p.pSets[fUseOr], p.pSets[fUseOr], pDiv, p.nWords, !fComp );
            if p.fVerbose {
                fmt.Printf("\n"); } 
            iResLit = Gia_ManResubPerform_rec( p, nLimit-1 );
            if iResLit >= 0 {
                var iNode int32  = nVars + Vec_IntSize(p.vGates)/2;
                Vec_IntPushTwo( p.vGates, Abc_LitNot(iDiv), Abc_LitNotCond(iResLit, fUseOr) );
                return Abc_Var2Lit( iNode, fUseOr );
            }
        }
        if Max2 == 0 {
            return -1; }
/*
        if ( Max2 == TopTwoW[0] || Max2 == TopTwoW[1] )
        {
            int fUseOr  = Max2 == TopTwoW[0];
            int iDiv    = Vec_IntEntry( p->vUnatePairs[!fUseOr], 0 );
            int fComp   = Abc_LitIsCompl(iDiv);
            Gia_ManDeriveDivPair( iDiv, p->vDivs, p->nWords, p->pDivA );
            Abc_TtAndSharp( p->pSets[fUseOr], p->pSets[fUseOr], p->pDivA, p->nWords, !fComp );
            if ( p->fVerbose )
                printf( "\n      " ); 
            iResLit = Gia_ManResubPerform_rec( p, nLimit-2 );
            if ( iResLit >= 0 ) 
            {
                int iNode = nVars + Vec_IntSize(p->vGates)/2;
                int iDiv0 = Abc_Lit2Var(iDiv) & 0x7FFF;   
                int iDiv1 = Abc_Lit2Var(iDiv) >> 15;      
                Vec_IntPushTwo( p->vGates, iDiv0, iDiv1 );
                Vec_IntPushTwo( p->vGates, Abc_LitNotCond(iResLit, fUseOr), Abc_Var2Lit(iNode, !fComp) );
                return Abc_Var2Lit( iNode+1, fUseOr );
            }
        }
*/
    
    } else {
        if nLimit >= 3 && (Max2 == TopTwoW[0] || Max2 == TopTwoW[1]) {
            var fUseOr int32   = Max2 == TopTwoW[0];
            var iDiv int32     = Vec_IntEntry( p.vUnatePairs[!fUseOr], 0 );
            var fComp int32    = Abc_LitIsCompl(iDiv);
            Gia_ManDeriveDivPair( iDiv, p.vDivs, p.nWords, p.pDivA );
            Abc_TtAndSharp( p.pSets[fUseOr], p.pSets[fUseOr], p.pDivA, p.nWords, !fComp );
            if p.fVerbose {
                fmt.Printf("\n"); } 
            iResLit = Gia_ManResubPerform_rec( p, nLimit-2 );
            if iResLit >= 0 {
                var iNode int32  = nVars + Vec_IntSize(p.vGates)/2;
                var iDiv0 int32  = Abc_Lit2Var(iDiv) & 0x7FFF;   
                var iDiv1 int32  = Abc_Lit2Var(iDiv) >> 15;      
                Vec_IntPushTwo( p.vGates, iDiv0, iDiv1 );
                Vec_IntPushTwo( p.vGates, Abc_LitNotCond(iResLit, fUseOr), Abc_Var2Lit(iNode, !fComp) );
                return Abc_Var2Lit( iNode+1, fUseOr );
            }
        }
        if Max1 == 0 {
            return -1; }
/*
        if ( Max1 == TopOneW[0] || Max1 == TopOneW[1] )
        {
            int fUseOr  = Max1 == TopOneW[0];
            int iDiv    = Vec_IntEntry( p->vUnateLits[!fUseOr], 0 );
            int fComp   = Abc_LitIsCompl(iDiv);
            word * pDiv = (word *)Vec_PtrEntry( p->vDivs, Abc_Lit2Var(iDiv) );
            Abc_TtAndSharp( p->pSets[fUseOr], p->pSets[fUseOr], pDiv, p->nWords, !fComp );
            if ( p->fVerbose )
                printf( "\n      " ); 
            iResLit = Gia_ManResubPerform_rec( p, nLimit-1 );
            if ( iResLit >= 0 ) 
            {
                int iNode = nVars + Vec_IntSize(p->vGates)/2;
                Vec_IntPushTwo( p->vGates, Abc_LitNot(iDiv), Abc_LitNotCond(iResLit, fUseOr) );
                return Abc_Var2Lit( iNode, fUseOr );
            }
        }
*/
    }
    return -1;
}
func Gia_ManResubPerform(p *Gia_ResbMan_t, vDivs *Vec_Ptr_t, nWords int32, nLimit int32, nDivsMax int32, iChoice int32, fUseXor int32, fDebug int32, fVerbose int32) {
    var Res int32;
    Gia_ResbInit( p, vDivs, nWords, nLimit, nDivsMax, iChoice, fUseXor, fDebug, fVerbose, fVerbose );
    Res = Gia_ManResubPerform_rec( p, nLimit );
    if Res >= 0 { 
        Vec_IntPush( p.vGates, Res );
    } else {
        Vec_IntClear( p.vGates ); }
    if fVerbose != 0 {
        fmt.Printf("\n"); }
}
func Gia_ManResubOne(vDivs *Vec_Ptr_t, nWords int32, nLimit int32, nDivsMax int32, iChoice int32, fUseXor int32, fDebug int32, fVerbose int32, pFunc *word) *Vec_Int_t {
    var vRes *Vec_Int_t;
    var p *Gia_ResbMan_t  = Gia_ResbAlloc( nWords );
    Gia_ManResubPerform( p, vDivs, nWords, nLimit, nDivsMax, iChoice, fUseXor, fDebug, fVerbose );
    if fVerbose != 0 {
        Gia_ManResubPrint( p.vGates, Vec_PtrSize(vDivs) ); }
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    if !(Gia_ManResubVerify(p, pFunc) != 0) {
        Gia_ManResubPrint( p.vGates, Vec_PtrSize(vDivs) );
        fmt.Printf("Verification FAILED.\n");
    
    } else if (fDebug != 0) && (fVerbose != 0) {
        fmt.Printf("Verification succeeded."); }
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    vRes = Vec_IntDup( p.vGates );
    Gia_ResbFree( p );
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Top level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var s_pResbMan *Gia_ResbMan_t = nil;

func Abc_ResubPrepareManager(nWords int32) {
    if s_pResbMan != nil {
        Gia_ResbFree( s_pResbMan ); }
    s_pResbMan = nil;
    if nWords > 0 {
        s_pResbMan = Gia_ResbAlloc( nWords ); }
}

func Abc_ResubComputeFunction(ppDivs *any, nDivs int32, nWords int32, nLimit int32, nDivsMax int32, iChoice int32, fUseXor int32, fDebug int32, fVerbose int32, ppArray **int32) int32 {
    var Divs Vec_Ptr_t  = Vec_Ptr_t{ nDivs, nDivs, ppDivs };
    assert(s_pResbMan != nil, "s_pResbMan != nil"); // first call Abc_ResubPrepareManager()
    Gia_ManResubPerform( s_pResbMan, &Divs, nWords, nLimit, nDivsMax, iChoice, fUseXor, fDebug, fVerbose==2 );
    if fVerbose != 0 {
        var nGates int32  = Vec_IntSize(s_pResbMan.vGates)/2;
        if nGates != 0 {
            fmt.Printf("      Gain = %2d  Gates = %2d  __________  F = ", nLimit+1-nGates, nGates);
            Gia_ManResubPrint( s_pResbMan.vGates, nDivs );
            fmt.Printf("\n");
        }
    }
    if fDebug != 0 {
        if !(Gia_ManResubVerify(s_pResbMan, nil) != 0) {
            Gia_ManResubPrint( s_pResbMan.vGates, nDivs );
            fmt.Printf("Verification FAILED.\n");
        }
        //else
        //    printf( "Verification succeeded.\n" );
    }
    *ppArray = Vec_IntArray(s_pResbMan.vGates);
    assert(Vec_IntSize(s_pResbMan.vGates)/2 <= nLimit, "Vec_IntSize(s_pResbMan.vGates)/2 <= nLimit");
    return Vec_IntSize(s_pResbMan.vGates);
}

func Abc_ResubDumpProblem(pFileName string, ppDivs *any, nDivs int32, nWords int32) {
    var vSims *Vec_Wrd_t  = Vec_WrdAlloc( nDivs * nWords );
    var pDivs **word  = (word **)ppDivs;
    var d, w int32;
    for  d = 0; d < nDivs;  d++  {
    for  w = 0; w < nWords; w++  {
        Vec_WrdPush( vSims, pDivs[d][w] );
    }
    }
    Vec_WrdDumpHex( pFileName, vSims, nWords, 1 );
    Vec_WrdFree( vSims );
}

/**Function*************************************************************

  Synopsis    [Top level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

// extern void Extra_PrintHex( FILE * pFile, unsigned * pTruth, int nVars );

// extern void Dau_DsdPrintFromTruth2( word * pTruth, int nVarsInit );


func Gia_ManResubTest3() {
    var nVars int32  = 4;
    var fVerbose int32  = 1;
    var Divs [6]word = [6]word{ 0, 0, 
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
        ABC_CONST(0xFF00FF00FF00FF00), 
    };
    var vDivs *Vec_Ptr_t  = Vec_PtrAlloc( 6 );
    var vRes *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i int32;
    var k int32;
    var ArraySize int32;
    var pArray *int32; 
    for  i = 0; i < 6; i++  {
        Vec_PtrPush( vDivs, Divs+i );
    }
    Abc_ResubPrepareManager( 1 );
    for  i = 0; i < (1<<(1<<nVars)); i++  { //if ( i == 0xCA ) 
    
        var Truth word  = Abc_Tt6Stretch( i, nVars );
        Divs[0] = ~Truth;
        Divs[1] =  Truth;
        fmt.Printf("%3d : ", i);
        Extra_PrintHex( os.Stdout, (unsigned*)&Truth, nVars );
        fmt.Printf(" ");
        Dau_DsdPrintFromTruth2( &Truth, nVars );
        fmt.Printf("           ");

        //Abc_ResubDumpProblem( "temp.resub", (void **)Vec_PtrArray(vDivs), Vec_PtrSize(vDivs), 1 );
        ArraySize = Abc_ResubComputeFunction( (void **)Vec_PtrArray(vDivs), Vec_PtrSize(vDivs), 1, 16, 50, 0, 0, 1, fVerbose, &pArray );
        fmt.Printf("\n");

        Vec_IntClear( vRes );
        for  k = 0; k < ArraySize; k++  {
            Vec_IntPush( vRes, pArray[k] );
        }

        if i == 1000 {
            break; }
    }
    Abc_ResubPrepareManager( 0 );
    Vec_IntFree( vRes );
    Vec_PtrFree( vDivs );
}
func Gia_ManResubTest3_() {
    var p *Gia_ResbMan_t  = Gia_ResbAlloc( 1 );
    var Divs [6]word = [6]word{ 0, 0, 
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
        ABC_CONST(0xFF00FF00FF00FF00), 
    };
    var vDivs *Vec_Ptr_t  = Vec_PtrAlloc( 6 );
    var vRes *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i int32; 
    for  i = 0; i < 6; i++  {
        Vec_PtrPush( vDivs, Divs+i );
    }

    {
        var Truth word  = (Divs[2] | Divs[3]) & (Divs[4] & Divs[5]);
//        word Truth = (~Divs[2] | Divs[3]) | ~Divs[4];
        Divs[0] = ~Truth;
        Divs[1] =  Truth;
        Extra_PrintHex( os.Stdout, (unsigned*)&Truth, 6 );
        fmt.Printf(" ");
        Dau_DsdPrintFromTruth2( &Truth, 6 );
        fmt.Printf("       ");
        Gia_ManResubPerform( p, vDivs, 1, 100, 0, 50, 1, 1, 0 );
    }
    Gia_ResbFree( p );
    Vec_IntFree( vRes );
    Vec_PtrFree( vDivs );
}

/**Function*************************************************************

  Synopsis    [Top level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckResub(vDivs *Vec_Ptr_t, nWords int32) {
    //int i, nVars = 6, pVarSet[10] = { 2, 189, 2127, 2125, 177, 178 };
    var i int32;
    var nVars int32 = 3;
    var pVarSet [10]int32 = [10]int32{ 2, 3, 4 };
    var pOff *word  = (word *)Vec_PtrEntry( vDivs, 0 );
    var pOn *word   = (word *)Vec_PtrEntry( vDivs, 1 );
    var vValue *Vec_Int_t  = Vec_IntStartFull( 1 << 6 );
    fmt.Printf("Verifying resub:\n");
    for  i = 0; i < 64*nWords; i++  {
        var v int32;
        var Mint int32;
        var Value int32 = Abc_TtGetBit(pOn, i);
        if !Abc_TtGetBit(pOff, i) && !(Value != 0) {
            continue; }
        for  v = 0; v < nVars; v++  {
            if Abc_TtGetBit((word *)Vec_PtrEntry(vDivs, pVarSet[v]), i) {
                Mint |= 1 << v; }
        }
        if Vec_IntEntry(vValue, Mint) == -1 {
            Vec_IntWriteEntry(vValue, Mint, Value);
        } else if Vec_IntEntry(vValue, Mint) != Value {
            fmt.Printf("Mismatch in pattern %d\n", i); }
    }
    fmt.Printf("Finished verifying resub.\n");
    Vec_IntFree( vValue );
}
func Gia_ManDeriveDivs(vSims *Vec_Wrd_t, nWords int32) *Vec_Ptr_t {
    var i int32;
    var nDivs int32 = Vec_WrdSize(vSims)/nWords;
    var vDivs *Vec_Ptr_t  = Vec_PtrAlloc( nDivs );
    for  i = 0; i < nDivs; i++  {
        Vec_PtrPush( vDivs, Vec_WrdEntryP(vSims, nWords*i) );
    }
    return vDivs;
}
func Gia_ManResub2(pGia *Gia_Man_t, nNodes int32, nSupp int32, nDivs int32, iChoice int32, fUseXor int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    return nil;
}
func Gia_ManResub1(pFileName string, nNodes int32, nSupp int32, nDivs int32, iChoice int32, fUseXor int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var nWords int32;
    var pMan *Gia_Man_t    = nil;
    var vSims *Vec_Wrd_t   = Vec_WrdReadHex( pFileName, &nWords, 1 );
    var vDivs *Vec_Ptr_t;
    if vSims != nil {
        vDivs = Gia_ManDeriveDivs( vSims, nWords );
    } else {
        vDivs = nil;
    }
    var p *Gia_ResbMan_t  = Gia_ResbAlloc( nWords );
    //Gia_ManCheckResub( vDivs, nWords );
    if Vec_PtrSize(vDivs) >= (1<<14) {
        fmt.Printf("Reducing all divs from %d to %d.\n", Vec_PtrSize(vDivs), (1<<14)-1);
        Vec_PtrShrink( vDivs, (1<<14)-1 );
    }
    assert(Vec_PtrSize(vDivs) < (1<<14), "Vec_PtrSize(vDivs) < (1<<14)");
    Gia_ManResubPerform( p, vDivs, nWords, 100, 50, iChoice, fUseXor, 1, 1 );
    if Vec_IntSize(p.vGates) {
        var vGates *Vec_Wec_t  = Vec_WecStart(1);
        Vec_IntAppend( Vec_WecEntry(vGates, 0), p.vGates );
        pMan = Gia_ManConstructFromGates( vGates, Vec_PtrSize(vDivs) );
        Vec_WecFree( vGates );
    
    } else {
        fmt.Printf("Decomposition did not succeed.\n"); }
    Gia_ResbFree( p );
    Vec_PtrFree( vDivs );
    Vec_WrdFree( vSims );
    return pMan;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnivTfo_rec(p *Gia_Man_t, iObj int32, vNodes *Vec_Int_t, vPos *Vec_Int_t) int32 {
    var i int32;
    var iFan int32;
    var Count int32 = 1;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return 0; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    if (vNodes != nil) && Gia_ObjIsCo(Gia_ManObj(p, iObj)) {
        Vec_IntPush( vNodes, iObj ); }    
    if (vPos != nil) && Gia_ObjIsCo(Gia_ManObj(p, iObj)) {
        Vec_IntPush( vPos, iObj ); }
    // inlining of macro Gia_ObjForEachFanoutStaticId
    for  i = 0; (i < Gia_ObjFanoutNumId(p, iObj))   && (((iFan) = Gia_ObjFanoutId(p, iObj, i)), 1); i++  {
        Count += Gia_ManUnivTfo_rec( p, iFan, vNodes, vPos );
    }
    return Count;
}
func Gia_ManUnivTfo(p *Gia_Man_t, pObjs *int32, nObjs int32, pvNodes **Vec_Int_t, pvPos **Vec_Int_t) int32 {
    var i, Count int32;
    if pvNodes {
        if *pvNodes {
            Vec_IntClear( *pvNodes );
        } else {
            *pvNodes = Vec_IntAlloc( 100 ); }
    }
    if pvPos {
        if *pvPos {
            Vec_IntClear( *pvPos );
        } else {
            *pvPos = Vec_IntAlloc( 100 ); }
    }
    Gia_ManIncrementTravId( p );
    for  i = 0; i < nObjs; i++  {
        Count += Gia_ManUnivTfo_rec( p, pObjs[i], pvNodes ? *pvNodes : nil, pvPos ? *pvPos : nil );
    }
    if pvNodes {
        Vec_IntSort( *pvNodes, 0 ); }
    if pvPos {
        Vec_IntSort( *pvPos, 0 ); }
    return Count;
}

/**Function*************************************************************

  Synopsis    [Tuning resub.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTryResub(p *Gia_Man_t) {
    var nLimit int32    =  20;
    var nDivsMax int32  = 200;
    var iChoice int32;
    var fUseXor int32   =   1;
    var fDebug int32    =   1;
    var fVerbose int32;
    var clk abctime;
    var clkResub abctime;
    var clkStart abctime = Abc_Clock();
    var vvSims *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
    var vSims *Vec_Wrd_t;
    var pSets [2]*word;
    var pFunc *word;
    var pObj, pObj2 *Gia_Obj_t; 
    var i, i2, nWords, nNonDec, nTotal int32;
    assert(Gia_ManCiNum(p) < 16, "Gia_ManCiNum(p) < 16");
    Vec_WrdFreeP( &p.vSimsPi );
    p.vSimsPi = Vec_WrdStartTruthTables( Gia_ManCiNum(p) );
    nWords = Vec_WrdSize(p.vSimsPi) / Gia_ManCiNum(p);
    //Vec_WrdPrintHex( p->vSimsPi, nWords );
    // inlining of macro ABC_CALLOC
    pSets[0] = ((word *) calloc((size_t)(nWords), sizeof(word)));
    // inlining of macro ABC_CALLOC
    pSets[1] = ((word *) calloc((size_t)(nWords), sizeof(word)));
    vSims = Gia_ManSimPatSim( p );
    Gia_ManLevelNum(p);
    Gia_ManCreateRefs(p);
    Abc_ResubPrepareManager( nWords );
    Gia_ManStaticFanoutStart( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var vGates Vec_Int_t;
        var pArray *int32;
        var nArray int32;
        var nTfo int32;
        var iObj int32 = Gia_ObjId(p, pObj);
        var Level int32  = Gia_ObjLevel(p, pObj);
        var nMffc int32  = Gia_NodeMffcSizeMark(p, pObj);
        pFunc = Vec_WrdEntryP( vSims, nWords*iObj );
        Abc_TtCopy( pSets[0], pFunc, nWords, 1 );
        Abc_TtCopy( pSets[1], pFunc, nWords, 0 );
        Vec_PtrClear( vvSims );
        Vec_PtrPushTwo( vvSims, pSets[0], pSets[1] );
        nTfo = Gia_ManUnivTfo( p, &iObj, 1, nil, nil );
        // inlining of macro Gia_ManForEachCi
        for  i2 = 0; (i2 < Vec_IntSize(p.vCis)) && ((pObj2) = Gia_ManCi(p, i2)); i2++  {
            Vec_PtrPush( vvSims, Vec_WrdEntryP(vSims, nWords*Gia_ObjId(p, pObj2)) );
        }
        // inlining of macro Gia_ManForEachAnd
        for  i2 = 0; (i2 < p.nObjs) && ((pObj2) = Gia_ManObj(p, i2)); i2++  {      if !Gia_ObjIsAnd(pObj2) {} else
            if !Gia_ObjIsTravIdCurrent(p, pObj2) && !Gia_ObjIsTravIdPrevious(p, pObj2) && Gia_ObjLevel(p, pObj2) <= Level {
                Vec_PtrPush( vvSims, Vec_WrdEntryP(vSims, nWords*Gia_ObjId(p, pObj2)) ); }
        }
        if fVerbose != 0 {
        fmt.Printf("%3d : Lev = %2d  Mffc = %2d  Divs = %3d  Tfo = %3d\n", iObj, Level, nMffc, Vec_PtrSize(vvSims)-2, nTfo); }
        clk = Abc_Clock();
        nArray = Abc_ResubComputeFunction( (void **)Vec_PtrArray(vvSims), Vec_PtrSize(vvSims), nWords, Abc_MinInt(nMffc-1, nLimit), nDivsMax, iChoice, fUseXor, fDebug, fVerbose, &pArray );
        clkResub += Abc_Clock() - clk;
        vGates.nSize = vGates.nCap = nArray;
        vGates.pArray = pArray;
        assert(nMffc > Vec_IntSize(&vGates)/2, "nMffc > Vec_IntSize(&vGates)/2");
        if Vec_IntSize(&vGates) > 0 {
            nTotal += nMffc - Vec_IntSize(&vGates)/2; }
        nNonDec += Vec_IntSize(&vGates) == 0;
    }
    }
    fmt.Printf("Total nodes = %5d.  Non-realizable = %5d.  Gain = %6d.  ", Gia_ManAndNum(p), nNonDec, nTotal);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
    Abc_PrintTime( 1, "Pure resub time", clkResub );
    Abc_ResubPrepareManager( 0 );
    Gia_ManStaticFanoutStop( p );
    Vec_PtrFree( vvSims );
    Vec_WrdFree( vSims );
    // inlining of macro ABC_FREE
    ((pSets[0]) ? (free((char *) (pSets[0])), (pSets[0]) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pSets[1]) ? (free((char *) (pSets[1])), (pSets[1]) = 0) : 0);
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaEquiv.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Manipulation of equivalence classes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaEquiv.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "proof/cec/cec.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Manipulating original IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManOrigIdsInit(p *Gia_Man_t) {
    Vec_IntFreeP( &p.vIdsOrig );
    p.vIdsOrig = Vec_IntStartNatural( Gia_ManObjNum(p) );
}
func Gia_ManOrigIdsStart(p *Gia_Man_t) {
    Vec_IntFreeP( &p.vIdsOrig );
    p.vIdsOrig = Vec_IntStartFull( Gia_ManObjNum(p) );
}
func Gia_ManOrigIdsRemap(p *Gia_Man_t, pNew *Gia_Man_t) {
    var pObj *Gia_Obj_t; var i int32;
    if p.vIdsOrig == nil {
        return; }
    Gia_ManOrigIdsStart( pNew );
    Vec_IntWriteEntry( pNew.vIdsOrig, 0, 0 );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if ~pObj.Value && Abc_Lit2Var(pObj.Value) && Vec_IntEntry(p.vIdsOrig, i) != -1 && Vec_IntEntry(pNew.vIdsOrig, Abc_Lit2Var(pObj.Value)) == -1 {
            Vec_IntWriteEntry( pNew.vIdsOrig, Abc_Lit2Var(pObj.Value), Vec_IntEntry(p.vIdsOrig, i) ); }
    }
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pNew.nObjs) && ((pObj) = Gia_ManObj(pNew, i)); i++  {
        assert(Vec_IntEntry(pNew.vIdsOrig, i) >= 0, "Vec_IntEntry(pNew.vIdsOrig, i) >= 0");
    }
}
// input is a set of equivalent node pairs in any order
// output is the mapping of each node into the equiv node with the smallest ID
func Gia_ManOrigIdsRemapPairsInsert(vMap *Vec_Int_t, One int32, Two int32) {
    var Smo int32;
    if One < Two {
        Smo = One;
    } else {
        Smo = Two;
    }
    var Big int32;
    if One < Two {
        Big = Two;
    } else {
        Big = One;
    }
    assert(Smo != Big, "Smo != Big");
    if Vec_IntEntry(vMap, Big) == -1 {
        Vec_IntWriteEntry( vMap, Big, Smo );
    } else {
        Gia_ManOrigIdsRemapPairsInsert( vMap, Smo, Vec_IntEntry(vMap, Big) ); }
}
func Gia_ManOrigIdsRemapPairsExtract(vMap *Vec_Int_t, One int32) int32 {
    if Vec_IntEntry(vMap, One) == -1 {
        return One; }
    return Gia_ManOrigIdsRemapPairsExtract( vMap, Vec_IntEntry(vMap, One) );
}
func Gia_ManOrigIdsRemapPairs(vEquivPairs *Vec_Int_t, nObjs int32) *Vec_Int_t {
    var vMapResult *Vec_Int_t;
    var vMap2Smaller *Vec_Int_t;
    var i, One, Two int32;
    // map bigger into smaller one
    vMap2Smaller = Vec_IntStartFull( nObjs );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vEquivPairs)) && (((One) = Vec_IntEntry(vEquivPairs, i)), 1) && (((Two) = Vec_IntEntry(vEquivPairs, i+1)), 1); i += 2  {
        Gia_ManOrigIdsRemapPairsInsert( vMap2Smaller, One, Two );
    }
    // collect results in the topo order
    vMapResult = Vec_IntStartFull( nObjs );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vMap2Smaller)) && (((One) = Vec_IntEntry(vMap2Smaller, i)), 1); i++  {
        if One >= 0 {
            Vec_IntWriteEntry( vMapResult, i, Gia_ManOrigIdsRemapPairsExtract(vMap2Smaller, One) ); }
    }
    Vec_IntFree( vMap2Smaller );
    return vMapResult;
}
// remap the AIG using the equivalent pairs proved
// returns the reduced AIG and the equivalence classes of the original AIG
func Gia_ManOrigIdsReduce(p *Gia_Man_t, vPairs *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t  = nil;
    var pObj, pRepr *Gia_Obj_t; var i int32;
    var vMap *Vec_Int_t  = Gia_ManOrigIdsRemapPairs( vPairs, Gia_ManObjNum(p) );
    Gia_ManSetPhase( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManFillValue(p);
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Vec_IntEntry(vMap, i) == -1 {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else {
            pRepr = Gia_ManObj( p, Vec_IntEntry(vMap, i) );
            pObj.Value = Abc_LitNotCond( pRepr.Value, pRepr.fPhase ^ pObj.fPhase );
        }
    }
    }
    Gia_ManHashStop( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Vec_IntFree( vMap );
    // compute equivalences
    assert(!p.pReprs && !p.pNexts, "!p.pReprs && !p.pNexts");
    // inlining of macro ABC_CALLOC
    p.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Gia_Rpr_t)));
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        Gia_ObjSetRepr( p, i, GIA_VOID );
    }
    Gia_ManFillValue(pNew);
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var iRepr int32  = Abc_Lit2Var(pObj.Value);
        if iRepr == 0 {
            Gia_ObjSetRepr( p, i, 0 );
            continue;
        }
        pRepr = Gia_ManObj( pNew, iRepr );
        if !~pRepr.Value { // first time
        
            pRepr.Value = i;
            continue;
        }
        // add equivalence
        Gia_ObjSetRepr( p, i, pRepr.Value );
    }
    }
    p.pNexts = Gia_ManDeriveNexts( p );
    return pNew;
}
func Gia_ManOrigIdsReduceTest(p *Gia_Man_t, vPairs *Vec_Int_t) *Gia_Man_t {
    var pTemp *Gia_Man_t;
    var pNew *Gia_Man_t = Gia_ManOrigIdsReduce( p, vPairs );
    Gia_ManPrintStats( p, nil );
    Gia_ManPrintStats( pNew, nil );
    //Gia_ManStop( pNew );
    // cleanup the resulting one
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Compute equivalence classes of nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManComputeGiaEquivs(pGia *Gia_Man_t, nConfs int32, fVerbose int32) *Gia_Man_t {
    var pTemp *Gia_Man_t;
    var ParsFra Cec_ParFra_t;
    var pPars *Cec_ParFra_t = &ParsFra;
    Cec_ManFraSetDefaultParams( pPars );
    pPars.nItersMax = 100;
    pPars.fUseOrigIds = 1;
    pPars.fSatSweeping = 1;
    pPars.nBTLimit = nConfs;
    pPars.fVerbose = fVerbose;
    pTemp = Cec_ManSatSweeping( pGia, pPars, 0 );
    Gia_ManStop( pTemp );
    return Gia_ManOrigIdsReduce( pGia, pGia.vIdsEquiv );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if AIG is not in the required topo order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckTopoOrder_rec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    var pRepr *Gia_Obj_t;
    if pObj.Value == 0 {
        return 1; }
    pObj.Value = 0;
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if !(Gia_ManCheckTopoOrder_rec( p, Gia_ObjFanin0(pObj) ) != 0) {
        return 0; }
    if !(Gia_ManCheckTopoOrder_rec( p, Gia_ObjFanin1(pObj) ) != 0) {
        return 0; }
    if p.pReprs {
        pRepr = Gia_ObjReprObj( p, Gia_ObjId(p,pObj) );
    } else {
        pRepr = nil;
    }
    return pRepr == nil || pRepr.Value == 0;
}

/**Function*************************************************************

  Synopsis    [Returns 0 if AIG is not in the required topo order.]

  Description [AIG should be in such an order that the representative
  is always traversed before the node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckTopoOrder(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i int32;
    var RetValue int32 = 1;
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = 0;
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        RetValue &= Gia_ManCheckTopoOrder_rec( p, Gia_ObjFanin0(pObj) );
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Given representatives, derives pointers to the next objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Given points to the next objects, derives representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivCountLitsAll(p *Gia_Man_t) int32 {
    var i, nLits int32;
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        nLits += (Gia_ObjRepr(p, i) != GIA_VOID);
    }
    return nLits;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivCountClasses(p *Gia_Man_t) int32 {
    var i, Counter int32;
    if p.pReprs == nil {
        return 0; }
    for  i = 1; i < Gia_ManObjNum(p); i++  {
        Counter += Gia_ObjIsHead(p, i);
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivCheckLits(p *Gia_Man_t, nLits int32) int32 {
    var nLitsReal int32  = Gia_ManEquivCountLitsAll( p );
    if nLitsReal != nLits {
        Abc_Print( 1, "Detected a mismatch in counting equivalence classes (%d).\n", nLitsReal - nLits ); }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintStatsClasses(p *Gia_Man_t) {
    var i, Counter, Counter0, CounterX, Proved, nLits int32;
    for  i = 1; i < Gia_ManObjNum(p); i++  {
        if Gia_ObjIsHead(p, i) {
            Counter++;
        } else if Gia_ObjIsConst(p, i) {
            Counter0++;
        } else if Gia_ObjIsNone(p, i) {
            CounterX++; }
        if Gia_ObjProved(p, i) {
            Proved++; }
    }
    CounterX -= Gia_ManCoNum(p);
    nLits = Gia_ManCiNum(p) + Gia_ManAndNum(p) - Counter - CounterX;

//    Abc_Print( 1, "i/o/ff =%5d/%5d/%5d  ", Gia_ManPiNum(p), Gia_ManPoNum(p), Gia_ManRegNum(p) );
//    Abc_Print( 1, "and =%5d  ", Gia_ManAndNum(p) );
//    Abc_Print( 1, "lev =%3d  ", Gia_ManLevelNum(p) );
    Abc_Print( 1, "cst =%3d  cls =%6d  lit =%8d\n", Counter0, Counter, nLits );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivCountLits(p *Gia_Man_t) int32 {
    var i, Counter, Counter0, CounterX int32;
    if p.pReprs == nil || p.pNexts == nil {
        return 0; }
    for  i = 1; i < Gia_ManObjNum(p); i++  {
        if Gia_ObjIsHead(p, i) {
            Counter++;
        } else if Gia_ObjIsConst(p, i) {
            Counter0++;
        } else if Gia_ObjIsNone(p, i) {
            CounterX++; }
    }
    CounterX -= Gia_ManCoNum(p);
    return Gia_ManCiNum(p) + Gia_ManAndNum(p) - Counter - CounterX;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivCountOne(p *Gia_Man_t, i int32) int32 {
    var Ent int32;
    var nLits int32 = 1;
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p, i), "Gia_ObjIsHead(p, i)"), Ent = Gia_ObjNext(p, i); Ent > 0; Ent = Gia_ObjNext(p, Ent)  {
        assert(Gia_ObjRepr(p, Ent) == i, "Gia_ObjRepr(p, Ent) == i");
        nLits++;
    }
    return nLits;
}
func Gia_ManEquivPrintOne(p *Gia_Man_t, i int32, Counter int32) {
    var Ent int32;
    Abc_Print( 1, "Class %4d :  Num = %2d  {", Counter, Gia_ManEquivCountOne(p, i) );
    // inlining of macro Gia_ClassForEachObj
    for  assert(Gia_ObjIsHead(p, i) && (i != 0), "Gia_ObjIsHead(p, i) && i"), Ent = i; Ent > 0; Ent = Gia_ObjNext(p, Ent)  {
        Abc_Print( 1," %d", Ent );
        if p.pReprs[Ent].fColorA || p.pReprs[Ent].fColorB {
            Abc_Print( 1," <%d%d>", p.pReprs[Ent].fColorA, p.pReprs[Ent].fColorB ); }
    }
    Abc_Print( 1, " }\n" );
}
func Gia_ManEquivPrintClasses(p *Gia_Man_t, fVerbose int32, Mem float32) {
    var i, Counter, Counter0, CounterX, Proved, nLits int32;
    for  i = 1; i < Gia_ManObjNum(p); i++  {
        if Gia_ObjIsHead(p, i) {
            Counter++;
        } else if Gia_ObjIsConst(p, i) {
            Counter0++;
        } else if Gia_ObjIsNone(p, i) {
            CounterX++; }
        if Gia_ObjProved(p, i) {
            Proved++; }
    }
    CounterX -= Gia_ManCoNum(p);
    nLits = Gia_ManCiNum(p) + Gia_ManAndNum(p) - Counter - CounterX;
//    Abc_Print( 1, "cst =%8d  cls =%7d  lit =%8d  unused =%8d  proof =%6d  mem =%5.2f MB\n",
//        Counter0, Counter, nLits, CounterX, Proved, (Mem == 0.0) ? 8.0*Gia_ManObjNum(p)/(1<<20) : Mem );
    Abc_Print( 1, "cst =%8d  cls =%7d  lit =%8d  unused =%8d  proof =%6d\n",
        Counter0, Counter, nLits, CounterX, Proved );
    assert((Gia_ManEquivCheckLits( p, nLits ) != 0), "Gia_ManEquivCheckLits( p, nLits )");
    if fVerbose != 0 {
//        int Ent;
        Abc_Print( 1, "Const0 (%d) = ", Counter0 );
        // inlining of macro Gia_ManForEachConst
        for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsConst(p, i) {} else {
            Abc_Print( 1, "%d ", i ); }
        }
        Abc_Print( 1, "\n" );
        Counter = 0;
        // inlining of macro Gia_ManForEachClass
        for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
            Gia_ManEquivPrintOne( p, i, ++Counter ); }
        }
/*
        Gia_ManLevelNum( p );
        Gia_ManForEachClass( p, i )
            if ( i % 100 == 0 )
            {
//                Abc_Print( 1, "%d ", Gia_ManEquivCountOne(p, i) );
                Gia_ClassForEachObj( p, i, Ent )
                {
                    Abc_Print( 1, "%d ", Gia_ObjLevel( p, Gia_ManObj(p, Ent) ) );
                }
                Abc_Print( 1, "\n" );
            }
*/
    }
}


/**Function*************************************************************

  Synopsis    [Map representatives into class members with minimum level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManChoiceMinLevel_rec(p *Gia_Man_t, iPivot int32, fDiveIn int32, vMap *Vec_Int_t) int32 {
    var Level0, Level1, LevelMax int32;
    var pPivot *Gia_Obj_t  = Gia_ManObj( p, iPivot );
    if Gia_ObjIsCi(pPivot) {
        return 0; }
    if Gia_ObjLevel(p, pPivot) {
        return Gia_ObjLevel(p, pPivot); }
    if (fDiveIn != 0) && Gia_ObjIsClass(p, iPivot) {
        var iObj int32;
        var ObjMin int32 = -1;
        var iRepr int32 = Gia_ObjRepr(p, iPivot);
        var LevMin int32 = ABC_INFINITY;
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, iRepr) && (iRepr != 0), "Gia_ObjIsHead(p, iRepr) && iRepr"), iObj = iRepr; iObj > 0; iObj = Gia_ObjNext(p, iObj)  {
            var LevCur int32  = Gia_ManChoiceMinLevel_rec( p, iObj, 0, vMap );
            if LevMin > LevCur {
                LevMin = LevCur;
                ObjMin = iObj;
            }
        }
        assert(LevMin > 0, "LevMin > 0");
        Vec_IntWriteEntry( vMap, iRepr, ObjMin );
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, iRepr) && (iRepr != 0), "Gia_ObjIsHead(p, iRepr) && iRepr"), iObj = iRepr; iObj > 0; iObj = Gia_ObjNext(p, iObj)  {
            Gia_ObjSetLevelId( p, iObj, LevMin );
        }
        return LevMin;
    }
    assert(Gia_ObjIsAnd(pPivot), "Gia_ObjIsAnd(pPivot)");
    Level0 = Gia_ManChoiceMinLevel_rec( p, Gia_ObjFaninId0(pPivot, iPivot), 1, vMap );
    Level1 = Gia_ManChoiceMinLevel_rec( p, Gia_ObjFaninId1(pPivot, iPivot), 1, vMap );
    LevelMax = 1 + Abc_MaxInt(Level0, Level1);
    Gia_ObjSetLevel( p, pPivot, LevelMax );
    return LevelMax;
}
func Gia_ManChoiceMinLevel(p *Gia_Man_t) *Vec_Int_t {
    var vMap *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    var pObj *Gia_Obj_t;
    var i, LevelCur, LevelMax int32;
//    assert( Gia_ManRegNum(p) == 0 );
    Gia_ManCleanLevels( p, Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        LevelCur = Gia_ManChoiceMinLevel_rec( p, Gia_ObjFaninId0p(p, pObj), 1, vMap );
        LevelMax = Abc_MaxInt(LevelMax, LevelCur);
    }
    //printf( "Max level %d\n", LevelMax );
    return vMap;
} 

/**Function*************************************************************

  Synopsis    [Returns representative node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivRepr(p *Gia_Man_t, pObj *Gia_Obj_t, fUseAll int32, fDualOut int32) *Gia_Obj_t {
    if fUseAll != 0 {
        if Gia_ObjRepr(p, Gia_ObjId(p,pObj)) == GIA_VOID {
            return nil; }
    
    } else {
        if !Gia_ObjProved(p, Gia_ObjId(p,pObj)) {
            return nil; }
    }
//    if ( fDualOut && !Gia_ObjDiffColors( p, Gia_ObjId(p, pObj), Gia_ObjRepr(p, Gia_ObjId(p,pObj)) ) )
    if (fDualOut != 0) && !Gia_ObjDiffColors2( p, Gia_ObjId(p, pObj), Gia_ObjRepr(p, Gia_ObjId(p,pObj)) ) {
        return nil; }
    return Gia_ManObj( p, Gia_ObjRepr(p, Gia_ObjId(p,pObj)) );
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivReduce_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, fUseAll int32, fDualOut int32) {
    var pRepr *Gia_Obj_t;
    if pRepr = Gia_ManEquivRepr(p, pObj, fUseAll, fDualOut) {
        Gia_ManEquivReduce_rec( pNew, p, pRepr, fUseAll, fDualOut );
        pObj.Value = Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) );
        return;
    }
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManEquivReduce_rec( pNew, p, Gia_ObjFanin0(pObj), fUseAll, fDualOut );
    Gia_ManEquivReduce_rec( pNew, p, Gia_ObjFanin1(pObj), fUseAll, fDualOut );
    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Reduces AIG using equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivReduce(p *Gia_Man_t, fUseAll int32, fDualOut int32, fSkipPhase int32, fVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    if !p.pReprs && p.pSibls {
        // inlining of macro ABC_FALLOC
        var pMap *int32  = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p))));
        // inlining of macro ABC_CALLOC
        p.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Gia_Rpr_t)));
        for  i = 0; i < Gia_ManObjNum(p); i++  {
            Gia_ObjSetRepr( p, i, GIA_VOID );
        }
        for  i = 0; i < Gia_ManObjNum(p); i++  {
            if p.pSibls[i] > 0 {
                if pMap[p.pSibls[i]] == -1 {
                    pMap[p.pSibls[i]] = p.pSibls[i]; }
                pMap[i] = pMap[p.pSibls[i]];
            }
        }
        for  i = 0; i < Gia_ManObjNum(p); i++  {
            if p.pSibls[i] > 0 {
                Gia_ObjSetRepr( p, i, pMap[i] ); }
        }
        //printf( "Created equivalence classes.\n" );
        // inlining of macro ABC_FREE
        p.pNexts = nil;
        p.pNexts = Gia_ManDeriveNexts( p );
        // inlining of macro ABC_FREE
        pMap = nil;
    }
    if !p.pReprs {
        Abc_Print( 1, "Gia_ManEquivReduce(): Equivalence classes are not available.\n" );
        return nil;
    }
    if (fDualOut != 0) && (Gia_ManPoNum(p) & 1) {
        Abc_Print( 1, "Gia_ManEquivReduce(): Dual-output miter should have even number of POs.\n" );
        return nil;
    }
    // check if there are any equivalences defined
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjReprObj(p, i) != nil {
            break; }
    }
    if i == Gia_ManObjNum(p) {
//        Abc_Print( 1, "Gia_ManEquivReduce(): There are no equivalences to reduce.\n" );
//        return NULL;
        return Gia_ManDup( p );
    }
/*
    if ( !Gia_ManCheckTopoOrder( p ) )
    {
        Abc_Print( 1, "Gia_ManEquivReduce(): AIG is not in a correct topological order.\n" );
        return NULL;
    }
*/
    if !(fSkipPhase != 0) {
        Gia_ManSetPhase( p ); }
    if fDualOut != 0 {
        Gia_ManEquivSetColors( p, fVerbose ); }
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManEquivReduce_rec( pNew, p, Gia_ObjFanin0(pObj), fUseAll, fDualOut );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivReduce2_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, vMap *Vec_Int_t, fDiveIn int32) {
    var pRepr *Gia_Obj_t;
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if fDiveIn && (pRepr = Gia_ManEquivRepr(p, pObj, 1, 0)) {
        var iTemp int32;
        var iRepr int32 = Gia_ObjId(p, pRepr);
        var pRepr2 *Gia_Obj_t  = Gia_ManObj( p, Vec_IntEntry(vMap, iRepr) );
        Gia_ManEquivReduce2_rec( pNew, p, pRepr2, vMap, 0 );
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, iRepr) && (iRepr != 0), "Gia_ObjIsHead(p, iRepr) && iRepr"), iTemp = iRepr; iTemp > 0; iTemp = Gia_ObjNext(p, iTemp)  {
            var pTemp *Gia_Obj_t  = Gia_ManObj(p, iTemp);
            pTemp.Value = Abc_LitNotCond( pRepr2.Value, Gia_ObjPhaseReal(pRepr2) ^ Gia_ObjPhaseReal(pTemp) );
        }
        assert(~pObj.Value, "~pObj.Value");
        assert(~pRepr.Value, "~pRepr.Value");
        assert(~pRepr2.Value, "~pRepr2.Value");
        return;
    }
    Gia_ManEquivReduce2_rec( pNew, p, Gia_ObjFanin0(pObj), vMap, 1 );
    Gia_ManEquivReduce2_rec( pNew, p, Gia_ObjFanin1(pObj), vMap, 1 );
    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}
func Gia_ManEquivReduce2(p *Gia_Man_t) *Gia_Man_t {
    var vMap *Vec_Int_t;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    if !p.pReprs && p.pSibls {
        // inlining of macro ABC_FALLOC
        var pMap *int32  = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p))));
        // inlining of macro ABC_CALLOC
        p.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Gia_Rpr_t)));
        for  i = 0; i < Gia_ManObjNum(p); i++  {
            Gia_ObjSetRepr( p, i, GIA_VOID );
        }
        for  i = 0; i < Gia_ManObjNum(p); i++  {
            if p.pSibls[i] > 0 {
                if pMap[p.pSibls[i]] == -1 {
                    pMap[p.pSibls[i]] = p.pSibls[i]; }
                pMap[i] = pMap[p.pSibls[i]];
            }
        }
        for  i = 0; i < Gia_ManObjNum(p); i++  {
            if p.pSibls[i] > 0 {
                Gia_ObjSetRepr( p, i, pMap[i] ); }
        }
        //printf( "Created equivalence classes.\n" );
        // inlining of macro ABC_FREE
        p.pNexts = nil;
        p.pNexts = Gia_ManDeriveNexts( p );
        // inlining of macro ABC_FREE
        pMap = nil;
    }
    if !p.pReprs {
        Abc_Print( 1, "Gia_ManEquivReduce(): Equivalence classes are not available.\n" );
        return nil;
    }
    // check if there are any equivalences defined
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjReprObj(p, i) != nil {
            break; }
    }
    if i == Gia_ManObjNum(p) {
        return Gia_ManDup( p ); }
    vMap = Gia_ManChoiceMinLevel( p );
    Gia_ManSetPhase( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManEquivReduce2_rec( pNew, p, Gia_ObjFanin0(pObj), vMap, 1 );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    Vec_IntFree( vMap );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Reduces AIG using equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivFixOutputPairs(p *Gia_Man_t) {
    var pObj0, pObj1 *Gia_Obj_t;
    var i int32;
    assert((Gia_ManPoNum(p) & 1) == 0, "(Gia_ManPoNum(p) & 1) == 0");
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj0) = Gia_ManCo(p, i)); i++  {
        pObj1 = Gia_ManPo( p, ++i );
        if Gia_ObjChild0(pObj0) != Gia_ObjChild0(pObj1) {
            continue; }
        pObj0.iDiff0  = Gia_ObjId(p, pObj0);
        pObj0.fCompl0 = 0;
        pObj1.iDiff0  = Gia_ObjId(p, pObj1);
        pObj1.fCompl0 = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Removes pointers to the unmarked nodes..]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivUpdatePointers(p *Gia_Man_t, pNew *Gia_Man_t) {
    var pObj, pObjNew *Gia_Obj_t;
    var i int32;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if !~pObj.Value {
            continue; }
        pObjNew = Gia_ManObj( pNew, Abc_Lit2Var(pObj.Value) );
        if pObjNew.fMark0 {
            pObj.Value = ~0; }
    }
}

/**Function*************************************************************

  Synopsis    [Removes pointers to the unmarked nodes..]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivDeriveReprs(p *Gia_Man_t, pNew *Gia_Man_t, pFinal *Gia_Man_t) {
    var vClass *Vec_Int_t;
    var pObj, pObjNew *Gia_Obj_t;
    var i, k, iNode, iRepr, iPrev int32;
    // start representatives
    // inlining of macro ABC_CALLOC
    pFinal.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(pFinal)), sizeof(Gia_Rpr_t)));
    for  i = 0; i < Gia_ManObjNum(pFinal); i++  {
        Gia_ObjSetRepr( pFinal, i, GIA_VOID );
    }
    // iterate over constant candidates
    // inlining of macro Gia_ManForEachConst
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsConst(p, i) {} else {
        pObj = Gia_ManObj( p, i );
        if !~pObj.Value {
            continue; }
        pObjNew = Gia_ManObj( pNew, Abc_Lit2Var(pObj.Value) );
        if Abc_Lit2Var(pObjNew.Value) == 0 {
            continue; }
        Gia_ObjSetRepr( pFinal, Abc_Lit2Var(pObjNew.Value), 0 );
    }
    }
    // iterate over class candidates
    vClass = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        Vec_IntClear( vClass );
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, i) && (i != 0), "Gia_ObjIsHead(p, i) && i"), k = i; k > 0; k = Gia_ObjNext(p, k)  {
            pObj = Gia_ManObj( p, k );
            if !~pObj.Value {
                continue; }
            pObjNew = Gia_ManObj( pNew, Abc_Lit2Var(pObj.Value) );
            Vec_IntPushUnique( vClass, Abc_Lit2Var(pObjNew.Value) );
        }
        if Vec_IntSize( vClass ) < 2 {
            continue; }
        Vec_IntSort( vClass, 0 );
        iRepr = iPrev = Vec_IntEntry( vClass, 0 );
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vClass)) && (((iNode) = Vec_IntEntry(vClass, k)), 1); k++  {
            Gia_ObjSetRepr( pFinal, iNode, iRepr );
            assert(iPrev < iNode, "iPrev < iNode");
            iPrev = iNode;
        }
    }
    }
    Vec_IntFree( vClass );
    pFinal.pNexts = Gia_ManDeriveNexts( pFinal );
}

/**Function*************************************************************

  Synopsis    [Removes pointers to the unmarked nodes..]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivRemapDfs(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var vClass *Vec_Int_t;
    var i, k, iNode, iRepr, iPrev int32;
    pNew = Gia_ManDupDfs( p );
    // start representatives
    // inlining of macro ABC_CALLOC
    pNew.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(pNew)), sizeof(Gia_Rpr_t)));
    for  i = 0; i < Gia_ManObjNum(pNew); i++  {
        Gia_ObjSetRepr( pNew, i, GIA_VOID );
    }
    // iterate over constant candidates
    // inlining of macro Gia_ManForEachConst
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsConst(p, i) {} else {
        Gia_ObjSetRepr( pNew, Abc_Lit2Var(Gia_ManObj(p, i).Value), 0 ); }
    }
    // iterate over class candidates
    vClass = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        Vec_IntClear( vClass );
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, i) && (i != 0), "Gia_ObjIsHead(p, i) && i"), k = i; k > 0; k = Gia_ObjNext(p, k)  {
            Vec_IntPushUnique( vClass, Abc_Lit2Var(Gia_ManObj(p, k).Value) );
        }
        assert(Vec_IntSize( vClass ) > 1, "Vec_IntSize( vClass ) > 1");
        Vec_IntSort( vClass, 0 );
        iRepr = iPrev = Vec_IntEntry( vClass, 0 );
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vClass)) && (((iNode) = Vec_IntEntry(vClass, k)), 1); k++  {
            Gia_ObjSetRepr( pNew, iNode, iRepr );
            assert(iPrev < iNode, "iPrev < iNode");
            iPrev = iNode;
        }
    }
    }
    Vec_IntFree( vClass );
    pNew.pNexts = Gia_ManDeriveNexts( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Reduces AIG while remapping equivalence classes.]

  Description [Drops the pairs of outputs if they are proved equivalent.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivReduceAndRemap(p *Gia_Man_t, fSeq int32, fMiterPairs int32) *Gia_Man_t {
    var pNew, pFinal *Gia_Man_t;
    pNew = Gia_ManEquivReduce( p, 0, 0, 0, 0 );
    if pNew == nil {
        return nil; }
    Gia_ManOrigIdsRemap( p, pNew );
    if fMiterPairs != 0 {
        Gia_ManEquivFixOutputPairs( pNew ); }
    if fSeq != 0 {
        Gia_ManSeqMarkUsed( pNew );
    } else {
        Gia_ManCombMarkUsed( pNew ); }
    Gia_ManEquivUpdatePointers( p, pNew );
    pFinal = Gia_ManDupMarked( pNew );
    Gia_ManOrigIdsRemap( pNew, pFinal );
    Gia_ManEquivDeriveReprs( p, pNew, pFinal );
    Gia_ManStop( pNew );
    pFinal = Gia_ManEquivRemapDfs( pNew = pFinal );
    Gia_ManOrigIdsRemap( pNew, pFinal );
    Gia_ManStop( pNew );
    return pFinal;
}

/**Function*************************************************************

  Synopsis    [Marks CIs/COs/ANDs unreachable from POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivSetColor_rec(p *Gia_Man_t, pObj *Gia_Obj_t, fOdds int32) int32 {
    if Gia_ObjVisitColor( p, Gia_ObjId(p,pObj), fOdds ) {
        return 0; }
    if Gia_ObjIsRo(p, pObj) {
        return 1 + Gia_ManEquivSetColor_rec( p, Gia_ObjFanin0(Gia_ObjRoToRi(p, pObj)), fOdds ); }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    return 1 + Gia_ManEquivSetColor_rec( p, Gia_ObjFanin0(pObj), fOdds )
             + Gia_ManEquivSetColor_rec( p, Gia_ObjFanin1(pObj), fOdds );
}

/**Function*************************************************************

  Synopsis    [Marks CIs/COs/ANDs unreachable from POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivSetColors(p *Gia_Man_t, fVerbose int32) int32 {
    var pObj *Gia_Obj_t;
    var i int32;
    var nNodes [2]int32;
    var nDiffs [2]int32;
    assert((Gia_ManPoNum(p) & 1) == 0, "(Gia_ManPoNum(p) & 1) == 0");
    Gia_ObjSetColors( p, 0 );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Gia_ObjSetColors( p, Gia_ObjId(p,pObj) );
    }
    nNodes[0] = nNodes[1] = Gia_ManPiNum(p);
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        nNodes[i&1] += Gia_ManEquivSetColor_rec( p, Gia_ObjFanin0(pObj), i&1 );
    }
//    Gia_ManForEachObj( p, pObj, i )
//        if ( Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj) )
//            assert( Gia_ObjColors(p, i) );
    nDiffs[0] = Gia_ManCandNum(p) - nNodes[0];
    nDiffs[1] = Gia_ManCandNum(p) - nNodes[1];
    if fVerbose != 0 {
        Abc_Print( 1, "CI+AND = %7d  A = %7d  B = %7d  Ad = %7d  Bd = %7d  AB = %7d.\n",
            Gia_ManCandNum(p), nNodes[0], nNodes[1], nDiffs[0], nDiffs[1],
            Gia_ManCandNum(p) - nDiffs[0] - nDiffs[1] );
    }
    return (nDiffs[0] + nDiffs[1]) / 2;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSpecBuild(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, vXorLits *Vec_Int_t, fDualOut int32, fSpeculate int32, vTrace *Vec_Int_t, vGuide *Vec_Int_t, vMap *Vec_Int_t) {
    var pRepr *Gia_Obj_t;
    var iLitNew uint32;
    pRepr = Gia_ObjReprObj( p, Gia_ObjId(p,pObj) );
    if pRepr == nil {
        return; }
//    if ( fDualOut && !Gia_ObjDiffColors( p, Gia_ObjId(p, pObj), Gia_ObjId(p, pRepr) ) )
    if (fDualOut != 0) && !Gia_ObjDiffColors2( p, Gia_ObjId(p, pObj), Gia_ObjId(p, pRepr) ) {
        return; }
    iLitNew = Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) );
    if pObj.Value != iLitNew && !Gia_ObjProved(p, Gia_ObjId(p,pObj)) {
        if vTrace != nil {
            Vec_IntPush( vTrace, 1 ); }
        if vGuide == nil || Vec_IntEntry( vGuide, Vec_IntSize(vTrace)-1 ) {
            if vMap != nil {  
                Vec_IntPush( vMap, Gia_ObjId(p, pObj) ); }
            Vec_IntPush( vXorLits, Gia_ManHashXor(pNew, pObj.Value, iLitNew) );
        }
    
    } else {
        if vTrace != nil {
            Vec_IntPush( vTrace, 0 ); }
    }
    if fSpeculate != 0 {
        pObj.Value = iLitNew; }
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSpecReduce_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, vXorLits *Vec_Int_t, fDualOut int32, fSpeculate int32, vTrace *Vec_Int_t, vGuide *Vec_Int_t, vMap *Vec_Int_t) {
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManSpecReduce_rec( pNew, p, Gia_ObjFanin0(pObj), vXorLits, fDualOut, fSpeculate, vTrace, vGuide, vMap );
    Gia_ManSpecReduce_rec( pNew, p, Gia_ObjFanin1(pObj), vXorLits, fDualOut, fSpeculate, vTrace, vGuide, vMap );
    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    Gia_ManSpecBuild( pNew, p, pObj, vXorLits, fDualOut, fSpeculate, vTrace, vGuide, vMap );
}

/**Function*************************************************************

  Synopsis    [Reduces AIG using equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSpecReduceTrace(p *Gia_Man_t, vTrace *Vec_Int_t, vMap *Vec_Int_t) *Gia_Man_t {
    var vXorLits *Vec_Int_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, iLitNew int32;
    if !p.pReprs {
        Abc_Print( 1, "Gia_ManSpecReduce(): Equivalence classes are not available.\n" );
        return nil;
    }
    Vec_IntClear( vTrace );
    vXorLits = Vec_IntAlloc( 1000 );
    Gia_ManSetPhase( p );
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ManSpecBuild( pNew, p, pObj, vXorLits, 0, 1, vTrace, nil, vMap );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManSpecReduce_rec( pNew, p, Gia_ObjFanin0(pObj), vXorLits, 0, 1, vTrace, nil, vMap );
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXorLits)) && (((iLitNew) = Vec_IntEntry(vXorLits, i)), 1); i++  {
        Gia_ManAppendCo( pNew, iLitNew );
    }
    if Vec_IntSize(vXorLits) == 0 {
        Abc_Print( 1, "Speculatively reduced model has no primary outputs.\n" );
        Gia_ManAppendCo( pNew, 0 );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Vec_IntFree( vXorLits );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Reduces AIG using equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSpecReduce(p *Gia_Man_t, fDualOut int32, fSynthesis int32, fSpeculate int32, fSkipSome int32, fVerbose int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vXorLits *Vec_Int_t;
    var i, iLitNew int32;
    var vTrace *Vec_Int_t = nil;
    var vGuide *Vec_Int_t = nil;
    if !p.pReprs {
        Abc_Print( 1, "Gia_ManSpecReduce(): Equivalence classes are not available.\n" );
        return nil;
    }
    if (fDualOut != 0) && (Gia_ManPoNum(p) & 1) {
        Abc_Print( 1, "Gia_ManSpecReduce(): Dual-output miter should have even number of POs.\n" );
        return nil;
    }
    if fSkipSome != 0 {
        vGuide = Vec_IntAlloc( 100 );
        pTemp  = Gia_ManSpecReduceTrace( p, vGuide, nil );
        Gia_ManStop( pTemp );
        assert(Vec_IntSize(vGuide) == Gia_ManEquivCountLitsAll(p), "Vec_IntSize(vGuide) == Gia_ManEquivCountLitsAll(p)");
        vTrace = Vec_IntAlloc( 100 );
    }
    vXorLits = Vec_IntAlloc( 1000 );
    Gia_ManSetPhase( p );
    Gia_ManFillValue( p );
    if fDualOut != 0 {
        Gia_ManEquivSetColors( p, fVerbose ); }
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ManSpecBuild( pNew, p, pObj, vXorLits, fDualOut, fSpeculate, vTrace, vGuide, nil );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManSpecReduce_rec( pNew, p, Gia_ObjFanin0(pObj), vXorLits, fDualOut, fSpeculate, vTrace, vGuide, nil );
    }
    if !(fSynthesis != 0) {
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXorLits)) && (((iLitNew) = Vec_IntEntry(vXorLits, i)), 1); i++  {
        Gia_ManAppendCo( pNew, iLitNew );
    }
    if Vec_IntSize(vXorLits) == 0 {
        Abc_Print( 1, "Speculatively reduced model has no primary outputs.\n" );
        Gia_ManAppendCo( pNew, 0 );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Vec_IntFree( vXorLits );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );

    // update using trace
    if fSkipSome != 0 {
        // count the number of non-zero entries
        var iLit, nAddPos int32;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vGuide)) && (((iLit) = Vec_IntEntry(vGuide, i)), 1); i++  {
            if iLit != 0 {
                nAddPos++; }
        }
        if nAddPos != 0 {
            assert(Gia_ManPoNum(pNew) == Gia_ManPoNum(p) + nAddPos, "Gia_ManPoNum(pNew) == Gia_ManPoNum(p) + nAddPos"); }
    }
    Vec_IntFreeP( &vTrace );
    Vec_IntFreeP( &vGuide );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSpecBuildInit(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, vXorLits *Vec_Int_t, f int32, fDualOut int32) {
    var pRepr *Gia_Obj_t;
    var iLitNew int32;
    pRepr = Gia_ObjReprObj( p, Gia_ObjId(p,pObj) );
    if pRepr == nil {
        return; }
//    if ( fDualOut && !Gia_ObjDiffColors( p, Gia_ObjId(p, pObj), Gia_ObjId(p, pRepr) ) )
    if (fDualOut != 0) && !Gia_ObjDiffColors2( p, Gia_ObjId(p, pObj), Gia_ObjId(p, pRepr) ) {
        return; }
    iLitNew = Abc_LitNotCond( Gia_ObjCopyF(p, f, pRepr), Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) );
    if Gia_ObjCopyF(p, f, pObj) != iLitNew && !Gia_ObjProved(p, Gia_ObjId(p,pObj)) {
        Vec_IntPush( vXorLits, Gia_ManHashXor(pNew, Gia_ObjCopyF(p, f, pObj), iLitNew) ); }
    Gia_ObjSetCopyF( p, f, pObj, iLitNew );
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSpecReduceInit_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, vXorLits *Vec_Int_t, f int32, fDualOut int32) {
    if ~Gia_ObjCopyF(p, f, pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManSpecReduceInit_rec( pNew, p, Gia_ObjFanin0(pObj), vXorLits, f, fDualOut );
    Gia_ManSpecReduceInit_rec( pNew, p, Gia_ObjFanin1(pObj), vXorLits, f, fDualOut );
    Gia_ObjSetCopyF( p, f, pObj, Gia_ManHashAnd(pNew, Gia_ObjFanin0CopyF(p, f, pObj), Gia_ObjFanin1CopyF(p, f, pObj)) );
    Gia_ManSpecBuildInit( pNew, p, pObj, vXorLits, f, fDualOut );
}

/**Function*************************************************************

  Synopsis    [Creates initialized SRM with the given number of frames.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSpecReduceInit(p *Gia_Man_t, pInit *Abc_Cex_t, nFrames int32, fDualOut int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var vXorLits *Vec_Int_t;
    var f, i, iLitNew int32;
    if !p.pReprs {
        Abc_Print( 1, "Gia_ManSpecReduceInit(): Equivalence classes are not available.\n" );
        return nil;
    }
    if Gia_ManRegNum(p) == 0 {
        Abc_Print( 1, "Gia_ManSpecReduceInit(): Circuit is not sequential.\n" );
        return nil;
    }
    if Gia_ManRegNum(p) != pInit.nRegs {
        Abc_Print( 1, "Gia_ManSpecReduceInit(): Mismatch in the number of registers.\n" );
        return nil;
    }
    if (fDualOut != 0) && (Gia_ManPoNum(p) & 1) {
        Abc_Print( 1, "Gia_ManSpecReduceInit(): Dual-output miter should have even number of POs.\n" );
        return nil;
    }

/*
    if ( !Gia_ManCheckTopoOrder( p ) )
    {
        Abc_Print( 1, "Gia_ManSpecReduceInit(): AIG is not in a correct topological order.\n" );
        return NULL;
    }
*/
    assert(pInit.nRegs == Gia_ManRegNum(p) && pInit.nPis == 0, "pInit.nRegs == Gia_ManRegNum(p) && pInit.nPis == 0");
    Vec_IntFill( &p.vCopies, nFrames * Gia_ManObjNum(p), -1 );
    vXorLits = Vec_IntAlloc( 1000 );
    Gia_ManSetPhase( p );
    if fDualOut != 0 {
        Gia_ManEquivSetColors( p, 0 ); }
    pNew = Gia_ManStart( nFrames * Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ObjSetCopyF( p, 0, pObj, Abc_InfoHasBit(pInit.pData, i) );
    }
    for  f = 0; f < nFrames; f++  {
        Gia_ObjSetCopyF( p, f, Gia_ManConst0(p), 0 );
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            Gia_ObjSetCopyF( p, f, pObj, Gia_ManAppendCi(pNew) );
        }
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            Gia_ManSpecBuildInit( pNew, p, pObj, vXorLits, f, fDualOut );
        }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            Gia_ManSpecReduceInit_rec( pNew, p, Gia_ObjFanin0(pObj), vXorLits, f, fDualOut );
            Gia_ObjSetCopyF( p, f, pObj, Gia_ObjFanin0CopyF(p, f, pObj) );
        }
        if f == nFrames - 1 {
            break; }
        // inlining of macro Gia_ManForEachRiRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            Gia_ObjSetCopyF( p, f+1, pObjRo, Gia_ObjCopyF(p, f, pObjRi) );
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXorLits)) && (((iLitNew) = Vec_IntEntry(vXorLits, i)), 1); i++  {
        Gia_ManAppendCo( pNew, iLitNew );
    }
    if Vec_IntSize(vXorLits) == 0 {
//        Abc_Print( 1, "Speculatively reduced model has no primary outputs.\n" );
        Gia_ManAppendCo( pNew, 0 );
    }
    Vec_IntErase( &p.vCopies );
    Vec_IntFree( vXorLits );
    Gia_ManHashStop( pNew );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Creates initialized SRM with the given number of frames.]

  Description [Uses as many frames as needed to create the number of 
  output not less than the number of equivalence literals.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSpecReduceInitFrames(p *Gia_Man_t, pInit *Abc_Cex_t, nFramesMax int32, pnFrames *int32, fDualOut int32, nMinOutputs int32) *Gia_Man_t {
    var pFrames *Gia_Man_t;
    var f, nLits int32;
    nLits = Gia_ManEquivCountLits( p );
    for  f = 1; ; f++  {
        pFrames = Gia_ManSpecReduceInit( p, pInit, f, fDualOut );
        if (nMinOutputs == 0 && Gia_ManPoNum(pFrames) >= nLits/2+1) ||
             (nMinOutputs != 0 && Gia_ManPoNum(pFrames) >= nMinOutputs) {
            break; }
        if f == nFramesMax {
            break; }
        if Gia_ManAndNum(pFrames) > 500000 {
            Gia_ManStop( pFrames );
            return nil;
        }
        Gia_ManStop( pFrames );
        pFrames = nil;
    }
    if f == nFramesMax {
        Abc_Print( 1, "Stopped unrolling after %d frames.\n", nFramesMax ); }
    if pnFrames != nil {
        *pnFrames = f; }
    return pFrames;
}

/**Function*************************************************************

  Synopsis    [Transforms equiv classes by removing the AB nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivTransform(p *Gia_Man_t, fVerbose int32) {
    extern void Cec_ManSimClassCreate( Gia_Man_t * p, Vec_Int_t * vClass );
    var vClass, vClassNew *Vec_Int_t;
    var iRepr, iNode, Ent, k int32;
    var nRemovedLits, nRemovedClas int32;
    var nTotalLits, nTotalClas int32;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(p.pReprs && p.pNexts, "p.pReprs && p.pNexts");
    vClass = Vec_IntAlloc( 100 );
    vClassNew = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) || Gia_ObjIsAnd(pObj) {
            assert(Gia_ObjColors(p, i), "Gia_ObjColors(p, i)"); }
    }
    // inlining of macro Gia_ManForEachClassReverse
    for  iRepr = Gia_ManObjNum(p) - 1; iRepr > 0; iRepr--  { if !Gia_ObjIsHead(p, iRepr) {} else {
        nTotalClas++;
        Vec_IntClear( vClass );
        Vec_IntClear( vClassNew );
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, iRepr) && (iRepr != 0), "Gia_ObjIsHead(p, iRepr) && iRepr"), iNode = iRepr; iNode > 0; iNode = Gia_ObjNext(p, iNode)  {
            nTotalLits++;
            Vec_IntPush( vClass, iNode );
            assert(Gia_ObjColors(p, iNode), "Gia_ObjColors(p, iNode)");
            if Gia_ObjColors(p, iNode) != 3 {
                Vec_IntPush( vClassNew, iNode );
            } else {
                nRemovedLits++; }
        }
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vClass)) && (((Ent) = Vec_IntEntry(vClass, k)), 1); k++  {
            p.pReprs[Ent].fFailed = p.pReprs[Ent].fProved = 0;
            p.pReprs[Ent].iRepr = GIA_VOID;
            p.pNexts[Ent] = 0;
        }
        if Vec_IntSize(vClassNew) < 2 {
            nRemovedClas++;
            continue;
        }
        Cec_ManSimClassCreate( p, vClassNew );
    }
    }
    Vec_IntFree( vClass );
    Vec_IntFree( vClassNew );
    if fVerbose != 0 {
    Abc_Print( 1, "Removed classes = %6d (out of %6d). Removed literals = %6d (out of %6d).\n",
        nRemovedClas, nTotalClas, nRemovedLits, nTotalLits ); }
}

/**Function*************************************************************

  Synopsis    [Marks proved equivalences.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivMark(p *Gia_Man_t, pFileName string, fSkipSome int32, fVerbose int32) {
    var pMiter, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, iLit, nAddPos, nLits int32;
    var nLitsAll, Counter int32;
    nLitsAll = Gia_ManEquivCountLitsAll( p );
    if nLitsAll == 0 {
        Abc_Print( 1, "Gia_ManEquivMark(): Current AIG does not have equivalences.\n" );
        return;
    }
    // read AIGER file
    pMiter = Gia_AigerRead( pFileName, 0, 0, 0 );
    if pMiter == nil {
        Abc_Print( 1, "Gia_ManEquivMark(): Input file %s could not be read.\n", pFileName );
        return;
    }
    if fSkipSome != 0 {
        var vTrace *Vec_Int_t  = Vec_IntAlloc( 100 );
        pTemp = Gia_ManSpecReduceTrace( p, vTrace, nil );
        Gia_ManStop( pTemp );
        assert(Vec_IntSize(vTrace) == nLitsAll, "Vec_IntSize(vTrace) == nLitsAll");
        // count the number of non-zero entries
        nAddPos = 0;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vTrace)) && (((iLit) = Vec_IntEntry(vTrace, i)), 1); i++  {
            if iLit != 0 {
                nAddPos++; }
        }
        // check the number
        if Gia_ManPoNum(pMiter) != Gia_ManPoNum(p) + nAddPos {
            Abc_Print( 1, "Gia_ManEquivMark(): The number of POs is not correct: MiterPONum(%d) != AIGPONum(%d) + AIGFilteredEquivNum(%d).\n",
                Gia_ManPoNum(pMiter), Gia_ManPoNum(p), nAddPos );
            Gia_ManStop( pMiter );
            Vec_IntFreeP( &vTrace );
            return;
        }
        // mark corresponding POs as solved
        nLits = iLit = Counter = 0;
        for  i = 0; i < Gia_ManObjNum(p); i++  {
            if Gia_ObjRepr(p, i) == GIA_VOID {
                continue; }
            if Vec_IntEntry( vTrace, nLits++ ) == 0 {
                continue; }
            pObj = Gia_ManPo( pMiter, Gia_ManPoNum(p) + iLit++ );
            if Gia_ObjFaninLit0p(pMiter, pObj) == 0 { // const 0 - proven
            
                Gia_ObjSetProved(p, i);
                Counter++;
            }
        }
        assert(nLits == nLitsAll, "nLits == nLitsAll");
        assert(iLit == nAddPos, "iLit == nAddPos");
        Vec_IntFreeP( &vTrace );
    
    } else {
        if Gia_ManPoNum(pMiter) != Gia_ManPoNum(p) + nLitsAll {
            Abc_Print( 1, "Gia_ManEquivMark(): The number of POs is not correct: MiterPONum(%d) != AIGPONum(%d) + AIGEquivNum(%d).\n",
                Gia_ManPoNum(pMiter), Gia_ManPoNum(p), nLitsAll );
            Gia_ManStop( pMiter );
            return;
        }
        // mark corresponding POs as solved
        nLits = 0;
        for  i = 0; i < Gia_ManObjNum(p); i++  {
            if Gia_ObjRepr(p, i) == GIA_VOID {
                continue; }
            pObj = Gia_ManPo( pMiter, Gia_ManPoNum(p) + nLits++ );
            if Gia_ObjFaninLit0p(pMiter, pObj) == 0 { // const 0 - proven
            
                Gia_ObjSetProved(p, i);
                Counter++;
            }
        }
        assert(nLits == nLitsAll, "nLits == nLitsAll");
    }
    if fVerbose != 0 {
        Abc_Print( 1, "Set %d equivalences as proved.\n", Counter ); }
    Gia_ManStop( pMiter );
}

/**Function*************************************************************

  Synopsis    [Transforms equiv classes by filtering those that correspond to disproved outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivFilter(p *Gia_Man_t, vPoIds *Vec_Int_t, fVerbose int32) {
    var pSrm *Gia_Man_t;
    var vTrace, vMap *Vec_Int_t;
    var i int32;
    var iObjId int32;
    var Entry int32;
    var Prev int32 = -1;
    // check if there are equivalences
    if p.pReprs == nil || p.pNexts == nil {
        Abc_Print( 1, "Gia_ManEquivFilter(): Equivalence classes are not defined.\n" );
        return;
    }
    // check if PO indexes are available
    if vPoIds == nil {
        Abc_Print( 1, "Gia_ManEquivFilter(): Array of disproved POs is not available.\n" );
        return;
    }
    if Vec_IntSize(vPoIds) == 0 {
        return; }
    // create SRM with mapping into POs
    vMap = Vec_IntAlloc( 1000 );
    vTrace = Vec_IntAlloc( 1000 );
    pSrm = Gia_ManSpecReduceTrace( p, vTrace, vMap );
    Vec_IntFree( vTrace );
    // the resulting array (vMap) maps PO indexes of the SRM into object IDs
    assert(Gia_ManPoNum(pSrm) == Gia_ManPoNum(p) + Vec_IntSize(vMap), "Gia_ManPoNum(pSrm) == Gia_ManPoNum(p) + Vec_IntSize(vMap)");
    Gia_ManStop( pSrm );
    if fVerbose != 0 {
        fmt.Printf("Design POs = %d. SRM POs = %d. Spec POs = %d. Disproved POs = %d.\n", Gia_ManPoNum(p), Gia_ManPoNum(p) + Vec_IntSize(vMap), Vec_IntSize(vMap), Vec_IntSize(vPoIds)); }
    // check if disproved POs satisfy the range
    Vec_IntSort( vPoIds, 0 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPoIds)) && (((Entry) = Vec_IntEntry(vPoIds, i)), 1); i++  {
        if Entry < 0 || Entry >= Gia_ManPoNum(p) + Vec_IntSize(vMap) {
            Abc_Print( 1, "Gia_ManEquivFilter(): Array of disproved POs contains PO index (%d),\n", Entry );
            Abc_Print( 1, "which does not fit into the range of available PO indexes of the SRM: [%d; %d].\n", 0, Gia_ManPoNum(p) + Vec_IntSize(vMap)-1 );
            Vec_IntFree( vMap );
            return;
        }
        if Entry < Gia_ManPoNum(p) {
            Abc_Print( 0, "Gia_ManEquivFilter(): One of the original POs (%d) have failed.\n", Entry ); }
        if Prev == Entry {
            Abc_Print( 1, "Gia_ManEquivFilter(): Array of disproved POs contains at least one duplicate entry (%d),\n", Entry );
            Vec_IntFree( vMap );
            return;
        }
        Prev = Entry;
    }
    // perform the reduction of the equivalence classes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPoIds)) && (((Entry) = Vec_IntEntry(vPoIds, i)), 1); i++  {
        if Entry < Gia_ManPoNum(p) {
            continue; }
        iObjId = Vec_IntEntry( vMap, Entry - Gia_ManPoNum(p) );
        Gia_ObjUnsetRepr( p, iObjId );
//        Gia_ObjSetNext( p, iObjId, 0 );
    }
    Vec_IntFree( vMap );
    // inlining of macro ABC_FREE
    p.pNexts = nil;
    p.pNexts = Gia_ManDeriveNexts( p );
}

/**Function*************************************************************

  Synopsis    [Transforms equiv classes by filtering those that correspond to disproved outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivFilterTest(p *Gia_Man_t) {
    var vPoIds *Vec_Int_t;
    var i int32;
    vPoIds = Vec_IntAlloc( 1000 );
    for  i = 0; i < 10; i++  {
        Vec_IntPush( vPoIds, Gia_ManPoNum(p) + 2 * i + 2 );
        fmt.Printf("%d ", Gia_ManPoNum(p) + 2*i + 2);
    }
    fmt.Printf("\n");
    Gia_ManEquivFilter( p, vPoIds, 1 );
    Vec_IntFree( vPoIds );
}

/**Function*************************************************************

  Synopsis    [Transforms equiv classes by setting a good representative.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivImprove(p *Gia_Man_t) {
    var vClass *Vec_Int_t;
    var i, k, iNode, iRepr int32;
    var iReprBest, iLevelBest, iLevelCur, iMffcBest, iMffcCur int32;
    assert(p.pReprs != nil && p.pNexts != nil, "p.pReprs != nil && p.pNexts != nil");
    Gia_ManLevelNum( p );
    Gia_ManCreateRefs( p );
    // iterate over class candidates
    vClass = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        Vec_IntClear( vClass );
        iReprBest = -1;
        iLevelBest = iMffcBest = ABC_INFINITY;
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, i) && (i != 0), "Gia_ObjIsHead(p, i) && i"), k = i; k > 0; k = Gia_ObjNext(p, k)  {
            iLevelCur = Gia_ObjLevel( p,Gia_ManObj(p, k) );
            iMffcCur  = Gia_NodeMffcSize( p, Gia_ManObj(p, k) );
            if iLevelBest > iLevelCur || (iLevelBest == iLevelCur && iMffcBest > iMffcCur) {
                iReprBest  = k;
                iLevelBest = iLevelCur;
                iMffcBest  = iMffcCur;
            }
            Vec_IntPush( vClass, k );
        }
        assert(Vec_IntSize( vClass ) > 1, "Vec_IntSize( vClass ) > 1");
        assert(iReprBest > 0, "iReprBest > 0");
        if i == iReprBest {
            continue; }
/*
        Abc_Print( 1, "Repr/Best = %6d/%6d. Lev = %3d/%3d. Mffc = %3d/%3d.\n", 
            i, iReprBest, Gia_ObjLevel( p,Gia_ManObj(p, i) ), Gia_ObjLevel( p,Gia_ManObj(p, iReprBest) ),
            Gia_NodeMffcSize( p, Gia_ManObj(p, i) ), Gia_NodeMffcSize( p, Gia_ManObj(p, iReprBest) ) );
*/
        iRepr = iReprBest;
        Gia_ObjSetRepr( p, iRepr, GIA_VOID );
        Gia_ObjSetProved( p, i );
        Gia_ObjUnsetProved( p, iRepr );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vClass)) && (((iNode) = Vec_IntEntry(vClass, k)), 1); k++  {
            if iNode != iRepr {
                Gia_ObjSetRepr( p, iNode, iRepr ); }
        }
    }
    }
    Vec_IntFree( vClass );
    // inlining of macro ABC_FREE
    p.pNexts = nil;
//    p->pNexts = Gia_ManDeriveNexts( p );
}


/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNode.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjCheckTfi_rec(p *Gia_Man_t, pOld *Gia_Obj_t, pNode *Gia_Obj_t, vVisited *Vec_Ptr_t) int32 {
    // check the trivial cases
    if pNode == nil {
        return 0; }
    if Gia_ObjIsCi(pNode) {
        return 0; }
//    if ( pNode->Id < pOld->Id ) // cannot use because of choices of pNode
//        return 0;
    if pNode == pOld {
        return 1; }
    // skip the visited node
    if pNode.fMark0 {
        return 0; }
    pNode.fMark0 = 1;
    Vec_PtrPush( vVisited, pNode );
    // check the children
    if Gia_ObjCheckTfi_rec( p, pOld, Gia_ObjFanin0(pNode), vVisited ) != 0 {
        return 1; }
    if Gia_ObjCheckTfi_rec( p, pOld, Gia_ObjFanin1(pNode), vVisited ) != 0 {
        return 1; }
    // check equivalent nodes
    return Gia_ObjCheckTfi_rec( p, pOld, Gia_ObjNextObj(p, Gia_ObjId(p, pNode)), vVisited );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNode.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjCheckTfi(p *Gia_Man_t, pOld *Gia_Obj_t, pNode *Gia_Obj_t) int32 {
    var vVisited *Vec_Ptr_t;
    var pObj *Gia_Obj_t;
    var RetValue, i int32;
    assert(!Gia_IsComplement(pOld), "!Gia_IsComplement(pOld)");
    assert(!Gia_IsComplement(pNode), "!Gia_IsComplement(pNode)");
    vVisited = Vec_PtrAlloc( 100 );
    RetValue = Gia_ObjCheckTfi_rec( p, pOld, pNode, vVisited );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vVisited)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vVisited, i)), 1); i++  {
        pObj.fMark0 = 0;
    }
    Vec_PtrFree( vVisited );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Adds the next entry while making choices.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAddNextEntry_rec(p *Gia_Man_t, pOld *Gia_Obj_t, pNode *Gia_Obj_t) {
    if Gia_ObjNext(p, Gia_ObjId(p, pOld)) == 0 {
        Gia_ObjSetNext( p, Gia_ObjId(p, pOld), Gia_ObjId(p, pNode) );
        return;
    }
    Gia_ManAddNextEntry_rec( p, Gia_ObjNextObj(p, Gia_ObjId(p, pOld)), pNode );
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivToChoices_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    var pRepr, pReprNew, pObjNew *Gia_Obj_t;
    if ~pObj.Value {
        return; }
    if pRepr = Gia_ObjReprObj(p, Gia_ObjId(p, pObj)) {
        if Gia_ObjIsConst0(pRepr) {
            pObj.Value = Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) );
            return;
        }
        Gia_ManEquivToChoices_rec( pNew, p, pRepr );
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        Gia_ManEquivToChoices_rec( pNew, p, Gia_ObjFanin0(pObj) );
        Gia_ManEquivToChoices_rec( pNew, p, Gia_ObjFanin1(pObj) );
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        if Abc_LitRegular(pObj.Value) == Abc_LitRegular(pRepr.Value) {
            assert(int32(pObj.Value) == Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) ), "int32(pObj.Value) == Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) )");
            return;
        }
        if pRepr.Value > pObj.Value { // should never happen with high resource limit
            return; }
        assert(pRepr.Value < pObj.Value, "pRepr.Value < pObj.Value");
        pReprNew = Gia_ManObj( pNew, Abc_Lit2Var(pRepr.Value) );
        pObjNew  = Gia_ManObj( pNew, Abc_Lit2Var(pObj.Value) );
        if Gia_ObjReprObj( pNew, Gia_ObjId(pNew, pObjNew) ) {
//            assert( Gia_ObjReprObj( pNew, Gia_ObjId(pNew, pObjNew) ) == pReprNew );
            if Gia_ObjReprObj( pNew, Gia_ObjId(pNew, pObjNew) ) != pReprNew {
                return; }
            pObj.Value = Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) );
            return;
        }
        if !(Gia_ObjCheckTfi( pNew, pReprNew, pObjNew ) != 0) {
            assert(Gia_ObjNext(pNew, Gia_ObjId(pNew, pObjNew)) == 0, "Gia_ObjNext(pNew, Gia_ObjId(pNew, pObjNew)) == 0");
            Gia_ObjSetRepr( pNew, Gia_ObjId(pNew, pObjNew), Gia_ObjId(pNew, pReprNew) );
            Gia_ManAddNextEntry_rec( pNew, pReprNew, pObjNew );
        }
        pObj.Value = Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManEquivToChoices_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Gia_ManEquivToChoices_rec( pNew, p, Gia_ObjFanin1(pObj) );
    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Removes choices, which contain fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRemoveBadChoices(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i, iObj, iPrev, Counter int32;
    // mark nodes with fanout
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        pObj.fMark0 = 0;
        if Gia_ObjIsAnd(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1;
            Gia_ObjFanin1(pObj).fMark0 = 1;
        
        } else if Gia_ObjIsCo(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1; }
    }
    // go through the classes and remove 
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        for  iPrev = i, iObj = Gia_ObjNext(p, i); (iObj != 0); iObj = Gia_ObjNext(p, iPrev)  {
            if !Gia_ManObj(p, iObj).fMark0 {
                iPrev = iObj;
                continue;
            }
            Gia_ObjSetRepr( p, iObj, GIA_VOID );
            Gia_ObjSetNext( p, iPrev, Gia_ObjNext(p, iObj) );
            Gia_ObjSetNext( p, iObj, 0 );
            Counter++;
        }
    }
    }
    // remove the marks
    Gia_ManCleanMark0( p );
//    Abc_Print( 1, "Removed %d bad choices.\n", Counter );
}

/**Function*************************************************************

  Synopsis    [Reduces AIG using equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEquivToChoices(p *Gia_Man_t, nSnapshots int32) *Gia_Man_t {
    var vNodes *Vec_Int_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj, pRepr *Gia_Obj_t;
    var i int32;
//Gia_ManEquivPrintClasses( p, 0, 0 );
    assert((Gia_ManCoNum(p) % nSnapshots) == 0, "(Gia_ManCoNum(p) % nSnapshots) == 0");
    Gia_ManSetPhase( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.pName = Abc_UtilStrsav( p.pName );
    // inlining of macro ABC_CALLOC
    pNew.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Gia_Rpr_t)));
    // inlining of macro ABC_CALLOC
    pNew.pNexts = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        pNew.pReprs[i].iRepr = GIA_VOID;
    }
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if pRepr = Gia_ObjReprObj(p, Gia_ObjId(p, pObj)) {
            assert(Gia_ObjIsConst0(pRepr) || Gia_ObjIsRo(p, pRepr), "Gia_ObjIsConst0(pRepr) || Gia_ObjIsRo(p, pRepr)");
            pObj.Value = pRepr.Value;
        }
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManEquivToChoices_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    vNodes = Gia_ManGetDangling( p );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
        Gia_ManEquivToChoices_rec( pNew, p, pObj );
    }
    Vec_IntFree( vNodes );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if i % nSnapshots == 0 {
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    Gia_ManRemoveBadChoices( pNew );
//Gia_ManEquivPrintClasses( pNew, 0, 0 );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
//Gia_ManEquivPrintClasses( pNew, 0, 0 );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Counts the number of choice nodes]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountChoiceNodes(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    if p.pReprs == nil || p.pNexts == nil {
        return 0; }
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        Counter += Gia_ObjIsHead( p, i );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of choices]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountChoices(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    if p.pReprs == nil || p.pNexts == nil {
        return 0; }
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        Counter += int32((Gia_ObjNext( p, i ) > 0));
    }
    return Counter;
}

//ABC_NAMESPACE_IMPL_END

//#include "aig/aig/aig.h"
//#include "aig/saig/saig.h"
//#include "proof/cec/cec.h"
//#include "giaAig.h"

//ABC_NAMESPACE_IMPL_START


/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has dangling nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManHasNoEquivs(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i int32;
    if p.pReprs == nil {
        return 1; }
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjReprObj(p, i) != nil {
            break; }
    }
    return i == Gia_ManObjNum(p);
}

/**Function*************************************************************

  Synopsis    [Implements iteration during speculation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_CommandSpecI(pGia *Gia_Man_t, nFramesInit int32, nBTLimitInit int32, fStart int32, fCheckMiter int32, fVerbose int32) int32 {
//    extern int Cec_ManCheckNonTrivialCands( Gia_Man_t * pAig );
    var pTemp *Aig_Man_t;
    var pSrm, pReduce, pAux *Gia_Man_t;
    var nIter, nStart int32;
    if pGia.pReprs == nil || pGia.pNexts == nil {
        Abc_Print( 1, "Gia_CommandSpecI(): Equivalence classes are not defined.\n" );
        return 0;
    }
    // (spech)*  where spech = &srm; restore save3; bmc2 -F 100 -C 25000; &resim
    Gia_ManCleanMark0( pGia );
    Gia_ManPrintStats( pGia, nil );
    for  nIter = 0; ; nIter++  {
        if Gia_ManHasNoEquivs(pGia) != 0 {
            Abc_Print( 1, "Gia_CommandSpecI: No equivalences left.\n" );
            break;
        }
        Abc_Print( 1, "ITER %3d : ", nIter );
//      if ( fVerbose )
//            Abc_Print( 1, "Starting BMC from frame %d.\n", nStart );
//      if ( fVerbose )
//            Gia_ManPrintStats( pGia, 0 );
            Gia_ManPrintStatsClasses( pGia );
        // perform speculative reduction
//        if ( Gia_ManPoNum(pSrm) <= Gia_ManPoNum(pGia) )
        if !Cec_ManCheckNonTrivialCands(pGia) {
            Abc_Print( 1, "Gia_CommandSpecI: There are only trivial equiv candidates left (PO drivers). Quitting.\n" );
            break;
        }
        pSrm = Gia_ManSpecReduce( pGia, 0, 0, 1, 0, 0 );
        // bmc2 -F 100 -C 25000
        {
            var pCex *Abc_Cex_t;
            var nFrames int32      = nFramesInit; // different from default
            var nNodeDelta int32   = 2000;
            var nBTLimit int32     = nBTLimitInit; // different from default
            var nBTLimitAll int32  = 2000000;
            pTemp = Gia_ManToAig( pSrm, 0 );
//            Aig_ManPrintStats( pTemp );
            Gia_ManStop( pSrm );
            Saig_BmcPerform( pTemp, nStart, nFrames, nNodeDelta, 0, nBTLimit, nBTLimitAll, fVerbose, 0, nil, 0, 0 );
            pCex = pTemp.pSeqModel; pTemp.pSeqModel = nil;
            Aig_ManStop( pTemp );
            if pCex == nil {
                Abc_Print( 1, "Gia_CommandSpecI(): Internal BMC could not find a counter-example.\n" );
                break;
            }
            if fStart != 0 {
                nStart = pCex.IFrame; }
            // perform simulation
            {
                var Pars Cec_ParSim_t;
                var pPars *Cec_ParSim_t = &Pars;
                Cec_ManSimSetDefaultParams( pPars );
                pPars.fCheckMiter = fCheckMiter;
                if Cec_ManSeqResimulateCounter( pGia, pPars, pCex ) {
                    // inlining of macro ABC_FREE
                    pCex = nil;
                    break;
                }
                // inlining of macro ABC_FREE
                pCex = nil;
            }
        }
        // write equivalence classes
        Gia_AigerWrite( pGia, "gore.aig", 0, 0, 0 );
        // reduce the model
        pReduce = Gia_ManSpecReduce( pGia, 0, 0, 1, 0, 0 );
        if pReduce {
            pReduce = Gia_ManSeqStructSweep( pAux = pReduce, 1, 1, 0 );
            Gia_ManStop( pAux );
            Gia_AigerWrite( pReduce, "gsrm.aig", 0, 0, 0 );
//            Abc_Print( 1, "Speculatively reduced model was written into file \"%s\".\n", "gsrm.aig" );
//          Gia_ManPrintStatsShort( pReduce );
            Gia_ManStop( pReduce );
        }
    }
    return 1;
}




/**Function*************************************************************

  Synopsis    [Reduces AIG using equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFilterEquivsForSpeculation(pGia *Gia_Man_t, pName1 string, pName2 string, fLatchA int32, fLatchB int32) int32 {
    var pGia1, pGia2, pMiter *Gia_Man_t;
    var pObj1, pObj2, pObjM, pObj *Gia_Obj_t;
    var i, iObj, iNext, Counter int32;
    if pGia.pReprs == nil || pGia.pNexts == nil {
        Abc_Print( 1, "Equivalences are not defined.\n" );
        return 0;
    }
    pGia1 = Gia_AigerRead( pName1, 0, 0, 0 );
    if pGia1 == nil {
        Abc_Print( 1, "Cannot read first file %s.\n", pName1 );
        return 0;
    }
    pGia2 = Gia_AigerRead( pName2, 0, 0, 0 );
    if pGia2 == nil {
        Gia_ManStop( pGia2 );
        Abc_Print( 1, "Cannot read second file %s.\n", pName2 );
        return 0;
    }
    pMiter = Gia_ManMiter( pGia1, pGia2, 0, 0, 1, 0, 0 );
    if pMiter == nil {
        Gia_ManStop( pGia1 );
        Gia_ManStop( pGia2 );
        Abc_Print( 1, "Cannot create sequential miter.\n" );
        return 0;
    }
    // make sure the miter is isomorphic
    if Gia_ManObjNum(pGia) != Gia_ManObjNum(pMiter) {
        Gia_ManStop( pGia1 );
        Gia_ManStop( pGia2 );
        Gia_ManStop( pMiter );
        Abc_Print( 1, "The number of objects in different.\n" );
        return 0;
    }
    if memcmp( pGia.pObjs, pMiter.pObjs, sizeof(Gia_Obj_t) *  Gia_ManObjNum(pGia) ) {
        Gia_ManStop( pGia1 );
        Gia_ManStop( pGia2 );
        Gia_ManStop( pMiter );
        Abc_Print( 1, "The AIG structure of the miter does not match.\n" );
        return 0;
    }
    // transfer copies
    Gia_ManCleanMark0( pGia );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia1.nObjs) && ((pObj1) = Gia_ManObj(pGia1, i)); i++  {
        if pObj1.Value == ~0 {
            continue; }
        pObjM = Gia_ManObj( pMiter, Abc_Lit2Var(pObj1.Value) );
        pObj = Gia_ManObj( pGia, Gia_ObjId(pMiter, pObjM) );
        pObj.fMark0 = 1;
    }
    Gia_ManCleanMark1( pGia );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia2.nObjs) && ((pObj2) = Gia_ManObj(pGia2, i)); i++  {
        if pObj2.Value == ~0 {
            continue; }
        pObjM = Gia_ManObj( pMiter, Abc_Lit2Var(pObj2.Value) );
        pObj = Gia_ManObj( pGia, Gia_ObjId(pMiter, pObjM) );
        pObj.fMark1 = 1;
    }

    // filter equivalences
    // inlining of macro Gia_ManForEachConst
    for  i = 1; i < Gia_ManObjNum(pGia); i++  { if !Gia_ObjIsConst(pGia, i) {} else {
        Gia_ObjUnsetRepr( pGia, i );
        assert(pGia.pNexts[i] == 0, "pGia.pNexts[i] == 0");
    }
    }
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(pGia); i++  { if !Gia_ObjIsHead(pGia, i) {} else {
        // find the first colorA and colorB
        var ClassA int32 = -1;
        var ClassB int32 = -1;
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(pGia, i) && (i != 0), "Gia_ObjIsHead(pGia, i) && i"), iObj = i; iObj > 0; iObj = Gia_ObjNext(pGia, iObj)  {
            pObj = Gia_ManObj( pGia, iObj );
            if ClassA == -1 && pObj.fMark0 && !pObj.fMark1 {
                if (fLatchA != 0) && !Gia_ObjIsRo(pGia, pObj) {
                    continue; }
                ClassA = iObj;
            }
            if ClassB == -1 && pObj.fMark1 && !pObj.fMark0 {
                if (fLatchB != 0) && !Gia_ObjIsRo(pGia, pObj) {
                    continue; }
                ClassB = iObj;
            }
        }
        // undo equivalence classes
        for  iObj = i, iNext = Gia_ObjNext(pGia, iObj); (iObj != 0);
              iObj = iNext, iNext = Gia_ObjNext(pGia, iObj)  {
            Gia_ObjUnsetRepr( pGia, iObj );
            Gia_ObjSetNext( pGia, iObj, 0 );
        }
        assert(!Gia_ObjIsHead(pGia, i), "!Gia_ObjIsHead(pGia, i)");
        if ClassA > 0 && ClassB > 0 {
            if ClassA > ClassB {
                ClassA ^= ClassB;
                ClassB ^= ClassA;
                ClassA ^= ClassB;
            }
            assert(ClassA < ClassB, "ClassA < ClassB");
            Gia_ObjSetNext( pGia, ClassA, ClassB );
            Gia_ObjSetRepr( pGia, ClassB, ClassA );
            Counter++;
            assert(Gia_ObjIsHead(pGia, ClassA), "Gia_ObjIsHead(pGia, ClassA)");
        }
    }
    }
    Abc_Print( 1, "The number of two-node classes after filtering = %d.\n", Counter );
//Gia_ManEquivPrintClasses( pGia, 1, 0 );

    Gia_ManCleanMark0( pGia );
    Gia_ManCleanMark1( pGia );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Reduces AIG using equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFilterEquivsUsingParts(pGia *Gia_Man_t, pName1 string, pName2 string) int32 {
    var vNodes *Vec_Int_t;
    var pGia1, pGia2, pMiter *Gia_Man_t;
    var pObj1 *Gia_Obj_t;
    var pObj2 *Gia_Obj_t;
    var pObjM *Gia_Obj_t;
    var pObj *Gia_Obj_t = nil;
    var i, k, iObj, iNext, iPrev, iRepr int32;
    var iLitsOld, iLitsNew int32;
    if pGia.pReprs == nil || pGia.pNexts == nil {
        Abc_Print( 1, "Equivalences are not defined.\n" );
        return 0;
    }
    pGia1 = Gia_AigerRead( pName1, 0, 0, 0 );
    if pGia1 == nil {
        Abc_Print( 1, "Cannot read first file %s.\n", pName1 );
        return 0;
    }
    pGia2 = Gia_AigerRead( pName2, 0, 0, 0 );
    if pGia2 == nil {
        Gia_ManStop( pGia2 );
        Abc_Print( 1, "Cannot read second file %s.\n", pName2 );
        return 0;
    }
    pMiter = Gia_ManMiter( pGia1, pGia2, 0, 0, 1, 0, 0 );
    if pMiter == nil {
        Gia_ManStop( pGia1 );
        Gia_ManStop( pGia2 );
        Abc_Print( 1, "Cannot create sequential miter.\n" );
        return 0;
    }
    // make sure the miter is isomorphic
    if Gia_ManObjNum(pGia) != Gia_ManObjNum(pMiter) {
        Gia_ManStop( pGia1 );
        Gia_ManStop( pGia2 );
        Gia_ManStop( pMiter );
        Abc_Print( 1, "The number of objects in different.\n" );
        return 0;
    }
    if memcmp( pGia.pObjs, pMiter.pObjs, sizeof(Gia_Obj_t) *  Gia_ManObjNum(pGia) ) {
        Gia_ManStop( pGia1 );
        Gia_ManStop( pGia2 );
        Gia_ManStop( pMiter );
        Abc_Print( 1, "The AIG structure of the miter does not match.\n" );
        return 0;
    }
    // transfer copies
    Gia_ManCleanMark0( pGia );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia1.nObjs) && ((pObj1) = Gia_ManObj(pGia1, i)); i++  {
        if pObj1.Value == ~0 {
            continue; }
        pObjM = Gia_ManObj( pMiter, Abc_Lit2Var(pObj1.Value) );
        pObj = Gia_ManObj( pGia, Gia_ObjId(pMiter, pObjM) );
        pObj.fMark0 = 1;
    }
    Gia_ManCleanMark1( pGia );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia2.nObjs) && ((pObj2) = Gia_ManObj(pGia2, i)); i++  {
        if pObj2.Value == ~0 {
            continue; }
        pObjM = Gia_ManObj( pMiter, Abc_Lit2Var(pObj2.Value) );
        pObj = Gia_ManObj( pGia, Gia_ObjId(pMiter, pObjM) );
        pObj.fMark1 = 1;
    }

    // filter equivalences
    iLitsOld = iLitsNew = 0;
    // inlining of macro Gia_ManForEachConst
    for  i = 1; i < Gia_ManObjNum(pGia); i++  { if !Gia_ObjIsConst(pGia, i) {} else {
        iLitsOld++;
        pObj = Gia_ManObj( pGia, i );
        assert(pGia.pNexts[i] == 0, "pGia.pNexts[i] == 0");
        assert(pObj.fMark0 || pObj.fMark1, "pObj.fMark0 || pObj.fMark1");
        if pObj.fMark0 && pObj.fMark1 { // belongs to both A and B
            Gia_ObjUnsetRepr( pGia, i );
        } else {
            iLitsNew++; }
    }
    }
    // filter equivalences
    vNodes = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(pGia); i++  { if !Gia_ObjIsHead(pGia, i) {} else {
        var fSeenA, fSeenB int32;
        assert(pObj.fMark0 || pObj.fMark1, "pObj.fMark0 || pObj.fMark1");
        Vec_IntClear( vNodes );
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(pGia, i) && (i != 0), "Gia_ObjIsHead(pGia, i) && i"), iObj = i; iObj > 0; iObj = Gia_ObjNext(pGia, iObj)  {
            pObj = Gia_ManObj( pGia, iObj );
            if pObj.fMark0 && !pObj.fMark1 {
                fSeenA = 1;
                Vec_IntPush( vNodes, iObj );
            }
            if !pObj.fMark0 && pObj.fMark1 {
                fSeenB = 1;
                Vec_IntPush( vNodes, iObj );
            }
            iLitsOld++;
        }
        iLitsOld--;
        // undo equivalence classes
        for  iObj = i, iNext = Gia_ObjNext(pGia, iObj); (iObj != 0);
              iObj = iNext, iNext = Gia_ObjNext(pGia, iObj)  {
            Gia_ObjUnsetRepr( pGia, iObj );
            Gia_ObjSetNext( pGia, iObj, 0 );
        }
        assert(!Gia_ObjIsHead(pGia, i), "!Gia_ObjIsHead(pGia, i)");
        if fSeenA && (fSeenB != 0) && Vec_IntSize(vNodes) > 1 {
            // create new class
            iPrev = iRepr = Vec_IntEntry( vNodes, 0 );
            // inlining of macro Vec_IntForEachEntryStart
            for  k = 1; (k < Vec_IntSize(vNodes)) && (((iObj) = Vec_IntEntry(vNodes, k)), 1); k++  {
                Gia_ObjSetRepr( pGia, iObj, iRepr );
                Gia_ObjSetNext( pGia, iPrev, iObj );
                iPrev = iObj;
                iLitsNew++;
            }
            assert(Gia_ObjNext(pGia, iPrev) == 0, "Gia_ObjNext(pGia, iPrev) == 0");
        }
    }
    }
    Vec_IntFree( vNodes );
    Abc_Print( 1, "The number of literals: Before = %d. After = %d.\n", iLitsOld, iLitsNew );
//Gia_ManEquivPrintClasses( pGia, 1, 0 );

    Gia_ManCleanMark0( pGia );
    Gia_ManCleanMark1( pGia );
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFilterEquivsUsingLatches(pGia *Gia_Man_t, fFlopsOnly int32, fFlopsWith int32, fUseRiDrivers int32) {
    var pObjR *Gia_Obj_t;
    var vNodes, vFfIds *Vec_Int_t;
    var i, k, iObj, iNext, iPrev, iRepr int32;
    var iLitsOld, iLitsNew int32;
    assert(fFlopsOnly ^ fFlopsWith, "fFlopsOnly ^ fFlopsWith");
    vNodes = Vec_IntAlloc( 100 );
    // select nodes "flop" node IDs
    vFfIds = Vec_IntStart( Gia_ManObjNum(pGia) );
    if fUseRiDrivers != 0 {
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(pGia)) && ((pObjR) = Gia_ManCo(pGia, Gia_ManPoNum(pGia)+i)); i++  {
            Vec_IntWriteEntry( vFfIds, Gia_ObjFaninId0p(pGia, pObjR), 1 );
        }
    
    } else {
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(pGia)) && ((pObjR) = Gia_ManCi(pGia, Gia_ManPiNum(pGia)+i)); i++  {
            Vec_IntWriteEntry( vFfIds, Gia_ObjId(pGia, pObjR), 1 );
        }
    }
    // remove all non-flop constants
    // inlining of macro Gia_ManForEachConst
    for  i = 1; i < Gia_ManObjNum(pGia); i++  { if !Gia_ObjIsConst(pGia, i) {} else {
        iLitsOld++;
        assert(pGia.pNexts[i] == 0, "pGia.pNexts[i] == 0");
        if !Vec_IntEntry(vFfIds, i) {
            Gia_ObjUnsetRepr( pGia, i );
        } else {
            iLitsNew++; }
    }
    }
    // clear the classes
    if fFlopsOnly != 0 {
        // inlining of macro Gia_ManForEachClass
        for  i = 1; i < Gia_ManObjNum(pGia); i++  { if !Gia_ObjIsHead(pGia, i) {} else {
            Vec_IntClear( vNodes );
            // inlining of macro Gia_ClassForEachObj
            for  assert(Gia_ObjIsHead(pGia, i) && (i != 0), "Gia_ObjIsHead(pGia, i) && i"), iObj = i; iObj > 0; iObj = Gia_ObjNext(pGia, iObj)  {
                if Vec_IntEntry(vFfIds, iObj) {
                    Vec_IntPush( vNodes, iObj ); }
                iLitsOld++;
            }
            iLitsOld--;
            // undo equivalence classes
            for  iObj = i, iNext = Gia_ObjNext(pGia, iObj); (iObj != 0);
                  iObj = iNext, iNext = Gia_ObjNext(pGia, iObj)  {
                Gia_ObjUnsetRepr( pGia, iObj );
                Gia_ObjSetNext( pGia, iObj, 0 );
            }
            assert(!Gia_ObjIsHead(pGia, i), "!Gia_ObjIsHead(pGia, i)");
            if Vec_IntSize(vNodes) > 1 {
                // create new class
                iPrev = iRepr = Vec_IntEntry( vNodes, 0 );
                // inlining of macro Vec_IntForEachEntryStart
                for  k = 1; (k < Vec_IntSize(vNodes)) && (((iObj) = Vec_IntEntry(vNodes, k)), 1); k++  {
                    Gia_ObjSetRepr( pGia, iObj, iRepr );
                    Gia_ObjSetNext( pGia, iPrev, iObj );
                    iPrev = iObj;
                    iLitsNew++;
                }
                assert(Gia_ObjNext(pGia, iPrev) == 0, "Gia_ObjNext(pGia, iPrev) == 0");
            }
        }
        }
    
    } else {
        // inlining of macro Gia_ManForEachClass
        for  i = 1; i < Gia_ManObjNum(pGia); i++  { if !Gia_ObjIsHead(pGia, i) {} else {
            var fSeenFlop int32;
            // inlining of macro Gia_ClassForEachObj
            for  assert(Gia_ObjIsHead(pGia, i) && (i != 0), "Gia_ObjIsHead(pGia, i) && i"), iObj = i; iObj > 0; iObj = Gia_ObjNext(pGia, iObj)  {
                if Vec_IntEntry(vFfIds, iObj) {
                    fSeenFlop = 1; }
                iLitsOld++;
                iLitsNew++;
            }
            iLitsOld--;
            iLitsNew--;
            if fSeenFlop != 0 {
                continue; }
            // undo equivalence classes
            for  iObj = i, iNext = Gia_ObjNext(pGia, iObj); (iObj != 0);
                  iObj = iNext, iNext = Gia_ObjNext(pGia, iObj)  {
                Gia_ObjUnsetRepr( pGia, iObj );
                Gia_ObjSetNext( pGia, iObj, 0 );
                iLitsNew--;
            }
            iLitsNew++;
            assert(!Gia_ObjIsHead(pGia, i), "!Gia_ObjIsHead(pGia, i)");
        }
        }
    }
    Vec_IntFree( vNodes );
    Vec_IntFree( vFfIds );
    Abc_Print( 1, "The number of literals: Before = %d. After = %d.\n", iLitsOld, iLitsNew );
}

/**Function*************************************************************

  Synopsis    [Changing node order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManChangeOrder_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if ~pObj.Value {
        return pObj.Value; }
    if Gia_ObjIsCi(pObj) {
        return pObj.Value = Gia_ManAppendCi(pNew); }
    Gia_ManChangeOrder_rec( pNew, p, Gia_ObjFanin0(pObj) );
    if Gia_ObjIsCo(pObj) {
        return pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    Gia_ManChangeOrder_rec( pNew, p, Gia_ObjFanin1(pObj) );
    return pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
} 
func Gia_ManChangeOrder(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, k int32;
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, i) && (i != 0), "Gia_ObjIsHead(p, i) && i"), k = i; k > 0; k = Gia_ObjNext(p, k)  {
            Gia_ManChangeOrder_rec( pNew, p, Gia_ManObj(p, k) );
        } }
    }
    // inlining of macro Gia_ManForEachConst
    for  k = 1; k < Gia_ManObjNum(p); k++  { if !Gia_ObjIsConst(p, k) {} else {
        Gia_ManChangeOrder_rec( pNew, p, Gia_ManObj(p, k) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManChangeOrder_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    assert(Gia_ManObjNum(pNew) == Gia_ManObjNum(p), "Gia_ManObjNum(pNew) == Gia_ManObjNum(p)");
    return pNew;
}
func Gia_ManTransferEquivs(p *Gia_Man_t, pNew *Gia_Man_t) {
    var vClass *Vec_Int_t;
    var i, k, iNode, iRepr int32;
    assert(Gia_ManObjNum(p) == Gia_ManObjNum(pNew), "Gia_ManObjNum(p) == Gia_ManObjNum(pNew)");
    assert(p.pReprs != nil, "p.pReprs != nil");
    assert(p.pNexts != nil, "p.pNexts != nil");
    assert(pNew.pReprs == nil, "pNew.pReprs == nil");
    assert(pNew.pNexts == nil, "pNew.pNexts == nil");
    // start representatives
    // inlining of macro ABC_CALLOC
    pNew.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(pNew)), sizeof(Gia_Rpr_t)));
    for  i = 0; i < Gia_ManObjNum(pNew); i++  {
        Gia_ObjSetRepr( pNew, i, GIA_VOID );
    }
    // iterate over constant candidates
    // inlining of macro Gia_ManForEachConst
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsConst(p, i) {} else {
        Gia_ObjSetRepr( pNew, Abc_Lit2Var(Gia_ManObj(p, i).Value), 0 ); }
    }
    // iterate over class candidates
    vClass = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        Vec_IntClear( vClass );
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, i) && (i != 0), "Gia_ObjIsHead(p, i) && i"), k = i; k > 0; k = Gia_ObjNext(p, k)  {
            Vec_IntPushUnique( vClass, Abc_Lit2Var(Gia_ManObj(p, k).Value) );
        }
        assert(Vec_IntSize( vClass ) > 1, "Vec_IntSize( vClass ) > 1");
        Vec_IntSort( vClass, 0 );
        iRepr = Vec_IntEntry( vClass, 0 );
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vClass)) && (((iNode) = Vec_IntEntry(vClass, k)), 1); k++  {
            Gia_ObjSetRepr( pNew, iNode, iRepr );
        }
    }
    }
    Vec_IntFree( vClass );
    pNew.pNexts = Gia_ManDeriveNexts( pNew );
}
func Gia_ManTransferTest(p *Gia_Man_t) {
    var pObj *Gia_Obj_t; var i int32;
    var pReprs *Gia_Rpr_t  = p.pReprs;  // representatives (for CIs and ANDs)
    var pNexts *int32  = p.pNexts;  // next nodes in the equivalence classes
    var pNew *Gia_Man_t  = Gia_ManChangeOrder(p);
    //Gia_ManEquivPrintClasses( p, 1, 0 );
    assert(Gia_ManObjNum(p) == Gia_ManObjNum(pNew), "Gia_ManObjNum(p) == Gia_ManObjNum(pNew)");
    Gia_ManTransferEquivs( p, pNew );
    p.pReprs = nil;
    p.pNexts = nil;
    // make new point to old
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        assert(!Abc_LitIsCompl(pObj.Value), "!Abc_LitIsCompl(pObj.Value)");
        Gia_ManObj(pNew, Abc_Lit2Var(pObj.Value)).Value = Abc_Var2Lit(i, 0);
    }
    Gia_ManTransferEquivs( pNew, p );
    //Gia_ManEquivPrintClasses( p, 1, 0 );
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        pReprs[i].fProved = 0;
    }
        //printf( "%5d :    %5d %5d    %5d %5d\n", i, *(int*)&p->pReprs[i], *(int*)&pReprs[i], (int)p->pNexts[i], (int)pNexts[i] );
    if memcmp(p.pReprs, pReprs, sizeof(int)*Gia_ManObjNum(p)) {
        fmt.Printf("Verification of reprs failed.\n");
    } else {
        fmt.Printf("Verification of reprs succeeded.\n"); }
    if memcmp(p.pNexts, pNexts, sizeof(int)*Gia_ManObjNum(p)) {
        fmt.Printf("Verification of nexts failed.\n");
    } else {
        fmt.Printf("Verification of nexts succeeded.\n"); }
    // inlining of macro ABC_FREE
    pNew.pReprs = nil;
    // inlining of macro ABC_FREE
    pNew.pNexts = nil;
    // inlining of macro ABC_FREE
    pReprs = nil;
    // inlining of macro ABC_FREE
    pNexts = nil;
    Gia_ManStop( pNew );
}


/**Function*************************************************************

  Synopsis    [Converting AIG after SAT sweeping into AIG with choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec4_ManMarkIndependentClasses_rec(p *Gia_Man_t, iObj int32) int32 {
    var pObj *Gia_Obj_t;
    assert(iObj > 0, "iObj > 0");
    if Gia_ObjIsTravIdPreviousId(p, iObj) { // failed
        return 0; }
    if Gia_ObjIsTravIdCurrentId(p, iObj) { // passed
        return 1; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        return 1; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if (Cec4_ManMarkIndependentClasses_rec( p, Gia_ObjFaninId0(pObj, iObj) ) != 0) && 
         (Cec4_ManMarkIndependentClasses_rec( p, Gia_ObjFaninId1(pObj, iObj) ) != 0) {
        return 1; }
    Gia_ObjSetTravIdPreviousId(p, iObj);
    return 0;
}
func Cec4_ManMarkIndependentClasses(p *Gia_Man_t, pNew *Gia_Man_t) int32 {
    var iObjNew, iRepr, iObj, Res, fHaveChoices int32;
    Gia_ManCleanMark01(p);
    // inlining of macro Gia_ManForEachClass
    for  iRepr = 1; iRepr < Gia_ManObjNum(p); iRepr++  { if !Gia_ObjIsHead(p, iRepr) {} else {
        Gia_ManIncrementTravId( pNew );
        Gia_ManIncrementTravId( pNew );
        iObjNew = Abc_Lit2Var( Gia_ManObj(p, iRepr).Value );
        Res = Cec4_ManMarkIndependentClasses_rec( pNew, iObjNew );
        assert(Res == 1, "Res == 1");
        Gia_ObjSetTravIdPreviousId( pNew, iObjNew );
        p.pReprs[iRepr].fColorA = 1;
        // inlining of macro Gia_ClassForEachObj1
        for  assert(Gia_ObjIsHead(p, iRepr), "Gia_ObjIsHead(p, iRepr)"), iObj = Gia_ObjNext(p, iRepr); iObj > 0; iObj = Gia_ObjNext(p, iObj)  {
            assert(p.pReprs[iObj].iRepr == uint32(iRepr), "p.pReprs[iObj].iRepr == uint32(iRepr)");
            iObjNew = Abc_Lit2Var( Gia_ManObj(p, iObj).Value );
            if Cec4_ManMarkIndependentClasses_rec( pNew, iObjNew ) != 0 {
                p.pReprs[iObj].fColorA = 1;
                fHaveChoices = 1;
            }
            Gia_ObjSetTravIdPreviousId( pNew, iObjNew );
        }
    }
    }
    return fHaveChoices;
}
func Cec4_ManSatSolverAnd_rec(pCho *Gia_Man_t, p *Gia_Man_t, pNew *Gia_Man_t, iObj int32) int32 {
    return 0;
}
func Cec4_ManSatSolverChoices_rec(pCho *Gia_Man_t, p *Gia_Man_t, pNew *Gia_Man_t, iObj int32) int32 {
    if !Gia_ObjIsClass(p, iObj) {
        return Cec4_ManSatSolverAnd_rec( pCho, p, pNew, iObj );
    } else {
        var vLits *Vec_Int_t  = Vec_IntAlloc( 100 );
        var i int32;
        var iHead int32;
        var iNext int32;
        var iRepr int32;
        if Gia_ObjIsHead(p, iObj) {
            iRepr = iObj;
        } else {
            iRepr = Gia_ObjRepr(p, iObj);
        }
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, iRepr) && (iRepr != 0), "Gia_ObjIsHead(p, iRepr) && iRepr"), iObj = iRepr; iObj > 0; iObj = Gia_ObjNext(p, iObj)  {
            if p.pReprs[iObj].fColorA {
                Vec_IntPush( vLits, Cec4_ManSatSolverAnd_rec( pCho, p, pNew, iObj ) ); }
        }
        Vec_IntSort( vLits, 1 );
        iHead = Abc_Lit2Var( Vec_IntEntry(vLits, 0) );
        if Vec_IntSize(vLits) > 1 {
            // inlining of macro Vec_IntForEachEntryStart
            for  i = 1; (i < Vec_IntSize(vLits)) && (((iNext) = Vec_IntEntry(vLits, i)), 1); i++  {
                pCho.pSibls[iHead] = Abc_Lit2Var(iNext);  
                iHead = Abc_Lit2Var(iNext);
            }
        }
        return Abc_LitNotCond( Vec_IntEntry(vLits, 0), Gia_ManObj(p, iHead).fPhase );
    }
}
func Cec4_ManSatSolverChoices(p *Gia_Man_t, pNew *Gia_Man_t) *Gia_Man_t {
    var pCho *Gia_Man_t;
    var pObj *Gia_Obj_t; 
    var i, DriverId int32;
    // mark topologically dependent equivalent nodes
    if !(Cec4_ManMarkIndependentClasses( p, pNew ) != 0) {
        return Gia_ManDup( pNew ); }
    // rebuild AIG in a different order with choices
    pCho = Gia_ManStart( Gia_ManObjNum(pNew) );
    pCho.pName = Abc_UtilStrsav( p.pName );
    pCho.pSpec = Abc_UtilStrsav( p.pSpec );
    // inlining of macro ABC_CALLOC
    pCho.pSibls = ((int *) calloc((size_t)(Gia_ManObjNum(pNew)), sizeof(int)));
    Gia_ManFillValue(pNew);
    Gia_ManConst0(pNew).Value = 0;
    for  i = 0; i < Gia_ManCiNum(pNew); i++  {
        Gia_ManCi(pNew, i).Value = Gia_ManAppendCi( pCho );
    }
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && (((DriverId) = Gia_ObjFaninId0p(p, Gia_ManCo(p, i))), 1); i++  {
        Cec4_ManSatSolverChoices_rec( pCho, p, pNew, DriverId );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pNew.vCos)) && ((pObj) = Gia_ManCo(pNew, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pCho, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pCho, Gia_ManRegNum(p) );
    return pCho;
}

/**Function*************************************************************

  Synopsis    [Converting AIG after SAT sweeping into AIG with choices.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCombSpecReduce(p *Gia_Man_t) *Gia_Man_t {
    var pObj, pRepr *Gia_Obj_t; var i, iLit int32;
    var vXors *Vec_Int_t  = Vec_IntAlloc( 100 );
    var pTemp *Gia_Man_t;
    var pNew *Gia_Man_t = Gia_ManStart( Gia_ManObjNum(p) );
    assert(p.pReprs && p.pNexts, "p.pReprs && p.pNexts");
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManLevelNum(p);
    Gia_ManSetPhase(p);
    Gia_ManFillValue(p);
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        pRepr = Gia_ObjReprObj( p, i );
        if (pRepr != nil) && Abc_Lit2Var(pObj.Value) != Abc_Lit2Var(pRepr.Value) {
            //if ( Gia_ObjLevel(p, pRepr) > Gia_ObjLevel(p, pObj) + 50 )
            //printf( "%d %d  ", Gia_ObjLevel(p, pRepr), Gia_ObjLevel(p, pObj) );
            iLit = Abc_LitNotCond( pRepr.Value, pObj.fPhase ^ pRepr.fPhase );
            Vec_IntPush( vXors, Gia_ManHashXor( pNew, pObj.Value, iLit ) );
            pObj.Value = iLit;
        }
    }
    }
    Gia_ManHashStop( pNew );
    if Vec_IntSize(vXors) == 0 {
        Vec_IntPush( vXors, 0 ); }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXors)) && (((iLit) = Vec_IntEntry(vXors, i)), 1); i++  {
        Gia_ManAppendCo( pNew, iLit );
    }
    Vec_IntFree( vXors );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Gia_ManCombSpecReduceTest(p *Gia_Man_t, pFileName string) {
    var pSrm *Gia_Man_t  = Gia_ManCombSpecReduce( p );
    if pFileName == nil {
        pFileName = "test.aig"; }
    Gia_AigerWrite( pSrm, pFileName, 0, 0, 0 );
    Abc_Print( 1, "Speculatively reduced model was written into file \"%s\".\n", pFileName );
    Gia_ManStop( pSrm );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

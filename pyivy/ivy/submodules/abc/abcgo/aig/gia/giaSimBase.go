package gia

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [giaSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Fast sequential simulator.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"
//#include <immintrin.h>

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


   
type Gia_SimRsbMan_t struct {
    pGia     *Gia_Man_t;
    vTfo     *Vec_Int_t;
    vCands     *Vec_Int_t;
    vFanins     *Vec_Int_t;
    vFanins2     *Vec_Int_t;
    vSimsObj     *Vec_Wrd_t;
    vSimsObj2     *Vec_Wrd_t;
    nWords            int32;
    word *         pFunc[3];
}


   
type Gia_SimAbsMan_t struct {
    // problem formulation
    pGia     *Gia_Man_t;       // AIG manager
    word *         pSet[2];    // offset/onset truth tables
    nCands            int32;     // candidate count
    nWords            int32;     // word count
    vSims     *Vec_Wrd_t;      // candidate simulation info
    vResub     *Vec_Int_t;     // the result
    fVerbose            int32;   // verbose
    // intermediate result
    vValues     *Vec_Int_t;       // function values in each pattern
    vPatPairs     *Vec_Int_t;     // used minterms
    nWordsTable            int32;   // words of table data
    pTableTemp          *word;    // temporary table pattern
    vCoverTable     *Vec_Wrd_t;   // columns = minterms; rows = classes
    vTtMints     *Vec_Int_t;      // truth table minterms
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

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





/**Function*************************************************************

  Synopsis    []

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

func Gia_ManSimBitPackOne(nWords int32, vSimsIn *Vec_Wrd_t, vSimsCare *Vec_Wrd_t, iPat int32, pLits *int32, nLits int32) int32 {
    var pSimsI, pSimsC *word; var i, k int32;
    for  i = 0; i < iPat; i++  {
        for  k = 0; k < nLits; k++  {
            var iVar int32  = Abc_Lit2Var( pLits[k] );
            pSimsI = Vec_WrdEntryP( vSimsIn,   nWords * iVar );
            pSimsC = Vec_WrdEntryP( vSimsCare, nWords * iVar );
            if Abc_TtGetBit(pSimsC, i) && (Abc_TtGetBit(pSimsI, i) == Abc_LitIsCompl(pLits[k])) {
                break; }
        }
        if k == nLits {
            break; }
    }
    for  k = 0; k < nLits; k++  {
        var iVar int32  = Abc_Lit2Var( pLits[k] );
        pSimsI = Vec_WrdEntryP( vSimsIn,   nWords * iVar );
        pSimsC = Vec_WrdEntryP( vSimsCare, nWords * iVar );
        if !Abc_TtGetBit(pSimsC, i) && Abc_TtGetBit(pSimsI, i) == Abc_LitIsCompl(pLits[k]) {
            Abc_TtXorBit( pSimsI, i ); }
        Abc_TtSetBit( pSimsC, i );
        assert(Abc_TtGetBit(pSimsC, i) && (Abc_TtGetBit(pSimsI, i) != Abc_LitIsCompl(pLits[k])), "Abc_TtGetBit(pSimsC, i) && (Abc_TtGetBit(pSimsI, i) != Abc_LitIsCompl(pLits[k]))");
    }
    return int32((i == iPat));
}
func Gia_ManSimBitPacking(p *Gia_Man_t, vCexStore *Vec_Int_t, nCexes int32, nUnDecs int32) *Vec_Wrd_t {
    var c, iCur, iPat int32;
    var nWordsMax int32  = Abc_Bit6WordNum( nCexes ); 
    var vSimsIn *Vec_Wrd_t    = Vec_WrdStartRandom( Gia_ManCiNum(p) * nWordsMax );
    var vSimsCare *Vec_Wrd_t  = Vec_WrdStart( Gia_ManCiNum(p) * nWordsMax );
    var vSimsRes *Vec_Wrd_t   = nil;
    for  c = 0; c < nCexes + nUnDecs; c++  {
        var Out int32   = Vec_IntEntry( vCexStore, iCur++ );
        var Size int32  = Vec_IntEntry( vCexStore, iCur++ );
        if Size == -1 {
            continue; }
        iPat += Gia_ManSimBitPackOne( nWordsMax, vSimsIn, vSimsCare, iPat, Vec_IntEntryP(vCexStore, iCur), Size );
        iCur += Size;
        assert(iPat <= nCexes + nUnDecs, "iPat <= nCexes + nUnDecs");
        Out = 0;
    }
    assert(iCur == Vec_IntSize(vCexStore), "iCur == Vec_IntSize(vCexStore)");
    vSimsRes = Gia_ManSimCombine( Gia_ManCiNum(p), p.vSimsPi, vSimsIn, Abc_Bit6WordNum(iPat+1) );
    fmt.Printf("Compressed %d CEXes into %d patterns and added %d words to available %d words.\n", nCexes, iPat, Abc_Bit6WordNum(iPat+1), Vec_WrdSize(p.vSimsPi) / Gia_ManCiNum(p));
    Vec_WrdFree( vSimsIn );
    Vec_WrdFree( vSimsCare );
    return vSimsRes;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimPatHashPatterns(p *Gia_Man_t, nWords int32, vSims *Vec_Wrd_t, pnC0 *int32, pnC1 *int32) int32 {
    var pObj *Gia_Obj_t; 
    var i, nUnique int32;
    var vStore *Vec_Mem_t;
    vStore = Vec_MemAlloc( nWords, 12 ); // 2^12 N-word entries per page
    Vec_MemHashAlloc( vStore, 1 << 12 );
    // inlining of macro Gia_ManForEachCand
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsCand(pObj) {} else {
        var pSim *word  = Vec_WrdEntryP(vSims, i*nWords);
        if (pnC0 != nil) && Abc_TtIsConst0(pSim, nWords) {
            (*pnC0)++; }
        if (pnC1 != nil) && Abc_TtIsConst1(pSim, nWords) {
            (*pnC1)++; }
        Vec_MemHashInsert( vStore, pSim );
    }
    }
    nUnique = Vec_MemEntryNum( vStore );
    Vec_MemHashFree( vStore );
    Vec_MemFree( vStore );
    return nUnique;
}
func Gia_ManSimPatGenMiter(p *Gia_Man_t, vSims *Vec_Wrd_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var nWords int32 = Vec_WrdSize(vSims) / Gia_ManObjNum(p);
    pNew = Gia_ManStart( Gia_ManObjNum(p) + Gia_ManCoNum(p) );
    Gia_ManHashStart( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var pSim *word  = Vec_WrdEntryP(vSims, i*nWords);
        if Abc_TtIsConst0(pSim, nWords) {
            Gia_ManAppendCo( pNew, Abc_LitNotCond(pObj.Value, 0) ); }
        if Abc_TtIsConst1(pSim, nWords) {
            Gia_ManAppendCo( pNew, Abc_LitNotCond(pObj.Value, 1) ); }
    }
    }
    Gia_ManHashStop( pNew );
    return pNew;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimProfile(pGia *Gia_Man_t) {
    var vSims *Vec_Wrd_t  = Gia_ManSimPatSim( pGia );
    var nWords int32  = Vec_WrdSize(vSims) / Gia_ManObjNum(pGia);
    var nC0s int32;
    var nC1s int32;
    var nUnique int32 = Gia_ManSimPatHashPatterns( pGia, nWords, vSims, &nC0s, &nC1s );
    fmt.Printf("Simulating %d patterns leads to %d unique objects (%.2f %% out of %d), Const0 = %d. Const1 = %d.\n", 64*nWords, nUnique, 100.0*nUnique/Gia_ManCandNum(pGia), Gia_ManCandNum(pGia), nC0s, nC1s);
    Vec_WrdFree( vSims );
}
func Gia_ManPatSatImprove(p *Gia_Man_t, nWords0 int32, fVerbose int32) {
    extern Vec_Int_t * Cbs2_ManSolveMiterNc( Gia_Man_t * pAig, int nConfs, Vec_Str_t ** pvStatus, int fVerbose );
    var i int32;
    var Status int32;
    var Counts [3]int32 ;
    var pGia *Gia_Man_t;
    var vSimsIn *Vec_Wrd_t  = nil;
    var vStatus *Vec_Str_t  = nil;
    var vCexStore *Vec_Int_t  = nil;
    var vSims *Vec_Wrd_t  = Gia_ManSimPatSim( p );
    //Gia_ManSimProfile( p );
    pGia  = Gia_ManSimPatGenMiter( p, vSims );
    vCexStore = Cbs2_ManSolveMiterNc( pGia, 1000, &vStatus, 0 );
    Gia_ManStop( pGia );
    // inlining of macro Vec_StrForEachEntry
    for  i = 0; (i < Vec_StrSize(vStatus)) && (((Status) = Vec_StrEntry(vStatus, i)), 1); i++  {
        assert(Status >= -1 && Status <= 1, "Status >= -1 && Status <= 1");
        Counts[Status+1]++;
    }
    if fVerbose != 0 {
        fmt.Printf("Total = %d : SAT = %d.  UNSAT = %d.  UNDEC = %d.\n", Counts[1]+Counts[2]+Counts[0], Counts[1], Counts[2], Counts[0]); }
    if Counts[1] == 0 {
        fmt.Printf("There are no counter-examples.  No need for more simulation.\n");
    } else {
        vSimsIn = Gia_ManSimBitPacking( p, vCexStore, Counts[1], Counts[0] );
        Vec_WrdFreeP( &p.vSimsPi );
        p.vSimsPi = vSimsIn;
        //Gia_ManSimProfile( p );
    }
    Vec_StrFree( vStatus );
    Vec_IntFree( vCexStore );
    Vec_WrdFree( vSims );
}




/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SimRsbAlloc(pGia *Gia_Man_t) *Gia_SimRsbMan_t {
    // inlining of macro ABC_CALLOC
    var p *Gia_SimRsbMan_t  = &Gia_SimRsbMan_t{};
    p.pGia      = pGia;
    p.nWords    = Vec_WrdSize(pGia.vSimsPi) / Gia_ManCiNum(pGia); assert(Vec_WrdSize(pGia.vSimsPi) % Gia_ManCiNum(pGia) == 0, "Vec_WrdSize(pGia.vSimsPi) % Gia_ManCiNum(pGia) == 0");
    // inlining of macro ABC_CALLOC
    p.pFunc[0]  = ((word *) calloc((size_t)(p.nWords), sizeof(word)));
    // inlining of macro ABC_CALLOC
    p.pFunc[1]  = ((word *) calloc((size_t)(p.nWords), sizeof(word)));
    // inlining of macro ABC_CALLOC
    p.pFunc[2]  = ((word *) calloc((size_t)(p.nWords), sizeof(word)));
    p.vTfo      = Vec_IntAlloc( 1000 );
    p.vCands    = Vec_IntAlloc( 1000 );
    p.vFanins   = Vec_IntAlloc( 10 );
    p.vFanins2  = Vec_IntAlloc( 10 );
    p.vSimsObj  = Gia_ManSimPatSim( pGia );
    p.vSimsObj2 = Vec_WrdStart( Vec_WrdSize(p.vSimsObj) );
    assert(p.nWords == Vec_WrdSize(p.vSimsObj) / Gia_ManObjNum(pGia), "p.nWords == Vec_WrdSize(p.vSimsObj) / Gia_ManObjNum(pGia)");
    Gia_ManStaticFanoutStart( pGia );
    return p;
}
func Gia_SimRsbFree(p *Gia_SimRsbMan_t) {
    Gia_ManStaticFanoutStop( p.pGia );
    Vec_IntFree( p.vTfo );
    Vec_IntFree( p.vCands );
    Vec_IntFree( p.vFanins );
    Vec_IntFree( p.vFanins2 );
    Vec_WrdFree( p.vSimsObj );
    Vec_WrdFree( p.vSimsObj2 );
    // inlining of macro ABC_FREE
    p.pFunc[0] = nil;
    // inlining of macro ABC_FREE
    p.pFunc[1] = nil;
    // inlining of macro ABC_FREE
    p.pFunc[2] = nil;
    // inlining of macro ABC_FREE
    p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SimRsbTfo_rec(p *Gia_Man_t, iObj int32, iFanout int32, vTfo *Vec_Int_t) {
    var i, iFan int32;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    // inlining of macro Gia_ObjForEachFanoutStaticId
    for  i = 0; (i < Gia_ObjFanoutNumId(p, iObj))   && (((iFan) = Gia_ObjFanoutId(p, iObj, i)), 1); i++  {
        if iFanout == -1 || iFan == iFanout {
            Gia_SimRsbTfo_rec( p, iFan, -1, vTfo ); }
    }
    Vec_IntPush( vTfo, iObj );
}
func Gia_SimRsbTfo(p *Gia_SimRsbMan_t, iObj int32, iFanout int32) *Vec_Int_t {
    assert(iObj > 0, "iObj > 0");
    Vec_IntClear( p.vTfo );
    Gia_ManIncrementTravId( p.pGia );
    Gia_SimRsbTfo_rec( p.pGia, iObj, iFanout, p.vTfo );
    assert(Vec_IntEntryLast(p.vTfo) == iObj, "Vec_IntEntryLast(p.vTfo) == iObj");
    Vec_IntPop( p.vTfo );
    Vec_IntReverseOrder( p.vTfo );
    Vec_IntSort( p.vTfo, 0 );
    return p.vTfo;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SimRsbFunc(p *Gia_SimRsbMan_t, iObj int32, vFanins *Vec_Int_t, fOnSet int32) *word {
    var nTruthWords int32  = Abc_Truth6WordNum( Vec_IntSize(vFanins) );
    // inlining of macro ABC_CALLOC
    var pTruth *word    = ((word *) calloc((size_t)(nTruthWords), sizeof(word))); 
    var pFunc *word     = Vec_WrdEntryP( p.vSimsObj, p.nWords*iObj );
    var pFanins [16]*word = [16]*word{nil};  var s, b, iMint, i, iFanin int32;
    assert(Vec_IntSize(vFanins) <= 16, "Vec_IntSize(vFanins) <= 16");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFanins)) && (((iFanin) = Vec_IntEntry(vFanins, i)), 1); i++  {
        pFanins[i] = Vec_WrdEntryP( p.vSimsObj, p.nWords*iFanin );
    }
    for  s = 0; s < 64*p.nWords; s++  {
        if !Abc_TtGetBit(p.pFunc[2], s) || Abc_TtGetBit(pFunc, s) != fOnSet {
            continue; }
        iMint = 0;
        for  b = 0; b < Vec_IntSize(vFanins); b++  {
            if Abc_TtGetBit(pFanins[b], s) {
                iMint |= 1 << b; }
        }
        Abc_TtSetBit( pTruth, iMint );
    }
    return pTruth;
}
func Gia_SimRsbResubVerify(p *Gia_SimRsbMan_t, iObj int32, vFanins *Vec_Int_t) int32 {
    var pTruth0 *word  = Gia_SimRsbFunc( p, iObj, p.vFanins, 0 );
    var pTruth1 *word  = Gia_SimRsbFunc( p, iObj, p.vFanins, 1 );
    var Res int32  = !Abc_TtIntersect( pTruth0, pTruth1, p.nWords, 0 );
    // inlining of macro ABC_FREE
    pTruth0 = nil;
    // inlining of macro ABC_FREE
    pTruth1 = nil;
    return Res;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SimRsbSimAndCareSet(p *Gia_Man_t, i int32, pObj *Gia_Obj_t, nWords int32, vSims *Vec_Wrd_t, vSims2 *Vec_Wrd_t) {
    var pComps [2]word = [2]word{ 0, ~(word)0 };
    var Diff0 word  = pComps[Gia_ObjFaninC0(pObj)];
    var Diff1 word  = pComps[Gia_ObjFaninC1(pObj)];
    var vSims0 *Vec_Wrd_t;
    if Gia_ObjIsTravIdCurrentId(p, Gia_ObjFaninId0(pObj, i)) {
        vSims0 = vSims2;
    } else {
        vSims0 = vSims;
    }
    var vSims1 *Vec_Wrd_t;
    if Gia_ObjIsTravIdCurrentId(p, Gia_ObjFaninId1(pObj, i)) {
        vSims1 = vSims2;
    } else {
        vSims1 = vSims;
    }
    var pSims0 *word  = Vec_WrdEntryP( vSims0, nWords*Gia_ObjFaninId0(pObj, i) );
    var pSims1 *word  = Vec_WrdEntryP( vSims1, nWords*Gia_ObjFaninId1(pObj, i) );
    var pSims2 *word  = Vec_WrdEntryP( vSims2, nWords*i ); var w int32;
    if Gia_ObjIsXor(pObj) {
        for  w = 0; w < nWords; w++  {
            pSims2[w] = (pSims0[w] ^ Diff0) ^ (pSims1[w] ^ Diff1);
        }
    } else {
        for  w = 0; w < nWords; w++  {
            pSims2[w] = (pSims0[w] ^ Diff0) & (pSims1[w] ^ Diff1);
        } }
}
func Gia_SimRsbCareSet(p *Gia_SimRsbMan_t, iObj int32, vTfo *Vec_Int_t) *word {
    var pSims *word   = Vec_WrdEntryP( p.vSimsObj,  p.nWords*iObj );
    var pSims2 *word  = Vec_WrdEntryP( p.vSimsObj2, p.nWords*iObj );  var iNode, i int32;
    Abc_TtCopy( pSims2, pSims, p.nWords, 1 );
    Abc_TtClear( p.pFunc[2], p.nWords );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vTfo)) && (((iNode) = Vec_IntEntry(vTfo, i)), 1); i++  {
        var pNode *Gia_Obj_t  = Gia_ManObj(p.pGia, iNode);
        if Gia_ObjIsAnd(pNode) {
            Gia_SimRsbSimAndCareSet( p.pGia, iNode, pNode, p.nWords, p.vSimsObj, p.vSimsObj2 );
        } else if Gia_ObjIsCo(pNode) {
            var pSimsA *word  = Vec_WrdEntryP( p.vSimsObj,  p.nWords*Gia_ObjFaninId0p(p.pGia, pNode) );
            var pSimsB *word  = Vec_WrdEntryP( p.vSimsObj2, p.nWords*Gia_ObjFaninId0p(p.pGia, pNode) );
            Abc_TtOrXor( p.pFunc[2], pSimsA, pSimsB, p.nWords );
        
        } else { assert(false, "0"); }
    }
    return p.pFunc[2];
}


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjSimCollect(p *Gia_SimRsbMan_t) {
    var i, k, iTemp, iFanout int32;
    Vec_IntClear( p.vFanins2 );
    assert(Vec_IntSize(p.vFanins) > 0, "Vec_IntSize(p.vFanins) > 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vFanins)) && (((iTemp) = Vec_IntEntry(p.vFanins, i)), 1); i++  {
        var pObj *Gia_Obj_t  = Gia_ManObj( p.pGia, iTemp );
        if Gia_ObjIsAnd(pObj) && !Gia_ObjIsTravIdCurrentId( p.pGia, Gia_ObjFaninId0(pObj, iTemp) ) {
            Vec_IntPush( p.vFanins2, Gia_ObjFaninId0(pObj, iTemp) ); }
        if Gia_ObjIsAnd(pObj) && !Gia_ObjIsTravIdCurrentId( p.pGia, Gia_ObjFaninId1(pObj, iTemp) ) {
            Vec_IntPush( p.vFanins2, Gia_ObjFaninId1(pObj, iTemp) ); }
        // inlining of macro Gia_ObjForEachFanoutStaticId
        for  k = 0; (k < Gia_ObjFanoutNumId(p.pGia, iTemp))   && (((iFanout) = Gia_ObjFanoutId(p.pGia, iTemp, k)), 1); k++  {
            if Gia_ObjIsAnd(Gia_ManObj(p.pGia, iFanout)) && !Gia_ObjIsTravIdCurrentId( p.pGia, iFanout ) {
                Vec_IntPush( p.vFanins2, iFanout ); }
        }
    }
}
func Gia_ObjSimCands(p *Gia_SimRsbMan_t, iObj int32, nCands int32) *Vec_Int_t {
    assert(iObj > 0, "iObj > 0");
    assert(Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj)), "Gia_ObjIsAnd(Gia_ManObj(p.pGia, iObj))");
    Vec_IntClear( p.vCands );
    Vec_IntFill( p.vFanins, 1, iObj );
    for  Vec_IntSize(p.vFanins) > 0 && Vec_IntSize(p.vCands) < nCands  {
        var i, iTemp int32;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vFanins)) && (((iTemp) = Vec_IntEntry(p.vFanins, i)), 1); i++  {
            Gia_ObjSetTravIdCurrentId( p.pGia, iTemp );
        }
        Gia_ObjSimCollect( p ); // p->vFanins -> p->vFanins2
        Vec_IntAppend( p.vCands, p.vFanins2 );
        // inlining of macro ABC_SWAP
        { var t *Vec_Int_t  = p.vFanins; p.vFanins = p.vFanins2; p.vFanins2 = t; };
    }
    assert(Vec_IntSize(p.vFanins) == 0 || Vec_IntSize(p.vCands) >= nCands, "Vec_IntSize(p.vFanins) == 0 || Vec_IntSize(p.vCands) >= nCands");
    if Vec_IntSize(p.vCands) > nCands {
        Vec_IntShrink( p.vCands, nCands ); }
    return p.vCands;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjSimRsb(p *Gia_SimRsbMan_t, iObj int32, nCands int32, fVerbose int32, pnBufs *int32, pnInvs *int32) int32 {
    var i, iCand, RetValue int32;
    var vTfo *Vec_Int_t    = Gia_SimRsbTfo( p, iObj, -1 );
    var pCareSet *word     = Gia_SimRsbCareSet( p, iObj, vTfo );
    var pFunc *word        = Vec_WrdEntryP( p.vSimsObj, p.nWords*iObj );
    var vCands *Vec_Int_t  = Gia_ObjSimCands( p, iObj, nCands );
    Abc_TtAndSharp( p.pFunc[0], pCareSet, pFunc, p.nWords, 1 );
    Abc_TtAndSharp( p.pFunc[1], pCareSet, pFunc, p.nWords, 0 );

/*
printf( "Considering node %d with %d candidates:\n", iObj, Vec_IntSize(vCands) );
Vec_IntPrint( vTfo );
Vec_IntPrint( vCands );
Extra_PrintBinary( stdout, (unsigned *)pCareSet,    64 );  printf( "\n" );
Extra_PrintBinary( stdout, (unsigned *)pFunc,       64 );  printf( "\n" );
Extra_PrintBinary( stdout, (unsigned *)p->pFunc[0], 64 );  printf( "\n" );
Extra_PrintBinary( stdout, (unsigned *)p->pFunc[1], 64 );  printf( "\n" );
*/
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCands)) && (((iCand) = Vec_IntEntry(vCands, i)), 1); i++  {
        var pDiv *word  = Vec_WrdEntryP( p.vSimsObj, p.nWords*iCand );
        if !Abc_TtIntersect(pDiv, p.pFunc[0], p.nWords, 0) &&
             !Abc_TtIntersect(pDiv, p.pFunc[1], p.nWords, 1) { (*pnBufs)++; if fVerbose != 0 { fmt.Printf("Level %3d : %d = buf(%d)\n", Gia_ObjLevelId(p.pGia, iObj), iObj, iCand); }
 RetValue = 1; }
        if !Abc_TtIntersect(pDiv, p.pFunc[0], p.nWords, 1) &&
             !Abc_TtIntersect(pDiv, p.pFunc[1], p.nWords, 0) { (*pnInvs)++; if fVerbose != 0 { fmt.Printf("Level %3d : %d = inv(%d)\n", Gia_ObjLevelId(p.pGia, iObj), iObj, iCand); }
 RetValue = 1; }
    }
    return RetValue;
}

func Gia_ManSimRsb(pGia *Gia_Man_t, nCands int32, fVerbose int32) int32 {
    var clk abctime  = Abc_Clock();
    var pObj *Gia_Obj_t; var iObj, nCount, nBufs, nInvs int32;
    var p *Gia_SimRsbMan_t  = Gia_SimRsbAlloc( pGia );
    assert(pGia.vSimsPi != nil, "pGia.vSimsPi != nil");
    Gia_ManLevelNum( pGia );
    // inlining of macro Gia_ManForEachAnd
    for  iObj = 0; (iObj < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, iObj)); iObj++  {      if !Gia_ObjIsAnd(pObj) {} else {
        //if ( iObj == 6 )
        nCount += Gia_ObjSimRsb( p, iObj, nCands, fVerbose, &nBufs, &nInvs ); }
    }
    fmt.Printf("Can resubstitute %d nodes (%.2f %% out of %d) (Bufs = %d Invs = %d)  ", nCount, 100.0*nCount/Gia_ManAndNum(pGia), Gia_ManAndNum(pGia), nBufs, nInvs);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Gia_SimRsbFree( p );
    return nCount;
}




/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimRelAssignInputs(p *Gia_Man_t, nWords int32, vSims *Vec_Wrd_t, nWordsIn int32, vSimsIn *Vec_Wrd_t) {
    var i int32;
    var m int32;
    var Id int32;
    var nMints int32 = nWords / nWordsIn;
    assert(Vec_WrdSize(vSims)   == nWords   * Gia_ManObjNum(p), "Vec_WrdSize(vSims)   == nWords   * Gia_ManObjNum(p)");
    assert(Vec_WrdSize(vSimsIn) == nWordsIn * Gia_ManCiNum(p), "Vec_WrdSize(vSimsIn) == nWordsIn * Gia_ManCiNum(p)");
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        for  m = 0; m < nMints; m++  {
            memcpy( Vec_WrdEntryP(vSims, Id * nWords + nWordsIn * m), 
                                Vec_WrdEntryP(vSimsIn, i * nWordsIn), sizeof(word) * nWordsIn );
        }
    }
}
func Gia_ManSimRelCompare(p *Gia_Man_t, nWords int32, vSims *Vec_Wrd_t, nWordsOut int32, vSimsOut *Vec_Wrd_t, iPat int32, iMint int32) int32 {
    var i, Id int32;
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
        var pSim *word     = Vec_WrdEntryP( vSims, nWords * Id + iMint * nWordsOut );
        var pSimOut *word  = Vec_WrdEntryP( vSimsOut, nWordsOut * i );
/*
        int k;
        for ( k = 0; k < 64*nWordsOut; k++ )
            printf( "%d", Abc_TtGetBit( pSim, k ) );
        printf( "\n" );
        for ( k = 0; k < 64*nWordsOut; k++ )
            printf( "%d", Abc_TtGetBit( pSimOut, k ) );
        printf( "\n\n" );
*/
        if Abc_TtGetBit(pSim, iPat) != Abc_TtGetBit(pSimOut, iPat) {
            return 0; }
    }
    return 1;
}
func Gia_ManSimRelCollectOutputs(p *Gia_Man_t, nWords int32, vSims *Vec_Wrd_t, nWordsOut int32, vSimsOut *Vec_Wrd_t, vRel *Vec_Wrd_t) int32 {
    var i int32;
    var m int32;
    var nMints int32 = nWords / nWordsOut;
    var Count int32;
    assert(Vec_WrdSize(vSims)    == nWords   * Gia_ManObjNum(p), "Vec_WrdSize(vSims)    == nWords   * Gia_ManObjNum(p)");
    assert(Vec_WrdSize(vSimsOut) == nWordsOut * Gia_ManCoNum(p), "Vec_WrdSize(vSimsOut) == nWordsOut * Gia_ManCoNum(p)");
    assert(Vec_WrdSize(vRel)     == nWordsOut * nMints, "Vec_WrdSize(vRel)     == nWordsOut * nMints");
    for  i = 0; i < 64 * nWordsOut; i++  {
        var CountMints int32;
        for  m = 0; m < nMints; m++  {
            if Gia_ManSimRelCompare(p, nWords, vSims, nWordsOut, vSimsOut, i, m) != 0 {
                Abc_TtSetBit( Vec_WrdArray(vRel), i*nMints+m );
CountMints++; }
        }
        Count += CountMints == 0;
    }
    if Count != 0 {
        fmt.Printf("The relation is not well-defined for %d (out of %d) patterns.\n", Count, 64 * nWordsOut); }
    return Count;
}
func Gia_ManSimRel(p *Gia_Man_t, vObjs *Vec_Int_t, vVals *Vec_Wrd_t) *Vec_Wrd_t {
    var nWords int32  = Vec_WrdSize(p.vSimsPi) / Gia_ManCiNum(p);
    var nMints int32 = 1 << Vec_IntSize(vObjs);
    var i int32;
    var m int32;
    var iObj int32;
    var pObj *Gia_Obj_t;
    var vRel *Vec_Wrd_t   = Vec_WrdStart( nWords * nMints );
    var vSims *Vec_Wrd_t  = Vec_WrdStart( Gia_ManObjNum(p) * nWords * nMints );
    Gia_ManSimRelAssignInputs( p, nWords * nMints, vSims, nWords, p.vSimsPi );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        for  m = 0; m < nMints; m++  {
            if (m >> i) & 1 {
                memset( Vec_WrdEntryP(vSims, iObj*nMints*nWords + nWords*m), 0xFF, sizeof(word)*nWords );    
            } else {
                memset( Vec_WrdEntryP(vSims, iObj*nMints*nWords + nWords*m), 0x00, sizeof(word)*nWords ); }
        }
    }    
    Gia_ManCleanPhase( p );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        pObj.fPhase = 1;
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else 
        if !pObj.fPhase {
            Gia_ManSimPatSimAnd( p, i, pObj, nWords * nMints, vSims ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if !pObj.fPhase {
            Gia_ManSimPatSimPo( p, Gia_ObjId(p, pObj), pObj, nWords * nMints, vSims ); }
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        pObj.fPhase = 0;
    }
    if Gia_ManSimRelCollectOutputs( p, nWords * nMints, vSims, nWords, vVals, vRel ) != 0 {
        Vec_WrdFreeP( &vRel ); }
    Vec_WrdFree( vSims );
    return vRel;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimRelCheckFuncs(p *Gia_Man_t, vRel *Vec_Wrd_t, nOuts int32, vFuncs *Vec_Wrd_t) {
    var i int32;
    var k int32;
    var m int32;
    var Values [32]int32;
    var nErrors int32;
    var nMints int32 = 1 << nOuts;
    var nWords int32 = Vec_WrdSize(vRel) / nMints;
    assert(Vec_WrdSize(vFuncs) == 2 * nOuts * nWords, "Vec_WrdSize(vFuncs) == 2 * nOuts * nWords");
    assert(nOuts <= 32, "nOuts <= 32");
    for  i = 0; i < 64 * nWords; i++  {
        for  k = 0; k < nOuts; k++  {
            var Value0 int32  = Abc_TtGetBit( Vec_WrdEntryP(vFuncs, (2*k+0)*nWords), i );
            var Value1 int32  = Abc_TtGetBit( Vec_WrdEntryP(vFuncs, (2*k+1)*nWords), i );
            if (Value0 != 0) && !(Value1 != 0) {
                Values[k] = 1;
            } else if !(Value0 != 0) && (Value1 != 0) {
                Values[k] = 2;
            } else if !(Value0 != 0) && !(Value1 != 0) {
                Values[k] = 3;
            } else { assert(false, "0"); }
        }
        for  m = 0; m < nMints; m++  {
            for  k = 0; k < nOuts; k++  {
                if ((Values[k] >> ((m >> k) & 1)) & 1) == 0 {
                    break; }
            }
            if k < nOuts {
                continue; }
            if Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m ) {
                continue; }
            if nErrors++ == 0 {
                fmt.Printf("For pattern %d, minterm %d produced by function is not in the relation.\n", i, m); }
        }
    }
    if nErrors != 0 {    
        fmt.Printf("Total number of similar errors = %d.\n", nErrors);
    } else {
        fmt.Printf("The function agrees with the relation.\n"); }
}
func Gia_ManSimRelDeriveFuncs(p *Gia_Man_t, vRel *Vec_Wrd_t, nOuts int32) *Vec_Wrd_t {
    var i int32;
    var k int32;
    var m int32;
    var Count int32;
    var nMints int32 = 1 << nOuts;
    var nWords int32 = Vec_WrdSize(vRel) / nMints;
    var vFuncs *Vec_Wrd_t  = Vec_WrdStart( 2 * nOuts * nWords );
    assert(Vec_WrdSize(vRel) % nMints == 0, "Vec_WrdSize(vRel) % nMints == 0");
    for  i = 0; i < 64 * nWords; i++  {
        for  m = 0; m < nMints; m++  {
            if Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m ) {
                break; }
        }
        Count += m == nMints;
        for  k = 0; k < nOuts; k++  {
            if (m >> k) & 1 {
                Abc_TtSetBit( Vec_WrdEntryP(vFuncs, (2*k+1)*nWords), i );
            } else {
                Abc_TtSetBit( Vec_WrdEntryP(vFuncs, (2*k+0)*nWords), i ); }
        }
    }
    if Count != 0 {
        fmt.Printf("The relation is not well-defined for %d (out of %d) patterns.\n", Count, 64 * nWords);
    } else {
        fmt.Printf("The relation was successfully determized without don't-cares for %d patterns.\n", 64 * nWords); }
    Gia_ManSimRelCheckFuncs( p, vRel, nOuts, vFuncs );
    return vFuncs;
}
func Gia_ManSimRelDeriveFuncs2(p *Gia_Man_t, vRel *Vec_Wrd_t, nOuts int32) *Vec_Wrd_t {
    var i int32;
    var k int32;
    var m int32;
    var nDCs [32]int32 ;
    var Count int32;
    var nMints int32 = 1 << nOuts;
    var nWords int32 = Vec_WrdSize(vRel) / nMints;
    var vFuncs *Vec_Wrd_t  = Vec_WrdStart( 2 * nOuts * nWords );
    assert(Vec_WrdSize(vRel) % nMints == 0, "Vec_WrdSize(vRel) % nMints == 0");
    assert(nOuts <= 32, "nOuts <= 32");
    for  i = 0; i < 64 * nWords; i++  {
        for  m = 0; m < nMints; m++  {
            if Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m ) {
                break; }
        }
        Count += m == nMints;
        for  k = 0; k < nOuts; k++  {
            if Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+(m^(1<<k)) ) {
                nDCs[k]++;
                continue;
            }
            if (m >> k) & 1 {
                Abc_TtSetBit( Vec_WrdEntryP(vFuncs, (2*k+1)*nWords), i );
            } else {
                Abc_TtSetBit( Vec_WrdEntryP(vFuncs, (2*k+0)*nWords), i ); }
        }
        if false {
            for  m = 0; m < nMints; m++  {
                fmt.Printf("%d", Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m ));
            }
            fmt.Printf(" ");
            for  k = 0; k < nOuts; k++  {
                if Abc_TtGetBit( Vec_WrdEntryP(vFuncs, (2*k+0)*nWords), i ) {
                    fmt.Printf("0");
                } else if Abc_TtGetBit( Vec_WrdEntryP(vFuncs, (2*k+1)*nWords), i ) {
                    fmt.Printf("1");
                } else {
                    fmt.Printf("-"); }
            }
            fmt.Printf("\n");
        }
    }
    if Count != 0 {
        fmt.Printf("The relation is not well-defined for %d (out of %d) patterns.\n", Count, 64 * nWords);
    } else {
        fmt.Printf("The relation was successfully determized with don't-cares for %d patterns.\n", 64 * nWords);
        for  k = 0; k < nOuts; k++  {
            var nOffs int32  = Abc_TtCountOnesVec( Vec_WrdEntryP(vFuncs, (2*k+0)*nWords), nWords );
            var nOns int32   = Abc_TtCountOnesVec( Vec_WrdEntryP(vFuncs, (2*k+1)*nWords), nWords );
            fmt.Printf("%4d : Off = %6d  On = %6d  Dc = %6d (%6.2f %%)\n", k, nOffs, nOns, nDCs[k], 100.0*nDCs[k]/(64*nWords));
        }
        fmt.Printf("\n");
    }
    Gia_ManSimRelCheckFuncs( p, vRel, nOuts, vFuncs );
    return vFuncs;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimRelPrint(p *Gia_Man_t, vRel *Vec_Wrd_t, vOutMints *Vec_Int_t) {
    var nWords int32  = Vec_WrdSize(p.vSimsPi) / Gia_ManCiNum(p);
    var nMints int32  = Vec_WrdSize(vRel) / nWords; 
    var i, m, Count int32;
/*
    for ( i = 0; i < 64 * nWords; i++ )
    {
        int k;
        for ( k = 0; k < Gia_ManCiNum(p); k++ )
            printf( "%d", Abc_TtGetBit( Vec_WrdEntryP(p->vSimsPi, k), i ) );
        printf( " " );
        Count = 0;
        for ( m = 0; m < nMints; m++ )
        {
            printf( "%d", Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m ) );
            Count += Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m );
        }
        printf( "  Count = %2d ", Count );
        if ( vOutMints )
        {
            printf( "   %3d ", Vec_IntEntry(vOutMints, i) );
            if ( Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+Vec_IntEntry(vOutMints, i) ) )
                printf( "yes" );
            else
                printf( "no" );
        }
        printf( "\n" );
    }
*/
/*
    for ( i = 0; i < 64 * nWords; i++ )
    {
        Count = 0;
        for ( m = 0; m < nMints; m++ )
            Count += Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m );
        printf( "%d ", Count );
    }
    printf( "\n" );
*/
    for  i = 0; i < 64 * nWords; i++  {
        Count = 0;
        for  m = 0; m < nMints; m++  {
            fmt.Printf("%d", Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m ));
            Count += Abc_TtGetBit( Vec_WrdArray(vRel), i*nMints+m );
        }
        fmt.Printf("  Count = %2d \n", Count);
    }
}
func Gia_ManSimPatStart(nItems int32) *Vec_Int_t {
    var vValues *Vec_Int_t  = Vec_IntAlloc( nItems );
    Vec_IntPush( vValues, 17 );
    Vec_IntPush( vValues, 39 );
    Vec_IntPush( vValues, 56 );
    Vec_IntPush( vValues, 221 );
    return vValues;
}
func Gia_ManSimRelTest(p *Gia_Man_t) {
    //int nWords        = Vec_WrdSize(p->vSimsPi) / Gia_ManCiNum(p);
    var vObjs *Vec_Int_t  = Gia_ManSimPatStart( 4 ); // can be CI/AND/CO
    var vVals *Vec_Wrd_t  = Gia_ManSimPatValues( p );
    var vRel *Vec_Wrd_t   = Gia_ManSimRel( p, vObjs, vVals );
    assert(p.vSimsPi != nil, "p.vSimsPi != nil");
    Gia_ManSimRelPrint( p, vRel, nil );
    Vec_IntFree( vObjs );
    Vec_WrdFree( vVals );
    Vec_WrdFree( vRel );
}



/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Sim5CollectValues(pOffSet *word, pOnSet *word, nWords int32) *Vec_Int_t {
    var vBits *Vec_Int_t  = Vec_IntAlloc( 64*nWords ); var i int32;
var Count [2]int32 ;
    for  i = 0; i < 64*nWords; i++  {
        if Abc_TtGetBit( pOffSet, i ) {
            Vec_IntPush( vBits, 0 );
Count[0]++;
        } else if Abc_TtGetBit( pOnSet, i ) {
            Vec_IntPush( vBits, 1 );
Count[1]++;
        } else {
            Vec_IntPush( vBits, -1 ); }
    }
    //printf( "Offset = %d.  Onset = %d.  Dcset = %d.\n", Count[0], Count[1], 64*nWords - Count[0] - Count[1] );
    return vBits;
}
func Gia_SimAbsAlloc(pGia *Gia_Man_t, pOffSet *word, pOnSet *word, vSims *Vec_Wrd_t, nWords int32, vResub *Vec_Int_t, fVerbose int32) *Gia_SimAbsMan_t {
    // inlining of macro ABC_CALLOC
    var p *Gia_SimAbsMan_t  = &Gia_SimAbsMan_t{};
    p.pGia      = pGia;
    p.pSet[0]   = pOffSet;
    p.pSet[1]   = pOnSet;
    p.nCands    = Vec_WrdSize(vSims)/nWords;
    p.nWords    = nWords;
    p.vSims     = vSims;
    p.vResub    = vResub;
    p.fVerbose  = fVerbose;
    p.vValues      = Gia_Sim5CollectValues( pOffSet, pOnSet, nWords );
    p.vPatPairs    = Vec_IntAlloc( 100 );
    p.vCoverTable  = Vec_WrdAlloc( 10000 );
    p.vTtMints     = Vec_IntAlloc( 100 );
    assert(Vec_WrdSize(vSims) % nWords == 0, "Vec_WrdSize(vSims) % nWords == 0");
    return p;
}
func Gia_SimAbsFree(p *Gia_SimAbsMan_t) {
    Vec_IntFree( p.vValues );
    Vec_IntFree( p.vPatPairs );
    Vec_WrdFree( p.vCoverTable );
    Vec_IntFree( p.vTtMints );
    // inlining of macro ABC_FREE
    p = nil;
}


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SimAbsCheckSolution(p *Gia_SimAbsMan_t) {
    var x, y, z, w, fFound int32;
    assert(Vec_WrdSize(p.vCoverTable) == p.nWordsTable * (p.nCands+1), "Vec_WrdSize(p.vCoverTable) == p.nWordsTable * (p.nCands+1)");

    Abc_TtClear( p.pTableTemp, p.nWordsTable );
    for  x = 0; x < Vec_IntSize(p.vPatPairs)/2; x++  {
        Abc_TtXorBit( p.pTableTemp, x );
    }

    for  x =   0; x < p.nCands; x++  {
        var pSimTableX *word  = Vec_WrdEntryP( p.vCoverTable, p.nWordsTable * x );
        for  w = 0; w < p.nWordsTable; w++  {
            if p.pTableTemp[w] != pSimTableX[w] {
                break; }
        }
        if w == p.nWordsTable {
            fmt.Printf("Found solution { %d }\n", x);
            fFound = 1;
        }
    }
    if fFound != 0 {
        return; }

    for  x =   0; x < p.nCands; x++  {
    for  y =   0; y < x;         y++  {
        var pSimTableX *word  = Vec_WrdEntryP( p.vCoverTable, p.nWordsTable * x );
        var pSimTableY *word  = Vec_WrdEntryP( p.vCoverTable, p.nWordsTable * y );
        for  w = 0; w < p.nWordsTable; w++  {
            if p.pTableTemp[w] != (pSimTableX[w] | pSimTableY[w]) {
                break; }
        }
        if w == p.nWordsTable {
            fmt.Printf("Found solution { %d %d }\n", y, x);
            fFound = 1;
        }
    }
    }
    if fFound != 0 {
        return; }

    for  x =   0; x < p.nCands; x++  {
    for  y =   0; y < x;         y++  {
    for  z =   0; z < y;         z++  {
        var pSimTableX *word  = Vec_WrdEntryP( p.vCoverTable, p.nWordsTable * x );
        var pSimTableY *word  = Vec_WrdEntryP( p.vCoverTable, p.nWordsTable * y );
        var pSimTableZ *word  = Vec_WrdEntryP( p.vCoverTable, p.nWordsTable * z );
        for  w = 0; w < p.nWordsTable; w++  {
            if p.pTableTemp[w] != (pSimTableX[w] | pSimTableY[w] | pSimTableZ[w]) {
                break; }
        }
        if w == p.nWordsTable {
            fmt.Printf("Found solution { %d %d %d }\n", z, y, x); }
    }
    }
    }
}

func Gia_SimAbsSolve(p *Gia_SimAbsMan_t) {
    var clk abctime  = Abc_Clock();
    var i, k, iPat, iPat2 int32;
/*
    Vec_Int_t * vSimPats = Vec_IntDup( p->vPatPairs );
    Vec_IntUniqify( vSimPats );
    printf( "Selected %d pattern pairs contain %d unique patterns.\n", Vec_IntSize(p->vPatPairs)/2, Vec_IntSize(vSimPats) );
    Vec_IntFree( vSimPats );
*/
    // set up the covering problem
    p.nWordsTable = Abc_Bit6WordNum( Vec_IntSize(p.vPatPairs)/2 );
    Vec_WrdFill( p.vCoverTable, p.nWordsTable * (p.nCands + 1), 0 );
    p.pTableTemp = Vec_WrdEntryP( p.vCoverTable, p.nWordsTable * p.nCands );
    for  i = 0; i < p.nCands; i++  {
        var pSimCand *word   = Vec_WrdEntryP( p.vSims, p.nWords * i );
        var pSimTable *word  = Vec_WrdEntryP( p.vCoverTable, p.nWordsTable * i );
        //printf( "%4d : ", i );
        //Extra_PrintBinary( stdout, (word *)pSimCand, p->nCands );  printf( "\n" );
        // inlining of macro Vec_IntForEachEntryDouble
        for  k = 0; (k+1 < Vec_IntSize(p.vPatPairs)) && (((iPat) = Vec_IntEntry(p.vPatPairs, k)), 1) && (((iPat2) = Vec_IntEntry(p.vPatPairs, k+1)), 1); k += 2  {
            assert(Vec_IntEntry(p.vValues, iPat)  == 0, "Vec_IntEntry(p.vValues, iPat)  == 0");
            assert(Vec_IntEntry(p.vValues, iPat2) == 1, "Vec_IntEntry(p.vValues, iPat2) == 1");
            if Abc_TtGetBit(pSimCand, iPat) != Abc_TtGetBit(pSimCand, iPat2) {
                Abc_TtXorBit(pSimTable, k/2); }
        }
        assert(k == Vec_IntSize(p.vPatPairs), "k == Vec_IntSize(p.vPatPairs)");
    }

    if false {
        fmt.Printf("                  ");
        for  i = 0; i < p.nCands; i++  {
            fmt.Printf("%d", i % 10);
        }
        fmt.Printf("\n");

        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(p.vPatPairs)) && (((iPat) = Vec_IntEntry(p.vPatPairs, i)), 1) && (((iPat2) = Vec_IntEntry(p.vPatPairs, i+1)), 1); i += 2  {
            fmt.Printf("%4d  ", i/2);
            fmt.Printf("%4d  ", iPat);
            fmt.Printf("%4d  ", iPat2);
            for  k = 0; k < p.nCands; k++  {
                var pSimTable *word  = Vec_WrdEntryP( p.vCoverTable, p.nWordsTable * k );
                fmt.Printf("%c", Abc_TtGetBit(pSimTable, i/2) ? '*' : ' ');
            }
            fmt.Printf("\n");
        }
    }

    //Gia_SimAbsCheckSolution(p);

    Vec_IntClear( p.vResub );
    Abc_TtClear( p.pTableTemp, p.nWordsTable );
    for  i = 0; i < Vec_IntSize(p.vPatPairs)/2; i++  {
        Abc_TtXorBit( p.pTableTemp, i );
    }

    for  !Abc_TtIsConst0(p.pTableTemp, p.nWordsTable)  {
        var pSimTable *word;
        var iArgMax int32 = -1;
        var CostThis int32;
        var CostMax int32 = -1;
        // compute the cost of each column
        for  i = 0; i < p.nCands; i++  {
            pSimTable = Vec_WrdEntryP( p.vCoverTable, p.nWordsTable * i );
            CostThis = Abc_TtCountOnesVecMask( pSimTable, p.pTableTemp, p.nWordsTable, 0 );
            if CostMax >= CostThis {
                continue; }
            CostMax = CostThis;
            iArgMax = i;
        }
        // find the best column
        Vec_IntPush( p.vResub, iArgMax );
        // delete values of this column
        pSimTable = Vec_WrdEntryP( p.vCoverTable, p.nWordsTable * iArgMax );
        Abc_TtSharp( p.pTableTemp, p.pTableTemp, pSimTable, p.nWordsTable );
    }
    if p.fVerbose {
        fmt.Printf("Solution %2d for covering problem [%5d x %5d]: ", Vec_IntSize(p.vResub), Vec_IntSize(p.vPatPairs)/2, p.nCands);
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vResub)) && (((iPat) = Vec_IntEntry(p.vResub, i)), 1); i++  {
            fmt.Printf("%6d ", iPat);
        }
        for  ; i < 12; i++  {
            fmt.Printf("       ");
        }
        fmt.Printf("   ");
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
}
func Gia_SimAbsRefine(p *Gia_SimAbsMan_t) int32 {
    var i, b, Value, iPat, iMint, iObj, Count int32;
    // inlining of macro ABC_ALLOC
    var pFanins **word  = ((word * *) malloc(sizeof(word *) * (size_t)(Vec_IntSize(p.vResub))));
    assert(Vec_IntSize(p.vResub) > 0, "Vec_IntSize(p.vResub) > 0");
    // inlining of macro Vec_IntForEachEntry
    for  b = 0; (b < Vec_IntSize(p.vResub)) && (((iObj) = Vec_IntEntry(p.vResub, b)), 1); b++  {
        pFanins[b] = Vec_WrdEntryP( p.vSims, p.nWords * iObj );
    }
    Vec_IntFill( p.vTtMints, 1 << Vec_IntSize(p.vResub), -1 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vValues)) && (((Value) = Vec_IntEntry(p.vValues, i)), 1); i++  {
        if Value == -1 {
            continue; }
        iMint = 0;
        for  b = 0; b < Vec_IntSize(p.vResub); b++  {
            if Abc_TtGetBit(pFanins[b], i) {
                iMint |= 1 << b; }
        }
        iPat = Vec_IntEntry( p.vTtMints, iMint );
        if iPat == -1 {
            Vec_IntWriteEntry( p.vTtMints, iMint, i );
            continue;
        }
        assert(Abc_TtGetBit(p.pSet[Value], i), "Abc_TtGetBit(p.pSet[Value], i)");
        if Abc_TtGetBit(p.pSet[Value], iPat) {
            continue; }
        assert(Abc_TtGetBit(p.pSet[!Value], iPat), "Abc_TtGetBit(p.pSet[!Value], iPat)");
        Vec_IntPushTwo( p.vPatPairs, Value ? iPat : i, Value ? i : iPat );
        //printf( "iPat1 = %d  iPat2 = %d   Mint = %d\n", Value ? iPat : i, Value ? i : iPat, iMint );
        Count++;
        if Count == 64 {
            // inlining of macro ABC_FREE
            pFanins = nil;
            return 1;
        }
    }
    //printf( "Refinement added %d minterm pairs.\n", Count );
    // inlining of macro ABC_FREE
    pFanins = nil;
    return Count != 0;
}
func Gia_SimAbsFind(vValues *Vec_Int_t, Value int32) *Vec_Int_t {
    var vSubset *Vec_Int_t  = Vec_IntAlloc( 100 ); var i, Entry int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vValues)) && (((Entry) = Vec_IntEntry(vValues, i)), 1); i++  {
        if Entry == Value {
            Vec_IntPush( vSubset, i ); }
    }
    return vSubset;
}
func Gia_SimAbsInit(p *Gia_SimAbsMan_t) {
    var n int32;
    var nPairsInit int32 = 64;
    var vValue0 *Vec_Int_t  = Gia_SimAbsFind( p.vValues, 0 );
    var vValue1 *Vec_Int_t  = Gia_SimAbsFind( p.vValues, 1 );
    Vec_IntClear( p.vPatPairs );
    fmt.Printf("There are %d offset and %d onset minterms (%d pairs) and %d divisors.\n", Vec_IntSize(vValue0), Vec_IntSize(vValue1), Vec_IntSize(vValue0)*Vec_IntSize(vValue1), p.nCands);
    Abc_Random( 1 );
    assert(Vec_IntSize(vValue0) > 0, "Vec_IntSize(vValue0) > 0");
    assert(Vec_IntSize(vValue1) > 0, "Vec_IntSize(vValue1) > 0");
    for  n = 0; n < nPairsInit; n++  {
        Vec_IntPushTwo( p.vPatPairs, 
            Vec_IntEntry(vValue0, Abc_Random(0) % Vec_IntSize(vValue0)), 
            Vec_IntEntry(vValue1, Abc_Random(0) % Vec_IntSize(vValue1)) );
    }
    Vec_IntFree( vValue0 );
    Vec_IntFree( vValue1 );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SimAbsPerformOne(pGia *Gia_Man_t, pOffSet *word, pOnSet *word, vSimsCands *Vec_Wrd_t, nWords int32, fVerbose int32) *Vec_Int_t {
    var clk abctime  = Abc_Clock();
    var vResub *Vec_Int_t   = Vec_IntAlloc( 10 );
    var p *Gia_SimAbsMan_t  = Gia_SimAbsAlloc( pGia, pOffSet, pOnSet, vSimsCands, nWords, vResub, fVerbose );
    Gia_SimAbsInit( p );
    for {
        Gia_SimAbsSolve( p );
        if !(Gia_SimAbsRefine( p ) != 0) {
            break; }
    }
    Gia_SimAbsFree( p );
    Abc_PrintTime( 1, "Resubstitution time", Abc_Clock() - clk );
    return vResub;
}



/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

   
type Gia_RsbMan_t struct {
    pGia     *Gia_Man_t;
    pOffSet          *word;
    pOnSet          *word;
    nWords            int32;
    nWordsT            int32;
    vSims     *Vec_Wrd_t;
    vSimsT     *Vec_Wrd_t;
    vCands     *Vec_Int_t;
    vObjs     *Vec_Int_t;
    vObjs2     *Vec_Int_t;
    Vec_Wec_t *    vSets[2];
    word *         pSet[3];
    vActive     *Vec_Int_t;
}
func Gia_RsbAlloc(pGia *Gia_Man_t, pOffSet *word, pOnSet *word, vSims *Vec_Wrd_t, nWords int32, vSimsT *Vec_Wrd_t, nWordsT int32, vCands *Vec_Int_t) *Gia_RsbMan_t {
    var i, iObj int32;
    // inlining of macro ABC_CALLOC
    var p *Gia_RsbMan_t  = &Gia_RsbMan_t{};
    assert(nWords <= 1024, "nWords <= 1024");
    assert(Vec_WrdSize(vSims) == 64 * nWords * nWordsT, "Vec_WrdSize(vSims) == 64 * nWords * nWordsT");
    assert(Vec_WrdSize(vSims) == Vec_WrdSize(vSimsT), "Vec_WrdSize(vSims) == Vec_WrdSize(vSimsT)");
    p.pGia      = pGia;
    p.pOffSet   = pOffSet;
    p.pOnSet    = pOnSet;
    p.nWords    = nWords;
    p.nWordsT   = nWordsT;
    p.vSims     = vSims;
    p.vSimsT    = vSimsT;
    p.vCands    = vCands;
    p.vObjs     = Vec_IntAlloc( 100 );
    p.vObjs2    = Vec_IntAlloc( 100 );
    p.vSets[0]  = Vec_WecAlloc( 1024 );
    p.vSets[1]  = Vec_WecAlloc( 1024 );
    // inlining of macro ABC_CALLOC
    p.pSet[0]   = ((word *) calloc((size_t)(nWordsT), sizeof(word)));
    // inlining of macro ABC_CALLOC
    p.pSet[1]   = ((word *) calloc((size_t)(nWordsT), sizeof(word)));
    // inlining of macro ABC_CALLOC
    p.pSet[2]   = ((word *) calloc((size_t)(nWordsT), sizeof(word)));
    p.vActive   = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCands)) && (((iObj) = Vec_IntEntry(vCands, i)), 1); i++  {
        assert(iObj < nWordsT * 64, "iObj < nWordsT * 64");
        Abc_TtSetBit( p.pSet[0], iObj );
    }
    Vec_WecPushLevel( p.vSets[0] );
    Vec_WecPushLevel( p.vSets[1] );
    for  i = 0; i < 64*nWords; i++  {
        var Value0 int32  = Abc_TtGetBit( pOffSet, i );
        var Value1 int32  = Abc_TtGetBit( pOnSet,  i );
        if (Value0 != 0) && !(Value1 != 0) {
            Vec_WecPush( p.vSets[0], 0, i );
        } else if !(Value0 != 0) && (Value1 != 0) {
            Vec_WecPush( p.vSets[1], 0, i );
        } else { assert(!(Value0 != 0) || !(Value1 != 0), "!Value0 || !Value1"); }
    }
    assert(Vec_WecSize(p.vSets[0]) == 1, "Vec_WecSize(p.vSets[0]) == 1");
    assert(Vec_WecSize(p.vSets[1]) == 1, "Vec_WecSize(p.vSets[1]) == 1");
    Abc_Random( 1 );
    //Extra_PrintBinary2( stdout, (unsigned*)pOffSet, 64*nWords ); printf( "\n" );
    //Extra_PrintBinary2( stdout, (unsigned*)pOnSet,  64*nWords ); printf( "\n" );
    return p;
}
func Gia_RsbFree(p *Gia_RsbMan_t) {
    Vec_IntFree( p.vActive );
    Vec_IntFree( p.vObjs );
    Vec_IntFree( p.vObjs2 );
    Vec_WecFree( p.vSets[0] );
    Vec_WecFree( p.vSets[1] );
    // inlining of macro ABC_FREE
    p.pSet[0] = nil;
    // inlining of macro ABC_FREE
    p.pSet[1] = nil;
    // inlining of macro ABC_FREE
    p.pSet[2] = nil;
    // inlining of macro ABC_FREE
    p = nil;
}


func Gia_RsbCost(p *Gia_RsbMan_t) int32 {
    var vLevel [2]*Vec_Int_t; var i, Cost int32;
    // inlining of macro Vec_WecForEachLevelTwo
    for  i = 0; (i < Vec_WecSize(p.vSets[0])) && (((vLevel[0]) = Vec_WecEntry(p.vSets[0], i)), 1) && (((vLevel[1]) = Vec_WecEntry(p.vSets[1], i)), 1); i++  {
        Cost += Vec_IntSize(vLevel[0]) * Vec_IntSize(vLevel[1]);
    }
    return Cost;
}
func Gia_RsbPrint(p *Gia_RsbMan_t) {
    var vLevel [2]*Vec_Int_t; 
    var n int32;
    var i int32;
    var nLeaves int32 = 1 << Vec_IntSize(p.vObjs);
    assert(Vec_WecSize(p.vSets[0]) == nLeaves, "Vec_WecSize(p.vSets[0]) == nLeaves");
    assert(Vec_WecSize(p.vSets[1]) == nLeaves, "Vec_WecSize(p.vSets[1]) == nLeaves");
    fmt.Printf("Database for %d objects and cost %d:\n", Vec_IntSize(p.vObjs), Gia_RsbCost(p));
    // inlining of macro Vec_WecForEachLevelTwo
    for  i = 0; (i < Vec_WecSize(p.vSets[0])) && (((vLevel[0]) = Vec_WecEntry(p.vSets[0], i)), 1) && (((vLevel[1]) = Vec_WecEntry(p.vSets[1], i)), 1); i++  {
        for  n = 0; n < 2; n++  {
            fmt.Printf("%5d : ", i);
            Extra_PrintBinary2( os.Stdout, (unsigned*)&i, Vec_IntSize(p.vObjs) ); fmt.Printf(" %d ", n);
            Vec_IntPrint( vLevel[n] );
        }
    }    
}
func Gia_RsbUpdateAdd(p *Gia_RsbMan_t, iObj int32) {
    var n int32;
    var i int32;
    var nLeaves int32 = 1 << Vec_IntSize(p.vObjs);
    assert(Vec_WecSize(p.vSets[0]) == nLeaves, "Vec_WecSize(p.vSets[0]) == nLeaves");
    assert(Vec_WecSize(p.vSets[1]) == nLeaves, "Vec_WecSize(p.vSets[1]) == nLeaves");
    for  i = 0; i < nLeaves; i++  {
        for  n = 0; n < 2; n++  {
            var vLevelN *Vec_Int_t  = Vec_WecPushLevel(p.vSets[n]);  
            var vLevel *Vec_Int_t   = Vec_WecEntry(p.vSets[n], i);
            var iMint, j, k int32;
            // inlining of macro Vec_IntForEachEntry
            for  j = 0; (j < Vec_IntSize(vLevel)) && (((iMint) = Vec_IntEntry(vLevel, j)), 1); j++  {
                if Abc_TtGetBit(Vec_WrdEntryP(p.vSims, p.nWords*iObj), iMint) {
                    Vec_IntPush( vLevelN, iMint );
                } else {
                    Vec_IntWriteEntry( vLevel, k++, iMint ); }
            }
            Vec_IntShrink( vLevel, k );
        }
    }
    Vec_IntPush( p.vObjs, iObj );
    assert(Vec_WecSize(p.vSets[0]) == 2*nLeaves, "Vec_WecSize(p.vSets[0]) == 2*nLeaves");
    assert(Vec_WecSize(p.vSets[1]) == 2*nLeaves, "Vec_WecSize(p.vSets[1]) == 2*nLeaves");
}
func Gia_RsbUpdateRemove(p *Gia_RsbMan_t, Index int32) {
    var vLevel [2]*Vec_Int_t;
    var vTemp [2][2]*Vec_Int_t;
    var k int32;
    var m int32;
    var m2 int32;
    var nLeaves int32 = 1 << Vec_IntSize(p.vObjs);
    assert(Index < Vec_IntSize(p.vObjs), "Index < Vec_IntSize(p.vObjs)");
    assert(Vec_WecSize(p.vSets[0]) == nLeaves, "Vec_WecSize(p.vSets[0]) == nLeaves");
    assert(Vec_WecSize(p.vSets[1]) == nLeaves, "Vec_WecSize(p.vSets[1]) == nLeaves");
    for  m = 0; m < nLeaves; m++  {
        if m & (1 << Index) {
            continue; }
        m2 = m ^ (1 << Index);
        vTemp[0][0] = Vec_WecEntry(p.vSets[0], m);
        vTemp[0][1] = Vec_WecEntry(p.vSets[1], m);
        vTemp[1][0] = Vec_WecEntry(p.vSets[0], m2);
        vTemp[1][1] = Vec_WecEntry(p.vSets[1], m2);
        Vec_IntAppend( vTemp[0][0], vTemp[1][0] );
        Vec_IntAppend( vTemp[0][1], vTemp[1][1] );
        Vec_IntClear( vTemp[1][0] );
        Vec_IntClear( vTemp[1][1] );
    }
    Vec_IntDrop( p.vObjs, Index );
    // inlining of macro Vec_WecForEachLevelTwo
    for  m = 0; (m < Vec_WecSize(p.vSets[0])) && (((vLevel[0]) = Vec_WecEntry(p.vSets[0], m)), 1) && (((vLevel[1]) = Vec_WecEntry(p.vSets[1], m)), 1); m++  {
        if m & (1 << Index) {
            continue; }
        // inlining of macro ABC_SWAP
        { var t Vec_Int_t  = Vec_WecArray(p.vSets[0])[k]; Vec_WecArray(p.vSets[0])[k] = Vec_WecArray(p.vSets[0])[m]; Vec_WecArray(p.vSets[0])[m] = t; };
        // inlining of macro ABC_SWAP
        { var t Vec_Int_t  = Vec_WecArray(p.vSets[1])[k]; Vec_WecArray(p.vSets[1])[k] = Vec_WecArray(p.vSets[1])[m]; Vec_WecArray(p.vSets[1])[m] = t; };
        k++;
    }
    assert(k == nLeaves/2, "k == nLeaves/2");
    Vec_WecShrink( p.vSets[0], k );
    Vec_WecShrink( p.vSets[1], k );
}
func Gia_RsbRemovalCost(p *Gia_RsbMan_t, Index int32) int32 {
    var vTemp [2][2]*Vec_Int_t;
    //unsigned Mask = Abc_InfoMask( Index );
    var m int32;
    var m2 int32;
    var Cost int32;
    var nLeaves int32 = 1 << Vec_IntSize(p.vObjs);
    assert(Vec_WecSize(p.vSets[0]) == (1 << Vec_IntSize(p.vObjs)), "Vec_WecSize(p.vSets[0]) == (1 << Vec_IntSize(p.vObjs))");
    assert(Vec_WecSize(p.vSets[1]) == (1 << Vec_IntSize(p.vObjs)), "Vec_WecSize(p.vSets[1]) == (1 << Vec_IntSize(p.vObjs))");
    for  m = 0; m < nLeaves; m++  {
        if m & (1 << Index) {
            continue; }
        m2 = m ^ (1 << Index);
        vTemp[0][0] = Vec_WecEntry(p.vSets[0], m);
        vTemp[0][1] = Vec_WecEntry(p.vSets[1], m);
        vTemp[1][0] = Vec_WecEntry(p.vSets[0], m2);
        vTemp[1][1] = Vec_WecEntry(p.vSets[1], m2);
        Cost += (Vec_IntSize(vTemp[0][0]) + Vec_IntSize(vTemp[1][0])) * (Vec_IntSize(vTemp[0][1]) + Vec_IntSize(vTemp[1][1]));
    }
    return Cost;
}
func Gia_RsbFindNodeToRemove(p *Gia_RsbMan_t, pMinCost *int32) int32 {
    var i int32;
    var iObj int32;
    var iMin int32 = -1;
    var CostMin int32 = ABC_INFINITY;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vObjs)) && (((iObj) = Vec_IntEntry(p.vObjs, i)), 1); i++  {
        var Cost int32  = Gia_RsbRemovalCost( p, i );
        if CostMin > Cost {
            CostMin = Cost;
            iMin = i;
        }
    }
    if pMinCost != nil {
        *pMinCost = CostMin; }
    return iMin;
}

func Gia_RsbFindMints(p *Gia_RsbMan_t, pMint0 *int32, pMint1 *int32) {
    var iSetI int32  = Abc_Random(0) % Vec_IntSize(p.vActive);
    var iSet int32   = Vec_IntEntry( p.vActive, iSetI );
    var vArray0 *Vec_Int_t  = Vec_WecEntry(p.vSets[0], iSet);
    var vArray1 *Vec_Int_t  = Vec_WecEntry(p.vSets[1], iSet);
    var iMint0i int32  = Abc_Random(0) % Vec_IntSize(vArray0);
    var iMint1i int32  = Abc_Random(0) % Vec_IntSize(vArray1);
    var iMint0 int32  = Vec_IntEntry( vArray0, iMint0i );
    var iMint1 int32  = Vec_IntEntry( vArray1, iMint1i );
    *pMint0 = iMint0;
    *pMint1 = iMint1;
}
func Gia_RsbFindNode(p *Gia_RsbMan_t) int32 {
    var i int32;
    var iObj int32;
    var nNodes int32;
    var nNodesNew int32 = -1;
    var nNodesOld int32 = -1;
    var Mint0 int32;
    var Mint1 int32;
    var Shift int32;
    Abc_TtCopy( p.pSet[1], p.pSet[0], p.nWordsT, 0 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vObjs)) && (((iObj) = Vec_IntEntry(p.vObjs, i)), 1); i++  {
        assert(Abc_TtGetBit(p.pSet[1], iObj), "Abc_TtGetBit(p.pSet[1], iObj)");
        Abc_TtXorBit(p.pSet[1], iObj);
    }
    Abc_TtCopy( p.pSet[2], p.pSet[1], p.nWordsT, 0 );
    Gia_RsbFindMints( p, &Mint0, &Mint1 );
    nNodes = Abc_TtAndXorSum( p.pSet[1], Vec_WrdEntryP(p.vSimsT, p.nWordsT*Mint0), Vec_WrdEntryP(p.vSimsT, p.nWordsT*Mint1), p.nWordsT );
    for  i = 0; i < 5 && nNodes > 1; i++  {
        nNodesOld = nNodes;
        Abc_TtCopy( p.pSet[2], p.pSet[1], p.nWordsT, 0 );
        Gia_RsbFindMints( p, &Mint0, &Mint1 );
        nNodesNew = Abc_TtAndXorSum( p.pSet[1], Vec_WrdEntryP(p.vSimsT, p.nWordsT*Mint0), Vec_WrdEntryP(p.vSimsT, p.nWordsT*Mint1), p.nWordsT );
        assert(nNodesNew <= nNodes, "nNodesNew <= nNodes");
        if nNodesNew < nNodes {
            i = 0; }
        nNodes = nNodesNew;
    }
    Shift = Abc_Random(0) % (64*p.nWordsT);
    for  i = 0; i < 64*p.nWordsT; i++  {
        var Index int32  = (i+Shift) % (64*p.nWordsT);
        if Abc_TtGetBit( p.pSet[2], Index ) {
            return Index; }
    }
    assert(false, "0");      
    return -1;
}
func Gia_RsbCollectValid(p *Gia_RsbMan_t) int32 {
    var vLevel [2]*Vec_Int_t; var i int32;
    Vec_IntClear( p.vActive );
    assert(Vec_WecSize(p.vSets[0]) == Vec_WecSize(p.vSets[1]), "Vec_WecSize(p.vSets[0]) == Vec_WecSize(p.vSets[1])");
    // inlining of macro Vec_WecForEachLevelTwo
    for  i = 0; (i < Vec_WecSize(p.vSets[0])) && (((vLevel[0]) = Vec_WecEntry(p.vSets[0], i)), 1) && (((vLevel[1]) = Vec_WecEntry(p.vSets[1], i)), 1); i++  {
        if Vec_IntSize(vLevel[0]) && Vec_IntSize(vLevel[1]) {
            Vec_IntPush( p.vActive, i ); }
    }
    if Vec_IntSize(p.vActive) == 0 {
        return 0; }
    return 1;
}
func Gia_RsbSolve(p *Gia_RsbMan_t) *Vec_Int_t {
    var i, iMin int32;
    Vec_IntClear( p.vObjs );
    for  Gia_RsbCollectValid(p) != 0  {
        Gia_RsbUpdateAdd( p, Gia_RsbFindNode(p) );
    }
    for  i = 0; i < 100; i++  {
        var k int32;
        var nUndo int32 = 1 + Abc_Random(0) % Vec_IntSize(p.vObjs);
        for  k = 0; k < nUndo; k++  {
            iMin = Gia_RsbFindNodeToRemove( p, nil );// &MinCost );
            Gia_RsbUpdateRemove( p, iMin );
        }
        for  Gia_RsbCollectValid(p) != 0  {
            Gia_RsbUpdateAdd( p, Gia_RsbFindNode(p) );
        }
        if Vec_IntSize(p.vObjs2) == 0 || Vec_IntSize(p.vObjs2) > Vec_IntSize(p.vObjs) {
            Vec_IntClear( p.vObjs2 );
            Vec_IntAppend( p.vObjs2, p.vObjs );
        }
    }
    //Gia_RsbPrint( p );
    return Vec_IntDup( p.vObjs2 );
}
func Gia_RsbSetFind(pOffSet *word, pOnSet *word, vSims *Vec_Wrd_t, nWords int32, vSimsT *Vec_Wrd_t, nWordsT int32, vCands *Vec_Int_t) *Vec_Int_t {
    var p *Gia_RsbMan_t   = Gia_RsbAlloc( nil, pOffSet, pOnSet, vSims, nWords, vSimsT, nWordsT, vCands );
    var vObjs *Vec_Int_t  = Gia_RsbSolve( p );
    Gia_RsbFree( p );
    Vec_IntSort( vObjs, 0 );
    return vObjs;
}

/**Function*************************************************************

  Synopsis    [Improving quality of simulation patterns.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SimQualityOne(p *Gia_Man_t, vPat *Vec_Int_t, fPoOnly int32) *Vec_Int_t {
    var i int32;
    var Id int32;
    var Value int32;
    var nWords int32 = Abc_Bit6WordNum( 1+Gia_ManCiNum(p) );
    var vTemp *Vec_Wrd_t;
    var vSims *Vec_Wrd_t;
    var vSimsPi *Vec_Wrd_t = Vec_WrdStart( Gia_ManCiNum(p) * nWords );
    var vRes *Vec_Int_t;
    assert(Vec_IntSize(vPat) == Gia_ManCiNum(p), "Vec_IntSize(vPat) == Gia_ManCiNum(p)");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPat)) && (((Value) = Vec_IntEntry(vPat, i)), 1); i++  {
        var pSim *word  = Vec_WrdEntryP( vSimsPi, i*nWords );
        if Value != 0 {
            Abc_TtFill( pSim, nWords ); }
        Abc_TtXorBit( pSim, i+1 );
    }
    vTemp = p.vSimsPi;
    p.vSimsPi = vSimsPi;
    vSims = Gia_ManSimPatSim( p );
    p.vSimsPi = vTemp;
    if fPoOnly != 0 {
        vRes = Vec_IntStart( Gia_ManCoNum(p) );
        // inlining of macro Gia_ManForEachCoId
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
            var pSim *word  = Vec_WrdEntryP( vSims, Id*nWords );
            if pSim[0] & 1 {
                Abc_TtNot( pSim, nWords ); }
            Vec_IntWriteEntry( vRes, i, Abc_TtCountOnesVec(pSim, nWords) );
        }
        assert(Vec_IntSize(vRes) == Gia_ManCoNum(p), "Vec_IntSize(vRes) == Gia_ManCoNum(p)");
    
    } else {
        vRes = Vec_IntStart( Gia_ManObjNum(p) );
        // inlining of macro Gia_ManForEachAndId
        for  Id = 0; (Id < p.nObjs); Id++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p, Id)) {} else {
            var pSim *word  = Vec_WrdEntryP( vSims, Id*nWords );
            if pSim[0] & 1 {
                Abc_TtNot( pSim, nWords ); }
            Vec_IntWriteEntry( vRes, Id, Abc_TtCountOnesVec(pSim, nWords) );
        }
        }
        assert(Vec_IntSize(vRes) == Gia_ManObjNum(p), "Vec_IntSize(vRes) == Gia_ManObjNum(p)");
    }
    Vec_WrdFree( vSims );
    Vec_WrdFree( vSimsPi );
    return vRes;
}
func Gia_SimQualityTest(p *Gia_Man_t) {
    var vPat, vRes *Vec_Int_t;
    var k int32;
    var m int32;
    var nMints int32 = (1 << Gia_ManCiNum(p));
    assert(Gia_ManCiNum(p) <= 10, "Gia_ManCiNum(p) <= 10");
    for  m = 0; m < nMints; m++  {
        fmt.Printf("%d : ", m);
        Extra_PrintBinary( os.Stdout, (unsigned*)&m, Gia_ManCiNum(p) );
        fmt.Printf(" ");
        vPat = Vec_IntAlloc( Gia_ManCiNum(p) );
        for  k = 0; k < Gia_ManCiNum(p); k++  {
            Vec_IntPush( vPat, (m >> k) & 1 );
        }
        vRes = Gia_SimQualityOne( p, vPat, 1 );
        fmt.Printf("%d ", Vec_IntSum(vRes));
        Vec_IntFree( vRes );
        Vec_IntFree( vPat );
        fmt.Printf("\n");
    }
}
func Gia_SimGenerateStats(p *Gia_Man_t) *Vec_Int_t {
    var vTotal *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var vRes, vPat *Vec_Int_t;
    var i, k, Value int32;
    Abc_Random(1);
    for  i = 0; i < 1000; i++  {
        vPat = Vec_IntAlloc( Gia_ManCiNum(p) );
        for  k = 0; k < Gia_ManCiNum(p); k++  {
            Vec_IntPush( vPat, Abc_Random(0) & 1 );
        }
        vRes = Gia_SimQualityOne( p, vPat, 0 );
        assert(Vec_IntSize(vRes) == Gia_ManObjNum(p), "Vec_IntSize(vRes) == Gia_ManObjNum(p)");
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vRes)) && (((Value) = Vec_IntEntry(vRes, k)), 1); k++  {
            Vec_IntAddToEntry( vTotal, k, Value );
        }
        Vec_IntFree( vRes );
        Vec_IntFree( vPat );
    }
    //Vec_IntPrint( vTotal );
    return vTotal;
}   
func Gia_SimComputeScore(p *Gia_Man_t, vTotal *Vec_Int_t, vThis *Vec_Int_t) float64 {
    var TotalScore float64;
    var i, Total, This int32; 
    assert(Vec_IntSize(vTotal) == Vec_IntSize(vThis), "Vec_IntSize(vTotal) == Vec_IntSize(vThis)");
    // inlining of macro Vec_IntForEachEntryTwo
    for  i = 0; (i < Vec_IntSize(vTotal)) && (((Total) = Vec_IntEntry(vTotal, i)), 1) && (((This) = Vec_IntEntry(vThis, i)), 1); i++  {
        if Total == 0 { 
            Total = 1; }
        TotalScore += 1000.0*This/Total;
    }
    if TotalScore == 0 {
        return 1.0;
    } else {
        return TotalScore/Gia_ManAndNum(p);
    }
}
func Gia_SimQualityPatternsMax(p *Gia_Man_t, vPat *Vec_Int_t, Iter int32, fVerbose int32, vStats *Vec_Int_t) int32 {
    var k int32;
    var MaxIn int32 = -1;
    var vTries *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vRes *Vec_Int_t  = Gia_SimQualityOne( p, vPat, 0 );
    var Value float64;
    var InitValue float64;
    var MaxValue float64 = InitValue = Gia_SimComputeScore( p, vStats, vRes );
    Vec_IntFree( vRes );

    if fVerbose != 0 {
        fmt.Printf("Iter %5d : Init = %6.3f  ", Iter, InitValue); }

    for  k = 0; k < Gia_ManCiNum(p); k++  {
        Vec_IntArray(vPat)[k] ^= 1;
        //Vec_IntPrint( vPat );

        vRes = Gia_SimQualityOne( p, vPat, 0 );
        Value = Gia_SimComputeScore( p, vStats, vRes );
        if MaxValue <= Value {
            if MaxValue < Value {
                Vec_IntClear( vTries ); }
            Vec_IntPush( vTries, k );
            MaxValue = Value;
            MaxIn = k;
        }
        Vec_IntFree( vRes );

        Vec_IntArray(vPat)[k] ^= 1;
    }
    if Vec_IntSize(vTries) {
        MaxIn = Vec_IntEntry( vTries, rand()%Vec_IntSize(vTries) );
    } else {
        MaxIn = -1;
    }
    if fVerbose != 0 {
        fmt.Printf("Final = %6.3f  Ratio = %4.2f  Tries = %5d  ", MaxValue, MaxValue/InitValue, Vec_IntSize(vTries));
        fmt.Printf("Choosing %5d\r", MaxIn);
    }
    Vec_IntFree( vTries );
    return MaxIn;
}
func Gia_ManPatCollectOne(p *Gia_Man_t, vPatterns *Vec_Wrd_t, n int32, nWords int32) *Vec_Int_t {
    var vPat *Vec_Int_t  = Vec_IntAlloc( Gia_ManCiNum(p) ); var k int32;
    for  k = 0; k < Gia_ManCiNum(p); k++  {
        Vec_IntPush( vPat, Abc_TtGetBit( Vec_WrdEntryP(vPatterns, k*nWords), n ) );
    }
    return vPat;
}
func Gia_ManPatUpdateOne(p *Gia_Man_t, vPatterns *Vec_Wrd_t, n int32, nWords int32, vPat *Vec_Int_t) {
    var k, Value int32;
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vPat)) && (((Value) = Vec_IntEntry(vPat, k)), 1); k++  {
        var pSim *word  = Vec_WrdEntryP( vPatterns, k*nWords );
        if Abc_TtGetBit(pSim, n) != Value {
            Abc_TtXorBit( pSim, n ); }
    }
}
func Gia_ManPatDistImprove(p *Gia_Man_t, fVerbose int32) {
    var n int32;
    var k int32;
    var nWords int32 = Vec_WrdSize(p.vSimsPi) / Gia_ManCiNum(p);
    var InitValue, InitTotal, FinalValue, FinalTotal float64;
    var vPat *Vec_Int_t;
    var vRes *Vec_Int_t;
    var vStats *Vec_Int_t = Gia_SimGenerateStats( p );
    var vPatterns *Vec_Wrd_t  = p.vSimsPi; p.vSimsPi = nil;
    Abc_Random(1);
    for  n = 0; n < 64*nWords; n++  {
        var clk abctime  = Abc_Clock();
//        if ( n == 32 )
//            break;

        vPat = Gia_ManPatCollectOne( p, vPatterns, n, nWords );
        vRes = Gia_SimQualityOne( p, vPat, 0 );
        InitValue = Gia_SimComputeScore(p, vStats, vRes);
        InitTotal += InitValue;
        Vec_IntFree( vRes );

        for  k = 0; k < 100; k++  {
            var MaxIn int32  = Gia_SimQualityPatternsMax( p, vPat, k, fVerbose, vStats );
            if MaxIn == -1 {
                break; }
            assert(MaxIn >= 0 && MaxIn < Gia_ManCiNum(p), "MaxIn >= 0 && MaxIn < Gia_ManCiNum(p)");
            Vec_IntArray(vPat)[MaxIn] ^= 1;            
        }
        //Vec_IntPrint( vPat );

        vRes = Gia_SimQualityOne( p, vPat, 0 );
        FinalValue = Gia_SimComputeScore(p, vStats, vRes);
        FinalTotal += FinalValue;
        Vec_IntFree( vRes );

        if fVerbose != 0 {
            fmt.Printf("Pat %5d : Tries = %5d  InitValue = %6.3f  FinalValue = %6.3f  Ratio = %4.2f  ", n, k, InitValue, FinalValue, FinalValue/InitValue);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }

        Gia_ManPatUpdateOne( p, vPatterns, n, nWords, vPat );
        Vec_IntFree( vPat );
    }
    Vec_IntFree( vStats );
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    fmt.Printf("Improved %d patterns with average init value %.2f and average final value %.2f.\n", 64*nWords, 1.0*InitTotal/(64*nWords), 1.0*FinalTotal/(64*nWords));
    p.vSimsPi = vPatterns;
}

/**Function*************************************************************

  Synopsis    [Improving quality of simulation patterns.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SimCollectRare(p *Gia_Man_t, vPatterns *Vec_Wrd_t, RareLimit int32) *Vec_Int_t {
    var vRareCounts *Vec_Int_t  = Vec_IntAlloc( 100 );  // (node, rare_count) pairs
    var Id int32;
    var nWords int32 = Vec_WrdSize(vPatterns) / Gia_ManCiNum(p);
    var TotalBits int32 = 64*nWords;
    var vSims *Vec_Wrd_t;
    var vTemp *Vec_Wrd_t = p.vSimsPi;
    assert(Vec_WrdSize(vPatterns) % Gia_ManCiNum(p) == 0, "Vec_WrdSize(vPatterns) % Gia_ManCiNum(p) == 0");
    p.vSimsPi = vPatterns;
    vSims = Gia_ManSimPatSim( p );
    p.vSimsPi = vTemp;
    // inlining of macro Gia_ManForEachAndId
    for  Id = 0; (Id < p.nObjs); Id++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p, Id)) {} else {
        var pSim *word    = Vec_WrdEntryP( vSims, Id*nWords );
        var Count int32      = Abc_TtCountOnesVec( pSim, nWords );
        var fRareOne int32   = Count < TotalBits/2; // fRareOne is 1 if rare value is 1
        var CountRare int32;
        if fRareOne != 0 {
            CountRare = Count;
        } else {
            CountRare = TotalBits - Count;
        }
        assert(CountRare <= TotalBits/2, "CountRare <= TotalBits/2");
        if CountRare <= RareLimit {
            Vec_IntPushTwo( vRareCounts, Abc_Var2Lit(Id, fRareOne), CountRare ); }
    }
    }
    Vec_WrdFree( vSims );
    return vRareCounts;
}
func Gia_SimQualityImpact(p *Gia_Man_t, vPat *Vec_Int_t, vRareCounts *Vec_Int_t) *Vec_Flt_t {
    var vQuoIncs *Vec_Flt_t  = Vec_FltStart( Gia_ManCiNum(p) );
    var nWordsNew int32  = Abc_Bit6WordNum( 1+Gia_ManCiNum(p) );
    var vSimsPiNew *Vec_Wrd_t  = Vec_WrdStart( Gia_ManCiNum(p) * nWordsNew );
    var vTemp, vSims *Vec_Wrd_t;
    var i, k, Value, RareLit, RareCount int32;
    assert(Vec_IntSize(vPat) == Gia_ManCiNum(p), "Vec_IntSize(vPat) == Gia_ManCiNum(p)");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPat)) && (((Value) = Vec_IntEntry(vPat, i)), 1); i++  {
        var pSim *word  = Vec_WrdEntryP( vSimsPiNew, i*nWordsNew );
        if Value != 0 {
            Abc_TtFill( pSim, nWordsNew ); }
        Abc_TtXorBit( pSim, i+1 );
    }
    vTemp = p.vSimsPi;
    p.vSimsPi = vSimsPiNew;
    vSims = Gia_ManSimPatSim( p );
    p.vSimsPi = vTemp;
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vRareCounts)) && (((RareLit) = Vec_IntEntry(vRareCounts, i)), 1) && (((RareCount) = Vec_IntEntry(vRareCounts, i+1)), 1); i += 2  {
        var Incrm float32  = float32(1.0)/(RareCount+1);
        var RareObj int32  = Abc_Lit2Var(RareLit);
        var RareVal int32  = Abc_LitIsCompl(RareLit);
        var pSim *word  = Vec_WrdEntryP( vSims, RareObj*nWordsNew );
        var OrigVal int32  = pSim[0] & 1;
        if OrigVal != 0 {
            Abc_TtNot( pSim, nWordsNew ); }
        for  k = 0; k < Gia_ManCiNum(p); k++  {
            if Abc_TtGetBit(pSim, k+1) { // value changed
                Vec_FltAddToEntry( vQuoIncs, k, OrigVal != RareVal ? Incrm : -Incrm ); }
        }
    }
    Vec_WrdFree( vSims );
    Vec_WrdFree( vSimsPiNew );
    return vQuoIncs;
}
func Gia_SimCollectBest(vQuo *Vec_Flt_t) *Vec_Int_t {
    var vRes *Vec_Int_t; var i int32;
    var Value float32;
    var ValueMax float32 = Vec_FltFindMax( vQuo );
    if ValueMax <= 0 {
        return nil; }
    vRes = Vec_IntAlloc( 100 );  // variables with max quo
    // inlining of macro Vec_FltForEachEntry
    for  i = 0; (i < Vec_FltSize(vQuo)) && (((Value) = Vec_FltEntry(vQuo, i)), 1); i++  {
        if Value == ValueMax {
            Vec_IntPush( vRes, i ); }
    }
    return vRes;
}
func Gia_ManPatGetQuo(p *Gia_Man_t, vRareCounts *Vec_Int_t, vSims *Vec_Wrd_t, n int32, nWords int32) float32 {
    var Quality float32;
    var RareLit, RareCount, i int32; 
    assert(Vec_WrdSize(vSims) == Gia_ManObjNum(p), "Vec_WrdSize(vSims) == Gia_ManObjNum(p)");
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vRareCounts)) && (((RareLit) = Vec_IntEntry(vRareCounts, i)), 1) && (((RareCount) = Vec_IntEntry(vRareCounts, i+1)), 1); i += 2  {
        var Incrm float32  = float32(1.0)/(RareCount+1);
        var RareObj int32  = Abc_Lit2Var(RareLit);
        var RareVal int32  = Abc_LitIsCompl(RareLit);
        var pSim *word  = Vec_WrdEntryP( vSims, RareObj*nWords );
        if Abc_TtGetBit(pSim, n) == RareVal {
            Quality += Incrm; }
    }
    return Quality;
}
func Gia_ManPatGetTotalQuo(p *Gia_Man_t, RareLimit int32, vPatterns *Vec_Wrd_t, nWords int32) float32 {
    var Total float32; var n int32;
    var vRareCounts *Vec_Int_t  = Gia_SimCollectRare( p, vPatterns, RareLimit );
    var vSims *Vec_Wrd_t;
    var vTemp *Vec_Wrd_t = p.vSimsPi;
    p.vSimsPi = vPatterns;
    vSims = Gia_ManSimPatSim( p );
    p.vSimsPi = vTemp;
    for  n = 0; n < 64*nWords; n++  {
        Total += Gia_ManPatGetQuo( p, vRareCounts, vSims, n, nWords );
    }
    Vec_IntFree( vRareCounts );
    Vec_WrdFree( vSims );
    return Total;
}
func Gia_ManPatGetOneQuo(p *Gia_Man_t, RareLimit int32, vPatterns *Vec_Wrd_t, nWords int32, n int32) float32 {
    var Total float32; 
    var vRareCounts *Vec_Int_t  = Gia_SimCollectRare( p, vPatterns, RareLimit );
    var vSims *Vec_Wrd_t;
    var vTemp *Vec_Wrd_t = p.vSimsPi;
    p.vSimsPi = vPatterns;
    vSims = Gia_ManSimPatSim( p );
    p.vSimsPi = vTemp;
    Total += Gia_ManPatGetQuo( p, vRareCounts, vSims, n, nWords );
    Vec_IntFree( vRareCounts );
    Vec_WrdFree( vSims );
    return Total;
}
func Gia_ManPatRareImprove(p *Gia_Man_t, RareLimit int32, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var FinalTotal, InitTotal float32;
    var n int32;
    var nRares int32;
    var nChanges int32;
    var nWords int32 = Vec_WrdSize(p.vSimsPi) / Gia_ManCiNum(p);
    var vPatterns *Vec_Wrd_t  = p.vSimsPi; p.vSimsPi = nil;
    InitTotal = Gia_ManPatGetTotalQuo( p, RareLimit, vPatterns, nWords );
    for  n = 0; n < 64*nWords; n++  {
        var clk abctime  = Abc_Clock();
        var vRareCounts *Vec_Int_t  = Gia_SimCollectRare( p, vPatterns, RareLimit );
        var vPat *Vec_Int_t         = Gia_ManPatCollectOne( p, vPatterns, n, nWords );
        var vQuoIncs *Vec_Flt_t     = Gia_SimQualityImpact( p, vPat, vRareCounts );
        var vBest *Vec_Int_t        = Gia_SimCollectBest( vQuoIncs );
        if fVerbose != 0 {
            var PatQuo float32  = Gia_ManPatGetOneQuo( p, RareLimit, vPatterns, nWords, n );
            fmt.Printf("Pat %5d : Rare = %4d  Cands = %3d  Value = %8.3f  Change = %8.3f  ", n, Vec_IntSize(vRareCounts)/2, vBest ? Vec_IntSize(vBest) : 0, PatQuo, vBest ? Vec_FltEntry(vQuoIncs, Vec_IntEntry(vBest,0)) : 0);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }
        if vBest != nil {
            var VarBest int32  = Vec_IntEntry( vBest, rand()%Vec_IntSize(vBest) );
            Abc_TtXorBit( Vec_WrdEntryP(vPatterns, VarBest*nWords), n );
            nChanges++;
        }
        nRares = Vec_IntSize(vRareCounts)/2;
        Vec_IntFree( vRareCounts );
        Vec_IntFree( vPat );
        Vec_FltFree( vQuoIncs );
        Vec_IntFreeP( &vBest );
    }
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    FinalTotal = Gia_ManPatGetTotalQuo( p, RareLimit, vPatterns, nWords );
    p.vSimsPi = vPatterns;

    fmt.Printf("Improved %d out of %d patterns using %d rare nodes: %.2f -> %.2f.  ", nChanges, 64*nWords, nRares, InitTotal, FinalTotal);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    [Trying vectorized simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimTest(pGia *Gia_Man_t) {
    var n int32;
    var nWords int32 = 4;
    var vSim1, vSim2 *Vec_Wrd_t;
    var vSim0 *Vec_Wrd_t  = Vec_WrdStartRandom( Gia_ManCiNum(pGia) * nWords );
    var clk abctime  = Abc_Clock();

    pGia.vSimsPi = vSim0;
    for  n = 0; n < 20; n++  {
        vSim1 = Gia_ManSimPatSim( pGia );
        Vec_WrdFree( vSim1 );
    }
    Abc_PrintTime( 1, "Time1", Abc_Clock() - clk );

    clk = Abc_Clock();
    for  n = 0; n < 20; n++  {
        vSim2 = Gia_ManSimPatSim2( pGia );
        Vec_WrdFree( vSim2 );
    }
    Abc_PrintTime( 1, "Time2", Abc_Clock() - clk );

    pGia.vSimsPi = nil;
    Vec_WrdFree( vSim0 );
}

/**Function*************************************************************

  Synopsis    [Trying compiled simulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimGen(pGia *Gia_Man_t) {
    var nWords int32  = 4;
    var pObj *Gia_Obj_t;
    var vSim0 *Vec_Wrd_t  = Vec_WrdStartRandom( Gia_ManCiNum(pGia) * nWords );
    var pFile *os.File  = os.OpenFile("comp_sim.c", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    var i, k, Id int32;
    fmt.Fprintf(pFile, "#include <stdio.h>\n");
    fmt.Fprintf(pFile, "#include <stdlib.h>\n");
    fmt.Fprintf(pFile, "#include <time.h>\n");
    fmt.Fprintf(pFile, "int main()\n");
    fmt.Fprintf(pFile, "{\n");
    fmt.Fprintf(pFile, "  clock_t clkThis = clock();\n");
    fmt.Fprintf(pFile, "  unsigned long Res = 0;\n");
    fmt.Fprintf(pFile, "  int i;\n");
    fmt.Fprintf(pFile, "  srand(time(NULL));\n");
    fmt.Fprintf(pFile, "  for ( i = 0; i < 2000; i++ )\n");
    fmt.Fprintf(pFile, "  {\n");
    for  k = 0; k < nWords; k++  {
        fmt.Fprintf(pFile, "  unsigned long s%07d_%d = 0x%08x%08x;\n", 0, k, 0, 0);
    }
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((Id) = Gia_ObjId(pGia, Gia_ManCi(pGia, i))); i++  {
        //word * pSim = Vec_WrdEntryP(vSim0, i*nWords);
        //unsigned * pSimU = (unsigned *)pSim;
        for  k = 0; k < nWords; k++  {
            fmt.Fprintf(pFile, "  unsigned long s%07d_%d = ((unsigned long)rand() << 48) | ((unsigned long)rand() << 32) | ((unsigned long)rand() << 16) | (unsigned long)rand();\n", Id, k);
        }
    }
    // inlining of macro Gia_ManForEachAnd
    for  Id = 0; (Id < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, Id)); Id++  {      if !Gia_ObjIsAnd(pObj) {} else {
        for  k = 0; k < nWords; k++  {
            fmt.Fprintf(pFile, "  unsigned long s%07d_%d = %cs%07d_%d & %cs%07d_%d;\n", Id, k, Gia_ObjFaninC0(pObj) ? '~' : ' ', Gia_ObjFaninId0(pObj, Id), k, Gia_ObjFaninC1(pObj) ? ' ' : '~', Gia_ObjFaninId1(pObj, Id), k);
        }
    }
    }
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((Id) = Gia_ObjId(pGia, Gia_ManCo(pGia, i))); i++  {
        pObj = Gia_ManObj(pGia, Id);
        for  k = 0; k < nWords; k++  {
            fmt.Fprintf(pFile, "  Res ^= %cs%07d_%d;\n", Gia_ObjFaninC0(pObj) ? '~' : ' ', Gia_ObjFaninId0(pObj, Id), k);
        }
    }
    Vec_WrdFree( vSim0 );
    fmt.Fprintf(pFile, "  }\n");
    fmt.Fprintf(pFile, "  printf( \"Res = 0x%%08x    \", (unsigned)Res );\n");
    fmt.Fprintf(pFile, "  printf( \"Time = %%6.2f sec\\n\", (float)(clock() - clkThis)/CLOCKS_PER_SEC );\n");
    fmt.Fprintf(pFile, "  return 1;\n");
    fmt.Fprintf(pFile, "}\n");
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


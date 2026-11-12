package cec

/**CFile****************************************************************

  FileName    [cecClass.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Equivalence class refinement.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecClass.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Cec_ObjSim(p *Cec_ManSim_t, Id int32) *uint32 { return p.pMems + p.pSimInfo[Id] + 1;   }
func Cec_ObjSetSim(p *Cec_ManSim_t, Id int32, n int32) { p.pSimInfo[Id] = n;                     }

func Cec_MemUsage(p *Cec_ManSim_t) float32 { return 1.0*p.nMemsMax*(p.pPars.nWords+1)/(1<<20);   }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Compares simulation info of one node with constant 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimCompareConst(p *uint32, nWords int32) int32 {
    var w int32;
    if p[0] & 1 {
        for  w = 0; w < nWords; w++  {
            if p[w] != ~0 {
                return 0; }
        }
        return 1;
    
    } else {
        for  w = 0; w < nWords; w++  {
            if p[w] != 0 {
                return 0; }
        }
        return 1;
    }
}

/**Function*************************************************************

  Synopsis    [Compares simulation info of two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimCompareEqual(p0 *uint32, p1 *uint32, nWords int32) int32 {
    var w int32;
    if (p0[0] & 1) == (p1[0] & 1) {
        for  w = 0; w < nWords; w++  {
            if p0[w] != p1[w] {
                return 0; }
        }
        return 1;
    
    } else {
        for  w = 0; w < nWords; w++  {
            if p0[w] != ~p1[w] {
                return 0; }
        }
        return 1;
    }
}

/**Function*************************************************************

  Synopsis    [Returns the number of the first non-equal bit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimCompareConstFirstBit(p *uint32, nWords int32) int32 {
    var w int32;
    if p[0] & 1 {
        for  w = 0; w < nWords; w++  {
            if p[w] != ~0 {
                return 32*w + Gia_WordFindFirstBit( ~p[w] ); }
        }
        return -1;
    
    } else {
        for  w = 0; w < nWords; w++  {
            if p[w] != 0 {
                return 32*w + Gia_WordFindFirstBit( p[w] ); }
        }
        return -1;
    }
}

/**Function*************************************************************

  Synopsis    [Compares simulation info of two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimCompareEqualFirstBit(p0 *uint32, p1 *uint32, nWords int32) int32 {
    var w int32;
    if (p0[0] & 1) == (p1[0] & 1) {
        for  w = 0; w < nWords; w++  {
            if p0[w] != p1[w] {
                return 32*w + Gia_WordFindFirstBit( p0[w] ^ p1[w] ); }
        }
        return -1;
    
    } else {
        for  w = 0; w < nWords; w++  {
            if p0[w] != ~p1[w] {
                return 32*w + Gia_WordFindFirstBit( p0[w] ^ ~p1[w] ); }
        }
        return -1;
    }
}

/**Function*************************************************************

  Synopsis    [Returns the number of the first non-equal bit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimCompareConstScore(p *uint32, nWords int32, pScores *int32) {
    var w, b int32;
    if p[0] & 1 {
        for  w = 0; w < nWords; w++  {
            if p[w] != ~0 {
                for  b = 0; b < 32; b++  {
                    if ((~p[w]) >> b ) & 1 {
                        pScores[32*w + b]++; }
                } }
        }
    
    } else {
        for  w = 0; w < nWords; w++  {
            if p[w] != 0 {
                for  b = 0; b < 32; b++  {
                    if ((p[w]) >> b ) & 1 {
                        pScores[32*w + b]++; }
                } }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Compares simulation info of two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimCompareEqualScore(p0 *uint32, p1 *uint32, nWords int32, pScores *int32) {
    var w, b int32;
    if (p0[0] & 1) == (p1[0] & 1) {
        for  w = 0; w < nWords; w++  {
            if p0[w] != p1[w] {
                for  b = 0; b < 32; b++  {
                    if ((p0[w] ^ p1[w]) >> b ) & 1 {
                        pScores[32*w + b]++; }
                } }
        }
    
    } else {
        for  w = 0; w < nWords; w++  {
            if p0[w] != ~p1[w] {
                for  b = 0; b < 32; b++  {
                    if ((p0[w] ^ ~p1[w]) >> b ) & 1 {
                        pScores[32*w + b]++; }
                } }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Creates equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimClassCreate(p *Gia_Man_t, vClass *Vec_Int_t) {
    var Repr int32 = GIA_VOID;
    var EntPrev int32 = -1;
    var Ent int32;
    var i int32;
    assert(Vec_IntSize(vClass) > 0, "Vec_IntSize(vClass) > 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClass)) && (((Ent) = Vec_IntEntry(vClass, i)), 1); i++  {
        if i == 0 {
            Repr = Ent;
            Gia_ObjSetRepr( p, Ent, GIA_VOID );
            EntPrev = Ent;
        
        } else {
            assert(Repr < Ent, "Repr < Ent");
            Gia_ObjSetRepr( p, Ent, Repr );
            Gia_ObjSetNext( p, EntPrev, Ent );
            EntPrev = Ent;
        }
    }
    Gia_ObjSetNext( p, EntPrev, 0 );
}

/**Function*************************************************************

  Synopsis    [Refines one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimClassRefineOne(p *Cec_ManSim_t, i int32) int32 {
    var pSim0, pSim1 *uint32;
    var Ent int32;
    Vec_IntClear( p.vClassOld );
    Vec_IntClear( p.vClassNew );
    Vec_IntPush( p.vClassOld, i );
    pSim0 = Cec_ObjSim(p, i);
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p.pAig, i), "Gia_ObjIsHead(p.pAig, i)"), Ent = Gia_ObjNext(p.pAig, i); Ent > 0; Ent = Gia_ObjNext(p.pAig, Ent)  {
        pSim1 = Cec_ObjSim(p, Ent);
        if Cec_ManSimCompareEqual( pSim0, pSim1, p.nWords ) != 0 {
            Vec_IntPush( p.vClassOld, Ent );
        } else {
            Vec_IntPush( p.vClassNew, Ent );
            if p.pBestState {
                Cec_ManSimCompareEqualScore( pSim0, pSim1, p.nWords, p.pScores ); }
        }
    }
    if Vec_IntSize( p.vClassNew ) == 0 {
        return 0; }
    Cec_ManSimClassCreate( p.pAig, p.vClassOld );
    Cec_ManSimClassCreate( p.pAig, p.vClassNew );
    if Vec_IntSize(p.vClassNew) > 1 {
        return 1 + Cec_ManSimClassRefineOne( p, Vec_IntEntry(p.vClassNew,0) ); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Refines one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimClassRemoveOne(p *Cec_ManSim_t, i int32) int32 {
    var iRepr, Ent int32;
    if Gia_ObjIsConst(p.pAig, i) {
        Gia_ObjSetRepr( p.pAig, i, GIA_VOID );
        return 1;
    }
    if !Gia_ObjIsClass(p.pAig, i) {
        return 0; }
    assert(Gia_ObjIsClass(p.pAig, i), "Gia_ObjIsClass(p.pAig, i)");
    iRepr = Gia_ObjRepr( p.pAig, i );
    if iRepr == GIA_VOID {
        iRepr = i; }
    // collect nodes
    Vec_IntClear( p.vClassOld );
    Vec_IntClear( p.vClassNew );
    // inlining of macro Gia_ClassForEachObj
    for  assert(Gia_ObjIsHead(p.pAig, iRepr) && iRepr, "Gia_ObjIsHead(p.pAig, iRepr) && iRepr"), Ent = iRepr; Ent > 0; Ent = Gia_ObjNext(p.pAig, Ent)  {
        if Ent == i {
            Vec_IntPush( p.vClassNew, Ent );
        } else {
            Vec_IntPush( p.vClassOld, Ent ); }
    }
    assert(Vec_IntSize( p.vClassNew ) == 1, "Vec_IntSize( p.vClassNew ) == 1");
    Cec_ManSimClassCreate( p.pAig, p.vClassOld );
    Cec_ManSimClassCreate( p.pAig, p.vClassNew );
    assert(!Gia_ObjIsClass(p.pAig, i), "!Gia_ObjIsClass(p.pAig, i)");
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes hash key of the simuation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Cec_ManSimHashKey_s_Primes [16]int32 = [16]int32{ 
        1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 
        4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147 };
func Cec_ManSimHashKey(pSim *uint32, nWords int32, nTableSize int32) int32 {
    
    var uHash uint32;
    var i int32;
    if pSim[0] & 1 {
        for  i = 0; i < nWords; i++  {
            uHash ^= ~pSim[i] * spStatic_Cec_ManSimHashKey_s_Primes[i & 0xf];
        }
    } else {
        for  i = 0; i < nWords; i++  {
            uHash ^= pSim[i] * spStatic_Cec_ManSimHashKey_s_Primes[i & 0xf];
        } }
    return int32((uHash % nTableSize));

}

/**Function*************************************************************

  Synopsis    [Resets pointers to the simulation memory.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimMemRelink(p *Cec_ManSim_t) {
    var pPlace *uint32;
    var Ent uint32;
    pPlace = (unsigned *)&p.MemFree;
    for  Ent = p.nMems * (p.nWords + 1); 
          Ent + p.nWords + 1 < uint32(p.nWordsAlloc); 
          Ent += p.nWords + 1  {
        *pPlace = Ent;
        pPlace = p.pMems + Ent;
    }
    *pPlace = 0;
    p.nWordsOld = p.nWords;
}

/**Function*************************************************************

  Synopsis    [References simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimSimRef(p *Cec_ManSim_t, i int32) *uint32 {
    var pSim *uint32;
    assert(p.pSimInfo[i] == 0, "p.pSimInfo[i] == 0");
    if p.MemFree == 0 {
        if p.nWordsAlloc == 0 {
            assert(p.pMems == nil, "p.pMems == nil");
            p.nWordsAlloc = (1<<17); // -> 1Mb
            p.nMems = 1;
        }
        p.nWordsAlloc *= 2;
        // inlining of macro ABC_REALLOC
        
        if p.pMems {
            p.pMems = ((unsigned *) realloc((char *)(p.pMems), sizeof(unsigned) * (size_t)(p.nWordsAlloc)));
        } else {
            p.pMems = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nWordsAlloc)));
        }
        Cec_ManSimMemRelink( p );
    }
    p.pSimInfo[i] = p.MemFree;
    pSim = p.pMems + p.MemFree;
    p.MemFree = pSim[0];
    pSim[0] = Gia_ObjValue( Gia_ManObj(p.pAig, i) );
    p.nMems++;
    if p.nMemsMax < p.nMems {
        p.nMemsMax = p.nMems; }
    return pSim;
}

/**Function*************************************************************

  Synopsis    [Dereferences simulaton info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimSimDeref(p *Cec_ManSim_t, i int32) *uint32 {
    var pSim *uint32;
    assert(p.pSimInfo[i] > 0, "p.pSimInfo[i] > 0");
    pSim = p.pMems + p.pSimInfo[i];
    if --pSim[0] == 0 {
        pSim[0] = p.MemFree;
        p.MemFree = p.pSimInfo[i];
        p.pSimInfo[i] = 0;
        p.nMems--;
    }
    return pSim;
}

/**Function*************************************************************

  Synopsis    [Refines nodes belonging to candidate constant class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimProcessRefined(p *Cec_ManSim_t, vRefined *Vec_Int_t) {
    var pSim *uint32;
    var pTable *int32;
    var nTableSize int32;
    var i int32;
    var k int32;
    var Key int32;
    if Vec_IntSize(vRefined) == 0 {
        return; }
    nTableSize = Abc_PrimeCudd( 100 + Vec_IntSize(vRefined) / 3 );
    // inlining of macro ABC_CALLOC
    pTable = ((int *) calloc((size_t)(nTableSize), sizeof(int)));
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vRefined)) && (((i) = Vec_IntEntry(vRefined, k)), 1); k++  {
        pSim = Cec_ObjSim( p, i );
        assert(!(Cec_ManSimCompareConst( pSim, p.nWords ) != 0), "!Cec_ManSimCompareConst( pSim, p.nWords )");
        Key = Cec_ManSimHashKey( pSim, p.nWords, nTableSize );
        if pTable[Key] == 0 {
            assert(Gia_ObjRepr(p.pAig, i) == 0, "Gia_ObjRepr(p.pAig, i) == 0");
            assert(Gia_ObjNext(p.pAig, i) == 0, "Gia_ObjNext(p.pAig, i) == 0");
            Gia_ObjSetRepr( p.pAig, i, GIA_VOID );
        
        } else {
            Gia_ObjSetNext( p.pAig, pTable[Key], i );
            Gia_ObjSetRepr( p.pAig, i, Gia_ObjRepr(p.pAig, pTable[Key]) );
            if Gia_ObjRepr(p.pAig, i) == GIA_VOID {
                Gia_ObjSetRepr( p.pAig, i, pTable[Key] ); }
            assert(Gia_ObjRepr(p.pAig, i) > 0, "Gia_ObjRepr(p.pAig, i) > 0");
        }
        pTable[Key] = i;
    }
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vRefined)) && (((i) = Vec_IntEntry(vRefined, k)), 1); k++  {
        if Gia_ObjIsHead( p.pAig, i ) {
            Cec_ManSimClassRefineOne( p, i ); }
    }
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vRefined)) && (((i) = Vec_IntEntry(vRefined, k)), 1); k++  {
        Cec_ManSimSimDeref( p, i );
    }
    // inlining of macro ABC_FREE
    ((pTable) ? (free((char *) (pTable)), (pTable) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Saves the input pattern with the given number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimSavePattern(p *Cec_ManSim_t, iPat int32) {
    var pInfo *uint32;
    var i int32;
    assert(p.pCexComb == nil, "p.pCexComb == nil");
    assert(iPat >= 0 && iPat < 32 * p.nWords, "iPat >= 0 && iPat < 32 * p.nWords");
    // inlining of macro ABC_CALLOC
    p.pCexComb = (Abc_Cex_t *)((char *) calloc((size_t)(sizeof(Abc_Cex_t) + sizeof(unsigned) * Abc_BitWordNum(Gia_ManCiNum(p.pAig))), sizeof(char)));
    p.pCexComb.iPo = p.iOut;
    p.pCexComb.nPis = Gia_ManCiNum(p.pAig);
    p.pCexComb.nBits = Gia_ManCiNum(p.pAig);
    for  i = 0; i < Gia_ManCiNum(p.pAig); i++  {
        pInfo = (unsigned *)Vec_PtrEntry( p.vCiSimInfo, i );
        if Abc_InfoHasBit( pInfo, iPat ) {
            Abc_InfoSetBit( p.pCexComb.pData, i ); }
    }
}

/**Function*************************************************************

  Synopsis    [Find the best pattern using the scores.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimFindBestPattern(p *Cec_ManSim_t) { 
    var pInfo *uint32;
    var i int32;
    var ScoreBest int32;
    var iPatBest int32 = 1; // set the first pattern
    // find the best pattern
    for  i = 0; i < 32 * p.nWords; i++  {
        if ScoreBest < p.pScores[i] {
            ScoreBest = p.pScores[i];
            iPatBest = i;
        }
    }
    // compare this with the available patterns - and save
    if p.pBestState.iPo <= ScoreBest {
        assert(p.pBestState.nRegs == Gia_ManRegNum(p.pAig), "p.pBestState.nRegs == Gia_ManRegNum(p.pAig)");
        for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
            pInfo = (unsigned *)Vec_PtrEntry( p.vCiSimInfo, Gia_ManPiNum(p.pAig) + i );
            if Abc_InfoHasBit(p.pBestState.pData, i) != Abc_InfoHasBit(pInfo, iPatBest) {
                Abc_InfoXorBit( p.pBestState.pData, i ); }
        }
        p.pBestState.iPo = ScoreBest;
    }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if computation should stop.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimAnalyzeOutputs(p *Cec_ManSim_t) int32 {
    var pInfo, pInfo2 *uint32;
    var i int32;
    if !p.pPars.fCheckMiter {
        return 0; }
    assert(p.vCoSimInfo != nil, "p.vCoSimInfo != nil");
    // compare outputs with 0
    if p.pPars.fDualOut {
        assert((Gia_ManPoNum(p.pAig) & 1) == 0, "(Gia_ManPoNum(p.pAig) & 1) == 0");
        for  i = 0; i < Gia_ManPoNum(p.pAig); i++  {
            pInfo  = (unsigned *)Vec_PtrEntry( p.vCoSimInfo, i );
            pInfo2 = (unsigned *)Vec_PtrEntry( p.vCoSimInfo, ++i );
            if !(Cec_ManSimCompareEqual( pInfo, pInfo2, p.nWords ) != 0) {
                if p.iOut == -1 {
                    p.iOut = i/2;
                    Cec_ManSimSavePattern( p, Cec_ManSimCompareEqualFirstBit(pInfo, pInfo2, p.nWords) );
                }
                if p.pCexes == nil {
                    // inlining of macro ABC_CALLOC
                    p.pCexes = ((void * *) calloc((size_t)(Gia_ManPoNum(p.pAig)/2), sizeof(void *))); }
                if p.pCexes[i/2] == nil {
                    p.nOuts++;
                    p.pCexes[i/2] = (void *)1;
                }
            }
        }
    
    } else {
        for  i = 0; i < Gia_ManPoNum(p.pAig); i++  {
            pInfo = (unsigned *)Vec_PtrEntry( p.vCoSimInfo, i );
            if !(Cec_ManSimCompareConst( pInfo, p.nWords ) != 0) {
                if p.iOut == -1 {
                    p.iOut = i;
                    Cec_ManSimSavePattern( p, Cec_ManSimCompareConstFirstBit(pInfo, p.nWords) );
                }
                if p.pCexes == nil {
                    // inlining of macro ABC_CALLOC
                    p.pCexes = ((void * *) calloc((size_t)(Gia_ManPoNum(p.pAig)), sizeof(void *))); }
                if p.pCexes[i] == nil {
                    p.nOuts++;
                    p.pCexes[i] = (void *)1;
                }
            }
        }
    }
    return p.pCexes != nil;
}

/**Function*************************************************************

  Synopsis    [Simulates one round.]

  Description [Returns the number of PO entry if failed; 0 otherwise.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimSimulateRound(p *Cec_ManSim_t, vInfoCis *Vec_Ptr_t, vInfoCos *Vec_Ptr_t) int32 {
    var pObj *Gia_Obj_t;
    var pRes0, pRes1, pRes *uint32;
    var i, k, w, Ent, iCiId, iCoId int32;
    // prepare internal storage
    if p.nWordsOld != p.nWords {
        Cec_ManSimMemRelink( p ); }
    p.nMemsMax = 0;
    // allocate score counters
    // inlining of macro ABC_FREE
    ((p.pScores) ? (free((char *) (p.pScores)), (p.pScores) = 0) : 0);
    if p.pBestState {
        // inlining of macro ABC_CALLOC
        p.pScores = ((int *) calloc((size_t)(32 * p.nWords), sizeof(int))); }
    // simulate nodes
    Vec_IntClear( p.vRefinedC );
    if Gia_ObjValue(Gia_ManConst0(p.pAig)) {
        pRes = Cec_ManSimSimRef( p, 0 );
        for  w = 1; w <= p.nWords; w++  {
            pRes[w] = 0;
        }
    }
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.pAig.nObjs) && ((pObj) = Gia_ManObj(p.pAig, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            if Gia_ObjValue(pObj) == 0 {
                iCiId++;
                continue;
            }
            pRes = Cec_ManSimSimRef( p, i );
            if vInfoCis != nil {
                pRes0 = (unsigned *)Vec_PtrEntry( vInfoCis, iCiId++ );
                for  w = 1; w <= p.nWords; w++  {
                    pRes[w] = pRes0[w-1];
                }
            
            } else {
                for  w = 1; w <= p.nWords; w++  {
                    pRes[w] = Gia_ManRandom( 0 );
                }
            }
            // make sure the first pattern is always zero
            pRes[1] ^= (pRes[1] & 1);
            goto references;
        }
        if Gia_ObjIsCo(pObj) { // co always has non-zero 1st fanin and zero 2nd fanin
        
            pRes0 = Cec_ManSimSimDeref( p, Gia_ObjFaninId0(pObj,i) );
            if vInfoCos != nil {
                pRes = (unsigned *)Vec_PtrEntry( vInfoCos, iCoId++ );
                if Gia_ObjFaninC0(pObj) {
                    for  w = 1; w <= p.nWords; w++  {
                        pRes[w-1] = ~pRes0[w];
                    }
                } else { 
                    for  w = 1; w <= p.nWords; w++  {
                        pRes[w-1] = pRes0[w];
                    } }
            }
            continue;
        }
        assert(Gia_ObjValue(pObj), "Gia_ObjValue(pObj)");
        pRes  = Cec_ManSimSimRef( p, i );
        pRes0 = Cec_ManSimSimDeref( p, Gia_ObjFaninId0(pObj,i) );
        pRes1 = Cec_ManSimSimDeref( p, Gia_ObjFaninId1(pObj,i) );

//        Abc_Print( 1, "%d,%d  ", Gia_ObjValue( Gia_ObjFanin0(pObj) ), Gia_ObjValue( Gia_ObjFanin1(pObj) ) );

        if Gia_ObjFaninC0(pObj) {
            if Gia_ObjFaninC1(pObj) {
                for  w = 1; w <= p.nWords; w++  {
                    pRes[w] = ~(pRes0[w] | pRes1[w]);
                }
            } else {
                for  w = 1; w <= p.nWords; w++  {
                    pRes[w] = ~pRes0[w] & pRes1[w];
                } }
        
        } else {
            if Gia_ObjFaninC1(pObj) {
                for  w = 1; w <= p.nWords; w++  {
                    pRes[w] = pRes0[w] & ~pRes1[w];
                }
            } else {
                for  w = 1; w <= p.nWords; w++  {
                    pRes[w] = pRes0[w] & pRes1[w];
                } }
        }

references:
        // if this node is candidate constant, collect it
        if Gia_ObjIsConst(p.pAig, i) && !(Cec_ManSimCompareConst(pRes + 1, p.nWords) != 0) {
            pRes[0]++;
            Vec_IntPush( p.vRefinedC, i );
            if p.pBestState {
                Cec_ManSimCompareConstScore( pRes + 1, p.nWords, p.pScores ); }
        }
        // if the node belongs to a class, save it
        if Gia_ObjIsClass(p.pAig, i) {
            pRes[0]++; }
        // if this is the last node of the class, process it
        if Gia_ObjIsTail(p.pAig, i) {
            Vec_IntClear( p.vClassTemp );
            // inlining of macro Gia_ClassForEachObj
            for  assert(Gia_ObjIsHead(p.pAig, Gia_ObjRepr(p.pAig, i)) && Gia_ObjRepr(p.pAig, i), "Gia_ObjIsHead(p.pAig, Gia_ObjRepr(p.pAig, i)) && Gia_ObjRepr(p.pAig, i)"), Ent = Gia_ObjRepr(p.pAig, i); Ent > 0; Ent = Gia_ObjNext(p.pAig, Ent)  {
                Vec_IntPush( p.vClassTemp, Ent );
            }
            Cec_ManSimClassRefineOne( p, Gia_ObjRepr(p.pAig, i) );
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(p.vClassTemp)) && (((Ent) = Vec_IntEntry(p.vClassTemp, k)), 1); k++  {
                Cec_ManSimSimDeref( p, Ent );
            }
        }
    }

    if p.pPars.fConstCorr {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(p.vRefinedC)) && (((i) = Vec_IntEntry(p.vRefinedC, k)), 1); k++  {
            Gia_ObjSetRepr( p.pAig, i, GIA_VOID );
            Cec_ManSimSimDeref( p, i );
        }
        Vec_IntClear( p.vRefinedC );
    }

    if Vec_IntSize(p.vRefinedC) > 0 {
        Cec_ManSimProcessRefined( p, p.vRefinedC ); }
    assert(vInfoCis == nil || iCiId == Gia_ManCiNum(p.pAig), "vInfoCis == nil || iCiId == Gia_ManCiNum(p.pAig)");
    assert(vInfoCos == nil || iCoId == Gia_ManCoNum(p.pAig), "vInfoCos == nil || iCoId == Gia_ManCoNum(p.pAig)");
    assert(p.nMems == 1, "p.nMems == 1");
    if p.nMems != 1 {
        Abc_Print( 1, "Cec_ManSimSimulateRound(): Memory management error!\n" ); }
    if p.pPars.fVeryVerbose {
        Gia_ManEquivPrintClasses( p.pAig, 0, Cec_MemUsage(p) ); }
    if p.pBestState {
        Cec_ManSimFindBestPattern( p ); }
/*
    if ( p->nMems > 1 ) {
        for ( i = 1; i < p->nObjs; i++ )
        if ( p->pSims[i] ) {
            int x = 0;
        }
    }
*/
    return Cec_ManSimAnalyzeOutputs( p );
}



/**Function*************************************************************

  Synopsis    [Creates simulation info for this round.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimCreateInfo(p *Cec_ManSim_t, vInfoCis *Vec_Ptr_t, vInfoCos *Vec_Ptr_t) {
    var pRes0, pRes1 *uint32;
    var i, w int32;
    if p.pPars.fSeqSimulate && Gia_ManRegNum(p.pAig) > 0 {
        assert((vInfoCis != nil) && (vInfoCos != nil), "vInfoCis && vInfoCos");
        for  i = 0; i < Gia_ManPiNum(p.pAig); i++  {
            pRes0 = (unsigned *)Vec_PtrEntry( vInfoCis, i );
            for  w = 0; w < p.nWords; w++  {
                pRes0[w] = Gia_ManRandom( 0 );
            }
        }
        for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
            pRes0 = (unsigned *)Vec_PtrEntry( vInfoCis, Gia_ManPiNum(p.pAig) + i );
            pRes1 = (unsigned *)Vec_PtrEntry( vInfoCos, Gia_ManPoNum(p.pAig) + i );
            for  w = 0; w < p.nWords; w++  {
                pRes0[w] = pRes1[w];
            }
        }
    
    } else {
        for  i = 0; i < Gia_ManCiNum(p.pAig); i++  {
            pRes0 = (unsigned *)Vec_PtrEntry( vInfoCis, i );
            for  w = 0; w < p.nWords; w++  {
                pRes0[w] = Gia_ManRandom( 0 );
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the bug is found.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimClassesPrepare(p *Cec_ManSim_t, LevelMax int32) int32 {
    var pObj *Gia_Obj_t;
    var i int32;
    assert(p.pAig.pReprs == nil, "p.pAig.pReprs == nil");
    // allocate representation
    // inlining of macro ABC_CALLOC
    p.pAig.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p.pAig)), sizeof(Gia_Rpr_t)));
    // inlining of macro ABC_CALLOC
    p.pAig.pNexts = ((int *) calloc((size_t)(Gia_ManObjNum(p.pAig)), sizeof(int)));
    // create references
    Gia_ManCreateValueRefs( p.pAig );
    // set starting representative of internal nodes to be constant 0
    if p.pPars.fLatchCorr {
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < p.pAig.nObjs) && ((pObj) = Gia_ManObj(p.pAig, i)); i++  {
            Gia_ObjSetRepr( p.pAig, i, GIA_VOID );
        }
    } else if LevelMax == -1 {
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < p.pAig.nObjs) && ((pObj) = Gia_ManObj(p.pAig, i)); i++  {
            Gia_ObjSetRepr( p.pAig, i, Gia_ObjIsAnd(pObj) ? 0 : GIA_VOID );
        }
    } else {
        Gia_ManLevelNum( p.pAig );
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < p.pAig.nObjs) && ((pObj) = Gia_ManObj(p.pAig, i)); i++  {
            Gia_ObjSetRepr( p.pAig, i, (Gia_ObjIsAnd(pObj) && Gia_ObjLevel(p.pAig,pObj) <= LevelMax) ? 0 : GIA_VOID );
        }
        Vec_IntFreeP( &p.pAig.vLevels );
    }
    // if sequential simulation, set starting representative of ROs to be constant 0
    if p.pPars.fSeqSimulate {
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p.pAig)) && ((pObj) = Gia_ManCi(p.pAig, Gia_ManPiNum(p.pAig)+i)); i++  {
            if pObj.Value {
                Gia_ObjSetRepr( p.pAig, Gia_ObjId(p.pAig, pObj), 0 ); }
        } }
    // perform simulation
    if p.pAig.nSimWords {
        p.nWords = 2*p.pAig.nSimWords;
        assert(Vec_WrdSize(p.pAig.vSimsPi) == Gia_ManCiNum(p.pAig) * p.pAig.nSimWords, "Vec_WrdSize(p.pAig.vSimsPi) == Gia_ManCiNum(p.pAig) * p.pAig.nSimWords"); 
        //Cec_ManSimCreateInfo( p, p->vCiSimInfo, p->vCoSimInfo );
        for  i = 0; i < Gia_ManCiNum(p.pAig); i++  {
            memmove( Vec_PtrEntry(p.vCiSimInfo, i), Vec_WrdEntryP(p.pAig.vSimsPi, i*p.pAig.nSimWords), sizeof(word)*p.pAig.nSimWords );
        }
        if Cec_ManSimSimulateRound( p, p.vCiSimInfo, p.vCoSimInfo ) != 0 {
            return 1; }
        if p.pPars.fVerbose {
            Gia_ManEquivPrintClasses( p.pAig, 0, Cec_MemUsage(p) ); }
    
    } else {
        p.nWords = 1;
        for {
            if p.pPars.fVerbose {
                Gia_ManEquivPrintClasses( p.pAig, 0, Cec_MemUsage(p) ); }
            for  i = 0; i < 4; i++  {
                Cec_ManSimCreateInfo( p, p.vCiSimInfo, p.vCoSimInfo );
                if Cec_ManSimSimulateRound( p, p.vCiSimInfo, p.vCoSimInfo ) != 0 {
                    return 1; }
            }
            p.nWords = 2 * p.nWords + 1;
        
if !( p.nWords <= p.pPars.nWords ) { break }
}
        
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the bug is found.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimClassesRefine(p *Cec_ManSim_t) int32 {
    var i int32;
    Gia_ManCreateValueRefs( p.pAig );
    p.nWords = p.pPars.nWords;
    for  i = 0; i < p.pPars.nRounds; i++  {
        if (i % (p.pPars.nRounds / 5)) == 0 && p.pPars.fVerbose {
            Gia_ManEquivPrintClasses( p.pAig, 0, Cec_MemUsage(p) ); }
        Cec_ManSimCreateInfo( p, p.vCiSimInfo, p.vCoSimInfo );
        if Cec_ManSimSimulateRound( p, p.vCiSimInfo, p.vCoSimInfo ) != 0 {
            return 1; }
    }
    if p.pPars.fVerbose {
        Gia_ManEquivPrintClasses( p.pAig, 0, Cec_MemUsage(p) ); }
    return 0;
}
////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


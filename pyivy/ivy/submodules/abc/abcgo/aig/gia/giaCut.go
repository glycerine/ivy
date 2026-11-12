package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaCut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Stand-alone cut computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaCut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const GIA_MAX_CUTSIZE =    8
const GIA_MAX_CUTNUM =     51
#define GIA_MAX_TT_WORDS   ((GIA_MAX_CUTSIZE > 6) ? 1 << (GIA_MAX_CUTSIZE-6) : 1)

const GIA_CUT_NO_LEAF =   0xF

    
type Gia_Cut_t struct {
    Sign            word;                      // signature
    iFunc             int32;                     // functionality
    Cost             int32;                      // cut cost
    CostLev             int32;                   // cut cost
    nTreeLeaves uint32; /* C bitfield: 28 bits; review packing and truncation */         // tree leaves
    nLeaves uint32; /* C bitfield: 4 bits; review packing and truncation */         // leaf count
    pLeaves [GIA_MAX_CUTSIZE]int32;  // leaves
}

    
type Gia_Sto_t struct {
    nCutSize             int32;
    nCutNum             int32;
    fCutMin             int32;
    fTruthMin             int32;
    fVerbose             int32;
    pGia      *Gia_Man_t;                      // user's AIG manager (will be modified by adding nodes)
    vRefs      *Vec_Int_t;                     // refs for each node
    vCuts      *Vec_Wec_t;                     // cuts for each node
    vTtMem      *Vec_Mem_t;                    // truth tables
    pCuts [3][GIA_MAX_CUTNUM]Gia_Cut_t;  // temporary cuts
    ppCuts [GIA_MAX_CUTNUM]*Gia_Cut_t;    // temporary cut pointers
    nCutsR             int32;                    // the number of cuts
    Pivot             int32;                     // current object
    iCutBest             int32;                  // best-delay cut
    nCutsOver             int32;                 // overflow cuts
    CutCount [4]float64;               // cut counters
    clkStart         abctime;                  // starting time
}

func Gia_CutTruth(p *Gia_Sto_t, pCut *Gia_Cut_t) *word { return Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut.iFunc));           }

// Original macro:
// #define Sdb_ForEachCut( pList, pCut, i ) for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += pCut[0] + 2 )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Check correctness of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_CutGetSign(pCut *Gia_Cut_t) word {
    var Sign word; var i int32; 
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Sign |= ((word)1) << (pCut.pLeaves[i] & 0x3F);
    }
    return Sign;
}
static inline int Gia_CutCheck( Gia_Cut_t * pBase, Gia_Cut_t * pCut ) // check if pCut is contained in pBase
{
    var nSizeB int32  = pBase.nLeaves;
    var nSizeC int32  = pCut.nLeaves;
    var i int32;
    var pB *int32 = pBase.pLeaves;
    var k int32;
    var pC *int32 = pCut.pLeaves;
    for  i = 0; i < nSizeC; i++  {
        for  k = 0; k < nSizeB; k++  {
            if pC[i] == pB[k] {
                break; }
        }
        if k == nSizeB {
            return 0; }
    }
    return 1;
}
func Gia_CutSetCheckArray(ppCuts **Gia_Cut_t, nCuts int32) int32 {
    var pCut0, pCut1 *Gia_Cut_t; 
    var i, k, m, n, Value int32;
    assert(nCuts > 0, "nCuts > 0");
    for  i = 0; i < nCuts; i++  {
        pCut0 = ppCuts[i];
        assert(pCut0.nLeaves <= GIA_MAX_CUTSIZE, "pCut0.nLeaves <= GIA_MAX_CUTSIZE");
        assert(pCut0.Sign == Gia_CutGetSign(pCut0), "pCut0.Sign == Gia_CutGetSign(pCut0)");
        // check duplicates
        for  m = 0; m < int32(pCut0.nLeaves); m++  {
        for  n = m + 1; n < int32(pCut0.nLeaves); n++  {
            assert(pCut0.pLeaves[m] < pCut0.pLeaves[n], "pCut0.pLeaves[m] < pCut0.pLeaves[n]");
        }
        }
        // check pairs
        for  k = 0; k < nCuts; k++  {
            pCut1 = ppCuts[k];
            if pCut0 == pCut1 {
                continue; }
            // check containments
            Value = Gia_CutCheck( pCut0, pCut1 );
            assert(Value == 0, "Value == 0");
        }
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_CutMergeOrder(pCut0 *Gia_Cut_t, pCut1 *Gia_Cut_t, pCut *Gia_Cut_t, nCutSize int32) int32 { 
    var nSize0 int32    = pCut0.nLeaves;
    var nSize1 int32    = pCut1.nLeaves;
    var i int32;
    var pC0 *int32 = pCut0.pLeaves;
    var k int32;
    var pC1 *int32 = pCut1.pLeaves;
    var c int32;
    var pC *int32 = pCut.pLeaves;
    // the case of the largest cut sizes
    if nSize0 == nCutSize && nSize1 == nCutSize {
        for  i = 0; i < nSize0; i++  {
            if pC0[i] != pC1[i] {  return 0; }
            pC[i] = pC0[i];
        }
        pCut.nLeaves = nCutSize;
        pCut.iFunc = -1;
        pCut.Sign = pCut0.Sign | pCut1.Sign;
        return 1;
    }
    // compare two cuts with different numbers
    i = k = c = 0;
    if nSize0 == 0 { goto FlushCut1; }
    if nSize1 == 0 { goto FlushCut0; }
    for {
        if c == nCutSize { return 0; }
        if pC0[i] < pC1[k] {
            pC[c++] = pC0[i++];
            if i >= nSize0 { goto FlushCut1; }
        
        } else if pC0[i] > pC1[k] {
            pC[c++] = pC1[k++];
            if k >= nSize1 { goto FlushCut0; }
        
        } else {
            pC[c++] = pC0[i++]; k++;
            if i >= nSize0 { goto FlushCut1; }
            if k >= nSize1 { goto FlushCut0; }
        }
    }

FlushCut0:
    if c + nSize0 > nCutSize + i { return 0; }
    for  i < nSize0  {
        pC[c++] = pC0[i++];
    }
    pCut.nLeaves = c;
    pCut.iFunc = -1;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    return 1;

FlushCut1:
    if c + nSize1 > nCutSize + k { return 0; }
    for  k < nSize1  {
        pC[c++] = pC1[k++];
    }
    pCut.nLeaves = c;
    pCut.iFunc = -1;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    return 1;
}
func Gia_CutMergeOrder2(pCut0 *Gia_Cut_t, pCut1 *Gia_Cut_t, pCut *Gia_Cut_t, nCutSize int32) int32 { 
    var x0 int32;
    var i0 int32;
    var nSize0 int32 = pCut0.nLeaves;
    var pC0 *int32 = pCut0.pLeaves;
    var x1 int32;
    var i1 int32;
    var nSize1 int32 = pCut1.nLeaves;
    var pC1 *int32 = pCut1.pLeaves;
    var xMin int32;
    var c int32;
    var pC *int32 = pCut.pLeaves;
    for {
        if i0 == nSize0 {
            x0 = ABC_INFINITY;
        } else {
            x0 = pC0[i0];
        }
        if i1 == nSize1 {
            x1 = ABC_INFINITY;
        } else {
            x1 = pC1[i1];
        }
        xMin = Abc_MinInt(x0, x1);
        if xMin == ABC_INFINITY { break; }
        if c == nCutSize { return 0; }
        pC[c++] = xMin;
        if x0 == xMin { i0++; }
        if x1 == xMin { i1++; }
    }
    pCut.nLeaves = c;
    pCut.iFunc = -1;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    return 1;
}
static inline int Gia_CutSetCutIsContainedOrder( Gia_Cut_t * pBase, Gia_Cut_t * pCut ) // check if pCut is contained in pBase
{
    var i int32;
    var nSizeB int32 = pBase.nLeaves;
    var k int32;
    var nSizeC int32 = pCut.nLeaves;
    if nSizeB == nSizeC {
        for  i = 0; i < nSizeB; i++  {
            if pBase.pLeaves[i] != pCut.pLeaves[i] {
                return 0; }
        }
        return 1;
    }
    assert(nSizeB > nSizeC, "nSizeB > nSizeC"); 
    if nSizeC == 0 {
        return 1; }
    for  i = k = 0; i < nSizeB; i++  {
        if pBase.pLeaves[i] > pCut.pLeaves[k] {
            return 0; }
        if pBase.pLeaves[i] == pCut.pLeaves[k] {
            if k++; k == nSizeC {
                return 1; }
        }
    }
    return 0;
}
func Gia_CutSetLastCutIsContained(pCuts **Gia_Cut_t, nCuts int32) int32 {
    var i int32;
    for  i = 0; i < nCuts; i++  {
        if pCuts[i].nLeaves <= pCuts[nCuts].nLeaves && (pCuts[i].Sign & pCuts[nCuts].Sign) == pCuts[i].Sign && Gia_CutSetCutIsContainedOrder(pCuts[nCuts], pCuts[i]) {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_CutCompare(pCut0 *Gia_Cut_t, pCut1 *Gia_Cut_t) int32 {
    if pCut0.nTreeLeaves < pCut1.nTreeLeaves {  return -1; }
    if pCut0.nTreeLeaves > pCut1.nTreeLeaves {  return  1; }
    if pCut0.nLeaves     < pCut1.nLeaves {      return -1; }
    if pCut0.nLeaves     > pCut1.nLeaves {      return  1; }
    return 0;
}
func Gia_CutSetLastCutContains(pCuts **Gia_Cut_t, nCuts int32) int32 {
    var i, k, fChanges int32;
    for  i = 0; i < nCuts; i++  {
        if pCuts[nCuts].nLeaves < pCuts[i].nLeaves && (pCuts[nCuts].Sign & pCuts[i].Sign) == pCuts[nCuts].Sign && Gia_CutSetCutIsContainedOrder(pCuts[i], pCuts[nCuts]) {
            pCuts[i].nLeaves = GIA_CUT_NO_LEAF, fChanges = 1; }
    }
    if !(fChanges != 0) {
        return nCuts; }
    for  i = k = 0; i <= nCuts; i++  {
        if pCuts[i].nLeaves == GIA_CUT_NO_LEAF {
            continue; }
        if k < i {
            // inlining of macro ABC_SWAP
             var t *Gia_Cut_t  = pCuts[k]; pCuts[k] = pCuts[i]; pCuts[i] = t; };
        k++;
    }
    return k - 1;
}
func Gia_CutSetSortByCost(pCuts **Gia_Cut_t, nCuts int32) {
    var i int32;
    for  i = nCuts; i > 0; i--  {
        if Gia_CutCompare(pCuts[i - 1], pCuts[i]) < 0 {//!= 1 )
            return; }
        // inlining of macro ABC_SWAP
        { var t *Gia_Cut_t  = pCuts[i - 1]; pCuts[i - 1] = pCuts[i]; pCuts[i] = t; };
    }
}
func Gia_CutSetAddCut(pCuts **Gia_Cut_t, nCuts int32, nCutNum int32) int32 {
    if nCuts == 0 {
        return 1; }
    nCuts = Gia_CutSetLastCutContains(pCuts, nCuts);
    assert(nCuts >= 0, "nCuts >= 0");
    Gia_CutSetSortByCost( pCuts, nCuts );
    // add new cut if there is room
    return Abc_MinInt( nCuts + 1, nCutNum - 1 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_CutComputeTruth6(p *Gia_Sto_t, pCut0 *Gia_Cut_t, pCut1 *Gia_Cut_t, fCompl0 int32, fCompl1 int32, pCutR *Gia_Cut_t, fIsXor int32) int32 {
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var fCompl int32; var t word;
    var t0 word  = *Gia_CutTruth(p, pCut0);
    var t1 word  = *Gia_CutTruth(p, pCut1);
    if Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 { t0 = ~t0; }
    if Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 { t1 = ~t1; }
    t0 = Abc_Tt6Expand( t0, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    t1 = Abc_Tt6Expand( t1, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    if fIsXor != 0 {
        t = t0 ^ t1;
    } else {
        t = t0 & t1;
    }
    if fCompl = int32((t & 1)) { t = ~t; }
    if p.fTruthMin {
        pCutR.nLeaves = Abc_Tt6MinBase( &t, pCutR.pLeaves, pCutR.nLeaves ); }
    assert(int32((t & 1)) == 0, "int32((t & 1)) == 0");
    truthId        = Vec_MemHashInsert(p.vTtMem, &t);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
}
func Gia_CutComputeTruth(p *Gia_Sto_t, pCut0 *Gia_Cut_t, pCut1 *Gia_Cut_t, fCompl0 int32, fCompl1 int32, pCutR *Gia_Cut_t, fIsXor int32) int32 {
    if p.nCutSize <= 6 {
        return Gia_CutComputeTruth6( p, pCut0, pCut1, fCompl0, fCompl1, pCutR, fIsXor ); }
    {
    var uTruth, uTruth0, uTruth1 [GIA_MAX_TT_WORDS]word;
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var nCutSize int32 = p.nCutSize;
    var fCompl int32;
    var nWords int32      = Abc_Truth6WordNum(nCutSize);
    var pTruth0 *word  = Gia_CutTruth(p, pCut0);
    var pTruth1 *word  = Gia_CutTruth(p, pCut1);
    Abc_TtCopy( uTruth0, pTruth0, nWords, Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 );
    Abc_TtCopy( uTruth1, pTruth1, nWords, Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 );
    Abc_TtExpand( uTruth0, nCutSize, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    Abc_TtExpand( uTruth1, nCutSize, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    if fIsXor != 0 {
        Abc_TtXor( uTruth, uTruth0, uTruth1, nWords, (fCompl = int32(((uTruth0[0] ^ uTruth1[0]) & 1))) );
    } else {
        Abc_TtAnd( uTruth, uTruth0, uTruth1, nWords, (fCompl = int32(((uTruth0[0] & uTruth1[0]) & 1))) ); }
    if p.fTruthMin {
        pCutR.nLeaves = Abc_TtMinBase( uTruth, pCutR.pLeaves, pCutR.nLeaves, nCutSize ); }
    assert((uTruth[0] & 1) == 0, "(uTruth[0] & 1) == 0");
//Kit_DsdPrintFromTruth( uTruth, pCutR->nLeaves ), printf("\n" ), printf("\n" );
    truthId        = Vec_MemHashInsert(p.vTtMem, uTruth);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_CutCountBits(i word) int32 {
    i = i - ((i >> 1) & 0x5555555555555555);
    i = (i & 0x3333333333333333) + ((i >> 2) & 0x3333333333333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F0F0F0F0F);
    return (i*(0x0101010101010101))>>56;
}
func Gia_CutAddUnit(p *Gia_Sto_t, iObj int32) {
    var vThis *Vec_Int_t  = Vec_WecEntry( p.vCuts, iObj );
    if Vec_IntSize(vThis) == 0 {
        Vec_IntPush( vThis, 1 );
    } else { 
        Vec_IntAddToEntry( vThis, 0, 1 ); }
    Vec_IntPush( vThis, 1 );
    Vec_IntPush( vThis, iObj );
    Vec_IntPush( vThis, 2 );
}
func Gia_CutAddZero(p *Gia_Sto_t, iObj int32) {
    var vThis *Vec_Int_t  = Vec_WecEntry( p.vCuts, iObj );
    assert(Vec_IntSize(vThis) == 0, "Vec_IntSize(vThis) == 0");
    Vec_IntPush( vThis, 1 );
    Vec_IntPush( vThis, 0 );
    Vec_IntPush( vThis, 0 );
}
func Gia_CutTreeLeaves(p *Gia_Sto_t, pCut *Gia_Cut_t) int32 {
    var i, Cost int32;
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Cost += Vec_IntEntry( p.vRefs, pCut.pLeaves[i] ) == 1;
    }
    return Cost;
}
func Gia_StoPrepareSet(p *Gia_Sto_t, iObj int32, Index int32) int32 {
    var vThis *Vec_Int_t  = Vec_WecEntry( p.vCuts, iObj );
    var i int32;
    var v int32;
    var pCut *int32;
    var pList *int32 = Vec_IntArray( vThis );
    // inlining of macro Sdb_ForEachCut
    for  i, pCut = 0, pList + 1; i < pList[0]; i++, pCut += pCut[0] + 2  {
        var pCutTemp *Gia_Cut_t  = &p.pCuts[Index][i];
        pCutTemp.nLeaves = pCut[0];
        for  v = 1; v <= pCut[0]; v++  {
            pCutTemp.pLeaves[v-1] = pCut[v];
        }
        pCutTemp.iFunc = pCut[pCut[0]+1];
        pCutTemp.Sign = Gia_CutGetSign( pCutTemp );
        pCutTemp.nTreeLeaves = Gia_CutTreeLeaves( p, pCutTemp );
    }
    return pList[0];
}
func Gia_StoInitResult(p *Gia_Sto_t) {
    var i int32; 
    for  i = 0; i < GIA_MAX_CUTNUM; i++  {
        p.ppCuts[i] = &p.pCuts[2][i];
    }
}
func Gia_StoStoreResult(p *Gia_Sto_t, iObj int32, pCuts **Gia_Cut_t, nCuts int32) {
    var i, v int32;
    var vList *Vec_Int_t  = Vec_WecEntry( p.vCuts, iObj );
    Vec_IntPush( vList, nCuts );
    for  i = 0; i < nCuts; i++  {
        Vec_IntPush( vList, pCuts[i].nLeaves );
        for  v = 0; v < int32(pCuts[i].nLeaves); v++  {
            Vec_IntPush( vList, pCuts[i].pLeaves[v] );
        }
        Vec_IntPush( vList, pCuts[i].iFunc );
    }
}
func Gia_CutPrint(p *Gia_Sto_t, iObj int32, pCut *Gia_Cut_t) {
    var i int32;
    var nDigits int32 = Abc_Base10Log(Gia_ManObjNum(p.pGia)); 
    if pCut == nil { fmt.Printf("No cut.\n"); return; }
    fmt.Printf("%d  {", pCut.nLeaves);
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        fmt.Printf(" %*d", nDigits, pCut.pLeaves[i]);
    }
    for  ; i < int32(p.nCutSize); i++  {
        fmt.Printf(" %*s", nDigits, " ");
    }
    fmt.Printf("  }  Cost = %3d  CostL = %3d  Tree = %d  ", pCut.Cost, pCut.CostLev, pCut.nTreeLeaves);
    fmt.Printf("\n");
}
func Gia_StoMergeCuts(p *Gia_Sto_t, iObj int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    var fIsXor int32        = Gia_ObjIsXor(pObj);
    var nCutSize int32      = p.nCutSize;
    var nCutNum int32       = p.nCutNum;
    var fComp0 int32        = Gia_ObjFaninC0(pObj);
    var fComp1 int32        = Gia_ObjFaninC1(pObj);
    var Fan0 int32          = Gia_ObjFaninId0(pObj, iObj);
    var Fan1 int32          = Gia_ObjFaninId1(pObj, iObj);
    var nCuts0 int32        = Gia_StoPrepareSet( p, Fan0, 0 );
    var nCuts1 int32        = Gia_StoPrepareSet( p, Fan1, 1 );
    var i, k, nCutsR int32;
    var pCut0 *Gia_Cut_t;
    var pCut1 *Gia_Cut_t;
    var pCutsR **Gia_Cut_t = p.ppCuts;
    assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
    assert(!Gia_ObjIsMux(p.pGia, pObj), "!Gia_ObjIsMux(p.pGia, pObj)");
    Gia_StoInitResult( p );
    p.CutCount[0] += nCuts0 * nCuts1;
    for  i = 0, pCut0 = p.pCuts[0]; i < nCuts0; i, pCut0 = i+1, pCut0+1  {
    for  k = 0, pCut1 = p.pCuts[1]; k < nCuts1; k, pCut1 = k+1, pCut1+1  {
        if int32((pCut0.nLeaves + pCut1.nLeaves)) > nCutSize && Gia_CutCountBits(pCut0.Sign | pCut1.Sign) > nCutSize {
            continue; }
        p.CutCount[1]++; 
        if !(Gia_CutMergeOrder(pCut0, pCut1, pCutsR[nCutsR], nCutSize) != 0) {
            continue; }
        if Gia_CutSetLastCutIsContained(pCutsR, nCutsR) != 0 {
            continue; }
        p.CutCount[2]++;
        if p.fCutMin && (Gia_CutComputeTruth(p, pCut0, pCut1, fComp0, fComp1, pCutsR[nCutsR], fIsXor) != 0) {
            pCutsR[nCutsR].Sign = Gia_CutGetSign(pCutsR[nCutsR]); }
        pCutsR[nCutsR].nTreeLeaves = Gia_CutTreeLeaves( p, pCutsR[nCutsR] );
        nCutsR = Gia_CutSetAddCut( pCutsR, nCutsR, nCutNum );
    }
    }
    p.CutCount[3] += nCutsR;
    p.nCutsOver += nCutsR == nCutNum-1;
    p.nCutsR = nCutsR;
    p.Pivot = iObj;
    // debug printout
    if false {
        fmt.Printf("*** Obj = %4d  NumCuts = %4d\n", iObj, nCutsR);
        for  i = 0; i < nCutsR; i++  {
            Gia_CutPrint( p, iObj, pCutsR[i] );
        }
        fmt.Printf("\n");
    }
    // verify
    assert(nCutsR > 0 && nCutsR < nCutNum, "nCutsR > 0 && nCutsR < nCutNum");
    assert((Gia_CutSetCheckArray(pCutsR, nCutsR) != 0), "Gia_CutSetCheckArray(pCutsR, nCutsR)");
    // store the cutset
    Gia_StoStoreResult( p, iObj, pCutsR, nCutsR );
    if nCutsR > 1 || pCutsR[0].nLeaves > 1 {
        Gia_CutAddUnit( p, iObj ); }
}


/**Function*************************************************************

  Synopsis    [Incremental cut computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_StoAlloc(pGia *Gia_Man_t, nCutSize int32, nCutNum int32, fCutMin int32, fTruthMin int32, fVerbose int32) *Gia_Sto_t {
    
    assert(nCutSize < GIA_CUT_NO_LEAF, "nCutSize < GIA_CUT_NO_LEAF");
    assert(nCutSize > 1 && nCutSize <= GIA_MAX_CUTSIZE, "nCutSize > 1 && nCutSize <= GIA_MAX_CUTSIZE");
    assert(nCutNum > 1 && nCutNum < GIA_MAX_CUTNUM, "nCutNum > 1 && nCutNum < GIA_MAX_CUTNUM");
    // inlining of macro ABC_CALLOC
    p := &Gia_Sto_t{};
    p.clkStart  = Abc_Clock();
    p.nCutSize  = nCutSize;
    p.nCutNum   = nCutNum;
    p.fCutMin   = fCutMin;
    p.fTruthMin = fTruthMin;
    p.fVerbose  = fVerbose;
    p.pGia      = pGia;
    p.vRefs     = Vec_IntAlloc( Gia_ManObjNum(pGia) );
    p.vCuts     = Vec_WecStart( Gia_ManObjNum(pGia) );
    if fCutMin != 0 {
        p.vTtMem = Vec_MemAllocForTT( nCutSize, 0 );
    } else {
        p.vTtMem = nil;
    }
    return p;
}
func Gia_StoFree(p *Gia_Sto_t) {
    Vec_IntFree( p.vRefs );
    Vec_WecFree( p.vCuts );
    if p.fCutMin {
        Vec_MemHashFree( p.vTtMem ); }
    if p.fCutMin {
        Vec_MemFree( p.vTtMem ); }
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Gia_StoComputeCutsConst0(p *Gia_Sto_t, iObj int32) {
    Gia_CutAddZero( p, iObj );
}
func Gia_StoComputeCutsCi(p *Gia_Sto_t, iObj int32) {
    Gia_CutAddUnit( p, iObj );
}
func Gia_StoComputeCutsNode(p *Gia_Sto_t, iObj int32) {
    Gia_StoMergeCuts( p, iObj );
}
func Gia_StoRefObj(p *Gia_Sto_t, iObj int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    assert(iObj == Vec_IntSize(p.vRefs), "iObj == Vec_IntSize(p.vRefs)");
    Vec_IntPush( p.vRefs, 0 );
    if Gia_ObjIsAnd(pObj) {
        Vec_IntAddToEntry( p.vRefs, Gia_ObjFaninId0(pObj, iObj), 1 );
        Vec_IntAddToEntry( p.vRefs, Gia_ObjFaninId1(pObj, iObj), 1 );
    
    } else if Gia_ObjIsCo(pObj) {
        Vec_IntAddToEntry( p.vRefs, Gia_ObjFaninId0(pObj, iObj), 1 ); }
}
func Gia_StoComputeCuts(pGia *Gia_Man_t) {
    var nCutSize int32   =  6;
    var nCutNum int32    = 25;
    var fCutMin int32    =  1;
    var fTruthMin int32  =  1;
    var fVerbose int32   =  1;
    var p *Gia_Sto_t  = Gia_StoAlloc( pGia, nCutSize, nCutNum, fCutMin, fTruthMin, fVerbose );
    var pObj *Gia_Obj_t;  var i, iObj int32;
    assert(nCutSize <= GIA_MAX_CUTSIZE, "nCutSize <= GIA_MAX_CUTSIZE");
    assert(nCutNum  <  GIA_MAX_CUTNUM, "nCutNum  <  GIA_MAX_CUTNUM");
    // prepare references
    // inlining of macro Gia_ManForEachObj
    for  iObj = 0; (iObj < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, iObj)); iObj++  {
        Gia_StoRefObj( p, iObj );
    }
    // compute cuts
    Gia_StoComputeCutsConst0( p, 0 );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((iObj) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i))); i++  {
        Gia_StoComputeCutsCi( p, iObj );
    }
    // inlining of macro Gia_ManForEachAnd
    for  iObj = 0; (iObj < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, iObj)); iObj++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Gia_StoComputeCutsNode( p, iObj ); }
    }
    if p.fVerbose {
        fmt.Printf("Running cut computation with CutSize = %d  CutNum = %d  CutMin = %s  TruthMin = %s\n", p.nCutSize, p.nCutNum, p.fCutMin ? "yes":"no", p.fTruthMin ? "yes":"no");
        fmt.Printf("CutPair = %.0f  ", p.CutCount[0]);
        fmt.Printf("Merge = %.0f (%.2f %%)  ", p.CutCount[1], 100.0*p.CutCount[1]/p.CutCount[0]);
        fmt.Printf("Eval = %.0f (%.2f %%)  ", p.CutCount[2], 100.0*p.CutCount[2]/p.CutCount[0]);
        fmt.Printf("Cut = %.0f (%.2f %%)  ", p.CutCount[3], 100.0*p.CutCount[3]/p.CutCount[0]);
        fmt.Printf("Cut/Node = %.2f  ", p.CutCount[3] / Gia_ManAndNum(p.pGia));
        fmt.Printf("\n");
        fmt.Printf("The number of nodes with cut count over the limit (%d cuts) = %d nodes (out of %d).  ", p.nCutNum, p.nCutsOver, Gia_ManAndNum(pGia));
        Abc_PrintTime( 0, "Time", Abc_Clock() - p.clkStart );
    }
    Gia_StoFree( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


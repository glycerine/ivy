package sbd

import "fmt"

/**CFile****************************************************************

  FileName    [sbdCut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Cut computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sbdCut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sbdInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const SBD_MAX_CUTSIZE =    10
const SBD_MAX_CUTNUM =     501
#define SBD_MAX_TT_WORDS   ((SBD_MAX_CUTSIZE > 6) ? 1 << (SBD_MAX_CUTSIZE-6) : 1)

const SBD_CUT_NO_LEAF =   0xF

    
type Sbd_Cut_t struct {
    Sign            word;                      // signature
    iFunc             int32;                     // functionality
    Cost             int32;                      // cut cost
    CostLev             int32;                   // cut cost
    nTreeLeaves uint32; /* C bitfield: 9 bits; review packing and truncation */         // tree leaves
    nSlowLeaves uint32; /* C bitfield: 9 bits; review packing and truncation */         // slow leaves
    nTopLeaves uint32; /* C bitfield: 10 bits; review packing and truncation */         // top leaves
    nLeaves uint32; /* C bitfield: 4 bits; review packing and truncation */         // leaf count
    pLeaves [SBD_MAX_CUTSIZE]int32;  // leaves
}

type Sbd_Sto_t_ struct {
    nLutSize             int32;
    nCutSize             int32;
    nCutNum             int32;
    fCutMin             int32;
    fVerbose             int32;
    pGia      *Gia_Man_t;                      // user's AIG manager (will be modified by adding nodes)
    vMirrors      *Vec_Int_t;                  // mirrors for each node
    vDelays      *Vec_Int_t;                   // delays for each node
    vLevels      *Vec_Int_t;                   // levels for each node
    vRefs      *Vec_Int_t;                     // refs for each node
    vCuts      *Vec_Wec_t;                     // cuts for each node
    vTtMem      *Vec_Mem_t;                    // truth tables
    pCuts [3][SBD_MAX_CUTNUM]Sbd_Cut_t;  // temporary cuts
    ppCuts [SBD_MAX_CUTNUM]*Sbd_Cut_t;    // temporary cut pointers
    nCutsR             int32;                    // the number of cuts
    Pivot             int32;                     // current object
    iCutBest             int32;                  // best-delay cut
    nCutsSpec             int32;                 // special cuts
    nCutsOver             int32;                 // overflow cuts
    DelayMin             int32;                  // minimum delay
    CutCount [4]float64;               // cut counters
    clkStart         abctime;                  // starting time
}

func Sbd_CutTruth(p *Sbd_Sto_t, pCut *Sbd_Cut_t) *word { return Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut.iFunc));           }

// Original macro:
// #define Sbd_ForEachCut( pList, pCut, i ) for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += pCut[0] + 2 )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Check correctness of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_CutGetSign(pCut *Sbd_Cut_t) word {
    var Sign word; var i int32; 
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Sign |= ((word)1) << (pCut.pLeaves[i] & 0x3F);
    }
    return Sign;
}
static inline int Sbd_CutCheck( Sbd_Cut_t * pBase, Sbd_Cut_t * pCut ) // check if pCut is contained in pBase
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
func Sbd_CutSetCheckArray(ppCuts **Sbd_Cut_t, nCuts int32) int32 {
    var pCut0, pCut1 *Sbd_Cut_t; 
    var i, k, m, n, Value int32;
    assert(nCuts > 0, "nCuts > 0");
    for  i = 0; i < nCuts; i++  {
        pCut0 = ppCuts[i];
        assert(pCut0.nLeaves <= SBD_MAX_CUTSIZE, "pCut0.nLeaves <= SBD_MAX_CUTSIZE");
        assert(pCut0.Sign == Sbd_CutGetSign(pCut0), "pCut0.Sign == Sbd_CutGetSign(pCut0)");
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
            Value = Sbd_CutCheck( pCut0, pCut1 );
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
func Sbd_CutMergeOrder(pCut0 *Sbd_Cut_t, pCut1 *Sbd_Cut_t, pCut *Sbd_Cut_t, nCutSize int32) int32 { 
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
func Sbd_CutMergeOrder2(pCut0 *Sbd_Cut_t, pCut1 *Sbd_Cut_t, pCut *Sbd_Cut_t, nCutSize int32) int32 { 
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
static inline int Sbd_CutSetCutIsContainedOrder( Sbd_Cut_t * pBase, Sbd_Cut_t * pCut ) // check if pCut is contained in pBase
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
func Sbd_CutSetLastCutIsContained(pCuts **Sbd_Cut_t, nCuts int32) int32 {
    var i int32;
    for  i = 0; i < nCuts; i++  {
        if pCuts[i].nLeaves <= pCuts[nCuts].nLeaves && (pCuts[i].Sign & pCuts[nCuts].Sign) == pCuts[i].Sign && Sbd_CutSetCutIsContainedOrder(pCuts[nCuts], pCuts[i]) {
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
func Sbd_CutCompare(pCut0 *Sbd_Cut_t, pCut1 *Sbd_Cut_t) int32 {
    if pCut0.nLeaves <= 4 && pCut1.nLeaves <= 4 {
        if pCut0.nLeaves     < pCut1.nLeaves {      return -1; }
        if pCut0.nLeaves     > pCut1.nLeaves {      return  1; }
        if pCut0.Cost        < pCut1.Cost {      return -1; }
        if pCut0.Cost        > pCut1.Cost {      return  1; }
        if pCut0.CostLev     < pCut1.CostLev {      return -1; }
        if pCut0.CostLev     > pCut1.CostLev {      return  1; }
    
    } else if pCut0.nLeaves <= 4 {
        return -1;
    } else if pCut1.nLeaves <= 4 {
        return  1;
    } else {
        if pCut0.nTreeLeaves < pCut1.nTreeLeaves {  return -1; }
        if pCut0.nTreeLeaves > pCut1.nTreeLeaves {  return  1; }
        if pCut0.Cost        < pCut1.Cost {      return -1; }
        if pCut0.Cost        > pCut1.Cost {      return  1; }
        if pCut0.CostLev     < pCut1.CostLev {      return -1; }
        if pCut0.CostLev     > pCut1.CostLev {      return  1; }
        if pCut0.nLeaves     < pCut1.nLeaves {      return -1; }
        if pCut0.nLeaves     > pCut1.nLeaves {      return  1; }
    }
    return 0;
}
func Sbd_CutCompare2(pCut0 *Sbd_Cut_t, pCut1 *Sbd_Cut_t) int32 {
    assert(pCut0.nLeaves > 4 && pCut1.nLeaves > 4, "pCut0.nLeaves > 4 && pCut1.nLeaves > 4");
    if pCut0.nSlowLeaves < pCut1.nSlowLeaves {  return -1; }
    if pCut0.nSlowLeaves > pCut1.nSlowLeaves {  return  1; }
    if pCut0.nTreeLeaves < pCut1.nTreeLeaves {  return -1; }
    if pCut0.nTreeLeaves > pCut1.nTreeLeaves {  return  1; }
    if pCut0.Cost        < pCut1.Cost {      return -1; }
    if pCut0.Cost        > pCut1.Cost {      return  1; }
    if pCut0.CostLev     < pCut1.CostLev {      return -1; }
    if pCut0.CostLev     > pCut1.CostLev {      return  1; }
    if pCut0.nLeaves     < pCut1.nLeaves {      return -1; }
    if pCut0.nLeaves     > pCut1.nLeaves {      return  1; }
    return 0;
}

func Sbd_CutSetLastCutContains(pCuts **Sbd_Cut_t, nCuts int32) int32 {
    var i, k, fChanges int32;
    for  i = 0; i < nCuts; i++  {
        if pCuts[nCuts].nLeaves < pCuts[i].nLeaves && (pCuts[nCuts].Sign & pCuts[i].Sign) == pCuts[nCuts].Sign && Sbd_CutSetCutIsContainedOrder(pCuts[i], pCuts[nCuts]) {
            pCuts[i].nLeaves = SBD_CUT_NO_LEAF, fChanges = 1; }
    }
    if !(fChanges != 0) {
        return nCuts; }
    for  i = k = 0; i <= nCuts; i++  {
        if pCuts[i].nLeaves == SBD_CUT_NO_LEAF {
            continue; }
        if k < i {
            ABC_SWAP( Sbd_Cut_t *, pCuts[k], pCuts[i] ); }
        k++;
    }
    return k - 1;
}
func Sbd_CutSetSortByCost(pCuts **Sbd_Cut_t, nCuts int32) {
    var i int32;
    for  i = nCuts; i > 0; i--  {
        if Sbd_CutCompare(pCuts[i - 1], pCuts[i]) < 0 {//!= 1 )
            return; }
        ABC_SWAP( Sbd_Cut_t *, pCuts[i - 1], pCuts[i] );
    }
}
func Sbd_CutSetAddCut(pCuts **Sbd_Cut_t, nCuts int32, nCutNum int32) int32 {
    if nCuts == 0 {
        return 1; }
    nCuts = Sbd_CutSetLastCutContains(pCuts, nCuts);
    assert(nCuts >= 0, "nCuts >= 0");
    Sbd_CutSetSortByCost( pCuts, nCuts );
    // add new cut if there is room
    return Abc_MinInt( nCuts + 1, nCutNum - 1 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_CutComputeTruth6(p *Sbd_Sto_t, pCut0 *Sbd_Cut_t, pCut1 *Sbd_Cut_t, fCompl0 int32, fCompl1 int32, pCutR *Sbd_Cut_t, fIsXor int32) int32 {
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var fCompl int32; var t word;
    var t0 word  = *Sbd_CutTruth(p, pCut0);
    var t1 word  = *Sbd_CutTruth(p, pCut1);
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
    pCutR.nLeaves = Abc_Tt6MinBase( &t, pCutR.pLeaves, pCutR.nLeaves );
    assert(int32((t & 1)) == 0, "int32((t & 1)) == 0");
    truthId        = Vec_MemHashInsert(p.vTtMem, &t);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
}
func Sbd_CutComputeTruth(p *Sbd_Sto_t, pCut0 *Sbd_Cut_t, pCut1 *Sbd_Cut_t, fCompl0 int32, fCompl1 int32, pCutR *Sbd_Cut_t, fIsXor int32) int32 {
    if p.nCutSize <= 6 {
        return Sbd_CutComputeTruth6( p, pCut0, pCut1, fCompl0, fCompl1, pCutR, fIsXor ); }
    {
    var uTruth, uTruth0, uTruth1 [SBD_MAX_TT_WORDS]word;
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var nCutSize int32 = p.nCutSize;
    var fCompl int32;
    var nWords int32      = Abc_Truth6WordNum(nCutSize);
    var pTruth0 *word  = Sbd_CutTruth(p, pCut0);
    var pTruth1 *word  = Sbd_CutTruth(p, pCut1);
    Abc_TtCopy( uTruth0, pTruth0, nWords, Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 );
    Abc_TtCopy( uTruth1, pTruth1, nWords, Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 );
    Abc_TtExpand( uTruth0, nCutSize, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    Abc_TtExpand( uTruth1, nCutSize, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    if fIsXor != 0 {
        Abc_TtXor( uTruth, uTruth0, uTruth1, nWords, (fCompl = int32(((uTruth0[0] ^ uTruth1[0]) & 1))) );
    } else {
        Abc_TtAnd( uTruth, uTruth0, uTruth1, nWords, (fCompl = int32(((uTruth0[0] & uTruth1[0]) & 1))) ); }
    pCutR.nLeaves = Abc_TtMinBase( uTruth, pCutR.pLeaves, pCutR.nLeaves, nCutSize );
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
func Sbd_CutCountBits(i word) int32 {
    i = i - ((i >> 1) & 0x5555555555555555);
    i = (i & 0x3333333333333333) + ((i >> 2) & 0x3333333333333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F0F0F0F0F);
    return (i*(0x0101010101010101))>>56;
}
func Sbd_CutCost(p *Sbd_Sto_t, pCut *Sbd_Cut_t) int32 {
    var i, Cost int32;
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Cost += Vec_IntEntry( p.vDelays, pCut.pLeaves[i] );
    }
    return Cost;
}
func Sbd_CutCostLev(p *Sbd_Sto_t, pCut *Sbd_Cut_t) int32 {
    var i, Cost int32;
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Cost += Vec_IntEntry( p.vLevels, pCut.pLeaves[i] );
    }
    return Cost;
}
func Sbd_CutTreeLeaves(p *Sbd_Sto_t, pCut *Sbd_Cut_t) int32 {
    var i, Cost int32;
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Cost += Vec_IntEntry( p.vRefs, pCut.pLeaves[i] ) == 1;
    }
    return Cost;
}
func Sbd_CutSlowLeaves(p *Sbd_Sto_t, iObj int32, pCut *Sbd_Cut_t) int32 {
    var i int32;
    var Count int32;
    var Delay int32 = Vec_IntEntry(p.vDelays, iObj);
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Count += (Vec_IntEntry(p.vDelays, pCut.pLeaves[i]) - Delay >= -1);
    }
    return Count;
}
func Sbd_CutTopLeaves(p *Sbd_Sto_t, iObj int32, pCut *Sbd_Cut_t) int32 {
    var i int32;
    var Count int32;
    var Delay int32 = Vec_IntEntry(p.vDelays, iObj);
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Count += (Vec_IntEntry(p.vDelays, pCut.pLeaves[i]) - Delay == -2);
    }
    return Count;
}
func Sbd_CutAddUnit(p *Sbd_Sto_t, iObj int32) {
    var vThis *Vec_Int_t  = Vec_WecEntry( p.vCuts, iObj );
    if Vec_IntSize(vThis) == 0 {
        Vec_IntPush( vThis, 1 );
    } else { 
        Vec_IntAddToEntry( vThis, 0, 1 ); }
    Vec_IntPush( vThis, 1 );
    Vec_IntPush( vThis, iObj );
    Vec_IntPush( vThis, 2 );
}
func Sbd_CutAddZero(p *Sbd_Sto_t, iObj int32) {
    var vThis *Vec_Int_t  = Vec_WecEntry( p.vCuts, iObj );
    assert(Vec_IntSize(vThis) == 0, "Vec_IntSize(vThis) == 0");
    Vec_IntPush( vThis, 1 );
    Vec_IntPush( vThis, 0 );
    Vec_IntPush( vThis, 0 );
}
func Sbd_StoPrepareSet(p *Sbd_Sto_t, iObj int32, Index int32) int32 {
    var vThis *Vec_Int_t  = Vec_WecEntry( p.vCuts, iObj );
    var i int32;
    var v int32;
    var pCut *int32;
    var pList *int32 = Vec_IntArray( vThis );
    // inlining of macro Sbd_ForEachCut
    for  i, pCut = 0, pList + 1; i < pList[0]; i++, pCut += pCut[0] + 2  {
        var pCutTemp *Sbd_Cut_t  = &p.pCuts[Index][i];
        pCutTemp.nLeaves = pCut[0];
        for  v = 1; v <= pCut[0]; v++  {
            pCutTemp.pLeaves[v-1] = pCut[v];
        }
        pCutTemp.iFunc = pCut[pCut[0]+1];
        pCutTemp.Sign = Sbd_CutGetSign( pCutTemp );
        pCutTemp.Cost = Sbd_CutCost( p, pCutTemp );
        pCutTemp.CostLev = Sbd_CutCostLev( p, pCutTemp );
        pCutTemp.nTreeLeaves = Sbd_CutTreeLeaves( p, pCutTemp );
        pCutTemp.nSlowLeaves = Sbd_CutSlowLeaves( p, iObj, pCutTemp );
        pCutTemp.nTopLeaves  = Sbd_CutTopLeaves( p, iObj, pCutTemp );
    }
    return pList[0];
}
func Sbd_StoInitResult(p *Sbd_Sto_t) {
    var i int32; 
    for  i = 0; i < SBD_MAX_CUTNUM; i++  {
        p.ppCuts[i] = &p.pCuts[2][i];
    }
}
func Sbd_StoStoreResult(p *Sbd_Sto_t, iObj int32, pCuts **Sbd_Cut_t, nCuts int32) {
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
func Sbd_StoComputeDelay(p *Sbd_Sto_t, iObj int32, pCuts **Sbd_Cut_t, nCuts int32) {
    var i int32;
    var v int32;
    var Delay int32;
    var DelayMin int32 = ABC_INFINITY;
    assert(nCuts > 0, "nCuts > 0");
    p.iCutBest = -1;
    for  i = 0; i < nCuts; i++  {
        if int32(pCuts[i].nLeaves) > p.nLutSize {
            continue; }
        Delay = 0;
        for  v = 0; v < int32(pCuts[i].nLeaves); v++  {
            Delay = Abc_MaxInt( Delay, Vec_IntEntry(p.vDelays, pCuts[i].pLeaves[v]) );
        }
        //DelayMin = Abc_MinInt( DelayMin, Delay );
        if DelayMin > Delay {
            DelayMin = Delay;
            p.iCutBest = i;
        
        } else if DelayMin == Delay && p.iCutBest >= 0 && pCuts[p.iCutBest].nLeaves > pCuts[i].nLeaves {
            p.iCutBest = i; }
    }
    assert(p.iCutBest >= 0, "p.iCutBest >= 0");
    assert(DelayMin < ABC_INFINITY, "DelayMin < ABC_INFINITY");
    if nCuts > 1 || pCuts[0].nLeaves > 1 {
        DelayMin = DelayMin + 1;
    } else {
        DelayMin = DelayMin;
    }
    Vec_IntWriteEntry( p.vDelays, iObj, DelayMin );
    p.DelayMin = Abc_MaxInt( p.DelayMin, DelayMin );
}
func Sbd_StoComputeSpec(p *Sbd_Sto_t, iObj int32, pCuts **Sbd_Cut_t, nCuts int32) {
    var i int32;
    for  i = 0; i < nCuts; i++  {
        pCuts[i].nTopLeaves  = Sbd_CutTopLeaves( p, iObj, pCuts[i] );
        pCuts[i].nSlowLeaves = Sbd_CutSlowLeaves( p, iObj, pCuts[i] );
        p.nCutsSpec += (pCuts[i].nSlowLeaves == 0);
    }
}
func Sbd_CutPrint(p *Sbd_Sto_t, iObj int32, pCut *Sbd_Cut_t) {
    var i int32;
    var nDigits int32 = Abc_Base10Log(Gia_ManObjNum(p.pGia)); 
    var Delay int32  = Vec_IntEntry(p.vDelays, iObj);
    if pCut == nil { fmt.Printf("No cut.\n"); return; }
    fmt.Printf("%d  {", pCut.nLeaves);
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        fmt.Printf(" %*d", nDigits, pCut.pLeaves[i]);
    }
    for  ; i < int32(p.nCutSize); i++  {
        fmt.Printf(" %*s", nDigits, " ");
    }
    fmt.Printf("  }  Cost = %3d  CostL = %3d  Tree = %d  Slow = %d  Top = %d  ", pCut.Cost, pCut.CostLev, pCut.nTreeLeaves, pCut.nSlowLeaves, pCut.nTopLeaves);
    fmt.Printf("%c ", pCut.nSlowLeaves == 0 ? '*' : ' ');
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        fmt.Printf("%3d ", Vec_IntEntry(p.vDelays, pCut.pLeaves[i]) - Delay);
    }
    fmt.Printf("\n");
}
func Sbd_StoMergeCuts(p *Sbd_Sto_t, iObj int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    var fIsXor int32        = Gia_ObjIsXor(pObj);
    var nCutSize int32      = p.nCutSize;
    var nCutNum int32       = p.nCutNum;
    var Lit0m int32;
    if p.vMirrors {
        Lit0m = Vec_IntEntry( p.vMirrors, Gia_ObjFaninId0(pObj, iObj) );
    } else {
        Lit0m = -1;
    }
    var Lit1m int32;
    if p.vMirrors {
        Lit1m = Vec_IntEntry( p.vMirrors, Gia_ObjFaninId1(pObj, iObj) );
    } else {
        Lit1m = -1;
    }
    var fComp0 int32        = Gia_ObjFaninC0(pObj) ^ (Lit0m >= 0 && Abc_LitIsCompl(Lit0m));
    var fComp1 int32        = Gia_ObjFaninC1(pObj) ^ (Lit1m >= 0 && Abc_LitIsCompl(Lit1m));
    var Fan0 int32;
    if Lit0m >= 0 {
        Fan0 = Abc_Lit2Var(Lit0m);
    } else {
        Fan0 = Gia_ObjFaninId0(pObj, iObj);
    }
    var Fan1 int32;
    if Lit1m >= 0 {
        Fan1 = Abc_Lit2Var(Lit1m);
    } else {
        Fan1 = Gia_ObjFaninId1(pObj, iObj);
    }
    var nCuts0 int32        = Sbd_StoPrepareSet( p, Fan0, 0 );
    var nCuts1 int32        = Sbd_StoPrepareSet( p, Fan1, 1 );
    var i, k, nCutsR int32;
    var pCut0 *Sbd_Cut_t;
    var pCut1 *Sbd_Cut_t;
    var pCutsR **Sbd_Cut_t = p.ppCuts;
    assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
    assert(!Gia_ObjIsMux(p.pGia, pObj), "!Gia_ObjIsMux(p.pGia, pObj)");
    Sbd_StoInitResult( p );
    p.CutCount[0] += nCuts0 * nCuts1;
    for  i = 0, pCut0 = p.pCuts[0]; i < nCuts0; i, pCut0 = i+1, pCut0+1  {
    for  k = 0, pCut1 = p.pCuts[1]; k < nCuts1; k, pCut1 = k+1, pCut1+1  {
        if int32((pCut0.nLeaves + pCut1.nLeaves)) > nCutSize && Sbd_CutCountBits(pCut0.Sign | pCut1.Sign) > nCutSize {
            continue; }
        p.CutCount[1]++; 
        if !(Sbd_CutMergeOrder(pCut0, pCut1, pCutsR[nCutsR], nCutSize) != 0) {
            continue; }
        if Sbd_CutSetLastCutIsContained(pCutsR, nCutsR) != 0 {
            continue; }
        p.CutCount[2]++;
        if p.fCutMin && (Sbd_CutComputeTruth(p, pCut0, pCut1, fComp0, fComp1, pCutsR[nCutsR], fIsXor) != 0) {
            pCutsR[nCutsR].Sign = Sbd_CutGetSign(pCutsR[nCutsR]); }
        pCutsR[nCutsR].Cost = Sbd_CutCost( p, pCutsR[nCutsR] );
        pCutsR[nCutsR].CostLev = Sbd_CutCostLev( p, pCutsR[nCutsR] );
        pCutsR[nCutsR].nTreeLeaves = Sbd_CutTreeLeaves( p, pCutsR[nCutsR] );
        nCutsR = Sbd_CutSetAddCut( pCutsR, nCutsR, nCutNum );
    }
    }
    Sbd_StoComputeDelay( p, iObj, pCutsR, nCutsR );
    Sbd_StoComputeSpec( p, iObj, pCutsR, nCutsR );
    p.CutCount[3] += nCutsR;
    p.nCutsOver += nCutsR == nCutNum-1;
    p.nCutsR = nCutsR;
    p.Pivot = iObj;
    // debug printout
    if false {
        fmt.Printf("*** Obj = %4d  Delay = %4d  NumCuts = %4d\n", iObj, Vec_IntEntry(p.vDelays, iObj), nCutsR);
        for  i = 0; i < nCutsR; i++  {
            if int32(pCutsR[i].nLeaves) <= p.nLutSize || pCutsR[i].nSlowLeaves < 2 {
                Sbd_CutPrint( p, iObj, pCutsR[i] ); }
        }
        fmt.Printf("\n");
    }
    // verify
    assert(nCutsR > 0 && nCutsR < nCutNum, "nCutsR > 0 && nCutsR < nCutNum");
    assert((Sbd_CutSetCheckArray(pCutsR, nCutsR) != 0), "Sbd_CutSetCheckArray(pCutsR, nCutsR)");
    // store the cutset
    Sbd_StoStoreResult( p, iObj, pCutsR, nCutsR );
    if nCutsR > 1 || pCutsR[0].nLeaves > 1 {
        Sbd_CutAddUnit( p, iObj ); }
}

/**Function*************************************************************

  Synopsis    [Incremental cut computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_StoAlloc(pGia *Gia_Man_t, vMirrors *Vec_Int_t, nLutSize int32, nCutSize int32, nCutNum int32, fCutMin int32, fVerbose int32) *Sbd_Sto_t {
    var p *Sbd_Sto_t;
    assert(nLutSize <= nCutSize, "nLutSize <= nCutSize");
    assert(nCutSize < SBD_CUT_NO_LEAF, "nCutSize < SBD_CUT_NO_LEAF");
    assert(nCutSize > 1 && nCutSize <= SBD_MAX_CUTSIZE, "nCutSize > 1 && nCutSize <= SBD_MAX_CUTSIZE");
    assert(nCutNum > 1 && nCutNum < SBD_MAX_CUTNUM, "nCutNum > 1 && nCutNum < SBD_MAX_CUTNUM");
    p = new(Sbd_Sto_t);
    p.clkStart = Abc_Clock();
    p.nLutSize = nLutSize;
    p.nCutSize = nCutSize;
    p.nCutNum  = nCutNum;
    p.fCutMin  = fCutMin;
    p.fVerbose = fVerbose;
    p.pGia     = pGia;
    p.vMirrors = vMirrors;
    p.vDelays  = Vec_IntStart( Gia_ManObjNum(pGia) );
    p.vLevels  = Vec_IntStart( Gia_ManObjNum(pGia) );
    p.vRefs    = Vec_IntAlloc( Gia_ManObjNum(pGia) );
    p.vCuts    = Vec_WecStart( Gia_ManObjNum(pGia) );
    if fCutMin != 0 {
        p.vTtMem = Vec_MemAllocForTT( nCutSize, 0 );
    } else {
        p.vTtMem = nil;
    }
    return p;
}
func Sbd_StoFree(p *Sbd_Sto_t) {
    Vec_IntFree( p.vDelays );
    Vec_IntFree( p.vLevels );
    Vec_IntFree( p.vRefs );
    Vec_WecFree( p.vCuts );
    if p.fCutMin {
        Vec_MemHashFree( p.vTtMem ); }
    if p.fCutMin {
        Vec_MemFree( p.vTtMem ); }
    ABC_FREE( p );
}
func Sbd_StoComputeCutsObj(p *Sbd_Sto_t, iObj int32, Delay int32, Level int32) {
    if iObj < Vec_IntSize(p.vDelays) {
        Vec_IntWriteEntry( p.vDelays, iObj, Delay );
        Vec_IntWriteEntry( p.vLevels, iObj, Level );
    
    } else {
        assert(iObj == Vec_IntSize(p.vDelays), "iObj == Vec_IntSize(p.vDelays)");
        assert(iObj == Vec_IntSize(p.vLevels), "iObj == Vec_IntSize(p.vLevels)");
        assert(iObj == Vec_WecSize(p.vCuts), "iObj == Vec_WecSize(p.vCuts)");
        Vec_IntPush( p.vDelays, Delay );
        Vec_IntPush( p.vLevels, Level );
        Vec_WecPushLevel( p.vCuts );
    }
}
func Sbd_StoComputeCutsConst0(p *Sbd_Sto_t, iObj int32) {
    Sbd_StoComputeCutsObj( p, iObj, 0, 0 );
    Sbd_CutAddZero( p, iObj );
}
func Sbd_StoComputeCutsCi(p *Sbd_Sto_t, iObj int32, Delay int32, Level int32) {
    Sbd_StoComputeCutsObj( p, iObj, Delay, Level );
    Sbd_CutAddUnit( p, iObj );
}
func Sbd_StoComputeCutsNode(p *Sbd_Sto_t, iObj int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    var Lev0 int32  = Vec_IntEntry( p.vLevels, Gia_ObjFaninId0(pObj, iObj) );
    var Lev1 int32  = Vec_IntEntry( p.vLevels, Gia_ObjFaninId1(pObj, iObj) );
    Sbd_StoComputeCutsObj( p, iObj, -1, 1 + Abc_MaxInt(Lev0, Lev1) );
    Sbd_StoMergeCuts( p, iObj );
    return Vec_IntEntry( p.vDelays, iObj );
}
func Sbd_StoSaveBestDelayCut(p *Sbd_Sto_t, iObj int32, pCut *int32) {
    var pCutBest *Sbd_Cut_t  = p.ppCuts[p.iCutBest]; var i int32;
    assert(iObj == p.Pivot, "iObj == p.Pivot");
    pCut[0] = pCutBest.nLeaves;
    for  i = 0; i < int32(pCutBest.nLeaves); i++  {
        pCut[i+1] = pCutBest.pLeaves[i];
    }
}
func Sbd_StoObjRefs(p *Sbd_Sto_t, iObj int32) int32 {
    return Vec_IntEntry(p.vRefs, iObj);
}
func Sbd_StoRefObj(p *Sbd_Sto_t, iObj int32, iMirror int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    assert(iObj == Vec_IntSize(p.vRefs), "iObj == Vec_IntSize(p.vRefs)");
    assert(iMirror < iObj, "iMirror < iObj");
    Vec_IntPush( p.vRefs, 0 );
//printf( "Ref %d\n", iObj );
    if iMirror > 0 {
        Vec_IntWriteEntry( p.vRefs, iObj, Vec_IntEntry(p.vRefs, iMirror) );
        Vec_IntWriteEntry( p.vRefs, iMirror, 1 );
    }
    if Gia_ObjIsAnd(pObj) {
        var Lit0m int32  = Vec_IntEntry( p.vMirrors, Gia_ObjFaninId0(pObj, iObj) );
        var Lit1m int32  = Vec_IntEntry( p.vMirrors, Gia_ObjFaninId1(pObj, iObj) );
        var Fan0 int32;
        if Lit0m >= 0 {
            Fan0 = Abc_Lit2Var(Lit0m);
        } else {
            Fan0 = Gia_ObjFaninId0(pObj, iObj);
        }
        var Fan1 int32;
        if Lit1m >= 0 {
            Fan1 = Abc_Lit2Var(Lit1m);
        } else {
            Fan1 = Gia_ObjFaninId1(pObj, iObj);
        }
        Vec_IntAddToEntry( p.vRefs, Fan0, 1 );
        Vec_IntAddToEntry( p.vRefs, Fan1, 1 );
    
    } else if Gia_ObjIsCo(pObj) {
        var Lit0m int32  = Vec_IntEntry( p.vMirrors, Gia_ObjFaninId0(pObj, iObj) );
        assert(Lit0m == -1, "Lit0m == -1");
        Vec_IntAddToEntry( p.vRefs, Gia_ObjFaninId0(pObj, iObj), 1 );
    }
}
func Sbd_StoDerefObj(p *Sbd_Sto_t, iObj int32) {
    var pObj *Gia_Obj_t;
    var Lit0m, Lit1m, Fan0, Fan1 int32;
    return;

    pObj = Gia_ManObj(p.pGia, iObj);
    if Vec_IntEntry(p.vRefs, iObj) == 0 {
        fmt.Printf("Ref count mismatch at node %d\n", iObj); }
    assert(Vec_IntEntry(p.vRefs, iObj) > 0, "Vec_IntEntry(p.vRefs, iObj) > 0");
    Vec_IntAddToEntry( p.vRefs, iObj, -1 );
    if Vec_IntEntry( p.vRefs, iObj ) > 0 {
        return; }
    if Gia_ObjIsCi(pObj) {
        return; }
//printf( "Deref %d\n", iObj );
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Lit0m = Vec_IntEntry( p.vMirrors, Gia_ObjFaninId0(pObj, iObj) );
    Lit1m = Vec_IntEntry( p.vMirrors, Gia_ObjFaninId1(pObj, iObj) );
    if Lit0m >= 0 {
        Fan0 = Abc_Lit2Var(Lit0m);
    } else {
        Fan0 = Gia_ObjFaninId0(pObj, iObj);
    }
    if Lit1m >= 0 {
        Fan1 = Abc_Lit2Var(Lit1m);
    } else {
        Fan1 = Gia_ObjFaninId1(pObj, iObj);
    }
    if Fan0 != 0 { Sbd_StoDerefObj( p, Fan0 ); }
    if Fan1 != 0 { Sbd_StoDerefObj( p, Fan1 ); }
}
func Sbd_StoObjBestCut(p *Sbd_Sto_t, iObj int32, nSize int32, pLeaves *int32) int32 {
    var fVerbose int32;
    var pCutBest *Sbd_Cut_t  = nil;   var i int32;
    assert(p.Pivot == iObj, "p.Pivot == iObj");
    if (fVerbose != 0) && iObj % 1000 == 0 {
        fmt.Printf("Node %6d : \n", iObj); }
    for  i = 0; i < p.nCutsR; i++  {
        if (fVerbose != 0) && iObj % 1000 == 0 {
            Sbd_CutPrint( p, iObj, p.ppCuts[i] ); }
        if (nSize != 0) && int32(p.ppCuts[i].nLeaves) != nSize {
            continue; }
        if int32(p.ppCuts[i].nLeaves) > p.nLutSize && 
             int32(p.ppCuts[i].nSlowLeaves) <= 1 &&
             int32(p.ppCuts[i].nTopLeaves) <= p.nLutSize-1 &&
             (pCutBest == nil || Sbd_CutCompare2(pCutBest, p.ppCuts[i]) == 1) {
            pCutBest = p.ppCuts[i]; }
    }
    if (fVerbose != 0) && iObj % 1000 == 0 {
        fmt.Printf("Best cut of size %d:\n", nSize);
        Sbd_CutPrint( p, iObj, pCutBest );
    }
    if pCutBest == nil {
        return -1; }
    assert(pCutBest.nLeaves <= SBD_DIV_MAX, "pCutBest.nLeaves <= SBD_DIV_MAX");
    for  i = 0; i < int32(pCutBest.nLeaves); i++  {
        pLeaves[i] = pCutBest.pLeaves[i];
    }
    return pCutBest.nLeaves;
}
func Sbd_StoComputeCutsTest(pGia *Gia_Man_t) {
    var p *Sbd_Sto_t  = Sbd_StoAlloc( pGia, nil, 4, 8, 100, 1, 1 );
    var pObj *Gia_Obj_t;   
    var i, iObj int32;
    // prepare references
    // inlining of macro Gia_ManForEachObj
    for  iObj = 0; (iObj < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, iObj)); iObj++  {
        Sbd_StoRefObj( p, iObj, -1 );
    }
    // compute cuts
    Sbd_StoComputeCutsConst0( p, 0 );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((iObj) = Gia_ObjId(p.pGia, Gia_ManCi(p.pGia, i))); i++  {
        Sbd_StoComputeCutsCi( p, iObj, 0, 0 );
    }
    // inlining of macro Gia_ManForEachAnd
    for  iObj = 0; (iObj < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, iObj)); iObj++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Sbd_StoComputeCutsNode( p, iObj ); }
    }
    if p.fVerbose {
        fmt.Printf("Running cut computation with LutSize = %d  CutSize = %d  CutNum = %d:\n", p.nLutSize, p.nCutSize, p.nCutNum);
        fmt.Printf("CutPair = %.0f  ", p.CutCount[0]);
        fmt.Printf("Merge = %.0f (%.2f %%)  ", p.CutCount[1], 100.0*p.CutCount[1]/p.CutCount[0]);
        fmt.Printf("Eval = %.0f (%.2f %%)  ", p.CutCount[2], 100.0*p.CutCount[2]/p.CutCount[0]);
        fmt.Printf("Cut = %.0f (%.2f %%)  ", p.CutCount[3], 100.0*p.CutCount[3]/p.CutCount[0]);
        fmt.Printf("Cut/Node = %.2f  ", p.CutCount[3] / Gia_ManAndNum(p.pGia));
        fmt.Printf("\n");
        fmt.Printf("Spec = %4d  ", p.nCutsSpec);
        fmt.Printf("Over = %4d  ", p.nCutsOver);
        fmt.Printf("Lev = %4d  ", p.DelayMin);
        Abc_PrintTime( 0, "Time", Abc_Clock() - p.clkStart );
    }
    Sbd_StoFree( p );
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaBalance.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [AIG balancing.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaBalance.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecHash.h"
//#include "misc/vec/vecQue.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const BAL_LEAF_MAX =    6
const BAL_CUT_MAX =     8
const BAL_SUPER =      50
const BAL_NO_LEAF =    31
const BAL_NO_FUNC =    134217727     // (1<<27)-1

    
type Bal_Cut_t struct {
    Sign             word;           // signature
    Delay              int32;          // delay
    iFunc uint32; /* C bitfield: 27 bits; review packing and truncation */   // function (BAL_NO_FUNC)
    nLeaves uint32; /* C bitfield: 5 bits; review packing and truncation */   // leaf number (Bal_NO_LEAF)
    pLeaves [BAL_LEAF_MAX]int32; // leaves
}

// operation manager
   
type Bal_Man_t struct {
    pGia       *Gia_Man_t;           // user AIG
    nLutSize              int32;       // LUT size
    nCutNum              int32;        // cut number
    fCutMin              int32;        // cut minimization
    fVerbose              int32;       // verbose
    pNew       *Gia_Man_t;           // derived AIG
    vCosts       *Vec_Int_t;         // cost of supergate nodes
    vCutSets       *Vec_Ptr_t;       // object cutsets
    clkStart          abctime;       // starting the clock
}

func Bal_GiaMan(p *Gia_Man_t) *Bal_Man_t { return (Bal_Man_t *)p.pData;           }

func Bal_ObjCost(p *Bal_Man_t, i int32) int32 { return Vec_IntEntry(p.vCosts, i);      }
func Bal_LitCost(p *Bal_Man_t, i int32) int32 { return Bal_ObjCost(p, Abc_Lit2Var(i));  }
func Bal_ObjDelay(p *Bal_Man_t, i int32) int32 { return Bal_ObjCost(p, i) >> 4;          }
func Bal_LitDelay(p *Bal_Man_t, i int32) int32 { return Bal_ObjDelay(p, Abc_Lit2Var(i)); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bal_ManAlloc(pGia *Gia_Man_t, pNew *Gia_Man_t, nLutSize int32, nCutNum int32, fVerbose int32) *Bal_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Bal_Man_t{};
    p.clkStart = Abc_Clock();
    p.pGia     = pGia;
    p.pNew     = pNew;
    p.nLutSize = nLutSize;
    p.nCutNum  = nCutNum;
    p.fVerbose = fVerbose;
    p.vCosts   = Vec_IntAlloc( 3 * Gia_ManObjNum(pGia) / 2 );
    p.vCutSets = Vec_PtrAlloc( 3 * Gia_ManObjNum(pGia) / 2 );
    Vec_IntFill( p.vCosts, Gia_ManObjNum(pNew), 0 );
    Vec_PtrFill( p.vCutSets, Gia_ManObjNum(pNew), nil );
    pNew.pData = p;
    return p;
}
func Bal_ManFree(p *Bal_Man_t) {
    Vec_PtrFreeFree( p.vCutSets );
    Vec_IntFree( p.vCosts );
    // inlining of macro ABC_FREE
    p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bal_CutCountBits(i word) int32 {
    i = i - ((i >> 1) & 0x5555555555555555);
    i = (i & 0x3333333333333333) + ((i >> 2) & 0x3333333333333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F0F0F0F0F);
    return (i*(0x0101010101010101))>>56;
}
func Bal_CutGetSign(pLeaves *int32, nLeaves int32) word {
    var Sign word; var i int32; 
    for  i = 0; i < nLeaves; i++  {
        Sign |= ((word)1) << (pLeaves[i] & 0x3F);
    }
    return Sign;
}
func Bal_CutCreateUnit(p *Bal_Cut_t, i int32, Delay int32) int32 {
    p.iFunc = 2;
    p.Delay = Delay;
    p.nLeaves = 1;
    p.pLeaves[0] = i;
    p.Sign = ((word)1) << (i & 0x3F);
    return 1;
}
var spStatic_Bal_ManPrepareSet_CutTemp [3]Bal_Cut_t;
func Bal_ManPrepareSet(p *Bal_Man_t, iObj int32, Index int32, fUnit int32, ppCutSet **Bal_Cut_t) int32 {
     var i int32;
    if Vec_PtrEntry(p.vCutSets, iObj) == nil || (fUnit != 0) {
        return Bal_CutCreateUnit( (*ppCutSet = spStatic_Bal_ManPrepareSet_CutTemp + Index), iObj, Bal_ObjDelay(p, iObj)+1 ); }
    *ppCutSet = (Bal_Cut_t *)Vec_PtrEntry(p.vCutSets, iObj);
    for  i = 0; i < p.nCutNum; i++  { 
        if (*ppCutSet)[i].nLeaves == BAL_NO_LEAF {
            return i; }
    }
    return i;
}


/**Function*************************************************************

  Synopsis    [Check correctness of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
static inline int Bal_CutCheck( Bal_Cut_t * pBase, Bal_Cut_t * pCut ) // check if pCut is contained in pBase
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
func Bal_SetCheckArray(ppCuts **Bal_Cut_t, nCuts int32) int32 {
    var pCut0, pCut1 *Bal_Cut_t; 
    var i, k, m, n, Value int32;
    assert(nCuts > 0, "nCuts > 0");
    for  i = 0; i < nCuts; i++  {
        pCut0 = ppCuts[i];
        assert(pCut0.nLeaves <= BAL_LEAF_MAX, "pCut0.nLeaves <= BAL_LEAF_MAX");
        assert(pCut0.Sign == Bal_CutGetSign(pCut0.pLeaves, pCut0.nLeaves), "pCut0.Sign == Bal_CutGetSign(pCut0.pLeaves, pCut0.nLeaves)");
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
            Value = Bal_CutCheck( pCut0, pCut1 );
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
func Bal_CutMergeOrder(pCut0 *Bal_Cut_t, pCut1 *Bal_Cut_t, pCut *Bal_Cut_t, nLutSize int32) int32 { 
    var nSize0 int32    = pCut0.nLeaves;
    var nSize1 int32    = pCut1.nLeaves;
    var i int32;
    var pC0 *int32 = pCut0.pLeaves;
    var k int32;
    var pC1 *int32 = pCut1.pLeaves;
    var c int32;
    var pC *int32 = pCut.pLeaves;
    // the case of the largest cut sizes
    if nSize0 == nLutSize && nSize1 == nLutSize {
        for  i = 0; i < nSize0; i++  {
            if pC0[i] != pC1[i] {  return 0; }
            pC[i] = pC0[i];
        }
        pCut.nLeaves = nLutSize;
        pCut.iFunc = BAL_NO_FUNC;
        pCut.Sign = pCut0.Sign | pCut1.Sign;
        pCut.Delay = Abc_MaxInt( pCut0.Delay, pCut1.Delay );
        return 1;
    }
    // compare two cuts with different numbers
    i = k = c = 0;
    for {
        if c == nLutSize { return 0; }
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
    if c + nSize0 > nLutSize + i { return 0; }
    for  i < nSize0  {
        pC[c++] = pC0[i++];
    }
    pCut.nLeaves = c;
    pCut.iFunc = BAL_NO_FUNC;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    pCut.Delay = Abc_MaxInt( pCut0.Delay, pCut1.Delay );
    return 1;

FlushCut1:
    if c + nSize1 > nLutSize + k { return 0; }
    for  k < nSize1  {
        pC[c++] = pC1[k++];
    }
    pCut.nLeaves = c;
    pCut.iFunc = BAL_NO_FUNC;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    pCut.Delay = Abc_MaxInt( pCut0.Delay, pCut1.Delay );
    return 1;
}
func Bal_CutMergeOrderMux(pCut0 *Bal_Cut_t, pCut1 *Bal_Cut_t, pCut2 *Bal_Cut_t, pCut *Bal_Cut_t, nLutSize int32) int32 { 
    var x0 int32;
    var i0 int32;
    var nSize0 int32 = pCut0.nLeaves;
    var pC0 *int32 = pCut0.pLeaves;
    var x1 int32;
    var i1 int32;
    var nSize1 int32 = pCut1.nLeaves;
    var pC1 *int32 = pCut1.pLeaves;
    var x2 int32;
    var i2 int32;
    var nSize2 int32 = pCut2.nLeaves;
    var pC2 *int32 = pCut2.pLeaves;
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
        if i2 == nSize2 {
            x2 = ABC_INFINITY;
        } else {
            x2 = pC2[i2];
        }
        xMin = Abc_MinInt( Abc_MinInt(x0, x1), x2 );
        if xMin == ABC_INFINITY { break; }
        if c == nLutSize { return 0; }
        pC[c++] = xMin;
        if x0 == xMin { i0++; }
        if x1 == xMin { i1++; }
        if x2 == xMin { i2++; }
    }
    pCut.nLeaves = c;
    pCut.iFunc = BAL_NO_FUNC;
    pCut.Sign = pCut0.Sign | pCut1.Sign | pCut2.Sign;
    pCut.Delay = Abc_MaxInt( pCut0.Delay, Abc_MaxInt(pCut1.Delay, pCut2.Delay) );
    return 1;
}
static inline int Bal_SetCutIsContainedOrder( Bal_Cut_t * pBase, Bal_Cut_t * pCut ) // check if pCut is contained in pBase
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
func Bal_SetLastCutIsContained(pCuts **Bal_Cut_t, nCuts int32) int32 {
    var i int32;
    for  i = 0; i < nCuts; i++  {
        if pCuts[i].nLeaves <= pCuts[nCuts].nLeaves && (pCuts[i].Sign & pCuts[nCuts].Sign) == pCuts[i].Sign && Bal_SetCutIsContainedOrder(pCuts[nCuts], pCuts[i]) {
            return 1; }
    }
    return 0;
}
func Bal_SetLastCutContains(pCuts **Bal_Cut_t, nCuts int32) int32 {
    var i, k, fChanges int32;
    for  i = 0; i < nCuts; i++  {
        if pCuts[nCuts].nLeaves < pCuts[i].nLeaves && (pCuts[nCuts].Sign & pCuts[i].Sign) == pCuts[nCuts].Sign && Bal_SetCutIsContainedOrder(pCuts[i], pCuts[nCuts]) {
            pCuts[i].nLeaves = BAL_NO_LEAF, fChanges = 1; }
    }
    if !(fChanges != 0) {
        return nCuts; }
    for  i = k = 0; i <= nCuts; i++  {
        if pCuts[i].nLeaves == BAL_NO_LEAF {
            continue; }
        if k < i {
            // inlining of macro ABC_SWAP
             var t *Bal_Cut_t  = pCuts[k]; pCuts[k] = pCuts[i]; pCuts[i] = t; };
        k++;
    }
    return k - 1;
}
func Bal_CutCompareArea(pCut0 *Bal_Cut_t, pCut1 *Bal_Cut_t) int32 {
    if pCut0.Delay   < pCut1.Delay {  return -1; }
    if pCut0.Delay   > pCut1.Delay {  return  1; }
    if pCut0.nLeaves < pCut1.nLeaves {  return -1; }
    if pCut0.nLeaves > pCut1.nLeaves {  return  1; }
    return 0;
}
func Bal_SetSortByDelay(pCuts **Bal_Cut_t, nCuts int32) {
    var i int32;
    for  i = nCuts; i > 0; i--  {
        if Bal_CutCompareArea(pCuts[i - 1], pCuts[i]) < 0 {//!= 1 )
            return; }
        // inlining of macro ABC_SWAP
        { var t *Bal_Cut_t  = pCuts[i - 1]; pCuts[i - 1] = pCuts[i]; pCuts[i] = t; };
    }
}
func Bal_SetAddCut(pCuts **Bal_Cut_t, nCuts int32, nCutNum int32) int32 {
    if nCuts == 0 {
        return 1; }
    nCuts = Bal_SetLastCutContains(pCuts, nCuts);
    Bal_SetSortByDelay( pCuts, nCuts );
    return Abc_MinInt( nCuts + 1, nCutNum - 1 );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bal_ManDeriveCuts(p *Bal_Man_t, iFan0 int32, iFan1 int32, iFan2 int32, fCompl0 int32, fCompl1 int32, fCompl2 int32, fUnit0 int32, fUnit1 int32, fUnit2 int32, fIsXor int32, Target int32, fSave int32) int32 {
    var pCutSet [BAL_CUT_MAX]Bal_Cut_t;
    var pCutsR [BAL_CUT_MAX]*Bal_Cut_t;
    var pCutSet0, pCutSet1, pCutSet2 *Bal_Cut_t;
    var nCuts0 int32  = Bal_ManPrepareSet( p, iFan0, 0, fUnit0, &pCutSet0 );
    var nCuts1 int32  = Bal_ManPrepareSet( p, iFan1, 1, fUnit1, &pCutSet1 );
    var pCut0 *Bal_Cut_t;
    var pCut0Lim *Bal_Cut_t = pCutSet0 + nCuts0;
    var pCut1 *Bal_Cut_t;
    var pCut1Lim *Bal_Cut_t = pCutSet1 + nCuts1;
    var i, Cost, nCutsR int32;
    memset( pCutSet, 0, sizeof(Bal_Cut_t) * p.nCutNum );
    for  i = 0; i < p.nCutNum; i++  {
        pCutsR[i] = pCutSet + i;
    }
    // enumerate cuts
    if iFan2 > 0 {
        var nCuts2 int32  = Bal_ManPrepareSet( p, iFan2, 2, fUnit2, &pCutSet2 );
        var pCut2 *Bal_Cut_t;
        var pCut2Lim *Bal_Cut_t = pCutSet2 + nCuts2;
        for  pCut0 = pCutSet0; pCut0 < pCut0Lim; pCut0++  {
        for  pCut1 = pCutSet1; pCut1 < pCut1Lim; pCut1++  {
        for  pCut2 = pCutSet2; pCut2 < pCut2Lim; pCut2++  {
            if Bal_CutCountBits(pCut0.Sign | pCut1.Sign | pCut2.Sign) > p.nLutSize {
                continue; }
            if !(Bal_CutMergeOrderMux(pCut0, pCut1, pCut2, pCutsR[nCutsR], p.nLutSize) != 0) {
                continue; }
            assert(pCutsR[nCutsR].Delay == Target, "pCutsR[nCutsR].Delay == Target");
            if Bal_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
//            if ( p->fCutMin && Bal_CutComputeTruthMux(p, pCut0, pCut1, pCut2, fCompl0, fCompl1, fCompl2, pCutsR[nCutsR]) )
//                pCutsR[nCutsR]->Sign = Bal_CutGetSign(pCutsR[nCutsR]->pLeaves, pCutsR[nCutsR]->nLeaves);
            nCutsR = Bal_SetAddCut( pCutsR, nCutsR, p.nCutNum );
        }
        }
        }
    
    } else {
        for  pCut0 = pCutSet0; pCut0 < pCut0Lim; pCut0++  {
        for  pCut1 = pCutSet1; pCut1 < pCut1Lim; pCut1++  {
            if Bal_CutCountBits(pCut0.Sign | pCut1.Sign) > p.nLutSize {
                continue; }
            if !(Bal_CutMergeOrder(pCut0, pCut1, pCutsR[nCutsR], p.nLutSize) != 0) {
                continue; }
            assert(pCutsR[nCutsR].Delay == Target, "pCutsR[nCutsR].Delay == Target");
            if Bal_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
//            if ( p->fCutMin && Bal_CutComputeTruth(p, pCut0, pCut1, fCompl0, fCompl1, pCutsR[nCutsR], fIsXor) )
//                pCutsR[nCutsR]->Sign = Bal_CutGetSign(pCutsR[nCutsR]->pLeaves, pCutsR[nCutsR]->nLeaves);
            nCutsR = Bal_SetAddCut( pCutsR, nCutsR, p.nCutNum );
        }
        }
    }
    if nCutsR == 0 {
        return -1; }
//printf( "%d ", nCutsR );
    Cost = ((pCutsR[0].Delay << 4) | pCutsR[0].nLeaves);
    // verify
    assert(nCutsR > 0 && nCutsR < p.nCutNum, "nCutsR > 0 && nCutsR < p.nCutNum");
    assert((Bal_SetCheckArray(pCutsR, nCutsR) != 0), "Bal_SetCheckArray(pCutsR, nCutsR)");
    // save cuts
    if (fSave != 0) && Cost >= 0 {
        // inlining of macro ABC_CALLOC
        pCutSet0 = ((Bal_Cut_t *) calloc((size_t)(p.nCutNum), sizeof(Bal_Cut_t)));
        Vec_PtrPush( p.vCutSets, pCutSet0 );
        assert(Vec_PtrSize(p.vCutSets) == Gia_ManObjNum(p.pNew), "Vec_PtrSize(p.vCutSets) == Gia_ManObjNum(p.pNew)");
        for  i = 0; i < nCutsR; i++  {
            pCutSet0[i] = *pCutsR[i];
        }        
        for  ; i < p.nCutNum; i++  {
            pCutSet0[i].nLeaves = BAL_NO_LEAF;
        }
        Vec_IntPush( p.vCosts, Cost );
        assert(Vec_IntSize(p.vCosts) == Gia_ManObjNum(p.pNew), "Vec_IntSize(p.vCosts) == Gia_ManObjNum(p.pNew)");
    }
    return Cost;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bal_ManSetGateLevel(p *Bal_Man_t, pObjOld *Gia_Obj_t, iLitNew int32) int32 {
    var iFan0, iFan1, iFan2, Cost int32;
    var fCompl0, fCompl1, fCompl2 int32;
    var fUnit0, fUnit1, fUnit2 int32;
    var Delay0, Delay1, Delay2, DelayMax int32;
    var iObjNew int32  = Abc_Lit2Var(iLitNew);
    var pObjNew *Gia_Obj_t  = Gia_ManObj( p.pNew, iObjNew );
    var fMux int32  = Gia_ObjIsMux(p.pNew, pObjNew);
    if iObjNew < Vec_PtrSize(p.vCutSets) {
        return -1; }
    iFan0    = Gia_ObjFaninId0( pObjNew, iObjNew );
    iFan1    = Gia_ObjFaninId1( pObjNew, iObjNew );
    if fMux != 0 {
        iFan2 = Gia_ObjFaninId2(p.pNew, iObjNew);
    } else {
        iFan2 = 0;
    }
    fCompl0  = Gia_ObjFaninC0( pObjNew );
    fCompl1  = Gia_ObjFaninC1( pObjNew );
    if fMux != 0 {
        fCompl2 = Gia_ObjFaninC2(p.pNew, pObjNew);
    } else {
        fCompl2 = 0;
    }
    Delay0   = Bal_ObjDelay( p, iFan0 );
    Delay1   = Bal_ObjDelay( p, iFan1 );
    Delay2   = Bal_ObjDelay( p, iFan2 );
    DelayMax = Abc_MaxInt( Delay0, Abc_MaxInt(Delay1, Delay2) );
    fUnit0   = int32((Delay0 != DelayMax)); 
    fUnit1   = int32((Delay1 != DelayMax)); 
    fUnit2   = int32((Delay2 != DelayMax)); 
    if DelayMax > 0 {
//printf( "A" );
        Cost = Bal_ManDeriveCuts(p, iFan0, iFan1, iFan2, fCompl0, fCompl1, fCompl2, fUnit0, fUnit1, fUnit2, Gia_ObjIsXor(pObjNew), DelayMax, 1 );
//printf( "B" );
        if Cost >= 0 {
            return Cost; }
    }
    DelayMax++;
    fUnit0 = fUnit1 = fUnit2 = 1;
//printf( "A" );
    Cost = Bal_ManDeriveCuts(p, iFan0, iFan1, iFan2, fCompl0, fCompl1, fCompl2, fUnit0, fUnit1, fUnit2, Gia_ObjIsXor(pObjNew), DelayMax, 1 );
//printf( "B" );
    assert(Cost >= 0, "Cost >= 0");
    return Cost;
}
func Bal_ManEvalTwo(p *Bal_Man_t, iLitNew0 int32, iLitNew1 int32, iLitNew2 int32, fIsXor int32) int32 {
    var iFan0 int32     = Abc_Lit2Var( iLitNew0 );
    var iFan1 int32     = Abc_Lit2Var( iLitNew1 );
    var iFan2 int32     = Abc_Lit2Var( iLitNew2 );
    var fCompl0 int32   = Abc_LitIsCompl( iLitNew0 );
    var fCompl1 int32   = Abc_LitIsCompl( iLitNew1 );
    var fCompl2 int32   = Abc_LitIsCompl( iLitNew2 );
    var Delay0 int32    = Bal_ObjDelay( p, iFan0 );
    var Delay1 int32    = Bal_ObjDelay( p, iFan1 );
    var Delay2 int32    = Bal_ObjDelay( p, iFan2 );
    var DelayMax int32  = Abc_MaxInt( Delay0, Abc_MaxInt(Delay1, Delay2) );
    var fUnit0 int32    = int32((Delay0 != DelayMax)); 
    var fUnit1 int32    = int32((Delay1 != DelayMax)); 
    var fUnit2 int32    = int32((Delay2 != DelayMax)); 
    if DelayMax == 0 {
        return -1; }
    return Bal_ManDeriveCuts(p, iFan0, iFan1, iFan2, fCompl0, fCompl1, fCompl2, fUnit0, fUnit1, fUnit2, fIsXor, DelayMax, 0 );
}

/**Function*************************************************************

  Synopsis    [Sort literals by their cost.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntSelectSortCostLit(vSuper *Vec_Int_t, vCosts *Vec_Int_t) {
    var pArray *int32  = Vec_IntArray(vSuper);
    var nSize int32  = Vec_IntSize(vSuper);
    var i, j, best_i int32;
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if Vec_IntEntry(vCosts, Abc_Lit2Var(pArray[j])) > Vec_IntEntry(vCosts, Abc_Lit2Var(pArray[best_i])) {
                best_i = j; }
        }
        // inlining of macro ABC_SWAP
        { var t int32  = pArray[i]; pArray[i] = pArray[best_i]; pArray[best_i] = t; };
    }
}
func Vec_IntPushOrderCost2(vSuper *Vec_Int_t, vCosts *Vec_Int_t, iLit int32) {
    var i int32;
    var nSize int32;
    var pArray *int32;
    Vec_IntPush( vSuper, iLit );
    pArray = Vec_IntArray(vSuper);
    nSize = Vec_IntSize(vSuper);
    for  i = nSize-1; i > 0; i--  {
        if Vec_IntEntry(vCosts, Abc_Lit2Var(pArray[i])) <= Vec_IntEntry(vCosts, Abc_Lit2Var(pArray[i - 1])) {
            return; }
        // inlining of macro ABC_SWAP
        { var t int32  = pArray[i]; pArray[i] = pArray[i - 1]; pArray[i - 1] = t; };
    }
}
func Vec_IntFindFirstSameDelayAsLast(p *Bal_Man_t, vSuper *Vec_Int_t) int32 {
    var i int32;
    var DelayCur int32;
    var Delay int32 = Bal_LitDelay( p, Vec_IntEntryLast(vSuper) );
    assert(Vec_IntSize(vSuper) > 1, "Vec_IntSize(vSuper) > 1");
    for  i = Vec_IntSize(vSuper)-1; i > 0; i--  {
        DelayCur = Bal_LitDelay( p, Vec_IntEntry(vSuper, i-1) );
        assert(DelayCur >= Delay, "DelayCur >= Delay");
        if DelayCur > Delay {
            return i; }
    }
    return i;
}


/**Function*************************************************************

  Synopsis    [Select the best pair to merge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bal_ManFindBestPair(p *Bal_Man_t, vSuper *Vec_Int_t, pObj *Gia_Obj_t) int32 {
    var pSuper *int32  = Vec_IntArray(vSuper);
    var iBeg int32  = Vec_IntFindFirstSameDelayAsLast( p, vSuper );
    var iEnd int32  = Vec_IntSize(vSuper)-1;
    var i int32;
    var k int32;
    var iBest int32 = -1;
    var kBest int32 = -1;
    var BestCost int32 = ABC_INFINITY;
    var Cost int32;
    assert(iBeg <= iEnd, "iBeg <= iEnd");
    // check if we can add to the higher levels without increasing cost
    for  k = iBeg-1; k >= 0; k--  {
    for  i = iEnd; i >= iBeg; i--  {
        Cost = Bal_ManEvalTwo( p, pSuper[i], pSuper[k], 0, Gia_ObjIsXor(pObj) );
        if Cost == -1 {
            continue; }
        if Cost == Bal_LitCost(p, pSuper[k]) {
//            printf( "A" );
            return (k << 16)|i;
        }
        if BestCost > Cost {
            BestCost = Cost;
iBest = i;
kBest = k; }
    }
    }
    if BestCost != ABC_INFINITY && (BestCost >> 4) == Bal_LitDelay(p, pSuper[kBest]) {
//        printf( "B" );
        return (kBest << 16)|iBest;
    }
    // check if some can be added to lowest level without increasing cost
    BestCost = ABC_INFINITY;
    for  i = iBeg; i <= iEnd; i++  {
    for  k = i+1;  k <= iEnd; k++  {
        Cost = Bal_ManEvalTwo( p, pSuper[i], pSuper[k], 0, Gia_ObjIsXor(pObj) );
        if Cost == -1 {
            continue; }
        if Cost == Abc_MaxInt(Bal_LitCost(p, pSuper[i]), Bal_LitCost(p, pSuper[k])) {
//            printf( "C" );
            return (k << 16)|i;
        }
        if BestCost > Cost {
            BestCost = Cost;
iBest = i;
kBest = k; }
    }
    }
    if BestCost != ABC_INFINITY {
//        printf( "D" );
        return (kBest << 16)|iBest;
    }
//    printf( "E" );
    // group pairs from lowest level based on proximity
    return (iEnd << 16)|(iEnd-1);
}

/**Function*************************************************************

  Synopsis    [Simplify multi-input AND/XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimplifyXor(vSuper *Vec_Int_t) {
    var i int32;
    var k int32;
    var Prev int32 = -1;
    var This int32;
    var fCompl int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSuper)) && (((This) = Vec_IntEntry(vSuper, i)), 1); i++  {
        if This == 0 {
            continue; }
        if This == 1 {
            fCompl ^= 1; 
        } else if Prev != This {
            Vec_IntWriteEntry( vSuper, k++, This );
Prev = This;
        } else {
            Prev = -1;
k--; }
    }
    Vec_IntShrink( vSuper, k );
    if Vec_IntSize( vSuper ) == 0 {
        Vec_IntPush( vSuper, fCompl );
    } else if fCompl != 0 {
        Vec_IntWriteEntry( vSuper, 0, Abc_LitNot(Vec_IntEntry(vSuper, 0)) ); }
}
func Gia_ManSimplifyAnd(vSuper *Vec_Int_t) {
    var i int32;
    var k int32;
    var Prev int32 = -1;
    var This int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSuper)) && (((This) = Vec_IntEntry(vSuper, i)), 1); i++  {
        if This == 0 { Vec_IntFill(vSuper, 1, 0); return; }
        if This == 1 {
            continue; }
        if Prev == -1 || Abc_Lit2Var(Prev) != Abc_Lit2Var(This) {
            Vec_IntWriteEntry( vSuper, k++, This );
Prev = This;
        } else if Prev != This { Vec_IntFill(vSuper, 1, 0); return; }
    }
    Vec_IntShrink( vSuper, k );
    if Vec_IntSize( vSuper ) == 0 {
        Vec_IntPush( vSuper, 1 ); }
}

/**Function*************************************************************

  Synopsis    [Collect multi-input AND/XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSuperCollectXor_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    if !Gia_ObjIsXor(pObj) || 
//        Gia_ObjRefNum(p, pObj) > 1 || 
        Gia_ObjRefNum(p, pObj) > 3 || 
//        (Gia_ObjRefNum(p, pObj) == 2 && (Gia_ObjRefNum(p, Gia_ObjFanin0(pObj)) == 1 || Gia_ObjRefNum(p, Gia_ObjFanin1(pObj)) == 1)) || 
        Vec_IntSize(p.vSuper) > BAL_SUPER {
        Vec_IntPush( p.vSuper, Gia_ObjToLit(p, pObj) );
        return;
    }
    assert(!Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj), "!Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj)");
    Gia_ManSuperCollectXor_rec( p, Gia_ObjFanin0(pObj) );
    Gia_ManSuperCollectXor_rec( p, Gia_ObjFanin1(pObj) );
}
func Gia_ManSuperCollectAnd_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if Gia_IsComplement(pObj) || 
        !Gia_ObjIsAndReal(p, pObj) || 
//        Gia_ObjRefNum(p, pObj) > 1 || 
        Gia_ObjRefNum(p, pObj) > 3 || 
//        (Gia_ObjRefNum(p, pObj) == 2 && (Gia_ObjRefNum(p, Gia_ObjFanin0(pObj)) == 1 || Gia_ObjRefNum(p, Gia_ObjFanin1(pObj)) == 1)) || 
        Vec_IntSize(p.vSuper) > BAL_SUPER {
        Vec_IntPush( p.vSuper, Gia_ObjToLit(p, pObj) );
        return;
    }
    Gia_ManSuperCollectAnd_rec( p, Gia_ObjChild0(pObj) );
    Gia_ManSuperCollectAnd_rec( p, Gia_ObjChild1(pObj) );
}
func Gia_ManSuperCollect(p *Gia_Man_t, pObj *Gia_Obj_t) {
//    int nSize;
    if p.vSuper == nil {
        p.vSuper = Vec_IntAlloc( 1000 );
    } else {
        Vec_IntClear( p.vSuper ); }
    if Gia_ObjIsXor(pObj) {
        assert(!Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj), "!Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj)");
        Gia_ManSuperCollectXor_rec( p, Gia_ObjFanin0(pObj) );
        Gia_ManSuperCollectXor_rec( p, Gia_ObjFanin1(pObj) );
//        nSize = Vec_IntSize(vSuper);
        Vec_IntSort( p.vSuper, 0 );
        Gia_ManSimplifyXor( p.vSuper );
//        if ( nSize != Vec_IntSize(vSuper) )
//            printf( "X %d->%d  ", nSize, Vec_IntSize(vSuper) );
    
    } else if Gia_ObjIsAndReal(p, pObj) {
        Gia_ManSuperCollectAnd_rec( p, Gia_ObjChild0(pObj) );
        Gia_ManSuperCollectAnd_rec( p, Gia_ObjChild1(pObj) );
//        nSize = Vec_IntSize(vSuper);
        Vec_IntSort( p.vSuper, 0 );
        Gia_ManSimplifyAnd( p.vSuper );
//        if ( nSize != Vec_IntSize(vSuper) )
//            printf( "A %d->%d  ", nSize, Vec_IntSize(vSuper) );
    
    } else { assert(false, "0"); }
//    if ( nSize > 10 )
//        printf( "%d ", nSize );
    assert(Vec_IntSize(p.vSuper) > 0, "Vec_IntSize(p.vSuper) > 0");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCreateGate(pNew *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t) {
    var iLit0 int32  = Vec_IntPop(vSuper);
    var iLit1 int32  = Vec_IntPop(vSuper);
    var iLit, i int32;
    if !Gia_ObjIsXor(pObj) {
        iLit = Gia_ManHashAnd( pNew, iLit0, iLit1 );
    } else if pNew.pMuxes {
        iLit = Gia_ManHashXorReal( pNew, iLit0, iLit1 );
    } else { 
        iLit = Gia_ManHashXor( pNew, iLit0, iLit1 ); }
    Vec_IntPush( vSuper, iLit );
    Bal_ManSetGateLevel( Bal_GiaMan(pNew), pObj, iLit );
//    Gia_ObjSetGateLevel( pNew, Gia_ManObj(pNew, Abc_Lit2Var(iLit)) );
    // shift to the corrent location
    for  i = Vec_IntSize(vSuper)-1; i > 0; i--  {
        var iLit1 int32  = Vec_IntEntry(vSuper, i);
        var iLit2 int32  = Vec_IntEntry(vSuper, i-1);
        if Gia_ObjLevelId(pNew, Abc_Lit2Var(iLit1)) <= Gia_ObjLevelId(pNew, Abc_Lit2Var(iLit2)) {
            break; }
        Vec_IntWriteEntry( vSuper, i,   iLit2 );
        Vec_IntWriteEntry( vSuper, i-1, iLit1 );
    }
}
func Gia_ManBalanceGate(pNew *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t, pLits *int32, nLits int32) int32 {
    Vec_IntClear( vSuper );
    if nLits == 1 {
        Vec_IntPush( vSuper, pLits[0] );
    } else if nLits == 2 {
        Vec_IntPush( vSuper, pLits[0] );
        Vec_IntPush( vSuper, pLits[1] );
        Gia_ManCreateGate( pNew, pObj, vSuper );
    
    } else if nLits > 2 {
        var p *Bal_Man_t  = Bal_GiaMan(pNew); var i int32;
        for  i = 0; i < nLits; i++  {
            Vec_IntPush( vSuper, pLits[i] );
        }
        // sort by level/cut-size
        Vec_IntSelectSortCostLit( vSuper, p.vCosts );
        // iterate till everything is grouped
        for  Vec_IntSize(vSuper) > 1  {
            var iLit int32;
            var Res int32 = Bal_ManFindBestPair( p, vSuper, pObj );
            var iBest int32  = Vec_IntEntry( vSuper, Res >> 16 );
            var kBest int32  = Vec_IntEntry( vSuper, Res & 0xFFFF );
            Vec_IntRemove( vSuper, iBest );
            Vec_IntRemove( vSuper, kBest );
            if Gia_ObjIsXor(pObj) {
                iLit = Gia_ManHashXorReal( pNew, iBest, kBest );
            } else { 
                iLit = Gia_ManHashAnd( pNew, iBest, kBest ); }
            Bal_ManSetGateLevel( p, pObj, iLit );
            Vec_IntPushOrderCost2( vSuper, p.vCosts, iLit );
        }
    }
    // consider trivial case
    assert(Vec_IntSize(vSuper) == 1, "Vec_IntSize(vSuper) == 1");
    return Vec_IntEntry(vSuper, 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManBalance_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    var i, iLit, iBeg, iEnd int32;
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    // handle MUX
    if Gia_ObjIsMux(p, pObj) {
        Gia_ManBalance_rec( pNew, p, Gia_ObjFanin0(pObj) );
        Gia_ManBalance_rec( pNew, p, Gia_ObjFanin1(pObj) );
        Gia_ManBalance_rec( pNew, p, Gia_ObjFanin2(p, pObj) );
        pObj.Value = Gia_ManHashMuxReal( pNew, Gia_ObjFanin2Copy(p, pObj), Gia_ObjFanin1Copy(pObj), Gia_ObjFanin0Copy(pObj) );
        Bal_ManSetGateLevel( Bal_GiaMan(pNew), pObj, pObj.Value );
//        Gia_ObjSetGateLevel( pNew, Gia_ManObj(pNew, Abc_Lit2Var(pObj->Value)) );
        return;
    }
    // find supergate
    Gia_ManSuperCollect( p, pObj );
    // save entries
    if p.vStore == nil {
        p.vStore = Vec_IntAlloc( 1000 ); }
    iBeg = Vec_IntSize( p.vStore );
    Vec_IntAppend( p.vStore, p.vSuper );
    iEnd = Vec_IntSize( p.vStore );
    // call recursively
    // inlining of macro Vec_IntForEachEntryStartStop
    for  i = iBeg; (i < iEnd) && (((iLit) = Vec_IntEntry(p.vStore, i)), 1); i++  {
        var pTemp *Gia_Obj_t  = Gia_ManObj( p, Abc_Lit2Var(iLit) );
        Gia_ManBalance_rec( pNew, p, pTemp );
        Vec_IntWriteEntry( p.vStore, i, Abc_LitNotCond(pTemp.Value, Abc_LitIsCompl(iLit)) );
    }
    assert(Vec_IntSize(p.vStore) == iEnd, "Vec_IntSize(p.vStore) == iEnd");
    // consider general case
    pObj.Value = Gia_ManBalanceGate( pNew, pObj, p.vSuper, Vec_IntEntryP(p.vStore, iBeg), iEnd-iBeg );
    Vec_IntShrink( p.vStore, iBeg );
}
func Gia_ManBalanceInt(p *Gia_Man_t, nLutSize int32, nCutNum int32, fVerbose int32) *Gia_Man_t {
    var pMan *Bal_Man_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManFillValue( p );
    Gia_ManCreateRefs( p ); 
    // start the new manager
    pNew = Gia_ManStart( 3*Gia_ManObjNum(p)/2 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // inlining of macro ABC_CALLOC
    pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned)));
    pNew.vLevels = Vec_IntStart( pNew.nObjsAlloc );
    // create constant and inputs
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // create balancing manager
    pMan = Bal_ManAlloc( p, pNew, nLutSize, nCutNum, fVerbose );
    // create internal nodes
    Gia_ManHashStart( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManBalance_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
//    if ( fVerbose )
    {
        var nLevelMax int32;
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(pNew.vCos)) && ((pObj) = Gia_ManCo(pNew, i)); i++  {
            nLevelMax = Abc_MaxInt( nLevelMax, Bal_ObjDelay(pMan, Gia_ObjFaninId0p(pNew, pObj)) );
//            printf( "%d=%d ", i, Bal_ObjDelay(pMan, Gia_ObjFaninId0p(pNew, pObj)) );
        }
        fmt.Printf("Best delay = %d\n", nLevelMax);
    }

//    assert( Gia_ManObjNum(pNew) <= Gia_ManObjNum(p) );
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // delete manager
    Bal_ManFree( pMan );
    // perform cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Gia_ManBalanceLut(p *Gia_Man_t, nLutSize int32, nCutNum int32, fVerbose int32) *Gia_Man_t {
    var pNew, pNew1, pNew2 *Gia_Man_t;
    if fVerbose != 0 {      Gia_ManPrintStats( p, nil ); }
    pNew = Gia_ManDupMuxes( p, 2 );
    if fVerbose != 0 {      Gia_ManPrintStats( pNew, nil ); }
    pNew1 = Gia_ManBalanceInt( pNew, nLutSize, nCutNum, fVerbose );
    if fVerbose != 0 {      Gia_ManPrintStats( pNew1, nil ); }
    Gia_ManStop( pNew );
    pNew2 = Gia_ManDupNoMuxes( pNew1, 0 );
    if fVerbose != 0 {      Gia_ManPrintStats( pNew2, nil ); }
    Gia_ManStop( pNew1 );
    return pNew2;
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


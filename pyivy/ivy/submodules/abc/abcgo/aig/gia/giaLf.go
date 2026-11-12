package gia

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [giaLf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Cut computation.]

  Author      [Alan Mishchenko]`
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaLf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/tim/tim.h"
//#include "misc/vec/vecSet.h"
//#include "misc/vec/vecMem.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const LF_LEAF_MAX =   13
const LF_CUT_MAX =    32
const LF_LOG_PAGE =   12
const LF_NO_LEAF =   255
#define LF_CUT_WORDS (4+LF_LEAF_MAX/2)
#define LF_TT_WORDS  ((LF_LEAF_MAX > 6) ? 1 << (LF_LEAF_MAX-6) : 1)
const LF_EPSILON = 0.005

    
type Lf_Cut_t struct {
    Sign            word;            // signature
    Delay             int32;           // delay
    Flow           float32;            // flow 
    iFunc             int32;           // functionality
    Cost uint32; /* C bitfield: 22 bits; review packing and truncation */  // misc cut cost
    fLate uint32; /* C bitfield: 1 bits; review packing and truncation */  // fails timing
    fMux7 uint32; /* C bitfield: 1 bits; review packing and truncation */  // specialized cut
    nLeaves uint32; /* C bitfield: 8 bits; review packing and truncation */  // the number of leaves
    pLeaves [0]int32;      // leaves
}
    
type Lf_Plc_t struct {
    fUsed uint32; /* C bitfield: 1 bits; review packing and truncation */   // the cut is used
    Handle uint32; /* C bitfield: 31 bits; review packing and truncation */   // the cut handle
}
    
type Lf_Bst_t struct {
    Delay [3]int32;        // delay
    Flow [3]float32;         // flow 
    Cut [2]Lf_Plc_t;          // cut info
}
    
type Lf_Mem_t struct {
    LogPage             int32;         // log size of memory page
    MaskPage             int32;        // page mask
    nCutWords             int32;       // cut size in words
    iCur             int32;            // writing position 
    vPages       Vec_Ptr_t;          // memory pages
    vFree      *Vec_Ptr_t;           // free pages 
}
    
type Lf_Man_t struct {
    // user data
    pGia      *Gia_Man_t;            // manager
    pPars       *Jf_Par_t;           // parameters
    // cut data
    nCutWords             int32;       // cut size in words
    nSetWords             int32;       // set size in words
    pObjBests       *Lf_Bst_t;       // best cuts
    vMemSets       Vec_Ptr_t;        // memory for cutsets
    vFreeSets       Vec_Int_t;       // free cutsets
    vTtMem      *Vec_Mem_t;          // truth tables
    vFreePages       Vec_Ptr_t;      // free memory pages
    vStoreOld        Lf_Mem_t;       // previous cuts
    vStoreNew        Lf_Mem_t;       // current cuts
    // mapper data
    vOffsets       Vec_Int_t;        // offsets
    vRequired       Vec_Int_t;       // required times
    vCutSets       Vec_Int_t;        // cutsets (pObj->Value stores cut refs)
    vFlowRefs       Vec_Flt_t;       // flow refs
    vMapRefs       Vec_Int_t;        // mapping refs
    vSwitches       Vec_Flt_t;       // switching activity
    vCiArrivals       Vec_Int_t;     // arrival times of the CIs
    // statistics
    clkStart         abctime;        // starting time
    CutCount [4]float64;     // cut counts
    Switches          float64;        // switching activity
    nFrontMax             int32;       // frontier
    nCoDrivers             int32;      // CO drivers
    nInverters             int32;      // inverters
    nTimeFails             int32;      // timing fails
    Iter             int32;            // mapping iteration
    fUseEla             int32;         // use exact local area
    nCutMux             int32;         // non-trivial MUX cuts
    nCutEqual             int32;       // equal two cuts
    nCutCounts [LF_LEAF_MAX+1]int32;
}

func Lf_CutCopy(p *Lf_Cut_t, q *Lf_Cut_t, n int32) { memcpy(p, q, sizeof(word) * n);                                         }
func Lf_CutNext(p *Lf_Cut_t, n int32) *Lf_Cut_t { return (Lf_Cut_t *)((word *)p + n);                                     }
func Lf_CutTruth(p *Lf_Man_t, pCut *Lf_Cut_t) *word { return Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut.iFunc));           }

func Lf_ObjOff(p *Lf_Man_t, i int32) int32 { return Vec_IntEntry(&p.vOffsets, i);                                   }
func Lf_ObjRequired(p *Lf_Man_t, i int32) int32 { return Vec_IntEntry(&p.vRequired, i);                                  }
func Lf_ObjSetRequired(p *Lf_Man_t, i int32, t int32) { Vec_IntDowndateEntry(&p.vRequired, i, t);                              }
func Lf_ObjReadBest(p *Lf_Man_t, i int32) *Lf_Bst_t { return p.pObjBests + Lf_ObjOff(p,i);                                   }
func Lf_ObjFlowRefs(p *Lf_Man_t, i int32) float32 { return Vec_FltEntry(&p.vFlowRefs, Lf_ObjOff(p,i));                     }
func Lf_ObjMapRefNum(p *Lf_Man_t, i int32) int32 { return Vec_IntEntry(&p.vMapRefs, Lf_ObjOff(p,i));                      }
func Lf_ObjMapRefInc(p *Lf_Man_t, i int32) int32 { return (*Vec_IntEntryP(&p.vMapRefs, Lf_ObjOff(p,i)))++;                }
func Lf_ObjMapRefDec(p *Lf_Man_t, i int32) int32 { return --(*Vec_IntEntryP(&p.vMapRefs, Lf_ObjOff(p,i)));                }
func Lf_ObjSwitches(p *Lf_Man_t, i int32) float32 { return Vec_FltEntry(&p.vSwitches, i);                                  }
func Lf_BestDiffCuts(p *Lf_Bst_t) int32 { return p.Cut[0].Handle != p.Cut[1].Handle;                            }
func Lf_BestIsMapped(p *Lf_Bst_t) int32 { return int32((p.Cut[0].fUsed ^ p.Cut[1].fUsed));                        }
func Lf_BestIndex(p *Lf_Bst_t) int32 { return p.Cut[1].fUsed;                                                 }
func Lf_BestCutIndex(p *Lf_Bst_t) int32 { if p.Cut[0].fUsed { return 0; }
 if p.Cut[1].fUsed { return 1; }
 return 2; }

// Original macro:
// #define Lf_CutSetForEachCut( nWords, pCutSet, pCut, i, nCuts )  for ( i = 0, pCut = pCutSet; i < nCuts; pCut = Lf_CutNext(pCut, nWords), i++ ) 

// Original macro:
// #define Lf_CutForEachVar( pCut, Var, i )                        for ( i = 0; i < (int)pCut->nLeaves && (Var = pCut->pLeaves[i]); i++ ) if ( Lf_ObjOff(p, Var) < 0 ) {} else


// extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_ObjSetCiArrival(p *Lf_Man_t, iCi int32, Time int32) {
    Vec_IntWriteEntry( &p.vCiArrivals, iCi, Time );
}
func Lf_ObjCiArrival(p *Lf_Man_t, iCi int32) int32 {
    return Vec_IntEntry( &p.vCiArrivals, iCi );
}
func Lf_ObjArrival_rec(p *Lf_Man_t, pDriver *Gia_Obj_t) int32 {
    if Gia_ObjIsBuf(pDriver) {
        return Lf_ObjArrival_rec( p, Gia_ObjFanin0(pDriver) ); }
    if Gia_ObjIsAnd(pDriver) {
        return Lf_ObjReadBest(p, Gia_ObjId(p.pGia, pDriver)).Delay[0]; }
    if Gia_ObjIsCi(pDriver) {
        return Lf_ObjCiArrival(p, Gia_ObjCioId(pDriver)); }
    return 0;
}
func Lf_ObjCoArrival(p *Lf_Man_t, iCo int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManCo(p.pGia, iCo);
    var pDriver *Gia_Obj_t  = Gia_ObjFanin0(pObj);
    return Lf_ObjArrival_rec( p, pDriver );
//    if ( Gia_ObjIsAnd(pDriver) )
//        return Lf_ObjReadBest(p, Gia_ObjId(p->pGia, pDriver))->Delay[0];
//    if ( Gia_ObjIsCi(pDriver) )
//        return Lf_ObjCiArrival(p, Gia_ObjCioId(pDriver));
//    return 0;
}
func Lf_ObjCoArrival2_rec(p *Lf_Man_t, pDriver *Gia_Obj_t) int32 {
    if Gia_ObjIsBuf(pDriver) {
        return Lf_ObjCoArrival2_rec( p, Gia_ObjFanin0(pDriver) ); }
    if Gia_ObjIsAnd(pDriver) {
        var pBest *Lf_Bst_t  = Lf_ObjReadBest(p, Gia_ObjId(p.pGia, pDriver));
        var Index int32  = Lf_BestCutIndex( pBest );
        assert(Index < 2 || Gia_ObjIsMux(p.pGia, pDriver), "Index < 2 || Gia_ObjIsMux(p.pGia, pDriver)");
        return pBest.Delay[Index];
    }
    if Gia_ObjIsCi(pDriver) {
        return Lf_ObjCiArrival(p, Gia_ObjCioId(pDriver)); }
    return 0;
}
func Lf_ObjCoArrival2(p *Lf_Man_t, iCo int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManCo(p.pGia, iCo);
    var pDriver *Gia_Obj_t  = Gia_ObjFanin0(pObj);
    return Lf_ObjCoArrival2_rec( p, pDriver );
//    if ( Gia_ObjIsAnd(pDriver) )
//    {
//        Lf_Bst_t * pBest = Lf_ObjReadBest(p, Gia_ObjId(p->pGia, pDriver));
//        int Index = Lf_BestCutIndex( pBest );
//        assert( Index < 2 || Gia_ObjIsMux(p->pGia, pDriver) );
//        return pBest->Delay[Index];
//    }
//    if ( Gia_ObjIsCi(pDriver) )
//        return Lf_ObjCiArrival(p, Gia_ObjCioId(pDriver));
//    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_ManComputeCrossCut(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, nCutMax, nCutCur int32;
    assert(p.pMuxes == nil, "p.pMuxes == nil");
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        pObj.Value = 0;
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjIsAnd(Gia_ObjFanin0(pObj)) {
            Gia_ObjFanin0(pObj).Value++; }
        if Gia_ObjIsAnd(Gia_ObjFanin1(pObj)) {
            Gia_ObjFanin1(pObj).Value++; }
    }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if pObj.Value {
            nCutCur++; }
        if nCutMax < nCutCur {
            nCutMax = nCutCur; }
        if Gia_ObjIsAnd(Gia_ObjFanin0(pObj)) && --Gia_ObjFanin0(pObj).Value == 0 {
            nCutCur--; }
        if Gia_ObjIsAnd(Gia_ObjFanin1(pObj)) && --Gia_ObjFanin1(pObj).Value == 0 {
            nCutCur--; }
    }
    }
    assert(nCutCur == 0, "nCutCur == 0");
    if nCutCur != 0 {
        fmt.Printf("Cutset is not 0\n"); }
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        assert(pObj.Value == 0, "pObj.Value == 0");
    }
    fmt.Printf("CutMax = %d\n", nCutMax);
    return nCutMax;
}

/**Function*************************************************************

  Synopsis    [Detect MUX truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Lf_ManTtIsMux_s_Muxes [24]uint32 = [24]uint32{
        (~0xAAAAAAAA & ~0xCCCCCCCC) | ( 0xAAAAAAAA & ~0xF0F0F0F0),
        (~0xAAAAAAAA & ~0xCCCCCCCC) | ( 0xAAAAAAAA &  0xF0F0F0F0),
        (~0xAAAAAAAA &  0xCCCCCCCC) | ( 0xAAAAAAAA & ~0xF0F0F0F0),
        (~0xAAAAAAAA &  0xCCCCCCCC) | ( 0xAAAAAAAA &  0xF0F0F0F0),
        ( 0xAAAAAAAA & ~0xCCCCCCCC) | (~0xAAAAAAAA & ~0xF0F0F0F0),
        ( 0xAAAAAAAA & ~0xCCCCCCCC) | (~0xAAAAAAAA &  0xF0F0F0F0),
        ( 0xAAAAAAAA &  0xCCCCCCCC) | (~0xAAAAAAAA & ~0xF0F0F0F0),
        ( 0xAAAAAAAA &  0xCCCCCCCC) | (~0xAAAAAAAA &  0xF0F0F0F0),

        (~0xCCCCCCCC & ~0xAAAAAAAA) | ( 0xCCCCCCCC & ~0xF0F0F0F0),
        (~0xCCCCCCCC & ~0xAAAAAAAA) | ( 0xCCCCCCCC &  0xF0F0F0F0),
        (~0xCCCCCCCC &  0xAAAAAAAA) | ( 0xCCCCCCCC & ~0xF0F0F0F0),
        (~0xCCCCCCCC &  0xAAAAAAAA) | ( 0xCCCCCCCC &  0xF0F0F0F0),
        ( 0xCCCCCCCC & ~0xAAAAAAAA) | (~0xCCCCCCCC & ~0xF0F0F0F0),
        ( 0xCCCCCCCC & ~0xAAAAAAAA) | (~0xCCCCCCCC &  0xF0F0F0F0),
        ( 0xCCCCCCCC &  0xAAAAAAAA) | (~0xCCCCCCCC & ~0xF0F0F0F0),
        ( 0xCCCCCCCC &  0xAAAAAAAA) | (~0xCCCCCCCC &  0xF0F0F0F0),

        (~0xF0F0F0F0 & ~0xCCCCCCCC) | ( 0xF0F0F0F0 & ~0xAAAAAAAA),
        (~0xF0F0F0F0 & ~0xCCCCCCCC) | ( 0xF0F0F0F0 &  0xAAAAAAAA),
        (~0xF0F0F0F0 &  0xCCCCCCCC) | ( 0xF0F0F0F0 & ~0xAAAAAAAA),
        (~0xF0F0F0F0 &  0xCCCCCCCC) | ( 0xF0F0F0F0 &  0xAAAAAAAA),
        ( 0xF0F0F0F0 & ~0xCCCCCCCC) | (~0xF0F0F0F0 & ~0xAAAAAAAA),
        ( 0xF0F0F0F0 & ~0xCCCCCCCC) | (~0xF0F0F0F0 &  0xAAAAAAAA),
        ( 0xF0F0F0F0 &  0xCCCCCCCC) | (~0xF0F0F0F0 & ~0xAAAAAAAA),
        ( 0xF0F0F0F0 &  0xCCCCCCCC) | (~0xF0F0F0F0 &  0xAAAAAAAA),
    };
func Lf_ManTtIsMux(t word) int32 {
    
    var i int32;
    for  i = 0; i < 24; i++  {
        if (uint32(t)) == spStatic_Lf_ManTtIsMux_s_Muxes[i] {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Count the number of unique drivers and invertors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_ManAnalyzeCoDrivers(p *Gia_Man_t, pnDrivers *int32, pnInverts *int32) {
    var pObj *Gia_Obj_t;
    var i, Entry, nDrivers, nInverts int32;
    var vMarks *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    nDrivers = nInverts = 0;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        *Vec_IntEntryP( vMarks, Gia_ObjFaninId0p(p, pObj) ) |= Gia_ObjFaninC0(pObj) ? 2 : 1;
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vMarks)) && (((Entry) = Vec_IntEntry(vMarks, i)), 1); i++  {
        nDrivers += int32((Entry != 0)), nInverts += int32((Entry == 3));
    }
    Vec_IntFree( vMarks );
    *pnDrivers = nDrivers;
    *pnInverts = nInverts;
}
func Lf_ManComputeSwitching(p *Gia_Man_t, vSwitches *Vec_Flt_t) {
//    abctime clk = Abc_Clock();
    var vSwitching *Vec_Flt_t  = (Vec_Flt_t *)Gia_ManComputeSwitchProbs( p, 48, 16, 0 );
    assert(Vec_FltCap(vSwitches) == 0, "Vec_FltCap(vSwitches) == 0");
    *vSwitches = *vSwitching;
    // inlining of macro ABC_FREE
    vSwitching = nil;
//    Abc_PrintTime( 1, "Computing switching activity", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_CutCreateUnit(p *Lf_Cut_t, i int32) int32 {
    p.fLate = 0;
    p.fMux7 = 0;
    p.iFunc = 2;
    p.nLeaves = 1;
    p.pLeaves[0] = i;
    p.Sign = ((word)1) << (i & 0x3F);
    return 1;
}
func Lf_ManFetchSet(p *Lf_Man_t, i int32) *Lf_Cut_t {
    var uMaskPage int32  = (1 << LF_LOG_PAGE) - 1;
    var pObj *Gia_Obj_t  = Gia_ManObj( p.pGia, i );
    var iOffSet int32  = Vec_IntEntry( &p.vOffsets, i );
    var Entry int32  = Vec_IntEntry( &p.vCutSets, iOffSet );
    assert(Gia_ObjIsAndNotBuf(pObj), "Gia_ObjIsAndNotBuf(pObj)");
    assert(pObj.Value > 0, "pObj.Value > 0");
    if Entry == -1 { // first visit
    
        if Vec_IntSize(&p.vFreeSets) == 0 { // add new
        
            // inlining of macro ABC_CALLOC
            var pCut *Lf_Cut_t  = (Lf_Cut_t *)((word *) calloc((size_t)(p.nSetWords * (1 << LF_LOG_PAGE)), sizeof(word)));
            var uMaskShift int32  = Vec_PtrSize(&p.vMemSets) << LF_LOG_PAGE;
            Vec_PtrPush( &p.vMemSets, pCut );
            for  Entry = uMaskPage; Entry >= 0; Entry--  {
                Vec_IntPush( &p.vFreeSets, uMaskShift | Entry );
                pCut[Entry].nLeaves   = LF_NO_LEAF;
            }
        }
        Entry = Vec_IntPop( &p.vFreeSets );
        Vec_IntWriteEntry( &p.vCutSets, iOffSet, Entry );
        p.nFrontMax = Abc_MaxInt( p.nFrontMax, Entry + 1 );
    
    } else if --pObj.Value == 0 {
        Vec_IntPush( &p.vFreeSets, Entry );
        Vec_IntWriteEntry( &p.vCutSets, iOffSet, -1 );
    }
    return (Lf_Cut_t *)((word *)Vec_PtrEntry(&p.vMemSets, Entry >> LF_LOG_PAGE) + p.nSetWords * (Entry & uMaskPage));
}
var spStatic_Lf_ManPrepareSet_CutTemp [3][LF_CUT_WORDS]word;
func Lf_ManPrepareSet(p *Lf_Man_t, iObj int32, Index int32, ppCutSet **Lf_Cut_t) int32 {
    
    if Vec_IntEntry(&p.vOffsets, iObj) == -1 {
        return Lf_CutCreateUnit( (*ppCutSet = (Lf_Cut_t *)spStatic_Lf_ManPrepareSet_CutTemp[Index]), iObj ); }
    {
        var pCut *Lf_Cut_t; 
        var i int32;
        var nCutNum int32 = p.pPars.nCutNum;
        *ppCutSet = Lf_ManFetchSet(p, iObj);
        // inlining of macro Lf_CutSetForEachCut
        for  i, pCut = 0, *ppCutSet; i < nCutNum; pCut = Lf_CutNext(pCut, p.nCutWords), i++  { 
            if pCut.nLeaves == LF_NO_LEAF {
                return i; }
        }
        return i;
    }
}

/**Function*************************************************************

  Synopsis    [Cut manipulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_CutGetSign(pCut *Lf_Cut_t) word {
    var Sign word; var i int32; 
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Sign |= ((word)1) << (pCut.pLeaves[i] & 0x3F);
    }
    return Sign;
}
func Lf_CutCountBits(i word) int32 {
    i = i - ((i >> 1) & 0x5555555555555555);
    i = (i & 0x3333333333333333) + ((i >> 2) & 0x3333333333333333);
    i = ((i + (i >> 4)) & 0x0F0F0F0F0F0F0F0F);
    return (i*(0x0101010101010101))>>56;
}
func Lf_CutEqual(pCut0 *Lf_Cut_t, pCut1 *Lf_Cut_t) int32 {
    var i int32;
    if pCut0.iFunc != pCut1.iFunc {
        return 0; }
    if pCut0.nLeaves != pCut1.nLeaves {
        return 0; }
    for  i = 0; i < int32(pCut0.nLeaves); i++  {
        if pCut0.pLeaves[i] != pCut1.pLeaves[i] {
            return 0; }
    }
    return 1;
}
func Lf_CutSwitches(p *Lf_Man_t, pCut *Lf_Cut_t) float32 {
    var Switches float32; var i int32; 
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        Switches += Lf_ObjSwitches(p, pCut.pLeaves[i]);
    }
//printf( "%.2f ", Switches );
    return Switches;
}
func Lf_CutPrint(p *Lf_Man_t, pCut *Lf_Cut_t) {
    var i int32;
    var nDigits int32 = Abc_Base10Log(Gia_ManObjNum(p.pGia)); 
    fmt.Printf("%d  {", pCut.nLeaves);
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        fmt.Printf(" %*d", nDigits, pCut.pLeaves[i]);
    }
    for  ; i < int32(p.pPars.nLutSize); i++  {
        fmt.Printf(" %*s", nDigits, " ");
    }
    fmt.Printf("  }   Late = %d  D = %4d  A = %9.4f  F = %6d\n", pCut.fLate, pCut.Delay, pCut.Flow, pCut.iFunc);
}
func Lf_CutArea(p *Lf_Man_t, pCut *Lf_Cut_t) float32 {
    if pCut.nLeaves < 2 || pCut.fMux7 {
        return 0; }
    if p.pPars.fPower {
        return 1.0 * pCut.nLeaves + Lf_CutSwitches( p, pCut ); }
    if p.pPars.fOptEdge {
        return (pCut.nLeaves + p.pPars.nAreaTuner) * (1 + (p.pPars.fCutGroup && int32(pCut.nLeaves) > p.pPars.nLutSize/2)); }
    return 1 + (p.pPars.fCutGroup && int32(pCut.nLeaves) > p.pPars.nLutSize/2);
}
func Lf_CutIsMux(p *Lf_Man_t, pCut *Lf_Cut_t, pMux *Gia_Obj_t) int32 {
    var i, Id int32;
    if pCut.nLeaves != 3 {
        return 0; }
    assert(Gia_ObjIsMux(p.pGia, pMux), "Gia_ObjIsMux(p.pGia, pMux)");
    if Gia_ObjIsCi(Gia_ObjFanin0(pMux)) || Gia_ObjIsCi(Gia_ObjFanin1(pMux)) {
        return 0; }
    Id = Gia_ObjFaninId0p( p.pGia, pMux );
    for  i = 0; i < 3; i++  {
        if pCut.pLeaves[i] == Id {
            break; }
    }
    if i == 3 { 
        return 0; }
    Id = Gia_ObjFaninId1p( p.pGia, pMux );
    for  i = 0; i < 3; i++  {
        if pCut.pLeaves[i] == Id {
            break; }
    }
    if i == 3 { 
        return 0; }
    Id = Gia_ObjFaninId2p( p.pGia, pMux );
    for  i = 0; i < 3; i++  {
        if pCut.pLeaves[i] == Id {
            break; }
    }
    if i == 3 { 
        return 0; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Cut packing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_MemAlloc(p *Lf_Mem_t, LogPage int32, vFree *Vec_Ptr_t, nCutWords int32) {
    memset( p, 0, sizeof(Lf_Mem_t) );
    p.LogPage   = LogPage;
    p.MaskPage  = (1 << LogPage) - 1;
    p.nCutWords = nCutWords;
    p.vFree     = vFree;
}
func Lf_MemSaveCut(p *Lf_Mem_t, pCut *Lf_Cut_t, iObj int32) int32 {
    unsigned char * pPlace;
    var i int32;
    var iPlace int32;
    var Prev int32 = iObj;
    var iCur int32 = p.iCur;
    assert(!pCut.fMux7, "!pCut.fMux7");
    if Vec_PtrSize(&p.vPages) == (p.iCur >> p.LogPage) {
        // inlining of macro ABC_ALLOC
        Vec_PtrPush( &p.vPages, Vec_PtrSize(p.vFree) ? Vec_PtrPop(p.vFree) : ((char *) malloc(sizeof(char) * (size_t)(p.MaskPage+1))) ); }
    assert(p.MaskPage - (p.iCur & p.MaskPage) >= 4 * (LF_LEAF_MAX + 2), "p.MaskPage - (p.iCur & p.MaskPage) >= 4 * (LF_LEAF_MAX + 2)");
    iPlace = iCur & p.MaskPage;
    pPlace = (unsigned char *)Vec_PtrEntry(&p.vPages, p.iCur >> p.LogPage);
    iPlace = Gia_AigerWriteUnsignedBuffer( pPlace, iPlace, pCut.nLeaves );
    for  i = pCut.nLeaves - 1; i >= 0; i--  {
        iPlace = Gia_AigerWriteUnsignedBuffer( pPlace, iPlace, Prev - pCut.pLeaves[i] );
Prev = pCut.pLeaves[i];
    }
    assert(pCut.nLeaves >= 2 || pCut.iFunc <= 3, "pCut.nLeaves >= 2 || pCut.iFunc <= 3");
    if pCut.iFunc >= 0 {
        iPlace = Gia_AigerWriteUnsignedBuffer( pPlace, iPlace, pCut.iFunc ); }
    if p.MaskPage - (iPlace & p.MaskPage) < 4 * (LF_LEAF_MAX + 2) {
        p.iCur = ((p.iCur >> p.LogPage) + 1) << p.LogPage;
    } else {
        p.iCur = (p.iCur & ~p.MaskPage) | iPlace; }
    return iCur;
}
func Lf_MemLoadCut(p *Lf_Mem_t, iCur int32, iObj int32, pCut *Lf_Cut_t, fTruth int32, fRecycle int32) *Lf_Cut_t {
    unsigned char * pPlace;  
    var i int32;
    var Prev int32 = iObj;
    var Page int32 = iCur >> p.LogPage;
    assert(Page < Vec_PtrSize(&p.vPages), "Page < Vec_PtrSize(&p.vPages)");
    pPlace = (unsigned char *)Vec_PtrEntry(&p.vPages, Page) + (iCur & p.MaskPage);
    pCut.nLeaves = Gia_AigerReadUnsigned(&pPlace);
    assert(pCut.nLeaves <= LF_LEAF_MAX, "pCut.nLeaves <= LF_LEAF_MAX");
    for  i = pCut.nLeaves - 1; i >= 0; i--  {
        pCut.pLeaves[i] = Prev - Gia_AigerReadUnsigned(&pPlace), Prev = pCut.pLeaves[i];
    }
    if fTruth != 0 {
        pCut.iFunc = Gia_AigerReadUnsigned(&pPlace);
    } else {
        pCut.iFunc = -1;
    }
    assert(pCut.nLeaves >= 2 || pCut.iFunc <= 3, "pCut.nLeaves >= 2 || pCut.iFunc <= 3");
    if (fRecycle != 0) && (Page != 0) && Vec_PtrEntry(&p.vPages, Page-1) {
        Vec_PtrPush( p.vFree, Vec_PtrEntry(&p.vPages, Page-1) );
        Vec_PtrWriteEntry( &p.vPages, Page-1, nil );
    }
    if fRecycle != 0 {
        pCut.Sign = Lf_CutGetSign(pCut);
    } else {
        pCut.Sign = 0;
    }
    pCut.fMux7 = 0;
    return pCut;
}
func Lf_MemRecycle(p *Lf_Mem_t) {
    var pPlace any; var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&p.vPages)) && (((pPlace) = (void *)Vec_PtrEntry(&p.vPages, i)), 1); i++  {
        if pPlace {
            Vec_PtrPush( p.vFree, pPlace ); }
    }
    Vec_PtrClear( &p.vPages );
    p.iCur = 0;
}
func Lf_MemLoadMuxCut(p *Lf_Man_t, iObj int32, pCut *Lf_Cut_t) *Lf_Cut_t {
    var pMux *Gia_Obj_t  = Gia_ManObj( p.pGia, iObj );
    assert(Gia_ObjIsMux(p.pGia, pMux), "Gia_ObjIsMux(p.pGia, pMux)");
    if p.pPars.fCutMin {
        pCut.iFunc = 4;
    } else {
        pCut.iFunc = -1;
    }
    pCut.pLeaves[0] = Gia_ObjFaninId0( pMux, iObj );
    pCut.pLeaves[1] = Gia_ObjFaninId1( pMux, iObj );
    pCut.pLeaves[2] = Gia_ObjFaninId2( p.pGia, iObj );
    pCut.nLeaves = 3;
    pCut.fMux7 = 1;
    return pCut;
}
var spStatic_Lf_ObjCutMux_CutSet [LF_CUT_WORDS]word;
func Lf_ObjCutMux(p *Lf_Man_t, i int32) *Lf_Cut_t {
    
    return Lf_MemLoadMuxCut( p, i, (Lf_Cut_t *)spStatic_Lf_ObjCutMux_CutSet );
}
var spStatic_Lf_ObjCutBest_CutSet [LF_CUT_WORDS]word;
func Lf_ObjCutBest(p *Lf_Man_t, i int32) *Lf_Cut_t {
    
    var pBest *Lf_Bst_t  = Lf_ObjReadBest( p, i );
    var pCut *Lf_Cut_t  = (Lf_Cut_t *)spStatic_Lf_ObjCutBest_CutSet;
    var Index int32  = Lf_BestCutIndex( pBest );
    pCut.Delay = pBest.Delay[Index];
    pCut.Flow  = pBest.Flow[Index];
    if Index == 2 {
        return Lf_MemLoadMuxCut( p, i, pCut ); }
    return Lf_MemLoadCut( &p.vStoreOld, pBest.Cut[Index].Handle, i, pCut, p.pPars.fCutMin, 0 );
}
func Lf_ObjCutBestNew(p *Lf_Man_t, i int32, pCut *Lf_Cut_t) *Lf_Cut_t {
    var pBest *Lf_Bst_t  = Lf_ObjReadBest( p, i );
    var Index int32  = Lf_BestCutIndex( pBest );
    pCut.Delay = pBest.Delay[Index];
    pCut.Flow  = pBest.Flow[Index];
    if Index == 2 {
        return Lf_MemLoadMuxCut( p, i, pCut ); }
    return Lf_MemLoadCut( &p.vStoreNew, pBest.Cut[Index].Handle, i, pCut, 0, 0 );
}

/**Function*************************************************************

  Synopsis    [Check correctness of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
static inline int Lf_CutCheck( Lf_Cut_t * pBase, Lf_Cut_t * pCut ) // check if pCut is contained in pBase
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
func Lf_SetCheckArray(ppCuts **Lf_Cut_t, nCuts int32) int32 {
    var pCut0, pCut1 *Lf_Cut_t; 
    var i, k, m, n, Value int32;
    assert(nCuts > 0, "nCuts > 0");
    for  i = 0; i < nCuts; i++  {
        pCut0 = ppCuts[i];
        assert(!pCut0.fMux7, "!pCut0.fMux7");
        assert(pCut0.nLeaves < LF_LEAF_MAX, "pCut0.nLeaves < LF_LEAF_MAX");
        assert(pCut0.Sign == Lf_CutGetSign(pCut0), "pCut0.Sign == Lf_CutGetSign(pCut0)");
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
            Value = Lf_CutCheck( pCut0, pCut1 );
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
func Lf_CutMergeOrder(pCut0 *Lf_Cut_t, pCut1 *Lf_Cut_t, pCut *Lf_Cut_t, nLutSize int32) int32 { 
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
        pCut.iFunc = -1;
        pCut.Sign = pCut0.Sign | pCut1.Sign;
        return 1;
    }
    // compare two cuts with different numbers
    i = k = c = 0;
    if nSize0 == 0 { goto FlushCut1; }
    if nSize1 == 0 { goto FlushCut0; }
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
    pCut.iFunc = -1;
    pCut.fMux7 = 0;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    return 1;

FlushCut1:
    if c + nSize1 > nLutSize + k { return 0; }
    for  k < nSize1  {
        pC[c++] = pC1[k++];
    }
    pCut.nLeaves = c;
    pCut.iFunc = -1;
    pCut.fMux7 = 0;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    return 1;
}
func Lf_CutMergeOrder2(pCut0 *Lf_Cut_t, pCut1 *Lf_Cut_t, pCut *Lf_Cut_t, nLutSize int32) int32 { 
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
        if c == nLutSize { return 0; }
        pC[c++] = xMin;
        if x0 == xMin { i0++; }
        if x1 == xMin { i1++; }
    }
    pCut.nLeaves = c;
    pCut.iFunc = -1;
    pCut.fMux7 = 0;
    pCut.Sign = pCut0.Sign | pCut1.Sign;
    return 1;
}
func Lf_CutMergeOrderMux(pCut0 *Lf_Cut_t, pCut1 *Lf_Cut_t, pCut2 *Lf_Cut_t, pCut *Lf_Cut_t, nLutSize int32) int32 { 
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
    pCut.iFunc = -1;
    pCut.fMux7 = 0;
    pCut.Sign = pCut0.Sign | pCut1.Sign | pCut2.Sign;
    return 1;
}

static inline int Lf_SetCutIsContainedOrder( Lf_Cut_t * pBase, Lf_Cut_t * pCut ) // check if pCut is contained in pBase
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
func Lf_SetLastCutIsContained(pCuts **Lf_Cut_t, nCuts int32) int32 {
    var i int32;
    for  i = 0; i < nCuts; i++  {
        if pCuts[i].nLeaves <= pCuts[nCuts].nLeaves && (pCuts[i].Sign & pCuts[nCuts].Sign) == pCuts[i].Sign && Lf_SetCutIsContainedOrder(pCuts[nCuts], pCuts[i]) {
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
func Lf_CutCompareDelay(pCut0 *Lf_Cut_t, pCut1 *Lf_Cut_t) int32 {
    if pCut0.Delay   < pCut1.Delay {  return -1; }
    if pCut0.Delay   > pCut1.Delay {  return  1; }
    if pCut0.nLeaves < pCut1.nLeaves {  return -1; }
    if pCut0.nLeaves > pCut1.nLeaves {  return  1; }
    if pCut0.Flow    < pCut1.Flow - LF_EPSILON {  return -1; }
    if pCut0.Flow    > pCut1.Flow + LF_EPSILON {  return  1; }
    return 0;
}
func Lf_CutCompareArea(pCut0 *Lf_Cut_t, pCut1 *Lf_Cut_t) int32 {
    if pCut0.fLate   < pCut1.fLate {  return -1; }
    if pCut0.fLate   > pCut1.fLate {  return  1; }
    if pCut0.Flow    < pCut1.Flow - LF_EPSILON {  return -1; }
    if pCut0.Flow    > pCut1.Flow + LF_EPSILON {  return  1; }
    if pCut0.Delay   < pCut1.Delay {  return -1; }
    if pCut0.Delay   > pCut1.Delay {  return  1; }
    if pCut0.nLeaves < pCut1.nLeaves {  return -1; }
    if pCut0.nLeaves > pCut1.nLeaves {  return  1; }
    return 0;
}
func Lf_SetLastCutContainsArea(pCuts **Lf_Cut_t, nCuts int32) int32 {
    var i, k, fChanges int32;
    for  i = 1; i < nCuts; i++  {
        if pCuts[nCuts].nLeaves < pCuts[i].nLeaves && (pCuts[nCuts].Sign & pCuts[i].Sign) == pCuts[nCuts].Sign && Lf_SetCutIsContainedOrder(pCuts[i], pCuts[nCuts]) {
            pCuts[i].nLeaves = LF_NO_LEAF, fChanges = 1; }
    }
    if !(fChanges != 0) {
        return nCuts; }
    for  i = k = 1; i <= nCuts; i++  {
        if pCuts[i].nLeaves == LF_NO_LEAF {
            continue; }
        if k < i {
            // inlining of macro ABC_SWAP
             var t *Lf_Cut_t  = pCuts[k]; pCuts[k] = pCuts[i]; pCuts[i] = t; };
        k++;
    }
    return k - 1;
}
func Lf_SetSortByArea(pCuts **Lf_Cut_t, nCuts int32) {
    var i int32;
    for  i = nCuts; i > 1; i--  {
        if Lf_CutCompareArea(pCuts[i - 1], pCuts[i]) < 0 {//!= 1 )
            return; }
        // inlining of macro ABC_SWAP
        { var t *Lf_Cut_t  = pCuts[i - 1]; pCuts[i - 1] = pCuts[i]; pCuts[i] = t; };
    }
}
func Lf_SetAddCut(pCuts **Lf_Cut_t, nCuts int32, nCutNum int32) int32 {
    if nCuts == 0 {
        return 1; }
    nCuts = Lf_SetLastCutContainsArea(pCuts, nCuts);
    assert(nCuts >= 1, "nCuts >= 1");
    if Lf_CutCompareDelay(pCuts[0], pCuts[nCuts]) == 1 { // new cut is better for delay
    
        // inlining of macro ABC_SWAP
        { var t *Lf_Cut_t  = pCuts[0]; pCuts[0] = pCuts[nCuts]; pCuts[nCuts] = t; };
        // if old cut (now cut number nCuts) is contained - remove it
        if pCuts[0].nLeaves < pCuts[nCuts].nLeaves && (pCuts[0].Sign & pCuts[nCuts].Sign) == pCuts[0].Sign && Lf_SetCutIsContainedOrder(pCuts[nCuts], pCuts[0]) {
            return nCuts; }
    }
    // sort area cuts by area
    Lf_SetSortByArea( pCuts, nCuts );
    // add new cut if there is room
    return Abc_MinInt( nCuts + 1, nCutNum - 1 );
}
func Lf_SetSortBySize(pCutsR **Lf_Cut_t, nCutsR int32) {
    var i, j, best_i int32;
    for  i = 1; i < nCutsR-1; i++  {
        best_i = i;
        for  j = i+1; j < nCutsR; j++  {
            if pCutsR[j].nLeaves > pCutsR[best_i].nLeaves {
                best_i = j; }
        }
        // inlining of macro ABC_SWAP
        { var t *Lf_Cut_t  = pCutsR[i]; pCutsR[i] = pCutsR[best_i]; pCutsR[best_i] = t; };
    }
}

/**Function*************************************************************

  Synopsis    [Check if truth table has non-const-cof cofactoring variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_ManFindCofVar(pTruth *word, nWords int32, nVars int32) int32 {
    var uTruthCof [LF_TT_WORDS]word; var iVar int32;
    for  iVar = 0; iVar < nVars; iVar++  {
        Abc_TtCofactor0p( uTruthCof, pTruth, nWords, iVar );
        if Abc_TtSupportSize(uTruthCof, nVars) < 2 {
            continue; }
        Abc_TtCofactor1p( uTruthCof, pTruth, nWords, iVar );
        if Abc_TtSupportSize(uTruthCof, nVars) < 2 {
            continue; }
        return iVar;
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_CutComputeTruth6(p *Lf_Man_t, pCut0 *Lf_Cut_t, pCut1 *Lf_Cut_t, fCompl0 int32, fCompl1 int32, pCutR *Lf_Cut_t, fIsXor int32) int32 {
//    extern int Mf_ManTruthCanonicize( word * t, int nVars );
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var fCompl int32; var t word;
    var t0 word  = *Lf_CutTruth(p, pCut0);
    var t1 word  = *Lf_CutTruth(p, pCut1);
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
//    p->nCutMux += Lf_ManTtIsMux( t );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
//    Mf_ManTruthCanonicize( &t, pCutR->nLeaves );
    return int32(pCutR.nLeaves) < nOldSupp;
}
func Lf_CutComputeTruth(p *Lf_Man_t, pCut0 *Lf_Cut_t, pCut1 *Lf_Cut_t, fCompl0 int32, fCompl1 int32, pCutR *Lf_Cut_t, fIsXor int32) int32 {
    if p.pPars.nLutSize <= 6 {
        return Lf_CutComputeTruth6( p, pCut0, pCut1, fCompl0, fCompl1, pCutR, fIsXor ); }
    {
    var uTruth, uTruth0, uTruth1 [LF_TT_WORDS]word;
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var LutSize int32 = p.pPars.nLutSize;
    var fCompl int32;
    var nWords int32      = Abc_Truth6WordNum(LutSize);
    var pTruth0 *word  = Lf_CutTruth(p, pCut0);
    var pTruth1 *word  = Lf_CutTruth(p, pCut1);
    Abc_TtCopy( uTruth0, pTruth0, nWords, Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 );
    Abc_TtCopy( uTruth1, pTruth1, nWords, Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 );
    Abc_TtExpand( uTruth0, LutSize, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    Abc_TtExpand( uTruth1, LutSize, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    if fIsXor != 0 {
        Abc_TtXor( uTruth, uTruth0, uTruth1, nWords, (fCompl = int32(((uTruth0[0] ^ uTruth1[0]) & 1))) );
    } else {
        Abc_TtAnd( uTruth, uTruth0, uTruth1, nWords, (fCompl = int32(((uTruth0[0] & uTruth1[0]) & 1))) ); }
    pCutR.nLeaves = Abc_TtMinBase( uTruth, pCutR.pLeaves, pCutR.nLeaves, LutSize );
    assert((uTruth[0] & 1) == 0, "(uTruth[0] & 1) == 0");
//Kit_DsdPrintFromTruth( uTruth, pCutR->nLeaves ), printf("\n" ), printf("\n" );
    truthId        = Vec_MemHashInsert(p.vTtMem, uTruth);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
    }
}
func Lf_CutComputeTruthMux6(p *Lf_Man_t, pCut0 *Lf_Cut_t, pCut1 *Lf_Cut_t, pCutC *Lf_Cut_t, fCompl0 int32, fCompl1 int32, fComplC int32, pCutR *Lf_Cut_t) int32 {
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var fCompl int32; var t word;
    var t0 word  = *Lf_CutTruth(p, pCut0);
    var t1 word  = *Lf_CutTruth(p, pCut1);
    var tC word  = *Lf_CutTruth(p, pCutC);
    if Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 { t0 = ~t0; }
    if Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 { t1 = ~t1; }
    if Abc_LitIsCompl(pCutC.iFunc) ^ fComplC { tC = ~tC; }
    t0 = Abc_Tt6Expand( t0, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    t1 = Abc_Tt6Expand( t1, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    tC = Abc_Tt6Expand( tC, pCutC.pLeaves, pCutC.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    t = (tC & t1) | (~tC & t0);
    if fCompl = int32((t & 1)) { t = ~t; }
    pCutR.nLeaves = Abc_Tt6MinBase( &t, pCutR.pLeaves, pCutR.nLeaves );
    assert(int32((t & 1)) == 0, "int32((t & 1)) == 0");
    truthId        = Vec_MemHashInsert(p.vTtMem, &t);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
}
func Lf_CutComputeTruthMux(p *Lf_Man_t, pCut0 *Lf_Cut_t, pCut1 *Lf_Cut_t, pCutC *Lf_Cut_t, fCompl0 int32, fCompl1 int32, fComplC int32, pCutR *Lf_Cut_t) int32 {
    if p.pPars.nLutSize <= 6 {
        return Lf_CutComputeTruthMux6( p, pCut0, pCut1, pCutC, fCompl0, fCompl1, fComplC, pCutR ); }
    {
    var uTruth, uTruth0, uTruth1, uTruthC [LF_TT_WORDS]word;
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var LutSize int32 = p.pPars.nLutSize;
    var fCompl int32;
    var nWords int32      = Abc_Truth6WordNum(LutSize);
    var pTruth0 *word  = Lf_CutTruth(p, pCut0);
    var pTruth1 *word  = Lf_CutTruth(p, pCut1);
    var pTruthC *word  = Lf_CutTruth(p, pCutC);
    Abc_TtCopy( uTruth0, pTruth0, nWords, Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 );
    Abc_TtCopy( uTruth1, pTruth1, nWords, Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 );
    Abc_TtCopy( uTruthC, pTruthC, nWords, Abc_LitIsCompl(pCutC.iFunc) ^ fComplC );
    Abc_TtExpand( uTruth0, LutSize, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    Abc_TtExpand( uTruth1, LutSize, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    Abc_TtExpand( uTruthC, LutSize, pCutC.pLeaves, pCutC.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    Abc_TtMux( uTruth, uTruthC, uTruth1, uTruth0, nWords );
    fCompl         = int32((uTruth[0] & 1));
    if fCompl != 0 { Abc_TtNot( uTruth, nWords ); }
    pCutR.nLeaves = Abc_TtMinBase( uTruth, pCutR.pLeaves, pCutR.nLeaves, LutSize );
    assert((uTruth[0] & 1) == 0, "(uTruth[0] & 1) == 0");
    truthId        = Vec_MemHashInsert(p.vTtMem, uTruth);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
    }
}

/**Function*************************************************************

  Synopsis    [Exact local area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_CutRef_rec(p *Lf_Man_t, pCut *Lf_Cut_t) float32 {
    var CutTemp [LF_CUT_WORDS]word ;
    var Count float32  = Lf_CutArea(p, pCut);
    var i, Var int32; 
    // inlining of macro Lf_CutForEachVar
    for  i = 0; i < int32(pCut.nLeaves) && (Var = pCut.pLeaves[i]); i++  { if Lf_ObjOff(p, Var) < 0 {} else
        if !(Lf_ObjMapRefInc(p, Var) != 0) {
            Count += Lf_CutRef_rec( p, Lf_ObjCutBestNew(p, Var, (Lf_Cut_t *)CutTemp) ); }
    }
    return Count;
}
func Lf_CutDeref_rec(p *Lf_Man_t, pCut *Lf_Cut_t) float32 {
    var CutTemp [LF_CUT_WORDS]word ;
    var Count float32  = Lf_CutArea(p, pCut);
    var i, Var int32; 
    // inlining of macro Lf_CutForEachVar
    for  i = 0; i < int32(pCut.nLeaves) && (Var = pCut.pLeaves[i]); i++  { if Lf_ObjOff(p, Var) < 0 {} else
        if !(Lf_ObjMapRefDec(p, Var) != 0) {
            Count += Lf_CutDeref_rec( p, Lf_ObjCutBestNew(p, Var, (Lf_Cut_t *)CutTemp) ); }
    }
    return Count;
}
func Lf_CutAreaDerefed(p *Lf_Man_t, pCut *Lf_Cut_t) float32 {
    var Ela1 float32  = Lf_CutRef_rec( p, pCut );
    Lf_CutDeref_rec( p, pCut );
//    float Ela2 = Lf_CutDeref_rec( p, pCut );
//    assert( Ela1 == Ela2 );
    return Ela1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_CutRequired(p *Lf_Man_t, pCut *Lf_Cut_t) int32 {
    var i, Arr, Req, Arrival, Required int32;
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        if Lf_ObjOff(p, pCut.pLeaves[i]) < 0 {
//            Arr = Lf_ObjCiArrival( p, Gia_ObjCioId(Gia_ManObj(p->pGia, pCut->pLeaves[i])) ); 
            Arr = Lf_ObjArrival_rec( p, Gia_ManObj(p.pGia, pCut.pLeaves[i]) );
        } else {
            Arr = Lf_ObjReadBest(p, pCut.pLeaves[i]).Delay[0]; }
        Arrival = Abc_MaxInt( Arrival, Arr );
        Req = Lf_ObjRequired(p, pCut.pLeaves[i]);
        if Req < ABC_INFINITY {
            Required = Abc_MaxInt( Required, Req ); }
    }
    return Abc_MaxInt( Required + 2, Arrival + 1 ); 
}
func Lf_CutParams(p *Lf_Man_t, pCut *Lf_Cut_t, Required int32, FlowRefs float32, pMux *Gia_Obj_t) {
    var pBest *Lf_Bst_t;
    var i, Index, Delay int32;
    assert(!pCut.fMux7 || Gia_ObjIsMux(p.pGia, pMux), "!pCut.fMux7 || Gia_ObjIsMux(p.pGia, pMux)");
    pCut.fLate = 0;
    pCut.Delay = 0;
    pCut.Flow  = 0;
    assert(pCut.nLeaves < LF_NO_LEAF, "pCut.nLeaves < LF_NO_LEAF");
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        if Lf_ObjOff(p, pCut.pLeaves[i]) < 0 {
//            Delay = Lf_ObjCiArrival( p, Gia_ObjCioId(Gia_ManObj(p->pGia, pCut->pLeaves[i])) );   
            Delay = Lf_ObjArrival_rec( p, Gia_ManObj(p.pGia, pCut.pLeaves[i]) );
        } else {
            pBest = Lf_ObjReadBest(p, pCut.pLeaves[i]);
            assert(pBest.Delay[0] <= pBest.Delay[1], "pBest.Delay[0] <= pBest.Delay[1]");
            assert(pBest.Flow[0]  >= pBest.Flow[1], "pBest.Flow[0]  >= pBest.Flow[1]");
            if p.fUseEla {
                Index = Lf_BestIndex(pBest);
            } else {
                Index = int32((pBest.Delay[1] + 1 <= Required && Required != ABC_INFINITY));
                //pCut->Flow += pBest->Flow[Index];
                if pCut.Flow >= float32(1e32) || pBest.Flow[Index] >= float32(1e32) {
                    pCut.Flow = float32(1e32);
                } else {
                    pCut.Flow += pBest.Flow[Index];
                    if pCut.Flow > float32(1e32) {
                         pCut.Flow = float32(1e32); }
                }
            }
            Delay = pBest.Delay[Index];
        }
//        if ( pCut->fMux7 && pCut->pLeaves[i] == Gia_ObjFaninId2p(p->pGia, pMux) )
//            Delay += 1;
        pCut.Delay = Abc_MaxInt( pCut.Delay, Delay );
    }
    pCut.Delay += int32((pCut.nLeaves > 1));// && !pCut->fMux7;
    if pCut.Delay > Required {
        pCut.fLate = 1; }
    if p.fUseEla {
        pCut.Flow = Lf_CutAreaDerefed(p, pCut) / FlowRefs;
    } else {
        pCut.Flow = (pCut.Flow + Lf_CutArea(p, pCut)) / FlowRefs;
        if pCut.Flow > float32(1e32) {
             pCut.Flow = float32(1e32); }
    }
}

func Lf_ObjMergeOrder(p *Lf_Man_t, iObj int32) {
    var CutSet [LF_CUT_MAX][LF_CUT_WORDS]word = [LF_CUT_MAX][LF_CUT_WORDS]word{{0}};
    var pCutSet0, pCutSet1, pCutSet2, pCut0, pCut1, pCut2 *Lf_Cut_t;
    var pCutSet *Lf_Cut_t = (Lf_Cut_t *)CutSet;
    var pCutsR [LF_CUT_MAX]*Lf_Cut_t;
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    var pBest *Lf_Bst_t  = Lf_ObjReadBest(p, iObj);
    var FlowRefs float32  = Lf_ObjFlowRefs(p, iObj);
    var Required int32    = Lf_ObjRequired(p, iObj);
    var nLutSize int32;
    if p.pPars.fCutGroup {
        nLutSize = p.pPars.nLutSize/2;
    } else {
        nLutSize = p.pPars.nLutSize;
    }
    var nCutNum int32     = p.pPars.nCutNum;
    var nCutWords int32   = p.nCutWords;
    var fComp0 int32      = Gia_ObjFaninC0(pObj);
    var fComp1 int32      = Gia_ObjFaninC1(pObj);
    var nCuts0 int32      = Lf_ManPrepareSet( p, Gia_ObjFaninId0(pObj, iObj), 0, &pCutSet0 );
    var nCuts1 int32      = Lf_ManPrepareSet( p, Gia_ObjFaninId1(pObj, iObj), 1, &pCutSet1 );
    var iSibl int32       = Gia_ObjSibl(p.pGia, iObj);
    var i, k, n, iCutUsed, nCutsR int32;
    var Value1 float32 = -1;
    var Value2 float32 = -1;
    assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
    // inlining of macro Lf_CutSetForEachCut
    for  i, pCut0 = 0, pCutSet; i < nCutNum; pCut0 = Lf_CutNext(pCut0, nCutWords), i++  { 
        pCutsR[i] = pCut0;
    }
    if p.Iter {
        assert(nCutsR == 0, "nCutsR == 0");
        // load cuts
        Lf_MemLoadCut( &p.vStoreOld, pBest.Cut[0].Handle, iObj, pCutsR[0], p.pPars.fCutMin, 1 );
        if Lf_BestDiffCuts(pBest) != 0 {
            Lf_MemLoadCut( &p.vStoreOld, pBest.Cut[1].Handle, iObj, pCutsR[1], p.pPars.fCutMin, 1 ); }
        // deref the cut
        if p.fUseEla && Lf_ObjMapRefNum(p, iObj) > 0 {
            Value1 = Lf_CutDeref_rec( p, pCutsR[Lf_BestIndex(pBest)] ); }
        // update required times
        if Required == ABC_INFINITY {//&& !p->fUseEla )
            Required = Lf_CutRequired( p, pCutsR[0] ); }
        // compute parameters
        Lf_CutParams( p, pCutsR[nCutsR++], Required, FlowRefs, pObj );
        if Lf_BestDiffCuts(pBest) != 0 {
            assert(nCutsR == 1, "nCutsR == 1");
            Lf_CutParams( p, pCutsR[nCutsR], Required, FlowRefs, pObj );
            nCutsR = Lf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
        if pCutsR[0].fLate {
            p.nTimeFails++; }
    }
    if iSibl != 0 {
        var pObjE *Gia_Obj_t  = Gia_ObjSiblObj(p.pGia, iObj);
        var fCompE int32  = Gia_ObjPhase(pObj) ^ Gia_ObjPhase(pObjE);
        var nCutsE int32  = Lf_ManPrepareSet( p, iSibl, 2, &pCutSet2 );
        // inlining of macro Lf_CutSetForEachCut
        for  n, pCut2 = 0, pCutSet2; n < nCutsE; pCut2 = Lf_CutNext(pCut2, nCutWords), n++  {
            if pCut2.pLeaves[0] == iSibl {
                continue; }
            Lf_CutCopy( pCutsR[nCutsR], pCut2, nCutWords );
            if pCutsR[nCutsR].iFunc >= 0 {
                pCutsR[nCutsR].iFunc = Abc_LitNotCond( pCutsR[nCutsR].iFunc, fCompE ); }
            Lf_CutParams( p, pCutsR[nCutsR], Required, FlowRefs, pObj );
            nCutsR = Lf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
    }
    if Gia_ObjIsMuxId(p.pGia, iObj) {
        var pCutSave *Lf_Cut_t  = nil;
        var fComp2 int32  = Gia_ObjFaninC2(p.pGia, pObj);
        var nCuts2 int32  = Lf_ManPrepareSet( p, Gia_ObjFaninId2(p.pGia, iObj), 2, &pCutSet2 );
        p.CutCount[0] += nCuts0 * nCuts1 * nCuts2;
        // inlining of macro Lf_CutSetForEachCut
        for  i, pCut0 = 0, pCutSet0; i < nCuts0; pCut0 = Lf_CutNext(pCut0, nCutWords), i++  {  if int32(pCut0.nLeaves) <= nLutSize {
        // inlining of macro Lf_CutSetForEachCut
        for  k, pCut1 = 0, pCutSet1; k < nCuts1; pCut1 = Lf_CutNext(pCut1, nCutWords), k++  {  if int32(pCut1.nLeaves) <= nLutSize {
        // inlining of macro Lf_CutSetForEachCut
        for  n, pCut2 = 0, pCutSet2; n < nCuts2; pCut2 = Lf_CutNext(pCut2, nCutWords), n++  {  if int32(pCut2.nLeaves) <= nLutSize {
            pCutSave = pCut2;
            if Lf_CutCountBits(pCut0.Sign | pCut1.Sign | pCut2.Sign) > nLutSize {
                continue; }
            p.CutCount[1]++; 
            if !(Lf_CutMergeOrderMux(pCut0, pCut1, pCut2, pCutsR[nCutsR], nLutSize) != 0) {
                continue; }
            if Lf_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
            p.CutCount[2]++;
            if p.pPars.fCutMin && (Lf_CutComputeTruthMux(p, pCut0, pCut1, pCut2, fComp0, fComp1, fComp2, pCutsR[nCutsR]) != 0) {
                pCutsR[nCutsR].Sign = Lf_CutGetSign(pCutsR[nCutsR]); }
            if p.pPars.nLutSizeMux && p.pPars.nLutSizeMux == int32(pCutsR[nCutsR].nLeaves) && 
                Lf_ManFindCofVar(Lf_CutTruth(p,pCutsR[nCutsR]), Abc_Truth6WordNum(nLutSize), pCutsR[nCutsR].nLeaves) == -1 {
                continue; }
            Lf_CutParams( p, pCutsR[nCutsR], Required, FlowRefs, pObj );
            nCutsR = Lf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
        } }
        } }
        }
        if p.pPars.fCutGroup {
            assert(pCutSave.nLeaves == 1, "pCutSave.nLeaves == 1");
            assert(pCutSave.pLeaves[0] == Gia_ObjFaninId2(p.pGia, iObj), "pCutSave.pLeaves[0] == Gia_ObjFaninId2(p.pGia, iObj)");
            // inlining of macro Lf_CutSetForEachCut
            for  i, pCut0 = 0, pCutSet0; i < nCuts0; pCut0 = Lf_CutNext(pCut0, nCutWords), i++  {  if int32(pCut0.nLeaves) <= nLutSize {
            // inlining of macro Lf_CutSetForEachCut
            for  k, pCut1 = 0, pCutSet1; k < nCuts1; pCut1 = Lf_CutNext(pCut1, nCutWords), k++  {  if int32(pCut1.nLeaves) <= nLutSize {
                assert(int32(pCut0.nLeaves) + int32(pCut1.nLeaves) + 1 <= p.pPars.nLutSize, "int32(pCut0.nLeaves) + int32(pCut1.nLeaves) + 1 <= p.pPars.nLutSize");
    //            if ( Lf_CutCountBits(pCut0->Sign | pCut1->Sign | pCutSave->Sign) > p->pPars->nLutSize )
    //                continue;
                p.CutCount[1]++; 
                if !(Lf_CutMergeOrderMux(pCut0, pCut1, pCutSave, pCutsR[nCutsR], p.pPars.nLutSize) != 0) {
                    continue; }
                if Lf_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                    continue; }
                p.CutCount[2]++;
                if p.pPars.fCutMin && (Lf_CutComputeTruthMux(p, pCut0, pCut1, pCutSave, fComp0, fComp1, fComp2, pCutsR[nCutsR]) != 0) {
                    pCutsR[nCutsR].Sign = Lf_CutGetSign(pCutsR[nCutsR]); }
    //            if ( p->pPars->nLutSizeMux && p->pPars->nLutSizeMux == (int)pCutsR[nCutsR]->nLeaves && 
    //                Lf_ManFindCofVar(Lf_CutTruth(p,pCutsR[nCutsR]), Abc_Truth6WordNum(nLutSize), pCutsR[nCutsR]->nLeaves) == -1 )
    //                continue;
                Lf_CutParams( p, pCutsR[nCutsR], Required, FlowRefs, pObj );
                nCutsR = Lf_SetAddCut( pCutsR, nCutsR, nCutNum );
            }
            } }
            }
        }
    
    } else {
        var fIsXor int32  = Gia_ObjIsXor(pObj);
        p.CutCount[0] += nCuts0 * nCuts1;
        // inlining of macro Lf_CutSetForEachCut
        for  i, pCut0 = 0, pCutSet0; i < nCuts0; pCut0 = Lf_CutNext(pCut0, nCutWords), i++  {  if int32(pCut0.nLeaves) <= nLutSize {
        // inlining of macro Lf_CutSetForEachCut
        for  k, pCut1 = 0, pCutSet1; k < nCuts1; pCut1 = Lf_CutNext(pCut1, nCutWords), k++  {  if int32(pCut1.nLeaves) <= nLutSize {
            if int32((pCut0.nLeaves + pCut1.nLeaves)) > nLutSize && Lf_CutCountBits(pCut0.Sign | pCut1.Sign) > nLutSize {
                continue; }
            p.CutCount[1]++; 
            if !(Lf_CutMergeOrder(pCut0, pCut1, pCutsR[nCutsR], nLutSize) != 0) {
                continue; }
            if Lf_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
            p.CutCount[2]++;
            if p.pPars.fCutMin && (Lf_CutComputeTruth(p, pCut0, pCut1, fComp0, fComp1, pCutsR[nCutsR], fIsXor) != 0) {
                pCutsR[nCutsR].Sign = Lf_CutGetSign(pCutsR[nCutsR]); }
            if p.pPars.nLutSizeMux && p.pPars.nLutSizeMux == int32(pCutsR[nCutsR].nLeaves) && 
                Lf_ManFindCofVar(Lf_CutTruth(p,pCutsR[nCutsR]), Abc_Truth6WordNum(nLutSize), pCutsR[nCutsR].nLeaves) == -1 {
                continue; }
            Lf_CutParams( p, pCutsR[nCutsR], Required, FlowRefs, pObj );
            nCutsR = Lf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
        } }
        }
    }
    // debug printout
    if false {
        fmt.Printf("*** Obj = %d  FlowRefs = %.2f  MapRefs = %2d  Required = %2d\n", iObj, FlowRefs, Lf_ObjMapRefNum(p, iObj), Required);
        for  i = 0; i < nCutsR; i++  {
            Lf_CutPrint( p, pCutsR[i] );
        }
        fmt.Printf("\n");
    }
    // verify
    assert(nCutsR > 0 && nCutsR < nCutNum, "nCutsR > 0 && nCutsR < nCutNum");
//    assert( Lf_SetCheckArray(pCutsR, nCutsR) );
    // delay cut
    assert(nCutsR == 1 || pCutsR[0].Delay <= pCutsR[1].Delay, "nCutsR == 1 || pCutsR[0].Delay <= pCutsR[1].Delay");
    pBest.Cut[0].fUsed = pBest.Cut[1].fUsed = 0;
    pBest.Cut[0].Handle = pBest.Cut[1].Handle = Lf_MemSaveCut(&p.vStoreNew, pCutsR[0], iObj);
    pBest.Delay[0] = pBest.Delay[1] = pCutsR[0].Delay;
    pBest.Flow[0] = pBest.Flow[1] = pCutsR[0].Flow;
    p.nCutCounts[pCutsR[0].nLeaves]++;
    p.CutCount[3] += nCutsR;
    p.nCutEqual++;
    // area cut
    iCutUsed = 0;
    if nCutsR > 1 && pCutsR[0].Flow > pCutsR[1].Flow + LF_EPSILON {//&& !pCutsR[1]->fLate ) // can remove !fLate
    
        pBest.Cut[1].Handle = Lf_MemSaveCut(&p.vStoreNew, pCutsR[1], iObj);
        pBest.Delay[1] = pCutsR[1].Delay;
        pBest.Flow[1] = pCutsR[1].Flow;
        p.nCutCounts[pCutsR[1].nLeaves]++;
        p.nCutEqual--;
        if !pCutsR[1].fLate {
            iCutUsed = 1; }
    }
    // mux cut
    if p.pPars.fUseMux7 && Gia_ObjIsMuxId(p.pGia, iObj) {
        pCut2 = Lf_ObjCutMux( p, iObj );
        Lf_CutParams( p, pCut2, Required, FlowRefs, pObj );
        pBest.Delay[2] = pCut2.Delay;
        pBest.Flow[2] = pCut2.Flow;
        // update area value of the best area cut
//        if ( !pCut2->fLate )
//            pBest->Flow[1] = Abc_MinFloat( pBest->Flow[1], pBest->Flow[2] );
    }
    // reference resulting cut
    if p.fUseEla {
        pBest.Cut[iCutUsed].fUsed = 1;
        if Lf_ObjMapRefNum(p, iObj) > 0 {
            Value2 = Lf_CutRef_rec( p, pCutsR[iCutUsed] ); }
//        if ( Value1 < Value2 )
//            printf( "ELA degradated cost at node %d from %d to %d.\n", iObj, Value1, Value2 ), fflush(stdout);
//        assert( Value1 >= Value2 );
//        if ( Value1 != -1 )
//            printf( "%.2f -> %.2f    ", Value1, Value2 );
    }
    if pObj.Value == 0 {
        return; }
    // store the cutset
    pCutSet = Lf_ManFetchSet(p, iObj);
    // inlining of macro Lf_CutSetForEachCut
    for  i, pCut0 = 0, pCutSet; i < nCutNum; pCut0 = Lf_CutNext(pCut0, nCutWords), i++  {
        assert(!pCut0.fMux7, "!pCut0.fMux7");
        if i < nCutsR {
            Lf_CutCopy( pCut0, pCutsR[i], nCutWords );
        } else if i == nCutsR && pCutsR[0].nLeaves > 1 && (nCutsR == 1 || pCutsR[1].nLeaves > 1) {
            Lf_CutCreateUnit( pCut0, iObj );
        } else {
            pCut0.nLeaves = LF_NO_LEAF; }
    }
}

/**Function*************************************************************

  Synopsis    [Computing delay/area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_ManSetFlowRefInc(p *Gia_Man_t, vRefs *Vec_Flt_t, vOffsets *Vec_Int_t, i int32) {
    if Gia_ObjIsAndNotBuf(Gia_ManObj(p, i)) {
        Vec_FltAddToEntry( vRefs, Vec_IntEntry(vOffsets, i), 1 ); }
}
func Lf_ManSetFlowRefs(p *Gia_Man_t, vRefs *Vec_Flt_t, vOffsets *Vec_Int_t) {
    var fDiscount int32  = 1;
    var pObj, pCtrl, pData0, pData1 *Gia_Obj_t; 
    var i, Id int32;
    Vec_FltFill( vRefs, Gia_ManAndNotBufNum(p), 0 );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjIsAndNotBuf(Gia_ObjFanin0(pObj)) {
            Vec_FltAddToEntry( vRefs, Vec_IntEntry(vOffsets, Gia_ObjFaninId0(pObj, i)), 1 ); }
        if Gia_ObjIsBuf(pObj) {
            continue; }
        if Gia_ObjIsAndNotBuf(Gia_ObjFanin1(pObj)) {
            Vec_FltAddToEntry( vRefs, Vec_IntEntry(vOffsets, Gia_ObjFaninId1(pObj, i)), 1 ); }
        if p.pMuxes {
            if Gia_ObjIsMuxId(p, i) && Gia_ObjIsAndNotBuf(Gia_ObjFanin2(p, pObj)) {
                Vec_FltAddToEntry( vRefs, Vec_IntEntry(vOffsets, Gia_ObjFaninId2(p, i)), 1 ); }
        
        } else if (fDiscount != 0) && Gia_ObjIsMuxType(pObj) { // discount XOR/MUX
        
            pCtrl  = Gia_Regular(Gia_ObjRecognizeMux(pObj, &pData1, &pData0));
            pData0 = Gia_Regular(pData0);
            pData1 = Gia_Regular(pData1);
            if Gia_ObjIsAndNotBuf(pCtrl) {
                Vec_FltAddToEntry( vRefs, Vec_IntEntry(vOffsets, Gia_ObjId(p, pCtrl)), -1 ); }
            if pData0 == pData1 && Gia_ObjIsAndNotBuf(pData0) {
                Vec_FltAddToEntry( vRefs, Vec_IntEntry(vOffsets, Gia_ObjId(p, pData0)), -1 ); }
        }
    }
    }
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && (((Id) = Gia_ObjFaninId0p(p, Gia_ManCo(p, i))), 1); i++  {
        if Gia_ObjIsAndNotBuf(Gia_ManObj(p, Id)) {
            Vec_FltAddToEntry( vRefs, Vec_IntEntry(vOffsets, Id), 1 ); }
    }
    for  i = 0; i < Vec_FltSize(vRefs); i++  {
        Vec_FltUpdateEntry( vRefs, i, 1 );
    }
}
func Lf_ManSetCutRefs(p *Lf_Man_t) {
    var pObj *Gia_Obj_t; var i int32;
    if Vec_PtrSize(&p.vMemSets) * (1 << LF_LOG_PAGE) != Vec_IntSize(&p.vFreeSets) {
        fmt.Printf("The number of used cutsets = %d.\n", Vec_PtrSize(&p.vMemSets) * (1 << LF_LOG_PAGE) - Vec_IntSize(&p.vFreeSets)); }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        assert(pObj.Value == 0, "pObj.Value == 0");
        if Gia_ObjIsBuf(pObj) {
            continue; }
        if Gia_ObjIsAndNotBuf(Gia_ObjFanin0(pObj)) {
            Gia_ObjFanin0(pObj).Value++; }
        if Gia_ObjIsAndNotBuf(Gia_ObjFanin1(pObj)) {
            Gia_ObjFanin1(pObj).Value++; }
        if Gia_ObjIsMuxId(p.pGia, i) && Gia_ObjIsAndNotBuf(Gia_ObjFanin2(p.pGia, pObj)) {
            Gia_ObjFanin2(p.pGia, pObj).Value++; }
        if Gia_ObjSibl(p.pGia, i) && Gia_ObjIsAndNotBuf(Gia_ObjSiblObj(p.pGia, i)) {
            Gia_ObjSiblObj(p.pGia, i).Value++; }
    }
    }
}

func Lf_ManSetMuxCut(p *Lf_Man_t, pBest *Lf_Bst_t, iObj int32, Required int32) int32 {
    var pMux *Gia_Obj_t;
    if !Gia_ObjIsMuxId(p.pGia, iObj) {
        return 0; }
    if pBest.Delay[2] > Required {
        return 0; }
    if pBest.Flow[2] > 1.1 * pBest.Flow[1] {
        return 0; }
    pMux = Gia_ManObj(p.pGia, iObj);
    if pMux.fMark0 || Gia_ObjFanin0(pMux).fMark0 || Gia_ObjFanin1(pMux).fMark0 {
        return 0; }
    Gia_ObjFanin0(pMux).fMark0 = 1;
    Gia_ObjFanin1(pMux).fMark0 = 1;
    return 1;
}
func Lf_ManSetMapRefsOne(p *Lf_Man_t, iObj int32) {
    var pCut *Lf_Cut_t;
    var pBest *Lf_Bst_t  = Lf_ObjReadBest( p, iObj );
    var k int32;
    var Index int32;
    var Required int32 = Lf_ObjRequired( p, iObj );
    assert(Lf_ObjMapRefNum(p, iObj) > 0, "Lf_ObjMapRefNum(p, iObj) > 0");
    assert(!pBest.Cut[0].fUsed && !pBest.Cut[1].fUsed, "!pBest.Cut[0].fUsed && !pBest.Cut[1].fUsed");
    if !p.pPars.fUseMux7 || !(Lf_ManSetMuxCut(p, pBest, iObj, Required) != 0) {
        Index = int32((Lf_BestDiffCuts(pBest) && pBest.Delay[1] <= Required));
        pBest.Cut[Index].fUsed = 1;
    }
    pCut = Lf_ObjCutBest( p, iObj );
    assert(!pCut.fMux7 || pCut.nLeaves == 3, "!pCut.fMux7 || pCut.nLeaves == 3");
//    assert( pCut->Delay <= Required );
    for  k = 0; k < int32(pCut.nLeaves); k++  {
//        if ( pCut->fMux7 && pCut->pLeaves[k] != Gia_ObjFaninId2(p->pGia, iObj) )
//            Lf_ObjSetRequired( p, pCut->pLeaves[k], Required );
//        else
            Lf_ObjSetRequired( p, pCut.pLeaves[k], Required - 1 );
        if Gia_ObjIsAndNotBuf(Gia_ManObj(p.pGia, pCut.pLeaves[k])) {
            Lf_ObjMapRefInc( p, pCut.pLeaves[k] ); }
    }
    if pCut.fMux7 {
        p.pPars.Mux7++;
        p.pPars.Edge++;
        return;
    }
    if Vec_FltSize(&p.vSwitches) {
        p.Switches += Lf_CutSwitches(p, pCut); }
    p.pPars.Edge += pCut.nLeaves;
    p.pPars.Area++;
}
func Lf_ManSetMapRefs(p *Lf_Man_t) int32 {
    var Coef float32  = 1.0 / (1.0 + (p.Iter + 1) * (p.Iter + 1));
    var pFlowRefs *float32; 
    var pMapRefs *int32;
    var i int32;
    var pObj *Gia_Obj_t;
    // compute delay
    var Delay int32;
    for  i = 0; i < Gia_ManCoNum(p.pGia); i++  {
        Delay = Abc_MaxInt( Delay, Lf_ObjCoArrival(p, i) );
    }
    // check delay target
    if p.pPars.DelayTarget == -1 && p.pPars.nRelaxRatio {
        p.pPars.DelayTarget = int32((float32(Delay) * (100.0 + p.pPars.nRelaxRatio) / 100.0)); }
    if p.pPars.DelayTarget != -1 {
        if Delay < p.pPars.DelayTarget + 0.01 {
            Delay = p.pPars.DelayTarget;
        } else if p.pPars.nRelaxRatio == 0 {
            Abc_Print( 0, "Relaxing user-specified delay target from %d to %d.\n", p.pPars.DelayTarget, Delay ); }
    } 
    p.pPars.Delay = Delay;
    // compute area/edges/required
    p.pPars.Mux7 = p.pPars.Area = p.pPars.Edge = p.Switches = 0;
    Vec_IntFill( &p.vMapRefs, Gia_ManAndNotBufNum(p.pGia), 0 );
    Vec_IntFill( &p.vRequired, Gia_ManObjNum(p.pGia), ABC_INFINITY );
    if p.pPars.fUseMux7 {
        Gia_ManCleanMark0(p.pGia);
        // inlining of macro Gia_ManForEachCi
        for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((pObj) = Gia_ManCi(p.pGia, i)); i++  {
            pObj.fMark0 = 1;
        }
    }
    if p.pGia.pManTime != nil {
        assert(!Gia_ManBufNum(p.pGia), "!Gia_ManBufNum(p.pGia)");
        Tim_ManIncrementTravId( (Tim_Man_t*)p.pGia.pManTime );
        if p.pPars.fDoAverage {
            for  i = 0; i < Gia_ManCoNum(p.pGia); i++  {
               Tim_ManSetCoRequired( (Tim_Man_t*)p.pGia.pManTime, i, int32((Lf_ObjCoArrival(p, i) * (100.0 + p.pPars.nRelaxRatio) / 100.0)) );
            }
        } else {  
            Tim_ManInitPoRequiredAll( (Tim_Man_t*)p.pGia.pManTime, Delay ); }
        // inlining of macro Gia_ManForEachObjReverse1
        for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {
            if Gia_ObjIsBuf(pObj) {
                Lf_ObjSetRequired( p, Gia_ObjFaninId0(pObj, i), Lf_ObjRequired(p, i) );
            } else if Gia_ObjIsAnd(pObj) {
                if Lf_ObjMapRefNum(p, i) != 0 {
                    Lf_ManSetMapRefsOne( p, i ); }
            
            } else if Gia_ObjIsCi(pObj) {
                Tim_ManSetCiRequired( (Tim_Man_t*)p.pGia.pManTime, Gia_ObjCioId(pObj), Lf_ObjRequired(p, i) );
            } else if Gia_ObjIsCo(pObj) {
                var iDriverId int32  = Gia_ObjFaninId0(pObj, i);
                var reqTime int32  = Tim_ManGetCoRequired( (Tim_Man_t*)p.pGia.pManTime, Gia_ObjCioId(pObj) );
                Lf_ObjSetRequired( p, iDriverId, reqTime );
                if Gia_ObjIsAndNotBuf(Gia_ObjFanin0(pObj)) {
                    Lf_ObjMapRefInc( p, iDriverId ); }
            
            } else { assert(false, "0"); }
        }
    
    } else {
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
            var iDriverId int32  = Gia_ObjFaninId0p(p.pGia, pObj);
            var reqTime int32;
            if p.pPars.fDoAverage {
                reqTime = int32((Lf_ObjCoArrival(p, i) * (100.0 + p.pPars.nRelaxRatio) / 100.0));
            } else {
                reqTime = Delay;
            }
            Lf_ObjSetRequired( p, iDriverId, reqTime );
            if Gia_ObjIsAndNotBuf(Gia_ObjFanin0(pObj)) {
                Lf_ObjMapRefInc( p, iDriverId ); }
        }
        // inlining of macro Gia_ManForEachAndReverse
        for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
            if Gia_ObjIsBuf(pObj) {
                Lf_ObjSetRequired( p, Gia_ObjFaninId0(pObj, i), Lf_ObjRequired(p, i) );
                if Gia_ObjIsAndNotBuf(Gia_ObjFanin0(pObj)) {
                    Lf_ObjMapRefInc( p, Gia_ObjFaninId0(pObj, i) ); }
            
            } else if Lf_ObjMapRefNum(p, i) != 0 {
                Lf_ManSetMapRefsOne( p, i ); }
        }
        }
    }
    if p.pPars.fUseMux7 {
        Gia_ManCleanMark0(p.pGia); }
    // blend references
    assert(Vec_IntSize(&p.vMapRefs)  == Gia_ManAndNotBufNum(p.pGia), "Vec_IntSize(&p.vMapRefs)  == Gia_ManAndNotBufNum(p.pGia)");
    assert(Vec_FltSize(&p.vFlowRefs) == Gia_ManAndNotBufNum(p.pGia), "Vec_FltSize(&p.vFlowRefs) == Gia_ManAndNotBufNum(p.pGia)");
    pMapRefs  = Vec_IntArray(&p.vMapRefs);
    pFlowRefs = Vec_FltArray(&p.vFlowRefs);
    for  i = 0; i < Vec_IntSize(&p.vMapRefs); i++  {
        pFlowRefs[i] = Coef * pFlowRefs[i] + (1.0 - Coef) * Abc_MaxFloat(1, pMapRefs[i]);
    }
//        pFlowRefs[i] = 0.2 * pFlowRefs[i] + 0.8 * Abc_MaxFloat(1, pMapRefs[i]);
    return p.pPars.Area;
}

func Lf_ManCountMapRefsOne(p *Lf_Man_t, iObj int32) {
    var pBest *Lf_Bst_t  = Lf_ObjReadBest( p, iObj );
    var pCut *Lf_Cut_t  = Lf_ObjCutBest( p, iObj );
    var k int32;
    var Required int32 = Lf_ObjRequired( p, iObj );
    assert(Lf_ObjMapRefNum(p, iObj) > 0, "Lf_ObjMapRefNum(p, iObj) > 0");
    assert((Lf_BestIsMapped(pBest) != 0), "Lf_BestIsMapped(pBest)");
    assert(!pCut.fMux7, "!pCut.fMux7");
//    assert( pCut->Delay <= Required );
    for  k = 0; k < int32(pCut.nLeaves); k++  {
            Lf_ObjSetRequired( p, pCut.pLeaves[k], Required - 1 );
    }
    if Vec_FltSize(&p.vSwitches) {
        p.Switches += Lf_CutSwitches(p, pCut); }
    p.pPars.Edge += pCut.nLeaves;
    p.pPars.Area++;
}
func Lf_ManCountMapRefs(p *Lf_Man_t) {
    // compute delay
    var pObj *Gia_Obj_t;
    var i, Id, Delay int32;
    for  i = 0; i < Gia_ManCoNum(p.pGia); i++  {
        Delay = Abc_MaxInt( Delay, Lf_ObjCoArrival2(p, i) );
    }
    // check delay target
    if p.pPars.DelayTarget == -1 && p.pPars.nRelaxRatio {
        p.pPars.DelayTarget = int32((float32(Delay) * (100.0 + p.pPars.nRelaxRatio) / 100.0)); }
    if p.pPars.DelayTarget != -1 {
        if Delay < p.pPars.DelayTarget + 0.01 {
            Delay = p.pPars.DelayTarget;
        } else if p.pPars.nRelaxRatio == 0 {
            Abc_Print( 0, "Relaxing user-specified delay target from %d to %d.\n", p.pPars.DelayTarget, Delay ); }
    }
    p.pPars.Delay = Delay;
    // compute area/edges/required
    p.pPars.Mux7 = p.pPars.Area = p.pPars.Edge = p.Switches = 0;
    Vec_IntFill( &p.vRequired, Gia_ManObjNum(p.pGia), ABC_INFINITY );
    if p.pPars.fUseMux7 {
        Gia_ManCleanMark0(p.pGia); }
    if p.pGia.pManTime != nil {
        Tim_ManIncrementTravId( (Tim_Man_t*)p.pGia.pManTime );
        if p.pPars.fDoAverage {
            for  i = 0; i < Gia_ManCoNum(p.pGia); i++  {
               Tim_ManSetCoRequired( (Tim_Man_t*)p.pGia.pManTime, i, int32((Lf_ObjCoArrival(p, i) * (100.0 + p.pPars.nRelaxRatio) / 100.0)) );
            }
        } else {  
            Tim_ManInitPoRequiredAll( (Tim_Man_t*)p.pGia.pManTime, Delay ); }
        // inlining of macro Gia_ManForEachObjReverse1
        for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {
            if Gia_ObjIsBuf(pObj) {
                Lf_ObjSetRequired( p, Gia_ObjFaninId0(pObj, i), Lf_ObjRequired(p, i) );
            } else if Gia_ObjIsAnd(pObj) {
                if Lf_ObjMapRefNum(p, i) != 0 {
                    Lf_ManCountMapRefsOne( p, i ); }
            
            } else if Gia_ObjIsCi(pObj) {
                Tim_ManSetCiRequired( (Tim_Man_t*)p.pGia.pManTime, Gia_ObjCioId(pObj), Lf_ObjRequired(p, i) );
            } else if Gia_ObjIsCo(pObj) {
                var reqTime int32  = Tim_ManGetCoRequired( (Tim_Man_t*)p.pGia.pManTime, Gia_ObjCioId(pObj) );
                Lf_ObjSetRequired( p, Gia_ObjFaninId0(pObj, i), reqTime );
            
            } else { assert(false, "0"); }
        }
    
    } else {
        // inlining of macro Gia_ManForEachCoDriverId
        for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && (((Id) = Gia_ObjFaninId0p(p.pGia, Gia_ManCo(p.pGia, i))), 1); i++  {
            Lf_ObjSetRequired( p, Id, p.pPars.fDoAverage ? int32((Lf_ObjCoArrival(p, i) * (100.0 + p.pPars.nRelaxRatio) / 100.0)) : Delay );
        }
        // inlining of macro Gia_ManForEachAndReverse
        for  i = p.pGia.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p.pGia, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else
            if Gia_ObjIsBuf(pObj) {
                Lf_ObjSetRequired( p, Gia_ObjFaninId0(pObj, i), Lf_ObjRequired(p, i) );
            } else if Lf_ObjMapRefNum(p, i) != 0 {
                Lf_ManCountMapRefsOne( p, i ); }
        }
    }
    if p.pPars.fUseMux7 {
        Gia_ManCleanMark0(p.pGia); }
}    


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_ManDeriveMapping(p *Lf_Man_t) *Gia_Man_t {
    var vMapping *Vec_Int_t;
    var pCut *Lf_Cut_t;
    var i, k int32;
    assert(!p.pPars.fCutMin && p.pGia.vMapping == nil, "!p.pPars.fCutMin && p.pGia.vMapping == nil");
    vMapping = Vec_IntAlloc( Gia_ManObjNum(p.pGia) + int32(p.pPars.Edge) + int32(p.pPars.Area) * 2 );
    Vec_IntFill( vMapping, Gia_ManObjNum(p.pGia), 0 );
    // inlining of macro Gia_ManForEachAndId
    for  i = 0; (i < p.pGia.nObjs); i++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p.pGia, i)) {} else {
        if !(Lf_ObjMapRefNum(p, i) != 0) {
            continue; }
        assert(!Gia_ObjIsBuf(Gia_ManObj(p.pGia,i)), "!Gia_ObjIsBuf(Gia_ManObj(p.pGia,i))");
        pCut = Lf_ObjCutBest( p, i );
        assert(!pCut.fMux7, "!pCut.fMux7");
        Vec_IntWriteEntry( vMapping, i, Vec_IntSize(vMapping) );
        Vec_IntPush( vMapping, pCut.nLeaves );
        for  k = 0; k < int32(pCut.nLeaves); k++  {
            Vec_IntPush( vMapping, pCut.pLeaves[k] );
        }
        Vec_IntPush( vMapping, i );
    }
    }
    assert(Vec_IntCap(vMapping) == 16 || Vec_IntSize(vMapping) == Vec_IntCap(vMapping), "Vec_IntCap(vMapping) == 16 || Vec_IntSize(vMapping) == Vec_IntCap(vMapping)");
    p.pGia.vMapping = vMapping;
    return p.pGia;
}
func Lf_ManDeriveMappingCoarse(p *Lf_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pGia *Gia_Man_t = p.pGia;
    var pObj *Gia_Obj_t;
    var pCut *Lf_Cut_t;
    var i, k int32;
    assert(!p.pPars.fCutMin && pGia.pMuxes, "!p.pPars.fCutMin && pGia.pMuxes");
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(pGia) );
    pNew.pName = Abc_UtilStrsav( pGia.pName );
    pNew.pSpec = Abc_UtilStrsav( pGia.pSpec );
    // start mapping
    pNew.vMapping = Vec_IntAlloc( Gia_ManObjNum(pGia) + 2*Gia_ManXorNum(pGia) + 2*Gia_ManMuxNum(pGia) + int32(p.pPars.Edge) + 2*int32(p.pPars.Area) + 4*int32(p.pPars.Mux7) );
    Vec_IntFill( pNew.vMapping, Gia_ManObjNum(pGia) + 2*Gia_ManXorNum(pGia) + 2*Gia_ManMuxNum(pGia), 0 );
    // process objects
    Gia_ManConst0(pGia).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {
        if Gia_ObjIsCi(pObj) { pObj.Value = Gia_ManAppendCi( pNew ); continue; }
        if Gia_ObjIsCo(pObj) { pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); continue; }
        if Gia_ObjIsBuf(pObj) { pObj.Value = Gia_ManAppendBuf( pNew, Gia_ObjFanin0Copy(pObj) ); continue; }
        if Gia_ObjIsMuxId(pGia, i) {
            pObj.Value = Gia_ManAppendMux( pNew, Gia_ObjFanin2Copy(pGia, pObj), Gia_ObjFanin1Copy(pObj), Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsXor(pObj) {
            pObj.Value = Gia_ManAppendXor( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else { 
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        if !(Lf_ObjMapRefNum(p, i) != 0) {
            continue; }
        pCut = Lf_ObjCutBest( p, i );
        Vec_IntWriteEntry( pNew.vMapping, Abc_Lit2Var(pObj.Value), Vec_IntSize(pNew.vMapping) );
        Vec_IntPush( pNew.vMapping, pCut.nLeaves );
        for  k = 0; k < int32(pCut.nLeaves); k++  {
            Vec_IntPush( pNew.vMapping, Abc_Lit2Var(Gia_ManObj(pGia, pCut.pLeaves[k]).Value) );
        }
        Vec_IntPush( pNew.vMapping, pCut.fMux7 ? -Abc_Lit2Var(pObj.Value) : Abc_Lit2Var(pObj.Value) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(pGia) );
    assert(Vec_IntCap(pNew.vMapping) == 16 || Vec_IntSize(pNew.vMapping) == Vec_IntCap(pNew.vMapping), "Vec_IntCap(pNew.vMapping) == 16 || Vec_IntSize(pNew.vMapping) == Vec_IntCap(pNew.vMapping)");
    return pNew;
}
func Lf_ManDerivePart(p *Lf_Man_t, pNew *Gia_Man_t, vMapping *Vec_Int_t, vMapping2 *Vec_Int_t, vCopies *Vec_Int_t, pCut *Lf_Cut_t, vLeaves *Vec_Int_t, vCover *Vec_Int_t, pObj *Gia_Obj_t) int32 {
    var pTruth *word;
    var k, iLit, iTemp int32;
    if p.pPars.nLutSizeMux && p.pPars.nLutSizeMux == int32(pCut.nLeaves) {
        var pTruthCof [LF_TT_WORDS]word;
        var pTruth *word = Lf_CutTruth( p, pCut );
        var pVarsNew [LF_LEAF_MAX]int32;
        var nVarsNew int32;
        var iLitCofs [2]int32; 
        var LutSize int32    = p.pPars.nLutSize;
        var nWords int32     = Abc_Truth6WordNum(LutSize);
        var c int32;
        var iVar int32 = Lf_ManFindCofVar( pTruth, nWords, pCut.nLeaves );
        assert(iVar >= 0 && iVar < int32(pCut.nLeaves), "iVar >= 0 && iVar < int32(pCut.nLeaves)");
        for  c = 0; c < 2; c++  {
            for  k = 0; k < int32(pCut.nLeaves); k++  {
                pVarsNew[k] = k;
            }
            if c != 0 {
                Abc_TtCofactor1p( pTruthCof, pTruth, nWords, iVar );
            } else {
                Abc_TtCofactor0p( pTruthCof, pTruth, nWords, iVar ); }
            nVarsNew = Abc_TtMinBase( pTruthCof, pVarsNew, pCut.nLeaves, LutSize );
            assert(nVarsNew > 0, "nVarsNew > 0");
            // derive LUT
            Vec_IntClear( vLeaves );
            for  k = 0; k < nVarsNew; k++  {
                Vec_IntPush( vLeaves, Vec_IntEntry(vCopies, pCut.pLeaves[pVarsNew[k]]) );
            }
            iLitCofs[c] = Kit_TruthToGia( pNew, (unsigned *)pTruthCof, nVarsNew, vCover, vLeaves, 0 );
            // create mapping
            Vec_IntSetEntry( vMapping, Abc_Lit2Var(iLitCofs[c]), Vec_IntSize(vMapping2) );
            Vec_IntPush( vMapping2, Vec_IntSize(vLeaves) );
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vLeaves)) && (((iTemp) = Vec_IntEntry(vLeaves, k)), 1); k++  {
                Vec_IntPush( vMapping2, Abc_Lit2Var(iTemp) );
            }
            Vec_IntPush( vMapping2, Abc_Lit2Var(iLitCofs[c]) );
        }
        // derive MUX
        pTruthCof[0] = ABC_CONST(0xCACACACACACACACA);
        Vec_IntClear( vLeaves );
        Vec_IntPush( vLeaves, iLitCofs[0] );
        Vec_IntPush( vLeaves, iLitCofs[1] );
        Vec_IntPush( vLeaves, Vec_IntEntry(vCopies, pCut.pLeaves[iVar]) );
        iLit = Kit_TruthToGia( pNew, (unsigned *)pTruthCof, Vec_IntSize(vLeaves), vCover, vLeaves, 0 );
        // create mapping
        Vec_IntSetEntry( vMapping, Abc_Lit2Var(iLit), Vec_IntSize(vMapping2) );
        Vec_IntPush( vMapping2, Vec_IntSize(vLeaves) );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLeaves)) && (((iTemp) = Vec_IntEntry(vLeaves, k)), 1); k++  {
            Vec_IntPush( vMapping2, Abc_Lit2Var(iTemp) );
        }
        Vec_IntPush( vMapping2, -Abc_Lit2Var(iLit) );
        return iLit;
    }
    Vec_IntClear( vLeaves );
    if pCut.fMux7 {
        assert(pCut.nLeaves == 3, "pCut.nLeaves == 3");
        Vec_IntPush( vLeaves, Abc_LitNotCond(Vec_IntEntry(vCopies, pCut.pLeaves[0]), Gia_ObjFaninC0(pObj)) );
        Vec_IntPush( vLeaves, Abc_LitNotCond(Vec_IntEntry(vCopies, pCut.pLeaves[1]), Gia_ObjFaninC1(pObj)) );
        Vec_IntPush( vLeaves, Abc_LitNotCond(Vec_IntEntry(vCopies, pCut.pLeaves[2]), Gia_ObjFaninC2(p.pGia,pObj)) );
    
    } else {
        for  k = 0; k < int32(pCut.nLeaves); k++  {
            Vec_IntPush( vLeaves, Vec_IntEntry(vCopies, pCut.pLeaves[k]) );
        }
    }
    pTruth = Lf_CutTruth( p, pCut );
    iLit = Kit_TruthToGia( pNew, (unsigned *)pTruth, Vec_IntSize(vLeaves), vCover, vLeaves, 0 );
    // create mapping
    Vec_IntSetEntry( vMapping, Abc_Lit2Var(iLit), Vec_IntSize(vMapping2) );
    Vec_IntPush( vMapping2, Vec_IntSize(vLeaves) );
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vLeaves)) && (((iTemp) = Vec_IntEntry(vLeaves, k)), 1); k++  {
        Vec_IntPush( vMapping2, Abc_Lit2Var(iTemp) );
    }
    Vec_IntPush( vMapping2, pCut.fMux7 ? -Abc_Lit2Var(iLit) : Abc_Lit2Var(iLit) );
    return iLit;
}
func Lf_ManDeriveMappingGia(p *Lf_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t; 
    var pObj *Gia_Obj_t; 
    var vCopies *Vec_Int_t    = Vec_IntStartFull( Gia_ManObjNum(p.pGia) );
    var vMapping *Vec_Int_t   = Vec_IntStart( 2*Gia_ManObjNum(p.pGia) + int32(p.pPars.Edge) + 2*int32(p.pPars.Area) + 4*int32(p.pPars.Mux7) );
    var vMapping2 *Vec_Int_t  = Vec_IntStart( int32(p.pPars.Edge) + 2*int32(p.pPars.Area) + 1000 );
    var vCover *Vec_Int_t     = Vec_IntAlloc( 1 << 16 );
    var vLeaves *Vec_Int_t    = Vec_IntAlloc( 16 );
    var pCut *Lf_Cut_t;
    var i, iLit int32; 
    assert(p.pPars.fCutMin, "p.pPars.fCutMin");
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p.pGia) );
    pNew.pName = Abc_UtilStrsav( p.pGia.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pGia.pSpec );
    Vec_IntWriteEntry( vCopies, 0, 0 );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            Vec_IntWriteEntry( vCopies, i, Gia_ManAppendCi(pNew) );
            continue;
        }
        if Gia_ObjIsCo(pObj) {
            iLit = Vec_IntEntry( vCopies, Gia_ObjFaninId0p(p.pGia, pObj) );
            iLit = Gia_ManAppendCo( pNew, Abc_LitNotCond(iLit, Gia_ObjFaninC0(pObj)) );
            continue;
        }
        if Gia_ObjIsBuf(pObj) {
            iLit = Vec_IntEntry( vCopies, Gia_ObjFaninId0p(p.pGia, pObj) );
            iLit = Gia_ManAppendBuf( pNew, Abc_LitNotCond(iLit, Gia_ObjFaninC0(pObj)) );
            Vec_IntWriteEntry( vCopies, i, iLit );
            continue;
        }
        if !(Lf_ObjMapRefNum(p, i) != 0) {
            continue; }
        pCut = Lf_ObjCutBest( p, i );
        assert(pCut.iFunc >= 0, "pCut.iFunc >= 0");
        if pCut.nLeaves == 0 {
            assert(Abc_Lit2Var(pCut.iFunc) == 0, "Abc_Lit2Var(pCut.iFunc) == 0");
            Vec_IntWriteEntry( vCopies, i, pCut.iFunc );
            continue;
        }
        if pCut.nLeaves == 1 {
            assert(Abc_Lit2Var(pCut.iFunc) == 1, "Abc_Lit2Var(pCut.iFunc) == 1");
            iLit = Vec_IntEntry( vCopies, pCut.pLeaves[0] );
            Vec_IntWriteEntry( vCopies, i, Abc_LitNotCond(iLit, Abc_LitIsCompl(pCut.iFunc)) );
            continue;
        }
        iLit = Lf_ManDerivePart( p, pNew, vMapping, vMapping2, vCopies, pCut, vLeaves, vCover, pObj );
        Vec_IntWriteEntry( vCopies, i, Abc_LitNotCond(iLit, Abc_LitIsCompl(pCut.iFunc)) );
    }
    Vec_IntFree( vCopies );
    Vec_IntFree( vCover );
    Vec_IntFree( vLeaves );
    // finish mapping 
    if Vec_IntSize(vMapping) > Gia_ManObjNum(pNew) {
        Vec_IntShrink( vMapping, Gia_ManObjNum(pNew) );
    } else {
        Vec_IntFillExtra( vMapping, Gia_ManObjNum(pNew), 0 ); }
    assert(Vec_IntSize(vMapping) == Gia_ManObjNum(pNew), "Vec_IntSize(vMapping) == Gia_ManObjNum(pNew)");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vMapping)) && (((iLit) = Vec_IntEntry(vMapping, i)), 1); i++  {
        if iLit > 0 {
            Vec_IntAddToEntry( vMapping, i, Gia_ManObjNum(pNew) ); }
    }
    Vec_IntAppend( vMapping, vMapping2 );
    Vec_IntFree( vMapping2 );
    // attach mapping and packing
    assert(pNew.vMapping == nil, "pNew.vMapping == nil");
    pNew.vMapping = vMapping;
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p.pGia) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_ManAlloc(pGia *Gia_Man_t, pPars *Jf_Par_t) *Lf_Man_t {
    var p *Lf_Man_t; var i, k int32;
    assert(pPars.nCutNum > 1  && pPars.nCutNum <= LF_CUT_MAX, "pPars.nCutNum > 1  && pPars.nCutNum <= LF_CUT_MAX");
    assert(pPars.nLutSize > 1 && pPars.nLutSize <= LF_LEAF_MAX, "pPars.nLutSize > 1 && pPars.nLutSize <= LF_LEAF_MAX");
    // inlining of macro ABC_FREE
    pGia.Refs = nil;
    Vec_IntFreeP( &pGia.vMapping );
    Gia_ManCleanValue( pGia );
    if Gia_ManHasChoices(pGia) {
        Gia_ManSetPhase(pGia); }
    // inlining of macro ABC_CALLOC
    p = &Lf_Man_t{};
    Lf_ManAnalyzeCoDrivers( pGia, &p.nCoDrivers, &p.nInverters );
    if pPars.fPower {
        Lf_ManComputeSwitching( pGia, &p.vSwitches ); }
    p.clkStart  = Abc_Clock();
    p.pGia      = pGia;
    p.pPars     = pPars;
    p.nCutWords = (sizeof(Lf_Cut_t)/sizeof(int) + pPars.nLutSize + 1) >> 1;
    p.nSetWords = p.nCutWords * pPars.nCutNum;
    if pPars.fCutMin {
        p.vTtMem = Vec_MemAllocForTT( pPars.nLutSize, 0 );
    } else {
        p.vTtMem = nil;
    }
    if pPars.fCutMin && pPars.fUseMux7 {
        Vec_MemAddMuxTT( p.vTtMem, pPars.nLutSize ); }
    // inlining of macro ABC_CALLOC
    p.pObjBests = ((Lf_Bst_t *) calloc((size_t)(Gia_ManAndNotBufNum(pGia)), sizeof(Lf_Bst_t)));
    Vec_IntGrow( &p.vFreeSets, (1<<14) );
    Vec_PtrGrow( &p.vFreePages, 256 );
    Lf_MemAlloc( &p.vStoreOld, 16, &p.vFreePages, p.nCutWords );
    Lf_MemAlloc( &p.vStoreNew, 16, &p.vFreePages, p.nCutWords );
    Vec_IntFill( &p.vOffsets,  Gia_ManObjNum(pGia), -1 );
    Vec_IntFill( &p.vRequired, Gia_ManObjNum(pGia), ABC_INFINITY );
    Vec_IntFill( &p.vCutSets,  Gia_ManAndNotBufNum(pGia), -1 );
    Vec_FltFill( &p.vFlowRefs, Gia_ManAndNotBufNum(pGia), 0 );
    Vec_IntFill( &p.vMapRefs,  Gia_ManAndNotBufNum(pGia), 0 );
    Vec_IntFill( &p.vCiArrivals, Gia_ManCiNum(pGia), 0 );
    // inlining of macro Gia_ManForEachAndId
    for  i = 0; (i < pGia.nObjs); i++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(pGia, i)) {} else
        if !Gia_ObjIsBuf(Gia_ManObj(pGia, i)) {
            Vec_IntWriteEntry( &p.vOffsets, i, k++ ); }
    }
    assert(k == Gia_ManAndNotBufNum(pGia), "k == Gia_ManAndNotBufNum(pGia)");
    Lf_ManSetFlowRefs( pGia, &p.vFlowRefs, &p.vOffsets );
    if pPars.pTimesArr {
        for  i = 0; i < Gia_ManPiNum(pGia); i++  {
            Vec_IntWriteEntry( &p.vCiArrivals, i, pPars.pTimesArr[i] );
        } }
    return p;
}
func Lf_ManFree(p *Lf_Man_t) {
    // inlining of macro ABC_FREE
    p.pPars.pTimesArr = nil;
    // inlining of macro ABC_FREE
    p.pPars.pTimesReq = nil;
    if p.pPars.fCutMin {
        Vec_MemHashFree( p.vTtMem ); }
    if p.pPars.fCutMin {
        Vec_MemFree( p.vTtMem ); }
    Vec_PtrFreeData( &p.vMemSets );
    Vec_PtrFreeData( &p.vFreePages );
    Vec_PtrFreeData( &p.vStoreOld.vPages );
    Vec_PtrFreeData( &p.vStoreNew.vPages );
    // inlining of macro ABC_FREE
    p.vMemSets.pArray = nil;
    // inlining of macro ABC_FREE
    p.vFreePages.pArray = nil;
    // inlining of macro ABC_FREE
    p.vStoreOld.vPages.pArray = nil;
    // inlining of macro ABC_FREE
    p.vStoreNew.vPages.pArray = nil;
    // inlining of macro ABC_FREE
    p.vFreePages.pArray = nil;
    // inlining of macro ABC_FREE
    p.vFreeSets.pArray = nil;
    // inlining of macro ABC_FREE
    p.vOffsets.pArray = nil;
    // inlining of macro ABC_FREE
    p.vRequired.pArray = nil;
    // inlining of macro ABC_FREE
    p.vCutSets.pArray = nil;
    // inlining of macro ABC_FREE
    p.vFlowRefs.pArray = nil;
    // inlining of macro ABC_FREE
    p.vMapRefs.pArray = nil;
    // inlining of macro ABC_FREE
    p.vSwitches.pArray = nil;
    // inlining of macro ABC_FREE
    p.vCiArrivals.pArray = nil;
    // inlining of macro ABC_FREE
    p.pObjBests = nil;
    // inlining of macro ABC_FREE
    p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Lf_ManSetDefaultPars(pPars *Jf_Par_t) {
    memset( pPars, 0, sizeof(Jf_Par_t) );
    pPars.nLutSize     =  6;
    pPars.nCutNum      =  8;
    pPars.nProcNum     =  0;
    pPars.nRounds      =  4;
    pPars.nRoundsEla   =  1;
    pPars.nRelaxRatio  =  0;
    pPars.nCoarseLimit =  3;
    pPars.nAreaTuner   =  1;
    pPars.nVerbLimit   =  5;
    pPars.DelayTarget  = -1;
    pPars.fAreaOnly    =  0;
    pPars.fOptEdge     =  1; 
    pPars.fUseMux7     =  0;
    pPars.fPower       =  0;
    pPars.fCoarsen     =  1;
    pPars.fCutMin      =  0;
    pPars.fFuncDsd     =  0;
    pPars.fGenCnf      =  0;
    pPars.fPureAig     =  0;
    pPars.fCutHashing  =  0;
    pPars.fCutSimple   =  0;
    pPars.fVerbose     =  0;
    pPars.fVeryVerbose =  0;
    pPars.nLutSizeMax  =  LF_LEAF_MAX;
    pPars.nCutNumMax   =  LF_CUT_MAX;
}
func Lf_ManPrintStats(p *Lf_Man_t, pTitle string) {
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("%s :  ", pTitle);
    fmt.Printf("Level =%6d   ", int64(p.pPars.Delay));
    fmt.Printf("Area =%9d   ", int64(p.pPars.Area));
    fmt.Printf("Edge =%9d   ", int64(p.pPars.Edge));
    fmt.Printf("LUT =%9d  ", int64(p.pPars.Area)+p.nInverters);
    if Vec_FltSize(&p.vSwitches) {
        fmt.Printf("Swt =%8.1f  ", p.Switches); }
    if p.pPars.fUseMux7 {
        fmt.Printf("Mux7 =%7d  ", int64(p.pPars.Mux7)); }
    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    fflush( os.Stdout );
}
func Lf_ManPrintInit(p *Lf_Man_t) {
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("LutSize = %d  ", p.pPars.nLutSize);
    fmt.Printf("CutNum = %d  ", p.pPars.nCutNum);
    fmt.Printf("Iter = %d  ", p.pPars.nRounds + p.pPars.nRoundsEla);
    if p.pPars.nRelaxRatio {
    fmt.Printf("Ratio = %d  ", p.pPars.nRelaxRatio); }
    fmt.Printf("Edge = %d  ", p.pPars.fOptEdge);
    if p.pPars.DelayTarget != -1 {
    fmt.Printf("Delay = %d  ", p.pPars.DelayTarget); }
    fmt.Printf("CutMin = %d  ", p.pPars.fCutMin);
    fmt.Printf("Coarse = %d  ", p.pPars.fCoarsen);
    fmt.Printf("Cut/Set = %d/%d Bytes", 8*p.nCutWords, 8*p.nSetWords);
    fmt.Printf("\n");
    fmt.Printf("Computing cuts...\r");
    fflush( os.Stdout );
}
func Lf_ManPrintQuit(p *Lf_Man_t, pNew *Gia_Man_t) {
    var MemGia float32    = Gia_ManMemory(p.pGia) / (1<<20);
    var MemMan float32    = 1.0 * sizeof(int) * (2 * Gia_ManObjNum(p.pGia) + 3 * Gia_ManAndNotBufNum(p.pGia)) / (1<<20); // offset, required, cutsets, maprefs, flowrefs
    var MemCutsB float32  = 1.0 * (p.vStoreOld.MaskPage + 1) * (Vec_PtrSize(&p.vFreePages) + Vec_PtrSize(&p.vStoreOld.vPages)) / (1<<20) + 1.0 * sizeof(Lf_Bst_t) * Gia_ManAndNotBufNum(p.pGia) / (1<<20);
    var MemCutsF float32  = 1.0 * sizeof(word) * p.nSetWords * (1<<LF_LOG_PAGE) * Vec_PtrSize(&p.vMemSets) / (1<<20);
    var MemTt float32;
    if p.vTtMem {
        MemTt = Vec_MemMemory(p.vTtMem) / (1<<20);
    } else {
        MemTt = 0;
    }
    var MemMap float32    = Vec_IntMemory(pNew.vMapping) / (1<<20);
    if p.CutCount[0] == 0 {
        p.CutCount[0] = 1; }
    if !p.pPars.fVerbose {
        var i int32;
        var CountOver [2]int32 ;
        var nLutSize int32;
        if p.pPars.fCutGroup {
            nLutSize = p.pPars.nLutSize/2;
        } else {
            nLutSize = p.pPars.nLutSize;
        }
        // inlining of macro Gia_ManForEachLut
        for  i = 1; i < Gia_ManObjNum(pNew); i++  { if !Gia_ObjIsLut(pNew, i) {} else {
            CountOver[Gia_ObjLutSize(pNew, i) > nLutSize]++; }
        }
        if p.pPars.fCutGroup {
            fmt.Printf("Created %d regular %d-LUTs and %d dual %d-LUTs. The total of %d %d-LUTs.\n", CountOver[0], nLutSize, CountOver[1], nLutSize, CountOver[0] + 2*CountOver[1], nLutSize); }
        return;
    }
    fmt.Printf("CutPair = %.0f  ", p.CutCount[0]);
    fmt.Printf("Merge = %.0f (%.2f %%)  ", p.CutCount[1], 100.0*p.CutCount[1]/p.CutCount[0]);
    fmt.Printf("Eval = %.0f (%.2f %%)  ", p.CutCount[2], 100.0*p.CutCount[2]/p.CutCount[0]);
    fmt.Printf("Cut = %.0f (%.2f %%)  ", p.CutCount[3], 100.0*p.CutCount[3]/p.CutCount[0]);
    fmt.Printf("\n");
    fmt.Printf("Gia = %.2f MB  ", MemGia);
    fmt.Printf("Man = %.2f MB  ", MemMan); 
    fmt.Printf("Best = %.2f MB  ", MemCutsB);
    fmt.Printf("Front = %.2f MB   ", MemCutsF);
    fmt.Printf("Map = %.2f MB  ", MemMap); 
    fmt.Printf("TT = %.2f MB  ", MemTt); 
    fmt.Printf("Total = %.2f MB", MemGia + MemMan + MemCutsB + MemCutsF + MemMap + MemTt); 
    fmt.Printf("\n");
    if true {
        var i int32;
        for  i = 0; i <= p.pPars.nLutSize; i++  {
            fmt.Printf("%d:%d  ", i, p.nCutCounts[i]);
        }
        fmt.Printf("Equal = %d (%.0f %%) ", p.nCutEqual, 100.0 * p.nCutEqual / p.Iter / Gia_ManAndNotBufNum(p.pGia));
        if p.vTtMem {
            fmt.Printf("TT = %d (%.2f %%)  ", Vec_MemEntryNum(p.vTtMem), 100.0 * Vec_MemEntryNum(p.vTtMem) / p.CutCount[2]); }
        if p.pGia.pMuxes && p.nCutMux {
            fmt.Printf("MuxTT = %d (%.0f %%) ", p.nCutMux, 100.0 * p.nCutMux / p.Iter / Gia_ManMuxNum(p.pGia)); }
        fmt.Printf("\n");
    }
    fmt.Printf("CoDrvs = %d (%.2f %%)  ", p.nCoDrivers, 100.0*p.nCoDrivers/Gia_ManCoNum(p.pGia));
    fmt.Printf("CoInvs = %d (%.2f %%)  ", p.nInverters, 100.0*p.nInverters/Gia_ManCoNum(p.pGia));
    fmt.Printf("Front = %d (%.2f %%)  ", p.nFrontMax, 100.0*p.nFrontMax/Gia_ManAndNum(p.pGia));
    fmt.Printf("TimeFails = %d   ", p.nTimeFails);
    Abc_PrintTime( 1, "Time",    Abc_Clock() - p.clkStart );
    fflush( os.Stdout );
}
func Lf_ManComputeMapping(p *Lf_Man_t) {
    var pObj *Gia_Obj_t;
    var i, arrTime int32;
    assert(p.vStoreNew.iCur == 0, "p.vStoreNew.iCur == 0");
    Lf_ManSetCutRefs( p );
    if p.pGia.pManTime != nil {
        assert(!Gia_ManBufNum(p.pGia), "!Gia_ManBufNum(p.pGia)");
        Tim_ManIncrementTravId( (Tim_Man_t*)p.pGia.pManTime );
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {
            if Gia_ObjIsBuf(pObj) {
                continue; }
            if Gia_ObjIsAnd(pObj) {
                Lf_ObjMergeOrder( p, i );
            } else if Gia_ObjIsCi(pObj) {
                arrTime = Tim_ManGetCiArrival( (Tim_Man_t*)p.pGia.pManTime, Gia_ObjCioId(pObj) );
                Lf_ObjSetCiArrival( p, Gia_ObjCioId(pObj), arrTime );
            
            } else if Gia_ObjIsCo(pObj) {
                arrTime = Lf_ObjCoArrival( p, Gia_ObjCioId(pObj) );
                Tim_ManSetCoArrival( (Tim_Man_t*)p.pGia.pManTime, Gia_ObjCioId(pObj), arrTime );
            
            } else { assert(false, "0"); }
        }
//        Tim_ManPrint( p->pGia->pManTime );
    
    } else {
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
            if !Gia_ObjIsBuf(pObj) {
                Lf_ObjMergeOrder( p, i ); }
        }
    }
    Lf_MemRecycle( &p.vStoreOld );
    // inlining of macro ABC_SWAP
    { var t Lf_Mem_t  = p.vStoreOld; p.vStoreOld = p.vStoreNew; p.vStoreNew = t; };
    if p.fUseEla {
        Lf_ManCountMapRefs( p );
    } else {
        Lf_ManSetMapRefs( p ); }
    Lf_ManPrintStats( p, (char *)(p.fUseEla ? "Ela  " : (p.Iter ? "Area " : "Delay")) );
}
func Lf_ManPerformMappingInt(pGia *Gia_Man_t, pPars *Jf_Par_t) *Gia_Man_t {
    var fUsePowerMode int32;
    var p *Lf_Man_t;
    var pNew, pCls *Gia_Man_t;
    if pPars.fUseMux7 {
        pPars.fCoarsen = 1, pPars.nRoundsEla = 0; }
    if Gia_ManHasChoices(pGia) || pPars.nLutSizeMux {
        pPars.fCutMin = 1; } 
    if pPars.fCoarsen {
        pCls = Gia_ManDupMuxes(pGia, pPars.nCoarseLimit);
        pCls.pManTime = pGia.pManTime; pGia.pManTime = nil;
    
    } else { pCls = pGia; }
    p = Lf_ManAlloc( pCls, pPars );
    if pPars.fVerbose && pPars.fCoarsen {
        fmt.Printf("Initial ");  Gia_ManPrintMuxStats( pGia );  fmt.Printf("\n");
        fmt.Printf("Derived ");  Gia_ManPrintMuxStats( pCls );  fmt.Printf("\n");
    }
    Lf_ManPrintInit( p );

    // power mode
    if (fUsePowerMode != 0) && Vec_FltSize(&p.vSwitches) {
        pPars.fPower = 0; }

    // perform mapping
    for  p.Iter = 0; p.Iter < p.pPars.nRounds; p.Iter++  {
        Lf_ManComputeMapping( p );
    }
    p.fUseEla = 1;
    for  ; p.Iter < p.pPars.nRounds + pPars.nRoundsEla; p.Iter++  {
        Lf_ManComputeMapping( p );
    }

    // power mode
    if (fUsePowerMode != 0) && Vec_FltSize(&p.vSwitches) {
        pPars.fPower = 1;
        for  ; p.Iter < p.pPars.nRounds + pPars.nRoundsEla + 2; p.Iter++  {
            Lf_ManComputeMapping( p );
        }
    }

    if pPars.fVeryVerbose && pPars.fCutMin {
        Vec_MemDumpTruthTables( p.vTtMem, Gia_ManName(p.pGia), pPars.nLutSize ); }
    if pPars.fCutMin {
        pNew = Lf_ManDeriveMappingGia( p );
    } else if pPars.fCoarsen {
        pNew = Lf_ManDeriveMappingCoarse( p );
    } else {
        pNew = Lf_ManDeriveMapping( p ); }
    Gia_ManMappingVerify( pNew );
    Lf_ManPrintQuit( p, pNew );
    Lf_ManFree( p );
    if pCls != pGia {
        pGia.pManTime = pCls.pManTime; pCls.pManTime = nil;
        Gia_ManStop( pCls );
    }
    return pNew;
}
func Lf_ManPerformMapping(p *Gia_Man_t, pPars *Jf_Par_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    if p.pManTime && Tim_ManBoxNum((Tim_Man_t*)p.pManTime) && Gia_ManIsNormalized(p) {
        var pTimOld *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
        p.pManTime = Tim_ManDup( pTimOld, 1 );
        pNew = Gia_ManDupUnnormalize( p );
        if pNew == nil {
            return nil; }
        Gia_ManTransferTiming( pNew, p );
        p = pNew;
        // mapping
        pNew = Lf_ManPerformMappingInt( p, pPars );
        if pNew != p {
            Gia_ManTransferTiming( pNew, p );
            Gia_ManStop( p );
        }
        // normalize
        pNew = Gia_ManDupNormalize( p = pNew, 0 );
        Gia_ManTransferMapping( pNew, p );
//        Gia_ManTransferPacking( pNew, p );
        Gia_ManTransferTiming( pNew, p );
        Gia_ManStop( p ); // do not delete if the original one!
        // cleanup
        Tim_ManStop( (Tim_Man_t *)pNew.pManTime );
        pNew.pManTime = pTimOld;
        assert(Gia_ManIsNormalized(pNew), "Gia_ManIsNormalized(pNew)");
    
    } else {
        // mapping
        pNew = Lf_ManPerformMappingInt( p, pPars );
        Gia_ManTransferTiming( pNew, p );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Interface of LUT mapping package.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Gia_ManPerformLfMapping(p *Gia_Man_t, pPars *Jf_Par_t, fNormalized int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    assert(!pPars.fCutGroup || pPars.nLutSize == 9 || pPars.nLutSize == 11 || pPars.nLutSize == 13, "!pPars.fCutGroup || pPars.nLutSize == 9 || pPars.nLutSize == 11 || pPars.nLutSize == 13");
    // reconstruct GIA according to the hierarchy manager
    assert(pPars.pTimesArr == nil, "pPars.pTimesArr == nil");
    assert(pPars.pTimesReq == nil, "pPars.pTimesReq == nil");
    if p.pManTime {
        if fNormalized != 0 {
            pNew = Gia_ManDupUnnormalize( p );
            if pNew == nil {
                return nil; }
            Gia_ManTransferTiming( pNew, p );
            p = pNew;
            // set arrival and required times
            pPars.pTimesArr = Tim_ManGetArrTimes( (Tim_Man_t *)p.pManTime );
            pPars.pTimesReq = Tim_ManGetReqTimes( (Tim_Man_t *)p.pManTime );
        
        } else {
            p = Gia_ManDup( p ); }
    
    } else { 
        p = Gia_ManDup( p ); }
    // perform mapping
    pNew = Lf_ManPerformMappingInt( p, pPars );
    if pNew != p {
        // transfer name
        // inlining of macro ABC_FREE
        pNew.pName = "";
        // inlining of macro ABC_FREE
        pNew.pSpec = "";
        pNew.pName = Abc_UtilStrsav( p.pName );
        pNew.pSpec = Abc_UtilStrsav( p.pSpec );
        Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
        // return the original (unmodified by the mapper) timing manager
        Gia_ManTransferTiming( pNew, p );
        Gia_ManStop( p );
    }
    // normalize and transfer mapping
    pNew = Gia_ManDupNormalize( p = pNew, 0 );
    Gia_ManTransferMapping( pNew, p );
//    Gia_ManTransferPacking( pNew, p );
    Gia_ManTransferTiming( pNew, p );
    Gia_ManStop( p );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


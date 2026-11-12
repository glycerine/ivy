package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcUnroll.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Unrolling manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmc.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const UNR_DIFF_NULL = 0x7FFF

    // 24 bytes + (RankMax-1) * 4 bytes
type Unr_Obj_t struct {
    hFan0              uint32;        // address of the fanin
    hFan1              uint32;        // address of the fanin
    fCompl0 uint32; /* C bitfield: 1 bits; review packing and truncation */ // complemented bit
    fCompl1 uint32; /* C bitfield: 1 bits; review packing and truncation */ // complemented bit
    uRDiff0 uint32; /* C bitfield: 15 bits; review packing and truncation */ // rank diff of the fanin
    uRDiff1 uint32; /* C bitfield: 15 bits; review packing and truncation */ // rank diff of the fanin
    fItIsPi uint32; /* C bitfield: 1 bits; review packing and truncation */ // remember attributes
    fItIsPo uint32; /* C bitfield: 1 bits; review packing and truncation */ // remember attributes
    RankMax uint32; /* C bitfield: 15 bits; review packing and truncation */ // max rank diff between node and its fanout
    RankCur uint32; /* C bitfield: 15 bits; review packing and truncation */ // cur rank of the node
    OrigId              uint32;       // original object ID
    Res [1]uint32;       // RankMax entries
}

type Unr_Man_t_ struct {
    // input data
    pGia            *Gia_Man_t;           // the user's AIG manager
    pFrames            *Gia_Man_t;        // unrolled manager
    nObjs                   int32;          // the number of objects
    // intermediate data
    vOrder            *Vec_Int_t;         // ordering of GIA objects
    vOrderLim            *Vec_Int_t;      // beginning of each time frame
    vTents            *Vec_Int_t;         // tents of GIA objects
    vRanks            *Vec_Int_t;         // ranks of GIA objects
    // unrolling data
    pObjs                  *int32;          // storage for unroling objects
    pEnd                  *int32;           // end of storage
    vObjLim            *Vec_Int_t;        // handle of the first object in each frame
    vCiMap            *Vec_Int_t;         // mapping of GIA CIs into unrolling objects
    vCoMap            *Vec_Int_t;         // mapping of GIA POs into unrolling objects
    vPiLits            *Vec_Int_t;        // storage for PI literals
}

func Unr_ManObj(p *Unr_Man_t, h int32) *Unr_Obj_t { assert(h >= 0 && h < p.pEnd - p.pObjs, "h >= 0 && h < p.pEnd - p.pObjs"); return (Unr_Obj_t *)(p.pObjs + h);  }
func Unr_ObjSizeInt(Rank int32) int32 { return 0xFFFFFFFE & (sizeof(Unr_Obj_t) / sizeof(int) + Rank);                    }
func Unr_ObjSize(pObj *Unr_Obj_t) int32 { return Unr_ObjSizeInt(pObj.RankMax);                                            }

func Unr_ManFanin0Value(p *Unr_Man_t, pObj *Unr_Obj_t) int32 {
    var pFanin *Unr_Obj_t  = Unr_ManObj( p, pObj.hFan0 );
    var Index int32  = (pFanin.RankCur + pFanin.RankMax - pObj.uRDiff0) % pFanin.RankMax;
    assert(pFanin.RankCur < pFanin.RankMax, "pFanin.RankCur < pFanin.RankMax");
    assert(pObj.uRDiff0 < pFanin.RankMax, "pObj.uRDiff0 < pFanin.RankMax");
    return Abc_LitNotCond( pFanin.Res[Index], pObj.fCompl0 );
}
func Unr_ManFanin1Value(p *Unr_Man_t, pObj *Unr_Obj_t) int32 {
    var pFanin *Unr_Obj_t  = Unr_ManObj( p, pObj.hFan1 );
    var Index int32  = (pFanin.RankCur + pFanin.RankMax - pObj.uRDiff1) % pFanin.RankMax;
    assert(pFanin.RankCur < pFanin.RankMax, "pFanin.RankCur < pFanin.RankMax");
    assert(pObj.uRDiff1 < pFanin.RankMax, "pObj.uRDiff1 < pFanin.RankMax");
    return Abc_LitNotCond( pFanin.Res[Index], pObj.fCompl1 );
}
func Unr_ManObjReadValue(pObj *Unr_Obj_t) int32 {
    assert(pObj.RankCur >= 0 && pObj.RankCur < pObj.RankMax, "pObj.RankCur >= 0 && pObj.RankCur < pObj.RankMax");
    return pObj.Res[ pObj.RankCur ];
}
func Unr_ManObjSetValue(pObj *Unr_Obj_t, Value int32) {
    assert(Value >= 0, "Value >= 0");
    pObj.RankCur = (UNR_DIFF_NULL & (pObj.RankCur + 1)) % pObj.RankMax;
    pObj.Res[ pObj.RankCur ] = Value;
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
func Vec_IntWriteMaxEntry(p *Vec_Int_t, i int32, Entry int32) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Abc_MaxInt( p.pArray[i], Entry );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unr_ManProfileRanks(vRanks *Vec_Int_t) {
    var RankMax int32  = Vec_IntFindMax( vRanks );
    var vCounts *Vec_Int_t  = Vec_IntStart( RankMax+1 );
    var i, Rank, Count, nExtras int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRanks)) && (((Rank) = Vec_IntEntry(vRanks, i)), 1); i++  {
        Vec_IntAddToEntry( vCounts, Rank, 1 );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCounts)) && (((Count) = Vec_IntEntry(vCounts, i)), 1); i++  {
        if Count == 0 {
            continue; }
        fmt.Printf("%2d : %8d  (%6.2f %%)\n", i, Count, 100.0 * Count / Vec_IntSize(vRanks));
        nExtras += Count * i;
    }
    fmt.Printf("Extra space = %d (%6.2f %%)  ", nExtras, 100.0 * nExtras / Vec_IntSize(vRanks));
    Vec_IntFree( vCounts );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unr_ManSetup_rec(p *Unr_Man_t, iObj int32, iTent int32, vRoots *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var iFanin int32;
    if Vec_IntEntry(p.vTents, iObj) >= 0 {
        return; }
    Vec_IntWriteEntry(p.vTents, iObj, iTent);
    pObj = Gia_ManObj( p.pGia, iObj );
    if Gia_ObjIsAnd(pObj) || Gia_ObjIsCo(pObj) {
        Unr_ManSetup_rec( p, (iFanin = Gia_ObjFaninId0(pObj, iObj)), iTent, vRoots );
        Vec_IntWriteMaxEntry( p.vRanks, iFanin, Abc_MaxInt(0, iTent - Vec_IntEntry(p.vTents, iFanin) - 1) );
    }
    if Gia_ObjIsAnd(pObj) {
        Unr_ManSetup_rec( p, (iFanin = Gia_ObjFaninId1(pObj, iObj)), iTent, vRoots );
        Vec_IntWriteMaxEntry( p.vRanks, iFanin, Abc_MaxInt(0, iTent - Vec_IntEntry(p.vTents, iFanin) - 1) );
    
    } else if Gia_ObjIsRo(p.pGia, pObj) {
        Vec_IntPush( vRoots, (iFanin = Gia_ObjId(p.pGia, Gia_ObjRoToRi(p.pGia, pObj))) );
        Vec_IntWriteMaxEntry( p.vRanks, iFanin, 0 );
    }
    Vec_IntPush( p.vOrder, iObj );
}
func Unr_ManSetup(p *Unr_Man_t, fVerbose int32) {
    var vRoots, vRoots2, vMap *Vec_Int_t;
    var pUnrObj *Unr_Obj_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var k int32;
    var t int32;
    var iObj int32;
    var nInts int32;
    var pInts *int32;
    var clk abctime  = Abc_Clock();
    // create zero rank
    assert(Vec_IntSize(p.vOrder) == 0, "Vec_IntSize(p.vOrder) == 0");
    Vec_IntPush( p.vOrder, 0 );
    Vec_IntPush( p.vOrderLim, Vec_IntSize(p.vOrder) );
    Vec_IntWriteEntry( p.vTents, 0, 0 );
    Vec_IntWriteEntry( p.vRanks, 0, 0 );
    // start from the POs
    vRoots  = Vec_IntAlloc( 100 );
    vRoots2 = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p.pGia)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        Unr_ManSetup_rec( p, Gia_ObjId(p.pGia, pObj), 0, vRoots );
    }
    // collect tents
    for  Vec_IntSize(vRoots) > 0  {
        Vec_IntPush( p.vOrderLim, Vec_IntSize(p.vOrder) );
        Vec_IntClear( vRoots2 );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vRoots)) && (((iObj) = Vec_IntEntry(vRoots, i)), 1); i++  {
            Unr_ManSetup_rec( p, iObj, Vec_IntSize(p.vOrderLim)-1, vRoots2 );
        }
        // inlining of macro ABC_SWAP
        { var t *Vec_Int_t  = vRoots; vRoots = vRoots2; vRoots2 = t; };
    }
    Vec_IntPush( p.vOrderLim, Vec_IntSize(p.vOrder) );
    Vec_IntFree( vRoots );
    Vec_IntFree( vRoots2 );
    // allocate memory
    nInts = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vOrder)) && (((iObj) = Vec_IntEntry(p.vOrder, i)), 1); i++  {
        nInts += Unr_ObjSizeInt( Vec_IntEntry(p.vRanks, iObj) + 1 );
    }
    // inlining of macro ABC_CALLOC
    p.pObjs = pInts = ((int *) calloc((size_t)(nInts), sizeof(int)));
    p.pEnd = p.pObjs + nInts;
    // create const0 node
    pUnrObj = Unr_ManObj( p, pInts - p.pObjs );
    pUnrObj.RankMax = Vec_IntEntry(p.vRanks, 0) + 1;
    pUnrObj.uRDiff0 = pUnrObj.uRDiff1 = UNR_DIFF_NULL;
    pUnrObj.Res[0]  = 0; // const0
    // map the objects
    vMap = Vec_IntStartFull( p.nObjs );
    Vec_IntWriteEntry( vMap, 0, pInts - p.pObjs );
    pInts += Unr_ObjSize(pUnrObj);
    // mark up the entries
    assert(Vec_IntSize(p.vObjLim) == 0, "Vec_IntSize(p.vObjLim) == 0");
    for  t = Vec_IntSize(p.vOrderLim) - 2; t >= 0; t--  {
        var Beg int32  = Vec_IntEntry(p.vOrderLim, t);
        var End int32  = Vec_IntEntry(p.vOrderLim, t+1);
        Vec_IntPush( p.vObjLim, pInts - p.pObjs );
        // inlining of macro Vec_IntForEachEntryStartStop
        for  i = Beg; (i < End) && (((iObj) = Vec_IntEntry(p.vOrder, i)), 1); i++  {
            pObj = Gia_ManObj( p.pGia, iObj );
            pUnrObj = Unr_ManObj( p, pInts - p.pObjs );
            pUnrObj.uRDiff0 = pUnrObj.uRDiff1 = UNR_DIFF_NULL;
            if Gia_ObjIsAnd(pObj) || Gia_ObjIsCo(pObj) {
                pUnrObj.uRDiff0 = Abc_MaxInt(0, Vec_IntEntry(p.vTents, iObj) - Vec_IntEntry(p.vTents, Gia_ObjFaninId0(pObj, iObj)) - 1); }
            if Gia_ObjIsAnd(pObj) {
                pUnrObj.uRDiff1 = Abc_MaxInt(0, Vec_IntEntry(p.vTents, iObj) - Vec_IntEntry(p.vTents, Gia_ObjFaninId1(pObj, iObj)) - 1);
            } else if Gia_ObjIsRo(p.pGia, pObj) {
                pUnrObj.uRDiff0 = 0; }
            pUnrObj.RankMax = Vec_IntEntry(p.vRanks, iObj) + 1;
            pUnrObj.RankCur = UNR_DIFF_NULL;
            pUnrObj.OrigId  = iObj;
            for  k = 0; k < int32(pUnrObj.RankMax); k++  {
                pUnrObj.Res[k] = -1;
            }
            assert(((pInts - p.pObjs) & 1) == 0, "((pInts - p.pObjs) & 1) == 0"); // align for 64-bits
            Vec_IntWriteEntry( vMap, iObj, pInts - p.pObjs );
            pInts += Unr_ObjSize( pUnrObj );
        }
    }
    assert(pInts - p.pObjs == nInts, "pInts - p.pObjs == nInts");
    // label the objects
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {
        if Vec_IntEntry(vMap, i) == -1 {
            continue; }
        pUnrObj = Unr_ManObj( p, Vec_IntEntry(vMap, i) );
        if Gia_ObjIsAnd(pObj) || Gia_ObjIsCo(pObj) {
            pUnrObj.hFan0   = Vec_IntEntry( vMap, Gia_ObjFaninId0(pObj, i) );
            pUnrObj.fCompl0 = Gia_ObjFaninC0(pObj);
            pUnrObj.fItIsPo = Gia_ObjIsPo(p.pGia, pObj);
        }
        if Gia_ObjIsAnd(pObj) {
            pUnrObj.hFan1   = Vec_IntEntry( vMap, Gia_ObjFaninId1(pObj, i) );
            pUnrObj.fCompl1 = Gia_ObjFaninC1(pObj);
        
        } else if Gia_ObjIsRo(p.pGia, pObj) {
            pUnrObj.hFan0   = Vec_IntEntry( vMap, Gia_ObjId(p.pGia, Gia_ObjRoToRi(p.pGia, pObj)) );
            pUnrObj.fCompl0 = 0;
        
        } else if Gia_ObjIsPi(p.pGia, pObj) {
            pUnrObj.hFan0   = Gia_ObjCioId(pObj);           // remember CIO id
            pUnrObj.hFan1   = Vec_IntEntry(p.vTents, i);   // remember tent
            pUnrObj.fItIsPi = 1;
        }
    }
    // store CI/PO objects;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.pGia.vCis)) && ((pObj) = Gia_ManCi(p.pGia, i)); i++  {
        Vec_IntPush( p.vCiMap, Vec_IntEntry(vMap, Gia_ObjId(p.pGia, pObj)) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        Vec_IntPush( p.vCoMap, Vec_IntEntry(vMap, Gia_ObjId(p.pGia, pObj)) );
    }
    Vec_IntFreeP( &vMap );
    // print stats
    if fVerbose != 0 {
        Unr_ManProfileRanks( p.vRanks );
        fmt.Printf("Memory usage = %6.2f MB  ", 4.0 * nInts / (1<<20));
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    Vec_IntFreeP( &p.vOrder );
    Vec_IntFreeP( &p.vOrderLim );
    Vec_IntFreeP( &p.vRanks );
    Vec_IntFreeP( &p.vTents );
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unr_ManAlloc(pGia *Gia_Man_t) *Unr_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Unr_Man_t{};
    p.pGia      = pGia;
    p.nObjs     = Gia_ManObjNum(pGia);
    p.vOrder    = Vec_IntAlloc( p.nObjs );
    p.vOrderLim = Vec_IntAlloc( 100 );
    p.vTents    = Vec_IntStartFull( p.nObjs );
    p.vRanks    = Vec_IntStart( p.nObjs );
    p.vObjLim   = Vec_IntAlloc( 100 );
    p.vCiMap    = Vec_IntAlloc( Gia_ManCiNum(pGia) );
    p.vCoMap    = Vec_IntAlloc( Gia_ManCoNum(pGia) );
    p.vPiLits   = Vec_IntAlloc( 10000 );
    p.pFrames   = Gia_ManStart( 10000 );
    p.pFrames.pName = Abc_UtilStrsav( pGia.pName );
    Gia_ManHashStart( p.pFrames );
    return p;
}
func Unr_ManFree(p *Unr_Man_t) {
    Gia_ManStop( p.pFrames );
    // intermediate data
    Vec_IntFreeP( &p.vOrder );
    Vec_IntFreeP( &p.vOrderLim );
    Vec_IntFreeP( &p.vTents );
    Vec_IntFreeP( &p.vRanks );
    // unrolling data
    Vec_IntFreeP( &p.vObjLim );
    Vec_IntFreeP( &p.vCiMap );
    Vec_IntFreeP( &p.vCoMap );
    Vec_IntFreeP( &p.vPiLits );
    p.pObjs = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Perform smart unrolling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unr_ManUnrollStart(pGia *Gia_Man_t, fVerbose int32) *Unr_Man_t {
    var i, iHandle int32;
    var p *Unr_Man_t;
    p = Unr_ManAlloc( pGia );
    Unr_ManSetup( p, fVerbose );
    for  i = 0; i < Gia_ManRegNum(p.pGia); i++  {
        if (iHandle = Vec_IntEntry(p.vCoMap, Gia_ManPoNum(p.pGia) + i)) != -1 {
            Unr_ManObjSetValue( Unr_ManObj(p, iHandle), 0 ); }
    }
    return p;
}
func Unr_ManUnrollFrame(p *Unr_Man_t, f int32) *Gia_Man_t {
    var i, iLit, iLit0, iLit1, hStart int32;
    for  i = 0; i < Gia_ManPiNum(p.pGia); i++  {
        Vec_IntPush( p.vPiLits, Gia_ManAppendCi(p.pFrames) );
    }
    hStart = Vec_IntEntry( p.vObjLim, Abc_MaxInt(0, Vec_IntSize(p.vObjLim)-1-f) );
    for  p.pObjs + hStart < p.pEnd  {
        var pUnrObj *Unr_Obj_t  = Unr_ManObj( p, hStart );
        if pUnrObj.uRDiff0 != UNR_DIFF_NULL && pUnrObj.uRDiff1 != UNR_DIFF_NULL { // AND node
        
            iLit0 = Unr_ManFanin0Value( p, pUnrObj );
            iLit1 = Unr_ManFanin1Value( p, pUnrObj );
            iLit  = Gia_ManHashAnd( p.pFrames, iLit0, iLit1 );
            Unr_ManObjSetValue( pUnrObj, iLit );
        
        } else if pUnrObj.uRDiff0 != UNR_DIFF_NULL && pUnrObj.uRDiff1 == UNR_DIFF_NULL { // PO/RI/RO
        
            iLit  = Unr_ManFanin0Value( p, pUnrObj );
            Unr_ManObjSetValue( pUnrObj, iLit );
            if pUnrObj.fItIsPo {
                Gia_ManAppendCo( p.pFrames, iLit ); }
        
        } else { // PI  (pUnrObj->hFan0 is CioId; pUnrObj->hFan1 is tent)
        
            assert(pUnrObj.fItIsPi && f >= int32(pUnrObj.hFan1), "pUnrObj.fItIsPi && f >= int32(pUnrObj.hFan1)");
            iLit = Vec_IntEntry( p.vPiLits, Gia_ManPiNum(p.pGia) * (f - pUnrObj.hFan1) + pUnrObj.hFan0 );
            Unr_ManObjSetValue( pUnrObj, iLit );
        }
        hStart += Unr_ObjSize( pUnrObj );
    }
    assert(p.pObjs + hStart == p.pEnd, "p.pObjs + hStart == p.pEnd");
    assert(Gia_ManPoNum(p.pFrames) == (f + 1) * Gia_ManPoNum(p.pGia), "Gia_ManPoNum(p.pFrames) == (f + 1) * Gia_ManPoNum(p.pGia)");
    return p.pFrames;
}
func Unr_ManUnroll(pGia *Gia_Man_t, nFrames int32) *Gia_Man_t {
    var p *Unr_Man_t;
    var pFrames *Gia_Man_t;
    var f int32;
    p = Unr_ManUnrollStart( pGia, 1 );
    for  f = 0; f < nFrames; f++  {
        Unr_ManUnrollFrame( p, f );
    }
    pFrames = Gia_ManCleanup( p.pFrames );
    Unr_ManFree( p );
    return pFrames;
}

/**Function*************************************************************

  Synopsis    [Perform naive unrolling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unr_ManUnrollSimple(pGia *Gia_Man_t, nFrames int32) *Gia_Man_t {
    var pFrames *Gia_Man_t;
    var pObj, pObjRi *Gia_Obj_t;
    var f, i int32; 
    pFrames = Gia_ManStart( 10000 );
    pFrames.pName = Abc_UtilStrsav( pGia.pName );
    Gia_ManHashAlloc( pFrames );
    Gia_ManConst0(pGia).Value = 0;
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(pGia)) && ((pObj) = Gia_ManCo(pGia, Gia_ManPoNum(pGia)+i)); i++  {
        pObj.Value = 0;
    }
    for  f = 0; f < nFrames; f++  {
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(pGia)) && ((pObj) = Gia_ManCi(pGia, i)); i++  {
            pObj.Value = Gia_ManAppendCi(pFrames);
        }
        // inlining of macro Gia_ManForEachRiRo
        for  i = 0; (i < Gia_ManRegNum(pGia)) && ((pObjRi) = Gia_ManCo(pGia, Gia_ManPoNum(pGia)+i)) && ((pObj) = Gia_ManCi(pGia, Gia_ManPiNum(pGia)+i)); i++  {
            pObj.Value = pObjRi.Value;
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = Gia_ManHashAnd( pFrames, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((pObj) = Gia_ManCo(pGia, i)); i++  {
            pObj.Value = Gia_ObjFanin0Copy(pObj);
        }
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(pGia)) && ((pObj) = Gia_ManCo(pGia, i)); i++  {
            Gia_ManAppendCo( pFrames, pObj.Value );
        }
    }
    Gia_ManHashStop( pFrames );
    Gia_ManSetRegNum( pFrames, 0 );
    pFrames = Gia_ManCleanup( pGia = pFrames );
    Gia_ManStop( pGia );
    return pFrames;
}

/**Function*************************************************************

  Synopsis    [Perform evaluation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unr_ManTest(pGia *Gia_Man_t, nFrames int32) {
    var pFrames0, pFrames1 *Gia_Man_t;
    var clk abctime  = Abc_Clock();
    pFrames0 = Unr_ManUnroll( pGia, nFrames );
    Abc_PrintTime( 1, "Unroll ", Abc_Clock() - clk );
    clk = Abc_Clock();
    pFrames1 = Unr_ManUnrollSimple( pGia, nFrames );
    Abc_PrintTime( 1, "UnrollS", Abc_Clock() - clk );

Gia_ManPrintStats( pFrames0, nil );
Gia_ManPrintStats( pFrames1, nil );
Gia_AigerWrite( pFrames0, "frames0.aig", 0, 0, 0 );
Gia_AigerWrite( pFrames1, "frames1.aig", 0, 0, 0 );

    Gia_ManStop( pFrames0 );
    Gia_ManStop( pFrames1 );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


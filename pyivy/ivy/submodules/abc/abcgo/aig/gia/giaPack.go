package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaPack.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [LUT packing.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaPack.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Collects LUT nodes in the increasing order of distance from COs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManLutCollect2(p *Gia_Man_t) *Vec_Int_t {
    var pObj *Gia_Obj_t;
    var vOrder *Vec_Int_t;
    var i, k, Id, iFan int32;
    vOrder = Vec_IntAlloc( Gia_ManLutNum(p) );
    Gia_ManIncrementTravId( p );
    // inlining of macro Gia_ManForEachCoDriver
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ObjFanin0(Gia_ManCo(p, i))); i++  {
        if !Gia_ObjIsAnd(pObj) {
            continue; }
        Id = Gia_ObjId( p, pObj );
        assert(Gia_ObjIsLut(p, Id), "Gia_ObjIsLut(p, Id)");
        if Gia_ObjIsTravIdCurrentId(p, Id) {
            continue; }
        Gia_ObjSetTravIdCurrentId(p, Id);
        Vec_IntPush( vOrder, Id );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOrder)) && (((Id) = Vec_IntEntry(vOrder, i)), 1); i++  {
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,Id) && ((iFan = Gia_ObjLutFanins(p,Id)[k]),1); k++  {
            if !Gia_ObjIsAnd(Gia_ManObj(p, iFan)) {
                continue; }
            assert(Gia_ObjIsLut(p, iFan), "Gia_ObjIsLut(p, iFan)");
            if Gia_ObjIsTravIdCurrentId(p, iFan) {
                continue; }
            Gia_ObjSetTravIdCurrentId(p, iFan);
            Vec_IntPush( vOrder, iFan );
        }
    }
    assert(Vec_IntCap(vOrder) == 16 || Vec_IntSize(vOrder) == Vec_IntCap(vOrder), "Vec_IntCap(vOrder) == 16 || Vec_IntSize(vOrder) == Vec_IntCap(vOrder)");
    Vec_IntReverseOrder( vOrder );
    return vOrder;
}
func Gia_ManLutCollect(p *Gia_Man_t) *Vec_Int_t {
    var vLuts, vDist, vOrder *Vec_Int_t;
    var i int32;
    var k int32;
    var Id int32;
    var iFan int32;
    var pPerm *int32;
    // collect LUTs
    vLuts = Vec_IntAlloc( Gia_ManAndNum(p) );
    // inlining of macro Gia_ManForEachLut
    for  Id = 1; Id < Gia_ManObjNum(p); Id++  { if !Gia_ObjIsLut(p, Id) {} else {
        Vec_IntPush( vLuts, Id ); }
    }
    // propagate distance
    vDist = Vec_IntStart( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && (((Id) = Gia_ObjFaninId0p(p, Gia_ManCo(p, i))), 1); i++  {
        Vec_IntWriteEntry( vDist, Id, 1 );
    }
    // inlining of macro Vec_IntForEachEntryReverse
    for  i = Vec_IntSize(vLuts) - 1; (i >= 0) && (((Id) = Vec_IntEntry(vLuts, i)), 1); i--  {
        var Dist int32  = 1 + Vec_IntEntry(vDist, Id);
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,Id) && ((iFan = Gia_ObjLutFanins(p,Id)[k]),1); k++  {
            Vec_IntUpdateEntry( vDist, iFan, Dist );
        }
    }
    // sort LUTs by distance
    k = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLuts)) && (((Id) = Vec_IntEntry(vLuts, i)), 1); i++  {
        Vec_IntWriteEntry( vDist, k++, -Vec_IntEntry(vDist, Id) );
    }
    Vec_IntShrink( vDist, k );
    pPerm = Abc_MergeSortCost( Vec_IntArray(vDist), Vec_IntSize(vDist) );
    // collect
    vOrder = Vec_IntAlloc( Vec_IntSize(vLuts) );
    for  i = 0; i < Vec_IntSize(vLuts); i++  {
        Vec_IntPush( vOrder, Vec_IntEntry(vLuts, pPerm[i]) );
    }
    Vec_IntFree( vDist );
    Vec_IntFree( vLuts );
    // inlining of macro ABC_FREE
    pPerm = nil;
    return vOrder;
}

/**Function*************************************************************

  Synopsis    [LUT packing algorithm.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManLutPacking(p *Gia_Man_t, nBlockSize int32, DelayRoute int32, DelayDir int32, fVerbose int32) {
    var Delays, Perm [32]int32;
    var vPacking, vStarts *Vec_Int_t;
    var vOrder *Vec_Int_t  = Gia_ManLutCollect( p );
    var vDelay *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var vBlock *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var vBSize *Vec_Int_t  = Vec_IntAlloc( 2 * Vec_IntSize(vOrder) / nBlockSize );
    var i, k, Id, iFan, nSize, iBlock, Delay, DelayMax int32;
    // create blocks
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOrder)) && (((Id) = Vec_IntEntry(vOrder, i)), 1); i++  {
        nSize = Gia_ObjLutSize( p, Id );
        assert(nSize <= 32, "nSize <= 32");
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,Id) && ((iFan = Gia_ObjLutFanins(p,Id)[k]),1); k++  {
            Delays[k] = Vec_IntEntry(vDelay, iFan);
            Perm[k] = iFan;
        }
        Vec_IntSelectSortCost2Reverse( Perm, nSize, Delays );
        assert(nSize < 2 || Delays[0] >= Delays[nSize-1], "nSize < 2 || Delays[0] >= Delays[nSize-1]");
        assert(Delays[0] >= 0 && Delays[nSize-1] >= 0, "Delays[0] >= 0 && Delays[nSize-1] >= 0");
        // check if we can reduce delay by adding it to the same bin as the latest one
        iBlock = Vec_IntEntry( vBlock, Perm[0] );
        if Delays[0] > 0 && Delays[0] > Delays[1] && Vec_IntEntry(vBSize, iBlock) < nBlockSize {
            Delay = Delays[0] + DelayDir;
            Vec_IntWriteEntry( vBlock, Id, iBlock );
            Vec_IntAddToEntry( vBSize, iBlock, 1 );
        
        } else { // clean new block
        
            Delay = Delays[0] + DelayRoute;
            Vec_IntWriteEntry( vBlock, Id, Vec_IntSize(vBSize) );
            Vec_IntPush( vBSize, 1 );
        }
        // calculate delay
        for  k = 1; k < nSize; k++  {
            Delay = Abc_MaxInt( Delay, Delays[k] + DelayRoute );
        }
        Vec_IntWriteEntry( vDelay, Id, Delay );
        DelayMax = Abc_MaxInt( DelayMax, Delay );
    }
    assert(Vec_IntSum(vBSize) == Vec_IntSize(vOrder), "Vec_IntSum(vBSize) == Vec_IntSize(vOrder)");
    // create packing info
    vPacking = Vec_IntAlloc( Vec_IntSize(vBSize) + Vec_IntSize(vOrder) + 1 );
    Vec_IntPush( vPacking, Vec_IntSize(vBSize) );
    // create starting places for each block
    vStarts = Vec_IntAlloc( Vec_IntSize(vBSize) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vBSize)) && (((nSize) = Vec_IntEntry(vBSize, i)), 1); i++  {
        Vec_IntPush( vPacking, nSize );
        Vec_IntPush( vStarts, Vec_IntSize(vPacking) );
        Vec_IntFillExtra( vPacking, Vec_IntSize(vPacking) + nSize, -1 );
    }
    assert(Vec_IntCap(vPacking) == 16 || Vec_IntSize(vPacking) == Vec_IntCap(vPacking), "Vec_IntCap(vPacking) == 16 || Vec_IntSize(vPacking) == Vec_IntCap(vPacking)");
    // collect LUTs from the block
    // inlining of macro Vec_IntForEachEntryReverse
    for  i = Vec_IntSize(vOrder) - 1; (i >= 0) && (((Id) = Vec_IntEntry(vOrder, i)), 1); i--  {
        var Block int32  = Vec_IntEntry( vBlock, Id );
        var Start int32  = Vec_IntEntry( vStarts, Block );
        assert(Vec_IntEntry(vPacking, Start) == -1, "Vec_IntEntry(vPacking, Start) == -1");
        Vec_IntWriteEntry( vPacking, Start, Id );
        Vec_IntAddToEntry( vStarts, Block, 1 );
    }
    assert(Vec_IntCountEntry(vPacking, -1) == 0, "Vec_IntCountEntry(vPacking, -1) == 0");
    // cleanup
    Vec_IntFree( vOrder );
    Vec_IntFree( vDelay );
    Vec_IntFree( vBlock );
    Vec_IntFree( vBSize );
    Vec_IntFree( vStarts );
    Vec_IntFreeP( &p.vPacking );
    p.vPacking = vPacking;
    fmt.Printf("Global delay = %d.\n", DelayMax);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


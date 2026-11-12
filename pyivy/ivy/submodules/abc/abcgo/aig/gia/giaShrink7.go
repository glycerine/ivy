package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaShrink7.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Implementation of DAG-aware unmapping for 6-input cuts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaShrink6.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecHash.h"
//#include "misc/util/utilTruth.h"


//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// operation manager
   
/**Function*************************************************************

  Synopsis    [Computes information about node pairs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unm_ManPrintPairStats(pHash *Hash_IntMan_t, nTotal0 int32, nPairs0 int32, nPairs1 int32, fUseLit int32) int32 {
    var i int32;
    var Num int32;
    var nRefs int32;
    var nPairs int32;
    var nTotal int32;
    var Counter [21]int32 ;
    Num = Hash_IntManEntryNum( pHash );
    for  i = 1; i <= Num; i++  {
        nRefs = Abc_MinInt( 20, Hash_IntObjData2(pHash, i) );
        nTotal += nRefs;
        Counter[nRefs]++;
        nPairs += (nRefs > 1);
/*
        if ( fUseLit )
            printf( "(%c%c, %c%c) %d\n", 
                Abc_LitIsCompl(Hash_IntObjData0(pHash, i)-2) ? '!' : ' ', 
                'a' + Abc_Lit2Var(Hash_IntObjData0(pHash, i)-2), 
                Abc_LitIsCompl(Hash_IntObjData1(pHash, i)-2) ? '!' : ' ', 
                'a' + Abc_Lit2Var(Hash_IntObjData1(pHash, i)-2), nRefs );
        else
            printf( "( %c,  %c) %d\n", 
                'a' + Hash_IntObjData0(pHash, i)-1, 
                'a' + Hash_IntObjData1(pHash, i)-1, nRefs );
*/
//        printf( "(%4d, %4d) %d\n", Hash_IntObjData0(pHash, i), Hash_IntObjData1(pHash, i), nRefs );

    }
    fmt.Printf("Statistics for pairs appearing less than 20 times:\n");
    for  i = 0; i < 21; i++  {
        if Counter[i] > 0 {
            fmt.Printf("%3d : %7d  %7.2f %%\n", i, Counter[i], 100.0 * Counter[i] * i / Abc_MaxInt(nTotal, 1)); }
    }
    fmt.Printf("Pairs:  Total = %8d    Init = %8d %7.2f %%    Final = %8d %7.2f %%    Real = %8d %7.2f %%\n", nTotal0, nPairs0, 100.0 * nPairs0 / Abc_MaxInt(nTotal0, 1), nPairs, 100.0 * nPairs / Abc_MaxInt(nTotal0, 1), nPairs1, 100.0 * nPairs1 / Abc_MaxInt(nTotal0, 1));
    return nPairs;
}
func Unm_ManComputePairs(p *Unm_Man_t, fVerbose int32) *Vec_Int_t {
    var pObj *Gia_Obj_t;
    var vPairs *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vNum2Obj *Vec_Int_t  = Vec_IntStart( 1 );
    var pHash *Hash_IntMan_t  = Hash_IntManStart( 1000 );
    var nTotal, nPairs0, nPairs int32;
    var i, k, j, FanK, FanJ, Num, nRefs int32;
    Gia_ManSetRefsMapped( p.pGia );
    // inlining of macro Gia_ManForEachLut
    for  i = 1; i < Gia_ManObjNum(p.pGia); i++  { if !Gia_ObjIsLut(p.pGia, i) {} else {
        nTotal += Gia_ObjLutSize(p.pGia, i) * (Gia_ObjLutSize(p.pGia, i) - 1) / 2;
        pObj = Gia_ManObj( p.pGia, i );
        // collect leaves of this gate  
        Vec_IntClear( p.vLeaves );
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p.pGia,i) && ((Num = Gia_ObjLutFanins(p.pGia,i)[k]),1); k++  {
            if Gia_ObjRefNumId(p.pGia, Num) > 1 {
                Vec_IntPush( p.vLeaves, Num ); }
        }
        if Vec_IntSize(p.vLeaves) < 2 {
            continue; }
        nPairs0 += Vec_IntSize(p.vLeaves) * (Vec_IntSize(p.vLeaves) - 1) / 2;
        // enumerate pairs
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(p.vLeaves)) && (((FanK) = Vec_IntEntry(p.vLeaves, k)), 1); k++  {
        // inlining of macro Vec_IntForEachEntryStart
        for  j = k+1; (j < Vec_IntSize(p.vLeaves)) && (((FanJ) = Vec_IntEntry(p.vLeaves, j)), 1); j++  {
            if FanK > FanJ {
                // inlining of macro ABC_SWAP
                 var t int32  = FanK; FanK = FanJ; FanJ = t; };
            Num = Hash_Int2ManInsert( pHash, FanK, FanJ, 0 );
            nRefs = Hash_Int2ObjInc(pHash, Num);
            if nRefs == 0 {
                assert(Num == Hash_IntManEntryNum(pHash), "Num == Hash_IntManEntryNum(pHash)");
                assert(Num == Vec_IntSize(vNum2Obj), "Num == Vec_IntSize(vNum2Obj)");
                Vec_IntPush( vNum2Obj, i );
                continue;
            }
            if nRefs == 1 {
                assert(Num < Vec_IntSize(vNum2Obj), "Num < Vec_IntSize(vNum2Obj)");
                Vec_IntPush( vPairs, Vec_IntEntry(vNum2Obj, Num) );
                Vec_IntPush( vPairs, FanK );
                Vec_IntPush( vPairs, FanJ);
            }
            Vec_IntPush( vPairs, i );
            Vec_IntPush( vPairs, FanK );
            Vec_IntPush( vPairs, FanJ );
        }
        }
    }
    }
    Vec_IntFree( vNum2Obj );
    if fVerbose != 0 {
        nPairs = Unm_ManPrintPairStats( pHash, nTotal, nPairs0, Vec_IntSize(vPairs) / 3, 0 ); }
    Hash_IntManStop( pHash );
    return vPairs;    
}
// finds used nodes
/**Function*************************************************************

  Synopsis    [Collects decomposable pairs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unm_ManCollectDecomp(p *Unm_Man_t, vPairs *Vec_Int_t, fVerbose int32) *Vec_Int_t {
    var uTruth word; var nNonUnique int32;
    var i, k, j, s, iObj, iNode, iUsed, FanK, FanJ, Res, Num, nRefs int32;
    var vNum2Obj *Vec_Int_t  = Vec_IntStart( 1 );
    var vPairs2 *Vec_Int_t  = Vec_IntAlloc( 1000 );
    assert(Hash_IntManEntryNum(p.pHash) == 0, "Hash_IntManEntryNum(p.pHash) == 0");
    for  i = 0; i < Vec_IntSize(vPairs); i += 3  {
        iObj = Vec_IntEntry( vPairs, i );
        assert(Gia_ObjIsLut(p.pGia, iObj), "Gia_ObjIsLut(p.pGia, iObj)");
        // collect leaves of this gate  
        Vec_IntClear( p.vLeaves );
        // inlining of macro Gia_LutForEachFanin
        for  s = 0; s < Gia_ObjLutSize(p.pGia,iObj) && ((iNode = Gia_ObjLutFanins(p.pGia,iObj)[s]),1); s++  {
            Vec_IntPush( p.vLeaves, iNode );
        }
        assert(Vec_IntSize(p.vLeaves) <= 6, "Vec_IntSize(p.vLeaves) <= 6");
        FanK = Vec_IntEntry(vPairs, i+1);
        FanJ = Vec_IntEntry(vPairs, i+2);
        k = Vec_IntFind( p.vLeaves, FanK );
        j = Vec_IntFind( p.vLeaves, FanJ );
        assert(FanK < FanJ, "FanK < FanJ");
        iUsed = Vec_IntEntry( p.vId2Used, iObj );
        uTruth = Vec_WrdEntry( p.vTruths, iUsed );
        Res = Abc_TtCheckDsdAnd( uTruth, k, j, nil );
        if Res == -1 {
            continue; }
        // derive literals
        FanK = Abc_Var2Lit( FanK, ((Res >> 0) & 1) );
        FanJ = Abc_Var2Lit( FanJ, ((Res >> 1) & 1) );
        if Res == 4 {
            // inlining of macro ABC_SWAP
             var t int32  = FanK; FanK = FanJ; FanJ = t; };
        Num = Hash_Int2ManInsert( p.pHash, FanK, FanJ, 0 );
        nRefs = Hash_Int2ObjInc(p.pHash, Num);
        if nRefs == 0 {
            assert(Num == Hash_IntManEntryNum(p.pHash), "Num == Hash_IntManEntryNum(p.pHash)");
            assert(Num == Vec_IntSize(vNum2Obj), "Num == Vec_IntSize(vNum2Obj)");
            Vec_IntPush( vNum2Obj, iObj );
            continue;
        }
        if nRefs == 1 {
            assert(Num < Vec_IntSize(vNum2Obj), "Num < Vec_IntSize(vNum2Obj)");
            Vec_IntPush( vPairs2, Vec_IntEntry(vNum2Obj, Num) );
            Vec_IntPush( vPairs2, FanK );
            Vec_IntPush( vPairs2, FanJ );
            nNonUnique++;
        }
        Vec_IntPush( vPairs2, iObj );
        Vec_IntPush( vPairs2, FanK );
        Vec_IntPush( vPairs2, FanJ );
    }
    Vec_IntFree( vNum2Obj );
    if fVerbose != 0 {
        Unm_ManPrintPairStats( p.pHash, Vec_IntSize(vPairs)/3, Hash_IntManEntryNum(p.pHash), Vec_IntSize(vPairs2)/3, 1 ); }
//    Hash_IntManStop( pHash );
    return vPairs2;    
}

/**Function*************************************************************

  Synopsis    [Compute truth tables for the selected nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unm_ManWork(p *Unm_Man_t) {
    var vPairs, vPairs2 *Vec_Int_t;
    // find the duplicated pairs
    vPairs = Unm_ManComputePairs( p, 1 );
    // find the used nodes
    p.vUsed = Unm_ManFindUsedNodes( vPairs, Gia_ManObjNum(p.pGia) );
    p.vId2Used = Vec_IntInvert( p.vUsed, -1 );
    Vec_IntFillExtra( p.vId2Used, Gia_ManObjNum(p.pGia), -1 );
    // compute truth tables for used nodes
    p.vTruths = Unm_ManComputeTruths( p );
    // derive new pairs
    vPairs2 = Unm_ManCollectDecomp( p, vPairs, 1 );
    Vec_IntFreeP( &vPairs );
    Vec_IntFreeP( &vPairs2 );
}


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Unm_ManTest(pGia *Gia_Man_t) *Gia_Man_t {
    var p *Unm_Man_t;
    p = Unm_ManAlloc( pGia );
    Unm_ManWork( p );

    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    return Unm_ManFree( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


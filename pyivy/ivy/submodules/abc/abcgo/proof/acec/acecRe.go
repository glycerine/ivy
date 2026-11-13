package acec

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [acecRe.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecRe.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecHash.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Ree_ForEachCut( pList, pCut, i ) for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += pCut[0] + 2 )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Detecting FADDs in the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ree_TruthPrecompute() {
    var Truths [8]word = [8]word{ 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77 };
    var Truth word;
    var i int32;
    for  i = 0; i < 8; i++  {
        Truth = Truths[i];
        Truth = Abc_Tt6SwapAdjacent( Truth, 1 );
        Abc_TtPrintHexRev( os.Stdout, &Truth, 3 );
        fmt.Printf("\n");
    }
    fmt.Printf("\n");
    for  i = 0; i < 8; i++  {
        Truth = Truths[i];
        Truth = Abc_Tt6SwapAdjacent( Truth, 1 );
        Truth = Abc_Tt6SwapAdjacent( Truth, 0 );
        Abc_TtPrintHexRev( os.Stdout, &Truth, 3 );
        fmt.Printf("\n");
    }
    fmt.Printf("\n");
}
func Ree_TruthPrecompute2() {
    var i, b int32;
    for  i = 0; i < 8; i++  {
        var Truth word  = 0xE8;
        for  b = 0; b < 3; b++  {
            if (i >> b) & 1 {
                Truth = Abc_Tt6Flip( Truth, b ); }
        }
        fmt.Printf("%d = %X\n", i, 0xFF & int32(Truth));
    }
}

/**Function*************************************************************

  Synopsis    [Detecting FADDs in the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ree_ManCutPrint(pCut *int32, Count int32, Truth word, iObj int32) {
    var c int32;
    fmt.Printf("%d : %d : ", Count, iObj);
    for  c = 1; c <= pCut[0]; c++  {
        fmt.Printf("%3d ", pCut[c]);
    }
    for       ; c <= 4; c++  {
        fmt.Printf("    ");
    }
    fmt.Printf("0x");
    Abc_TtPrintHexRev( os.Stdout, &Truth, 3 );
    fmt.Printf("\n");
}
func Ree_ManCutMerge(p *Gia_Man_t, iObj int32, pList0 *int32, pList1 *int32, vCuts *Vec_Int_t, pHash *Hash_IntMan_t, vData *Vec_Int_t, vXors *Vec_Int_t) {
    var fVerbose int32;
    var i int32;
    var k int32;
    var c int32;
    var Value int32;
    var Truth int32;
    var TruthC int32;
    var pCut0 *int32;
    var pCut1 *int32;
    var pCut [6]int32;
    var Count int32;
    var iXor2 int32 = -1;
    var iXor3 int32 = -1;
    if fVerbose != 0 {
        fmt.Printf("Object %d\n", iObj); }
    Vec_IntFill( vCuts, 2, 1 );
    Vec_IntPush( vCuts, iObj );
    Vec_IntPush( vCuts, 0xAA );
    // inlining of macro Ree_ForEachCut
    for  i, pCut0 = 0, pList0 + 1; i < pList0[0]; i++, pCut0 += pCut0[0] + 2  {
    // inlining of macro Ree_ForEachCut
    for  k, pCut1 = 0, pList1 + 1; k < pList1[0]; k++, pCut1 += pCut1[0] + 2  {
        if !(Ree_ManCutMergeOne(pCut0, pCut1, pCut) != 0) {
            continue; }
        if Ree_ManCutCheckEqual(vCuts, pCut) != 0 {
            continue; }
        Truth = TruthC = Ree_ManCutTruth(Gia_ManObj(p, iObj), pCut0, pCut1, pCut);
        //assert( Truth == Ree_ObjComputeTruth(p, iObj, pCut) );
        if Truth & 0x80 {
            Truth = 0xFF & ~Truth; }
        if Truth == 0x66 && iXor2 == -1 {
            iXor2 = Vec_IntSize(vCuts);
        } else if Truth == 0x69 && iXor3 == -1 {
            iXor3 = Vec_IntSize(vCuts); }
        Vec_IntAddToEntry( vCuts, 0, 1 );  
        for  c = 0; c <= pCut[0]; c++  {
            Vec_IntPush( vCuts, pCut[c] );
        }
        Vec_IntPush( vCuts, TruthC );
        if (Truth == 0x66 || Truth == 0x11 || Truth == 0x22 || Truth == 0x44 || Truth == 0x77) && pCut[0] == 2 {
            assert(pCut[0] == 2, "pCut[0] == 2");
            Value = Hsh_Int3ManInsert( pHash, pCut[1], pCut[2], 0 );
            Vec_IntPushThree( vData, iObj, Value, TruthC );
        
        } else if Truth == 0x69 || Truth == 0x17 || Truth == 0x2B || Truth == 0x4D || Truth == 0x71 {
            assert(pCut[0] == 3, "pCut[0] == 3");
            Value = Hsh_Int3ManInsert( pHash, pCut[1], pCut[2], pCut[3] );
            Vec_IntPushThree( vData, iObj, Value, TruthC );
        }
        if fVerbose != 0 {
            Ree_ManCutPrint( pCut, ++Count, TruthC, iObj ); }
    }
    }
    if !(vXors != nil) {
        return; }
    if iXor2 > 0 {
        pCut0 = Vec_IntEntryP( vCuts, iXor2 );
    } else if iXor3 > 0 {
        pCut0 = Vec_IntEntryP( vCuts, iXor3 );
    } else {
        return; }
    Vec_IntPush( vXors, iObj );
    for  c = 1; c <= pCut0[0]; c++  {
        Vec_IntPush( vXors, pCut0[c] );
    }
    if pCut0[0] == 2 {
        Vec_IntPush( vXors, 0 ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ree_ManDeriveAdds(p *Hash_IntMan_t, vData *Vec_Int_t, fVerbose int32) *Vec_Int_t {
    var i, j, k, iObj, iObj2, Value, Truth, Truth2, CountX, CountM, Index int32;
    var nEntries int32  = Hash_IntManEntryNum(p);
    var vAdds *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vXors *Vec_Int_t  = Vec_IntStart( nEntries + 1 );
    var vMajs *Vec_Int_t  = Vec_IntStart( nEntries + 1 );
    var vIndex *Vec_Int_t  = Vec_IntStartFull( nEntries + 1 );
    var vIndexRev *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vXorMap, vMajMap *Vec_Wec_t;
    // inlining of macro Vec_IntForEachEntryTriple
    for  i = 0; (i+2 < Vec_IntSize(vData)) && (((iObj) = Vec_IntEntry(vData, i)), 1) && (((Value) = Vec_IntEntry(vData, i+1)), 1) && (((Truth) = Vec_IntEntry(vData, i+2)), 1); i += 3  {
        assert(Value <= nEntries, "Value <= nEntries");
        if Truth == 0x66 || Truth == 0x99 || Truth == 0x69 || Truth == 0x96 {
            Vec_IntAddToEntry( vXors, Value, 1 );
        } else {
            Vec_IntAddToEntry( vMajs, Value, 1 ); }
    }
    // remap these into indexes
    // inlining of macro Vec_IntForEachEntryTwo
    for  i = 0; (i < Vec_IntSize(vXors)) && (((CountX) = Vec_IntEntry(vXors, i)), 1) && (((CountM) = Vec_IntEntry(vMajs, i)), 1); i++  {
        if CountX && CountM {
            Vec_IntPush( vIndexRev, i );
            Vec_IntWriteEntry( vIndex, i, Index++ );
        }
    }
    Vec_IntFree( vXors );
    Vec_IntFree( vMajs );
    //if ( fVerbose )
    //    printf( "Detected %d shared cuts among %d hashed cuts.\n", Index, nEntries );
    // collect nodes
    vXorMap = Vec_WecStart( Index );
    vMajMap = Vec_WecStart( Index );
    // inlining of macro Vec_IntForEachEntryTriple
    for  i = 0; (i+2 < Vec_IntSize(vData)) && (((iObj) = Vec_IntEntry(vData, i)), 1) && (((Value) = Vec_IntEntry(vData, i+1)), 1) && (((Truth) = Vec_IntEntry(vData, i+2)), 1); i += 3  {
        Index = Vec_IntEntry( vIndex, Value );
        if Index == -1 {
            continue; }
        if Truth == 0x66 || Truth == 0x99 || Truth == 0x69 || Truth == 0x96 {
            Vec_IntPushTwo( Vec_WecEntry(vXorMap, Index), iObj, Truth );
        } else {
            Vec_IntPushTwo( Vec_WecEntry(vMajMap, Index), iObj, Truth ); }
    }
    Vec_IntFree( vIndex );
    // create pairs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vIndexRev)) && (((Value) = Vec_IntEntry(vIndexRev, i)), 1); i++  {
        var vXorOne *Vec_Int_t  = Vec_WecEntry( vXorMap, i );
        var vMajOne *Vec_Int_t  = Vec_WecEntry( vMajMap, i );
        var pObj *Hash_IntObj_t  = Hash_IntObj( p, Value );
        // inlining of macro Vec_IntForEachEntryDouble
        for  j = 0; (j+1 < Vec_IntSize(vXorOne)) && (((iObj) = Vec_IntEntry(vXorOne, j)), 1) && (((Truth) = Vec_IntEntry(vXorOne, j+1)), 1); j += 2  {
        // inlining of macro Vec_IntForEachEntryDouble
        for  k = 0; (k+1 < Vec_IntSize(vMajOne)) && (((iObj2) = Vec_IntEntry(vMajOne, k)), 1) && (((Truth2) = Vec_IntEntry(vMajOne, k+1)), 1); k += 2  {
            var SignAnd [8]int32 = [8]int32{0x88, 0x44, 0x22, 0x11, 0x77, 0xBB, 0xDD, 0xEE};
            var SignMaj [8]int32 = [8]int32{0xE8, 0xD4, 0xB2, 0x71, 0x8E, 0x4D, 0x2B, 0x17};
            var n int32;
            var SignXor int32 = (Truth == 0x99 || Truth == 0x69) << 3;
            for  n = 0; n < 8; n++  {
                if Truth2 == SignMaj[n] {
                    break; }
            }
            if n == 8 { 
                for  n = 0; n < 8; n++  {
                    if Truth2 == SignAnd[n] {
                        break; }
                } }
            assert(n < 8, "n < 8");
            Vec_IntPushThree( vAdds, pObj.iData0, pObj.iData1, pObj.iData2 ); 
            Vec_IntPushThree( vAdds, iObj, iObj2, SignXor | n );
        }
        }
    }
    Vec_IntFree( vIndexRev );
    Vec_WecFree( vXorMap );
    Vec_WecFree( vMajMap );
    return vAdds;
}
func Ree_ManCompare(pCut0 *int32, pCut1 *int32) int32 {
    if pCut0[3] < pCut1[3] { return -1; }
    if pCut0[3] > pCut1[3] { return  1; }
    if pCut0[4] < pCut1[4] { return -1; }
    if pCut0[4] > pCut1[4] { return  1; }
    return 0;
}
func Ree_ManComputeCuts(p *Gia_Man_t, pvXors **Vec_Int_t, fVerbose int32) *Vec_Int_t {
    extern void Ree_ManRemoveTrivial( Gia_Man_t * p, Vec_Int_t * vAdds );
    extern void Ree_ManRemoveContained( Gia_Man_t * p, Vec_Int_t * vAdds );
    var pObj *Gia_Obj_t; 
    var pList0 *int32;
    var pList1 *int32;
    var i int32;
    var nCuts int32;
    var pHash *Hash_IntMan_t  = Hash_IntManStart( 1000 );
    var vAdds *Vec_Int_t;
    var vTemp *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vData *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vCuts *Vec_Int_t  = Vec_IntAlloc( 30 * Gia_ManAndNum(p) );
    Vec_IntFill( vCuts, Gia_ManObjNum(p), 0 );
    Gia_ManCleanValue( p );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Vec_IntWriteEntry( vCuts, Gia_ObjId(p, pObj), Vec_IntSize(vCuts) );
        Vec_IntPush( vCuts, 1 );
        Vec_IntPush( vCuts, 1 );
        Vec_IntPush( vCuts, Gia_ObjId(p, pObj) );
        Vec_IntPush( vCuts, 0xAA );
    }
    if pvXors { *pvXors = Vec_IntAlloc( 1000 ); }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pList0 = Vec_IntEntryP( vCuts, Vec_IntEntry(vCuts, Gia_ObjFaninId0(pObj, i)) );
        pList1 = Vec_IntEntryP( vCuts, Vec_IntEntry(vCuts, Gia_ObjFaninId1(pObj, i)) );
        Ree_ManCutMerge( p, i, pList0, pList1, vTemp, pHash, vData, pvXors ? *pvXors : nil );
        Vec_IntWriteEntry( vCuts, i, Vec_IntSize(vCuts) );
        Vec_IntAppend( vCuts, vTemp );
        nCuts += Vec_IntEntry( vTemp, 0 );
    }
    }
    if fVerbose != 0 {
        fmt.Printf("AIG nodes = %d.  Cuts = %d.  Cuts/Node = %.2f.  Ints/Node = %.2f.\n", Gia_ManAndNum(p), nCuts, 1.0*nCuts/Gia_ManAndNum(p), 1.0*Vec_IntSize(vCuts)/Gia_ManAndNum(p)); }
    Vec_IntFree( vTemp );
    Vec_IntFree( vCuts );
    vAdds = Ree_ManDeriveAdds( pHash, vData, fVerbose );
    qsort( Vec_IntArray(vAdds), (size_t)(Vec_IntSize(vAdds)/6), 24, (int (*)(void *, void *))Ree_ManCompare );
    if fVerbose != 0 {
        fmt.Printf("Adders = %d.  Total cuts = %d.  Hashed cuts = %d.  Hashed/Adders = %.2f.\n", Vec_IntSize(vAdds)/6, Vec_IntSize(vData)/3, Hash_IntManEntryNum(pHash), 6.0*Hash_IntManEntryNum(pHash)/Vec_IntSize(vAdds)); }
    Vec_IntFree( vData );
    Hash_IntManStop( pHash );
    Ree_ManRemoveTrivial( p, vAdds );
    Ree_ManRemoveContained( p, vAdds );
    //Ree_ManPrintAdders( vAdds, 1 );
    return vAdds;
}

/**Function*************************************************************

  Synopsis    [Highlight nodes inside FAs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ree_CollectInsiders_rec(pGia *Gia_Man_t, iObj int32, vVisited *Vec_Bit_t, vInsiders *Vec_Bit_t) {
    if Vec_BitEntry(vVisited, iObj) {
        return; }
    Vec_BitSetEntry( vVisited, iObj, 1 );
    Ree_CollectInsiders_rec( pGia, Gia_ObjFaninId0p(pGia, Gia_ManObj(pGia, iObj)), vVisited, vInsiders );
    Ree_CollectInsiders_rec( pGia, Gia_ObjFaninId1p(pGia, Gia_ManObj(pGia, iObj)), vVisited, vInsiders );
    Vec_BitSetEntry( vInsiders, iObj, 1 );
}
func Ree_CollectInsiders(pGia *Gia_Man_t, vAdds *Vec_Int_t) *Vec_Bit_t {
    var vVisited *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(pGia) );
    var vInsiders *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(pGia) );
    var i, Entry1, Entry2, Entry3 int32;
    for  i = 0; 6*i < Vec_IntSize(vAdds); i++  {
        if Vec_IntEntry(vAdds, 6*i+2) == 0 { // HADD
            continue; }
        // mark inputs
        Entry1 = Vec_IntEntry( vAdds, 6*i + 0 );
        Entry2 = Vec_IntEntry( vAdds, 6*i + 1 );
        Entry3 = Vec_IntEntry( vAdds, 6*i + 2 );
        Vec_BitWriteEntry( vVisited, Entry1, 1 );
        Vec_BitWriteEntry( vVisited, Entry2, 1 );
        Vec_BitWriteEntry( vVisited, Entry3, 1 );
        // traverse from outputs
        Entry1 = Vec_IntEntry( vAdds, 6*i + 3 );
        Entry2 = Vec_IntEntry( vAdds, 6*i + 4 );
        Ree_CollectInsiders_rec( pGia, Entry1, vVisited, vInsiders );
        Ree_CollectInsiders_rec( pGia, Entry2, vVisited, vInsiders );
    }
    Vec_BitFree( vVisited );
    return vInsiders;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// removes HAs whose AND2 is part of XOR2 without additional fanout
func Ree_ManRemoveTrivial(p *Gia_Man_t, vAdds *Vec_Int_t) {
    var pObjX, pObjM *Gia_Obj_t;
    var i, k int32;
    p.pRefs = nil;
    Gia_ManCreateRefs( p );
    for  i = 0; 6*i < Vec_IntSize(vAdds); i++  {
        if Vec_IntEntry(vAdds, 6*i+2) == 0 { // HADD
        
            pObjX = Gia_ManObj( p, Vec_IntEntry(vAdds, 6*i+3) );
            pObjM = Gia_ManObj( p, Vec_IntEntry(vAdds, 6*i+4) );
            // rule out if MAJ is a fanout of XOR
            //if ( pObjX == Gia_ObjFanin0(pObjM) || pObjX == Gia_ObjFanin1(pObjM) )
            //    continue;
            // rule out if MAJ is a fanin of XOR and has no other fanouts
            if (pObjM == Gia_ObjFanin0(pObjX) || pObjM == Gia_ObjFanin1(pObjX)) && Gia_ObjRefNum(p, pObjM) == 1 {
                continue; }
        }
        memmove( Vec_IntArray(vAdds) + 6*k++, Vec_IntArray(vAdds) + 6*i, 6*sizeof(int) );
    }
    assert(k <= i, "k <= i");
    Vec_IntShrink( vAdds, 6*k );
}
// removes HAs fully contained inside FAs
func Ree_ManRemoveContained(p *Gia_Man_t, vAdds *Vec_Int_t) {
    var vInsiders *Vec_Bit_t  = Ree_CollectInsiders( p, vAdds );
    var i, k int32;
    for  i = 0; 6*i < Vec_IntSize(vAdds); i++  {
        if Vec_IntEntry(vAdds, 6*i+2) == 0 { // HADD
            if Vec_BitEntry(vInsiders, Vec_IntEntry(vAdds, 6*i+3)) && Vec_BitEntry(vInsiders, Vec_IntEntry(vAdds, 6*i+4)) {
                continue; } }
        memmove( Vec_IntArray(vAdds) + 6*k++, Vec_IntArray(vAdds) + 6*i, 6*sizeof(int) );
    }
    assert(k <= i, "k <= i");
    Vec_IntShrink( vAdds, 6*k );
    Vec_BitFree( vInsiders );
}

func Ree_ManCountFadds(vAdds *Vec_Int_t) int32 {
    var i, Count int32;
    for  i = 0; 6*i < Vec_IntSize(vAdds); i++  {
        if Vec_IntEntry(vAdds, 6*i+2) != 0 {
            Count++; }
    }
    return Count;
}
func Ree_ManPrintAdders(vAdds *Vec_Int_t, fVerbose int32) {
    var i int32;
    for  i = 0; 6*i < Vec_IntSize(vAdds); i++  {
        //if ( Vec_IntEntry(vAdds, 6*i+2) == 0 )
        //    continue;
        if !(fVerbose != 0) {
            continue; }
        fmt.Printf("%6d : ", i);
        fmt.Printf("%6d ", Vec_IntEntry(vAdds, 6*i+0));
        fmt.Printf("%6d ", Vec_IntEntry(vAdds, 6*i+1));
        fmt.Printf("%6d ", Vec_IntEntry(vAdds, 6*i+2));
        fmt.Printf("   ->  ");
        fmt.Printf("%6d ", Vec_IntEntry(vAdds, 6*i+3));
        fmt.Printf("%6d ", Vec_IntEntry(vAdds, 6*i+4));
        fmt.Printf("  (%d)", Vec_IntEntry(vAdds, 6*i+5));
        fmt.Printf("\n");
    }
}
func Ree_ManComputeCutsTest(p *Gia_Man_t) {
    var clk abctime  = Abc_Clock();
    var vAdds *Vec_Int_t  = Ree_ManComputeCuts( p, nil, 1 );
    var nFadds int32  = Ree_ManCountFadds( vAdds );
    Ree_ManPrintAdders( vAdds, 1 );
    fmt.Printf("Detected %d FAs and %d HAs.  ", nFadds, Vec_IntSize(vAdds)/6-nFadds);
    Vec_IntFree( vAdds );
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


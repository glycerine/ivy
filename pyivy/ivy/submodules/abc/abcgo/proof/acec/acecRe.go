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


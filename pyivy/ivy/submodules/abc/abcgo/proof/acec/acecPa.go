package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecPa.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecPa.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecWec.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pas_ManVerifyPhaseOne_rec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    var Truth0, Truth1 int32;
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return pObj.Value; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    assert(!Gia_ObjIsXor(pObj), "!Gia_ObjIsXor(pObj)");
    Truth0 = Pas_ManVerifyPhaseOne_rec( p, Gia_ObjFanin0(pObj) );
    Truth1 = Pas_ManVerifyPhaseOne_rec( p, Gia_ObjFanin1(pObj) );
    if Gia_ObjFaninC0(pObj) {
        Truth0 = 0xFF & ~Truth0;
    } else {
        Truth0 = Truth0;
    }
    if Gia_ObjFaninC1(pObj) {
        Truth1 = 0xFF & ~Truth1;
    } else {
        Truth1 = Truth1;
    }
    return (pObj.Value = Truth0 & Truth1);
}
func Pas_ManVerifyPhaseOne(p *Gia_Man_t, vAdds *Vec_Int_t, iBox int32, vPhase *Vec_Bit_t) {
    var pObj *Gia_Obj_t;
    var TruthXor uint32;
    var TruthMaj uint32;
    var Truths [3]uint32 = [3]uint32{ 0xAA, 0xCC, 0xF0 };
    var k int32;
    var iObj int32;
    var fFadd int32 = Vec_IntEntry(vAdds, 6*iBox+2) > 0;

    if !(fFadd != 0) {
        return; }

    Gia_ManIncrementTravId( p );
    for  k = 0; k < 3; k++  {
        iObj = Vec_IntEntry( vAdds, 6*iBox+k );
        if iObj == 0 {
            continue; }
        pObj = Gia_ManObj( p, iObj );
        if Vec_BitEntry(vPhase, iObj) {
            pObj.Value = 0xFF & ~Truths[k];
        } else {
            pObj.Value = Truths[k];
        }
        Gia_ObjSetTravIdCurrent( p, pObj );
    }

    iObj = Vec_IntEntry( vAdds, 6*iBox+3 );
    TruthXor = Pas_ManVerifyPhaseOne_rec( p, Gia_ManObj(p, iObj) );
    if Vec_BitEntry(vPhase, iObj) {
        TruthXor = 0xFF & ~TruthXor;
    } else {
        TruthXor = TruthXor;
    }

    iObj = Vec_IntEntry( vAdds, 6*iBox+4 );
    TruthMaj = Pas_ManVerifyPhaseOne_rec( p, Gia_ManObj(p, iObj) );
    if Vec_BitEntry(vPhase, iObj) {
        TruthMaj = 0xFF & ~TruthMaj;
    } else {
        TruthMaj = TruthMaj;
    }

    if fFadd != 0 { // FADD
    
        if TruthXor != 0x96 {
            fmt.Printf("Fadd %d sum is wrong.\n", iBox); }
        if TruthMaj != 0xE8 {
            fmt.Printf("Fadd %d carry is wrong.\n", iBox); }
    
    } else {
        if TruthXor != 0x66 {
            fmt.Printf("Hadd %d sum is wrong.\n", iBox); }
        if TruthMaj != 0x88 {
            fmt.Printf("Hadd %d carry is wrong.\n", iBox); }
    }
}
func Pas_ManVerifyPhase(p *Gia_Man_t, vAdds *Vec_Int_t, vOrder *Vec_Int_t, vPhase *Vec_Bit_t) {
    var k, iBox int32;
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vOrder)) && (((iBox) = Vec_IntEntry(vOrder, k)), 1); k++  {
        Pas_ManVerifyPhaseOne( p, vAdds, iBox, vPhase );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pas_ManPhase_rec(p *Gia_Man_t, vAdds *Vec_Int_t, vMap *Vec_Int_t, pObj *Gia_Obj_t, fPhase int32, vPhase *Vec_Bit_t, vConstPhase *Vec_Bit_t) {
    var k, iBox, iXor, Sign, fXorPhase int32;
    assert(pObj != Gia_ManConst0(p), "pObj != Gia_ManConst0(p)");
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if fPhase != 0 {
        Vec_BitWriteEntry( vPhase, Gia_ObjId(p, pObj), fPhase ); }
    if !Gia_ObjIsAnd(pObj) {
        return; }
    iBox = Vec_IntEntry( vMap, Gia_ObjId(p, pObj) );
    if iBox == -1 {
        return; }
    iXor = Vec_IntEntry( vAdds, 6*iBox+3 );
    Sign = Vec_IntEntry( vAdds, 6*iBox+5 );
    fXorPhase = ((Sign >> 3) & 1);
    // remember complemented HADD
    if Vec_IntEntry(vAdds, 6*iBox+2) == 0 && (fPhase != 0) {
        Vec_BitWriteEntry( vConstPhase, iBox, 1 ); }
    for  k = 0; k < 3; k++  {
        var iObj int32  = Vec_IntEntry( vAdds, 6*iBox+k );
        var fPhaseThis int32  = ((Sign >> k) & 1) ^ fPhase;
        fXorPhase ^= fPhaseThis;
        if iObj == 0 {
            continue; }
        Pas_ManPhase_rec( p, vAdds, vMap, Gia_ManObj(p, iObj), fPhaseThis, vPhase, vConstPhase );
    }
    Vec_BitWriteEntry( vPhase, iXor, fXorPhase );
}
func Pas_ManPhase(p *Gia_Man_t, vAdds *Vec_Int_t, vMap *Vec_Int_t, vRoots *Vec_Int_t, pvConstPhase **Vec_Bit_t) *Vec_Bit_t {
    var vPhase *Vec_Bit_t  = Vec_BitStart( Vec_IntSize(vMap) );
    var vConstPhase *Vec_Bit_t  = Vec_BitStart( Vec_IntSize(vAdds)/6 );
    var i, iRoot int32;
    Gia_ManIncrementTravId( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRoots)) && (((iRoot) = Vec_IntEntry(vRoots, i)), 1); i++  {
        Pas_ManPhase_rec( p, vAdds, vMap, Gia_ManObj(p, iRoot), 1, vPhase, vConstPhase );
    }
    *pvConstPhase = vConstPhase;
    return vPhase;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pas_ManComputeCuts(p *Gia_Man_t, vAdds *Vec_Int_t, vOrder *Vec_Int_t, vIns *Vec_Int_t, vOuts *Vec_Int_t) int32 {
    var vUnique *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var vUsed *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var vMap *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    var vRoots *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vPhase, vConstPhase *Vec_Bit_t;
    var i, k, Entry, nTrees int32;

    // map carries into adder indexes and mark driven nodes
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vOrder)) && (((i) = Vec_IntEntry(vOrder, k)), 1); k++  {
        var Carry int32  = Vec_IntEntry(vAdds, 6*i+4);
        if Vec_BitEntry(vUnique, Carry) {
            fmt.Printf("Carry %d participates more than once.\n", Carry); }
        Vec_BitWriteEntry( vUnique, Carry, 1 );
        Vec_IntWriteEntry( vMap, Carry, i );
        // mark driven 
        Vec_BitWriteEntry( vUsed, Vec_IntEntry(vAdds, 6*i+0), 1 );
        Vec_BitWriteEntry( vUsed, Vec_IntEntry(vAdds, 6*i+1), 1 );
        Vec_BitWriteEntry( vUsed, Vec_IntEntry(vAdds, 6*i+2), 1 );
    }
    // collect carries that do not drive other adders
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        if Vec_BitEntry(vUnique, i) && !Vec_BitEntry(vUsed, i) {
            Vec_IntPush( vRoots, i ); }
    }
    nTrees = Vec_IntSize( vRoots );

    Vec_IntPrint( vRoots );

    // compute phases
    if Vec_IntSize(vRoots) > 0 {
        var nCompls int32;

        vPhase = Pas_ManPhase( p, vAdds, vMap, vRoots, &vConstPhase );
        Pas_ManVerifyPhase( p, vAdds, vOrder, vPhase );

        fmt.Printf("Outputs: ");
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vOuts)) && (((Entry) = Vec_IntEntry(vOuts, i)), 1); i++  {
            fmt.Printf("%d(%d) ", Entry, Vec_BitEntry(vPhase, Entry));
        }
        fmt.Printf("\n");

        fmt.Printf("Inputs: ");
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vIns)) && (((Entry) = Vec_IntEntry(vIns, i)), 1); i++  {
            fmt.Printf("%d(%d) ", Entry, Vec_BitEntry(vPhase, Entry));
            nCompls += Vec_BitEntry(vPhase, Entry);
        }
        fmt.Printf("   Compl = %d\n", nCompls);

        Vec_BitFreeP( &vPhase );
        Vec_BitFreeP( &vConstPhase );
    }

    Vec_IntFree( vRoots );
    Vec_IntFree( vMap );
    Vec_BitFree( vUnique );
    Vec_BitFree( vUsed );
    return nTrees;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pas_ManComputeCutsTest(p *Gia_Man_t) {
    var clk abctime  = Abc_Clock();
    var vAdds *Vec_Int_t  = Ree_ManComputeCuts( p, nil, 1 );
    var vIns, vOuts *Vec_Int_t;
    var vOrder *Vec_Int_t  = Gia_PolynCoreOrder( p, vAdds, nil, &vIns, &vOuts );
    var nTrees int32;
    var nFadds int32 = Ree_ManCountFadds( vAdds );
    //Ree_ManPrintAdders( vAdds, 1 );
    fmt.Printf("Detected %d FAs and %d HAs.  Collected %d adders.  ", nFadds, Vec_IntSize(vAdds)/6-nFadds, Vec_IntSize(vOrder));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    // detect trees
    clk = Abc_Clock();
    nTrees = Pas_ManComputeCuts( p, vAdds, vOrder, vIns, vOuts );
    Vec_IntFree( vAdds );
    Vec_IntFree( vOrder );
    Vec_IntFree( vIns );
    Vec_IntFree( vOuts );

    fmt.Printf("Detected %d adder trees. ", nTrees);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


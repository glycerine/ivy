package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecFadds.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Detecting half-adders and full-adders.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecFadds.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecWec.h"
//#include "misc/tim/tim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Dtc_ForEachCut( pList, pCut, i ) for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += pCut[0] + 1 )

// Original macro:
// #define Dtc_ForEachFadd( vFadds, i )     for ( i = 0; i < Vec_IntSize(vFadds)/5; i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Detecting HADDs in the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIllustrateBoxes(p *Gia_Man_t) {
    var pManTime *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
    var nBoxes int32  = Tim_ManBoxNum( pManTime );
    var i, k, curCi, curCo, nBoxIns, nBoxOuts int32;
    var pObj *Gia_Obj_t;
    // walk through the boxes
    curCi = Tim_ManPiNum(pManTime);
    curCo = 0;
    for  i = 0; i < nBoxes; i++  {
        nBoxIns = Tim_ManBoxInputNum(pManTime, i);
        nBoxOuts = Tim_ManBoxOutputNum(pManTime, i);
        fmt.Printf("Box %4d  [%d x %d] :   ", i, nBoxIns, nBoxOuts);
        fmt.Printf("Input obj IDs = ");
        for  k = 0; k < nBoxIns; k++  {
            pObj = Gia_ManCo( p, curCo + k );
            fmt.Printf("%d ", Gia_ObjId(p, pObj));
        }
        fmt.Printf("  Output obj IDs = ");
        for  k = 0; k < nBoxOuts; k++  {
            pObj = Gia_ManCi( p, curCi + k );
            fmt.Printf("%d ", Gia_ObjId(p, pObj));
        }
        curCo += nBoxIns;
        curCi += nBoxOuts;
        fmt.Printf("\n");
    }
    curCo += Tim_ManPoNum(pManTime);
    // verify counts
    assert(curCi == Gia_ManCiNum(p), "curCi == Gia_ManCiNum(p)");
    assert(curCo == Gia_ManCoNum(p), "curCo == Gia_ManCoNum(p)");
}

/**Function*************************************************************

  Synopsis    [Detecting FADDs in the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dtc_ManCutMergeOne(pCut0 *int32, pCut1 *int32, pCut *int32) int32 {
    var i, k int32;
    for  k = 0; k <= pCut1[0]; k++  {
        pCut[k] = pCut1[k];
    }
    for  i = 1; i <= pCut0[0]; i++  {
        for  k = 1; k <= pCut1[0]; k++  {
            if pCut0[i] == pCut1[k] {
                break; }
        }
        if k <= pCut1[0] {
            continue; }
        if pCut[0] == 3 {
            return 0; }
        pCut[1+pCut[0]++] = pCut0[i];
    }
    assert(pCut[0] == 2 || pCut[0] == 3, "pCut[0] == 2 || pCut[0] == 3");
    if pCut[1] > pCut[2] {
        // inlining of macro ABC_SWAP
         var t int32  = pCut[1]; pCut[1] = pCut[2]; pCut[2] = t; };
    assert(pCut[1] < pCut[2], "pCut[1] < pCut[2]");
    if pCut[0] == 2 {
        return 1; }
    if pCut[2] > pCut[3] {
        // inlining of macro ABC_SWAP
         var t int32  = pCut[2]; pCut[2] = pCut[3]; pCut[3] = t; };
    if pCut[1] > pCut[2] {
        // inlining of macro ABC_SWAP
         var t int32  = pCut[1]; pCut[1] = pCut[2]; pCut[2] = t; };
    assert(pCut[1] < pCut[2], "pCut[1] < pCut[2]");
    assert(pCut[2] < pCut[3], "pCut[2] < pCut[3]");
    return 1;
}
func Dtc_ManCutCheckEqual(vCuts *Vec_Int_t, pCutNew *int32) int32 {
    var pList *int32  = Vec_IntArray( vCuts );
    var i int32;
    var k int32;
    var pCut *int32;
    // inlining of macro Dtc_ForEachCut
    for  i, pCut = 0, pList + 1; i < pList[0]; i++, pCut += pCut[0] + 1  {
        for  k = 0; k <= pCut[0]; k++  {
            if pCut[k] != pCutNew[k] {
                break; }
        }
        if k > pCut[0] {
            return 1; }
    }
    return 0;
}
func Dtc_ObjComputeTruth_rec(pObj *Gia_Obj_t) int32 {
    var Truth0, Truth1 int32;
    if pObj.Value {
        return pObj.Value; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Truth0 = Dtc_ObjComputeTruth_rec( Gia_ObjFanin0(pObj) );
    Truth1 = Dtc_ObjComputeTruth_rec( Gia_ObjFanin1(pObj) );
    if Gia_ObjIsXor(pObj) {
        return (pObj.Value = (Gia_ObjFaninC0(pObj) ? ~Truth0 : Truth0) ^ (Gia_ObjFaninC1(pObj) ? ~Truth1 : Truth1));
    } else {
        return (pObj.Value = (Gia_ObjFaninC0(pObj) ? ~Truth0 : Truth0) & (Gia_ObjFaninC1(pObj) ? ~Truth1 : Truth1)); }
}
func Dtc_ObjCleanTruth_rec(pObj *Gia_Obj_t) {
    if !pObj.Value {
        return; }
    pObj.Value = 0;
    if !Gia_ObjIsAnd(pObj) {
        return; }
    Dtc_ObjCleanTruth_rec( Gia_ObjFanin0(pObj) );
    Dtc_ObjCleanTruth_rec( Gia_ObjFanin1(pObj) );
}
func Dtc_ObjComputeTruth(p *Gia_Man_t, iObj int32, pCut *int32, pTruth *int32) int32 {
    var Truth uint32;
    var Truths [3]uint32 = [3]uint32{ 0xAA, 0xCC, 0xF0 }; var i int32;
    for  i = 1; i <= pCut[0]; i++  {
        Gia_ManObj(p, pCut[i]).Value = Truths[i-1];
    }
    Truth = 0xFF & Dtc_ObjComputeTruth_rec( Gia_ManObj(p, iObj) );
    Dtc_ObjCleanTruth_rec( Gia_ManObj(p, iObj) );
    if pTruth != nil { 
        *pTruth = Truth; }
    if Truth == 0x66 || Truth == 0x99 {
        return 3; }
    if Truth == 0x96 || Truth == 0x69 {
        return 1; }
    if Truth == 0xE8 || Truth == 0xD4 || Truth == 0xB2 || Truth == 0x71 ||
         Truth == 0x17 || Truth == 0x2B || Truth == 0x4D || Truth == 0x8E {
        return 2; }
    return 0;
}
func Dtc_ManCutMerge(p *Gia_Man_t, iObj int32, pList0 *int32, pList1 *int32, vCuts *Vec_Int_t, vCutsXor2 *Vec_Int_t, vCutsXor *Vec_Int_t, vCutsMaj *Vec_Int_t) {
    var fVerbose int32;
    var vTemp *Vec_Int_t;
    var i int32;
    var k int32;
    var c int32;
    var Type int32;
    var pCut0 *int32;
    var pCut1 *int32;
    var pCut [4]int32;
    if fVerbose != 0 {
        fmt.Printf("Object %d = :\n", iObj); }
    Vec_IntFill( vCuts, 2, 1 );
    Vec_IntPush( vCuts, iObj );
    // inlining of macro Dtc_ForEachCut
    for  i, pCut0 = 0, pList0 + 1; i < pList0[0]; i++, pCut0 += pCut0[0] + 1  {
    // inlining of macro Dtc_ForEachCut
    for  k, pCut1 = 0, pList1 + 1; k < pList1[0]; k++, pCut1 += pCut1[0] + 1  {
        if !(Dtc_ManCutMergeOne(pCut0, pCut1, pCut) != 0) {
            continue; }
        if Dtc_ManCutCheckEqual(vCuts, pCut) != 0 {
            continue; }
        Vec_IntAddToEntry( vCuts, 0, 1 );  
        if fVerbose != 0 {
            fmt.Printf("%d : ", pCut[0]); }
        for  c = 0; c <= pCut[0]; c++  {
            Vec_IntPush( vCuts, pCut[c] );
            if (fVerbose != 0) && (c != 0) {
                fmt.Printf("%d ", pCut[c]); }
        }
        if fVerbose != 0 {
            fmt.Printf("\n"); }
        if pCut[0] == 2 {
            var Value int32  = Dtc_ObjComputeTruth( p, iObj, pCut, nil );
            assert(Value == 3 || Value == 0, "Value == 3 || Value == 0");
            if Value == 3 {
                Vec_IntPush( vCutsXor2, pCut[1] );
                Vec_IntPush( vCutsXor2, pCut[2] );
                Vec_IntPush( vCutsXor2, iObj );
            }
            continue;
        }
        if pCut[0] != 3 {
            continue; }
        Type = Dtc_ObjComputeTruth( p, iObj, pCut, nil );
        if Type == 0 {
            continue; }
        if Type == 1 {
            vTemp = vCutsXor;
        } else {
            vTemp = vCutsMaj;
        }
        if fVerbose != 0 {
            fmt.Printf("%d = %s(", iObj, Type == 1 ? "XOR" : "MAJ"); }
        for  c = 1; c <= pCut[0]; c++  {
            if fVerbose != 0 {
                fmt.Printf(" %d", pCut[c]); }
            Vec_IntPush( vTemp, pCut[c] );
        }
        if fVerbose != 0 {
            fmt.Printf(" )\n"); }
        Vec_IntPush( vTemp, iObj );
    }
    }
}
func Dtc_ManComputeCuts(p *Gia_Man_t, pvCutsXor2 **Vec_Int_t, pvCutsXor **Vec_Int_t, pvCutsMaj **Vec_Int_t, fVerbose int32) {
    var pObj *Gia_Obj_t; 
    var pList0 *int32;
    var pList1 *int32;
    var i int32;
    var nCuts int32;
    var vTemp *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vCutsXor2 *Vec_Int_t  = Vec_IntAlloc( Gia_ManAndNum(p) );
    var vCutsXor *Vec_Int_t  = Vec_IntAlloc( Gia_ManAndNum(p) );
    var vCutsMaj *Vec_Int_t  = Vec_IntAlloc( Gia_ManAndNum(p) );
    var vCuts *Vec_Int_t  = Vec_IntAlloc( 30 * Gia_ManAndNum(p) );
    Vec_IntFill( vCuts, Gia_ManObjNum(p), 0 );
    Gia_ManCleanValue( p );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Vec_IntWriteEntry( vCuts, Gia_ObjId(p, pObj), Vec_IntSize(vCuts) );
        Vec_IntPush( vCuts, 1 );
        Vec_IntPush( vCuts, 1 );
        Vec_IntPush( vCuts, Gia_ObjId(p, pObj) );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pList0 = Vec_IntEntryP( vCuts, Vec_IntEntry(vCuts, Gia_ObjFaninId0(pObj, i)) );
        pList1 = Vec_IntEntryP( vCuts, Vec_IntEntry(vCuts, Gia_ObjFaninId1(pObj, i)) );
        Dtc_ManCutMerge( p, i, pList0, pList1, vTemp, vCutsXor2, vCutsXor, vCutsMaj );
        Vec_IntWriteEntry( vCuts, i, Vec_IntSize(vCuts) );
        Vec_IntAppend( vCuts, vTemp );
        nCuts += Vec_IntEntry( vTemp, 0 );
    }
    }
    if fVerbose != 0 {
        fmt.Printf("Nodes = %d.  Cuts = %d.  Cuts/Node = %.2f.  Ints/Node = %.2f.\n", Gia_ManAndNum(p), nCuts, 1.0*nCuts/Gia_ManAndNum(p), 1.0*Vec_IntSize(vCuts)/Gia_ManAndNum(p)); }
    Vec_IntFree( vTemp );
    Vec_IntFree( vCuts );
    if pvCutsXor2 {
        *pvCutsXor2 = vCutsXor2;
    } else {
        Vec_IntFree( vCutsXor2 ); }
    *pvCutsXor = vCutsXor;
    *pvCutsMaj = vCutsMaj;
}
func Dtc_ManFindCommonCuts(p *Gia_Man_t, vCutsXor *Vec_Int_t, vCutsMaj *Vec_Int_t) *Vec_Int_t {
    var pCuts0 *int32  = Vec_IntArray(vCutsXor);
    var pCuts1 *int32  = Vec_IntArray(vCutsMaj);
    var pLimit0 *int32  = Vec_IntLimit(vCutsXor);
    var pLimit1 *int32  = Vec_IntLimit(vCutsMaj);   var i int32;
    var vFadds *Vec_Int_t  = Vec_IntAlloc( 1000 );
    assert(Vec_IntSize(vCutsXor) % 4 == 0, "Vec_IntSize(vCutsXor) % 4 == 0");
    assert(Vec_IntSize(vCutsMaj) % 4 == 0, "Vec_IntSize(vCutsMaj) % 4 == 0");
    for  pCuts0 < pLimit0 && pCuts1 < pLimit1  {
        for  i = 0; i < 3; i++  {
            if pCuts0[i] != pCuts1[i] {
                break; }
        }
        if i == 3 {
            for  i = 0; i < 4; i++  {
                Vec_IntPush( vFadds, pCuts0[i] );
            }
            Vec_IntPush( vFadds, pCuts1[3] );
            pCuts0 += 4;
            pCuts1 += 4;
        
        } else if pCuts0[i] < pCuts1[i] {
            pCuts0 += 4;
        } else if pCuts0[i] > pCuts1[i] {
            pCuts1 += 4; }
    }
    assert(Vec_IntSize(vFadds) % 5 == 0, "Vec_IntSize(vFadds) % 5 == 0");
    return vFadds;
}
func Dtc_ManPrintFadds(vFadds *Vec_Int_t) {
    var i int32;
    // inlining of macro Dtc_ForEachFadd
    for  i = 0; i < Vec_IntSize(vFadds)/5; i++  {
        fmt.Printf("%6d : ", i);
        fmt.Printf("%6d ", Vec_IntEntry(vFadds, 5*i+0));
        fmt.Printf("%6d ", Vec_IntEntry(vFadds, 5*i+1));
        fmt.Printf("%6d ", Vec_IntEntry(vFadds, 5*i+2));
        fmt.Printf(" ->  ");
        fmt.Printf("%6d ", Vec_IntEntry(vFadds, 5*i+3));
        fmt.Printf("%6d ", Vec_IntEntry(vFadds, 5*i+4));
        fmt.Printf("\n");

        if i == 100 {
            fmt.Printf("Skipping other FADDs.\n");
            break;
        }
    }
}
func Dtc_ManCompare(pCut0 *int32, pCut1 *int32) int32 {
    if pCut0[0] < pCut1[0] { return -1; }
    if pCut0[0] > pCut1[0] { return  1; }
    if pCut0[1] < pCut1[1] { return -1; }
    if pCut0[1] > pCut1[1] { return  1; }
    if pCut0[2] < pCut1[2] { return -1; }
    if pCut0[2] > pCut1[2] { return  1; }
    return 0;
}
func Dtc_ManCompare2(pCut0 *int32, pCut1 *int32) int32 {
    if pCut0[4] < pCut1[4] { return -1; }
    if pCut0[4] > pCut1[4] { return  1; }
    return 0;
}
// returns array of 5-tuples containing inputs/sum/cout of each full adder
func Gia_ManDetectFullAdders(p *Gia_Man_t, fVerbose int32, pvCutsXor2 **Vec_Int_t) *Vec_Int_t {
    var vCutsXor, vCutsMaj, vFadds *Vec_Int_t;
    Dtc_ManComputeCuts( p, pvCutsXor2, &vCutsXor, &vCutsMaj, fVerbose );
    qsort( Vec_IntArray(vCutsXor), (size_t)(Vec_IntSize(vCutsXor)/4), 16, (int (*)(void *, void *))Dtc_ManCompare );
    qsort( Vec_IntArray(vCutsMaj), (size_t)(Vec_IntSize(vCutsMaj)/4), 16, (int (*)(void *, void *))Dtc_ManCompare );
    vFadds = Dtc_ManFindCommonCuts( p, vCutsXor, vCutsMaj );
    qsort( Vec_IntArray(vFadds), (size_t)(Vec_IntSize(vFadds)/5), 20, (int (*)(void *, void *))Dtc_ManCompare2 );
    if fVerbose != 0 {
        fmt.Printf("XOR3 cuts = %d.  MAJ cuts = %d.  Full-adders = %d.\n", Vec_IntSize(vCutsXor)/4, Vec_IntSize(vCutsMaj)/4, Vec_IntSize(vFadds)/5); }
    if fVerbose != 0 {
        Dtc_ManPrintFadds( vFadds ); }
    Vec_IntFree( vCutsXor );
    Vec_IntFree( vCutsMaj );
    return vFadds;
}

/**Function*************************************************************

  Synopsis    [Map each MAJ into the topmost MAJ of its chain.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// maps MAJ nodes into FADD indexes
func Gia_ManCreateMap(p *Gia_Man_t, vFadds *Vec_Int_t) *Vec_Int_t {
    var vMap *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );  var i int32;
    // inlining of macro Dtc_ForEachFadd
    for  i = 0; i < Vec_IntSize(vFadds)/5; i++  {
        Vec_IntWriteEntry( vMap, Vec_IntEntry(vFadds, 5*i+4), i );
    }
    return vMap;
}
// find chain length (for each MAJ, how many FADDs are rooted in its first input)
func Gia_ManFindChains_rec(p *Gia_Man_t, iMaj int32, vFadds *Vec_Int_t, vMap *Vec_Int_t, vLength *Vec_Int_t) int32 {
    assert(Vec_IntEntry(vMap, iMaj) >= 0, "Vec_IntEntry(vMap, iMaj) >= 0"); // MAJ
    if Vec_IntEntry(vLength, iMaj) >= 0 {
        return Vec_IntEntry(vLength, iMaj); }
    assert(Gia_ObjIsAnd(Gia_ManObj(p, iMaj)), "Gia_ObjIsAnd(Gia_ManObj(p, iMaj))");
    {
        var iFadd int32  = Vec_IntEntry( vMap, iMaj );
        var iXor0 int32  = Vec_IntEntry( vFadds, 5*iFadd+0 );
        var iXor1 int32  = Vec_IntEntry( vFadds, 5*iFadd+1 );
        var iXor2 int32  = Vec_IntEntry( vFadds, 5*iFadd+2 );
        var iLen0 int32;
        if Vec_IntEntry( vMap, iXor0 ) == -1 {
            iLen0 = 0;
        } else {
            iLen0 = Gia_ManFindChains_rec( p, iXor0, vFadds, vMap, vLength );
        }
        var iLen1 int32;
        if Vec_IntEntry( vMap, iXor1 ) == -1 {
            iLen1 = 0;
        } else {
            iLen1 = Gia_ManFindChains_rec( p, iXor1, vFadds, vMap, vLength );
        }
        var iLen2 int32;
        if Vec_IntEntry( vMap, iXor2 ) == -1 {
            iLen2 = 0;
        } else {
            iLen2 = Gia_ManFindChains_rec( p, iXor2, vFadds, vMap, vLength );
        }
        var iLen int32  = Abc_MaxInt( iLen0, Abc_MaxInt(iLen1, iLen2) );
        if iLen0 < iLen {
            if iLen == iLen1 {
                // inlining of macro ABC_SWAP
                { var t int32  = Vec_IntArray(vFadds)[5*iFadd+0]; Vec_IntArray(vFadds)[5*iFadd+0] = Vec_IntArray(vFadds)[5*iFadd+1]; Vec_IntArray(vFadds)[5*iFadd+1] = t; };
            
            } else if iLen == iLen2 {
                // inlining of macro ABC_SWAP
                { var t int32  = Vec_IntArray(vFadds)[5*iFadd+0]; Vec_IntArray(vFadds)[5*iFadd+0] = Vec_IntArray(vFadds)[5*iFadd+2]; Vec_IntArray(vFadds)[5*iFadd+2] = t; };
            }
        }
        Vec_IntWriteEntry( vLength, iMaj, iLen + 1 );
        return iLen + 1;
    }
}
// for each FADD find the longest chain and reorder its inputs
func Gia_ManFindChains(p *Gia_Man_t, vFadds *Vec_Int_t, vMap *Vec_Int_t) {
    var i int32;
    // for each FADD find the longest chain rooted in it
    var vLength *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    // inlining of macro Dtc_ForEachFadd
    for  i = 0; i < Vec_IntSize(vFadds)/5; i++  {
        Gia_ManFindChains_rec( p, Vec_IntEntry(vFadds, 5*i+4), vFadds, vMap, vLength );
    }
    Vec_IntFree( vLength );
}
// collect one carry-chain
func Gia_ManCollectOneChain(p *Gia_Man_t, vFadds *Vec_Int_t, iFaddTop int32, vMap *Vec_Int_t, vChain *Vec_Int_t) {
    var iFadd int32;
    Vec_IntClear( vChain );
    for  iFadd = iFaddTop; iFadd >= 0 && 
          !Gia_ObjIsTravIdCurrentId(p, Vec_IntEntry(vFadds, 5*iFadd+3)) && 
          !Gia_ObjIsTravIdCurrentId(p, Vec_IntEntry(vFadds, 5*iFadd+4)); 
          iFadd = Vec_IntEntry(vMap, Vec_IntEntry(vFadds, 5*iFadd+0))  {
                Vec_IntPush( vChain, iFadd );
          }
    Vec_IntReverseOrder( vChain );
}
func Gia_ManMarkWithTravId_rec(p *Gia_Man_t, Id int32) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, Id) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, Id);
    pObj = Gia_ManObj( p, Id );
    if Gia_ObjIsAnd(pObj) {
        Gia_ManMarkWithTravId_rec( p, Gia_ObjFaninId0(pObj, Id) ); }
    if Gia_ObjIsAnd(pObj) {
        Gia_ManMarkWithTravId_rec( p, Gia_ObjFaninId1(pObj, Id) ); }
}
// returns mapping of each MAJ into the topmost elements of its chain
func Gia_ManCollectTopmost(p *Gia_Man_t, vFadds *Vec_Int_t, vMap *Vec_Int_t, nFaddMin int32) *Vec_Wec_t {
    var i, j, iFadd int32;
    var vChain *Vec_Int_t   = Vec_IntAlloc( 100 );
    var vChains *Vec_Wec_t  = Vec_WecAlloc( Vec_IntSize(vFadds)/5 );
    // erase elements appearing as FADD inputs
    var vMarksTop *Vec_Bit_t  = Vec_BitStart( Vec_IntSize(vFadds)/5 );
    // inlining of macro Dtc_ForEachFadd
    for  i = 0; i < Vec_IntSize(vFadds)/5; i++  {
        if (iFadd = Vec_IntEntry(vMap, Vec_IntEntry(vFadds, 5*i+0))) >= 0 {
            Vec_BitWriteEntry( vMarksTop, iFadd, 1 ); }
    }
    // compress the remaining ones
    Gia_ManIncrementTravId( p );
    // inlining of macro Dtc_ForEachFadd
    for  i = 0; i < Vec_IntSize(vFadds)/5; i++  {
        if Vec_BitEntry(vMarksTop, i) {
            continue; }
        Gia_ManCollectOneChain( p, vFadds, i, vMap, vChain );
        if Vec_IntSize(vChain) < nFaddMin {
            continue; }
        Vec_IntAppend( Vec_WecPushLevel(vChains), vChain );
        // inlining of macro Vec_IntForEachEntry
        for  j = 0; (j < Vec_IntSize(vChain)) && (((iFadd) = Vec_IntEntry(vChain, j)), 1); j++  {
            assert(!Gia_ObjIsTravIdCurrentId(p, Vec_IntEntry(vFadds, 5*iFadd+3)), "!Gia_ObjIsTravIdCurrentId(p, Vec_IntEntry(vFadds, 5*iFadd+3))");
            assert(!Gia_ObjIsTravIdCurrentId(p, Vec_IntEntry(vFadds, 5*iFadd+4)), "!Gia_ObjIsTravIdCurrentId(p, Vec_IntEntry(vFadds, 5*iFadd+4))");
            Gia_ManMarkWithTravId_rec( p, Vec_IntEntry(vFadds, 5*iFadd+3) );
            Gia_ManMarkWithTravId_rec( p, Vec_IntEntry(vFadds, 5*iFadd+4) );
        }
    }
    // cleanup
    Vec_BitFree( vMarksTop );
    Vec_IntFree( vChain );
    return vChains;
}
// prints chains beginning in majority nodes contained in vTops
func Gia_ManPrintChains(p *Gia_Man_t, vFadds *Vec_Int_t, vMap *Vec_Int_t, vChains *Vec_Wec_t) {
    var vChain *Vec_Int_t;
    var i, k, iFadd, Count int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vChains)) && (((vChain) = Vec_WecEntry(vChains, i)), 1); i++  {
        Count += Vec_IntSize(vChain);
        if i < 10 {
            fmt.Printf("Chain %4d : %4d    ", i, Vec_IntSize(vChain));
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vChain)) && (((iFadd) = Vec_IntEntry(vChain, k)), 1); k++  {
                fmt.Printf("%d(%d) ", iFadd, Vec_IntEntry(vFadds, 5*iFadd+4));
                if k != Vec_IntSize(vChain) - 1 {
                    fmt.Printf("-> "); }
                if k > 6 {
                    fmt.Printf("...");
                    break;
                }
            }
            fmt.Printf("\n");
        
        } else if i == 10 {
            fmt.Printf("...\n"); }

    }
    fmt.Printf("Total chains = %d. Total full-adders = %d.\n", Vec_WecSize(vChains), Count);
}
// map SUM bits and topmost MAJ into topmost FADD number
func Gia_ManFindMapping(p *Gia_Man_t, vFadds *Vec_Int_t, vMap *Vec_Int_t, vChains *Vec_Wec_t) *Vec_Int_t {
    var vChain *Vec_Int_t;
    var i int32;
    var k int32;
    var iFadd int32 = -1;
    var vMap2Chain *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vChains)) && (((vChain) = Vec_WecEntry(vChains, i)), 1); i++  {
        assert(Vec_IntSize(vChain) > 0, "Vec_IntSize(vChain) > 0");
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vChain)) && (((iFadd) = Vec_IntEntry(vChain, k)), 1); k++  {
            //printf( "Chain %d: setting SUM %d (obj %d)\n", i, k, Vec_IntEntry(vFadds, 5*iFadd+3) );
            assert(Vec_IntEntry(vMap2Chain, Vec_IntEntry(vFadds, 5*iFadd+3)) == -1, "Vec_IntEntry(vMap2Chain, Vec_IntEntry(vFadds, 5*iFadd+3)) == -1");
            Vec_IntWriteEntry( vMap2Chain, Vec_IntEntry(vFadds, 5*iFadd+3), i );
        }
        //printf( "Chain %d: setting CARRY (obj %d)\n", i, Vec_IntEntry(vFadds, 5*iFadd+4) );
        assert(Vec_IntEntry(vMap2Chain, Vec_IntEntry(vFadds, 5*iFadd+4)) == -1, "Vec_IntEntry(vMap2Chain, Vec_IntEntry(vFadds, 5*iFadd+4)) == -1");
        Vec_IntWriteEntry( vMap2Chain, Vec_IntEntry(vFadds, 5*iFadd+4), i );
    }
    return vMap2Chain;
}

/**Function*************************************************************

  Synopsis    [Derive GIA with boxes containing adder-chains.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectTruthTables(p *Gia_Man_t, vFadds *Vec_Int_t) *Vec_Int_t {
    var i int32;
    var k int32;
    var Type int32;
    var Truth int32;
    var pCut [4]int32 = [4]int32{3};
    var vTruths *Vec_Int_t  = Vec_IntAlloc( 2*Vec_IntSize(vFadds)/5 );
    Gia_ManCleanValue( p );
    // inlining of macro Dtc_ForEachFadd
    for  i = 0; i < Vec_IntSize(vFadds)/5; i++  {
        for  k = 0; k < 3; k++  {
            pCut[k+1] = Vec_IntEntry( vFadds, 5*i+k );
        }
        Type = Dtc_ObjComputeTruth( p, Vec_IntEntry(vFadds, 5*i+3), pCut, &Truth );
        assert(Type == 1, "Type == 1");
        Vec_IntPush( vTruths, Truth );
        Type = Dtc_ObjComputeTruth( p, Vec_IntEntry(vFadds, 5*i+4), pCut, &Truth );
        assert(Type == 2, "Type == 2");
        Vec_IntPush( vTruths, Truth );
    }
    return vTruths;
}
func Gia_ManGenerateDelayTableFloat(nIns int32, nOuts int32) *float32 {
    var i int32;
    var Total int32 = nIns * nOuts;
    // inlining of macro ABC_ALLOC
    var pDelayTable *float32  = ((float *) malloc(sizeof(float) * (size_t)(Total + 3)));
    pDelayTable[0] = 0;
    pDelayTable[1] = nIns;
    pDelayTable[2] = nOuts;
    for  i = 0; i < Total; i++  {
        pDelayTable[i+3] = 1;
    }
    pDelayTable[i+3 - nIns] = -ABC_INFINITY;
    return pDelayTable;
}
func Gia_ManGenerateTim(nPis int32, nPos int32, nBoxes int32, nIns int32, nOuts int32) *Tim_Man_t {
    var pMan *Tim_Man_t;
    var i, curPi, curPo int32;
    var vDelayTables *Vec_Ptr_t  = Vec_PtrAlloc( 1 );
    Vec_PtrPush( vDelayTables, Gia_ManGenerateDelayTableFloat(nIns, nOuts) );
    pMan = Tim_ManStart( nPis + nOuts * nBoxes, nPos + nIns * nBoxes );
    Tim_ManSetDelayTables( pMan, vDelayTables );
    curPi = nPis;
    curPo = 0;
    for  i = 0; i < nBoxes; i++  {
        Tim_ManCreateBox( pMan, curPo, nIns, curPi, nOuts, 0, 0 );
        curPi += nOuts;
        curPo += nIns;
    }
    curPo += nPos;
    assert(curPi == Tim_ManCiNum(pMan), "curPi == Tim_ManCiNum(pMan)");
    assert(curPo == Tim_ManCoNum(pMan), "curPo == Tim_ManCoNum(pMan)");
    //Tim_ManPrint( pMan );
    return pMan;
}
func Gia_ManGenerateExtraAig(nBoxes int32, nIns int32, nOuts int32) *Gia_Man_t {
    var pNew *Gia_Man_t  = Gia_ManStart( nBoxes * 20 );
    var i int32;
    var k int32;
    var pInLits [16]int32;
    var pOutLits [16]int32;
    assert(nIns < 16 && nOuts < 16, "nIns < 16 && nOuts < 16");
    for  i = 0; i < nIns; i++  {
        pInLits[i] = Gia_ManAppendCi( pNew );
    }
    pOutLits[0] = Gia_ManAppendXor( pNew, Gia_ManAppendXor(pNew, pInLits[0], pInLits[1]), pInLits[2] );
    pOutLits[1] = Gia_ManAppendMaj( pNew, pInLits[0], pInLits[1], pInLits[2] );
    for  i = 0; i < nBoxes; i++  {
        for  k = 0; k < nOuts; k++  {
            Gia_ManAppendCo( pNew, pOutLits[k] );
        }
    }
    return pNew;
}
func Gia_ManDupFadd(pNew *Gia_Man_t, p *Gia_Man_t, vChain *Vec_Int_t, vFadds *Vec_Int_t, vMap *Vec_Int_t, vChains *Vec_Wec_t, vMap2Chain *Vec_Int_t, vTruths *Vec_Int_t) {
    extern void Gia_ManDupWithFaddBoxes_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vFadds, Vec_Int_t * vMap, Vec_Wec_t * vChains, Vec_Int_t * vMap2Chain, Vec_Int_t * vTruths );
    var i int32;
    var k int32;
    var iFadd int32 = -1;
    var iCiLit int32;
    var pLits [3]int32;
    var pObj *Gia_Obj_t;
    // construct FADD inputs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vChain)) && (((iFadd) = Vec_IntEntry(vChain, i)), 1); i++  {
        for  k = 0; k < 3; k++  {
            if (i != 0) && !(k != 0) { continue; }
            pObj = Gia_ManObj( p, Vec_IntEntry(vFadds, 5*iFadd+k) );
            Gia_ManDupWithFaddBoxes_rec( pNew, p, pObj, vFadds, vMap, vChains, vMap2Chain, vTruths );
        }
    }
    // construct boxes
    iCiLit = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vChain)) && (((iFadd) = Vec_IntEntry(vChain, i)), 1); i++  {
        var iXorTruth int32  = Vec_IntEntry( vTruths, 2*iFadd+0 );
        var iMajTruth int32  = Vec_IntEntry( vTruths, 2*iFadd+1 );
        for  k = 0; k < 3; k++  {
            pObj = Gia_ManObj( p, Vec_IntEntry(vFadds, 5*iFadd+k) );
            pLits[k] = (!k && iCiLit) ? iCiLit : pObj.Value;
            assert(pLits[k] >= 0, "pLits[k] >= 0");
        }
        // normalize truth table
        //    if ( Truth == 0xE8 || Truth == 0xD4 || Truth == 0xB2 || Truth == 0x71 ||
        //         Truth == 0x17 || Truth == 0x2B || Truth == 0x4D || Truth == 0x8E )
        if iMajTruth == 0x4D {
            pLits[0] = Abc_LitNot(pLits[0]), iMajTruth = 0x8E, iXorTruth = 0xFF & ~iXorTruth;
        } else if iMajTruth == 0xD4 {
            pLits[0] = Abc_LitNot(pLits[0]), iMajTruth = 0xE8, iXorTruth = 0xFF & ~iXorTruth;
        } else if iMajTruth == 0x2B {
            pLits[1] = Abc_LitNot(pLits[1]), iMajTruth = 0x8E, iXorTruth = 0xFF & ~iXorTruth;
        } else if iMajTruth == 0xB2 {
            pLits[1] = Abc_LitNot(pLits[1]), iMajTruth = 0xE8, iXorTruth = 0xFF & ~iXorTruth; }
        if iMajTruth == 0x8E {
            pLits[2] = Abc_LitNot(pLits[2]), iMajTruth = 0xE8, iXorTruth = 0xFF & ~iXorTruth;
        } else if iMajTruth == 0x71 {
            pLits[2] = Abc_LitNot(pLits[2]), iMajTruth = 0x17, iXorTruth = 0xFF & ~iXorTruth;
        } else { assert(iMajTruth == 0xE8 || iMajTruth == 0x17, "iMajTruth == 0xE8 || iMajTruth == 0x17"); }
        // normalize carry-in
        if Abc_LitIsCompl(pLits[0]) {
            for  k = 0; k < 3; k++  {
                pLits[k] = Abc_LitNot(pLits[k]);
            }
            iXorTruth = 0xFF & ~iXorTruth;
            iMajTruth = 0xFF & ~iMajTruth;
        }
        // add COs
        assert(!Abc_LitIsCompl(pLits[0]), "!Abc_LitIsCompl(pLits[0])");
        for  k = 0; k < 3; k++  {
            Gia_ManAppendCo( pNew, pLits[k] );
        }
        // create CI
        assert(iXorTruth == 0x96 || iXorTruth == 0x69, "iXorTruth == 0x96 || iXorTruth == 0x69");
        pObj = Gia_ManObj( p, Vec_IntEntry(vFadds, 5*iFadd+3) );
        pObj.Value = Abc_LitNotCond( Gia_ManAppendCi(pNew), int32((iXorTruth == 0x69)) );
        // create CI
        assert(iMajTruth == 0xE8 || iMajTruth == 0x17, "iMajTruth == 0xE8 || iMajTruth == 0x17");
        iCiLit = Abc_LitNotCond( Gia_ManAppendCi(pNew), int32((iMajTruth == 0x17)) );
    }   
    // assign carry out
    assert(iFadd == Vec_IntEntryLast(vChain), "iFadd == Vec_IntEntryLast(vChain)");
    pObj = Gia_ManObj( p, Vec_IntEntry(vFadds, 5*iFadd+4) );
    pObj.Value = iCiLit;
}
func Gia_ManDupWithFaddBoxes_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, vFadds *Vec_Int_t, vMap *Vec_Int_t, vChains *Vec_Wec_t, vMap2Chain *Vec_Int_t, vTruths *Vec_Int_t) {
    var iChain int32;
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    iChain = Vec_IntEntry( vMap2Chain, Gia_ObjId(p, pObj) );
/*
    assert( iChain == -1 );
    if ( iChain >= 0 )
    {
        Gia_ManDupFadd( pNew, p, Vec_WecEntry(vChains, iChain), vFadds, vMap, vChains, vMap2Chain, vTruths );
        assert( ~pObj->Value );
        return;
    }
*/
    Gia_ManDupWithFaddBoxes_rec( pNew, p, Gia_ObjFanin0(pObj), vFadds, vMap, vChains, vMap2Chain, vTruths );
    Gia_ManDupWithFaddBoxes_rec( pNew, p, Gia_ObjFanin1(pObj), vFadds, vMap, vChains, vMap2Chain, vTruths );
    pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}
func Gia_ManDupWithNaturalBoxes(p *Gia_Man_t, nFaddMin int32, fVerbose int32) *Gia_Man_t {
    var clk abctime  = Abc_Clock();
    var pNew *Gia_Man_t;//, * pTemp;
    var vFadds, vMap, vMap2Chain, vTruths, vChain *Vec_Int_t;
    var vChains *Vec_Wec_t;
    var pObj *Gia_Obj_t;  
    var i, nBoxes int32;
    if Gia_ManBoxNum(p) > 0 {
        fmt.Printf("Currently natural carry-chains cannot be detected when boxes are present.\n");
        return nil;
    }
    assert(Gia_ManBoxNum(p) == 0, "Gia_ManBoxNum(p) == 0");

    // detect FADDs
    vFadds = Gia_ManDetectFullAdders( p, fVerbose, nil );
    assert(Vec_IntSize(vFadds) % 5 == 0, "Vec_IntSize(vFadds) % 5 == 0");
    // map MAJ into its FADD
    vMap = Gia_ManCreateMap( p, vFadds );
    // for each FADD, find the longest chain and reorder its inputs
    Gia_ManFindChains( p, vFadds, vMap );
    // returns the set of topmost MAJ nodes
    vChains = Gia_ManCollectTopmost( p, vFadds, vMap, nFaddMin );
    if fVerbose != 0 {
        Gia_ManPrintChains( p, vFadds, vMap, vChains ); }
    if Vec_WecSize(vChains) == 0 {
        Vec_IntFree( vFadds );
        Vec_IntFree( vMap );
        Vec_WecFree( vChains );
        return Gia_ManDup( p );
    }
    // returns mapping of each MAJ into the topmost elements of its chain
    vMap2Chain = Gia_ManFindMapping( p, vFadds, vMap, vChains );
    // compute truth tables for FADDs
    vTruths = Gia_ManCollectTruthTables( p, vFadds );
    if fVerbose != 0 {
        Abc_PrintTime( 1, "Carry-chain detection time", Abc_Clock() - clk ); }

    // duplicate
    clk = Abc_Clock();
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vChains)) && (((vChain) = Vec_WecEntry(vChains, i)), 1); i++  {
        Gia_ManDupFadd( pNew, p, vChain, vFadds, vMap, vChains, vMap2Chain, vTruths );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupWithFaddBoxes_rec( pNew, p, Gia_ObjFanin0(pObj), vFadds, vMap, vChains, vMap2Chain, vTruths );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    if Gia_ManRegNum(p) {
        if fVerbose != 0 {
            fmt.Printf("Warning: Sequential design is coverted into combinational one by adding white boxes.\n"); }
        pNew.nRegs = 0;
    }
    assert(!Gia_ManHasDangling(pNew), "!Gia_ManHasDangling(pNew)");

    // cleanup
    Vec_IntFree( vFadds );
    Vec_IntFree( vMap );
    Vec_WecFree( vChains );
    Vec_IntFree( vMap2Chain );
    Vec_IntFree( vTruths );
    
    // other information
    nBoxes = (Gia_ManCiNum(pNew) - Gia_ManCiNum(p)) / 2;
    assert(nBoxes == (Gia_ManCoNum(pNew) - Gia_ManCoNum(p)) / 3, "nBoxes == (Gia_ManCoNum(pNew) - Gia_ManCoNum(p)) / 3");
    pNew.pManTime  = Gia_ManGenerateTim( Gia_ManCiNum(p), Gia_ManCoNum(p), nBoxes, 3, 2 );
    pNew.pAigExtra = Gia_ManGenerateExtraAig( nBoxes, 3, 2 );
/*
    // normalize
    pNew = Gia_ManDupNormalize( pTemp = pNew, 0 );
    pNew->pManTime  = pTemp->pManTime;  pTemp->pManTime  = NULL;
    pNew->pAigExtra = pTemp->pAigExtra; pTemp->pAigExtra = NULL;
    Gia_ManStop( pTemp );
*/
    //pNew = Gia_ManDupCollapse( pTemp = pNew, pNew->pAigExtra, NULL );
    //Gia_ManStop( pTemp );

    //Gia_ManIllustrateBoxes( pNew );
    if fVerbose != 0 {
        Abc_PrintTime( 1, "AIG with boxes construction time", Abc_Clock() - clk ); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Converting AIG with annotated carry-chains into AIG with boxes.]

  Description [Assumes that annotations are pObj->fMark0 or pObj->fMark1.
  Only one of these can be set to 1.  If fMark0 (fMark1) is set to 1, 
  the first (second) input of an AND-gate is chained.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjFanin0CopyCarry(vCarries *Vec_Int_t, pObj *Gia_Obj_t, Id int32) int32 {
    if vCarries == nil || Vec_IntEntry(vCarries, Gia_ObjFaninId0(pObj, Id)) == -1 {
        return Gia_ObjFanin0Copy(pObj); }
    return Abc_LitNotCond( Vec_IntEntry(vCarries, Gia_ObjFaninId0(pObj, Id)), Gia_ObjFaninC0(pObj) );
}
func Gia_ObjFanin1CopyCarry(vCarries *Vec_Int_t, pObj *Gia_Obj_t, Id int32) int32 {
    if vCarries == nil || Vec_IntEntry(vCarries, Gia_ObjFaninId1(pObj, Id)) == -1 {
        return Gia_ObjFanin1Copy(pObj); }
    return Abc_LitNotCond( Vec_IntEntry(vCarries, Gia_ObjFaninId1(pObj, Id)), Gia_ObjFaninC1(pObj) );
}
func Gia_ManDupWithArtificalFaddBoxes(p *Gia_Man_t, fUseFanout int32, fXorTrick int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;  
    var nBoxes int32  = Gia_ManBoxNum(p);
    var i, nRealPis, nRealPos int32;
    var vCarries *Vec_Int_t  = nil;
    // make sure two chains do not overlap
    Gia_ManCleanPhase( p );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        assert(!pObj.fMark0 && !pObj.fMark1, "!pObj.fMark0 && !pObj.fMark1");
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        assert(!pObj.fMark0 && !pObj.fMark1, "!pObj.fMark0 && !pObj.fMark1");
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        assert(!pObj.fMark0 || !pObj.fMark1, "!pObj.fMark0 || !pObj.fMark1");
        if pObj.fMark0 {
            assert(Gia_ObjFanin0(pObj).fPhase == 0, "Gia_ObjFanin0(pObj).fPhase == 0");
            Gia_ObjFanin0(pObj).fPhase = 1;
        }
        if pObj.fMark1 {
            assert(Gia_ObjFanin1(pObj).fPhase == 0, "Gia_ObjFanin1(pObj).fPhase == 0");
            Gia_ObjFanin1(pObj).fPhase = 1;
        }
    }
    }
    // create mapping for carry-chains
    if !(fUseFanout != 0) {
        vCarries = Vec_IntStartFull( Gia_ManObjNum(p) ); }
    // create references and discount carries
    if vCarries != nil {
        Gia_ManCreateRefs( p );
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
            if pObj.fMark0 {
                Gia_ObjRefFanin0Dec( p, pObj );
            } else if pObj.fMark1 {
                Gia_ObjRefFanin1Dec( p, pObj ); }
        }
    }
    // if AIG already has (natural) FADD boxes, it should not un-normalized
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else if !pObj.fMark0 && !pObj.fMark1 { // AND-gate
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else { // AND-gate with chain
        
            var iCiLit, iOtherLit, iLit0, iLit1, iLit2, iXorLit int32;
            assert(pObj.fMark0 != pObj.fMark1, "pObj.fMark0 != pObj.fMark1");
            if pObj.fMark0 {
                iCiLit = Gia_ObjFanin0CopyCarry(vCarries, pObj, i);
            } else {
                iCiLit = Gia_ObjFanin1CopyCarry(vCarries, pObj, i);
            }
            if pObj.fMark0 {
                iOtherLit = Gia_ObjFanin1Copy(pObj);
            } else {
                iOtherLit = Gia_ObjFanin0Copy(pObj);
            }
            assert(iCiLit >= 0 && iOtherLit >= 0, "iCiLit >= 0 && iOtherLit >= 0");
            iLit0 = Abc_LitNotCond( iCiLit,    Abc_LitIsCompl(iCiLit) );
            iLit1 = Abc_LitNotCond( iOtherLit, Abc_LitIsCompl(iCiLit) );
            iLit2 = Abc_LitNotCond( 0,         Abc_LitIsCompl(iCiLit) );
            // add COs
            assert(!Abc_LitIsCompl(iLit0), "!Abc_LitIsCompl(iLit0)");
            Gia_ManAppendCo( pNew, iLit0 );
            Gia_ManAppendCo( pNew, iLit1 );
            Gia_ManAppendCo( pNew, iLit2 );
            // add CI (unused sum bit)
            iXorLit = Gia_ManAppendCi(pNew);
            // add CI (carry bit)
            pObj.Value = Abc_LitNotCond( Gia_ManAppendCi(pNew), Abc_LitIsCompl(iCiLit) );
            if (vCarries != nil) && pObj.fPhase {
                Vec_IntWriteEntry( vCarries, i, pObj.Value );
                if Gia_ObjRefNum(p, pObj) > 0 {
                    if fXorTrick != 0 {
                        pObj.Value = Gia_ManAppendAnd( pNew, Abc_LitNotCond(iXorLit, !Abc_LitIsCompl(iCiLit)), iOtherLit );
                    } else {
                        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
                }
            }
            nBoxes++;
        }
    }
    Gia_ManCleanPhase( p );
    Vec_IntFreeP( &vCarries );
    p.pRefs = nil;
    assert(!Gia_ManHasDangling(pNew), "!Gia_ManHasDangling(pNew)");
    // other information
//    nBoxes += (Gia_ManCiNum(pNew) - Gia_ManCiNum(p)) / 2;
//    assert( nBoxes == Gia_ManBoxNum(p) + (Gia_ManCoNum(pNew) - Gia_ManCoNum(p)) / 3 );
    if Gia_ManBoxNum(p) {
        nRealPis = Tim_ManPiNum((Tim_Man_t *)p.pManTime);
    } else {
        nRealPis = Gia_ManCiNum(p);
    }
    if Gia_ManBoxNum(p) {
        nRealPos = Tim_ManPoNum((Tim_Man_t *)p.pManTime);
    } else {
        nRealPos = Gia_ManCoNum(p);
    }
    pNew.pManTime  = Gia_ManGenerateTim( nRealPis, nRealPos, nBoxes, 3, 2 );
    pNew.pAigExtra = Gia_ManGenerateExtraAig( nBoxes, 3, 2 );
    // optionally normalize the AIG
    return pNew;
}
func Gia_ManDupWithArtificalFaddBoxesTest(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // label some and-gates
    Gia_ManCleanMark01( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.fMark0 = i % 5;
        pObj.fMark1 = i % 7;
        if pObj.fMark0 && pObj.fMark1 {
            pObj.fMark0 = pObj.fMark1 = 0; }
    }
    }

    // output new AIG
    pNew = Gia_ManDupWithArtificalFaddBoxes( p, 0, 0 );
    Gia_ManCleanMark01( p );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Adds artificial carry chains to the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// computes AIG delay information when boxes are used
func Gia_ManFindAnnotatedDelay(p *Gia_Man_t, DelayC int32, pnBoxes *int32, fIgnoreBoxDelays int32) int32 {
    var pObj *Gia_Obj_t;
    var nRealPis int32;
    if Gia_ManBoxNum(p) {
        nRealPis = Tim_ManPiNum((Tim_Man_t *)p.pManTime);
    } else {
        nRealPis = Gia_ManCiNum(p);
    }
    var pDelays *int32  = Vec_IntArray(p.vLevels);
    var i, k, iBox, iBoxOutId, Delay, Delay0, Delay1, DelayMax, nBoxes int32;
    Vec_IntFill( p.vLevels, Gia_ManObjNum(p), 0 );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            if fIgnoreBoxDelays != 0 {
                continue; }
            // check if it is real PI
            iBoxOutId = Gia_ObjCioId(pObj) - nRealPis;
            if iBoxOutId < 0 {
                continue; }
            // if it is a box output, find box number
            iBox = iBoxOutId / 2;
            assert(iBox < Gia_ManBoxNum(p), "iBox < Gia_ManBoxNum(p)");
            // check find the maximum delay of the box inputs
            Delay = 0;
            for  k = 0; k < 3; k++  {
                var Id int32  = Gia_ObjId( p, Gia_ManCo(p, iBox*3+k) );
                assert(Id < i, "Id < i");
                Delay = Abc_MaxInt( Delay, pDelays[Id] );
            }
            // consider outputs
            if iBoxOutId & 1 { // carry output
                Delay += DelayC;
            } else { // sum output
                Delay += 100; }
            pDelays[i] = Delay;
            continue;
        }
        if Gia_ObjIsCo(pObj) {
            pDelays[i] = pDelays[Gia_ObjFaninId0(pObj, i)];
            DelayMax = Abc_MaxInt( DelayMax, pDelays[i] );
            continue;
        }
        assert(!pObj.fMark0 || !pObj.fMark1, "!pObj.fMark0 || !pObj.fMark1");
        Delay0 = pDelays[Gia_ObjFaninId0(pObj, i)];
        Delay1 = pDelays[Gia_ObjFaninId1(pObj, i)];
        if pObj.fMark0 {
            Delay = Abc_MaxInt( Delay0 + DelayC, Delay1 + 100 );
            nBoxes++;
        
        } else if pObj.fMark1 {
            Delay = Abc_MaxInt( Delay1 + DelayC, Delay0 + 100 );
            nBoxes++;
        
        } else {
            Delay = Abc_MaxInt( Delay0 + 100, Delay1 + 100 ); }
        pDelays[i] = Delay;
    }
    if pnBoxes != nil {
        *pnBoxes = nBoxes; }
    return DelayMax;
}
// check if the object is already used in some chain
func Gia_ObjIsUsed(pObj *Gia_Obj_t) int32 {
    return pObj.fMark0 || pObj.fMark1 || pObj.fPhase;
}
// finds internal node that can begin a new chain
func Gia_ManFindChainStart(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var pDelays *int32  = Vec_IntArray(p.vLevels);
    var i int32;
    var iMax int32 = -1;
    var DelayMax int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjIsUsed(pObj) != 0 {
            continue; }
        if DelayMax > pDelays[i] {
            continue; }
        DelayMax = pDelays[i];
        iMax = i;
    }
    }
    return iMax;
}
// finds a sequence of internal nodes that creates a new chain
func Gia_ManFindPath(p *Gia_Man_t, DelayC int32, nPathMin int32, nPathMax int32, vPath *Vec_Int_t) int32 {
    var pObj, pFanin0, pFanin1 *Gia_Obj_t;
    var pDelays *int32  = Vec_IntArray(p.vLevels);
    var i int32;
    var iLit int32;
    var iMax int32 = Gia_ManFindChainStart( p );
    if iMax == -1 {
        return -1; }
    Vec_IntClear( vPath );
    pObj = Gia_ManObj(p, iMax);
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    for  Gia_ObjIsAnd(pObj)  {
        assert(!(Gia_ObjIsUsed(pObj) != 0), "!Gia_ObjIsUsed(pObj)");
        pFanin0 = Gia_ObjFanin0(pObj);
        pFanin1 = Gia_ObjFanin1(pObj);
        if (Gia_ObjIsUsed(pFanin0) != 0) && (Gia_ObjIsUsed(pFanin1) != 0) {
            break; }
        if Gia_ObjIsUsed(pFanin0) != 0 {
            Vec_IntPush( vPath, Abc_Var2Lit(Gia_ObjId(p, pObj), 1) );
            pObj = pFanin1;
        
        } else if Gia_ObjIsUsed(pFanin1) != 0 {
            Vec_IntPush( vPath, Abc_Var2Lit(Gia_ObjId(p, pObj), 0) );
            pObj = pFanin0;
        
        } else {
            if pDelays[Gia_ObjId(p, pFanin1)] > pDelays[Gia_ObjId(p, pFanin0)] {
                Vec_IntPush( vPath, Abc_Var2Lit(Gia_ObjId(p, pObj), 1) );
                pObj = pFanin1;
            
            } else {
                Vec_IntPush( vPath, Abc_Var2Lit(Gia_ObjId(p, pObj), 0) );
                pObj = pFanin0;
            }
        }
    }
    if Vec_IntSize(vPath) < nPathMin {
        Gia_ManObj(p, iMax).fPhase = 1;
        return 0;
    }
    // label nodes
    if Vec_IntSize(vPath) > nPathMax {
        Vec_IntShrink( vPath, nPathMax ); }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPath)) && (((iLit) = Vec_IntEntry(vPath, i)), 1); i++  {
        pObj = Gia_ManObj( p, Abc_Lit2Var(iLit) );
        if Abc_LitIsCompl(iLit) {
            assert(pObj.fMark1 == 0, "pObj.fMark1 == 0");
            pObj.fMark1 = 1;
            assert(Gia_ObjFanin1(pObj).fPhase == 0, "Gia_ObjFanin1(pObj).fPhase == 0");
            Gia_ObjFanin1(pObj).fPhase = 1;
        
        } else {
            assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
            pObj.fMark0 = 1;
            assert(Gia_ObjFanin0(pObj).fPhase == 0, "Gia_ObjFanin0(pObj).fPhase == 0");
            Gia_ObjFanin0(pObj).fPhase = 1;
        }
    }
    return Vec_IntSize(vPath);
}
// iteratively create the given number of chains
func Gia_ManIteratePaths(p *Gia_Man_t, DelayC int32, nPathMin int32, nPathMax int32, nPathLimit int32, fIgnoreBoxDelays int32, fVerbose int32) int32 {
    var pObj *Gia_Obj_t;
    var vPath *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i, RetValue, nBoxes, MaxDelay, nPaths int32;
    assert(p.vLevels == nil, "p.vLevels == nil");
    p.vLevels = Vec_IntStart( Gia_ManObjNum(p) );
    Gia_ManCleanMark01( p );
    Gia_ManCleanPhase( p );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.fPhase = 1;
    }
    if fVerbose != 0 {
        fmt.Printf("Running path detection: BoxDelay = %d, PathMin = %d, PathMax = %d, PathLimit = %d.\n", DelayC, nPathMin, nPathMax, nPathLimit); }
    for  i = 0; i < nPathLimit; i++  {
        MaxDelay = Gia_ManFindAnnotatedDelay( p, DelayC, &nBoxes, fIgnoreBoxDelays );
        RetValue = Gia_ManFindPath( p, DelayC, nPathMin, nPathMax, vPath );
        if RetValue == -1 {
            break; }
        nPaths += (RetValue > 0);
        if fVerbose != 0 {
            fmt.Printf("Iter %5d : Paths = %2d. Boxes = %2d. Total boxes = %6d.  Max delay = %5d.\n", i, nPaths, RetValue, nBoxes, MaxDelay); }
    }
    Vec_IntFree( vPath );
    Vec_IntFreeP( &p.vLevels );
    Gia_ManCleanPhase( p );
    return 1;
}
// annotate artificial chains and then put them into boxes
func Gia_ManDupWithArtificialBoxes(p *Gia_Man_t, DelayC int32, nPathMin int32, nPathMax int32, nPathLimit int32, fUseFanout int32, fXorTrick int32, fIgnoreBoxDelays int32, fVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
/*
    if ( Gia_ManBoxNum(p) > 0 )
    {
        printf( "Currently artifical carry-chains cannot be detected when natural ones are present.\n" );
        return NULL;
    }
*/
    Gia_ManIteratePaths( p, DelayC, nPathMin, nPathMax, nPathLimit, fIgnoreBoxDelays, fVerbose );
    pNew = Gia_ManDupWithArtificalFaddBoxes( p, fUseFanout, fXorTrick );
    Gia_ManCleanMark01( p );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


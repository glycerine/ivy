package wlc

import (
	"fmt"
	"os"
	"unsafe"
)

/**CFile****************************************************************

  FileName    [wlcGraft.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Verilog parser.]

  Synopsis    [Parses several flavors of word-level Verilog.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 22, 2014.]

  Revision    [$Id: wlcGraft.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wlc.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Internal simulation APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_ObjSim(p *Gia_Man_t, iObj int32) *word {
    return Vec_WrdEntryP( p.vSims, p.nSimWords * iObj );
}
func Wlc_ObjSimPi(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSim *word  = Wlc_ObjSim( p, iObj );
    for  w = 0; w < p.nSimWords; w++  {
        pSim[w] = Gia_ManRandomW( 0 );
    }
    pSim[0] <<= 1;
}
func Wlc_ObjSimRo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pSimRo *word  = Wlc_ObjSim( p, iObj );
    var pSimRi *word  = Wlc_ObjSim( p, Gia_ObjRoToRiId(p, iObj) );
    for  w = 0; w < p.nSimWords; w++  {
        pSimRo[w] = pSimRi[w];
    }
}
func Wlc_ObjSimCo(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSimCo *word   = Wlc_ObjSim( p, iObj );
    var pSimDri *word  = Wlc_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] = ~pSimDri[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSimCo[w] =  pSimDri[w];
        } }
}
func Wlc_ObjSimAnd(p *Gia_Man_t, iObj int32) {
    var w int32;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var pSim *word   = Wlc_ObjSim( p, iObj );
    var pSim0 *word  = Wlc_ObjSim( p, Gia_ObjFaninId0(pObj, iObj) );
    var pSim1 *word  = Wlc_ObjSim( p, Gia_ObjFaninId1(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = ~pSim0[w] & ~pSim1[w];
        }
    } else if Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = ~pSim0[w] & pSim1[w];
        }
    } else if !Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = pSim0[w] & ~pSim1[w];
        }
    } else {
        for  w = 0; w < p.nSimWords; w++  {
            pSim[w] = pSim0[w] & pSim1[w];
        } }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_NtkCollectObjs_rec(p *Wlc_Ntk_t, pObj *Wlc_Obj_t, vObjs *Vec_Int_t) int32 {
    var i, iFanin, Count int32;
    if Wlc_ObjIsCi(pObj) {
        return 0; }
    if pObj.Mark {
        return 0; }
    pObj.Mark = 1;
    // inlining of macro Wlc_ObjForEachFanin
    for i = 0; i < Wlc_ObjFaninNum(pObj); i++ {
        iFanin = Wlc_ObjFaninId(pObj, i)
        Count += Wlc_NtkCollectObjs_rec( p, Wlc_NtkObj(p, iFanin), vObjs );
    }
    Vec_IntPush( vObjs, Wlc_ObjId(p, pObj) );
    if pObj.Type == WLC_OBJ_ARI_MULTI { Count++ }
    return Count;
}
func Wlc_NtkCollectObjs(p *Wlc_Ntk_t, fEven int32, pCount *int32) *Vec_Int_t {
    var vObjs *Vec_Int_t  = Vec_IntAlloc( 100 );
    var pObj *Wlc_Obj_t; 
    var i, Count int32;
    Wlc_NtkCleanMarks( p );
    // inlining of macro Wlc_NtkForEachCo
    for i = 0; i < Wlc_NtkCoNum(p); i++ {
        pObj = Wlc_NtkCo(p, i)
        if (i & 1) == fEven {
            Count += Wlc_NtkCollectObjs_rec( p, pObj, vObjs ); }
    }
    Wlc_NtkCleanMarks( p );
    if pCount != nil {
        *pCount = Count; }
    return vObjs;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_NtkSaveOneNode(p *Wlc_Ntk_t, pWlcObj *Wlc_Obj_t, pGia *Gia_Man_t, vTtMem *Vec_Mem_t) {
    var k, Entry int32;
    var nWords int32  = Vec_MemEntrySize(vTtMem);
    var nBits int32  = Wlc_ObjRange(pWlcObj);
    var iFirst int32  = Vec_IntEntry(p.vCopies, Wlc_ObjId(p, pWlcObj) );
    for  k = 0; k < nBits; k++  {
        var iLit int32  = Vec_IntEntry(p.vBits, iFirst + k );
        var pInfoObj *word  = Wlc_ObjSim( pGia, Abc_Lit2Var(iLit) );
        var fCompl int32  = int32(pInfoObj[0] & 1);
        if fCompl != 0 { Abc_TtNot( pInfoObj, nWords ); }
        Entry = Vec_MemHashInsert( vTtMem, pInfoObj );
        if fCompl != 0 { Abc_TtNot( pInfoObj, nWords ); }
        fmt.Printf("%2d(%d) ", Entry, fCompl ^ Abc_LitIsCompl(iLit));
        Extra_PrintHex( os.Stdout, (*uint32)(unsafe.Pointer(pInfoObj)), 8 );
        fmt.Printf("\n");
    }
    fmt.Printf("\n");
}
func Wlc_NtkFindOneNode(p *Wlc_Ntk_t, pWlcObj *Wlc_Obj_t, pGia *Gia_Man_t, vTtMem *Vec_Mem_t) {
    var k, Entry int32;
    var nWords int32  = Vec_MemEntrySize(vTtMem);
    var nBits int32  = Wlc_ObjRange(pWlcObj);
    var iFirst int32  = Vec_IntEntry(p.vCopies, Wlc_ObjId(p, pWlcObj) );
    for  k = 0; k < nBits; k++  {
        var iLit int32  = Vec_IntEntry(p.vBits, iFirst + k );
        var pInfoObj *word  = Wlc_ObjSim( pGia, Abc_Lit2Var(iLit) );
        var fCompl int32  = int32(pInfoObj[0] & 1);
        if fCompl != 0 { Abc_TtNot( pInfoObj, nWords ); }
        Entry = *Vec_MemHashLookup( vTtMem, pInfoObj );
        if Entry > 0 {
            fmt.Printf("Obj %4d.  Range = %2d.  Bit %2d.  Entry %d(%d).  %s\n", Wlc_ObjId(p, pWlcObj), Wlc_ObjRange(pWlcObj), k, Entry, fCompl ^ Abc_LitIsCompl(iLit), Wlc_ObjName(p, Wlc_ObjId(p, pWlcObj))); }
        if fCompl != 0 { Abc_TtNot( pInfoObj, nWords ); }
        //printf( "%2d ", Entry );
        //Extra_PrintHex( stdout, (*uint32)(unsafe.Pointer(pInfoObj)), 8 );
        //printf( "\n" );
    }
    //printf( "\n" );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wlc_NtkGraftMulti(p *Wlc_Ntk_t, fVerbose int32) *Wlc_Ntk_t {
    var nWords int32  = 4;
    var i, nMultiLHS, nMultiRHS int32;
    var pInfoObj *word;
    var pNew *Wlc_Ntk_t  = nil;
    var pWlcObj *Wlc_Obj_t;
    var pObj *Gia_Obj_t;
    var vObjsLHS *Vec_Int_t  = Wlc_NtkCollectObjs( p, 0, &nMultiLHS );
    var vObjsRHS *Vec_Int_t  = Wlc_NtkCollectObjs( p, 1, &nMultiRHS );
    var pGia *Gia_Man_t  = Wlc_NtkBitBlast( p, nil ); //, -1, 0, 0, 0, 0, 1, 0, 0 ); // <= no cleanup
    var vTtMem *Vec_Mem_t  = Vec_MemAlloc( nWords, 10 );
    Vec_MemHashAlloc( vTtMem, 10000 );

    // check if there are multipliers
    if nMultiLHS == 0 && nMultiRHS == 0 {
        fmt.Printf("No multipliers are present.\n");
        return nil;
    }
    // compare multipliers
    if nMultiLHS > 0 && nMultiRHS > 0 {
        fmt.Printf("Multipliers are present in both sides of the miter.\n");
        return nil;
    }
    // swap if wrong side
    if nMultiRHS > 0 {
        vObjsLHS, vObjsRHS = vObjsRHS, vObjsLHS
        nMultiLHS, nMultiRHS = nMultiRHS, nMultiLHS
    }
    assert(nMultiLHS > 0, "nMultiLHS > 0");
    assert(nMultiRHS == 0, "nMultiRHS == 0");

    // allocate simulation info for one timeframe
    Vec_WrdFreeP(&pGia.vSims);
    pGia.vSims = Vec_WrdStart( Gia_ManObjNum(pGia) * nWords );
    pGia.nSimWords = nWords;
    // perform simulation
    Gia_ManRandomW( 1 );
    for i = 1; i < Gia_ManObjNum(pGia); i++ {
        pObj = Gia_ManObj(pGia, i)
        if Gia_ObjIsAnd(pObj) {
            Wlc_ObjSimAnd( pGia, i );
        } else if Gia_ObjIsCo(pObj) {
            Wlc_ObjSimCo( pGia, i );
        } else if Gia_ObjIsCi(pObj) {
            Wlc_ObjSimPi( pGia, i );
        } else { assert(false, "0"); }
    }

    // hash constant 0
    pInfoObj = Wlc_ObjSim( pGia, 0 );
    Vec_MemHashInsert( vTtMem, pInfoObj );

    // hash sim info on the multiplier boundary
    // inlining of macro Wlc_NtkForEachObjVec
    for i = 0; i < Vec_IntSize(vObjsLHS); i++ {
        pWlcObj = Wlc_NtkObj(p, Vec_IntEntry(vObjsLHS, i))
        if Wlc_ObjType(pWlcObj) == WLC_OBJ_ARI_MULTI {
            Wlc_NtkSaveOneNode( p, Wlc_ObjFanin0(p, pWlcObj), pGia, vTtMem );
            Wlc_NtkSaveOneNode( p, Wlc_ObjFanin1(p, pWlcObj), pGia, vTtMem );
            Wlc_NtkSaveOneNode( p, pWlcObj, pGia, vTtMem );
        }
    }

    // check if there are similar signals in LHS
    // inlining of macro Wlc_NtkForEachObjVec
    for i = 0; i < Vec_IntSize(vObjsRHS); i++ {
        pWlcObj = Wlc_NtkObj(p, Vec_IntEntry(vObjsRHS, i))
        Wlc_NtkFindOneNode( p, pWlcObj, pGia, vTtMem );
    }

    // perform grafting


    Vec_MemHashFree( vTtMem );
    Vec_MemFreeP(&vTtMem);

    // cleanup
    Vec_WrdFreeP(&pGia.vSims);
    pGia.nSimWords = 0;

    Vec_IntFree( vObjsLHS );
    Vec_IntFree( vObjsRHS );
    Gia_ManStop( pGia );
    return pNew;
}





/**Function*************************************************************

  Synopsis    [Generate simulation vectors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbc_Mult(a word, b word, r []word) {
    var Msk word  = 0xFFFFFFFF;
    var pL word   = (a & Msk) * (b & Msk);
    var pM1 word  = (a & Msk) * (b >> 32);
    var pM2 word  = (a >> 32) * (b & Msk);
    var pH word   = (a >> 32) * (b >> 32);
    var Car word  = (pM1 & Msk) + (pM2 & Msk) + (pL >> 32);
    r[0] = a * b;
    r[1] = pH + (pM1 >> 32) + (pM2 >> 32) + (Car >> 32);
}
func Sbc_SimMult(A []word, B []word, R []word, nIns int32) {
    var a word;
    var b word;
    var r [2]word;
    var Mask word = Abc_Tt6Mask(nIns); var i, k int32;
    for  i = 0; i < 64; i++  {
        A[i], B[i], R[i], R[i+64] = 0, 0, 0, 0;
    }
    Gia_ManRandom(1);
    for  i = 0; i < 64; i++  {
        if i != 0 {
            a = (Mask & Gia_ManRandomW(0));
        } else {
            a = 0;
        }
        if i != 0 {
            b = (Mask & Gia_ManRandomW(0));
        } else {
            b = 0;
        }        
        Sbc_Mult( a, b, r );
        for  k = 0; k < 64; k++  {
            if ((a >> k) & 1) != 0 {  A[k]    |= (word(1) << i); }
            if ((b >> k) & 1) != 0 {  B[k]    |= (word(1) << i); }
            if ((r[0] >> k) & 1) != 0 {  R[k]    |= (word(1) << i); }
            if ((r[1] >> k) & 1) != 0 {  R[k+64] |= (word(1) << i); }
        }
    }
//    for ( i = 0; i < 128; i++ )
//        for ( k = 0; k < 64; k++, printf( "\n" ) )
//            printf( "%d", (R[i] >> k) & 1 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbc_ManDetectMult(p *Gia_Man_t, vIns *Vec_Int_t) *Vec_Int_t {
    var nWords int32  = 1;
    var vGia2Out *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    var pObj *Gia_Obj_t; var i int32;
var Entry int32;
var nIns int32 = Vec_IntSize(vIns)/2;
    var A [64]word;
    var B [64]word;
    var R [128]word;
    var pInfoObj *word; var Temp word;

    // create hash table
    var vTtMem *Vec_Mem_t  = Vec_MemAlloc( nWords, 10 );
    Vec_MemHashAlloc( vTtMem, 1000 );
    Sbc_SimMult(A[:], B[:], R[:], nIns);
    for  i = 0; i < 2*nIns; i++  {
        Vec_MemHashInsert(vTtMem, &R[i]);
        //printf( "Out %2d : ", i );
        //Extra_PrintHex( stdout, (unsigned *)(R+i), 6 ); printf( "\n" );
    }
    assert(Vec_MemEntryNum(vTtMem) == 2*nIns, "Vec_MemEntryNum(vTtMem) == 2*nIns");

    // alloc simulation info
    Vec_WrdFreeP(&p.vSims);
    p.vSims = Vec_WrdStart( Gia_ManObjNum(p) * nWords );
    p.nSimWords = nWords;

    // mark inputs
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrentId( p, 0 );
    for i = 0; i < Gia_ManCiNum(p); i++ {
        pObj = Gia_ManCi(p, i)
        Gia_ObjSetTravIdCurrent( p, pObj );
        //Wlc_ObjSimPi( p, Gia_ObjId(p, pObj) );
    }

    // assign inputs
    assert(Vec_IntSize(vIns) % 2 == 0, "Vec_IntSize(vIns) % 2 == 0");
    for i = 0; i < Vec_IntSize(vIns); i++ {
        pObj = Gia_ManObj(p, Vec_IntEntry(vIns, i))
        Gia_ObjSetTravIdCurrent( p, pObj );
        pInfoObj = Wlc_ObjSim( p, Gia_ObjId(p, pObj) );
        if i < nIns {
            *pInfoObj = A[i]
        } else {
            *pInfoObj = B[i-nIns]
        }
    }

    // perform simulation
    for i = 1; i < Gia_ManObjNum(p); i++ {
        pObj = Gia_ManObj(p, i)
        if Gia_ObjIsTravIdCurrent(p, pObj) {
            continue; }

        if Gia_ObjIsAnd(pObj) {
            Wlc_ObjSimAnd( p, i );
        } else if Gia_ObjIsCo(pObj) {
            Wlc_ObjSimCo( p, i );
        } else { assert(false, "0"); }

        // mark direct polarity
        pInfoObj = Wlc_ObjSim( p, i );
        Entry = *Vec_MemHashLookup( vTtMem, pInfoObj );
        if Entry >= 0 {
            Vec_IntWriteEntry( vGia2Out, i, Abc_Var2Lit(Entry, 0) );
            continue;
        }

        // mark negated polarity
        Temp = *pInfoObj;
        Abc_TtNot( pInfoObj, nWords );
        Entry = *Vec_MemHashLookup( vTtMem, pInfoObj );
        Abc_TtNot( pInfoObj, nWords );
        assert(Temp == *pInfoObj, "Temp == *pInfoObj");
        if Entry >= 0 {
            Vec_IntWriteEntry( vGia2Out, i, Abc_Var2Lit(Entry, 1) );
            continue;
        }
    }

    for i = 0; i < Gia_ManCoNum(p); i++ {
        pObj = Gia_ManCo(p, i)
        pInfoObj = Wlc_ObjSim( p, Gia_ObjId(p, pObj) );
        //printf( "Out %2d : Driver = %5d(%d)", i, Gia_ObjFaninId0p(p, pObj), Gia_ObjFaninC0(pObj) );
        //Extra_PrintHex( stdout, (*uint32)(unsafe.Pointer(pInfoObj)), 6 ); printf( "\n" );
    }

    // cleanup
    Vec_MemHashFree( vTtMem );
    Vec_MemFreeP(&vTtMem);
    //Vec_WrdFreeP( &p->vSims );
    //p->nSimWords = 0;
    return vGia2Out;
}
func Sbc_ManWlcNodes2(pNtk *Wlc_Ntk_t, p *Gia_Man_t, vGiaLits *Vec_Int_t) *Vec_Int_t {
    var pObj *Wlc_Obj_t;  var i, k, iGiaLit, iFirst, nBits int32;
    var vRes *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vMap *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    for i = 0; i < Vec_IntSize(vGiaLits); i++ {
        iGiaLit = Vec_IntEntry(vGiaLits, i)
        if iGiaLit != -1 {
            Vec_IntWriteEntry(vMap, Abc_Lit2Var(iGiaLit), Abc_Var2Lit(i, Abc_LitIsCompl(iGiaLit)))
        }
    }
    // inlining of macro Wlc_NtkForEachObj
    for i = 1; i < Wlc_NtkObjNumMax(pNtk); i++ {
        pObj = Wlc_NtkObj(pNtk, i)
        iFirst = Vec_IntEntry(pNtk.vCopies, i );
        nBits  = Wlc_ObjRange(pObj);
        for  k = 0; k < nBits; k++  {
            var iLitGia int32  = Vec_IntEntry(pNtk.vBits, iFirst + k );
            var iLitOut int32  = Vec_IntEntry( vMap, Abc_Lit2Var(iLitGia) );
            if iLitOut == -1 {
                continue; }
            Vec_IntWriteEntry( vMap, Abc_Lit2Var(iLitGia), -1 );
            iLitOut = Abc_LitNotCond( iLitOut, Abc_LitIsCompl(iLitGia) );
            fmt.Printf("Matched out %d in phase %d with object %d (%s) bit %d (out of %d).\n", Abc_Lit2Var(iLitOut), Abc_LitIsCompl(iLitOut), i, Wlc_ObjName(pNtk, Wlc_ObjId(pNtk, pObj)), k, nBits);
            Vec_IntPushUnique( vRes, i );
        }
    }
    Vec_IntFree( vMap );
    Vec_IntSort( vRes, 0 );
    // consider the last one
    pObj   = Wlc_NtkObj( pNtk, Vec_IntEntryLast(vRes) );
    iFirst = Vec_IntEntry(pNtk.vCopies, Wlc_ObjId(pNtk, pObj) );
    nBits  = Wlc_ObjRange(pObj);
    fmt.Printf("Considering object %d (%s):\n", Wlc_ObjId(pNtk, pObj), Wlc_ObjName(pNtk, Wlc_ObjId(pNtk, pObj)));
    for  k = 0; k < nBits; k++  {
        var iLitGia int32   = Vec_IntEntry(pNtk.vBits, iFirst + k );
        var iLitOutP int32  = Vec_IntFind( vGiaLits, iLitGia );
        var iLitOutN int32  = Vec_IntFind( vGiaLits, Abc_LitNot(iLitGia) );
        fmt.Printf("Matching bit %d with output %d / %d.\n", k, iLitOutP, iLitOutN);
        // print simulation signature
        {
            var pInfoObj *word  = Wlc_ObjSim( p, Abc_Lit2Var(iLitGia) );
            Extra_PrintHex( os.Stdout, (*uint32)(unsafe.Pointer(pInfoObj)), 6 ); fmt.Printf("\n");
        }
    }
    return vRes;
}
func Sbc_ManWlcNodes(pNtk *Wlc_Ntk_t, p *Gia_Man_t, vGia2Out *Vec_Int_t, nOuts int32) int32 {
    var pObj *Wlc_Obj_t;  
    var i int32;
    var k int32;
    var iLitGia int32;
    var iLitOut int32;
    var iFirst int32;
    var nBits int32;
    var iObjFound int32 = -1;
    var vMatched *Vec_Int_t  = Vec_IntAlloc( 100 );
    // inlining of macro Wlc_NtkForEachObj
    for i = 1; i < Wlc_NtkObjNumMax(pNtk); i++ {
        pObj = Wlc_NtkObj(pNtk, i)
        iFirst = Vec_IntEntry(pNtk.vCopies, i );
        nBits  = Wlc_ObjRange(pObj);
        Vec_IntClear( vMatched );
        for  k = 0; k < nBits; k++  {
            iLitGia = Vec_IntEntry(pNtk.vBits, iFirst + k );
            iLitOut = Vec_IntEntry( vGia2Out, Abc_Lit2Var(iLitGia) );
            if iLitOut == -1 {
                continue; }
            iLitOut = Abc_LitNotCond( iLitOut, Abc_LitIsCompl(iLitGia) );
            fmt.Printf("Matched node %5d (%10s) bit %3d (out of %3d) with output %3d(%d).\n", i, Wlc_ObjName(pNtk, Wlc_ObjId(pNtk, pObj)), k, nBits, Abc_Lit2Var(iLitOut), Abc_LitIsCompl(iLitOut));
            Vec_IntPushOrder( vMatched, Abc_Lit2Var(iLitOut) );
        }
        if Vec_IntSize(vMatched) > 0 {
            fmt.Printf("\n"); }
        if Vec_IntSize(vMatched) == nOuts {
            if iObjFound == -1 {
                iObjFound = i; }
            fmt.Printf("Found object %d with all bits matched.\n", i);
            /*
            for ( k = nBits-2; k < nBits; k++ )
            {
                iLitGia = Vec_IntEntry( &pNtk->vBits, iFirst + k );
                {
                    word * pInfoObj = Wlc_ObjSim( p, Abc_Lit2Var(iLitGia) );
                    Extra_PrintHex( stdout, (*uint32)(unsafe.Pointer(pInfoObj)), 6 ); printf( "\n" );
                }
            }
            */
            break;
        }
    }
    Vec_IntFree( vMatched );
    return iObjFound;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbc_ManDetectMultTest(pNtk *Wlc_Ntk_t, fVerbose int32) {
    var p *Gia_Man_t  = Wlc_NtkBitBlast( pNtk, nil );//, -1, 0, 0, 0, 0, 1, 0, 0 ); // <= no cleanup
    var vIns, vGia2Out *Vec_Int_t;
    var iObjFound int32  = -1;
//    Gia_Obj_t * pObj; int i;
//    Gia_ManForEachCo( p, pObj, i )
//        printf( "Output %2d - driver %5d (%d)\n", i, Gia_ObjFaninId0p(p, pObj), Gia_ObjFaninC0(pObj) );

    vIns = Sdb_StoComputeCutsDetect( p );
    if vIns == nil || Vec_IntSize(vIns) == 0 || (Vec_IntSize(vIns) % 2) != 0 {
        fmt.Printf("Input identification did not work out.\n");
        return;
    }

    vGia2Out = Sbc_ManDetectMult( p, vIns );

    iObjFound = Sbc_ManWlcNodes( pNtk, p, vGia2Out, Vec_IntSize(vIns) );
    _ = iObjFound

    Vec_IntFree( vGia2Out );
    Vec_IntFree( vIns );

    Gia_ManStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


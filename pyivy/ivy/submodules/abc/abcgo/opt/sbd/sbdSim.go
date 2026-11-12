package sbd

import "fmt"

/**CFile****************************************************************

  FileName    [sbdSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sbdSim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sbdInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Sbd_ObjSims(p *Gia_Man_t, i int32) *word { return Vec_WrdEntryP( p.vSims,   p.iPatsPi * i );  }
func Sbd_ObjCtrl(p *Gia_Man_t, i int32) *word { return Vec_WrdEntryP( p.vSimsPi, p.iPatsPi * i );  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This does not work.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_GiaSimRoundBack2(p *Gia_Man_t) {
    var nWords int32  = p.iPatsPi;
    var pObj *Gia_Obj_t;
    var w, i, Id int32;
    // init primary outputs
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
        for  w = 0; w < nWords; w++  {
            Sbd_ObjSims(p, Id)[w] = Gia_ManRandomW(0);
        }
    }
    // transfer to nodes
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        var pSims *word  = Sbd_ObjSims(p, Gia_ObjId(p, pObj));
        Abc_TtCopy( Sbd_ObjSims(p, Gia_ObjFaninId0p(p, pObj)), pSims, nWords, Gia_ObjFaninC0(pObj) );
    }
    // simulate nodes
    // inlining of macro Gia_ManForEachAndReverse
    for  i = p.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
        var pSims *word   = Sbd_ObjSims(p, i);
        var pSims0 *word  = Sbd_ObjSims(p, Gia_ObjFaninId0(pObj, i));
        var pSims1 *word  = Sbd_ObjSims(p, Gia_ObjFaninId1(pObj, i));
        var Rand word  = Gia_ManRandomW(0);
        for  w = 0; w < nWords; w++  {
            pSims0[w] = pSims[w] | Rand;
            pSims1[w] = pSims[w] | ~Rand;
        }
        if Gia_ObjFaninC0(pObj) { Abc_TtNot( pSims0, nWords ); }
        if Gia_ObjFaninC1(pObj) { Abc_TtNot( pSims1, nWords ); }
    }
    }
    // primary inputs are initialized
}


/**Function*************************************************************

  Synopsis    [Tries to falsify a sequence of two-literal SAT problems.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_GiaSatOne_rec(p *Gia_Man_t, pObj *Gia_Obj_t, Value int32, fFirst int32, iPat int32) int32 {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return int32(pObj.fMark0) == Value; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    pObj.fMark0 = Value;
    if Gia_ObjIsCi(pObj) {
        var pSims *word  = Sbd_ObjSims(p, Gia_ObjId(p, pObj));
        if Abc_TtGetBit( pSims, iPat ) != Value {
            Abc_TtXorBit( pSims, iPat ); }
        return 1;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    assert(!Gia_ObjIsXor(pObj), "!Gia_ObjIsXor(pObj)");
    if Value != 0 {
        return Sbd_GiaSatOne_rec( p, Gia_ObjFanin0(pObj), !Gia_ObjFaninC0(pObj), fFirst, iPat ) &&
               Sbd_GiaSatOne_rec( p, Gia_ObjFanin1(pObj), !Gia_ObjFaninC1(pObj), fFirst, iPat ); }
    if fFirst != 0 {
        return Sbd_GiaSatOne_rec( p, Gia_ObjFanin0(pObj), Gia_ObjFaninC0(pObj), fFirst, iPat );
    } else {
        return Sbd_GiaSatOne_rec( p, Gia_ObjFanin1(pObj), Gia_ObjFaninC1(pObj), fFirst, iPat ); }
}
func Sbd_GiaSatOne(p *Gia_Man_t, vPairs *Vec_Int_t) int32 {
    var k, n, Var1, Var2, iPat int32;
    //Gia_ManSetPhase( p );
    Vec_IntForEachEntryDouble( vPairs, Var1, Var2, k )
    {
        var pObj1 *Gia_Obj_t  = Gia_ManObj( p, Var1 );
        var pObj2 *Gia_Obj_t  = Gia_ManObj( p, Var2 );
        assert(Var2 > 0, "Var2 > 0");
        if Var1 == 0 {
            for  n = 0; n < 2; n++  {
                Gia_ManIncrementTravId( p );
                if Sbd_GiaSatOne_rec(p, pObj2, !pObj2.fPhase, n, iPat) != 0 {
                    iPat++;
                    break;
                }
            }
            fmt.Printf("%c", n == 2 ? '.' : 'c');
        
        } else {
            for  n = 0; n < 2; n++  {
                Gia_ManIncrementTravId( p );
                if (Sbd_GiaSatOne_rec(p, pObj1, !pObj1.fPhase, n, iPat) != 0) && (Sbd_GiaSatOne_rec(p, pObj2, pObj2.fPhase, n, iPat) != 0) {
                    iPat++;
                    break;
                }
                Gia_ManIncrementTravId( p );
                if (Sbd_GiaSatOne_rec(p, pObj1, pObj1.fPhase, n, iPat) != 0) && (Sbd_GiaSatOne_rec(p, pObj2, !pObj2.fPhase, n, iPat) != 0) {
                    iPat++;
                    break;
                }
            }
            fmt.Printf("%c", n == 2 ? '.' : 'e');
        }
        if iPat == 64 * p.iPatsPi - 1 {
            break; }
    }
    fmt.Printf("\n");
    return iPat;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_GiaSimRoundBack(p *Gia_Man_t) {
    extern void Sbd_GiaSimRound( Gia_Man_t * p, int fTry, Vec_Int_t ** pvMap );
    var vReprs *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var vPairs *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vMap *Vec_Int_t; // maps each node into its class
    var i, nConsts, nClasses, nPats int32;
    Sbd_GiaSimRound( p, 0, &vMap );
    // inlining of macro Gia_ManForEachAndId
    for  i = 0; (i < p.nObjs); i++  {                                     if !Gia_ObjIsAnd(Gia_ManObj(p, i)) {} else {
        if Vec_IntEntry(vMap, i) == 0 {
            Vec_IntPushTwo( vPairs, 0, i );
nConsts++;
        } else if Vec_IntEntry(vReprs, Vec_IntEntry(vMap, i)) == 0 {
            Vec_IntWriteEntry( vReprs, Vec_IntEntry(vMap, i), i );
        } else if Vec_IntEntry(vReprs, Vec_IntEntry(vMap, i)) != -1 {
            Vec_IntPushTwo( vPairs, Vec_IntEntry(vReprs, Vec_IntEntry(vMap, i)), i );
            Vec_IntWriteEntry( vReprs, Vec_IntEntry(vMap, i), -1 );
            nClasses++;
        }
    }
    }
    Vec_IntFree( vMap );
    Vec_IntFree( vReprs );
    fmt.Printf("Constants = %d.   Classes = %d.\n", nConsts, nClasses);

    nPats = Sbd_GiaSatOne( p, vPairs );
    Vec_IntFree( vPairs );

    fmt.Printf("Generated %d patterns.\n", nPats);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbd_GiaSimRound(p *Gia_Man_t, fTry int32, pvMap **Vec_Int_t) {
    var nWords int32  = p.iPatsPi;
    var vStore *Vec_Mem_t;
    var pObj *Gia_Obj_t;
    var vMap *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var w, i, Id, fCompl, RetValue int32;
    // init primary inputs
    if fTry != 0 {
        Sbd_GiaSimRoundBack( p );
        // inlining of macro Gia_ManForEachCiId
        for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
            Sbd_ObjSims(p, Id)[0] <<= 1;
        }
    
    } else {
        // inlining of macro Gia_ManForEachCiId
        for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
            for  w = 0; w < nWords; w++  {
                Sbd_ObjSims(p, Id)[w] = Gia_ManRandomW(0) << !w;
            }
        }
    }
    // simulate internal nodes
    vStore = Vec_MemAlloc( nWords, 16 ); // 2^12 N-word entries per page
    Vec_MemHashAlloc( vStore, 1 << 16 );
    RetValue = Vec_MemHashInsert( vStore, Sbd_ObjSims(p, 0) ); // const zero
    assert(RetValue == 0, "RetValue == 0");
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var pSims *word  = Sbd_ObjSims(p, i);
        if Gia_ObjIsXor(pObj) {
            Abc_TtXor( pSims, 
                Sbd_ObjSims(p, Gia_ObjFaninId0(pObj, i)), 
                Sbd_ObjSims(p, Gia_ObjFaninId1(pObj, i)), 
                nWords, 
                Gia_ObjFaninC0(pObj) ^ Gia_ObjFaninC1(pObj) );
        } else {
            Abc_TtAndCompl( pSims, 
                Sbd_ObjSims(p, Gia_ObjFaninId0(pObj, i)), Gia_ObjFaninC0(pObj), 
                Sbd_ObjSims(p, Gia_ObjFaninId1(pObj, i)), Gia_ObjFaninC1(pObj), 
                nWords ); }
        // hash sim info
        fCompl = int32((pSims[0] & 1));
        if fCompl { Abc_TtNot( pSims, nWords ); }
        Vec_IntWriteEntry( vMap, i, Vec_MemHashInsert(vStore, pSims) );
        if fCompl { Abc_TtNot( pSims, nWords ); }
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        var pSims *word  = Sbd_ObjSims(p, Gia_ObjId(p, pObj));
        Abc_TtCopy( pSims, Sbd_ObjSims(p, Gia_ObjFaninId0p(p, pObj)), nWords, Gia_ObjFaninC0(pObj) );
//        printf( "%d ", Abc_TtCountOnesVec(pSims, nWords) );
        assert(Gia_ObjPhase(pObj) == int32((pSims[0] & 1)), "Gia_ObjPhase(pObj) == int32((pSims[0] & 1))");
    }
//    printf( "\n" );
    Vec_MemHashFree( vStore );
    Vec_MemFree( vStore );
    fmt.Printf("Objects = %6d.  Unique = %6d.\n", Gia_ManAndNum(p), Vec_IntCountUnique(vMap));
    if pvMap {
        *pvMap = vMap;
    } else {
        Vec_IntFree( vMap ); }
}
func Sbd_GiaSimTest(pGia *Gia_Man_t) {
    Gia_ManSetPhase( pGia );

    // allocate simulation info
    pGia.iPatsPi = 32;
    pGia.vSims   = Vec_WrdStart( Gia_ManObjNum(pGia) * pGia.iPatsPi );
    pGia.vSimsPi = Vec_WrdStart( Gia_ManObjNum(pGia) * pGia.iPatsPi );

    Gia_ManRandom( 1 );

    Sbd_GiaSimRound( pGia, 0, nil );
    Sbd_GiaSimRound( pGia, 0, nil );
    Sbd_GiaSimRound( pGia, 0, nil );

    fmt.Printf("\n");
    Sbd_GiaSimRound( pGia, 1, nil );
    fmt.Printf("\n");
    Sbd_GiaSimRound( pGia, 1, nil );
    fmt.Printf("\n");
    Sbd_GiaSimRound( pGia, 1, nil );

    Vec_WrdFreeP( &pGia.vSims );
    Vec_WrdFreeP( &pGia.vSimsPi );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


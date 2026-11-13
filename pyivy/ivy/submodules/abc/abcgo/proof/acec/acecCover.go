package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecCover.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecCover.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"

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
func Gia_AcecExplore(p *Gia_Man_t, fVerbose int32) {
    var vNodes *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vFadds, vHadds, vXors *Vec_Int_t;
    var vMap *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var pObj *Gia_Obj_t; 
    var i, nSupp, nCone, nHadds int32;
    assert(p.pMuxes != nil, "p.pMuxes != nil");
    vFadds = Gia_ManDetectFullAdders( p, fVerbose, nil );
    vHadds = Gia_ManDetectHalfAdders( p, fVerbose );

    pObj = Gia_ManObj( p, 352 );
    fmt.Printf("Xor = %d.\n", Gia_ObjIsXor(pObj));
    fmt.Printf("Fanin0 = %d.  Fanin1 = %d.\n", Gia_ObjFaninId0(pObj, 352), Gia_ObjFaninId1(pObj, 352));
    fmt.Printf("Fan00 = %d.  Fan01 = %d.   Fan10 = %d.  Fan11 = %d.\n", Gia_ObjFaninId0(Gia_ObjFanin0(pObj), Gia_ObjFaninId0(pObj, 352)), Gia_ObjFaninId1(Gia_ObjFanin0(pObj), Gia_ObjFaninId0(pObj, 352)), Gia_ObjFaninId0(Gia_ObjFanin1(pObj), Gia_ObjFaninId1(pObj, 352)), Gia_ObjFaninId1(Gia_ObjFanin1(pObj), Gia_ObjFaninId1(pObj, 352)));

    // create a map of all HADD/FADD outputs
    for  i = 0; i < Vec_IntSize(vHadds)/2; i++  {
        var pSigs *int32  = Vec_IntEntryP(vHadds, 2*i);
        Vec_BitWriteEntry( vMap, pSigs[0], 1 );
        Vec_BitWriteEntry( vMap, pSigs[1], 1 );
    }
    for  i = 0; i < Vec_IntSize(vFadds)/5; i++  {
        var pSigs *int32  = Vec_IntEntryP(vFadds, 5*i);
        Vec_BitWriteEntry( vMap, pSigs[3], 1 );
        Vec_BitWriteEntry( vMap, pSigs[4], 1 );
    }

    Gia_ManCleanMark01( p );

    // mark outputs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ObjFanin0(pObj).fMark0 = 1;
    }

    // collect XORs
    vXors = Gia_AcecCollectXors( p, vMap );
    Vec_BitFree( vMap );

    fmt.Printf("Collected XORs: ");
    Vec_IntPrint( vXors );

    // mark their fanins
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vXors)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vXors,i))); i++  {
        pObj.fMark1 = 1;
        Gia_ObjFanin0(pObj).fMark0 = 1;
        Gia_ObjFanin1(pObj).fMark0 = 1;
    }

    // mark FADDs
    for  i = 0; i < Vec_IntSize(vFadds)/5; i++  {
        Gia_AcecMarkFadd( p, Vec_IntEntryP(vFadds, 5*i) );
    }

    // iterate through HADDs and find those that fit in
    for {
        var fChange int32;
        for  i = 0; i < Vec_IntSize(vHadds)/2; i++  {
            var pSigs *int32  = Vec_IntEntryP(vHadds, 2*i);
            if !Gia_ManObj(p, pSigs[0]).fMark0 || !Gia_ManObj(p, pSigs[1]).fMark0 {
                continue; }
            if Gia_ManObj(p, pSigs[0]).fMark1 || Gia_ManObj(p, pSigs[1]).fMark1 {
                continue; }
            Gia_AcecMarkHadd( p, pSigs );
            fChange = 1;
            nHadds++;
        }
        if !(fChange != 0) {
            break; }
    }
    // print inputs to the adder network
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if pObj.fMark0 && !pObj.fMark1 {
            nSupp = Gia_ManSuppSize( p, &i, 1 );
            nCone = Gia_ManConeSize( p, &i, 1 );
            fmt.Printf("Node %5d : Supp = %5d.  Cone = %5d.\n", i, nSupp, nCone);
            Vec_IntPush( vNodes, i );
        }
    }
    fmt.Printf("Fadds = %d. Hadds = %d.  Root nodes found = %d.\n", Vec_IntSize(vFadds)/5, nHadds, Vec_IntSize(vNodes));

    Gia_ManCleanMark01( p );

    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
        pObj.fMark0 = 1;
    }

    Vec_IntFree( vFadds );
    Vec_IntFree( vHadds );
    Vec_IntFree( vNodes );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_AcecCover(p *Gia_Man_t) {
    var fVerbose int32  = 1;
    var i, k, Entry int32;
    var pObj *Gia_Obj_t;
    var vCutsXor2 *Vec_Int_t  = nil;
    var vFadds *Vec_Int_t  = Gia_ManDetectFullAdders( p, fVerbose, &vCutsXor2 );

    // mark FADDs
    Gia_ManCleanMark01( p );
    for  i = 0; i < Vec_IntSize(vFadds)/5; i++  {
        Gia_AcecMarkFadd( p, Vec_IntEntryP(vFadds, 5*i) );
    }

    k = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCutsXor2)) && (((Entry) = Vec_IntEntry(vCutsXor2, i)), 1); i++  {
        if i % 3 != 2 {
            continue; }
        pObj = Gia_ManObj( p, Entry );
        if pObj.fMark1 {
            continue; }
        fmt.Printf("%d ", Entry);
    }
    fmt.Printf("\n");

    Gia_ManCleanMark01( p );

    Vec_IntFree( vFadds );
    Vec_IntFree( vCutsXor2 );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


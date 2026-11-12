package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaIso3.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaIso3.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var Iso_Nodes [6]uint32 = [6]uint32{ 0x04892ed6, 0xc2523d7d, 0xdc36cd2e, 0xf2db64f8, 0xde3126bb, 0xdebbdff0 }; // ab, a!b, !a!b, pi, po, const0
var Iso_Fanio [2]uint32 = [2]uint32{ 0x855ee0cf, 0x946e1b5f }; // fanin, fanout
var Iso_Compl [2]uint32 = [2]uint32{ 0x8ba63e50, 0x14d87f02 }; // non-compl, compl

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Iso3Node(pObj *Gia_Obj_t) uint32 {
    if Gia_ObjIsAnd(pObj) {
        return Iso_Nodes[Gia_ObjFaninC0(pObj) + Gia_ObjFaninC1(pObj)]; }
    if Gia_ObjIsCi(pObj) {
        return Iso_Nodes[3]; }
    if Gia_ObjIsCo(pObj) {
        return Iso_Nodes[4]; }
    return Iso_Nodes[5];
}
func Gia_Iso3Init(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        pObj.Value = Gia_Iso3Node( pObj );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Iso3ComputeEdge(p *Gia_Man_t, pObj *Gia_Obj_t, pFanin *Gia_Obj_t, fCompl int32, vSign *Vec_Int_t) {
    pObj.Value   += Vec_IntEntry(vSign, Gia_ObjId(p, pFanin)) + Iso_Compl[fCompl] + Iso_Fanio[0];
    pFanin.Value += Vec_IntEntry(vSign, Gia_ObjId(p, pObj))   + Iso_Compl[fCompl] + Iso_Fanio[1];
}
func Gia_Iso3Compute(p *Gia_Man_t, vSign *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) || Gia_ObjIsCo(pObj) {
            Gia_Iso3ComputeEdge( p, pObj, Gia_ObjFanin0(pObj), Gia_ObjFaninC0(pObj), vSign ); }
        if Gia_ObjIsAnd(pObj) {
            Gia_Iso3ComputeEdge( p, pObj, Gia_ObjFanin1(pObj), Gia_ObjFaninC1(pObj), vSign ); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Iso3Save(p *Gia_Man_t) *Vec_Int_t {
    var vSign *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    vSign = Vec_IntAlloc( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        Vec_IntPush( vSign, pObj.Value );
    }
    return vSign;
}
func Gia_Iso3Unique(vSign *Vec_Int_t) int32 {
    var nUnique int32;
    var vCopy *Vec_Int_t  = Vec_IntDup( vSign );
    Vec_IntUniqify( vCopy );
    nUnique = Vec_IntSize(vCopy);
    Vec_IntFree( vCopy );
    return nUnique;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Iso3Test(p *Gia_Man_t) {
    var nIterMax int32  = 500;
    var i int32;
    var Prev int32 = -1;
    var This int32; 
    var clk abctime  = Abc_Clock();
    var vSign *Vec_Int_t  = nil;
    Gia_Iso3Init( p );
    for  i = 0; i < nIterMax; i++  {
        vSign = Gia_Iso3Save( p );
//        This = Gia_Iso3Unique( vSign );
        This = Vec_IntUniqueCount( vSign, 1, nil );
        fmt.Printf("Iter %3d : %6d  out of %6d  ", i, This, Vec_IntSize(vSign));
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        if This == Prev {
            break; }
        Prev = This;
        Gia_Iso3Compute( p, vSign );
        Vec_IntFreeP( &vSign );
    }
    Vec_IntFreeP( &vSign );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Iso4Gia(p *Gia_Man_t) *Vec_Wec_t {
    var vLevs *Vec_Wec_t  = Gia_ManLevelizeR( p );
    var vLevel *Vec_Int_t; var l int32;
    Abc_Random( 1 );
    // inlining of macro Vec_WecForEachLevel
    for  l = 0; (l < Vec_WecSize(vLevs)) && (((vLevel) = Vec_WecEntry(vLevs, l)), 1); l++  {
        var pObj *Gia_Obj_t;  var i int32;
        var RandC [2]int32 = [2]int32{ Abc_Random(0), Abc_Random(0) };
        if l == 0 {
            // inlining of macro Gia_ManForEachObjVec
            for  i = 0; (i < Vec_IntSize(vLevel)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vLevel,i))); i++  {
                assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
                pObj.Value = Abc_Random(0);
                Gia_ObjFanin0(pObj).Value += pObj.Value + RandC[Gia_ObjFaninC0(pObj)];
            }
        
        } else {
            // inlining of macro Gia_ManForEachObjVec
            for  i = 0; (i < Vec_IntSize(vLevel)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vLevel,i))); i++  { if Gia_ObjIsAnd(pObj) {
                Gia_ObjFanin0(pObj).Value += pObj.Value + RandC[Gia_ObjFaninC0(pObj)];
                Gia_ObjFanin1(pObj).Value += pObj.Value + RandC[Gia_ObjFaninC1(pObj)];
            }
            }
        }
    }
    return vLevs;
}
func Gia_Iso4Test(p *Gia_Man_t) {
    var vLevs *Vec_Wec_t  = Gia_Iso4Gia( p );
    var vLevel *Vec_Int_t; var l int32;
    // inlining of macro Vec_WecForEachLevel
    for  l = 0; (l < Vec_WecSize(vLevs)) && (((vLevel) = Vec_WecEntry(vLevs, l)), 1); l++  {
        var pObj *Gia_Obj_t;  var i int32;
        fmt.Printf("Level %d\n", l);
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vLevel)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vLevel,i))); i++  {
            fmt.Printf("Obj = %5d.  Value = %08x.\n", Gia_ObjId(p, pObj), pObj.Value);
        }
    }
    Vec_WecFree( vLevs );
}
func Gia_IsoCollectData(p *Gia_Man_t, vObjs *Vec_Int_t) *Vec_Int_t {
    var pObj *Gia_Obj_t;  var i int32;
    var vData *Vec_Int_t  = Vec_IntAlloc( Vec_IntSize(vObjs) );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        Vec_IntPush( vData, pObj.Value );
    }
    return vData;
}
func Gia_IsoCompareVecs(pGia0 *Gia_Man_t, vLevs0 *Vec_Wec_t, pGia1 *Gia_Man_t, vLevs1 *Vec_Wec_t) {
    var i int32;
    var Common int32;
    var nLevels int32 = Abc_MinInt( Vec_WecSize(vLevs0), Vec_WecSize(vLevs1) );
    Gia_ManPrintStats( pGia0, nil );
    Gia_ManPrintStats( pGia1, nil );
    fmt.Printf("Printing %d shared levels:\n", nLevels);
    for  i = 0; i < nLevels; i++  {
        var vLev0 *Vec_Int_t  = Vec_WecEntry(vLevs0, i);
        var vLev1 *Vec_Int_t  = Vec_WecEntry(vLevs1, i);
        var vData0 *Vec_Int_t  = Gia_IsoCollectData( pGia0, vLev0 );
        var vData1 *Vec_Int_t  = Gia_IsoCollectData( pGia1, vLev1 );
        Vec_IntSort( vData0, 0 );
        Vec_IntSort( vData1, 0 );
        Common = Vec_IntTwoCountCommon( vData0, vData1 );
        fmt.Printf("Level = %3d. One = %6d. Two = %6d.  Common = %6d.\n", i, Vec_IntSize(vData0)-Common, Vec_IntSize(vData1)-Common, Common);
        Vec_IntFree( vData0 );
        Vec_IntFree( vData1 );        
    }
}
func Gia_Iso4TestTwo(pGia0 *Gia_Man_t, pGia1 *Gia_Man_t) {
    var vLevs0 *Vec_Wec_t  = Gia_Iso4Gia( pGia0 );
    var vLevs1 *Vec_Wec_t  = Gia_Iso4Gia( pGia1 );
    Gia_IsoCompareVecs( pGia0, vLevs0, pGia1, vLevs1 );
    Vec_WecFree( vLevs0 );
    Vec_WecFree( vLevs1 );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


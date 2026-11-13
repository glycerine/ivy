package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecCo.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecCo.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecWec.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Collect non-XOR inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynCoreOrderArray(pGia *Gia_Man_t, vAdds *Vec_Int_t, vRootBoxes *Vec_Int_t) *Vec_Wec_t {
    extern Vec_Bit_t * Acec_ManPoolGetPointed( Gia_Man_t * p, Vec_Int_t * vAdds );
    var vMarks *Vec_Bit_t  = Acec_ManPoolGetPointed( pGia, vAdds );
    var vMap *Vec_Wec_t    = Gia_PolynComputeMap( vAdds, Gia_ManObjNum(pGia) );
    var vRes *Vec_Wec_t    = Vec_WecStart( Vec_IntSize(vRootBoxes) );
    var vRoots *Vec_Int_t  = Vec_IntAlloc( 64 );
    var vOrder *Vec_Int_t;
    var i, iBox int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRootBoxes)) && (((iBox) = Vec_IntEntry(vRootBoxes, i)), 1); i++  {
        Gia_PolyCollectRoots( vAdds, vMap, vMarks, iBox, vRoots );
        vOrder = Gia_PolynCoreOrder_int( pGia, vAdds, vMap, vRoots, nil );
        Vec_IntAppend( Vec_WecEntry(vRes, i), vOrder );
        Vec_IntFree( vOrder );
    }
    Vec_BitFree( vMarks );
    Vec_IntFree( vRoots );
    Vec_WecFree( vMap );
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Collect internal node order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynCoreCollect_rec(pGia *Gia_Man_t, iObj int32, vNodes *Vec_Int_t, vVisited *Vec_Bit_t) {
    if Vec_BitEntry(vVisited, iObj) {
        return; }
    Vec_BitSetEntry( vVisited, iObj, 1 );
    Gia_PolynCoreCollect_rec( pGia, Gia_ObjFaninId0p(pGia, Gia_ManObj(pGia, iObj)), vNodes, vVisited );
    Gia_PolynCoreCollect_rec( pGia, Gia_ObjFaninId1p(pGia, Gia_ManObj(pGia, iObj)), vNodes, vVisited );
    Vec_IntPush( vNodes, iObj );
}
func Gia_PolynCoreCollect(pGia *Gia_Man_t, vAdds *Vec_Int_t, vOrder *Vec_Int_t) *Vec_Int_t {
    var vNodes *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var vVisited *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(pGia) );
    var i, Index, Entry1, Entry2, Entry3 int32;
    // inlining of macro Vec_IntForEachEntryReverse
    for  i = Vec_IntSize(vOrder) - 1; (i >= 0) && (((Index) = Vec_IntEntry(vOrder, i)), 1); i--  {
        // mark inputs
        Entry1 = Vec_IntEntry( vAdds, 6*Index + 0 );
        Entry2 = Vec_IntEntry( vAdds, 6*Index + 1 );
        Entry3 = Vec_IntEntry( vAdds, 6*Index + 2 );
        Vec_BitWriteEntry( vVisited, Entry1, 1 );
        Vec_BitWriteEntry( vVisited, Entry2, 1 );
        Vec_BitWriteEntry( vVisited, Entry3, 1 );
        // traverse from outputs
        Entry1 = Vec_IntEntry( vAdds, 6*Index + 3 );
        Entry2 = Vec_IntEntry( vAdds, 6*Index + 4 );
        Gia_PolynCoreCollect_rec( pGia, Entry1, vNodes, vVisited );
        Gia_PolynCoreCollect_rec( pGia, Entry2, vNodes, vVisited );
    }
    Vec_BitFree( vVisited );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynCorePrintCones(pGia *Gia_Man_t, vLeaves *Vec_Int_t, fVerbose int32) {
    var i, iObj int32;
    if fVerbose != 0 {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLeaves)) && (((iObj) = Vec_IntEntry(vLeaves, i)), 1); i++  {
            fmt.Printf("%4d : ", i);
            fmt.Printf("Supp = %3d  ", Gia_ManSuppSize(pGia, &iObj, 1));
            fmt.Printf("Cone = %3d  ", Gia_ManConeSize(pGia, &iObj, 1));
            fmt.Printf("\n");
        }
    
    } else {
        var SuppMax, ConeMax int32;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLeaves)) && (((iObj) = Vec_IntEntry(vLeaves, i)), 1); i++  {
            SuppMax = Abc_MaxInt( SuppMax, Gia_ManSuppSize(pGia, &iObj, 1) );
            ConeMax = Abc_MaxInt( ConeMax, Gia_ManConeSize(pGia, &iObj, 1) );
        }
        fmt.Printf("Remaining cones:  Count = %d.  SuppMax = %d.  ConeMax = %d.\n", Vec_IntSize(vLeaves), SuppMax, ConeMax);
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynCoreDupTreePlus_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if ~pObj.Value {
        return pObj.Value; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_PolynCoreDupTreePlus_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Gia_PolynCoreDupTreePlus_rec( pNew, p, Gia_ObjFanin1(pObj) );
    return pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
} 
func Gia_PolynCoreDupTree(p *Gia_Man_t, vAddCos *Vec_Int_t, vLeaves *Vec_Int_t, vNodes *Vec_Int_t, fAddCones int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    if fAddCones != 0 {
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.Value = Gia_ManAppendCi(pNew);
        }
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vLeaves)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vLeaves,i))); i++  {
            pObj.Value = Gia_PolynCoreDupTreePlus_rec( pNew, p, pObj );
        }
    
    } else {
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vLeaves)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vLeaves,i))); i++  {
            pObj.Value = Gia_ManAppendCi(pNew);
        }
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vAddCos)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vAddCos,i))); i++  {
        Gia_ManAppendCo( pNew, pObj.Value );
    }
    return pNew;

}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynCoreDetectTest_int(pGia *Gia_Man_t, vAddCos *Vec_Int_t, fAddCones int32, fVerbose int32) *Gia_Man_t {
    extern Vec_Int_t * Ree_ManComputeCuts( Gia_Man_t * p, Vec_Int_t ** pvXors, int fVerbose );
    var clk abctime  = Abc_Clock();
    var pNew *Gia_Man_t;
    var vAdds *Vec_Int_t  = Ree_ManComputeCuts( pGia, nil, 1 );
    var vLeaves *Vec_Int_t;
    var vRoots *Vec_Int_t;
    var vOrder *Vec_Int_t = Gia_PolynCoreOrder( pGia, vAdds, vAddCos, &vLeaves, &vRoots );
    var vNodes *Vec_Int_t  = Gia_PolynCoreCollect( pGia, vAdds, vOrder );

    //Gia_ManShow( pGia, vNodes, 0 );
    
    fmt.Printf("Detected %d FAs/HAs. Roots = %d. Leaves = %d. Nodes = %d. Adds = %d. ", Vec_IntSize(vAdds)/6, Vec_IntSize(vLeaves), Vec_IntSize(vRoots), Vec_IntSize(vNodes), Vec_IntSize(vOrder));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    Gia_PolynCorePrintCones( pGia, vLeaves, fVerbose );

    pNew = Gia_PolynCoreDupTree( pGia, vAddCos, vLeaves, vNodes, fAddCones );

    Vec_IntFree( vAdds );
    Vec_IntFree( vLeaves );
    Vec_IntFree( vRoots );
    Vec_IntFree( vOrder );
    Vec_IntFree( vNodes );
    return pNew;
}
func Gia_PolynCoreDetectTest(pGia *Gia_Man_t, fAddExtra int32, fAddCones int32, fVerbose int32) *Gia_Man_t {
    var vAddCos *Vec_Int_t;
    if fAddExtra != 0 {
        vAddCos = Gia_PolynAddHaRoots( pGia );
    } else {
        vAddCos = Vec_IntAlloc(0);
    }
    var pNew *Gia_Man_t  = Gia_PolynCoreDetectTest_int( pGia, vAddCos, fAddCones, fVerbose );
    fmt.Printf("On top of %d COs, created %d new adder outputs.\n", Gia_ManCoNum(pGia), Vec_IntSize(vAddCos));
    Vec_IntFree( vAddCos );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


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
func Gia_PolynCoreNonXors_rec(pGia *Gia_Man_t, pObj *Gia_Obj_t, vXorPairs *Vec_Int_t) {
    extern int Gia_ManSuppSizeOne( Gia_Man_t * p, Gia_Obj_t * pObj );
    var pFan0, pFan1 *Gia_Obj_t;
    if !Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) {
        return; }
    Gia_PolynCoreNonXors_rec( pGia, Gia_Regular(pFan0), vXorPairs );
    Gia_PolynCoreNonXors_rec( pGia, Gia_Regular(pFan1), vXorPairs );
    //if ( Gia_ManSuppSizeOne(pGia, pObj) > 4 )
    //if ( Gia_ObjId(pGia, pObj) >= 73 )
    Vec_IntPushTwo( vXorPairs, Gia_ObjId(pGia, Gia_Regular(pFan0)), Gia_ObjId(pGia, Gia_Regular(pFan1)) );
}
func Gia_PolynAddHaRoots(pGia *Gia_Man_t) *Vec_Int_t {
    var i, iFan0, iFan1 int32;
    var vNewOuts *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vXorPairs *Vec_Int_t  = Vec_IntAlloc( 100 );
    var pObj *Gia_Obj_t  = Gia_ManCo( pGia, Gia_ManCoNum(pGia)-1 );
    Gia_PolynCoreNonXors_rec( pGia, Gia_ObjFanin0(pObj), vXorPairs );
    // add new outputs
    Gia_ManSetPhase( pGia );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vXorPairs)) && (((iFan0) = Vec_IntEntry(vXorPairs, i)), 1) && (((iFan1) = Vec_IntEntry(vXorPairs, i+1)), 1); i += 2  {
        var pFan0 *Gia_Obj_t  = Gia_ManObj( pGia, iFan0 );
        var pFan1 *Gia_Obj_t  = Gia_ManObj( pGia, iFan1 );
        var iLit0 int32  = Abc_Var2Lit( iFan0, pFan0.fPhase );
        var iLit1 int32  = Abc_Var2Lit( iFan1, pFan1.fPhase );
        var iRoot int32  = Gia_ManAppendAnd( pGia, iLit0, iLit1 );
        Vec_IntPush( vNewOuts, Abc_Lit2Var(iRoot) );
    }
    Vec_IntFree( vXorPairs );
    Vec_IntReverseOrder( vNewOuts );
//    Vec_IntPop( vNewOuts );
    return vNewOuts;
}


/**Function*************************************************************

  Synopsis    [Detects the order of adders.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynComputeMap(vAdds *Vec_Int_t, nObjs int32) *Vec_Wec_t {
    // map nodes driven by adders into adder indexes
    var vMap *Vec_Wec_t  = Vec_WecStart( nObjs );  var i int32;
    for  i = 0; 6*i < Vec_IntSize(vAdds); i++  {
        var Entry1 int32  = Vec_IntEntry( vAdds, 6*i + 3 );
        var Entry2 int32  = Vec_IntEntry( vAdds, 6*i + 4 );
        Vec_WecPush( vMap, Entry1, i );
        Vec_WecPush( vMap, Entry1, Entry2 );
        Vec_WecPush( vMap, Entry2, i );
        Vec_WecPush( vMap, Entry2, Entry1 );
    }
    return vMap;
}
func Gia_PolynCoreOrder_int(pGia *Gia_Man_t, vAdds *Vec_Int_t, vMap *Vec_Wec_t, vRoots *Vec_Int_t, pvIns **Vec_Int_t) *Vec_Int_t {
    var vOrder *Vec_Int_t   = Vec_IntAlloc( 1000 );
    var vIsRoot *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(pGia) );
    var i int32;
    var k int32;
    var Index int32 = -1;
    var Driver int32;
    var Entry1 int32;
    var Entry2 int32 = -1;
    // mark roots
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRoots)) && (((Driver) = Vec_IntEntry(vRoots, i)), 1); i++  {
        Vec_BitWriteEntry( vIsRoot, Driver, 1 );
    }
    // collect boxes
    for {
        // iterate through boxes driving this one
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vRoots)) && (((Entry1) = Vec_IntEntry(vRoots, i)), 1); i++  {
            var vLevel *Vec_Int_t  = Vec_WecEntry( vMap, Entry1 );
            // inlining of macro Vec_IntForEachEntryDouble
            for  k = 0; (k+1 < Vec_IntSize(vLevel)) && (((Index) = Vec_IntEntry(vLevel, k)), 1) && (((Entry2) = Vec_IntEntry(vLevel, k+1)), 1); k += 2  {
                if Vec_BitEntry(vIsRoot, Entry2) {
                    break; }
            }
            if k == Vec_IntSize(vLevel) {
                continue; }
            assert(Vec_BitEntry(vIsRoot, Entry1), "Vec_BitEntry(vIsRoot, Entry1)");
            assert(Vec_BitEntry(vIsRoot, Entry2), "Vec_BitEntry(vIsRoot, Entry2)");
            // collect adder
            Vec_IntPush( vOrder, Index );
            // clean marks
            Vec_BitWriteEntry( vIsRoot, Entry1, 0 );
            Vec_BitWriteEntry( vIsRoot, Entry2, 0 );
            Vec_IntRemove( vRoots, Entry1 );
            Vec_IntRemove( vRoots, Entry2 );
            // set new marks
            Entry1 = Vec_IntEntry( vAdds, 6*Index + 0 );
            Entry2 = Vec_IntEntry( vAdds, 6*Index + 1 );
            Driver = Vec_IntEntry( vAdds, 6*Index + 2 );
            Vec_BitWriteEntry( vIsRoot, Entry1, 1 );
            Vec_BitWriteEntry( vIsRoot, Entry2, 1 );
            Vec_BitWriteEntry( vIsRoot, Driver, 1 );
            Vec_IntPushUnique( vRoots, Entry1 );
            Vec_IntPushUnique( vRoots, Entry2 );
            Vec_IntPushUnique( vRoots, Driver );
            break;
        }
        if i == Vec_IntSize(vRoots) {
            break; }
    }
    // collect remaining leaves
    if pvIns {
        *pvIns = Vec_IntAlloc( Vec_BitSize(vIsRoot) );
        // inlining of macro Vec_BitForEachEntryStart
        for  i = 1; (i < Vec_BitSize(vIsRoot)) && (((Driver) = Vec_BitEntry(vIsRoot, i)), 1); i++  {
            if Driver != 0 {
                Vec_IntPush( *pvIns, i ); }
        }
    }
    Vec_BitFree( vIsRoot );
    return vOrder;
}
func Gia_PolynCoreOrder(pGia *Gia_Man_t, vAdds *Vec_Int_t, vAddCos *Vec_Int_t, pvIns **Vec_Int_t, pvOuts **Vec_Int_t) *Vec_Int_t {
    var vOrder *Vec_Int_t;
    var vMap *Vec_Wec_t  = Gia_PolynComputeMap( vAdds, Gia_ManObjNum(pGia) );
    var vRoots *Vec_Int_t  = Vec_IntAlloc( Gia_ManCoNum(pGia) );
    var i, Driver int32;
    // collect roots
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && (((Driver) = Gia_ObjFaninId0p(pGia, Gia_ManCo(pGia, i))), 1); i++  {
        Vec_IntPush( vRoots, Driver );
    }
    // collect additional outputs
    if vAddCos != nil {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vAddCos)) && (((Driver) = Vec_IntEntry(vAddCos, i)), 1); i++  {
            Vec_IntPush( vRoots, Driver );
        } }
    // remember roots
    if pvOuts {
        *pvOuts = Vec_IntDup( vRoots ); }
    // create order
    vOrder = Gia_PolynCoreOrder_int( pGia, vAdds, vMap, vRoots, pvIns );
    Vec_IntFree( vRoots );
    Vec_WecFree( vMap );
    fmt.Printf("Collected %d boxes.\n", Vec_IntSize(vOrder));
    return vOrder;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolyCollectRoots_rec(vAdds *Vec_Int_t, vMap *Vec_Wec_t, vMarks *Vec_Bit_t, iBox int32, vRoots *Vec_Int_t) {
    var k int32;
    for  k = 0; k < 3; k++  {
        var i int32;
        var Index int32;
        var Sum int32;
        var Carry int32 = Vec_IntEntry( vAdds, 6*iBox+k );
        var vLevel *Vec_Int_t  = Vec_WecEntry( vMap, Carry );
        if Carry == 0 {
            continue; }
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(vLevel)) && (((Index) = Vec_IntEntry(vLevel, i)), 1) && (((Sum) = Vec_IntEntry(vLevel, i+1)), 1); i += 2  {
            if Vec_IntEntry(vAdds, 6*Index+4) == Carry && !Vec_BitEntry(vMarks, Sum) {
                Vec_IntPush( vRoots, Sum );
                Gia_PolyCollectRoots_rec( vAdds, vMap, vMarks, Index, vRoots );
            }
        }
    }
}
func Gia_PolyCollectRoots(vAdds *Vec_Int_t, vMap *Vec_Wec_t, vMarks *Vec_Bit_t, iBox int32, vRoots *Vec_Int_t) {
    Vec_IntClear( vRoots );
    Vec_IntPush( vRoots, Vec_IntEntry(vAdds, 6*iBox+3) );
    Vec_IntPush( vRoots, Vec_IntEntry(vAdds, 6*iBox+4) );
    Gia_PolyCollectRoots_rec( vAdds, vMap, vMarks, iBox, vRoots );
}
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


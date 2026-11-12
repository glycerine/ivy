package saig

import "fmt"

/**CFile****************************************************************

  FileName    [aigIsoFast.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Graph isomorphism package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigIsoFast.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const AIG_ISO_NUM = 16

   
type Iso_Dat_t struct {
    nFiNeg uint32; /* C bitfield: 3 bits; review packing and truncation */
    nFoNeg uint32; /* C bitfield: 2 bits; review packing and truncation */
    nFoPos uint32; /* C bitfield: 2 bits; review packing and truncation */
    Fi0Lev uint32; /* C bitfield: 3 bits; review packing and truncation */
    Fi1Lev uint32; /* C bitfield: 3 bits; review packing and truncation */
    Level uint32; /* C bitfield: 3 bits; review packing and truncation */
    fVisit uint32; /* C bitfield: 16 bits; review packing and truncation */
}

   
type Iso_Dat2_t struct {
    Data uint32; /* C bitfield: 16 bits; review packing and truncation */
}

   
type Iso_Sto_t struct {
    pAig    *Aig_Man_t;       // user's AIG manager
    nObjs           int32;      // number of objects
    pData    *Iso_Dat_t;      // data for each object
    vVisited    *Vec_Int_t;   // visited nodes
    vRoots    *Vec_Ptr_t;     // root nodes
    vPlaces    *Vec_Int_t;    // places in the counter lists
    pCounters          *int32;  // counters    
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Iso_StoStart(pAig *Aig_Man_t) *Iso_Sto_t {
    
    // inlining of macro ABC_CALLOC
    p := &Iso_Sto_t{};
    p.pAig      = pAig;
    p.nObjs     = Aig_ManObjNumMax( pAig );
    // inlining of macro ABC_CALLOC
    p.pData     = ((Iso_Dat_t *) calloc((size_t)(p.nObjs), sizeof(Iso_Dat_t)));
    p.vVisited  = Vec_IntStart( 1000 );
    p.vPlaces   = Vec_IntStart( 1000 );
    p.vRoots    = Vec_PtrStart( 1000 );
    // inlining of macro ABC_CALLOC
    p.pCounters = ((int *) calloc((size_t)((1 << AIG_ISO_NUM)), sizeof(int)));
    return p;
}
func Iso_StoStop(p *Iso_Sto_t) {
    Vec_IntFree( p.vPlaces );
    Vec_IntFree( p.vVisited );
    Vec_PtrFree( p.vRoots );
    p.pCounters = nil;
    p.pData = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Collect statistics about one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Iso_StoCollectInfo_rec(p *Aig_Man_t, pObj *Aig_Obj_t, fCompl int32, vVisited *Vec_Int_t, pData *Iso_Dat_t, vRoots *Vec_Ptr_t) {
    var pThis *Iso_Dat_t  = pData + Aig_ObjId(pObj);
    assert(Aig_ObjIsCi(pObj) || Aig_ObjIsNode(pObj), "Aig_ObjIsCi(pObj) || Aig_ObjIsNode(pObj)");
    if pThis.fVisit {
        if fCompl != 0 {
            pThis.nFoNeg++;
        } else {
            pThis.nFoPos++; }
        return;
    }
    assert(*((int *)pThis) == 0, "*((int *)pThis) == 0");
    pThis.fVisit = 1;
    if fCompl != 0 {
        pThis.nFoNeg++;
    } else {
        pThis.nFoPos++; }
    pThis.Level = pObj.Level;
    pThis.nFiNeg = Aig_ObjFaninC0(pObj) + Aig_ObjFaninC1(pObj);
    if Aig_ObjIsNode(pObj) {
        if Aig_ObjFaninC0(pObj) < Aig_ObjFaninC1(pObj) || (Aig_ObjFaninC0(pObj) == Aig_ObjFaninC1(pObj) && Aig_ObjFanin0(pObj).Level < Aig_ObjFanin1(pObj).Level) {
            pThis.Fi0Lev = pObj.Level - Aig_ObjFanin0(pObj).Level;
            pThis.Fi1Lev = pObj.Level - Aig_ObjFanin1(pObj).Level;
        
        } else {
            pThis.Fi0Lev = pObj.Level - Aig_ObjFanin1(pObj).Level;
            pThis.Fi1Lev = pObj.Level - Aig_ObjFanin0(pObj).Level;
        }
        Iso_StoCollectInfo_rec( p, Aig_ObjFanin0(pObj), Aig_ObjFaninC0(pObj), vVisited, pData, vRoots );
        Iso_StoCollectInfo_rec( p, Aig_ObjFanin1(pObj), Aig_ObjFaninC1(pObj), vVisited, pData, vRoots );
    
    } else if Saig_ObjIsLo(p, pObj) {
        pThis.Fi0Lev = 1;
        pThis.Fi1Lev = 0;
        Vec_PtrPush( vRoots, Saig_ObjLoToLi(p, pObj) );
    
    } else if Saig_ObjIsPi(p, pObj) {
        pThis.Fi0Lev = 0;
        pThis.Fi1Lev = 0;
    
    } else {
        assert(false, "0"); }
    assert(pThis.nFoNeg + pThis.nFoPos, "pThis.nFoNeg + pThis.nFoPos");
    Vec_IntPush( vVisited, Aig_ObjId(pObj) );
}

//static abctime time_Trav = 0;

/**Function*************************************************************

  Synopsis    [Collect statistics about one output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Iso_StoCollectInfo(p *Iso_Sto_t, pPo *Aig_Obj_t) *Vec_Int_t {
    var fVerboseShow int32;
    var vInfo *Vec_Int_t;
    var pData2 *Iso_Dat2_t  = (Iso_Dat2_t *)p.pData;
    var pAig *Aig_Man_t  = p.pAig;
    var pObj *Aig_Obj_t;
    var i int32;
    var Value int32;
    var Entry int32;
    var pPerm *int32;
//    abctime clk = Abc_Clock();

    assert(Aig_ObjIsCo(pPo), "Aig_ObjIsCo(pPo)");

    // collect initial POs
    Vec_IntClear( p.vVisited );
    Vec_PtrClear( p.vRoots );
    Vec_PtrPush( p.vRoots, pPo );

    // mark internal nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vRoots)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vRoots, i)), 1); i++  {
        if !Aig_ObjIsConst1(Aig_ObjFanin0(pObj)) {
            Iso_StoCollectInfo_rec( pAig, Aig_ObjFanin0(pObj), Aig_ObjFaninC0(pObj), p.vVisited, p.pData, p.vRoots ); }
    }
//    time_Trav += Abc_Clock() - clk;

    // count how many times each data entry appears
    Vec_IntClear( p.vPlaces );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vVisited)) && (((Entry) = Vec_IntEntry(p.vVisited, i)), 1); i++  {
        Value = pData2[Entry].Data;
//        assert( Value > 0 );
        if p.pCounters[Value]++ == 0 {
            Vec_IntPush( p.vPlaces, Value ); }
//        pData2[Entry].Data = 0;
        *((int *)(p.pData + Entry)) = 0;
    }

    // collect non-trivial counters
    Vec_IntClear( p.vVisited );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vPlaces)) && (((Entry) = Vec_IntEntry(p.vPlaces, i)), 1); i++  {
        assert(p.pCounters[Entry], "p.pCounters[Entry]");
        Vec_IntPush( p.vVisited, p.pCounters[Entry] );
        p.pCounters[Entry] = 0;
    }
//    printf( "%d ", Vec_IntSize(p->vVisited) );

    // sort the costs in the increasing order
    pPerm = Abc_MergeSortCost( Vec_IntArray(p.vVisited), Vec_IntSize(p.vVisited) );
    assert(Vec_IntEntry(p.vVisited, pPerm[0]) <= Vec_IntEntry(p.vVisited, pPerm[Vec_IntSize(p.vVisited)-1]), "Vec_IntEntry(p.vVisited, pPerm[0]) <= Vec_IntEntry(p.vVisited, pPerm[Vec_IntSize(p.vVisited)-1])");

    // create information
    vInfo = Vec_IntAlloc( Vec_IntSize(p.vVisited) );
    for  i = Vec_IntSize(p.vVisited)-1; i >= 0; i--  {
        Entry = Vec_IntEntry( p.vVisited, pPerm[i] );
        Entry = (Entry << AIG_ISO_NUM) | Vec_IntEntry( p.vPlaces, pPerm[i] );
        Vec_IntPush( vInfo, Entry );
    }
    pPerm = nil;

    // show the final result
    if fVerboseShow != 0 {
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vInfo)) && (((Entry) = Vec_IntEntry(vInfo, i)), 1); i++  {
        var Data Iso_Dat2_t  = Iso_Dat2_t{ uint32(Entry) & 0xFFFF };
        var pData *Iso_Dat_t  = (Iso_Dat_t *)&Data;

        fmt.Printf("%6d : ", i);
        fmt.Printf("Freq =%6d ", Entry >> 16);

        fmt.Printf("FiNeg =%3d ", pData.nFiNeg);
        fmt.Printf("FoNeg =%3d ", pData.nFoNeg);
        fmt.Printf("FoPos =%3d ", pData.nFoPos);
        fmt.Printf("Fi0L  =%3d ", pData.Fi0Lev);
        fmt.Printf("Fi1L  =%3d ", pData.Fi1Lev);
        fmt.Printf("Lev   =%3d ", pData.Level);
        fmt.Printf("\n");
    } }
    return vInfo;
}

/**Function*************************************************************

  Synopsis    [Takes multi-output sequential AIG.]

  Description [Returns candidate equivalence classes of POs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Iso_StoCompareVecInt(p1 **Vec_Int_t, p2 **Vec_Int_t) int32 {
    return Vec_IntCompareVec( *p1, *p2 );
}

/**Function*************************************************************

  Synopsis    [Takes multi-output sequential AIG.]

  Description [Returns candidate equivalence classes of POs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_IsoDetectFast(pAig *Aig_Man_t) *Vec_Vec_t {
    var pMan *Iso_Sto_t;
    var pObj *Aig_Obj_t;
    var vClasses, vInfos *Vec_Ptr_t;
    var vInfo, vPrev, vLevel *Vec_Int_t;
    var i, Number, nUnique int32;
    var clk abctime  = Abc_Clock();

    // collect infos and remember their number
    pMan = Iso_StoStart( pAig );
    vInfos = Vec_PtrAlloc( Saig_ManPoNum(pAig) );
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        vInfo = Iso_StoCollectInfo(pMan, pObj);
        Vec_IntPush( vInfo, i );
        Vec_PtrPush( vInfos, vInfo );
    }
    Iso_StoStop( pMan );
    Abc_PrintTime( 1, "Info computation time", Abc_Clock() - clk );

    // sort the infos
    clk = Abc_Clock();
    Vec_PtrSort( vInfos, (int (*)(void))Iso_StoCompareVecInt );

    // create classes
    clk = Abc_Clock();
    vClasses = Vec_PtrAlloc( Saig_ManPoNum(pAig) );
    // start the first class
    Vec_PtrPush( vClasses, (vLevel = Vec_IntAlloc(4)) );
    vPrev = (Vec_Int_t *)Vec_PtrEntry( vInfos, 0 );
    Vec_IntPush( vLevel, Vec_IntPop(vPrev) );
    // consider other classes
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vInfos)) && (((vInfo) = (Vec_Int_t *)Vec_PtrEntry(vInfos, i)), 1); i++  {
        Number = Vec_IntPop( vInfo );
        if Vec_IntCompareVec(vPrev, vInfo) {
            Vec_PtrPush( vClasses, Vec_IntAlloc(4) ); }
        vLevel = (Vec_Int_t *)Vec_PtrEntryLast( vClasses );
        Vec_IntPush( vLevel, Number );
        vPrev = vInfo;
    }
    Vec_VecFree( (Vec_Vec_t *)vInfos );
    Abc_PrintTime( 1, "Sorting time", Abc_Clock() - clk );
//    Abc_PrintTime( 1, "Traversal time", time_Trav );

    // report the results
//    Vec_VecPrintInt( (Vec_Vec_t *)vClasses );
    fmt.Printf("Devided %d outputs into %d cand equiv classes.\n", Saig_ManPoNum(pAig), Vec_PtrSize(vClasses));

    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vClasses)) && (((vLevel) = (Vec_Int_t *)Vec_PtrEntry(vClasses, i)), 1); i++  {
        if Vec_IntSize(vLevel) > 1 {
            fmt.Printf("%d ", Vec_IntSize(vLevel));
        } else {
            nUnique++; }
    }
    fmt.Printf(" Unique = %d\n", nUnique);

//    return (Vec_Vec_t *)vClasses;
    Vec_VecFree( (Vec_Vec_t *)vClasses );
    return nil;
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


package sbd

import "fmt"

/**CFile****************************************************************

  FileName    [sbdPath.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Critical path.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sbdPath.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sbdInt.h"
//#include "misc/tim/tim.h"

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
func Sbc_ManAddInternalToPath_rec(p *Gia_Man_t, iObj int32, vPath *Vec_Bit_t) int32 {
    var pObj *Gia_Obj_t; 
    var k, iFan, Value int32;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return Vec_BitEntry(vPath, iObj); }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsCi(pObj) {
        return Vec_BitEntry(vPath, iObj); }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    // inlining of macro Gia_LutForEachFanin
    for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFan = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
        Value |= Sbc_ManAddInternalToPath_rec( p, iFan, vPath );
    }
    if Value != 0 {
        Vec_BitWriteEntry( vPath, iObj, 1 ); }
    return Value;
}
func Sbc_ManAddInternalToPath(p *Gia_Man_t, vPath *Vec_Bit_t) {
    var k, iFan, iObj int32;
    // inlining of macro Gia_ManForEachLut
    for  iObj = 1; iObj < Gia_ManObjNum(p); iObj++  { if !Gia_ObjIsLut(p, iObj) {} else {
        if !Vec_BitEntry(vPath, iObj) {
            continue; }
        Gia_ManIncrementTravId( p );
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFan = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
            Gia_ObjSetTravIdCurrentId(p, iFan);
        }
        Sbc_ManAddInternalToPath_rec( p, iObj, vPath );
    }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbc_ManCriticalPath_rec(p *Gia_Man_t, pLevels *int32, iObj int32, LevelFan int32, vPath *Vec_Bit_t, Slack int32) {
    var pObj *Gia_Obj_t; var k, iFan int32;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    Vec_BitWriteEntry( vPath, iObj, 1 );
    if Gia_ObjIsCi(pObj) {
        var pManTime *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
        var iBox int32;
        if pManTime != nil {
            iBox = Tim_ManBoxForCi( pManTime, Gia_ObjCioId(pObj) );
        } else {
            iBox = -1;
        }
        if iBox >= 0 {
            var curCo int32  = Tim_ManBoxInputFirst( pManTime, iBox );
            var nBoxInputs int32  = Tim_ManBoxInputNum( pManTime, iBox );
            for  k = 0; k < nBoxInputs; k++  {
                var pCo *Gia_Obj_t  = Gia_ManCo( p, curCo + k );
                var iDriver int32  = Gia_ObjFaninId0p( p, pCo );
                if (pLevels[iDriver]+Slack >= LevelFan-1) && (iDriver != 0) {
                    Sbc_ManCriticalPath_rec( p, pLevels, iDriver, pLevels[iDriver], vPath, Abc_MaxInt(0, pLevels[iDriver]+Slack-(LevelFan-1)) ); }
            }
        }
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    // inlining of macro Gia_LutForEachFanin
    for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFan = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
        if pLevels[iFan]+Slack >= LevelFan-1 {
            Sbc_ManCriticalPath_rec( p, pLevels, iFan, pLevels[iFan], vPath, Abc_MaxInt(0, pLevels[iFan]+Slack-(LevelFan-1)) ); }
    }
}
func Sbc_ManCriticalPath(p *Gia_Man_t) *Vec_Bit_t {
    var pLevels *int32 = nil;
    var k int32;
    var iDriver int32;
    var Slack int32 = 1;
    var nLevels int32;
    if p.pManTime {
        nLevels = Gia_ManLutLevelWithBoxes(p);
    } else {
        nLevels = Gia_ManLutLevel(p, &pLevels);
    }
    var vPath *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    if p.pManTime {
        pLevels = Vec_IntArray( p.vLevels ); }
    Gia_ManIncrementTravId( p );
    // inlining of macro Gia_ManForEachCoDriverId
    for  k = 0; (k < Vec_IntSize(p.vCos)) && (((iDriver) = Gia_ObjFaninId0p(p, Gia_ManCo(p, k))), 1); k++  {
        if (pLevels[iDriver] == nLevels) && (iDriver != 0) {
            Sbc_ManCriticalPath_rec( p, pLevels, iDriver, pLevels[iDriver], vPath, Slack ); }
    }
    if !p.pManTime {
        ABC_FREE( pLevels ); }
    Sbc_ManAddInternalToPath( p, vPath );
    return vPath;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sbc_ManDelayTrace(p *Gia_Man_t) {
    var vPath *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var i int32;
    var k int32;
    var iFan int32;
    var nLevels int32;
    var pLevels *int32;
    var nLuts, nNodes, nEdges, nEdgesAll int32;
    if !Gia_ManHasMapping(p) {
        fmt.Printf("No mapping is available.\n");
        return;
    }
    assert(Gia_ManHasMapping(p), "Gia_ManHasMapping(p)");
    // set critical CO drivers
    nLevels = Gia_ManLutLevel( p, &pLevels );
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && (((iFan) = Gia_ObjFaninId0p(p, Gia_ManCo(p, i))), 1); i++  {
        if pLevels[iFan] == nLevels {
            Vec_BitWriteEntry( vPath, iFan, 1 ); }
    }
    // set critical internal nodes
    // inlining of macro Gia_ManForEachLutReverse
    for  i = Gia_ManObjNum(p) - 1; i > 0; i--  { if !Gia_ObjIsLut(p, i) {} else { 
        nLuts++;
        if !Vec_BitEntry(vPath, i) {
            continue; }
        nNodes++;
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,i) && ((iFan = Gia_ObjLutFanins(p,i)[k]),1); k++  {
            if pLevels[iFan] +1 < pLevels[i] {
                continue; }
            assert(pLevels[iFan] + 1 == pLevels[i], "pLevels[iFan] + 1 == pLevels[i]");
            Vec_BitWriteEntry( vPath, iFan, 1 );
            nEdges++;
            //printf( "%d -> %d\n", i, iFan );
        }
    }
    }
    // inlining of macro Gia_ManForEachLut
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsLut(p, i) {} else {
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,i) && ((iFan = Gia_ObjLutFanins(p,i)[k]),1); k++  {
            nEdgesAll += (Vec_BitEntry(vPath, i) && Vec_BitEntry(vPath, iFan));
        } }
    }

    ABC_FREE( pLevels );
    Vec_BitFree( vPath );
    fmt.Printf("AIG = %d. LUT = %d. Lev = %d.   Path nodes = %d.  Path edges = %d. (%d.)\n", Gia_ManAndNum(p), nLuts, nLevels, nNodes, nEdges, nEdgesAll);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


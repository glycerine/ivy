package scl

import "fmt"

/**CFile****************************************************************

  FileName    [sclLoad.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Wire/gate load computations.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclLoad.c,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sclSize.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns estimated wire capacitances for each fanout count.]

  Description []
               
  SideEffects []`

  SeeAlso     []

***********************************************************************/
func Abc_SclFindWireCaps(pWL *SC_WireLoad, nFanoutMax int32) *Vec_Flt_t {
    var vCaps *Vec_Flt_t  = nil;
    var EntryPrev, EntryCur, Slope float32;
    var i, iPrev, k, Entry, EntryMax int32;
    assert(pWL != nil, "pWL != nil");
    // find the biggest fanout count
    EntryMax = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(&pWL.vFanout)) && (((Entry) = Vec_IntEntry(&pWL.vFanout, i)), 1); i++  {
        EntryMax = Abc_MaxInt( EntryMax, Entry );
    }
    // create the array
    vCaps = Vec_FltStart( Abc_MaxInt(nFanoutMax, EntryMax) + 1 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(&pWL.vFanout)) && (((Entry) = Vec_IntEntry(&pWL.vFanout, i)), 1); i++  {
        Vec_FltWriteEntry( vCaps, Entry, Vec_FltEntry(&pWL.vLen, i) * pWL.cap );
    }
    if Vec_FltEntry(vCaps, 1) == 0 {
        return vCaps; }
    // interpolate between the values
    assert(Vec_FltEntry(vCaps, 1) != 0, "Vec_FltEntry(vCaps, 1) != 0");
    iPrev = 1;
    EntryPrev = Vec_FltEntry(vCaps, 1);
    // inlining of macro Vec_FltForEachEntryStart
    for  i = 2; (i < Vec_FltSize(vCaps)) && (((EntryCur) = Vec_FltEntry(vCaps, i)), 1); i++  {
        if EntryCur == 0 {
            continue; }
        Slope = (EntryCur - EntryPrev) / (i - iPrev);
        for  k = iPrev + 1; k < i; k++  {
            Vec_FltWriteEntry( vCaps, k, EntryPrev + Slope * (k - iPrev) );
        }
        EntryPrev = EntryCur;
        iPrev = i;
    }
    // extrapolate after the largest value
    Slope = pWL.cap * pWL.slope;
    for  k = iPrev + 1; k < i; k++  {
        Vec_FltWriteEntry( vCaps, k, EntryPrev + Slope * (k - iPrev) );
    }
    // show
//    Vec_FltForEachEntry( vCaps, EntryCur, i )
//        printf( "%3d : %f\n", i, EntryCur );
    return vCaps;
}

/**Function*************************************************************

  Synopsis    [Computes load for all nodes in the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindWireLoad(vWireCaps *Vec_Flt_t, nFans int32) float32 {
    if vWireCaps == nil {
        return 0; }
    return Vec_FltEntry( vWireCaps, Abc_MinInt(nFans, Vec_FltSize(vWireCaps)-1) );
}
func Abc_SclAddWireLoad(p *SC_Man, pObj *Abc_Obj_t, fSubtr int32) {
    var Load float32  = Abc_SclFindWireLoad( p.vWireCaps, Abc_ObjFanoutNum(pObj) );
    Abc_SclObjLoad(p, pObj).rise += fSubtr ? -Load : Load;
    Abc_SclObjLoad(p, pObj).fall += fSubtr ? -Load : Load;
}
func Abc_SclComputeLoad(p *SC_Man) {
    var pObj, pFanin *Abc_Obj_t;
    var i, k int32;
    // clear load storage
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((p.pNtk).vObjs)) && (((pObj) = Abc_NtkObj(p.pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        var pLoad *SC_Pair  = Abc_SclObjLoad( p, pObj );
        if !Abc_ObjIsPo(pObj) {
            pLoad.rise = pLoad.fall = 0.0; }
    }
    }
    // add cell load
    // inlining of macro Abc_NtkForEachNode1
    for  i = 0; (i < Vec_PtrSize((p.pNtk).vObjs)) && (((pObj) = Abc_NtkObj(p.pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || !Abc_ObjFaninNum(pObj) {} else {
        var pCell *SC_Cell  = Abc_SclObjCell( pObj );
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            var pLoad *SC_Pair  = Abc_SclObjLoad( p, pFanin );
            var pPin *SC_Pin  = SC_CellPin( pCell, k );
            pLoad.rise += pPin.rise_cap;
            pLoad.fall += pPin.fall_cap;
        }
    }
    }
    // add PO load
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        var pLoadPo *SC_Pair  = Abc_SclObjLoad( p, pObj );
        var pLoad *SC_Pair  = Abc_SclObjLoad( p, Abc_ObjFanin0(pObj) );
        pLoad.rise += pLoadPo.rise;
        pLoad.fall += pLoadPo.fall;
    }
    // add wire load
    if p.pWLoadUsed != nil {
        if p.vWireCaps == nil {
            p.vWireCaps = Abc_SclFindWireCaps( p.pWLoadUsed, Abc_NtkGetFanoutMax(p.pNtk) ); }
        // inlining of macro Abc_NtkForEachNode1
        for  i = 0; (i < Vec_PtrSize((p.pNtk).vObjs)) && (((pObj) = Abc_NtkObj(p.pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || !Abc_ObjFaninNum(pObj) {} else {
            Abc_SclAddWireLoad( p, pObj, 0 ); }
        }
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(p.pNtk)) && (((pObj) = Abc_NtkPi(p.pNtk, i)), 1); i++  {
            Abc_SclAddWireLoad( p, pObj, 0 );
        }
    }
    // check input loads
    if p.vInDrive != nil {
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(p.pNtk)) && (((pObj) = Abc_NtkPi(p.pNtk, i)), 1); i++  {
            var pLoad *SC_Pair  = Abc_SclObjLoad( p, pObj );
            if Abc_SclObjInDrive(p, pObj) != 0 && (pLoad.rise > Abc_SclObjInDrive(p, pObj) || pLoad.fall > Abc_SclObjInDrive(p, pObj)) {
                fmt.Printf("Maximum input drive strength is exceeded at primary input %d.\n", i); }
        }
    }
/*
    // transfer load from barbufs
    Abc_NtkForEachBarBuf( p->pNtk, pObj, i )
    {
        SC_Pair * pLoad = Abc_SclObjLoad( p, pObj );
        SC_Pair * pLoadF = Abc_SclObjLoad( p, Abc_ObjFanin(pObj, 0) );
        SC_PairAdd( pLoadF, pLoad );
    }
*/
    // calculate average load
//    if ( p->EstLoadMax )
    {
        var TotalLoad float64;
        var nObjs int32;
        // inlining of macro Abc_NtkForEachNode1
        for  i = 0; (i < Vec_PtrSize((p.pNtk).vObjs)) && (((pObj) = Abc_NtkObj(p.pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || !Abc_ObjFaninNum(pObj) {} else {
            var pLoad *SC_Pair  = Abc_SclObjLoad( p, pObj );
            TotalLoad += 0.5 * pLoad.fall + 0.5 * pLoad.rise;
            nObjs++;
        }
        }
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(p.pNtk)) && (((pObj) = Abc_NtkPi(p.pNtk, i)), 1); i++  {
            var pLoad *SC_Pair  = Abc_SclObjLoad( p, pObj );
            TotalLoad += 0.5 * pLoad.fall + 0.5 * pLoad.rise;
            nObjs++;
        }
        p.EstLoadAve = float32((TotalLoad / nObjs));
//        printf( "Average load = %.2f\n", p->EstLoadAve );
    }
}

/**Function*************************************************************

  Synopsis    [Updates load of the node's fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclUpdateLoad(p *SC_Man, pObj *Abc_Obj_t, pOld *SC_Cell, pNew *SC_Cell) {
    var pFanin *Abc_Obj_t;
    var k int32;
    // inlining of macro Abc_ObjForEachFanin
    for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
        var pLoad *SC_Pair  = Abc_SclObjLoad( p, pFanin );
        var pPinOld *SC_Pin  = SC_CellPin( pOld, k );
        var pPinNew *SC_Pin  = SC_CellPin( pNew, k );
        pLoad.rise += pPinNew.rise_cap - pPinOld.rise_cap;
        pLoad.fall += pPinNew.fall_cap - pPinOld.fall_cap;
    }
}
func Abc_SclUpdateLoadSplit(p *SC_Man, pBuffer *Abc_Obj_t, pFanout *Abc_Obj_t) {
    var pPin *SC_Pin;
    var pLoad *SC_Pair;
    var iFanin int32  = Abc_NodeFindFanin( pFanout, pBuffer );
    assert(iFanin >= 0, "iFanin >= 0");
    assert(Abc_ObjFaninNum(pBuffer) == 1, "Abc_ObjFaninNum(pBuffer) == 1");
    pPin = SC_CellPin( Abc_SclObjCell(pFanout), iFanin );
    // update load of the buffer
    pLoad = Abc_SclObjLoad( p, pBuffer );
    pLoad.rise -= pPin.rise_cap;
    pLoad.fall -= pPin.fall_cap;
    // update load of the fanin
    pLoad = Abc_SclObjLoad( p, Abc_ObjFanin0(pBuffer) );
    pLoad.rise += pPin.rise_cap;
    pLoad.fall += pPin.fall_cap;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


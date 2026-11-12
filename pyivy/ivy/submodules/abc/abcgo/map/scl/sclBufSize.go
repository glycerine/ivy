package scl

import (
	"fmt"
	"math"
)

/**CFile****************************************************************

  FileName    [sclBufSize.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Buffering and sizing combined.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclBufSize.c,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sclSize.h"
//#include "map/mio/mio.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Bus_Man_t struct {
    // user data
    pPars    *SC_BusPars;      // parameters    
    pNtk     *Abc_Ntk_t;       // user's network
    pPiDrive       *SC_Cell;   // PI driver
    // library
    pLib        *SC_Lib;       // cell library
    pInv       *SC_Cell;       // base interter (largest/average/???)
    pWLoadUsed   *SC_WireLoad; // name of the used WireLoad model
    vWireCaps     *Vec_Flt_t;  // estimated wire loads
    // internal
    vCins     *Vec_Flt_t;      // input cap for fanouts
    vETimes     *Vec_Flt_t;    // fanout edge departures
    vLoads     *Vec_Flt_t;     // loads for all nodes
    vDepts     *Vec_Flt_t;     // departure times
    vFanouts     *Vec_Ptr_t;   // fanout array
}


func Bus_SclObjMan(p *Abc_Obj_t) *Bus_Man_t { return (Bus_Man_t *)p.pNtk.pBSMan;                                  }
func Bus_SclObjCin(p *Abc_Obj_t) float32 { return Vec_FltEntry( Bus_SclObjMan(p).vCins, Abc_ObjId(p) );         }
func Bus_SclObjSetCin(p *Abc_Obj_t, cap float32) { Vec_FltWriteEntry( Bus_SclObjMan(p).vCins, Abc_ObjId(p), cap );      }
func Bus_SclObjETime(p *Abc_Obj_t) float32 { return Vec_FltEntry( Bus_SclObjMan(p).vETimes, Abc_ObjId(p) );       }
func Bus_SclObjSetETime(p *Abc_Obj_t, time float32) { Vec_FltWriteEntry( Bus_SclObjMan(p).vETimes, Abc_ObjId(p), time );   }
func Bus_SclObjLoad(p *Abc_Obj_t) float32 { return Vec_FltEntry( Bus_SclObjMan(p).vLoads, Abc_ObjId(p) );        }
func Bus_SclObjSetLoad(p *Abc_Obj_t, cap float32) { Vec_FltWriteEntry( Bus_SclObjMan(p).vLoads, Abc_ObjId(p), cap );     }
func Bus_SclObjDept(p *Abc_Obj_t) float32 { return Vec_FltEntry( Bus_SclObjMan(p).vDepts, Abc_ObjId(p) );        }
func Bus_SclObjUpdateDept(p *Abc_Obj_t, time float32) { var q *float32  = Vec_FltEntryP( Bus_SclObjMan(p).vDepts, Abc_ObjId(p) ); if *q < time { *q = time; }  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bus_ManStart(pNtk *Abc_Ntk_t, pLib *SC_Lib, pPars *SC_BusPars) *Bus_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Bus_Man_t{};
    p.pPars     = pPars;
    p.pNtk      = pNtk;
    p.pLib      = pLib;
    p.pInv      = Abc_SclFindInvertor(pLib, pPars.fAddBufs).pRepr.pPrev;//->pAve;
    if pPars.fUseWireLoads { 
        if pNtk.pWLoadUsed == nil {            
            p.pWLoadUsed = Abc_SclFindWireLoadModel( pLib, Abc_SclGetTotalArea(pNtk) );
            if p.pWLoadUsed {
            pNtk.pWLoadUsed = Abc_UtilStrsav( p.pWLoadUsed.pName ); }
        
        } else {
            p.pWLoadUsed = Abc_SclFetchWireLoadModel( pLib, pNtk.pWLoadUsed ); }
    }
    if p.pWLoadUsed {
    p.vWireCaps = Abc_SclFindWireCaps( p.pWLoadUsed, Abc_NtkGetFanoutMax(pNtk) ); }
    p.vFanouts  = Vec_PtrAlloc( 100 );
    p.vCins     = Vec_FltAlloc( 2*Abc_NtkObjNumMax(pNtk) + 1000 );
    p.vETimes   = Vec_FltAlloc( 2*Abc_NtkObjNumMax(pNtk) + 1000 );
    p.vLoads    = Vec_FltAlloc( 2*Abc_NtkObjNumMax(pNtk) + 1000 );
    p.vDepts    = Vec_FltAlloc( 2*Abc_NtkObjNumMax(pNtk) + 1000 );
    Vec_FltFill( p.vCins,   Abc_NtkObjNumMax(pNtk), 0 );
    Vec_FltFill( p.vETimes, Abc_NtkObjNumMax(pNtk), 0 );
    Vec_FltFill( p.vLoads,  Abc_NtkObjNumMax(pNtk), 0 );
    Vec_FltFill( p.vDepts,  Abc_NtkObjNumMax(pNtk), 0 );
    pNtk.pBSMan = p;
    return p;
}
func Bus_ManStop(p *Bus_Man_t) {
    Vec_PtrFreeP( &p.vFanouts );
    Vec_FltFreeP( &p.vWireCaps );
    Vec_FltFreeP( &p.vCins );
    Vec_FltFreeP( &p.vETimes );
    Vec_FltFreeP( &p.vLoads );
    Vec_FltFreeP( &p.vDepts );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bus_ManReadInOutLoads(p *Bus_Man_t) {
    if Abc_FrameReadMaxLoad() {
        var pObj *Abc_Obj_t; var i int32;
        var MaxLoad float32  = Abc_FrameReadMaxLoad();
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
            Bus_SclObjSetCin( pObj, MaxLoad );
        }
//        printf( "Default output load is specified (%f ff).\n", MaxLoad );
    }
    if Abc_FrameReadDrivingCell() {
        var iCell int32  = Abc_SclCellFind( p.pLib, Abc_FrameReadDrivingCell() );
        if iCell == -1 {
            fmt.Printf("Cannot find the default PI driving cell (%s) in the library.\n", Abc_FrameReadDrivingCell());
        } else {
//            printf( "Default PI driving cell is specified (%s).\n", Abc_FrameReadDrivingCell() );
            p.pPiDrive = SC_LibCell( p.pLib, iCell );
            assert(p.pPiDrive != nil, "p.pPiDrive != nil");
            assert(p.pPiDrive.n_inputs == 1, "p.pPiDrive.n_inputs == 1");
//            printf( "Default input driving cell is specified (%s).\n", p->pPiDrive->pName );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Compute load and departure times of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkComputeEdgeDept(pFanout *Abc_Obj_t, iFanin int32, Slew float32) float32 {
    var Load float32  = Bus_SclObjLoad( pFanout );
    var Dept float32  = Bus_SclObjDept( pFanout );
    var Edge float32  = Scl_LibPinArrivalEstimate( Abc_SclObjCell(pFanout), iFanin, Slew, Load );
    assert(Edge > 0, "Edge > 0");
    return Dept + Edge;
}
func Abc_NtkComputeNodeDeparture(pObj *Abc_Obj_t, Slew float32) float32 {
    var pFanout *Abc_Obj_t;
    var i int32;
    assert(Bus_SclObjDept(pObj) == 0, "Bus_SclObjDept(pObj) == 0");
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if Abc_ObjIsBarBuf(pFanout) {
            Bus_SclObjUpdateDept( pObj, Bus_SclObjDept(pFanout) );
        } else if !Abc_ObjIsCo(pFanout) { // add required times here
            Bus_SclObjUpdateDept( pObj, Abc_NtkComputeEdgeDept(pFanout, Abc_NodeFindFanin(pFanout, pObj), Slew) ); }
    }
    return Bus_SclObjDept( pObj );
}
func Abc_NtkComputeFanoutInfo(pObj *Abc_Obj_t, Slew float32) {
    var pFanout *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if Abc_ObjIsBarBuf(pFanout) {
            Bus_SclObjSetETime( pFanout, Bus_SclObjDept(pFanout) );
            Bus_SclObjSetCin( pFanout, Bus_SclObjLoad(pFanout) );
        
        } else if !Abc_ObjIsCo(pFanout) {
            var iFanin int32  = Abc_NodeFindFanin(pFanout, pObj);
            Bus_SclObjSetETime( pFanout, Abc_NtkComputeEdgeDept(pFanout, iFanin, Slew) );
            Bus_SclObjSetCin( pFanout, SC_CellPinCap( Abc_SclObjCell(pFanout), iFanin ) );
        }
    }
}
func Abc_NtkComputeNodeLoad(p *Bus_Man_t, pObj *Abc_Obj_t) float32 {
    var pFanout *Abc_Obj_t;
    var Load float32;
    var i int32;
    assert(Bus_SclObjLoad(pObj) == 0, "Bus_SclObjLoad(pObj) == 0");
    Load = Abc_SclFindWireLoad( p.vWireCaps, Abc_ObjFanoutNum(pObj) );
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        Load += Bus_SclObjCin( pFanout );
    }
    Bus_SclObjSetLoad( pObj, Load );
    return Load;
}
func Abc_NtkComputeFanoutLoad(p *Bus_Man_t, vFanouts *Vec_Ptr_t) float32 {
    var pFanout *Abc_Obj_t;
    var Load float32;
    var i int32;
    Load = Abc_SclFindWireLoad( p.vWireCaps, Vec_PtrSize(vFanouts) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanouts)) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i)), 1); i++  {
        Load += Bus_SclObjCin( pFanout );
    }
    return Load;
}
func Abc_NtkPrintFanoutProfile(pObj *Abc_Obj_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    fmt.Printf("Obj %6d fanouts (%d):\n", Abc_ObjId(pObj), Abc_ObjFanoutNum(pObj));
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        fmt.Printf("%3d : time = %7.2f ps   load = %7.2f ff  ", i, Bus_SclObjETime(pFanout), Bus_SclObjCin(pFanout));
        fmt.Printf("%s\n", Abc_ObjFaninPhase( pFanout, Abc_NodeFindFanin(pFanout, pObj) ) ? "*" : " ");
    }
    fmt.Printf("\n");
}
func Abc_NtkPrintFanoutProfileVec(pObj *Abc_Obj_t, vFanouts *Vec_Ptr_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    fmt.Printf("Fanout profile (%d):\n", Vec_PtrSize(vFanouts));
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanouts)) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i)), 1); i++  {
        fmt.Printf("%3d : time = %7.2f ps   load = %7.2f ff  ", i, Bus_SclObjETime(pFanout), Bus_SclObjCin(pFanout));
        if pObj.pNtk.vPhases {
            fmt.Printf("%s", (pObj && Abc_ObjFanoutNum(pObj) == Vec_PtrSize(vFanouts) && Abc_ObjFaninPhase( pFanout, Abc_NodeFindFanin(pFanout, pObj) )) ? "*" : " "); }
        fmt.Printf("\n");
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Compare two fanouts by their departure times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bus_SclCompareFanouts(pp1 **Abc_Obj_t, pp2 **Abc_Obj_t) int32 {
    var Espilon float32;//10; // 10 ps
    if Bus_SclObjETime(*pp1) < Bus_SclObjETime(*pp2) - Espilon {
        return -1; }
    if Bus_SclObjETime(*pp1) > Bus_SclObjETime(*pp2) + Espilon {
        return 1; }
    if Bus_SclObjCin(*pp1) > Bus_SclObjCin(*pp2) {
        return -1; }
    if Bus_SclObjCin(*pp1) < Bus_SclObjCin(*pp2) {
        return 1; }
    return -1;
}
func Bus_SclInsertFanout(vFanouts *Vec_Ptr_t, pObj *Abc_Obj_t) {
    var pCur *Abc_Obj_t;
    var i, k int32;
    // compact array
    for  i = k = 0; i < Vec_PtrSize(vFanouts); i++  {
        if Vec_PtrEntry(vFanouts, i) != nil {
            Vec_PtrWriteEntry( vFanouts, k++, Vec_PtrEntry(vFanouts, i) ); }
    }
    Vec_PtrShrink( vFanouts, k );
    // insert new entry
    Vec_PtrPush( vFanouts, pObj );
    for  i = Vec_PtrSize(vFanouts) - 1; i > 0; i--  {
        pCur = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i-1);
        pObj = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i);
        if Bus_SclCompareFanouts( &pCur, &pObj ) == -1 {
            break; }
        // inlining of macro ABC_SWAP
        { var t any  = Vec_PtrArray(vFanouts)[i-1]; Vec_PtrArray(vFanouts)[i-1] = Vec_PtrArray(vFanouts)[i]; Vec_PtrArray(vFanouts)[i] = t; };
    }
}
func Bus_SclCheckSortedFanout(vFanouts *Vec_Ptr_t) {
    var pObj, pNext *Abc_Obj_t;
    var i int32;
    for  i = 0; i < Vec_PtrSize(vFanouts) - 1; i++  {
        pObj  = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i);
        pNext = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i+1);
        if Bus_SclCompareFanouts( &pObj, &pNext ) != -1 {
            fmt.Printf("Fanouts %d and %d are out of order.\n", i, i+1);
            Abc_NtkPrintFanoutProfileVec( nil, vFanouts );
            return;
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclOneNodePrint(p *Bus_Man_t, pObj *Abc_Obj_t) {
    var pCell *SC_Cell  = Abc_SclObjCell(pObj);
    fmt.Printf("%s%7d :  ", (Abc_ObjFaninNum(pObj) == 0) ? " Inv" : "Node", Abc_ObjId(pObj));
    fmt.Printf("%d/%2d   ", Abc_ObjFaninNum(pObj) ? Abc_ObjFaninNum(pObj) : 1, Abc_ObjFanoutNum(pObj));
    fmt.Printf("%12s ", pCell.pName);
    fmt.Printf("(%2d/%2d)  ", pCell.Order, pCell.nGates);
    fmt.Printf("gain =%5d  ", int32((100.0 * Bus_SclObjLoad(pObj) / SC_CellPinCapAve(pCell))));
    fmt.Printf("dept =%7.0f ps  ", Bus_SclObjDept(pObj));
    fmt.Printf("\n");
}
func Abc_SclAddOneInv(p *Bus_Man_t, pObj *Abc_Obj_t, vFanouts *Vec_Ptr_t, Gain float32) *Abc_Obj_t {
    var pCellNew *SC_Cell;
    var pFanout, pInv *Abc_Obj_t;
    var Target float32  = SC_CellPinCap(p.pInv, 0) * Gain;
    var LoadWirePrev, LoadWireThis, LoadNew, Load float32;
    var Limit int32  = Abc_MinInt( p.pPars.nDegree, Vec_PtrSize(vFanouts) );
    var i, iStop int32;
    Bus_SclCheckSortedFanout( vFanouts );
    // inlining of macro Vec_PtrForEachEntryStop
    for  iStop = 0; (iStop < Limit) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, iStop)), 1); iStop++  {
        LoadWirePrev = Abc_SclFindWireLoad( p.vWireCaps, iStop );
        LoadWireThis = Abc_SclFindWireLoad( p.vWireCaps, iStop+1 );
        Load += Bus_SclObjCin( pFanout ) - LoadWirePrev + LoadWireThis;
        if Load > Target {
            iStop++;
            break;
        }
    }
    // create inverter
    if p.pPars.fAddBufs {
        pInv = Abc_NtkCreateNodeBuf( p.pNtk, nil );
    } else {
        pInv = Abc_NtkCreateNodeInv( p.pNtk, nil ); }
    assert(int32(Abc_ObjId(pInv)) == Vec_FltSize(p.vCins), "int32(Abc_ObjId(pInv)) == Vec_FltSize(p.vCins)");
    Vec_FltPush( p.vCins,   0 );
    Vec_FltPush( p.vETimes, 0 );
    Vec_FltPush( p.vLoads,  0 );
    Vec_FltPush( p.vDepts,  0 );
    Limit = Abc_MinInt( Abc_MaxInt(iStop, 2), Vec_PtrSize(vFanouts) );
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Limit) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i)), 1); i++  {
        Vec_PtrWriteEntry( vFanouts, i, nil );
        if Abc_ObjFaninNum(pFanout) == 0 {
            Abc_ObjAddFanin( pFanout, pInv );
        } else {
            Abc_ObjPatchFanin( pFanout, pObj, pInv ); }
    }
    // set the gate
    pCellNew = Abc_SclFindSmallestGate( p.pInv, Load / Gain );
    Vec_IntSetEntry( p.pNtk.vGates, Abc_ObjId(pInv), pCellNew.Id );
    // set departure and load
    Abc_NtkComputeNodeDeparture( pInv, p.pPars.Slew );
    LoadNew = Abc_NtkComputeNodeLoad( p, pInv );
    assert(LoadNew - Load < 1 && Load - LoadNew < 1, "LoadNew - Load < 1 && Load - LoadNew < 1");
    // set fanout info for the inverter
    Bus_SclObjSetCin( pInv, SC_CellPinCap(pCellNew, 0) );
    Bus_SclObjSetETime( pInv, Abc_NtkComputeEdgeDept(pInv, 0, p.pPars.Slew) );
    // update phases
    if p.pNtk.vPhases && Abc_SclIsInv(pInv) {
        Abc_NodeInvUpdateFanPolarity( pInv ); }
    return pInv;
}
func Abc_SclBufSize(p *Bus_Man_t, Gain float32) {
    var pCell, pCellNew *SC_Cell;
    var pObj, pFanout *Abc_Obj_t;
    var clk abctime  = Abc_Clock();
    var i int32;
    var k int32;
    var nObjsOld int32 = Abc_NtkObjNumMax(p.pNtk);
    var GainGate, GainInv, Load, LoadNew, Cin, DeptMax float32;
    if p.pPars.fAddBufs {
        GainGate = float32(math.Pow(float64(float64(Gain)), float64(float64(2.0))));
    } else {
        GainGate = Gain;
    }
    if p.pPars.fAddBufs {
        GainInv = float32(math.Pow(float64(float64(Gain)), float64(float64(2.0))));
    } else {
        GainInv = Gain;
    }
    // inlining of macro Abc_NtkForEachObjReverse
    for  i = Vec_PtrSize((p.pNtk).vObjs) - 1; (i >= 0) && (((pObj) = Abc_NtkObj(p.pNtk, i)), 1); i--  {         if (pObj) == nil {} else {
        if !((Abc_ObjIsNode(pObj) && Abc_ObjFaninNum(pObj) > 0) || (Abc_ObjIsCi(pObj) && p.pPiDrive)) {
            continue; }
        if 2 * nObjsOld < Abc_NtkObjNumMax(p.pNtk) {
            fmt.Printf("Buffering could not be completed because the gain value (%d) is too low.\n", p.pPars.GainRatio);
            break;
        }
        // compute load
        Abc_NtkComputeFanoutInfo( pObj, p.pPars.Slew );
        Load = Abc_NtkComputeNodeLoad( p, pObj );
        // consider the gate
        if Abc_ObjIsCi(pObj) || Abc_ObjIsBarBuf(pObj) {
            pCell = p.pPiDrive;
            // if PI driver is not given, assume Cin to be equal to Load
            // this way, buffering of the PIs is performed
            if pCell != nil {
                Cin = SC_CellPinCapAve(pCell);
            } else {
                Cin = Load;
            }
        
        } else {
            pCell = Abc_SclObjCell( pObj );
            Cin = SC_CellPinCapAve( pCell.pAve );
//        Cin = SC_CellPinCapAve( pCell->pRepr->pNext );
        }
        // consider buffering this gate
        if !p.pPars.fSizeOnly && (Abc_ObjFanoutNum(pObj) > p.pPars.nDegree || Load > GainGate * Cin) {
            // add one or more inverters
//            Abc_NtkPrintFanoutProfile( pObj );
            Abc_NodeCollectFanouts( pObj, p.vFanouts );
            Vec_PtrSort( p.vFanouts, (int(*)(void))Bus_SclCompareFanouts );
            for 
            {
                var pInv *Abc_Obj_t;
                if p.pPars.fVeryVerbose {//|| Vec_PtrSize(p->vFanouts) == Abc_ObjFanoutNum(pObj) )
                    Abc_NtkPrintFanoutProfileVec( pObj, p.vFanouts ); }
                pInv = Abc_SclAddOneInv( p, pObj, p.vFanouts, GainInv );
                if p.pPars.fVeryVerbose {
                    Abc_SclOneNodePrint( p, pInv ); }
                Bus_SclInsertFanout( p.vFanouts, pInv );
                Load = Abc_NtkComputeFanoutLoad( p, p.vFanouts );
            
if !( Vec_PtrSize(p.vFanouts) > p.pPars.nDegree || (Vec_PtrSize(p.vFanouts) > 1 && Load > GainGate * Cin) ) { break }
}
            
            // update node fanouts
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(p.vFanouts)) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(p.vFanouts, k)), 1); k++  {
                if Abc_ObjFaninNum(pFanout) == 0 {
                    Abc_ObjAddFanin( pFanout, pObj ); }
            }
            Bus_SclObjSetLoad( pObj, 0 );
            LoadNew = Abc_NtkComputeNodeLoad( p, pObj );
            assert(LoadNew - Load < 1 && Load - LoadNew < 1, "LoadNew - Load < 1 && Load - LoadNew < 1");
        } 
        if Abc_ObjIsCi(pObj) {
            continue; }
        Abc_NtkComputeNodeDeparture( pObj, p.pPars.Slew );
        if Abc_ObjIsBarBuf(pObj) {
            continue; }
        // create cell
        pCellNew = Abc_SclFindSmallestGate( pCell, Load / GainGate );
        Abc_SclObjSetCell( pObj, pCellNew );
        if p.pPars.fVeryVerbose {
            Abc_SclOneNodePrint( p, pObj ); }
        assert(p.pPars.fSizeOnly || Abc_ObjFanoutNum(pObj) <= p.pPars.nDegree, "p.pPars.fSizeOnly || Abc_ObjFanoutNum(pObj) <= p.pPars.nDegree");
    }
    }
    // compute departure time of the PI
    if i < 0 { // finished buffering
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(p.pNtk)) && (((pObj) = Abc_NtkCi(p.pNtk, i)), 1); i++  {
        var DeptCur float32  = Abc_NtkComputeNodeDeparture(pObj, p.pPars.Slew);
        if p.pPiDrive {
            var Load float32  = Bus_SclObjLoad( pObj );
            var ArrOut SC_Pair;
            var SlewOut SC_Pair;
            var LoadIn SC_Pair = SC_Pair{ Load, Load }; 
            Scl_LibHandleInputDriver( p.pPiDrive, &LoadIn, &ArrOut, &SlewOut );
            DeptCur += 0.5 * ArrOut.fall +  0.5 * ArrOut.rise;
        }       
        DeptMax = Abc_MaxFloat( DeptMax, DeptCur );
    } }
    if p.pPars.fVerbose {
        fmt.Printf("WireLoads = %d  Degree = %d  Target slew =%4d ps   Gain2 =%5d  Buf = %6d  Delay =%7.0f ps   ", p.pPars.fUseWireLoads, p.pPars.nDegree, p.pPars.Slew, p.pPars.GainRatio, Abc_NtkObjNumMax(p.pNtk) - nObjsOld, DeptMax);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
}
func Abc_SclBufferingPerform(pNtk *Abc_Ntk_t, pLib *SC_Lib, pPars *SC_BusPars) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var p *Bus_Man_t;
    if !Abc_SclCheckNtk( pNtk, 0 ) {
        return nil; }
    Abc_SclReportDupFanins( pNtk );
    Abc_SclMioGates2SclGates( pLib, pNtk );
    p = Bus_ManStart( pNtk, pLib, pPars );
    Bus_ManReadInOutLoads( p );
    Abc_SclBufSize( p, 0.01 * pPars.GainRatio );
    Bus_ManStop( p );
    Abc_SclSclGates2MioGates( pLib, pNtk );
    if pNtk.vPhases {
        Vec_IntFillExtra( pNtk.vPhases, Abc_NtkObjNumMax(pNtk), 0 ); }
    pNtkNew = Abc_NtkDupDfs( pNtk );
    return pNtkNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


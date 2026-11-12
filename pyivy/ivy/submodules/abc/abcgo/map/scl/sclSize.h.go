package scl

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [sclSize.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Timing/gate-sizing manager.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclSize.h,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "base/abc/abc.h"
//#include "misc/vec/vecQue.h"
//#include "misc/vec/vecWec.h"
//#include "sclLib.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

            
type SC_Man struct {
    pLib        *SC_Lib;          // library
    pNtk     *Abc_Ntk_t;          // network
    nObjs            int32;         // allocated size
    // get assignment
    vGatesBest     *Vec_Int_t;    // best gate sizes found so far
    vUpdates     *Vec_Int_t;      // sizing updates in this round
    vUpdates2     *Vec_Int_t;     // sizing updates in this round
    // timing information
    pWLoadUsed   *SC_WireLoad;    // name of the used WireLoad model
    vWireCaps     *Vec_Flt_t;     // wire capacitances
    pLoads       *SC_Pair;        // loads for each gate
    pDepts       *SC_Pair;        // departures for each gate
    pTimes       *SC_Pair;        // arrivals for each gate
    pSlews       *SC_Pair;        // slews for each gate
    vInDrive     *Vec_Flt_t;      // maximum input drive strength
    vTimesOut     *Vec_Flt_t;     // output arrival times
    vQue     *Vec_Que_t;          // outputs by their time
    pPiDrive       *SC_Cell;      // cell driving primary inputs
    // backup information
    vLoads2     *Vec_Flt_t;       // backup storage for loads
    vLoads3     *Vec_Flt_t;       // backup storage for loads
    vTimes2     *Vec_Flt_t;       // backup storage for times
    vTimes3     *Vec_Flt_t;       // backup storage for slews
    // buffer trees
    EstLoadMax          float32;    // max ratio of Cout/Cin when this kicks in
    EstLoadAve          float32;    // average load of the gate
    EstLinear          float32;     // linear coefficient
    nEstNodes            int32;     // the number of estimations
    // intermediate data
    vNodeByGain     *Vec_Que_t;   // nodes by gain
    vNode2Gain     *Vec_Flt_t;    // mapping node into its gain
    vNode2Gate     *Vec_Int_t;    // mapping node into its best gate
    vNodeIter     *Vec_Int_t;     // the last iteration the node was upsized
    vBestFans     *Vec_Int_t;     // best fanouts
    // incremental timing update
    vLevels     *Vec_Wec_t;
    vChanged     *Vec_Int_t; 
    nIncUpdates            int32;
    // optimization parameters
    SumArea          float32;       // total area
    MaxDelay          float32;      // max delay
    SumArea0          float32;      // total area at the begining 
    MaxDelay0          float32;     // max delay at the begining
    BestDelay          float32;     // best delay in the middle
    ReportDelay          float32;   // delay to report
    // runtime statistics
    timeTotal        abctime;     // starting/total time
    timeCone        abctime;      // critical path selection 
    timeSize        abctime;      // incremental sizing
    timeTime        abctime;      // timing update
    timeOther        abctime;     // everything else
}

////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

func Abc_SclObjLib(p *Abc_Obj_t) *SC_Lib { return (SC_Lib *)p.pNtk.pSCLib;    }
func Abc_SclObjCellId(p *Abc_Obj_t) int32 { return Vec_IntEntry( p.pNtk.vGates, Abc_ObjId(p) );                               }
func Abc_SclObjCell(p *Abc_Obj_t) *SC_Cell { var c int32  = Abc_SclObjCellId(p); if c == -1 {
    return nil;
} else {
    return SC_LibCell(Abc_SclObjLib(p), c);
} }
func Abc_SclObjSetCell(p *Abc_Obj_t, pCell *SC_Cell) { Vec_IntWriteEntry( p.pNtk.vGates, Abc_ObjId(p), pCell.Id );                      }

func Abc_SclObjLoad(p *SC_Man, pObj *Abc_Obj_t) *SC_Pair { return p.pLoads + Abc_ObjId(pObj);  }
func Abc_SclObjDept(p *SC_Man, pObj *Abc_Obj_t) *SC_Pair { return p.pDepts + Abc_ObjId(pObj);  }
func Abc_SclObjTime(p *SC_Man, pObj *Abc_Obj_t) *SC_Pair { return p.pTimes + Abc_ObjId(pObj);  }
func Abc_SclObjSlew(p *SC_Man, pObj *Abc_Obj_t) *SC_Pair { return p.pSlews + Abc_ObjId(pObj);  }

func Abc_SclObjLoadMax(p *SC_Man, pObj *Abc_Obj_t) float64 { return Abc_MaxFloat(Abc_SclObjLoad(p, pObj).rise, Abc_SclObjLoad(p, pObj).fall);  }
func Abc_SclObjLoadAve(p *SC_Man, pObj *Abc_Obj_t) float32 { return 0.5 * Abc_SclObjLoad(p, pObj).rise + 0.5 * Abc_SclObjLoad(p, pObj).fall;   }
func Abc_SclObjTimeOne(p *SC_Man, pObj *Abc_Obj_t, fRise int32) float64 { if fRise != 0 {
    return Abc_SclObjTime(p, pObj).rise;
} else {
    return Abc_SclObjTime(p, pObj).fall;
}       }
func Abc_SclObjTimeMax(p *SC_Man, pObj *Abc_Obj_t) float32 { return Abc_MaxFloat(Abc_SclObjTime(p, pObj).rise, Abc_SclObjTime(p, pObj).fall);  }
func Abc_SclObjSlewMax(p *SC_Man, pObj *Abc_Obj_t) float64 { return Abc_MaxFloat(Abc_SclObjSlew(p, pObj).rise, Abc_SclObjSlew(p, pObj).fall);  }
func Abc_SclObjGetSlackR(p *SC_Man, pObj *Abc_Obj_t, D float32) float32 { return D - (Abc_SclObjTime(p, pObj).rise + Abc_SclObjDept(p, pObj).rise);         }
func Abc_SclObjGetSlackF(p *SC_Man, pObj *Abc_Obj_t, D float32) float32 { return D - (Abc_SclObjTime(p, pObj).fall + Abc_SclObjDept(p, pObj).fall);         }
func Abc_SclObjGetSlack(p *SC_Man, pObj *Abc_Obj_t, D float32) float32 { return D - Abc_MaxFloat(Abc_SclObjTime(p, pObj).rise + Abc_SclObjDept(p, pObj).rise, Abc_SclObjTime(p, pObj).fall + Abc_SclObjDept(p, pObj).fall);  }
func Abc_SclObjSlackMax(p *SC_Man, pObj *Abc_Obj_t, D float32) float64 { return Abc_SclObjGetSlack(p, pObj, D);                                              }
func Abc_SclObjDupFanin(p *SC_Man, pObj *Abc_Obj_t) { assert(Abc_ObjIsCo(pObj), "Abc_ObjIsCo(pObj)"); *Abc_SclObjTime(p, pObj) = *Abc_SclObjTime(p, Abc_ObjFanin0(pObj));  }
func Abc_SclObjInDrive(p *SC_Man, pObj *Abc_Obj_t) float32 { return Vec_FltEntry( p.vInDrive, pObj.iData );                                    }
func Abc_SclObjSetInDrive(p *SC_Man, pObj *Abc_Obj_t, c float32) { Vec_FltWriteEntry( p.vInDrive, pObj.iData, c );                                   }


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Constructor/destructor of STA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclManAlloc(pLib *SC_Lib, pNtk *Abc_Ntk_t) *SC_Man {
    
    var pObj *Abc_Obj_t;
    var i int32;
    assert(pLib.unit_time == 12, "pLib.unit_time == 12");
    assert(pLib.unit_cap_snd == 15, "pLib.unit_cap_snd == 15");
    assert(Abc_NtkHasMapping(pNtk), "Abc_NtkHasMapping(pNtk)");
    // inlining of macro ABC_CALLOC
    p := &SC_Man{};
    p.pLib      = pLib;
    p.pNtk      = pNtk;
    p.nObjs     = Abc_NtkObjNumMax(pNtk);
    // inlining of macro ABC_CALLOC
    p.pLoads    = ((SC_Pair *) calloc((size_t)(p.nObjs), sizeof(SC_Pair)));
    // inlining of macro ABC_CALLOC
    p.pDepts    = ((SC_Pair *) calloc((size_t)(p.nObjs), sizeof(SC_Pair)));
    // inlining of macro ABC_CALLOC
    p.pTimes    = ((SC_Pair *) calloc((size_t)(p.nObjs), sizeof(SC_Pair)));
    // inlining of macro ABC_CALLOC
    p.pSlews    = ((SC_Pair *) calloc((size_t)(p.nObjs), sizeof(SC_Pair)));
    p.vBestFans = Vec_IntStart( p.nObjs );
    p.vTimesOut = Vec_FltStart( Abc_NtkCoNum(pNtk) );
    p.vQue      = Vec_QueAlloc( Abc_NtkCoNum(pNtk) );
    Vec_QueSetPriority( p.vQue, Vec_FltArrayP(p.vTimesOut) );
    for  i = 0; i < Abc_NtkCoNum(pNtk); i++  {
        Vec_QuePush( p.vQue, i );
    }
    p.vUpdates  = Vec_IntAlloc( 1000 );
    p.vUpdates2 = Vec_IntAlloc( 1000 );
    p.vLoads2   = Vec_FltAlloc( 1000 );
    p.vLoads3   = Vec_FltAlloc( 1000 );
    p.vTimes2   = Vec_FltAlloc( 1000 );
    p.vTimes3   = Vec_FltAlloc( 1000 );
    // intermediate data
    p.vNode2Gain  = Vec_FltStart( p.nObjs );
    p.vNode2Gate  = Vec_IntStart( p.nObjs );
    p.vNodeByGain = Vec_QueAlloc( p.nObjs );
    Vec_QueSetPriority( p.vNodeByGain, Vec_FltArrayP(p.vNode2Gain) );
    p.vNodeIter   = Vec_IntStartFull( p.nObjs );
    p.vLevels     = Vec_WecStart( 2 * Abc_NtkLevel(pNtk) + 1 );
    p.vChanged    = Vec_IntAlloc( 100 );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pObj.Level = Abc_ObjFanin0(pObj).Level + 1;
    }
    // set CI/CO ids
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.iData = i;
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pObj.iData = i;
    }
    return p;
}
func Abc_SclManFree(p *SC_Man) {
    var pObj *Abc_Obj_t;
    var i int32;
    // set CI/CO ids
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(p.pNtk)) && (((pObj) = Abc_NtkCi(p.pNtk, i)), 1); i++  {
        pObj.iData = 0;
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        pObj.iData = 0;
    }
    // other
    p.pNtk.pSCLib = nil;
    Vec_IntFreeP( &p.pNtk.vGates );
    Vec_IntFreeP( &p.vNodeIter );
    Vec_QueFreeP( &p.vNodeByGain );
    Vec_FltFreeP( &p.vNode2Gain );
    Vec_IntFreeP( &p.vNode2Gate );
    // intermediate data
    Vec_FltFreeP( &p.vLoads2 );
    Vec_FltFreeP( &p.vLoads3 );
    Vec_FltFreeP( &p.vTimes2 );
    Vec_FltFreeP( &p.vTimes3 );
    Vec_IntFreeP( &p.vUpdates );
    Vec_IntFreeP( &p.vUpdates2 );
    Vec_IntFreeP( &p.vGatesBest );
    Vec_WecFreeP( &p.vLevels );
    Vec_IntFreeP( &p.vChanged );
//    Vec_QuePrint( p->vQue );
    Vec_QueCheck( p.vQue );
    Vec_QueFreeP( &p.vQue );
    Vec_FltFreeP( &p.vTimesOut );
    Vec_IntFreeP( &p.vBestFans );
    Vec_FltFreeP( &p.vInDrive );
    Vec_FltFreeP( &p.vWireCaps );
    p.pLoads = nil;
    p.pDepts = nil;
    p.pTimes = nil;
    p.pSlews = nil;
    p = nil;
}
/*
static inline void Abc_SclManCleanTime( SC_Man * p )
{
    Vec_Flt_t * vSlews;
    Abc_Obj_t * pObj;
    int i;
    vSlews = Vec_FltAlloc( 2 * Abc_NtkPiNum(p->pNtk) );
    Abc_NtkForEachPi( p->pNtk, pObj, i )
    {
        SC_Pair * pSlew = Abc_SclObjSlew( p, pObj );
        Vec_FltPush( vSlews, pSlew->rise );
        Vec_FltPush( vSlews, pSlew->fall );
    }
    memset( p->pDepts, 0, sizeof(SC_Pair) * p->nObjs );
    memset( p->pTimes, 0, sizeof(SC_Pair) * p->nObjs );
    memset( p->pSlews, 0, sizeof(SC_Pair) * p->nObjs );
    Abc_NtkForEachPi( p->pNtk, pObj, i )
    {
        SC_Pair * pSlew = Abc_SclObjSlew( p, pObj );
        pSlew->rise = Vec_FltEntry( vSlews, 2 * i + 0 );
        pSlew->fall = Vec_FltEntry( vSlews, 2 * i + 1 );
    }
    Vec_FltFree( vSlews );
}
*/
func Abc_SclManCleanTime(p *SC_Man) {
    memset( p.pTimes, 0, sizeof(SC_Pair) * p.nObjs );
    memset( p.pSlews, 0, sizeof(SC_Pair) * p.nObjs );
    memset( p.pDepts, 0, sizeof(SC_Pair) * p.nObjs );
/*
    if ( p->pPiDrive != NULL )
    {
        SC_Pair * pSlew, * pTime, * pLoad;
        Abc_Obj_t * pObj;
        int i;
        Abc_NtkForEachPi( p->pNtk, pObj, i )
        {
            pLoad = Abc_SclObjLoad( p, pObj );
            pTime = Abc_SclObjTime( p, pObj );
            pSlew = Abc_SclObjSlew( p, pObj );
            Scl_LibHandleInputDriver( p->pPiDrive, pLoad, pTime, pSlew );
        }
    }
*/
}


/**Function*************************************************************

  Synopsis    [Stores/retrivies information for the logic cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclLoadStore(p *SC_Man, pObj *Abc_Obj_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    Vec_FltClear( p.vLoads2 );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Vec_FltPush( p.vLoads2, Abc_SclObjLoad(p, pFanin).rise );
        Vec_FltPush( p.vLoads2, Abc_SclObjLoad(p, pFanin).fall );
    }
}
func Abc_SclLoadRestore(p *SC_Man, pObj *Abc_Obj_t) {
    var pFanin *Abc_Obj_t;
    var i, k int32;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_SclObjLoad(p, pFanin).rise = Vec_FltEntry(p.vLoads2, k++);
        Abc_SclObjLoad(p, pFanin).fall = Vec_FltEntry(p.vLoads2, k++);
    }
    assert(Vec_FltSize(p.vLoads2) == k, "Vec_FltSize(p.vLoads2) == k");
}

func Abc_SclLoadStore3(p *SC_Man, pObj *Abc_Obj_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    Vec_FltClear( p.vLoads3 );
    Vec_FltPush( p.vLoads3, Abc_SclObjLoad(p, pObj).rise );
    Vec_FltPush( p.vLoads3, Abc_SclObjLoad(p, pObj).fall );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Vec_FltPush( p.vLoads3, Abc_SclObjLoad(p, pFanin).rise );
        Vec_FltPush( p.vLoads3, Abc_SclObjLoad(p, pFanin).fall );
    }
}
func Abc_SclLoadRestore3(p *SC_Man, pObj *Abc_Obj_t) {
    var pFanin *Abc_Obj_t;
    var i, k int32;
    Abc_SclObjLoad(p, pObj).rise = Vec_FltEntry(p.vLoads3, k++);
    Abc_SclObjLoad(p, pObj).fall = Vec_FltEntry(p.vLoads3, k++);
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Abc_SclObjLoad(p, pFanin).rise = Vec_FltEntry(p.vLoads3, k++);
        Abc_SclObjLoad(p, pFanin).fall = Vec_FltEntry(p.vLoads3, k++);
    }
    assert(Vec_FltSize(p.vLoads3) == k, "Vec_FltSize(p.vLoads3) == k");
}
func Abc_SclConeStore(p *SC_Man, vCone *Vec_Int_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    Vec_FltClear( p.vTimes2 );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCone) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vCone,i))), 1); i++  {         if (pObj) == nil {} else {
        Vec_FltPush( p.vTimes2, Abc_SclObjTime(p, pObj).rise );
        Vec_FltPush( p.vTimes2, Abc_SclObjTime(p, pObj).fall );
        Vec_FltPush( p.vTimes2, Abc_SclObjSlew(p, pObj).rise );
        Vec_FltPush( p.vTimes2, Abc_SclObjSlew(p, pObj).fall );
    }
    }
}
func Abc_SclConeRestore(p *SC_Man, vCone *Vec_Int_t) {
    var pObj *Abc_Obj_t;
    var i, k int32;
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCone) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vCone,i))), 1); i++  {         if (pObj) == nil {} else {
        Abc_SclObjTime(p, pObj).rise = Vec_FltEntry(p.vTimes2, k++);
        Abc_SclObjTime(p, pObj).fall = Vec_FltEntry(p.vTimes2, k++);
        Abc_SclObjSlew(p, pObj).rise = Vec_FltEntry(p.vTimes2, k++);
        Abc_SclObjSlew(p, pObj).fall = Vec_FltEntry(p.vTimes2, k++);
    }
    }
    assert(Vec_FltSize(p.vTimes2) == k, "Vec_FltSize(p.vTimes2) == k");
}
func Abc_SclEvalStore(p *SC_Man, vCone *Vec_Int_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    Vec_FltClear( p.vTimes3 );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCone) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vCone,i))), 1); i++  {         if (pObj) == nil {} else {
        Vec_FltPush( p.vTimes3, Abc_SclObjTime(p, pObj).rise );
        Vec_FltPush( p.vTimes3, Abc_SclObjTime(p, pObj).fall );
    }
    }
}
func Abc_SclEvalPerform(p *SC_Man, vCone *Vec_Int_t) float32 {
    var pObj *Abc_Obj_t;
    var Diff float32;
    var Multi float32 = 1.5;
    var Eval float32;
    var i, k int32;
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCone) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vCone,i))), 1); i++  {         if (pObj) == nil {} else {
        Diff  = (Vec_FltEntry(p.vTimes3, k++) - Abc_SclObjTime(p, pObj).rise);
        Diff += (Vec_FltEntry(p.vTimes3, k++) - Abc_SclObjTime(p, pObj).fall);
        Eval += 0.5 * (Diff > 0 ? Diff : Multi * Diff);
    }
    }
    assert(Vec_FltSize(p.vTimes3) == k, "Vec_FltSize(p.vTimes3) == k");
    return Eval / Vec_IntSize(vCone);
}
func Abc_SclEvalPerformLegal(p *SC_Man, vCone *Vec_Int_t, D float32) float32 {
    var pObj *Abc_Obj_t;
    var Rise float32;
    var Fall float32;
    var Multi float32 = 1.0;
    var Eval float32;
    var i, k int32;
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCone) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vCone,i))), 1); i++  {         if (pObj) == nil {} else {
        Rise = Vec_FltEntry(p.vTimes3, k++) - Abc_SclObjTime(p, pObj).rise;
        Fall = Vec_FltEntry(p.vTimes3, k++) - Abc_SclObjTime(p, pObj).fall;
        if Rise + Multi * Abc_SclObjGetSlackR(p, pObj, D) < 0 || Fall + Multi * Abc_SclObjGetSlackF(p, pObj, D) < 0 {
             return -1; }
        Eval += 0.5 * Rise + 0.5 * Fall;
    }
    }
    assert(Vec_FltSize(p.vTimes3) == k, "Vec_FltSize(p.vTimes3) == k");
    return Eval / Vec_IntSize(vCone);
}
func Abc_SclConeClean(p *SC_Man, vCone *Vec_Int_t) {
    var Zero SC_Pair  = SC_Pair{ 0.0, 0.0 };
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCone) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vCone,i))), 1); i++  {         if (pObj) == nil {} else {
        *Abc_SclObjTime(p, pObj) = Zero;
        *Abc_SclObjSlew(p, pObj) = Zero;
    }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclGetBufInvCount(pNtk *Abc_Ntk_t) int32 {
    var pObj *Abc_Obj_t;
    var i, Count int32;
    // inlining of macro Abc_NtkForEachNodeNotBarBuf1
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || !Abc_ObjFaninNum(pObj) || Abc_ObjIsBarBuf(pObj) {} else {
        Count += (Abc_ObjFaninNum(pObj) == 1); }
    }
    return Count;
}
func Abc_SclGetAverageSize(pNtk *Abc_Ntk_t) float32 {
    var pObj *Abc_Obj_t;
    var Total float64;
    var i, Count int32;
    // inlining of macro Abc_NtkForEachNodeNotBarBuf1
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || !Abc_ObjFaninNum(pObj) || Abc_ObjIsBarBuf(pObj) {} else {
        Count++, Total += 100.0*Abc_SclObjCell(pObj).Order/Abc_SclObjCell(pObj).nGates; }
    }
    return float32((Total / Count));
}
func Abc_SclGetTotalArea(pNtk *Abc_Ntk_t) float32 {
    var Area float64;
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachNodeNotBarBuf1
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || !Abc_ObjFaninNum(pObj) || Abc_ObjIsBarBuf(pObj) {} else {
        Area += Abc_SclObjCell(pObj).area; }
    }
    return Area;
}
func Abc_SclGetMaxDelay(p *SC_Man) float32 {
    var fMaxArr float32;
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        fMaxArr = Abc_MaxFloat( fMaxArr, Abc_SclObjTimeMax(p, pObj) );
    }
    return fMaxArr;
}
func Abc_SclGetMaxDelayNodeFanins(p *SC_Man, pNode *Abc_Obj_t) float32 {
    var fMaxArr float32;
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pObj) = Abc_ObjFanin(pNode, i)), 1); i++  {
        fMaxArr = Abc_MaxFloat( fMaxArr, Abc_SclObjTimeMax(p, pObj) );
    }
    return fMaxArr;
}
func Abc_SclReadMaxDelay(p *SC_Man) float32 {
    return Abc_SclObjTimeMax( p, Abc_NtkCo(p.pNtk, Vec_QueTop(p.vQue)) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclObjResiable(p *SC_Man, pObj *Abc_Obj_t, fUpsize int32) *SC_Cell {
    var pOld *SC_Cell  = Abc_SclObjCell(pObj);
    if fUpsize != 0 {
        if pOld.pNext.Order > pOld.Order {
            return pOld.pNext;
        } else {
            return nil;
        }
    } else {
        if pOld.pPrev.Order < pOld.Order {
            return pOld.pPrev;
        } else {
            return nil;
        } }
}

/**Function*************************************************************

  Synopsis    [Dumps timing results into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_SclDumpStats_FileNameOld string // was [1000]byte in C

var spStatic_Abc_SclDumpStats_nNodesOld, spStatic_Abc_SclDumpStats_nAreaOld, spStatic_Abc_SclDumpStats_nDelayOld int32;
var spStatic_Abc_SclDumpStats_clk abctime;
func Abc_SclDumpStats(p *SC_Man, pFileName string, Time abctime) {
    
    
    
    var pTable *os.File;
    pTable = os.OpenFile(pFileName, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    if strcmp( spStatic_Abc_SclDumpStats_FileNameOld, p.pNtk.pName ) {
        spStatic_Abc_SclDumpStats_FileNameOld = fmt.Sprintf("%s", p.pNtk.pName);
        fmt.Fprintf(pTable, "\n");
        fmt.Fprintf(pTable, "%s ", Extra_FileNameWithoutPath(p.pNtk.pName));
        fmt.Fprintf(pTable, "%d ", Abc_NtkPiNum(p.pNtk));
        fmt.Fprintf(pTable, "%d ", Abc_NtkPoNum(p.pNtk));
        fmt.Fprintf(pTable, "%d ", (spStatic_Abc_SclDumpStats_nNodesOld = Abc_NtkNodeNum(p.pNtk)));
        fmt.Fprintf(pTable, "%d ", (spStatic_Abc_SclDumpStats_nAreaOld  = int32(p.SumArea)));
        fmt.Fprintf(pTable, "%d ", (spStatic_Abc_SclDumpStats_nDelayOld = int32(p.ReportDelay)));
        spStatic_Abc_SclDumpStats_clk = Abc_Clock();
    
    } else {
        fmt.Fprintf(pTable, " ");
        fmt.Fprintf(pTable, "%.1f ", 100.0 * Abc_NtkNodeNum(p.pNtk) / spStatic_Abc_SclDumpStats_nNodesOld);
        fmt.Fprintf(pTable, "%.1f ", 100.0 * int32(p.SumArea)         / spStatic_Abc_SclDumpStats_nAreaOld);
        fmt.Fprintf(pTable, "%.1f ", 100.0 * int32(p.ReportDelay)     / spStatic_Abc_SclDumpStats_nDelayOld);
        fmt.Fprintf(pTable, "%.2f", 1.0*(Abc_Clock() - spStatic_Abc_SclDumpStats_clk)/int64(time.Second));
    }
    _ = (pTable).Close();
}

/*=== sclBuffer.c ===============================================================*/
// extern Abc_Ntk_t *   Abc_SclBufferingPerform( Abc_Ntk_t * pNtk, SC_Lib * pLib, SC_BusPars * pPars );

/*=== sclBufferOld.c ===============================================================*/
// extern int           Abc_SclIsInv( Abc_Obj_t * pObj );

// extern void          Abc_NodeInvUpdateFanPolarity( Abc_Obj_t * pObj );

// extern void          Abc_NodeInvUpdateObjFanoutPolarity( Abc_Obj_t * pObj, Abc_Obj_t * pFanout );

// extern void          Abc_SclReportDupFanins( Abc_Ntk_t * pNtk );

// extern Abc_Ntk_t *   Abc_SclUnBufferPerform( Abc_Ntk_t * pNtk, int fVerbose );

// extern Abc_Ntk_t *   Abc_SclUnBufferPhase( Abc_Ntk_t * pNtk, int fVerbose );

// extern Abc_Ntk_t *   Abc_SclBufferPhase( Abc_Ntk_t * pNtk, int fVerbose );

// extern int           Abc_SclCheckNtk( Abc_Ntk_t * p, int fVerbose );

// extern Abc_Ntk_t *   Abc_SclPerformBuffering( Abc_Ntk_t * p, int DegreeR, int Degree, int fUseInvs, int fVerbose );

// extern Abc_Ntk_t *   Abc_SclBufPerform( Abc_Ntk_t * pNtk, int FanMin, int FanMax, int fBufPis, int fSkipDup, int fVerbose );

/*=== sclDnsize.c ===============================================================*/
// extern void          Abc_SclDnsizePerform( SC_Lib * pLib, Abc_Ntk_t * pNtk, SC_SizePars * pPars );

/*=== sclLoad.c ===============================================================*/
// extern Vec_Flt_t *   Abc_SclFindWireCaps( SC_WireLoad * pWL, int nFanoutMax );

// extern float         Abc_SclFindWireLoad( Vec_Flt_t * vWireCaps, int nFans );

// extern void          Abc_SclAddWireLoad( SC_Man * p, Abc_Obj_t * pObj, int fSubtr );

// extern void          Abc_SclComputeLoad( SC_Man * p );

// extern void          Abc_SclUpdateLoad( SC_Man * p, Abc_Obj_t * pObj, SC_Cell * pOld, SC_Cell * pNew );

// extern void          Abc_SclUpdateLoadSplit( SC_Man * p, Abc_Obj_t * pBuffer, Abc_Obj_t * pFanout );

/*=== sclSize.c ===============================================================*/
// extern Abc_Obj_t *   Abc_SclFindCriticalCo( SC_Man * p, int * pfRise );

// extern Abc_Obj_t *   Abc_SclFindMostCriticalFanin( SC_Man * p, int * pfRise, Abc_Obj_t * pNode );

// extern void          Abc_SclTimeNtkPrint( SC_Man * p, int fShowAll, int fPrintPath );

// extern SC_Man *      Abc_SclManStart( SC_Lib * pLib, Abc_Ntk_t * pNtk, int fUseWireLoads, int fDept, float DUser, int nTreeCRatio );

// extern void          Abc_SclTimeCone( SC_Man * p, Vec_Int_t * vCone );

// extern void          Abc_SclTimeNtkRecompute( SC_Man * p, float * pArea, float * pDelay, int fReverse, float DUser );

// extern int           Abc_SclTimeIncUpdate( SC_Man * p );

// extern void          Abc_SclTimeIncInsert( SC_Man * p, Abc_Obj_t * pObj );

// extern void          Abc_SclTimeIncUpdateLevel( Abc_Obj_t * pObj );

// extern void          Abc_SclTimePerform( SC_Lib * pLib, Abc_Ntk_t * pNtk, int nTreeCRatio, int fUseWireLoads, int fShowAll, int fPrintPath, int fDumpStats );

// extern void          Abc_SclPrintBuffers( SC_Lib * pLib, Abc_Ntk_t * pNtk, int fVerbose );

/*=== sclUpsize.c ===============================================================*/
// extern int           Abc_SclCountNearCriticalNodes( SC_Man * p );

// extern void          Abc_SclUpsizePerform( SC_Lib * pLib, Abc_Ntk_t * pNtk, SC_SizePars * pPars );

/*=== sclUtil.c ===============================================================*/
// extern void          Abc_SclMioGates2SclGates( SC_Lib * pLib, Abc_Ntk_t * p );

// extern void          Abc_SclSclGates2MioGates( SC_Lib * pLib, Abc_Ntk_t * p );

// extern void          Abc_SclTransferGates( Abc_Ntk_t * pOld, Abc_Ntk_t * pNew );

// extern void          Abc_SclPrintGateSizes( SC_Lib * pLib, Abc_Ntk_t * p );

// extern void          Abc_SclMinsizePerform( SC_Lib * pLib, Abc_Ntk_t * p, int fUseMax, int fVerbose );

// extern int           Abc_SclCountMinSize( SC_Lib * pLib, Abc_Ntk_t * p, int fUseMax );

// extern Vec_Int_t *   Abc_SclExtractBarBufs( Abc_Ntk_t * pNtk );

// extern void          Abc_SclInsertBarBufs( Abc_Ntk_t * pNtk, Vec_Int_t * vBufs );



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

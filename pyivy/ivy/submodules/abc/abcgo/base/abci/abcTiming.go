package abci

import (
	"fmt"
	"math"
)

/**CFile****************************************************************

  FileName    [abcTiming.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Computation of timing info for mapped circuits.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcTiming.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <math.h>

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Abc_ManTime_t_ struct {
    tArrDef     Abc_Time_t;
    tReqDef     Abc_Time_t;
    vArrs     *Vec_Ptr_t;
    vReqs     *Vec_Ptr_t;
    tInDriveDef     Abc_Time_t;
    tOutLoadDef     Abc_Time_t;
    tInDrive    *Abc_Time_t;
    tOutLoad    *Abc_Time_t;
}

const TOLERANCE = 0.001

func Abc_FloatEqual(x float32, y float32) int32 { return math.Abs(float64(x-y)) < TOLERANCE; }

// static functions
// static Abc_ManTime_t *     Abc_ManTimeStart( Abc_Ntk_t * pNtk );

// static void                Abc_ManTimeExpand( Abc_ManTime_t * p, int nSize, int fProgressive );


// accessing the arrival and required times of a node
func Abc_NodeArrival(pNode *Abc_Obj_t) *Abc_Time_t {  return (Abc_Time_t *)pNode.pNtk.pManTime.vArrs.pArray[pNode.Id];  }
func Abc_NodeRequired(pNode *Abc_Obj_t) *Abc_Time_t {  return (Abc_Time_t *)pNode.pNtk.pManTime.vReqs.pArray[pNode.Id];  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reads the arrival.required time of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkReadDefaultArrival(pNtk *Abc_Ntk_t) *Abc_Time_t {
    assert(pNtk.pManTime, "pNtk.pManTime");
    return &pNtk.pManTime.tArrDef;
}
func Abc_NtkReadDefaultRequired(pNtk *Abc_Ntk_t) *Abc_Time_t {
    assert(pNtk.pManTime, "pNtk.pManTime");
    return &pNtk.pManTime.tReqDef;
}
func Abc_NodeReadArrival(pNode *Abc_Obj_t) *Abc_Time_t {
    assert(pNode.pNtk.pManTime, "pNode.pNtk.pManTime");
    return Abc_NodeArrival(pNode);
}
func Abc_NodeReadRequired(pNode *Abc_Obj_t) *Abc_Time_t {
    assert(pNode.pNtk.pManTime, "pNode.pNtk.pManTime");
    return Abc_NodeRequired(pNode);
}
func Abc_NtkReadDefaultArrivalWorst(pNtk *Abc_Ntk_t) float32 {
    return 0.5 * pNtk.pManTime.tArrDef.Rise + 0.5 * pNtk.pManTime.tArrDef.Fall;
}
func Abc_NtkReadDefaultRequiredWorst(pNtk *Abc_Ntk_t) float32 {
    return 0.5 * pNtk.pManTime.tReqDef.Rise + 0.5 * pNtk.pManTime.tReqDef.Fall;
}
func Abc_NodeReadArrivalAve(pNode *Abc_Obj_t) float32 {
    return 0.5 * Abc_NodeArrival(pNode).Rise + 0.5 * Abc_NodeArrival(pNode).Fall;
}
func Abc_NodeReadRequiredAve(pNode *Abc_Obj_t) float32 {
    return 0.5 * Abc_NodeReadRequired(pNode).Rise + 0.5 * Abc_NodeReadRequired(pNode).Fall;
}
func Abc_NodeReadArrivalWorst(pNode *Abc_Obj_t) float32 {
    return Abc_MaxFloat( Abc_NodeArrival(pNode).Rise, Abc_NodeArrival(pNode).Fall );
}
func Abc_NodeReadRequiredWorst(pNode *Abc_Obj_t) float32 {
    return Abc_MinFloat( Abc_NodeReadRequired(pNode).Rise, Abc_NodeReadRequired(pNode).Fall );
}

/**Function*************************************************************

  Synopsis    [Reads the input drive / output load of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkReadDefaultInputDrive(pNtk *Abc_Ntk_t) *Abc_Time_t {
    assert(pNtk.pManTime, "pNtk.pManTime");
    return &pNtk.pManTime.tInDriveDef;
}
func Abc_NtkReadDefaultOutputLoad(pNtk *Abc_Ntk_t) *Abc_Time_t {
    assert(pNtk.pManTime, "pNtk.pManTime");
    return &pNtk.pManTime.tOutLoadDef;
}
func Abc_NodeReadInputDrive(pNtk *Abc_Ntk_t, iPi int32) *Abc_Time_t {
    assert(pNtk.pManTime, "pNtk.pManTime");
    if pNtk.pManTime.tInDrive {
        return pNtk.pManTime.tInDrive + iPi;
    } else {
        return nil;
    }
}
func Abc_NodeReadOutputLoad(pNtk *Abc_Ntk_t, iPo int32) *Abc_Time_t {
    assert(pNtk.pManTime, "pNtk.pManTime");
    if pNtk.pManTime.tOutLoad {
        return pNtk.pManTime.tOutLoad + iPo;
    } else {
        return nil;
    }
}
func Abc_NodeReadInputDriveWorst(pNtk *Abc_Ntk_t, iPi int32) float32 {
    return Abc_MaxFloat( Abc_NodeReadInputDrive(pNtk, iPi).Rise, Abc_NodeReadInputDrive(pNtk, iPi).Fall );
}
func Abc_NodeReadOutputLoadWorst(pNtk *Abc_Ntk_t, iPo int32) float32 {
    return Abc_MaxFloat( Abc_NodeReadOutputLoad(pNtk, iPo).Rise, Abc_NodeReadOutputLoad(pNtk, iPo).Fall );
}

/**Function*************************************************************

  Synopsis    [Sets the default arrival time for the network.]

  Description [Please note that .default_input_arrival and
  .default_output_required should precede .input_arrival and 
  .output required. Otherwise, an overwrite may happen.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTimeSetDefaultArrival(pNtk *Abc_Ntk_t, Rise float32, Fall float32) {
    var pObj *Abc_Obj_t; var i int32;
    if pNtk.pManTime == nil {
        pNtk.pManTime = Abc_ManTimeStart(pNtk); }
    pNtk.pManTime.tArrDef.Rise  = Rise;
    pNtk.pManTime.tArrDef.Fall  = Fall;
    // set the arrival times for each input
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Abc_NtkTimeSetArrival( pNtk, Abc_ObjId(pObj), Rise, Fall );
    }    
}
func Abc_NtkTimeSetDefaultRequired(pNtk *Abc_Ntk_t, Rise float32, Fall float32) {
    var pObj *Abc_Obj_t; var i int32;
    if pNtk.pManTime == nil {
        pNtk.pManTime = Abc_ManTimeStart(pNtk); }
    pNtk.pManTime.tReqDef.Rise  = Rise;
    pNtk.pManTime.tReqDef.Fall  = Fall;
    // set the required times for each output
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_NtkTimeSetRequired( pNtk, Abc_ObjId(pObj), Rise, Fall );
    }        
}

/**Function*************************************************************

  Synopsis    [Sets the arrival time for an object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTimeSetArrival(pNtk *Abc_Ntk_t, ObjId int32, Rise float32, Fall float32) {
    var vTimes *Vec_Ptr_t;
    var pTime *Abc_Time_t;
    if pNtk.pManTime == nil {
        pNtk.pManTime = Abc_ManTimeStart(pNtk); }
    Abc_ManTimeExpand( pNtk.pManTime, ObjId + 1, 1 );
    // set the arrival time
    vTimes = pNtk.pManTime.vArrs;
    pTime = (Abc_Time_t *)vTimes.pArray[ObjId];
    pTime.Rise  = Rise;
    pTime.Fall  = Fall;
}
func Abc_NtkTimeSetRequired(pNtk *Abc_Ntk_t, ObjId int32, Rise float32, Fall float32) {
    var vTimes *Vec_Ptr_t;
    var pTime *Abc_Time_t;
    if pNtk.pManTime == nil {
        pNtk.pManTime = Abc_ManTimeStart(pNtk); }
    Abc_ManTimeExpand( pNtk.pManTime, ObjId + 1, 1 );
    // set the required time
    vTimes = pNtk.pManTime.vReqs;
    pTime = (Abc_Time_t *)vTimes.pArray[ObjId];
    pTime.Rise  = Rise;
    pTime.Fall  = Fall;
}

/**Function*************************************************************

  Synopsis    [Sets the default arrival time for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTimeSetDefaultInputDrive(pNtk *Abc_Ntk_t, Rise float32, Fall float32) {
//    if ( Rise == 0.0 && Fall == 0.0 )
//        return;
    if pNtk.pManTime == nil {
        pNtk.pManTime = Abc_ManTimeStart(pNtk); }
    pNtk.pManTime.tInDriveDef.Rise  = Rise;
    pNtk.pManTime.tInDriveDef.Fall  = Fall;
    if pNtk.pManTime.tInDrive != nil {
        var i int32;
        for  i = 0; i < Abc_NtkCiNum(pNtk); i++  {
            if pNtk.pManTime.tInDrive[i].Rise == 0 && pNtk.pManTime.tInDrive[i].Fall == 0 {
                pNtk.pManTime.tInDrive[i] = pNtk.pManTime.tInDriveDef; }
        }
    }
}
func Abc_NtkTimeSetDefaultOutputLoad(pNtk *Abc_Ntk_t, Rise float32, Fall float32) {
//    if ( Rise == 0.0 && Fall == 0.0 )
//        return;
    if pNtk.pManTime == nil {
        pNtk.pManTime = Abc_ManTimeStart(pNtk); }
    pNtk.pManTime.tOutLoadDef.Rise  = Rise;
    pNtk.pManTime.tOutLoadDef.Fall  = Fall;
    if pNtk.pManTime.tOutLoad != nil {
        var i int32;
        for  i = 0; i < Abc_NtkCoNum(pNtk); i++  {
            if pNtk.pManTime.tOutLoad[i].Rise == 0 && pNtk.pManTime.tOutLoad[i].Fall == 0 {
                pNtk.pManTime.tOutLoad[i] = pNtk.pManTime.tOutLoadDef; }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Sets the arrival time for an object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTimeSetInputDrive(pNtk *Abc_Ntk_t, PiNum int32, Rise float32, Fall float32) {
    var pTime *Abc_Time_t;
    assert(PiNum >= 0 && PiNum < Abc_NtkCiNum(pNtk), "PiNum >= 0 && PiNum < Abc_NtkCiNum(pNtk)");
    if pNtk.pManTime == nil {
        pNtk.pManTime = Abc_ManTimeStart(pNtk); }
    if pNtk.pManTime.tInDriveDef.Rise == Rise && pNtk.pManTime.tInDriveDef.Fall == Fall {
        return; }
    if pNtk.pManTime.tInDrive == nil {
        var i int32;
        // inlining of macro ABC_CALLOC
        pNtk.pManTime.tInDrive = ((Abc_Time_t *) calloc((size_t)(Abc_NtkCiNum(pNtk)), sizeof(Abc_Time_t)));
        for  i = 0; i < Abc_NtkCiNum(pNtk); i++  {
            pNtk.pManTime.tInDrive[i] = pNtk.pManTime.tInDriveDef;
        }
    }
    pTime = pNtk.pManTime.tInDrive + PiNum;
    pTime.Rise  = Rise;
    pTime.Fall  = Fall;
}
func Abc_NtkTimeSetOutputLoad(pNtk *Abc_Ntk_t, PoNum int32, Rise float32, Fall float32) {
    var pTime *Abc_Time_t;
    assert(PoNum >= 0 && PoNum < Abc_NtkCoNum(pNtk), "PoNum >= 0 && PoNum < Abc_NtkCoNum(pNtk)");
    if pNtk.pManTime == nil {
        pNtk.pManTime = Abc_ManTimeStart(pNtk); }
    if pNtk.pManTime.tOutLoadDef.Rise == Rise && pNtk.pManTime.tOutLoadDef.Fall == Fall {
        return; }
    if pNtk.pManTime.tOutLoad == nil {
        var i int32;
        // inlining of macro ABC_CALLOC
        pNtk.pManTime.tOutLoad = ((Abc_Time_t *) calloc((size_t)(Abc_NtkCoNum(pNtk)), sizeof(Abc_Time_t)));
        for  i = 0; i < Abc_NtkCoNum(pNtk); i++  {
            pNtk.pManTime.tOutLoad[i] = pNtk.pManTime.tOutLoadDef;
        }
    }
    pTime = pNtk.pManTime.tOutLoad + PoNum;
    pTime.Rise  = Rise;
    pTime.Fall  = Fall;
}

/**Function*************************************************************

  Synopsis    [Finalizes the timing manager after setting arr/req times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTimeInitialize(pNtk *Abc_Ntk_t, pNtkOld *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var ppTimes **Abc_Time_t;
    var i int32;
    assert(pNtkOld == nil || pNtkOld.pManTime != nil, "pNtkOld == nil || pNtkOld.pManTime != nil");
    assert(pNtkOld == nil || Abc_NtkCiNum(pNtk) == Abc_NtkCiNum(pNtkOld), "pNtkOld == nil || Abc_NtkCiNum(pNtk) == Abc_NtkCiNum(pNtkOld)");
    assert(pNtkOld == nil || Abc_NtkCoNum(pNtk) == Abc_NtkCoNum(pNtkOld), "pNtkOld == nil || Abc_NtkCoNum(pNtk) == Abc_NtkCoNum(pNtkOld)");
    if pNtk.pManTime == nil {
        return; }
    // create timing manager with default values
    Abc_ManTimeExpand( pNtk.pManTime, Abc_NtkObjNumMax(pNtk), 0 );
    // set global defaults from pNtkOld
    if pNtkOld != nil {
        pNtk.pManTime.tArrDef = pNtkOld.pManTime.tArrDef;
        pNtk.pManTime.tReqDef = pNtkOld.pManTime.tReqDef;
        pNtk.AndGateDelay = pNtkOld.AndGateDelay;
    }
    // set the default timing for CI and COs
    ppTimes = (Abc_Time_t **)pNtk.pManTime.vArrs.pArray;
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        *ppTimes[pObj.Id] = pNtkOld ? *Abc_NodeReadArrival(Abc_NtkCi(pNtkOld, i)) : pNtk.pManTime.tArrDef;
    }
    ppTimes = (Abc_Time_t **)pNtk.pManTime.vReqs.pArray;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        *ppTimes[pObj.Id] = pNtkOld ? *Abc_NodeReadRequired(Abc_NtkCo(pNtkOld, i)) : pNtk.pManTime.tReqDef;
    }
}

/**Function*************************************************************

  Synopsis    [This procedure scales user timing by multiplicative factor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTimeScale(pNtk *Abc_Ntk_t, Scale float32) {
    var pObj *Abc_Obj_t;
    var ppTimes **Abc_Time_t;
    var pTime *Abc_Time_t;
    var i int32;
    if pNtk.pManTime == nil {
        return; }
    // arrival
    pNtk.pManTime.tArrDef.Fall *= Scale;
    pNtk.pManTime.tArrDef.Rise *= Scale;
    // departure
    pNtk.pManTime.tReqDef.Fall *= Scale;
    pNtk.pManTime.tReqDef.Rise *= Scale;
    // set the default timing
    ppTimes = (Abc_Time_t **)pNtk.pManTime.vArrs.pArray;
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pTime = ppTimes[pObj.Id];
        pTime.Fall *= Scale;
        pTime.Rise *= Scale;
    }
    // set the default timing
    ppTimes = (Abc_Time_t **)pNtk.pManTime.vReqs.pArray;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pTime = ppTimes[pObj.Id];
        pTime.Fall *= Scale;
        pTime.Rise *= Scale;
    }
}

/**Function*************************************************************

  Synopsis    [Prepares the timing manager for delay trace.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTimePrepare(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var ppTimes **Abc_Time_t;
    var pTime *Abc_Time_t;
    var i int32;
    // if there is no timing manager, allocate and initialize
    if pNtk.pManTime == nil {
        pNtk.pManTime = Abc_ManTimeStart(pNtk);
        Abc_NtkTimeInitialize( pNtk, nil );
        return;
    }
    // if timing manager is given, expand it if necessary
    Abc_ManTimeExpand( pNtk.pManTime, Abc_NtkObjNumMax(pNtk), 0 );
    // clean arrivals except for PIs
    ppTimes = (Abc_Time_t **)pNtk.pManTime.vArrs.pArray;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        pTime = ppTimes[pObj.Id];
        if pTime.Rise = Abc_ObjFaninNum(pObj) {
            pTime.Fall = -ABC_INFINITY;
        } else {
            pTime.Fall = 0;
        } // set contant node arrivals to zero
    }
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pTime = ppTimes[pObj.Id];
        pTime.Fall = pTime.Rise = -ABC_INFINITY;
    }
    // clean required except for POs
    ppTimes = (Abc_Time_t **)pNtk.pManTime.vReqs.pArray;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        pTime = ppTimes[pObj.Id];
        pTime.Fall = pTime.Rise = ABC_INFINITY;
    }
    }
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pTime = ppTimes[pObj.Id];
        pTime.Fall = pTime.Rise = ABC_INFINITY;
    }
}




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManTimeStart(pNtk *Abc_Ntk_t) *Abc_ManTime_t {
    var fUseZeroDefaultOutputRequired int32  = 1;
    var p *Abc_ManTime_t;
    var pObj *Abc_Obj_t; var i int32;
    // inlining of macro ABC_ALLOC
    p = pNtk.pManTime = ((Abc_ManTime_t *) malloc(sizeof(Abc_ManTime_t) * (size_t)(1)));
    memset( p, 0, sizeof(Abc_ManTime_t) );
    p.vArrs = Vec_PtrAlloc( 0 );
    p.vReqs = Vec_PtrAlloc( 0 );
    // set default default input=arrivals (assumed to be 0)
    // set default default output-requireds (can be either 0 or +infinity, based on the flag)
    if fUseZeroDefaultOutputRequired != 0 {
        p.tReqDef.Rise = 0;
    } else {
        p.tReqDef.Rise = ABC_INFINITY;
    }
    if fUseZeroDefaultOutputRequired != 0 {
        p.tReqDef.Fall = 0;
    } else {
        p.tReqDef.Fall = ABC_INFINITY;
    }
    // extend manager
    Abc_ManTimeExpand( p, Abc_NtkObjNumMax(pNtk) + 1, 0 );
    // set the default timing for CIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Abc_NtkTimeSetArrival( pNtk, Abc_ObjId(pObj), p.tArrDef.Rise, p.tArrDef.Rise );
    }   
    // set the default timing for COs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_NtkTimeSetRequired( pNtk, Abc_ObjId(pObj), p.tReqDef.Rise, p.tReqDef.Rise );
    }        
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManTimeStop(p *Abc_ManTime_t) {
    if p.tInDrive {
        p.tInDrive = nil; }
    if p.tOutLoad {
        p.tOutLoad = nil; }
    if Vec_PtrSize(p.vArrs) > 0 {
        p.vArrs.pArray[0] = nil; }
    Vec_PtrFree( p.vArrs );
    if Vec_PtrSize(p.vReqs) > 0 {
        p.vReqs.pArray[0] = nil; }
    Vec_PtrFree( p.vReqs );
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Duplicates the timing manager with the PI/PO timing info.]

  Description [The PIs/POs of the new network should be allocated.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManTimeDup(pNtkOld *Abc_Ntk_t, pNtkNew *Abc_Ntk_t) {
    extern void Abc_NtkTimePrint( Abc_Ntk_t * pNtk );

    var pObj *Abc_Obj_t;
    var ppTimesOld, ppTimesNew **Abc_Time_t;
    var i int32;
    if pNtkOld.pManTime == nil {
        return; }
    assert(Abc_NtkCiNum(pNtkOld) == Abc_NtkCiNum(pNtkNew), "Abc_NtkCiNum(pNtkOld) == Abc_NtkCiNum(pNtkNew)");
    assert(Abc_NtkCoNum(pNtkOld) == Abc_NtkCoNum(pNtkNew), "Abc_NtkCoNum(pNtkOld) == Abc_NtkCoNum(pNtkNew)");
    assert(Abc_NtkLatchNum(pNtkOld) == Abc_NtkLatchNum(pNtkNew), "Abc_NtkLatchNum(pNtkOld) == Abc_NtkLatchNum(pNtkNew)");
    // create the new timing manager
    pNtkNew.pManTime = Abc_ManTimeStart(pNtkNew);
    Abc_ManTimeExpand( pNtkNew.pManTime, Abc_NtkObjNumMax(pNtkNew), 0 );
    // set the default timing
    pNtkNew.pManTime.tArrDef = pNtkOld.pManTime.tArrDef;
    pNtkNew.pManTime.tReqDef = pNtkOld.pManTime.tReqDef;
    // set the CI timing
    ppTimesOld = (Abc_Time_t **)pNtkOld.pManTime.vArrs.pArray;
    ppTimesNew = (Abc_Time_t **)pNtkNew.pManTime.vArrs.pArray;
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtkOld)) && (((pObj) = Abc_NtkCi(pNtkOld, i)), 1); i++  {
        *ppTimesNew[ Abc_NtkCi(pNtkNew,i).Id ] = *ppTimesOld[ pObj.Id ];
    }
    // set the CO timing
    ppTimesOld = (Abc_Time_t **)pNtkOld.pManTime.vReqs.pArray;
    ppTimesNew = (Abc_Time_t **)pNtkNew.pManTime.vReqs.pArray;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtkOld)) && (((pObj) = Abc_NtkCo(pNtkOld, i)), 1); i++  {
        *ppTimesNew[ Abc_NtkCo(pNtkNew,i).Id ] = *ppTimesOld[ pObj.Id ];
    }
    // duplicate input drive
    pNtkNew.pManTime.tInDriveDef = pNtkOld.pManTime.tInDriveDef; 
    pNtkNew.pManTime.tOutLoadDef = pNtkOld.pManTime.tOutLoadDef; 
    if pNtkOld.pManTime.tInDrive {
        // inlining of macro ABC_ALLOC
        pNtkNew.pManTime.tInDrive = ((Abc_Time_t *) malloc(sizeof(Abc_Time_t) * (size_t)(Abc_NtkCiNum(pNtkOld))));
        memcpy( pNtkNew.pManTime.tInDrive, pNtkOld.pManTime.tInDrive, sizeof(Abc_Time_t) * Abc_NtkCiNum(pNtkOld) );
    }
    if pNtkOld.pManTime.tOutLoad {
        // inlining of macro ABC_ALLOC
        pNtkNew.pManTime.tOutLoad = ((Abc_Time_t *) malloc(sizeof(Abc_Time_t) * (size_t)(Abc_NtkCiNum(pNtkOld))));
        memcpy( pNtkNew.pManTime.tOutLoad, pNtkOld.pManTime.tOutLoad, sizeof(Abc_Time_t) * Abc_NtkCoNum(pNtkOld) );
    }
}

/**Function*************************************************************

  Synopsis    [Prints out the timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTimePrint(pNtk *Abc_Ntk_t) {
    if pNtk.pManTime == nil {
        fmt.Printf("There is no timing manager\n");
    } else {
        var pObj *Abc_Obj_t; var i int32;
        fmt.Printf("Default arrival = %8f\n", pNtk.pManTime.tArrDef.Fall);
        fmt.Printf("Default required = %8f\n", pNtk.pManTime.tReqDef.Fall);
        fmt.Printf("Inputs (%d):\n", Abc_NtkCiNum(pNtk));
        // inlining of macro Abc_NtkForEachCi
        for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
            fmt.Printf("%20s   arrival = %8f   required = %8f\n", Abc_ObjName(pObj), Abc_NodeReadArrivalWorst(pObj), Abc_NodeReadRequiredWorst(pObj));
        }
        fmt.Printf("Outputs (%d):\n", Abc_NtkCoNum(pNtk));
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
            fmt.Printf("%20s   arrival = %8f   required = %8f\n", Abc_ObjName(pObj), Abc_NodeReadArrivalWorst(pObj), Abc_NodeReadRequiredWorst(pObj));
        }
    }
}

/**Function*************************************************************

  Synopsis    [Expends the storage for timing information.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManTimeExpand(p *Abc_ManTime_t, nSize int32, fProgressive int32) {
    var vTimes *Vec_Ptr_t;
    var ppTimes, ppTimesOld, pTime *Abc_Time_t;
    var nSizeOld, nSizeNew, i int32;

    nSizeOld = p.vArrs.nSize;
    if nSizeOld >= nSize {
        return; }
    if fProgressive != 0 {
        nSizeNew = 2 * nSize;
    } else {
        nSizeNew = nSize;
    }
    if nSizeNew < 100 {
        nSizeNew = 100; }

    vTimes = p.vArrs;
    Vec_PtrGrow( vTimes, nSizeNew );
    vTimes.nSize = nSizeNew;
    if nSizeOld == 0 {
        ppTimesOld = nil;
    } else {
        ppTimesOld = (Abc_Time_t *)vTimes.pArray[0];
    }
    // inlining of macro ABC_REALLOC
    
    if ppTimesOld {
        ppTimes = ((Abc_Time_t *) realloc((char *)(ppTimesOld), sizeof(Abc_Time_t) * (size_t)(nSizeNew)));
    } else {
        ppTimes = ((Abc_Time_t *) malloc(sizeof(Abc_Time_t) * (size_t)(nSizeNew)));
    }
    for  i = 0; i < nSizeNew; i++  {
        vTimes.pArray[i] = ppTimes + i;
    }
    for  i = nSizeOld; i < nSizeNew; i++  {
        pTime = (Abc_Time_t *)vTimes.pArray[i];
        pTime.Rise  = -ABC_INFINITY;
        pTime.Fall  = -ABC_INFINITY;
    }

    vTimes = p.vReqs;
    Vec_PtrGrow( vTimes, nSizeNew );
    vTimes.nSize = nSizeNew;
    if nSizeOld == 0 {
        ppTimesOld = nil;
    } else {
        ppTimesOld = (Abc_Time_t *)vTimes.pArray[0];
    }
    // inlining of macro ABC_REALLOC
    
    if ppTimesOld {
        ppTimes = ((Abc_Time_t *) realloc((char *)(ppTimesOld), sizeof(Abc_Time_t) * (size_t)(nSizeNew)));
    } else {
        ppTimes = ((Abc_Time_t *) malloc(sizeof(Abc_Time_t) * (size_t)(nSizeNew)));
    }
    for  i = 0; i < nSizeNew; i++  {
        vTimes.pArray[i] = ppTimes + i;
    }
    for  i = nSizeOld; i < nSizeNew; i++  {
        pTime = (Abc_Time_t *)vTimes.pArray[i];
        pTime.Rise  = ABC_INFINITY;
        pTime.Fall  = ABC_INFINITY;
    }
}






/**Function*************************************************************

  Synopsis    [Sets the CI node levels according to the arrival info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSetNodeLevelsArrival(pNtkOld *Abc_Ntk_t) {
    var pNodeOld, pNodeNew *Abc_Obj_t;
    var tAndDelay float32;
    var i int32;
    if pNtkOld.pManTime == nil {
        return; }
    if Abc_FrameReadLibGen() == nil || Mio_LibraryReadNand2((Mio_Library_t *)Abc_FrameReadLibGen()) == nil {
        return; }
    tAndDelay = Mio_LibraryReadDelayNand2Max((Mio_Library_t *)Abc_FrameReadLibGen());
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtkOld)) && (((pNodeOld) = Abc_NtkCi(pNtkOld, i)), 1); i++  {
        pNodeNew = pNodeOld.pCopy;
        pNodeNew.Level = int32((Abc_NodeReadArrivalWorst(pNodeOld) / tAndDelay));
    }
}

/**Function*************************************************************

  Synopsis    [Sets the CI node levels according to the arrival info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkGetCiArrivalTimes(pNtk *Abc_Ntk_t) *Abc_Time_t {
    var p *Abc_Time_t;
    var pNode *Abc_Obj_t;
    var i int32;
    // inlining of macro ABC_CALLOC
    p = ((Abc_Time_t *) calloc((size_t)(Abc_NtkCiNum(pNtk)), sizeof(Abc_Time_t)));
    if pNtk.pManTime == nil {
        return p; }
    // set the PI arrival times
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        p[i] = *Abc_NodeArrival(pNode);
    }
    return p;
}
func Abc_NtkGetCoRequiredTimes(pNtk *Abc_Ntk_t) *Abc_Time_t {
    var p *Abc_Time_t;
    var pNode *Abc_Obj_t;
    var i int32;
    // inlining of macro ABC_CALLOC
    p = ((Abc_Time_t *) calloc((size_t)(Abc_NtkCoNum(pNtk)), sizeof(Abc_Time_t)));
    if pNtk.pManTime == nil {
        return p; }
    // set the PO required times
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        p[i] = *Abc_NodeRequired(pNode);
    }
    return p;
}


/**Function*************************************************************

  Synopsis    [Sets the CI node levels according to the arrival info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkGetCiArrivalFloats(pNtk *Abc_Ntk_t) *float32 {
    var p *float32;
    var pNode *Abc_Obj_t;
    var i int32;
    // inlining of macro ABC_CALLOC
    p = ((float *) calloc((size_t)(Abc_NtkCiNum(pNtk)), sizeof(float)));
    if pNtk.pManTime == nil {
        return p; }
    // set the PI arrival times
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        p[i] = Abc_NodeReadArrivalWorst(pNode);
    }
    return p;
}
func Abc_NtkGetCoRequiredFloats(pNtk *Abc_Ntk_t) *float32 {
    var p *float32;
    var pNode *Abc_Obj_t;
    var i int32;
    if pNtk.pManTime == nil {
        return nil; }
    // set the PO required times
    // inlining of macro ABC_CALLOC
    p = ((float *) calloc((size_t)(Abc_NtkCoNum(pNtk)), sizeof(float)));
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        p[i] = Abc_NodeReadRequiredWorst(pNode);
    }
    return p;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDelayTraceSlackStart(pNtk *Abc_Ntk_t) *Vec_Int_t {
    var vSlacks *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i, k int32;
    vSlacks = Vec_IntAlloc( Abc_NtkObjNumMax(pNtk) + Abc_NtkGetTotalFanins(pNtk) );
    Vec_IntFill( vSlacks, Abc_NtkObjNumMax(pNtk), -1 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        Vec_IntWriteEntry( vSlacks, i, Vec_IntSize(vSlacks) );
        for  k = 0; k < Abc_ObjFaninNum(pObj); k++  {
            Vec_IntPush( vSlacks, -1 );
        }
    }
    }
//    assert( Abc_MaxInt(16, Vec_IntSize(vSlacks)) == Vec_IntCap(vSlacks) );
    return vSlacks;
}

/**Function*************************************************************

  Synopsis    [Read/write edge slacks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDelayTraceSlack(vSlacks *Vec_Int_t, pObj *Abc_Obj_t, iFanin int32) float32 {
    return Abc_Int2Float( Vec_IntEntry( vSlacks, Vec_IntEntry(vSlacks, Abc_ObjId(pObj)) + iFanin ) );
}
func Abc_NtkDelayTraceSetSlack(vSlacks *Vec_Int_t, pObj *Abc_Obj_t, iFanin int32, Num float32) {
    Vec_IntWriteEntry( vSlacks, Vec_IntEntry(vSlacks, Abc_ObjId(pObj)) + iFanin, Abc_Float2Int(Num) );
}

/**Function*************************************************************

  Synopsis    [Find most-critical path (the path with smallest slacks).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDelayTraceCritPath_rec(vSlacks *Vec_Int_t, pNode *Abc_Obj_t, pLeaf *Abc_Obj_t, vBest *Vec_Int_t) int32 {
    var pFanin *Abc_Obj_t;
    var pFaninBest *Abc_Obj_t = nil;
    var SlackMin float32  = ABC_INFINITY;  var i int32;
    // check primary inputs
    if Abc_ObjIsCi(pNode) {
        return (pLeaf == nil || pLeaf == pNode); }
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    // check visited
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return Vec_IntEntry(vBest, Abc_ObjId(pNode)) >= 0; }
    Abc_NodeSetTravIdCurrent( pNode );
    // check the node
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        if !(Abc_NtkDelayTraceCritPath_rec( vSlacks, pFanin, pLeaf, vBest ) != 0) {
            continue; }
        if pFaninBest == nil || SlackMin > Abc_NtkDelayTraceSlack(vSlacks, pNode, i) {
            pFaninBest = pFanin;
            SlackMin = Abc_NtkDelayTraceSlack(vSlacks, pNode, i);
        }
    }
    if pFaninBest != nil {
        Vec_IntWriteEntry( vBest, Abc_ObjId(pNode), Abc_NodeFindFanin(pNode, pFaninBest) ); }
    return (pFaninBest != nil);
}

/**Function*************************************************************

  Synopsis    [Find most-critical path (the path with smallest slacks).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDelayTraceCritPathCollect_rec(vSlacks *Vec_Int_t, pNode *Abc_Obj_t, vBest *Vec_Int_t, vPath *Vec_Ptr_t) {
    assert(Abc_ObjIsCi(pNode) || Abc_ObjIsNode(pNode), "Abc_ObjIsCi(pNode) || Abc_ObjIsNode(pNode)");
    if Abc_ObjIsNode(pNode) {
        var iFanin int32  = Vec_IntEntry( vBest, Abc_ObjId(pNode) );
        assert(iFanin >= 0, "iFanin >= 0");
        Abc_NtkDelayTraceCritPathCollect_rec( vSlacks, Abc_ObjFanin(pNode, iFanin), vBest, vPath );
    }
    Vec_PtrPush( vPath, pNode );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeDelayTraceArrival(pNode *Abc_Obj_t, vSlacks *Vec_Int_t) {
    var pFanin *Abc_Obj_t;
    var pTimeIn, pTimeOut *Abc_Time_t;
    var tDelayBlockRise, tDelayBlockFall float32;
    var PinPhase Mio_PinPhase_t;
    var pPin *Mio_Pin_t;
    var i int32;

    // start the arrival time of the node
    pTimeOut = Abc_NodeArrival(pNode);
    pTimeOut.Rise = pTimeOut.Fall = -ABC_INFINITY; 
    // consider the buffer
    if Abc_ObjIsBarBuf(pNode) {
        pTimeIn = Abc_NodeArrival(Abc_ObjFanin0(pNode));
        *pTimeOut = *pTimeIn;
        return;
    }
    // go through the pins of the gate
    pPin = Mio_GateReadPins((Mio_Gate_t *)pNode.pData);
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        pTimeIn = Abc_NodeArrival(pFanin);
        // get the interesting parameters of this pin
        PinPhase = Mio_PinReadPhase(pPin);
        tDelayBlockRise = float32(Mio_PinReadDelayBlockRise( pPin ));  
        tDelayBlockFall = float32(Mio_PinReadDelayBlockFall( pPin ));  
        // compute the arrival times of the positive phase
        if PinPhase != MIO_PHASE_INV {  // NONINV phase is present
        
            if pTimeOut.Rise < pTimeIn.Rise + tDelayBlockRise {
                pTimeOut.Rise = pTimeIn.Rise + tDelayBlockRise; }
            if pTimeOut.Fall < pTimeIn.Fall + tDelayBlockFall {
                pTimeOut.Fall = pTimeIn.Fall + tDelayBlockFall; }
        }
        if PinPhase != MIO_PHASE_NONINV {  // INV phase is present
        
            if pTimeOut.Rise < pTimeIn.Fall + tDelayBlockRise {
                pTimeOut.Rise = pTimeIn.Fall + tDelayBlockRise; }
            if pTimeOut.Fall < pTimeIn.Rise + tDelayBlockFall {
                pTimeOut.Fall = pTimeIn.Rise + tDelayBlockFall; }
        }
        pPin = Mio_PinReadNext(pPin);
    }

    // compute edge slacks
    if vSlacks != nil {
        var Slack float32;
        // go through the pins of the gate
        pPin = Mio_GateReadPins((Mio_Gate_t *)pNode.pData);
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
            pTimeIn = Abc_NodeArrival(pFanin);
            // get the interesting parameters of this pin
            PinPhase = Mio_PinReadPhase(pPin);
            tDelayBlockRise = float32(Mio_PinReadDelayBlockRise( pPin ));  
            tDelayBlockFall = float32(Mio_PinReadDelayBlockFall( pPin ));  
            // compute the arrival times of the positive phase
            Slack = ABC_INFINITY;
            if PinPhase != MIO_PHASE_INV {  // NONINV phase is present
            
                Slack = Abc_MinFloat( Slack, Abc_AbsFloat(pTimeIn.Rise + tDelayBlockRise - pTimeOut.Rise) );
                Slack = Abc_MinFloat( Slack, Abc_AbsFloat(pTimeIn.Fall + tDelayBlockFall - pTimeOut.Fall) );
            }
            if PinPhase != MIO_PHASE_NONINV {  // INV phase is present
            
                Slack = Abc_MinFloat( Slack, Abc_AbsFloat(pTimeIn.Fall + tDelayBlockRise - pTimeOut.Rise) );
                Slack = Abc_MinFloat( Slack, Abc_AbsFloat(pTimeIn.Rise + tDelayBlockFall - pTimeOut.Fall) );
            }
            pPin = Mio_PinReadNext(pPin);
            Abc_NtkDelayTraceSetSlack( vSlacks, pNode, i, Slack );
        }
    }
}


/**Function*************************************************************

  Synopsis    [Performs delay-trace of the network. If input (pIn) or 
  output (pOut) are given, finds the most-timing-critical path between 
  them and prints it to the standard output. If input and/or output are 
  not given, finds the most-critical path in the network and prints it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDelayTrace(pNtk *Abc_Ntk_t, pOut *Abc_Obj_t, pIn *Abc_Obj_t, fPrint int32) float32 {
    var vSlacks *Vec_Int_t  = nil;
    var pNode, pDriver *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var pTime *Abc_Time_t;
    var tArrivalMax float32;
    var i int32;

    assert(Abc_NtkIsMappedLogic(pNtk), "Abc_NtkIsMappedLogic(pNtk)");
    assert(pOut == nil || Abc_ObjIsCo(pOut), "pOut == nil || Abc_ObjIsCo(pOut)");
    assert(pIn == nil || Abc_ObjIsCi(pIn), "pIn == nil || Abc_ObjIsCi(pIn)");

    // create slacks (need slacks if printing is requested even if pIn/pOut are not given)
    if (pOut != nil) || (pIn != nil) || (fPrint != 0) {
        vSlacks = Abc_NtkDelayTraceSlackStart( pNtk ); }

    // compute the timing
    Abc_NtkTimePrepare( pNtk );
    vNodes = Abc_NtkDfs( pNtk, 1 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Abc_NodeDelayTraceArrival( pNode, vSlacks );
    }
    Vec_PtrFree( vNodes );

    // get the latest arrival times
    tArrivalMax = -ABC_INFINITY;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pDriver = Abc_ObjFanin0(pNode);
        pTime   = Abc_NodeArrival(pDriver);
        if tArrivalMax < Abc_MaxFloat(pTime.Fall, pTime.Rise) {
            tArrivalMax = Abc_MaxFloat(pTime.Fall, pTime.Rise); }
    }

    // determine the output to print
    if (fPrint != 0) && pOut == nil {
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
            pDriver = Abc_ObjFanin0(pNode);
            pTime   = Abc_NodeArrival(pDriver);
            if tArrivalMax == Abc_MaxFloat(pTime.Fall, pTime.Rise) {
                pOut = pNode; }
        }
        assert(pOut != nil, "pOut != nil");
    }

    if fPrint != 0 {
        var vPath *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
        var vBest *Vec_Int_t  = Vec_IntStartFull( Abc_NtkObjNumMax(pNtk) );
        // traverse to determine the critical path
        Abc_NtkIncrementTravId( pNtk );
        if !(Abc_NtkDelayTraceCritPath_rec( vSlacks, Abc_ObjFanin0(pOut), pIn, vBest ) != 0) {
            if pIn == nil {
                fmt.Printf("The logic cone of PO \"%s\" has no primary inputs.\n", Abc_ObjName(pOut));
            } else {
                fmt.Printf("There is no combinational path between PI \"%s\" and PO \"%s\".\n", Abc_ObjName(pIn), Abc_ObjName(pOut)); }
        
        } else {
            var Slack float32 = 0.0;
            var SlackAdd float32;
            var k, iFanin, Length int32;
            var pFanin *Abc_Obj_t;
            // check the additional slack
            SlackAdd = Abc_NodeReadRequiredWorst(pOut) - Abc_NodeReadArrivalWorst(Abc_ObjFanin0(pOut));
            // collect the critical path
            Abc_NtkDelayTraceCritPathCollect_rec( vSlacks, Abc_ObjFanin0(pOut), vBest, vPath );
            if pIn == nil {
                pIn = (Abc_Obj_t *)Vec_PtrEntry( vPath, 0 ); }
            // find the longest gate name
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(vPath)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vPath, i)), 1); i++  {
                if Abc_ObjIsNode(pNode) {
                    Length = Abc_MaxInt( Length, len(Mio_GateReadName((Mio_Gate_t *)pNode.pData)) ); }
            }
            // print critical path
            Abc_NtkLevel( pNtk );
            fmt.Printf("Critical path from PI \"%s\" to PO \"%s\":\n", Abc_ObjName(pIn), Abc_ObjName(pOut)); 
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(vPath)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vPath, i)), 1); i++  {
                fmt.Printf("Level %3d : ", Abc_ObjLevel(pNode));
                if Abc_ObjIsCi(pNode) {
                    fmt.Printf("Primary input \"%s\".   ", Abc_ObjName(pNode));
                    fmt.Printf("Arrival time =%6.1f. ", Abc_NodeReadArrivalWorst(pNode));
                    fmt.Printf("\n");
                    continue;
                }
                if Abc_ObjIsCo(pNode) {
                    fmt.Printf("Primary output \"%s\".   ", Abc_ObjName(pNode));
                    fmt.Printf("Arrival =%6.1f. ", Abc_NodeReadArrivalWorst(pNode));
                
                } else {
                    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
                    iFanin = Abc_NodeFindFanin( pNode, (Abc_Obj_t *)Vec_PtrEntry(vPath,i-1) );
                    Slack = Abc_NtkDelayTraceSlack(vSlacks, pNode, iFanin);
                    fmt.Printf("%10s/", Abc_ObjName(pNode));
                    fmt.Printf("%-4s", Mio_GateReadPinName((Mio_Gate_t *)pNode.pData, iFanin));
                    fmt.Printf(" (%s)", Mio_GateReadName((Mio_Gate_t *)pNode.pData));
                    for  k = len(Mio_GateReadName((Mio_Gate_t *)pNode.pData)); k < Length; k++  {
                        fmt.Printf(" ");
                    }
                    fmt.Printf("   ");
                    fmt.Printf("Arrival =%6.1f.   ", Abc_NodeReadArrivalWorst(pNode));
                    fmt.Printf("I/O times: (");
                    // inlining of macro Abc_ObjForEachFanin
                    for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
                        fmt.Printf("%s%.1f", (k? ", ":""), Abc_NodeReadArrivalWorst(pFanin));
                    }
//                    printf( " -> %.1f)", Abc_NodeReadArrival(pNode)->Worst + Slack + SlackAdd );
                    fmt.Printf(" -> %.1f)", Abc_NodeReadArrivalWorst(pNode));
                }
                fmt.Printf("\n");
            }
            fmt.Printf("Level %3d : ", Abc_ObjLevel(Abc_ObjFanin0(pOut)) + 1);
            fmt.Printf("Primary output \"%s\".   ", Abc_ObjName(pOut));
            fmt.Printf("Required time = %6.1f.  ", Abc_NodeReadRequiredWorst(pOut));
            fmt.Printf("Path slack = %6.1f.\n", SlackAdd);
        }
        Vec_PtrFree( vPath );
        Vec_IntFree( vBest );
    }

    Vec_IntFreeP( &vSlacks );
    return tArrivalMax;
}



/**Function*************************************************************

  Synopsis    [Computes the level of the node using its fanin levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjLevelNew(pObj *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    var i, Level int32;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Level = Abc_MaxFloat( Level, Abc_ObjLevel(pFanin) );
    }
    return Level + int32((Abc_ObjFaninNum(pObj) > 0));
}

/**Function*************************************************************

  Synopsis    [Computes the reverse level of the node using its fanout levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjReverseLevelNew(pObj *Abc_Obj_t) int32 {
    var pFanout *Abc_Obj_t;
    var i, LevelCur, Level int32;
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        LevelCur = Abc_ObjReverseLevel( pFanout );
        Level = Abc_MaxFloat( Level, LevelCur );
    }
    return Level + 1;
}

/**Function*************************************************************

  Synopsis    [Returns required level of the node.]

  Description [Converts the reverse levels of the node into its required 
  level as follows: ReqLevel(Node) = MaxLevels(Ntk) + 1 - LevelR(Node).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjRequiredLevel(pObj *Abc_Obj_t) int32 {
    var pNtk *Abc_Ntk_t  = pObj.pNtk;
    assert(pNtk.vLevelsR, "pNtk.vLevelsR");
    return pNtk.LevelMax + 1 - Abc_ObjReverseLevel(pObj);
}

/**Function*************************************************************

  Synopsis    [Returns the reverse level of the node.]

  Description [The reverse level is the level of the node in reverse
  topological order, starting from the COs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjReverseLevel(pObj *Abc_Obj_t) int32 {
    var pNtk *Abc_Ntk_t  = pObj.pNtk;
    assert(pNtk.vLevelsR, "pNtk.vLevelsR");
    Vec_IntFillExtra( pNtk.vLevelsR, pObj.Id + 1, 0 );
    return Vec_IntEntry(pNtk.vLevelsR, pObj.Id);
}

/**Function*************************************************************

  Synopsis    [Sets the reverse level of the node.]

  Description [The reverse level is the level of the node in reverse
  topological order, starting from the COs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjSetReverseLevel(pObj *Abc_Obj_t, LevelR int32) {
    var pNtk *Abc_Ntk_t  = pObj.pNtk;
    assert(pNtk.vLevelsR, "pNtk.vLevelsR");
    Vec_IntFillExtra( pNtk.vLevelsR, pObj.Id + 1, 0 );
    Vec_IntWriteEntry( pNtk.vLevelsR, pObj.Id, LevelR );
}

/**Function*************************************************************

  Synopsis    [Prepares for the computation of required levels.]

  Description [This procedure should be called before the required times
  are used. It starts internal data structures, which records the level 
  from the COs of the network nodes in reverse topologogical order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkStartReverseLevels(pNtk *Abc_Ntk_t, nMaxLevelIncrease int32) {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    // remember the maximum number of direct levels
    pNtk.LevelMax = Abc_NtkLevel(pNtk) + nMaxLevelIncrease;
    // start the reverse levels
    pNtk.vLevelsR = Vec_IntAlloc( 0 );
    Vec_IntFill( pNtk.vLevelsR, 1 + Abc_NtkObjNumMax(pNtk), 0 );
    // compute levels in reverse topological order
    vNodes = Abc_NtkDfsReverse( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Abc_ObjSetReverseLevel( pObj, Abc_ObjReverseLevelNew(pObj) );
    }
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Cleans the data structures used to compute required levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkStopReverseLevels(pNtk *Abc_Ntk_t) {
    assert(pNtk.vLevelsR, "pNtk.vLevelsR");
    Vec_IntFree( pNtk.vLevelsR );
    pNtk.vLevelsR = nil;
    pNtk.LevelMax = 0;

}

/**Function*************************************************************

  Synopsis    [Incrementally updates level of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkUpdateLevel(pObjNew *Abc_Obj_t, vLevels *Vec_Vec_t) {
    var pFanout, pTemp *Abc_Obj_t;
    var LevelOld, Lev, k, m int32;
//    int Counter = 0, CounterMax = 0;
    // check if level has changed
    LevelOld = Abc_ObjLevel(pObjNew);
    if LevelOld == Abc_ObjLevelNew(pObjNew) {
        return; }
    // start the data structure for level update
    // we cannot fail to visit a node when using this structure because the 
    // nodes are stored by their _old_ levels, which are assumed to be correct
    Vec_VecClear( vLevels );
    Vec_VecPush( vLevels, LevelOld, pObjNew );
    pObjNew.fMarkA = 1;
    // recursively update level
    // inlining of macro Vec_VecForEachEntryStart
    // inlining of macro Vec_PtrForEachEntry
    for  Lev = LevelOld; Lev < Vec_VecSize(vLevels); Lev++  {                                               for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(vLevels, Lev))) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(Vec_VecEntry(vLevels, Lev), k)), 1); k++  {
//        Counter--;
        pTemp.fMarkA = 0;
        assert(Abc_ObjLevel(pTemp) == Lev, "Abc_ObjLevel(pTemp) == Lev");
        Abc_ObjSetLevel( pTemp, Abc_ObjLevelNew(pTemp) );
        // if the level did not change, no need to check the fanout levels
        if Abc_ObjLevel(pTemp) == Lev {
            continue; }
        // schedule fanout for level update
        // inlining of macro Abc_ObjForEachFanout
        for  m = 0; (m < Abc_ObjFanoutNum(pTemp)) && (((pFanout) = Abc_ObjFanout(pTemp, m)), 1); m++  {
            if !Abc_ObjIsCo(pFanout) && !pFanout.fMarkA {
                assert(Abc_ObjLevel(pFanout) >= Lev, "Abc_ObjLevel(pFanout) >= Lev");
                Vec_VecPush( vLevels, Abc_ObjLevel(pFanout), pFanout );
//                Counter++;
//                CounterMax = Abc_MaxFloat( CounterMax, Counter );
                pFanout.fMarkA = 1;
            }
        }
    }
    }
//    printf( "%d ", CounterMax );
}

/**Function*************************************************************

  Synopsis    [Incrementally updates level of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkUpdateReverseLevel(pObjNew *Abc_Obj_t, vLevels *Vec_Vec_t) {
    var pFanin, pTemp *Abc_Obj_t;
    var LevelOld, LevFanin, Lev, k, m int32;
    // check if level has changed
    LevelOld = Abc_ObjReverseLevel(pObjNew);
    if LevelOld == Abc_ObjReverseLevelNew(pObjNew) {
        return; }
    // start the data structure for level update
    // we cannot fail to visit a node when using this structure because the 
    // nodes are stored by their _old_ levels, which are assumed to be correct
    Vec_VecClear( vLevels );
    Vec_VecPush( vLevels, LevelOld, pObjNew );
    pObjNew.fMarkA = 1;
    // recursively update level
    // inlining of macro Vec_VecForEachEntryStart
    // inlining of macro Vec_PtrForEachEntry
    for  Lev = LevelOld; Lev < Vec_VecSize(vLevels); Lev++  {                                               for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(vLevels, Lev))) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(Vec_VecEntry(vLevels, Lev), k)), 1); k++  {
        pTemp.fMarkA = 0;
        LevelOld = Abc_ObjReverseLevel(pTemp); 
        assert(LevelOld == Lev, "LevelOld == Lev");
        Abc_ObjSetReverseLevel( pTemp, Abc_ObjReverseLevelNew(pTemp) );
        // if the level did not change, no need to check the fanout levels
        if Abc_ObjReverseLevel(pTemp) == Lev {
            continue; }
        // schedule fanins for level update
        // inlining of macro Abc_ObjForEachFanin
        for  m = 0; (m < Abc_ObjFaninNum(pTemp)) && (((pFanin) = Abc_ObjFanin(pTemp, m)), 1); m++  {
            if !Abc_ObjIsCi(pFanin) && !pFanin.fMarkA {
                LevFanin = Abc_ObjReverseLevel( pFanin );
                assert(LevFanin >= Lev, "LevFanin >= Lev");
                Vec_VecPush( vLevels, LevFanin, pFanin );
                pFanin.fMarkA = 1;
            }
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Replaces the node and incrementally updates levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkUpdate(pObj *Abc_Obj_t, pObjNew *Abc_Obj_t, vLevels *Vec_Vec_t) {
    // replace the old node by the new node
    pObjNew.Level = pObj.Level;
    Abc_ObjReplace( pObj, pObjNew );
    // update the level of the node
    Abc_NtkUpdateLevel( pObjNew, vLevels );
    Abc_ObjSetReverseLevel( pObjNew, 0 );
    Abc_NtkUpdateReverseLevel( pObjNew, vLevels );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


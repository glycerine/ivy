package sfm

import "fmt"

/**CFile****************************************************************

  FileName    [sfmTim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Timing manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sfmTim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sfmInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Sfm_Tim_t_ struct {
    // external
    pLib    *Mio_Library_t;        // library
    pExt        *Scl_Con_t;        // external timing
    pNtk        *Abc_Ntk_t;        // mapped network
    Delay               int32;       // the largest delay
    DeltaCrit               int32;   // critical delay delta
    // timing info
    vTimArrs         Vec_Int_t;    // arrivals (rise/fall)
    vTimReqs         Vec_Int_t;    // required (rise/fall)
    // incremental timing
    vLevels         Vec_Wec_t;     // levels
    // critical path
    vPath         Vec_Int_t;       // critical path
    vSortData         Vec_Wrd_t;   // node priority order
}

func Sfm_TimArrId(p *Sfm_Tim_t, Id int32) *int32 { return Vec_IntEntryP( &p.vTimArrs,  Abc_Var2Lit(Id, 0) );               }
func Sfm_TimReqId(p *Sfm_Tim_t, Id int32) *int32 { return Vec_IntEntryP( &p.vTimReqs,  Abc_Var2Lit(Id, 0) );               }

func Sfm_TimArr(p *Sfm_Tim_t, pNode *Abc_Obj_t) *int32 { return Vec_IntEntryP( &p.vTimArrs,  Abc_Var2Lit(Abc_ObjId(pNode), 0) ); }
func Sfm_TimReq(p *Sfm_Tim_t, pNode *Abc_Obj_t) *int32 { return Vec_IntEntryP( &p.vTimReqs,  Abc_Var2Lit(Abc_ObjId(pNode), 0) ); }

func Sfm_TimArrMaxId(p *Sfm_Tim_t, Id int32) int32 { var a *int32  = Sfm_TimArrId(p, Id); return Abc_MaxInt(a[0], a[1]);            }

func Sfm_TimArrMax(p *Sfm_Tim_t, pNode *Abc_Obj_t) int32 { var a *int32  = Sfm_TimArr(p, pNode); return Abc_MaxInt(a[0], a[1]);           }
func Sfm_TimSetReq(p *Sfm_Tim_t, pNode *Abc_Obj_t, t int32) { var r *int32  = Sfm_TimReq(p, pNode); r[0] = r[1] = t;                         }
func Sfm_TimSlack(p *Sfm_Tim_t, pNode *Abc_Obj_t) int32 { var r *int32 = Sfm_TimReq(p, pNode);
var a *int32 = Sfm_TimArr(p, pNode); return Abc_MinInt(r[0]-a[0], r[1]-a[1]); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TimEdgeArrival(p *Sfm_Tim_t, pPin *Mio_Pin_t, pTimeIn *int32, pTimeOut *int32) {
    var PinPhase Mio_PinPhase_t  = Mio_PinReadPhase(pPin);
    var tDelayBlockRise int32  = Scl_Flt2Int(Mio_PinReadDelayBlockRise(pPin));  
    var tDelayBlockFall int32  = Scl_Flt2Int(Mio_PinReadDelayBlockFall(pPin));  
    if PinPhase != MIO_PHASE_INV {  // NONINV phase is present
    
        pTimeOut[0] = Abc_MaxInt( pTimeOut[0], pTimeIn[0] + tDelayBlockRise );
        pTimeOut[1] = Abc_MaxInt( pTimeOut[1], pTimeIn[1] + tDelayBlockFall );
    }
    if PinPhase != MIO_PHASE_NONINV {  // INV phase is present
    
        pTimeOut[0] = Abc_MaxInt( pTimeOut[0], pTimeIn[1] + tDelayBlockRise );
        pTimeOut[1] = Abc_MaxInt( pTimeOut[1], pTimeIn[0] + tDelayBlockFall );
    }
}
func Sfm_TimGateArrival(p *Sfm_Tim_t, pGate *Mio_Gate_t, pTimesIn **int32, pTimeOut *int32) {
    var pPin *Mio_Pin_t;  var i int32;
    pTimeOut[0] = pTimeOut[1] = 0;
    // inlining of macro Mio_GateForEachPin
    for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
        Sfm_TimEdgeArrival( p, pPin, pTimesIn[i++], pTimeOut );
    }
    assert(i == Mio_GateReadPinNum(pGate), "i == Mio_GateReadPinNum(pGate)");
}
func Sfm_TimNodeArrival(p *Sfm_Tim_t, pNode *Abc_Obj_t) {
    var i int32;
    var iFanin int32;
    var pTimesIn [6]*int32;
    var pTimeOut *int32  = Sfm_TimArr(p, pNode);
    assert(Abc_ObjFaninNum(pNode) <= 6, "Abc_ObjFaninNum(pNode) <= 6");
    // inlining of macro Abc_ObjForEachFaninId
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((iFanin) = Abc_ObjFaninId(pNode, i)), 1); i++  {
        pTimesIn[i] = Sfm_TimArrId( p, iFanin );
    }
    Sfm_TimGateArrival( p, (Mio_Gate_t *)pNode.pData, pTimesIn, pTimeOut );
}

func Sfm_TimEdgeRequired(p *Sfm_Tim_t, pPin *Mio_Pin_t, pTimeIn *int32, pTimeOut *int32) {
    var PinPhase Mio_PinPhase_t  = Mio_PinReadPhase(pPin);
    var tDelayBlockRise int32  = Scl_Flt2Int(Mio_PinReadDelayBlockRise(pPin));  
    var tDelayBlockFall int32  = Scl_Flt2Int(Mio_PinReadDelayBlockFall(pPin));  
    if PinPhase != MIO_PHASE_INV {  // NONINV phase is present
    
        pTimeIn[0] = Abc_MinInt( pTimeIn[0], pTimeOut[0] - tDelayBlockRise );
        pTimeIn[1] = Abc_MinInt( pTimeIn[1], pTimeOut[1] - tDelayBlockFall );
    }
    if PinPhase != MIO_PHASE_NONINV {  // INV phase is present
    
        pTimeIn[0] = Abc_MinInt( pTimeIn[0], pTimeOut[1] - tDelayBlockRise );
        pTimeIn[1] = Abc_MinInt( pTimeIn[1], pTimeOut[0] - tDelayBlockFall );
    }
}
func Sfm_TimGateRequired(p *Sfm_Tim_t, pGate *Mio_Gate_t, pTimesIn **int32, pTimeOut *int32) {
    var pPin *Mio_Pin_t;  var i int32;
    // inlining of macro Mio_GateForEachPin
    for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
        Sfm_TimEdgeRequired( p, pPin, pTimesIn[i++], pTimeOut );
    }
    assert(i == Mio_GateReadPinNum(pGate), "i == Mio_GateReadPinNum(pGate)");
}
func Sfm_TimNodeRequired(p *Sfm_Tim_t, pNode *Abc_Obj_t) {
    var i int32;
    var iFanin int32;
    var pTimesIn [6]*int32;
    var pTimeOut *int32  = Sfm_TimReq(p, pNode);
    assert(Abc_ObjFaninNum(pNode) <= 6, "Abc_ObjFaninNum(pNode) <= 6");
    // inlining of macro Abc_ObjForEachFaninId
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((iFanin) = Abc_ObjFaninId(pNode, i)), 1); i++  {
        pTimesIn[i] = Sfm_TimReqId( p, iFanin );
    }
    Sfm_TimGateRequired( p, (Mio_Gate_t *)pNode.pData, pTimesIn, pTimeOut );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TimCriticalPath_int(p *Sfm_Tim_t, pObj *Abc_Obj_t, vPath *Vec_Int_t, SlackMax int32) {
    var pNext *Abc_Obj_t; var i int32;
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if Abc_ObjIsCi(pNext) || Abc_ObjFaninNum(pNext) == 0 {
            continue; }
        assert(Abc_ObjIsNode(pNext), "Abc_ObjIsNode(pNext)");
        if Sfm_TimSlack(p, pNext) <= SlackMax {
            Sfm_TimCriticalPath_int( p, pNext, vPath, SlackMax ); }
    }
    if Abc_ObjFaninNum(pObj) > 0 {
        Vec_IntPush( vPath, Abc_ObjId(pObj) ); }
}
func Sfm_TimCriticalPath(p *Sfm_Tim_t, Window int32) int32 {
    var i int32;
    var SlackMax int32 = p.Delay * Window / 100;
    var pObj, pNext *Abc_Obj_t; 
    Vec_IntClear( &p.vPath );
    Abc_NtkIncrementTravId( p.pNtk ); 
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        pNext = Abc_ObjFanin0(pObj);
        if Abc_ObjIsCi(pNext) || Abc_ObjFaninNum(pNext) == 0 {
            continue; }
        assert(Abc_ObjIsNode(pNext), "Abc_ObjIsNode(pNext)");
        if Sfm_TimSlack(p, pNext) <= SlackMax {
            Sfm_TimCriticalPath_int( p, pNext, &p.vPath, SlackMax ); }
    }
    return Vec_IntSize(&p.vPath);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TimTrace(p *Sfm_Tim_t) int32 {
    var pObj *Abc_Obj_t; var i, Delay int32;
    var vNodes *Vec_Ptr_t  = Abc_NtkDfs( p.pNtk, 1 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Sfm_TimNodeArrival( p, pObj );
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        Delay = Abc_MaxInt( Delay, Sfm_TimArrMax(p, Abc_ObjFanin0(pObj)) );
    }
    Vec_IntFill( &p.vTimReqs, 2*Abc_NtkObjNumMax(p.pNtk), ABC_INFINITY );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        Sfm_TimSetReq( p, Abc_ObjFanin0(pObj), Delay );
    }
    // inlining of macro Vec_PtrForEachEntryReverse
    for  i = Vec_PtrSize(vNodes) - 1; (i >= 0) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i--  {
        Sfm_TimNodeRequired( p, pObj );
    }
    Vec_PtrFree( vNodes );
    return Delay;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TimStart(pLib *Mio_Library_t, pExt *Scl_Con_t, pNtk *Abc_Ntk_t, DeltaCrit int32) *Sfm_Tim_t {
    // inlining of macro ABC_CALLOC
    var p *Sfm_Tim_t  = &Sfm_Tim_t{};
    p.pLib = pLib;
    p.pExt = pExt;
    p.pNtk = pNtk;
    Vec_IntFill( &p.vTimArrs,  3*Abc_NtkObjNumMax(pNtk), 0 );
    Vec_IntFill( &p.vTimReqs,  3*Abc_NtkObjNumMax(pNtk), 0 );
    p.Delay = Sfm_TimTrace( p );
    assert(DeltaCrit > 0 && DeltaCrit < Scl_Flt2Int(1000.0), "DeltaCrit > 0 && DeltaCrit < Scl_Flt2Int(1000.0)");
    p.DeltaCrit = DeltaCrit;
    return p;
}
func Sfm_TimStop(p *Sfm_Tim_t) {
    Vec_IntErase( &p.vTimArrs );
    Vec_IntErase( &p.vTimReqs );
    Vec_WecErase( &p.vLevels );
    Vec_IntErase( &p.vPath );
    Vec_WrdErase( &p.vSortData );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Sfm_TimReadNtkDelay(p *Sfm_Tim_t) int32 {
    return p.Delay;
}
func Sfm_TimReadObjDelay(p *Sfm_Tim_t, iObj int32) int32 {
    return Sfm_TimArrMaxId(p, iObj);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TimTest(pNtk *Abc_Ntk_t) {
    var pLib *Mio_Library_t  = (Mio_Library_t *)pNtk.pManFunc;
    var p *Sfm_Tim_t  = Sfm_TimStart( pLib, nil, pNtk, 100 );
    fmt.Printf("Max delay = %.2f.  Path = %d (%d).\n", Scl_Int2Flt(p.Delay), Sfm_TimCriticalPath(p, 1), Abc_NtkNodeNum(p.pNtk));
    Sfm_TimStop( p );
}

/**Function*************************************************************

  Synopsis    [Levelized structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TimUpdateClean(p *Sfm_Tim_t) {
    var vLevel *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i, k int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(&p.vLevels)) && (((vLevel) = Vec_WecEntry(&p.vLevels, i)), 1); i++  {
        // inlining of macro Abc_NtkForEachObjVec
        for  k = 0; k < Vec_IntSize(vLevel) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vLevel,k))), 1); k++  {         if (pObj) == nil {} else {
            assert(pObj.fMarkC == 1, "pObj.fMarkC == 1");
            pObj.fMarkC = 0;
        }
        }
        Vec_IntClear( vLevel );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TimUpdateTiming(p *Sfm_Tim_t, vTimeNodes *Vec_Int_t) {
    assert(Vec_IntSize(vTimeNodes) > 0 && Vec_IntSize(vTimeNodes) <= 2, "Vec_IntSize(vTimeNodes) > 0 && Vec_IntSize(vTimeNodes) <= 2");
    Vec_IntFillExtra( &p.vTimArrs, 2*Abc_NtkObjNumMax(p.pNtk), 0 );
    Vec_IntFillExtra( &p.vTimReqs, 2*Abc_NtkObjNumMax(p.pNtk), 0 );
    p.Delay = Sfm_TimTrace( p );
}

/**Function*************************************************************

  Synopsis    [Sort an array of nodes using their max arrival times.]

  Description [Returns the number of new divisor nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TimSortArrayByArrival(p *Sfm_Tim_t, vNodes *Vec_Int_t, iPivot int32) int32 {
    var Entry word; 
    var i int32;
    var Id int32;
    var Time int32;
    var nDivNew int32 = -1; 
    var MaxDelay int32  = ABC_INFINITY/2+Sfm_TimArrMaxId(p, iPivot);
    assert(p.DeltaCrit > 0, "p.DeltaCrit > 0");
    // collect nodes
    Vec_WrdClear( &p.vSortData );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((Id) = Vec_IntEntry(vNodes, i)), 1); i++  {
        Time = Sfm_TimArrMaxId( p, Id );
        assert(-ABC_INFINITY/2 < Time && Time < ABC_INFINITY/2, "-ABC_INFINITY/2 < Time && Time < ABC_INFINITY/2");
        Vec_WrdPush( &p.vSortData, ((word)Id << 32) | (ABC_INFINITY/2+Time) );
    }
    // sort nodes by delay
    Abc_QuickSort3( Vec_WrdArray(&p.vSortData), Vec_WrdSize(&p.vSortData), 0 );
    // collect sorted nodes and find place where divisors end
    Vec_IntClear( vNodes );
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(&p.vSortData)) && (((Entry) = Vec_WrdEntry(&p.vSortData, i)), 1); i++  {
        Vec_IntPush( vNodes, int32((Entry >> 32)) );
        if nDivNew == -1 && (int32(Entry)) + p.DeltaCrit > MaxDelay {
            nDivNew = i; }
    }
    return nDivNew;
}

/**Function*************************************************************

  Synopsis    [Priority of nodes to try remapping for delay.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TimPriorityNodes(p *Sfm_Tim_t, vCands *Vec_Int_t, Window int32) int32 {
    var vLevel *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i, k int32;
    assert(Window >= 0 && Window <= 100, "Window >= 0 && Window <= 100");
    // collect critical path
    Sfm_TimCriticalPath( p, Window );
    // add nodes to the levelized structure
    Sfm_TimUpdateClean( p );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(&p.vPath) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(&p.vPath,i))), 1); i++  {         if (pObj) == nil {} else {
        assert(pObj.fMarkC == 0, "pObj.fMarkC == 0");
        pObj.fMarkC = 1;
        Vec_WecPush( &p.vLevels, Abc_ObjLevel(pObj), Abc_ObjId(pObj) );
    }
    }
    // prioritize nodes by expected gain
    Vec_WecSort( &p.vLevels, 0 );
    Vec_IntClear( vCands );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(&p.vLevels)) && (((vLevel) = Vec_WecEntry(&p.vLevels, i)), 1); i++  {
        // inlining of macro Abc_NtkForEachObjVec
        for  k = 0; k < Vec_IntSize(vLevel) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(vLevel,k))), 1); k++  {         if (pObj) == nil {} else
            if !pObj.fMarkA {
                Vec_IntPush( vCands, Abc_ObjId(pObj) ); }
        }
    }
//    printf( "Path = %5d   Cand = %5d\n", Vec_IntSize(&p->vPath) );
    return Vec_IntSize(vCands) > 0;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if node is relatively non-critical compared to the pivot.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TimNodeIsNonCritical(p *Sfm_Tim_t, pPivot *Abc_Obj_t, pNode *Abc_Obj_t) int32 {
    return Sfm_TimArrMax(p, pNode) + p.DeltaCrit <= Sfm_TimArrMax(p, pPivot);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_TimEvalRemapping(p *Sfm_Tim_t, vFanins *Vec_Int_t, vMap *Vec_Int_t, pGate1 *Mio_Gate_t, pFans1 string, pGate2 *Mio_Gate_t, pFans2 string) int32 {
    var TimeOut [2][2]int32;
    var pTimesIn1, pTimesIn2 [6]*int32;
    var i, nFanins1, nFanins2 int32;
    // process the first gate
    nFanins1 = Mio_GateReadPinNum( pGate1 );
    for  i = 0; i < nFanins1; i++  {
        pTimesIn1[i] = Sfm_TimArrId( p, Vec_IntEntry(vMap, Vec_IntEntry(vFanins, int32(pFans1[i]))) );
    }
    Sfm_TimGateArrival( p, pGate1, pTimesIn1, TimeOut[0] );
    if pGate2 == nil {
        return Abc_MaxInt(TimeOut[0][0], TimeOut[0][1]); }
    // process the second gate
    nFanins2 = Mio_GateReadPinNum( pGate2 );
    for  i = 0; i < nFanins2; i++  {
        if int32(pFans2[i]) == 16 {
            pTimesIn2[i] = TimeOut[0];
        } else {
            pTimesIn2[i] = Sfm_TimArrId( p, Vec_IntEntry(vMap, Vec_IntEntry(vFanins, int32(pFans2[i]))) ); }
    }
    Sfm_TimGateArrival( p, pGate2, pTimesIn2, TimeOut[1] );
    return Abc_MaxInt(TimeOut[1][0], TimeOut[1][1]);
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


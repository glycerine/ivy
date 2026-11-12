package abs

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [saigSimExt2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Extending simulation trace to contain ternary values.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigSimExt2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abs.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const SAIG_ZER = 1
const SAIG_ONE = 2
const SAIG_UND = 3

func Saig_ManSimInfoNot(Value int32) int32 {
    if Value == SAIG_ZER {
        return SAIG_ONE; }
    if Value == SAIG_ONE {
        return SAIG_ZER; }
    return SAIG_UND;
}

func Saig_ManSimInfoAnd(Value0 int32, Value1 int32) int32 {
    if Value0 == SAIG_ZER || Value1 == SAIG_ZER {
        return SAIG_ZER; }
    if Value0 == SAIG_ONE && Value1 == SAIG_ONE {
        return SAIG_ONE; }
    return SAIG_UND;
}

func Saig_ManSimInfoGet(vSimInfo *Vec_Ptr_t, pObj *Aig_Obj_t, iFrame int32) int32 {
    var pInfo *uint32  = (unsigned *)Vec_PtrEntry( vSimInfo, Aig_ObjId(pObj) );
    return 3 & (pInfo[iFrame >> 4] >> ((iFrame & 15) << 1));
}

func Saig_ManSimInfoSet(vSimInfo *Vec_Ptr_t, pObj *Aig_Obj_t, iFrame int32, Value int32) {
    var pInfo *uint32  = (unsigned *)Vec_PtrEntry( vSimInfo, Aig_ObjId(pObj) );
    assert(Value >= SAIG_ZER && Value <= SAIG_UND, "Value >= SAIG_ZER && Value <= SAIG_UND");
    Value ^= Saig_ManSimInfoGet( vSimInfo, pObj, iFrame );
    pInfo[iFrame >> 4] ^= (Value << ((iFrame & 15) << 1));
}



const SAIG_ZER_NEW = 0   // 0 not visited
const SAIG_ONE_NEW = 1   // 1 not visited
const SAIG_ZER_OLD = 2   // 0 visited
const SAIG_ONE_OLD = 3   // 1 visited

func Saig_ManSimInfo2IsOld(Value int32) int32 {
    return Value == SAIG_ZER_OLD || Value == SAIG_ONE_OLD;
}

func Saig_ManSimInfo2SetOld(Value int32) int32 {
    if Value == SAIG_ZER_NEW {
        return SAIG_ZER_OLD; }
    if Value == SAIG_ONE_NEW {
        return SAIG_ONE_OLD; }
    assert(false, "0");
    return 0;
}

func Saig_ManSimInfo2Not(Value int32) int32 {
    if Value == SAIG_ZER_NEW {
        return SAIG_ONE_NEW; }
    if Value == SAIG_ONE_NEW {
        return SAIG_ZER_NEW; }
    if Value == SAIG_ZER_OLD {
        return SAIG_ONE_OLD; }
    if Value == SAIG_ONE_OLD {
        return SAIG_ZER_OLD; }
    assert(false, "0");
    return 0;
}

func Saig_ManSimInfo2And(Value0 int32, Value1 int32) int32 {
    if Value0 == SAIG_ZER_NEW || Value1 == SAIG_ZER_NEW {
        return SAIG_ZER_NEW; }
    if Value0 == SAIG_ONE_NEW && Value1 == SAIG_ONE_NEW {
        return SAIG_ONE_NEW; }
    assert(false, "0");
    return 0;
}

func Saig_ManSimInfo2Get(vSimInfo *Vec_Ptr_t, pObj *Aig_Obj_t, iFrame int32) int32 {
    var pInfo *uint32  = (unsigned *)Vec_PtrEntry( vSimInfo, Aig_ObjId(pObj) );
    return 3 & (pInfo[iFrame >> 4] >> ((iFrame & 15) << 1));
}

func Saig_ManSimInfo2Set(vSimInfo *Vec_Ptr_t, pObj *Aig_Obj_t, iFrame int32, Value int32) {
    var pInfo *uint32  = (unsigned *)Vec_PtrEntry( vSimInfo, Aig_ObjId(pObj) );
    Value ^= Saig_ManSimInfo2Get( vSimInfo, pObj, iFrame );
    pInfo[iFrame >> 4] ^= (Value << ((iFrame & 15) << 1));
}

// performs ternary simulation
//extern int Saig_ManSimDataInit( Aig_Man_t * p, Abc_Cex_t * pCex, Vec_Ptr_t * vSimInfo, Vec_Int_t * vRes );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs ternary simulation for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManExtendOneEval(vSimInfo *Vec_Ptr_t, pObj *Aig_Obj_t, iFrame int32) int32 {
    var Value0, Value1, Value int32;
    Value0 = Saig_ManSimInfoGet( vSimInfo, Aig_ObjFanin0(pObj), iFrame );
    if Aig_ObjFaninC0(pObj) {
        Value0 = Saig_ManSimInfoNot( Value0 ); }
    if Aig_ObjIsCo(pObj) {
        Saig_ManSimInfoSet( vSimInfo, pObj, iFrame, Value0 );
        return Value0;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Value1 = Saig_ManSimInfoGet( vSimInfo, Aig_ObjFanin1(pObj), iFrame );
    if Aig_ObjFaninC1(pObj) {
        Value1 = Saig_ManSimInfoNot( Value1 ); }
    Value = Saig_ManSimInfoAnd( Value0, Value1 );
    Saig_ManSimInfoSet( vSimInfo, pObj, iFrame, Value );
    return Value;
}

/**Function*************************************************************

  Synopsis    [Performs ternary simulation for one design.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManSimDataInit(p *Aig_Man_t, pCex *Abc_Cex_t, vSimInfo *Vec_Ptr_t, vRes *Vec_Int_t) int32 {
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i, f, Entry, iBit int32;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        Saig_ManSimInfoSet( vSimInfo, pObj, 0, Abc_InfoHasBit(pCex.pData, iBit++)?SAIG_ONE:SAIG_ZER );
    }
    for  f = 0; f <= pCex.iFrame; f++  {
        Saig_ManSimInfoSet( vSimInfo, Aig_ManConst1(p), f, SAIG_ONE );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            Saig_ManSimInfoSet( vSimInfo, pObj, f, Abc_InfoHasBit(pCex.pData, iBit++)?SAIG_ONE:SAIG_ZER );
        }
        if vRes != nil {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vRes)) && (((Entry) = Vec_IntEntry(vRes, i)), 1); i++  {
            Saig_ManSimInfoSet( vSimInfo, Aig_ManCi(p, Entry), f, SAIG_UND );
        } }
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            Saig_ManExtendOneEval( vSimInfo, pObj, f ); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            Saig_ManExtendOneEval( vSimInfo, pObj, f );
        }
        if f == pCex.iFrame {
            break; }
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObjLo)=Saig_ManLo(p, i)), 1); i++  {
            Saig_ManSimInfoSet( vSimInfo, pObjLo, f+1, Saig_ManSimInfoGet(vSimInfo, pObjLi, f) );
        }
    }
    // make sure the output of the property failed
    pObj = Aig_ManCo( p, pCex.iPo );
    return Saig_ManSimInfoGet( vSimInfo, pObj, pCex.iFrame );
}

/**Function*************************************************************

  Synopsis    [Performs ternary simulation for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManExtendOneEval2(vSimInfo *Vec_Ptr_t, pObj *Aig_Obj_t, iFrame int32) int32 {
    var Value0, Value1, Value int32;
    Value0 = Saig_ManSimInfo2Get( vSimInfo, Aig_ObjFanin0(pObj), iFrame );
    if Aig_ObjFaninC0(pObj) {
        Value0 = Saig_ManSimInfo2Not( Value0 ); }
    if Aig_ObjIsCo(pObj) {
        Saig_ManSimInfo2Set( vSimInfo, pObj, iFrame, Value0 );
        return Value0;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Value1 = Saig_ManSimInfo2Get( vSimInfo, Aig_ObjFanin1(pObj), iFrame );
    if Aig_ObjFaninC1(pObj) {
        Value1 = Saig_ManSimInfo2Not( Value1 ); }
    Value = Saig_ManSimInfo2And( Value0, Value1 );
    Saig_ManSimInfo2Set( vSimInfo, pObj, iFrame, Value );
    return Value;
}

/**Function*************************************************************

  Synopsis    [Performs sensitization analysis for one design.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManSimDataInit2(p *Aig_Man_t, pCex *Abc_Cex_t, vSimInfo *Vec_Ptr_t) int32 {
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i, f, iBit int32;
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        Saig_ManSimInfo2Set( vSimInfo, pObj, 0, Abc_InfoHasBit(pCex.pData, iBit++)?SAIG_ONE_NEW:SAIG_ZER_NEW );
    }
    for  f = 0; f <= pCex.iFrame; f++  {
        Saig_ManSimInfo2Set( vSimInfo, Aig_ManConst1(p), f, SAIG_ONE_NEW );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            Saig_ManSimInfo2Set( vSimInfo, pObj, f, Abc_InfoHasBit(pCex.pData, iBit++)?SAIG_ONE_NEW:SAIG_ZER_NEW );
        }
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            Saig_ManExtendOneEval2( vSimInfo, pObj, f ); }
        }
        // inlining of macro Aig_ManForEachCo
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
            Saig_ManExtendOneEval2( vSimInfo, pObj, f );
        }
        if f == pCex.iFrame {
            break; }
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObjLo)=Saig_ManLo(p, i)), 1); i++  {
            Saig_ManSimInfo2Set( vSimInfo, pObjLo, f+1, Saig_ManSimInfo2Get(vSimInfo, pObjLi, f) );
        }
    }
    // make sure the output of the property failed
    pObj = Aig_ManCo( p, pCex.iPo );
    return Saig_ManSimInfo2Get( vSimInfo, pObj, pCex.iFrame );
}

/**Function*************************************************************

  Synopsis    [Drive implications of the given node towards primary outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManSetAndDriveImplications_rec(p *Aig_Man_t, pObj *Aig_Obj_t, f int32, fMax int32, vSimInfo *Vec_Ptr_t) {
    var pFanout *Aig_Obj_t;
    var k int32;
    var iFanout int32 = -1;
    var Value0 int32;
    var Value1 int32;
    var Value int32  = Saig_ManSimInfo2Get( vSimInfo, pObj, f );
    assert(!(Saig_ManSimInfo2IsOld( Value ) != 0), "!Saig_ManSimInfo2IsOld( Value )");
    Saig_ManSimInfo2Set( vSimInfo, pObj, f, Saig_ManSimInfo2SetOld(Value) );
    if (Aig_ObjIsCo(pObj) && f == fMax) || Saig_ObjIsPo(p, pObj) {
        return; }
    if Saig_ObjIsLi( p, pObj ) {
        assert(f < fMax, "f < fMax");
        pFanout = Saig_ObjLiToLo(p, pObj);
        Value = Saig_ManSimInfo2Get( vSimInfo, pFanout, f+1 );
        if !(Saig_ManSimInfo2IsOld( Value ) != 0) {
            Saig_ManSetAndDriveImplications_rec( p, pFanout, f+1, fMax, vSimInfo ); }
        return;
    }
    assert(Aig_ObjIsCi(pObj) || Aig_ObjIsNode(pObj) || Aig_ObjIsConst1(pObj), "Aig_ObjIsCi(pObj) || Aig_ObjIsNode(pObj) || Aig_ObjIsConst1(pObj)");
    // inlining of macro Aig_ObjForEachFanout
    for  assert(p.pFanData, "p.pFanData"), k = 0; (k < int32((pObj).nRefs)) &&                         (((iFanout) = k? Aig_ObjFanoutNext(p, iFanout) : Aig_ObjFanout0Int(p, pObj.Id)), 1) &&           (((pFanout) = Aig_ManObj(p, iFanout>>1)), 1); k++  {
        Value = Saig_ManSimInfo2Get( vSimInfo, pFanout, f );
        if Saig_ManSimInfo2IsOld( Value ) != 0 {
            continue; }
        if Aig_ObjIsCo(pFanout) {
            Saig_ManSetAndDriveImplications_rec( p, pFanout, f, fMax, vSimInfo );
            continue;
        }
        assert(Aig_ObjIsNode(pFanout), "Aig_ObjIsNode(pFanout)");
        Value0 = Saig_ManSimInfo2Get( vSimInfo, Aig_ObjFanin0(pFanout), f );
        Value1 = Saig_ManSimInfo2Get( vSimInfo, Aig_ObjFanin1(pFanout), f );
        if Aig_ObjFaninC0(pFanout) {
            Value0 = Saig_ManSimInfo2Not( Value0 ); }
        if Aig_ObjFaninC1(pFanout) {
            Value1 = Saig_ManSimInfo2Not( Value1 ); }
        if Value0 == SAIG_ZER_OLD || Value1 == SAIG_ZER_OLD || 
            (Value0 == SAIG_ONE_OLD && Value1 == SAIG_ONE_OLD) {
            Saig_ManSetAndDriveImplications_rec( p, pFanout, f, fMax, vSimInfo ); }
    }
}

/**Function*************************************************************

  Synopsis    [Performs recursive sensetization analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManExplorePaths_rec(p *Aig_Man_t, pObj *Aig_Obj_t, f int32, fMax int32, vSimInfo *Vec_Ptr_t) {
    var Value int32  = Saig_ManSimInfo2Get( vSimInfo, pObj, f );
    if Saig_ManSimInfo2IsOld( Value ) != 0 {
        return; }
    Saig_ManSetAndDriveImplications_rec( p, pObj, f, fMax, vSimInfo );
    assert(!Aig_ObjIsConst1(pObj), "!Aig_ObjIsConst1(pObj)");
    if Saig_ObjIsLo(p, pObj) && f == 0 {
        return; }
    if Saig_ObjIsPi(p, pObj) {
        // propagate implications of this assignment
        var i int32;
        var iPiNum int32 = Aig_ObjCioId(pObj);
        for  i = fMax; i >= 0; i--  {
            if i != f {
                Saig_ManSetAndDriveImplications_rec( p, Aig_ManCi(p, iPiNum), i, fMax, vSimInfo ); }
        }
        return;
    }
    if Saig_ObjIsLo( p, pObj ) {
        assert(f > 0, "f > 0");
        Saig_ManExplorePaths_rec( p, Saig_ObjLoToLi(p, pObj), f-1, fMax, vSimInfo );
        return;
    }
    if Aig_ObjIsCo(pObj) {
        Saig_ManExplorePaths_rec( p, Aig_ObjFanin0(pObj), f, fMax, vSimInfo );
        return;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    if Value == SAIG_ZER_OLD {
//        if ( (Aig_ObjId(pObj) & 1) == 0 )
            Saig_ManExplorePaths_rec( p, Aig_ObjFanin0(pObj), f, fMax, vSimInfo );
//        else
//            Saig_ManExplorePaths_rec( p, Aig_ObjFanin1(pObj), f, fMax, vSimInfo );
    
    } else {
        Saig_ManExplorePaths_rec( p, Aig_ObjFanin0(pObj), f, fMax, vSimInfo );
        Saig_ManExplorePaths_rec( p, Aig_ObjFanin1(pObj), f, fMax, vSimInfo );
    }
}

/**Function*************************************************************

  Synopsis    [Returns the array of PIs for flops that should not be absracted.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManProcessCex(p *Aig_Man_t, iFirstFlopPi int32, pCex *Abc_Cex_t, vSimInfo *Vec_Ptr_t, fVerbose int32) *Vec_Int_t {
    var pObj *Aig_Obj_t;
    var vRes, vResInv *Vec_Int_t;
    var i, f, Value int32;
//    assert( Aig_ManRegNum(p) > 0 );
    assert((unsigned *)Vec_PtrEntry(vSimInfo,1) - (unsigned *)Vec_PtrEntry(vSimInfo,0) >= Abc_BitWordNum(2*(pCex.iFrame+1)), "(unsigned *)Vec_PtrEntry(vSimInfo,1) - (unsigned *)Vec_PtrEntry(vSimInfo,0) >= Abc_BitWordNum(2*(pCex.iFrame+1))");
    // start simulation data
    Value = Saig_ManSimDataInit2( p, pCex, vSimInfo );
    assert(Value == SAIG_ONE_NEW, "Value == SAIG_ONE_NEW");
    // derive implications of constants and primary inputs
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
        Saig_ManSetAndDriveImplications_rec( p, pObj, 0, pCex.iFrame, vSimInfo );
    }
    for  f = pCex.iFrame; f >= 0; f--  {
        Saig_ManSetAndDriveImplications_rec( p, Aig_ManConst1(p), f, pCex.iFrame, vSimInfo );
        for  i = 0; i < iFirstFlopPi; i++  {
            Saig_ManSetAndDriveImplications_rec( p, Aig_ManCi(p, i), f, pCex.iFrame, vSimInfo );
        }
    }
    // recursively compute justification
    Saig_ManExplorePaths_rec( p, Aig_ManCo(p, pCex.iPo), pCex.iFrame, pCex.iFrame, vSimInfo );
    // select the result
    vRes = Vec_IntAlloc( 1000 );
    vResInv = Vec_IntAlloc( 1000 );
    for  i = iFirstFlopPi; i < Saig_ManPiNum(p); i++  {
        for  f = pCex.iFrame; f >= 0; f--  {
            Value = Saig_ManSimInfo2Get( vSimInfo, Aig_ManCi(p, i), f );
            if Saig_ManSimInfo2IsOld( Value ) != 0 {
                break; }
        }
        if f >= 0 {
            Vec_IntPush( vRes, i );
        } else {
            Vec_IntPush( vResInv, i ); }
    }
    // resimulate to make sure it is valid
    Value = Saig_ManSimDataInit( p, pCex, vSimInfo, vResInv );
    assert(Value == SAIG_ONE, "Value == SAIG_ONE");
    Vec_IntFree( vResInv );
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Returns the array of PIs for flops that should not be absracted.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManExtendCounterExampleTest2(p *Aig_Man_t, iFirstFlopPi int32, pCex *Abc_Cex_t, fVerbose int32) *Vec_Int_t {
    var vRes *Vec_Int_t;
    var vSimInfo *Vec_Ptr_t;
    var clk abctime;
    if Saig_ManPiNum(p) != pCex.nPis {
        fmt.Printf("Saig_ManExtendCounterExampleTest2(): The PI count of AIG (%d) does not match that of cex (%d).\n", Aig_ManCiNum(p), pCex.nPis);
        return nil;
    }
    Aig_ManFanoutStart( p );
    vSimInfo = Vec_PtrAllocSimInfo( Aig_ManObjNumMax(p), Abc_BitWordNum(2*(pCex.iFrame+1)) );
    Vec_PtrCleanSimInfo( vSimInfo, 0, Abc_BitWordNum(2*(pCex.iFrame+1)) );

clk = Abc_Clock();
    vRes = Saig_ManProcessCex( p, iFirstFlopPi, pCex, vSimInfo, fVerbose );
    if fVerbose != 0 {
        fmt.Printf("Total new PIs = %3d. Non-removable PIs = %3d.  ", Saig_ManPiNum(p)-iFirstFlopPi, Vec_IntSize(vRes));
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    Vec_PtrFree( vSimInfo );
    Aig_ManFanoutStop( p );
    return vRes;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


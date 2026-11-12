package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaRetime.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Performs most-forward retiming for AIG with flop classes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaRetime.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Marks objects reachables from Const0 and PIs/

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMarkAutonomous_rec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return pObj.fMark0; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
    if Gia_ObjIsPi(p, pObj) || Gia_ObjIsConst0(pObj) {
        return pObj.fMark0 = 1; }
    if Gia_ObjIsCo(pObj) {
        return pObj.fMark0 = Gia_ManMarkAutonomous_rec( p, Gia_ObjFanin0(pObj) ); }
    if Gia_ObjIsCi(pObj) {
        return pObj.fMark0 = Gia_ManMarkAutonomous_rec( p, Gia_ObjRoToRi(p, pObj) ); }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if Gia_ManMarkAutonomous_rec( p, Gia_ObjFanin0(pObj) ) != 0 {
        return pObj.fMark0 = 1; }
    return pObj.fMark0 = Gia_ManMarkAutonomous_rec( p, Gia_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Marks with current trav ROs reachable from Const0 and PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMarkAutonomous(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManCleanMark0( p );
    Gia_ManIncrementTravId( p );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ManMarkAutonomous_rec( p, pObj );
    }
    Gia_ManIncrementTravId( p );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if pObj.fMark0 {
            Gia_ObjSetTravIdCurrent( p, pObj ); }
    }
    Gia_ManCleanMark0( p );
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRetimeDup_rec(pNew *Gia_Man_t, pObj *Gia_Obj_t) {
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManRetimeDup_rec( pNew, Gia_ObjFanin0(pObj) );
    Gia_ManRetimeDup_rec( pNew, Gia_ObjFanin1(pObj) );
    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG while retiming the registers to the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRetimeDupForward(p *Gia_Man_t, vCut *Vec_Ptr_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var i int32;
    // create the new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    // create the true PIs
    Gia_ManFillValue( p );
    Gia_ManSetPhase( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // create the registers
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCut)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vCut, i)), 1); i++  {
        pObj.Value = Abc_LitNotCond( Gia_ManAppendCi(pNew), pObj.fPhase );
    }
    // duplicate logic above the cut
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManRetimeDup_rec( pNew, Gia_ObjFanin0(pObj) );
    }
    // create the true POs
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // remember value in LI
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.Value = Gia_ObjFanin0Copy(pObj);
    }
    // transfer values from the LIs to the LOs
    // inlining of macro Gia_ManForEachRiRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        pObjRo.Value = pObjRi.Value;
    }
    // erase the data values on the internal nodes of the cut
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCut)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vCut, i)), 1); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = ~0; }
    }
    // duplicate logic below the cut
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCut)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vCut, i)), 1); i++  {
        Gia_ManRetimeDup_rec( pNew, pObj );
        Gia_ManAppendCo( pNew, Abc_LitNotCond( pObj.Value, pObj.fPhase ) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Vec_PtrSize(vCut) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Derives the cut for forward retiming.]

  Description [Assumes topological ordering of the nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRetimeForwardOne(p *Gia_Man_t, pnRegFixed *int32, pnRegMoves *int32) *Gia_Man_t {
    var vFlopClasses *Vec_Int_t  = nil;
    var vObjClasses *Vec_Int_t  = nil;
    var pNew *Gia_Man_t;
    var vCut *Vec_Ptr_t;
    var pObj *Gia_Obj_t;
    var i int32;
    if p.vFlopClasses {
//        printf( "Performing retiming with register classes.\n" );
        vObjClasses = Vec_IntAlloc( Gia_ManObjNum(p) );
        for  i = 0; i < Gia_ManObjNum(p); i++  {
            Vec_IntPush( vObjClasses, -1 );
        }
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            Vec_IntWriteEntry( vObjClasses, Gia_ObjId(p, pObj), Vec_IntEntry(p.vFlopClasses, i) );
        }
        vFlopClasses = Vec_IntAlloc( Gia_ManRegNum(p) );
    }
    // mark the retimable nodes
    Gia_ManIncrementTravId( p );
    Gia_ManMarkAutonomous( p );
    // mark the retimable registers with the fresh trav ID
    Gia_ManIncrementTravId( p );
    *pnRegFixed = 0;
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if Gia_ObjIsTravIdPrevious(p, pObj) {
            Gia_ObjSetTravIdCurrent(p, pObj);
        } else {
            (*pnRegFixed)++; }
    }
    // mark all the nodes that can be retimed forward
    *pnRegMoves = 0;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Gia_ObjIsTravIdCurrent(p, Gia_ObjFanin0(pObj)) && Gia_ObjIsTravIdCurrent(p, Gia_ObjFanin1(pObj)) {
            if (vObjClasses != nil) && Vec_IntEntry(vObjClasses, Gia_ObjFaninId0(pObj, i)) != Vec_IntEntry(vObjClasses, Gia_ObjFaninId1(pObj, i)) {
                continue; }
            if vObjClasses != nil {
                Vec_IntWriteEntry( vObjClasses, Gia_ObjId(p, pObj), Vec_IntEntry(vObjClasses, Gia_ObjFaninId0(pObj, i)) ); }
            Gia_ObjSetTravIdCurrent(p, pObj);
            (*pnRegMoves)++;
        }
    }
    // mark the remaining registers
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ObjSetTravIdCurrent(p, pObj);
    }
    // find the cut (all such marked objects that fanout into unmarked nodes)
    vCut = Vec_PtrAlloc( 1000 );
    Gia_ManIncrementTravId( p );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsTravIdPrevious(p, pObj) {
            continue; }
        if (Gia_ObjIsCo(pObj) || Gia_ObjIsAnd(pObj)) && Gia_ObjIsTravIdPrevious(p, Gia_ObjFanin0(pObj)) {
            if vFlopClasses != nil {
                Vec_IntPush( vFlopClasses, Vec_IntEntry(vObjClasses, Gia_ObjFaninId0(pObj, i)) ); }
            Vec_PtrPush( vCut, Gia_ObjFanin0(pObj) );
            Gia_ObjSetTravIdCurrent( p, Gia_ObjFanin0(pObj) );
        }
        if Gia_ObjIsAnd(pObj) && Gia_ObjIsTravIdPrevious(p, Gia_ObjFanin1(pObj)) {
            if vFlopClasses != nil {
                Vec_IntPush( vFlopClasses, Vec_IntEntry(vObjClasses, Gia_ObjFaninId1(pObj, i)) ); }
            Vec_PtrPush( vCut, Gia_ObjFanin1(pObj) );
            Gia_ObjSetTravIdCurrent( p, Gia_ObjFanin1(pObj) );
        }
    }
    assert(vFlopClasses == nil || Vec_IntSize(vFlopClasses) == Vec_PtrSize(vCut), "vFlopClasses == nil || Vec_IntSize(vFlopClasses) == Vec_PtrSize(vCut)");
    // finally derive the new manager
    pNew = Gia_ManRetimeDupForward( p, vCut );
    Vec_PtrFree( vCut );
    if vObjClasses != nil {
    Vec_IntFree( vObjClasses ); }
    pNew.vFlopClasses = vFlopClasses;
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Derives the cut for forward retiming.]

  Description [Assumes topological ordering of the nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManRetimeForward(p *Gia_Man_t, nMaxIters int32, fVerbose int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var i int32;
    var nRegFixed int32;
    var nRegMoves int32 = 1;
    var clk abctime;
    pNew = p;
    for  i = 0; i < nMaxIters && nRegMoves > 0; i++  {
        clk = Abc_Clock();
        pNew = Gia_ManRetimeForwardOne( pTemp = pNew, &nRegFixed, &nRegMoves );
        if fVerbose != 0 {
            fmt.Printf("%2d : And = %6d. Reg = %5d. Unret = %5d. Move = %6d. ", i + 1, Gia_ManAndNum(pTemp), Gia_ManRegNum(pTemp), nRegFixed, nRegMoves);
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }
        if pTemp != p {
            Gia_ManStop( pTemp ); }
    }
/*
    clk = Abc_Clock();
    pNew = Gia_ManReduceLaches( pNew, fVerbose );
    if ( fVerbose )
    {
        ABC_PRT( "Register sharing time", Abc_Clock() - clk );
    }
*/
    return pNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


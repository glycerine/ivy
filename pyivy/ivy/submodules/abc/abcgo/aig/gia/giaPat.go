package gia

import "fmt"

/**CFile****************************************************************

  FileName    [gia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Sat_ObjXValue(pObj *Gia_Obj_t) int32 { return (pObj.fMark1 << 1) | pObj.fMark0;             }
func Sat_ObjSetXValue(pObj *Gia_Obj_t, v int32) { pObj.fMark0 = (v & 1); pObj.fMark1 = ((v >> 1) & 1); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Collects nodes in the cone and initialized them to x.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SatCollectCone_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vVisit *Vec_Int_t) {
    if Sat_ObjXValue(pObj) == GIA_UND {
        return; }
    if Gia_ObjIsAnd(pObj) {
        Gia_SatCollectCone_rec( p, Gia_ObjFanin0(pObj), vVisit );
        Gia_SatCollectCone_rec( p, Gia_ObjFanin1(pObj), vVisit );
    }
    assert(Sat_ObjXValue(pObj) == 0, "Sat_ObjXValue(pObj) == 0");
    Sat_ObjSetXValue( pObj, GIA_UND );
    Vec_IntPush( vVisit, Gia_ObjId(p, pObj) );
}

/**Function*************************************************************

  Synopsis    [Collects nodes in the cone and initialized them to x.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SatCollectCone(p *Gia_Man_t, pObj *Gia_Obj_t, vVisit *Vec_Int_t) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    assert(!Gia_ObjIsConst0(pObj), "!Gia_ObjIsConst0(pObj)");
    assert(Sat_ObjXValue(pObj) == 0, "Sat_ObjXValue(pObj) == 0");
    Vec_IntClear( vVisit );
    Gia_SatCollectCone_rec( p, pObj, vVisit );
}

/**Function*************************************************************

  Synopsis    [Checks if the counter-examples asserts the output.]

  Description [Assumes that fMark0 and fMark1 are clean. Leaves them clean.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SatVerifyPattern(p *Gia_Man_t, pRoot *Gia_Obj_t, vCex *Vec_Int_t, vVisit *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i, Entry, Value, Value0, Value1 int32;
    assert(Gia_ObjIsCo(pRoot), "Gia_ObjIsCo(pRoot)");
    assert(!Gia_ObjIsConst0(Gia_ObjFanin0(pRoot)), "!Gia_ObjIsConst0(Gia_ObjFanin0(pRoot))");
    // collect nodes and initialized them to x
    Gia_SatCollectCone( p, Gia_ObjFanin0(pRoot), vVisit );
    // set binary values to nodes in the counter-example
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCex)) && (((Entry) = Vec_IntEntry(vCex, i)), 1); i++  {
//        Sat_ObjSetXValue( Gia_ManObj(p, Abc_Lit2Var(Entry)), Abc_LitIsCompl(Entry)? GIA_ZER : GIA_ONE );
        Sat_ObjSetXValue( Gia_ManCi(p, Abc_Lit2Var(Entry)), Abc_LitIsCompl(Entry)? GIA_ZER : GIA_ONE );
    }
    // simulate
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vVisit)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vVisit,i))); i++  {
        if Gia_ObjIsCi(pObj) {
            continue; }
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        Value0 = Sat_ObjXValue( Gia_ObjFanin0(pObj) );
        Value1 = Sat_ObjXValue( Gia_ObjFanin1(pObj) );
        Value = Gia_XsimAndCond( Value0, Gia_ObjFaninC0(pObj), Value1, Gia_ObjFaninC1(pObj) );
        Sat_ObjSetXValue( pObj, Value );
    }
    Value = Sat_ObjXValue( Gia_ObjFanin0(pRoot) );
    Value = Gia_XsimNotCond( Value, Gia_ObjFaninC0(pRoot) );
    if Value != GIA_ONE {
        fmt.Printf("Gia_SatVerifyPattern(): Verification FAILED.\n"); }
//    else
//        printf( "Gia_SatVerifyPattern(): Verification succeeded.\n" );
//    assert( Value == GIA_ONE );
    // clean the nodes
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vVisit)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vVisit,i))); i++  {
        Sat_ObjSetXValue( pObj, 0 );
    }
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


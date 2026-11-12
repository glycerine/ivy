package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaTis.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Technology independent synthesis.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaTis.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Derives GIA with MUXes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTisDupMuxes(p *Gia_Man_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pFan0, pFan1, pFanC *Gia_Obj_t;
    var i int32;
    assert(p.pMuxes == nil, "p.pMuxes == nil");
    // inlining of macro ABC_FREE
    ((p.pRefs) ? (free((char *) (p.pRefs)), (p.pRefs) = 0) : 0);
    Gia_ManCreateRefs( p ); 
    // start the new manager
    pNew = Gia_ManStart( 5000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // inlining of macro ABC_CALLOC
    pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned)));
    // create constant
    Gia_ManConst0(p).Value = 0;
    // create PIs
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // create internal nodes
    Gia_ManHashStart( pNew );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !Gia_ObjIsMuxType(pObj) || (Gia_ObjRefNum(p, Gia_ObjFanin0(pObj)) > 1 && Gia_ObjRefNum(p, Gia_ObjFanin1(pObj)) > 1) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) {
            pObj.Value = Gia_ManHashXorReal( pNew, Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan0)), Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan1)) );
        } else {
            pFanC = Gia_ObjRecognizeMux( pObj, &pFan1, &pFan0 );
            pObj.Value = Gia_ManHashMuxReal( pNew, Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFanC)), Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan1)), Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan0)) );
        }
    }
    }
    Gia_ManHashStop( pNew );
    // create ROs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // perform cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTisCollectMffc_rec(p *Gia_Man_t, Id int32, vMffc *Vec_Int_t, vLeaves *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, Id) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, Id);
    if Gia_ObjRefNumId(p, Id) > 1 {
        Vec_IntPush( vLeaves, Id );
        return;
    }
    pObj = Gia_ManObj( p, Id );
    if Gia_ObjIsCi(pObj) {
        Vec_IntPush( vLeaves, Id );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManTisCollectMffc_rec( p, Gia_ObjFaninId0(pObj, Id), vMffc, vLeaves );
    Gia_ManTisCollectMffc_rec( p, Gia_ObjFaninId1(pObj, Id), vMffc, vLeaves );
    if Gia_ObjIsMuxId(p, Id) {
        Gia_ManTisCollectMffc_rec( p, Gia_ObjFaninId2(p, Id), vMffc, vLeaves ); }
    Vec_IntPush( vMffc, Id );
}
func Gia_ManTisCollectMffc(p *Gia_Man_t, Id int32, vMffc *Vec_Int_t, vLeaves *Vec_Int_t) {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, Id );
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Vec_IntClear( vMffc );
    Vec_IntClear( vLeaves );
    Gia_ManIncrementTravId( p );
    Gia_ManTisCollectMffc_rec( p, Gia_ObjFaninId0(pObj, Id), vMffc, vLeaves );
    Gia_ManTisCollectMffc_rec( p, Gia_ObjFaninId1(pObj, Id), vMffc, vLeaves );
    if Gia_ObjIsMuxId(p, Id) {
        Gia_ManTisCollectMffc_rec( p, Gia_ObjFaninId2(p, Id), vMffc, vLeaves ); }
    Vec_IntPush( vMffc, Id );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTisPrintMffc(p *Gia_Man_t, Id int32, vMffc *Vec_Int_t, vLeaves *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    fmt.Printf("MFFC %d has %d nodes and %d leaves:\n", Id, Vec_IntSize(vMffc), Vec_IntSize(vLeaves));
    // inlining of macro Gia_ManForEachObjVecReverse
    for  i = Vec_IntSize(vMffc) - 1; (i >= 0) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vMffc,i))); i--  {
        fmt.Printf("Node %2d : ", Vec_IntSize(vMffc) - 1 - i);
        Gia_ObjPrint( p, pObj );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vLeaves)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vLeaves,i))); i++  {
        fmt.Printf("Leaf %2d : ", i);
        Gia_ObjPrint( p, pObj );
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTisTest(pInit *Gia_Man_t) {
    var p *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vMffc, vLeaves *Vec_Int_t;
    var i int32;
    vMffc = Vec_IntAlloc( 10 );
    vLeaves = Vec_IntAlloc( 10 );
    p = Gia_ManTisDupMuxes( pInit );
    Gia_ManCreateRefs( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjRefNumId(p, i) == 1 {
            continue; }
        Gia_ManTisCollectMffc( p, i, vMffc, vLeaves );
        Gia_ManTisPrintMffc( p, i, vMffc, vLeaves );
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if Gia_ObjRefNumId(p, Gia_ObjFaninId0p(p, pObj)) > 1 {
            continue; }
        Gia_ManTisCollectMffc( p, Gia_ObjFaninId0p(p, pObj), vMffc, vLeaves );
        Gia_ManTisPrintMffc( p, Gia_ObjFaninId0p(p, pObj), vMffc, vLeaves );
    }
    Gia_ManStop( p );
    Vec_IntFree( vMffc );
    Vec_IntFree( vLeaves );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


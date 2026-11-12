package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecUtil.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"

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
func Gia_PolynCollectXors_rec(pGia *Gia_Man_t, iObj int32, vXors *Vec_Int_t) {
    var pObj *Gia_Obj_t  = Gia_ManObj( pGia, iObj );
    if Gia_ObjIsTravIdCurrent(pGia, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(pGia, pObj);
    if !Gia_ObjIsAnd(pObj) || !Gia_ObjIsXor(pObj) || Gia_ObjRefNum(pGia, pObj) > 1 {
        return; }
    Gia_PolynCollectXors_rec( pGia, Gia_ObjFaninId0(pObj, iObj), vXors );
    Gia_PolynCollectXors_rec( pGia, Gia_ObjFaninId1(pObj, iObj), vXors );
    Vec_IntPushUnique( vXors, iObj );
}
func Gia_PolynCollectLastXor(pGia *Gia_Man_t, fVerbose int32) *Vec_Int_t {
    var vXors *Vec_Int_t  = Vec_IntAlloc( 100 );
    var pObj *Gia_Obj_t  = Gia_ManCo( pGia, Gia_ManCoNum(pGia)-1 );
    pGia.pRefs = nil;
    Gia_ManCreateRefs( pGia );
    Gia_ManIncrementTravId( pGia );
    Gia_PolynCollectXors_rec( pGia, Gia_ObjFaninId0p(pGia, pObj), vXors );
    Vec_IntReverseOrder( vXors );
    pGia.pRefs = nil;
    return vXors;
}
func Gia_PolynAnalyzeXors(pGia *Gia_Man_t, fVerbose int32) {
    var i, iDriver, Count int32;
    var vXors *Vec_Int_t  = Vec_IntAlloc( 100 );
    if pGia.pMuxes == nil {
        fmt.Printf("AIG does not have XORs extracted.\n");
        return;
    }
    assert(pGia.pMuxes, "pGia.pMuxes");
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && (((iDriver) = Gia_ObjFaninId0p(pGia, Gia_ManCo(pGia, i))), 1); i++  {
        Vec_IntClear( vXors );
        Gia_ManIncrementTravId( pGia );
        Gia_PolynCollectXors_rec( pGia, iDriver, vXors );
        //printf( "%3d : ", i );
        //Vec_IntPrint( vXors );
        fmt.Printf("%d=%d  ", i, Vec_IntSize(vXors));
        Count += Vec_IntSize(vXors);
    }
    fmt.Printf("Total = %d.\n", Count);
    Vec_IntFree( vXors );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupTopMostRange(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var vTops *Vec_Int_t  = Vec_IntAlloc( 10 );
    var i int32;
    for  i = 45; i < 52; i++  {
        Vec_IntPush( vTops, Gia_ObjId( p, Gia_ObjFanin0(Gia_ManCo(p, i)) ) );
    }
    pNew = Gia_ManDupAndConesLimit( p, Vec_IntArray(vTops), Vec_IntSize(vTops), 100 );
    Vec_IntFree( vTops );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


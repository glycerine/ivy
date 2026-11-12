package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecBo.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecBo.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecWec.h"
//#include "misc/extra/extra.h"

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
func Acec_DetectBoothXorMux(p *Gia_Man_t, pMux *Gia_Obj_t, pXor *Gia_Obj_t, pIns []int32) int32 {
    var pFan0, pFan1 *Gia_Obj_t;
    var pDat0, pDat1, pCtrl *Gia_Obj_t;
    if !Gia_ObjIsMuxType(pMux) || !Gia_ObjIsMuxType(pXor) {
        return 0; }
    if !Gia_ObjRecognizeExor( pXor, &pFan0, &pFan1 ) {
        return 0; }
    pFan0 = Gia_Regular(pFan0);
    pFan1 = Gia_Regular(pFan1);
    if Gia_ObjId(p, pFan0) > Gia_ObjId(p, pFan1) {
        // inlining of macro ABC_SWAP
         var t *Gia_Obj_t  = pFan0; pFan0 = pFan1; pFan1 = t; };
    if !(pCtrl = Gia_ObjRecognizeMux( pMux, &pDat0, &pDat1 )) {
        return 0; }
    pDat0 = Gia_Regular(pDat0);
    pDat1 = Gia_Regular(pDat1);
    pCtrl = Gia_Regular(pCtrl);
    if !Gia_ObjIsAnd(pDat0) || !Gia_ObjIsAnd(pDat1) {
        return 0; }
    if Gia_ObjFaninId0p(p, pDat0) != Gia_ObjFaninId0p(p, pDat1) ||
         Gia_ObjFaninId1p(p, pDat0) != Gia_ObjFaninId1p(p, pDat1) {
         return 0; }
    if Gia_ObjFaninId0p(p, pDat0) != Gia_ObjId(p, pFan0) ||
         Gia_ObjFaninId1p(p, pDat0) != Gia_ObjId(p, pFan1) {
         return 0; }
    pIns[0] = Gia_ObjId(p, pFan0);
    pIns[1] = Gia_ObjId(p, pFan1);
    pIns[2] = Gia_ObjId(p, pCtrl);
    return 1;
}
func Acec_DetectBoothXorFanin(p *Gia_Man_t, pObj *Gia_Obj_t, pIns []int32) int32 {
    var pFan0, pFan1 *Gia_Obj_t;
    //int Id = Gia_ObjId(p, pObj);
    if !Gia_ObjIsAnd(pObj) {
        return 0; }
    if !Gia_ObjFaninC0(pObj) || !Gia_ObjFaninC1(pObj) {
        return 0; }
    pFan0 = Gia_ObjFanin0(pObj);
    pFan1 = Gia_ObjFanin1(pObj);
    if !Gia_ObjIsAnd(pFan0) || !Gia_ObjIsAnd(pFan1) {
        return 0; }
    if Acec_DetectBoothXorMux(p, Gia_ObjFanin0(pFan0), Gia_ObjFanin0(pFan1), pIns) != 0 {
        pIns[3] = Gia_ObjId(p, Gia_ObjFanin1(pFan0));
        pIns[4] = Gia_ObjId(p, Gia_ObjFanin1(pFan1));
        return 1;
    }
    if Acec_DetectBoothXorMux(p, Gia_ObjFanin0(pFan0), Gia_ObjFanin1(pFan1), pIns) != 0 {
        pIns[3] = Gia_ObjId(p, Gia_ObjFanin1(pFan0));
        pIns[4] = Gia_ObjId(p, Gia_ObjFanin0(pFan1));
        return 1;
    }
    if Acec_DetectBoothXorMux(p, Gia_ObjFanin1(pFan0), Gia_ObjFanin0(pFan1), pIns) != 0 {
        pIns[3] = Gia_ObjId(p, Gia_ObjFanin0(pFan0));
        pIns[4] = Gia_ObjId(p, Gia_ObjFanin1(pFan1));
        return 1;
    }
    if Acec_DetectBoothXorMux(p, Gia_ObjFanin1(pFan0), Gia_ObjFanin1(pFan1), pIns) != 0 {
        pIns[3] = Gia_ObjId(p, Gia_ObjFanin0(pFan0));
        pIns[4] = Gia_ObjId(p, Gia_ObjFanin0(pFan1));
        return 1;
    }
    return 0;
}
func Acec_DetectBoothOne(p *Gia_Man_t, pObj *Gia_Obj_t, pIns []int32) int32 {
    var pFan0, pFan1 *Gia_Obj_t;
    if !Gia_ObjRecognizeExor( pObj, &pFan0, &pFan1 ) {
        return 0; }
    pFan0 = Gia_Regular(pFan0);
    pFan1 = Gia_Regular(pFan1);
    if (Acec_DetectBoothXorFanin( p, pFan0, pIns ) != 0) && pIns[2] == Gia_ObjId(p, pFan1) {
        return 1; }
    if (Acec_DetectBoothXorFanin( p, pFan1, pIns ) != 0) && pIns[2] == Gia_ObjId(p, pFan0) {
        return 1; }
    return 0;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_DetectBoothTwoXor(p *Gia_Man_t, pObj *Gia_Obj_t, pIns []int32) int32 {
    var pFan0, pFan1 *Gia_Obj_t;
    if !Gia_ObjIsAnd(pObj) {
        return 0; }
    if Gia_ObjRecognizeExor( Gia_ObjFanin0(pObj), &pFan0, &pFan1 ) {
        pIns[0] = Gia_ObjId(p, Gia_Regular(pFan0));
        pIns[1] = Gia_ObjId(p, Gia_Regular(pFan1));
        pIns[2] = -1;
        pIns[3] = 0;
        pIns[4] = Gia_ObjId(p, Gia_ObjFanin1(pObj));
        return 1;
    }
    if Gia_ObjRecognizeExor( Gia_ObjFanin1(pObj), &pFan0, &pFan1 ) {
        pIns[0] = Gia_ObjId(p, Gia_Regular(pFan0));
        pIns[1] = Gia_ObjId(p, Gia_Regular(pFan1));
        pIns[2] = -1;
        pIns[3] = 0;
        pIns[4] = Gia_ObjId(p, Gia_ObjFanin0(pObj));
        return 1;
    }
    return 0;
}
func Acec_DetectBoothTwo(p *Gia_Man_t, pObj *Gia_Obj_t, pIns []int32) int32 {
    var pFan0, pFan1 *Gia_Obj_t;
    if !Gia_ObjRecognizeExor( pObj, &pFan0, &pFan1 ) {
        return 0; }
    pFan0 = Gia_Regular(pFan0);
    pFan1 = Gia_Regular(pFan1);
    if Acec_DetectBoothTwoXor( p, pFan0, pIns ) != 0 {
        pIns[2] = Gia_ObjId(p, pFan1);
        return 1;
    }
    if Acec_DetectBoothTwoXor( p, pFan1, pIns ) != 0 {
        pIns[2] = Gia_ObjId(p, pFan0);
        return 1;
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_DetectBoothTest(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    var pIns [5]int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {        
        if !(Acec_DetectBoothOne(p, pObj, pIns) != 0) && !(Acec_DetectBoothTwo(p, pObj, pIns) != 0) {
                continue; }
        fmt.Printf("obj = %4d  :  b0 = %4d  b1 = %4d  b2 = %4d    a0 = %4d  a1 = %4d\n", i, pIns[0], pIns[1], pIns[2], pIns[3], pIns[4]);
    }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


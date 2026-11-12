package gia

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [giaSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Fast sequential simulator.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Derives signal implications.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Finds signals, which reset flops to have constant values.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns index of the first pattern that failed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns index of the PO and pattern that failed it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimSimulate(g *Gia_Man_t, pars *Gia_ParSim_t) int32 {
    return giaSimSimulate(g, pars, Gia_ManSimSimulateEquiv)
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Bit-parallel simulation during AIG construction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Finds a satisfying assignment.]

  Description [Returns 1 if a sat assignment is found; 0 otherwise.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManObjCheckSat_rec(p *Gia_Man_t, iLit int32, vObjs *Vec_Int_t) int32 {
    var iObj int32  = Abc_Lit2Var(iLit);
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if pObj.fMark0 {
        return pObj.fMark1 == uint32(Abc_LitIsCompl(iLit)); }
    pObj.fMark0 = 1;
    pObj.fMark1 = Abc_LitIsCompl(iLit);
    Vec_IntPush( vObjs, iObj );
    if Gia_ObjIsAnd(pObj) {
        if pObj.fMark1 == 0 { // node value is 1
        
            if !(Gia_ManObjCheckSat_rec( p, Gia_ObjFaninLit0(pObj, iObj), vObjs ) != 0) {   return 0; }
            if !(Gia_ManObjCheckSat_rec( p, Gia_ObjFaninLit1(pObj, iObj), vObjs ) != 0) {   return 0; }
        
        } else {
            if !(Gia_ManObjCheckSat_rec( p, Abc_LitNot(Gia_ObjFaninLit0(pObj, iObj)), vObjs ) != 0) {   return 0; }
        }
    }
    return 1;
}
func Gia_ManObjCheckOverlap1(p *Gia_Man_t, iLit0 int32, iLit1 int32, vObjs *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Res0, Res1 int32;
//    Gia_ManForEachObj( p, pObj, i )
//        assert( pObj->fMark0 == 0 && pObj->fMark1 == 0 );
    Vec_IntClear( vObjs );
    Res0 = Gia_ManObjCheckSat_rec( p, iLit0, vObjs );
    if Res0 {
        Res1 = Gia_ManObjCheckSat_rec( p, iLit1, vObjs ); }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        pObj.fMark0 = pObj.fMark1 = 0;
    }
    return Res0 && Res1;
}
func Gia_ManObjCheckOverlap(p *Gia_Man_t, iLit0 int32, iLit1 int32, vObjs *Vec_Int_t) int32 {
    if Gia_ManObjCheckOverlap1( p, iLit0, iLit1, vObjs ) != 0 {
        return 1; }
    return Gia_ManObjCheckOverlap1( p, iLit1, iLit0, vObjs );
}





/**Function*************************************************************

  Synopsis    [Bit-parallel simulation during AIG construction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


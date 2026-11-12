package cec

/**CFile****************************************************************

  FileName    [cecChoice.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Computation of structural choices.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecChoice.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"
//#include "aig/gia/giaAig.h"
//#include "proof/dch/dch.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Cec_ManCombSpecReduce_rec( Gia_Man_t * pNew, Gia_Man_t * p, Gia_Obj_t * pObj );


// extern int Cec_ManResimulateCounterExamplesComb( Cec_ManSim_t * pSim, Vec_Int_t * vCexStore );

// extern int Gia_ManCheckRefinements( Gia_Man_t * p, Vec_Str_t * vStatus, Vec_Int_t * vOutputs, Cec_ManSim_t * pSim, int fRings );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes the real value of the literal w/o spec reduction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManCombSpecReal(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Cec_ManCombSpecReduce_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Cec_ManCombSpecReduce_rec( pNew, p, Gia_ObjFanin1(pObj) );
    return Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Recursively performs speculative reduction for the object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManCombSpecReduce_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    var pRepr *Gia_Obj_t;
    if ~pObj.Value {
        return; }
    if pRepr = Gia_ObjReprObj(p, Gia_ObjId(p, pObj)) {
        Cec_ManCombSpecReduce_rec( pNew, p, pRepr );
        pObj.Value = Abc_LitNotCond( pRepr.Value, Gia_ObjPhase(pRepr) ^ Gia_ObjPhase(pObj) );
        return;
    }
    pObj.Value = Cec_ManCombSpecReal( pNew, p, pObj );
}

/**Function*************************************************************

  Synopsis    [Derives SRM for signal correspondence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManCombSpecReduce(p *Gia_Man_t, pvOutputs **Vec_Int_t, fRings int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pRepr *Gia_Obj_t;
    var vXorLits *Vec_Int_t;
    var i, iPrev, iObj, iPrevNew, iObjNew int32;
    assert(p.pReprs != nil, "p.pReprs != nil");
    Gia_ManSetPhase( p );
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    *pvOutputs = Vec_IntAlloc( 1000 );
    vXorLits = Vec_IntAlloc( 1000 );
    if fRings != 0 {
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if Gia_ObjIsConst( p, i ) {
                iObjNew = Cec_ManCombSpecReal( pNew, p, pObj );
                iObjNew = Abc_LitNotCond( iObjNew, Gia_ObjPhase(pObj) );
                if iObjNew != 0 {
                    Vec_IntPush( *pvOutputs, 0 );
                    Vec_IntPush( *pvOutputs, i );
                    Vec_IntPush( vXorLits, iObjNew );
                }
            
            } else if Gia_ObjIsHead( p, i ) {
                iPrev = i;
                // inlining of macro Gia_ClassForEachObj1
                for  assert(Gia_ObjIsHead(p, i), "Gia_ObjIsHead(p, i)"), iObj = Gia_ObjNext(p, i); iObj > 0; iObj = Gia_ObjNext(p, iObj)  {
                    iPrevNew = Cec_ManCombSpecReal( pNew, p, Gia_ManObj(p, iPrev) );
                    iObjNew  = Cec_ManCombSpecReal( pNew, p, Gia_ManObj(p, iObj) );
                    iPrevNew = Abc_LitNotCond( iPrevNew, Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iPrev)) );
                    iObjNew  = Abc_LitNotCond( iObjNew,  Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iObj)) );
                    if iPrevNew != iObjNew && iPrevNew != 0 && iObjNew != 1 {
                        Vec_IntPush( *pvOutputs, iPrev );
                        Vec_IntPush( *pvOutputs, iObj );
                        Vec_IntPush( vXorLits, Gia_ManHashAnd(pNew, iPrevNew, Abc_LitNot(iObjNew)) );
                    }
                    iPrev = iObj;
                }
                iObj = i;
                iPrevNew = Cec_ManCombSpecReal( pNew, p, Gia_ManObj(p, iPrev) );
                iObjNew  = Cec_ManCombSpecReal( pNew, p, Gia_ManObj(p, iObj) );
                iPrevNew = Abc_LitNotCond( iPrevNew, Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iPrev)) );
                iObjNew  = Abc_LitNotCond( iObjNew,  Gia_ObjPhase(pObj) ^ Gia_ObjPhase(Gia_ManObj(p, iObj)) );
                if iPrevNew != iObjNew && iPrevNew != 0 && iObjNew != 1 {
                    Vec_IntPush( *pvOutputs, iPrev );
                    Vec_IntPush( *pvOutputs, iObj );
                    Vec_IntPush( vXorLits, Gia_ManHashAnd(pNew, iPrevNew, Abc_LitNot(iObjNew)) );
                }
            }
        }
    
    } else {
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            pRepr = Gia_ObjReprObj( p, Gia_ObjId(p,pObj) );
            if pRepr == nil {
                continue; }
            if Gia_ObjIsConst(p, i) {
                iPrevNew = 0;
            } else {
                iPrevNew = Cec_ManCombSpecReal( pNew, p, pRepr );
            }
            iObjNew  = Cec_ManCombSpecReal( pNew, p, pObj );
            iObjNew  = Abc_LitNotCond( iObjNew, Gia_ObjPhase(pRepr) ^ Gia_ObjPhase(pObj) );
            if iPrevNew != iObjNew {
                Vec_IntPush( *pvOutputs, Gia_ObjId(p, pRepr) );
                Vec_IntPush( *pvOutputs, Gia_ObjId(p, pObj) );
                Vec_IntPush( vXorLits, Gia_ManHashXor(pNew, iPrevNew, iObjNew) );
            }
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXorLits)) && (((iObjNew) = Vec_IntEntry(vXorLits, i)), 1); i++  {
        Gia_ManAppendCo( pNew, iObjNew );
    }
    Vec_IntFree( vXorLits );
    Gia_ManHashStop( pNew );
//Abc_Print( 1, "Before sweeping = %d\n", Gia_ManAndNum(pNew) );
    pNew = Gia_ManCleanup( pTemp = pNew );
//Abc_Print( 1, "After sweeping = %d\n", Gia_ManAndNum(pNew) );
    Gia_ManStop( pTemp );
    return pNew;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Computes choices for the vector of AIGs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Computes choices for one AIGs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Performs computation of AIGs with choices.]

  Description [Takes several AIGs and performs choicing.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Performs computation of AIGs with choices.]

  Description [Takes several AIGs and performs choicing.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


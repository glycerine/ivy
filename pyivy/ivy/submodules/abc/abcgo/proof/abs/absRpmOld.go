package abs

import "fmt"

/**CFile****************************************************************

  FileName    [absRpmOld.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Old min-cut-based reparametrization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absRpmOld.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abs.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Specialized duplication.]

  Description [Replaces registers by PIs/POs and PIs by registers.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupIn2Ff(p *Gia_Man_t) *Gia_Man_t {
    var vPiOuts *Vec_Int_t;
    var pNew *Gia_Man_t; 
    var pObj *Gia_Obj_t;
    var i int32;
    vPiOuts = Vec_IntAlloc( Gia_ManPiNum(p) );
    pNew = Gia_ManStart( Gia_ManObjNum(p) + 2 * Gia_ManPiNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Vec_IntPush( vPiOuts, Gia_ManAppendCi(pNew) );
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Gia_ManAppendCo( pNew, Vec_IntEntry(vPiOuts, i) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManPiNum(p) );
    Vec_IntFree( vPiOuts );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Reverses the above step.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupFf2In_rec(pNew *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if pObj.Value != ~0 {
        return pObj.Value; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManDupFf2In_rec( pNew, Gia_ObjFanin0(pObj) );
    Gia_ManDupFf2In_rec( pNew, Gia_ObjFanin1(pObj) );
    return pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Reverses the above step.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupFf2In(p *Gia_Man_t, nFlopsOld int32) *Gia_Man_t {
    var pNew *Gia_Man_t; 
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    for  i = Gia_ManPiNum(p) - nFlopsOld; i < Gia_ManPiNum(p); i++  {
        Gia_ManPi(p, i).Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupFf2In_rec( pNew, Gia_ObjFanin0(pObj) );
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, nFlopsOld );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Reparameterized to get rid of useless primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_RpmPerformOld(p *Gia_Man_t, fVerbose int32) *Gia_Man_t {
//    extern Aig_Man_t * Saig_ManRetimeMinArea( Aig_Man_t * p, int nMaxIters, int fForwardOnly, int fBackwardOnly, int fInitial, int fVerbose );
    var pMan, pTemp *Aig_Man_t;
    var pNew, pTmp *Gia_Man_t;
    var nFlopsOld int32  = Gia_ManRegNum(p);
    if fVerbose != 0 {
        fmt.Printf("Original AIG:\n");
        Gia_ManPrintStats( p, nil );
    }

    // perform input trimming
    pNew = Gia_ManDupTrimmed( p, 1, 0, 0, -1 );
    if fVerbose != 0 {
        fmt.Printf("After PI trimming:\n");
        Gia_ManPrintStats( pNew, nil );
    }
    // transform GIA
    pNew = Gia_ManDupIn2Ff( pTmp = pNew );
    Gia_ManStop( pTmp );
    if fVerbose != 0 {
        fmt.Printf("After PI-2-FF transformation:\n");
        Gia_ManPrintStats( pNew, nil );
    }

    // derive AIG
    pMan = Gia_ManToAigSimple( pNew );
    Gia_ManStop( pNew );
    // perform min-reg retiming
    pMan = Saig_ManRetimeMinArea( pTemp = pMan, 10, 0, 0, 1, 0 );
    Aig_ManStop( pTemp );
    // derive GIA
    pNew = Gia_ManFromAigSimple( pMan );
    Aig_ManStop( pMan );
    if fVerbose != 0 {
        fmt.Printf("After min-area retiming:\n");
        Gia_ManPrintStats( pNew, nil );
    }

    // transform back
    pNew = Gia_ManDupFf2In( pTmp = pNew, nFlopsOld );
    Gia_ManStop( pTmp );
    if fVerbose != 0 {
        fmt.Printf("After FF-2-PI tranformation:\n");
        Gia_ManPrintStats( pNew, nil );
    }
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaAbs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Counter-example-guided abstraction refinement.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaAbs.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 
//#include "gia.h"

//#include "sat/bsat/satSolver.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Resimulates the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManVerifyCex(pAig *Gia_Man_t, p *Abc_Cex_t, fDualOut int32) int32 {
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var RetValue, i, k, iBit int32;
    Gia_ManCleanMark0(pAig);
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+i)); i++  {
        pObj.fMark0 = Abc_InfoHasBit(p.pData, iBit++);
    }
    for  i = 0; i <= p.iFrame; i++  {
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(pAig)) && ((pObj) = Gia_ManCi(pAig, k)); k++  {
            pObj.fMark0 = Abc_InfoHasBit(p.pData, iBit++);
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                           (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(pAig.vCos)) && ((pObj) = Gia_ManCo(pAig, k)); k++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
        if i == p.iFrame {
            break; }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(pAig)) && ((pObjRi) = Gia_ManCo(pAig, Gia_ManPoNum(pAig)+k)) && ((pObjRo) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+k)); k++  {
            pObjRo.fMark0 = pObjRi.fMark0;
        }
    }
    assert(iBit == p.nBits, "iBit == p.nBits");
    if fDualOut != 0 {
        RetValue = Gia_ManPo(pAig, 2*p.iPo).fMark0 ^ Gia_ManPo(pAig, 2*p.iPo+1).fMark0;
    } else {
        RetValue = Gia_ManPo(pAig, p.iPo).fMark0; }
    Gia_ManCleanMark0(pAig);
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Resimulates the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFindFailedPoCex(pAig *Gia_Man_t, p *Abc_Cex_t, nOutputs int32) int32 {
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var RetValue, i, k, iBit int32;
    assert(Gia_ManPiNum(pAig) == p.nPis, "Gia_ManPiNum(pAig) == p.nPis");
    Gia_ManCleanMark0(pAig);
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+i)); i++  {
        pObj.fMark0 = Abc_InfoHasBit(p.pData, iBit++);
    }
    iBit = p.nRegs;
    for  i = 0; i <= p.iFrame; i++  {
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(pAig)) && ((pObj) = Gia_ManCi(pAig, k)); k++  {
            pObj.fMark0 = Abc_InfoHasBit(p.pData, iBit++);
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                           (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(pAig.vCos)) && ((pObj) = Gia_ManCo(pAig, k)); k++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(pAig)) && ((pObjRi) = Gia_ManCo(pAig, Gia_ManPoNum(pAig)+k)) && ((pObjRo) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+k)); k++  {
            pObjRo.fMark0 = pObjRi.fMark0;
        }
    }
    assert(iBit == p.nBits, "iBit == p.nBits");
    // figure out the number of failed output
    RetValue = -1;
//    for ( i = Gia_ManPoNum(pAig) - 1; i >= nOutputs; i-- )
    for  i = nOutputs; i < Gia_ManPoNum(pAig); i++  {
        if Gia_ManPo(pAig, i).fMark0 {
            RetValue = i;
            break;
        }
    }
    Gia_ManCleanMark0(pAig);
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Determines the failed PO when its exact frame is not known.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSetFailedPoCex(pAig *Gia_Man_t, p *Abc_Cex_t) int32 {
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var i, k, iBit int32;
    assert(Gia_ManPiNum(pAig) == p.nPis, "Gia_ManPiNum(pAig) == p.nPis");
    Gia_ManCleanMark0(pAig);
    p.iPo = -1;
//    Gia_ManForEachRo( pAig, pObj, i )
//       pObj->fMark0 = Abc_InfoHasBit(p->pData, iBit++);
    iBit = p.nRegs;
    for  i = 0; i <= p.iFrame; i++  {
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(pAig)) && ((pObj) = Gia_ManCi(pAig, k)); k++  {
            pObj.fMark0 = Abc_InfoHasBit(p.pData, iBit++);
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                           (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(pAig.vCos)) && ((pObj) = Gia_ManCo(pAig, k)); k++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(pAig)) && ((pObjRi) = Gia_ManCo(pAig, Gia_ManPoNum(pAig)+k)) && ((pObjRo) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+k)); k++  {
            pObjRo.fMark0 = pObjRi.fMark0;
        }
        // check the POs
        // inlining of macro Gia_ManForEachPo
        for  k = 0; (k < Gia_ManPoNum(pAig)) && ((pObj) = Gia_ManCo(pAig, k)); k++  {
            if !pObj.fMark0 {
                continue; }
            p.iPo = k;
            p.iFrame = i;
            p.nBits = iBit;
            break;
        }
    }
    Gia_ManCleanMark0(pAig);
    return p.iPo;
}


/**Function*************************************************************

  Synopsis    [Starts the process of returning values for internal nodes.]

  Description [Should be called when pCex is available, before probing 
  any object for its value using Gia_ManCounterExampleValueLookup().]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCounterExampleValueStart(pGia *Gia_Man_t, pCex *Abc_Cex_t) {
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var Val0, Val1, nObjs, i, k, iBit int32;
    assert(Gia_ManRegNum(pGia) > 0, "Gia_ManRegNum(pGia) > 0"); // makes sense only for sequential AIGs
    assert(pGia.pData2 == nil, "pGia.pData2 == nil");    // if this fail, there may be a memory leak
    // allocate memory to store simulation bits for internal nodes
    // inlining of macro ABC_CALLOC
    pGia.pData2 = ((unsigned *) calloc((size_t)(Abc_BitWordNum( (pCex.iFrame + 1) * Gia_ManObjNum(pGia) )), sizeof(unsigned)));
    // the register values in the counter-example should be zero
    // inlining of macro Gia_ManForEachRo
    for  k = 0; (k < Gia_ManRegNum(pGia)) && ((pObj) = Gia_ManCi(pGia, Gia_ManPiNum(pGia)+k)); k++  {
        assert(Abc_InfoHasBit(pCex.pData, iBit++) == 0, "Abc_InfoHasBit(pCex.pData, iBit++) == 0");
    }
    // iterate through the timeframes
    nObjs = Gia_ManObjNum(pGia);
    for  i = 0; i <= pCex.iFrame; i++  {
        // no need to set constant-0 node
        // set primary inputs according to the counter-example
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(pGia)) && ((pObj) = Gia_ManCi(pGia, k)); k++  {
            if Abc_InfoHasBit(pCex.pData, iBit++) {
                Abc_InfoSetBit( (unsigned *)pGia.pData2, nObjs * i + Gia_ObjId(pGia, pObj) ); }
        }
        // compute values for each node
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            Val0 = Abc_InfoHasBit( (unsigned *)pGia.pData2, nObjs * i + Gia_ObjFaninId0p(pGia, pObj) );
            Val1 = Abc_InfoHasBit( (unsigned *)pGia.pData2, nObjs * i + Gia_ObjFaninId1p(pGia, pObj) );
            if (Val0 ^ Gia_ObjFaninC0(pObj)) & (Val1 ^ Gia_ObjFaninC1(pObj)) {
                Abc_InfoSetBit( (unsigned *)pGia.pData2, nObjs * i + Gia_ObjId(pGia, pObj) ); }
        }
        }
        // derive values for combinational outputs
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(pGia.vCos)) && ((pObj) = Gia_ManCo(pGia, k)); k++  {
            Val0 = Abc_InfoHasBit( (unsigned *)pGia.pData2, nObjs * i + Gia_ObjFaninId0p(pGia, pObj) );
            if Val0 ^ Gia_ObjFaninC0(pObj) {
                Abc_InfoSetBit( (unsigned *)pGia.pData2, nObjs * i + Gia_ObjId(pGia, pObj) ); }
        }
        if i == pCex.iFrame {
            continue; }
        // transfer values to the register output of the next frame
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(pGia)) && ((pObjRi) = Gia_ManCo(pGia, Gia_ManPoNum(pGia)+k)) && ((pObjRo) = Gia_ManCi(pGia, Gia_ManPiNum(pGia)+k)); k++  {
            if Abc_InfoHasBit( (unsigned *)pGia.pData2, nObjs * i + Gia_ObjId(pGia, pObjRi) ) {
                Abc_InfoSetBit( (unsigned *)pGia.pData2, nObjs * (i+1) + Gia_ObjId(pGia, pObjRo) ); }
        }
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    // check that the counter-example is correct, that is, the corresponding output is asserted
    assert(Abc_InfoHasBit( (unsigned *)pGia.pData2, nObjs * pCex.iFrame + Gia_ObjId(pGia, Gia_ManCo(pGia, pCex.iPo)) ), "Abc_InfoHasBit( (unsigned *)pGia.pData2, nObjs * pCex.iFrame + Gia_ObjId(pGia, Gia_ManCo(pGia, pCex.iPo)) )");
}

/**Function*************************************************************

  Synopsis    [Stops the process of returning values for internal nodes.]

  Description [Should be called when probing is no longer needed]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCounterExampleValueStop(pGia *Gia_Man_t) {
    assert(pGia.pData2 != nil, "pGia.pData2 != nil");    // if this fail, we try to call this procedure more than once
    // inlining of macro ABC_FREE
    ((pGia.pData2) ? (free((char *) (pGia.pData2)), (pGia.pData2) = 0) : 0);
    pGia.pData2 = nil;
}

/**Function*************************************************************

  Synopsis    [Returns the value of the given object in the given timeframe.]

  Description [Should be called to probe the value of an object with 
  the given ID (iFrame is a 0-based number of a time frame - should not 
  exceed the number of timeframes in the original counter-example).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCounterExampleValueLookup(pGia *Gia_Man_t, Id int32, iFrame int32) int32 {
    assert(Id >= 0 && Id < Gia_ManObjNum(pGia), "Id >= 0 && Id < Gia_ManObjNum(pGia)");
    return Abc_InfoHasBit( (unsigned *)pGia.pData2, Gia_ManObjNum(pGia) * iFrame + Id );
}

/**Function*************************************************************

  Synopsis    [Procedure to test the above code.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCounterExampleValueTest(pGia *Gia_Man_t, pCex *Abc_Cex_t) {
    var pObj *Gia_Obj_t  = Gia_ManObj( pGia, Gia_ManObjNum(pGia)/2 );
    var iFrame int32  = Abc_MaxInt( 0, pCex.iFrame - 1 );
    fmt.Printf("\nUsing counter-example, which asserts output %d in frame %d.\n", pCex.iPo, pCex.iFrame);
    Gia_ManCounterExampleValueStart( pGia, pCex );
    fmt.Printf("Value of object %d in frame %d is %d.\n", Gia_ObjId(pGia, pObj), iFrame, Gia_ManCounterExampleValueLookup(pGia, Gia_ObjId(pGia, pObj), iFrame));
    Gia_ManCounterExampleValueStop( pGia );
}


/**Function*************************************************************

  Synopsis    [Returns CEX containing PI+CS values for each timeframe.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCexExtendToIncludeCurrentStates(p *Gia_Man_t, pCex *Abc_Cex_t) *Abc_Cex_t {
    var pNew *Abc_Cex_t;
    var pObj, pObjRo, pObjRi *Gia_Obj_t;
    var i, k, iBit int32;
    assert(pCex.nRegs > 0, "pCex.nRegs > 0");
    // start the counter-example
    pNew = Abc_CexAlloc( 0, Gia_ManCiNum(p), pCex.iFrame + 1 );
    pNew.iFrame = pCex.iFrame;
    pNew.iPo    = pCex.iPo;
    // set const0
    Gia_ManConst0(p).fMark0 = 0;
    // set init state
    // inlining of macro Gia_ManForEachRi
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)); k++  {
        pObjRi.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
    }
    assert(iBit == pCex.nRegs, "iBit == pCex.nRegs");
    for  i = 0; i <= pCex.iFrame; i++  {
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
            pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
        }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            pObjRo.fMark0 = pObjRi.fMark0;
        }
        // inlining of macro Gia_ManForEachCi
        for  k = 0; (k < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, k)); k++  {
            if pObj.fMark0 {
                Abc_InfoSetBit( pNew.pData, pNew.nPis * i + k ); }
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    assert(Gia_ManPo(p, pCex.iPo).fMark0 == 1, "Gia_ManPo(p, pCex.iPo).fMark0 == 1");
    Gia_ManCleanMark0(p);
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Returns CEX containing all object valuess for each timeframe.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCexExtendToIncludeAllObjects(p *Gia_Man_t, pCex *Abc_Cex_t) *Abc_Cex_t {
    var pNew *Abc_Cex_t;
    var pObj, pObjRo, pObjRi *Gia_Obj_t;
    var i, k, iBit int32;
    assert(pCex.nRegs > 0, "pCex.nRegs > 0");
    // start the counter-example
    pNew = Abc_CexAlloc( 0, Gia_ManObjNum(p), pCex.iFrame + 1 );
    pNew.iFrame = pCex.iFrame;
    pNew.iPo    = pCex.iPo;
    // set const0
    Gia_ManConst0(p).fMark0 = 0;
    // set init state
    // inlining of macro Gia_ManForEachRi
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)); k++  {
        pObjRi.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
    }
    assert(iBit == pCex.nRegs, "iBit == pCex.nRegs");
    for  i = 0; i <= pCex.iFrame; i++  {
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
            pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++);
        }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            pObjRo.fMark0 = pObjRi.fMark0;
        }
        // inlining of macro Gia_ManForEachObj
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {
            if pObj.fMark0 {
                Abc_InfoSetBit( pNew.pData, pNew.nPis * i + k ); }
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    assert(Gia_ManPo(p, pCex.iPo).fMark0 == 1, "Gia_ManPo(p, pCex.iPo).fMark0 == 1");
    Gia_ManCleanMark0(p);
    return pNew;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFramesForCexMin(p *Gia_Man_t, nFrames int32) *Gia_Man_t {
    var pFrames, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t; var i, f int32;
    assert(Gia_ManPoNum(p) == 1, "Gia_ManPoNum(p) == 1");
    pFrames = Gia_ManStart( Gia_ManObjNum(p) );
    pFrames.pName = Abc_UtilStrsav( p.pName );
    pFrames.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pFrames );
    Gia_ManConst0(p).Value = 0;
    for  f = 0; f < nFrames; f++  {
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            if f != 0 {
                pObj.Value = Gia_ObjRoToRi( p, pObj ).Value;
            } else {
                pObj.Value = 0;
            }
        }
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pFrames );
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = Gia_ManHashAnd( pFrames, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
            pObj.Value = Gia_ObjFanin0Copy(pObj);
        }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pFrames, Gia_ObjFanin0Copy(pObj) );
    }
    pFrames = Gia_ManCleanup( pTemp = pFrames );
    //printf( "Before cleanup = %d nodes. After cleanup = %d nodes.\n", 
    //    Gia_ManAndNum(pTemp), Gia_ManAndNum(pFrames) );
    Gia_ManStop( pTemp );
    return pFrames;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMinCex(p *Gia_Man_t, pCex *Abc_Cex_t) {
    var clk abctime  = Abc_Clock();
    var n, i, iFirstVar, iLit, status, Counter int32;//, Id;
    var vLits *Vec_Int_t;
    var pSat *sat_solver;
    var pCnf *Cnf_Dat_t;
    var nFinal int32;
    var pFinal *int32;
    var pCexCare *Abc_Cex_t;
    var pFrames *Gia_Man_t;

    // CEX minimization
    clk = Abc_Clock();
    pCexCare = Bmc_CexCareMinimizeAig( p, Gia_ManPiNum(p), pCex, 1, 1, 1 );
    for  i = pCexCare.nRegs; i < pCexCare.nBits; i++  {
        Counter += Abc_InfoHasBit(pCexCare.pData, i);
    }
    Abc_CexFree( pCexCare );
    fmt.Printf("Care bits = %d. ", Counter);
    Abc_PrintTime( 1, "CEX minimization", Abc_Clock() - clk );

    // SAT instance
    clk = Abc_Clock();
    pFrames = Gia_ManFramesForCexMin( p, pCex.iFrame + 1 );
    pCnf = (Cnf_Dat_t*)Mf_ManGenerateCnf( pFrames, 8, 0, 0, 0, 0 );
    iFirstVar = pCnf.nVars - (pCex.iFrame+1) * pCex.nPis;
    pSat = (sat_solver*)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    iLit = Abc_Var2Lit( 1, 1 );
    status = sat_solver_addclause( pSat, &iLit, &iLit + 1 );
    assert((status != 0), "status");
    // create literals
    vLits = Vec_IntAlloc( 100 );
    for  i = pCex.nRegs; i < pCex.nBits; i++  {
        Vec_IntPush( vLits, Abc_Var2Lit(iFirstVar + i - pCex.nRegs, !Abc_InfoHasBit(pCex.pData, i)) );
    }
    Abc_PrintTime( 1, "SAT solver", Abc_Clock() - clk );

    for  n = 0; n < 2; n++  {
        if n { Vec_IntReverseOrder( vLits ); }

        // SAT-based minimization
        clk = Abc_Clock();
        status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits), 0, 0, 0, 0 );
        nFinal = sat_solver_final( pSat, &pFinal );
        fmt.Printf("Status %d.  Selected %d assumptions out of %d.  ", status, nFinal, Vec_IntSize(vLits));
        Abc_PrintTime( 1, "Analyze_final", Abc_Clock() - clk );

        // SAT-based minimization
        clk = Abc_Clock();
        nFinal = sat_solver_minimize_assumptions( pSat, Vec_IntArray(vLits), Vec_IntSize(vLits), 0 );
        fmt.Printf("Status %d.  Selected %d assumptions out of %d.  ", status, nFinal, Vec_IntSize(vLits));
        Abc_PrintTime( 1, "LEXUNSAT", Abc_Clock() - clk );
    }

    // cleanup
    Vec_IntFree( vLits );
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    Gia_ManStop( pFrames );
}


func Bmc_CexCareDeriveCex(pCex *Abc_Cex_t, iFirstVar int32, pLits *int32, nLits int32) *Abc_Cex_t {
    var pCexMin *Abc_Cex_t; var i int32;
    pCexMin = Abc_CexAlloc( pCex.nRegs, pCex.nPis, pCex.iFrame + 1 );
    pCexMin.iPo    = pCex.iPo;
    pCexMin.iFrame = pCex.iFrame;
    for  i = 0; i < nLits; i++  {
        var PiNum int32  = Abc_Lit2Var(pLits[i]) - iFirstVar;
        assert(PiNum >= 0 && PiNum < pCex.nBits - pCex.nRegs, "PiNum >= 0 && PiNum < pCex.nBits - pCex.nRegs");
        Abc_InfoSetBit( pCexMin.pData, pCexMin.nRegs + PiNum );
    }
    return pCexMin;
}
func Bmc_CexCareSatBasedMinimizeAig(p *Gia_Man_t, pCex *Abc_Cex_t, fHighEffort int32, fVerbose int32) *Abc_Cex_t {
    var clk abctime  = Abc_Clock();
    var n, i, iFirstVar, iLit, status int32;
    var vLits *Vec_Int_t = nil;
    var vTemp *Vec_Int_t;
    var pSat *sat_solver;
    var pCnf *Cnf_Dat_t;
    var nFinal int32;
    var pFinal *int32;
    var pCexBest *Abc_Cex_t  = nil; 
    var CountBest int32;
    var pFrames *Gia_Man_t;

    // CEX minimization
    clk = Abc_Clock();
    pCexBest = Bmc_CexCareMinimizeAig( p, Gia_ManPiNum(p), pCex, 1, 1, fVerbose );
    for  i = pCexBest.nRegs; i < pCexBest.nBits; i++  {
        CountBest += Abc_InfoHasBit(pCexBest.pData, i);
    }
    if fVerbose != 0 {
        fmt.Printf("Care bits = %d. ", CountBest);
        Abc_PrintTime( 1, "Non-SAT-based CEX minimization", Abc_Clock() - clk );
    }

    // SAT instance
    clk = Abc_Clock();
    pFrames = Gia_ManFramesForCexMin( p, pCex.iFrame + 1 );
    pCnf = (Cnf_Dat_t*)Mf_ManGenerateCnf( pFrames, 8, 0, 0, 0, 0 );
    iFirstVar = pCnf.nVars - (pCex.iFrame+1) * pCex.nPis;
    pSat = (sat_solver*)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    iLit = Abc_Var2Lit( 1, 1 );
    status = sat_solver_addclause( pSat, &iLit, &iLit + 1 );
    assert((status != 0), "status");
    // create literals
    vTemp = Vec_IntAlloc( 100 );
    for  i = pCex.nRegs; i < pCex.nBits; i++  {
        Vec_IntPush( vTemp, Abc_Var2Lit(iFirstVar + i - pCex.nRegs, !Abc_InfoHasBit(pCex.pData, i)) );
    }
    if fVerbose != 0 {
    Abc_PrintTime( 1, "Constructing SAT solver", Abc_Clock() - clk ); }

    for  n = 0; n < 2; n++  {
        Vec_IntFreeP( &vLits );

        vLits = Vec_IntDup( vTemp );
        if n { Vec_IntReverseOrder( vLits ); }

        // SAT-based minimization
        clk = Abc_Clock();
        status = sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntLimit(vLits), 0, 0, 0, 0 );
        nFinal = sat_solver_final( pSat, &pFinal );
        if fVerbose != 0 {
            fmt.Printf("Status %s   Selected %5d assumptions out of %5d.  ", status == l_False ? "OK ":"BUG", nFinal, Vec_IntSize(vLits));
            Abc_PrintTime( 1, "Analyze_final", Abc_Clock() - clk );
        }
        if CountBest > nFinal {
            CountBest = nFinal;
            // inlining of macro ABC_FREE
            ((pCexBest) ? (free((char *) (pCexBest)), (pCexBest) = 0) : 0);
            pCexBest = Bmc_CexCareDeriveCex( pCex, iFirstVar, pFinal, nFinal );
        }
        if !(fHighEffort != 0) {
            continue; }

        // SAT-based minimization
        clk = Abc_Clock();
        nFinal = sat_solver_minimize_assumptions( pSat, Vec_IntArray(vLits), Vec_IntSize(vLits), 0 );
        if fVerbose != 0 {
            fmt.Printf("Status %s   Selected %5d assumptions out of %5d.  ", status == l_False ? "OK ":"BUG", nFinal, Vec_IntSize(vLits));
            Abc_PrintTime( 1, "LEXUNSAT     ", Abc_Clock() - clk );
        }
        if CountBest > nFinal {
            CountBest = nFinal;
            // inlining of macro ABC_FREE
            ((pCexBest) ? (free((char *) (pCexBest)), (pCexBest) = 0) : 0);
            pCexBest = Bmc_CexCareDeriveCex( pCex, iFirstVar, Vec_IntArray(vLits), nFinal );
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Final    :    ");
        Bmc_CexPrint( pCexBest, pCexBest.nPis, 0 );
    }
    // cleanup
    Vec_IntFreeP( &vLits );
    Vec_IntFreeP( &vTemp );
    sat_solver_delete( pSat );
    Cnf_DataFree( pCnf );
    Gia_ManStop( pFrames );
    return pCexBest;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


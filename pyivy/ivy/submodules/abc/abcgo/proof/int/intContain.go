package int

import "fmt"

/**CFile****************************************************************

  FileName    [intContain.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Interpolant containment checking.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 24, 2008.]

  Revision    [$Id: intContain.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "intInt.h"
//#include "proof/fra/fra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int Inter_ManCheckUniqueness( Aig_Man_t * p, sat_solver * pSat, Cnf_Dat_t * pCnf, int nFrames );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checks constainment of two interpolants.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManCheckContainment(pNew *Aig_Man_t, pOld *Aig_Man_t) int32 {
    var pMiter, pAigTemp *Aig_Man_t;
    var RetValue int32;
    pMiter = Aig_ManCreateMiter( pNew, pOld, 1 );
//    pMiter = Dar_ManRwsat( pAigTemp = pMiter, 1, 0 );
//    Aig_ManStop( pAigTemp );
    RetValue = Fra_FraigMiterStatus( pMiter );
    if RetValue == -1 {
        pAigTemp = Fra_FraigEquivence( pMiter, 1000000, 1 );
        RetValue = Fra_FraigMiterStatus( pAigTemp );
        Aig_ManStop( pAigTemp );
//        RetValue = Fra_FraigSat( pMiter, 1000000, 0, 0, 0, 0, 0, 0 );
    }
    assert(RetValue != -1, "RetValue != -1");
    Aig_ManStop( pMiter );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Checks constainment of two interpolants.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManCheckEquivalence(pNew *Aig_Man_t, pOld *Aig_Man_t) int32 {
    var pMiter, pAigTemp *Aig_Man_t;
    var RetValue int32;
    pMiter = Aig_ManCreateMiter( pNew, pOld, 0 );
//    pMiter = Dar_ManRwsat( pAigTemp = pMiter, 1, 0 );
//    Aig_ManStop( pAigTemp );
    RetValue = Fra_FraigMiterStatus( pMiter );
    if RetValue == -1 {
        pAigTemp = Fra_FraigEquivence( pMiter, 1000000, 1 );
        RetValue = Fra_FraigMiterStatus( pAigTemp );
        Aig_ManStop( pAigTemp );
//        RetValue = Fra_FraigSat( pMiter, 1000000, 0, 0, 0, 0, 0, 0 );
    }
    assert(RetValue != -1, "RetValue != -1");
    Aig_ManStop( pMiter );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Create timeframes of the manager for interpolation.]

  Description [The resulting manager is combinational. The primary inputs
  corresponding to register outputs are ordered first.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManFramesLatches(pAig *Aig_Man_t, nFrames int32, pvMapReg **Vec_Ptr_t) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i, f int32;
    assert(Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0");
    pFrames = Aig_ManStart( Aig_ManNodeNum(pAig) * nFrames );
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pFrames );
    // create variables for register outputs
    *pvMapReg = Vec_PtrAlloc( (nFrames+1) * Saig_ManRegNum(pAig) );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.pData = Aig_ObjCreateCi( pFrames );
        Vec_PtrPush( *pvMapReg, pObj.pData );
    }
    // add timeframes
    for  f = 0; f < nFrames; f++  {
        // create PI nodes for this frame
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            pObj.pData = Aig_ObjCreateCi( pFrames );
        }
        // add internal nodes of this frame
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObj.pData = Aig_And( pFrames, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
        }
        // save register inputs
        // inlining of macro Saig_ManForEachLi
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
            pObj.pData = Aig_ObjChild0Copy(pObj);
        }
        // transfer to register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
            pObjLo.pData = pObjLi.pData;
            Vec_PtrPush( *pvMapReg, pObjLo.pData );
        }
    }
    return pFrames;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while mapping PIs into the given array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManAppendCone(pOld *Aig_Man_t, pNew *Aig_Man_t, ppNewPis **Aig_Obj_t, fCompl int32) {
    var pObj *Aig_Obj_t;
    var i int32;
    assert(Aig_ManCoNum(pOld) == 1, "Aig_ManCoNum(pOld) == 1");
    // create the PIs
    Aig_ManCleanData( pOld );
    Aig_ManConst1(pOld).pData = Aig_ManConst1(pNew);
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pOld.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pOld.vCis, i)), 1); i++  {
        pObj.pData = ppNewPis[i];
    }
    // duplicate internal nodes
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pOld.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pOld.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        pObj.pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) ); }
    }
    // add one PO to new
    pObj = Aig_ManCo( pOld, 0 );
    Aig_ObjCreateCo( pNew, Aig_NotCond( Aig_ObjChild0Copy(pObj), fCompl ) );
}


/**Function*************************************************************

  Synopsis    [Checks constainment of two interpolants inductively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManCheckInductiveContainment(pTrans *Aig_Man_t, pInter *Aig_Man_t, nSteps int32, fBackward int32) int32 {
    var pFrames *Aig_Man_t;
    var ppNodes **Aig_Obj_t;
    var vMapRegs *Vec_Ptr_t;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    var f, nRegs, status int32;
    nRegs = Saig_ManRegNum(pTrans);
    assert(nRegs > 0, "nRegs > 0");
    // generate the timeframes 
    pFrames = Inter_ManFramesLatches( pTrans, nSteps, &vMapRegs );
    assert(Vec_PtrSize(vMapRegs) == (nSteps + 1) * nRegs, "Vec_PtrSize(vMapRegs) == (nSteps + 1) * nRegs");
    // add main constraints to the timeframes
    ppNodes = (Aig_Obj_t **)Vec_PtrArray(vMapRegs);
    if !(fBackward != 0) { 
        // forward inductive check: p -> p -> ... -> !p
        for  f = 0; f < nSteps; f++  {
            Inter_ManAppendCone( pInter, pFrames, ppNodes + f * nRegs, 0 );
        }
        Inter_ManAppendCone( pInter, pFrames, ppNodes + f * nRegs, 1 );
    
    } else {
        // backward inductive check: p -> !p -> ... -> !p
        Inter_ManAppendCone( pInter, pFrames, ppNodes + 0 * nRegs, 1 );
        for  f = 1; f <= nSteps; f++  {
            Inter_ManAppendCone( pInter, pFrames, ppNodes + f * nRegs, 0 );
        }
    }
    Vec_PtrFree( vMapRegs );
    Aig_ManCleanup( pFrames );

    // convert to CNF
    pCnf = Cnf_Derive( pFrames, 0 ); 
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
//    Cnf_DataFree( pCnf );
//    Aig_ManStop( pFrames );

    if pSat == nil {
        Cnf_DataFree( pCnf );
        Aig_ManStop( pFrames );
        return 1;
    }

     // solve the problem
    status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );

//    Inter_ManCheckUniqueness( pTrans, pSat, pCnf, nSteps );

    Cnf_DataFree( pCnf );
    Aig_ManStop( pFrames );

    sat_solver_delete( pSat );
    return status == l_False;
}
//ABC_NAMESPACE_IMPL_END

//#include "proof/fra/fra.h"

//ABC_NAMESPACE_IMPL_START


/**Function*************************************************************

  Synopsis    [Check if cex satisfies uniqueness constraints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManCheckUniqueness(p *Aig_Man_t, pSat *sat_solver, pCnf *Cnf_Dat_t, nFrames int32) int32 {
    extern int Fra_SmlNodesCompareInFrame( Fra_Sml_t * p, Aig_Obj_t * pObj0, Aig_Obj_t * pObj1, int iFrame0, int iFrame1 );
    extern void Fra_SmlAssignConst( Fra_Sml_t * p, Aig_Obj_t * pObj, int fConst1, int iFrame );
    extern void Fra_SmlSimulateOne( Fra_Sml_t * p );

    var pSml *Fra_Sml_t;
    var vPis *Vec_Int_t;
    var pObj, pObj0 *Aig_Obj_t;
    var i int32;
    var k int32;
    var v int32;
    var iBit int32;
    var pCounterEx *int32;
    var Counter int32;
    if nFrames == 1 {
        return 1; }
//    if ( pSat->model.size == 0 )
    
    // possible consequences here!!!
    assert(false, "0");

    if sat_solver_nvars(pSat) == 0 {
        return 1; }
//    assert( Saig_ManPoNum(p) == 1 );
    assert(Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0");
    assert(Aig_ManRegNum(p) < Aig_ManCiNum(p), "Aig_ManRegNum(p) < Aig_ManCiNum(p)");

    // get the counter-example
    vPis = Vec_IntAlloc( 100 );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(pCnf.pMan.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pCnf.pMan.vCis, k)), 1); k++  {
        Vec_IntPush( vPis, pCnf.pVarNums[Aig_ObjId(pObj)] );
    }
    assert(Vec_IntSize(vPis) == Aig_ManRegNum(p) + nFrames * Saig_ManPiNum(p), "Vec_IntSize(vPis) == Aig_ManRegNum(p) + nFrames * Saig_ManPiNum(p)");
    pCounterEx = Sat_SolverGetModel( pSat, vPis.pArray, vPis.nSize );
    Vec_IntFree( vPis );

    // start a new sequential simulator
    pSml = Fra_SmlStart( p, 0, nFrames, 1 );
    // assign simulation info for the registers
    iBit = 0;
    // inlining of macro Aig_ManForEachLoSeq
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCiNum(p)-Aig_ManRegNum(p); (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Fra_SmlAssignConst( pSml, pObj, pCounterEx[iBit++], 0 );
    }
    // assign simulation info for the primary inputs
    for  i = 0; i < nFrames; i++  {
        // inlining of macro Aig_ManForEachPiSeq
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < Aig_ManCiNum(p)-Aig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, k)), 1); k++  {
            Fra_SmlAssignConst( pSml, pObj, pCounterEx[iBit++], i );
        }
    }
    assert(iBit == Aig_ManCiNum(pCnf.pMan), "iBit == Aig_ManCiNum(pCnf.pMan)");
    // run simulation
    Fra_SmlSimulateOne( pSml );

    // check if the given output has failed
//    RetValue = !Fra_SmlNodeIsZero( pSml, Aig_ManCo(pAig, 0) );
//    assert( RetValue );

    // check values at the internal nodes
    Counter = 0;
    for  i = 0; i < nFrames; i++  {
    for  k = i+1; k < nFrames; k++  {
        for  v = 0; v < Aig_ManRegNum(p); v++  {
            pObj0 = Aig_ManLo(p, v);
            if !Fra_SmlNodesCompareInFrame( pSml, pObj0, pObj0, i, k ) {
                break; }
        }
        if v == Aig_ManRegNum(p) {
            Counter++; }
    }
    }
    fmt.Printf("Uniquness does not hold in %d frames.\n", Counter);

    Fra_SmlStop( pSml );
    // inlining of macro ABC_FREE
    ((pCounterEx) ? (free((char *) (pCounterEx)), (pCounterEx) = 0) : 0);
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


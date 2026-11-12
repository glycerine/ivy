package int

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [intCtrex.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Counter-example generation after disproving the property.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 24, 2008.]

  Revision    [$Id: intCtrex.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "intInt.h"
//#include "proof/ssw/ssw.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Unroll the circuit the given number of timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManFramesBmc(pAig *Aig_Man_t, nFrames int32) *Aig_Man_t {
    var pFrames *Aig_Man_t;
    var pObj, pObjLi, pObjLo *Aig_Obj_t;
    var i, f int32;
    assert(Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0");
    assert(Saig_ManPoNum(pAig) == 1, "Saig_ManPoNum(pAig) == 1");
    pFrames = Aig_ManStart( Aig_ManNodeNum(pAig) * nFrames );
    // map the constant node
    Aig_ManConst1(pAig).pData = Aig_ManConst1( pFrames );
    // create variables for register outputs
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        pObj.pData = Aig_ManConst0( pFrames );
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
        if f == nFrames - 1 {
            break; }
        // transfer to register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
            pObjLi.pData = Aig_ObjChild0Copy(pObjLi);
        }
        // transfer to register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObjLi) = Saig_ManLi(pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(pAig, i)), 1); i++  {
            pObjLo.pData = pObjLi.pData;
        }
    }
    // create POs for the output of the last frame
    pObj = Aig_ManCo( pAig, 0 );
    Aig_ObjCreateCo( pFrames, Aig_ObjChild0Copy(pObj) );
    Aig_ManCleanup( pFrames );
    return pFrames;
}

/**Function*************************************************************

  Synopsis    [Run the SAT solver on the unrolled instance.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inter_ManGetCounterExample(pAig *Aig_Man_t, nFrames int32, fVerbose int32) any {
    var nConfLimit int32  = 1000000;
    var pCtrex *Abc_Cex_t  = nil;
    var pFrames *Aig_Man_t;
    var pSat *sat_solver;
    var pCnf *Cnf_Dat_t;
    var status int32;
    var clk abctime  = Abc_Clock();
    var vCiIds *Vec_Int_t;
    // create timeframes
    assert(Saig_ManPoNum(pAig) == 1, "Saig_ManPoNum(pAig) == 1");
    pFrames = Inter_ManFramesBmc( pAig, nFrames );
    // derive CNF
    pCnf = Cnf_Derive( pFrames, 0 );
    Cnf_DataTranformPolarity( pCnf, 0 );
    vCiIds = Cnf_DataCollectPiSatNums( pCnf, pFrames );
    Aig_ManStop( pFrames );
    // convert into SAT solver
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    Cnf_DataFree( pCnf );
    if pSat == nil {
        fmt.Printf("Counter-example generation in command \"int\" has failed.\n");
        fmt.Printf("Use command \"bmc2\" to produce a valid counter-example.\n");
        Vec_IntFree( vCiIds );
        return nil;
    }
    // simplify the problem
    status = sat_solver_simplify(pSat);
    if status == 0 {
        Vec_IntFree( vCiIds );
        sat_solver_delete( pSat );
        return nil;
    }
    // solve the miter
    status = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    // if the problem is SAT, get the counterexample
    if status == l_True {
        var i int32;
        var pModel *int32 = Sat_SolverGetModel( pSat, vCiIds.pArray, vCiIds.nSize );
        pCtrex = Abc_CexAlloc( Saig_ManRegNum(pAig), Saig_ManPiNum(pAig), nFrames );
        pCtrex.iFrame = nFrames - 1;
        pCtrex.iPo = 0;
        for  i = 0; i < Vec_IntSize(vCiIds); i++  {
            if pModel[i] {
                Abc_InfoSetBit( pCtrex.pData, Saig_ManRegNum(pAig) + i ); }
        }
        pModel = nil;
    }
    // free the sat_solver
    sat_solver_delete( pSat );
    Vec_IntFree( vCiIds );
    // verify counter-example
    status = Saig_ManVerifyCex( pAig, pCtrex );
    if status == 0 {
        fmt.Printf("Inter_ManGetCounterExample(): Counter-example verification has FAILED.\n"); }
    // report the results
    if fVerbose != 0 {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Total ctrex generation time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    return pCtrex;

}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


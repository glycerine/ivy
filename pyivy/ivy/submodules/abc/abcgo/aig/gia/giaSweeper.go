package gia

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [giaSweeper.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Incremental SAT sweeper.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSweeper.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "base/main/main.h"
//#include "sat/bsat/satSolver.h"
//#include "proof/ssc/ssc.h"

//ABC_NAMESPACE_IMPL_START

/*

SAT sweeping/equivalence checking requires the following steps:
- Creating probes
  These APIs should be called for all internal points in the logic, which may be used as
      - nodes representing conditions to be used as constraints
      - nodes representing functions to be equivalence checked
      - nodes representing functions needed by the user at the end of SAT sweeping
  Creating new probe using  Gia_SweeperProbeCreate():   int Gia_SweeperProbeCreate( Gia_Man_t * p, int iLit );
  Delete existing probe using Gia_SweeperProbeDelete(): int Gia_SweeperProbeDelete( Gia_Man_t * p, int ProbeId );
  Update existing probe using Gia_SweeperProbeUpdate(): int Gia_SweeperProbeUpdate( Gia_Man_t * p, int ProbeId, int iLit );
  Comments:
      - a probe is identified by its 0-based ID, which is returned by above procedures
      - GIA literal of the probe is returned by       int Gia_SweeperProbeLit( Gia_Man_t * p, int ProbeId )
- Adding/removing conditions on the current path by calling Gia_SweeperCondPush() and Gia_SweeperCondPop()
      extern void Gia_SweeperCondPush( Gia_Man_t * p, int ProbeId );
      extern void Gia_SweeperCondPop( Gia_Man_t * p );
- Performing equivalence checking by calling     int Gia_SweeperCheckEquiv( Gia_Man_t * pGia, int Probe1, int Probe2 )
      (resource limits, such as the number of conflicts, will be controllable by dedicated GIA APIs)
- The resulting AIG to be returned to the user by calling Gia_SweeperExtractUserLogic()
      Gia_Man_t * Gia_SweeperExtractUserLogic( Gia_Man_t * p, Vec_Int_t * vProbeIds, Vec_Ptr_t * vOutNames )

*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
/**Function*************************************************************

  Synopsis    [Sweeper cleanup.]

  Description [Returns new GIA with sweeper defined, which is the same
  as the original sweeper, with all the dangling logic removed and SAT 
  solver restarted. The probe IDs are guaranteed to have the same logic
  functions as in the original manager.]
               
  SideEffects [The input manager is deleted inside this procedure.]

  SeeAlso     []

***********************************************************************/
func Gia_SweeperCleanup(p *Gia_Man_t, pCommLime string) *Gia_Man_t {
    var pSwp *Swp_Man_t  = (Swp_Man_t *)p.pData;
    var vObjIds *Vec_Int_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, iLit, ProbeId int32;

    // collect all internal nodes pointed to by currently-used probes
    Gia_ManIncrementTravId( p );
    vObjIds = Vec_IntAlloc( 1000 );
    // inlining of macro Vec_IntForEachEntry
    for  ProbeId = 0; (ProbeId < Vec_IntSize(pSwp.vProbes)) && (((iLit) = Vec_IntEntry(pSwp.vProbes, ProbeId)), 1); ProbeId++  {
        if iLit < 0 { continue; }
        pObj = Gia_Lit2Obj( p, iLit );
        Gia_ManExtract_rec( p, Gia_Regular(pObj), vObjIds );
    }
    // create new manager
    pNew = Gia_ManStart( 1 + Gia_ManPiNum(p) + Vec_IntSize(vObjIds) + 100 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // create internal nodes
    Gia_ManHashStart( pNew );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjIds)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjIds,i))); i++  {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    // create outputs
    // inlining of macro Vec_IntForEachEntry
    for  ProbeId = 0; (ProbeId < Vec_IntSize(pSwp.vProbes)) && (((iLit) = Vec_IntEntry(pSwp.vProbes, ProbeId)), 1); ProbeId++  {
        if iLit < 0 { continue; }
        pObj = Gia_Lit2Obj( p, iLit );
        iLit = Gia_Regular(pObj).Value ^ Gia_IsComplement(pObj);
        Vec_IntWriteEntry( pSwp.vProbes, ProbeId, iLit );
    }
    Vec_IntFree( vObjIds );
    // duplicate if needed
    if Gia_ManHasDangling(pNew) {
        pNew = Gia_ManCleanup( pTemp = pNew );
        Gia_ManStop( pTemp );
    }
    // execute command line
    if pCommLime {
        // set pNew to be current GIA in ABC
        Abc_FrameUpdateGia( Abc_FrameGetGlobalFrame(), pNew );
        // execute command line pCommLine using Abc_CmdCommandExecute()
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), pCommLime );
        // get pNew to be current GIA in ABC     
        pNew = Abc_FrameGetGia( Abc_FrameGetGlobalFrame() );
    }
    // restart the SAT solver
    Vec_IntClear( pSwp.vId2Lit );
    sat_solver_delete( pSwp.pSat );
    pSwp.pSat         = sat_solver_new();
    pSwp.nSatVars     = 1;
    sat_solver_setnvars( pSwp.pSat, 1000 );
    Swp_ManSetObj2Lit( pSwp, 0, (iLit = Abc_Var2Lit(pSwp.nSatVars++, 0)) );
    iLit = Abc_LitNot(iLit);
    sat_solver_addclause( pSwp.pSat, &iLit, &iLit + 1 );
    pSwp.timeStart    = Abc_Clock();
    // return the result
    pNew.pData = p.pData; p.pData = nil;
    Gia_ManStop( p );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Addes clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Runs equivalence test for probes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SweeperCheckEquiv(pGia *Gia_Man_t, Probe1 int32, Probe2 int32) int32 {
    var p *Swp_Man_t  = (Swp_Man_t *)pGia.pData;
    var iLitOld int32;
    var iLitNew int32;
    var iLitAig int32;
    var pLitsSat [2]int32;
    var RetValue int32;
    var RetValue1 int32;
    var ProbeId int32;
    var i int32;
    var clk abctime;
    p.nSatCalls++;
    assert(p.pSat != nil, "p.pSat != nil");
    p.vCexUser = nil;

    // get the literals
    iLitOld = Gia_SweeperProbeLit( pGia, Probe1 );
    iLitNew = Gia_SweeperProbeLit( pGia, Probe2 );
    // if the literals are identical, the probes are equivalent
    if iLitOld == iLitNew {
        return 1; }
    // if the literals are opposites, the probes are not equivalent
    if Abc_LitRegular(iLitOld) == Abc_LitRegular(iLitNew) {
        Vec_IntFill( p.vCexSwp, Gia_ManPiNum(pGia), 2 );
        p.vCexUser = p.vCexSwp;
        return 0;
    }
    // order the literals
    if iLitOld < iLitNew {
        // inlining of macro ABC_SWAP
         var t int32  = iLitOld; iLitOld = iLitNew; iLitNew = t; };
    assert(iLitOld > iLitNew, "iLitOld > iLitNew");

    // create logic cones and the array of assumptions
    Vec_IntClear( p.vCondAssump );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCondProbes)) && (((ProbeId) = Vec_IntEntry(p.vCondProbes, i)), 1); i++  {
        iLitAig = Gia_SweeperProbeLit( pGia, ProbeId );
        Gia_ManCnfNodeAddToSolver( p, Abc_Lit2Var(iLitAig) );
        Vec_IntPush( p.vCondAssump, Abc_LitNot(Swp_ManLit2Lit(p, iLitAig)) );
    }
    Gia_ManCnfNodeAddToSolver( p, Abc_Lit2Var(iLitOld) );
    Gia_ManCnfNodeAddToSolver( p, Abc_Lit2Var(iLitNew) );
    sat_solver_compress( p.pSat );

    // set the SAT literals
    pLitsSat[0] = Swp_ManLit2Lit( p, iLitOld );
    pLitsSat[1] = Swp_ManLit2Lit( p, iLitNew );

    // solve under assumptions
    // A = 1; B = 0     OR     A = 1; B = 1 
    Vec_IntPush( p.vCondAssump, pLitsSat[0] );
    Vec_IntPush( p.vCondAssump, Abc_LitNot(pLitsSat[1]) );

    // set runtime limit for this call
    if p.nTimeOut {
        sat_solver_set_runtime_limit( p.pSat, p.nTimeOut * int64(time.Second) + Abc_Clock() ); }

clk = Abc_Clock();
    RetValue1 = sat_solver_solve( p.pSat, Vec_IntArray(p.vCondAssump), Vec_IntArray(p.vCondAssump) + Vec_IntSize(p.vCondAssump), 
        (ABC_INT64_T)p.nConfMax, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    Vec_IntShrink( p.vCondAssump, Vec_IntSize(p.vCondAssump) - 2 );
p.timeSat += Abc_Clock() - clk;
    if RetValue1 == l_False {
        pLitsSat[0] = Abc_LitNot( pLitsSat[0] );
        RetValue = sat_solver_addclause( p.pSat, pLitsSat, pLitsSat + 2 );
        assert((RetValue != 0), "RetValue");
        pLitsSat[0] = Abc_LitNot( pLitsSat[0] );
p.timeSatUnsat += Abc_Clock() - clk;
        p.nSatCallsUnsat++;
    
    } else if RetValue1 == l_True {
        p.vCexUser = Gia_ManGetCex( p.pGia, p.vId2Lit, p.pSat, p.vCexSwp );
p.timeSatSat += Abc_Clock() - clk;
        p.nSatCallsSat++;
        return 0;
    
    } else { // if ( RetValue1 == l_Undef )
    
p.timeSatUndec += Abc_Clock() - clk;
        p.nSatCallsUndec++;
        return -1;
    }

    // if the old node was constant 0, we already know the answer
    if Gia_ManIsConstLit(iLitNew) {
        p.nSatProofs++;
        return 1;
    }

    // solve under assumptions
    // A = 0; B = 1     OR     A = 0; B = 0 
    Vec_IntPush( p.vCondAssump, Abc_LitNot(pLitsSat[0]) );
    Vec_IntPush( p.vCondAssump, pLitsSat[1] );

clk = Abc_Clock();
    RetValue1 = sat_solver_solve( p.pSat, Vec_IntArray(p.vCondAssump), Vec_IntArray(p.vCondAssump) + Vec_IntSize(p.vCondAssump), 
        (ABC_INT64_T)p.nConfMax, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    Vec_IntShrink( p.vCondAssump, Vec_IntSize(p.vCondAssump) - 2 );
p.timeSat += Abc_Clock() - clk;
    if RetValue1 == l_False {
        pLitsSat[1] = Abc_LitNot( pLitsSat[1] );
        RetValue = sat_solver_addclause( p.pSat, pLitsSat, pLitsSat + 2 );
        assert((RetValue != 0), "RetValue");
        pLitsSat[1] = Abc_LitNot( pLitsSat[1] );
p.timeSatUnsat += Abc_Clock() - clk;
        p.nSatCallsUnsat++;
    
    } else if RetValue1 == l_True {
        p.vCexUser = Gia_ManGetCex( p.pGia, p.vId2Lit, p.pSat, p.vCexSwp );
p.timeSatSat += Abc_Clock() - clk;
        p.nSatCallsSat++;
        return 0;
    
    } else { // if ( RetValue1 == l_Undef )
    
p.timeSatUndec += Abc_Clock() - clk;
        p.nSatCallsUndec++;
        return -1;
    }
    // return SAT proof
    p.nSatProofs++;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the set of conditions is UNSAT (0 if SAT; -1 if undecided).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SweeperCondCheckUnsat(pGia *Gia_Man_t) int32 {
    var p *Swp_Man_t  = (Swp_Man_t *)pGia.pData;
    var RetValue, ProbeId, iLitAig, i int32;
    var clk abctime;
    assert(p.pSat != nil, "p.pSat != nil");
    p.nSatCalls++;
    p.vCexUser = nil;

    // create logic cones and the array of assumptions
    Vec_IntClear( p.vCondAssump );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCondProbes)) && (((ProbeId) = Vec_IntEntry(p.vCondProbes, i)), 1); i++  {
        iLitAig = Gia_SweeperProbeLit( pGia, ProbeId );
        Gia_ManCnfNodeAddToSolver( p, Abc_Lit2Var(iLitAig) );
        Vec_IntPush( p.vCondAssump, Abc_LitNot(Swp_ManLit2Lit(p, iLitAig)) );
    }
    sat_solver_compress( p.pSat );

    // set runtime limit for this call
    if p.nTimeOut {
        sat_solver_set_runtime_limit( p.pSat, p.nTimeOut * int64(time.Second) + Abc_Clock() ); }

clk = Abc_Clock();
    RetValue = sat_solver_solve( p.pSat, Vec_IntArray(p.vCondAssump), Vec_IntArray(p.vCondAssump) + Vec_IntSize(p.vCondAssump), 
        (ABC_INT64_T)p.nConfMax, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
p.timeSat += Abc_Clock() - clk;
    if RetValue == l_False {
        assert(Vec_IntSize(p.vCondProbes) > 0, "Vec_IntSize(p.vCondProbes) > 0");
p.timeSatUnsat += Abc_Clock() - clk;
        p.nSatCallsUnsat++;
        p.nSatProofs++;
        return 1;
    
    } else if RetValue == l_True {
        p.vCexUser = Gia_ManGetCex( p.pGia, p.vId2Lit, p.pSat, p.vCexSwp );
p.timeSatSat += Abc_Clock() - clk;
        p.nSatCallsSat++;
        return 0;
    
    } else { // if ( RetValue1 == l_Undef )
    
p.timeSatUndec += Abc_Clock() - clk;
        p.nSatCallsUndec++;
        return -1;
    }
}

/**Function*************************************************************

  Synopsis    [Performs grafting from another manager.]

  Description [Returns the array of resulting literals in the destination manager.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Performs conditional sweeping of the cone.]

  Description [Returns the result as a new GIA manager with as many inputs 
  as the original manager and as many outputs as there are logic cones.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SweeperSweep(p *Gia_Man_t, vProbeOuts *Vec_Int_t, nWords int32, nConfs int32, fVerify int32, fVerbose int32) *Gia_Man_t {
    var vProbeConds *Vec_Int_t;
    var pGiaCond, pGiaOuts, pGiaRes *Gia_Man_t;
    var Pars Ssc_Pars_t;
    var pPars *Ssc_Pars_t = &Pars;
    Ssc_ManSetDefaultParams( pPars );
    pPars.nWords   = nWords;
    pPars.nBTLimit = nConfs;
    pPars.fVerify  = fVerify;
    pPars.fVerbose = fVerbose;
    // sweeper is running
    assert((Gia_SweeperIsRunning(p) != 0), "Gia_SweeperIsRunning(p)");
    // extract conditions and logic cones
    vProbeConds = Gia_SweeperCondVector( p );
    pGiaCond = Gia_SweeperExtractUserLogic( p, vProbeConds, nil, nil );
    pGiaOuts = Gia_SweeperExtractUserLogic( p, vProbeOuts, nil, nil );
    Gia_ManSetPhase( pGiaOuts );
    // if there is no conditions, define constant true constraint (constant 0 output)
    if Gia_ManPoNum(pGiaCond) == 0 {
        Gia_ManAppendCo( pGiaCond, Gia_ManConst0Lit() ); }
    // perform sweeping under constraints
    pGiaRes = Ssc_PerformSweeping( pGiaOuts, pGiaCond, pPars );
    Gia_ManStop( pGiaCond );
    Gia_ManStop( pGiaOuts );
    return pGiaRes;
}

/**Function*************************************************************

  Synopsis    [Procedure to perform conditional fraig sweeping on separate logic cones.]

  Description [The procedure takes GIA with the sweeper defined. The sweeper
  is assumed to have some conditions currently pushed, which will be used
  as constraints for fraig sweeping. The second argument (vProbes) contains
  the array of probe IDs pointing to the user's logic cones to be SAT swept.
  Finally, the optional command line (pCommLine) is an ABC command line
  to be applied to the resulting GIA after SAT sweeping before it is 
  grafted back into the original GIA manager. The return value is the status
  (success/failure) and the array of original probes possibly pointing to the
  new literals in the original GIA manager, corresponding to the user's
  logic cones after sweeping, synthesis and grafting.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SweeperFraig(p *Gia_Man_t, vProbeIds *Vec_Int_t, pCommLime string, nWords int32, nConfs int32, fVerify int32, fVerbose int32) int32 {
    var pNew *Gia_Man_t;
    var vLits *Vec_Int_t;
    var ProbeId, i int32;
    // sweeper is running
    assert((Gia_SweeperIsRunning(p) != 0), "Gia_SweeperIsRunning(p)");
    // sweep the logic
    pNew = Gia_SweeperSweep( p, vProbeIds, nWords, nConfs, fVerify, fVerbose );
    if pNew == nil {
        return 0; }
    // execute command line
    if pCommLime {
        // set pNew to be current GIA in ABC
        Abc_FrameUpdateGia( Abc_FrameGetGlobalFrame(), pNew );
        // execute command line pCommLine using Abc_CmdCommandExecute()
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), pCommLime );
        // get pNew to be current GIA in ABC     
        pNew = Abc_FrameGetGia( Abc_FrameGetGlobalFrame() );
    }
    // return logic back into the main manager
    vLits = Gia_SweeperGraft( p, nil, pNew );
    Gia_ManStop( pNew );
    // update the array of probes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vProbeIds)) && (((ProbeId) = Vec_IntEntry(vProbeIds, i)), 1); i++  {
        Gia_SweeperProbeUpdate( p, ProbeId, Vec_IntEntry(vLits, i) );
    }
    Vec_IntFree( vLits );        
    return 1;
}

/**Function*************************************************************

  Synopsis    [Executes given command line for the logic defined by the probes.]

  Description [ ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SweeperRun(p *Gia_Man_t, vProbeIds *Vec_Int_t, pCommLime string, fVerbose int32) int32 {
    var pNew *Gia_Man_t;
    var vLits *Vec_Int_t;
    var ProbeId, i int32;
    // sweeper is running
    assert((Gia_SweeperIsRunning(p) != 0), "Gia_SweeperIsRunning(p)");
    // sweep the logic
    pNew = Gia_SweeperExtractUserLogic( p, vProbeIds, nil, nil );
    // execute command line
    if pCommLime {
        if fVerbose != 0 {
            fmt.Printf("GIA manager statistics before and after applying \"%s\":\n", pCommLime); }
        if fVerbose != 0 {
            Gia_ManPrintStats( pNew, nil ); }
        // set pNew to be current GIA in ABC
        Abc_FrameUpdateGia( Abc_FrameGetGlobalFrame(), pNew );
        // execute command line pCommLine using Abc_CmdCommandExecute()
        Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), pCommLime );
        // get pNew to be current GIA in ABC     
        pNew = Abc_FrameGetGia( Abc_FrameGetGlobalFrame() );
        if fVerbose != 0 {
            Gia_ManPrintStats( pNew, nil ); }
    }
    // return logic back into the main manager
    vLits = Gia_SweeperGraft( p, nil, pNew );
    Gia_ManStop( pNew );
    // update the array of probes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vProbeIds)) && (((ProbeId) = Vec_IntEntry(vProbeIds, i)), 1); i++  {
        Gia_SweeperProbeUpdate( p, ProbeId, Vec_IntEntry(vLits, i) );
    }
    Vec_IntFree( vLits );        
    return 1;
}

/**Function*************************************************************

  Synopsis    [Sweeper sweeper test.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SweeperFraigTest(pInit *Gia_Man_t, nWords int32, nConfs int32, fVerbose int32) *Gia_Man_t {
    var p, pGia *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vOuts *Vec_Int_t;
    var i int32;
    // add one-hotness constraints
    p = Gia_ManDupOneHot( pInit );
    // create sweeper
    Gia_SweeperStart( p );
    // collect outputs and create conditions
    vOuts = Vec_IntAlloc( Gia_ManPoNum(p) );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if i < Gia_ManPoNum(p) - p.nConstrs {  // this is the user's output
            Vec_IntPush( vOuts, Gia_SweeperProbeCreate( p,  Gia_ObjFaninLit0p(p, pObj) ) );
        } else { // this is a constraint
            Gia_SweeperCondPush( p, Gia_SweeperProbeCreate( p, Gia_ObjFaninLit0p(p, pObj) ) ); }
    }
    // perform the sweeping
    pGia = Gia_SweeperSweep( p, vOuts, nWords, nConfs, fVerbose, 0 );
//    pGia = Gia_ManDup( p );
    Vec_IntFree( vOuts );
    // sop the sweeper
    Gia_SweeperStop( p );
    Gia_ManStop( p );
    return pGia;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END


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


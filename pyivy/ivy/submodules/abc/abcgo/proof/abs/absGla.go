package abs

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [absGla2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Scalable gate-level abstraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absGla2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/main/main.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver2.h"
//#include "bool/kit/kit.h"
//#include "abs.h"
//#include "absRef.h"
//#include "absRef2.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


    // manager
////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes truth table for the marked node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Returns AIG marked for CNF generation.]

  Description [The marking satisfies the following requirements:
  Each marked node has the number of marked fanins no more than N.]
               
  SideEffects [Uses pObj->fPhase to store the markings.]

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Computes a minimized truth table.]

  Description [Input literals can be 0/1 (const 0/1), non-trivial literals 
  (integers that are more than 1) and unassigned literals (large integers).
  This procedure computes the truth table that essentially depends on input
  variables ordered in the increasing order of their positive literals.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ga2_ObjComputeTruthSpecial_uTruth5 [5]uint32 = [5]uint32{ 0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000 };
/**Function*************************************************************

  Synopsis    [Returns CNF of the function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Derives CNF for one node.]

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
var spStatic_Saig_ManBmcHashKey_s_Primes [5]int32 = [5]int32{ 12582917, 25165843, 50331653, 100663319, 201326611 };
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

  Synopsis    [Creates a new manager.]

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
func Gia_ManPerformGla(pAig *Gia_Man_t, pPars *Abs_Par_t) int32 {
    var fUseSecondCore int32  = 1;
    var p *Ga2_Man_t;
    var vCore, vPPis *Vec_Int_t;
    var clk2 abctime;
    var clk abctime = Abc_Clock();
    var Status int32 = l_Undef;
    var RetValue int32 = -1;
    var iFrameTryToProve int32 = -1;
    var fOneIsSent int32;
    var i, c, f, Lit int32;
    pPars.iFrame = -1;
    // check trivial case 
    assert(Gia_ManPoNum(pAig) == 1, "Gia_ManPoNum(pAig) == 1"); 
    ABC_FREE( pAig.pCexSeq );
    if Gia_ObjIsConst0(Gia_ObjFanin0(Gia_ManPo(pAig,0))) {
        if !Gia_ObjFaninC0(Gia_ManPo(pAig,0)) {
            Abc_Print( 1, "Sequential miter is trivially UNSAT.\n" );
            return 1;
        }
        pAig.pCexSeq = Abc_CexMakeTriv( Gia_ManRegNum(pAig), Gia_ManPiNum(pAig), 1, 0 );
        Abc_Print( 1, "Sequential miter is trivially SAT.\n" );
        return 0;
    }
    // create gate classes if not given
    if pAig.vGateClasses == nil {
        pAig.vGateClasses = Vec_IntStart( Gia_ManObjNum(pAig) );
        Vec_IntWriteEntry( pAig.vGateClasses, 0, 1 );
        Vec_IntWriteEntry( pAig.vGateClasses, Gia_ObjFaninId0p(pAig, Gia_ManPo(pAig, 0)), 1 );
    }
    // start the manager
    p = Ga2_ManStart( pAig, pPars );
    p.timeInit = Abc_Clock() - clk;
    // perform initial abstraction
    if p.pPars.fVerbose {
        Abc_Print( 1, "Running gate-level abstraction (GLA) with the following parameters:\n" );
        Abc_Print( 1, "FrameMax = %d  ConfMax = %d  Timeout = %d  Limit = %d %%  Limit2 = %d %%  RatioMax = %d %%\n", 
            pPars.nFramesMax, pPars.nConfLimit, pPars.nTimeOut, pPars.nRatioMin, pPars.nRatioMin2, pPars.nRatioMax );
        Abc_Print( 1, "LrnStart = %d  LrnDelta = %d  LrnRatio = %d %%  Skip = %d  SimpleCNF = %d  Dump = %d\n", 
            pPars.nLearnedStart, pPars.nLearnedDelta, pPars.nLearnedPerce, pPars.fUseSkip, pPars.fUseSimple, pPars.fDumpVabs|pPars.fDumpMabs );
        if pPars.fDumpVabs || pPars.fDumpMabs {
            Abc_Print( 1, "%s will be continously dumped into file \"%s\".\n", 
                pPars.fDumpVabs ? "Abstracted model" : "Miter with abstraction map",
                Ga2_GlaGetFileName(p, pPars.fDumpVabs) ); }
        if pPars.fDumpMabs {
            {
                var Command string // was [1000]byte in C

                Abc_FrameSetStatus( -1 );
                Abc_FrameSetCex( nil );
                Abc_FrameSetNFrames( -1 );
                Command = fmt.Sprintf("write_status %s", Extra_FileNameGenericAppend((char *)(p.pPars.pFileVabs ? p.pPars.pFileVabs : "glabs.aig"), ".status"));
                Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
            }
            {
                // create trivial abstraction map
                var pObj *Gia_Obj_t; 
                var pFileName string  = Ga2_GlaGetFileName(p, 0);
                var vMap *Vec_Int_t  = pAig.vGateClasses; pAig.vGateClasses = nil;
                pAig.vGateClasses = Vec_IntStart( Gia_ManObjNum(pAig) );
                Vec_IntWriteEntry( pAig.vGateClasses, 0, 1 );
                Gia_ManForEachAnd( pAig, pObj, i )
                    Vec_IntWriteEntry( pAig.vGateClasses, i, 1 );
                Gia_ManForEachRo( pAig, pObj, i )
                    Vec_IntWriteEntry( pAig.vGateClasses, Gia_ObjId(pAig, pObj), 1 );
                Gia_AigerWrite( pAig, pFileName, 0, 0, 0 );
                Vec_IntFree( pAig.vGateClasses );
                pAig.vGateClasses = vMap;
                if p.pPars.fVerbose {
                    Abc_Print( 1, "Dumping miter with abstraction map into file \"%s\"...\n", pFileName ); }
            }
        }
        Abc_Print( 1, " Frame   %%   Abs  PPI   FF   LUT   Confl  Cex   Vars   Clas   Lrns     Time        Mem\n" );
    }
    // iterate unrolling
    for  i = f = 0; !pPars.nFramesMax || f < pPars.nFramesMax; i++  {
        var nAbsOld int32;
        // remember the timeframe
        p.pPars.iFrame = -1;
        // create new SAT solver
        Ga2_ManRestart( p );
        // remember abstraction size after the last restart
        nAbsOld = Vec_IntSize(p.vAbs);
        // unroll the circuit
        for  f = 0; !pPars.nFramesMax || f < pPars.nFramesMax; f++  {
            // remember current limits
            var nConflsBeg int32  = sat_solver2_nconflicts(p.pSat);
            var nAbs int32        = Vec_IntSize(p.vAbs);
            var nValues int32     = Vec_IntSize(p.vValues);
            var nVarsOld int32;
            // remember the timeframe
            p.pPars.iFrame = f;
            // extend and clear storage
            if f == Vec_PtrSize(p.vId2Lit) {
                Vec_PtrPush( p.vId2Lit, Vec_IntAlloc(0) ); }
            Vec_IntFillExtra( Ga2_MapFrameMap(p, f), Vec_IntSize(p.vValues), -1 );
            // add static clauses to this timeframe
            Ga2_ManAddAbsClauses( p, f );
            // skip checking if skipcheck is enabled (&gla -s)
            if p.pPars.fUseSkip && f <= p.pPars.iFrameProved {
                continue; }
            // skip checking if we need to skip several starting frames (&gla -S <num>)
            if p.pPars.nFramesStart && f <= p.pPars.nFramesStart {
                continue; }
            // get the output literal
//            Lit = Ga2_ManUnroll_rec( p, Gia_ManPo(pAig,0), f );
            Lit = Ga2_ObjFindLit( p, Gia_ObjFanin0(Gia_ManPo(pAig,0)), f );
            assert(Lit >= 0, "Lit >= 0");
            Lit = Abc_LitNotCond( Lit, Gia_ObjFaninC0(Gia_ManPo(pAig,0)) );
            if Lit == 0 {
                continue; }
            assert(Lit > 1, "Lit > 1");
            // check for counter-examples
            if p.nSatVars > sat_solver2_nvars(p.pSat) {
                sat_solver2_setnvars( p.pSat, p.nSatVars ); }
            nVarsOld = p.nSatVars;
            for  c = 0; ; c++  {
                // consider the special case when the target literal is implied false
                // by implications which happened as a result of previous refinements
                // note that incremental UNSAT core cannot be computed because there is no learned clauses
                // in this case, we will assume that UNSAT core cannot reduce the problem
                if var_is_assigned(p.pSat, Abc_Lit2Var(Lit)) {
                    Prf_ManStopP( &p.pSat.pPrf2 );
                    break;
                }
                // perform SAT solving
                clk2 = Abc_Clock();
                Status = sat_solver2_solve( p.pSat, &Lit, &Lit+1, (ABC_INT64_T)pPars.nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
                if Status == l_True { // perform refinement
                
                    p.nCexes++;
                    p.timeSat += Abc_Clock() - clk2;

                    // cancel old one if it was sent
                    if Abc_FrameIsBridgeMode() && (fOneIsSent != 0) {
                        Gia_Ga2SendCancel( p, pPars.fVerbose );
                        fOneIsSent = 0;
                    }
                    if iFrameTryToProve >= 0 {
                        Gia_GlaProveCancel( pPars.fVerbose );
                        iFrameTryToProve = -1;
                    }

                    // perform refinement
                    clk2 = Abc_Clock();
                    Rnm_ManSetRefId( p.pRnm, c );
                    vPPis = Ga2_ManRefine( p );
                    p.timeCex += Abc_Clock() - clk2;
                    if vPPis == nil {
                        if pPars.fVerbose {
                            Ga2_ManAbsPrintFrame( p, f, sat_solver2_nconflicts(p.pSat)-nConflsBeg, c, Abc_Clock() - clk, 1 ); }
                        goto finish;
                    }

                    if c == 0 {
//                        Ga2_ManRefinePrintPPis( p );
                        // create bookmark to be used for rollback
                        assert(nVarsOld == p.pSat.size, "nVarsOld == p.pSat.size");
                        sat_solver2_bookmark( p.pSat );
                        // start incremental proof manager
                        assert(p.pSat.pPrf2 == nil, "p.pSat.pPrf2 == nil");
                        p.pSat.pPrf2 = Prf_ManAlloc();
                        if p.pSat.pPrf2 {
                            p.nProofIds = 0;
                            Vec_IntFill( p.vProofIds, Gia_ManObjNum(p.pGia), -1 );
                            Prf_ManRestart( p.pSat.pPrf2, p.vProofIds, sat_solver2_nlearnts(p.pSat), Vec_IntSize(vPPis) );
                        }
                    
                    } else {
                        // resize the proof logger
                        if p.pSat.pPrf2 {
                            Prf_ManGrow( p.pSat.pPrf2, p.nProofIds + Vec_IntSize(vPPis) ); }
                    }

                    Ga2_ManAddToAbs( p, vPPis );
                    Vec_IntFree( vPPis );
                    if pPars.fVerbose {
                        Ga2_ManAbsPrintFrame( p, f, sat_solver2_nconflicts(p.pSat)-nConflsBeg, c+1, Abc_Clock() - clk, 0 ); }
                    // check if the number of objects is below limit
                    if pPars.nRatioMin2 && Vec_IntSize(p.vAbs) >= p.nMarked * pPars.nRatioMin2 / 100 {
                        Status = l_Undef;
                        goto finish;
                    }
                    continue;
                }
                p.timeUnsat += Abc_Clock() - clk2;
                if Status == l_Undef { // ran out of resources
                    goto finish; }
                if p.pSat.nRuntimeLimit && Abc_Clock() > p.pSat.nRuntimeLimit { // timeout
                    goto finish; }
                if c == 0 {
                    if f > p.pPars.iFrameProved {
                        p.pPars.nFramesNoChange++; }
                    break;
                }
                if f > p.pPars.iFrameProved {
                    p.pPars.nFramesNoChange = 0; }

                // derive the core
                assert(p.pSat.pPrf2 != nil, "p.pSat.pPrf2 != nil");
                vCore = (Vec_Int_t *)Sat_ProofCore( p.pSat );
                Prf_ManStopP( &p.pSat.pPrf2 );
                // update the SAT solver
                sat_solver2_rollback( p.pSat );
                // reduce abstraction
                Ga2_ManShrinkAbs( p, nAbs, nValues, nVarsOld );

                // purify UNSAT core
                if fUseSecondCore != 0 {
//                    int nOldCore = Vec_IntSize(vCore);
                    // reverse the order of objects in the core
//                    Vec_IntSort( vCore, 0 );
//                    Vec_IntPrint( vCore );
                    // create bookmark to be used for rollback
                    assert(nVarsOld == p.pSat.size, "nVarsOld == p.pSat.size");
                    sat_solver2_bookmark( p.pSat );
                    // start incremental proof manager
                    assert(p.pSat.pPrf2 == nil, "p.pSat.pPrf2 == nil");
                    p.pSat.pPrf2 = Prf_ManAlloc();
                    if p.pSat.pPrf2 {
                        p.nProofIds = 0;
                        Vec_IntFill( p.vProofIds, Gia_ManObjNum(p.pGia), -1 );
                        Prf_ManRestart( p.pSat.pPrf2, p.vProofIds, sat_solver2_nlearnts(p.pSat), Vec_IntSize(vCore) );

                        Ga2_ManAddToAbs( p, vCore );
                        Vec_IntFree( vCore );
                    }
                    // run SAT solver
                    clk2 = Abc_Clock();
                    Status = sat_solver2_solve( p.pSat, &Lit, &Lit+1, (ABC_INT64_T)pPars.nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
                    if Status == l_Undef {
                        goto finish; }
                    assert(Status == l_False, "Status == l_False");
                    p.timeUnsat += Abc_Clock() - clk2;

                    // derive the core
                    assert(p.pSat.pPrf2 != nil, "p.pSat.pPrf2 != nil");
                    vCore = (Vec_Int_t *)Sat_ProofCore( p.pSat );
                    Prf_ManStopP( &p.pSat.pPrf2 );
                    // update the SAT solver
                    sat_solver2_rollback( p.pSat );
                    // reduce abstraction
                    Ga2_ManShrinkAbs( p, nAbs, nValues, nVarsOld );
//                    printf( "\n%4d -> %4d\n", nOldCore, Vec_IntSize(vCore) );
                }

                Ga2_ManAddToAbs( p, vCore );
//                Ga2_ManRefinePrint( p, vCore );
                Vec_IntFree( vCore );
                break;
            }
            // remember the last proved frame
            if p.pPars.iFrameProved < f {
                p.pPars.iFrameProved = f; }
            // print statistics
            if pPars.fVerbose {
                Ga2_ManAbsPrintFrame( p, f, sat_solver2_nconflicts(p.pSat)-nConflsBeg, c, Abc_Clock() - clk, 1 ); }
            // check if abstraction was proved
            if Gia_GlaProveCheck( pPars.fVerbose ) {
                RetValue = 1;
                goto finish;
            }
            if c > 0 {
                if p.pPars.fVeryVerbose {
                    Abc_Print( 1, "\n" ); }
                // recompute the abstraction
                Vec_IntFreeP( &pAig.vGateClasses );
                pAig.vGateClasses = Ga2_ManAbsTranslate( p );
                // check if the number of objects is below limit
                if pPars.nRatioMin && Vec_IntSize(p.vAbs) >= p.nMarked * pPars.nRatioMin / 100 {
                    Status = l_Undef;
                    goto finish;
                }
            }
            // check the number of stable frames
            if p.pPars.nFramesNoChange == p.pPars.nFramesNoChangeLim {
                // dump the model into file
                if p.pPars.fDumpVabs || p.pPars.fDumpMabs {
                    var Command string // was [1000]byte in C

                    Abc_FrameSetStatus( -1 );
                    Abc_FrameSetCex( nil );
                    Abc_FrameSetNFrames( f );
                    Command = fmt.Sprintf("write_status %s", Extra_FileNameGenericAppend((char *)(p.pPars.pFileVabs ? p.pPars.pFileVabs : "glabs.aig"), ".status"));
                    Cmd_CommandExecute( Abc_FrameGetGlobalFrame(), Command );
                    Ga2_GlaDumpAbsracted( p, pPars.fVerbose );
                }
                // call the prover
                if p.pPars.fCallProver {
                    // cancel old one if it is proving
                    if iFrameTryToProve >= 0 {
                        Gia_GlaProveCancel( pPars.fVerbose ); }
                    // prove new one
                    Gia_GlaProveAbsracted( pAig, pPars.fSimpProver, pPars.fVeryVerbose );
                    iFrameTryToProve = f;
                    p.nPdrCalls++;
                }
                // speak to the bridge
                if Abc_FrameIsBridgeMode() {
                    // cancel old one if it was sent
                    if fOneIsSent != 0 {
                        Gia_Ga2SendCancel( p, pPars.fVerbose ); }
                    // send new one 
                    Gia_Ga2SendAbsracted( p, pPars.fVerbose );
                    fOneIsSent = 1;
                }
            }
            // if abstraction grew more than a certain percentage, force a restart
            if pPars.nRatioMax == 0 {
                continue; }
            if c > 0 && (f > 20 || Vec_IntSize(p.vAbs) > 100) && Vec_IntSize(p.vAbs) - nAbsOld >= nAbsOld * pPars.nRatioMax / 100 {
                if p.pPars.fVerbose {
                Abc_Print( 1, "Forcing restart because abstraction grew from %d to %d (more than %d %%).\n", 
                    nAbsOld, Vec_IntSize(p.vAbs), pPars.nRatioMax ); }
                break;
            }
        }
    }
finish:
    Prf_ManStopP( &p.pSat.pPrf2 );
    // cancel old one if it is proving
    if iFrameTryToProve >= 0 {
        Gia_GlaProveCancel( pPars.fVerbose ); }
    // analize the results
    if !p.fUseNewLine {
        Abc_Print( 1, "\n" ); }
    if RetValue == 1 {
        Abc_Print( 1, "GLA completed %d frames and proved abstraction derived in frame %d  ", p.pPars.iFrameProved+1, iFrameTryToProve );
    } else if pAig.pCexSeq == nil {
        Vec_IntFreeP( &pAig.vGateClasses );
        pAig.vGateClasses = Ga2_ManAbsTranslate( p );
        if p.pPars.nTimeOut && Abc_Clock() >= p.pSat.nRuntimeLimit { 
            Abc_Print( 1, "GLA reached timeout %d sec in frame %d with a %d-stable abstraction.    ", p.pPars.nTimeOut, p.pPars.iFrameProved+1, p.pPars.nFramesNoChange );
        } else if pPars.nConfLimit && sat_solver2_nconflicts(p.pSat) >= pPars.nConfLimit {
            Abc_Print( 1, "GLA exceeded %d conflicts in frame %d with a %d-stable abstraction.  ", pPars.nConfLimit, p.pPars.iFrameProved+1, p.pPars.nFramesNoChange );
        } else if pPars.nRatioMin2 && Vec_IntSize(p.vAbs) >= p.nMarked * pPars.nRatioMin2 / 100 {
            Abc_Print( 1, "GLA found that the size of abstraction exceeds %d %% in frame %d during refinement.  ", pPars.nRatioMin2, p.pPars.iFrameProved+1 );
        } else if pPars.nRatioMin && Vec_IntSize(p.vAbs) >= p.nMarked * pPars.nRatioMin / 100 {
            Abc_Print( 1, "GLA found that the size of abstraction exceeds %d %% in frame %d.  ", pPars.nRatioMin, p.pPars.iFrameProved+1 );
        } else {
            Abc_Print( 1, "GLA finished %d frames and produced a %d-stable abstraction.  ", p.pPars.iFrameProved+1, p.pPars.nFramesNoChange ); }
        p.pPars.iFrame = p.pPars.iFrameProved;
    
    } else {
        if p.pPars.fVerbose {
            Abc_Print( 1, "\n" ); }
        if !Gia_ManVerifyCex( pAig, pAig.pCexSeq, 0 ) {
            Abc_Print( 1, "    Gia_ManPerformGlaOld(): CEX verification has failed!\n" ); }
        Abc_Print( 1, "True counter-example detected in frame %d.  ", f );
        p.pPars.iFrame = f - 1;
        Vec_IntFreeP( &pAig.vGateClasses );
        RetValue = 0;
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    if p.pPars.fVerbose {
        p.timeOther = (Abc_Clock() - clk) - p.timeUnsat - p.timeSat - p.timeCex - p.timeInit;
        ABC_PRTP( "Runtime: Initializing", p.timeInit,   Abc_Clock() - clk );
        ABC_PRTP( "Runtime: Solver UNSAT", p.timeUnsat,  Abc_Clock() - clk );
        ABC_PRTP( "Runtime: Solver SAT  ", p.timeSat,    Abc_Clock() - clk );
        ABC_PRTP( "Runtime: Refinement  ", p.timeCex,    Abc_Clock() - clk );
        ABC_PRTP( "Runtime: Other       ", p.timeOther,  Abc_Clock() - clk );
        ABC_PRTP( "Runtime: TOTAL       ", Abc_Clock() - clk, Abc_Clock() - clk );
        Ga2_ManReportMemory( p );
    }
//    Ga2_ManDumpStats( p->pGia, p->pPars, p->pSat, p->pPars->iFrameProved, 0 );
    Ga2_ManStop( p );
    fflush( os.Stdout );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


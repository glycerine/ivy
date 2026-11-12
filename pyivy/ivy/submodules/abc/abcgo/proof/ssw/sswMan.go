package ssw

import "time"

/**CFile****************************************************************

  FileName    [sswMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Calls to the SAT solver.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswMan.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManCreate(pAig *Aig_Man_t, pPars *Ssw_Pars_t) *Ssw_Man_t {
    var p *Ssw_Man_t;
    // prepare the sequential AIG
    assert(Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0");
    Aig_ManFanoutStart( pAig );
    Aig_ManSetCioIds( pAig );
    // create interpolation manager
    // inlining of macro ABC_ALLOC
    p = ((Ssw_Man_t *) malloc(sizeof(Ssw_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Ssw_Man_t) );
    p.pPars         = pPars;
    p.pAig          = pAig;
    p.nFrames       = pPars.nFramesK + 1;
    // inlining of macro ABC_CALLOC
    p.pNodeToFrames = ((Aig_Obj_t * *) calloc((size_t)(Aig_ManObjNumMax(p.pAig) * p.nFrames), sizeof(Aig_Obj_t *)));
    p.vCommon       = Vec_PtrAlloc( 100 );
    p.iOutputLit    = -1;
    // allocate storage for sim pattern
    p.nPatWords     = Abc_BitWordNum( Saig_ManPiNum(pAig) * p.nFrames + Saig_ManRegNum(pAig) );
    // inlining of macro ABC_CALLOC
    p.pPatWords     = ((unsigned *) calloc((size_t)(p.nPatWords), sizeof(unsigned)));
    // other
    p.vNewLos       = Vec_PtrAlloc( 100 );
    p.vNewPos       = Vec_IntAlloc( 100 );
    p.vResimConsts  = Vec_PtrAlloc( 100 );
    p.vResimClasses = Vec_PtrAlloc( 100 );
//    p->pPars->fVerbose = 1;
    return p;
}

/**Function*************************************************************

  Synopsis    [Prints stats of the manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManCountEquivs(p *Ssw_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i, nEquivs int32;
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        nEquivs += ( Aig_ObjRepr(p.pAig, pObj) != nil ); }
    }
    return nEquivs;
}

/**Function*************************************************************

  Synopsis    [Prints stats of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManPrintStats(p *Ssw_Man_t) {
    var nMemory float64  = 1.0*Aig_ManObjNumMax(p.pAig)*p.nFrames*(2*sizeof(int)+2*sizeof(void*))/(1<<20);

    Abc_Print( 1, "Parameters: F = %d. AddF = %d. C-lim = %d. Constr = %d. MaxLev = %d. Mem = %0.2f MB.\n",
        p.pPars.nFramesK, p.pPars.nFramesAddSim, p.pPars.nBTLimit, Saig_ManConstrNum(p.pAig), p.pPars.nMaxLevs, nMemory );
    Abc_Print( 1, "AIG       : PI = %d. PO = %d. Latch = %d. Node = %d.  Ave SAT vars = %d.\n",
        Saig_ManPiNum(p.pAig), Saig_ManPoNum(p.pAig), Saig_ManRegNum(p.pAig), Aig_ManNodeNum(p.pAig),
        0/(p.pPars.nIters+1) );
    Abc_Print( 1, "SAT calls : Proof = %d. Cex = %d. Fail = %d. Lits proved = %d.\n",
        p.nSatProof, p.nSatCallsSat, p.nSatFailsReal, Ssw_ManCountEquivs(p) );
    Abc_Print( 1, "SAT solver: Vars max = %d. Calls max = %d. Recycles = %d. Sim rounds = %d.\n",
        p.nVarsMax, p.nCallsMax, p.nRecyclesTotal, p.nSimRounds );
    Abc_Print( 1, "NBeg = %d. NEnd = %d. (Gain = %6.2f %%).  RBeg = %d. REnd = %d. (Gain = %6.2f %%).\n",
        p.nNodesBeg, p.nNodesEnd, 100.0*(p.nNodesBeg-p.nNodesEnd)/(p.nNodesBeg?p.nNodesBeg:1),
        p.nRegsBeg, p.nRegsEnd, 100.0*(p.nRegsBeg-p.nRegsEnd)/(p.nRegsBeg?p.nRegsBeg:1) );

    p.timeOther = p.timeTotal-p.timeBmc-p.timeReduce-p.timeMarkCones-p.timeSimSat-p.timeSat;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("BMC        "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeBmc)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeBmc)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Spec reduce"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeReduce)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeReduce)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Mark cones "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeMarkCones)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeMarkCones)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Sim SAT    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSimSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSimSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("SAT solving"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("  unsat    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnsat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUnsat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("  sat      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("  undecided"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUndec)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUndec)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Other      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeOther)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("TOTAL      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;

    // report the reductions
    if p.pAig.nConstrs {
        Abc_Print( 1, "Statistics reflecting the use of constraints:\n" );
        Abc_Print( 1, "Total cones  = %6d.  Constraint cones = %6d. (%6.2f %%)\n",
            p.nConesTotal, p.nConesConstr, 100.0*p.nConesConstr/p.nConesTotal );
        Abc_Print( 1, "Total equivs = %6d.  Removed equivs   = %6d. (%6.2f %%)\n",
            p.nEquivsTotal, p.nEquivsConstr, 100.0*p.nEquivsConstr/p.nEquivsTotal );
        Abc_Print( 1, "NBeg = %d. NEnd = %d. (Gain = %6.2f %%).  RBeg = %d. REnd = %d. (Gain = %6.2f %%).\n",
            p.nNodesBegC, p.nNodesEndC, 100.0*(p.nNodesBegC-p.nNodesEndC)/(p.nNodesBegC?p.nNodesBegC:1),
            p.nRegsBegC, p.nRegsEndC,   100.0*(p.nRegsBegC-p.nRegsEndC)/(p.nRegsBegC?p.nRegsBegC:1) );
    }
}

/**Function*************************************************************

  Synopsis    [Frees the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManCleanup(p *Ssw_Man_t) {
//    Aig_ManCleanMarkAB( p->pAig );
    assert(p.pMSat == nil, "p.pMSat == nil");
    if p.pFrames {
        Aig_ManCleanMarkAB( p.pFrames );
        Aig_ManStop( p.pFrames );
        p.pFrames = nil;
        memset( p.pNodeToFrames, 0, sizeof(Aig_Obj_t *) * Aig_ManObjNumMax(p.pAig) * p.nFrames );
    }
    if p.vSimInfo {
        Vec_PtrFree( p.vSimInfo );
        p.vSimInfo = nil;
    }
    p.nConstrTotal = 0;
    p.nConstrReduced = 0;
}

/**Function*************************************************************

  Synopsis    [Frees the manager.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManStop(p *Ssw_Man_t) {
    p.pVisited = nil;
    if p.pPars.fVerbose {//&& p->pPars->nStepsMax == -1 )
        Ssw_ManPrintStats( p ); }
    if p.ppClasses {
        Ssw_ClassesStop( p.ppClasses ); }
    if p.pSml {
        Ssw_SmlStop( p.pSml ); }
    if p.vDiffPairs {
        Vec_IntFree( p.vDiffPairs ); }
    if p.vInits {
        Vec_IntFree( p.vInits ); }
    Vec_PtrFree( p.vResimConsts );
    Vec_PtrFree( p.vResimClasses );
    Vec_PtrFree( p.vNewLos );
    Vec_IntFree( p.vNewPos );
    Vec_PtrFree( p.vCommon );
    p.pNodeToFrames = nil;
    p.pPatWords = nil;
    p = nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

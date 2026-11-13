package ssw

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [sswCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [The core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswCore.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

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

  Synopsis    [Reports improvements for property cones.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ReportConeReductions(p *Ssw_Man_t, pAigInit *Aig_Man_t, pAigStop *Aig_Man_t) {
    var pAig1, pAig2, pAux *Aig_Man_t;
    pAig1 = Aig_ManDupOneOutput( pAigInit, 0, 1 );
    pAig1 = Aig_ManScl( pAux = pAig1, 1, 1, 0, -1, -1, 0, 0 );
    Aig_ManStop( pAux );
    pAig2 = Aig_ManDupOneOutput( pAigStop, 0, 1 );
    pAig2 = Aig_ManScl( pAux = pAig2, 1, 1, 0, -1, -1, 0, 0 );
    Aig_ManStop( pAux );

    p.nNodesBegC = Aig_ManNodeNum(pAig1);
    p.nNodesEndC = Aig_ManNodeNum(pAig2);
    p.nRegsBegC  = Aig_ManRegNum(pAig1);
    p.nRegsEndC  = Aig_ManRegNum(pAig2);

    Aig_ManStop( pAig1 );
    Aig_ManStop( pAig2 );
}

/**Function*************************************************************

  Synopsis    [Reports one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ReportOneOutput(p *Aig_Man_t, pObj *Aig_Obj_t) {
    if pObj == Aig_ManConst1(p) {
        Abc_Print( 1, "1" );
    } else if pObj == Aig_ManConst0(p) {
        Abc_Print( 1, "0" );
    } else {
        Abc_Print( 1, "X" ); }
}

/**Function*************************************************************

  Synopsis    [Reports improvements for property cones.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ReportOutputs(pAig *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        if i < Saig_ManPoNum(pAig)-Saig_ManConstrNum(pAig) {
            Abc_Print( 1, "o" );
        } else {
            Abc_Print( 1, "c" ); }
        Ssw_ReportOneOutput( pAig, Aig_ObjChild0(pObj) );
    }
    Abc_Print( 1, "\n" );
}

/**Function*************************************************************

  Synopsis    [Remove from-equivs that are in the cone of constraints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManUpdateEquivs(p *Ssw_Man_t, pAig *Aig_Man_t, fVerbose int32) {
    var vCones *Vec_Ptr_t;
    var pArray **Aig_Obj_t;
    var pObj *Aig_Obj_t;
    var i, nTotal, nRemoved int32;
    // collect the nodes in the cone of constraints
    pArray  = (Aig_Obj_t **)Vec_PtrArray(pAig.vCos);
    pArray += Saig_ManPoNum(pAig) - Saig_ManConstrNum(pAig);
    vCones  = Aig_ManDfsNodes( pAig, pArray, Saig_ManConstrNum(pAig) );
    // remove all the node that are equiv to something and are in the cones
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Aig_ObjIsCi(pObj) && !Aig_ObjIsNode(pObj) {
            continue; }
        if pAig.pReprs[i] != nil {
            nTotal++; }
        if !Aig_ObjIsTravIdCurrent(pAig, pObj) {
            continue; }
        if pAig.pReprs[i] {
            if p.pPars.fConstrs && !p.pPars.fMergeFull {
                pAig.pReprs[i] = nil;
                nRemoved++;
            }
        }
    }
    }
    // collect statistics    
    p.nConesTotal   = Aig_ManCiNum(pAig) + Aig_ManNodeNum(pAig);
    p.nConesConstr  = Vec_PtrSize(vCones);
    p.nEquivsTotal  = nTotal;
    p.nEquivsConstr = nRemoved;
    Vec_PtrFree( vCones );
}

/**Function*************************************************************

  Synopsis    [Performs computation of signal correspondence with constraints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SignalCorrespondenceRefine(p *Ssw_Man_t) *Aig_Man_t {
    var nSatProof, nSatCallsSat, nRecycles, nSatFailsReal, nUniques int32;
    var pAigNew *Aig_Man_t;
    var RetValue int32;
    var nIter int32 = -1;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    // get the starting stats
    p.nLitsBeg  = Ssw_ClassesLitNum( p.ppClasses );
    p.nNodesBeg = Aig_ManNodeNum(p.pAig);
    p.nRegsBeg  = Aig_ManRegNum(p.pAig);
    // refine classes using BMC
    if p.pPars.fVerbose {
        Abc_Print( 1, "Before BMC: " );
        Ssw_ClassesPrint( p.ppClasses, 0 );
    }
    if !p.pPars.fLatchCorr || p.pPars.nFramesK > 1 {
        p.pMSat = Ssw_SatStart( 0 );
        if p.pPars.fConstrs {
            Ssw_ManSweepBmcConstr( p );
        } else {
            Ssw_ManSweepBmc( p ); }
        Ssw_SatStop( p.pMSat );
        p.pMSat = nil;
        Ssw_ManCleanup( p );
    }
    if p.pPars.fVerbose {
        Abc_Print( 1, "After  BMC: " );
        Ssw_ClassesPrint( p.ppClasses, 0 );
    }
    // apply semi-formal filtering
/*
    if ( p->pPars->fSemiFormal )
    {
        Aig_Man_t * pSRed;
        Ssw_FilterUsingSemi( p, 0, 2000, p->pPars->fVerbose );
//        Ssw_FilterUsingSemi( p, 1, 100000, p->pPars->fVerbose );
        pSRed = Ssw_SpeculativeReduction( p );
        Aig_ManDumpBlif( pSRed, "srm.blif", NULL, NULL );
        Aig_ManStop( pSRed );
    }
*/
    if p.pPars.pFunc {
        ((int (*)(void *))p.pPars.pFunc)( p.pPars.pData );
        ((int (*)(void *))p.pPars.pFunc)( p.pPars.pData );
    }
    if p.pPars.nStepsMax == 0 {
        Abc_Print( 1, "Stopped signal correspondence after BMC.\n" );
        goto finalize;
    }
    // refine classes using induction
    nSatProof = nSatCallsSat = nRecycles = nSatFailsReal = nUniques = 0;
    for  nIter = 0; ; nIter++  {
        if p.pPars.nStepsMax == nIter {
            Abc_Print( 1, "Stopped signal correspondence after %d refiment iterations.\n", nIter );
            goto finalize;
        }
        if p.pPars.nItersStop >= 0 && p.pPars.nItersStop == nIter {
            var pSRed *Aig_Man_t  = Ssw_SpeculativeReduction( p );
            Aig_ManDumpBlif( pSRed, "srm.blif", nil, nil );
            Aig_ManStop( pSRed );
            Abc_Print( 1, "Iterative refinement is stopped before iteration %d.\n", nIter );
            Abc_Print( 1, "The network is reduced using candidate equivalences.\n" );
            Abc_Print( 1, "Speculatively reduced miter is saved in file \"%s\".\n", "srm.blif" );
            Abc_Print( 1, "If the miter is SAT, the reduced result is incorrect.\n" );
            break;
        }

clk = Abc_Clock();
        p.pMSat = Ssw_SatStart( 0 );
        if p.pPars.fLatchCorrOpt {
            RetValue = Ssw_ManSweepLatch( p );
            if p.pPars.fVerbose {
                Abc_Print( 1, "%3d : C =%7d. Cl =%7d. Pr =%6d. Cex =%5d. R =%4d. F =%4d. ",
                    nIter, Ssw_ClassesCand1Num(p.ppClasses), Ssw_ClassesClassNum(p.ppClasses),
                    p.nSatProof-nSatProof, p.nSatCallsSat-nSatCallsSat,
                    p.nRecycles-nRecycles, p.nSatFailsReal-nSatFailsReal );
                // inlining of macro ABC_PRT
                Abc_Print(1, "%s =", ("T"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
            }
        
        } else {
            if p.pPars.fConstrs {
                RetValue = Ssw_ManSweepConstr( p );
            } else if p.pPars.fDynamic {
                RetValue = Ssw_ManSweepDyn( p );
            } else {
                RetValue = Ssw_ManSweep( p ); }

            p.pPars.nConflicts += p.pMSat.pSat.stats.conflicts;
            if p.pPars.fVerbose {
                Abc_Print( 1, "%3d : C =%7d. Cl =%7d. LR =%6d. NR =%6d. ",
                    nIter, Ssw_ClassesCand1Num(p.ppClasses), Ssw_ClassesClassNum(p.ppClasses),
                    p.nConstrReduced, Aig_ManNodeNum(p.pFrames) );
                if p.pPars.fDynamic {
                    Abc_Print( 1, "Cex =%5d. ", p.nSatCallsSat-nSatCallsSat );
                    Abc_Print( 1, "R =%4d. ",   p.nRecycles-nRecycles );
                }
                Abc_Print( 1, "F =%5d. %s ", p.nSatFailsReal-nSatFailsReal,
                    (Saig_ManPoNum(p.pAig)==1 && Ssw_ObjIsConst1Cand(p.pAig,Aig_ObjFanin0(Aig_ManCo(p.pAig,0))))? "+" : "-" );
                // inlining of macro ABC_PRT
                Abc_Print(1, "%s =", ("T"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
            }
//            if ( p->pPars->fDynamic && p->nSatCallsSat-nSatCallsSat < 100 )
//                p->pPars->nBTLimit = 10000;

            if p.pPars.fStopWhenGone && Saig_ManPoNum(p.pAig) == 1 && !Ssw_ObjIsConst1Cand(p.pAig,Aig_ObjFanin0(Aig_ManCo(p.pAig,0))) {
                fmt.Printf("Iterative refinement is stopped after iteration %d\n", nIter);
                fmt.Printf("because the property output is no longer a candidate constant.\n");
                // prepare to quite
                p.nLitsEnd  = p.nLitsBeg;
                p.nNodesEnd = p.nNodesBeg;
                p.nRegsEnd  = p.nRegsBeg;
                // cleanup
                Ssw_SatStop( p.pMSat );
                p.pMSat = nil;
                Ssw_ManCleanup( p );
                // cleanup
                Aig_ManSetPhase( p.pAig );
                Aig_ManCleanMarkB( p.pAig );
                return Aig_ManDupSimple( p.pAig );
            }
        }
        nSatProof     = p.nSatProof;
        nSatCallsSat  = p.nSatCallsSat;
        nRecycles     = p.nRecycles;
        nSatFailsReal = p.nSatFailsReal;
        nUniques      = p.nUniques;

        p.nVarsMax  = Abc_MaxInt( p.nVarsMax,  p.pMSat.nSatVars );
        p.nCallsMax = Abc_MaxInt( p.nCallsMax, p.pMSat.nSolverCalls );
        Ssw_SatStop( p.pMSat );
        p.pMSat = nil;
        Ssw_ManCleanup( p );
        if !(RetValue != 0) {
            break; }
        if p.pPars.pFunc {
            ((int (*)(void *))p.pPars.pFunc)( p.pPars.pData ); }
    }

finalize:
    p.pPars.nIters = nIter + 1;
p.timeTotal = Abc_Clock() - clkTotal;

    Ssw_ManUpdateEquivs( p, p.pAig, p.pPars.fVerbose );
    pAigNew = Aig_ManDupRepr( p.pAig, 0 );
    Aig_ManSeqCleanup( pAigNew );
//Ssw_ClassesPrint( p->ppClasses, 1 );
    // get the final stats
    p.nLitsEnd  = Ssw_ClassesLitNum( p.ppClasses );
    p.nNodesEnd = Aig_ManNodeNum(pAigNew);
    p.nRegsEnd  = Aig_ManRegNum(pAigNew);
    // cleanup
    Aig_ManSetPhase( p.pAig );
    Aig_ManCleanMarkB( p.pAig );
    return pAigNew;
}

/**Function*************************************************************

  Synopsis    [Performs computation of signal correspondence with constraints.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SignalCorrespondence(pAig *Aig_Man_t, pPars *Ssw_Pars_t) *Aig_Man_t {
    var Pars Ssw_Pars_t;
    var pAigNew *Aig_Man_t;
    var p *Ssw_Man_t;
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    // reset random numbers
    Aig_ManRandom( 1 );
    // if parameters are not given, create them
    if pPars == nil {
        Ssw_ManSetDefaultParams( pPars = &Pars ); }
    // consider the case of empty AIG
    if Aig_ManNodeNum(pAig) == 0 {
        pPars.nIters = 0;
        // Ntl_ManFinalize() needs the following to satisfy an assertion
        Aig_ManReprStart( pAig,Aig_ManObjNumMax(pAig) );
        return Aig_ManDupOrdered(pAig);
    }
    // check and update parameters
    if pPars.fLatchCorrOpt {
        pPars.fLatchCorr = 1;
        pPars.nFramesAddSim = 0;
        if pAig.vClockDoms && Vec_VecSize(pAig.vClockDoms) > 0 {
           return Ssw_SignalCorrespondencePart( pAig, pPars ); }
    
    } else {
        assert(pPars.nFramesK > 0, "pPars.nFramesK > 0");
        // perform partitioning
        if (pPars.nPartSize > 0 && pPars.nPartSize < Aig_ManRegNum(pAig))
             || (pAig.vClockDoms && Vec_VecSize(pAig.vClockDoms) > 0) {
            return Ssw_SignalCorrespondencePart( pAig, pPars ); }
    }

    if pPars.fScorrGia {
        if pPars.fLatchCorrOpt {
            extern Aig_Man_t * Cec_LatchCorrespondence( Aig_Man_t * pAig, int nConfs, int fUseCSat );
            return Cec_LatchCorrespondence( pAig, pPars.nBTLimit, pPars.fUseCSat );
        
        } else {
            extern Aig_Man_t * Cec_SignalCorrespondence( Aig_Man_t * pAig, int nConfs, int fUseCSat );
            return Cec_SignalCorrespondence( pAig, pPars.nBTLimit, pPars.fUseCSat );
        }
    }

    // start the induction manager
    p = Ssw_ManCreate( pAig, pPars );
    // compute candidate equivalence classes
//    p->pPars->nConstrs = 1;
    if p.pPars.fConstrs {
        // create trivial equivalence classes with all nodes being candidates for constant 1
        p.ppClasses = Ssw_ClassesPrepareSimple( pAig, pPars.fLatchCorr, pPars.nMaxLevs );
        Ssw_ClassesSetData( p.ppClasses, nil, nil, Ssw_SmlObjIsConstBit, Ssw_SmlObjsAreEqualBit );
        // derive phase bits to satisfy the constraints
        if Ssw_ManSetConstrPhases( pAig, p.pPars.nFramesK + 1, &p.vInits ) != 0 {
            Abc_Print( 1, "Ssw_SignalCorrespondence(): The init state does not satisfy the constraints!\n" );
            p.pPars.fVerbose = 0;
            Ssw_ManStop( p );
            return nil;
        }
        // perform simulation of the first timeframes
        Ssw_ManRefineByConstrSim( p );
    
    } else {
        // perform one round of seq simulation and generate candidate equivalence classes
        p.ppClasses = Ssw_ClassesPrepare( pAig, pPars.nFramesK, pPars.fLatchCorr, pPars.fConstCorr, pPars.fOutputCorr, pPars.nMaxLevs, pPars.fVerbose );
//        p->ppClasses = Ssw_ClassesPrepareTargets( pAig );
        if pPars.fLatchCorrOpt {
            p.pSml = Ssw_SmlStart( pAig, 0, 2, 1 );
        } else if pPars.fDynamic {
            p.pSml = Ssw_SmlStart( pAig, 0, p.nFrames + p.pPars.nFramesAddSim, 1 );
        } else {
            p.pSml = Ssw_SmlStart( pAig, 0, 1 + p.pPars.nFramesAddSim, 1 ); }
        Ssw_ClassesSetData( p.ppClasses, p.pSml, (unsigned(*)(void *,Aig_Obj_t *))Ssw_SmlObjHashWord, (int(*)(void *,Aig_Obj_t *))Ssw_SmlObjIsConstWord, (int(*)(void *,Aig_Obj_t *,Aig_Obj_t *))Ssw_SmlObjsAreEqualWord );
    }
    // allocate storage
    if p.pPars.fLocalSim && p.pSml {
        // inlining of macro ABC_CALLOC
        p.pVisited = ((int *) calloc((size_t)(Ssw_SmlNumFrames( p.pSml ) * Aig_ManObjNumMax(p.pAig)), sizeof(int))); }
    // perform refinement of classes
    pAigNew = Ssw_SignalCorrespondenceRefine( p );
//    Ssw_ReportOutputs( pAigNew );
    if pPars.fConstrs && pPars.fVerbose {
        Ssw_ReportConeReductions( p, pAig, pAigNew ); }
    // cleanup
    Ssw_ManStop( p );
    return pAigNew;
}

/**Function*************************************************************

  Synopsis    [Performs computation of latch correspondence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_LatchCorrespondence(pAig *Aig_Man_t, pPars *Ssw_Pars_t) *Aig_Man_t {
    var pRes *Aig_Man_t;
    var Pars Ssw_Pars_t;
    if pPars == nil {
        Ssw_ManSetDefaultParamsLcorr( pPars = &Pars ); }
    pRes = Ssw_SignalCorrespondence( pAig, pPars );
//    if ( pPars->fConstrs && pPars->fVerbose )
//        Ssw_ReportConeReductions( pAig, pRes );
    return pRes;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

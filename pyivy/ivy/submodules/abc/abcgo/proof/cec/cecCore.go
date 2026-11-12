package cec

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [cecCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecCore.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
  

/**Function************  *************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
 

/**Function************  *************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
 

/**Function************  *************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
 

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
  

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
  

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
  

/**Function*************************************************************

  Synopsis    [Core procedure for SAT sweeping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSatSolving(pAig *Gia_Man_t, pPars *Cec_ParSat_t, f0Proved int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pPat *Cec_ManPat_t;
    pPat = Cec_ManPatStart();
    if pPars.SolverType == -1 {
        Cec_ManSatSolve( pPat, pAig, pPars, nil, nil, nil, f0Proved );
    } else {
        CecG_ManSatSolve( pPat, pAig, pPars, f0Proved ); }
//    pNew = Gia_ManDupDfsSkip( pAig );
    pNew = Gia_ManCleanup( pAig );
    Cec_ManPatStop( pPat );
    pNew.vSeqModelVec = pAig.vSeqModelVec;
    pAig.vSeqModelVec = nil;
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Core procedure for simulation.]

  Description [Returns 1 if refinement has happened.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimulationOne(pAig *Gia_Man_t, pPars *Cec_ParSim_t) int32 {
    var pSim *Cec_ManSim_t;
    var RetValue int32;
    var clkTotal abctime  = Abc_Clock();
    pSim = Cec_ManSimStart( pAig, pPars );
    if (pAig.pReprs == nil && (RetValue = Cec_ManSimClassesPrepare( pSim, -1 ))) ||
         (RetValue == 0 &&        (RetValue = Cec_ManSimClassesRefine( pSim ))) {
        Abc_Print( 1, "The number of failed outputs of the miter = %6d. (Words = %4d. Frames = %4d.)\n", 
            pSim.nOuts, pPars.nWords, pPars.nFrames ); }
    if pPars.fVerbose {
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal ); }
    Cec_ManSimStop( pSim );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Core procedure for simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManSimulation(pAig *Gia_Man_t, pPars *Cec_ParSim_t) {
    var r, nLitsOld, nLitsNew, nCountNoRef, fStop int32;
    Gia_ManRandom( 1 );
    if pPars.fSeqSimulate {
        Abc_Print( 1, "Performing rounds of random simulation of %d frames with %d words.\n", 
            pPars.nRounds, pPars.nFrames, pPars.nWords ); }
    nLitsOld = Gia_ManEquivCountLits( pAig );
    for  r = 0; r < pPars.nRounds; r++  {
        if Cec_ManSimulationOne( pAig, pPars ) != 0 {
            fStop = 1;
            break;
        }
        // decide when to stop
        nLitsNew = Gia_ManEquivCountLits( pAig );
        if nLitsOld == 0 || nLitsOld > nLitsNew {
            nLitsOld = nLitsNew;
            nCountNoRef = 0;
        
        } else if nCountNoRef++; nCountNoRef == pPars.nNonRefines {
            r++;
            break;
        }
        assert(nLitsOld == nLitsNew, "nLitsOld == nLitsNew");
    }
//    if ( pPars->fVerbose )
    if r == pPars.nRounds || (fStop != 0) {
        Abc_Print( 1, "Random simulation is stopped after %d rounds.\n", r );
    } else {
        Abc_Print( 1, "Random simulation saturated after %d rounds.\n", r ); }
    if pPars.fCheckMiter {
        var nNonConsts int32  = Cec_ManCountNonConstOutputs( pAig );
        if nNonConsts != 0 {
            Abc_Print( 1, "The number of POs that are not const-0 candidates = %d.\n", nNonConsts ); }
    }
}

/**Function*************************************************************

  Synopsis    [Core procedure for SAT sweeping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


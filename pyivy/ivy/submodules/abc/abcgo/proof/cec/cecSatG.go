package cec

/**CFile****************************************************************

  FileName    [cecSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Detection of structural isomorphism.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSat.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/gia/gia.h"
//#include "misc/util/utilTruth.h"
//#include "sat/glucose/AbcGlucose.h"
//#include "cec.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// sweeping manager

// SAT solving manager

type Cec3_Man_t struct {
	pPars *Cec3_Par_t // parameters
	pAig  *Gia_Man_t  // user's AIG
	pNew  *Gia_Man_t  // internal AIG
	// SAT solving
	pSat         *bmcg_sat_solver // SAT solver
	vFrontier    *Vec_Ptr_t       // CNF construction
	vFanins      *Vec_Ptr_t       // CNF construction
	vSims        *Vec_Wrd_t       // CI simulation info
	vNodesNew    *Vec_Int_t       // nodes
	vSatVars     *Vec_Int_t       // nodes
	vObjSatPairs *Vec_Int_t       // nodes
	vCexTriples  *Vec_Int_t       // nodes
	// statistics
	nPatterns    int32
	nSatSat      int32
	nSatUnsat    int32
	nSatUndec    int32
	timeSatSat   abctime
	timeSatUnsat abctime
	timeSatUndec abctime
	timeSim      abctime
	timeRefine   abctime
	timeExtra    abctime
	timeStart    abctime
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Sets parameter defaults.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Adds clauses to the solver.]

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

  Synopsis    [Verify counter-example.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// Cec3_ManPerformSweeping runs simulation-guided SAT sweeping using the
// source parameters, retaining the source's miter result convention.
func Cec3_ManPerformSweeping(p *Gia_Man_t, pPars *Cec3_Par_t, ppNew **Gia_Man_t) int32 {
	settings := cec3SweepSettings{
		words: int(pPars.nSimWords), rounds: int(pPars.nSimRounds),
		iterations: int(pPars.nItersMax), conflictLimit: int64(pPars.nConfLimit),
		useCones: pPars.fUseCones != 0, isMiter: pPars.fIsMiter != 0, finishGraph: ppNew != nil, verbose: pPars.fVerbose != 0, veryVerbose: pPars.fVeryVerbose != 0,
	}
	reduced, _, cex := cec3SweepEquivalences(p, settings)
	if ppNew != nil && reduced != nil {
		*ppNew = reduced
	}
	if cex != nil || p.SequentialCounterexample() != nil {
		return 0
	}
	return 1
}

func Cec3_ManSimulateTest(p *Gia_Man_t, pPars0 *Cec_ParFra_t) *Gia_Man_t {
	var pars Cec3_Par_t
	Cec3_SetDefaultParams(&pars)
	pars.nConfLimit = pPars0.nBTLimit
	pars.fUseCones = pPars0.fUseCones
	pars.fVerbose = pPars0.fVerbose
	var reduced *Gia_Man_t
	Cec3_ManPerformSweeping(p, &pars, &reduced)
	return reduced
}

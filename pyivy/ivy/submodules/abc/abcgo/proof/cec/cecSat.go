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
//#include "sat/satoko/satoko.h"
//#include "cec.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// sweeping manager

type Cec2_Par_t struct {
	nSimWords    int32 // simulation words
	nSimRounds   int32 // simulation rounds
	nItersMax    int32 // max number of iterations
	nConfLimit   int32 // SAT solver conflict limit
	fIsMiter     int32 // this is a miter
	fUseCones    int32 // use logic cones
	fVeryVerbose int32 // verbose stats
	fVerbose     int32 // verbose stats
}

// SAT solving manager

type Cec2_Man_t struct {
	pPars *Cec2_Par_t // parameters
	pAig  *Gia_Man_t  // user's AIG
	pNew  *Gia_Man_t  // internal AIG
	// SAT solving
	pSat         *satoko_t  // SAT solver
	vFrontier    *Vec_Ptr_t // CNF construction
	vFanins      *Vec_Ptr_t // CNF construction
	vSims        *Vec_Wrd_t // CI simulation info
	vNodesNew    *Vec_Int_t // nodes
	vSatVars     *Vec_Int_t // nodes
	vObjSatPairs *Vec_Int_t // nodes
	vCexTriples  *Vec_Int_t // nodes
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

/*
*Function*************************************************************

	Synopsis    [Sets parameter defaults.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cec2_SetDefaultParams(p *Cec2_Par_t) {
	*p = Cec2_Par_t{nSimWords: 12, nSimRounds: 4, nItersMax: 10, nConfLimit: 1000, fUseCones: 1}
}

/*
*Function*************************************************************

	Synopsis    [Adds clauses to the solver.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Cec2_ObjGetCnfVar(p *Cec2_Man_t, iObj int32) int32 {
	return Gia_ObjGetCnfVar(p.pNew, iObj, p.vFrontier, p.vFanins, p.pSat)
}

/**Function*************************************************************

  Synopsis    [Internal simulation APIs.]

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
// Cec2_ManPerformSweeping runs simulation-guided SAT sweeping using the
// source parameters, retaining the source's miter result convention.
func Cec2_ManPerformSweeping(p *Gia_Man_t, pPars *Cec2_Par_t, ppNew **Gia_Man_t) int32 {
	settings := cec2SweepSettings{
		words: int(pPars.nSimWords), rounds: int(pPars.nSimRounds),
		iterations: int(pPars.nItersMax), conflictLimit: int64(pPars.nConfLimit),
		useCones: pPars.fUseCones != 0, isMiter: pPars.fIsMiter != 0, finishGraph: ppNew != nil, verbose: pPars.fVerbose != 0, veryVerbose: pPars.fVeryVerbose != 0,
	}
	reduced, _, cex := cec2SweepEquivalences(p, settings)
	if ppNew != nil && reduced != nil {
		*ppNew = reduced
	}
	if cex != nil || p.SequentialCounterexample() != nil {
		return 0
	}
	return 1
}

func Cec2_ManSimulateTest(p *Gia_Man_t, pPars0 *Cec_ParFra_t) *Gia_Man_t {
	var pars Cec2_Par_t
	Cec2_SetDefaultParams(&pars)
	pars.nConfLimit = pPars0.NBTLimit
	pars.fUseCones = pPars0.FUseCones
	pars.fVerbose = pPars0.FVerbose
	var reduced *Gia_Man_t
	Cec2_ManPerformSweeping(p, &pars, &reduced)
	return reduced
}

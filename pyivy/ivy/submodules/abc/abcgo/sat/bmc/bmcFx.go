package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/cnf"
)

/**CFile****************************************************************

  FileName    [bmcFx.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [INT-FX: Interpolation-based logic sharing extraction.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcFx.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "misc/vec/vecWec.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Create hash table to hash divisors.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bmc_FxCompute(p *Gia_Man_t) int32 {
	// create dual-output circuit with on-set/off-set
	//extern Gia_Man_t * Gia_ManDupOnsetOffset( Gia_Man_t * p );
	var p2 *Gia_Man_t = Gia_ManDupOnsetOffset(p)
	// create SAT solver
	pCnf, err := gia.Mf_ManGenerateCnf(p2, 8, 0, 0, 0, 0)
	if err != nil {
		panic(err)
	}
	sat_solver := cnf.Cnf_DataWriteIntoSolver(pCnf, 1, 0)
	// compute parameters
	var nOuts int32 = Gia_ManCoNum(p)
	var nCiVars int32 = Gia_ManCiNum(p)
	var iCiVarBeg int32 = pCnf.nVars - nCiVars // - 1;
	var o int32
	var i int32
	var n int32
	var RetValue int32
	var nCounter int32
	var iAuxVarStart int32 = sat_solver_nvars(pSat)
	var nCubes [2][2]int32 = [2][2]int32{{0}}
	// create variables
	var vVars *Vec_Int_t = Vec_IntAlloc(nCiVars)
	for n = 0; n < nCiVars; n++ {
		Vec_IntPush(vVars, iCiVarBeg+n)
	}
	sat_solver_setnvars(pSat, iAuxVarStart+4*nOuts)
	// iterate through outputs
	for o = 0; o < nOuts; o++ {
		for i = 0; i < 2; i++ {
			for n = 0; n < 2; n++ { // 0=onset, 1=offset
				//    for ( n = 1; n >= 0; n-- ) // 0=onset, 1=offset

				onOff := " on"
				if n != 0 {
					onOff = "off"
				}
				fmt.Printf("Out %3d %sset ", o, onOff)
				RetValue = Bmc_FxSolve(pSat, Abc_Var2Lit(o, n), iAuxVarStart+2*i*nOuts+Abc_Var2Lit(o, n), vVars, 0, 0, &nCounter, nil)
				if RetValue == 0 {
					fmt.Printf("Mismatch\n")
				}
				if RetValue == -1 {
					fmt.Printf("Timeout\n")
				}
				nCubes[i][n] += nCounter
			}
		}
	}
	// cleanup
	Vec_IntFree(vVars)
	sat_solver_delete(pSat)
	Cnf_DataFree(pCnf)
	Gia_ManStop(p2)
	fmt.Printf("Onset = %5d.   Offset = %5d.      Onset = %5d.   Offset = %5d.\n", nCubes[0][0], nCubes[0][1], nCubes[1][0], nCubes[1][1])
	return 1
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Bmc_FxComputeOne(p *Gia_Man_t, nIterMax int32, nDiv2Add int32) int32 {
	var Extra int32 = 1000
	// create SAT solver
	pCnf, err := gia.Mf_ManGenerateCnf(p, 8, 0, 0, 0, 0)
	if err != nil {
		panic(err)
	}
	pSat := cnf.Cnf_DataWriteIntoSolver(pCnf, 1, 0)
	// compute parameters
	var nCiVars int32 = Gia_ManCiNum(p)        // PI count
	var iCiVarBeg int32 = pCnf.nVars - nCiVars //- 1;   // first PI var
	var iCiVarCur int32 = iCiVarBeg + nCiVars  // current unused PI var
	var n, Iter, RetValue int32
	// create variables
	var iAuxVarStart int32 = sat_solver_nvars(pSat) + Extra // the aux var
	sat_solver_setnvars(pSat, iAuxVarStart+1+nIterMax)
	for Iter = 0; Iter < nIterMax; Iter++ {
		var vCubes *Vec_Wec_t = Vec_WecAlloc(1000)
		// collect variables
		var vVar2Sat *Vec_Int_t = Vec_IntAlloc(iCiVarCur - iCiVarBeg)
		var vDivs *Vec_Int_t
		//        for ( n = iCiVarCur - 1; n >= iCiVarBeg; n-- )
		for n = iCiVarBeg; n < iCiVarCur; n++ {
			Vec_IntPush(vVar2Sat, n)
		}
		// iterate through outputs
		fmt.Printf("\nIteration %d (Aux = %d)\n", Iter, iAuxVarStart+Iter)
		RetValue = Bmc_FxSolve(pSat, 0, iAuxVarStart+Iter, vVar2Sat, 1, 1, nil, vCubes)
		if RetValue == 0 {
			fmt.Printf("Mismatch\n")
		}
		if RetValue == -1 {
			fmt.Printf("Timeout\n")
		}
		// print cubes
		//Div_CubePrint( vCubes, nCiVars );
		vDivs = Div_CubePairs(vCubes, nCiVars, nDiv2Add)
		Vec_WecFree(vCubes)
		// add clauses and update variables
		Bmc_FxAddClauses(pSat, vDivs, iCiVarBeg, iCiVarCur)
		iCiVarCur += Vec_IntSize(vDivs) / 4
		Vec_IntFree(vDivs)
		// cleanup
		assert(Vec_IntSize(vVar2Sat) <= nCiVars+Extra, "Vec_IntSize(vVar2Sat) <= nCiVars + Extra")
		Vec_IntFree(vVar2Sat)
	}
	// cleanup
	sat_solver_delete(pSat)
	Cnf_DataFree(pCnf)
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

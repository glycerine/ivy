package bsat

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [satTrace.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sat_solver.]

  Synopsis    [Records the trace of SAT solving in the CNF form.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satTrace.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

//#include "satSolver.h"

//ABC_NAMESPACE_IMPL_START


/*
    The trace of SAT solving contains the original clause of the problem
    along with the learned clauses derived during SAT solving.
    The first line of the resulting file contains 3 numbers instead of 2:
    c <num_vars> <num_all_clauses> <num_root_clauses>
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Start the trace recording.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_SolverTraceStart(pSat *sat_solver, pName string) {
    assert(pSat.pFile == nil, "pSat.pFile == nil");
    pSat.pFile = os.OpenFile(pName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pSat.pFile, "                                        \n");
    pSat.nClauses = 0;
    pSat.nRoots = 0;
}   

/**Function*************************************************************

  Synopsis    [Stops the trace recording.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_SolverTraceStop(pSat *sat_solver) {
    if pSat.pFile == nil {
        return; }
    rewind( pSat.pFile );
    fmt.Fprintf(pSat.pFile, "p %d %d %d", sat_solver_nvars(pSat), pSat.nClauses, pSat.nRoots);
    _ = (pSat.pFile).Close();
    pSat.pFile = nil;
}


/**Function*************************************************************

  Synopsis    [Writes one clause into the trace file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_SolverTraceWrite(pSat *sat_solver, pBeg *int32, pEnd *int32, fRoot int32) {
    if pSat.pFile == nil {
        return; }
    pSat.nClauses++;
    pSat.nRoots += fRoot;
    for  ; pBeg < pEnd ; pBeg++  {
        fmt.Fprintf(pSat.pFile, " %d", lit_print(*pBeg));
    }
    fmt.Fprintf(pSat.pFile, " 0\n");
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


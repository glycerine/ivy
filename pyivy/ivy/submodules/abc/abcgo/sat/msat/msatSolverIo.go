package msat

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [msatSolverIo.c]

  PackageName [A C version of SAT solver MINISAT, originally developed 
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of 
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [Input/output of CNFs.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatSolverIo.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

//#include "msatInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static char * Msat_TimeStamp();


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverPrintAssignment(p *Msat_Solver_t) {
    var i int32;
    fmt.Printf("Current assignments are: \n");
    for  i = 0; i < p.nVars; i++  {
        fmt.Printf("%d", i % 10);
    }
    fmt.Printf("\n");
    for  i = 0; i < p.nVars; i++  {
        if p.pAssigns[i] == MSAT_VAR_UNASSIGNED {
            fmt.Printf(".");
        } else {
            // inlining of macro MSAT_LIT2VAR
            assert(i == ((p.pAssigns[i])>>1), "i == /*__starting_point_inline_macro:MSAT_LIT2VAR__*/((p.pAssigns[i])>>1)");
            // inlining of macro MSAT_LITSIGN
            if  (p.pAssigns[i])&1 {
                fmt.Printf("0");
            } else { 
                fmt.Printf("1"); }
        }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverPrintClauses(p *Msat_Solver_t) {
    var pClauses **Msat_Clause_t;
    var nClauses, i int32;

    fmt.Printf("Original clauses: \n");
    nClauses = Msat_ClauseVecReadSize( p.vClauses );
    pClauses = Msat_ClauseVecReadArray( p.vClauses );
    for  i = 0; i < nClauses; i++  {
        fmt.Printf("%3d: ", i);
        Msat_ClausePrint( pClauses[i] );
    }

    fmt.Printf("Learned clauses: \n");
    nClauses = Msat_ClauseVecReadSize( p.vLearned );
    pClauses = Msat_ClauseVecReadArray( p.vLearned );
    for  i = 0; i < nClauses; i++  {
        fmt.Printf("%3d: ", i);
        Msat_ClausePrint( pClauses[i] );
    }

    fmt.Printf("Variable activity: \n");
    for  i = 0; i < p.nVars; i++  {
        fmt.Printf("%3d : %.4f\n", i, p.pdActivity[i]);
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverWriteDimacs(p *Msat_Solver_t, pFileName string) {
    var pFile *os.File;
    var pClauses **Msat_Clause_t;
    var nClauses, i int32;

    nClauses = Msat_ClauseVecReadSize(p.vClauses) + Msat_ClauseVecReadSize(p.vLearned);
    for  i = 0; i < p.nVars; i++  {
        nClauses += ( p.pLevel[i] == 0 );
    }

    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pFile, "c Produced by Msat_SolverWriteDimacs() on %s\n", Msat_TimeStamp());
    fmt.Fprintf(pFile, "p cnf %d %d\n", p.nVars, nClauses);

    nClauses = Msat_ClauseVecReadSize( p.vClauses );
    pClauses = Msat_ClauseVecReadArray( p.vClauses );
    for  i = 0; i < nClauses; i++  {
        Msat_ClauseWriteDimacs( pFile, pClauses[i], 1 );
    }

    nClauses = Msat_ClauseVecReadSize( p.vLearned );
    pClauses = Msat_ClauseVecReadArray( p.vLearned );
    for  i = 0; i < nClauses; i++  {
        Msat_ClauseWriteDimacs( pFile, pClauses[i], 1 );
    }

    // write zero-level assertions
    for  i = 0; i < p.nVars; i++  {
        if p.pLevel[i] == 0 {
            fmt.Fprintf(pFile, "%s%d 0\n", ((p.pAssigns[i]&1)? "-": ""), i + 1); }
    }

    fmt.Fprintf(pFile, "\n");
    _ = (pFile).Close();
}


/**Function*************************************************************

  Synopsis    [Returns the time stamp.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Msat_TimeStamp_Buffer [100]byte;
func Msat_TimeStamp() string {
    
    var ltime int64;
    var TimeStamp string;
    // get the current time
    time( &ltime );
    TimeStamp = asctime( localtime( &ltime ) );
    TimeStamp[ len(TimeStamp) - 1 ] = 0;
    strcpy( spStatic_Msat_TimeStamp_Buffer, TimeStamp );
    return spStatic_Msat_TimeStamp_Buffer;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


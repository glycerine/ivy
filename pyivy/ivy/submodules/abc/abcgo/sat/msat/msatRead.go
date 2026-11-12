package msat

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [msatRead.c]

  PackageName [A C version of SAT solver MINISAT, originally developed 
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of 
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [The reader of the CNF formula in DIMACS format.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatRead.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

//#include "msatInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static char * Msat_FileRead( FILE * pFile );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Read the file into the internal buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_FileRead(pFile *os.File) string {
    var nFileSize int32;
    var pBuffer string;
    var RetValue int32;
    // get the file size, in bytes
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile );  
    // move the file current reading position to the beginning
    rewind( pFile ); 
    // load the contents of the file into memory
    // inlining of macro ABC_ALLOC
    pBuffer = ((char *) malloc(sizeof(char) * (size_t)(nFileSize + 3)));
    RetValue = fread( pBuffer, nFileSize, 1, pFile );
    // terminate the string with '\0'
    pBuffer[ nFileSize + 0] = '\n';
    pBuffer[ nFileSize + 1] = '\0';
    return pBuffer;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ReadWhitespace(pIn *string) {
    for (**pIn >= 9 && **pIn <= 13) || **pIn == 32 {
        (*pIn)++;
    } 
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ReadNotWhitespace(pIn *string) {
    for  !((**pIn >= 9 && **pIn <= 13) || **pIn == 32)  {
        (*pIn)++;
    } 
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func skipLine(pIn *string) {
    for {
        if **pIn == 0 { 
            return; }
        if **pIn == '\n' { 
            (*pIn)++; 
            return; 
        }
        (*pIn)++; 
    } 
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ReadInt(pIn *string) int32 {
    var val int32;
    var neg int32;

    Msat_ReadWhitespace( pIn );
    if **pIn == '-' { 
        neg = 1;
(*pIn)++;
    } else if **pIn == '+' { 
        (*pIn)++; }
    if **pIn < '0' || **pIn > '9' { 
        fmt.Fprintf(os.Stderr, "PARSE ERROR! Unexpected char: %c\n", **pIn);
os.Exit(int(1)); }
    for  **pIn >= '0' && **pIn <= '9'  {
        val = val*10 + (**pIn - '0');
(*pIn)++;
    }
    if neg != 0 {
        return -val;
    } else {
        return val;
    } 
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ReadClause(pIn *string, p *Msat_Solver_t, pLits *Msat_IntVec_t) {
    var nVars int32  = Msat_SolverReadVarNum( p );
    var parsed_lit, var, sign int32;

    Msat_IntVecClear( pLits );
    for {
        parsed_lit = Msat_ReadInt(pIn);
        if parsed_lit == 0 { 
            break; }
        var = abs(parsed_lit) - 1;
        sign = (parsed_lit > 0);
        if var >= nVars {
            fmt.Printf("Variable %d is larger than the number of allocated variables (%d).\n", var+1, nVars);
            os.Exit(int(1));
        }
        // inlining of macro MSAT_VAR2LIT
        Msat_IntVecPush( pLits, (2*(var)+(!sign))  );
    }
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ReadDimacs(pText string, pS **Msat_Solver_t, fVerbose int32) int32 {
    var p *Msat_Solver_t  = nil; // Suppress "might be used uninitialized"
    var pLits *Msat_IntVec_t  = nil; // Suppress "might be used uninitialized"
    var pIn string  = pText;
    var nVars, nClas int32;
    for {
        Msat_ReadWhitespace( &pIn );
        if *pIn == 0 {
            break;
        } else if *pIn == 'c' {
            skipLine( &pIn );
        } else if *pIn == 'p' {
            pIn++;
            Msat_ReadWhitespace( &pIn );
            Msat_ReadNotWhitespace( &pIn );

            nVars = Msat_ReadInt( &pIn );
            nClas = Msat_ReadInt( &pIn );
            skipLine( &pIn );
            // start the solver
            p = Msat_SolverAlloc( nVars, 1, 1, 1, 1, 0 ); 
            Msat_SolverClean( p, nVars );
            Msat_SolverSetVerbosity( p, fVerbose );
            // allocate the vector
            pLits = Msat_IntVecAlloc( nVars );
        
        } else {
            if p == nil {
                fmt.Printf("There is no parameter line.\n");
                os.Exit(int(1));
            }
            Msat_ReadClause( &pIn, p, pLits );
            if !Msat_SolverAddClause( p, pLits ) {
                return 0; }
        }
    }
    Msat_IntVecFree( pLits );
    *pS = p;
    return Msat_SolverSimplifyDB( p );
}

/**Function*************************************************************

  Synopsis    [Starts the solver and reads the DIMAC file.]

  Description [Returns FALSE upon immediate conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverParseDimacs(pFile *os.File, p **Msat_Solver_t, fVerbose int32) int32 {
    var pText string;
    var Value int32;
    pText = Msat_FileRead( pFile );
    Value = Msat_ReadDimacs( pText, p, fVerbose );
    // inlining of macro ABC_FREE
    ((pText) ? (free((char *) (pText)), (pText) = 0) : 0);
    return Value;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


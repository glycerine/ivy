package xsat

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [xsatCnfReader.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [xSAT - A SAT solver written in C.
               Read the license file for more info.]

  Synopsis    [CNF DIMACS file format parser.]

  Author      [Bruno Schmitt <boschmitt@inf.ufrgs.br>]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - November 10, 2016.]

  Revision    []

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
//#include <ctype.h>

//#include "misc/util/abc_global.h"
//#include "misc/vec/vecInt.h"

//#include "xsatSolver.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Read the file into the internal buffer.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_FileRead(pFile *os.File) string {
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
func xSAT_ReadInt(pIn *string) int32 {
    var val int32;
    var neg int32;

    for ; isspace(**pIn); (*pIn)++ {}
    if **pIn == '-' {
        neg = 1;
(*pIn)++;
    } else if **pIn == '+' {
        (*pIn)++; }
    if !isdigit(**pIn) {
        fmt.Fprintf(os.Stderr, "PARSE ERROR! Unexpected char: %c\n", **pIn);
os.Exit(int(1)); }
    for  isdigit(**pIn)  {
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
func xSAT_ReadClause(pIn *string, p *xSAT_Solver_t, vLits *Vec_Int_t) {
    var token, var, sign int32;

    Vec_IntClear( vLits );
    for {
        token = xSAT_ReadInt( pIn );
        if token == 0 {
            break; }
        var = abs(token) - 1;
        sign = (token > 0);
        Vec_IntPush( vLits, xSAT_Var2Lit( var, !sign ) );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_ParseDimacs(pText string, pS **xSAT_Solver_t) int32 {
    var p *xSAT_Solver_t  = nil;
    var vLits *Vec_Int_t  = nil;
    var pIn string  = pText;
    var nVars, nClas int32;
    for {
        for ; isspace(*pIn); pIn++ {}
        if *pIn == 0 {
            break;
        } else if *pIn == 'c' {
            skipLine( &pIn );
        } else if *pIn == 'p' {
            pIn++;
            for ; isspace(*pIn); pIn++ {}
            for ; !isspace(*pIn); pIn++ {}

            nVars = xSAT_ReadInt( &pIn );
            nClas = xSAT_ReadInt( &pIn );
            skipLine( &pIn );

            /* start the solver */
            p = xSAT_SolverCreate();
            /* allocate the vector */
            vLits = Vec_IntAlloc( nVars );
        
        } else {
            if p == nil {
                fmt.Printf("There is no parameter line.\n");
                os.Exit(int(1));
            }
            xSAT_ReadClause( &pIn, p, vLits );
            if !xSAT_SolverAddClause( p, vLits ) {
                Vec_IntPrint(vLits);
                return 0;
            }
        }
    }
    Vec_IntFree( vLits );
    *pS = p;
    return xSAT_SolverSimplify( p );
}

/**Function*************************************************************

  Synopsis    [Starts the solver and reads the DIMAC file.]

  Description [Returns FALSE upon immediate conflict.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_SolverParseDimacs(pFile *os.File, p **xSAT_Solver_t) int32 {
    var pText string;
    var Value int32;
    pText = xSAT_FileRead( pFile );
    Value = xSAT_ParseDimacs( pText, p );
    pText = nil;
    return Value;
}

//ABC_NAMESPACE_IMPL_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

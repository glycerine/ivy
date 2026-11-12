package bsat

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [satChecker.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sat_solver.]

  Synopsis    [Resolution proof checker.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satChecker.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_PrintClause(vClauses *Vec_Vec_t, Clause int32) {
    var vClause *Vec_Int_t;
    var i, Entry int32;
    fmt.Printf("Clause %d:  {", Clause);
    vClause = Vec_VecEntry( vClauses, Clause );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClause)) && (((Entry) = Vec_IntEntry(vClause, i)), 1); i++  {
        fmt.Printf(" %d", Entry);
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_ProofResolve(vClauses *Vec_Vec_t, Result int32, Clause1 int32, Clause2 int32) int32 {
    var vResult *Vec_Int_t   = Vec_VecEntry( vClauses, Result );
    var vClause1 *Vec_Int_t  = Vec_VecEntry( vClauses, Clause1 );
    var vClause2 *Vec_Int_t  = Vec_VecEntry( vClauses, Clause2 );
    var Entry1, Entry2, ResVar int32;
    var i, j, Counter int32;

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClause1)) && (((Entry1) = Vec_IntEntry(vClause1, i)), 1); i++  {
    // inlining of macro Vec_IntForEachEntry
    for  j = 0; (j < Vec_IntSize(vClause2)) && (((Entry2) = Vec_IntEntry(vClause2, j)), 1); j++  {
    if Entry1 == -Entry2 {
        ResVar = Entry1;
        Counter++;
    }
    }
    }
    if Counter != 1 {
        fmt.Printf("Error: Clause %d = Resolve(%d, %d): The number of pivot vars is %d.\n", Result, Clause1, Clause2, Counter);
        Sat_PrintClause( vClauses, Clause1 );
        Sat_PrintClause( vClauses, Clause2 );
        return 0;
    }
    // create new clause
    assert(Vec_IntSize(vResult) == 0, "Vec_IntSize(vResult) == 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClause1)) && (((Entry1) = Vec_IntEntry(vClause1, i)), 1); i++  {
        if Entry1 != ResVar && Entry1 != -ResVar {
            Vec_IntPushUnique( vResult, Entry1 ); }
    }
    assert(Vec_IntSize(vResult) + 1 == Vec_IntSize(vClause1), "Vec_IntSize(vResult) + 1 == Vec_IntSize(vClause1)");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClause2)) && (((Entry2) = Vec_IntEntry(vClause2, i)), 1); i++  {
        if Entry2 != ResVar && Entry2 != -ResVar {
            Vec_IntPushUnique( vResult, Entry2 ); }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sat_ProofChecker(pFileName string) {
    var pFile *os.File;
    var vClauses *Vec_Vec_t;
    var c, i, Num, RetValue, Counter, Counter2, Clause1, Clause2 int32;
    var RetValue int32;
    // open the file
    pFile = os.Open(pFileName);
    if pFile == nil {
        return; }
    // count the number of clauses
    Counter = Counter2 = 0;
    for  (c = fgetc(pFile)) != -1  {
        Counter += (c == '\n');
        Counter2 += (c == '*');
    }
    vClauses = Vec_VecStart( Counter+1 );
    fmt.Printf("The proof contains %d roots and %d resolution steps.\n", Counter-Counter2, Counter2);
    // read the clauses
    rewind( pFile );
    for  i = 1 ; ; i++  {
        RetValue = RetValue = fscanf( pFile, "%d", &Num );
        if RetValue != 1 {
            break; }
        assert(Num == i, "Num == i");
        for  (c = fgetc( pFile )) == ' '  {; }
        if c == '*' {
            RetValue = fscanf( pFile, "%d %d", &Clause1, &Clause2 );
            assert(RetValue == 2, "RetValue == 2");
            RetValue = fscanf( pFile, "%d", &Num );
            assert(RetValue == 1, "RetValue == 1");
            assert(Num == 0, "Num == 0");
            if !(Sat_ProofResolve( vClauses, i, Clause1, Clause2 ) != 0) {
                fmt.Printf("Error detected in the resolution proof.\n");
                Vec_VecFree( vClauses );
                _ = (pFile).Close();
                return;
            }
        
        } else {
            ungetc( c, pFile );
            for {
                RetValue = fscanf( pFile, "%d", &Num );
                assert(RetValue == 1, "RetValue == 1");
                if Num == 0 {
                    break; }
                Vec_VecPush( vClauses, i, (void *)Num );
            }
            RetValue = fscanf( pFile, "%d", &Num );
            assert(RetValue == 1, "RetValue == 1");
            assert(Num == 0, "Num == 0");
        }
    }
    assert(i-1 == Counter, "i-1 == Counter");
    if Vec_IntSize( Vec_VecEntry(vClauses, Counter) ) != 0 {
        fmt.Printf("The last clause is not empty.\n");
    } else {
        fmt.Printf("The empty clause is derived.\n"); }
    Vec_VecFree( vClauses );
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


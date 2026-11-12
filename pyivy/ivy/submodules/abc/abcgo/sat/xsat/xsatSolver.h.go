package xsat

import "fmt"

/**CFile****************************************************************

  FileName    [xsatSolver.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [xSAT - A SAT solver written in C.
               Read the license file for more info.]

  Synopsis    [Internal definitions of the solver.]

  Author      [Bruno Schmitt <boschmitt@inf.ufrgs.br>]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - November 10, 2016.]

  Revision    []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/util/abc_global.h"
//#include "misc/vec/vecStr.h"

//#include "xsat.h"
//#include "xsatBQueue.h"
//#include "xsatClause.h"
//#include "xsatHeap.h"
//#include "xsatMemory.h"
//#include "xsatWatchList.h"

ABC_NAMESPACE_HEADER_START

#ifndef __cplusplus
#ifndef false
#  define false 0
#endif
#ifndef true
#  define true 1
#endif
#endif

const
(
    Var0 = 1

    Var1 = 0

    VarX = 3

)


const
(
    LBoolUndef =  0

    LBoolTrue  =  1

    LBoolFalse = -1

)


const
(
    VarUndef = -1

    LitUndef = -2

)


const CRefUndef = 0xFFFFFFFF

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
   
type xSAT_SolverOptions_t struct {
    fVerbose byte;

    // Limits
    nConfLimit   iword;    // external limit on the number of conflicts
    nInsLimit   iword;     // external limit on the number of implications
    nRuntimeLimit abctime; // external limit on runtime

    // Constants used for restart heuristic
    K float64; // Forces a restart
    R float64; // Block a restart
    nFirstBlockRestart int32; // Lower bound number of conflicts for start blocking restarts
    nSizeLBDQueue int32;      // Size of the moving avarege queue for LBD (force restart)
    nSizeTrailQueue int32;    // Size of the moving avarege queue for Trail size (block restart)

    // Constants used for clause database reduction heuristic
    nConfFirstReduce int32;  // Number of conflicts before first reduction
    nIncReduce int32;        // Increment to reduce
    nSpecialIncReduce int32; // Special increment to reduce
    nLBDFrozenClause uint32;
}

   
type xSAT_Stats_t struct {
    nStarts uint32;
    nReduceDB uint32;

    nDecisions iword;
    nPropagations iword;
    nInspects iword;
    nConflicts iword;

    nClauseLits iword;
    nLearntLits iword;
}

type xSAT_Solver_t_ struct {
    /* Clauses Database */
    pMemory  *xSAT_Mem_t;
    vLearnts   *Vec_Int_t;
    vClauses   *Vec_Int_t;
    vWatches  *xSAT_VecWatchList_t;
    vBinWatches  *xSAT_VecWatchList_t;

    /* Activity heuristic */
    nVarActInc int32;       /* Amount to bump next variable with. */
    nClaActInc int32;       /* Amount to bump next clause with. */

    /* Variable Information */
    vActivity  *Vec_Int_t;      /* A heuristic measurement of the activity of a variable. */
    hOrder  *xSAT_Heap_t;
    vLevels  *Vec_Int_t;        /* Decision level of the current assignment */
    vReasons  *Vec_Int_t;       /* Reason (clause) of the current assignment */
    vAssigns  *Vec_Str_t;       /* Current assignment. */
    vPolarity  *Vec_Str_t;
    vTags  *Vec_Str_t;

    /* Assignments */
    vTrail  *Vec_Int_t;
    vTrailLim  *Vec_Int_t; // Separator indices for different decision levels in 'trail'.
    iQhead int32;            // Head of propagation queue (as index into the trail).

    nAssignSimplify     int32; /* Number of top-level assignments since last
                              * execution of 'simplify()'. */
    nPropSimplify   iword;   /* Remaining number of propagations that must be
                              * made before next execution of 'simplify()'. */

    /* Temporary data used by Search method */
    bqTrail  *xSAT_BQueue_t;
    bqLBD  *xSAT_BQueue_t;
    nSumLBD float32;
    nConfBeforeReduce  int32;
    nRC1 long;
    nRC2  int32;

    /* Temporary data used by Analyze */
    vLearntClause  *Vec_Int_t;
    vSeen  *Vec_Str_t;
    vTagged  *Vec_Int_t;
    vStack  *Vec_Int_t;
    vLastDLevel  *Vec_Int_t;

    /* Misc temporary */
    nStamp uint32;
    vStamp  *Vec_Int_t;        /* Multipurpose stamp used to calculate LBD and
                                * clauses minimization with binary resolution */

    Config xSAT_SolverOptions_t;
    Stats xSAT_Stats_t;
}

func xSAT_Var2Lit(Var int32, c int32) int32 {
    return Var + Var + ( c != 0 );
}

func xSAT_NegLit(Lit int32) int32 {
    return Lit ^ 1;
}

func xSAT_Lit2Var(Lit int32) int32 {
    return Lit >> 1;
}

func xSAT_LitSign(Lit int32) int32 {
    return Lit & 1;
}

func xSAT_SolverDecisionLevel(s *xSAT_Solver_t) int32 {
    return Vec_IntSize( s.vTrailLim );
}

func xSAT_SolverReadClause(s *xSAT_Solver_t, h uint32) *xSAT_Clause_t {
    return xSAT_MemClauseHand( s.pMemory, h );
}

func xSAT_SolverIsClauseSatisfied(s *xSAT_Solver_t, pCla *xSAT_Clause_t) int32 {
    var i int32;
    var Lits *int32  = &( pCla.pData[0].Lit );

    for  i = 0; i < pCla.nSize; i++  {
        if Vec_StrEntry( s.vAssigns, xSAT_Lit2Var( Lits[i] ) ) == xSAT_LitSign( ( Lits[i] ) ) {
            return true; }
    }

    return false;
}

func xSAT_SolverPrintClauses(s *xSAT_Solver_t) {
    var i int32;
    var CRef uint32;

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(s.vClauses)) && (((CRef) = Vec_IntEntry(s.vClauses, i)), 1); i++  {
        xSAT_ClausePrint( xSAT_SolverReadClause( s, CRef ) );
    }
}

func xSAT_SolverPrintState(s *xSAT_Solver_t) {
    fmt.Printf("starts        : %10d\n", s.Stats.nStarts);
    fmt.Printf("conflicts     : %10d\n", s.Stats.nConflicts);
    fmt.Printf("decisions     : %10d\n", s.Stats.nDecisions);
    fmt.Printf("propagations  : %10d\n", s.Stats.nPropagations);
}


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DECLARATIONS                        ///
////////////////////////////////////////////////////////////////////////
// extern unsigned xSAT_SolverClaNew( xSAT_Solver_t* s, Vec_Int_t * vLits, int fLearnt );

// extern char    xSAT_SolverSearch( xSAT_Solver_t * s );


// extern void xSAT_SolverGarbageCollect( xSAT_Solver_t * s );


// extern int xSAT_SolverEnqueue( xSAT_Solver_t* s, int Lit, unsigned From );

// extern void xSAT_SolverCancelUntil( xSAT_Solver_t* s, int Level);

// extern unsigned xSAT_SolverPropagate( xSAT_Solver_t* s );

// extern void xSAT_SolverRebuildOrderHeap( xSAT_Solver_t* s );


//ABC_NAMESPACE_HEADER_END



package msat

/**CFile****************************************************************

  FileName    [msatInt.c]

  PackageName [A C version of SAT solver MINISAT, originally developed 
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of 
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [Internal definitions of the solver.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatInt.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>
//#include <math.h>

//#include "misc/util/abc_global.h"
//#include "msat.h"

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

// By default, custom memory management is used
// which guarantees constant time allocation/deallocation 
// for SAT clauses and other frequently modified objects.
// For debugging, it is possible use system memory management
// directly. In which case, uncomment the macro below.
//#define USE_SYSTEM_MEMORY_MANAGEMENT 

// internal data structures
type Msat_Clause_t =      Msat_Clause_t_;
type Msat_Queue_t =       Msat_Queue_t_;
type Msat_Order_t =       Msat_Order_t_;
// memory managers (duplicated from Extra for stand-aloneness)
type Msat_MmFixed_t =     Msat_MmFixed_t_;    
type Msat_MmFlex_t =      Msat_MmFlex_t_;     
type Msat_MmStep_t =      Msat_MmStep_t_;     
// variables and literals
typedef int                       Msat_Lit_t;  
typedef int                       Msat_Var_t;  
// the type of return value
#define MSAT_VAR_UNASSIGNED (-1)
#define MSAT_LIT_UNASSIGNED (-2)
#define MSAT_ORDER_UNKNOWN  (-3)

// printing the search tree
const L_IND =      "%-*d"
#define L_ind      Msat_SolverReadDecisionLevel(p)*3+3,Msat_SolverReadDecisionLevel(p)
const L_LIT =      "%s%d"
// Original macro:
// #define L_lit(Lit) MSAT_LITSIGN(Lit)?"-":"", MSAT_LIT2VAR(Lit)+1


   
type Msat_SolverStats_t struct {
    nStarts   ABC_INT64_T;        // the number of restarts
    nDecisions   ABC_INT64_T;     // the number of decisions
    nPropagations   ABC_INT64_T;  // the number of implications
    nInspects   ABC_INT64_T;      // the number of times clauses are vising while watching them
    nConflicts   ABC_INT64_T;     // the number of conflicts
    nSuccesses   ABC_INT64_T;     // the number of sat assignments found
}

   
type Msat_SearchParams_t struct {
    dVarDecay  float64;
    dClaDecay  float64;
}

// sat solver data structure visible through all the internal files
type Msat_Solver_t_ struct {
    nClauses                 int32;    // the total number of clauses
    nClausesStart                 int32; // the number of clauses before adding
    vClauses   *Msat_ClauseVec_t;    // problem clauses
    vLearned   *Msat_ClauseVec_t;    // learned clauses
    dClaInc              float64;     // Amount to bump next clause with.
    dClaDecay              float64;   // INVERSE decay factor for clause activity: stores 1/decay.

    pdActivity             *float64;  // A heuristic measurement of the activity of a variable.
    pFactors              *float32;    // the multiplicative factors of variable activity
    dVarInc              float64;     // Amount to bump next variable with.
    dVarDecay              float64;   // INVERSE decay factor for variable activity: stores 1/decay. Use negative value for static variable order.
    pOrder       *Msat_Order_t;      // Keeps track of the decision variable order.

    pvWatched  **Msat_ClauseVec_t;   // 'pvWatched[lit]' is a list of constraints watching 'lit' (will go there if literal becomes true).
    pQueue       *Msat_Queue_t;      // Propagation queue.

    nVars                 int32;       // the current number of variables
    nVarsAlloc                 int32;  // the maximum allowed number of variables
    pAssigns                *int32;    // The current assignments (literals or MSAT_VAR_UNKOWN)
    pModel                *int32;      // The satisfying assignment
    vTrail      *Msat_IntVec_t;      // List of assignments made. 
    vTrailLim      *Msat_IntVec_t;   // Separator indices for different decision levels in 'trail'.
    pReasons     **Msat_Clause_t;    // 'reason[var]' is the clause that implied the variables current value, or 'NULL' if none.
    pLevel                *int32;      // 'level[var]' is the decision level at which assignment was made. 
    nLevelRoot                 int32;  // Level of first proper decision.

    dRandSeed              float64;   // For the internal random number generator (makes solver deterministic over different platforms). 

    fVerbose                 int32;    // the verbosity flag
    dProgress              float64;   // Set by 'search()'.

    // the variable cone and variable connectivity
    vConeVars      *Msat_IntVec_t;
    vVarsUsed      *Msat_IntVec_t;
    vAdjacents   *Msat_ClauseVec_t;

    // internal data used during conflict analysis
    pSeen                *int32;       // time when a lit was seen for the last time
    nSeenId                 int32;     // the id of current seeing
    vReason      *Msat_IntVec_t;     // the temporary array of literals
    vTemp      *Msat_IntVec_t;       // the temporary array of literals
    pFreq                *int32;       // the number of times each var participated in conflicts

    // the memory manager
    pMem      *Msat_MmStep_t;

    // statistics
    Stats  Msat_SolverStats_t;
    nTwoLits                 int32;
    nTwoLitsL                 int32;
    nClausesInit                 int32;
    nClausesAlloc                 int32;
    nClausesAllocL                 int32;
    nBackTracks                 int32;
}

type Msat_ClauseVec_t_ struct {
    pArray     **Msat_Clause_t;
    nSize                 int32;
    nCap                 int32;
}

type Msat_IntVec_t_ struct {
    pArray                *int32;
    nSize                 int32;
    nCap                 int32;
}

////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DECLARATIONS                        ///
////////////////////////////////////////////////////////////////////////

/*=== satActivity.c ===========================================================*/
// extern void                Msat_SolverVarDecayActivity( Msat_Solver_t * p );

// extern void                Msat_SolverVarRescaleActivity( Msat_Solver_t * p );

// extern void                Msat_SolverClaDecayActivity( Msat_Solver_t * p );

// extern void                Msat_SolverClaRescaleActivity( Msat_Solver_t * p );

/*=== satSolverApi.c ===========================================================*/
// extern Msat_Clause_t *     Msat_SolverReadClause( Msat_Solver_t * p, int Num );

/*=== satSolver.c ===========================================================*/
// extern int                 Msat_SolverReadDecisionLevel( Msat_Solver_t * p );

// extern int *               Msat_SolverReadDecisionLevelArray( Msat_Solver_t * p );

// extern Msat_Clause_t **    Msat_SolverReadReasonArray( Msat_Solver_t * p );

// extern Msat_Type_t         Msat_SolverReadVarValue( Msat_Solver_t * p, Msat_Var_t Var );

// extern Msat_ClauseVec_t *  Msat_SolverReadLearned( Msat_Solver_t * p );

// extern Msat_ClauseVec_t ** Msat_SolverReadWatchedArray( Msat_Solver_t * p );

// extern int *               Msat_SolverReadSeenArray( Msat_Solver_t * p );

// extern int                 Msat_SolverIncrementSeenId( Msat_Solver_t * p );

// extern Msat_MmStep_t *     Msat_SolverReadMem( Msat_Solver_t * p );

// extern void                Msat_SolverClausesIncrement( Msat_Solver_t * p );

// extern void                Msat_SolverClausesDecrement( Msat_Solver_t * p );

// extern void                Msat_SolverClausesIncrementL( Msat_Solver_t * p );

// extern void                Msat_SolverClausesDecrementL( Msat_Solver_t * p );

// extern void                Msat_SolverVarBumpActivity( Msat_Solver_t * p, Msat_Lit_t Lit );

// extern void                Msat_SolverClaBumpActivity( Msat_Solver_t * p, Msat_Clause_t * pC );

// extern int                 Msat_SolverEnqueue( Msat_Solver_t * p, Msat_Lit_t Lit, Msat_Clause_t * pC );

// extern double              Msat_SolverProgressEstimate( Msat_Solver_t * p );

/*=== satSolverSearch.c ===========================================================*/
// extern int                 Msat_SolverAssume( Msat_Solver_t * p, Msat_Lit_t Lit );

// extern Msat_Clause_t *     Msat_SolverPropagate( Msat_Solver_t * p );

// extern void                Msat_SolverCancelUntil( Msat_Solver_t * p, int Level );

// extern Msat_Type_t         Msat_SolverSearch( Msat_Solver_t * p, int nConfLimit, int nLearnedLimit, int nBackTrackLimit, Msat_SearchParams_t * pPars );

/*=== satQueue.c ===========================================================*/
// extern Msat_Queue_t *      Msat_QueueAlloc( int nVars );

// extern void                Msat_QueueFree( Msat_Queue_t * p );

// extern int                 Msat_QueueReadSize( Msat_Queue_t * p );

// extern void                Msat_QueueInsert( Msat_Queue_t * p, int Lit );

// extern int                 Msat_QueueExtract( Msat_Queue_t * p );

// extern void                Msat_QueueClear( Msat_Queue_t * p );

/*=== satOrder.c ===========================================================*/
// extern Msat_Order_t *      Msat_OrderAlloc( Msat_Solver_t * pSat );

// extern void                Msat_OrderSetBounds( Msat_Order_t * p, int nVarsMax );

// extern void                Msat_OrderClean( Msat_Order_t * p, Msat_IntVec_t * vCone );

// extern int                 Msat_OrderCheck( Msat_Order_t * p );

// extern void                Msat_OrderFree( Msat_Order_t * p );

// extern int                 Msat_OrderVarSelect( Msat_Order_t * p );

// extern void                Msat_OrderVarAssigned( Msat_Order_t * p, int Var );

// extern void                Msat_OrderVarUnassigned( Msat_Order_t * p, int Var );

// extern void                Msat_OrderUpdate( Msat_Order_t * p, int Var );

/*=== satClause.c ===========================================================*/
// extern int                 Msat_ClauseCreate( Msat_Solver_t * p, Msat_IntVec_t * vLits, int  fLearnt, Msat_Clause_t ** pClause_out );

// extern Msat_Clause_t *     Msat_ClauseCreateFake( Msat_Solver_t * p, Msat_IntVec_t * vLits );

// extern Msat_Clause_t *     Msat_ClauseCreateFakeLit( Msat_Solver_t * p, Msat_Lit_t Lit );

// extern int                 Msat_ClauseReadLearned( Msat_Clause_t * pC );

// extern int                 Msat_ClauseReadSize( Msat_Clause_t * pC );

// extern int *               Msat_ClauseReadLits( Msat_Clause_t * pC );

// extern int                 Msat_ClauseReadMark( Msat_Clause_t * pC );

// extern void                Msat_ClauseSetMark( Msat_Clause_t * pC, int  fMark );

// extern int                 Msat_ClauseReadNum( Msat_Clause_t * pC );

// extern void                Msat_ClauseSetNum( Msat_Clause_t * pC, int Num );

// extern int                 Msat_ClauseReadTypeA( Msat_Clause_t * pC );

// extern void                Msat_ClauseSetTypeA( Msat_Clause_t * pC, int  fTypeA );

// extern int                 Msat_ClauseIsLocked( Msat_Solver_t * p, Msat_Clause_t * pC );

// extern float               Msat_ClauseReadActivity( Msat_Clause_t * pC );

// extern void                Msat_ClauseWriteActivity( Msat_Clause_t * pC, float Num );

// extern void                Msat_ClauseFree( Msat_Solver_t * p, Msat_Clause_t * pC, int  fRemoveWatched );

// extern int                 Msat_ClausePropagate( Msat_Clause_t * pC, Msat_Lit_t Lit, int * pAssigns, Msat_Lit_t * pLit_out );

// extern int                 Msat_ClauseSimplify( Msat_Clause_t * pC, int * pAssigns );

// extern void                Msat_ClauseCalcReason( Msat_Solver_t * p, Msat_Clause_t * pC, Msat_Lit_t Lit, Msat_IntVec_t * vLits_out );

// extern void                Msat_ClauseRemoveWatch( Msat_ClauseVec_t * vClauses, Msat_Clause_t * pC );

// extern void                Msat_ClausePrint( Msat_Clause_t * pC );

// extern void                Msat_ClausePrintSymbols( Msat_Clause_t * pC );

// extern void                Msat_ClauseWriteDimacs( FILE * pFile, Msat_Clause_t * pC, int  fIncrement );

// extern unsigned            Msat_ClauseComputeTruth( Msat_Solver_t * p, Msat_Clause_t * pC );

/*=== satSort.c ===========================================================*/
// extern void                Msat_SolverSortDB( Msat_Solver_t * p );

/*=== satClauseVec.c ===========================================================*/
// extern Msat_ClauseVec_t *  Msat_ClauseVecAlloc( int nCap );

// extern void                Msat_ClauseVecFree( Msat_ClauseVec_t * p );

// extern Msat_Clause_t **    Msat_ClauseVecReadArray( Msat_ClauseVec_t * p );

// extern int                 Msat_ClauseVecReadSize( Msat_ClauseVec_t * p );

// extern void                Msat_ClauseVecGrow( Msat_ClauseVec_t * p, int nCapMin );

// extern void                Msat_ClauseVecShrink( Msat_ClauseVec_t * p, int nSizeNew );

// extern void                Msat_ClauseVecClear( Msat_ClauseVec_t * p );

// extern void                Msat_ClauseVecPush( Msat_ClauseVec_t * p, Msat_Clause_t * Entry );

// extern Msat_Clause_t *     Msat_ClauseVecPop( Msat_ClauseVec_t * p );

// extern void                Msat_ClauseVecWriteEntry( Msat_ClauseVec_t * p, int i, Msat_Clause_t * Entry );

// extern Msat_Clause_t *     Msat_ClauseVecReadEntry( Msat_ClauseVec_t * p, int i );


/*=== satMem.c ===========================================================*/
// fixed-size-block memory manager
// extern Msat_MmFixed_t *    Msat_MmFixedStart( int nEntrySize );

// extern void                Msat_MmFixedStop( Msat_MmFixed_t * p, int fVerbose );

// extern char *              Msat_MmFixedEntryFetch( Msat_MmFixed_t * p );

// extern void                Msat_MmFixedEntryRecycle( Msat_MmFixed_t * p, char * pEntry );

// extern void                Msat_MmFixedRestart( Msat_MmFixed_t * p );

// extern int                 Msat_MmFixedReadMemUsage( Msat_MmFixed_t * p );

// flexible-size-block memory manager
// extern Msat_MmFlex_t *     Msat_MmFlexStart();

// extern void                Msat_MmFlexStop( Msat_MmFlex_t * p, int fVerbose );

// extern char *              Msat_MmFlexEntryFetch( Msat_MmFlex_t * p, int nBytes );

// extern int                 Msat_MmFlexReadMemUsage( Msat_MmFlex_t * p );

// hierarchical memory manager
// extern Msat_MmStep_t *     Msat_MmStepStart( int nSteps );

// extern void                Msat_MmStepStop( Msat_MmStep_t * p, int fVerbose );

// extern char *              Msat_MmStepEntryFetch( Msat_MmStep_t * p, int nBytes );

// extern void                Msat_MmStepEntryRecycle( Msat_MmStep_t * p, char * pEntry, int nBytes );

// extern int                 Msat_MmStepReadMemUsage( Msat_MmStep_t * p );


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_HEADER_END



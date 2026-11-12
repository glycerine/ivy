package msat

/**CFile****************************************************************

  FileName    [msatSolverApi.c]

  PackageName [A C version of SAT solver MINISAT, originally developed 
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of 
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [APIs of the SAT solver.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatSolverApi.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

//#include "msatInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Msat_SolverSetupTruthTables( unsigned uTruths[][2] );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Simple SAT solver APIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverReadVarNum(p *Msat_Solver_t) int32 { return p.nVars;      }
func Msat_SolverReadClauseNum(p *Msat_Solver_t) int32 { return p.nClauses;   }
func Msat_SolverReadVarAllocNum(p *Msat_Solver_t) int32 { return p.nVarsAlloc; }
func Msat_SolverReadDecisionLevel(p *Msat_Solver_t) int32 { return Msat_IntVecReadSize(p.vTrailLim); }
func Msat_SolverReadDecisionLevelArray(p *Msat_Solver_t) *int32 { return p.pLevel;     }
func Msat_SolverReadReasonArray(p *Msat_Solver_t) **Msat_Clause_t { return p.pReasons;   }
func Msat_SolverReadVarValue(p *Msat_Solver_t, Var Msat_Var_t) Msat_Type_t { return (Msat_Type_t)p.pAssigns[Var]; }
func Msat_SolverReadLearned(p *Msat_Solver_t) *Msat_ClauseVec_t { return p.vLearned;   }
func Msat_SolverReadWatchedArray(p *Msat_Solver_t) **Msat_ClauseVec_t { return p.pvWatched;  }
func Msat_SolverReadAssignsArray(p *Msat_Solver_t) *int32 { return p.pAssigns;   }
func Msat_SolverReadModelArray(p *Msat_Solver_t) *int32 { return p.pModel;     }
func Msat_SolverReadBackTracks(p *Msat_Solver_t) int32 { return int32(p.Stats.nConflicts); }
func Msat_SolverReadInspects(p *Msat_Solver_t) int32 { return int32(p.Stats.nInspects);  }
func Msat_SolverReadMem(p *Msat_Solver_t) *Msat_MmStep_t { return p.pMem;       }
func Msat_SolverReadSeenArray(p *Msat_Solver_t) *int32 { return p.pSeen;      }
func Msat_SolverIncrementSeenId(p *Msat_Solver_t) int32 { return ++p.nSeenId;  }
func Msat_SolverSetVerbosity(p *Msat_Solver_t, fVerbose int32) { p.fVerbose = fVerbose; }
func Msat_SolverClausesIncrement(p *Msat_Solver_t) { p.nClausesAlloc++;   }
func Msat_SolverClausesDecrement(p *Msat_Solver_t) { p.nClausesAlloc--;   }
func Msat_SolverClausesIncrementL(p *Msat_Solver_t) { p.nClausesAllocL++;  }
func Msat_SolverClausesDecrementL(p *Msat_Solver_t) { p.nClausesAllocL--;  }
func Msat_SolverMarkLastClauseTypeA(p *Msat_Solver_t) { Msat_ClauseSetTypeA( Msat_ClauseVecReadEntry( p.vClauses, Msat_ClauseVecReadSize(p.vClauses)-1 ), 1 ); }
func Msat_SolverMarkClausesStart(p *Msat_Solver_t) { p.nClausesStart = Msat_ClauseVecReadSize(p.vClauses); }
func Msat_SolverReadFactors(p *Msat_Solver_t) *float32 { return p.pFactors;   }

/**Function*************************************************************

  Synopsis    [Reads the clause with the given number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverReadClause(p *Msat_Solver_t, Num int32) *Msat_Clause_t {
    var nClausesP int32;
    assert(Num < p.nClauses, "Num < p.nClauses");
    nClausesP = Msat_ClauseVecReadSize( p.vClauses );
    if Num < nClausesP {
        return Msat_ClauseVecReadEntry( p.vClauses, Num ); }
    return Msat_ClauseVecReadEntry( p.vLearned, Num - nClausesP );
}

/**Function*************************************************************

  Synopsis    [Reads the clause with the given number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverReadAdjacents(p *Msat_Solver_t) *Msat_ClauseVec_t {
    return p.vAdjacents;
}

/**Function*************************************************************

  Synopsis    [Reads the clause with the given number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverReadConeVars(p *Msat_Solver_t) *Msat_IntVec_t {
    return p.vConeVars;
}

/**Function*************************************************************

  Synopsis    [Reads the clause with the given number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverReadVarsUsed(p *Msat_Solver_t) *Msat_IntVec_t {
    return p.vVarsUsed;
}


/**Function*************************************************************

  Synopsis    [Allocates the solver.]

  Description [After the solver is allocated, the procedure
  Msat_SolverClean() should be called to set the number of variables.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverAlloc(nVarsAlloc int32, dClaInc float64, dClaDecay float64, dVarInc float64, dVarDecay float64, fVerbose int32) *Msat_Solver_t {
    var p *Msat_Solver_t;
    var i int32;

    assert(sizeof(Msat_Lit_t) == sizeof(unsigned), "sizeof(Msat_Lit_t) == sizeof(unsigned)");
    assert(sizeof(float)     == sizeof(unsigned), "sizeof(float)     == sizeof(unsigned)");

    // inlining of macro ABC_ALLOC
    p = ((Msat_Solver_t *) malloc(sizeof(Msat_Solver_t) * (size_t)(1)));
    memset( p, 0, sizeof(Msat_Solver_t) );

    p.nVarsAlloc = nVarsAlloc;
    p.nVars     = 0;

    p.nClauses  = 0;
    p.vClauses  = Msat_ClauseVecAlloc( 512 );
    p.vLearned  = Msat_ClauseVecAlloc( 512 );

    p.dClaInc   = dClaInc;
    p.dClaDecay = dClaDecay;
    p.dVarInc   = dVarInc;
    p.dVarDecay = dVarDecay;

    // inlining of macro ABC_ALLOC
    p.pdActivity = ((double *) malloc(sizeof(double) * (size_t)(p.nVarsAlloc)));
    // inlining of macro ABC_ALLOC
    p.pFactors   = ((float *) malloc(sizeof(float) * (size_t)(p.nVarsAlloc)));
    for  i = 0; i < p.nVarsAlloc; i++  {
        p.pdActivity[i] = 0.0;
        p.pFactors[i]   = 1.0;
    }

    // inlining of macro ABC_ALLOC
    p.pAssigns  = ((int *) malloc(sizeof(int) * (size_t)(p.nVarsAlloc))); 
    // inlining of macro ABC_ALLOC
    p.pModel    = ((int *) malloc(sizeof(int) * (size_t)(p.nVarsAlloc))); 
    for  i = 0; i < p.nVarsAlloc; i++  {
        p.pAssigns[i] = MSAT_VAR_UNASSIGNED;
    }
//    p->pOrder    = Msat_OrderAlloc( p->pAssigns, p->pdActivity, p->nVarsAlloc );
    p.pOrder    = Msat_OrderAlloc( p );

    // inlining of macro ABC_ALLOC
    p.pvWatched = ((Msat_ClauseVec_t * *) malloc(sizeof(Msat_ClauseVec_t *) * (size_t)(2 * p.nVarsAlloc)));
    for  i = 0; i < 2 * p.nVarsAlloc; i++  {
        p.pvWatched[i] = Msat_ClauseVecAlloc( 16 );
    }
    p.pQueue    = Msat_QueueAlloc( p.nVarsAlloc );

    p.vTrail    = Msat_IntVecAlloc( p.nVarsAlloc );
    p.vTrailLim = Msat_IntVecAlloc( p.nVarsAlloc );
    // inlining of macro ABC_ALLOC
    p.pReasons  = ((Msat_Clause_t * *) malloc(sizeof(Msat_Clause_t *) * (size_t)(p.nVarsAlloc)));
    memset( p.pReasons, 0, sizeof(Msat_Clause_t *) * p.nVarsAlloc );
    // inlining of macro ABC_ALLOC
    p.pLevel = ((int *) malloc(sizeof(int) * (size_t)(p.nVarsAlloc)));
    for  i = 0; i < p.nVarsAlloc; i++  {
        p.pLevel[i] = -1;
    }
    p.dRandSeed = 91648253;
    p.fVerbose  = fVerbose;
    p.dProgress = 0.0;
//    p->pModel = Msat_IntVecAlloc( p->nVarsAlloc );
    p.pMem = Msat_MmStepStart( 10 );

    p.vConeVars   = Msat_IntVecAlloc( p.nVarsAlloc ); 
    p.vAdjacents  = Msat_ClauseVecAlloc( p.nVarsAlloc );
    for  i = 0; i < p.nVarsAlloc; i++  {
        Msat_ClauseVecPush( p.vAdjacents, (Msat_Clause_t *)Msat_IntVecAlloc(5) );
    }
    p.vVarsUsed   = Msat_IntVecAlloc( p.nVarsAlloc ); 
    Msat_IntVecFill( p.vVarsUsed, p.nVarsAlloc, 1 );


    // inlining of macro ABC_ALLOC
    p.pSeen     = ((int *) malloc(sizeof(int) * (size_t)(p.nVarsAlloc)));
    memset( p.pSeen, 0, sizeof(int) * p.nVarsAlloc );
    p.nSeenId   = 1;
    p.vReason   = Msat_IntVecAlloc( p.nVarsAlloc );
    p.vTemp     = Msat_IntVecAlloc( p.nVarsAlloc );
    return p;
}

/**Function*************************************************************

  Synopsis    [Resizes the solver.]

  Description [Assumes that the solver contains some clauses, and that 
  it is currently between the calls. Resizes the solver to accomodate 
  more variables.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverResize(p *Msat_Solver_t, nVarsAlloc int32) {
    var nVarsAllocOld, i int32;

    nVarsAllocOld = p.nVarsAlloc;
    p.nVarsAlloc = nVarsAlloc;

    // inlining of macro ABC_REALLOC
    
    if p.pdActivity {
        p.pdActivity = ((double *) realloc((char *)(p.pdActivity), sizeof(double) * (size_t)(p.nVarsAlloc)));
    } else {
        p.pdActivity = ((double *) malloc(sizeof(double) * (size_t)(p.nVarsAlloc)));
    }
    // inlining of macro ABC_REALLOC
    
    if p.pFactors {
        p.pFactors = ((float *) realloc((char *)(p.pFactors), sizeof(float) * (size_t)(p.nVarsAlloc)));
    } else {
        p.pFactors = ((float *) malloc(sizeof(float) * (size_t)(p.nVarsAlloc)));
    }
    for  i = nVarsAllocOld; i < p.nVarsAlloc; i++  {
        p.pdActivity[i] = 0.0;
        p.pFactors[i]   = 1.0;
    }

    // inlining of macro ABC_REALLOC
    
    if p.pAssigns {
        p.pAssigns = ((int *) realloc((char *)(p.pAssigns), sizeof(int) * (size_t)(p.nVarsAlloc)));
    } else {
        p.pAssigns = ((int *) malloc(sizeof(int) * (size_t)(p.nVarsAlloc)));
    }
    // inlining of macro ABC_REALLOC
    
    if p.pModel {
        p.pModel = ((int *) realloc((char *)(p.pModel), sizeof(int) * (size_t)(p.nVarsAlloc)));
    } else {
        p.pModel = ((int *) malloc(sizeof(int) * (size_t)(p.nVarsAlloc)));
    }
    for  i = nVarsAllocOld; i < p.nVarsAlloc; i++  {
        p.pAssigns[i] = MSAT_VAR_UNASSIGNED;
    }

//    Msat_OrderRealloc( p->pOrder, p->pAssigns, p->pdActivity, p->nVarsAlloc );
    Msat_OrderSetBounds( p.pOrder, p.nVarsAlloc );

    // inlining of macro ABC_REALLOC
    
    if p.pvWatched {
        p.pvWatched = ((Msat_ClauseVec_t * *) realloc((char *)(p.pvWatched), sizeof(Msat_ClauseVec_t *) * (size_t)(2 * p.nVarsAlloc)));
    } else {
        p.pvWatched = ((Msat_ClauseVec_t * *) malloc(sizeof(Msat_ClauseVec_t *) * (size_t)(2 * p.nVarsAlloc)));
    }
    for  i = 2 * nVarsAllocOld; i < 2 * p.nVarsAlloc; i++  {
        p.pvWatched[i] = Msat_ClauseVecAlloc( 16 );
    }

    Msat_QueueFree( p.pQueue );
    p.pQueue    = Msat_QueueAlloc( p.nVarsAlloc );

    // inlining of macro ABC_REALLOC
    
    if p.pReasons {
        p.pReasons = ((Msat_Clause_t * *) realloc((char *)(p.pReasons), sizeof(Msat_Clause_t *) * (size_t)(p.nVarsAlloc)));
    } else {
        p.pReasons = ((Msat_Clause_t * *) malloc(sizeof(Msat_Clause_t *) * (size_t)(p.nVarsAlloc)));
    }
    // inlining of macro ABC_REALLOC
    
    if p.pLevel {
        p.pLevel = ((int *) realloc((char *)(p.pLevel), sizeof(int) * (size_t)(p.nVarsAlloc)));
    } else {
        p.pLevel = ((int *) malloc(sizeof(int) * (size_t)(p.nVarsAlloc)));
    }
    for  i = nVarsAllocOld; i < p.nVarsAlloc; i++  {
        p.pReasons[i] = nil;
        p.pLevel[i] = -1;
    }

    // inlining of macro ABC_REALLOC
    
    if p.pSeen {
        p.pSeen = ((int *) realloc((char *)(p.pSeen), sizeof(int) * (size_t)(p.nVarsAlloc)));
    } else {
        p.pSeen = ((int *) malloc(sizeof(int) * (size_t)(p.nVarsAlloc)));
    }
    for  i = nVarsAllocOld; i < p.nVarsAlloc; i++  {
        p.pSeen[i] = 0;
    }

    Msat_IntVecGrow( p.vTrail, p.nVarsAlloc );
    Msat_IntVecGrow( p.vTrailLim, p.nVarsAlloc );

    // make sure the array of adjucents has room to store the variable numbers
    for  i = Msat_ClauseVecReadSize(p.vAdjacents); i < p.nVarsAlloc; i++  {
        Msat_ClauseVecPush( p.vAdjacents, (Msat_Clause_t *)Msat_IntVecAlloc(5) );
    }
    Msat_IntVecFill( p.vVarsUsed, p.nVarsAlloc, 1 );
}

/**Function*************************************************************

  Synopsis    [Prepares the solver.]

  Description [Cleans the solver assuming that the problem will involve 
  the given number of variables (nVars). This procedure is useful 
  for many small (incremental) SAT problems, to prevent the solver
  from being reallocated each time.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverClean(p *Msat_Solver_t, nVars int32) {
    var i int32;
    // free the clauses
    var nClauses int32;
    var pClauses **Msat_Clause_t;

    assert(p.nVarsAlloc >= nVars, "p.nVarsAlloc >= nVars");
    p.nVars    = nVars;
    p.nClauses = 0;

    nClauses = Msat_ClauseVecReadSize( p.vClauses );
    pClauses = Msat_ClauseVecReadArray( p.vClauses );
    for  i = 0; i < nClauses; i++  {
        Msat_ClauseFree( p, pClauses[i], 0 );
    }
//    Msat_ClauseVecFree( p->vClauses );
    Msat_ClauseVecClear( p.vClauses );

    nClauses = Msat_ClauseVecReadSize( p.vLearned );
    pClauses = Msat_ClauseVecReadArray( p.vLearned );
    for  i = 0; i < nClauses; i++  {
        Msat_ClauseFree( p, pClauses[i], 0 );
    }
//    Msat_ClauseVecFree( p->vLearned );
    Msat_ClauseVecClear( p.vLearned );

//    ABC_FREE( p->pdActivity );
    for  i = 0; i < p.nVars; i++  {
        p.pdActivity[i] = 0;
    }

//    Msat_OrderFree( p->pOrder );
//    Msat_OrderClean( p->pOrder, p->nVars, NULL );
    Msat_OrderSetBounds( p.pOrder, p.nVars );

    for  i = 0; i < 2 * p.nVars; i++  {
//        Msat_ClauseVecFree( p->pvWatched[i] );
        Msat_ClauseVecClear( p.pvWatched[i] );
    }
//    ABC_FREE( p->pvWatched );
//    Msat_QueueFree( p->pQueue );
    Msat_QueueClear( p.pQueue );

//    ABC_FREE( p->pAssigns );
    for  i = 0; i < p.nVars; i++  {
        p.pAssigns[i] = MSAT_VAR_UNASSIGNED;
    }
//    Msat_IntVecFree( p->vTrail );
    Msat_IntVecClear( p.vTrail );
//    Msat_IntVecFree( p->vTrailLim );
    Msat_IntVecClear( p.vTrailLim );
//    ABC_FREE( p->pReasons );
    memset( p.pReasons, 0, sizeof(Msat_Clause_t *) * p.nVars );
//    ABC_FREE( p->pLevel );
    for  i = 0; i < p.nVars; i++  {
        p.pLevel[i] = -1;
    }
//    Msat_IntVecFree( p->pModel );
//    Msat_MmStepStop( p->pMem, 0 );
    p.dRandSeed = 91648253;
    p.dProgress = 0.0;

//    ABC_FREE( p->pSeen );
    memset( p.pSeen, 0, sizeof(int) * p.nVars );
    p.nSeenId = 1;
//    Msat_IntVecFree( p->vReason );
    Msat_IntVecClear( p.vReason );
//    Msat_IntVecFree( p->vTemp );
    Msat_IntVecClear( p.vTemp );
//    printf(" The number of clauses remaining = %d (%d).\n", p->nClausesAlloc, p->nClausesAllocL );
//    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Frees the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverFree(p *Msat_Solver_t) {
    var i int32;

    // free the clauses
    var nClauses int32;
    var pClauses **Msat_Clause_t;
//printf( "clauses = %d. learned = %d.\n", Msat_ClauseVecReadSize( p->vClauses ), 
//                                         Msat_ClauseVecReadSize( p->vLearned ) );

    nClauses = Msat_ClauseVecReadSize( p.vClauses );
    pClauses = Msat_ClauseVecReadArray( p.vClauses );
    for  i = 0; i < nClauses; i++  {
        Msat_ClauseFree( p, pClauses[i], 0 );
    }
    Msat_ClauseVecFree( p.vClauses );

    nClauses = Msat_ClauseVecReadSize( p.vLearned );
    pClauses = Msat_ClauseVecReadArray( p.vLearned );
    for  i = 0; i < nClauses; i++  {
        Msat_ClauseFree( p, pClauses[i], 0 );
    }
    Msat_ClauseVecFree( p.vLearned );

    // inlining of macro ABC_FREE
    ((p.pdActivity) ? (free((char *) (p.pdActivity)), (p.pdActivity) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pFactors) ? (free((char *) (p.pFactors)), (p.pFactors) = 0) : 0);
    Msat_OrderFree( p.pOrder );

    for  i = 0; i < 2 * p.nVarsAlloc; i++  {
        Msat_ClauseVecFree( p.pvWatched[i] );
    }
    // inlining of macro ABC_FREE
    ((p.pvWatched) ? (free((char *) (p.pvWatched)), (p.pvWatched) = 0) : 0);
    Msat_QueueFree( p.pQueue );

    // inlining of macro ABC_FREE
    ((p.pAssigns) ? (free((char *) (p.pAssigns)), (p.pAssigns) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pModel) ? (free((char *) (p.pModel)), (p.pModel) = 0) : 0);
    Msat_IntVecFree( p.vTrail );
    Msat_IntVecFree( p.vTrailLim );
    // inlining of macro ABC_FREE
    ((p.pReasons) ? (free((char *) (p.pReasons)), (p.pReasons) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pLevel) ? (free((char *) (p.pLevel)), (p.pLevel) = 0) : 0);

    Msat_MmStepStop( p.pMem, 0 );

    nClauses = Msat_ClauseVecReadSize( p.vAdjacents );
    pClauses = Msat_ClauseVecReadArray( p.vAdjacents );
    for  i = 0; i < nClauses; i++  {
        Msat_IntVecFree( (Msat_IntVec_t *)pClauses[i] );
    }
    Msat_ClauseVecFree( p.vAdjacents );
    Msat_IntVecFree( p.vConeVars );
    Msat_IntVecFree( p.vVarsUsed );

    // inlining of macro ABC_FREE
    ((p.pSeen) ? (free((char *) (p.pSeen)), (p.pSeen) = 0) : 0);
    Msat_IntVecFree( p.vReason );
    Msat_IntVecFree( p.vTemp );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Prepares the solver to run on a subset of variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverPrepare(p *Msat_Solver_t, vVars *Msat_IntVec_t) {

    var i int32;
    // undo the previous data
    for  i = 0; i < p.nVarsAlloc; i++  {
        p.pAssigns[i]   = MSAT_VAR_UNASSIGNED;
        p.pReasons[i]   = nil;
        p.pLevel[i]     = -1;
        p.pdActivity[i] = 0.0;
    }

    // set the new variable order
    Msat_OrderClean( p.pOrder, vVars );

    Msat_QueueClear( p.pQueue );
    Msat_IntVecClear( p.vTrail );
    Msat_IntVecClear( p.vTrailLim );
    p.dProgress = 0.0;
}

/**Function*************************************************************

  Synopsis    [Sets up the truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverSetupTruthTables(uTruths [][2]uint32) {
    var m, v int32;
    // set up the truth tables
    for  m = 0; m < 32; m++  {
        for  v = 0; v < 5; v++  {
            if m & (1 << v) {
                uTruths[v][0] |= (1 << m); }
        }
    }
    // make adjustments for the case of 6 variables
    for  v = 0; v < 5; v++  {
        uTruths[v][1] = uTruths[v][0];
    }
    uTruths[5][0] = 0;
    uTruths[5][1] = ~(uint32(0));
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


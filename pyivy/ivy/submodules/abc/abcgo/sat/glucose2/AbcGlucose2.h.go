package glucose2

/**CFile****************************************************************

  FileName    [AbcGlucose.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver Glucose 3.0 by Gilles Audemard and Laurent Simon.]

  Synopsis    [Interface to Glucose.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 6, 2017.]

  Revision    [$Id: AbcGlucose.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

const (
	GLUCOSE_UNSAT int32 = -1
	GLUCOSE_SAT   int32 = 1
	GLUCOSE_UNDEC int32 = 0
)

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// Glucose2Params configures DIMACS or AIG solving.
type Glucose2Params struct {
	Preprocess    bool
	Verbose       bool
	Custom        int32
	ConflictLimit int32 // 0 means unlimited
}

// Glucose2_Pars retains the source-facing name for existing callers.
type Glucose2_Pars = Glucose2Params

func NewGlucose2Params(preprocess, verbose bool, conflictLimit int32) Glucose2Params {
	return Glucose2Params{Preprocess: preprocess, Verbose: verbose, ConflictLimit: conflictLimit}
}

func Glucose_CreatePars(p, v, c, nConfls int32) Glucose2Params {
	return NewGlucose2Params(p != 0, v != 0, nConfls)
}

// ABC's former opaque C handle is represented by the concrete Go solver.
type bmcg2_sat_solver = SimpSolver

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

// extern bmcg2_sat_solver * bmcg2_sat_solver_start();

// extern void              bmcg2_sat_solver_stop( bmcg2_sat_solver* s );

// extern void              bmcg2_sat_solver_reset( bmcg2_sat_solver* s );

// extern int               bmcg2_sat_solver_addclause( bmcg2_sat_solver* s, int * plits, int nlits );

// extern void              bmcg2_sat_solver_setcallback( bmcg2_sat_solver* s, void * pman, int(*pfunc)(void*, int, int*) );

// extern int               bmcg2_sat_solver_solve( bmcg2_sat_solver* s, int * plits, int nlits );

// extern int               bmcg2_sat_solver_final( bmcg2_sat_solver* s, int ** ppArray );

// extern int               bmcg2_sat_solver_addvar( bmcg2_sat_solver* s );

// extern void              bmcg2_sat_solver_set_nvars( bmcg2_sat_solver* s, int nvars );

// extern int               bmcg2_sat_solver_eliminate( bmcg2_sat_solver* s, int turn_off_elim );

// extern int               bmcg2_sat_solver_var_is_elim( bmcg2_sat_solver* s, int v );

// extern void              bmcg2_sat_solver_var_set_frozen( bmcg2_sat_solver* s, int v, int freeze );

// extern int               bmcg2_sat_solver_elim_varnum(bmcg2_sat_solver* s);

// extern int *             bmcg2_sat_solver_read_cex( bmcg2_sat_solver* s );

// extern int               bmcg2_sat_solver_read_cex_varvalue( bmcg2_sat_solver* s, int );

// extern void              bmcg2_sat_solver_set_stop( bmcg2_sat_solver* s, int * pstop );

// extern abctime           bmcg2_sat_solver_set_runtime_limit( bmcg2_sat_solver* s, abctime Limit );

// extern void              bmcg2_sat_solver_set_conflict_budget( bmcg2_sat_solver* s, int Limit );

// extern int               bmcg2_sat_solver_varnum( bmcg2_sat_solver* s );

// extern int               bmcg2_sat_solver_clausenum( bmcg2_sat_solver* s );

// extern int               bmcg2_sat_solver_learntnum( bmcg2_sat_solver* s );

// extern int               bmcg2_sat_solver_conflictnum( bmcg2_sat_solver* s );

// extern int               bmcg2_sat_solver_minimize_assumptions( bmcg2_sat_solver * s, int * plits, int nlits, int pivot );

// extern int               bmcg2_sat_solver_add_and( bmcg2_sat_solver * s, int iVar, int iVar0, int iVar1, int fCompl0, int fCompl1, int fCompl );

// extern int               bmcg2_sat_solver_add_xor( bmcg2_sat_solver * s, int iVarA, int iVarB, int iVarC, int fCompl );

// extern int               bmcg2_sat_solver_quantify( bmcg2_sat_solver * s[], Gia_Man_t * p, int iLit, int fHash, int(*pFuncCiToKeep)(void *, int), void * pData, Vec_Int_t * vDLits );

// extern int               bmcg2_sat_solver_equiv_overlap_check( bmcg2_sat_solver * s, Gia_Man_t * p, int iLit0, int iLit1, int fEquiv );

// extern Vec_Str_t *       bmcg2_sat_solver_sop( Gia_Man_t * p, int CubeLimit );

// extern int               bmcg2_sat_solver_jftr( bmcg2_sat_solver * s );

// extern void              bmcg2_sat_solver_set_jftr( bmcg2_sat_solver * s, int jftr );

// extern void              bmcg2_sat_solver_set_var_fanin_lit( bmcg2_sat_solver * s, int var, int lit0, int lit1 );

// extern void              bmcg2_sat_solver_start_new_round( bmcg2_sat_solver * s );

// extern void              bmcg2_sat_solver_mark_cone( bmcg2_sat_solver * s, int var );

// extern void              Glucose2_SolveCnf( char * pFilename, Glucose2_Pars * pPars );

// extern int               Glucose2_SolveAig( Gia_Man_t * p, Glucose2_Pars * pPars );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

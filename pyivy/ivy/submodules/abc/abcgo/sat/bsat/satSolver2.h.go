package bsat

/**************************************************************************************************
MiniSat -- Copyright (c) 2005, Niklas Sorensson
http://www.cs.chalmers.se/Cs/Research/FormalMethods/MiniSat/

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/
// Modified to compile with MS Visual Studio 6.0 by Alan Mishchenko



 

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "satVec.h"
//#include "satClause.h"
//#include "misc/vec/vecSet.h"
//#include "satProof2.h"

//ABC_NAMESPACE_HEADER_START

//#define USE_FLOAT_ACTIVITY2

//=================================================================================================
// Public interface:

struct sat_solver2_t;
type sat_solver2 = sat_solver2_t;
type Int2_Man_t = Int2_Man_t_;

// extern sat_solver2* sat_solver2_new(void);

// extern void         sat_solver2_delete(sat_solver2* s);


// extern int          sat_solver2_addclause(sat_solver2* s, lit* begin, lit* end, int Id);

// extern int          sat_solver2_simplify(sat_solver2* s);

// extern int          sat_solver2_solve(sat_solver2* s, lit* begin, lit* end, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, ABC_INT64_T nConfLimitGlobal, ABC_INT64_T nInsLimitGlobal);

// extern void         sat_solver2_rollback(sat_solver2* s);

// extern void         sat_solver2_reducedb(sat_solver2* s);

// extern double       sat_solver2_memory( sat_solver2* s, int fAll );

// extern double       sat_solver2_memory_proof( sat_solver2* s );


// extern void         sat_solver2_setnvars(sat_solver2* s,int n);


// extern void         Sat_Solver2WriteDimacs( sat_solver2 * p, char * pFileName, lit* assumptionsBegin, lit* assumptionsEnd, int incrementVars );

// extern void         Sat_Solver2PrintStats( FILE * pFile, sat_solver2 * p );

// extern int *        Sat_Solver2GetModel( sat_solver2 * p, int * pVars, int nVars );

// extern void         Sat_Solver2DoubleClauses( sat_solver2 * p, int iVar );


// global variables
// extern int          var_is_assigned(sat_solver2* s, int v);

// extern int          var_is_partA   (sat_solver2* s, int v);

// extern void         var_set_partA  (sat_solver2* s, int v, int partA);


// proof-based APIs
// extern void *       Sat_ProofCore( sat_solver2 * s );

// extern void *       Sat_ProofInterpolant( sat_solver2 * s, void * pGloVars );

// extern word *       Sat_ProofInterpolantTruth( sat_solver2 * s, void * pGloVars );

// extern void         Sat_ProofCheck( sat_solver2 * s );


// interpolation APIs
// extern Int2_Man_t * Int2_ManStart( sat_solver2 * pSat, int * pGloVars, int nGloVars );

// extern void         Int2_ManStop( Int2_Man_t * p );

// extern int          Int2_ManChainStart( Int2_Man_t * p, clause * c );

// extern int          Int2_ManChainResolve( Int2_Man_t * p, clause * c, int iLit, int varA );

// extern void *       Int2_ManReadInterpolant( sat_solver2 * s );



//=================================================================================================
// Solver representation:

struct varinfo_t;
type varinfo2 = varinfo2_t;

type sat_solver2_t struct {
    size             int32;           // nof variables
    cap             int32;            // size of varmaps
    qhead             int32;          // Head index of queue.
    qtail             int32;          // Tail index of queue.

    root_level             int32;     // Level of first proper decision.
    random_seed          float64;
    progress_estimate          float64;
    verbosity             int32;      // Verbosity level. 0=silent, 1=some progress report, 2=everything    // activities

#ifdef USE_FLOAT_ACTIVITY2
    double          var_inc;        // Amount to bump next variable with.
    double          var_decay;      // INVERSE decay factor for variable activity: stores 1/decay. 
    float           cla_inc;        // Amount to bump next clause with.
    float           cla_decay;      // INVERSE decay factor for clause activity: stores 1/decay.
    double*         activity;       // A heuristic measurement of the activity of a variable.
#else
    int             var_inc;        // Amount to bump next variable with.
    int             var_inc2;       // Amount to bump next variable with.
    int             cla_inc;        // Amount to bump next clause with.
    unsigned*       activity;       // A heuristic measurement of the activity of a variable
    unsigned*       activity2;      // backup variable activity
#endif

    nUnits             int32;         // the total number of unit clauses
    nof_learnts             int32;    // the number of clauses to trigger reduceDB
    nLearntMax             int32;     // enables using reduce DB method
    nLearntStart             int32;   // starting learned clause limit
    nLearntDelta             int32;   // delta of learned clause limit
    nLearntRatio             int32;   // ratio percentage of learned clauses
    nDBreduces             int32;     // number of DB reductions
    fNotUseRandom             int32;  // do not allow random decisions with a fixed probability
    fSkipSimplify             int32;  // set to one to skip simplification of the clause database
    fProofLogging             int32;  // enable proof-logging
    fVerbose             int32;

    // clauses
    Mem       Sat_Mem_t;
    wlists           *veci;         // watcher lists (for each literal)
    act_clas            veci;       // clause activities
    claProofs            veci;      // clause proofs

    // rollback
    iVarPivot             int32;      // the pivot for variables
    iTrailPivot             int32;    // the pivot for trail
    hProofPivot             int32;    // the pivot for proof records

    // internal state
    vi       *varinfo2;             // variable information
    levels            *int32;         // 
    assigns           string;        // 
    trail            *lit;          // sequence of assignment and implications
    orderpos            *int32;       // Index in variable order.
    reasons            *cla;        // reason clauses
    units            *cla;          // unit clauses
    model            *int32;          // If problem is solved, this vector contains the model (contains: lbool).

    tagged            veci;         // (contains: var)
    stack            veci;          // (contains: var)
    order            veci;          // Variable order. (heap) (contains: var)
    trail_lim            veci;      // Separator indices for different decision levels in 'trail'. (contains: int)
    temp_clause            veci;    // temporary storage for a CNF clause
    conf_final            veci;     // If problem is unsatisfiable (possibly under assumptions),
                                    // this vector represent the final conflict clause expressed in the assumptions.
    mark_levels            veci;    // temporary storage for labeled levels
    min_lit_order            veci;  // ordering of removable literals
    min_step_order            veci; // ordering of resolution steps
    learnt_live            veci;    // remaining clauses after reduce DB

    // proof logging
    pPrf1      *Vec_Set_t;          // sequence of proof records
    temp_proof            veci;     // temporary place to store proofs
    hLearntLast             int32;    // in proof-logging mode, the ID of the final conflict clause (conf_final)
    hProofLast             int32;     // in proof-logging mode, the ID of the final conflict clause (conf_final)
    pPrf2      *Prf_Man_t;          // another proof manager
    dPrfMemory          float64;     // memory used by the proof-logger
    pInt2     *Int2_Man_t;          // interpolation manager
    tempInter             int32;      // temporary storage for the interpolant

    // statistics
    stats         stats_t;
    nConfLimit     ABC_INT64_T;     // external limit on the number of conflicts
    nInsLimit     ABC_INT64_T;      // external limit on the number of implications
    nRuntimeLimit         abctime;  // external limit on runtime
}

func clause2_read(s *sat_solver2, h cla) *clause { return Sat_MemClauseHand( &s.Mem, h ); }
func clause2_proofid(s *sat_solver2, c *clause, varA int32) int32 { if c.lrn {
    return (veci_begin(&s.claProofs)[clause_id(c)]<<2) | (varA<<1);
} else {
    return (clause_id(c)<<2) | (varA<<1) | 1;
} }

// these two only work after creating a clause before the solver is called
func clause2_is_partA(s *sat_solver2, h int32) int32 { return clause2_read(s, h).partA;       }
func clause2_set_partA(s *sat_solver2, h int32, partA int32) { clause2_read(s, h).partA = partA;      }
func clause2_id(s *sat_solver2, h int32) int32 { return clause_id(clause2_read(s, h));   }
func clause2_set_id(s *sat_solver2, h int32, id int32) { clause_set_id(clause2_read(s, h), id);  }

//=================================================================================================
// Public APIs:

func sat_solver2_nvars(s *sat_solver2) int32 {
    return s.size;
}

func sat_solver2_nclauses(s *sat_solver2) int32 {
    return int32(s.stats.clauses);
}

func sat_solver2_nlearnts(s *sat_solver2) int32 {
    return int32(s.stats.learnts);
}

func sat_solver2_nconflicts(s *sat_solver2) int32 {
    return int32(s.stats.conflicts);
}

func sat_solver2_var_value(s *sat_solver2, v int32) int32 {
    assert(v >= 0 && v < s.size, "v >= 0 && v < s.size");
    return int32((s.model[v] == l_True));
}
func sat_solver2_var_literal(s *sat_solver2, v int32) int32 {
    assert(v >= 0 && v < s.size, "v >= 0 && v < s.size");
    return toLitCond( v, s.model[v] != l_True );
}
func sat_solver2_act_var_clear(s *sat_solver2) {
    var i int32;
    for i = 0; i < s.size; i++ {
        s.activity[i] = 0;
    }//.0;
    s.var_inc = 1.0;
}

func sat_solver2_final(s *sat_solver2, ppArray **int32) int32 {
    *ppArray = s.conf_final.ptr;
    return s.conf_final.size;
}

func sat_solver2_set_runtime_limit(s *sat_solver2, Limit abctime) abctime {
    var temp abctime  = s.nRuntimeLimit;
    s.nRuntimeLimit = Limit;
    return temp;
}

func sat_solver2_set_learntmax(s *sat_solver2, nLearntMax int32) int32 {
    var temp int32  = s.nLearntMax;
    s.nLearntMax = nLearntMax;
    return temp;
}

func sat_solver2_bookmark(s *sat_solver2) {
    assert(s.qhead == s.qtail, "s.qhead == s.qtail");
    s.iVarPivot    = s.size;
    s.iTrailPivot  = s.qhead;
    if s.pPrf1 {
        s.hProofPivot  = Vec_SetHandCurrent(s.pPrf1); }
    Sat_MemBookMark( &s.Mem );
    if s.activity2 {
        s.var_inc2 = s.var_inc;
        memcpy( s.activity2, s.activity, sizeof(unsigned) * s.iVarPivot );
    }
}

func sat_solver2_add_const(pSat *sat_solver2, iVar int32, fCompl int32, fMark int32, Id int32) int32 {
    var Lits [1]lit;
    var Cid int32;
    assert(iVar >= 0, "iVar >= 0");

    Lits[0] = toLitCond( iVar, fCompl );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 1, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }
    return 1;
}
func sat_solver2_add_buffer(pSat *sat_solver2, iVarA int32, iVarB int32, fCompl int32, fMark int32, Id int32) int32 {
    var Lits [2]lit;
    var Cid int32;
    assert(iVarA >= 0 && iVarB >= 0, "iVarA >= 0 && iVarB >= 0");

    Lits[0] = toLitCond( iVarA, 0 );
    Lits[1] = toLitCond( iVarB, !fCompl );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 2, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }

    Lits[0] = toLitCond( iVarA, 1 );
    Lits[1] = toLitCond( iVarB, fCompl );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 2, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }
    return 2;
}
func sat_solver2_add_and(pSat *sat_solver2, iVar int32, iVar0 int32, iVar1 int32, fCompl0 int32, fCompl1 int32, fMark int32, Id int32) int32 {
    var Lits [3]lit;
    var Cid int32;

    Lits[0] = toLitCond( iVar, 1 );
    Lits[1] = toLitCond( iVar0, fCompl0 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 2, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }

    Lits[0] = toLitCond( iVar, 1 );
    Lits[1] = toLitCond( iVar1, fCompl1 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 2, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }

    Lits[0] = toLitCond( iVar, 0 );
    Lits[1] = toLitCond( iVar0, !fCompl0 );
    Lits[2] = toLitCond( iVar1, !fCompl1 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 3, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }
    return 3;
}
func sat_solver2_add_xor(pSat *sat_solver2, iVarA int32, iVarB int32, iVarC int32, fCompl int32, fMark int32, Id int32) int32 {
    var Lits [3]lit;
    var Cid int32;
    assert(iVarA >= 0 && iVarB >= 0 && iVarC >= 0, "iVarA >= 0 && iVarB >= 0 && iVarC >= 0");

    Lits[0] = toLitCond( iVarA, !fCompl );
    Lits[1] = toLitCond( iVarB, 1 );
    Lits[2] = toLitCond( iVarC, 1 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 3, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }

    Lits[0] = toLitCond( iVarA, !fCompl );
    Lits[1] = toLitCond( iVarB, 0 );
    Lits[2] = toLitCond( iVarC, 0 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 3, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }

    Lits[0] = toLitCond( iVarA, fCompl );
    Lits[1] = toLitCond( iVarB, 1 );
    Lits[2] = toLitCond( iVarC, 0 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 3, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }

    Lits[0] = toLitCond( iVarA, fCompl );
    Lits[1] = toLitCond( iVarB, 0 );
    Lits[2] = toLitCond( iVarC, 1 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 3, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }
    return 4;
}
func sat_solver2_add_constraint(pSat *sat_solver2, iVar int32, iVar2 int32, fCompl int32, fMark int32, Id int32) int32 {
    var Lits [2]lit;
    var Cid int32;
    assert(iVar >= 0, "iVar >= 0");

    Lits[0] = toLitCond( iVar, fCompl );
    Lits[1] = toLitCond( iVar2, 0 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 2, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }

    Lits[0] = toLitCond( iVar, fCompl );
    Lits[1] = toLitCond( iVar2, 1 );
    Cid = sat_solver2_addclause( pSat, Lits, Lits + 2, Id );
    if fMark != 0 {
        clause2_set_partA( pSat, Cid, 1 ); }
    return 2;
}


//ABC_NAMESPACE_HEADER_END



package bsat

import "os"

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
//#include "misc/util/utilDouble.h"

//ABC_NAMESPACE_HEADER_START

//=================================================================================================
// Public interface:

struct sat_solver3_t;
type sat_solver3 = sat_solver3_t;

// extern sat_solver3* sat_solver3_new(void);

// extern sat_solver3* zsat_solver3_new_seed(double seed);

// extern void        sat_solver3_delete(sat_solver3* s);


// extern int         sat_solver3_addclause(sat_solver3* s, lit* begin, lit* end);

// extern int         sat_solver3_clause_new(sat_solver3* s, lit* begin, lit* end, int learnt);

// extern int         sat_solver3_simplify(sat_solver3* s);

// extern int         sat_solver3_solve(sat_solver3* s, lit* begin, lit* end, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, ABC_INT64_T nConfLimitGlobal, ABC_INT64_T nInsLimitGlobal);

// extern int         sat_solver3_solve_internal(sat_solver3* s);

// extern int         sat_solver3_solve_lexsat(sat_solver3* s, int * pLits, int nLits);

// extern int         sat_solver3_minimize_assumptions( sat_solver3* s, int * pLits, int nLits, int nConfLimit );

// extern int         sat_solver3_minimize_assumptions2( sat_solver3* s, int * pLits, int nLits, int nConfLimit );

// extern int         sat_solver3_push(sat_solver3* s, int p);

// extern void        sat_solver3_pop(sat_solver3* s);

// extern void        sat_solver3_set_resource_limits(sat_solver3* s, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, ABC_INT64_T nConfLimitGlobal, ABC_INT64_T nInsLimitGlobal);

// extern void        sat_solver3_restart( sat_solver3* s );

// extern void        zsat_solver3_restart_seed( sat_solver3* s, double seed );

// extern void        sat_solver3_rollback( sat_solver3* s );


// extern int         sat_solver3_nvars(sat_solver3* s);

// extern int         sat_solver3_nclauses(sat_solver3* s);

// extern int         sat_solver3_nconflicts(sat_solver3* s);

// extern double      sat_solver3_memory(sat_solver3* s);

// extern int         sat_solver3_count_assigned(sat_solver3* s);


// extern void        sat_solver3_setnvars(sat_solver3* s,int n);

// extern int         sat_solver3_get_var_value(sat_solver3* s, int v);

// extern void        sat_solver3_set_var_activity(sat_solver3* s, int * pVars, int nVars);


// extern void        sat_solver3WriteDimacs( sat_solver3 * p, char * pFileName, lit* assumptionsBegin, lit* assumptionsEnd, int incrementVars );

// extern void        sat_solver3PrintStats( FILE * pFile, sat_solver3 * p );

// extern int *       sat_solver3GetModel( sat_solver3 * p, int * pVars, int nVars );

// extern void        sat_solver3DoubleClauses( sat_solver3 * p, int iVar );


// trace recording
// extern void        sat_solver3TraceStart( sat_solver3 * pSat, char * pName );

// extern void        sat_solver3TraceStop( sat_solver3 * pSat );

// extern void        sat_solver3TraceWrite( sat_solver3 * pSat, int * pBeg, int * pEnd, int fRoot );


// clause storage
// extern void        sat_solver3_store_alloc( sat_solver3 * s );

// extern void        sat_solver3_store_write( sat_solver3 * s, char * pFileName );

// extern void        sat_solver3_store_free( sat_solver3 * s );

// extern void        sat_solver3_store_mark_roots( sat_solver3 * s );

// extern void        sat_solver3_store_mark_clauses_a( sat_solver3 * s );

// extern void *      sat_solver3_store_release( sat_solver3 * s );
 

//=================================================================================================
// Solver representation:

//struct clause_t;
//typedef struct clause_t clause;

struct varinfo_t;
type varinfo = varinfo_t;

type sat_solver3_t struct {
    size         int32;          // nof variables
    cap         int32;           // size of varmaps
    qhead         int32;         // Head index of queue.
    qtail         int32;         // Tail index of queue.

    // clauses
    Mem   Sat_Mem_t;
    hLearnts         int32;      // the first learnt clause
    hBinary         int32;       // the special binary clause
    binary     *clause;
    wlists       *veci;        // watcher lists

    // rollback
    iVarPivot         int32;     // the pivot for variables
    iTrailPivot         int32;   // the pivot for trail
    hProofPivot         int32;   // the pivot for proof records

    // activities
    VarActType         int32;
    ClaActType         int32;
    var_inc        word;       // Amount to bump next variable with.
    var_inc2        word;      // Amount to bump next variable with.
    var_decay        word;     // INVERSE decay factor for variable activity: stores 1/decay. 
    activity       *word;      // A heuristic measurement of the activity of a variable.
    activity2       *word;     // backup variable activity
    cla_inc    uint32;       // Amount to bump next clause with.
    cla_decay    uint32;     // INVERSE decay factor for clause activity: stores 1/decay.
    act_clas        veci;      // contain clause activities

    pFreqs       string;        // how many times this variable was assigned a value
    nVarUsed         int32;

//    varinfo *   vi;            // variable information
    levels        *int32;        //
    assigns       string;       // Current values of variables.
    polarity       string;      //
    tags       string;          //
    loads       string;         //

    orderpos        *int32;      // Index in variable order.
    reasons        *int32;       //
    trail        *lit;
    tagged        veci;        // (contains: var)
    stack        veci;         // (contains: var)

    order        veci;         // Variable order. (heap) (contains: var)
    trail_lim        veci;     // Separator indices for different decision levels in 'trail'. (contains: int)
//    veci        model;         // If problem is solved, this vector contains the model (contains: lbool).
    model        *int32;         // If problem is solved, this vector contains the model (contains: lbool).
    conf_final        veci;    // If problem is unsatisfiable (possibly under assumptions),
                               // this vector represent the final conflict clause expressed in the assumptions.

    root_level         int32;    // Level of first proper decision.
    simpdb_assigns         int32;// Number of top-level assignments at last 'simplifyDB()'.
    simpdb_props         int32;  // Number of propagations before next 'simplifyDB()'.
    random_seed      float64;
    progress_estimate      float64;
    verbosity         int32;     // Verbosity level. 0=silent, 1=some progress report, 2=everything
    fVerbose         int32;

    stats     stats_t;
    nLearntMax         int32;    // max number of learned clauses
    nLearntStart         int32;  // starting learned clause limit
    nLearntDelta         int32;  // delta of learned clause limit
    nLearntRatio         int32;  // ratio percentage of learned clauses
    nDBreduces         int32;    // number of DB reductions

    nConfLimit ABC_INT64_T;    // external limit on the number of conflicts
    nInsLimit ABC_INT64_T;     // external limit on the number of implications
    nRuntimeLimit     abctime; // external limit on runtime

    act_vars        veci;      // variables whose activity has changed
    factors     *float64;       // the activity factors
    nRestarts         int32;     // the number of local restarts
    nCalls         int32;        // the number of local restarts
    nCalls2         int32;       // the number of local restarts
    unit_lits        veci;     // variables whose activity has changed
    pivot_vars        veci;    // pivot variables

    fSkipSimplify         int32; // set to one to skip simplification of the clause database
    fNotUseRandom         int32; // do not allow random decisions with a fixed probability
    fNoRestarts         int32;   // disables periodic restarts

    pGlobalVars        *int32;   // for experiments with global vars during interpolation
    // clause store
    pStore       any;
    fSolved         int32;

    // trace recording
    pFile       *os.File;
    nClauses         int32;
    nRoots         int32;

    temp_clause        veci;    // temporary storage for a CNF clause

    // CNF loading
    pCnfMan       any;           // external CNF manager
    int(*pCnfFunc)(void * p, int); // external callback
}

func clause_read(s *sat_solver3, h cla) *clause { 
    return Sat_MemClauseHand( &s.Mem, h );      
}

func sat_solver3_var_value(s *sat_solver3, v int32) int32 {
    assert(v >= 0 && v < s.size, "v >= 0 && v < s.size");
    return int32((s.model[v] == l_True));
}
func sat_solver3_var_literal(s *sat_solver3, v int32) int32 {
    assert(v >= 0 && v < s.size, "v >= 0 && v < s.size");
    return toLitCond( v, s.model[v] != l_True );
}
func sat_solver3_act_var_clear(s *sat_solver3) {
    var i int32;
    if s.VarActType == 0 {
        for i = 0; i < s.size; i++ {
            s.activity[i] = (1 << 10);
        }
        s.var_inc = (1 << 5);
    
    } else if s.VarActType == 1 {
        for i = 0; i < s.size; i++ {
            s.activity[i] = 0;
        }
        s.var_inc = 1;
    
    } else if s.VarActType == 2 {
        for i = 0; i < s.size; i++ {
            s.activity[i] = Xdbl_Const1();
        }
        s.var_inc = Xdbl_Const1();
    
    } else { assert(false, "0"); }
}
func sat_solver3_compress(s *sat_solver3) {
    if s.qtail != s.qhead {
        var RetValue int32  = sat_solver3_simplify(s);
        assert(RetValue != 0, "RetValue != 0");
        (void) RetValue;
    }
}
func sat_solver3_delete_p(ps **sat_solver3) {
    if *ps {
        sat_solver3_delete( *ps ); }
    *ps = nil;
}
func sat_solver3_clean_polarity(s *sat_solver3, pVars *int32, nVars int32) {
    var i int32;
    for  i = 0; i < nVars; i++  {
        s.polarity[pVars[i]] = 0;
    }
}
func sat_solver3_set_polarity(s *sat_solver3, pVars *int32, nVars int32) {
    var i int32;
    for  i = 0; i < s.size; i++  {
        s.polarity[i] = 0;
    }
    for  i = 0; i < nVars; i++  {
        s.polarity[pVars[i]] = 1;
    }
}
func sat_solver3_set_literal_polarity(s *sat_solver3, pLits *int32, nLits int32) {
    var i int32;
    for  i = 0; i < nLits; i++  {
        s.polarity[Abc_Lit2Var(pLits[i])] = !Abc_LitIsCompl(pLits[i]);
    }
}

func sat_solver3_final(s *sat_solver3, ppArray **int32) int32 {
    *ppArray = s.conf_final.ptr;
    return s.conf_final.size;
}

func sat_solver3_set_runtime_limit(s *sat_solver3, Limit abctime) abctime {
    var nRuntimeLimit abctime  = s.nRuntimeLimit;
    s.nRuntimeLimit = Limit;
    return nRuntimeLimit;
}

func sat_solver3_set_random(s *sat_solver3, fNotUseRandom int32) int32 {
    var fNotUseRandomOld int32  = s.fNotUseRandom;
    s.fNotUseRandom = fNotUseRandom;
    return fNotUseRandomOld;
}

func sat_solver3_bookmark(s *sat_solver3) {
    assert(s.qhead == s.qtail, "s.qhead == s.qtail");
    s.iVarPivot    = s.size;
    s.iTrailPivot  = s.qhead;
    Sat_MemBookMark( &s.Mem );
    if s.activity2 {
        s.var_inc2 = s.var_inc;
        memcpy( s.activity2, s.activity, sizeof(word) * s.iVarPivot );
    }
}
func sat_solver3_set_pivot_variables(s *sat_solver3, pPivots *int32, nPivots int32) {
    s.pivot_vars.cap = nPivots;
    s.pivot_vars.size = nPivots;
    s.pivot_vars.ptr = pPivots;
}
func sat_solver3_count_usedvars(s *sat_solver3) int32 {
    var i, nVars int32;
    for  i = 0; i < s.size; i++  {
        if s.pFreqs[i] {
            s.pFreqs[i] = 0;
            nVars++;
        }
    }
    return nVars;
}

func sat_solver3_add_const(pSat *sat_solver3, iVar int32, fCompl int32) int32 {
    var Lits [1]lit;
    var Cid int32;
    assert(iVar >= 0, "iVar >= 0");

    Lits[0] = toLitCond( iVar, fCompl );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 1 );
    assert((Cid != 0), "Cid");
    return 1;
}
func sat_solver3_add_buffer(pSat *sat_solver3, iVarA int32, iVarB int32, fCompl int32) int32 {
    var Lits [2]lit;
    var Cid int32;
    assert(iVarA >= 0 && iVarB >= 0, "iVarA >= 0 && iVarB >= 0");

    Lits[0] = toLitCond( iVarA, 0 );
    Lits[1] = toLitCond( iVarB, !fCompl );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 2 );
    if Cid == 0 {
        return 0; }
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarA, 1 );
    Lits[1] = toLitCond( iVarB, fCompl );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 2 );
    if Cid == 0 {
        return 0; }
    assert((Cid != 0), "Cid");
    return 2;
}
func sat_solver3_add_buffer_enable(pSat *sat_solver3, iVarA int32, iVarB int32, iVarEn int32, fCompl int32) int32 {
    var Lits [3]lit;
    var Cid int32;
    assert(iVarA >= 0 && iVarB >= 0 && iVarEn >= 0, "iVarA >= 0 && iVarB >= 0 && iVarEn >= 0");

    Lits[0] = toLitCond( iVarA, 0 );
    Lits[1] = toLitCond( iVarB, !fCompl );
    Lits[2] = toLitCond( iVarEn, 1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarA, 1 );
    Lits[1] = toLitCond( iVarB, fCompl );
    Lits[2] = toLitCond( iVarEn, 1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");
    return 2;
}
func sat_solver3_add_and(pSat *sat_solver3, iVar int32, iVar0 int32, iVar1 int32, fCompl0 int32, fCompl1 int32, fCompl int32) int32 {
    var Lits [3]lit;
    var Cid int32;

    Lits[0] = toLitCond( iVar, !fCompl );
    Lits[1] = toLitCond( iVar0, fCompl0 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 2 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVar, !fCompl );
    Lits[1] = toLitCond( iVar1, fCompl1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 2 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVar, fCompl );
    Lits[1] = toLitCond( iVar0, !fCompl0 );
    Lits[2] = toLitCond( iVar1, !fCompl1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");
    return 3;
}
func sat_solver3_add_xor(pSat *sat_solver3, iVarA int32, iVarB int32, iVarC int32, fCompl int32) int32 {
    var Lits [3]lit;
    var Cid int32;
    assert(iVarA >= 0 && iVarB >= 0 && iVarC >= 0, "iVarA >= 0 && iVarB >= 0 && iVarC >= 0");

    Lits[0] = toLitCond( iVarA, !fCompl );
    Lits[1] = toLitCond( iVarB, 1 );
    Lits[2] = toLitCond( iVarC, 1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarA, !fCompl );
    Lits[1] = toLitCond( iVarB, 0 );
    Lits[2] = toLitCond( iVarC, 0 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarA, fCompl );
    Lits[1] = toLitCond( iVarB, 1 );
    Lits[2] = toLitCond( iVarC, 0 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarA, fCompl );
    Lits[1] = toLitCond( iVarB, 0 );
    Lits[2] = toLitCond( iVarC, 1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");
    return 4;
}
func sat_solver3_add_mux(pSat *sat_solver3, iVarZ int32, iVarC int32, iVarT int32, iVarE int32, iComplC int32, iComplT int32, iComplE int32, iComplZ int32) int32 {
    var Lits [3]lit;
    var Cid int32;
    assert(iVarC >= 0 && iVarT >= 0 && iVarE >= 0 && iVarZ >= 0, "iVarC >= 0 && iVarT >= 0 && iVarE >= 0 && iVarZ >= 0");

    Lits[0] = toLitCond( iVarC, 1 ^ iComplC );
    Lits[1] = toLitCond( iVarT, 1 ^ iComplT );
    Lits[2] = toLitCond( iVarZ, 0 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarC, 1 ^ iComplC );
    Lits[1] = toLitCond( iVarT, 0 ^ iComplT );
    Lits[2] = toLitCond( iVarZ, 1 ^ iComplZ );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarC, 0 ^ iComplC );
    Lits[1] = toLitCond( iVarE, 1 ^ iComplE );
    Lits[2] = toLitCond( iVarZ, 0 ^ iComplZ );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarC, 0 ^ iComplC );
    Lits[1] = toLitCond( iVarE, 0 ^ iComplE );
    Lits[2] = toLitCond( iVarZ, 1 ^ iComplZ );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");

    if iVarT == iVarE {
        return 4; }

    Lits[0] = toLitCond( iVarT, 0 ^ iComplT );
    Lits[1] = toLitCond( iVarE, 0 ^ iComplE );
    Lits[2] = toLitCond( iVarZ, 1 ^ iComplZ );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarT, 1 ^ iComplT );
    Lits[1] = toLitCond( iVarE, 1 ^ iComplE );
    Lits[2] = toLitCond( iVarZ, 0 ^ iComplZ );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");
    return 6;
}
func sat_solver3_add_mux41(pSat *sat_solver3, iVarZ int32, iVarC0 int32, iVarC1 int32, iVarD0 int32, iVarD1 int32, iVarD2 int32, iVarD3 int32) int32 {
    var Lits [4]lit;
    var Cid int32;
    assert(iVarC0 >= 0 && iVarC1 >= 0 && iVarD0 >= 0 && iVarD1 >= 0 && iVarD2 >= 0 && iVarD3 >= 0 && iVarZ >= 0, "iVarC0 >= 0 && iVarC1 >= 0 && iVarD0 >= 0 && iVarD1 >= 0 && iVarD2 >= 0 && iVarD3 >= 0 && iVarZ >= 0");

    Lits[0] = toLitCond( iVarD0, 1 );
    Lits[1] = toLitCond( iVarC0, 0 );
    Lits[2] = toLitCond( iVarC1, 0 );
    Lits[3] = toLitCond( iVarZ,  0 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 4 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarD1, 1 );
    Lits[1] = toLitCond( iVarC0, 1 );
    Lits[2] = toLitCond( iVarC1, 0 );
    Lits[3] = toLitCond( iVarZ,  0 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 4 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarD2, 1 );
    Lits[1] = toLitCond( iVarC0, 0 );
    Lits[2] = toLitCond( iVarC1, 1 );
    Lits[3] = toLitCond( iVarZ,  0 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 4 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarD3, 1 );
    Lits[1] = toLitCond( iVarC0, 1 );
    Lits[2] = toLitCond( iVarC1, 1 );
    Lits[3] = toLitCond( iVarZ,  0 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 4 );
    assert((Cid != 0), "Cid");


    Lits[0] = toLitCond( iVarD0, 0 );
    Lits[1] = toLitCond( iVarC0, 0 );
    Lits[2] = toLitCond( iVarC1, 0 );
    Lits[3] = toLitCond( iVarZ,  1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 4 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarD1, 0 );
    Lits[1] = toLitCond( iVarC0, 1 );
    Lits[2] = toLitCond( iVarC1, 0 );
    Lits[3] = toLitCond( iVarZ,  1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 4 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarD2, 0 );
    Lits[1] = toLitCond( iVarC0, 0 );
    Lits[2] = toLitCond( iVarC1, 1 );
    Lits[3] = toLitCond( iVarZ,  1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 4 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarD3, 0 );
    Lits[1] = toLitCond( iVarC0, 1 );
    Lits[2] = toLitCond( iVarC1, 1 );
    Lits[3] = toLitCond( iVarZ,  1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 4 );
    assert((Cid != 0), "Cid");
    return 8;
}
func sat_solver3_add_xor_and(pSat *sat_solver3, iVarF int32, iVarA int32, iVarB int32, iVarC int32) int32 {
    // F = (a (+) b) * c
    var Lits [4]lit;
    var Cid int32;
    assert(iVarF >= 0 && iVarA >= 0 && iVarB >= 0 && iVarC >= 0, "iVarF >= 0 && iVarA >= 0 && iVarB >= 0 && iVarC >= 0");

    Lits[0] = toLitCond( iVarF, 1 );
    Lits[1] = toLitCond( iVarA, 1 );
    Lits[2] = toLitCond( iVarB, 1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarF, 1 );
    Lits[1] = toLitCond( iVarA, 0 );
    Lits[2] = toLitCond( iVarB, 0 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarF, 1 );
    Lits[1] = toLitCond( iVarC, 0 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 2 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarF, 0 );
    Lits[1] = toLitCond( iVarA, 1 );
    Lits[2] = toLitCond( iVarB, 0 );
    Lits[3] = toLitCond( iVarC, 1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 4 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarF, 0 );
    Lits[1] = toLitCond( iVarA, 0 );
    Lits[2] = toLitCond( iVarB, 1 );
    Lits[3] = toLitCond( iVarC, 1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 4 );
    assert((Cid != 0), "Cid");
    return 5;
}
func sat_solver3_add_constraint(pSat *sat_solver3, iVar int32, iVar2 int32, fCompl int32) int32 {
    var Lits [2]lit;
    var Cid int32;
    assert(iVar >= 0, "iVar >= 0");

    Lits[0] = toLitCond( iVar, fCompl );
    Lits[1] = toLitCond( iVar2, 0 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 2 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVar, fCompl );
    Lits[1] = toLitCond( iVar2, 1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 2 );
    assert((Cid != 0), "Cid");
    return 2;
}

func sat_solver3_add_half_sorter(pSat *sat_solver3, iVarA int32, iVarB int32, iVar0 int32, iVar1 int32) int32 {
    var Lits [3]lit;
    var Cid int32;

    Lits[0] = toLitCond( iVarA, 0 );
    Lits[1] = toLitCond( iVar0, 1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 2 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarA, 0 );
    Lits[1] = toLitCond( iVar1, 1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 2 );
    assert((Cid != 0), "Cid");

    Lits[0] = toLitCond( iVarB, 0 );
    Lits[1] = toLitCond( iVar0, 1 );
    Lits[2] = toLitCond( iVar1, 1 );
    Cid = sat_solver3_addclause( pSat, Lits, Lits + 3 );
    assert((Cid != 0), "Cid");
    return 3;
}


//ABC_NAMESPACE_HEADER_END



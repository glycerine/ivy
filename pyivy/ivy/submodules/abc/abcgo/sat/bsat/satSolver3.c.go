package bsat

import (
	"fmt"
	"math"
)

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
//#include <assert.h>
//#include <string.h>
//#include <math.h>

//#include "satSolver3.h"

//ABC_NAMESPACE_IMPL_START

#define SAT_USE_ANALYZE_FINAL

//=================================================================================================
// Debug:

//#define VERBOSEDEBUG

// For derivation output (verbosity level 2)
const L_IND =    "%-*d"
#define L_ind    sat_solver3_dl(s)*2+2,sat_solver3_dl(s)
const L_LIT =    "%sx%d"
// Original macro:
// #define L_lit(p) lit_sign(p)?"~":"", (lit_var(p))


// Just like 'assert()' but expression will be evaluated in the release version as well.
func check(expr int32) { assert((expr != 0), "expr"); }

func printlits(begin *lit, end *lit) {
    var i int32;
    for i = 0; i < end - begin; i++ {
        // inlining of macro L_lit
        fmt.Printf(L_LIT" ", lit_sign(begin[i])?"~":"", (lit_var(begin[i])));
    }
}

//=================================================================================================
// Random numbers:


// Returns a random float 0 <= x < 1. Seed must never be 0.
func drand(seed *float64) float64 {
    var q int32;
    *seed *= 1389796;
    q = int32((*seed / 2147483647));
    *seed -= float64(q) * 2147483647;
    return *seed / 2147483647; }


// Returns a random integer 0 <= x < size. Seed must never be 0.
func irand(seed *float64, size int32) int32 {
    return int32((drand(seed) * size)); }


//=================================================================================================
// Variable datatype + minor functions:

var var0 int32 = 1;
var var1 int32;
var varX int32 = 3;

type varinfo_t struct {
    val uint32; /* C bitfield: 2 bits; review packing and truncation */  // variable value 
    pol uint32; /* C bitfield: 1 bits; review packing and truncation */  // last polarity
    tag uint32; /* C bitfield: 1 bits; review packing and truncation */  // conflict analysis tag
    lev uint32; /* C bitfield: 28 bits; review packing and truncation */  // variable level
}

func var_level(s *sat_solver3, v int32) int32 { return s.levels[v];   }
func var_value(s *sat_solver3, v int32) int32 { return s.assigns[v];  }
func var_polar(s *sat_solver3, v int32) int32 { return s.polarity[v]; }

func var_set_level(s *sat_solver3, v int32, lev int32) { s.levels[v] = lev;    }
func var_set_value(s *sat_solver3, v int32, val int32) { s.assigns[v] = val;   }
func var_set_polar(s *sat_solver3, v int32, pol int32) { s.polarity[v] = pol;  }

// variable tags
func var_tag(s *sat_solver3, v int32) int32 { return s.tags[v]; }
func var_set_tag(s *sat_solver3, v int32, tag int32) { 
    assert(tag > 0 && tag < 16, "tag > 0 && tag < 16");
    if s.tags[v] == 0 {
        veci_push( &s.tagged, v ); }
    s.tags[v] = tag;                           
}
func var_add_tag(s *sat_solver3, v int32, tag int32) { 
    assert(tag > 0 && tag < 16, "tag > 0 && tag < 16");
    if s.tags[v] == 0 {
        veci_push( &s.tagged, v ); }
    s.tags[v] |= tag;                           
}
func solver2_clear_tags(s *sat_solver3, start int32) { 
    var i int32;
    var tagged *int32 = veci_begin(&s.tagged);
    for i = start; i < veci_size(&s.tagged); i++ {
        s.tags[tagged[i]] = 0;
    }
    veci_resize(&s.tagged,start);
}

func sat_solver3_get_var_value(s *sat_solver3, v int32) int32 {
    if var_value(s, v) == var0 {
        return l_False; }
    if var_value(s, v) == var1 {
        return l_True; }
    if var_value(s, v) == varX {
        return l_Undef; }
    assert(false, "0");
    return 0;
}

//=================================================================================================
// Simple helpers:

func sat_solver3_dl(s *sat_solver3) int32 { return veci_size(&s.trail_lim); }
func sat_solver3_read_wlist(s *sat_solver3, l lit) *veci { return &s.wlists[l];            }

//=================================================================================================
// Variable order functions:

static inline void order_update(sat_solver3* s, int v) // updateorder
{
    var orderpos *int32  = s.orderpos;
    var heap *int32      = veci_begin(&s.order);
    var i int32         = orderpos[v];
    var x int32         = heap[i];
    var parent int32    = (i - 1) / 2;

    assert(s.orderpos[v] != -1, "s.orderpos[v] != -1");

    for i != 0 && s.activity[x] > s.activity[heap[parent]]{
        heap[i]           = heap[parent];
        orderpos[heap[i]] = i;
        i                 = parent;
        parent            = (i - 1) / 2;
    }

    heap[i]     = x;
    orderpos[x] = i;
}

func order_assigned(s *sat_solver3, v int32) {
}

static inline void order_unassigned(sat_solver3* s, int v) // undoorder
{
    var orderpos *int32  = s.orderpos;
    if orderpos[v] == -1{
        orderpos[v] = veci_size(&s.order);
        veci_push(&s.order,v);
        order_update(s,v);
//printf( "+%d ", v );
    }
}

static inline int  order_select(sat_solver3* s, float random_var_freq) // selectvar
{
    var heap *int32      = veci_begin(&s.order);
    var orderpos *int32  = s.orderpos;
    // Random decision:
    if drand(&s.random_seed) < random_var_freq{
        var next int32  = irand(&s.random_seed,s.size);
        assert(next >= 0 && next < s.size, "next >= 0 && next < s.size");
        if var_value(s, next) == varX {
            return next; }
    }
    // Activity based decision:
    for veci_size(&s.order) > 0{
        var next int32   = heap[0];
        var size int32   = veci_size(&s.order)-1;
        var x int32      = heap[size];
        veci_resize(&s.order,size);
        orderpos[next] = -1;
        if size > 0{
            var i int32;
            var child int32  = 1;
            for child < size{

                if child+1 < size && s.activity[heap[child]] < s.activity[heap[child+1]] {
                    child++; }
                assert(child < size, "child < size");
                if s.activity[x] >= s.activity[heap[child]] {
                    break; }

                heap[i]           = heap[child];
                orderpos[heap[i]] = i;
                i                 = child;
                child             = 2 * child + 1;
            }
            heap[i]           = x;
            orderpos[heap[i]] = i;
        }
        if var_value(s, next) == varX {
            return next; }
    }
    return var_Undef;
}

func sat_solver3_set_var_activity(s *sat_solver3, pVars *int32, nVars int32) {
    var i int32;
    assert(s.VarActType == 1, "s.VarActType == 1");
    for i = 0; i < s.size; i++ {
        s.activity[i] = 0;
    }
    s.var_inc = Abc_Dbl2Word(1);
    for  i = 0; i < nVars; i++  {
        var iVar int32;
        if pVars != nil {
            iVar = pVars[i];
        } else {
            iVar = i;
        }
        s.activity[iVar] = Abc_Dbl2Word(nVars-i);
        order_update( s, iVar );
    }
}

//=================================================================================================
// variable activities

func solver_init_activities(s *sat_solver3) {
    // variable activities
    s.VarActType             = 0;
    if s.VarActType == 0 {
        s.var_inc            = (1 <<  5);
        s.var_decay          = -1;
    
    } else if s.VarActType == 1 {
        s.var_inc            = Abc_Dbl2Word(1.0);
        s.var_decay          = Abc_Dbl2Word(1.0 / 0.95);
    
    } else if s.VarActType == 2 {
        s.var_inc            = Xdbl_FromDouble(1.0);
        s.var_decay          = Xdbl_FromDouble(1.0 / 0.950);
    
    } else { assert(false, "0"); }

    // clause activities
    s.ClaActType             = 0;
    if s.ClaActType == 0 {
        s.cla_inc            = (1 << 11);
        s.cla_decay          = -1;
    
    } else {
        s.cla_inc            = 1;
        s.cla_decay          = float32((1 / 0.999));
    }
}

func act_var_rescale(s *sat_solver3) {
    if s.VarActType == 0 {
        var activity *word  = s.activity;
        var i int32;
        for i = 0; i < s.size; i++ {
            activity[i] >>= 19;
        }
        s.var_inc >>= 19;
        s.var_inc = Abc_MaxInt( uint32(s.var_inc), (1<<4) );
    
    } else if s.VarActType == 1 {
        var activity *float64  = (double*)s.activity;
        var i int32;
        for i = 0; i < s.size; i++ {
            activity[i] *= 1e-100;
        }
        s.var_inc = Abc_Dbl2Word( Abc_Word2Dbl(s.var_inc) * 1e-100 );
        //printf( "Rescaling var activity...\n" ); 
    
    } else if s.VarActType == 2 {
        var activity *xdbl  = s.activity;
        var i int32;
        for i = 0; i < s.size; i++ {
            activity[i] = Xdbl_Div( activity[i], 200 );
        } // activity[i] / 2^200
        s.var_inc = Xdbl_Div( s.var_inc, 200 ); 
    
    } else { assert(false, "0"); }
}
func act_var_bump(s *sat_solver3, v int32) {
    if s.VarActType == 0 {
        s.activity[v] += s.var_inc;
        if uint32(s.activity[v]) & 0x80000000 {
            act_var_rescale(s); }
        if s.orderpos[v] != -1 {
            order_update(s,v); }
    
    } else if s.VarActType == 1 {
        var act float64  = Abc_Word2Dbl(s.activity[v]) + Abc_Word2Dbl(s.var_inc);
        s.activity[v] = Abc_Dbl2Word(act);
        if act > 1e100 {
            act_var_rescale(s); }
        if s.orderpos[v] != -1 {
            order_update(s,v); }
    
    } else if s.VarActType == 2 {
        s.activity[v] = Xdbl_Add( s.activity[v], s.var_inc );
        if s.activity[v] > ABC_CONST(0x014c924d692ca61b) {
            act_var_rescale(s); }
        if s.orderpos[v] != -1 {
            order_update(s,v); }
    
    } else { assert(false, "0"); }
}
func act_var_bump_global(s *sat_solver3, v int32) {
    if !s.pGlobalVars || !s.pGlobalVars[v] {
        return; }
    if s.VarActType == 0 {
        s.activity[v] += int32((uint32(s.var_inc) * 3));
        if s.activity[v] & 0x80000000 {
            act_var_rescale(s); }
        if s.orderpos[v] != -1 {
            order_update(s,v); }
    
    } else if s.VarActType == 1 {
        var act float64  = Abc_Word2Dbl(s.activity[v]) + Abc_Word2Dbl(s.var_inc) * 3.0;
        s.activity[v] = Abc_Dbl2Word(act);
        if act > 1e100 {
            act_var_rescale(s); }
        if s.orderpos[v] != -1 {
            order_update(s,v); }
    
    } else if s.VarActType == 2 {
        s.activity[v] = Xdbl_Add( s.activity[v], Xdbl_Mul(s.var_inc, Xdbl_FromDouble(3.0)) );
        if s.activity[v] > ABC_CONST(0x014c924d692ca61b) {
            act_var_rescale(s); }
        if s.orderpos[v] != -1 {
            order_update(s,v); }
    
    } else { assert(false, "0"); }
}
func act_var_bump_factor(s *sat_solver3, v int32) {
    if !s.factors {
        return; }
    if s.VarActType == 0 {
        s.activity[v] += int32((uint32(s.var_inc) * float32(s.factors[v])));
        if s.activity[v] & 0x80000000 {
            act_var_rescale(s); }
        if s.orderpos[v] != -1 {
            order_update(s,v); }
    
    } else if s.VarActType == 1 {
        var act float64  = Abc_Word2Dbl(s.activity[v]) + Abc_Word2Dbl(s.var_inc) * s.factors[v];
        s.activity[v] = Abc_Dbl2Word(act);
        if act > 1e100 {
            act_var_rescale(s); }
        if s.orderpos[v] != -1 {
            order_update(s,v); }
    
    } else if s.VarActType == 2 {
        s.activity[v] = Xdbl_Add( s.activity[v], Xdbl_Mul(s.var_inc, Xdbl_FromDouble(s.factors[v])) );
        if s.activity[v] > ABC_CONST(0x014c924d692ca61b) {
            act_var_rescale(s); }
        if s.orderpos[v] != -1 {
            order_update(s,v); }
    
    } else { assert(false, "0"); }
}

func act_var_decay(s *sat_solver3) { 
    if s.VarActType == 0 {
        s.var_inc += (s.var_inc >>  4); 
    } else if s.VarActType == 1 {
        s.var_inc = Abc_Dbl2Word( Abc_Word2Dbl(s.var_inc) * Abc_Word2Dbl(s.var_decay) );
    } else if s.VarActType == 2 {
        s.var_inc = Xdbl_Mul(s.var_inc, s.var_decay); 
    } else { assert(false, "0"); }
}

// clause activities
func act_clause_rescale(s *sat_solver3) {
    if s.ClaActType == 0 {
        var activity *uint32  = (unsigned *)veci_begin(&s.act_clas);
        var i int32;
        for i = 0; i < veci_size(&s.act_clas); i++ {
            activity[i] >>= 14;
        }
        s.cla_inc >>= 14;
        s.cla_inc = Abc_MaxInt( s.cla_inc, (1<<10) );
    
    } else {
        var activity *float32  = (float *)veci_begin(&s.act_clas);
        var i int32;
        for i = 0; i < veci_size(&s.act_clas); i++ {
            activity[i] *= float32(1e-20);
        }
        s.cla_inc *= float32(1e-20);
    }
}
func act_clause_bump(s *sat_solver3, c *clause) {
    if s.ClaActType == 0 {
        var act *uint32  = (unsigned *)veci_begin(&s.act_clas) + c.lits[c.size];
        *act += s.cla_inc;
        if *act & 0x80000000 {
            act_clause_rescale(s); }
    
    } else {
        var act *float32  = (float *)veci_begin(&s.act_clas) + c.lits[c.size];
        *act += s.cla_inc;
        if *act > 1e20 { 
            act_clause_rescale(s); }
    }
}
func act_clause_decay(s *sat_solver3) { 
    if s.ClaActType == 0 {
        s.cla_inc += (s.cla_inc >> 10);
    } else {
        s.cla_inc *= s.cla_decay; }
}


//=================================================================================================
// Sorting functions (sigh):

func selectionsort(array *any, size int32, comp func(any, any) int32) {
    var i, j, best_i int32;
    var tmp any;

    for i = 0; i < size-1; i++{
        best_i = i;
        for j = i+1; j < size; j++{
            if comp(array[j], array[best_i]) < 0 {
                best_i = j; }
        }
        tmp = array[i]; array[i] = array[best_i]; array[best_i] = tmp;
    }
}

func sortrnd(array *any, size int32, comp func(any, any) int32, seed *float64) {
    if size <= 15 {
        selectionsort(array, size, comp);

    } else{
        var pivot any  = array[irand(seed, size)];
        var tmp any;
        var i int32  = -1;
        var j int32  = size;

        for {
            do i++; while(comp(array[i], pivot)<0);
            do j--; while(comp(pivot, array[j])<0);

            if i >= j { break; }

            tmp = array[i]; array[i] = array[j]; array[j] = tmp;
        }

        sortrnd(array    , i     , comp, seed);
        sortrnd(&array[i], size-i, comp, seed);
    }
}

//=================================================================================================
// Clause functions:

func sat_clause_compute_lbd(s *sat_solver3, c *clause) int32 {
    var i, lev, minl, lbd int32;
    for i = 0; i < int32(c.size); i++ {
        lev = var_level(s, lit_var(c.lits[i]));
        if !(minl & (1 << (lev & 31))) {
            minl |= 1 << (lev & 31);
            lbd++;
//            printf( "%d ", lev );
        }
    }
//    printf( " -> %d\n", lbd );
    return lbd;
}

/* pre: size > 1 && no variable occurs twice
 */
func sat_solver3_clause_new(s *sat_solver3, begin *lit, end *lit, learnt int32) int32 {
    var fUseBinaryClauses int32  = 1;
    var size int32;
    var c *clause;
    var h int32;

    assert(end - begin > 1, "end - begin > 1");
    assert(learnt >= 0 && learnt < 2, "learnt >= 0 && learnt < 2");
    size           = end - begin;

    // do not allocate memory for the two-literal problem clause
    if (fUseBinaryClauses != 0) && size == 2 && !learnt {
        veci_push(sat_solver3_read_wlist(s,lit_neg(begin[0])),(clause_from_lit(begin[1])));
        veci_push(sat_solver3_read_wlist(s,lit_neg(begin[1])),(clause_from_lit(begin[0])));
        s.stats.clauses++;
        s.stats.clauses_literals += size;
        return 0;
    }

    // create new clause
//    h = Vec_SetAppend( &s->Mem, NULL, size + learnt + 1 + 1 ) << 1;
    h = Sat_MemAppend( &s.Mem, begin, size, learnt, 0 );
    assert(!(h & 1), "!(h & 1)");
    if s.hLearnts == -1 && learnt {
        s.hLearnts = h; }
    if learnt {
        c = clause_read( s, h );
        c.lbd = sat_clause_compute_lbd( s, c );
        assert(clause_id(c) == veci_size(&s.act_clas), "clause_id(c) == veci_size(&s.act_clas)");
//        veci_push(&s->learned, h);
//        act_clause_bump(s,clause_read(s, h));
        if s.ClaActType == 0 {
            veci_push(&s.act_clas, (1<<10));
        } else {
            veci_push(&s.act_clas, s.cla_inc); }
        s.stats.learnts++;
        s.stats.learnts_literals += size;
    
    } else {
        s.stats.clauses++;
        s.stats.clauses_literals += size;
    }

    assert(begin[0] >= 0, "begin[0] >= 0");
    assert(begin[0] < s.size*2, "begin[0] < s.size*2");
    assert(begin[1] >= 0, "begin[1] >= 0");
    assert(begin[1] < s.size*2, "begin[1] < s.size*2");

    assert(lit_neg(begin[0]) < s.size*2, "lit_neg(begin[0]) < s.size*2");
    assert(lit_neg(begin[1]) < s.size*2, "lit_neg(begin[1]) < s.size*2");

    //veci_push(sat_solver3_read_wlist(s,lit_neg(begin[0])),c);
    //veci_push(sat_solver3_read_wlist(s,lit_neg(begin[1])),c);
    veci_push(sat_solver3_read_wlist(s,lit_neg(begin[0])),(size > 2 ? h : clause_from_lit(begin[1])));
    veci_push(sat_solver3_read_wlist(s,lit_neg(begin[1])),(size > 2 ? h : clause_from_lit(begin[0])));

    return h;
}


//=================================================================================================
// Minor (solver) functions:

func sat_solver3_enqueue(s *sat_solver3, l lit, from int32) int32 {
    var v int32   = lit_var(l);
    if s.pFreqs[v] == 0 {
//    {
        s.pFreqs[v] = 1; }
//        s->nVarUsed++;
//    }

#ifdef VERBOSEDEBUG
    printf(L_IND"enqueue("L_LIT")\n", L_ind, L_lit(l));
#endif
    if var_value(s, v) != varX {
        return var_value(s, v) == lit_sign(l);
    } else{
/*
        if ( s->pCnfFunc )
        {
            if ( lit_sign(l) )
            {
                if ( (s->loads[v] & 1) == 0 )
                {
                    s->loads[v] ^= 1;
                    s->pCnfFunc( s->pCnfMan, l );
                }
            }
            else
            {
                if ( (s->loads[v] & 2) == 0 )
                {
                    s->loads[v] ^= 2;
                    s->pCnfFunc( s->pCnfMan, l );
                }
            }
        }
*/
        // New fact -- store it.
#ifdef VERBOSEDEBUG
        printf(L_IND"bind("L_LIT")\n", L_ind, L_lit(l));
#endif
        var_set_value(s, v, lit_sign(l));
        var_set_level(s, v, sat_solver3_dl(s));
        s.reasons[v] = from;
        s.trail[s.qtail++] = l;
        order_assigned(s, v);
        return true;
    }
}


func sat_solver3_decision(s *sat_solver3, l lit) int32 {
    assert(s.qtail == s.qhead, "s.qtail == s.qhead");
    assert(var_value(s, lit_var(l)) == varX, "var_value(s, lit_var(l)) == varX");
#ifdef VERBOSEDEBUG
    printf(L_IND"assume("L_LIT")  ", L_ind, L_lit(l));
    printf( "act = %.20f\n", s->activity[lit_var(l)] );
#endif
    veci_push(&s.trail_lim,s.qtail);
    return sat_solver3_enqueue(s,l,0);
}


func sat_solver3_canceluntil(s *sat_solver3, level int32) {
    var bound int32;
    var lastLev int32;
    var c int32;
    
    if sat_solver3_dl(s) <= level {
        return; }

    assert(veci_size(&s.trail_lim) > 0, "veci_size(&s.trail_lim) > 0");
    bound   = (veci_begin(&s.trail_lim))[level];
    lastLev = (veci_begin(&s.trail_lim))[veci_size(&s.trail_lim)-1];

    ////////////////////////////////////////
    // added to cancel all assignments
//    if ( level == -1 )
//        bound = 0;
    ////////////////////////////////////////

    for c = s.qtail-1; c >= bound; c-- {
        var x int32   = lit_var(s.trail[c]);
        var_set_value(s, x, varX);
        s.reasons[x] = 0;
        if c < lastLev {
            var_set_polar( s, x, !lit_sign(s.trail[c]) ); }
    }
    //printf( "\n" );

    for c = s.qhead-1; c >= bound; c-- {
        order_unassigned(s,lit_var(s.trail[c]));
    }

    s.qhead = s.qtail = bound;
    veci_resize(&s.trail_lim,level);
}

func sat_solver3_canceluntil_rollback(s *sat_solver3, NewBound int32) {
    var c, x int32;
   
    assert(sat_solver3_dl(s) == 0, "sat_solver3_dl(s) == 0");
    assert(s.qtail == s.qhead, "s.qtail == s.qhead");
    assert(s.qtail >= NewBound, "s.qtail >= NewBound");

    for c = s.qtail-1; c >= NewBound; c-- {
        x = lit_var(s.trail[c]);
        var_set_value(s, x, varX);
        s.reasons[x] = 0;
    }

    for c = s.qhead-1; c >= NewBound; c-- {
        order_unassigned(s,lit_var(s.trail[c]));
    }

    s.qhead = s.qtail = NewBound;
}

func sat_solver3_record(s *sat_solver3, cls *veci) {
    var begin *lit  = veci_begin(cls);
    var end *lit    = begin + veci_size(cls);
    var h int32;
    if veci_size(cls) > 1 {
        h = sat_solver3_clause_new(s,begin,end,1);
    } else {
        h = 0;
    }
    sat_solver3_enqueue(s,*begin,h);
    assert(veci_size(cls) > 0, "veci_size(cls) > 0");
    if h == 0 {
        veci_push( &s.unit_lits, *begin ); }

/*
    if (h != 0) {
        act_clause_bump(s,clause_read(s, h));
        s->stats.learnts++;
        s->stats.learnts_literals += veci_size(cls);
    }
*/
}

func sat_solver3_count_assigned(s *sat_solver3) int32 {
    // count top-level assignments
    var i, Count int32;
    assert(sat_solver3_dl(s) == 0, "sat_solver3_dl(s) == 0");
    for  i = 0; i < s.size; i++  {
        if var_value(s, i) != varX {
            Count++; }
    }
    return Count;
}

func sat_solver3_progress(s *sat_solver3) float64 {
    var i int32;
    var progress float64;
    var F float64         = 1.0 / s.size;
    for i = 0; i < s.size; i++ {
        if var_value(s, i) != varX {
            progress += math.Pow(float64(F), float64(var_level(s, i))); }
    }
    return progress / s.size;
}

//=================================================================================================
// Major methods:

func sat_solver3_lit_removable(s *sat_solver3, x int32, minl int32) int32 {
    var top int32      = veci_size(&s.tagged);

    assert(s.reasons[x] != 0, "s.reasons[x] != 0");
    veci_resize(&s.stack,0);
    veci_push(&s.stack,x);

    for veci_size(&s.stack){
        var v int32  = veci_pop(&s.stack);
        assert(s.reasons[v] != 0, "s.reasons[v] != 0");
        if clause_is_lit(s.reasons[v]){
            v = lit_var(clause_read_lit(s.reasons[v]));
            if !(var_tag(s,v) != 0) && (var_level(s, v) != 0){
                if s.reasons[v] != 0 && ((1 << (var_level(s, v) & 31)) & minl){
                    veci_push(&s.stack,v);
                    var_set_tag(s, v, 1);
                }else{
                    solver2_clear_tags(s, top);
                    return 0;
                }
            }
        }else{
            var c *clause  = clause_read(s, s.reasons[v]);
            var lits *lit  = clause_begin(c);
            var i int32;
            for i = 1; i < clause_size(c); i++{
                var v int32  = lit_var(lits[i]);
                if !(var_tag(s,v) != 0) && (var_level(s, v) != 0){
                    if s.reasons[v] != 0 && ((1 << (var_level(s, v) & 31)) & minl){
                        veci_push(&s.stack,lit_var(lits[i]));
                        var_set_tag(s, v, 1);
                    }else{
                        solver2_clear_tags(s, top);
                        return 0;
                    }
                }
            }
        }
    }
    return 1;
}


/*_________________________________________________________________________________________________
|
|  analyzeFinal : (p : Lit)  ->  [void]
|  
|  Description:
|    Specialized analysis procedure to express the final conflict in terms of assumptions.
|    Calculates the (possibly empty) set of assumptions that led to the assignment of 'p', and
|    stores the result in 'out_conflict'.
|________________________________________________________________________________________________@*/
/*
void Solver::analyzeFinal(Clause* confl, bool skip_first)
{
    // -- NOTE! This code is relatively untested. Please report bugs!
    conflict.clear();
    if (root_level == 0) return;

    vec<char>& seen  = analyze_seen;
    for (int i = skip_first ? 1 : 0; i < confl->size(); i++){
        Var x = var((*confl)[i]);
        if (level[x] > 0)
            seen[x] = 1;
    }

    int start = (root_level >= trail_lim.size()) ? trail.size()-1 : trail_lim[root_level];
    for (int i = start; i >= trail_lim[0]; i--){
        Var     x = var(trail[i]);
        if (seen[x]){
            GClause r = reason[x];
            if (r == GClause_NULL){
                assert(level[x] > 0);
                conflict.push(~trail[i]);
            }else{
                if (r.isLit()){
                    Lit p = r.lit();
                    if (level[var(p)] > 0)
                        seen[var(p)] = 1;
                }else{
                    Clause& c = *r.clause();
                    for (int j = 1; j < c.size(); j++)
                        if (level[var(c[j])] > 0)
                            seen[var(c[j])] = 1;
                }
            }
            seen[x] = 0;
        }
    }
}
*/

                            

func sat_solver3_analyze_final(s *sat_solver3, hConf int32, skip_first int32) {
    var conf *clause  = clause_read(s, hConf);
    var i, j, start int32;
    veci_resize(&s.conf_final,0);
    if s.root_level == 0 {
        return; }
    assert(veci_size(&s.tagged) == 0, "veci_size(&s.tagged) == 0");
//    assert( s->tags[lit_var(p)] == l_Undef );
//    s->tags[lit_var(p)] = l_True;
    for i = skip_first ? 1 : 0; i < clause_size(conf); i++ {
        var x int32  = lit_var(clause_begin(conf)[i]);
        if var_level(s, x) > 0 {
            var_set_tag(s, x, 1); }
    }

    if s.root_level >= veci_size(&s.trail_lim) {
        start = s.qtail-1;
    } else {
        start = (veci_begin(&s.trail_lim))[s.root_level];
    }
    for i = start; i >= (veci_begin(&s.trail_lim))[0]; i--{
        var x int32  = lit_var(s.trail[i]);
        if var_tag(s,x) != 0{
            if s.reasons[x] == 0{
                assert(var_level(s, x) > 0, "var_level(s, x) > 0");
                veci_push(&s.conf_final,lit_neg(s.trail[i]));
            }else{
                if clause_is_lit(s.reasons[x]){
                    var q lit  = clause_read_lit(s.reasons[x]);
                    assert(lit_var(q) >= 0 && lit_var(q) < s.size, "lit_var(q) >= 0 && lit_var(q) < s.size");
                    if var_level(s, lit_var(q)) > 0 {
                        var_set_tag(s, lit_var(q), 1); }
                
                } else{
                    var c *clause  = clause_read(s, s.reasons[x]);
                    var lits *int32  = clause_begin(c);
                    for j = 1; j < clause_size(c); j++ {
                        if var_level(s, lit_var(lits[j])) > 0 {
                            var_set_tag(s, lit_var(lits[j]), 1); }
                    }
                }
            }
        }
    }
    solver2_clear_tags(s,0);
}

      

func sat_solver3_analyze(s *sat_solver3, h int32, learnt *veci) {
    var trail *lit    = s.trail;
    var cnt int32;
    var p lit        = lit_Undef;
    var ind int32      = s.qtail-1;
    var lits *lit;
    var i, j, minl int32;
    veci_push(learnt,lit_Undef);
    for{
        assert(h != 0, "h != 0");
        if clause_is_lit(h){
            var x int32  = lit_var(clause_read_lit(h));
            if var_tag(s, x) == 0 && var_level(s, x) > 0{
                var_set_tag(s, x, 1);
                act_var_bump(s,x);
                if var_level(s, x) == sat_solver3_dl(s) {
                    cnt++;
                } else {
                    veci_push(learnt,clause_read_lit(h)); }
            }
        }else{
            var c *clause  = clause_read(s, h);
            
            if clause_learnt(c) {
                act_clause_bump(s,c); }
            lits = clause_begin(c);
            //printlits(lits,lits+clause_size(c)); printf("\n");
            for j = (p == lit_Undef ? 0 : 1); j < clause_size(c); j++{
                var x int32  = lit_var(lits[j]);
                if var_tag(s, x) == 0 && var_level(s, x) > 0{
                    var_set_tag(s, x, 1);
                    act_var_bump(s,x);
                    // bump variables propaged by the LBD=2 clause
//                    if ( s->reasons[x] && clause_read(s, s->reasons[x])->lbd <= 2 )
//                        act_var_bump(s,x);
                    if var_level(s,x) == sat_solver3_dl(s) {
                        cnt++;
                    } else {
                        veci_push(learnt,lits[j]); }
                }
            }
        }

        for  !var_tag(s, lit_var(trail[ind--]))  {; }

        p = trail[ind+1];
        h = s.reasons[lit_var(p)];
        cnt--;

    
if !(cnt > 0) { break }
}

    *veci_begin(learnt) = lit_neg(p);

    lits = veci_begin(learnt);
    minl = 0;
    for i = 1; i < veci_size(learnt); i++{
        var lev int32  = var_level(s, lit_var(lits[i]));
        minl    |= 1 << (lev & 31);
    }

    // simplify (full)
    for i = j = 1; i < veci_size(learnt); i++{
        if s.reasons[lit_var(lits[i])] == 0 || !(sat_solver3_lit_removable(s,lit_var(lits[i]),minl) != 0) {
            lits[j++] = lits[i]; }
    }

    // update size of learnt + statistics
    veci_resize(learnt,j);
    s.stats.tot_literals += j;


    // clear tags
    solver2_clear_tags(s,0);

#ifdef DEBUG
    for (i = 0; i < s->size; i++)
        assert(!var_tag(s, i));
#endif

#ifdef VERBOSEDEBUG
    printf(L_IND"Learnt {", L_ind);
    for (i = 0; i < veci_size(learnt); i++) printf(" "L_LIT, L_lit(lits[i]));
#endif
    if veci_size(learnt) > 1{
        var max_i int32  = 1;
        var max int32    = var_level(s, lit_var(lits[1]));
        var tmp lit;

        for i = 2; i < veci_size(learnt); i++ {
            if var_level(s, lit_var(lits[i])) > max{
                max   = var_level(s, lit_var(lits[i]));
                max_i = i;
            }
        }

        tmp         = lits[1];
        lits[1]     = lits[max_i];
        lits[max_i] = tmp;
    }
#ifdef VERBOSEDEBUG
    {
        int lev = veci_size(learnt) > 1 ? var_level(s, lit_var(lits[1])) : 0;
        printf(" } at level %d\n", lev);
    }
#endif
}

//#define TEST_CNF_LOAD

int sat_solver3_propagate(sat_solver3* s)
{
    int     hConfl = 0;
    lit*    lits;
    lit false_lit;

    //printf("sat_solver3_propagate\n");
    while (hConfl == 0 && s->qtail - s->qhead > 0){
        lit p = s->trail[s->qhead++];

#ifdef TEST_CNF_LOAD
        int v = lit_var(p);
        if ( s->pCnfFunc )
        {
            if ( lit_sign(p) )
            {
                if ( (s->loads[v] & 1) == 0 )
                {
                    s->loads[v] ^= 1;
                    s->pCnfFunc( s->pCnfMan, p );
                }
            }
            else
            {
                if ( (s->loads[v] & 2) == 0 )
                {
                    s->loads[v] ^= 2;
                    s->pCnfFunc( s->pCnfMan, p );
                }
            }
        }
        {
#endif

        veci* ws    = sat_solver3_read_wlist(s,p);
        int*  begin = veci_begin(ws);
        int*  end   = begin + veci_size(ws);
        int*i, *j;

        s->stats.propagations++;
//        s->simpdb_props--;

        //printf("checking lit %d: "L_LIT"\n", veci_size(ws), L_lit(p));
        for (i = j = begin; i < end; ){
            if (clause_is_lit(*i)){

                int Lit = clause_read_lit(*i);
                if (var_value(s, lit_var(Lit)) == lit_sign(Lit)){
                    *j++ = *i++;
                    continue;
                }

                *j++ = *i;
                if (!sat_solver3_enqueue(s,clause_read_lit(*i),clause_from_lit(p))){
                    hConfl = s->hBinary;
                    (clause_begin(s->binary))[1] = lit_neg(p);
                    (clause_begin(s->binary))[0] = clause_read_lit(*i++);
                    // Copy the remaining watches:
                    while (i < end)
                        *j++ = *i++;
                }
            }else{

                clause* c = clause_read(s,*i);
                lits = clause_begin(c);

                // Make sure the false literal is data[1]:
                false_lit = lit_neg(p);
                if (lits[0] == false_lit){
                    lits[0] = lits[1];
                    lits[1] = false_lit;
                }
                assert(lits[1] == false_lit);

                // If 0th watch is true, then clause is already satisfied.
                if (var_value(s, lit_var(lits[0])) == lit_sign(lits[0]))
                    *j++ = *i;
                else{
                    // Look for new watch:
                    lit* stop = lits + clause_size(c);
                    lit* k;
                    for (k = lits + 2; k < stop; k++){
                        if (var_value(s, lit_var(*k)) != !lit_sign(*k)){
                            lits[1] = *k;
                            *k = false_lit;
                            veci_push(sat_solver3_read_wlist(s,lit_neg(lits[1])),*i);
                            goto next; }
                    }

                    *j++ = *i;
                    // Clause is unit under assignment:
                    if ( c->lrn )
                        c->lbd = sat_clause_compute_lbd(s, c);
                    if (!sat_solver3_enqueue(s,lits[0], *i)){
                        hConfl = *i++;
                        // Copy the remaining watches:
                        while (i < end)
                            *j++ = *i++;
                    }
                }
            }
        next:
            i++;
        }

        s->stats.inspects += j - veci_begin(ws);
        veci_resize(ws,j - veci_begin(ws));
#ifdef TEST_CNF_LOAD
        }
#endif
    }

    return hConfl;
}

//=================================================================================================
// External solver functions:

sat_solver3* sat_solver3_new(void)
{
    sat_solver3* s = (sat_solver3*)ABC_CALLOC( char, sizeof(sat_solver3));

//    Vec_SetAlloc_(&s->Mem, 15);
    Sat_MemAlloc_(&s->Mem, 17);
    s->hLearnts = -1;
    s->hBinary = Sat_MemAppend( &s->Mem, NULL, 2, 0, 0 );
    s->binary = clause_read( s, s->hBinary );

    s->nLearntStart = LEARNT_MAX_START_DEFAULT;  // starting learned clause limit
    s->nLearntDelta = LEARNT_MAX_INCRE_DEFAULT;  // delta of learned clause limit
    s->nLearntRatio = LEARNT_MAX_RATIO_DEFAULT;  // ratio of learned clause limit
    s->nLearntMax   = s->nLearntStart;

    // initialize vectors
    veci_new(&s->order);
    veci_new(&s->trail_lim);
    veci_new(&s->tagged);
//    veci_new(&s->learned);
    veci_new(&s->act_clas);
    veci_new(&s->stack);
//    veci_new(&s->model);
    veci_new(&s->unit_lits);
    veci_new(&s->temp_clause);
    veci_new(&s->conf_final);

    // initialize arrays
    s->wlists    = 0;
    s->activity  = 0;
    s->orderpos  = 0;
    s->reasons   = 0;
    s->trail     = 0;

    // initialize other vars
    s->size                   = 0;
    s->cap                    = 0;
    s->qhead                  = 0;
    s->qtail                  = 0;

    solver_init_activities(s);
    veci_new(&s->act_vars);

    s->root_level             = 0;
//    s->simpdb_assigns         = 0;
//    s->simpdb_props           = 0;
    s->random_seed            = 91648253;
    s->progress_estimate      = 0;
//    s->binary                 = (clause*)ABC_ALLOC( char, sizeof(clause) + sizeof(lit)*2);
//    s->binary->size_learnt    = (2 << 1);
    s->verbosity              = 0;

    s->stats.starts           = 0;
    s->stats.decisions        = 0;
    s->stats.propagations     = 0;
    s->stats.inspects         = 0;
    s->stats.conflicts        = 0;
    s->stats.clauses          = 0;
    s->stats.clauses_literals = 0;
    s->stats.learnts          = 0;
    s->stats.learnts_literals = 0;
    s->stats.tot_literals     = 0;
    return s;
}

sat_solver3* zsat_solver3_new_seed(double seed)
{
    sat_solver3* s = (sat_solver3*)ABC_CALLOC( char, sizeof(sat_solver3));

//    Vec_SetAlloc_(&s->Mem, 15);
    Sat_MemAlloc_(&s->Mem, 15);
    s->hLearnts = -1;
    s->hBinary = Sat_MemAppend( &s->Mem, NULL, 2, 0, 0 );
    s->binary = clause_read( s, s->hBinary );

    s->nLearntStart = LEARNT_MAX_START_DEFAULT;  // starting learned clause limit
    s->nLearntDelta = LEARNT_MAX_INCRE_DEFAULT;  // delta of learned clause limit
    s->nLearntRatio = LEARNT_MAX_RATIO_DEFAULT;  // ratio of learned clause limit
    s->nLearntMax   = s->nLearntStart;

    // initialize vectors
    veci_new(&s->order);
    veci_new(&s->trail_lim);
    veci_new(&s->tagged);
//    veci_new(&s->learned);
    veci_new(&s->act_clas);
    veci_new(&s->stack);
//    veci_new(&s->model);
    veci_new(&s->unit_lits);
    veci_new(&s->temp_clause);
    veci_new(&s->conf_final);

    // initialize arrays
    s->wlists    = 0;
    s->activity  = 0;
    s->orderpos  = 0;
    s->reasons   = 0;
    s->trail     = 0;

    // initialize other vars
    s->size                   = 0;
    s->cap                    = 0;
    s->qhead                  = 0;
    s->qtail                  = 0;

    solver_init_activities(s);
    veci_new(&s->act_vars);

    s->root_level             = 0;
//    s->simpdb_assigns         = 0;
//    s->simpdb_props           = 0;
    s->random_seed            = seed;
    s->progress_estimate      = 0;
//    s->binary                 = (clause*)ABC_ALLOC( char, sizeof(clause) + sizeof(lit)*2);
//    s->binary->size_learnt    = (2 << 1);
    s->verbosity              = 0;

    s->stats.starts           = 0;
    s->stats.decisions        = 0;
    s->stats.propagations     = 0;
    s->stats.inspects         = 0;
    s->stats.conflicts        = 0;
    s->stats.clauses          = 0;
    s->stats.clauses_literals = 0;
    s->stats.learnts          = 0;
    s->stats.learnts_literals = 0;
    s->stats.tot_literals     = 0;
    return s;
}

void sat_solver3_setnvars(sat_solver3* s,int n)
{
    int var;

    if (s->cap < n){
        int old_cap = s->cap;
        while (s->cap < n) s->cap = s->cap*2+1;
        if ( s->cap < 50000 )
            s->cap = 50000;

        s->wlists    = ABC_REALLOC(veci,   s->wlists,   s->cap*2);
//        s->vi        = ABC_REALLOC(varinfo,s->vi,       s->cap);
        s->levels    = ABC_REALLOC(int,    s->levels,   s->cap);
        s->assigns   = ABC_REALLOC(char,   s->assigns,  s->cap);
        s->polarity  = ABC_REALLOC(char,   s->polarity, s->cap);
        s->tags      = ABC_REALLOC(char,   s->tags,     s->cap);
        s->loads     = ABC_REALLOC(char,   s->loads,    s->cap);
        s->activity  = ABC_REALLOC(word,   s->activity, s->cap);
        s->activity2 = ABC_REALLOC(word,   s->activity2,s->cap);
        s->pFreqs    = ABC_REALLOC(char,   s->pFreqs,   s->cap);

        if ( s->factors )
        s->factors   = ABC_REALLOC(double, s->factors,  s->cap);
        s->orderpos  = ABC_REALLOC(int,    s->orderpos, s->cap);
        s->reasons   = ABC_REALLOC(int,    s->reasons,  s->cap);
        s->trail     = ABC_REALLOC(lit,    s->trail,    s->cap);
        s->model     = ABC_REALLOC(int,    s->model,    s->cap);
        memset( s->wlists + 2*old_cap, 0, 2*(s->cap-old_cap)*sizeof(veci) );
    } 

    for (var = s->size; var < n; var++){
        assert(!s->wlists[2*var].size);
        assert(!s->wlists[2*var+1].size);
        if ( s->wlists[2*var].ptr == NULL )
            veci_new(&s->wlists[2*var]);
        if ( s->wlists[2*var+1].ptr == NULL )
            veci_new(&s->wlists[2*var+1]);

        if ( s->VarActType == 0 )
            s->activity[var] = (1<<10);
        else if ( s->VarActType == 1 )
            s->activity[var] = 0;
        else if ( s->VarActType == 2 )
            s->activity[var] = 0;
        else assert(0);

        s->pFreqs[var]   = 0;
        if ( s->factors )
        s->factors [var] = 0;
//        *((int*)s->vi + var) = 0; s->vi[var].val = varX;
        s->levels  [var] = 0;
        s->assigns [var] = varX;
        s->polarity[var] = 0;
        s->tags    [var] = 0;
        s->loads   [var] = 0;
        s->orderpos[var] = veci_size(&s->order);
        s->reasons [var] = 0;
        s->model   [var] = 0; 
        
        /* does not hold because variables enqueued at top level will not be reinserted in the heap
           assert(veci_size(&s->order) == var); 
         */
        veci_push(&s->order,var);
        order_update(s, var);
    }

    s->size = n > s->size ? n : s->size;
}

void sat_solver3_delete(sat_solver3* s)
{
//    Vec_SetFree_( &s->Mem );
    Sat_MemFree_( &s->Mem );

    // delete vectors
    veci_delete(&s->order);
    veci_delete(&s->trail_lim);
    veci_delete(&s->tagged);
//    veci_delete(&s->learned);
    veci_delete(&s->act_clas);
    veci_delete(&s->stack);
//    veci_delete(&s->model);
    veci_delete(&s->act_vars);
    veci_delete(&s->unit_lits);
    veci_delete(&s->pivot_vars);
    veci_delete(&s->temp_clause);
    veci_delete(&s->conf_final);

    // delete arrays
    if (s->reasons != 0){
        int i;
        for (i = 0; i < s->cap*2; i++)
            veci_delete(&s->wlists[i]);
        ABC_FREE(s->wlists   );
//        ABC_FREE(s->vi       );
        ABC_FREE(s->levels   );
        ABC_FREE(s->assigns  );
        ABC_FREE(s->polarity );
        ABC_FREE(s->tags     );
        ABC_FREE(s->loads    );
        ABC_FREE(s->activity );
        ABC_FREE(s->activity2);
        ABC_FREE(s->pFreqs   );
        ABC_FREE(s->factors  );
        ABC_FREE(s->orderpos );
        ABC_FREE(s->reasons  );
        ABC_FREE(s->trail    );
        ABC_FREE(s->model    );
    }

    ABC_FREE(s);
}

void sat_solver3_restart( sat_solver3* s )
{
    int i;
    Sat_MemRestart( &s->Mem );
    s->hLearnts = -1;
    s->hBinary = Sat_MemAppend( &s->Mem, NULL, 2, 0, 0 );
    s->binary = clause_read( s, s->hBinary );

    veci_resize(&s->trail_lim, 0);
    veci_resize(&s->order, 0);
    for ( i = 0; i < s->size*2; i++ )
        s->wlists[i].size = 0;

    s->nDBreduces = 0;

    // initialize other vars
    s->size                   = 0;
//    s->cap                    = 0;
    s->qhead                  = 0;
    s->qtail                  = 0;


    // variable activities
    solver_init_activities(s);
    veci_resize(&s->act_clas, 0);


    s->root_level             = 0;
//    s->simpdb_assigns         = 0;
//    s->simpdb_props           = 0;
    s->random_seed            = 91648253;
    s->progress_estimate      = 0;
    s->verbosity              = 0;

    s->stats.starts           = 0;
    s->stats.decisions        = 0;
    s->stats.propagations     = 0;
    s->stats.inspects         = 0;
    s->stats.conflicts        = 0;
    s->stats.clauses          = 0;
    s->stats.clauses_literals = 0;
    s->stats.learnts          = 0;
    s->stats.learnts_literals = 0;
    s->stats.tot_literals     = 0;
}

void zsat_solver3_restart_seed( sat_solver3* s, double seed )
{
    int i;
    Sat_MemRestart( &s->Mem );
    s->hLearnts = -1;
    s->hBinary = Sat_MemAppend( &s->Mem, NULL, 2, 0, 0 );
    s->binary = clause_read( s, s->hBinary );

    veci_resize(&s->trail_lim, 0);
    veci_resize(&s->order, 0);
    for ( i = 0; i < s->size*2; i++ )
        s->wlists[i].size = 0;

    s->nDBreduces = 0;

    // initialize other vars
    s->size                   = 0;
//    s->cap                    = 0;
    s->qhead                  = 0;
    s->qtail                  = 0;

    solver_init_activities(s);
    veci_resize(&s->act_clas, 0);

    s->root_level             = 0;
//    s->simpdb_assigns         = 0;
//    s->simpdb_props           = 0;
    s->random_seed            = seed;
    s->progress_estimate      = 0;
    s->verbosity              = 0;

    s->stats.starts           = 0;
    s->stats.decisions        = 0;
    s->stats.propagations     = 0;
    s->stats.inspects         = 0;
    s->stats.conflicts        = 0;
    s->stats.clauses          = 0;
    s->stats.clauses_literals = 0;
    s->stats.learnts          = 0;
    s->stats.learnts_literals = 0;
    s->stats.tot_literals     = 0;
}

// returns memory in bytes used by the SAT solver
double sat_solver3_memory( sat_solver3* s )
{
    int i;
    double Mem = sizeof(sat_solver3);
    for (i = 0; i < s->cap*2; i++)
        Mem += s->wlists[i].cap * sizeof(int);
    Mem += s->cap * sizeof(veci);     // ABC_FREE(s->wlists   );
    Mem += s->cap * sizeof(int);      // ABC_FREE(s->levels   );
    Mem += s->cap * sizeof(char);     // ABC_FREE(s->assigns  );
    Mem += s->cap * sizeof(char);     // ABC_FREE(s->polarity );
    Mem += s->cap * sizeof(char);     // ABC_FREE(s->tags     );
    Mem += s->cap * sizeof(char);     // ABC_FREE(s->loads    );
    Mem += s->cap * sizeof(word);     // ABC_FREE(s->activity );
    if ( s->activity2 )
    Mem += s->cap * sizeof(word);     // ABC_FREE(s->activity );
    if ( s->factors )
    Mem += s->cap * sizeof(double);   // ABC_FREE(s->factors  );
    Mem += s->cap * sizeof(int);      // ABC_FREE(s->orderpos );
    Mem += s->cap * sizeof(int);      // ABC_FREE(s->reasons  );
    Mem += s->cap * sizeof(lit);      // ABC_FREE(s->trail    );
    Mem += s->cap * sizeof(int);      // ABC_FREE(s->model    );

    Mem += s->order.cap * sizeof(int);
    Mem += s->trail_lim.cap * sizeof(int);
    Mem += s->tagged.cap * sizeof(int);
//    Mem += s->learned.cap * sizeof(int);
    Mem += s->stack.cap * sizeof(int);
    Mem += s->act_vars.cap * sizeof(int);
    Mem += s->unit_lits.cap * sizeof(int);
    Mem += s->act_clas.cap * sizeof(int);
    Mem += s->temp_clause.cap * sizeof(int);
    Mem += s->conf_final.cap * sizeof(int);
    Mem += Sat_MemMemoryAll( &s->Mem );
    return Mem;
}

int sat_solver3_simplify(sat_solver3* s)
{
    assert(sat_solver3_dl(s) == 0);
    if (sat_solver3_propagate(s) != 0)
        return false;
    return true;
}

void sat_solver3_reducedb(sat_solver3* s)
{
    static abctime TimeTotal = 0;
    abctime clk = Abc_Clock();
    Sat_Mem_t * pMem = &s->Mem;
    int nLearnedOld = veci_size(&s->act_clas);
    int * act_clas = veci_begin(&s->act_clas);
    int * pPerm, * pArray, * pSortValues, nCutoffValue;
    int i, k, j, Id, Counter, CounterStart, nSelected;
    clause * c;

    assert( s->nLearntMax > 0 );
    assert( nLearnedOld == Sat_MemEntryNum(pMem, 1) );
    assert( nLearnedOld == (int)s->stats.learnts );

    s->nDBreduces++;

    //printf( "Calling reduceDB with %d learned clause limit.\n", s->nLearntMax );
    s->nLearntMax = s->nLearntStart + s->nLearntDelta * s->nDBreduces;
//    return;

    // create sorting values
    pSortValues = ABC_ALLOC( int, nLearnedOld );
    Sat_MemForEachLearned( pMem, c, i, k )
    {
        Id = clause_id(c);
//        pSortValues[Id] = act[Id];
        if ( s->ClaActType == 0 )
            pSortValues[Id] = ((7 - Abc_MinInt(c->lbd, 7)) << 28) | (act_clas[Id] >> 4);
        else
            pSortValues[Id] = ((7 - Abc_MinInt(c->lbd, 7)) << 28);// | (act_clas[Id] >> 4);
        assert( pSortValues[Id] >= 0 );
    }

    // preserve 1/20 of last clauses
    CounterStart  = nLearnedOld - (s->nLearntMax / 20);

    // preserve 3/4 of most active clauses
    nSelected = nLearnedOld*s->nLearntRatio/100;

    // find non-decreasing permutation
    pPerm = Abc_MergeSortCost( pSortValues, nLearnedOld );
    assert( pSortValues[pPerm[0]] <= pSortValues[pPerm[nLearnedOld-1]] );
    nCutoffValue = pSortValues[pPerm[nLearnedOld-nSelected]];
    ABC_FREE( pPerm );
//    ActCutOff = ABC_INFINITY;

    // mark learned clauses to remove
    Counter = j = 0;
    Sat_MemForEachLearned( pMem, c, i, k )
    {
        assert( c->mark == 0 );
        if ( Counter++ > CounterStart || clause_size(c) < 3 || pSortValues[clause_id(c)] > nCutoffValue || s->reasons[lit_var(c->lits[0])] == Sat_MemHand(pMem, i, k) )
            act_clas[j++] = act_clas[clause_id(c)];
        else // delete
        {
            c->mark = 1;
            s->stats.learnts_literals -= clause_size(c);
            s->stats.learnts--;
        }
    }
    assert( s->stats.learnts == (unsigned)j );
    assert( Counter == nLearnedOld );
    veci_resize(&s->act_clas,j);
    ABC_FREE( pSortValues );

    // update ID of each clause to be its new handle
    Counter = Sat_MemCompactLearned( pMem, 0 );
    assert( Counter == (int)s->stats.learnts );

    // update reasons
    for ( i = 0; i < s->size; i++ )
    {
        if ( !s->reasons[i] ) // no reason
            continue;
        if ( clause_is_lit(s->reasons[i]) ) // 2-lit clause
            continue;
        if ( !clause_learnt_h(pMem, s->reasons[i]) ) // problem clause
            continue;
        c = clause_read( s, s->reasons[i] );
        assert( c->mark == 0 );
        s->reasons[i] = clause_id(c); // updating handle here!!!
    }

    // update watches
    for ( i = 0; i < s->size*2; i++ )
    {
        pArray = veci_begin(&s->wlists[i]);
        for ( j = k = 0; k < veci_size(&s->wlists[i]); k++ )
        {
            if ( clause_is_lit(pArray[k]) ) // 2-lit clause
                pArray[j++] = pArray[k];
            else if ( !clause_learnt_h(pMem, pArray[k]) ) // problem clause
                pArray[j++] = pArray[k];
            else 
            {
                c = clause_read(s, pArray[k]);
                if ( !c->mark ) // useful learned clause
                   pArray[j++] = clause_id(c); // updating handle here!!!
            }
        }
        veci_resize(&s->wlists[i],j);
    }

    // perform final move of the clauses
    Counter = Sat_MemCompactLearned( pMem, 1 );
    assert( Counter == (int)s->stats.learnts );

    // report the results
    TimeTotal += Abc_Clock() - clk;
    if ( s->fVerbose )
    {
    Abc_Print(1, "reduceDB: Keeping %7d out of %7d clauses (%5.2f %%)  ",
        s->stats.learnts, nLearnedOld, 100.0 * s->stats.learnts / nLearnedOld );
    Abc_PrintTime( 1, "Time", TimeTotal );
    }
}


// reverses to the previously bookmarked point
void sat_solver3_rollback( sat_solver3* s )
{
    Sat_Mem_t * pMem = &s->Mem;
    int i, k, j;
    static int Count = 0;
    Count++;
    assert( s->iVarPivot >= 0 && s->iVarPivot <= s->size );
    assert( s->iTrailPivot >= 0 && s->iTrailPivot <= s->qtail );
    // reset implication queue
    sat_solver3_canceluntil_rollback( s, s->iTrailPivot );
    // update order 
    if ( s->iVarPivot < s->size )
    { 
        if ( s->activity2 )
        {
            s->var_inc = s->var_inc2;
            memcpy( s->activity, s->activity2, sizeof(word) * s->iVarPivot );
        }
        veci_resize(&s->order, 0);
        for ( i = 0; i < s->iVarPivot; i++ )
        {
            if ( var_value(s, i) != varX )
                continue;
            s->orderpos[i] = veci_size(&s->order);
            veci_push(&s->order,i);
            order_update(s, i);
        }
    }
    // compact watches
    for ( i = 0; i < s->iVarPivot*2; i++ )
    {
        cla* pArray = veci_begin(&s->wlists[i]);
        for ( j = k = 0; k < veci_size(&s->wlists[i]); k++ )
        {
            if ( clause_is_lit(pArray[k]) )
            {
                if ( clause_read_lit(pArray[k]) < s->iVarPivot*2 )
                    pArray[j++] = pArray[k];
            }
            else if ( Sat_MemClauseUsed(pMem, pArray[k]) )
                pArray[j++] = pArray[k];
        }
        veci_resize(&s->wlists[i],j);
    }
    // reset watcher lists
    for ( i = 2*s->iVarPivot; i < 2*s->size; i++ )
        s->wlists[i].size = 0;

    // reset clause counts
    s->stats.clauses = pMem->BookMarkE[0];
    s->stats.learnts = pMem->BookMarkE[1];
    // rollback clauses
    Sat_MemRollBack( pMem );

    // resize learned arrays
    veci_resize(&s->act_clas,  s->stats.learnts);

    // initialize other vars
    s->size = s->iVarPivot;
    if ( s->size == 0 )
    {
    //    s->size                   = 0;
    //    s->cap                    = 0;
        s->qhead                  = 0;
        s->qtail                  = 0;

        solver_init_activities(s);

        s->root_level             = 0;
        s->random_seed            = 91648253;
        s->progress_estimate      = 0;
        s->verbosity              = 0;

        s->stats.starts           = 0;
        s->stats.decisions        = 0;
        s->stats.propagations     = 0;
        s->stats.inspects         = 0;
        s->stats.conflicts        = 0;
        s->stats.clauses          = 0;
        s->stats.clauses_literals = 0;
        s->stats.learnts          = 0;
        s->stats.learnts_literals = 0;
        s->stats.tot_literals     = 0;

        // initialize rollback
        s->iVarPivot              =  0; // the pivot for variables
        s->iTrailPivot            =  0; // the pivot for trail
        s->hProofPivot            =  1; // the pivot for proof records
    }
}


int sat_solver3_addclause(sat_solver3* s, lit* begin, lit* end)
{
    int fVerbose = 0;
    lit *i,*j;
    int maxvar;
    lit last;
    assert( begin < end );
    if ( fVerbose )
    {
        for ( i = begin; i < end; i++ )
            printf( "%s%d ", (*i)&1 ? "!":"", (*i)>>1 );
        printf( "\n" );
    }

    veci_resize( &s->temp_clause, 0 );
    for ( i = begin; i < end; i++ )
        veci_push( &s->temp_clause, *i );
    begin = veci_begin( &s->temp_clause );
    end = begin + veci_size( &s->temp_clause );

    // insertion sort
    maxvar = lit_var(*begin);
    for (i = begin + 1; i < end; i++){
        lit l = *i;
        maxvar = lit_var(l) > maxvar ? lit_var(l) : maxvar;
        for (j = i; j > begin && *(j-1) > l; j--)
            *j = *(j-1);
        *j = l;
    }
    sat_solver3_setnvars(s,maxvar+1);

    // delete duplicates
    last = lit_Undef;
    for (i = j = begin; i < end; i++){
        //printf("lit: "L_LIT", value = %d\n", L_lit(*i), (lit_sign(*i) ? -s->assignss[lit_var(*i)] : s->assignss[lit_var(*i)]));
        if (*i == lit_neg(last) || var_value(s, lit_var(*i)) == lit_sign(*i))
            return true;   // tautology
        else if (*i != last && var_value(s, lit_var(*i)) == varX)
            last = *j++ = *i;
    }
//    j = i;

    if (j == begin)          // empty clause
        return false;

    if (j - begin == 1) // unit clause
        return sat_solver3_enqueue(s,*begin,0);

    // create new clause
    sat_solver3_clause_new(s,begin,j,0);
    return true;
}

static double luby(double y, int x)
{
    int size, seq;
    for (size = 1, seq = 0; size < x+1; seq++, size = 2*size + 1);
    while (size-1 != x){
        size = (size-1) >> 1;
        seq--;
        x = x % size;
    }
    return pow(y, (double)seq);
} 

static void luby_test()
{
    int i;
    for ( i = 0; i < 20; i++ )
        printf( "%d ", (int)luby(2,i) );
    printf( "\n" );
}

static lbool sat_solver3_search(sat_solver3* s, ABC_INT64_T nof_conflicts)
{
//    double  var_decay       = 0.95;
//    double  clause_decay    = 0.999;
    double  random_var_freq = s->fNotUseRandom ? 0.0 : 0.02;
    ABC_INT64_T  conflictC  = 0;
    veci    learnt_clause;
    int     i;

    assert(s->root_level == sat_solver3_dl(s));

    s->nRestarts++;
    s->stats.starts++;
//    s->var_decay = (float)(1 / var_decay   );  // move this to sat_solver3_new()
//    s->cla_decay = (float)(1 / clause_decay);  // move this to sat_solver3_new()
//    veci_resize(&s->model,0);
    veci_new(&learnt_clause);

    // use activity factors in every even restart
    if ( (s->nRestarts & 1) && veci_size(&s->act_vars) > 0 )
//    if ( veci_size(&s->act_vars) > 0 )
        for ( i = 0; i < s->act_vars.size; i++ )
            act_var_bump_factor(s, s->act_vars.ptr[i]);

    // use activity factors in every restart
    if ( s->pGlobalVars && veci_size(&s->act_vars) > 0 )
        for ( i = 0; i < s->act_vars.size; i++ )
            act_var_bump_global(s, s->act_vars.ptr[i]);

    for (;;){
        int hConfl = sat_solver3_propagate(s);
        if (hConfl != 0){
            // CONFLICT
            int blevel;

#ifdef VERBOSEDEBUG
            printf(L_IND"**CONFLICT**\n", L_ind);
#endif
            s->stats.conflicts++; conflictC++;
            if (sat_solver3_dl(s) == s->root_level){
                            
                sat_solver3_analyze_final(s, hConfl, 0);
      
                veci_delete(&learnt_clause);
                return l_False;
            }

            veci_resize(&learnt_clause,0);
            sat_solver3_analyze(s, hConfl, &learnt_clause);
            blevel = veci_size(&learnt_clause) > 1 ? var_level(s, lit_var(veci_begin(&learnt_clause)[1])) : s->root_level;
            blevel = s->root_level > blevel ? s->root_level : blevel;
            sat_solver3_canceluntil(s,blevel);
            sat_solver3_record(s,&learnt_clause);
                            
//            if (learnt_clause.size() == 1) level[var(learnt_clause[0])] = 0;    // (this is ugly (but needed for 'analyzeFinal()') -- in future versions, we will backtrack past the 'root_level' and redo the assumptions)
            if ( learnt_clause.size == 1 ) 
                var_set_level(s, lit_var(learnt_clause.ptr[0]), 0);
      
            act_var_decay(s);
            act_clause_decay(s);

        }else{
            // NO CONFLICT
            int next;

            // Reached bound on number of conflicts:
            if ( (!s->fNoRestarts && nof_conflicts >= 0 && conflictC >= nof_conflicts) || (s->nRuntimeLimit && (s->stats.conflicts & 63) == 0 && Abc_Clock() > s->nRuntimeLimit)){
                s->progress_estimate = sat_solver3_progress(s);
                sat_solver3_canceluntil(s,s->root_level);
                veci_delete(&learnt_clause);
                return l_Undef; }

            // Reached bound on number of conflicts:
            if ( (s->nConfLimit && s->stats.conflicts > s->nConfLimit) ||
                 (s->nInsLimit  && s->stats.propagations > s->nInsLimit) )
            {
                s->progress_estimate = sat_solver3_progress(s);
                sat_solver3_canceluntil(s,s->root_level);
                veci_delete(&learnt_clause);
                return l_Undef; 
            }

            // Simplify the set of problem clauses:
            if (sat_solver3_dl(s) == 0 && !s->fSkipSimplify)
                sat_solver3_simplify(s);

            // Reduce the set of learnt clauses:
//            if (s->nLearntMax && veci_size(&s->learned) - s->qtail >= s->nLearntMax)
            if (s->nLearntMax && veci_size(&s->act_clas) >= s->nLearntMax)
                sat_solver3_reducedb(s);

            // New variable decision:
            s->stats.decisions++;
            next = order_select(s,(float)random_var_freq);

            if (next == var_Undef){
                // Model found:
                int i;
                for (i = 0; i < s->size; i++)
                    s->model[i] = (var_value(s,i)==var1 ? l_True : l_False);
                sat_solver3_canceluntil(s,s->root_level);
                veci_delete(&learnt_clause);

                /*
                veci apa; veci_new(&apa);
                for (i = 0; i < s->size; i++) 
                    veci_push(&apa,(int)(s->model.ptr[i] == l_True ? toLit(i) : lit_neg(toLit(i))));
                printf("model: "); printlits((lit*)apa.ptr, (lit*)apa.ptr + veci_size(&apa)); printf("\n");
                veci_delete(&apa);
                */

                return l_True;
            }

            if ( var_polar(s, next) ) // positive polarity
                sat_solver3_decision(s,toLit(next));
            else
                sat_solver3_decision(s,lit_neg(toLit(next)));
        }
    }

    return l_Undef; // cannot happen
}

// internal call to the SAT solver
int sat_solver3_solve_internal(sat_solver3* s)
{
    lbool status = l_Undef;
    int restart_iter = 0;
    veci_resize(&s->unit_lits, 0);
    s->nCalls++;

    if (s->verbosity >= 1){
        printf("==================================[MINISAT]===================================\n");
        printf("| Conflicts |     ORIGINAL     |              LEARNT              | Progress |\n");
        printf("|           | Clauses Literals |   Limit Clauses Literals  Lit/Cl |          |\n");
        printf("==============================================================================\n");
    }

    while (status == l_Undef){
        ABC_INT64_T nof_conflicts;
        double Ratio = (s->stats.learnts == 0)? 0.0 :
            s->stats.learnts_literals / (double)s->stats.learnts;
        if ( s->nRuntimeLimit && Abc_Clock() > s->nRuntimeLimit )
            break;
        if (s->verbosity >= 1)
        {
            printf("| %9.0f | %7.0f %8.0f | %7.0f %7.0f %8.0f %7.1f | %6.3f %% |\n", 
                (double)s->stats.conflicts,
                (double)s->stats.clauses, 
                (double)s->stats.clauses_literals,
                (double)0, 
                (double)s->stats.learnts, 
                (double)s->stats.learnts_literals,
                Ratio,
                s->progress_estimate*100);
            fflush(stdout);
        }
        nof_conflicts = (ABC_INT64_T)( 100 * luby(2, restart_iter++) );
        status = sat_solver3_search(s, nof_conflicts);
        // quit the loop if reached an external limit
        if ( s->nConfLimit && s->stats.conflicts > s->nConfLimit )
            break;
        if ( s->nInsLimit  && s->stats.propagations > s->nInsLimit )
            break;
        if ( s->nRuntimeLimit && Abc_Clock() > s->nRuntimeLimit )
            break;
    }
    if (s->verbosity >= 1)
        printf("==============================================================================\n");

    sat_solver3_canceluntil(s,s->root_level);
    return status;
}

// pushing one assumption to the stack of assumptions
int sat_solver3_push(sat_solver3* s, int p)
{
    assert(lit_var(p) < s->size);
    veci_push(&s->trail_lim,s->qtail);
    s->root_level++;
    if (!sat_solver3_enqueue(s,p,0))
    {
        int h = s->reasons[lit_var(p)];
        if (h)
        {
            if (clause_is_lit(h))
            {
                (clause_begin(s->binary))[1] = lit_neg(p);
                (clause_begin(s->binary))[0] = clause_read_lit(h);
                h = s->hBinary;
            }
            sat_solver3_analyze_final(s, h, 1);
            veci_push(&s->conf_final, lit_neg(p));
        }
        else
        {
            veci_resize(&s->conf_final,0);
            veci_push(&s->conf_final, lit_neg(p));
            // the two lines below are a bug fix by Siert Wieringa 
            if (var_level(s, lit_var(p)) > 0)
                veci_push(&s->conf_final, p);
        }
        //sat_solver3_canceluntil(s, 0);
        return false; 
    }
    else
    {
        int fConfl = sat_solver3_propagate(s);
        if (fConfl){
            sat_solver3_analyze_final(s, fConfl, 0);
            //assert(s->conf_final.size > 0);
            //sat_solver3_canceluntil(s, 0);
            return false; }
    }
    return true;
}

// removing one assumption from the stack of assumptions
void sat_solver3_pop(sat_solver3* s)
{
    assert( sat_solver3_dl(s) > 0 );
    sat_solver3_canceluntil(s, --s->root_level);
}

void sat_solver3_set_resource_limits(sat_solver3* s, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, ABC_INT64_T nConfLimitGlobal, ABC_INT64_T nInsLimitGlobal)
{
    // set the external limits
    s->nRestarts  = 0;
    s->nConfLimit = 0;
    s->nInsLimit  = 0;
    if ( nConfLimit )
        s->nConfLimit = s->stats.conflicts + nConfLimit;
    if ( nInsLimit )
//        s->nInsLimit = s->stats.inspects + nInsLimit;
        s->nInsLimit = s->stats.propagations + nInsLimit;
    if ( nConfLimitGlobal && (s->nConfLimit == 0 || s->nConfLimit > nConfLimitGlobal) )
        s->nConfLimit = nConfLimitGlobal;
    if ( nInsLimitGlobal && (s->nInsLimit == 0 || s->nInsLimit > nInsLimitGlobal) )
        s->nInsLimit = nInsLimitGlobal;
}

int sat_solver3_solve(sat_solver3* s, lit* begin, lit* end, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, ABC_INT64_T nConfLimitGlobal, ABC_INT64_T nInsLimitGlobal)
{
    lbool status;
    lit * i;
    if ( s->fSolved )
        return l_False;

    if ( s->fVerbose )
        printf( "Running SAT solver with parameters %d and %d and %d.\n", s->nLearntStart, s->nLearntDelta, s->nLearntRatio );

    sat_solver3_set_resource_limits( s, nConfLimit, nInsLimit, nConfLimitGlobal, nInsLimitGlobal );

                            
    // Perform assumptions:
    s->root_level = 0;
    for ( i = begin; i < end; i++ )
        if ( !sat_solver3_push(s, *i) )
        {
            sat_solver3_canceluntil(s,0);
            s->root_level = 0;
            return l_False;
        }
    assert(s->root_level == sat_solver3_dl(s));
     
                                                             
                                  
                                                                                       
                                   
                              
                  
                             
                                        
                                              
                      
                          
                              
                                          
                           
         
     
                                      
      

    status = sat_solver3_solve_internal(s);

    sat_solver3_canceluntil(s,0);
    s->root_level = 0;
    return status;
}

// This LEXSAT procedure should be called with a set of literals (pLits, nLits),
// which defines both (1) variable order, and (2) assignment to begin search from.
// It retuns the LEXSAT assigment that is the same or larger than the given one.
// (It assumes that there is no smaller assignment than the one given!)
// The resulting assignment is returned in the same set of literals (pLits, nLits).
// It pushes/pops assumptions internally and will undo them before terminating.
int sat_solver3_solve_lexsat( sat_solver3* s, int * pLits, int nLits )
{
    int i, iLitFail = -1;
    lbool status;
    assert( nLits > 0 );
    // help the SAT solver by setting desirable polarity
    sat_solver3_set_literal_polarity( s, pLits, nLits );
    // check if there exists a satisfying assignment
    status = sat_solver3_solve_internal( s );
    if ( status != l_True ) // no assignment
        return status;
    // there is at least one satisfying assignment
    assert( status == l_True );
    // find the first mismatching literal
    for ( i = 0; i < nLits; i++ )
        if ( pLits[i] != sat_solver3_var_literal(s, Abc_Lit2Var(pLits[i])) )
            break;
    if ( i == nLits ) // no mismatch - the current assignment is the minimum one!
        return l_True;
    // mismatch happens in literal i
    iLitFail = i;
    // create assumptions up to this literal (as in pLits) - including this literal!
    for ( i = 0; i <= iLitFail; i++ )
        if ( !sat_solver3_push(s, pLits[i]) ) // can become UNSAT while adding the last assumption
            break;
    if ( i < iLitFail + 1 ) // the solver became UNSAT while adding assumptions
        status = l_False;
    else // solve under the assumptions
        status = sat_solver3_solve_internal( s );
    if ( status == l_True )
    {
        // we proved that there is a sat assignment with literal (iLitFail) having polarity as in pLits
        // continue solving recursively
        if ( iLitFail + 1 < nLits )
            status = sat_solver3_solve_lexsat( s, pLits + iLitFail + 1, nLits - iLitFail - 1 );
    }
    else if ( status == l_False )
    {
        // we proved that there is no assignment with iLitFail having polarity as in pLits
        assert( Abc_LitIsCompl(pLits[iLitFail]) ); // literal is 0 
        // (this assert may fail only if there is a sat assignment smaller than one originally given in pLits)
        // now we flip this literal (make it 1), change the last assumption
        // and contiue looking for the 000...0-assignment of other literals
        sat_solver3_pop( s );
        pLits[iLitFail] = Abc_LitNot(pLits[iLitFail]);
        if ( !sat_solver3_push(s, pLits[iLitFail]) )
            printf( "sat_solver3_solve_lexsat(): A satisfying assignment should exist.\n" ); // because we know that the problem is satisfiable
        // update other literals to be 000...0
        for ( i = iLitFail + 1; i < nLits; i++ )
            pLits[i] = Abc_LitNot( Abc_LitRegular(pLits[i]) );
        // continue solving recursively
        if ( iLitFail + 1 < nLits )
            status = sat_solver3_solve_lexsat( s, pLits + iLitFail + 1, nLits - iLitFail - 1 );
        else
            status = l_True;
    }
    // undo the assumptions
    for ( i = iLitFail; i >= 0; i-- )
        sat_solver3_pop( s );
    return status;
}

// This procedure is called on a set of assumptions to minimize their number.
// The procedure relies on the fact that the current set of assumptions is UNSAT.  
// It receives and returns SAT solver without assumptions. It returns the number 
// of assumptions after minimization. The set of assumptions is returned in pLits.
int sat_solver3_minimize_assumptions( sat_solver3* s, int * pLits, int nLits, int nConfLimit )
{
    int i, k, nLitsL, nLitsR, nResL, nResR;
    if ( nLits == 1 )
    {
        // since the problem is UNSAT, we will try to solve it without assuming the last literal
        // if the result is UNSAT, the last literal can be dropped; otherwise, it is needed
        int status = l_False;
        int Temp = s->nConfLimit; 
        s->nConfLimit = nConfLimit;
        status = sat_solver3_solve_internal( s );
        s->nConfLimit = Temp;
        return (int)(status != l_False); // return 1 if the problem is not UNSAT
    }
    assert( nLits >= 2 );
    nLitsL = nLits / 2;
    nLitsR = nLits - nLitsL;
    // assume the left lits
    for ( i = 0; i < nLitsL; i++ )
        if ( !sat_solver3_push(s, pLits[i]) )
        {
            for ( k = i; k >= 0; k-- )
                sat_solver3_pop(s);
            return sat_solver3_minimize_assumptions( s, pLits, i+1, nConfLimit );
        }
    // solve for the right lits
    nResL = sat_solver3_minimize_assumptions( s, pLits + nLitsL, nLitsR, nConfLimit );
    for ( i = 0; i < nLitsL; i++ )
        sat_solver3_pop(s);
    // swap literals
//    assert( nResL <= nLitsL );
//    for ( i = 0; i < nResL; i++ )
//        ABC_SWAP( int, pLits[i], pLits[nLitsL+i] );
    veci_resize( &s->temp_clause, 0 );
    for ( i = 0; i < nLitsL; i++ )
        veci_push( &s->temp_clause, pLits[i] );
    for ( i = 0; i < nResL; i++ )
        pLits[i] = pLits[nLitsL+i];
    for ( i = 0; i < nLitsL; i++ )
        pLits[nResL+i] = veci_begin(&s->temp_clause)[i];
    // assume the right lits
    for ( i = 0; i < nResL; i++ )
        if ( !sat_solver3_push(s, pLits[i]) )
        {
            for ( k = i; k >= 0; k-- )
                sat_solver3_pop(s);
            return sat_solver3_minimize_assumptions( s, pLits, i+1, nConfLimit );
        }
    // solve for the left lits
    nResR = sat_solver3_minimize_assumptions( s, pLits + nResL, nLitsL, nConfLimit );
    for ( i = 0; i < nResL; i++ )
        sat_solver3_pop(s);
    return nResL + nResR;
}

// This is a specialized version of the above procedure with several custom changes:
// - makes sure that at least one of the marked literals is preserved in the clause
// - sets literals to zero when they do not have to be used
// - sets literals to zero for disproved variables
int sat_solver3_minimize_assumptions2( sat_solver3* s, int * pLits, int nLits, int nConfLimit )
{
    int i, k, nLitsL, nLitsR, nResL, nResR;
    if ( nLits == 1 )
    {
        // since the problem is UNSAT, we will try to solve it without assuming the last literal
        // if the result is UNSAT, the last literal can be dropped; otherwise, it is needed
        int RetValue = 1, LitNot = Abc_LitNot(pLits[0]);
        int status = l_False;
        int Temp = s->nConfLimit; 
        s->nConfLimit = nConfLimit;

        RetValue = sat_solver3_push( s, LitNot ); assert( RetValue );
        status = sat_solver3_solve_internal( s );
        sat_solver3_pop( s );

        // if the problem is UNSAT, add clause
        if ( status == l_False )
        {
            RetValue = sat_solver3_addclause( s, &LitNot, &LitNot+1 );
            assert( RetValue );
        }

        s->nConfLimit = Temp;
        return (int)(status != l_False); // return 1 if the problem is not UNSAT
    }
    assert( nLits >= 2 );
    nLitsL = nLits / 2;
    nLitsR = nLits - nLitsL;
    // assume the left lits
    for ( i = 0; i < nLitsL; i++ )
        if ( !sat_solver3_push(s, pLits[i]) )
        {
            for ( k = i; k >= 0; k-- )
                sat_solver3_pop(s);

            // add clauses for these literal
            for ( k = i+1; k > nLitsL; k++ )
            {
                int LitNot = Abc_LitNot(pLits[i]);
                int RetValue = sat_solver3_addclause( s, &LitNot, &LitNot+1 );
                assert( RetValue );
            }

            return sat_solver3_minimize_assumptions2( s, pLits, i+1, nConfLimit );
        }
    // solve for the right lits
    nResL = sat_solver3_minimize_assumptions2( s, pLits + nLitsL, nLitsR, nConfLimit );
    for ( i = 0; i < nLitsL; i++ )
        sat_solver3_pop(s);
    // swap literals
//    assert( nResL <= nLitsL );
    veci_resize( &s->temp_clause, 0 );
    for ( i = 0; i < nLitsL; i++ )
        veci_push( &s->temp_clause, pLits[i] );
    for ( i = 0; i < nResL; i++ )
        pLits[i] = pLits[nLitsL+i];
    for ( i = 0; i < nLitsL; i++ )
        pLits[nResL+i] = veci_begin(&s->temp_clause)[i];
    // assume the right lits
    for ( i = 0; i < nResL; i++ )
        if ( !sat_solver3_push(s, pLits[i]) )
        {
            for ( k = i; k >= 0; k-- )
                sat_solver3_pop(s);

            // add clauses for these literal
            for ( k = i+1; k > nResL; k++ )
            {
                int LitNot = Abc_LitNot(pLits[i]);
                int RetValue = sat_solver3_addclause( s, &LitNot, &LitNot+1 );
                assert( RetValue );
            }

            return sat_solver3_minimize_assumptions2( s, pLits, i+1, nConfLimit );
        }
    // solve for the left lits
    nResR = sat_solver3_minimize_assumptions2( s, pLits + nResL, nLitsL, nConfLimit );
    for ( i = 0; i < nResL; i++ )
        sat_solver3_pop(s);
    return nResL + nResR;
}



int sat_solver3_nvars(sat_solver3* s)
{
    return s->size;
}


int sat_solver3_nclauses(sat_solver3* s)
{
    return s->stats.clauses;
}


int sat_solver3_nconflicts(sat_solver3* s)
{
    return (int)s->stats.conflicts;
}



ABC_NAMESPACE_IMPL_END


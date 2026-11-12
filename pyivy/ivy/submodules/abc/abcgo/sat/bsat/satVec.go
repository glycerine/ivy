package bsat

import (
	"fmt"
	"os"
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




//#include "misc/util/abc_global.h"

//ABC_NAMESPACE_HEADER_START


// vector of 32-bit intergers (added for 64-bit portability)
type veci_t struct {
    cap    int32;
    size    int32;
    ptr   *int32;
}
type veci = veci_t;

func veci_new(v *veci) {
    v.cap  = 4;
    v.size = 0;
    // inlining of macro ABC_ALLOC
    v.ptr  = (int*)((char *) malloc(sizeof(char) * (size_t)(sizeof(int)*v.cap)));
}

// inlining of macro ABC_FREE
func veci_delete(v *veci) { ((v.ptr) ? (free((char *) (v.ptr)), (v.ptr) = 0) : 0);   }
func veci_begin(v *veci) *int32 { return v.ptr;  }
func veci_size(v *veci) int32 { return v.size; }
func veci_resize(v *veci, k int32) { 
    assert(k <= v.size, "k <= v.size"); 
//    memset( veci_begin(v) + k, -1, sizeof(int) * (veci_size(v) - k) );
    v.size = k;         
} // only safe to shrink !!
func veci_pop(v *veci) int32 { assert(v.size, "v.size"); return v.ptr[--v.size]; }
func veci_push(v *veci, e int32) {
    if v.size == v.cap {
//        int newsize = v->cap * 2;//+1;
        var newsize int32;
        if v.cap < 4 {
            newsize = v.cap * 2;
        } else {
            newsize = (v.cap / 2) * 3;
        }
        // inlining of macro ABC_REALLOC
        
        if v.ptr {
            v.ptr = ((int *) realloc((char *)(v.ptr), sizeof(int) * (size_t)(newsize)));
        } else {
            v.ptr = ((int *) malloc(sizeof(int) * (size_t)(newsize)));
        }
        if v.ptr == nil {
            fmt.Printf("Failed to realloc memory from %.1f MB to %.1f MB.\n", 1.0 * v.cap / (1<<20), 1.0 * newsize / (1<<20));
            fflush( os.Stdout );
        }
        v.cap = newsize; }
    v.ptr[v.size++] = e;
}
func veci_remove(v *veci, e int32) {
    var ws *int32  = (int*)veci_begin(v);
    var j int32;
    for ; ws[j] != e  ; j++ {}
    assert(j < veci_size(v), "j < veci_size(v)");
    for ; j < veci_size(v)-1; j++ { ws[j] = ws[j+1]; }
    veci_resize(v,veci_size(v)-1);
}


// vector of 32- or 64-bit pointers
type vecp_t struct {
    cap    int32;
    size    int32;
    ptr *any;
}
type vecp = vecp_t;

func vecp_new(v *vecp) {
    v.size = 0;
    v.cap  = 4;
    // inlining of macro ABC_ALLOC
    v.ptr  = (void**)((char *) malloc(sizeof(char) * (size_t)(sizeof(void*)*v.cap)));
}

// inlining of macro ABC_FREE
func vecp_delete(v *vecp) { ((v.ptr) ? (free((char *) (v.ptr)), (v.ptr) = 0) : 0);   }
func vecp_begin(v *vecp) *any { return v.ptr;  }
func vecp_size(v *vecp) int32 { return v.size; }
func vecp_resize(v *vecp, k int32) { assert(k <= v.size, "k <= v.size"); v.size = k;    } // only safe to shrink !!
func vecp_push(v *vecp, e any) {
    if v.size == v.cap {
//        int newsize = v->cap * 2;//+1;
        var newsize int32;
        if v.cap < 4 {
            newsize = v.cap * 2;
        } else {
            newsize = (v.cap / 2) * 3;
        }
        // inlining of macro ABC_REALLOC
        
        if v.ptr {
            v.ptr = ((void* *) realloc((char *)(v.ptr), sizeof(void*) * (size_t)(newsize)));
        } else {
            v.ptr = ((void* *) malloc(sizeof(void*) * (size_t)(newsize)));
        }
        v.cap = newsize; }
    v.ptr[v.size++] = e;
}
func vecp_remove(v *vecp, e any) {
    var ws *any  = vecp_begin(v);
    var j int32;
    for ; ws[j] != e  ; j++ {}
    assert(j < vecp_size(v), "j < vecp_size(v)");
    for ; j < vecp_size(v)-1; j++ { ws[j] = ws[j+1]; }
    vecp_resize(v,vecp_size(v)-1);
}



//=================================================================================================
// Simple types:

#ifndef __cplusplus
#ifndef false
#  define false 0
#endif
#ifndef true
#  define true 1
#endif
#endif

typedef int    lit;
typedef int    cla;

typedef char               lbool;

var var_Undef int32 = -1;
var lit_Undef lit = -2;

var l_Undef lbool;
var l_True lbool = 1;
var l_False lbool = -1;

func toLit(v int32) lit { return v + v; }
func toLitCond(v int32, c int32) lit { return v + v + (c != 0); }
func lit_neg(l lit) lit { return l ^ 1; }
func lit_var(l lit) int32 { return l >> 1; }
func lit_sign(l lit) int32 { return l & 1; }
func lit_print(l lit) int32 { if lit_sign(l) != 0 {
    return -lit_var(l)-1;
} else {
    return lit_var(l)+1;
} }
func lit_read(s int32) lit { if s > 0 {
    return toLit(s-1);
} else {
    return lit_neg(toLit(-s-1));
} }
func lit_check(l lit, n int32) int32 { return l >= 0 && lit_var(l) < n;                  }

type stats_t struct {
    starts, clauses, learnts uint32;
    ABC_INT64_T decisions, propagations, inspects, conflicts;
    ABC_INT64_T clauses_literals, learnts_literals, tot_literals;
}
   

//ABC_NAMESPACE_HEADER_END



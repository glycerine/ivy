package vec

import (
	"fmt"
	"os"
)

//===--- vec_int.h ----------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===



//#include <assert.h>
//#include <stdio.h>
//#include <string.h>

//#include "../mem.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START

   
type vec_int_t struct {
    cap uint32;
    size uint32;
    data *int32;
}

//===------------------------------------------------------------------------===
// Vector Macros
//===------------------------------------------------------------------------===
// Original macro:
// #define vec_int_foreach(vec, entry, i) \
//     for (i = 0; (i < vec->size) && (((entry) = vec_int_at(vec, i)), 1); i++)


// Original macro:
// #define vec_int_foreach_start(vec, entry, i, start) \
//     for (i = start; (i < vec_int_size(vec)) && (((entry) = vec_int_at(vec, i)), 1); i++)


// Original macro:
// #define vec_int_foreach_stop(vec, entry, i, stop) \
//     for (i = 0; (i < stop) && (((entry) = vec_int_at(vec, i)), 1); i++)


//===------------------------------------------------------------------------===
// Vector API
//===------------------------------------------------------------------------===
func vec_int_alloc(cap uint32) *vec_int_t {
    // inlining of macro satoko_alloc
    var p *vec_int_t  = ((vec_int_t *) malloc((1) * sizeof(vec_int_t)));

    if cap > 0 && cap < 16 {
        cap = 16; }
    p.size = 0;
    p.cap = cap;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((int *) malloc((p.cap) * sizeof(int)));
    } else {
        p.data = nil;
    }
    return p;
}

func vec_int_alloc_exact(cap uint32) *vec_int_t {
    // inlining of macro satoko_alloc
    var p *vec_int_t  = ((vec_int_t *) malloc((1) * sizeof(vec_int_t)));

    p.size = 0;
    p.cap = cap;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((int *) malloc((p.cap) * sizeof(int)));
    } else {
        p.data = nil;
    }
    return p;
}

func vec_int_init(size uint32, value int32) *vec_int_t {
    // inlining of macro satoko_alloc
    var p *vec_int_t  = ((vec_int_t *) malloc((1) * sizeof(vec_int_t)));

    p.cap = size;
    p.size = size;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((int *) malloc((p.cap) * sizeof(int)));
    } else {
        p.data = nil;
    }
    memset(p.data, value, sizeof(int) * p.size);
    return p;
}

func vec_int_free(p *vec_int_t) {
    if p.data != nil {
        // inlining of macro satoko_free
        for { free(p.data); p.data = nil; 
if !(false) { break }
}  }
    // inlining of macro satoko_free
    for { free(p); p = nil; 
if !(false) { break }
} 
}

func vec_int_size(p *vec_int_t) uint32 {
    return p.size;
}

func vec_int_resize(p *vec_int_t, new_size uint32) {
    p.size = new_size;
    if p.cap >= new_size {
        return; }
    // inlining of macro satoko_realloc
    p.data = ((int *) realloc(p.data, (new_size) * sizeof(int)));
    assert(p.data != nil, "p.data != nil");
    p.cap = new_size;
}

func vec_int_reserve(p *vec_int_t, new_cap uint32) {
    if p.cap >= new_cap {
        return; }
    // inlining of macro satoko_realloc
    p.data = ((int *) realloc(p.data, (new_cap) * sizeof(int)));
    assert(p.data != nil, "p.data != nil");
    p.cap = new_cap;
}

func vec_int_capacity(p *vec_int_t) uint32 {
    return p.cap;
}

func vec_int_empty(p *vec_int_t) int32 {
    if p.size {
        return 0;
    } else {
        return 1;
    }
}

func vec_int_erase(p *vec_int_t) {
    // inlining of macro satoko_free
    for { free(p.data); p.data = nil; 
if !(false) { break }
} 
    p.size = 0;
    p.cap = 0;
}

func vec_int_at(p *vec_int_t, i uint32) int32 {
    assert(i >= 0 && i < p.size, "i >= 0 && i < p.size");
    return p.data[i];
}

func vec_int_at_ptr(p *vec_int_t, i uint32) *int32 {
    assert(i >= 0 && i < p.size, "i >= 0 && i < p.size");
    return p.data + i;
}

func vec_int_duplicate(dest *vec_int_t, src *vec_int_t) {
    assert(dest != nil && src != nil, "dest != nil && src != nil");
    vec_int_resize(dest, src.cap);
    memcpy(dest.data, src.data, sizeof(int) * src.cap);
    dest.size = src.size;
}

func vec_int_copy(dest *vec_int_t, src *vec_int_t) {
    assert(dest != nil && src != nil, "dest != nil && src != nil");
    vec_int_resize(dest, src.size);
    memcpy(dest.data, src.data, sizeof(int) * src.size);
    dest.size = src.size;
}

func vec_int_push_back(p *vec_int_t, value int32) {
    if p.size == p.cap {
        if p.cap < 16 {
            vec_int_reserve(p, 16);
        } else {
            vec_int_reserve(p, 2 * p.cap); }
    }
    p.data[p.size] = value;
    p.size++;
}

func vec_int_assign(p *vec_int_t, i uint32, value int32) {
    assert((i >= 0) && (i < vec_int_size(p)), "(i >= 0) && (i < vec_int_size(p))");
    p.data[i] = value;
}

func vec_int_insert(p *vec_int_t, i uint32, value int32) {
    assert((i >= 0) && (i < vec_int_size(p)), "(i >= 0) && (i < vec_int_size(p))");
    vec_int_push_back(p, 0);
    memmove(p.data + i + 1, p.data + i, (p.size - i - 2) * sizeof(int));
    p.data[i] = value;
}

func vec_int_drop(p *vec_int_t, i uint32) {
    assert((i >= 0) && (i < vec_int_size(p)), "(i >= 0) && (i < vec_int_size(p))");
    memmove(p.data + i, p.data + i + 1, (p.size - i - 1) * sizeof(int));
    p.size -= 1;
}

func vec_int_clear(p *vec_int_t) {
    p.size = 0;
}

func vec_int_asc_compare(p1 any, p2 any) int32 {
    int *pp1 = (int *) p1;
    int *pp2 = (int *) p2;

    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 {
        return 1; }
    return 0;
}

func vec_int_desc_compare(p1 any, p2 any) int32 {
    int *pp1 = (int *) p1;
    int *pp2 = (int *) p2;

    if *pp1 > *pp2 {
        return -1; }
    if *pp1 < *pp2 {
        return 1; }
    return 0;
}

func vec_int_sort(p *vec_int_t, ascending int32) {
    if ascending != 0 {
        qsort((void *) p.data, (size_t)p.size, sizeof(int),
                      (int (*)(void*, void*)) vec_int_asc_compare);
    } else {
        qsort((void *) p.data, (size_t)p.size, sizeof(int),
                      (int (*)(void*, void*)) vec_int_desc_compare); }
}

func vec_int_memory(p *vec_int_t) long {
    if p == nil {
        return 0;
    } else {
        return sizeof(int) * p.cap + sizeof(vec_int_t);
    }
}

func vec_int_print(p *vec_int_t) {
    var i uint32;
    assert(p != nil, "p != nil");
    fmt.Fprintf(os.Stdout, "Vector has %d(%d) entries: {", p.size, p.cap);
    for i = 0; i < p.size; i++ {
        fmt.Fprintf(os.Stdout, " %d", p.data[i]);
    }
    fmt.Fprintf(os.Stdout, " }\n");
}

//ABC_NAMESPACE_HEADER_END


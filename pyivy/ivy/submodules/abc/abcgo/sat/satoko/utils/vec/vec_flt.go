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

   
type vec_flt_t struct {
    cap uint32;
    size uint32;
    data *float32;
}

//===------------------------------------------------------------------------===
// Vector Macros
//===------------------------------------------------------------------------===
// Original macro:
// #define vec_flt_foreach(vec, entry, i) \
//     for (i = 0; (i < vec->size) && (((entry) = vec_flt_at(vec, i)), 1); i++)


// Original macro:
// #define vec_flt_foreach_start(vec, entry, i, start) \
//     for (i = start; (i < vec_flt_size(vec)) && (((entry) = vec_flt_at(vec, i)), 1); i++)


// Original macro:
// #define vec_flt_foreach_stop(vec, entry, i, stop) \
//     for (i = 0; (i < stop) && (((entry) = vec_flt_at(vec, i)), 1); i++)


//===------------------------------------------------------------------------===
// Vector API
//===------------------------------------------------------------------------===
func vec_flt_alloc(cap uint32) *vec_flt_t {
    // inlining of macro satoko_alloc
    var p *vec_flt_t  = ((vec_flt_t *) malloc((1) * sizeof(vec_flt_t)));

    if cap > 0 && cap < 16 {
        cap = 16; }
    p.size = 0;
    p.cap = cap;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((float *) malloc((p.cap) * sizeof(float)));
    } else {
        p.data = nil;
    }
    return p;
}

func vec_flt_alloc_exact(cap uint32) *vec_flt_t {
    // inlining of macro satoko_alloc
    var p *vec_flt_t  = ((vec_flt_t *) malloc((1) * sizeof(vec_flt_t)));

    p.size = 0;
    p.cap = cap;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((float *) malloc((p.cap) * sizeof(float)));
    } else {
        p.data = nil;
    }
    return p;
}

func vec_flt_init(size uint32, value float32) *vec_flt_t {
    // inlining of macro satoko_alloc
    var p *vec_flt_t  = ((vec_flt_t *) malloc((1) * sizeof(vec_flt_t)));

    p.cap = size;
    p.size = size;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((float *) malloc((p.cap) * sizeof(float)));
    } else {
        p.data = nil;
    }
    memset(p.data, value, sizeof(float) * p.size);
    return p;
}

func vec_flt_free(p *vec_flt_t) {
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

func vec_flt_size(p *vec_flt_t) uint32 {
    return p.size;
}

func vec_flt_resize(p *vec_flt_t, new_size uint32) {
    p.size = new_size;
    if p.cap >= new_size {
        return; }
    // inlining of macro satoko_realloc
    p.data = ((float *) realloc(p.data, (new_size) * sizeof(float)));
    assert(p.data != nil, "p.data != nil");
    p.cap = new_size;
}

func vec_flt_reserve(p *vec_flt_t, new_cap uint32) {
    if p.cap >= new_cap {
        return; }
    // inlining of macro satoko_realloc
    p.data = ((float *) realloc(p.data, (new_cap) * sizeof(float)));
    assert(p.data != nil, "p.data != nil");
    p.cap = new_cap;
}

func vec_flt_capacity(p *vec_flt_t) uint32 {
    return p.cap;
}

func vec_flt_empty(p *vec_flt_t) int32 {
    if p.size {
        return 0;
    } else {
        return 1;
    }
}

func vec_flt_erase(p *vec_flt_t) {
    // inlining of macro satoko_free
    for { free(p.data); p.data = nil; 
if !(false) { break }
} 
    p.size = 0;
    p.cap = 0;
}

func vec_flt_at(p *vec_flt_t, i uint32) float32 {
    assert(i >= 0 && i < p.size, "i >= 0 && i < p.size");
    return p.data[i];
}

func vec_flt_at_ptr(p *vec_flt_t, i uint32) *float32 {
    assert(i >= 0 && i < p.size, "i >= 0 && i < p.size");
    return p.data + i;
}

func vec_flt_data(p *vec_flt_t) *float32 {
    assert((p != nil), "p");
    return p.data;
}

func vec_flt_duplicate(dest *vec_flt_t, src *vec_flt_t) {
    assert(dest != nil && src != nil, "dest != nil && src != nil");
    vec_flt_resize(dest, src.cap);
    memcpy(dest.data, src.data, sizeof(float) * src.cap);
    dest.size = src.size;
}

func vec_flt_copy(dest *vec_flt_t, src *vec_flt_t) {
    assert(dest != nil && src != nil, "dest != nil && src != nil");
    vec_flt_resize(dest, src.size);
    memcpy(dest.data, src.data, sizeof(float) * src.size);
    dest.size = src.size;
}

func vec_flt_push_back(p *vec_flt_t, value float32) {
    if p.size == p.cap {
        if p.cap < 16 {
            vec_flt_reserve(p, 16);
        } else {
            vec_flt_reserve(p, 2 * p.cap); }
    }
    p.data[p.size] = value;
    p.size++;
}

func vec_flt_assign(p *vec_flt_t, i uint32, value float32) {
    assert((i >= 0) && (i < vec_flt_size(p)), "(i >= 0) && (i < vec_flt_size(p))");
    p.data[i] = value;
}

func vec_flt_insert(p *vec_flt_t, i uint32, value float32) {
    assert((i >= 0) && (i < vec_flt_size(p)), "(i >= 0) && (i < vec_flt_size(p))");
    vec_flt_push_back(p, 0);
    memmove(p.data + i + 1, p.data + i, (p.size - i - 2) * sizeof(float));
    p.data[i] = value;
}

func vec_flt_drop(p *vec_flt_t, i uint32) {
    assert((i >= 0) && (i < vec_flt_size(p)), "(i >= 0) && (i < vec_flt_size(p))");
    memmove(p.data + i, p.data + i + 1, (p.size - i - 1) * sizeof(float));
    p.size -= 1;
}

func vec_flt_clear(p *vec_flt_t) {
    p.size = 0;
}

func vec_flt_asc_compare(p1 any, p2 any) int32 {
    float *pp1 = (float *) p1;
    float *pp2 = (float *) p2;

    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 {
        return 1; }
    return 0;
}

func vec_flt_desc_compare(p1 any, p2 any) int32 {
    float *pp1 = (float *) p1;
    float *pp2 = (float *) p2;

    if *pp1 > *pp2 {
        return -1; }
    if *pp1 < *pp2 {
        return 1; }
    return 0;
}

func vec_flt_sort(p *vec_flt_t, ascending int32) {
    if ascending != 0 {
        qsort((void *) p.data, (size_t)p.size, sizeof(float),
                      (int (*)(void*, void*)) vec_flt_asc_compare);
    } else {
        qsort((void *) p.data, (size_t)p.size, sizeof(float),
                      (int (*)(void*, void*)) vec_flt_desc_compare); }
}

func vec_flt_memory(p *vec_flt_t) long {
    if p == nil {
        return 0;
    } else {
        return sizeof(float) * p.cap + sizeof(vec_flt_t);
    }
}

func vec_flt_print(p *vec_flt_t) {
    var i uint32;
    assert(p != nil, "p != nil");
    fmt.Fprintf(os.Stdout, "Vector has %d(%d) entries: {", p.size, p.cap);
    for i = 0; i < p.size; i++ {
        fmt.Fprintf(os.Stdout, " %f", p.data[i]);
    }
    fmt.Fprintf(os.Stdout, " }\n");
}

//ABC_NAMESPACE_HEADER_END


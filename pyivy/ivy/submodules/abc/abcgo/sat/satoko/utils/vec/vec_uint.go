package vec

import (
	"fmt"
	"os"
)

//===--- vec_uint.h ---------------------------------------------------------===
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

   
type vec_uint_t struct {
    cap uint32;
    size uint32;
    data *uint32;
}

//===------------------------------------------------------------------------===
// Vector Macros
//===------------------------------------------------------------------------===
// Original macro:
// #define vec_uint_foreach(vec, entry, i) \
//     for (i = 0; (i < vec_uint_size(vec)) && (((entry) = vec_uint_at(vec, i)), 1); i++)


// Original macro:
// #define vec_uint_foreach_start(vec, entry, i, start) \
//     for (i = start; (i < vec_uint_size(vec)) && (((entry) = vec_uint_at(vec, i)), 1); i++)


// Original macro:
// #define vec_uint_foreach_stop(vec, entry, i, stop) \
//     for (i = 0; (i < stop) && (((entry) = vec_uint_at(vec, i)), 1); i++)


//===------------------------------------------------------------------------===
// Vector API
//===------------------------------------------------------------------------===
func vec_uint_alloc(cap uint32) *vec_uint_t {
    // inlining of macro satoko_alloc
    var p *vec_uint_t  = ((vec_uint_t *) malloc((1) * sizeof(vec_uint_t)));

    if cap > 0 && cap < 16 {
        cap = 16; }
    p.size = 0;
    p.cap = cap;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((unsigned *) malloc((p.cap) * sizeof(unsigned)));
    } else {
        p.data = nil;
    }
    return p;
}

func vec_uint_alloc_exact(cap uint32) *vec_uint_t {
    // inlining of macro satoko_alloc
    var p *vec_uint_t  = ((vec_uint_t *) malloc((1) * sizeof(vec_uint_t)));

    cap = 0;
    p.size = 0;
    p.cap = cap;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((unsigned *) malloc((p.cap) * sizeof(unsigned)));
    } else {
        p.data = nil;
    }
    return p;
}

func vec_uint_init(size uint32, value uint32) *vec_uint_t {
    // inlining of macro satoko_alloc
    var p *vec_uint_t  = ((vec_uint_t *) malloc((1) * sizeof(vec_uint_t)));

    p.cap = size;
    p.size = size;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((unsigned *) malloc((p.cap) * sizeof(unsigned)));
    } else {
        p.data = nil;
    }
    memset(p.data, value, sizeof(unsigned) * p.size);
    return p;
}

func vec_uint_free(p *vec_uint_t) {
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

func vec_uint_size(p *vec_uint_t) uint32 {
    return p.size;
}

func vec_uint_resize(p *vec_uint_t, new_size uint32) {
    p.size = new_size;
    if p.cap >= new_size {
        return; }
    // inlining of macro satoko_realloc
    p.data = ((unsigned *) realloc(p.data, (new_size) * sizeof(unsigned)));
    assert(p.data != nil, "p.data != nil");
    p.cap = new_size;
}

func vec_uint_shrink(p *vec_uint_t, new_size uint32) {
    assert(p.cap >= new_size, "p.cap >= new_size");
    p.size = new_size;
}

func vec_uint_reserve(p *vec_uint_t, new_cap uint32) {
    if p.cap >= new_cap {
        return; }
    // inlining of macro satoko_realloc
    p.data = ((unsigned *) realloc(p.data, (new_cap) * sizeof(unsigned)));
    assert(p.data != nil, "p.data != nil");
    p.cap = new_cap;
}

func vec_uint_capacity(p *vec_uint_t) uint32 {
    return p.cap;
}

func vec_uint_empty(p *vec_uint_t) int32 {
    if p.size {
        return 0;
    } else {
        return 1;
    }
}

func vec_uint_erase(p *vec_uint_t) {
    // inlining of macro satoko_free
    for { free(p.data); p.data = nil; 
if !(false) { break }
} 
    p.size = 0;
    p.cap = 0;
}

func vec_uint_at(p *vec_uint_t, idx uint32) uint32 {
    assert(idx >= 0 && idx < p.size, "idx >= 0 && idx < p.size");
    return p.data[idx];
}

func vec_uint_at_ptr(p *vec_uint_t, idx uint32) *uint32 {
    assert(idx >= 0 && idx < p.size, "idx >= 0 && idx < p.size");
    return p.data + idx;
}

func vec_uint_find(p *vec_uint_t, entry uint32) uint32 {
    var i uint32;
    for i = 0; i < p.size; i++ {
        if p.data[i] == entry {
            return 1; }
    }
    return 0;
}

func vec_uint_data(p *vec_uint_t) *uint32 {
    assert((p != nil), "p");
    return p.data;
}

func vec_uint_duplicate(dest *vec_uint_t, src *vec_uint_t) {
    assert(dest != nil && src != nil, "dest != nil && src != nil");
    vec_uint_resize(dest, src.cap);
    memcpy(dest.data, src.data, sizeof(unsigned) * src.cap);
    dest.size = src.size;
}

func vec_uint_copy(dest *vec_uint_t, src *vec_uint_t) {
    assert(dest != nil && src != nil, "dest != nil && src != nil");
    vec_uint_resize(dest, src.size);
    memcpy(dest.data, src.data, sizeof(unsigned) * src.size);
    dest.size = src.size;
}

func vec_uint_push_back(p *vec_uint_t, value uint32) {
    if p.size == p.cap {
        if p.cap < 16 {
            vec_uint_reserve(p, 16);
        } else {
            vec_uint_reserve(p, 2 * p.cap); }
    }
    p.data[p.size] = value;
    p.size++;
}

func vec_uint_pop_back(p *vec_uint_t) uint32 {
    assert((p != nil) && p.size, "p && p.size");
    return p.data[--p.size];
}

func vec_uint_assign(p *vec_uint_t, idx uint32, value uint32) {
    assert((idx >= 0) && (idx < vec_uint_size(p)), "(idx >= 0) && (idx < vec_uint_size(p))");
    p.data[idx] = value;
}

func vec_uint_insert(p *vec_uint_t, idx uint32, value uint32) {
    assert((idx >= 0) && (idx < vec_uint_size(p)), "(idx >= 0) && (idx < vec_uint_size(p))");
    vec_uint_push_back(p, 0);
    memmove(p.data + idx + 1, p.data + idx, (p.size - idx - 2) * sizeof(unsigned));
    p.data[idx] = value;
}

func vec_uint_drop(p *vec_uint_t, idx uint32) {
    assert((idx >= 0) && (idx < vec_uint_size(p)), "(idx >= 0) && (idx < vec_uint_size(p))");
    memmove(p.data + idx, p.data + idx + 1, (p.size - idx - 1) * sizeof(unsigned));
    p.size -= 1;
}

func vec_uint_clear(p *vec_uint_t) {
    p.size = 0;
}

func vec_uint_asc_compare(p1 any, p2 any) int32 {
    unsigned *pp1 = (unsigned *) p1;
    unsigned *pp2 = (unsigned *) p2;

    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 {
        return 1; }
    return 0;
}

func vec_uint_desc_compare(p1 any, p2 any) int32 {
    unsigned *pp1 = (unsigned *) p1;
    unsigned *pp2 = (unsigned *) p2;

    if *pp1 > *pp2 {
        return -1; }
    if *pp1 < *pp2 {
        return 1; }
    return 0;
}

func vec_uint_sort(p *vec_uint_t, ascending int32) {
    if ascending != 0 {
        qsort((void *) p.data, (size_t)p.size, sizeof(unsigned),
                      (int (*)(void *, void *)) vec_uint_asc_compare);
    } else {
        qsort((void*) p.data, (size_t)p.size, sizeof(unsigned),
                      (int (*)(void *, void *)) vec_uint_desc_compare); }
}

func vec_uint_memory(p *vec_uint_t) long {
    if p == nil {
        return 0;
    } else {
        return sizeof(unsigned) * p.cap + sizeof(vec_uint_t);
    }
}

func vec_uint_print(p *vec_uint_t) {
    var i uint32;
    assert(p != nil, "p != nil");
    fmt.Fprintf(os.Stdout, "Vector has %d(%d) entries: {", p.size, p.cap);
    for i = 0; i < p.size; i++ {
        fmt.Fprintf(os.Stdout, " %d", p.data[i]);
    }
    fmt.Fprintf(os.Stdout, " }\n");
}

//ABC_NAMESPACE_HEADER_END


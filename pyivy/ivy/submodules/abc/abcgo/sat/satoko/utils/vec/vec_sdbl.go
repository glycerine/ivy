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
//#include "../sdbl.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START

   
type vec_sdbl_t struct {
    cap uint32;
    size uint32;
    data *uint64;
}

//===------------------------------------------------------------------------===
// Vector Macros
//===------------------------------------------------------------------------===
// Original macro:
// #define vec_sdbl_foreach(vec, entry, i) \
//     for (i = 0; (i < vec->size) && (((entry) = vec_sdbl_at(vec, i)), 1); i++)


// Original macro:
// #define vec_sdbl_foreach_start(vec, entry, i, start) \
//     for (i = start; (i < vec_sdbl_size(vec)) && (((entry) = vec_sdbl_at(vec, i)), 1); i++)


// Original macro:
// #define vec_sdbl_foreach_stop(vec, entry, i, stop) \
//     for (i = 0; (i < stop) && (((entry) = vec_sdbl_at(vec, i)), 1); i++)


//===------------------------------------------------------------------------===
// Vector API
//===------------------------------------------------------------------------===
func vec_sdbl_alloc(cap uint32) *vec_sdbl_t {
    // inlining of macro satoko_alloc
    var p *vec_sdbl_t  = ((vec_sdbl_t *) malloc((1) * sizeof(vec_sdbl_t)));

    if cap > 0 && cap < 16 {
        cap = 16; }
    p.size = 0;
    p.cap = cap;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((sdbl_t *) malloc((p.cap) * sizeof(sdbl_t)));
    } else {
        p.data = nil;
    }
    return p;
}

func vec_sdbl_alloc_exact(cap uint32) *vec_sdbl_t {
    // inlining of macro satoko_alloc
    var p *vec_sdbl_t  = ((vec_sdbl_t *) malloc((1) * sizeof(vec_sdbl_t)));

    p.size = 0;
    p.cap = cap;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((sdbl_t *) malloc((p.cap) * sizeof(sdbl_t)));
    } else {
        p.data = nil;
    }
    return p;
}

func vec_sdbl_init(size uint32, value uint64) *vec_sdbl_t {
    // inlining of macro satoko_alloc
    var p *vec_sdbl_t  = ((vec_sdbl_t *) malloc((1) * sizeof(vec_sdbl_t)));

    p.cap = size;
    p.size = size;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((sdbl_t *) malloc((p.cap) * sizeof(sdbl_t)));
    } else {
        p.data = nil;
    }
    memset(p.data, value, sizeof(sdbl_t) * p.size);
    return p;
}

func vec_sdbl_free(p *vec_sdbl_t) {
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

func vec_sdbl_size(p *vec_sdbl_t) uint32 {
    return p.size;
}

func vec_sdbl_shrink(p *vec_sdbl_t, new_size uint32) {
    assert(new_size <= p.cap, "new_size <= p.cap");
    p.size = new_size;
}

func vec_sdbl_resize(p *vec_sdbl_t, new_size uint32) {
    p.size = new_size;
    if p.cap >= new_size {
        return; }
    // inlining of macro satoko_realloc
    p.data = ((sdbl_t *) realloc(p.data, (new_size) * sizeof(sdbl_t)));
    assert(p.data != nil, "p.data != nil");
    p.cap = new_size;
}

func vec_sdbl_reserve(p *vec_sdbl_t, new_cap uint32) {
    if p.cap >= new_cap {
        return; }
    // inlining of macro satoko_realloc
    p.data = ((sdbl_t *) realloc(p.data, (new_cap) * sizeof(sdbl_t)));
    assert(p.data != nil, "p.data != nil");
    p.cap = new_cap;
}

func vec_sdbl_capacity(p *vec_sdbl_t) uint32 {
    return p.cap;
}

func vec_sdbl_empty(p *vec_sdbl_t) int32 {
    if p.size {
        return 0;
    } else {
        return 1;
    }
}

func vec_sdbl_erase(p *vec_sdbl_t) {
    // inlining of macro satoko_free
    for { free(p.data); p.data = nil; 
if !(false) { break }
} 
    p.size = 0;
    p.cap = 0;
}

func vec_sdbl_at(p *vec_sdbl_t, i uint32) uint64 {
    assert(i >= 0 && i < p.size, "i >= 0 && i < p.size");
    return p.data[i];
}

func vec_sdbl_at_ptr(p *vec_sdbl_t, i uint32) *uint64 {
    assert(i >= 0 && i < p.size, "i >= 0 && i < p.size");
    return p.data + i;
}

func vec_sdbl_data(p *vec_sdbl_t) *uint64 {
    assert((p != nil), "p");
    return p.data;
}

func vec_sdbl_duplicate(dest *vec_sdbl_t, src *vec_sdbl_t) {
    assert(dest != nil && src != nil, "dest != nil && src != nil");
    vec_sdbl_resize(dest, src.cap);
    memcpy(dest.data, src.data, sizeof(sdbl_t) * src.cap);
    dest.size = src.size;
}

func vec_sdbl_copy(dest *vec_sdbl_t, src *vec_sdbl_t) {
    assert(dest != nil && src != nil, "dest != nil && src != nil");
    vec_sdbl_resize(dest, src.size);
    memcpy(dest.data, src.data, sizeof(sdbl_t) * src.size);
    dest.size = src.size;
}

func vec_sdbl_push_back(p *vec_sdbl_t, value uint64) {
    if p.size == p.cap {
        if p.cap < 16 {
            vec_sdbl_reserve(p, 16);
        } else {
            vec_sdbl_reserve(p, 2 * p.cap); }
    }
    p.data[p.size] = value;
    p.size++;
}

func vec_sdbl_assign(p *vec_sdbl_t, i uint32, value uint64) {
    assert((i >= 0) && (i < vec_sdbl_size(p)), "(i >= 0) && (i < vec_sdbl_size(p))");
    p.data[i] = value;
}

func vec_sdbl_insert(p *vec_sdbl_t, i uint32, value uint64) {
    assert((i >= 0) && (i < vec_sdbl_size(p)), "(i >= 0) && (i < vec_sdbl_size(p))");
    vec_sdbl_push_back(p, 0);
    memmove(p.data + i + 1, p.data + i, (p.size - i - 2) * sizeof(sdbl_t));
    p.data[i] = value;
}

func vec_sdbl_drop(p *vec_sdbl_t, i uint32) {
    assert((i >= 0) && (i < vec_sdbl_size(p)), "(i >= 0) && (i < vec_sdbl_size(p))");
    memmove(p.data + i, p.data + i + 1, (p.size - i - 1) * sizeof(sdbl_t));
    p.size -= 1;
}

func vec_sdbl_clear(p *vec_sdbl_t) {
    p.size = 0;
}

func vec_sdbl_asc_compare(p1 any, p2 any) int32 {
    sdbl_t *pp1 = (sdbl_t *) p1;
    sdbl_t *pp2 = (sdbl_t *) p2;

    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 {
        return 1; }
    return 0;
}

func vec_sdbl_desc_compare(p1 any, p2 any) int32 {
    sdbl_t *pp1 = (sdbl_t *) p1;
    sdbl_t *pp2 = (sdbl_t *) p2;

    if *pp1 > *pp2 {
        return -1; }
    if *pp1 < *pp2 {
        return 1; }
    return 0;
}

func vec_sdbl_sort(p *vec_sdbl_t, ascending int32) {
    if ascending != 0 {
        qsort((void *) p.data, (size_t)p.size, sizeof(sdbl_t),
                      (int (*)(void*, void*)) vec_sdbl_asc_compare);
    } else {
        qsort((void *) p.data, (size_t)p.size, sizeof(sdbl_t),
                      (int (*)(void*, void*)) vec_sdbl_desc_compare); }
}

func vec_sdbl_memory(p *vec_sdbl_t) long {
    if p == nil {
        return 0;
    } else {
        return sizeof(sdbl_t) * p.cap + sizeof(vec_sdbl_t);
    }
}

func vec_sdbl_print(p *vec_sdbl_t) {
    var i uint32;
    assert(p != nil, "p != nil");
    fmt.Fprintf(os.Stdout, "Vector has %d(%d) entries: {", p.size, p.cap);
    for i = 0; i < p.size; i++ {
        fmt.Fprintf(os.Stdout, " %f", sdbl2double(p.data[i]));
    }
    fmt.Fprintf(os.Stdout, " }\n");
}

//ABC_NAMESPACE_HEADER_END


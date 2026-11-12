package vec

import (
	"fmt"
	"os"
)

//===--- vec_char.h ---------------------------------------------------------===
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

   
type vec_char_t struct {
    cap uint32;
    size uint32;
    data string;
}

//===------------------------------------------------------------------------===
// Vector Macros
//===------------------------------------------------------------------------===
// Original macro:
// #define vec_char_foreach(vec, entry, i) \
//     for (i = 0; (i < vec_char_size(vec)) && (((entry) = vec_char_at(vec, i)), 1); i++)


// Original macro:
// #define vec_char_foreach_start(vec, entry, i, start) \
//     for (i = start; (i < vec_char_size(vec)) && (((entry) = vec_char_at(vec, i)), 1); i++)


// Original macro:
// #define vec_char_foreach_stop(vec, entry, i, stop) \
//     for (i = 0; (i < stop) && (((entry) = vec_char_at(vec, i)), 1); i++)


//===------------------------------------------------------------------------===
// Vector API
//===------------------------------------------------------------------------===
func vec_char_alloc(cap uint32) *vec_char_t {
    // inlining of macro satoko_alloc
    var p *vec_char_t  = ((vec_char_t *) malloc((1) * sizeof(vec_char_t)));

    if cap > 0 && cap < 16 {
        cap = 16; }
    p.size = 0;
    p.cap = cap;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((char *) malloc((p.cap) * sizeof(char)));
    } else {
        p.data = nil;
    }
    return p;
}

func vec_char_alloc_exact(cap uint32) *vec_char_t {
    // inlining of macro satoko_alloc
    var p *vec_char_t  = ((vec_char_t *) malloc((1) * sizeof(vec_char_t)));

    cap = 0;
    p.size = 0;
    p.cap = cap;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((char *) malloc((p.cap) * sizeof(char)));
    } else {
        p.data = nil;
    }
    return p;
}

func vec_char_init(size uint32, value byte) *vec_char_t {
    // inlining of macro satoko_alloc
    var p *vec_char_t  = ((vec_char_t *) malloc((1) * sizeof(vec_char_t)));

    p.cap = size;
    p.size = size;
    if p.cap {
        // inlining of macro satoko_alloc
        p.data = ((char *) malloc((p.cap) * sizeof(char)));
    } else {
        p.data = nil;
    }
    memset(p.data, value, sizeof(char) * p.size);
    return p;
}

func vec_char_free(p *vec_char_t) {
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

func vec_char_size(p *vec_char_t) uint32 {
    return p.size;
}

func vec_char_resize(p *vec_char_t, new_size uint32) {
    p.size = new_size;
    if p.cap >= new_size {
        return; }
    // inlining of macro satoko_realloc
    p.data = ((char *) realloc(p.data, (new_size) * sizeof(char)));
    assert(p.data != nil, "p.data != nil");
    p.cap = new_size;
}

func vec_char_shrink(p *vec_char_t, new_size uint32) {
    assert(p.cap > new_size, "p.cap > new_size");
    p.size = new_size;
}

func vec_char_reserve(p *vec_char_t, new_cap uint32) {
    if p.cap >= new_cap {
        return; }
    // inlining of macro satoko_realloc
    p.data = ((char *) realloc(p.data, (new_cap) * sizeof(char)));
    assert(p.data != nil, "p.data != nil");
    p.cap = new_cap;
}

func vec_char_capacity(p *vec_char_t) uint32 {
    return p.cap;
}

func vec_char_empty(p *vec_char_t) int32 {
    if p.size {
        return 0;
    } else {
        return 1;
    }
}

func vec_char_erase(p *vec_char_t) {
    // inlining of macro satoko_free
    for { free(p.data); p.data = nil; 
if !(false) { break }
} 
    p.size = 0;
    p.cap = 0;
}

func vec_char_at(p *vec_char_t, idx uint32) byte {
    assert(idx >= 0 && idx < p.size, "idx >= 0 && idx < p.size");
    return p.data[idx];
}

func vec_char_at_ptr(p *vec_char_t, idx uint32) string {
    assert(idx >= 0 && idx < p.size, "idx >= 0 && idx < p.size");
    return p.data + idx;
}

func vec_char_data(p *vec_char_t) string {
    assert((p != nil), "p");
    return p.data;
}

func vec_char_duplicate(dest *vec_char_t, src *vec_char_t) {
    assert(dest != nil && src != nil, "dest != nil && src != nil");
    vec_char_resize(dest, src.cap);
    memcpy(dest.data, src.data, sizeof(char) * src.cap);
    dest.size = src.size;
}

func vec_char_copy(dest *vec_char_t, src *vec_char_t) {
    assert(dest != nil && src != nil, "dest != nil && src != nil");
    vec_char_resize(dest, src.size);
    memcpy(dest.data, src.data, sizeof(char) * src.size);
    dest.size = src.size;
}

func vec_char_push_back(p *vec_char_t, value byte) {
    if p.size == p.cap {
        if p.cap < 16 {
            vec_char_reserve(p, 16);
        } else {
            vec_char_reserve(p, 2 * p.cap); }
    }
    p.data[p.size] = value;
    p.size++;
}

func vec_char_pop_back(p *vec_char_t) byte {
    assert((p != nil) && p.size, "p && p.size");
    return p.data[--p.size];
}

func vec_char_assign(p *vec_char_t, idx uint32, value byte) {
    assert((idx >= 0) && (idx < vec_char_size(p)), "(idx >= 0) && (idx < vec_char_size(p))");
    p.data[idx] = value;
}

func vec_char_insert(p *vec_char_t, idx uint32, value byte) {
    assert((idx >= 0) && (idx < vec_char_size(p)), "(idx >= 0) && (idx < vec_char_size(p))");
    vec_char_push_back(p, 0);
    memmove(p.data + idx + 1, p.data + idx, (p.size - idx - 2) * sizeof(char));
    p.data[idx] = value;
}

func vec_char_drop(p *vec_char_t, idx uint32) {
    assert((idx >= 0) && (idx < vec_char_size(p)), "(idx >= 0) && (idx < vec_char_size(p))");
    memmove(p.data + idx, p.data + idx + 1, (p.size - idx - 1) * sizeof(char));
    p.size -= 1;
}

func vec_char_clear(p *vec_char_t) {
    p.size = 0;
}

func vec_char_asc_compare(p1 any, p2 any) int32 {
    char *pp1 = (char *)p1;
    char *pp2 = (char *)p2;

    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 {
        return 1; }
    return 0;
}

func vec_char_desc_compare(p1 any, p2 any) int32 {
    char *pp1 = (char *)p1;
    char *pp2 = (char *)p2;

    if *pp1 > *pp2 {
        return -1; }
    if *pp1 < *pp2 {
        return 1; }
    return 0;
}

func vec_char_sort(p *vec_char_t, ascending int32) {
    if ascending != 0 {
        qsort((void *) p.data, (size_t)p.size, sizeof(char),
                      (int (*)(void *, void *)) vec_char_asc_compare);
    } else {
        qsort((void*) p.data, (size_t)p.size, sizeof(char),
                      (int (*)(void *, void *)) vec_char_desc_compare); }
}


func vec_char_memory(p *vec_char_t) long {
    if p == nil {
        return 0;
    } else {
        return sizeof(char) * p.cap + sizeof(vec_char_t);
    }
}

func vec_char_print(p *vec_char_t) {
    var i uint32;
    assert(p != nil, "p != nil");
    fmt.Fprintf(os.Stdout, "Vector has %d(%d) entries: {", p.size, p.cap);
    for i = 0; i < p.size; i++ {
        fmt.Fprintf(os.Stdout, " %d", p.data[i]);
    }
    fmt.Fprintf(os.Stdout, " }\n");
}

//ABC_NAMESPACE_HEADER_END


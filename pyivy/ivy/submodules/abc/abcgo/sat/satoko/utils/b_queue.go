package utils

//===--- b_queue.h ----------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===



//#include "mem.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START

/* Bounded Queue */
   
type b_queue_t struct {
    size uint32;
    cap uint32;
    i_first uint32;
    i_empty uint32;
    unsigned long sum;
    data *uint32;
}

//===------------------------------------------------------------------------===
// Bounded Queue API
//===------------------------------------------------------------------------===
func b_queue_alloc(cap uint32) *b_queue_t {
    // inlining of macro satoko_calloc
    var p *b_queue_t  = &b_queue_t{};
    p.cap = cap;
    // inlining of macro satoko_calloc
    p.data = ((unsigned *) calloc((cap), sizeof(unsigned)));
    return p;
}

func b_queue_free(p *b_queue_t) {
    // inlining of macro satoko_free
    for { free(p.data); p.data = nil; 
if !(false) { break }
} 
    // inlining of macro satoko_free
    for { free(p); p = nil; 
if !(false) { break }
} 
}

func b_queue_push(p *b_queue_t, Value uint32) {
    if p.size == p.cap {
        assert(p.i_first == p.i_empty, "p.i_first == p.i_empty");
        p.sum -= p.data[p.i_first];
        p.i_first = (p.i_first + 1) % p.cap;
    } else {
        p.size++; }

    p.sum += Value;
    p.data[p.i_empty] = Value;
    if (++p.i_empty) == p.cap {
        p.i_empty = 0;
        p.i_first = 0;
    }
}

func b_queue_avg(p *b_queue_t) uint32 {
    return uint32((p.sum / ((unsigned long) p.size)));
}

func b_queue_is_valid(p *b_queue_t) uint32 {
    return (p.cap == p.size);
}

func b_queue_clean(p *b_queue_t) {
    p.i_first = 0;
    p.i_empty = 0;
    p.size = 0;
    p.sum = 0;
}

//ABC_NAMESPACE_HEADER_END


package utils

//===--- heap.h ----------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===



//#include "mem.h"
//#include "../types.h"
//#include "vec/vec_sdbl.h"
//#include "vec/vec_int.h"
//#include "vec/vec_uint.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START

   
type heap_t struct {
    indices *vec_int_t;
    data *vec_uint_t;
    weights *vec_act_t;
}
//===------------------------------------------------------------------------===
// Heap internal functions
//===------------------------------------------------------------------------===
func left(i uint32) uint32 { return 2 * i + 1; }
func right(i uint32) uint32 { return (i + 1) * 2; }
func parent(i uint32) uint32 { return (i - 1) >> 1; }

func compare(p *heap_t, x uint32, y uint32) int32 {
    // inlining of macro vec_act_at
    return vec_sdbl_at(p.weights, x) > vec_sdbl_at(p.weights, y);
}

func heap_percolate_up(h *heap_t, i uint32) {
    var x uint32  = vec_uint_at(h.data, i);
    var p uint32  = parent(i);

    for i != 0 && (compare(h, x, vec_uint_at(h.data, p)) != 0) {
        vec_uint_assign(h.data, i, vec_uint_at(h.data, p));
        vec_int_assign(h.indices, vec_uint_at(h.data, p), int32( i));
        i = p;
        p = parent(p);
    }
    vec_uint_assign(h.data, i, x);
    vec_int_assign(h.indices, x, int32( i));
}

func heap_percolate_down(h *heap_t, i uint32) {
    var x uint32  = vec_uint_at(h.data, i);

    for left(i) < vec_uint_size(h.data) {
        var child uint32;
        if right(i) < vec_uint_size(h.data) &&
                     (compare(h, vec_uint_at(h.data, right(i)), vec_uint_at(h.data, left(i))) != 0) {
            child = right(i);
        } else {
            child = left(i);
        }

        if !(compare(h, vec_uint_at(h.data, child), x) != 0) {
            break; }

        vec_uint_assign(h.data, i, vec_uint_at(h.data, child));
        vec_int_assign(h.indices, vec_uint_at(h.data, i), int32( i));
        i = child;
    }
    vec_uint_assign(h.data, i, x);
    vec_int_assign(h.indices, x, int32( i));
}

//===------------------------------------------------------------------------===
// Heap API
//===------------------------------------------------------------------------===
func heap_alloc(weights *vec_act_t) *heap_t {
    // inlining of macro satoko_alloc
    var p *heap_t  = ((heap_t *) malloc((1) * sizeof(heap_t)));
    p.weights = weights;
    p.indices = vec_int_alloc(0);
    p.data = vec_uint_alloc(0);
    return p;
}

func heap_free(p *heap_t) {
    vec_int_free(p.indices);
    vec_uint_free(p.data);
    // inlining of macro satoko_free
    for { free(p); p = nil; 
if !(false) { break }
} 
}

func heap_size(p *heap_t) uint32 {
    return vec_uint_size(p.data);
}

func heap_in_heap(p *heap_t, entry uint32) int32 {
    return (entry < vec_int_size(p.indices)) &&
           (vec_int_at(p.indices, entry) >= 0);
}

func heap_increase(p *heap_t, entry uint32) {
    assert((heap_in_heap(p, entry) != 0), "heap_in_heap(p, entry)");
    heap_percolate_down(p, uint32( vec_int_at(p.indices, entry)));
}

func heap_decrease(p *heap_t, entry uint32) {
    assert((heap_in_heap(p, entry) != 0), "heap_in_heap(p, entry)");
    heap_percolate_up(p, uint32( vec_int_at(p.indices, entry)));
}

func heap_insert(p *heap_t, entry uint32) {
    if vec_int_size(p.indices) < entry + 1 {
        var old_size uint32  = vec_int_size(p.indices);
        var i uint32;
        var e int32;
        vec_int_resize(p.indices, entry + 1);
        // inlining of macro vec_int_foreach_start
        for i = old_size; (i < vec_int_size(p.indices)) && (((e) = vec_int_at(p.indices, i)), 1); i++ {
            vec_int_assign(p.indices, i, -1);
        }
    }
    assert(!(heap_in_heap(p, entry) != 0), "!heap_in_heap(p, entry)");
    vec_int_assign(p.indices, entry, int32( vec_uint_size(p.data)));
    vec_uint_push_back(p.data, entry);
    heap_percolate_up(p, uint32( vec_int_at(p.indices, entry)));
}

func heap_update(p *heap_t, i uint32) {
    if !(heap_in_heap(p, i) != 0) {
        heap_insert(p, i);
    } else {
        heap_percolate_up(p, uint32( vec_int_at(p.indices, i)));
        heap_percolate_down(p, uint32( vec_int_at(p.indices, i)));
    }
}

func heap_build(p *heap_t, entries *vec_uint_t) {
    var i int32;
    var j, entry uint32;

    // inlining of macro vec_uint_foreach
    for j = 0; (j < vec_uint_size(p.data)) && (((entry) = vec_uint_at(p.data, j)), 1); j++ {
        vec_int_assign(p.indices, entry, -1);
    }
    vec_uint_clear(p.data);
    // inlining of macro vec_uint_foreach
    for j = 0; (j < vec_uint_size(entries)) && (((entry) = vec_uint_at(entries, j)), 1); j++ {
        vec_int_assign(p.indices, entry, int32( j));
        vec_uint_push_back(p.data, entry);
    }
    for (i = vec_uint_size(p.data) / 2 - 1); i >= 0; i-- {
        heap_percolate_down(p, uint32( i));
    }
}

func heap_clear(p *heap_t) {
    vec_int_clear(p.indices);
    vec_uint_clear(p.data);
}

func heap_remove_min(p *heap_t) uint32 {
    var x uint32  = vec_uint_at(p.data, 0);
    vec_uint_assign(p.data, 0, vec_uint_at(p.data, vec_uint_size(p.data) - 1));
    vec_int_assign(p.indices, vec_uint_at(p.data, 0), 0);
    vec_int_assign(p.indices, x, -1);
    vec_uint_pop_back(p.data);
    if vec_uint_size(p.data) > 1 {
        heap_percolate_down(p, 0); }
    return x;
}

//ABC_NAMESPACE_HEADER_END


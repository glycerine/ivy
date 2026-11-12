package satoko

//===--- cdb.h --------------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===



//#include "clause.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START

/* Clauses DB data structure */
type cdb struct {
    size uint32;
    cap uint32;
    wasted uint32;
    data *uint32;
}

//===------------------------------------------------------------------------===
// Clause DB API
//===------------------------------------------------------------------------===
static inline struct clause *cdb_handler(struct cdb *p, unsigned cref)
{
    if cref != 0xFFFFFFFF {
        return (struct clause *)(p.data + cref);
    } else {
        return nil;
    }
}

static inline unsigned cdb_cref(struct cdb *p, unsigned *clause)
{
    return uint32((clause - &(p.data[0])));
}

static inline void cdb_grow(struct cdb *p, unsigned cap)
{
    var prev_cap uint32  = p.cap;

    if p.cap >= cap {
        return; }
    for p.cap < cap {
        var delta uint32  = ((p.cap >> 1) + (p.cap >> 3) + 2) & uint32((~1));
        p.cap += delta;
        assert(p.cap >= prev_cap, "p.cap >= prev_cap");
    }
    assert(p.cap > 0, "p.cap > 0");
    // inlining of macro satoko_realloc
    p.data = ((unsigned *) realloc(p.data, (p.cap) * sizeof(unsigned)));
}

func cdb_alloc(cap uint32) *struct cdb {
    // inlining of macro satoko_calloc
    struct cdb *p = ((struct cdb *) calloc((1), sizeof(struct cdb)));
    if cap <= 0 {
        cap = 1024 * 1024; }
    cdb_grow(p, cap);
    return p;
}

static inline void cdb_free(struct cdb *p)
{
    // inlining of macro satoko_free
    for { free(p.data); p.data = nil; 
if !(false) { break }
} 
    // inlining of macro satoko_free
    for { free(p); p = nil; 
if !(false) { break }
} 
}

static inline unsigned cdb_append(struct cdb *p, unsigned size)
{
    var prev_size uint32;
    assert(size > 0, "size > 0");
    cdb_grow(p, p.size + size);
    prev_size = p.size;
    p.size += size;
    assert(p.size > prev_size, "p.size > prev_size");
    return prev_size;
}

static inline void cdb_remove(struct cdb *p, struct clause *clause)
{
    p.wasted += clause.size;
}

static inline void cdb_clear(struct cdb *p)
{
    p.wasted = 0;
    p.size = 0;
}

static inline unsigned cdb_capacity(struct cdb *p)
{
    return p.cap;
}

static inline unsigned cdb_size(struct cdb *p)
{
    return p.size;
}

static inline unsigned cdb_wasted(struct cdb *p)
{
    return p.wasted;
}

//ABC_NAMESPACE_HEADER_END


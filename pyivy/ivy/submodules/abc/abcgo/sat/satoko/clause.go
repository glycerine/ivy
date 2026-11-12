package satoko

import "fmt"

//===--- clause.h -----------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===



//#include "types.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START

type clause struct {
    f_learnt uint32; /* C bitfield: 1 bits; review packing and truncation */
    f_mark uint32; /* C bitfield: 1 bits; review packing and truncation */
    f_reallocd uint32; /* C bitfield: 1 bits; review packing and truncation */
    f_deletable uint32; /* C bitfield: 1 bits; review packing and truncation */
    lbd uint32; /* C bitfield: 28 bits; review packing and truncation */
    size uint32;
    union {
        unsigned lit;
        clause_act_t act;
    }
 data[0];
}

//===------------------------------------------------------------------------===
// Clause API
//===------------------------------------------------------------------------===
func clause_compare(p1 any, p2 any) int32 {
    struct clause *c1 = (struct clause *)p1;
    struct clause *c2 = (struct clause *)p2;

    if c1.size > 2 && c2.size == 2 {
        return 1; }
    if c1.size == 2 && c2.size > 2 {
        return 0; }
    if c1.size == 2 && c2.size == 2 {
        return 0; }

    if c1.lbd > c2.lbd {
        return 1; }
    if c1.lbd < c2.lbd {
        return 0; }

    return c1.data[c1.size].act < c2.data[c2.size].act;
}

static inline void clause_print(struct clause *clause)
{
    var i uint32;
    fmt.Printf("{ ");
    for i = 0; i < clause.size; i++ {
        fmt.Printf("%d ", clause.data[i].lit);
    }
    fmt.Printf("}\n");
}

static inline void clause_dump(FILE *file, struct clause *clause, int no_zero_var) 
{
    var i uint32;
    for i = 0; i < clause.size; i++ {
        var var int32  = (clause.data[i].lit >> 1);
        var pol byte  = (clause.data[i].lit & 1);
        fmt.Fprintf(file, "%d ", pol ? -(var + no_zero_var) : (var + no_zero_var));
    }
    if no_zero_var {
        fmt.Fprintf(file, "0\n");
    } else {
        fmt.Fprintf(file, "\n"); }
}

//ABC_NAMESPACE_HEADER_END


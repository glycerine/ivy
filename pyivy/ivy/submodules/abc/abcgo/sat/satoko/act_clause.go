package satoko

//===--- act_var.h ----------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===



//#include "solver.h"
//#include "types.h"

//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START

func clause_act_rescale(s *solver_t) {
    var i, cref uint32;
    struct clause *clause;

    // inlining of macro vec_uint_foreach
    for i = 0; (i < vec_uint_size(s.learnts)) && (((cref) = vec_uint_at(s.learnts, i)), 1); i++ {
        clause = clause_fetch(s, cref);
        clause.data[clause.size].act >>= 10;
    }
    s.clause_act_inc = stk_uint_max((s.clause_act_inc >> 10), (1 << 11));
}

static inline void clause_act_bump(solver_t *s, struct clause *clause)
{
    clause.data[clause.size].act += s.clause_act_inc;
    if clause.data[clause.size].act & 0x80000000 {
        clause_act_rescale(s); }
}

func clause_act_decay(s *solver_t) {
    s.clause_act_inc += (s.clause_act_inc >> 10);
}

//ABC_NAMESPACE_HEADER_END


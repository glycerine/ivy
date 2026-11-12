package utils

//===--- misc.h -------------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===



//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START

// Original macro:
// #define stk_swap(type, a, b)  { type t = a; a = b; b = t; }


func stk_uint_max(a uint32, b uint32) uint32 {
    if a > b {
        return a;
    } else {
        return b;
    }
}

func stk_uint_compare(p1 any, p2 any) int32 {
    unsigned pp1 = *(unsigned *)p1;
    unsigned pp2 = *(unsigned *)p2;

    if pp1 < pp2 {
        return -1; }
    if pp1 > pp2 {
        return 1; }
    return 0;
}

//ABC_NAMESPACE_HEADER_END


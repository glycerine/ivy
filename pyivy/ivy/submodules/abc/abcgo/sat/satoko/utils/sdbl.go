package utils

import "fmt"

//===--- sdbl.h -------------------------------------------------------------===
//
//                     satoko: Satisfiability solver
//
// This file is distributed under the BSD 2-Clause License.
// See LICENSE for details.
//
//===------------------------------------------------------------------------===
// by Alan Mishchenko



//#include "misc/util/abc_global.h"
//ABC_NAMESPACE_HEADER_START
/*
    The sdbl_t floating-point number is represented as a 64-bit unsigned int.
    The number is (2^expt)*mnt, where expt is a 16-bit exponent and mnt is a
    48-bit mantissa.  The decimal point is located between the MSB of mnt,
    which is always 1, and the remaining 15 digits of mnt.

    Currently, only positive numbers are represented.

    The range of possible values is [1.0; 2^(2^16-1)*1.111111111111111]
    that is, the smallest possible number is 1.0 and the largest possible
    number is 2^(---16 ones---).(1.---47 ones---)

    Comparison of numbers can be done by comparing the underlying unsigned ints.

    Only addition, multiplication, and division by 2^n are currently implemented.
*/

typedef word sdbl_t;

var SDBL_CONST1 uint64 = ABC_CONST(0x0000800000000000);
var SDBL_MAX uint64 = ~(sdbl_t)(0);

union ui64_dbl { word ui64; double dbl; };

func sdbl_exp(a uint64) word { return a >> 48;         }
func sdbl_mnt(a uint64) word { return (a << 16) >> 16; }

func sdbl2double(a uint64) float64 {
    union ui64_dbl temp;
    assert(sdbl_exp(a) < 1023, "sdbl_exp(a) < 1023");
    temp.ui64 = ((sdbl_exp(a) + 1023) << 52) | (((a << 17) >> 17) << 5);
    return temp.dbl;
}

func double2sdbl(value float64) uint64 {
    union ui64_dbl temp;
    var expt, mnt uint64;
    assert(value >= 1.0, "value >= 1.0");
    temp.dbl = value;
    expt = (temp.ui64 >> 52) - 1023;
    mnt = SDBL_CONST1 | ((temp.ui64 << 12) >> 17);
    return (expt << 48) + mnt;
}

func sdbl_add(a uint64, b uint64) uint64 {
    var expt, mnt uint64;
    if a < b {
        a ^= b;
        b ^= a;
        a ^= b;
    }
    assert(a >= b, "a >= b");
    expt = sdbl_exp(a);
    mnt = sdbl_mnt(a) + (sdbl_mnt(b) >> (sdbl_exp(a) - sdbl_exp(b)));
    /* Check for carry */
    if mnt >> 48 {
        expt++;
        mnt >>= 1;
    }
    if expt >> 16 { /* overflow */
        return SDBL_MAX; }
    return (expt << 48) + mnt;
}

func sdbl_mult(a uint64, b uint64) uint64 {
    var expt, mnt uint64;
    var a_mnt, a_mnt_hi, a_mnt_lo uint64;
    var b_mnt, b_mnt_hi, b_mnt_lo uint64;
    if a < b {
        a ^= b;
        b ^= a;
        a ^= b;
    }
    assert(a >= b, "a >= b");
    a_mnt  = sdbl_mnt(a);
    b_mnt  = sdbl_mnt(b);
    a_mnt_hi = a_mnt>>32;
    b_mnt_hi = b_mnt>>32;
    a_mnt_lo = (a_mnt<<32)>>32;
    b_mnt_lo = (b_mnt<<32)>>32;
    mnt = ((a_mnt_hi * b_mnt_hi) << 17) +
          ((a_mnt_lo * b_mnt_lo) >> 47) +
          ((a_mnt_lo * b_mnt_hi) >> 15) +
          ((a_mnt_hi * b_mnt_lo) >> 15);
    expt = sdbl_exp(a) + sdbl_exp(b);
    /* Check for carry */
    if mnt >> 48 {
        expt++;
        mnt >>= 1;
    }
    if expt >> 16 { /* overflow */
        return SDBL_MAX; }
    return (expt << 48) + mnt;
}

func sdbl_div(a uint64, deg2 uint32) uint64 {
    if sdbl_exp(a) >= (word)deg2 {
        return ((sdbl_exp(a) - deg2) << 48) + sdbl_mnt(a); }
    return SDBL_CONST1;
}

func sdbl_test() {
    var ten100_ uint64  = ABC_CONST(0x014c924d692ca61b);
    fmt.Printf("%f\n", sdbl2double(ten100_));
    //printf("%016lX\n", double2sdbl(1 /0.95));
    //printf("%016lX\n", SDBL_CONST1);
    fmt.Printf("%f\n", sdbl2double(SDBL_CONST1));
    fmt.Printf("%f\n", sdbl2double(ABC_CONST(0x000086BCA1AF286B)));

}

//ABC_NAMESPACE_HEADER_END



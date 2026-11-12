package util

import "fmt"

/**CFile****************************************************************

  FileName    [utilFloat.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName []

  Synopsis    [Floating point number implementation.]

  Author      [Alan Mishchenko, Bruno Schmitt]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - January 28, 2017.]

  Revision    []

***********************************************************************/



//#include "misc/util/abc_global.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/* 
    The xFloat_t floating-point number is represented as a 32-bit unsigned int.
    The number is (2^Exp)*Mnt, where Exp is a 16-bit exponent and Mnt is a 
    16-bit mantissa.  The decimal point is located between the MSB of Mnt,
    which is always 1, and the remaining 15 digits of Mnt.

    Currently, only positive numbers are represented.

    The range of possible values is [1.0; 2^(2^16-1)*1.111111111111111]
    that is, the smallest possible number is 1.0 and the largest possible 
    number is 2^(---16 ones---).(1.---15 ones---)

    Comparison of numbers can be done by comparing the underlying unsigned ints.

    Only addition, multiplication, and division by 2^n are currently implemented.
*/

   
type xFloat_t struct {
    Mnt uint32; /* C bitfield: 16 bits; review packing and truncation */
    Exp uint32; /* C bitfield: 16 bits; review packing and truncation */
}

func xSat_Float2Uint(f xFloat_t) uint32 { union { var f xFloat_t; var u uint32; }
 temp; temp.f = f; return temp.u; }
func xSat_Uint2Float(u uint32) xFloat_t { union { var f xFloat_t; var u uint32; }
 temp; temp.u = u; return temp.f; }
func xSat_LessThan(a xFloat_t, b xFloat_t) int32 { return a.Exp < b.Exp || (a.Exp == b.Exp && a.Mnt < b.Mnt);         }
func xSat_Equal(a xFloat_t, b xFloat_t) int32 { return a.Exp == b.Exp && a.Mnt == b.Mnt;                           }

func xSat_FloatCreate(Exp uint32, Mnt uint32) xFloat_t { var res xFloat_t; res.Exp = Exp; res.Mnt = Mnt; return res;            }

func xSat_FloatCreateConst1() xFloat_t { return xSat_FloatCreate( 0, 1 << 15 );                             }
func xSat_FloatCreateConst2() xFloat_t { return xSat_FloatCreate( 1, 1 << 15 );                             }
func xSat_FloatCreateConst3() xFloat_t { return xSat_FloatCreate( 1, 3 << 14 );                             }
func xSat_FloatCreateConst12() xFloat_t { return xSat_FloatCreate( 3, 3 << 14 );                             }
func xSat_FloatCreateConst1point5() xFloat_t { return xSat_FloatCreate( 0, 3 << 14 );                             }
func xSat_FloatCreateConst2point5() xFloat_t { return xSat_FloatCreate( 1, 5 << 13 );                             }
func xSat_FloatCreateMaximum() xFloat_t { return xSat_Uint2Float( 0xFFFFFFFF );                              }

func xSat_Float2Float(a xFloat_t) float32 { assert(a.Exp < 127, "a.Exp < 127"); return Abc_Int2Float(((a.Exp + 127) << 23) | ((a.Mnt & 0x7FFF) << 8)); }
func xSat_FloatFromFloat(a float32) xFloat_t { var A int32  = Abc_Float2Int(a); assert(a >= 1.0, "a >= 1.0"); return xSat_FloatCreate((A >> 23)-127, 0x8000 | ((A >> 8) & 0x7FFF)); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Adding two floating-point numbers.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSat_FloatAdd(a xFloat_t, b xFloat_t) xFloat_t {
    var Exp, Mnt uint32;
    if a.Exp < b.Exp {
        return xSat_FloatAdd(b, a); }
    assert(a.Exp >= b.Exp, "a.Exp >= b.Exp");
    // compute new mantissa
    Mnt = a.Mnt + (b.Mnt >> (a.Exp - b.Exp));
    // compute new exponent
    Exp = a.Exp;
    // update exponent and mantissa if new MSB is created
    if Mnt & 0xFFFF0000 { // new MSB bit is created
        Exp++, Mnt >>= 1; }
    // check overflow
    if Exp & 0xFFFF0000 { // overflow
        return xSat_Uint2Float( 0xFFFFFFFF ); }
    assert((Exp & 0xFFFF0000) == 0, "(Exp & 0xFFFF0000) == 0");
    assert((Mnt & 0xFFFF0000) == 0, "(Mnt & 0xFFFF0000) == 0");
    assert(Mnt & 0x00008000, "Mnt & 0x00008000");
    return xSat_FloatCreate( Exp, Mnt );
}

/**Function*************************************************************

  Synopsis    [Multiplying two floating-point numbers.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSat_FloatMul(a xFloat_t, b xFloat_t) xFloat_t {
    var Exp, Mnt uint32;
    if a.Exp < b.Exp {
        return xSat_FloatMul(b, a); }
    assert(a.Exp >= b.Exp, "a.Exp >= b.Exp");
    // compute new mantissa
    Mnt = (a.Mnt * b.Mnt) >> 15;
    // compute new exponent
    Exp = a.Exp + b.Exp;
    // update exponent and mantissa if new MSB is created
    if Mnt & 0xFFFF0000 { // new MSB bit is created
        Exp++, Mnt >>= 1; }
    // check overflow
    if Exp & 0xFFFF0000 { // overflow
        return xSat_Uint2Float( 0xFFFFFFFF ); }
    assert((Exp & 0xFFFF0000) == 0, "(Exp & 0xFFFF0000) == 0");
    assert((Mnt & 0xFFFF0000) == 0, "(Mnt & 0xFFFF0000) == 0");
    assert(Mnt & 0x00008000, "Mnt & 0x00008000");
    return xSat_FloatCreate( Exp, Mnt );
}

/**Function*************************************************************

  Synopsis    [Dividing floating point number by a degree of 2.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSat_FloatDiv(a xFloat_t, Deg2 uint32) xFloat_t {
    assert(Deg2 < 0xFFFF, "Deg2 < 0xFFFF");
    if a.Exp >= Deg2 {
        return xSat_FloatCreate( a.Exp - Deg2, a.Mnt ); }
    return xSat_FloatCreateConst1(); // underflow
}

/**Function*************************************************************

  Synopsis    [Testing procedure.]

  Description [Helpful link https://www.h-schmidt.net/FloatConverter/IEEE754.html]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSat_FloatTest() {
    var c1 xFloat_t    = xSat_FloatCreateConst1();
    var c2 xFloat_t    = xSat_FloatCreateConst2();
    var c3 xFloat_t    = xSat_FloatCreateConst3();
    var c12 xFloat_t   = xSat_FloatCreateConst12();
    var c1p5 xFloat_t  = xSat_FloatCreateConst1point5();
    var c2p5 xFloat_t  = xSat_FloatCreateConst2point5();

    var c1_ xFloat_t    = xSat_FloatFromFloat(1.0);
    var c2_ xFloat_t    = xSat_FloatFromFloat(2.0);
    var c3_ xFloat_t    = xSat_FloatFromFloat(3.0);
    var c12_ xFloat_t   = xSat_FloatFromFloat(12.0);
    var c1p5_ xFloat_t  = xSat_FloatFromFloat(1.5);
    var c2p5_ xFloat_t  = xSat_FloatFromFloat(2.5);

    var sum1 xFloat_t  = xSat_FloatAdd(c1, c1p5);
    var mul1 xFloat_t  = xSat_FloatMul(c2, c1p5);

    var sum2 xFloat_t  = xSat_FloatAdd(c1p5, c2p5);
    var mul2 xFloat_t  = xSat_FloatMul(c1p5, c2p5);

//    float f1 = xSat_Float2Float(c1);
//    Extra_PrintBinary( stdout, (int *)&c1, 32 ); printf( "\n" );
//    Extra_PrintBinary( stdout, (int *)&f1, 32 ); printf( "\n" );

    fmt.Printf("1 = %f\n", xSat_Float2Float(c1));
    fmt.Printf("2 = %f\n", xSat_Float2Float(c2));
    fmt.Printf("3 = %f\n", xSat_Float2Float(c3));
    fmt.Printf("12 = %f\n", xSat_Float2Float(c12));
    fmt.Printf("1.5 = %f\n", xSat_Float2Float(c1p5));
    fmt.Printf("2.5 = %f\n", xSat_Float2Float(c2p5));

    fmt.Printf("Converted 1 = %f\n", xSat_Float2Float(c1_));
    fmt.Printf("Converted 2 = %f\n", xSat_Float2Float(c2_));
    fmt.Printf("Converted 3 = %f\n", xSat_Float2Float(c3_));
    fmt.Printf("Converted 12 = %f\n", xSat_Float2Float(c12_));
    fmt.Printf("Converted 1.5 = %f\n", xSat_Float2Float(c1p5_));
    fmt.Printf("Converted 2.5 = %f\n", xSat_Float2Float(c2p5_));

    fmt.Printf("1.0 + 1.5 = %f\n", xSat_Float2Float(sum1));
    fmt.Printf("2.0 * 1.5 = %f\n", xSat_Float2Float(mul1));

    fmt.Printf("1.5 + 2.5 = %f\n", xSat_Float2Float(sum2));
    fmt.Printf("1.5 * 2.5 = %f\n", xSat_Float2Float(mul2));
    fmt.Printf("12 / 2^2  = %f\n", xSat_Float2Float(xSat_FloatDiv(c12, 2)));

    assert((xSat_Equal(sum1, c2p5) != 0), "xSat_Equal(sum1, c2p5)");
    assert((xSat_Equal(mul1, c3) != 0), "xSat_Equal(mul1, c3)");
}

//ABC_NAMESPACE_HEADER_END



package util

import "fmt"

/**CFile****************************************************************

  FileName    [utilDouble.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName []

  Synopsis    [Double floating point number implementation.]

  Author      [Alan Mishchenko, Bruno Schmitt]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - February 11, 2017.]

  Revision    []

***********************************************************************/




//#include "misc/util/abc_global.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/* 
    The xdbl floating-point number is represented as a 64-bit unsigned int.
    The number is (2^Exp)*Mnt, where Exp is a 16-bit exponent and Mnt is a 
    48-bit mantissa.  The decimal point is located between the MSB of Mnt,
    which is always 1, and the remaining 15 digits of Mnt.

    Currently, only positive numbers are represented.

    The range of possible values is [1.0; 2^(2^16-1)*1.111111111111111]
    that is, the smallest possible number is 1.0 and the largest possible 
    number is 2^(---16 ones---).(1.---47 ones---)

    Comparison of numbers can be done by comparing the underlying unsigned ints.

    Only addition, multiplication, and division by 2^n are currently implemented.
*/

typedef word xdbl;

func Xdbl_Exp(a xdbl) word { return a >> 48;           }
func Xdbl_Mnt(a xdbl) word { return (a << 16) >> 16;   }

func Xdbl_Create(Exp word, Mnt word) xdbl { assert(!(Exp>>16) && (Mnt>>47)==(word)1, "!(Exp>>16) && (Mnt>>47)==(word)1"); return (Exp<<48) | Mnt;  }

func Xdbl_Const1() xdbl { return Xdbl_Create( (word)0, (word)1 << 47 );                      }
func Xdbl_Const2() xdbl { return Xdbl_Create( (word)1, (word)1 << 47 );                      }
func Xdbl_Const3() xdbl { return Xdbl_Create( (word)1, (word)3 << 46 );                      }
func Xdbl_Const12() xdbl { return Xdbl_Create( (word)3, (word)3 << 46 );                      }
func Xdbl_Const1point5() xdbl { return Xdbl_Create( (word)0, (word)3 << 46 );                      }
func Xdbl_Const2point5() xdbl { return Xdbl_Create( (word)1, (word)5 << 45 );                      }
func Xdbl_Maximum() xdbl { return ~(word)0;                                                   }

func Xdbl_ToDouble(a xdbl) float64 { assert(Xdbl_Exp(a) < 1023, "Xdbl_Exp(a) < 1023"); return Abc_Word2Dbl(((Xdbl_Exp(a) + 1023) << 52) | (((a<<17)>>17) << 5));                  }
func Xdbl_FromDouble(a float64) xdbl { var A word  = Abc_Dbl2Word(a); assert(a >= 1.0, "a >= 1.0"); return Xdbl_Create((A >> 52)-1023, (((word)1) << 47) | ((A << 12) >> 17)); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Adding two floating-point numbers.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Xdbl_Add(a xdbl, b xdbl) xdbl {
    var Exp, Mnt word;
    if a < b { a ^= b, b ^= a, a ^= b; }
    assert(a >= b, "a >= b");
    Mnt = Xdbl_Mnt(a) + (Xdbl_Mnt(b) >> (Xdbl_Exp(a) - Xdbl_Exp(b)));
    Exp = Xdbl_Exp(a);
    if Mnt >> 48 { // new MSB is created
        Exp++, Mnt >>= 1; }
    if Exp >> 16 { // overflow
        return Xdbl_Maximum(); }
    return Xdbl_Create( Exp, Mnt );
}

/**Function*************************************************************

  Synopsis    [Multiplying two floating-point numbers.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Xdbl_Mul(a xdbl, b xdbl) xdbl {
    var Exp, Mnt, MntA, MntB, MntAh, MntBh, MntAl, MntBl word;
    if a < b { a ^= b, b ^= a, a ^= b; }
    assert(a >= b, "a >= b");
    MntA  = Xdbl_Mnt(a);
    MntB  = Xdbl_Mnt(b);
    MntAh = MntA>>32;
    MntBh = MntB>>32;
    MntAl = (MntA<<32)>>32;
    MntBl = (MntB<<32)>>32;
    Mnt = ((MntAh * MntBh) << 17) + ((MntAl * MntBl) >> 47) + ((MntAl * MntBh) >> 15) + ((MntAh * MntBl) >> 15);
    Exp = Xdbl_Exp(a) + Xdbl_Exp(b);
    if Mnt >> 48 { // new MSB is created
        Exp++, Mnt >>= 1; }
    if Exp >> 16 { // overflow
        return Xdbl_Maximum(); }
    return Xdbl_Create( Exp, Mnt );
}

/**Function*************************************************************

  Synopsis    [Dividing floating point number by a degree of 2.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Xdbl_Div(a xdbl, Deg2 uint32) xdbl {
    if Xdbl_Exp(a) >= (word)Deg2 {
        return Xdbl_Create( Xdbl_Exp(a) - Deg2, Xdbl_Mnt(a) ); }
    return Xdbl_Const1(); // underflow
}

/**Function*************************************************************

  Synopsis    [Testing procedure.]

  Description [Helpful link https://www.h-schmidt.net/FloatConverter/IEEE754.html]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Xdbl_Test() {
    var c1 xdbl    = Xdbl_Const1();
    var c2 xdbl    = Xdbl_Const2();
    var c3 xdbl    = Xdbl_Const3();
    var c12 xdbl   = Xdbl_Const12();
    var c1p5 xdbl  = Xdbl_Const1point5();
    var c2p5 xdbl  = Xdbl_Const2point5();

    var c1_ xdbl    = Xdbl_FromDouble(1.0);
    var c2_ xdbl    = Xdbl_FromDouble(2.0);
    var c3_ xdbl    = Xdbl_FromDouble(3.0);
    var c12_ xdbl   = Xdbl_FromDouble(12.0);
    var c1p5_ xdbl  = Xdbl_FromDouble(1.5);
    var c2p5_ xdbl  = Xdbl_FromDouble(2.5);

    var sum1 xdbl  = Xdbl_Add(c1, c1p5);
    var mul1 xdbl  = Xdbl_Mul(c2, c1p5);

    var sum2 xdbl  = Xdbl_Add(c1p5, c2p5);
    var mul2 xdbl  = Xdbl_Mul(c1p5, c2p5);

    var a xdbl     = Xdbl_FromDouble(1.2929725);
    var b xdbl     = Xdbl_FromDouble(10.28828287);
    var ab xdbl    = Xdbl_Mul(a, b);

    var ten100 xdbl  = Xdbl_FromDouble( 1e100 );
    var ten100_ xdbl  = ABC_CONST(0x014c924d692ca61b);

    assert(ten100 == ten100_, "ten100 == ten100_");

//    float f1 = Xdbl_ToDouble(c1);
//    Extra_PrintBinary( stdout, (int *)&c1, 32 ); printf( "\n" );
//    Extra_PrintBinary( stdout, (int *)&f1, 32 ); printf( "\n" );

    fmt.Printf("1 = %f\n", Xdbl_ToDouble(c1));
    fmt.Printf("2 = %f\n", Xdbl_ToDouble(c2));
    fmt.Printf("3 = %f\n", Xdbl_ToDouble(c3));
    fmt.Printf("12 = %f\n", Xdbl_ToDouble(c12));
    fmt.Printf("1.5 = %f\n", Xdbl_ToDouble(c1p5));
    fmt.Printf("2.5 = %f\n", Xdbl_ToDouble(c2p5));

    fmt.Printf("Converted 1 = %f\n", Xdbl_ToDouble(c1_));
    fmt.Printf("Converted 2 = %f\n", Xdbl_ToDouble(c2_));
    fmt.Printf("Converted 3 = %f\n", Xdbl_ToDouble(c3_));
    fmt.Printf("Converted 12 = %f\n", Xdbl_ToDouble(c12_));
    fmt.Printf("Converted 1.5 = %f\n", Xdbl_ToDouble(c1p5_));
    fmt.Printf("Converted 2.5 = %f\n", Xdbl_ToDouble(c2p5_));

    fmt.Printf("1.0 + 1.5 = %f\n", Xdbl_ToDouble(sum1));
    fmt.Printf("2.0 * 1.5 = %f\n", Xdbl_ToDouble(mul1));

    fmt.Printf("1.5 + 2.5 = %f\n", Xdbl_ToDouble(sum2));
    fmt.Printf("1.5 * 2.5 = %f\n", Xdbl_ToDouble(mul2));
    fmt.Printf("12 / 2^2  = %f\n", Xdbl_ToDouble(Xdbl_Div(c12, 2)));

    fmt.Printf("12 / 2^2  = %f\n", Xdbl_ToDouble(Xdbl_Div(c12, 2)));

    fmt.Printf("%.16f * %.16f = %.16f (%.16f)\n", Xdbl_ToDouble(a), Xdbl_ToDouble(b), Xdbl_ToDouble(ab), 1.2929725 * 10.28828287);

    assert(sum1 == c2p5, "sum1 == c2p5");
    assert(mul1 == c3, "mul1 == c3");
}

//ABC_NAMESPACE_HEADER_END



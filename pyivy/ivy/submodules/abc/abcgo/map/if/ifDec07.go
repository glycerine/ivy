package if_

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [ifDec07.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Performs additional check.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifDec07.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"
//#include "misc/extra/extra.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the bit count for the first 256 integer numbers
var BitCount8 [256]int32 = [256]int32{
    0,1,1,2,1,2,2,3,1,2,2,3,2,3,3,4,1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,
    1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
    1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
    2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
    1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
    2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
    2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
    3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,4,5,5,6,5,6,6,7,5,6,6,7,6,7,7,8,
};
// variable swapping code
var PMasks [5][3]word = [5][3]word{
    { ABC_CONST(0x9999999999999999), ABC_CONST(0x2222222222222222), ABC_CONST(0x4444444444444444) },
    { ABC_CONST(0xC3C3C3C3C3C3C3C3), ABC_CONST(0x0C0C0C0C0C0C0C0C), ABC_CONST(0x3030303030303030) },
    { ABC_CONST(0xF00FF00FF00FF00F), ABC_CONST(0x00F000F000F000F0), ABC_CONST(0x0F000F000F000F00) },
    { ABC_CONST(0xFF0000FFFF0000FF), ABC_CONST(0x0000FF000000FF00), ABC_CONST(0x00FF000000FF0000) },
    { ABC_CONST(0xFFFF00000000FFFF), ABC_CONST(0x00000000FFFF0000), ABC_CONST(0x0000FFFF00000000) },
};
// elementary truth tables
var Truth6 [6]word = [6]word{
    ABC_CONST(0xAAAAAAAAAAAAAAAA),
    ABC_CONST(0xCCCCCCCCCCCCCCCC),
    ABC_CONST(0xF0F0F0F0F0F0F0F0),
    ABC_CONST(0xFF00FF00FF00FF00),
    ABC_CONST(0xFFFF0000FFFF0000),
    ABC_CONST(0xFFFFFFFF00000000),
};
var Truth7 [7][2]word = [7][2]word{
    {ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA)},
    {ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC)},
    {ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0)},
    {ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00)},
    {ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000)},
    {ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000)},
    {ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF)},
};

// extern void Kit_DsdPrintFromTruth( unsigned * pTruth, int nVars );

// extern void Extra_PrintBinary( FILE * pFile, unsigned Sign[], int nBits );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

func If_DecPrintConfig(z word) {
   var S [1]uint32;
   S[0] = (z & 0xffff) | ((z & 0xffff) << 16);
   Extra_PrintBinary( os.Stdout, S, 16 ); 
   fmt.Printf(" ");
   Kit_DsdPrintFromTruth( S, 4 );
   fmt.Printf(" ");
   fmt.Printf(" %d", int32(((z >> 16) & 7)));
   fmt.Printf(" %d", int32(((z >> 20) & 7)));
   fmt.Printf(" %d", int32(((z >> 24) & 7)));
   fmt.Printf(" %d", int32(((z >> 28) & 7)));
   fmt.Printf("   ");
   S[0] = ((z >> 32) & 0xffff) | (((z >> 32) & 0xffff) << 16);
   Extra_PrintBinary( os.Stdout, S, 16 );
   fmt.Printf(" ");
   Kit_DsdPrintFromTruth( S, 4 );
   fmt.Printf(" ");
   fmt.Printf(" %d", int32(((z >> 48) & 7)));
   fmt.Printf(" %d", int32(((z >> 52) & 7)));
   fmt.Printf(" %d", int32(((z >> 56) & 7)));
   fmt.Printf(" %d", int32(((z >> 60) & 7)));
   fmt.Printf("\n");
}

// verification for 6-input function
func If_Dec6ComposeLut4(t int32, f []word) word {
    var m, v int32;
    var c, r word;
    for  m = 0; m < 16; m++  {
        if !((t >> m) & 1) {
            continue; }
        c = ~(word)0;
        for  v = 0; v < 4; v++  {
            c &= ((m >> v) & 1) ? f[v] : ~f[v];
        }
        r |= c;
    }
    return r;
}
func If_Dec6Truth(z word) word {
    var r word;
    var q word;
    var f [4]word;
    var i, v int32;
    assert(z, "z");
    for  i = 0; i < 4; i++  {
        v = (z >> (16+(i<<2))) & 7;
        assert(v != 7, "v != 7");
        if v == 6 {
            continue; }
        f[i] = Truth6[v];
    }
    q = If_Dec6ComposeLut4( int32((z & 0xffff)), f );
    for  i = 0; i < 4; i++  {
        v = (z >> (48+(i<<2))) & 7;
        if v == 6 {
            continue; }
        f[i] = (v == 7) ? q : Truth6[v];
    }
    r = If_Dec6ComposeLut4( int32(((z >> 32) & 0xffff)), f );
    return r;
}
func If_Dec6Verify(t word, z word) {
    var r word  = If_Dec6Truth( z );
    if r != t {
        If_DecPrintConfig( z );
        Kit_DsdPrintFromTruth( (unsigned*)&t, 6 ); fmt.Printf("\n");
//        Kit_DsdPrintFromTruth( (unsigned*)&q, 6 ); printf( "\n" );
        Kit_DsdPrintFromTruth( (unsigned*)&r, 6 ); fmt.Printf("\n");
        fmt.Printf("Verification failed!\n");
    }
}
// verification for 7-input function
func If_Dec7ComposeLut4(t int32, f [][2]word, r []word) {
    var m, v int32;
    var c [2]word;
    r[0] = r[1] = 0;
    for  m = 0; m < 16; m++  {
        if !((t >> m) & 1) {
            continue; }
        c[0] = c[1] = ~(word)0;
        for  v = 0; v < 4; v++  {
            c[0] &= ((m >> v) & 1) ? f[v][0] : ~f[v][0];
            c[1] &= ((m >> v) & 1) ? f[v][1] : ~f[v][1];
        }
        r[0] |= c[0];
        r[1] |= c[1];
    }
}
func If_Dec7Verify(t []word, z word) {
    var f [4][2]word;
    var r [2]word;
    var i, v int32;
    assert(z, "z");
    for  i = 0; i < 4; i++  {
        v = (z >> (16+(i<<2))) & 7;
        f[i][0] = Truth7[v][0];
        f[i][1] = Truth7[v][1];
    }
    If_Dec7ComposeLut4( int32((z & 0xffff)), f, r );
    f[3][0] = r[0];
    f[3][1] = r[1];
    for  i = 0; i < 3; i++  {
        v = (z >> (48+(i<<2))) & 7;
        f[i][0] = Truth7[v][0];
        f[i][1] = Truth7[v][1];
    }
    If_Dec7ComposeLut4( int32(((z >> 32) & 0xffff)), f, r );
    if r[0] != t[0] || r[1] != t[1] {
        If_DecPrintConfig( z );
        Kit_DsdPrintFromTruth( (unsigned*)t, 7 ); fmt.Printf("\n");
        Kit_DsdPrintFromTruth( (unsigned*)r, 7 ); fmt.Printf("\n");
        fmt.Printf("Verification failed!\n");
    }
}


// count the number of unique cofactors
func If_Dec6CofCount2(t word) int32 {
    var i, Mask int32;
    for  i = 0; i < 16; i++  {
        Mask |= (1 << ((t >> (i<<2)) & 15));
    }
    return BitCount8[((unsigned char*)&Mask)[0]] + BitCount8[((unsigned char*)&Mask)[1]];
}
// count the number of unique cofactors (up to 3)
func If_Dec7CofCount3(t []word) int32 {
    unsigned char * pTruth = (unsigned char *)t;
    var i, iCof2 int32;
    for  i = 1; i < 16; i++  {
        if pTruth[i] == pTruth[0] {
            continue; }
        if iCof2 == 0 {
            iCof2 = i;
        } else if pTruth[i] != pTruth[iCof2] {
            return 3; }
    }
    assert((iCof2 != 0), "iCof2");
    return 2;
}

// permute 6-input function
func If_Dec6SwapAdjacent(t word, v int32) word {
    assert(v < 5, "v < 5");
    return (t & PMasks[v][0]) | ((t & PMasks[v][1]) << (1 << v)) | ((t & PMasks[v][2]) >> (1 << v));
}
func If_Dec6MoveTo(t word, v int32, p int32, Pla2Var []int32, Var2Pla []int32) word {
    var iPlace0, iPlace1 int32;
    assert(Var2Pla[v] >= p, "Var2Pla[v] >= p");
    for  Var2Pla[v] != p  {
        iPlace0 = Var2Pla[v]-1;
        iPlace1 = Var2Pla[v];
        t = If_Dec6SwapAdjacent( t, iPlace0 );
        Var2Pla[Pla2Var[iPlace0]]++;
        Var2Pla[Pla2Var[iPlace1]]--;
        Pla2Var[iPlace0] ^= Pla2Var[iPlace1];
        Pla2Var[iPlace1] ^= Pla2Var[iPlace0];
        Pla2Var[iPlace0] ^= Pla2Var[iPlace1];
    }
    assert(Pla2Var[p] == v, "Pla2Var[p] == v");
    return t;
}

// permute 7-input function
func If_Dec7SwapAdjacent(t []word, v int32) {
    if v == 5 {
        var Temp uint32  = (t[0] >> 32);
        t[0]  = (t[0] & 0xFFFFFFFF) | ((t[1] & 0xFFFFFFFF) << 32);
        t[1] ^= (t[1] & 0xFFFFFFFF) ^ Temp;
        return;
    }
    assert(v < 5, "v < 5");
    t[0] = If_Dec6SwapAdjacent( t[0], v );
    t[1] = If_Dec6SwapAdjacent( t[1], v );
}
func If_Dec7MoveTo(t []word, v int32, p int32, Pla2Var []int32, Var2Pla []int32) {
    var iPlace0, iPlace1 int32;
    assert(Var2Pla[v] >= p, "Var2Pla[v] >= p");
    for  Var2Pla[v] != p  {
        iPlace0 = Var2Pla[v]-1;
        iPlace1 = Var2Pla[v];
        If_Dec7SwapAdjacent( t, iPlace0 );
        Var2Pla[Pla2Var[iPlace0]]++;
        Var2Pla[Pla2Var[iPlace1]]--;
        Pla2Var[iPlace0] ^= Pla2Var[iPlace1];
        Pla2Var[iPlace1] ^= Pla2Var[iPlace0];
        Pla2Var[iPlace0] ^= Pla2Var[iPlace1];
    }
    assert(Pla2Var[p] == v, "Pla2Var[p] == v");
}

// derive binary function
func If_Dec6DeriveCount2(t word, pCof0 *int32, pCof1 *int32) int32 {
    var i, Mask int32;
    *pCof0 = (t & 15);
    *pCof1 = (t & 15);
    for  i = 1; i < 16; i++  {
        if *pCof0 != ((t >> (i<<2)) & 15) {
            *pCof1 = ((t >> (i<<2)) & 15);
            Mask |= (1 << i);
        }
    }
    return Mask;
}
func If_Dec7DeriveCount3(t []word, pCof0 *int32, pCof1 *int32) int32 {
    unsigned char * pTruth = (unsigned char *)t;
    var i, Mask int32;
    *pCof0 = pTruth[0];
    *pCof1 = pTruth[0];
    for  i = 1; i < 16; i++  {
        if *pCof0 != pTruth[i] {
            *pCof1 = pTruth[i];
            Mask |= (1 << i);
        }
    }
    return Mask;
}

// derives decomposition
func If_Dec6Cofactor(t word, iVar int32, fCof1 int32) word {
    assert(iVar >= 0 && iVar < 6, "iVar >= 0 && iVar < 6");
    if fCof1 != 0 {
        return (t & Truth6[iVar]) | ((t & Truth6[iVar]) >> (1<<iVar));
    } else {
        return (t &~Truth6[iVar]) | ((t &~Truth6[iVar]) << (1<<iVar)); }
}
func If_Dec6DeriveDisjoint(t word, Pla2Var []int32, Var2Pla []int32) word {
    var i, Cof0, Cof1 int32;
    var z word  = If_Dec6DeriveCount2( t, &Cof0, &Cof1 );
    for  i = 0; i < 4; i++  {
        z |= (((word)Pla2Var[i+2]) << (16 + 4*i));
    }
    z |= ((word)((Cof1 << 4) | Cof0) << 32);
    z |= ((word)((Cof1 << 4) | Cof0) << 40);
    for  i = 0; i < 2; i++  {
        z |= (((word)Pla2Var[i]) << (48 + 4*i));
    }
    z |= (((word)7) << (48 + 4*i++));
    assert(i == 3, "i == 3");
    return z;
}
func If_Dec6DeriveNonDisjoint(t word, s int32, Pla2Var0 []int32, Var2Pla0 []int32) word {
    var z, c0, c1 word;
    var Cof0, Cof1 [2]int32;
    var Truth0, Truth1, i int32;
    var Pla2Var, Var2Pla [6]int32;
    assert(s >= 2 && s <= 5, "s >= 2 && s <= 5");
    for  i = 0; i < 6; i++  {
        Pla2Var[i] = Pla2Var0[i];
        Var2Pla[i] = Var2Pla0[i];
    }
    for  i = s; i < 5; i++  {
        t = If_Dec6SwapAdjacent( t, i );
        Var2Pla[Pla2Var[i]]++;
        Var2Pla[Pla2Var[i+1]]--;
        Pla2Var[i] ^= Pla2Var[i+1];
        Pla2Var[i+1] ^= Pla2Var[i];
        Pla2Var[i] ^= Pla2Var[i+1];
    }
    c0 = If_Dec6Cofactor( t, 5, 0 );
    c1 = If_Dec6Cofactor( t, 5, 1 );
    assert(2 >= If_Dec6CofCount2(c0), "2 >= If_Dec6CofCount2(c0)");
    assert(2 >= If_Dec6CofCount2(c1), "2 >= If_Dec6CofCount2(c1)");
    Truth0 = If_Dec6DeriveCount2( c0, &Cof0[0], &Cof0[1] );
    Truth1 = If_Dec6DeriveCount2( c1, &Cof1[0], &Cof1[1] );
    z = ((Truth1 & 0xFF) << 8) | (Truth0 & 0xFF);
    for  i = 0; i < 4; i++  {
        z |= (((word)Pla2Var[i+2]) << (16 + 4*i));
    }
    z |= ((word)((Cof0[1] << 4) | Cof0[0]) << 32);
    z |= ((word)((Cof1[1] << 4) | Cof1[0]) << 40);
    for  i = 0; i < 2; i++  {
        z |= (((word)Pla2Var[i]) << (48 + 4*i));
    }
    z |= (((word)7) << (48 + 4*i++));
    z |= (((word)Pla2Var[5]) << (48 + 4*i++));
    assert(i == 4, "i == 4");
    return z;
}
func If_Dec7DeriveDisjoint(t []word, Pla2Var []int32, Var2Pla []int32) word {
    var i, Cof0, Cof1 int32;
    var z word  = If_Dec7DeriveCount3( t, &Cof0, &Cof1 );
    for  i = 0; i < 4; i++  {
        z |= (((word)Pla2Var[i+3]) << (16 + 4*i));
    }
    z |= ((word)((Cof1 << 8) | Cof0) << 32);
    for  i = 0; i < 3; i++  {
        z |= (((word)Pla2Var[i]) << (48 + 4*i));
    }
    z |= (((word)7) << (48 + 4*i));
    return z;
}

func If_Dec6CountOnes(t word) int32 {
    t =    (t & ABC_CONST(0x5555555555555555)) + ((t>> 1) & ABC_CONST(0x5555555555555555));
    t =    (t & ABC_CONST(0x3333333333333333)) + ((t>> 2) & ABC_CONST(0x3333333333333333));
    t =    (t & ABC_CONST(0x0F0F0F0F0F0F0F0F)) + ((t>> 4) & ABC_CONST(0x0F0F0F0F0F0F0F0F));
    t =    (t & ABC_CONST(0x00FF00FF00FF00FF)) + ((t>> 8) & ABC_CONST(0x00FF00FF00FF00FF));
    t =    (t & ABC_CONST(0x0000FFFF0000FFFF)) + ((t>>16) & ABC_CONST(0x0000FFFF0000FFFF));
    return (t & ABC_CONST(0x00000000FFFFFFFF)) +  (t>>32);
}
func If_Dec6HasVar(t word, v int32) int32 {
    return ((t & Truth6[v]) >> (1<<v)) != (t & ~Truth6[v]);
}
func If_Dec7HasVar(t []word, v int32) int32 {
    assert(v >= 0 && v < 7, "v >= 0 && v < 7");
    if v == 6 {
        return t[0] != t[1]; } 
    return ((t[0] & Truth6[v]) >> (1<<v)) != (t[0] & ~Truth6[v])
        || ((t[1] & Truth6[v]) >> (1<<v)) != (t[1] & ~Truth6[v]);
}

func If_DecVerifyPerm(Pla2Var []int32, Var2Pla []int32) {
    var i int32;
    for  i = 0; i < 6; i++  {
        assert(Pla2Var[Var2Pla[i]] == i, "Pla2Var[Var2Pla[i]] == i");
    }
} 
func If_Dec6Perform(t word, fDerive int32) word { 
    var r word;
    var i int32;
    var v int32;
    var u int32;
    var x int32;
    var Count int32;
    var Pla2Var [6]int32;
    var Var2Pla [6]int32;
    // start arrays
    for  i = 0; i < 6; i++  {
        assert((If_Dec6HasVar( t, i ) != 0), "If_Dec6HasVar( t, i )");
        Pla2Var[i] = Var2Pla[i] = i;
    }
    // generate permutations
    i = 0;
    for  v = 0;   v < 6; v++  {
    for  u = v+1; u < 6; u, i = u+1, i+1  {
        t = If_Dec6MoveTo( t, v, 0, Pla2Var, Var2Pla );
        t = If_Dec6MoveTo( t, u, 1, Pla2Var, Var2Pla );
//        If_DecVerifyPerm( Pla2Var, Var2Pla );
        Count = If_Dec6CofCount2( t );
        assert(Count > 1, "Count > 1");
        if Count == 2 {
            if !(fDerive != 0) {
                return 1;
            } else {
                return If_Dec6DeriveDisjoint( t, Pla2Var, Var2Pla );
            } }
        // check non-disjoint decomposition
        if !r && (Count == 3 || Count == 4) {
            for  x = 0; x < 4; x++  {
                var c0 word  = If_Dec6Cofactor( t, x+2, 0 );
                var c1 word  = If_Dec6Cofactor( t, x+2, 1 );
                if If_Dec6CofCount2( c0 ) <= 2 && If_Dec6CofCount2( c1 ) <= 2 {
                    if !(fDerive != 0) {
                        r = 1;
                    } else {
                        r = If_Dec6DeriveNonDisjoint( t, x+2, Pla2Var, Var2Pla );
                    }
                    break;
                }
            }
        }
    }
    }
    assert(i == 15, "i == 15");
    return r;
}
func If_Dec7Perform(t0 []word, fDerive int32) word {
    var t [2]word = [2]word{t0[0], t0[1]}; 
    var i int32;
    var v int32;
    var u int32;
    var y int32;
    var Pla2Var [7]int32;
    var Var2Pla [7]int32;
    // start arrays
    for  i = 0; i < 7; i++  {
/*
        if ( i < 6 )
            assert( If_Dec6HasVar( t[0], i ) || If_Dec6HasVar( t[1], i ) );
        else
            assert( t[0] != t[1] );
*/
        Pla2Var[i] = Var2Pla[i] = i;
    }
    // generate permutations
    for  v = 0;   v < 7; v++  {
    for  u = v+1; u < 7; u++  {
    for  y = u+1; y < 7; y++  {
        If_Dec7MoveTo( t, v, 0, Pla2Var, Var2Pla );
        If_Dec7MoveTo( t, u, 1, Pla2Var, Var2Pla );
        If_Dec7MoveTo( t, y, 2, Pla2Var, Var2Pla );
//        If_DecVerifyPerm( Pla2Var, Var2Pla );
        if If_Dec7CofCount3( t ) == 2 {
            if !(fDerive != 0) {
                return 1;
            } else {
                return If_Dec7DeriveDisjoint( t, Pla2Var, Var2Pla );
            }
        }
    }
    }
    }
    return 0;
}


// support minimization
func If_DecSuppIsMinBase(Supp int32) int32 {
    return (Supp & (Supp+1)) == 0;
}
func If_Dec6TruthShrink(uTruth word, nVars int32, nVarsAll int32, Phase uint32) word {
    var i, k, Var int32;
    assert(nVarsAll <= 6, "nVarsAll <= 6");
    for  i = 0; i < nVarsAll; i++  {
        if Phase & (1 << i) {
            for  k = i-1; k >= Var; k--  {
                uTruth = If_Dec6SwapAdjacent( uTruth, k );
            }
            Var++;
        }
    }
    assert(Var == nVars, "Var == nVars");
    return uTruth;
}
func If_Dec6MinimumBase(uTruth word, pSupp *int32, nVarsAll int32, pnVars *int32) word {
    var v, iVar, uSupp int32;
    assert(nVarsAll <= 6, "nVarsAll <= 6");
    for  v = 0; v < nVarsAll; v++  {
        if If_Dec6HasVar( uTruth, v ) != 0 {
            uSupp |= (1 << v);
            if pSupp != nil {
                pSupp[iVar] = pSupp[v]; }
            iVar++;
        }
    }
    if pnVars != nil {
        *pnVars = iVar; }
    if If_DecSuppIsMinBase( uSupp ) != 0 {
        return uTruth; }
    return If_Dec6TruthShrink( uTruth, iVar, nVarsAll, uSupp );
}

func If_Dec7TruthShrink(uTruth []word, nVars int32, nVarsAll int32, Phase uint32) {
    var i, k, Var int32;
    assert(nVarsAll <= 7, "nVarsAll <= 7");
    for  i = 0; i < nVarsAll; i++  {
        if Phase & (1 << i) {
            for  k = i-1; k >= Var; k--  {
                If_Dec7SwapAdjacent( uTruth, k );
            }
            Var++;
        }
    }
    assert(Var == nVars, "Var == nVars");
}
func If_Dec7MinimumBase(uTruth []word, pSupp *int32, nVarsAll int32, pnVars *int32) {
    var v, iVar, uSupp int32;
    assert(nVarsAll <= 7, "nVarsAll <= 7");
    for  v = 0; v < nVarsAll; v++  {
        if If_Dec7HasVar( uTruth, v ) != 0 {
            uSupp |= (1 << v);
            if pSupp != nil {
                pSupp[iVar] = pSupp[v]; }
            iVar++;
        }
    }
    if pnVars != nil {
        *pnVars = iVar; }
    if If_DecSuppIsMinBase( uSupp ) != 0 {
        return; }
    If_Dec7TruthShrink( uTruth, iVar, nVarsAll, uSupp );
}



// check for MUX decomposition
func If_Dec6SuppSize(t word) int32 {
    var v, Count int32;
    for  v = 0; v < 6; v++  {
        if If_Dec6Cofactor(t, v, 0) != If_Dec6Cofactor(t, v, 1) {
            Count++; }
    }
    return Count;
}
func If_Dec6CheckMux(t word) int32 {
    var v int32;
    for  v = 0; v < 6; v++  {
        if If_Dec6SuppSize(If_Dec6Cofactor(t, v, 0)) < 5 &&
             If_Dec6SuppSize(If_Dec6Cofactor(t, v, 1)) < 5 {
             return v; }
    }
    return -1;
}

// check for MUX decomposition
func If_Dec7Cofactor(t []word, iVar int32, fCof1 int32, r []word) {
    assert(iVar >= 0 && iVar < 7, "iVar >= 0 && iVar < 7");
    if iVar == 6 {
        if fCof1 != 0 {
            r[0] = r[1] = t[1];
        } else {
            r[0] = r[1] = t[0]; }
    
    } else {
        if fCof1 != 0 {
            r[0] = (t[0] & Truth6[iVar]) | ((t[0] & Truth6[iVar]) >> (1<<iVar));
            r[1] = (t[1] & Truth6[iVar]) | ((t[1] & Truth6[iVar]) >> (1<<iVar));
        
        } else {
            r[0] = (t[0] &~Truth6[iVar]) | ((t[0] &~Truth6[iVar]) << (1<<iVar));
            r[1] = (t[1] &~Truth6[iVar]) | ((t[1] &~Truth6[iVar]) << (1<<iVar));
        }
    }
}
func If_Dec7SuppSize(t []word) int32 {
    var c0, c1 [2]word;
    var v, Count int32;
    for  v = 0; v < 7; v++  {
        If_Dec7Cofactor( t, v, 0, c0 );
        If_Dec7Cofactor( t, v, 1, c1 );
        if c0[0] != c1[0] || c0[1] != c1[1] {
            Count++; }
    }
    return Count;
}
func If_Dec7CheckMux(t []word) int32 {
    var c0, c1 [2]word;
    var v int32;
    for  v = 0; v < 7; v++  {
        If_Dec7Cofactor( t, v, 0, c0 );
        If_Dec7Cofactor( t, v, 1, c1 );
        if If_Dec7SuppSize(c0) < 5 && If_Dec7SuppSize(c1) < 5 {
            return v; }
    }
    return -1;
}

// find the best MUX decomposition
func If_Dec6PickBestMux(t word, Cofs []word) int32 {
    var v int32;
    var vBest int32 = -1;
    var Count0 int32;
    var Count1 int32;
    var CountBest int32 = 1000;
    for  v = 0; v < 6; v++  {
        Count0 = If_Dec6SuppSize( If_Dec6Cofactor(t, v, 0) );
        Count1 = If_Dec6SuppSize( If_Dec6Cofactor(t, v, 1) );
        if Count0 < 5 && Count1 < 5 && CountBest > Count0 + Count1 {
            CountBest = Count0 + Count1;
            vBest = v;
            Cofs[0] = If_Dec6Cofactor(t, v, 0);
            Cofs[1] = If_Dec6Cofactor(t, v, 1);
        }
    }
    return vBest;
}
func If_Dec7PickBestMux(t []word, c0r []word, c1r []word) int32 {
    var c0, c1 [2]word;
    var v int32;
    var vBest int32 = -1;
    var Count0 int32;
    var Count1 int32;
    var CountBest int32 = 1000;
    for  v = 0; v < 7; v++  {
        If_Dec7Cofactor( t, v, 0, c0 );
        If_Dec7Cofactor( t, v, 1, c1 );
        Count0 = If_Dec7SuppSize(c0);
        Count1 = If_Dec7SuppSize(c1);
        if Count0 < 5 && Count1 < 5 && CountBest > Count0 + Count1 {
            CountBest = Count0 + Count1;
            vBest = v;
            c0r[0] = c0[0]; c0r[1] = c0[1];
            c1r[0] = c1[0]; c1r[1] = c1[1];
        }
    }
    return vBest;
}



/**Function*************************************************************

  Synopsis    [Checks decomposability ]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// count the number of unique cofactors
func If_Dec5CofCount2(t word, x int32, y int32, Pla2Var *int32, t0 word, fDerive int32) word {
    var m, i, Mask int32;
    assert(x >= 0 && x < 4, "x >= 0 && x < 4");
    assert(y >= 0 && y < 4, "y >= 0 && y < 4");
    for  m = 0; m < 4; m++  {
        for  Mask = i = 0; i < 16; i++  {
            if ((i >> x) & 1) == ((m >> 0) & 1) && ((i >> y) & 1) == ((m >> 1) & 1) {
                Mask |= (1 << ((t >> (i<<1)) & 3)); }
        }
        if BitCount8[Mask & 0xF] > 2 {
            return 0; }
    }
//    Kit_DsdPrintFromTruth( (unsigned *)&t, 5 ); printf( "\n" );
    if !(fDerive != 0) {
        return 1;
    } else {
        var fHas2, fHas3 int32;
        // composition function C depends on {x, y, Out, v[0]}
        // decomposed function D depends on {x, y, z1, z2}
        var F [4]word = [4]word{ 0, ABC_CONST(0x5555555555555555), ABC_CONST(0xAAAAAAAAAAAAAAAA), ~((word)0) };
        var C2 [4]word;
        var D2 [4]word ;
        var C1 [2]word;
        var D1 [2]word;
        var C word;
        var D word;
        var z word;
        var v int32;
        var zz1 int32 = -1;
        var zz2 int32 = -1;
        // find two variables
        for  v = 0; v < 4; v++  {
            if v != x && v != y {  zz1 = v; break; }
        }
        for  v = 1; v < 4; v++  {
            if v != x && v != y && v != zz1 {  zz2 = v; break; }
        }
        assert(zz1 != -1 && zz2 != -1, "zz1 != -1 && zz2 != -1");
        // find the cofactors
        for  m = 0; m < 4; m++  {
            // for each cofactor, derive C2 and D2
            for  Mask = i = 0; i < 16; i++  {
                if ((i >> x) & 1) == ((m >> 0) & 1) && ((i >> y) & 1) == ((m >> 1) & 1) {
                    Mask |= (1 << ((t >> (i<<1)) & 3)); }
            }
            // find the values
            if BitCount8[Mask & 0xF] == 1 {
                C2[m] = F[Abc_Tt6FirstBit( Mask )];
                D2[m] = ~(word)0;
            
            } else if BitCount8[Mask & 0xF] == 2 {
                var Bit0 int32  = Abc_Tt6FirstBit( Mask );
                var Bit1 int32  = Abc_Tt6FirstBit( Mask ^ (((word)1)<<Bit0) );
                C2[m] = (F[Bit1] & Truth6[1]) | (F[Bit0] & ~Truth6[1]);
                // find Bit1 appears
                for  Mask = i = 0; i < 16; i++  {
                    if ((i >> x) & 1) == ((m >> 0) & 1) && ((i >> y) & 1) == ((m >> 1) & 1) {
                        if Bit1 == ((t >> (i<<1)) & 3) {
                            D2[m] |= (((word)1) << ( (((i >> zz2) & 1) << 1) | ((i >> zz1) & 1) )); } }
                }
            
            } else { assert(false, "0"); }
            D2[m] = Abc_Tt6Stretch( D2[m], 2 );
        }

        // combine
        C1[0] = (C2[1] & Truth6[2]) | (C2[0] & ~Truth6[2]);
        C1[1] = (C2[3] & Truth6[2]) | (C2[2] & ~Truth6[2]);
        C     = (C1[1] & Truth6[3]) | (C1[0] & ~Truth6[3]);

        // combine
        D1[0] = (D2[1] & Truth6[2]) | (D2[0] & ~Truth6[2]);
        D1[1] = (D2[3] & Truth6[2]) | (D2[2] & ~Truth6[2]);
        D     = (D1[1] & Truth6[3]) | (D1[0] & ~Truth6[3]);

//        printf( "\n" );
//        Kit_DsdPrintFromTruth( (unsigned *)&C, 5 ); printf("\n");
//        Kit_DsdPrintFromTruth( (unsigned *)&D, 5 ); printf("\n");

        // create configuration
        // find one
        fHas2 = Abc_TtHasVar(&D, 5, 2);
        fHas3 = Abc_TtHasVar(&D, 5, 3);
        if fHas2 && (fHas3 != 0) {
            z = D & 0xFFFF;
            z |= (((word)Pla2Var[zz1+1]) << (16 + 4*0));
            z |= (((word)Pla2Var[zz2+1]) << (16 + 4*1));
            z |= (((word)Pla2Var[x+1])   << (16 + 4*2));
            z |= (((word)Pla2Var[y+1])   << (16 + 4*3));
        
        } else if fHas2 && !(fHas3 != 0) {
            z = D & 0xFFFF;
            z |= (((word)Pla2Var[zz1+1]) << (16 + 4*0));
            z |= (((word)Pla2Var[zz2+1]) << (16 + 4*1));
            z |= (((word)Pla2Var[x+1])   << (16 + 4*2));
            z |= (((word)6)              << (16 + 4*3));
        
        } else if !fHas2 && (fHas3 != 0) {
            Abc_TtSwapVars( &D, 5, 2, 3 );
            z = D & 0xFFFF;
            z |= (((word)Pla2Var[zz1+1]) << (16 + 4*0));
            z |= (((word)Pla2Var[zz2+1]) << (16 + 4*1));
            z |= (((word)Pla2Var[y+1])   << (16 + 4*2));
            z |= (((word)6)              << (16 + 4*3));
        
        } else {
            z = D & 0xFFFF;
            z |= (((word)Pla2Var[zz1+1]) << (16 + 4*0));
            z |= (((word)Pla2Var[zz2+1]) << (16 + 4*1));
            z |= (((word)6)              << (16 + 4*2));
            z |= (((word)6)              << (16 + 4*3));
        }

        // second one
        fHas2 = Abc_TtHasVar(&C, 5, 2);
        fHas3 = Abc_TtHasVar(&C, 5, 3);
        if fHas2 && (fHas3 != 0) {
            z |= ((C & 0xFFFF) << 32);
            z |= (((word)Pla2Var[0])     << (48 + 4*0));
            z |= (((word)7)              << (48 + 4*1));
            z |= (((word)Pla2Var[x+1])   << (48 + 4*2));
            z |= (((word)Pla2Var[y+1])   << (48 + 4*3));
        
        } else if fHas2 && !(fHas3 != 0) {
            z |= ((C & 0xFFFF) << 32);
            z |= (((word)Pla2Var[0])     << (48 + 4*0));
            z |= (((word)7)              << (48 + 4*1));
            z |= (((word)Pla2Var[x+1])   << (48 + 4*2));
            z |= (((word)6)              << (48 + 4*3));
        
        } else if !fHas2 && (fHas3 != 0) {
            Abc_TtSwapVars( &C, 5, 2, 3 );
            z |= ((C & 0xFFFF) << 32);
            z |= (((word)Pla2Var[0])     << (48 + 4*0));
            z |= (((word)7)              << (48 + 4*1));
            z |= (((word)Pla2Var[y+1])   << (48 + 4*2));
            z |= (((word)6)              << (48 + 4*3));
        
        } else {
            z |= ((C & 0xFFFF) << 32);
            z |= (((word)Pla2Var[0])     << (48 + 4*0));
            z |= (((word)7)              << (48 + 4*1));
            z |= (((word)6)              << (48 + 4*2));
            z |= (((word)6)              << (48 + 4*3));
        }

        // verify
        {
            var t1 word  = If_Dec6Truth( z );
            if t1 != t0 {
                fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&C2[0], 5 ); fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&C2[1], 5 ); fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&C2[2], 5 ); fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&C2[3], 5 ); fmt.Printf("\n");

                fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&D2[0], 5 ); fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&D2[1], 5 ); fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&D2[2], 5 ); fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&D2[3], 5 ); fmt.Printf("\n");

                fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&C, 5 ); fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&D, 5 ); fmt.Printf("\n");

                fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&t1, 5 ); fmt.Printf("\n");
                Kit_DsdPrintFromTruth( (unsigned *)&t0, 5 ); fmt.Printf("\n");
            }
            assert(t1 == t0, "t1 == t0");
        }
        return z;
    }
}
func If_Dec5Perform(t word, fDerive int32) word {
    var Pla2Var, Var2Pla [7]int32;
    var i, j, v int32;
    var t0 word  = t;
/*
    word c0, c1, c00, c01, c10, c11;
    for ( i = 0; i < 5; i++ )
    {
        c0 = If_Dec6Cofactor( t, i, 0 );
        c1 = If_Dec6Cofactor( t, i, 1 );
        if ( c0 == 0 )
            return 1;
        if ( ~c0 == 0 )
            return 1;
        if ( c1 == 0 )
            return 1;
        if ( ~c1 == 0 )
            return 1;
       if ( c0 == ~c1 )
            return 1;
    }
    for ( i = 0; i < 4; i++ )
    {
        c0 = If_Dec6Cofactor( t, i, 0 );
        c1 = If_Dec6Cofactor( t, i, 1 );
        for ( j = i + 1; j < 5; j++ )
        {
            c00 = If_Dec6Cofactor( c0, j, 0 );
            c01 = If_Dec6Cofactor( c0, j, 1 );
            c10 = If_Dec6Cofactor( c1, j, 0 );
            c11 = If_Dec6Cofactor( c1, j, 1 );
            if ( c00 == c01 && c00 == c10 )
                return 1;
            if ( c11 == c01 && c11 == c10 )
                return 1;
            if ( c11 == c00 && c11 == c01 )
                return 1;
            if ( c11 == c00 && c11 == c10 )
                return 1;
            if ( c00 == c11 && c01 == c10 )
                return 1;
        }
    }
*/
    // start arrays
    for  i = 0; i < 7; i++  {
        Pla2Var[i] = Var2Pla[i] = i;
    }
    // generate permutations
    for  v = 0; v < 5; v++  {
        t = If_Dec6MoveTo( t, v, 0, Pla2Var, Var2Pla );
        If_DecVerifyPerm( Pla2Var, Var2Pla );
        for  i = 0; i < 4; i++  {
            for  j = i + 1; j < 4; j++  {
                var z word  = If_Dec5CofCount2( t, i, j, Pla2Var, t0, fDerive );
                if z {
/*
                    if ( v == 0 && i == 1 && j == 2 )
                    {
                          Kit_DsdPrintFromTruth( (unsigned *)&t, 5 ); printf( "\n" );
                    }
*/
                    return z;
                }
            }
        }
    }

/*
    // start arrays
    for ( i = 0; i < 7; i++ )
        Pla2Var[i] = Var2Pla[i] = i;

    t = t0;
    for ( v = 0; v < 5; v++ )
    {
        int x, y;

        t = If_Dec6MoveTo( t, v, 0, Pla2Var, Var2Pla );
        If_DecVerifyPerm( Pla2Var, Var2Pla );

        for ( i = 0; i < 16; i++ )
            printf( "%d ", ((t >> (i<<1)) & 3) );
        printf( "\n" );

        for ( x = 0; x < 4; x++ )
        for ( y = x + 1; y < 4; y++ )
        {
            for ( i = 0; i < 16; i++ )
                if ( !((i >> x) & 1) && !((i >> y) & 1) )
                    printf( "%d ", ((t >> (i<<1)) & 3) );
            printf( "\n" );

            for ( i = 0; i < 16; i++ )
                if ( ((i >> x) & 1) && !((i >> y) & 1) )
                    printf( "%d ", ((t >> (i<<1)) & 3) );
            printf( "\n" );

            for ( i = 0; i < 16; i++ )
                if ( !((i >> x) & 1) && ((i >> y) & 1) )
                    printf( "%d ", ((t >> (i<<1)) & 3) );
            printf( "\n" );

            for ( i = 0; i < 16; i++ )
                if ( ((i >> x) & 1) && ((i >> y) & 1) )
                    printf( "%d ", ((t >> (i<<1)) & 3) );
            printf( "\n" );
            printf( "\n" );
        }
    }
*/

//    Kit_DsdPrintFromTruth( (unsigned *)&t, 5 ); printf( "\n" );
    return 0;
}

func If_Dec5PerformEx() word {
    var z word;
    // find one
    z = (word)(0x17ac & 0xFFFF);
    z |= (((word)3) << (16 + 4*0));
    z |= (((word)4) << (16 + 4*1));
    z |= (((word)1) << (16 + 4*2));
    z |= (((word)2) << (16 + 4*3));
    // second one
    z |= (((word)(0x179a & 0xFFFF)) << 32);
    z |= (((word)0) << (48 + 4*0));
    z |= (((word)7) << (48 + 4*1));
    z |= (((word)1) << (48 + 4*2));
    z |= (((word)2) << (48 + 4*3));
    return z;
}
func If_Dec5PerformTest() {
    var z, t, t1 word;
//    s = If_Dec5PerformEx();
//    t = If_Dec6Truth( s );
    t = ABC_CONST(0xB0F3B0FFB0F3B0FF);

    Kit_DsdPrintFromTruth( (unsigned *)&t, 5 ); fmt.Printf("\n");

    z = If_Dec5Perform( t, 1 );
    t1 = If_Dec6Truth( z );
    assert(t == t1, "t == t1");
}


/**Function*************************************************************

  Synopsis    [Performs additional check.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutPerformDerive07(p *If_Man_t, pTruth *uint32, nVars int32, nLeaves int32, pStr string) word {
    if nLeaves < 5 {
        return 1; }
    if nLeaves == 5 {
        var z word;
        var t word = ((word)pTruth[0] << 32) | (word)pTruth[0];
        z = If_Dec5Perform( t, 1 );
        If_Dec6Verify( t, z );
        return z;
    }
    if nLeaves == 6 {
        var z word;
        var t word = ((word *)pTruth)[0];
        z = If_Dec6Perform( t, 1 );
        If_Dec6Verify( t, z );
        return z;
    }
    if nLeaves == 7 {
        var z word;
        var t [2]word;
        t[0] = ((word *)pTruth)[0];
        t[1] = ((word *)pTruth)[1];
        z = If_Dec7Perform( t, 1 );
        If_Dec7Verify( t, z );
        return z;
    }
    assert(false, "0");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Performs additional check.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutPerformCheck07(p *If_Man_t, pTruth *uint32, nVars int32, nLeaves int32, pStr string) int32 {
    var fDerive int32;
    var v int32;
    // skip non-support minimal
    for  v = 0; v < nLeaves; v++  {
        if !Abc_TtHasVar( (word *)pTruth, nVars, v ) {
            return 0; }
    }
    // check
    if nLeaves < 5 {
        return 1; }
    if nLeaves == 5 {
        var z word;
        var t word = ((word)pTruth[0] << 32) | (word)pTruth[0];
        z = If_Dec5Perform( t, fDerive );
        if (fDerive != 0) && z {
            If_Dec6Verify( t, z ); }
        return z != 0;
    }
    if nLeaves == 6 {
        var z word;
        var t word = ((word *)pTruth)[0];
        z = If_Dec6Perform( t, fDerive );
        if (fDerive != 0) && z {
//            If_DecPrintConfig( z );
            If_Dec6Verify( t, z );
        }
//        if ( z == 0 )
//            Extra_PrintHex(stdout, (unsigned *)&t, 6), printf( "  " ), Kit_DsdPrintFromTruth( (unsigned *)&t, 6 ), printf( "\n" );
        return z != 0;
    }
    if nLeaves == 7 {
        var z word;
        var t [2]word;
        t[0] = ((word *)pTruth)[0];
        t[1] = ((word *)pTruth)[1];
//        if ( If_Dec7CheckMux(t) >= 0 )
//            return 1;
        z = If_Dec7Perform( t, fDerive );
        if (fDerive != 0) && z {
            If_Dec7Verify( t, z ); }
        return z != 0;
    }
    assert(false, "0");
    return 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


package if_

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [ifDec08.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Performs additional check.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifDec08.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

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
var Truth10 [10][16]word = [10][16]word{
    {ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA)},
    {ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC)},
    {ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0)},
    {ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00)},
    {ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000)},
    {ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000)},
    {ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF)},
    {ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF)},
    {ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF)},
    {ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF)},
};

// extern void Kit_DsdPrintFromTruth( unsigned * pTruth, int nVars );

// extern void Extra_PrintBinary( FILE * pFile, unsigned Sign[], int nBits );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

func If_Dec08WordNum(nVars int32) int32 {
    if nVars <= 6 {
        return 1;
    } else {
        return 1 << (nVars-6);
    }
}
func If_Dec08PrintConfigOne(z uint32) {
    var t uint32;
    t = (z & 0xffff) | ((z & 0xffff) << 16);
    Extra_PrintBinary( os.Stdout, &t, 16 ); 
    fmt.Printf(" ");
    Kit_DsdPrintFromTruth( &t, 4 );
    fmt.Printf(" ");
    fmt.Printf(" %d", (z >> 16) & 7);
    fmt.Printf(" %d", (z >> 20) & 7);
    fmt.Printf(" %d", (z >> 24) & 7);
    fmt.Printf(" %d", (z >> 28) & 7);
    fmt.Printf("\n");
}
func If_Dec08PrintConfig(pZ *uint32) {
    for  *pZ  {
        If_Dec08PrintConfigOne( *pZ++ );
    }
}
func If_Dec08ComposeLut4(t int32, pF **word, pR *word, nVars int32) {
    var pC [16]word;
    var m, w, v, nWords int32;
    assert(nVars <= 10, "nVars <= 10");
    nWords = If_Dec08WordNum( nVars );
    for  w = 0; w < nWords; w++  {
        pR[w] = 0;
    }
    for  m = 0; m < 16; m++  {
        if !((t >> m) & 1) {
            continue; }
        for  w = 0; w < nWords; w++  {
            pC[w] = ~(word)0;
        }
        for  v = 0; v < 4; v++  {
            for  w = 0; w < nWords; w++  {
                pC[w] &= ((m >> v) & 1) ? pF[v][w] : ~pF[v][w];
            }
        }
        for  w = 0; w < nWords; w++  {
            pR[w] |= pC[w];
        }
    }
}
func If_Dec08Verify(pF *word, nVars int32, pZ *uint32) {
    var pN [16][16]word;
    var pG [4]*word;
    var i, w, v, k, nWords int32;
    var z uint32;
    nWords = If_Dec08WordNum( nVars );
    for  k = 0; k < nVars; k++  {
        for  w = 0; w < nWords; w++  {
            pN[k][w] = Truth10[k][w];
        }
    }
    for  i = 0; (z = pZ[i]); i, k = i+1, k+1  {
        for  v = 0; v < 4; v++  {
            pG[v] = pN[ (z >> (16+(v<<2))) & 7 ];
        }
        If_Dec08ComposeLut4( int32((z & 0xffff)), pG, pN[k], nVars );
    }
    k--;
    for  w = 0; w < nWords; w++  {
        if pN[k][w] != pF[w] {
            If_Dec08PrintConfig( pZ );
            Kit_DsdPrintFromTruth( (unsigned*)pF, nVars ); fmt.Printf("\n");
            Kit_DsdPrintFromTruth( (unsigned*)pN[k], nVars ); fmt.Printf("\n");
            fmt.Printf("Verification failed!\n");
            break;
        }
    }
}


// count the number of unique cofactors
func If_Dec08CofCount2(pF *word, nVars int32) int32 {
    var nShift int32  = (1 << (nVars - 3));
    var Mask word   = (((word)1) << nShift) - 1;
    var iCof0 word  = pF[0] & Mask;
    var iCof1 word = iCof0;
    var iCof word;
    var i int32;
    assert(nVars >= 6 && nVars <= 8, "nVars >= 6 && nVars <= 8");
//    if ( nVars == 10 )
//        Mask = ~(word)0;
    for  i = 1; i < 8; i++  {
        iCof = (pF[(i * nShift) / 64] >> ((i * nShift) & 63)) & Mask;
        if iCof == iCof0 {
            continue; }
        if iCof1 == iCof0 {
            iCof1 = iCof;
        } else if iCof != iCof1 {
            return 3; }
    }
    return 2;
}
func If_Dec08CofCount(pF *word, nVars int32) int32 {
    var nShift int32  = (1 << (nVars - 3));
    var Mask word   = (((word)1) << nShift) - 1;
    var iCofs [16]word;
    var iCof word;
    var i int32;
    var c int32;
    var nCofs int32 = 1;
//    if ( nVars == 10 )
//        Mask = ~(word)0;
    iCofs[0] = pF[0] & Mask;
    for  i = 1; i < 8; i++  {
        iCof = (pF[(i * nShift) / 64] >> ((i * nShift) & 63)) & Mask;
        for  c = 0; c < nCofs; c++  {
            if iCof == iCofs[c] {
                break; }
        }
        if c == nCofs {
            iCofs[nCofs++] = iCof; }
    }
    return nCofs;
}


// variable permutation for large functions
func If_Dec08Copy(pOut *word, pIn *word, nVars int32) {
    var w int32;
    var nWords int32 = If_Dec08WordNum( nVars );
    for  w = 0; w < nWords; w++  {
        pOut[w] = pIn[w];
    }
}
func If_Dec08SwapAdjacent(pOut *word, pIn *word, iVar int32, nVars int32) {
    var i int32;
    var k int32;
    var nWords int32 = If_Dec08WordNum( nVars );
    assert(iVar < nVars - 1, "iVar < nVars - 1");
    if iVar < 5 {
        var Shift int32  = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            pOut[i] = (pIn[i] & PMasks[iVar][0]) | ((pIn[i] & PMasks[iVar][1]) << Shift) | ((pIn[i] & PMasks[iVar][2]) >> Shift);
        }
    
    } else if iVar > 5 {
        var Step int32  = (1 << (iVar - 6));
        for  k = 0; k < nWords; k += 4*Step  {
            for  i = 0; i < Step; i++  {
                pOut[i] = pIn[i];
            }
            for  i = 0; i < Step; i++  {
                pOut[Step+i] = pIn[2*Step+i];
            }
            for  i = 0; i < Step; i++  {
                pOut[2*Step+i] = pIn[Step+i];
            }
            for  i = 0; i < Step; i++  {
                pOut[3*Step+i] = pIn[3*Step+i];
            }
            pIn  += 4*Step;
            pOut += 4*Step;
        }
    
    } else { // if ( iVar == 5 )
    
        for  i = 0; i < nWords; i += 2  {
            pOut[i]   = (pIn[i]   & ABC_CONST(0x00000000FFFFFFFF)) | ((pIn[i+1] & ABC_CONST(0x00000000FFFFFFFF)) << 32);
            pOut[i+1] = (pIn[i+1] & ABC_CONST(0xFFFFFFFF00000000)) | ((pIn[i]   & ABC_CONST(0xFFFFFFFF00000000)) >> 32);
        }
    }
}
func If_Dec08MoveTo(pF *word, nVars int32, v int32, p int32, Pla2Var []int32, Var2Pla []int32) {
    var pG [16]word;
    var pIn *word = pF;
    var pOut *word = pG;
    var pTemp *word;
    var iPlace0, iPlace1, Count int32;
    assert(Var2Pla[v] <= p, "Var2Pla[v] <= p");
    for  Var2Pla[v] != p  {
        iPlace0 = Var2Pla[v];
        iPlace1 = Var2Pla[v]+1;
        If_Dec08SwapAdjacent( pOut, pIn, iPlace0, nVars );
        pTemp = pIn; pIn = pOut;
pOut = pTemp;
        Var2Pla[Pla2Var[iPlace0]]++;
        Var2Pla[Pla2Var[iPlace1]]--;
        Pla2Var[iPlace0] ^= Pla2Var[iPlace1];
        Pla2Var[iPlace1] ^= Pla2Var[iPlace0];
        Pla2Var[iPlace0] ^= Pla2Var[iPlace1];
        Count++;
    }
    if Count & 1 {
        If_Dec08Copy( pF, pIn, nVars ); }
    assert(Pla2Var[p] == v, "Pla2Var[p] == v");
}
/*
// derive binary function
static inline int If_Dec08DeriveCount2( word * pF, word * pRes, int nVars )
{
    int nShift = (1 << (nVars - 4));
    word Mask  = (((word)1) << nShift) - 1;
    int i, MaskDec = 0;
    word iCof0 = pF[0] & Mask;
    word iCof1 = pF[0] & Mask;
    word iCof, * pCof0, * pCof1;
    if ( nVars == 10 )
        Mask = ~(word)0;
    for ( i = 1; i < 16; i++ )
    {
        iCof = (pF[(i * nShift) / 64] >> ((i * nShift) & 63)) & Mask;
        if ( *pCof0 != iCof )
        {
            *pCof1 = iCof;
            MaskDec |= (1 << i);
        }
    }
    *pRes = ((iCof1 << nShift) | iCof0);
    return MaskDec;
}
static inline word If_DecTruthStretch( word t, int nVars )
{
    assert( nVars > 1 );
    if ( nVars == 1 )
        nVars++, t = (t & 0x3) | ((t & 0x3) << 2);
    if ( nVars == 2 )
        nVars++, t = (t & 0xF) | ((t & 0xF) << 4);
    if ( nVars == 3 )
        nVars++, t = (t & 0xFF) | ((t & 0xFF) << 8);
    if ( nVars == 4 )
        nVars++, t = (t & 0xFFFF) | ((t & 0xFFFF) << 16);
    if ( nVars == 5 )
        nVars++, t = (t & 0xFFFFFFFF) | ((t & 0xFFFFFFFF) << 32);
    assert( nVars >= 6 );
}
*/

// support minimization
func If_DecSuppIsMinBase(Supp int32) int32 {
    return (Supp & (Supp+1)) == 0;
}
func If_Dec08HasVar(t *word, nVars int32, iVar int32) int32 {
    var nWords int32  = If_Dec08WordNum( nVars );
    assert(iVar < nVars, "iVar < nVars");
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            if (t[i] & ~Truth6[iVar]) != ((t[i] & Truth6[iVar]) >> Shift) {
                return 1; }
        }
        return 0;
    
    } else {
        var i int32;
        var k int32;
        var Step int32 = (1 << (iVar - 6));
        for  k = 0; k < nWords; k += 2*Step  {
            for  i = 0; i < Step; i++  {
                if t[i] != t[Step+i] {
                    return 1; }
            }
            t += 2*Step;
        }
        return 0;
    }
}
func If_Dec08Support(t *word, nVars int32) int32 {
    var v, Supp int32;
    for  v = 0; v < nVars; v++  {
        if If_Dec08HasVar( t, nVars, v ) != 0 {
            Supp |= (1 << v); }
    }
    return Supp;
}

// checks
func If_Dec08Cofactors(pF *word, nVars int32, iVar int32, pCof0 *word, pCof1 *word) {
    var nWords int32  = If_Dec08WordNum( nVars );
    assert(iVar < nVars, "iVar < nVars");
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            pCof0[i] = (pF[i] & ~Truth6[iVar]) | ((pF[i] & ~Truth6[iVar]) << Shift);
            pCof1[i] = (pF[i] &  Truth6[iVar]) | ((pF[i] &  Truth6[iVar]) >> Shift);
        }
        return;
    
    } else {
        var i int32;
        var k int32;
        var Step int32 = (1 << (iVar - 6));
        for  k = 0; k < nWords; k += 2*Step  {
            for  i = 0; i < Step; i++  {
                pCof0[i] = pCof0[Step+i] = pF[i];
                pCof1[i] = pCof1[Step+i] = pF[Step+i];
            }
            pF    += 2*Step;
            pCof0 += 2*Step;
            pCof1 += 2*Step;
        }
        return;
    }
}
func If_Dec08Count16(Num16 int32) int32 {
    assert(Num16 < (1<<16)-1, "Num16 < (1<<16)-1");
    return BitCount8[Num16 & 255] + BitCount8[(Num16 >> 8) & 255];
}
func If_DecVerifyPerm(Pla2Var []int32, Var2Pla []int32, nVars int32) {
    var i int32;
    for  i = 0; i < nVars; i++  {
        assert(Pla2Var[Var2Pla[i]] == i, "Pla2Var[Var2Pla[i]] == i");
    }
} 
func If_Dec08Perform(pF *word, nVars int32, fDerive int32) int32 { 
//    static int Cnt = 0;
    var pCof0, pCof1 [16]word;
    var Pla2Var [10]int32;
    var Var2Pla [10]int32;
    var Count [210]int32;
    var Masks [210]int32;
    var i, i0, i1, i2, v, x int32;
    assert(nVars >= 6 && nVars <= 8, "nVars >= 6 && nVars <= 8");
    // start arrays
    for  i = 0; i < nVars; i++  {
        assert((If_Dec08HasVar( pF, nVars, i ) != 0), "If_Dec08HasVar( pF, nVars, i )");
        Pla2Var[i] = Var2Pla[i] = i;
    }
/*
    Cnt++;
//if ( Cnt == 108 )
{
printf( "%d\n", Cnt );
//Extra_PrintHex( stdout, (unsigned *)pF, nVars );
//printf( "\n" ); 
Kit_DsdPrintFromTruth( (unsigned *)pF, nVars );
printf( "\n" );
printf( "\n" );
}
*/
    // generate permutations
    v = 0;
    for  i0 = 0;    i0 < nVars; i0++  {
    for  i1 = i0+1; i1 < nVars; i1++  {
    for  i2 = i1+1; i2 < nVars; i2, v = i2+1, v+1  {
        If_Dec08MoveTo( pF, nVars, i0, nVars-1, Pla2Var, Var2Pla );
        If_Dec08MoveTo( pF, nVars, i1, nVars-2, Pla2Var, Var2Pla );
        If_Dec08MoveTo( pF, nVars, i2, nVars-3, Pla2Var, Var2Pla );
        If_DecVerifyPerm( Pla2Var, Var2Pla, nVars );
        Count[v] = If_Dec08CofCount( pF, nVars );
        Masks[v] = (1 << i0) | (1 << i1) | (1 << i2);
        assert(Count[v] > 1, "Count[v] > 1");
//printf( "%d ", Count[v] );
        if Count[v] == 2 || Count[v] > 5 {
            continue; }
        for  x = 0; x < 4; x++  {
            If_Dec08Cofactors( pF, nVars, nVars-1-x, pCof0, pCof1 );
            if If_Dec08CofCount2(pCof0, nVars) <= 2 && If_Dec08CofCount2(pCof1, nVars) <= 2 {
                Count[v] = -Count[v];
                break;
            }
        }
    }
    }
    }
//printf( "\n" );
    assert(v <= 210, "v <= 210");
    // check if there are compatible bound sets
    for  i0 = 0; i0 < v; i0++  {
    for  i1 = i0+1; i1 < v; i1++  {
        if If_Dec08Count16(Masks[i0] & Masks[i1]) > 8 - nVars {
            continue; }
        if nVars == 8 {
            if Count[i0] == 2 && Count[i1] == 2 {
                return 1; }
        
        } else if nVars == 7 {
            if (Count[i0] == 2 && Count[i1] == 2) || 
                 (Count[i0] == 2 && Count[i1] <  0) || 
                 (Count[i0] <  0 && Count[i1] == 2) {
                return 1; }
        
        } else {
            if (Count[i0] == 2 && Count[i1] == 2) || 
                 (Count[i0] == 2 && Count[i1] <  0) || 
                 (Count[i0] <  0 && Count[i1] == 2) || 
                 (Count[i0] <  0 && Count[i1] <  0) {
                return 1; }
        }
    }
    }
//    printf( "not found\n" );
    return 0;
}


/**Function*************************************************************

  Synopsis    [Performs additional check.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutPerformCheck08(p *If_Man_t, pTruth *uint32, nVars int32, nLeaves int32, pStr string) int32 {
    var nSupp, fDerive int32;
    var pF [16]word;
    if nLeaves <= 5 {
        return 1; }
    If_Dec08Copy( pF, (word *)pTruth, nVars );
    nSupp = If_Dec08Support( pF, nLeaves );
    if !nSupp || !(If_DecSuppIsMinBase(nSupp) != 0) {
        return 0; }
    if If_Dec08Perform( pF, nLeaves, fDerive ) != 0 {
//        printf( "1" );
        return 1;
//        If_Dec08Verify( t, nLeaves, NULL );
    }
//    printf( "0" );
    return 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


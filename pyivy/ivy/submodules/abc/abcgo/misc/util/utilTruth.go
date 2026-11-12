package util

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [utilTruth.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Truth table manipulation.]

  Synopsis    [Truth table manipulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 28, 2012.]

  Revision    [$Id: utilTruth.h,v 1.00 2012/10/28 00:00:00 alanmi Exp $]

***********************************************************************/
 



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

var s_Truths6 [6]word = [6]word{
    ABC_CONST(0xAAAAAAAAAAAAAAAA),
    ABC_CONST(0xCCCCCCCCCCCCCCCC),
    ABC_CONST(0xF0F0F0F0F0F0F0F0),
    ABC_CONST(0xFF00FF00FF00FF00),
    ABC_CONST(0xFFFF0000FFFF0000),
    ABC_CONST(0xFFFFFFFF00000000),
};

var s_Truths6Neg [6]word = [6]word{
    ABC_CONST(0x5555555555555555),
    ABC_CONST(0x3333333333333333),
    ABC_CONST(0x0F0F0F0F0F0F0F0F),
    ABC_CONST(0x00FF00FF00FF00FF),
    ABC_CONST(0x0000FFFF0000FFFF),
    ABC_CONST(0x00000000FFFFFFFF),
};

var s_TruthXors [6]word = [6]word{
    ABC_CONST(0x0000000000000000),
    ABC_CONST(0x6666666666666666),
    ABC_CONST(0x6969696969696969),
    ABC_CONST(0x6996699669966996),
    ABC_CONST(0x6996966969969669),
    ABC_CONST(0x6996966996696996),
};

var s_PMasks [5][3]word = [5][3]word{
    { ABC_CONST(0x9999999999999999), ABC_CONST(0x2222222222222222), ABC_CONST(0x4444444444444444) },
    { ABC_CONST(0xC3C3C3C3C3C3C3C3), ABC_CONST(0x0C0C0C0C0C0C0C0C), ABC_CONST(0x3030303030303030) },
    { ABC_CONST(0xF00FF00FF00FF00F), ABC_CONST(0x00F000F000F000F0), ABC_CONST(0x0F000F000F000F00) },
    { ABC_CONST(0xFF0000FFFF0000FF), ABC_CONST(0x0000FF000000FF00), ABC_CONST(0x00FF000000FF0000) },
    { ABC_CONST(0xFFFF00000000FFFF), ABC_CONST(0x00000000FFFF0000), ABC_CONST(0x0000FFFF00000000) },
};

static word s_PPMasks[5][6][3] = {
    { 
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 0 0  
        { ABC_CONST(0x9999999999999999), ABC_CONST(0x2222222222222222), ABC_CONST(0x4444444444444444) }, // 0 1  
        { ABC_CONST(0xA5A5A5A5A5A5A5A5), ABC_CONST(0x0A0A0A0A0A0A0A0A), ABC_CONST(0x5050505050505050) }, // 0 2 
        { ABC_CONST(0xAA55AA55AA55AA55), ABC_CONST(0x00AA00AA00AA00AA), ABC_CONST(0x5500550055005500) }, // 0 3 
        { ABC_CONST(0xAAAA5555AAAA5555), ABC_CONST(0x0000AAAA0000AAAA), ABC_CONST(0x5555000055550000) }, // 0 4 
        { ABC_CONST(0xAAAAAAAA55555555), ABC_CONST(0x00000000AAAAAAAA), ABC_CONST(0x5555555500000000) }  // 0 5 
    },
    { 
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 1 0  
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 1 1  
        { ABC_CONST(0xC3C3C3C3C3C3C3C3), ABC_CONST(0x0C0C0C0C0C0C0C0C), ABC_CONST(0x3030303030303030) }, // 1 2 
        { ABC_CONST(0xCC33CC33CC33CC33), ABC_CONST(0x00CC00CC00CC00CC), ABC_CONST(0x3300330033003300) }, // 1 3 
        { ABC_CONST(0xCCCC3333CCCC3333), ABC_CONST(0x0000CCCC0000CCCC), ABC_CONST(0x3333000033330000) }, // 1 4 
        { ABC_CONST(0xCCCCCCCC33333333), ABC_CONST(0x00000000CCCCCCCC), ABC_CONST(0x3333333300000000) }  // 1 5 
    },
    { 
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 2 0  
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 2 1  
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 2 2 
        { ABC_CONST(0xF00FF00FF00FF00F), ABC_CONST(0x00F000F000F000F0), ABC_CONST(0x0F000F000F000F00) }, // 2 3 
        { ABC_CONST(0xF0F00F0FF0F00F0F), ABC_CONST(0x0000F0F00000F0F0), ABC_CONST(0x0F0F00000F0F0000) }, // 2 4 
        { ABC_CONST(0xF0F0F0F00F0F0F0F), ABC_CONST(0x00000000F0F0F0F0), ABC_CONST(0x0F0F0F0F00000000) }  // 2 5 
    },
    { 
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 3 0  
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 3 1  
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 3 2 
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 3 3 
        { ABC_CONST(0xFF0000FFFF0000FF), ABC_CONST(0x0000FF000000FF00), ABC_CONST(0x00FF000000FF0000) }, // 3 4 
        { ABC_CONST(0xFF00FF0000FF00FF), ABC_CONST(0x00000000FF00FF00), ABC_CONST(0x00FF00FF00000000) }  // 3 5 
    },
    { 
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 4 0  
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 4 1  
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 4 2 
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 4 3 
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000) }, // 4 4 
        { ABC_CONST(0xFFFF00000000FFFF), ABC_CONST(0x00000000FFFF0000), ABC_CONST(0x0000FFFF00000000) }  // 4 5 
    }
};

// the bit count for the first 256 integer numbers
var Abc_TtBitCount8 [256]int32 = [256]int32{
    0,1,1,2,1,2,2,3,1,2,2,3,2,3,3,4,1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,
    1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
    1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
    2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
    1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
    2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
    2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
    3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,4,5,5,6,5,6,6,7,5,6,6,7,6,7,7,8,
};
func Abc_TtBitCount16(i int32) int32 { return Abc_TtBitCount8[i & 0xFF] + Abc_TtBitCount8[i >> 8]; }

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// read/write/flip i-th bit of a bit string table:
func Abc_TtGetBit(p *word, i int32) int32 { return int32((p[i>>6] >> (word)(i & 63))) & 1;        }
func Abc_TtSetBit(p *word, i int32) { p[i>>6] |= (word)(((word)1)<<(i & 63));             }
func Abc_TtXorBit(p *word, i int32) { p[i>>6] ^= (word)(((word)1)<<(i & 63));             }

// read/write k-th digit d of a quaternary number:
func Abc_TtGetQua(p *word, k int32) int32 { return int32((p[k>>5] >> (word)((k<<1) & 63))) & 3;   }
func Abc_TtSetQua(p *word, k int32, d int32) { p[k>>5] |= (word)(((word)d)<<((k<<1) & 63));        }
func Abc_TtXorQua(p *word, k int32, d int32) { p[k>>5] ^= (word)(((word)d)<<((k<<1) & 63));        }

// read/write k-th digit d of a hexadecimal number:
func Abc_TtGetHex(p *word, k int32) int32 { return int32((p[k>>4] >> (word)((k<<2) & 63))) & 15;  }
func Abc_TtSetHex(p *word, k int32, d int32) { p[k>>4] |= (word)(((word)d)<<((k<<2) & 63));        }
func Abc_TtXorHex(p *word, k int32, d int32) { p[k>>4] ^= (word)(((word)d)<<((k<<2) & 63));        }

// read/write k-th digit d of a 256-base number:
func Abc_TtGet256(p *word, k int32) int32 { return int32((p[k>>3] >> (word)((k<<3) & 63))) & 255; }
func Abc_TtSet256(p *word, k int32, d int32) { p[k>>3] |= (word)(((word)d)<<((k<<3) & 63));        }
func Abc_TtXor256(p *word, k int32, d int32) { p[k>>3] ^= (word)(((word)d)<<((k<<3) & 63));        }

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtWordNum(nVars int32) int32 { if nVars <= 6 {
    return 1;
} else {
    return 1 << (nVars-6);
} }
func Abc_TtByteNum(nVars int32) int32 { if nVars <= 3 {
    return 1;
} else {
    return 1 << (nVars-3);
} }
func Abc_TtHexDigitNum(nVars int32) int32 { if nVars <= 2 {
    return 1;
} else {
    return 1 << (nVars-2);
} }

/**Function*************************************************************

  Synopsis    [Bit mask.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6Mask(nBits int32) word { assert(nBits >= 0 && nBits <= 64, "nBits >= 0 && nBits <= 64"); return (~(word)0) >> (64-nBits);        }
func Abc_TtMask(pTruth *word, nWords int32, nBits int32) { 
    var w int32;
    assert(nBits >= 0 && nBits <= nWords * 64, "nBits >= 0 && nBits <= nWords * 64");
    for  w = 0; w < nWords; w++  {
        if nBits >= (w + 1) * 64 {
            pTruth[w] = ~(word)0;
        } else if nBits > w * 64 {
            pTruth[w] = Abc_Tt6Mask( nBits - w * 64 );
        } else {
            pTruth[w] = 0; }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtVec(pOut *word, nWords int32, Entry word) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] = Entry;
    }
}
func Abc_TtConst(pOut *word, nWords int32, fConst1 int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] = fConst1 ? ~(word)0 : 0;
    }
}
func Abc_TtClear(pOut *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] = 0;
    }
}
func Abc_TtFill(pOut *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] = ~(word)0;
    }
}
func Abc_TtUnit(pOut *word, nWords int32, fCompl int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] = fCompl ? ~s_Truths6[0] : s_Truths6[0];
    }
}
func Abc_TtNot(pOut *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] = ~pOut[w];
    }
}
func Abc_TtCopy(pOut *word, pIn *word, nWords int32, fCompl int32) {
    var w int32;
    if fCompl != 0 {
        for  w = 0; w < nWords; w++  {
            pOut[w] = ~pIn[w];
        }
    } else {
        for  w = 0; w < nWords; w++  {
            pOut[w] = pIn[w];
        } }
}
func Abc_TtAnd(pOut *word, pIn1 *word, pIn2 *word, nWords int32, fCompl int32) {
    var w int32;
    if fCompl != 0 {
        for  w = 0; w < nWords; w++  {
            pOut[w] = ~(pIn1[w] & pIn2[w]);
        }
    } else {
        for  w = 0; w < nWords; w++  {
            pOut[w] = pIn1[w] & pIn2[w];
        } }
}
func Abc_TtAndCompl(pOut *word, pIn1 *word, fCompl1 int32, pIn2 *word, fCompl2 int32, nWords int32) {
    var w int32;
    if fCompl1 != 0 {
        if fCompl2 != 0 {
            for  w = 0; w < nWords; w++  {
                pOut[w] = ~pIn1[w] & ~pIn2[w];
            }
        } else {
            for  w = 0; w < nWords; w++  {
                pOut[w] = ~pIn1[w] & pIn2[w];
            } }
    
    } else {
        if fCompl2 != 0 {
            for  w = 0; w < nWords; w++  {
                pOut[w] = pIn1[w] & ~pIn2[w];
            }
        } else {
            for  w = 0; w < nWords; w++  {
                pOut[w] = pIn1[w] & pIn2[w];
            } }
    }
}
func Abc_TtAndSharp(pOut *word, pIn1 *word, pIn2 *word, nWords int32, fCompl int32) {
    var w int32;
    if fCompl != 0 {
        for  w = 0; w < nWords; w++  {
            pOut[w] = pIn1[w] & ~pIn2[w];
        }
    } else {
        for  w = 0; w < nWords; w++  {
            pOut[w] = pIn1[w] & pIn2[w];
        } }
}
func Abc_TtSharp(pOut *word, pIn1 *word, pIn2 *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] = pIn1[w] & ~pIn2[w];
    }
}
func Abc_TtOr(pOut *word, pIn1 *word, pIn2 *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] = pIn1[w] | pIn2[w];
    }
}
func Abc_TtOrXor(pOut *word, pIn1 *word, pIn2 *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] |= pIn1[w] ^ pIn2[w];
    }
}
func Abc_TtOrAnd(pOut *word, pIn1 *word, pIn2 *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] |= pIn1[w] & pIn2[w];
    }
}
func Abc_TtXor(pOut *word, pIn1 *word, pIn2 *word, nWords int32, fCompl int32) {
    var w int32;
    if fCompl != 0 {
        for  w = 0; w < nWords; w++  {
            pOut[w] = pIn1[w] ^ ~pIn2[w];
        }
    } else {
        for  w = 0; w < nWords; w++  {
            pOut[w] = pIn1[w] ^ pIn2[w];
        } }
}
func Abc_TtXorMask(pOut *word, pIn1 *word, pIn2 *word, pMask *word, nWords int32, fCompl int32) {
    var w int32;
    if fCompl != 0 {
        for  w = 0; w < nWords; w++  {
            pOut[w] = (pIn1[w] ^ pIn2[w]) & ~pMask[w];
        }
    } else {
        for  w = 0; w < nWords; w++  {
            pOut[w] = (pIn1[w] ^ pIn2[w]) & pMask[w];
        } }
}
func Abc_TtMux(pOut *word, pCtrl *word, pIn1 *word, pIn0 *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] = (pCtrl[w] & pIn1[w]) | (~pCtrl[w] & pIn0[w]);
    }
}
func Abc_TtMaj(pOut *word, pIn0 *word, pIn1 *word, pIn2 *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] = (pIn0[w] & pIn1[w]) | (pIn0[w] & pIn2[w]) | (pIn1[w] & pIn2[w]);
    }
}
func Abc_TtIntersect(pIn1 *word, pIn2 *word, nWords int32, fCompl int32) int32 {
    var w int32;
    if fCompl != 0 {
        for  w = 0; w < nWords; w++  {
            if ~pIn1[w] & pIn2[w] {
                return 1; }
        }
    
    } else {
        for  w = 0; w < nWords; w++  {
            if pIn1[w] & pIn2[w] {
                return 1; }
        }
    }
    return 0;
}
func Abc_TtIntersectOne(pOut *word, fComp int32, pIn *word, fComp0 int32, nWords int32) int32 {
    var w int32;
    if fComp0 != 0 {
        if fComp != 0 {
            for  w = 0; w < nWords; w++  {
                if ~pIn[w] & ~pOut[w] {
                    return 1; }
            }
        
        } else {
            for  w = 0; w < nWords; w++  {
                if ~pIn[w] & pOut[w] {
                    return 1; }
            }
        }
    
    } else {
        if fComp != 0 {
            for  w = 0; w < nWords; w++  {
                if pIn[w] & ~pOut[w] {
                    return 1; }
            }
        
        } else {
            for  w = 0; w < nWords; w++  {
                if pIn[w] & pOut[w] {
                    return 1; }
            }
        }
    }
    return 0;
}
func Abc_TtIntersectTwo(pOut *word, fComp int32, pIn0 *word, fComp0 int32, pIn1 *word, fComp1 int32, nWords int32) int32 {
    var w int32;
    if (fComp0 != 0) && (fComp1 != 0) {
        if fComp != 0 {
            for  w = 0; w < nWords; w++  {
                if ~pIn0[w] & ~pIn1[w] & ~pOut[w] {
                    return 1; }
            }
        
        } else {
            for  w = 0; w < nWords; w++  {
                if ~pIn0[w] & ~pIn1[w] & pOut[w] {
                    return 1; }
            }
        }
    
    } else if fComp0 != 0 {
        if fComp != 0 {
            for  w = 0; w < nWords; w++  {
                if ~pIn0[w] & pIn1[w] & ~pOut[w] {
                    return 1; }
            }
        
        } else {
            for  w = 0; w < nWords; w++  {
                if ~pIn0[w] & pIn1[w] & pOut[w] {
                    return 1; }
            }
        }
    
    } else if fComp1 != 0 {
        if fComp != 0 {
            for  w = 0; w < nWords; w++  {
                if pIn0[w] & ~pIn1[w] & ~pOut[w] {
                    return 1; }
            }
        
        } else {
            for  w = 0; w < nWords; w++  {
                if pIn0[w] & ~pIn1[w] & pOut[w] {
                    return 1; }
            }
        }
    
    } else {
        if fComp != 0 {
            for  w = 0; w < nWords; w++  {
                if pIn0[w] & pIn1[w] & ~pOut[w] {
                    return 1; }
            }
        
        } else {
            for  w = 0; w < nWords; w++  {
                if pIn0[w] & pIn1[w] & pOut[w] {
                    return 1; }
            }
        }
    }
    return 0;
}
func Abc_TtIntersectXor(pOut *word, fComp int32, pIn0 *word, pIn1 *word, fComp01 int32, nWords int32) int32 {
    var w int32;
    if fComp01 != 0 {
        if fComp != 0 {
            for  w = 0; w < nWords; w++  {
                if ~(pIn0[w] ^ pIn1[w]) & ~pOut[w] {
                    return 1; }
            }
        
        } else {
            for  w = 0; w < nWords; w++  {
                if ~(pIn0[w] ^ pIn1[w]) & pOut[w] {
                    return 1; }
            }
        }
    
    } else {
        if fComp != 0 {
            for  w = 0; w < nWords; w++  {
                if (pIn0[w] ^ pIn1[w]) & ~pOut[w] {
                    return 1; }
            }
        
        } else {
            for  w = 0; w < nWords; w++  {
                if (pIn0[w] ^ pIn1[w]) & pOut[w] {
                    return 1; }
            }
        }
    }
    return 0;
}
func Abc_TtEqual(pIn1 *word, pIn2 *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if pIn1[w] != pIn2[w] {
            return 0; }
    }
    return 1;
}
func Abc_TtOpposite(pIn1 *word, pIn2 *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if pIn1[w] != ~pIn2[w] {
            return 0; }
    }
    return 1;
}
func Abc_TtImply(pIn1 *word, pIn2 *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if (pIn1[w] & pIn2[w]) != pIn1[w] {
            return 0; }
    }
    return 1;
}
func Abc_TtCompare(pIn1 *word, pIn2 *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if pIn1[w] != pIn2[w] {
            if pIn1[w] < pIn2[w] {
                return -1;
            } else {
                return 1;
            } }
    }
    return 0;
}
func Abc_TtCompareRev(pIn1 *word, pIn2 *word, nWords int32) int32 {
    var w int32;
    for  w = nWords - 1; w >= 0; w--  {
        if pIn1[w] != pIn2[w] {
            if pIn1[w] < pIn2[w] {
                return -1;
            } else {
                return 1;
            } }
    }
    return 0;
}
func Abc_TtIsConst0(pIn1 *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if pIn1[w] {
            return 0; }
    }
    return 1;
}
func Abc_TtIsConst1(pIn1 *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if ~pIn1[w] {
            return 0; }
    }
    return 1;
}
func Abc_TtConst0(pIn1 *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pIn1[w] = 0;
    }
}
func Abc_TtConst1(pIn1 *word, nWords int32) {
    var w int32;
    for  w = 0; w < nWords; w++  {
        pIn1[w] = ~(word)0;
    }
}
func Abc_TtIthVar(pOut *word, iVar int32, nVars int32) {
    var k int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    if iVar < 6 {
        for  k = 0; k < nWords; k++  {
            pOut[k] = s_Truths6[iVar];
        }
    
    } else {
        for  k = 0; k < nWords; k++  {
            if k & (1 << (iVar-6)) {
                pOut[k] = ~(word)0;
            } else {
                pOut[k] = 0; }
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtIsAndCompl(pOut *word, fCompl int32, pIn1 *word, fCompl1 int32, pIn2 *word, fCompl2 int32, pCare *word, nWords int32) int32 {
    var w int32;
    if fCompl != 0 {
        if fCompl1 != 0 {
            if fCompl2 != 0 {
                for  w = 0; w < nWords; w++  {
                    if (~pOut[w] & pCare[w]) != (~pIn1[w] & ~pIn2[w] & pCare[w]) {
                        return 0; }
                }
            
            } else {
                for  w = 0; w < nWords; w++  {
                    if (~pOut[w] & pCare[w]) != (~pIn1[w] & pIn2[w] & pCare[w]) {
                        return 0; }
                }
            }
        
        } else {
            if fCompl2 != 0 {
                for  w = 0; w < nWords; w++  {
                    if (~pOut[w] & pCare[w]) != (pIn1[w] & ~pIn2[w] & pCare[w]) {
                        return 0; }
                }
            
            } else {
                for  w = 0; w < nWords; w++  {
                    if (~pOut[w] & pCare[w]) != (pIn1[w] & pIn2[w] & pCare[w]) {
                        return 0; }
                }
            }
        }
    
    } else {
        if fCompl1 != 0 {
            if fCompl2 != 0 {
                for  w = 0; w < nWords; w++  {
                    if (pOut[w] & pCare[w]) != (~pIn1[w] & ~pIn2[w] & pCare[w]) {
                        return 0; }
                }
            
            } else {
                for  w = 0; w < nWords; w++  {
                    if (pOut[w] & pCare[w]) != (~pIn1[w] & pIn2[w] & pCare[w]) {
                        return 0; }
                }
            }
        
        } else {
            if fCompl2 != 0 {
                for  w = 0; w < nWords; w++  {
                    if (pOut[w] & pCare[w]) != (pIn1[w] & ~pIn2[w] & pCare[w]) {
                        return 0; }
                }
            
            } else {
                for  w = 0; w < nWords; w++  {
                    if (pOut[w] & pCare[w]) != (pIn1[w] & pIn2[w] & pCare[w]) {
                        return 0; }
                }
            }
        }
    }
    return 1;
}

func Abc_TtIsXorCompl(pOut *word, fCompl int32, pIn1 *word, pIn2 *word, pCare *word, nWords int32) int32 {
    var w int32;
    if fCompl != 0 {
        for  w = 0; w < nWords; w++  {
            if (~pOut[w] & pCare[w]) != ((pIn1[w] ^ pIn2[w]) & pCare[w]) {
                return 0; }
        }
    
    } else {
        for  w = 0; w < nWords; w++  {
            if ( pOut[w] & pCare[w]) != ((pIn1[w] ^ pIn2[w]) & pCare[w]) {
                return 0; }
        }
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Compares Cof0 and Cof1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtCompare1VarCofs(pTruth *word, nWords int32, iVar int32) int32 {
    if nWords == 1 {
        var Cof0 word  = pTruth[0] & s_Truths6Neg[iVar];
        var Cof1 word  = (pTruth[0] >> (1 << iVar)) & s_Truths6Neg[iVar];
        if Cof0 != Cof1 {
            if Cof0 < Cof1 {
                return -1;
            } else {
                return 1;
            } }
        return 0;
    }
    if iVar <= 5 {
        var Cof0, Cof1 word;
        var w int32;
        var shift int32 = (1 << iVar);
        for  w = 0; w < nWords; w++  {
            Cof0 = pTruth[w] & s_Truths6Neg[iVar];
            Cof1 = (pTruth[w] >> shift) & s_Truths6Neg[iVar];
            if Cof0 != Cof1 {
                if Cof0 < Cof1 {
                    return -1;
                } else {
                    return 1;
                } }
        }
        return 0;
    }
    // if ( iVar > 5 )
    {
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        assert(nWords >= 2, "nWords >= 2");
        for  ; pTruth < pLimit; pTruth += 2*iStep  {
            for  i = 0; i < iStep; i++  {
                if pTruth[i] != pTruth[i + iStep] {
                    if pTruth[i] < pTruth[i + iStep] {
                        return -1;
                    } else {
                        return 1;
                    } }
            }
        }
        return 0;
    }    
}
func Abc_TtCompare1VarCofsRev(pTruth *word, nWords int32, iVar int32) int32 {
    if nWords == 1 {
        var Cof0 word  = pTruth[0] & s_Truths6Neg[iVar];
        var Cof1 word  = (pTruth[0] >> (1 << iVar)) & s_Truths6Neg[iVar];
        if Cof0 != Cof1 {
            if Cof0 < Cof1 {
                return -1;
            } else {
                return 1;
            } }
        return 0;
    }
    if iVar <= 5 {
        var Cof0, Cof1 word;
        var w int32;
        var shift int32 = (1 << iVar);
        for  w = nWords - 1; w >= 0; w--  {
            Cof0 = pTruth[w] & s_Truths6Neg[iVar];
            Cof1 = (pTruth[w] >> shift) & s_Truths6Neg[iVar];
            if Cof0 != Cof1 {
                if Cof0 < Cof1 {
                    return -1;
                } else {
                    return 1;
                } }
        }
        return 0;
    }
    // if ( iVar > 5 )
    {
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        assert(nWords >= 2, "nWords >= 2");
        for  pLimit -= 2*iStep; pLimit >= pTruth; pLimit -= 2*iStep  {
            for  i = iStep - 1; i >= 0; i--  {
                if pLimit[i] != pLimit[i + iStep] {
                    if pLimit[i] < pLimit[i + iStep] {
                        return -1;
                    } else {
                        return 1;
                    } }
            }
        }
        return 0;
    }    
}

/**Function*************************************************************

  Synopsis    [Compute elementary truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtElemInit(pTtElems **word, nVars int32) {
    var i int32;
    var k int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    for  i = 0; i < nVars; i++  {
        if i < 6 {
            for  k = 0; k < nWords; k++  {
                pTtElems[i][k] = s_Truths6[i];
            }
        } else {
            for  k = 0; k < nWords; k++  {
                pTtElems[i][k] = (k & (1 << (i-6))) ? ~(word)0 : 0;
            } }
    }
}
func Abc_TtElemInit2(pTtElems *word, nVars int32) {
    var i int32;
    var k int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    for  i = 0; i < nVars; i++  {
        var pTruth *word  = pTtElems + i * nWords;
        if i < 6 {
            for  k = 0; k < nWords; k++  {
                pTruth[k] = s_Truths6[i];
            }
        } else {
            for  k = 0; k < nWords; k++  {
                pTruth[k] = (k & (1 << (i-6))) ? ~(word)0 : 0;
            } }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6Cofactor0(t word, iVar int32) word {
    assert(iVar >= 0 && iVar < 6, "iVar >= 0 && iVar < 6");
    return (t &s_Truths6Neg[iVar]) | ((t &s_Truths6Neg[iVar]) << (1<<iVar));
}
func Abc_Tt6Cofactor1(t word, iVar int32) word {
    assert(iVar >= 0 && iVar < 6, "iVar >= 0 && iVar < 6");
    return (t & s_Truths6[iVar]) | ((t & s_Truths6[iVar]) >> (1<<iVar));
}

func Abc_TtCofactor0p(pOut *word, pIn *word, nWords int32, iVar int32) {
    if nWords == 1 {
        pOut[0] = ((pIn[0] & s_Truths6Neg[iVar]) << (1 << iVar)) | (pIn[0] & s_Truths6Neg[iVar]);
    } else if iVar <= 5 {
        var w int32;
        var shift int32 = (1 << iVar);
        for  w = 0; w < nWords; w++  {
            pOut[w] = ((pIn[w] & s_Truths6Neg[iVar]) << shift) | (pIn[w] & s_Truths6Neg[iVar]);
        }
    
    } else { // if ( iVar > 5 )
    
        var pLimit *word  = pIn + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        for  ; pIn < pLimit; pIn += 2*iStep, pOut += 2*iStep  {
            for  i = 0; i < iStep; i++  {
                pOut[i]         = pIn[i];
                pOut[i + iStep] = pIn[i];
            }
        }
    }    
}
func Abc_TtCofactor1p(pOut *word, pIn *word, nWords int32, iVar int32) {
    if nWords == 1 {
        pOut[0] = (pIn[0] & s_Truths6[iVar]) | ((pIn[0] & s_Truths6[iVar]) >> (1 << iVar));
    } else if iVar <= 5 {
        var w int32;
        var shift int32 = (1 << iVar);
        for  w = 0; w < nWords; w++  {
            pOut[w] = (pIn[w] & s_Truths6[iVar]) | ((pIn[w] & s_Truths6[iVar]) >> shift);
        }
    
    } else { // if ( iVar > 5 )
    
        var pLimit *word  = pIn + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        for  ; pIn < pLimit; pIn += 2*iStep, pOut += 2*iStep  {
            for  i = 0; i < iStep; i++  {
                pOut[i]         = pIn[i + iStep];
                pOut[i + iStep] = pIn[i + iStep];
            }
        }
    }    
}
func Abc_TtCofactor0(pTruth *word, nWords int32, iVar int32) {
    if nWords == 1 {
        pTruth[0] = ((pTruth[0] & s_Truths6Neg[iVar]) << (1 << iVar)) | (pTruth[0] & s_Truths6Neg[iVar]);
    } else if iVar <= 5 {
        var w int32;
        var shift int32 = (1 << iVar);
        for  w = 0; w < nWords; w++  {
            pTruth[w] = ((pTruth[w] & s_Truths6Neg[iVar]) << shift) | (pTruth[w] & s_Truths6Neg[iVar]);
        }
    
    } else { // if ( iVar > 5 )
    
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        for  ; pTruth < pLimit; pTruth += 2*iStep  {
            for  i = 0; i < iStep; i++  {
                pTruth[i + iStep] = pTruth[i];
            }
        }
    }
}
func Abc_TtCofactor1(pTruth *word, nWords int32, iVar int32) {
    if nWords == 1 {
        pTruth[0] = (pTruth[0] & s_Truths6[iVar]) | ((pTruth[0] & s_Truths6[iVar]) >> (1 << iVar));
    } else if iVar <= 5 {
        var w int32;
        var shift int32 = (1 << iVar);
        for  w = 0; w < nWords; w++  {
            pTruth[w] = (pTruth[w] & s_Truths6[iVar]) | ((pTruth[w] & s_Truths6[iVar]) >> shift);
        }
    
    } else { // if ( iVar > 5 )
    
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        for  ; pTruth < pLimit; pTruth += 2*iStep  {
            for  i = 0; i < iStep; i++  {
                pTruth[i] = pTruth[i + iStep];
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Checks pairs of cofactors w.r.t. two variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtCheckEqualCofs(pTruth *word, nWords int32, iVar int32, jVar int32, Num1 int32, Num2 int32) int32 {
    assert(Num1 < Num2 && Num2 < 4, "Num1 < Num2 && Num2 < 4");
    assert(iVar < jVar, "iVar < jVar");
    if nWords == 1 {
        var Mask word  = s_Truths6Neg[jVar] & s_Truths6Neg[iVar];
        var shift1 int32  = (Num1 >> 1) * (1 << jVar) + (Num1 & 1) * (1 << iVar);
        var shift2 int32  = (Num2 >> 1) * (1 << jVar) + (Num2 & 1) * (1 << iVar);
        return ((pTruth[0] >> shift1) & Mask) == ((pTruth[0] >> shift2) & Mask);
    }
    if jVar <= 5 {
        var Mask word  = s_Truths6Neg[jVar] & s_Truths6Neg[iVar];
        var shift1 int32  = (Num1 >> 1) * (1 << jVar) + (Num1 & 1) * (1 << iVar);
        var shift2 int32  = (Num2 >> 1) * (1 << jVar) + (Num2 & 1) * (1 << iVar);
        var w int32;
        for  w = 0; w < nWords; w++  {
            if ((pTruth[w] >> shift1) & Mask) != ((pTruth[w] >> shift2) & Mask) {
                return 0; }
        }
        return 1;
    }
    if iVar <= 5 && jVar > 5 {
        var pLimit *word  = pTruth + nWords;
        var j int32;
        var jStep int32 = Abc_TtWordNum(jVar);
        var shift1 int32  = (Num1 & 1) * (1 << iVar);
        var shift2 int32  = (Num2 & 1) * (1 << iVar);
        var Offset1 int32  = (Num1 >> 1) * jStep;
        var Offset2 int32  = (Num2 >> 1) * jStep;
        for  ; pTruth < pLimit; pTruth += 2*jStep  {
            for  j = 0; j < jStep; j++  {
                if ((pTruth[j + Offset1] >> shift1) & s_Truths6Neg[iVar]) != ((pTruth[j + Offset2] >> shift2) & s_Truths6Neg[iVar]) {
                    return 0; }
            }
        }
        return 1;
    }
    {
        var pLimit *word  = pTruth + nWords;
        var j int32;
        var jStep int32 = Abc_TtWordNum(jVar);
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        var Offset1 int32  = (Num1 >> 1) * jStep + (Num1 & 1) * iStep;
        var Offset2 int32  = (Num2 >> 1) * jStep + (Num2 & 1) * iStep;
        for  ; pTruth < pLimit; pTruth += 2*jStep  {
            for  i = 0; i < jStep; i += 2*iStep  {
                for  j = 0; j < iStep; j++  {
                    if pTruth[Offset1 + i + j] != pTruth[Offset2 + i + j] {
                        return 0; }
                }
            }
        }
        return 1;
    }    
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6Cof0IsConst0(t word, iVar int32) int32 { return (t & s_Truths6Neg[iVar]) == 0;                                          }
func Abc_Tt6Cof0IsConst1(t word, iVar int32) int32 { return (t & s_Truths6Neg[iVar]) == s_Truths6Neg[iVar];                         }
func Abc_Tt6Cof1IsConst0(t word, iVar int32) int32 { return (t & s_Truths6[iVar]) == 0;                                             }
func Abc_Tt6Cof1IsConst1(t word, iVar int32) int32 { return (t & s_Truths6[iVar]) == s_Truths6[iVar];                               }
func Abc_Tt6CofsOpposite(t word, iVar int32) int32 { return (~t & s_Truths6Neg[iVar]) == ((t >> (1 << iVar)) & s_Truths6Neg[iVar]); } 
func Abc_Tt6Cof0EqualCof1(t1 word, t2 word, iVar int32) int32 { return (t1 & s_Truths6Neg[iVar]) == ((t2 >> (1 << iVar)) & s_Truths6Neg[iVar]); } 
func Abc_Tt6Cof0EqualCof0(t1 word, t2 word, iVar int32) int32 { return (t1 & s_Truths6Neg[iVar]) == (t2 & s_Truths6Neg[iVar]); } 
func Abc_Tt6Cof1EqualCof1(t1 word, t2 word, iVar int32) int32 { return (t1 & s_Truths6[iVar])    == (t2 & s_Truths6[iVar]); } 

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtTruthIsConst0(p *word, nWords int32) int32 { var w int32; for  w = 0; w < nWords; w++  { if p[w] != 0 { return 0; } }
 return 1; }
func Abc_TtTruthIsConst1(p *word, nWords int32) int32 { var w int32; for  w = 0; w < nWords; w++  { if p[w] != ~(word)0 { return 0; } }
 return 1; }

func Abc_TtCof0IsConst0(t *word, nWords int32, iVar int32) int32 { 
    if iVar < 6 {
        var i int32;
        for  i = 0; i < nWords; i++  {
            if t[i] & s_Truths6Neg[iVar] {
                return 0; }
        }
        return 1;
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + nWords;
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                if t[i] {
                    return 0; }
            }
        }
        return 1;
    }
}
func Abc_TtCof0IsConst1(t *word, nWords int32, iVar int32) int32 { 
    if iVar < 6 {
        var i int32;
        for  i = 0; i < nWords; i++  {
            if (t[i] & s_Truths6Neg[iVar]) != s_Truths6Neg[iVar] {
                return 0; }
        }
        return 1;
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + nWords;
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                if ~t[i] {
                    return 0; }
            }
        }
        return 1;
    }
}
func Abc_TtCof1IsConst0(t *word, nWords int32, iVar int32) int32 { 
    if iVar < 6 {
        var i int32;
        for  i = 0; i < nWords; i++  {
            if t[i] & s_Truths6[iVar] {
                return 0; }
        }
        return 1;
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + nWords;
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                if t[i+Step] {
                    return 0; }
            }
        }
        return 1;
    }
}
func Abc_TtCof1IsConst1(t *word, nWords int32, iVar int32) int32 { 
    if iVar < 6 {
        var i int32;
        for  i = 0; i < nWords; i++  {
            if (t[i] & s_Truths6[iVar]) != s_Truths6[iVar] {
                return 0; }
        }
        return 1;
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + nWords;
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                if ~t[i+Step] {
                    return 0; }
            }
        }
        return 1;
    }
}
func Abc_TtCofsOpposite(t *word, nWords int32, iVar int32) int32 { 
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            if ((t[i] << Shift) & s_Truths6[iVar]) != (~t[i] & s_Truths6[iVar]) {
                return 0; }
        }
        return 1;
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + nWords;
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                if t[i] != ~t[i+Step] {
                    return 0; }
            }
        }
        return 1;
    }
}

/**Function*************************************************************

  Synopsis    [Stretch truthtable to have more input variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtStretch5(pInOut *uint32, nVarS int32, nVarB int32) {
    var w, i, step, nWords int32;
    if nVarS == nVarB {
        return; }
    assert(nVarS < nVarB, "nVarS < nVarB");
    step = Abc_TruthWordNum(nVarS);
    nWords = Abc_TruthWordNum(nVarB);
    if step == nWords {
        return; }
    assert(step < nWords, "step < nWords");
    for  w = 0; w < nWords; w += step  {
        for  i = 0; i < step; i++  {
            pInOut[w + i] = pInOut[i];
        }
    }              
}
func Abc_TtStretch6(pInOut *word, nVarS int32, nVarB int32) {
    var w, i, step, nWords int32;
    if nVarS == nVarB {
        return; }
    assert(nVarS < nVarB, "nVarS < nVarB");
    step = Abc_Truth6WordNum(nVarS);
    nWords = Abc_Truth6WordNum(nVarB);
    if step == nWords {
        return; }
    assert(step < nWords, "step < nWords");
    for  w = 0; w < nWords; w += step  {
        for  i = 0; i < step; i++  {
            pInOut[w + i] = pInOut[i];
        }
    }              
}
func Abc_Tt6Stretch(t word, nVars int32) word {
    assert(nVars >= 0, "nVars >= 0");
    if nVars == 0 {
        nVars++, t = (t & 0x1) | ((t & 0x1) << 1); }
    if nVars == 1 {
        nVars++, t = (t & 0x3) | ((t & 0x3) << 2); }
    if nVars == 2 {
        nVars++, t = (t & 0xF) | ((t & 0xF) << 4); }
    if nVars == 3 {
        nVars++, t = (t & 0xFF) | ((t & 0xFF) << 8); }
    if nVars == 4 {
        nVars++, t = (t & 0xFFFF) | ((t & 0xFFFF) << 16); }
    if nVars == 5 {
        nVars++, t = (t & 0xFFFFFFFF) | ((t & 0xFFFFFFFF) << 32); }
    assert(nVars == 6, "nVars == 6");
    return t;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtIsHexDigit(HexChar byte) int32 {
    return (HexChar >= '0' && HexChar <= '9') || (HexChar >= 'A' && HexChar <= 'F') || (HexChar >= 'a' && HexChar <= 'f');
}
func Abc_TtPrintDigit(Digit int32) byte {
    assert(Digit >= 0 && Digit < 16, "Digit >= 0 && Digit < 16");
    if Digit < 10 {
        return '0' + Digit; }
    return 'A' + Digit-10;
}
func Abc_TtPrintDigitLower(Digit int32) byte {
    assert(Digit >= 0 && Digit < 16, "Digit >= 0 && Digit < 16");
    if Digit < 10 {
        return '0' + Digit; }
    return 'a' + Digit-10;
}
func Abc_TtReadHexDigit(HexChar byte) int32 {
    if HexChar >= '0' && HexChar <= '9' {
        return HexChar - '0'; }
    if HexChar >= 'A' && HexChar <= 'F' {
        return HexChar - 'A' + 10; }
    if HexChar >= 'a' && HexChar <= 'f' {
        return HexChar - 'a' + 10; }
    assert(false, "0"); // not a hexadecimal symbol
    return -1; // return value which makes no sense
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtPrintHex(pTruth *word, nVars int32) {
    var pThis *word;
    var pLimit *word = pTruth + Abc_TtWordNum(nVars);
    var k int32;
    assert(nVars >= 2, "nVars >= 2");
    for  pThis = pTruth; pThis < pLimit; pThis++  {
        for  k = 0; k < 16; k++  {
            fmt.Printf("%c", Abc_TtPrintDigit(int32((pThis[0] >> (k << 2))) & 15));
        }
    }
    fmt.Printf("\n");
}
func Abc_TtPrintHexRev(pFile *os.File, pTruth *word, nVars int32) {
    var pThis *word;
    var k int32;
    var StartK int32;
    if nVars >= 6 {
        StartK = 16;
    } else {
        StartK = (1 << (nVars - 2));
    }
    assert(nVars >= 2, "nVars >= 2");
    for  pThis = pTruth + Abc_TtWordNum(nVars) - 1; pThis >= pTruth; pThis--  {
        for  k = StartK - 1; k >= 0; k--  {
            fmt.Fprintf(pFile, "%c", Abc_TtPrintDigit(int32((pThis[0] >> (k << 2))) & 15));
        }
    }
//    printf( "\n" );
}
func Abc_TtPrintHexSpecial(pTruth *word, nVars int32) {
    var pThis *word;
    var k int32;
    assert(nVars >= 2, "nVars >= 2");
    for  pThis = pTruth + Abc_TtWordNum(nVars) - 1; pThis >= pTruth; pThis--  {
        for  k = 0; k < 16; k++  {
            fmt.Printf("%c", Abc_TtPrintDigit(int32((pThis[0] >> (k << 2))) & 15));
        }
    }
    fmt.Printf("\n");
}
func Abc_TtWriteHexRev(pStr string, pTruth *word, nVars int32) int32 {
    var pThis *word;
    var pStrInit string  = pStr;
    var k int32;
    var StartK int32;
    if nVars >= 6 {
        StartK = 16;
    } else {
        StartK = (1 << (nVars - 2));
    }
    assert(nVars >= 2, "nVars >= 2");
    for  pThis = pTruth + Abc_TtWordNum(nVars) - 1; pThis >= pTruth; pThis--  {
        for  k = StartK - 1; k >= 0; k--  {
            *pStr++ = Abc_TtPrintDigit( int32((pThis[0] >> (k << 2))) & 15 );
        }
    }
    return pStr - pStrInit;
}
func Abc_TtPrintHexArrayRev(pFile *os.File, pTruth *word, nDigits int32) {
    var k int32;
    for  k = nDigits - 1; k >= 0; k--  {
        fmt.Fprintf(pFile, "%c", Abc_TtPrintDigitLower( Abc_TtGetHex(pTruth, k) ));
    }
}

/**Function*************************************************************

  Synopsis    [Reads hex truth table from a string.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtReadHex(pTruth *word, pString string) int32 {
    var k, nVars, Digit, nDigits int32;
    // skip the first 2 symbols if they are "0x"
    if pString[0] == '0' && pString[1] == 'x' {
        pString += 2; }
    // count the number of hex digits
    nDigits = 0;
    for  k = 0; (Abc_TtIsHexDigit(pString[k]) != 0); k++  {
        nDigits++;
    }
    if nDigits == 1 {
        if pString[0] == '0' || pString[0] == 'F' {
            pTruth[0] = (pString[0] == '0') ? 0 : ~(word)0;
            return 0;
        }
        if pString[0] == '5' || pString[0] == 'A' {
            pTruth[0] = (pString[0] == '5') ? s_Truths6Neg[0] : s_Truths6[0];
            return 1;
        }
    }
    // determine the number of variables
    nVars = 2 + (nDigits == 1 ? 0 : Abc_Base2Log(nDigits));
    // clean storage
    for  k = Abc_TtWordNum(nVars) - 1; k >= 0; k--  {
        pTruth[k] = 0;
    }
    // read hexadecimal digits in the reverse order
    // (the last symbol in the string is the least significant digit)
    for  k = 0; k < nDigits; k++  {
        Digit = Abc_TtReadHexDigit( pString[nDigits - 1 - k] );
        assert(Digit >= 0 && Digit < 16, "Digit >= 0 && Digit < 16");
        Abc_TtSetHex( pTruth, k, Digit );
    }
    if nVars < 6 {
        pTruth[0] = Abc_Tt6Stretch( pTruth[0], nVars ); }
    return nVars;
}
func Abc_TtReadHexNumber(pTruth *word, pString string) int32 {
    // count the number of hex digits
    var k, Digit, nDigits int32;
    for  k = 0; (Abc_TtIsHexDigit(pString[k]) != 0); k++  {
        nDigits++;
    }
    // read hexadecimal digits in the reverse order
    // (the last symbol in the string is the least significant digit)
    for  k = 0; k < nDigits; k++  {
        Digit = Abc_TtReadHexDigit( pString[nDigits - 1 - k] );
        assert(Digit >= 0 && Digit < 16, "Digit >= 0 && Digit < 16");
        Abc_TtSetHex( pTruth, k, Digit );
    }
    return nDigits;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtPrintBinary(pTruth *word, nVars int32) {
    var pThis *word;
    var pLimit *word = pTruth + Abc_TtWordNum(nVars);
    var k int32;
    var Limit int32 = Abc_MinInt( 64, (1 << nVars) );
    assert(nVars >= 2, "nVars >= 2");
    for  pThis = pTruth; pThis < pLimit; pThis++  {
        for  k = 0; k < Limit; k++  {
            fmt.Printf("%d", Abc_InfoHasBit( (unsigned *)pThis, k ));
        }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtSuppFindFirst(Supp int32) int32 {
    var i int32;
    assert(Supp > 0, "Supp > 0");
    for  i = 0; i < 32; i++  {
        if Supp & (1 << i) {
            return i; }
    }
    return -1;
}
func Abc_TtSuppOnlyOne(Supp int32) int32 {
    if Supp == 0 {
        return 0; }
    return (Supp & (Supp-1)) == 0;
}
func Abc_TtSuppIsMinBase(Supp int32) int32 {
    assert(Supp > 0, "Supp > 0");
    return (Supp & (Supp+1)) == 0;
}
func Abc_Tt6HasVar(t word, iVar int32) int32 {
    return ((t >> (1<<iVar)) & s_Truths6Neg[iVar]) != (t & s_Truths6Neg[iVar]);
}
func Abc_Tt6XorVar(t word, iVar int32) int32 {
    return ((t >> (1<<iVar)) & s_Truths6Neg[iVar]) == ~(t & s_Truths6Neg[iVar]);
}
func Abc_TtHasVar(t *word, nVars int32, iVar int32) int32 {
    assert(iVar < nVars, "iVar < nVars");
    if nVars <= 6 {
        return Abc_Tt6HasVar( t[0], iVar ); }
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        var nWords int32  = Abc_TtWordNum( nVars );
        for  i = 0; i < nWords; i++  {
            if ((t[i] >> Shift) & s_Truths6Neg[iVar]) != (t[i] & s_Truths6Neg[iVar]) {
                return 1; }
        }
        return 0;
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + Abc_TtWordNum( nVars );
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                if t[i] != t[Step+i] {
                    return 1; }
            }
        }
        return 0;
    }
}
func Abc_TtSupport(t *word, nVars int32) int32 {
    var v, Supp int32;
    for  v = 0; v < nVars; v++  {
        if Abc_TtHasVar( t, nVars, v ) != 0 {
            Supp |= (1 << v); }
    }
    return Supp;
}
func Abc_TtSupportSize(t *word, nVars int32) int32 {
    var v, SuppSize int32;
    for  v = 0; v < nVars; v++  {
        if Abc_TtHasVar( t, nVars, v ) != 0 {
            SuppSize++; }
    }
    return SuppSize;
}
func Abc_TtSupportAndSize(t *word, nVars int32, pSuppSize *int32) int32 {
    var v, Supp int32;
    *pSuppSize = 0;
    for  v = 0; v < nVars; v++  {
        if Abc_TtHasVar( t, nVars, v ) != 0 {
            Supp |= (1 << v), (*pSuppSize)++; }
    }
    return Supp;
}
func Abc_Tt6SupportAndSize(t word, nVars int32, pSuppSize *int32) int32 {
    var v, Supp int32;
    *pSuppSize = 0;
    assert(nVars <= 6, "nVars <= 6");
    for  v = 0; v < nVars; v++  {
        if Abc_Tt6HasVar( t, v ) != 0 {
            Supp |= (1 << v), (*pSuppSize)++; }
    }
    return Supp;
}

/**Function*************************************************************

  Synopsis    [Checks if there is a var whose both cofs have supp <= nSuppLim.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtCheckCondDep2(pTruth *word, nVars int32, nSuppLim int32) int32 {
    var v int32;
    var d int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    if nVars <= nSuppLim + 1 {
        return 0; }
    for  v = 0; v < nVars; v++  {
        var nDep0, nDep1 int32;
        for  d = 0; d < nVars; d++  {
            if v == d {
                continue; }
            if v < d {
                nDep0 += !Abc_TtCheckEqualCofs( pTruth, nWords, v, d, 0, 2 );
                nDep1 += !Abc_TtCheckEqualCofs( pTruth, nWords, v, d, 1, 3 );
            
            } else { // if ( v > d )
            
                nDep0 += !Abc_TtCheckEqualCofs( pTruth, nWords, d, v, 0, 1 );
                nDep1 += !Abc_TtCheckEqualCofs( pTruth, nWords, d, v, 2, 3 );
            }
            if nDep0 > nSuppLim || nDep1 > nSuppLim {
                break; }
        }
        if d == nVars {
            return v; }
    }
    return nVars;
}
func Abc_TtCheckCondDep(pTruth *word, nVars int32, nSuppLim int32) int32 {
    var nVarsMax int32  = 13;
    var Cof0, Cof1 [128]word; // pow( 2, nVarsMax-6 )
    var v int32;
    var d int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    assert(nVars <= nVarsMax, "nVars <= nVarsMax");
    if nVars <= nSuppLim + 1 {
        return 0; }
    for  v = 0; v < nVars; v++  {
        var nDep0, nDep1 int32;
        Abc_TtCofactor0p( Cof0, pTruth, nWords, v );
        Abc_TtCofactor1p( Cof1, pTruth, nWords, v );
        for  d = 0; d < nVars; d++  {
            if v == d {
                continue; }
            nDep0 += Abc_TtHasVar( Cof0, nVars, d );
            nDep1 += Abc_TtHasVar( Cof1, nVars, d );
            if nDep0 > nSuppLim || nDep1 > nSuppLim {
                break; }
        }
        if d == nVars {
            return v; }
    }
    return nVars;
}


/**Function*************************************************************

  Synopsis    [Detecting elementary functions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtOnlyOneOne(t word) int32 {
    if t == 0 {
        return 0; }
    return (t & (t-1)) == 0;
}
func Abc_Tt6IsAndType(t word, nVars int32) int32 {
    return Abc_TtOnlyOneOne( t & Abc_Tt6Mask(1 << nVars) );
}
func Abc_Tt6IsOrType(t word, nVars int32) int32 {
    return Abc_TtOnlyOneOne( ~t & Abc_Tt6Mask(1 << nVars) );
}
func Abc_Tt6IsXorType(t word, nVars int32) int32 {
    return ((((t & 1) ? ~t : t) ^ s_TruthXors[nVars]) & Abc_Tt6Mask(1 << nVars)) == 0;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6Flip(Truth word, iVar int32) word {
    return Truth = ((Truth << (1 << iVar)) & s_Truths6[iVar]) | ((Truth & s_Truths6[iVar]) >> (1 << iVar));
}
func Abc_TtFlip(pTruth *word, nWords int32, iVar int32) {
    if nWords == 1 {
        pTruth[0] = ((pTruth[0] << (1 << iVar)) & s_Truths6[iVar]) | ((pTruth[0] & s_Truths6[iVar]) >> (1 << iVar));
    } else if iVar <= 5 {
        var w int32;
        var shift int32 = (1 << iVar);
        for  w = 0; w < nWords; w++  {
            pTruth[w] = ((pTruth[w] << shift) & s_Truths6[iVar]) | ((pTruth[w] & s_Truths6[iVar]) >> shift);
        }
    
    } else { // if ( iVar > 5 )
    
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        for  ; pTruth < pLimit; pTruth += 2*iStep  {
            for  i = 0; i < iStep; i++  {
                ABC_SWAP( word, pTruth[i], pTruth[i + iStep] );
            }
        }
    }    
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6Permute_rec(t word, pPerm *int32, nVars int32) word {
    var uRes0, uRes1 word; var Var int32;
    if t == 0 { return 0; }
    if ~t == 0 { return ~(word)0; }
    for  Var = nVars-1; Var >= 0; Var--  {
        if Abc_Tt6HasVar( t, Var ) != 0 {
             break; }
    }
    assert(Var >= 0, "Var >= 0");
    uRes0 = Abc_Tt6Permute_rec( Abc_Tt6Cofactor0(t, Var), pPerm, Var );
    uRes1 = Abc_Tt6Permute_rec( Abc_Tt6Cofactor1(t, Var), pPerm, Var );
    return (uRes0 & s_Truths6Neg[pPerm[Var]]) | (uRes1 & s_Truths6[pPerm[Var]]);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6SwapAdjacent(Truth word, iVar int32) word {
    return (Truth & s_PMasks[iVar][0]) | ((Truth & s_PMasks[iVar][1]) << (1 << iVar)) | ((Truth & s_PMasks[iVar][2]) >> (1 << iVar));
}
func Abc_TtSwapAdjacent(pTruth *word, nWords int32, iVar int32) {
    if iVar < 5 {
        var i int32;
        var Shift int32 = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            pTruth[i] = (pTruth[i] & s_PMasks[iVar][0]) | ((pTruth[i] & s_PMasks[iVar][1]) << Shift) | ((pTruth[i] & s_PMasks[iVar][2]) >> Shift);
        }
    
    } else if iVar == 5 {
        var pTruthU *uint32  = (unsigned *)pTruth;
        var pLimitU *uint32  = (unsigned *)(pTruth + nWords);
        for  ; pTruthU < pLimitU; pTruthU += 4  {
            ABC_SWAP( unsigned, pTruthU[1], pTruthU[2] );
        }
    
    } else { // if ( iVar > 5 )
    
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        for  ; pTruth < pLimit; pTruth += 4*iStep  {
            for  i = 0; i < iStep; i++  {
                ABC_SWAP( word, pTruth[i + iStep], pTruth[i + 2*iStep] );
            }
        }
    }
}
func Abc_Tt6SwapVars(t word, iVar int32, jVar int32) word {
    var s_PMasks *word  = s_PPMasks[iVar][jVar];
    var shift int32  = (1 << jVar) - (1 << iVar);
    assert(iVar < jVar, "iVar < jVar");
    return (t & s_PMasks[0]) | ((t & s_PMasks[1]) << shift) | ((t & s_PMasks[2]) >> shift);
}
func Abc_TtSwapVars(pTruth *word, nVars int32, iVar int32, jVar int32) {
    if iVar == jVar {
        return; }
    if jVar < iVar {
        ABC_SWAP( int, iVar, jVar ); }
    assert(iVar < jVar && jVar < nVars, "iVar < jVar && jVar < nVars");
    if nVars <= 6 {
        pTruth[0] = Abc_Tt6SwapVars( pTruth[0], iVar, jVar );
        return;
    }
    if jVar <= 5 {
        var s_PMasks *word  = s_PPMasks[iVar][jVar];
        var nWords int32  = Abc_TtWordNum(nVars);
        var w int32;
        var shift int32 = (1 << jVar) - (1 << iVar);
        for  w = 0; w < nWords; w++  {
            pTruth[w] = (pTruth[w] & s_PMasks[0]) | ((pTruth[w] & s_PMasks[1]) << shift) | ((pTruth[w] & s_PMasks[2]) >> shift);
        }
        return;
    }
    if iVar <= 5 && jVar > 5 {
        var low2High, high2Low word;
        var pLimit *word  = pTruth + Abc_TtWordNum(nVars);
        var j int32;
        var jStep int32 = Abc_TtWordNum(jVar);
        var shift int32  = 1 << iVar;
        for  ; pTruth < pLimit; pTruth += 2*jStep  {
            for  j = 0; j < jStep; j++  {
                low2High = (pTruth[j] & s_Truths6[iVar]) >> shift;
                high2Low = (pTruth[j+jStep] << shift) & s_Truths6[iVar];
                pTruth[j] = (pTruth[j] & ~s_Truths6[iVar]) | high2Low;
                pTruth[j+jStep] = (pTruth[j+jStep] & s_Truths6[iVar]) | low2High;
            }
        }
        return;
    }
    {
        var pLimit *word  = pTruth + Abc_TtWordNum(nVars);
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        var j int32;
        var jStep int32 = Abc_TtWordNum(jVar);
        for  ; pTruth < pLimit; pTruth += 2*jStep  {
            for  i = 0; i < jStep; i += 2*iStep  {
                for  j = 0; j < iStep; j++  {
                    ABC_SWAP( word, pTruth[iStep + i + j], pTruth[jStep + i + j] );
                }
            }
        }
        return;
    }    
}
// moves one var (v) to the given position (p)
func Abc_TtMoveVar(pF *word, nVars int32, V2P *int32, P2V *int32, v int32, p int32) {
    var iVar int32 = V2P[v];
    var jVar int32 = p;
    if iVar == jVar {
        return; }
    Abc_TtSwapVars( pF, nVars, iVar, jVar );
    V2P[P2V[iVar]] = jVar;
    V2P[P2V[jVar]] = iVar;
    P2V[iVar] ^= P2V[jVar];
    P2V[jVar] ^= P2V[iVar];
    P2V[iVar] ^= P2V[jVar];
}
func Abc_Tt6RemoveVar(t word, iVar int32) word {
    assert(!(Abc_Tt6HasVar(t, iVar) != 0), "!Abc_Tt6HasVar(t, iVar)");
    for  iVar < 5  {
        t = Abc_Tt6SwapAdjacent( t, iVar++ );
    }
    return t;
}

/**Function*************************************************************

  Synopsis    [Support minimization.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtShrink(pF *word, nVars int32, nVarsAll int32, Phase uint32) {
    var i, k, Var int32;
    assert(nVarsAll <= 16, "nVarsAll <= 16");
    for  i = 0; i < nVarsAll; i++  {
        if Phase & (1 << i) {
            for  k = i-1; k >= Var; k--  {
                Abc_TtSwapAdjacent( pF, Abc_TtWordNum(nVarsAll), k );
            }
            Var++;
        }
    }
    assert(Var == nVars, "Var == nVars");
}
func Abc_TtMinimumBase(t *word, pSupp *int32, nVarsAll int32, pnVars *int32) int32 {
    var v, iVar, uSupp int32;
    assert(nVarsAll <= 16, "nVarsAll <= 16");
    for  v = 0; v < nVarsAll; v++  {
        if Abc_TtHasVar( t, nVarsAll, v ) != 0 {
            uSupp |= (1 << v);
            if pSupp != nil {
                pSupp[iVar] = pSupp[v]; }
            iVar++;
        }
    }
    if pnVars != nil {
        *pnVars = iVar; }
    if uSupp == 0 || (Abc_TtSuppIsMinBase( uSupp ) != 0) {
        return 0; }
    Abc_TtShrink( t, iVar, nVarsAll, uSupp );
    return 1;
}
func Abc_TtSimplify(t *word, pLits *int32, nVarsAll int32, pnVars *int32) int32 {
    var v, u int32;
    for  v = 0; v < nVarsAll; v++  {
        if pLits[v] == 0 {
            Abc_TtCofactor0( t, Abc_TtWordNum(nVarsAll), v );
        } else if pLits[v] == 1 {
            Abc_TtCofactor1( t, Abc_TtWordNum(nVarsAll), v ); }
    }
    for  v = 0;   v < nVarsAll; v++  {
    for  u = v+1; u < nVarsAll; u++  {
        if Abc_Lit2Var(pLits[v]) == Abc_Lit2Var(pLits[u]) {
            assert(nVarsAll <= 6, "nVarsAll <= 6");
            if pLits[v] == pLits[u] {
                var t0 word  = Abc_Tt6Cofactor0(Abc_Tt6Cofactor0(*t, v), u);
                var t1 word  = Abc_Tt6Cofactor1(Abc_Tt6Cofactor1(*t, v), u);
                *t = (t0 & s_Truths6Neg[v]) | (t1 & s_Truths6[v]);
            
            } else { // if ( pLits[v] == Abc_LitNot(pLits[u]) )
            
                var t0 word  = Abc_Tt6Cofactor1(Abc_Tt6Cofactor0(*t, v), u);
                var t1 word  = Abc_Tt6Cofactor0(Abc_Tt6Cofactor1(*t, v), u);
                *t = (t0 & s_Truths6Neg[v]) | (t1 & s_Truths6[v]);
            }
        }
    }
    }
    return Abc_TtMinimumBase( t, pLits, nVarsAll, pnVars );
}

/**Function*************************************************************

  Synopsis    [Cut minimization.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6Expand(t word, pCut0 *int32, nCutSize0 int32, pCut *int32, nCutSize int32) word {
    var i, k int32;
    for  i, k = nCutSize - 1, nCutSize0 - 1; i >= 0 && k >= 0; i--  {
        if pCut[i] > pCut0[k] {
            continue; }
        assert(pCut[i] == pCut0[k], "pCut[i] == pCut0[k]");
        if k < i {
            t = Abc_Tt6SwapVars( t, k, i ); }
        k--;
    }
    assert(k == -1, "k == -1");
    return t;
}
func Abc_TtExpand(pTruth0 *word, nVars int32, pCut0 *int32, nCutSize0 int32, pCut *int32, nCutSize int32) {
    var i, k int32;
    for  i, k = nCutSize - 1, nCutSize0 - 1; i >= 0 && k >= 0; i--  {
        if pCut[i] > pCut0[k] {
            continue; }
        assert(pCut[i] == pCut0[k], "pCut[i] == pCut0[k]");
        if k < i {
            Abc_TtSwapVars( pTruth0, nVars, k, i ); }
        k--;
    }
    assert(k == -1, "k == -1");
}
func Abc_Tt6MinBase(pTruth *word, pVars *int32, nVars int32) int32 {
    var t word  = *pTruth;
    var i, k int32;
    for  i = k = 0; i < nVars; i++  {
        if !(Abc_Tt6HasVar( t, i ) != 0) {
            continue; }
        if k < i {
            if pVars != nil { pVars[k] = pVars[i]; }
            t = Abc_Tt6SwapVars( t, k, i );
        }
        k++;
    }
    if k == nVars {
        return k; }
    assert(k < nVars, "k < nVars");
    *pTruth = t;
    return k;
}
func Abc_TtMinBase(pTruth *word, pVars *int32, nVars int32, nVarsAll int32) int32 {
    var i, k int32;
    assert(nVars <= nVarsAll, "nVars <= nVarsAll");
    for  i = k = 0; i < nVars; i++  {
        if !(Abc_TtHasVar( pTruth, nVarsAll, i ) != 0) {
            continue; }
        if k < i {
            if pVars != nil { pVars[k] = pVars[i]; }
            Abc_TtSwapVars( pTruth, nVarsAll, k, i );
        }
        k++;
    }
    if k == nVars {
        return k; }
    assert(k < nVars, "k < nVars");
//    assert( k == Abc_TtSupportSize(pTruth, nVars) );
    return k;
}

/**Function*************************************************************

  Synopsis    [Implemeting given NPN config.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtImplementNpnConfig(pTruth *word, nVars int32, pCanonPerm string, uCanonPhase uint32) {
    var i int32;
    var k int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    if (uCanonPhase >> nVars) & 1 {
        Abc_TtNot( pTruth, nWords ); }
    for  i = 0; i < nVars; i++  {
        if (uCanonPhase >> i) & 1 {
            Abc_TtFlip( pTruth, nWords, i ); }
    }
    if pCanonPerm {
    for  i = 0; i < nVars; i++  {
        for  k = i; k < nVars; k++  {
            if pCanonPerm[k] == i {
                break; }
        }
        assert(k < nVars, "k < nVars");
        if i == k {
            continue; }
        Abc_TtSwapVars( pTruth, nVars, i, k );
        ABC_SWAP( int, pCanonPerm[i], pCanonPerm[k] );
    } }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtCountOnesSlow(t word) int32 {
    t =    (t & ABC_CONST(0x5555555555555555)) + ((t>> 1) & ABC_CONST(0x5555555555555555));
    t =    (t & ABC_CONST(0x3333333333333333)) + ((t>> 2) & ABC_CONST(0x3333333333333333));
    t =    (t & ABC_CONST(0x0F0F0F0F0F0F0F0F)) + ((t>> 4) & ABC_CONST(0x0F0F0F0F0F0F0F0F));
    t =    (t & ABC_CONST(0x00FF00FF00FF00FF)) + ((t>> 8) & ABC_CONST(0x00FF00FF00FF00FF));
    t =    (t & ABC_CONST(0x0000FFFF0000FFFF)) + ((t>>16) & ABC_CONST(0x0000FFFF0000FFFF));
    return (t & ABC_CONST(0x00000000FFFFFFFF)) +  (t>>32);
}
func Abc_TtCountOnes(x word) int32 {
    x = x - ((x >> 1) & ABC_CONST(0x5555555555555555));   
    x = (x & ABC_CONST(0x3333333333333333)) + ((x >> 2) & ABC_CONST(0x3333333333333333));    
    x = (x + (x >> 4)) & ABC_CONST(0x0F0F0F0F0F0F0F0F);    
    x = x + (x >> 8);
    x = x + (x >> 16);
    x = x + (x >> 32); 
    return int32((x & 0xFF));
}
func Abc_TtCountOnesVec(x *word, nWords int32) int32 {
    var w, Count int32;
    for  w = 0; w < nWords; w++  {
        Count += Abc_TtCountOnes( x[w] );
    }
    return Count;
}
func Abc_TtCountOnesVecMask(x *word, pMask *word, nWords int32, fCompl int32) int32 {
    var w, Count int32;
    if fCompl != 0 {
        for  w = 0; w < nWords; w++  {
            Count += Abc_TtCountOnes( pMask[w] & ~x[w] );
        }
    } else {
        for  w = 0; w < nWords; w++  {
            Count += Abc_TtCountOnes( pMask[w] & x[w] );
        } }
    return Count;
}
func Abc_TtCountOnesVecMask2(x0 *word, x1 *word, fComp0 int32, fComp1 int32, pMask *word, nWords int32) int32 {
    var w, Count int32;
    if !(fComp0 != 0) && !(fComp1 != 0) {
        for  w = 0; w < nWords; w++  {
            Count += Abc_TtCountOnes( pMask[w] &  x0[w] &  x1[w] );
        }
    } else if (fComp0 != 0) && !(fComp1 != 0) {
        for  w = 0; w < nWords; w++  {
            Count += Abc_TtCountOnes( pMask[w] & ~x0[w] &  x1[w] );
        }
    } else if !(fComp0 != 0) &&  (fComp1 != 0) {
        for  w = 0; w < nWords; w++  {
            Count += Abc_TtCountOnes( pMask[w] &  x0[w] & ~x1[w] );
        }
    } else { 
        for  w = 0; w < nWords; w++  {
            Count += Abc_TtCountOnes( pMask[w] & ~x0[w] & ~x1[w] );
        } }
    return Count;
}
func Abc_TtCountOnesVecXor(x *word, y *word, nWords int32) int32 {
    var w, Count int32;
    for  w = 0; w < nWords; w++  {
        Count += Abc_TtCountOnes( x[w] ^ y[w] );
    }
    return Count;
}
func Abc_TtCountOnesVecXorMask(x *word, y *word, fCompl int32, pMask *word, nWords int32) int32 {
    var w, Count int32;
    if fCompl != 0 {
        for  w = 0; w < nWords; w++  {
            Count += Abc_TtCountOnes( pMask[w] & (x[w] ^ ~y[w]) );
        }
    } else {
        for  w = 0; w < nWords; w++  {
            Count += Abc_TtCountOnes( pMask[w] & (x[w] ^ y[w]) );
        } }
    return Count;
}
func Abc_TtAndXorSum(pOut *word, pIn1 *word, pIn2 *word, nWords int32) int32 {
    var w, Count int32;
    for  w = 0; w < nWords; w++  {
        pOut[w] &= pIn1[w] ^ pIn2[w];
        Count += Abc_TtCountOnes( pOut[w] );
    }
    return Count;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6FirstBit(t word) int32 {
    var n int32;
    if t == 0 { return -1; }
    if (t & ABC_CONST(0x00000000FFFFFFFF)) == 0 { n += 32; t >>= 32; }
    if (t & ABC_CONST(0x000000000000FFFF)) == 0 { n += 16; t >>= 16; }
    if (t & ABC_CONST(0x00000000000000FF)) == 0 { n +=  8; t >>=  8; }
    if (t & ABC_CONST(0x000000000000000F)) == 0 { n +=  4; t >>=  4; }
    if (t & ABC_CONST(0x0000000000000003)) == 0 { n +=  2; t >>=  2; }
    if (t & ABC_CONST(0x0000000000000001)) == 0 { n++; }
    return n;
}
func Abc_Tt6LastBit(t word) int32 {
    var n int32;
    if t == 0 { return -1; }
    if (t & ABC_CONST(0xFFFFFFFF00000000)) == 0 { n += 32; t <<= 32; }
    if (t & ABC_CONST(0xFFFF000000000000)) == 0 { n += 16; t <<= 16; }
    if (t & ABC_CONST(0xFF00000000000000)) == 0 { n +=  8; t <<=  8; }
    if (t & ABC_CONST(0xF000000000000000)) == 0 { n +=  4; t <<=  4; }
    if (t & ABC_CONST(0xC000000000000000)) == 0 { n +=  2; t <<=  2; }
    if (t & ABC_CONST(0x8000000000000000)) == 0 { n++; }
    return 63-n;
}
func Abc_TtFindFirstBit(pIn *word, nVars int32) int32 {
    var w int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    for  w = 0; w < nWords; w++  {
        if pIn[w] {
            return 64*w + Abc_Tt6FirstBit(pIn[w]); }
    }
    return -1;
}
func Abc_TtFindFirstBit2(pIn *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if pIn[w] {
            return 64*w + Abc_Tt6FirstBit(pIn[w]); }
    }
    return -1;
}
func Abc_TtFindLastBit(pIn *word, nVars int32) int32 {
    var w int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    for  w = nWords - 1; w >= 0; w--  {
        if pIn[w] {
            return 64*w + Abc_Tt6LastBit(pIn[w]); }
    }
    return -1;
}
func Abc_TtFindLastBit2(pIn *word, nWords int32) int32 {
    var w int32;
    for  w = nWords - 1; w >= 0; w--  {
        if pIn[w] {
            return 64*w + Abc_Tt6LastBit(pIn[w]); }
    }
    return -1;
}
func Abc_TtFindFirstDiffBit(pIn1 *word, pIn2 *word, nVars int32) int32 {
    var w int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    for  w = 0; w < nWords; w++  {
        if pIn1[w] ^ pIn2[w] {
            return 64*w + Abc_Tt6FirstBit(pIn1[w] ^ pIn2[w]); }
    }
    return -1;
}
func Abc_TtFindFirstDiffBit2(pIn1 *word, pIn2 *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if pIn1[w] ^ pIn2[w] {
            return 64*w + Abc_Tt6FirstBit(pIn1[w] ^ pIn2[w]); }
    }
    return -1;
}
func Abc_TtFindLastDiffBit(pIn1 *word, pIn2 *word, nVars int32) int32 {
    var w int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    for  w = nWords - 1; w >= 0; w--  {
        if pIn1[w] ^ pIn2[w] {
            return 64*w + Abc_Tt6LastBit(pIn1[w] ^ pIn2[w]); }
    }
    return -1;
}
func Abc_TtFindLastDiffBit2(pIn1 *word, pIn2 *word, nWords int32) int32 {
    var w int32;
    for  w = nWords - 1; w >= 0; w--  {
        if pIn1[w] ^ pIn2[w] {
            return 64*w + Abc_Tt6LastBit(pIn1[w] ^ pIn2[w]); }
    }
    return -1;
}
func Abc_TtFindFirstZero(pIn *word, nVars int32) int32 {
    var w int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    for  w = 0; w < nWords; w++  {
        if ~pIn[w] {
            return 64*w + Abc_Tt6FirstBit(~pIn[w]); }
    }
    return -1;
}
func Abc_TtFindLastZero(pIn *word, nVars int32) int32 {
    var w int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    for  w = nWords - 1; w >= 0; w--  {
        if ~pIn[w] {
            return 64*w + Abc_Tt6LastBit(~pIn[w]); }
    }
    return -1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtReverseVars(pTruth *word, nVars int32) {
    var k int32;
    for  k = 0; k < nVars/2 ; k++  {
        Abc_TtSwapVars( pTruth, nVars, k, nVars - 1 - k );
    }
}
func Abc_TtReverseBits(pTruth *word, nVars int32) {
    static unsigned char pMirror[256] = {
          0, 128,  64, 192,  32, 160,  96, 224,  16, 144,  80, 208,  48, 176, 112, 240,
          8, 136,  72, 200,  40, 168, 104, 232,  24, 152,  88, 216,  56, 184, 120, 248,
          4, 132,  68, 196,  36, 164, 100, 228,  20, 148,  84, 212,  52, 180, 116, 244,
         12, 140,  76, 204,  44, 172, 108, 236,  28, 156,  92, 220,  60, 188, 124, 252,
          2, 130,  66, 194,  34, 162,  98, 226,  18, 146,  82, 210,  50, 178, 114, 242,
         10, 138,  74, 202,  42, 170, 106, 234,  26, 154,  90, 218,  58, 186, 122, 250,
          6, 134,  70, 198,  38, 166, 102, 230,  22, 150,  86, 214,  54, 182, 118, 246,
         14, 142,  78, 206,  46, 174, 110, 238,  30, 158,  94, 222,  62, 190, 126, 254,
          1, 129,  65, 193,  33, 161,  97, 225,  17, 145,  81, 209,  49, 177, 113, 241,
          9, 137,  73, 201,  41, 169, 105, 233,  25, 153,  89, 217,  57, 185, 121, 249,
          5, 133,  69, 197,  37, 165, 101, 229,  21, 149,  85, 213,  53, 181, 117, 245,
         13, 141,  77, 205,  45, 173, 109, 237,  29, 157,  93, 221,  61, 189, 125, 253,
          3, 131,  67, 195,  35, 163,  99, 227,  19, 147,  83, 211,  51, 179, 115, 243,
         11, 139,  75, 203,  43, 171, 107, 235,  27, 155,  91, 219,  59, 187, 123, 251,
          7, 135,  71, 199,  39, 167, 103, 231,  23, 151,  87, 215,  55, 183, 119, 247,
         15, 143,  79, 207,  47, 175, 111, 239,  31, 159,  95, 223,  63, 191, 127, 255
    };
    unsigned char Temp, * pTruthC = (unsigned char *)pTruth;
    var i int32;
    var nBytes int32;
    if nVars > 6 {
        nBytes = (1 << (nVars - 3));
    } else {
        nBytes = 8;
    }
    for  i = 0; i < nBytes/2; i++  {
        Temp = pMirror[pTruthC[i]];
        pTruthC[i] = pMirror[pTruthC[nBytes-1-i]];
        pTruthC[nBytes-1-i] = Temp;
    }
}


/**Function*************************************************************

  Synopsis    [Checks unateness of a function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6PosVar(t word, iVar int32) int32 {
    return ((t >> (1<<iVar)) & t & s_Truths6Neg[iVar]) == (t & s_Truths6Neg[iVar]);
}
func Abc_Tt6NegVar(t word, iVar int32) int32 {
    return ((t << (1<<iVar)) & t & s_Truths6[iVar]) == (t & s_Truths6[iVar]);
}
func Abc_TtPosVar(t *word, nVars int32, iVar int32) int32 {
    assert(iVar < nVars, "iVar < nVars");
    if nVars <= 6 {
        return Abc_Tt6PosVar( t[0], iVar ); }
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        var nWords int32  = Abc_TtWordNum( nVars );
        for  i = 0; i < nWords; i++  {
            if ((t[i] >> Shift) & t[i] & s_Truths6Neg[iVar]) != (t[i] & s_Truths6Neg[iVar]) {
                return 0; }
        }
        return 1;
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + Abc_TtWordNum( nVars );
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                if t[i] != (t[i] & t[Step+i]) {
                    return 0; }
            }
        }
        return 1;
    }
}
func Abc_TtNegVar(t *word, nVars int32, iVar int32) int32 {
    assert(iVar < nVars, "iVar < nVars");
    if nVars <= 6 {
        return Abc_Tt6NegVar( t[0], iVar ); }
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        var nWords int32  = Abc_TtWordNum( nVars );
        for  i = 0; i < nWords; i++  {
            if ((t[i] << Shift) & t[i] & s_Truths6[iVar]) != (t[i] & s_Truths6[iVar]) {
                return 0; }
        }
        return 1;
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + Abc_TtWordNum( nVars );
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                if (t[i] & t[Step+i]) != t[Step+i] {
                    return 0; }
            }
        }
        return 1;
    }
}
func Abc_TtIsUnate(t *word, nVars int32) int32 {
    var i int32;
    for  i = 0; i < nVars; i++  {
        if !(Abc_TtNegVar(t, nVars, i) != 0) && !(Abc_TtPosVar(t, nVars, i) != 0) {
            return 0; }
    }
    return 1;
}
func Abc_TtIsPosUnate(t *word, nVars int32) int32 {
    var i int32;
    for  i = 0; i < nVars; i++  {
        if !(Abc_TtPosVar(t, nVars, i) != 0) {
            return 0; }
    }
    return 1;
}
func Abc_TtMakePosUnate(t *word, nVars int32) {
    var i int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    for  i = 0; i < nVars; i++  {
        if Abc_TtNegVar(t, nVars, i) != 0 {
            Abc_TtFlip( t, nWords, i );
        } else { assert((Abc_TtPosVar(t, nVars, i) != 0), "Abc_TtPosVar(t, nVars, i)"); }
    }
}


/**Function*************************************************************

  Synopsis    [Computes ISOP for 6 variables or less.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6Isop(uOn word, uOnDc word, nVars int32, pnCubes *int32) word {
    var uOn0, uOn1, uOnDc0, uOnDc1, uRes0, uRes1, uRes2 word;
    var Var int32;
    assert(nVars <= 6, "nVars <= 6");
    assert((uOn & ~uOnDc) == 0, "(uOn & ~uOnDc) == 0");
    if uOn == 0 {
        return 0; }
    if uOnDc == ~(word)0 {
        (*pnCubes)++;
        return ~(word)0;
    }
    assert(nVars > 0, "nVars > 0");
    // find the topmost var
    for  Var = nVars-1; Var >= 0; Var--  {
        if (Abc_Tt6HasVar( uOn, Var ) != 0) || (Abc_Tt6HasVar( uOnDc, Var ) != 0) {
             break; }
    }
    assert(Var >= 0, "Var >= 0");
    // cofactor
    uOn0   = Abc_Tt6Cofactor0( uOn,   Var );
    uOn1   = Abc_Tt6Cofactor1( uOn  , Var );
    uOnDc0 = Abc_Tt6Cofactor0( uOnDc, Var );
    uOnDc1 = Abc_Tt6Cofactor1( uOnDc, Var );
    // solve for cofactors
    uRes0 = Abc_Tt6Isop( uOn0 & ~uOnDc1, uOnDc0, Var, pnCubes );
    uRes1 = Abc_Tt6Isop( uOn1 & ~uOnDc0, uOnDc1, Var, pnCubes );
    uRes2 = Abc_Tt6Isop( (uOn0 & ~uRes0) | (uOn1 & ~uRes1), uOnDc0 & uOnDc1, Var, pnCubes );
    // derive the final truth table
    uRes2 |= (uRes0 & s_Truths6Neg[Var]) | (uRes1 & s_Truths6[Var]);
    assert((uOn & ~uRes2) == 0, "(uOn & ~uRes2) == 0");
    assert((uRes2 & ~uOnDc) == 0, "(uRes2 & ~uOnDc) == 0");
    return uRes2;
}
func Abc_Tt7Isop(uOn []word, uOnDc []word, nVars int32, uRes []word) int32 {
    var nCubes int32;
    if nVars <= 6 || (uOn[0] == uOn[1] && uOnDc[0] == uOnDc[1]) {
        uRes[0] = uRes[1] = Abc_Tt6Isop( uOn[0], uOnDc[0], Abc_MinInt(nVars, 6), &nCubes );
    } else {
        var uRes0, uRes1, uRes2 word;
        assert(nVars == 7, "nVars == 7");
        // solve for cofactors
        uRes0 = Abc_Tt6Isop( uOn[0] & ~uOnDc[1], uOnDc[0], 6, &nCubes );
        uRes1 = Abc_Tt6Isop( uOn[1] & ~uOnDc[0], uOnDc[1], 6, &nCubes );
        uRes2 = Abc_Tt6Isop( (uOn[0] & ~uRes0) | (uOn[1] & ~uRes1), uOnDc[0] & uOnDc[1], 6, &nCubes );
        // derive the final truth table
        uRes[0] = uRes2 | uRes0;
        uRes[1] = uRes2 | uRes1;
        assert((uOn[0] & ~uRes[0]) == 0 && (uOn[1] & ~uRes[1]) == 0, "(uOn[0] & ~uRes[0]) == 0 && (uOn[1] & ~uRes[1]) == 0");
        assert((uRes[0] & ~uOnDc[0])==0 && (uRes[1] & ~uOnDc[1])==0, "(uRes[0] & ~uOnDc[0])==0 && (uRes[1] & ~uOnDc[1])==0");
    }
    return nCubes;
}
func Abc_Tt8Isop(uOn []word, uOnDc []word, nVars int32, uRes []word) int32 {
    var nCubes int32;
    if nVars <= 6 {
        uRes[0] = uRes[1] = uRes[2] = uRes[3] = Abc_Tt6Isop( uOn[0], uOnDc[0], nVars, &nCubes );
    } else if nVars == 7 || (uOn[0] == uOn[2] && uOn[1] == uOn[3] && uOnDc[0] == uOnDc[2] && uOnDc[1] == uOnDc[3]) {
        nCubes = Abc_Tt7Isop( uOn, uOnDc, 7, uRes );
        uRes[2] = uRes[0];
        uRes[3] = uRes[1];
    
    } else {
        var uOn0, uOn1, uOn2, uOnDc2, uRes0, uRes1, uRes2 [2]word;
        assert(nVars == 8, "nVars == 8");
        // cofactor
        uOn0[0] = uOn[0] & ~uOnDc[2];
        uOn0[1] = uOn[1] & ~uOnDc[3];
        uOn1[0] = uOn[2] & ~uOnDc[0];
        uOn1[1] = uOn[3] & ~uOnDc[1];
        uOnDc2[0] = uOnDc[0] & uOnDc[2];
        uOnDc2[1] = uOnDc[1] & uOnDc[3];
        // solve for cofactors
        nCubes += Abc_Tt7Isop( uOn0, uOnDc+0, 7, uRes0 );
        nCubes += Abc_Tt7Isop( uOn1, uOnDc+2, 7, uRes1 );
        uOn2[0] = (uOn[0] & ~uRes0[0]) | (uOn[2] & ~uRes1[0]);
        uOn2[1] = (uOn[1] & ~uRes0[1]) | (uOn[3] & ~uRes1[1]);
        nCubes += Abc_Tt7Isop( uOn2, uOnDc2, 7, uRes2 );
        // derive the final truth table
        uRes[0] = uRes2[0] | uRes0[0];
        uRes[1] = uRes2[1] | uRes0[1];
        uRes[2] = uRes2[0] | uRes1[0];
        uRes[3] = uRes2[1] | uRes1[1];
        assert((uOn[0] & ~uRes[0]) == 0 && (uOn[1] & ~uRes[1]) == 0 && (uOn[2] & ~uRes[2]) == 0 && (uOn[3] & ~uRes[3]) == 0, "(uOn[0] & ~uRes[0]) == 0 && (uOn[1] & ~uRes[1]) == 0 && (uOn[2] & ~uRes[2]) == 0 && (uOn[3] & ~uRes[3]) == 0");
        assert((uRes[0] & ~uOnDc[0])==0 && (uRes[1] & ~uOnDc[1])==0 && (uRes[2] & ~uOnDc[2])==0 && (uRes[3] & ~uOnDc[3])==0, "(uRes[0] & ~uOnDc[0])==0 && (uRes[1] & ~uOnDc[1])==0 && (uRes[2] & ~uOnDc[2])==0 && (uRes[3] & ~uOnDc[3])==0");
    }
    return nCubes;
}

/**Function*************************************************************

  Synopsis    [Computes CNF size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6CnfSize(t word, nVars int32) int32 {
    var nCubes int32;
    Abc_Tt6Isop(  t,  t, nVars, &nCubes );
    Abc_Tt6Isop( ~t, ~t, nVars, &nCubes );
    assert(nCubes <= 64, "nCubes <= 64");
    return nCubes;
}
func Abc_Tt8CnfSize(t []word, nVars int32) int32 {
    var uRes [4]word;
    var tc [4]word = [4]word{~t[0], ~t[1], ~t[2], ~t[3]};
    var nCubes int32;
    nCubes += Abc_Tt8Isop( t,  t,  nVars, uRes );
    nCubes += Abc_Tt8Isop( tc, tc, nVars, uRes );
    assert(nCubes <= 256, "nCubes <= 256");
    return nCubes;
}

/**Function*************************************************************

  Synopsis    [Derives ISOP cover for the function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6IsopCover(uOn word, uOnDc word, nVars int32, pCover *int32, pnCubes *int32) word {
    var uOn0, uOn1, uOnDc0, uOnDc1, uRes0, uRes1, uRes2 word;
    var c, Var, nBeg0, nEnd0, nEnd1 int32;
    assert(nVars <= 6, "nVars <= 6");
    assert((uOn & ~uOnDc) == 0, "(uOn & ~uOnDc) == 0");
    if uOn == 0 {
        return 0; }
    if uOnDc == ~(word)0 {
        pCover[(*pnCubes)++] = 0;
        return ~(word)0;
    }
    assert(nVars > 0, "nVars > 0");
    // find the topmost var
    for  Var = nVars-1; Var >= 0; Var--  {
        if (Abc_Tt6HasVar( uOn, Var ) != 0) || (Abc_Tt6HasVar( uOnDc, Var ) != 0) {
             break; }
    }
    assert(Var >= 0, "Var >= 0");
    // cofactor
    uOn0   = Abc_Tt6Cofactor0( uOn,   Var );
    uOn1   = Abc_Tt6Cofactor1( uOn  , Var );
    uOnDc0 = Abc_Tt6Cofactor0( uOnDc, Var );
    uOnDc1 = Abc_Tt6Cofactor1( uOnDc, Var );
    // solve for cofactors
    nBeg0 = *pnCubes; 
    uRes0 = Abc_Tt6IsopCover( uOn0 & ~uOnDc1, uOnDc0, Var, pCover, pnCubes );
    nEnd0 = *pnCubes;
    uRes1 = Abc_Tt6IsopCover( uOn1 & ~uOnDc0, uOnDc1, Var, pCover, pnCubes );
    nEnd1 = *pnCubes;
    uRes2 = Abc_Tt6IsopCover( (uOn0 & ~uRes0) | (uOn1 & ~uRes1), uOnDc0 & uOnDc1, Var, pCover, pnCubes );
    // derive the final truth table
    uRes2 |= (uRes0 & s_Truths6Neg[Var]) | (uRes1 & s_Truths6[Var]);
    for  c = nBeg0; c < nEnd0; c++  {
        pCover[c] |= (1 << (2*Var+0));
    }
    for  c = nEnd0; c < nEnd1; c++  {
        pCover[c] |= (1 << (2*Var+1));
    }
    assert((uOn & ~uRes2) == 0, "(uOn & ~uRes2) == 0");
    assert((uRes2 & ~uOnDc) == 0, "(uRes2 & ~uOnDc) == 0");
    return uRes2;
}
func Abc_Tt7IsopCover(uOn []word, uOnDc []word, nVars int32, uRes []word, pCover *int32, pnCubes *int32) {
    if nVars <= 6 || (uOn[0] == uOn[1] && uOnDc[0] == uOnDc[1]) {
        uRes[0] = uRes[1] = Abc_Tt6IsopCover( uOn[0], uOnDc[0], Abc_MinInt(nVars, 6), pCover, pnCubes );
    } else {
        var uRes0, uRes1, uRes2 word;
        var c, nBeg0, nEnd0, nEnd1 int32;
        assert(nVars == 7, "nVars == 7");
        // solve for cofactors
        nBeg0 = *pnCubes; 
        uRes0 = Abc_Tt6IsopCover( uOn[0] & ~uOnDc[1], uOnDc[0], 6, pCover, pnCubes );   
        nEnd0 = *pnCubes;
        uRes1 = Abc_Tt6IsopCover( uOn[1] & ~uOnDc[0], uOnDc[1], 6, pCover, pnCubes );   
        nEnd1 = *pnCubes;
        uRes2 = Abc_Tt6IsopCover( (uOn[0] & ~uRes0) | (uOn[1] & ~uRes1), uOnDc[0] & uOnDc[1], 6, pCover, pnCubes );
        // derive the final truth table
        uRes[0] = uRes2 | uRes0;
        uRes[1] = uRes2 | uRes1;
        for  c = nBeg0; c < nEnd0; c++  {
            pCover[c] |= (1 << (2*6+0));
        }
        for  c = nEnd0; c < nEnd1; c++  {
            pCover[c] |= (1 << (2*6+1));
        }
        assert((uOn[0] & ~uRes[0]) == 0 && (uOn[1] & ~uRes[1]) == 0, "(uOn[0] & ~uRes[0]) == 0 && (uOn[1] & ~uRes[1]) == 0");
        assert((uRes[0] & ~uOnDc[0])==0 && (uRes[1] & ~uOnDc[1])==0, "(uRes[0] & ~uOnDc[0])==0 && (uRes[1] & ~uOnDc[1])==0");
    }
}
func Abc_Tt8IsopCover(uOn []word, uOnDc []word, nVars int32, uRes []word, pCover *int32, pnCubes *int32) {
    if nVars <= 6 {
        uRes[0] = uRes[1] = uRes[2] = uRes[3] = Abc_Tt6IsopCover( uOn[0], uOnDc[0], nVars, pCover, pnCubes );
    } else if nVars == 7 || (uOn[0] == uOn[2] && uOn[1] == uOn[3] && uOnDc[0] == uOnDc[2] && uOnDc[1] == uOnDc[3]) {
        Abc_Tt7IsopCover( uOn, uOnDc, 7, uRes, pCover, pnCubes );
        uRes[2] = uRes[0];
        uRes[3] = uRes[1];
    
    } else {
        var uOn0, uOn1, uOn2, uOnDc2, uRes0, uRes1, uRes2 [2]word;
        var c, nBeg0, nEnd0, nEnd1 int32;
        assert(nVars == 8, "nVars == 8");
        // cofactor
        uOn0[0] = uOn[0] & ~uOnDc[2];
        uOn0[1] = uOn[1] & ~uOnDc[3];
        uOn1[0] = uOn[2] & ~uOnDc[0];
        uOn1[1] = uOn[3] & ~uOnDc[1];
        uOnDc2[0] = uOnDc[0] & uOnDc[2];
        uOnDc2[1] = uOnDc[1] & uOnDc[3];
        // solve for cofactors
        nBeg0 = *pnCubes; 
        Abc_Tt7IsopCover( uOn0, uOnDc+0, 7, uRes0, pCover, pnCubes );
        nEnd0 = *pnCubes;
        Abc_Tt7IsopCover( uOn1, uOnDc+2, 7, uRes1, pCover, pnCubes );
        nEnd1 = *pnCubes;
        uOn2[0] = (uOn[0] & ~uRes0[0]) | (uOn[2] & ~uRes1[0]);
        uOn2[1] = (uOn[1] & ~uRes0[1]) | (uOn[3] & ~uRes1[1]);
        Abc_Tt7IsopCover( uOn2, uOnDc2, 7, uRes2, pCover, pnCubes );
        // derive the final truth table
        uRes[0] = uRes2[0] | uRes0[0];
        uRes[1] = uRes2[1] | uRes0[1];
        uRes[2] = uRes2[0] | uRes1[0];
        uRes[3] = uRes2[1] | uRes1[1];
        for  c = nBeg0; c < nEnd0; c++  {
            pCover[c] |= (1 << (2*7+0));
        }
        for  c = nEnd0; c < nEnd1; c++  {
            pCover[c] |= (1 << (2*7+1));
        }
        assert((uOn[0] & ~uRes[0]) == 0 && (uOn[1] & ~uRes[1]) == 0 && (uOn[2] & ~uRes[2]) == 0 && (uOn[3] & ~uRes[3]) == 0, "(uOn[0] & ~uRes[0]) == 0 && (uOn[1] & ~uRes[1]) == 0 && (uOn[2] & ~uRes[2]) == 0 && (uOn[3] & ~uRes[3]) == 0");
        assert((uRes[0] & ~uOnDc[0])==0 && (uRes[1] & ~uOnDc[1])==0 && (uRes[2] & ~uOnDc[2])==0 && (uRes[3] & ~uOnDc[3])==0, "(uRes[0] & ~uOnDc[0])==0 && (uRes[1] & ~uOnDc[1])==0 && (uRes[2] & ~uOnDc[2])==0 && (uRes[3] & ~uOnDc[3])==0");
    }
}

/**Function*************************************************************

  Synopsis    [Computes CNF for the function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6Cnf(t word, nVars int32, pCover *int32) int32 {
    var c, nCubes int32;
    Abc_Tt6IsopCover( t, t, nVars, pCover, &nCubes );
    for  c = 0; c < nCubes; c++  {
        pCover[c] |= (1 << (2*nVars+0));
    }
    Abc_Tt6IsopCover( ~t, ~t, nVars, pCover, &nCubes );
    for  ; c < nCubes; c++  {
        pCover[c] |= (1 << (2*nVars+1));
    }
    assert(nCubes <= 64, "nCubes <= 64");
    return nCubes;
}
func Abc_Tt8Cnf(t []word, nVars int32, pCover *int32) int32 {
    var uRes [4]word;
    var tc [4]word = [4]word{~t[0], ~t[1], ~t[2], ~t[3]};
    var c, nCubes int32;
    Abc_Tt8IsopCover( t,  t,  nVars, uRes, pCover, &nCubes );
    for  c = 0; c < nCubes; c++  {
        pCover[c] |= (1 << (2*nVars+0));
    }
    Abc_Tt8IsopCover( tc, tc, nVars, uRes, pCover, &nCubes );
    for  ; c < nCubes; c++  {
        pCover[c] |= (1 << (2*nVars+1));
    }
    assert(nCubes <= 256, "nCubes <= 256");
    return nCubes;
}


/**Function*************************************************************

  Synopsis    [Computes ISOP for 6 variables or less.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6Esop(t word, nVars int32, pCover *int32) int32 {
    var c0, c1 word;
    var Var, r0, r1, r2, Max, i int32;
    assert(nVars <= 6, "nVars <= 6");
    if t == 0 {
        return 0; }
    if t == ~(word)0 {
        if pCover != nil { *pCover = 0; }
        return 1;
    }
    assert(nVars > 0, "nVars > 0");
    // find the topmost var
    for  Var = nVars-1; Var >= 0; Var--  {
        if Abc_Tt6HasVar( t, Var ) != 0 {
             break; }
    }
    assert(Var >= 0, "Var >= 0");
    // cofactor
    c0 = Abc_Tt6Cofactor0( t, Var );
    c1 = Abc_Tt6Cofactor1( t, Var );
    // call recursively
    r0 = Abc_Tt6Esop( c0,      Var, pCover ? pCover : nil );
    r1 = Abc_Tt6Esop( c1,      Var, pCover ? pCover + r0 : nil );
    r2 = Abc_Tt6Esop( c0 ^ c1, Var, pCover ? pCover + r0 + r1 : nil );
    Max = Abc_MaxInt( r0, Abc_MaxInt(r1, r2) );
    // add literals
    if pCover != nil {
        if Max == r0 {
            for  i = 0; i < r1; i++  {
                pCover[i] = pCover[r0+i];
            }
            for  i = 0; i < r2; i++  {
                pCover[r1+i] = pCover[r0+r1+i] | (1 << (2*Var+0));
            }
        
        } else if Max == r1 {
            for  i = 0; i < r2; i++  {
                pCover[r0+i] = pCover[r0+r1+i] | (1 << (2*Var+1));
            }
        
        } else {
            for  i = 0; i < r0; i++  {
                pCover[i] |= (1 << (2*Var+0));
            }
            for  i = 0; i < r1; i++  {
                pCover[r0+i] |= (1 << (2*Var+1));
            }
        }
    }
    return r0 + r1 + r2 - Max;
}
func Abc_Tt6EsopBuild(nVars int32, pCover *int32, nCubes int32) word {
    var p, t word; var c, v int32;
    for  c = 0; c < nCubes; c++  {
        p = ~(word)0;
        for  v = 0; v < nVars; v++  {
            if ((pCover[c] >> (v << 1)) & 3) == 1 {
                p &= ~s_Truths6[v];
            } else if ((pCover[c] >> (v << 1)) & 3) == 2 {
                p &= s_Truths6[v]; }
        }
        t ^= p;
    }
    return t;
}
func Abc_Tt6EsopVerify(t word, nVars int32) int32 {
    var pCover [64]int32;
    var nCubes int32  = Abc_Tt6Esop( t, nVars, pCover );
    var t2 word  = Abc_Tt6EsopBuild( nVars, pCover, nCubes );
    if t != t2 {
        fmt.Printf("Verification failed.\n"); }
    return nCubes;
}

/**Function*************************************************************

  Synopsis    [Check if the function is output-decomposable with the given var.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6CheckOutDec(t word, i int32, pOut *word) int32 {
    var c0 word  = Abc_Tt6Cofactor0( t, i );
    var c1 word  = Abc_Tt6Cofactor1( t, i );
    assert(c0 != c1, "c0 != c1");
    if c0 == 0 { //  F = i * G
    
        if pOut != nil { *pOut = c1; }
        return 0;
    }
    if c1 == 0 { //  F = ~i * G
    
        if pOut != nil { *pOut = c0; }
        return 1;
    }
    if ~c0 == 0 { //  F = ~i + G
    
        if pOut != nil { *pOut = c1; }
        return 2;
    }
    if ~c1 == 0 { //  F = i + G
    
        if pOut != nil { *pOut = c0; }
        return 3;
    }
    if c0 == ~c1 { //  F = i # G
    
        if pOut != nil { *pOut = c0; }
        return 4;
    }
    return -1;
}
func Abc_TtCheckOutDec(pTruth *word, nVars int32, v int32, pOut *word) int32 {
    var Cof0, Cof1 [4]word;
    var nWords int32  = Abc_TtWordNum(nVars);
    assert(nVars <= 8, "nVars <= 8");
    Abc_TtCofactor0p( Cof0, pTruth, nWords, v );
    Abc_TtCofactor1p( Cof1, pTruth, nWords, v );
    assert(!(Abc_TtEqual(Cof0, Cof1, nWords) != 0), "!Abc_TtEqual(Cof0, Cof1, nWords)");
    if Abc_TtIsConst0(Cof0, nWords) != 0 { //if ( c0 == 0 ) //  F = i * G
    
        if pOut != nil { Abc_TtCopy( pOut, Cof1, nWords, 0 ); } //*pOut = c1;
        return 0;
    }
    if Abc_TtIsConst0(Cof1, nWords) != 0 { //if ( c1 == 0 ) //  F = ~i * G
    
        if pOut != nil { Abc_TtCopy( pOut, Cof0, nWords, 0 ); } //*pOut = c0;
        return 1;
    }
    if Abc_TtIsConst1(Cof0, nWords) != 0 { //if ( ~c0 == 0 ) //  F = ~i + G
    
        if pOut != nil { Abc_TtCopy( pOut, Cof1, nWords, 0 ); } //*pOut = c1;
        return 2;
    }
    if Abc_TtIsConst1(Cof1, nWords) != 0 { //if ( ~c1 == 0 ) //  F = i + G
    
        if pOut != nil { Abc_TtCopy( pOut, Cof0, nWords, 0 ); } //*pOut = c0;
        return 3;
    }
    if Abc_TtOpposite(Cof0, Cof1, nWords) != 0 { //if ( c0 == ~c1 )  //  F = i # G
    
        if pOut != nil { Abc_TtCopy( pOut, Cof0, nWords, 0 ); } //*pOut = c0;
        return 4;
    }
    return -1;
}
func Abc_TtCheckDecOutOne7(t *word, piVar *int32, pType *int32) word {
    var v, Type, Type2 int32; var Out [2]word;
    for  v = 6; v >= 0; v--  {
        if (Type = Abc_TtCheckOutDec(t, 7, v, nil)) != -1 {
            Abc_TtSwapVars( t, 7, 6, v );
            Type2 = Abc_TtCheckOutDec( t, 7, 6, Out );
            assert(Type == Type2, "Type == Type2");
            *piVar = v;
            *pType = Type;
            return Out[0];
        }
    }
    return 0;
}
func Abc_TtCheckDecOutOne8(t *word, piVar1 *int32, piVar2 *int32, pType1 *int32, pType2 *int32) word {
    var v, Type1, Type12, Type2, Type22 int32; var Out [4]word;
var Out2 [2]word;
    for  v = 7; v >= 0; v--  {
        if (Type1 = Abc_TtCheckOutDec(t, 8, v, nil)) != -1 {
            Abc_TtSwapVars( t, 8, 7, v );
            Type12 = Abc_TtCheckOutDec( t, 8, 7, Out );
            assert(Type1 == Type12, "Type1 == Type12");
            *piVar1 = v;
            *pType1 = Type1;
            break;
        }
    }
    if v == -1 {
        return 0; }
    for  v = 6; v >= 0; v--  {
        if (Type2 = Abc_TtCheckOutDec(Out, 7, v, nil)) != -1 && Abc_Lit2Var(Type2) == Abc_Lit2Var(Type1) {
            Abc_TtSwapVars( Out, 7, 6, v );
            Type22 = Abc_TtCheckOutDec(Out, 7, 6, Out2);
            assert(Type2 == Type22, "Type2 == Type22");
            *piVar2 = v;
            *pType2 = Type2;
            assert(*piVar2 < *piVar1, "*piVar2 < *piVar1");
            return Out2[0];
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Check if the function is input-decomposable with the given pair.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtCheckDsdAnd(t word, i int32, j int32, pOut *word) int32 {
    var c0 word  = Abc_Tt6Cofactor0( t, i );
    var c1 word  = Abc_Tt6Cofactor1( t, i );
    var c00 word  = Abc_Tt6Cofactor0( c0, j );
    var c01 word  = Abc_Tt6Cofactor1( c0, j );
    var c10 word  = Abc_Tt6Cofactor0( c1, j );
    var c11 word  = Abc_Tt6Cofactor1( c1, j );
    if c00 == c01 && c00 == c10 { //  i *  j
    
        if pOut != nil { *pOut = (~s_Truths6[i] & c00) | (s_Truths6[i] & c11); }
        return 0;
    }
    if c11 == c00 && c11 == c10 { //  i * !j
    
        if pOut != nil { *pOut = (~s_Truths6[i] & c11) | (s_Truths6[i] & c01); }
        return 1;
    }
    if c11 == c00 && c11 == c01 { // !i *  j
    
        if pOut != nil { *pOut = (~s_Truths6[i] & c11) | (s_Truths6[i] & c10); }
        return 2;
    }
    if c11 == c01 && c11 == c10 { // !i * !j
    
        if pOut != nil { *pOut = (~s_Truths6[i] & c11) | (s_Truths6[i] & c00); }
        return 3;
    }
    if c00 == c11 && c01 == c10 {
        if pOut != nil { *pOut = (~s_Truths6[i] & c11) | (s_Truths6[i] & c10); }
        return 4;
    }
    return -1;
}
func Abc_TtCheckDsdMux(t word, i int32, pOut *word) int32 {
    var c0 word  = Abc_Tt6Cofactor0( t, i );
    var c1 word  = Abc_Tt6Cofactor1( t, i );
    var c00, c01, c10, c11 word;
    var k int32;
    var fPres0 int32;
    var fPres1 int32;
    var iVar0 int32 = -1;
    var iVar1 int32 = -1;
    for  k = 0; k < 6; k++  {
        if k == i { continue; }
        fPres0 = Abc_Tt6HasVar( c0, k );
        fPres1 = Abc_Tt6HasVar( c1, k );
        if (fPres0 != 0) && !(fPres1 != 0) {
            if iVar0 >= 0 {
                return -1; }
            iVar0 = k;
        }
        if !(fPres0 != 0) && (fPres1 != 0) {
            if iVar1 >= 0 {
                return -1; }
            iVar1 = k;
        }
    }
    if iVar0 == -1 || iVar1 == -1 {
        return -1; }
    c00 = Abc_Tt6Cofactor0( c0, iVar0 );
    c01 = Abc_Tt6Cofactor1( c0, iVar0 );
    c10 = Abc_Tt6Cofactor0( c1, iVar1 );
    c11 = Abc_Tt6Cofactor1( c1, iVar1 );
    if c00 ==  c10 && c01 ==  c11 { //  ITE(i,  iVar1,  iVar0)
    
        if pOut != nil { *pOut = (~s_Truths6[i] & c10) | (s_Truths6[i] & c11); }
        return (Abc_Var2Lit(iVar1, 0) << 16) | Abc_Var2Lit(iVar0, 0);
    }
    if c00 == ~c10 && c01 == ~c11 { //  ITE(i,  iVar1, !iVar0)
    
        if pOut != nil { *pOut = (~s_Truths6[i] & c10) | (s_Truths6[i] & c11); }
        return (Abc_Var2Lit(iVar1, 0) << 16) | Abc_Var2Lit(iVar0, 1);
    }
    return -1;
}
func Unm_ManCheckTest2() {
    var t, t1, Out, Var0, Var1, Var0_, Var1_ word;
    var iVar0, iVar1, i, Res int32;
    for  iVar0 = 0; iVar0 < 6; iVar0++  {
    for  iVar1 = 0; iVar1 < 6; iVar1++  {
        if iVar0 == iVar1 {
            continue; }
        Var0 = s_Truths6[iVar0];
        Var1 = s_Truths6[iVar1];
        for  i = 0; i < 5; i++  {
            if (i >> 0) & 1 {
                Var0_ = ~Var0;
            } else {
                Var0_ = Var0;
            }
            if (i >> 1) & 1 {
                Var1_ = ~Var1;
            } else {
                Var1_ = Var1;
            }

            t = Var0_ & Var1_;
            if i == 4 {
                t = ~(Var0_ ^ Var1_); }

//            Kit_DsdPrintFromTruth( (unsigned *)&t, 6 ), printf( "\n" );

            Res = Abc_TtCheckDsdAnd( t, iVar0, iVar1, &Out );
            if Res == -1 {
                fmt.Printf("No decomposition\n");
                continue;
            }

            Var0_ = s_Truths6[iVar0];
            if (Res >> 0) & 1 {
                Var0_ = ~Var0_;
            } else {
                Var0_ = Var0_;
            }

            Var1_ = s_Truths6[iVar1];
            if (Res >> 1) & 1 {
                Var1_ = ~Var1_;
            } else {
                Var1_ = Var1_;
            }

            t1 = Var0_ & Var1_;
            if Res == 4 {
                t1 = Var0_ ^ Var1_; }

            t1 = (~t1 & Abc_Tt6Cofactor0(Out, iVar0)) | (t1 & Abc_Tt6Cofactor1(Out, iVar0));

//            Kit_DsdPrintFromTruth( (unsigned *)&t1, 6 ), printf( "\n" );

            if t1 != t {
                fmt.Printf("Verification failed.\n");
            } else {
                fmt.Printf("Verification succeeded.\n"); }
        }
    }
    }
}
func Unm_ManCheckTest() {
    var t, t1, Out, Ctrl, Var0, Var1, Ctrl_, Var0_, Var1_ word;
    var iVar0, iVar1, iCtrl, i, Res int32;
    for  iCtrl = 0; iCtrl < 6; iCtrl++  {
    for  iVar0 = 0; iVar0 < 6; iVar0++  {
    for  iVar1 = 0; iVar1 < 6; iVar1++  {
        if iCtrl == iVar0 || iCtrl == iVar1 || iVar0 == iVar1 {
            continue; }
        Ctrl = s_Truths6[iCtrl];
        Var0 = s_Truths6[iVar0];
        Var1 = s_Truths6[iVar1];
        for  i = 0; i < 8; i++  {
            if (i >> 0) & 1 {
                Ctrl_ = ~Ctrl;
            } else {
                Ctrl_ = Ctrl;
            }
            if (i >> 1) & 1 {
                Var0_ = ~Var0;
            } else {
                Var0_ = Var0;
            }
            if (i >> 2) & 1 {
                Var1_ = ~Var1;
            } else {
                Var1_ = Var1;
            }

            t = (~Ctrl_ & Var0_) | (Ctrl_ & Var1_);

//            Kit_DsdPrintFromTruth( (unsigned *)&t, 6 ), printf( "\n" );

            Res = Abc_TtCheckDsdMux( t, iCtrl, &Out );
            if Res == -1 {
                fmt.Printf("No decomposition\n");
                continue;
            }

//            Kit_DsdPrintFromTruth( (unsigned *)&Out, 6 ), printf( "\n" );

            Ctrl_ = s_Truths6[iCtrl];
            Var0_ = s_Truths6[Abc_Lit2Var(Res & 0xFFFF)];
            if Abc_LitIsCompl(Res & 0xFFFF) {
                Var0_ = ~Var0_;
            } else {
                Var0_ = Var0_;
            }

            Res >>= 16;
            Var1_ = s_Truths6[Abc_Lit2Var(Res & 0xFFFF)];
            if Abc_LitIsCompl(Res & 0xFFFF) {
                Var1_ = ~Var1_;
            } else {
                Var1_ = Var1_;
            }

            t1 = (~Ctrl_ & Var0_) | (Ctrl_ & Var1_);

//            Kit_DsdPrintFromTruth( (unsigned *)&t1, 6 ), printf( "\n" );
//            Kit_DsdPrintFromTruth( (unsigned *)&Out, 6 ), printf( "\n" );

            t1 = (~t1 & Abc_Tt6Cofactor0(Out, iCtrl)) | (t1 & Abc_Tt6Cofactor1(Out, iCtrl));

//            Kit_DsdPrintFromTruth( (unsigned *)&t1, 6 ), printf( "\n" );

            if t1 != t {
                fmt.Printf("Verification failed.\n");
            } else {
                fmt.Printf("Verification succeeded.\n"); }
        }
    }
    }
    }
}


/**Function*************************************************************

  Synopsis    [Truth table evaluation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtEvalLut6(Ins []word, Lut word, nVars int32) word {
    var Cube, Res word; var k, i int32;
    for  k = 0; k < (1<<nVars); k++  {
        if ((Lut >> k) & 1) == 0 {
            continue; }
        Cube = ~(word)0;
        for  i = 0; i < nVars; i++  {
            Cube &= ((k >> i) & 1) ? Ins[i] : ~Ins[i];
        }
        Res |= Cube;
    }
    return Res;
}
func Abc_TtEvalLut5(Ins []uint32, Lut int32, nVars int32) uint32 {
    var Cube, Res uint32; var k, i int32;
    for  k = 0; k < (1<<nVars); k++  {
        if ((Lut >> k) & 1) == 0 {
            continue; }
        Cube = ~uint32(0);
        for  i = 0; i < nVars; i++  {
            Cube &= ((k >> i) & 1) ? Ins[i] : ~Ins[i];
        }
        Res |= Cube;
    }
    return Res;
}
func Abc_TtEvalLut4(Ins []int32, Lut int32, nVars int32) int32 {
    var Cube, Res int32; var k, i int32;
    for  k = 0; k < (1<<nVars); k++  {
        if ((Lut >> k) & 1) == 0 {
            continue; }
        Cube = ~int32(0);
        for  i = 0; i < nVars; i++  {
            Cube &= ((k >> i) & 1) ? Ins[i] : ~Ins[i];
        }
        Res |= Cube;
    }
    return Res & ~(~0 << (1<<nVars));
}


/**Function*************************************************************

  Synopsis    [Checks existence of bi-decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtComputeGraph(pTruth *word, v int32, nVars int32, pGraph *int32) {
    var Cof0, Cof1 [64]word; // pow( 2, nVarsMax-6 )
    var Cof00, Cof01, Cof10, Cof11 [64]word;
    var CofXor, CofAndTest word; 
    var i int32;
    var w int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    pGraph[v] |= (1 << v);
    if v == nVars - 1 {
        return; }
    assert(v < nVars - 1, "v < nVars - 1");
    Abc_TtCofactor0p( Cof0, pTruth, nWords, v );
    Abc_TtCofactor1p( Cof1, pTruth, nWords, v );
    for  i = v + 1; i < nVars; i++  {
        Abc_TtCofactor0p( Cof00, Cof0, nWords, i );
        Abc_TtCofactor1p( Cof01, Cof0, nWords, i );
        Abc_TtCofactor0p( Cof10, Cof1, nWords, i );
        Abc_TtCofactor1p( Cof11, Cof1, nWords, i );
        for  w = 0; w < nWords; w++  {
            CofXor     =  Cof00[w] ^ Cof01[w]  ^  Cof10[w] ^ Cof11[w];
            CofAndTest = (Cof00[w] & Cof01[w]) | (Cof10[w] & Cof11[w]);
            if CofXor & CofAndTest {
                pGraph[v] |= (1 << i);
                pGraph[i] |= (1 << v);
            
            } else if CofXor & ~CofAndTest {
                pGraph[v] |= (1 << (16+i));
                pGraph[i] |= (1 << (16+v));
            }
        }
    }
}
func Abc_TtPrintVarSet(Mask int32, nVars int32) {
    var i int32;
    for  i = 0; i < nVars; i++  {
        if (Mask >> i) & 1 {
            fmt.Printf("1");
        } else {
            fmt.Printf("."); }
    }
}
func Abc_TtPrintBiDec(pTruth *word, nVars int32) {
    var v int32;
    var pGraph [12]int32 ;
    assert(nVars <= 12, "nVars <= 12");
    for  v = 0; v < nVars; v++  {
        Abc_TtComputeGraph( pTruth, v, nVars, pGraph );
        Abc_TtPrintVarSet( pGraph[v], nVars );
        fmt.Printf("    ");
        Abc_TtPrintVarSet( pGraph[v] >> 16, nVars );
        fmt.Printf("\n");
    }
}
func Abc_TtVerifyBiDec(pTruth *word, nVars int32, This int32, That int32, nSuppLim int32, wThis word, wThat word) int32 {
    var pVarsThis, pVarsThat, pVarsAll [12]int32;
    var nThis int32  = Abc_TtBitCount16(This);
    var nThat int32  = Abc_TtBitCount16(That);
    var i int32;
    var k int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    var pThis [64]word = [64]word{wThis};
    var pThat [64]word = [64]word{wThat};
    assert(nVars <= 12, "nVars <= 12");
    for  i = 0; i < nVars; i++  {
        pVarsAll[i] = i;
    }
    for  i = k = 0; i < nVars; i++  {
        if (This >> i) & 1 {
            pVarsThis[k++] = i; }
    }
    assert(k == nThis, "k == nThis");
    for  i = k = 0; i < nVars; i++  {
        if (That >> i) & 1 {
            pVarsThat[k++] = i; }
    }
    assert(k == nThat, "k == nThat");
    Abc_TtStretch6( pThis, nThis, nVars );
    Abc_TtStretch6( pThat, nThat, nVars );
    Abc_TtExpand( pThis, nVars, pVarsThis, nThis, pVarsAll, nVars );
    Abc_TtExpand( pThat, nVars, pVarsThat, nThat, pVarsAll, nVars );
    for  k = 0; k < nWords; k++  {
        if pTruth[k] != (pThis[k] & pThat[k]) {
            return 0; }
    }
    return 1;
}
func Abc_TtExist(pTruth *word, iVar int32, nWords int32) {
    var Cof0, Cof1 [64]word; 
    Abc_TtCofactor0p( Cof0, pTruth, nWords, iVar );
    Abc_TtCofactor1p( Cof1, pTruth, nWords, iVar );
    Abc_TtOr( pTruth, Cof0, Cof1, nWords );
}
func Abc_TtCheckBiDec(pTruth *word, nVars int32, This int32, That int32) int32 {
    var VarMask [2]int32 = [2]int32{This & ~That, That & ~This};
    var v int32;
    var c int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    var pTempR [2][64]word; 
    for  c = 0; c < 2; c++  {
        Abc_TtCopy( pTempR[c], pTruth, nWords, 0 );
        for  v = 0; v < nVars; v++  {
            if (VarMask[c] >> v) & 1 {
                Abc_TtExist( pTempR[c], v, nWords ); }
        }
    }
    for  v = 0; v < nWords; v++  {
        if ~pTruth[v] & pTempR[0][v] & pTempR[1][v] {
            return 0; }
    }
    return 1;
}
func Abc_TtDeriveBiDecOne(pTruth *word, nVars int32, This int32) word {
    var pTemp [64]word; 
    var nThis int32  = Abc_TtBitCount16(This);
    var v int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    Abc_TtCopy( pTemp, pTruth, nWords, 0 );
    for  v = 0; v < nVars; v++  {
        if !((This >> v) & 1) {
            Abc_TtExist( pTemp, v, nWords ); }
    }
    Abc_TtShrink( pTemp, nThis, nVars, This );
    return Abc_Tt6Stretch( pTemp[0], nThis );
}
func Abc_TtDeriveBiDec(pTruth *word, nVars int32, This int32, That int32, nSuppLim int32, pThis *word, pThat *word) {
    assert(Abc_TtBitCount16(This) <= nSuppLim, "Abc_TtBitCount16(This) <= nSuppLim");
    assert(Abc_TtBitCount16(That) <= nSuppLim, "Abc_TtBitCount16(That) <= nSuppLim");
    pThis[0] = Abc_TtDeriveBiDecOne( pTruth, nVars, This );
    pThat[0] = Abc_TtDeriveBiDecOne( pTruth, nVars, That );
    if !(Abc_TtVerifyBiDec(pTruth, nVars, This, That, nSuppLim, pThis[0], pThat[0] ) != 0) {
        fmt.Printf("Bi-decomposition verification failed.\n"); }
}
// detect simple case of decomposition with topmost AND gate
func Abc_TtCheckBiDecSimple(pTruth *word, nVars int32, nSuppLim int32) int32 {
    var Cof0, Cof1 [64]word; 
    var v int32;
    var Res int32;
    var nDecVars int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    for  v = 0; v < nVars; v++  {
        Abc_TtCofactor0p( Cof0, pTruth, nWords, v );
        Abc_TtCofactor1p( Cof1, pTruth, nWords, v );
        if !(Abc_TtIsConst0(Cof0, nWords) != 0) && !(Abc_TtIsConst0(Cof1, nWords) != 0) {
            continue; }
        nDecVars++;
        Res |= 1 << v;
        if nDecVars >= nVars - nSuppLim {
            return ((Res ^ int32(Abc_Tt6Mask(nVars))) << 16) | Res; }
    }
    return 0;
}
func Abc_TtProcessBiDecInt(pTruth *word, nVars int32, nSuppLim int32) int32 {
    var i int32;
    var v int32;
    var Res int32;
    var nSupp int32;
    var CountShared int32;
    var pGraph [12]int32 ;
    assert(nSuppLim < nVars && nVars <= 2 * nSuppLim && nVars <= 12, "nSuppLim < nVars && nVars <= 2 * nSuppLim && nVars <= 12");
    assert(2 <= nSuppLim && nSuppLim <= 6, "2 <= nSuppLim && nSuppLim <= 6");
    Res = Abc_TtCheckBiDecSimple( pTruth, nVars, nSuppLim );
    if Res {
        return Res; }
    for  v = 0; v < nVars; v++  {
        Abc_TtComputeGraph( pTruth, v, nVars, pGraph );
        nSupp = Abc_TtBitCount16(pGraph[v] & 0xFFFF);
        if nSupp > nSuppLim {
            // this variable is shared - check if the limit is exceeded
            if CountShared++; CountShared > 2*nSuppLim - nVars {
                return 0; }
        
        } else if nVars - nSupp <= nSuppLim {
            var This int32  = pGraph[v] & 0xFFFF;
            var That int32  = This ^ int32(Abc_Tt6Mask(nVars));
            // find the other component
            var Graph int32  = That;
            for  i = 0; i < nVars; i++  {
                if (That >> i) & 1 {
                    Graph |= pGraph[i] & 0xFFFF; }
            }
            // check if this can be done
            if Abc_TtBitCount16(Graph) > nSuppLim {
                continue; }
            // try decomposition
            if Abc_TtCheckBiDec(pTruth, nVars, This, Graph) != 0 {
                return (Graph << 16) | This; }
        }
    }
    return 0;
}
func Abc_TtProcessBiDec(pTruth *word, nVars int32, nSuppLim int32) int32 {
    var pFunc [64]word;
    var Res int32;
    var nWords int32 = Abc_TtWordNum(nVars);
    Abc_TtCopy( pFunc, pTruth, nWords, 0 );
    Res = Abc_TtProcessBiDecInt( pFunc, nVars, nSuppLim );
    if Res {
        return Res; }
    Abc_TtCopy( pFunc, pTruth, nWords, 1 );
    Res = Abc_TtProcessBiDecInt( pFunc, nVars, nSuppLim );
    if Res {
        return Res | (1 << 30); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Tests decomposition procedures.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Abc_TtProcessBiDecTest(pTruth *word, nVars int32, nSuppLim int32) {
    var This word;
    var That word;
    var pTemp [64]word;
    var Res, resThis, resThat int32;//, nThis, nThat;
    var nWords int32  = Abc_TtWordNum(nVars);
    Abc_TtCopy( pTemp, pTruth, nWords, 0 );
    Res = Abc_TtProcessBiDec( pTemp, nVars, nSuppLim );
    if Res == 0 {
        //Dau_DsdPrintFromTruth( pTemp, nVars );
        //printf( "Non_dec\n\n" );
        return;
    }

    resThis = Res & 0xFFFF;
    resThat = Res >> 16;

    Abc_TtDeriveBiDec( pTemp, nVars, resThis, resThat, nSuppLim, &This, &That );
    return;

    //if ( !(resThis & resThat) )
    //    return;

//    Dau_DsdPrintFromTruth( pTemp, nVars );

    //nThis = Abc_TtBitCount16(resThis);
    //nThat = Abc_TtBitCount16(resThat);

    fmt.Printf("Variable sets:    ");
    Abc_TtPrintVarSet( resThis, nVars );
    fmt.Printf("    ");
    Abc_TtPrintVarSet( resThat, nVars );
    fmt.Printf("\n");
    Abc_TtDeriveBiDec( pTemp, nVars, resThis, resThat, nSuppLim, &This, &That );
//    Dau_DsdPrintFromTruth( &This, nThis );
//    Dau_DsdPrintFromTruth( &That, nThat );
    fmt.Printf("\n");
}
func Abc_TtProcessBiDecExperiment() {
    var nVars int32  = 3;
    var nSuppLim int32  = 2;
    var Res, resThis, resThat int32;
    var This, That word;
//    word t = ABC_CONST(0x8000000000000000);
//    word t = (s_Truths6[0] | s_Truths6[1]) & (s_Truths6[2] | s_Truths6[3] | s_Truths6[4] | s_Truths6[5]);
//    word t = ((s_Truths6[0] & s_Truths6[1]) | (~s_Truths6[1] & s_Truths6[2]));
    var t word  = ((s_Truths6[0] | s_Truths6[1]) & (s_Truths6[1] | s_Truths6[2]));
    Abc_TtPrintBiDec( &t, nVars );
    Res = Abc_TtProcessBiDec( &t, nVars, nSuppLim );
    resThis = Res & 0xFFFF;
    resThat = Res >> 16;
    Abc_TtDeriveBiDec( &t, nVars, resThis, resThat, nSuppLim, &This, &That );
//    Dau_DsdPrintFromTruth( &This, Abc_TtBitCount16(resThis) );
//    Dau_DsdPrintFromTruth( &That, Abc_TtBitCount16(resThat) );
    nVars = nSuppLim;
}

/**Function*************************************************************

  Synopsis    [Truth table checking procedure.]

  Description []
               
  SideEffects []

  SeeAlso     [] 

***********************************************************************/
func Abc_Tt4Equal3(c0 int32, c1 int32, c2 int32, c3 int32) int32 {
    if c0 == c1 && c0 == c2 { return 3; }
    if c0 == c1 && c0 == c3 { return 2; }
    if c0 == c3 && c0 == c2 { return 1; }
    if c3 == c1 && c3 == c2 { return 0; }
    return -1;
}
func Abc_Tt4Check2(t int32, i int32, j int32, f *int32, r *int32) int32 {
    var c0 int32   =  t & r[j];
    var c1 int32   = (t & f[j]) >> (1 << j);
    var c00 int32  =  c0 & r[i];
    var c01 int32  = (c0 & f[i]) >> (1 << i);
    var c10 int32  =  c1 & r[i];
    var c11 int32  = (c1 & f[i]) >> (1 << i);
    return Abc_Tt4Equal3( c00, c01, c10, c11 );
}
func Abc_Tt4CheckTwoLevel(t int32) int32 {
    var pair1, pair2 int32;
    var f [4]int32 = [4]int32{ 0xAAAA, 0xCCCC, 0xF0F0, 0xFF00 };
    var r [4]int32 = [4]int32{ 0x5555, 0x3333, 0x0F0F, 0x00FF };
    if (pair1 = Abc_Tt4Check2(t, 0, 1, f, r)) >= 0 && (pair2 = Abc_Tt4Check2(t, 2, 3, f, r)) >= 0 { return (1 << 4) | (pair2 << 2) | pair1; }
    if (pair1 = Abc_Tt4Check2(t, 0, 2, f, r)) >= 0 && (pair2 = Abc_Tt4Check2(t, 1, 3, f, r)) >= 0 { return (2 << 4) | (pair2 << 2) | pair1; }
    if (pair1 = Abc_Tt4Check2(t, 0, 3, f, r)) >= 0 && (pair2 = Abc_Tt4Check2(t, 1, 2, f, r)) >= 0 { return (3 << 4) | (pair2 << 2) | pair1; }
    return -1;
}
func Abc_Tt4CountOnes(t int32) int32 {
    t = (t & (0x5555)) + ((t >> 1) & (0x5555));
    t = (t & (0x3333)) + ((t >> 2) & (0x3333));
    t = (t & (0x0f0f)) + ((t >> 4) & (0x0f0f));
    t = (t & (0x00ff)) + ((t >> 8) & (0x00ff));
    return t;
}
func Abc_Tt4FirstBit(t int32) int32 {
    var n int32;
    if t == 0 { return -1; }
    if (t & 0x00FF) == 0 { n +=  8; t >>=  8; }
    if (t & 0x000F) == 0 { n +=  4; t >>=  4; }
    if (t & 0x0003) == 0 { n +=  2; t >>=  2; }
    if (t & 0x0001) == 0 { n++; }
    return n;
}
func Abc_Tt4Check(t int32) int32 {
    var Count int32;
    var tn int32 = 0xFFFF & ~t;
    if t == 0x6996 || tn == 0x6996 { return 1; }
    if (t & (t-1)) == 0 {            return 1; }
    if (tn & (tn-1)) == 0 {          return 1; }
    Count = Abc_Tt4CountOnes( t );
    if Count == 7 && Abc_Tt4CheckTwoLevel(t)  > 0 { return 1; }
    if Count == 9 && Abc_Tt4CheckTwoLevel(tn) > 0 { return 1; }
    return 0;
}


/**Function*************************************************************

  Synopsis    [Returns symmetry profile of the function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6VarsAreSymmetric(t word, iVar int32, jVar int32) int32 {
    var s_PMasks *word  = s_PPMasks[iVar][jVar];
    var shift int32  = (1 << jVar) - (1 << iVar);
    assert(iVar < jVar, "iVar < jVar");
    return ((t & s_PMasks[1]) << shift) == (t & s_PMasks[2]);
}
func Abc_Tt6VarsAreAntiSymmetric(t word, iVar int32, jVar int32) int32 {
    var s_PMasks *word  = s_PPMasks[iVar][jVar];
    var shift int32  = (1 << jVar) + (1 << iVar);
    assert(iVar < jVar, "iVar < jVar");
    return ((t & (s_PMasks[1] >> (1 << iVar))) << shift) == (t & (s_PMasks[2] << (1 << iVar)));
}
func Abc_TtVarsAreSymmetric(pTruth *word, nVars int32, i int32, j int32, pCof0 *word, pCof1 *word) int32 {
    var nWords int32  = Abc_TtWordNum( nVars );
    assert(i < nVars && j < nVars, "i < nVars && j < nVars");
    Abc_TtCofactor0p( pCof0, pTruth, nWords, i );
    Abc_TtCofactor1p( pCof1, pTruth, nWords, i );
    Abc_TtCofactor1( pCof0, nWords, j );
    Abc_TtCofactor0( pCof1, nWords, j );
    return Abc_TtEqual( pCof0, pCof1, nWords );
}
func Abc_TtVarsAreAntiSymmetric(pTruth *word, nVars int32, i int32, j int32, pCof0 *word, pCof1 *word) int32 {
    var nWords int32  = Abc_TtWordNum( nVars );
    assert(i < nVars && j < nVars, "i < nVars && j < nVars");
    Abc_TtCofactor0p( pCof0, pTruth, nWords, i );
    Abc_TtCofactor1p( pCof1, pTruth, nWords, i );
    Abc_TtCofactor0( pCof0, nWords, j );
    Abc_TtCofactor1( pCof1, nWords, j );
    return Abc_TtEqual( pCof0, pCof1, nWords );
}
func Abc_TtIsFullySymmetric(pTruth *word, nVars int32) int32 {
    var m, v, Polar, Seen int32; 
    for  m = 0; m < (1<<nVars); m++  {
        var Count int32;
        var Value int32  = Abc_TtGetBit( pTruth, m );
        for  v = 0; v < nVars; v++  {
            Count += ((m >> v) & 1);
        }
        if (Seen >> Count) & 1 { // seen this count
        
            if Value != ((Polar >> Count) & 1) {
                return -1; }
        
        } else { // new count
        
            Seen  |= 1 << Count;
            if Value != 0 {
                Polar |= 1 << Count; }
        }
    }
    return Polar;
}
func Abc_TtGenFullySymmetric(pTruth *word, nVars int32, Polar int32) {
    var m int32;
    var v int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    Abc_TtClear( pTruth, nWords );
    for  m = 0; m < (1<<nVars); m++  {
        var Count int32;
        for  v = 0; v < nVars; v++  {
            Count += ((m >> v) & 1);
        }
        if (Polar >> Count) & 1 {
            Abc_TtSetBit( pTruth, m ); }
    }
}
func Abc_TtTestFullySymmetric() {
    var pTruth [4]word; // 8-var function
    var PolarOut int32;
    var PolarIn int32 = 271;
    Abc_TtGenFullySymmetric( pTruth, 8, PolarIn );
    //Abc_TtXorBit( pTruth, 171 );
    PolarOut = Abc_TtIsFullySymmetric( pTruth, 8 );
    assert(PolarIn == PolarOut, "PolarIn == PolarOut");
}


/**Function*************************************************************

  Synopsis    [Generates truth table of a symmetric function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtSymFunGenerate(pOnes string, nVars int32) *word {
    var m, k, Count int32;
    var pTruth *word  = ABC_CALLOC( word, Abc_TtWordNum(nVars) );
    assert(int32(len(pOnes)) == nVars + 1, "int32(strlen(pOnes)) == nVars + 1");
    for  m = 0; m < (1 << nVars); m++  {
        Count = 0;
        for  k = 0; k < nVars; k++  {
            Count += (m >> k) & 1;
        }
        if pOnes[Count] == '1' {
            Abc_TtXorBit( pTruth, m ); }
    }
    return pTruth;
}


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

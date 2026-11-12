package pla

import "fmt"

/**CFile****************************************************************

  FileName    [pla.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: pla.h,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"
//#include "misc/extra/extra.h"
//#include "base/main/mainInt.h"
//#include "misc/util/utilTruth.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

#define MASK55 ABC_CONST(0x5555555555555555)

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// file types
 type Pla_File_t int32
const (
    PLA_FILE_FD Pla_File_t = 0

    PLA_FILE_F Pla_File_t = PLA_FILE_FD + 1

    PLA_FILE_FR Pla_File_t = PLA_FILE_F + 1

    PLA_FILE_FDR Pla_File_t = PLA_FILE_FR + 1

    PLA_FILE_NONE Pla_File_t = PLA_FILE_FDR + 1

)

 

// literal types
 type Pla_Lit_t int32
const (
    PLA_LIT_DASH Pla_Lit_t = 0

    PLA_LIT_ZERO Pla_Lit_t = PLA_LIT_DASH + 1

    PLA_LIT_ONE Pla_Lit_t = PLA_LIT_ZERO + 1

    PLA_LIT_FULL Pla_Lit_t = PLA_LIT_ONE + 1

)

 


   
type Pla_Man_t struct {
    pName            string;      // model name
    pSpec            string;      // input file
    Type       Pla_File_t;       // file type
    nIns              int32;       // inputs
    nOuts              int32;      // outputs
    nInWords              int32;   // words of input data
    nOutWords              int32;  // words of output data
    vCubes        Vec_Int_t;     // cubes
    vHashes        Vec_Int_t;    // hash values
    vInBits        Vec_Wrd_t;    // input bits
    vOutBits        Vec_Wrd_t;   // output bits
    vCubeLits        Vec_Wec_t;  // cubes as interger arrays
    vOccurs        Vec_Wec_t;    // occurence counters for the literals
    vDivs        Vec_Int_t;      // divisor definitions
}

func Pla_ManName(p *Pla_Man_t) string { return p.pName;                   }
func Pla_ManInNum(p *Pla_Man_t) int32 { return p.nIns;                    }
func Pla_ManOutNum(p *Pla_Man_t) int32 { return p.nOuts;                   }
func Pla_ManCubeNum(p *Pla_Man_t) int32 { return Vec_IntSize( &p.vCubes );  }
func Pla_ManDivNum(p *Pla_Man_t) int32 { return Vec_IntSize( &p.vDivs );   }

func Pla_CubeIn(p *Pla_Man_t, i int32) *word { return Vec_WrdEntryP(&p.vInBits,  i * p.nInWords);  }
func Pla_CubeOut(p *Pla_Man_t, i int32) *word { return Vec_WrdEntryP(&p.vOutBits, i * p.nOutWords); }

func Pla_CubeNum(hCube int32) int32 { return hCube >> 8;    }
func Pla_CubeLit(hCube int32) int32 { return hCube & 0xFF;  }
func Pla_CubeHandle(iCube int32, iLit int32) int32 { assert(!(iCube >> 24) && !(iLit >> 8), "!(iCube >> 24) && !(iLit >> 8)"); return iCube << 8 | iLit; }

// read/write/flip i-th bit of a bit string table
func Pla_TtGetBit(p *word, i int32) int32 { return int32((p[i>>6] >> (i & 63))) & 1;      }
func Pla_TtSetBit(p *word, i int32) { p[i>>6] |= (((word)1)<<(i & 63));           }
func Pla_TtXorBit(p *word, i int32) { p[i>>6] ^= (((word)1)<<(i & 63));           }

// read/write/flip i-th literal in a cube
func Pla_CubeGetLit(p *word, i int32) int32 { return int32((p[i>>5] >> ((i<<1) & 63))) & 3; }
func Pla_CubeSetLit(p *word, i int32, d Pla_Lit_t) { p[i>>5] |= (((word)d)<<((i<<1) & 63));      }
func Pla_CubeXorLit(p *word, i int32, d Pla_Lit_t) { p[i>>5] ^= (((word)d)<<((i<<1) & 63));      }


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Pla_ForEachCubeIn( p, pCube, i )                            \
//     for ( i = 0; (i < Pla_ManCubeNum(p))  && (((pCube) = Pla_CubeIn(p, i)), 1); i++ )

// Original macro:
// #define Pla_ForEachCubeInStart( p, pCube, i, Start )                \
//     for ( i = Start; (i < Pla_ManCubeNum(p))  && (((pCube) = Pla_CubeIn(p, i)), 1); i++ )


// Original macro:
// #define Pla_ForEachCubeOut( p, pCube, i )                           \
//     for ( i = 0; (i < Pla_ManCubeNum(p))  && (((pCube) = Pla_CubeOut(p, i)), 1); i++ )

// Original macro:
// #define Pla_ForEachCubeInOut( p, pCubeIn, pCubeOut, i )             \
//     for ( i = 0; (i < Pla_ManCubeNum(p))  && (((pCubeIn) = Pla_CubeIn(p, i)), 1)  && (((pCubeOut) = Pla_CubeOut(p, i)), 1); i++ )


// Original macro:
// #define Pla_CubeForEachLit( nVars, pCube, Lit, i )                  \
//     for ( i = 0; (i < nVars)  && (((Lit) = Pla_CubeGetLit(pCube, i)), 1); i++ )

// Original macro:
// #define Pla_CubeForEachLitIn( p, pCube, Lit, i )                    \
//     Pla_CubeForEachLit( Pla_ManInNum(p), pCube, Lit, i )

// Original macro:
// #define Pla_CubeForEachLitOut( p, pCube, Lit, i )                   \
//     Pla_CubeForEachLit( Pla_ManOutNum(p), pCube, Lit, i )



////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checks if cubes are distance-1.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pla_OnlyOneOne(t word) int32 {
    if t {
        return ((t & (t-1)) == 0);
    } else {
        return 0;
    }
}
func Pla_CubesAreDistance1(p *word, q *word, nWords int32) int32 {
    var Test word; var c, fFound int32;
    for  c = 0; c < nWords; c++  {
        if p[c] == q[c] {
            continue; }
        if fFound != 0 {
            return 0; }
        // check if the number of 1s is one, which means exactly one different literal (0/1, -/1, -/0)
        Test = ((p[c] ^ q[c]) | ((p[c] ^ q[c]) >> 1)) & MASK55;
        if !(Pla_OnlyOneOne(Test) != 0) {
            return 0; }
        fFound = 1;
    }
    return fFound;
}
func Pla_CubesAreConsensus(p *word, q *word, nWords int32, piVar *int32) int32 {
    var Test word; var c, fFound int32;
    for  c = 0; c < nWords; c++  {
        if p[c] == q[c] {
            continue; }
        if fFound != 0 {
            return 0; }
        // check if the number of 1s is one, which means exactly one opposite literal (0/1) but may have other diffs (-/0 or -/1)
        Test = ((p[c] ^ q[c]) & ((p[c] ^ q[c]) >> 1)) & MASK55;
        if !(Pla_OnlyOneOne(Test) != 0) {
            return 0; }
        fFound = 1;
        if piVar != nil { *piVar = c * 32 + Abc_Tt6FirstBit(Test)/2; }
    }
    return fFound;
}
func Pla_TtCountOnesOne(x word) int32 {
    x = x - ((x >> 1) & ABC_CONST(0x5555555555555555));
    x = (x & ABC_CONST(0x3333333333333333)) + ((x >> 2) & ABC_CONST(0x3333333333333333));
    x = (x + (x >> 4)) & ABC_CONST(0x0F0F0F0F0F0F0F0F);
    x = x + (x >> 8);
    x = x + (x >> 16);
    x = x + (x >> 32);
    return int32((x & 0xFF));
}
func Pla_TtCountOnes(p *word, nWords int32) int32 {
    var i, Count int32;
    for  i = 0; i < nWords; i++  {
        Count += Pla_TtCountOnesOne( p[i] );
    }
    return Count;
}

/**Function*************************************************************

  Synopsis    [Manager APIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pla_ManAlloc(pFileName string, nIns int32, nOuts int32, nCubes int32) *Pla_Man_t {
    var p *Pla_Man_t  = new(Pla_Man_t);
    p.pName = Extra_FileDesignName( pFileName );
    p.pSpec = Abc_UtilStrsav( pFileName );
    p.nIns  = nIns;
    p.nOuts = nOuts;
    p.nInWords  = Abc_Bit6WordNum( 2*nIns );
    p.nOutWords = Abc_Bit6WordNum( 2*nOuts );
    Vec_IntFillNatural( &p.vCubes, nCubes );
    Vec_WrdFill( &p.vInBits,  Pla_ManCubeNum(p) * p.nInWords,  0 );
    Vec_WrdFill( &p.vOutBits, Pla_ManCubeNum(p) * p.nOutWords, 0 );
    return p;
}
func Pla_ManFree(p *Pla_Man_t) {
    Vec_IntErase( &p.vCubes );
    Vec_IntErase( &p.vHashes );
    Vec_WrdErase( &p.vInBits );
    Vec_WrdErase( &p.vOutBits );
    Vec_WecErase( &p.vCubeLits );
    Vec_WecErase( &p.vOccurs );
    Vec_IntErase( &p.vDivs );
    ABC_FREE( p.pName );
    ABC_FREE( p.pSpec );
    ABC_FREE( p );
}
func Pla_ManLitInNum(p *Pla_Man_t) int32 {
    var pCube *word; var i, k, Lit, Count int32;
    // inlining of macro Pla_ForEachCubeIn
    for  i = 0; (i < Pla_ManCubeNum(p))  && (((pCube) = Pla_CubeIn(p, i)), 1); i++  {
        // inlining of macro Pla_CubeForEachLitIn
        // inlining of macro Pla_CubeForEachLit
        for  k = 0; (k < Pla_ManInNum(p))  && (((Lit) = Pla_CubeGetLit(pCube, k)), 1); k++  {
            Count += Lit != PLA_LIT_DASH;
        }
    }
    return Count;
}
func Pla_ManLitOutNum(p *Pla_Man_t) int32 {
    var pCube *word; var i, k, Lit, Count int32;
    // inlining of macro Pla_ForEachCubeOut
    for  i = 0; (i < Pla_ManCubeNum(p))  && (((pCube) = Pla_CubeOut(p, i)), 1); i++  {
        // inlining of macro Pla_CubeForEachLitOut
        // inlining of macro Pla_CubeForEachLit
        for  k = 0; (k < Pla_ManOutNum(p))  && (((Lit) = Pla_CubeGetLit(pCube, k)), 1); k++  {
            Count += Lit == PLA_LIT_ONE;
        }
    }
    return Count;
}
func Pla_ManPrintStats(p *Pla_Man_t, fVerbose int32) {
    fmt.Printf("%-16s :  ", Pla_ManName(p));
    fmt.Printf("In =%4d  ", Pla_ManInNum(p));
    fmt.Printf("Out =%4d  ", Pla_ManOutNum(p));
    fmt.Printf("Cube =%8d  ", Pla_ManCubeNum(p));
    fmt.Printf("LitIn =%8d  ", Pla_ManLitInNum(p));
    fmt.Printf("LitOut =%8d  ", Pla_ManLitOutNum(p));
    fmt.Printf("Div =%6d  ", Pla_ManDivNum(p));
    fmt.Printf("\n");
}



/*=== plaHash.c ========================================================*/
// extern int                 Pla_ManHashDist1NumTest( Pla_Man_t * p );

// extern void                Pla_ManComputeDist1Test( Pla_Man_t * p );

/*=== plaMan.c ========================================================*/
// extern Vec_Bit_t *         Pla_ManPrimesTable( int nVars );

// extern Pla_Man_t *         Pla_ManPrimesDetector( int nVars );

// extern Pla_Man_t *         Pla_ManGenerate( int nIns, int nOuts, int nCubes, int fVerbose );

// extern void                Pla_ManConvertFromBits( Pla_Man_t * p );

// extern void                Pla_ManConvertToBits( Pla_Man_t * p );

// extern int                 Pla_ManDist1NumTest( Pla_Man_t * p );

/*=== plaMerge.c ========================================================*/
// extern int                 Pla_ManDist1Merge( Pla_Man_t * p );

/*=== plaSimple.c ========================================================*/
// extern int                 Pla_ManFxPerformSimple( int nVars );

/*=== plaRead.c ========================================================*/
// extern Pla_Man_t *         Pla_ReadPla( char * pFileName );

/*=== plaWrite.c ========================================================*/
// extern void                Pla_WritePla( Pla_Man_t * p, char * pFileName );


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

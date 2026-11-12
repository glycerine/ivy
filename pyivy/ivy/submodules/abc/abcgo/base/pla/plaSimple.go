package pla

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [plaSimple.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SOP manager.]

  Synopsis    [Scalable SOP transformations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - March 18, 2015.]

  Revision    [$Id: plaSimple.c,v 1.00 2014/09/12 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pla.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Dump PLA manager into a BLIF file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pla_ManDumpPla(p *Pla_Man_t, pFileName string) {
    // find the number of original variables
    //int nVarsInit = Pla_ManDivNum(p) ? Vec_IntCountZero(&p->vDivs) : Pla_ManInNum(p);
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
    } else {
        //char * pLits = "-01?";
        var vStr *Vec_Str_t;
        var vCube *Vec_Int_t;
        var i, k, Lit int32;
        // comment
        fmt.Fprintf(pFile, "# PLA file written via PLA package in ABC on ");
        fmt.Fprintf(pFile, "%s", Extra_TimeStamp());
        fmt.Fprintf(pFile, "\n\n");
        // header
        fmt.Fprintf(pFile, ".i %d\n", Pla_ManInNum(p));
        fmt.Fprintf(pFile, ".o %d\n", 1);
        fmt.Fprintf(pFile, ".p %d\n", Vec_WecSize(&p.vCubeLits));
        // SOP
        vStr = Vec_StrStart( Pla_ManInNum(p) + 1 );
        Vec_WecForEachLevel( &p.vCubeLits, vCube, i )
        {
            if !Vec_IntSize(vCube) {
                continue; }
            for  k = 0; k < Pla_ManInNum(p); k++  {
                Vec_StrWriteEntry( vStr, k, '-' );
            }
            Vec_IntForEachEntry( vCube, Lit, k )
                Vec_StrWriteEntry( vStr, Abc_Lit2Var(Lit), byte((Abc_LitIsCompl(Lit) ? '0' : '1')) );
            fmt.Fprintf(pFile, "%s 1\n", Vec_StrArray(vStr));
        }
        Vec_StrFree( vStr );
        fmt.Fprintf(pFile, ".e\n\n");
        _ = (pFile).Close();
        fmt.Printf("Written file \"%s\".\n", pFileName);
    }
}
func Pla_ManDumpBlif(p *Pla_Man_t, pFileName string) {
    // find the number of original variables
    var nVarsInit int32;
    if Pla_ManDivNum(p) {
        nVarsInit = Vec_IntCountZero(&p.vDivs);
    } else {
        nVarsInit = Pla_ManInNum(p);
    }
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
    } else {
        //char * pLits = "-01?";
        var vStr *Vec_Str_t;
        var vCube *Vec_Int_t;
        var i, k, Lit, Div int32;
        // comment
        fmt.Fprintf(pFile, "# BLIF file written via PLA package in ABC on ");
        fmt.Fprintf(pFile, "%s", Extra_TimeStamp());
        fmt.Fprintf(pFile, "\n\n");
        // header
        fmt.Fprintf(pFile, ".model %s\n", Pla_ManName(p));
        fmt.Fprintf(pFile, ".inputs");
        for  i = 0; i < nVarsInit; i++  {
            fmt.Fprintf(pFile, " i%d", i);
        }
        fmt.Fprintf(pFile, "\n");
        fmt.Fprintf(pFile, ".outputs o");
        fmt.Fprintf(pFile, "\n");
        // SOP
        fmt.Fprintf(pFile, ".names");
        for  i = 0; i < Pla_ManInNum(p); i++  {
            fmt.Fprintf(pFile, " i%d", i);
        }
        fmt.Fprintf(pFile, " o\n");
        vStr = Vec_StrStart( Pla_ManInNum(p) + 1 );
        Vec_WecForEachLevel( &p.vCubeLits, vCube, i )
        {
            for  k = 0; k < Pla_ManInNum(p); k++  {
                Vec_StrWriteEntry( vStr, k, '-' );
            }
            Vec_IntForEachEntry( vCube, Lit, k )
                Vec_StrWriteEntry( vStr, Abc_Lit2Var(Lit), byte((Abc_LitIsCompl(Lit) ? '0' : '1')) );
            fmt.Fprintf(pFile, "%s 1\n", Vec_StrArray(vStr));
        }
        Vec_StrFree( vStr );
        // divisors
        Vec_IntForEachEntryStart( &p.vDivs, Div, i, nVarsInit )
        {
            var pLits [3]int32 = [3]int32{ (Div >> 2) & 0x3FF, (Div >> 12) & 0x3FF, (Div >> 22) & 0x3FF };
            fmt.Fprintf(pFile, ".names");
            fmt.Fprintf(pFile, " i%d", Abc_Lit2Var(pLits[0]));
            fmt.Fprintf(pFile, " i%d", Abc_Lit2Var(pLits[1]));
            if (Div & 3) == 3 { // MUX
                fmt.Fprintf(pFile, " i%d", Abc_Lit2Var(pLits[2])); }
            fmt.Fprintf(pFile, " i%d\n", i);
            if (Div & 3) == 1 { // AND
                fmt.Fprintf(pFile, "%d%d 1\n", !Abc_LitIsCompl(pLits[0]), !Abc_LitIsCompl(pLits[1]));
            } else if (Div & 3) == 2 { // XOR
            
                assert(!Abc_LitIsCompl(pLits[0]), "!Abc_LitIsCompl(pLits[0])");
                assert(!Abc_LitIsCompl(pLits[1]), "!Abc_LitIsCompl(pLits[1])");
                fmt.Fprintf(pFile, "10 1\n01 1\n");
            
            } else if (Div & 3) == 3 { // MUX
            
                assert(!Abc_LitIsCompl(pLits[1]), "!Abc_LitIsCompl(pLits[1])");
                assert(!Abc_LitIsCompl(pLits[2]), "!Abc_LitIsCompl(pLits[2])");
                fmt.Fprintf(pFile, "%d-0 1\n-11 1\n", !Abc_LitIsCompl(pLits[0]));
            
            } else { assert(false, "0"); }
        }
        fmt.Fprintf(pFile, ".end\n\n");
        _ = (pFile).Close();
        fmt.Printf("Written file \"%s\".\n", pFileName);
    }
}

/**Function*************************************************************

  Synopsis    [Transforms truth table into an SOP manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pla_ManExpendDirNum(pOn *word, nBits int32, iMint int32, pVars *int32) int32 {
    var v, nVars int32;
    for  v = 0; v < nBits; v++  {
        if Pla_TtGetBit(pOn, iMint ^ (1 << v)) {
            pVars[nVars++] = v; }
    }
    return nVars;
}
func Pla_PrintBinary(pT *word, nBits int32) {
    var v int32;
    for  v = 0; v < nBits; v++  {
        fmt.Printf("%d", Pla_TtGetBit(pT, v));
    }
    fmt.Printf("\n");
}
func Pla_ManFxMinimize(pOn *word, nVars int32) *Vec_Wrd_t {
    var i int32;
    var v int32;
    var iMint int32;
    var iVar int32;
    var nMints int32 = (1 << nVars);
    var nWords int32  = Abc_Bit6WordNum( nMints );
    var vCubes *Vec_Wrd_t  = Vec_WrdAlloc( 1000 );
    var pDc *word  = ABC_CALLOC( word, nWords );
    var Count [32]int32 ;
    var Cubes [32]int32 ;
    var vStore *Vec_Int_t  = Vec_IntAlloc( 1000 );

    // count the number of expansion directions
    for  i = 0; i < nMints; i++  {
        if Pla_TtGetBit(pOn, i) && !Pla_TtGetBit(pDc, i) {
            var pDirs [32]int32;
            var nDirs int32 = Pla_ManExpendDirNum(pOn, nVars, i, pDirs);
            Count[nDirs]++;
            if nDirs == 0 {
                Pla_TtSetBit(pDc, i);
                Cubes[0]++;
                // save
                Vec_IntPushTwo( vStore, i, -1 );
                continue;
            }
            if nDirs == 1 {
                //Pla_PrintBinary( (word *)&i, nVars );
                //printf( " %d \n", pDirs[0] );

                Pla_TtSetBit(pDc, i);
                Pla_TtSetBit(pDc, i ^ (1 << pDirs[0]));
                Cubes[1]++;
                // save
                Vec_IntPushTwo( vStore, i, pDirs[0] );
                continue;
            }
            if nDirs == 2 && Pla_TtGetBit(pOn, i ^ (1 << pDirs[0]) ^ (1 << pDirs[1])) {
                assert(false, "0");
                Pla_TtSetBit(pDc, i);
                Pla_TtSetBit(pDc, i ^ (1 << pDirs[0]));
                Pla_TtSetBit(pDc, i ^ (1 << pDirs[1]));
                Pla_TtSetBit(pDc, i ^ (1 << pDirs[0]) ^ (1 << pDirs[1]));
                Cubes[2]++;
                continue;
            }
        }
    }

    // go through the remaining cubes
    for  i = 0; i < nMints; i++  {
        if Pla_TtGetBit(pOn, i) && !Pla_TtGetBit(pDc, i) {
            var pDirs [32]int32;
            var nDirs int32 = Pla_ManExpendDirNum(pOn, nVars, i, pDirs);
            // find direction, which is not taken
            for  v = 0; v < nDirs; v++  {
                if Pla_TtGetBit(pOn, i ^ (1 << pDirs[v])) && !Pla_TtGetBit(pDc, i ^ (1 << pDirs[v])) {
                    break; }
            }
            // if there is no open directions, use any one
            if v == nDirs {
                v = 0; }
            // mark this one
            Pla_TtSetBit(pDc, i);
            Pla_TtSetBit(pDc, i ^ (1 << pDirs[v]));
            Cubes[10]++;
            // save
            Vec_IntPushTwo( vStore, i, pDirs[v] );
            continue;
        }
    }

    fmt.Printf("\n");
    fmt.Printf("Truth = %d. ", Pla_TtCountOnes(pOn, nWords));
    fmt.Printf("Cover = %d. ", Pla_TtCountOnes(pDc, nWords));
    fmt.Printf("\n");

    fmt.Printf("Count: ");
    for  i = 0; i < 16; i++  {
        if Count[i] {
            fmt.Printf("%d=%d ", i, Count[i]); }
    }
    fmt.Printf("\n");

    fmt.Printf("Cubes: ");
    for  i = 0; i < 16; i++  {
        if Cubes[i] {
            fmt.Printf("%d=%d ", i, Cubes[i]); }
    }
    fmt.Printf("\n");

/*
    // extract cubes one at a time
    for ( i = 0; i < nMints; i++ )
        if ( Pla_TtGetBit(pOn, i) )
        {
            word Cube = 0;
            for ( v = 0; v < nVars; v++ )
                if ( (i >> v) & 1 )
                    Pla_CubeSetLit( &Cube, v, PLA_LIT_ONE );
                else
                    Pla_CubeSetLit( &Cube, v, PLA_LIT_ZERO );
            Vec_WrdPush( vCubes, Cube );
        }
*/
    Vec_IntForEachEntryDouble( vStore, iMint, iVar, i )
    {
        var Cube word;
        for  v = 0; v < nVars; v++  {
            if v == iVar {
                continue; }
            if (iMint >> v) & 1 {
                Pla_CubeSetLit( &Cube, v, PLA_LIT_ONE );
            } else {
                Pla_CubeSetLit( &Cube, v, PLA_LIT_ZERO ); }
        }
        Vec_WrdPush( vCubes, Cube );
    }
    Vec_IntFree( vStore );

    // collect the minterms
    ABC_FREE( pDc );
    return vCubes;
}
func Pla_ManFxPrepare(nVars int32) *Pla_Man_t {
    var p *Pla_Man_t; var Buffer string // was [1000]byte in C
 
    var vFunc *Vec_Bit_t  = Pla_ManPrimesTable( nVars );
    var vSop *Vec_Wrd_t  = Pla_ManFxMinimize( (word *)Vec_BitArray(vFunc), nVars );
    var Cube word;
    var pCube *word = &Cube; var i, k, Lit int32;
    Buffer = fmt.Sprintf("primes%02d", nVars);
    p = Pla_ManAlloc( Buffer, nVars, 1, Vec_WrdSize(vSop) );
    Vec_WecInit( &p.vCubeLits, Pla_ManCubeNum(p) );
    Vec_WecInit( &p.vOccurs, 2*Pla_ManInNum(p) );
    Vec_WrdForEachEntry( vSop, Cube, i )
        // inlining of macro Pla_CubeForEachLit
        for  k = 0; (k < nVars)  && (((Lit) = Pla_CubeGetLit(pCube, k)), 1); k++  {
            if Lit != PLA_LIT_DASH {
                Lit = Abc_Var2Lit( k, Lit == PLA_LIT_ZERO );
                Vec_WecPush( &p.vCubeLits, i, Lit );
                Vec_WecPush( &p.vOccurs, Lit, i );
            }
        }
    Vec_BitFree( vFunc );
    Vec_WrdFree( vSop );
    return p;
}
func Pla_ManFxPerformSimple(nVars int32) int32 {
    var Buffer string // was [100]byte in C

    var p *Pla_Man_t  = Pla_ManFxPrepare( nVars );
    Buffer = fmt.Sprintf("primesmin%02d.pla", nVars);
    Pla_ManDumpPla( p, Buffer );
    Pla_ManFree( p );
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


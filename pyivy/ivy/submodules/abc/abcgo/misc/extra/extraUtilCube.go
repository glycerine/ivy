package extra

import "fmt"

/**CFile****************************************************************

  FileName    [extraUtilCube.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Permutation computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilCube.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>
//#include "misc/vec/vec.h"
//#include "misc/vec/vecHsh.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Abc_StatePush(vData *Vec_Int_t, pState string, k int32) { var i int32; for  i = 0; i < 6; i++  { Vec_IntWriteEntry(vData, 6*k+i, ((int*)pState)[i]); }  }
func Abc_StatePerm(pState string, pPerm string, pRes string) { var i int32; for  i = 0; i < 24; i++  { pRes[i] = pState[int32(pPerm[i])]; }                    }
func Abc_StatePrint(pState string) { var i int32; for  i = 0; i < 24; i++  { fmt.Printf(" %2d", pState[i]); }
 fmt.Printf("\n");          }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_EnumerateCubeStates2() {
    var pXYZ [3][9][2]int32 = [3][9][2]int32{
        { {3, 5}, {3,17}, {3,15}, {1, 6}, {1,16}, {1,14}, {2, 4}, {2,18}, {2,13} },
        { {2,14}, {2,24}, {2,12}, {3,13}, {3,23}, {3,10}, {1,15}, {1,22}, {1,11} },
        { {1,10}, {1, 7}, {1, 4}, {3,12}, {3, 9}, {3, 6}, {2,11}, {2, 8}, {2, 5} }  };

    var vData *Vec_Int_t  = Vec_IntStart( 6 * (1 << 22) );            // 96 MB
    var vHash *Hsh_IntMan_t  = Hsh_IntManStart( vData, 6, 1 << 22 );  // 32 MB
    var i, k, v, RetValue, Beg, End, Counter int32;
    var pStart [24]byte;
    var pFirst [9][24]byte;
    var clk abctime  = Abc_Clock();
    fmt.Printf("Enumerating states of 2x2x2 cube.\n");
    // init state
    for  v = 0; v < 24; v++  {
        pStart[v] = v;
    }
    Abc_StatePush( vData, pStart, Counter );
    RetValue = Hsh_IntManAdd( vHash, Counter );
    assert(RetValue == Counter, "RetValue == Counter");
    Counter++;
    // first nine states
    for  i = 0; i < 3; i++  {
        memcpy( pFirst[i], pStart, 24 );
        for  v = 0; v < 9; v++  {
            // inlining of macro ABC_SWAP
             var t byte  = pFirst[i][pXYZ[i][v][0]-1]; pFirst[i][pXYZ[i][v][0]-1] = pFirst[i][pXYZ[i][v][1]-1]; pFirst[i][pXYZ[i][v][1]-1] = t; };
        Abc_StatePush( vData, pFirst[i], Counter );
        RetValue = Hsh_IntManAdd( vHash, Counter );
        assert(RetValue == Counter, "RetValue == Counter");
        Counter++;
        //Abc_StatePrint( pFirst[i] );

        memcpy( pFirst[3+i], pFirst[i], 24 );
        for  v = 0; v < 9; v++  {
            // inlining of macro ABC_SWAP
             var t byte  = pFirst[3+i][pXYZ[i][v][0]-1]; pFirst[3+i][pXYZ[i][v][0]-1] = pFirst[3+i][pXYZ[i][v][1]-1]; pFirst[3+i][pXYZ[i][v][1]-1] = t; };
        Abc_StatePush( vData, pFirst[3+i], Counter );
        RetValue = Hsh_IntManAdd( vHash, Counter );
        assert(RetValue == Counter, "RetValue == Counter");
        Counter++;
        //Abc_StatePrint( pFirst[3+i] );

        memcpy( pFirst[6+i], pFirst[3+i], 24 );
        for  v = 0; v < 9; v++  {
            // inlining of macro ABC_SWAP
             var t byte  = pFirst[6+i][pXYZ[i][v][0]-1]; pFirst[6+i][pXYZ[i][v][0]-1] = pFirst[6+i][pXYZ[i][v][1]-1]; pFirst[6+i][pXYZ[i][v][1]-1] = t; };
        Abc_StatePush( vData, pFirst[6+i], Counter );
        RetValue = Hsh_IntManAdd( vHash, Counter );
        assert(RetValue == Counter, "RetValue == Counter");
        Counter++;
        //Abc_StatePrint( pFirst[6+i] );
    }
    fmt.Printf("Iter %2d -> %8d   ", 0, 1);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    fmt.Printf("Iter %2d -> %8d   ", 1, Counter);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Beg = 1;  End = 10;
    for  i = 2; i <= 100; i++  {
        if Beg == End {
            break; }
        for  k = Beg; k < End; k++  {
            for  v = 0; v < 9; v++  {
                Abc_StatePerm( (char *)Vec_IntEntryP(vData, 6*k), pFirst[v], (char *)Vec_IntEntryP(vData, 6*Counter) );
                RetValue = Hsh_IntManAdd( vHash, Counter );
                if RetValue == Counter {
                    Counter++; }
                if Counter == (1<<22) {
                    fmt.Printf("Did not converge.  ");
                    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
                    return;
                }
            }
        }
        Beg = End;  End = Counter;
        fmt.Printf("Iter %2d -> %8d   ", i, Counter);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    Hsh_IntManStop( vHash );
    Vec_IntFree( vData );
}

/*
Enumerating states of 2x2x2 cube.
Iter  0 ->        1   Time =     0.00 sec
Iter  1 ->       10   Time =     0.00 sec
Iter  2 ->       64   Time =     0.00 sec
Iter  3 ->      385   Time =     0.00 sec
Iter  4 ->     2232   Time =     0.00 sec
Iter  5 ->    12224   Time =     0.00 sec
Iter  6 ->    62360   Time =     0.02 sec
Iter  7 ->   289896   Time =     0.09 sec
Iter  8 ->  1159968   Time =     0.90 sec
Iter  9 ->  3047716   Time =    11.62 sec
Iter 10 ->  3671516   Time =    52.00 sec
Iter 11 ->  3674160   Time =    70.38 sec
Iter 12 ->  3674160   Time =    70.48 sec 
*/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_DataHasBit(p *word, i word) int32 { return (p[(i)>>6] & (((word)1)<<((i) & 63))) > 0; }
func Abc_DataXorBit(p *word, i word) { p[(i)>>6] ^= (((word)1)<<((i) & 63));             }
func Abc_DataGetCube(w word, i int32) int32 { return (w >> (5*i)) & 31;            }
func Abc_DataXorCube(w word, i int32, c int32) word { return w ^ (((word)c) << (5*i));     }
var spStatic_Abc_CubeGenerateSign_Var2Cor [24]int32 = [24]int32{ 0, 0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4, 5, 5, 5, -1, -1, -1, 6, 6, 6 };
var spStatic_Abc_CubeGenerateSign_Var2Per [24]int32 = [24]int32{ 0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2, -1, -1, -1, 0, 1, 2 };
func Abc_CubeGenerateSign(pState string) word {
    
    
    var Res word; 
    Res = Abc_DataXorCube( Res, 0, (spStatic_Abc_CubeGenerateSign_Var2Cor[int32(pState[ 0])] << 2) | spStatic_Abc_CubeGenerateSign_Var2Per[int32(pState[ 0])] );
    Res = Abc_DataXorCube( Res, 1, (spStatic_Abc_CubeGenerateSign_Var2Cor[int32(pState[ 3])] << 2) | spStatic_Abc_CubeGenerateSign_Var2Per[int32(pState[ 3])] );
    Res = Abc_DataXorCube( Res, 2, (spStatic_Abc_CubeGenerateSign_Var2Cor[int32(pState[ 6])] << 2) | spStatic_Abc_CubeGenerateSign_Var2Per[int32(pState[ 6])] );
    Res = Abc_DataXorCube( Res, 3, (spStatic_Abc_CubeGenerateSign_Var2Cor[int32(pState[ 9])] << 2) | spStatic_Abc_CubeGenerateSign_Var2Per[int32(pState[ 9])] );
    Res = Abc_DataXorCube( Res, 4, (spStatic_Abc_CubeGenerateSign_Var2Cor[int32(pState[12])] << 2) | spStatic_Abc_CubeGenerateSign_Var2Per[int32(pState[12])] );
    Res = Abc_DataXorCube( Res, 5, (spStatic_Abc_CubeGenerateSign_Var2Cor[int32(pState[15])] << 2) | spStatic_Abc_CubeGenerateSign_Var2Per[int32(pState[15])] );
    Res = Abc_DataXorCube( Res, 6, (spStatic_Abc_CubeGenerateSign_Var2Cor[int32(pState[21])] << 2) | spStatic_Abc_CubeGenerateSign_Var2Per[int32(pState[21])] );
    return Res;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_EnumerateCubeStates() {
    var pXYZ [3][9][2]int32 = [3][9][2]int32{
        { {3, 5}, {3,17}, {3,15}, {1, 6}, {1,16}, {1,14}, {2, 4}, {2,18}, {2,13} },
        { {2,14}, {2,24}, {2,12}, {3,13}, {3,23}, {3,10}, {1,15}, {1,22}, {1,11} },
        { {1,10}, {1, 7}, {1, 4}, {3,12}, {3, 9}, {3, 6}, {2,11}, {2, 8}, {2, 5} }  };

    var vData *Vec_Int_t  = Vec_IntStart( 6 * (1 << 22) );  // 96 MB
    // inlining of macro ABC_CALLOC
    word State, * pHash = ((word *) calloc((size_t)(1 << 29), sizeof(word)));  //  4 GB
    var i, k, v, Beg, End, Counter int32;
    var pPrev string;
    var pState string;
    var pStart [24]byte;
    var pFirst [9][24]byte;
    var clk abctime  = Abc_Clock();
    fmt.Printf("Enumerating states of 2x2x2 cube.\n");
    // init state
    for  v = 0; v < 24; v++  {
        pStart[v] = v;
    }
    Abc_StatePush( vData, pStart, Counter );
    Abc_DataXorBit( pHash, Abc_CubeGenerateSign(pStart) );
    Counter++;
    // first nine states
    for  i = 0; i < 3; i++  {
        memcpy( pFirst[i], pStart, 24 );
        for  v = 0; v < 9; v++  {
            // inlining of macro ABC_SWAP
             var t byte  = pFirst[i][pXYZ[i][v][0]-1]; pFirst[i][pXYZ[i][v][0]-1] = pFirst[i][pXYZ[i][v][1]-1]; pFirst[i][pXYZ[i][v][1]-1] = t; };
        Abc_StatePush( vData, pFirst[i], Counter );
        Abc_DataXorBit( pHash, Abc_CubeGenerateSign(pFirst[i]) );
        Counter++;
        //Abc_StatePrint( pFirst[i] );

        memcpy( pFirst[3+i], pFirst[i], 24 );
        for  v = 0; v < 9; v++  {
            // inlining of macro ABC_SWAP
             var t byte  = pFirst[3+i][pXYZ[i][v][0]-1]; pFirst[3+i][pXYZ[i][v][0]-1] = pFirst[3+i][pXYZ[i][v][1]-1]; pFirst[3+i][pXYZ[i][v][1]-1] = t; };
        Abc_StatePush( vData, pFirst[3+i], Counter );
        Abc_DataXorBit( pHash, Abc_CubeGenerateSign(pFirst[3+i]) );
        Counter++;
        //Abc_StatePrint( pFirst[3+i] );

        memcpy( pFirst[6+i], pFirst[3+i], 24 );
        for  v = 0; v < 9; v++  {
            // inlining of macro ABC_SWAP
             var t byte  = pFirst[6+i][pXYZ[i][v][0]-1]; pFirst[6+i][pXYZ[i][v][0]-1] = pFirst[6+i][pXYZ[i][v][1]-1]; pFirst[6+i][pXYZ[i][v][1]-1] = t; };
        Abc_StatePush( vData, pFirst[6+i], Counter );
        Abc_DataXorBit( pHash, Abc_CubeGenerateSign(pFirst[6+i]) );
        Counter++;
        //Abc_StatePrint( pFirst[6+i] );
    }
    fmt.Printf("Iter %2d -> %8d   ", 0, 1);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    fmt.Printf("Iter %2d -> %8d   ", 1, Counter);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Beg = 1;  End = 10;
    for  i = 2; i <= 100; i++  {
        if Beg == End {
            break; }
        for  k = Beg; k < End; k++  {
            pPrev = (char *)Vec_IntEntryP(vData, 6*k);
            for  v = 0; v < 9; v++  {
                pState = (char *)Vec_IntEntryP(vData, 6*Counter);
                Abc_StatePerm( pPrev, pFirst[v], pState );
                State = Abc_CubeGenerateSign( pState );
                if !(Abc_DataHasBit(pHash, State) != 0) {
                    Abc_DataXorBit(pHash, State);
Counter++; }
                if Counter == (1<<22) {
                    fmt.Printf("Did not converge.  ");
                    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
                    return;
                }
            }
        }
        Beg = End;  End = Counter;
        fmt.Printf("Iter %2d -> %8d   ", i, Counter);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    Vec_IntFree( vData );
    // inlining of macro ABC_FREE
    ((pHash) ? (free((char *) (pHash)), (pHash) = 0) : 0);
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


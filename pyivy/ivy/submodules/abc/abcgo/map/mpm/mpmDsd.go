package mpm

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [mpmDsd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Configurable technology mapper.]

  Synopsis    [DSD manipulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 1, 2013.]

  Revision    [$Id: mpmDsd.c,v 1.00 2013/06/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mpmInt.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManPrintDsdStats(p *Mpm_Man_t) {
    var i, Absent int32;
    for  i = 0; i < 595; i++  {
        if p.nCountDsd[i] == 0 {
            Absent++;
            continue;
        }
        if p.pPars.fVeryVerbose {
            fmt.Printf("%5d  :  ", i);
            fmt.Printf("%-20s   ", p.pDsd6[i].pStr);
            fmt.Printf("%8d ", p.nCountDsd[i]);
            fmt.Printf("\n");
        }
    }
    fmt.Printf("Unused classes = %d (%.2f %%).  ", Absent, 100.0 * Absent / 595);
    fmt.Printf("Non-DSD cuts = %d (%.2f %%).  ", p.nNonDsd, 100.0 * p.nNonDsd / p.nCutsMergedAll);
    fmt.Printf("No-match cuts = %d (%.2f %%).\n", p.nNoMatch, 100.0 * p.nNoMatch / p.nCutsMergedAll);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManBuildHashTable(pvConfgRes **Vec_Int_t) *Hsh_IntMan_t {
    var pFile *os.File;
    var pFileName string  = "dsdfuncs6.dat";
    var RetValue int32;
    var size int32 = Extra_FileSize( pFileName ) / 12;  // 2866420
    var vTruthRes *Vec_Wrd_t  = Vec_WrdAlloc( size );
    var vConfgRes *Vec_Int_t  = Vec_IntAlloc( size );
    var pHash *Hsh_IntMan_t;

    pFile = os.Open(pFileName);
    RetValue = fread( Vec_WrdArray(vTruthRes), sizeof(word), size, pFile );
    RetValue = fread( Vec_IntArray(vConfgRes), sizeof(int), size, pFile );
    vTruthRes.nSize = size;
    vConfgRes.nSize = size;
    // create hash table
    pHash = Hsh_WrdManHashArrayStart( vTruthRes, 1 );
    // cleanup
    if pvConfgRes {
        *pvConfgRes = vConfgRes;
    } else {
        Vec_IntFree( vConfgRes ); }
    Vec_WrdFree( vTruthRes );
//    Hsh_IntManStop( pHash );
    return pHash;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManPrintPerm(s uint32) {
    var i int32;
    for  i = 0; i < 6; i++  {
        fmt.Printf("%d ", (s >> (3*i)) & 7);
    }
    fmt.Printf("  ");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManPrecomputePerms(p *Mpm_Man_t) {
    var nVars int32  = 6;
    // 0(1:1) 1(2:1) 2(4:2) 3(10:6) 4(33:23) 5(131:98) 6(595:464)
    var nClasses [7]int32 = [7]int32{ 1, 2, 4, 10, 33, 131, 595 };
    var nPerms int32  = Extra_Factorial( nVars );
//    int nSwaps = (1 << nVars);
    var pComp, pPerm *int32;
    var i, k, x, One, OneCopy, Num int32;
    var vVars *Vec_Int_t;
    var clk abctime  = Abc_Clock();
    assert(p.pDsd6 == nil, "p.pDsd6 == nil");
    p.pDsd6 = s_DsdClass6;
    // precompute schedules
    pComp = Extra_GreyCodeSchedule( nVars );
    pPerm = Extra_PermSchedule( nVars );
    // map numbers into perms
    p.vMap2Perm = Vec_IntStartFull( (1<<(3*nVars)) );
    // store permutations
    One = 0;
    for  x = 0; x < nVars; x++  {
        p.Perm6[0][x] = byte(x);
        One |= (x << (3*x));
    }
//    Vec_IntWriteEntry( p->vMap2Perm, One, 0 );
    OneCopy = One;
    for  k = 0; k < nPerms; k++  {
        if k > 0 {
        for  x = 0; x < nVars; x++  {
            p.Perm6[k][x] = p.Perm6[k-1][x];
        } }
        // inlining of macro ABC_SWAP
        { var t byte  = p.Perm6[k][pPerm[k]]; p.Perm6[k][pPerm[k]] = p.Perm6[k][pPerm[k]+1]; p.Perm6[k][pPerm[k]+1] = t; };

        Num = ( (One >> (3*(pPerm[k]  ))) ^ (One >> (3*(pPerm[k]+1))) ) & 7;
        One ^=  (Num << (3*(pPerm[k]  )));
        One ^=  (Num << (3*(pPerm[k]+1)));

        Vec_IntWriteEntry( p.vMap2Perm, One, k );
        
//        Mpm_ManPrintPerm( One );
//        for ( x = 0; x < nVars; x++ )
//            printf( "%d ", p->Perm6[k][x] );
//        printf( "\n" );
    }
    assert(OneCopy == One, "OneCopy == One");
    // fill in the gaps
    vVars = Vec_IntAlloc( 6 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vMap2Perm)) && (((Num) = Vec_IntEntry(p.vMap2Perm, i)), 1); i++  {
        // mark used variables
        var Count int32;
        One = i;
        Vec_IntFill( vVars, 6, 0 );
        for  k = 0; k < nVars; k++  {
            var iVar int32  = ((One >> (3*k)) & 7);
            if iVar >= nVars && iVar < 7 {
                break; }
            if iVar != 7 {
                if Vec_IntEntry( vVars, iVar ) == 1 {
                    break; }
                Vec_IntWriteEntry( vVars, iVar, 1 );
                Count++;
            }
        }
        // skip ones with dups and complete
        if k < nVars || Count == nVars {
            continue; }
        // find unused variables
        for  x = k = 0; k < 6; k++  {
            if Vec_IntEntry(vVars, k) == 0 {
                Vec_IntWriteEntry( vVars, x++, k ); }
        }
        Vec_IntShrink( vVars, x );
        // fill in used variables
        x = 0;
        for  k = 0; k < nVars; k++  {
            var iVar int32  = ((One >> (3*k)) & 7);
            if iVar == 7 {
                One ^= ((Vec_IntEntry(vVars, x++) ^ 7) << (3*k)); }
        }
        assert(x == Vec_IntSize(vVars), "x == Vec_IntSize(vVars)");
        // save this one
        assert(Vec_IntEntry( p.vMap2Perm, One ) != -1, "Vec_IntEntry( p.vMap2Perm, One ) != -1");
        Vec_IntWriteEntry( p.vMap2Perm, i, Vec_IntEntry(p.vMap2Perm, One) );
/*
        // mapping
        Mpm_ManPrintPerm( i );
        printf( "->  " );
        Mpm_ManPrintPerm( One );
        printf( "\n" );
*/
    }
    Vec_IntFree( vVars );

    // store permuted truth tables
    assert(p.vPerm6 == nil, "p.vPerm6 == nil");
    p.vPerm6 = Vec_WrdAlloc( nPerms * 595 );
    for  i = 0; i < nClasses[nVars]; i++  {
        var uTruth word  = s_DsdClass6[i].uTruth;
        for  k = 0; k < nPerms; k++  {
            uTruth = Abc_Tt6SwapAdjacent( uTruth, pPerm[k] );
            Vec_WrdPush( p.vPerm6, uTruth );
        }
        assert(uTruth == s_DsdClass6[i].uTruth, "uTruth == s_DsdClass6[i].uTruth");
    }
    pPerm = nil;
    pComp = nil;
    // build hash table
    p.pHash = Mpm_ManBuildHashTable( &p.vConfgRes );
    Abc_PrintTime( 1, "Setting up DSD information", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_CutTruthFromDsd(pMan *Mpm_Man_t, pCut *Mpm_Cut_t, iClass int32) word {
    var i int32;
    var uTruth word  = pMan.pDsd6[iClass].uTruth;
    assert(pMan.pDsd6[iClass].nVars == int32(pCut.nLeaves), "pMan.pDsd6[iClass].nVars == int32(pCut.nLeaves)");
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        if Abc_LitIsCompl(pCut.pLeaves[i]) {
            uTruth = Abc_Tt6Flip( uTruth, i ); }
    } 
    return uTruth;
}

/**Function*************************************************************

  Synopsis    [Checks hash table for DSD class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_CutCheckDsd6(p *Mpm_Man_t, t word) int32 {
    var fCompl, Entry, Config int32;
    if fCompl = (t & 1) {
        t = ~t; }
    Entry = *Hsh_IntManLookup( p.pHash, (unsigned *)&t );
    if Entry == -1 {
        return -1; }
    Config = Vec_IntEntry( p.vConfgRes, Entry );
    if fCompl != 0 {
        Config ^= (1 << 16); }
    return Config;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_CutComputeDsd6(p *Mpm_Man_t, pCut *Mpm_Cut_t, pCut0 *Mpm_Cut_t, pCut1 *Mpm_Cut_t, pCutC *Mpm_Cut_t, fCompl0 int32, fCompl1 int32, fComplC int32, Type int32) int32 {
    var fVerbose int32;
    var i, Config, iClass, fCompl int32;
    var pLeavesNew [6]int32 = [6]int32{ -1, -1, -1, -1, -1, -1 };
    var t word;
    if pCutC == nil {
        var t0, t1 word;
        var iClass0 int32   = Abc_Lit2Var(pCut0.iFunc);
        var iClass1 int32   = Abc_Lit2Var(pCut1.iFunc);
        var Truth0 word   = p.pDsd6[iClass0].uTruth;
        var Perm1 int32     = Vec_IntEntry( p.vMap2Perm, p.uPermMask[1] );
        var Truth1p word  = Vec_WrdEntry( p.vPerm6, iClass1 * 720 + Perm1 );
        if p.uComplMask[1] {
            for  i = 0; i < 6; i++  {
                if (p.uComplMask[1] >> i) & 1 {
                    Truth1p = Abc_Tt6Flip( Truth1p, i ); }
            }            
        }
        if fCompl0 ^ pCut0.fCompl ^ Abc_LitIsCompl(pCut0.iFunc) {
            t0 = ~Truth0;
        } else {
            t0 = Truth0;
        }
        if fCompl1 ^ pCut1.fCompl ^ Abc_LitIsCompl(pCut1.iFunc) {
            t1 = ~Truth1p;
        } else {
            t1 = Truth1p;
        }
        if Type == 1 {
            t = t0 & t1;
        } else if Type == 2 {
            t = t0 ^ t1;
        } else { assert(false, "0"); }

if fVerbose != 0 {
Mpm_ManPrintPerm( p.uPermMask[1] );               fmt.Printf("\n");
Kit_DsdPrintFromTruth( (unsigned *)&Truth0, 6 );   fmt.Printf("\n");
Kit_DsdPrintFromTruth( (unsigned *)&Truth1p, 6 );  fmt.Printf("\n");
Kit_DsdPrintFromTruth( (unsigned *)&t, 6 );        fmt.Printf("\n");
}
    
    } else {
        var t0, t1, tC word;
        var iClass0 int32   = Abc_Lit2Var(pCut0.iFunc);
        var iClass1 int32   = Abc_Lit2Var(pCut1.iFunc);
        var iClassC int32   = Abc_Lit2Var(pCutC.iFunc);
        var Truth0 word   = p.pDsd6[iClass0].uTruth;
        var Perm1 int32     = Vec_IntEntry( p.vMap2Perm, p.uPermMask[1] );
        var PermC int32     = Vec_IntEntry( p.vMap2Perm, p.uPermMask[2] );
        var Truth1p word  = Vec_WrdEntry( p.vPerm6, iClass1 * 720 + Perm1 );
        var TruthCp word  = Vec_WrdEntry( p.vPerm6, iClassC * 720 + PermC );
        if p.uComplMask[1] {
            for  i = 0; i < 6; i++  {
                if (p.uComplMask[1] >> i) & 1 {
                    Truth1p = Abc_Tt6Flip( Truth1p, i ); }
            }            
        }
        if p.uComplMask[2] {
            for  i = 0; i < 6; i++  {
                if (p.uComplMask[2] >> i) & 1 {
                    TruthCp = Abc_Tt6Flip( TruthCp, i ); }
            }            
        }
        if fCompl0 ^ pCut0.fCompl ^ Abc_LitIsCompl(pCut0.iFunc) {
            t0 = ~Truth0;
        } else {
            t0 = Truth0;
        }
        if fCompl1 ^ pCut1.fCompl ^ Abc_LitIsCompl(pCut1.iFunc) {
            t1 = ~Truth1p;
        } else {
            t1 = Truth1p;
        }
        if fComplC ^ pCutC.fCompl ^ Abc_LitIsCompl(pCutC.iFunc) {
            tC = ~TruthCp;
        } else {
            tC = TruthCp;
        }
        t = (tC & t1) | (~tC & t0);
    }

    // find configuration
    Config = Mpm_CutCheckDsd6( p, t );
    if Config == -1 {
        p.nNonDsd++;
        return 0;
    }

    // get the class
    iClass = Config >> 17;
    fCompl = (Config >> 16) & 1;
    Config &= 0xFFFF;

    // check if the gate exists
    if p.pPars.fMap4Gates {
        if Vec_IntSize(Vec_WecEntry(p.vNpnConfigs, iClass)) == 0 {
            p.nNoMatch++;
            return 0;
        }
    }

    // set the function
    pCut.iFunc = Abc_Var2Lit( iClass, fCompl );

if fVerbose != 0 {
Mpm_CutPrint( pCut0 );
Mpm_CutPrint( pCut1 );
Mpm_CutPrint( pCut );
}

    // update cut
    assert((Config >> 6) < 720, "(Config >> 6) < 720");
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        pLeavesNew[int32((p.Perm6[Config >> 6][i]))] = Abc_LitNotCond( pCut.pLeaves[i], (Config >> i) & 1 );
    }
    pCut.nLeaves = p.pDsd6[iClass].nVars;
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        assert(pLeavesNew[i] != -1, "pLeavesNew[i] != -1");
    }
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        pCut.pLeaves[i] = pLeavesNew[i];
    }
    p.nCountDsd[iClass]++;
    p.nSmallSupp += int32((pCut.nLeaves < 2));

if fVerbose != 0 {
fmt.Printf("Computed ");
Mpm_CutPrint( pCut );
fmt.Printf("\n");
}
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


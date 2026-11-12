package io

import (
	"fmt"
	"strconv"
)

import "os"

/**CFile****************************************************************

  FileName    [ioReadPlaMo.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedure to read network from file.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: ioReadPlaMo.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ioAbc.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Mop_Man_t struct {
    nIns              int32;
    nOuts              int32;
    nWordsIn              int32;
    nWordsOut              int32;
    vWordsIn       *Vec_Wrd_t;
    vWordsOut       *Vec_Wrd_t;
    vCubes       *Vec_Int_t; 
    vFree       *Vec_Int_t;
}

func Mop_ManIsSopSymb(c byte) int32 {  return c == '0' || c == '1' || c == '-'; }
func Mop_ManIsSpace(c byte) int32 {  return c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f' || c == '\r'; }

func Mop_ManCubeIn(p *Mop_Man_t, i int32) *word { return Vec_WrdEntryP(p.vWordsIn,   p.nWordsIn  * i); }
func Mop_ManCubeOut(p *Mop_Man_t, i int32) *word { return Vec_WrdEntryP(p.vWordsOut,  p.nWordsOut * i); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mop_ManAlloc(nIns int32, nOuts int32, nCubes int32) *Mop_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Mop_Man_t    = &Mop_Man_t{};
    p.nIns         = nIns;
    p.nOuts        = nOuts;
    p.nWordsIn     = Abc_Bit6WordNum( 2 * nIns );
    p.nWordsOut    = Abc_Bit6WordNum( nOuts );
    p.vWordsIn     = Vec_WrdStart( 2 * p.nWordsIn * nCubes );
    p.vWordsOut    = Vec_WrdStart( 2 * p.nWordsOut * nCubes );
    p.vCubes       = Vec_IntAlloc( 2 * nCubes );
    p.vFree        = Vec_IntAlloc( 2 * nCubes );
    return p;
}
func Mop_ManStop(p *Mop_Man_t) {
    Vec_WrdFree( p.vWordsIn );
    Vec_WrdFree( p.vWordsOut );
    Vec_IntFree( p.vCubes );
    Vec_IntFree( p.vFree );
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Reads the file into a character buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mop_ManLoadFile(pFileName string) string {
    var pFile *os.File;
    var nFileSize, RetValue int32;
    var pContents string;
    pFile = os.Open(pFileName);
    if pFile == nil {
        Abc_Print( -1, "Mop_ManLoadFile(): The file is unavailable (absent or open).\n" );
        return nil;
    }
    fseek( pFile, 0, SEEK_END );  
    nFileSize = ftell( pFile ); 
    if nFileSize == 0 {
        Abc_Print( -1, "Mop_ManLoadFile(): The file is empty.\n" );
        return nil;
    }
    // inlining of macro ABC_ALLOC
    pContents = ((char *) malloc(sizeof(char) * (size_t)(nFileSize + 10)));
    rewind( pFile );
    RetValue = fread( pContents, nFileSize, 1, pFile );
    _ = (pFile).Close();
    strcpy( pContents + nFileSize, "\n" );
    return pContents;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mop_ManReadParams(pBuffer string, pnIns *int32, pnOuts *int32) int32 {
    var pIns string   = strstr( pBuffer, ".i " );
    var pOuts string  = strstr( pBuffer, ".o " );
    var pStr string   = pBuffer; var nCubes int32;
    if pIns == nil || pOuts == nil {
        return -1; }
    *pnIns  = strconv.Atoi(pIns + 2);
    *pnOuts = strconv.Atoi(pOuts + 2);
    for  *pStr  {
        nCubes += (*pStr++ == '\n');
    }
    return nCubes;
}
func Mop_ManRead(pFileName string) *Mop_Man_t {
    var p *Mop_Man_t;
    var nIns, nOuts, nCubes, iCube int32;
    var pToken string;
    var pBuffer string = Mop_ManLoadFile( pFileName );
    if pBuffer == nil {
        return nil; }
    nCubes = Mop_ManReadParams( pBuffer, &nIns, &nOuts );
    if nCubes == -1 {
        return nil; }
    p = Mop_ManAlloc( nIns, nOuts, nCubes );
    // get the first cube
    pToken = strtok( pBuffer, "\n" );
    for  pToken  {
        for  Mop_ManIsSpace(*pToken) != 0  {
            pToken++;
        }
        if Mop_ManIsSopSymb(*pToken) != 0 {
            break; }
        pToken = strtok( nil, "\n" );
    }
    // read cubes
    for  iCube = 0; pToken && (Mop_ManIsSopSymb(*pToken) != 0); iCube++  {
        var pTokenCopy string  = pToken;
        var i int32;
        var o int32;
        var nVars [2]int32 = [2]int32{nIns, nOuts};
        var pCube [2]*word = [2]*word{ Mop_ManCubeIn(p, iCube), Mop_ManCubeOut(p, iCube) }; 
        for  o = 0; o < 2; o++  {
            for  Mop_ManIsSpace(*pToken) != 0  {
                pToken++;
            }
            for  i = 0; i < nVars[o]; i, pToken = i+1, pToken+1  {
                if !(Mop_ManIsSopSymb(*pToken) != 0) {
                    fmt.Printf("Cannot read cube %d (%s).\n", iCube+1, pTokenCopy);
                    pBuffer = nil;
                    Mop_ManStop( p );
                    return nil;
                }
                if o == 1 {
                    if *pToken == '1' {
                        Abc_TtSetBit( pCube[o], i ); }
                
                } else if *pToken == '0' {
                    Abc_TtSetBit( pCube[o], 2*i );
                } else if *pToken == '1' {
                    Abc_TtSetBit( pCube[o], 2*i+1 ); }
            }
        }
        assert(iCube < nCubes, "iCube < nCubes");
        Vec_IntPush( p.vCubes, iCube );
        pToken = strtok( nil, "\n" );
    }
    for  ; iCube < 2 * nCubes; iCube++  {
        Vec_IntPush( p.vFree, iCube );
    }
    pBuffer = nil;
    return p;
}
func Mop_ManPrintOne(p *Mop_Man_t, iCube int32) {
    var k int32;
    var Symb [4]byte = [4]byte{ '-', '0', '1', '?' };
    var pCubeIn *word   = Mop_ManCubeIn( p, iCube );
    var pCubeOut *word  = Mop_ManCubeOut( p, iCube );
    for  k = 0; k < p.nIns; k++  {
        fmt.Printf("%c", Symb[Abc_TtGetQua(pCubeIn, k)]);
    }
    fmt.Printf(" ");
    for  k = 0; k < p.nOuts; k++  {
        fmt.Printf("%d", Abc_TtGetBit(pCubeOut, k));
    }
    fmt.Printf("\n");
}
func Mop_ManPrint(p *Mop_Man_t) {
    var i, iCube int32;
    fmt.Printf(".%d\n", p.nIns);
    fmt.Printf(".%d\n", p.nOuts);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCubes)) && (((iCube) = Vec_IntEntry(p.vCubes, i)), 1); i++  {
        Mop_ManPrintOne( p, iCube );
    }
    fmt.Printf(".e\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mop_ManCountOnes(pCube *word, nWords int32) int32 {
    var w, Count int32;
    for  w = 0; w < nWords; w++  {
        Count += Abc_TtCountOnes( pCube[w] );
    }
    return Count;
}
func Mop_ManCountOutputLits(p *Mop_Man_t) int32 {
    var i, iCube, nOutLits int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCubes)) && (((iCube) = Vec_IntEntry(p.vCubes, i)), 1); i++  {
        nOutLits += Mop_ManCountOnes( Mop_ManCubeOut(p, iCube), p.nWordsOut );
    }
    return nOutLits;
}
func Mop_ManCreateGroups(p *Mop_Man_t) *Vec_Wec_t {
    var i, iCube int32;
    var vGroups *Vec_Wec_t  = Vec_WecStart( p.nIns );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCubes)) && (((iCube) = Vec_IntEntry(p.vCubes, i)), 1); i++  {
        Vec_WecPush( vGroups, Mop_ManCountOnes(Mop_ManCubeIn(p, iCube), p.nWordsIn), iCube );
    }
    return vGroups;
}
func Mop_ManUnCreateGroups(p *Mop_Man_t, vGroups *Vec_Wec_t) int32 {
    var i, c1, iCube1 int32;
    var nBefore int32  = Vec_IntSize(p.vCubes);
    var vGroup *Vec_Int_t; 
    Vec_IntClear( p.vCubes );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vGroups)) && (((vGroup) = Vec_WecEntry(vGroups, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  c1 = 0; (c1 < Vec_IntSize(vGroup)) && (((iCube1) = Vec_IntEntry(vGroup, c1)), 1); c1++  {
            if iCube1 != -1 {
                Vec_IntPush( p.vCubes, iCube1 ); }
        }
    }
    return nBefore - Vec_IntSize(p.vCubes);
}
func Mop_ManCheckContain(pBig *word, pSmall *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if pSmall[w] != (pSmall[w] & pBig[w]) {
            return 0; }
    }
    return 1;
}
func Mop_ManRemoveEmpty(p *Mop_Man_t) {
    var w, i, k, iCube int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCubes)) && (((iCube) = Vec_IntEntry(p.vCubes, i)), 1); i++  {
        var pCube *word  = Mop_ManCubeOut( p, iCube );
        for  w = 0; w < p.nWordsOut; w++  {
            if pCube[w] {
                break; }
        }
        if w < p.nWordsOut {
            Vec_IntWriteEntry( p.vCubes, k++, iCube ); }
    }
    Vec_IntShrink( p.vCubes, k );
}

/**Function*************************************************************

  Synopsis    [Count how many times each variable appears in the input parts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mop_ManCollectStats(p *Mop_Man_t) *Vec_Int_t {
    var i int32;
    var v int32;
    var iCube int32;
    var nVars int32 = 32 * p.nWordsIn;
    var vStats *Vec_Int_t  = Vec_IntStart( nVars );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCubes)) && (((iCube) = Vec_IntEntry(p.vCubes, i)), 1); i++  {
        var pCube *word  = Mop_ManCubeIn(p, iCube);
        var nOutLits int32  = Mop_ManCountOnes( Mop_ManCubeOut(p, iCube), p.nWordsOut );
        for  v = 0; v < nVars; v++  {
            if Abc_TtGetQua(pCube, v) {
                Vec_IntAddToEntry( vStats, v, nOutLits ); }
        }
    }
    return vStats;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// knowing that cubes are distance-1, find the different input variable
func Mop_ManFindDiffVar(pCube1 *word, pCube2 *word, nWords int32) int32 {
    var w, i int32;
    for  w = 0; w < nWords; w++  {
        var Xor word  = pCube1[w] ^ pCube2[w];
        for  i = 0; i < 32; i++  {
            if (Xor >> (i << 1)) & 0x3 {
                return w * 32 + i; }
        }
    }
    assert(false, "0");
    return -1;
}
// check containment of input parts of two cubes
func Mop_ManCheckDist1(pCube1 *word, pCube2 *word, nWords int32) int32 {
    var w, fFound1 int32;
    for  w = 0; w < nWords; w++  {
        var Xor word  = pCube1[w] ^ pCube2[w];
        if Xor == 0 { // equal
            continue; }
        if (Xor ^ (Xor >> 1)) & ABC_CONST(0x5555555555555555) { // not pairs
            return 0; }
        Xor &= (Xor >> 1) & ABC_CONST(0x5555555555555555);
        if Xor == 0 { // not equal and not distance-1
            return 0; }
        if fFound1 != 0 { // distance-2 or more
            return 0; }
        if Xor & (Xor-1) { // distance-2 or more
            return 0; }
        fFound1 = 1; // distance 1 so far
    }
    return fFound1;
}
// compresses cubes in the group
func Map_ManGroupCompact(vGroup *Vec_Int_t) {
    var i, Entry, k int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vGroup)) && (((Entry) = Vec_IntEntry(vGroup, i)), 1); i++  {
        if Entry != -1 {
            Vec_IntWriteEntry( vGroup, k++, Entry ); }
    }
    Vec_IntShrink( vGroup, k );
}
// takes cubes with identical literal count and removes duplicates
func Mop_ManRemoveIdentical(p *Mop_Man_t, vGroup *Vec_Int_t) int32 {
    var w, c1, c2, iCube1, iCube2, nEqual int32;
    // inlining of macro Vec_IntForEachEntry
    for  c1 = 0; (c1 < Vec_IntSize(vGroup)) && (((iCube1) = Vec_IntEntry(vGroup, c1)), 1); c1++  {
    if iCube1 != -1 {
        var pCube1Out *word;
        var pCube1 *word = Mop_ManCubeIn( p, iCube1 );
        // inlining of macro Vec_IntForEachEntryStart
        for  c2 = c1+1; (c2 < Vec_IntSize(vGroup)) && (((iCube2) = Vec_IntEntry(vGroup, c2)), 1); c2++  {
        if iCube2 != -1 {
            var pCube2Out *word;
            var pCube2 *word = Mop_ManCubeIn( p, iCube2 );
            if memcmp(pCube1, pCube2, sizeof(word)*p.nWordsIn) {
                continue; }
            // merge cubes
            pCube1Out = Mop_ManCubeOut( p, iCube1 );
            pCube2Out = Mop_ManCubeOut( p, iCube2 );
            for  w = 0; w < p.nWordsOut; w++  {
                pCube1Out[w] |= pCube2Out[w];
            }
            Vec_IntWriteEntry( vGroup, c2, -1 );
            Vec_IntPush( p.vFree, iCube2 );
            nEqual++;
        }
        }
    }
    }
    if nEqual != 0 {
        Map_ManGroupCompact( vGroup ); }
    return nEqual;
}
// reduces the set of pairs
func Mop_ManCompatiblePairs(vPairs *Vec_Int_t, nObjs int32) *Vec_Int_t {
    var i, Entry, Entry2 int32;
    var vCounts *Vec_Int_t  = Vec_IntStart( nObjs );
    var vPairsNew *Vec_Int_t  = Vec_IntAlloc( Vec_IntSize(vPairs) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPairs)) && (((Entry) = Vec_IntEntry(vPairs, i)), 1); i++  {
        Vec_IntAddToEntry( vCounts, Entry, 1 );
    }
    // include pairs which have those that appear only once
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vPairs)) && (((Entry) = Vec_IntEntry(vPairs, i)), 1) && (((Entry2) = Vec_IntEntry(vPairs, i+1)), 1); i += 2  {
        if Vec_IntEntry(vCounts, Entry) == 1 || Vec_IntEntry(vCounts, Entry2) == 1 {
            if Vec_IntEntry(vCounts, Entry) == 1 {
                Vec_IntPushTwo( vPairsNew, Entry, Entry2 );
            } else {
                Vec_IntPushTwo( vPairsNew, Entry2, Entry ); }
            Vec_IntWriteEntry( vCounts, Entry, -1 );
            Vec_IntWriteEntry( vCounts, Entry2, -1 );
        }
    }
    // add those remaining pairs that are both present
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vPairs)) && (((Entry) = Vec_IntEntry(vPairs, i)), 1) && (((Entry2) = Vec_IntEntry(vPairs, i+1)), 1); i += 2  {
        if Vec_IntEntry(vCounts, Entry) > 0 && Vec_IntEntry(vCounts, Entry2) > 0 {
            Vec_IntPushTwo( vPairsNew, Entry, Entry2 );
            Vec_IntWriteEntry( vCounts, Entry, -1 );
            Vec_IntWriteEntry( vCounts, Entry2, -1 );
        }
    }
    // add remaining pairs
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vPairs)) && (((Entry) = Vec_IntEntry(vPairs, i)), 1) && (((Entry2) = Vec_IntEntry(vPairs, i+1)), 1); i += 2  {
        if Vec_IntEntry(vCounts, Entry) > 0 || Vec_IntEntry(vCounts, Entry2) > 0 {
            if Vec_IntEntry(vCounts, Entry) > 0 {
                Vec_IntPushTwo( vPairsNew, Entry, Entry2 );
            } else {
                Vec_IntPushTwo( vPairsNew, Entry2, Entry ); }
            Vec_IntWriteEntry( vCounts, Entry, -1 );
            Vec_IntWriteEntry( vCounts, Entry2, -1 );
        }
    }
    Vec_IntFree( vCounts );
    // verify the result
    if false {
        var vTemp1 *Vec_Int_t  = Vec_IntDup( vPairs );
        var vTemp2 *Vec_Int_t  = Vec_IntDup( vPairsNew );
        Vec_IntUniqify( vTemp1 );
        Vec_IntUniqify( vTemp2 );
        assert(Vec_IntEqual( vTemp1, vTemp2 ), "Vec_IntEqual( vTemp1, vTemp2 )");
        Vec_IntFree( vTemp1 );
        Vec_IntFree( vTemp2 );
    }
    return vPairsNew;
}
// detects pairs of distance-1 cubes with identical outputs
func Mop_ManFindDist1Pairs(p *Mop_Man_t, vGroup *Vec_Int_t) *Vec_Int_t {
    var c1, c2, iCube1, iCube2 int32;
    var vPairs *Vec_Int_t  = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntry
    for  c1 = 0; (c1 < Vec_IntSize(vGroup)) && (((iCube1) = Vec_IntEntry(vGroup, c1)), 1); c1++  {
        var pCube1Out *word;
        var pCube1 *word = Mop_ManCubeIn( p, iCube1 );
        // inlining of macro Vec_IntForEachEntryStart
        for  c2 = c1+1; (c2 < Vec_IntSize(vGroup)) && (((iCube2) = Vec_IntEntry(vGroup, c2)), 1); c2++  {
            var pCube2Out *word;
            var pCube2 *word = Mop_ManCubeIn( p, iCube2 );
            if !(Mop_ManCheckDist1(pCube1, pCube2, p.nWordsIn) != 0) {
                continue; }
            pCube1Out = Mop_ManCubeOut( p, iCube1 );
            pCube2Out = Mop_ManCubeOut( p, iCube2 );
            if !memcmp(pCube1Out, pCube2Out, sizeof(word)*p.nWordsOut) {
                Vec_IntPushTwo( vPairs, c1, c2 ); }
        }
    }
    return vPairs;
}
// merge distance-1 with identical output part
func Mop_ManMergeDist1Pairs(p *Mop_Man_t, vGroup *Vec_Int_t, vGroupPrev *Vec_Int_t, vStats *Vec_Int_t, nLimit int32) int32 {
    var vPairs *Vec_Int_t     = Mop_ManFindDist1Pairs( p, vGroup );
    var vPairsNew *Vec_Int_t  = Mop_ManCompatiblePairs( vPairs, Vec_IntSize(vGroup) );
    var nCubes int32  = Vec_IntSize(vGroup) + Vec_IntSize(vGroupPrev);
    var w, i, c1, c2, iCubeNew, iVar int32;
    // move cubes to the previous group
    var pCube, pCube1, pCube2 *word;
    var vToFree *Vec_Int_t  = Vec_IntAlloc( Vec_IntSize(vPairsNew) );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vPairsNew)) && (((c1) = Vec_IntEntry(vPairsNew, i)), 1) && (((c2) = Vec_IntEntry(vPairsNew, i+1)), 1); i += 2  {
        pCube1 = Mop_ManCubeIn( p, Vec_IntEntry(vGroup, c1) );
        pCube2 = Mop_ManCubeIn( p, Vec_IntEntry(vGroup, c2) );
        assert((Mop_ManCheckDist1(pCube1, pCube2, p.nWordsIn) != 0), "Mop_ManCheckDist1(pCube1, pCube2, p.nWordsIn)");

        // skip those cubes that have frequently appearing variables
        iVar = Mop_ManFindDiffVar( pCube1, pCube2, p.nWordsIn );
        if Vec_IntEntry( vStats, iVar ) > nLimit {
            continue; }
        Vec_IntPush( vToFree, c1 );
        Vec_IntPush( vToFree, c2 );

        iCubeNew  = Vec_IntPop( p.vFree );
        pCube  = Mop_ManCubeIn( p, iCubeNew );
        for  w = 0; w < p.nWordsIn; w++  {
            pCube[w] = pCube1[w] & pCube2[w];
        }

        pCube  = Mop_ManCubeOut( p, iCubeNew );
        pCube1 = Mop_ManCubeOut( p, Vec_IntEntry(vGroup, c1) );
        pCube2 = Mop_ManCubeOut( p, Vec_IntEntry(vGroup, c2) );

        assert(!memcmp(pCube1, pCube2, sizeof(word)*p.nWordsOut), "!memcmp(pCube1, pCube2, sizeof(word)*p.nWordsOut)");
        for  w = 0; w < p.nWordsOut; w++  {
            pCube[w] = pCube1[w];
        }

        Vec_IntPush( vGroupPrev, iCubeNew );
    }
//    Vec_IntForEachEntry( vPairsNew, c1, i )
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vToFree)) && (((c1) = Vec_IntEntry(vToFree, i)), 1); i++  {
        if Vec_IntEntry(vGroup, c1) == -1 {
            continue; }
        Vec_IntPush( p.vFree, Vec_IntEntry(vGroup, c1) );
        Vec_IntWriteEntry( vGroup, c1, -1 );
    }
    Vec_IntFree( vToFree );
    if Vec_IntSize(vPairsNew) > 0 {
        Map_ManGroupCompact( vGroup ); }
    Vec_IntFree( vPairs );
    Vec_IntFree( vPairsNew );
    return nCubes - Vec_IntSize(vGroup) - Vec_IntSize(vGroupPrev);
}
// merge distance-1 with contained output part
func Mop_ManMergeDist1Pairs2(p *Mop_Man_t, vGroup *Vec_Int_t, vGroupPrev *Vec_Int_t) int32 {
    var w, c1, c2, iCube1, iCube2, Count int32;
    // inlining of macro Vec_IntForEachEntry
    for  c1 = 0; (c1 < Vec_IntSize(vGroup)) && (((iCube1) = Vec_IntEntry(vGroup, c1)), 1); c1++  {
    if iCube1 != -1 {
        var pCube1Out *word;
        var pCube1 *word = Mop_ManCubeIn( p, iCube1 );
        // inlining of macro Vec_IntForEachEntryStart
        for  c2 = c1+1; (c2 < Vec_IntSize(vGroup)) && (((iCube2) = Vec_IntEntry(vGroup, c2)), 1); c2++  {
        if iCube2 != -1 {
            var pCube2Out *word;
            var pCube2 *word = Mop_ManCubeIn( p, iCube2 );
            if !(Mop_ManCheckDist1(pCube1, pCube2, p.nWordsIn) != 0) {
                continue; }
            pCube1Out = Mop_ManCubeOut( p, iCube1 );
            pCube2Out = Mop_ManCubeOut( p, iCube2 );
            assert(memcmp(pCube1Out, pCube2Out, sizeof(word)*p.nWordsOut), "memcmp(pCube1Out, pCube2Out, sizeof(word)*p.nWordsOut)");
            if Mop_ManCheckContain(pCube1Out, pCube2Out, p.nWordsOut) != 0 { // pCube1 has more outputs
            
                // update the input part
                for  w = 0; w < p.nWordsIn; w++  {
                    pCube2[w] &= pCube1[w];
                }
                // sharp the output part
                for  w = 0; w < p.nWordsOut; w++  {
                    pCube1Out[w] &= ~pCube2Out[w];
                }
                // move to another group
                Vec_IntPush( vGroupPrev, iCube2 );
                Vec_IntWriteEntry( vGroup, c2, -1 );
                Count++;
            
            } else if Mop_ManCheckContain(pCube2Out, pCube1Out, p.nWordsOut) != 0 { // pCube2 has more outputs
            
                // update the input part
                for  w = 0; w < p.nWordsIn; w++  {
                    pCube1[w] &= pCube2[w];
                }
                // sharp the output part
                for  w = 0; w < p.nWordsOut; w++  {
                    pCube2Out[w] &= ~pCube1Out[w];
                }
                // move to another group
                Vec_IntPush( vGroupPrev, iCube1 );
                Vec_IntWriteEntry( vGroup, c1, -1 );
                Count++;
            }
        }
        }
    }
    }
    if Count != 0 {
        Map_ManGroupCompact( vGroup ); }
    return Count;
}
func Mop_ManMergeDist1All(p *Mop_Man_t, vGroups *Vec_Wec_t, vStats *Vec_Int_t, nLimit int32) int32 {
    var vGroup *Vec_Int_t;
    var i, nEqual, nReduce, Count int32;
    // inlining of macro Vec_WecForEachLevelReverse
    for  i = Vec_WecSize(vGroups)-1; (i >= 0) && (((vGroup) = Vec_WecEntry(vGroups, i)), 1); i--  {
        if Vec_IntSize(vGroup) == 0 {
            continue; }
        if i == 0 {
            fmt.Printf("Detected constant-1 cover.\n");
            fflush( os.Stdout );
            return -1;
        }
        nEqual  = Mop_ManRemoveIdentical( p, vGroup );
        nReduce = Mop_ManMergeDist1Pairs( p, vGroup, Vec_WecEntry(vGroups, i-1), vStats, nLimit );
        //Mop_ManMergeDist1Pairs2( p, vGroup, Vec_WecEntry(vGroups, i-1) );
        Count  += nEqual + nReduce;
        //printf( "Group %3d : Equal =%5d. Reduce =%5d.\n", i, nEqual, nReduce );
    }
    return Count;
}
// reduce contained cubes
func Mop_ManMergeContainTwo(p *Mop_Man_t, vGroup *Vec_Int_t, vGroup2 *Vec_Int_t) int32 {
    var w, c1, c2, iCube1, iCube2, Count int32;
    // inlining of macro Vec_IntForEachEntry
    for  c1 = 0; (c1 < Vec_IntSize(vGroup)) && (((iCube1) = Vec_IntEntry(vGroup, c1)), 1); c1++  {
        var pCube1Out *word;
        var pCube1 *word = Mop_ManCubeIn( p, iCube1 );
        // inlining of macro Vec_IntForEachEntry
        for  c2 = 0; (c2 < Vec_IntSize(vGroup2)) && (((iCube2) = Vec_IntEntry(vGroup2, c2)), 1); c2++  {
        if iCube2 != -1 {
            var pCube2Out *word;
            var pCube2 *word = Mop_ManCubeIn( p, iCube2 );
            if !(Mop_ManCheckContain(pCube2, pCube1, p.nWordsIn) != 0) {
                continue; }
            pCube1Out = Mop_ManCubeOut( p, iCube1 );
            pCube2Out = Mop_ManCubeOut( p, iCube2 );
            for  w = 0; w < p.nWordsOut; w++  {
                pCube2Out[w] &= ~pCube1Out[w];
            }
            for  w = 0; w < p.nWordsOut; w++  {
                if pCube2Out[w] {
                    break; }
            }
            if w < p.nWordsOut { // has output literals
                continue; }
            // remove larger cube
            Vec_IntWriteEntry( vGroup2, c2, -1 );
            Vec_IntPush( p.vFree, iCube2 );
            Count++;
        }
        }
    }
    if Count != 0 {
        Map_ManGroupCompact( vGroup2 ); }
    return Count;
}
func Mop_ManMergeContainAll(p *Mop_Man_t, vGroups *Vec_Wec_t) int32 {
    var vGroup, vGroup2 *Vec_Int_t;
    var i, k, Count int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vGroups)) && (((vGroup) = Vec_WecEntry(vGroups, i)), 1); i++  {
        Count += Mop_ManRemoveIdentical( p, vGroup );
        // inlining of macro Vec_WecForEachLevelStart
        for  k = i+1; (k < Vec_WecSize(vGroups)) && (((vGroup2) = Vec_WecEntry(vGroups, k)), 1); k++  {
            Count += Mop_ManMergeContainTwo( p, vGroup, vGroup2 );
        }
    }
    return Count;
}
func Mop_ManReduce2(p *Mop_Man_t) {
    var clk abctime  = Abc_Clock();
    var nCubes int32   = Vec_IntSize(p.vCubes);
    var vStats *Vec_Int_t  = Mop_ManCollectStats( p );
    var vGroups *Vec_Wec_t  = Mop_ManCreateGroups( p );
    var nLimit int32     = ABC_INFINITY; // 5 * Vec_IntSum(vStats) / Vec_IntSize(vStats) + 1;
    var nOutLits int32   = Mop_ManCountOutputLits( p );
    var Count1 int32     = Mop_ManMergeContainAll( p, vGroups );
    var Count2 int32     = Mop_ManMergeDist1All( p, vGroups, vStats, nLimit );
    var Count3 int32     = Mop_ManMergeContainAll( p, vGroups );
    var Count4 int32     = Mop_ManMergeDist1All( p, vGroups, vStats, nLimit );
    var Count5 int32     = Mop_ManMergeContainAll( p, vGroups );
    var Removed int32    = Mop_ManUnCreateGroups( p, vGroups );
    var nOutLits2 int32  = Mop_ManCountOutputLits( p );
    Vec_WecFree( vGroups );
//Vec_IntPrint( vStats );
    Vec_IntFree( vStats );
    assert(Removed == Count1 + Count2 + Count3, "Removed == Count1 + Count2 + Count3");
    // report
    fmt.Printf("Cubes: %d -> %d.  C = %d.  M = %d.  C = %d.  M = %d.  C = %d.  Output lits: %d -> %d.   ", nCubes, Vec_IntSize(p.vCubes), Count1, Count2, Count3, Count4, Count5, nOutLits, nOutLits2);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mop_ManReduce(p *Mop_Man_t) {
    var clk abctime  = Abc_Clock();
    var vGroup, vGroup2 *Vec_Int_t;
    var i, k, nOutLits, nOutLits2, nEqual, nContain int32;
    var vGroups *Vec_Wec_t  = Mop_ManCreateGroups( p );
    // initial stats
    nOutLits = Mop_ManCountOutputLits( p );
    // check identical cubes within each group
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vGroups)) && (((vGroup) = Vec_WecEntry(vGroups, i)), 1); i++  {
        nEqual += Mop_ManRemoveIdentical( p, vGroup );
    }
    // check contained cubes
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vGroups)) && (((vGroup) = Vec_WecEntry(vGroups, i)), 1); i++  {
        // inlining of macro Vec_WecForEachLevelStart
        for  k = i+1; (k < Vec_WecSize(vGroups)) && (((vGroup2) = Vec_WecEntry(vGroups, k)), 1); k++  {
            nContain += Mop_ManMergeContainTwo( p, vGroup, vGroup2 );
        }
    }
    // final stats
    nOutLits2 = Mop_ManCountOutputLits( p );
    Mop_ManUnCreateGroups( p, vGroups );
    Vec_WecFree( vGroups );
    // report
    fmt.Printf("Total = %d. Reduced %d equal and %d contained cubes. Output lits: %d -> %d.   ", Vec_IntSize(p.vCubes), nEqual, nContain, nOutLits, nOutLits2);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mop_ManCubeCount(p *Mop_Man_t) *Vec_Wec_t {
    var i, k, iCube int32;
    var vOuts *Vec_Wec_t  = Vec_WecStart( p.nOuts );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCubes)) && (((iCube) = Vec_IntEntry(p.vCubes, i)), 1); i++  {
    if iCube != -1 {
        var pCube *word  = Mop_ManCubeOut( p, iCube );
        for  k = 0; k < p.nOuts; k++  {
            if Abc_TtGetBit( pCube, k ) {
                Vec_WecPush( vOuts, k, iCube ); }
        }
    }
    }
    return vOuts;
}
func Mop_ManDerive(p *Mop_Man_t, pFileName string) *Abc_Ntk_t {
    var i, k, c, iCube int32; 
    var Symb [4]byte = [4]byte{ '-', '0', '1', '?' };     // cube symbols
    var vSop *Vec_Str_t   = Vec_StrAlloc( 1000 );  // storage for one SOP
    var vOuts *Vec_Wec_t  = Mop_ManCubeCount( p ); // cube count for each output
    var pNtk *Abc_Ntk_t   = Abc_NtkAlloc( ABC_NTK_LOGIC, ABC_FUNC_SOP, 1 );
    pNtk.pName = Extra_UtilStrsav( pFileName );
    pNtk.pSpec = Extra_UtilStrsav( pFileName );
    for  i = 0; i < p.nIns; i++  {
        Abc_NtkCreatePi(pNtk);
    }
    for  i = 0; i < p.nOuts; i++  {
        var vThis *Vec_Int_t  = Vec_WecEntry( vOuts, i );
        var pPo *Abc_Obj_t    = Abc_NtkCreatePo(pNtk);
        var pNode *Abc_Obj_t  = Abc_NtkCreateNode(pNtk);
        Abc_ObjAddFanin( pPo, pNode );
        if Vec_IntSize(vThis) == 0 {
            pNode.pData = Abc_SopRegister( (Mem_Flex_t *)pNtk.pManFunc, " 0\n" );
            continue;
        }
        for  k = 0; k < p.nIns; k++  {
            Abc_ObjAddFanin( pNode, Abc_NtkPi(pNtk, k) );
        }
        Vec_StrClear( vSop );
        // inlining of macro Vec_IntForEachEntry
        for  c = 0; (c < Vec_IntSize(vThis)) && (((iCube) = Vec_IntEntry(vThis, c)), 1); c++  {
            var pCube *word  = Mop_ManCubeIn( p, iCube );
            for  k = 0; k < p.nIns; k++  {
                Vec_StrPush( vSop, Symb[Abc_TtGetQua(pCube, k)] );
            }
            Vec_StrAppend( vSop, " 1\n" );
        }
        Vec_StrPush( vSop, '\0' );
        pNode.pData = Abc_SopRegister( (Mem_Flex_t *)pNtk.pManFunc, Vec_StrArray(vSop) );
    }
    Vec_StrFree( vSop );
    Vec_WecFree( vOuts );
    Abc_NtkAddDummyPiNames( pNtk );
    Abc_NtkAddDummyPoNames( pNtk );
    return pNtk;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mop_ManTest(pFileName string, fMerge int32, fVerbose int32) *Abc_Ntk_t {
    var pNtk *Abc_Ntk_t  = nil;
    var p *Mop_Man_t  = Mop_ManRead( pFileName );
    if p == nil {
        return nil; }
    Mop_ManRemoveEmpty( p );
    //Mop_ManPrint( p );
    if fMerge != 0 {
        Mop_ManReduce2( p );
    } else {
        Mop_ManReduce( p ); }
    //Mop_ManPrint( p );
    pNtk = Mop_ManDerive( p, pFileName );
    Mop_ManStop( p );
    return pNtk;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_IMPL_END


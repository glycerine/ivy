package if_

import "fmt"

/**CFile****************************************************************

  FileName    [ifDec16.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [Fast checking procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifDec16.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"
//#include "bool/kit/kit.h"
//#include "misc/vec/vecMem.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const CLU_VAR_MAX =  16
#define CLU_WRD_MAX  (1 << ((CLU_VAR_MAX)-6))
const CLU_MEM_MAX =  1000  // 1 GB
const CLU_UNUSED =   0xff

// decomposition
   
type If_Grp_t struct {
    nVars       byte;
    nMyu       byte;
    pVars [CLU_VAR_MAX]byte;
}

// hash table entry
   
type If_Hte_t struct {
    pNext  *If_Hte_t;
    Group   uint32;
    Counter   uint32;
    word       pTruth[1];
}

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
var TruthAll [CLU_VAR_MAX][CLU_WRD_MAX]word = [CLU_VAR_MAX][CLU_WRD_MAX]word{{0}};

// extern void Kit_DsdPrintFromTruth( unsigned * pTruth, int nVars );

// extern void Extra_PrintBinary( FILE * pFile, unsigned Sign[], int nBits );


// extern int If_CluSupportSize( word * t, int nVars );


var s_Count2 int32;
var s_Count3 int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

func If_CluGrp2Uns(pG *If_Grp_t) uint32 {
    var pChar string  = (char *)pG;
    var Res uint32;
    var i int32;
    for  i = 0; i < 8; i++  {
        Res |= ((pChar[i] & 15) << (i << 2));
    }
    return Res;
}

func If_CluUns2Grp(Group uint32, pG *If_Grp_t) {
    var pChar string  = (char *)pG;
    var i int32;
    for  i = 0; i < 8; i++  {
        pChar[i] = ((Group >> (i << 2)) & 15);
    }
}

unsigned int If_CluPrimeCudd( unsigned int p )
{
    var i, pn int32;

    p--;
    for {
        p++;
        if p&1 {
        pn = 1;
        i = 3;
        for uint32( (i * i)) <= p {
        if p % i == 0 {
            pn = 0;
            break;
        }
        i += 2;
        }
    } else {
        pn = 0;
    }
    
if !(!(pn != 0)) { break }
} 
    return(p);

} /* end of Cudd_Prime */

// hash table
func If_CluWordNum(nVars int32) int32 {
    if nVars <= 6 {
        return 1;
    } else {
        return 1 << (nVars-6);
    }
}
func If_CluCountOnes(t word) int32 {
    t =    (t & ABC_CONST(0x5555555555555555)) + ((t>> 1) & ABC_CONST(0x5555555555555555));
    t =    (t & ABC_CONST(0x3333333333333333)) + ((t>> 2) & ABC_CONST(0x3333333333333333));
    t =    (t & ABC_CONST(0x0F0F0F0F0F0F0F0F)) + ((t>> 4) & ABC_CONST(0x0F0F0F0F0F0F0F0F));
    t =    (t & ABC_CONST(0x00FF00FF00FF00FF)) + ((t>> 8) & ABC_CONST(0x00FF00FF00FF00FF));
    t =    (t & ABC_CONST(0x0000FFFF0000FFFF)) + ((t>>16) & ABC_CONST(0x0000FFFF0000FFFF));
    return (t & ABC_CONST(0x00000000FFFFFFFF)) +  (t>>32);
}

func If_CluHashTableCheck(p *If_Man_t) {
    var t int32  = 1;
    var pEntry *If_Hte_t;
    var i, RetValue, Status int32;
    for  i = 0; i < p.nTableSize[t]; i++  {
        for  pEntry = ((If_Hte_t **)p.pHashTable[t])[i]; (pEntry != nil); pEntry = pEntry.pNext  {        
            Status = ((pEntry.Group & 15) > 0);
            RetValue = If_CutPerformCheck16( nil, (unsigned *)pEntry.pTruth, 13, If_CluSupportSize(pEntry.pTruth, 13), "555" );
            if RetValue != Status {
                Kit_DsdPrintFromTruth( (unsigned*)pEntry.pTruth, 13 ); fmt.Printf("\n");
                RetValue = If_CutPerformCheck16( nil, (unsigned *)pEntry.pTruth, 13, If_CluSupportSize(pEntry.pTruth, 13), "555" );
                fmt.Printf("Hash table problem!!!\n");
            }
        }
    }
}
func If_CluHashPrintStats(p *If_Man_t, t int32) {
    var pEntry *If_Hte_t;
    var i, Counter int32;
    for  i = 0; i < p.nTableSize[t]; i++  {
        Counter = 0;
        for  pEntry = ((If_Hte_t **)p.pHashTable[t])[i]; (pEntry != nil); pEntry = pEntry.pNext  {
            Counter++;
        }
        if Counter == 0 {
            continue; }
        if Counter < 10 {
            continue; }
        fmt.Printf("%d=%d ", i, Counter);
    }
}
func If_CluHashFindMedian(p *If_Man_t, t int32) int32 {
    var pEntry *If_Hte_t;
    var vCounters *Vec_Int_t;
    var i, Max, Total, Half int32;
    vCounters = Vec_IntStart( 1000 );
    for  i = 0; i < p.nTableSize[t]; i++  {
        for  pEntry = ((If_Hte_t **)p.pHashTable[t])[i]; (pEntry != nil); pEntry = pEntry.pNext  {
            if Max < int32(pEntry.Counter) {
                Max = pEntry.Counter;
                Vec_IntSetEntry( vCounters, pEntry.Counter, 0 );
            }
            Vec_IntAddToEntry( vCounters, pEntry.Counter, 1 );
            Total++;
        }
    }
    for  i = Max; i > 0; i--  {
        Half += Vec_IntEntry( vCounters, i );
        if Half > Total/2 {
            break; }
    }
/*
    printf( "total = %d  ", Total );
    printf( "half = %d  ", Half );
    printf( "i = %d  ", i );
    printf( "Max = %d.\n", Max );
*/
    Vec_IntFree( vCounters );
    return Abc_MaxInt( i, 1 );
}
var spStatic_If_CluHashKey_BigPrimes [8]uint32 = [8]uint32{12582917, 25165843, 50331653, 100663319, 201326611, 402653189, 805306457, 1610612741};
func If_CluHashKey(pTruth *word, nWords int32, Size int32) int32 {
    
    var Value uint32;
    var i int32;
    if nWords < 4 {
        unsigned char * s = (unsigned char *)pTruth;
        for  i = 0; i < 8 * nWords; i++  {
            Value ^= spStatic_If_CluHashKey_BigPrimes[i % 7] * s[i];
        }
    
    } else {
        var s *uint32  = (unsigned *)pTruth;
        for  i = 0; i < 2 * nWords; i++  {
            Value ^= spStatic_If_CluHashKey_BigPrimes[i % 7] * s[i];
        }
    }
    return Value % Size;
}
func If_CluHashLookup(p *If_Man_t, pTruth *word, t int32) *uint32 {
    var pEntry, pPrev *If_Hte_t;
    var nWords, HashKey int32;
    if p == nil {
        return nil; }
    nWords = If_CluWordNum(p.pPars.nLutSize);
    if p.pMemEntries == nil {
        p.pMemEntries = Mem_FixedStart( sizeof(If_Hte_t) + sizeof(word) * (If_CluWordNum(p.pPars.nLutSize) - 1) ); }
    if p.pHashTable[t] == nil {
        // decide how large should be the table
        var nEntriesMax1 int32  = 4 * If_CluPrimeCudd( Vec_PtrSize(p.vObjs) * p.pPars.nCutsMax );
        var nEntriesMax2 int32  = int32(((float64(CLU_MEM_MAX) * (1 << 20)) / If_CluWordNum(p.pPars.nLutSize) / 8));
//        int nEntriesMax2 = 10000;
        // create table
        p.nTableSize[t] = If_CluPrimeCudd( Abc_MinInt(nEntriesMax1, nEntriesMax2)/2 );
        p.pHashTable[t] = ABC_CALLOC( void *, p.nTableSize[t] );
    }
    // check if this entry exists
    HashKey = If_CluHashKey( pTruth, nWords, p.nTableSize[t] );
    for  pEntry = ((If_Hte_t **)p.pHashTable[t])[HashKey]; (pEntry != nil); pEntry = pEntry.pNext  {
        if memcmp(pEntry.pTruth, pTruth, sizeof(word) * nWords) == 0 {
            pEntry.Counter++;
            return &pEntry.Group;
        }
    }
    // resize the hash table
    if p.nTableEntries[t] >= 2 * p.nTableSize[t] {
        // collect useful entries
        var pPrev *If_Hte_t;
        var vUseful *Vec_Ptr_t  = Vec_PtrAlloc( p.nTableEntries[t] );
        var i int32;
        var Median int32 = If_CluHashFindMedian( p, t );
        for  i = 0; i < p.nTableSize[t]; i++  {
            for  pEntry = ((If_Hte_t **)p.pHashTable[t])[i]; (pEntry != nil);  {
                if int32(pEntry.Counter) > Median {
                    Vec_PtrPush( vUseful, pEntry );
                    pEntry = pEntry.pNext;
                
                } else {
                    pPrev = pEntry.pNext;
                    Mem_FixedEntryRecycle( p.pMemEntries, (char *)pEntry );
                    pEntry = pPrev;
                }
            }
        }
        // add useful entries
        memset( p.pHashTable[t], 0, sizeof(void *) * p.nTableSize[t] );
        Vec_PtrForEachEntry( If_Hte_t *, vUseful, pEntry, i )
        {
            HashKey = If_CluHashKey( pEntry.pTruth, nWords, p.nTableSize[t] );
            pPrev = ((If_Hte_t **)p.pHashTable[t])[HashKey];
            if pPrev == nil || pEntry.Counter >= pPrev.Counter {
                pEntry.pNext = pPrev;
                ((If_Hte_t **)p.pHashTable[t])[HashKey] = pEntry;
            
            } else {
                for  pPrev.pNext && pEntry.Counter < pPrev.pNext.Counter  {
                    pPrev = pPrev.pNext;
                }
                pEntry.pNext = pPrev.pNext;
                pPrev.pNext = pEntry;
            }
        }
        p.nTableEntries[t] = Vec_PtrSize( vUseful );
        Vec_PtrFree( vUseful );
    }
    // create entry
    p.nTableEntries[t]++;
    pEntry = (If_Hte_t *)Mem_FixedEntryFetch( p.pMemEntries );
    memcpy( pEntry.pTruth, pTruth, sizeof(word) * nWords );
    pEntry.Group = CLU_UNUSED;
    pEntry.Counter = 1;
    // insert at the beginning
//    pEntry->pNext = ((If_Hte_t **)p->pHashTable[t])[HashKey];
//    ((If_Hte_t **)p->pHashTable[t])[HashKey] = pEntry;
    // insert at the end
    pEntry.pNext = nil;
    for  pPrev = ((If_Hte_t **)p.pHashTable[t])[HashKey]; (pPrev != nil) && pPrev.pNext; pPrev = pPrev.pNext  {}
    if pPrev == nil {
        ((If_Hte_t **)p.pHashTable[t])[HashKey] = pEntry;
    } else {
        pPrev.pNext = pEntry; }
    return &pEntry.Group;
}

// variable permutation for large functions
func If_CluClear(pIn *word, nVars int32) {
    var w int32;
    var nWords int32 = If_CluWordNum( nVars );
    for  w = 0; w < nWords; w++  {
        pIn[w] = 0;
    }
}
func If_CluFill(pIn *word, nVars int32) {
    var w int32;
    var nWords int32 = If_CluWordNum( nVars );
    for  w = 0; w < nWords; w++  {
        pIn[w] = ~(word)0;
    }
}
func If_CluCopy(pOut *word, pIn *word, nVars int32) {
    var w int32;
    var nWords int32 = If_CluWordNum( nVars );
    for  w = 0; w < nWords; w++  {
        pOut[w] = pIn[w];
    }
}
func If_CluEqual(pOut *word, pIn *word, nVars int32) int32 {
    var w int32;
    var nWords int32 = If_CluWordNum( nVars );
    for  w = 0; w < nWords; w++  {
        if pOut[w] != pIn[w] {
            return 0; }
    }
    return 1;
}
func If_CluAnd(pRes *word, pIn1 *word, pIn2 *word, nVars int32) {
    var w int32;
    var nWords int32 = If_CluWordNum( nVars );
    for  w = 0; w < nWords; w++  {
        pRes[w] = pIn1[w] & pIn2[w];
    }
}
func If_CluSharp(pRes *word, pIn1 *word, pIn2 *word, nVars int32) {
    var w int32;
    var nWords int32 = If_CluWordNum( nVars );
    for  w = 0; w < nWords; w++  {
        pRes[w] = pIn1[w] & ~pIn2[w];
    }
} 
func If_CluOr(pRes *word, pIn1 *word, pIn2 *word, nVars int32) {
    var w int32;
    var nWords int32 = If_CluWordNum( nVars );
    for  w = 0; w < nWords; w++  {
        pRes[w] = pIn1[w] | pIn2[w];
    }
}
func If_CluAdjust(t word, nVars int32) word {
    assert(nVars >= 0 && nVars <= 6, "nVars >= 0 && nVars <= 6");
    if nVars == 6 {
        return t; }
    t &= (((word)1) << (1 << nVars)) - 1;
    if nVars == 0 {
        t |= t << (1<<nVars++); }
    if nVars == 1 {
        t |= t << (1<<nVars++); }
    if nVars == 2 {
        t |= t << (1<<nVars++); }
    if nVars == 3 {
        t |= t << (1<<nVars++); }
    if nVars == 4 {
        t |= t << (1<<nVars++); }
    if nVars == 5 {
        t |= t << (1<<nVars++); }
    return t;
}
func If_CluAdjustBig(pF *word, nVarsCur int32, nVarsMax int32) {
    var v, nWords int32;
    if nVarsCur == nVarsMax {
        return; }
    assert(nVarsCur < nVarsMax, "nVarsCur < nVarsMax");
    for  v = Abc_MaxInt( nVarsCur, 6 ); v < nVarsMax; v++  {
        nWords = If_CluWordNum( v );
        If_CluCopy( pF + nWords, pF, v );
    }
}
func If_CluSwapAdjacent(pOut *word, pIn *word, iVar int32, nVars int32) {
    var i int32;
    var k int32;
    var nWords int32 = If_CluWordNum( nVars );
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

func If_CluChangePhase(pF *word, nVars int32, iVar int32) {
    var nWords int32  = If_CluWordNum( nVars );
    assert(iVar < nVars, "iVar < nVars");
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            pF[i] = ((pF[i] & ~Truth6[iVar]) << Shift) | ((pF[i] & Truth6[iVar]) >> Shift);
        }
    
    } else {
        var Temp word;
        var i int32;
        var k int32;
        var Step int32 = (1 << (iVar - 6));
        for  k = 0; k < nWords; k += 2*Step  {
            for  i = 0; i < Step; i++  {
                Temp = pF[i];
                pF[i] = pF[Step+i];
                pF[Step+i] = Temp;
            }
            pF += 2*Step;
        }
    }
}
func If_CluCountOnesInCofs(pTruth *word, nVars int32, pStore *int32) {
    var nWords int32  = If_CluWordNum( nVars );
    var i int32;
    var k int32;
    var nOnes int32;
    var Limit int32 = Abc_MinInt( nVars, 6 );
    memset( pStore, 0, sizeof(int) * 2 * nVars );
    // compute positive cofactors
    for  k = 0; k < nWords; k++  {
        for  i = 0; i < Limit; i++  {
            pStore[2*i+1] += If_CluCountOnes( pTruth[k] & Truth6[i] );
        }
    }
    if nVars > 6 {
    for  k = 0; k < nWords; k++  {
        for  i = 6; i < nVars; i++  {
            if k & (1 << (i-6)) {
                pStore[2*i+1] += If_CluCountOnes( pTruth[k] ); }
        }
    } }
    // compute negative cofactors
    for  k = 0; k < nWords; k++  {
        nOnes += If_CluCountOnes( pTruth[k] );
    }
    for  i = 0; i < nVars; i++  {
        pStore[2*i] = nOnes - pStore[2*i+1];
    }
}
func If_CluSemiCanonicize(pTruth *word, nVars int32, pCanonPerm *int32) uint32 {
    var pFunc [CLU_WRD_MAX]word;
    var pIn *word = pTruth;
    var pOut *word = pFunc;
    var pTemp *word;
    var pStore [CLU_VAR_MAX*2]int32;
    var uCanonPhase uint32;
    var i, Temp, fChange, Counter int32;
//Kit_DsdPrintFromTruth( (unsigned*)pTruth, nVars ); printf( "\n" );

    // collect signatures 
    If_CluCountOnesInCofs( pTruth, nVars, pStore );
    // canonicize phase
    for  i = 0; i < nVars; i++  {
        if pStore[2*i+0] <= pStore[2*i+1] {
            continue; }
        uCanonPhase |= (1 << i);
        Temp = pStore[2*i+0];
        pStore[2*i+0] = pStore[2*i+1];
        pStore[2*i+1] = Temp;
        If_CluChangePhase( pIn, nVars, i );
    }
    // compute permutation
    for  i = 0; i < nVars; i++  {
        pCanonPerm[i] = i;
    }
    do {
        fChange = 0;
        for  i = 0; i < nVars-1; i++  {
            if pStore[2*i] <= pStore[2*(i+1)] {
                continue; }
            Counter++;
            fChange = 1;

            Temp = pCanonPerm[i];
            pCanonPerm[i] = pCanonPerm[i+1];
            pCanonPerm[i+1] = Temp;

            Temp = pStore[2*i];
            pStore[2*i] = pStore[2*(i+1)];
            pStore[2*(i+1)] = Temp;

            Temp = pStore[2*i+1];
            pStore[2*i+1] = pStore[2*(i+1)+1];
            pStore[2*(i+1)+1] = Temp;

            If_CluSwapAdjacent( pOut, pIn, i, nVars );
            pTemp = pIn; pIn = pOut; pOut = pTemp;
        }
    } while ( fChange );
    // swap if it was moved an odd number of times
    if Counter & 1 {
        If_CluCopy( pOut, pIn, nVars ); }
    return uCanonPhase;
}
func If_CluSemiCanonicizeVerify(pTruth *word, pTruth0 *word, nVars int32, pCanonPerm *int32, uCanonPhase uint32) {
    var pFunc [CLU_WRD_MAX]word;
    var pGunc [CLU_WRD_MAX]word;
    var pIn *word = pTruth;
    var pOut *word = pFunc;
    var pTemp *word;
    var i, Temp, fChange, Counter int32;
    If_CluCopy( pGunc, pTruth, nVars );
    // undo permutation
    do {
        fChange = 0;
        for  i = 0; i < nVars-1; i++  {
            if pCanonPerm[i] < pCanonPerm[i+1] {
                continue; }

            Counter++;
            fChange = 1;

            Temp = pCanonPerm[i];
            pCanonPerm[i] = pCanonPerm[i+1];
            pCanonPerm[i+1] = Temp;

            If_CluSwapAdjacent( pOut, pIn, i, nVars );
            pTemp = pIn; pIn = pOut; pOut = pTemp;
        }
    } while ( fChange );
    if Counter & 1 {
        If_CluCopy( pOut, pIn, nVars ); }
    // undo phase
    for  i = 0; i < nVars; i++  {
        if (uCanonPhase >> i) & 1 {
            If_CluChangePhase( pTruth, nVars, i ); }
    }
    // compare
    if !(If_CluEqual(pTruth0, pTruth, nVars) != 0) {
        Kit_DsdPrintFromTruth( (unsigned*)pTruth0, nVars ); fmt.Printf("\n");
        Kit_DsdPrintFromTruth( (unsigned*)pGunc, nVars ); fmt.Printf("\n");
        Kit_DsdPrintFromTruth( (unsigned*)pTruth, nVars ); fmt.Printf("\n");
        fmt.Printf("SemiCanonical verification FAILED!\n");
    }
}


func If_CluPrintGroup(g *If_Grp_t) {
    var i int32;
    fmt.Printf("Vars = %d   ", g.nVars);
    fmt.Printf("Myu = %d   {", g.nMyu);
    for  i = 0; i < g.nVars; i++  {
        fmt.Printf(" %c", 'a' + g.pVars[i]);
    }
    fmt.Printf(" }\n");
}

func If_CluPrintConfig(nVars int32, g *If_Grp_t, r *If_Grp_t, BStruth word, pFStruth *word) {
    assert(r.nVars == nVars - g.nVars + 1 + (g.nMyu > 2), "r.nVars == nVars - g.nVars + 1 + (g.nMyu > 2)");
    If_CluPrintGroup( g );
    if g.nVars < 6 {
        BStruth = If_CluAdjust( BStruth, g.nVars ); }
    Kit_DsdPrintFromTruth( (unsigned *)&BStruth, g.nVars );
    fmt.Printf("\n");
    If_CluPrintGroup( r );
    if r.nVars < 6 {
        pFStruth[0] = If_CluAdjust( pFStruth[0], r.nVars ); }
    Kit_DsdPrintFromTruth( (unsigned *)pFStruth, r.nVars );
    fmt.Printf("\n");
}


func If_CluInitTruthTables() {
    var i, k int32;
    assert(CLU_VAR_MAX <= 16, "CLU_VAR_MAX <= 16");
    for  i = 0; i < 6; i++  {
        for  k = 0; k < CLU_WRD_MAX; k++  {
            TruthAll[i][k] = Truth6[i];
        }
    }
    for  i = 6; i < CLU_VAR_MAX; i++  {
        for  k = 0; k < CLU_WRD_MAX; k++  {
            TruthAll[i][k] = ((k >> (i-6)) & 1) ? ~(word)0 : 0;
        }
    }

//    Extra_PrintHex( stdout, TruthAll[6], 8 ); printf( "\n" );
//    Extra_PrintHex( stdout, TruthAll[7], 8 ); printf( "\n" );
}


// verification
func If_CluComposeLut(nVars int32, g *If_Grp_t, t *word, f [][CLU_WRD_MAX]word, r *word) {
    var c [CLU_WRD_MAX]word;
    var m, v int32;
    If_CluClear( r, nVars ); 
    for  m = 0; m < (1<<g.nVars); m++  {
        if !((t[m >> 6] >> (m & 63)) & 1) {
            continue; }
        If_CluFill( c, nVars );
        for  v = 0; v < g.nVars; v++  {
            if (m >> v) & 1 {
                If_CluAnd( c, c, f[v], nVars );
            } else {
                If_CluSharp( c, c, f[v], nVars ); }
        }
        If_CluOr( r, r, c, nVars );
    }
}
func If_CluVerify(pF *word, nVars int32, g *If_Grp_t, r *If_Grp_t, BStruth word, pFStruth *word) {
    var pTTFans [6][CLU_WRD_MAX]word;
    var pTTWire [CLU_WRD_MAX]word;
    var pTTRes [CLU_WRD_MAX]word;
    var i int32;
    assert(g.nVars <= 6 && r.nVars <= 6, "g.nVars <= 6 && r.nVars <= 6");

    if TruthAll[0][0] == 0 {
        If_CluInitTruthTables(); }

    for  i = 0; i < g.nVars; i++  {
        If_CluCopy( pTTFans[i], TruthAll[int32(g.pVars[i])], nVars );
    }
    If_CluComposeLut( nVars, g, &BStruth, pTTFans, pTTWire );

    for  i = 0; i < r.nVars; i++  {
        if r.pVars[i] == nVars {
            If_CluCopy( pTTFans[i], pTTWire, nVars );
        } else {
            If_CluCopy( pTTFans[i], TruthAll[int32(r.pVars[i])], nVars ); }
    }
    If_CluComposeLut( nVars, r, pFStruth, pTTFans, pTTRes );

    if !(If_CluEqual(pTTRes, pF, nVars) != 0) {
        fmt.Printf("\n");
        If_CluPrintConfig( nVars, g, r, BStruth, pFStruth );
        Kit_DsdPrintFromTruth( (unsigned*)pTTRes, nVars ); fmt.Printf("\n");
        Kit_DsdPrintFromTruth( (unsigned*)pF, nVars ); fmt.Printf("\n");
//        Extra_PrintHex( stdout, (unsigned *)pF, nVars ); printf( "\n" );
        fmt.Printf("Verification FAILED!\n");
    }
//    else
//        printf( "Verification succeed!\n" );
}
func If_CluVerify3(pF *word, nVars int32, g *If_Grp_t, g2 *If_Grp_t, r *If_Grp_t, BStruth word, BStruth2 word, FStruth word) {
    var pTTFans [6][CLU_WRD_MAX]word;
    var pTTWire [CLU_WRD_MAX]word;
    var pTTWire2 [CLU_WRD_MAX]word;
    var pTTRes [CLU_WRD_MAX]word;
    var i int32;
    assert(g.nVars >= 2 && g2.nVars >= 2 && r.nVars >= 2, "g.nVars >= 2 && g2.nVars >= 2 && r.nVars >= 2");
    assert(g.nVars <= 6 && g2.nVars <= 6 && r.nVars <= 6, "g.nVars <= 6 && g2.nVars <= 6 && r.nVars <= 6");

    if TruthAll[0][0] == 0 {
        If_CluInitTruthTables(); }

    for  i = 0; i < g.nVars; i++  {
        If_CluCopy( pTTFans[i], TruthAll[int32(g.pVars[i])], nVars );
    }
    If_CluComposeLut( nVars, g, &BStruth, pTTFans, pTTWire );

    for  i = 0; i < g2.nVars; i++  {
        If_CluCopy( pTTFans[i], TruthAll[int32(g2.pVars[i])], nVars );
    }
    If_CluComposeLut( nVars, g2, &BStruth2, pTTFans, pTTWire2 );

    for  i = 0; i < r.nVars; i++  {
        if r.pVars[i] == nVars {
            If_CluCopy( pTTFans[i], pTTWire, nVars );
        } else if r.pVars[i] == nVars + 1 {
            If_CluCopy( pTTFans[i], pTTWire2, nVars );
        } else {
            If_CluCopy( pTTFans[i], TruthAll[int32(r.pVars[i])], nVars ); }
    }
    If_CluComposeLut( nVars, r, &FStruth, pTTFans, pTTRes );

    if !(If_CluEqual(pTTRes, pF, nVars) != 0) {
        fmt.Printf("%d\n", nVars);
//        If_CluPrintConfig( nVars, g, r, BStruth, pFStruth );
//        Extra_PrintHex( stdout, (unsigned *)pF, nVars ); printf( "\n" );

        Kit_DsdPrintFromTruth( (unsigned*)&BStruth, g.nVars );   fmt.Printf("    "); If_CluPrintGroup(g);  // printf( "\n" );
        Kit_DsdPrintFromTruth( (unsigned*)&BStruth2, g2.nVars ); fmt.Printf("    "); If_CluPrintGroup(g2); // printf( "\n" );
        Kit_DsdPrintFromTruth( (unsigned*)&FStruth, r.nVars );   fmt.Printf("    "); If_CluPrintGroup(r);  // printf( "\n" );

        Kit_DsdPrintFromTruth( (unsigned*)pTTWire, nVars ); fmt.Printf("\n");
        Kit_DsdPrintFromTruth( (unsigned*)pTTWire2, nVars ); fmt.Printf("\n");
        Kit_DsdPrintFromTruth( (unsigned*)pTTRes, nVars ); fmt.Printf("\n");
        Kit_DsdPrintFromTruth( (unsigned*)pF, nVars ); fmt.Printf("\n");
//        Extra_PrintHex( stdout, (unsigned *)pF, nVars ); printf( "\n" );
        fmt.Printf("Verification FAILED!\n");
    }
//    else
//        printf( "Verification succeed!\n" );
}



func If_CluSwapVars(pTruth *word, nVars int32, V2P *int32, P2V *int32, iVar int32, jVar int32) {
    var low2High, high2Low, temp word;
    var nWords int32  = If_CluWordNum(nVars);
    var shift, step, iStep, jStep int32;
    var w, i, j int32;
    static word PPMasks[6][6] = {
        { ABC_CONST(0x2222222222222222), ABC_CONST(0x0A0A0A0A0A0A0A0A), ABC_CONST(0x00AA00AA00AA00AA), ABC_CONST(0x0000AAAA0000AAAA), ABC_CONST(0x00000000AAAAAAAA), ABC_CONST(0xAAAAAAAAAAAAAAAA) },
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0C0C0C0C0C0C0C0C), ABC_CONST(0x00CC00CC00CC00CC), ABC_CONST(0x0000CCCC0000CCCC), ABC_CONST(0x00000000CCCCCCCC), ABC_CONST(0xCCCCCCCCCCCCCCCC) },
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x00F000F000F000F0), ABC_CONST(0x0000F0F00000F0F0), ABC_CONST(0x00000000F0F0F0F0), ABC_CONST(0xF0F0F0F0F0F0F0F0) },
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000FF000000FF00), ABC_CONST(0x00000000FF00FF00), ABC_CONST(0xFF00FF00FF00FF00) },
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x00000000FFFF0000), ABC_CONST(0xFFFF0000FFFF0000) },
        { ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0xFFFFFFFF00000000) }
    };
    if iVar == jVar {
        return; }
    if jVar < iVar {
        var varTemp int32  = jVar;
        jVar = iVar;
        iVar = varTemp;
    }
    if iVar <= 5 && jVar <= 5 {
        shift = (1 <<  jVar) - (1 << iVar);
        for  w = 0; w < nWords; w++  {
            low2High = (pTruth[w] & PPMasks[iVar][jVar - 1] ) << shift;
            pTruth[w] &= ~PPMasks[iVar][jVar - 1];
            high2Low = (pTruth[w] & (PPMasks[iVar][jVar - 1] << shift )) >> shift;
            pTruth[w] &= ~ (PPMasks[iVar][jVar - 1] << shift);
            pTruth[w] = pTruth[w] | low2High | high2Low;
        }
    
    } else if iVar <= 5 && jVar > 5 {
        step = If_CluWordNum(jVar + 1)/2;
        shift = 1 << iVar;
        for  w = 0; w < nWords; w += 2*step  {
            for j = 0; j < step; j++ {
                low2High = (pTruth[w + j] & PPMasks[iVar][5]) >> shift;
                pTruth[w + j] &= ~PPMasks[iVar][5];
                high2Low = (pTruth[w + step + j] & (PPMasks[iVar][5] >> shift)) << shift;
                pTruth[w + step + j] &= ~(PPMasks[iVar][5] >> shift);
                pTruth[w + j] |= high2Low;
                pTruth[w + step + j] |= low2High;            
            }
        }
    
    } else {
        iStep = If_CluWordNum(iVar + 1)/2;
        jStep = If_CluWordNum(jVar + 1)/2;
        for w = 0; w < nWords; w += 2*jStep {
            for i = 0; i < jStep; i += 2*iStep {
                for j = 0; j < iStep; j++ {
                    temp = pTruth[w + iStep + i + j];
                    pTruth[w + iStep + i + j] = pTruth[w + jStep + i + j];
                    pTruth[w + jStep + i + j] = temp;
                }
            }
        }
    }    
    if (V2P != nil) && (P2V != nil) {
        V2P[P2V[iVar]] = jVar;
        V2P[P2V[jVar]] = iVar;
        P2V[iVar] ^= P2V[jVar];
        P2V[jVar] ^= P2V[iVar];
        P2V[iVar] ^= P2V[jVar];
    }
}
func If_CluReverseOrder(pTruth *word, nVars int32, V2P *int32, P2V *int32, iVarStart int32) {
    var i, j, k int32;
    for  k = 0; k < (nVars-iVarStart)/2 ; k++  {
        i = iVarStart + k;
        j = nVars - 1 - k;
        If_CluSwapVars( pTruth, nVars, V2P, P2V, i, j );
    }
}

// moves one var (v) to the given position (p)
func If_CluMoveVar2(pF *word, nVars int32, Var2Pla *int32, Pla2Var *int32, v int32, p int32) {
    If_CluSwapVars( pF, nVars, Var2Pla, Pla2Var, Var2Pla[v], p );
}

// moves one var (v) to the given position (p)
func If_CluMoveVar(pF *word, nVars int32, Var2Pla *int32, Pla2Var *int32, v int32, p int32) {
    var pG [CLU_WRD_MAX]word;
    var pIn *word = pF;
    var pOut *word = pG;
    var pTemp *word;
    var iPlace0, iPlace1, Count int32;
    assert(v >= 0 && v < nVars, "v >= 0 && v < nVars");
    for  Var2Pla[v] < p  {
        iPlace0 = Var2Pla[v];
        iPlace1 = Var2Pla[v]+1;
        If_CluSwapAdjacent( pOut, pIn, iPlace0, nVars );
        pTemp = pIn; pIn = pOut;
pOut = pTemp;
        Var2Pla[Pla2Var[iPlace0]]++;
        Var2Pla[Pla2Var[iPlace1]]--;
        Pla2Var[iPlace0] ^= Pla2Var[iPlace1];
        Pla2Var[iPlace1] ^= Pla2Var[iPlace0];
        Pla2Var[iPlace0] ^= Pla2Var[iPlace1];
        Count++;
    }
    for  Var2Pla[v] > p  {
        iPlace0 = Var2Pla[v]-1;
        iPlace1 = Var2Pla[v];
        If_CluSwapAdjacent( pOut, pIn, iPlace0, nVars );
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
        If_CluCopy( pF, pIn, nVars ); }
    assert(Pla2Var[p] == v, "Pla2Var[p] == v");
}

// moves vars to be the most signiticant ones (Group[0] is MSB)
func If_CluMoveGroupToMsb(pF *word, nVars int32, V2P *int32, P2V *int32, g *If_Grp_t) {
    var v int32;
    for  v = 0; v < g.nVars; v++  {
        If_CluMoveVar( pF, nVars, V2P, P2V, g.pVars[g.nVars - 1 - v], nVars - 1 - v );
    }
}


// reverses the variable order
func If_CluReverseOrder_old(pF *word, nVars int32, V2P *int32, P2V *int32, iVarStart int32) {
    var pG [CLU_WRD_MAX]word;
    var v int32;

    If_CluCopy( pG, pF, nVars );

//    for ( v = 0; v < nVars; v++ )
//        printf( "%c ", 'a' + P2V[v] );
//    printf( "  ---  " );

    for  v = iVarStart; v < nVars; v++  {
        If_CluMoveVar( pF, nVars, V2P, P2V, P2V[iVarStart], nVars - 1 - (v - iVarStart) );
    }

//    for ( v = 0; v < nVars; v++ )
//        printf( "%c ", 'a' + P2V[v] );
//    printf( "\n" );

//    if ( iVarStart > 0 )
//        return;

    If_CluReverseOrder( pG, nVars, nil, nil, iVarStart );
    if If_CluEqual( pG, pF, nVars ) != 0 {
//        printf( "+" );
    
    } else {
/*
        printf( "\n" );
        Kit_DsdPrintFromTruth( (unsigned*)pF, nVars ); printf( "\n" );
        Kit_DsdPrintFromTruth( (unsigned*)pG, nVars ); 
        printf( "\n" );
*/
        fmt.Printf("%d ", nVars);
    }
}

// return the number of cofactors w.r.t. the topmost vars (nBSsize)
func If_CluCountCofs(pF *word, nVars int32, nBSsize int32, iShift int32, pCofs [][CLU_WRD_MAX/4]word) int32 {
    var iCofs [128]word ;
    var iCof word;
    var Result word;
    var pCofA, pCofB *word;
    var nMints int32  = (1 << nBSsize);
    var i, c, w, nCofs int32;
    assert(nBSsize >= 2 && nBSsize <= 7 && nBSsize < nVars, "nBSsize >= 2 && nBSsize <= 7 && nBSsize < nVars");
    if nVars - nBSsize < 6 {
        var nShift int32  = (1 << (nVars - nBSsize));
        var Mask word   = ((((word)1) << nShift) - 1);
        for  nCofs = i = 0; i < nMints; i++  {
            iCof = (pF[(iShift + i * nShift) / 64] >> ((iShift + i * nShift) & 63)) & Mask;
            for  c = 0; c < nCofs; c++  {
                if iCof == iCofs[c] {
                    break; }
            }
            if c == nCofs {
                iCofs[nCofs++] = iCof; }
            if pCofs && iCof != iCofs[0] {
                Result |= (((word)1) << i); }
            if nCofs == 5 {
                break; }
        }
        if nCofs <= 2 && pCofs {
            assert(nBSsize <= 6, "nBSsize <= 6");
            pCofs[0][0] = iCofs[0];
            pCofs[1][0] = (nCofs == 2) ? iCofs[1] : iCofs[0];
            pCofs[2][0] = Result;
        }
    
    } else {
        var nWords int32  = If_CluWordNum( nVars - nBSsize );
        assert(nWords * nMints == If_CluWordNum(nVars), "nWords * nMints == If_CluWordNum(nVars)");
        for  nCofs = i = 0; i < nMints; i++  {
            pCofA = pF + i * nWords;
            for  c = 0; c < nCofs; c++  {
                pCofB = pF + iCofs[c] * nWords;
                for  w = 0; w < nWords; w++  {
                    if pCofA[w] != pCofB[w] {
                        break; }
                }
                if w == nWords {
                    break; }
            }
            if c == nCofs {
                iCofs[nCofs++] = i; }
            if pCofs {
                assert(nBSsize <= 6, "nBSsize <= 6");
                pCofB = pF + iCofs[0] * nWords;
                for  w = 0; w < nWords; w++  {
                    if pCofA[w] != pCofB[w] {
                        break; }
                }
                if w != nWords {
                    Result |= (((word)1) << i); }
            }
            if nCofs == 5 {
                break; }
        }
        if nCofs <= 2 && pCofs {
            If_CluCopy( pCofs[0], pF + iCofs[0] * nWords, nVars - nBSsize );
            If_CluCopy( pCofs[1], pF + ((nCofs == 2) ? iCofs[1] : iCofs[0]) * nWords, nVars - nBSsize );
            pCofs[2][0] = Result;
        }
    }
    assert(nCofs >= 1 && nCofs <= 5, "nCofs >= 1 && nCofs <= 5");
    return nCofs;
}

// return the number of cofactors w.r.t. the topmost vars (nBSsize)
func If_CluCountCofs4(pF *word, nVars int32, nBSsize int32, pCofs [][CLU_WRD_MAX/4]word) int32 {
    var iCofs [128]word ;
    var iCof word;
    var Result0 word;
    var Result1 word;
    var nMints int32  = (1 << nBSsize);
    var i, c, nCofs int32;
    assert(pCofs, "pCofs");
    assert(nBSsize >= 2 && nBSsize <= 6 && nBSsize < nVars, "nBSsize >= 2 && nBSsize <= 6 && nBSsize < nVars");
    if nVars - nBSsize < 6 {
        var nShift int32  = (1 << (nVars - nBSsize));
        var Mask word   = ((((word)1) << nShift) - 1);
        for  nCofs = i = 0; i < nMints; i++  {
            iCof = (pF[(i * nShift) / 64] >> ((i * nShift) & 63)) & Mask;
            for  c = 0; c < nCofs; c++  {
                if iCof == iCofs[c] {
                    break; }
            }
            if c == nCofs {
                iCofs[nCofs++] = iCof; }
            if c == 1 || c == 3 {
                Result0 |= (((word)1) << i); }
            if c == 2 || c == 3 {
                Result1 |= (((word)1) << i); }
        }
        assert(nCofs >= 3 && nCofs <= 4, "nCofs >= 3 && nCofs <= 4");
        pCofs[0][0] = iCofs[0];
        pCofs[1][0] = iCofs[1];
        pCofs[2][0] = iCofs[2];
        pCofs[3][0] = (nCofs == 4) ? iCofs[3] : iCofs[2];
        pCofs[4][0] = Result0;
        pCofs[5][0] = Result1;
    
    } else {
        assert(false, "0"); }
    return nCofs;
}

func If_CluCofactors(pF *word, nVars int32, iVar int32, pCof0 *word, pCof1 *word) {
    var nWords int32  = If_CluWordNum( nVars );
    assert(iVar < nVars, "iVar < nVars");
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            pCof0[i] = (pF[i] & ~Truth6[iVar]) | ((pF[i] & ~Truth6[iVar]) << Shift);
            pCof1[i] = (pF[i] &  Truth6[iVar]) | ((pF[i] &  Truth6[iVar]) >> Shift);
        }
    
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
    }
}

// returns 1 if we have special case of cofactors; otherwise, returns 0
func If_CluDetectSpecialCaseCofs(pF *word, nVars int32, iVar int32) int32 {
    var Cof0, Cof1 word;
    var State [6]int32 ;
    var i int32;
    var nWords int32 = If_CluWordNum( nVars );
    assert(iVar < nVars, "iVar < nVars");
    if iVar < 6 {
        var Shift int32  = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            Cof0 =  (pF[i] & ~Truth6[iVar]);
            Cof1 = ((pF[i] &  Truth6[iVar]) >> Shift);

            if Cof0 == 0 {
                State[0]++;
            } else if Cof0 == ~Truth6[iVar] {
                State[1]++;
            } else if Cof1 == 0 {
                State[2]++;
            } else if Cof1 == ~Truth6[iVar] {
                State[3]++;
            } else if Cof0 == ~Cof1 {
                State[4]++;
            } else if Cof0 == Cof1 {
                State[5]++; }
        }
    
    } else {
        var k int32;
        var Step int32 = (1 << (iVar - 6));
        for  k = 0; k < nWords; k += 2*Step  {
            for  i = 0; i < Step; i++  {
                Cof0 = pF[i];
                Cof1 = pF[Step+i];

                if Cof0 == 0 {
                    State[0]++;
                } else if Cof0 == ~(word)0 {
                    State[1]++;
                } else if Cof1 == 0 {
                    State[2]++;
                } else if Cof1 == ~(word)0 {
                    State[3]++;
                } else if Cof0 == ~Cof1 {
                    State[4]++;
                } else if Cof0 == Cof1 {
                    State[5]++; }
            }
            pF    += 2*Step;
        }
        nWords /= 2;
    }
    assert(State[5] != nWords, "State[5] != nWords");
    for  i = 0; i < 5; i++  {
        assert(State[i] <= nWords, "State[i] <= nWords");
        if State[i] == nWords {
            return i; }
    }
    return -1;
}

// returns 1 if we have special case of cofactors; otherwise, returns 0
func If_CluDecUsingCofs(pTruth *word, nVars int32, nLutLeaf int32) If_Grp_t {
    var G If_Grp_t  ;
    var pF2 [CLU_WRD_MAX]word;
    var pF *word = pF2;
    var Var2Pla, Pla2Var [CLU_VAR_MAX+2]int32;
    var V2P, P2V [CLU_VAR_MAX+2]int32;
    var nVarsNeeded int32  = nVars - nLutLeaf;
    var v, i, k, iVar, State int32;
//Kit_DsdPrintFromTruth( (unsigned*)pTruth, nVars ); printf( "\n" );
    // create local copy
    If_CluCopy( pF, pTruth, nVars );
    for  k = 0; k < nVars; k++  {
        Var2Pla[k] = Pla2Var[k] = k;
    }
    // find decomposable vars 
    for  i = 0; i < nVarsNeeded; i++  {
        for  v = nVars - 1; v >= 0; v--  {
            State = If_CluDetectSpecialCaseCofs( pF, nVars, v );
            if State == -1 {
                continue; }
            // update the variable place
            iVar = Pla2Var[v];
            for  Var2Pla[iVar] < nVars - 1  {
                var iPlace0 int32  = Var2Pla[iVar];
                var iPlace1 int32  = Var2Pla[iVar]+1;
                Var2Pla[Pla2Var[iPlace0]]++;
                Var2Pla[Pla2Var[iPlace1]]--;
                Pla2Var[iPlace0] ^= Pla2Var[iPlace1];
                Pla2Var[iPlace1] ^= Pla2Var[iPlace0];
                Pla2Var[iPlace0] ^= Pla2Var[iPlace1];
            }
            // move this variable to the top
            for  k = 0; k < nVars; k++  {
                V2P[k] = P2V[k] = k;
            }
//Kit_DsdPrintFromTruth( (unsigned*)pF, nVars ); printf( "\n" );
            If_CluMoveVar( pF, nVars, V2P, P2V, v, nVars - 1 );
//Kit_DsdPrintFromTruth( (unsigned*)pF, nVars ); printf( "\n" );
            // choose cofactor to follow
            iVar = nVars - 1;
            if State == 0 || State == 1 { // need cof1
            
                if iVar < 6 {
                    pF[0] = (pF[0] &  Truth6[iVar]) | ((pF[0] &  Truth6[iVar]) >> (1 << iVar));
                } else {
                    pF += If_CluWordNum( nVars ) / 2; }
            
            } else { // need cof0
            
                if iVar < 6 {
                    pF[0] = (pF[0] & ~Truth6[iVar]) | ((pF[0] & ~Truth6[iVar]) << (1 << iVar)); }
            }
            // update the variable count
            nVars--;
            break;
        }
        if v == -1 {
            return G; }
    }
    // create the resulting group
    G.nVars = nLutLeaf;
    G.nMyu = 2;
    for  v = 0; v < G.nVars; v++  {
        G.pVars[v] = Pla2Var[v];
    }
    return G;
}



// deriving decomposition
func If_CluDeriveDisjoint(pF *word, nVars int32, V2P *int32, P2V *int32, g *If_Grp_t, r *If_Grp_t) word {
    var pCofs [3][CLU_WRD_MAX/4]word;
    var i int32;
    var RetValue int32;
    var nFSset int32 = nVars - g.nVars;
    RetValue = If_CluCountCofs( pF, nVars, g.nVars, 0, pCofs );
//    assert( RetValue == 2 );

    if nFSset < 6 {
        pF[0] = (pCofs[1][0] << (1 << nFSset)) | pCofs[0][0];
    } else {
        If_CluCopy( pF, pCofs[0], nFSset );
        If_CluCopy( pF + If_CluWordNum(nFSset), pCofs[1], nFSset );
    }
    // create the resulting group
    if r {
        r.nVars = nFSset + 1;
        r.nMyu = 0;
        for  i = 0; i < nFSset; i++  {
            r.pVars[i] = P2V[i];
        }
        r.pVars[nFSset] = nVars;
    }
    return pCofs[2][0];
}
func If_CluDeriveDisjoint4(pF *word, nVars int32, V2P *int32, P2V *int32, g *If_Grp_t, r *If_Grp_t, pTruth0 *word, pTruth1 *word) {
    var pCofs [6][CLU_WRD_MAX/4]word;
    var Cof0, Cof1 word;
    var i int32;
    var RetValue int32;
    var nFSset int32 = nVars - g.nVars;

    assert(g.nVars <= 6 && nFSset <= 4, "g.nVars <= 6 && nFSset <= 4");

    RetValue = If_CluCountCofs4( pF, nVars, g.nVars, pCofs );
    if RetValue != 3 && RetValue != 4 {
        fmt.Printf("If_CluDeriveDisjoint4(): Error!!!\n"); }

    Cof0  = (pCofs[1][0] << (1 << nFSset)) | pCofs[0][0];
    Cof1  = (pCofs[3][0] << (1 << nFSset)) | pCofs[2][0];
    pF[0] = (Cof1 << (1 << (nFSset+1))) | Cof0;
    pF[0] = If_CluAdjust( pF[0], nFSset + 2 );

    // create the resulting group
    r.nVars = nFSset + 2;
    r.nMyu = 0;
    for  i = 0; i < nFSset; i++  {
        r.pVars[i] = P2V[i];
    }
    r.pVars[nFSset] = nVars;
    r.pVars[nFSset+1] = nVars+1;

    *pTruth0 = If_CluAdjust( pCofs[4][0], g.nVars );
    *pTruth1 = If_CluAdjust( pCofs[5][0], g.nVars );
}

func If_CluDeriveNonDisjoint(pF *word, nVars int32, V2P *int32, P2V *int32, g *If_Grp_t, r *If_Grp_t) word {
    var pCofs [2][CLU_WRD_MAX]word;
    var Truth0, Truth1, Truth word;
    var i int32;
    var nFSset int32 = nVars - g.nVars;
    var nFSset1 int32 = nFSset + 1;
    If_CluCofactors( pF, nVars, nVars - 1, pCofs[0], pCofs[1] );

//    Extra_PrintHex( stdout, (unsigned *)pCofs[0], nVars ); printf( "\n" );
//    Extra_PrintHex( stdout, (unsigned *)pCofs[1], nVars ); printf( "\n" );

    g.nVars--;
    Truth0 = If_CluDeriveDisjoint( pCofs[0], nVars - 1, V2P, P2V, g, nil );
    Truth1 = If_CluDeriveDisjoint( pCofs[1], nVars - 1, V2P, P2V, g, nil );
    Truth  = (Truth1 << (1 << g.nVars)) | Truth0;
    g.nVars++;
    if nFSset1 < 6 {
        pF[0] = (pCofs[1][0] << (1 << nFSset1)) | pCofs[0][0];
    } else {
        If_CluCopy( pF, pCofs[0], nFSset1 );
        If_CluCopy( pF + If_CluWordNum(nFSset1), pCofs[1], nFSset1 );
    }

//    Extra_PrintHex( stdout, (unsigned *)&Truth0, 6 ); printf( "\n" );
//    Extra_PrintHex( stdout, (unsigned *)&Truth1, 6 ); printf( "\n" );
//    Extra_PrintHex( stdout, (unsigned *)&pCofs[0][0], 6 ); printf( "\n" );
//    Extra_PrintHex( stdout, (unsigned *)&pCofs[1][0], 6 ); printf( "\n" );
//    Extra_PrintHex( stdout, (unsigned *)&Truth, 6 ); printf( "\n" );
//    Extra_PrintHex( stdout, (unsigned *)&pF[0], 6 ); printf( "\n" );

    // create the resulting group
    r.nVars = nFSset + 2;
    r.nMyu = 0;
    for  i = 0; i < nFSset; i++  {
        r.pVars[i] = P2V[i];
    }
    r.pVars[nFSset] = nVars;
    r.pVars[nFSset+1] = g.pVars[g.nVars - 1];
    return Truth;
}

// check non-disjoint decomposition
func If_CluCheckNonDisjointGroup(pF *word, nVars int32, V2P *int32, P2V *int32, g *If_Grp_t) int32 {
    var v, i, nCofsBest2 int32;
    if g.nMyu == 3 || g.nMyu == 4 {
        var pCofs [2][CLU_WRD_MAX]word;
        // try cofactoring w.r.t. each variable
        for  v = 0; v < g.nVars; v++  {
            If_CluCofactors( pF, nVars, V2P[int32(g.pVars[v])], pCofs[0], pCofs[1] );
            nCofsBest2 = If_CluCountCofs( pCofs[0], nVars, g.nVars, 0, nil );
            if nCofsBest2 > 2 {
                continue; }
            nCofsBest2 = If_CluCountCofs( pCofs[1], nVars, g.nVars, 0, nil );
            if nCofsBest2 > 2 {
                continue; }
            // found good shared variable - move to the end
            If_CluMoveVar( pF, nVars, V2P, P2V, g.pVars[v], nVars-1 );
            for  i = 0; i < g.nVars; i++  {
                g.pVars[i] = P2V[nVars-g.nVars+i];
            }
            return 1;
        }
    }
    return 0;
}


// finds a good var group (cof count < 6; vars are MSBs)
func If_CluFindGroup(pF *word, nVars int32, iVarStart int32, iVarStop int32, V2P *int32, P2V *int32, nBSsize int32, fDisjoint int32) If_Grp_t {
    var fVerbose int32;
    var nRounds int32  = 2;//nBSsize;
    var G If_Grp_t ;
    var g *If_Grp_t = &G;//, BestG = {0};
    var i, r, v, nCofs, VarBest, nCofsBest2 int32;
    assert(nVars > nBSsize && nVars >= nBSsize + iVarStart && nVars <= CLU_VAR_MAX, "nVars > nBSsize && nVars >= nBSsize + iVarStart && nVars <= CLU_VAR_MAX");
    assert(nBSsize >= 2 && nBSsize <= 6, "nBSsize >= 2 && nBSsize <= 6");
    assert(!(iVarStart != 0) || !(iVarStop != 0), "!iVarStart || !iVarStop");
    // start with the default group
    g.nVars = nBSsize;
    g.nMyu = If_CluCountCofs( pF, nVars, nBSsize, 0, nil );
    for  i = 0; i < nBSsize; i++  {
        g.pVars[i] = P2V[nVars-nBSsize+i];
    }
    // check if good enough
    if g.nMyu == 2 {
        return G; }
    if !(fDisjoint != 0) && (If_CluCheckNonDisjointGroup( pF, nVars, V2P, P2V, g ) != 0) {
//        BestG = G;
        return G;
    }
    if nVars == nBSsize + iVarStart {
        g.nVars = 0;
        return G;
    }

    if fVerbose != 0 {
        fmt.Printf("Iter %2d  ", -1);
        If_CluPrintGroup( g );
    }

    // try to find better group
    for  r = 0; r < nRounds; r++  {
        if nBSsize < nVars-1 {
            // find the best var to add
            VarBest = P2V[nVars-1-nBSsize];
            nCofsBest2 = If_CluCountCofs( pF, nVars, nBSsize+1, 0, nil );
            for  v = nVars-2-nBSsize; v >= iVarStart; v--  {
//                If_CluMoveVar( pF, nVars, V2P, P2V, P2V[v], nVars-1-nBSsize );
                If_CluMoveVar2( pF, nVars, V2P, P2V, P2V[v], nVars-1-nBSsize );
                nCofs = If_CluCountCofs( pF, nVars, nBSsize+1, 0, nil );
                if nCofsBest2 >= nCofs {
                    nCofsBest2 = nCofs;
                    VarBest = P2V[nVars-1-nBSsize];
                }
            }
            // go back
//            If_CluMoveVar( pF, nVars, V2P, P2V, VarBest, nVars-1-nBSsize );
            If_CluMoveVar2( pF, nVars, V2P, P2V, VarBest, nVars-1-nBSsize );
            // update best bound set
            nCofs = If_CluCountCofs( pF, nVars, nBSsize+1, 0, nil );
            assert(nCofs == nCofsBest2, "nCofs == nCofsBest2");
        }

        // find the best var to remove
        VarBest = P2V[nVars-1-nBSsize];
        nCofsBest2 = If_CluCountCofs( pF, nVars, nBSsize, 0, nil );
        for  v = nVars-nBSsize; v < nVars-iVarStop; v++  {
//            If_CluMoveVar( pF, nVars, V2P, P2V, P2V[v], nVars-1-nBSsize );
            If_CluMoveVar2( pF, nVars, V2P, P2V, P2V[v], nVars-1-nBSsize );
            nCofs = If_CluCountCofs( pF, nVars, nBSsize, 0, nil );
            if nCofsBest2 >= nCofs {
                nCofsBest2 = nCofs;
                VarBest = P2V[nVars-1-nBSsize];
            }
        }

        // go back
//        If_CluMoveVar( pF, nVars, V2P, P2V, VarBest, nVars-1-nBSsize );
        If_CluMoveVar2( pF, nVars, V2P, P2V, VarBest, nVars-1-nBSsize );
        // update best bound set
        nCofs = If_CluCountCofs( pF, nVars, nBSsize, 0, nil );
        assert(nCofs == nCofsBest2, "nCofs == nCofsBest2");
        if g.nMyu >= nCofs {
            g.nVars = nBSsize;
            g.nMyu = nCofs;
            for  i = 0; i < nBSsize; i++  {
                g.pVars[i] = P2V[nVars-nBSsize+i];
            }
        }

        if fVerbose != 0 {
            fmt.Printf("Iter %2d  ", r);
            If_CluPrintGroup( g );
        }

        // check if good enough
        if g.nMyu == 2 {
            return G; }
        if !(fDisjoint != 0) && (If_CluCheckNonDisjointGroup( pF, nVars, V2P, P2V, g ) != 0) {
//            BestG = G;
            return G;
        }
    }

    assert(r == nRounds, "r == nRounds");
    g.nVars = 0;
    return G;
//    return BestG;
}


// double check that the given group has a decomposition
func If_CluCheckGroup(pTruth *word, nVars int32, g *If_Grp_t) {
    var pF [CLU_WRD_MAX]word;
    var v int32;
    var nCofs int32;
    var V2P [CLU_VAR_MAX]int32;
    var P2V [CLU_VAR_MAX]int32;
    assert(g.nVars >= 2 && g.nVars <= 6, "g.nVars >= 2 && g.nVars <= 6"); // vars
    assert(g.nMyu >= 2 && g.nMyu <= 4, "g.nMyu >= 2 && g.nMyu <= 4"); // cofs
    // create permutation
    for  v = 0; v < nVars; v++  {
        V2P[v] = P2V[v] = v;
    }
    // create truth table
    If_CluCopy( pF, pTruth, nVars );
    // move group up
    If_CluMoveGroupToMsb( pF, nVars, V2P, P2V, g );
    // check the number of cofactors
    nCofs = If_CluCountCofs( pF, nVars, g.nVars, 0, nil );
    if nCofs != g.nMyu {
        fmt.Printf("Group check 0 has failed.\n"); }
    // check compatible
    if nCofs > 2 {
        nCofs = If_CluCountCofs( pF, nVars-1, g.nVars-1, 0, nil );
        if nCofs > 2 {
            fmt.Printf("Group check 1 has failed.\n"); }
        nCofs = If_CluCountCofs( pF, nVars-1, g.nVars-1, (1 << (nVars-1)), nil );
        if nCofs > 2 {
            fmt.Printf("Group check 2 has failed.\n"); }
    }
}


// double check that the permutation derived is correct
func If_CluCheckPerm(pTruth *word, pF *word, nVars int32, V2P *int32, P2V *int32) {
    var i int32;
    for  i = 0; i < nVars; i++  {
        If_CluMoveVar( pF, nVars, V2P, P2V, i, i );
    }

    if !(If_CluEqual( pTruth, pF, nVars ) != 0) {
        fmt.Printf("Permutation FAILED.\n"); }
//    else
//        printf( "Permutation successful\n" );
}




func If_CluSuppIsMinBase(Supp int32) int32 {
    return (Supp & (Supp+1)) == 0;
}
func If_CluHasVar(t *word, nVars int32, iVar int32) int32 {
    var nWords int32  = If_CluWordNum( nVars );
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
func If_CluSupport(t *word, nVars int32) int32 {
    var v, Supp int32;
    for  v = 0; v < nVars; v++  {
        if If_CluHasVar( t, nVars, v ) != 0 {
            Supp |= (1 << v); }
    }
    return Supp;
}
func If_CluSupportSize(t *word, nVars int32) int32 {
    var v, SuppSize int32;
    for  v = 0; v < nVars; v++  {
        if If_CluHasVar( t, nVars, v ) != 0 {
            SuppSize++; }
    }
    return SuppSize;
}
func If_CluTruthShrink(pF *word, nVars int32, nVarsAll int32, Phase uint32) {
    var pG [CLU_WRD_MAX]word;
    var pIn *word = pF;
    var pOut *word = pG;
    var pTemp *word;
    var i, k, Var, Counter int32;
    assert(nVarsAll <= 16, "nVarsAll <= 16");
    for  i = 0; i < nVarsAll; i++  {
        if Phase & (1 << i) {
            for  k = i-1; k >= Var; k--  {
                If_CluSwapAdjacent( pOut, pIn, k, nVarsAll );
                pTemp = pIn; pIn = pOut;
pOut = pTemp;
                Counter++;
            }
            Var++;
        }
    }
    assert(Var == nVars, "Var == nVars");
    // swap if it was moved an odd number of times
    if Counter & 1 {
        If_CluCopy( pOut, pIn, nVarsAll ); }
}
func If_CluMinimumBase(t *word, pSupp *int32, nVarsAll int32, pnVars *int32) int32 {
    var v, iVar, uSupp int32;
    assert(nVarsAll <= 16, "nVarsAll <= 16");
    for  v = 0; v < nVarsAll; v++  {
        if If_CluHasVar( t, nVarsAll, v ) != 0 {
            uSupp |= (1 << v);
            if pSupp != nil {
                pSupp[iVar] = pSupp[v]; }
            iVar++;
        }
    }
    if pnVars != nil {
        *pnVars = iVar; }
    if If_CluSuppIsMinBase( uSupp ) != 0 {
        return 0; }
    If_CluTruthShrink( t, iVar, nVarsAll, uSupp );
    return 1;
}

// returns the best group found
func If_CluCheck(p *If_Man_t, pTruth0 *word, nVars int32, iVarStart int32, iVarStop int32, nLutLeaf int32, nLutRoot int32, pR *If_Grp_t, pFunc0 *word, pFunc1 *word, pLeftOver *word, fHashing int32) If_Grp_t {
//    int fEnableHashing = 0;
    var G1 If_Grp_t ;
    var R If_Grp_t ;
    var pHashed *uint32  = nil;
    var Truth word;
    var pTruth [CLU_WRD_MAX]word;
    var pF [CLU_WRD_MAX]word;//, pG[CLU_WRD_MAX];
    var V2P [CLU_VAR_MAX+2]int32;
    var P2V [CLU_VAR_MAX+2]int32;
    var pCanonPerm [CLU_VAR_MAX]int32;
    var i, nSupp, uCanonPhase int32;
    var nLutSize int32;
    if p {
        nLutSize = p.pPars.nLutSize;
    } else {
        nLutSize = nVars;
    }
    assert(nVars <= CLU_VAR_MAX, "nVars <= CLU_VAR_MAX");
    assert(nVars <= nLutLeaf + nLutRoot - 1, "nVars <= nLutLeaf + nLutRoot - 1");

    if pR != nil {
        pR.nVars = 0;
        *pFunc0 = 0;
        *pFunc1 = 0;
    }

    // canonicize truth table
    If_CluCopy( pTruth, pTruth0, nLutSize );

    if false {
        uCanonPhase = If_CluSemiCanonicize( pTruth, nVars, pCanonPerm );
        If_CluAdjustBig( pTruth, nVars, nLutSize );
    }

//    If_CluSemiCanonicizeVerify( pTruth, pTruth0, nVars, pCanonPerm, uCanonPhase );
//    If_CluCopy( pTruth, pTruth0, nLutSize );

 /*
    {
        int pCanonPerm[32];
        short pStore[32];
        unsigned uCanonPhase;
        If_CluCopy( pF, pTruth, nVars );
        uCanonPhase = Kit_TruthSemiCanonicize( pF, pG, nVars, pCanonPerm );
        G1.nVars = 1;
        return G1;
    }
*/
    // check minnimum base
    If_CluCopy( pF, pTruth, nVars );
    for  i = 0; i < nVars; i++  {
        V2P[i] = P2V[i] = i;
    }
    // check support
    nSupp = If_CluSupport( pF, nVars );
//Extra_PrintBinary( stdout, &nSupp, 16 );  printf( "\n" );
    if !nSupp || !(If_CluSuppIsMinBase(nSupp) != 0) {
//        assert( 0 );     
        return G1;
    }

    // check hash table
    if p && (fHashing != 0) {
        pHashed = If_CluHashLookup( p, pTruth, 0 );
        if (pHashed != nil) && *pHashed != CLU_UNUSED {
            If_CluUns2Grp( *pHashed, &G1 ); }
    }

    // update the variable order so that the first var was the last one
    if iVarStop != 0 {
        If_CluMoveVar( pF, nVars, V2P, P2V, 0, nVars-1 ); }

    if G1.nVars == 0 {
        s_Count2++;

        // detect easy cofs
        if iVarStart == 0 {
            G1 = If_CluDecUsingCofs( pTruth, nVars, nLutLeaf ); }
        if G1.nVars == 0 {
            // perform testing
            G1 = If_CluFindGroup( pF, nVars, iVarStart, iVarStop, V2P, P2V, nLutLeaf, nLutLeaf + nLutRoot == nVars + 1 );
    //        If_CluCheckPerm( pTruth, pF, nVars, V2P, P2V );
            if G1.nVars == 0 {
                // perform testing with a smaller set
                if nVars < nLutLeaf + nLutRoot - 2 {
                    nLutLeaf--;
                    G1 = If_CluFindGroup( pF, nVars, iVarStart, iVarStop, V2P, P2V, nLutLeaf, nLutLeaf + nLutRoot == nVars + 1 );
                    nLutLeaf++;
                }
                // perform testing with a smaller set
                if nLutLeaf > 4 && nVars < nLutLeaf + nLutRoot - 3 {
                    nLutLeaf--;
                    nLutLeaf--;
                    G1 = If_CluFindGroup( pF, nVars, iVarStart, iVarStop, V2P, P2V, nLutLeaf, nLutLeaf + nLutRoot == nVars + 1 );
                    nLutLeaf++;
                    nLutLeaf++;
                }
                if G1.nVars == 0 {
                    // perform testing with a different order
                    If_CluReverseOrder( pF, nVars, V2P, P2V, iVarStart );
                    G1 = If_CluFindGroup( pF, nVars, iVarStart, iVarStop, V2P, P2V, nLutLeaf, nLutLeaf + nLutRoot == nVars + 1 );

                    // check permutation
    //                If_CluCheckPerm( pTruth, pF, nVars, V2P, P2V );
                    if G1.nVars == 0 {
                        // remember free set, just in case
//                        for ( i = 0; i < nVars - nLutLeaf; i++ )
///                           G1.pVars[nLutLeaf+i] = P2V[i];
                        // if <XY>, this will not be used
                        // if <XYZ>, this will not be hashed

    /*
                        if ( nVars == 6 )
                        {
                            Extra_PrintHex( stdout, (unsigned *)pF, nVars );  printf( "    " );
                            Kit_DsdPrintFromTruth( (unsigned*)pF, nVars );  printf( "\n" );
                            if ( !If_CutPerformCheck07( (unsigned *)pF, 6, 6, NULL ) )
                                printf( "no\n" );
                        } 
    */
                        if pHashed != nil {
                            *pHashed = If_CluGrp2Uns( &G1 ); }
                        return G1;
                    }
                }
            }
        }
    }

    // derive
    if pR != nil {
        var iNewPos int32;

        If_CluMoveGroupToMsb( pF, nVars, V2P, P2V, &G1 );
        if G1.nMyu == 2 {
            Truth = If_CluDeriveDisjoint( pF, nVars, V2P, P2V, &G1, &R );
            iNewPos = R.nVars - 1;
        
        } else {
            Truth = If_CluDeriveNonDisjoint( pF, nVars, V2P, P2V, &G1, &R );
            iNewPos = R.nVars - 2;
        }
        assert(R.pVars[iNewPos] == nVars, "R.pVars[iNewPos] == nVars");

        // adjust the functions
        Truth = If_CluAdjust( Truth, G1.nVars );
        if R.nVars < 6 {
            pF[0] = If_CluAdjust( pF[0], R.nVars ); }

//        Kit_DsdPrintFromTruth( (unsigned*)&Truth, G1.nVars ); printf( "  ...1\n" );
//        Kit_DsdPrintFromTruth( (unsigned*)pF, R.nVars );      printf( "  ...1\n" );

        // update the variable order of R so that the new var was the first one
//        if ( iVarStart == 0 )
        {
            var k int32;
            var V2P2 [CLU_VAR_MAX+2]int32;
            var P2V2 [CLU_VAR_MAX+2]int32;
            assert(iNewPos >= iVarStart, "iNewPos >= iVarStart");
            for  k = 0; k < R.nVars; k++  {
                V2P2[k] = P2V2[k] = k;
            }
            If_CluMoveVar( pF, R.nVars, V2P2, P2V2, iNewPos, iVarStart );
            for  k = iNewPos; k > iVarStart; k--  {
                R.pVars[k] = R.pVars[k-1];
            }
            R.pVars[iVarStart] = nVars;
        }

//        Kit_DsdPrintFromTruth( (unsigned*)pF, R.nVars ); printf( "  ...2\n" );

        if pLeftOver != nil {
            if R.nVars < 6 {
                *pLeftOver = If_CluAdjust( pF[0], R.nVars );
            } else {
                If_CluCopy( pLeftOver, pF, R.nVars ); }
            If_CluAdjustBig( pLeftOver, R.nVars, nLutSize );
        }

        // perform checking
        if false {
            If_CluCheckGroup( pTruth, nVars, &G1 );
            If_CluVerify( pTruth, nVars, &G1, &R, Truth, pF );
        } 

        // save functions
        *pR = R;
        if pFunc0 != nil {
            *pFunc0 = pF[0]; }
        if pFunc1 != nil {
            *pFunc1 = Truth; }
    }

    if pHashed != nil {
        *pHashed = If_CluGrp2Uns( &G1 ); }
    return G1;
}

/*
static inline word Abc_Tt6Cofactor0( word t, int iVar )
{
    assert( iVar >= 0 && iVar < 6 );
    return (t &~Truth6[iVar]) | ((t &~Truth6[iVar]) << (1<<iVar));
}
static inline word Abc_Tt6Cofactor1( word t, int iVar )
{
    assert( iVar >= 0 && iVar < 6 );
    return (t & Truth6[iVar]) | ((t & Truth6[iVar]) >> (1<<iVar));
}
*/
func If_CluCheckDecInAny(t word, nVars int32) int32 {
    var v int32;
    var u int32;
    var Cof2 [2]int32;
    var Cof4 [4]int32;
    for  v = 0; v < nVars; v++  {
        Cof2[0] = Abc_Tt6Cofactor0( t, v );
        Cof2[1] = Abc_Tt6Cofactor1( t, v );
        for  u = v+1; u < nVars; u++  {
            Cof4[0] = Abc_Tt6Cofactor0( Cof2[0], u );
            Cof4[1] = Abc_Tt6Cofactor1( Cof2[0], u );
            Cof4[2] = Abc_Tt6Cofactor0( Cof2[1], u );
            Cof4[3] = Abc_Tt6Cofactor1( Cof2[1], u );
            if Cof4[0] == Cof4[1] && Cof4[0] == Cof4[2] {
                return 1; }
            if Cof4[0] == Cof4[2] && Cof4[0] == Cof4[3] {
                return 1; }
            if Cof4[0] == Cof4[1] && Cof4[0] == Cof4[3] {
                return 1; }
            if Cof4[1] == Cof4[2] && Cof4[1] == Cof4[3] {
                return 1; }
        }
    }
    return 0;
}
func If_CluCheckDecIn(t word, nVars int32) int32 {
    var v int32;
    var u int32;
    var Cof2 [2]int32;
    var Cof4 [4]int32;
//    for ( v = 0; v < nVars; v++ )
    for  v = 0; v < 1; v++  { // restrict to the first (decomposed) input
    
        Cof2[0] = Abc_Tt6Cofactor0( t, v );
        Cof2[1] = Abc_Tt6Cofactor1( t, v );
        for  u = v+1; u < nVars; u++  {
            Cof4[0] = Abc_Tt6Cofactor0( Cof2[0], u );
            Cof4[1] = Abc_Tt6Cofactor1( Cof2[0], u );
            Cof4[2] = Abc_Tt6Cofactor0( Cof2[1], u );
            Cof4[3] = Abc_Tt6Cofactor1( Cof2[1], u );
            if Cof4[0] == Cof4[1] && Cof4[0] == Cof4[2] {
                return 1; }
            if Cof4[0] == Cof4[2] && Cof4[0] == Cof4[3] {
                return 1; }
            if Cof4[0] == Cof4[1] && Cof4[0] == Cof4[3] {
                return 1; }
            if Cof4[1] == Cof4[2] && Cof4[1] == Cof4[3] {
                return 1; }
        }
    }
    return 0;
}
func If_CluCheckDecInU(t word, nVars int32) int32 {
    var v int32;
    var u int32;
    var Cof2 [2]int32;
    var Cof4 [4]int32;
//    for ( v = 0; v < nVars; v++ )
    for  v = 0; v < 1; v++  { // restrict to the first (decomposed) input
    
        Cof2[0] = Abc_Tt6Cofactor0( t, v );
        Cof2[1] = Abc_Tt6Cofactor1( t, v );
        for  u = v+1; u < nVars; u++  {
            Cof4[0] = Abc_Tt6Cofactor0( Cof2[0], u ); // 00
            Cof4[1] = Abc_Tt6Cofactor1( Cof2[0], u ); // 01
            Cof4[2] = Abc_Tt6Cofactor0( Cof2[1], u ); // 10
            Cof4[3] = Abc_Tt6Cofactor1( Cof2[1], u ); // 11 
            if Cof4[0] == Cof4[1] && Cof4[0] == Cof4[2] { //  F * a
                return 1; }
            if Cof4[0] == Cof4[2] && Cof4[0] == Cof4[3] { // !F * a
                return 1; }
        }
    }
    return 0;
}
func If_CluCheckDecOut(t word, nVars int32) int32 {
    var v int32;
    for  v = 0; v < nVars; v++  {
        if (t & Truth6[v]) == 0   ||  //  F * !a
             (~t & Truth6[v]) == 0  ||  // !F * !a
             (t & ~Truth6[v]) == 0  ||  //  F *  a
             (~t & ~Truth6[v]) == 0     // !F *  a   
            {
            return 1; }
    }
    return 0;
}
func If_CluCheckDecOutU(t word, nVars int32) int32 {
    var v int32;
    for  v = 0; v < nVars; v++  {
        if (t & ~Truth6[v]) == 0  ||  //  F *  a
             (~t & ~Truth6[v]) == 0     // !F *  a   
            {
            return 1; }
    }
    return 0;
}

func If_CutPerformCheck45(p *If_Man_t, pTruth *uint32, nVars int32, nLeaves int32, pStr string) int32 {    
    // 5LUT -> 4LUT
    var G, R If_Grp_t;
    var Func0, Func1 word;
    G = If_CluCheck( p, (word *)pTruth, nLeaves, 0, 0, 5, 4, &R, &Func0, &Func1, nil, 0 );
    if G.nVars == 0 {
        return 0; }
    Func0 = If_CluAdjust( Func0, R.nVars );
    Func1 = If_CluAdjust( Func1, G.nVars );
     
                                                           
                                                                        
                                                                        
                           
                           
      
    if G.nVars < 5 || (p.pPars.fEnableCheck75 && (If_CluCheckDecOut(Func1, 5) != 0)) || (p.pPars.fEnableCheck75u && (If_CluCheckDecOutU(Func1, 5) != 0)) {
        return 1; }
    return 0;
}
func If_CutPerformCheck54(p *If_Man_t, pTruth *uint32, nVars int32, nLeaves int32, pStr string) int32 {    
    // 4LUT -> 5LUT
    var G, R If_Grp_t;
    var Func0, Func1 word;
    G = If_CluCheck( p, (word *)pTruth, nLeaves, 0, 0, 4, 5, &R, &Func0, &Func1, nil, 0 );
    if G.nVars == 0 {
        return 0; }
    Func0 = If_CluAdjust( Func0, R.nVars );
    Func1 = If_CluAdjust( Func1, G.nVars );
     
                                                           
                                                                        
                                                                        
                           
                           
      
    if R.nVars < 5 || (p.pPars.fEnableCheck75 && (If_CluCheckDecIn(Func0, 5) != 0)) || (p.pPars.fEnableCheck75u && (If_CluCheckDecInU(Func0, 5) != 0)) {
        return 1; }
    return 0;
}

// returns the best group found
var spStatic_If_CluCheck3_Counter int32;
func If_CluCheck3(p *If_Man_t, pTruth0 *word, nVars int32, nLutLeaf int32, nLutLeaf2 int32, nLutRoot int32, pR *If_Grp_t, pG2 *If_Grp_t, pFunc0 *word, pFunc1 *word, pFunc2 *word) If_Grp_t {
    var fEnableHashing int32;
    
    var pHashed *uint32  = nil;
    var pLeftOver [CLU_WRD_MAX]word;
    var Func0 word;
    var Func1 word;
    var Func2 word;
    var G1 If_Grp_t ;
    var G2 If_Grp_t ;
    var R If_Grp_t ;
    var R2 If_Grp_t ;
    var i int32;
    spStatic_If_CluCheck3_Counter++;
//    if ( Counter == 37590 )
//    {
//        int ns = 0;
//    }

    // check hash table
    if p && (fEnableHashing != 0) {
        pHashed = If_CluHashLookup( p, pTruth0, 1 );
        if (pHashed != nil) && *pHashed != CLU_UNUSED {
            If_CluUns2Grp( *pHashed, &G1 );
            return G1;
        }
    }
    s_Count3++;

    // check two-node decomposition
    G1 = If_CluCheck( p, pTruth0, nVars, 0, 0, nLutLeaf, nLutRoot + nLutLeaf2 - 1, &R2, &Func0, &Func1, pLeftOver, 0 );
    // decomposition does not exist
    if G1.nVars == 0 {
        // check for decomposition with two outputs
        if (G1.nMyu == 3 || G1.nMyu == 4) && nLutLeaf == nLutLeaf2 && nVars - nLutLeaf + 2 <= nLutRoot {
            var V2P, P2V [CLU_VAR_MAX+2]int32;
            var Func0, Func1, Func2 word;
            var iVar0, iVar1 int32;

            G1.nVars = nLutLeaf;
            If_CluCopy( pLeftOver, pTruth0, nVars );
            for  i = 0; i < nVars; i++  {
                V2P[i] = P2V[i] = i;
            }

            If_CluMoveGroupToMsb( pLeftOver, nVars, V2P, P2V, &G1 );
            If_CluDeriveDisjoint4( pLeftOver, nVars, V2P, P2V, &G1, &R, &Func1, &Func2 );

            // move the two vars to the front
            for  i = 0; i < R.nVars; i++  {
                V2P[i] = P2V[i] = i;
            }
            If_CluMoveVar( pLeftOver, R.nVars, V2P, P2V, R.nVars-2, 0 );
            If_CluMoveVar( pLeftOver, R.nVars, V2P, P2V, R.nVars-1, 1 );
            iVar0 = R.pVars[R.nVars-2];
            iVar1 = R.pVars[R.nVars-1];
            for  i = R.nVars-1; i > 1; i--  {
                R.pVars[i] = R.pVars[i-2];
            }
            R.pVars[0] = iVar0;
            R.pVars[1] = iVar1;

            Func0 = pLeftOver[0];
            If_CluVerify3( pTruth0, nVars, &G1, &G1, &R, Func1, Func2, Func0 );
            if (pFunc1 != nil) && (pFunc2 != nil) {
                *pFunc0 = Func0;
                *pFunc1 = Func1;
                *pFunc2 = Func2;
                *pG2 = G1;
                *pR = R;
            }

            if pHashed != nil {
                *pHashed = If_CluGrp2Uns( &G1 ); }
//                Kit_DsdPrintFromTruth( (unsigned*)pTruth0, nVars );  printf( "\n" );
//                If_CluPrintGroup( &G1 );
            return G1;
        }

/*
//        if ( nVars == 6 )
        {
//            Extra_PrintHex( stdout, (unsigned *)pTruth0, nVars );  printf( "    " );
            Kit_DsdPrintFromTruth( (unsigned*)pTruth0, nVars );  printf( "\n" );
            if ( p != NULL )
            If_CluCheck3( NULL, pTruth0, nVars, nLutLeaf, nLutLeaf2, nLutRoot, pR, pG2, pFunc0, pFunc1, pFunc2 );
        } 
*/
        if pHashed != nil {
            *pHashed = If_CluGrp2Uns( &G1 ); }
        return G1;
    }
    // decomposition exists and sufficient
    if R2.nVars <= nLutRoot {
        if pG2 != nil {     *pG2 = G2; }
        if pR != nil {      *pR  = R2; }
        if pFunc0 != nil {  *pFunc0 = Func0; }
        if pFunc1 != nil {  *pFunc1 = Func1; }
        if pFunc2 != nil {  *pFunc2 = 0; }
        if pHashed != nil {
            *pHashed = If_CluGrp2Uns( &G1 ); }
        return G1;
    }

    // try second decomposition
    {
        var Test int32;
        if Test != 0 {
            Kit_DsdPrintFromTruth( (unsigned*)&pLeftOver, R2.nVars ); fmt.Printf("\n");
        }
    }

    // the new variable is at the bottom - skip it (iVarStart = 1)
    if p.pPars.nStructType == 0 { // allowed
        G2 = If_CluCheck( p, pLeftOver, R2.nVars, 0, 0, nLutLeaf2, nLutRoot, &R, &Func0, &Func2, nil, 0 );
    } else if p.pPars.nStructType == 1 { // not allowed
        G2 = If_CluCheck( p, pLeftOver, R2.nVars, 1, 0, nLutLeaf2, nLutRoot, &R, &Func0, &Func2, nil, 0 );
    } else if p.pPars.nStructType == 2 { // required
        G2 = If_CluCheck( p, pLeftOver, R2.nVars, 0, 1, nLutLeaf2, nLutRoot, &R, &Func0, &Func2, nil, 0 );
    } else { assert(false, "0"); }

    if G2.nVars == 0 {
        if pHashed != nil {
            *pHashed = If_CluGrp2Uns( &G2 ); }
        return G2;
    }
    // remap variables
    for  i = 0; i < G2.nVars; i++  {
        assert(G2.pVars[i] < R2.nVars, "G2.pVars[i] < R2.nVars");
        G2.pVars[i] = R2.pVars[ int32(G2.pVars[i]) ];
    }
    // remap variables
    for  i = 0; i < R.nVars; i++  {
        if R.pVars[i] == R2.nVars {
            R.pVars[i] = nVars + 1;
        } else {
            R.pVars[i] = R2.pVars[ int32(R.pVars[i]) ]; }
    }

    // decomposition exist
    if pG2 != nil {     *pG2 = G2; }
    if pR != nil {      *pR  = R; }
    if pFunc0 != nil {  *pFunc0 = Func0; }
    if pFunc1 != nil {  *pFunc1 = Func1; }
    if pFunc2 != nil {  *pFunc2 = Func2; }
    if pHashed != nil {
        *pHashed = If_CluGrp2Uns( &G1 ); }

    // verify
//    If_CluVerify3( pTruth0, nVars, &G1, &G2, &R, Func1, Func2, Func0 );
    return G1;
}

// returns the best group found
func If_CluCheckExt(pMan any, pTruth *word, nVars int32, nLutLeaf int32, nLutRoot int32, pLut0 string, pLut1 string, pFunc0 *word, pFunc1 *word) int32 {
    var p *If_Man_t  = (If_Man_t *)pMan;
    var G, R If_Grp_t;
    G = If_CluCheck( p, pTruth, nVars, 0, 0, nLutLeaf, nLutRoot, &R, pFunc0, pFunc1, nil, 0 );
    memcpy( pLut0, &R, sizeof(If_Grp_t) );
    memcpy( pLut1, &G, sizeof(If_Grp_t) );
//    memcpy( pLut2, &G2, sizeof(If_Grp_t) );
    return (G.nVars > 0);
}

// returns the best group found
func If_CluCheckExt3(pMan any, pTruth *word, nVars int32, nLutLeaf int32, nLutLeaf2 int32, nLutRoot int32, pLut0 string, pLut1 string, pLut2 string, pFunc0 *word, pFunc1 *word, pFunc2 *word) int32 {
    var p *If_Man_t  = (If_Man_t *)pMan;
    var G, G2, R If_Grp_t;
    G = If_CluCheck3( p, pTruth, nVars, nLutLeaf, nLutLeaf2, nLutRoot, &R, &G2, pFunc0, pFunc1, pFunc2 );
    memcpy( pLut0, &R, sizeof(If_Grp_t) );
    memcpy( pLut1, &G, sizeof(If_Grp_t) );
    memcpy( pLut2, &G2, sizeof(If_Grp_t) );
    return (G.nVars > 0);
}


// computes delay of the decomposition
func If_CluDelayMax(g *If_Grp_t, pDelays *float32) float32 {
    var Delay float32  = 0.0;
    var i int32;
    for  i = 0; i < g.nVars; i++  {
        Delay = Abc_MaxFloat( Delay, pDelays[int32(g.pVars[i])] );
    }
    return Delay;
}

// returns delay of the decomposition;  sets area of the cut as its cost
func If_CutDelayLutStruct(p *If_Man_t, pCut *If_Cut_t, pStr string, WireDelay float32) float32 {
    var Delays [CLU_VAR_MAX+2]float32;
    var fUsed [CLU_VAR_MAX+2]int32 ;
    var pLeaf *If_Obj_t;
    var G1 If_Grp_t ;
    var G2 If_Grp_t ;
    var G3 If_Grp_t ;
    var nLeaves int32  = If_CutLeaveNum(pCut);
    var i, nLutLeaf, nLutRoot int32;
    // mark the cut as user cut
//    pCut->fUser = 1;
    // quit if parameters are wrong
    if len(pStr) != 2 {
        fmt.Printf("Wrong LUT struct (%s)\n", pStr);
        return ABC_INFINITY;
    }
    nLutLeaf = pStr[0] - '0';
    if nLutLeaf < 3 || nLutLeaf > 6 {
        fmt.Printf("Leaf size (%d) should belong to {3,4,5,6}.\n", nLutLeaf);
        return ABC_INFINITY;
    }
    nLutRoot = pStr[1] - '0';
    if nLutRoot < 3 || nLutRoot > 6 {
        fmt.Printf("Root size (%d) should belong to {3,4,5,6}.\n", nLutRoot);
        return ABC_INFINITY;
    }
    if nLeaves > nLutLeaf + nLutRoot - 1 {
        fmt.Printf("The cut size (%d) is too large for the LUT structure %d%d.\n", If_CutLeaveNum(pCut), nLutLeaf, nLutRoot);
        return ABC_INFINITY;
    }

    // remember the delays
    If_CutForEachLeaf( p, pCut, pLeaf, i )
        Delays[i] = If_ObjCutBest(pLeaf).Delay;

    // consider easy case
    if nLeaves <= Abc_MaxInt( nLutLeaf, nLutRoot ) {
        var pPerm string  = If_CutPerm( pCut );
        assert(nLeaves <= 6, "nLeaves <= 6");
        for  i = 0; i < nLeaves; i++  {
            pPerm[i] = 1;
            G1.pVars[i] = i;
        }
        G1.nVars = nLeaves;
        return 1.0 + If_CluDelayMax( &G1, Delays );
    }

    // derive the first group
    G1 = If_CluCheck( p, If_CutTruthW(p, pCut), nLeaves, 0, 0, nLutLeaf, nLutRoot, nil, nil, nil, nil, 1 );
    if G1.nVars == 0 {
        return ABC_INFINITY; }

    // compute the delay
    Delays[nLeaves] = If_CluDelayMax( &G1, Delays ) + ((WireDelay == 0.0)?1.0:WireDelay);
    if G2.nVars {
        Delays[nLeaves+1] = If_CluDelayMax( &G2, Delays ) + ((WireDelay == 0.0)?1.0:WireDelay); }

    // mark used groups
    for  i = 0; i < G1.nVars; i++  {
        fUsed[int32(G1.pVars[i])] = 1;
    }
    for  i = 0; i < G2.nVars; i++  {
        fUsed[int32(G2.pVars[i])] = 1;
    }
    // mark unused groups
    assert(G1.nMyu >= 2 && G1.nMyu <= 4, "G1.nMyu >= 2 && G1.nMyu <= 4");
    if G1.nMyu > 2 {
        fUsed[int32(G1.pVars[G1.nVars-1])] = 0; }
    assert(!G2.nVars || (G2.nMyu >= 2 && G2.nMyu <= 4), "!G2.nVars || (G2.nMyu >= 2 && G2.nMyu <= 4)");
    if G2.nMyu > 2 {
        fUsed[int32(G2.pVars[G2.nVars-1])] = 0; }

    // create remaining group
    assert(G3.nVars == 0, "G3.nVars == 0");
    for  i = 0; i < nLeaves; i++  {
        if !fUsed[i] {
            G3.pVars[int32(G3.nVars++)] = i; }
    }
    G3.pVars[int32(G3.nVars++)] = nLeaves;
    if G2.nVars {
        G3.pVars[int32(G3.nVars++)] = nLeaves+1; }
    assert(G1.nVars + G2.nVars + G3.nVars == nLeaves + 
            (G1.nVars > 0) + (G2.nVars > 0) + (G1.nMyu > 2) + (G2.nMyu > 2), "G1.nVars + G2.nVars + G3.nVars == nLeaves + \n        (G1.nVars > 0) + (G2.nVars > 0) + (G1.nMyu > 2) + (G2.nMyu > 2)");
    // what if both non-disjoint vars are the same???

    pCut.Cost = 2 + (G2.nVars > 0);
    return 1.0 + If_CluDelayMax( &G3, Delays );
}

//#define IF_TRY_NEW

#ifdef IF_TRY_NEW
static Vec_Mem_t * s_vTtMem = NULL;
static Vec_Mem_t * s_vTtMem2 = NULL;
int If_TtMemCutNum()  { return Vec_MemEntryNum(s_vTtMem); }
int If_TtMemCutNum2() { return Vec_MemEntryNum(s_vTtMem2); }
//        printf( "Unique TTs = %d.  Unique classes = %d.    ", If_TtMemCutNum(), If_TtMemCutNum2() );
//        printf( "Check2 = %d.  Check3 = %d.\n", s_Count2, s_Count3 );
#endif

/**Function*************************************************************

  Synopsis    [Performs additional check.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func If_CutPerformCheck16(p *If_Man_t, pTruth0 *uint32, nVars int32, nLeaves int32, pStr string) int32 {
    var pTruth [IF_MAX_FUNC_LUTSIZE > 5 ? 1 << (IF_MAX_FUNC_LUTSIZE - 5) : 1]uint32;
    var G1 If_Grp_t  ;//, G3 = {0};
    var i, nLutLeaf, nLutLeaf2, nLutRoot, Length int32;
    // stretch the truth table
    assert(nVars >= 6, "nVars >= 6");
    memcpy( pTruth, pTruth0, sizeof(word) * Abc_TtWordNum(nVars) );
    Abc_TtStretch6( (word *)pTruth, nLeaves, p.pPars.nLutSize );

#ifdef IF_TRY_NEW
    {
        word pCopy[1024];
        int nWords = Abc_TruthWordNum(nVars);
        int iNum;
        if ( s_vTtMem == NULL )
        {
            s_vTtMem = Vec_MemAlloc( Abc_Truth6WordNum(nVars), 12 ); // 32 KB/page for 6-var functions
            Vec_MemHashAlloc( s_vTtMem, 10000 );
        }
        if ( s_vTtMem2 == NULL )
        {
            s_vTtMem2 = Vec_MemAlloc( Abc_Truth6WordNum(nVars), 12 ); // 32 KB/page for 6-var functions
            Vec_MemHashAlloc( s_vTtMem2, 10000 );
        }
        memcpy( pCopy, pTruth, sizeof(word) * Abc_Truth6WordNum(nVars) );
        if ( pCopy[0] & 1 )
            for ( i = 0; i < nWords; i++ )
                pCopy[i] = ~pCopy[i];
        iNum = Vec_MemHashInsert( s_vTtMem, pCopy );
        if ( iNum == Vec_MemEntryNum(s_vTtMem) - 1 )
        {
            int pCanonPerm[16];
            char pCanonPermC[16];
            Abc_TtCanonicize( pCopy, nVars, pCanonPermC );
//            If_CluSemiCanonicize( pCopy, nVars, pCanonPerm );
            Vec_MemHashInsert( s_vTtMem2, pCopy );
        }
    }
#endif

    // if cutmin is disabled, minimize the function
    if !p.pPars.fCutMin {
        nLeaves = Abc_TtMinBase( (word *)pTruth, nil, nLeaves, nVars ); }

    // quit if parameters are wrong
    Length = len(pStr);
    if Length != 2 && Length != 3 {
        fmt.Printf("Wrong LUT struct (%s)\n", pStr);
        return 0;
    }
    for  i = 0; i < Length; i++  {
        if pStr[i] - '0' < 3 || pStr[i] - '0' > 6 {
            fmt.Printf("The LUT size (%d) should belong to {3,4,5,6}.\n", pStr[i] - '0');
            return 0;
        }
    }

    nLutLeaf  =                   pStr[0] - '0';
    if Length == 3 {
        nLutLeaf2 = pStr[1] - '0';
    } else {
        nLutLeaf2 = 0;
    }
    nLutRoot  =                   pStr[Length-1] - '0';
    if nLeaves > nLutLeaf - 1 + (nLutLeaf2 ? nLutLeaf2 - 1 : 0) + nLutRoot {
        fmt.Printf("The cut size (%d) is too large for the LUT structure %s.\n", nLeaves, pStr);
        return 0;
    }
    // consider easy case
    if nLeaves <= Abc_MaxInt( nLutLeaf2, Abc_MaxInt(nLutLeaf, nLutRoot) ) {
        return 1; }

    // derive the first group
    if Length == 2 {
        G1 = If_CluCheck( p, (word *)pTruth, nLeaves, 0, 0, nLutLeaf, nLutRoot, nil, nil, nil, nil, 1 );
    } else {
        G1 = If_CluCheck3( p, (word *)pTruth, nLeaves, nLutLeaf, nLutLeaf2, nLutRoot, nil, nil, nil, nil, nil ); }

//    if ( G1.nVars > 0 )
//        If_CluPrintGroup( &G1 );

    return int32((G1.nVars > 0));
}


// testing procedure
func If_CluTest() {
//    word t = 0xff00f0f0ccccaaaa;
//    word t = 0xfedcba9876543210;
//    word t = 0xec64000000000000;
//    word t = 0x0100200000000001;
//    word t2[4] = { 0x0000800080008000, 0x8000000000008000, 0x8000000080008000, 0x0000000080008000 };
//    word t = 0x07770FFF07770FFF;
//    word t = 0x002000D000D00020;
//    word t = 0x000F000E000F000F;
//    word t = 0xF7FFF7F7F7F7F7F7;
//    word t = 0x0234AFDE23400BCE;
//    word t = 0x0080008880A088FF;
//    word s = t;
//    word t = 0xFFBBBBFFF0B0B0F0;
    var t word  = ABC_CONST(0x6DD9926D962D6996);
    var nVars int32      = 6;
    var nLutLeaf int32   = 4;
    var nLutLeaf2 int32  = 4;
    var nLutRoot int32   = 4;
/*
    word t2[2] = { 0x7f807f807f80807f, 0x7f807f807f807f80 };
    int nVars     = 7;
    int nLutLeaf  = 3;
    int nLutLeaf2 = 3;
    int nLutRoot  = 3;
*/

    var G If_Grp_t;
    var G2, R If_Grp_t;
    var Func0, Func1, Func2 word;


    return;

/*
    int pCanonPerm[CLU_VAR_MAX];
    int uCanonPhase;

    Kit_DsdPrintFromTruth( (unsigned*)&s, nVars ); printf( "\n" );
    uCanonPhase = If_CluSemiCanonicize( &s, nVars, pCanonPerm );
    Kit_DsdPrintFromTruth( (unsigned*)&s, nVars ); printf( "\n" );

    If_CluSemiCanonicizeVerify( &s, &t, nVars, pCanonPerm, uCanonPhase );
*/

    Kit_DsdPrintFromTruth( (unsigned*)&t, nVars ); fmt.Printf("\n");
    G = If_CluCheck3( nil, &t, nVars, nLutLeaf, nLutLeaf2, nLutRoot, &R, &G2, &Func0, &Func1, &Func2 );
    If_CluPrintGroup( &G );
    If_CluPrintGroup( &G2 );
    If_CluPrintGroup( &R );

//    If_CluVerify3( &t, nVars, &G, &G2, &R, Func1, Func2, Func0 );

    return;

//    If_CutPerformCheck07( NULL, (unsigned *)&t, 6, 6, NULL );
//    return;

//    Kit_DsdPrintFromTruth( (unsigned*)&t, nVars ); printf( "\n" );
//    G = If_CluCheck( NULL, &t, nVars, 0, nLutLeaf, nLutRoot, NULL, NULL, NULL, NULL, 0 );
//    If_CluPrintGroup( &G );
}




////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


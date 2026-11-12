package abci

import (
	spFmt "fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [abcExact.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Find minimum size networks with a SAT solver.]

  Author      [Mathias Soeken]

  Affiliation [EPFL]

  Date        [Ver. 1.0. Started - July 15, 2016.]

  Revision    [$Id: abcFanio.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

/* This implementation is based on Exercises 477 and 478 in
 * Donald E. Knuth TAOCP Fascicle 6 (Satisfiability) Section 7.2.2.2
 */

//#include "base/abc/abc.h"

//#include "aig/gia/gia.h"
//#include "misc/util/utilTruth.h"
//#include "misc/vec/vecInt.h"
//#include "misc/vec/vecPtr.h"
//#include "proof/cec/cec.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/***********************************************************************

  Synopsis    [Some truth table helper functions.]

***********************************************************************/

var s_Truths8 [32]word = [32]word{
    ABC_CONST(0xAAAAAAAAAAAAAAAA), ABC_CONST(0xAAAAAAAAAAAAAAAA), ABC_CONST(0xAAAAAAAAAAAAAAAA), ABC_CONST(0xAAAAAAAAAAAAAAAA),
    ABC_CONST(0xCCCCCCCCCCCCCCCC), ABC_CONST(0xCCCCCCCCCCCCCCCC), ABC_CONST(0xCCCCCCCCCCCCCCCC), ABC_CONST(0xCCCCCCCCCCCCCCCC),
    ABC_CONST(0xF0F0F0F0F0F0F0F0), ABC_CONST(0xF0F0F0F0F0F0F0F0), ABC_CONST(0xF0F0F0F0F0F0F0F0), ABC_CONST(0xF0F0F0F0F0F0F0F0),
    ABC_CONST(0xFF00FF00FF00FF00), ABC_CONST(0xFF00FF00FF00FF00), ABC_CONST(0xFF00FF00FF00FF00), ABC_CONST(0xFF00FF00FF00FF00),
    ABC_CONST(0xFFFF0000FFFF0000), ABC_CONST(0xFFFF0000FFFF0000), ABC_CONST(0xFFFF0000FFFF0000), ABC_CONST(0xFFFF0000FFFF0000),
    ABC_CONST(0xFFFFFFFF00000000), ABC_CONST(0xFFFFFFFF00000000), ABC_CONST(0xFFFFFFFF00000000), ABC_CONST(0xFFFFFFFF00000000),
    ABC_CONST(0x0000000000000000), ABC_CONST(0xFFFFFFFFFFFFFFFF), ABC_CONST(0x0000000000000000), ABC_CONST(0xFFFFFFFFFFFFFFFF),
    ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000), ABC_CONST(0xFFFFFFFFFFFFFFFF), ABC_CONST(0xFFFFFFFFFFFFFFFF),
};

var s_Truths8Neg [32]word = [32]word{
    ABC_CONST(0x5555555555555555), ABC_CONST(0x5555555555555555), ABC_CONST(0x5555555555555555), ABC_CONST(0x5555555555555555),
    ABC_CONST(0x3333333333333333), ABC_CONST(0x3333333333333333), ABC_CONST(0x3333333333333333), ABC_CONST(0x3333333333333333),
    ABC_CONST(0x0F0F0F0F0F0F0F0F), ABC_CONST(0x0F0F0F0F0F0F0F0F), ABC_CONST(0x0F0F0F0F0F0F0F0F), ABC_CONST(0x0F0F0F0F0F0F0F0F),
    ABC_CONST(0x00FF00FF00FF00FF), ABC_CONST(0x00FF00FF00FF00FF), ABC_CONST(0x00FF00FF00FF00FF), ABC_CONST(0x00FF00FF00FF00FF),
    ABC_CONST(0x0000FFFF0000FFFF), ABC_CONST(0x0000FFFF0000FFFF), ABC_CONST(0x0000FFFF0000FFFF), ABC_CONST(0x0000FFFF0000FFFF),
    ABC_CONST(0x00000000FFFFFFFF), ABC_CONST(0x00000000FFFFFFFF), ABC_CONST(0x00000000FFFFFFFF), ABC_CONST(0x00000000FFFFFFFF),
    ABC_CONST(0xFFFFFFFFFFFFFFFF), ABC_CONST(0x0000000000000000), ABC_CONST(0xFFFFFFFFFFFFFFFF), ABC_CONST(0x0000000000000000),
    ABC_CONST(0xFFFFFFFFFFFFFFFF), ABC_CONST(0xFFFFFFFFFFFFFFFF), ABC_CONST(0x0000000000000000), ABC_CONST(0x0000000000000000),
};

func Abc_TtIsSubsetWithMask(pSmall *word, pLarge *word, pMask *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; ++w  {
        if ( pSmall[w] & pLarge[w] & pMask[w] ) != ( pSmall[w] & pMask[w] ) {
            return 0; }
    }
    return 1;
}

func Abc_TtCofsOppositeWithMask(pTruth *word, pMask *word, nWords int32, iVar int32) int32 {
    if iVar < 6 {
        var w int32;
        var Shift int32 = ( 1 << iVar );
        for  w = 0; w < nWords; ++w  {
            if ( ( pTruth[w] << Shift ) & s_Truths6[iVar] & pMask[w] ) != ( ~pTruth[w] & s_Truths6[iVar] & pMask[w] ) {
                return 0; }
        }
        return 1;
    
    } else {
        var w int32;
        var Step int32 = ( 1 << ( iVar - 6 ) );
        var p *word = pTruth;
        var m *word = pMask;
        var pLimit *word = pTruth + nWords;
        for  ; p < pLimit; p += 2 * Step, m += 2 * Step  {
            for  w = 0; w < Step; ++w  {
                if ( p[w] & m[w] ) != ( ~p[w + Step] & m[w + Step] ) {
                    return 0; }
            }
        }
        return 1;
    }
}

// checks whether we can decompose as OP(x^p, g) where OP in {AND, OR} and p in {0, 1}
// returns p if OP = AND, and 2 + p if OP = OR
func Abc_TtIsTopDecomposable(pTruth *word, pMask *word, nWords int32, iVar int32) int32 {
    assert(iVar < 8, "iVar < 8");

    if Abc_TtIsSubsetWithMask( pTruth, &s_Truths8[iVar << 2], pMask, nWords ) != 0 { return 1; }
    if Abc_TtIsSubsetWithMask( pTruth, &s_Truths8Neg[iVar << 2], pMask, nWords ) != 0 { return 2; }
    if Abc_TtIsSubsetWithMask( &s_Truths8[iVar << 2], pTruth, pMask, nWords ) != 0 { return 3; }
    if Abc_TtIsSubsetWithMask( &s_Truths8Neg[iVar << 2], pTruth, pMask, nWords ) != 0 { return 4; }
    if Abc_TtCofsOppositeWithMask( pTruth, pMask, nWords, iVar ) != 0 { return 5; }

    return 0;
}

// checks whether we can decompose as OP(x1, OP(x2, OP(x3, ...))) where pVars = {x1, x2, x3, ...}
// OP can be different and vars can be complemented
func Abc_TtIsStairDecomposable(pTruth *word, nWords int32, pVars *int32, nSize int32, pStairFunc *int32) int32 {
    var i, d int32;
    var pMask [4]word;
    var pCopy [4]word;

    Abc_TtCopy( pCopy, pTruth, nWords, 0 );
    Abc_TtMask( pMask, nWords, nWords * 64 );

    for  i = 0; i < nSize; ++i  {
        d = Abc_TtIsTopDecomposable( pCopy, pMask, nWords, pVars[i] );
        if !(d != 0) {
            return 0; } /* not decomposable */

        pStairFunc[i] = d;

        switch  d  {
        case 1, /* AND(x, g) */
         4: /* OR(!x, g) */
            Abc_TtAnd( pMask, pMask, &s_Truths8[pVars[i] << 2], nWords, 0 );
            break;
        case 2, /* AND(!x, g) */
         3: /* OR(x, g) */
            Abc_TtAnd( pMask, pMask, &s_Truths8Neg[pVars[i] << 2], nWords, 0 );
            break;
        case 5:
            Abc_TtXor( pCopy, pCopy, &s_Truths8[pVars[i] << 2], nWords, 0 );
            break;
        }
    }

    return 1; /* decomposable */
}

/***********************************************************************

  Synopsis    [Some printing utilities.]

***********************************************************************/

func Abc_DebugPrint(str string, fCond int32) {
    if fCond != 0 {
        spFmt.Printf("%s", str);
        fflush( os.Stdout );
    }
}

func Abc_DebugPrintInt(fmt string, n int32, fCond int32) {
    if fCond != 0 {
        spFmt.Printf(fmt, n);
        fflush( os.Stdout );
    }
}

func Abc_DebugPrintIntInt(fmt string, n1 int32, n2 int32, fCond int32) {
    if fCond != 0 {
        spFmt.Printf(fmt, n1, n2);
        fflush( os.Stdout );
    }
}

func Abc_DebugErase(n int32, fCond int32) {
    var i int32;
    if fCond != 0 {
        for  i = 0; i < n; ++i  {
            spFmt.Printf("\b");
        }
        fflush( os.Stdout );
    }
}

/***********************************************************************

  Synopsis    [BMS.]

***********************************************************************/

const ABC_EXACT_SOL_NVARS =  0
const ABC_EXACT_SOL_NFUNC =  1
const ABC_EXACT_SOL_NGATES = 2

const ANSI_COLOR_RED =     "\x1b[31m"
const ANSI_COLOR_GREEN =   "\x1b[32m"
const ANSI_COLOR_YELLOW =  "\x1b[33m"
const ANSI_COLOR_BLUE =    "\x1b[34m"
const ANSI_COLOR_MAGENTA = "\x1b[35m"
const ANSI_COLOR_CYAN =    "\x1b[36m"
const ANSI_COLOR_RESET =   "\x1b[0m"

   
type Ses_Man_t struct {
    pSat  *sat_solver;                  /* SAT solver */

    pSpec        *word;                 /* specification */
    bSpecInv          int32;              /* remembers whether spec was inverted for normalization */
    nSpecVars          int32;             /* number of variables in specification */
    nSpecFunc          int32;             /* number of functions to synthesize */
    nSpecWords          int32;            /* number of words for function */
    nRows          int32;                 /* number of rows in the specification (without 0) */
    nMaxDepth          int32;             /* maximum depth (-1 if depth is not constrained) */
    nMaxDepthTmp          int32;          /* temporary copy to modify nMaxDepth temporarily */
    pArrTimeProfile         *int32;       /* arrival times of inputs (NULL if arrival times are ignored) */
    pArrTimeProfileTmp [8]int32; /* temporary copy to modify pArrTimeProfile temporarily */
    nArrTimeDelta          int32;         /* delta to the original arrival times (arrival times are normalized to have 0 as minimum element) */
    nArrTimeMax          int32;           /* maximum normalized arrival time */
    nBTLimit          int32;              /* conflict limit */
    fMakeAIG          int32;              /* create AIG instead of general network */
    fVerbose          int32;              /* be verbose */
    fVeryVerbose          int32;          /* be very verbose */
    fExtractVerbose          int32;       /* be verbose about solution extraction */
    fSatVerbose          int32;           /* be verbose about SAT solving */
    fReasonVerbose          int32;        /* be verbose about give-up reasons */
    word         pTtValues[4];          /* truth table values to assign */
    vPolar   *Vec_Int_t;                /* variables with positive polarity */
    vAssump   *Vec_Int_t;               /* assumptions */
    nRandRowAssigns          int32;       /* number of random row assignments to initialize CEGAR */
    fKeepRowAssigns          int32;       /* if 1, keep counter examples in CEGAR for next number of gates */

    nGates          int32;                /* number of gates */
    nStartGates          int32;           /* number of gates to start search (-1), i.e., to start from 1 gate, one needs to specify 0 */
    nMaxGates          int32;             /* maximum number of gates given max. delay and arrival times */
    fDecStructure          int32;         /* set to 1 or higher if nSpecFunc = 1 and f = x_i OP g(X \ {x_i}), otherwise 0 (determined when solving) */
    pDecVars          int32;              /* mask of variables that can be decomposed at top-level */
    vStairDecVars   *Vec_Int_t;         /* list of stair decomposable variables */
    pStairDecFunc [8]int32;      /* list of stair decomposable functions */
    word         pTtObjs[100];          /* temporary truth tables */

    nSimVars          int32;              /* number of simulation vars x(i, t) */
    nOutputVars          int32;           /* number of output variables g(h, i) */
    nGateVars          int32;             /* number of gate variables f(i, p, q) */
    nSelectVars          int32;           /* number of select variables s(i, j, k) */
    nDepthVars          int32;            /* number of depth variables d(i, j) */

    nSimOffset          int32;            /* offset where gate variables start */
    nOutputOffset          int32;         /* offset where output variables start */
    nGateOffset          int32;           /* offset where gate variables start */
    nSelectOffset          int32;         /* offset where select variables start */
    nDepthOffset          int32;          /* offset where depth variables start */

    fHitResLimit          int32;          /* SAT solver gave up due to resource limit */

    timeSat      abctime;               /* SAT runtime */
    timeSatSat      abctime;            /* SAT runtime (sat instance) */
    timeSatUnsat      abctime;          /* SAT runtime (unsat instance) */
    timeSatUndef      abctime;          /* SAT runtime (undef instance) */
    timeInstance      abctime;          /* creating instance runtime */
    timeTotal      abctime;             /* all runtime */

    nSatCalls          int32;             /* number of SAT calls */
    nUnsatCalls          int32;           /* number of UNSAT calls */
    nUndefCalls          int32;           /* number of UNDEF calls */

    nDebugOffset          int32;          /* for debug printing */
}

/***********************************************************************

  Synopsis    [Store truth tables based on normalized arrival times.]

***********************************************************************/

// The hash table is a list of pointers to Ses_TruthEntry_t elements, which
// are arranged in a linked list, each of which pointing to a linked list
// of Ses_TimesEntry_t elements which contain the char* representation of the
// optimum netlist according to then normalized arrival times:

   
type Ses_TimesEntry_t struct {
    pArrTimeProfile [8]int32; /* normalized arrival time profile */
    fResLimit                int32;          /* solution found after resource limit */
    next  *Ses_TimesEntry_t;               /* linked list pointer */
    pNetwork              string;           /* pointer to char array representation of optimum network */
}

   
type Ses_TruthEntry_t struct {
    word               pTruth[4]; /* truth table for comparison */
    nVars                int32;     /* number of variables */
    next  *Ses_TruthEntry_t;      /* linked list pointer */
    head  *Ses_TimesEntry_t;      /* pointer to head of sub list with arrival times */
}

const SES_STORE_TABLE_SIZE = 1024
   
type Ses_Store_t struct {
    fMakeAIG                int32;                       /* create AIG instead of general network */
    fVerbose                int32;                       /* be verbose */
    fVeryVerbose                int32;                   /* be very verbose */
    nBTLimit                int32;                       /* conflict limit */
    nEntriesCount                int32;                  /* number of entries */
    nValidEntriesCount                int32;             /* number of entries with network */
    pEntries [SES_STORE_TABLE_SIZE]*Ses_TruthEntry_t; /* hash table for truth table entries */
    pSat        *sat_solver;                           /* own SAT solver instance to reuse when calling exact algorithm */
    pDebugEntries              *os.File;                  /* debug unsynth. (rl) entries */
    szDBName              string;                       /* if given, database is written every time a new entry is added */

    /* statistics */
    unsigned long      nCutCount;                      /* number of cuts investigated */
    unsigned long      pCutCount[9];                   /* -> per cut size */
    unsigned long      nUnsynthesizedImp;              /* number of cuts which couldn't be optimized at all, opt. stopped because of imp. constraints */
    unsigned long      pUnsynthesizedImp[9];           /* -> per cut size */
    unsigned long      nUnsynthesizedRL;               /* number of cuts which couldn't be optimized at all, opt. stopped because of resource limits */
    unsigned long      pUnsynthesizedRL[9];            /* -> per cut size */
    unsigned long      nSynthesizedTrivial;            /* number of cuts which could be synthesized trivially (n < 2) */
    unsigned long      pSynthesizedTrivial[9];         /* -> per cut size */
    unsigned long      nSynthesizedImp;                /* number of cuts which could be synthesized, opt. stopped because of imp. constraints */
    unsigned long      pSynthesizedImp[9];             /* -> per cut size */
    unsigned long      nSynthesizedRL;                 /* number of cuts which could be synthesized, opt. stopped because of resource limits */
    unsigned long      pSynthesizedRL[9];              /* -> per cut size */
    unsigned long      nCacheHits;                     /* number of cache hits */
    unsigned long      pCacheHits[9];                  /* -> per cut size */

    unsigned long      nSatCalls;                      /* number of total SAT calls */
    unsigned long      nUnsatCalls;                    /* number of total UNSAT calls */
    unsigned long      nUndefCalls;                    /* number of total UNDEF calls */

    timeExact            abctime;                      /* Exact synthesis runtime */
    timeSat            abctime;                        /* SAT runtime */
    timeSatSat            abctime;                     /* SAT runtime (sat instance) */
    timeSatUnsat            abctime;                   /* SAT runtime (unsat instance) */
    timeSatUndef            abctime;                   /* SAT runtime (undef instance) */
    timeInstance            abctime;                   /* creating instance runtime */
    timeTotal            abctime;                      /* all runtime */
}

var s_pSesStore *Ses_Store_t = nil;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

func Abc_NormalizeArrivalTimes(pArrTimeProfile *int32, nVars int32, maxNormalized *int32) int32 {
    var p *int32 = pArrTimeProfile;
    var pEnd *int32 = pArrTimeProfile + nVars;
    var delta int32  = *p;

    for  ++p < pEnd  {
        if *p < delta {
            delta = *p; }
    }

    *maxNormalized = 0;
    p = pArrTimeProfile;
    for  p < pEnd  {
        *p -= delta;
        if *p > *maxNormalized {
            *maxNormalized = *p; }
        ++p;
    }

    *maxNormalized += 1;

    return delta;
}

func Ses_StoreAlloc(nBTLimit int32, fMakeAIG int32, fVerbose int32) *Ses_Store_t {
    // inlining of macro ABC_CALLOC
    var pStore *Ses_Store_t  = &Ses_Store_t{};
    pStore.fMakeAIG           = fMakeAIG;
    pStore.fVerbose           = fVerbose;
    pStore.nBTLimit           = nBTLimit;
    memset( pStore.pEntries, 0, sizeof(pStore.pEntries) );

    pStore.pSat = sat_solver_new();

    return pStore;
}

func Ses_StoreClean(pStore *Ses_Store_t) {
    var i int32;
    var pTEntry, pTEntry2 *Ses_TruthEntry_t;
    var pTiEntry, pTiEntry2 *Ses_TimesEntry_t;

    for  i = 0; i < SES_STORE_TABLE_SIZE; ++i  {
        if pStore.pEntries[i] {
            pTEntry = pStore.pEntries[i];

            for  pTEntry != nil  {
                pTiEntry = pTEntry.head;
                for  pTiEntry != nil  {
                    // inlining of macro ABC_FREE
                    ((pTiEntry.pNetwork) ? (free((char *) (pTiEntry.pNetwork)), (pTiEntry.pNetwork) = 0) : 0);
                    pTiEntry2 = pTiEntry;
                    pTiEntry = pTiEntry.next;
                    // inlining of macro ABC_FREE
                    ((pTiEntry2) ? (free((char *) (pTiEntry2)), (pTiEntry2) = 0) : 0);
                }
                pTEntry2 = pTEntry;
                pTEntry = pTEntry.next;
                // inlining of macro ABC_FREE
                ((pTEntry2) ? (free((char *) (pTEntry2)), (pTEntry2) = 0) : 0);
            }
        }
    }

    sat_solver_delete( pStore.pSat );

    if pStore.szDBName {
        // inlining of macro ABC_FREE
        ((pStore.szDBName) ? (free((char *) (pStore.szDBName)), (pStore.szDBName) = 0) : 0); }
    // inlining of macro ABC_FREE
    ((pStore) ? (free((char *) (pStore)), (pStore) = 0) : 0);
}

var spStatic_Ses_StoreTableHash_s_Primes [4]int32 = [4]int32{ 1291, 1699, 1999, 2357 };
func Ses_StoreTableHash(pTruth *word, nVars int32) int32 {
    
    var i int32;
    var uHash uint32;
    for  i = 0; i < Abc_TtWordNum( nVars ); ++i  {
        uHash ^= pTruth[i] * spStatic_Ses_StoreTableHash_s_Primes[i & 0xf];
    }
    return int32((uHash % SES_STORE_TABLE_SIZE ));
}

func Ses_StoreTruthEqual(pEntry *Ses_TruthEntry_t, pTruth *word, nVars int32) int32 {
    var i int32;

    if pEntry.nVars != nVars {
        return 0; }

    for  i = 0; i < Abc_TtWordNum( nVars ); ++i  {
        if pEntry.pTruth[i] != pTruth[i] {
            return 0; }
    }
    return 1;
}

func Ses_StoreTruthCopy(pEntry *Ses_TruthEntry_t, pTruthSrc *word, nVars int32) {
    var i int32;
    pEntry.nVars = nVars;
    for  i = 0; i < Abc_TtWordNum( nVars ); ++i  {
        pEntry.pTruth[i] = pTruthSrc[i];
    }
}

func Ses_StoreTimesEqual(pTimes1 *int32, pTimes2 *int32, nVars int32) int32 {
    var i int32;
    for  i = 0; i < nVars; ++i  {
        if pTimes1[i] != pTimes2[i] {
            return 0; }
    }
    return 1;
}

func Ses_StoreTimesCopy(pTimesDest *int32, pTimesSrc *int32, nVars int32) {
    var i int32;
    for  i = 0; i < nVars; ++i  {
        pTimesDest[i] = pTimesSrc[i];
    }
}

func Ses_StorePrintEntry(pEntry *Ses_TruthEntry_t, pTiEntry *Ses_TimesEntry_t) {
    var i int32;

    spFmt.Printf("f = ");
    Abc_TtPrintHexRev( os.Stdout, pEntry.pTruth, pEntry.nVars );
    spFmt.Printf(", n = %d", pEntry.nVars);
    spFmt.Printf(", arrival =");
    for  i = 0; i < pEntry.nVars; ++i  {
        spFmt.Printf(" %d", pTiEntry.pArrTimeProfile[i]);
    }
    spFmt.Printf("\n");
}

func Ses_StorePrintDebugEntry(pStore *Ses_Store_t, pTruth *word, nVars int32, pNormalArrTime *int32, nMaxDepth int32, pSol string, nStartGates int32) {
    var l int32;

    spFmt.Fprintf(pStore.pDebugEntries, "abc -c \"exact -v -C %d", pStore.nBTLimit);
    if s_pSesStore.fMakeAIG { spFmt.Fprintf(pStore.pDebugEntries, " -a"); }
    spFmt.Fprintf(pStore.pDebugEntries, " -S %d -D %d -A", nStartGates + 1, nMaxDepth);
    for  l = 0; l < nVars; ++l  {
        spFmt.Fprintf(pStore.pDebugEntries, "%c%d", ( l == 0 ? ' ' : ',' ), pNormalArrTime[l]);
    }
    spFmt.Fprintf(pStore.pDebugEntries, " ");
    Abc_TtPrintHexRev( pStore.pDebugEntries, pTruth, nVars );
    spFmt.Fprintf(pStore.pDebugEntries, "\" # ");

    if !pSol {
        spFmt.Fprintf(pStore.pDebugEntries, "no "); }
    spFmt.Fprintf(pStore.pDebugEntries, "solution found before\n");
}

func Abc_ExactNormalizeArrivalTimesForNetwork(nVars int32, pArrTimeProfile *int32, pSol string) {
    var nGates, i, j, k, nMax int32;
    var vLevels *Vec_Int_t;

    nGates = pSol[ABC_EXACT_SOL_NGATES];

    /* printf( "NORMALIZE\n" ); */
    /* printf( "  #vars  = %d\n", nVars ); */
    /* printf( "  #gates = %d\n", nGates ); */

    vLevels = Vec_IntAllocArrayCopy( pArrTimeProfile, nVars );

    /* compute level of each gate based on arrival time profile (to compute depth) */
    for  i = 0; i < nGates; ++i  {
        j = pSol[3 + i * 4 + 2];
        k = pSol[3 + i * 4 + 3];

        Vec_IntPush( vLevels, Abc_MaxInt( Vec_IntEntry( vLevels, j ), Vec_IntEntry( vLevels, k ) ) + 1 );

        /* printf( "  gate %d = (%d,%d)\n", nVars + i, j, k ); */
    }

    /* Vec_IntPrint( vLevels ); */

    /* reset all levels except for the last one */
    for  i = 0; i < nVars + nGates - 1; ++i  {
        Vec_IntSetEntry( vLevels, i, Vec_IntEntry( vLevels, nVars + nGates - 1 ) );
    }

    /* Vec_IntPrint( vLevels ); */

    /* compute levels from top to bottom */
    for  i = nGates - 1; i >= 0; --i  {
        j = pSol[3 + i * 4 + 2];
        k = pSol[3 + i * 4 + 3];

        Vec_IntSetEntry( vLevels, j, Abc_MinInt( Vec_IntEntry( vLevels, j ), Vec_IntEntry( vLevels, nVars + i ) - 1 ) );
        Vec_IntSetEntry( vLevels, k, Abc_MinInt( Vec_IntEntry( vLevels, k ), Vec_IntEntry( vLevels, nVars + i ) - 1 ) );
    }

    /* Vec_IntPrint( vLevels ); */

    /* normalize arrival times */
    Abc_NormalizeArrivalTimes( Vec_IntArray( vLevels ), nVars, &nMax );
    memcpy( pArrTimeProfile, Vec_IntArray( vLevels ), sizeof(int) * nVars );

    /* printf( "  nMax = %d\n", nMax ); */
    /* Vec_IntPrint( vLevels ); */

    Vec_IntFree( vLevels );
}

func Ses_StoreWrite(pStore *Ses_Store_t, pFilename string, fSynthImp int32, fSynthRL int32, fUnsynthImp int32, fUnsynthRL int32) {
    var i int32;
    var zero byte  = '\0';
    unsigned long nEntries = 0;
    var pTEntry *Ses_TruthEntry_t;
    var pTiEntry *Ses_TimesEntry_t;
    var pFile *os.File;

    pFile = os.OpenFile(pFilename, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        spFmt.Printf("cannot open file \"%s\" for writing\n", pFilename);
        return;
    }

    if fSynthImp != 0 {   nEntries += pStore.nSynthesizedImp; }
    if fSynthRL != 0 {    nEntries += pStore.nSynthesizedRL; }
    if fUnsynthImp != 0 { nEntries += pStore.nUnsynthesizedImp; }
    if fUnsynthRL != 0 {  nEntries += pStore.nUnsynthesizedRL; }
    fwrite( &nEntries, sizeof( unsigned long ), 1, pFile );

    for  i = 0; i < SES_STORE_TABLE_SIZE; ++i  {
        if pStore.pEntries[i] {
            pTEntry = pStore.pEntries[i];

            for  pTEntry != nil  {
                pTiEntry = pTEntry.head;
                for  pTiEntry != nil  {
                    if !(fSynthImp != 0) && pTiEntry.pNetwork && !pTiEntry.fResLimit    { pTiEntry = pTiEntry.next; continue; }
                    if !(fSynthRL != 0) && pTiEntry.pNetwork && pTiEntry.fResLimit      { pTiEntry = pTiEntry.next; continue; }
                    if !(fUnsynthImp != 0) && !pTiEntry.pNetwork && !pTiEntry.fResLimit { pTiEntry = pTiEntry.next; continue; }
                    if !(fUnsynthRL != 0) && !pTiEntry.pNetwork && pTiEntry.fResLimit   { pTiEntry = pTiEntry.next; continue; }

                    fwrite( pTEntry.pTruth, sizeof( word ), 4, pFile );
                    fwrite( &pTEntry.nVars, sizeof( int ), 1, pFile );
                    fwrite( pTiEntry.pArrTimeProfile, sizeof( int ), 8, pFile );
                    fwrite( &pTiEntry.fResLimit, sizeof( int ), 1, pFile );

                    if pTiEntry.pNetwork {
                        fwrite( pTiEntry.pNetwork, sizeof( char ), 3 + 4 * pTiEntry.pNetwork[ABC_EXACT_SOL_NGATES] + 2 + pTiEntry.pNetwork[ABC_EXACT_SOL_NVARS], pFile );
                    
                    } else {
                        fwrite( &zero, sizeof( char ), 1, pFile );
                        fwrite( &zero, sizeof( char ), 1, pFile );
                        fwrite( &zero, sizeof( char ), 1, pFile );
                    }

                    pTiEntry = pTiEntry.next;
                }
                pTEntry = pTEntry.next;
            }
        }
    }


    _ = (pFile).Close();
}

// pArrTimeProfile is normalized
// returns 1 if and only if a new TimesEntry has been created
func Ses_StoreAddEntry(pStore *Ses_Store_t, pTruth *word, nVars int32, pArrTimeProfile *int32, pSol string, fResLimit int32) int32 {
    var key, fAdded int32;
    var pTEntry *Ses_TruthEntry_t;
    var pTiEntry *Ses_TimesEntry_t;

    if pSol {
        Abc_ExactNormalizeArrivalTimesForNetwork( nVars, pArrTimeProfile, pSol ); }

    key = Ses_StoreTableHash( pTruth, nVars );
    pTEntry = pStore.pEntries[key];

    /* does truth table already exist? */
    for  pTEntry != nil  {
        if Ses_StoreTruthEqual( pTEntry, pTruth, nVars ) != 0 {
            break;
        } else {
            pTEntry = pTEntry.next; }
    }

    /* entry does not yet exist, so create new one and enqueue */
    if !(pTEntry != nil) {
        // inlining of macro ABC_CALLOC
        pTEntry = &Ses_TruthEntry_t{};
        Ses_StoreTruthCopy( pTEntry, pTruth, nVars );
        pTEntry.next = pStore.pEntries[key];
        pStore.pEntries[key] = pTEntry;
    }

    /* does arrival time already exist? */
    pTiEntry = pTEntry.head;
    for  pTiEntry != nil  {
        if Ses_StoreTimesEqual( pArrTimeProfile, pTiEntry.pArrTimeProfile, nVars ) != 0 {
            break;
        } else {
            pTiEntry = pTiEntry.next; }
    }

    /* entry does not yet exist, so create new one and enqueue */
    if !(pTiEntry != nil) {
        // inlining of macro ABC_CALLOC
        pTiEntry = &Ses_TimesEntry_t{};
        Ses_StoreTimesCopy( pTiEntry.pArrTimeProfile, pArrTimeProfile, nVars );
        pTiEntry.pNetwork = pSol;
        pTiEntry.fResLimit = fResLimit;
        pTiEntry.next = pTEntry.head;
        pTEntry.head = pTiEntry;

        /* item has been added */
        fAdded = 1;
        pStore.nEntriesCount++;
        if pSol {
            pStore.nValidEntriesCount++; }
    
    } else {
        //assert( 0 );
        /* item was already present */
        fAdded = 0;
    }

    /* statistics */
    if pSol {
        if fResLimit != 0 {
            pStore.nSynthesizedRL++;
            pStore.pSynthesizedRL[nVars]++;
        
        } else {
            pStore.nSynthesizedImp++;
            pStore.pSynthesizedImp[nVars]++;
        }
    
    } else {
        if fResLimit != 0 {
            pStore.nUnsynthesizedRL++;
            pStore.pUnsynthesizedRL[nVars]++;
        
        } else {
            pStore.nUnsynthesizedImp++;
            pStore.pUnsynthesizedImp[nVars]++;
        }
    }

    if (fAdded != 0) && pStore.szDBName {
        Ses_StoreWrite( pStore, pStore.szDBName, 1, 0, 0, 0 ); }

    return fAdded;
}

// pArrTimeProfile is normalized
// returns 1 if entry was in store, pSol may still be 0 if it couldn't be computed
func Ses_StoreGetEntrySimple(pStore *Ses_Store_t, pTruth *word, nVars int32, pArrTimeProfile *int32, pSol *string) int32 {
    var key int32;
    var pTEntry *Ses_TruthEntry_t;
    var pTiEntry *Ses_TimesEntry_t;

    key = Ses_StoreTableHash( pTruth, nVars );
    pTEntry = pStore.pEntries[key];

    /* find truth table entry */
    for  pTEntry != nil  {
        if Ses_StoreTruthEqual( pTEntry, pTruth, nVars ) != 0 {
            break;
        } else {
            pTEntry = pTEntry.next; }
    }

    /* no entry found? */
    if !(pTEntry != nil) {
        return 0; }

    /* find times entry */
    pTiEntry = pTEntry.head;
    for  pTiEntry != nil  {
        if Ses_StoreTimesEqual( pArrTimeProfile, pTiEntry.pArrTimeProfile, nVars ) != 0 {
            break;
        } else {
            pTiEntry = pTiEntry.next; }
    }

    /* no entry found? */
    if !(pTiEntry != nil) {
        return 0; }

    *pSol = pTiEntry.pNetwork;
    return 1;
}

func Ses_StoreGetEntry(pStore *Ses_Store_t, pTruth *word, nVars int32, pArrTimeProfile *int32, pSol *string) int32 {
    var key int32;
    var pTEntry *Ses_TruthEntry_t;
    var pTiEntry *Ses_TimesEntry_t;
    var pTimes [8]int32;

    key = Ses_StoreTableHash( pTruth, nVars );
    pTEntry = pStore.pEntries[key];

    /* find truth table entry */
    for  pTEntry != nil  {
        if Ses_StoreTruthEqual( pTEntry, pTruth, nVars ) != 0 {
            break;
        } else {
            pTEntry = pTEntry.next; }
    }

    /* no entry found? */
    if !(pTEntry != nil) {
        return 0; }

    /* find times entry */
    pTiEntry = pTEntry.head;
    for  pTiEntry != nil  {
        /* found after normalization wrt. to network */
        if pTiEntry.pNetwork {
            memcpy( pTimes, pArrTimeProfile, sizeof(int) * nVars );
            Abc_ExactNormalizeArrivalTimesForNetwork( nVars, pTimes, pTiEntry.pNetwork );

            if Ses_StoreTimesEqual( pTimes, pTiEntry.pArrTimeProfile, nVars ) != 0 {
                break; }
        
        /* found for non synthesized network */
        } else if Ses_StoreTimesEqual( pArrTimeProfile, pTiEntry.pArrTimeProfile, nVars ) != 0 {
            break;
        } else {
            pTiEntry = pTiEntry.next; }
    }

    /* no entry found? */
    if !(pTiEntry != nil) {
        return 0; }

    *pSol = pTiEntry.pNetwork;
    return 1;
}

func Ses_StoreRead(pStore *Ses_Store_t, pFilename string, fSynthImp int32, fSynthRL int32, fUnsynthImp int32, fUnsynthRL int32) {
    var i int32;
    unsigned long nEntries;
    var pTruth [4]word;
    var nVars, fResLimit int32;
    var pArrTimeProfile [8]int32;
    var pHeader [3]byte;
    var pNetwork string;
    var pFile *os.File;
    var value int32;

    if pStore.szDBName {
        spFmt.Printf("cannot read from database when szDBName is set");
        return;
    }

    pFile = os.Open(pFilename);
    if pFile == nil {
        spFmt.Printf("cannot open file \"%s\" for reading\n", pFilename);
        return;
    }

    value = fread( &nEntries, sizeof( unsigned long ), 1, pFile );

    for  i = 0; i < int32(nEntries); ++i  {
        value = fread( pTruth, sizeof( word ), 4, pFile );
        value = fread( &nVars, sizeof( int ), 1, pFile );
        value = fread( pArrTimeProfile, sizeof( int ), 8, pFile );
        value = fread( &fResLimit, sizeof( int ), 1, pFile );
        value = fread( pHeader, sizeof( char ), 3, pFile );

        if pHeader[0] == '\0' {
            pNetwork = nil;
        } else {
            // inlining of macro ABC_CALLOC
            pNetwork = ((char *) calloc((size_t)(3 + 4 * pHeader[ABC_EXACT_SOL_NGATES] + 2 + pHeader[ABC_EXACT_SOL_NVARS]), sizeof(char)));
            pNetwork[0] = pHeader[0];
            pNetwork[1] = pHeader[1];
            pNetwork[2] = pHeader[2];

            value = fread( pNetwork + 3, sizeof( char ), 4 * pHeader[ABC_EXACT_SOL_NGATES] + 2 + pHeader[ABC_EXACT_SOL_NVARS], pFile );
        }

        if !(fSynthImp != 0) && pNetwork && !(fResLimit != 0) {    continue; }
        if !(fSynthRL != 0) && pNetwork && (fResLimit != 0) {      continue; }
        if !(fUnsynthImp != 0) && !pNetwork && !(fResLimit != 0) { continue; }
        if !(fUnsynthRL != 0) && !pNetwork && (fResLimit != 0) {   continue; }

        Ses_StoreAddEntry( pStore, pTruth, nVars, pArrTimeProfile, pNetwork, fResLimit );
    }

    _ = (pFile).Close();

    spFmt.Printf("read %d entries from file\n", int64(nEntries));
}

// computes top decomposition of variables wrt. to AND and OR
func Ses_ManComputeTopDec(pSes *Ses_Man_t) {
    var l int32;
    var pMask [4]word;

    Abc_TtMask( pMask, pSes.nSpecWords, pSes.nSpecWords * 64 );

    for  l = 0; l < pSes.nSpecVars; ++l  {
        if Abc_TtIsTopDecomposable( pSes.pSpec, pMask, pSes.nSpecWords, l ) != 0 {
            pSes.pDecVars |= ( 1 << l ); }
    }
}

func Ses_ManAlloc(pTruth *word, nVars int32, nFunc int32, nMaxDepth int32, pArrTimeProfile *int32, fMakeAIG int32, nBTLimit int32, fVerbose int32) *Ses_Man_t {
    var h, i int32;

    // inlining of macro ABC_CALLOC
    var p *Ses_Man_t  = &Ses_Man_t{};
    p.pSat       = nil;
    p.bSpecInv   = 0;
    for  h = 0; h < nFunc; ++h  {
        if pTruth[h << 2] & 1 {
            for  i = 0; i < 4; ++i  {
                pTruth[(h << 2) + i] = ~pTruth[(h << 2) + i];
            }
            p.bSpecInv |= ( 1 << h );
        }
    }
    p.pSpec           = pTruth;
    p.nSpecVars       = nVars;
    p.nSpecFunc       = nFunc;
    p.nSpecWords      = Abc_TtWordNum( nVars );
    p.nRows           = ( 1 << nVars ) - 1;
    p.nMaxDepth       = nMaxDepth;
    if nMaxDepth >= 0 {
        p.pArrTimeProfile = pArrTimeProfile;
    } else {
        p.pArrTimeProfile = nil;
    }
    if p.pArrTimeProfile {
        p.nArrTimeDelta = Abc_NormalizeArrivalTimes( p.pArrTimeProfile, nVars, &p.nArrTimeMax );
    } else {
        p.nArrTimeDelta = p.nArrTimeMax = 0; }
    p.fMakeAIG        = fMakeAIG;
    p.nBTLimit        = nBTLimit;
    p.fVerbose        = fVerbose;
    p.fVeryVerbose    = 0;
    p.fExtractVerbose = 0;
    p.fSatVerbose     = 0;
    p.vPolar          = Vec_IntAlloc( 100 );
    p.vAssump         = Vec_IntAlloc( 10 );
    p.vStairDecVars   = Vec_IntAlloc( nVars );
    p.nRandRowAssigns = 2 * nVars;
    p.fKeepRowAssigns = 0;

    if p.nSpecFunc == 1 {
        Ses_ManComputeTopDec( p ); }

    srand( 0xCAFE );

    return p;
}

func Ses_ManCleanLight(pSes *Ses_Man_t) {
    var h, i int32;
    for  h = 0; h < pSes.nSpecFunc; ++h  {
        if ( pSes.bSpecInv >> h ) & 1 {
            for  i = 0; i < 4; ++i  {
                pSes.pSpec[(h << 2) + i] = ~( pSes.pSpec[(h << 2) + i] );
            } }
    }

    if pSes.pArrTimeProfile {
        for  i = 0; i < pSes.nSpecVars; ++i  {
            pSes.pArrTimeProfile[i] += pSes.nArrTimeDelta;
        } }

    Vec_IntFree( pSes.vPolar );
    Vec_IntFree( pSes.vAssump );
    Vec_IntFree( pSes.vStairDecVars );

    // inlining of macro ABC_FREE
    ((pSes) ? (free((char *) (pSes)), (pSes) = 0) : 0);
}

func Ses_ManClean(pSes *Ses_Man_t) {
    if pSes.pSat {
        sat_solver_delete( pSes.pSat ); }
    Ses_ManCleanLight( pSes );
}

/**Function*************************************************************

  Synopsis    [Access variables based on indexes.]

***********************************************************************/
func Ses_ManSimVar(pSes *Ses_Man_t, i int32, t int32) int32 {
    assert(i < pSes.nGates, "i < pSes.nGates");
    assert(t < pSes.nRows, "t < pSes.nRows");

    return pSes.nSimOffset + pSes.nRows * i + t;
}

func Ses_ManOutputVar(pSes *Ses_Man_t, h int32, i int32) int32 {
    assert(h < pSes.nSpecFunc, "h < pSes.nSpecFunc");
    assert(i < pSes.nGates, "i < pSes.nGates");

    return pSes.nOutputOffset + pSes.nGates * h + i;
}

func Ses_ManGateVar(pSes *Ses_Man_t, i int32, p int32, q int32) int32 {
    assert(i < pSes.nGates, "i < pSes.nGates");
    assert(p < 2, "p < 2");
    assert(q < 2, "q < 2");
    assert(p > 0 || q > 0, "p > 0 || q > 0");

    return pSes.nGateOffset + i * 3 + ( p << 1 ) + q - 1;
}

func Ses_ManSelectVar(pSes *Ses_Man_t, i int32, j int32, k int32) int32 {
    var a int32;
    var offset int32;

    assert(i < pSes.nGates, "i < pSes.nGates");
    assert(k < pSes.nSpecVars + i, "k < pSes.nSpecVars + i");
    assert(j < k, "j < k");

    offset = pSes.nSelectOffset;
    for  a = pSes.nSpecVars; a < pSes.nSpecVars + i; ++a  {
        offset += a * ( a - 1 ) / 2;
    }

    return offset + ( -j * ( 1 + j - 2 * ( pSes.nSpecVars + i ) ) ) / 2 + ( k - j - 1 );
}

func Ses_ManDepthVar(pSes *Ses_Man_t, i int32, j int32) int32 {
    assert(i < pSes.nGates, "i < pSes.nGates");
    assert(j <= pSes.nArrTimeMax + i, "j <= pSes.nArrTimeMax + i");

    return pSes.nDepthOffset + i * pSes.nArrTimeMax + ( ( i * ( i + 1 ) ) / 2 ) + j;
}

/**Function*************************************************************

  Synopsis    [Compute truth table from a solution.]

***********************************************************************/
func Ses_ManDeriveTruth(pSes *Ses_Man_t, pSol string, fInvert int32) *word {
    var i int32;
    var f int32;
    var j int32;
    var k int32;
    var w int32;
    var nGates int32 = pSol[ABC_EXACT_SOL_NGATES];
    var p string;
    var pTruth *word = nil;
    var pTruth0 *word;
    var pTruth1 *word;
    assert(pSol[ABC_EXACT_SOL_NFUNC] == 1, "pSol[ABC_EXACT_SOL_NFUNC] == 1");

    p = pSol + 3;

    memset( pSes.pTtObjs, 0, sizeof( word ) * 4 * nGates );

    for  i = 0; i < nGates; ++i  {
        f = *p++;
        assert(*p++ == 2, "*p++ == 2");
        j = *p++;
        k = *p++;

        if j < pSes.nSpecVars {
            pTruth0 = &s_Truths8[j << 2];
        } else {
            pTruth0 = &pSes.pTtObjs[( j - pSes.nSpecVars ) << 2];
        }
        if k < pSes.nSpecVars {
            pTruth1 = &s_Truths8[k << 2];
        } else {
            pTruth1 = &pSes.pTtObjs[( k - pSes.nSpecVars ) << 2];
        }

        pTruth = &pSes.pTtObjs[i << 2];

        if f & 1 {
            for  w = 0; w < pSes.nSpecWords; ++w  {
                pTruth[w] |= ~pTruth0[w] & pTruth1[w];
            } }
        if ( f >> 1 ) & 1 {
            for  w = 0; w < pSes.nSpecWords; ++w  {
                pTruth[w] |= pTruth0[w] & ~pTruth1[w];
            } }
        if ( f >> 2 ) & 1 {
            for  w = 0; w < pSes.nSpecWords; ++w  {
                pTruth[w] |= pTruth0[w] & pTruth1[w];
            } }
    }

    assert(Abc_Lit2Var( *p ) == nGates - 1, "Abc_Lit2Var( *p ) == nGates - 1");
    if (fInvert != 0) && Abc_LitIsCompl( *p ) {
        Abc_TtNot( pTruth, pSes.nSpecWords ); }

    return pTruth;
}

/**Function*************************************************************

  Synopsis    [Setup variables to find network with nGates gates.]

***********************************************************************/
func Ses_ManCreateVars(pSes *Ses_Man_t, nGates int32) {
    var i int32;

    if pSes.fSatVerbose {
        spFmt.Printf("create variables for network with %d functions over %d variables and %d/%d gates\n", pSes.nSpecFunc, pSes.nSpecVars, nGates, pSes.nMaxGates);
    }

    pSes.nGates      = nGates;
    pSes.nSimVars    = nGates * pSes.nRows;
    pSes.nOutputVars = pSes.nSpecFunc * nGates;
    pSes.nGateVars   = nGates * 3;
    pSes.nSelectVars = 0;
    for  i = pSes.nSpecVars; i < pSes.nSpecVars + nGates; ++i  {
        pSes.nSelectVars += ( i * ( i - 1 ) ) / 2;
    }
    if pSes.nMaxDepth > 0 {
        pSes.nDepthVars = nGates * pSes.nArrTimeMax + ( nGates * ( nGates + 1 ) ) / 2;
    } else {
        pSes.nDepthVars = 0;
    }

    /* pSes->nSimOffset    = 0; */
    /* pSes->nOutputOffset = pSes->nSimVars; */
    /* pSes->nGateOffset   = pSes->nSimVars + pSes->nOutputVars; */
    /* pSes->nSelectOffset = pSes->nSimVars + pSes->nOutputVars + pSes->nGateVars; */
    /* pSes->nDepthOffset  = pSes->nSimVars + pSes->nOutputVars + pSes->nGateVars + pSes->nSelectVars; */

    pSes.nDepthOffset  = 0;
    pSes.nSelectOffset = pSes.nDepthVars;
    pSes.nGateOffset   = pSes.nDepthVars + pSes.nSelectVars;
    pSes.nOutputOffset = pSes.nDepthVars + pSes.nSelectVars + pSes.nGateVars;
    pSes.nSimOffset    = pSes.nDepthVars + pSes.nSelectVars + pSes.nGateVars + pSes.nOutputVars;

    /* pSes->nDepthOffset  = 0; */
    /* pSes->nSelectOffset = pSes->nDepthVars; */
    /* pSes->nOutputOffset = pSes->nDepthVars + pSes->nSelectVars; */
    /* pSes->nGateOffset   = pSes->nDepthVars + pSes->nSelectVars + pSes->nOutputVars; */
    /* pSes->nSimOffset    = pSes->nDepthVars + pSes->nSelectVars + pSes->nOutputVars + pSes->nGateVars; */

    /* if we already have a SAT solver, then restart it (this saves a lot of time) */
    if pSes.pSat {
        sat_solver_restart( pSes.pSat );
    } else {
        pSes.pSat = sat_solver_new(); }
    sat_solver_setnvars( pSes.pSat, pSes.nSimVars + pSes.nOutputVars + pSes.nGateVars + pSes.nSelectVars + pSes.nDepthVars );
}

/**Function*************************************************************

  Synopsis    [Create clauses.]

***********************************************************************/
func Ses_ManGateCannotHaveChild(pSes *Ses_Man_t, i int32, j int32) {
    var k int32;

    for  k = 0; k < j; ++k  {
        Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManSelectVar( pSes, i, k, j ), 1 ) );
    }
    for  k = j + 1; k < pSes.nSpecVars + i; ++k  {
        Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, k ), 1 ) );
    }
}

func Ses_ManCreateMainClause(pSes *Ses_Man_t, t int32, i int32, j int32, k int32, a int32, b int32, c int32) {
    var pLits [5]int32;
    var ctr int32;

    pLits[ctr++] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, k ), 1 );
    pLits[ctr++] = Abc_Var2Lit( Ses_ManSimVar( pSes, i, t ), a );

    if j < pSes.nSpecVars {
        if ( ( ( t + 1 ) & ( 1 << j ) ) ? 1 : 0 ) != b {
            return; }
    
    } else {
        pLits[ctr++] = Abc_Var2Lit( Ses_ManSimVar( pSes, j - pSes.nSpecVars, t ), b ); }

    if k < pSes.nSpecVars {
        if ( ( ( t + 1 ) & ( 1 << k ) ) ? 1 : 0 ) != c {
            return; }
    
    } else {
        pLits[ctr++] = Abc_Var2Lit( Ses_ManSimVar( pSes, k - pSes.nSpecVars, t ), c ); }

    if b > 0 || c > 0 {
        pLits[ctr++] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, b, c ), 1 - a ); }

    sat_solver_addclause( pSes.pSat, pLits, pLits + ctr );
}

func Ses_ManCreateTruthTableClause(pSes *Ses_Man_t, t int32) int inline {
    var i, j, k, h int32;
    var pLits [3]int32;

    for  i = 0; i < pSes.nGates; ++i  {
        /* main clauses */
        for  j = 0; j < pSes.nSpecVars + i; ++j  {
            for  k = j + 1; k < pSes.nSpecVars + i; ++k  {
                Ses_ManCreateMainClause( pSes, t, i, j, k, 0, 0, 1 );
                Ses_ManCreateMainClause( pSes, t, i, j, k, 0, 1, 0 );
                Ses_ManCreateMainClause( pSes, t, i, j, k, 0, 1, 1 );
                Ses_ManCreateMainClause( pSes, t, i, j, k, 1, 0, 0 );
                Ses_ManCreateMainClause( pSes, t, i, j, k, 1, 0, 1 );
                Ses_ManCreateMainClause( pSes, t, i, j, k, 1, 1, 0 );
                Ses_ManCreateMainClause( pSes, t, i, j, k, 1, 1, 1 );
            }
        }

        /* output clauses */
        if pSes.nSpecFunc != 1 {
            for  h = 0; h < pSes.nSpecFunc; ++h  {
                pLits[0] = Abc_Var2Lit( Ses_ManOutputVar( pSes, h, i ), 1 );
                pLits[1] = Abc_Var2Lit( Ses_ManSimVar( pSes, i, t ), 1 - Abc_TtGetBit( &pSes.pSpec[h << 2], t + 1 ) );
                if !sat_solver_addclause( pSes.pSat, pLits, pLits + 2 ) {
                    return 0; }
            } }
    }

    if pSes.nSpecFunc == 1 {
        Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManSimVar( pSes, pSes.nGates - 1, t ), 1 - Abc_TtGetBit( pSes.pSpec, t + 1 ) ) ); }

    return 1;
}

func Ses_ManCreateClauses(pSes *Ses_Man_t) int32 {
    extern int Extra_TruthVarsSymm( unsigned * pTruth, int nVars, int iVar0, int iVar1 );

    var h, i, j, k, t, ii, jj, kk, iii, p, q int32;
    var pLits [3]int32;
    var vLits *Vec_Int_t  = nil;

    for  t = 0; t < pSes.nRows; ++t  {
        if Abc_TtGetBit( pSes.pTtValues, t ) {
            if !(Ses_ManCreateTruthTableClause( pSes, t ) != 0) {
                return 0; } }
    }

    /* if there is only one output, we know it must point to the last gate  */
    if pSes.nSpecFunc == 1 {
        for  i = 0; i < pSes.nGates - 1; ++i  {
            Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManOutputVar( pSes, 0, i ), 1 ) );
        }
        Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManOutputVar( pSes, 0, pSes.nGates - 1 ), 0 ) );

        vLits = Vec_IntAlloc( uint32(0) );
    
    } else {
        vLits = Vec_IntAlloc( uint32(0) );

        /* some output is selected */
        for  h = 0; h < pSes.nSpecFunc; ++h  {
            Vec_IntGrowResize( vLits, pSes.nGates );
            for  i = 0; i < pSes.nGates; ++i  {
                Vec_IntSetEntry( vLits, i, Abc_Var2Lit( Ses_ManOutputVar( pSes, h, i ), 0 ) );
            }
            sat_solver_addclause( pSes.pSat, Vec_IntArray( vLits ), Vec_IntArray( vLits ) + pSes.nGates );
        }
    }

    /* each gate has two operands */
    for  i = 0; i < pSes.nGates; ++i  {
        Vec_IntGrowResize( vLits, ( ( pSes.nSpecVars + i ) * ( pSes.nSpecVars + i - 1 ) ) / 2 );
        jj = 0;
        for  j = 0; j < pSes.nSpecVars + i; ++j  {
            for  k = j + 1; k < pSes.nSpecVars + i; ++k  {
                Vec_IntSetEntry( vLits, jj++, Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, k ), 0 ) );
            }
        }
        sat_solver_addclause( pSes.pSat, Vec_IntArray( vLits ), Vec_IntArray( vLits ) + jj );
    }

    /* gate decomposition (makes it worse) */
    /* if ( fDecVariable >= 0 && pSes->nGates >= 2 ) */
    /* { */
    /*     pLits[0] = Abc_Var2Lit( Ses_ManSelectVar( pSes, pSes->nGates - 1, fDecVariable, pSes->nSpecVars + pSes->nGates - 2 ), 0 ); */
    /*     if ( !sat_solver_addclause( pSes->pSat, pLits, pLits + 1 ) ) */
    /*     { */
    /*         Vec_IntFree( vLits ); */
    /*         return 0; */
    /*     } */

    /*     for ( k = 1; k < pSes->nSpecVars + pSes->nGates - 1; ++k ) */
    /*         for ( j = 0; j < k; ++j ) */
    /*             if ( j != fDecVariable || k != pSes->nSpecVars + pSes->nGates - 2 ) */
    /*             { */
    /*                 pLits[0] = Abc_Var2Lit( Ses_ManSelectVar( pSes, pSes->nGates - 1, j, k ), 1 ); */
    /*                 if ( !sat_solver_addclause( pSes->pSat, pLits, pLits + 1 ) ) */
    /*                 { */
    /*                     Vec_IntFree( vLits ); */
    /*                     return 0; */
    /*                 } */
    /*             } */
    /* } */

    /* only AIG */
    if pSes.fMakeAIG {
        for  i = 0; i < pSes.nGates; ++i  {
            /* not 2 ones */
            pLits[0] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 0, 1 ), 1 );
            pLits[1] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 0 ), 1 );
            pLits[2] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 1 ), 0 );
            sat_solver_addclause( pSes.pSat, pLits, pLits + 3 );

            pLits[0] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 0, 1 ), 1 );
            pLits[1] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 0 ), 0 );
            pLits[2] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 1 ), 1 );
            sat_solver_addclause( pSes.pSat, pLits, pLits + 3 );

            pLits[0] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 0, 1 ), 0 );
            pLits[1] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 0 ), 1 );
            pLits[2] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 1 ), 1 );
            sat_solver_addclause( pSes.pSat, pLits, pLits + 3 );
        }
    }

    /* only binary clauses */
    if true { /* TODO: add some meaningful parameter */
    
        for  i = 0; i < pSes.nGates; ++i  {
            pLits[0] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 0, 1 ), 0 );
            pLits[1] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 0 ), 0 );
            pLits[2] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 1 ), 0 );
            sat_solver_addclause( pSes.pSat, pLits, pLits + 3 );

            pLits[0] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 0, 1 ), 1 );
            pLits[1] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 0 ), 0 );
            pLits[2] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 1 ), 1 );
            sat_solver_addclause( pSes.pSat, pLits, pLits + 3 );

            pLits[0] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 0, 1 ), 0 );
            pLits[1] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 0 ), 1 );
            pLits[2] = Abc_Var2Lit( Ses_ManGateVar( pSes, i, 1, 1 ), 1 );
            sat_solver_addclause( pSes.pSat, pLits, pLits + 3 );
        }

        for  i = 0; i < pSes.nGates; ++i  {
            for  k = 1; k < pSes.nSpecVars + i; ++k  {
                for  j = 0; j < k; ++j  {
                    pLits[0] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, k ), 1 );

                    for  kk = 1; kk < pSes.nSpecVars + i; ++kk  {
                        for  jj = 0; jj < kk; ++jj  {
                            if k == kk && j == jj { continue; }
                            pLits[1] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i, jj, kk ), 1 );

                            if pLits[0] < pLits[1] {
                                sat_solver_addclause( pSes.pSat, pLits, pLits + 2 ); }
                        }
                    }
                }
            }
        }

        /* Vec_IntGrowResize( vLits, pSes->nGates * ( pSes->nSpecVars + pSes->nGates - 2 ) ); */

        /* for ( j = 0; j < pSes->nSpecVars; ++j ) */
        /* { */
        /*     jj = 0; */
        /*     for ( i = 0; i < pSes->nGates; ++i ) */
        /*     { */
        /*         for ( k = 0; k < j; ++k ) */
        /*             Vec_IntSetEntry( vLits, jj++, Abc_Var2Lit( Ses_ManSelectVar( pSes, i, k, j ), 0 ) ); */
        /*         for ( k = j + 1; k < pSes->nSpecVars + i; ++k ) */
        /*             Vec_IntSetEntry( vLits, jj++, Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, k ), 0 ) ); */
        /*     } */
        /*     sat_solver_addclause( pSes->pSat, Vec_IntArray( vLits ), Vec_IntArray( vLits ) + jj ); */
        /* } */
    }

    /* EXTRA stair decomposition */
    if Vec_IntSize( pSes.vStairDecVars ) {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(pSes.vStairDecVars)) && (((j) = Vec_IntEntry(pSes.vStairDecVars, i)), 1); i++  {
            if pSes.nGates - 2 - i < j {
                continue;
            }
            Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManSelectVar( pSes, pSes.nGates - 1 - i, j, pSes.nSpecVars + pSes.nGates - 2 - i ), 0 ) );

            //printf( "dec %d for var %d\n", pSes->pStairDecFunc[i], j );

            switch  pSes.pStairDecFunc[i]  {
            case 1: /* AND(x,g) */
                Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes.nGates - 1 - i, 0, 1 ), 1 ) );
                //Vec_IntPush( pSes->vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes->nGates - 1 - i, 1, 0 ), 1 ) );
                //Vec_IntPush( pSes->vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes->nGates - 1 - i, 1, 1 ), 0 ) );
                break;
            case 2: /* AND(!x,g) */
                //Vec_IntPush( pSes->vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes->nGates - 1 - i, 0, 1 ), 0 ) );
                Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes.nGates - 1 - i, 1, 0 ), 1 ) );
                Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes.nGates - 1 - i, 1, 1 ), 1 ) );
                break;
            case 3: /* OR(x,g) */
                //Vec_IntPush( pSes->vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes->nGates - 1 - i, 0, 1 ), 0 ) );
                Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes.nGates - 1 - i, 1, 0 ), 0 ) );
                Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes.nGates - 1 - i, 1, 1 ), 0 ) );
                break;
            case 4: /* OR(!x,g) */
                ////Vec_IntPush( pSes->vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes->nGates - 1 - i, 0, 1 ), 0 ) );
                ////Vec_IntPush( pSes->vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes->nGates - 1 - i, 1, 0 ), 1 ) );
                ////Vec_IntPush( pSes->vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes->nGates - 1 - i, 1, 1 ), 0 ) );
                break;
            case 5: /* XOR(x,g) */
                Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes.nGates - 1 - i, 0, 1 ), 0 ) );
                Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes.nGates - 1 - i, 1, 0 ), 0 ) );
                Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManGateVar( pSes, pSes.nGates - 1 - i, 1, 1 ), 1 ) );
                break;
            default:
                spFmt.Printf("func: %d\n", pSes.pStairDecFunc[i]);
                assert(false, "0");
            }
        }
    }

    /* EXTRA clauses: use gate i at least once */
    Vec_IntGrowResize( vLits, pSes.nSpecFunc + pSes.nGates * ( pSes.nSpecVars + pSes.nGates - 2 ) );
    for  i = 0; i < pSes.nGates; ++i  {
        jj = 0;
        for  h = 0; h < pSes.nSpecFunc; ++h  {
            Vec_IntSetEntry( vLits, jj++, Abc_Var2Lit( Ses_ManOutputVar( pSes, h, i ), 0 ) );
        }
        for  ii = i + 1; ii < pSes.nGates; ++ii  {
            for  j = 0; j < pSes.nSpecVars + i; ++j  {
                Vec_IntSetEntry( vLits, jj++, Abc_Var2Lit( Ses_ManSelectVar( pSes, ii, j, pSes.nSpecVars + i ), 0 ) );
            }
            for  j = pSes.nSpecVars + i + 1; j < pSes.nSpecVars + ii; ++j  {
                Vec_IntSetEntry( vLits, jj++, Abc_Var2Lit( Ses_ManSelectVar( pSes, ii, pSes.nSpecVars + i, j ), 0 ) );
            }
        }
        sat_solver_addclause( pSes.pSat, Vec_IntArray( vLits ), Vec_IntArray( vLits ) + jj );
    }
    Vec_IntFree( vLits );

    /* EXTRA clauses: no reapplying operand */
    if pSes.nGates > 1 {
        for  i = 0; i < pSes.nGates - 1; ++i  {
            for  ii = i + 1; ii < pSes.nGates; ++ii  {
                for  k = 1; k < pSes.nSpecVars + i; ++k  {
                    for  j = 0; j < k; ++j  {
                        pLits[0] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, k ), 1 );
                        pLits[1] = Abc_Var2Lit( Ses_ManSelectVar( pSes, ii, j, pSes.nSpecVars + i ), 1 );
                        sat_solver_addclause( pSes.pSat, pLits, pLits + 2 );

                        pLits[1] = Abc_Var2Lit( Ses_ManSelectVar( pSes, ii, k, pSes.nSpecVars + i ), 1 );
                        sat_solver_addclause( pSes.pSat, pLits, pLits + 2 );
                    }
                }
            }
        } }
    if pSes.nGates > 2 {
        for  i = 0; i < pSes.nGates - 2; ++i  {
            for  ii = i + 1; ii < pSes.nGates - 1; ++ii  {
                for  iii = ii + 1; iii < pSes.nGates; ++iii  {
                    for  k = 1; k < pSes.nSpecVars + i; ++k  {
                        for  j = 0; j < k; ++j  {
                                pLits[0] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, k ), 1 );
                                pLits[1] = Abc_Var2Lit( Ses_ManSelectVar( pSes, ii, j, k ), 1 );
                                pLits[2] = Abc_Var2Lit( Ses_ManSelectVar( pSes, iii, i, ii ), 1 );
                                sat_solver_addclause( pSes.pSat, pLits, pLits + 3 );
                            }
                    }
                }
            }
        } }

    /* EXTRA clauses: co-lexicographic order */
    for  i = 0; i < pSes.nGates - 1; ++i  {
        for  k = 2; k < pSes.nSpecVars + i; ++k  {
            for  j = 1; j < k; ++j  {
                for  jj = 0; jj < j; ++jj  {
                    pLits[0] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, k ), 1 );
                    pLits[1] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i + 1, jj, k ), 1 );
                    sat_solver_addclause( pSes.pSat, pLits, pLits + 2 );
                }
            }

            for  j = 0; j < k; ++j  {
                for  kk = 1; kk < k; ++kk  {
                    for  jj = 0; jj < kk; ++jj  {
                        pLits[0] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, k ), 1 );
                        pLits[1] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i + 1, jj, kk ), 1 );
                        sat_solver_addclause( pSes.pSat, pLits, pLits + 2 );
                    }
                }
            }
        }
    }

    /* EXTRA clauses: symmetric variables */
    if  /*pSes->nMaxDepth == -1 &&*/ pSes.nSpecFunc == 1 { /* only check if there is one output function */
        for  q = 1; q < pSes.nSpecVars; ++q  {
            for  p = 0; p < q; ++p  {
                if Extra_TruthVarsSymm( (unsigned*)( pSes.pSpec ), pSes.nSpecVars, p, q ) &&
                     ( !pSes.pArrTimeProfile || ( pSes.pArrTimeProfile[p] <= pSes.pArrTimeProfile[q] ) ) {
                    if pSes.fSatVerbose {
                        spFmt.Printf("variables %d and %d are symmetric\n", p, q); }
                    for  i = 0; i < pSes.nGates; ++i  {
                        for  j = 0; j < q; ++j  {
                            if j == p { continue; }

                            vLits = Vec_IntAlloc( 0 );
                            Vec_IntPush( vLits, Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, q ), 1 ) );
                            for  ii = 0; ii < i; ++ii  {
                                for  kk = 1; kk < pSes.nSpecVars + ii; ++kk  {
                                    for  jj = 0; jj < kk; ++jj  {
                                        if jj == p || kk == p {
                                            Vec_IntPush( vLits, Abc_Var2Lit( Ses_ManSelectVar( pSes, ii, jj, kk ), 0 ) ); }
                                    }
                                }
                            }
                            sat_solver_addclause( pSes.pSat, Vec_IntArray( vLits ), Vec_IntLimit( vLits ) );
                            Vec_IntFree( vLits );
                        }
                    }
                }
            }
        } }

    return 1;
}

func Ses_ManCreateDepthClauses(pSes *Ses_Man_t) int32 {
    var i, j, k, jj, kk, d, h int32;
    var pLits [3]int32;

    for  i = 0; i < pSes.nGates; ++i  {
        /* propagate depths from children */
        for  k = 1; k < i; ++k  {
            for  j = 0; j < k; ++j  {
                pLits[0] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i, pSes.nSpecVars + j, pSes.nSpecVars + k ), 1 );
                for  jj = 0; jj <= pSes.nArrTimeMax + j; ++jj  {
                    pLits[1] = Abc_Var2Lit( Ses_ManDepthVar( pSes, j, jj ), 1 );
                    pLits[2] = Abc_Var2Lit( Ses_ManDepthVar( pSes, i, jj + 1 ), 0 );
                    sat_solver_addclause( pSes.pSat, pLits, pLits + 3 );
                }
            }
        }

        for  k = 0; k < i; ++k  {
            for  j = 0; j < pSes.nSpecVars + k; ++j  {
                pLits[0] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, pSes.nSpecVars + k ), 1 );
                for  kk = 0; kk <= pSes.nArrTimeMax + k; ++kk  {
                    pLits[1] = Abc_Var2Lit( Ses_ManDepthVar( pSes, k, kk ), 1 );
                    pLits[2] = Abc_Var2Lit( Ses_ManDepthVar( pSes, i, kk + 1 ), 0 );
                    sat_solver_addclause( pSes.pSat, pLits, pLits + 3 );
                }
            }
        }

        /* propagate depths from arrival times at PIs */
        if pSes.pArrTimeProfile {
            for  k = 1; k < pSes.nSpecVars + i; ++k  {
                for  j = 0; j < ( ( k < pSes.nSpecVars ) ? k : pSes.nSpecVars ); ++j  {
                    d = pSes.pArrTimeProfile[j];
                    if k < pSes.nSpecVars && pSes.pArrTimeProfile[k] > d {
                        d = pSes.pArrTimeProfile[k]; }

                    pLits[0] = Abc_Var2Lit( Ses_ManSelectVar( pSes, i, j, k ), 1 );
                    pLits[1] = Abc_Var2Lit( Ses_ManDepthVar( pSes, i, d ), 0 );
                    sat_solver_addclause( pSes.pSat, pLits, pLits + 2 );
                }
            }
        
        } else {
            /* arrival times are 0 */
            Vec_IntPush( pSes.vAssump, Abc_Var2Lit( Ses_ManDepthVar( pSes, i, 0 ), 0 ) ); }

        /* reverse order encoding of depth variables */
        for  j = 1; j <= pSes.nArrTimeMax + i; ++j  {
            pLits[0] = Abc_Var2Lit( Ses_ManDepthVar( pSes, i, j ), 1 );
            pLits[1] = Abc_Var2Lit( Ses_ManDepthVar( pSes, i, j - 1 ), 0 );
            sat_solver_addclause( pSes.pSat, pLits, pLits + 2 );
        }

        /* constrain maximum depth */
        if pSes.nMaxDepth < pSes.nArrTimeMax + i {
            for  h = 0; h < pSes.nSpecFunc; ++h  {
                pLits[0] = Abc_Var2Lit( Ses_ManOutputVar( pSes, h, i ), 1 );
                pLits[1] = Abc_Var2Lit( Ses_ManDepthVar( pSes, i, pSes.nMaxDepth ), 1 );
                if !sat_solver_addclause( pSes.pSat, pLits, pLits + 2 ) {
                    return 0; }
            } }
    }

    return 1;
}

/**Function*************************************************************

  Synopsis    [Solve.]

***********************************************************************/
func Sat_Wrd2Dbl(w word) float64 { return float64(uint32((w&0x3FFFFFFF))) + float64((1<<30))*uint32((w>>30)); }

func Ses_ManSolve(pSes *Ses_Man_t) int32 {
    var status int32;
    var timeStart, timeDelta abctime;

    if pSes.fSatVerbose {
        spFmt.Printf("SAT   CL: %7d   VA: %5d", sat_solver_nclauses( pSes.pSat ), sat_solver_nvars( pSes.pSat ));
        fflush( os.Stdout );
    }

    timeStart = Abc_Clock();
    status = sat_solver_solve( pSes.pSat, Vec_IntArray( pSes.vAssump ), Vec_IntLimit( pSes.vAssump ), pSes.nBTLimit, 0, 0, 0 );
    timeDelta = Abc_Clock() - timeStart;

    if pSes.fSatVerbose {
        spFmt.Printf("   RE:   %2d   ST: %4.f   CO: %7.0f   DE: %6.0f    PR: %6.0f\n", status, Sat_Wrd2Dbl( pSes.pSat.stats.starts ), Sat_Wrd2Dbl( pSes.pSat.stats.conflicts ), Sat_Wrd2Dbl( pSes.pSat.stats.decisions ), Sat_Wrd2Dbl( pSes.pSat.stats.propagations )); }
        //Sat_SolverPrintStats( stdout, pSes->pSat );

    pSes.timeSat += timeDelta;

    if status == l_True {
        pSes.nSatCalls++;
        pSes.timeSatSat += timeDelta;
        return 1;
    
    } else if status == l_False {
        pSes.nUnsatCalls++;
        pSes.timeSatUnsat += timeDelta;
        return 0;
    
    } else {
        pSes.nUndefCalls++;
        pSes.timeSatUndef += timeDelta;
        if pSes.fSatVerbose {
            //Sat_SolverWriteDimacs( pSes->pSat, "/tmp/test.cnf", Vec_IntArray( pSes->vAssump ), Vec_IntLimit( pSes->vAssump ), 1 );
            //exit( 42 );

            spFmt.Printf("resource limit reached\n");
        }
        return 2;
    }
}

/**Function*************************************************************

  Synopsis    [Extract solution.]

***********************************************************************/
// char is an array of short integers that stores the synthesized network
// using the following format
// | nvars | nfunc | ngates | gate[1] | ... | gate[ngates] | func[1] | .. | func[nfunc] |
// nvars:       integer with number of variables
// nfunc:       integer with number of functions
// ngates:      integer with number of gates
// gate[i]:     | op | nfanin | fanin[1] | ... | fanin[nfanin] |
//   op:        integer of gate's truth table (divided by 2, because gate is normal)
//   nfanin:    integer with number of fanins
//   fanin[i]:  integer to primary input or other gate
// func[i]:     | fanin | delay | pin[1] | ... | pin[nvars] |
//   fanin:     integer as literal to some gate (not primary input), can be complemented
//   delay:     maximum delay to output (taking arrival times into account, not normalized) or 0 if not specified
//   pin[i]:    pin to pin delay to input i or 0 if not specified or if there is no connection to input i
// NOTE: since outputs can only point to gates, delay and pin-to-pin times cannot be 0
func Ses_ManExtractSolution(pSes *Ses_Man_t) string {
    var nSol, h, i, j, k, l, aj, ak, d, nOp int32;
    var pSol, p string;
    var pPerm *int32  = nil; /* will be a 2d array [i][l] where is is gate id and l is PI id */

    /* compute length of solution, for now all gates have 2 inputs */
    nSol = 3 + pSes.nGates * 4 + pSes.nSpecFunc * ( 2 + pSes.nSpecVars );

    // inlining of macro ABC_CALLOC
    p = pSol = ((char *) calloc((size_t)(nSol), sizeof(char)));

    /* header */
    *p++ = pSes.nSpecVars;
    *p++ = pSes.nSpecFunc;
    *p++ = pSes.nGates;

    /* gates */
    for  i = 0; i < pSes.nGates; ++i  {
        nOp  = sat_solver_var_value( pSes.pSat, Ses_ManGateVar( pSes, i, 0, 1 ) );
        nOp |= sat_solver_var_value( pSes.pSat, Ses_ManGateVar( pSes, i, 1, 0 ) ) << 1;
        nOp |= sat_solver_var_value( pSes.pSat, Ses_ManGateVar( pSes, i, 1, 1 ) ) << 2;

        *p++ = nOp;
        *p++ = 2;

        if pSes.fExtractVerbose {
            spFmt.Printf("add gate %d with operation %d", pSes.nSpecVars + i, nOp); }

        for  k = 0; k < pSes.nSpecVars + i; ++k  {
            for  j = 0; j < k; ++j  {
                if sat_solver_var_value( pSes.pSat, Ses_ManSelectVar( pSes, i, j, k ) ) {
                    if pSes.fExtractVerbose {
                        spFmt.Printf(" and operands %d and %d", j, k); }
                    *p++ = j;
                    *p++ = k;
                    k = pSes.nSpecVars + i;
                    break;
                }
            }
        }

        if pSes.fExtractVerbose {
            if pSes.nMaxDepth > 0 {
                spFmt.Printf(" and depth vector ");
                for  j = 0; j <= pSes.nArrTimeMax + i; ++j  {
                    spFmt.Printf("%d", sat_solver_var_value( pSes.pSat, Ses_ManDepthVar( pSes, i, j ) ));
                }
            }
            spFmt.Printf("\n");
        }
    }

    /* pin-to-pin delay */
    if pSes.nMaxDepth != -1 {
        // inlining of macro ABC_CALLOC
        pPerm = ((int *) calloc((size_t)(pSes.nGates * pSes.nSpecVars), sizeof(int)));
        for  i = 0; i < pSes.nGates; ++i  {
            /* since all gates are binary for now */
            j = pSol[3 + i * 4 + 2];
            k = pSol[3 + i * 4 + 3];

            for  l = 0; l < pSes.nSpecVars; ++l  {
                /* pin-to-pin delay to input l of child nodes */
                if j < pSes.nSpecVars {
                    aj = 0;
                } else {
                    aj = pPerm[(j - pSes.nSpecVars) * pSes.nSpecVars + l];
                }
                if k < pSes.nSpecVars {
                    ak = 0;
                } else {
                    ak = pPerm[(k - pSes.nSpecVars) * pSes.nSpecVars + l];
                }

                if aj == 0 && ak == 0 {
                    pPerm[i * pSes.nSpecVars + l] = ( l == j || l == k ) ? 1 : 0;
                } else {
                    pPerm[i * pSes.nSpecVars + l] = Abc_MaxInt( aj, ak ) + 1; }
            }
        }
    }

    /* outputs */
    for  h = 0; h < pSes.nSpecFunc; ++h  {
        for  i = 0; i < pSes.nGates; ++i  {
            if sat_solver_var_value( pSes.pSat, Ses_ManOutputVar( pSes, h, i ) ) {
                *p++ = Abc_Var2Lit( i, ( pSes.bSpecInv >> h ) & 1 );
                d = 0;
                if pSes.nMaxDepth != -1 {
                    for  l = 0; l < pSes.nSpecVars; ++l  {
                        if pSes.pArrTimeProfile {
                            d = Abc_MaxInt( d, pSes.pArrTimeProfile[l] + pPerm[i * pSes.nSpecVars + l] );
                        } else {
                            d = Abc_MaxInt( d, pPerm[i * pSes.nSpecVars + l] ); }
                    } }
                *p++ = d;
                if pSes.pArrTimeProfile && pSes.fExtractVerbose {
                    spFmt.Printf("output %d points to gate %d and has normalized delay %d (nArrTimeDelta = %d)\n", h, pSes.nSpecVars + i, d, pSes.nArrTimeDelta); }
                for  l = 0; l < pSes.nSpecVars; ++l  {
                    if pSes.nMaxDepth != -1 {
                        d = pPerm[i * pSes.nSpecVars + l];
                    } else {
                        d = 0;
                    }
                    if pSes.pArrTimeProfile && pSes.fExtractVerbose {
                        spFmt.Printf("  pin-to-pin arrival time from input %d is %d (pArrTimeProfile = %d)\n", l, d, pSes.pArrTimeProfile[l]); }
                    *p++ = d;
                }
            }
        }
    }

    /* pin-to-pin delays */
    if pSes.nMaxDepth != -1 {
        // inlining of macro ABC_FREE
        ((pPerm) ? (free((char *) (pPerm)), (pPerm) = 0) : 0); }

    /* have we used all the fields? */
    assert(( p - pSol ) == nSol, "( p - pSol ) == nSol");

    /* printf( "found network: " ); */
    /* Abc_TtPrintHexRev( stdout, Ses_ManDeriveTruth( pSes, pSol, 1 ), pSes->nSpecVars ); */
    /* printf( "\n" ); */

    return pSol;
}

func Ses_ManExtractNtk(pSol string) *Abc_Ntk_t {
    var h, i int32;
    char * p;
    var pNtk *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var pGates, vNames *Vec_Ptr_t;
    var pGateTruth [5]byte;
    var pSopCover string;

    pNtk = Abc_NtkAlloc( ABC_NTK_LOGIC, ABC_FUNC_SOP, 1 );
    pNtk.pName = Extra_UtilStrsav( "exact" );
    pGates = Vec_PtrAlloc( pSol[ABC_EXACT_SOL_NVARS] + pSol[ABC_EXACT_SOL_NGATES] );
    pGateTruth[3] = '0';
    pGateTruth[4] = '\0';
    vNames = Abc_NodeGetFakeNames( pSol[ABC_EXACT_SOL_NVARS] + pSol[ABC_EXACT_SOL_NFUNC] );

    /* primary inputs */
    Vec_PtrPush( pNtk.vObjs, nil );
    for  i = 0; i < pSol[ABC_EXACT_SOL_NVARS]; ++i  {
        pObj = Abc_NtkCreatePi( pNtk );
        Abc_ObjAssignName( pObj, (char*)Vec_PtrEntry( vNames, i ), nil );
        Vec_PtrPush( pGates, pObj );
    }

    /* gates */
    p = pSol + 3;
    for  i = 0; i < pSol[ABC_EXACT_SOL_NGATES]; ++i  {
        pGateTruth[2] = '0' + ( *p & 1 );
        pGateTruth[1] = '0' + ( ( *p >> 1 ) & 1 );
        pGateTruth[0] = '0' + ( ( *p >> 2 ) & 1 );
        ++p;

        assert(*p == 2, "*p == 2"); /* binary gate */
        ++p;

        pSopCover = Abc_SopFromTruthBin( pGateTruth );
        pObj = Abc_NtkCreateNode( pNtk );
        pObj.pData = Abc_SopRegister( (Mem_Flex_t*)pNtk.pManFunc, pSopCover );
        Vec_PtrPush( pGates, pObj );
        // inlining of macro ABC_FREE
        ((pSopCover) ? (free((char *) (pSopCover)), (pSopCover) = 0) : 0);

        Abc_ObjAddFanin( pObj, (Abc_Obj_t *)Vec_PtrEntry( pGates, *p++ ) );
        Abc_ObjAddFanin( pObj, (Abc_Obj_t *)Vec_PtrEntry( pGates, *p++ ) );
    }

    /* outputs */
    for  h = 0; h < pSol[ABC_EXACT_SOL_NFUNC]; ++h  {
        pObj = Abc_NtkCreatePo( pNtk );
        Abc_ObjAssignName( pObj, (char*)Vec_PtrEntry( vNames, pSol[ABC_EXACT_SOL_NVARS] + h ), nil );
        if Abc_LitIsCompl( *p ) {
            Abc_ObjAddFanin( pObj, Abc_NtkCreateNodeInv( pNtk, (Abc_Obj_t *)Vec_PtrEntry( pGates, pSol[ABC_EXACT_SOL_NVARS] + Abc_Lit2Var( *p ) ) ) );
        } else {
            Abc_ObjAddFanin( pObj, (Abc_Obj_t *)Vec_PtrEntry( pGates, pSol[ABC_EXACT_SOL_NVARS] + Abc_Lit2Var( *p ) ) ); }
        p += ( 2 + pSol[ABC_EXACT_SOL_NVARS] );
    }
    Abc_NodeFreeNames( vNames );

    Vec_PtrFree( pGates );

    if !Abc_NtkCheck( pNtk ) {
        spFmt.Printf("Ses_ManExtractSolution(): Network check has failed.\n"); }

    return pNtk;
}

func Ses_ManExtractGia(pSol string) *Gia_Man_t {
    var h, i int32;
    char * p;
    var pGia *Gia_Man_t;
    var pGates *Vec_Int_t;
    var vNames *Vec_Ptr_t;
    var nObj, nChild1, nChild2, fChild1Comp, fChild2Comp int32;

    pGia = Gia_ManStart( pSol[ABC_EXACT_SOL_NVARS] + pSol[ABC_EXACT_SOL_NGATES] + pSol[ABC_EXACT_SOL_NFUNC] + 1 );
    pGia.nConstrs = 0;
    pGia.pName = Extra_UtilStrsav( "exact" );

    pGates = Vec_IntAlloc( pSol[ABC_EXACT_SOL_NVARS] + pSol[ABC_EXACT_SOL_NGATES] );
    vNames = Abc_NodeGetFakeNames( pSol[ABC_EXACT_SOL_NVARS] + pSol[ABC_EXACT_SOL_NFUNC] );

    /* primary inputs */
    pGia.vNamesIn = Vec_PtrStart( pSol[ABC_EXACT_SOL_NVARS] );
    for  i = 0; i < pSol[ABC_EXACT_SOL_NVARS]; ++i  {
        nObj = Gia_ManAppendCi( pGia );
        Vec_IntPush( pGates, nObj );
        Vec_PtrSetEntry( pGia.vNamesIn, i, Extra_UtilStrsav( (char*)Vec_PtrEntry( vNames, i ) ) );
    }

    /* gates */
    p = pSol + 3;
    for  i = 0; i < pSol[ABC_EXACT_SOL_NGATES]; ++i  {
        assert(p[1] == 2, "p[1] == 2");
        nChild1 = Vec_IntEntry( pGates, p[2] );
        nChild2 = Vec_IntEntry( pGates, p[3] );
        fChild1Comp = fChild2Comp = 0;

        if *p & 1 {
            nChild1 = Abc_LitNot( nChild1 );
            fChild1Comp = 1;
        }
        if ( *p >> 1 ) & 1 {
            nChild2 = Abc_LitNot( nChild2 );
            fChild2Comp = 1;
        }
        nObj = Gia_ManAppendAnd( pGia, nChild1, nChild2 );
        if fChild1Comp && (fChild2Comp != 0) {
            assert(( *p >> 2 ) & 1, "( *p >> 2 ) & 1");
            nObj = Abc_LitNot( nObj );
        }

        Vec_IntPush( pGates, nObj );

        p += 4;
    }

    /* outputs */
    pGia.vNamesOut = Vec_PtrStart( pSol[ABC_EXACT_SOL_NFUNC] );
    for  h = 0; h < pSol[ABC_EXACT_SOL_NFUNC]; ++h  {
        nObj = Vec_IntEntry( pGates, pSol[ABC_EXACT_SOL_NVARS] + Abc_Lit2Var( *p ) );
        if Abc_LitIsCompl( *p ) {
            nObj = Abc_LitNot( nObj ); }
        Gia_ManAppendCo( pGia, nObj );
        Vec_PtrSetEntry( pGia.vNamesOut, h, Extra_UtilStrsav( (char*)Vec_PtrEntry( vNames, pSol[ABC_EXACT_SOL_NVARS] + h ) ) );
        p += ( 2 + pSol[ABC_EXACT_SOL_NVARS] );
    }
    Abc_NodeFreeNames( vNames );

    Vec_IntFree( pGates );

    return pGia;
}

/**Function*************************************************************

  Synopsis    [Debug.]

***********************************************************************/
func Ses_ManPrintRuntime(pSes *Ses_Man_t) {
    spFmt.Printf("Runtime breakdown:\n");
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Sat     "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((pSes.timeSat)))/(float64(int64(time.Second))), (float64((pSes.timeTotal)))? 100.0*(float64((pSes.timeSat)))/(float64((pSes.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Sat    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((pSes.timeSatSat)))/(float64(int64(time.Second))), (float64((pSes.timeTotal)))? 100.0*(float64((pSes.timeSatSat)))/(float64((pSes.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Unsat  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((pSes.timeSatUnsat)))/(float64(int64(time.Second))), (float64((pSes.timeTotal)))? 100.0*(float64((pSes.timeSatUnsat)))/(float64((pSes.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Undef  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((pSes.timeSatUndef)))/(float64(int64(time.Second))), (float64((pSes.timeTotal)))? 100.0*(float64((pSes.timeSatUndef)))/(float64((pSes.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Instance"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((pSes.timeInstance)))/(float64(int64(time.Second))), (float64((pSes.timeTotal)))? 100.0*(float64((pSes.timeInstance)))/(float64((pSes.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("ALL     "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((pSes.timeTotal)))/(float64(int64(time.Second))), (float64((pSes.timeTotal)))? 100.0*(float64((pSes.timeTotal)))/(float64((pSes.timeTotal))) : 0.0);;
}

func Ses_ManPrintFuncs(pSes *Ses_Man_t) {
    var h int32;

    spFmt.Printf("find optimum circuit for %d %d-variable functions:\n", pSes.nSpecFunc, pSes.nSpecVars);
    for  h = 0; h < pSes.nSpecFunc; ++h  {
        spFmt.Printf("  func %d: ", h + 1);
        Abc_TtPrintHexRev( os.Stdout, &pSes.pSpec[h << 2], pSes.nSpecVars );
        spFmt.Printf("\n");
    }

    if pSes.nMaxDepth != -1 {
        spFmt.Printf("  max depth = %d\n", pSes.nMaxDepth);
        if pSes.pArrTimeProfile {
            spFmt.Printf("  arrival times =");
            for  h = 0; h < pSes.nSpecVars; ++h  {
                spFmt.Printf(" %d", pSes.pArrTimeProfile[h]);
            }
            spFmt.Printf("\n");
        }
    }
}

func Ses_ManPrintVars(pSes *Ses_Man_t) {
    var h, i, j, k, p, q, t int32;

    for  i = 0; i < pSes.nGates; ++i  {
        for  t = 0; t < pSes.nRows; ++t  {
            spFmt.Printf("x(%d, %d) : %d\n", i, t, Ses_ManSimVar( pSes, i, t ));
        }
    }

    for  h = 0; h < pSes.nSpecFunc; ++h  {
        for  i = 0; i < pSes.nGates; ++i  {
            spFmt.Printf("h(%d, %d) : %d\n", h, i, Ses_ManOutputVar( pSes, h, i ));
        }
    }

    for  i = 0; i < pSes.nGates; ++i  {
        for  p = 0; p <= 1; ++p  {
            for  q = 0; q <= 1; ++ q {
                if p == 0 && q == 0 { continue; }
                spFmt.Printf("f(%d, %d, %d) : %d\n", i, p, q, Ses_ManGateVar( pSes, i, p, q ));
            }
        }
    }

    for  i = 0; i < pSes.nGates; ++i  {
        for  j = 0; j < pSes.nSpecVars + i; ++j  {
            for  k = j + 1; k < pSes.nSpecVars + i; ++k  {
                spFmt.Printf("s(%d, %d, %d) : %d\n", i, j, k, Ses_ManSelectVar( pSes, i, j, k ));
            }
        }
    }

    if pSes.nMaxDepth > 0 {
        for  i = 0; i < pSes.nGates; ++i  {
            for  j = 0; j <= i; ++j  {
                spFmt.Printf("d(%d, %d) : %d\n", i, j, Ses_ManDepthVar( pSes, i, j ));
            }
        } }

}

/**Function*************************************************************

  Synopsis    [Synthesis algorithm.]

***********************************************************************/
func Ses_ManComputeMaxGates(pSes *Ses_Man_t) {
    var s int32 = 1;
    var lvl int32 = pSes.nMaxDepth;
    var avail int32 = pSes.nSpecVars;
    var l int32;

    pSes.nMaxGates = 0;

    /* s is the number of gates we have in the current level */
    for  (s != 0) && /* while there are nodes to branch from */
            (lvl != 0) && /* while we are at some level */
            avail * 2 > s /* while there are still enough available nodes (heuristic) */  {
        /* erase from s if we have arriving nodes */
        for  l = 0; l < pSes.nSpecVars; ++l  {
            if pSes.pArrTimeProfile[l] == lvl {
                --s;
                --avail;
            }
        }

        --lvl;
        pSes.nMaxGates += s;
        s *= 2;
    }
}

// returns 0, if current max depth and arrival times are not consistent
func Ses_CheckDepthConsistency(pSes *Ses_Man_t) int32 {
    var l, i, fAdded, nLevel int32;
    var fMaxGatesLevel2 int32  = 1;

    Vec_IntClear( pSes.vStairDecVars );
    pSes.fDecStructure = 0;

    for  l = 0; l < pSes.nSpecVars; ++l  {
        if pSes.pArrTimeProfile[l] >= pSes.nMaxDepth {
            if pSes.fReasonVerbose {
                spFmt.Printf("give up due to impossible arrival time (depth = %d, input = %d, arrival time = %d)", pSes.nMaxDepth, l, pSes.pArrTimeProfile[l]); }
            return 0;
        
        } else if pSes.nSpecFunc == 1 && pSes.pArrTimeProfile[l] + 1 == pSes.nMaxDepth {
            if ( pSes.fDecStructure == 1 && pSes.nSpecVars > 2 ) || pSes.fDecStructure == 2 {
                if pSes.fReasonVerbose {
                    spFmt.Printf("give up due to impossible decomposition (depth = %d, input = %d, arrival time = %d)", pSes.nMaxDepth, l, pSes.pArrTimeProfile[l]); }
                return 0;
            }

            pSes.fDecStructure++;

            /* A subset B <=> A and B = A */
            if !( ( pSes.pDecVars >> l ) & 1 ) {
                if pSes.fReasonVerbose {
                    spFmt.Printf("give up due to impossible decomposition (depth = %d, input = %d, arrival time = %d)", pSes.nMaxDepth, l, pSes.pArrTimeProfile[l]); }
                return 0;
            }
        }
    }

    /* more complicated decomposition */
    if pSes.fDecStructure {
        nLevel = 1;
        for {
            fAdded = 0;

            for  l = 0; l < pSes.nSpecVars; ++l  {
                if pSes.pArrTimeProfile[l] + nLevel == pSes.nMaxDepth {
                    if fAdded != 0 {
                        assert(nLevel == Vec_IntSize( pSes.vStairDecVars ), "nLevel == Vec_IntSize( pSes.vStairDecVars )");
                        if fAdded > 1 || ( nLevel + 1 < pSes.nSpecVars ) {
                            if pSes.fReasonVerbose {
                                spFmt.Printf("give up due to impossible decomposition at level %d", nLevel); }
                            return 0;
                        }
                    
                    } else {
                        Vec_IntPush( pSes.vStairDecVars, l );
                    }
                    fAdded++;
                }
            }

            if !(fAdded != 0) { break; }
            ++nLevel;
        }

        if Vec_IntSize( pSes.vStairDecVars ) && !(Abc_TtIsStairDecomposable( pSes.pSpec, pSes.nSpecWords, Vec_IntArray( pSes.vStairDecVars ), Vec_IntSize( pSes.vStairDecVars ), pSes.pStairDecFunc ) != 0) {
            if pSes.fReasonVerbose {
                spFmt.Printf("give up due to impossible stair decomposition at level %d: ", nLevel);
                Vec_IntPrint( pSes.vStairDecVars );
            }
            return 0;
        }
    }

    /* check if depth's match with structure at second level from top */
    if pSes.fDecStructure {
        if pSes.nSpecVars == 3 {
            fMaxGatesLevel2 = 2;
        } else {
            fMaxGatesLevel2 = 1;
        }
    } else {
        if pSes.nSpecVars == 4 {
            fMaxGatesLevel2 = 4;
        } else {
            fMaxGatesLevel2 = 3;
        } }

    i = 0;
    for  l = 0; l < pSes.nSpecVars; ++l  {
        if pSes.pArrTimeProfile[l] + 2 == pSes.nMaxDepth {
            if i++; i > fMaxGatesLevel2 {
                if pSes.fReasonVerbose {
                    spFmt.Printf("give up due to impossible decomposition at second level (depth = %d, input = %d, arrival time = %d)", pSes.nMaxDepth, l, pSes.pArrTimeProfile[l]); }
                return 0;
            } }
    }

    /* check if depth's match with structure at third level from top */
    if pSes.nSpecVars > 4 && pSes.fDecStructure && i == 1 { /* we have f = AND(x_i, AND(x_j, g)) (x_i and x_j may be complemented) */
    
        i = 0;
        for  l = 0; l < pSes.nSpecVars; ++l  {
            if pSes.pArrTimeProfile[l] + 3 == pSes.nMaxDepth {
                if i++; i > 1 {
                    if pSes.fReasonVerbose {
                        spFmt.Printf("give up due to impossible decomposition at third level (depth = %d, input = %d, arrival time = %d)", pSes.nMaxDepth, l, pSes.pArrTimeProfile[l]); }
                    return 0;
                } }
        }
    }

    return 1;
}

// returns 0, if current max depth and #gates are not consistent
func Ses_CheckGatesConsistency(pSes *Ses_Man_t, nGates int32) int32 {
    /* give up if number of gates is impossible for given depth */
    if pSes.nMaxDepth != -1 && nGates >= ( 1 << pSes.nMaxDepth ) {
        if pSes.fReasonVerbose {
            spFmt.Printf("give up due to impossible depth (depth = %d, gates = %d)", pSes.nMaxDepth, nGates); }
        return 0;
    }

    /* give up if number of gates is impossible for given depth and arrival times */
    if pSes.nMaxDepth != -1 && pSes.pArrTimeProfile && nGates > pSes.nMaxGates {
        if pSes.fReasonVerbose {
            spFmt.Printf("give up due to impossible depth and arrival times (depth = %d, gates = %d)", pSes.nMaxDepth, nGates); }
        return 0;
    }

    if pSes.fDecStructure && nGates >= ( 1 << ( pSes.nMaxDepth - 1 ) ) + 1 {
        if pSes.fReasonVerbose {
            spFmt.Printf("give up due to impossible depth in AND-dec structure (depth = %d, gates = %d)", pSes.nMaxDepth, nGates); }
        return 0;
    }

    /* give up if number of gates gets practically too large */
    if nGates >= ( 1 << pSes.nSpecVars ) {
        if pSes.fReasonVerbose {
            spFmt.Printf("give up due to impossible number of gates"); }
        return 0;
    }

    return 1;
}

func Abc_ExactCopyDepthAndArrivalTimes(nVars int32, nDepthFrom int32, nDepthTo *int32, pTimesFrom *int32, pTimesTo *int32) {
    var l int32;

    *nDepthTo = nDepthFrom;
    for  l = 0; l < nVars; ++l  {
        pTimesTo[l] = pTimesFrom[l];
    }
}

func Ses_ManStoreDepthAndArrivalTimes(pSes *Ses_Man_t) {
    if pSes.nMaxDepth == -1 { return; }
    Abc_ExactCopyDepthAndArrivalTimes( pSes.nSpecVars, pSes.nMaxDepth, &pSes.nMaxDepthTmp, pSes.pArrTimeProfile, pSes.pArrTimeProfileTmp );
}

func Ses_ManRestoreDepthAndArrivalTimes(pSes *Ses_Man_t) {
    if pSes.nMaxDepth == -1 { return; }
    Abc_ExactCopyDepthAndArrivalTimes( pSes.nSpecVars, pSes.nMaxDepthTmp, &pSes.nMaxDepth, pSes.pArrTimeProfileTmp, pSes.pArrTimeProfile );
}

func Abc_ExactAdjustDepthAndArrivalTimes(nVars int32, nGates int32, nDepthFrom int32, nDepthTo *int32, pTimesFrom *int32, pTimesTo *int32, nTimesMax int32) {
    var l int32;

    *nDepthTo = Abc_MinInt( nDepthFrom, nGates );
    for  l = 0; l < nVars; ++l  {
        pTimesTo[l] = Abc_MinInt( pTimesFrom[l], Abc_MaxInt( 0, nGates - 1 - nTimesMax + pTimesFrom[l] ) );
    }
}

func Ses_AdjustDepthAndArrivalTimes(pSes *Ses_Man_t, nGates int32) {
    Abc_ExactAdjustDepthAndArrivalTimes( pSes.nSpecVars, nGates, pSes.nMaxDepthTmp, &pSes.nMaxDepth, pSes.pArrTimeProfileTmp, pSes.pArrTimeProfile, pSes.nArrTimeMax - 1 );
}

/* return: (2: continue, 1: found, 0: gave up) */
func Ses_ManFindNetworkExact(pSes *Ses_Man_t, nGates int32) int32 {
    var f, fSat int32;
    var timeStart abctime;

    /* solve */
    timeStart = Abc_Clock();
    Vec_IntClear( pSes.vPolar );
    Vec_IntClear( pSes.vAssump );
    Ses_ManCreateVars( pSes, nGates );

    if pSes.nMaxDepth != -1 {
        f = Ses_ManCreateDepthClauses( pSes );
        pSes.timeInstance += ( Abc_Clock() - timeStart );
        if !(f != 0) { return 2; } /* proven UNSAT while creating clauses */
    }

    sat_solver_set_polarity( pSes.pSat, Vec_IntArray( pSes.vPolar ), Vec_IntSize( pSes.vPolar ) );

    /* first solve */
    fSat = Ses_ManSolve( pSes );
    if fSat == 0 {
        return 2; /* UNSAT, continue with next # of gates */
    } else if fSat == 2 {
        pSes.fHitResLimit = 1;
        return 0;
    }

    timeStart = Abc_Clock();
    f = Ses_ManCreateClauses( pSes );
    pSes.timeInstance += ( Abc_Clock() - timeStart );
    if !(f != 0) { return 2; } /* proven UNSAT while creating clauses */

    fSat = Ses_ManSolve( pSes );
    if fSat == 1 {
        return 1;
    } else if fSat == 2 {
        pSes.fHitResLimit = 1;
        return 0;
    }

    return 2; /* UNSAT continue */
}

// is there a network for a given number of gates
/* return: (3: impossible, 2: continue, 1: found, 0: gave up) */
func Ses_ManFindNetworkExactCEGAR(pSes *Ses_Man_t, nGates int32, pSol *string) int32 {
    var fRes, iMint, fSat, i int32;
    var pTruth [4]word;

    /* debug */
    Abc_DebugErase( pSes.nDebugOffset + ( nGates > 10 ? 5 : 4 ), pSes.fVeryVerbose );
    Abc_DebugPrintIntInt( " (%d/%d)", nGates, pSes.nMaxGates, pSes.fVeryVerbose );

    /* do #gates and max depth allow for a network? */
    if !(Ses_CheckGatesConsistency( pSes, nGates ) != 0) {
        return 3; }

    for  i = 0; i < pSes.nRandRowAssigns; ++i  {
        Abc_TtSetBit( pSes.pTtValues, rand() % pSes.nRows );
    }

    fRes = Ses_ManFindNetworkExact( pSes, nGates );
    if fRes != 1 { return fRes; }

    for  true  {
        *pSol = Ses_ManExtractSolution( pSes );
        Abc_TtXor( pTruth, Ses_ManDeriveTruth( pSes, *pSol, 0 ), pSes.pSpec, pSes.nSpecWords, 0 );
        iMint = Abc_TtFindFirstBit( pTruth, pSes.nSpecVars );

        if iMint == -1 || (pSes.nSpecVars < 6 && iMint > pSes.nRows) {
            assert(fRes == 1, "fRes == 1");
            return 1;
        }
        // inlining of macro ABC_FREE
        ((*pSol) ? (free((char *) (*pSol)), (*pSol) = 0) : 0);

        if pSes.fKeepRowAssigns {
            Abc_TtSetBit( pSes.pTtValues, iMint - 1 ); }
        if !(Ses_ManCreateTruthTableClause( pSes, iMint - 1 ) != 0) { /* UNSAT, continue */
            return 2; }

        if ( fSat = Ses_ManSolve( pSes ) ) == 1 { continue; }

        if fSat == 2 {
            return 0;
        } else {
            return 2;
        }
    }
}

// find minimum size by increasing the number of gates
func Ses_ManFindMinimumSizeBottomUp(pSes *Ses_Man_t) string {
    var nGates int32 = pSes.nStartGates;
    var fRes int32;
    var pSol string  = nil;

    /* store whether call was unsuccessful due to resource limit and not due to impossible constraint */
    pSes.fHitResLimit = 0;

    if pSes.nMaxGates >= 10 {
        pSes.nDebugOffset = 3;
    } else {
        pSes.nDebugOffset = 2;
    }

    /* adjust number of gates if there is a stair decomposition */
    if Vec_IntSize( pSes.vStairDecVars ) {
        nGates = Abc_MaxInt( nGates, Vec_IntSize( pSes.vStairDecVars ) - 1 ); }

    //Ses_ManStoreDepthAndArrivalTimes( pSes );

    memset( pSes.pTtValues, 0, 4 * sizeof( word ) );

    Abc_DebugPrintIntInt( " (%d/%d)", nGates, pSes.nMaxGates, pSes.fVeryVerbose );

    for  true  {
        ++nGates;

        fRes = Ses_ManFindNetworkExactCEGAR( pSes, nGates, &pSol );

        if fRes == 0 {
            pSes.fHitResLimit = 1;
            break;
        
        } else if fRes == 1 || fRes == 3 {
            break; }
    }

    Abc_DebugErase( pSes.nDebugOffset + ( nGates >= 10 ? 5 : 4 ), pSes.fVeryVerbose );

    return pSol;
}

func Ses_ManFindMinimumSizeTopDown(pSes *Ses_Man_t, nMinGates int32) string {
    var nGates int32 = pSes.nMaxGates;
    var fRes int32;
    var pSol string = nil;
    var pSol2 string = nil;

    pSes.fHitResLimit = 0;

    Abc_DebugPrintIntInt( " (%d/%d)", nGates, pSes.nMaxGates, pSes.fVeryVerbose );

    for  true  {
        fRes = Ses_ManFindNetworkExactCEGAR( pSes, nGates, &pSol2 );

        if fRes == 0 {
            pSes.fHitResLimit = 1;
            break;
        
        } else if fRes == 2 || fRes == 3 {
            break; }

        pSol = pSol2;

        if nGates == nMinGates {
            break; }

        --nGates;
    }

    Abc_DebugErase( pSes.nDebugOffset + ( nGates >= 10 ? 5 : 4 ), pSes.fVeryVerbose );

    return pSol;
}

func Ses_ManFindMinimumSize(pSes *Ses_Man_t) string {
    var pSol string  = nil;
    var i int32 = pSes.nStartGates + 1;
    var fRes int32;

    /* if more than one function, no CEGAR */
    if pSes.nSpecFunc > 1 {
      for  true  {
        if pSes.fVerbose {
          spFmt.Printf("try with %d gates\n", i);
        }

        memset( pSes.pTtValues, ~0, 4 * sizeof( word ) );
        fRes = Ses_ManFindNetworkExact( pSes, i++ );
        if fRes == 2 { continue; }
        if fRes == 0 { break; }

        pSol = Ses_ManExtractSolution( pSes );
        break;
      }

      return pSol;
    }

    /* do the arrival times allow for a network? */
    if pSes.nMaxDepth != -1 && pSes.pArrTimeProfile {
        if !(Ses_CheckDepthConsistency( pSes ) != 0) {
            return 0; }
        Ses_ManComputeMaxGates( pSes );
    }

    pSol = Ses_ManFindMinimumSizeBottomUp( pSes );

    if !pSol && pSes.nMaxDepth != -1 && pSes.fHitResLimit && pSes.nGates != pSes.nMaxGates {
        return Ses_ManFindMinimumSizeTopDown( pSes, pSes.nGates + 1 );
    } else {
        return pSol; }
}


/**Function*************************************************************

  Synopsis    [Find minimum size networks with a SAT solver.]

  Description [If nMaxDepth is -1, then depth constraints are ignored.
               If nMaxDepth is not -1, one can set pArrTimeProfile which should have the length of nVars.
               One can ignore pArrTimeProfile by setting it to NULL.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFindExact(pTruth *word, nVars int32, nFunc int32, nMaxDepth int32, pArrTimeProfile *int32, nBTLimit int32, nStartGates int32, fVerbose int32) *Abc_Ntk_t {
    var pSes *Ses_Man_t;
    var pSol string;
    var pNtk *Abc_Ntk_t  = nil;
    var timeStart abctime;

    /* some checks */
    assert(nVars >= 2 && nVars <= 8, "nVars >= 2 && nVars <= 8");

    timeStart = Abc_Clock();

    pSes = Ses_ManAlloc( pTruth, nVars, nFunc, nMaxDepth, pArrTimeProfile, 0, nBTLimit, fVerbose );
    pSes.nStartGates = nStartGates;
    pSes.fReasonVerbose = 0;
    pSes.fSatVerbose = 0;
    if fVerbose != 0 {
        Ses_ManPrintFuncs( pSes ); }

    if ( pSol = Ses_ManFindMinimumSize( pSes ) ) != nil {
        pNtk = Ses_ManExtractNtk( pSol );
        // inlining of macro ABC_FREE
        ((pSol) ? (free((char *) (pSol)), (pSol) = 0) : 0);
    }

    pSes.timeTotal = Abc_Clock() - timeStart;

    if fVerbose != 0 {
        Ses_ManPrintRuntime( pSes ); }

    /* cleanup */
    Ses_ManClean( pSes );

    return pNtk;
}

func Gia_ManFindExact(pTruth *word, nVars int32, nFunc int32, nMaxDepth int32, pArrTimeProfile *int32, nBTLimit int32, nStartGates int32, fVerbose int32) *Gia_Man_t {
    var pSes *Ses_Man_t;
    var pSol string;
    var pGia *Gia_Man_t  = nil;
    var timeStart abctime;

    /* some checks */
    assert(nVars >= 2 && nVars <= 8, "nVars >= 2 && nVars <= 8");

    timeStart = Abc_Clock();

    pSes = Ses_ManAlloc( pTruth, nVars, nFunc, nMaxDepth, pArrTimeProfile, 1, nBTLimit, fVerbose );
    pSes.nStartGates = nStartGates;
    pSes.fVeryVerbose = 1;
    pSes.fExtractVerbose = 0;
    pSes.fSatVerbose = 0;
    pSes.fReasonVerbose = 1;
    if fVerbose != 0 {
        Ses_ManPrintFuncs( pSes ); }

    if ( pSol = Ses_ManFindMinimumSize( pSes ) ) != nil {
        pGia = Ses_ManExtractGia( pSol );
        // inlining of macro ABC_FREE
        ((pSol) ? (free((char *) (pSol)), (pSol) = 0) : 0);
    }

    pSes.timeTotal = Abc_Clock() - timeStart;

    if fVerbose != 0 {
        Ses_ManPrintRuntime( pSes ); }

    /* cleanup */
    Ses_ManClean( pSes );

    return pGia;
}

/**Function*************************************************************

  Synopsis    [Some test cases.]

***********************************************************************/

func Abc_NtkFromTruthTable(pTruth *word, nVars int32) *Abc_Ntk_t {
    var pNtk *Abc_Ntk_t;
    var pMan *Mem_Flex_t;
    var pSopCover string;

    pMan = Mem_FlexStart();
    pSopCover = Abc_SopCreateFromTruth( pMan, nVars, (unsigned*)pTruth );
    pNtk = Abc_NtkCreateWithNode( pSopCover );
    Abc_NtkShortNames( pNtk );
    Mem_FlexStop( pMan, 0 );

    return pNtk;
}

func Abc_ExactTestSingleOutput(fVerbose int32) {
    extern void Abc_NtkCecSat( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nConfLimit, int nInsLimit );

    var pTruth [4]word = [4]word{0xcafe, 0, 0, 0};
    var pNtk, pNtk2, pNtk3, pNtk4 *Abc_Ntk_t;
    var pArrTimeProfile [4]int32 = [4]int32{6, 2, 8, 5};

    pNtk = Abc_NtkFromTruthTable( pTruth, 4 );

    pNtk2 = Abc_NtkFindExact( pTruth, 4, 1, -1, nil, 0, 0, fVerbose );
    Abc_NtkShortNames( pNtk2 );
    Abc_NtkCecSat( pNtk, pNtk2, 10000, 0 );
    assert(pNtk2, "pNtk2");
    assert(Abc_NtkNodeNum( pNtk2 ) == 6, "Abc_NtkNodeNum( pNtk2 ) == 6");
    Abc_NtkDelete( pNtk2 );

    pNtk3 = Abc_NtkFindExact( pTruth, 4, 1, 3, nil, 0, 0, fVerbose );
    Abc_NtkShortNames( pNtk3 );
    Abc_NtkCecSat( pNtk, pNtk3, 10000, 0 );
    assert(pNtk3, "pNtk3");
    assert(Abc_NtkLevel( pNtk3 ) <= 3, "Abc_NtkLevel( pNtk3 ) <= 3");
    Abc_NtkDelete( pNtk3 );

    pNtk4 = Abc_NtkFindExact( pTruth, 4, 1, 9, pArrTimeProfile, 50000, 0, fVerbose );
    Abc_NtkShortNames( pNtk4 );
    Abc_NtkCecSat( pNtk, pNtk4, 10000, 0 );
    assert((pNtk4 != nil), "pNtk4");
    assert(Abc_NtkLevel( pNtk4 ) <= 9, "Abc_NtkLevel( pNtk4 ) <= 9");
    Abc_NtkDelete( pNtk4 );

    assert(!(Abc_NtkFindExact( pTruth, 4, 1, 2, nil, 50000, 0, fVerbose ) != nil), "!Abc_NtkFindExact( pTruth, 4, 1, 2, nil, 50000, 0, fVerbose )");

    assert(!(Abc_NtkFindExact( pTruth, 4, 1, 8, pArrTimeProfile, 50000, 0, fVerbose ) != nil), "!Abc_NtkFindExact( pTruth, 4, 1, 8, pArrTimeProfile, 50000, 0, fVerbose )");

    Abc_NtkDelete( pNtk );
}

func Abc_ExactTestSingleOutputAIG(fVerbose int32) {
    var pTruth [4]word = [4]word{0xcafe, 0, 0, 0};
    var pNtk *Abc_Ntk_t;
    var pGia, pGia2, pGia3, pGia4, pMiter *Gia_Man_t;
    var ParsCec Cec_ParCec_t;
    var pPars *Cec_ParCec_t = &ParsCec;
    var pArrTimeProfile [4]int32 = [4]int32{6, 2, 8, 5};

    Cec_ManCecSetDefaultParams( pPars );

    pNtk = Abc_NtkFromTruthTable( pTruth, 4 );
    Abc_NtkToAig( pNtk );
    pGia = Abc_NtkAigToGia( pNtk, 1 );

    pGia2 = Gia_ManFindExact( pTruth, 4, 1, -1, nil, 0, 0, fVerbose );
    pMiter = Gia_ManMiter( pGia, pGia2, 0, 1, 0, 0, 1 );
    assert((pMiter != nil), "pMiter");
    Cec_ManVerify( pMiter, pPars );
    Gia_ManStop( pMiter );

    pGia3 = Gia_ManFindExact( pTruth, 4, 1, 3, nil, 0, 0, fVerbose );
    pMiter = Gia_ManMiter( pGia, pGia3, 0, 1, 0, 0, 1 );
    assert((pMiter != nil), "pMiter");
    Cec_ManVerify( pMiter, pPars );
    Gia_ManStop( pMiter );

    pGia4 = Gia_ManFindExact( pTruth, 4, 1, 9, pArrTimeProfile, 50000, 0, fVerbose );
    pMiter = Gia_ManMiter( pGia, pGia4, 0, 1, 0, 0, 1 );
    assert((pMiter != nil), "pMiter");
    Cec_ManVerify( pMiter, pPars );
    Gia_ManStop( pMiter );

    assert(!(Gia_ManFindExact( pTruth, 4, 1, 2, nil, 50000, 0, fVerbose ) != nil), "!Gia_ManFindExact( pTruth, 4, 1, 2, nil, 50000, 0, fVerbose )");

    assert(!(Gia_ManFindExact( pTruth, 4, 1, 8, pArrTimeProfile, 50000, 0, fVerbose ) != nil), "!Gia_ManFindExact( pTruth, 4, 1, 8, pArrTimeProfile, 50000, 0, fVerbose )");

    Gia_ManStop( pGia );
    Gia_ManStop( pGia2 );
    Gia_ManStop( pGia3 );
    Gia_ManStop( pGia4 );
}

func Abc_ExactTest(fVerbose int32) {
    Abc_ExactTestSingleOutput( fVerbose );
    Abc_ExactTestSingleOutputAIG( fVerbose );

    spFmt.Printf("\n");
}


/**Function*************************************************************

  Synopsis    [APIs for integraging with the mapper.]

***********************************************************************/
// may need to have a static pointer to the SAT-based synthesis engine and/or loaded library
// this procedure should return 1, if the engine/library are available, and 0 otherwise
func Abc_ExactIsRunning() int32 {
    return s_pSesStore != nil;
}
// this procedure returns the number of inputs of the library
// for example, somebody may try to map into 10-cuts while the library only contains 8-functions
func Abc_ExactInputNum() int32 {
    return 8;
}
// start exact store manager
func Abc_ExactStart(nBTLimit int32, fMakeAIG int32, fVerbose int32, fVeryVerbose int32, pFilename string) {
    if !(s_pSesStore != nil) {
        s_pSesStore = Ses_StoreAlloc( nBTLimit, fMakeAIG, fVerbose );
        s_pSesStore.fVeryVerbose = fVeryVerbose;
        if pFilename {
            Ses_StoreRead( s_pSesStore, pFilename, 1, 0, 0, 0 );

            // inlining of macro ABC_CALLOC
            s_pSesStore.szDBName = ((char *) calloc((size_t)(len(pFilename) + 1), sizeof(char)));
            strcpy( s_pSesStore.szDBName, pFilename );
        }
        if s_pSesStore.fVeryVerbose {
            s_pSesStore.pDebugEntries = os.OpenFile("bms.debug", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        }
    
    } else {
        spFmt.Printf("BMS manager already started\n"); }
}
// stop exact store manager
func Abc_ExactStop(pFilename string) {
    if s_pSesStore != nil {
        if pFilename {
            Ses_StoreWrite( s_pSesStore, pFilename, 1, 0, 0, 0 ); }
        if s_pSesStore.pDebugEntries {
            _ = (s_pSesStore.pDebugEntries).Close(); }
        Ses_StoreClean( s_pSesStore );
    
    } else {
        spFmt.Printf("BMS manager has not been started\n"); }
}
// show statistics about store manager
func Abc_ExactStats() {
    var i int32;

    if !(s_pSesStore != nil) {
        spFmt.Printf("BMS manager has not been started\n");
        return;
    }

    spFmt.Printf("-------------------------------------------------------------------------------------------------------------------------------\n");
    spFmt.Printf("                                    0         1         2         3         4         5         6         7         8     total\n");
    spFmt.Printf("-------------------------------------------------------------------------------------------------------------------------------\n");
    spFmt.Printf("number of considered cuts :");
    for  i = 0; i < 9; ++i  {
        spFmt.Printf("%10d", s_pSesStore.pCutCount[i]);
    }
    spFmt.Printf("%10d\n", s_pSesStore.nCutCount);
    spFmt.Printf(" - trivial                :");
    for  i = 0; i < 9; ++i  {
        spFmt.Printf("%10d", s_pSesStore.pSynthesizedTrivial[i]);
    }
    spFmt.Printf("%10d\n", s_pSesStore.nSynthesizedTrivial);
    spFmt.Printf(" - synth (imp)            :");
    for  i = 0; i < 9; ++i  {
        spFmt.Printf("%10d", s_pSesStore.pSynthesizedImp[i]);
    }
    spFmt.Printf("%10d\n", s_pSesStore.nSynthesizedImp);
    spFmt.Printf(" - synth (res)            :");
    for  i = 0; i < 9; ++i  {
        spFmt.Printf("%10d", s_pSesStore.pSynthesizedRL[i]);
    }
    spFmt.Printf("%10d\n", s_pSesStore.nSynthesizedRL);
    spFmt.Printf(" - not synth (imp)        :");
    for  i = 0; i < 9; ++i  {
        spFmt.Printf("%10d", s_pSesStore.pUnsynthesizedImp[i]);
    }
    spFmt.Printf("%10d\n", s_pSesStore.nUnsynthesizedImp);
    spFmt.Printf(" - not synth (res)        :");
    for  i = 0; i < 9; ++i  {
        spFmt.Printf("%10d", s_pSesStore.pUnsynthesizedRL[i]);
    }
    spFmt.Printf("%10d\n", s_pSesStore.nUnsynthesizedRL);
    spFmt.Printf(" - cache hits             :");
    for  i = 0; i < 9; ++i  {
        spFmt.Printf("%10d", s_pSesStore.pCacheHits[i]);
    }
    spFmt.Printf("%10d\n", s_pSesStore.nCacheHits);
    spFmt.Printf("-------------------------------------------------------------------------------------------------------------------------------\n");
    spFmt.Printf("number of entries         : %d\n", s_pSesStore.nEntriesCount);
    spFmt.Printf("number of valid entries   : %d\n", s_pSesStore.nValidEntriesCount);
    spFmt.Printf("number of invalid entries : %d\n", s_pSesStore.nEntriesCount - s_pSesStore.nValidEntriesCount);
    spFmt.Printf("-------------------------------------------------------------------------------------------------------------------------------\n");
    spFmt.Printf("number of SAT calls       : %d\n", s_pSesStore.nSatCalls);
    spFmt.Printf("number of UNSAT calls     : %d\n", s_pSesStore.nUnsatCalls);
    spFmt.Printf("number of UNDEF calls     : %d\n", s_pSesStore.nUndefCalls);

    spFmt.Printf("-------------------------------------------------------------------------------------------------------------------------------\n");
    spFmt.Printf("Runtime breakdown:\n");
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Exact    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((s_pSesStore.timeExact)))/(float64(int64(time.Second))), (float64((s_pSesStore.timeTotal)))? 100.0*(float64((s_pSesStore.timeExact)))/(float64((s_pSesStore.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Sat     "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((s_pSesStore.timeSat)))/(float64(int64(time.Second))), (float64((s_pSesStore.timeTotal)))? 100.0*(float64((s_pSesStore.timeSat)))/(float64((s_pSesStore.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("  Sat    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((s_pSesStore.timeSatSat)))/(float64(int64(time.Second))), (float64((s_pSesStore.timeTotal)))? 100.0*(float64((s_pSesStore.timeSatSat)))/(float64((s_pSesStore.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("  Unsat  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((s_pSesStore.timeSatUnsat)))/(float64(int64(time.Second))), (float64((s_pSesStore.timeTotal)))? 100.0*(float64((s_pSesStore.timeSatUnsat)))/(float64((s_pSesStore.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("  Undef  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((s_pSesStore.timeSatUndef)))/(float64(int64(time.Second))), (float64((s_pSesStore.timeTotal)))? 100.0*(float64((s_pSesStore.timeSatUndef)))/(float64((s_pSesStore.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Instance"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((s_pSesStore.timeInstance)))/(float64(int64(time.Second))), (float64((s_pSesStore.timeTotal)))? 100.0*(float64((s_pSesStore.timeInstance)))/(float64((s_pSesStore.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Other    "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((s_pSesStore.timeTotal - s_pSesStore.timeExact)))/(float64(int64(time.Second))), (float64((s_pSesStore.timeTotal)))? 100.0*(float64((s_pSesStore.timeTotal - s_pSesStore.timeExact)))/(float64((s_pSesStore.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("ALL      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((s_pSesStore.timeTotal)))/(float64(int64(time.Second))), (float64((s_pSesStore.timeTotal)))? 100.0*(float64((s_pSesStore.timeTotal)))/(float64((s_pSesStore.timeTotal))) : 0.0);;
}
// this procedure takes TT and input arrival times (pArrTimeProfile) and return the smallest output arrival time;
// it also returns the pin-to-pin delays (pPerm) between each cut leaf and the cut output and the cut area cost (Cost)
// the area cost should not exceed 2048, if the cut is implementable; otherwise, it should be ABC_INFINITY
func Abc_ExactDelayCost(pTruth *word, nVars int32, pArrTimeProfile *int32, pPerm string, Cost *int32, AigLevel int32) int32 {
    var i, nMaxArrival, nDelta, l int32;
    var pSes *Ses_Man_t  = nil;
    var pSol string = nil;
    var pSol2 string = nil;
    var p string;
    var pNormalArrTime [8]int32;
    var Delay int32 = ABC_INFINITY;
    var nMaxDepth int32;
    var fResLimit int32;
    var timeStart abctime = Abc_Clock();
    var timeStartExact abctime;

    /* some checks */
    if nVars < 0 || nVars > 8 {
        spFmt.Printf("invalid truth table size %d\n", nVars);
        assert(false, "0");
    }

    /* statistics */
    s_pSesStore.nCutCount++;
    s_pSesStore.pCutCount[nVars]++;

    if nVars == 0 {
        s_pSesStore.nSynthesizedTrivial++;
        s_pSesStore.pSynthesizedTrivial[0]++;

        *Cost = 0;
        s_pSesStore.timeTotal += ( Abc_Clock() - timeStart );
        return 0;
    }

    if nVars == 1 {
        s_pSesStore.nSynthesizedTrivial++;
        s_pSesStore.pSynthesizedTrivial[1]++;

        *Cost = 0;
        pPerm[0] = byte(0);
        s_pSesStore.timeTotal += ( Abc_Clock() - timeStart );
        return pArrTimeProfile[0];
    }

    for  l = 0; l < nVars; ++l  {
        pNormalArrTime[l] = pArrTimeProfile[l];
    }

    nDelta = Abc_NormalizeArrivalTimes( pNormalArrTime, nVars, &nMaxArrival );

    *Cost = ABC_INFINITY;

    if Ses_StoreGetEntry( s_pSesStore, pTruth, nVars, pNormalArrTime, &pSol ) != 0 {
        s_pSesStore.nCacheHits++;
        s_pSesStore.pCacheHits[nVars]++;
    
    } else {
        if s_pSesStore.fVeryVerbose {
            spFmt.Printf(ANSI_COLOR_CYAN);
            Abc_TtPrintHexRev( os.Stdout, pTruth, nVars );
            spFmt.Printf(ANSI_COLOR_RESET);
            spFmt.Printf(" [%d", pNormalArrTime[0]);
            for  l = 1; l < nVars; ++l  {
                spFmt.Printf(" %d", pNormalArrTime[l]);
            }
            spFmt.Printf("]@%d:", AigLevel);
            fflush( os.Stdout );
        }

        nMaxDepth = pNormalArrTime[0];
        for  i = 1; i < nVars; ++i  {
            nMaxDepth = Abc_MaxInt( nMaxDepth, pNormalArrTime[i] );
        }
        nMaxDepth += nVars + 1;
        if AigLevel != -1 {
            nMaxDepth = Abc_MinInt( AigLevel - nDelta, nMaxDepth + nVars + 1 ); }

        timeStartExact = Abc_Clock();

        pSes = Ses_ManAlloc( pTruth, nVars, 1 /* nSpecFunc */, nMaxDepth, pNormalArrTime, s_pSesStore.fMakeAIG, s_pSesStore.nBTLimit, s_pSesStore.fVerbose );
        pSes.fVeryVerbose = s_pSesStore.fVeryVerbose;
        pSes.pSat = s_pSesStore.pSat;
        pSes.nStartGates = nVars - 2;

        for  pSes.nMaxDepth  { /* there is improvement */
        
            if s_pSesStore.fVeryVerbose {
                spFmt.Printf(" %d", pSes.nMaxDepth);
                fflush( os.Stdout );
            }

            if ( pSol2 = Ses_ManFindMinimumSize( pSes ) ) != nil {
                if s_pSesStore.fVeryVerbose {
                    if pSes.nMaxDepth >= 10 { spFmt.Printf("\b"); }
                    spFmt.Printf("\b" ANSI_COLOR_GREEN "%d" ANSI_COLOR_RESET, pSes.nMaxDepth);
                }
                if pSol {
                    // inlining of macro ABC_FREE
                    ((pSol) ? (free((char *) (pSol)), (pSol) = 0) : 0); }
                pSol = pSol2;
                pSes.nMaxDepth--;
            
            } else {
                if s_pSesStore.fVeryVerbose {
                    if pSes.nMaxDepth >= 10 { spFmt.Printf("\b"); }
                    spFmt.Printf("\b%s%d" ANSI_COLOR_RESET, pSes.fHitResLimit ? ANSI_COLOR_RED : ANSI_COLOR_YELLOW, pSes.nMaxDepth);
                }
                break;
            }
        }

        if s_pSesStore.fVeryVerbose {
            spFmt.Printf("        \n"); }

        /* log unsuccessful case for debugging */
        if s_pSesStore.pDebugEntries && pSes.fHitResLimit {
            Ses_StorePrintDebugEntry( s_pSesStore, pTruth, nVars, pNormalArrTime, pSes.nMaxDepth, pSol, nVars - 2 ); }

        pSes.timeTotal = Abc_Clock() - timeStartExact;

        /* statistics */
        s_pSesStore.nSatCalls += pSes.nSatCalls;
        s_pSesStore.nUnsatCalls += pSes.nUnsatCalls;
        s_pSesStore.nUndefCalls += pSes.nUndefCalls;

        s_pSesStore.timeSat += pSes.timeSat;
        s_pSesStore.timeSatSat += pSes.timeSatSat;
        s_pSesStore.timeSatUnsat += pSes.timeSatUnsat;
        s_pSesStore.timeSatUndef += pSes.timeSatUndef;
        s_pSesStore.timeInstance += pSes.timeInstance;
        s_pSesStore.timeExact += pSes.timeTotal;

        /* cleanup (we need to clean before adding since pTruth may have been modified by pSes) */
        fResLimit = pSes.fHitResLimit;
        Ses_ManCleanLight( pSes );

        /* store solution */
        Ses_StoreAddEntry( s_pSesStore, pTruth, nVars, pNormalArrTime, pSol, fResLimit );
    }

    if pSol {
        *Cost = pSol[ABC_EXACT_SOL_NGATES];
        p = pSol + 3 + 4 * pSol[ABC_EXACT_SOL_NGATES] + 1;
        Delay = *p++;
        for  l = 0; l < nVars; ++l  {
            pPerm[l] = *p++;
        }
    }

    if pSol {
        var Delay2 int32;
        for  l = 0; l < nVars; ++l  {
            //printf( "%d ", pPerm[l] );
            Delay2 = Abc_MaxInt( Delay2, pArrTimeProfile[l] + pPerm[l] );
        }
        //printf( "  output arrival = %d    recomputed = %d\n", Delay, Delay2 );
        //if ( Delay != Delay2 )
        //{
        //    printf( "^--- BUG!\n" );
        //    assert( 0 );
        //}

        s_pSesStore.timeTotal += ( Abc_Clock() - timeStart );
        return Delay2;
    
    } else {
        assert(*Cost == ABC_INFINITY, "*Cost == ABC_INFINITY");

        s_pSesStore.timeTotal += ( Abc_Clock() - timeStart );
        return ABC_INFINITY;
    }
}
// this procedure returns a new node whose output in terms of the given fanins
// has the smallest possible arrival time (in agreement with the above Abc_ExactDelayCost)
func Abc_ExactBuildNode(pTruth *word, nVars int32, pArrTimeProfile *int32, pFanins **Abc_Obj_t, pNtk *Abc_Ntk_t) *Abc_Obj_t {
    var pSol string  = nil;
    var i, j, nMaxArrival int32;
    var pNormalArrTime [8]int32;
    char * p;
    var pObj *Abc_Obj_t;
    var pGates *Vec_Ptr_t;
    var pGateTruth [5]byte;
    var pSopCover string;
    var timeStart abctime  = Abc_Clock();

    if nVars == 0 {
        s_pSesStore.timeTotal += ( Abc_Clock() - timeStart );
        if pTruth[0] & 1 {
            return Abc_NtkCreateNodeConst1(pNtk);
        } else {
            return Abc_NtkCreateNodeConst0(pNtk);
        }
    }
    if nVars == 1 {
        s_pSesStore.timeTotal += ( Abc_Clock() - timeStart );
        if pTruth[0] & 1 {
            return Abc_NtkCreateNodeInv(pNtk, pFanins[0]);
        } else {
            return Abc_NtkCreateNodeBuf(pNtk, pFanins[0]);
        }
    }

    for  i = 0; i < nVars; ++i  {
        pNormalArrTime[i] = pArrTimeProfile[i];
    }
    Abc_NormalizeArrivalTimes( pNormalArrTime, nVars, &nMaxArrival );
    assert((Ses_StoreGetEntry( s_pSesStore, pTruth, nVars, pNormalArrTime, &pSol ) != 0), "Ses_StoreGetEntry( s_pSesStore, pTruth, nVars, pNormalArrTime, &pSol )");
    if !pSol {
        s_pSesStore.timeTotal += ( Abc_Clock() - timeStart );
        return nil;
    }

    assert(pSol[ABC_EXACT_SOL_NVARS] == nVars, "pSol[ABC_EXACT_SOL_NVARS] == nVars");
    assert(pSol[ABC_EXACT_SOL_NFUNC] == 1, "pSol[ABC_EXACT_SOL_NFUNC] == 1");

    pGates = Vec_PtrAlloc( nVars + pSol[ABC_EXACT_SOL_NGATES] );
    pGateTruth[3] = '0';
    pGateTruth[4] = '\0';

    /* primary inputs */
    for  i = 0; i < nVars; ++i  {
        assert(pFanins[i], "pFanins[i]");
        Vec_PtrPush( pGates, pFanins[i] );
    }

    /* gates */
    p = pSol + 3;
    for  i = 0; i < pSol[ABC_EXACT_SOL_NGATES]; ++i  {
        pGateTruth[2] = '0' + ( *p & 1 );
        pGateTruth[1] = '0' + ( ( *p >> 1 ) & 1 );
        pGateTruth[0] = '0' + ( ( *p >> 2 ) & 1 );
        ++p;

        assert(*p == 2, "*p == 2"); /* binary gate */
        ++p;

        /* invert truth table if we are last gate and inverted */
        if i + 1 == pSol[ABC_EXACT_SOL_NGATES] && Abc_LitIsCompl( *( p + 2 ) ) {
            for  j = 0; j < 4; ++j  {
                pGateTruth[j] = ( pGateTruth[j] == '0' ) ? '1' : '0';
            } }

        pSopCover = Abc_SopFromTruthBin( pGateTruth );
        pObj = Abc_NtkCreateNode( pNtk );
        assert((pObj != nil), "pObj");
        pObj.pData = Abc_SopRegister( (Mem_Flex_t*)pNtk.pManFunc, pSopCover );
        Vec_PtrPush( pGates, pObj );
        // inlining of macro ABC_FREE
        ((pSopCover) ? (free((char *) (pSopCover)), (pSopCover) = 0) : 0);

        Abc_ObjAddFanin( pObj, (Abc_Obj_t *)Vec_PtrEntry( pGates, *p++ ) );
        Abc_ObjAddFanin( pObj, (Abc_Obj_t *)Vec_PtrEntry( pGates, *p++ ) );
    }

    /* output */
    pObj = (Abc_Obj_t *)Vec_PtrEntry( pGates, nVars + Abc_Lit2Var( *p ) );

    Vec_PtrFree( pGates );

    s_pSesStore.timeTotal += ( Abc_Clock() - timeStart );
    return pObj;
}

func Abc_ExactStoreTest(fVerbose int32) {
    var i int32;
    var pTruth [4]word = [4]word{0xcafe, 0, 0, 0};
    var pArrTimeProfile [4]int32 = [4]int32{6, 2, 8, 5};
    var pNtk *Abc_Ntk_t;
    var pFanins [4]*Abc_Obj_t;
    var vNames *Vec_Ptr_t;
    var pPerm [4]byte ;
    var Cost int32;

    pNtk = Abc_NtkAlloc( ABC_NTK_LOGIC, ABC_FUNC_SOP, 1 );
    pNtk.pName = Extra_UtilStrsav( "exact" );
    vNames = Abc_NodeGetFakeNames( uint32(4) );

    /* primary inputs */
    Vec_PtrPush( pNtk.vObjs, nil );
    for  i = 0; i < 4; ++i  {
        pFanins[i] = Abc_NtkCreatePi( pNtk );
        Abc_ObjAssignName( pFanins[i], (char*)Vec_PtrEntry( vNames, i ), nil );
    }
    Abc_NodeFreeNames( vNames );

    Abc_ExactStart( 10000, 1, fVerbose, 0, nil );

    assert(!(Abc_ExactBuildNode( pTruth, 4, pArrTimeProfile, pFanins, pNtk ) != nil), "!Abc_ExactBuildNode( pTruth, 4, pArrTimeProfile, pFanins, pNtk )");

    assert(Abc_ExactDelayCost( pTruth, 4, pArrTimeProfile, pPerm, &Cost, 12 ) == 1, "Abc_ExactDelayCost( pTruth, 4, pArrTimeProfile, pPerm, &Cost, 12 ) == 1");

    assert((Abc_ExactBuildNode( pTruth, 4, pArrTimeProfile, pFanins, pNtk ) != nil), "Abc_ExactBuildNode( pTruth, 4, pArrTimeProfile, pFanins, pNtk )");

    (*pArrTimeProfile)++;
    assert(!(Abc_ExactBuildNode( pTruth, 4, pArrTimeProfile, pFanins, pNtk ) != nil), "!Abc_ExactBuildNode( pTruth, 4, pArrTimeProfile, pFanins, pNtk )");
    (*pArrTimeProfile)--;

    Abc_ExactStop( nil );

    Abc_NtkDelete( pNtk );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

package abci

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcNpn.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures for testing and comparing semi-canonical forms.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcNpn.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "misc/extra/extra.h"
//#include "misc/vec/vec.h"

//#include "bool/kit/kit.h"
//#include "bool/lucky/lucky.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
// semi-canonical form types
// 0 - none
// 1 - based on counting 1s in cofactors
// 2 - based on minimum truth table value
// 3 - exact NPN

// data-structure to store a bunch of truth tables
    
type Abc_TtStore_t struct {
    nVars                int32;
    nWords                int32;
    nFuncs                int32;
    pFuncs             **word;
}

// extern Abc_TtStore_t * Abc_TtStoreLoad( char * pFileName, int nVarNum );

// extern void            Abc_TtStoreFree( Abc_TtStore_t * p, int nVarNum );

// extern void            Abc_TtStoreWrite( char * pFileName, Abc_TtStore_t * p, int fBinary );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Counts the number of unique truth tables.]

  Description []
               
  SideEffects [] 

  SeeAlso     []

***********************************************************************/
// returns hash key of the truth table
var spStatic_Abc_TruthHashKey_s_BigPrimes [7]uint32 = [7]uint32{12582917, 25165843, 50331653, 100663319, 201326611, 402653189, 805306457};
func Abc_TruthHashKey(pFunc *word, nWords int32, nTableSize int32) int32 {
    
    var w int32;
    var Key word;
    for  w = 0; w < nWords; w++  {
        Key += pFunc[w] * spStatic_Abc_TruthHashKey_s_BigPrimes[w % 7];
    }
    return int32((Key % nTableSize));
}
// returns 1 if the entry with this truth table exits
func Abc_TruthHashLookup(pFuncs **word, iThis int32, nWords int32, pTable *int32, pNexts *int32, Key int32) int32 {
    var iThat int32;
    for  iThat = pTable[Key]; iThat != -1; iThat = pNexts[iThat]  {
        if !memcmp( pFuncs[iThat], pFuncs[iThis], sizeof(word) * nWords ) {
            return 1; }
    }
    return 0;
}
// hashes truth tables and collects unique ones
func Abc_TruthNpnCountUnique(p *Abc_TtStore_t) int32 {
    // allocate hash table
    var nTableSize int32  = Abc_PrimeCudd(p.nFuncs);
    var pTable *int32  = ABC_FALLOC( int, nTableSize );
    var pNexts *int32  = ABC_FALLOC( int, nTableSize );
    // hash functions
    var i, k, Key int32;
    for  i = 0; i < p.nFuncs; i++  {
        Key = Abc_TruthHashKey( p.pFuncs[i], p.nWords, nTableSize );
        if Abc_TruthHashLookup( p.pFuncs, i, p.nWords, pTable, pNexts, Key ) != 0 { // found equal
            p.pFuncs[i] = nil;
        } else { // there is no equal (the first time this one occurs so far)
            pNexts[i] = pTable[Key], pTable[Key] = i; }
    }
    ABC_FREE( pTable );
    ABC_FREE( pNexts );
    // count the number of unqiue functions
    assert(p.pFuncs[0] != nil, "p.pFuncs[0] != nil");
    for  i = k = 1; i < p.nFuncs; i++  {
        if p.pFuncs[i] != nil {
            p.pFuncs[k++] = p.pFuncs[i]; }
    }
    return (p.nFuncs = k);
}

/**Function*************************************************************

  Synopsis    [Counts the number of unique truth tables.]

  Description []
               
  SideEffects [] 

  SeeAlso     []

***********************************************************************/
var nWords int32; // unfortunate global variable
func Abc_TruthCompare(p1 **word, p2 **word) int32 { return memcmp(*p1, *p2, sizeof(word) * nWords); }
func Abc_TruthNpnCountUniqueSort(p *Abc_TtStore_t) int32 {
    var i, k int32;
    // sort them by value
    nWords = p.nWords;
    assert(nWords > 0, "nWords > 0");
    qsort( (void *)p.pFuncs, (size_t)p.nFuncs, sizeof(word *), (int(*)(void *,void *))Abc_TruthCompare );
    // count the number of unqiue functions
    for  i = k = 1; i < p.nFuncs; i++  {
        if memcmp( p.pFuncs[i-1], p.pFuncs[i], sizeof(word) * nWords ) {
            p.pFuncs[k++] = p.pFuncs[i]; }
    }
    return (p.nFuncs = k);
}

/**Function*************************************************************

  Synopsis    [Prints out one NPN transform.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TruthNpnPrint(pCanonPermInit string, uCanonPhase uint32, nVars int32) {
    var pCanonPerm [16]byte; var i int32;
    assert(nVars <= 16, "nVars <= 16");
    for  i = 0; i < nVars; i++  {
        pCanonPerm[i] = pCanonPermInit ? pCanonPermInit[i] : 'a' + i;
    }
    fmt.Printf("   %c = ( ", Abc_InfoHasBit(&uCanonPhase, nVars) ? 'Z':'z');
    for  i = 0; i < nVars; i++  {
        fmt.Printf("%c%s", pCanonPerm[i] + ('A'-'a') * Abc_InfoHasBit(&uCanonPhase, pCanonPerm[i]-'a'), i == nVars-1 ? "":",");
    }
    fmt.Printf(" )  ");
}

/**Function*************************************************************

  Synopsis    [Apply decomposition to the truth table.]

  Description [Returns the number of AIG nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TruthNpnPerform(p *Abc_TtStore_t, NpnType int32, fVerbose int32) {
    var pAux [2048]uint32;
    var pAuxWord, pAuxWord1 [1024]word;
    var pCanonPerm [16]byte;
    var uCanonPhase uint32;
    var clk abctime  = Abc_Clock();
    var i int32;

    var pAlgoName string  = nil;
    if NpnType == 0 {
        pAlgoName = "uniqifying          ";
    } else if NpnType == 1 {
        pAlgoName = "exact NPN           ";
    } else if NpnType == 2 {
        pAlgoName = "counting 1s         ";
    } else if NpnType == 3 {
        pAlgoName = "Jake's hybrid fast  ";
    } else if NpnType == 4 {
        pAlgoName = "Jake's hybrid good  ";
    } else if NpnType == 5 {
        pAlgoName = "new hybrid fast     ";
    } else if NpnType == 6 {
        pAlgoName = "new phase flipping  ";
    } else if NpnType == 7 {
        pAlgoName = "new hier. matching  ";
    } else if NpnType == 8 {
        pAlgoName = "new adap. matching  ";
    } else if NpnType == 9 {
        pAlgoName = "adjustable algorithm (heuristic) ";
    } else if NpnType == 10 {
        pAlgoName = "adjustable algorithm (exact)     ";
    } else if NpnType == 11 {
        pAlgoName = "new cost-aware exact algorithm   "; }

    assert(p.nVars <= 16, "p.nVars <= 16");
    if pAlgoName {
        fmt.Printf("Applying %-20s to %8d func%s of %2d vars...  ", pAlgoName, p.nFuncs, (p.nFuncs == 1 ? "":"s"), p.nVars); }
    if fVerbose != 0 {
        fmt.Printf("\n"); }

    if NpnType == 0 {
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d : ", i); }
            if fVerbose != 0 {
                Extra_PrintHex( os.Stdout, (unsigned *)p.pFuncs[i], p.nVars );
fmt.Printf("\n"); }
        }
    
    } else if NpnType == 1 {
        var pi *permInfo; 
        Abc_TruthNpnCountUnique(p);
        pi = setPermInfoPtr(p.nVars);
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d : ", i); }
            simpleMinimal(p.pFuncs[i], pAuxWord, pAuxWord1, pi, p.nVars);
            if fVerbose != 0 {
                Extra_PrintHex( os.Stdout, (unsigned *)p.pFuncs[i], p.nVars );
Abc_TruthNpnPrint(pCanonPerm, uCanonPhase, p.nVars);
fmt.Printf("\n"); }
        }
        freePermInfoPtr(pi);
    
    } else if NpnType == 2 {
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d : ", i); }
            resetPCanonPermArray(pCanonPerm, p.nVars);
            uCanonPhase = Kit_TruthSemiCanonicize( (unsigned *)p.pFuncs[i], pAux, p.nVars, pCanonPerm );
            if fVerbose != 0 {
                Extra_PrintHex( os.Stdout, (unsigned *)p.pFuncs[i], p.nVars );
Abc_TruthNpnPrint(pCanonPerm, uCanonPhase, p.nVars);
fmt.Printf("\n"); }
        }
    
    } else if NpnType == 3 {
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d : ", i); }
            resetPCanonPermArray(pCanonPerm, p.nVars);
            uCanonPhase = luckyCanonicizer_final_fast( p.pFuncs[i], p.nVars, pCanonPerm );
            if fVerbose != 0 {
                Extra_PrintHex( os.Stdout, (unsigned *)p.pFuncs[i], p.nVars );
Abc_TruthNpnPrint(pCanonPerm, uCanonPhase, p.nVars);
fmt.Printf("\n"); }
        }
    
    } else if NpnType == 4 {
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d : ", i); }
            resetPCanonPermArray(pCanonPerm, p.nVars);
            uCanonPhase = luckyCanonicizer_final_fast1( p.pFuncs[i], p.nVars, pCanonPerm );
            if fVerbose != 0 {
                Extra_PrintHex( os.Stdout, (unsigned *)p.pFuncs[i], p.nVars );
Abc_TruthNpnPrint(pCanonPerm, uCanonPhase, p.nVars);
fmt.Printf("\n"); }
        }
    
    } else if NpnType == 5 {
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d : ", i); }
            uCanonPhase = Abc_TtCanonicize( p.pFuncs[i], p.nVars, pCanonPerm );
            if fVerbose != 0 {
                Extra_PrintHex( os.Stdout, (unsigned *)p.pFuncs[i], p.nVars );
Abc_TruthNpnPrint(pCanonPerm, uCanonPhase, p.nVars);
fmt.Printf("\n"); }
        }
    
    } else if NpnType == 6 {
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d : ", i); }
            uCanonPhase = Abc_TtCanonicizePhase( p.pFuncs[i], p.nVars );
            if fVerbose != 0 {
                Extra_PrintHex( os.Stdout, (unsigned *)p.pFuncs[i], p.nVars );
Abc_TruthNpnPrint(nil, uCanonPhase, p.nVars);
fmt.Printf("\n"); }
        }
    
    } else if NpnType == 7 {
        var fExact int32;
		var pMan *Abc_TtHieMan_t  = Abc_TtHieManStart( p.nVars, 5 );
        for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d : ", i); }
            uCanonPhase = Abc_TtCanonicizeHie( pMan, p.pFuncs[i], p.nVars, pCanonPerm, fExact );
            if fVerbose != 0 {
//                Extra_PrintHex( stdout, (unsigned *)p->pFuncs[i], p->nVars ), Abc_TruthNpnPrint(NULL, uCanonPhase, p->nVars), printf( "\n" );
                fmt.Printf("\n"); }
        }
        // nClasses = Abc_TtManNumClasses( pMan );
		Abc_TtHieManStop( pMan );
    
    } else if NpnType == 8 {
//        typedef unsigned(*TtCanonicizeFunc)(Abc_TtHieMan_t * p, word * pTruth, int nVars, char * pCanonPerm, int flag);
        unsigned Abc_TtCanonicizeWrap(TtCanonicizeFunc func, Abc_TtHieMan_t * p, word * pTruth, int nVars, char * pCanonPerm, int flag);
        unsigned Abc_TtCanonicizeAda(Abc_TtHieMan_t * p, word * pTruth, int nVars, char * pCanonPerm, int iThres);
		
		var fHigh int32 = 1;
		var iEnumThres int32 = 25;
		var pMan *Abc_TtHieMan_t  = Abc_TtHieManStart(p.nVars, 5);
		for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d : ", i); }
            uCanonPhase = Abc_TtCanonicizeWrap(Abc_TtCanonicizeAda, pMan, p.pFuncs[i], p.nVars, pCanonPerm, fHigh*100 + iEnumThres);
            if fVerbose != 0 {
                Extra_PrintHex( os.Stdout, (unsigned *)p.pFuncs[i], p.nVars );
Abc_TruthNpnPrint(pCanonPerm, uCanonPhase, p.nVars);
fmt.Printf("\n"); }
        }
		Abc_TtHieManStop(pMan);
    
    } else if NpnType == 9 || NpnType == 10 || NpnType == 11 {
//        typedef unsigned(*TtCanonicizeFunc)(Abc_TtHieMan_t * p, word * pTruth, int nVars, char * pCanonPerm, int flag);
        unsigned Abc_TtCanonicizeWrap(TtCanonicizeFunc func, Abc_TtHieMan_t * p, word * pTruth, int nVars, char * pCanonPerm, int flag);
        unsigned Abc_TtCanonicizeAda(Abc_TtHieMan_t * p, word * pTruth, int nVars, char * pCanonPerm, int iThres);
        unsigned Abc_TtCanonicizeCA(Abc_TtHieMan_t * p, word * pTruth, int nVars, char * pCanonPerm, int iThres);
		
		var pMan *Abc_TtHieMan_t  = Abc_TtHieManStart(p.nVars, 5);
		for  i = 0; i < p.nFuncs; i++  {
            if fVerbose != 0 {
                fmt.Printf("%7d : ", i); }
            if NpnType == 9 {
                uCanonPhase = Abc_TtCanonicizeWrap(Abc_TtCanonicizeAda, pMan, p.pFuncs[i], p.nVars, pCanonPerm,  125); // -A 8, adjustable algorithm (heuristic)
            } else if NpnType == 10 {
                uCanonPhase = Abc_TtCanonicizeWrap(Abc_TtCanonicizeAda, pMan, p.pFuncs[i], p.nVars, pCanonPerm, 1199); // -A 9, adjustable algorithm (exact)
            } else if NpnType == 11 {
                uCanonPhase = Abc_TtCanonicizeWrap(Abc_TtCanonicizeCA,  pMan, p.pFuncs[i], p.nVars, pCanonPerm,    1); } // -A 10, new cost-aware exact algorithm
            if fVerbose != 0 {
                Extra_PrintHex( os.Stdout, (unsigned *)p.pFuncs[i], p.nVars );
Abc_TruthNpnPrint(pCanonPerm, uCanonPhase, p.nVars);
fmt.Printf("\n"); }
        }
		Abc_TtHieManStop(pMan);
    
    } else { assert(false, "0"); }
    clk = Abc_Clock() - clk;
    fmt.Printf("Classes =%9d  ", Abc_TruthNpnCountUnique(p));
    Abc_PrintTime( 1, "Time", clk );
}

/**Function*************************************************************

  Synopsis    [Apply decomposition to truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TruthNpnTest(pFileName string, NpnType int32, nVarNum int32, fDumpRes int32, fBinary int32, fVerbose int32) {
    var p *Abc_TtStore_t;
    var pFileNameOut string;

    // read info from file
    p = Abc_TtStoreLoad( pFileName, nVarNum );
    if p == nil {
        return; }

    // consider functions from the file
    Abc_TruthNpnPerform( p, NpnType, fVerbose );

    // write the result
    if fDumpRes != 0 {
        if fBinary != 0 {
            pFileNameOut = Extra_FileNameGenericAppend( pFileName, "_out.tt" );
        } else {
            pFileNameOut = Extra_FileNameGenericAppend( pFileName, "_out.txt" ); }
        Abc_TtStoreWrite( pFileNameOut, p, fBinary );
        if fVerbose != 0 {
            fmt.Printf("The resulting functions are written into file \"%s\".\n", pFileNameOut); }
    }

    // delete data-structure
    Abc_TtStoreFree( p, nVarNum );
//    printf( "Finished computing canonical forms for functions from file \"%s\".\n", pFileName );
}


/**Function*************************************************************

  Synopsis    [Testbench for decomposition algorithms.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NpnTest(pFileName string, NpnType int32, nVarNum int32, fDumpRes int32, fBinary int32, fVerbose int32) int32 {
    if fVerbose != 0 {
        fmt.Printf("Using truth tables from file \"%s\"...\n", pFileName); }
    if NpnType >= 0 && NpnType <= 11 {
        Abc_TruthNpnTest( pFileName, NpnType, nVarNum, fDumpRes, fBinary, fVerbose );
    } else {
        fmt.Printf("Unknown canonical form value (%d).\n", NpnType); }
    fflush( os.Stdout );
    return 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


package super

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [superAnd.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Pre-computation of supergates.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 8, 2003.]

  Revision    [$Id: superAnd.c,v 1.3 2004/06/28 14:20:25 alanmi Exp $]

***********************************************************************/

//#include "superInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the bit masks
// Original macro:
// #define SUPER_MASK(n)     ((~((unsigned)0)) >> (32-n))

#define SUPER_FULL         (~((unsigned)0))

// data structure for AND2 subgraph precomputation
type Super2_Man_t =     Super2_ManStruct_t_;   // manager
type Super2_Lib_t =     Super2_LibStruct_t_;   // library
type Super2_Gate_t =    Super2_GateStruct_t_;  // supergate

type Super2_ManStruct_t_ struct {
    pMem    *Extra_MmFixed_t;         // memory manager for all supergates
    tTable         *stmm_table;       // mapping of truth tables into gates
    nTried                 int32;       // the total number of tried
}

type Super2_LibStruct_t_ struct {
    i                 int32;            // used to iterate through the table
    k                 int32;            // used to iterate through the table
    nInputs                 int32;      // the number of inputs
    nMints                 int32;       // the number of minterms
    nLevels                 int32;      // the number of logic levels
    nGates                 int32;       // the number of gates in the library
    nGatesAlloc                 int32;  // the number of allocated places
    pGates     **Super2_Gate_t;       // the gates themselves
    uMaskBit            uint32;     // the mask used to determine the compl bit
}

type Super2_GateStruct_t_ struct {
    uTruth            uint32;       // the truth table of this supergate
    pOne      *Super2_Gate_t;         // the left wing
    pTwo      *Super2_Gate_t;         // the right wing
    pNext      *Super2_Gate_t;        // the next gate in the table
}


// manipulation of complemented attributes
// Original macro:
// #define Super2_IsComplement(p)    (((int)((ABC_PTRUINT_T) (p) & 01)))

// Original macro:
// #define Super2_Regular(p)         ((Super2_Gate_t *)((ABC_PTRUINT_T)(p) & ~01)) 

// Original macro:
// #define Super2_Not(p)             ((Super2_Gate_t *)((ABC_PTRUINT_T)(p) ^ 01)) 

// Original macro:
// #define Super2_NotCond(p,c)       ((Super2_Gate_t *)((ABC_PTRUINT_T)(p) ^ (c)))


// iterating through the gates in the library
// Original macro:
// #define Super2_LibForEachGate( Lib, Gate )                        \
//     for ( Lib->i = 0;                                             \
//           Lib->i < Lib->nGates && (Gate = Lib->pGates[Lib->i]);   \
//           Lib->i++ )

// Original macro:
// #define Super2_LibForEachGate2( Lib, Gate2 )                      \
//     for ( Lib->k = 0;                                             \
//           Lib->k < Lib->i && (Gate2 = Lib->pGates[Lib->k]);       \
//           Lib->k++ )


// static functions
// static Super2_Man_t * Super2_ManStart();

// static void           Super2_ManStop( Super2_Man_t * pMan );

// static Super2_Lib_t * Super2_LibStart();

// static Super2_Lib_t * Super2_LibDup( Super2_Lib_t * pLib );

// static void           Super2_LibStop( Super2_Lib_t * pLib );

// static void           Super2_LibAddGate( Super2_Lib_t * pLib, Super2_Gate_t * pGate );

// static Super2_Lib_t * Super2_LibFirst( Super2_Man_t * pMan, int nInputs );

// static Super2_Lib_t * Super2_LibCompute( Super2_Man_t * pMan, Super2_Lib_t * pLib );


// static void           Super2_LibWrite( Super2_Lib_t * pLib );

// static void           Super2_LibWriteGate( FILE * pFile, Super2_Lib_t * pLib, Super2_Gate_t * pGate );

// static char *         Super2_LibWriteGate_rec( Super2_Gate_t * pGate, int fInv, int Level );

// static int            Super2_LibWriteCompare( char * pStr1, char * pStr2 );

// static int            Super2_LibCompareGates( Super2_Gate_t ** ppG1, Super2_Gate_t ** ppG2 );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Precomputes the library of AND2 gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_Precompute(nInputs int32, nLevels int32, fVerbose int32) {
    var pMan *Super2_Man_t;
    var pLibCur, pLibNext *Super2_Lib_t;
    var Level int32;
    var clk abctime;

    assert(nInputs < 6, "nInputs < 6");

    // start the manager
    pMan = Super2_ManStart();

    // get the starting supergates
    pLibCur = Super2_LibFirst( pMan, nInputs );

    // perform the computation of supergates
fmt.Printf("Computing supergates for %d inputs and %d levels:\n", nInputs, nLevels);
    for  Level = 1; Level <= nLevels; Level++  {
clk = Abc_Clock();
        pLibNext = Super2_LibCompute( pMan, pLibCur );
        pLibNext.nLevels = Level;
        Super2_LibStop( pLibCur );
        pLibCur = pLibNext;
fmt.Printf("Level %d:  Tried = %7d.  Computed = %7d.  ", Level, pMan.nTried, pLibCur.nGates);
ABC_PRT( "Runtime", Abc_Clock() - clk );
fflush( os.Stdout );
    }

fmt.Printf("Writing the output file...\n");
fflush( os.Stdout );
    // write them into a file
    Super2_LibWrite( pLibCur );
    Super2_LibStop( pLibCur );

    // stop the manager
    Super2_ManStop( pMan );
}




/**Function*************************************************************

  Synopsis    [Starts the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_ManStart() *Super2_Man_t {
    var pMan *Super2_Man_t;
    pMan = new(Super2_Man_t);
    *pMan = Super2_Man_t{};
    pMan.pMem   = Extra_MmFixedStart( sizeof(Super2_Gate_t) );
    pMan.tTable = stmm_init_table( st__ptrcmp, st__ptrhash );
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Stops the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_ManStop(pMan *Super2_Man_t) {
    Extra_MmFixedStop( pMan.pMem );
    stmm_free_table( pMan.tTable );
    ABC_FREE( pMan );
}

/**Function*************************************************************

  Synopsis    [Starts the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_LibStart() *Super2_Lib_t {
    var pLib *Super2_Lib_t;
    pLib = new(Super2_Lib_t);
    *pLib = Super2_Lib_t{};
    return pLib;
}

/**Function*************************************************************

  Synopsis    [Duplicates the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_LibDup(pLib *Super2_Lib_t) *Super2_Lib_t {
    var pLibNew *Super2_Lib_t;
    pLibNew = Super2_LibStart();
    pLibNew.nInputs     = pLib.nInputs;
    pLibNew.nMints      = pLib.nMints;
    pLibNew.nLevels     = pLib.nLevels;
    pLibNew.nGates      = pLib.nGates;
    pLibNew.uMaskBit    = pLib.uMaskBit;
    pLibNew.nGatesAlloc = 1000 + pLib.nGatesAlloc;
    pLibNew.pGates      = ABC_ALLOC( Super2_Gate_t *, pLibNew.nGatesAlloc );
    memcpy( pLibNew.pGates, pLib.pGates, pLibNew.nGates * sizeof(Super2_Gate_t *) );
    return pLibNew;
}

/**Function*************************************************************

  Synopsis    [Add gate to the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_LibAddGate(pLib *Super2_Lib_t, pGate *Super2_Gate_t) {
    if pLib.nGates == pLib.nGatesAlloc {
        pLib.pGates  = ABC_REALLOC( Super2_Gate_t *, pLib.pGates,  3 * pLib.nGatesAlloc );
        pLib.nGatesAlloc *= 3;
    }
    pLib.pGates[ pLib.nGates++ ] = pGate;
}

/**Function*************************************************************

  Synopsis    [Stops the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_LibStop(pLib *Super2_Lib_t) {
    ABC_FREE( pLib.pGates );
    ABC_FREE( pLib );
}

/**Function*************************************************************

  Synopsis    [Derives the starting supergates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_LibFirst(pMan *Super2_Man_t, nInputs int32) *Super2_Lib_t {
    var pLib *Super2_Lib_t;
    var v, m int32;    

    // start the library
    pLib = Super2_LibStart();

    // create the starting supergates
    pLib.nInputs     = nInputs;
    pLib.nMints      = (1 << nInputs);
    pLib.nLevels     = 0;
    pLib.nGates      = nInputs + 1;
    pLib.nGatesAlloc = nInputs + 1;
    pLib.uMaskBit    = (1 << (pLib.nMints-1));
    pLib.pGates      = ABC_ALLOC( Super2_Gate_t *, nInputs + 1 );
    // add the constant 0
    pLib.pGates[0] = (Super2_Gate_t *)Extra_MmFixedEntryFetch( pMan.pMem );
    memset( pLib.pGates[0], 0, sizeof(Super2_Gate_t) );
    // add the elementary gates
    for  v = 0; v < nInputs; v++  {
        pLib.pGates[v+1] = (Super2_Gate_t *)Extra_MmFixedEntryFetch( pMan.pMem );
        memset( pLib.pGates[v+1], 0, sizeof(Super2_Gate_t) );
        pLib.pGates[v+1].pTwo = (Super2_Gate_t *)(ABC_PTRUINT_T)v;
    }

    // set up their truth tables
    for  m = 0; m < pLib.nMints; m++  {
        for  v = 0; v < nInputs; v++  {
            if m & (1 << v) {
                pLib.pGates[v+1].uTruth |= (1 << m); }
        }
    }
    return pLib;
}

/**Function*************************************************************

  Synopsis    [Precomputes one level of supergates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_LibCompute(pMan *Super2_Man_t, pLib *Super2_Lib_t) *Super2_Lib_t {
    var pLibNew *Super2_Lib_t;
    var pGate1, pGate2, pGateNew *Super2_Gate_t;
    var ppGate **Super2_Gate_t;
    // inlining of macro SUPER_MASK
    var Mask uint32  = ((~(uint32(0))) >> (32-pLib.nMints));
    var uTruth, uTruthR, uTruth1, uTruth2, uTruth1c, uTruth2c uint32;

    // start the new library
    pLibNew = Super2_LibDup( pLib );

    // reset the hash table
    stmm_free_table( pMan.tTable );
    pMan.tTable = stmm_init_table( st__ptrcmp, st__ptrhash );
    // set the starting things into the hash table
    // inlining of macro Super2_LibForEachGate
    for  pLibNew.i = 0;                                                       pLibNew.i < pLibNew.nGates && (pGate1 = pLibNew.pGates[pLibNew.i]);             pLibNew.i++  {
        if pGate1.uTruth & pLibNew.uMaskBit {
            uTruthR = Mask & ~pGate1.uTruth;
        } else {
            uTruthR = pGate1.uTruth;
        }

        if stmm_lookup( pMan.tTable, (char *)(ABC_PTRUINT_T)uTruthR, (char **)&pGate2 ) {
            fmt.Printf("New gate:\n");
            Super2_LibWriteGate( os.Stdout, pLibNew, pGate1 );
            fmt.Printf("Gate in the table:\n");
            Super2_LibWriteGate( os.Stdout, pLibNew, pGate2 );
            assert(false, "0");
        }
        stmm_insert( pMan.tTable, (char *)(ABC_PTRUINT_T)uTruthR, (char *)(ABC_PTRUINT_T)pGate1 );
    }


    // set the number of gates tried
    pMan.nTried = pLibNew.nGates;

    // go through the gate pairs
    // inlining of macro Super2_LibForEachGate
    for  pLib.i = 0;                                                       pLib.i < pLib.nGates && (pGate1 = pLib.pGates[pLib.i]);             pLib.i++  {
        if pLib.i && pLib.i % 300 == 0 {
            fmt.Printf("Tried %5d first gates...\n", pLib.i);
            fflush( os.Stdout );
        }

        // inlining of macro Super2_LibForEachGate2
        for  pLib.k = 0;                                                       pLib.k < pLib.i && (pGate2 = pLib.pGates[pLib.k]);                 pLib.k++  {
            uTruth1  = pGate1.uTruth;
            uTruth2  = pGate2.uTruth;
            uTruth1c = Mask & ~uTruth1;
            uTruth2c = Mask & ~uTruth2;

            // none complemented
            uTruth  = uTruth1  & uTruth2;
            if uTruth & pLibNew.uMaskBit {
                uTruthR = Mask & ~uTruth;
            } else {
                uTruthR = uTruth;
            }

            if !stmm_find_or_add( pMan.tTable, (char *)(ABC_PTRUINT_T)uTruthR, (char ***)&ppGate ) {
                pGateNew = (Super2_Gate_t *)Extra_MmFixedEntryFetch( pMan.pMem );
                pGateNew.pOne  = pGate1;
                pGateNew.pTwo  = pGate2;
                pGateNew.uTruth = uTruth;
                *ppGate = pGateNew;
                Super2_LibAddGate( pLibNew, pGateNew );
            }                

            // one complemented
            uTruth  = uTruth1c & uTruth2;
            if uTruth & pLibNew.uMaskBit {
                uTruthR = Mask & ~uTruth;
            } else {
                uTruthR = uTruth;
            }

            if !stmm_find_or_add( pMan.tTable, (char *)(ABC_PTRUINT_T)uTruthR, (char ***)&ppGate ) {
                pGateNew = (Super2_Gate_t *)Extra_MmFixedEntryFetch( pMan.pMem );
                // inlining of macro Super2_Not
                pGateNew.pOne  = ((Super2_Gate_t *)((ABC_PTRUINT_T)(pGate1) ^ 01)) ;
                pGateNew.pTwo  = pGate2;
                pGateNew.uTruth = uTruth;
                *ppGate = pGateNew;
                Super2_LibAddGate( pLibNew, pGateNew );
            }                

            // another complemented
            uTruth  = uTruth1  & uTruth2c;
            if uTruth & pLibNew.uMaskBit {
                uTruthR = Mask & ~uTruth;
            } else {
                uTruthR = uTruth;
            }

            if !stmm_find_or_add( pMan.tTable, (char *)(ABC_PTRUINT_T)uTruthR, (char ***)&ppGate ) {
                pGateNew = (Super2_Gate_t *)Extra_MmFixedEntryFetch( pMan.pMem );
                pGateNew.pOne  = pGate1;
                // inlining of macro Super2_Not
                pGateNew.pTwo  = ((Super2_Gate_t *)((ABC_PTRUINT_T)(pGate2) ^ 01)) ;
                pGateNew.uTruth = uTruth;
                *ppGate = pGateNew;
                Super2_LibAddGate( pLibNew, pGateNew );
            }                

            // both complemented
            uTruth  = uTruth1c & uTruth2c;
            if uTruth & pLibNew.uMaskBit {
                uTruthR = Mask & ~uTruth;
            } else {
                uTruthR = uTruth;
            }

            if !stmm_find_or_add( pMan.tTable, (char *)(ABC_PTRUINT_T)uTruthR, (char ***)&ppGate ) {
                pGateNew = (Super2_Gate_t *)Extra_MmFixedEntryFetch( pMan.pMem );
                // inlining of macro Super2_Not
                pGateNew.pOne  = ((Super2_Gate_t *)((ABC_PTRUINT_T)(pGate1) ^ 01)) ;
                // inlining of macro Super2_Not
                pGateNew.pTwo  = ((Super2_Gate_t *)((ABC_PTRUINT_T)(pGate2) ^ 01)) ;
                pGateNew.uTruth = uTruth;
                *ppGate = pGateNew;
                Super2_LibAddGate( pLibNew, pGateNew );
            }                

            pMan.nTried += 4;
        }
    }
    return pLibNew;
}


var s_uMaskBit uint32;
var s_uMaskAll uint32;

/**Function*************************************************************

  Synopsis    [Writes the library into the file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_LibWrite(pLib *Super2_Lib_t) {
    var pGate *Super2_Gate_t;
    var pFile *os.File;
    var FileName string // was [100]byte in C

    var clk abctime;

    if pLib.nLevels > 5 {
        fmt.Printf("Cannot write file for %d levels.\n", pLib.nLevels);
        return;
    }

clk = Abc_Clock();
    // sort the supergates by truth table
    s_uMaskBit = pLib.uMaskBit;
    // inlining of macro SUPER_MASK
    s_uMaskAll = ((~(uint32(0))) >> (32-pLib.nMints));
    qsort( (void *)pLib.pGates, (size_t)pLib.nGates, sizeof(Super2_Gate_t *), 
                (int (*)(void *, void *)) Super2_LibCompareGates );
    assert(Super2_LibCompareGates( pLib.pGates, pLib.pGates + pLib.nGates - 1 ) < 0, "Super2_LibCompareGates( pLib.pGates, pLib.pGates + pLib.nGates - 1 ) < 0");
ABC_PRT( "Sorting", Abc_Clock() - clk );


    // start the file
    FileName = fmt.Sprintf("superI%dL%d", pLib.nInputs, pLib.nLevels);
    pFile = fopen( FileName, "w" );
    fmt.Fprintf(pFile, "# AND2/INV supergates derived on %s.\n", Extra_TimeStamp());
    fmt.Fprintf(pFile, "# Command line: \"super2 -i %d -l %d\".\n", pLib.nInputs, pLib.nLevels);
    fmt.Fprintf(pFile, "# The number of inputs     = %6d.\n", pLib.nInputs);
    fmt.Fprintf(pFile, "# The number of levels     = %6d.\n", pLib.nLevels);
    fmt.Fprintf(pFile, "# The number of supergates = %6d.\n", pLib.nGates);
    fmt.Fprintf(pFile, "# The total functions      = %6d.\n", (1<<(pLib.nMints-1)));
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "%6d\n", pLib.nGates);

    // print the gates
    // inlining of macro Super2_LibForEachGate
    for  pLib.i = 0;                                                       pLib.i < pLib.nGates && (pGate = pLib.pGates[pLib.i]);             pLib.i++  {
        Super2_LibWriteGate( pFile, pLib, pGate );
    }
    _ = (pFile).Close();

    printf( "The supergates are written into file \"%s\" ", FileName );
    fmt.Printf("(%0.2f MB).\n", (float64(Extra_FileSize(FileName)))/(1<<20));
}

/**Function*************************************************************

  Synopsis    [Writes the gate into the file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_LibCompareGates(ppG1 **Super2_Gate_t, ppG2 **Super2_Gate_t) int32 {
    var pG1 *Super2_Gate_t   = *ppG1;
    var pG2 *Super2_Gate_t   = *ppG2;
    var uTruth1, uTruth2 uint32;

    if pG1.uTruth & s_uMaskBit {
        uTruth1 = s_uMaskAll & ~pG1.uTruth;
    } else {
        uTruth1 = pG1.uTruth;
    }
    if pG2.uTruth & s_uMaskBit {
        uTruth2 = s_uMaskAll & ~pG2.uTruth;
    } else {
        uTruth2 = pG2.uTruth;
    }

    if uTruth1 < uTruth2 {
        return -1; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Writes the gate into the file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_LibWriteGate(pFile *os.File, pLib *Super2_Lib_t, pGate *Super2_Gate_t) {
//    unsigned uTruthR;
    var uTruth uint32;
    var fInv int32;

    // check whether the gate need complementation
    fInv = int32((pGate.uTruth & pLib.uMaskBit));
    if fInv != 0 {
        uTruth = ~pGate.uTruth;
    } else {
        uTruth = pGate.uTruth;
    }
/*
    // reverse the truth table
    uTruthR = 0;
    for ( m = 0; m < pLib->nMints; m++ )
        if ( uTruth & (1 << m) )
            uTruthR |= (1 << (pLib->nMints-1-m));
*/
    // write the truth table
    Extra_PrintBinary( pFile, &uTruth, pLib.nMints );
    fmt.Fprintf(pFile, "   ");
    // write the symbolic expression
    fmt.Fprintf(pFile, "%s", Super2_LibWriteGate_rec( pGate, fInv, pLib.nLevels ));
    fmt.Fprintf(pFile, "\n");
}


/**Function*************************************************************

  Synopsis    [Recursively writes the gate into the file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_LibWriteGate_rec(pGate *Super2_Gate_t, fInv int32, Level int32) string {
    static char Buff01[  3], Buff02[  3];    // Max0              =  1
    static char Buff11[  6], Buff12[  6];    // Max1 = 2*Max0 + 2 =  4
    static char Buff21[ 12], Buff22[ 12];    // Max2 = 2*Max1 + 2 = 10
    static char Buff31[ 25], Buff32[ 25];    // Max3 = 2*Max2 + 2 = 22
    static char Buff41[ 50], Buff42[ 50];    // Max4 = 2*Max3 + 2 = 46
    static char Buff51[100], Buff52[100];    // Max5 = 2*Max4 + 2 = 94
    static char * pBuffs1[6] = { Buff01, Buff11, Buff21, Buff31, Buff41, Buff51 };
    static char * pBuffs2[6] = { Buff02, Buff12, Buff22, Buff32, Buff42, Buff52 };
    var pBranch string;
    var pBuffer1 string  = pBuffs1[Level];
    var pBuffer2 string  = pBuffs2[Level];
    var pGateNext1, pGateNext2 *Super2_Gate_t;
    var fInvNext1, fInvNext2 int32;
    var RetValue int32;

    // consider the last level
    assert(Level >= 0, "Level >= 0");
    if pGate.pOne == nil {
        if pGate.uTruth == 0 {
            pBuffer1[0] = (fInv? '1': '0');
            pBuffer1[1] = '$';
            pBuffer1[2] = 0;
        
        } else {
            pBuffer1[0] = (fInv? 'A' + (int32((ABC_PTRUINT_T))pGate.pTwo): 'a' + (int32((ABC_PTRUINT_T))pGate.pTwo));
            pBuffer1[1] = 0;
        }
        return pBuffer1;
    }
    assert(Level > 0, "Level > 0");


    // get the left branch
    // inlining of macro Super2_Regular
    pGateNext1 = ((Super2_Gate_t *)((ABC_PTRUINT_T)(pGate.pOne) & ~01)) ;
    // inlining of macro Super2_IsComplement
    fInvNext1  = ((int32(((ABC_PTRUINT_T) (pGate.pOne) & 01))));
    pBranch    = Super2_LibWriteGate_rec(pGateNext1, fInvNext1, Level - 1);
    // copy into Buffer1
    pBuffer1 = pBranch;

    // get the right branch
    // inlining of macro Super2_Regular
    pGateNext2 = ((Super2_Gate_t *)((ABC_PTRUINT_T)(pGate.pTwo) & ~01)) ;
    // inlining of macro Super2_IsComplement
    fInvNext2  = ((int32(((ABC_PTRUINT_T) (pGate.pTwo) & 01))));
    pBranch    = Super2_LibWriteGate_rec(pGateNext2, fInvNext2, Level - 1);

    // consider the case when comparison is not necessary
    if fInvNext1 ^ fInvNext2 {
        if fInvNext1 > fInvNext2 {
            pBuffer2 = fmt.Sprintf("%c%s%s%c", (fInv? '<': '('), pBuffer1, pBranch, (fInv? '>': ')'));
        } else {
            pBuffer2 = fmt.Sprintf("%c%s%s%c", (fInv? '<': '('), pBranch, pBuffer1, (fInv? '>': ')')); }
    
    } else {
        // compare the two branches
        RetValue = Super2_LibWriteCompare( pBuffer1, pBranch );
        if RetValue == 1 {
            pBuffer2 = fmt.Sprintf("%c%s%s%c", (fInv? '<': '('), pBuffer1, pBranch, (fInv? '>': ')'));
        } else { // if ( RetValue == -1 )
        
            pBuffer2 = fmt.Sprintf("%c%s%s%c", (fInv? '<': '('), pBranch, pBuffer1, (fInv? '>': ')'));
            if RetValue == 0 {
                fmt.Printf("Strange!\n"); }
        }
    }
    return pBuffer2;
}

/**Function*************************************************************

  Synopsis    [Compares the two branches of the tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super2_LibWriteCompare(pStr1 string, pStr2 string) int32 {
    for {
        // skip extra symbols
        for  *pStr1 && *pStr1 < 'A'  {
            pStr1++;
        }
        for  *pStr2 && *pStr2 < 'A'  {
            pStr2++;
        }

        // check if any one is finished
        if *pStr1 == 0 || *pStr2 == 0 {
            if *pStr2 {
                return 1; }
            return -1;
        }

        // compare
        if *pStr1 == *pStr2 {
            pStr1++;
            pStr2++;
        
        } else {
            if *pStr1 < *pStr2 {
                return 1; }
            return -1;
        }
    }
    return 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


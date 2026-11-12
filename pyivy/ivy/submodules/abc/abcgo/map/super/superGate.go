package super

import (
	"fmt"
	"math"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [superGate.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Pre-computation of supergates.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 8, 2003.]

  Revision    [$Id: superGate.c,v 1.7 2004/08/03 00:11:40 satrajit Exp $]

***********************************************************************/

//#include <math.h>
//#include "superInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the bit masks
// Original macro:
// #define SUPER_MASK(n)     ((~((unsigned)0)) >> (32-(n)))

#define SUPER_FULL         (~((unsigned)0))
#define SUPER_NO_VAR         (-9999.0)
#define SUPER_EPSILON        (0.001)

// data structure for supergate precomputation
type Super_Man_t =     Super_ManStruct_t_;   // manager
type Super_Gate_t =    Super_GateStruct_t_;  // supergate

type Super_ManStruct_t_ struct {
    // parameters
    pName               string;        // the original genlib file name
    nVarsMax                 int32;     // the number of inputs
    nMints                 int32;       // the number of minterms
    nLevels                 int32;      // the number of logic levels
    nGatesMax                 int32;    // the number of gates computed
    tDelayMax               float32;    // the max delay of the supergates in the library
    tAreaMax               float32;     // the max area of the supergates in the library
    fSkipInv                 int32;     // the flag says about skipping inverters
    fWriteOldFormat                 int32; // in addition, writes the file in the old format
    fVerbose                 int32;

    // supergates
    pInputs [10]*Super_Gate_t;  // the input supergates
    nGates                 int32;       // the number of gates in the library
    pGates      **Super_Gate_t;       // the gates themselves
    tTable         *stmm_table;       // mapping of truth tables into gates

    // memory managers
    pMem    *Extra_MmFixed_t;         // memory manager for the supergates
    pMemFlex     *Extra_MmFlex_t;     // memory manager for the fanin arrays

    // statistics
    nTried                 int32;       // the total number of tried
    nAdded                 int32;       // the number of entries added
    nRemoved                 int32;     // the number of entries removed
    nClasses                 int32;     // the number of gate classes
    nUnique                 int32;      // the number of unique gates
    nLookups                 int32;     // the number of hash table lookups
    nAliases                 int32;     // the number of hash table lookups thrown away due to aliasing

    // runtime
    Time             abctime;         // the runtime of the generation procedure
    TimeLimit                 int32;    // the runtime limit (in seconds)
    TimeSec                 int32;      // the time passed (in seconds)
    TimeStop             abctime;     // the time to stop computation (in miliseconds)
    TimePrint             abctime;    // the time to print message
}

type Super_GateStruct_t_ struct {
    pRoot         *Mio_Gate_t;        // the root gate for this supergate
    fVar uint32; /* C bitfield: 1 bits; review packing and truncation */ // the flag signaling the elementary variable
    fSuper uint32; /* C bitfield: 1 bits; review packing and truncation */ // the flag signaling the elementary variable
    nFanins uint32; /* C bitfield: 6 bits; review packing and truncation */ // the number of fanin gates
    Number uint32; /* C bitfield: 24 bits; review packing and truncation */ // the number assigned in the process
    uTruth [2]uint32;    // the truth table of this supergate
    pFanins [6]*Super_Gate_t;   // the fanins of the gate
    Area               float32;         // the area of this gate
    ptDelays [6]float32;  // the pin-to-pin delays for all inputs
    tDelayMax               float32;    // the maximum delay
    pNext       *Super_Gate_t;        // the next gate in the table
}


// iterating through the gates in the library
// Original macro:
// #define Super_ManForEachGate( GateArray, Limit, Index, Gate )    \
//     for ( Index = 0;                                             \
//           Index < Limit && (Gate = GateArray[Index]);            \
//           Index++ )


// static functions
// static Super_Man_t *  Super_ManStart();

// static void           Super_ManStop( Super_Man_t * pMan );


// static void           Super_AddGateToTable( Super_Man_t * pMan, Super_Gate_t * pGate );

// static void           Super_First( Super_Man_t * pMan, int nVarsMax );

// static Super_Man_t *  Super_Compute( Super_Man_t * pMan, Mio_Gate_t ** ppGates, int nGates, int nGatesMax, int fSkipInv );

// static Super_Gate_t * Super_CreateGateNew( Super_Man_t * pMan, Mio_Gate_t * pRoot, Super_Gate_t ** pSupers, int nSupers, unsigned uTruth[], float Area, float tPinDelaysRes[], float tDelayMax, int nPins );

// static int            Super_CompareGates( Super_Man_t * pMan, unsigned uTruth[], float Area, float tPinDelaysRes[], int nPins );

// static int            Super_DelayCompare( Super_Gate_t ** ppG1, Super_Gate_t ** ppG2 );

// static int            Super_AreaCompare( Super_Gate_t ** ppG1, Super_Gate_t ** ppG2 );

// static void           Super_TranferGatesToArray( Super_Man_t * pMan );

// static int            Super_CheckTimeout( ProgressBar * pPro, Super_Man_t * pMan );

 
// static Vec_Str_t *    Super_Write( Super_Man_t * pMan );

// static int            Super_WriteCompare( Super_Gate_t ** ppG1, Super_Gate_t ** ppG2 );

// static void           Super_WriteFileHeader( Super_Man_t * pMan, FILE * pFile );


// static void           Super_WriteLibrary( Super_Man_t * pMan );


// static void           Super_WriteLibraryTreeFile( Super_Man_t * pMan );

// static Vec_Str_t *    Super_WriteLibraryTreeStr( Super_Man_t * pMan );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Precomputes the library of supergates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_Precompute(pLibGen *Mio_Library_t, nVarsMax int32, nLevels int32, nGatesMax int32, tDelayMax float32, tAreaMax float32, TimeLimit int32, fSkipInv int32, fVerbose int32, pFileName string) {
    var vStr *Vec_Str_t;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {     
        fmt.Printf("Cannot open output file \"%s\".\n", pFileName);
        return;
    }
    vStr = Super_PrecomputeStr( pLibGen, nVarsMax, nLevels, nGatesMax, tDelayMax, tAreaMax, TimeLimit, fSkipInv, fVerbose );
    fwrite( Vec_StrArray(vStr), 1, Vec_StrSize(vStr), pFile );
    _ = (pFile).Close();
    Vec_StrFree( vStr );
    // report the result of writing
    if fVerbose != 0 {
        fmt.Printf("The supergates are written using new format \"%s\" ", pFileName);
        fmt.Printf("(%0.3f MB).\n", (float64(Extra_FileSize(pFileName)))/(1<<20));
    }
}

/**Function*************************************************************

  Synopsis    [Precomputes the library of supergates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_PrecomputeStr(pLibGen *Mio_Library_t, nVarsMax int32, nLevels int32, nGatesMax int32, tDelayMax float32, tAreaMax float32, TimeLimit int32, fSkipInv int32, fVerbose int32) *Vec_Str_t {
    var vStr *Vec_Str_t;
    var pMan *Super_Man_t;
    var ppGates **Mio_Gate_t;
    var nGates, Level int32;
    var clk, clockStart abctime;

    assert(nVarsMax < 7, "nVarsMax < 7");
    if (nGatesMax != 0) && nGatesMax < nVarsMax {
        fmt.Fprintf(os.Stderr, "Erro! The number of supergates requested (%d) in less than the number of variables (%d).\n", nGatesMax, nVarsMax);
        fmt.Fprintf(os.Stderr, "The library cannot be computed.\n");
        return nil;
    }

    // get the root gates
    ppGates = Mio_CollectRoots( pLibGen, nVarsMax, tDelayMax, 0, &nGates, fVerbose );
    if (nGatesMax != 0) && nGates >= nGatesMax {
        fmt.Fprintf(os.Stdout, "Warning! Genlib library contains more gates than can be computed.\n");
        fmt.Fprintf(os.Stdout, "Only one-gate supergates are included in the supergate library.\n");
    }

    // start the manager
    pMan = Super_ManStart();
    pMan.pName     = Mio_LibraryReadName(pLibGen);
    pMan.nGatesMax = nGatesMax;
    pMan.fSkipInv  = fSkipInv;
    pMan.tDelayMax = tDelayMax;
    pMan.tAreaMax  = tAreaMax;
    pMan.TimeLimit = TimeLimit; // in seconds
    if TimeLimit {
        pMan.TimeStop = TimeLimit * int64(time.Second) + Abc_Clock();
    } else {
        pMan.TimeStop = 0;
    } // in CPU ticks
    pMan.fVerbose  = fVerbose;

    if nGates == 0 {
        fmt.Fprintf(os.Stderr, "Error: No genlib gates satisfy the limits criteria. Stop.\n");
        fmt.Fprintf(os.Stderr, "Limits: max delay =  %.2f, max area =  %.2f, time limit = %d sec.\n", pMan.tDelayMax, pMan.tAreaMax, pMan.TimeLimit);

        // stop the manager
        Super_ManStop( pMan );
        ABC_FREE( ppGates );

        return nil;
    }

    // get the starting supergates
    Super_First( pMan, nVarsMax );

    // perform the computation of supergates
    clockStart = Abc_Clock();
if fVerbose != 0 {
    fmt.Printf("Computing supergates with %d inputs, %d levels, and %d max gates.\n", pMan.nVarsMax, nLevels, nGatesMax);
    fmt.Printf("Limits: max delay =  %.2f, max area =  %.2f, time limit = %d sec.\n", pMan.tDelayMax, pMan.tAreaMax, pMan.TimeLimit);
}

    for  Level = 1; Level <= nLevels; Level++  {
        if pMan.TimeStop && Abc_Clock() > pMan.TimeStop {
            break; }
clk = Abc_Clock();
        Super_Compute( pMan, ppGates, nGates, nGatesMax, fSkipInv );
        pMan.nLevels = Level;
if fVerbose != 0 {
        fmt.Printf("Lev %d: Try =%12d. Add =%6d. Rem =%5d. Save =%6d. Lookups =%12d. Aliases =%12d. ", Level, pMan.nTried, pMan.nAdded, pMan.nRemoved, pMan.nAdded - pMan.nRemoved, pMan.nLookups, pMan.nAliases);
ABC_PRT( "Time", Abc_Clock() - clk );
fflush( os.Stdout );
}
    }
    pMan.Time = Abc_Clock() - clockStart;

if fVerbose != 0 {
fmt.Printf("Writing the output file...\n");
fflush( os.Stdout );
}
    // write them into a file
    vStr = Super_Write( pMan );

    // stop the manager
    Super_ManStop( pMan );
    ABC_FREE( ppGates );
    return vStr;
}


/**Function*************************************************************

  Synopsis    [Derives the starting supergates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_First(pMan *Super_Man_t, nVarsMax int32) {
    var pSuper *Super_Gate_t;
    var nMintLimit, nVarLimit int32;
    var v, m int32;
    // set the parameters
    pMan.nVarsMax  = nVarsMax;
    pMan.nMints    = (1 << nVarsMax);
    pMan.nLevels   = 0;
    // allocate room for the gates
    pMan.nGates    = nVarsMax;  
    pMan.pGates    = ABC_ALLOC( Super_Gate_t *, nVarsMax + 2 ); 
    // create the gates corresponding to the elementary variables
    for  v = 0; v < nVarsMax; v++  {
        // get a new gate
        pSuper = (Super_Gate_t *)Extra_MmFixedEntryFetch( pMan.pMem );
        *pSuper = Super_Gate_t{};
        // assign the elementary variable, the truth table, and the delays
        pSuper.fVar = 1;
        pSuper.Number = v;
        for  m = 0; m < nVarsMax; m++  {
            pSuper.ptDelays[m] = SUPER_NO_VAR;
        }
        pSuper.ptDelays[v] = 0.0;
        // set the gate
        pMan.pGates[v] = pSuper;
        Super_AddGateToTable( pMan, pSuper );
        pMan.pInputs[v] = pSuper;
    }
    // set up their truth tables
    if nVarsMax >= 5 {
        nVarLimit = 5;
    } else {
        nVarLimit = nVarsMax;
    }
    nMintLimit = (1 << nVarLimit);
    for  m = 0; m < nMintLimit; m++  {
        for  v = 0; v < nVarLimit; v++  {
            if m & (1 << v) {
                pMan.pGates[v].uTruth[0] |= (1 << m); }
        }
    }
    // make adjustments for the case of 6 variables
    if nVarsMax == 6 {
        for  v = 0; v < 5; v++  {
            pMan.pGates[v].uTruth[1] = pMan.pGates[v].uTruth[0];
        }
        pMan.pGates[5].uTruth[0] = 0;
        pMan.pGates[5].uTruth[1] = ~(uint32(0));
    
    } else {
        for  v = 0; v < nVarsMax; v++  {
            pMan.pGates[v].uTruth[1] = 0;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Precomputes one level of supergates.]

  Description [This procedure computes the set of supergates that can be
  derived from the given set of root gates (from genlib library) by composing
  the root gates with the currently available supergates. This procedure is
  smart in the sense that it tries to avoid useless emuration by imposing
  tight bounds by area and delay. Only the supergates and are guaranteed to 
  have smaller area and delay are enumereated. See comments below for details.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_Compute(pMan *Super_Man_t, ppGates **Mio_Gate_t, nGates int32, nGatesMax int32, fSkipInv int32) *Super_Man_t {
    var pSupers [6]*Super_Gate_t;
    var pGate0 *Super_Gate_t;
    var pGate1 *Super_Gate_t;
    var pGate2 *Super_Gate_t;
    var pGate3 *Super_Gate_t;
    var pGate4 *Super_Gate_t;
    var pGate5 *Super_Gate_t;
    var pGateNew *Super_Gate_t;
    var tPinDelaysRes [6]float32;
    var ptPinDelays [6]*float32;
    var tPinDelayMax float32;
    var tDelayMio float32;
    var Area float32  = 0.0; // Suppress "might be used uninitialized"
    var Area0, Area1, Area2, Area3, Area4, AreaMio float32;
    var uTruth [2]uint32;
    var uTruths [6][2]uint32;
    var i0, i1, i2, i3, i4, i5 int32; 
    var ppGatesLimit **Super_Gate_t;
    var nFanins, nGatesLimit, k, s, t int32;
    var pProgress *ProgressBar;
    var fTimeOut int32;
    var fPrune int32  = 1;                     // Shall we prune?
    var iPruneLimit int32  = 3;                // Each of the gates plugged into the root gate will have 
                                        // less than these many fanins
    var iPruneLimitRoot int32  = 4;            // The root gate may have only less than these many fanins

    // put the gates from the unique table into the array
    // the gates from the array will be used to compose other gates
    // the gates in tbe table are used to check uniqueness of collected gates
    Super_TranferGatesToArray( pMan );

    // sort the gates in the increasing order of maximum delay
    if pMan.nGates > 10000 {
        fmt.Printf("Sorting array of %d supergates...\r", pMan.nGates);
        fflush( os.Stdout );
    }
    qsort( (void *)pMan.pGates, (size_t)pMan.nGates, sizeof(Super_Gate_t *), 
                (int (*)(void *, void *)) Super_DelayCompare );
    assert(Super_DelayCompare( pMan.pGates, pMan.pGates + pMan.nGates - 1 ) <= 0, "Super_DelayCompare( pMan.pGates, pMan.pGates + pMan.nGates - 1 ) <= 0");
    if pMan.nGates > 10000 {
        fmt.Printf("                                       \r");
    }

    pProgress = Extra_ProgressBarStart( os.Stdout, pMan.TimeLimit );
    pMan.TimePrint = Abc_Clock() + int64(time.Second);
    ppGatesLimit = ABC_ALLOC( Super_Gate_t *, pMan.nGates );
    // go through the root gates
    // the root gates are sorted in the increasing gelay
    fTimeOut = 0;
    for  k = 0; k < nGates; k++  {
        if fTimeOut != 0 { break; }

        if fPrune != 0 {
            if pMan.nLevels >= 1 {  // First level gates have been computed
            
                if Mio_GateReadPinNum(ppGates[k]) >= iPruneLimitRoot {
                    continue; }
            }
        }
/*
        if ( strcmp(Mio_GateReadName(ppGates[k]), "MUX2IX0") == 0 )
        {
            int s = 0;
        }
*/
        // select the subset of gates to be considered with this root gate
        // all the gates past this point will lead to delay larger than the limit
        tDelayMio = float32(Mio_GateReadDelayMax(ppGates[k]));
        for  s, t = 0, 0; s < pMan.nGates; s++  {
            if (fPrune != 0) && ( pMan.nLevels >= 1 ) && ( (int32(pMan.pGates[s].nFanins)) >= iPruneLimit ) {
                continue; }
            
            ppGatesLimit[t] = pMan.pGates[s];
            if ppGatesLimit[t++].tDelayMax + tDelayMio > pMan.tDelayMax && pMan.tDelayMax > 0.0 {
                break; }
        }
        nGatesLimit = t;

        if pMan.fVerbose {
            fmt.Printf("Trying %d choices for %d inputs\r", t, Mio_GateReadPinNum(ppGates[k]));
        }

        // resort part of this range by area
        // now we can prune the search by going up in the list until we reach the limit on area
        // all the gates beyond this point can be skipped because their area can be only larger
        if nGatesLimit > 10000 {
            fmt.Printf("Sorting array of %d supergates...\r", nGatesLimit); }
        qsort( (void *)ppGatesLimit, (size_t)nGatesLimit, sizeof(Super_Gate_t *), 
                        (int (*)(void *, void *)) Super_AreaCompare );
        assert(Super_AreaCompare( ppGatesLimit, ppGatesLimit + nGatesLimit - 1 ) <= 0, "Super_AreaCompare( ppGatesLimit, ppGatesLimit + nGatesLimit - 1 ) <= 0");
        if nGatesLimit > 10000 {
            fmt.Printf("                                       \r"); }

        // consider the combinations of gates with the root gate on top
        AreaMio = float32(Mio_GateReadArea(ppGates[k]));
        nFanins = Mio_GateReadPinNum(ppGates[k]);
        switch  nFanins  {
        case 0: // should not happen
            assert(false, "0"); 
            break;
        case 1: // interter root
            // inlining of macro Super_ManForEachGate
            for  i0 = 0;                                                       i0 < nGatesLimit && (pGate0 = ppGatesLimit[i0]);                      i0++  {
              if fTimeOut != 0 { break; }
              fTimeOut = Super_CheckTimeout( pProgress, pMan );
              // skip the inverter as the root gate before the elementary variable
              // as a result, the supergates will not have inverters on the input side
              // but inverters still may occur at the output of or inside complex supergates
              if (fSkipInv != 0) && pGate0.tDelayMax == 0 {
                  continue; }
              // compute area
              Area = AreaMio + pGate0.Area;
              if pMan.tAreaMax > 0.0 && Area > pMan.tAreaMax {
                  break; }

              pSupers[0] = pGate0;  uTruths[0][0] = pGate0.uTruth[0];  uTruths[0][1] = pGate0.uTruth[1];  ptPinDelays[0] = pGate0.ptDelays; 
              Mio_DeriveGateDelays( ppGates[k], ptPinDelays, nFanins, pMan.nVarsMax, SUPER_NO_VAR, tPinDelaysRes, &tPinDelayMax );
              Mio_DeriveTruthTable( ppGates[k], uTruths, nFanins, pMan.nVarsMax, uTruth );
              if !(Super_CompareGates( pMan, uTruth, Area, tPinDelaysRes, pMan.nVarsMax ) != 0) {
                  continue; }
              // create a new gate
              pGateNew = Super_CreateGateNew( pMan, ppGates[k], pSupers, nFanins, uTruth, Area, tPinDelaysRes, tPinDelayMax, pMan.nVarsMax );
              Super_AddGateToTable( pMan, pGateNew );
              if (nGatesMax != 0) && pMan.nClasses > nGatesMax {
                  goto done; }
            }
            break;
        case 2: // two-input root gate
            // inlining of macro Super_ManForEachGate
            for  i0 = 0;                                                       i0 < nGatesLimit && (pGate0 = ppGatesLimit[i0]);                      i0++  {
              Area0 = AreaMio + pGate0.Area;
              if pMan.tAreaMax > 0.0 && Area0 > pMan.tAreaMax {
                  break; }
              pSupers[0] = pGate0;  uTruths[0][0] = pGate0.uTruth[0];  uTruths[0][1] = pGate0.uTruth[1];  ptPinDelays[0] = pGate0.ptDelays; 
              // inlining of macro Super_ManForEachGate
              for  i1 = 0;                                                       i1 < nGatesLimit && (pGate1 = ppGatesLimit[i1]);                      i1++  {
              if i1 != i0 {
                if fTimeOut != 0 { goto done; }
                fTimeOut = Super_CheckTimeout( pProgress, pMan );
                // compute area
                Area = Area0 + pGate1.Area;
                if pMan.tAreaMax > 0.0 && Area > pMan.tAreaMax {
                    break; }

                pSupers[1] = pGate1;  uTruths[1][0] = pGate1.uTruth[0];  uTruths[1][1] = pGate1.uTruth[1];  ptPinDelays[1] = pGate1.ptDelays;
                Mio_DeriveGateDelays( ppGates[k], ptPinDelays, nFanins, pMan.nVarsMax, SUPER_NO_VAR, tPinDelaysRes, &tPinDelayMax );
                Mio_DeriveTruthTable( ppGates[k], uTruths, nFanins, pMan.nVarsMax, uTruth );
                if !(Super_CompareGates( pMan, uTruth, Area, tPinDelaysRes, pMan.nVarsMax ) != 0) {
                    continue; }
                // create a new gate
                pGateNew = Super_CreateGateNew( pMan, ppGates[k], pSupers, nFanins, uTruth, Area, tPinDelaysRes, tPinDelayMax, pMan.nVarsMax );
                Super_AddGateToTable( pMan, pGateNew );
                if (nGatesMax != 0) && pMan.nClasses > nGatesMax {
                    goto done; }
              }
              }
            }
            break;
        case 3: // three-input root gate
            // inlining of macro Super_ManForEachGate
            for  i0 = 0;                                                       i0 < nGatesLimit && (pGate0 = ppGatesLimit[i0]);                      i0++  {
              Area0 = AreaMio + pGate0.Area;
              if pMan.tAreaMax > 0.0 && Area0 > pMan.tAreaMax {
                  break; }
              pSupers[0] = pGate0;  uTruths[0][0] = pGate0.uTruth[0];  uTruths[0][1] = pGate0.uTruth[1];  ptPinDelays[0] = pGate0.ptDelays; 

              // inlining of macro Super_ManForEachGate
              for  i1 = 0;                                                       i1 < nGatesLimit && (pGate1 = ppGatesLimit[i1]);                      i1++  {
              if i1 != i0 {
                Area1 = Area0 + pGate1.Area;
                if pMan.tAreaMax > 0.0 && Area1 > pMan.tAreaMax {
                    break; }
                pSupers[1] = pGate1;  uTruths[1][0] = pGate1.uTruth[0];  uTruths[1][1] = pGate1.uTruth[1];  ptPinDelays[1] = pGate1.ptDelays;

                // inlining of macro Super_ManForEachGate
                for  i2 = 0;                                                       i2 < nGatesLimit && (pGate2 = ppGatesLimit[i2]);                      i2++  {
                if i2 != i0 && i2 != i1 {
                  if fTimeOut != 0 { goto done; }
                  fTimeOut = Super_CheckTimeout( pProgress, pMan );
                  // compute area
                  Area = Area1 + pGate2.Area;
                  if pMan.tAreaMax > 0.0 && Area > pMan.tAreaMax {
                      break; }
                  pSupers[2] = pGate2;  uTruths[2][0] = pGate2.uTruth[0];  uTruths[2][1] = pGate2.uTruth[1];   ptPinDelays[2] = pGate2.ptDelays;

                  Mio_DeriveGateDelays( ppGates[k], ptPinDelays, nFanins, pMan.nVarsMax, SUPER_NO_VAR, tPinDelaysRes, &tPinDelayMax );
                  Mio_DeriveTruthTable( ppGates[k], uTruths, nFanins, pMan.nVarsMax, uTruth );
                  if !(Super_CompareGates( pMan, uTruth, Area, tPinDelaysRes, pMan.nVarsMax ) != 0) {
                      continue; }
                  // create a new gate
                  pGateNew = Super_CreateGateNew( pMan, ppGates[k], pSupers, nFanins, uTruth, Area, tPinDelaysRes, tPinDelayMax, pMan.nVarsMax );
                  Super_AddGateToTable( pMan, pGateNew );
                  if (nGatesMax != 0) && pMan.nClasses > nGatesMax {
                      goto done; }
                }
                }
              }
              }
            }
            break;
        case 4: // four-input root gate
            // inlining of macro Super_ManForEachGate
            for  i0 = 0;                                                       i0 < nGatesLimit && (pGate0 = ppGatesLimit[i0]);                      i0++  {
              Area0 = AreaMio + pGate0.Area;
              if pMan.tAreaMax > 0.0 && Area0 > pMan.tAreaMax {
                  break; }
              pSupers[0] = pGate0;  uTruths[0][0] = pGate0.uTruth[0];  uTruths[0][1] = pGate0.uTruth[1];  ptPinDelays[0] = pGate0.ptDelays; 

              // inlining of macro Super_ManForEachGate
              for  i1 = 0;                                                       i1 < nGatesLimit && (pGate1 = ppGatesLimit[i1]);                      i1++  {
              if i1 != i0 {
                Area1 = Area0 + pGate1.Area;
                if pMan.tAreaMax > 0.0 && Area1 > pMan.tAreaMax {
                    break; }
                pSupers[1] = pGate1;  uTruths[1][0] = pGate1.uTruth[0];  uTruths[1][1] = pGate1.uTruth[1];  ptPinDelays[1] = pGate1.ptDelays;

                // inlining of macro Super_ManForEachGate
                for  i2 = 0;                                                       i2 < nGatesLimit && (pGate2 = ppGatesLimit[i2]);                      i2++  {
                if i2 != i0 && i2 != i1 {
                  Area2 = Area1 + pGate2.Area;
                  if pMan.tAreaMax > 0.0 && Area2 > pMan.tAreaMax {
                      break; }
                  pSupers[2] = pGate2;  uTruths[2][0] = pGate2.uTruth[0];  uTruths[2][1] = pGate2.uTruth[1];   ptPinDelays[2] = pGate2.ptDelays;

                  // inlining of macro Super_ManForEachGate
                  for  i3 = 0;                                                       i3 < nGatesLimit && (pGate3 = ppGatesLimit[i3]);                      i3++  {
                  if i3 != i0 && i3 != i1 && i3 != i2 {
                    if fTimeOut != 0 { goto done; }
                    fTimeOut = Super_CheckTimeout( pProgress, pMan );
                    // compute area
                    Area = Area2 + pGate3.Area;
                    if pMan.tAreaMax > 0.0 && Area > pMan.tAreaMax {
                        break; }
                    pSupers[3] = pGate3;   uTruths[3][0] = pGate3.uTruth[0];  uTruths[3][1] = pGate3.uTruth[1];   ptPinDelays[3] = pGate3.ptDelays;

                    Mio_DeriveGateDelays( ppGates[k], ptPinDelays, nFanins, pMan.nVarsMax, SUPER_NO_VAR, tPinDelaysRes, &tPinDelayMax );
                    Mio_DeriveTruthTable( ppGates[k], uTruths, nFanins, pMan.nVarsMax, uTruth );
                    if !(Super_CompareGates( pMan, uTruth, Area, tPinDelaysRes, pMan.nVarsMax ) != 0) {
                        continue; }
                    // create a new gate
                    pGateNew = Super_CreateGateNew( pMan, ppGates[k], pSupers, nFanins, uTruth, Area, tPinDelaysRes, tPinDelayMax, pMan.nVarsMax );
                    Super_AddGateToTable( pMan, pGateNew );
                    if (nGatesMax != 0) && pMan.nClasses > nGatesMax {
                        goto done; }
                  }
                  }
                }
                }
              }
              }
            }
            break;
        case 5: // five-input root gate
            // inlining of macro Super_ManForEachGate
            for  i0 = 0;                                                       i0 < nGatesLimit && (pGate0 = ppGatesLimit[i0]);                      i0++  {
              Area0 = AreaMio + pGate0.Area;
              if pMan.tAreaMax > 0.0 && Area0 > pMan.tAreaMax {
                  break; }
              pSupers[0] = pGate0;  uTruths[0][0] = pGate0.uTruth[0];  uTruths[0][1] = pGate0.uTruth[1];  ptPinDelays[0] = pGate0.ptDelays; 

              // inlining of macro Super_ManForEachGate
              for  i1 = 0;                                                       i1 < nGatesLimit && (pGate1 = ppGatesLimit[i1]);                      i1++  {
              if i1 != i0 {
                Area1 = Area0 + pGate1.Area;
                if pMan.tAreaMax > 0.0 && Area1 > pMan.tAreaMax {
                    break; }
                pSupers[1] = pGate1;  uTruths[1][0] = pGate1.uTruth[0];  uTruths[1][1] = pGate1.uTruth[1];  ptPinDelays[1] = pGate1.ptDelays;

                // inlining of macro Super_ManForEachGate
                for  i2 = 0;                                                       i2 < nGatesLimit && (pGate2 = ppGatesLimit[i2]);                      i2++  {
                if i2 != i0 && i2 != i1 {
                  Area2 = Area1 + pGate2.Area;
                  if pMan.tAreaMax > 0.0 && Area2 > pMan.tAreaMax {
                      break; }
                  pSupers[2] = pGate2;  uTruths[2][0] = pGate2.uTruth[0];  uTruths[2][1] = pGate2.uTruth[1];   ptPinDelays[2] = pGate2.ptDelays;

                  // inlining of macro Super_ManForEachGate
                  for  i3 = 0;                                                       i3 < nGatesLimit && (pGate3 = ppGatesLimit[i3]);                      i3++  {
                  if i3 != i0 && i3 != i1 && i3 != i2 {
                    Area3 = Area2 + pGate3.Area;
                    if pMan.tAreaMax > 0.0 && Area3 > pMan.tAreaMax {
                        break; }
                    pSupers[3] = pGate3;   uTruths[3][0] = pGate3.uTruth[0];  uTruths[3][1] = pGate3.uTruth[1];   ptPinDelays[3] = pGate3.ptDelays;

                    // inlining of macro Super_ManForEachGate
                    for  i4 = 0;                                                       i4 < nGatesLimit && (pGate4 = ppGatesLimit[i4]);                      i4++  {
                    if i4 != i0 && i4 != i1 && i4 != i2 && i4 != i3 {
                      if fTimeOut != 0 { goto done; }
                      fTimeOut = Super_CheckTimeout( pProgress, pMan );
                      // compute area
                      Area = Area3 + pGate4.Area;
                      if pMan.tAreaMax > 0.0 && Area > pMan.tAreaMax {
                          break; }
                      pSupers[4] = pGate4;   uTruths[4][0] = pGate4.uTruth[0];  uTruths[4][1] = pGate4.uTruth[1];  ptPinDelays[4] = pGate4.ptDelays;

                      Mio_DeriveGateDelays( ppGates[k], ptPinDelays, nFanins, pMan.nVarsMax, SUPER_NO_VAR, tPinDelaysRes, &tPinDelayMax );
                      Mio_DeriveTruthTable( ppGates[k], uTruths, nFanins, pMan.nVarsMax, uTruth );
                      if !(Super_CompareGates( pMan, uTruth, Area, tPinDelaysRes, pMan.nVarsMax ) != 0) {
                          continue; }
                      // create a new gate
                      pGateNew = Super_CreateGateNew( pMan, ppGates[k], pSupers, nFanins, uTruth, Area, tPinDelaysRes, tPinDelayMax, pMan.nVarsMax );
                      Super_AddGateToTable( pMan, pGateNew );
                      if (nGatesMax != 0) && pMan.nClasses > nGatesMax {
                          goto done; }
                    }
                    }
                  }
                  }
                }
                }
              }
              }
            }
            break;
        case 6: // six-input root gate
            // inlining of macro Super_ManForEachGate
            for  i0 = 0;                                                       i0 < nGatesLimit && (pGate0 = ppGatesLimit[i0]);                      i0++  {
              Area0 = AreaMio + pGate0.Area;
              if pMan.tAreaMax > 0.0 && Area0 > pMan.tAreaMax {
                  break; }
              pSupers[0] = pGate0;  uTruths[0][0] = pGate0.uTruth[0];  uTruths[0][1] = pGate0.uTruth[1];  ptPinDelays[0] = pGate0.ptDelays; 

              // inlining of macro Super_ManForEachGate
              for  i1 = 0;                                                       i1 < nGatesLimit && (pGate1 = ppGatesLimit[i1]);                      i1++  {
              if i1 != i0 {
                Area1 = Area0 + pGate1.Area;
                if pMan.tAreaMax > 0.0 && Area1 > pMan.tAreaMax {
                    break; }
                pSupers[1] = pGate1;  uTruths[1][0] = pGate1.uTruth[0];  uTruths[1][1] = pGate1.uTruth[1];  ptPinDelays[1] = pGate1.ptDelays;

                // inlining of macro Super_ManForEachGate
                for  i2 = 0;                                                       i2 < nGatesLimit && (pGate2 = ppGatesLimit[i2]);                      i2++  {
                if i2 != i0 && i2 != i1 {
                  Area2 = Area1 + pGate2.Area;
                  if pMan.tAreaMax > 0.0 && Area2 > pMan.tAreaMax {
                      break; }
                  pSupers[2] = pGate2;  uTruths[2][0] = pGate2.uTruth[0];  uTruths[2][1] = pGate2.uTruth[1];   ptPinDelays[2] = pGate2.ptDelays;

                  // inlining of macro Super_ManForEachGate
                  for  i3 = 0;                                                       i3 < nGatesLimit && (pGate3 = ppGatesLimit[i3]);                      i3++  {
                  if i3 != i0 && i3 != i1 && i3 != i2 {
                    Area3 = Area2 + pGate3.Area;
                    if pMan.tAreaMax > 0.0 && Area3 > pMan.tAreaMax {
                        break; }
                    pSupers[3] = pGate3;   uTruths[3][0] = pGate3.uTruth[0];  uTruths[3][1] = pGate3.uTruth[1];   ptPinDelays[3] = pGate3.ptDelays;

                    // inlining of macro Super_ManForEachGate
                    for  i4 = 0;                                                       i4 < nGatesLimit && (pGate4 = ppGatesLimit[i4]);                      i4++  {
                    if i4 != i0 && i4 != i1 && i4 != i2 && i4 != i3 {
                      if fTimeOut != 0 { break; }
                      fTimeOut = Super_CheckTimeout( pProgress, pMan );
                      // compute area
                      Area4 = Area3 + pGate4.Area;
                      if pMan.tAreaMax > 0.0 && Area > pMan.tAreaMax {
                          break; }
                      pSupers[4] = pGate4;   uTruths[4][0] = pGate4.uTruth[0];  uTruths[4][1] = pGate4.uTruth[1];  ptPinDelays[4] = pGate4.ptDelays;

                      // inlining of macro Super_ManForEachGate
                      for  i5 = 0;                                                       i5 < nGatesLimit && (pGate5 = ppGatesLimit[i5]);                      i5++  {
                      if i5 != i0 && i5 != i1 && i5 != i2 && i5 != i3 && i5 != i4 {
                        if fTimeOut != 0 { goto done; }
                        fTimeOut = Super_CheckTimeout( pProgress, pMan );
                        // compute area
                        Area = Area4 + pGate5.Area;
                        if pMan.tAreaMax > 0.0 && Area > pMan.tAreaMax {
                            break; }
                        pSupers[5] = pGate5;   uTruths[5][0] = pGate5.uTruth[0];  uTruths[5][1] = pGate5.uTruth[1];  ptPinDelays[5] = pGate5.ptDelays;

                        Mio_DeriveGateDelays( ppGates[k], ptPinDelays, nFanins, pMan.nVarsMax, SUPER_NO_VAR, tPinDelaysRes, &tPinDelayMax );
                        Mio_DeriveTruthTable( ppGates[k], uTruths, nFanins, pMan.nVarsMax, uTruth );
                        if !(Super_CompareGates( pMan, uTruth, Area, tPinDelaysRes, pMan.nVarsMax ) != 0) {
                            continue; }
                        // create a new gate
                        pGateNew = Super_CreateGateNew( pMan, ppGates[k], pSupers, nFanins, uTruth, Area, tPinDelaysRes, tPinDelayMax, pMan.nVarsMax );
                        Super_AddGateToTable( pMan, pGateNew );
                        if (nGatesMax != 0) && pMan.nClasses > nGatesMax {
                            goto done; }
                      }
                      }
                    }
                    }
                  }
                  }
                }
                }
              }
              }
            }
            break;
        default :
            assert(false, "0");
            break;
        }
    }
done: 
    Extra_ProgressBarStop( pProgress );
    ABC_FREE( ppGatesLimit );
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Transfers gates from table into the array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_CheckTimeout(pPro *ProgressBar, pMan *Super_Man_t) int32 {
    var TimeNow abctime  = Abc_Clock();
    if TimeNow > pMan.TimePrint {
        Extra_ProgressBarUpdate( pPro, ++pMan.TimeSec, nil );
        pMan.TimePrint = Abc_Clock() + int64(time.Second);
    }
    if pMan.TimeStop && TimeNow > pMan.TimeStop {
        fmt.Printf("Timeout!\n");
        return 1;
    }
    pMan.nTried++;
    return 0;
}


/**Function*************************************************************

  Synopsis    [Transfers gates from table into the array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_TranferGatesToArray(pMan *Super_Man_t) {
    var gen *stmm_generator;
    var pGate, pList *Super_Gate_t;
    var Key ABC_PTRUINT_T;

    // put the gates fron the table into the array
    ABC_FREE( pMan.pGates );
    pMan.pGates = ABC_ALLOC( Super_Gate_t *, pMan.nAdded );
    pMan.nGates = 0;
    stmm_foreach_item( pMan.tTable, gen, (char **)&Key, (char **)&pList )
    {
        for  pGate = pList; (pGate != nil); pGate = pGate.pNext  {
            pMan.pGates[ pMan.nGates++ ] = pGate;
        }
    }
//    assert( pMan->nGates == pMan->nAdded - pMan->nRemoved );
}

/**Function*************************************************************

  Synopsis    [Adds one supergate into the unique table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_AddGateToTable(pMan *Super_Man_t, pGate *Super_Gate_t) {
    var ppList **Super_Gate_t;
    var Key ABC_PTRUINT_T;
//    Key = pGate->uTruth[0] + 2003 * pGate->uTruth[1];
    Key = pGate.uTruth[0] ^ pGate.uTruth[1];
    if !stmm_find_or_add( pMan.tTable, (char *)Key, (char ***)&ppList ) {
        *ppList = nil;
        pMan.nClasses++;
    }
    pGate.pNext = *ppList;
    *ppList = pGate;
    pMan.nAdded++;
}

/**Function*************************************************************

  Synopsis    [Check the manager's unique table for comparable gates.]

  Description [Returns 0 if the gate is dominated by others. Returns 1 
  if the gate is new or is better than the available ones. In this case, 
  cleans the table by removing the gates that are worse than the given one.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_CompareGates(pMan *Super_Man_t, uTruth []uint32, Area float32, tPinDelaysRes []float32, nPins int32) int32 {
    var ppList **Super_Gate_t;
    var pPrev *Super_Gate_t;
    var pGate *Super_Gate_t;
    var pGate2 *Super_Gate_t;
    var i, fNewIsBetter, fGateIsBetter int32;
    var Key ABC_PTRUINT_T;

    // skip constant functions
    if pMan.nVarsMax < 6 {
        if uTruth[0] == 0 || ~uTruth[0] == 0 {
            return 0; }
    
    } else {
        if ( uTruth[0] == 0 && uTruth[1] == 0 ) || ( ~uTruth[0] == 0 && ~uTruth[1] == 0 ) {
            return 0; }
    }

    // get hold of the place where the entry is stored
//    Key = uTruth[0] + 2003 * uTruth[1];
    Key = uTruth[0] ^ uTruth[1];
    if !stmm_find( pMan.tTable, (char *)Key, (char ***)&ppList ) {
        return 1; } 
    // the entry with this truth table is found
    pPrev = nil;
    for  pGate = *ppList, pGate2 = pGate? pGate.pNext: nil; (pGate != nil); 
        pGate = pGate2, pGate2 = pGate? pGate.pNext: nil  {
        pMan.nLookups++;
        if pGate.uTruth[0] != uTruth[0] || pGate.uTruth[1] != uTruth[1] {
            pMan.nAliases++;
            continue;
        }
        fGateIsBetter = 0;
        fNewIsBetter  = 0;
        if pGate.Area + SUPER_EPSILON < Area {
            fGateIsBetter = 1;
        } else if pGate.Area > Area + SUPER_EPSILON {
            fNewIsBetter = 1; }
        for  i = 0; i < nPins; i++  {
            if pGate.ptDelays[i] == SUPER_NO_VAR || tPinDelaysRes[i] == SUPER_NO_VAR {
                continue; }
            if pGate.ptDelays[i] + SUPER_EPSILON < tPinDelaysRes[i] {
                fGateIsBetter = 1;
            } else if pGate.ptDelays[i] > tPinDelaysRes[i] + SUPER_EPSILON {
                fNewIsBetter = 1; }
            if (fGateIsBetter != 0) && fNewIsBetter {
                break; }
        }
        // consider 4 cases
        if (fGateIsBetter != 0) && fNewIsBetter { // Pareto points; save both
            pPrev = pGate;
        } else if fNewIsBetter { // gate is worse; remove the gate
        
            if pPrev == nil {
                *ppList = pGate.pNext;
            } else {
                pPrev.pNext = pGate.pNext; }
            Extra_MmFixedEntryRecycle( pMan.pMem, (char *)pGate );
            pMan.nRemoved++;
        
        } else if fGateIsBetter != 0 { // new is worse, already dominated no need to see others
            return 0;
        } else { // if ( !fGateIsBetter && !fNewIsBetter ) // they are identical, no need to see others
            return 0; }
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Create a new supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_CreateGateNew(pMan *Super_Man_t, pRoot *Mio_Gate_t, pSupers **Super_Gate_t, nSupers int32, uTruth []uint32, Area float32, tPinDelaysRes []float32, tDelayMax float32, nPins int32) *Super_Gate_t {
    var pSuper *Super_Gate_t;
    pSuper = (Super_Gate_t *)Extra_MmFixedEntryFetch( pMan.pMem );
    *pSuper = Super_Gate_t{};
    pSuper.pRoot     = pRoot;
    pSuper.uTruth[0] = uTruth[0];
    pSuper.uTruth[1] = uTruth[1];
    memcpy( pSuper.ptDelays, tPinDelaysRes, sizeof(float) * nPins );
    pSuper.Area      = Area;
    pSuper.nFanins   = nSupers;
    memcpy( pSuper.pFanins, pSupers, sizeof(Super_Gate_t *) * nSupers );
    pSuper.pNext     = nil;
    pSuper.tDelayMax = tDelayMax;
    return pSuper;
}

/**Function*************************************************************

  Synopsis    [Starts the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_ManStart() *Super_Man_t {
    var pMan *Super_Man_t;
    pMan = new(Super_Man_t);
    pMan.pMem     = Extra_MmFixedStart( sizeof(Super_Gate_t) );
    pMan.tTable   = stmm_init_table( st__ptrcmp, st__ptrhash );
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Stops the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_ManStop(pMan *Super_Man_t) {
    Extra_MmFixedStop( pMan.pMem );
    if pMan.tTable { stmm_free_table( pMan.tTable ); }
    ABC_FREE( pMan.pGates );
    ABC_FREE( pMan );
}





/**Function*************************************************************

  Synopsis    [Writes the supergate library into the file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_Write(pMan *Super_Man_t) *Vec_Str_t {
    var vStr *Vec_Str_t;
    var pGateRoot, pGate *Super_Gate_t;
    var gen *stmm_generator;
    var fZeroFound, v int32;
    var clk abctime;
    var Key ABC_PTRUINT_T;

    if pMan.nGates < 1 {
        fmt.Printf("The generated library is empty. No output file written.\n");
        return nil;
    }

    // Filters the supergates by removing those that have fewer inputs than 
    // the given limit, provided that the inputs are not consequtive. 
    // For example, NAND2(a,c) is removed, but NAND2(a,b) is left, 
    // because a and b are consequtive.
    ABC_FREE( pMan.pGates );
    pMan.pGates = ABC_ALLOC( Super_Gate_t *, pMan.nAdded );
    pMan.nGates = 0;
    stmm_foreach_item( pMan.tTable, gen, (char **)&Key, (char **)&pGateRoot )
    {
        for  pGate = pGateRoot; (pGate != nil); pGate = pGate.pNext  {
            // skip the elementary variables
            if pGate.pRoot == nil {
                continue; }
            // skip the non-consequtive gates
            fZeroFound = 0;
            for  v = 0; v < pMan.nVarsMax; v++  {
                if pGate.ptDelays[v] < SUPER_NO_VAR + SUPER_EPSILON {
                    fZeroFound = 1;
                } else if fZeroFound {
                    break; }
            }
            if v < pMan.nVarsMax {
                continue; }
            // save the unique gate
            pMan.pGates[ pMan.nGates++ ] = pGate;
        }
    }

clk = Abc_Clock();
    // sort the supergates by truth table
    qsort( (void *)pMan.pGates, (size_t)pMan.nGates, sizeof(Super_Gate_t *), 
                (int (*)(void *, void *)) Super_WriteCompare );
    assert(Super_WriteCompare( pMan.pGates, pMan.pGates + pMan.nGates - 1 ) <= 0, "Super_WriteCompare( pMan.pGates, pMan.pGates + pMan.nGates - 1 ) <= 0");
if pMan.fVerbose {
ABC_PRT( "Sorting", Abc_Clock() - clk );
}


    // write library in the old format
clk = Abc_Clock();
    if pMan.fWriteOldFormat {
        Super_WriteLibrary( pMan ); }
if pMan.fVerbose {
ABC_PRT( "Writing old format", Abc_Clock() - clk );
}

    // write the tree-like structure of supergates
clk = Abc_Clock();
    vStr = Super_WriteLibraryTreeStr( pMan );
if pMan.fVerbose {
ABC_PRT( "Writing new format", Abc_Clock() - clk );
}
    return vStr;
}


/**Function*************************************************************

  Synopsis    [Writes the file header.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_WriteFileHeader(pMan *Super_Man_t, pFile *os.File) {
    fmt.Fprintf(pFile, "#\n");
    fmt.Fprintf(pFile, "# Supergate library derived for \"%s\" on %s.\n", pMan.pName, Extra_TimeStamp());
    fmt.Fprintf(pFile, "#\n");
    fmt.Fprintf(pFile, "# Command line: \"super -I %d -L %d -N %d -T %d -D %.2f -A %.2f %s %s\".\n", pMan.nVarsMax, pMan.nLevels, pMan.nGatesMax, pMan.TimeLimit, pMan.tDelayMax, pMan.tAreaMax, (pMan.fSkipInv? "" : "-s"), pMan.pName);
    fmt.Fprintf(pFile, "#\n");
    fmt.Fprintf(pFile, "# The number of inputs      = %10d.\n", pMan.nVarsMax);
    fmt.Fprintf(pFile, "# The number of levels      = %10d.\n", pMan.nLevels);
    fmt.Fprintf(pFile, "# The maximum delay         = %10.2f.\n", pMan.tDelayMax);
    fmt.Fprintf(pFile, "# The maximum area          = %10.2f.\n", pMan.tAreaMax);
    fmt.Fprintf(pFile, "# The maximum runtime (sec) = %10d.\n", pMan.TimeLimit);
    fmt.Fprintf(pFile, "#\n");
    fmt.Fprintf(pFile, "# The number of attempts    = %10d.\n", pMan.nTried);
    fmt.Fprintf(pFile, "# The number of supergates  = %10d.\n", pMan.nGates);
    fmt.Fprintf(pFile, "# The number of functions   = %10d.\n", pMan.nUnique);
    fmt.Fprintf(pFile, "# The total functions       = %.0f (2^%d).\n", math.Pow(float64(float64(2)), float64(pMan.nMints)), pMan.nMints);
    fmt.Fprintf(pFile, "#\n");
    fmt.Fprintf(pFile, "# Generation time           = %10.2f sec.\n", float32((pMan.Time))/float32((int64(time.Second))));
    fmt.Fprintf(pFile, "#\n");
    fmt.Fprintf(pFile, "%s\n", pMan.pName);
    fmt.Fprintf(pFile, "%d\n", pMan.nVarsMax);
    fmt.Fprintf(pFile, "%d\n", pMan.nGates);
}
func Super_WriteFileHeaderStr(pMan *Super_Man_t, vStr *Vec_Str_t) {
    var pBuffer string // was [1000]byte in C

    pBuffer = fmt.Sprintf("#\n");
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# Supergate library derived for \"%s\" on %s.\n", pMan.pName, Extra_TimeStamp());
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("#\n");
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# Command line: \"super -I %d -L %d -N %d -T %d -D %.2f -A %.2f %s %s\".\n", pMan.nVarsMax, pMan.nLevels, pMan.nGatesMax, pMan.TimeLimit, pMan.tDelayMax, pMan.tAreaMax, (pMan.fSkipInv? "" : "-s"), pMan.pName);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("#\n");
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# The number of inputs      = %10d.\n", pMan.nVarsMax);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# The number of levels      = %10d.\n", pMan.nLevels);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# The maximum delay         = %10.2f.\n", pMan.tDelayMax);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# The maximum area          = %10.2f.\n", pMan.tAreaMax);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# The maximum runtime (sec) = %10d.\n", pMan.TimeLimit);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("#\n");
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# The number of attempts    = %10d.\n", pMan.nTried);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# The number of supergates  = %10d.\n", pMan.nGates);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# The number of functions   = %10d.\n", pMan.nUnique);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# The total functions       = %.0f (2^%d).\n", math.Pow(float64(float64(2)), float64(pMan.nMints)), pMan.nMints);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("#\n");
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("# Generation time           = %10.2f sec.\n", float32((pMan.Time))/float32((int64(time.Second))));
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("#\n");
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("%s\n", pMan.pName);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("%d\n", pMan.nVarsMax);
    Vec_StrPrintStr( vStr, pBuffer );
    pBuffer = fmt.Sprintf("%d\n", pMan.nGates);
    Vec_StrPrintStr( vStr, pBuffer );
}

/**Function*************************************************************

  Synopsis    [Compares two gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_WriteCompare(ppG1 **Super_Gate_t, ppG2 **Super_Gate_t) int32 {
    var pTruth1 *uint32  = (*ppG1).uTruth;
    var pTruth2 *uint32  = (*ppG2).uTruth;
    if pTruth1[1] < pTruth2[1] {
        return -1; }
    if pTruth1[1] > pTruth2[1] {
        return 1; }
    if pTruth1[0] < pTruth2[0] {
        return -1; }
    if pTruth1[0] > pTruth2[0] {
        return 1; }
    return 0;
}
func Super_DelayCompare(ppG1 **Super_Gate_t, ppG2 **Super_Gate_t) int32 {
    if (*ppG1).tDelayMax < (*ppG2).tDelayMax {
        return -1; }
    if (*ppG1).tDelayMax > (*ppG2).tDelayMax {
        return 1; }
    return 0;
}
func Super_AreaCompare(ppG1 **Super_Gate_t, ppG2 **Super_Gate_t) int32 {
    if (*ppG1).Area < (*ppG2).Area {
        return -1; }
    if (*ppG1).Area > (*ppG2).Area {
        return 1; }
    return 0;
}






/**Function*************************************************************

  Synopsis    [Writes the gates into the file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_WriteLibraryGateName_rec(pGate *Super_Gate_t, pBuffer string) {
    var Buffer string // was [10]byte in C

    var i int32;

    if pGate.pRoot == nil {
        Buffer = fmt.Sprintf("%c", 'a' + pGate.Number);
        strcat( pBuffer, Buffer );
        return;
    }
    strcat( pBuffer, Mio_GateReadName(pGate.pRoot) );
    strcat( pBuffer, "(" );
    for  i = 0; i < int32(pGate.nFanins); i++  {
        if i != 0 {
            strcat( pBuffer, "," ); }
        Super_WriteLibraryGateName_rec( pGate.pFanins[i], pBuffer );
    }
    strcat( pBuffer, ")" );
}
var spStatic_Super_WriteLibraryGateName_Buffer [2000]byte;
func Super_WriteLibraryGateName(pGate *Super_Gate_t) string {
    
    spStatic_Super_WriteLibraryGateName_Buffer[0] = 0;
    Super_WriteLibraryGateName_rec( pGate, spStatic_Super_WriteLibraryGateName_Buffer );
    return spStatic_Super_WriteLibraryGateName_Buffer;
}
func Super_WriteLibraryGate(pFile *os.File, pMan *Super_Man_t, pGate *Super_Gate_t, Num int32) {
    var i int32;
    fmt.Fprintf(pFile, "%04d  ", Num);                         // the number
    Extra_PrintBinary( pFile, pGate.uTruth, pMan.nMints ); // the truth table
    fmt.Fprintf(pFile, "   %5.2f", pGate.tDelayMax);          // the max delay
    fmt.Fprintf(pFile, "  ");                                  
    for  i = 0; i < pMan.nVarsMax; i++  {                   // the pin-to-pin delays
        fmt.Fprintf(pFile, " %5.2f", pGate.ptDelays[i]==SUPER_NO_VAR? 0.0 : pGate.ptDelays[i]);
    }  
    fmt.Fprintf(pFile, "   %5.2f", pGate.Area);               // the area
    fmt.Fprintf(pFile, "   ");
    fmt.Fprintf(pFile, "%s", Super_WriteLibraryGateName(pGate));      // the symbolic expression
    fmt.Fprintf(pFile, "\n");
}
func Super_WriteLibrary(pMan *Super_Man_t) {
    var pGate, pGateNext *Super_Gate_t;
    var pFile *os.File;
    var FileName string;
    var pNameGeneric string;
    var i, Counter int32;

    FileName = ABC_ALLOC( char, 10000 );

    // get the file name
    pNameGeneric = Extra_FileNameGeneric( pMan.pName );
    FileName = fmt.Sprintf("%s.super_old", pNameGeneric);
    ABC_FREE( pNameGeneric );

    // count the number of unique functions
    pMan.nUnique = 1;
    // inlining of macro Super_ManForEachGate
    for  i = 0;                                                       i < pMan.nGates && (pGate = pMan.pGates[i]);                      i++  {
        if i == pMan.nGates - 1 {
            break; }
        // print the newline if this gate is different from the following one
        pGateNext = pMan.pGates[i+1];
        if pGateNext.uTruth[0] != pGate.uTruth[0] || pGateNext.uTruth[1] != pGate.uTruth[1] {
            pMan.nUnique++; }
    }

    // start the file
    pFile = os.OpenFile(FileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    Super_WriteFileHeader( pMan, pFile );

    // print the gates
    Counter = 0;
    // inlining of macro Super_ManForEachGate
    for  i = 0;                                                       i < pMan.nGates && (pGate = pMan.pGates[i]);                      i++  {
        Super_WriteLibraryGate( pFile, pMan, pGate, ++Counter );
        if i == pMan.nGates - 1 {
            break; }
        // print the newline if this gate is different from the following one
        pGateNext = pMan.pGates[i+1];
        if pGateNext.uTruth[0] != pGate.uTruth[0] || pGateNext.uTruth[1] != pGate.uTruth[1] {
            fmt.Fprintf(pFile, "\n"); }
    }
    assert(Counter == pMan.nGates, "Counter == pMan.nGates");
    _ = (pFile).Close();

    if pMan.fVerbose {
        fmt.Printf("The supergates are written using old format \"%s\" ", FileName);
        fmt.Printf("(%0.3f MB).\n", (float64(Extra_FileSize(FileName)))/(1<<20));
    }

    ABC_FREE( FileName );
}




/**Function*************************************************************

  Synopsis    [Recursively writes the gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_WriteLibraryTreeFile_rec(pFile *os.File, pMan *Super_Man_t, pSuper *Super_Gate_t, pCounter *int32) {
    var nFanins, i int32;
    // skip an elementary variable and a gate that was already written
    if pSuper.fVar || pSuper.Number > 0 {
        return; }
    // write the fanins
    nFanins = Mio_GateReadPinNum(pSuper.pRoot);
    for  i = 0; i < nFanins; i++  {
        Super_WriteLibraryTreeFile_rec( pFile, pMan, pSuper.pFanins[i], pCounter );
    }
    // finally write the gate
    pSuper.Number = (*pCounter)++;
    fmt.Fprintf(pFile, "%s", pSuper.fSuper? "* " : "");
    fmt.Fprintf(pFile, "%s", Mio_GateReadName(pSuper.pRoot));
    for  i = 0; i < nFanins; i++  {
        fmt.Fprintf(pFile, " %d", pSuper.pFanins[i].Number);
    }
    // write the formula 
    // this step is optional, the resulting library will work in any case
    // however, it may be helpful to for debugging to compare the same library 
    // written in the old format and written in the new format with formulas
//    fprintf( pFile, "    # %s", Super_WriteLibraryGateName( pSuper ) );
    fmt.Fprintf(pFile, "\n");
}
func Super_WriteLibraryTreeFile(pMan *Super_Man_t) {
    var pSuper *Super_Gate_t;
    var pFile *os.File;
    var FileName string;
    var pNameGeneric string;
    var i, Counter int32;
    var posStart int32;

    FileName = ABC_ALLOC( char, 10000 );

    // get the file name
    pNameGeneric = Extra_FileNameGeneric( pMan.pName );
    FileName = fmt.Sprintf("%s.super", pNameGeneric);
    ABC_FREE( pNameGeneric );
 
    // write the elementary variables
    pFile = os.OpenFile(FileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    Super_WriteFileHeader( pMan, pFile );
    // write the place holder for the number of lines
    posStart = ftell( pFile );
    fmt.Fprintf(pFile, "         \n");
    // mark the real supergates
    // inlining of macro Super_ManForEachGate
    for  i = 0;                                                       i < pMan.nGates && (pSuper = pMan.pGates[i]);                      i++  {
        pSuper.fSuper = 1;
    }
    // write the supergates
    Counter = pMan.nVarsMax;
    // inlining of macro Super_ManForEachGate
    for  i = 0;                                                       i < pMan.nGates && (pSuper = pMan.pGates[i]);                      i++  {
        Super_WriteLibraryTreeFile_rec( pFile, pMan, pSuper, &Counter );
    }
    _ = (pFile).Close();
    // write the number of lines
    pFile = os.OpenFile(FileName, os.O_RDWR, 0666);
    fseek( pFile, posStart, SEEK_SET );
    fmt.Fprintf(pFile, "%d", Counter);
    _ = (pFile).Close();

if pMan.fVerbose {
    fmt.Printf("The supergates are written using new format \"%s\" ", FileName);
    fmt.Printf("(%0.3f MB).\n", (float64(Extra_FileSize(FileName)))/(1<<20));
}

    ABC_FREE( FileName );
}


/**Function*************************************************************

  Synopsis    [Recursively writes the gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Super_WriteLibraryTreeStr_rec(vStr *Vec_Str_t, pMan *Super_Man_t, pSuper *Super_Gate_t, pCounter *int32) {
    var nFanins, i int32;
    // skip an elementary variable and a gate that was already written
    if pSuper.fVar || pSuper.Number > 0 {
        return; }
    // write the fanins
    nFanins = Mio_GateReadPinNum(pSuper.pRoot);
    for  i = 0; i < nFanins; i++  {
        Super_WriteLibraryTreeStr_rec( vStr, pMan, pSuper.pFanins[i], pCounter );
    }
    // finally write the gate
    pSuper.Number = (*pCounter)++;
//    fprintf( pFile, "%s", pSuper->fSuper? "* " : "" );
//    fprintf( pFile, "%s", Mio_GateReadName(pSuper->pRoot) );
//    for ( i = 0; i < nFanins; i++ )
//        fprintf( pFile, " %d", pSuper->pFanins[i]->Number );
    Vec_StrPrintStr( vStr, pSuper.fSuper? "* " : "" );
    Vec_StrPrintStr( vStr, Mio_GateReadName(pSuper.pRoot) );
    for  i = 0; i < nFanins; i++  {
        Vec_StrPrintStr( vStr, " " );
        Vec_StrPrintNum( vStr, pSuper.pFanins[i].Number );    
    }
    // write the formula 
    // this step is optional, the resulting library will work in any case
    // however, it may be helpful to for debugging to compare the same library 
    // written in the old format and written in the new format with formulas
//    fprintf( pFile, "    # %s", Super_WriteLibraryGateName( pSuper ) );
//    fprintf( pFile, "\n" );
    Vec_StrPrintStr( vStr, "\n" );
}
func Super_WriteLibraryTreeStr(pMan *Super_Man_t) *Vec_Str_t {
    var pInsert string // was [16]byte in C

    var vStr *Vec_Str_t;
    var pSuper *Super_Gate_t;
    var i, Counter int32;
    var posStart int32;
     // write the elementary variables
    vStr = Vec_StrAlloc( 1000 );
    Super_WriteFileHeaderStr( pMan, vStr );
    // write the place holder for the number of lines
    posStart = Vec_StrSize( vStr );
    for  i = 0; i < 9; i++  {
        Vec_StrPush( vStr, ' ' );
    }
    Vec_StrPush( vStr, '\n' );
    // mark the real supergates
    // inlining of macro Super_ManForEachGate
    for  i = 0;                                                       i < pMan.nGates && (pSuper = pMan.pGates[i]);                      i++  {
        pSuper.fSuper = 1;
    }
    // write the supergates
    Counter = pMan.nVarsMax;
    // inlining of macro Super_ManForEachGate
    for  i = 0;                                                       i < pMan.nGates && (pSuper = pMan.pGates[i]);                      i++  {
        Super_WriteLibraryTreeStr_rec( vStr, pMan, pSuper, &Counter );
    }
    Vec_StrPush( vStr, 0 );
    // write the number of lines
    pInsert = fmt.Sprintf("%d", Counter);
    for  i = 0; i < int32(strlen(pInsert)); i++  {
        Vec_StrWriteEntry( vStr, posStart + i, pInsert[i] );
    }
    return vStr;
}
func Super_WriteLibraryTree(pMan *Super_Man_t) {
    var vStr *Vec_Str_t;
    var pFileName string  = Extra_FileNameGenericAppend( pMan.pName, ".super" );
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {     
        fmt.Printf("Cannot open output file \"%s\".\n", pFileName);
        return;
    }
    vStr = Super_WriteLibraryTreeStr( pMan );
    fwrite( Vec_StrArray(vStr), 1, Vec_StrSize(vStr), pFile );
    _ = (pFile).Close();
    Vec_StrFree( vStr );
    // report the result of writing
    if pMan.fVerbose {
        fmt.Printf("The supergates are written using new format \"%s\" ", pFileName);
        fmt.Printf("(%0.3f MB).\n", (float64(Extra_FileSize(pFileName)))/(1<<20));
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END


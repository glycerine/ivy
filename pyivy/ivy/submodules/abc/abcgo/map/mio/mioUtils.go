package mio

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [mioUtils.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [File reading/writing for technology mapping.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 8, 2003.]

  Revision    [$Id: mioUtils.c,v 1.6 2004/09/03 18:02:20 satrajit Exp $]

***********************************************************************/

//#include <math.h>
//#include "mioInt.h"
//#include "base/main/main.h"
//#include "exp.h"
//#include "misc/util/utilTruth.h"
//#include "opt/dau/dau.h"
//#include "misc/util/utilNam.h"
//#include "map/scl/sclLib.h"
//#include "map/scl/sclCon.h"

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
func Mio_LibraryDelete(pLib *Mio_Library_t) {
    var pGate, pGate2 *Mio_Gate_t;
    if pLib == nil {
        return; }
    Mio_LibraryMatchesStop( pLib );
    Mio_LibraryMatches2Stop( pLib );
    // free the bindings of nodes to gates from this library for all networks
    Abc_FrameUnmapAllNetworks( Abc_FrameGetGlobalFrame() );
    // free the library
    pLib.pName = nil;
    // inlining of macro Mio_LibraryForEachGateSafe
    for  pGate = Mio_LibraryReadGates(pLib),                             pGate2 = (pGate? Mio_GateReadNext(pGate): nil);                 (pGate != nil);                                                         pGate = pGate2,                                                 pGate2 = (pGate? Mio_GateReadNext(pGate): nil)  {
        Mio_GateDelete( pGate );
    }
    Mem_FlexStop( pLib.pMmFlex, 0 );
    Vec_StrFree( pLib.vCube );
    if pLib.tName2Gate {
        st__free_table( pLib.tName2Gate ); }
//    if ( pLib->dd )
//        Cudd_Quit( pLib->dd );
    pLib.ppGates0 = nil;
    pLib.ppGatesName = nil;
    pLib = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_GateDelete(pGate *Mio_Gate_t) {
    var pPin, pPin2 *Mio_Pin_t;
    if pGate.nInputs > 6 {
        pGate.pTruth = nil; }
    Vec_IntFreeP( &pGate.vExpr );
    pGate.pOutName = nil;
    pGate.pName = nil;
    pGate.pForm = nil;
//    if ( pGate->bFunc )
//        Cudd_RecursiveDeref( pGate->pLib->dd, pGate->bFunc );
    // inlining of macro Mio_GateForEachPinSafe
    for  pPin = Mio_GateReadPins(pGate),                                 pPin2 = (pPin? Mio_PinReadNext(pPin): nil);                     (pPin != nil);                                                          pPin = pPin2,                                                   pPin2 = (pPin? Mio_PinReadNext(pPin): nil)  {
        Mio_PinDelete( pPin );
    }   
    pGate = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_PinDelete(pPin *Mio_Pin_t) {
    pPin.pName = nil;
    pPin = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_PinDup(pPin *Mio_Pin_t) *Mio_Pin_t {
    var pPinNew *Mio_Pin_t;

    // inlining of macro ABC_ALLOC
    pPinNew = ((Mio_Pin_t *) malloc(sizeof(Mio_Pin_t) * (size_t)(1)));
    *pPinNew = *pPin;
    if pPinNew.pName {
        pPinNew.pName = Abc_UtilStrsav(pPinNew.pName);
    } else {
        pPinNew.pName = nil;
    }
    pPinNew.pNext = nil;

    return pPinNew;
}




/**Function*************************************************************

  Synopsis    [Check if pin characteristics are the same.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_CheckPins(pPin1 *Mio_Pin_t, pPin2 *Mio_Pin_t) int32 {
    if pPin1 == nil || pPin2 == nil {
        return 1; }
    if pPin1.dLoadInput != pPin2.dLoadInput {
        return 0; }
    if pPin1.dLoadMax != pPin2.dLoadMax {
        return 0; }
    if pPin1.dDelayBlockRise != pPin2.dDelayBlockRise {
        return 0; }
    if pPin1.dDelayFanoutRise != pPin2.dDelayFanoutRise {
        return 0; }
    if pPin1.dDelayBlockFall != pPin2.dDelayBlockFall {
        return 0; }
    if pPin1.dDelayFanoutFall != pPin2.dDelayFanoutFall {
        return 0; }
    return 1;
}
func Mio_CheckGates(pLib *Mio_Library_t) int32 {
    var pGate *Mio_Gate_t;
    var pPin0 *Mio_Pin_t = nil;
    var pPin *Mio_Pin_t = nil;
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        // inlining of macro Mio_GateForEachPin
        for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
            if Mio_CheckPins( pPin0, pPin ) != 0 {
                pPin0 = pPin;
            } else {
                return 0; }
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_WritePin(pFile *os.File, pPin *Mio_Pin_t, NameLen int32, fAllPins int32) {
    var pPhaseNames [10]string = [10]string{ "UNKNOWN", "INV", "NONINV" };
    if fAllPins != 0 {
        fmt.Fprintf(pFile, "PIN *  ");
    } else {
        fmt.Fprintf(pFile, "\n    PIN %*s  ", NameLen, pPin.pName); }
    fmt.Fprintf(pFile, "%7s ", pPhaseNames[pPin.Phase]);
    fmt.Fprintf(pFile, "%3d ", int32(pPin.dLoadInput));
    fmt.Fprintf(pFile, "%3d ", int32(pPin.dLoadMax));
    fmt.Fprintf(pFile, "%8.2f ", pPin.dDelayBlockRise);
    fmt.Fprintf(pFile, "%8.2f ", pPin.dDelayFanoutRise);
    fmt.Fprintf(pFile, "%8.2f ", pPin.dDelayBlockFall);
    fmt.Fprintf(pFile, "%8.2f", pPin.dDelayFanoutFall);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_WriteGate(pFile *os.File, pGate *Mio_Gate_t, GateLen int32, NameLen int32, FormLen int32, fPrintSops int32, fAllPins int32) {
    var Buffer string // was [5000]byte in C

    var pPin *Mio_Pin_t;
    assert(NameLen+FormLen+2 < 5000, "NameLen+FormLen+2 < 5000");
    Buffer = fmt.Sprintf("%s=%s;", pGate.pOutName, pGate.pForm);
    fmt.Fprintf(pFile, "GATE %-*s ", GateLen, pGate.pName);
    fmt.Fprintf(pFile, "%8.2f  ", pGate.dArea);
    fprintf( pFile, "%-*s ",      Abc_MinInt(NameLen+FormLen+2, 60), Buffer );
    // print the pins
    if fPrintSops != 0 {
        fmt.Fprintf(pFile, "%s", pGate.pSop? pGate.pSop : "unspecified\n"); }
    if (fAllPins != 0) && pGate.pPins { // equal pins
        Mio_WritePin( pFile, pGate.pPins, NameLen, 1 );
    } else { // different pins
        // inlining of macro Mio_GateForEachPin
        for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
            Mio_WritePin( pFile, pPin, NameLen, 0 );
        } }
    fmt.Fprintf(pFile, "\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_WriteLibrary(pFile *os.File, pLib *Mio_Library_t, fPrintSops int32, fShort int32, fSelected int32) {
    var pGate *Mio_Gate_t;
    var pPin *Mio_Pin_t;
    var vGates *Vec_Ptr_t  = Vec_PtrAlloc( 1000 );
    var i, nCells, GateLen, NameLen, FormLen int32;
    var fAllPins int32  = fShort || Mio_CheckGates( pLib );
    if fSelected != 0 {
        var pCells *Mio_Cell2_t  = Mio_CollectRootsNewDefault2( 6, &nCells, 0 );
        for  i = 0; i < nCells; i++  {
            Vec_PtrPush( vGates, pCells[i].pMioGate );
        }
        pCells = nil;
    
    } else {
        for  i = 0; i < pLib.nGates; i++  {
            Vec_PtrPush( vGates, pLib.ppGates0[i] );
        }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGates)) && (((pGate) = (Mio_Gate_t *)Vec_PtrEntry(vGates, i)), 1); i++  {
        GateLen = Abc_MaxInt( GateLen, len(pGate.pName) );
        NameLen = Abc_MaxInt( NameLen, len(pGate.pOutName) );
        FormLen = Abc_MaxInt( FormLen, len(pGate.pForm) );
        // inlining of macro Mio_GateForEachPin
        for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
            NameLen = Abc_MaxInt( NameLen, len(pPin.pName) );
        }
    }
    fmt.Fprintf(pFile, "# The genlib library \"%s\" with %d gates written by ABC on %s\n", pLib.pName, Vec_PtrSize(vGates), Extra_TimeStamp());
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGates)) && (((pGate) = (Mio_Gate_t *)Vec_PtrEntry(vGates, i)), 1); i++  {
        Mio_WriteGate( pFile, pGate, GateLen, NameLen, FormLen, fPrintSops, fAllPins );
    }
    Vec_PtrFree( vGates );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Exp_PrintNodeVerilog(pFile *os.File, nVars int32, p *Vec_Int_t, vNames *Vec_Ptr_t, Node int32, fCompl int32) {
    extern void Exp_PrintLitVerilog( FILE * pFile, int nVars, Vec_Int_t * p, Vec_Ptr_t * vNames, int Lit );
    if Vec_IntEntry(p, 2*Node+1) >= 2*nVars {
        fmt.Fprintf(pFile, "("); }
    Exp_PrintLitVerilog( pFile, nVars, p, vNames, Vec_IntEntry(p, 2*Node+1) ^ fCompl );
    if Vec_IntEntry(p, 2*Node+1) >= 2*nVars {
        fmt.Fprintf(pFile, ")"); }
    fmt.Fprintf(pFile, " %c ", fCompl ? '|' : '&');
    if Vec_IntEntry(p, 2*Node+0) >= 2*nVars {
        fmt.Fprintf(pFile, "("); }
    Exp_PrintLitVerilog( pFile, nVars, p, vNames, Vec_IntEntry(p, 2*Node+0) ^ fCompl );
    if Vec_IntEntry(p, 2*Node+0) >= 2*nVars {
        fmt.Fprintf(pFile, ")"); }
}
func Exp_PrintLitVerilog(pFile *os.File, nVars int32, p *Vec_Int_t, vNames *Vec_Ptr_t, Lit int32) {
    if Lit == EXP_CONST0 {
        fmt.Fprintf(pFile, "1\'b0");
    } else if Lit == EXP_CONST1 {
        fmt.Fprintf(pFile, "1\'b1");
    } else if Lit < 2 * nVars {
        fmt.Fprintf(pFile, "%s%s", (Lit&1) ? "~" : "", (char *)Vec_PtrEntry(vNames, Lit/2));
    } else {
        Exp_PrintNodeVerilog( pFile, nVars, p, vNames, Lit/2-nVars, Lit&1 ); }
}
func Exp_PrintVerilog(pFile *os.File, nVars int32, p *Vec_Int_t, vNames *Vec_Ptr_t) {
    Exp_PrintLitVerilog( pFile, nVars, p, vNames, Vec_IntEntryLast(p) );
}
func Mio_WriteGateVerilog(pFile *os.File, pGate *Mio_Gate_t, vNames *Vec_Ptr_t) {
    var pName string; var i int32;
    fmt.Fprintf(pFile, "module %s ( ", pGate.pName);
    fmt.Fprintf(pFile, "%s", pGate.pOutName);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNames)) && (((pName) = (char *)Vec_PtrEntry(vNames, i)), 1); i++  {
        fmt.Fprintf(pFile, ", %s", pName);
    }
    fmt.Fprintf(pFile, " );\n");
    fmt.Fprintf(pFile, "  output %s;\n", pGate.pOutName);
    if Vec_PtrSize(vNames) > 0 {
        fmt.Fprintf(pFile, "  input %s", (char *)Vec_PtrEntry(vNames, 0));
        // inlining of macro Vec_PtrForEachEntryStart
        for  i = 1; (i < Vec_PtrSize(vNames)) && (((pName) = (char *)Vec_PtrEntry(vNames, i)), 1); i++  {
            fmt.Fprintf(pFile, ", %s", pName);
        }
        fmt.Fprintf(pFile, ";\n");
    }
    fmt.Fprintf(pFile, "  assign %s = ", pGate.pOutName);
    Exp_PrintVerilog( pFile, Vec_PtrSize(vNames), pGate.vExpr, vNames );
    fmt.Fprintf(pFile, ";\n");
    fmt.Fprintf(pFile, "endmodule\n\n");
}
func Mio_WriteLibraryVerilog(pFile *os.File, pLib *Mio_Library_t, fPrintSops int32, fShort int32, fSelected int32) {
    var pGate *Mio_Gate_t;
    var pPin *Mio_Pin_t;
    var vGates *Vec_Ptr_t  = Vec_PtrAlloc( 1000 );
    var vNames *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
    var i, nCells int32;
    if fSelected != 0 {
        var pCells *Mio_Cell2_t  = Mio_CollectRootsNewDefault2( 6, &nCells, 0 );
        for  i = 0; i < nCells; i++  {
            Vec_PtrPush( vGates, pCells[i].pMioGate );
        }
        pCells = nil;
    
    } else {
        for  i = 0; i < pLib.nGates; i++  {
            Vec_PtrPush( vGates, pLib.ppGates0[i] );
        }
    }
    fmt.Fprintf(pFile, "// Verilog for genlib library \"%s\" with %d gates written by ABC on %s\n\n", pLib.pName, Vec_PtrSize(vGates), Extra_TimeStamp());
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGates)) && (((pGate) = (Mio_Gate_t *)Vec_PtrEntry(vGates, i)), 1); i++  {
        Vec_PtrClear( vNames );
        // inlining of macro Mio_GateForEachPin
        for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
            Vec_PtrPush( vNames, pPin.pName );
        }
        Mio_WriteGateVerilog( pFile, pGate, vNames );
    }
    Vec_PtrFree( vNames );
    Vec_PtrFree( vGates );
}

/**Function*************************************************************

  Synopsis    [Compares the max delay of two gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_DelayCompare(ppG1 **Mio_Gate_t, ppG2 **Mio_Gate_t) int32 {
    var Comp int32;
    var Eps float32  = float32(0.0094636);
    if (*ppG1).dDelayMax < (*ppG2).dDelayMax - Eps {
        return -1; }
    if (*ppG1).dDelayMax > (*ppG2).dDelayMax + Eps {
        return 1; }
    // compare names
    Comp = int32(strings.Compare((*ppG1).pName, (*ppG2).pName));
    if Comp < 0 {
        return -1; }
    if Comp > 0 {
        return 1; }
    assert(false, "0");
    return 0;
}
func Mio_AreaCompare(pG1 *Mio_Cell_t, pG2 *Mio_Cell_t) int32 {
    var Comp int32;
    var Eps float32  = float32(0.0094636);
    if pG1.nFanins < pG2.nFanins {
        return -1; }
    if pG1.nFanins > pG2.nFanins {
        return 1; }
    if pG1.Area < pG2.Area - Eps {
        return -1; }
    if pG1.Area > pG2.Area + Eps {
        return 1; }
    // compare names
    Comp = int32(strings.Compare(pG1.pName, pG2.pName));
    if Comp < 0 {
        return -1; }
    if Comp > 0 {
        return 1; }
    assert(false, "0");
    return 0;
}
func Mio_AreaCompare2(pG1 *Mio_Cell2_t, pG2 *Mio_Cell2_t) int32 {
    var Comp int32;
    if pG1.nFanins < pG2.nFanins {
        return -1; }
    if pG1.nFanins > pG2.nFanins {
        return 1; }
    if pG1.AreaW < pG2.AreaW {
        return -1; }
    if pG1.AreaW > pG2.AreaW {
        return 1; }
    // compare names
    Comp = int32(strings.Compare(pG1.pName, pG2.pName));
    if Comp < 0 {
        return -1; }
    if Comp > 0 {
        return 1; }
    assert(false, "0");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Collects the set of root gates.]

  Description [Only collects the gates with unique functionality, 
  which have fewer inputs and shorter delay than the given limits.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_CellDelayAve(pCell *Mio_Cell_t) float32 {
    var CellDelay float32; var k int32;
    for  k = 0; k < int32(pCell.nFanins); k++  {
        CellDelay += pCell.Delays[k];
    }
    if pCell.nFanins {
        CellDelay /= pCell.nFanins; }
    return CellDelay;
}
func Mio_GateDelayAve(pGate *Mio_Gate_t) float32 {
    var GateDelay float32;
    var pPin *Mio_Pin_t; 
    // inlining of macro Mio_GateForEachPin
    for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
        GateDelay += float32((0.5 * pPin.dDelayBlockRise + 0.5 * pPin.dDelayBlockFall));
    }
    if pGate.nInputs {
        GateDelay /= pGate.nInputs; }
    return GateDelay;
}
func Mio_CompareTwoGates(pCell *Mio_Gate_t, pGate *Mio_Gate_t) int32 {
    var Comp int32;
    var Eps float32  = float32(0.0094636);
    var CellDelay, GateDelay float32;
    // compare areas
    if pCell.dArea > float32(pGate.dArea) + Eps {
        return 1; }
    if pCell.dArea < float32(pGate.dArea) - Eps {
        return 0; }
    // compare delays
    CellDelay = Mio_GateDelayAve( pCell );
    GateDelay = Mio_GateDelayAve( pGate );
    if CellDelay > GateDelay + Eps {
        return 1; }
    if CellDelay < GateDelay - Eps {
        return 0; }
    // compare names
    Comp = int32(strings.Compare(pCell.pName, pGate.pName));
    if Comp > 0 {
        return 1; }
    if Comp < 0 {
        return 0; }
    assert(false, "0");
    return 0;
}
func Mio_CollectRoots(pLib *Mio_Library_t, nInputs int32, tDelay float32, fSkipInv int32, pnGates *int32, fVerbose int32) **Mio_Gate_t {
    var pGate *Mio_Gate_t;
    var ppGates **Mio_Gate_t;
    var i, nGates, iGate, fProfile int32;
    nGates = Mio_LibraryReadGateNum( pLib );
    // inlining of macro ABC_ALLOC
    ppGates = ((Mio_Gate_t * *) malloc(sizeof(Mio_Gate_t *) * (size_t)(nGates)));
    iGate = 0;
    // check if profile is entered
    fProfile = Mio_LibraryHasProfile( pLib );
    if fProfile != 0 {
        fmt.Printf("Mio_CollectRoots(): Using gate profile to select gates for mapping.\n"); }
    // for each functionality, select gate with the smallest area
    // if equal areas, select gate with lexicographically smaller name
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        if pGate.nInputs > nInputs {
            continue; }
        if (fProfile != 0) && Mio_GateReadProfile(pGate) == 0 && pGate.nInputs > 1 {
            continue; }
        if tDelay > 0.0 && pGate.dDelayMax > float64(tDelay) {
            continue; }
        if pGate.uTruth == 0 || pGate.uTruth == ~(word)0 {
            continue; }
        if pGate.uTruth == ABC_CONST(0xAAAAAAAAAAAAAAAA) {
            continue; }
        if pGate.uTruth == ~ABC_CONST(0xAAAAAAAAAAAAAAAA) && (fSkipInv != 0) {
            continue; }
        if pGate.pTwin { // skip multi-output gates for now
            continue; }
        // check if the gate with this functionality already exists
        for  i = 0; i < iGate; i++  {
            if ppGates[i].uTruth == pGate.uTruth {
                if Mio_CompareTwoGates(ppGates[i], pGate) != 0 {
                    ppGates[i] = pGate; }
                break;
            }
        }
        if i < iGate {
            continue; }
        assert(iGate < nGates, "iGate < nGates");
        ppGates[ iGate++ ] = pGate;
        if fVerbose != 0 {
            fmt.Printf("Selected gate %3d:   %-20s  A = %7.2f  D = %7.2f  %3s = %-s\n", iGate+1, pGate.pName, pGate.dArea, pGate.dDelayMax, pGate.pOutName, pGate.pForm); }
    }
    // sort by delay
    if iGate > 0 {
        qsort( (void *)ppGates, (size_t)iGate, sizeof(Mio_Gate_t *), 
                        (int (*)(void *, void *)) Mio_DelayCompare );
        assert(Mio_DelayCompare( ppGates, ppGates + iGate - 1 ) <= 0, "Mio_DelayCompare( ppGates, ppGates + iGate - 1 ) <= 0");
    }
    if pnGates != nil {
        *pnGates = iGate; }
    return ppGates;
}


/**Function*************************************************************

  Synopsis    [Collects the set of root gates.]

  Description [Only collects the gates with unique functionality, 
  which have fewer inputs and shorter delay than the given limits.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_CompareTwo(pCell *Mio_Cell_t, pGate *Mio_Gate_t) int32 {
    var Comp int32;
    var Eps float32  = float32(0.0094636);
    var CellDelay, GateDelay float32;
    // compare areas
    if pCell.Area > float32(pGate.dArea) + Eps {
        return 1; }
    if pCell.Area < float32(pGate.dArea) - Eps {
        return 0; }
    // compare delays
    CellDelay = Mio_CellDelayAve( pCell );
    GateDelay = Mio_GateDelayAve( pGate );
    if CellDelay > GateDelay + Eps {
        return 1; }
    if CellDelay < GateDelay - Eps {
        return 0; }
    // compare names
    Comp = int32(strings.Compare(pCell.pName, pGate.pName));
    if Comp > 0 {
        return 1; }
    if Comp < 0 {
        return 0; }
    assert(false, "0");
    return 0;
}
func Mio_CollectCopy(pCell *Mio_Cell_t, pGate *Mio_Gate_t) {
    var pPin *Mio_Pin_t;  var k int32;
    pCell.pName   = pGate.pName;
    pCell.uTruth  = pGate.uTruth;
    pCell.Area    = float32(pGate.dArea);
    pCell.nFanins = pGate.nInputs;
    for  k, pPin = 0, pGate.pPins; (pPin != nil); pPin, k = pPin.pNext, k+1  {
        pCell.Delays[k] = float32((0.5 * pPin.dDelayBlockRise + 0.5 * pPin.dDelayBlockFall));
    }
}

func Mio_CollectRootsNew(pLib *Mio_Library_t, nInputs int32, pnGates *int32, fVerbose int32) *Mio_Cell_t {
    var pGate *Mio_Gate_t;
    var ppCells *Mio_Cell_t;
    var i int32;
    var nGates int32;
    var iCell int32 = 4;
    nGates = Mio_LibraryReadGateNum( pLib );
    // inlining of macro ABC_CALLOC
    ppCells = ((Mio_Cell_t *) calloc((size_t)(nGates + 4), sizeof(Mio_Cell_t)));
    // for each functionality, select gate with the smallest area
    // if equal areas, select gate with smaller average pin delay
    // if these are also equal, select lexicographically smaller name
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        if pGate.nInputs > nInputs || pGate.pTwin { // skip large and multi-output
            continue; }
        // check if the gate with this functionality already exists
        for  i = 0; i < iCell; i++  {
            if ppCells[i].pName && ppCells[i].uTruth == pGate.uTruth {
                if Mio_CompareTwo( ppCells + i, pGate ) != 0 {
                    Mio_CollectCopy( ppCells + i, pGate ); }
                break;
            }
        }
        if i < iCell {
            continue; }
        if pGate.uTruth == 0 || pGate.uTruth == ~(word)0 {
            var Idx int32  = int32((pGate.uTruth == ~(word)0));
            assert(pGate.nInputs == 0, "pGate.nInputs == 0");
            Mio_CollectCopy( ppCells + Idx, pGate );
            continue;
        }
        if pGate.uTruth == ABC_CONST(0xAAAAAAAAAAAAAAAA) || pGate.uTruth == ~ABC_CONST(0xAAAAAAAAAAAAAAAA) {
            var Idx int32  = 2 + int32((pGate.uTruth == ~ABC_CONST(0xAAAAAAAAAAAAAAAA)));
            assert(pGate.nInputs == 1, "pGate.nInputs == 1");
            Mio_CollectCopy( ppCells + Idx, pGate );
            continue;
        }
        Mio_CollectCopy( ppCells + iCell++, pGate );
    }
    if ppCells[0].pName == nil { fmt.Printf("Error: Cannot find constant 0 gate in the library.\n"); return nil; }
    if ppCells[1].pName == nil { fmt.Printf("Error: Cannot find constant 1 gate in the library.\n"); return nil; }
    if ppCells[2].pName == nil { fmt.Printf("Error: Cannot find buffer gate in the library.\n");     return nil; }
    if ppCells[3].pName == nil { fmt.Printf("Error: Cannot find inverter gate in the library.\n");   return nil; }
    // sort by delay
    if iCell > 5 {
        qsort( (void *)(ppCells + 4), (size_t)(iCell - 4), sizeof(Mio_Cell_t), 
                        (int (*)(void *, void *)) Mio_AreaCompare );
        assert(Mio_AreaCompare( ppCells + 4, ppCells + iCell - 1 ) <= 0, "Mio_AreaCompare( ppCells + 4, ppCells + iCell - 1 ) <= 0");
    }
    // assign IDs
    for  i = 0; i < iCell; i++  {
        ppCells[i].Id = ppCells[i].pName ? i : -1;
    }

    // report
    if fVerbose != 0 {
        // count gates
        // inlining of macro ABC_CALLOC
        var pCounts *int32  = ((int *) calloc((size_t)(nGates + 4), sizeof(int)));
        // inlining of macro Mio_LibraryForEachGate
        for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
            if pGate.nInputs > nInputs || pGate.pTwin { // skip large and multi-output
                continue; }
            for  i = 0; i < iCell; i++  {
                if ppCells[i].pName && ppCells[i].uTruth == pGate.uTruth {
                    pCounts[i]++;
                    break;
                }
            }
            assert(i < iCell, "i < iCell");
        }
        for  i = 0; i < iCell; i++  {
            var pCell *Mio_Cell_t  = ppCells + i;
            fmt.Printf("%4d : ", i);
            if pCell.pName == nil {
                fmt.Printf("None\n");
            } else {
                fmt.Printf("%-20s   In = %d   N = %3d   A = %12.6f   D = %12.6f\n", pCell.pName, pCell.nFanins, pCounts[i], pCell.Area, Mio_CellDelayAve(pCell)); }
        }
        pCounts = nil;
    }
    if pnGates != nil {
        *pnGates = iCell; }
    return ppCells;
}
func Mio_CollectRootsNewDefault(nInputs int32, pnGates *int32, fVerbose int32) *Mio_Cell_t {
    return Mio_CollectRootsNew( (Mio_Library_t *)Abc_FrameReadLibGen(), nInputs, pnGates, fVerbose );
}

/**Function*************************************************************

  Synopsis    [Collects the set of root gates.]

  Description [Only collects the gates with unique functionality, 
  which have fewer inputs and shorter delay than the given limits.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_CompareTwo2(pCell1 *Mio_Cell2_t, pCell2 *Mio_Cell2_t) int32 {
    var Comp int32;
    // compare areas
    if pCell1.AreaW > pCell2.AreaW {
        return 1; }
    if pCell1.AreaW < pCell2.AreaW {
        return 0; }
    // compare delays
    if pCell1.iDelayAve > pCell2.iDelayAve {
        return 1; }
    if pCell1.iDelayAve < pCell2.iDelayAve {
        return 0; }
    // compare names
    Comp = int32(strings.Compare(pCell1.pName, pCell2.pName));
    if Comp > 0 {
        return 1; }
    if Comp < 0 {
        return 0; }
    assert(false, "0");
    return 0;
}
func Mio_CollectCopy2(pCell *Mio_Cell2_t, pGate *Mio_Gate_t) {
    var pPin *Mio_Pin_t;  var k int32;
    pCell.pName     = pGate.pName;
    pCell.vExpr     = pGate.vExpr;
    pCell.uTruth    = pGate.uTruth;
    pCell.AreaF     = pGate.dArea;
    pCell.AreaW     = (word)(SCL_NUM * pGate.dArea);
    pCell.nFanins   = pGate.nInputs;
    pCell.pMioGate  = pGate;
    pCell.iDelayAve = 0;
    for  k, pPin = 0, pGate.pPins; (pPin != nil); pPin, k = pPin.pNext, k+1  {
        pCell.iDelays[k] = int32((SCL_NUM/2 * pPin.dDelayBlockRise + SCL_NUM/2 * pPin.dDelayBlockFall));
        pCell.iDelayAve += pCell.iDelays[k];
    }
    if pCell.nFanins {
        pCell.iDelayAve /= pCell.nFanins; }
}

func Mio_CollectRootsNew2(pLib *Mio_Library_t, nInputs int32, pnGates *int32, fVerbose int32) *Mio_Cell2_t {
    var pGate0 *Mio_Gate_t;
    var ppCells0, ppCells, pCell *Mio_Cell2_t;
    var i int32;
    var nGates int32;
    var iCell0 int32;
    var iCell int32 = 4;
    // create space for new cells
    nGates = Mio_LibraryReadGateNum( pLib );
    // inlining of macro ABC_CALLOC
    ppCells  = ((Mio_Cell2_t *) calloc((size_t)(nGates + 4), sizeof(Mio_Cell2_t)));
    // copy all gates first
    // inlining of macro ABC_CALLOC
    ppCells0 = ((Mio_Cell2_t *) calloc((size_t)(nGates), sizeof(Mio_Cell2_t)));
    // inlining of macro Mio_LibraryForEachGate
    for  pGate0 = Mio_LibraryReadGates(pLib);                             (pGate0 != nil);                                                         pGate0 = Mio_GateReadNext(pGate0)  {
        if !(pGate0.nInputs > nInputs || pGate0.pTwin) { // skip large and multi-output
            Mio_CollectCopy2( ppCells0 + iCell0++, pGate0 ); }
    }
    assert(iCell0 <= nGates, "iCell0 <= nGates");
    // for each functionality, select gate with the smallest area
    // if equal areas, select gate with smaller average pin delay
    // if these are also equal, select lexicographically smaller name
    for  pCell = ppCells0; pCell < ppCells0 + iCell0; pCell++  {
        // check if the gate with this functionality already exists
        for  i = 0; i < iCell; i++  {
            if ppCells[i].pName && ppCells[i].uTruth == pCell.uTruth {
                if Mio_CompareTwo2( ppCells + i, pCell ) != 0 {
                    ppCells[i] = *pCell; }
                break;
            }
        }
        if i < iCell {
            continue; }
        if pCell.uTruth == 0 || pCell.uTruth == ~(word)0 {
            var Idx int32  = int32((pCell.uTruth == ~(word)0));
            assert(pCell.nFanins == 0, "pCell.nFanins == 0");
            ppCells[Idx] = *pCell;
            continue;
        }
        if pCell.uTruth == ABC_CONST(0xAAAAAAAAAAAAAAAA) || pCell.uTruth == ~ABC_CONST(0xAAAAAAAAAAAAAAAA) {
            var Idx int32  = 2 + int32((pCell.uTruth == ~ABC_CONST(0xAAAAAAAAAAAAAAAA)));
            assert(pCell.nFanins == 1, "pCell.nFanins == 1");
            ppCells[Idx] = *pCell;
            continue;
        }
        ppCells[iCell++] = *pCell;
    }
    ppCells0 = nil;
    if ppCells[0].pName == nil { fmt.Printf("Error: Cannot find constant 0 gate in the library.\n"); return nil; }
    if ppCells[1].pName == nil { fmt.Printf("Error: Cannot find constant 1 gate in the library.\n"); return nil; }
    if ppCells[2].pName == nil { fmt.Printf("Error: Cannot find buffer gate in the library.\n");     return nil; }
    if ppCells[3].pName == nil { fmt.Printf("Error: Cannot find inverter gate in the library.\n");   return nil; }
    // sort by delay
    if iCell > 5 {
        qsort( (void *)(ppCells + 4), (size_t)(iCell - 4), sizeof(Mio_Cell2_t), 
                        (int (*)(void *, void *)) Mio_AreaCompare2 );
        assert(Mio_AreaCompare2( ppCells + 4, ppCells + iCell - 1 ) <= 0, "Mio_AreaCompare2( ppCells + 4, ppCells + iCell - 1 ) <= 0");
    }
    // assign IDs
    // inlining of macro Mio_LibraryForEachGate
    for  pGate0 = Mio_LibraryReadGates(pLib);                             (pGate0 != nil);                                                         pGate0 = Mio_GateReadNext(pGate0)  {
        Mio_GateSetCell( pGate0, -1 );
    }
    for  i = 0; i < iCell; i++  {
        ppCells[i].Id = ppCells[i].pName ? i : -1;
        Mio_GateSetCell( (Mio_Gate_t *)ppCells[i].pMioGate, i );
    }

    // report
    if fVerbose != 0 {
        // count gates
        // inlining of macro ABC_CALLOC
        var pCounts *int32  = ((int *) calloc((size_t)(nGates + 4), sizeof(int)));
        // inlining of macro Mio_LibraryForEachGate
        for  pGate0 = Mio_LibraryReadGates(pLib);                             (pGate0 != nil);                                                         pGate0 = Mio_GateReadNext(pGate0)  {
            if pGate0.nInputs > nInputs || pGate0.pTwin { // skip large and multi-output
                continue; }
            for  i = 0; i < iCell; i++  {
                if ppCells[i].pName && ppCells[i].uTruth == pGate0.uTruth {
                    pCounts[i]++;
                    break;
                }
            }
            assert(i < iCell, "i < iCell");
        }
        for  i = 0; i < iCell; i++  {
            var pCell *Mio_Cell2_t  = ppCells + i;
            fmt.Printf("%4d : ", i);
            if pCell.pName == nil {
                fmt.Printf("None\n");
            } else {
                fmt.Printf("%-20s   In = %d   N = %3d   A = %12.6f   D = %12.6f\n", pCell.pName, pCell.nFanins, pCounts[i], pCell.AreaF, Scl_Int2Flt(pCell.iDelayAve)); }
        }
        pCounts = nil;
    }
    if pnGates != nil {
        *pnGates = iCell; }
    return ppCells;
}
func Mio_CollectRootsNewDefault2(nInputs int32, pnGates *int32, fVerbose int32) *Mio_Cell2_t {
    return Mio_CollectRootsNew2( (Mio_Library_t *)Abc_FrameReadLibGen(), nInputs, pnGates, fVerbose );
}

/**Function*************************************************************

  Synopsis    [Derives the truth table of the gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_CollectRootsNewDefault3(nInputs int32, pvNames **Vec_Ptr_t, pvTruths **Vec_Wrd_t) int32 {
    var pLib *Mio_Library_t  = (Mio_Library_t *)Abc_FrameReadLibGen();
    var i int32;
    var iGate int32;
    var nGates int32;
    if pLib != nil {
        nGates = Mio_LibraryReadGateNum( pLib );
    } else {
        nGates = 0;
    }
    var pGate0 *Mio_Gate_t;
    var ppGates **Mio_Gate_t; var pTruth *word;
    if pLib == nil {
        return 0; }
    // inlining of macro ABC_CALLOC
    ppGates = ((Mio_Gate_t * *) calloc((size_t)(nGates), sizeof(Mio_Gate_t *)));
    // inlining of macro Mio_LibraryForEachGate
    for  pGate0 = Mio_LibraryReadGates(pLib);                             (pGate0 != nil);                                                         pGate0 = Mio_GateReadNext(pGate0)  {
        ppGates[pGate0.Cell] = pGate0;
    }
    *pvNames  = Vec_PtrAlloc( nGates );
    *pvTruths = Vec_WrdStart( nGates * 4 );
    for  i = 0; i < nGates; i++  {
        pGate0 = ppGates[i];
        if pGate0.nInputs > nInputs || pGate0.pTwin { // skip large and multi-output
            continue; }
        Vec_PtrPush( *pvNames, pGate0.pName );
        pTruth = Vec_WrdEntryP( *pvTruths, iGate++*4 );
        if pGate0.nInputs <= 6 {
            pTruth[0] = pTruth[1] = pTruth[2] = pTruth[3] = pGate0.uTruth;
        } else if pGate0.nInputs == 7 {
            pTruth[0] = pTruth[2] = pGate0.pTruth[0];
            pTruth[1] = pTruth[3] = pGate0.pTruth[1];
        
        } else if pGate0.nInputs == 8 {
            memcpy( pTruth, pGate0.pTruth, (size_t)(4*sizeof(word)) ); }
    }
    assert(iGate == nGates, "iGate == nGates");
    assert(Vec_WrdEntry(*pvTruths,  0) ==        0, "Vec_WrdEntry(*pvTruths,  0) ==        0");
    assert(Vec_WrdEntry(*pvTruths,  4) == ~(word)0, "Vec_WrdEntry(*pvTruths,  4) == ~(word)0");
    assert(Vec_WrdEntry(*pvTruths,  8) ==  s_Truths6[0], "Vec_WrdEntry(*pvTruths,  8) ==  s_Truths6[0]");
    assert(Vec_WrdEntry(*pvTruths, 12) == ~s_Truths6[0], "Vec_WrdEntry(*pvTruths, 12) == ~s_Truths6[0]");
    ppGates = nil;
    return nGates;
}

/**Function*************************************************************

  Synopsis    [Derives the truth table of the gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Mio_DeriveTruthTable6_uTruths6 [6][2]uint32 = [6][2]uint32{
        { 0xAAAAAAAA, 0xAAAAAAAA },
        { 0xCCCCCCCC, 0xCCCCCCCC },
        { 0xF0F0F0F0, 0xF0F0F0F0 },
        { 0xFF00FF00, 0xFF00FF00 },
        { 0xFFFF0000, 0xFFFF0000 },
        { 0x00000000, 0xFFFFFFFF },
    };
func Mio_DeriveTruthTable6(pGate *Mio_Gate_t) word {
    
    union {
      var u [2]uint32;
      var w word;
    }
 uTruthRes;
    assert(pGate.nInputs <= 6, "pGate.nInputs <= 6");
    Mio_DeriveTruthTable( pGate, spStatic_Mio_DeriveTruthTable6_uTruths6, pGate.nInputs, 6, uTruthRes.u );
    return uTruthRes.w;
}

     

                                                                        

                                                                

                
               
                

                

                                                                        
                                                                                              
 
                            
                            

                                                                            
                                   
     
                                                                          
                                     
                                     
               
     

                                                                       
                                           
     
                                
                                
               
     

                                           
                                                                     
                                                                     

                                                                    
                                                                   
                                                                  
                                                                   
                                                                  
 

                                                                        

                                                    

                
               
                

                

                                                                        
                                                                                                                       
 
                                                                   
 

      

/**Function*************************************************************

  Synopsis    [Derives the truth table of the gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_DeriveTruthTable(pGate *Mio_Gate_t, uTruthsIn [][2]uint32, nSigns int32, nInputs int32, uTruthRes []uint32) {
    var uRes word;
    var uFanins [6]word;
    var i int32;
    assert(pGate.nInputs == nSigns, "pGate.nInputs == nSigns");
    for  i = 0; i < nSigns; i++  {
        uFanins[i] = (((word)uTruthsIn[i][1]) << 32) | (word)uTruthsIn[i][0];
    }
    uRes = Exp_Truth6( nSigns, pGate.vExpr, (word *)uFanins );
    uTruthRes[0] = uRes & 0xFFFFFFFF;
    uTruthRes[1] = uRes >> 32;
}

/**Function*************************************************************

  Synopsis    [Reads the number of variables in the cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_SopGetVarNum(pSop string) int32 {
    var pCur string;
    for  pCur = pSop; *pCur != '\n'; pCur++  {
        if *pCur == 0 {
            return -1; }
    }
    return pCur - pSop - 2;
}

/**Function*************************************************************

  Synopsis    [Derives the truth table of the root of the gate.]

  Description [Given the truth tables of the leaves of the gate,
  this procedure derives the truth table of the root.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_DeriveTruthTable2(pGate *Mio_Gate_t, uTruthsIn [][2]uint32, nTruths int32, nInputs int32, uTruthRes []uint32) {
    var uSignCube [2]uint32;
    var i, nFanins int32;
    var pCube string;

    // make sure that the number of input truth tables in equal to the number of gate inputs
    assert(pGate.nInputs == nTruths, "pGate.nInputs == nTruths");
    assert(nInputs < 7, "nInputs < 7");

    nFanins = Mio_SopGetVarNum( pGate.pSop );
    assert(nFanins == nInputs, "nFanins == nInputs");

    // clean the resulting truth table
    uTruthRes[0] = 0;
    uTruthRes[1] = 0;
    if nInputs < 6 {
//        Abc_SopForEachCube( pGate->pSop, nFanins, pCube )
        for  pCube = pGate.pSop; *pCube; pCube += (nFanins) + 3  {
            // add the clause
            uSignCube[0] = MIO_FULL;
            for  i = 0; i < nFanins; i++  {
                if pCube[i] == '0' {
                    uSignCube[0] &= ~uTruthsIn[i][0];
                } else if pCube[i] == '1' {
                    uSignCube[0] &=  uTruthsIn[i][0]; }
            }
        }
        if nInputs < 5 {
            // inlining of macro MIO_MASK
            uTruthRes[0] &= ((~(uint32(0))) >> (32-(1<<nInputs))); }
    
    } else {
        // consider the case when two unsigneds should be used
//        Abc_SopForEachCube( pGate->pSop, nFanins, pCube )
        for  pCube = pGate.pSop; *pCube; pCube += (nFanins) + 3  {
            uSignCube[0] = MIO_FULL;
            uSignCube[1] = MIO_FULL;
            for  i = 0; i < nFanins; i++  {
                if pCube[i] == '0' {
                    uSignCube[0] &= ~uTruthsIn[i][0];
                    uSignCube[1] &= ~uTruthsIn[i][1];
                
                } else if pCube[i] == '1' {
                    uSignCube[0] &=  uTruthsIn[i][0];
                    uSignCube[1] &=  uTruthsIn[i][1];
                }
            }
            uTruthRes[0] |= uSignCube[0];
            uTruthRes[1] |= uSignCube[1];
        }
    }
}

/**Function*************************************************************

  Synopsis    [Derives the area and delay of the root of the gate.]

  Description [Array of the resulting delays should be initialized 
  to the (negative) SUPER_NO_VAR value.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_DeriveGateDelays(pGate *Mio_Gate_t, ptPinDelays **float32, nPins int32, nInputs int32, tDelayZero float32, ptDelaysRes *float32, ptPinDelayMax *float32) {
    var pPin *Mio_Pin_t;
    var Delay, DelayMax float32;
    var i, k int32;
    assert(pGate.nInputs == nPins, "pGate.nInputs == nPins");
    // set all the delays to the unused delay
    for  i = 0; i < nInputs; i++  {
        ptDelaysRes[i] = tDelayZero;
    }
    // compute the delays for each input and the max delay at the same time
    DelayMax = 0;
    for  i = 0; i < nInputs; i++  {
        for  k, pPin = 0, pGate.pPins; (pPin != nil); pPin, k = pPin.pNext, k+1  {
            if ptPinDelays[k][i] < 0 {
                continue; }
            Delay = ptPinDelays[k][i] + float32(pPin.dDelayBlockMax);
            if ptDelaysRes[i] < Delay {
                ptDelaysRes[i] = Delay; }
        }
        if k != nPins {
            fmt.Printf("DEBUG: problem gate is %s\n", Mio_GateReadName( pGate ));
        }
        assert(k == nPins, "k == nPins");
        if DelayMax < ptDelaysRes[i] {
            DelayMax = ptDelaysRes[i]; }
    }
    *ptPinDelayMax = DelayMax;
}


/**Function*************************************************************

  Synopsis    [Creates a pseudo-gate.]

  Description [The pseudo-gate is a N-input gate with all info set to 0.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_GateCreatePseudo(nInputs int32) *Mio_Gate_t {
    var pGate *Mio_Gate_t;
    var pPin *Mio_Pin_t;
    var i int32;
    // allocate the gate structure
    // inlining of macro ABC_ALLOC
    pGate = ((Mio_Gate_t *) malloc(sizeof(Mio_Gate_t) * (size_t)(1)));
    memset( pGate, 0, sizeof(Mio_Gate_t) );
    pGate.nInputs = nInputs;
    // create pins
    for  i = 0; i < nInputs; i++  {
        // inlining of macro ABC_ALLOC
        pPin = ((Mio_Pin_t *) malloc(sizeof(Mio_Pin_t) * (size_t)(1)));
        memset( pPin, 0, sizeof(Mio_Pin_t) );
        pPin.pNext = pGate.pPins;
        pGate.pPins = pPin;
    }
    return pGate;
}

/**Function*************************************************************

  Synopsis    [Adds constant value to all delay values.]

  Description [The pseudo-gate is a N-input gate with all info set to 0.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryShiftDelay(pLib *Mio_Library_t, Shift float64) {
    var pGate *Mio_Gate_t;
    var pPin *Mio_Pin_t;
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        pGate.dDelayMax += Shift;
        // inlining of macro Mio_GateForEachPin
        for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
            pPin.dDelayBlockRise += Shift;
            pPin.dDelayBlockFall += Shift;
            pPin.dDelayBlockMax  += Shift;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Multiply areas/delays by values proportional to fanin count.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryMultiArea(pLib *Mio_Library_t, Multi float64) {
    var pGate *Mio_Gate_t;
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        if pGate.nInputs < 2 {
            continue; }
//        printf( "Before %8.3f  ", pGate->dArea );
        pGate.dArea *= math.Pow(float64(pGate.nInputs), float64(Multi));
//        printf( "After %8.3f  Inputs = %d. Factor = %8.3f\n", pGate->dArea, pGate->nInputs, pow( pGate->nInputs, Multi ) );
    }
}
func Mio_LibraryMultiDelay(pLib *Mio_Library_t, Multi float64) {
    var pGate *Mio_Gate_t;
    var pPin *Mio_Pin_t;
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        if pGate.nInputs < 2 {
            continue; }
//        printf( "Before %8.3f  ", pGate->dDelayMax );
        pGate.dDelayMax *= math.Pow(float64(pGate.nInputs), float64(Multi));
//        printf( "After %8.3f  Inputs = %d. Factor = %8.3f\n", pGate->dDelayMax, pGate->nInputs, pow( pGate->nInputs, Multi ) );
        // inlining of macro Mio_GateForEachPin
        for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
            pPin.dDelayBlockRise *= math.Pow(float64(pGate.nInputs), float64(Multi));
            pPin.dDelayBlockFall *= math.Pow(float64(pGate.nInputs), float64(Multi));
            pPin.dDelayBlockMax  *= math.Pow(float64(pGate.nInputs), float64(Multi));
        }
    }
}

/**Function*************************************************************

  Synopsis    [Transfers delays from the second to the first.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryTransferDelays(pLibD *Mio_Library_t, pLibS *Mio_Library_t) {
    var pGateD, pGateS *Mio_Gate_t;
    var pPinD, pPinS *Mio_Pin_t;
    // inlining of macro Mio_LibraryForEachGate
    for  pGateS = Mio_LibraryReadGates(pLibS);                             (pGateS != nil);                                                         pGateS = Mio_GateReadNext(pGateS)  {
        // inlining of macro Mio_LibraryForEachGate
        for  pGateD = Mio_LibraryReadGates(pLibD);                             pGateD;                                                         pGateD = Mio_GateReadNext(pGateD)  {
            if pGateD.uTruth != pGateS.uTruth {
                continue; }
            pPinS = Mio_GateReadPins( pGateS );
            // inlining of macro Mio_GateForEachPin
            for  pPinD = Mio_GateReadPins(pGateD);                                 pPinD;                                                          pPinD = Mio_PinReadNext(pPinD)  {
                if pPinS != nil {
                    pPinD.dDelayBlockRise = pPinS.dDelayBlockRise;
                    pPinD.dDelayBlockFall = pPinS.dDelayBlockFall;
                    pPinD.dDelayBlockMax  = pPinS.dDelayBlockMax;
                    pPinS = Mio_PinReadNext(pPinS);
                
                } else {
                    pPinD.dDelayBlockRise = 0;
                    pPinD.dDelayBlockFall = 0;
                    pPinD.dDelayBlockMax  = 0;
                }
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nf_ManPrepareGate(nVars int32, uTruth word, pComp *int32, pPerm *int32, vResult *Vec_Wrd_t) {
    var nPerms int32  = Extra_Factorial( nVars );
    var nMints int32  = (1 << nVars);
    var tCur, tTemp1, tTemp2 word;
    var i, p, c int32;
    Vec_WrdClear( vResult );
    for  i = 0; i < 2; i++  {
        if i != 0 {
            tCur = ~uTruth;
        } else {
            tCur = uTruth;
        }
        tTemp1 = tCur;
        for  p = 0; p < nPerms; p++  {
            tTemp2 = tCur;
            for  c = 0; c < nMints; c++  {
                Vec_WrdPush( vResult, tCur );
                tCur = Abc_Tt6Flip( tCur, pComp[c] );
            }
            assert(tTemp2 == tCur, "tTemp2 == tCur");
            tCur = Abc_Tt6SwapAdjacent( tCur, pPerm[p] );
        }
        assert(tTemp1 == tCur, "tTemp1 == tCur");
    }
}
func Nf_ManPreparePrint(nVars int32, pComp *int32, pPerm *int32, Line [][8]byte) {
    var nPerms int32  = Extra_Factorial( nVars );
    var nMints int32  = (1 << nVars);
    var pChar, pChar2 string;
    var i, p, c, n int32;
    for  i = 0; i < nVars; i++  {
        Line[0][i] = 'A' + nVars - 1 - i;
    }
    Line[0][nVars] = '+';
    Line[0][nVars+1] = 0;
    for  i = 0; i < 2; i++  {
        Line[n][nVars] = i ? '-' : '+';
        for  p = 0; p < nPerms; p++  {
            for  c = 0; c < nMints; c++  {
                strcpy( Line[n+1], Line[n] ); n++;
                pChar = &Line[n][pComp[c]];
                if *pChar >= 'A' && *pChar <= 'Z' {
                    *pChar += 'a' - 'A';
                } else if *pChar >= 'a' && *pChar <= 'z' {
                    *pChar -= 'a' - 'A'; }
            }
            pChar  = &Line[n][pPerm[p]];
            pChar2 = pChar + 1;
            // inlining of macro ABC_SWAP
            { var t byte  = *pChar; *pChar = *pChar2; *pChar2 = t; };
        }
    }
    assert(n == 2*nPerms*nMints, "n == 2*nPerms*nMints");
    n = 0;
    for  i = 0; i < 2; i++  {
        for  p = 0; p < nPerms; p++  {
            for  c = 0; c < nMints; c++  {
                fmt.Printf("%8d : %d %3d %2d : %s\n", n, i, p, c, Line[n]);
n++;
            }
        }
    }
}

func Nf_ManPrepareLibrary(pLib *Mio_Library_t) {
//    char Lines[2*720*64][8];
//    Nf_ManPreparePrint( 6, pComp, pPerm, Lines );
    var pComp [7]*int32;
    var pPerm [7]*int32;
    var ppGates **Mio_Gate_t;
    var vResult *Vec_Wrd_t;
    var pTruths *word;
    var pSizes *int32;
    var nGates, i, nClasses, nTotal int32;
    var clk abctime  = Abc_Clock();

    for  i = 2; i <= 6; i++  {
        pComp[i] = Extra_GreyCodeSchedule( i );
    }
    for  i = 2; i <= 6; i++  {
        pPerm[i] = Extra_PermSchedule( i );
    }

    // collect truth tables
    ppGates = Mio_CollectRoots( pLib, 6, float32(1.0e+20), 1, &nGates, 0 );
    // inlining of macro ABC_CALLOC
    pSizes  = ((int *) calloc((size_t)(nGates), sizeof(int)));
    // inlining of macro ABC_CALLOC
    pTruths = ((word *) calloc((size_t)(nGates), sizeof(word)));
    vResult = Vec_WrdAlloc( 2 * 720 * 64 );
    for  i = 0; i < nGates; i++  {
        pSizes[i] = Mio_GateReadPinNum( ppGates[i] );
        assert(pSizes[i] > 1 && pSizes[i] <= 6, "pSizes[i] > 1 && pSizes[i] <= 6");
        pTruths[i] = Mio_GateReadTruth( ppGates[i] );

        Nf_ManPrepareGate( pSizes[i], pTruths[i], pComp[pSizes[i]], pPerm[pSizes[i]], vResult );
        Vec_WrdUniqify(vResult);
        nClasses += Vec_WrdSize(vResult);
        nTotal = (1 << (pSizes[i]+1)) * Extra_Factorial(pSizes[i]);

        fmt.Printf("%6d : ", i);
        fmt.Printf("%16s : ", Mio_GateReadName( ppGates[i] ));
        fmt.Printf("%48s : ", Mio_GateReadForm( ppGates[i] ));
        fmt.Printf("Inputs = %2d   ", pSizes[i]);
        fmt.Printf("Total = %6d  ", nTotal);
        fmt.Printf("Classes = %6d ", Vec_WrdSize(vResult));
        fmt.Printf("Configs = %8.2f ", 1.0*nTotal/Vec_WrdSize(vResult));
        fmt.Printf("%6.2f %%  ", 100.0*Vec_WrdSize(vResult)/nTotal);
        Dau_DsdPrintFromTruth( &pTruths[i], pSizes[i] );
//        printf( "\n" );
    }
    Vec_WrdFree( vResult );
    ppGates = nil;
    pSizes = nil;
    pTruths = nil;

    for  i = 2; i <= 6; i++  {
        pComp[i] = nil;
    }
    for  i = 2; i <= 6; i++  {
        pPerm[i] = nil;
    }

    fmt.Printf("Classes = %d.  ", nClasses);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}

func Nf_ManPrepareLibraryTest2() {
    var pLib *Mio_Library_t  = (Mio_Library_t *)Abc_FrameReadLibGen();
    if pLib != nil {
        Nf_ManPrepareLibrary( pLib );
    } else {
        fmt.Printf("Standard cell library is not available.\n"); }

}

/**Function*************************************************************

  Synopsis    [Install library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryTransferCellIds() {
    var pGate *Mio_Gate_t;
    var pLib *Mio_Library_t  = (Mio_Library_t *)Abc_FrameReadLibGen();
    var pScl *SC_Lib  = (SC_Lib *)Abc_FrameReadLibScl();
    var CellId int32;
    if pScl == nil {
        fmt.Printf("SC library cannot be found.\n");
        return;
    }
    if pLib == nil {
        fmt.Printf("Genlib library cannot be found.\n");
        return;
    }
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        if Mio_GateReadPinNum(pGate) == 0 {
            continue; }
        CellId = Abc_SclCellFind( pScl, Mio_GateReadName(pGate) );
        if CellId < 0 {
            fmt.Printf("Cannot find cell ID of gate %s.\n", Mio_GateReadName(pGate));
        } else {
            Mio_GateSetCell( pGate, CellId ); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryReadProfile(pFile *os.File, pLib *Mio_Library_t) {
    var pGate *Mio_Gate_t;    
    var pBuffer [1000]byte;
    for  fgets( pBuffer, 1000, pFile ) != nil  {
        var pToken string  = strtok( pBuffer, " \t\n" );
        if pToken == nil {
            continue; }
        if pToken[0] == '#' {
            continue; }
        // read gate
        pGate = Mio_LibraryReadGateByName( pLib, pToken, nil );
        if pGate == nil {
            fmt.Printf("Cannot find gate \"%s\" in library \"%s\".\n", pToken, Mio_LibraryReadName(pLib));
            continue;
        }
        // read profile
        pToken = strtok( nil, " \t\n" );
        Mio_GateSetProfile( pGate, strconv.Atoi(pToken) );
    }
}

func Mio_LibraryWriteProfile(pFile *os.File, pLib *Mio_Library_t) {
    var pGate *Mio_Gate_t;    
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        if Mio_GateReadProfile(pGate) > 0 {
            fmt.Fprintf(pFile, "%-24s  %6d\n", Mio_GateReadName(pGate), Mio_GateReadProfile(pGate)); }
    }
}

func Mio_LibraryHasProfile(pLib *Mio_Library_t) int32 {
    var pGate *Mio_Gate_t;    
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        if Mio_GateReadProfile(pGate) > 0 {
            return 1; }
    }
    return 0;
}


func Mio_LibraryTransferProfile(pLibDst *Mio_Library_t, pLibSrc *Mio_Library_t) {
    var pGateSrc, pGateDst *Mio_Gate_t;    
    // inlining of macro Mio_LibraryForEachGate
    for  pGateDst = Mio_LibraryReadGates(pLibDst);                             (pGateDst != nil);                                                         pGateDst = Mio_GateReadNext(pGateDst)  {
        Mio_GateSetProfile( pGateDst, 0 );
    }
    // inlining of macro Mio_LibraryForEachGate
    for  pGateSrc = Mio_LibraryReadGates(pLibSrc);                             pGateSrc;                                                         pGateSrc = Mio_GateReadNext(pGateSrc)  {
        if Mio_GateReadProfile(pGateSrc) > 0 {
            // find gate by name
            pGateDst = Mio_LibraryReadGateByName( pLibDst, Mio_GateReadName(pGateSrc), nil );
            if pGateDst == nil {
                // find gate by function
                // inlining of macro Mio_LibraryForEachGate
                for  pGateDst = Mio_LibraryReadGates(pLibDst);                             (pGateDst != nil);                                                         pGateDst = Mio_GateReadNext(pGateDst)  {
                    if pGateDst.uTruth == pGateSrc.uTruth {
                        break; }
                }
                if pGateDst == nil {
                    fmt.Printf("Cannot find gate \"%s\" in library \"%s\".\n", Mio_GateReadName(pGateSrc), Mio_LibraryReadName(pLibDst));
                    continue;
                }
            }
            Mio_GateAddToProfile( pGateDst, Mio_GateReadProfile(pGateSrc) );
        }
    }
}
func Mio_LibraryTransferProfile2(pLibDst *Mio_Library_t, pLibSrc *Mio_Library_t) {
    var pGateSrc, pGateDst *Mio_Gate_t;    
    // inlining of macro Mio_LibraryForEachGate
    for  pGateDst = Mio_LibraryReadGates(pLibDst);                             (pGateDst != nil);                                                         pGateDst = Mio_GateReadNext(pGateDst)  {
        Mio_GateSetProfile2( pGateDst, 0 );
    }
    // inlining of macro Mio_LibraryForEachGate
    for  pGateSrc = Mio_LibraryReadGates(pLibSrc);                             pGateSrc;                                                         pGateSrc = Mio_GateReadNext(pGateSrc)  {
        if Mio_GateReadProfile2(pGateSrc) > 0 {
            // find gate by name
            pGateDst = Mio_LibraryReadGateByName( pLibDst, Mio_GateReadName(pGateSrc), nil );
            if pGateDst == nil {
                // find gate by function
                // inlining of macro Mio_LibraryForEachGate
                for  pGateDst = Mio_LibraryReadGates(pLibDst);                             (pGateDst != nil);                                                         pGateDst = Mio_GateReadNext(pGateDst)  {
                    if pGateDst.uTruth == pGateSrc.uTruth {
                        break; }
                }
                if pGateDst == nil {
                    fmt.Printf("Cannot find gate \"%s\" in library \"%s\".\n", Mio_GateReadName(pGateSrc), Mio_LibraryReadName(pLibDst));
                    continue;
                }
            }
            Mio_GateAddToProfile2( pGateDst, Mio_GateReadProfile2(pGateSrc) );
        }
    }
}

func Mio_LibraryCleanProfile2(pLib *Mio_Library_t) {
    var pGate *Mio_Gate_t;    
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        Mio_GateSetProfile2( pGate, 0 );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryHashGates(pLib *Mio_Library_t) {
    var pGate *Mio_Gate_t;    
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        if pGate.pTwin {
            fmt.Printf("Gates with multiple outputs are not supported.\n");
            return;
        }
    }
    if pLib.tName2Gate {
        st__free_table( pLib.tName2Gate ); }
    pLib.tName2Gate = st__init_table(st__strcmp, st__strhash);
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        st__insert( pLib.tName2Gate, pGate.pName, (char *)pGate );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclIsChar(c byte) int32 {
    return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_';
}
func Abc_SclIsName(c byte) int32 {
    return Abc_SclIsChar(c) || (c >= '0' && c <= '9');
}
func Abc_SclFindLimit(pName string) string {
    assert((Abc_SclIsChar(*pName) != 0), "Abc_SclIsChar(*pName)");
    for  Abc_SclIsName(*pName) != 0  {
        pName++;
    }
    return pName;
}
func Abc_SclAreEqual(pBase string, pName string, pLimit string) int32 {
    return !strncmp( pBase, pName, pLimit - pName );
}
func Mio_LibraryShortFormula(pCell *Mio_Gate_t, pForm string, pBuffer string) {
    var pPin *Mio_Pin_t; 
    var pTemp, pLimit string; var i int32;
    if !strncmp(pForm, "CONST", 5) {
        pBuffer = fmt.Sprintf("%s", pForm);
        return;
    }
    for  pTemp = pForm; *pTemp;  {
        if !(Abc_SclIsChar(*pTemp) != 0) {
            *pBuffer++ = *pTemp++;
            continue;
        }
        pLimit = Abc_SclFindLimit( pTemp );
        i = 0;
        // inlining of macro Mio_GateForEachPin
        for  pPin = Mio_GateReadPins(pCell);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
            if Abc_SclAreEqual( pPin.pName, pTemp, pLimit ) != 0 {
                *pBuffer++ = 'a' + i;
                break;
            }
            i++;
        }
        pTemp = pLimit;
    }
    *pBuffer++ = 0;
}
func Mio_LibraryShortNames(pLib *Mio_Library_t) {
    var Buffer string // was [10000]byte in C

    var pGate *Mio_Gate_t; var pPin *Mio_Pin_t;
    var c int32;
    var i int32;
    var nDigits int32 = Abc_Base10Log( Mio_LibraryReadGateNum(pLib) );
    // itereate through classes
    // inlining of macro Mio_LibraryForEachGate
    for  pGate = Mio_LibraryReadGates(pLib);                             (pGate != nil);                                                         pGate = Mio_GateReadNext(pGate)  {
        pGate.pName = nil;
        Buffer = fmt.Sprintf("g%0*d", nDigits, ++c);
        pGate.pName = Abc_UtilStrsav( Buffer );
        // update formula
        Mio_LibraryShortFormula( pGate, pGate.pForm, Buffer );
        pGate.pForm = nil;
        pGate.pForm = Abc_UtilStrsav( Buffer );
        // pin names
        i = 0;
        // inlining of macro Mio_GateForEachPin
        for  pPin = Mio_GateReadPins(pGate);                                 (pPin != nil);                                                          pPin = Mio_PinReadNext(pPin)  {
            pPin.pName = nil;
            Buffer = fmt.Sprintf("%c", 'a'+i);
            pPin.pName = Abc_UtilStrsav( Buffer );
            i++;
        }
        // output pin
        pGate.pOutName = nil;
        Buffer = fmt.Sprintf("z");
        pGate.pOutName = Abc_UtilStrsav( Buffer );
    }
    Mio_LibraryHashGates( pLib );
    // update library name
    fmt.Printf("Renaming library \"%s\" into \"%s%d\".\n", pLib.pName, "lib", Mio_LibraryReadGateNum(pLib));
    pLib.pName = nil;
    Buffer = fmt.Sprintf("lib%d", Mio_LibraryReadGateNum(pLib));
    pLib.pName = Abc_UtilStrsav( Buffer );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryMatchesStop(pLib *Mio_Library_t) {
    if !pLib.vTtMem {
        return; }
    Vec_WecFree( pLib.vTt2Match );
    Vec_MemHashFree( pLib.vTtMem );
    Vec_MemFree( pLib.vTtMem );
    pLib.pCells = nil;
}
func Mio_LibraryMatchesStart(pLib *Mio_Library_t, fPinFilter int32, fPinPerm int32, fPinQuick int32) {
    extern Mio_Cell2_t * Nf_StoDeriveMatches( Vec_Mem_t * vTtMem, Vec_Wec_t * vTt2Match, int * pnCells, int fPinFilter, int fPinPerm, int fPinQuick );
    if pLib.vTtMem && pLib.fPinFilter == fPinFilter && pLib.fPinPerm == fPinPerm && pLib.fPinQuick == fPinQuick {
        return; }
    if pLib.vTtMem {
        Mio_LibraryMatchesStop( pLib ); }
    pLib.fPinFilter = fPinFilter;  // pin filtering
    pLib.fPinPerm   = fPinPerm;    // pin permutation
    pLib.fPinQuick  = fPinQuick;   // pin permutation
    pLib.vTtMem     = Vec_MemAllocForTT( 6, 0 );          
    pLib.vTt2Match  = Vec_WecAlloc( 1000 ); 
    Vec_WecPushLevel( pLib.vTt2Match );
    Vec_WecPushLevel( pLib.vTt2Match );
    assert(Vec_WecSize(pLib.vTt2Match) == Vec_MemEntryNum(pLib.vTtMem), "Vec_WecSize(pLib.vTt2Match) == Vec_MemEntryNum(pLib.vTtMem)");
    pLib.pCells = Nf_StoDeriveMatches( pLib.vTtMem, pLib.vTt2Match, &pLib.nCells, fPinFilter, fPinPerm, fPinQuick );
}
func Mio_LibraryMatchesFetch(pLib *Mio_Library_t, pvTtMem **Vec_Mem_t, pvTt2Match **Vec_Wec_t, ppCells **Mio_Cell2_t, pnCells *int32, fPinFilter int32, fPinPerm int32, fPinQuick int32) {
    Mio_LibraryMatchesStart( pLib, fPinFilter, fPinPerm, fPinQuick );
    *pvTtMem    = pLib.vTtMem;     // truth tables
    *pvTt2Match = pLib.vTt2Match;  // matches for truth tables
    *ppCells    = pLib.pCells;     // library gates
    *pnCells    = pLib.nCells;     // library gate count
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mio_LibraryMatches2Stop(pLib *Mio_Library_t) {
    var i int32;
    if !pLib.vNames {
        return; }
    Vec_PtrFree( pLib.vNames );
    Vec_WrdFree( pLib.vTruths );
    Vec_IntFree( pLib.vTt2Match4 );
    Vec_IntFree( pLib.vConfigs );
    for  i = 0; i < 3; i++  {
        Vec_MemHashFree( pLib.vTtMem2[i] );
        Vec_MemFree( pLib.vTtMem2[i] );
        Vec_IntFree( pLib.vTt2Match2[i] );
    }
}
func Mio_LibraryMatches2Start(pLib *Mio_Library_t) {
    extern int Gia_ManDeriveMatches( Vec_Ptr_t ** pvNames, Vec_Wrd_t ** pvTruths, Vec_Int_t ** pvTt2Match4, Vec_Int_t ** pvConfigs, Vec_Mem_t * pvTtMem2[3], Vec_Int_t * pvTt2Match2[3] );
    if pLib.vNames {
        return; }
    if pLib.vTtMem {
        Mio_LibraryMatches2Stop( pLib ); }
    Gia_ManDeriveMatches( &pLib.vNames, &pLib.vTruths, &pLib.vTt2Match4, &pLib.vConfigs, pLib.vTtMem2, pLib.vTt2Match2 );
}
func Mio_LibraryMatches2Fetch(pLib *Mio_Library_t, pvNames **Vec_Ptr_t, pvTruths **Vec_Wrd_t, pvTt2Match4 **Vec_Int_t, pvConfigs **Vec_Int_t, pvTtMem2 []*Vec_Mem_t, pvTt2Match2 []*Vec_Int_t) {
    var i int32;
    Mio_LibraryMatches2Start( pLib );
    *pvNames     = pLib.vNames;    
    *pvTruths    = pLib.vTruths;    
    *pvTt2Match4 = pLib.vTt2Match4;    
    *pvConfigs   = pLib.vConfigs;    
    for  i = 0; i < 3; i++  {
        pvTtMem2[i]  = pLib.vTtMem2[i];    
        pvTt2Match2[i] = pLib.vTt2Match2[i];   
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

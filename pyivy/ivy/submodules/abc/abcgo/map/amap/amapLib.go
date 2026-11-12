package amap

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [amapLib.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Technology mapper for standard cells.]

  Synopsis    [Standard-cell library.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: amapLib.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "amapInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocs a library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibAlloc() *Amap_Lib_t {
    var p *Amap_Lib_t;
    // inlining of macro ABC_ALLOC
    p = (Amap_Lib_t *)((Amap_Lib_t *) malloc(sizeof(Amap_Lib_t) * (size_t)(1)));
    memset( p, 0, sizeof(Amap_Lib_t) );
    p.vGates = Vec_PtrAlloc( 100 );
    p.pMemGates = Aig_MmFlexStart();
    p.pMemSet = Aig_MmFlexStart();
    return p;
}

/**Function*************************************************************

  Synopsis    [Deallocs a library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibFree(p *Amap_Lib_t) {
    if p == nil {
        return; }
    if p.vSelect {
        Vec_PtrFree( p.vSelect ); }
    if p.vSorted {
        Vec_PtrFree( p.vSorted ); }
    if p.vGates {
        Vec_PtrFree( p.vGates ); }
    if p.vRules {
        Vec_VecFree( (Vec_Vec_t *)p.vRules ); }
    if p.vRulesX {
        Vec_VecFree( (Vec_Vec_t *)p.vRulesX ); }
    if p.vRules3 {
        Vec_IntFree( p.vRules3 ); }
    Aig_MmFlexStop( p.pMemGates, 0 );
    Aig_MmFlexStop( p.pMemSet, 0 );
    p.pRules = nil;
    p.pRulesX = nil;
    p.pNodes = nil;
    p.pName = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Returns the largest gate size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibNumPinsMax(p *Amap_Lib_t) int32 {
    var pGate *Amap_Gat_t;
    var i, Counter int32;
    // inlining of macro Amap_LibForEachGate
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vGates)) && (((pGate) = (Amap_Gat_t *)Vec_PtrEntry(p.vGates, i)), 1); i++  {
        if Counter < int32(pGate.nPins) {
            Counter = pGate.nPins; }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Writes one pin.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibWritePin(pFile *os.File, pPin *Amap_Pin_t) {
    var pPhaseNames [10]string = [10]string{ "UNKNOWN", "INV", "NONINV" };
    fmt.Fprintf(pFile, "    PIN ");
    fmt.Fprintf(pFile, "%9s ", pPin.pName);
    fmt.Fprintf(pFile, "%10s ", pPhaseNames[pPin.Phase]);
    fmt.Fprintf(pFile, "%6d ", int32(pPin.dLoadInput));
    fmt.Fprintf(pFile, "%6d ", int32(pPin.dLoadMax));
    fmt.Fprintf(pFile, "%6.2f ", pPin.dDelayBlockRise);
    fmt.Fprintf(pFile, "%6.2f ", pPin.dDelayFanoutRise);
    fmt.Fprintf(pFile, "%6.2f ", pPin.dDelayBlockFall);
    fmt.Fprintf(pFile, "%6.2f", pPin.dDelayFanoutFall);
    fmt.Fprintf(pFile, "\n");
}

/**Function*************************************************************

  Synopsis    [Writes one gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibWriteGate(pFile *os.File, pGate *Amap_Gat_t, fPrintDsd int32) {
    var pPin *Amap_Pin_t;
    fmt.Fprintf(pFile, "GATE ");
    fmt.Fprintf(pFile, "%12s ", pGate.pName);
    fmt.Fprintf(pFile, "%10.2f   ", pGate.dArea);
    fmt.Fprintf(pFile, "%s=%s;\n", pGate.pOutName, pGate.pForm);
    if fPrintDsd != 0 {
        if pGate.pFunc == nil {
            fmt.Printf("Truth table is not available.\n");
        } else {
            Kit_DsdPrintFromTruth( pGate.pFunc, pGate.nPins ); }
    }
    // inlining of macro Amap_GateForEachPin
    for  pPin = pGate.Pins; pPin < pGate.Pins + pGate.nPins; pPin++  {
        Amap_LibWritePin( pFile, pPin );
    }
}

/**Function*************************************************************

  Synopsis    [Writes library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibWrite(pFile *os.File, pLib *Amap_Lib_t, fPrintDsd int32) {
    var pGate *Amap_Gat_t;
    var i int32;
    fmt.Fprintf(pFile, "# The genlib library \"%s\".\n", pLib.pName);
    // inlining of macro Amap_LibForEachGate
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pLib.vGates)) && (((pGate) = (Amap_Gat_t *)Vec_PtrEntry(pLib.vGates, i)), 1); i++  {
        Amap_LibWriteGate( pFile, pGate, fPrintDsd );
    }
}
        
/**Function*************************************************************

  Synopsis    [Compares two gates by area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibCompareGatesByArea(pp1 **Amap_Gat_t, pp2 **Amap_Gat_t) int32 {
    var Diff float64  = (*pp1).dArea - (*pp2).dArea;
    if Diff < 0.0 {
        return -1; }
    if Diff > 0.0 { 
        return 1; }
    return 0; 
}
        
/**Function*************************************************************

  Synopsis    [Compares gates by area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibSortGatesByArea(pLib *Amap_Lib_t) *Vec_Ptr_t {
    var vSorted *Vec_Ptr_t;
    vSorted = Vec_PtrDup( pLib.vGates );
    qsort( (void *)Vec_PtrArray(vSorted), (size_t)Vec_PtrSize(vSorted), sizeof(void *), 
                (int (*)(void *, void *)) Amap_LibCompareGatesByArea );
    return vSorted;
}

/**Function*************************************************************

  Synopsis    [Finds min-area gate with the given function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibFindGate(p *Amap_Lib_t, uTruth uint32) *Amap_Gat_t {
    var pGate *Amap_Gat_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vSorted)) && (((pGate) = (Amap_Gat_t *)Vec_PtrEntry(p.vSorted, i)), 1); i++  {
        if ( pGate == nil ) || ( pGate.pFunc == nil ) { continue; }
        if pGate.nPins <= 5 && pGate.pFunc[0] == uTruth {
            return pGate; }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Selects gates useful for area-only mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibSelectGates(p *Amap_Lib_t, fVerbose int32) *Vec_Ptr_t {
    var vSelect *Vec_Ptr_t;
    var pGate, pGate2 *Amap_Gat_t;
    var i, k int32;//, clk = Abc_Clock();
    p.pGate0   = Amap_LibFindGate( p, 0 );
    p.pGate1   = Amap_LibFindGate( p, ~0 );
    p.pGateBuf = Amap_LibFindGate( p, 0xAAAAAAAA );
    p.pGateInv = Amap_LibFindGate( p, ~0xAAAAAAAA );
    vSelect = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vSorted)) && (((pGate) = (Amap_Gat_t *)Vec_PtrEntry(p.vSorted, i)), 1); i++  {
        if pGate.pFunc == nil || pGate.pTwin != nil {
            continue; }
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < i) && (((pGate2) = (Amap_Gat_t *)Vec_PtrEntry(p.vSorted, k)), 1); k++  {
            if pGate2.pFunc == nil || pGate2.pTwin != nil {
                continue; }
            if pGate2.nPins != pGate.nPins {
                continue; }
            if !memcmp( pGate2.pFunc, pGate.pFunc, sizeof(unsigned) * Abc_TruthWordNum(pGate.nPins) ) {
                break; }
        }
        if k < i {
            continue; }
        Vec_PtrPush( vSelect, pGate );            
    }
    return vSelect;
}

/**Function*************************************************************

  Synopsis    [Selects gates useful for area-only mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibPrintSelectedGates(p *Amap_Lib_t, fAllGates int32) {
    var vArray *Vec_Ptr_t;
    var pGate *Amap_Gat_t;
    var i int32;
    if fAllGates != 0 {
        vArray = p.vGates;
    } else {
        vArray = p.vSelect;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vArray)) && (((pGate) = (Amap_Gat_t *)Vec_PtrEntry(vArray, i)), 1); i++  {
        fmt.Printf("%3d :%12s %d %9.2f  ", i, pGate.pName, pGate.nPins, pGate.dArea);
        fmt.Printf("%4s=%40s  ", pGate.pOutName, pGate.pForm);
        fmt.Printf("DSD: ");
        Kit_DsdPrintFromTruth( pGate.pFunc, pGate.nPins );
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Parses equations for the gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_LibReadAndPrepare(pFileName string, pBuffer string, fVerbose int32, fVeryVerbose int32) *Amap_Lib_t {
    var p *Amap_Lib_t;
    var clk abctime  = Abc_Clock();
    if pBuffer == nil {
        p = Amap_LibReadFile( pFileName, fVerbose );
    } else {
        p = Amap_LibReadBuffer( pBuffer, fVerbose );
        if p != nil {
            p.pName = Abc_UtilStrsav( pFileName ); }
    }
    if fVerbose != 0 {
        fmt.Printf("Read %d gates from file \"%s\".\n", Vec_PtrSize(p.vGates), pFileName); }
    if p == nil {
        return nil; }
    if !Amap_LibParseEquations( p, fVerbose ) {
        Amap_LibFree( p );
        return nil;
    }
    p.vSorted = Amap_LibSortGatesByArea( p );
    p.vSelect = Amap_LibSelectGates( p, fVerbose );
    if fVerbose != 0 {
        fmt.Printf("Selected %d functionally unique gates. ", Vec_PtrSize(p.vSelect));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
//       Amap_LibPrintSelectedGates( p, 0 );
    }
    clk = Abc_Clock();
    Amap_LibCreateRules( p, fVeryVerbose );
    if fVerbose != 0 {
        fmt.Printf("Created %d rules and %d matches. ", p.nNodes, p.nSets);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    return p;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


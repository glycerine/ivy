package mpm

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [mpm.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Configurable technology mapper.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 1, 2013.]

  Revision    [$Id: mpm.c,v 1.00 2013/06/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mpmInt.h"

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
func Mpm_ManStart(pMig *Mig_Man_t, pPars *Mpm_Par_t) *Mpm_Man_t {
    
    var i int32;
    assert(sizeof(Mpm_Uni_t) % sizeof(word) == 0, "sizeof(Mpm_Uni_t) % sizeof(word) == 0");      // aligned info to word boundary
    assert(pPars.nNumCuts <= MPM_CUT_MAX, "pPars.nNumCuts <= MPM_CUT_MAX");
    assert(!pPars.fUseTruth || pPars.pLib.LutMax <= 16, "!pPars.fUseTruth || pPars.pLib.LutMax <= 16");
    assert(!pPars.fUseDsd || pPars.pLib.LutMax <= 6, "!pPars.fUseDsd || pPars.pLib.LutMax <= 6");
    Mig_ManSetRefs( pMig );
    // alloc
    // inlining of macro ABC_CALLOC
    p := &Mpm_Man_t{};
    p.pMig      = pMig;
    p.pPars     = pPars;
    p.pLibLut   = pPars.pLib;
    p.nLutSize  = pPars.pLib.LutMax;
    if pPars.fUseTruth {
        p.nTruWords = Abc_Truth6WordNum(p.nLutSize);
    } else {
        p.nTruWords = 0;
    }
    p.nNumCuts  = pPars.nNumCuts;
    // cuts
    assert(Mpm_CutWordNum(32) < 32, "Mpm_CutWordNum(32) < 32"); // using 5 bits for word count
    p.pManCuts  = Mmr_StepStart( 13, Abc_Base2Log(Mpm_CutWordNum(p.nLutSize) + 1) );
    Vec_PtrGrow( &p.vFreeUnits, p.nNumCuts + 1 );
    for  i = p.nNumCuts; i >= 0; i--  {
        Vec_PtrPush( &p.vFreeUnits, p.pCutUnits + i );
    }
    p.vTemp     = Vec_PtrAlloc( 1000 );
    // mapping attributes
    Vec_IntFill( &p.vCutBests, Mig_ManObjNum(pMig), 0 );
    Vec_IntFill( &p.vCutLists, Mig_ManObjNum(pMig), 0 );
    Vec_IntFill( &p.vMigRefs, Mig_ManObjNum(pMig), 0 );
    Vec_IntFill( &p.vMapRefs, Mig_ManObjNum(pMig), 0 );
    Vec_IntFill( &p.vEstRefs, Mig_ManObjNum(pMig), 0 );
    Vec_IntFill( &p.vRequireds, Mig_ManObjNum(pMig), ABC_INFINITY );
    Vec_IntFill( &p.vTimes, Mig_ManObjNum(pMig), 0 );
    Vec_IntFill( &p.vAreas, Mig_ManObjNum(pMig), 0 );
    Vec_IntFill( &p.vEdges, Mig_ManObjNum(pMig), 0 );
    // start DSD manager
    assert(!p.pPars.fUseTruth || !p.pPars.fUseDsd, "!p.pPars.fUseTruth || !p.pPars.fUseDsd");
    if p.pPars.fUseTruth { 
        p.vTtMem = Vec_MemAlloc( p.nTruWords, 12 ); // 32 KB/page for 6-var functions
        Vec_MemHashAlloc( p.vTtMem, 10000 );
        p.funcCst0 = Vec_MemHashInsert( p.vTtMem, p.Truth );
        Abc_TtUnit( p.Truth, p.nTruWords, 0 );
        p.funcVar0 = Vec_MemHashInsert( p.vTtMem, p.Truth );
    
    } else if p.pPars.fUseDsd {
        Mpm_ManPrecomputePerms( p );
        p.funcVar0 = 1;
    }
    // finish
    p.timeTotal = Abc_Clock();
    pMig.pMan = p;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManStop(p *Mpm_Man_t) {
    if p.pPars.fUseTruth && p.pPars.fVeryVerbose {
        var pFileName string  = "truths.txt";
        var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        Vec_MemDump( pFile, p.vTtMem );
        _ = (pFile).Close();
        fmt.Printf("Dumped %d %d-var truth tables into file \"%s\" (%.2f MB).\n", Vec_MemEntryNum(p.vTtMem), p.nLutSize, pFileName, (16.0 * p.nTruWords + 1.0) * Vec_MemEntryNum(p.vTtMem) / (1 << 20));
    }
    if p.pPars.fUseDsd && p.pPars.fVerbose {
        Mpm_ManPrintDsdStats( p ); }
    if p.vTtMem {
        Vec_MemHashFree( p.vTtMem );
        Vec_MemFree( p.vTtMem );
    }
    if p.pHash {
        Vec_WrdFree( p.vPerm6 );
        Vec_IntFree( p.vMap2Perm );
        Vec_IntFree( p.vConfgRes );
        Vec_IntFree( p.pHash.vData );
        Hsh_IntManStop( p.pHash );
    }
    Vec_WecFreeP( &p.vNpnConfigs );
    Vec_PtrFree( p.vTemp );
    Mmr_StepStop( p.pManCuts );
    p.vFreeUnits.pArray = nil;
    // mapping attributes
    p.vCutBests.pArray = nil;
    p.vCutLists.pArray = nil;
    p.vMigRefs.pArray = nil;
    p.vMapRefs.pArray = nil;
    p.vEstRefs.pArray = nil;
    p.vRequireds.pArray = nil;
    p.vTimes.pArray = nil;
    p.vAreas.pArray = nil;
    p.vEdges.pArray = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mpm_ManPrintStatsInit(p *Mpm_Man_t) {
    fmt.Printf("K = %d.  C = %d.  Cand = %d. XOR = %d. MUX = %d. Choice = %d.  CutMin = %d. Truth = %d. DSD = %d.\n", p.nLutSize, p.nNumCuts, Mig_ManCandNum(p.pMig), Mig_ManXorNum(p.pMig), Mig_ManMuxNum(p.pMig), p.pMig.nChoices, p.pPars.fCutMin, p.pPars.fUseTruth, p.pPars.fUseDsd);
}
func Mpm_ManPrintStats(p *Mpm_Man_t) {
    fmt.Printf("Memory usage:  Mig = %.2f MB  Map = %.2f MB  Cut = %.2f MB    Total = %.2f MB.  ", 1.0 * Mig_ManObjNum(p.pMig) * sizeof(Mig_Obj_t) / (1 << 20), 1.0 * Mig_ManObjNum(p.pMig) * 48 / (1 << 20), 1.0 * Mmr_StepMemory(p.pManCuts) / (1 << 17), 1.0 * Mig_ManObjNum(p.pMig) * sizeof(Mig_Obj_t) / (1 << 20) + 
            1.0 * Mig_ManObjNum(p.pMig) * 48 / (1 << 20) +
            1.0 * Mmr_StepMemory(p.pManCuts) / (1 << 17));
    if p.timeDerive {
        fmt.Printf("\n");
        p.timeTotal = Abc_Clock() - p.timeTotal;
        p.timeOther = p.timeTotal - p.timeDerive;

        Abc_Print( 1, "Runtime breakdown:\n" );
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Complete cut computation   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeDerive)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeDerive)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("- Merging cuts             "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeMerge)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeMerge)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("- Evaluting cut parameters "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeEval)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeEval)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("- Checking cut containment "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCompare)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeCompare)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("- Adding cuts to storage   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeStore)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeStore)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Other                      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeOther)))/(float64((p.timeTotal))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("TOTAL                      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;
    
    } else {
        Abc_PrintTime( 1, "Time", Abc_Clock() - p.timeTotal ); }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


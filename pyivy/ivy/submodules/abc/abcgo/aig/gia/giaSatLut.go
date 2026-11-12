package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaSatLut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSatLut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/tim/tim.h"
//#include "sat/bsat/satStore.h"
//#include "misc/util/utilNam.h"
//#include "map/scl/sclCon.h"
//#include "misc/vec/vecHsh.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
/**Function*************************************************************

  Synopsis    [For each node in the window, create fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Timing computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Given mapping in p->vSolCur, check if mapping meets delay.]

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

func Sbl_ManPrintRuntime(p *Sbl_Man_t) {
    fmt.Printf("Runtime breakdown:\n");
    p.timeOther = p.timeTotal - p.timeWin - p.timeCut - p.timeSat - p.timeTime;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Win   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeWin)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeWin)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Cut   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCut)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeCut)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Sat   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Sat  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Unsat"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUns)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUns)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Undec"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnd)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUnd)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Timing"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTime)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTime)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Other "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeOther)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("ALL   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;
}
func Gia_ManLutSat(pGia *Gia_Man_t, LutSize int32, nNumber int32, nImproves int32, nBTLimit int32, DelayMax int32, nEdges int32, fDelay int32, fReverse int32, fVerbose int32, fVeryVerbose int32) {
    var iLut, nImproveCount int32;
    var p *Sbl_Man_t    = Sbl_ManAlloc( pGia, nNumber );
    p.LutSize      = LutSize;      // LUT size
    p.nBTLimit     = nBTLimit;     // conflicts
    p.DelayMax     = DelayMax;     // external delay
    p.nEdges       = nEdges;       // the number of edges
    p.fDelay       = fDelay;       // delay mode
    p.fReverse     = fReverse;     // reverse windowing
    p.fVerbose     = fVerbose | fVeryVerbose;
    p.fVeryVerbose = fVeryVerbose;
    if p.fVerbose {
    fmt.Printf("Parameters: WinSize = %d AIG nodes.  Conf = %d.  DelayMax = %d.\n", p.nVars, p.nBTLimit, p.DelayMax); }
    // determine delay limit
    if (fDelay != 0) && pGia.vEdge1 && p.DelayMax == 0 {
        p.DelayMax = Gia_ManEvalEdgeDelay( pGia ); }
    // iterate through the internal nodes
    Gia_ManComputeOneWinStart( pGia, nNumber, fReverse );
    // inlining of macro Gia_ManForEachLut2
    for  iLut = 1; iLut < Gia_ManObjNum(pGia); iLut++  { if !Gia_ObjIsLut2(pGia, iLut) {} else {
        if Sbl_ManTestSat( p, iLut ) != 2 {
            continue; }
        if nImproveCount++; nImproveCount == nImproves {
            break; }
    }
    }
    Gia_ManComputeOneWin( pGia, -1, nil, nil, nil, nil );
    if p.fVerbose {
    fmt.Printf("Tried = %d. Used = %d. HashWin = %d. SmallWin = %d. LargeWin = %d. IterOut = %d.  SAT runs = %d.\n", p.nTried, p.nImproved, p.nHashWins, p.nSmallWins, p.nLargeWins, p.nIterOuts, p.nRuns); }
    if p.fVerbose {
    Sbl_ManPrintRuntime( p ); }
    Sbl_ManStop( p );
    Vec_IntFreeP( &pGia.vPacking );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


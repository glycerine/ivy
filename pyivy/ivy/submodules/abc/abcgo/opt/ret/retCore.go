package ret

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [retCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Retiming package.]

  Synopsis    [The core retiming procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Oct 31, 2006.]

  Revision    [$Id: retCore.c,v 1.00 2006/10/31 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "retInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var timeRetime abctime;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Implementation of retiming.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetime(pNtk *Abc_Ntk_t, Mode int32, nDelayLim int32, fForwardOnly int32, fBackwardOnly int32, fOneStep int32, fUseOldNames int32, fVerbose int32) int32 {
    var nLatches int32  = Abc_NtkLatchNum(pNtk);
    var nLevels int32   = Abc_NtkLevel(pNtk);
    var RetValue int32;
    var clkTotal abctime  = Abc_Clock();
    var nNodesOld, nLatchesOld int32;
    assert(Mode > 0 && Mode < 7, "Mode > 0 && Mode < 7");
    assert(!(fForwardOnly != 0) || !(fBackwardOnly != 0), "!fForwardOnly || !fBackwardOnly");

    // cleanup the network
    nNodesOld   = Abc_NtkNodeNum(pNtk);
    nLatchesOld = Abc_NtkLatchNum(pNtk);
    Abc_NtkCleanupSeq(pNtk, 0, 0, 0);
    if nNodesOld > Abc_NtkNodeNum(pNtk) || nLatchesOld > Abc_NtkLatchNum(pNtk) {
        fmt.Printf("Cleanup before retiming removed %d dangling nodes and %d dangling latches.\n", nNodesOld - Abc_NtkNodeNum(pNtk), nLatchesOld - Abc_NtkLatchNum(pNtk)); }

    // perform retiming
    switch  Mode  {
    case 1: // forward 
        RetValue = Abc_NtkRetimeIncremental( pNtk, nDelayLim, 1, 0, 0, fUseOldNames, fVerbose );
        break;
    case 2: // backward 
        RetValue = Abc_NtkRetimeIncremental( pNtk, nDelayLim, 0, 0, 0, fUseOldNames, fVerbose );
        break;
    case 3: // min-area 
        RetValue = Abc_NtkRetimeMinArea( pNtk, fForwardOnly, fBackwardOnly, fUseOldNames, fVerbose );
        break;
    case 4: // min-delay
        if ( !fBackwardOnly )
            RetValue += Abc_NtkRetimeIncremental( pNtk, nDelayLim, 1, 1, fOneStep, fUseOldNames, fVerbose );
        if !(fForwardOnly != 0) {
            RetValue += Abc_NtkRetimeIncremental( pNtk, nDelayLim, 0, 1, fOneStep, fUseOldNames, fVerbose ); }
        break;
    case 5: // min-area + min-delay
        RetValue  = Abc_NtkRetimeMinArea( pNtk, fForwardOnly, fBackwardOnly, fUseOldNames, fVerbose );
        if !(fBackwardOnly != 0) {
            RetValue += Abc_NtkRetimeIncremental( pNtk, nDelayLim, 1, 1, 0, fUseOldNames, fVerbose ); }
        if !(fForwardOnly != 0) {
            RetValue += Abc_NtkRetimeIncremental( pNtk, nDelayLim, 0, 1, 0, fUseOldNames, fVerbose ); }
        break;
    case 6: // Pan's algorithm
        RetValue = Abc_NtkRetimeLValue( pNtk, 500, fVerbose );
        break;
    default:
        fmt.Printf("Unknown retiming option.\n");
        break;
    }
    if fVerbose != 0 {
        fmt.Printf("Reduction in area = %3d. Reduction in delay = %3d. ", nLatches - Abc_NtkLatchNum(pNtk), nLevels - Abc_NtkLevel(pNtk));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Total runtime"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
    }
    timeRetime = Abc_Clock() - clkTotal;
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Used for automated debugging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeDebug(pNtk *Abc_Ntk_t) int32 {
    extern int Abc_NtkSecFraig( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int nSeconds, int nFrames, int fVerbose );
    var pNtkRet *Abc_Ntk_t;
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    Abc_NtkToSop( pNtk, -1, ABC_INFINITY );
//    if ( !Abc_NtkCheck( pNtk ) )
//        fprintf( stdout, "Abc_NtkRetimeDebug(): Network check has failed.\n" );
//    Io_WriteBlifLogic( pNtk, "debug_temp.blif", 1 );
    pNtkRet = Abc_NtkDup( pNtk );
    Abc_NtkRetime( pNtkRet, 3, 0, 0, 1, 0, 1, 0 ); // debugging backward flow
    return !Abc_NtkSecFraig( pNtk, pNtkRet, 10000, 3, 0 );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


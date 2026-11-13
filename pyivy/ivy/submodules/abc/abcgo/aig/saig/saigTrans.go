package saig

import "time"

/**CFile****************************************************************

  FileName    [saigTrans.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Dynamic simplication of the transition relation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigTrans.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"
//#include "proof/fra/fra.h"

//ABC_NAMESPACE_IMPL_START


/* 
    A similar approach is presented in the his paper:
    A. Kuehlmann. Dynamic transition relation simplification for 
    bounded property checking. ICCAD'04, pp. 50-57.
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Maps a node/frame into a node of a different manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Implements dynamic simplification.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManTimeframeSimplify(pAig *Aig_Man_t, nFrames int32, nFramesMax int32, fInit int32, fVerbose int32) *Aig_Man_t {
//    extern Aig_Man_t * Fra_FraigEquivence( Aig_Man_t * pManAig, int nConfMax, int fProve );
    var pFrames, pFraig, pRes1, pRes2 *Aig_Man_t;
    var clk abctime;
    // create uninitialized timeframes with map1
    pFrames = Saig_ManFramesNonInitial( pAig, nFrames );
    // perform fraiging for the unrolled timeframes
clk = Abc_Clock();
    pFraig = Fra_FraigEquivence( pFrames, 1000, 0 );
    // report the results
    if fVerbose != 0 {
        Aig_ManPrintStats( pFrames );
        Aig_ManPrintStats( pFraig );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Fraiging"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    Aig_ManStop( pFraig );
    assert(pFrames.pReprs != nil, "pFrames.pReprs != nil");
    // create AIG with map2
    Saig_ManCreateMapping( pAig, pFrames, nFrames );
    Aig_ManStop( pFrames );
    Saig_ManStopMap1( pAig );
    // create reduced initialized timeframes
clk = Abc_Clock();
    pRes2 = Saig_ManFramesInitialMapped( pAig, nFrames, nFramesMax, fInit );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Mapped"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    // free mapping
    Saig_ManStopMap2( pAig );
clk = Abc_Clock();
    pRes1 = Saig_ManFramesInitialMapped( pAig, nFrames, nFramesMax, fInit );
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Normal"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    // report the results
    if fVerbose != 0 {
        Aig_ManPrintStats( pRes1 );
        Aig_ManPrintStats( pRes2 );
    }
    Aig_ManStop( pRes1 );
    assert(!(Saig_ManHasMap1(pAig) != 0), "!Saig_ManHasMap1(pAig)");
    assert(!(Saig_ManHasMap2(pAig) != 0), "!Saig_ManHasMap2(pAig)");
    return pRes2;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


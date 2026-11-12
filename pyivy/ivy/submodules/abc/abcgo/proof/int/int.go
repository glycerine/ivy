package int

/**CFile****************************************************************

  FileName    [int.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 24, 2008.]

  Revision    [$Id: int.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




/* 
    The interpolation algorithm implemented here was introduced in the paper:
    K. L. McMillan. Interpolation and SAT-based model checking. CAV_03, pp. 1-13.
*/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// simulation manager
   
type Inter_ManParams_t struct {
    nBTLimit  int32;      // limit on the number of conflicts
    nFramesMax  int32;    // the max number timeframes to unroll
    nSecLimit  int32;     // time limit in seconds
    nFramesK  int32;      // the number of timeframes to use in induction
    fRewrite  int32;      // use additional rewriting to simplify timeframes
    fTransLoop  int32;    // add transition into the init state under new PI var
    fUsePudlak  int32;    // use Pudluk interpolation procedure
    fUseOther  int32;     // use other undisclosed option
    fUseMiniSat  int32;   // use MiniSat-1.14p instead of internal proof engine
    fCheckKstep  int32;   // check using K-step induction
    fUseBias  int32;      // bias decisions to global variables
    fUseBackward  int32;  // perform backward interpolation
    fUseSeparate  int32;  // solve each output separately
    fUseTwoFrames  int32; // create the OR of two last timeframes
    fDropSatOuts  int32;  // replace by 1 the solved outputs
    fDropInvar  int32;    // dump inductive invariant into file
    fVerbose  int32;      // print verbose statistics
    iFrameMax  int32;     // the time frame reached
    pFileName  string;   // file name to dump interpolant
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== intCore.c ==========================================================*/
// extern void       Inter_ManSetDefaultParams( Inter_ManParams_t * p );

// extern int        Inter_ManPerformInterpolation( Aig_Man_t * pAig, Inter_ManParams_t * pPars, int * piFrame );





//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


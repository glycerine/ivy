package int2

/**CFile****************************************************************

  FileName    [int2.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 1, 2013.]

  Revision    [$Id: int2.h,v 1.00 2013/12/01 00:00:00 alanmi Exp $]

***********************************************************************/
 




/* 
    The interpolation algorithm implemented here was introduced in the papers:
    K. L. McMillan. Interpolation and SAT-based model checking. CAV_03, pp. 1-13.
    C.-Y. Wu et al. A CEX-Guided Interpolant Generation Algorithm for 
    SAT-based Model Checking. DAC'13.
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
   
type Int2_ManPars_t struct {
    nBTLimit     int32;       // limit on the number of conflicts
    nFramesS     int32;       // the starting number timeframes
    nFramesMax     int32;     // the max number timeframes to unroll
    nSecLimit     int32;      // time limit in seconds
    nFramesK     int32;       // the number of timeframes to use in induction
    fRewrite     int32;       // use additional rewriting to simplify timeframes
    fTransLoop     int32;     // add transition into the init state under new PI var
    fDropInvar     int32;     // dump inductive invariant into file
    fVerbose     int32;       // print verbose statistics
    iFrameMax     int32;      // the time frame reached
    pFileName   string;      // file name to dump interpolant
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== intCore.c ==========================================================*/
// extern void       Int2_ManSetDefaultParams( Int2_ManPars_t * p );

// extern int        Int2_ManPerformInterpolation( Gia_Man_t * p, Int2_ManPars_t * pPars );





//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


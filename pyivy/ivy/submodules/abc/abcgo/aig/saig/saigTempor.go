package saig

import "fmt"

/**CFile****************************************************************

  FileName    [saigTempor.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Temporal decomposition.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigTempor.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates initialized timeframes for temporal decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Find index of first non-zero entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntLastNonZeroBeforeLimit(vTemp *Vec_Int_t, Limit int32) int32 {
    var Entry, i int32;
    if vTemp == nil {
        return -1; }
    // inlining of macro Vec_IntForEachEntryReverse
    for  i = Vec_IntSize(vTemp) - 1; (i >= 0) && (((Entry) = Vec_IntEntry(vTemp, i)), 1); i--  {
        if i >= Limit {
            continue; }
        if Entry != 0 {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManTempor(pAig *Aig_Man_t, nFrames int32, TimeOut int32, nConfLimit int32, fUseBmc int32, fUseTransSigs int32, fVerbose int32, fVeryVerbose int32) *Aig_Man_t { 
    extern int Saig_ManPhasePrefixLength( Aig_Man_t * p, int fVerbose, int fVeryVerbose, Vec_Int_t ** pvTrans );

    var vTransSigs *Vec_Int_t  = nil;
    var RetValue int32;
    var nFramesFinished int32 = -1;
    assert(nFrames >= 0, "nFrames >= 0"); 
    if nFrames == 0 {
        nFrames = Saig_ManPhasePrefixLength( pAig, fVerbose, fVeryVerbose, &vTransSigs );
        if nFrames == 0 {
            Vec_IntFreeP( &vTransSigs );
            fmt.Printf("The leading sequence has length 0. Temporal decomposition is not performed.\n");
            return nil;
        }
        if nFrames == 1 {
            Vec_IntFreeP( &vTransSigs );
            fmt.Printf("The leading sequence has length 1. Temporal decomposition is not performed.\n");
            return nil;
        }
        if fUseTransSigs != 0 {
            var Entry int32;
            var i int32;
            var iLast int32 = -1;
            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(vTransSigs)) && (((Entry) = Vec_IntEntry(vTransSigs, i)), 1); i++  {
                if Entry != 0 {
                    iLast = i;
                } else {
                    iLast = iLast;
                }
            }
            if iLast > 0 && iLast < nFrames {
                Abc_Print( 1, "Reducing frame count from %d to %d to fit the last transient.\n", nFrames, iLast );
                nFrames = iLast;
            }
        }
        Abc_Print( 1, "Using computed frame number (%d).\n", nFrames );
    
    } else {
        Abc_Print( 1, "Using user-given frame number (%d).\n", nFrames ); }
    // run BMC2
    if fUseBmc != 0 {
        RetValue = Saig_BmcPerform( pAig, 0, nFrames, 2000, TimeOut, nConfLimit, 0, fVerbose, 0, &nFramesFinished, 0, 0 );
        if RetValue == 0 {
            Vec_IntFreeP( &vTransSigs );
            fmt.Printf("A cex found in the first %d frames.\n", nFrames);
            return nil;
        }
        if nFramesFinished + 1 < nFrames {
            var iLastBefore int32  = Vec_IntLastNonZeroBeforeLimit( vTransSigs, nFramesFinished );
            if iLastBefore < 1 || !(fUseTransSigs != 0) {
                Vec_IntFreeP( &vTransSigs );
                fmt.Printf("BMC for %d frames could not be completed. A cex may exist!\n", nFrames);
                return nil;
            }
            assert(iLastBefore < nFramesFinished, "iLastBefore < nFramesFinished");
            fmt.Printf("BMC succeeded to frame %d. Adjusting frame count to be (%d) based on the last transient signal.\n", nFramesFinished, iLastBefore);
            nFrames = iLastBefore;
        }
    }
    Vec_IntFreeP( &vTransSigs );
    return Saig_ManTemporDecompose( pAig, nFrames );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END


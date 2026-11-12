package lpk

/**CFile****************************************************************

  FileName    [lpk.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast Boolean matching for LUT structures.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: lpk.h,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/
 




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

   
type Lpk_Par_t struct {
    // user-controlled parameters
    nLutsMax               int32;      // (N) the maximum number of LUTs in the structure
    nLutsOver               int32;     // (Q) the maximum number of LUTs not in the MFFC
    nVarsShared               int32;   // (S) the maximum number of shared variables (crossbars)
    nGrowthLevel               int32;  // (L) the maximum increase in the node level after resynthesis
    fSatur               int32;        // iterate till saturation
    fZeroCost               int32;     // accept zero-cost replacements
    fFirst               int32;        // use root node and first cut only
    fOldAlgo               int32;      // use old algorithm
    fVerbose               int32;      // the verbosiness flag
    fVeryVerbose               int32;  // additional verbose info printout
    // internal parameters
    nLutSize               int32;      // (K) the LUT size (determined by the input network)
    nVarsMax               int32;      // (V) the largest number of variables: V = N * (K-1) + 1
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                           ITERATORS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== lpkCore.c ========================================================*/
// extern int     Lpk_Resynthesize( Abc_Ntk_t * pNtk, Lpk_Par_t * pPars );





//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


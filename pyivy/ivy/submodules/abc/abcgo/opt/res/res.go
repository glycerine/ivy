package res

/**CFile****************************************************************

  FileName    [res.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resynthesis package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 15, 2007.]

  Revision    [$Id: res.h,v 1.00 2007/01/15 00:00:00 alanmi Exp $]

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

   
type Res_Par_t struct {
    // general parameters
    nWindow           int32;       // window size
    nGrowthLevel           int32;  // the maximum allowed growth in level after one iteration of resynthesis
    nSimWords           int32;     // the number of simulation words 
    nCands           int32;        // the number of candidates to try
    fArea           int32;         // performs optimization for area
    fDelay           int32;        // performs optimization for delay
    fVerbose           int32;      // enable basic stats
    fVeryVerbose           int32;  // enable detailed stats
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== resCore.c ==========================================================*/
// extern int        Abc_NtkResynthesize( Abc_Ntk_t * pNtk, Res_Par_t * pPars );





//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


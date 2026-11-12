package mfs

/**CFile****************************************************************

  FileName    [mfs.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The good old minimization with complete don't-cares.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mfs.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

   
type Mfs_Par_t struct {
    // general parameters
    nWinTfoLevs           int32;   // the maximum fanout levels
    nFanoutsMax           int32;   // the maximum number of fanouts
    nDepthMax           int32;     // the maximum number of logic levels
    nWinMax           int32;       // the maximum size of the window
    nGrowthLevel           int32;  // the maximum allowed growth in level
    nBTLimit           int32;      // the maximum number of conflicts in one SAT run
    fRrOnly           int32;       // perform redundance removal
    fResub           int32;        // performs resubstitution
    fArea           int32;         // performs optimization for area
    fMoreEffort           int32;   // performs high-affort minimization
    fSwapEdge           int32;     // performs edge swapping
    fOneHotness           int32;   // adds one-hotness conditions
    fDelay           int32;        // performs optimization for delay
    fPower           int32;        // performs power-aware optimization
    fGiaSat           int32;       // use new SAT solver
    fVerbose           int32;      // enable basic stats
    fVeryVerbose           int32;  // enable detailed stats
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== mfsCore.c ==========================================================*/
// extern void        Abc_NtkMfsParsDefault( Mfs_Par_t * pPars );

// extern int         Abc_NtkMfs( Abc_Ntk_t * pNtk, Mfs_Par_t * pPars );


 


//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


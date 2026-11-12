package ssc

/**CFile****************************************************************

  FileName    [ssc.h] 

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sweeping under constraints.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: ssc.h,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

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

// choicing parameters
   
type Ssc_Pars_t struct {
    nWords              int32;        // the number of simulation words
    nBTLimit              int32;      // conflict limit at a node
    nSatVarMax              int32;    // the max number of SAT variables
    nCallsRecycle              int32; // calls to perform before recycling SAT solver
    fAppend              int32;       // append constraints to the resulting AIG
    fVerbose              int32;      // verbose stats
    fVerify              int32;       // enable internal verification
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== sscCore.c ==========================================================*/
// extern void          Ssc_ManSetDefaultParams( Ssc_Pars_t * p );

// extern Gia_Man_t *   Ssc_PerformSweeping( Gia_Man_t * pAig, Gia_Man_t * pCare, Ssc_Pars_t * pPars );

// extern Gia_Man_t *   Ssc_PerformSweepingConstr( Gia_Man_t * p, Ssc_Pars_t * pPars );



//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


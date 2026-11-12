package fsim

/**CFile****************************************************************

  FileName    [fsim.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast sequential AIG simulator.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: fsim.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

type Fsim_Man_t = Fsim_Man_t_;

// simulation parameters
   
type Fsim_ParSim_t struct {
    // user-controlled parameters
    nWords             int32;       // the number of machine words
    nIters             int32;       // the number of timeframes
    TimeLimit             int32;    // time limit in seconds
    fCheckMiter             int32;  // check if miter outputs are non-zero
    fVerbose             int32;     // enables verbose output
    // internal parameters
    fCompressAig             int32; // compresses internal data
}

// switching estimation parameters
   
type Fsim_ParSwitch_t struct {
    // user-controlled parameters
    nWords             int32;       // the number of machine words
    nIters             int32;       // the number of timeframes
    nPref             int32;        // the number of first timeframes to skip
    nRandPiNum             int32;   // PI trans prob (0=1/2; 1=1/4; 2=1/8, etc)
    fProbOne             int32;     // collect probability of one
    fProbTrans             int32;   // collect probatility of switching
    fVerbose             int32;     // enables verbose output
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== fsimCore.c ==========================================================*/
// extern void           Fsim_ManSetDefaultParamsSim( Fsim_ParSim_t * p );

// extern void           Fsim_ManSetDefaultParamsSwitch( Fsim_ParSwitch_t * p );

/*=== fsimSim.c ==========================================================*/
// extern int            Fsim_ManSimulate( Aig_Man_t * pAig, Fsim_ParSim_t * pPars );

/*=== fsimSwitch.c ==========================================================*/
// extern Vec_Int_t *    Fsim_ManSwitchSimulate( Aig_Man_t * pAig, Fsim_ParSwitch_t * pPars );

/*=== fsimTsim.c ==========================================================*/
// extern Vec_Ptr_t *    Fsim_ManTerSimulate( Aig_Man_t * pAig, int fVerbose );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


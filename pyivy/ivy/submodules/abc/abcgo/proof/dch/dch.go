package dch

/**CFile****************************************************************

  FileName    [dch.h] 

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Choice computation for tech-mapping.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: dch.h,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

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
   
type Dch_Pars_t struct {
    nWords              int32;        // the number of simulation words
    nBTLimit              int32;      // conflict limit at a node
    nSatVarMax              int32;    // the max number of SAT variables
    fSynthesis              int32;    // set to 1 to perform synthesis
    fPolarFlip              int32;    // uses polarity adjustment
    fSimulateTfo              int32;  // uses simulation of TFO classes
    fPower              int32;        // uses power-aware rewriting
    fUseGia              int32;       // uses GIA package 
    fUseCSat              int32;      // uses circuit-based solver
    fUseNew              int32;       // uses new implementation
    fLightSynth              int32;   // uses lighter version of synthesis
    fSkipRedSupp              int32;  // skip choices with redundant support vars
    fVerbose              int32;      // verbose stats
    timeSynth          abctime;     // synthesis runtime
    nNodesAhead              int32;   // the lookahead in terms of nodes
    nCallsRecycle              int32; // calls to perform before recycling SAT solver
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== dchAig.c ==========================================================*/
// extern Aig_Man_t *   Dch_DeriveTotalAig( Vec_Ptr_t * vAigs );

/*=== dchCore.c ==========================================================*/
// extern void          Dch_ManSetDefaultParams( Dch_Pars_t * p );

// extern int           Dch_ManReadVerbose( Dch_Pars_t * p );

// extern Aig_Man_t *   Dch_ComputeChoices( Aig_Man_t * pAig, Dch_Pars_t * pPars );

// extern void          Dch_ComputeEquivalences( Aig_Man_t * pAig, Dch_Pars_t * pPars );

/*=== dchScript.c ==========================================================*/
// extern Aig_Man_t *   Dar_ManChoiceNew( Aig_Man_t * pAig, Dch_Pars_t * pPars );



//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


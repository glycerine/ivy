package sfm

/**CFile****************************************************************

  FileName    [sfm.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sfm.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "misc/vec/vecWec.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Sfm_Ntk_t = Sfm_Ntk_t_;
   
type Sfm_Par_t struct {
    nTfoLevMax             int32;    // the maximum fanout levels
    nTfiLevMax             int32;    // the maximum fanin levels
    nFanoutMax             int32;    // the maximum number of fanouts
    nDepthMax             int32;     // the maximum depth to try
    nVarMax             int32;       // the maximum variable count
    nMffcMin             int32;      // the minimum MFFC size
    nMffcMax             int32;      // the maximum MFFC size
    nDecMax             int32;       // the maximum number of decompositions
    nWinSizeMax             int32;   // the maximum window size
    nGrowthLevel             int32;  // the maximum allowed growth in level
    nBTLimit             int32;      // the maximum number of conflicts in one SAT run
    nNodesMax             int32;     // the maximum number of nodes to try
    iNodeOne             int32;      // one particular node to try
    nFirstFixed             int32;   // the number of first nodes to be treated as fixed
    nTimeWin             int32;      // the size of timing window in percents
    DeltaCrit             int32;     // delay delta in picoseconds
    DelAreaRatio             int32;  // delay/area tradeoff (how many ps we trade for a unit of area)
    fRrOnly             int32;       // perform redundance removal
    fArea             int32;         // performs optimization for area
    fAreaRev             int32;      // performs optimization for area in reverse order
    fMoreEffort             int32;   // performs high-affort minimization
    fUseAndOr             int32;     // enable internal detection of AND/OR gates
    fZeroCost             int32;     // enable zero-cost replacement
    fUseSim             int32;       // enable simulation
    fUseDcs             int32;       // enable deriving don't-cares
    fPrintDecs             int32;    // enable printing decompositions
    fAllBoxes             int32;     // enable preserving all boxes
    fLibVerbose             int32;   // enable library stats
    fDelayVerbose             int32; // enable delay stats
    fVerbose             int32;      // enable basic stats
    fVeryVerbose             int32;  // enable detailed stats
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== sfmCnf.c ==========================================================*/
/*=== sfmCore.c ==========================================================*/
// extern void         Sfm_ParSetDefault( Sfm_Par_t * pPars );

// extern int          Sfm_NtkPerform( Sfm_Ntk_t * p, Sfm_Par_t * pPars );

/*=== sfmNtk.c ==========================================================*/
// extern Sfm_Ntk_t *  Sfm_NtkConstruct( Vec_Wec_t * vFanins, int nPis, int nPos, Vec_Str_t * vFixed, Vec_Str_t * vEmpty, Vec_Wrd_t * vTruths, Vec_Int_t * vStarts, Vec_Wrd_t * vTruths2 );

// extern void         Sfm_NtkFree( Sfm_Ntk_t * p );

// extern Vec_Int_t *  Sfm_NodeReadFanins( Sfm_Ntk_t * p, int i );

// extern word *       Sfm_NodeReadTruth( Sfm_Ntk_t * p, int i );

// extern int          Sfm_NodeReadFixed( Sfm_Ntk_t * p, int i );

// extern int          Sfm_NodeReadUsed( Sfm_Ntk_t * p, int i );

/*=== sfmWin.c ==========================================================*/
// extern Vec_Int_t *  Sfm_NtkDfs( Sfm_Ntk_t * p, Vec_Wec_t * vGroups, Vec_Int_t * vGroupMap, Vec_Int_t * vBoxesLeft, int fAllBoxes );



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


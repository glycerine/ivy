package dar

/**CFile****************************************************************

  FileName    [dar.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: dar.h,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

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

type Dar_RwrPar_t =            Dar_RwrPar_t_;
type Dar_RefPar_t =            Dar_RefPar_t_;

type Dar_RwrPar_t_ struct {
    nCutsMax              int32;       // the maximum number of cuts to try
    nSubgMax              int32;       // the maximum number of subgraphs to try
    fFanout              int32;        // support fanout representation
    fUpdateLevel              int32;   // update level 
    fUseZeros              int32;      // performs zero-cost replacement
    fPower              int32;         // enables power-aware rewriting
    fRecycle              int32;       // enables cut recycling
    fVerbose              int32;       // enables verbose output
    fVeryVerbose              int32;   // enables very verbose output
}

type Dar_RefPar_t_ struct {
    nMffcMin              int32;       // the min MFFC size for which refactoring is used
    nLeafMax              int32;       // the max number of leaves of a cut
    nCutsMax              int32;       // the max number of cuts to consider  
    fExtend              int32;        // extends the cut below MFFC
    fUpdateLevel              int32;   // updates the level after each move
    fUseZeros              int32;      // perform zero-cost replacements
    fVerbose              int32;       // verbosity level
    fVeryVerbose              int32;   // enables very verbose output
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                             ITERATORS                            ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== darLib.c ========================================================*/
// extern void            Dar_LibStart();

// extern void            Dar_LibStop();

// extern void            Dar_LibPrepare( int nSubgraphs );

// extern int             Dar_LibReturnClass( unsigned uTruth );

/*=== darBalance.c ========================================================*/
// extern Aig_Man_t *     Dar_ManBalance( Aig_Man_t * p, int fUpdateLevel );

// extern Aig_Man_t *     Dar_ManBalanceXor( Aig_Man_t * pAig, int fExor, int fUpdateLevel, int fVerbose );

// extern void            Dar_BalancePrintStats( Aig_Man_t * p );

/*=== darCore.c ========================================================*/
// extern void            Dar_ManDefaultRwrParams( Dar_RwrPar_t * pPars );

// extern int             Dar_ManRewrite( Aig_Man_t * pAig, Dar_RwrPar_t * pPars );

// extern Aig_MmFixed_t * Dar_ManComputeCuts( Aig_Man_t * pAig, int nCutsMax, int fSkipTtMin, int fVerbose );

/*=== darRefact.c ========================================================*/
// extern void            Dar_ManDefaultRefParams( Dar_RefPar_t * pPars );

// extern int             Dar_ManRefactor( Aig_Man_t * pAig, Dar_RefPar_t * pPars );

/*=== darScript.c ========================================================*/
// extern Aig_Man_t *     Dar_ManRewriteDefault( Aig_Man_t * pAig );

// extern Aig_Man_t *     Dar_ManRwsat( Aig_Man_t * pAig, int fBalance, int fVerbose );

// extern Aig_Man_t *     Dar_ManCompress( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fVerbose );

// extern Aig_Man_t *     Dar_ManCompress2( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fFanout, int fPower, int fVerbose );

// extern Aig_Man_t *     Dar_ManChoice( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fConstruct, int nConfMax, int nLevelMax, int fVerbose );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


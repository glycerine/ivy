package int

/**CFile****************************************************************

  FileName    [intInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 24, 2008.]

  Revision    [$Id: intInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/saig/saig.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver.h"
//#include "sat/bsat/satStore.h"
//#include "int.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// interpolation manager
   
type Inter_Man_t struct {
    // AIG manager
    pAig       *Aig_Man_t;         // the original AIG manager
    pAigTrans       *Aig_Man_t;    // the transformed original AIG manager
    pCnfAig       *Cnf_Dat_t;      // CNF for the original manager
    // interpolant
    pInter       *Aig_Man_t;       // the current interpolant
    pCnfInter       *Cnf_Dat_t;    // CNF for the current interplant
    // timeframes
    pFrames       *Aig_Man_t;      // the timeframes      
    pCnfFrames       *Cnf_Dat_t;   // CNF for the timeframes 
    // other data
    vVarsAB       *Vec_Int_t;      // the variables participating in 
    // temporary place for the new interpolant
    pInterNew       *Aig_Man_t;
    vInters       *Vec_Ptr_t;
    // parameters
    nFrames              int32;      // the number of timeframes
    nConfCur              int32;     // the current number of conflicts
    nConfLimit              int32;   // the limit on the number of conflicts
    fVerbose              int32;     // the verbosiness flag
    pFileName            string;
    // runtime
    timeRwr          abctime;
    timeCnf          abctime;
    timeSat          abctime;
    timeInt          abctime;
    timeEqu          abctime;
    timeOther          abctime;
    timeTotal          abctime;
}

// containment checking manager
type Inter_Check_t = Inter_Check_t_;

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== intCheck.c ============================================================*/
// extern Inter_Check_t * Inter_CheckStart( Aig_Man_t * pTrans, int nFramesK );

// extern void            Inter_CheckStop( Inter_Check_t * p );

// extern int             Inter_CheckPerform( Inter_Check_t * p, Cnf_Dat_t * pCnf, abctime nTimeNewOut );


/*=== intContain.c ============================================================*/
// extern int             Inter_ManCheckContainment( Aig_Man_t * pNew, Aig_Man_t * pOld );

// extern int             Inter_ManCheckEquivalence( Aig_Man_t * pNew, Aig_Man_t * pOld );

// extern int             Inter_ManCheckInductiveContainment( Aig_Man_t * pTrans, Aig_Man_t * pInter, int nSteps, int fBackward );


/*=== intCtrex.c ============================================================*/
// extern void *          Inter_ManGetCounterExample( Aig_Man_t * pAig, int nFrames, int fVerbose );


/*=== intDup.c ============================================================*/
// extern Aig_Man_t *     Inter_ManStartInitState( int nRegs );

// extern Aig_Man_t *     Inter_ManStartDuplicated( Aig_Man_t * p );

// extern Aig_Man_t *     Inter_ManStartOneOutput( Aig_Man_t * p, int fAddFirstPo );


/*=== intFrames.c ============================================================*/
// extern Aig_Man_t *     Inter_ManFramesInter( Aig_Man_t * pAig, int nFrames, int fAddRegOuts, int fUseTwoFrames );


/*=== intMan.c ============================================================*/
// extern Inter_Man_t *   Inter_ManCreate( Aig_Man_t * pAig, Inter_ManParams_t * pPars );

// extern void            Inter_ManClean( Inter_Man_t * p );

// extern void            Inter_ManStop( Inter_Man_t * p, int fProved );


/*=== intM114.c ============================================================*/
// extern int             Inter_ManPerformOneStep( Inter_Man_t * p, int fUseBias, int fUseBackward, abctime nTimeNewOut );


/*=== intM114p.c ============================================================*/
#ifdef ABC_USE_LIBRARIES
extern int             Inter_ManPerformOneStepM114p( Inter_Man_t * p, int fUsePudlak, int fUseOther );
#endif

/*=== intUtil.c ============================================================*/
// extern int             Inter_ManCheckInitialState( Aig_Man_t * p );

// extern int             Inter_ManCheckAllStates( Aig_Man_t * p );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


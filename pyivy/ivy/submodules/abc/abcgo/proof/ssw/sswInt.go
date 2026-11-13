package ssw

/**CFile****************************************************************

  FileName    [sswInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswInt.h,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/saig/saig.h"
//#include "sat/bsat/satSolver.h"
//#include "ssw.h"
//#include "aig/ioa/ioa.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== sswAig.c ===================================================*/
// extern Ssw_Frm_t *   Ssw_FrmStart( Aig_Man_t * pAig );

// extern void          Ssw_FrmStop( Ssw_Frm_t * p );

// extern Aig_Man_t *   Ssw_FramesWithClasses( Ssw_Man_t * p );

// extern Aig_Man_t *   Ssw_SpeculativeReduction( Ssw_Man_t * p );

/*=== sswBmc.c ===================================================*/
/*=== sswClass.c =================================================*/
// extern Ssw_Cla_t *   Ssw_ClassesStart( Aig_Man_t * pAig );

// extern void          Ssw_ClassesSetData( Ssw_Cla_t * p, void * pManData,
//                          unsigned (*pFuncNodeHash)(void *,Aig_Obj_t *),
//                          int (*pFuncNodeIsConst)(void *,Aig_Obj_t *),
//                          int (*pFuncNodesAreEqual)(void *,Aig_Obj_t *, Aig_Obj_t *) );

// extern void          Ssw_ClassesStop( Ssw_Cla_t * p );

// extern Aig_Man_t *   Ssw_ClassesReadAig( Ssw_Cla_t * p );

// extern Vec_Ptr_t *   Ssw_ClassesGetRefined( Ssw_Cla_t * p );

// extern void          Ssw_ClassesClearRefined( Ssw_Cla_t * p );

// extern int           Ssw_ClassesCand1Num( Ssw_Cla_t * p );

// extern int           Ssw_ClassesClassNum( Ssw_Cla_t * p );

// extern int           Ssw_ClassesLitNum( Ssw_Cla_t * p );

// extern Aig_Obj_t **  Ssw_ClassesReadClass( Ssw_Cla_t * p, Aig_Obj_t * pRepr, int * pnSize );

// extern void          Ssw_ClassesCollectClass( Ssw_Cla_t * p, Aig_Obj_t * pRepr, Vec_Ptr_t * vClass );

// extern void          Ssw_ClassesCheck( Ssw_Cla_t * p );

// extern void          Ssw_ClassesPrint( Ssw_Cla_t * p, int fVeryVerbose );

// extern void          Ssw_ClassesRemoveNode( Ssw_Cla_t * p, Aig_Obj_t * pObj );

// extern Ssw_Cla_t *   Ssw_ClassesPrepare( Aig_Man_t * pAig, int nFramesK, int fLatchCorr, int fConstCorr, int fOutputCorr, int nMaxLevs, int fVerbose );

// extern Ssw_Cla_t *   Ssw_ClassesPrepareSimple( Aig_Man_t * pAig, int fLatchCorr, int nMaxLevs );

// extern Ssw_Cla_t *   Ssw_ClassesPrepareFromReprs( Aig_Man_t * pAig );

// extern Ssw_Cla_t *   Ssw_ClassesPrepareTargets( Aig_Man_t * pAig );

// extern Ssw_Cla_t *   Ssw_ClassesPreparePairs( Aig_Man_t * pAig, Vec_Int_t ** pvClasses );

// extern Ssw_Cla_t *   Ssw_ClassesPreparePairsSimple( Aig_Man_t * pMiter, Vec_Int_t * vPairs );

// extern int           Ssw_ClassesRefine( Ssw_Cla_t * p, int fRecursive );

// extern int           Ssw_ClassesRefineGroup( Ssw_Cla_t * p, Vec_Ptr_t * vReprs, int fRecursive );

// extern int           Ssw_ClassesRefineOneClass( Ssw_Cla_t * p, Aig_Obj_t * pRepr, int fRecursive );

// extern int           Ssw_ClassesRefineConst1Group( Ssw_Cla_t * p, Vec_Ptr_t * vRoots, int fRecursive );

// extern int           Ssw_ClassesRefineConst1( Ssw_Cla_t * p, int fRecursive );

// extern int           Ssw_ClassesPrepareRehash( Ssw_Cla_t * p, Vec_Ptr_t * vCands, int fConstCorr );

/*=== sswCnf.c ===================================================*/
// extern Ssw_Sat_t *   Ssw_SatStart( int fPolarFlip );

// extern void          Ssw_SatStop( Ssw_Sat_t * p );

// extern void          Ssw_CnfNodeAddToSolver( Ssw_Sat_t * p, Aig_Obj_t * pObj );

// extern int           Ssw_CnfGetNodeValue( Ssw_Sat_t * p, Aig_Obj_t * pObjFraig );

/*=== sswConstr.c ===================================================*/
// extern int           Ssw_ManSweepBmcConstr( Ssw_Man_t * p );

// extern int           Ssw_ManSweepConstr( Ssw_Man_t * p );

// extern void          Ssw_ManRefineByConstrSim( Ssw_Man_t * p );

/*=== sswCore.c ===================================================*/
// extern Aig_Man_t *   Ssw_SignalCorrespondenceRefine( Ssw_Man_t * p );

/*=== sswDyn.c ===================================================*/
// extern void          Ssw_ManLoadSolver( Ssw_Man_t * p, Aig_Obj_t * pRepr, Aig_Obj_t * pObj );

// extern int           Ssw_ManSweepDyn( Ssw_Man_t * p );

/*=== sswLcorr.c ==========================================================*/
// extern int           Ssw_ManSweepLatch( Ssw_Man_t * p );

/*=== sswMan.c ===================================================*/
// extern Ssw_Man_t *   Ssw_ManCreate( Aig_Man_t * pAig, Ssw_Pars_t * pPars );

// extern void          Ssw_ManCleanup( Ssw_Man_t * p );

// extern void          Ssw_ManStop( Ssw_Man_t * p );

/*=== sswSat.c ===================================================*/
// extern int           Ssw_NodesAreEquiv( Ssw_Man_t * p, Aig_Obj_t * pOld, Aig_Obj_t * pNew );

// extern int           Ssw_NodesAreConstrained( Ssw_Man_t * p, Aig_Obj_t * pOld, Aig_Obj_t * pNew );

// extern int           Ssw_NodeIsConstrained( Ssw_Man_t * p, Aig_Obj_t * pPoObj );

/*=== sswSemi.c ===================================================*/
// extern int           Ssw_FilterUsingSemi( Ssw_Man_t * pMan, int fCheckTargets, int nConfMax, int fVerbose );

/*=== sswSim.c ===================================================*/
// extern unsigned      Ssw_SmlObjHashWord( Ssw_Sml_t * p, Aig_Obj_t * pObj );

// extern int           Ssw_SmlObjIsConstWord( Ssw_Sml_t * p, Aig_Obj_t * pObj );

// extern int           Ssw_SmlObjsAreEqualWord( Ssw_Sml_t * p, Aig_Obj_t * pObj0, Aig_Obj_t * pObj1 );

// extern int           Ssw_SmlObjIsConstBit( void * p, Aig_Obj_t * pObj );

// extern int           Ssw_SmlObjsAreEqualBit( void * p, Aig_Obj_t * pObj0, Aig_Obj_t * pObj1 );

// extern void          Ssw_SmlAssignRandomFrame( Ssw_Sml_t * p, Aig_Obj_t * pObj, int iFrame );

// extern Ssw_Sml_t *   Ssw_SmlStart( Aig_Man_t * pAig, int nPref, int nFrames, int nWordsFrame );

// extern void          Ssw_SmlClean( Ssw_Sml_t * p );

// extern void          Ssw_SmlStop( Ssw_Sml_t * p );

// extern void          Ssw_SmlObjAssignConst( Ssw_Sml_t * p, Aig_Obj_t * pObj, int fConst1, int iFrame );

// extern void          Ssw_SmlObjSetWord( Ssw_Sml_t * p, Aig_Obj_t * pObj, unsigned Word, int iWord, int iFrame );

// extern void          Ssw_SmlAssignDist1Plus( Ssw_Sml_t * p, unsigned * pPat );

// extern void          Ssw_SmlSimulateOne( Ssw_Sml_t * p );

// extern void          Ssw_SmlSimulateOneFrame( Ssw_Sml_t * p );

// extern void          Ssw_SmlSimulateOneDyn_rec( Ssw_Sml_t * p, Aig_Obj_t * pObj, int f, int * pVisited, int nVisCounter );

// extern void          Ssw_SmlResimulateSeq( Ssw_Sml_t * p );

/*=== sswSimSat.c ===================================================*/
// extern void          Ssw_ManResimulateBit( Ssw_Man_t * p, Aig_Obj_t * pObj, Aig_Obj_t * pRepr );

// extern void          Ssw_ManResimulateWord( Ssw_Man_t * p, Aig_Obj_t * pCand, Aig_Obj_t * pRepr, int f );

/*=== sswSweep.c ===================================================*/
// extern int           Ssw_ManGetSatVarValue( Ssw_Man_t * p, Aig_Obj_t * pObj, int f );

// extern void          Ssw_SmlSavePatternAig( Ssw_Man_t * p, int f );

// extern int           Ssw_ManSweepNode( Ssw_Man_t * p, Aig_Obj_t * pObj, int f, int fBmc, Vec_Int_t * vPairs );

// extern int           Ssw_ManSweepBmc( Ssw_Man_t * p );

// extern int           Ssw_ManSweep( Ssw_Man_t * p );

/*=== sswUnique.c ===================================================*/
// extern void          Ssw_UniqueRegisterPairInfo( Ssw_Man_t * p );

// extern int           Ssw_ManUniqueOne( Ssw_Man_t * p, Aig_Obj_t * pRepr, Aig_Obj_t * pObj, int fVerbose );

// extern int           Ssw_ManUniqueAddConstraint( Ssw_Man_t * p, Vec_Ptr_t * vCommon, int f1, int f2 );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

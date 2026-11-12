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

type Ssw_Man_t = Ssw_Man_t_; // signal correspondence manager
type Ssw_Frm_t = Ssw_Frm_t_; // unrolled frames manager
type Ssw_Sat_t = Ssw_Sat_t_; // SAT solver manager
type Ssw_Cla_t = Ssw_Cla_t_; // equivalence classe manager

type Ssw_Man_t_ struct {
    // parameters
    pPars      *Ssw_Pars_t;          // parameters
    nFrames              int32;        // for quick lookup
    // AIGs used in the package
    pAig       *Aig_Man_t;           // user-given AIG
    pFrames       *Aig_Man_t;        // final AIG
    pNodeToFrames      **Aig_Obj_t;  // mapping of AIG nodes into FRAIG nodes
    // equivalence classes
    ppClasses       *Ssw_Cla_t;      // equivalence classes of nodes
    fRefined              int32;       // is set to 1 when refinement happens
    // SAT solving 
    pMSatBmc       *Ssw_Sat_t;       // SAT manager for base case
    pMSat       *Ssw_Sat_t;          // SAT manager for inductive case
    // SAT solving (latch corr only)
    vSimInfo       *Vec_Ptr_t;       // simulation information for the framed PIs
    nPatterns              int32;      // the number of patterns saved
    nSimRounds              int32;     // the number of simulation rounds performed
    nCallsCount              int32;    // the number of calls in this round
    nCallsDelta              int32;    // the number of calls to skip
    nCallsSat              int32;      // the number of SAT calls in this round
    nCallsUnsat              int32;    // the number of UNSAT calls in this round
    nRecycleCalls              int32;  // the number of calls since last recycling
    nRecycles              int32;      // the number of time SAT solver was recycled
    nRecyclesTotal              int32; // the number of time SAT solver was recycled
    nVarsMax              int32;       // the maximum variables in the solver
    nCallsMax              int32;      // the maximum number of SAT calls
    // uniqueness
    vCommon       *Vec_Ptr_t;        // the set of common variables in the logic cones
    iOutputLit              int32;     // the output literal of the uniqueness constraint
    vDiffPairs       *Vec_Int_t;     // is set to 1 if reg pair can be diff
    nUniques              int32;       // the number of uniqueness constraints used 
    nUniquesAdded              int32;  // useful uniqueness constraints
    nUniquesUseful              int32; // useful uniqueness constraints
    // dynamic constraint addition
    nSRMiterMaxId              int32;  // max ID after which the last frame begins
    vNewLos       *Vec_Ptr_t;        // new time frame LOs of to constrain
    vNewPos       *Vec_Int_t;        // new time frame POs of to add constraints
    pVisited             *int32;       // flags to label visited nodes in each frame
    nVisCounter              int32;    // the traversal ID
    // sequential simulation
    pSml       *Ssw_Sml_t;           // the simulator
    iNodeStart              int32;     // the first node considered
    iNodeLast              int32;      // the last node considered
    vResimConsts       *Vec_Ptr_t;   // resimulation constants
    vResimClasses       *Vec_Ptr_t;  // resimulation classes
    vInits       *Vec_Int_t;         // the init values of primary inputs under constraints
    // counter example storage
    nPatWords              int32;      // the number of words in the counter example
    pPatWords        *uint32;      // the counter example
    // constraints
    nConstrTotal              int32;   // the number of total constraints
    nConstrReduced              int32; // the number of reduced constraints
    nStrangers              int32;     // the number of strange situations
    // SAT calls statistics
    nSatCalls              int32;      // the number of SAT calls
    nSatProof              int32;      // the number of proofs
    nSatFailsReal              int32;  // the number of timeouts
    nSatCallsUnsat              int32; // the number of unsat SAT calls
    nSatCallsSat              int32;   // the number of sat SAT calls
    // node/register/lit statistics
    nLitsBeg              int32;
    nLitsEnd              int32;
    nNodesBeg              int32;
    nNodesEnd              int32;
    nRegsBeg              int32;
    nRegsEnd              int32;
    // equiv statistis
    nConesTotal              int32;
    nConesConstr              int32;
    nEquivsTotal              int32;
    nEquivsConstr              int32;
    nNodesBegC              int32;
    nNodesEndC              int32;
    nRegsBegC              int32;
    nRegsEndC              int32;
    // runtime stats
    timeBmc          abctime;        // bounded model checking
    timeReduce          abctime;     // speculative reduction
    timeMarkCones          abctime;  // marking the cones not to be refined
    timeSimSat          abctime;     // simulation of the counter-examples
    timeSat          abctime;        // solving SAT
    timeSatSat          abctime;     // sat
    timeSatUnsat          abctime;   // unsat
    timeSatUndec          abctime;   // undecided
    timeOther          abctime;      // other runtime
    timeTotal          abctime;      // total runtime
}

// internal SAT manager
type Ssw_Sat_t_ struct {
    pAig       *Aig_Man_t;           // the AIG manager
    fPolarFlip              int32;     // flips polarity 
    pSat      *sat_solver;           // recyclable SAT solver
    nSatVars              int32;       // the counter of SAT variables
    vSatVars       *Vec_Int_t;       // mapping of each node into its SAT var
    vFanins       *Vec_Ptr_t;        // fanins of the CNF node
    vUsedPis       *Vec_Ptr_t;       // the PIs with SAT variables 
    nSolverCalls              int32;   // the total number of SAT calls
}

// internal frames manager
type Ssw_Frm_t_ struct {
    pAig       *Aig_Man_t;           // user-given AIG
    nObjs              int32;          // offset in terms of AIG nodes
    nFrames              int32;        // the number of frames in current unrolling
    pFrames       *Aig_Man_t;        // unrolled AIG
    vAig2Frm       *Vec_Ptr_t;       // mapping of AIG nodes into frame nodes
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Ssw_ObjSatNum(p *Ssw_Sat_t, pObj *Aig_Obj_t) int32 { return Vec_IntGetEntry( p.vSatVars, pObj.Id );  }
func Ssw_ObjSetSatNum(p *Ssw_Sat_t, pObj *Aig_Obj_t, Num int32) { Vec_IntSetEntry(p.vSatVars, pObj.Id, Num);      }

func Ssw_ObjIsConst1Cand(pAig *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    return Aig_ObjRepr(pAig, pObj) == Aig_ManConst1(pAig);
}
func Ssw_ObjSetConst1Cand(pAig *Aig_Man_t, pObj *Aig_Obj_t) {
    assert(!(Ssw_ObjIsConst1Cand( pAig, pObj ) != 0), "!Ssw_ObjIsConst1Cand( pAig, pObj )");
    Aig_ObjSetRepr( pAig, pObj, Aig_ManConst1(pAig) );
}

func Ssw_ObjFrame(p *Ssw_Man_t, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { return p.pNodeToFrames[p.nFrames*pObj.Id + i];  }
func Ssw_ObjSetFrame(p *Ssw_Man_t, pObj *Aig_Obj_t, i int32, pNode *Aig_Obj_t) { p.pNodeToFrames[p.nFrames*pObj.Id + i] = pNode; }

func Ssw_ObjChild0Fra(p *Ssw_Man_t, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin0(pObj) {
    return Aig_NotCond(Ssw_ObjFrame(p, Aig_ObjFanin0(pObj), i), Aig_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Ssw_ObjChild1Fra(p *Ssw_Man_t, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin1(pObj) {
    return Aig_NotCond(Ssw_ObjFrame(p, Aig_ObjFanin1(pObj), i), Aig_ObjFaninC1(pObj));
} else {
    return nil;
}  }

func Ssw_ObjFrame_(p *Ssw_Frm_t, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { return (Aig_Obj_t *)Vec_PtrGetEntry( p.vAig2Frm, p.nObjs*i+pObj.Id );     }
func Ssw_ObjSetFrame_(p *Ssw_Frm_t, pObj *Aig_Obj_t, i int32, pNode *Aig_Obj_t) { Vec_PtrSetEntry( p.vAig2Frm, p.nObjs*i+pObj.Id, pNode );     }

func Ssw_ObjChild0Fra_(p *Ssw_Frm_t, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin0(pObj) {
    return Aig_NotCond(Ssw_ObjFrame_(p, Aig_ObjFanin0(pObj), i), Aig_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Ssw_ObjChild1Fra_(p *Ssw_Frm_t, pObj *Aig_Obj_t, i int32) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin1(pObj) {
    return Aig_NotCond(Ssw_ObjFrame_(p, Aig_ObjFanin1(pObj), i), Aig_ObjFaninC1(pObj));
} else {
    return nil;
}  }

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


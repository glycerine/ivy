package fra

/**CFile****************************************************************

  FileName    [fra.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [[New FRAIG package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fra.h,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "aig/aig/aig.h"
//#include "opt/dar/dar.h"
//#include "sat/bsat/satSolver.h"
//#include "aig/ioa/ioa.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START
 

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Fra_Par_t =   Fra_Par_t_;
type Fra_Ssw_t =   Fra_Ssw_t_;
type Fra_Sec_t =   Fra_Sec_t_;
type Fra_Man_t =   Fra_Man_t_;
type Fra_Cla_t =   Fra_Cla_t_;
type Fra_Sml_t =   Fra_Sml_t_;
type Fra_Bmc_t =   Fra_Bmc_t_;

// FRAIG parameters
type Fra_Par_t_ struct {
    nSimWords              int32;         // the number of words in the simulation info
    dSimSatur           float64;         // the ratio of refined classes when saturation is reached
    fPatScores              int32;        // enables simulation pattern scoring
    MaxScore              int32;          // max score after which resimulation is used
    dActConeRatio           float64;     // the ratio of cone to be bumped
    dActConeBumpMax           float64;   // the largest bump in activity
    fChoicing              int32;         // enables choicing
    fSpeculate              int32;        // use speculative reduction
    fProve              int32;            // prove the miter outputs
    fVerbose              int32;          // verbose output
    fDoSparse              int32;         // skip sparse functions
    fConeBias              int32;         // bias variables in the cone (good for unsat runs)
    nBTLimitNode              int32;      // conflict limit at a node
    nBTLimitMiter              int32;     // conflict limit at an output
    nLevelMax              int32;         // the max level to consider seriously
    nFramesP              int32;          // the number of timeframes to in the prefix
    nFramesK              int32;          // the number of timeframes to unroll
    nMaxImps              int32;          // the maximum number of implications to consider
    nMaxLevs              int32;          // the maximum number of levels to consider
    fRewrite              int32;          // use rewriting for constraint reduction
    fLatchCorr              int32;        // computes latch correspondence only
    fUseImps              int32;          // use implications
    fUse1Hot              int32;          // use one-hotness conditions
    fWriteImps              int32;        // record implications
    fDontShowBar              int32;      // does not show progressbar during fraiging
}

// seq SAT sweeping parameters
type Fra_Ssw_t_ struct {
    nPartSize              int32;         // size of the partition
    nOverSize              int32;         // size of the overlap between partitions
    nFramesP              int32;          // number of frames in the prefix
    nFramesK              int32;          // number of frames for induction (1=simple) 
    nMaxImps              int32;          // max implications to consider
    nMaxLevs              int32;          // max levels to consider
    nMinDomSize              int32;       // min clock domain considered for optimization
    fUseImps              int32;          // use implications  
    fRewrite              int32;          // enable rewriting of the specualatively reduced model
    fFraiging              int32;         // enable comb SAT sweeping as preprocessing 
    fLatchCorr              int32;        // perform register correspondence
    fWriteImps              int32;        // write implications into a file
    fUse1Hot              int32;          // use one-hotness constraints
    fVerbose              int32;          // enable verbose output 
    fSilent              int32;           // disable any output 
    nIters              int32;            // the number of iterations performed
    TimeLimit            float32;         // the runtime budget for this call
}

// SEC parametesr
type Fra_Sec_t_ struct {
    fTryComb              int32;          // try CEC call as a preprocessing step
    fTryBmc              int32;           // try BMC call as a preprocessing step 
    nFramesMax              int32;        // the max number of frames used for induction
    nBTLimit              int32;          // the conflict limit at a node
    nBTLimitGlobal              int32;    // the global conflict limit
    nBTLimitInter              int32;     // the conflict limit for interpolation
    nBddVarsMax              int32;       // the state space limit for BDD reachability
    nBddMax              int32;           // the max number of BDD nodes
    nBddIterMax              int32;       // the limit on the number of BDD iterations
    nPdrTimeout              int32;       // the timeout for PDR in the end
    fPhaseAbstract              int32;    // enables phase abstraction
    fRetimeFirst              int32;      // enables most-forward retiming at the beginning
    fRetimeRegs              int32;       // enables min-register retiming at the beginning
    fFraiging              int32;         // enables fraiging at the beginning
    fInduction              int32;        // enable the use of induction
    fInterpolation              int32;    // enables interpolation
    fInterSeparate              int32;    // enables interpolation for each outputs separately
    fReachability              int32;     // enables BDD based reachability
    fReorderImage              int32;     // enables BDD reordering during image computation
    fStopOnFirstFail              int32;  // enables stopping after first output of a miter has failed to prove
    fUseNewProver              int32;     // the new prover
    fUsePdr              int32;           // the PDR
    fSilent              int32;           // disables all output
    fVerbose              int32;          // enables verbose reporting of statistics
    fVeryVerbose              int32;      // enables very verbose reporting  
    TimeLimit              int32;         // enables the timeout
    fReadUnsolved              int32;     // inserts the unsolved model back
    nSMnumber              int32;         // the number of model written
    // internal parameters
    fRecursive              int32;        // set to 1 when SEC is called recursively
    fReportSolution              int32;   // enables report solution in a special form
}

// FRAIG equivalence classes
type Fra_Cla_t_ struct {
    pAig       *Aig_Man_t;              // the original AIG manager
    pMemRepr      **Aig_Obj_t;          // pointers to representatives of each node
    vClasses       *Vec_Ptr_t;          // equivalence classes
    vClasses1       *Vec_Ptr_t;         // equivalence class of Const1 node
    vClassesTemp       *Vec_Ptr_t;      // temporary storage for new classes
    pMemClasses      **Aig_Obj_t;       // memory allocated for equivalence classes
    pMemClassesFree      **Aig_Obj_t;   // memory allocated for equivalence classes to be used
    vClassOld       *Vec_Ptr_t;         // old equivalence class after splitting
    vClassNew       *Vec_Ptr_t;         // new equivalence class(es) after splitting
    nPairs              int32;            // the number of pairs of nodes
    fRefinement              int32;       // set to 1 when refinement has happened
    vImps       *Vec_Int_t;             // implications
    // procedures used for class refinement
    int (*pFuncNodeHash)     (Aig_Obj_t *, int);         // returns has key of the node
    int (*pFuncNodeIsConst)  (Aig_Obj_t *);              // returns 1 if the node is a constant
    int (*pFuncNodesAreEqual)(Aig_Obj_t *, Aig_Obj_t *); // returns 1 if nodes are equal up to a complement
}

// simulation manager
type Fra_Sml_t_ struct {
    pAig       *Aig_Man_t;              // the original AIG manager
    nPref              int32;             // the number of times frames in the prefix
    nFrames              int32;           // the number of times frames 
    nWordsFrame              int32;       // the number of words in each time frame
    nWordsTotal              int32;       // the total number of words at a node
    nWordsPref              int32;        // the number of word in the prefix
    fNonConstOut              int32;      // have seen a non-const-0 output during simulation
    nSimRounds              int32;        // statistics
    timeSim              int32;           // statistics
    pData [0]uint32;          // simulation data for the nodes
}

// FRAIG manager
type Fra_Man_t_ struct {
    // high-level data    
    pPars       *Fra_Par_t;             // parameters governing fraiging
    // AIG managers
    pManAig       *Aig_Man_t;           // the starting AIG manager
    pManFraig       *Aig_Man_t;         // the final AIG manager
    // mapping AIG into FRAIG
    nFramesAll              int32;        // the number of timeframes used
    pMemFraig      **Aig_Obj_t;         // memory allocated for points to the fraig nodes
    nSizeAlloc              int32;        // allocated size of the arrays for timeframe nodes
    // equivalence classes 
    pCla       *Fra_Cla_t;              // representation of (candidate) equivalent nodes
    // simulation info
    pSml       *Fra_Sml_t;              // simulation manager
    // bounded model checking manager
    pBmc       *Fra_Bmc_t;
    // counter example storage
    nPatWords              int32;         // the number of words in the counter example
    pPatWords        *uint32;         // the counter example
    vCex       *Vec_Int_t;
    // one-hotness conditions
    vOneHots       *Vec_Int_t;          
    // satisfiability solving
    pSat      *sat_solver;              // SAT solver
    nSatVars              int32;          // the number of variables currently used
    vPiVars       *Vec_Ptr_t;           // the PIs of the cone used 
    nBTLimitGlobal           ABC_INT64_T;    // resource limit
    nInsLimitGlobal           ABC_INT64_T;   // resource limit
    pMemFanins      **Vec_Ptr_t;        // the arrays of fanins for some FRAIG nodes
    pMemSatNums             *int32;       // the array of SAT numbers for some FRAIG nodes
    nMemAlloc              int32;         // allocated size of the arrays for FRAIG varnums and fanins
    vTimeouts       *Vec_Ptr_t;         // the nodes, for which equivalence checking timed out
    // statistics
    nSimRounds              int32;
    nNodesMiter              int32;
    nLitsBeg              int32;
    nLitsEnd              int32;
    nNodesBeg              int32;
    nNodesEnd              int32;
    nRegsBeg              int32;
    nRegsEnd              int32;
    nSatCalls              int32;
    nSatCallsSat              int32;
    nSatCallsUnsat              int32;
    nSatProof              int32;
    nSatFails              int32;
    nSatFailsReal              int32;
    nSpeculs              int32;   
    nChoices              int32;
    nChoicesFake              int32;
    nSatCallsRecent              int32;
    nSatCallsSkipped              int32;
    // runtime
    timeSim          abctime;
    timeTrav          abctime;
    timeRwr          abctime;
    timeSat          abctime;
    timeSatUnsat          abctime;
    timeSatSat          abctime;
    timeSatFail          abctime;
    timeRef          abctime;
    timeTotal          abctime;
    time1          abctime;
    time2          abctime;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Fra_ObjSim(p *Fra_Sml_t, Id int32) *uint32 { return p.pData + p.nWordsTotal * Id; }
func Fra_ObjRandomSim() uint32 { return Aig_ManRandom(0);               }

func Fra_ObjFraig(pObj *Aig_Obj_t, i int32) *Aig_Obj_t { return ((Fra_Man_t *)pObj.pData).pMemFraig[((Fra_Man_t *)pObj.pData).nFramesAll*pObj.Id + i];  }
func Fra_ObjSetFraig(pObj *Aig_Obj_t, i int32, pNode *Aig_Obj_t) { ((Fra_Man_t *)pObj.pData).pMemFraig[((Fra_Man_t *)pObj.pData).nFramesAll*pObj.Id + i] = pNode; }

func Fra_ObjFaninVec(pObj *Aig_Obj_t) *Vec_Ptr_t { return ((Fra_Man_t *)pObj.pData).pMemFanins[pObj.Id];      }
func Fra_ObjSetFaninVec(pObj *Aig_Obj_t, vFanins *Vec_Ptr_t) { ((Fra_Man_t *)pObj.pData).pMemFanins[pObj.Id] = vFanins;   }

func Fra_ObjSatNum(pObj *Aig_Obj_t) int32 { return ((Fra_Man_t *)pObj.pData).pMemSatNums[pObj.Id];     }
func Fra_ObjSetSatNum(pObj *Aig_Obj_t, Num int32) { ((Fra_Man_t *)pObj.pData).pMemSatNums[pObj.Id] = Num;      }

func Fra_ClassObjRepr(pObj *Aig_Obj_t) *Aig_Obj_t { return ((Fra_Man_t *)pObj.pData).pCla.pMemRepr[pObj.Id];  }
func Fra_ClassObjSetRepr(pObj *Aig_Obj_t, pNode *Aig_Obj_t) { ((Fra_Man_t *)pObj.pData).pCla.pMemRepr[pObj.Id] = pNode; }

func Fra_ObjChild0Fra(pObj *Aig_Obj_t, i int32) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin0(pObj) {
    return Aig_NotCond(Fra_ObjFraig(Aig_ObjFanin0(pObj),i), Aig_ObjFaninC0(pObj));
} else {
    return nil;
}  }
func Fra_ObjChild1Fra(pObj *Aig_Obj_t, i int32) *Aig_Obj_t { assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)"); if Aig_ObjFanin1(pObj) {
    return Aig_NotCond(Fra_ObjFraig(Aig_ObjFanin1(pObj),i), Aig_ObjFaninC1(pObj));
} else {
    return nil;
}  }

func Fra_ImpLeft(Imp int32) int32 { return Imp & 0xFFFF;         }
func Fra_ImpRight(Imp int32) int32 { return Imp >> 16;            }
func Fra_ImpCreate(Left int32, Right int32) int32 { return (Right << 16) | Left; }

////////////////////////////////////////////////////////////////////////
///                         ITERATORS                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== fraCec.c ========================================================*/
// extern int                 Fra_FraigSat( Aig_Man_t * pMan, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, int nLearnedStart, int nLearnedDelta, int nLearnedPerce, int fFlipBits, int fAndOuts, int fNewSolver, int fVerbose );

// extern int                 Fra_FraigCec( Aig_Man_t ** ppAig, int nConfLimit, int fVerbose );

// extern int                 Fra_FraigCecPartitioned( Aig_Man_t * pMan1, Aig_Man_t * pMan2, int nConfLimit, int nPartSize, int fSmart, int fVerbose );

/*=== fraClass.c ========================================================*/
// extern int                 Fra_BmcNodeIsConst( Aig_Obj_t * pObj );

// extern int                 Fra_BmcNodesAreEqual( Aig_Obj_t * pObj0, Aig_Obj_t * pObj1 );

// extern void                Fra_BmcStop( Fra_Bmc_t * p );

// extern void                Fra_BmcPerform( Fra_Man_t * p, int nPref, int nDepth );

// extern void                Fra_BmcPerformSimple( Aig_Man_t * pAig, int nFrames, int nBTLimit, int fRewrite, int fVerbose );

/*=== fraClass.c ========================================================*/
// extern Fra_Cla_t *         Fra_ClassesStart( Aig_Man_t * pAig );

// extern void                Fra_ClassesStop( Fra_Cla_t * p );

// extern void                Fra_ClassesCopyReprs( Fra_Cla_t * p, Vec_Ptr_t * vFailed );

// extern void                Fra_ClassesPrint( Fra_Cla_t * p, int fVeryVerbose );

// extern void                Fra_ClassesPrepare( Fra_Cla_t * p, int fLatchCorr, int nMaxLevs );

// extern int                 Fra_ClassesRefine( Fra_Cla_t * p );

// extern int                 Fra_ClassesRefine1( Fra_Cla_t * p, int fRefineNewClass, int * pSkipped );

// extern int                 Fra_ClassesCountLits( Fra_Cla_t * p );

// extern int                 Fra_ClassesCountPairs( Fra_Cla_t * p );

// extern void                Fra_ClassesTest( Fra_Cla_t * p, int Id1, int Id2 );

// extern void                Fra_ClassesLatchCorr( Fra_Man_t * p );

// extern void                Fra_ClassesPostprocess( Fra_Cla_t * p );

// extern void                Fra_ClassesSelectRepr( Fra_Cla_t * p );

// extern Aig_Man_t *         Fra_ClassesDeriveAig( Fra_Cla_t * p, int nFramesK );

/*=== fraCnf.c ========================================================*/
// extern void                Fra_CnfNodeAddToSolver( Fra_Man_t * p, Aig_Obj_t * pOld, Aig_Obj_t * pNew );

/*=== fraCore.c ========================================================*/
// extern void                Fra_FraigSweep( Fra_Man_t * pManAig );

// extern int                 Fra_FraigMiterStatus( Aig_Man_t * p );

// extern int                 Fra_FraigMiterAssertedOutput( Aig_Man_t * p );

// extern Aig_Man_t *         Fra_FraigPerform( Aig_Man_t * pManAig, Fra_Par_t * pPars );

// extern Aig_Man_t *         Fra_FraigChoice( Aig_Man_t * pManAig, int nConfMax, int nLevelMax );

// extern Aig_Man_t *         Fra_FraigEquivence( Aig_Man_t * pManAig, int nConfMax, int fProve );

/*=== fraHot.c ========================================================*/
// extern Vec_Int_t *         Fra_OneHotCompute( Fra_Man_t * p, Fra_Sml_t * pSim );

// extern void                Fra_OneHotAssume( Fra_Man_t * p, Vec_Int_t * vOneHots );

// extern void                Fra_OneHotCheck( Fra_Man_t * p, Vec_Int_t * vOneHots );

// extern int                 Fra_OneHotRefineUsingCex( Fra_Man_t * p, Vec_Int_t * vOneHots );

// extern int                 Fra_OneHotCount( Fra_Man_t * p, Vec_Int_t * vOneHots );

// extern void                Fra_OneHotEstimateCoverage( Fra_Man_t * p, Vec_Int_t * vOneHots );

// extern Aig_Man_t *         Fra_OneHotCreateExdc( Fra_Man_t * p, Vec_Int_t * vOneHots );

// extern void                Fra_OneHotAddKnownConstraint( Fra_Man_t * p, Vec_Ptr_t * vOnehots );

/*=== fraImp.c ========================================================*/
// extern Vec_Int_t *         Fra_ImpDerive( Fra_Man_t * p, int nImpMaxLimit, int nImpUseLimit, int fLatchCorr );

// extern void                Fra_ImpAddToSolver( Fra_Man_t * p, Vec_Int_t * vImps, int * pSatVarNums );

// extern int                 Fra_ImpCheckForNode( Fra_Man_t * p, Vec_Int_t * vImps, Aig_Obj_t * pNode, int Pos );

// extern int                 Fra_ImpRefineUsingCex( Fra_Man_t * p, Vec_Int_t * vImps );

// extern void                Fra_ImpCompactArray( Vec_Int_t * vImps );

// extern double              Fra_ImpComputeStateSpaceRatio( Fra_Man_t * p );

// extern int                 Fra_ImpVerifyUsingSimulation( Fra_Man_t * p );

// extern void                Fra_ImpRecordInManager( Fra_Man_t * p, Aig_Man_t * pNew );

/*=== fraInd.c ========================================================*/
// extern Aig_Man_t *         Fra_FraigInduction( Aig_Man_t * p, Fra_Ssw_t * pPars );

/*=== fraIndVer.c =====================================================*/
// extern int                 Fra_InvariantVerify( Aig_Man_t * p, int nFrames, Vec_Int_t * vClauses, Vec_Int_t * vLits );

/*=== fraLcr.c ========================================================*/
// extern Aig_Man_t *         Fra_FraigLatchCorrespondence( Aig_Man_t * pAig, int nFramesP, int nConfMax, int fProve, int fVerbose, int * pnIter, float TimeLimit );

/*=== fraMan.c ========================================================*/
// extern void                Fra_ParamsDefault( Fra_Par_t * pParams );

// extern void                Fra_ParamsDefaultSeq( Fra_Par_t * pParams );

// extern Fra_Man_t *         Fra_ManStart( Aig_Man_t * pManAig, Fra_Par_t * pParams );

// extern void                Fra_ManClean( Fra_Man_t * p, int nNodesMax );

// extern Aig_Man_t *         Fra_ManPrepareComb( Fra_Man_t * p );

// extern void                Fra_ManFinalizeComb( Fra_Man_t * p );

// extern void                Fra_ManStop( Fra_Man_t * p );

// extern void                Fra_ManPrint( Fra_Man_t * p );

/*=== fraSat.c ========================================================*/
// extern int                 Fra_NodesAreEquiv( Fra_Man_t * p, Aig_Obj_t * pOld, Aig_Obj_t * pNew );

// extern int                 Fra_NodesAreImp( Fra_Man_t * p, Aig_Obj_t * pOld, Aig_Obj_t * pNew, int fComplL, int fComplR );

// extern int                 Fra_NodesAreClause( Fra_Man_t * p, Aig_Obj_t * pOld, Aig_Obj_t * pNew, int fComplL, int fComplR );

// extern int                 Fra_NodeIsConst( Fra_Man_t * p, Aig_Obj_t * pNew );

/*=== fraSec.c ========================================================*/
// extern void                Fra_SecSetDefaultParams( Fra_Sec_t * p );

// extern int                 Fra_FraigSec( Aig_Man_t * p, Fra_Sec_t * pParSec, Aig_Man_t ** ppResult );

/*=== fraSim.c ========================================================*/
// extern int                 Fra_SmlNodeHash( Aig_Obj_t * pObj, int nTableSize );

// extern int                 Fra_SmlNodeIsConst( Aig_Obj_t * pObj );

// extern int                 Fra_SmlNodesAreEqual( Aig_Obj_t * pObj0, Aig_Obj_t * pObj1 );

// extern int                 Fra_SmlNodeNotEquWeight( Fra_Sml_t * p, int Left, int Right );

// extern int                 Fra_SmlNodeCountOnes( Fra_Sml_t * p, Aig_Obj_t * pObj );

// extern int                 Fra_SmlCheckOutput( Fra_Man_t * p );

// extern void                Fra_SmlSavePattern( Fra_Man_t * p );

// extern void                Fra_SmlSimulate( Fra_Man_t * p, int fInit );

// extern void                Fra_SmlResimulate( Fra_Man_t * p );

// extern Fra_Sml_t *         Fra_SmlStart( Aig_Man_t * pAig, int nPref, int nFrames, int nWordsFrame );

// extern void                Fra_SmlStop( Fra_Sml_t * p );

// extern Fra_Sml_t *         Fra_SmlSimulateComb( Aig_Man_t * pAig, int nWords, int fCheckMiter );

// extern Fra_Sml_t *         Fra_SmlSimulateCombGiven( Aig_Man_t * pAig, char * pFileName, int fCheckMiter, int fVerbose );

// extern Fra_Sml_t *         Fra_SmlSimulateSeq( Aig_Man_t * pAig, int nPref, int nFrames, int nWords, int fCheckMiter );

// extern Abc_Cex_t *         Fra_SmlGetCounterExample( Fra_Sml_t * p );

// extern Abc_Cex_t *         Fra_SmlCopyCounterExample( Aig_Man_t * pAig, Aig_Man_t * pFrames, int * pModel );


//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


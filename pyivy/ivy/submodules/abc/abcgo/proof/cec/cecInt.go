package cec

/**CFile****************************************************************

  FileName    [cecInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "sat/bsat/satSolver.h"
//#include "sat/glucose2/AbcGlucose2.h"
//#include "misc/bar/bar.h"
//#include "aig/gia/gia.h"
//#include "cec.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// simulation pattern manager
   
type Cec_ManPat_t struct {
    vPattern1       *Vec_Int_t;      // pattern in terms of primary inputs
    vPattern2       *Vec_Int_t;      // pattern in terms of primary inputs
    vStorage       *Vec_Str_t;       // storage for compressed patterns
    iStart              int32;         // position in the array where recent patterns begin
    nPats              int32;          // total number of recent patterns
    nPatsAll              int32;       // total number of all patterns
    nPatLits              int32;       // total number of literals in recent patterns
    nPatLitsAll              int32;    // total number of literals in all patterns
    nPatLitsMin              int32;    // total number of literals in minimized recent patterns
    nPatLitsMinAll              int32; // total number of literals in minimized all patterns
    nSeries              int32;        // simulation series
    fVerbose              int32;       // verbose stats
    // runtime statistics
    timeFind          abctime;       // detecting the pattern  
    timeShrink          abctime;     // minimizing the pattern
    timeVerify          abctime;     // verifying the result of minimisation
    timeSort          abctime;       // sorting literals 
    timePack          abctime;       // packing into sim info structures 
    timeTotal          abctime;      // total runtime  
    timeTotalSave          abctime;  // total runtime for saving  
}

// SAT solving manager
   
type Cec_ManSat_t struct {
    // parameters
    pPars    *Cec_ParSat_t;          
    // AIGs used in the package
    pAig       *Gia_Man_t;           // the AIG whose outputs are considered
    vStatus       *Vec_Int_t;        // status for each output
    // SAT solving
    pSat      *sat_solver;           // recyclable SAT solver (MiniSAT)
    pSat2 *bmcg2_sat_solver;          // recyclable SAT solver (Glucose)
    nSatVars              int32;       // the counter of SAT variables
    pSatVars             *int32;       // mapping of each node into its SAT var
    vUsedNodes       *Vec_Ptr_t;     // nodes whose SAT vars are assigned
    nRecycles              int32;      // the number of times SAT solver was recycled
    nCallsSince              int32;    // the number of calls since the last recycle
    vFanins       *Vec_Ptr_t;        // fanins of the CNF node
    // counter-examples
    vCex       *Vec_Int_t;           // the latest counter-example
    vVisits       *Vec_Int_t;        // temporary array for visited nodes  
    // SAT calls statistics
    nSatUnsat              int32;      // the number of proofs
    nSatSat              int32;        // the number of failure
    nSatUndec              int32;      // the number of timeouts
    nSatTotal              int32;      // the number of calls
    nCexLits              int32;
    // conflicts
    nConfUnsat              int32;     // conflicts in unsat problems
    nConfSat              int32;       // conflicts in sat problems
    nConfUndec              int32;     // conflicts in undec problems
    // runtime stats
    timeSatUnsat              int32;   // unsat
    timeSatSat              int32;     // sat
    timeSatUndec              int32;   // undecided
    timeTotal              int32;      // total runtime
}

// combinational simulation manager
   
type Cec_ManSim_t struct {
    // parameters
    pAig       *Gia_Man_t;           // the AIG to be used for simulation
    pPars    *Cec_ParSim_t;          // simulation parameters 
    nWords              int32;         // the number of simulation words
    // recycable memory
    pSimInfo             *int32;       // simulation information offsets
    pMems        *uint32;          // allocated simulaton memory
    nWordsAlloc              int32;    // the number of allocated entries
    nMems              int32;          // the number of used entries  
    nMemsMax              int32;       // the max number of used entries 
    MemFree              int32;        // next free entry
    nWordsOld              int32;      // the number of simulation words after previous relink
    // internal simulation info
    vCiSimInfo       *Vec_Ptr_t;     // CI simulation info  
    vCoSimInfo       *Vec_Ptr_t;     // CO simulation info  
    // counter examples
    pCexes           *any;         // counter-examples for each output
    iOut              int32;           // first failed output
    nOuts              int32;          // the number of failed outputs
    pCexComb       *Abc_Cex_t;       // counter-example for the first failed output
    pBestState       *Abc_Cex_t;     // the state that led to most of the refinements
    // scoring simulation patterns
    pScores             *int32;        // counters of refinement for each pattern
    // temporaries
    vClassOld       *Vec_Int_t;      // old class numbers
    vClassNew       *Vec_Int_t;      // new class numbers
    vClassTemp       *Vec_Int_t;     // temporary storage
    vRefinedC       *Vec_Int_t;      // refined const reprs
}

// combinational simulation manager
   
type Cec_ManFra_t struct {
    // parameters
    pAig       *Gia_Man_t;           // the AIG to be used for simulation
    pPars    *Cec_ParFra_t;          // SAT sweeping parameters 
    // simulation patterns
    vXorNodes       *Vec_Int_t;      // nodes used in speculative reduction
    nAllProved              int32;     // total number of proved nodes
    nAllDisproved              int32;  // total number of disproved nodes
    nAllFailed              int32;     // total number of failed nodes
    nAllProvedS              int32;    // total number of proved nodes
    nAllDisprovedS              int32; // total number of disproved nodes
    nAllFailedS              int32;    // total number of failed nodes
    // runtime stats
    timeSim          abctime;        // unsat
    timePat          abctime;        // unsat
    timeSat          abctime;        // sat
    timeTotal          abctime;      // total runtime
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== cecCorr.c ============================================================*/
// extern void                 Cec_ManRefinedClassPrintStats( Gia_Man_t * p, Vec_Str_t * vStatus, int iIter, abctime Time );

/*=== cecClass.c ============================================================*/
// extern int                  Cec_ManSimClassRemoveOne( Cec_ManSim_t * p, int i );

// extern int                  Cec_ManSimClassesPrepare( Cec_ManSim_t * p, int LevelMax );

// extern int                  Cec_ManSimClassesRefine( Cec_ManSim_t * p );

// extern int                  Cec_ManSimSimulateRound( Cec_ManSim_t * p, Vec_Ptr_t * vInfoCis, Vec_Ptr_t * vInfoCos );

/*=== cecIso.c ============================================================*/
// extern int *                Cec_ManDetectIsomorphism( Gia_Man_t * p );

/*=== cecMan.c ============================================================*/
// extern Cec_ManSat_t *       Cec_ManSatCreate( Gia_Man_t * pAig, Cec_ParSat_t * pPars );

// extern void                 Cec_ManSatPrintStats( Cec_ManSat_t * p );

// extern void                 Cec_ManSatStop( Cec_ManSat_t * p );

// extern Cec_ManPat_t *       Cec_ManPatStart();

// extern void                 Cec_ManPatPrintStats( Cec_ManPat_t * p );

// extern void                 Cec_ManPatStop( Cec_ManPat_t * p );

// extern Cec_ManSim_t *       Cec_ManSimStart( Gia_Man_t * pAig, Cec_ParSim_t *  pPars );
 
// extern void                 Cec_ManSimStop( Cec_ManSim_t * p );
  
// extern Cec_ManFra_t *       Cec_ManFraStart( Gia_Man_t * pAig, Cec_ParFra_t *  pPars );
  
// extern void                 Cec_ManFraStop( Cec_ManFra_t * p );

/*=== cecPat.c ============================================================*/
// extern void                 Cec_ManPatSavePattern( Cec_ManPat_t *  pPat, Cec_ManSat_t *  p, Gia_Obj_t * pObj );

// extern void                 Cec_ManPatSavePatternCSat( Cec_ManPat_t * pMan, Vec_Int_t * vPat );

// extern Vec_Ptr_t *          Cec_ManPatCollectPatterns( Cec_ManPat_t *  pMan, int nInputs, int nWords );

// extern Vec_Ptr_t *          Cec_ManPatPackPatterns( Vec_Int_t * vCexStore, int nInputs, int nRegs, int nWordsInit );

/*=== cecSeq.c ============================================================*/
// extern int                  Cec_ManSeqResimulate( Cec_ManSim_t * p, Vec_Ptr_t * vInfo );

// extern int                  Cec_ManSeqResimulateInfo( Gia_Man_t * pAig, Vec_Ptr_t * vSimInfo, Abc_Cex_t * pBestState, int fCheckMiter );

// extern void                 Cec_ManSeqDeriveInfoInitRandom( Vec_Ptr_t * vInfo, Gia_Man_t * pAig, Abc_Cex_t * pCex );

// extern int                  Cec_ManCountNonConstOutputs( Gia_Man_t * pAig );

// extern int                  Cec_ManCheckNonTrivialCands( Gia_Man_t * pAig );

/*=== cecSolve.c ============================================================*/
// extern int                  Cec_ObjSatVarValue( Cec_ManSat_t * p, Gia_Obj_t * pObj );

// extern void                 Cec_ManSatSolve( Cec_ManPat_t * pPat, Gia_Man_t * pAig, Cec_ParSat_t * pPars, Vec_Int_t * vIdsOrig, Vec_Int_t * vMiterPairs, Vec_Int_t * vEquivPairs, int f0Proved );

// extern void                 Cec_ManSatSolveCSat( Cec_ManPat_t * pPat, Gia_Man_t * pAig, Cec_ParSat_t * pPars );

// extern Vec_Str_t *          Cec_ManSatSolveSeq( Vec_Ptr_t * vPatts, Gia_Man_t * pAig, Cec_ParSat_t * pPars, int nRegs, int * pnPats );

// extern Vec_Int_t *          Cec_ManSatSolveMiter( Gia_Man_t * pAig, Cec_ParSat_t * pPars, Vec_Str_t ** pvStatus );

// extern int                  Cec_ManSatCheckNode( Cec_ManSat_t * p, Gia_Obj_t * pObj );

// extern int                  Cec_ManSatCheckNodeTwo( Cec_ManSat_t * p, Gia_Obj_t * pObj1, Gia_Obj_t * pObj2 );

// extern void                 Cec_ManSavePattern( Cec_ManSat_t * p, Gia_Obj_t * pObj1, Gia_Obj_t * pObj2 );

// extern Vec_Int_t *          Cec_ManSatReadCex( Cec_ManSat_t * p );

/*=== cecSolveG.c ============================================================*/
// extern void                 CecG_ManSatSolve( Cec_ManPat_t * pPat, Gia_Man_t * pAig, Cec_ParSat_t * pPars, int f0Proved );

/*=== ceFraeep.c ============================================================*/
// extern Gia_Man_t *          Cec_ManFraSpecReduction( Cec_ManFra_t * p );

// extern int                  Cec_ManFraClassesUpdate( Cec_ManFra_t * p, Cec_ManSim_t * pSim, Cec_ManPat_t * pPat, Gia_Man_t * pNew );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


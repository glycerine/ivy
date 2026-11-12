package abs

/**CFile****************************************************************

  FileName    [abs.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abs.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/gia/gia.h"
//#include "aig/gia/giaAig.h"
//#include "aig/saig/saig.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// abstraction parameters
   
type Abs_Par_t struct {
    nFramesMax            int32;         // maximum frames
    nFramesStart            int32;       // starting frame 
    nFramesPast            int32;        // overlap frames
    nConfLimit            int32;         // conflict limit
    nLearnedMax            int32;        // max number of learned clauses
    nLearnedStart            int32;      // max number of learned clauses
    nLearnedDelta            int32;      // delta increase of learned clauses
    nLearnedPerce            int32;      // percentage of clauses to leave
    nTimeOut            int32;           // timeout in seconds
    nRatioMin            int32;          // stop when less than this % of object is unabstracted
    nRatioMin2            int32;         // stop when less than this % of object is unabstracted during refinement
    nRatioMax            int32;          // restart when the number of abstracted object is more than this
    fUseTermVars            int32;       // use terminal variables
    fUseRollback            int32;       // use rollback to the starting number of frames
    fPropFanout            int32;        // propagate fanout implications
    fAddLayer            int32;          // refinement strategy by adding layers
    fNewRefine            int32;         // uses new refinement heuristics
    fUseSkip            int32;           // skip proving intermediate timeframes
    fUseSimple            int32;         // use simple CNF construction
    fSkipHash            int32;          // skip hashing CNF while unrolling
    fUseFullProof            int32;      // use full proof for UNSAT cores
    fDumpVabs            int32;          // dumps the abstracted model
    fDumpMabs            int32;          // dumps the original AIG with abstraction map
    fCallProver            int32;        // calls the prover
    fSimpProver            int32;        // calls simplification before prover
    pFileVabs          string;          // dumps the abstracted model into this file
    fVerbose            int32;           // verbose flag
    fVeryVerbose            int32;       // print additional information
    iFrame            int32;             // the number of frames covered
    iFrameProved            int32;       // the number of frames proved
    nFramesNoChange            int32;    // the number of last frames without changes
    nFramesNoChangeLim            int32; // the number of last frames without changes to dump abstraction
}

// old abstraction parameters
   
type Gia_ParAbs_t struct {
    Algo            int32;               // the algorithm to be used
    nFramesMax            int32;         // timeframes for PBA
    nConfMax            int32;           // conflicts for PBA
    fDynamic            int32;           // dynamic unfolding for PBA
    fConstr            int32;            // use constraints
    nFramesBmc            int32;         // timeframes for BMC
    nConfMaxBmc            int32;        // conflicts for BMC
    nStableMax            int32;         // the number of stable frames to quit
    nRatio            int32;             // ratio of flops to quit
    TimeOut            int32;            // approximate timeout in seconds
    TimeOutVT            int32;          // approximate timeout in seconds
    nBobPar            int32;            // Bob's parameter
    fUseBdds            int32;           // use BDDs to refine abstraction
    fUseDprove            int32;         // use 'dprove' to refine abstraction
    fUseStart            int32;          // use starting frame
    fVerbose            int32;           // verbose output
    fVeryVerbose            int32;       // printing additional information
    Status            int32;             // the problem status
    nFramesDone            int32;        // the number of frames covered
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Ga2_ObjOffset(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Vec_IntEntry(p.vMapping, Gia_ObjId(p, pObj));                                                         }
func Ga2_ObjLeaveNum(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Vec_IntEntry(p.vMapping, Ga2_ObjOffset(p, pObj));                                                     }
func Ga2_ObjLeavePtr(p *Gia_Man_t, pObj *Gia_Obj_t) *int32 { return Vec_IntEntryP(p.vMapping, Ga2_ObjOffset(p, pObj) + 1);                                                }
func Ga2_ObjTruth(p *Gia_Man_t, pObj *Gia_Obj_t) uint32 { return uint32(Vec_IntEntry(p.vMapping, Ga2_ObjOffset(p, pObj) + Ga2_ObjLeaveNum(p, pObj) + 1));            }
func Ga2_ObjRefNum(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return uint32(Vec_IntEntry(p.vMapping, Ga2_ObjOffset(p, pObj) + Ga2_ObjLeaveNum(p, pObj) + 2));            }
var spStatic_Ga2_ObjLeaves_v Vec_Int_t;
func Ga2_ObjLeaves(p *Gia_Man_t, pObj *Gia_Obj_t) *Vec_Int_t {  spStatic_Ga2_ObjLeaves_v.nSize = Ga2_ObjLeaveNum(p, pObj), spStatic_Ga2_ObjLeaves_v.pArray = Ga2_ObjLeavePtr(p, pObj); return &spStatic_Ga2_ObjLeaves_v;       }

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== abs.c =========================================================*/
/*=== absDup.c =========================================================*/
// extern Gia_Man_t *       Gia_ManDupAbsFlops( Gia_Man_t * p, Vec_Int_t * vFlopClasses );

// extern Gia_Man_t *       Gia_ManDupAbsGates( Gia_Man_t * p, Vec_Int_t * vGateClasses );

// extern void              Gia_ManGlaCollect( Gia_Man_t * p, Vec_Int_t * vGateClasses, Vec_Int_t ** pvPis, Vec_Int_t ** pvPPis, Vec_Int_t ** pvFlops, Vec_Int_t ** pvNodes );

// extern void              Gia_ManPrintFlopClasses( Gia_Man_t * p );

// extern void              Gia_ManPrintObjClasses( Gia_Man_t * p );

// extern void              Gia_ManPrintGateClasses( Gia_Man_t * p );

/*=== absGla.c =========================================================*/
// extern int               Gia_ManPerformGla( Gia_Man_t * p, Abs_Par_t * pPars );

/*=== absGlaOld.c =========================================================*/
// extern int               Gia_ManPerformGlaOld( Gia_Man_t * p, Abs_Par_t * pPars, int fStartVta );

/*=== absIter.c =========================================================*/
// extern Gia_Man_t *       Gia_ManShrinkGla( Gia_Man_t * p, int nFrameMax, int nTimeOut, int fUsePdr, int fUseSat, int fUseBdd, int fVerbose );

/*=== absPth.c =========================================================*/
// extern void              Gia_GlaProveAbsracted( Gia_Man_t * p, int fSimpProver, int fVerbose );

// extern void              Gia_GlaProveCancel( int fVerbose );

// extern int               Gia_GlaProveCheck( int fVerbose );

/*=== absVta.c =========================================================*/
// extern int               Gia_VtaPerform( Gia_Man_t * pAig, Abs_Par_t * pPars );

/*=== absUtil.c =========================================================*/
// extern void              Abs_ParSetDefaults( Abs_Par_t * p );

// extern Vec_Int_t *       Gia_VtaConvertToGla( Gia_Man_t * p, Vec_Int_t * vVta );

// extern Vec_Int_t *       Gia_VtaConvertFromGla( Gia_Man_t * p, Vec_Int_t * vGla, int nFrames );

// extern Vec_Int_t *       Gia_FlaConvertToGla( Gia_Man_t * p, Vec_Int_t * vFla );

// extern Vec_Int_t *       Gia_GlaConvertToFla( Gia_Man_t * p, Vec_Int_t * vGla );

// extern int               Gia_GlaCountFlops( Gia_Man_t * p, Vec_Int_t * vGla );

// extern int               Gia_GlaCountNodes( Gia_Man_t * p, Vec_Int_t * vGla );


/*=== absRpm.c =========================================================*/
// extern Gia_Man_t *       Abs_RpmPerform( Gia_Man_t * p, int nCutMax, int fVerbose, int fVeryVerbose );

/*=== absRpmOld.c =========================================================*/
// extern Gia_Man_t *       Abs_RpmPerformOld( Gia_Man_t * p, int fVerbose );

// extern void              Gia_ManAbsSetDefaultParams( Gia_ParAbs_t * p );

// extern Vec_Int_t *       Saig_ManCexAbstractionFlops( Aig_Man_t * p, Gia_ParAbs_t * pPars );


/*=== absOldCex.c ==========================================================*/
// extern Vec_Int_t *       Saig_ManCbaFilterFlops( Aig_Man_t * pAig, Abc_Cex_t * pAbsCex, Vec_Int_t * vFlopClasses, Vec_Int_t * vAbsFfsToAdd, int nFfsToSelect );

// extern Abc_Cex_t *       Saig_ManCbaFindCexCareBits( Aig_Man_t * pAig, Abc_Cex_t * pCex, int nInputs, int fVerbose );

/*=== absOldRef.c ==========================================================*/
// extern int               Gia_ManCexAbstractionRefine( Gia_Man_t * pGia, Abc_Cex_t * pCex, int nFfToAddMax, int fTryFour, int fSensePath, int fVerbose );

/*=== absOldSat.c ==========================================================*/
// extern Vec_Int_t *       Saig_ManExtendCounterExampleTest3( Aig_Man_t * pAig, int iFirstFlopPi, Abc_Cex_t * pCex, int fVerbose );

/*=== absOldSim.c ==========================================================*/
// extern Vec_Int_t *       Saig_ManExtendCounterExampleTest2( Aig_Man_t * p, int iFirstPi, Abc_Cex_t * pCex, int fVerbose );



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


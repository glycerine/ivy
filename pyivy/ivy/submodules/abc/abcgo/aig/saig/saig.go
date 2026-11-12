package saig

/**CFile****************************************************************

  FileName    [saig.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saig.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/aig/aig.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

   
type Sec_MtrStatus_t struct {
    nInputs         int32;      // the total number of inputs
    nNodes         int32;       // the total number of nodes
    nOutputs         int32;     // the total number of outputs
    nUnsat         int32;       // the number of UNSAT outputs
    nSat         int32;         // the number of SAT outputs
    nUndec         int32;       // the number of undecided outputs
    iOut         int32;         // the satisfied output
} 
 
   
type Saig_ParBbr_t struct {
    TimeLimit         int32;
    nBddMax         int32;
    nIterMax         int32;
    fPartition         int32;
    fReorder         int32;
    fReorderImage         int32;
    fVerbose         int32;
    fSilent         int32;
    fSkipOutCheck         int32;// skip output checking
    iFrame         int32;       // explored up to this frame
}


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Saig_ManPiNum(p *Aig_Man_t) int32 { return p.nTruePis;                     }
func Saig_ManPoNum(p *Aig_Man_t) int32 { return p.nTruePos;                     }
func Saig_ManCiNum(p *Aig_Man_t) int32 { return p.nTruePis + p.nRegs;          }
func Saig_ManCoNum(p *Aig_Man_t) int32 { return p.nTruePos + p.nRegs;          }
func Saig_ManRegNum(p *Aig_Man_t) int32 { return p.nRegs;                        }
func Saig_ManConstrNum(p *Aig_Man_t) int32 { return p.nConstrs;                     }
func Saig_ManLo(p *Aig_Man_t, i int32) *Aig_Obj_t { return (Aig_Obj_t *)Vec_PtrEntry(p.vCis, Saig_ManPiNum(p)+i);   }
func Saig_ManLi(p *Aig_Man_t, i int32) *Aig_Obj_t { return (Aig_Obj_t *)Vec_PtrEntry(p.vCos, Saig_ManPoNum(p)+i);   }

func Saig_ObjIsPi(p *Aig_Man_t, pObj *Aig_Obj_t) int32 { return Aig_ObjIsCi(pObj) && Aig_ObjCioId(pObj) < Saig_ManPiNum(p); }
func Saig_ObjIsPo(p *Aig_Man_t, pObj *Aig_Obj_t) int32 { return Aig_ObjIsCo(pObj) && Aig_ObjCioId(pObj) < Saig_ManPoNum(p); }
func Saig_ObjIsLo(p *Aig_Man_t, pObj *Aig_Obj_t) int32 { return Aig_ObjIsCi(pObj) && Aig_ObjCioId(pObj) >= Saig_ManPiNum(p); }
func Saig_ObjIsLi(p *Aig_Man_t, pObj *Aig_Obj_t) int32 { return Aig_ObjIsCo(pObj) && Aig_ObjCioId(pObj) >= Saig_ManPoNum(p); }
func Saig_ObjLoToLi(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { assert((Saig_ObjIsLo(p, pObj) != 0), "Saig_ObjIsLo(p, pObj)"); return (Aig_Obj_t *)Vec_PtrEntry(p.vCos, Saig_ManPoNum(p)+Aig_ObjCioId(pObj)-Saig_ManPiNum(p));   }
func Saig_ObjLiToLo(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { assert((Saig_ObjIsLi(p, pObj) != 0), "Saig_ObjIsLi(p, pObj)"); return (Aig_Obj_t *)Vec_PtrEntry(p.vCis, Saig_ManPiNum(p)+Aig_ObjCioId(pObj)-Saig_ManPoNum(p));   }
func Saig_ObjRegId(p *Aig_Man_t, pObj *Aig_Obj_t) int32 { if Saig_ObjIsLo(p, pObj) != 0 { return Aig_ObjCioId(pObj)-Saig_ManPiNum(p); }
 if Saig_ObjIsLi(p, pObj) != 0 { return Aig_ObjCioId(pObj)-Saig_ManPoNum(p); } else { assert(false, "0"); }
  return -1; }

// iterator over the primary inputs/outputs
// Original macro:
// #define Saig_ManForEachPi( p, pObj, i )                                           \
//     Vec_PtrForEachEntryStop( Aig_Obj_t *, p->vCis, pObj, i, Saig_ManPiNum(p) )

// Original macro:
// #define Saig_ManForEachPo( p, pObj, i )                                           \
//     Vec_PtrForEachEntryStop( Aig_Obj_t *, p->vCos, pObj, i, Saig_ManPoNum(p) )

// iterator over the latch inputs/outputs
// Original macro:
// #define Saig_ManForEachLo( p, pObj, i )                                           \
//     for ( i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p->vCis, i+Saig_ManPiNum(p))), 1); i++ )

// Original macro:
// #define Saig_ManForEachLi( p, pObj, i )                                           \
//     for ( i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p->vCos, i+Saig_ManPoNum(p))), 1); i++ )

// iterator over the latch input and outputs
// Original macro:
// #define Saig_ManForEachLiLo( p, pObjLi, pObjLo, i )                               \
//     for ( i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)    \
//         && (((pObjLo)=Saig_ManLo(p, i)), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== saigCone.c ==========================================================*/
// extern void              Saig_ManPrintCones( Aig_Man_t * p );

/*=== saigConstr.c ==========================================================*/
// extern Aig_Man_t *       Saig_ManDupUnfoldConstrs( Aig_Man_t * pAig );

// extern Aig_Man_t *       Saig_ManDupFoldConstrs( Aig_Man_t * pAig, Vec_Int_t * vConstrs );

// extern int               Saig_ManDetectConstrTest( Aig_Man_t * p );

// extern void              Saig_ManDetectConstrFuncTest( Aig_Man_t * p, int nFrames, int nConfs, int nProps, int fOldAlgo, int fVerbose );

/*=== saigConstr2.c ==========================================================*/
// extern Aig_Man_t *       Saig_ManDupFoldConstrsFunc( Aig_Man_t * pAig, int fCompl, int fVerbose );

// extern Aig_Man_t *       Saig_ManDupUnfoldConstrsFunc( Aig_Man_t * pAig, int nFrames, int nConfs, int nProps, int fOldAlgo, int fVerbose );

// -- jlong -- begin
// extern Aig_Man_t *       Saig_ManDupFoldConstrsFunc2( Aig_Man_t * pAig, int fCompl, int fVerbose, int typeII_cnt );

// extern Aig_Man_t *       Saig_ManDupUnfoldConstrsFunc2( Aig_Man_t * pAig, int nFrames, int nConfs, int nProps, int fOldAlgo, int fVerbose , int * typeII_cnt);

// --jlong -- end

/*=== saigDual.c ==========================================================*/
// extern Aig_Man_t *       Saig_ManDupDual( Aig_Man_t * pAig, Vec_Int_t * vDcFlops, int nDualPis, int fDualFfs, int fMiterFfs, int fComplPo, int fCheckZero, int fCheckOne );

// extern void              Saig_ManBlockPo( Aig_Man_t * pAig, int nCycles );

/*=== saigDup.c ==========================================================*/
// extern Aig_Man_t *       Saig_ManDupOrpos( Aig_Man_t * p );

// extern Aig_Man_t *       Saig_ManCreateEquivMiter( Aig_Man_t * pAig, Vec_Int_t * vPairs, int fAddOuts );

// extern Aig_Man_t *       Saig_ManDupAbstraction( Aig_Man_t * pAig, Vec_Int_t * vFlops );

// extern int               Saig_ManVerifyCex( Aig_Man_t * pAig, Abc_Cex_t * p );

// extern Abc_Cex_t *       Saig_ManExtendCex( Aig_Man_t * pAig, Abc_Cex_t * p );

// extern int               Saig_ManFindFailedPoCex( Aig_Man_t * pAig, Abc_Cex_t * p );

// extern Aig_Man_t *       Saig_ManDupWithPhase( Aig_Man_t * pAig, Vec_Int_t * vInit );

// extern Aig_Man_t *       Saig_ManDupCones( Aig_Man_t * pAig, int * pPos, int nPos );

/*=== saigHaig.c ==========================================================*/
// extern Aig_Man_t *       Saig_ManHaigRecord( Aig_Man_t * p, int nIters, int nSteps, int fRetimingOnly, int fAddBugs, int fUseCnf, int fVerbose );

/*=== saigInd.c ==========================================================*/
// extern int               Saig_ManInduction( Aig_Man_t * p, int nTimeOut, int nFramesMax, int nConfMax, int fUnique, int fUniqueAll, int fGetCex, int fVerbose, int fVeryVerbose );

/*=== saigIoa.c ==========================================================*/
// extern void              Saig_ManDumpBlif( Aig_Man_t * p, char * pFileName );

// extern Aig_Man_t *       Saig_ManReadBlif( char * pFileName );

/*=== saigIso.c ==========================================================*/
// extern Vec_Int_t *       Saig_ManFindIsoPerm( Aig_Man_t * pAig, int fVerbose );

// extern Aig_Man_t *       Saig_ManDupIsoCanonical( Aig_Man_t * pAig, int fVerbose );

// extern Aig_Man_t *       Saig_ManIsoReduce( Aig_Man_t * pAig, Vec_Ptr_t ** pvCosEquivs, int fVerbose );

/*=== saigIsoFast.c ==========================================================*/
// extern Vec_Vec_t *       Saig_IsoDetectFast( Aig_Man_t * pAig );

/*=== saigMiter.c ==========================================================*/
// extern Sec_MtrStatus_t   Sec_MiterStatus( Aig_Man_t * p );

// extern Aig_Man_t *       Saig_ManCreateMiter( Aig_Man_t * p1, Aig_Man_t * p2, int Oper );

// extern Aig_Man_t *       Saig_ManCreateMiterComb( Aig_Man_t * p1, Aig_Man_t * p2, int Oper );

// extern Aig_Man_t *       Saig_ManDualRail( Aig_Man_t * p, int fMiter );

// extern Aig_Man_t *       Saig_ManCreateMiterTwo( Aig_Man_t * pOld, Aig_Man_t * pNew, int nFrames );

// extern int               Saig_ManDemiterSimple( Aig_Man_t * p, Aig_Man_t ** ppAig0, Aig_Man_t ** ppAig1 );

// extern int               Saig_ManDemiterSimpleDiff( Aig_Man_t * p, Aig_Man_t ** ppAig0, Aig_Man_t ** ppAig1 );

// extern int               Saig_ManDemiterDual( Aig_Man_t * p, Aig_Man_t ** ppAig0, Aig_Man_t ** ppAig1 );

// extern int               Ssw_SecSpecialMiter( Aig_Man_t * p0, Aig_Man_t * p1, int nFrames, int fVerbose );

// extern int               Saig_ManDemiterNew( Aig_Man_t * pMan );

/*=== saigOutdec.c ==========================================================*/
// extern Aig_Man_t *       Saig_ManDecPropertyOutput( Aig_Man_t * pAig, int nLits, int fVerbose );

/*=== saigPhase.c ==========================================================*/
// extern Aig_Man_t *       Saig_ManPhaseAbstract( Aig_Man_t * p, Vec_Int_t * vInits, int nFrames, int nPref, int fIgnore, int fPrint, int fVerbose );

/*=== saigRetFwd.c ==========================================================*/
// extern void              Saig_ManMarkAutonomous( Aig_Man_t * p );

// extern Aig_Man_t *       Saig_ManRetimeForward( Aig_Man_t * p, int nMaxIters, int fVerbose );

/*=== saigRetMin.c ==========================================================*/
// extern Aig_Man_t *       Saig_ManRetimeDupForward( Aig_Man_t * p, Vec_Ptr_t * vCut );

// extern Aig_Man_t *       Saig_ManRetimeMinArea( Aig_Man_t * p, int nMaxIters, int fForwardOnly, int fBackwardOnly, int fInitial, int fVerbose );

/*=== saigRetStep.c ==========================================================*/
// extern int               Saig_ManRetimeSteps( Aig_Man_t * p, int nSteps, int fForward, int fAddBugs );

/*=== saigScl.c ==========================================================*/
// extern void              Saig_ManReportUselessRegisters( Aig_Man_t * pAig );

/*=== saigSimMv.c ==========================================================*/
// extern Vec_Ptr_t *       Saig_MvManSimulate( Aig_Man_t * pAig, int nFramesSymb, int nFramesSatur, int fVerbose, int fVeryVerbose );

/*=== saigStrSim.c ==========================================================*/
// extern Vec_Int_t *       Saig_StrSimPerformMatching( Aig_Man_t * p0, Aig_Man_t * p1, int nDist, int fVerbose, Aig_Man_t ** ppMiter );

/*=== saigSwitch.c ==========================================================*/
// extern Vec_Int_t *       Saig_ManComputeSwitchProb2s( Aig_Man_t * p, int nFrames, int nPref, int fProbOne );

/*=== saigSynch.c ==========================================================*/
// extern Aig_Man_t *       Saig_ManDupInitZero( Aig_Man_t * p );

/*=== saigTrans.c ==========================================================*/
// extern Aig_Man_t *       Saig_ManTimeframeSimplify( Aig_Man_t * pAig, int nFrames, int nFramesMax, int fInit, int fVerbose );

/*=== saigWnd.c ==========================================================*/
// extern Aig_Man_t *       Saig_ManWindowExtract( Aig_Man_t * p, Aig_Obj_t * pObj, int nDist );

// extern Aig_Man_t *       Saig_ManWindowInsert( Aig_Man_t * p, Aig_Obj_t * pObj, int nDist, Aig_Man_t * pWnd );

// extern Aig_Obj_t *       Saig_ManFindPivot( Aig_Man_t * p );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


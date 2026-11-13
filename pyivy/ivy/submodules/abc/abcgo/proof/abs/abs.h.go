package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

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

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

type Gia_Man_t = gia.Gia_Man_t
type Gia_Obj_t = gia.Gia_Obj_t
type Vec_Int_t = vec.Vec_Int_t

func Ga2_ObjOffset(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return vec.Vec_IntEntry(gia.Gia_ManMapping(p), int(gia.Gia_ObjId(p, obj)))
}
func Ga2_ObjLeaveNum(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return vec.Vec_IntEntry(gia.Gia_ManMapping(p), int(Ga2_ObjOffset(p, obj)))
}
func Ga2_ObjLeavePtr(p *Gia_Man_t, obj *Gia_Obj_t) *int32 {
	return vec.Vec_IntEntryP(gia.Gia_ManMapping(p), int(Ga2_ObjOffset(p, obj)+1))
}
func Ga2_ObjTruth(p *Gia_Man_t, obj *Gia_Obj_t) uint32 {
	return uint32(vec.Vec_IntEntry(gia.Gia_ManMapping(p), int(Ga2_ObjOffset(p, obj)+Ga2_ObjLeaveNum(p, obj)+1)))
}
func Ga2_ObjRefNum(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	return vec.Vec_IntEntry(gia.Gia_ManMapping(p), int(Ga2_ObjOffset(p, obj)+Ga2_ObjLeaveNum(p, obj)+2))
}

// Ga2_ObjLeaves borrows the selected cut's leaf slice from the manager.
func Ga2_ObjLeaves(p *Gia_Man_t, obj *Gia_Obj_t) *Vec_Int_t {
	begin := int(Ga2_ObjOffset(p, obj) + 1)
	count := int(Ga2_ObjLeaveNum(p, obj))
	return vec.Vec_IntAllocArray(vec.Vec_IntArray(gia.Gia_ManMapping(p))[begin:begin+count], count)
}

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

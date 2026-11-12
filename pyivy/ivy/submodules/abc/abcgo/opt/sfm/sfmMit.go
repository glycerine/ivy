package sfm

/**CFile****************************************************************

  FileName    [sfmMit.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Timing manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sfmMit.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sfmInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Sfm_Mit_t_ struct {
    Value1      int32;
    Value2      int32;
    Value3      int32;
    Value4      int32;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Temporary place holders.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_MitStart(pLib *Mio_Library_t, pScl *SC_Lib, pExt *Scl_Con_t, pNtk *Abc_Ntk_t, DeltaCrit int32) *Sfm_Mit_t { return nil; }
func Sfm_MitStop(p *Sfm_Mit_t) {}
func Sfm_MitReadNtkDelay(p *Sfm_Mit_t) int32 { return 0;}
func Sfm_MitReadNtkMinSlack(p *Sfm_Mit_t) int32 { return 0;}
func Sfm_MitReadObjDelay(p *Sfm_Mit_t, iObj int32) int32 { return 0;}
func Sfm_MitTransferLoad(p *Sfm_Mit_t, pNew *Abc_Obj_t, pOld *Abc_Obj_t) {};
func Sfm_MitTimingGrow(p *Sfm_Mit_t) {};
func Sfm_MitUpdateLoad(p *Sfm_Mit_t, vTimeNodes *Vec_Int_t, fAdd int32) {}
func Sfm_MitUpdateTiming(p *Sfm_Mit_t, vTimeNodes *Vec_Int_t) {}
func Sfm_MitSortArrayByArrival(p *Sfm_Mit_t, vNodes *Vec_Int_t, iPivot int32) int32 { return 0;}
func Sfm_MitPriorityNodes(p *Sfm_Mit_t, vCands *Vec_Int_t, Window int32) int32 { return 0;}
func Sfm_MitNodeIsNonCritical(p *Sfm_Mit_t, pPivot *Abc_Obj_t, pNode *Abc_Obj_t) int32 { return 0;}
func Sfm_MitEvalRemapping(p *Sfm_Mit_t, vMffc *Vec_Int_t, pObj *Abc_Obj_t, vFanins *Vec_Int_t, vMap *Vec_Int_t, pGate1 *Mio_Gate_t, pFans1 string, pGate2 *Mio_Gate_t, pFans2 string) int32 { return 0;}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


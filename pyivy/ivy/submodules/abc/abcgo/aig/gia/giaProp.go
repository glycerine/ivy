package gia

/**CFile****************************************************************

  FileName    [giaProp.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Constraint propagation on the AIG.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaProp.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const GIA_SAT_SHIFT = 12
#define GIA_ROOT_MASK 
#define GIA_PATH00_MASK
#define GIA_PATH10_MASK
#define GIA_PATH20_MASK
#define GIA_PATH30_MASK
#define GIA_PATH00_MASK
#define GIA_PATH10_MASK
#define GIA_PATH20_MASK
#define GIA_PATH30_MASK

func Gia_SatObjIsRoot(p *Gia_Obj_t) int32 { return 0; }
func Gia_SatObjXorRoot(p *Gia_Obj_t) int32 { return 0; }


func Gia_SatObjIsAssigned(p *Gia_Obj_t) int32 { return 0; }
func Gia_SatObjIsHeld(p *Gia_Obj_t) int32 { return 0; }
func Gia_SatObjValue(p *Gia_Obj_t) int32 { return 0; }


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checks if the give cut is satisfied.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SatPathCheckCutSat_rec(p *Gia_Obj_t, fCompl int32) int32 {
    if Gia_SatObjIsRoot(p) != 0 {
        return Gia_ObjIsAssigned(p) && Gia_SatObjValue(p) == fCompl; }
    if Gia_SatObjPath0(p) && !(Gia_SatPathCheckCutSat_rec( Gia_ObjFanin0(p), fCompl ^ Gia_ObjFaninC0(p) ) != 0) {
        return 0; }
    if Gia_SatObjPath1(p) && !(Gia_SatPathCheckCutSat_rec( Gia_ObjFanin1(p), fCompl ^ Gia_ObjFaninC1(p) ) != 0) {
        return 0; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Checks if the give cut is satisfied.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SatPathCheckCutSat(p *Gia_Obj_t) int32 {
    var RetValue int32;
    assert((Gia_SatObjIsRoot(p) != 0), "Gia_SatObjIsRoot(p)");
    Gia_SatObjXorRoot(p);
    RetValue = Gia_SatPathCheckCutSat_rec( p );
    Gia_SatObjXorRoot(p);
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Unbinds literals on the path.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SatPathUnbind_rec(p *Gia_Obj_t) int32 {
}

/**Function*************************************************************

  Synopsis    [Creates a feasible path from the node to a terminal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SatPathStart_rec(p *Gia_Obj_t, fDiffs int32, fCompl int32) int32 {
    if Gia_SatObjIsRoot(p) != 0 {
        return fDiffs && (!Gia_ObjIsAssigned(p) || Gia_SatObjValue(p) != fCompl); }
    if fCompl == 0 {
        if (Gia_SatPathStart_rec( Gia_ObjFanin0(p), fDiffs + !Gia_SatObjPath0(p), fCompl ^ Gia_ObjFaninC0(p) ) != 0) &&
             (Gia_SatPathStart_rec( Gia_ObjFanin1(p), fDiffs + !Gia_SatObjPath1(p), fCompl ^ Gia_ObjFaninC1(p) ) != 0) {
            return Gia_ObjSetDraftPath0(p) + Gia_ObjSetDraftPath1(p); }
    
    } else {
        if Gia_SatPathStart_rec( Gia_ObjFanin0(p), fDiffs + !Gia_SatObjPath0(p), fCompl ^ Gia_ObjFaninC0(p) ) != 0 {
            Gia_ObjUnsetDraftPath1(p);
            return Gia_ObjSetDraftPath0(p);
        }
        if Gia_SatPathStart_rec( Gia_ObjFanin1(p), fDiffs + !Gia_SatObjPath1(p), fCompl ^ Gia_ObjFaninC1(p) ) != 0 {
            Gia_ObjUnsetDraftPath0(p);
            return Gia_ObjSetDraftPath1(p);
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Creates a feasible path from the node to a terminal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_SatPathStart(p *Gia_Obj_t) int32 {
    var RetValue int32;
    assert((Gia_SatObjIsRoot(p) != 0), "Gia_SatObjIsRoot(p)");
    Gia_SatObjXorRoot(p);
    RetValue = Gia_SatPathStart_rec( p, 0, 0 );
    Gia_SatObjXorRoot(p);
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


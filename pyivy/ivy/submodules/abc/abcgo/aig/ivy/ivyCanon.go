package ivy

/**CFile****************************************************************

  FileName    [ivyCanon.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Finding canonical form of objects.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyCanon.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Ivy_Obj_t * Ivy_TableLookupPair_rec( Ivy_Man_t * p, Ivy_Obj_t * pObj0, Ivy_Obj_t * pObj1, int fCompl0, int fCompl1, Ivy_Type_t Type );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates the canonical form of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_CanonPair_rec(p *Ivy_Man_t, pGhost *Ivy_Obj_t) *Ivy_Obj_t {
    var pResult, pLat0, pLat1 *Ivy_Obj_t;
    var Init, Init0, Init1 Ivy_Init_t;
    var fCompl0, fCompl1 int32;
    var Type Ivy_Type_t;
    assert(Ivy_ObjIsNode(pGhost), "Ivy_ObjIsNode(pGhost)");
    assert(Ivy_ObjIsAnd(pGhost) || (!Ivy_ObjFaninC0(pGhost) && !Ivy_ObjFaninC1(pGhost)), "Ivy_ObjIsAnd(pGhost) || (!Ivy_ObjFaninC0(pGhost) && !Ivy_ObjFaninC1(pGhost))");
    assert(Ivy_ObjFaninId0(pGhost) != 0 && Ivy_ObjFaninId1(pGhost) != 0, "Ivy_ObjFaninId0(pGhost) != 0 && Ivy_ObjFaninId1(pGhost) != 0");
    // consider the case when the pair is canonical
    if !Ivy_ObjIsLatch(Ivy_ObjFanin0(pGhost)) || !Ivy_ObjIsLatch(Ivy_ObjFanin1(pGhost)) {
        if pResult = Ivy_TableLookup( p, pGhost ) {
            return pResult; }
        return Ivy_ObjCreate( p, pGhost );
    }
    /// remember the latches
    pLat0 = Ivy_ObjFanin0(pGhost);
    pLat1 = Ivy_ObjFanin1(pGhost);
    // remember type and compls
    Type = Ivy_ObjType(pGhost);
    fCompl0 = Ivy_ObjFaninC0(pGhost);
    fCompl1 = Ivy_ObjFaninC1(pGhost);
    // call recursively
    pResult = Ivy_Oper( p, Ivy_NotCond(Ivy_ObjFanin0(pLat0), fCompl0), Ivy_NotCond(Ivy_ObjFanin0(pLat1), fCompl1), Type );
    // build latch on top of this
    Init0 = Ivy_InitNotCond( Ivy_ObjInit(pLat0), fCompl0 );
    Init1 = Ivy_InitNotCond( Ivy_ObjInit(pLat1), fCompl1 );
    if Type == IVY_AND {
        Init = Ivy_InitAnd(Init0, Init1);
    } else {
        Init = Ivy_InitExor(Init0, Init1);
    }
    return Ivy_Latch( p, pResult, Init );
}

/**Function*************************************************************

  Synopsis    [Creates the canonical form of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_CanonAnd(p *Ivy_Man_t, pObj0 *Ivy_Obj_t, pObj1 *Ivy_Obj_t) *Ivy_Obj_t {
    var pGhost, pResult *Ivy_Obj_t;
    pGhost = Ivy_ObjCreateGhost( p, pObj0, pObj1, IVY_AND, IVY_INIT_NONE );
    pResult = Ivy_CanonPair_rec( p, pGhost );
    return pResult;
}

/**Function*************************************************************

  Synopsis    [Creates the canonical form of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_CanonExor(p *Ivy_Man_t, pObj0 *Ivy_Obj_t, pObj1 *Ivy_Obj_t) *Ivy_Obj_t {
    var pGhost, pResult *Ivy_Obj_t;
    var fCompl int32  = Ivy_IsComplement(pObj0) ^ Ivy_IsComplement(pObj1);
    pObj0 = Ivy_Regular(pObj0);
    pObj1 = Ivy_Regular(pObj1);
    pGhost = Ivy_ObjCreateGhost( p, pObj0, pObj1, IVY_EXOR, IVY_INIT_NONE );
    pResult = Ivy_CanonPair_rec( p, pGhost );
    return Ivy_NotCond( pResult, fCompl );
}

/**Function*************************************************************

  Synopsis    [Creates the canonical form of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_CanonLatch(p *Ivy_Man_t, pObj *Ivy_Obj_t, Init Ivy_Init_t) *Ivy_Obj_t {
    var pGhost, pResult *Ivy_Obj_t;
    var fCompl int32  = Ivy_IsComplement(pObj);
    pObj = Ivy_Regular(pObj);
    pGhost = Ivy_ObjCreateGhost( p, pObj, nil, IVY_LATCH, Ivy_InitNotCond(Init, fCompl) );
    pResult = Ivy_TableLookup( p, pGhost );
    if pResult == nil {
        pResult = Ivy_ObjCreate( p, pGhost ); }
    return Ivy_NotCond( pResult, fCompl );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


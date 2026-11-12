package ivy

/**CFile****************************************************************

  FileName    [ivyFanout.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Representation of the fanouts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyFanout.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// getting hold of the next fanout of the node
func Ivy_ObjNextFanout(pObj *Ivy_Obj_t, pFanout *Ivy_Obj_t) *Ivy_Obj_t {
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(!Ivy_IsComplement(pFanout), "!Ivy_IsComplement(pFanout)");
    if pFanout == nil {
        return nil; }
    if Ivy_ObjFanin0(pFanout) == pObj {
        return pFanout.pNextFan0; }
    assert(Ivy_ObjFanin1(pFanout) == pObj, "Ivy_ObjFanin1(pFanout) == pObj");
    return pFanout.pNextFan1;
}

// getting hold of the previous fanout of the node
func Ivy_ObjPrevFanout(pObj *Ivy_Obj_t, pFanout *Ivy_Obj_t) *Ivy_Obj_t {
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(!Ivy_IsComplement(pFanout), "!Ivy_IsComplement(pFanout)");
    if pFanout == nil {
        return nil; }
    if Ivy_ObjFanin0(pFanout) == pObj {
        return pFanout.pPrevFan0; }
    assert(Ivy_ObjFanin1(pFanout) == pObj, "Ivy_ObjFanin1(pFanout) == pObj");
    return pFanout.pPrevFan1;
}

// getting hold of the place where the next fanout will be attached
func Ivy_ObjNextFanoutPlace(pObj *Ivy_Obj_t, pFanout *Ivy_Obj_t) **Ivy_Obj_t {
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(!Ivy_IsComplement(pFanout), "!Ivy_IsComplement(pFanout)");
    if Ivy_ObjFanin0(pFanout) == pObj {
        return &pFanout.pNextFan0; }
    assert(Ivy_ObjFanin1(pFanout) == pObj, "Ivy_ObjFanin1(pFanout) == pObj");
    return &pFanout.pNextFan1;
}

// getting hold of the place where the next fanout will be attached
func Ivy_ObjPrevFanoutPlace(pObj *Ivy_Obj_t, pFanout *Ivy_Obj_t) **Ivy_Obj_t {
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(!Ivy_IsComplement(pFanout), "!Ivy_IsComplement(pFanout)");
    if Ivy_ObjFanin0(pFanout) == pObj {
        return &pFanout.pPrevFan0; }
    assert(Ivy_ObjFanin1(pFanout) == pObj, "Ivy_ObjFanin1(pFanout) == pObj");
    return &pFanout.pPrevFan1;
}

// getting hold of the place where the next fanout will be attached
func Ivy_ObjPrevNextFanoutPlace(pObj *Ivy_Obj_t, pFanout *Ivy_Obj_t) **Ivy_Obj_t {
    var pTemp *Ivy_Obj_t;
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(!Ivy_IsComplement(pFanout), "!Ivy_IsComplement(pFanout)");
    pTemp = Ivy_ObjPrevFanout(pObj, pFanout);
    if pTemp == nil {
        return &pObj.pFanout; }
    if Ivy_ObjFanin0(pTemp) == pObj {
        return &pTemp.pNextFan0; }
    assert(Ivy_ObjFanin1(pTemp) == pObj, "Ivy_ObjFanin1(pTemp) == pObj");
    return &pTemp.pNextFan1;
}

// getting hold of the place where the next fanout will be attached
func Ivy_ObjNextPrevFanoutPlace(pObj *Ivy_Obj_t, pFanout *Ivy_Obj_t) **Ivy_Obj_t {
    var pTemp *Ivy_Obj_t;
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(!Ivy_IsComplement(pFanout), "!Ivy_IsComplement(pFanout)");
    pTemp = Ivy_ObjNextFanout(pObj, pFanout);
    if pTemp == nil {
        return nil; }
    if Ivy_ObjFanin0(pTemp) == pObj {
        return &pTemp.pPrevFan0; }
    assert(Ivy_ObjFanin1(pTemp) == pObj, "Ivy_ObjFanin1(pTemp) == pObj");
    return &pTemp.pPrevFan1;
}

// iterator through the fanouts of the node
// Original macro:
// #define Ivy_ObjForEachFanoutInt( pObj, pFanout )                 \
//     for ( pFanout = (pObj)->pFanout; pFanout;                    \
//           pFanout = Ivy_ObjNextFanout(pObj, pFanout) )


// safe iterator through the fanouts of the node
// Original macro:
// #define Ivy_ObjForEachFanoutIntSafe( pObj, pFanout, pFanout2 )   \
//     for ( pFanout  = (pObj)->pFanout,                            \
//           pFanout2 = Ivy_ObjNextFanout(pObj, pFanout);           \
//           pFanout;                                               \
//           pFanout  = pFanout2,                                   \
//           pFanout2 = Ivy_ObjNextFanout(pObj, pFanout) )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the fanout representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManStartFanout(p *Ivy_Man_t) {
    var pObj *Ivy_Obj_t;
    var i int32;
    assert(!p.fFanout, "!p.fFanout");
    p.fFanout = 1;
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Ivy_ObjFanin0(pObj) {
            Ivy_ObjAddFanout( p, Ivy_ObjFanin0(pObj), pObj ); }
        if Ivy_ObjFanin1(pObj) {
            Ivy_ObjAddFanout( p, Ivy_ObjFanin1(pObj), pObj ); }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Stops the fanout representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManStopFanout(p *Ivy_Man_t) {
    var pObj *Ivy_Obj_t;
    var i int32;
    assert(p.fFanout, "p.fFanout");
    p.fFanout = 0;
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.pFanout = pObj.pNextFan0 = pObj.pNextFan1 = pObj.pPrevFan0 = pObj.pPrevFan1 = nil; }
    }
}

/**Function*************************************************************

  Synopsis    [Add the fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjAddFanout(p *Ivy_Man_t, pFanin *Ivy_Obj_t, pFanout *Ivy_Obj_t) {
    assert(p.fFanout, "p.fFanout");
    if pFanin.pFanout {
        *Ivy_ObjNextFanoutPlace(pFanin, pFanout) = pFanin.pFanout;
        *Ivy_ObjPrevFanoutPlace(pFanin, pFanin.pFanout) = pFanout;
    }
    pFanin.pFanout = pFanout;
}

/**Function*************************************************************

  Synopsis    [Removes the fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjDeleteFanout(p *Ivy_Man_t, pFanin *Ivy_Obj_t, pFanout *Ivy_Obj_t) {
    var ppPlace1, ppPlace2, ppPlaceN **Ivy_Obj_t;
    assert(pFanin.pFanout != nil, "pFanin.pFanout != nil");

    ppPlace1 = Ivy_ObjNextFanoutPlace(pFanin, pFanout);
    ppPlaceN = Ivy_ObjPrevNextFanoutPlace(pFanin, pFanout);
    assert(*ppPlaceN == pFanout, "*ppPlaceN == pFanout");
    if ppPlaceN {
        *ppPlaceN = *ppPlace1; }

    ppPlace2 = Ivy_ObjPrevFanoutPlace(pFanin, pFanout);
    ppPlaceN = Ivy_ObjNextPrevFanoutPlace(pFanin, pFanout);
    assert(ppPlaceN == nil || *ppPlaceN == pFanout, "ppPlaceN == nil || *ppPlaceN == pFanout");
    if ppPlaceN {
        *ppPlaceN = *ppPlace2; }

    *ppPlace1 = nil;
    *ppPlace2 = nil;
}

/**Function*************************************************************

  Synopsis    [Replaces the fanout of pOld to be pFanoutNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjPatchFanout(p *Ivy_Man_t, pFanin *Ivy_Obj_t, pFanoutOld *Ivy_Obj_t, pFanoutNew *Ivy_Obj_t) {
    var ppPlace **Ivy_Obj_t;
    ppPlace = Ivy_ObjPrevNextFanoutPlace(pFanin, pFanoutOld);
    assert(*ppPlace == pFanoutOld, "*ppPlace == pFanoutOld");
    if ppPlace {
        *ppPlace = pFanoutNew; }
    ppPlace = Ivy_ObjNextPrevFanoutPlace(pFanin, pFanoutOld);
    assert(ppPlace == nil || *ppPlace == pFanoutOld, "ppPlace == nil || *ppPlace == pFanoutOld");
    if ppPlace {
        *ppPlace = pFanoutNew; }
    // assuming that pFanoutNew already points to the next fanout
}

/**Function*************************************************************

  Synopsis    [Starts iteration through the fanouts.]

  Description [Copies the currently available fanouts into the array.]
               
  SideEffects [Can be used while the fanouts are being removed.]

  SeeAlso     []

***********************************************************************/
func Ivy_ObjCollectFanouts(p *Ivy_Man_t, pObj *Ivy_Obj_t, vArray *Vec_Ptr_t) {
    var pFanout *Ivy_Obj_t;
    assert(p.fFanout, "p.fFanout");
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    Vec_PtrClear( vArray );
    // inlining of macro Ivy_ObjForEachFanoutInt
    for  pFanout = (pObj).pFanout; (pFanout != nil);                              pFanout = Ivy_ObjNextFanout(pObj, pFanout)  {
        Vec_PtrPush( vArray, pFanout );
    }
}

/**Function*************************************************************

  Synopsis    [Reads one fanout.]

  Description [Returns fanout if there is only one fanout.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjReadFirstFanout(p *Ivy_Man_t, pObj *Ivy_Obj_t) *Ivy_Obj_t {
    return pObj.pFanout;
}

/**Function*************************************************************

  Synopsis    [Reads one fanout.]

  Description [Returns fanout if there is only one fanout.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjFanoutNum(p *Ivy_Man_t, pObj *Ivy_Obj_t) int32 {
    var pFanout *Ivy_Obj_t;
    var Counter int32;
    // inlining of macro Ivy_ObjForEachFanoutInt
    for  pFanout = (pObj).pFanout; (pFanout != nil);                              pFanout = Ivy_ObjNextFanout(pObj, pFanout)  {
        Counter++;
    }
    return Counter;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


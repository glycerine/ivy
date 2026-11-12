package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaAig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Transformation between AIG manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaAig.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "giaAig.h"
//#include "proof/fra/fra.h"
//#include "proof/dch/dch.h"
//#include "opt/dar/dar.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Gia_ObjChild0Copy(pObj *Aig_Obj_t) int32 { return Abc_LitNotCond( Aig_ObjFanin0(pObj).iData, Aig_ObjFaninC0(pObj) ); }
func Gia_ObjChild1Copy(pObj *Aig_Obj_t) int32 { return Abc_LitNotCond( Aig_ObjFanin1(pObj).iData, Aig_ObjFaninC1(pObj) ); }

func Gia_ObjChild0Copy2(ppNodes **Aig_Obj_t, pObj *Gia_Obj_t, Id int32) *Aig_Obj_t { return Aig_NotCond( ppNodes[Gia_ObjFaninId0(pObj, Id)], Gia_ObjFaninC0(pObj) ); }
func Gia_ObjChild1Copy2(ppNodes **Aig_Obj_t, pObj *Gia_Obj_t, Id int32) *Aig_Obj_t { return Aig_NotCond( ppNodes[Gia_ObjFaninId1(pObj, Id)], Gia_ObjFaninC1(pObj) ); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFromAig_rec(pNew *Gia_Man_t, p *Aig_Man_t, pObj *Aig_Obj_t) {
    var pNext *Aig_Obj_t;
    if pObj.iData {
        return; }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Gia_ManFromAig_rec( pNew, p, Aig_ObjFanin0(pObj) );
    Gia_ManFromAig_rec( pNew, p, Aig_ObjFanin1(pObj) );
    pObj.iData = Gia_ManAppendAnd( pNew, Gia_ObjChild0Copy(pObj), Gia_ObjChild1Copy(pObj) );
    if p.pEquivs && (pNext = Aig_ObjEquiv(p, pObj)) {
        var iObjNew, iNextNew int32;
        Gia_ManFromAig_rec( pNew, p, pNext );
        iObjNew  = Abc_Lit2Var(pObj.iData);
        iNextNew = Abc_Lit2Var(pNext.iData);
        if pNew.pNexts {
            pNew.pNexts[iObjNew] = iNextNew; }        
    }
}
func Gia_ManFromAig(p *Aig_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // create the new manager
    pNew = Gia_ManStart( Aig_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nConstrs = p.nConstrs;
    // create room to store equivalences
    if p.pEquivs {
        // inlining of macro ABC_CALLOC
        pNew.pNexts = ((int *) calloc((size_t)(Aig_ManObjNum(p)), sizeof(int))); }
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).iData = 1;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.iData = Gia_ManAppendCi( pNew );
    }
    // add logic for the POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Gia_ManFromAig_rec( pNew, p, Aig_ObjFanin0(pObj) );
    }        
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjChild0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    if pNew.pNexts {
        Gia_ManDeriveReprs( pNew ); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Checks integrity of choice nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckChoices_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if !pObj || !Gia_ObjIsAnd(pObj) || pObj.fPhase {
        return; }
    pObj.fPhase = 1;
    Gia_ManCheckChoices_rec( p, Gia_ObjFanin0(pObj) );
    Gia_ManCheckChoices_rec( p, Gia_ObjFanin1(pObj) );
    Gia_ManCheckChoices_rec( p, Gia_ObjSiblObj(p, Gia_ObjId(p, pObj)) );
}
func Gia_ManCheckChoices(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i, fFound int32;
    Gia_ManCleanPhase( p );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManCheckChoices_rec( p, Gia_ObjFanin0(pObj) );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if !pObj.fPhase {
            fmt.Printf("Object %d is dangling.\n", i);
fFound = 1; }
    }
    if !(fFound != 0) {
        fmt.Printf("There are no dangling objects.\n"); }
    Gia_ManCleanPhase( p );
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFromAigChoices_rec(pNew *Gia_Man_t, p *Aig_Man_t, pObj *Aig_Obj_t) {
    if pObj == nil || pObj.iData {
        return; }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Gia_ManFromAigChoices_rec( pNew, p, Aig_ObjFanin0(pObj) );
    Gia_ManFromAigChoices_rec( pNew, p, Aig_ObjFanin1(pObj) );
    Gia_ManFromAigChoices_rec( pNew, p, Aig_ObjEquiv(p, pObj) );
    pObj.iData = Gia_ManAppendAnd( pNew, Gia_ObjChild0Copy(pObj), Gia_ObjChild1Copy(pObj) );
    if Aig_ObjEquiv(p, pObj) {
        var iObjNew, iNextNew int32;
        iObjNew  = Abc_Lit2Var(pObj.iData);
        iNextNew = Abc_Lit2Var(Aig_ObjEquiv(p, pObj).iData);
        assert(iObjNew > iNextNew, "iObjNew > iNextNew");
        assert(Gia_ObjIsAnd(Gia_ManObj(pNew, iNextNew)), "Gia_ObjIsAnd(Gia_ManObj(pNew, iNextNew))");
        pNew.pSibls[iObjNew] = iNextNew;        
    }
}
func Gia_ManFromAigChoices(p *Aig_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    assert(p.pEquivs != nil, "p.pEquivs != nil");
    // create the new manager
    pNew = Gia_ManStart( Aig_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nConstrs = p.nConstrs;
    // create room to store equivalences
    // inlining of macro ABC_CALLOC
    pNew.pSibls = ((int *) calloc((size_t)(Aig_ManObjNum(p)), sizeof(int)));
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).iData = 1;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.iData = Gia_ManAppendCi( pNew );
    }
    // add logic for the POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Gia_ManFromAigChoices_rec( pNew, p, Aig_ObjFanin0(pObj) );
    }        
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjChild0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    //assert( Gia_ManObjNum(pNew) == Aig_ManObjNum(p) );
    //Gia_ManCheckChoices( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFromAigSimple(p *Aig_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // create the new manager
    pNew = Gia_ManStart( Aig_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nConstrs = p.nConstrs;
    // create the PIs
    Aig_ManCleanData( p );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsAnd(pObj) {
            pObj.iData = Gia_ManAppendAnd( pNew, Gia_ObjChild0Copy(pObj), Gia_ObjChild1Copy(pObj) );
        } else if Aig_ObjIsCi(pObj) {
            pObj.iData = Gia_ManAppendCi( pNew );
        } else if Aig_ObjIsCo(pObj) {
            pObj.iData = Gia_ManAppendCo( pNew, Gia_ObjChild0Copy(pObj) );
        } else if Aig_ObjIsConst1(pObj) {
            pObj.iData = 1;
        } else {
            assert(false, "0"); }
    }
    }
    Gia_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Handles choices as additional combinational outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFromAigSwitch(p *Aig_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // create the new manager
    pNew = Gia_ManStart( Aig_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nConstrs = p.nConstrs;
    // create the PIs
    Aig_ManCleanData( p );
    Aig_ManConst1(p).iData = 1;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        pObj.iData = Gia_ManAppendCi( pNew );
    }
    // add POs corresponding to the nodes with choices
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else
        if Aig_ObjRefs(pObj) == 0 {
            Gia_ManFromAig_rec( pNew, p, pObj );        
            Gia_ManAppendCo( pNew, pObj.iData );
        }
    }
    // add logic for the POs
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Gia_ManFromAig_rec( pNew, p, Aig_ObjFanin0(pObj) );
    }        
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        pObj.iData = Gia_ManAppendCo( pNew, Gia_ObjChild0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManToAig_rec(pNew *Aig_Man_t, ppNodes **Aig_Obj_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    var pNext *Gia_Obj_t;
    if ppNodes[Gia_ObjId(p, pObj)] {
        return; }
    if Gia_ObjIsCi(pObj) {
        ppNodes[Gia_ObjId(p, pObj)] = Aig_ObjCreateCi( pNew );
    } else {
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        Gia_ManToAig_rec( pNew, ppNodes, p, Gia_ObjFanin0(pObj) );
        Gia_ManToAig_rec( pNew, ppNodes, p, Gia_ObjFanin1(pObj) );
        ppNodes[Gia_ObjId(p, pObj)] = Aig_And( pNew, Gia_ObjChild0Copy2(ppNodes, pObj, Gia_ObjId(p, pObj)), Gia_ObjChild1Copy2(ppNodes, pObj, Gia_ObjId(p, pObj)) );
    }
    if pNew.pEquivs && (pNext = Gia_ObjNextObj(p, Gia_ObjId(p, pObj))) {
        var pObjNew, pNextNew *Aig_Obj_t;
        Gia_ManToAig_rec( pNew, ppNodes, p, pNext );
        pObjNew  = ppNodes[Gia_ObjId(p, pObj)];
        pNextNew = ppNodes[Gia_ObjId(p, pNext)];
        if pNew.pEquivs {
            pNew.pEquivs[Aig_Regular(pObjNew).Id] = Aig_Regular(pNextNew); }        
    }
}
func Gia_ManToAig(p *Gia_Man_t, fChoices int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var ppNodes **Aig_Obj_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(!(fChoices != 0) || (p.pNexts && p.pReprs), "!fChoices || (p.pNexts && p.pReprs)");
    // create the new manager
    pNew = Aig_ManStart( Gia_ManAndNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nConstrs = p.nConstrs;
//    pNew->pSpec = Abc_UtilStrsav( p->pName );
    // duplicate representation of choice nodes
    if fChoices != 0 {
        // inlining of macro ABC_CALLOC
        pNew.pEquivs = ((Aig_Obj_t * *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Aig_Obj_t *))); }
    // create the PIs
    // inlining of macro ABC_CALLOC
    ppNodes = ((Aig_Obj_t * *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Aig_Obj_t *)));
    ppNodes[0] = Aig_ManConst0(pNew);
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        ppNodes[Gia_ObjId(p, pObj)] = Aig_ObjCreateCi( pNew );
    }
    // transfer level
    if p.vLevels {
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Aig_ObjSetLevel( ppNodes[Gia_ObjId(p, pObj)], Gia_ObjLevel(p, pObj) );
    } }
    // add logic for the POs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManToAig_rec( pNew, ppNodes, p, Gia_ObjFanin0(pObj) );        
        ppNodes[Gia_ObjId(p, pObj)] = Aig_ObjCreateCo( pNew, Gia_ObjChild0Copy2(ppNodes, pObj, Gia_ObjId(p, pObj)) );
    }
    Aig_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // inlining of macro ABC_FREE
    ((ppNodes) ? (free((char *) (ppNodes)), (ppNodes) = 0) : 0);
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManToAigSkip(p *Gia_Man_t, nOutDelta int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var ppNodes **Aig_Obj_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(p.pNexts == nil && p.pReprs == nil, "p.pNexts == nil && p.pReprs == nil");
    assert(nOutDelta > 0 && Gia_ManCoNum(p) % nOutDelta == 0, "nOutDelta > 0 && Gia_ManCoNum(p) % nOutDelta == 0");
    // create the new manager
    pNew = Aig_ManStart( Gia_ManAndNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nConstrs = p.nConstrs;
//    pNew->pSpec = Abc_UtilStrsav( p->pName );
    // create the PIs
    // inlining of macro ABC_CALLOC
    ppNodes = ((Aig_Obj_t * *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Aig_Obj_t *)));
    ppNodes[0] = Aig_ManConst0(pNew);
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        ppNodes[Gia_ObjId(p, pObj)] = Aig_ObjCreateCi( pNew );
    }
    // add logic for the POs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManToAig_rec( pNew, ppNodes, p, Gia_ObjFanin0(pObj) );        
        if i % nOutDelta != 0 {
            continue; }
        ppNodes[Gia_ObjId(p, pObj)] = Aig_ObjCreateCo( pNew, Gia_ObjChild0Copy2(ppNodes, pObj, Gia_ObjId(p, pObj)) );
    }
    Aig_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // inlining of macro ABC_FREE
    ((ppNodes) ? (free((char *) (ppNodes)), (ppNodes) = 0) : 0);
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManToAigSimple(p *Gia_Man_t) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var ppNodes **Aig_Obj_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // inlining of macro ABC_FALLOC
    ppNodes = ((Aig_Obj_t * *) memset(malloc(sizeof(Aig_Obj_t *) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(Aig_Obj_t *) * (size_t)(Gia_ManObjNum(p))));
    // create the new manager
    pNew = Aig_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nConstrs = p.nConstrs;
    // create the PIs
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            ppNodes[i] = Aig_And( pNew, Gia_ObjChild0Copy2(ppNodes, pObj, Gia_ObjId(p, pObj)), Gia_ObjChild1Copy2(ppNodes, pObj, Gia_ObjId(p, pObj)) );
        } else if Gia_ObjIsCi(pObj) {
            ppNodes[i] = Aig_ObjCreateCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            ppNodes[i] = Aig_ObjCreateCo( pNew, Gia_ObjChild0Copy2(ppNodes, pObj, Gia_ObjId(p, pObj)) );
        } else if Gia_ObjIsConst0(pObj) {
            ppNodes[i] = Aig_ManConst0(pNew);
        } else {
            assert(false, "0"); }
        pObj.Value = Abc_Var2Lit( Aig_ObjId(Aig_Regular(ppNodes[i])), Aig_IsComplement(ppNodes[i]) );
        assert(i == 0 || Aig_ObjId(ppNodes[i]) == i, "i == 0 || Aig_ObjId(ppNodes[i]) == i");
    }
    Aig_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // inlining of macro ABC_FREE
    ((ppNodes) ? (free((char *) (ppNodes)), (ppNodes) = 0) : 0);
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCofactorAig(p *Aig_Man_t, nFrames int32, nCofFanLit int32) *Aig_Man_t {
    var pMan *Aig_Man_t;
    var pGia, pTemp *Gia_Man_t;
    pGia = Gia_ManFromAig( p );
    pGia = Gia_ManUnrollAndCofactor( pTemp = pGia, nFrames, nCofFanLit, 1 );
    Gia_ManStop( pTemp );
    pMan = Gia_ManToAig( pGia, 0 );
    Gia_ManStop( pGia );
    return pMan;
}


/**Function*************************************************************

  Synopsis    [Transfers representatives from pGia to pAig.]

  Description [Assumes that pGia was created from pAig.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManReprToAigRepr(pAig *Aig_Man_t, pGia *Gia_Man_t) {
    var pObj *Aig_Obj_t;
    var pGiaObj, pGiaRepr *Gia_Obj_t;
    var i int32;
    assert(pAig.pReprs == nil, "pAig.pReprs == nil");
    assert(pGia.pReprs != nil, "pGia.pReprs != nil");
    // move pointers from AIG to GIA
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(i == 0 || !Abc_LitIsCompl(pObj.iData), "i == 0 || !Abc_LitIsCompl(pObj.iData)");
        pGiaObj = Gia_ManObj( pGia, Abc_Lit2Var(pObj.iData) );
        pGiaObj.Value = i;
    }
    }
    // set the pointers to the nodes in AIG
    Aig_ManReprStart( pAig, Aig_ManObjNumMax(pAig) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia.nObjs) && ((pGiaObj) = Gia_ManObj(pGia, i)); i++  {
        pGiaRepr = Gia_ObjReprObj( pGia, i );
        if pGiaRepr == nil {
            continue; }
        Aig_ObjCreateRepr( pAig, Aig_ManObj(pAig, pGiaRepr.Value), Aig_ManObj(pAig, pGiaObj.Value) );
    }
}
func Gia_ManReprToAigRepr2(pAig *Aig_Man_t, pGia *Gia_Man_t) {
    var pGiaObj, pGiaRepr *Gia_Obj_t;
    var i int32;
    assert(pAig.pReprs == nil, "pAig.pReprs == nil");
    assert(pGia.pReprs != nil, "pGia.pReprs != nil");
    // set the pointers to the nodes in AIG
    Aig_ManReprStart( pAig, Aig_ManObjNumMax(pAig) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia.nObjs) && ((pGiaObj) = Gia_ManObj(pGia, i)); i++  {
        pGiaRepr = Gia_ObjReprObj( pGia, i );
        if pGiaRepr == nil {
            continue; }
        Aig_ObjCreateRepr( pAig, Aig_ManObj(pAig, Abc_Lit2Var(pGiaRepr.Value)), Aig_ManObj(pAig, Abc_Lit2Var(pGiaObj.Value)) );
    }
}

/**Function*************************************************************

  Synopsis    [Transfers representatives from pAig to pGia.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManReprFromAigRepr(pAig *Aig_Man_t, pGia *Gia_Man_t) {
    var pObjGia *Gia_Obj_t; 
    var pObjAig, pReprAig *Aig_Obj_t;
    var i int32;
    assert(pAig.pReprs != nil, "pAig.pReprs != nil");
    assert(pGia.pReprs == nil, "pGia.pReprs == nil");
    assert(Gia_ManObjNum(pGia) - Gia_ManCoNum(pGia) == Aig_ManObjNum(pAig) - Aig_ManCoNum(pAig), "Gia_ManObjNum(pGia) - Gia_ManCoNum(pGia) == Aig_ManObjNum(pAig) - Aig_ManCoNum(pAig)");
    // inlining of macro ABC_CALLOC
    pGia.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(pGia)), sizeof(Gia_Rpr_t)));
    for  i = 0; i < Gia_ManObjNum(pGia); i++  {
        Gia_ObjSetRepr( pGia, i, GIA_VOID );
    }
    // move pointers from GIA to AIG
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia.nObjs) && ((pObjGia) = Gia_ManObj(pGia, i)); i++  {
        if Gia_ObjIsCo(pObjGia) {
            continue; }
        assert(i == 0 || !Abc_LitIsCompl(Gia_ObjValue(pObjGia)), "i == 0 || !Abc_LitIsCompl(Gia_ObjValue(pObjGia))");
        pObjAig  = Aig_ManObj( pAig, Abc_Lit2Var(Gia_ObjValue(pObjGia)) );
        pObjAig.iData = i;
    }
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObjAig) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObjAig) == nil {} else {
        if Aig_ObjIsCo(pObjAig) {
            continue; }
        if pAig.pReprs[i] == nil {
            continue; }
        pReprAig = pAig.pReprs[i];
        Gia_ObjSetRepr( pGia, pObjAig.iData, pReprAig.iData );
    }
    }
    pGia.pNexts = Gia_ManDeriveNexts( pGia );
}
func Gia_ManReprFromAigRepr2(pAig *Aig_Man_t, pGia *Gia_Man_t) {
    var pObjAig, pReprAig *Aig_Obj_t;
    var i int32;
    assert(pAig.pReprs != nil, "pAig.pReprs != nil");
    assert(pGia.pReprs == nil, "pGia.pReprs == nil");
    assert(Gia_ManObjNum(pGia) - Gia_ManCoNum(pGia) == Aig_ManObjNum(pAig) - Aig_ManCoNum(pAig), "Gia_ManObjNum(pGia) - Gia_ManCoNum(pGia) == Aig_ManObjNum(pAig) - Aig_ManCoNum(pAig)");
    // inlining of macro ABC_CALLOC
    pGia.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(pGia)), sizeof(Gia_Rpr_t)));
    for  i = 0; i < Gia_ManObjNum(pGia); i++  {
        Gia_ObjSetRepr( pGia, i, GIA_VOID );
    }
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObjAig) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObjAig) == nil {} else {
        if Aig_ObjIsCo(pObjAig) {
            continue; }
        if pAig.pReprs[i] == nil {
            continue; }
        pReprAig = pAig.pReprs[i];
        Gia_ObjSetRepr( pGia, Abc_Lit2Var(pObjAig.iData), Abc_Lit2Var(pReprAig.iData) );
    }
    }
    pGia.pNexts = Gia_ManDeriveNexts( pGia );
}

/**Function*************************************************************

  Synopsis    [Applies DC2 to the GIA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCompress2(p *Gia_Man_t, fUpdateLevel int32, fVerbose int32) *Gia_Man_t {
    var pGia *Gia_Man_t;
    var pNew, pTemp *Aig_Man_t;
    if p.pManTime && p.vLevels == nil {
        Gia_ManLevelWithBoxes( p ); }
    pNew = Gia_ManToAig( p, 0 );
    pNew = Dar_ManCompress2( pTemp = pNew, 1, fUpdateLevel, 1, 0, fVerbose );
    Aig_ManStop( pTemp );
    pGia = Gia_ManFromAig( pNew );
    Aig_ManStop( pNew );
    Gia_ManTransferTiming( pGia, p );
    return pGia;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformDch(p *Gia_Man_t, pPars any) *Gia_Man_t {
    var fUseMapping int32;
    var pGia, pGia1 *Gia_Man_t;
    var pNew *Aig_Man_t;
    if p.pManTime && p.vLevels == nil {
        Gia_ManLevelWithBoxes( p ); }
    if (fUseMapping != 0) && Gia_ManHasMapping(p) {
        pGia1 = (Gia_Man_t *)Dsm_ManDeriveGia( p, 0 );
    } else {
        pGia1 = Gia_ManDup( p ); }
    pNew = Gia_ManToAig( pGia1, 0 );
    Gia_ManStop( pGia1 );
    pNew = Dar_ManChoiceNew( pNew, (Dch_Pars_t *)pPars );
//    pGia = Gia_ManFromAig( pNew );
    pGia = Gia_ManFromAigChoices( pNew );
    Aig_ManStop( pNew );
    Gia_ManTransferTiming( pGia, p );
    return pGia;
}

/**Function*************************************************************

  Synopsis    [Computes equivalences after structural sequential cleanup.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSeqCleanupClasses(p *Gia_Man_t, fConst int32, fEquiv int32, fVerbose int32) {
    var pNew, pTemp *Aig_Man_t;
    pNew  = Gia_ManToAigSimple( p );
    pTemp = Aig_ManScl( pNew, fConst, fEquiv, 0, -1, -1, fVerbose, 0 );
    Gia_ManReprFromAigRepr( pNew, p );
    Aig_ManStop( pTemp );
    Aig_ManStop( pNew );
}

/**Function*************************************************************

  Synopsis    [Solves SAT problem.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSolveSat(p *Gia_Man_t) int32 {
//    extern int Fra_FraigSat( Aig_Man_t * pMan, ABC_INT64_T nConfLimit, ABC_INT64_T nInsLimit, int fFlipBits, int fAndOuts, int fNewSolver, int fVerbose );
    var pNew *Aig_Man_t;
    var RetValue int32;//, clk = Abc_Clock();
    pNew = Gia_ManToAig( p, 0 );
    RetValue = Fra_FraigSat( pNew, 10000000, 0, 0, 0, 0, 1, 1, 0, 0 );
    if RetValue == 0 {
        var pObj *Gia_Obj_t;
        var i int32;
        var pInit *int32 = (int *)pNew.pData;
        Gia_ManConst0(p).fMark0 = 0;
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.fMark0 = pInit[i];
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                           (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj));
        }
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if pObj.fMark0 != 1 {
                break; }
        }
        if i != Gia_ManPoNum(p) {
            Abc_Print( 1, "Counter-example verification has failed.  " ); }
//        else
//            Abc_Print( 1, "Counter-example verification succeeded.  " );
    }
/*
    else if ( RetValue == 1 )
        Abc_Print( 1, "The SAT problem is unsatisfiable.  " );
    else if ( RetValue == -1 )
        Abc_Print( 1, "The SAT problem is undecided.  " );
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
*/
    Aig_ManStop( pNew );
    return RetValue;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


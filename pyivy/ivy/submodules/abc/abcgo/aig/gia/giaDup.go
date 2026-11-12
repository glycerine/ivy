package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaDup.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Duplication procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaDup.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/tim/tim.h"
//#include "misc/vec/vecWec.h"
//#include "proof/cec/cec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Removes pointers to the unmarked nodes..]

  Description [Array vLits contains literals of p. At the same time, 
  each object pObj of p points to a literal of pNew. This procedure
  remaps literals in array vLits into literals of pNew.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupRemapLiterals(vLits *Vec_Int_t, p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i, iLit, iLitNew int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLits)) && (((iLit) = Vec_IntEntry(vLits, i)), 1); i++  {
        if iLit < 0 {
            continue; }
        pObj = Gia_ManObj( p, Abc_Lit2Var(iLit) );
        if ~pObj.Value == 0 {
            iLitNew = -1;
        } else {
            iLitNew = Abc_LitNotCond( pObj.Value, Abc_LitIsCompl(iLit) ); }
        Vec_IntWriteEntry( vLits, i, iLitNew );
    }
}

/**Function*************************************************************

  Synopsis    [Removes pointers to the unmarked nodes..]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupRemapEquiv(pNew *Gia_Man_t, p *Gia_Man_t) {
    var vClass *Vec_Int_t;
    var i, k, iNode, iRepr, iPrev int32;
    if p.pReprs == nil {
        return; }
    assert(pNew.pReprs == nil && pNew.pNexts == nil, "pNew.pReprs == nil && pNew.pNexts == nil");
    // start representatives
    // inlining of macro ABC_CALLOC
    pNew.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(pNew)), sizeof(Gia_Rpr_t)));
    for  i = 0; i < Gia_ManObjNum(pNew); i++  {
        Gia_ObjSetRepr( pNew, i, GIA_VOID );
    }
    // iterate over constant candidates
    // inlining of macro Gia_ManForEachConst
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsConst(p, i) {} else {
        Gia_ObjSetRepr( pNew, Abc_Lit2Var(Gia_ManObj(p, i).Value), 0 ); }
    }
    // iterate over class candidates
    vClass = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        Vec_IntClear( vClass );
        // inlining of macro Gia_ClassForEachObj
        for  assert(Gia_ObjIsHead(p, i) && (i != 0), "Gia_ObjIsHead(p, i) && i"), k = i; k > 0; k = Gia_ObjNext(p, k)  {
            Vec_IntPushUnique( vClass, Abc_Lit2Var(Gia_ManObj(p, k).Value) );
        }
        assert(Vec_IntSize( vClass ) > 1, "Vec_IntSize( vClass ) > 1");
        Vec_IntSort( vClass, 0 );
        iRepr = iPrev = Vec_IntEntry( vClass, 0 );
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vClass)) && (((iNode) = Vec_IntEntry(vClass, k)), 1); k++  {
            Gia_ObjSetRepr( pNew, iNode, iRepr );
            assert(iPrev < iNode, "iPrev < iNode");
            iPrev = iNode;
        }
    }
    }
    Vec_IntFree( vClass );
    pNew.pNexts = Gia_ManDeriveNexts( pNew );
}

/**Function*************************************************************

  Synopsis    [Remaps combinational inputs when objects are DFS ordered.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupRemapCis(pNew *Gia_Man_t, p *Gia_Man_t) {
    var pObj, pObjNew *Gia_Obj_t;
    var i int32;
    assert(Vec_IntSize(p.vCis) == Vec_IntSize(pNew.vCis), "Vec_IntSize(p.vCis) == Vec_IntSize(pNew.vCis)");
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        assert(Gia_ObjCioId(pObj) == i, "Gia_ObjCioId(pObj) == i");
        pObjNew = Gia_ObjFromLit( pNew, pObj.Value );
        assert(!Gia_IsComplement(pObjNew), "!Gia_IsComplement(pObjNew)");
        Vec_IntWriteEntry( pNew.vCis, i, Gia_ObjId(pNew, pObjNew) );
        Gia_ObjSetCioId( pObjNew, i );
    }
}

/**Function*************************************************************

  Synopsis    [Remaps combinational outputs when objects are DFS ordered.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupRemapCos(pNew *Gia_Man_t, p *Gia_Man_t) {
    var pObj, pObjNew *Gia_Obj_t;
    var i int32;
    assert(Vec_IntSize(p.vCos) == Vec_IntSize(pNew.vCos), "Vec_IntSize(p.vCos) == Vec_IntSize(pNew.vCos)");
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        assert(Gia_ObjCioId(pObj) == i, "Gia_ObjCioId(pObj) == i");
        pObjNew = Gia_ObjFromLit( pNew, pObj.Value );
        assert(!Gia_IsComplement(pObjNew), "!Gia_IsComplement(pObjNew)");
        Vec_IntWriteEntry( pNew.vCos, i, Gia_ObjId(pNew, pObjNew) );
        Gia_ObjSetCioId( pObjNew, i );
    }
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupOrderDfs_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if ~pObj.Value {
        return pObj.Value; }
    if Gia_ObjIsCi(pObj) {
        return pObj.Value = Gia_ManAppendCi(pNew); }
//    if ( p->pNexts && Gia_ObjNext(p, Gia_ObjId(p, pObj)) )
//        Gia_ManDupOrderDfs_rec( pNew, p, Gia_ObjNextObj(p, Gia_ObjId(p, pObj)) );
    Gia_ManDupOrderDfs_rec( pNew, p, Gia_ObjFanin0(pObj) );
    if Gia_ObjIsCo(pObj) {
        return pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    Gia_ManDupOrderDfs_rec( pNew, p, Gia_ObjFanin1(pObj) );
    return pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
} 

/**Function*************************************************************

  Synopsis    [Duplicates AIG while putting objects in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupOrderDfs(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupOrderDfs_rec( pNew, p, pObj );
    }
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if !~pObj.Value {
            pObj.Value = Gia_ManAppendCi(pNew); }
    }
    assert(Gia_ManCiNum(pNew) == Gia_ManCiNum(p), "Gia_ManCiNum(pNew) == Gia_ManCiNum(p)");
    Gia_ManDupRemapCis( pNew, p );
    Gia_ManDupRemapEquiv( pNew, p );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while putting objects in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupAbs(p *Gia_Man_t, vMapPpi2Ff *Vec_Int_t, vMapFf2Ppi *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var k, Flop, Used int32;
    assert(Vec_IntSize(vMapFf2Ppi) == Vec_IntSize(vMapPpi2Ff) + Vec_IntCountEntry(vMapFf2Ppi, -1), "Vec_IntSize(vMapFf2Ppi) == Vec_IntSize(vMapPpi2Ff) + Vec_IntCountEntry(vMapFf2Ppi, -1)");
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // create inputs
    // inlining of macro Gia_ManForEachPi
    for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vMapPpi2Ff)) && (((Flop) = Vec_IntEntry(vMapPpi2Ff, k)), 1); k++  {
        pObj = Gia_ManCi( p, Gia_ManPiNum(p) + Flop );
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Vec_IntForEachEntry
    for  Flop = 0; (Flop < Vec_IntSize(vMapFf2Ppi)) && (((Used) = Vec_IntEntry(vMapFf2Ppi, Flop)), 1); Flop++  {
        pObj = Gia_ManCi( p, Gia_ManPiNum(p) + Flop );
        if Used >= 0 {
            assert(pObj.Value != ~0, "pObj.Value != ~0");
            continue;
        }
        assert(pObj.Value == ~0, "pObj.Value == ~0");
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachCi
    for  k = 0; (k < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, k)); k++  {
        assert(pObj.Value != ~0, "pObj.Value != ~0");
    }
    // create nodes
    // inlining of macro Gia_ManForEachPo
    for  k = 0; (k < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, k)); k++  {
        Gia_ManDupOrderDfs_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    // inlining of macro Vec_IntForEachEntry
    for  Flop = 0; (Flop < Vec_IntSize(vMapFf2Ppi)) && (((Used) = Vec_IntEntry(vMapFf2Ppi, Flop)), 1); Flop++  {
        if Used >= 0 {
            continue; }
        pObj = Gia_ManCi( p, Gia_ManPiNum(p) + Flop );
        pObj = Gia_ObjRoToRi( p, pObj );
        Gia_ManDupOrderDfs_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    // create outputs
    // inlining of macro Gia_ManForEachPo
    for  k = 0; (k < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, k)); k++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Vec_IntForEachEntry
    for  Flop = 0; (Flop < Vec_IntSize(vMapFf2Ppi)) && (((Used) = Vec_IntEntry(vMapFf2Ppi, Flop)), 1); Flop++  {
        if Used >= 0 {
            continue; }
        pObj = Gia_ManCi( p, Gia_ManPiNum(p) + Flop );
        pObj = Gia_ObjRoToRi( p, pObj );
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) - Vec_IntSize(vMapPpi2Ff) );
    assert(Gia_ManPiNum(pNew) == Gia_ManPiNum(p) + Vec_IntSize(vMapPpi2Ff), "Gia_ManPiNum(pNew) == Gia_ManPiNum(p) + Vec_IntSize(vMapPpi2Ff)");
    assert(Gia_ManCiNum(pNew) == Gia_ManCiNum(p), "Gia_ManCiNum(pNew) == Gia_ManCiNum(p)");
    assert(Gia_ManPoNum(pNew) == Gia_ManPoNum(p), "Gia_ManPoNum(pNew) == Gia_ManPoNum(p)");
    assert(Gia_ManCoNum(pNew) == Gia_ManCoNum(p) - Vec_IntSize(vMapPpi2Ff), "Gia_ManCoNum(pNew) == Gia_ManCoNum(p) - Vec_IntSize(vMapPpi2Ff)");
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Duplicates AIG while putting objects in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupOutputGroup(p *Gia_Man_t, iOutStart int32, iOutStop int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    for  i = iOutStart; i < iOutStop; i++  {
        pObj = Gia_ManCo( p, i );
        Gia_ManDupOrderDfs_rec( pNew, p, pObj );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while putting objects in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupOutputVec(p *Gia_Man_t, vOutPres *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(Gia_ManPoNum(p) == Vec_IntSize(vOutPres), "Gia_ManPoNum(p) == Vec_IntSize(vOutPres)");
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if Vec_IntEntry(vOutPres, i) {
            Gia_ManDupOrderDfs_rec( pNew, p, Gia_ObjFanin0(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if Vec_IntEntry(vOutPres, i) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while putting objects in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupSelectedOutputs(p *Gia_Man_t, vOutsLeft *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, iOut int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(Gia_ManPoNum(p) >= Vec_IntSize(vOutsLeft), "Gia_ManPoNum(p) >= Vec_IntSize(vOutsLeft)");
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOutsLeft)) && (((iOut) = Vec_IntEntry(vOutsLeft, i)), 1); i++  {
        Gia_ManDupOrderDfs_rec( pNew, p, Gia_ObjFanin0(Gia_ManPo(p, iOut)) );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOutsLeft)) && (((iOut) = Vec_IntEntry(vOutsLeft, i)), 1); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(Gia_ManPo(p, iOut)) );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupOrderDfsChoices_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    var pNext *Gia_Obj_t;
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    pNext = Gia_ObjNextObj( p, Gia_ObjId(p, pObj) );
    if pNext != nil {
        Gia_ManDupOrderDfsChoices_rec( pNew, p, pNext ); }
    Gia_ManDupOrderDfsChoices_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Gia_ManDupOrderDfsChoices_rec( pNew, p, Gia_ObjFanin1(pObj) );
    pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    if pNext != nil {
        pNew.pNexts[Abc_Lit2Var(pObj.Value)] = Abc_Lit2Var( Abc_Lit2Var(pNext.Value) );
        assert(Abc_Lit2Var(pObj.Value) > Abc_Lit2Var(pNext.Value), "Abc_Lit2Var(pObj.Value) > Abc_Lit2Var(pNext.Value)");
    }
} 

/**Function*************************************************************

  Synopsis    [Duplicates AIG while putting objects in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupOrderDfsChoices(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(p.pReprs && p.pNexts, "p.pReprs && p.pNexts");
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // inlining of macro ABC_CALLOC
    pNew.pNexts = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupOrderDfsChoices_rec( pNew, p, Gia_ObjFanin0(pObj) );
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
//    Gia_ManDeriveReprs( pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while putting objects in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupOrderDfsReverse(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCoReverse
    for  i = Vec_IntSize(p.vCos) - 1; (i >= 0) && ((pObj) = Gia_ManCo(p, i)); i--  {
        Gia_ManDupOrderDfs_rec( pNew, p, pObj );
    }
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if !~pObj.Value {
            pObj.Value = Gia_ManAppendCi(pNew); }
    }
    assert(Gia_ManCiNum(pNew) == Gia_ManCiNum(p), "Gia_ManCiNum(pNew) == Gia_ManCiNum(p)");
    Gia_ManDupRemapCis( pNew, p );
    Gia_ManDupRemapCos( pNew, p );
    Gia_ManDupRemapEquiv( pNew, p );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while putting first PIs, then nodes, then POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupOrderAiger(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManDupRemapEquiv( pNew, p );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    assert(Gia_ManIsNormalized(pNew), "Gia_ManIsNormalized(pNew)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while putting first PIs, then nodes, then POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupOnsetOffset(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        pObj.Value = Gia_ManAppendCo( pNew, Abc_LitNot(Gia_ObjFanin0Copy(pObj)) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while putting first PIs, then nodes, then POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupLastPis(p *Gia_Man_t, nLastPis int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if i < Gia_ManCiNum(p) - nLastPis {
            pObj.Value = ~0;
        } else {
            pObj.Value = Gia_ManAppendCi(pNew);
        }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while complementing the flops.]

  Description [The array of initial state contains the init state
  for each state bit of the flops in the design.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupFlip(p *Gia_Man_t, pInitState *int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
            if Gia_ObjCioId(pObj) >= Gia_ManPiNum(p) {
                pObj.Value = Abc_LitNotCond( pObj.Value, Abc_InfoHasBit((unsigned *)pInitState, Gia_ObjCioId(pObj) - Gia_ManPiNum(p)) ); }
        
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ObjFanin0Copy(pObj);
            if Gia_ObjCioId(pObj) >= Gia_ManPoNum(p) {
                pObj.Value = Abc_LitNotCond( pObj.Value, Abc_InfoHasBit((unsigned *)pInitState, Gia_ObjCioId(pObj) - Gia_ManPoNum(p)) ); }
            pObj.Value = Gia_ManAppendCo( pNew, pObj.Value );
        }
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Cycles AIG using random input.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCycle(p *Gia_Man_t, pCex *Abc_Cex_t, nFrames int32) {
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var i, k int32;
    Gia_ManRandom( 1 );
    assert(pCex == nil || nFrames <= pCex.IFrame, "pCex == nil || nFrames <= pCex.IFrame");
    // iterate for the given number of frames
    for  i = 0; i < nFrames; i++  {
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
            if pCex != nil {
                pObj.fMark0 = Abc_InfoHasBit(pCex.PData, pCex.NRegs+i*pCex.NPis+k);
            } else {
                pObj.fMark0 = (1 & Gia_ManRandom(0));
            }
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                           (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            pObjRo.fMark0 = pObjRi.fMark0;
        }
    }
}
func Gia_ManDupCycled(p *Gia_Man_t, pCex *Abc_Cex_t, nFrames int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var vInits *Vec_Bit_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManCleanMark0(p);
    Gia_ManCycle( p, pCex, nFrames );
    vInits = Vec_BitAlloc( Gia_ManRegNum(p) );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Vec_BitPush( vInits, pObj.fMark0 );
    }
    pNew = Gia_ManDupFlip( p, Vec_BitArray(vInits) );
    Vec_BitFree( vInits );
    Gia_ManCleanMark0(p);
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Duplicates AIG without any changes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDup(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    if Gia_ManHasChoices(p) {
        // inlining of macro ABC_CALLOC
        pNew.pSibls = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int))); }
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsBuf(pObj) {
            pObj.Value = Gia_ManAppendBuf( pNew, Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
            if Gia_ObjSibl(p, Gia_ObjId(p, pObj)) {
                pNew.pSibls[Abc_Lit2Var(pObj.Value)] = Abc_Lit2Var(Gia_ObjSiblObj(p, Gia_ObjId(p, pObj)).Value); }  
        
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    if p.pCexSeq {
        pNew.pCexSeq = Abc_CexDup( p.pCexSeq, Gia_ManRegNum(p) ); }
    return pNew;
}
func Gia_ManDup2(p1 *Gia_Man_t, p2 *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Gia_ManCiNum(p1) == Gia_ManCiNum(p2), "Gia_ManCiNum(p1) == Gia_ManCiNum(p2)");
    assert(Gia_ManCoNum(p1) == Gia_ManCoNum(p2), "Gia_ManCoNum(p1) == Gia_ManCoNum(p2)");
    pNew = Gia_ManStart( Gia_ManObjNum(p1) + Gia_ManObjNum(p2) );
    Gia_ManHashStart( pNew );
    Gia_ManConst0(p1).Value = 0;
    Gia_ManConst0(p2).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p1.vCis)) && ((pObj) = Gia_ManCi(p1, i)); i++  {
        pObj.Value = Gia_ManCi(p2, i).Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p1.nObjs) && ((pObj) = Gia_ManObj(p1, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p2.nObjs) && ((pObj) = Gia_ManObj(p2, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p1.vCos)) && ((pObj) = Gia_ManCo(p1, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p2.vCos)) && ((pObj) = Gia_ManCo(p2, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p1) );
    Gia_ManHashStop( pNew );
    return pNew;
}
func Gia_ManDupWithAttributes(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t  = Gia_ManDup(p);
    Gia_ManTransferMapping( pNew, p );
    Gia_ManTransferPacking( pNew, p );
    if p.pManTime {
        pNew.pManTime = Tim_ManDup( (Tim_Man_t *)p.pManTime, 0 ); }
    if p.pAigExtra {
        pNew.pAigExtra = Gia_ManDup( p.pAigExtra ); }
    if p.nAnd2Delay {
        pNew.nAnd2Delay = p.nAnd2Delay; }
    if p.vRegClasses {
        pNew.vRegClasses = Vec_IntDup( p.vRegClasses ); }
    if p.vRegInits {
        pNew.vRegInits = Vec_IntDup( p.vRegInits ); }
    if p.vConfigs {
        pNew.vConfigs = Vec_IntDup( p.vConfigs ); }
    if p.pCellStr {
        pNew.pCellStr = Abc_UtilStrsav( p.pCellStr ); }
    return pNew;
}
func Gia_ManDupRemovePis(p *Gia_Man_t, nRemPis int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) && Gia_ObjCioId(pObj) < Gia_ManCiNum(p)-nRemPis {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG without any changes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupZero(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t; var i int32;
    pNew = Gia_ManStart( 1 + Gia_ManCiNum(p) + Gia_ManCoNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    for  i = 0; i < Gia_ManCiNum(p); i++  {
        Gia_ManAppendCi( pNew );
    }
    for  i = 0; i < Gia_ManCoNum(p); i++  {
        Gia_ManAppendCo( pNew, 0 );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG without any changes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupPerm(p *Gia_Man_t, vPiPerm *Vec_Int_t) *Gia_Man_t {
//    Vec_Int_t * vPiPermInv;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Vec_IntSize(vPiPerm) == Gia_ManPiNum(p), "Vec_IntSize(vPiPerm) == Gia_ManPiNum(p)");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
//    vPiPermInv = Vec_IntInvert( vPiPerm, -1 );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
//        Gia_ManPi(p, Vec_IntEntry(vPiPermInv,i))->Value = Gia_ManAppendCi( pNew );
        Gia_ManPi(p, Vec_IntEntry(vPiPerm,i)).Value = Gia_ManAppendCi( pNew );
    }
//    Vec_IntFree( vPiPermInv );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) {
            if Gia_ObjIsRo(p, pObj) {
                pObj.Value = Gia_ManAppendCi( pNew ); }
        
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}
func Gia_ManDupPermFlop(p *Gia_Man_t, vFfPerm *Vec_Int_t) *Gia_Man_t {
    //Vec_Int_t * vPermInv;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Vec_IntSize(vFfPerm) == Gia_ManRegNum(p), "Vec_IntSize(vFfPerm) == Gia_ManRegNum(p)");
    //vPermInv = Vec_IntInvert( vFfPerm, -1 );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        //Gia_ManRo(p, Vec_IntEntry(vPermInv, i))->Value = Gia_ManAppendCi(pNew);
        Gia_ManRo(p, Vec_IntEntry(vFfPerm, i)).Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        //pObj->Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy( Gia_ManRi(p, Vec_IntEntry(vPermInv, i)) ) );
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy( Gia_ManRi(p, Vec_IntEntry(vFfPerm, i)) ) );
    }
    //Vec_IntFree( vPermInv );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}
func Gia_ManDupSpreadFlop(p *Gia_Man_t, vFfMask *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, k, Entry int32;
    assert(Vec_IntSize(vFfMask) >= Gia_ManRegNum(p), "Vec_IntSize(vFfMask) >= Gia_ManRegNum(p)");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    k = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFfMask)) && (((Entry) = Vec_IntEntry(vFfMask, i)), 1); i++  {
        if Entry == -1 {
            Gia_ManAppendCi(pNew);
        } else {
            Gia_ManRo(p, k++).Value = Gia_ManAppendCi(pNew); }
    }
    assert(k == Gia_ManRegNum(p), "k == Gia_ManRegNum(p)");
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    k = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFfMask)) && (((Entry) = Vec_IntEntry(vFfMask, i)), 1); i++  {
        if Entry == -1 {
            Gia_ManAppendCo( pNew, 0 );
        } else {
            pObj = Gia_ManRi( p, k++ );
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        }
    }
    assert(k == Gia_ManRegNum(p), "k == Gia_ManRegNum(p)");
    Gia_ManSetRegNum( pNew, Vec_IntSize(vFfMask) );
    return pNew;
}
func Gia_ManDupPermFlopGap(p *Gia_Man_t, vFfMask *Vec_Int_t) *Gia_Man_t {
    var vPerm *Vec_Int_t  = Vec_IntCondense( vFfMask, -1 );
    var pPerm *Gia_Man_t  = Gia_ManDupPermFlop( p, vPerm );
    var pSpread *Gia_Man_t  = Gia_ManDupSpreadFlop( pPerm, vFfMask );
    Vec_IntFree( vPerm );
    Gia_ManStop( pPerm );
    return pSpread;
}

/**Function*************************************************************

  Synopsis    [Appends second AIG without any changes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupPiPerm(p *Gia_Man_t) *Gia_Man_t {
    var pNew, pOne *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManRandom(1);
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var iLit0 int32  = Gia_ObjFanin0Copy(pObj);
        var iLit1 int32  = Gia_ObjFanin1Copy(pObj);
        var iPlace0 int32  = Gia_ManRandom(0) % Gia_ManCiNum(p);
        var iPlace1 int32  = Gia_ManRandom(0) % Gia_ManCiNum(p);
        if Abc_Lit2Var(iLit0) <= Gia_ManCiNum(p) {
            iLit0 = Abc_Var2Lit( iPlace0+1, Abc_LitIsCompl(iLit0) ); }
        if Abc_Lit2Var(iLit1) <= Gia_ManCiNum(p) {
            iLit1 = Abc_Var2Lit( iPlace1+1, Abc_LitIsCompl(iLit1) ); }
        pObj.Value = Gia_ManHashAnd( pNew, iLit0, iLit1 );
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pOne = pNew );
    Gia_ManStop( pOne );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Appends second AIG without any changes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupAppend(pNew *Gia_Man_t, pTwo *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    if pNew.nRegs > 0 {
        pNew.nRegs = 0; }
    if Vec_IntSize(&pNew.vHTable) == 0 {
        Gia_ManHashStart( pNew ); }
    Gia_ManConst0(pTwo).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < pTwo.nObjs) && ((pObj) = Gia_ManObj(pTwo, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
}
func Gia_ManDupAppendShare(pNew *Gia_Man_t, pTwo *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Gia_ManCiNum(pNew) == Gia_ManCiNum(pTwo), "Gia_ManCiNum(pNew) == Gia_ManCiNum(pTwo)");
    if Vec_IntSize(&pNew.vHTable) == 0 {
        Gia_ManHashStart( pNew ); }
    Gia_ManConst0(pTwo).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < pTwo.nObjs) && ((pObj) = Gia_ManObj(pTwo, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_Obj2Lit( pNew, Gia_ManCi( pNew, Gia_ObjCioId(pObj) ) );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
}
func Gia_ManDupAppendNew(pOne *Gia_Man_t, pTwo *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(pOne) + Gia_ManObjNum(pTwo) );
    pNew.pName = Abc_UtilStrsav( pOne.pName );
    pNew.pSpec = Abc_UtilStrsav( pOne.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(pOne).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < pOne.nObjs) && ((pObj) = Gia_ManObj(pOne, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew ); }
    }
    Gia_ManConst0(pTwo).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < pTwo.nObjs) && ((pObj) = Gia_ManObj(pTwo, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsPi(pTwo, pObj) {
            pObj.Value = Gia_ManPi(pOne, Gia_ObjCioId(pObj)).Value;
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew ); }
    }
    Gia_ManHashStop( pNew );
    // primary outputs
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(pOne)) && ((pObj) = Gia_ManCo(pOne, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(pTwo)) && ((pObj) = Gia_ManCo(pTwo, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // flop inputs
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(pOne)) && ((pObj) = Gia_ManCo(pOne, Gia_ManPoNum(pOne)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(pTwo)) && ((pObj) = Gia_ManCo(pTwo, Gia_ManPoNum(pTwo)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(pOne) + Gia_ManRegNum(pTwo) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Creates a miter for inductive checking of the invariant.]

  Description [The first GIA (p) is a sequential AIG whose transition
  relation is used. The second GIA (pInv) is a combinational AIG representing
  the invariant over the register outputs. If the resulting combination miter
  is UNSAT, the invariant holds by simple induction.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupInvMiter(p *Gia_Man_t, pInv *Gia_Man_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, Node1, Node2, Node int32;
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
    assert(Gia_ManRegNum(pInv) == 0, "Gia_ManRegNum(pInv) == 0");
    assert(Gia_ManCoNum(pInv) == 1, "Gia_ManCoNum(pInv) == 1");
    assert(Gia_ManRegNum(p) == Gia_ManCiNum(pInv), "Gia_ManRegNum(p) == Gia_ManCiNum(pInv)");
    Gia_ManFillValue(p);
    pNew = Gia_ManStart( Gia_ManObjNum(p) + 2*Gia_ManObjNum(pInv) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ObjFanin0Copy(pObj); }
    }
    // build invariant on top of register outputs in the first frame
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ManCi(pInv, i).Value = pObj.Value;
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pInv.nObjs) && ((pObj) = Gia_ManObj(pInv, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    pObj = Gia_ManCo( pInv, 0 );
    Node1 = Gia_ObjFanin0Copy(pObj);
    // build invariant on top of register outputs in the second frame
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        Gia_ManCi(pInv, i).Value = pObj.Value;
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pInv.nObjs) && ((pObj) = Gia_ManObj(pInv, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    pObj = Gia_ManCo( pInv, 0 );
    Node2 = Gia_ObjFanin0Copy(pObj);
    // create miter output
    Node = Gia_ManHashAnd( pNew, Node1, Abc_LitNot(Node2) );
    Gia_ManAppendCo( pNew, Node );
    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Appends logic cones as additional property outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupAppendCones(p *Gia_Man_t, ppCones **Gia_Man_t, nCones int32, fOnlyRegs int32) *Gia_Man_t {
    var pNew, pOne *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, k int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    for  k = 0; k < nCones; k++  {
        pOne = ppCones[k];
        assert(Gia_ManPoNum(pOne) == 1, "Gia_ManPoNum(pOne) == 1");
        assert(Gia_ManRegNum(pOne) == 0, "Gia_ManRegNum(pOne) == 0");
        if fOnlyRegs != 0 {
            assert(Gia_ManPiNum(pOne) == Gia_ManRegNum(p), "Gia_ManPiNum(pOne) == Gia_ManRegNum(p)");
        } else {
            assert(Gia_ManPiNum(pOne) == Gia_ManCiNum(p), "Gia_ManPiNum(pOne) == Gia_ManCiNum(p)"); }
        Gia_ManConst0(pOne).Value = 0;
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(pOne)) && ((pObj) = Gia_ManCi(pOne, i)); i++  {
            pObj.Value = Gia_ManCiLit( pNew, fOnlyRegs ? Gia_ManPiNum(p) + i : i );
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < pOne.nObjs) && ((pObj) = Gia_ManObj(pOne, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(pOne)) && ((pObj) = Gia_ManCo(pOne, i)); i++  {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        }
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pOne = pNew );
    Gia_ManStop( pOne );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Duplicates while adding self-loops to the registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupSelf(p *Gia_Man_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t; 
    var pObj *Gia_Obj_t;
    var i, iCtrl int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    iCtrl = Gia_ManAppendCi( pNew );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ObjFanin0Copy(pObj);
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.Value = Gia_ManHashMux( pNew, iCtrl, Gia_ObjFanin0Copy(pObj), Gia_ObjRiToRo(p, pObj).Value );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, pObj.Value );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates while adding self-loops to the registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupFlopClass(p *Gia_Man_t, iClass int32) *Gia_Man_t {
    var pNew *Gia_Man_t; 
    var pObj *Gia_Obj_t;
    var i, Counter1, Counter2 int32;
    assert(p.vFlopClasses != nil, "p.vFlopClasses != nil");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if Vec_IntEntry(p.vFlopClasses, i) != iClass {
            pObj.Value = Gia_ManAppendCi( pNew ); }
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if Vec_IntEntry(p.vFlopClasses, i) == iClass {
            pObj.Value = Gia_ManAppendCi( pNew ), Counter1++; }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        if Vec_IntEntry(p.vFlopClasses, i) != iClass {
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        if Vec_IntEntry(p.vFlopClasses, i) == iClass {
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
Counter2++; }
    }
    assert(Counter1 == Counter2, "Counter1 == Counter2");
    Gia_ManSetRegNum( pNew, Counter1 );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG without any changes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupMarked(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, nRos, nRis int32;
    var CountMarked int32;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        CountMarked += pObj.fMark0;
    }
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) - CountMarked );
    if p.pMuxes {
        // inlining of macro ABC_CALLOC
        pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned))); }
    pNew.nConstrs = p.nConstrs;
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if pObj.fMark0 {
            assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
            pObj.fMark0 = 0;
            continue;
        }
        if Gia_ObjIsBuf(pObj) {
            pObj.Value = Gia_ManAppendBuf( pNew, Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsAnd(pObj) {
            if Gia_ObjIsXor(pObj) {
                pObj.Value = Gia_ManAppendXorReal( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
            } else if Gia_ObjIsMux(p, pObj) {
                pObj.Value = Gia_ManAppendMuxReal( pNew, Gia_ObjFanin2Copy(p, pObj), Gia_ObjFanin1Copy(pObj), Gia_ObjFanin0Copy(pObj) );
            } else {
                pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
            nRos += Gia_ObjIsRo(p, pObj);
        
        } else if Gia_ObjIsCo(pObj) {
//            Gia_Obj_t * pFanin = Gia_ObjFanin0(pObj);
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
            nRis += Gia_ObjIsRi(p, pObj);
        }
    }
    assert(pNew.nObjsAlloc == pNew.nObjs, "pNew.nObjsAlloc == pNew.nObjs");
    assert(nRos == nRis, "nRos == nRis");
    Gia_ManSetRegNum( pNew, nRos );
    if p.pReprs && p.pNexts {
        var pRepr *Gia_Obj_t;
        // inlining of macro ABC_CALLOC
        pNew.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(pNew)), sizeof(Gia_Rpr_t)));
        for  i = 0; i < Gia_ManObjNum(pNew); i++  {
            Gia_ObjSetRepr( pNew, i, GIA_VOID );
        }
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if !~pObj.Value {
                continue; }
            pRepr = Gia_ObjReprObj( p, i );
            if pRepr == nil {
                continue; }
            if !~pRepr.Value {
                continue; }
            assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
            if Abc_Lit2Var(pObj.Value) != Abc_Lit2Var(pRepr.Value) {
                Gia_ObjSetRepr( pNew, Abc_Lit2Var(pObj.Value), Abc_Lit2Var(pRepr.Value) ); } 
        }
        pNew.pNexts = Gia_ManDeriveNexts( pNew );
    }
    if Gia_ManHasChoices(p) {
        var pSibl *Gia_Obj_t;
        // inlining of macro ABC_CALLOC
        pNew.pSibls = ((int *) calloc((size_t)(Gia_ManObjNum(pNew)), sizeof(int)));
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if !~pObj.Value {
                continue; }
            pSibl = Gia_ObjSiblObj( p, i );
            if pSibl == nil {
                continue; }
            if !~pSibl.Value {
                continue; }
            assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
            assert(Abc_Lit2Var(pObj.Value) > Abc_Lit2Var(pSibl.Value), "Abc_Lit2Var(pObj.Value) > Abc_Lit2Var(pSibl.Value)");
            pNew.pSibls[Abc_Lit2Var(pObj.Value)] = Abc_Lit2Var(pSibl.Value);
        }
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while creating "parallel" copies.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupTimes(p *Gia_Man_t, nTimes int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vPis, vPos, vRis, vRos *Vec_Int_t;
    var i, t, Entry int32;
    assert(nTimes > 0, "nTimes > 0");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    vPis = Vec_IntAlloc( Gia_ManPiNum(p) * nTimes );
    vPos = Vec_IntAlloc( Gia_ManPoNum(p) * nTimes );
    vRis = Vec_IntAlloc( Gia_ManRegNum(p) * nTimes );
    vRos = Vec_IntAlloc( Gia_ManRegNum(p) * nTimes );
    for  t = 0; t < nTimes; t++  {
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if Gia_ObjIsAnd(pObj) {
                pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
            } else if Gia_ObjIsCi(pObj) {
                pObj.Value = Gia_ManAppendCi( pNew );
                if Gia_ObjIsPi(p, pObj) {
                    Vec_IntPush( vPis, Abc_Lit2Var(pObj.Value) );
                } else {
                    Vec_IntPush( vRos, Abc_Lit2Var(pObj.Value) ); }
            
            } else if Gia_ObjIsCo(pObj) {
                pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
                if Gia_ObjIsPo(p, pObj) {
                    Vec_IntPush( vPos, Abc_Lit2Var(pObj.Value) );
                } else {
                    Vec_IntPush( vRis, Abc_Lit2Var(pObj.Value) ); }
            }
        }
    }
    Vec_IntClear( pNew.vCis );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPis)) && (((Entry) = Vec_IntEntry(vPis, i)), 1); i++  {
        Gia_ObjSetCioId( Gia_ManObj(pNew, Entry), Vec_IntSize(pNew.vCis) );
        Vec_IntPush( pNew.vCis, Entry );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRos)) && (((Entry) = Vec_IntEntry(vRos, i)), 1); i++  {
        Gia_ObjSetCioId( Gia_ManObj(pNew, Entry), Vec_IntSize(pNew.vCis) );
        Vec_IntPush( pNew.vCis, Entry );
    }
    Vec_IntClear( pNew.vCos );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPos)) && (((Entry) = Vec_IntEntry(vPos, i)), 1); i++  {
        Gia_ObjSetCioId( Gia_ManObj(pNew, Entry), Vec_IntSize(pNew.vCos) );
        Vec_IntPush( pNew.vCos, Entry );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRis)) && (((Entry) = Vec_IntEntry(vRis, i)), 1); i++  {
        Gia_ObjSetCioId( Gia_ManObj(pNew, Entry), Vec_IntSize(pNew.vCos) );
        Vec_IntPush( pNew.vCos, Entry );
    }
    Vec_IntFree( vPis );
    Vec_IntFree( vPos );
    Vec_IntFree( vRis );
    Vec_IntFree( vRos );
    Gia_ManSetRegNum( pNew, nTimes * Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupDfs2_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if ~pObj.Value {
        return pObj.Value; }
    if p.pReprsOld && ~p.pReprsOld[Gia_ObjId(p, pObj)] {
        var pRepr *Gia_Obj_t  = Gia_ManObj( p, p.pReprsOld[Gia_ObjId(p, pObj)] );
        pRepr.Value = Gia_ManDupDfs2_rec( pNew, p, pRepr );
        return pObj.Value = Abc_LitNotCond( pRepr.Value, Gia_ObjPhaseReal(pRepr) ^ Gia_ObjPhaseReal(pObj) );
    }
    if Gia_ObjIsCi(pObj) {
        return pObj.Value = Gia_ManAppendCi(pNew); }
    Gia_ManDupDfs2_rec( pNew, p, Gia_ObjFanin0(pObj) );
    if Gia_ObjIsCo(pObj) {
        return pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    Gia_ManDupDfs2_rec( pNew, p, Gia_ObjFanin1(pObj) );
    if Vec_IntSize(&pNew.vHTable) {
        return pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    return pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
} 

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupDfs2(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj, pObjNew *Gia_Obj_t;
    var i int32;
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupDfs2_rec( pNew, p, pObj );
    }
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if ~pObj.Value == 0 {
            pObj.Value = Gia_ManAppendCi(pNew); }
    }
    assert(Gia_ManCiNum(pNew) == Gia_ManCiNum(p), "Gia_ManCiNum(pNew) == Gia_ManCiNum(p)");
    // remap combinational inputs
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObjNew = Gia_ObjFromLit( pNew, pObj.Value );
        assert(!Gia_IsComplement(pObjNew), "!Gia_IsComplement(pObjNew)");
        Vec_IntWriteEntry( pNew.vCis, Gia_ObjCioId(pObj), Gia_ObjId(pNew, pObjNew) );
        Gia_ObjSetCioId( pObjNew, Gia_ObjCioId(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupDfs_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManDupDfs_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Gia_ManDupDfs_rec( pNew, p, Gia_ObjFanin1(pObj) );
    pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}
func Gia_ManDupDfs(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupDfs_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew.nConstrs = p.nConstrs;
    if p.pCexSeq {
        pNew.pCexSeq = Abc_CexDup( p.pCexSeq, Gia_ManRegNum(p) ); }
    return pNew;
}
func Gia_ManDupDfsOnePo(p *Gia_Man_t, iPo int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(iPo >= 0 && iPo < Gia_ManPoNum(p), "iPo >= 0 && iPo < Gia_ManPoNum(p)");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if !Gia_ObjIsPo(p, pObj) || i == iPo {
            Gia_ManDupDfs_rec( pNew, p, Gia_ObjFanin0(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if !Gia_ObjIsPo(p, pObj) || i == iPo {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Cofactors w.r.t. a primary input variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupCofactorVar_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManDupCofactorVar_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Gia_ManDupCofactorVar_rec( pNew, p, Gia_ObjFanin1(pObj) );
    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}
func Gia_ManDupCofactorVar(p *Gia_Man_t, iVar int32, Value int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(iVar >= 0 && iVar < Gia_ManPiNum(p), "iVar >= 0 && iVar < Gia_ManPiNum(p)");
    assert(Value == 0 || Value == 1, "Value == 0 || Value == 1");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    Gia_ManPi( p, iVar ).Value = Value; // modification!
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupCofactorVar_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew.nConstrs = p.nConstrs;
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Gia_ManDupMux(iVar int32, pCof1 *Gia_Man_t, pCof0 *Gia_Man_t) *Gia_Man_t {
    var pGia [2]*Gia_Man_t = [2]*Gia_Man_t{pCof0, pCof1};
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, n int32;
    assert(Gia_ManRegNum(pCof0) == 0, "Gia_ManRegNum(pCof0) == 0");
    assert(Gia_ManRegNum(pCof1) == 0, "Gia_ManRegNum(pCof1) == 0");
    assert(Gia_ManCoNum(pCof0) == 1, "Gia_ManCoNum(pCof0) == 1");
    assert(Gia_ManCoNum(pCof1) == 1, "Gia_ManCoNum(pCof1) == 1");
    assert(Gia_ManCiNum(pCof1) == Gia_ManCiNum(pCof0), "Gia_ManCiNum(pCof1) == Gia_ManCiNum(pCof0)");
    assert(iVar >= 0 && iVar < Gia_ManCiNum(pCof1), "iVar >= 0 && iVar < Gia_ManCiNum(pCof1)");
    pNew = Gia_ManStart( Gia_ManObjNum(pCof1) + Gia_ManObjNum(pCof0) );
    pNew.pName = Abc_UtilStrsav( pCof1.pName );
    pNew.pSpec = Abc_UtilStrsav( pCof1.pSpec );
    Gia_ManHashAlloc( pNew );
    for  n = 0; n < 2; n++  {
        Gia_ManFillValue( pGia[n] );
        Gia_ManConst0(pGia[n]).Value = 0;
        // inlining of macro Gia_ManForEachCi
        for  i = 0; (i < Vec_IntSize(pGia[n].vCis)) && ((pObj) = Gia_ManCi(pGia[n], i)); i++  {
            if n != 0 {
                pObj.Value = Gia_ManCi(pGia[0], i).Value;
            } else {
                pObj.Value = Gia_ManAppendCi(pNew);
            }
        }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(pGia[n].vCos)) && ((pObj) = Gia_ManCo(pGia[n], i)); i++  {
            Gia_ManDupCofactorVar_rec( pNew, pGia[n], Gia_ObjFanin0(pObj) );
        }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pGia[0].vCos)) && ((pObj) = Gia_ManCo(pGia[0], i)); i++  {
        var Ctrl int32  = Gia_ManCi(pGia[0], iVar).Value;
        var Lit1 int32  = Gia_ObjFanin0Copy(Gia_ManCo(pGia[1], i));
        var Lit0 int32  = Gia_ObjFanin0Copy(pObj);
        Gia_ManAppendCo( pNew, Gia_ManHashMux( pNew, Ctrl, Lit1, Lit0 ) );
    }
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Cofactors w.r.t. an internal node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupCofactorObj(p *Gia_Man_t, iObj int32, Value int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var iObjValue int32 = -1;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(iObj > 0 && iObj < Gia_ManObjNum(p), "iObj > 0 && iObj < Gia_ManObjNum(p)");
    assert(Gia_ObjIsCand(Gia_ManObj(p, iObj)), "Gia_ObjIsCand(Gia_ManObj(p, iObj))");
    assert(Value == 0 || Value == 1, "Value == 0 || Value == 1");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ManHashAnd(pNew, Gia_ObjFanin0Copy(pObj), iObjValue) ); }
        if i == iObj {
            iObjValue = Abc_LitNotCond(pObj.Value, !Value);
pObj.Value = Value; }
    }
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Reduce bit-width of GIA assuming it is Boolean.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupBlock(p *Gia_Man_t, nBlock int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t; var i int32;
    assert(Gia_ManCiNum(p) % nBlock == 0, "Gia_ManCiNum(p) % nBlock == 0");
    assert(Gia_ManCoNum(p) % nBlock == 0, "Gia_ManCoNum(p) % nBlock == 0");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if i % nBlock == 0 {
            pObj.Value = Gia_ManAppendCi(pNew);
        } else {
            pObj.Value = 0;
        }
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if i % nBlock == 0 {
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p)/nBlock );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Existentially quantified given variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupExist(p *Gia_Man_t, iVar int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(iVar >= 0 && iVar < Gia_ManPiNum(p), "iVar >= 0 && iVar < Gia_ManPiNum(p)");
    assert(Gia_ManPoNum(p) == 1, "Gia_ManPoNum(p) == 1");
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    Gia_ManFillValue( p );
    // find the cofactoring variable
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    // compute negative cofactor
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    Gia_ManPi( p, iVar ).Value = Abc_Var2Lit( 0, 0 );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ObjFanin0Copy(pObj);
    }
    // compute the positive cofactor
    Gia_ManPi( p, iVar ).Value = Abc_Var2Lit( 0, 1 );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // create OR gate
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ManHashOr(pNew, Gia_ObjFanin0Copy(pObj), pObj.Value) );
    }
    Gia_ManHashStop( pNew );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Existentially quantified given variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupUniv(p *Gia_Man_t, iVar int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(iVar >= 0 && iVar < Gia_ManPiNum(p), "iVar >= 0 && iVar < Gia_ManPiNum(p)");
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    Gia_ManFillValue( p );
    // find the cofactoring variable
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    // compute negative cofactor
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    Gia_ManPi( p, iVar ).Value = Abc_Var2Lit( 0, 0 );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ObjFanin0Copy(pObj);
    }
    // compute the positive cofactor
    Gia_ManPi( p, iVar ).Value = Abc_Var2Lit( 0, 1 );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // create OR gate
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if i == 0 {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ManHashAnd(pNew, Gia_ObjFanin0Copy(pObj), pObj.Value) );
        } else { 
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManHashStop( pNew );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Existentially quantifies the given variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupExist2(p *Gia_Man_t, iVar int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(iVar >= 0 && iVar < Gia_ManPiNum(p), "iVar >= 0 && iVar < Gia_ManPiNum(p)");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManHashAlloc( pNew );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // first part
    Gia_ManPi( p, iVar ).Value = 0; // modification!
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupCofactorVar_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ObjFanin0Copy(pObj);
    }
    // second part
    Gia_ManPi( p, iVar ).Value = 1; // modification!
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = ~0; }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupCofactorVar_rec( pNew, p, Gia_ObjFanin0(pObj) );
    }
    // combination
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ManHashOr(pNew, Gia_ObjFanin0Copy(pObj), pObj.Value) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew.nConstrs = p.nConstrs;
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order while putting CIs first.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupDfsSkip(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if pObj.fMark1 == 0 {
            Gia_ManDupDfs_rec( pNew, p, pObj ); }
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order while putting CIs first.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupDfsCone(p *Gia_Man_t, pRoot *Gia_Obj_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Gia_ObjIsCo(pRoot), "Gia_ObjIsCo(pRoot)");
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    Gia_ManDupDfs_rec( pNew, p, Gia_ObjFanin0(pRoot) );
    Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pRoot) );
    Gia_ManSetRegNum( pNew, 0 );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates logic cone of the literal and inserts it back.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupConeSupp_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, vObjs *Vec_Int_t) {
    var iLit0 int32;
    var iLit1 int32;
    var iObj int32 = Gia_ObjId( p, pObj );
    var iLit int32  = Gia_ObjCopyArray( p, iObj );
    if iLit >= 0 {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManDupConeSupp_rec( pNew, p, Gia_ObjFanin0(pObj), vObjs );
    Gia_ManDupConeSupp_rec( pNew, p, Gia_ObjFanin1(pObj), vObjs );
    iLit0 = Gia_ObjCopyArray( p, Gia_ObjFaninId0(pObj, iObj) );
    iLit1 = Gia_ObjCopyArray( p, Gia_ObjFaninId1(pObj, iObj) );
    iLit0 = Abc_LitNotCond( iLit0, Gia_ObjFaninC0(pObj) );
    iLit1 = Abc_LitNotCond( iLit1, Gia_ObjFaninC1(pObj) );
    iLit  = Gia_ManAppendAnd( pNew, iLit0, iLit1 );
    Gia_ObjSetCopyArray( p, iObj, iLit );
    Vec_IntPush( vObjs, iObj );
}
func Gia_ManDupConeSupp(p *Gia_Man_t, iLit int32, vCiIds *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t; var i, iLit0 int32;
    var pObj *Gia_Obj_t;
    var pRoot *Gia_Obj_t = Gia_ManObj( p, Abc_Lit2Var(iLit) );
    var vObjs *Vec_Int_t  = Vec_IntAlloc( 1000 );
    //assert( Gia_ObjIsAnd(pRoot) );
    if Vec_IntSize(&p.vCopies) < Gia_ManObjNum(p) {
        Vec_IntFillExtra( &p.vCopies, Gia_ManObjNum(p), -1 ); }
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // inlining of macro Gia_ManForEachCiVec
    for  i = 0; (i < Vec_IntSize(vCiIds)) && ((pObj) = Gia_ManCi(p, Vec_IntEntry(vCiIds,i))); i++  {
        Gia_ObjSetCopyArray( p, Gia_ObjId(p, pObj), Gia_ManAppendCi(pNew) );
    }
    Gia_ManDupConeSupp_rec( pNew, p, pRoot, vObjs );
    iLit0 = Gia_ObjCopyArray( p, Abc_Lit2Var(iLit) );
    iLit0 = Abc_LitNotCond( iLit0, Abc_LitIsCompl(iLit) );
    Gia_ManAppendCo( pNew, iLit0 );
    // inlining of macro Gia_ManForEachCiVec
    for  i = 0; (i < Vec_IntSize(vCiIds)) && ((pObj) = Gia_ManCi(p, Vec_IntEntry(vCiIds,i))); i++  {
        Gia_ObjSetCopyArray( p, Gia_ObjId(p, pObj), -1 );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vObjs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vObjs,i))); i++  {
        Gia_ObjSetCopyArray( p, Gia_ObjId(p, pObj), -1 );
    }
    Vec_IntFree( vObjs );
    //assert( Vec_IntCountLarger(&p->vCopies, -1) == 0 );
    return pNew;
}
func Gia_ManDupConeBack_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManDupConeBack_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Gia_ManDupConeBack_rec( pNew, p, Gia_ObjFanin1(pObj) );
    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
//    pObj->Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}
func Gia_ManDupConeBack(p *Gia_Man_t, pNew *Gia_Man_t, vCiIds *Vec_Int_t) int32 {
    var pObj, pRoot *Gia_Obj_t; var i int32;
    assert(Gia_ManCiNum(pNew) == Vec_IntSize(vCiIds), "Gia_ManCiNum(pNew) == Vec_IntSize(vCiIds)");
    Gia_ManFillValue(pNew);
    Gia_ManConst0(pNew).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pNew.vCis)) && ((pObj) = Gia_ManCi(pNew, i)); i++  {
        pObj.Value = Gia_Obj2Lit( p, Gia_ManCi(p, Vec_IntEntry(vCiIds, i)) );
    }
    pRoot = Gia_ManCo(pNew, 0);
    Gia_ManDupConeBack_rec( p, pNew, Gia_ObjFanin0(pRoot) );
    return Gia_ObjFanin0Copy(pRoot);
}
func Gia_ManDupConeBackObjs(p *Gia_Man_t, pNew *Gia_Man_t, vObjs *Vec_Int_t) int32 {
    var pObj, pRoot *Gia_Obj_t; var i int32;
    assert(Gia_ManCiNum(pNew) == Vec_IntSize(vObjs), "Gia_ManCiNum(pNew) == Vec_IntSize(vObjs)");
    Gia_ManFillValue(pNew);
    Gia_ManConst0(pNew).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(pNew.vCis)) && ((pObj) = Gia_ManCi(pNew, i)); i++  {
        pObj.Value = Abc_Var2Lit( Vec_IntEntry(vObjs, i), 0 );
    }
    pRoot = Gia_ManCo(pNew, 0);
    Gia_ManDupConeBack_rec( p, pNew, Gia_ObjFanin0(pRoot) );
    return Gia_ObjFanin0Copy(pRoot);
}


/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order while putting CIs first.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupDfs3_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    if ~pObj.Value {
        return; }
    if Gia_ObjIsCi(pObj) {
        pObj.Value = Gia_ManAppendCi(pNew);
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManDupDfs3_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Gia_ManDupDfs3_rec( pNew, p, Gia_ObjFanin1(pObj) );
    pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}
func Gia_ManDupDfsNode(p *Gia_Man_t, pRoot *Gia_Obj_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    assert(Gia_ObjIsAnd(pRoot), "Gia_ObjIsAnd(pRoot)");
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManDupDfs3_rec( pNew, p, pRoot );
    Gia_ManAppendCo( pNew, pRoot.Value );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order while putting CIs first.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupDfsLitArray(p *Gia_Man_t, vLits *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, iLit, iLitRes int32;
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLits)) && (((iLit) = Vec_IntEntry(vLits, i)), 1); i++  {
        iLitRes = Gia_ManDupDfs2_rec( pNew, p, Gia_ManObj(p, Abc_Lit2Var(iLit)) );
        Gia_ManAppendCo( pNew, Abc_LitNotCond( iLitRes, Abc_LitIsCompl(iLit)) );
    }
    Gia_ManSetRegNum( pNew, 0 );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Returns the array of non-const-0 POs of the dual-output miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupTrimmedNonZero(p *Gia_Man_t) *Vec_Int_t {
    var vNonZero *Vec_Int_t;
    var pTemp, pNonDual *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert((Gia_ManPoNum(p) & 1) == 0, "(Gia_ManPoNum(p) & 1) == 0");
    pNonDual = Gia_ManTransformMiter( p );
    pNonDual = Gia_ManSeqStructSweep( pTemp = pNonDual, 1, 1, 0 );
    Gia_ManStop( pTemp );
    assert(Gia_ManPiNum(pNonDual) > 0, "Gia_ManPiNum(pNonDual) > 0");
    assert(2 * Gia_ManPoNum(pNonDual) == Gia_ManPoNum(p), "2 * Gia_ManPoNum(pNonDual) == Gia_ManPoNum(p)");
    // skip PO pairs corresponding to const0 POs of the non-dual miter
    vNonZero = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(pNonDual)) && ((pObj) = Gia_ManCo(pNonDual, i)); i++  {
        if !Gia_ObjIsConst0(Gia_ObjFanin0(pObj)) {
            Vec_IntPush( vNonZero, i ); }
    }
    Gia_ManStop( pNonDual );
    return vNonZero;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if PO can be removed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPoIsToRemove(p *Gia_Man_t, pObj *Gia_Obj_t, Value int32) int32 {
    assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
    if Value == -1 {
        return Gia_ObjIsConst0(Gia_ObjFanin0(pObj)); }
    assert(Value == 0 || Value == 1, "Value == 0 || Value == 1");
    return Gia_ObjIsConst0(Gia_ObjFanin0(pObj)) && Value == Gia_ObjFaninC0(pObj);
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order while putting CIs first.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupTrimmed(p *Gia_Man_t, fTrimCis int32, fTrimCos int32, fDualOut int32, OutValue int32) *Gia_Man_t {
    var vNonZero *Vec_Int_t  = nil;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, Entry int32;
    // collect non-zero
    if (fDualOut != 0) && (fTrimCos != 0) {
        vNonZero = Gia_ManDupTrimmedNonZero( p ); }
    // start new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // check if there are PIs to be added
    Gia_ManCreateRefs( p );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if !(fTrimCis != 0) || Gia_ObjRefNum(p, pObj) {
            break; }
    }
    if i == Gia_ManPiNum(p) { // there is no PIs - add dummy PI
        Gia_ManAppendCi(pNew); }
    // add the ROs
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if !(fTrimCis != 0) || Gia_ObjRefNum(p, pObj) || Gia_ObjIsRo(p, pObj) {
            pObj.Value = Gia_ManAppendCi(pNew); }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    if (fDualOut != 0) && (fTrimCos != 0) {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vNonZero)) && (((Entry) = Vec_IntEntry(vNonZero, i)), 1); i++  {
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(Gia_ManPo(p, 2*Entry+0)) );
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(Gia_ManPo(p, 2*Entry+1)) );
        }
        if Gia_ManPoNum(pNew) == 0 { // nothing - add dummy PO
        
//            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(Gia_ManPo(p, 0)) );
//            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(Gia_ManPo(p, 1)) );
            Gia_ManAppendCo( pNew, 0 );
            Gia_ManAppendCo( pNew, 0 );
        }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        }
        Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
        // cleanup
        pNew = Gia_ManSeqStructSweep( pTemp = pNew, 1, 1, 0 );
        Gia_ManStop( pTemp );
        // trim the PIs
//        pNew = Gia_ManDupTrimmed( pTemp = pNew, 1, 0, 0 );
//        Gia_ManStop( pTemp );
    
    } else {
        // check if there are POs to be added
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if !(fTrimCos != 0) || !(Gia_ManPoIsToRemove(p, pObj, OutValue) != 0) {
                break; }
        }
        if i == Gia_ManPoNum(p) { // there is no POs - add dummy PO
            Gia_ManAppendCo( pNew, 0 ); }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if !(fTrimCos != 0) || !(Gia_ManPoIsToRemove(p, pObj, OutValue) != 0) || Gia_ObjIsRi(p, pObj) {
                Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
        }
        Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    }
    Vec_IntFreeP( &vNonZero );
    assert(!Gia_ManHasDangling( pNew ), "!Gia_ManHasDangling( pNew )");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Removes POs driven by PIs and PIs without fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupTrimmed2(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // start new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // check if there are PIs to be added
    Gia_ManCreateRefs( p );
    // discount references of POs
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ObjRefFanin0Dec( p, pObj );
    }
    // check if PIs are left
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if Gia_ObjRefNum(p, pObj) {
            break; }
    }
    if i == Gia_ManPiNum(p) { // there is no PIs - add dummy PI
        Gia_ManAppendCi(pNew); }
    // add the ROs
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if Gia_ObjRefNum(p, pObj) || Gia_ObjIsRo(p, pObj) {
            pObj.Value = Gia_ManAppendCi(pNew); }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // check if there are POs to be added
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if !Gia_ObjIsConst0(Gia_ObjFanin0(pObj)) && !Gia_ObjIsPi(p, Gia_ObjFanin0(pObj)) {
            break; }
    }
    if i == Gia_ManPoNum(p) { // there is no POs - add dummy PO
        Gia_ManAppendCo( pNew, 0 ); }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if (!Gia_ObjIsConst0(Gia_ObjFanin0(pObj)) && !Gia_ObjIsPi(p, Gia_ObjFanin0(pObj))) || Gia_ObjIsRi(p, pObj) {
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    assert(!Gia_ManHasDangling( pNew ), "!Gia_ManHasDangling( pNew )");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order while putting CIs first.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupOntop(p *Gia_Man_t, p2 *Gia_Man_t) *Gia_Man_t {
    var pTemp, pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Gia_ManPoNum(p) == Gia_ManPiNum(p2), "Gia_ManPoNum(p) == Gia_ManPiNum(p2)");
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(Gia_ManRegNum(p2) == 0, "Gia_ManRegNum(p2) == 0");
    pNew = Gia_ManStart( Gia_ManObjNum(p)+Gia_ManObjNum(p2) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    // dup first AIG
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // dup second AIG
    Gia_ManConst0(p2).Value = 0;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManPi(p2, i).Value = Gia_ObjFanin0Copy(pObj);
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p2.nObjs) && ((pObj) = Gia_ManObj(p2, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p2.vCos)) && ((pObj) = Gia_ManCo(p2, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
//    Gia_ManPrintStats( pGiaNew, 0 );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates transition relation from p1 and property from p2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupWithNewPo(p1 *Gia_Man_t, p2 *Gia_Man_t) *Gia_Man_t {
    var pTemp, pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // there is no flops in p2
    assert(Gia_ManRegNum(p2) == 0, "Gia_ManRegNum(p2) == 0");
    // there is only one PO in p2
//    assert( Gia_ManPoNum(p2) == 1 );
    // input count of p2 is equal to flop count of p1 
    assert(Gia_ManPiNum(p2) == Gia_ManRegNum(p1), "Gia_ManPiNum(p2) == Gia_ManRegNum(p1)");

    // start new AIG
    pNew = Gia_ManStart( Gia_ManObjNum(p1)+Gia_ManObjNum(p2) );
    pNew.pName = Abc_UtilStrsav( p1.pName );
    pNew.pSpec = Abc_UtilStrsav( p1.pSpec );
    Gia_ManHashAlloc( pNew );
    // dup first AIG
    Gia_ManConst0(p1).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p1.vCis)) && ((pObj) = Gia_ManCi(p1, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p1.nObjs) && ((pObj) = Gia_ManObj(p1, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // dup second AIG
    Gia_ManConst0(p2).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p2)) && ((pObj) = Gia_ManCi(p2, i)); i++  {
        pObj.Value = Gia_ManRo(p1, i).Value;
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p2.nObjs) && ((pObj) = Gia_ManObj(p2, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // add property output
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p2)) && ((pObj) = Gia_ManCo(p2, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // add flop inputs
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p1)) && ((pObj) = Gia_ManCo(p1, Gia_ManPoNum(p1)+i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p1) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Print representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintRepr(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if ~p.pReprsOld[i] {
            fmt.Printf("%d->%d ", i, p.pReprs[i].iRepr); }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupDfsCiMap(p *Gia_Man_t, pCi2Lit *int32, vLits *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
        if ~pCi2Lit[i] {
            pObj.Value = Abc_LitNotCond( Gia_ManObj(p, Abc_Lit2Var(pCi2Lit[i])).Value, Abc_LitIsCompl(pCi2Lit[i]) ); }
    }
    Gia_ManHashAlloc( pNew );
    if vLits != nil {
        var iLit, iLitRes int32;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLits)) && (((iLit) = Vec_IntEntry(vLits, i)), 1); i++  {
            iLitRes = Gia_ManDupDfs2_rec( pNew, p, Gia_ManObj(p, Abc_Lit2Var(iLit)) );
            Gia_ManAppendCo( pNew, Abc_LitNotCond( iLitRes, Abc_LitIsCompl(iLit)) );
        }
    
    } else {
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            Gia_ManDupDfs2_rec( pNew, p, Gia_ObjFanin0(pObj) );
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        }
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Permute inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPermuteInputs(p *Gia_Man_t, nPpis int32, nExtra int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    for  i = 0; i < Gia_ManPiNum(p) - nPpis - nExtra; i++  { // regular PIs
        Gia_ManCi(p, i).Value = Gia_ManAppendCi( pNew );
    }
    for  i = Gia_ManPiNum(p) - nExtra; i < Gia_ManPiNum(p); i++  { // extra PIs due to DC values
        Gia_ManCi(p, i).Value = Gia_ManAppendCi( pNew );
    }
    for  i = Gia_ManPiNum(p) - nPpis - nExtra; i < Gia_ManPiNum(p) - nExtra; i++  { // pseudo-PIs
        Gia_ManCi(p, i).Value = Gia_ManAppendCi( pNew );
    }
    for  i = Gia_ManPiNum(p); i < Gia_ManCiNum(p); i++  { // flop outputs
        Gia_ManCi(p, i).Value = Gia_ManAppendCi( pNew );
    }
    assert(Gia_ManCiNum(pNew) == Gia_ManCiNum(p), "Gia_ManCiNum(pNew) == Gia_ManCiNum(p)");
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupDfsClasses(p *Gia_Man_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(p.pReprsOld != nil, "p.pReprsOld != nil");
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupDfs_rec( pNew, p, pObj );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Detect topmost gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupTopAnd_iter(p *Gia_Man_t, fVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vFront, vLeaves *Vec_Int_t;
    var i int32;
    var iLit int32;
    var iObjId int32;
    var nCiLits int32;
    var pCi2Lit *int32;
    var pVar2Val string;
    // collect the frontier
    vFront = Vec_IntAlloc( 1000 );
    vLeaves = Vec_IntAlloc( 1000 );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if Gia_ObjIsConst0( Gia_ObjFanin0(pObj) ) {
            continue; }
        if Gia_ObjFaninC0(pObj) {
            Vec_IntPush( vLeaves, Gia_ObjFaninLit0p(p, pObj) );
        } else {
            Vec_IntPush( vFront, Gia_ObjFaninId0p(p, pObj) ); }
    }
    if Vec_IntSize(vFront) == 0 {
        if fVerbose != 0 {
            fmt.Printf("The AIG cannot be decomposed using AND-decomposition.\n"); }
        Vec_IntFree( vFront );
        Vec_IntFree( vLeaves );
        return Gia_ManDupNormalize( p, 0 );
    }
    // expand the frontier
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vFront)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vFront,i))); i++  {
        if Gia_ObjIsCi(pObj) {
            Vec_IntPush( vLeaves, Abc_Var2Lit( Gia_ObjId(p, pObj), 0 ) );
            continue;
        }
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        if Gia_ObjFaninC0(pObj) {
            Vec_IntPush( vLeaves, Gia_ObjFaninLit0p(p, pObj) );
        } else {
            Vec_IntPush( vFront, Gia_ObjFaninId0p(p, pObj) ); }
        if Gia_ObjFaninC1(pObj) {
            Vec_IntPush( vLeaves, Gia_ObjFaninLit1p(p, pObj) );
        } else {
            Vec_IntPush( vFront, Gia_ObjFaninId1p(p, pObj) ); }
    }
    Vec_IntFree( vFront );
    // sort the literals
    nCiLits = 0;
    // inlining of macro ABC_FALLOC
    pCi2Lit = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p))));
    // inlining of macro ABC_FALLOC
    pVar2Val = ((char *) memset(malloc(sizeof(char) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(char) * (size_t)(Gia_ManObjNum(p))));
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeaves)) && (((iLit) = Vec_IntEntry(vLeaves, i)), 1); i++  {
        iObjId = Abc_Lit2Var(iLit);
        pObj = Gia_ManObj(p, iObjId);
        if Gia_ObjIsCi(pObj) {
            pCi2Lit[Gia_ObjCioId(pObj)] = !Abc_LitIsCompl(iLit);
            nCiLits++;
        }
        if pVar2Val[iObjId] != 0 && pVar2Val[iObjId] != 1 {
            pVar2Val[iObjId] = Abc_LitIsCompl(iLit);
        } else if pVar2Val[iObjId] != Abc_LitIsCompl(iLit) {
            break; }
    }
    if i < Vec_IntSize(vLeaves) {
        fmt.Printf("Problem is trivially UNSAT.\n");
        // inlining of macro ABC_FREE
        pCi2Lit = nil;
        // inlining of macro ABC_FREE
        pVar2Val = "";
        Vec_IntFree( vLeaves );
        return Gia_ManDupNormalize( p, 0 );
    }
    // create array of input literals
    Vec_IntClear( vLeaves );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if !Gia_ObjIsCi(pObj) && (pVar2Val[i] == 0 || pVar2Val[i] == 1) {
            Vec_IntPush( vLeaves, Abc_Var2Lit(i, pVar2Val[i]) ); }
    }
    if fVerbose != 0 {
        fmt.Printf("Detected %6d AND leaves and %6d CI leaves.\n", Vec_IntSize(vLeaves), nCiLits); }
    // create the input map
    if nCiLits == 0 {
        pNew = Gia_ManDupDfsLitArray( p, vLeaves );
    } else {
        pNew = Gia_ManDupDfsCiMap( p, pCi2Lit, vLeaves ); }
    // inlining of macro ABC_FREE
    pCi2Lit = nil;
    // inlining of macro ABC_FREE
    pVar2Val = "";
    Vec_IntFree( vLeaves );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Detect topmost gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupTopAnd(p *Gia_Man_t, fVerbose int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var fContinue, iIter int32;
    pNew = Gia_ManDupNormalize( p, 0 );
    for  fContinue = 1; fContinue;  {
        pNew = Gia_ManDupTopAnd_iter( pTemp = pNew, fVerbose );
        if Gia_ManCoNum(pNew) == Gia_ManCoNum(pTemp) && Gia_ManAndNum(pNew) == Gia_ManAndNum(pTemp) {
            fContinue = 0; }
        Gia_ManStop( pTemp );
        if fVerbose != 0 {
            fmt.Printf("Iter %2d : ", ++iIter);
            Gia_ManPrintStatsShort( pNew );
        }
    }
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMiter_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if ~pObj.Value {
        return pObj.Value; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManMiter_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Gia_ManMiter_rec( pNew, p, Gia_ObjFanin1(pObj) );
    return pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Creates miter of two designs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMiter(p0 *Gia_Man_t, p1 *Gia_Man_t, nInsDup int32, fDualOut int32, fSeq int32, fImplic int32, fVerbose int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, iLit int32;
    if fSeq != 0 {
        if Gia_ManPiNum(p0) != Gia_ManPiNum(p1) {
            fmt.Printf("Gia_ManMiter(): Designs have different number of PIs.\n");
            return nil;
        }
        if Gia_ManPoNum(p0) != Gia_ManPoNum(p1) {
            fmt.Printf("Gia_ManMiter(): Designs have different number of POs.\n");
            return nil;
        }
        if Gia_ManRegNum(p0) == 0 || Gia_ManRegNum(p1) == 0 {
            fmt.Printf("Gia_ManMiter(): At least one of the designs has no registers.\n");
            return nil;
        }
    
    } else {
        if Gia_ManCiNum(p0) != Gia_ManCiNum(p1) {
            fmt.Printf("Gia_ManMiter(): Designs have different number of CIs.\n");
            return nil;
        }
        if Gia_ManCoNum(p0) != Gia_ManCoNum(p1) {
            fmt.Printf("Gia_ManMiter(): Designs have different number of COs.\n");
            return nil;
        }
    }
    // start the manager
    pNew = Gia_ManStart( Gia_ManObjNum(p0) + Gia_ManObjNum(p1) );
    pNew.pName = Abc_UtilStrsav( "miter" );
    // map combinational inputs
    Gia_ManFillValue( p0 );
    Gia_ManFillValue( p1 );
    Gia_ManConst0(p0).Value = 0;
    Gia_ManConst0(p1).Value = 0;
    // map internal nodes and outputs
    Gia_ManHashAlloc( pNew );
    if fSeq != 0 {
        // create primary inputs
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p0)) && ((pObj) = Gia_ManCi(p0, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p1)) && ((pObj) = Gia_ManCi(p1, i)); i++  {
            if i < Gia_ManPiNum(p1) - nInsDup {
                pObj.Value = Gia_ObjToLit( pNew, Gia_ManPi(pNew, i) );
            } else {
                pObj.Value = Gia_ManAppendCi( pNew ); }
        }
        // create latch outputs
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p0)) && ((pObj) = Gia_ManCi(p0, Gia_ManPiNum(p0)+i)); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p1)) && ((pObj) = Gia_ManCi(p1, Gia_ManPiNum(p1)+i)); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
        // create primary outputs
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p0)) && ((pObj) = Gia_ManCo(p0, i)); i++  {
            Gia_ManMiter_rec( pNew, p0, Gia_ObjFanin0(pObj) );
            Gia_ManMiter_rec( pNew, p1, Gia_ObjFanin0(Gia_ManPo(p1,i)) );
            if fDualOut != 0 {
                Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
                Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(Gia_ManPo(p1,i)) );
            
            } else if fImplic != 0 {
                iLit = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Abc_LitNot(Gia_ObjFanin0Copy(Gia_ManPo(p1,i))) );
                Gia_ManAppendCo( pNew, iLit );
            
            } else {
                iLit = Gia_ManHashXor( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin0Copy(Gia_ManPo(p1,i)) );
                Gia_ManAppendCo( pNew, iLit );
            }
        }
        // create register inputs
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p0)) && ((pObj) = Gia_ManCo(p0, Gia_ManPoNum(p0)+i)); i++  {
            Gia_ManMiter_rec( pNew, p0, Gia_ObjFanin0(pObj) );
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p1)) && ((pObj) = Gia_ManCo(p1, Gia_ManPoNum(p1)+i)); i++  {
            Gia_ManMiter_rec( pNew, p1, Gia_ObjFanin0(pObj) );
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        }
        Gia_ManSetRegNum( pNew, Gia_ManRegNum(p0) + Gia_ManRegNum(p1) );
    
    } else {
        // create combinational inputs
        // inlining of macro Gia_ManForEachCi
        for  i = 0; (i < Vec_IntSize(p0.vCis)) && ((pObj) = Gia_ManCi(p0, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
        // inlining of macro Gia_ManForEachCi
        for  i = 0; (i < Vec_IntSize(p1.vCis)) && ((pObj) = Gia_ManCi(p1, i)); i++  {
            if i < Gia_ManCiNum(p1) - nInsDup {
                pObj.Value = Gia_ObjToLit( pNew, Gia_ManCi(pNew, i) );
            } else {
                pObj.Value = Gia_ManAppendCi( pNew ); }
        }
        // create combinational outputs
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p0.vCos)) && ((pObj) = Gia_ManCo(p0, i)); i++  {
            Gia_ManMiter_rec( pNew, p0, Gia_ObjFanin0(pObj) );
            Gia_ManMiter_rec( pNew, p1, Gia_ObjFanin0(Gia_ManCo(p1,i)) );
            if fDualOut != 0 {
                Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
                Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(Gia_ManCo(p1,i)) );
            
            } else if fImplic != 0 {
                iLit = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Abc_LitNot(Gia_ObjFanin0Copy(Gia_ManPo(p1,i))) );
                Gia_ManAppendCo( pNew, iLit );
            
            } else {
                iLit = Gia_ManHashXor( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin0Copy(Gia_ManCo(p1,i)) );
                Gia_ManAppendCo( pNew, iLit );
            }
        }
    }
    Gia_ManHashStop( pNew );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );

    pNew = Gia_ManDupNormalize( pTemp = pNew, 0 );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Computes the AND of all POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupAndOr(p *Gia_Man_t, nOuts int32, fUseOr int32, fCompl int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, iResult int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    if fUseOr != 0 { // construct OR of all POs
    
        iResult = 0;
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            iResult = Gia_ManHashOr( pNew, iResult, Gia_ObjFanin0Copy(pObj) );
        }
    
    } else { // construct AND of all POs
    
        iResult = 1;
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            iResult = Gia_ManHashAnd( pNew, iResult, Gia_ObjFanin0Copy(pObj) );
        }
    }
    iResult = Abc_LitNotCond( iResult, int32((fCompl > 0)) );
//    Gia_ManForEachPo( p, pObj, i )
//        pObj->Value = Gia_ManAppendCo( pNew, iResult );
    for  i = 0; i < nOuts; i++  {
        Gia_ManAppendCo( pNew, iResult );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Transforms output names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMiterNames(p *Vec_Ptr_t, nOuts int32) *Vec_Ptr_t {
    var pName1 string;
    var pName2 string;
    var pBuffer string // was [1000]byte in C
 var i int32;
    var pNew *Vec_Ptr_t  = Vec_PtrAlloc( Vec_PtrSize(p) - nOuts/2 );
    assert(nOuts % 2 == 0, "nOuts % 2 == 0");
    assert(nOuts <= Vec_PtrSize(p), "nOuts <= Vec_PtrSize(p)");
    // inlining of macro Vec_PtrForEachEntryDouble
    for  i = 0; (i+1 < Vec_PtrSize(p)) && (((pName1) = (char *)Vec_PtrEntry(p, i)), 1) && (((pName2) = (char *)Vec_PtrEntry(p, i+1)), 1); i += 2  {
        if i == nOuts {
            break; }
        pBuffer = fmt.Sprintf("%s_xor_%s", pName1, pName2);
        Vec_PtrPush( pNew, Abc_UtilStrsav(pBuffer) );
    }
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = i; (i < Vec_PtrSize(p)) && (((pName1) = (char *)Vec_PtrEntry(p, i)), 1); i++  {
        Vec_PtrPush( pNew, Abc_UtilStrsav(pName1) );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Transforms the circuit into a regular miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTransformMiter(p *Gia_Man_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObj2 *Gia_Obj_t;
    var i, iLit int32;
    assert((Gia_ManPoNum(p) & 1) == 0, "(Gia_ManPoNum(p) & 1) == 0");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj2 = Gia_ManPo( p, ++i );
        iLit = Gia_ManHashXor( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin0Copy(pObj2) );
        Gia_ManAppendCo( pNew, iLit );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    if p.vNamesIn {
        pNew.vNamesIn = Vec_PtrDupStr(p.vNamesIn); }
    if p.vNamesOut {
        pNew.vNamesOut = Gia_ManMiterNames(p.vNamesOut, Gia_ManPoNum(p)); }
    return pNew;
}
func Gia_ManTransformMiter2(p *Gia_Man_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObj2 *Gia_Obj_t;
    var i int32;
    var iLit int32;
    var nPart int32 = Gia_ManPoNum(p)/2;
    assert((Gia_ManPoNum(p) & 1) == 0, "(Gia_ManPoNum(p) & 1) == 0");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if i == nPart {
            break; }
        pObj2 = Gia_ManPo( p, nPart + i );
        iLit = Gia_ManHashXor( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin0Copy(pObj2) );
        Gia_ManAppendCo( pNew, iLit );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Gia_ManTransformToDual(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        Gia_ManAppendCo( pNew, 0 );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}
func Gia_ManTransformTwoWord2DualOutput(p *Gia_Man_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObj2 *Gia_Obj_t;
    var i int32;
    var nPart int32 = Gia_ManPoNum(p)/2;
    assert((Gia_ManPoNum(p) & 1) == 0, "(Gia_ManPoNum(p) & 1) == 0");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if i == nPart {
            break; }
        pObj2 = Gia_ManPo( p, nPart + i );
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj2) );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

func Gia_ManCollectOneSide_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vNodes *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if !Gia_ObjIsAnd(pObj) {
        return; }
    Gia_ManCollectOneSide_rec( p, Gia_ObjFanin0(pObj), vNodes );
    Gia_ManCollectOneSide_rec( p, Gia_ObjFanin1(pObj), vNodes );
    Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
}
func Gia_ManCollectOneSide(p *Gia_Man_t, iSide int32) *Vec_Int_t {
    var pObj *Gia_Obj_t; var i int32;
    var vNodes *Vec_Int_t  = Vec_IntAlloc( Gia_ManAndNum(p) );
    Gia_ManIncrementTravId( p );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if (i & 1) == iSide {
            Gia_ManCollectOneSide_rec( p, Gia_ObjFanin0(pObj), vNodes ); }
    }
    return vNodes;
}
func Gia_ManTransformDualOutput(p *Gia_Man_t) *Gia_Man_t {
    var vNodes0 *Vec_Int_t  = Gia_ManCollectOneSide( p, 0 );
    var vNodes1 *Vec_Int_t  = Gia_ManCollectOneSide( p, 1 );
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObj2 *Gia_Obj_t;
    var i, fSwap int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert((Gia_ManPoNum(p) & 1) == 0, "(Gia_ManPoNum(p) & 1) == 0");
    if Vec_IntSize(vNodes0) > Vec_IntSize(vNodes1) {
        // inlining of macro ABC_SWAP
        { var t *Vec_Int_t  = vNodes0; vNodes0 = vNodes1; vNodes1 = t; };
        fSwap = 1;
    }
    assert(Vec_IntSize(vNodes0) <= Vec_IntSize(vNodes1), "Vec_IntSize(vNodes0) <= Vec_IntSize(vNodes1)");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes0)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes0,i))); i++  {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes1)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes1,i))); i++  {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
    Vec_IntFree( vNodes0 );
    Vec_IntFree( vNodes1 );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj2 = Gia_ManPo( p, i^fSwap );
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj2) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Performs 'zero' and 'undc' operation.]

  Description [The init string specifies 0/1/X for each flop.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupZeroUndc(p *Gia_Man_t, pInit string, nNewPis int32, fGiaSimple int32, fVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var CountPis int32 = Gia_ManPiNum(p);
    var pPiLits *int32;
    var i int32;
    var iResetFlop int32 = -1;
    var Count1 int32;
    //printf( "Using %s\n", pInit );
    // map X-valued flops into new PIs
    assert(int32(len(pInit)) == Gia_ManRegNum(p), "int32(strlen(pInit)) == Gia_ManRegNum(p)");
    // inlining of macro ABC_FALLOC
    pPiLits = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManRegNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManRegNum(p))));
    for  i = 0; i < Gia_ManRegNum(p); i++  {
        if pInit[i] == 'x' || pInit[i] == 'X' {
            pPiLits[i] = CountPis++; }
    }
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.fGiaSimple = fGiaSimple;
    Gia_ManConst0(p).Value = 0;
    // create primary inputs
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // create additional primary inputs
    for  i = Gia_ManPiNum(p); i < CountPis; i++  {
        Gia_ManAppendCi( pNew );
    }
    // create additional primary inputs
    for  i = 0; i < nNewPis; i++  {
        Gia_ManAppendCi( pNew );
    }
    // create flop outputs
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // create reset flop output
    if CountPis > Gia_ManPiNum(p) {
        iResetFlop = Gia_ManAppendCi( pNew ); }
    // update flop outputs
    Gia_ManMarkFanoutDrivers( p );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if pInit[i] == '1' {
            pObj.Value = Abc_LitNot(pObj.Value), Count1++;
        } else if pInit[i] == 'x' || pInit[i] == 'X' {
            if pObj.fMark0 { // only add MUX if the flop has fanout
                pObj.Value = Gia_ManAppendMux( pNew, iResetFlop, pObj.Value, Gia_Obj2Lit(pNew, Gia_ManPi(pNew, pPiLits[i])) ); }
        
        } else if pInit[i] != '0' {
            assert(false, "0"); }
    }
    Gia_ManCleanMark0( p );
    // inlining of macro ABC_FREE
    pPiLits = nil;
    // build internal nodes
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // create POs
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // create flop inputs
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        if pInit[i] == '1' {
            pObj.Value = Gia_ManAppendCo( pNew, Abc_LitNot(Gia_ObjFanin0Copy(pObj)) );
        } else {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    // create reset flop input
    if CountPis > Gia_ManPiNum(p) {
        Gia_ManAppendCo( pNew, 1 ); }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) + int32((CountPis > Gia_ManPiNum(p))) );
    if fVerbose != 0 {
        fmt.Printf("Converted %d 1-valued FFs and %d DC-valued FFs.\n", Count1, CountPis-Gia_ManPiNum(p)); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Creates miter of two designs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMiter2(pStart *Gia_Man_t, pInit string, fVerbose int32) *Gia_Man_t {
    var vCiValues, vCoValues0, vCoValues1 *Vec_Int_t;
    var pNew, pUndc, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var pInitNew string;
    var i, k int32;
    // check PI values
    for  i = 0; i < Gia_ManPiNum(pStart); i++  {
        assert(pInit[i] == 'x' || pInit[i] == 'X', "pInit[i] == 'x' || pInit[i] == 'X'");
    }
    // normalize the manager
    pUndc = Gia_ManDupZeroUndc( pStart, pInit + Gia_ManPiNum(pStart), 0, 0, fVerbose );
    // create new init string
    // inlining of macro ABC_ALLOC
    pInitNew = ((char *) malloc(sizeof(char) * (size_t)(Gia_ManPiNum(pUndc)+1)));
    for  i = 0; i < Gia_ManPiNum(pStart); i++  {
        pInitNew[i] = pInit[i];
    }
    for  i = k = Gia_ManPiNum(pStart); i < Gia_ManCiNum(pStart); i++  {
        if pInit[i] == 'x' || pInit[i] == 'X' {
            pInitNew[k++] = pInit[i]; }
    }
    pInitNew[k] = 0;
    assert(k == Gia_ManPiNum(pUndc), "k == Gia_ManPiNum(pUndc)");
    // derive miter
    pNew = Gia_ManStart( Gia_ManObjNum(pUndc) );
    pNew.pName = Abc_UtilStrsav( pUndc.pName );
    pNew.pSpec = Abc_UtilStrsav( pUndc.pSpec );
    Gia_ManConst0(pUndc).Value = 0;
    Gia_ManHashAlloc( pNew );
    // add PIs of the first side
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(pUndc)) && ((pObj) = Gia_ManCi(pUndc, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // add PIs of the second side
    vCiValues = Vec_IntAlloc( Gia_ManPiNum(pUndc) );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(pUndc)) && ((pObj) = Gia_ManCi(pUndc, i)); i++  {
        if pInitNew[i] == 'x' {
            Vec_IntPush( vCiValues, Gia_Obj2Lit( pNew, Gia_ManPi(pNew, i) ) );
        } else if pInitNew[i] == 'X' { 
            Vec_IntPush( vCiValues, Gia_ManAppendCi( pNew ) );
        } else { assert(false, "0"); }
    }
    // build flops and internal nodes
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(pUndc)) && ((pObj) = Gia_ManCi(pUndc, Gia_ManPiNum(pUndc)+i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pUndc.nObjs) && ((pObj) = Gia_ManObj(pUndc, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // collect CO values
    vCoValues0 = Vec_IntAlloc( Gia_ManPoNum(pUndc) );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pUndc.vCos)) && ((pObj) = Gia_ManCo(pUndc, i)); i++  {
        Vec_IntPush( vCoValues0, Gia_ObjFanin0Copy(pObj) );
    }
    // build the other side
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(pUndc)) && ((pObj) = Gia_ManCi(pUndc, i)); i++  {
        pObj.Value = Vec_IntEntry( vCiValues, i );
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(pUndc)) && ((pObj) = Gia_ManCi(pUndc, Gia_ManPiNum(pUndc)+i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < pUndc.nObjs) && ((pObj) = Gia_ManObj(pUndc, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // collect CO values
    vCoValues1 = Vec_IntAlloc( Gia_ManPoNum(pUndc) );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pUndc.vCos)) && ((pObj) = Gia_ManCo(pUndc, i)); i++  {
        Vec_IntPush( vCoValues1, Gia_ObjFanin0Copy(pObj) );
    }
    // create POs
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(pUndc)) && ((pObj) = Gia_ManCo(pUndc, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ManHashXor( pNew, Vec_IntEntry(vCoValues0, i), Vec_IntEntry(vCoValues1, i) ) );
    }
    // create flop inputs
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(pUndc)) && ((pObj) = Gia_ManCo(pUndc, Gia_ManPoNum(pUndc)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Vec_IntEntry(vCoValues0, Gia_ManPoNum(pUndc)+i) );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(pUndc)) && ((pObj) = Gia_ManCo(pUndc, Gia_ManPoNum(pUndc)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Vec_IntEntry(vCoValues1, Gia_ManPoNum(pUndc)+i) );
    }
    Vec_IntFree( vCoValues0 );
    Vec_IntFree( vCoValues1 );
    Vec_IntFree( vCiValues );
    // inlining of macro ABC_FREE
    pInitNew = "";
    // cleanup
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, 2*Gia_ManRegNum(pUndc) );
    Gia_ManStop( pUndc );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManChoiceMiter_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if ~pObj.Value {
        return pObj.Value; }
    Gia_ManChoiceMiter_rec( pNew, p, Gia_ObjFanin0(pObj) );
    if Gia_ObjIsCo(pObj) {
        return pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    Gia_ManChoiceMiter_rec( pNew, p, Gia_ObjFanin1(pObj) );
    return pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
} 

/**Function*************************************************************

  Synopsis    [Derives the miter of several AIGs for choice computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManChoiceMiter(vGias *Vec_Ptr_t) *Gia_Man_t {
    var pNew, pGia, pGia0 *Gia_Man_t;
    var i, k, iNode, nNodes int32;
    // make sure they have equal parameters
    assert(Vec_PtrSize(vGias) > 0, "Vec_PtrSize(vGias) > 0");
    pGia0 = (Gia_Man_t *)Vec_PtrEntry( vGias, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGias)) && (((pGia) = (Gia_Man_t *)Vec_PtrEntry(vGias, i)), 1); i++  {
        assert(Gia_ManCiNum(pGia)  == Gia_ManCiNum(pGia0), "Gia_ManCiNum(pGia)  == Gia_ManCiNum(pGia0)");
        assert(Gia_ManCoNum(pGia)  == Gia_ManCoNum(pGia0), "Gia_ManCoNum(pGia)  == Gia_ManCoNum(pGia0)");
        assert(Gia_ManRegNum(pGia) == Gia_ManRegNum(pGia0), "Gia_ManRegNum(pGia) == Gia_ManRegNum(pGia0)");
        Gia_ManFillValue( pGia );
        Gia_ManConst0(pGia).Value = 0;
    }
    // start the new manager
    pNew = Gia_ManStart( Vec_PtrSize(vGias) * Gia_ManObjNum(pGia0) );
    pNew.pName = Abc_UtilStrsav( pGia0.pName );
    pNew.pSpec = Abc_UtilStrsav( pGia0.pSpec );
    // create new CIs and assign them to the old manager CIs
    for  k = 0; k < Gia_ManCiNum(pGia0); k++  {
        iNode = Gia_ManAppendCi(pNew);
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vGias)) && (((pGia) = (Gia_Man_t *)Vec_PtrEntry(vGias, i)), 1); i++  {
            Gia_ManCi( pGia, k ).Value = iNode;
        } 
    }
    // create internal nodes
    Gia_ManHashAlloc( pNew );
    for  k = 0; k < Gia_ManCoNum(pGia0); k++  {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vGias)) && (((pGia) = (Gia_Man_t *)Vec_PtrEntry(vGias, i)), 1); i++  {
            Gia_ManChoiceMiter_rec( pNew, pGia, Gia_ManCo( pGia, k ) );
        }
    }
    Gia_ManHashStop( pNew );
    // check the presence of dangling nodes
    nNodes = Gia_ManHasDangling( pNew );
    //assert( nNodes == 0 );
    // finalize
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(pGia0) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG while putting first PIs, then nodes, then POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupWithConstraints(p *Gia_Man_t, vPoTypes *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, nConstr int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if Vec_IntEntry(vPoTypes, i) == 0 { // regular PO
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if Vec_IntEntry(vPoTypes, i) == 1 { // constraint (should be complemented!)
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ^ 1 ), nConstr++; }
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
//    Gia_ManDupRemapEquiv( pNew, p );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew.nConstrs = nConstr;
    assert(Gia_ManIsNormalized(pNew), "Gia_ManIsNormalized(pNew)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Copy an AIG structure related to the selected POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjCompareByCioId(pp1 **Gia_Obj_t, pp2 **Gia_Obj_t) int32 {
    var pObj1 *Gia_Obj_t  = *pp1;
    var pObj2 *Gia_Obj_t  = *pp2;
    return Gia_ObjCioId(pObj1) - Gia_ObjCioId(pObj2);
}
func Gia_ManDupCones_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vLeaves *Vec_Ptr_t, vNodes *Vec_Ptr_t, vRoots *Vec_Ptr_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsAnd(pObj) {
        Gia_ManDupCones_rec( p, Gia_ObjFanin0(pObj), vLeaves, vNodes, vRoots );
        Gia_ManDupCones_rec( p, Gia_ObjFanin1(pObj), vLeaves, vNodes, vRoots );
        Vec_PtrPush( vNodes, pObj );
    
    } else if Gia_ObjIsCo(pObj) {
        Gia_ManDupCones_rec( p, Gia_ObjFanin0(pObj), vLeaves, vNodes, vRoots );
    } else if Gia_ObjIsRo(p, pObj) {
        Vec_PtrPush( vRoots, Gia_ObjRoToRi(p, pObj) );
    } else if Gia_ObjIsPi(p, pObj) {
        Vec_PtrPush( vLeaves, pObj );
    } else { assert(false, "0"); }
}
func Gia_ManDupCones(p *Gia_Man_t, pPos *int32, nPos int32, fTrimPis int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var vLeaves, vNodes, vRoots *Vec_Ptr_t;
    var pObj *Gia_Obj_t;
    var i int32;

    // collect initial POs
    vLeaves = Vec_PtrAlloc( 100 );
    vNodes = Vec_PtrAlloc( 100 );
    vRoots = Vec_PtrAlloc( 100 );
    for  i = 0; i < nPos; i++  {
        Vec_PtrPush( vRoots, Gia_ManPo(p, pPos[i]) );
    }

    // mark internal nodes
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Gia_ManDupCones_rec( p, pObj, vLeaves, vNodes, vRoots );
    }
    Vec_PtrSort( vLeaves, (int (*)(void))Gia_ObjCompareByCioId );

    // start the new manager
//    Gia_ManFillValue( p );
    pNew = Gia_ManStart( (fTrimPis ? Vec_PtrSize(vLeaves) : Gia_ManCiNum(p)) + Vec_PtrSize(vNodes) + Vec_PtrSize(vRoots) + 1 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // map the constant node
    Gia_ManConst0(p).Value = 0;
    // create PIs
    if fTrimPis != 0 {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
    
    } else {
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
    }
    // create LOs
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = nPos; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Gia_ObjRiToRo(p, pObj).Value = Gia_ManAppendCi( pNew );
    }
    // create internal nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Gia_ObjIsXor(pObj) {
            pObj.Value = Gia_ManAppendXor( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // create COs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // finalize
    Gia_ManSetRegNum( pNew, Vec_PtrSize(vRoots)-nPos );
    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vNodes );
    Vec_PtrFree( vRoots );
    return pNew;

}
func Gia_ManDupAndCones(p *Gia_Man_t, pAnds *int32, nAnds int32, fTrimPis int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var vLeaves, vNodes, vRoots *Vec_Ptr_t;
    var pObj *Gia_Obj_t;
    var i int32;

    // collect initial POs
    vLeaves = Vec_PtrAlloc( 100 );
    vNodes = Vec_PtrAlloc( 100 );
    vRoots = Vec_PtrAlloc( 100 );
    for  i = 0; i < nAnds; i++  {
//        Vec_PtrPush( vRoots, Gia_ManPo(p, pPos[i]) );
        Vec_PtrPush( vRoots, Gia_ManObj(p, pAnds[i]) );
    }

    // mark internal nodes
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Gia_ManDupCones_rec( p, pObj, vLeaves, vNodes, vRoots );
    }
    Vec_PtrSort( vLeaves, (int (*)(void))Gia_ObjCompareByCioId );

    // start the new manager
//    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Vec_PtrSize(vLeaves) + Vec_PtrSize(vNodes) + Vec_PtrSize(vRoots) + 1);
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // map the constant node
    Gia_ManConst0(p).Value = 0;
    // create PIs
    if fTrimPis != 0 {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
    
    } else {
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
    }
    // create LOs
//    Vec_PtrForEachEntryStart( Gia_Obj_t *, vRoots, pObj, i, nPos )
//        Gia_ObjRiToRo(p, pObj)->Value = Gia_ManAppendCi( pNew );
    // create internal nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Gia_ObjIsMux(p, pObj) {
            pObj.Value = Gia_ManAppendMux( pNew, Gia_ObjFanin2Copy(p, pObj), Gia_ObjFanin1Copy(pObj), Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsXor(pObj) {
            pObj.Value = Gia_ManAppendXor( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else { 
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // create COs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
//        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        Gia_ManAppendCo( pNew, pObj.Value );
    }
    // finalize
//    Gia_ManSetRegNum( pNew, Vec_PtrSize(vRoots)-nPos );
    Gia_ManSetRegNum( pNew, 0 );
    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vNodes );
    Vec_PtrFree( vRoots );
    return pNew;

}
func Gia_ManDupAndConesLimit_rec(pNew *Gia_Man_t, p *Gia_Man_t, iObj int32, Level int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj(p, iObj);
    if ~pObj.Value {
        return; }
    if !Gia_ObjIsAnd(pObj) || Gia_ObjLevel(p, pObj) < Level {
        pObj.Value = Gia_ManAppendCi( pNew );
        //printf( "PI %d for %d.\n", Abc_Lit2Var(pObj->Value), iObj );
        return;
    }
    Gia_ManDupAndConesLimit_rec( pNew, p, Gia_ObjFaninId0(pObj, iObj), Level );
    Gia_ManDupAndConesLimit_rec( pNew, p, Gia_ObjFaninId1(pObj, iObj), Level );
    pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    //printf( "Obj %d for %d.\n", Abc_Lit2Var(pObj->Value), iObj );
}
func Gia_ManDupAndConesLimit(p *Gia_Man_t, pAnds *int32, nAnds int32, Level int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var i int32;
    pNew = Gia_ManStart( 1000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManLevelNum( p );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    for  i = 0; i < nAnds; i++  {
        Gia_ManDupAndConesLimit_rec( pNew, p, pAnds[i], Level );
    }
    for  i = 0; i < nAnds; i++  {
        Gia_ManAppendCo( pNew, Gia_ManObj(p, pAnds[i]).Value );
    }
    return pNew;
}

func Gia_ManDupAndConesLimit2_rec(pNew *Gia_Man_t, p *Gia_Man_t, iObj int32, Level int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj(p, iObj);
    if ~pObj.Value {
        return; }
    if !Gia_ObjIsAnd(pObj) || Level <= 0 {
        pObj.Value = Gia_ManAppendCi( pNew );
        //printf( "PI %d for %d.\n", Abc_Lit2Var(pObj->Value), iObj );
        return;
    }
    Gia_ManDupAndConesLimit2_rec( pNew, p, Gia_ObjFaninId0(pObj, iObj), Level-1 );
    Gia_ManDupAndConesLimit2_rec( pNew, p, Gia_ObjFaninId1(pObj, iObj), Level-1 );
    pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    //printf( "Obj %d for %d.\n", Abc_Lit2Var(pObj->Value), iObj );
}
func Gia_ManDupAndConesLimit2(p *Gia_Man_t, pAnds *int32, nAnds int32, Level int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var i int32;
    pNew = Gia_ManStart( 1000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    for  i = 0; i < nAnds; i++  {
        Gia_ManDupAndConesLimit2_rec( pNew, p, pAnds[i], Level );
    }
    for  i = 0; i < nAnds; i++  {
        Gia_ManAppendCo( pNew, Gia_ManObj(p, pAnds[i]).Value );
    }
    return pNew;

}

/**Function*************************************************************

  Synopsis    [Generates AIG representing 1-hot condition for N inputs.]

  Description [The condition is true of all POs are 0.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManOneHot(nSkips int32, nVars int32) *Gia_Man_t {
    var p *Gia_Man_t;
    var i int32;
    var b int32;
    var Shift int32;
    var iGiaLit int32;
    var nLogVars int32 = Abc_Base2Log( nVars );
    // inlining of macro ABC_CALLOC
    var pTemp *int32  = ((int *) calloc((size_t)((1 << nLogVars)), sizeof(int)));
    p = Gia_ManStart( nSkips + 4 * nVars + 1 );
    p.pName = Abc_UtilStrsav( "onehot" );
    for  i = 0; i < nSkips; i++  {
        Gia_ManAppendCi( p );
    }
    for  i = 0; i < nVars; i++  {
        pTemp[i] = Gia_ManAppendCi( p );
    }
    Gia_ManHashStart( p );
    for  b = 0; b < nLogVars; b++  {
        for  i = 0, Shift = (1<<b); i < (1 << nLogVars); i += 2*Shift  {
            iGiaLit = Gia_ManHashAnd( p, pTemp[i], pTemp[i + Shift] );
            if iGiaLit != 0 {
                Gia_ManAppendCo( p, iGiaLit ); }
            pTemp[i] = Gia_ManHashOr( p, pTemp[i], pTemp[i + Shift] );
        }
    }
    Gia_ManHashStop( p );
    Gia_ManAppendCo( p, Abc_LitNot(pTemp[0]) );
    // inlining of macro ABC_FREE
    pTemp = nil;
    assert(Gia_ManObjNum(p) <= nSkips + 4 * nVars + 1, "Gia_ManObjNum(p) <= nSkips + 4 * nVars + 1");
    return p;
}
func Gia_ManDupOneHot(p *Gia_Man_t) *Gia_Man_t {
    var pOneHot *Gia_Man_t;
    var pNew *Gia_Man_t = Gia_ManDup( p );
    if Gia_ManRegNum(pNew) == 0 {
        Abc_Print( 0, "Appending 1-hotness constraints to the PIs.\n" );
        pOneHot = Gia_ManOneHot( 0, Gia_ManCiNum(pNew) );
    
    } else {
        pOneHot = Gia_ManOneHot( Gia_ManPiNum(pNew), Gia_ManRegNum(pNew) ); }
    Gia_ManDupAppendShare( pNew, pOneHot );
    pNew.nConstrs += Gia_ManPoNum(pOneHot);
    Gia_ManStop( pOneHot );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG with nodes ordered by level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupLevelized(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var nLevels int32 = Gia_ManLevelNum( p );
    // inlining of macro ABC_CALLOC
    var pCounts *int32  = ((int *) calloc((size_t)(nLevels + 1), sizeof(int)));
    // inlining of macro ABC_ALLOC
    var pNodes *int32  = ((int *) malloc(sizeof(int) * (size_t)(Gia_ManAndNum(p))));
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pCounts[Gia_ObjLevel(p, pObj)]++; }
    }
    for  i = 1; i <= nLevels; i++  {
        pCounts[i] += pCounts[i-1];
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pNodes[pCounts[Gia_ObjLevel(p, pObj)-1]++] = i; }
    }
    // duplicate
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    for  i = 0; i < Gia_ManAndNum(p) && (pObj = Gia_ManObj(p, pNodes[i])); i++  {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // inlining of macro ABC_FREE
    pCounts = nil;
    // inlining of macro ABC_FREE
    pNodes = nil;
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupFromVecs(p *Gia_Man_t, vCis *Vec_Int_t, vAnds *Vec_Int_t, vCos *Vec_Int_t, nRegs int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // start the new manager
    pNew = Gia_ManStart( 5000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create constant
    Gia_ManConst0(p).Value = 0;
    // create PIs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCis)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vCis,i))); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // create internal nodes
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vAnds)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vAnds,i))); i++  {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
    // create ROs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vCos)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vCos,i))); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, nRegs );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupSliced(p *Gia_Man_t, nSuppMax int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // start the new manager
    pNew = Gia_ManStart( 5000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    // create constant and PIs
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // create internal nodes
    Gia_ManCleanMark01(p);
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Gia_ManSuppSize(p, &i, 1) <= nSuppMax {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
            pObj.fMark0 = 1;
        
        } else {
            Gia_ObjFanin0(pObj).fMark1 = 1;
            Gia_ObjFanin1(pObj).fMark1 = 1;
        }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ObjFanin0(pObj).fMark1 = 1;
    }
    // add POs for the nodes pointed to
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if pObj.fMark0 && pObj.fMark1 {
            Gia_ManAppendCo( pNew, pObj.Value ); }
    }
    // cleanup and leave
    Gia_ManCleanMark01(p);
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Extract constraints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupWithConstrCollectAnd_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t, fFirst int32) {
    if (Gia_IsComplement(pObj) || !Gia_ObjIsAnd(pObj)) && !(fFirst != 0) {
        Vec_IntPushUnique( vSuper, Gia_ObjToLit(p, pObj) );
        return;
    }
    Gia_ManDupWithConstrCollectAnd_rec( p, Gia_ObjChild0(pObj), vSuper, 0 );
    Gia_ManDupWithConstrCollectAnd_rec( p, Gia_ObjChild1(pObj), vSuper, 0 );
}
func Gia_ManDupWithConstr(p *Gia_Man_t) *Gia_Man_t {
    var vSuper *Vec_Int_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var iDriver int32;
    var iLit int32;
    var iLitBest int32 = -1;
    var LevelBest int32 = -1;
    assert(Gia_ManPoNum(p) == 1, "Gia_ManPoNum(p) == 1");
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    pObj = Gia_ManPo( p, 0 );
    if Gia_ObjFaninC0(pObj) {
        fmt.Printf("The miter's output is not AND-decomposable.\n");
        return nil;
    }
    if Gia_ObjFaninId0p(p, pObj) == 0 {
        fmt.Printf("The miter's output is a constant.\n");
        return nil;
    }
    vSuper = Vec_IntAlloc( 100 );
    Gia_ManDupWithConstrCollectAnd_rec( p, Gia_ObjChild0(pObj), vSuper, 1 );
    assert(Vec_IntSize(vSuper) > 1, "Vec_IntSize(vSuper) > 1");
    // find the highest level
    Gia_ManLevelNum( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSuper)) && (((iLit) = Vec_IntEntry(vSuper, i)), 1); i++  {
        if LevelBest < Gia_ObjLevelId(p, Abc_Lit2Var(iLit)) {
            LevelBest = Gia_ObjLevelId(p, Abc_Lit2Var(iLit));
iLitBest = iLit; }
    }
    assert(iLitBest != -1, "iLitBest != -1");
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew ); }
    }
    // create AND of nodes
    iDriver = -1;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSuper)) && (((iLit) = Vec_IntEntry(vSuper, i)), 1); i++  {
        if iLit == iLitBest {
            continue; }
        if iDriver == -1 {
            iDriver = Gia_ObjLitCopy(p, iLit);
        } else {
            iDriver = Gia_ManHashAnd( pNew, iDriver, Gia_ObjLitCopy(p, iLit) ); }
    }
    // create the main PO
    Gia_ManAppendCo( pNew, Gia_ObjLitCopy(p, iLitBest) );
    // create the constraint PO
    Gia_ManAppendCo( pNew, Abc_LitNot(iDriver) );
    pNew.nConstrs = 1;
    // rehash
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    Vec_IntFree( vSuper );
    return pNew;

}

/**Function*************************************************************

  Synopsis    [Compares two objects by their distance.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSortByValue(pp1 **Gia_Obj_t, pp2 **Gia_Obj_t) int32 {
    var Diff int32  = Gia_Regular(*pp1).Value - Gia_Regular(*pp2).Value;
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Decomposes the miter outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupOuts(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Gia_ManAppendCo( pNew, pObj.Value ); }
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    assert(Gia_ManIsNormalized(pNew), "Gia_ManIsNormalized(pNew)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Computes supports for each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCreateNodeSupps(p *Gia_Man_t, vNodes *Vec_Int_t, fVerbose int32) *Vec_Wec_t {
    var clk abctime  = Abc_Clock();
    var pObj *Gia_Obj_t; var i, Id int32;
    var vSuppsNo *Vec_Wec_t  = Vec_WecStart( Vec_IntSize(vNodes) );
    var vSupps *Vec_Wec_t  = Vec_WecStart( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        Vec_IntPush( Vec_WecEntry(vSupps, Id), i );
    }
    // inlining of macro Gia_ManForEachAnd
    for  Id = 0; (Id < p.nObjs) && ((pObj) = Gia_ManObj(p, Id)); Id++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Vec_IntTwoMerge2( Vec_WecEntry(vSupps, Gia_ObjFaninId0(pObj, Id)), 
                          Vec_WecEntry(vSupps, Gia_ObjFaninId1(pObj, Id)), 
                          Vec_WecEntry(vSupps, Id) ); }
    } 
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
        Vec_IntAppend( Vec_WecEntry(vSuppsNo, i), Vec_WecEntry(vSupps, Gia_ObjId(p, pObj)) );
    }
    Vec_WecFree( vSupps );
    if fVerbose != 0 {
        Abc_PrintTime( 1, "Support computation", Abc_Clock() - clk ); }
    return vSuppsNo;
}

func Gia_ManCreateCoSupps(p *Gia_Man_t, fVerbose int32) *Vec_Wec_t {
    var clk abctime  = Abc_Clock();
    var pObj *Gia_Obj_t; var i, Id int32;
    var vSuppsCo *Vec_Wec_t  = Vec_WecStart( Gia_ManCoNum(p) );
    var vSupps *Vec_Wec_t  = Vec_WecStart( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        Vec_IntPush( Vec_WecEntry(vSupps, Id), i );
    }
    // inlining of macro Gia_ManForEachAnd
    for  Id = 0; (Id < p.nObjs) && ((pObj) = Gia_ManObj(p, Id)); Id++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Vec_IntTwoMerge2( Vec_WecEntry(vSupps, Gia_ObjFaninId0(pObj, Id)), 
                          Vec_WecEntry(vSupps, Gia_ObjFaninId1(pObj, Id)), 
                          Vec_WecEntry(vSupps, Id) ); }
    } 
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Vec_IntAppend( Vec_WecEntry(vSuppsCo, i), Vec_WecEntry(vSupps, Gia_ObjFaninId0p(p, pObj)) );
    }
    Vec_WecFree( vSupps );
    if fVerbose != 0 {
        Abc_PrintTime( 1, "Support computation", Abc_Clock() - clk ); }
    return vSuppsCo;
}
func Gia_ManCoSuppSizeMax(p *Gia_Man_t, vSupps *Vec_Wec_t) int32 {
    var pObj *Gia_Obj_t; 
    var vSuppOne *Vec_Int_t;
    var i int32;
    var nSuppMax int32 = 1;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        vSuppOne = Vec_WecEntry( vSupps, i );
        nSuppMax = Abc_MaxInt( nSuppMax, Vec_IntSize(vSuppOne) );
    }
    return nSuppMax;
}
func Gia_ManCoLargestSupp(p *Gia_Man_t, vSupps *Vec_Wec_t) int32 {
    var pObj *Gia_Obj_t; 
    var vSuppOne *Vec_Int_t;
    var i int32;
    var iCoMax int32 = -1;
    var nSuppMax int32 = -1;
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        vSuppOne = Vec_WecEntry( vSupps, i );
        if nSuppMax < Vec_IntSize(vSuppOne) {
            nSuppMax = Vec_IntSize(vSuppOne);
            iCoMax = i;
        }
    }
    return iCoMax;
}
func Gia_ManSortCoBySuppSize(p *Gia_Man_t, vSupps *Vec_Wec_t) *Vec_Int_t {
    var vOrder *Vec_Int_t  = Vec_IntAlloc( Gia_ManCoNum(p) );
    var vSortData *Vec_Wrd_t  = Vec_WrdAlloc( Gia_ManCoNum(p) );
    var vSuppOne *Vec_Int_t; var Entry word; var i int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vSupps)) && (((vSuppOne) = Vec_WecEntry(vSupps, i)), 1); i++  {
        Vec_WrdPush( vSortData, ((word)i << 32) | Vec_IntSize(vSuppOne) );
    }
    Abc_QuickSort3( Vec_WrdArray(vSortData), Vec_WrdSize(vSortData), 1 );
    // inlining of macro Vec_WrdForEachEntry
    for  i = 0; (i < Vec_WrdSize(vSortData)) && (((Entry) = Vec_WrdEntry(vSortData, i)), 1); i++  {
        Vec_IntPush( vOrder, int32((Entry >> 32)) );
    }
    Vec_WrdFree( vSortData );
    return vOrder;
}

/**Function*************************************************************

  Synopsis    [Remaps each CO cone to depend on the first CI variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupHashDfs_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if ~pObj.Value {
        return pObj.Value; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManDupHashDfs_rec( pNew, p, Gia_ObjFanin0(pObj) );
    Gia_ManDupHashDfs_rec( pNew, p, Gia_ObjFanin1(pObj) );
    return pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
} 
func Gia_ManDupCleanDfs_rec(pObj *Gia_Obj_t) {
    if !~pObj.Value {
        return; }
    pObj.Value = ~0;
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManDupCleanDfs_rec( Gia_ObjFanin0(pObj) );
    Gia_ManDupCleanDfs_rec( Gia_ObjFanin1(pObj) );
} 
func Gia_ManDupStrashReduce(p *Gia_Man_t, vSupps *Vec_Wec_t, pvCoMap **Vec_Int_t) *Gia_Man_t {
    var pObj *Gia_Obj_t;
    var pNew, pTemp *Gia_Man_t;
    var vSuppOne, vCoMapLit *Vec_Int_t;
    var i, k, iCi, iLit, nSuppMax int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    Gia_ManFillValue( p );
    vCoMapLit = Vec_IntAlloc( Gia_ManCoNum(p) );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    nSuppMax = Gia_ManCoSuppSizeMax( p, vSupps );
    for  i = 0; i < nSuppMax; i++  {
        Gia_ManAppendCi(pNew);
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        vSuppOne = Vec_WecEntry( vSupps, i );
        if Vec_IntSize(vSuppOne) == 0 {
            Vec_IntPush( vCoMapLit, Abc_Var2Lit(0, Gia_ObjFaninC0(pObj)) );
        } else if Vec_IntSize(vSuppOne) == 1 {
            Vec_IntPush( vCoMapLit, Abc_Var2Lit(1, Gia_ObjFaninC0(pObj)) );
        } else {
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vSuppOne)) && (((iCi) = Vec_IntEntry(vSuppOne, k)), 1); k++  {
                Gia_ManCi(p, iCi).Value = Gia_Obj2Lit(pNew, Gia_ManCi(pNew, k) );
            }
            Gia_ManDupHashDfs_rec( pNew, p, Gia_ObjFanin0(pObj) );
            assert(Gia_ObjFanin0Copy(pObj) < 2 * Gia_ManObjNum(pNew), "Gia_ObjFanin0Copy(pObj) < 2 * Gia_ManObjNum(pNew)");
            Vec_IntPush( vCoMapLit, Gia_ObjFanin0Copy(pObj) );
            Gia_ManDupCleanDfs_rec( Gia_ObjFanin0(pObj) );
        }
    }
    Gia_ManHashStop( pNew );
    assert(Vec_IntSize(vCoMapLit) == Gia_ManCoNum(p), "Vec_IntSize(vCoMapLit) == Gia_ManCoNum(p)");
    if pvCoMap == nil { // do not remap
    
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vCoMapLit)) && (((iLit) = Vec_IntEntry(vCoMapLit, i)), 1); i++  {
            Gia_ManAppendCo( pNew, iLit );
        }
    
    } else { // remap
    
        var vCoMapRes *Vec_Int_t  = Vec_IntAlloc( Gia_ManCoNum(p) );      // map old CO into new CO 
        var vMap *Vec_Int_t  = Vec_IntStartFull( 2*Gia_ManObjNum(pNew) ); // map new lit into new CO
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vCoMapLit)) && (((iLit) = Vec_IntEntry(vCoMapLit, i)), 1); i++  {
            if Vec_IntEntry(vMap, iLit) == -1 {
                Vec_IntWriteEntry( vMap, iLit, Gia_ManCoNum(pNew) );
                Gia_ManAppendCo( pNew, iLit );
            }
            Vec_IntPush( vCoMapRes, Vec_IntEntry(vMap, iLit) );
        }
        Vec_IntFree( vMap );
        *pvCoMap = vCoMapRes;
    }
    Vec_IntFree( vCoMapLit );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Gia_ManIsoStrashReduce2(p *Gia_Man_t, pvPosEquivs **Vec_Ptr_t, fVerbose int32) *Gia_Man_t {
    var vCoMap *Vec_Int_t;
    var vSupps *Vec_Wec_t  = Gia_ManCreateCoSupps( p, fVerbose );
    var pNew *Gia_Man_t  = Gia_ManDupStrashReduce( p, vSupps, &vCoMap );
    Vec_IntFree( vCoMap );
    Vec_WecFree( vSupps );
    *pvPosEquivs = nil;
    return pNew;
}

func Gia_ManIsoStrashReduceOne(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, vSupp *Vec_Int_t) int32 {
    var k, iCi, iLit int32;
    assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
    if Vec_IntSize(vSupp) == 0 {
        return Abc_Var2Lit(0, Gia_ObjFaninC0(pObj)); }
    if Vec_IntSize(vSupp) == 1 {
        return Abc_Var2Lit(1, Gia_ObjFaninC0(pObj)); }
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vSupp)) && (((iCi) = Vec_IntEntry(vSupp, k)), 1); k++  {
        Gia_ManCi(p, iCi).Value = Gia_Obj2Lit(pNew, Gia_ManCi(pNew, k) );
    }
    Gia_ManDupHashDfs_rec( pNew, p, Gia_ObjFanin0(pObj) );
    iLit = Gia_ObjFanin0Copy(pObj);
    Gia_ManDupCleanDfs_rec( Gia_ObjFanin0(pObj) );
    return iLit;
}
func Gia_ManIsoStrashReduceInt(p *Gia_Man_t, vSupps *Vec_Wec_t, fVerbose int32) *Vec_Wec_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vPosEquivs *Vec_Wec_t  = Vec_WecAlloc( 100 );
    var vSuppOne *Vec_Int_t;
    var vMap *Vec_Int_t = Vec_IntAlloc( 10000 );
    var i int32;
    var iLit int32;
    var nSuppMax int32 = Gia_ManCoSuppSizeMax( p, vSupps );
    // count how many times each support size appears
    var vSizeCount *Vec_Int_t  = Vec_IntStart( nSuppMax + 1 );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vSupps)) && (((vSuppOne) = Vec_WecEntry(vSupps, i)), 1); i++  {
        Vec_IntAddToEntry( vSizeCount, Vec_IntSize(vSuppOne), 1 );
    }
    // create array of unique outputs
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    Gia_ManConst0(p).Value = 0;
    for  i = 0; i < nSuppMax; i++  {
        Gia_ManAppendCi(pNew);
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        vSuppOne = Vec_WecEntry( vSupps, i );
        if Vec_IntEntry(vSizeCount, Vec_IntSize(vSuppOne)) == 1 {
            Vec_IntPush( Vec_WecPushLevel(vPosEquivs), i );
            continue;
        }
        iLit = Gia_ManIsoStrashReduceOne( pNew, p, pObj, vSuppOne );
        Vec_IntFillExtra( vMap, iLit + 1, -1 );
        if Vec_IntEntry(vMap, iLit) == -1 {
            Vec_IntWriteEntry( vMap, iLit, Vec_WecSize(vPosEquivs) );
            Vec_IntPush( Vec_WecPushLevel(vPosEquivs), i );
            continue;
        }
        Vec_IntPush( Vec_WecEntry(vPosEquivs, Vec_IntEntry(vMap, iLit)), i );
    }
    Gia_ManHashStop( pNew );
    Gia_ManStop( pNew );
    Vec_IntFree( vSizeCount );
    Vec_IntFree( vMap );
    return vPosEquivs;
}
func Gia_ManIsoStrashReduce(p *Gia_Man_t, pvPosEquivs **Vec_Ptr_t, fVerbose int32) *Gia_Man_t {
    var vSupps *Vec_Wec_t  = Gia_ManCreateCoSupps( p, fVerbose );
    var vPosEquivs *Vec_Wec_t  = Gia_ManIsoStrashReduceInt( p, vSupps, fVerbose );
    // find the first outputs and derive GIA
    var vFirsts *Vec_Int_t  = Vec_WecCollectFirsts( vPosEquivs );
    var pNew *Gia_Man_t  = Gia_ManDupCones( p, Vec_IntArray(vFirsts), Vec_IntSize(vFirsts), 0 );
    Vec_IntFree( vFirsts );
    Vec_WecFree( vSupps );
    // report and return    
    if fVerbose != 0 { 
        fmt.Printf("Nontrivial classes:\n");
        Vec_WecPrint( vPosEquivs, 1 );
    }
    if pvPosEquivs {
        *pvPosEquivs = Vec_WecConvertToVecPtr( vPosEquivs ); }
    Vec_WecFree( vPosEquivs );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Decomposes the miter outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupDemiter(p *Gia_Man_t, fVerbose int32) *Gia_Man_t {
    var vSuper *Vec_Int_t;
    var vSuperPtr *Vec_Ptr_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjPo *Gia_Obj_t;
    var i, iLit int32;
    assert(Gia_ManPoNum(p) == 1, "Gia_ManPoNum(p) == 1");
    // decompose
    pObjPo = Gia_ManPo( p, 0 );
    vSuper = Vec_IntAlloc( 100 );
    Gia_ManDupWithConstrCollectAnd_rec( p, Gia_ObjFanin0(pObjPo), vSuper, 1 );
    assert(Vec_IntSize(vSuper) > 1, "Vec_IntSize(vSuper) > 1");
    // report the result
    fmt.Printf("The miter is %s-decomposable into %d parts.\n", Gia_ObjFaninC0(pObjPo) ? "OR":"AND", Vec_IntSize(vSuper));
    // create levels
    Gia_ManLevelNum( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSuper)) && (((iLit) = Vec_IntEntry(vSuper, i)), 1); i++  {
        Gia_ManObj(p, Abc_Lit2Var(iLit)).Value = Gia_ObjLevelId(p, Abc_Lit2Var(iLit));
    }
    // create pointer array
    vSuperPtr = Vec_PtrAlloc( Vec_IntSize(vSuper) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSuper)) && (((iLit) = Vec_IntEntry(vSuper, i)), 1); i++  {
        Vec_PtrPush( vSuperPtr, Gia_Lit2Obj(p, iLit) );
    }
    Vec_PtrSort( vSuperPtr, (int (*)(void))Gia_ManSortByValue );
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // create the outputs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuperPtr)) && (((pObj) = (Gia_Obj_t *)Vec_PtrEntry(vSuperPtr, i)), 1); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjLitCopy(p, Gia_Obj2Lit(p, pObj)) ^ Gia_ObjFaninC0(pObjPo) );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // rehash
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    Vec_IntFree( vSuper );
    Vec_PtrFree( vSuperPtr );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupDemiterOrderXors2(p *Gia_Man_t, vXors *Vec_Int_t) {
    var i int32;
    var iObj int32;
    var pPerm *int32;
    var vSizes *Vec_Int_t  = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXors)) && (((iObj) = Vec_IntEntry(vXors, i)), 1); i++  {
        Vec_IntPush( vSizes, Gia_ManSuppSize(p, &iObj, 1) );
    }
    pPerm = Abc_MergeSortCost( Vec_IntArray(vSizes), Vec_IntSize(vSizes) );
    Vec_IntClear( vSizes );
    for  i = 0; i < Vec_IntSize(vXors); i++  {
        Vec_IntPush( vSizes, Vec_IntEntry(vXors, pPerm[i]) );
    }
    // inlining of macro ABC_FREE
    pPerm = nil;
    Vec_IntClear( vXors );
    Vec_IntAppend( vXors, vSizes );
    Vec_IntFree( vSizes );
}
func Gia_ManDupDemiterFindMin(vSupps *Vec_Wec_t, vTakenIns *Vec_Int_t, vTakenOuts *Vec_Int_t) int32 {
    var vLevel *Vec_Int_t;
    var i int32;
    var k int32;
    var iObj int32;
    var iObjBest int32 = -1;
    var Count int32;
    var CountBest int32 = ABC_INFINITY;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vSupps)) && (((vLevel) = Vec_WecEntry(vSupps, i)), 1); i++  {
        if Vec_IntEntry(vTakenOuts, i) {
            continue; }
        Count = 0;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((iObj) = Vec_IntEntry(vLevel, k)), 1); k++  {
            Count += !Vec_IntEntry(vTakenIns, iObj);
        }
        if CountBest > Count {
            CountBest = Count;
            iObjBest = i;
        }
    }
    return iObjBest;
}
func Gia_ManDupDemiterOrderXors(p *Gia_Man_t, vXors *Vec_Int_t) {
    extern Vec_Wec_t * Gia_ManCreateNodeSupps( Gia_Man_t * p, Vec_Int_t * vNodes, int fVerbose );
    var vSupps *Vec_Wec_t  = Gia_ManCreateNodeSupps( p, vXors, 0 );
    var vTakenIns *Vec_Int_t  = Vec_IntStart( Gia_ManCiNum(p) );
    var vTakenOuts *Vec_Int_t  = Vec_IntStart( Vec_IntSize(vXors) );
    var vOrder *Vec_Int_t  = Vec_IntAlloc( Vec_IntSize(vXors) );
    var i, k, iObj int32;
    // add outputs in the order of increasing supports
    for  i = 0; i < Vec_IntSize(vXors); i++  {
        var Index int32  = Gia_ManDupDemiterFindMin( vSupps, vTakenIns, vTakenOuts );
        assert(Index >= 0 && Index < Vec_IntSize(vXors), "Index >= 0 && Index < Vec_IntSize(vXors)");
        Vec_IntPush( vOrder, Vec_IntEntry(vXors, Index) );
        assert(!Vec_IntEntry( vTakenOuts, Index ), "!Vec_IntEntry( vTakenOuts, Index )");
        Vec_IntWriteEntry( vTakenOuts, Index, 1 );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(Vec_WecEntry(vSupps, Index))) && (((iObj) = Vec_IntEntry(Vec_WecEntry(vSupps, Index), k)), 1); k++  {
            Vec_IntWriteEntry( vTakenIns, iObj, 1 );
        }
    }
    Vec_WecFree( vSupps );
    Vec_IntFree( vTakenIns );
    Vec_IntFree( vTakenOuts );
    // reload
    Vec_IntClear( vXors );
    Vec_IntAppend( vXors, vOrder );
    Vec_IntFree( vOrder );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSetMark0Dfs_rec(p *Gia_Man_t, iObj int32) {
    var pObj *Gia_Obj_t;
    pObj = Gia_ManObj( p, iObj );
    if pObj.fMark0 {
        return; }
    pObj.fMark0 = 1;
    if !Gia_ObjIsAnd(pObj) {
        return; }
    Gia_ManSetMark0Dfs_rec( p, Gia_ObjFaninId0(pObj, iObj) );
    Gia_ManSetMark0Dfs_rec( p, Gia_ObjFaninId1(pObj, iObj) );
}
func Gia_ManSetMark1Dfs_rec(p *Gia_Man_t, iObj int32) {
    var pObj *Gia_Obj_t;
    pObj = Gia_ManObj( p, iObj );
    if pObj.fMark1 {
        return; }
    pObj.fMark1 = 1;
    if !Gia_ObjIsAnd(pObj) {
        return; }
    Gia_ManSetMark1Dfs_rec( p, Gia_ObjFaninId0(pObj, iObj) );
    Gia_ManSetMark1Dfs_rec( p, Gia_ObjFaninId1(pObj, iObj) );
}

func Gia_ManCountMark0Dfs_rec(p *Gia_Man_t, iObj int32) int32 {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return 0; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if !Gia_ObjIsAnd(pObj) {
        return pObj.fMark0; }
    return Gia_ManCountMark0Dfs_rec( p, Gia_ObjFaninId0(pObj, iObj) ) + 
           Gia_ManCountMark0Dfs_rec( p, Gia_ObjFaninId1(pObj, iObj) ) + pObj.fMark0;
}
func Gia_ManCountMark0Dfs(p *Gia_Man_t, iObj int32) int32 {
    Gia_ManIncrementTravId( p );
    return Gia_ManCountMark0Dfs_rec( p, iObj );
}
func Gia_ManCountMark1Dfs_rec(p *Gia_Man_t, iObj int32) int32 {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return 0; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if !Gia_ObjIsAnd(pObj) {
        return pObj.fMark1; }
    return Gia_ManCountMark1Dfs_rec( p, Gia_ObjFaninId0(pObj, iObj) ) + 
           Gia_ManCountMark1Dfs_rec( p, Gia_ObjFaninId1(pObj, iObj) ) + pObj.fMark1;
}
func Gia_ManCountMark1Dfs(p *Gia_Man_t, iObj int32) int32 {
    Gia_ManIncrementTravId( p );
    return Gia_ManCountMark1Dfs_rec( p, iObj );
}

func Gia_ManDecideWhereToAdd(p *Gia_Man_t, vPart []*Vec_Int_t, pFan []*Gia_Obj_t) int32 {
    var Count0 int32 = 1;
    var Count1 int32;
    assert(Vec_IntSize(vPart[0]) == Vec_IntSize(vPart[1]), "Vec_IntSize(vPart[0]) == Vec_IntSize(vPart[1])");
    if Vec_IntSize(vPart[0]) > 0 {
        Count0 = Gia_ManCountMark0Dfs(p, Gia_ObjId(p, pFan[0])) + Gia_ManCountMark1Dfs(p, Gia_ObjId(p, pFan[1]));
        Count1 = Gia_ManCountMark0Dfs(p, Gia_ObjId(p, pFan[1])) + Gia_ManCountMark1Dfs(p, Gia_ObjId(p, pFan[0]));
    }
    return Count0 < Count1;
}
func Gia_ManCollectTopXors_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vXors *Vec_Int_t) {
    var pFan0, pFan1 *Gia_Obj_t;
    var iObj int32  = Gia_ObjId( p, pObj );
    if Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) || !Gia_ObjIsAnd(pObj) {
        Vec_IntPushUnique( vXors, Gia_ObjId(p, pObj) );
        return;
    }
    if Gia_ObjFaninC0(pObj) {
        Vec_IntPushUnique( vXors, Gia_ObjFaninId0(pObj, iObj) );
    } else {
        Gia_ManCollectTopXors_rec( p, Gia_ObjFanin0(pObj), vXors ); }
    if Gia_ObjFaninC1(pObj) {
        Vec_IntPushUnique( vXors, Gia_ObjFaninId1(pObj, iObj) );
    } else {
        Gia_ManCollectTopXors_rec( p, Gia_ObjFanin1(pObj), vXors ); }
}
func Gia_ManCollectTopXors(p *Gia_Man_t) *Vec_Int_t {
    var i, iObj, iObj2, fFlip, Count1 int32;
    var vXors *Vec_Int_t;
    var vPart [2]*Vec_Int_t;
    var vOrder *Vec_Int_t; 
    var pFan [2]*Gia_Obj_t;
    var pObj *Gia_Obj_t = Gia_ManCo(p, 0);
    vXors = Vec_IntAlloc( 100 );
    if Gia_ManCoNum(p) == 1 {
        if Gia_ObjFaninC0(pObj) {
            Gia_ManCollectTopXors_rec( p, Gia_ObjFanin0(pObj), vXors );
        } else {
            Vec_IntPush( vXors, Gia_ObjId(p, Gia_ObjFanin0(pObj)) ); }
    
    } else {
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if Gia_ObjFaninId0p(p, pObj) > 0 {
                Vec_IntPush( vXors, Gia_ObjFaninId0p(p, pObj) ); }
        }
    }
    // order by support size
    Gia_ManDupDemiterOrderXors( p, vXors );
    //Vec_IntPrint( vXors );
    Vec_IntReverseOrder( vXors ); // from MSB to LSB
    // divide into groups
    Gia_ManCleanMark01(p);
    vPart[0] = Vec_IntAlloc( 100 );
    vPart[1] = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vXors)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vXors,i))); i++  {
        var fCompl int32;
        if !Gia_ObjRecognizeExor(pObj, &pFan[0], &pFan[1]) {
            pFan[0] = pObj, pFan[1] = Gia_ManConst0(p), Count1++;
        } else {
            fCompl ^= Gia_IsComplement(pFan[0]);
            fCompl ^= Gia_IsComplement(pFan[1]);
            pFan[0] = Gia_Regular(pFan[0]);
            pFan[1] = Gia_Regular(pFan[1]);
        }
        fFlip = Gia_ManDecideWhereToAdd( p, vPart, pFan );
        Vec_IntPush( vPart[0], Gia_ObjId(p, pFan[fFlip]) );
        Vec_IntPush( vPart[1], Gia_ObjId(p, pFan[!fFlip]) );
        Gia_ManSetMark0Dfs_rec( p, Gia_ObjId(p, pFan[fFlip]) );
        Gia_ManSetMark1Dfs_rec( p, Gia_ObjId(p, pFan[!fFlip]) );
    }
    //printf( "Detected %d single-output XOR miters and %d other miters.\n", Vec_IntSize(vXors) - Count1, Count1 );
    Vec_IntFree( vXors );
    Gia_ManCleanMark01(p);
    // create new order
    vOrder = Vec_IntAlloc( 100 );
    // inlining of macro Vec_IntForEachEntryTwo
    for  i = 0; (i < Vec_IntSize(vPart[0])) && (((iObj) = Vec_IntEntry(vPart[0], i)), 1) && (((iObj2) = Vec_IntEntry(vPart[1], i)), 1); i++  {
        Vec_IntPushTwo( vOrder, iObj, iObj2 );
    }
    Vec_IntFree( vPart[0] );
    Vec_IntFree( vPart[1] );
    Vec_IntReverseOrder( vOrder ); // from LSB to MSB
    //Vec_IntPrint( vOrder );
    return vOrder;
}
func Gia_ManDemiterToDual(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t; var pObj *Gia_Obj_t; var i int32;
    var vNodes *Vec_Int_t;
    var vOrder *Vec_Int_t  = Gia_ManCollectTopXors( p );
    if vOrder == nil {
        fmt.Printf("Cannot demiter because the top-most gate is an AND-gate.\n");
        return nil;
    }
    assert(Vec_IntSize(vOrder) % 2 == 0, "Vec_IntSize(vOrder) % 2 == 0");
    vNodes = Vec_IntAlloc( Gia_ManObjNum(p) );
    Gia_ManIncrementTravId( p );
    Gia_ManCollectAnds( p, Vec_IntArray(vOrder), Vec_IntSize(vOrder), vNodes, nil );
    pNew = Gia_ManStart( 1 + Gia_ManCiNum(p) + Vec_IntSize(vNodes) + Vec_IntSize(vOrder) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
    pObj = Gia_ManCo(p, 0);
    if Gia_ObjFanin0(pObj) == Gia_ManConst0(p) {
        Gia_ManAppendCo( pNew, 0 );
        Gia_ManAppendCo( pNew, Gia_ObjFaninC0(pObj) );
    
    } else {
        Gia_ManSetPhase( p );
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vOrder)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vOrder,i))); i++  {
            Gia_ManAppendCo( pNew, Abc_LitNotCond(pObj.Value, pObj.fPhase) );
        }
    }
    Vec_IntFree( vNodes );
    Vec_IntFree( vOrder );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Collect nodes reachable from odd/even outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectDfs_rec(p *Gia_Man_t, iObj int32, vNodes *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    if !Gia_ObjIsAnd(pObj) {
        return; }
    Gia_ManCollectDfs_rec( p, Gia_ObjFaninId0(pObj, iObj), vNodes );
    Gia_ManCollectDfs_rec( p, Gia_ObjFaninId1(pObj, iObj), vNodes );
    Vec_IntPush( vNodes, iObj );
}
func Gia_ManCollectReach(p *Gia_Man_t, fOdd int32) *Vec_Int_t {
    var i, iDriver int32;
    var vNodes *Vec_Int_t  = Vec_IntAlloc( Gia_ManObjNum(p) );
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && (((iDriver) = Gia_ObjFaninId0p(p, Gia_ManCo(p, i))), 1); i++  {
        if (i & 1) == fOdd {
            Gia_ManCollectDfs_rec( p, iDriver, vNodes ); }
    }
    return vNodes;
}
func Gia_ManDemiterDual(p *Gia_Man_t, pp0 **Gia_Man_t, pp1 **Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, fOdd int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(Gia_ManCoNum(p) % 2 == 0, "Gia_ManCoNum(p) % 2 == 0");
    *pp0 = *pp1 = nil;
    for  fOdd = 0; fOdd < 2; fOdd++  {
        var vNodes *Vec_Int_t  = Gia_ManCollectReach( p, fOdd );
        var pNew *Gia_Man_t  = Gia_ManStart( 1 + Gia_ManCiNum(p) + Vec_IntSize(vNodes) + Gia_ManCoNum(p)/2 );
        pNew.pName = Abc_UtilStrsav( p.pName );
        pNew.pSpec = Abc_UtilStrsav( p.pSpec );
        Gia_ManConst0(p).Value = 0;
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if (i & 1) == fOdd {
                Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
        }
        Vec_IntFree( vNodes );
        if fOdd != 0 {
            *pp1 = pNew;
        } else {
            *pp0 = pNew; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Collect nodes reachable from first/second half of outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectReach2(p *Gia_Man_t, fSecond int32) *Vec_Int_t {
    var i, iDriver int32;
    var vNodes *Vec_Int_t  = Vec_IntAlloc( Gia_ManObjNum(p) );
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachCoDriverId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && (((iDriver) = Gia_ObjFaninId0p(p, Gia_ManCo(p, i))), 1); i++  {
        if (i < Gia_ManCoNum(p)/2) ^ fSecond {
            Gia_ManCollectDfs_rec( p, iDriver, vNodes ); }
    }
    return vNodes;
}
func Gia_ManDemiterTwoWords(p *Gia_Man_t, pp0 **Gia_Man_t, pp1 **Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, fSecond int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(Gia_ManCoNum(p) % 2 == 0, "Gia_ManCoNum(p) % 2 == 0");
    *pp0 = *pp1 = nil;
    for  fSecond = 0; fSecond < 2; fSecond++  {
        var vNodes *Vec_Int_t  = Gia_ManCollectReach2( p, fSecond );
        var pNew *Gia_Man_t  = Gia_ManStart( 1 + Gia_ManCiNum(p) + Vec_IntSize(vNodes) + Gia_ManCoNum(p)/2 );
        pNew.pName = Abc_UtilStrsav( p.pName );
        pNew.pSpec = Abc_UtilStrsav( p.pSpec );
        Gia_ManConst0(p).Value = 0;
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            if (i < Gia_ManCoNum(p)/2) ^ fSecond {
                Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
        }
        Vec_IntFree( vNodes );
        if fSecond != 0 {
            *pp1 = pNew;
        } else {
            *pp0 = pNew; }
    }
    return 1;
}



/**Function*************************************************************

  Synopsis    [Extracts "half" of the sequential AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupHalfSeq(p *Gia_Man_t, fSecond int32) *Gia_Man_t {
    var i int32; var pObj *Gia_Obj_t;
    var pNew *Gia_Man_t  = Gia_ManStart( Gia_ManObjNum(p) ); 
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManFillValue(p);
    Gia_ManConst0(p).Value = 0;
    if fSecond != 0 {
        // inlining of macro Gia_ManForEachCi
        for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            Gia_ManDupOrderDfs_rec( pNew, p, pObj );
        }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
            if i >= Gia_ManRegNum(p)/2 {
                Gia_ManDupOrderDfs_rec( pNew, p, pObj ); }
        }
        Gia_ManSetRegNum( pNew, Gia_ManRegNum(p)- Gia_ManRegNum(p)/2 );
    
    } else {
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pNew );
        }
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            if i >= Gia_ManRegNum(p)/2 {
                pObj.Value = Gia_ManAppendCi( pNew ); }
        }
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            if i < Gia_ManRegNum(p)/2 {
                pObj.Value = Gia_ManAppendCi( pNew ); }
        }
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            Gia_ManDupOrderDfs_rec( pNew, p, pObj );
        }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
            if i < Gia_ManRegNum(p)/2 {
                Gia_ManDupOrderDfs_rec( pNew, p, pObj ); }
        }
        Gia_ManSetRegNum( pNew, Gia_ManRegNum(p)/2 );
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Merge two sets of sequential equivalences.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSeqEquivMerge(p *Gia_Man_t, pPart []*Gia_Man_t) {
    // inlining of macro ABC_FALLOC
    int i, iObj, * pClasses    = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p))));
    // inlining of macro ABC_FALLOC
    int n, Repr, * pClass2Repr = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p))));
    // initialize equiv class representation in the big AIG
    assert(p.pReprs  == nil && p.pNexts  == nil, "p.pReprs  == nil && p.pNexts  == nil");
    // inlining of macro ABC_CALLOC
    p.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Gia_Rpr_t)));
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        Gia_ObjSetRepr( p, i, GIA_VOID );
    }
    // map equivalences of p into classes
    pClasses[0] = 0;
    for  n = 0; n < 2; n++  {
        assert(pPart[n].pReprs != nil && pPart[n].pNexts != nil, "pPart[n].pReprs != nil && pPart[n].pNexts != nil");
        for  i = 0; i < Gia_ManObjNum(pPart[n]); i++  {
            if Gia_ObjRepr(pPart[n], i) == 0 {
                pClasses[Gia_ManObj(pPart[n], i).Value] = 0; }
        }
        // inlining of macro Gia_ManForEachClass
        for  i = 1; i < Gia_ManObjNum(pPart[n]); i++  { if !Gia_ObjIsHead(pPart[n], i) {} else {
            Repr = Gia_ManObj(pPart[n], i).Value;
            if n == 1 {
                // inlining of macro Gia_ClassForEachObj
                for  assert(Gia_ObjIsHead(pPart[n], i) && (i != 0), "Gia_ObjIsHead(pPart[n], i) && i"), iObj = i; iObj > 0; iObj = Gia_ObjNext(pPart[n], iObj)  {
                    if pClasses[Gia_ManObj(pPart[n], iObj).Value] != -1 {
                        Repr = pClasses[Gia_ManObj(pPart[n], iObj).Value]; }
                }
            }
            // inlining of macro Gia_ClassForEachObj
            for  assert(Gia_ObjIsHead(pPart[n], i) && (i != 0), "Gia_ObjIsHead(pPart[n], i) && i"), iObj = i; iObj > 0; iObj = Gia_ObjNext(pPart[n], iObj)  {
                pClasses[Gia_ManObj(pPart[n], iObj).Value] = Repr;
            }
        }
        }
    }
    // map representatives of each class
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        if pClasses[i] != -1 && pClass2Repr[pClasses[i]] == -1 {
            pClass2Repr[pClasses[i]] = i;
            pClasses[i] = -1;
        }
    }
    // remap the remaining classes
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        if pClasses[i] != -1 {
            p.pReprs[i].iRepr = pClass2Repr[pClasses[i]]; }
    }
    // inlining of macro ABC_FREE
    pClasses = nil;
    // inlining of macro ABC_FREE
    pClass2Repr = nil;
    // create next pointers
    p.pNexts = Gia_ManDeriveNexts( p );
}

/**Function*************************************************************

  Synopsis    [Print equivalences.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintEquivs(p *Gia_Man_t) {
    var pObj *Gia_Obj_t; var i, iObj int32;
    fmt.Printf("Const0:");
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Gia_ObjRepr(p, i) == 0 {
            fmt.Printf(" %d", i); }
    }
    fmt.Printf("\n");
    // inlining of macro Gia_ManForEachClass
    for  i = 1; i < Gia_ManObjNum(p); i++  { if !Gia_ObjIsHead(p, i) {} else {
        fmt.Printf("%d:", i);
        // inlining of macro Gia_ClassForEachObj1
        for  assert(Gia_ObjIsHead(p, i), "Gia_ObjIsHead(p, i)"), iObj = Gia_ObjNext(p, i); iObj > 0; iObj = Gia_ObjNext(p, iObj)  {
            fmt.Printf(" %d", iObj);
        }
        fmt.Printf("\n");
    }
    }
}

/**Function*************************************************************

  Synopsis    [Computing seq equivs by dividing AIG into two parts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSeqEquivDivide(p *Gia_Man_t, pPars *Cec_ParCor_t) {
    var pParts [2]*Gia_Man_t;
    var pObj *Gia_Obj_t;
    var n, i int32;
    for  n = 0; n < 2; n++  {
        // derive n-th part of the AIG
        pParts[n] = Gia_ManDupHalfSeq( p, n );
        //Gia_ManPrintStats( pParts[n], NULL );
        // compute equivalences (recorded internally using pReprs and pNexts)
        Cec_ManLSCorrespondenceClasses( pParts[n], pPars );
        // make the nodes of the part AIG point to their prototypes in the AIG
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if ~pObj.Value {
                Gia_ManObj( pParts[n], Abc_Lit2Var(pObj.Value) ).Value = i; }
        }
    }
    Gia_ManSeqEquivMerge( p, pParts );
    Gia_ManStop( pParts[0] );
    Gia_ManStop( pParts[1] );
}
func Gia_ManScorrDivideTest(p *Gia_Man_t, pPars *Cec_ParCor_t) *Gia_Man_t {
    extern Gia_Man_t * Gia_ManCorrReduce( Gia_Man_t * p );
    var pNew, pTemp *Gia_Man_t;
    // inlining of macro ABC_FREE
    p.pReprs = nil; p.pReprs = nil;
    // inlining of macro ABC_FREE
    p.pNexts = nil; p.pNexts = nil;
    Gia_ManSeqEquivDivide( p, pPars );
    //Gia_ManPrintEquivs( p );
    pNew = Gia_ManCorrReduce( p );
    pNew = Gia_ManSeqCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicate AIG by creating a cut between logic fed by PIs]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManHighLightFlopLogic(p *Gia_Man_t) {
    var pObj *Gia_Obj_t; var i int32;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.fMark0 = 0;
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        pObj.fMark0 = 1;
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 | Gia_ObjFanin1(pObj).fMark0; }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0;
    }
}
func Gia_ManDupReplaceCut(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;  var i int32;
    var pObj, pFanin *Gia_Obj_t;
    Gia_ManHighLightFlopLogic( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create PIs for nodes pointed to from above the cut
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !pObj.fMark0 {
            continue; }
        pFanin = Gia_ObjFanin0(pObj);
        if !pFanin.fMark0 && !~pFanin.Value {
            pFanin.Value = Gia_ManAppendCi(pNew); }
        pFanin = Gia_ObjFanin1(pObj);
        if !pFanin.fMark0 && !~pFanin.Value {
            pFanin.Value = Gia_ManAppendCi(pNew); }
    }
    }
    // create flop outputs
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // create internal nodes
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupOrderDfs_rec( pNew, p, pObj );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    Gia_ManCleanMark0( p );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


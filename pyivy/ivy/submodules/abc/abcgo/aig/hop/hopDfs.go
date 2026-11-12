package hop

import "fmt"

/**CFile****************************************************************

  FileName    [hopDfs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Minimalistic And-Inverter Graph package.]

  Synopsis    [DFS traversal procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: hopDfs.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "hop.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Collects internal nodes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManDfs_rec(pObj *Hop_Obj_t, vNodes *Vec_Ptr_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return; }
    Hop_ManDfs_rec( Hop_ObjFanin0(pObj), vNodes );
    Hop_ManDfs_rec( Hop_ObjFanin1(pObj), vNodes );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA(pObj);
    Vec_PtrPush( vNodes, pObj );
}

/**Function*************************************************************

  Synopsis    [Collects internal nodes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManDfs(p *Hop_Man_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Hop_Obj_t;
    var i int32;
    vNodes = Vec_PtrAlloc( Hop_ManNodeNum(p) );
    // inlining of macro Hop_ManForEachNode
    for  i = 0; i < p.nTableSize; i++  {                                               if ((pObj) = p.pTable[i]) == nil {} else {
        Hop_ManDfs_rec( pObj, vNodes ); }
    }
    // inlining of macro Hop_ManForEachNode
    for  i = 0; i < p.nTableSize; i++  {                                               if ((pObj) = p.pTable[i]) == nil {} else {
        Hop_ObjClearMarkA(pObj); }
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Collects internal nodes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManDfsNode(p *Hop_Man_t, pNode *Hop_Obj_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Hop_Obj_t;
    var i int32;
    assert(!Hop_IsComplement(pNode), "!Hop_IsComplement(pNode)");
    vNodes = Vec_PtrAlloc( 16 );
    Hop_ManDfs_rec( pNode, vNodes );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Hop_ObjClearMarkA(pObj);
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Computes the max number of levels in the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManCountLevels(p *Hop_Man_t) int32 {
    var vNodes *Vec_Ptr_t;
    var pObj *Hop_Obj_t;
    var i, LevelsMax, Level0, Level1 int32;
    // initialize the levels
    Hop_ManConst1(p).pData = nil;
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        pObj.pData = nil;
    }
    // compute levels in a DFS order
    vNodes = Hop_ManDfs( p );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Level0 = int32((ABC_PTRUINT_T))Hop_ObjFanin0(pObj).pData;
        Level1 = int32((ABC_PTRUINT_T))Hop_ObjFanin1(pObj).pData;
        pObj.pData = (void *)(ABC_PTRUINT_T)(1 + Hop_ObjIsExor(pObj) + Abc_MaxInt(Level0, Level1));
    }
    Vec_PtrFree( vNodes );
    // get levels of the POs
    LevelsMax = 0;
    // inlining of macro Hop_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        LevelsMax = Abc_MaxInt( LevelsMax, int32((ABC_PTRUINT_T))Hop_ObjFanin0(pObj).pData );
    }
    return LevelsMax;
}

/**Function*************************************************************

  Synopsis    [Creates correct reference counters at each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManCreateRefs(p *Hop_Man_t) {
    var pObj *Hop_Obj_t;
    var i int32;
    if p.fRefCount {
        return; }
    p.fRefCount = 1;
    // clear refs
    Hop_ObjClearRef( Hop_ManConst1(p) );
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        Hop_ObjClearRef( pObj );
    }
    // inlining of macro Hop_ManForEachNode
    for  i = 0; i < p.nTableSize; i++  {                                               if ((pObj) = p.pTable[i]) == nil {} else {
        Hop_ObjClearRef( pObj ); }
    }
    // inlining of macro Hop_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        Hop_ObjClearRef( pObj );
    }
    // set refs
    // inlining of macro Hop_ManForEachNode
    for  i = 0; i < p.nTableSize; i++  {                                               if ((pObj) = p.pTable[i]) == nil {} else {
        Hop_ObjRef( Hop_ObjFanin0(pObj) );
        Hop_ObjRef( Hop_ObjFanin1(pObj) );
    }
    }
    // inlining of macro Hop_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        Hop_ObjRef( Hop_ObjFanin0(pObj) );
    }
}

/**Function*************************************************************

  Synopsis    [Counts the number of AIG nodes rooted at this cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ConeMark_rec(pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return; }
    Hop_ConeMark_rec( Hop_ObjFanin0(pObj) );
    Hop_ConeMark_rec( Hop_ObjFanin1(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
}

/**Function*************************************************************

  Synopsis    [Counts the number of AIG nodes rooted at this cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ConeCleanAndMark_rec(pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return; }
    Hop_ConeCleanAndMark_rec( Hop_ObjFanin0(pObj) );
    Hop_ConeCleanAndMark_rec( Hop_ObjFanin1(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
    pObj.pData = nil;
}

/**Function*************************************************************

  Synopsis    [Counts the number of AIG nodes rooted at this cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ConeCountAndMark_rec(pObj *Hop_Obj_t) int32 {
    var Counter int32;
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return 0; }
    Counter = 1 + Hop_ConeCountAndMark_rec( Hop_ObjFanin0(pObj) ) + 
        Hop_ConeCountAndMark_rec( Hop_ObjFanin1(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of AIG nodes rooted at this cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ConeUnmark_rec(pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || !Hop_ObjIsMarkA(pObj) {
        return; }
    Hop_ConeUnmark_rec( Hop_ObjFanin0(pObj) ); 
    Hop_ConeUnmark_rec( Hop_ObjFanin1(pObj) );
    assert(Hop_ObjIsMarkA(pObj), "Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjClearMarkA( pObj );
}

/**Function*************************************************************

  Synopsis    [Counts the number of AIG nodes rooted at this cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_DagSize(pObj *Hop_Obj_t) int32 {
    var Counter int32;
    Counter = Hop_ConeCountAndMark_rec( Hop_Regular(pObj) );
    Hop_ConeUnmark_rec( Hop_Regular(pObj) );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts how many fanout the given node has.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ObjFanoutCount_rec(pObj *Hop_Obj_t, pPivot *Hop_Obj_t) int32 {
    var Counter int32;
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return int32((pObj == pPivot)); }
    Counter = Hop_ObjFanoutCount_rec( Hop_ObjFanin0(pObj), pPivot ) + 
              Hop_ObjFanoutCount_rec( Hop_ObjFanin1(pObj), pPivot );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
    return Counter;
}
func Hop_ObjFanoutCount(pObj *Hop_Obj_t, pPivot *Hop_Obj_t) int32 {
    var Counter int32;
    assert(!Hop_IsComplement(pPivot), "!Hop_IsComplement(pPivot)");
    Counter = Hop_ObjFanoutCount_rec( Hop_Regular(pObj), pPivot );
    Hop_ConeUnmark_rec( Hop_Regular(pObj) );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Transfers the AIG from one manager into another.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_Transfer_rec(pDest *Hop_Man_t, pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return; }
    Hop_Transfer_rec( pDest, Hop_ObjFanin0(pObj) ); 
    Hop_Transfer_rec( pDest, Hop_ObjFanin1(pObj) );
    pObj.pData = Hop_And( pDest, Hop_ObjChild0Copy(pObj), Hop_ObjChild1Copy(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
}

/**Function*************************************************************

  Synopsis    [Transfers the AIG from one manager into another.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_Transfer(pSour *Hop_Man_t, pDest *Hop_Man_t, pRoot *Hop_Obj_t, nVars int32) *Hop_Obj_t {
    var pObj *Hop_Obj_t;
    var i int32;
    // solve simple cases
    if pSour == pDest {
        return pRoot; }
    if Hop_ObjIsConst1( Hop_Regular(pRoot) ) {
        return Hop_NotCond( Hop_ManConst1(pDest), Hop_IsComplement(pRoot) ); }
    // set the PI mapping
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pSour.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(pSour.vPis, i)), 1); i++  {
        if i == nVars {
           break; }
        pObj.pData = Hop_IthVar(pDest, i);
    }
    // transfer and set markings
    Hop_Transfer_rec( pDest, Hop_Regular(pRoot) );
    // clear the markings
    Hop_ConeUnmark_rec( Hop_Regular(pRoot) );
    return Hop_NotCond( (Hop_Obj_t *)Hop_Regular(pRoot).pData, Hop_IsComplement(pRoot) );
}

/**Function*************************************************************

  Synopsis    [Composes the AIG (pRoot) with the function (pFunc) using PI var (iVar).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_Compose_rec(p *Hop_Man_t, pObj *Hop_Obj_t, pFunc *Hop_Obj_t, pVar *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if Hop_ObjIsMarkA(pObj) {
        return; }
    if Hop_ObjIsConst1(pObj) || Hop_ObjIsPi(pObj) {
        if pObj == pVar {
            pObj.pData = pFunc;
        } else {
            pObj.pData = pObj;
        }
        return;
    }
    Hop_Compose_rec( p, Hop_ObjFanin0(pObj), pFunc, pVar ); 
    Hop_Compose_rec( p, Hop_ObjFanin1(pObj), pFunc, pVar );
    pObj.pData = Hop_And( p, Hop_ObjChild0Copy(pObj), Hop_ObjChild1Copy(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
}

/**Function*************************************************************

  Synopsis    [Composes the AIG (pRoot) with the function (pFunc) using PI var (iVar).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_Compose(p *Hop_Man_t, pRoot *Hop_Obj_t, pFunc *Hop_Obj_t, iVar int32) *Hop_Obj_t {
    // quit if the PI variable is not defined
    if iVar >= Hop_ManPiNum(p) {
        fmt.Printf("Hop_Compose(): The PI variable %d is not defined.\n", iVar);
        return nil;
    }
    // recursively perform composition
    Hop_Compose_rec( p, Hop_Regular(pRoot), pFunc, Hop_ManPi(p, iVar) );
    // clear the markings
    Hop_ConeUnmark_rec( Hop_Regular(pRoot) );
    return Hop_NotCond( (Hop_Obj_t *)Hop_Regular(pRoot).pData, Hop_IsComplement(pRoot) );
}

/**Function*************************************************************

  Synopsis    [Complements the AIG (pRoot) with the function (pFunc) using PI var (iVar).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_Complement_rec(p *Hop_Man_t, pObj *Hop_Obj_t, pVar *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if Hop_ObjIsMarkA(pObj) {
        return; }
    if Hop_ObjIsConst1(pObj) || Hop_ObjIsPi(pObj) {
        if pObj == pVar {
            pObj.pData = Hop_Not(pObj);
        } else {
            pObj.pData = pObj;
        }
        return;
    }
    Hop_Complement_rec( p, Hop_ObjFanin0(pObj), pVar ); 
    Hop_Complement_rec( p, Hop_ObjFanin1(pObj), pVar );
    pObj.pData = Hop_And( p, Hop_ObjChild0Copy(pObj), Hop_ObjChild1Copy(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
}

/**Function*************************************************************

  Synopsis    [Complements the AIG (pRoot) with the function (pFunc) using PI var (iVar).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_Complement(p *Hop_Man_t, pRoot *Hop_Obj_t, iVar int32) *Hop_Obj_t {
    // quit if the PI variable is not defined
    if iVar >= Hop_ManPiNum(p) {
        fmt.Printf("Hop_Complement(): The PI variable %d is not defined.\n", iVar);
        return nil;
    }
    // recursively perform composition
    Hop_Complement_rec( p, Hop_Regular(pRoot), Hop_ManPi(p, iVar) );
    // clear the markings
    Hop_ConeUnmark_rec( Hop_Regular(pRoot) );
    return Hop_NotCond( (Hop_Obj_t *)Hop_Regular(pRoot).pData, Hop_IsComplement(pRoot) );
}

/**Function*************************************************************

  Synopsis    [Remaps the AIG (pRoot) to have the given support (uSupp).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_Remap_rec(p *Hop_Man_t, pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return; }
    Hop_Remap_rec( p, Hop_ObjFanin0(pObj) ); 
    Hop_Remap_rec( p, Hop_ObjFanin1(pObj) );
    pObj.pData = Hop_And( p, Hop_ObjChild0Copy(pObj), Hop_ObjChild1Copy(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
}

/**Function*************************************************************

  Synopsis    [Remaps the AIG (pRoot) to have the given support (uSupp).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_Remap(p *Hop_Man_t, pRoot *Hop_Obj_t, uSupp uint32, nVars int32) *Hop_Obj_t {
    var pObj *Hop_Obj_t;
    var i, k int32;
    // quit if the PI variable is not defined
    if nVars > Hop_ManPiNum(p) {
        fmt.Printf("Hop_Remap(): The number of variables (%d) is more than the manager size (%d).\n", nVars, Hop_ManPiNum(p));
        return nil;
    }
    // return if constant
    if Hop_ObjIsConst1( Hop_Regular(pRoot) ) {
        return pRoot; }
    if uSupp == 0 {
        return Hop_NotCond( Hop_ManConst0(p), Hop_ObjPhaseCompl(pRoot) ); }
    // set the PI mapping
    k = 0;
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        if i == nVars {
           break; }
        if uSupp & (1 << i) {
            pObj.pData = Hop_IthVar(p, k++);
        } else {
            pObj.pData = Hop_ManConst0(p); }
    }
    assert(k > 0 && k < nVars, "k > 0 && k < nVars");
    // recursively perform composition
    Hop_Remap_rec( p, Hop_Regular(pRoot) );
    // clear the markings
    Hop_ConeUnmark_rec( Hop_Regular(pRoot) );
    return Hop_NotCond( (Hop_Obj_t *)Hop_Regular(pRoot).pData, Hop_IsComplement(pRoot) );
}

/**Function*************************************************************

  Synopsis    [Permute the AIG according to the given permutation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_Permute(p *Hop_Man_t, pRoot *Hop_Obj_t, nRootVars int32, pPermute *int32) *Hop_Obj_t {
    var pObj *Hop_Obj_t;
    var i int32;
    // return if constant
    if Hop_ObjIsConst1( Hop_Regular(pRoot) ) {
        return pRoot; }
    // create mapping
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        if i == nRootVars {
            break; }
        assert(pPermute[i] >= 0 && pPermute[i] < Hop_ManPiNum(p), "pPermute[i] >= 0 && pPermute[i] < Hop_ManPiNum(p)");
        pObj.pData = Hop_IthVar( p, pPermute[i] );
    }
    // recursively perform composition
    Hop_Remap_rec( p, Hop_Regular(pRoot) );
    // clear the markings
    Hop_ConeUnmark_rec( Hop_Regular(pRoot) );
    return Hop_NotCond( (Hop_Obj_t *)Hop_Regular(pRoot).pData, Hop_IsComplement(pRoot) );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


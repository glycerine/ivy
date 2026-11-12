package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigFactor.c]

  SystemName  []

  PackageName []

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation []

  Date        [Ver. 1.0. Started - April 17, 2009.]

  Revision    [$Id: aigFactor.c,v 1.00 2009/04/17 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Detects multi-input AND gate rooted at this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFindImplications_rec(pObj *Aig_Obj_t, vImplics *Vec_Ptr_t) {
    if Aig_IsComplement(pObj) || Aig_ObjIsCi(pObj) {
        Vec_PtrPushUnique( vImplics, pObj );
        return;
    }
    Aig_ManFindImplications_rec( Aig_ObjChild0(pObj), vImplics );
    Aig_ManFindImplications_rec( Aig_ObjChild1(pObj), vImplics );
}

/**Function*************************************************************

  Synopsis    [Returns the nodes whose values are implied by pNode.]

  Description [Attention!  Both pNode and results can be complemented!
  Also important: Currently, this procedure only does backward propagation. 
  In general, it may find more implications if forward propagation is enabled.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFindImplications(p *Aig_Man_t, pNode *Aig_Obj_t) *Vec_Ptr_t {
    var vImplics *Vec_Ptr_t;
    vImplics = Vec_PtrAlloc( 100 );
    Aig_ManFindImplications_rec( pNode, vImplics );
    return vImplics;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the cone of the node overlaps with the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFindConeOverlap_rec(p *Aig_Man_t, pNode *Aig_Obj_t) int32 {
    if Aig_ObjIsTravIdPrevious( p, pNode ) {
        return 1; }
    if Aig_ObjIsTravIdCurrent( p, pNode ) {
        return 0; }
    Aig_ObjSetTravIdCurrent( p, pNode );
    if Aig_ObjIsCi(pNode) {
        return 0; }
    if Aig_ManFindConeOverlap_rec( p, Aig_ObjFanin0(pNode) ) != 0 {
        return 1; }
    if Aig_ManFindConeOverlap_rec( p, Aig_ObjFanin1(pNode) ) != 0 {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the cone of the node overlaps with the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFindConeOverlap(p *Aig_Man_t, vImplics *Vec_Ptr_t, pNode *Aig_Obj_t) int32 {
    var pTemp *Aig_Obj_t;
    var i int32;
    assert(!Aig_IsComplement(pNode), "!Aig_IsComplement(pNode)");
    assert(!Aig_ObjIsConst1(pNode), "!Aig_ObjIsConst1(pNode)");
    Aig_ManIncrementTravId( p );
    Vec_PtrForEachEntry( Aig_Obj_t *, vImplics, pTemp, i )
        Aig_ObjSetTravIdCurrent( p, Aig_Regular(pTemp) );
    Aig_ManIncrementTravId( p );
    return Aig_ManFindConeOverlap_rec( p, pNode );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the cone of the node overlaps with the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDeriveNewCone_rec(p *Aig_Man_t, pNode *Aig_Obj_t) *Aig_Obj_t {
    if Aig_ObjIsTravIdCurrent( p, pNode ) {
        return (Aig_Obj_t *)pNode.pData; }
    Aig_ObjSetTravIdCurrent( p, pNode );
    if Aig_ObjIsCi(pNode) {
        return (Aig_Obj_t *)(pNode.pData = pNode); }
    Aig_ManDeriveNewCone_rec( p, Aig_ObjFanin0(pNode) );
    Aig_ManDeriveNewCone_rec( p, Aig_ObjFanin1(pNode) );
    return (Aig_Obj_t *)(pNode.pData = Aig_And( p, Aig_ObjChild0Copy(pNode), Aig_ObjChild1Copy(pNode) ));
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the cone of the node overlaps with the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManDeriveNewCone(p *Aig_Man_t, vImplics *Vec_Ptr_t, pNode *Aig_Obj_t) *Aig_Obj_t {
    var pTemp *Aig_Obj_t;
    var i int32;
    assert(!Aig_IsComplement(pNode), "!Aig_IsComplement(pNode)");
    assert(!Aig_ObjIsConst1(pNode), "!Aig_ObjIsConst1(pNode)");
    Aig_ManIncrementTravId( p );
    Vec_PtrForEachEntry( Aig_Obj_t *, vImplics, pTemp, i )
    {
        Aig_ObjSetTravIdCurrent( p, Aig_Regular(pTemp) );
        Aig_Regular(pTemp).pData = Aig_NotCond( Aig_ManConst1(p), Aig_IsComplement(pTemp) );
    }
    return Aig_ManDeriveNewCone_rec( p, pNode );
}

/**Function*************************************************************

  Synopsis    [Returns algebraic factoring of B in terms of A.]

  Description [Returns internal node C (an AND gate) that is equal to B 
  under assignment A = 'Value', or NULL if there is no such node C. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFactorAlgebraic_int(p *Aig_Man_t, pPoA *Aig_Obj_t, pPoB *Aig_Obj_t, Value int32) *Aig_Obj_t {
    var pNodeA, pNodeC *Aig_Obj_t;
    var vImplics *Vec_Ptr_t;
    var RetValue int32;
    if Aig_ObjIsConst1(Aig_ObjFanin0(pPoA)) || Aig_ObjIsConst1(Aig_ObjFanin0(pPoB)) {
        return nil; }
    if Aig_ObjIsCi(Aig_ObjFanin0(pPoB)) {
        return nil; }
    // get the internal node representing function of A under assignment A = 'Value'
    pNodeA = Aig_ObjChild0( pPoA );
    pNodeA = Aig_NotCond( pNodeA, Value==0 );
    // find implications of this signal (nodes whose value is fixed under assignment A = 'Value')
    vImplics = Aig_ManFindImplications( p, pNodeA );
    // check if the TFI cone of B overlaps with the implied nodes
    RetValue = Aig_ManFindConeOverlap( p, vImplics, Aig_ObjFanin0(pPoB) );
    if RetValue == 0 { // no overlap
    
        Vec_PtrFree( vImplics );
        return nil;
    }
    // there is overlap - derive node representing value of B under assignment A = 'Value'
    pNodeC = Aig_ManDeriveNewCone( p, vImplics, Aig_ObjFanin0(pPoB) );
    pNodeC = Aig_NotCond( pNodeC, Aig_ObjFaninC0(pPoB) );
    Vec_PtrFree( vImplics );
    return pNodeC;
}

/**Function*************************************************************

  Synopsis    [Returns algebraic factoring of B in terms of A.]

  Description [Returns internal node C (an AND gate) that is equal to B 
  under assignment A = 'Value', or NULL if there is no such node C. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFactorAlgebraic(p *Aig_Man_t, iPoA int32, iPoB int32, Value int32) *Aig_Obj_t {
    assert(iPoA >= 0 && iPoA < Aig_ManCoNum(p), "iPoA >= 0 && iPoA < Aig_ManCoNum(p)");
    assert(iPoB >= 0 && iPoB < Aig_ManCoNum(p), "iPoB >= 0 && iPoB < Aig_ManCoNum(p)");
    assert(Value == 0 || Value == 1, "Value == 0 || Value == 1");
    return Aig_ManFactorAlgebraic_int( p, Aig_ManCo(p, iPoA), Aig_ManCo(p, iPoB), Value ); 
}

/**Function*************************************************************

  Synopsis    [Testing procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManFactorAlgebraicTest(p *Aig_Man_t) {
    var iPoA int32;
    var iPoB int32   = 1;
    var Value int32;
    var pRes *Aig_Obj_t;
//    Aig_Obj_t * pObj;
//    int i;
    pRes = Aig_ManFactorAlgebraic( p, iPoA, iPoB, Value );
    Aig_ManShow( p, 0, nil );
    Aig_ObjPrint( p, pRes );
    fmt.Printf("\n");
/*
    printf( "Results:\n" );
    Aig_ManForEachObj( p, pObj, i )
    {
        printf( "Object = %d.\n", i );
        Aig_ObjPrint( p, pObj );
        printf( "\n" );
        Aig_ObjPrint( p, pObj->pData );
        printf( "\n" );
    }
*/
}



/**Function*************************************************************

  Synopsis    [Determines what support variables can be cofactored.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_SuppMinPerform(p *Aig_Man_t, vOrGate *Vec_Ptr_t, vNodes *Vec_Ptr_t, vSupp *Vec_Ptr_t) *Vec_Ptr_t {
    var pObj *Aig_Obj_t;
    var vTrSupp, vTrNode, vCofs *Vec_Ptr_t;
    var uFunc *uint32 = nil;
    var uCare *uint32;
    var uFunc0 *uint32;
    var uFunc1 *uint32;
    var uCof0, uCof1, uCare0, uCare1 *uint32;
    var i int32;
    var nWords int32 = Abc_TruthWordNum( Vec_PtrSize(vSupp) );
    // assign support nodes
    vTrSupp = Vec_PtrAllocTruthTables( Vec_PtrSize(vSupp) );
    Vec_PtrForEachEntry( Aig_Obj_t *, vSupp, pObj, i )
        pObj.pData = Vec_PtrEntry( vTrSupp, i );
    // compute internal nodes
    vTrNode = Vec_PtrAllocSimInfo( Vec_PtrSize(vNodes) + 5, nWords );
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
    {
        pObj.pData = uFunc = (unsigned *)Vec_PtrEntry( vTrNode, i );
        uFunc0 = (unsigned *)Aig_ObjFanin0(pObj).pData;
        uFunc1 = (unsigned *)Aig_ObjFanin1(pObj).pData;
        Kit_TruthAndPhase( uFunc, uFunc0, uFunc1, Vec_PtrSize(vSupp), Aig_ObjFaninC0(pObj), Aig_ObjFaninC1(pObj) );
    }
    // uFunc contains the result of computation
    // compute care set
    uCare = (unsigned *)Vec_PtrEntry( vTrNode, Vec_PtrSize(vNodes) );
    Kit_TruthClear( uCare, Vec_PtrSize(vSupp) );
    Vec_PtrForEachEntry( Aig_Obj_t *, vOrGate, pObj, i )
        Kit_TruthOrPhase( uCare, uCare, (unsigned *)Aig_Regular(pObj).pData, Vec_PtrSize(vSupp), 0, Aig_IsComplement(pObj) );
    // try cofactoring each variable in both polarities
    vCofs = Vec_PtrAlloc( 10 );
    uCof0  = (unsigned *)Vec_PtrEntry( vTrNode, Vec_PtrSize(vNodes)+1 );
    uCof1  = (unsigned *)Vec_PtrEntry( vTrNode, Vec_PtrSize(vNodes)+2 );
    uCare0 = (unsigned *)Vec_PtrEntry( vTrNode, Vec_PtrSize(vNodes)+3 );
    uCare1 = (unsigned *)Vec_PtrEntry( vTrNode, Vec_PtrSize(vNodes)+4 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vSupp, pObj, i )
    {
        Kit_TruthCofactor0New( uCof0, uFunc, Vec_PtrSize(vSupp), i );
        Kit_TruthCofactor1New( uCof1, uFunc, Vec_PtrSize(vSupp), i );
        Kit_TruthCofactor0New( uCare0, uCare, Vec_PtrSize(vSupp), i );
        Kit_TruthCofactor1New( uCare1, uCare, Vec_PtrSize(vSupp), i );
        if Kit_TruthIsEqualWithCare( uCof0, uCof1, uCare1, Vec_PtrSize(vSupp) ) {
            Vec_PtrPush( vCofs, Aig_Not(pObj) );
        } else if Kit_TruthIsEqualWithCare( uCof0, uCof1, uCare0, Vec_PtrSize(vSupp) ) {
            Vec_PtrPush( vCofs, pObj ); }
    }
    Vec_PtrFree( vTrNode );
    Vec_PtrFree( vTrSupp );
    return vCofs;
}



/**Function*************************************************************

  Synopsis    [Returns the new node after cofactoring.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_SuppMinReconstruct(p *Aig_Man_t, vCofs *Vec_Ptr_t, vNodes *Vec_Ptr_t, vSupp *Vec_Ptr_t) *Aig_Obj_t {
    var pObj *Aig_Obj_t  = nil;
    var i int32;
    // set the value of the support variables
    Vec_PtrForEachEntry( Aig_Obj_t *, vSupp, pObj, i )
        assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    Vec_PtrForEachEntry( Aig_Obj_t *, vSupp, pObj, i )
        pObj.pData = pObj;
    // set the value of the cofactoring variables
    Vec_PtrForEachEntry( Aig_Obj_t *, vCofs, pObj, i )
        Aig_Regular(pObj).pData = Aig_NotCond( Aig_ManConst1(p), Aig_IsComplement(pObj) );
    // reconstruct the node
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
        pObj.pData = Aig_And( p, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
    return (Aig_Obj_t *)pObj.pData;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if all nodes of vOrGate are in vSupp.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_SuppMinGateIsInSupport(p *Aig_Man_t, vOrGate *Vec_Ptr_t, vSupp *Vec_Ptr_t) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    Aig_ManIncrementTravId( p );
    Vec_PtrForEachEntry( Aig_Obj_t *, vSupp, pObj, i )
        Aig_ObjSetTravIdCurrent( p, pObj );
    Vec_PtrForEachEntry( Aig_Obj_t *, vOrGate, pObj, i )
        if ( !Aig_ObjIsTravIdCurrent( p, Aig_Regular(pObj) ) )
            return 0;
    return 1;
}


/**Function*************************************************************

  Synopsis    [Collects fanins of the marked nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_SuppMinCollectSupport(p *Aig_Man_t, vNodes *Vec_Ptr_t) *Vec_Ptr_t {
    var vSupp *Vec_Ptr_t;
    var pObj, pFanin *Aig_Obj_t;
    var i int32;
    vSupp = Vec_PtrAlloc( 4 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
    {
        assert(Aig_ObjIsTravIdCurrent(p, pObj), "Aig_ObjIsTravIdCurrent(p, pObj)");
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        pFanin = Aig_ObjFanin0( pObj );
        if !Aig_ObjIsTravIdCurrent(p, pFanin) {
            Aig_ObjSetTravIdCurrent( p, pFanin );
            Vec_PtrPush( vSupp, pFanin );
        }
        pFanin = Aig_ObjFanin1( pObj );
        if !Aig_ObjIsTravIdCurrent(p, pFanin) {
            Aig_ObjSetTravIdCurrent( p, pFanin );
            Vec_PtrPush( vSupp, pFanin );
        }
    }
    return vSupp;
}

/**Function*************************************************************

  Synopsis    [Marks the nodes in the cone with current trav ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_SuppMinCollectCone_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vNodes *Vec_Ptr_t) {
    if Aig_ObjIsTravIdCurrent( p, pObj ) { // visited
        return; }
    if !Aig_ObjIsTravIdPrevious( p, pObj ) { // not visited, but outside
        return; }
    assert(Aig_ObjIsTravIdPrevious(p, pObj), "Aig_ObjIsTravIdPrevious(p, pObj)"); // not visited, inside
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Aig_ObjSetTravIdCurrent( p, pObj );
    Aig_SuppMinCollectCone_rec( p, Aig_ObjFanin0(pObj), vNodes );
    Aig_SuppMinCollectCone_rec( p, Aig_ObjFanin1(pObj), vNodes );
    Vec_PtrPush( vNodes, pObj );
}

/**Function*************************************************************

  Synopsis    [Collects nodes with the current trav ID rooted in the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_SuppMinCollectCone(p *Aig_Man_t, pRoot *Aig_Obj_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    assert(!Aig_IsComplement(pRoot), "!Aig_IsComplement(pRoot)"); 
//    assert( Aig_ObjIsTravIdCurrent( p, pRoot ) );
    vNodes = Vec_PtrAlloc( 4 );
    Aig_ManIncrementTravId( p );
    Aig_SuppMinCollectCone_rec( p, Aig_Regular(pRoot), vNodes );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Marks the nodes in the cone with current trav ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_SuppMinHighlightCone_rec(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    var RetValue int32;
    if Aig_ObjIsTravIdCurrent( p, pObj ) { // visited, marks there
        return 1; }
    if Aig_ObjIsTravIdPrevious( p, pObj ) { // visited, no marks there
        return 0; }
    Aig_ObjSetTravIdPrevious( p, pObj );
    if Aig_ObjIsCi(pObj) {
        return 0; }
    RetValue = Aig_SuppMinHighlightCone_rec( p, Aig_ObjFanin0(pObj) ) |
               Aig_SuppMinHighlightCone_rec( p, Aig_ObjFanin1(pObj) );
//    printf( "%d %d\n", Aig_ObjId(pObj), RetValue );
    if RetValue != 0 {
        Aig_ObjSetTravIdCurrent( p, pObj ); }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Marks the nodes in the cone with current trav ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_SuppMinHighlightCone(p *Aig_Man_t, pRoot *Aig_Obj_t, vOrGate *Vec_Ptr_t) int32 {
    var pLeaf *Aig_Obj_t;
    var i, RetValue int32;
    assert(!Aig_IsComplement(pRoot), "!Aig_IsComplement(pRoot)");
    Aig_ManIncrementTravId( p );
    Aig_ManIncrementTravId( p );
    Vec_PtrForEachEntry( Aig_Obj_t *, vOrGate, pLeaf, i )
        Aig_ObjSetTravIdCurrent( p, Aig_Regular(pLeaf) );
    RetValue = Aig_SuppMinHighlightCone_rec( p, pRoot );
    Vec_PtrForEachEntry( Aig_Obj_t *, vOrGate, pLeaf, i )
        Aig_ObjSetTravIdPrevious( p, Aig_Regular(pLeaf) );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_SuppMinCollectSuper_rec(pObj *Aig_Obj_t, vSuper *Vec_Ptr_t) {
    // if the new node is complemented or a PI, another gate begins
    if Aig_IsComplement(pObj) || Aig_ObjIsCi(pObj) { // || (Aig_ObjRefs(pObj) > 1) )
    
        Vec_PtrPushUnique( vSuper, Aig_Not(pObj) );
        return;
    }
    // go through the branches
    Aig_SuppMinCollectSuper_rec( Aig_ObjChild0(pObj), vSuper );
    Aig_SuppMinCollectSuper_rec( Aig_ObjChild1(pObj), vSuper );
}

/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_SuppMinCollectSuper(pObj *Aig_Obj_t) *Vec_Ptr_t {
    var vSuper *Vec_Ptr_t;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(!Aig_ObjIsCi(pObj), "!Aig_ObjIsCi(pObj)");
    vSuper = Vec_PtrAlloc( 4 );
    Aig_SuppMinCollectSuper_rec( Aig_ObjChild0(pObj), vSuper );
    Aig_SuppMinCollectSuper_rec( Aig_ObjChild1(pObj), vSuper );
    return vSuper;
}

/**Function*************************************************************

  Synopsis    [Returns the result of support minimization.]

  Description [Returns internal AIG node that is equal to pFunc under 
  assignment pCond == 1, or NULL if there is no such node. status is 
  -1 if condition is not OR; 
  -2 if cone is too large or no cone; 
  -3 if no support reduction is possible.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSupportMinimization(p *Aig_Man_t, pCond *Aig_Obj_t, pFunc *Aig_Obj_t, pStatus *int32) *Aig_Obj_t {
    var nSuppMax int32  = 16;
    var vOrGate, vNodes, vSupp, vCofs *Vec_Ptr_t;
    var pResult *Aig_Obj_t;
    var RetValue int32;
    *pStatus = 0;
    // if pCond is not OR
    if !Aig_IsComplement(pCond) || Aig_ObjIsCi(Aig_Regular(pCond)) || Aig_ObjIsConst1(Aig_Regular(pCond)) {
        *pStatus = -1;
        return nil;
    }
    // if pFunc is not a node
    if !Aig_ObjIsNode(Aig_Regular(pFunc)) {
        *pStatus = -2;
        return nil;
    }
    // collect the multi-input OR gate rooted in the condition
    vOrGate = Aig_SuppMinCollectSuper( Aig_Regular(pCond) );
    if Vec_PtrSize(vOrGate) > nSuppMax {
        Vec_PtrFree( vOrGate );
        *pStatus = -2;
        return nil;
    }
    // highlight the cone limited by these gates
    RetValue = Aig_SuppMinHighlightCone( p, Aig_Regular(pFunc), vOrGate );
    if RetValue == 0 { // no overlap
    
        Vec_PtrFree( vOrGate );
        *pStatus = -2;
        return nil;
    }
    // collect the cone rooted in pFunc limited by vOrGate
    vNodes = Aig_SuppMinCollectCone( p, Aig_Regular(pFunc) );
    // collect the support nodes reachable from the cone
    vSupp = Aig_SuppMinCollectSupport( p, vNodes );
    if Vec_PtrSize(vSupp) > nSuppMax {
        Vec_PtrFree( vOrGate );
        Vec_PtrFree( vNodes );
        Vec_PtrFree( vSupp );
        *pStatus = -2;
        return nil;
    }
    // check if all nodes belonging to OR gate are included in the support
    // (if this is not the case, don't-care minimization is not possible)
    if !(Aig_SuppMinGateIsInSupport( p, vOrGate, vSupp ) != 0) {
        Vec_PtrFree( vOrGate );
        Vec_PtrFree( vNodes );
        Vec_PtrFree( vSupp );
        *pStatus = -3;
        return nil;
    }
    // create truth tables of all nodes and find the maximal number
    // of support varialbles that can be replaced by constants
    vCofs = Aig_SuppMinPerform( p, vOrGate, vNodes, vSupp );
    if Vec_PtrSize(vCofs) == 0 {
        Vec_PtrFree( vCofs );
        Vec_PtrFree( vOrGate );
        Vec_PtrFree( vNodes );
        Vec_PtrFree( vSupp );
        *pStatus = -3;
        return nil;
    }
    // reconstruct the cone
    pResult = Aig_SuppMinReconstruct( p, vCofs, vNodes, vSupp );
    pResult = Aig_NotCond( pResult, Aig_IsComplement(pFunc) );
    Vec_PtrFree( vCofs );
    Vec_PtrFree( vOrGate );
    Vec_PtrFree( vNodes );
    Vec_PtrFree( vSupp );
    return pResult;
}
/**Function*************************************************************

  Synopsis    [Testing procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManSupportMinimizationTest() {
    var p *Aig_Man_t;
    var pFunc, pCond, pRes *Aig_Obj_t;
    var i, Status int32;
    p = Aig_ManStart( 100 );
    for  i = 0; i < 5; i++  {
        Aig_IthVar(p,i);
    }
    pFunc = Aig_Mux( p, Aig_IthVar(p,3), Aig_IthVar(p,1), Aig_IthVar(p,0) );
    pFunc = Aig_Mux( p, Aig_IthVar(p,4), Aig_IthVar(p,2), pFunc );
    pCond = Aig_Or( p, Aig_IthVar(p,3), Aig_IthVar(p,4) );
    pRes  = Aig_ManSupportMinimization( p, pCond, pFunc, &Status );
    assert(Status == 0, "Status == 0");

    Aig_ObjPrint( p, Aig_Regular(pRes) );                fmt.Printf("\n");
    Aig_ObjPrint( p, Aig_ObjFanin0(Aig_Regular(pRes)) ); fmt.Printf("\n");
    Aig_ObjPrint( p, Aig_ObjFanin1(Aig_Regular(pRes)) ); fmt.Printf("\n");

    Aig_ManStop( p );
}
func Aig_ManSupportMinimizationTest2() {
    var p *Aig_Man_t;
    var node09, node10, node11, node12, node13, pRes *Aig_Obj_t;
    var i, Status int32;
    p = Aig_ManStart( 100 );
    for  i = 0; i < 3; i++  {
        Aig_IthVar(p,i);
    }

    node09 = Aig_And( p, Aig_IthVar(p,0), Aig_Not(Aig_IthVar(p,1)) );
    node10 = Aig_And( p, Aig_Not(node09), Aig_Not(Aig_IthVar(p,2)) );
    node11 = Aig_And( p, node10, Aig_IthVar(p,1) );

    node12 = Aig_Or( p, Aig_IthVar(p,1), Aig_IthVar(p,2) );
    node13 = Aig_Or( p, node12, Aig_IthVar(p,0) );

    pRes  = Aig_ManSupportMinimization( p, node13, node11, &Status );
    assert(Status == 0, "Status == 0");

    fmt.Printf("Compl = %d  ", Aig_IsComplement(pRes));
    Aig_ObjPrint( p, Aig_Regular(pRes) );                fmt.Printf("\n");
    if Aig_ObjIsNode(Aig_Regular(pRes)) {
        Aig_ObjPrint( p, Aig_ObjFanin0(Aig_Regular(pRes)) ); fmt.Printf("\n");
        Aig_ObjPrint( p, Aig_ObjFanin1(Aig_Regular(pRes)) ); fmt.Printf("\n");
    }

    Aig_ManStop( p );
}
////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END


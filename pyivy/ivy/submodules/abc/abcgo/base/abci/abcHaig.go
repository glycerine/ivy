package abci

import "fmt"

/**CFile****************************************************************

  FileName    [abcHaig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Implements history AIG for combinational rewriting.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcHaig.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Collects the nodes in the classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHaigCollectMembers(p *Hop_Man_t) *Vec_Ptr_t {
    var vObjs *Vec_Ptr_t;
    var pObj *Hop_Obj_t;
    var i int32;
    vObjs = Vec_PtrAlloc( 4098 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if pObj.pData == nil {
            continue; }
        pObj.pData = Hop_ObjRepr( pObj );
        Vec_PtrPush( vObjs, pObj );
    }
    return vObjs;
}

/**Function*************************************************************

  Synopsis    [Creates classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHaigCreateClasses(vMembers *Vec_Ptr_t) *Vec_Ptr_t {
    var vClasses *Vec_Ptr_t;
    var pObj, pRepr *Hop_Obj_t;
    var i int32;

    // count classes
    vClasses = Vec_PtrAlloc( 4098 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMembers)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(vMembers, i)), 1); i++  {
        pRepr = (Hop_Obj_t *)pObj.pData;
        assert(pRepr.pData == nil, "pRepr.pData == nil");
        if pRepr.fMarkA == 0 { // new
        
            pRepr.fMarkA = 1;
            Vec_PtrPush( vClasses, pRepr );
        }
    }

    // set representatives as representatives
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vClasses)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(vClasses, i)), 1); i++  {
        pObj.fMarkA = 0;
        pObj.pData = pObj;
    }

    // go through the members and update
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMembers)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(vMembers, i)), 1); i++  {
        pRepr = (Hop_Obj_t *)pObj.pData;
        if ((Hop_Obj_t *)pRepr.pData).Id > pObj.Id {
            pRepr.pData = pObj; }
    }

    // change representatives of the class
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMembers)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(vMembers, i)), 1); i++  {
        pRepr = (Hop_Obj_t *)pObj.pData;
        pObj.pData = pRepr.pData;
        assert(((Hop_Obj_t *)pObj.pData).Id <= pObj.Id, "((Hop_Obj_t *)pObj.pData).Id <= pObj.Id");
    }

    // update classes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vClasses)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(vClasses, i)), 1); i++  {
        pRepr = (Hop_Obj_t *)pObj.pData;
        assert(pRepr.pData == pRepr, "pRepr.pData == pRepr");
//        pRepr->pData = NULL;
        Vec_PtrWriteEntry( vClasses, i, pRepr );
        Vec_PtrPush( vMembers, pObj );
    }

    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vMembers)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(vMembers, i)), 1); i++  {
        if pObj.pData == pObj {
            pObj.pData = nil; }
    }

/*
    Vec_PtrForEachEntry( Hop_Obj_t *, vMembers, pObj, i )
    {
        printf( "ObjId = %4d : ", pObj->Id );
        if ( pObj->pData == NULL )
        {
            printf( "NULL" );
        }
        else
        {
            printf( "%4d", ((Hop_Obj_t *)pObj->pData)->Id );
            assert( ((Hop_Obj_t *)pObj->pData)->Id <= pObj->Id );
        }
        printf( "\n" );
    }
*/
    return vClasses;
}

/**Function*************************************************************

  Synopsis    [Counts how many data members have non-trivial fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHaigCountFans(p *Hop_Man_t) int32 {
    var pObj *Hop_Obj_t;
    var i, Counter int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if pObj.pData == nil {
            continue; }
        if Hop_ObjRefs(pObj) > 0 {
            Counter++; }
    }
    fmt.Printf("The number of class members with fanouts = %5d.\n", Counter);
    return Counter;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ObjReprHop(pObj *Hop_Obj_t) *Hop_Obj_t {
    var pRepr *Hop_Obj_t;
    assert(pObj.pNext != nil, "pObj.pNext != nil");
    if pObj.pData == nil {
        return pObj.pNext; }
    pRepr = (Hop_Obj_t *)pObj.pData;
    assert(pRepr.pData == pRepr, "pRepr.pData == pRepr");
    return Hop_NotCond( pRepr.pNext, pObj.fPhase ^ pRepr.fPhase );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ObjChild0Hop(pObj *Hop_Obj_t) *Hop_Obj_t { return Hop_NotCond( Hop_ObjReprHop(Hop_ObjFanin0(pObj)), Hop_ObjFaninC0(pObj) ); }
func Hop_ObjChild1Hop(pObj *Hop_Obj_t) *Hop_Obj_t { return Hop_NotCond( Hop_ObjReprHop(Hop_ObjFanin1(pObj)), Hop_ObjFaninC1(pObj) ); }

/**Function*************************************************************

  Synopsis    [Stops history AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHaigReconstruct(p *Hop_Man_t) *Hop_Man_t { 
    var pNew *Hop_Man_t;
    var pObj *Hop_Obj_t;
    var i, Counter int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        pObj.pNext = nil;
    }
    // start the HOP package
    pNew = Hop_ManStart();
    pNew.vObjs = Vec_PtrAlloc( p.nCreated );
    Vec_PtrPush( pNew.vObjs, Hop_ManConst1(pNew) );
    // map the constant node
    Hop_ManConst1(p).pNext = Hop_ManConst1(pNew);
    // map the CIs
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        pObj.pNext = Hop_ObjCreatePi(pNew);
    }
    // map the internal nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if !Hop_ObjIsNode(pObj) {
            continue; }
        pObj.pNext = Hop_And( pNew, Hop_ObjChild0Hop(pObj), Hop_ObjChild1Hop(pObj) );
//        assert( !Hop_IsComplement(pObj->pNext) );
        if Hop_ManConst1(pNew) == Hop_Regular(pObj.pNext) {
            Counter++; }
        if pObj.pData { // member of the class
            Hop_Regular(pObj.pNext).pData = Hop_Regular(((Hop_Obj_t *)pObj.pData).pNext); }
    }
//    printf( " Counter = %d.\n", Counter );
    // transfer the POs
    // inlining of macro Hop_ManForEachPo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPos)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPos, i)), 1); i++  {
        Hop_ObjCreatePo( pNew, Hop_ObjChild0Hop(pObj) );
    }
    // check the new manager
    if !Hop_ManCheck(pNew) {
        fmt.Printf("Abc_NtkHaigReconstruct: Check for History AIG has failed.\n");
        Hop_ManStop(pNew);
        return nil;
    }
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHaigCheckTfi_rec(pNode *Abc_Obj_t, pOld *Abc_Obj_t) int32 {
    if pNode == nil {
        return 0; }
    if pNode == pOld {
        return 1; }
    // check the trivial cases
    if Abc_ObjIsCi(pNode) {
        return 0; }
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    // if this node is already visited, skip
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return 0; }
    // mark the node as visited
    Abc_NodeSetTravIdCurrent( pNode );
    // check the children
    if Abc_NtkHaigCheckTfi_rec( Abc_ObjFanin0(pNode), pOld ) != 0 {
        return 1; }
    if Abc_NtkHaigCheckTfi_rec( Abc_ObjFanin1(pNode), pOld ) != 0 {
        return 1; }
    // check equivalent nodes
    return Abc_NtkHaigCheckTfi_rec( (Abc_Obj_t *)pNode.pData, pOld );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHaigCheckTfi(pNtk *Abc_Ntk_t, pOld *Abc_Obj_t, pNew *Abc_Obj_t) int32 {
    assert(!Abc_ObjIsComplement(pOld), "!Abc_ObjIsComplement(pOld)");
    assert(!Abc_ObjIsComplement(pNew), "!Abc_ObjIsComplement(pNew)");
    Abc_NtkIncrementTravId(pNtk);
    return Abc_NtkHaigCheckTfi_rec( pNew, pOld );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ObjChild0Next(pObj *Hop_Obj_t) *Abc_Obj_t { return Abc_ObjNotCond( (Abc_Obj_t *)Hop_ObjFanin0(pObj).pNext, Hop_ObjFaninC0(pObj) );  }
func Hop_ObjChild1Next(pObj *Hop_Obj_t) *Abc_Obj_t { return Abc_ObjNotCond( (Abc_Obj_t *)Hop_ObjFanin1(pObj).pNext, Hop_ObjFaninC1(pObj) );  }

/**Function*************************************************************

  Synopsis    [Stops history AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHaigRecreateAig(pNtk *Abc_Ntk_t, p *Hop_Man_t) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pObjOld, pObjAbcThis, pObjAbcRepr *Abc_Obj_t;
    var pObj *Hop_Obj_t;
    var i int32;
    assert(p.nCreated == Vec_PtrSize(p.vObjs), "p.nCreated == Vec_PtrSize(p.vObjs)");

    // start the new network
    pNtkAig = Abc_NtkStartFrom( pNtk, ABC_NTK_STRASH, ABC_FUNC_AIG );

    // transfer new nodes to the PIs of HOP
    Hop_ManConst1(p).pNext = (Hop_Obj_t *)Abc_AigConst1( pNtkAig );
    // inlining of macro Hop_ManForEachPi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vPis)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vPis, i)), 1); i++  {
        pObj.pNext = (Hop_Obj_t *)Abc_NtkCi( pNtkAig, i );
    }

    // construct new nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if !Hop_ObjIsNode(pObj) {
            continue; }
        pObj.pNext = (Hop_Obj_t *)Abc_AigAnd( (Abc_Aig_t *)pNtkAig.pManFunc, Hop_ObjChild0Next(pObj), Hop_ObjChild1Next(pObj) );
        assert(!Hop_IsComplement(pObj.pNext), "!Hop_IsComplement(pObj.pNext)");
    }

    // set the COs
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObjOld) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_ObjAddFanin( pObjOld.pCopy, Hop_ObjChild0Next(Hop_ManPo(p,i)) );
    }

    // construct choice nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        // skip the node without choices
        if pObj.pData == nil {
            continue; }
        // skip the representative of the class
        if pObj.pData == pObj {
            continue; }
        // do not create choices for constant 1 and PIs
        if !Hop_ObjIsNode((Hop_Obj_t *)pObj.pData) {
            continue; }
        // get the corresponding new nodes
        pObjAbcThis = (Abc_Obj_t *)pObj.pNext;
        pObjAbcRepr = (Abc_Obj_t *)((Hop_Obj_t *)pObj.pData).pNext;
        // the new node cannot be already in the class
        assert(pObjAbcThis.pData == nil, "pObjAbcThis.pData == nil");
        // the new node cannot have fanouts
        assert(Abc_ObjFanoutNum(pObjAbcThis) == 0, "Abc_ObjFanoutNum(pObjAbcThis) == 0");
        // these should be different nodes
        assert(pObjAbcRepr != pObjAbcThis, "pObjAbcRepr != pObjAbcThis");
        // do not create choices if there is a path from pObjAbcThis to pObjAbcRepr
        if !(Abc_NtkHaigCheckTfi( pNtkAig, pObjAbcRepr, pObjAbcThis ) != 0) {
            // find the last node in the class
            for  pObjAbcRepr.pData  {
                pObjAbcRepr = (Abc_Obj_t *)pObjAbcRepr.pData;
            }
            // add the new node at the end of the list
            pObjAbcRepr.pData = pObjAbcThis;
        }
    }

    // finish the new network
//    Abc_NtkFinalize( pNtk, pNtkAig );
//    Abc_AigCleanup( pNtkAig->pManFunc );
    // check correctness of the network
    if !Abc_NtkCheck( pNtkAig ) {
        fmt.Printf("Abc_NtkHaigUse: The network check has failed.\n");
        Abc_NtkDelete( pNtkAig );
        return nil;
    }
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Resets representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHaigResetReprsOld(pMan *Hop_Man_t) {
    var vMembers, vClasses *Vec_Ptr_t;

    // collect members of the classes and make them point to reprs
    vMembers = Abc_NtkHaigCollectMembers( pMan );
    fmt.Printf("Collected %6d class members.\n", Vec_PtrSize(vMembers));

    // create classes
    vClasses = Abc_NtkHaigCreateClasses( vMembers );
    fmt.Printf("Collected %6d classes. (Ave = %5.2f)\n", Vec_PtrSize(vClasses), float32((Vec_PtrSize(vMembers)))/Vec_PtrSize(vClasses));

    Vec_PtrFree( vMembers );
    Vec_PtrFree( vClasses );
}

/**Function*************************************************************

  Synopsis    [Resets representatives.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHaigResetReprs(p *Hop_Man_t) int32 {
    var pObj, pRepr *Hop_Obj_t;
    var i, nClasses, nMembers, nFanouts, nNormals int32;
    // clear self-classes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        // fix the strange situation of double-loop
        pRepr = (Hop_Obj_t *)pObj.pData;
        if (pRepr != nil) && pRepr.pData == pObj {
            pRepr.pData = pRepr; }
        // remove self-loops
        if pObj.pData == pObj {
            pObj.pData = nil; }
    }
    // set representatives
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if pObj.pData == nil {
            continue; }
        // get representative of the node
        pRepr = Hop_ObjRepr( pObj );
        pRepr.pData = pRepr;
        // set the representative
        pObj.pData = pRepr;
    }
    // make each class point to the smallest topological order
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if pObj.pData == nil {
            continue; }
        pRepr = Hop_ObjRepr( pObj );
        if pRepr.Id > pObj.Id {
            pRepr.pData = pObj;
            pObj.pData = pObj;
        
        } else {
            pObj.pData = pRepr; }
    }
    // count classes, members, and fanouts - and verify
    nMembers = nClasses = nFanouts = nNormals = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Hop_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {
        if pObj.pData == nil {
            continue; }
        // count members
        nMembers++;
        // count the classes and fanouts
        if pObj.pData == pObj {
            nClasses++;
        } else if Hop_ObjRefs(pObj) > 0 {
            nFanouts++;
        } else {
            nNormals++; }
        // compare representatives
        pRepr = Hop_ObjRepr( pObj );
        assert(pObj.pData == pRepr, "pObj.pData == pRepr");
        assert(pRepr.Id <= pObj.Id, "pRepr.Id <= pObj.Id");
    }
//    printf( "Nodes = %7d.  Member = %7d.  Classes = %6d.  Fanouts = %6d.  Normals = %6d.\n", 
//        Hop_ManNodeNum(p), nMembers, nClasses, nFanouts, nNormals );
    return nFanouts;
}

/**Function*************************************************************

  Synopsis    [Transform HOP manager into the one without loops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHopRemoveLoops(pNtk *Abc_Ntk_t, pMan *Hop_Man_t) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pManTemp *Hop_Man_t;

    // iteratively reconstruct the HOP manager to create choice nodes
    for  Abc_NtkHaigResetReprs( pMan ) != 0  {
        pMan = Abc_NtkHaigReconstruct( pManTemp = pMan );
        Hop_ManStop( pManTemp );
    }

    // traverse in the topological order and create new AIG
    pNtkAig = Abc_NtkHaigRecreateAig( pNtk, pMan );
    Hop_ManStop( pMan );
    return pNtkAig;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


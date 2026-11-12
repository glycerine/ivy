package ivy

import "fmt"

/**CFile****************************************************************

  FileName    [ivyRwrAlg.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Algebraic AIG rewriting.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyRwrAlg.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int Ivy_ManFindAlgCut( Ivy_Obj_t * pRoot, Vec_Ptr_t * vFront, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vCone );

// static Ivy_Obj_t * Ivy_NodeRewriteAlg( Ivy_Obj_t * pObj, Vec_Ptr_t * vFront, Vec_Ptr_t * vLeaves, Vec_Ptr_t * vCone, Vec_Ptr_t * vSols, int LevelR, int fUseZeroCost );

// static int Ivy_NodeCountMffc( Ivy_Obj_t * pNode );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Algebraic AIG rewriting.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManRewriteAlg(p *Ivy_Man_t, fUpdateLevel int32, fUseZeroCost int32) int32 {
    var vRequired *Vec_Int_t;
    var vFront, vLeaves, vCone, vSol *Vec_Ptr_t;
    var pObj, pResult *Ivy_Obj_t;
    var i, RetValue, LevelR, nNodesOld int32;
    var CountUsed, CountUndo int32;
    if fUpdateLevel != 0 {
        vRequired = Ivy_ManRequiredLevels( p );
    } else {
        vRequired = nil;
    }
    vFront    = Vec_PtrAlloc( 100 );
    vLeaves   = Vec_PtrAlloc( 100 );
    vCone     = Vec_PtrAlloc( 100 );
    vSol      = Vec_PtrAlloc( 100 );
    // go through the nodes in the topological order
    CountUsed = CountUndo = 0;
    nNodesOld = Ivy_ManObjIdNext(p);
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        assert(!Ivy_ObjIsBuf(pObj), "!Ivy_ObjIsBuf(pObj)");
        if i >= nNodesOld {
            break; }
        // skip no-nodes and MUX roots
        if !Ivy_ObjIsNode(pObj) || Ivy_ObjIsExor(pObj) || Ivy_ObjIsMuxType(pObj) {
            continue; }
//        if ( pObj->Id > 297 ) // 296 --- 297 
//            break;
        if pObj.Id == 297 {
            var x int32;
        }
        // get the largest algebraic cut
        RetValue = Ivy_ManFindAlgCut( pObj, vFront, vLeaves, vCone );
        // the case of a trivial tree cut
        if RetValue == 1 {
            continue; }
        // the case of constant 0 cone
        if RetValue == -1 {
            Ivy_ObjReplace( pObj, Ivy_ManConst0(p), 1, 0, 1 ); 
            continue;
        }
        assert(Vec_PtrSize(vLeaves) > 2, "Vec_PtrSize(vLeaves) > 2");
        // get the required level for this node
        if vRequired != nil {
            LevelR = Vec_IntEntry(vRequired, pObj.Id);
        } else {
            LevelR = 1000000;
        }
        // create a new cone
        pResult = Ivy_NodeRewriteAlg( pObj, vFront, vLeaves, vCone, vSol, LevelR, fUseZeroCost );
        if pResult == nil || pResult == pObj {
            continue; }
        assert(Vec_PtrSize(vSol) == 1 || !Ivy_IsComplement(pResult), "Vec_PtrSize(vSol) == 1 || !Ivy_IsComplement(pResult)");
        if Ivy_ObjLevel(Ivy_Regular(pResult)) > LevelR && Ivy_ObjRefs(Ivy_Regular(pResult)) == 0 {
            Ivy_ObjDelete_rec(Ivy_Regular(pResult), 1);
CountUndo++;
        } else {
            Ivy_ObjReplace( pObj, pResult, 1, 0, 1 );
CountUsed++; } 
    }
    }
    fmt.Printf("Used = %d. Undo = %d.\n", CountUsed, CountUndo);
    Vec_PtrFree( vFront );
    Vec_PtrFree( vCone );
    Vec_PtrFree( vSol );
    if vRequired != nil { Vec_IntFree( vRequired ); }
    if i = Ivy_ManCleanup(p) {
        fmt.Printf("Cleanup after rewriting removed %d dangling nodes.\n", i); }
    if !Ivy_ManCheck(p) {
        fmt.Printf("Ivy_ManRewriteAlg(): The check has failed.\n"); }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Analizes one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeRewriteAlg(pObj *Ivy_Obj_t, vFront *Vec_Ptr_t, vLeaves *Vec_Ptr_t, vCone *Vec_Ptr_t, vSols *Vec_Ptr_t, LevelR int32, fUseZeroCost int32) *Ivy_Obj_t {
    var fVerbose int32;
    var pTemp *Ivy_Obj_t;
    var k, Counter, nMffc, RetValue int32;

    if fVerbose != 0 {
        if Ivy_ObjIsExor(pObj) {
            fmt.Printf("x ");
        } else {
            fmt.Printf("  "); }
    }

/*       
    printf( "%d ", Vec_PtrSize(vFront) );
    printf( "( " );
    Vec_PtrForEachEntry( Ivy_Obj_t *, vFront, pTemp, k )
        printf( "%d ", Ivy_ObjRefs(Ivy_Regular(pTemp)) );
    printf( ")\n" );
*/
    // collect nodes in the cone
    if Ivy_ObjIsExor(pObj) {
        Ivy_ManCollectCone( pObj, vFront, vCone );
    } else {
        Ivy_ManCollectCone( pObj, vLeaves, vCone ); }

    // deref nodes in the cone
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vCone)) && (((pTemp) = (Ivy_Obj_t *)Vec_PtrEntry(vCone, k)), 1); k++  {
        Ivy_ObjRefsDec( Ivy_ObjFanin0(pTemp) );
        Ivy_ObjRefsDec( Ivy_ObjFanin1(pTemp) );
        pTemp.fMarkB = 1;
    }

    // count the MFFC size
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vFront)) && (((pTemp) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, k)), 1); k++  {
        Ivy_Regular(pTemp).fMarkA = 1;
    }
    nMffc = Ivy_NodeCountMffc( pObj );
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vFront)) && (((pTemp) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, k)), 1); k++  {
        Ivy_Regular(pTemp).fMarkA = 0;
    }

    if fVerbose != 0 {
    Counter = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vCone)) && (((pTemp) = (Ivy_Obj_t *)Vec_PtrEntry(vCone, k)), 1); k++  {
        Counter += (Ivy_ObjRefs(pTemp) > 0);
    }
    fmt.Printf("%5d : Leaves = %2d. Cone = %2d. ConeRef = %2d.   Mffc = %d.   Lev = %d.  LevR = %d.\n", pObj.Id, Vec_PtrSize(vFront), Vec_PtrSize(vCone), Counter-1, nMffc, Ivy_ObjLevel(pObj), LevelR);
    }
/*
    printf( "Leaves:" );
    Vec_PtrForEachEntry( Ivy_Obj_t *, vLeaves, pTemp, k )
        printf( " %d%s", Ivy_Regular(pTemp)->Id, Ivy_IsComplement(pTemp)? "\'" : "" );
    printf( "\n" );
    printf( "Cone:\n" );
    Vec_PtrForEachEntry( Ivy_Obj_t *, vCone, pTemp, k )
        printf( " %5d = %d%s %d%s\n", pTemp->Id,
            Ivy_ObjFaninId0(pTemp), Ivy_ObjFaninC0(pTemp)? "\'" : "",
            Ivy_ObjFaninId1(pTemp), Ivy_ObjFaninC1(pTemp)? "\'" : "" );
*/

    RetValue = Ivy_MultiPlus( vLeaves, vCone, Ivy_ObjType(pObj), nMffc + fUseZeroCost, vSols );

    // ref nodes in the cone
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vCone)) && (((pTemp) = (Ivy_Obj_t *)Vec_PtrEntry(vCone, k)), 1); k++  {
        Ivy_ObjRefsInc( Ivy_ObjFanin0(pTemp) );
        Ivy_ObjRefsInc( Ivy_ObjFanin1(pTemp) );
        pTemp.fMarkA = 0;
        pTemp.fMarkB = 0;
    }

    if !(RetValue != 0) {
        return nil; }

    if Vec_PtrSize( vSols ) == 1 {
        return Vec_PtrEntry( vSols, 0 ); }
    return Ivy_NodeBalanceBuildSuper( vSols, Ivy_ObjType(pObj), 1 );
}

/**Function*************************************************************

  Synopsis    [Comparison for node pointers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCountMffc_rec(pNode *Ivy_Obj_t) int32 {
    if Ivy_ObjRefs(pNode) > 0 || Ivy_ObjIsCi(pNode) || pNode.fMarkA {
        return 0; }
    assert(pNode.fMarkB, "pNode.fMarkB");
    pNode.fMarkA = 1;
//    printf( "%d ", pNode->Id );
    if Ivy_ObjIsBuf(pNode) {
        return Ivy_NodeCountMffc_rec( Ivy_ObjFanin0(pNode) ); }
    return 1 + Ivy_NodeCountMffc_rec( Ivy_ObjFanin0(pNode) ) + Ivy_NodeCountMffc_rec( Ivy_ObjFanin1(pNode) );
}

/**Function*************************************************************

  Synopsis    [Comparison for node pointers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCountMffc(pNode *Ivy_Obj_t) int32 {
    assert(pNode.fMarkB, "pNode.fMarkB");
    return 1 + Ivy_NodeCountMffc_rec( Ivy_ObjFanin0(pNode) ) + Ivy_NodeCountMffc_rec( Ivy_ObjFanin1(pNode) );
}

/**Function*************************************************************

  Synopsis    [Comparison for node pointers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManFindAlgCutCompare(pp1 **Ivy_Obj_t, pp2 **Ivy_Obj_t) int32 {
    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Computing one algebraic cut.]

  Description [Returns 1 if the tree-leaves of this node where traversed 
  and found to have no external references (and have not been collected). 
  Returns 0 if the tree-leaves have external references and are collected.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManFindAlgCut_rec(pObj *Ivy_Obj_t, Type Ivy_Type_t, vFront *Vec_Ptr_t, vCone *Vec_Ptr_t) int32 {
    var RetValue0, RetValue1 int32;
    var pObjR *Ivy_Obj_t  = Ivy_Regular(pObj);
    assert(!Ivy_ObjIsBuf(pObjR), "!Ivy_ObjIsBuf(pObjR)");
    assert(Type != IVY_EXOR || !Ivy_IsComplement(pObj), "Type != IVY_EXOR || !Ivy_IsComplement(pObj)");

    // make sure the node is not visited twice in different polarities
    if Ivy_IsComplement(pObj) { // if complemented, mark B
        if pObjR.fMarkA {
            return -1; }
        pObjR.fMarkB = 1;
    
    } else { // if non-complicated, mark A
        if pObjR.fMarkB {
            return -1; }
        pObjR.fMarkA = 1;
    }
    Vec_PtrPush( vCone, pObjR );

    // if the node is the end of the tree, return
    if Ivy_IsComplement(pObj) || Ivy_ObjType(pObj) != Type {
        if Ivy_ObjRefs(pObjR) == 1 {
            return 1; }
        assert(Ivy_ObjRefs(pObjR) > 1, "Ivy_ObjRefs(pObjR) > 1");
        Vec_PtrPush( vFront, pObj );
        return 0;
    }

    // branch on the node
    assert(!Ivy_IsComplement(pObj), "!Ivy_IsComplement(pObj)");
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    // what if buffer has more than one fanout???
    RetValue0 = Ivy_ManFindAlgCut_rec( Ivy_ObjReal( Ivy_ObjChild0(pObj) ), Type, vFront, vCone );
    RetValue1 = Ivy_ManFindAlgCut_rec( Ivy_ObjReal( Ivy_ObjChild1(pObj) ), Type, vFront, vCone );
    if RetValue0 == -1 || RetValue1 == -1 {
        return -1; }

    // the case when both have no external references
    if RetValue0 && (RetValue1 != 0) {
        if Ivy_ObjRefs(pObj) == 1 {
            return 1; }
        assert(Ivy_ObjRefs(pObj) > 1, "Ivy_ObjRefs(pObj) > 1");
        Vec_PtrPush( vFront, pObj );
        return 0;
    }
    // the case when one of them has external references
    if RetValue0 {
        Vec_PtrPush( vFront, Ivy_ObjReal( Ivy_ObjChild0(pObj) ) ); }
    if RetValue1 != 0 {
        Vec_PtrPush( vFront, Ivy_ObjReal( Ivy_ObjChild1(pObj) ) ); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Computing one algebraic cut.]

  Description [Algebraic cut stops when we hit (a) CI, (b) complemented edge,
  (c) boundary of different gates. Returns 1 if this is a pure tree.
  Returns -1 if the contant 0 is detected. Return 0 if the array can be used.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManFindAlgCut(pRoot *Ivy_Obj_t, vFront *Vec_Ptr_t, vLeaves *Vec_Ptr_t, vCone *Vec_Ptr_t) int32 {
    var pObj, pPrev *Ivy_Obj_t;
    var RetValue, i int32;
    assert(!Ivy_IsComplement(pRoot), "!Ivy_IsComplement(pRoot)");
    assert(Ivy_ObjIsNode(pRoot), "Ivy_ObjIsNode(pRoot)");
    // clear the frontier and collect the nodes
    Vec_PtrClear( vCone );
    Vec_PtrClear( vFront );
    Vec_PtrClear( vLeaves );
    RetValue = Ivy_ManFindAlgCut_rec( pRoot, Ivy_ObjType(pRoot), vFront, vCone );
    // clean the marks
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCone)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vCone, i)), 1); i++  {
        pObj.fMarkA = pObj.fMarkB = 0;
    }
    // quit if the same node is found in both polarities
    if RetValue == -1 {
        return -1; }
    // return if the node is the root of a tree
    if RetValue == 1 {
        return 1; }
    // return if the cut is composed of two nodes
    if Vec_PtrSize(vFront) <= 2 {
        return 1; }
    // sort the entries in increasing order
    Vec_PtrSort( vFront, (int (*)(void))Ivy_ManFindAlgCutCompare );
    // remove duplicates from vFront and save the nodes in vLeaves
    pPrev = Vec_PtrEntry(vFront, 0);
    Vec_PtrPush( vLeaves, pPrev );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vFront)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        // compare current entry and the previous entry
        if pObj == pPrev {
            if Ivy_ObjIsExor(pRoot) { // A <+> A = 0
            
                // vLeaves are no longer structural support of pRoot!!!
                Vec_PtrPop(vLeaves);  
                if Vec_PtrSize(vLeaves) == 0 {
                    pPrev = nil;
                } else {
                    pPrev = Vec_PtrEntryLast(vLeaves);
                }
            }
            continue;
        }
        if pObj == Ivy_Not(pPrev) {
            assert(Ivy_ObjIsAnd(pRoot), "Ivy_ObjIsAnd(pRoot)");
            return -1;
        }
        pPrev = pObj;
        Vec_PtrPush( vLeaves, pObj );
    }
    if Vec_PtrSize(vLeaves) == 0 {
        return -1; }
    if Vec_PtrSize(vLeaves) <= 2 {
        return 1; }
    return 0;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


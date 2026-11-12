package ivy

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [ivyCut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [And-Inverter Graph package.]

  Synopsis    [Computes reconvergence driven sequential cut.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: ivyCut.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "ivy.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Ivy_NodeCutHashValue(NodeId int32) int32 { return 1 << (NodeId % 31); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Evaluate the cost of removing the node from the set of leaves.]

  Description [Returns the number of new leaves that will be brought in.
  Returns large number if the node cannot be removed from the set of leaves.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeGetLeafCostOne(p *Ivy_Man_t, Leaf int32, vInside *Vec_Int_t) int32 {
    var pNode *Ivy_Obj_t;
    var nLatches, FaninLeaf, Cost int32;
    // make sure leaf is not a contant node
    assert(Leaf > 0, "Leaf > 0"); 
    // get the node
    pNode = Ivy_ManObj( p, Ivy_LeafId(Leaf) );
    // cannot expand over the PI node
    if Ivy_ObjIsPi(pNode) || Ivy_ObjIsConst1(pNode) {
        return 999; }
    // get the number of latches
    nLatches = Ivy_LeafLat(Leaf) + Ivy_ObjIsLatch(pNode);
    if nLatches > 15 {
        return 999; }
    // get the first fanin
    FaninLeaf = Ivy_LeafCreate( Ivy_ObjFaninId0(pNode), nLatches );
    Cost = FaninLeaf && (Vec_IntFind(vInside, FaninLeaf) == -1);
    // quit if this is the one fanin node
    if Ivy_ObjIsLatch(pNode) || Ivy_ObjIsBuf(pNode) {
        return Cost; }
    assert(Ivy_ObjIsNode(pNode), "Ivy_ObjIsNode(pNode)");
    // get the second fanin
    FaninLeaf = Ivy_LeafCreate( Ivy_ObjFaninId1(pNode), nLatches );
    Cost += FaninLeaf && (Vec_IntFind(vInside, FaninLeaf) == -1);
    return Cost;
}

/**Function*************************************************************

  Synopsis    [Builds reconvergence-driven cut by changing one leaf at a time.]

  Description [This procedure looks at the current leaves and tries to change 
  one leaf at a time in such a way that the cut grows as little as possible.
  In evaluating the fanins, this procedure looks only at their immediate 
  predecessors (this is why it is called a one-level construction procedure).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManSeqFindCut_int(p *Ivy_Man_t, vFront *Vec_Int_t, vInside *Vec_Int_t, nSizeLimit int32) int32 {
    var pNode *Ivy_Obj_t;
    var CostBest, CostCur, Leaf, LeafBest, Next, nLatches, i int32;
    var LeavesBest [10]int32;
    var Counter int32;

    // add random selection of the best fanin!!!

    // find the best fanin
    CostBest = 99;
    LeafBest = -1;
    Counter = -1;
//printf( "Evaluating fanins of the cut:\n" );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFront)) && (((Leaf) = Vec_IntEntry(vFront, i)), 1); i++  {
        CostCur = Ivy_NodeGetLeafCostOne( p, Leaf, vInside );
//printf( "    Fanin %s has cost %d.\n", Ivy_ObjName(pNode), CostCur );
        if CostBest > CostCur {
            CostBest = CostCur;
            LeafBest = Leaf;
            LeavesBest[0] = Leaf;
            Counter = 1;
        
        } else if CostBest == CostCur {
            LeavesBest[Counter++] = Leaf; }

        if CostBest <= 1 { // can be if ( CostBest <= 1 )
            break; }
    }
    if CostBest == 99 {
        return 0; }
//        return Ivy_NodeBuildCutLevelTwo_int( vInside, vFront, nFaninLimit );

    assert(CostBest < 3, "CostBest < 3");
    if Vec_IntSize(vFront) - 1 + CostBest > nSizeLimit {
        return 0; }
//        return Ivy_NodeBuildCutLevelTwo_int( vInside, vFront, nFaninLimit );

    assert(Counter > 0, "Counter > 0");
fmt.Printf("%d", Counter);

    LeafBest = LeavesBest[rand() % Counter];

    // remove the node from the array
    assert(LeafBest >= 0, "LeafBest >= 0");
    Vec_IntRemove( vFront, LeafBest );
//printf( "Removing fanin %s.\n", Ivy_ObjName(pNode) );

    // get the node and its latches
    pNode = Ivy_ManObj( p, Ivy_LeafId(LeafBest) );
    nLatches = Ivy_LeafLat(LeafBest) + Ivy_ObjIsLatch(pNode);
    assert(Ivy_ObjIsNode(pNode) || Ivy_ObjIsLatch(pNode) || Ivy_ObjIsBuf(pNode), "Ivy_ObjIsNode(pNode) || Ivy_ObjIsLatch(pNode) || Ivy_ObjIsBuf(pNode)");

    // add the left child to the fanins
    Next = Ivy_LeafCreate( Ivy_ObjFaninId0(pNode), nLatches );
    if Next && Vec_IntFind(vInside, Next) == -1 {
//printf( "Adding fanin %s.\n", Ivy_ObjName(pNext) );
        Vec_IntPush( vFront, Next );
        Vec_IntPush( vInside, Next );
    }

    // quit if this is the one fanin node
    if Ivy_ObjIsLatch(pNode) || Ivy_ObjIsBuf(pNode) {
        return 1; }
    assert(Ivy_ObjIsNode(pNode), "Ivy_ObjIsNode(pNode)");

    // add the right child to the fanins
    Next = Ivy_LeafCreate( Ivy_ObjFaninId1(pNode), nLatches );
    if Next && Vec_IntFind(vInside, Next) == -1 {
//printf( "Adding fanin %s.\n", Ivy_ObjName(pNext) );
        Vec_IntPush( vFront, Next );
        Vec_IntPush( vInside, Next );
    }
    assert(Vec_IntSize(vFront) <= nSizeLimit, "Vec_IntSize(vFront) <= nSizeLimit");
    // keep doing this
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes one sequential cut of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManSeqFindCut(p *Ivy_Man_t, pRoot *Ivy_Obj_t, vFront *Vec_Int_t, vInside *Vec_Int_t, nSize int32) {
    assert(!Ivy_IsComplement(pRoot), "!Ivy_IsComplement(pRoot)");
    assert(Ivy_ObjIsNode(pRoot), "Ivy_ObjIsNode(pRoot)");
    assert(Ivy_ObjFaninId0(pRoot), "Ivy_ObjFaninId0(pRoot)");
    assert(Ivy_ObjFaninId1(pRoot), "Ivy_ObjFaninId1(pRoot)");

    // start the cut 
    Vec_IntClear( vFront );
    Vec_IntPush( vFront, Ivy_LeafCreate(Ivy_ObjFaninId0(pRoot), 0) );
    Vec_IntPush( vFront, Ivy_LeafCreate(Ivy_ObjFaninId1(pRoot), 0) );

    // start the visited nodes
    Vec_IntClear( vInside );
    Vec_IntPush( vInside, Ivy_LeafCreate(pRoot.Id, 0) );
    Vec_IntPush( vInside, Ivy_LeafCreate(Ivy_ObjFaninId0(pRoot), 0) );
    Vec_IntPush( vInside, Ivy_LeafCreate(Ivy_ObjFaninId1(pRoot), 0) );

    // compute the cut
    for  Ivy_ManSeqFindCut_int( p, vFront, vInside, nSize ) != 0  {; }
    assert(Vec_IntSize(vFront) <= nSize, "Vec_IntSize(vFront) <= nSize");
}





/**Function*************************************************************

  Synopsis    [Computing Boolean cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManFindBoolCut_rec(p *Ivy_Man_t, pObj *Ivy_Obj_t, vLeaves *Vec_Ptr_t, vVolume *Vec_Ptr_t, pPivot *Ivy_Obj_t) int32 {
    var RetValue0, RetValue1 int32;
    if pObj == pPivot {
        Vec_PtrPushUnique( vLeaves, pObj );
        Vec_PtrPushUnique( vVolume, pObj );
        return 1;        
    }
    if pObj.fMarkA {
        return 0; }

//    assert( !Ivy_ObjIsCi(pObj) );
    if Ivy_ObjIsCi(pObj) {
        return 0; }

    if Ivy_ObjIsBuf(pObj) {
        RetValue0 = Ivy_ManFindBoolCut_rec( p, Ivy_ObjFanin0(pObj), vLeaves, vVolume, pPivot );
        if !RetValue0 {
            return 0; }
        Vec_PtrPushUnique( vVolume, pObj );
        return 1;
    }
    assert(Ivy_ObjIsNode(pObj), "Ivy_ObjIsNode(pObj)");
    RetValue0 = Ivy_ManFindBoolCut_rec( p, Ivy_ObjFanin0(pObj), vLeaves, vVolume, pPivot );
    RetValue1 = Ivy_ManFindBoolCut_rec( p, Ivy_ObjFanin1(pObj), vLeaves, vVolume, pPivot );
    if !RetValue0 && !(RetValue1 != 0) {
        return 0; }
    // add new leaves
    if !RetValue0 {
        Vec_PtrPushUnique( vLeaves, Ivy_ObjFanin0(pObj) );
        Vec_PtrPushUnique( vVolume, Ivy_ObjFanin0(pObj) );
    }
    if !(RetValue1 != 0) {
        Vec_PtrPushUnique( vLeaves, Ivy_ObjFanin1(pObj) );
        Vec_PtrPushUnique( vVolume, Ivy_ObjFanin1(pObj) );
    }
    Vec_PtrPushUnique( vVolume, pObj );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns the cost of one node (how many new nodes are added.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManFindBoolCutCost(pObj *Ivy_Obj_t) int32 {
    var Cost int32;
    // make sure the node is in the construction zone
    assert(pObj.fMarkA == 1, "pObj.fMarkA == 1");  
    // cannot expand over the PI node
    if Ivy_ObjIsCi(pObj) {
        return 999; }
    // always expand over the buffer
    if Ivy_ObjIsBuf(pObj) {
        return !Ivy_ObjFanin0(pObj).fMarkA; }
    // get the cost of the cone
    Cost = (!Ivy_ObjFanin0(pObj).fMarkA) + (!Ivy_ObjFanin1(pObj).fMarkA);
    // return the number of nodes to be added to the leaves if this node is removed
    return Cost;
}

/**Function*************************************************************

  Synopsis    [Computing Boolean cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManFindBoolCut(p *Ivy_Man_t, pRoot *Ivy_Obj_t, vFront *Vec_Ptr_t, vVolume *Vec_Ptr_t, vLeaves *Vec_Ptr_t) int32 {
    var pObj *Ivy_Obj_t  = nil; // Suppress "might be used uninitialized"
    var pFaninC, pFanin0, pFanin1, pPivot *Ivy_Obj_t;
    var RetValue, LevelLimit, Lev, k int32;
    assert(!Ivy_IsComplement(pRoot), "!Ivy_IsComplement(pRoot)");
    // clear the frontier and collect the nodes
    Vec_PtrClear( vFront );
    Vec_PtrClear( vVolume );
    if Ivy_ObjIsMuxType(pRoot) {
        pFaninC = Ivy_ObjRecognizeMux( pRoot, &pFanin0, &pFanin1 );
    } else {
        pFaninC = nil;
        pFanin0 = Ivy_ObjFanin0(pRoot);
        pFanin1 = Ivy_ObjFanin1(pRoot); 
    }
    // start cone A
    pFanin0.fMarkA = 1;
    Vec_PtrPush( vFront, pFanin0 );
    Vec_PtrPush( vVolume, pFanin0 );
    // start cone B
    pFanin1.fMarkB = 1;
    Vec_PtrPush( vFront, pFanin1 );
    Vec_PtrPush( vVolume, pFanin1 );
    // iteratively expand until the common node (pPivot) is found or limit is reached
    assert(Ivy_ObjLevel(pRoot) == Ivy_ObjLevelNew(pRoot), "Ivy_ObjLevel(pRoot) == Ivy_ObjLevelNew(pRoot)");
    pPivot = nil;
    // inlining of macro IVY_MAX
    
    if (Ivy_ObjLevel(pRoot) - 10) > (1) {
        LevelLimit = (Ivy_ObjLevel(pRoot) - 10);
    } else {
        LevelLimit = (1);
    }
    for  Lev = Ivy_ObjLevel(pRoot) - 1; Lev >= LevelLimit; Lev--  {
        for {
            // find the next node to expand on this level
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vFront)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, k)), 1); k++  {
                if int32(pObj.Level) == Lev {
                    break; }
            }
            if k == Vec_PtrSize(vFront) {
                break; }
            assert(int32(pObj.Level) <= Lev, "int32(pObj.Level) <= Lev");
            assert(pObj.fMarkA ^ pObj.fMarkB, "pObj.fMarkA ^ pObj.fMarkB");
            // remove the old node
            Vec_PtrRemove( vFront, pObj );

            // expand this node
            pFanin0 = Ivy_ObjFanin0(pObj);
            if !pFanin0.fMarkA && !pFanin0.fMarkB {
                Vec_PtrPush( vFront, pFanin0 );
                Vec_PtrPush( vVolume, pFanin0 );
            }
            // mark the new nodes
            if pObj.fMarkA {
                pFanin0.fMarkA = 1; } 
            if pObj.fMarkB {
                pFanin0.fMarkB = 1; } 

            if Ivy_ObjIsBuf(pObj) {
                if pFanin0.fMarkA && pFanin0.fMarkB {
                    pPivot = pFanin0;
                    break;
                }
                continue;
            }

            // expand this node
            pFanin1 = Ivy_ObjFanin1(pObj); 
            if !pFanin1.fMarkA && !pFanin1.fMarkB {
                Vec_PtrPush( vFront, pFanin1 );
                Vec_PtrPush( vVolume, pFanin1 );
            }
            // mark the new nodes
            if pObj.fMarkA {
                pFanin1.fMarkA = 1; } 
            if pObj.fMarkB {
                pFanin1.fMarkB = 1; } 

            // consider if it is time to quit
            if pFanin0.fMarkA && pFanin0.fMarkB {
                pPivot = pFanin0;
                break;
            }
            if pFanin1.fMarkA && pFanin1.fMarkB {
                pPivot = pFanin1;
                break;
            }
        }
        if pPivot != nil {
            break; }
    }
    if pPivot == nil {
        return 0; }
    // if the MUX control is defined, it should not be
    if pFaninC && !pFaninC.fMarkA && !pFaninC.fMarkB {
        Vec_PtrPush( vFront, pFaninC ); }
    // clean the markings
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vVolume)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vVolume, k)), 1); k++  {
        pObj.fMarkA = pObj.fMarkB = 0;
    }

    // mark the nodes on the frontier (including the pivot)
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vFront)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, k)), 1); k++  {
        pObj.fMarkA = 1;
    }
    // cut exists, collect all the nodes on the shortest path to the pivot
    Vec_PtrClear( vLeaves );
    Vec_PtrClear( vVolume );
    RetValue = Ivy_ManFindBoolCut_rec( p, pRoot, vLeaves, vVolume, pPivot );
    assert(RetValue == 1, "RetValue == 1");
    // unmark the nodes on the frontier (including the pivot)
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vFront)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vFront, k)), 1); k++  {
        pObj.fMarkA = 0;
    }

    // mark the nodes in the volume
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vVolume)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vVolume, k)), 1); k++  {
        pObj.fMarkA = 1;
    }
    // expand the cut without increasing its size
    for {
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vLeaves)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vLeaves, k)), 1); k++  {
            if Ivy_ManFindBoolCutCost(pObj) < 2 {
                break; }
        }
        if k == Vec_PtrSize(vLeaves) {
            break; }
        // the node can be expanded
        // remove the old node
        Vec_PtrRemove( vLeaves, pObj );
        // expand this node
        pFanin0 = Ivy_ObjFanin0(pObj);
        if !pFanin0.fMarkA {
            pFanin0.fMarkA = 1;
            Vec_PtrPush( vVolume, pFanin0 );
            Vec_PtrPush( vLeaves, pFanin0 );
        }
        if Ivy_ObjIsBuf(pObj) {
            continue; }
        // expand this node
        pFanin1 = Ivy_ObjFanin1(pObj);
        if !pFanin1.fMarkA {
            pFanin1.fMarkA = 1;
            Vec_PtrPush( vVolume, pFanin1 );
            Vec_PtrPush( vLeaves, pFanin1 );
        }        
    }
    // unmark the nodes in the volume
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vVolume)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(vVolume, k)), 1); k++  {
        pObj.fMarkA = 0;
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManTestCutsBool(p *Ivy_Man_t) {
    var vFront, vVolume, vLeaves *Vec_Ptr_t;
    var pObj *Ivy_Obj_t;//, * pTemp;
    var i, RetValue int32;//, k;
    vFront = Vec_PtrAlloc( 100 );
    vVolume = Vec_PtrAlloc( 100 );
    vLeaves = Vec_PtrAlloc( 100 );
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Ivy_ObjIsNode(pObj) {
            continue; }
        if Ivy_ObjIsMuxType(pObj) {
            fmt.Printf("m");
            continue;
        }
        if Ivy_ObjIsExor(pObj) {
            fmt.Printf("x"); }
        RetValue = Ivy_ManFindBoolCut( p, pObj, vFront, vVolume, vLeaves );
        if RetValue == 0 {
            fmt.Printf("- ");
        } else {
            fmt.Printf("%d ", Vec_PtrSize(vLeaves)); }
/*        
        printf( "( " );
        Vec_PtrForEachEntry( Ivy_Obj_t *, vFront, pTemp, k )
            printf( "%d ", Ivy_ObjRefs(Ivy_Regular(pTemp)) );
        printf( ")\n" );
*/
    }
    }
    fmt.Printf("\n");
    Vec_PtrFree( vFront );
    Vec_PtrFree( vVolume );
    Vec_PtrFree( vLeaves );
}



/**Function*************************************************************

  Synopsis    [Find the hash value of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCutHash(pCut *Ivy_Cut_t) uint32 {
    var i int32;
//    for ( i = 1; i < pCut->nSize; i++ )
//        assert( pCut->pArray[i-1] < pCut->pArray[i] );
    pCut.uHash = 0;
    for  i = 0; i < pCut.nSize; i++  {
        pCut.uHash |= (1 << (pCut.pArray[i] % 31));
    }
    return pCut.uHash;
}

/**Function*************************************************************

  Synopsis    [Removes one node to the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCutShrink(pCut *Ivy_Cut_t, iOld int32) {
    var i, k int32;
    for  i = k = 0; i < pCut.nSize; i++  {
        if pCut.pArray[i] != iOld {
            pCut.pArray[k++] = pCut.pArray[i]; }
    }
    assert(k == pCut.nSize - 1, "k == pCut.nSize - 1");
    pCut.nSize--;
}

/**Function*************************************************************

  Synopsis    [Adds one node to the cut.]

  Description [Returns 1 if the cuts is still okay.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCutExtend(pCut *Ivy_Cut_t, iNew int32) int32 {
    var i int32;
    for  i = 0; i < pCut.nSize; i++  {
        if pCut.pArray[i] == iNew {
            return 1; }
    }
    // check if there is room
    if pCut.nSize == pCut.nSizeMax {
        return 0; }
    // add the new one
    for  i = pCut.nSize - 1; i >= 0; i--  {
        if pCut.pArray[i] > iNew {
            pCut.pArray[i+1] = pCut.pArray[i];
        } else {
            assert(pCut.pArray[i] < iNew, "pCut.pArray[i] < iNew");
            break;
        }
    }
    pCut.pArray[i+1] = iNew;
    pCut.nSize++;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the cut can be constructed; 0 otherwise.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCutPrescreen(pCut *Ivy_Cut_t, Id0 int32, Id1 int32) int32 {
    var i int32;
    if pCut.nSize < pCut.nSizeMax {
        return 1; }
    for  i = 0; i < pCut.nSize; i++  {
        if pCut.pArray[i] == Id0 || pCut.pArray[i] == Id1 {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Derives new cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCutDeriveNew(pCut *Ivy_Cut_t, pCutNew *Ivy_Cut_t, IdOld int32, IdNew0 int32, IdNew1 int32) int32 {
    var uHash uint32;
    var i, k int32; 
    assert(pCut.nSize > 0, "pCut.nSize > 0");
    assert(IdNew0 < IdNew1, "IdNew0 < IdNew1");
    for  i = k = 0; i < pCut.nSize; i++  {
        if pCut.pArray[i] == IdOld {
            continue; }
        if IdNew0 <= pCut.pArray[i] {
            if IdNew0 < pCut.pArray[i] {
                pCutNew.pArray[ k++ ] = IdNew0;
                uHash |= Ivy_NodeCutHashValue( IdNew0 );
            }
            IdNew0 = 0x7FFFFFFF;
        }
        if IdNew1 <= pCut.pArray[i] {
            if IdNew1 < pCut.pArray[i] {
                pCutNew.pArray[ k++ ] = IdNew1;
                uHash |= Ivy_NodeCutHashValue( IdNew1 );
            }
            IdNew1 = 0x7FFFFFFF;
        }
        pCutNew.pArray[ k++ ] = pCut.pArray[i];
        uHash |= Ivy_NodeCutHashValue( pCut.pArray[i] );
    }
    if IdNew0 < 0x7FFFFFFF {
        pCutNew.pArray[ k++ ] = IdNew0;
        uHash |= Ivy_NodeCutHashValue( IdNew0 );
    }
    if IdNew1 < 0x7FFFFFFF {
        pCutNew.pArray[ k++ ] = IdNew1;
        uHash |= Ivy_NodeCutHashValue( IdNew1 );
    }
    pCutNew.nSize = k;
    pCutNew.uHash = uHash;
    assert(pCutNew.nSize <= pCut.nSizeMax, "pCutNew.nSize <= pCut.nSizeMax");
//    for ( i = 1; i < pCutNew->nSize; i++ )
//        assert( pCutNew->pArray[i-1] < pCutNew->pArray[i] );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Check if the cut exists.]

  Description [Returns 1 if the cut exists.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCutFindOrAdd(pCutStore *Ivy_Store_t, pCutNew *Ivy_Cut_t) int32 {
    var pCut *Ivy_Cut_t;
    var i, k int32;
    assert(pCutNew.uHash, "pCutNew.uHash");
    // try to find the cut
    for  i = 0; i < pCutStore.nCuts; i++  {
        pCut = pCutStore.pCuts + i;
        if pCut.uHash == pCutNew.uHash && pCut.nSize == pCutNew.nSize {
            for  k = 0; k < pCutNew.nSize; k++  {
                if pCut.pArray[k] != pCutNew.pArray[k] {
                    break; }
            }
            if k == pCutNew.nSize {
                return 1; }
        }
    }
    assert(pCutStore.nCuts < pCutStore.nCutsMax, "pCutStore.nCuts < pCutStore.nCutsMax");
    // add the cut
    pCut = pCutStore.pCuts + pCutStore.nCuts++;
    *pCut = *pCutNew;
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pDom is contained in pCut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_CutCheckDominance(pDom *Ivy_Cut_t, pCut *Ivy_Cut_t) int32 {
    var i, k int32;
    for  i = 0; i < pDom.nSize; i++  {
        for  k = 0; k < pCut.nSize; k++  {
            if pDom.pArray[i] == pCut.pArray[k] {
                break; }
        }
        if k == pCut.nSize { // node i in pDom is not contained in pCut
            return 0; }
    }
    // every node in pDom is contained in pCut
    return 1;
}

/**Function*************************************************************

  Synopsis    [Check if the cut exists.]

  Description [Returns 1 if the cut exists.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCutFindOrAddFilter(pCutStore *Ivy_Store_t, pCutNew *Ivy_Cut_t) int32 {
    var pCut *Ivy_Cut_t;
    var i, k int32;
    assert(pCutNew.uHash, "pCutNew.uHash");
    // try to find the cut
    for  i = 0; i < pCutStore.nCuts; i++  {
        pCut = pCutStore.pCuts + i;
        if pCut.nSize == 0 {
            continue; }
        if pCut.nSize == pCutNew.nSize {
            if pCut.uHash == pCutNew.uHash {
                for  k = 0; k < pCutNew.nSize; k++  {
                    if pCut.pArray[k] != pCutNew.pArray[k] {
                        break; }
                }
                if k == pCutNew.nSize {
                    return 1; }
            }
            continue;
        }
        if pCut.nSize < pCutNew.nSize {
            // skip the non-contained cuts
            if (pCut.uHash & pCutNew.uHash) != pCut.uHash {
                continue; }
            // check containment seriously
            if Ivy_CutCheckDominance( pCut, pCutNew ) != 0 {
                return 1; }
            continue;
        }
        // check potential containment of other cut

        // skip the non-contained cuts
        if (pCut.uHash & pCutNew.uHash) != pCutNew.uHash {
            continue; }
        // check containment seriously
        if Ivy_CutCheckDominance( pCutNew, pCut ) != 0 {
            // remove the current cut
//            --pCutStore->nCuts;
//            for ( k = i; k < pCutStore->nCuts; k++ )
//                pCutStore->pCuts[k] = pCutStore->pCuts[k+1];
//            i--;
            pCut.nSize = 0;
        }
    }
    assert(pCutStore.nCuts < pCutStore.nCutsMax, "pCutStore.nCuts < pCutStore.nCutsMax");
    // add the cut
    pCut = pCutStore.pCuts + pCutStore.nCuts++;
    *pCut = *pCutNew;
    return 0;
}

/**Function*************************************************************

  Synopsis    [Print the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodeCompactCuts(pCutStore *Ivy_Store_t) {
    var pCut *Ivy_Cut_t;
    var i, k int32;
    for  i = k = 0; i < pCutStore.nCuts; i++  {
        pCut = pCutStore.pCuts + i;
        if pCut.nSize == 0 {
            continue; }
        pCutStore.pCuts[k++] = *pCut;
    }
    pCutStore.nCuts = k;
}

/**Function*************************************************************

  Synopsis    [Print the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodePrintCut(pCut *Ivy_Cut_t) {
    var i int32;
    assert(pCut.nSize > 0, "pCut.nSize > 0");
    fmt.Printf("%d : {", pCut.nSize);
    for  i = 0; i < pCut.nSize; i++  {
        fmt.Printf(" %d", pCut.pArray[i]);
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_NodePrintCuts(pCutStore *Ivy_Store_t) {
    var i int32;
    fmt.Printf("Node %d\n", pCutStore.pCuts[0].pArray[0]);
    for  i = 0; i < pCutStore.nCuts; i++  {
        Ivy_NodePrintCut( pCutStore.pCuts + i );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ObjRealFanin(pObj *Ivy_Obj_t) *Ivy_Obj_t {
    if !Ivy_ObjIsBuf(pObj) {
        return pObj; }
    return Ivy_ObjRealFanin( Ivy_ObjFanin0(pObj) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ivy_NodeFindCutsAll_CutStore Ivy_Store_t;
var spStatic_Ivy_NodeFindCutsAll_pCutStore *Ivy_Store_t = &spStatic_Ivy_NodeFindCutsAll_CutStore;
func Ivy_NodeFindCutsAll(p *Ivy_Man_t, pObj *Ivy_Obj_t, nLeaves int32) *Ivy_Store_t {
    
    var CutNew Ivy_Cut_t;
    var pCutNew *Ivy_Cut_t = &CutNew;
    var pCut *Ivy_Cut_t;
    var pLeaf *Ivy_Obj_t;
    var i, k, iLeaf0, iLeaf1 int32;

    assert(nLeaves <= IVY_CUT_INPUT, "nLeaves <= IVY_CUT_INPUT");

    // start the structure
    spStatic_Ivy_NodeFindCutsAll_pCutStore.nCuts = 0;
    spStatic_Ivy_NodeFindCutsAll_pCutStore.nCutsMax = IVY_CUT_LIMIT;
    // start the trivial cut
    pCutNew.uHash = 0;
    pCutNew.nSize = 1;
    pCutNew.nSizeMax = nLeaves;
    pCutNew.pArray[0] = pObj.Id;
    Ivy_NodeCutHash( pCutNew );
    // add the trivial cut
    Ivy_NodeCutFindOrAdd( spStatic_Ivy_NodeFindCutsAll_pCutStore, pCutNew );
    assert(spStatic_Ivy_NodeFindCutsAll_pCutStore.nCuts == 1, "spStatic_Ivy_NodeFindCutsAll_pCutStore.nCuts == 1");

    // explore the cuts
    for  i = 0; i < spStatic_Ivy_NodeFindCutsAll_pCutStore.nCuts; i++  {
        // expand this cut
        pCut = spStatic_Ivy_NodeFindCutsAll_pCutStore.pCuts + i;
        if pCut.nSize == 0 {
            continue; }
        for  k = 0; k < pCut.nSize; k++  {
            pLeaf = Ivy_ManObj( p, pCut.pArray[k] );
            if Ivy_ObjIsCi(pLeaf) {
                continue; }
/*
            *pCutNew = *pCut;
            Ivy_NodeCutShrink( pCutNew, pLeaf->Id );
            if ( !Ivy_NodeCutExtend( pCutNew, Ivy_ObjFaninId0(pLeaf) ) )
                continue;
            if ( Ivy_ObjIsNode(pLeaf) && !Ivy_NodeCutExtend( pCutNew, Ivy_ObjFaninId1(pLeaf) ) )
                continue;
            Ivy_NodeCutHash( pCutNew );
*/
            iLeaf0 = Ivy_ObjId( Ivy_ObjRealFanin(Ivy_ObjFanin0(pLeaf)) );
            iLeaf1 = Ivy_ObjId( Ivy_ObjRealFanin(Ivy_ObjFanin1(pLeaf)) );
//            if ( iLeaf0 == iLeaf1 ) // strange situation observed on Jan 18, 2007
//                continue;
            if !(Ivy_NodeCutPrescreen( pCut, iLeaf0, iLeaf1 ) != 0) {
                continue; }
            if iLeaf0 > iLeaf1 {
                Ivy_NodeCutDeriveNew( pCut, pCutNew, pCut.pArray[k], iLeaf1, iLeaf0 );
            } else {
                Ivy_NodeCutDeriveNew( pCut, pCutNew, pCut.pArray[k], iLeaf0, iLeaf1 ); }
            Ivy_NodeCutFindOrAddFilter( spStatic_Ivy_NodeFindCutsAll_pCutStore, pCutNew );
            if spStatic_Ivy_NodeFindCutsAll_pCutStore.nCuts == IVY_CUT_LIMIT {
                break; }
        }
        if spStatic_Ivy_NodeFindCutsAll_pCutStore.nCuts == IVY_CUT_LIMIT {
            break; }
    }
    Ivy_NodeCompactCuts( spStatic_Ivy_NodeFindCutsAll_pCutStore );
//    Ivy_NodePrintCuts( pCutStore );
    return spStatic_Ivy_NodeFindCutsAll_pCutStore;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ivy_ManTestCutsAll(p *Ivy_Man_t) {
    var pObj *Ivy_Obj_t;
    var i, nCutsCut, nCutsTotal, nNodeTotal, nNodeOver int32;
    var clk abctime  = Abc_Clock();
    nNodeTotal = nNodeOver = 0;
    nCutsTotal = -Ivy_ManNodeNum(p);
    // inlining of macro Ivy_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Ivy_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if !Ivy_ObjIsNode(pObj) {
            continue; }
        nCutsCut    = Ivy_NodeFindCutsAll( p, pObj, 5 ).nCuts;
        nCutsTotal += nCutsCut;
        nNodeOver  += (nCutsCut == IVY_CUT_LIMIT);
        nNodeTotal++;
    }
    }
    fmt.Printf("Total cuts = %6d. Trivial = %6d.   Nodes = %6d. Satur = %6d.  ", nCutsTotal, Ivy_ManPiNum(p) + Ivy_ManNodeNum(p), nNodeTotal, nNodeOver);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


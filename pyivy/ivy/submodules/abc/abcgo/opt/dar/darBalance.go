package dar

import "fmt"

/**CFile****************************************************************

  FileName    [darBalance.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting.]

  Synopsis    [Algebraic AIG balancing.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: darBalance.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "darInt.h"
//#include "misc/tim/tim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

//#define USE_LUTSIZE_BALANCE

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Uniqifies the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ObjCompareLits(pp1 **Aig_Obj_t, pp2 **Aig_Obj_t) int32 {
    var Diff int32  = Aig_ObjToLit(*pp1) - Aig_ObjToLit(*pp2);
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    return 0; 
}
func Dar_BalanceUniqify(pObj *Aig_Obj_t, vNodes *Vec_Ptr_t, fExor int32) {
    var pTemp, pTempNext *Aig_Obj_t;
    var i, k int32;
    // sort the nodes by their literal
    Vec_PtrSort( vNodes, (int (*)())Dar_ObjCompareLits );
    // remove duplicates
    k = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pTemp) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if i + 1 == Vec_PtrSize(vNodes) {
            Vec_PtrWriteEntry( vNodes, k++, pTemp );
            break;
        }
        pTempNext = (Aig_Obj_t *)Vec_PtrEntry( vNodes, i+1 );
        if !(fExor != 0) && pTemp == Aig_Not(pTempNext) { // pos_lit & neg_lit = 0
        
            Vec_PtrClear( vNodes );
            return;
        }
        if pTemp != pTempNext {  // save if different
            Vec_PtrWriteEntry( vNodes, k++, pTemp );
        } else if fExor != 0 { // in case of XOR, remove identical
            i++; }
    }
    Vec_PtrShrink( vNodes, k );
    if Vec_PtrSize(vNodes) < 2 {
        return; }
    // check that there is no duplicates
    pTemp = (Aig_Obj_t *)Vec_PtrEntry( vNodes, 0 );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vNodes)) && (((pTempNext) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        assert(pTemp != pTempNext, "pTemp != pTempNext");
        pTemp = pTempNext;
    }
}

/**Function*************************************************************

  Synopsis    [Collects the nodes of the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_BalanceCone_rec(pRoot *Aig_Obj_t, pObj *Aig_Obj_t, vSuper *Vec_Ptr_t) {
    if pObj != pRoot && (Aig_IsComplement(pObj) || Aig_ObjType(pObj) != Aig_ObjType(pRoot) || Aig_ObjRefs(pObj) > 1 || Vec_PtrSize(vSuper) > 10000) {
        Vec_PtrPush( vSuper, pObj );
    } else {
        assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        // go through the branches
        Dar_BalanceCone_rec( pRoot, Aig_ObjReal_rec( Aig_ObjChild0(pObj) ), vSuper );
        Dar_BalanceCone_rec( pRoot, Aig_ObjReal_rec( Aig_ObjChild1(pObj) ), vSuper );
    }
}
func Dar_BalanceCone(pObj *Aig_Obj_t, vStore *Vec_Vec_t, Level int32) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    // extend the storage
    if Vec_VecSize( vStore ) <= Level {
        Vec_VecPush( vStore, Level, 0 ); }
    // get the temporary array of nodes
    vNodes = Vec_VecEntry( vStore, Level );
    Vec_PtrClear( vNodes );
    // collect the nodes in the implication supergate
    Dar_BalanceCone_rec( pObj, pObj, vNodes );
    // remove duplicates
    Dar_BalanceUniqify( pObj, vNodes, Aig_ObjIsExor(pObj) );
    return vNodes;
}


/**Function*************************************************************

  Synopsis    [Collects the nodes of the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
int Dar_BalanceCone_rec( Aig_Obj_t * pRoot, Aig_Obj_t * pObj, Vec_Ptr_t * vSuper )
{
    int RetValue1, RetValue2, i;
    // check if the node is visited
    if ( Aig_Regular(pObj)->fMarkB )
    {
        if ( Aig_ObjIsExor(pRoot) )
        {
            assert( !Aig_IsComplement(pObj) );
            // check if the node occurs in the same polarity
            Vec_PtrRemove( vSuper, pObj );
            Aig_Regular(pObj)->fMarkB = 0;
//printf( " Duplicated EXOR input!!! " );
            return 1;
        }
        else
        {
            // check if the node occurs in the same polarity
            for ( i = 0; i < vSuper->nSize; i++ )
                if ( vSuper->pArray[i] == pObj )
                    return 1;
            // check if the node is present in the opposite polarity
            for ( i = 0; i < vSuper->nSize; i++ )
                if ( vSuper->pArray[i] == Aig_Not(pObj) )
                    return -1;
        }
        assert( 0 );
        return 0;
    }
    // if the new node is complemented or a PI, another gate begins
    if ( pObj != pRoot && (Aig_IsComplement(pObj) || Aig_ObjType(pObj) != Aig_ObjType(pRoot) || Aig_ObjRefs(pObj) > 1 || Vec_PtrSize(vSuper) > 10000) )
    {
        Vec_PtrPush( vSuper, pObj );
        Aig_Regular(pObj)->fMarkB = 1;
        return 0;
    }
    assert( !Aig_IsComplement(pObj) );
    assert( Aig_ObjIsNode(pObj) );
    // go through the branches
    RetValue1 = Dar_BalanceCone_rec( pRoot, Aig_ObjReal_rec( Aig_ObjChild0(pObj) ), vSuper );
    RetValue2 = Dar_BalanceCone_rec( pRoot, Aig_ObjReal_rec( Aig_ObjChild1(pObj) ), vSuper );
    if ( RetValue1 == -1 || RetValue2 == -1 )
        return -1;
    // return 1 if at least one branch has a duplicate
    return RetValue1 || RetValue2;
}
Vec_Ptr_t * Dar_BalanceCone( Aig_Obj_t * pObj, Vec_Vec_t * vStore, int Level )
{
    Vec_Ptr_t * vNodes;
    int RetValue, i;
    assert( !Aig_IsComplement(pObj) );
    // extend the storage
    if ( Vec_VecSize( vStore ) <= Level )
        Vec_VecPush( vStore, Level, 0 );
    // get the temporary array of nodes
    vNodes = Vec_VecEntry( vStore, Level );
    Vec_PtrClear( vNodes );
    // collect the nodes in the implication supergate
    RetValue = Dar_BalanceCone_rec( pObj, pObj, vNodes );
    assert( RetValue != 0 || vNodes->nSize > 1 );
    // unmark the visited nodes
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
        Aig_Regular(pObj)->fMarkB = 0;
    // if we found the node and its complement in the same implication supergate, 
    // return empty set of nodes (meaning that we should use constant-0 node)
    if ( RetValue == -1 )
        vNodes->nSize = 0;
    return vNodes;
}
*/

/**Function*************************************************************

  Synopsis    [Finds the left bound on the next candidate to be paired.]

  Description [The nodes in the array are in the decreasing order of levels. 
  The last node in the array has the smallest level. By default it would be paired 
  with the next node on the left. However, it may be possible to pair it with some
  other node on the left, in such a way that the new node is shared. This procedure
  finds the index of the left-most node, which can be paired with the last node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_BalanceFindLeft(vSuper *Vec_Ptr_t) int32 {
    var pObjRight, pObjLeft *Aig_Obj_t;
    var Current int32;
    // if two or less nodes, pair with the first
    if Vec_PtrSize(vSuper) < 3 {
        return 0; }
    // set the pointer to the one before the last
    Current = Vec_PtrSize(vSuper) - 2;
    pObjRight = (Aig_Obj_t *)Vec_PtrEntry( vSuper, Current );
    // go through the nodes to the left of this one
    for  Current--; Current >= 0; Current--  {
        // get the next node on the left
        pObjLeft = (Aig_Obj_t *)Vec_PtrEntry( vSuper, Current );
        // if the level of this node is different, quit the loop
        if Aig_ObjLevel(Aig_Regular(pObjLeft)) != Aig_ObjLevel(Aig_Regular(pObjRight)) {
            break; }
    }
    Current++;    
    // get the node, for which the equality holds
    pObjLeft = (Aig_Obj_t *)Vec_PtrEntry( vSuper, Current );
    assert(Aig_ObjLevel(Aig_Regular(pObjLeft)) == Aig_ObjLevel(Aig_Regular(pObjRight)), "Aig_ObjLevel(Aig_Regular(pObjLeft)) == Aig_ObjLevel(Aig_Regular(pObjRight))");
    return Current;
}

/**Function*************************************************************

  Synopsis    [Moves closer to the end the node that is best for sharing.]

  Description [If there is no node with sharing, randomly chooses one of 
  the legal nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_BalancePermute(p *Aig_Man_t, vSuper *Vec_Ptr_t, LeftBound int32, fExor int32) {
    var pObj1, pObj2, pObj3, pGhost *Aig_Obj_t;
    var RightBound, i int32;
    // get the right bound
    RightBound = Vec_PtrSize(vSuper) - 2;
    assert(LeftBound <= RightBound, "LeftBound <= RightBound");
    if LeftBound == RightBound {
        return; }
    // get the two last nodes
    pObj1 = (Aig_Obj_t *)Vec_PtrEntry( vSuper, RightBound + 1 );
    pObj2 = (Aig_Obj_t *)Vec_PtrEntry( vSuper, RightBound     );
    if Aig_Regular(pObj1) == p.pConst1 || Aig_Regular(pObj2) == p.pConst1 || Aig_Regular(pObj1) == Aig_Regular(pObj2) {
        return; }
    // find the first node that can be shared
    for  i = RightBound; i >= LeftBound; i--  {
        pObj3 = (Aig_Obj_t *)Vec_PtrEntry( vSuper, i );
        if Aig_Regular(pObj3) == p.pConst1 {
            Vec_PtrWriteEntry( vSuper, i,          pObj2 );
            Vec_PtrWriteEntry( vSuper, RightBound, pObj3 );
            return;
        }
        if Aig_Regular(pObj1) == Aig_Regular(pObj3) {
            if pObj3 == pObj2 {
                return; }
            Vec_PtrWriteEntry( vSuper, i,          pObj2 );
            Vec_PtrWriteEntry( vSuper, RightBound, pObj3 );
            return;
        }
        pGhost = Aig_ObjCreateGhost( p, pObj1, pObj3, fExor? AIG_OBJ_EXOR : AIG_OBJ_AND );
        if Aig_TableLookup( p, pGhost ) {
            if pObj3 == pObj2 {
                return; }
            Vec_PtrWriteEntry( vSuper, i,          pObj2 );
            Vec_PtrWriteEntry( vSuper, RightBound, pObj3 );
            return;
        }
    }
/*
    // we did not find the node to share, randomize choice
    {
        int Choice = Aig_ManRandom(0) % (RightBound - LeftBound + 1);
        pObj3 = Vec_PtrEntry( vSuper, LeftBound + Choice );
        if ( pObj3 == pObj2 )
            return;
        Vec_PtrWriteEntry( vSuper, LeftBound + Choice, pObj2 );
        Vec_PtrWriteEntry( vSuper, RightBound,         pObj3 );
    }
*/
}

/**Function*************************************************************

  Synopsis    [Procedure used for sorting the nodes in decreasing order of levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeCompareLevelsDecrease(pp1 **Aig_Obj_t, pp2 **Aig_Obj_t) int32 {
    var Diff int32  = Aig_ObjLevel(Aig_Regular(*pp1)) - Aig_ObjLevel(Aig_Regular(*pp2));
    if Diff > 0 {
        return -1; }
    if Diff < 0 { 
        return 1; }
    Diff = Aig_ObjId(Aig_Regular(*pp1)) - Aig_ObjId(Aig_Regular(*pp2));
    if Diff > 0 {
        return -1; }
    if Diff < 0 { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Inserts a new node in the order by levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_BalancePushUniqueOrderByLevel(vStore *Vec_Ptr_t, pObj *Aig_Obj_t, fExor int32) {
    var pObj1, pObj2 *Aig_Obj_t;
    var i int32;
    if Vec_PtrPushUnique(vStore, pObj) {
        if fExor != 0 {
            Vec_PtrRemove(vStore, pObj); }
        return;
    }
    // find the p of the node
    for  i = vStore.nSize-1; i > 0; i--  {
        pObj1 = (Aig_Obj_t *)vStore.pArray[i  ];
        pObj2 = (Aig_Obj_t *)vStore.pArray[i-1];
        if Aig_ObjLevel(Aig_Regular(pObj1)) <= Aig_ObjLevel(Aig_Regular(pObj2)) {
            break; }
        vStore.pArray[i  ] = pObj2;
        vStore.pArray[i-1] = pObj1;
    }
}

/**Function*************************************************************

  Synopsis    [Builds implication supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_BalanceBuildSuper(p *Aig_Man_t, vSuper *Vec_Ptr_t, Type Aig_Type_t, fUpdateLevel int32) *Aig_Obj_t {
    var pObj1, pObj2 *Aig_Obj_t;
    var LeftBound int32;
    assert(vSuper.nSize > 1, "vSuper.nSize > 1");
    // sort the new nodes by level in the decreasing order
    Vec_PtrSort( vSuper, (int (*)(void))Aig_NodeCompareLevelsDecrease );
    // balance the nodes
    for  vSuper.nSize > 1  {
        // find the left bound on the node to be paired
        if !(fUpdateLevel != 0) {
            LeftBound = 0;
        } else {
            LeftBound = Dar_BalanceFindLeft( vSuper );
        }
        // find the node that can be shared (if no such node, randomize choice)
        Dar_BalancePermute( p, vSuper, LeftBound, Type == AIG_OBJ_EXOR );
        // pull out the last two nodes
        pObj1 = (Aig_Obj_t *)Vec_PtrPop(vSuper);
        pObj2 = (Aig_Obj_t *)Vec_PtrPop(vSuper);
        Dar_BalancePushUniqueOrderByLevel( vSuper, Aig_Oper(p, pObj1, pObj2, Type), Type == AIG_OBJ_EXOR );
    }
    return (Aig_Obj_t *)Vec_PtrEntry(vSuper, 0);
}


/**Function*************************************************************

  Synopsis    [Returns affective support size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_BaseSize(p *Aig_Man_t, pObj *Aig_Obj_t, nLutSize int32) int32 {
    var nBaseSize int32;
    pObj = Aig_Regular(pObj);
    if Aig_ObjIsConst1(pObj) {
        return 0; }
    if Aig_ObjLevel(pObj) >= nLutSize {
        return 1; }
    nBaseSize = Aig_SupportSize( p, pObj );
    if nBaseSize >= nLutSize {
        return 1; }
    return nBaseSize;
}

/**Function*************************************************************

  Synopsis    [Builds implication supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_BalanceBuildSuperTop(p *Aig_Man_t, vSuper *Vec_Ptr_t, Type Aig_Type_t, fUpdateLevel int32, nLutSize int32) *Aig_Obj_t {
    var vSubset *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i, nBaseSizeAll, nBaseSize int32;
    assert(vSuper.nSize > 1, "vSuper.nSize > 1");
    // sort the new nodes by level in the decreasing order
    Vec_PtrSort( vSuper, (int (*)(void))Aig_NodeCompareLevelsDecrease );
    // add one LUT at a time
    for  Vec_PtrSize(vSuper) > 1  {
        // isolate the group of nodes with nLutSize inputs
        nBaseSizeAll = 0;
        vSubset = Vec_PtrAlloc( nLutSize );
        // inlining of macro Vec_PtrForEachEntryReverse
        for  i = Vec_PtrSize(vSuper) - 1; (i >= 0) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i--  {
            nBaseSize = Aig_BaseSize( p, pObj, nLutSize );
            if nBaseSizeAll + nBaseSize > nLutSize && Vec_PtrSize(vSubset) > 1 {
                break; }
            nBaseSizeAll += nBaseSize;
            Vec_PtrPush( vSubset, pObj );
        }
        // remove them from vSuper
        Vec_PtrShrink( vSuper, Vec_PtrSize(vSuper) - Vec_PtrSize(vSubset) );
        // create the new supergate
        pObj = Dar_BalanceBuildSuper( p, vSubset, Type, fUpdateLevel );
        Vec_PtrFree( vSubset );
        // add the new output
        Dar_BalancePushUniqueOrderByLevel( vSuper, pObj, Type == AIG_OBJ_EXOR );
    }
    return (Aig_Obj_t *)Vec_PtrEntry(vSuper, 0);
}

/**Function*************************************************************

  Synopsis    [Returns the new node constructed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_Balance_rec(pNew *Aig_Man_t, pObjOld *Aig_Obj_t, vStore *Vec_Vec_t, Level int32, fUpdateLevel int32) *Aig_Obj_t {
    var pObjNew *Aig_Obj_t;
    var vSuper *Vec_Ptr_t;
    var i int32;
    assert(!Aig_IsComplement(pObjOld), "!Aig_IsComplement(pObjOld)");
    assert(!Aig_ObjIsBuf(pObjOld), "!Aig_ObjIsBuf(pObjOld)");
    // return if the result is known
    if pObjOld.pData {
        return (Aig_Obj_t *)pObjOld.pData; }
    assert(Aig_ObjIsNode(pObjOld), "Aig_ObjIsNode(pObjOld)");
    // get the implication supergate
    vSuper = Dar_BalanceCone( pObjOld, vStore, Level );
    // check if supergate contains two nodes in the opposite polarity
    if vSuper.nSize == 0 {
        return (Aig_Obj_t *)(pObjOld.pData = Aig_ManConst0(pNew)); }
    // for each old node, derive the new well-balanced node
    for  i = 0; i < Vec_PtrSize(vSuper); i++  {
        pObjNew = Dar_Balance_rec( pNew, Aig_Regular((Aig_Obj_t *)vSuper.pArray[i]), vStore, Level + 1, fUpdateLevel );
        if pObjNew == nil {
            return nil; }
        vSuper.pArray[i] = Aig_NotCond( pObjNew, Aig_IsComplement((Aig_Obj_t *)vSuper.pArray[i]) );
    }
    // check for exactly one node
    if vSuper.nSize == 1 {
        return (Aig_Obj_t *)Vec_PtrEntry(vSuper, 0); }
    // build the supergate
#ifdef USE_LUTSIZE_BALANCE
    pObjNew = Dar_BalanceBuildSuperTop( pNew, vSuper, Aig_ObjType(pObjOld), fUpdateLevel, 6 );
#else
    pObjNew = Dar_BalanceBuildSuper( pNew, vSuper, Aig_ObjType(pObjOld), fUpdateLevel );
#endif
    if pNew.Time2Quit && !(Aig_Regular(pObjNew).Id & 255) && Abc_Clock() > pNew.Time2Quit {
        return nil; }
    // make sure the balanced node is not assigned
//    assert( pObjOld->Level >= Aig_Regular(pObjNew)->Level );
    assert(pObjOld.pData == nil, "pObjOld.pData == nil");
    return (Aig_Obj_t *)(pObjOld.pData = pObjNew);
}

/**Function*************************************************************

  Synopsis    [Performs algebraic balancing of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManBalance(p *Aig_Man_t, fUpdateLevel int32) *Aig_Man_t {
    var pNew *Aig_Man_t;
    var pObj, pDriver, pObjNew *Aig_Obj_t;
    var vStore *Vec_Vec_t;
    var i int32;
    assert(Aig_ManVerifyTopoOrder(p), "Aig_ManVerifyTopoOrder(p)");
    // create the new manager 
    pNew = Aig_ManStart( Aig_ManObjNumMax(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.nAsserts = p.nAsserts;
    pNew.nConstrs = p.nConstrs;
    pNew.nBarBufs = p.nBarBufs;
    pNew.Time2Quit = p.Time2Quit;
    if p.vFlopNums {
        pNew.vFlopNums = Vec_IntDup( p.vFlopNums ); }
    // map the PI nodes
    Aig_ManCleanData( p );
    Aig_ManConst1(p).pData = Aig_ManConst1(pNew);
    vStore = Vec_VecAlloc( 50 );
    if p.pManTime != nil {
        var arrTime float32;
        Tim_ManIncrementTravId( (Tim_Man_t *)p.pManTime );
        Aig_ManSetCioIds( p );
        // inlining of macro Aig_ManForEachObj
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
            if Aig_ObjIsNode(pObj) || Aig_ObjIsConst1(pObj) {
                continue; }
            if Aig_ObjIsCi(pObj) {
                // copy the PI
                pObjNew = Aig_ObjCreateCi(pNew); 
                pObj.pData = pObjNew;
                // set the arrival time of the new PI
                arrTime = Tim_ManGetCiArrival( (Tim_Man_t *)p.pManTime, Aig_ObjCioId(pObj) );
                pObjNew.Level = int32(arrTime);
            
            } else if Aig_ObjIsCo(pObj) {
                // perform balancing
                pDriver = Aig_ObjReal_rec( Aig_ObjChild0(pObj) );
                pObjNew = Dar_Balance_rec( pNew, Aig_Regular(pDriver), vStore, 0, fUpdateLevel );
                if pObjNew == nil {
                    Vec_VecFree( vStore );
                    Aig_ManStop( pNew );
                    return nil;
                }
                pObjNew = Aig_NotCond( pObjNew, Aig_IsComplement(pDriver) );
                // save arrival time of the output
                arrTime = float32(Aig_Regular(pObjNew).Level);
                Tim_ManSetCoArrival( (Tim_Man_t *)p.pManTime, Aig_ObjCioId(pObj), arrTime );
                // create PO
                pObjNew = Aig_ObjCreateCo( pNew, pObjNew );
            
            } else {
                assert(false, "0"); }
        }
        }
        Aig_ManCleanCioIds( p );
        pNew.pManTime = Tim_ManDup( (Tim_Man_t *)p.pManTime, 0 );
    
    } else {
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
            pObjNew = Aig_ObjCreateCi(pNew); 
            pObjNew.Level = pObj.Level;
            pObj.pData = pObjNew;
        }
        if p.nBarBufs == 0 {
            // inlining of macro Aig_ManForEachCo
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                pDriver = Aig_ObjReal_rec( Aig_ObjChild0(pObj) );
                pObjNew = Dar_Balance_rec( pNew, Aig_Regular(pDriver), vStore, 0, fUpdateLevel );
                if pObjNew == nil {
                    Vec_VecFree( vStore );
                    Aig_ManStop( pNew );
                    return nil;
                }
                pObjNew = Aig_NotCond( pObjNew, Aig_IsComplement(pDriver) );
                pObjNew = Aig_ObjCreateCo( pNew, pObjNew );
            }
        
        } else {
            var vLits *Vec_Ptr_t  = Vec_PtrStart( Aig_ManCoNum(p) );
            // inlining of macro Aig_ManForEachCo
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                var k int32;
                if i < p.nBarBufs {
                    k = Aig_ManCoNum(p) - p.nBarBufs + i;
                } else {
                    k = i - p.nBarBufs;
                }
                pObj = Aig_ManCo( p, k );
                pDriver = Aig_ObjReal_rec( Aig_ObjChild0(pObj) );
                pObjNew = Dar_Balance_rec( pNew, Aig_Regular(pDriver), vStore, 0, fUpdateLevel );
                if pObjNew == nil {
                    Vec_VecFree( vStore );
                    Aig_ManStop( pNew );
                    return nil;
                }
                pObjNew = Aig_NotCond( pObjNew, Aig_IsComplement(pDriver) );
                Vec_PtrWriteEntry( vLits, k, pObjNew );
                if i < p.nBarBufs {
                    Aig_ManCi(pNew, Aig_ManCiNum(p) - p.nBarBufs + i).Level = Aig_Regular(pObjNew).Level; }
            }
            // inlining of macro Aig_ManForEachCo
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
                Aig_ObjCreateCo( pNew, (Aig_Obj_t *)Vec_PtrEntry(vLits, i) );
            }
            Vec_PtrFree(vLits);
        }
    }
    Vec_VecFree( vStore );
    // remove dangling nodes
    Aig_ManCleanup( pNew );
    Aig_ManSetRegNum( pNew, Aig_ManRegNum(p) );
    // check the resulting AIG
    if !Aig_ManCheck(pNew) {
        fmt.Printf("Dar_ManBalance(): The check has failed.\n"); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Reproduces script "compress2".]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_ManBalanceXor(pAig *Aig_Man_t, fExor int32, fUpdateLevel int32, fVerbose int32) *Aig_Man_t {
    var pAigXor, pRes *Aig_Man_t;
    if fExor != 0 {
        pAigXor = Aig_ManDupExor( pAig );
        if fVerbose != 0 {
            Dar_BalancePrintStats( pAigXor ); }
        pRes = Dar_ManBalance( pAigXor, fUpdateLevel );
        Aig_ManStop( pAigXor );
    
    } else {
        pRes = Dar_ManBalance( pAig, fUpdateLevel );
    }
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Inserts a new node in the order by levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_BalancePrintStats(p *Aig_Man_t) {
    var vSuper *Vec_Ptr_t;
    var pObj, pTemp *Aig_Obj_t;
    var i, k int32;
    if Aig_ManExorNum(p) == 0 {
        fmt.Printf("There is no EXOR gates.\n");
        return;
    }
    // inlining of macro Aig_ManForEachExor
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsExor(pObj) {} else {
        Aig_ObjFanin0(pObj).fMarkA = 1;
        Aig_ObjFanin1(pObj).fMarkA = 1;
        assert(!Aig_ObjFaninC0(pObj), "!Aig_ObjFaninC0(pObj)");
        assert(!Aig_ObjFaninC1(pObj), "!Aig_ObjFaninC1(pObj)");
    }
    }
    vSuper = Vec_PtrAlloc( 1000 );
    // inlining of macro Aig_ManForEachExor
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsExor(pObj) {} else {
        if pObj.fMarkA && pObj.nRefs == 1 {
            continue; }
        Vec_PtrClear( vSuper );
        Dar_BalanceCone_rec( pObj, pObj, vSuper );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vSuper)) && (((pTemp) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, k)), 1); k++  {
            pTemp.fMarkB = 0;
        }
        if Vec_PtrSize(vSuper) < 3 {
            continue; }
        fmt.Printf("  %d(", Vec_PtrSize(vSuper));
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vSuper)) && (((pTemp) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, k)), 1); k++  {
            fmt.Printf(" %d", pTemp.Level);
        }
        fmt.Printf(" )");
    }
    }
    Vec_PtrFree( vSuper );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pObj.fMarkA = 0; }
    }
    fmt.Printf("\n");
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


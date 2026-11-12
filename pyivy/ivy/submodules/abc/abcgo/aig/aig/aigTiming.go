package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigTiming.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Incremental updating of direct/reverse AIG levels.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigTiming.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the reverse level of the node.]

  Description [The reverse level is the level of the node in reverse
  topological order, starting from the COs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjReverseLevel(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    assert(p.vLevelR, "p.vLevelR");
    Vec_IntFillExtra( p.vLevelR, pObj.Id + 1, 0 );
    return Vec_IntEntry(p.vLevelR, pObj.Id);
}

/**Function*************************************************************

  Synopsis    [Sets the reverse level of the node.]

  Description [The reverse level is the level of the node in reverse
  topological order, starting from the COs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjSetReverseLevel(p *Aig_Man_t, pObj *Aig_Obj_t, LevelR int32) {
    assert(p.vLevelR, "p.vLevelR");
    Vec_IntFillExtra( p.vLevelR, pObj.Id + 1, 0 );
    Vec_IntWriteEntry( p.vLevelR, pObj.Id, LevelR );
}

/**Function*************************************************************

  Synopsis    [Resets reverse level of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjClearReverseLevel(p *Aig_Man_t, pObj *Aig_Obj_t) {
    Aig_ObjSetReverseLevel( p, pObj, 0 );
}

/**Function*************************************************************

  Synopsis    [Returns required level of the node.]

  Description [Converts the reverse levels of the node into its required 
  level as follows: ReqLevel(Node) = MaxLevels(Ntk) + 1 - LevelR(Node).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjRequiredLevel(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    assert(p.vLevelR, "p.vLevelR");
    return p.nLevelMax + 1 - Aig_ObjReverseLevel(p, pObj);
}

/**Function*************************************************************

  Synopsis    [Computes the reverse level of the node using its fanout levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjReverseLevelNew(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    var pFanout *Aig_Obj_t;
    var i int32;
    var iFanout int32 = -1;
    var LevelCur int32;
    var Level int32;
    // inlining of macro Aig_ObjForEachFanout
    for  assert(p.pFanData, "p.pFanData"), i = 0; (i < int32((pObj).nRefs)) &&                         (((iFanout) = i? Aig_ObjFanoutNext(p, iFanout) : Aig_ObjFanout0Int(p, pObj.Id)), 1) &&           (((pFanout) = Aig_ManObj(p, iFanout>>1)), 1); i++  {
        LevelCur = Aig_ObjReverseLevel( p, pFanout );
        Level = Abc_MaxInt( Level, LevelCur );
    }
    return Level + 1;
}

/**Function*************************************************************

  Synopsis    [Prepares for the computation of required levels.]

  Description [This procedure should be called before the required times
  are used. It starts internal data structures, which records the level 
  from the COs of the network nodes in reverse topologogical order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManStartReverseLevels(p *Aig_Man_t, nMaxLevelIncrease int32) {
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    assert(p.pFanData != nil, "p.pFanData != nil");
    assert(p.vLevelR == nil, "p.vLevelR == nil");
    // remember the maximum number of direct levels
    p.nLevelMax = Aig_ManLevels(p) + nMaxLevelIncrease;
    // start the reverse levels
    p.vLevelR = Vec_IntAlloc( 0 );
    Vec_IntFill( p.vLevelR, Aig_ManObjNumMax(p), 0 );
    // compute levels in reverse topological order
    vNodes = Aig_ManDfsReverse( p );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        assert(pObj.fMarkA == 0, "pObj.fMarkA == 0");
        Aig_ObjSetReverseLevel( p, pObj, Aig_ObjReverseLevelNew(p, pObj) );
    }
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Cleans the data structures used to compute required levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManStopReverseLevels(p *Aig_Man_t) {
    assert(p.vLevelR != nil, "p.vLevelR != nil");
    Vec_IntFree( p.vLevelR );
    p.vLevelR = nil;
    p.nLevelMax = 0;

}

/**Function*************************************************************

  Synopsis    [Incrementally updates level of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManUpdateLevel(p *Aig_Man_t, pObjNew *Aig_Obj_t) {
    var pFanout, pTemp *Aig_Obj_t;
    var iFanout int32 = -1;
    var LevelOld int32;
    var Lev int32;
    var k int32;
    var m int32;
    assert(p.pFanData != nil, "p.pFanData != nil");
    assert(Aig_ObjIsNode(pObjNew), "Aig_ObjIsNode(pObjNew)");
    // allocate level if needed
    if p.vLevels == nil {
        p.vLevels = Vec_VecAlloc( Aig_ManLevels(p) + 8 ); }
    // check if level has changed
    LevelOld = Aig_ObjLevel(pObjNew);
    if LevelOld == Aig_ObjLevelNew(pObjNew) {
        return; }
    // start the data structure for level update
    // we cannot fail to visit a node when using this structure because the 
    // nodes are stored by their _old_ levels, which are assumed to be correct
    Vec_VecClear( p.vLevels );
    Vec_VecPush( p.vLevels, LevelOld, pObjNew );
    pObjNew.fMarkA = 1;
    // recursively update level
    // inlining of macro Vec_VecForEachEntryStart
    // inlining of macro Vec_PtrForEachEntry
    for  Lev = LevelOld; Lev < Vec_VecSize(p.vLevels); Lev++  {                                               for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(p.vLevels, Lev))) && (((pTemp) = (Aig_Obj_t *)Vec_PtrEntry(Vec_VecEntry(p.vLevels, Lev), k)), 1); k++  {
        pTemp.fMarkA = 0;
        assert(Aig_ObjLevel(pTemp) == Lev, "Aig_ObjLevel(pTemp) == Lev");
        pTemp.Level = Aig_ObjLevelNew(pTemp);
        // if the level did not change, no need to check the fanout levels
        if Aig_ObjLevel(pTemp) == Lev {
            continue; }
        // schedule fanout for level update
        // inlining of macro Aig_ObjForEachFanout
        for  assert(p.pFanData, "p.pFanData"), m = 0; (m < int32((pTemp).nRefs)) &&                         (((iFanout) = m? Aig_ObjFanoutNext(p, iFanout) : Aig_ObjFanout0Int(p, pTemp.Id)), 1) &&           (((pFanout) = Aig_ManObj(p, iFanout>>1)), 1); m++  {
            if Aig_ObjIsNode(pFanout) && !pFanout.fMarkA {
                assert(Aig_ObjLevel(pFanout) >= Lev, "Aig_ObjLevel(pFanout) >= Lev");
                Vec_VecPush( p.vLevels, Aig_ObjLevel(pFanout), pFanout );
                pFanout.fMarkA = 1;
            }
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Incrementally updates level of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManUpdateReverseLevel(p *Aig_Man_t, pObjNew *Aig_Obj_t) {
    var pFanin, pTemp *Aig_Obj_t;
    var LevelOld, LevFanin, Lev, k int32;
    assert(p.vLevelR != nil, "p.vLevelR != nil");
    assert(Aig_ObjIsNode(pObjNew), "Aig_ObjIsNode(pObjNew)");
    // allocate level if needed
    if p.vLevels == nil {
        p.vLevels = Vec_VecAlloc( Aig_ManLevels(p) + 8 ); }
    // check if level has changed
    LevelOld = Aig_ObjReverseLevel(p, pObjNew);
    if LevelOld == Aig_ObjReverseLevelNew(p, pObjNew) {
        return; }
    // start the data structure for level update
    // we cannot fail to visit a node when using this structure because the 
    // nodes are stored by their _old_ levels, which are assumed to be correct
    Vec_VecClear( p.vLevels );
    Vec_VecPush( p.vLevels, LevelOld, pObjNew );
    pObjNew.fMarkA = 1;
    // recursively update level
    // inlining of macro Vec_VecForEachEntryStart
    // inlining of macro Vec_PtrForEachEntry
    for  Lev = LevelOld; Lev < Vec_VecSize(p.vLevels); Lev++  {                                               for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(p.vLevels, Lev))) && (((pTemp) = (Aig_Obj_t *)Vec_PtrEntry(Vec_VecEntry(p.vLevels, Lev), k)), 1); k++  {
        pTemp.fMarkA = 0;
        LevelOld = Aig_ObjReverseLevel(p, pTemp); 
        assert(LevelOld == Lev, "LevelOld == Lev");
        Aig_ObjSetReverseLevel( p, pTemp, Aig_ObjReverseLevelNew(p, pTemp) );
        // if the level did not change, to need to check the fanout levels
        if Aig_ObjReverseLevel(p, pTemp) == Lev {
            continue; }
        // schedule fanins for level update
        pFanin = Aig_ObjFanin0(pTemp);
        if Aig_ObjIsNode(pFanin) && !pFanin.fMarkA {
            LevFanin = Aig_ObjReverseLevel( p, pFanin );
            assert(LevFanin >= Lev, "LevFanin >= Lev");
            Vec_VecPush( p.vLevels, LevFanin, pFanin );
            pFanin.fMarkA = 1;
        }
        pFanin = Aig_ObjFanin1(pTemp);
        if Aig_ObjIsNode(pFanin) && !pFanin.fMarkA {
            LevFanin = Aig_ObjReverseLevel( p, pFanin );
            assert(LevFanin >= Lev, "LevFanin >= Lev");
            Vec_VecPush( p.vLevels, LevFanin, pFanin );
            pFanin.fMarkA = 1;
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Verifies direct level of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManVerifyLevel(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    assert(p.pFanData, "p.pFanData");
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else
        if Aig_ObjLevel(pObj) != Aig_ObjLevelNew(pObj) {
            fmt.Printf("Level of node %6d should be %4d instead of %4d.\n", pObj.Id, Aig_ObjLevelNew(pObj), Aig_ObjLevel(pObj));
            Counter++;
        }
    }
    if Counter != 0 {
    fmt.Printf("Levels of %d nodes are incorrect.\n", Counter); }
}

/**Function*************************************************************

  Synopsis    [Verifies reverse level of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManVerifyReverseLevel(p *Aig_Man_t) {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    assert(p.vLevelR, "p.vLevelR");
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else
        if Aig_ObjLevel(pObj) != Aig_ObjLevelNew(pObj) {
            fmt.Printf("Reverse level of node %6d should be %4d instead of %4d.\n", pObj.Id, Aig_ObjReverseLevelNew(p, pObj), Aig_ObjReverseLevel(p, pObj));
            Counter++;
        }
    }
    if Counter != 0 {
    fmt.Printf("Reverse levels of %d nodes are incorrect.\n", Counter); }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


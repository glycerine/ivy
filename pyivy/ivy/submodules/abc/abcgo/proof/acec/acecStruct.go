package acec

import "fmt"

/**CFile****************************************************************

  FileName    [acecStruct.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecStruct.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecWec.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_StructDetectXorRoots(p *Gia_Man_t) *Vec_Int_t {
    var vXors *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vXorIns *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var pFan0, pFan1, pObj *Gia_Obj_t; 
    var i, k, Entry int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) {
            continue; }
        Vec_IntPush( vXors, i );
        Vec_BitWriteEntry( vXorIns, Gia_ObjId(p, Gia_Regular(pFan0)), 1 );
        Vec_BitWriteEntry( vXorIns, Gia_ObjId(p, Gia_Regular(pFan1)), 1 );
    }
    }
    // collect XORs that not inputs of other XORs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXors)) && (((Entry) = Vec_IntEntry(vXors, i)), 1); i++  {
        if !Vec_BitEntry(vXorIns, Entry) {
            Vec_IntWriteEntry( vXors, k++, Entry ); }
    }
    Vec_IntShrink( vXors, k );
    Vec_BitFree( vXorIns );
    return vXors;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_StructAssignRanks(p *Gia_Man_t, vXorRoots *Vec_Int_t) *Vec_Int_t {
    var vDoubles *Vec_Int_t  = Vec_IntAlloc( 100 );
    var pFan0, pFan1, pObj *Gia_Obj_t; 
    var i int32;
    var k int32;
    var Fanins [2]int32;
    var Entry int32;
    var Rank int32;
    // map roots into their ranks
    var vRanks *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) ); 
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vXorRoots)) && (((Entry) = Vec_IntEntry(vXorRoots, i)), 1); i++  {
        Vec_IntWriteEntry( vRanks, Entry, i );
    }
    // map nodes into their ranks
    // inlining of macro Gia_ManForEachAndReverse
    for  i = p.nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p, i)); i--  {  if !Gia_ObjIsAnd(pObj) {} else {
        if !Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) {
            continue; }
        Rank = Vec_IntEntry( vRanks, i );
        // skip XORs that are not part of any tree
        if Rank == -1 {
            continue; }
        // iterate through XOR inputs
        Fanins[0] = Gia_ObjId(p, Gia_Regular(pFan0));
        Fanins[1] = Gia_ObjId(p, Gia_Regular(pFan1));
        for  k = 0; k < 2; k++  {
            Entry = Vec_IntEntry( vRanks, Fanins[k] );
            if Entry == Rank { // the same tree -- allow fanout in this tree
                continue; }
            if Entry == -1 {
                Vec_IntWriteEntry( vRanks, Fanins[k], Rank );
            } else {
                Vec_IntPush( vDoubles, Fanins[k] ); }
            if Entry != -1 && Gia_ObjIsAnd(Gia_ManObj(p, Fanins[k])) {
            fmt.Printf("Xor node %d belongs to Tree %d and Tree %d.\n", Fanins[k], Entry, Rank); }
        }
    }
    }
    // remove duplicated entries
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDoubles)) && (((Entry) = Vec_IntEntry(vDoubles, i)), 1); i++  {
        Vec_IntWriteEntry( vRanks, Entry, -1 );
    }
    Vec_IntFree( vDoubles );
    return vRanks;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_FindTreeLeaves(p *Gia_Man_t, vXorRoots *Vec_Int_t, vRanks *Vec_Int_t) *Vec_Wec_t {
    var vMapXors *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    var vTreeLeaves *Vec_Wec_t  = Vec_WecStart( Vec_IntSize(vXorRoots) );
    var pFan0, pFan1, pObj *Gia_Obj_t; 
    var i int32;
    var k int32;
    var Fanins [2]int32;
    var Rank int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) {
            continue; }
        Vec_BitWriteEntry( vMapXors, i, 1 );
        Rank = Vec_IntEntry( vRanks, i );
        // skip XORs that are not part of any tree
        if Rank == -1 {
            continue; }
        // iterate through XOR inputs
        Fanins[0] = Gia_ObjId(p, Gia_Regular(pFan0));
        Fanins[1] = Gia_ObjId(p, Gia_Regular(pFan1));
        for  k = 0; k < 2; k++  {
            if Vec_BitEntry(vMapXors, Fanins[k]) {
                assert(Rank == Vec_IntEntry(vRanks, Fanins[k]), "Rank == Vec_IntEntry(vRanks, Fanins[k])");
                continue;
            }
            Vec_WecPush( vTreeLeaves, Rank, Fanins[k] );
        }
    }
    }
    Vec_BitFree( vMapXors );
    return vTreeLeaves;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_FindShadows(p *Gia_Man_t, vRanks *Vec_Int_t) *Vec_Int_t {
    var vShadows *Vec_Int_t  = Vec_IntDup( vRanks );
    var pObj *Gia_Obj_t; var i, Shad0, Shad1 int32;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Vec_IntWriteEntry( vShadows, Gia_ObjId(p, pObj), -1 );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Vec_IntEntry(vShadows, i) >= 0 {
            continue; }
        Shad0 = Vec_IntEntry(vShadows, Gia_ObjFaninId0(pObj, i));
        Shad1 = Vec_IntEntry(vShadows, Gia_ObjFaninId1(pObj, i));
        if Shad0 == Shad1 && Shad0 != -1 {
            Vec_IntWriteEntry(vShadows, i, Shad0); }
    }
    }
    return vShadows;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_CollectSupp_rec(p *Gia_Man_t, iNode int32, Rank int32, vRanks *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t;
    var nSize int32;
    if Gia_ObjIsTravIdCurrentId(p, iNode) {
        return 0; }
    Gia_ObjSetTravIdCurrentId(p, iNode);
    pObj = Gia_ManObj(p, iNode);
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if Vec_IntEntry(vRanks, iNode) == Rank {
        return 1; }
    nSize = Acec_CollectSupp_rec( p, Gia_ObjFaninId0(pObj, iNode), Rank, vRanks );
    nSize += Acec_CollectSupp_rec( p, Gia_ObjFaninId1(pObj, iNode), Rank, vRanks );
    return nSize;
}
func Acec_FindNexts(p *Gia_Man_t, vRanks *Vec_Int_t, vShadows *Vec_Int_t, vTreeLeaves *Vec_Wec_t) *Vec_Wec_t {
    var vNexts *Vec_Wec_t  = Vec_WecStart( Vec_WecSize(vTreeLeaves) );
    var vTree *Vec_Int_t;
    var i int32;
    var k int32;
    var Node int32;
    var Fanins [2]int32;
    var Shad0 int32;
    var Shad1 int32;
    var Rank int32;
    var nSupp int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vTreeLeaves)) && (((vTree) = Vec_WecEntry(vTreeLeaves, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vTree)) && (((Node) = Vec_IntEntry(vTree, k)), 1); k++  {
            var pObj *Gia_Obj_t  = Gia_ManObj(p, Node);
            if !Gia_ObjIsAnd(pObj) {
                continue; }
            Fanins[0] = Gia_ObjFaninId0(pObj, Node);
            Fanins[1] = Gia_ObjFaninId1(pObj, Node);
            Shad0 = Vec_IntEntry(vShadows, Fanins[0]);
            Shad1 = Vec_IntEntry(vShadows, Fanins[1]);
            if Shad0 != Shad1 || Shad0 == -1 {
                continue; }
            // check support size of Node in terms of the shadow of its fanins
            Rank = Vec_IntEntry( vRanks, Node );
            assert(Rank != Shad0, "Rank != Shad0");
            Gia_ManIncrementTravId( p );
            nSupp = Acec_CollectSupp_rec( p, Node, Shad0, vRanks );
            assert(nSupp > 1, "nSupp > 1");
            if nSupp > 3 {
                continue; }
            Vec_IntPushUniqueOrder( Vec_WecEntry(vNexts, Shad0), Rank );
        }
    }
    return vNexts;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_StructTest(p *Gia_Man_t) {
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


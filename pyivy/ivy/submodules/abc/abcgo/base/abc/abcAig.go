package abc

import "fmt"

/**CFile****************************************************************

  FileName    [abcAig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Simple structural hashing package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcAig.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"

//ABC_NAMESPACE_IMPL_START

/*
    AIG is an And-Inv Graph with structural hashing.
    It is always structurally hashed. It means that at any time:
    - for each AND gate, there are no other AND gates with the same children
    - the constants are propagated
    - there is no single-input nodes (inverters/buffers)
    Additionally the following invariants are satisfied:
    - there are no dangling nodes (the nodes without fanout)
    - the level of each AND gate reflects the levels of this fanins
    - the EXOR-status of each node is up-to-date
    - the AND nodes are in the topological order
    - the constant 1 node has always number 0 in the object list
    The operations that are performed on AIGs:
    - building new nodes (Abc_AigAnd)
    - performing elementary Boolean operations (Abc_AigOr, Abc_AigXor, etc)
    - replacing one node by another (Abc_AigReplace)
    - propagating constants (Abc_AigReplace)
    When AIG is duplicated, the new graph is structurally hashed too.
    If this repeated hashing leads to fewer nodes, it means the original
    AIG was not strictly hashed (one of the conditions above is violated).
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the simple AIG manager
type Abc_Aig_t_ struct {
    pNtkAig        *Abc_Ntk_t;           // the AIG network
    pConst1        *Abc_Obj_t;           // the constant 1 object (not a node!)
    pBins       **Abc_Obj_t;             // the table bins
    nBins               int32;             // the size of the table
    nEntries               int32;          // the total number of entries in the table
    vNodes        *Vec_Ptr_t;            // the temporary array of nodes
    vStackReplaceOld        *Vec_Ptr_t;  // the nodes to be replaced
    vStackReplaceNew        *Vec_Ptr_t;  // the nodes to be used for replacement
    vLevels        *Vec_Vec_t;           // the nodes to be updated
    vLevelsR        *Vec_Vec_t;          // the nodes to be updated
    vAddedCells        *Vec_Ptr_t;       // the added nodes
    vUpdatedNets        *Vec_Ptr_t;      // the nodes whose fanouts have changed

    nStrash0               int32;
    nStrash1               int32;
    nStrash5               int32;
    nStrash2               int32;
}

// iterators through the entries in the linked lists of nodes
// Original macro:
// #define Abc_AigBinForEachEntry( pBin, pEnt )                   \
//     for ( pEnt = pBin;                                         \
//           pEnt;                                                \
//           pEnt = pEnt->pNext )

// Original macro:
// #define Abc_AigBinForEachEntrySafe( pBin, pEnt, pEnt2 )        \
//     for ( pEnt = pBin,                                         \
//           pEnt2 = pEnt? pEnt->pNext: NULL;                     \
//           pEnt;                                                \
//           pEnt = pEnt2,                                        \
//           pEnt2 = pEnt? pEnt->pNext: NULL )


// hash key for the structural hash table
//static inline unsigned Abc_HashKey2( Abc_Obj_t * p0, Abc_Obj_t * p1, int TableSize ) { return ((unsigned)(p0) + (unsigned)(p1) * 12582917) % TableSize; }
//static inline unsigned Abc_HashKey2( Abc_Obj_t * p0, Abc_Obj_t * p1, int TableSize ) { return ((unsigned)((a)->Id + (b)->Id) * ((a)->Id + (b)->Id + 1) / 2) % TableSize; }

// hashing the node
func Abc_HashKey2(p0 *Abc_Obj_t, p1 *Abc_Obj_t, TableSize int32) uint32 {
    var Key uint32;
    Key ^= Abc_ObjRegular(p0).Id * 7937;
    Key ^= Abc_ObjRegular(p1).Id * 2971;
    Key ^= Abc_ObjIsComplement(p0) * 911;
    Key ^= Abc_ObjIsComplement(p1) * 353;
    return Key % TableSize;
}

// structural hash table procedures
// static Abc_Obj_t * Abc_AigAndCreate( Abc_Aig_t * pMan, Abc_Obj_t * p0, Abc_Obj_t * p1 );

// static Abc_Obj_t * Abc_AigAndCreateFrom( Abc_Aig_t * pMan, Abc_Obj_t * p0, Abc_Obj_t * p1, Abc_Obj_t * pAnd );

// static void        Abc_AigAndDelete( Abc_Aig_t * pMan, Abc_Obj_t * pThis );

// static void        Abc_AigResize( Abc_Aig_t * pMan );

// incremental AIG procedures
// static void        Abc_AigReplace_int( Abc_Aig_t * pMan, Abc_Obj_t * pOld, Abc_Obj_t * pNew, int fUpdateLevel );

// static void        Abc_AigUpdateLevel_int( Abc_Aig_t * pMan );

// static void        Abc_AigUpdateLevelR_int( Abc_Aig_t * pMan );

// static void        Abc_AigRemoveFromLevelStructure( Vec_Vec_t * vStruct, Abc_Obj_t * pNode );

// static void        Abc_AigRemoveFromLevelStructureR( Vec_Vec_t * vStruct, Abc_Obj_t * pNode );



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the local AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigAlloc(pNtkAig *Abc_Ntk_t) *Abc_Aig_t {
    var pMan *Abc_Aig_t;
    // start the manager
    // inlining of macro ABC_ALLOC
    pMan = ((Abc_Aig_t *) malloc(sizeof(Abc_Aig_t) * (size_t)(1)));
    memset( pMan, 0, sizeof(Abc_Aig_t) );
    // allocate the table
    pMan.nBins    = Abc_PrimeCudd( 10000 );
    // inlining of macro ABC_ALLOC
    pMan.pBins    = ((Abc_Obj_t * *) malloc(sizeof(Abc_Obj_t *) * (size_t)(pMan.nBins)));
    memset( pMan.pBins, 0, sizeof(Abc_Obj_t *) * pMan.nBins );
    pMan.vNodes   = Vec_PtrAlloc( 100 );
    pMan.vLevels  = Vec_VecAlloc( 100 );
    pMan.vLevelsR = Vec_VecAlloc( 100 );
    pMan.vStackReplaceOld = Vec_PtrAlloc( 100 );
    pMan.vStackReplaceNew = Vec_PtrAlloc( 100 );
    // create the constant node
    assert(pNtkAig.vObjs.nSize == 0, "pNtkAig.vObjs.nSize == 0");
    pMan.pConst1 = Abc_NtkCreateObj( pNtkAig, ABC_OBJ_NODE );
    pMan.pConst1.Type = ABC_OBJ_CONST1;
    pMan.pConst1.fPhase = 1;
    pNtkAig.nObjCounts[ABC_OBJ_NODE]--;
    // save the current network
    pMan.pNtkAig = pNtkAig;
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Deallocates the local AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigFree(pMan *Abc_Aig_t) {
    assert(Vec_PtrSize( pMan.vStackReplaceOld ) == 0, "Vec_PtrSize( pMan.vStackReplaceOld ) == 0");
    assert(Vec_PtrSize( pMan.vStackReplaceNew ) == 0, "Vec_PtrSize( pMan.vStackReplaceNew ) == 0");
    // free the table
    if pMan.vAddedCells {
        Vec_PtrFree( pMan.vAddedCells ); }
    if pMan.vUpdatedNets {
        Vec_PtrFree( pMan.vUpdatedNets ); }
    Vec_VecFree( pMan.vLevels );
    Vec_VecFree( pMan.vLevelsR );
    Vec_PtrFree( pMan.vStackReplaceOld );
    Vec_PtrFree( pMan.vStackReplaceNew );
    Vec_PtrFree( pMan.vNodes );
    // inlining of macro ABC_FREE
    ((pMan.pBins) ? (free((char *) (pMan.pBins)), (pMan.pBins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pMan) ? (free((char *) (pMan)), (pMan) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Returns the number of dangling nodes removed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigCleanup(pMan *Abc_Aig_t) int32 {
    var vDangles *Vec_Ptr_t;
    var pAnd *Abc_Obj_t;
    var i, nNodesOld int32;
//    printf( "Strash0 = %d.  Strash1 = %d.  Strash100 = %d.  StrashM = %d.\n", 
//        pMan->nStrash0, pMan->nStrash1, pMan->nStrash5, pMan->nStrash2 );
    nNodesOld = pMan.nEntries;
    // collect the AND nodes that do not fanout
    vDangles = Vec_PtrAlloc( 100 );
    for  i = 0; i < pMan.nBins; i++  {
        // inlining of macro Abc_AigBinForEachEntry
        for  pAnd = pMan.pBins[i];                                                   (pAnd != nil);                                                          pAnd = pAnd.pNext  {
            if Abc_ObjFanoutNum(pAnd) == 0 {
                Vec_PtrPush( vDangles, pAnd ); }
        }
    }
    // process the dangling nodes and their MFFCs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vDangles)) && (((pAnd) = (Abc_Obj_t *)Vec_PtrEntry(vDangles, i)), 1); i++  {
        Abc_AigDeleteNode( pMan, pAnd );
    }
    Vec_PtrFree( vDangles );
    return nNodesOld - pMan.nEntries;
}

/**Function*************************************************************

  Synopsis    [Makes sure that every node in the table is in the network and vice versa.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigCheck(pMan *Abc_Aig_t) int32 {
    var pObj, pAnd *Abc_Obj_t;
    var i, nFanins, Counter int32;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pMan.pNtkAig).vObjs)) && (((pObj) = Abc_NtkObj(pMan.pNtkAig, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        nFanins = Abc_ObjFaninNum(pObj);
        if nFanins == 0 {
            if !Abc_AigNodeIsConst(pObj) {
                fmt.Printf("Abc_AigCheck: The AIG has non-standard constant nodes.\n");
                return 0;
            }
            continue;
        }
        if nFanins == 1 {
            fmt.Printf("Abc_AigCheck: The AIG has single input nodes.\n");
            return 0;
        }
        if nFanins > 2 {
            fmt.Printf("Abc_AigCheck: The AIG has non-standard nodes.\n");
            return 0;
        }
        if pObj.Level != 1 + uint32(Abc_MaxInt( Abc_ObjFanin0(pObj).Level, Abc_ObjFanin1(pObj).Level )) {
            fmt.Printf("Abc_AigCheck: Node \"%s\" has level that does not agree with the fanin levels.\n", Abc_ObjName(pObj)); }
        pAnd = Abc_AigAndLookup( pMan, Abc_ObjChild0(pObj), Abc_ObjChild1(pObj) );
        if pAnd != pObj {
            fmt.Printf("Abc_AigCheck: Node \"%s\" is not in the structural hashing table.\n", Abc_ObjName(pObj)); }
    }
    }
    // count the number of nodes in the table
    Counter = 0;
    for  i = 0; i < pMan.nBins; i++  {
        // inlining of macro Abc_AigBinForEachEntry
        for  pAnd = pMan.pBins[i];                                                   (pAnd != nil);                                                          pAnd = pAnd.pNext  {
            Counter++;
        }
    }
    if Counter != Abc_NtkNodeNum(pMan.pNtkAig) {
        fmt.Printf("Abc_AigCheck: The number of nodes in the structural hashing table is wrong.\n");
        return 0;
    }
    // if the node is a choice node, nodes in its class should not have fanouts
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pMan.pNtkAig).vObjs)) && (((pObj) = Abc_NtkObj(pMan.pNtkAig, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else
        if Abc_AigNodeIsChoice(pObj) {
            for  pAnd = (Abc_Obj_t *)pObj.pData; (pAnd != nil); pAnd = (Abc_Obj_t *)pAnd.pData  {
                if Abc_ObjFanoutNum(pAnd) > 0 {
                    fmt.Printf("Abc_AigCheck: Representative %s", Abc_ObjName(pAnd));
                    fmt.Printf(" of choice node %s has %d fanouts.\n", Abc_ObjName(pObj), Abc_ObjFanoutNum(pAnd));
                    return 0;
                }
            } }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Computes the number of logic levels not counting PIs/POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigLevel(pNtk *Abc_Ntk_t) int32 {
    var pNode *Abc_Obj_t;
    var i, LevelsMax int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    if pNtk.nBarBufs {
        return Abc_NtkLevel( pNtk ); }
    // perform the traversal
    LevelsMax = 0;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        if LevelsMax < int32(Abc_ObjFanin0(pNode).Level) {
            LevelsMax = int32(Abc_ObjFanin0(pNode).Level); }
    }
    return LevelsMax;
}


/**Function*************************************************************

  Synopsis    [Performs canonicization step.]

  Description [The argument nodes can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigAndCreate(pMan *Abc_Aig_t, p0 *Abc_Obj_t, p1 *Abc_Obj_t) *Abc_Obj_t {
    var pAnd *Abc_Obj_t;
    var Key uint32;
    // check if it is a good time for table resizing
    if pMan.nEntries > 2 * pMan.nBins {
        Abc_AigResize( pMan ); }
    // order the arguments
    if Abc_ObjRegular(p0).Id > Abc_ObjRegular(p1).Id {
        pAnd = p0;
p0 = p1;
p1 = pAnd; }
    // create the new node
    pAnd = Abc_NtkCreateNode( pMan.pNtkAig );
    Abc_ObjAddFanin( pAnd, p0 );
    Abc_ObjAddFanin( pAnd, p1 );
    // set the level of the new node
    pAnd.Level  = 1 + Abc_MaxInt( Abc_ObjRegular(p0).Level, Abc_ObjRegular(p1).Level ); 
    pAnd.fExor  = Abc_NodeIsExorType(pAnd);
    pAnd.fPhase = (Abc_ObjIsComplement(p0) ^ Abc_ObjRegular(p0).fPhase) & (Abc_ObjIsComplement(p1) ^ Abc_ObjRegular(p1).fPhase);
    // add the node to the corresponding linked list in the table
    Key = Abc_HashKey2( p0, p1, pMan.nBins );
    pAnd.pNext      = pMan.pBins[Key];
    pMan.pBins[Key] = pAnd;
    pMan.nEntries++;
    // create the cuts if defined
//    if ( pAnd->pNtk->pManCut )
//        Abc_NodeGetCuts( pAnd->pNtk->pManCut, pAnd );
    pAnd.pCopy = nil;
    // add the node to the list of updated nodes
    if pMan.vAddedCells {
        Vec_PtrPush( pMan.vAddedCells, pAnd ); }
    return pAnd;
}

/**Function*************************************************************

  Synopsis    [Performs canonicization step.]

  Description [The argument nodes can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigAndCreateFrom(pMan *Abc_Aig_t, p0 *Abc_Obj_t, p1 *Abc_Obj_t, pAnd *Abc_Obj_t) *Abc_Obj_t {
    var pTemp *Abc_Obj_t;
    var Key uint32;
    assert(!Abc_ObjIsComplement(pAnd), "!Abc_ObjIsComplement(pAnd)");
    // order the arguments
    if Abc_ObjRegular(p0).Id > Abc_ObjRegular(p1).Id {
        pTemp = p0;
p0 = p1;
p1 = pTemp; }
    // create the new node
    Abc_ObjAddFanin( pAnd, p0 );
    Abc_ObjAddFanin( pAnd, p1 );
    // set the level of the new node
    pAnd.Level      = 1 + Abc_MaxInt( Abc_ObjRegular(p0).Level, Abc_ObjRegular(p1).Level ); 
    pAnd.fExor      = Abc_NodeIsExorType(pAnd);
    // add the node to the corresponding linked list in the table
    Key = Abc_HashKey2( p0, p1, pMan.nBins );
    pAnd.pNext      = pMan.pBins[Key];
    pMan.pBins[Key] = pAnd;
    pMan.nEntries++;
    // create the cuts if defined
//    if ( pAnd->pNtk->pManCut )
//        Abc_NodeGetCuts( pAnd->pNtk->pManCut, pAnd );
    pAnd.pCopy = nil;
    // add the node to the list of updated nodes
//    if ( pMan->vAddedCells )
//        Vec_PtrPush( pMan->vAddedCells, pAnd );
    return pAnd;
}

/**Function*************************************************************

  Synopsis    [Performs canonicization step.]

  Description [The argument nodes can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigAndLookup(pMan *Abc_Aig_t, p0 *Abc_Obj_t, p1 *Abc_Obj_t) *Abc_Obj_t {
    var pAnd, pConst1 *Abc_Obj_t;
    var Key uint32;
    assert(Abc_ObjRegular(p0).pNtk.pManFunc == pMan, "Abc_ObjRegular(p0).pNtk.pManFunc == pMan");
    assert(Abc_ObjRegular(p1).pNtk.pManFunc == pMan, "Abc_ObjRegular(p1).pNtk.pManFunc == pMan");
    // check for trivial cases
    pConst1 = Abc_AigConst1(pMan.pNtkAig);
    if p0 == p1 {
        return p0; }
    if p0 == Abc_ObjNot(p1) {
        return Abc_ObjNot(pConst1); }
    if Abc_ObjRegular(p0) == pConst1 {
        if p0 == pConst1 {
            return p1; }
        return Abc_ObjNot(pConst1);
    }
    if Abc_ObjRegular(p1) == pConst1 {
        if p1 == pConst1 {
            return p0; }
        return Abc_ObjNot(pConst1);
    }
/*
    {
        int nFans0 = Abc_ObjFanoutNum( Abc_ObjRegular(p0) );
        int nFans1 = Abc_ObjFanoutNum( Abc_ObjRegular(p1) );
        if ( nFans0 == 0 || nFans1 == 0 )
            pMan->nStrash0++;
        else if ( nFans0 == 1 || nFans1 == 1 )
            pMan->nStrash1++;
        else if ( nFans0 <= 100 && nFans1 <= 100 )
            pMan->nStrash5++;
        else
            pMan->nStrash2++;
    }
*/
    {
        var nFans0 int32  = Abc_ObjFanoutNum( Abc_ObjRegular(p0) );
        var nFans1 int32  = Abc_ObjFanoutNum( Abc_ObjRegular(p1) );
        if nFans0 == 0 || nFans1 == 0 {
            return nil; }
    }

    // order the arguments
    if Abc_ObjRegular(p0).Id > Abc_ObjRegular(p1).Id {
        pAnd = p0;
p0 = p1;
p1 = pAnd; }
    // get the hash key for these two nodes
    Key = Abc_HashKey2( p0, p1, pMan.nBins );
    // find the matching node in the table
    // inlining of macro Abc_AigBinForEachEntry
    for  pAnd = pMan.pBins[Key];                                                   (pAnd != nil);                                                          pAnd = pAnd.pNext  {
        if p0 == Abc_ObjChild0(pAnd) && p1 == Abc_ObjChild1(pAnd) {
//            assert( Abc_ObjFanoutNum(Abc_ObjRegular(p0)) && Abc_ObjFanoutNum(p1) );
             return pAnd;
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Returns the gate implementing EXOR of the two arguments if it exists.]

  Description [The argument nodes can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigXorLookup(pMan *Abc_Aig_t, p0 *Abc_Obj_t, p1 *Abc_Obj_t, pType *int32) *Abc_Obj_t {
    var pNode1, pNode2, pNode *Abc_Obj_t;
    // set the flag to zero
    if pType != nil { *pType = 0; }
    // check the case of XOR(a,b) = OR(ab, a'b')'
    if (pNode1 = Abc_AigAndLookup(pMan, Abc_ObjNot(p0), Abc_ObjNot(p1))) &&
         (pNode2 = Abc_AigAndLookup(pMan, p0, p1)) {
        pNode = Abc_AigAndLookup( pMan, Abc_ObjNot(pNode1), Abc_ObjNot(pNode2) );
        if (pNode != nil) && (pType != nil) { *pType = 1; }
        return pNode;
    }
    // check the case of XOR(a,b) = OR(a'b, ab')
    if (pNode1 = Abc_AigAndLookup(pMan, p0, Abc_ObjNot(p1))) &&
         (pNode2 = Abc_AigAndLookup(pMan, Abc_ObjNot(p0), p1)) {
        pNode = Abc_AigAndLookup( pMan, Abc_ObjNot(pNode1), Abc_ObjNot(pNode2) );
        if pNode != nil {
            return Abc_ObjNot(pNode);
        } else {
            return nil;
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Returns the gate implementing EXOR of the two arguments if it exists.]

  Description [The argument nodes can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigMuxLookup(pMan *Abc_Aig_t, pC *Abc_Obj_t, pT *Abc_Obj_t, pE *Abc_Obj_t, pType *int32) *Abc_Obj_t {
    var pNode1, pNode2, pNode *Abc_Obj_t;
    // set the flag to zero
    if pType != nil { *pType = 0; }
    // check the case of MUX(c,t,e) = OR(ct', c'e')'
    if (pNode1 = Abc_AigAndLookup(pMan, pC, Abc_ObjNot(pT))) &&
         (pNode2 = Abc_AigAndLookup(pMan, Abc_ObjNot(pC), Abc_ObjNot(pE))) {
        pNode = Abc_AigAndLookup( pMan, Abc_ObjNot(pNode1), Abc_ObjNot(pNode2) );
        if (pNode != nil) && (pType != nil) { *pType = 1; }
        return pNode;
    }
    // check the case of MUX(c,t,e) = OR(ct, c'e)
    if (pNode1 = Abc_AigAndLookup(pMan, pC, pT)) &&
         (pNode2 = Abc_AigAndLookup(pMan, Abc_ObjNot(pC), pE)) {
        pNode = Abc_AigAndLookup( pMan, Abc_ObjNot(pNode1), Abc_ObjNot(pNode2) );
        if pNode != nil {
            return Abc_ObjNot(pNode);
        } else {
            return nil;
        }
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Deletes an AIG node from the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigAndDelete(pMan *Abc_Aig_t, pThis *Abc_Obj_t) {
    var pAnd *Abc_Obj_t;
    var pAnd0 *Abc_Obj_t;
    var pAnd1 *Abc_Obj_t;
    var ppPlace **Abc_Obj_t;
    var Key uint32;
    assert(!Abc_ObjIsComplement(pThis), "!Abc_ObjIsComplement(pThis)");
    assert(Abc_ObjIsNode(pThis), "Abc_ObjIsNode(pThis)");
    assert(Abc_ObjFaninNum(pThis) == 2, "Abc_ObjFaninNum(pThis) == 2");
    assert(pMan.pNtkAig == pThis.pNtk, "pMan.pNtkAig == pThis.pNtk");
    // get the hash key for these two nodes
    pAnd0 = Abc_ObjRegular( Abc_ObjChild0(pThis) );
    pAnd1 = Abc_ObjRegular( Abc_ObjChild1(pThis) );
    Key = Abc_HashKey2( Abc_ObjChild0(pThis), Abc_ObjChild1(pThis), pMan.nBins );
    // find the matching node in the table
    ppPlace = pMan.pBins + Key;
    // inlining of macro Abc_AigBinForEachEntry
    for  pAnd = pMan.pBins[Key];                                                   (pAnd != nil);                                                          pAnd = pAnd.pNext  {
        if pAnd != pThis {
            ppPlace = &pAnd.pNext;
            continue;
        }
        *ppPlace = pAnd.pNext;
        break;
    }
    assert(pAnd == pThis, "pAnd == pThis");
    pMan.nEntries--;
    // delete the cuts if defined
    if pThis.pNtk.pManCut {
        Abc_NodeFreeCuts( pThis.pNtk.pManCut, pThis ); }
}

/**Function*************************************************************

  Synopsis    [Resizes the hash table of AIG nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigResize(pMan *Abc_Aig_t) {
    var pBinsNew **Abc_Obj_t;
    var pEnt, pEnt2 *Abc_Obj_t;
    var nBinsNew, Counter, i int32;
    var clk abctime;
    var Key uint32;

clk = Abc_Clock();
    // get the new table size
    nBinsNew = Abc_PrimeCudd( 3 * pMan.nBins ); 
    // allocate a new array
    // inlining of macro ABC_ALLOC
    pBinsNew = ((Abc_Obj_t * *) malloc(sizeof(Abc_Obj_t *) * (size_t)(nBinsNew)));
    memset( pBinsNew, 0, sizeof(Abc_Obj_t *) * nBinsNew );
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < pMan.nBins; i++  {
        // inlining of macro Abc_AigBinForEachEntrySafe
        for  pEnt = pMan.pBins[i],                                                   pEnt2 = pEnt? pEnt.pNext: nil;                               (pEnt != nil);                                                          pEnt = pEnt2,                                                  pEnt2 = pEnt? pEnt.pNext: nil  {
            Key = Abc_HashKey2( Abc_ObjChild0(pEnt), Abc_ObjChild1(pEnt), nBinsNew );
            pEnt.pNext   = pBinsNew[Key];
            pBinsNew[Key] = pEnt;
            Counter++;
        }
    }
    assert(Counter == pMan.nEntries, "Counter == pMan.nEntries");
//    printf( "Increasing the structural table size from %6d to %6d. ", pMan->nBins, nBinsNew );
//    ABC_PRT( "Time", Abc_Clock() - clk );
    // replace the table and the parameters
    // inlining of macro ABC_FREE
    ((pMan.pBins) ? (free((char *) (pMan.pBins)), (pMan.pBins) = 0) : 0);
    pMan.pBins = pBinsNew;
    pMan.nBins = nBinsNew;
}

/**Function*************************************************************

  Synopsis    [Resizes the hash table of AIG nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigRehash(pMan *Abc_Aig_t) {
    var pBinsNew **Abc_Obj_t;
    var pEnt, pEnt2 *Abc_Obj_t;
    var pArray *int32;
    var Key uint32;
    var Counter, Temp, i int32;

    // allocate a new array
    // inlining of macro ABC_ALLOC
    pBinsNew = ((Abc_Obj_t * *) malloc(sizeof(Abc_Obj_t *) * (size_t)(pMan.nBins)));
    memset( pBinsNew, 0, sizeof(Abc_Obj_t *) * pMan.nBins );
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < pMan.nBins; i++  {
        // inlining of macro Abc_AigBinForEachEntrySafe
        for  pEnt = pMan.pBins[i],                                                   pEnt2 = pEnt? pEnt.pNext: nil;                               (pEnt != nil);                                                          pEnt = pEnt2,                                                  pEnt2 = pEnt? pEnt.pNext: nil  {
            // swap the fanins if needed
            pArray = pEnt.vFanins.pArray;
            if pArray[0] > pArray[1] {
                Temp = pArray[0];
                pArray[0] = pArray[1];
                pArray[1] = Temp;
                Temp = pEnt.fCompl0;
                pEnt.fCompl0 = pEnt.fCompl1;
                pEnt.fCompl1 = Temp;
            }
            // rehash the node
            Key = Abc_HashKey2( Abc_ObjChild0(pEnt), Abc_ObjChild1(pEnt), pMan.nBins );
            pEnt.pNext   = pBinsNew[Key];
            pBinsNew[Key] = pEnt;
            Counter++;
        }
    }
    assert(Counter == pMan.nEntries, "Counter == pMan.nEntries");
    // replace the table and the parameters
    // inlining of macro ABC_FREE
    ((pMan.pBins) ? (free((char *) (pMan.pBins)), (pMan.pBins) = 0) : 0);
    pMan.pBins = pBinsNew;
}






/**Function*************************************************************

  Synopsis    [Performs canonicization step.]

  Description [The argument nodes can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigConst1(pNtk *Abc_Ntk_t) *Abc_Obj_t {
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    return ((Abc_Aig_t *)pNtk.pManFunc).pConst1;
}

/**Function*************************************************************

  Synopsis    [Performs canonicization step.]

  Description [The argument nodes can be complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigAnd(pMan *Abc_Aig_t, p0 *Abc_Obj_t, p1 *Abc_Obj_t) *Abc_Obj_t {
    var pAnd *Abc_Obj_t;
    if pAnd = Abc_AigAndLookup( pMan, p0, p1 ) {
        return pAnd; }
    return Abc_AigAndCreate( pMan, p0, p1 );
}

/**Function*************************************************************

  Synopsis    [Implements Boolean OR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigOr(pMan *Abc_Aig_t, p0 *Abc_Obj_t, p1 *Abc_Obj_t) *Abc_Obj_t {
    return Abc_ObjNot( Abc_AigAnd( pMan, Abc_ObjNot(p0), Abc_ObjNot(p1) ) );
}

/**Function*************************************************************

  Synopsis    [Implements Boolean XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigXor(pMan *Abc_Aig_t, p0 *Abc_Obj_t, p1 *Abc_Obj_t) *Abc_Obj_t {
    return Abc_AigOr( pMan, Abc_AigAnd(pMan, p0, Abc_ObjNot(p1)), 
                            Abc_AigAnd(pMan, p1, Abc_ObjNot(p0)) );
}

/**Function*************************************************************

  Synopsis    [Implements Boolean XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigMux(pMan *Abc_Aig_t, pC *Abc_Obj_t, p1 *Abc_Obj_t, p0 *Abc_Obj_t) *Abc_Obj_t {
    return Abc_AigOr( pMan, Abc_AigAnd(pMan, pC, p1), Abc_AigAnd(pMan, Abc_ObjNot(pC), p0) );
}
 
/**Function*************************************************************

  Synopsis    [Implements the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigMiter_rec(pMan *Abc_Aig_t, ppObjs **Abc_Obj_t, nObjs int32) *Abc_Obj_t {
    var pObj1, pObj2 *Abc_Obj_t;
    if nObjs == 1 {
        return ppObjs[0]; }
    pObj1 = Abc_AigMiter_rec( pMan, ppObjs,           nObjs/2 );
    pObj2 = Abc_AigMiter_rec( pMan, ppObjs + nObjs/2, nObjs - nObjs/2 );
    return Abc_AigOr( pMan, pObj1, pObj2 );
}
 
/**Function*************************************************************

  Synopsis    [Implements the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigMiter(pMan *Abc_Aig_t, vPairs *Vec_Ptr_t, fImplic int32) *Abc_Obj_t {
    var i int32;
    if vPairs.nSize == 0 {
        return Abc_ObjNot( Abc_AigConst1(pMan.pNtkAig) ); }
    assert(vPairs.nSize % 2 == 0, "vPairs.nSize % 2 == 0");
    // go through the cubes of the node's SOP
    if fImplic != 0 {
        for  i = 0; i < vPairs.nSize; i += 2  {
            vPairs.pArray[i/2] = Abc_AigAnd( pMan, (Abc_Obj_t *)vPairs.pArray[i], Abc_ObjNot((Abc_Obj_t *)vPairs.pArray[i+1]) );
        }
    
    } else {
        for  i = 0; i < vPairs.nSize; i += 2  {
            vPairs.pArray[i/2] = Abc_AigXor( pMan, (Abc_Obj_t *)vPairs.pArray[i], (Abc_Obj_t *)vPairs.pArray[i+1] );
        }
    }
    vPairs.nSize = vPairs.nSize/2;
    return Abc_AigMiter_rec( pMan, (Abc_Obj_t **)vPairs.pArray, vPairs.nSize );
}
 
/**Function*************************************************************

  Synopsis    [Implements the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigMiter2(pMan *Abc_Aig_t, vPairs *Vec_Ptr_t) *Abc_Obj_t {
    var pMiter, pXor *Abc_Obj_t;
    var i int32;
    assert(vPairs.nSize % 2 == 0, "vPairs.nSize % 2 == 0");
    // go through the cubes of the node's SOP
    pMiter = Abc_ObjNot( Abc_AigConst1(pMan.pNtkAig) );
    for  i = 0; i < vPairs.nSize; i += 2  {
        pXor   = Abc_AigXor( pMan, (Abc_Obj_t *)vPairs.pArray[i], (Abc_Obj_t *)vPairs.pArray[i+1] );
        pMiter = Abc_AigOr( pMan, pMiter, pXor );
    }
    return pMiter;
}




/**Function*************************************************************

  Synopsis    [Replaces one AIG node by the other.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigReplace(pMan *Abc_Aig_t, pOld *Abc_Obj_t, pNew *Abc_Obj_t, fUpdateLevel int32) {
    assert(Vec_PtrSize(pMan.vStackReplaceOld) == 0, "Vec_PtrSize(pMan.vStackReplaceOld) == 0");
    assert(Vec_PtrSize(pMan.vStackReplaceNew) == 0, "Vec_PtrSize(pMan.vStackReplaceNew) == 0");
    Vec_PtrPush( pMan.vStackReplaceOld, pOld );
    Vec_PtrPush( pMan.vStackReplaceNew, pNew );
    assert(!Abc_ObjIsComplement(pOld), "!Abc_ObjIsComplement(pOld)");
    // process the replacements
    for  Vec_PtrSize(pMan.vStackReplaceOld)  {
        pOld = (Abc_Obj_t *)Vec_PtrPop( pMan.vStackReplaceOld );
        pNew = (Abc_Obj_t *)Vec_PtrPop( pMan.vStackReplaceNew );
        Abc_AigReplace_int( pMan, pOld, pNew, fUpdateLevel );
    }
    if fUpdateLevel != 0 {
        Abc_AigUpdateLevel_int( pMan );
        if pMan.pNtkAig.vLevelsR { 
            Abc_AigUpdateLevelR_int( pMan ); }
    }
}

/**Function*************************************************************

  Synopsis    [Performs internal replacement step.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigReplace_int(pMan *Abc_Aig_t, pOld *Abc_Obj_t, pNew *Abc_Obj_t, fUpdateLevel int32) {
    var pFanin1, pFanin2, pFanout, pFanoutNew, pFanoutFanout *Abc_Obj_t;
    var k, v, iFanin int32; 
    // make sure the old node is regular and has fanouts
    // (the new node can be complemented and can have fanouts)
    assert(!Abc_ObjIsComplement(pOld), "!Abc_ObjIsComplement(pOld)");
    assert(Abc_ObjFanoutNum(pOld) > 0, "Abc_ObjFanoutNum(pOld) > 0");
    // look at the fanouts of old node
    Abc_NodeCollectFanouts( pOld, pMan.vNodes );
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(pMan.vNodes)) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(pMan.vNodes, k)), 1); k++  {
        if Abc_ObjIsCo(pFanout) {
            Abc_ObjPatchFanin( pFanout, pOld, pNew );
            continue;
        }
        // find the old node as a fanin of this fanout
        iFanin = Vec_IntFind( &pFanout.vFanins, pOld.Id );
        assert(iFanin == 0 || iFanin == 1, "iFanin == 0 || iFanin == 1");
        // get the new fanin
        pFanin1 = Abc_ObjNotCond( pNew, Abc_ObjFaninC(pFanout, iFanin) );
        assert(Abc_ObjRegular(pFanin1) != pFanout, "Abc_ObjRegular(pFanin1) != pFanout");
        // get another fanin
        pFanin2 = Abc_ObjChild( pFanout, iFanin ^ 1 );
        assert(Abc_ObjRegular(pFanin2) != pFanout, "Abc_ObjRegular(pFanin2) != pFanout");
        // check if the node with these fanins exists
        if pFanoutNew = Abc_AigAndLookup( pMan, pFanin1, pFanin2 ) { // such node exists (it may be a constant)
            // schedule replacement of the old fanout by the new fanout
            Vec_PtrPush( pMan.vStackReplaceOld, pFanout );
            Vec_PtrPush( pMan.vStackReplaceNew, pFanoutNew );
            continue;
        }
        // such node does not exist - modify the old fanout node 
        // (this way the change will not propagate all the way to the COs)
        assert(Abc_ObjRegular(pFanin1) != Abc_ObjRegular(pFanin2), "Abc_ObjRegular(pFanin1) != Abc_ObjRegular(pFanin2)");             

        // if the node is in the level structure, remove it
        if pFanout.fMarkA {
            Abc_AigRemoveFromLevelStructure( pMan.vLevels, pFanout ); }
        // if the node is in the level structure, remove it
        if pFanout.fMarkB {
            Abc_AigRemoveFromLevelStructureR( pMan.vLevelsR, pFanout ); }

        // remove the old fanout node from the structural hashing table
        Abc_AigAndDelete( pMan, pFanout );
        // remove the fanins of the old fanout
        Abc_ObjRemoveFanins( pFanout );
        // recreate the old fanout with new fanins and add it to the table
        Abc_AigAndCreateFrom( pMan, pFanin1, pFanin2, pFanout );
        assert((Abc_AigNodeIsAcyclic(pFanout, pFanout) != 0), "Abc_AigNodeIsAcyclic(pFanout, pFanout)");

        if fUpdateLevel != 0 {
            // schedule the updated fanout for updating direct level
            assert(pFanout.fMarkA == 0, "pFanout.fMarkA == 0");
            pFanout.fMarkA = 1;
            Vec_VecPush( pMan.vLevels, pFanout.Level, pFanout );
            // schedule the updated fanout for updating reverse level
            if pMan.pNtkAig.vLevelsR {
                assert(pFanout.fMarkB == 0, "pFanout.fMarkB == 0");
                pFanout.fMarkB = 1;
                Vec_VecPush( pMan.vLevelsR, Abc_ObjReverseLevel(pFanout), pFanout );
            }
        }

        // the fanout has changed, update EXOR status of its fanouts
        // inlining of macro Abc_ObjForEachFanout
        for  v = 0; (v < Abc_ObjFanoutNum(pFanout)) && (((pFanoutFanout) = Abc_ObjFanout(pFanout, v)), 1); v++  {
            if Abc_AigNodeIsAnd(pFanoutFanout) {
                pFanoutFanout.fExor = Abc_NodeIsExorType(pFanoutFanout); }
        }
    }
    // if the node has no fanouts left, remove its MFFC
    if Abc_ObjFanoutNum(pOld) == 0 {
        Abc_AigDeleteNode( pMan, pOld ); }
}

/**Function*************************************************************

  Synopsis    [Performs internal deletion step.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigDeleteNode(pMan *Abc_Aig_t, pNode *Abc_Obj_t) {
    var pNode0, pNode1, pTemp *Abc_Obj_t;
    var i, k int32;

    // make sure the node is regular and dangling
    assert(!Abc_ObjIsComplement(pNode), "!Abc_ObjIsComplement(pNode)");
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    assert(Abc_ObjFaninNum(pNode) == 2, "Abc_ObjFaninNum(pNode) == 2");
    assert(Abc_ObjFanoutNum(pNode) == 0, "Abc_ObjFanoutNum(pNode) == 0");

    // when deleting an old node that is scheduled for replacement, remove it from the replacement queue
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vStackReplaceOld)) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(pMan.vStackReplaceOld, i)), 1); i++  {
        if pNode == pTemp {
            // remove the entry from the replacement array
            for  k = i; k < pMan.vStackReplaceOld.nSize - 1; k++  {
                pMan.vStackReplaceOld.pArray[k] = pMan.vStackReplaceOld.pArray[k+1];
                pMan.vStackReplaceNew.pArray[k] = pMan.vStackReplaceNew.pArray[k+1];
            }
            pMan.vStackReplaceOld.nSize--;
            pMan.vStackReplaceNew.nSize--;
        }
    }

    // when deleting a new node that should replace another node, do not delete
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pMan.vStackReplaceNew)) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(pMan.vStackReplaceNew, i)), 1); i++  {
        if pNode == Abc_ObjRegular(pTemp) {
            return; }
    }

    // remember the node's fanins
    pNode0 = Abc_ObjFanin0( pNode );
    pNode1 = Abc_ObjFanin1( pNode );

    // add the node to the list of updated nodes
    if pMan.vUpdatedNets {
        Vec_PtrPushUnique( pMan.vUpdatedNets, pNode0 );
        Vec_PtrPushUnique( pMan.vUpdatedNets, pNode1 );
    }

    // remove the node from the table
    Abc_AigAndDelete( pMan, pNode );
    // if the node is in the level structure, remove it
    if pNode.fMarkA {
        Abc_AigRemoveFromLevelStructure( pMan.vLevels, pNode ); }
    if pNode.fMarkB {
        Abc_AigRemoveFromLevelStructureR( pMan.vLevelsR, pNode ); }
    // remove the node from the network
    Abc_NtkDeleteObj( pNode );

    // call recursively for the fanins
    if Abc_ObjIsNode(pNode0) && pNode0.vFanouts.nSize == 0 {
        Abc_AigDeleteNode( pMan, pNode0 ); }
    if Abc_ObjIsNode(pNode1) && pNode1.vFanouts.nSize == 0 {
        Abc_AigDeleteNode( pMan, pNode1 ); }
}


/**Function*************************************************************

  Synopsis    [Updates the level of the node after it has changed.]

  Description [This procedure is based on the observation that
  after the node's level has changed, the fanouts levels can change too, 
  but the new fanout levels are always larger than the node's level.
  As a result, we can accumulate the nodes to be updated in the queue
  and process them in the increasing order of levels.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigUpdateLevel_int(pMan *Abc_Aig_t) {
    var pNode, pFanout *Abc_Obj_t;
    var vVec *Vec_Ptr_t;
    var LevelNew, i, k, v int32;

    // go through the nodes and update the level of their fanouts
    // inlining of macro Vec_VecForEachLevel
    for  i = 0; (i < Vec_VecSize(pMan.vLevels)) && (((vVec) = Vec_VecEntry(pMan.vLevels, i)), 1); i++  {
        if Vec_PtrSize(vVec) == 0 {
            continue; }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vVec)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vVec, k)), 1); k++  {
            if pNode == nil {
                continue; }
            assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
            assert(int32(pNode.Level) == i, "int32(pNode.Level) == i");
            // clean the mark
            assert(pNode.fMarkA == 1, "pNode.fMarkA == 1");
            pNode.fMarkA = 0;
            // iterate through the fanouts
            // inlining of macro Abc_ObjForEachFanout
            for  v = 0; (v < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, v)), 1); v++  {
                if Abc_ObjIsCo(pFanout) {
                    continue; }
                // get the new level of this fanout
                LevelNew = 1 + Abc_MaxInt( Abc_ObjFanin0(pFanout).Level, Abc_ObjFanin1(pFanout).Level );
                assert(LevelNew > i, "LevelNew > i");
                if int32(pFanout.Level) == LevelNew { // no change
                    continue; }
                // if the fanout is present in the data structure, pull it out
                if pFanout.fMarkA {
                    Abc_AigRemoveFromLevelStructure( pMan.vLevels, pFanout ); }
                // update the fanout level
                pFanout.Level = LevelNew;
                // add the fanout to the data structure to update its fanouts
                assert(pFanout.fMarkA == 0, "pFanout.fMarkA == 0");
                pFanout.fMarkA = 1;
                Vec_VecPush( pMan.vLevels, pFanout.Level, pFanout );
            }
        }
        Vec_PtrClear( vVec );
    }
}

/**Function*************************************************************

  Synopsis    [Updates the level of the node after it has changed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigUpdateLevelR_int(pMan *Abc_Aig_t) {
    var pNode, pFanin, pFanout *Abc_Obj_t;
    var vVec *Vec_Ptr_t;
    var LevelNew, i, k, v, j int32;

    // go through the nodes and update the level of their fanouts
    // inlining of macro Vec_VecForEachLevel
    for  i = 0; (i < Vec_VecSize(pMan.vLevelsR)) && (((vVec) = Vec_VecEntry(pMan.vLevelsR, i)), 1); i++  {
        if Vec_PtrSize(vVec) == 0 {
            continue; }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vVec)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vVec, k)), 1); k++  {
            if pNode == nil {
                continue; }
            assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
            assert(Abc_ObjReverseLevel(pNode) == i, "Abc_ObjReverseLevel(pNode) == i");
            // clean the mark
            assert(pNode.fMarkB == 1, "pNode.fMarkB == 1");
            pNode.fMarkB = 0;
            // iterate through the fanins
            // inlining of macro Abc_ObjForEachFanin
            for  v = 0; (v < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, v)), 1); v++  {
                if Abc_ObjIsCi(pFanin) {
                    continue; }
                // get the new reverse level of this fanin
                LevelNew = 0;
                // inlining of macro Abc_ObjForEachFanout
                for  j = 0; (j < Abc_ObjFanoutNum(pFanin)) && (((pFanout) = Abc_ObjFanout(pFanin, j)), 1); j++  {
                    if LevelNew < Abc_ObjReverseLevel(pFanout) {
                        LevelNew = Abc_ObjReverseLevel(pFanout); }
                }
                LevelNew += 1;
                assert(LevelNew > i, "LevelNew > i");
                if Abc_ObjReverseLevel(pFanin) == LevelNew { // no change
                    continue; }
                // if the fanin is present in the data structure, pull it out
                if pFanin.fMarkB {
                    Abc_AigRemoveFromLevelStructureR( pMan.vLevelsR, pFanin ); }
                // update the reverse level
                Abc_ObjSetReverseLevel( pFanin, LevelNew );
                // add the fanin to the data structure to update its fanins
                assert(pFanin.fMarkB == 0, "pFanin.fMarkB == 0");
                pFanin.fMarkB = 1;
                Vec_VecPush( pMan.vLevelsR, LevelNew, pFanin );
            }
        }
        Vec_PtrClear( vVec );
    }
}

/**Function*************************************************************

  Synopsis    [Removes the node from the level structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigRemoveFromLevelStructure(vStruct *Vec_Vec_t, pNode *Abc_Obj_t) {
    var vVecTemp *Vec_Ptr_t;
    var pTemp *Abc_Obj_t;
    var m int32;
    assert(pNode.fMarkA, "pNode.fMarkA");
    vVecTemp = Vec_VecEntry( vStruct, pNode.Level );
    // inlining of macro Vec_PtrForEachEntry
    for  m = 0; (m < Vec_PtrSize(vVecTemp)) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(vVecTemp, m)), 1); m++  {
        if pTemp != pNode {
            continue; }
        Vec_PtrWriteEntry( vVecTemp, m, nil );
        break;
    }
    assert(m < Vec_PtrSize(vVecTemp), "m < Vec_PtrSize(vVecTemp)"); // found
    pNode.fMarkA = 0;
}

/**Function*************************************************************

  Synopsis    [Removes the node from the level structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigRemoveFromLevelStructureR(vStruct *Vec_Vec_t, pNode *Abc_Obj_t) {
    var vVecTemp *Vec_Ptr_t;
    var pTemp *Abc_Obj_t;
    var m int32;
    assert(pNode.fMarkB, "pNode.fMarkB");
    vVecTemp = Vec_VecEntry( vStruct, Abc_ObjReverseLevel(pNode) );
    // inlining of macro Vec_PtrForEachEntry
    for  m = 0; (m < Vec_PtrSize(vVecTemp)) && (((pTemp) = (Abc_Obj_t *)Vec_PtrEntry(vVecTemp, m)), 1); m++  {
        if pTemp != pNode {
            continue; }
        Vec_PtrWriteEntry( vVecTemp, m, nil );
        break;
    }
    assert(m < Vec_PtrSize(vVecTemp), "m < Vec_PtrSize(vVecTemp)"); // found
    pNode.fMarkB = 0;
}




/**Function*************************************************************

  Synopsis    [Returns 1 if the node has at least one complemented fanout.]

  Description [A fanout is complemented if the fanout's fanin edge pointing
  to the given node is complemented.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigNodeHasComplFanoutEdge(pNode *Abc_Obj_t) int32 {
    var pFanout *Abc_Obj_t;
    var i, iFanin int32;
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, i)), 1); i++  {
        iFanin = Vec_IntFind( &pFanout.vFanins, pNode.Id );
        assert(iFanin >= 0, "iFanin >= 0");
        if Abc_ObjFaninC( pFanout, iFanin ) {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the node has at least one complemented fanout.]

  Description [A fanout is complemented if the fanout's fanin edge pointing
  to the given node is complemented. Only the fanouts with current TravId
  are counted.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigNodeHasComplFanoutEdgeTrav(pNode *Abc_Obj_t) int32 {
    var pFanout *Abc_Obj_t;
    var i, iFanin int32;
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pFanout) = Abc_ObjFanout(pNode, i)), 1); i++  {
        if !Abc_NodeIsTravIdCurrent(pFanout) {
            continue; }
        iFanin = Vec_IntFind( &pFanout.vFanins, pNode.Id );
        assert(iFanin >= 0, "iFanin >= 0");
        if Abc_ObjFaninC( pFanout, iFanin ) {
            return 1; }
    }
    return 0;
}


/**Function*************************************************************

  Synopsis    [Prints the AIG node for debugging purposes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigPrintNode(pNode *Abc_Obj_t) {
    var pNodeR *Abc_Obj_t  = Abc_ObjRegular(pNode);
    if Abc_ObjIsCi(pNodeR) {
        fmt.Printf("CI %4s%s.\n", Abc_ObjName(pNodeR), Abc_ObjIsComplement(pNode)? "\'" : "");
        return;
    }
    if Abc_AigNodeIsConst(pNodeR) {
        fmt.Printf("Constant 1 %s.\n", Abc_ObjIsComplement(pNode)? "(complemented)" : "");
        return;
    }
    // print the node's function
    fmt.Printf("%7s%s", Abc_ObjName(pNodeR), Abc_ObjIsComplement(pNode)? "\'" : "");
    fmt.Printf(" = ");
    fmt.Printf("%7s%s", Abc_ObjName(Abc_ObjFanin0(pNodeR)), Abc_ObjFaninC0(pNodeR)?     "\'" : "");
    fmt.Printf(" * ");
    fmt.Printf("%7s%s", Abc_ObjName(Abc_ObjFanin1(pNodeR)), Abc_ObjFaninC1(pNodeR)?     "\'" : "");
    fmt.Printf("\n");
}


/**Function*************************************************************

  Synopsis    [Check if the node has a combination loop of depth 1 or 2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigNodeIsAcyclic(pNode *Abc_Obj_t, pRoot *Abc_Obj_t) int32 {
    var pFanin0, pFanin1 *Abc_Obj_t;
    var pChild00, pChild01 *Abc_Obj_t;
    var pChild10, pChild11 *Abc_Obj_t;
    if !Abc_AigNodeIsAnd(pNode) {
        return 1; }
    pFanin0 = Abc_ObjFanin0(pNode);
    pFanin1 = Abc_ObjFanin1(pNode);
    if pRoot == pFanin0 || pRoot == pFanin1 {
        return 0; }
    if Abc_ObjIsCi(pFanin0) {
        pChild00 = nil;
        pChild01 = nil;
    
    } else {
        pChild00 = Abc_ObjFanin0(pFanin0);
        pChild01 = Abc_ObjFanin1(pFanin0);
        if pRoot == pChild00 || pRoot == pChild01 {
            return 0; }
    }
    if Abc_ObjIsCi(pFanin1) {
        pChild10 = nil;
        pChild11 = nil;
    
    } else {
        pChild10 = Abc_ObjFanin0(pFanin1);
        pChild11 = Abc_ObjFanin1(pFanin1);
        if pRoot == pChild10 || pRoot == pChild11 {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Resizes the hash table of AIG nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigCheckFaninOrder(pMan *Abc_Aig_t) {
    var pEnt *Abc_Obj_t;
    var i int32;
    for  i = 0; i < pMan.nBins; i++  {
        // inlining of macro Abc_AigBinForEachEntry
        for  pEnt = pMan.pBins[i];                                                   (pEnt != nil);                                                          pEnt = pEnt.pNext  {
            if Abc_ObjRegular(Abc_ObjChild0(pEnt)).Id > Abc_ObjRegular(Abc_ObjChild1(pEnt)).Id {
//                int i0 = Abc_ObjRegular(Abc_ObjChild0(pEnt))->Id;
//                int i1 = Abc_ObjRegular(Abc_ObjChild1(pEnt))->Id;
                fmt.Printf("Node %d has incorrect ordering of fanins.\n", pEnt.Id);
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Sets the correct phase of the nodes.]

  Description [The AIG nodes should be in the DFS order.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigSetNodePhases(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsDfsOrdered(pNtk), "Abc_NtkIsDfsOrdered(pNtk)");
    Abc_AigConst1(pNtk).fPhase = 1;
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pObj.fPhase = 0;
    }
    // inlining of macro Abc_NtkForEachLatchOutput
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)); i++  {                                                  if !(Abc_ObjIsLatch(Abc_NtkBox(pNtk, i)) && (((pObj) = Abc_ObjFanout0(Abc_NtkBox(pNtk, i))), 1)) {} else {
        pObj.fPhase = Abc_LatchIsInit1(pObj); }
    }
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        pObj.fPhase = (Abc_ObjFanin0(pObj).fPhase ^ Abc_ObjFaninC0(pObj)) & (Abc_ObjFanin1(pObj).fPhase ^ Abc_ObjFaninC1(pObj)); }
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        pObj.fPhase = (Abc_ObjFanin0(pObj).fPhase ^ Abc_ObjFaninC0(pObj));
    }
    // inlining of macro Abc_NtkForEachLatchInput
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)); i++  {                                                  if !(Abc_ObjIsLatch(Abc_NtkBox(pNtk, i)) && (((pObj) = Abc_ObjFanin0(Abc_NtkBox(pNtk, i))), 1)) {} else {
        pObj.fPhase = (Abc_ObjFanin0(pObj).fPhase ^ Abc_ObjFaninC0(pObj)); }
    }
}



/**Function*************************************************************

  Synopsis    [Start the update list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigUpdateStart(pMan *Abc_Aig_t, pvUpdatedNets **Vec_Ptr_t) *Vec_Ptr_t {
    assert(pMan.vAddedCells == nil, "pMan.vAddedCells == nil");
    pMan.vAddedCells  = Vec_PtrAlloc( 1000 );
    pMan.vUpdatedNets = Vec_PtrAlloc( 1000 );
    *pvUpdatedNets = pMan.vUpdatedNets;
    return pMan.vAddedCells;
}

/**Function*************************************************************

  Synopsis    [Start the update list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigUpdateStop(pMan *Abc_Aig_t) {
    assert(pMan.vAddedCells != nil, "pMan.vAddedCells != nil");
    Vec_PtrFree( pMan.vAddedCells );
    Vec_PtrFree( pMan.vUpdatedNets );
    pMan.vAddedCells = nil;
    pMan.vUpdatedNets = nil;
}

/**Function*************************************************************

  Synopsis    [Start the update list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigUpdateReset(pMan *Abc_Aig_t) {
    assert(pMan.vAddedCells != nil, "pMan.vAddedCells != nil");
    Vec_PtrClear( pMan.vAddedCells );
    Vec_PtrClear( pMan.vUpdatedNets );
}

/**Function*************************************************************

  Synopsis    [Start the update list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_AigCountNext(pMan *Abc_Aig_t) int32 {
    var pAnd *Abc_Obj_t;
    var i, Counter, CounterTotal int32;
    // count how many nodes have pNext set
    for  i = 0; i < pMan.nBins; i++  {
        // inlining of macro Abc_AigBinForEachEntry
        for  pAnd = pMan.pBins[i];                                                   (pAnd != nil);                                                          pAnd = pAnd.pNext  {
            Counter += (pAnd.pNext != nil);
            CounterTotal++;
        }
    }
    fmt.Printf("Counter = %d.  Nodes = %d.  Ave = %6.2f\n", Counter, CounterTotal, 1.0 * CounterTotal/pMan.nBins);
    return Counter;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


func Abc_NtkHelloWorld(pNtk *Abc_Ntk_t) {
    fmt.Printf("Hello, World!\n");
}


//ABC_NAMESPACE_IMPL_END


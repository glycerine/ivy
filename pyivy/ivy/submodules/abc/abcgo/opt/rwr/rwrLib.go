package rwr

import "fmt"

/**CFile****************************************************************

  FileName    [rwrLib.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rwrLib.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "rwr.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static Rwr_Node_t *        Rwr_ManTryNode( Rwr_Man_t * p, Rwr_Node_t * p0, Rwr_Node_t * p1, int fExor, int Level, int Volume );

// static void                Rwr_MarkUsed_rec( Rwr_Man_t * p, Rwr_Node_t * pNode );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Precomputes the forest in the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManPrecompute(p *Rwr_Man_t) {
    var p0, p1 *Rwr_Node_t;
    var i, k, Level, Volume int32;
    var LevelOld int32  = -1;
    var nNodes int32;

    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(p.vForest)) && (((p0) = (Rwr_Node_t *)Vec_PtrEntry(p.vForest, i)), 1); i++  {
    // inlining of macro Vec_PtrForEachEntryStart
    for  k = 1; (k < Vec_PtrSize(p.vForest)) && (((p1) = (Rwr_Node_t *)Vec_PtrEntry(p.vForest, k)), 1); k++  {
        if LevelOld < int32(p0.Level) {
            LevelOld = p0.Level;
            fmt.Printf("Starting level %d  (at %d nodes).\n", LevelOld+1, i);
            fmt.Printf("Considered = %5d M.   Found = %8d.   Classes = %6d.   Trying %7d.\n", p.nConsidered/1000000, p.vForest.nSize, p.nClasses, i);
        }

        if k == i {
            break; }
//        if ( p0->Level + p1->Level > 6 ) // hard
//            break;

        if p0.Level + p1.Level > 5 { // easy 
            break; }

//        if ( p0->Level + p1->Level > 6 || (p0->Level == 3 && p1->Level == 3) )
//            break;

        // compute the level and volume of the new nodes
        Level  = 1 + Abc_MaxInt( p0.Level, p1.Level );
        Volume = 1 + Rwr_ManNodeVolume( p, p0, p1 );
        // try four different AND nodes
        Rwr_ManTryNode( p,         p0 ,         p1 , 0, Level, Volume );
        Rwr_ManTryNode( p, Rwr_Not(p0),         p1 , 0, Level, Volume );
        Rwr_ManTryNode( p,         p0 , Rwr_Not(p1), 0, Level, Volume );
        Rwr_ManTryNode( p, Rwr_Not(p0), Rwr_Not(p1), 0, Level, Volume );
        // try EXOR
        Rwr_ManTryNode( p,         p0 ,         p1 , 1, Level, Volume + 1 );
        // report the progress
        if p.nConsidered % 50000000 == 0 {
            fmt.Printf("Considered = %5d M.   Found = %8d.   Classes = %6d.   Trying %7d.\n", p.nConsidered/1000000, p.vForest.nSize, p.nClasses, i); }
        // quit after some time
        if p.vForest.nSize == RWR_LIMIT + 5 {
            fmt.Printf("Considered = %5d M.   Found = %8d.   Classes = %6d.   Trying %7d.\n", p.nConsidered/1000000, p.vForest.nSize, p.nClasses, i);
            goto save;
        }
    }
    }
save :

    // mark the relevant ones
    Rwr_ManIncTravId( p );
    k = 5;
    nNodes = 0;
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 5; (i < Vec_PtrSize(p.vForest)) && (((p0) = (Rwr_Node_t *)Vec_PtrEntry(p.vForest, i)), 1); i++  {
        if p0.uTruth == p.puCanons[p0.uTruth] {
            Rwr_MarkUsed_rec( p, p0 );
            nNodes++;
        }
    }

    // compact the array by throwing away non-canonical
    k = 5;
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 5; (i < Vec_PtrSize(p.vForest)) && (((p0) = (Rwr_Node_t *)Vec_PtrEntry(p.vForest, i)), 1); i++  {
        if p0.fUsed {
            p.vForest.pArray[k] = p0;
            p0.Id = k++;
        }
    }
    p.vForest.nSize = k;
    fmt.Printf("Total canonical = %4d. Total used = %5d.\n", nNodes, p.vForest.nSize);
}

/**Function*************************************************************

  Synopsis    [Adds one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManTryNode(p *Rwr_Man_t, p0 *Rwr_Node_t, p1 *Rwr_Node_t, fExor int32, Level int32, Volume int32) *Rwr_Node_t {
    var pOld *Rwr_Node_t;
    var pNew *Rwr_Node_t;
    var ppPlace **Rwr_Node_t;
    var uTruth uint32;
    // compute truth table, level, volume
    p.nConsidered++;
    if fExor != 0 {
//        printf( "Considering EXOR of %d and %d.\n", p0->Id, p1->Id );
        uTruth = (p0.uTruth ^ p1.uTruth);
    
    } else {
        uTruth = (Rwr_IsComplement(p0)? ~Rwr_Regular(p0).uTruth : Rwr_Regular(p0).uTruth) & 
                 (Rwr_IsComplement(p1)? ~Rwr_Regular(p1).uTruth : Rwr_Regular(p1).uTruth) & 0xFFFF; }
    // skip non-practical classes
    if Level > 2 && !p.pPractical[p.puCanons[uTruth]] {
        return nil; }
    // enumerate through the nodes with the same canonical form
    ppPlace = p.pTable + uTruth;
    for  pOld = *ppPlace; (pOld != nil); ppPlace = &pOld.pNext, pOld = pOld.pNext  {
        if pOld.Level <  uint32(Level) && pOld.Volume < uint32(Volume) {
            return nil; }
        if pOld.Level == uint32(Level) && pOld.Volume < uint32(Volume) {
            return nil; }
//        if ( pOld->Level <  (unsigned)Level && pOld->Volume == (unsigned)Volume )
//            return NULL;
    }
/*
    // enumerate through the nodes with the opposite polarity
    for ( pOld = p->pTable[~uTruth & 0xFFFF]; pOld; pOld = pOld->pNext )
    {
        if ( pOld->Level <  (unsigned)Level && pOld->Volume < (unsigned)Volume )
            return NULL;
        if ( pOld->Level == (unsigned)Level && pOld->Volume < (unsigned)Volume )
            return NULL;
//        if ( pOld->Level <  (unsigned)Level && pOld->Volume == (unsigned)Volume )
//            return NULL;
    }
*/
    // count the classes
    if p.pTable[uTruth] == nil && p.puCanons[uTruth] == uTruth {
        p.nClasses++; }
    // create the new node
    pNew = (Rwr_Node_t *)Extra_MmFixedEntryFetch( p.pMmNode );
    pNew.Id     = p.vForest.nSize;
    pNew.TravId = 0;
    pNew.uTruth = uTruth;
    pNew.Level  = Level;
    pNew.Volume = Volume;
    pNew.fUsed  = 0;
    pNew.fExor  = fExor;
    pNew.p0     = p0;
    pNew.p1     = p1;
    pNew.pNext  = nil;
    Vec_PtrPush( p.vForest, pNew );
    *ppPlace     = pNew;
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Adds one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManAddNode(p *Rwr_Man_t, p0 *Rwr_Node_t, p1 *Rwr_Node_t, fExor int32, Level int32, Volume int32) *Rwr_Node_t {
    var pNew *Rwr_Node_t;
    var uTruth uint32;
    // compute truth table, leve, volume
    p.nConsidered++;
    if fExor != 0 {
        uTruth = (p0.uTruth ^ p1.uTruth);
    } else {
        uTruth = (Rwr_IsComplement(p0)? ~Rwr_Regular(p0).uTruth : Rwr_Regular(p0).uTruth) & 
                 (Rwr_IsComplement(p1)? ~Rwr_Regular(p1).uTruth : Rwr_Regular(p1).uTruth) & 0xFFFF; }
    // create the new node
    pNew = (Rwr_Node_t *)Extra_MmFixedEntryFetch( p.pMmNode );
    pNew.Id     = p.vForest.nSize;
    pNew.TravId = 0;
    pNew.uTruth = uTruth;
    pNew.Level  = Level;
    pNew.Volume = Volume;
    pNew.fUsed  = 0;
    pNew.fExor  = fExor;
    pNew.p0     = p0;
    pNew.p1     = p1;
    pNew.pNext  = nil;
    Vec_PtrPush( p.vForest, pNew );
    // do not add if the node is not essential
    if uTruth != p.puCanons[uTruth] {
        return pNew; }

    // add to the list
    p.nAdded++;
    if p.pTable[uTruth] == nil {
        p.nClasses++; }
    Rwr_ListAddToTail( p.pTable + uTruth, pNew );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Adds one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManAddVar(p *Rwr_Man_t, uTruth uint32, fPrecompute int32) *Rwr_Node_t {
    var pNew *Rwr_Node_t;
    pNew = (Rwr_Node_t *)Extra_MmFixedEntryFetch( p.pMmNode );
    pNew.Id     = p.vForest.nSize;
    pNew.TravId = 0;
    pNew.uTruth = uTruth;
    pNew.Level  = 0;
    pNew.Volume = 0;
    pNew.fUsed  = 1;
    pNew.fExor  = 0;
    pNew.p0     = nil;
    pNew.p1     = nil;    
    pNew.pNext  = nil;
    Vec_PtrPush( p.vForest, pNew );
    if fPrecompute != 0 {
        Rwr_ListAddToTail( p.pTable + uTruth, pNew ); }
    return pNew;
}



/**Function*************************************************************

  Synopsis    [Adds one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_MarkUsed_rec(p *Rwr_Man_t, pNode *Rwr_Node_t) {
    if pNode.fUsed || pNode.TravId == p.nTravIds {
        return; }
    pNode.TravId = p.nTravIds;
    pNode.fUsed = 1;
    Rwr_MarkUsed_rec( p, Rwr_Regular(pNode.p0) );
    Rwr_MarkUsed_rec( p, Rwr_Regular(pNode.p1) );
}

/**Function*************************************************************

  Synopsis    [Adds one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_Trav_rec(p *Rwr_Man_t, pNode *Rwr_Node_t, pVolume *int32) {
    if pNode.fUsed || pNode.TravId == p.nTravIds {
        return; }
    pNode.TravId = p.nTravIds;
    (*pVolume)++;
    if pNode.fExor {
        (*pVolume)++; }
    Rwr_Trav_rec( p, Rwr_Regular(pNode.p0), pVolume );
    Rwr_Trav_rec( p, Rwr_Regular(pNode.p1), pVolume );
}

/**Function*************************************************************

  Synopsis    [Adds one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManNodeVolume(p *Rwr_Man_t, p0 *Rwr_Node_t, p1 *Rwr_Node_t) int32 {
    var Volume int32;
    Rwr_ManIncTravId( p );
    Rwr_Trav_rec( p, p0, &Volume );
    Rwr_Trav_rec( p, p1, &Volume );
    return Volume;
}

/**Function*************************************************************

  Synopsis    [Adds one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManIncTravId(p *Rwr_Man_t) {
    var pNode *Rwr_Node_t;
    var i int32;
    if p.nTravIds++ < 0x8FFFFFFF {
        return; }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vForest)) && (((pNode) = (Rwr_Node_t *)Vec_PtrEntry(p.vForest, i)), 1); i++  {
        pNode.TravId = 0;
    }
    p.nTravIds = 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


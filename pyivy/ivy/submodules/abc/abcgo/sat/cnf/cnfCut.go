package cnf

import "fmt"

/**CFile****************************************************************

  FileName    [cnfCut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG-to-CNF conversion.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: cnfCut.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cnf.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates cut of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutAlloc(p *Cnf_Man_t, nLeaves int32) *Cnf_Cut_t {
    var pCut *Cnf_Cut_t;
    var nSize int32  = sizeof(Cnf_Cut_t) + sizeof(int) * nLeaves + sizeof(unsigned) * Abc_TruthWordNum(nLeaves);
    pCut = (Cnf_Cut_t *)Aig_MmFlexEntryFetch( p.pMemCuts, nSize );
    pCut.nFanins = nLeaves;
    pCut.nWords = Abc_TruthWordNum(nLeaves);
    pCut.vIsop[0] = pCut.vIsop[1] = nil;
    return pCut;
}

/**Function*************************************************************

  Synopsis    [Deallocates cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutFree(pCut *Cnf_Cut_t) {
    if pCut.vIsop[0] {
        Vec_IntFree( pCut.vIsop[0] ); }
    if pCut.vIsop[1] {
        Vec_IntFree( pCut.vIsop[1] ); }
}

/**Function*************************************************************

  Synopsis    [Creates cut for the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutCreate(p *Cnf_Man_t, pObj *Aig_Obj_t) *Cnf_Cut_t {
    var pCutBest *Dar_Cut_t;
    var pCut *Cnf_Cut_t;
    var pTruth *uint32;
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    pCutBest = Dar_ObjBestCut( pObj );
    assert(pCutBest != nil, "pCutBest != nil");
    assert(pCutBest.nLeaves <= 4, "pCutBest.nLeaves <= 4");
    pCut = Cnf_CutAlloc( p, pCutBest.nLeaves );
    memcpy( pCut.pFanins, pCutBest.pLeaves, sizeof(int) * pCutBest.nLeaves );
    pTruth = Cnf_CutTruth(pCut);
    *pTruth = (pCutBest.uTruth << 16) | pCutBest.uTruth;
    pCut.Cost = Cnf_CutSopCost( p, pCutBest );
    return pCut;
}

/**Function*************************************************************

  Synopsis    [Deallocates cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutPrint(pCut *Cnf_Cut_t) {
    var i int32;
    fmt.Printf("{");
    for  i = 0; i < pCut.nFanins; i++  {
        fmt.Printf("%d ", pCut.pFanins[i]);
    }
    fmt.Printf(" } ");
}

/**Function*************************************************************

  Synopsis    [Allocates cut of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutDeref(p *Cnf_Man_t, pCut *Cnf_Cut_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Cnf_CutForEachLeaf( p.pManAig, pCut, pObj, i )
    {
        assert(pObj.nRefs > 0, "pObj.nRefs > 0");
        pObj.nRefs--;
    }
}

/**Function*************************************************************

  Synopsis    [Allocates cut of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutRef(p *Cnf_Man_t, pCut *Cnf_Cut_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Cnf_CutForEachLeaf( p.pManAig, pCut, pObj, i )
    {
        pObj.nRefs++;
    }
}

/**Function*************************************************************

  Synopsis    [Allocates cut of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutUpdateRefs(p *Cnf_Man_t, pCut *Cnf_Cut_t, pCutFan *Cnf_Cut_t, pCutRes *Cnf_Cut_t) {
    Cnf_CutDeref( p, pCut );
    Cnf_CutDeref( p, pCutFan );
    Cnf_CutRef( p, pCutRes );
}

/**Function*************************************************************

  Synopsis    [Merges two arrays of integers.]

  Description [Returns the number of items.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutMergeLeaves(pCut *Cnf_Cut_t, pCutFan *Cnf_Cut_t, pFanins *int32) int32 {
    var i, k, nFanins int32;
    for  i = k = 0; i < pCut.nFanins && k < pCutFan.nFanins;  {
        if pCut.pFanins[i] == pCutFan.pFanins[k] {
            pFanins[nFanins++] = pCut.pFanins[i], i++, k++;
        } else if pCut.pFanins[i] < pCutFan.pFanins[k] {
            pFanins[nFanins++] = pCut.pFanins[i], i++;
        } else {
            pFanins[nFanins++] = pCutFan.pFanins[k], k++; }
    }
    for  ; i < pCut.nFanins; i++  {
        pFanins[nFanins++] = pCut.pFanins[i];
    }
    for  ; k < pCutFan.nFanins; k++  {
        pFanins[nFanins++] = pCutFan.pFanins[k];
    }
    return nFanins;
}

/**Function*************************************************************

  Synopsis    [Computes the stretching phase of the cut w.r.t. the merged cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_TruthPhase(pCut *Cnf_Cut_t, pCut1 *Cnf_Cut_t) uint32 {
    var uPhase uint32;
    var i, k int32;
    for  i = k = 0; i < pCut.nFanins; i++  {
        if k == pCut1.nFanins {
            break; }
        if pCut.pFanins[i] < pCut1.pFanins[k] {
            continue; }
        assert(pCut.pFanins[i] == pCut1.pFanins[k], "pCut.pFanins[i] == pCut1.pFanins[k]");
        uPhase |= (1 << i);
        k++;
    }
    return uPhase;
}

/**Function*************************************************************

  Synopsis    [Removes the fanin variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutRemoveIthVar(pCut *Cnf_Cut_t, iVar int32, iFan int32) {
    var i int32;
    assert(pCut.pFanins[iVar] == iFan, "pCut.pFanins[iVar] == iFan");
    pCut.nFanins--;
    for  i = iVar; i < pCut.nFanins; i++  {
        pCut.pFanins[i] = pCut.pFanins[i+1];
    }
}

/**Function*************************************************************

  Synopsis    [Inserts the fanin variable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cnf_CutInsertIthVar(pCut *Cnf_Cut_t, iVar int32, iFan int32) {
    var i int32;
    for  i = pCut.nFanins; i > iVar; i--  {
        pCut.pFanins[i] = pCut.pFanins[i-1];
    }
    pCut.pFanins[iVar] = iFan;
    pCut.nFanins++;
}

/**Function*************************************************************

  Synopsis    [Merges two cuts.]

  Description [Returns NULL of the cuts cannot be merged.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Cnf_CutCompose_pFanins [32]int32;
func Cnf_CutCompose(p *Cnf_Man_t, pCut *Cnf_Cut_t, pCutFan *Cnf_Cut_t, iFan int32) *Cnf_Cut_t {
    var pCutRes *Cnf_Cut_t;
    
    var pTruth, pTruthFan, pTruthRes *uint32;
    var pTop *uint32 = p.pTruths[0];
    var pFan *uint32 = p.pTruths[2];
    var pTemp *uint32 = p.pTruths[3];
    var uPhase, uPhaseFan uint32;
    var i, iVar, nFanins, RetValue int32;

    // make sure the second cut is the fanin of the first
    for  iVar = 0; iVar < pCut.nFanins; iVar++  {
        if pCut.pFanins[iVar] == iFan {
            break; }
    }
    assert(iVar < pCut.nFanins, "iVar < pCut.nFanins");
    // remove this variable
    Cnf_CutRemoveIthVar( pCut, iVar, iFan );
    // merge leaves of the cuts
    nFanins = Cnf_CutMergeLeaves( pCut, pCutFan, spStatic_Cnf_CutCompose_pFanins );
    if nFanins+1 > p.nMergeLimit {
        Cnf_CutInsertIthVar( pCut, iVar, iFan );
        return nil;
    }
    // create new cut
    pCutRes = Cnf_CutAlloc( p, nFanins );
    memcpy( pCutRes.pFanins, spStatic_Cnf_CutCompose_pFanins, sizeof(int) * nFanins );
    assert(pCutRes.nFanins <= pCut.nFanins + pCutFan.nFanins, "pCutRes.nFanins <= pCut.nFanins + pCutFan.nFanins");

    // derive its truth table
    // get the truth tables in the composition space
    pTruth    = Cnf_CutTruth(pCut);
    pTruthFan = Cnf_CutTruth(pCutFan);
    pTruthRes = Cnf_CutTruth(pCutRes);
    for  i = 0; i < 2*pCutRes.nWords; i++  {
        pTop[i] = pTruth[i % pCut.nWords];
    }
    for  i = 0; i < pCutRes.nWords; i++  {
        pFan[i] = pTruthFan[i % pCutFan.nWords];
    }
    // move the variable to the end
    uPhase = Kit_BitMask( pCutRes.nFanins+1 ) & ~(1 << iVar);
    Kit_TruthShrink( pTemp, pTop, pCutRes.nFanins, pCutRes.nFanins+1, uPhase, 1 );
    // compute the phases
    uPhase    = Cnf_TruthPhase( pCutRes, pCut    ) | (1 << pCutRes.nFanins);
    uPhaseFan = Cnf_TruthPhase( pCutRes, pCutFan );
    // permute truth-tables to the common support
    Kit_TruthStretch( pTemp, pTop, pCut.nFanins+1,  pCutRes.nFanins+1, uPhase,    1 );
    Kit_TruthStretch( pTemp, pFan, pCutFan.nFanins, pCutRes.nFanins,   uPhaseFan, 1 );
    // perform Boolean operation
    Kit_TruthMux( pTruthRes, pTop, pTop+pCutRes.nWords, pFan, pCutRes.nFanins );
    // return the cut to its original condition
    Cnf_CutInsertIthVar( pCut, iVar, iFan );
    // consider the simple case
    if pCutRes.nFanins < 5 {
        pCutRes.Cost = p.pSopSizes[0xFFFF & *pTruthRes] + p.pSopSizes[0xFFFF & ~*pTruthRes];
        return pCutRes;
    }

    // derive ISOP for positive phase
    RetValue = Kit_TruthIsop( pTruthRes, pCutRes.nFanins, p.vMemory, 0 );
    pCutRes.vIsop[1] = (RetValue == -1)? nil : Vec_IntDup( p.vMemory );
    // derive ISOP for negative phase
    Kit_TruthNot( pTruthRes, pTruthRes, pCutRes.nFanins );
    RetValue = Kit_TruthIsop( pTruthRes, pCutRes.nFanins, p.vMemory, 0 );
    pCutRes.vIsop[0] = (RetValue == -1)? nil : Vec_IntDup( p.vMemory );
    Kit_TruthNot( pTruthRes, pTruthRes, pCutRes.nFanins );

    // compute the cut cost
    if pCutRes.vIsop[0] == nil || pCutRes.vIsop[1] == nil {
        pCutRes.Cost = 127;
    } else if Vec_IntSize(pCutRes.vIsop[0]) + Vec_IntSize(pCutRes.vIsop[1]) > 127 {
        pCutRes.Cost = 127;
    } else {
        pCutRes.Cost = Vec_IntSize(pCutRes.vIsop[0]) + Vec_IntSize(pCutRes.vIsop[1]); }
    return pCutRes;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


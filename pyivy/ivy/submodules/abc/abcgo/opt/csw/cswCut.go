package csw

/**CFile****************************************************************

  FileName    [cswCut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Cut sweeping.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 11, 2007.]

  Revision    [$Id: cswCut.c,v 1.00 2007/07/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cswInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Compute the cost of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_CutFindCost(p *Csw_Man_t, pCut *Csw_Cut_t) int32 {
    var pLeaf *Aig_Obj_t;
    var i, Cost int32;
    assert(pCut.nFanins > 0, "pCut.nFanins > 0");
    // inlining of macro Csw_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nFanins)) && ((pLeaf) = Aig_ManObj(p.pManRes, (pCut).pFanins[i])); i++  {
//        Cost += pLeaf->nRefs;
        Cost += Csw_ObjRefs( p, pLeaf );
//        printf( "%d ", pLeaf->nRefs );
    }
//printf( "\n" );
    return Cost * 100 / pCut.nFanins;
}

/**Function*************************************************************

  Synopsis    [Compute the cost of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_CutFindCost2(p *Csw_Man_t, pCut *Csw_Cut_t) float32 {
    var pLeaf *Aig_Obj_t;
    var Cost float32  = 0.0;
    var i int32;
    assert(pCut.nFanins > 0, "pCut.nFanins > 0");
    // inlining of macro Csw_CutForEachLeaf
    for  i = 0; (i < int32((pCut).nFanins)) && ((pLeaf) = Aig_ManObj(p.pManRes, (pCut).pFanins[i])); i++  {
        Cost += float32(1.0)/pLeaf.nRefs;
    }
    return 1/Cost;
}

/**Function*************************************************************

  Synopsis    [Returns the next free cut to use.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_CutFindFree(p *Csw_Man_t, pObj *Aig_Obj_t) *Csw_Cut_t {
    var pCut, pCutMax *Csw_Cut_t;
    var i int32;
    pCutMax = nil;
    // inlining of macro Csw_ObjForEachCut
    for  i = 0, pCut = Csw_ObjCuts(p, pObj); i < p.nCutsMax; i++, pCut = Csw_CutNext(pCut)  {
        if pCut.nFanins == 0 {
            return pCut; }
        if pCutMax == nil || pCutMax.Cost < pCut.Cost {
            pCutMax = pCut; }
    }
    assert(pCutMax != nil, "pCutMax != nil");
    pCutMax.nFanins = 0;
    return pCutMax;
}

/**Function*************************************************************

  Synopsis    [Computes the stretching phase of the cut w.r.t. the merged cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_TruthPhase(pCut *Csw_Cut_t, pCut1 *Csw_Cut_t) uint32 {
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

  Synopsis    [Performs truth table computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_CutComputeTruth(p *Csw_Man_t, pCut *Csw_Cut_t, pCut0 *Csw_Cut_t, pCut1 *Csw_Cut_t, fCompl0 int32, fCompl1 int32) *uint32 {
    // permute the first table
    if fCompl0 != 0 { 
        Kit_TruthNot( p.puTemp[0], Csw_CutTruth(pCut0), p.nLeafMax );
    } else {
        Kit_TruthCopy( p.puTemp[0], Csw_CutTruth(pCut0), p.nLeafMax ); }
    Kit_TruthStretch( p.puTemp[2], p.puTemp[0], pCut0.nFanins, p.nLeafMax, Cut_TruthPhase(pCut, pCut0), 0 );
    // permute the second table
    if fCompl1 != 0 { 
        Kit_TruthNot( p.puTemp[1], Csw_CutTruth(pCut1), p.nLeafMax );
    } else {
        Kit_TruthCopy( p.puTemp[1], Csw_CutTruth(pCut1), p.nLeafMax ); }
    Kit_TruthStretch( p.puTemp[3], p.puTemp[1], pCut1.nFanins, p.nLeafMax, Cut_TruthPhase(pCut, pCut1), 0 );
    // produce the resulting table
    Kit_TruthAnd( Csw_CutTruth(pCut), p.puTemp[2], p.puTemp[3], p.nLeafMax );
//    assert( pCut->nFanins >= Kit_TruthSupportSize( Csw_CutTruth(pCut), p->nLeafMax ) );
    return Csw_CutTruth(pCut);
}

/**Function*************************************************************

  Synopsis    [Performs support minimization for the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_CutSupportMinimize(p *Csw_Man_t, pCut *Csw_Cut_t) int32 {
    var pTruth *uint32;
    var uSupp, nFansNew, i, k int32;
    // get truth table
    pTruth = Csw_CutTruth( pCut );
    // get support 
    uSupp = Kit_TruthSupport( pTruth, p.nLeafMax );
    // get the new support size
    nFansNew = Kit_WordCountOnes( uSupp );
    // check if there are redundant variables
    if nFansNew == pCut.nFanins {
        return nFansNew; }
    assert(nFansNew < pCut.nFanins, "nFansNew < pCut.nFanins");
    // minimize support
    Kit_TruthShrink( p.puTemp[0], pTruth, nFansNew, p.nLeafMax, uSupp, 1 );
    for  i = k = 0; i < pCut.nFanins; i++  {
        if uSupp & (1 << i) {
            pCut.pFanins[k++] = pCut.pFanins[i]; }
    }
    assert(k == nFansNew, "k == nFansNew");
    pCut.nFanins = nFansNew;
//    assert( nFansNew == Kit_TruthSupportSize( pTruth, p->nLeafMax ) );
//Extra_PrintBinary( stdout, pTruth, (1<<p->nLeafMax) ); printf( "\n" );
    return nFansNew;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pDom is contained in pCut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_CutCheckDominance(pDom *Csw_Cut_t, pCut *Csw_Cut_t) int32 {
    var i, k int32;
    for  i = 0; i < int32(pDom.nFanins); i++  {
        for  k = 0; k < int32(pCut.nFanins); k++  {
            if pDom.pFanins[i] == pCut.pFanins[k] {
                break; }
        }
        if k == int32(pCut.nFanins) { // node i in pDom is not contained in pCut
            return 0; }
    }
    // every node in pDom is contained in pCut
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the cut is contained.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_CutFilter(p *Csw_Man_t, pObj *Aig_Obj_t, pCut *Csw_Cut_t) int32 { 
    var pTemp *Csw_Cut_t;
    var i int32;
    // go through the cuts of the node
    // inlining of macro Csw_ObjForEachCut
    for  i = 0, pTemp = Csw_ObjCuts(p, pObj); i < p.nCutsMax; i++, pTemp = Csw_CutNext(pTemp)  {
        if pTemp.nFanins < 2 {
            continue; }
        if pTemp == pCut {
            continue; }
        if pTemp.nFanins > pCut.nFanins {
            // skip the non-contained cuts
            if (pTemp.uSign & pCut.uSign) != pCut.uSign {
                continue; }
            // check containment seriously
            if Csw_CutCheckDominance( pCut, pTemp ) != 0 {
                // remove contained cut
                pTemp.nFanins = 0;
            }
         
        } else {
            // skip the non-contained cuts
            if (pTemp.uSign & pCut.uSign) != pTemp.uSign {
                continue; }
            // check containment seriously
            if Csw_CutCheckDominance( pTemp, pCut ) != 0 {
                // remove the given
                pCut.nFanins = 0;
                return 1;
            }
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Merges two cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_CutMergeOrdered(p *Csw_Man_t, pC0 *Csw_Cut_t, pC1 *Csw_Cut_t, pC *Csw_Cut_t) int32 { 
    var i, k, c int32;
    assert(pC0.nFanins >= pC1.nFanins, "pC0.nFanins >= pC1.nFanins");
    // the case of the largest cut sizes
    if pC0.nFanins == p.nLeafMax && pC1.nFanins == p.nLeafMax {
        for  i = 0; i < pC0.nFanins; i++  {
            if pC0.pFanins[i] != pC1.pFanins[i] {
                return 0; }
        }
        for  i = 0; i < pC0.nFanins; i++  {
            pC.pFanins[i] = pC0.pFanins[i];
        }
        pC.nFanins = pC0.nFanins;
        return 1;
    }
    // the case when one of the cuts is the largest
    if pC0.nFanins == p.nLeafMax {
        for  i = 0; i < pC1.nFanins; i++  {
            for  k = pC0.nFanins - 1; k >= 0; k--  {
                if pC0.pFanins[k] == pC1.pFanins[i] {
                    break; }
            }
            if k == -1 { // did not find
                return 0; }
        }
        for  i = 0; i < pC0.nFanins; i++  {
            pC.pFanins[i] = pC0.pFanins[i];
        }
        pC.nFanins = pC0.nFanins;
        return 1;
    }

    // compare two cuts with different numbers
    i = k = 0;
    for  c = 0; c < p.nLeafMax; c++  {
        if k == pC1.nFanins {
            if i == pC0.nFanins {
                pC.nFanins = c;
                return 1;
            }
            pC.pFanins[c] = pC0.pFanins[i]; i++;
            continue;
        }
        if i == pC0.nFanins {
            if k == pC1.nFanins {
                pC.nFanins = c;
                return 1;
            }
            pC.pFanins[c] = pC1.pFanins[k]; k++;
            continue;
        }
        if pC0.pFanins[i] < pC1.pFanins[k] {
            pC.pFanins[c] = pC0.pFanins[i]; i++;
            continue;
        }
        if pC0.pFanins[i] > pC1.pFanins[k] {
            pC.pFanins[c] = pC1.pFanins[k]; k++;
            continue;
        }
        pC.pFanins[c] = pC0.pFanins[i]; i++; 
        k++;
    }
    if i < pC0.nFanins || k < pC1.nFanins {
        return 0; }
    pC.nFanins = c;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Prepares the object for FPGA mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_CutMerge(p *Csw_Man_t, pCut0 *Csw_Cut_t, pCut1 *Csw_Cut_t, pCut *Csw_Cut_t) int32 { 
    assert(p.nLeafMax > 0, "p.nLeafMax > 0");
    // merge the nodes
    if pCut0.nFanins < pCut1.nFanins {
        if !(Csw_CutMergeOrdered( p, pCut1, pCut0, pCut ) != 0) {
            return 0; }
    
    } else {
        if !(Csw_CutMergeOrdered( p, pCut0, pCut1, pCut ) != 0) {
            return 0; }
    }
    pCut.uSign = pCut0.uSign | pCut1.uSign;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Consider cut with more than 2 fanins having 2 true variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_ObjTwoVarCut(p *Csw_Man_t, pCut *Csw_Cut_t) *Aig_Obj_t {
    var pRes, pIn0, pIn1 *Aig_Obj_t;
    var nVars, uTruth, fCompl int32;
    assert(pCut.nFanins > 2, "pCut.nFanins > 2");
    // minimize support of this cut
    nVars = Csw_CutSupportMinimize( p, pCut );
    assert(nVars == 2, "nVars == 2");
    // get the fanins
    pIn0 = Aig_ManObj( p.pManRes, pCut.pFanins[0] );
    pIn1 = Aig_ManObj( p.pManRes, pCut.pFanins[1] );
    // derive the truth table
    uTruth = 0xF & *Csw_CutTruth(pCut);
    if uTruth == 14 || uTruth == 13 || uTruth == 11 || uTruth == 7 {
        uTruth = 0xF & ~uTruth;
        fCompl = 1;
    }
    // compute the result
    pRes = nil;
    if uTruth == 1 {  // 0001  // 1110  14
        pRes = Aig_And( p.pManRes, Aig_Not(pIn0), Aig_Not(pIn1) ); }
    if uTruth == 2 {  // 0010  // 1101  13 
        pRes = Aig_And( p.pManRes,         pIn0 , Aig_Not(pIn1) ); }
    if uTruth == 4 {  // 0100  // 1011  11
        pRes = Aig_And( p.pManRes, Aig_Not(pIn0),         pIn1  ); }
    if uTruth == 8 {  // 1000  // 0111   7
        pRes = Aig_And( p.pManRes,         pIn0 ,         pIn1  ); }
    if pRes {
        pRes = Aig_NotCond( pRes, fCompl ); }
    return pRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_ObjPrepareCuts(p *Csw_Man_t, pObj *Aig_Obj_t, fTriv int32) *Csw_Cut_t {
    var pCutSet, pCut *Csw_Cut_t;
    var i int32;
    // create the cutset of the node
    pCutSet = (Csw_Cut_t *)Aig_MmFixedEntryFetch( p.pMemCuts );
    Csw_ObjSetCuts( p, pObj, pCutSet );
    // inlining of macro Csw_ObjForEachCut
    for  i = 0, pCut = Csw_ObjCuts(p, pObj); i < p.nCutsMax; i++, pCut = Csw_CutNext(pCut)  {
        pCut.nFanins = 0;
        pCut.iNode = pObj.Id;
        pCut.nCutSize = p.nCutSize;
        pCut.nLeafMax = p.nLeafMax;
    }
    // add unit cut if needed
    if fTriv != 0 {
        pCut = pCutSet;
        pCut.Cost = 0;
        pCut.iNode = pObj.Id;
        pCut.nFanins = 1;
        pCut.pFanins[0] = pObj.Id;
        pCut.uSign = Aig_ObjCutSign( pObj.Id );
        memset( Csw_CutTruth(pCut), 0xAA, sizeof(unsigned) * p.nTruthWords );
    }
    return pCutSet;
}

/**Function*************************************************************

  Synopsis    [Derives cuts for one node and sweeps this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Csw_ObjSweep(p *Csw_Man_t, pObj *Aig_Obj_t, fTriv int32) *Aig_Obj_t {
    var fUseResub int32  = 1;
    var pCut0, pCut1, pCut, pCutSet *Csw_Cut_t;
    var pFanin0 *Aig_Obj_t  = Aig_ObjFanin0(pObj);
    var pFanin1 *Aig_Obj_t  = Aig_ObjFanin1(pObj);
    var pObjNew *Aig_Obj_t;
    var pTruth *uint32;
    var i, k, nVars, nFanins, iVar int32;
    var clk abctime;

    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if !Aig_ObjIsNode(pObj) {
        return pObj; }
    if Csw_ObjCuts(p, pObj) {
        return pObj; }
    // the node is not processed yet
    assert(Csw_ObjCuts(p, pObj) == nil, "Csw_ObjCuts(p, pObj) == nil");
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");

    // set up the first cut
    pCutSet = Csw_ObjPrepareCuts( p, pObj, fTriv );

    // compute pair-wise cut combinations while checking table
    // inlining of macro Csw_ObjForEachCut
    for  i = 0, pCut0 = Csw_ObjCuts(p, pFanin0); i < p.nCutsMax; i++, pCut0 = Csw_CutNext(pCut0)  { 
    if pCut0.nFanins > 0 {
    // inlining of macro Csw_ObjForEachCut
    for  k = 0, pCut1 = Csw_ObjCuts(p, pFanin1); k < p.nCutsMax; k++, pCut1 = Csw_CutNext(pCut1)  { 
    if pCut1.nFanins > 0 {
        // make sure K-feasible cut exists
        if Kit_WordCountOnes(pCut0.uSign | pCut1.uSign) > p.nLeafMax {
            continue; }
        // get the next cut of this node
        pCut = Csw_CutFindFree( p, pObj );
clk = Abc_Clock();
        // assemble the new cut
        if !(Csw_CutMerge( p, pCut0, pCut1, pCut ) != 0) {
            assert(pCut.nFanins == 0, "pCut.nFanins == 0");
            continue;
        }
        // check containment
        if Csw_CutFilter( p, pObj, pCut ) != 0 {
            assert(pCut.nFanins == 0, "pCut.nFanins == 0");
            continue;
        }
        // create its truth table
        pTruth = Csw_CutComputeTruth( p, pCut, pCut0, pCut1, Aig_ObjFaninC0(pObj), Aig_ObjFaninC1(pObj) );
        // support minimize the truth table
        nFanins = pCut.nFanins;
//        nVars = Csw_CutSupportMinimize( p, pCut ); // leads to quality degradation
        nVars = Kit_TruthSupportSize( pTruth, p.nLeafMax );
p.timeCuts += Abc_Clock() - clk;

        // check for trivial truth tables
        if nVars == 0 {
            p.nNodesTriv0++;
            return Aig_NotCond( Aig_ManConst1(p.pManRes), !(pTruth[0] & 1) );
        }
        if nVars == 1 {
            p.nNodesTriv1++;
            iVar = Kit_WordFindFirstBit( Kit_TruthSupport(pTruth, p.nLeafMax) );
            assert(iVar < pCut.nFanins, "iVar < pCut.nFanins");
            return Aig_NotCond( Aig_ManObj(p.pManRes, pCut.pFanins[iVar]), (pTruth[0] & 1) );
        }
        if nVars == 2 && nFanins > 2 && (fUseResub != 0) {
            if pObjNew = Csw_ObjTwoVarCut( p, pCut ) {
                p.nNodesTriv2++;
                return pObjNew;
            }
        }

        // check if an equivalent node with the same cut exists
clk = Abc_Clock();
        if pCut.nFanins > 2 {
            pObjNew = Csw_TableCutLookup( p, pCut );
        } else {
            pObjNew = nil;
        }
p.timeHash += Abc_Clock() - clk;
        if pObjNew != nil {
            p.nNodesCuts++;
            return pObjNew;
        }

        // assign the cost
        pCut.Cost = Csw_CutFindCost( p, pCut );
        assert(pCut.nFanins > 0, "pCut.nFanins > 0");
        assert(pCut.Cost > 0, "pCut.Cost > 0");
    }
    } }
    }
    p.nNodesTried++;

    // load the resulting cuts into the table
clk = Abc_Clock();
    // inlining of macro Csw_ObjForEachCut
    for  i = 0, pCut = Csw_ObjCuts(p, pObj); i < p.nCutsMax; i++, pCut = Csw_CutNext(pCut)  {
        if pCut.nFanins > 2 {
            assert(pCut.Cost > 0, "pCut.Cost > 0");
            Csw_TableCutInsert( p, pCut );
        }
    }
p.timeHash += Abc_Clock() - clk;

    // return the node if could not replace it
    return pObj;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


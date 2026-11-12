package aig

import "fmt"

/**CFile****************************************************************

  FileName    [aigCuts.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [AIG package.]

  Synopsis    [Computation of K-feasible priority cuts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: aigCuts.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the cut sweeping manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCutStart(pMan *Aig_Man_t, nCutsMax int32, nLeafMax int32, fTruth int32, fVerbose int32) *Aig_ManCut_t {
    var p *Aig_ManCut_t;
    assert(nCutsMax >= 2, "nCutsMax >= 2");
    assert(nLeafMax <= 16, "nLeafMax <= 16");
    // allocate the fraiging manager
    p = new(Aig_ManCut_t);
    memset( p, 0, sizeof(Aig_ManCut_t) );
    p.nCutsMax = nCutsMax;
    p.nLeafMax = nLeafMax;
    p.fTruth   = fTruth;
    p.fVerbose = fVerbose;
    p.pAig     = pMan;
    p.pCuts    = ABC_CALLOC( Aig_Cut_t *, Aig_ManObjNumMax(pMan) );
    // allocate memory manager
    p.nTruthWords = Abc_TruthWordNum(nLeafMax);
    p.nCutSize = sizeof(Aig_Cut_t) + sizeof(int) * nLeafMax + fTruth * sizeof(unsigned) * p.nTruthWords;
    p.pMemCuts = Aig_MmFixedStart( p.nCutSize * p.nCutsMax, 512 );
    // room for temporary truth tables
    if fTruth {
        p.puTemp[0] = ABC_ALLOC( unsigned, 4 * p.nTruthWords );
        p.puTemp[1] = p.puTemp[0] + p.nTruthWords;
        p.puTemp[2] = p.puTemp[1] + p.nTruthWords;
        p.puTemp[3] = p.puTemp[2] + p.nTruthWords;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the fraiging manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCutStop(p *Aig_ManCut_t) {
    Aig_MmFixedStop( p.pMemCuts, 0 );
    ABC_FREE( p.puTemp[0] );
    ABC_FREE( p.pCuts );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Prints one cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CutPrint(pCut *Aig_Cut_t) {
    var i int32;
    fmt.Printf("{");
    for  i = 0; i < pCut.nFanins; i++  {
        fmt.Printf(" %d", pCut.pFanins[i]);
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    [Prints one cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjCutPrint(p *Aig_ManCut_t, pObj *Aig_Obj_t) {
    var pCut *Aig_Cut_t;
    var i int32;
    fmt.Printf("Cuts for node %d:\n", pObj.Id);
    Aig_ObjForEachCut( p, pObj, pCut, i )
        if ( pCut.nFanins )
            Aig_CutPrint( pCut );
//    printf( "\n" );
}

/**Function*************************************************************

  Synopsis    [Computes the total number of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCutCount(p *Aig_ManCut_t, pnCutsK *int32) int32 {
    var pCut *Aig_Cut_t;
    var pObj *Aig_Obj_t;
    var i, k, nCuts, nCutsK int32;
    Aig_ManForEachNode( p.pAig, pObj, i )
        Aig_ObjForEachCut( p, pObj, pCut, k )
        {
            if pCut.nFanins == 0 {
                continue; }
            nCuts++;
            if pCut.nFanins == p.nLeafMax {
                nCutsK++; }
        }
    if pnCutsK != nil {
        *pnCutsK = nCutsK; }
    return nCuts;
}

/**Function*************************************************************

  Synopsis    [Compute the cost of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CutFindCost(p *Aig_ManCut_t, pCut *Aig_Cut_t) int32 {
    var pLeaf *Aig_Obj_t;
    var i, Cost int32;
    assert(pCut.nFanins > 0, "pCut.nFanins > 0");
    Aig_CutForEachLeaf( p.pAig, pCut, pLeaf, i )
        Cost += pLeaf.nRefs;
    return Cost * 1000 / pCut.nFanins;
}

/**Function*************************************************************

  Synopsis    [Compute the cost of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CutFindCost2(p *Aig_ManCut_t, pCut *Aig_Cut_t) float32 {
    var pLeaf *Aig_Obj_t;
    var Cost float32  = 0.0;
    var i int32;
    assert(pCut.nFanins > 0, "pCut.nFanins > 0");
    Aig_CutForEachLeaf( p.pAig, pCut, pLeaf, i )
        Cost += float32(1.0)/pLeaf.nRefs;
    return 1/Cost;
}

/**Function*************************************************************

  Synopsis    [Returns the next free cut to use.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CutFindFree(p *Aig_ManCut_t, pObj *Aig_Obj_t) *Aig_Cut_t {
    var pCut, pCutMax *Aig_Cut_t;
    var i int32;
    pCutMax = nil;
    Aig_ObjForEachCut( p, pObj, pCut, i )
    {
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
func Aig_CutTruthPhase(pCut *Aig_Cut_t, pCut1 *Aig_Cut_t) uint32 {
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
func Aig_CutComputeTruth(p *Aig_ManCut_t, pCut *Aig_Cut_t, pCut0 *Aig_Cut_t, pCut1 *Aig_Cut_t, fCompl0 int32, fCompl1 int32) *uint32 {
    // permute the first table
    if fCompl0 != 0 { 
        Kit_TruthNot( p.puTemp[0], Aig_CutTruth(pCut0), p.nLeafMax );
    } else {
        Kit_TruthCopy( p.puTemp[0], Aig_CutTruth(pCut0), p.nLeafMax ); }
    Kit_TruthStretch( p.puTemp[2], p.puTemp[0], pCut0.nFanins, p.nLeafMax, Aig_CutTruthPhase(pCut, pCut0), 0 );
    // permute the second table
    if fCompl1 != 0 { 
        Kit_TruthNot( p.puTemp[1], Aig_CutTruth(pCut1), p.nLeafMax );
    } else {
        Kit_TruthCopy( p.puTemp[1], Aig_CutTruth(pCut1), p.nLeafMax ); }
    Kit_TruthStretch( p.puTemp[3], p.puTemp[1], pCut1.nFanins, p.nLeafMax, Aig_CutTruthPhase(pCut, pCut1), 0 );
    // produce the resulting table
    Kit_TruthAnd( Aig_CutTruth(pCut), p.puTemp[2], p.puTemp[3], p.nLeafMax );
//    assert( pCut->nFanins >= Kit_TruthSupportSize( Aig_CutTruth(pCut), p->nLeafMax ) );
    return Aig_CutTruth(pCut);
}

/**Function*************************************************************

  Synopsis    [Performs support minimization for the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_CutSupportMinimize(p *Aig_ManCut_t, pCut *Aig_Cut_t) int32 {
    var pTruth *uint32;
    var uSupp, nFansNew, i, k int32;
    // get truth table
    pTruth = Aig_CutTruth( pCut );
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
func Aig_CutCheckDominance(pDom *Aig_Cut_t, pCut *Aig_Cut_t) int32 {
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
func Aig_CutFilter(p *Aig_ManCut_t, pObj *Aig_Obj_t, pCut *Aig_Cut_t) int32 { 
    var pTemp *Aig_Cut_t;
    var i int32;
    // go through the cuts of the node
    Aig_ObjForEachCut( p, pObj, pTemp, i )
    {
        if pTemp.nFanins < 2 {
            continue; }
        if pTemp == pCut {
            continue; }
        if pTemp.nFanins > pCut.nFanins {
            // skip the non-contained cuts
            if (pTemp.uSign & pCut.uSign) != pCut.uSign {
                continue; }
            // check containment seriously
            if Aig_CutCheckDominance( pCut, pTemp ) != 0 {
                // remove contained cut
                pTemp.nFanins = 0;
            }
         
        } else {
            // skip the non-contained cuts
            if (pTemp.uSign & pCut.uSign) != pTemp.uSign {
                continue; }
            // check containment seriously
            if Aig_CutCheckDominance( pTemp, pCut ) != 0 {
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
func Aig_CutMergeOrdered(p *Aig_ManCut_t, pC0 *Aig_Cut_t, pC1 *Aig_Cut_t, pC *Aig_Cut_t) int32 { 
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
func Aig_CutMerge(p *Aig_ManCut_t, pCut0 *Aig_Cut_t, pCut1 *Aig_Cut_t, pCut *Aig_Cut_t) int32 { 
    assert(p.nLeafMax > 0, "p.nLeafMax > 0");
    // merge the nodes
    if pCut0.nFanins < pCut1.nFanins {
        if !(Aig_CutMergeOrdered( p, pCut1, pCut0, pCut ) != 0) {
            return 0; }
    
    } else {
        if !(Aig_CutMergeOrdered( p, pCut0, pCut1, pCut ) != 0) {
            return 0; }
    }
    pCut.uSign = pCut0.uSign | pCut1.uSign;
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjPrepareCuts(p *Aig_ManCut_t, pObj *Aig_Obj_t, fTriv int32) *Aig_Cut_t {
    var pCutSet, pCut *Aig_Cut_t;
    var i int32;
    // create the cutset of the node
    pCutSet = (Aig_Cut_t *)Aig_MmFixedEntryFetch( p.pMemCuts );
    Aig_ObjSetCuts( p, pObj, pCutSet );
    Aig_ObjForEachCut( p, pObj, pCut, i )
    {
        pCut.nFanins  = 0;
        pCut.iNode    = pObj.Id;
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
        if p.fTruth {
        memset( Aig_CutTruth(pCut), 0xAA, sizeof(unsigned) * p.nTruthWords ); }
    }
    return pCutSet;
}

/**Function*************************************************************

  Synopsis    [Derives cuts for one node and sweeps this node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ObjComputeCuts(p *Aig_ManCut_t, pObj *Aig_Obj_t, fTriv int32) {
    var pCut0, pCut1, pCut, pCutSet *Aig_Cut_t;
    var pFanin0 *Aig_Obj_t  = Aig_ObjFanin0(pObj);
    var pFanin1 *Aig_Obj_t  = Aig_ObjFanin1(pObj);
    var i, k int32;
    // the node is not processed yet
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    assert(Aig_ObjCuts(p, pObj) == nil, "Aig_ObjCuts(p, pObj) == nil");
    // set up the first cut
    pCutSet = Aig_ObjPrepareCuts( p, pObj, fTriv );
    // compute pair-wise cut combinations while checking table
    Aig_ObjForEachCut( p, pFanin0, pCut0, i )
    if ( pCut0.nFanins > 0 )
    Aig_ObjForEachCut( p, pFanin1, pCut1, k )
    if pCut1.nFanins > 0 {
        // make sure K-feasible cut exists
        if Kit_WordCountOnes(pCut0.uSign | pCut1.uSign) > p.nLeafMax {
            continue; }
        // get the next cut of this node
        pCut = Aig_CutFindFree( p, pObj );
        // assemble the new cut
        if !(Aig_CutMerge( p, pCut0, pCut1, pCut ) != 0) {
            assert(pCut.nFanins == 0, "pCut.nFanins == 0");
            continue;
        }
        // check containment
        if Aig_CutFilter( p, pObj, pCut ) != 0 {
            assert(pCut.nFanins == 0, "pCut.nFanins == 0");
            continue;
        }
        // create its truth table
        if p.fTruth {
            Aig_CutComputeTruth( p, pCut, pCut0, pCut1, Aig_ObjFaninC0(pObj), Aig_ObjFaninC1(pObj) ); }
        // assign the cost
        pCut.Cost = Aig_CutFindCost( p, pCut );
        assert(pCut.nFanins > 0, "pCut.nFanins > 0");
        assert(pCut.Cost > 0, "pCut.Cost > 0");
    }
}

/**Function*************************************************************

  Synopsis    [Computes the cuts for all nodes in the static AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ComputeCuts(pAig *Aig_Man_t, nCutsMax int32, nLeafMax int32, fTruth int32, fVerbose int32) *Aig_ManCut_t {
    var p *Aig_ManCut_t;
    var pObj *Aig_Obj_t;
    var i int32;
    var clk abctime  = Abc_Clock();
    assert(pAig.pManCuts == nil, "pAig.pManCuts == nil");
    // start the manager
    p = Aig_ManCutStart( pAig, nCutsMax, nLeafMax, fTruth, fVerbose );
    // set elementary cuts at the PIs
    Aig_ManForEachCi( pAig, pObj, i )
        Aig_ObjPrepareCuts( p, pObj, 1 );
    // process the nodes
    Aig_ManForEachNode( pAig, pObj, i )
        Aig_ObjComputeCuts( p, pObj, 1 );
    // print stats
    if fVerbose != 0 {
        var nCuts, nCutsK int32;
        nCuts = Aig_ManCutCount( p, &nCutsK );
        fmt.Printf("Nodes = %6d. Total cuts = %6d. %d-input cuts = %6d.\n", Aig_ManObjNum(pAig), nCuts, nLeafMax, nCutsK);
        fmt.Printf("Cut size = %2d. Truth size = %2d. Total mem = %5.2f MB  ", p.nCutSize, 4*p.nTruthWords, 1.0*Aig_MmFixedReadMemUsage(p.pMemCuts)/(1<<20));
        ABC_PRT( "Runtime", Abc_Clock() - clk );
/*
        Aig_ManForEachNode( pAig, pObj, i )
            if ( i % 300 == 0 )
                Aig_ObjCutPrint( p, pObj );
*/
    }
    // remember the cut manager
    pAig.pManCuts = p;
    return p;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


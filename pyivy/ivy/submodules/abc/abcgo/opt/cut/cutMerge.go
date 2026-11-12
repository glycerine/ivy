package cut

/**CFile****************************************************************

  FileName    [cutMerge.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [K-feasible cut computation package.]

  Synopsis    [Procedure to merge two cuts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cutMerge.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cutInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Merges two cuts.]

  Description [This procedure works.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Cut_CutMergeTwo2_M [7][3]int32 = [7][3]int32{{0},{0},{0},{0},{0},{0},{0}};
func Cut_CutMergeTwo2(p *Cut_Man_t, pCut0 *Cut_Cut_t, pCut1 *Cut_Cut_t) *Cut_Cut_t { 
    
    var pRes *Cut_Cut_t;
    var pRow *int32;
    var nLeaves0, nLeaves1, Limit int32;
    var i, k, Count, nNodes int32;

    assert(pCut0.nLeaves >= pCut1.nLeaves, "pCut0.nLeaves >= pCut1.nLeaves");

    // the case of the largest cut sizes
    Limit = p.pParams.nVarsMax;
    nLeaves0 = pCut0.nLeaves;
    nLeaves1 = pCut1.nLeaves;
    if nLeaves0 == Limit && nLeaves1 == Limit {
        for  i = 0; i < nLeaves0; i++  {
            if pCut0.pLeaves[i] != pCut1.pLeaves[i] {
                return nil; }
        }
        pRes = Cut_CutAlloc( p );
        for  i = 0; i < nLeaves0; i++  {
            pRes.pLeaves[i] = pCut0.pLeaves[i];
        }
        pRes.nLeaves = nLeaves0;
        return pRes;
    }
    // the case when one of the cuts is the largest
    if nLeaves0 == Limit {
        for  i = 0; i < nLeaves1; i++  {
            for  k = nLeaves0 - 1; k >= 0; k--  {
                if pCut0.pLeaves[k] == pCut1.pLeaves[i] {
                    break; }
            }
            if k == -1 { // did not find
                return nil; }
        }
        pRes = Cut_CutAlloc( p );
        for  i = 0; i < nLeaves0; i++  {
            pRes.pLeaves[i] = pCut0.pLeaves[i];
        }
        pRes.nLeaves = nLeaves0;
        return pRes;
    }
    // other cases
    nNodes = nLeaves0;
    for  i = 0; i < nLeaves1; i++  {
        for  k = nLeaves0 - 1; k >= 0; k--  {
            if pCut0.pLeaves[k] > pCut1.pLeaves[i] {
                continue; }
            if pCut0.pLeaves[k] < pCut1.pLeaves[i] {
                pRow = spStatic_Cut_CutMergeTwo2_M[k+1];
                if pRow[0] == 0 {
                    pRow[0] = pCut1.pLeaves[i], pRow[1] = 0;
                } else if pRow[1] == 0 {
                    pRow[1] = pCut1.pLeaves[i], pRow[2] = 0;
                } else if pRow[2] == 0 {
                    pRow[2] = pCut1.pLeaves[i];
                } else { 
                    assert(false, "0"); }
                if nNodes++; nNodes > Limit {
                    for  i = 0; i <= nLeaves0; i++  {
                        spStatic_Cut_CutMergeTwo2_M[i][0] = 0;
                    }
                    return nil;
                }
            }
            break;
        }
        if k == -1 {
            pRow = spStatic_Cut_CutMergeTwo2_M[0];
            if pRow[0] == 0 {
                pRow[0] = pCut1.pLeaves[i], pRow[1] = 0;
            } else if pRow[1] == 0 {
                pRow[1] = pCut1.pLeaves[i], pRow[2] = 0;
            } else if pRow[2] == 0 {
                pRow[2] = pCut1.pLeaves[i];
            } else { 
                assert(false, "0"); }
            if nNodes++; nNodes > Limit {
                for  i = 0; i <= nLeaves0; i++  {
                    spStatic_Cut_CutMergeTwo2_M[i][0] = 0;
                }
                return nil;
            }
            continue;
        }
    }

    pRes = Cut_CutAlloc( p );
    for  Count, i = 0, 0; i <= nLeaves0; i++  {
        if i > 0 {
            pRes.pLeaves[Count++] = pCut0.pLeaves[i-1]; }
        pRow = spStatic_Cut_CutMergeTwo2_M[i];
        if pRow[0] {
            pRes.pLeaves[Count++] = pRow[0];
            if pRow[1] {
                pRes.pLeaves[Count++] = pRow[1];
                if pRow[2] {
                    pRes.pLeaves[Count++] = pRow[2]; }
            }
            pRow[0] = 0;
        }
    }
    assert(Count == nNodes, "Count == nNodes");
    pRes.nLeaves = nNodes;
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Merges two cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutMergeTwo(p *Cut_Man_t, pCut0 *Cut_Cut_t, pCut1 *Cut_Cut_t) *Cut_Cut_t { 
    var pRes *Cut_Cut_t;
    var pLeaves *int32;
    var Limit, nLeaves0, nLeaves1 int32;
    var i, k, c int32;

    assert(pCut0.nLeaves >= pCut1.nLeaves, "pCut0.nLeaves >= pCut1.nLeaves");

    // consider two cuts
    nLeaves0 = pCut0.nLeaves;
    nLeaves1 = pCut1.nLeaves;

    // the case of the largest cut sizes
    Limit = p.pParams.nVarsMax;
    if nLeaves0 == Limit && nLeaves1 == Limit {
        for  i = 0; i < nLeaves0; i++  {
            if pCut0.pLeaves[i] != pCut1.pLeaves[i] {
                return nil; }
        }
        pRes = Cut_CutAlloc( p );
        for  i = 0; i < nLeaves0; i++  {
            pRes.pLeaves[i] = pCut0.pLeaves[i];
        }
        pRes.nLeaves = pCut0.nLeaves;
        return pRes;
    }
    // the case when one of the cuts is the largest
    if nLeaves0 == Limit {
        for  i = 0; i < nLeaves1; i++  {
            for  k = nLeaves0 - 1; k >= 0; k--  {
                if pCut0.pLeaves[k] == pCut1.pLeaves[i] {
                    break; }
            }
            if k == -1 { // did not find
                return nil; }
        }
        pRes = Cut_CutAlloc( p );
        for  i = 0; i < nLeaves0; i++  {
            pRes.pLeaves[i] = pCut0.pLeaves[i];
        }
        pRes.nLeaves = pCut0.nLeaves;
        return pRes;
    }

    // prepare the cut
    if p.pReady == nil {
        p.pReady = Cut_CutAlloc( p ); }
    pLeaves = p.pReady.pLeaves;

    // compare two cuts with different numbers
    i = k = 0;
    for  c = 0; c < Limit; c++  {
        if k == nLeaves1 {
            if i == nLeaves0 {
                p.pReady.nLeaves = c;
                pRes = p.pReady;  p.pReady = nil;
                return pRes;
            }
            pLeaves[c] = pCut0.pLeaves[i]; i++;
            continue;
        }
        if i == nLeaves0 {
            if k == nLeaves1 {
                p.pReady.nLeaves = c;
                pRes = p.pReady;  p.pReady = nil;
                return pRes;
            }
            pLeaves[c] = pCut1.pLeaves[k]; k++;
            continue;
        }
        if pCut0.pLeaves[i] < pCut1.pLeaves[k] {
            pLeaves[c] = pCut0.pLeaves[i]; i++;
            continue;
        }
        if pCut0.pLeaves[i] > pCut1.pLeaves[k] {
            pLeaves[c] = pCut1.pLeaves[k]; k++;
            continue;
        }
        pLeaves[c] = pCut0.pLeaves[i]; i++; 
        k++;
    }
    if i < nLeaves0 || k < nLeaves1 {
        return nil; }
    p.pReady.nLeaves = c;
    pRes = p.pReady;  p.pReady = nil;
    return pRes;
}


/**Function*************************************************************

  Synopsis    [Merges two cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutMergeTwo3(p *Cut_Man_t, pCut0 *Cut_Cut_t, pCut1 *Cut_Cut_t) *Cut_Cut_t { 
    var pRes *Cut_Cut_t;
    var pLeaves *int32;
    var Limit, nLeaves0, nLeaves1 int32;
    var i, k, c int32;

    assert(pCut0.nLeaves >= pCut1.nLeaves, "pCut0.nLeaves >= pCut1.nLeaves");

    // prepare the cut
    if p.pReady == nil {
        p.pReady = Cut_CutAlloc( p ); }
    pLeaves = p.pReady.pLeaves;

    // consider two cuts
    Limit = p.pParams.nVarsMax;
    nLeaves0 = pCut0.nLeaves;
    nLeaves1 = pCut1.nLeaves;
    if nLeaves0 == Limit { // the case when one of the cuts is the largest
        if nLeaves1 == Limit { // the case when both cuts are the largest
            for  i = 0; i < nLeaves0; i++  {
                pLeaves[i] = pCut0.pLeaves[i];
                if pLeaves[i] != pCut1.pLeaves[i] {
                    return nil; }
            }
        
        } else {
            for  i = k = 0; i < nLeaves0; i++  {
                pLeaves[i] = pCut0.pLeaves[i];
                if k == int32(nLeaves1) {
                    continue; }
                if pLeaves[i] < pCut1.pLeaves[k] {
                    continue; }
                if pLeaves[i] == pCut1.pLeaves[k++] {
                    continue; }
                return nil;
            }
            if k < nLeaves1 {
                return nil; }
        }
        p.pReady.nLeaves = nLeaves0;
        pRes = p.pReady;  p.pReady = nil;
        return pRes;
    }

    // compare two cuts with different numbers
    i = k = 0;
    for  c = 0; c < Limit; c++  {
        if k == nLeaves1 {
            if i == nLeaves0 {
                p.pReady.nLeaves = c;
                pRes = p.pReady;  p.pReady = nil;
                return pRes;
            }
            pLeaves[c] = pCut0.pLeaves[i]; i++;
            continue;
        }
        if i == nLeaves0 {
            if k == nLeaves1 {
                p.pReady.nLeaves = c;
                pRes = p.pReady;  p.pReady = nil;
                return pRes;
            }
            pLeaves[c] = pCut1.pLeaves[k]; k++;
            continue;
        }
        if pCut0.pLeaves[i] < pCut1.pLeaves[k] {
            pLeaves[c] = pCut0.pLeaves[i]; i++;
            continue;
        }
        if pCut0.pLeaves[i] > pCut1.pLeaves[k] {
            pLeaves[c] = pCut1.pLeaves[k]; k++;
            continue;
        }
        pLeaves[c] = pCut0.pLeaves[i]; i++; 
        k++;
    }
    if i < nLeaves0 || k < nLeaves1 {
        return nil; }
    p.pReady.nLeaves = c;
    pRes = p.pReady;  p.pReady = nil;
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Merges two cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cut_CutMergeTwo4(p *Cut_Man_t, pCut0 *Cut_Cut_t, pCut1 *Cut_Cut_t) *Cut_Cut_t { 
    var pRes *Cut_Cut_t;
    var pLeaves *int32;
    var i, k, min, NodeTemp, Limit, nTotal int32;

    assert(pCut0.nLeaves >= pCut1.nLeaves, "pCut0.nLeaves >= pCut1.nLeaves");

    // prepare the cut
    if p.pReady == nil {
        p.pReady = Cut_CutAlloc( p ); }
    pLeaves = p.pReady.pLeaves;

    // consider two cuts
    Limit = p.pParams.nVarsMax;
    if pCut0.nLeaves == uint32(Limit) { // the case when one of the cuts is the largest
        if pCut1.nLeaves == uint32(Limit) { // the case when both cuts are the largest
            for  i = 0; i < int32(pCut0.nLeaves); i++  {
                pLeaves[i] = pCut0.pLeaves[i];
                if pLeaves[i] != pCut1.pLeaves[i] {
                    return nil; }
            }
        
        } else {
            for  i = k = 0; i < int32(pCut0.nLeaves); i++  {
                pLeaves[i] = pCut0.pLeaves[i];
                if k == int32(pCut1.nLeaves) {
                    continue; }
                if pLeaves[i] < pCut1.pLeaves[k] {
                    continue; }
                if pLeaves[i] == pCut1.pLeaves[k++] {
                    continue; }
                return nil;
            }
            if k < int32(pCut1.nLeaves) {
                return nil; }
        }
        p.pReady.nLeaves = pCut0.nLeaves;
        pRes = p.pReady;  p.pReady = nil;
        return pRes;
    }

    // count the number of unique entries in pCut1
    nTotal = pCut0.nLeaves;
    for  i = 0; i < int32(pCut1.nLeaves); i++  {
        // try to find this entry among the leaves of pCut0
        for  k = 0; k < int32(pCut0.nLeaves); k++  {
            if pCut1.pLeaves[i] == pCut0.pLeaves[k] {
                break; }
        }
        if k < int32(pCut0.nLeaves) { // found
            continue; }
        // we found a new entry to add
        if nTotal == Limit {
            return nil; }
        pLeaves[nTotal++] = pCut1.pLeaves[i];
    }
    // we know that the feasible cut exists

    // add the starting entries
    for  k = 0; k < int32(pCut0.nLeaves); k++  {
        pLeaves[k] = pCut0.pLeaves[k];
    }

    // selection-sort the entries
    for  i = 0; i < nTotal - 1; i++  {
        min = i;
        for  k = i+1; k < nTotal; k++  {
            if pLeaves[k] < pLeaves[min] {
                min = k; }
        }
        NodeTemp     = pLeaves[i];
        pLeaves[i]   = pLeaves[min];
        pLeaves[min] = NodeTemp;
    }
    p.pReady.nLeaves = nTotal;
    pRes = p.pReady;  p.pReady = nil;
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Merges two cuts.]

  Description [This procedure works.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Cut_CutMergeTwo5_M [7][3]int32 = [7][3]int32{{0},{0},{0},{0},{0},{0},{0}};
func Cut_CutMergeTwo5(p *Cut_Man_t, pCut0 *Cut_Cut_t, pCut1 *Cut_Cut_t) *Cut_Cut_t { 
    
    var pRes *Cut_Cut_t;
    var pRow *int32;
    var uSign0, uSign1 uint32;
    var i, k, nNodes, Count int32;
    var Limit uint32  = p.pParams.nVarsMax;

    assert(pCut0.nLeaves >= pCut1.nLeaves, "pCut0.nLeaves >= pCut1.nLeaves");

    // the case of the largest cut sizes
    if pCut0.nLeaves == Limit && pCut1.nLeaves == Limit {
        for  i = 0; i < int32(pCut0.nLeaves); i++  {
            if pCut0.pLeaves[i] != pCut1.pLeaves[i] {
                return nil; }
        }
        pRes = Cut_CutAlloc( p );
        for  i = 0; i < int32(pCut0.nLeaves); i++  {
            pRes.pLeaves[i] = pCut0.pLeaves[i];
        }
        pRes.nLeaves = pCut0.nLeaves;
        return pRes;
    }
    // the case when one of the cuts is the largest
    if pCut0.nLeaves == Limit {
        if !p.pParams.fTruth {
            for  i = 0; i < int32(pCut1.nLeaves); i++  {
                for  k = pCut0.nLeaves - 1; k >= 0; k--  {
                    if pCut0.pLeaves[k] == pCut1.pLeaves[i] {
                        break; }
                }
                if k == -1 { // did not find
                    return nil; }
            }
            pRes = Cut_CutAlloc( p );
        
        } else {
            uSign1 = 0;
            for  i = 0; i < int32(pCut1.nLeaves); i++  {
                for  k = pCut0.nLeaves - 1; k >= 0; k--  {
                    if pCut0.pLeaves[k] == pCut1.pLeaves[i] {
                        uSign1 |= (1 << i);
                        break;
                    }
                }
                if k == -1 { // did not find
                    return nil; }
            }
            pRes = Cut_CutAlloc( p );
            pRes.Num1 = uSign1;
        }
        for  i = 0; i < int32(pCut0.nLeaves); i++  {
            pRes.pLeaves[i] = pCut0.pLeaves[i];
        }
        pRes.nLeaves = pCut0.nLeaves;
        return pRes;
    }
    // other cases
    nNodes = pCut0.nLeaves;
    for  i = 0; i < int32(pCut1.nLeaves); i++  {
        for  k = pCut0.nLeaves - 1; k >= 0; k--  {
            if pCut0.pLeaves[k] > pCut1.pLeaves[i] {
                continue; }
            if pCut0.pLeaves[k] < pCut1.pLeaves[i] {
                pRow = spStatic_Cut_CutMergeTwo5_M[k+1];
                if pRow[0] == 0 {
                    pRow[0] = pCut1.pLeaves[i], pRow[1] = 0;
                } else if pRow[1] == 0 {
                    pRow[1] = pCut1.pLeaves[i], pRow[2] = 0;
                } else if pRow[2] == 0 {
                    pRow[2] = pCut1.pLeaves[i];
                } else { 
                    assert(false, "0"); }
                if nNodes++; nNodes > int32(Limit) {
                    for  i = 0; i <= int32(pCut0.nLeaves); i++  {
                        spStatic_Cut_CutMergeTwo5_M[i][0] = 0;
                    }
                    return nil;
                }
            }
            break;
        }
        if k == -1 {
            pRow = spStatic_Cut_CutMergeTwo5_M[0];
            if pRow[0] == 0 {
                pRow[0] = pCut1.pLeaves[i], pRow[1] = 0;
            } else if pRow[1] == 0 {
                pRow[1] = pCut1.pLeaves[i], pRow[2] = 0;
            } else if pRow[2] == 0 {
                pRow[2] = pCut1.pLeaves[i];
            } else { 
                assert(false, "0"); }
            if nNodes++; nNodes > int32(Limit) {
                for  i = 0; i <= int32(pCut0.nLeaves); i++  {
                    spStatic_Cut_CutMergeTwo5_M[i][0] = 0;
                }
                return nil;
            }
            continue;
        }
    }

    pRes = Cut_CutAlloc( p );
    if !p.pParams.fTruth {
        for  Count, i = 0, 0; i <= int32(pCut0.nLeaves); i++  {
            if i > 0 {
                pRes.pLeaves[Count++] = pCut0.pLeaves[i-1]; }
            pRow = spStatic_Cut_CutMergeTwo5_M[i];
            if pRow[0] {
                pRes.pLeaves[Count++] = pRow[0];
                if pRow[1] {
                    pRes.pLeaves[Count++] = pRow[1];
                    if pRow[2] {
                        pRes.pLeaves[Count++] = pRow[2]; }
                }
                pRow[0] = 0;
            }
        }
        assert(Count == nNodes, "Count == nNodes");
        pRes.nLeaves = nNodes;
/*
    // make sure that the cut is correct
    {
        for ( i = 1; i < (int)pRes->nLeaves; i++ )
            if ( pRes->pLeaves[i-1] >= pRes->pLeaves[i] )
            {
                int v = 0;
            }
    }
*/
        return pRes;
    }

    uSign0 = uSign1 = 0;
    for  Count, i = 0, 0; i <= int32(pCut0.nLeaves); i++  {
        if i > 0 {
            uSign0 |= (1 << Count);
            pRes.pLeaves[Count++] = pCut1.pLeaves[i-1];
        }
        pRow = spStatic_Cut_CutMergeTwo5_M[i];
        if pRow[0] {
            uSign1 |= (1 << Count);
            pRes.pLeaves[Count++] = pRow[0];
            if pRow[1] {
                uSign1 |= (1 << Count);
                pRes.pLeaves[Count++] = pRow[1];
                if pRow[2] {
                    uSign1 |= (1 << Count);
                    pRes.pLeaves[Count++] = pRow[2];
                }
            }
            pRow[0] = 0;
        }
    }
    assert(Count == nNodes, "Count == nNodes");
    pRes.nLeaves = nNodes;
    pRes.Num1 = uSign1;
    pRes.Num0 = uSign0;
    return pRes;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


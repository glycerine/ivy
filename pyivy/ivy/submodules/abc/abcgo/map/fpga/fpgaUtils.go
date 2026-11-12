package fpga

import "fmt"

/**CFile****************************************************************

  FileName    [fpgaUtils.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Technology mapping for variable-size-LUT FPGAs.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - August 18, 2004.]

  Revision    [$Id: fpgaUtils.c,v 1.3 2004/07/06 04:55:58 alanmi Exp $]

***********************************************************************/

//#include "fpgaInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const FPGA_CO_LIST_SIZE =  5

// static void  Fpga_MappingDfs_rec( Fpga_Node_t * pNode, Fpga_NodeVec_t * vNodes, int fCollectEquiv );

// static void  Fpga_MappingDfsCuts_rec( Fpga_Node_t * pNode, Fpga_NodeVec_t * vNodes );

// static int   Fpga_MappingCompareOutputDelay( Fpga_Node_t ** ppNode1, Fpga_Node_t ** ppNode2 );

// static void  Fpga_MappingFindLatest( Fpga_Man_t * p, int * pNodes, int nNodesMax );

// static void  Fpga_DfsLim_rec( Fpga_Node_t * pNode, int Level, Fpga_NodeVec_t * vNodes );

// static int   Fpga_CollectNodeTfo_rec( Fpga_Node_t * pNode, Fpga_Node_t * pPivot, Fpga_NodeVec_t * vVisited, Fpga_NodeVec_t * vTfo );

// static Fpga_NodeVec_t * Fpga_MappingOrderCosByLevel( Fpga_Man_t * pMan );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Computes the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingDfs(pMan *Fpga_Man_t, fCollectEquiv int32) *Fpga_NodeVec_t {
    var vNodes *Fpga_NodeVec_t;//, * vNodesCo;
    var pNode *Fpga_Node_t;
    var i int32;
    // collect the CO nodes by level
//    vNodesCo = Fpga_MappingOrderCosByLevel( pMan );
    // start the array
    vNodes = Fpga_NodeVecAlloc( 100 );
    // collect the PIs
    for  i = 0; i < pMan.nInputs; i++  {
        pNode = pMan.pInputs[i];
        Fpga_NodeVecPush( vNodes, pNode );
        pNode.fMark0 = 1;
    }
    // perform the traversal
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro Fpga_Regular
        Fpga_MappingDfs_rec( ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) , vNodes, fCollectEquiv );
    }
//    for ( i = vNodesCo->nSize - 1; i >= 0 ; i-- )
//        for ( pNode = vNodesCo->pArray[i]; pNode; pNode = (Fpga_Node_t *)pNode->pData0 )
//            Fpga_MappingDfs_rec( pNode, vNodes, fCollectEquiv );
    // clean the node marks
    for  i = 0; i < vNodes.nSize; i++  {
        vNodes.pArray[i].fMark0 = 0;
    }
//    for ( i = 0; i < pMan->nOutputs; i++ )
//        Fpga_MappingUnmark_rec( Fpga_Regular(pMan->pOutputs[i]) );
//    Fpga_NodeVecFree( vNodesCo );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Recursively computes the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingDfs_rec(pNode *Fpga_Node_t, vNodes *Fpga_NodeVec_t, fCollectEquiv int32) {
    // inlining of macro Fpga_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fpga_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    if pNode.fMark0 {
        return; }
    // visit the transitive fanin
    if Fpga_NodeIsAnd(pNode) {
        // inlining of macro Fpga_Regular
        Fpga_MappingDfs_rec( ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) , vNodes, fCollectEquiv );
        // inlining of macro Fpga_Regular
        Fpga_MappingDfs_rec( ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) , vNodes, fCollectEquiv );
    }
    // visit the equivalent nodes
    if (fCollectEquiv != 0) && pNode.pNextE {
        Fpga_MappingDfs_rec( pNode.pNextE, vNodes, fCollectEquiv ); }
    // make sure the node is not visited through the equivalent nodes
    assert(pNode.fMark0 == 0, "pNode.fMark0 == 0");
    // mark the node as visited
    pNode.fMark0 = 1;
    // add the node to the list
    Fpga_NodeVecPush( vNodes, pNode );
}

/**Function*************************************************************

  Synopsis    [Computes the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingDfsNodes(pMan *Fpga_Man_t, ppNodes **Fpga_Node_t, nNodes int32, fEquiv int32) *Fpga_NodeVec_t {
    var vNodes *Fpga_NodeVec_t;
    var i int32;
    // perform the traversal
    vNodes = Fpga_NodeVecAlloc( 200 );
    for  i = 0; i < nNodes; i++  {
        Fpga_MappingDfs_rec( ppNodes[i], vNodes, fEquiv );
    }
    for  i = 0; i < vNodes.nSize; i++  {
        vNodes.pArray[i].fMark0 = 0;
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingGetAreaFlow(p *Fpga_Man_t) float32 {
    var aFlowFlowTotal float32;
    var i int32;
    for  i = 0; i < p.nOutputs; i++  {
        if Fpga_NodeIsConst(p.pOutputs[i]) {
            continue; }
        // inlining of macro Fpga_Regular
        aFlowFlowTotal += ((Fpga_Node_t *)((ABC_PTRUINT_T)(p.pOutputs[i]) & ~01)) .pCutBest.aFlow;
    }
    return aFlowFlowTotal;
}

/**Function*************************************************************

  Synopsis    [Computes the area of the current mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingArea(pMan *Fpga_Man_t) float32 {
    var pNode *Fpga_Node_t;
    var aTotal float32;
    var i int32;
    // perform the traversal
    aTotal = 0;
    for  i = 0; i < pMan.vMapping.nSize; i++  {
        pNode = pMan.vMapping.pArray[i];
        aTotal += pMan.pLutLib.pLutAreas[int32(pNode.pCutBest.nLeaves)];
    }
    return aTotal;
}

/**Function*************************************************************

  Synopsis    [Recursively computes the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingArea_rec(pMan *Fpga_Man_t, pNode *Fpga_Node_t, vNodes *Fpga_NodeVec_t) float32 {
    var aArea float32;
    var i int32;
    // inlining of macro Fpga_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fpga_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    if !Fpga_NodeIsAnd(pNode) {
        return 0; }
    if pNode.fMark0 {
        return 0; }
    assert(pNode.pCutBest != nil, "pNode.pCutBest != nil");
    // visit the transitive fanin of the selected cut
    aArea = 0;
    for  i = 0; i < pNode.pCutBest.nLeaves; i++  {
        aArea += Fpga_MappingArea_rec( pMan, pNode.pCutBest.ppLeaves[i], vNodes );
    }
    // make sure the node is not visited through the fanin nodes
    assert(pNode.fMark0 == 0, "pNode.fMark0 == 0");
    // mark the node as visited
    pNode.fMark0 = 1;
    // add the node to the list
    aArea += pMan.pLutLib.pLutAreas[int32(pNode.pCutBest.nLeaves)];
    // add the node to the list
    Fpga_NodeVecPush( vNodes, pNode );
    return aArea;
}

/**Function*************************************************************

  Synopsis    [Computes the area of the current mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingAreaTrav(pMan *Fpga_Man_t) float32 {
    var vNodes *Fpga_NodeVec_t;
    var aTotal float32;
    var i int32;
    // perform the traversal
    aTotal = 0;
    vNodes = Fpga_NodeVecAlloc( 100 );
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro Fpga_Regular
        aTotal += Fpga_MappingArea_rec( pMan, ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) , vNodes );
    }
    for  i = 0; i < vNodes.nSize; i++  {
        vNodes.pArray[i].fMark0 = 0;
    }
    Fpga_NodeVecFree( vNodes );
    return aTotal;
}


/**Function*************************************************************

  Synopsis    [Recursively computes the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingSetRefsAndArea_rec(pMan *Fpga_Man_t, pNode *Fpga_Node_t, ppStore **Fpga_Node_t) float32 {
    var aArea float32;
    var i int32;
    // inlining of macro Fpga_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fpga_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    if pNode.nRefs++ {
        return 0; }
    if !Fpga_NodeIsAnd(pNode) {
        return 0; }
    assert(pNode.pCutBest != nil, "pNode.pCutBest != nil");
    // store the node in the structure by level
    pNode.pData0 = (char *)ppStore[pNode.Level]; 
    ppStore[pNode.Level] = pNode;
    // visit the transitive fanin of the selected cut
    aArea = pMan.pLutLib.pLutAreas[int32(pNode.pCutBest.nLeaves)];
    for  i = 0; i < pNode.pCutBest.nLeaves; i++  {
        aArea += Fpga_MappingSetRefsAndArea_rec( pMan, pNode.pCutBest.ppLeaves[i], ppStore );
    }
    return aArea;
}

/**Function*************************************************************

  Synopsis    [Sets the correct reference counts for the mapping.]

  Description [Collects the nodes in reverse topological order
  and places in them in array pMan->vMapping.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingSetRefsAndArea(pMan *Fpga_Man_t) float32 {
    var pNode *Fpga_Node_t;
    var ppStore **Fpga_Node_t;
    var aArea float32;
    var i, LevelMax int32;

    // clean all references
    for  i = 0; i < pMan.vNodesAll.nSize; i++  {
        pMan.vNodesAll.pArray[i].nRefs = 0;
    }

    // allocate place to store the nodes
    LevelMax = Fpga_MappingMaxLevel( pMan );
    // inlining of macro ABC_ALLOC
    ppStore = ((Fpga_Node_t * *) malloc(sizeof(Fpga_Node_t *) * (size_t)(LevelMax + 1)));
    memset( ppStore, 0, sizeof(Fpga_Node_t *) * (LevelMax + 1) );

    // collect nodes reachable from POs in the DFS order through the best cuts
    aArea = 0;
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro Fpga_Regular
        pNode = ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) ;
        if pNode == pMan.pConst1 {
            continue; }
        aArea += Fpga_MappingSetRefsAndArea_rec( pMan, pNode, ppStore );
        pNode.nRefs++;
    }

    // reconnect the nodes in reverse topological order
    pMan.vMapping.nSize = 0;
    for  i = LevelMax; i >= 0; i--  {
        for  pNode = ppStore[i]; (pNode != nil); pNode = (Fpga_Node_t *)pNode.pData0  {
            Fpga_NodeVecPush( pMan.vMapping, pNode );
        }
    }
    ppStore = nil;
    return aArea;
}


/**Function*************************************************************

  Synopsis    [Compares the outputs by their arrival times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingCompareOutputDelay(ppNode1 **Fpga_Node_t, ppNode2 **Fpga_Node_t) int32 {
    // inlining of macro Fpga_Regular
    var pNode1 *Fpga_Node_t  = ((Fpga_Node_t *)((ABC_PTRUINT_T)(*ppNode1) & ~01)) ;
    // inlining of macro Fpga_Regular
    var pNode2 *Fpga_Node_t  = ((Fpga_Node_t *)((ABC_PTRUINT_T)(*ppNode2) & ~01)) ;
    var Arrival1 float32;
    if pNode1.pCutBest {
        Arrival1 = pNode1.pCutBest.tArrival;
    } else {
        Arrival1 = 0;
    }
    var Arrival2 float32;
    if pNode2.pCutBest {
        Arrival2 = pNode2.pCutBest.tArrival;
    } else {
        Arrival2 = 0;
    }
    if Arrival1 < Arrival2 {
        return -1; }
    if Arrival1 > Arrival2 {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Finds given number of latest arriving COs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingFindLatest(p *Fpga_Man_t, pNodes *int32, nNodesMax int32) {
    var nNodes, i, k, v int32;
    assert(p.nOutputs >= nNodesMax, "p.nOutputs >= nNodesMax");
    pNodes[0] = 0;
    nNodes = 1;
    for  i = 1; i < p.nOutputs; i++  {
        for  k = nNodes - 1; k >= 0; k--  {
            if Fpga_MappingCompareOutputDelay( &p.pOutputs[pNodes[k]], &p.pOutputs[i] ) >= 0 {
                break; }
        }
        if k == nNodesMax - 1 {
            continue; }
        if nNodes < nNodesMax {
            nNodes++; }
        for  v = nNodes - 1; v > k+1; v--  {
            pNodes[v] = pNodes[v-1];
        }
        pNodes[k+1] = i;
    }
}

/**Function*************************************************************

  Synopsis    [Prints a bunch of latest arriving outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingPrintOutputArrivals(p *Fpga_Man_t) {
    var pNode *Fpga_Node_t;
    var pSorted [FPGA_CO_LIST_SIZE]int32;
    var fCompl, Limit, MaxNameSize, i int32;

    // determine the number of nodes to print
    if p.nOutputs > FPGA_CO_LIST_SIZE {
        Limit = FPGA_CO_LIST_SIZE;
    } else {
        Limit = p.nOutputs;
    }

    // determine the order
    Fpga_MappingFindLatest( p, pSorted, Limit );

    // determine max size of the node's name
    MaxNameSize = 0;
    for  i = 0; i < Limit; i++  {
        if MaxNameSize < int32(len(p.ppOutputNames[pSorted[i]])) {
            MaxNameSize = len(p.ppOutputNames[pSorted[i]]); }
    }

    // print the latest outputs
    for  i = 0; i < Limit; i++  {
        // get the i-th latest output
        // inlining of macro Fpga_Regular
        pNode  = ((Fpga_Node_t *)((ABC_PTRUINT_T)(p.pOutputs[pSorted[i]]) & ~01)) ;
        // inlining of macro Fpga_IsComplement
        fCompl = ((int32(((ABC_PTRUINT_T) (p.pOutputs[pSorted[i]]) & 01))));
        // print out the best arrival time
        fmt.Printf("Output  %-*s : ", MaxNameSize + 3, p.ppOutputNames[pSorted[i]]);
        fmt.Printf("Delay = %8.2f  ", float64(pNode.pCutBest.tArrival));
        if fCompl {
            fmt.Printf("NEG");
        } else {
            fmt.Printf("POS"); }
        fmt.Printf("\n");
    }
}


/**Function*************************************************************

  Synopsis    [Sets up the truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingSetupTruthTables(uTruths [][2]uint32) {
    var m, v int32;
    // set up the truth tables
    for  m = 0; m < 32; m++  {
        for  v = 0; v < 5; v++  {
            if m & (1 << v) {
                uTruths[v][0] |= (1 << m); }
        }
    }
    // make adjustments for the case of 6 variables
    for  v = 0; v < 5; v++  {
        uTruths[v][1] = uTruths[v][0];
    }
    uTruths[5][0] = 0;
    uTruths[5][1] = FPGA_FULL;
}

/**Function*************************************************************

  Synopsis    [Sets up the mask.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingSetupMask(uMask []uint32, nVarsMax int32) {
    if nVarsMax == 6 {
        uMask[0] = uMask[1] = FPGA_FULL;
    } else {
        // inlining of macro FPGA_MASK
        uMask[0] = ((~(uint32(0))) >> (32-(1 << nVarsMax)));
        uMask[1] = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Verify one useful property.]

  Description [This procedure verifies one useful property. After 
  the FRAIG construction with choice nodes is over, each primary node 
  should have fanins that are primary nodes. The primary nodes is the 
  one that does not have pNode->pRepr set to point to another node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_ManCheckConsistency(p *Fpga_Man_t) int32 {
    var pNode *Fpga_Node_t;
    var pVec *Fpga_NodeVec_t;
    var i int32;
    pVec = Fpga_MappingDfs( p, 0 );
    for  i = 0; i < pVec.nSize; i++  {
        pNode = pVec.pArray[i];
        if Fpga_NodeIsVar(pNode) {
            if pNode.pRepr {
                fmt.Printf("Primary input %d is a secondary node.\n", pNode.Num); }
        
        } else if Fpga_NodeIsConst(pNode) {
            if pNode.pRepr {
                fmt.Printf("Constant 1 %d is a secondary node.\n", pNode.Num); }
        
        } else {
            if pNode.pRepr {
                fmt.Printf("Internal node %d is a secondary node.\n", pNode.Num); }
            // inlining of macro Fpga_Regular
            if  ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .pRepr {
                fmt.Printf("Internal node %d has first fanin that is a secondary node.\n", pNode.Num); }
            // inlining of macro Fpga_Regular
            if  ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .pRepr {
                fmt.Printf("Internal node %d has second fanin that is a secondary node.\n", pNode.Num); }
        }
    }
    Fpga_NodeVecFree( pVec );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Compares the supergates by their level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CompareNodesByLevelDecreasing(ppS1 **Fpga_Node_t, ppS2 **Fpga_Node_t) int32 {
    // inlining of macro Fpga_Regular
    if  ((Fpga_Node_t *)((ABC_PTRUINT_T)(*ppS1) & ~01)) .Level > ((Fpga_Node_t *)((ABC_PTRUINT_T)(*ppS2) & ~01)) .Level {
        return -1; }
    // inlining of macro Fpga_Regular
    if  ((Fpga_Node_t *)((ABC_PTRUINT_T)(*ppS1) & ~01)) .Level < ((Fpga_Node_t *)((ABC_PTRUINT_T)(*ppS2) & ~01)) .Level {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Compares the supergates by their level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CompareNodesByLevelIncreasing(ppS1 **Fpga_Node_t, ppS2 **Fpga_Node_t) int32 {
    // inlining of macro Fpga_Regular
    if  ((Fpga_Node_t *)((ABC_PTRUINT_T)(*ppS1) & ~01)) .Level < ((Fpga_Node_t *)((ABC_PTRUINT_T)(*ppS2) & ~01)) .Level {
        return -1; }
    // inlining of macro Fpga_Regular
    if  ((Fpga_Node_t *)((ABC_PTRUINT_T)(*ppS1) & ~01)) .Level > ((Fpga_Node_t *)((ABC_PTRUINT_T)(*ppS2) & ~01)) .Level {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Orders the nodes in the decreasing order of levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingSortByLevel(pMan *Fpga_Man_t, vNodes *Fpga_NodeVec_t, fIncreasing int32) {
    if fIncreasing != 0 {
        qsort( (void *)vNodes.pArray, (size_t)vNodes.nSize, sizeof(Fpga_Node_t *), 
                        (int (*)(void *, void *)) Fpga_CompareNodesByLevelIncreasing );
    } else {
        qsort( (void *)vNodes.pArray, (size_t)vNodes.nSize, sizeof(Fpga_Node_t *), 
                        (int (*)(void *, void *)) Fpga_CompareNodesByLevelDecreasing ); }
//    assert( Fpga_CompareNodesByLevel( vNodes->pArray, vNodes->pArray + vNodes->nSize - 1 ) <= 0 );
}

/**Function*************************************************************

  Synopsis    [Computes the limited DFS ordering for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_DfsLim(pMan *Fpga_Man_t, pNode *Fpga_Node_t, nLevels int32) *Fpga_NodeVec_t {
    var vNodes *Fpga_NodeVec_t;
    var i int32;
    // perform the traversal
    vNodes = Fpga_NodeVecAlloc( 100 );
    Fpga_DfsLim_rec( pNode, nLevels, vNodes );
    for  i = 0; i < vNodes.nSize; i++  {
        vNodes.pArray[i].fMark0 = 0;
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Recursively computes the DFS ordering of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_DfsLim_rec(pNode *Fpga_Node_t, Level int32, vNodes *Fpga_NodeVec_t) {
    // inlining of macro Fpga_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fpga_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    if pNode.fMark0 {
        return; }
    pNode.fMark0 = 1;
    // visit the transitive fanin
    Level--;
    if Level > 0 && Fpga_NodeIsAnd(pNode) {
        // inlining of macro Fpga_Regular
        Fpga_DfsLim_rec( ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) , Level, vNodes );
        // inlining of macro Fpga_Regular
        Fpga_DfsLim_rec( ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) , Level, vNodes );
    }
    // add the node to the list
    Fpga_NodeVecPush( vNodes, pNode );
}

/**Function*************************************************************

  Synopsis    [Computes the limited DFS ordering for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_ManCleanData0(pMan *Fpga_Man_t) {
    var i int32;
    for  i = 0; i < pMan.vNodesAll.nSize; i++  {
        pMan.vNodesAll.pArray[i].pData0 = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Collects the TFO of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CollectNodeTfo(pMan *Fpga_Man_t, pNode *Fpga_Node_t) *Fpga_NodeVec_t {
    var vVisited, vTfo *Fpga_NodeVec_t;
    var i int32;
    // perform the traversal
    vVisited = Fpga_NodeVecAlloc( 100 );
    vTfo     = Fpga_NodeVecAlloc( 100 );
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro Fpga_Regular
        Fpga_CollectNodeTfo_rec( ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) , pNode, vVisited, vTfo );
    }
    for  i = 0; i < vVisited.nSize; i++  {
        vVisited.pArray[i].fMark0 = vVisited.pArray[i].fMark1 = 0;
    }
    Fpga_NodeVecFree( vVisited );
    return vTfo;
}

/**Function*************************************************************

  Synopsis    [Collects the TFO of the node.]

  Description [Returns 1 if the node should be collected.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CollectNodeTfo_rec(pNode *Fpga_Node_t, pPivot *Fpga_Node_t, vVisited *Fpga_NodeVec_t, vTfo *Fpga_NodeVec_t) int32 {
    var Ret1, Ret2 int32;
    // inlining of macro Fpga_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fpga_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    // skip visited nodes
    if pNode.fMark0 {
        return pNode.fMark1; }
    pNode.fMark0 = 1;
    Fpga_NodeVecPush( vVisited, pNode );

    // return the pivot node
    if pNode == pPivot {
        pNode.fMark1 = 1;
        return 1;
    }
    if pNode.Level < pPivot.Level {
        pNode.fMark1 = 0;
        return 0;
    }
    // visit the transitive fanin
    assert(Fpga_NodeIsAnd(pNode), "Fpga_NodeIsAnd(pNode)");
    // inlining of macro Fpga_Regular
    Ret1 = Fpga_CollectNodeTfo_rec( ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) , pPivot, vVisited, vTfo );
    // inlining of macro Fpga_Regular
    Ret2 = Fpga_CollectNodeTfo_rec( ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) , pPivot, vVisited, vTfo );
    if Ret1 || (Ret2 != 0) {
        pNode.fMark1 = 1;
        Fpga_NodeVecPush( vTfo, pNode );
    
    } else {
        pNode.fMark1 = 0; }
    return pNode.fMark1;
}

/**Function*************************************************************

  Synopsis    [Levelizes the nodes accessible from the POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingLevelize(pMan *Fpga_Man_t, vNodes *Fpga_NodeVec_t) *Fpga_NodeVec_t {
    var vLevels *Fpga_NodeVec_t;
    var ppNodes **Fpga_Node_t;
    var pNode *Fpga_Node_t;
    var nNodes, nLevelsMax, i int32;

    // reassign the levels (this may be necessary for networks which choices)
    ppNodes = vNodes.pArray;
    nNodes  = vNodes.nSize;
    for  i = 0; i < nNodes; i++  {
        pNode = ppNodes[i];
        if !Fpga_NodeIsAnd(pNode) {
            pNode.Level = 0;
            continue;
        }
        // inlining of macro FPGA_MAX
        // inlining of macro Fpga_Regular
        pNode.Level = 1 + (((((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .Level) > (((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .Level))? (((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .Level) : (((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .Level));
    }

    // get the max levels
    nLevelsMax = 0;
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro FPGA_MAX
        
        // inlining of macro Fpga_Regular
        if (nLevelsMax) > (int32(((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) .Level)) {
            nLevelsMax = (nLevelsMax);
        } else {
            // inlining of macro Fpga_Regular
            nLevelsMax = (int32(((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) .Level));
        }
    }
    nLevelsMax++;

    // allocate storage for levels
    vLevels = Fpga_NodeVecAlloc( nLevelsMax );
    for  i = 0; i < nLevelsMax; i++  {
        Fpga_NodeVecPush( vLevels, nil );
    }

    // go through the nodes and add them to the levels
    for  i = 0; i < nNodes; i++  {
        pNode = ppNodes[i];
        pNode.pLevel = nil;
        if !Fpga_NodeIsAnd(pNode) {
            continue; }
        // attach the node to this level
        pNode.pLevel = Fpga_NodeVecReadEntry( vLevels, pNode.Level );
        Fpga_NodeVecWriteEntry( vLevels, pNode.Level, pNode );
    }
    return vLevels;
}

/**Function*************************************************************

  Synopsis    [Sets up the mask.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingMaxLevel(pMan *Fpga_Man_t) int32 {
    var nLevelMax, i int32;
    nLevelMax = 0;
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro Fpga_Regular
        if nLevelMax > int32(((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) .Level) {
            nLevelMax = nLevelMax;
        } else {
            // inlining of macro Fpga_Regular
            nLevelMax = int32(((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) .Level);
        }
    }
    return nLevelMax;
}


/**Function*************************************************************

  Synopsis    [Analyses choice nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingUpdateLevel_rec(pMan *Fpga_Man_t, pNode *Fpga_Node_t, fMaximum int32) int32 {
    var pTemp *Fpga_Node_t;
    var Level1, Level2, LevelE int32;
    // inlining of macro Fpga_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fpga_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    if !Fpga_NodeIsAnd(pNode) {
        return pNode.Level; }
    // skip the visited node
    if pNode.TravId == pMan.nTravIds {
        return pNode.Level; }
    pNode.TravId = pMan.nTravIds;
    // compute levels of the children nodes
    // inlining of macro Fpga_Regular
    Level1 = Fpga_MappingUpdateLevel_rec( pMan, ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) , fMaximum );
    // inlining of macro Fpga_Regular
    Level2 = Fpga_MappingUpdateLevel_rec( pMan, ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) , fMaximum );
    // inlining of macro FPGA_MAX
    pNode.Level = 1 + (((Level1) > (Level2))? (Level1) : (Level2));
    if pNode.pNextE {
        LevelE = Fpga_MappingUpdateLevel_rec( pMan, pNode.pNextE, fMaximum );
        if fMaximum != 0 {
            if pNode.Level < uint32(LevelE) {
                pNode.Level = LevelE; }
        
        } else {
            if pNode.Level > uint32(LevelE) {
                pNode.Level = LevelE; }
        }
        // set the level of all equivalent nodes to be the same minimum
        if pNode.pRepr == nil { // the primary node
            for  pTemp = pNode.pNextE; (pTemp != nil); pTemp = pTemp.pNextE  {
                pTemp.Level = pNode.Level;
            } }
    }
    return pNode.Level;
}

/**Function*************************************************************

  Synopsis    [Resets the levels of the nodes in the choice graph.]

  Description [Makes the level of the choice nodes to be equal to the
  maximum of the level of the nodes in the equivalence class. This way
  sorting by level leads to the reverse topological order, which is
  needed for the required time computation.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingSetChoiceLevels(pMan *Fpga_Man_t) {
    var i int32;
    pMan.nTravIds++;
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro Fpga_Regular
        Fpga_MappingUpdateLevel_rec( pMan, ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) , 1 );
    }
}

/**Function*************************************************************

  Synopsis    [Reports statistics on choice nodes.]

  Description [The number of choice nodes is the number of primary nodes,
  which has pNextE set to a pointer. The number of choices is the number
  of entries in the equivalent-node lists of the primary nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_ManReportChoices(pMan *Fpga_Man_t) {
    var pNode, pTemp *Fpga_Node_t;
    var nChoiceNodes, nChoices int32;
    var i, LevelMax1, LevelMax2 int32;

    // report the number of levels
    LevelMax1 = Fpga_MappingMaxLevel( pMan );
    pMan.nTravIds++;
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro Fpga_Regular
        Fpga_MappingUpdateLevel_rec( pMan, ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) , 0 );
    }
    LevelMax2 = Fpga_MappingMaxLevel( pMan );

    // report statistics about choices
    nChoiceNodes = nChoices = 0;
    for  i = 0; i < pMan.vAnds.nSize; i++  {
        pNode = pMan.vAnds.pArray[i];
        if pNode.pRepr == nil && pNode.pNextE != nil { // this is a choice node = the primary node that has equivalent nodes
            nChoiceNodes++;
            for  pTemp = pNode; (pTemp != nil); pTemp = pTemp.pNextE  {
                nChoices++;
            }
        }
    }
    if pMan.fVerbose {
    fmt.Printf("Maximum level: Original = %d. Reduced due to choices = %d.\n", LevelMax1, LevelMax2);
    fmt.Printf("Choice stats:  Choice nodes = %d. Total choices = %d.\n", nChoiceNodes, nChoices);
    }
/*
    {
        FILE * pTable;
        pTable = fopen( "stats_choice.txt", "a+" );
        fprintf( pTable, "%s ", pMan->pFileName );
        fprintf( pTable, "%4d ", LevelMax1 );
        fprintf( pTable, "%4d ", pMan->vAnds->nSize - pMan->nInputs );
        fprintf( pTable, "%4d ", LevelMax2 );
        fprintf( pTable, "%7d ", nChoiceNodes );
        fprintf( pTable, "%7d ", nChoices + nChoiceNodes );
        fprintf( pTable, "\n" );
        fclose( pTable );
    }
*/
}

/**Function*************************************************************

  Synopsis    [Returns the array of CO nodes sorted by level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingOrderCosByLevel(pMan *Fpga_Man_t) *Fpga_NodeVec_t {
    var pNode *Fpga_Node_t;
    var vNodes *Fpga_NodeVec_t;
    var i, nLevels int32;
    // get the largest level of a CO
    nLevels = Fpga_MappingMaxLevel( pMan );
    // allocate the array of nodes
    vNodes = Fpga_NodeVecAlloc( nLevels + 1 );
    for  i = 0; i <= nLevels; i++  {
        Fpga_NodeVecPush( vNodes, nil );
    }
    // clean the marks
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro Fpga_Regular
        ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) .fMark0 = 0;
    }
    // put the nodes into the structure
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro Fpga_Regular
        pNode = ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) ;
        if pNode.fMark0 {
            continue; }
        pNode.fMark0 = 1;
        pNode.pData0 = (char *)Fpga_NodeVecReadEntry( vNodes, pNode.Level );
        Fpga_NodeVecWriteEntry( vNodes, pNode.Level, pNode );
    }
    for  i = 0; i < pMan.nOutputs; i++  {
        // inlining of macro Fpga_Regular
        ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pOutputs[i]) & ~01)) .fMark0 = 0;
    }
    return vNodes;

}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


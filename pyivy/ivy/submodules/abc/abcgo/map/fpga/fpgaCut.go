package fpga

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [fpgaCut.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - August 18, 2004.]

  Revision    [$Id: fpgaCut.c,v 1.3 2004/07/06 04:55:57 alanmi Exp $]

***********************************************************************/
 
//#include "fpgaInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Fpga_CutTable_t struct {
    pBins  **Fpga_Cut_t;        // the table used for linear probing
    nBins           int32;        // the size of the table
    pCuts          *int32;        // the array of cuts currently stored 
    nCuts           int32;        // the number of cuts currently stored 
    pArray  **Fpga_Cut_t;       // the temporary array of cuts
    pCuts1  **Fpga_Cut_t;       // the temporary array of cuts
    pCuts2  **Fpga_Cut_t;       // the temporary array of cuts
}

// the largest number of cuts considered
//#define  FPGA_CUTS_MAX_COMPUTE   500
const FPGA_CUTS_MAX_COMPUTE =   2000
// the largest number of cuts used
//#define  FPGA_CUTS_MAX_USE       200
const FPGA_CUTS_MAX_USE =       1000

// primes used to compute the hash key
var s_HashPrimes [10]int32 = [10]int32{ 109, 499, 557, 619, 631, 709, 797, 881, 907, 991 };

var bit_count [256]int32 = [256]int32{
  0,1,1,2,1,2,2,3,1,2,2,3,2,3,3,4,1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  1,2,2,3,2,3,3,4,2,3,3,4,3,4,4,5,2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  2,3,3,4,3,4,4,5,3,4,4,5,4,5,5,6,3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,
  3,4,4,5,4,5,5,6,4,5,5,6,5,6,6,7,4,5,5,6,5,6,6,7,5,6,6,7,6,7,7,8,
};

// Original macro:
// #define FPGA_COUNT_ONES(u) (bit_count[(u)&255]+bit_count[((u)>>8)&255]+bit_count[((u)>>16)&255]+bit_count[(u)>>24])


// static Fpga_Cut_t *      Fpga_CutCompute( Fpga_Man_t * p, Fpga_CutTable_t * pTable, Fpga_Node_t * pNode );

// static void              Fpga_CutFilter( Fpga_Man_t * p, Fpga_Node_t * pNode );

// static Fpga_Cut_t *      Fpga_CutMergeLists( Fpga_Man_t * p, Fpga_CutTable_t * pTable, Fpga_Cut_t * pList1, Fpga_Cut_t * pList2, int fComp1, int fComp2, int fPivot1, int fPivot2 );

// static int               Fpga_CutMergeTwo( Fpga_Cut_t * pCut1, Fpga_Cut_t * pCut2, Fpga_Node_t * ppNodes[], int nNodesMax );

// static Fpga_Cut_t *      Fpga_CutUnionLists( Fpga_Cut_t * pList1, Fpga_Cut_t * pList2 );

// static int               Fpga_CutBelongsToList( Fpga_Cut_t * pList, Fpga_Node_t * ppNodes[], int nNodes );

// extern Fpga_Cut_t *      Fpga_CutAlloc( Fpga_Man_t * p );

// extern int               Fpga_CutCountAll( Fpga_Man_t * pMan );


// static void              Fpga_CutListPrint( Fpga_Man_t * pMan, Fpga_Node_t * pRoot );

// static void              Fpga_CutListPrint2( Fpga_Man_t * pMan, Fpga_Node_t * pRoot );

// static void              Fpga_CutPrint_( Fpga_Man_t * pMan, Fpga_Cut_t * pCut, Fpga_Node_t * pRoot );


// static Fpga_CutTable_t * Fpga_CutTableStart( Fpga_Man_t * pMan );

// static void              Fpga_CutTableStop( Fpga_CutTable_t * p );

// static unsigned          Fpga_CutTableHash( Fpga_Node_t * ppNodes[], int nNodes );

// static int               Fpga_CutTableLookup( Fpga_CutTable_t * p, Fpga_Node_t * ppNodes[], int nNodes );

// static Fpga_Cut_t *      Fpga_CutTableConsider( Fpga_Man_t * pMan, Fpga_CutTable_t * p, Fpga_Node_t * ppNodes[], int nNodes );

// static void              Fpga_CutTableRestart( Fpga_CutTable_t * p );


// static int               Fpga_CutSortCutsCompare( Fpga_Cut_t ** pC1, Fpga_Cut_t ** pC2 );

// static Fpga_Cut_t *      Fpga_CutSortCuts( Fpga_Man_t * pMan, Fpga_CutTable_t * p, Fpga_Cut_t * pList );

// static int               Fpga_CutList2Array( Fpga_Cut_t ** pArray, Fpga_Cut_t * pList );

// static Fpga_Cut_t *      Fpga_CutArray2List( Fpga_Cut_t ** pArray, int nCuts );



// iterator through all the cuts of the list
// Original macro:
// #define Fpga_ListForEachCut( pList, pCut )                \
//     for ( pCut = pList;                                   \
//           pCut;                                           \
//           pCut = pCut->pNext )

// Original macro:
// #define Fpga_ListForEachCutSafe( pList, pCut, pCut2 )     \
//     for ( pCut = pList,                                   \
//           pCut2 = pCut? pCut->pNext: NULL;                \
//           pCut;                                           \
//           pCut = pCut2,                                   \
//           pCut2 = pCut? pCut->pNext: NULL )



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes the cuts for each node in the object graph.]

  Description [The cuts are computed in one sweep over the mapping graph. 
  First, the elementary cuts, which include the node itself, are assigned 
  to the PI nodes. The internal nodes are considered in the DFS order.
  Each node is two-input AND-gate. So to compute the cuts at a node, we
  need to merge the sets of cuts of its two predecessors. The merged set
  contains only unique cuts with the number of inputs equal to k or less.
  Finally, the elementary cut, composed of the node itself, is added to
  the set of cuts for the node.
  
  This procedure is pretty fast for 5-feasible cuts, but it dramatically
  slows down on some "dense" networks when computing 6-feasible cuts.
  The problem is that there are too many cuts in this case. We should
  think how to heuristically trim the number of cuts in such cases, 
  to have reasonable runtime.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingCuts(p *Fpga_Man_t) {
    var pProgress *ProgressBar;
    var pTable *Fpga_CutTable_t;
    var pNode *Fpga_Node_t;
    var nCuts, nNodes, i int32;
    var clk int64  = time.Now().UnixNano();

    // set the elementary cuts for the PI variables
    assert(p.nVarsMax > 1 && p.nVarsMax < 11, "p.nVarsMax > 1 && p.nVarsMax < 11");
    Fpga_MappingCreatePiCuts( p );

    // compute the cuts for the internal nodes
    nNodes = p.vAnds.nSize;
    pProgress = Extra_ProgressBarStart( os.Stdout, nNodes );
    pTable = Fpga_CutTableStart( p );
    for  i = 0; i < nNodes; i++  {
        Extra_ProgressBarUpdate( pProgress, i, "Cuts ..." );
        pNode = p.vAnds.pArray[i];
        if !Fpga_NodeIsAnd( pNode ) {
            continue; }
        Fpga_CutCompute( p, pTable, pNode );
    }
    Extra_ProgressBarStop( pProgress );
    Fpga_CutTableStop( pTable );

    // report the stats
    if p.fVerbose {
        nCuts = Fpga_CutCountAll(p);
        fmt.Printf("Nodes = %6d. Total %d-cuts = %d. Cuts per node = %.1f. ", p.nNodes, p.nVarsMax, nCuts, (float32(nCuts))/p.nNodes);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
    }

    // print the cuts for the first primary output
//    Fpga_CutListPrint( p, Fpga_Regular(p->pOutputs[0]) );
}

/**Function*************************************************************

  Synopsis    [Performs technology mapping for variable-size-LUTs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingCreatePiCuts(p *Fpga_Man_t) {
    var pCut *Fpga_Cut_t;
    var i int32;

    // set the elementary cuts for the PI variables
    for  i = 0; i < p.nInputs; i++  {
        pCut = Fpga_CutAlloc( p );
        pCut.nLeaves = 1;
        pCut.ppLeaves[0] = p.pInputs[i];
        pCut.uSign = (1 << (i%31));
        p.pInputs[i].pCuts   = pCut;
        p.pInputs[i].pCutBest = pCut;
        // set the input arrival times
//        p->pInputs[i]->pCut[1]->tArrival = p->pInputArrivals[i];
    }
}

/**Function*************************************************************

  Synopsis    [Computes the cuts for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutCompute(p *Fpga_Man_t, pTable *Fpga_CutTable_t, pNode *Fpga_Node_t) *Fpga_Cut_t {
    var pTemp *Fpga_Node_t;
    var pList, pList1, pList2 *Fpga_Cut_t;
    var pCut *Fpga_Cut_t;
    var fTree int32;
    // inlining of macro Fpga_NodeReadRef
    // inlining of macro Fpga_Regular
    var fPivot1 int32  = fTree && (((((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) ).nRefs)>2);
    // inlining of macro Fpga_NodeReadRef
    // inlining of macro Fpga_Regular
    var fPivot2 int32  = fTree && (((((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) ).nRefs)>2);

    // if the cuts are computed return them
    if pNode.pCuts {
        return pNode.pCuts; }

    // compute the cuts for the children
    // inlining of macro Fpga_Regular
    pList1 = ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .pCuts;
    // inlining of macro Fpga_Regular
    pList2 = ((Fpga_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .pCuts;
    // merge the lists
    pList = Fpga_CutMergeLists( p, pTable, pList1, pList2, 
        // inlining of macro Fpga_IsComplement
        ((int32(((ABC_PTRUINT_T) (pNode.p1) & 01)))), ((int32(((ABC_PTRUINT_T) (pNode.p2) & 01)))),
        fPivot1, fPivot2 );
    // if there are functionally equivalent nodes, union them with this list
    assert((pList != nil), "pList");
    // only add to the list of cuts if the node is a representative one
    if pNode.pRepr == nil {
        for  pTemp = pNode.pNextE; pTemp; pTemp = pTemp.pNextE  {
            assert(pTemp.pCuts, "pTemp.pCuts");
            pList = Fpga_CutUnionLists( pList, pTemp.pCuts );
            assert(pTemp.pCuts, "pTemp.pCuts");
            pList = Fpga_CutSortCuts( p, pTable, pList );
        }
    }
    // add the new cut
    pCut = Fpga_CutAlloc( p );
    pCut.nLeaves = 1;
    pCut.ppLeaves[0] = pNode;
    pCut.uSign = (1 << (pNode.Num%31));
    pCut.fLevel = float32(pCut.ppLeaves[0].Level);
    // append (it is important that the elementary cut is appended first)
    pCut.pNext = pList;
    // set at the node
    pNode.pCuts = pCut;
    // remove the dominated cuts
//    Fpga_CutFilter( p, pNode );
    // set the phase correctly
    if pNode.pRepr && Fpga_NodeComparePhase(pNode, pNode.pRepr) {
        // inlining of macro Fpga_ListForEachCut
        for  pCut = pNode.pCuts;                                             (pCut != nil);                                                     pCut = pCut.pNext  {
            pCut.Phase = 1;
        }
    }


/*
    { 
        Fpga_Cut_t * pPrev;
        int i, Counter = 0;
        for ( pCut = pNode->pCuts->pNext, pPrev = pNode->pCuts; pCut; pCut = pCut->pNext )
        {
            for ( i = 0; i < pCut->nLeaves; i++ )
                if ( pCut->ppLeaves[i]->Level >= pNode->Level )
                    break;
            if ( i != pCut->nLeaves )
                pPrev->pNext = pCut->pNext;
            else 
                pPrev = pCut; 
        }
    }
    { 
        int i, Counter = 0;;
        for ( pCut = pNode->pCuts->pNext; pCut; pCut = pCut->pNext )
            for ( i = 0; i < pCut->nLeaves; i++ )
                Counter += (pCut->ppLeaves[i]->Level >= pNode->Level);
        if ( Counter )
            printf( " %d", Counter );
    }
*/

    return pCut;
}

/**Function*************************************************************

  Synopsis    [Filter the cuts using dominance.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutFilter(p *Fpga_Man_t, pNode *Fpga_Node_t) { 
    var pTemp, pPrev, pCut, pCut2 *Fpga_Cut_t;
    var i, k, Counter int32;

    Counter = 0;
    pPrev = pNode.pCuts;
    // inlining of macro Fpga_ListForEachCutSafe
    for  pCut = pNode.pCuts.pNext,                                             pCut2 = pCut? pCut.pNext: nil;                          (pCut != nil);                                                     pCut = pCut2,                                             pCut2 = pCut? pCut.pNext: nil  {
        // go through all the previous cuts up to pCut
        for  pTemp = pNode.pCuts.pNext; pTemp != pCut; pTemp = pTemp.pNext  {
            // check if every node in pTemp is contained in pCut
            for  i = 0; i < pTemp.nLeaves; i++  {
                for  k = 0; k < pCut.nLeaves; k++  {
                    if pTemp.ppLeaves[i] == pCut.ppLeaves[k] {
                        break; }
                }
                if k == pCut.nLeaves { // node i in pTemp is not contained in pCut
                    break; }
            }
            if i == pTemp.nLeaves { // every node in pTemp is contained in pCut
            
                Counter++;
                break;
            }
        }
        if pTemp != pCut { // pTemp contain pCut
        
            pPrev.pNext = pCut.pNext;  // skip pCut
            // recycle pCut
            Fpga_CutFree( p, pCut );
        
        } else { 
            pPrev = pCut; } 
    }
//  printf( "Dominated = %3d. \n", Counter );
}


/**Function*************************************************************

  Synopsis    [Merges two lists of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutMergeLists(p *Fpga_Man_t, pTable *Fpga_CutTable_t, pList1 *Fpga_Cut_t, pList2 *Fpga_Cut_t, fComp1 int32, fComp2 int32, fPivot1 int32, fPivot2 int32) *Fpga_Cut_t {
    var ppNodes [FPGA_MAX_LEAVES]*Fpga_Node_t;
    var pListNew *Fpga_Cut_t;
    var ppListNew **Fpga_Cut_t;
    var pLists [FPGA_MAX_LEAVES+1]*Fpga_Cut_t = [FPGA_MAX_LEAVES+1]*Fpga_Cut_t{ nil };
    var pCut, pPrev, pTemp1, pTemp2 *Fpga_Cut_t;
    var nNodes, Counter, i int32;
    var ppArray1, ppArray2, ppArray3 **Fpga_Cut_t;
    var nCuts1, nCuts2, nCuts3, k, fComp3 int32;

    ppArray1 = pTable.pCuts1;
    ppArray2 = pTable.pCuts2;
    nCuts1 = Fpga_CutList2Array( ppArray1, pList1 );
    nCuts2 = Fpga_CutList2Array( ppArray2, pList2 );
    if fPivot1 != 0 {
        nCuts1 = 1; }
    if fPivot2 != 0 {
        nCuts2 = 1; }
    // swap the lists based on their length
    if nCuts1 > nCuts2 {
         ppArray3 = ppArray1;
         ppArray1 = ppArray2;
         ppArray2 = ppArray3;

         nCuts3 = nCuts1;
         nCuts1 = nCuts2;
         nCuts2 = nCuts3;

         fComp3 = fComp1;
         fComp1 = fComp2;
         fComp2 = fComp3;
    }
    // pList1 is shorter or equal length compared to pList2
 
    // prepare the manager for the cut computation
    Fpga_CutTableRestart( pTable );
    // go through the cut pairs
    Counter = 0;
//    for ( pTemp1 = pList1; pTemp1; pTemp1 = fPivot1? NULL: pTemp1->pNext )
//        for ( pTemp2 = pList2; pTemp2; pTemp2 = fPivot2? NULL: pTemp2->pNext )
    for  i = 0; i < nCuts1; i++  {
        for  k = 0; k <= i; k++  {
            pTemp1 = ppArray1[i];
            pTemp2 = ppArray2[k];

            if pTemp1.nLeaves == p.nVarsMax && pTemp2.nLeaves == p.nVarsMax {
                if pTemp1.ppLeaves[0] != pTemp2.ppLeaves[0] {
                    continue; }
                if pTemp1.ppLeaves[1] != pTemp2.ppLeaves[1] {
                    continue; }
            }

            // check if k-feasible cut exists
            nNodes = Fpga_CutMergeTwo( pTemp1, pTemp2, ppNodes, p.nVarsMax );
            if nNodes == 0 {
                continue; }
            // consider the cut for possible addition to the set of new cuts
            pCut = Fpga_CutTableConsider( p, pTable, ppNodes, nNodes );
            if pCut == nil {
                continue; }
            // add data to the cut
            // inlining of macro Fpga_CutNotCond
            pCut.pOne = ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pTemp1) ^ (fComp1)));
            // inlining of macro Fpga_CutNotCond
            pCut.pTwo = ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pTemp2) ^ (fComp2)));
            // create the signature
            pCut.uSign = pTemp1.uSign | pTemp2.uSign;
            // add it to the corresponding list
            pCut.pNext = pLists[int32(pCut.nLeaves)];
            pLists[int32(pCut.nLeaves)] = pCut;
            // count this cut and quit if limit is reached
            Counter++;
            if Counter == FPGA_CUTS_MAX_COMPUTE {
                goto QUITS; }
        }
        for  k = 0; k < i; k++  {
            pTemp1 = ppArray1[k];
            pTemp2 = ppArray2[i];

            if pTemp1.nLeaves == p.nVarsMax && pTemp2.nLeaves == p.nVarsMax {
                if pTemp1.ppLeaves[0] != pTemp2.ppLeaves[0] {
                    continue; }
                if pTemp1.ppLeaves[1] != pTemp2.ppLeaves[1] {
                    continue; }
            }


            // check if k-feasible cut exists
            nNodes = Fpga_CutMergeTwo( pTemp1, pTemp2, ppNodes, p.nVarsMax );
            if nNodes == 0 {
                continue; }
            // consider the cut for possible addition to the set of new cuts
            pCut = Fpga_CutTableConsider( p, pTable, ppNodes, nNodes );
            if pCut == nil {
                continue; }
            // add data to the cut
            // inlining of macro Fpga_CutNotCond
            pCut.pOne = ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pTemp1) ^ (fComp1)));
            // inlining of macro Fpga_CutNotCond
            pCut.pTwo = ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pTemp2) ^ (fComp2)));
            // create the signature
            pCut.uSign = pTemp1.uSign | pTemp2.uSign;
            // add it to the corresponding list
            pCut.pNext = pLists[int32(pCut.nLeaves)];
            pLists[int32(pCut.nLeaves)] = pCut;
            // count this cut and quit if limit is reached
            Counter++;
            if Counter == FPGA_CUTS_MAX_COMPUTE {
                goto QUITS; }
        }
    }
    // consider the rest of them
    for  i = nCuts1; i < nCuts2; i++  {
        for  k = 0; k < nCuts1; k++  {
            pTemp1 = ppArray1[k];
            pTemp2 = ppArray2[i];

            if pTemp1.nLeaves == p.nVarsMax && pTemp2.nLeaves == p.nVarsMax {
                if pTemp1.ppLeaves[0] != pTemp2.ppLeaves[0] {
                    continue; }
                if pTemp1.ppLeaves[1] != pTemp2.ppLeaves[1] {
                    continue; }
                if pTemp1.ppLeaves[2] != pTemp2.ppLeaves[2] {
                    continue; }
            }


            // check if k-feasible cut exists
            nNodes = Fpga_CutMergeTwo( pTemp1, pTemp2, ppNodes, p.nVarsMax );
            if nNodes == 0 {
                continue; }
            // consider the cut for possible addition to the set of new cuts
            pCut = Fpga_CutTableConsider( p, pTable, ppNodes, nNodes );
            if pCut == nil {
                continue; }
            // add data to the cut
            // inlining of macro Fpga_CutNotCond
            pCut.pOne = ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pTemp1) ^ (fComp1)));
            // inlining of macro Fpga_CutNotCond
            pCut.pTwo = ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pTemp2) ^ (fComp2)));
            // create the signature
            pCut.uSign = pTemp1.uSign | pTemp2.uSign;
            // add it to the corresponding list
            pCut.pNext = pLists[int32(pCut.nLeaves)];
            pLists[int32(pCut.nLeaves)] = pCut;
            // count this cut and quit if limit is reached
            Counter++;
            if Counter == FPGA_CUTS_MAX_COMPUTE {
                goto QUITS; }
        }
    }
QUITS :
    // combine all the lists into one
    pListNew  = nil;
    ppListNew = &pListNew;
    for  i = 1; i <= p.nVarsMax; i++  {
        if pLists[i] == nil {
            continue; }
        // find the last entry
        for  pPrev = pLists[i], pCut = pPrev.pNext; (pCut != nil); 
            pPrev = pCut, pCut = pCut.pNext  {}
        // connect these lists
        *ppListNew = pLists[i];
        ppListNew  = &pPrev.pNext;
    }
    *ppListNew = nil;
    // sort the cuts by arrival times and use only the first FPGA_CUTS_MAX_USE
    pListNew = Fpga_CutSortCuts( p, pTable, pListNew );
    return pListNew;
}


/**Function*************************************************************

  Synopsis    [Merges two lists of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutMergeLists2(p *Fpga_Man_t, pTable *Fpga_CutTable_t, pList1 *Fpga_Cut_t, pList2 *Fpga_Cut_t, fComp1 int32, fComp2 int32, fPivot1 int32, fPivot2 int32) *Fpga_Cut_t {
    var ppNodes [FPGA_MAX_LEAVES]*Fpga_Node_t;
    var pListNew *Fpga_Cut_t;
    var ppListNew **Fpga_Cut_t;
    var pLists [FPGA_MAX_LEAVES+1]*Fpga_Cut_t = [FPGA_MAX_LEAVES+1]*Fpga_Cut_t{ nil };
    var pCut, pPrev, pTemp1, pTemp2 *Fpga_Cut_t;
    var nNodes, Counter, i int32;

    // prepare the manager for the cut computation
    Fpga_CutTableRestart( pTable );
    // go through the cut pairs
    Counter = 0;
    for  pTemp1 = pList1; pTemp1; pTemp1 = fPivot1? nil: pTemp1.pNext  {
        for  pTemp2 = pList2; (pTemp2 != nil); pTemp2 = fPivot2? nil: pTemp2.pNext  {
            // check if k-feasible cut exists
            nNodes = Fpga_CutMergeTwo( pTemp1, pTemp2, ppNodes, p.nVarsMax );
            if nNodes == 0 {
                continue; }
            // consider the cut for possible addition to the set of new cuts
            pCut = Fpga_CutTableConsider( p, pTable, ppNodes, nNodes );
            if pCut == nil {
                continue; }
            // add data to the cut
            // inlining of macro Fpga_CutNotCond
            pCut.pOne = ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pTemp1) ^ (fComp1)));
            // inlining of macro Fpga_CutNotCond
            pCut.pTwo = ((Fpga_Cut_t *)((ABC_PTRUINT_T)(pTemp2) ^ (fComp2)));
            // add it to the corresponding list
            pCut.pNext = pLists[int32(pCut.nLeaves)];
            pLists[int32(pCut.nLeaves)] = pCut;
            // count this cut and quit if limit is reached
            Counter++;
            if Counter == FPGA_CUTS_MAX_COMPUTE {
                goto QUITS; }
        }
    }
QUITS :
    // combine all the lists into one
    pListNew  = nil;
    ppListNew = &pListNew;
    for  i = 1; i <= p.nVarsMax; i++  {
        if pLists[i] == nil {
            continue; }
        // find the last entry
        for  pPrev = pLists[i], pCut = pPrev.pNext; (pCut != nil); 
            pPrev = pCut, pCut = pCut.pNext  {}
        // connect these lists
        *ppListNew = pLists[i];
        ppListNew  = &pPrev.pNext;
    }
    *ppListNew = nil;
    // sort the cuts by arrival times and use only the first FPGA_CUTS_MAX_USE
    pListNew = Fpga_CutSortCuts( p, pTable, pListNew );
    return pListNew;
}

/**Function*************************************************************

  Synopsis    [Merges two cuts.]

  Description [Returns the number of nodes in the resulting cut, or 0 if the
  cut is infeasible. Returns the resulting nodes in the array ppNodes[].]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutMergeTwo(pCut1 *Fpga_Cut_t, pCut2 *Fpga_Cut_t, ppNodes []*Fpga_Node_t, nNodesMax int32) int32 {
    var pNodeTemp *Fpga_Node_t;
    var nTotal, i, k, min, Counter int32;
    var uSign uint32;

    // use quick prefiltering
    uSign = pCut1.uSign | pCut2.uSign;
    // inlining of macro FPGA_COUNT_ONES
    Counter = (bit_count[(uSign)&255]+bit_count[((uSign)>>8)&255]+bit_count[((uSign)>>16)&255]+bit_count[(uSign)>>24]);
    if Counter > nNodesMax {
        return 0; }
/*
    // check the special case when at least of the cuts is the largest
    if ( pCut1->nLeaves == nNodesMax )
    {
        if ( pCut2->nLeaves == nNodesMax )
        {
            // return 0 if the cuts are different
            for ( i = 0; i < nNodesMax; i++ )
                if ( pCut1->ppLeaves[i] != pCut2->ppLeaves[i] )
                    return 0;
            // return nNodesMax if they are the same
            for ( i = 0; i < nNodesMax; i++ )
                ppNodes[i] = pCut1->ppLeaves[i];
            return nNodesMax;
        }
        else if ( pCut2->nLeaves == nNodesMax - 1 ) 
        {
            // return 0 if the cuts are different
            fMismatch = 0;
            for ( i = 0; i < nNodesMax; i++ )
                if ( pCut1->ppLeaves[i] != pCut2->ppLeaves[i - fMismatch] )
                {
                    if ( fMismatch == 1 )
                        return 0;
                    fMismatch = 1;
                }
            // return nNodesMax if they are the same
            for ( i = 0; i < nNodesMax; i++ )
                ppNodes[i] = pCut1->ppLeaves[i];
            return nNodesMax;
        }
    }
    else if ( pCut1->nLeaves == nNodesMax - 1 && pCut2->nLeaves == nNodesMax )
    {
        // return 0 if the cuts are different
        fMismatch = 0;
        for ( i = 0; i < nNodesMax; i++ )
            if ( pCut1->ppLeaves[i - fMismatch] != pCut2->ppLeaves[i] )
            {
                if ( fMismatch == 1 )
                    return 0;
                fMismatch = 1;
            }
        // return nNodesMax if they are the same
        for ( i = 0; i < nNodesMax; i++ )
            ppNodes[i] = pCut2->ppLeaves[i];
        return nNodesMax;
    }
*/
    // count the number of unique entries in pCut2
    nTotal = pCut1.nLeaves;
    for  i = 0; i < pCut2.nLeaves; i++  {
        // try to find this entry among the leaves of pCut1
        for  k = 0; k < pCut1.nLeaves; k++  {
            if pCut2.ppLeaves[i] == pCut1.ppLeaves[k] {
                break; }
        }
        if k < pCut1.nLeaves { // found
            continue; }
        // we found a new entry to add
        if nTotal == nNodesMax {
            return 0; }
        ppNodes[nTotal++] = pCut2.ppLeaves[i];
    }
    // we know that the feasible cut exists

    // add the starting entries
    for  k = 0; k < pCut1.nLeaves; k++  {
        ppNodes[k] = pCut1.ppLeaves[k];
    }

    // selection-sort the entries
    for  i = 0; i < nTotal - 1; i++  {
        min = i;
        for  k = i+1; k < nTotal; k++  {
//            if ( ppNodes[k] < ppNodes[min] ) // reported bug fix (non-determinism!)
            if ppNodes[k].Num < ppNodes[min].Num {
                min = k; }
        }
        pNodeTemp    = ppNodes[i];
        ppNodes[i]   = ppNodes[min];
        ppNodes[min] = pNodeTemp;
    }

    return nTotal;
}

/**Function*************************************************************

  Synopsis    [Computes the union of the two lists of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutUnionLists(pList1 *Fpga_Cut_t, pList2 *Fpga_Cut_t) *Fpga_Cut_t {
    var pTemp, pRoot *Fpga_Cut_t;
    // find the last cut in the first list
    pRoot = pList1;
    // inlining of macro Fpga_ListForEachCut
    for  pTemp = pList1;                                             pTemp;                                                     pTemp = pTemp.pNext  {
        pRoot = pTemp;
    }
    // attach the non-trival part of the second cut to the end of the first
    assert(pRoot.pNext == nil, "pRoot.pNext == nil");
    pRoot.pNext = pList2.pNext;   
    pList2.pNext = nil;
    return pList1;
}


/**Function*************************************************************

  Synopsis    [Checks whether the given cut belongs to the list.]

  Description [This procedure takes most of the runtime in the cut 
  computation.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutBelongsToList(pList *Fpga_Cut_t, ppNodes []*Fpga_Node_t, nNodes int32) int32 {
    var pTemp *Fpga_Cut_t;
    var i int32;
    for  pTemp = pList; pTemp; pTemp = pTemp.pNext  {
        for  i = 0; i < nNodes; i++  {
            if pTemp.ppLeaves[i] != ppNodes[i] {
                break; }
        }
        if i == nNodes {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Counts all the cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutCountAll(pMan *Fpga_Man_t) int32 {
    var pNode *Fpga_Node_t;
    var pCut *Fpga_Cut_t;
    var i, nCuts int32;
    // go through all the nodes in the unique table of the manager
    nCuts = 0;
    for  i = 0; i < pMan.nBins; i++  {
        for  pNode = pMan.pBins[i]; (pNode != nil); pNode = pNode.pNext  {
            for  pCut = pNode.pCuts; (pCut != nil); pCut = pCut.pNext  {
                if pCut.nLeaves > 1 { // skip the elementary cuts
                
//                    Fpga_CutVolume( pCut );
                    nCuts++;
                }
            }
        }
    }
    return nCuts;
}


/**Function*************************************************************

  Synopsis    [Clean the signatures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutsCleanSign(pMan *Fpga_Man_t) {
    var pNode *Fpga_Node_t;
    var pCut *Fpga_Cut_t;
    var i int32;
    for  i = 0; i < pMan.nBins; i++  {
        for  pNode = pMan.pBins[i]; (pNode != nil); pNode = pNode.pNext  {
            for  pCut = pNode.pCuts; (pCut != nil); pCut = pCut.pNext  {
                pCut.uSign = 0;
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Clean the signatures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutsCleanRoot(pMan *Fpga_Man_t) {
    var pNode *Fpga_Node_t;
    var pCut *Fpga_Cut_t;
    var i int32;
    for  i = 0; i < pMan.nBins; i++  {
        for  pNode = pMan.pBins[i]; (pNode != nil); pNode = pNode.pNext  {
            for  pCut = pNode.pCuts; (pCut != nil); pCut = pCut.pNext  {
                pCut.pRoot = nil;
            }
        }
    }
}



/**Function*************************************************************

  Synopsis    [Prints the cuts in the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutListPrint(pMan *Fpga_Man_t, pRoot *Fpga_Node_t) {
    var pTemp *Fpga_Cut_t;
    var Counter int32;
    for  Counter, pTemp = 0, pRoot.pCuts; pTemp; pTemp, Counter = pTemp.pNext, Counter+1  {
        fmt.Printf("%2d : ", Counter + 1);
        Fpga_CutPrint_( pMan, pTemp, pRoot );
    }
}

/**Function*************************************************************

  Synopsis    [Prints the cuts in the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutListPrint2(pMan *Fpga_Man_t, pRoot *Fpga_Node_t) {
    var pTemp *Fpga_Cut_t;
    var Counter int32;
    for  Counter, pTemp = 0, pRoot.pCuts; pTemp; pTemp, Counter = pTemp.pNext, Counter+1  {
        fmt.Printf("%2d : ", Counter + 1);
        Fpga_CutPrint_( pMan, pTemp, pRoot );
    }
}

/**Function*************************************************************

  Synopsis    [Prints the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutPrint_(pMan *Fpga_Man_t, pCut *Fpga_Cut_t, pRoot *Fpga_Node_t) {
    var i int32;
    fmt.Printf("(%3d)  {", pRoot.Num);
    for  i = 0; i < pMan.nVarsMax; i++  {
        if pCut.ppLeaves[i] {
            fmt.Printf(" %3d", pCut.ppLeaves[i].Num); }
    }
    fmt.Printf(" }\n");
}








/**Function*************************************************************

  Synopsis    [Starts the hash table to canonicize cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutTableStart(pMan *Fpga_Man_t) *Fpga_CutTable_t {
    var p *Fpga_CutTable_t;
    // allocate the table
    // inlining of macro ABC_ALLOC
    p = ((Fpga_CutTable_t *) malloc(sizeof(Fpga_CutTable_t) * (size_t)(1)));
    memset( p, 0, sizeof(Fpga_CutTable_t) );
    p.nBins = Abc_PrimeCudd( 10 * FPGA_CUTS_MAX_COMPUTE );
    // inlining of macro ABC_ALLOC
    p.pBins = ((Fpga_Cut_t * *) malloc(sizeof(Fpga_Cut_t *) * (size_t)(p.nBins)));
    memset( p.pBins, 0, sizeof(Fpga_Cut_t *) * p.nBins );
    // inlining of macro ABC_ALLOC
    p.pCuts = ((int *) malloc(sizeof(int) * (size_t)(2 * FPGA_CUTS_MAX_COMPUTE)));
    // inlining of macro ABC_ALLOC
    p.pArray = ((Fpga_Cut_t * *) malloc(sizeof(Fpga_Cut_t *) * (size_t)(2 * FPGA_CUTS_MAX_COMPUTE)));
    // inlining of macro ABC_ALLOC
    p.pCuts1 = ((Fpga_Cut_t * *) malloc(sizeof(Fpga_Cut_t *) * (size_t)(2 * FPGA_CUTS_MAX_COMPUTE)));
    // inlining of macro ABC_ALLOC
    p.pCuts2 = ((Fpga_Cut_t * *) malloc(sizeof(Fpga_Cut_t *) * (size_t)(2 * FPGA_CUTS_MAX_COMPUTE)));
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutTableStop(p *Fpga_CutTable_t) {
    p.pCuts1 = nil;
    p.pCuts2 = nil;
    p.pArray = nil;
    p.pBins = nil;
    p.pCuts = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Computes the hash value of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutTableHash(ppNodes []*Fpga_Node_t, nNodes int32) uint32 {
    var uRes uint32;
    var i int32;
    uRes = 0;
    for  i = 0; i < nNodes; i++  {
        uRes += s_HashPrimes[i] * ppNodes[i].Num;
    }
    return uRes;
}

/**Function*************************************************************

  Synopsis    [Looks up the table for the available cut.]

  Description [Returns -1 if the same cut is found. Returns the index
  of the cell where the cut should be added, if it does not exist.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutTableLookup(p *Fpga_CutTable_t, ppNodes []*Fpga_Node_t, nNodes int32) int32 {
    var pCut *Fpga_Cut_t;
    var Key uint32;
    var b, i int32;

    Key = Fpga_CutTableHash(ppNodes, nNodes) % p.nBins; 
    for  b = Key; p.pBins[b]; b = (b+1) % p.nBins  {
        pCut = p.pBins[b];
        if pCut.nLeaves != nNodes {
            continue; }
        for  i = 0; i < nNodes; i++  {
            if pCut.ppLeaves[i] != ppNodes[i] {
                break; }
        }
        if i == nNodes {
            return -1; }
    }
    return b;
}


/**Function*************************************************************

  Synopsis    [Starts the hash table to canonicize cuts.]

  Description [Considers addition of the cut to the hash table.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutTableConsider(pMan *Fpga_Man_t, p *Fpga_CutTable_t, ppNodes []*Fpga_Node_t, nNodes int32) *Fpga_Cut_t {
    var pCut *Fpga_Cut_t;
    var Place, i int32;
    // check the cut
    Place = Fpga_CutTableLookup( p, ppNodes, nNodes );
    if Place == -1 {
        return nil; }
    assert(nNodes > 0, "nNodes > 0");
    // create the new cut
    pCut = Fpga_CutAlloc( pMan );
    pCut.nLeaves = nNodes;
    pCut.fLevel = 0.0;
    for  i = 0; i < nNodes; i++  {
        pCut.ppLeaves[i] = ppNodes[i];
        pCut.fLevel += ppNodes[i].Level;
    }
    pCut.fLevel /= nNodes;
    // add the cut to the table
    assert(p.pBins[Place] == nil, "p.pBins[Place] == nil");
    p.pBins[Place] = pCut;
    // add the cut to the new list
    p.pCuts[ p.nCuts++ ] = Place;
    return pCut;
}

/**Function*************************************************************

  Synopsis    [Prepares the table to be used with other cuts.]

  Description [Restarts the table by cleaning the info about cuts stored
  when the previous node was considered.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutTableRestart(p *Fpga_CutTable_t) {
    var i int32;
    for  i = 0; i < p.nCuts; i++  {
        assert(p.pBins[ p.pCuts[i] ], "p.pBins[ p.pCuts[i] ]");
        p.pBins[ p.pCuts[i] ] = nil;
    }
    p.nCuts = 0;
}



/**Function*************************************************************

  Synopsis    [Compares the cuts by the number of leaves and then by delay.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutSortCutsCompare(pC1 **Fpga_Cut_t, pC2 **Fpga_Cut_t) int32 {
    if (*pC1).nLeaves < (*pC2).nLeaves {
        return -1; }
    if (*pC1).nLeaves > (*pC2).nLeaves {
        return 1; }
/*
    if ( (*pC1)->fLevel > (*pC2)->fLevel )
        return -1;
    if ( (*pC1)->fLevel < (*pC2)->fLevel )
        return 1;
*/
    return 0;
}

/**Function*************************************************************

  Synopsis    [Sorts the cuts by average arrival time.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutSortCuts(pMan *Fpga_Man_t, p *Fpga_CutTable_t, pList *Fpga_Cut_t) *Fpga_Cut_t {
    var pListNew *Fpga_Cut_t;
    var nCuts, i int32;
    // move the cuts from the list into the array
    nCuts = Fpga_CutList2Array( p.pCuts1, pList );
    assert(nCuts <= FPGA_CUTS_MAX_COMPUTE, "nCuts <= FPGA_CUTS_MAX_COMPUTE");
    // sort the cuts
    qsort( (void *)p.pCuts1, (size_t)nCuts, sizeof(void *), 
                (int (*)(void *, void *)) Fpga_CutSortCutsCompare );
    // move them back into the list
    if nCuts > FPGA_CUTS_MAX_USE - 1 {
//        printf( "*" );
        // free the remaining cuts
        for  i = FPGA_CUTS_MAX_USE - 1; i < nCuts; i++  {
            Extra_MmFixedEntryRecycle( pMan.mmCuts, (char *)p.pCuts1[i] );
        }
        // update the number of cuts
        nCuts = FPGA_CUTS_MAX_USE - 1;
    }
    pListNew = Fpga_CutArray2List( p.pCuts1, nCuts );
    return pListNew;
}

/**Function*************************************************************

  Synopsis    [Moves the nodes from the list into the array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutList2Array(pArray **Fpga_Cut_t, pList *Fpga_Cut_t) int32 {
    var i int32;
    for  i = 0; (pList != nil); pList, i = pList.pNext, i+1  {
        pArray[i] = pList;
    }
    return i;
}

/**Function*************************************************************

  Synopsis    [Moves the nodes from the array into the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutArray2List(pArray **Fpga_Cut_t, nCuts int32) *Fpga_Cut_t {
    var pListNew *Fpga_Cut_t;
    var ppListNew **Fpga_Cut_t;
    var i int32;
    pListNew  = nil;
    ppListNew = &pListNew;
    for  i = 0; i < nCuts; i++  {
        // connect these lists
        *ppListNew = pArray[i];
        ppListNew  = &pArray[i].pNext;
//printf( " %d(%.2f)", pArray[i]->nLeaves, pArray[i]->fLevel );
    }
//printf( "\n" );

    *ppListNew = nil;
    return pListNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END


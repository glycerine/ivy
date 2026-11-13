package fra

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [fraImp.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [New FRAIG package.]

  Synopsis    [Detecting and proving implications.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 30, 2007.]

  Revision    [$Id: fraImp.c,v 1.00 2007/06/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Counts the number of 1s in each siminfo of each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlCountOnesOne(p *Fra_Sml_t, Node int32) int32 {
    var pSim *uint32;
    var k, Counter int32;
    pSim = Fra_ObjSim( p, Node );
    for  k = p.nWordsPref; k < p.nWordsTotal; k++  {
        Counter += Aig_WordCountOnes( pSim[k] );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of 1s in each siminfo of each node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlCountOnes(p *Fra_Sml_t) *int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    var pnBits *int32; 
    // inlining of macro ABC_ALLOC
    pnBits = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManObjNumMax(p.pAig))));  
    memset( pnBits, 0, sizeof(int) * Aig_ManObjNumMax(p.pAig) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        pnBits[i] = Fra_SmlCountOnesOne( p, i ); }
    }
    return pnBits;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if implications holds.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Counts the number of 1s in the complement of the implication.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sml_NodeNotImpWeight(p *Fra_Sml_t, Left int32, Right int32) int32 {
    var pSimL, pSimR *uint32;
    var k, Counter int32;
    pSimL = Fra_ObjSim( p, Left );
    pSimR = Fra_ObjSim( p, Right );
    for  k = p.nWordsPref; k < p.nWordsTotal; k++  {
        Counter += Aig_WordCountOnes( pSimL[k] & ~pSimR[k] );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Computes the complement of the implication.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sml_NodeSaveNotImpPatterns(p *Fra_Sml_t, Left int32, Right int32, pResult *uint32) {
    var pSimL, pSimR *uint32;
    var k int32;
    pSimL = Fra_ObjSim( p, Left );
    pSimR = Fra_ObjSim( p, Right );
    for  k = p.nWordsPref; k < p.nWordsTotal; k++  {
        pResult[k] |= pSimL[k] & ~pSimR[k];
    }
}

/**Function*************************************************************

  Synopsis    [Returns the array of nodes sorted by the number of 1s.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlSortUsingOnes(p *Fra_Sml_t, fLatchCorr int32) *Vec_Ptr_t {
    var pObj *Aig_Obj_t;
    var vNodes *Vec_Ptr_t;
    var i int32;
    var nNodes int32;
    var nTotal int32;
    var nBits int32;
    var pnNodes *int32;
    var pnBits *int32;
    var pMemory *int32;
    assert(p.nWordsTotal > 0, "p.nWordsTotal > 0");
    // count 1s in each node's siminfo
    pnBits = Fra_SmlCountOnes( p );
    // count number of nodes having that many 1s
    nNodes = 0;
    nBits = p.nWordsTotal * 32;
    // inlining of macro ABC_ALLOC
    pnNodes = ((int *) malloc(sizeof(int) * (size_t)(nBits + 1)));
    memset( pnNodes, 0, sizeof(int) * (nBits + 1) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if i == 0 { continue; }
        // skip non-PI and non-internal nodes
        if fLatchCorr != 0 {
            if !Aig_ObjIsCi(pObj) {
                continue; }
        
        } else {
            if !Aig_ObjIsNode(pObj) && !Aig_ObjIsCi(pObj) {
                continue; }
        }
        // skip nodes participating in the classes
//        if ( Fra_ClassObjRepr(pObj) )
//            continue;
        assert(pnBits[i] <= nBits, "pnBits[i] <= nBits"); // "<" because of normalized info
        pnNodes[pnBits[i]]++;
        nNodes++;
    }
    }
    // allocate memory for all the nodes
    // inlining of macro ABC_ALLOC
    pMemory = ((int *) malloc(sizeof(int) * (size_t)(nNodes + nBits + 1)));  
    // markup the memory for each node
    vNodes = Vec_PtrAlloc( nBits + 1 );
    Vec_PtrPush( vNodes, pMemory );
    for  i = 1; i <= nBits; i++  {
        pMemory += pnNodes[i-1] + 1;
        Vec_PtrPush( vNodes, pMemory );
    }
    // add the nodes
    memset( pnNodes, 0, sizeof(int) * (nBits + 1) );
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if i == 0 { continue; }
        // skip non-PI and non-internal nodes
        if fLatchCorr != 0 {
            if !Aig_ObjIsCi(pObj) {
                continue; }
        
        } else {
            if !Aig_ObjIsNode(pObj) && !Aig_ObjIsCi(pObj) {
                continue; }
        }
        // skip nodes participating in the classes
//        if ( Fra_ClassObjRepr(pObj) )
//            continue;
        pMemory = (int *)Vec_PtrEntry( vNodes, pnBits[i] );
        pMemory[ pnNodes[pnBits[i]]++ ] = i;
    }
    }
    // add 0s in the end
    nTotal = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pMemory) = (int *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pMemory[ pnNodes[i]++ ] = 0;
        nTotal += pnNodes[i];
    }
    assert(nTotal == nNodes + nBits + 1, "nTotal == nNodes + nBits + 1");
    pnNodes = nil;
    pnBits = nil;
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Returns the array of implications with the highest cost.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_SmlSelectMaxCost(vImps *Vec_Int_t, pCosts *int32, nCostMax int32, nImpLimit int32, pCostRange *int32) *Vec_Int_t {
    var vImpsNew *Vec_Int_t;
    var pCostCount *int32;
    var nImpCount int32;
    var Imp int32;
    var i int32;
    var c int32;
    assert(Vec_IntSize(vImps) >= nImpLimit, "Vec_IntSize(vImps) >= nImpLimit");
    // count how many implications have each cost
    // inlining of macro ABC_ALLOC
    pCostCount = ((int *) malloc(sizeof(int) * (size_t)(nCostMax + 1)));
    memset( pCostCount, 0, sizeof(int) * (nCostMax + 1) );
    for  i = 0; i < Vec_IntSize(vImps); i++  {
        assert(pCosts[i] <= nCostMax, "pCosts[i] <= nCostMax");
        pCostCount[ pCosts[i] ]++;
    }
    assert(pCostCount[0] == 0, "pCostCount[0] == 0");
    // select the bound on the cost (above this bound, implication will be included)
    nImpCount = 0;
    for  c = nCostMax; c > 0; c--  {
        nImpCount += pCostCount[c];
        if nImpCount >= nImpLimit {
            break; }
    }
//    printf( "Cost range >= %d.\n", c );
    // collect implications with the given costs
    vImpsNew = Vec_IntAlloc( nImpLimit );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vImps)) && (((Imp) = Vec_IntEntry(vImps, i)), 1); i++  {
        if pCosts[i] < c {
            continue; }
        Vec_IntPush( vImpsNew, Imp );
        if Vec_IntSize( vImpsNew ) == nImpLimit {
            break; }
    }
    pCostCount = nil;
    if pCostRange != nil {
        *pCostRange = c; }
    return vImpsNew;
}

/**Function*************************************************************

  Synopsis    [Compares two implications using their largest ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
int Sml_CompareMaxId( unsigned short * pImp1, unsigned short * pImp2 )
{
    var Max1 int32  = Abc_MaxInt( pImp1[0], pImp1[1] );
    var Max2 int32  = Abc_MaxInt( pImp2[0], pImp2[1] );
    if Max1 < Max2 {
        return -1; }
    if Max1 > Max2 {
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Derives implication candidates.]

  Description [Implication candidates have the property that 
  (1) they hold using sequential simulation information
  (2) they do not hold using combinational simulation information
  (3) they have as high expressive power as possible (heuristically)
      that is, they are easy to disprove combinationally
      meaning they cover relatively larger sequential subspace.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ImpDerive(p *Fra_Man_t, nImpMaxLimit int32, nImpUseLimit int32, fLatchCorr int32) *Vec_Int_t {
    var nSimWords int32  = 64;
    var pSeq, pComb *Fra_Sml_t;
    var vImps, vTemp *Vec_Int_t;
    var vNodes *Vec_Ptr_t;
    var pImpCosts, pNodesI, pNodesK *int32;
    var nImpsTotal, nImpsTried, nImpsNonSeq, nImpsComb, nImpsCollected int32;
    var CostMin int32 = ABC_INFINITY;
    var CostMax int32;
    var i, k, Imp, CostRange int32;
    var clk abctime  = Abc_Clock();
    assert(Aig_ManObjNumMax(p.pManAig) < (1 << 15), "Aig_ManObjNumMax(p.pManAig) < (1 << 15)");
    assert(nImpMaxLimit > 0 && nImpUseLimit > 0 && nImpUseLimit <= nImpMaxLimit, "nImpMaxLimit > 0 && nImpUseLimit > 0 && nImpUseLimit <= nImpMaxLimit");
    // normalize both managers
    pComb = Fra_SmlSimulateComb( p.pManAig, nSimWords, 0 );
    pSeq = Fra_SmlSimulateSeq( p.pManAig, p.pPars.nFramesP, nSimWords, 1, 1 );
    // get the nodes sorted by the number of 1s
    vNodes = Fra_SmlSortUsingOnes( pSeq, fLatchCorr );
    // count the total number of implications
    for  k = nSimWords * 32; k > 0; k--  {
    for  i = k - 1; i > 0; i--  {
    for  pNodesI = (int *)Vec_PtrEntry( vNodes, i ); *pNodesI; pNodesI++  {
    for  pNodesK = (int *)Vec_PtrEntry( vNodes, k ); *pNodesK; pNodesK++  {
        nImpsTotal++;
    }
    }
    }
    }

    // compute implications and their costs
    // inlining of macro ABC_ALLOC
    pImpCosts = ((int *) malloc(sizeof(int) * (size_t)(nImpMaxLimit)));
    vImps = Vec_IntAlloc( nImpMaxLimit );
    for  k = pSeq.nWordsTotal * 32; k > 0; k--  {
        for  i = k - 1; i > 0; i--  {
            // HERE WE ARE MISSING SOME POTENTIAL IMPLICATIONS (with complement!)

            for  pNodesI = (int *)Vec_PtrEntry( vNodes, i ); *pNodesI; pNodesI++  {
            for  pNodesK = (int *)Vec_PtrEntry( vNodes, k ); *pNodesK; pNodesK++  {
                nImpsTried++;
                if !(Sml_NodeCheckImp(pSeq, *pNodesI, *pNodesK) != 0) {
                    nImpsNonSeq++;
                    continue;
                }
                if Sml_NodeCheckImp(pComb, *pNodesI, *pNodesK) != 0 {
                    nImpsComb++;
                    continue;
                }
                nImpsCollected++;
                Imp = Fra_ImpCreate( *pNodesI, *pNodesK );
                pImpCosts[ Vec_IntSize(vImps) ] = Sml_NodeNotImpWeight(pComb, *pNodesI, *pNodesK);
                CostMin = Abc_MinInt( CostMin, pImpCosts[ Vec_IntSize(vImps) ] );
                CostMax = Abc_MaxInt( CostMax, pImpCosts[ Vec_IntSize(vImps) ] );
                Vec_IntPush( vImps, Imp );
                if Vec_IntSize(vImps) == nImpMaxLimit {
                    goto finish; }
            }
            } 
        }
    }
finish:
    Fra_SmlStop( pComb );
    Fra_SmlStop( pSeq );

    // select implications with the highest cost
    CostRange = CostMin;
    if Vec_IntSize(vImps) > nImpUseLimit {
        vImps = Fra_SmlSelectMaxCost( vTemp = vImps, pImpCosts, nSimWords * 32, nImpUseLimit, &CostRange );
        Vec_IntFree( vTemp );  
    }

    // dealloc
    pImpCosts = nil; 
    {
    var pTemp any  = Vec_PtrEntry(vNodes, 0);
    pTemp = nil;
    }
    Vec_PtrFree( vNodes );
    // reorder implications topologically
    qsort( (void *)Vec_IntArray(vImps), (size_t)Vec_IntSize(vImps), sizeof(int), 
                (int (*)(void *, void *)) Sml_CompareMaxId );
if p.pPars.fVerbose {
fmt.Printf("Implications: All = %d. Try = %d. NonSeq = %d. Comb = %d. Res = %d.\n", nImpsTotal, nImpsTried, nImpsNonSeq, nImpsComb, nImpsCollected);
fmt.Printf("Implication weight: Min = %d. Pivot = %d. Max = %d.   ", CostMin, CostRange, CostMax);
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}
    return vImps;
}


// the following three procedures are called to 
// - add implications to the SAT solver
// - check implications using the SAT solver
// - refine implications using after a cex is generated

/**Function*************************************************************

  Synopsis    [Add implication clauses to the SAT solver.]

  Description [Note that implications should be checked in the first frame!]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ImpAddToSolver(p *Fra_Man_t, vImps *Vec_Int_t, pSatVarNums *int32) {
    var pSat *sat_solver  = p.pSat;
    var pLeft, pRight *Aig_Obj_t;
    var pLeftF, pRightF *Aig_Obj_t;
    var pLits [2]int32;
    var Imp int32;
    var Left int32;
    var Right int32;
    var i int32;
    var f int32;
    var status int32;
    var fComplL, fComplR int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vImps)) && (((Imp) = Vec_IntEntry(vImps, i)), 1); i++  {
        // get the corresponding nodes
        pLeft = Aig_ManObj( p.pManAig, Fra_ImpLeft(Imp) );
        pRight = Aig_ManObj( p.pManAig, Fra_ImpRight(Imp) );
        // check if all the nodes are present
        for  f = 0; f < p.pPars.nFramesK; f++  {
            // map these info fraig
            pLeftF = Fra_ObjFraig( pLeft, f );
            pRightF = Fra_ObjFraig( pRight, f );
            if Aig_ObjIsNone(Aig_Regular(pLeftF)) || Aig_ObjIsNone(Aig_Regular(pRightF)) {
                Vec_IntWriteEntry( vImps, i, 0 );
                break;
            }
        } 
        if f < p.pPars.nFramesK {
            continue; }
        // add constraints in each timeframe
        for  f = 0; f < p.pPars.nFramesK; f++  {
            // map these info fraig
            pLeftF = Fra_ObjFraig( pLeft, f );
            pRightF = Fra_ObjFraig( pRight, f );
            // get the corresponding SAT numbers
            Left = pSatVarNums[ Aig_Regular(pLeftF).Id ];
            Right = pSatVarNums[ Aig_Regular(pRightF).Id ];
            assert(Left > 0  && Left < p.nSatVars, "Left > 0  && Left < p.nSatVars");
            assert(Right > 0 && Right < p.nSatVars, "Right > 0 && Right < p.nSatVars");
            // get the complemented attributes
            fComplL = pLeft.fPhase ^ Aig_IsComplement(pLeftF);
            fComplR = pRight.fPhase ^ Aig_IsComplement(pRightF);
            // get the constraint
            // L => R      L' v R     (complement = L & R')
            pLits[0] = 2 * Left  + !fComplL;
            pLits[1] = 2 * Right +  fComplR;
            // add constraint to solver
            if !sat_solver_addclause( pSat, pLits, pLits + 2 ) {
                sat_solver_delete( pSat );
                p.pSat = nil;
                return;
            }
        }
    }
    status = sat_solver_simplify(pSat);
    if status == 0 {
        sat_solver_delete( pSat );
        p.pSat = nil;
    }
//    printf( "Total imps = %d. ", Vec_IntSize(vImps) );
    Fra_ImpCompactArray( vImps );
//    printf( "Valid imps = %d. \n", Vec_IntSize(vImps) );
}

/**Function*************************************************************

  Synopsis    [Check implications for the node (if they are present).]

  Description [Returns the new position in the array.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ImpCheckForNode(p *Fra_Man_t, vImps *Vec_Int_t, pNode *Aig_Obj_t, Pos int32) int32 {
    var pLeft, pRight *Aig_Obj_t;
    var pLeftF, pRightF *Aig_Obj_t;
    var i, Imp, Left, Right, Max, RetValue int32;
    var fComplL, fComplR int32;
    // inlining of macro Vec_IntForEachEntryStart
    for  i = Pos; (i < Vec_IntSize(vImps)) && (((Imp) = Vec_IntEntry(vImps, i)), 1); i++  {
        if Imp == 0 {
            continue; }
        Left = Fra_ImpLeft(Imp);
        Right = Fra_ImpRight(Imp);
        Max = Abc_MaxInt( Left, Right );
        assert(Max >= pNode.Id, "Max >= pNode.Id");
        if Max > pNode.Id {
            return i; }
        // get the corresponding nodes
        pLeft  = Aig_ManObj( p.pManAig, Left );
        pRight = Aig_ManObj( p.pManAig, Right );
        // get the corresponding FRAIG nodes
        pLeftF  = Fra_ObjFraig( pLeft, p.pPars.nFramesK );
        pRightF = Fra_ObjFraig( pRight, p.pPars.nFramesK );
        // get the complemented attributes
        fComplL = pLeft.fPhase ^ Aig_IsComplement(pLeftF);
        fComplR = pRight.fPhase ^ Aig_IsComplement(pRightF);
        // check equality
        if Aig_Regular(pLeftF) == Aig_Regular(pRightF) {
            if fComplL == fComplR { // x => x  - always true
                continue; }
            assert(fComplL != fComplR, "fComplL != fComplR");
            // consider 4 possibilities:
            // NOT(1) => 1    or   0 => 1  - always true
            // 1 => NOT(1)    or   1 => 0  - never true
            // NOT(x) => x    or   x       - not always true
            // x => NOT(x)    or   NOT(x)  - not always true
            if Aig_ObjIsConst1(Aig_Regular(pLeftF)) && fComplL { // proved implication
                continue; }
            // disproved implication
            p.pCla.fRefinement = 1;
            Vec_IntWriteEntry( vImps, i, 0 );
            continue;
        }
        // check the implication 
        // - if true, a clause is added
        // - if false, a cex is simulated
        // make sure the implication is refined
        RetValue = Fra_NodesAreImp( p, Aig_Regular(pLeftF), Aig_Regular(pRightF), fComplL, fComplR );
        if RetValue != 1 {
            p.pCla.fRefinement = 1;
            if RetValue == 0 {
                Fra_SmlResimulate( p ); }
            if Vec_IntEntry(vImps, i) != 0 {
                fmt.Printf("Fra_ImpCheckForNode(): Implication is not refined!\n"); }
            assert(Vec_IntEntry(vImps, i) == 0, "Vec_IntEntry(vImps, i) == 0");
        }
    }
    return i;
}

/**Function*************************************************************

  Synopsis    [Removes those implications that no longer hold.]

  Description [Returns 1 if refinement has happened.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Removes empty implications.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Determines the ratio of the state space by computed implications.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns the number of failed implications.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ImpVerifyUsingSimulation(p *Fra_Man_t) int32 {
    var nFrames int32  = 2000;
    var nSimWords int32  = 8;
    var pSeq *Fra_Sml_t;
    var pfFails string;
    var Left, Right, Imp, i, Counter int32;
    if p.pCla.vImps == nil || Vec_IntSize(p.pCla.vImps) == 0 {
        return 0; }
    // simulate the AIG manager with combinational patterns
    pSeq = Fra_SmlSimulateSeq( p.pManAig, p.pPars.nFramesP, nFrames, nSimWords, 1  );
    // go through the implications and check how many of them do not hold
    // inlining of macro ABC_ALLOC
    pfFails = ((char *) malloc(sizeof(char) * (size_t)(Vec_IntSize(p.pCla.vImps))));
    memset( pfFails, 0, sizeof(char) * Vec_IntSize(p.pCla.vImps) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.pCla.vImps)) && (((Imp) = Vec_IntEntry(p.pCla.vImps, i)), 1); i++  {
        Left = Fra_ImpLeft(Imp);
        Right = Fra_ImpRight(Imp);
        pfFails[i] = !Sml_NodeCheckImp( pSeq, Left, Right );
    }
    // count how many has failed
    Counter = 0;
    for  i = 0; i < Vec_IntSize(p.pCla.vImps); i++  {
        Counter += pfFails[i];
    }
    pfFails = nil;
    Fra_SmlStop( pSeq );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Record proven implications in the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fra_ImpRecordInManager(p *Fra_Man_t, pNew *Aig_Man_t) {
    var pLeft, pRight, pMiter *Aig_Obj_t;
    var nPosOld, Imp, i int32;
    if p.pCla.vImps == nil || Vec_IntSize(p.pCla.vImps) == 0 {
        return; }
    // go through the implication
    nPosOld = Aig_ManCoNum(pNew);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.pCla.vImps)) && (((Imp) = Vec_IntEntry(p.pCla.vImps, i)), 1); i++  {
        pLeft = Aig_ManObj( p.pManAig, Fra_ImpLeft(Imp) );
        pRight = Aig_ManObj( p.pManAig, Fra_ImpRight(Imp) );
        // record the implication: L' + R
        pMiter = Aig_Or( pNew, 
            Aig_NotCond((Aig_Obj_t *)pLeft.pData, !pLeft.fPhase), 
            Aig_NotCond((Aig_Obj_t *)pRight.pData, pRight.fPhase) ); 
        Aig_ObjCreateCo( pNew, pMiter );
    }
    pNew.nAsserts = Aig_ManCoNum(pNew) - nPosOld;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


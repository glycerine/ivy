package sim

import "fmt"

/**CFile****************************************************************

  FileName    [simSupp.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Simulation to determine functional support.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: simSupp.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "proof/fraig/fraig.h"
//#include "sim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
// static int    Sim_ComputeSuppRound( Sim_Man_t * p, int  fUseTargets );

// static int    Sim_ComputeSuppRoundNode( Sim_Man_t * p, int iNumCi, int  fUseTargets );

// static void   Sim_ComputeSuppSetTargets( Sim_Man_t * p );


// static void   Sim_UtilAssignRandom( Sim_Man_t * p );

// static void   Sim_UtilAssignFromFifo( Sim_Man_t * p );

// static void   Sim_SolveTargetsUsingSat( Sim_Man_t * p, int nCounters );

// static int    Sim_SolveSuppModelVerify( Abc_Ntk_t * pNtk, int * pModel, int Input, int Output );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes structural supports.]

  Description [Supports are returned as an array of bit strings, one
  for each CO.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_ComputeStrSupp(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vSuppStr *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var pSimmNode, pSimmNode1, pSimmNode2 *uint32;
    var nSuppWords, i, k int32;
    // allocate room for structural supports
    // inlining of macro SIM_NUM_WORDS
    nSuppWords = (((Abc_NtkCiNum(pNtk))>>5) + (((Abc_NtkCiNum(pNtk))&31) > 0));
    vSuppStr   = Sim_UtilInfoAlloc( Abc_NtkObjNumMax(pNtk), nSuppWords, 1 );
    // assign the structural support to the PIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        // inlining of macro Sim_SuppStrSetVar
        // inlining of macro Sim_SetBit
        (((unsigned*)(vSuppStr).pArray[(pNode).Id])[((i))>>5] |= (1<<(((i)) & 31)));
    }
    // derive the structural supports of the internal nodes
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
//        if ( Abc_NodeIsConst(pNode) )
//            continue;
        pSimmNode  = (unsigned *)vSuppStr.pArray[ pNode.Id ];
        pSimmNode1 = (unsigned *)vSuppStr.pArray[ Abc_ObjFaninId0(pNode) ];
        pSimmNode2 = (unsigned *)vSuppStr.pArray[ Abc_ObjFaninId1(pNode) ];
        for  k = 0; k < nSuppWords; k++  {
            pSimmNode[k] = pSimmNode1[k] | pSimmNode2[k];
        }
    }
    }
    // set the structural supports of the PO nodes
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pNode) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pSimmNode  = (unsigned *)vSuppStr.pArray[ pNode.Id ];
        pSimmNode1 = (unsigned *)vSuppStr.pArray[ Abc_ObjFaninId0(pNode) ];
        for  k = 0; k < nSuppWords; k++  {
            pSimmNode[k] = pSimmNode1[k];
        }
    }
    return vSuppStr;
}

/**Function*************************************************************

  Synopsis    [Compute functional supports.]

  Description [Supports are returned as an array of bit strings, one
  for each CO.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_ComputeFunSupp(pNtk *Abc_Ntk_t, fVerbose int32) *Vec_Ptr_t {
    var p *Sim_Man_t;
    var vResult *Vec_Ptr_t;
    var nSolved, i int32;
    var clk abctime  = Abc_Clock();

    srand( 0xABC );

    // start the simulation manager
    p = Sim_ManStart( pNtk, 0 );

    // compute functional support using one round of random simulation
    Sim_UtilAssignRandom( p );
    Sim_ComputeSuppRound( p, 0 );

    // set the support targets 
    Sim_ComputeSuppSetTargets( p );
    if fVerbose != 0 {
        fmt.Printf("Number of support targets after simulation = %5d.\n", Vec_VecSizeSize(p.vSuppTargs)); }
    if Vec_VecSizeSize(p.vSuppTargs) == 0 {
        goto exit; }

    for  i = 0; i < 1; i++  {
        // compute patterns using one round of random simulation
        Sim_UtilAssignRandom( p );
        nSolved = Sim_ComputeSuppRound( p, 1 );
        if Vec_VecSizeSize(p.vSuppTargs) == 0 {
            goto exit; }

        if fVerbose != 0 {
            fmt.Printf("Targets = %5d.   Solved = %5d.  Fifo = %5d.\n", Vec_VecSizeSize(p.vSuppTargs), nSolved, Vec_PtrSize(p.vFifo)); }
    }

    // try to solve the support targets
    for  Vec_VecSizeSize(p.vSuppTargs) > 0  {
        // solve targets until the first disproved one (which gives counter-example)
        Sim_SolveTargetsUsingSat( p, p.nSimWords/p.nSuppWords );
        // compute additional functional support
        Sim_UtilAssignFromFifo( p );
        nSolved = Sim_ComputeSuppRound( p, 1 );

if fVerbose != 0 {
    fmt.Printf("Targets = %5d.   Solved = %5d.  Fifo = %5d.  SAT runs = %3d.\n", Vec_VecSizeSize(p.vSuppTargs), nSolved, Vec_PtrSize(p.vFifo), p.nSatRuns); }
    }

exit:
p.timeTotal = Abc_Clock() - clk;
    vResult = p.vSuppFun;  
    //  p->vSuppFun = NULL;
    Sim_ManStop( p );
    return vResult;
}

/**Function*************************************************************

  Synopsis    [Computes functional support using one round of simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_ComputeSuppRound(p *Sim_Man_t, fUseTargets int32) int32 {
    var vTargets *Vec_Ptr_t;
    var i, Counter int32;
    var clk abctime;
    // perform one round of random simulation
clk = Abc_Clock();
    Sim_UtilSimulate( p, 0 );
p.timeSim += Abc_Clock() - clk;
    // iterate through the CIs and detect COs that depend on them
    for  i = p.iInput; i < p.nInputs; i++  {
        vTargets = (Vec_Ptr_t *)p.vSuppTargs.pArray[i];
        if (fUseTargets != 0) && vTargets.nSize == 0 {
            continue; }
        Counter += Sim_ComputeSuppRoundNode( p, i, fUseTargets );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Computes functional support for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_ComputeSuppRoundNode(p *Sim_Man_t, iNumCi int32, fUseTargets int32) int32 {
    var fVerbose int32;
    var pPat *Sim_Pat_t;
    var vTargets *Vec_Ptr_t;
    var vNodesByLevel *Vec_Vec_t;
    var pNodeCi, pNode *Abc_Obj_t;
    var i, k, v, Output, LuckyPat, fType0, fType1 int32;
    var Counter int32;
    var fFirst int32  = 1;
    var clk abctime;
    // collect nodes by level in the TFO of the CI 
    // this procedure does not collect the CIs and COs
    // but it increments TravId of the collected nodes and CIs/COs
clk = Abc_Clock();
    pNodeCi       = Abc_NtkCi( p.pNtk, iNumCi );
    vNodesByLevel = Abc_DfsLevelized( pNodeCi, 0 );
p.timeTrav += Abc_Clock() - clk;
    // complement the simulation info of the selected CI
    Sim_UtilInfoFlip( p, pNodeCi );
    // simulate the levelized structure of nodes
    // inlining of macro Vec_VecForEachEntry
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; i < Vec_VecSize(vNodesByLevel); i++  {                                                        for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(vNodesByLevel, i))) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(Vec_VecEntry(vNodesByLevel, i), k)), 1); k++  {
        fType0 = Abc_NodeIsTravIdCurrent( Abc_ObjFanin0(pNode) );
        fType1 = Abc_NodeIsTravIdCurrent( Abc_ObjFanin1(pNode) );
clk = Abc_Clock();
        Sim_UtilSimulateNode( p, pNode, 1, fType0, fType1 );
p.timeSim += Abc_Clock() - clk;
    }
    }
    // set the simulation info of the affected COs
    if fUseTargets != 0 {
        vTargets = (Vec_Ptr_t *)p.vSuppTargs.pArray[iNumCi];
        for  i = vTargets.nSize - 1; i >= 0; i--  {
            // get the target output
            Output = int32((ABC_PTRUINT_T))vTargets.pArray[i];
            // get the target node
            pNode  = Abc_ObjFanin0( Abc_NtkCo(p.pNtk, Output) );
            // the output should be in the cone
            assert(Abc_NodeIsTravIdCurrent(pNode), "Abc_NodeIsTravIdCurrent(pNode)");

            // skip if the simulation info is equal
            if Sim_UtilInfoCompare( p, pNode ) {
                continue; }

            // otherwise, we solved a new target
            Vec_PtrRemove( vTargets, vTargets.pArray[i] );
if fVerbose != 0 {
    fmt.Printf("(%d,%d) ", iNumCi, Output); }
            Counter++;
            // make sure this variable is not yet detected
            // inlining of macro Sim_SuppFunHasVar
            // inlining of macro Sim_HasBit
            assert(!((((unsigned*)(p.vSuppFun).pArray[Output])[((iNumCi))>>5]  & (1<<(((iNumCi)) & 31))) > 0), "!/*__starting_point_inline_macro:Sim_SuppFunHasVar__*//*__starting_point_inline_macro:Sim_HasBit__*/((((unsigned*)(p.vSuppFun).pArray[Output])[((iNumCi))>>5]  & (1<<(((iNumCi)) & 31))) > 0)");
            // set this variable
            // inlining of macro Sim_SuppFunSetVar
            // inlining of macro Sim_SetBit
            (((unsigned*)(p.vSuppFun).pArray[Output])[((iNumCi))>>5] |= (1<<(((iNumCi)) & 31)));
            
            // detect the differences in the simulation info
            Sim_UtilInfoDetectDiffs( (unsigned *)p.vSim0.pArray[pNode.Id], (unsigned *)p.vSim1.pArray[pNode.Id], p.nSimWords, p.vDiffs );
            // create new patterns
            if !(fFirst != 0) && p.vFifo.nSize > 1000 {
                continue; }

            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(p.vDiffs)) && (((LuckyPat) = Vec_IntEntry(p.vDiffs, k)), 1); k++  {
                // set the new pattern
                pPat = Sim_ManPatAlloc( p );
                pPat.Input  = iNumCi;
                pPat.Output = Output;
                // inlining of macro Abc_NtkForEachCi
                for  v = 0; (v < Abc_NtkCiNum(p.pNtk)) && (((pNodeCi) = Abc_NtkCi(p.pNtk, v)), 1); v++  {
                    // inlining of macro Sim_SimInfoHasVar
                    // inlining of macro Sim_HasBit
                    if  (((unsigned*)(p.vSim0).pArray[(pNodeCi).Id])[((LuckyPat))>>5]  & (1<<(((LuckyPat)) & 31))) > 0 {
                        // inlining of macro Sim_SetBit
                        ((pPat.pData)[(v)>>5] |= (1<<((v) & 31))); }
                }
                Vec_PtrPush( p.vFifo, pPat );

                fFirst = 0;
                break;
            }
        }
if (fVerbose != 0) && (Counter != 0) {
fmt.Printf("\n"); }
    
    } else {
        // inlining of macro Abc_NtkForEachCo
        for  Output = 0; (Output < Abc_NtkCoNum(p.pNtk)) && (((pNode) = Abc_NtkCo(p.pNtk, Output)), 1); Output++  {
            if !Abc_NodeIsTravIdCurrent( pNode ) {
                continue; }
            if !Sim_UtilInfoCompare( p, Abc_ObjFanin0(pNode) ) {
                // inlining of macro Sim_SuppFunHasVar
                // inlining of macro Sim_HasBit
                if !((((unsigned*)(p.vSuppFun).pArray[Output])[((iNumCi))>>5]  & (1<<(((iNumCi)) & 31))) > 0) {
                    Counter++;
                    // inlining of macro Sim_SuppFunSetVar
                    // inlining of macro Sim_SetBit
                    (((unsigned*)(p.vSuppFun).pArray[Output])[((iNumCi))>>5] |= (1<<(((iNumCi)) & 31)));
                }
            }
        }
    }
    Vec_VecFree( vNodesByLevel );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Sets the simulation targets.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_ComputeSuppSetTargets(p *Sim_Man_t) {
    var pNode *Abc_Obj_t;
    var pSuppStr, pSuppFun *uint32;
    var i, k, Num int32;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pNode) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        pSuppStr = (unsigned *)p.vSuppStr.pArray[pNode.Id];
        pSuppFun = (unsigned *)p.vSuppFun.pArray[i];
        // find vars in the structural support that are not in the functional support
        Sim_UtilInfoDetectNews( pSuppFun, pSuppStr, p.nSuppWords, p.vDiffs );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(p.vDiffs)) && (((Num) = Vec_IntEntry(p.vDiffs, k)), 1); k++  {
            Vec_VecPush( p.vSuppTargs, Num, (void *)(ABC_PTRUINT_T)i );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Assigns random simulation info to the PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilAssignRandom(p *Sim_Man_t) {
    var pNode *Abc_Obj_t;
    var pSimInfo *uint32;
    var i, k int32;
    // assign the random/systematic simulation info to the PIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(p.pNtk)) && (((pNode) = Abc_NtkCi(p.pNtk, i)), 1); i++  {
        pSimInfo = (unsigned *)p.vSim0.pArray[pNode.Id];
        for  k = 0; k < p.nSimWords; k++  {
            pSimInfo[k] = SIM_RANDOM_UNSIGNED;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Sets the new patterns from fifo.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_UtilAssignFromFifo(p *Sim_Man_t) {
    var fUseOneWord int32;
    var pNode *Abc_Obj_t;
    var pPat *Sim_Pat_t;
    var pSimInfo *uint32;
    var nWordsNew, iWord, iWordLim, i, w int32;
    var iBeg, iEnd int32;
    var Counter int32;
    // go through the patterns and fill in the dist-1 minterms for each
    for  iWord = 0; p.vFifo.nSize > 0; iWord = iWordLim  {
        ++Counter;
        // get the pattern
        pPat = (Sim_Pat_t *)Vec_PtrPop( p.vFifo );
        if fUseOneWord != 0 {
            // get the first word of the next series
            iWordLim = iWord + 1; 
            // set the pattern for all PIs from iBit to iWord + p->nInputs
            iBeg = p.iInput;
            iEnd = Abc_MinInt( iBeg + 32, p.nInputs );
//            for ( i = iBeg; i < iEnd; i++ )
            // inlining of macro Abc_NtkForEachCi
            for  i = 0; (i < Abc_NtkCiNum(p.pNtk)) && (((pNode) = Abc_NtkCi(p.pNtk, i)), 1); i++  {
                pNode = Abc_NtkCi(p.pNtk,i);
                pSimInfo = (unsigned *)p.vSim0.pArray[pNode.Id];
                // inlining of macro Sim_HasBit
                if  ((pPat.pData)[(i)>>5]  & (1<<((i) & 31))) > 0 {
                    pSimInfo[iWord] = SIM_MASK_FULL;
                } else {
                    pSimInfo[iWord] = 0; }
                // flip one bit
                if i >= iBeg && i < iEnd {
                    // inlining of macro Sim_XorBit
                    ((pSimInfo + iWord)[(i-iBeg)>>5] ^= (1<<((i-iBeg) & 31))); }
            }
        
        } else {
            // get the number of words for the remaining inputs
            nWordsNew = p.nSuppWords;
//            nWordsNew = SIM_NUM_WORDS( p->nInputs - p->iInput );
            // get the first word of the next series
            if iWord + nWordsNew < p.nSimWords {
                iWordLim = iWord + nWordsNew;
            } else {
                iWordLim = p.nSimWords;
            } 
            // set the pattern for all CIs from iWord to iWord + nWordsNew
            // inlining of macro Abc_NtkForEachCi
            for  i = 0; (i < Abc_NtkCiNum(p.pNtk)) && (((pNode) = Abc_NtkCi(p.pNtk, i)), 1); i++  {
                pSimInfo = (unsigned *)p.vSim0.pArray[pNode.Id];
                // inlining of macro Sim_HasBit
                if  ((pPat.pData)[(i)>>5]  & (1<<((i) & 31))) > 0 {
                    for  w = iWord; w < iWordLim; w++  {
                        pSimInfo[w] = SIM_MASK_FULL;
                    }
                
                } else {
                    for  w = iWord; w < iWordLim; w++  {
                        pSimInfo[w] = 0;
                    }
                }
                // inlining of macro Sim_XorBit
                ((pSimInfo + iWord)[(i)>>5] ^= (1<<((i) & 31)));
                // flip one bit
//                if ( i >= p->iInput )
//                    Sim_XorBit( pSimInfo + iWord, i-p->iInput );
            }
        }
        Sim_ManPatFree( p, pPat );
        // stop if we ran out of room for patterns
        if iWordLim == p.nSimWords {
            break; }
//        if ( Counter == 1 )
//            break;
    }
}

/**Function*************************************************************

  Synopsis    [Get the given number of counter-examples using SAT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SolveTargetsUsingSat(p *Sim_Man_t, Limit int32) {
    var Params Fraig_Params_t;
    var pMan *Fraig_Man_t;
    var pNodeCi *Abc_Obj_t;
    var pMiter *Abc_Ntk_t;
    var pPat *Sim_Pat_t;
    var pEntry any;
    var pModel *int32;
    var RetValue, Output, Input, k, v int32;
    var Counter int32;
    var clk abctime;

    p.nSatRuns = 0;
    // put targets into one array
    // inlining of macro Vec_VecForEachEntryReverse
    // inlining of macro Vec_PtrForEachEntryReverse
    for  Input = 0; Input < Vec_VecSize(p.vSuppTargs); Input++  {                                                        for  k = Vec_PtrSize(Vec_VecEntry(p.vSuppTargs, Input)) - 1; (k >= 0) && (((pEntry) = (void *)Vec_PtrEntry(Vec_VecEntry(p.vSuppTargs, Input), k)), 1); k--  {
        p.nSatRuns++;
        Output = int32((ABC_PTRUINT_T))pEntry;

        // set up the miter for the two cofactors of this output w.r.t. this input
        pMiter = Abc_NtkMiterForCofactors( p.pNtk, Output, Input, -1 );

        // transform the miter into a fraig
        Fraig_ParamsSetDefault( &Params );
        Params.nSeconds = ABC_INFINITY;
        Params.fInternal = 1;
clk = Abc_Clock();
        pMan = (Fraig_Man_t *)Abc_NtkToFraig( pMiter, &Params, 0, 0 ); 
p.timeFraig += Abc_Clock() - clk;
clk = Abc_Clock();
        Fraig_ManProveMiter( pMan );
p.timeSat += Abc_Clock() - clk;

        // analyze the result
        RetValue = Fraig_ManCheckMiter( pMan );
        assert(RetValue >= 0, "RetValue >= 0");
        if RetValue == 1 { // unsat
        
            p.nSatRunsUnsat++;
            pModel = nil;
            Vec_PtrRemove( (Vec_Ptr_t *)p.vSuppTargs.pArray[Input], pEntry );
        
        } else { // sat
        
            p.nSatRunsSat++;
            pModel = Fraig_ManReadModel( pMan );
            assert(pModel != nil, "pModel != nil");
            assert((Sim_SolveSuppModelVerify( p.pNtk, pModel, Input, Output ) != 0), "Sim_SolveSuppModelVerify( p.pNtk, pModel, Input, Output )");

//printf( "Solved by SAT (%d,%d).\n", Input, Output );
            // set the new pattern
            pPat = Sim_ManPatAlloc( p );
            pPat.Input  = Input;
            pPat.Output = Output;
            // inlining of macro Abc_NtkForEachCi
            for  v = 0; (v < Abc_NtkCiNum(p.pNtk)) && (((pNodeCi) = Abc_NtkCi(p.pNtk, v)), 1); v++  {
                if pModel[v] {
                    // inlining of macro Sim_SetBit
                    ((pPat.pData)[(v)>>5] |= (1<<((v) & 31))); }
            }
            Vec_PtrPush( p.vFifo, pPat );
/*
            // set the new pattern
            pPat = Sim_ManPatAlloc( p );
            pPat->Input  = Input;
            pPat->Output = Output;
            Abc_NtkForEachCi( p->pNtk, pNodeCi, v )
                if ( pModel[v] )
                    Sim_SetBit( pPat->pData, v );
            Sim_XorBit( pPat->pData, Input );  // add this bit in the opposite polarity
            Vec_PtrPush( p->vFifo, pPat );
*/
            Counter++;
        }
        // delete the fraig manager
        Fraig_ManFree( pMan );
        // delete the miter
        Abc_NtkDelete( pMiter );

        // makr the input, which we are processing
        p.iInput = Input;

        // stop when we found enough patterns
//        if ( Counter == Limit )
        if Counter == 1 {
            return; }
    }
    }
}


/**Function*************************************************************

  Synopsis    [Saves the counter example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_NtkSimTwoPats_rec(pNode *Abc_Obj_t) int32 {
    var Value0, Value1 int32;
    if Abc_NodeIsTravIdCurrent( pNode ) {
        return int32((ABC_PTRUINT_T))pNode.pCopy; }
    Abc_NodeSetTravIdCurrent( pNode );
    Value0 = Sim_NtkSimTwoPats_rec( Abc_ObjFanin0(pNode) );
    Value1 = Sim_NtkSimTwoPats_rec( Abc_ObjFanin1(pNode) );
    if Abc_ObjFaninC0(pNode) {
        Value0 = ~Value0; }
    if Abc_ObjFaninC1(pNode) {
        Value1 = ~Value1; }
    pNode.pCopy = (Abc_Obj_t *)(ABC_PTRUINT_T)(Value0 & Value1);
    return Value0 & Value1;
}

/**Function*************************************************************

  Synopsis    [Verifies that pModel proves the presence of Input in the support of Output.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SolveSuppModelVerify(pNtk *Abc_Ntk_t, pModel *int32, Input int32, Output int32) int32 {
    var pNode *Abc_Obj_t;
    var RetValue, i int32;
    // set the PI values
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pNode) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pNode );
        if pNode == Abc_NtkCi(pNtk,Input) {
            pNode.pCopy = (Abc_Obj_t *)1;
        } else if pModel[i] == 1 {
            pNode.pCopy = (Abc_Obj_t *)3;
        } else {
            pNode.pCopy = nil; }
    }
    // perform the traversal
    RetValue = 3 & Sim_NtkSimTwoPats_rec( Abc_ObjFanin0( Abc_NtkCo(pNtk,Output) ) );
//    assert( RetValue == 1 || RetValue == 2 ); 
    return RetValue == 1 || RetValue == 2;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


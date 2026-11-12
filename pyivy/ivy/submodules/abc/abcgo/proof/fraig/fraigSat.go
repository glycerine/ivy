package fraig

import (
	"fmt"
	"math"
)

/**CFile****************************************************************

  FileName    [fraigSat.c]

  PackageName [FRAIG: Functionally reduced AND-INV graphs.]

  Synopsis    [Proving functional equivalence using SAT.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - October 1, 2004]

  Revision    [$Id: fraigSat.c,v 1.10 2005/07/08 01:01:32 alanmi Exp $]

***********************************************************************/

//#include <math.h>
//#include "fraigInt.h"
//#include "sat/msat/msatInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Fraig_OrderVariables( Fraig_Man_t * pMan, Fraig_Node_t * pOld, Fraig_Node_t * pNew );

// static void Fraig_SetupAdjacent( Fraig_Man_t * pMan, Msat_IntVec_t * vConeVars );

// static void Fraig_SetupAdjacentMark( Fraig_Man_t * pMan, Msat_IntVec_t * vConeVars );

// static void Fraig_PrepareCones( Fraig_Man_t * pMan, Fraig_Node_t * pOld, Fraig_Node_t * pNew );

// static void Fraig_PrepareCones_rec( Fraig_Man_t * pMan, Fraig_Node_t * pNode );


// static void Fraig_SupergateAddClauses( Fraig_Man_t * pMan, Fraig_Node_t * pNode, Fraig_NodeVec_t * vSuper );

// static void Fraig_SupergateAddClausesExor( Fraig_Man_t * pMan, Fraig_Node_t * pNode );

// static void Fraig_SupergateAddClausesMux( Fraig_Man_t * pMan, Fraig_Node_t * pNode );

//static void Fraig_DetectFanoutFreeCone( Fraig_Man_t * pMan, Fraig_Node_t * pNode );
// static void Fraig_DetectFanoutFreeConeMux( Fraig_Man_t * pMan, Fraig_Node_t * pNode );

// static void Fraig_SetActivity( Fraig_Man_t * pMan, Fraig_Node_t * pOld, Fraig_Node_t * pNew );


// The lesson learned seems to be that variable should be in reverse topological order
// from the output of the miter. The ordering of adjacency lists is very important.
// The best way seems to be fanins followed by fanouts. Slight changes to this order
// leads to big degradation in quality.

var nMuxes int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checks equivalence of two nodes.]

  Description [Returns 1 iff the nodes are equivalent.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodesAreEqual(p *Fraig_Man_t, pNode1 *Fraig_Node_t, pNode2 *Fraig_Node_t, nBTLimit int32, nTimeLimit int32) int32 {
    if pNode1 == pNode2 {
        return 1; }
    // inlining of macro Fraig_Not
    if pNode1 == ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2) ^ 01)) {
        return 0; }
    // inlining of macro Fraig_Regular
    return Fraig_NodeIsEquivalent( p, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode1) & ~01)) , ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2) & ~01)) , nBTLimit, nTimeLimit );
}

/**Function*************************************************************

  Synopsis    [Tries to prove the final miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManProveMiter(p *Fraig_Man_t) {
    var pNode *Fraig_Node_t;
    var i int32;
    var clk abctime;

    if !p.fTryProve {
        return; }
 
    clk = Abc_Clock();
    // consider all outputs of the multi-output miter
    for  i = 0; i < p.vOutputs.nSize; i++  {
        // inlining of macro Fraig_Regular
        pNode = ((Fraig_Node_t *)((ABC_PTRUINT_T)(p.vOutputs.pArray[i]) & ~01)) ;
        // skip already constant nodes
        if pNode == p.pConst1 {
            continue; }
        // skip nodes that are different according to simulation
        if !Fraig_CompareSimInfo( pNode, p.pConst1, p.nWordsRand, 1 ) {
            continue; }
        if Fraig_NodeIsEquivalent( p, p.pConst1, pNode, -1, p.nSeconds ) != 0 {
            // inlining of macro Fraig_IsComplement
            if  ((int32(((ABC_PTRUINT_T) (p.vOutputs.pArray[i]) & 01)))) ^ Fraig_NodeComparePhase(p.pConst1, pNode) {
                // inlining of macro Fraig_Not
                p.vOutputs.pArray[i] = ((Fraig_Node_t *)((ABC_PTRUINT_T)(p.pConst1) ^ 01)) ;
            } else {
                p.vOutputs.pArray[i] = p.pConst1; }
        }
    }
    if p.fVerboseP {
//        ABC_PRT( "Final miter proof time", Abc_Clock() - clk );
    }
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the miter is unsat; 0 if sat; -1 if undecided.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManCheckMiter(p *Fraig_Man_t) int32 {
    var pNode *Fraig_Node_t;
    var i int32;
    // inlining of macro ABC_FREE
    ((p.pModel) ? (free((char *) (p.pModel)), (p.pModel) = 0) : 0);
    for  i = 0; i < p.vOutputs.nSize; i++  {
        // get the output node (it can be complemented!)
        pNode = p.vOutputs.pArray[i];
        // if the miter is constant 0, the problem is UNSAT
        // inlining of macro Fraig_Not
        if pNode == ((Fraig_Node_t *)((ABC_PTRUINT_T)(p.pConst1) ^ 01)) {
            continue; }
        // consider the special case when the miter is constant 1
        if pNode == p.pConst1 {
            // in this case, any counter example will do to distinquish it from constant 0
            // here we pick the counter example composed of all zeros
            p.pModel = Fraig_ManAllocCounterExample( p );
            return 0;
        }
        // save the counter example
        p.pModel = Fraig_ManSaveCounterExample( p, pNode );
        // if the model is not found, return undecided
        if p.pModel == nil {
            return -1;
        } else {
            return 0; }
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_MarkTfi_rec(pMan *Fraig_Man_t, pNode *Fraig_Node_t) int32 {
    // skip the visited node
    if pNode.TravId == pMan.nTravIds {
        return 0; }
    pNode.TravId = pMan.nTravIds;
    // skip the PI node
    if pNode.NumPi >= 0 {
        return 1; }
    // check the children
    // inlining of macro Fraig_Regular
    return Fraig_MarkTfi_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01))  ) +
           // inlining of macro Fraig_Regular
           Fraig_MarkTfi_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01))  );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_MarkTfi2_rec(pMan *Fraig_Man_t, pNode *Fraig_Node_t) int32 {
    // skip the visited node
    if pNode.TravId == pMan.nTravIds {
        return 0; }
    // skip the boundary node
    if pNode.TravId == pMan.nTravIds-1 {
        pNode.TravId = pMan.nTravIds;
        return 1;
    }
    pNode.TravId = pMan.nTravIds;
    // skip the PI node
    if pNode.NumPi >= 0 {
        return 1; }
    // check the children
    // inlining of macro Fraig_Regular
    return Fraig_MarkTfi2_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01))  ) +
           // inlining of macro Fraig_Regular
           Fraig_MarkTfi2_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01))  );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if pOld is in the TFI of pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_MarkTfi3_rec(pMan *Fraig_Man_t, pNode *Fraig_Node_t) int32 {
    // skip the visited node
    if pNode.TravId == pMan.nTravIds {
        return 1; }
    // skip the boundary node
    if pNode.TravId == pMan.nTravIds-1 {
        pNode.TravId = pMan.nTravIds;
        return 1;
    }
    pNode.TravId = pMan.nTravIds;
    // skip the PI node
    if pNode.NumPi >= 0 {
        return 0; }
    // check the children
    // inlining of macro Fraig_Regular
    return Fraig_MarkTfi3_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01))  ) *
           // inlining of macro Fraig_Regular
           Fraig_MarkTfi3_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01))  );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_VarsStudy(p *Fraig_Man_t, pOld *Fraig_Node_t, pNew *Fraig_Node_t) {
    var NumPis, NumCut, fContain int32;

    // mark the TFI of pNew
    p.nTravIds++;
    NumPis = Fraig_MarkTfi_rec( p, pNew );
    fmt.Printf("(%d)(%d,%d):", NumPis, pOld.Level, pNew.Level);

    // check if the old is in the TFI
    if pOld.TravId == p.nTravIds {
        fmt.Printf("* ");
        return;
    }

    // count the boundary of nodes in pOld
    p.nTravIds++;
    NumCut = Fraig_MarkTfi2_rec( p, pOld );
    fmt.Printf("%d", NumCut);

    // check if the new is contained in the old's support
    p.nTravIds++;
    fContain = Fraig_MarkTfi3_rec( p, pNew );
    fmt.Printf("%c ", fContain? '+':'-');
}


/**Function*************************************************************

  Synopsis    [Checks whether two nodes are functinally equivalent.]

  Description [The flag (fComp) tells whether the nodes to be checked
  are in the opposite polarity. The second flag (fSkipZeros) tells whether
  the checking should be performed if the simulation vectors are zeros.
  Returns 1 if the nodes are equivalent; 0 othewise.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeIsEquivalent(p *Fraig_Man_t, pOld *Fraig_Node_t, pNew *Fraig_Node_t, nBTLimit int32, nTimeLimit int32) int32 {
    var RetValue, RetValue1, i, fComp int32;
    var clk abctime;
    var fVerbose int32;
    var fSwitch int32;

    // make sure the nodes are not complemented
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNew) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNew) & 01))))");
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pOld) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pOld) & 01))))");
    assert(pNew != pOld, "pNew != pOld");

    // if at least one of the nodes is a failed node, perform adjustments:
    // if the backtrack limit is small, simply skip this node
    // if the backtrack limit is > 10, take the quare root of the limit
    if nBTLimit > 0 && (pOld.fFailTfo || pNew.fFailTfo) {
        p.nSatFails++;
//            return 0;
//        if ( nBTLimit > 10 )
//            nBTLimit /= 10;
        if nBTLimit <= 10 {
            return 0; }
        nBTLimit = int32(math.Sqrt(float64(float64(nBTLimit))));
//        fSwitch = 1;
    }

    p.nSatCalls++;

    // make sure the solver is allocated and has enough variables
    if p.pSat == nil {
        Fraig_ManCreateSolver( p ); }
    // make sure the SAT solver has enough variables
    for  i = Msat_SolverReadVarNum(p.pSat); i < p.vNodes.nSize; i++  {
        Msat_SolverAddVar( p.pSat, p.vNodes.pArray[i].Level );
    }


 
/*
    {
        Fraig_Node_t * ppNodes[2] = { pOld, pNew };
        extern void Fraig_MappingShowNodes( Fraig_Man_t * pMan, Fraig_Node_t ** ppRoots, int nRoots, char * pFileName );
        Fraig_MappingShowNodes( p, ppNodes, 2, "temp_aig" );
    }
*/

    nMuxes = 0;


    // get the logic cone
clk = Abc_Clock();
//    Fraig_VarsStudy( p, pOld, pNew );
    Fraig_OrderVariables( p, pOld, pNew );
//    Fraig_PrepareCones( p, pOld, pNew );
p.timeTrav += Abc_Clock() - clk;

//    printf( "The number of MUXes detected = %d (%5.2f %% of logic).  ", nMuxes, 300.0*nMuxes/(p->vNodes->nSize - p->vInputs->nSize) );
//    ABC_PRT( "Time", Abc_Clock() - clk );

    if fVerbose != 0 {
        fmt.Printf("%d(%d) - ", Fraig_CountPis(p,p.vVarsInt), Msat_IntVecReadSize(p.vVarsInt)); }


    // prepare variable activity
    Fraig_SetActivity( p, pOld, pNew );

    // get the complemented attribute
    fComp = Fraig_NodeComparePhase( pOld, pNew );
//Msat_SolverPrintClauses( p->pSat );

    ////////////////////////////////////////////
    // prepare the solver to run incrementally on these variables
//clk = Abc_Clock();
    Msat_SolverPrepare( p.pSat, p.vVarsInt );
//p->time3 += Abc_Clock() - clk;


    // solve under assumptions
    // A = 1; B = 0     OR     A = 1; B = 1 
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pOld.Num)+(0))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNew.Num)+(!fComp))  );

//Msat_SolverWriteDimacs( p->pSat, "temp_fraig.cnf" );

    // run the solver
clk = Abc_Clock();
    RetValue1 = Msat_SolverSolve( p.pSat, p.vProj, nBTLimit, nTimeLimit );
p.timeSat += Abc_Clock() - clk;

    if RetValue1 == MSAT_FALSE {
//p->time1 += Abc_Clock() - clk;

if fVerbose != 0 {
//    printf( "unsat  %d  ", Msat_SolverReadBackTracks(p->pSat) );
//ABC_PRT( "time", Abc_Clock() - clk );
}

        // add the clause
        Msat_IntVecClear( p.vProj );
        // inlining of macro MSAT_VAR2LIT
        Msat_IntVecPush( p.vProj, (2*(pOld.Num)+(1))  );
        // inlining of macro MSAT_VAR2LIT
        Msat_IntVecPush( p.vProj, (2*(pNew.Num)+(fComp))  );
        RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
        assert((RetValue != 0), "RetValue");
        // continue solving the other implication
    
    } else if RetValue1 == MSAT_TRUE {
//p->time2 += Abc_Clock() - clk;

if fVerbose != 0 {
//    printf( "sat  %d  ", Msat_SolverReadBackTracks(p->pSat) );
//ABC_PRT( "time", Abc_Clock() - clk );
}

        // record the counter example
        Fraig_FeedBack( p, Msat_SolverReadModelArray(p.pSat), p.vVarsInt, pOld, pNew );

//        if ( pOld->fFailTfo || pNew->fFailTfo )
//            printf( "*" );
//        printf( "s(%d)", pNew->Level );
        if fSwitch != 0 {
             fmt.Printf("s(%d)", pNew.Level); }
        p.nSatCounter++;
        return 0;
    
    } else { // if ( RetValue1 == MSAT_UNKNOWN )
    
p.time3 += Abc_Clock() - clk;

//        if ( pOld->fFailTfo || pNew->fFailTfo )
//            printf( "*" );
//        printf( "T(%d)", pNew->Level );

        // mark the node as the failed node
        if pOld != p.pConst1 { 
            pOld.fFailTfo = 1; }
        pNew.fFailTfo = 1;
//        p->nSatFails++;
        if fSwitch != 0 {
             fmt.Printf("T(%d)", pNew.Level); }
        p.nSatFailsReal++;
        return 0;
    }

    // if the old node was constant 0, we already know the answer
    if pOld == p.pConst1 {
        return 1; }

    ////////////////////////////////////////////
    // prepare the solver to run incrementally 
//clk = Abc_Clock();
    Msat_SolverPrepare( p.pSat, p.vVarsInt );
//p->time3 += Abc_Clock() - clk;
    // solve under assumptions
    // A = 0; B = 1     OR     A = 0; B = 0 
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pOld.Num)+(1))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNew.Num)+(fComp))  );
    // run the solver
clk = Abc_Clock();
    RetValue1 = Msat_SolverSolve( p.pSat, p.vProj, nBTLimit, nTimeLimit );
p.timeSat += Abc_Clock() - clk;

    if RetValue1 == MSAT_FALSE {
//p->time1 += Abc_Clock() - clk;

if fVerbose != 0 {
//    printf( "unsat  %d  ", Msat_SolverReadBackTracks(p->pSat) );
//ABC_PRT( "time", Abc_Clock() - clk );
}

        // add the clause
        Msat_IntVecClear( p.vProj );
        // inlining of macro MSAT_VAR2LIT
        Msat_IntVecPush( p.vProj, (2*(pOld.Num)+(0))  );
        // inlining of macro MSAT_VAR2LIT
        Msat_IntVecPush( p.vProj, (2*(pNew.Num)+(!fComp))  );
        RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
        assert((RetValue != 0), "RetValue");
        // continue solving the other implication
    
    } else if RetValue1 == MSAT_TRUE {
//p->time2 += Abc_Clock() - clk;

if fVerbose != 0 {
//    printf( "sat  %d  ", Msat_SolverReadBackTracks(p->pSat) );
//ABC_PRT( "time", Abc_Clock() - clk );
}

        // record the counter example
        Fraig_FeedBack( p, Msat_SolverReadModelArray(p.pSat), p.vVarsInt, pOld, pNew );
        p.nSatCounter++;

//        if ( pOld->fFailTfo || pNew->fFailTfo )
//            printf( "*" );
//        printf( "s(%d)", pNew->Level );
        if fSwitch != 0 {
             fmt.Printf("s(%d)", pNew.Level); }
        return 0;
    
    } else { // if ( RetValue1 == MSAT_UNKNOWN )
    
p.time3 += Abc_Clock() - clk;

//        if ( pOld->fFailTfo || pNew->fFailTfo )
//            printf( "*" );
//        printf( "T(%d)", pNew->Level );
        if fSwitch != 0 {
             fmt.Printf("T(%d)", pNew.Level); }

        // mark the node as the failed node
        pOld.fFailTfo = 1;
        pNew.fFailTfo = 1;
//        p->nSatFails++;
        p.nSatFailsReal++;
        return 0;
    }

    // return SAT proof
    p.nSatProof++;

//    if ( pOld->fFailTfo || pNew->fFailTfo )
//        printf( "*" );
//    printf( "u(%d)", pNew->Level );

    if fSwitch != 0 {
         fmt.Printf("u(%d)", pNew.Level); }

    return 1;
}


/**Function*************************************************************

  Synopsis    [Checks whether pOld => pNew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_NodeIsImplication(p *Fraig_Man_t, pOld *Fraig_Node_t, pNew *Fraig_Node_t, nBTLimit int32) int32 {
    var RetValue, RetValue1, i, fComp int32;
    var clk abctime;
    var fVerbose int32;

    // make sure the nodes are not complemented
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNew) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNew) & 01))))");
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pOld) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pOld) & 01))))");
    assert(pNew != pOld, "pNew != pOld");

    p.nSatCallsImp++;

    // make sure the solver is allocated and has enough variables
    if p.pSat == nil {
        Fraig_ManCreateSolver( p ); }
    // make sure the SAT solver has enough variables
    for  i = Msat_SolverReadVarNum(p.pSat); i < p.vNodes.nSize; i++  {
        Msat_SolverAddVar( p.pSat, p.vNodes.pArray[i].Level );
    }

   // get the logic cone
clk = Abc_Clock();
    Fraig_OrderVariables( p, pOld, pNew );
//    Fraig_PrepareCones( p, pOld, pNew );
p.timeTrav += Abc_Clock() - clk;

    if fVerbose != 0 {
        fmt.Printf("%d(%d) - ", Fraig_CountPis(p,p.vVarsInt), Msat_IntVecReadSize(p.vVarsInt)); }


    // get the complemented attribute
    fComp = Fraig_NodeComparePhase( pOld, pNew );
//Msat_SolverPrintClauses( p->pSat );

    ////////////////////////////////////////////
    // prepare the solver to run incrementally on these variables
//clk = Abc_Clock();
    Msat_SolverPrepare( p.pSat, p.vVarsInt );
//p->time3 += Abc_Clock() - clk;

    // solve under assumptions
    // A = 1; B = 0     OR     A = 1; B = 1 
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pOld.Num)+(0))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNew.Num)+(!fComp))  );
    // run the solver
clk = Abc_Clock();
    RetValue1 = Msat_SolverSolve( p.pSat, p.vProj, nBTLimit, 1000000 );
p.timeSat += Abc_Clock() - clk;

    if RetValue1 == MSAT_FALSE {
//p->time1 += Abc_Clock() - clk;

if fVerbose != 0 {
//    printf( "unsat  %d  ", Msat_SolverReadBackTracks(p->pSat) );
//ABC_PRT( "time", Abc_Clock() - clk );
}

        // add the clause
        Msat_IntVecClear( p.vProj );
        // inlining of macro MSAT_VAR2LIT
        Msat_IntVecPush( p.vProj, (2*(pOld.Num)+(1))  );
        // inlining of macro MSAT_VAR2LIT
        Msat_IntVecPush( p.vProj, (2*(pNew.Num)+(fComp))  );
        RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
        assert((RetValue != 0), "RetValue");
//        p->nSatProofImp++;
        return 1;
    
    } else if RetValue1 == MSAT_TRUE {
//p->time2 += Abc_Clock() - clk;

if fVerbose != 0 {
//    printf( "sat  %d  ", Msat_SolverReadBackTracks(p->pSat) );
//ABC_PRT( "time", Abc_Clock() - clk );
}
        // record the counter example
        Fraig_FeedBack( p, Msat_SolverReadModelArray(p.pSat), p.vVarsInt, pOld, pNew );
        p.nSatCounterImp++;
        return 0;
    
    } else { // if ( RetValue1 == MSAT_UNKNOWN )
    
p.time3 += Abc_Clock() - clk;
        p.nSatFailsImp++;
        return 0;
    }
}

/**Function*************************************************************

  Synopsis    [Prepares the SAT solver to run on the two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_ManCheckClauseUsingSat(p *Fraig_Man_t, pNode1 *Fraig_Node_t, pNode2 *Fraig_Node_t, nBTLimit int32) int32 {
    var pNode1R, pNode2R *Fraig_Node_t;
    var RetValue, RetValue1, i int32;
    var clk abctime;
    var fVerbose int32;

    // inlining of macro Fraig_Regular
    pNode1R = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode1) & ~01)) ;
    // inlining of macro Fraig_Regular
    pNode2R = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode2) & ~01)) ;
    assert(pNode1R != pNode2R, "pNode1R != pNode2R");

    // make sure the solver is allocated and has enough variables
    if p.pSat == nil {
        Fraig_ManCreateSolver( p ); }
    // make sure the SAT solver has enough variables
    for  i = Msat_SolverReadVarNum(p.pSat); i < p.vNodes.nSize; i++  {
        Msat_SolverAddVar( p.pSat, p.vNodes.pArray[i].Level );
    }

   // get the logic cone
clk = Abc_Clock();
    Fraig_OrderVariables( p, pNode1R, pNode2R );
//    Fraig_PrepareCones( p, pNode1R, pNode2R );
p.timeTrav += Abc_Clock() - clk;

    ////////////////////////////////////////////
    // prepare the solver to run incrementally on these variables
//clk = Abc_Clock();
    Msat_SolverPrepare( p.pSat, p.vVarsInt );
//p->time3 += Abc_Clock() - clk;

    // solve under assumptions
    // A = 1; B = 0     OR     A = 1; B = 1 
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    // inlining of macro Fraig_IsComplement
    Msat_IntVecPush( p.vProj, (2*(pNode1R.Num)+(!((int32(((ABC_PTRUINT_T) (pNode1) & 01))))))  );
    // inlining of macro MSAT_VAR2LIT
    // inlining of macro Fraig_IsComplement
    Msat_IntVecPush( p.vProj, (2*(pNode2R.Num)+(!((int32(((ABC_PTRUINT_T) (pNode2) & 01))))))  );
    // run the solver
clk = Abc_Clock();
    RetValue1 = Msat_SolverSolve( p.pSat, p.vProj, nBTLimit, 1000000 );
p.timeSat += Abc_Clock() - clk;

    if RetValue1 == MSAT_FALSE {
//p->time1 += Abc_Clock() - clk;

if fVerbose != 0 {
//    printf( "unsat  %d  ", Msat_SolverReadBackTracks(p->pSat) );
//ABC_PRT( "time", Abc_Clock() - clk );
}

        // add the clause
        Msat_IntVecClear( p.vProj );
        // inlining of macro MSAT_VAR2LIT
        // inlining of macro Fraig_IsComplement
        Msat_IntVecPush( p.vProj, (2*(pNode1R.Num)+(((int32(((ABC_PTRUINT_T) (pNode1) & 01))))))  );
        // inlining of macro MSAT_VAR2LIT
        // inlining of macro Fraig_IsComplement
        Msat_IntVecPush( p.vProj, (2*(pNode2R.Num)+(((int32(((ABC_PTRUINT_T) (pNode2) & 01))))))  );
        RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
        assert((RetValue != 0), "RetValue");
//        p->nSatProofImp++;
        return 1;
    
    } else if RetValue1 == MSAT_TRUE {
//p->time2 += Abc_Clock() - clk;

if fVerbose != 0 {
//    printf( "sat  %d  ", Msat_SolverReadBackTracks(p->pSat) );
//ABC_PRT( "time", Abc_Clock() - clk );
}
        // record the counter example
//        Fraig_FeedBack( p, Msat_SolverReadModelArray(p->pSat), p->vVarsInt, pNode1R, pNode2R );
        p.nSatCounterImp++;
        return 0;
    
    } else { // if ( RetValue1 == MSAT_UNKNOWN )
    
p.time3 += Abc_Clock() - clk;
        p.nSatFailsImp++;
        return 0;
    }
}


/**Function*************************************************************

  Synopsis    [Prepares the SAT solver to run on the two nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_PrepareCones(pMan *Fraig_Man_t, pOld *Fraig_Node_t, pNew *Fraig_Node_t) {
//    Msat_IntVec_t * vAdjs;
//    int * pVars, nVars, i, k;
    var nVarsAlloc int32;

    assert(pOld != pNew, "pOld != pNew");
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pOld) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pOld) & 01))))");
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNew) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNew) & 01))))");
    // clean the variables
    nVarsAlloc = Msat_IntVecReadSize(pMan.vVarsUsed);
    Msat_IntVecFill( pMan.vVarsUsed, nVarsAlloc, 0 );
    Msat_IntVecClear( pMan.vVarsInt );

    pMan.nTravIds++;
    Fraig_PrepareCones_rec( pMan, pNew );
    Fraig_PrepareCones_rec( pMan, pOld );


/*
    nVars = Msat_IntVecReadSize( pMan->vVarsInt );
    pVars = Msat_IntVecReadArray( pMan->vVarsInt );
    for ( i = 0; i < nVars; i++ )
    {
        // process its connections
        vAdjs = (Msat_IntVec_t *)Msat_ClauseVecReadEntry( pMan->vAdjacents, pVars[i] );
        printf( "%d=%d { ", pVars[i], Msat_IntVecReadSize(vAdjs) );
        for ( k = 0; k < Msat_IntVecReadSize(vAdjs); k++ )
            printf( "%d ", Msat_IntVecReadEntry(vAdjs,k) );
        printf( "}\n" );

    }
    i = 0;
*/
}

/**Function*************************************************************

  Synopsis    [Traverses the cone, collects the numbers and adds the clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_PrepareCones_rec(pMan *Fraig_Man_t, pNode *Fraig_Node_t) {
    var pFanin *Fraig_Node_t;
    var vAdjs *Msat_IntVec_t;
    var fUseMuxes int32 = 1;
    var i int32;
    var fItIsTime int32;

    // skip if the node is aleady visited
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    if pNode.TravId == pMan.nTravIds {
        return; }
    pNode.TravId = pMan.nTravIds;

    // collect the node's number (closer to reverse topological order)
    Msat_IntVecPush( pMan.vVarsInt, pNode.Num );
    Msat_IntVecWriteEntry( pMan.vVarsUsed, pNode.Num, 1 );
    if !Fraig_NodeIsAnd( pNode ) {
        return; }

    // if the node does not have fanins, create them
    fItIsTime = 0;
    if pNode.vFanins == nil {
        fItIsTime = 1;
        // create the fanins of the supergate
        assert(pNode.fClauses == 0, "pNode.fClauses == 0");
        if (fUseMuxes != 0) && Fraig_NodeIsMuxType(pNode) {
            pNode.vFanins = Fraig_NodeVecAlloc( 4 );
            // inlining of macro Fraig_Regular
            Fraig_NodeVecPushUnique( pNode.vFanins, ((Fraig_Node_t *)((ABC_PTRUINT_T)(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .p1) & ~01))  );
            // inlining of macro Fraig_Regular
            Fraig_NodeVecPushUnique( pNode.vFanins, ((Fraig_Node_t *)((ABC_PTRUINT_T)(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .p2) & ~01))  );
            // inlining of macro Fraig_Regular
            Fraig_NodeVecPushUnique( pNode.vFanins, ((Fraig_Node_t *)((ABC_PTRUINT_T)(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .p1) & ~01))  );
            // inlining of macro Fraig_Regular
            Fraig_NodeVecPushUnique( pNode.vFanins, ((Fraig_Node_t *)((ABC_PTRUINT_T)(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .p2) & ~01))  );
            Fraig_SupergateAddClausesMux( pMan, pNode );
        
        } else {
            pNode.vFanins = Fraig_CollectSupergate( pNode, fUseMuxes );
            Fraig_SupergateAddClauses( pMan, pNode, pNode.vFanins );
        }
        assert(pNode.vFanins.nSize > 1, "pNode.vFanins.nSize > 1");
        pNode.fClauses = 1;
        pMan.nVarsClauses++;

        // add fanins
        vAdjs = (Msat_IntVec_t *)Msat_ClauseVecReadEntry( pMan.vAdjacents, pNode.Num );
        assert(Msat_IntVecReadSize( vAdjs ) == 0, "Msat_IntVecReadSize( vAdjs ) == 0");
        for  i = 0; i < pNode.vFanins.nSize; i++  {
            // inlining of macro Fraig_Regular
            pFanin = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.vFanins.pArray[i]) & ~01)) ;
            Msat_IntVecPush( vAdjs, pFanin.Num );
        }
    }

    // recursively visit the fanins
    for  i = 0; i < pNode.vFanins.nSize; i++  {
        // inlining of macro Fraig_Regular
        Fraig_PrepareCones_rec( pMan, ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.vFanins.pArray[i]) & ~01))  );
    }

    if fItIsTime != 0 {
        // recursively visit the fanins
        for  i = 0; i < pNode.vFanins.nSize; i++  {
            // inlining of macro Fraig_Regular
            pFanin = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.vFanins.pArray[i]) & ~01)) ;
            vAdjs = (Msat_IntVec_t *)Msat_ClauseVecReadEntry( pMan.vAdjacents, pFanin.Num );
            Msat_IntVecPush( vAdjs, pNode.Num );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Collect variables using their proximity from the nodes.]

  Description [This procedure creates a variable order based on collecting
  first the nodes that are the closest to the given two target nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_OrderVariables(pMan *Fraig_Man_t, pOld *Fraig_Node_t, pNew *Fraig_Node_t) {
    var pNode, pFanin *Fraig_Node_t;
    var i int32;
    var k int32;
    var Number int32;
    var fUseMuxes int32 = 1;
    var nVarsAlloc int32;

    assert(pOld != pNew, "pOld != pNew");
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pOld) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pOld) & 01))))");
    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNew) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNew) & 01))))");

    pMan.nTravIds++;

    // clean the variables
    nVarsAlloc = Msat_IntVecReadSize(pMan.vVarsUsed);
    Msat_IntVecFill( pMan.vVarsUsed, nVarsAlloc, 0 );
    Msat_IntVecClear( pMan.vVarsInt );

    // add the first node
    Msat_IntVecPush( pMan.vVarsInt, pOld.Num );
    Msat_IntVecWriteEntry( pMan.vVarsUsed, pOld.Num, 1 );
    pOld.TravId = pMan.nTravIds;

    // add the second node
    Msat_IntVecPush( pMan.vVarsInt, pNew.Num );
    Msat_IntVecWriteEntry( pMan.vVarsUsed, pNew.Num, 1 );
    pNew.TravId = pMan.nTravIds;

    // create the variable order
    for  i = 0; i < Msat_IntVecReadSize(pMan.vVarsInt); i++  {
        // get the new node on the frontier
        Number = Msat_IntVecReadEntry(pMan.vVarsInt, i);
        pNode = pMan.vNodes.pArray[Number];
        if !Fraig_NodeIsAnd(pNode) {
            continue; }

        // if the node does not have fanins, create them
        if pNode.vFanins == nil {
            // create the fanins of the supergate
            assert(pNode.fClauses == 0, "pNode.fClauses == 0");
            // detecting a fanout-free cone (experiment only)
//            Fraig_DetectFanoutFreeCone( pMan, pNode );

            if (fUseMuxes != 0) && Fraig_NodeIsMuxType(pNode) {
                pNode.vFanins = Fraig_NodeVecAlloc( 4 );
                // inlining of macro Fraig_Regular
                Fraig_NodeVecPushUnique( pNode.vFanins, ((Fraig_Node_t *)((ABC_PTRUINT_T)(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .p1) & ~01))  );
                // inlining of macro Fraig_Regular
                Fraig_NodeVecPushUnique( pNode.vFanins, ((Fraig_Node_t *)((ABC_PTRUINT_T)(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .p2) & ~01))  );
                // inlining of macro Fraig_Regular
                Fraig_NodeVecPushUnique( pNode.vFanins, ((Fraig_Node_t *)((ABC_PTRUINT_T)(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .p1) & ~01))  );
                // inlining of macro Fraig_Regular
                Fraig_NodeVecPushUnique( pNode.vFanins, ((Fraig_Node_t *)((ABC_PTRUINT_T)(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .p2) & ~01))  );
                Fraig_SupergateAddClausesMux( pMan, pNode );
//                Fraig_DetectFanoutFreeConeMux( pMan, pNode );

                nMuxes++;
            
            } else {
                pNode.vFanins = Fraig_CollectSupergate( pNode, fUseMuxes );
                Fraig_SupergateAddClauses( pMan, pNode, pNode.vFanins );
            }
            assert(pNode.vFanins.nSize > 1, "pNode.vFanins.nSize > 1");
            pNode.fClauses = 1;
            pMan.nVarsClauses++;

            pNode.fMark2 = 1; // goes together with Fraig_SetupAdjacentMark()
        }

        // explore the implication fanins of pNode
        for  k = 0; k < pNode.vFanins.nSize; k++  {
            // inlining of macro Fraig_Regular
            pFanin = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.vFanins.pArray[k]) & ~01)) ;
            if pFanin.TravId == pMan.nTravIds { // already collected
                continue; }
            // collect and mark
            Msat_IntVecPush( pMan.vVarsInt, pFanin.Num );
            Msat_IntVecWriteEntry( pMan.vVarsUsed, pFanin.Num, 1 );
            pFanin.TravId = pMan.nTravIds;
        }
    }

    // set up the adjacent variable information
//    Fraig_SetupAdjacent( pMan, pMan->vVarsInt );
    Fraig_SetupAdjacentMark( pMan, pMan.vVarsInt );
}



/**Function*************************************************************

  Synopsis    [Set up the adjacent variable information.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_SetupAdjacent(pMan *Fraig_Man_t, vConeVars *Msat_IntVec_t) {
    var pNode, pFanin *Fraig_Node_t;
    var vAdjs *Msat_IntVec_t;
    var pVars *int32;
    var nVars int32;
    var i int32;
    var k int32;

    // clean the adjacents for the variables
    nVars = Msat_IntVecReadSize( vConeVars );
    pVars = Msat_IntVecReadArray( vConeVars );
    for  i = 0; i < nVars; i++  {
        // process its connections
        vAdjs = (Msat_IntVec_t *)Msat_ClauseVecReadEntry( pMan.vAdjacents, pVars[i] );
        Msat_IntVecClear( vAdjs );

        pNode = pMan.vNodes.pArray[pVars[i]];
        if !Fraig_NodeIsAnd(pNode) {
            continue; }

        // add fanins
        vAdjs = (Msat_IntVec_t *)Msat_ClauseVecReadEntry( pMan.vAdjacents, pVars[i] );
        for  k = 0; k < pNode.vFanins.nSize; k++  {
//        for ( k = pNode->vFanins->nSize - 1; k >= 0; k-- )
        
            // inlining of macro Fraig_Regular
            pFanin = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.vFanins.pArray[k]) & ~01)) ;
            Msat_IntVecPush( vAdjs, pFanin.Num );
//            Msat_IntVecPushUniqueOrder( vAdjs, pFanin->Num );
        }
    }
    // add the fanouts
    for  i = 0; i < nVars; i++  {
        pNode = pMan.vNodes.pArray[pVars[i]];
        if !Fraig_NodeIsAnd(pNode) {
            continue; }

        // add the edges
        for  k = 0; k < pNode.vFanins.nSize; k++  {
//        for ( k = pNode->vFanins->nSize - 1; k >= 0; k-- )
        
            // inlining of macro Fraig_Regular
            pFanin = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.vFanins.pArray[k]) & ~01)) ;
            vAdjs = (Msat_IntVec_t *)Msat_ClauseVecReadEntry( pMan.vAdjacents, pFanin.Num );
            Msat_IntVecPush( vAdjs, pNode.Num );
//            Msat_IntVecPushUniqueOrder( vAdjs, pFanin->Num );
        }
    }
}


/**Function*************************************************************

  Synopsis    [Set up the adjacent variable information.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_SetupAdjacentMark(pMan *Fraig_Man_t, vConeVars *Msat_IntVec_t) {
    var pNode, pFanin *Fraig_Node_t;
    var vAdjs *Msat_IntVec_t;
    var pVars *int32;
    var nVars int32;
    var i int32;
    var k int32;

    // clean the adjacents for the variables
    nVars = Msat_IntVecReadSize( vConeVars );
    pVars = Msat_IntVecReadArray( vConeVars );
    for  i = 0; i < nVars; i++  {
        pNode = pMan.vNodes.pArray[pVars[i]];
        if pNode.fMark2 == 0 {
            continue; }
//        pNode->fMark2 = 0;

        // process its connections
//        vAdjs = (Msat_IntVec_t *)Msat_ClauseVecReadEntry( pMan->vAdjacents, pVars[i] );
//        Msat_IntVecClear( vAdjs );

        if !Fraig_NodeIsAnd(pNode) {
            continue; }

        // add fanins
        vAdjs = (Msat_IntVec_t *)Msat_ClauseVecReadEntry( pMan.vAdjacents, pVars[i] );
        for  k = 0; k < pNode.vFanins.nSize; k++  {
//        for ( k = pNode->vFanins->nSize - 1; k >= 0; k-- )
        
            // inlining of macro Fraig_Regular
            pFanin = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.vFanins.pArray[k]) & ~01)) ;
            Msat_IntVecPush( vAdjs, pFanin.Num );
//            Msat_IntVecPushUniqueOrder( vAdjs, pFanin->Num );
        }
    }
    // add the fanouts
    for  i = 0; i < nVars; i++  {
        pNode = pMan.vNodes.pArray[pVars[i]];
        if pNode.fMark2 == 0 {
            continue; }
        pNode.fMark2 = 0;

        if !Fraig_NodeIsAnd(pNode) {
            continue; }

        // add the edges
        for  k = 0; k < pNode.vFanins.nSize; k++  {
//        for ( k = pNode->vFanins->nSize - 1; k >= 0; k-- )
        
            // inlining of macro Fraig_Regular
            pFanin = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.vFanins.pArray[k]) & ~01)) ;
            vAdjs = (Msat_IntVec_t *)Msat_ClauseVecReadEntry( pMan.vAdjacents, pFanin.Num );
            Msat_IntVecPush( vAdjs, pNode.Num );
//            Msat_IntVecPushUniqueOrder( vAdjs, pFanin->Num );
        }
    }
}




/**Function*************************************************************

  Synopsis    [Adds clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_SupergateAddClauses(p *Fraig_Man_t, pNode *Fraig_Node_t, vSuper *Fraig_NodeVec_t) {
    var fComp1, RetValue, nVars, Var, Var1, i int32;

    assert(Fraig_NodeIsAnd( pNode ), "Fraig_NodeIsAnd( pNode )");
    nVars = Msat_SolverReadVarNum(p.pSat);

    Var = pNode.Num;
    assert(Var  < nVars, "Var  < nVars"); 
    for  i = 0; i < vSuper.nSize; i++  {
        // get the predecessor nodes
        // get the complemented attributes of the nodes
        // inlining of macro Fraig_IsComplement
        fComp1 = ((int32(((ABC_PTRUINT_T) (vSuper.pArray[i]) & 01))));
        // determine the variable numbers
        // inlining of macro Fraig_Regular
        Var1 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(vSuper.pArray[i]) & ~01)) .Num;
        // check that the variables are in the SAT manager
        assert(Var1 < nVars, "Var1 < nVars");

        // suppose the AND-gate is A * B = C
        // add !A => !C   or   A + !C
    //  fprintf( pFile, "%d %d 0%c", Var1, -Var, 10 );
        Msat_IntVecClear( p.vProj );
        // inlining of macro MSAT_VAR2LIT
        Msat_IntVecPush( p.vProj, (2*(Var1)+(fComp1))  );
        // inlining of macro MSAT_VAR2LIT
        Msat_IntVecPush( p.vProj, (2*(Var)+(1))  );
        RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
        assert((RetValue != 0), "RetValue");
    }

    // add A & B => C   or   !A + !B + C
//  fprintf( pFile, "%d %d %d 0%c", -Var1, -Var2, Var, 10 );
    Msat_IntVecClear( p.vProj );
    for  i = 0; i < vSuper.nSize; i++  {
        // get the predecessor nodes
        // get the complemented attributes of the nodes
        // inlining of macro Fraig_IsComplement
        fComp1 = ((int32(((ABC_PTRUINT_T) (vSuper.pArray[i]) & 01))));
        // determine the variable numbers
        // inlining of macro Fraig_Regular
        Var1 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(vSuper.pArray[i]) & ~01)) .Num;

        // add this variable to the array
        // inlining of macro MSAT_VAR2LIT
        Msat_IntVecPush( p.vProj, (2*(Var1)+(!fComp1))  );
    }
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(Var)+(0))  );
    RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
    assert((RetValue != 0), "RetValue");
}

/**Function*************************************************************

  Synopsis    [Adds clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_SupergateAddClausesExor(p *Fraig_Man_t, pNode *Fraig_Node_t) {
    var pNode1, pNode2 *Fraig_Node_t;
    var fComp, RetValue int32;

    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    assert(Fraig_NodeIsExorType( pNode ), "Fraig_NodeIsExorType( pNode )");
    // get nodes
    // inlining of macro Fraig_Regular
    pNode1 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .p1) & ~01)) ;
    // inlining of macro Fraig_Regular
    pNode2 = ((Fraig_Node_t *)((ABC_PTRUINT_T)(((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .p2) & ~01)) ;
    // get the complemented attribute of the EXOR/NEXOR gate
    fComp = Fraig_NodeIsExor( pNode ); // 1 if EXOR, 0 if NEXOR

    // create four clauses
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode.Num)+(fComp))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode1.Num)+(fComp))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode2.Num)+(fComp))  );
    RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
    assert((RetValue != 0), "RetValue");
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode.Num)+(fComp))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode1.Num)+(!fComp))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode2.Num)+(!fComp))  );
    RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
    assert((RetValue != 0), "RetValue");
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode.Num)+(!fComp))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode1.Num)+(fComp))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode2.Num)+(!fComp))  );
    RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
    assert((RetValue != 0), "RetValue");
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode.Num)+(!fComp))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode1.Num)+(!fComp))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(pNode2.Num)+(fComp))  );
    RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
    assert((RetValue != 0), "RetValue");
}

/**Function*************************************************************

  Synopsis    [Adds clauses to the solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_SupergateAddClausesMux(p *Fraig_Man_t, pNode *Fraig_Node_t) {
    var pNodeI, pNodeT, pNodeE *Fraig_Node_t;
    var RetValue, VarF, VarI, VarT, VarE, fCompT, fCompE int32;

    // inlining of macro Fraig_IsComplement
    assert(!(((int32(((ABC_PTRUINT_T) (pNode) & 01)) != 0))), "!/*__starting_point_inline_macro:Fraig_IsComplement__*/((int32(((ABC_PTRUINT_T) (pNode) & 01))))");
    assert(Fraig_NodeIsMuxType( pNode ), "Fraig_NodeIsMuxType( pNode )");
    // get nodes (I = if, T = then, E = else)
    pNodeI = Fraig_NodeRecognizeMux( pNode, &pNodeT, &pNodeE );
    // get the variable numbers
    VarF = pNode.Num;
    VarI = pNodeI.Num;
    // inlining of macro Fraig_Regular
    VarT = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNodeT) & ~01)) .Num;
    // inlining of macro Fraig_Regular
    VarE = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNodeE) & ~01)) .Num;
    // get the complementation flags
    // inlining of macro Fraig_IsComplement
    fCompT = ((int32(((ABC_PTRUINT_T) (pNodeT) & 01))));
    // inlining of macro Fraig_IsComplement
    fCompE = ((int32(((ABC_PTRUINT_T) (pNodeE) & 01))));

    // f = ITE(i, t, e)

    // i' + t' + f
    // i' + t  + f'
    // i  + e' + f
    // i  + e  + f'

    // create four clauses
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarI)+(1))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarT)+(1^fCompT))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarF)+(0))  );
    RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
    assert((RetValue != 0), "RetValue");
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarI)+(1))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarT)+(0^fCompT))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarF)+(1))  );
    RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
    assert((RetValue != 0), "RetValue");
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarI)+(0))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarE)+(1^fCompE))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarF)+(0))  );
    RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
    assert((RetValue != 0), "RetValue");
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarI)+(0))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarE)+(0^fCompE))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarF)+(1))  );
    RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
    assert((RetValue != 0), "RetValue");

    // two additional clauses
    // t' & e' -> f'
    // t  & e  -> f 

    // t  + e   + f'
    // t' + e'  + f 

    if VarT == VarE {
//        assert( fCompT == !fCompE );
        return;
    }

    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarT)+(0^fCompT))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarE)+(0^fCompE))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarF)+(1))  );
    RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
    assert((RetValue != 0), "RetValue");
    Msat_IntVecClear( p.vProj );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarT)+(1^fCompT))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarE)+(1^fCompE))  );
    // inlining of macro MSAT_VAR2LIT
    Msat_IntVecPush( p.vProj, (2*(VarF)+(0))  );
    RetValue = Msat_SolverAddClause( p.pSat, p.vProj );
    assert((RetValue != 0), "RetValue");

}





/**Function*************************************************************

  Synopsis    [Returns the array of nodes to be combined into one multi-input AND-gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_DetectFanoutFreeCone_rec(pNode *Fraig_Node_t, vSuper *Fraig_NodeVec_t, vInside *Fraig_NodeVec_t, fFirst int32) {
    // make the pointer regular
    // inlining of macro Fraig_Regular
    pNode = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) ;
    // if the new node is complemented or a PI, another gate begins
    if (!(fFirst != 0) && pNode.nRefs > 1) || Fraig_NodeIsVar(pNode) {
        Fraig_NodeVecPushUnique( vSuper, pNode );
        return;
    }
    // go through the branches
    Fraig_DetectFanoutFreeCone_rec( pNode.p1, vSuper, vInside, 0 );
    Fraig_DetectFanoutFreeCone_rec( pNode.p2, vSuper, vInside, 0 );
    // add the node
    Fraig_NodeVecPushUnique( vInside, pNode );
}

/**Function*************************************************************

  Synopsis    [Returns the array of nodes to be combined into one multi-input AND-gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
void Fraig_DetectFanoutFreeCone( Fraig_Man_t * pMan, Fraig_Node_t * pNode )
{
    Fraig_NodeVec_t * vFanins;
    Fraig_NodeVec_t * vInside;
    int nCubes;
    extern int Fraig_CutSopCountCubes( Fraig_Man_t * pMan, Fraig_NodeVec_t * vFanins, Fraig_NodeVec_t * vInside );

    vFanins = Fraig_NodeVecAlloc( 8 );
    vInside = Fraig_NodeVecAlloc( 8 );

    Fraig_DetectFanoutFreeCone_rec( pNode, vFanins, vInside, 1 );
    assert( vInside->pArray[vInside->nSize-1] == pNode );

    nCubes = Fraig_CutSopCountCubes( pMan, vFanins, vInside );

printf( "%d(%d)", vFanins->nSize, nCubes );
    Fraig_NodeVecFree( vFanins );
    Fraig_NodeVecFree( vInside );
}
*/



/**Function*************************************************************

  Synopsis    [Returns the array of nodes to be combined into one multi-input AND-gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_DetectFanoutFreeConeMux_rec(pNode *Fraig_Node_t, vSuper *Fraig_NodeVec_t, vInside *Fraig_NodeVec_t, fFirst int32) {
    // make the pointer regular
    // inlining of macro Fraig_Regular
    pNode = ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode) & ~01)) ;
    // if the new node is complemented or a PI, another gate begins
    if (!(fFirst != 0) && pNode.nRefs > 1) || Fraig_NodeIsVar(pNode) || !Fraig_NodeIsMuxType(pNode) {
        Fraig_NodeVecPushUnique( vSuper, pNode );
        return;
    }
    // go through the branches
    // inlining of macro Fraig_Regular
    Fraig_DetectFanoutFreeConeMux_rec( ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .p1, vSuper, vInside, 0 );
    // inlining of macro Fraig_Regular
    Fraig_DetectFanoutFreeConeMux_rec( ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .p2, vSuper, vInside, 0 );
    // inlining of macro Fraig_Regular
    Fraig_DetectFanoutFreeConeMux_rec( ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .p1, vSuper, vInside, 0 );
    // inlining of macro Fraig_Regular
    Fraig_DetectFanoutFreeConeMux_rec( ((Fraig_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .p2, vSuper, vInside, 0 );
    // add the node
    Fraig_NodeVecPushUnique( vInside, pNode );
}

/**Function*************************************************************

  Synopsis    [Returns the array of nodes to be combined into one multi-input AND-gate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_DetectFanoutFreeConeMux(pMan *Fraig_Man_t, pNode *Fraig_Node_t) {
    var vFanins *Fraig_NodeVec_t;
    var vInside *Fraig_NodeVec_t;
    var nCubes int32;
    extern int Fraig_CutSopCountCubes( Fraig_Man_t * pMan, Fraig_NodeVec_t * vFanins, Fraig_NodeVec_t * vInside );

    vFanins = Fraig_NodeVecAlloc( 8 );
    vInside = Fraig_NodeVecAlloc( 8 );

    Fraig_DetectFanoutFreeConeMux_rec( pNode, vFanins, vInside, 1 );
    assert(vInside.pArray[vInside.nSize-1] == pNode, "vInside.pArray[vInside.nSize-1] == pNode");

//    nCubes = Fraig_CutSopCountCubes( pMan, vFanins, vInside );
    nCubes = 0;

fmt.Printf("%d(%d)", vFanins.nSize, nCubes);
    Fraig_NodeVecFree( vFanins );
    Fraig_NodeVecFree( vInside );
}



/**Function*************************************************************

  Synopsis    [Collect variables using their proximity from the nodes.]

  Description [This procedure creates a variable order based on collecting
  first the nodes that are the closest to the given two target nodes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fraig_SetActivity(pMan *Fraig_Man_t, pOld *Fraig_Node_t, pNew *Fraig_Node_t) {
    var pNode *Fraig_Node_t;
    var i, Number, MaxLevel int32;
    var pFactors *float32  = Msat_SolverReadFactors(pMan.pSat);
    if pFactors == nil {
        return; }
    MaxLevel = Abc_MaxInt( pOld.Level, pNew.Level );
    // create the variable order
    for  i = 0; i < Msat_IntVecReadSize(pMan.vVarsInt); i++  {
        // get the new node on the frontier
        Number = Msat_IntVecReadEntry(pMan.vVarsInt, i);
        pNode = pMan.vNodes.pArray[Number];
        pFactors[pNode.Num] = float32(math.Pow(float64(0.97), float64(MaxLevel - pNode.Level)));
//        if ( pNode->Num % 50 == 0 )
//        printf( "(%d) %.2f  ", MaxLevel - pNode->Level, pFactors[pNode->Num] );
    }
//    printf( "\n" );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


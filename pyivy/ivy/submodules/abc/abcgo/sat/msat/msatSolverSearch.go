package msat

import "fmt"

/**CFile****************************************************************

  FileName    [msatSolverSearch.c]

  PackageName [A C version of SAT solver MINISAT, originally developed 
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of 
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [The search part of the solver.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatSolverSearch.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

//#include "msatInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void            Msat_SolverUndoOne( Msat_Solver_t * p );

// static void            Msat_SolverCancel( Msat_Solver_t * p );

// static Msat_Clause_t * Msat_SolverRecord( Msat_Solver_t * p, Msat_IntVec_t * vLits );

// static void            Msat_SolverAnalyze( Msat_Solver_t * p, Msat_Clause_t * pC, Msat_IntVec_t * vLits_out, int * pLevel_out );

// static void            Msat_SolverReduceDB( Msat_Solver_t * p );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Makes the next assumption (Lit).]

  Description [Returns FALSE if immediate conflict.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverAssume(p *Msat_Solver_t, Lit Msat_Lit_t) int32 {
    assert(Msat_QueueReadSize(p.pQueue) == 0, "Msat_QueueReadSize(p.pQueue) == 0");
    if p.fVerbose {
        // inlining of macro L_lit
        // inlining of macro MSAT_LITSIGN
        // inlining of macro MSAT_LIT2VAR
        fmt.Printf(L_IND"assume("L_LIT")\n", L_ind, ((Lit)&1)     ?"-":"", ((Lit)>>1)+1); }
    Msat_IntVecPush( p.vTrailLim, Msat_IntVecReadSize(p.vTrail) );
//    assert( Msat_IntVecReadSize(p->vTrailLim) <= Msat_IntVecReadSize(p->vTrail) + 1 );
//    assert( Msat_IntVecReadSize( p->vTrailLim ) < p->nVars );
    return Msat_SolverEnqueue( p, Lit, nil );
}

/**Function*************************************************************

  Synopsis    [Reverts one variable binding on the trail.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverUndoOne(p *Msat_Solver_t) {
    var Lit Msat_Lit_t;
    var Var Msat_Var_t;
    Lit = Msat_IntVecPop( p.vTrail ); 
    // inlining of macro MSAT_LIT2VAR
    Var = ((Lit)>>1);
    p.pAssigns[Var] = MSAT_VAR_UNASSIGNED;
    p.pReasons[Var] = nil;
    p.pLevel[Var]   = -1;
//    Msat_OrderUndo( p->pOrder, Var );
    Msat_OrderVarUnassigned( p.pOrder, Var );

    if p.fVerbose {
        // inlining of macro L_lit
        // inlining of macro MSAT_LITSIGN
        // inlining of macro MSAT_LIT2VAR
        fmt.Printf(L_IND"unbind("L_LIT")\n", L_ind, ((Lit)&1)     ?"-":"", ((Lit)>>1)+1); } 
}

/**Function*************************************************************

  Synopsis    [Reverts to the state before last Msat_SolverAssume().]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverCancel(p *Msat_Solver_t) {
    var c int32;
    assert(Msat_QueueReadSize(p.pQueue) == 0, "Msat_QueueReadSize(p.pQueue) == 0");
    if p.fVerbose {
        if Msat_IntVecReadSize(p.vTrail) != Msat_IntVecReadEntryLast(p.vTrailLim) {
            var Lit Msat_Lit_t;
            Lit = Msat_IntVecReadEntry( p.vTrail, Msat_IntVecReadEntryLast(p.vTrailLim) ); 
            // inlining of macro L_lit
            // inlining of macro MSAT_LITSIGN
            // inlining of macro MSAT_LIT2VAR
            fmt.Printf(L_IND"cancel("L_LIT")\n", L_ind, ((Lit)&1)     ?"-":"", ((Lit)>>1)+1);
        }
    }
    for  c = Msat_IntVecReadSize(p.vTrail) - Msat_IntVecPop( p.vTrailLim ); c != 0; c--  {
        Msat_SolverUndoOne( p );
    }
}

/**Function*************************************************************

  Synopsis    [Reverts to the state at given level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverCancelUntil(p *Msat_Solver_t, Level int32) {
    for  Msat_IntVecReadSize(p.vTrailLim) > Level  {
        Msat_SolverCancel(p);
    }
}


/**Function*************************************************************

  Synopsis    [Record a clause and drive backtracking.]

  Description [vLits[0] must contain the asserting literal.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverRecord(p *Msat_Solver_t, vLits *Msat_IntVec_t) *Msat_Clause_t {
    var pC *Msat_Clause_t;
    var Value int32;
    assert(Msat_IntVecReadSize(vLits) != 0, "Msat_IntVecReadSize(vLits) != 0");
    Value = Msat_ClauseCreate( p, vLits, 1, &pC );
    assert((Value != 0), "Value");
    Value = Msat_SolverEnqueue( p, Msat_IntVecReadEntry(vLits,0), pC );
    assert((Value != 0), "Value");
    if pC != nil {
        Msat_ClauseVecPush( p.vLearned, pC ); }
    return pC;
}

/**Function*************************************************************

  Synopsis    [Enqueues one variable assignment.]

  Description [Puts a new fact on the propagation queue and immediately 
  updates the variable value. Should a conflict arise, FALSE is returned.
  Otherwise returns TRUE.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverEnqueue(p *Msat_Solver_t, Lit Msat_Lit_t, pC *Msat_Clause_t) int32 {
    // inlining of macro MSAT_LIT2VAR
    var Var Msat_Var_t  = ((Lit)>>1);

    // skip literals that are not in the current cone
    if !Msat_IntVecReadEntry( p.vVarsUsed, Var ) {
        return 1; }

//    assert( Msat_QueueReadSize(p->pQueue) == Msat_IntVecReadSize(p->vTrail) );
    // if the literal is assigned
    // return 1 if the assignment is consistent
    // return 0 if the assignment is inconsistent (conflict)
    if p.pAssigns[Var] != MSAT_VAR_UNASSIGNED {
        return p.pAssigns[Var] == Lit; }
    // new fact - store it
    if p.fVerbose {
//        printf(L_IND"bind("L_LIT")\n", L_ind, L_lit(Lit));
        // inlining of macro L_lit
        // inlining of macro MSAT_LITSIGN
        // inlining of macro MSAT_LIT2VAR
        fmt.Printf(L_IND"bind("L_LIT")  ", L_ind, ((Lit)&1)     ?"-":"", ((Lit)>>1)+1);
        Msat_ClausePrintSymbols( pC );
    }
    p.pAssigns[Var] = Lit;
    p.pLevel[Var]   = Msat_IntVecReadSize(p.vTrailLim);
//    p->pReasons[Var] = p->pLevel[Var]? pC: NULL;
    p.pReasons[Var] = pC;
    Msat_IntVecPush( p.vTrail, Lit );
    Msat_QueueInsert( p.pQueue, Lit );

    Msat_OrderVarAssigned( p.pOrder, Var );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Propagates the assignments in the queue.]

  Description [Propagates all enqueued facts. If a conflict arises, 
  the conflicting clause is returned, otherwise NULL.]
               
  SideEffects [The propagation queue is empty, even if there was a conflict.]

  SeeAlso     []

***********************************************************************/
func Msat_SolverPropagate(p *Msat_Solver_t) *Msat_Clause_t {
    var pvWatched **Msat_ClauseVec_t  = p.pvWatched;
    var pClauses **Msat_Clause_t;
    var pConflict *Msat_Clause_t;
    var Lit, Lit_out Msat_Lit_t;
    var i, j, nClauses int32;

    // propagate all the literals in the queue
    for  (Lit = Msat_QueueExtract( p.pQueue )) >= 0  {
        p.Stats.nPropagations++;
        // get the clauses watched by this literal
        nClauses = Msat_ClauseVecReadSize( pvWatched[Lit] );
        pClauses = Msat_ClauseVecReadArray( pvWatched[Lit] );
        // go through the watched clauses and decide what to do with them
        for  i = j = 0; i < nClauses; i++  {
            p.Stats.nInspects++;
            // clear the returned literal
            Lit_out = -1;
            // propagate the clause
            if !Msat_ClausePropagate( pClauses[i], Lit, p.pAssigns, &Lit_out ) {   // the clause is unit
                // "Lit_out" contains the new assignment to be enqueued
                if Msat_SolverEnqueue( p, Lit_out, pClauses[i] ) != 0 { // consistent assignment 
                    // no changes to the implication queue; the watch is the same too
                    pClauses[j++] = pClauses[i];
                    continue;
                }
                // remember the reason of conflict (will be returned)
                pConflict = pClauses[i];
                // leave the remaning clauses in the same watched list
                for  ; i < nClauses; i++  {
                    pClauses[j++] = pClauses[i];
                }
                Msat_ClauseVecShrink( pvWatched[Lit], j );
                // clear the propagation queue
                Msat_QueueClear( p.pQueue );
                return pConflict;
            }
            // the clause is not unit
            // in this case "Lit_out" contains the new watch if it has changed
            if Lit_out >= 0 {
                Msat_ClauseVecPush( pvWatched[Lit_out], pClauses[i] );
            } else { // the watch did not change
                pClauses[j++] = pClauses[i]; }
        }
        Msat_ClauseVecShrink( pvWatched[Lit], j );
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Simplifies the data base.]

  Description [Simplify all constraints according to the current top-level 
  assigment (redundant constraints may be removed altogether).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverSimplifyDB(p *Msat_Solver_t) int32 {
    var vClauses *Msat_ClauseVec_t;
    var pClauses **Msat_Clause_t;
    var nClauses, Type, i, j int32;
    var pAssigns *int32;
    var Counter int32;

    assert(Msat_SolverReadDecisionLevel(p) == 0, "Msat_SolverReadDecisionLevel(p) == 0");
    if Msat_SolverPropagate(p) != nil {
        return 0; }
//Msat_SolverPrintClauses( p );
//Msat_SolverPrintAssignment( p );
//printf( "Simplification\n" );

    // simplify and reassign clause numbers
    Counter = 0;
    pAssigns = Msat_SolverReadAssignsArray( p );
    for  Type = 0; Type < 2; Type++  {
        if Type {
            vClauses = p.vLearned;
        } else {
            vClauses = p.vClauses;
        }
        nClauses = Msat_ClauseVecReadSize( vClauses );
        pClauses = Msat_ClauseVecReadArray( vClauses );
        for  i = j = 0; i < nClauses; i++  {
            if Msat_ClauseSimplify( pClauses[i], pAssigns ) {
                Msat_ClauseFree( p, pClauses[i], 1 );
            } else {
                pClauses[j++] = pClauses[i];
                Msat_ClauseSetNum( pClauses[i], Counter++ );
            }
        }
        Msat_ClauseVecShrink( vClauses, j );
    }
    p.nClauses = Counter;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Cleans the clause databased from the useless learnt clauses.]

  Description [Removes half of the learnt clauses, minus the clauses locked 
  by the current assignment. Locked clauses are clauses that are reason 
  to a some assignment.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverReduceDB(p *Msat_Solver_t) {
    var pLearned **Msat_Clause_t;
    var nLearned, i, j int32;
    var dExtraLim float64  = p.dClaInc / Msat_ClauseVecReadSize(p.vLearned); 
    // Remove any clause below this activity

    // sort the learned clauses in the increasing order of activity
    Msat_SolverSortDB( p );

    // discard the first half the clauses (the less active ones)
    nLearned = Msat_ClauseVecReadSize( p.vLearned );
    pLearned = Msat_ClauseVecReadArray( p.vLearned );
    for  i = j = 0; i < nLearned / 2; i++  {
        if !Msat_ClauseIsLocked( p, pLearned[i]) {
            Msat_ClauseFree( p, pLearned[i], 1 );
        } else {
            pLearned[j++] = pLearned[i]; }
    }
    // filter the more active clauses and leave those above the limit
    for   ; i < nLearned; i++  {
        if !Msat_ClauseIsLocked( p, pLearned[i] ) && 
            Msat_ClauseReadActivity(pLearned[i]) < dExtraLim {
            Msat_ClauseFree( p, pLearned[i], 1 );
        } else {
            pLearned[j++] = pLearned[i]; }
    }
    Msat_ClauseVecShrink( p.vLearned, j );
}

/**Function*************************************************************

  Synopsis    [Removes the learned clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverRemoveLearned(p *Msat_Solver_t) {
    var pLearned **Msat_Clause_t;
    var nLearned, i int32;

    // discard the learned clauses
    nLearned = Msat_ClauseVecReadSize( p.vLearned );
    pLearned = Msat_ClauseVecReadArray( p.vLearned );
    for  i = 0; i < nLearned; i++  {
        assert(!Msat_ClauseIsLocked( p, pLearned[i]), "!Msat_ClauseIsLocked( p, pLearned[i])");
        
        Msat_ClauseFree( p, pLearned[i], 1 );
    }
    Msat_ClauseVecShrink( p.vLearned, 0 );
    p.nClauses = Msat_ClauseVecReadSize(p.vClauses);

    for  i = 0; i < p.nVarsAlloc; i++  {
        p.pReasons[i] = nil;
    }
}

/**Function*************************************************************

  Synopsis    [Removes the recently added clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverRemoveMarked(p *Msat_Solver_t) {
    var pLearned, pClauses **Msat_Clause_t;
    var nLearned, nClauses, i int32;

    // discard the learned clauses
    nClauses = Msat_ClauseVecReadSize( p.vClauses );
    pClauses = Msat_ClauseVecReadArray( p.vClauses );
    for  i = p.nClausesStart; i < nClauses; i++  {
//        assert( !Msat_ClauseIsLocked( p, pClauses[i]) );
        Msat_ClauseFree( p, pClauses[i], 1 );
    }
    Msat_ClauseVecShrink( p.vClauses, p.nClausesStart );

    // discard the learned clauses
    nLearned = Msat_ClauseVecReadSize( p.vLearned );
    pLearned = Msat_ClauseVecReadArray( p.vLearned );
    for  i = 0; i < nLearned; i++  {
//        assert( !Msat_ClauseIsLocked( p, pLearned[i]) );
        Msat_ClauseFree( p, pLearned[i], 1 );
    }
    Msat_ClauseVecShrink( p.vLearned, 0 );
    p.nClauses = Msat_ClauseVecReadSize(p.vClauses);
/*
    // undo the previous data
    for ( i = 0; i < p->nVarsAlloc; i++ )
    {
        p->pAssigns[i]   = MSAT_VAR_UNASSIGNED;
        p->pReasons[i]   = NULL;
        p->pLevel[i]     = -1;
        p->pdActivity[i] = 0.0;
    }
    Msat_OrderClean( p->pOrder, p->nVars, NULL );
    Msat_QueueClear( p->pQueue );
*/
}



/**Function*************************************************************

  Synopsis    [Analyze conflict and produce a reason clause.]

  Description [Current decision level must be greater than root level.]
               
  SideEffects [vLits_out[0] is the asserting literal at level pLevel_out.]

  SeeAlso     []

***********************************************************************/
func Msat_SolverAnalyze(p *Msat_Solver_t, pC *Msat_Clause_t, vLits_out *Msat_IntVec_t, pLevel_out *int32) {
    var LitQ Msat_Lit_t;
    var Lit Msat_Lit_t = MSAT_LIT_UNASSIGNED;
    var VarQ, Var Msat_Var_t;
    var pReasonArray *int32;
    var nReasonSize int32;
    var j int32;
    var pathC int32;
    var nLevelCur int32 = Msat_IntVecReadSize(p.vTrailLim);
    var iStep int32  = Msat_IntVecReadSize(p.vTrail) - 1;

    // increment the seen counter
    p.nSeenId++;
    // empty the vector array
    Msat_IntVecClear( vLits_out );
    Msat_IntVecPush( vLits_out, -1 ); // (leave room for the asserting literal)
    *pLevel_out = 0;
    for {
        assert(pC != nil, "pC != nil");  // (otherwise should be UIP)
        // get the reason of conflict
        Msat_ClauseCalcReason( p, pC, Lit, p.vReason );
        nReasonSize  = Msat_IntVecReadSize( p.vReason );
        pReasonArray = Msat_IntVecReadArray( p.vReason );
        for  j = 0; j < nReasonSize; j++  {
            LitQ = pReasonArray[j];
            // inlining of macro MSAT_LIT2VAR
            VarQ = ((LitQ)>>1);
            if p.pSeen[VarQ] != p.nSeenId {
                p.pSeen[VarQ] = p.nSeenId;

                // added to better fine-tune the search
                Msat_SolverVarBumpActivity( p, LitQ );

                // skip all the literals on this decision level
                if p.pLevel[VarQ] == nLevelCur {
                    pathC++;
                } else if p.pLevel[VarQ] > 0 { 
                    // add the literals on other decision levels but
                    // exclude variables from decision level 0
                    // inlining of macro MSAT_LITNOT
                    Msat_IntVecPush( vLits_out, ((LitQ)^1) );
                    if *pLevel_out < p.pLevel[VarQ] {
                        *pLevel_out = p.pLevel[VarQ]; }
                }
            }
        }
        // Select next clause to look at:
        for {
//            Lit = Msat_IntVecReadEntryLast(p->vTrail);
            Lit = Msat_IntVecReadEntry( p.vTrail, iStep-- );
            // inlining of macro MSAT_LIT2VAR
            Var = ((Lit)>>1);
            pC = p.pReasons[Var];
//            Msat_SolverUndoOne( p );
        
if !( p.pSeen[Var] != p.nSeenId ) { break }
} 
        pathC--;
    
if !( pathC > 0 ) { break }
} 
    // we do not unbind the variables above
    // this will be done after conflict analysis

    // inlining of macro MSAT_LITNOT
    Msat_IntVecWriteEntry( vLits_out, 0, ((Lit)^1) );
    if p.fVerbose {
        fmt.Printf(L_IND"Learnt {", L_ind);
        nReasonSize  = Msat_IntVecReadSize( vLits_out );
        pReasonArray = Msat_IntVecReadArray( vLits_out );
        for  j = 0; j < nReasonSize; j++  { 
            // inlining of macro L_lit
            // inlining of macro MSAT_LITSIGN
            // inlining of macro MSAT_LIT2VAR
            fmt.Printf(" "L_LIT, ((pReasonArray[j])&1)     ?"-":"", ((pReasonArray[j])>>1)+1);
        }
        fmt.Printf(" } at level %d\n", *pLevel_out); 
    }
}

/**Function*************************************************************

  Synopsis    [The search procedure called between the restarts.]

  Description [Search for a satisfying solution as long as the number of 
  conflicts does not exceed the limit (nConfLimit) while keeping the number 
  of learnt clauses below the provided limit (nLearnedLimit). NOTE! Use 
  negative value for nConfLimit or nLearnedLimit to indicate infinity.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_SolverSearch(p *Msat_Solver_t, nConfLimit int32, nLearnedLimit int32, nBackTrackLimit int32, pPars *Msat_SearchParams_t) Msat_Type_t {
    var pConf *Msat_Clause_t;
    var Var Msat_Var_t;
    var nLevelBack, nConfs, nAssigns, Value int32;
    var i int32;

    assert(Msat_SolverReadDecisionLevel(p) == p.nLevelRoot, "Msat_SolverReadDecisionLevel(p) == p.nLevelRoot");
    p.Stats.nStarts++;
    p.dVarDecay = 1 / pPars.dVarDecay;
    p.dClaDecay = 1 / pPars.dClaDecay;

    // reset the activities
    for  i = 0; i < p.nVars; i++  {
       p.pdActivity[i] = float64(p.pFactors[i]);
    }
//       p->pdActivity[i] = 0.0;

    nConfs = 0;
    for {
        pConf = Msat_SolverPropagate( p );
        if pConf != nil{
            // CONFLICT
            if p.fVerbose {
//                printf(L_IND"**CONFLICT**\n", L_ind);
                fmt.Printf(L_IND"**CONFLICT**  ", L_ind);
                Msat_ClausePrintSymbols( pConf );
            }
            // count conflicts
            p.Stats.nConflicts++;
            nConfs++;

            // if top level, return UNSAT
            if Msat_SolverReadDecisionLevel(p) == p.nLevelRoot {
                return MSAT_FALSE; }

            // perform conflict analysis
            Msat_SolverAnalyze( p, pConf, p.vTemp, &nLevelBack );
            Msat_SolverCancelUntil( p, (p.nLevelRoot > nLevelBack)? p.nLevelRoot : nLevelBack );
            Msat_SolverRecord( p, p.vTemp );

            // it is important that recording is done after cancelling
            // because canceling cleans the queue while recording adds to it
            Msat_SolverVarDecayActivity( p );
            Msat_SolverClaDecayActivity( p );

        
        } else{
            // NO CONFLICT
            if Msat_IntVecReadSize(p.vTrailLim) == 0 {
                // Simplify the set of problem clauses:
//                Value = Msat_SolverSimplifyDB(p);
//                assert( Value );
            }
            nAssigns = Msat_IntVecReadSize( p.vTrail );
            if nLearnedLimit >= 0 && Msat_ClauseVecReadSize(p.vLearned) >= nLearnedLimit + nAssigns {
                // Reduce the set of learnt clauses:
                Msat_SolverReduceDB(p);
            }

            Var = Msat_OrderVarSelect( p.pOrder );
            if Var == MSAT_ORDER_UNKNOWN {
                // Model found and stored in p->pAssigns
                memcpy( p.pModel, p.pAssigns, sizeof(int) * p.nVars );
                Msat_QueueClear( p.pQueue );
                Msat_SolverCancelUntil( p, p.nLevelRoot );
                return MSAT_TRUE;
            }
            if nConfLimit > 0 && nConfs > nConfLimit {
                // Reached bound on number of conflicts:
                p.dProgress = Msat_SolverProgressEstimate( p );
                Msat_QueueClear( p.pQueue );
                Msat_SolverCancelUntil( p, p.nLevelRoot );
                return MSAT_UNKNOWN; 
            
            } else if nBackTrackLimit > 0 && int32(p.Stats.nConflicts) - p.nBackTracks > nBackTrackLimit {
                // Reached bound on number of conflicts:
                Msat_QueueClear( p.pQueue );
                Msat_SolverCancelUntil( p, p.nLevelRoot );
                return MSAT_UNKNOWN; 
            
            } else{
                // New variable decision:
                p.Stats.nDecisions++;
                assert(Var != MSAT_ORDER_UNKNOWN && Var >= 0 && Var < p.nVars, "Var != MSAT_ORDER_UNKNOWN && Var >= 0 && Var < p.nVars");
                // inlining of macro MSAT_VAR2LIT
                Value = Msat_SolverAssume(p, (2*(Var)+(0))  );
                assert((Value != 0), "Value");
            }
        }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


package bsat

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [satInter.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sat_solver.]

  Synopsis    [Interpolation package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satInter.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "satStore.h"
//#include "aig/aig/aig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// variable assignments 
var LIT_UNDEF lit = 0xffffffff;

// interpolation manager
type Inta_Man_t_ struct {
    // clauses of the problems
    pCnf      *Sto_Man_t;         // the set of CNF clauses for A and B
    vVarsAB      *Vec_Int_t;      // the array of global variables
    // various parameters
    fVerbose             int32;     // verbosiness flag
    fProofVerif             int32;  // verifies the proof
    fProofWrite             int32;  // writes the proof file
    nVarsAlloc             int32;   // the allocated size of var arrays
    nClosAlloc             int32;   // the allocated size of clause arrays
    // internal BCP
    nRootSize             int32;    // the number of root level assignments
    nTrailSize             int32;   // the number of assignments made 
    pTrail            *lit;       // chronological order of assignments (size nVars)
    pAssigns            *lit;     // assignments by variable (size nVars) 
    pSeens           string;       // temporary mark (size nVars)
    pReasons     **Sto_Cls_t;     // reasons for each assignment (size nVars)          
    pWatches     **Sto_Cls_t;     // watched clauses for each literal (size 2*nVars)          
    // interpolation data
    pAig      *Aig_Man_t;         // the AIG manager for recording the interpolant
    pVarTypes            *int32;    // variable type (size nVars) [1=A, 0=B, <0=AB]
    pInters     **Aig_Obj_t;      // storage for interpolants as truth tables (size nClauses)
    nIntersAlloc             int32; // the allocated size of truth table array
    // proof recording
    Counter             int32;      // counter of resolved clauses
    pProofNums            *int32;   // the proof numbers for each clause (size nClauses)
    pFile           *os.File;        // the file for proof recording
    // internal verification
    vResLits      *Vec_Int_t;
    // runtime stats
    timeBcp         abctime;      // the runtime for BCP
    timeTrace         abctime;    // the runtime of trace construction
    timeTotal         abctime;    // the total runtime of interpolation
}

// procedure to get hold of the clauses' truth table 
func Inta_ManAigRead(pMan *Inta_Man_t, pCls *Sto_Cls_t) **Aig_Obj_t { return pMan.pInters + pCls.Id;          }
func Inta_ManAigClear(pMan *Inta_Man_t, p **Aig_Obj_t) { *p = Aig_ManConst0(pMan.pAig);           }
func Inta_ManAigFill(pMan *Inta_Man_t, p **Aig_Obj_t) { *p = Aig_ManConst1(pMan.pAig);           }
func Inta_ManAigCopy(pMan *Inta_Man_t, p **Aig_Obj_t, q **Aig_Obj_t) { *p = *q;                                  }
func Inta_ManAigAnd(pMan *Inta_Man_t, p **Aig_Obj_t, q **Aig_Obj_t) { *p = Aig_And(pMan.pAig, *p, *q);         }
func Inta_ManAigOr(pMan *Inta_Man_t, p **Aig_Obj_t, q **Aig_Obj_t) { *p = Aig_Or(pMan.pAig, *p, *q);          }
func Inta_ManAigOrNot(pMan *Inta_Man_t, p **Aig_Obj_t, q **Aig_Obj_t) { *p = Aig_Or(pMan.pAig, *p, Aig_Not(*q)); }
func Inta_ManAigOrVar(pMan *Inta_Man_t, p **Aig_Obj_t, v int32) { *p = Aig_Or(pMan.pAig, *p, Aig_IthVar(pMan.pAig, v));          }
func Inta_ManAigOrNotVar(pMan *Inta_Man_t, p **Aig_Obj_t, v int32) { *p = Aig_Or(pMan.pAig, *p, Aig_Not(Aig_IthVar(pMan.pAig, v))); }

// reading/writing the proof for a clause
func Inta_ManProofGet(p *Inta_Man_t, pCls *Sto_Cls_t) int32 { return p.pProofNums[pCls.Id];           }
func Inta_ManProofSet(p *Inta_Man_t, pCls *Sto_Cls_t, n int32) { p.pProofNums[pCls.Id] = n;              }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocate proof manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManAlloc() *Inta_Man_t {
    var p *Inta_Man_t;
    // allocate the manager
    // inlining of macro ABC_ALLOC
    p = (Inta_Man_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Inta_Man_t))));
    memset( p, 0, sizeof(Inta_Man_t) );
    // verification
    p.vResLits = Vec_IntAlloc( 1000 );
    // parameters
    p.fProofWrite = 0;
    p.fProofVerif = 1;
    return p;    
}

/**Function*************************************************************

  Synopsis    [Count common variables in the clauses of A and B.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManGlobalVars(p *Inta_Man_t) int32 {
    var pClause *Sto_Cls_t;
    var LargeNum int32  = -100000000;
    var Var, nVarsAB, v int32;

    // mark the variable encountered in the clauses of A
    // inlining of macro Sto_ManForEachClauseRoot
    for  pClause = p.pCnf.pHead; (pClause != nil) && pClause.fRoot; pClause = pClause.pNext  {
        if !pClause.fA {
            break; }
        for  v = 0; v < int32(pClause.nLits); v++  {
            p.pVarTypes[lit_var(pClause.pLits[v])] = 1;
        }
    }

    // check variables that appear in clauses of B
    nVarsAB = 0;
    // inlining of macro Sto_ManForEachClauseRoot
    for  pClause = p.pCnf.pHead; (pClause != nil) && pClause.fRoot; pClause = pClause.pNext  {
        if pClause.fA {
            continue; }
        for  v = 0; v < int32(pClause.nLits); v++  {
            Var = lit_var(pClause.pLits[v]);
            if p.pVarTypes[Var] == 1 { // var of A
            
                // change it into a global variable
                nVarsAB++;
                p.pVarTypes[Var] = LargeNum;
            }
        }
    }
    assert(nVarsAB <= Vec_IntSize(p.vVarsAB), "nVarsAB <= Vec_IntSize(p.vVarsAB)");

    // order global variables
    nVarsAB = 0;
    // inlining of macro Vec_IntForEachEntry
    for  v = 0; (v < Vec_IntSize(p.vVarsAB)) && (((Var) = Vec_IntEntry(p.vVarsAB, v)), 1); v++  {
        p.pVarTypes[Var] = -(1+nVarsAB++);
    }

    // check that there is no extra global variables
    for  v = 0; v < p.pCnf.nVars; v++  {
        assert(p.pVarTypes[v] != LargeNum, "p.pVarTypes[v] != LargeNum");
    }
    return nVarsAB;
}

/**Function*************************************************************

  Synopsis    [Resize proof manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManResize(p *Inta_Man_t) {
    p.Counter = 0;
    // check if resizing is needed
    if p.nVarsAlloc < p.pCnf.nVars {
        // find the new size
        if p.nVarsAlloc == 0 {
            p.nVarsAlloc = 1; }
        for  p.nVarsAlloc < p.pCnf.nVars  { 
            p.nVarsAlloc *= 2;
        }
        // resize the arrays
        // inlining of macro ABC_REALLOC
        
        if p.pTrail {
            p.pTrail = ((lit *) realloc((char *)(p.pTrail), sizeof(lit) * (size_t)(p.nVarsAlloc)));
        } else {
            p.pTrail = ((lit *) malloc(sizeof(lit) * (size_t)(p.nVarsAlloc)));
        }
        // inlining of macro ABC_REALLOC
        
        if p.pAssigns {
            p.pAssigns = ((lit *) realloc((char *)(p.pAssigns), sizeof(lit) * (size_t)(p.nVarsAlloc)));
        } else {
            p.pAssigns = ((lit *) malloc(sizeof(lit) * (size_t)(p.nVarsAlloc)));
        }
        // inlining of macro ABC_REALLOC
        
        if p.pSeens {
            p.pSeens = ((char *) realloc((char *)(p.pSeens), sizeof(char) * (size_t)(p.nVarsAlloc)));
        } else {
            p.pSeens = ((char *) malloc(sizeof(char) * (size_t)(p.nVarsAlloc)));
        }
        // inlining of macro ABC_REALLOC
        
        if p.pVarTypes {
            p.pVarTypes = ((int *) realloc((char *)(p.pVarTypes), sizeof(int) * (size_t)(p.nVarsAlloc)));
        } else {
            p.pVarTypes = ((int *) malloc(sizeof(int) * (size_t)(p.nVarsAlloc)));
        }
        // inlining of macro ABC_REALLOC
        
        if p.pReasons {
            p.pReasons = ((Sto_Cls_t * *) realloc((char *)(p.pReasons), sizeof(Sto_Cls_t *) * (size_t)(p.nVarsAlloc)));
        } else {
            p.pReasons = ((Sto_Cls_t * *) malloc(sizeof(Sto_Cls_t *) * (size_t)(p.nVarsAlloc)));
        }
        // inlining of macro ABC_REALLOC
        
        if p.pWatches {
            p.pWatches = ((Sto_Cls_t * *) realloc((char *)(p.pWatches), sizeof(Sto_Cls_t *) * (size_t)(p.nVarsAlloc*2)));
        } else {
            p.pWatches = ((Sto_Cls_t * *) malloc(sizeof(Sto_Cls_t *) * (size_t)(p.nVarsAlloc*2)));
        }
    }

    // clean the free space
    memset( p.pAssigns , 0xff, sizeof(lit) * p.pCnf.nVars );
    memset( p.pSeens   , 0,    sizeof(char) * p.pCnf.nVars );
    memset( p.pVarTypes, 0,    sizeof(int) * p.pCnf.nVars );
    memset( p.pReasons , 0,    sizeof(Sto_Cls_t *) * p.pCnf.nVars );
    memset( p.pWatches , 0,    sizeof(Sto_Cls_t *) * p.pCnf.nVars*2 );

    // compute the number of common variables
    Inta_ManGlobalVars( p );

    // check if resizing of clauses is needed
    if p.nClosAlloc < p.pCnf.nClauses {
        // find the new size
        if p.nClosAlloc == 0 {
            p.nClosAlloc = 1; }
        for  p.nClosAlloc < p.pCnf.nClauses  { 
            p.nClosAlloc *= 2;
        }
        // resize the arrays
        // inlining of macro ABC_REALLOC
        
        if p.pProofNums {
            p.pProofNums = ((int *) realloc((char *)(p.pProofNums), sizeof(int) * (size_t)(p.nClosAlloc)));
        } else {
            p.pProofNums = ((int *) malloc(sizeof(int) * (size_t)(p.nClosAlloc)));
        }
    }
    memset( p.pProofNums, 0, sizeof(int) * p.pCnf.nClauses );

    // check if resizing of truth tables is needed
    if p.nIntersAlloc < p.pCnf.nClauses {
        p.nIntersAlloc = p.pCnf.nClauses;
        // inlining of macro ABC_REALLOC
        
        if p.pInters {
            p.pInters = ((Aig_Obj_t * *) realloc((char *)(p.pInters), sizeof(Aig_Obj_t *) * (size_t)(p.nIntersAlloc)));
        } else {
            p.pInters = ((Aig_Obj_t * *) malloc(sizeof(Aig_Obj_t *) * (size_t)(p.nIntersAlloc)));
        }
    }
    memset( p.pInters, 0, sizeof(Aig_Obj_t *) * p.pCnf.nClauses );
}

/**Function*************************************************************

  Synopsis    [Deallocate proof manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManFree(p *Inta_Man_t) {
/*
    printf( "Runtime stats:\n" );
ABC_PRT( "BCP     ", p->timeBcp   );
ABC_PRT( "Trace   ", p->timeTrace );
ABC_PRT( "TOTAL   ", p->timeTotal );
*/
    p.pInters = nil;
    p.pProofNums = nil;
    p.pTrail = nil;
    p.pAssigns = nil;
    p.pSeens = nil;
    p.pVarTypes = nil;
    p.pReasons = nil;
    p.pWatches = nil;
    Vec_IntFree( p.vResLits );
    p = nil;
}




/**Function*************************************************************

  Synopsis    [Prints the clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManPrintClause(p *Inta_Man_t, pClause *Sto_Cls_t) {
    var i int32;
    fmt.Printf("Clause ID = %d. Proof = %d. {", pClause.Id, Inta_ManProofGet(p, pClause));
    for  i = 0; i < int32(pClause.nLits); i++  {
        fmt.Printf(" %d", pClause.pLits[i]);
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    [Prints the resolvent.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManPrintResolvent(vResLits *Vec_Int_t) {
    var i, Entry int32;
    fmt.Printf("Resolvent: {");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vResLits)) && (((Entry) = Vec_IntEntry(vResLits, i)), 1); i++  {
        fmt.Printf(" %d", Entry);
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    [Prints the interpolant for one clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManPrintInterOne(p *Inta_Man_t, pClause *Sto_Cls_t) {
    fmt.Printf("Clause %2d :  ", pClause.Id);
//    Extra_PrintBinary___( stdout, Inta_ManAigRead(p, pClause), (1 << p->nVarsAB) );
    fmt.Printf("\n");
}



/**Function*************************************************************

  Synopsis    [Adds one clause to the watcher list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManWatchClause(p *Inta_Man_t, pClause *Sto_Cls_t, Lit lit) {
    assert(lit_check(Lit, p.pCnf.nVars), "lit_check(Lit, p.pCnf.nVars)");
    if pClause.pLits[0] == Lit {
        pClause.pNext0 = p.pWatches[lit_neg(Lit)];  
    } else {
        assert(pClause.pLits[1] == Lit, "pClause.pLits[1] == Lit");
        pClause.pNext1 = p.pWatches[lit_neg(Lit)];  
    }
    p.pWatches[lit_neg(Lit)] = pClause;
}


/**Function*************************************************************

  Synopsis    [Records implication.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManEnqueue(p *Inta_Man_t, Lit lit, pReason *Sto_Cls_t) int32 {
    var Var int32  = lit_var(Lit);
    if p.pAssigns[Var] != LIT_UNDEF {
        return p.pAssigns[Var] == Lit; }
    p.pAssigns[Var] = Lit;
    p.pReasons[Var] = pReason;
    p.pTrail[p.nTrailSize++] = Lit;
//printf( "assigning var %d value %d\n", Var, !lit_sign(Lit) );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Records implication.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManCancelUntil(p *Inta_Man_t, Level int32) {
    var Lit lit;
    var i, Var int32;
    for  i = p.nTrailSize - 1; i >= Level; i--  {
        Lit = p.pTrail[i];
        Var = lit_var( Lit );
        p.pReasons[Var] = nil;
        p.pAssigns[Var] = LIT_UNDEF;
//printf( "cancelling var %d\n", Var );
    }
    p.nTrailSize = Level;
}

/**Function*************************************************************

  Synopsis    [Propagate one assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManPropagateOne(p *Inta_Man_t, Lit lit) *Sto_Cls_t {
    var ppPrev **Sto_Cls_t;
    var pCur *Sto_Cls_t;
    var pTemp *Sto_Cls_t;
    var LitF lit  = lit_neg(Lit);
    var i int32;
    // iterate through the literals
    ppPrev = p.pWatches + Lit;
    for  pCur = p.pWatches[Lit]; (pCur != nil); pCur = *ppPrev  {
        // make sure the false literal is in the second literal of the clause
        if pCur.pLits[0] == LitF {
            pCur.pLits[0] = pCur.pLits[1];
            pCur.pLits[1] = LitF;
            pTemp = pCur.pNext0;
            pCur.pNext0 = pCur.pNext1;
            pCur.pNext1 = pTemp;
        }
        assert(pCur.pLits[1] == LitF, "pCur.pLits[1] == LitF");

        // if the first literal is true, the clause is satisfied
        if pCur.pLits[0] == p.pAssigns[lit_var(pCur.pLits[0])] {
            ppPrev = &pCur.pNext1;
            continue;
        }

        // look for a new literal to watch
        for  i = 2; i < int32(pCur.nLits); i++  {
            // skip the case when the literal is false
            if lit_neg(pCur.pLits[i]) == p.pAssigns[lit_var(pCur.pLits[i])] {
                continue; }
            // the literal is either true or unassigned - watch it
            pCur.pLits[1] = pCur.pLits[i];
            pCur.pLits[i] = LitF;
            // remove this clause from the watch list of Lit
            *ppPrev = pCur.pNext1;
            // add this clause to the watch list of pCur->pLits[i] (now it is pCur->pLits[1])
            Inta_ManWatchClause( p, pCur, pCur.pLits[1] );
            break;
        }
        if i < int32(pCur.nLits) { // found new watch
            continue; }

        // clause is unit - enqueue new implication
        if Inta_ManEnqueue(p, pCur.pLits[0], pCur) != 0 {
            ppPrev = &pCur.pNext1;
            continue;
        }

        // conflict detected - return the conflict clause
        return pCur;
    }
    return nil;
}

/**Function*************************************************************

  Synopsis    [Propagate the current assignments.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManPropagate(p *Inta_Man_t, Start int32) *Sto_Cls_t {
    var pClause *Sto_Cls_t;
    var i int32;
    var clk abctime  = Abc_Clock();
    for  i = Start; i < p.nTrailSize; i++  {
        pClause = Inta_ManPropagateOne( p, p.pTrail[i] );
        if pClause != nil {
p.timeBcp += Abc_Clock() - clk;
            return pClause;
        }
    }
p.timeBcp += Abc_Clock() - clk;
    return nil;
}


/**Function*************************************************************

  Synopsis    [Writes one root clause into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManProofWriteOne(p *Inta_Man_t, pClause *Sto_Cls_t) {
    Inta_ManProofSet( p, pClause, ++p.Counter );

    if p.fProofWrite {
        var v int32;
        fmt.Fprintf(p.pFile, "%d", Inta_ManProofGet(p, pClause));
        for  v = 0; v < int32(pClause.nLits); v++  {
            fmt.Fprintf(p.pFile, " %d", lit_print(pClause.pLits[v]));
        }
        fmt.Fprintf(p.pFile, " 0 0\n");
    }
}

/**Function*************************************************************

  Synopsis    [Traces the proof for one clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManProofTraceOne(p *Inta_Man_t, pConflict *Sto_Cls_t, pFinal *Sto_Cls_t) int32 {
    var pReason *Sto_Cls_t;
    var i, v, Var, PrevId int32;
    var fPrint int32;
    var clk abctime  = Abc_Clock();

    // collect resolvent literals
    if p.fProofVerif {
        Vec_IntClear( p.vResLits );
        for  i = 0; i < int32(pConflict.nLits); i++  {
            Vec_IntPush( p.vResLits, pConflict.pLits[i] );
        }
    }

    // mark all the variables in the conflict as seen
    for  v = 0; v < int32(pConflict.nLits); v++  {
        p.pSeens[lit_var(pConflict.pLits[v])] = 1;
    }

    // start the anticedents
//    pFinal->pAntis = Vec_PtrAlloc( 32 );
//    Vec_PtrPush( pFinal->pAntis, pConflict );

    if p.pCnf.nClausesA {
        Inta_ManAigCopy( p, Inta_ManAigRead(p, pFinal), Inta_ManAigRead(p, pConflict) ); }

    // follow the trail backwards
    PrevId = Inta_ManProofGet(p, pConflict);
    for  i = p.nTrailSize - 1; i >= 0; i--  {
        // skip literals that are not involved
        Var = lit_var(p.pTrail[i]);
        if !p.pSeens[Var] {
            continue; }
        p.pSeens[Var] = 0;

        // skip literals of the resulting clause
        pReason = p.pReasons[Var];
        if pReason == nil {
            continue; }
        assert(p.pTrail[i] == pReason.pLits[0], "p.pTrail[i] == pReason.pLits[0]");

        // add the variables to seen
        for  v = 1; v < int32(pReason.nLits); v++  {
            p.pSeens[lit_var(pReason.pLits[v])] = 1;
        }

        // record the reason clause
        assert(Inta_ManProofGet(p, pReason) > 0, "Inta_ManProofGet(p, pReason) > 0");
        p.Counter++;
        if p.fProofWrite {
            fmt.Fprintf(p.pFile, "%d * %d %d 0\n", p.Counter, PrevId, Inta_ManProofGet(p, pReason)); }
        PrevId = p.Counter;

        if p.pCnf.nClausesA {
            if p.pVarTypes[Var] == 1 { // var of A
                Inta_ManAigOr( p, Inta_ManAigRead(p, pFinal), Inta_ManAigRead(p, pReason) );
            } else {
                Inta_ManAigAnd( p, Inta_ManAigRead(p, pFinal), Inta_ManAigRead(p, pReason) ); }
        }
 
        // resolve the temporary resolvent with the reason clause
        if p.fProofVerif {
            var v1 int32;
            var v2 int32;
            var Entry int32 = -1; 
            if fPrint != 0 {
                Inta_ManPrintResolvent( p.vResLits ); }
            // check that the var is present in the resolvent
            // inlining of macro Vec_IntForEachEntry
            for  v1 = 0; (v1 < Vec_IntSize(p.vResLits)) && (((Entry) = Vec_IntEntry(p.vResLits, v1)), 1); v1++  {
                if lit_var(Entry) == Var {
                    break; }
            }
            if v1 == Vec_IntSize(p.vResLits) {
                fmt.Printf("Recording clause %d: Cannot find variable %d in the temporary resolvent.\n", pFinal.Id, Var); }
            if Entry != lit_neg(pReason.pLits[0]) {
                fmt.Printf("Recording clause %d: The resolved variable %d is in the wrong polarity.\n", pFinal.Id, Var); }
            // remove variable v1 from the resolvent
            assert(lit_var(Entry) == Var, "lit_var(Entry) == Var");
            Vec_IntRemove( p.vResLits, Entry );
            // add variables of the reason clause
            for  v2 = 1; v2 < int32(pReason.nLits); v2++  {
                // inlining of macro Vec_IntForEachEntry
                for  v1 = 0; (v1 < Vec_IntSize(p.vResLits)) && (((Entry) = Vec_IntEntry(p.vResLits, v1)), 1); v1++  {
                    if lit_var(Entry) == lit_var(pReason.pLits[v2]) {
                        break; }
                }
                // if it is a new variable, add it to the resolvent
                if v1 == Vec_IntSize(p.vResLits) {
                    Vec_IntPush( p.vResLits, pReason.pLits[v2] );
                    continue;
                }
                // if the variable is the same, the literal should be the same too
                if Entry == pReason.pLits[v2] {
                    continue; }
                // the literal is different
                fmt.Printf("Recording clause %d: Trying to resolve the clause with more than one opposite literal.\n", pFinal.Id);
            }
        }
//        Vec_PtrPush( pFinal->pAntis, pReason );
    }

    // unmark all seen variables
//    for ( i = p->nTrailSize - 1; i >= 0; i-- )
//        p->pSeens[lit_var(p->pTrail[i])] = 0;
    // check that the literals are unmarked
//    for ( i = p->nTrailSize - 1; i >= 0; i-- )
//        assert( p->pSeens[lit_var(p->pTrail[i])] == 0 );

    // use the resulting clause to check the correctness of resolution
    if p.fProofVerif {
        var v1 int32;
        var v2 int32;
        var Entry int32 = -1; 
        if fPrint != 0 {
            Inta_ManPrintResolvent( p.vResLits ); }
        // inlining of macro Vec_IntForEachEntry
        for  v1 = 0; (v1 < Vec_IntSize(p.vResLits)) && (((Entry) = Vec_IntEntry(p.vResLits, v1)), 1); v1++  {
            for  v2 = 0; v2 < int32(pFinal.nLits); v2++  {
                if pFinal.pLits[v2] == Entry {
                    break; }
            }
            if v2 < int32(pFinal.nLits) {
                continue; }
            break;
        }
        if v1 < Vec_IntSize(p.vResLits) {
            fmt.Printf("Recording clause %d: The final resolvent is wrong.\n", pFinal.Id);
            Inta_ManPrintClause( p, pConflict );
            Inta_ManPrintResolvent( p.vResLits );
            Inta_ManPrintClause( p, pFinal );
        }

        // if there are literals in the clause that are not in the resolvent
        // it means that the derived resolvent is stronger than the clause
        // we can replace the clause with the resolvent by removing these literals
        if Vec_IntSize(p.vResLits) != int32(pFinal.nLits) {
            for  v1 = 0; v1 < int32(pFinal.nLits); v1++  {
                // inlining of macro Vec_IntForEachEntry
                for  v2 = 0; (v2 < Vec_IntSize(p.vResLits)) && (((Entry) = Vec_IntEntry(p.vResLits, v2)), 1); v2++  {
                    if pFinal.pLits[v1] == Entry {
                        break; }
                }
                if v2 < Vec_IntSize(p.vResLits) {
                    continue; }
                // remove literal v1 from the final clause
                pFinal.nLits--;
                for  v2 = v1; v2 < int32(pFinal.nLits); v2++  {
                    pFinal.pLits[v2] = pFinal.pLits[v2+1];
                }
                v1--;
            }
            assert(Vec_IntSize(p.vResLits) == int32(pFinal.nLits), "Vec_IntSize(p.vResLits) == int32(pFinal.nLits)");
        }
    }
p.timeTrace += Abc_Clock() - clk;

    // return the proof pointer 
    if p.pCnf.nClausesA {
//        Inta_ManPrintInterOne( p, pFinal );
    }
    Inta_ManProofSet( p, pFinal, p.Counter );
    // make sure the same proof ID is not asssigned to two consecutive clauses
    assert(p.pProofNums[pFinal.Id-1] != p.Counter, "p.pProofNums[pFinal.Id-1] != p.Counter");
//    if ( p->pProofNums[pFinal->Id] == p->pProofNums[pFinal->Id-1] )
//        p->pProofNums[pFinal->Id] = p->pProofNums[pConflict->Id];
    return p.Counter;
}

/**Function*************************************************************

  Synopsis    [Records the proof for one clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManProofRecordOne(p *Inta_Man_t, pClause *Sto_Cls_t) int32 {
    var pConflict *Sto_Cls_t;
    var i int32;

    // empty clause never ends up there
    assert(pClause.nLits > 0, "pClause.nLits > 0");
    if pClause.nLits == 0 {
        fmt.Printf("Error: Empty clause is attempted.\n"); }

    assert(!pClause.fRoot, "!pClause.fRoot");
    assert(p.nTrailSize == p.nRootSize, "p.nTrailSize == p.nRootSize");

    // if any of the clause literals are already assumed
    // it means that the clause is redundant and can be skipped
    for  i = 0; i < int32(pClause.nLits); i++  {
        if p.pAssigns[lit_var(pClause.pLits[i])] == pClause.pLits[i] {
            return 1; }
    }

    // add assumptions to the trail
    for  i = 0; i < int32(pClause.nLits); i++  {
        if !(Inta_ManEnqueue( p, lit_neg(pClause.pLits[i]), nil ) != 0) {
            assert(false, "0"); // impossible
            return 0;
        }
    }

    // propagate the assumptions
    pConflict = Inta_ManPropagate( p, p.nRootSize );
    if pConflict == nil {
        assert(false, "0"); // cannot prove
        return 0;
    } 

    // skip the clause if it is weaker or the same as the conflict clause
    if pClause.nLits >= pConflict.nLits {
        // check if every literal of conflict clause can be found in the given clause
        var j int32;
        for  i = 0; i < int32(pConflict.nLits); i++  {
            for  j = 0; j < int32(pClause.nLits); j++  {
                if pConflict.pLits[i] == pClause.pLits[j] {
                    break; }
            }
            if j == int32(pClause.nLits) { // literal pConflict->pLits[i] is not found
                break; }
        }
        if i == int32(pConflict.nLits) { // all lits are found
        
            // undo to the root level
            Inta_ManCancelUntil( p, p.nRootSize );
            return 1;
        }
    }

    // construct the proof
    Inta_ManProofTraceOne( p, pConflict, pClause );

    // undo to the root level
    Inta_ManCancelUntil( p, p.nRootSize );

    // add large clauses to the watched lists
    if pClause.nLits > 1 {
        Inta_ManWatchClause( p, pClause, pClause.pLits[0] );
        Inta_ManWatchClause( p, pClause, pClause.pLits[1] );
        return 1;
    }
    assert(pClause.nLits == 1, "pClause.nLits == 1");

    // if the clause proved is unit, add it and propagate
    if !(Inta_ManEnqueue( p, pClause.pLits[0], pClause ) != 0) {
        assert(false, "0"); // impossible
        return 0;
    }

    // propagate the assumption
    pConflict = Inta_ManPropagate( p, p.nRootSize );
    if pConflict != nil {
        // construct the proof
        Inta_ManProofTraceOne( p, pConflict, p.pCnf.pEmpty );
//        if ( p->fVerbose )
//            printf( "Found last conflict after adding unit clause number %d!\n", pClause->Id );
        return 0;
    }

    // update the root level
    p.nRootSize = p.nTrailSize;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Propagate the root clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManProcessRoots(p *Inta_Man_t) int32 {
    var pClause *Sto_Cls_t;
    var Counter int32;

    // make sure the root clauses are preceeding the learnt clauses
    Counter = 0;
    // inlining of macro Sto_ManForEachClause
    for  pClause = p.pCnf.pHead; (pClause != nil); pClause = pClause.pNext  {
        assert(int32(pClause.fA)    == (Counter < int32(p.pCnf.nClausesA)), "int32(pClause.fA)    == (Counter < int32(p.pCnf.nClausesA))");
        assert(int32(pClause.fRoot) == (Counter < int32(p.pCnf.nRoots)), "int32(pClause.fRoot) == (Counter < int32(p.pCnf.nRoots))");
        Counter++;
    }
    assert(p.pCnf.nClauses == Counter, "p.pCnf.nClauses == Counter");

    // make sure the last clause if empty
    assert(p.pCnf.pTail.nLits == 0, "p.pCnf.pTail.nLits == 0");

    // go through the root unit clauses
    p.nTrailSize = 0;
    // inlining of macro Sto_ManForEachClauseRoot
    for  pClause = p.pCnf.pHead; (pClause != nil) && pClause.fRoot; pClause = pClause.pNext  {
        // create watcher lists for the root clauses
        if pClause.nLits > 1 {
            Inta_ManWatchClause( p, pClause, pClause.pLits[0] );
            Inta_ManWatchClause( p, pClause, pClause.pLits[1] );
        }
        // empty clause and large clauses
        if pClause.nLits != 1 {
            continue; }
        // unit clause
        assert(lit_check(pClause.pLits[0], p.pCnf.nVars), "lit_check(pClause.pLits[0], p.pCnf.nVars)");
        if !(Inta_ManEnqueue( p, pClause.pLits[0], pClause ) != 0) {
            // detected root level conflict
//            printf( "Error in Inta_ManProcessRoots(): Detected a root-level conflict too early!\n" );
//            assert( 0 );
            // detected root level conflict
            Inta_ManProofTraceOne( p, pClause, p.pCnf.pEmpty );
            if p.fVerbose {
                fmt.Printf("Found root level conflict!\n"); }
            return 0;
        }
    }

    // propagate the root unit clauses
    pClause = Inta_ManPropagate( p, 0 );
    if pClause != nil {
        // detected root level conflict
        Inta_ManProofTraceOne( p, pClause, p.pCnf.pEmpty );
        if p.fVerbose {
            fmt.Printf("Found root level conflict!\n"); }
        return 0;
    }

    // set the root level
    p.nRootSize = p.nTrailSize;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Records the proof.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManPrepareInter(p *Inta_Man_t) {
    var pClause *Sto_Cls_t;
    var Var, VarAB, v int32;

    // set interpolants for root clauses
    // inlining of macro Sto_ManForEachClauseRoot
    for  pClause = p.pCnf.pHead; (pClause != nil) && pClause.fRoot; pClause = pClause.pNext  {
        if !pClause.fA { // clause of B
        
            Inta_ManAigFill( p, Inta_ManAigRead(p, pClause) );
//            Inta_ManPrintInterOne( p, pClause );
            continue;
        }
        // clause of A
        Inta_ManAigClear( p, Inta_ManAigRead(p, pClause) );
        for  v = 0; v < int32(pClause.nLits); v++  {
            Var = lit_var(pClause.pLits[v]);
            if p.pVarTypes[Var] < 0 { // global var
            
                VarAB = -p.pVarTypes[Var]-1;
                assert(VarAB >= 0 && VarAB < Vec_IntSize(p.vVarsAB), "VarAB >= 0 && VarAB < Vec_IntSize(p.vVarsAB)");
                if lit_sign(pClause.pLits[v]) { // negative var
                    Inta_ManAigOrNotVar( p, Inta_ManAigRead(p, pClause), VarAB );
                } else {
                    Inta_ManAigOrVar( p, Inta_ManAigRead(p, pClause), VarAB ); }
            }
        }
//        Inta_ManPrintInterOne( p, pClause );
    }
}

/**Function*************************************************************

  Synopsis    [Computes interpolant for the given CNF.]

  Description [Takes the interpolation manager, the CNF deriving by the SAT
  solver, which includes ClausesA, ClausesB, and learned clauses. Additional
  arguments are the vector of variables common to AB and the verbosiness flag.
  Returns the AIG manager with a single output, containing the interpolant.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManInterpolate(p *Inta_Man_t, pCnf *Sto_Man_t, TimeToStop abctime, vVarsAB any, fVerbose int32) any {
    var pRes *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var pClause *Sto_Cls_t;
    var RetValue int32  = 1;
    var clkTotal abctime  = Abc_Clock();

    if TimeToStop && Abc_Clock() > TimeToStop {
        return nil; }

    // check that the CNF makes sense
    assert(pCnf.nVars > 0 && pCnf.nClauses > 0, "pCnf.nVars > 0 && pCnf.nClauses > 0");
    p.pCnf = pCnf;
    p.fVerbose = fVerbose;
    p.vVarsAB = (Vec_Int_t *)vVarsAB;
    p.pAig = pRes = Aig_ManStart( 10000 );
    Aig_IthVar( p.pAig, Vec_IntSize(p.vVarsAB) - 1 );

    // adjust the manager
    Inta_ManResize( p );

    // prepare the interpolant computation
    Inta_ManPrepareInter( p );

    // construct proof for each clause
    // start the proof
    if p.fProofWrite {
        p.pFile = os.OpenFile("proof.cnf_", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        p.Counter = 0;
    }

    // write the root clauses
    // inlining of macro Sto_ManForEachClauseRoot
    for  pClause = p.pCnf.pHead; (pClause != nil) && pClause.fRoot; pClause = pClause.pNext  {
        Inta_ManProofWriteOne( p, pClause );
        if TimeToStop && Abc_Clock() > TimeToStop {
            Aig_ManStop( pRes );
            p.pAig = nil;
            return nil;
        }
    }

    // propagate root level assignments
    if Inta_ManProcessRoots( p ) != 0 {
        // if there is no conflict, consider learned clauses
        // inlining of macro Sto_ManForEachClause
        for  pClause = p.pCnf.pHead; (pClause != nil); pClause = pClause.pNext  {
            if pClause.fRoot {
                continue; }
            if !(Inta_ManProofRecordOne( p, pClause ) != 0) {
                RetValue = 0;
                break;
            }
            if TimeToStop && Abc_Clock() > TimeToStop {
                Aig_ManStop( pRes );
                p.pAig = nil;
                return nil;
            }
        }
    }

    // stop the proof
    if p.fProofWrite { 
        _ = (p.pFile).Close();
//        Sat_ProofChecker( "proof.cnf_" );
        p.pFile = nil;    
    }

    if fVerbose != 0 {
//        ABC_PRT( "Interpo", Abc_Clock() - clkTotal );
    fmt.Printf("Vars = %d. Roots = %d. Learned = %d. Resol steps = %d.  Ave = %.2f.  Mem = %.2f MB  ", p.pCnf.nVars, p.pCnf.nRoots, p.pCnf.nClauses-p.pCnf.nRoots, p.Counter, 1.0*(p.Counter-p.pCnf.nRoots)/(p.pCnf.nClauses-p.pCnf.nRoots), 1.0*Sto_ManMemoryReport(p.pCnf)/(1<<20));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
p.timeTotal += Abc_Clock() - clkTotal;
    }

    pObj = *Inta_ManAigRead( p, p.pCnf.pTail );
    Aig_ObjCreateCo( pRes, pObj );
    Aig_ManCleanup( pRes );

    p.pAig = nil;
    return pRes;
    
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Inta_ManDeriveClauses(pMan *Inta_Man_t, pCnf *Sto_Man_t, fClausesA int32) *Aig_Man_t {
    var p *Aig_Man_t;
    var pMiter, pSum, pLit *Aig_Obj_t;
    var pClause *Sto_Cls_t;
    var Var, VarAB, v int32;
    p = Aig_ManStart( 10000 );
    pMiter = Aig_ManConst1(p);
    // inlining of macro Sto_ManForEachClauseRoot
    for  pClause = pCnf.pHead; (pClause != nil) && pClause.fRoot; pClause = pClause.pNext  {
        if fClausesA ^ pClause.fA { // clause of B
            continue; }
        // clause of A
        pSum = Aig_ManConst0(p);
        for  v = 0; v < int32(pClause.nLits); v++  {
            Var = lit_var(pClause.pLits[v]);
            if pMan.pVarTypes[Var] < 0 { // global var
            
                VarAB = -pMan.pVarTypes[Var]-1;
                assert(VarAB >= 0 && VarAB < Vec_IntSize(pMan.vVarsAB), "VarAB >= 0 && VarAB < Vec_IntSize(pMan.vVarsAB)");
                pLit = Aig_NotCond( Aig_IthVar(p, VarAB), lit_sign(pClause.pLits[v]) );
            
            } else {
                pLit = Aig_NotCond( Aig_IthVar(p, Vec_IntSize(pMan.vVarsAB)+1+Var), lit_sign(pClause.pLits[v]) ); }
            pSum = Aig_Or( p, pSum, pLit );
        }
        pMiter = Aig_And( p, pMiter, pSum );
    }
    Aig_ObjCreateCo( p, pMiter );
    return p;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


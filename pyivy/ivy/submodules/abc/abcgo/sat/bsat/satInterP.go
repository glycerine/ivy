package bsat

import (
	"fmt"
	"time"
)

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
//#include "misc/vec/vec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// variable assignments 
var LIT_UNDEF lit = 0xffffffff;

// interpolation manager
type Intp_Man_t_ struct {
    // clauses of the problems
    pCnf      *Sto_Man_t;         // the set of CNF clauses for A and B
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
    // proof data
//    Vec_Int_t *     vAnties;      // anticedents for all clauses
//    Vec_Int_t *     vBreaks;      // beginnings of anticedents for each clause
    vAntClas      *Vec_Ptr_t;     // anticedant clauses
    nAntStart             int32;    // starting antecedant clause
    // proof recording
    Counter             int32;      // counter of resolved clauses
    pProofNums            *int32;   // the proof numbers for each clause (size nClauses)
    pFile           *os.File;        // the file for proof recording
    // internal verification
    pResLits            *lit;     // the literals of the resolvent   
    nResLits             int32;     // the number of literals of the resolvent
    nResLitsAlloc             int32;// the number of literals of the resolvent
    // runtime stats
    timeBcp         abctime;      // the runtime for BCP
    timeTrace         abctime;    // the runtime of trace construction
    timeTotal         abctime;    // the total runtime of interpolation
}

// reading/writing the proof for a clause
func Intp_ManProofGet(p *Intp_Man_t, pCls *Sto_Cls_t) int32 { return p.pProofNums[pCls.Id];           }
func Intp_ManProofSet(p *Intp_Man_t, pCls *Sto_Cls_t, n int32) { p.pProofNums[pCls.Id] = n;              }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocate proof manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Intp_ManAlloc() *Intp_Man_t {
    var p *Intp_Man_t;
    // allocate the manager
    // inlining of macro ABC_ALLOC
    p = (Intp_Man_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Intp_Man_t))));
    memset( p, 0, sizeof(Intp_Man_t) );
    // verification
    p.nResLitsAlloc = (1<<16);
    // inlining of macro ABC_ALLOC
    p.pResLits = ((lit *) malloc(sizeof(lit) * (size_t)(p.nResLitsAlloc)));
    // proof recording
//    p->vAnties = Vec_IntAlloc( 1000 );
//    p->vBreaks = Vec_IntAlloc( 1000 );
    p.vAntClas = Vec_PtrAlloc( 1000 );
    p.nAntStart = 0;
    // parameters
    p.fProofWrite = 0;
    p.fProofVerif = 1;
    return p;    
}

/**Function*************************************************************

  Synopsis    [Resize proof manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Intp_ManResize(p *Intp_Man_t) {
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
//        p->pVarTypes = ABC_REALLOC(int,         p->pVarTypes, p->nVarsAlloc );
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
//    memset( p->pVarTypes, 0,    sizeof(int) * p->pCnf->nVars );
    memset( p.pReasons , 0,    sizeof(Sto_Cls_t *) * p.pCnf.nVars );
    memset( p.pWatches , 0,    sizeof(Sto_Cls_t *) * p.pCnf.nVars*2 );

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
}

/**Function*************************************************************

  Synopsis    [Deallocate proof manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Intp_ManFree(p *Intp_Man_t) {
/*
    printf( "Runtime stats:\n" );
ABC_PRT( "BCP     ", p->timeBcp   );
ABC_PRT( "Trace   ", p->timeTrace );
ABC_PRT( "TOTAL   ", p->timeTotal );
*/
//    Vec_IntFree( p->vAnties );
//    Vec_IntFree( p->vBreaks );
    Vec_VecFree( (Vec_Vec_t *)p.vAntClas );
//    ABC_FREE( p->pInters );
    // inlining of macro ABC_FREE
    ((p.pProofNums) ? (free((char *) (p.pProofNums)), (p.pProofNums) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pTrail) ? (free((char *) (p.pTrail)), (p.pTrail) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pAssigns) ? (free((char *) (p.pAssigns)), (p.pAssigns) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSeens) ? (free((char *) (p.pSeens)), (p.pSeens) = 0) : 0);
//    ABC_FREE( p->pVarTypes );
    // inlining of macro ABC_FREE
    ((p.pReasons) ? (free((char *) (p.pReasons)), (p.pReasons) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pWatches) ? (free((char *) (p.pWatches)), (p.pWatches) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pResLits) ? (free((char *) (p.pResLits)), (p.pResLits) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}




/**Function*************************************************************

  Synopsis    [Prints the clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Intp_ManPrintClause(p *Intp_Man_t, pClause *Sto_Cls_t) {
    var i int32;
    fmt.Printf("Clause ID = %d. Proof = %d. {", pClause.Id, Intp_ManProofGet(p, pClause));
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
func Intp_ManPrintResolvent(pResLits *lit, nResLits int32) {
    var i int32;
    fmt.Printf("Resolvent: {");
    for  i = 0; i < nResLits; i++  {
        fmt.Printf(" %d", pResLits[i]);
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    [Prints the interpolant for one clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Intp_ManPrintInterOne(p *Intp_Man_t, pClause *Sto_Cls_t) {
    fmt.Printf("Clause %2d :  ", pClause.Id);
//    Extra_PrintBinary___( stdout, Intp_ManAigRead(p, pClause), (1 << p->nVarsAB) );
    fmt.Printf("\n");
}



/**Function*************************************************************

  Synopsis    [Adds one clause to the watcher list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Intp_ManWatchClause(p *Intp_Man_t, pClause *Sto_Cls_t, Lit lit) {
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
func Intp_ManEnqueue(p *Intp_Man_t, Lit lit, pReason *Sto_Cls_t) int32 {
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
func Intp_ManCancelUntil(p *Intp_Man_t, Level int32) {
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
func Intp_ManPropagateOne(p *Intp_Man_t, Lit lit) *Sto_Cls_t {
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
            Intp_ManWatchClause( p, pCur, pCur.pLits[1] );
            break;
        }
        if i < int32(pCur.nLits) { // found new watch
            continue; }

        // clause is unit - enqueue new implication
        if Intp_ManEnqueue(p, pCur.pLits[0], pCur) != 0 {
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
func Intp_ManPropagate(p *Intp_Man_t, Start int32) *Sto_Cls_t {
    var pClause *Sto_Cls_t;
    var i int32;
    var clk abctime  = Abc_Clock();
    for  i = Start; i < p.nTrailSize; i++  {
        pClause = Intp_ManPropagateOne( p, p.pTrail[i] );
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
func Intp_ManProofWriteOne(p *Intp_Man_t, pClause *Sto_Cls_t) {
    Intp_ManProofSet( p, pClause, ++p.Counter );

    if p.fProofWrite {
        var v int32;
        fmt.Fprintf(p.pFile, "%d", Intp_ManProofGet(p, pClause));
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
func Intp_ManProofTraceOne(p *Intp_Man_t, pConflict *Sto_Cls_t, pFinal *Sto_Cls_t) int32 {
    var pReason *Sto_Cls_t;
    var i, v, Var, PrevId int32;
    var fPrint int32;
    var clk abctime  = Abc_Clock();

    // collect resolvent literals
    if p.fProofVerif {
        assert(int32(pConflict.nLits) <= p.nResLitsAlloc, "int32(pConflict.nLits) <= p.nResLitsAlloc");
        memcpy( p.pResLits, pConflict.pLits, sizeof(lit) * pConflict.nLits );
        p.nResLits = pConflict.nLits;
    }

    // mark all the variables in the conflict as seen
    for  v = 0; v < int32(pConflict.nLits); v++  {
        p.pSeens[lit_var(pConflict.pLits[v])] = 1;
    }

    // start the anticedents
//    pFinal->pAntis = Vec_PtrAlloc( 32 );
//    Vec_PtrPush( pFinal->pAntis, pConflict );

//    assert( pFinal->Id == Vec_IntSize(p->vBreaks) );
//    Vec_IntPush( p->vBreaks, Vec_IntSize(p->vAnties) );
//    Vec_IntPush( p->vAnties, pConflict->Id );
    {
        var vAnts *Vec_Int_t  = Vec_IntAlloc( 16 );
        assert(Vec_PtrSize(p.vAntClas) == pFinal.Id - p.nAntStart, "Vec_PtrSize(p.vAntClas) == pFinal.Id - p.nAntStart");
        Vec_IntPush( vAnts, pConflict.Id );
        Vec_PtrPush( p.vAntClas, vAnts );
    }

//    if ( p->pCnf->nClausesA )
//        Intp_ManAigCopy( p, Intp_ManAigRead(p, pFinal), Intp_ManAigRead(p, pConflict) );

    // follow the trail backwards
    PrevId = Intp_ManProofGet(p, pConflict);
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
        assert(Intp_ManProofGet(p, pReason) > 0, "Intp_ManProofGet(p, pReason) > 0");
        p.Counter++;
        if p.fProofWrite {
            fmt.Fprintf(p.pFile, "%d * %d %d 0\n", p.Counter, PrevId, Intp_ManProofGet(p, pReason)); }
        PrevId = p.Counter;

//        if ( p->pCnf->nClausesA )
//       {
//            if ( p->pVarTypes[Var] == 1 ) // var of A
//                Intp_ManAigOr( p, Intp_ManAigRead(p, pFinal), Intp_ManAigRead(p, pReason) );
//            else
//                Intp_ManAigAnd( p, Intp_ManAigRead(p, pFinal), Intp_ManAigRead(p, pReason) );
//        }
 
        // resolve the temporary resolvent with the reason clause
        if p.fProofVerif {
            var v1, v2 int32; 
            if fPrint != 0 {
                Intp_ManPrintResolvent( p.pResLits, p.nResLits ); }
            // check that the var is present in the resolvent
            for  v1 = 0; v1 < p.nResLits; v1++  {
                if lit_var(p.pResLits[v1]) == Var {
                    break; }
            }
            if v1 == p.nResLits {
                fmt.Printf("Recording clause %d: Cannot find variable %d in the temporary resolvent.\n", pFinal.Id, Var); }
            if p.pResLits[v1] != lit_neg(pReason.pLits[0]) {
                fmt.Printf("Recording clause %d: The resolved variable %d is in the wrong polarity.\n", pFinal.Id, Var); }
            // remove this variable from the resolvent
            assert(lit_var(p.pResLits[v1]) == Var, "lit_var(p.pResLits[v1]) == Var");
            p.nResLits--;
            for  ; v1 < p.nResLits; v1++  {
                p.pResLits[v1] = p.pResLits[v1+1];
            }
            // add variables of the reason clause
            for  v2 = 1; v2 < int32(pReason.nLits); v2++  {
                for  v1 = 0; v1 < p.nResLits; v1++  {
                    if lit_var(p.pResLits[v1]) == lit_var(pReason.pLits[v2]) {
                        break; }
                }
                // if it is a new variable, add it to the resolvent
                if v1 == p.nResLits {
                    if p.nResLits == p.nResLitsAlloc {
                        fmt.Printf("Recording clause %d: Ran out of space for intermediate resolvent.\n", pFinal.Id); }
                    p.pResLits[ p.nResLits++ ] = pReason.pLits[v2];
                    continue;
                }
                // if the variable is the same, the literal should be the same too
                if p.pResLits[v1] == pReason.pLits[v2] {
                    continue; }
                // the literal is different
                fmt.Printf("Recording clause %d: Trying to resolve the clause with more than one opposite literal.\n", pFinal.Id);
            }
        }

//        Vec_PtrPush( pFinal->pAntis, pReason );
//        Vec_IntPush( p->vAnties, pReason->Id );
        Vec_IntPush( (Vec_Int_t *)Vec_PtrEntryLast(p.vAntClas), pReason.Id );
    }

    // unmark all seen variables
//    for ( i = p->nTrailSize - 1; i >= 0; i-- )
//        p->pSeens[lit_var(p->pTrail[i])] = 0;
    // check that the literals are unmarked
//    for ( i = p->nTrailSize - 1; i >= 0; i-- )
//        assert( p->pSeens[lit_var(p->pTrail[i])] == 0 );

    // use the resulting clause to check the correctness of resolution
    if p.fProofVerif {
        var v1, v2 int32; 
        if fPrint != 0 {
            Intp_ManPrintResolvent( p.pResLits, p.nResLits ); }
        for  v1 = 0; v1 < p.nResLits; v1++  {
            for  v2 = 0; v2 < int32(pFinal.nLits); v2++  {
                if pFinal.pLits[v2] == p.pResLits[v1] {
                    break; }
            }
            if v2 < int32(pFinal.nLits) {
                continue; }
            break;
        }
        if v1 < p.nResLits {
            fmt.Printf("Recording clause %d: The final resolvent is wrong.\n", pFinal.Id);
            Intp_ManPrintClause( p, pConflict );
            Intp_ManPrintResolvent( p.pResLits, p.nResLits );
            Intp_ManPrintClause( p, pFinal );
        }

        // if there are literals in the clause that are not in the resolvent
        // it means that the derived resolvent is stronger than the clause
        // we can replace the clause with the resolvent by removing these literals
        if p.nResLits != int32(pFinal.nLits) {
            for  v1 = 0; v1 < int32(pFinal.nLits); v1++  {
                for  v2 = 0; v2 < p.nResLits; v2++  {
                    if pFinal.pLits[v1] == p.pResLits[v2] {
                        break; }
                }
                if v2 < p.nResLits {
                    continue; }
                // remove literal v1 from the final clause
                pFinal.nLits--;
                for  v2 = v1; v2 < int32(pFinal.nLits); v2++  {
                    pFinal.pLits[v2] = pFinal.pLits[v2+1];
                }
                v1--;
            }
            assert(p.nResLits == int32(pFinal.nLits), "p.nResLits == int32(pFinal.nLits)");
        }
    }
p.timeTrace += Abc_Clock() - clk;

    // return the proof pointer 
//    if ( p->pCnf->nClausesA )
//    {
//        Intp_ManPrintInterOne( p, pFinal );
//    }
    Intp_ManProofSet( p, pFinal, p.Counter );
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
func Intp_ManProofRecordOne(p *Intp_Man_t, pClause *Sto_Cls_t) int32 {
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
//            Vec_IntPush( p->vBreaks, Vec_IntSize(p->vAnties) );
            Vec_PtrPush( p.vAntClas, Vec_IntAlloc(0) );
            return 1;
        }
    }

    // add assumptions to the trail
    for  i = 0; i < int32(pClause.nLits); i++  {
        if !(Intp_ManEnqueue( p, lit_neg(pClause.pLits[i]), nil ) != 0) {
            assert(false, "0"); // impossible
            return 0;
        }
    }

    // propagate the assumptions
    pConflict = Intp_ManPropagate( p, p.nRootSize );
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
            Intp_ManCancelUntil( p, p.nRootSize );
//            Vec_IntPush( p->vBreaks, Vec_IntSize(p->vAnties) );
            Vec_PtrPush( p.vAntClas, Vec_IntAlloc(0) );
            return 1;
        }
    }

    // construct the proof
    Intp_ManProofTraceOne( p, pConflict, pClause );

    // undo to the root level
    Intp_ManCancelUntil( p, p.nRootSize );

    // add large clauses to the watched lists
    if pClause.nLits > 1 {
        Intp_ManWatchClause( p, pClause, pClause.pLits[0] );
        Intp_ManWatchClause( p, pClause, pClause.pLits[1] );
        return 1;
    }
    assert(pClause.nLits == 1, "pClause.nLits == 1");

    // if the clause proved is unit, add it and propagate
    if !(Intp_ManEnqueue( p, pClause.pLits[0], pClause ) != 0) {
        assert(false, "0"); // impossible
        return 0;
    }

    // propagate the assumption
    pConflict = Intp_ManPropagate( p, p.nRootSize );
    if pConflict != nil {
        // insert place-holders till the empty clause
        for  Vec_PtrSize(p.vAntClas) < p.pCnf.pEmpty.Id - p.nAntStart  {
            Vec_PtrPush( p.vAntClas, Vec_IntAlloc(0) );
        }
        // construct the proof for the empty clause
        Intp_ManProofTraceOne( p, pConflict, p.pCnf.pEmpty );
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
func Intp_ManProcessRoots(p *Intp_Man_t) int32 {
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
            Intp_ManWatchClause( p, pClause, pClause.pLits[0] );
            Intp_ManWatchClause( p, pClause, pClause.pLits[1] );
        }
        // empty clause and large clauses
        if pClause.nLits != 1 {
            continue; }
        // unit clause
        assert(lit_check(pClause.pLits[0], p.pCnf.nVars), "lit_check(pClause.pLits[0], p.pCnf.nVars)");
        if !(Intp_ManEnqueue( p, pClause.pLits[0], pClause ) != 0) {
            // detected root level conflict
//            printf( "Error in Intp_ManProcessRoots(): Detected a root-level conflict too early!\n" );
//            assert( 0 );
            // detected root level conflict
            Intp_ManProofTraceOne( p, pClause, p.pCnf.pEmpty );
            if p.fVerbose {
                fmt.Printf("Found root level conflict!\n"); }
            return 0;
        }
    }

    // propagate the root unit clauses
    pClause = Intp_ManPropagate( p, 0 );
    if pClause != nil {
        // detected root level conflict
        Intp_ManProofTraceOne( p, pClause, p.pCnf.pEmpty );
        if p.fVerbose {
            fmt.Printf("Found root level conflict!\n"); }
        return 0;
    }

    // set the root level
    p.nRootSize = p.nTrailSize;
    return 1;
}


/**Function*************************************************************

  Synopsis    [Verifies the UNSAT core.]

  Description [Takes the interpolation manager, the CNF derived by the SAT
  solver, which includes the root clauses and the learned clauses. Returns
  the array of integers representing the number of root clauses that are in
  the UNSAT core.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Intp_ManUnsatCoreVerify(pCnf *Sto_Man_t, vCore *Vec_Int_t) {
    var fVerbose int32;
    var nConfMax int32  = 1000000;
    var pSat *sat_solver;
    var pClause *Sto_Cls_t;
    var vClauses *Vec_Ptr_t;
    var i, iClause, RetValue int32;
    var clk abctime  = Abc_Clock();
    // collect the clauses
    vClauses = Vec_PtrAlloc( 1000 );
    // inlining of macro Sto_ManForEachClauseRoot
    for  pClause = pCnf.pHead; (pClause != nil) && pClause.fRoot; pClause = pClause.pNext  {
        assert(Vec_PtrSize(vClauses) == pClause.Id, "Vec_PtrSize(vClauses) == pClause.Id");
        Vec_PtrPush( vClauses, pClause );
    }
    // create new SAT solver
    pSat = sat_solver_new();
//    sat_solver_setnvars( pSat, nSatVars );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCore)) && (((iClause) = Vec_IntEntry(vCore, i)), 1); i++  {
        pClause = (Sto_Cls_t *)Vec_PtrEntry( vClauses, iClause );
        if !sat_solver_addclause( pSat, pClause.pLits, pClause.pLits+pClause.nLits ) {
            fmt.Printf("The core verification problem is trivially UNSAT.\n");
            break;
        }
    }
    Vec_PtrFree( vClauses );
    // solve the problem
    RetValue = sat_solver_solve( pSat, nil, nil, (ABC_INT64_T)nConfMax, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    sat_solver_delete( pSat );
    if fVerbose != 0 {
        if RetValue == l_Undef {
            fmt.Printf("Conflict limit is reached.  ");
        } else if RetValue == l_True {
            fmt.Printf("UNSAT core verification FAILED.  ");
        } else {
            fmt.Printf("UNSAT core verification succeeded.  "); }
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    
    } else {
        if RetValue == l_True {
            fmt.Printf("UNSAT core verification FAILED.  \n"); }
    }
}

/**Function*************************************************************

  Synopsis    [Recursively computes the UNSAT core.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Intp_ManUnsatCore_rec(vAntClas *Vec_Ptr_t, iThis int32, vCore *Vec_Int_t, nRoots int32, vVisited *Vec_Str_t, fLearned int32) {
//    int i, iStop, iStart;
    var vAnt *Vec_Int_t;
    var i, Entry int32;
    // skip visited clauses
    if Vec_StrEntry( vVisited, iThis ) {
        return; }
    Vec_StrWriteEntry( vVisited, iThis, 1 );
    // add a root clause to the core
    if iThis < nRoots {
        if !(fLearned != 0) {
            Vec_IntPush( vCore, iThis ); }
        return; 
    }
    // iterate through the clauses
//    iStart = Vec_IntEntry( vBreaks, iThis );
//    iStop = Vec_IntEntry( vBreaks, iThis+1 );
//    assert( iStop != -1 );
//    for ( i = iStart; i < iStop; i++ )
    vAnt = (Vec_Int_t *)Vec_PtrEntry( vAntClas, iThis - nRoots );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vAnt)) && (((Entry) = Vec_IntEntry(vAnt, i)), 1); i++  {
//        Intp_ManUnsatCore_rec( vAntClas, Vec_IntEntry(vAnties, i), vCore, nRoots, vVisited );
        Intp_ManUnsatCore_rec( vAntClas, Entry, vCore, nRoots, vVisited, fLearned );
    }
    // collect learned clause
    if fLearned != 0 {
        Vec_IntPush( vCore, iThis ); }
}

/**Function*************************************************************

  Synopsis    [Computes UNSAT core of the satisfiablity problem.]

  Description [Takes the interpolation manager, the CNF derived by the SAT
  solver, which includes the root clauses and the learned clauses. Returns
  the array of integers representing the number of root clauses that are in
  the UNSAT core.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Intp_ManUnsatCore(p *Intp_Man_t, pCnf *Sto_Man_t, fLearned int32, fVerbose int32) any {
    var vCore *Vec_Int_t;
    var vVisited *Vec_Str_t;
    var pClause *Sto_Cls_t;
    var RetValue int32  = 1;
    var clkTotal abctime  = Abc_Clock();

    // check that the CNF makes sense
    assert(pCnf.nVars > 0 && pCnf.nClauses > 0, "pCnf.nVars > 0 && pCnf.nClauses > 0");
    p.pCnf = pCnf;
    p.fVerbose = fVerbose;

    // adjust the manager
    Intp_ManResize( p ); 

    // construct proof for each clause
    // start the proof
    if p.fProofWrite {
        p.pFile = os.OpenFile("proof.cnf_", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        p.Counter = 0;
    }

    // write the root clauses
//    Vec_IntClear( p->vAnties );
//    Vec_IntFill( p->vBreaks, p->pCnf->nRoots, 0 );
    Vec_PtrClear( p.vAntClas );
    p.nAntStart = p.pCnf.nRoots;

    // inlining of macro Sto_ManForEachClauseRoot
    for  pClause = p.pCnf.pHead; (pClause != nil) && pClause.fRoot; pClause = pClause.pNext  {
        Intp_ManProofWriteOne( p, pClause );
    }

    // propagate root level assignments
    if Intp_ManProcessRoots( p ) != 0 {
        // if there is no conflict, consider learned clauses
        // inlining of macro Sto_ManForEachClause
        for  pClause = p.pCnf.pHead; (pClause != nil); pClause = pClause.pNext  {
            if pClause.fRoot {
                continue; }
            if !(Intp_ManProofRecordOne( p, pClause ) != 0) {
                RetValue = 0;
                break;
            }
        }
    }

    // add the last breaker
//    assert( p->pCnf->pEmpty->Id == Vec_IntSize(p->vBreaks) - 1 );
//    Vec_IntPush( p->vBreaks, Vec_IntSize(p->vAnties) );
    assert(p.pCnf.pEmpty.Id - p.nAntStart == Vec_PtrSize(p.vAntClas) - 1, "p.pCnf.pEmpty.Id - p.nAntStart == Vec_PtrSize(p.vAntClas) - 1");
    Vec_PtrPush( p.vAntClas, Vec_IntAlloc(0) );

    // stop the proof
    if p.fProofWrite {
        _ = (p.pFile).Close();
//        Sat_ProofChecker( "proof.cnf_" );
        p.pFile = nil;    
    }

    if fVerbose != 0 {
//        ABC_PRT( "Core", Abc_Clock() - clkTotal );
    fmt.Printf("Vars = %d. Roots = %d. Learned = %d. Resol steps = %d.  Ave = %.2f.  Mem = %.2f MB\n", p.pCnf.nVars, p.pCnf.nRoots, p.pCnf.nClauses-p.pCnf.nRoots, p.Counter, 1.0*(p.Counter-p.pCnf.nRoots)/(p.pCnf.nClauses-p.pCnf.nRoots), 1.0*Sto_ManMemoryReport(p.pCnf)/(1<<20));
p.timeTotal += Abc_Clock() - clkTotal;
    }

    // derive the UNSAT core
    vCore = Vec_IntAlloc( 1000 );
    vVisited = Vec_StrStart( p.pCnf.pEmpty.Id+1 );
    Intp_ManUnsatCore_rec( p.vAntClas, p.pCnf.pEmpty.Id, vCore, p.pCnf.nRoots, vVisited, fLearned );
    Vec_StrFree( vVisited );
    if fVerbose != 0 {
        fmt.Printf("Root clauses = %d. Learned clauses = %d. UNSAT core size = %d.\n", p.pCnf.nRoots, p.pCnf.nClauses-p.pCnf.nRoots, Vec_IntSize(vCore)); }
//    Intp_ManUnsatCoreVerify( p->pCnf, vCore );
    return vCore;   
}

/**Function*************************************************************

  Synopsis    [Prints learned clauses in terms of original problem varibles.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Intp_ManUnsatCorePrintForBmc(pFile *os.File, pCnf *Sto_Man_t, vCore0 any, vVarMap0 any) {
    var vCore *Vec_Int_t    = (Vec_Int_t *)vCore0;
    var vVarMap *Vec_Int_t  = (Vec_Int_t *)vVarMap0;
    var vClaMap *Vec_Ptr_t;
    var pClause *Sto_Cls_t;
    var v, i, iClause, fCompl, iObj, iFrame int32;
    // create map of clause
    vClaMap = Vec_PtrAlloc( pCnf.nClauses );
    // inlining of macro Sto_ManForEachClause
    for  pClause = pCnf.pHead; (pClause != nil); pClause = pClause.pNext  {
        Vec_PtrPush( vClaMap, pClause );
    }
    // print clauses
    fmt.Fprintf(pFile, "UNSAT contains %d learned clauses:\n", Vec_IntSize(vCore));
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCore)) && (((iClause) = Vec_IntEntry(vCore, i)), 1); i++  {
        pClause = (Sto_Cls_t *)Vec_PtrEntry(vClaMap, iClause);
        fmt.Fprintf(pFile, "%6d : %6d : ", i, iClause - pCnf.nRoots);
        for  v = 0; v < int32(pClause.nLits); v++  {
            fCompl = Abc_LitIsCompl(pClause.pLits[v]);
            iObj   = Vec_IntEntry(vVarMap, 2*Abc_Lit2Var(pClause.pLits[v]));
            iFrame = Vec_IntEntry(vVarMap, 2*Abc_Lit2Var(pClause.pLits[v])+1);
            fmt.Fprintf(pFile, "%s%d(%d) ", fCompl ? "!":"", iObj, iFrame);
        }
        if pClause.nLits == 0 {
            fmt.Fprintf(pFile, "Empty"); }
        fmt.Fprintf(pFile, "\n");
    }
    Vec_PtrFree( vClaMap );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


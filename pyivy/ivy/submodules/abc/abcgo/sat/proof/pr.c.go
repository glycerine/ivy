package proof

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [pr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Proof recording.]

  Synopsis    [Core procedures of the package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: pr.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/util/abc_global.h"
//#include "pr.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

typedef unsigned lit;

   
type Pr_Cls_t struct {
    uTruth        uint32;       // interpolant
    pProof           any;       // the proof node 
//    void *          pAntis;       // the anticedents
    pNext       *Pr_Cls_t;        // the next clause
    pNext0       *Pr_Cls_t;       // the next 0-watch
    pNext1       *Pr_Cls_t;       // the next 0-watch
    Id             int32;           // the clause ID
    fA uint32; /* C bitfield: 1 bits; review packing and truncation */  // belongs to A
    fRoot uint32; /* C bitfield: 1 bits; review packing and truncation */  // original clause
    fVisit uint32; /* C bitfield: 1 bits; review packing and truncation */  // visited clause
    nLits uint32; /* C bitfield: 24 bits; review packing and truncation */  // the number of literals
    lit             pLits[0];     // literals of this clause
}

type Pr_Man_t_ struct {
    // general data
    fProofWrite             int32;  // writes the proof file
    fProofVerif             int32;  // verifies the proof
    nVars             int32;        // the number of variables
    nVarsAB             int32;      // the number of global variables
    nRoots             int32;       // the number of root clauses
    nClauses             int32;     // the number of all clauses
    nClausesA             int32;    // the number of clauses of A 
    pHead       *Pr_Cls_t;        // the head clause
    pTail       *Pr_Cls_t;        // the tail clause
    pLearnt       *Pr_Cls_t;      // the tail clause
    pEmpty       *Pr_Cls_t;       // the empty clause
    // internal BCP
    nRootSize             int32;    // the number of root level assignments
    nTrailSize             int32;   // the number of assignments made 
    pTrail            *lit;       // chronological order of assignments (size nVars)
    pAssigns            *lit;     // assignments by variable (size nVars) 
    pSeens           string;       // temporary mark (size nVars)
    pVarTypes           string;    // variable type (size nVars) [1=A, 0=B, <0=AB]
    pReasons      **Pr_Cls_t;     // reasons for each assignment (size nVars)          
    pWatches      **Pr_Cls_t;     // watched clauses for each literal (size 2*nVars)          
    nVarsAlloc             int32;   // the allocated size of arrays
    // proof recording
    pManProof           any;    // proof manager
    Counter             int32;      // counter of resolved clauses
    // memory management
    nChunkSize             int32;   // the number of bytes in a chunk
    nChunkUsed             int32;   // the number of bytes used in the last chunk
    pChunkLast           string;   // the last memory chunk
    // internal verification
    pResLits            *lit;     // the literals of the resolvent   
    nResLits             int32;     // the number of literals of the resolvent
    nResLitsAlloc             int32;// the number of literals of the resolvent
    // runtime stats
    timeBcp         abctime;
    timeTrace         abctime;
    timeRead         abctime;
    timeTotal         abctime;
}

// variable assignments 
var LIT_UNDEF lit = 0xffffffff;

// variable/literal conversions (taken from MiniSat)
func toLit(v int32) lit { return v + v;  }
func toLitCond(v int32, c int32) lit { return v + v + (c != 0); }
func lit_neg(l lit) lit { return l ^ 1;  }
func lit_var(l lit) int32 { return l >> 1; }
func lit_sign(l lit) int32 { return l & 1;  }
func lit_print(l lit) int32 { if lit_sign(l) != 0 {
    return -lit_var(l)-1;
} else {
    return lit_var(l)+1;
} }
func lit_read(s int32) lit { if s > 0 {
    return toLit(s-1);
} else {
    return lit_neg(toLit(-s-1));
} }
func lit_check(l lit, n int32) int32 { return l >= 0 && lit_var(l) < n;                  }

// iterators through the clauses
// Original macro:
// #define Pr_ManForEachClause( p, pCls )        for( pCls = p->pHead; pCls; pCls = pCls->pNext )

// Original macro:
// #define Pr_ManForEachClauseRoot( p, pCls )    for( pCls = p->pHead; pCls != p->pLearnt; pCls = pCls->pNext )

// Original macro:
// #define Pr_ManForEachClauseLearnt( p, pCls )  for( pCls = p->pLearnt; pCls; pCls = pCls->pNext )


// static procedures
// static char *     Pr_ManMemoryFetch( Pr_Man_t * p, int nBytes );

// static void       Pr_ManMemoryStop( Pr_Man_t * p );

// static void       Pr_ManResize( Pr_Man_t * p, int nVarsNew );


// exported procedures
// extern Pr_Man_t * Pr_ManAlloc( int nVarsAlloc );

// extern void       Pr_ManFree( Pr_Man_t * p );

// extern int        Pr_ManAddClause( Pr_Man_t * p, lit * pBeg, lit * pEnd, int fRoot, int fClauseA );

// extern int        Pr_ManProofWrite( Pr_Man_t * p );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocate proof manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManAlloc(nVarsAlloc int32) *Pr_Man_t {
    var p *Pr_Man_t;
    // allocate the manager
    // inlining of macro ABC_ALLOC
    p = (Pr_Man_t *)((char *) malloc(sizeof(char) * (size_t)(sizeof(Pr_Man_t))));
    memset( p, 0, sizeof(Pr_Man_t) );
    // allocate internal arrays
    Pr_ManResize( p, nVarsAlloc? nVarsAlloc : 256 );
    // set the starting number of variables
    p.nVars = 0;
    // memory management
    p.nChunkSize = (1<<16); // use 64K chunks
    // verification
    p.nResLitsAlloc = (1<<16);
    // inlining of macro ABC_ALLOC
    p.pResLits = ((lit *) malloc(sizeof(lit) * (size_t)(p.nResLitsAlloc)));
    // parameters
    p.fProofWrite = 0;
    p.fProofVerif = 0;
    return p;    
}

/**Function*************************************************************

  Synopsis    [Resize proof manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManResize(p *Pr_Man_t, nVarsNew int32) {
    // check if resizing is needed
    if p.nVarsAlloc < nVarsNew {
        var nVarsAllocOld int32  = p.nVarsAlloc;
        // find the new size
        if p.nVarsAlloc == 0 {
            p.nVarsAlloc = 1; }
        for  p.nVarsAlloc < nVarsNew  { 
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
            p.pVarTypes = ((char *) realloc((char *)(p.pVarTypes), sizeof(char) * (size_t)(p.nVarsAlloc)));
        } else {
            p.pVarTypes = ((char *) malloc(sizeof(char) * (size_t)(p.nVarsAlloc)));
        }
        // inlining of macro ABC_REALLOC
        
        if p.pReasons {
            p.pReasons = ((Pr_Cls_t * *) realloc((char *)(p.pReasons), sizeof(Pr_Cls_t *) * (size_t)(p.nVarsAlloc)));
        } else {
            p.pReasons = ((Pr_Cls_t * *) malloc(sizeof(Pr_Cls_t *) * (size_t)(p.nVarsAlloc)));
        }
        // inlining of macro ABC_REALLOC
        
        if p.pWatches {
            p.pWatches = ((Pr_Cls_t * *) realloc((char *)(p.pWatches), sizeof(Pr_Cls_t *) * (size_t)(p.nVarsAlloc*2)));
        } else {
            p.pWatches = ((Pr_Cls_t * *) malloc(sizeof(Pr_Cls_t *) * (size_t)(p.nVarsAlloc*2)));
        }
        // clean the free space
        memset( p.pAssigns  + nVarsAllocOld, 0xff, sizeof(lit) * (p.nVarsAlloc - nVarsAllocOld) );
        memset( p.pSeens    + nVarsAllocOld, 0, sizeof(char) * (p.nVarsAlloc - nVarsAllocOld) );
        memset( p.pVarTypes + nVarsAllocOld, 0, sizeof(char) * (p.nVarsAlloc - nVarsAllocOld) );
        memset( p.pReasons  + nVarsAllocOld, 0, sizeof(Pr_Cls_t *) * (p.nVarsAlloc - nVarsAllocOld) );
        memset( p.pWatches  + nVarsAllocOld, 0, sizeof(Pr_Cls_t *) * (p.nVarsAlloc - nVarsAllocOld)*2 );
    }
    // adjust the number of variables
    if p.nVars < nVarsNew {
        p.nVars = nVarsNew; }
}

/**Function*************************************************************

  Synopsis    [Deallocate proof manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManFree(p *Pr_Man_t) {
    fmt.Printf("Runtime stats:\n");
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Reading "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeRead)))/(float64(int64(time.Second))));;
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("BCP     "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeBcp)))/(float64(int64(time.Second))));;
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Trace   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTrace)))/(float64(int64(time.Second))));;
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("TOTAL   "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;

    Pr_ManMemoryStop( p );
    p.pTrail = nil;
    p.pAssigns = nil;
    p.pSeens = nil;
    p.pVarTypes = nil;
    p.pReasons = nil;
    p.pWatches = nil;
    p.pResLits = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Adds one clause to the watcher list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManWatchClause(p *Pr_Man_t, pClause *Pr_Cls_t, Lit lit) {
    assert((lit_check(Lit, p.nVars) != 0), "lit_check(Lit, p.nVars)");
    if pClause.pLits[0] == Lit {
        pClause.pNext0 = p.pWatches[lit_neg(Lit)];  
    } else {
        assert(pClause.pLits[1] == Lit, "pClause.pLits[1] == Lit");
        pClause.pNext1 = p.pWatches[lit_neg(Lit)];  
    }
    p.pWatches[lit_neg(Lit)] = pClause;
}

/**Function*************************************************************

  Synopsis    [Adds one clause to the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManAddClause(p *Pr_Man_t, pBeg *lit, pEnd *lit, fRoot int32, fClauseA int32) int32 {
    var pClause *Pr_Cls_t;
    var Lit lit;
    var i *lit;
    var j *lit;
    var nSize, VarMax int32;

    // process the literals
    if pBeg < pEnd {
        // insertion sort
        VarMax = lit_var( *pBeg );
        for  i = pBeg + 1; i < pEnd; i++  {
            Lit = *i;
            if lit_var(Lit) > VarMax {
                VarMax = lit_var(Lit);
            } else {
                VarMax = VarMax;
            }
            for  j = i; j > pBeg && *(j-1) > Lit; j--  {
                *j = *(j-1);
            }
            *j = Lit;
        }
        // make sure there is no duplicated variables
        for  i = pBeg + 1; i < pEnd; i++  {
            assert(lit_var(*(i-1)) != lit_var(*i), "lit_var(*(i-1)) != lit_var(*i)");
        }
        // resize the manager
        Pr_ManResize( p, VarMax+1 );
    }

    // get memory for the clause
    nSize = sizeof(Pr_Cls_t) + sizeof(lit) * (pEnd - pBeg);
    pClause = (Pr_Cls_t *)Pr_ManMemoryFetch( p, nSize );
    *pClause = Pr_Cls_t{};

    // assign the clause
    assert(!(fClauseA != 0) || fRoot, "!fClauseA || fRoot"); // clause of A is always a root clause
    p.nRoots += fRoot;
    p.nClausesA += fClauseA;
    pClause.Id = p.nClauses++;
    pClause.fA = fClauseA;
    pClause.fRoot = fRoot;
    pClause.nLits = pEnd - pBeg;
    memcpy( pClause.pLits, pBeg, sizeof(lit) * (pEnd - pBeg) );

    // add the clause to the list
    if p.pHead == nil {
        p.pHead = pClause; }
    if p.pTail == nil {
        p.pTail = pClause;
    } else {
        p.pTail.pNext = pClause;
        p.pTail = pClause;
    }

    // mark the first learnt clause
    if p.pLearnt == nil && !pClause.fRoot {
        p.pLearnt = pClause; }

    // add the empty clause
    if pClause.nLits == 0 {
        if p.pEmpty {
            fmt.Printf("More than one empty clause!\n"); }
        p.pEmpty = pClause;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Fetches memory.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManMemoryFetch(p *Pr_Man_t, nBytes int32) string {
    var pMem string;
    if p.pChunkLast == nil || nBytes > p.nChunkSize - p.nChunkUsed {
        // inlining of macro ABC_ALLOC
        pMem = (char *)((char *) malloc(sizeof(char) * (size_t)(p.nChunkSize)));
        *(char **)pMem = p.pChunkLast;
        p.pChunkLast = pMem;
        p.nChunkUsed = sizeof(char *);
    }
    pMem = p.pChunkLast + p.nChunkUsed;
    p.nChunkUsed += nBytes;
    return pMem;
}

/**Function*************************************************************

  Synopsis    [Frees memory manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManMemoryStop(p *Pr_Man_t) {
    var pMem, pNext string;
    if p.pChunkLast == nil {
        return; }
    for  pMem = p.pChunkLast; pNext = *(char **)pMem; pMem = pNext  {
        pMem = nil;
    }
    pMem = nil;
}

/**Function*************************************************************

  Synopsis    [Reports memory usage in bytes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManMemoryReport(p *Pr_Man_t) int32 {
    var Total int32;
    var pMem, pNext string;
    if p.pChunkLast == nil {
        return 0; }
    Total = p.nChunkUsed; 
    for  pMem = p.pChunkLast; pNext = *(char **)pMem; pMem = pNext  {
        Total += p.nChunkSize;
    }
    return Total;
}

/**Function*************************************************************

  Synopsis    [Records the proof.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_PrintBinary_(pFile *os.File, Sign []uint32, nBits int32) {
    var Remainder, nWords int32;
    var w, i int32;

    Remainder = (nBits%(sizeof(unsigned)*8));
    nWords    = (nBits/(sizeof(unsigned)*8)) + (Remainder>0);

    for  w = nWords-1; w >= 0; w--  {
        for  i = ((w == nWords-1 && Remainder)? Remainder-1: 31); i >= 0; i--  {
            fmt.Fprintf(pFile, "%c", '0' + int32(((Sign[w] & (1<<i)) > 0)));
        }
    }
}

/**Function*************************************************************

  Synopsis    [Prints the interpolant for one clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManPrintInterOne(p *Pr_Man_t, pClause *Pr_Cls_t) {
    fmt.Printf("Clause %2d :  ", pClause.Id);
    Extra_PrintBinary_( os.Stdout, &pClause.uTruth, (1 << p.nVarsAB) );
    fmt.Printf("\n");
}



/**Function*************************************************************

  Synopsis    [Records implication.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManEnqueue(p *Pr_Man_t, Lit lit, pReason *Pr_Cls_t) int32 {
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
func Pr_ManCancelUntil(p *Pr_Man_t, Level int32) {
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
func Pr_ManPropagateOne(p *Pr_Man_t, Lit lit) *Pr_Cls_t {
    var ppPrev **Pr_Cls_t;
    var pCur *Pr_Cls_t;
    var pTemp *Pr_Cls_t;
    var LitF lit  = lit_neg(Lit);
    var i int32;
    // iterate through the literals
    ppPrev = p.pWatches + Lit;
    for  pCur = p.pWatches[Lit]; pCur; pCur = *ppPrev  {
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
            Pr_ManWatchClause( p, pCur, pCur.pLits[1] );
            break;
        }
        if i < int32(pCur.nLits) { // found new watch
            continue; }

        // clause is unit - enqueue new implication
        if Pr_ManEnqueue(p, pCur.pLits[0], pCur) != 0 {
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
func Pr_ManPropagate(p *Pr_Man_t, Start int32) *Pr_Cls_t {
    var pClause *Pr_Cls_t;
    var i int32;
    var clk abctime  = Abc_Clock();
    for  i = Start; i < p.nTrailSize; i++  {
        pClause = Pr_ManPropagateOne( p, p.pTrail[i] );
        if pClause != nil {
p.timeBcp += Abc_Clock() - clk;
            return pClause;
        }
    }
p.timeBcp += Abc_Clock() - clk;
    return nil;
}


/**Function*************************************************************

  Synopsis    [Prints the clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManPrintClause(pClause *Pr_Cls_t) {
    var i int32;
    fmt.Printf("Clause ID = %d. Proof = %d. {", pClause.Id, int32(pClause.pProof));
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
func Pr_ManPrintResolvent(pResLits *lit, nResLits int32) {
    var i int32;
    fmt.Printf("Resolvent: {");
    for  i = 0; i < nResLits; i++  {
        fmt.Printf(" %d", pResLits[i]);
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    [Writes one root clause into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManProofWriteOne(p *Pr_Man_t, pClause *Pr_Cls_t) {
    pClause.pProof = (void *)++p.Counter;

    if p.fProofWrite {
        var v int32;
        fmt.Fprintf((FILE *)p.pManProof, "%d", int32(pClause.pProof));
        for  v = 0; v < int32(pClause.nLits); v++  {
            fmt.Fprintf((FILE *)p.pManProof, " %d", lit_print(pClause.pLits[v]));
        }
        fmt.Fprintf((FILE *)p.pManProof, " 0 0\n");
    }
}

/**Function*************************************************************

  Synopsis    [Traces the proof for one clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManProofTraceOne(p *Pr_Man_t, pConflict *Pr_Cls_t, pFinal *Pr_Cls_t) int32 {
    var pReason *Pr_Cls_t;
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

    if p.nClausesA {
        pFinal.uTruth = pConflict.uTruth; }
   
    // follow the trail backwards
    PrevId = int32(pConflict.pProof);
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
        assert(pReason.pProof > 0, "pReason.pProof > 0");
        p.Counter++;
        if p.fProofWrite {
            fmt.Fprintf((FILE *)p.pManProof, "%d * %d %d 0\n", p.Counter, PrevId, int32(pReason.pProof)); }
        PrevId = p.Counter;

        if p.nClausesA {
            if p.pVarTypes[Var] == 1 { // var of A
                pFinal.uTruth |= pReason.uTruth;
            } else {
                pFinal.uTruth &= pReason.uTruth; }
        }
 
        // resolve the temporary resolvent with the reason clause
        if p.fProofVerif {
            var v1, v2 int32; 
            if fPrint != 0 {
                Pr_ManPrintResolvent( p.pResLits, p.nResLits ); }
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
                        fmt.Printf("Recording clause %d: Ran out of space for intermediate resolvent.\n, pFinal->Id"); }
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
            Pr_ManPrintResolvent( p.pResLits, p.nResLits ); }
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
            Pr_ManPrintClause( pConflict );
            Pr_ManPrintResolvent( p.pResLits, p.nResLits );
            Pr_ManPrintClause( pFinal );
        }
    }
p.timeTrace += Abc_Clock() - clk;

    // return the proof pointer 
    if p.nClausesA {
        Pr_ManPrintInterOne( p, pFinal );
    }
    return p.Counter;
}

/**Function*************************************************************

  Synopsis    [Records the proof for one clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManProofRecordOne(p *Pr_Man_t, pClause *Pr_Cls_t) int32 {
    var pConflict *Pr_Cls_t;
    var i int32;

    // empty clause never ends up there
    assert(pClause.nLits > 0, "pClause.nLits > 0");
    if pClause.nLits == 0 {
        fmt.Printf("Error: Empty clause is attempted.\n"); }

    // add assumptions to the trail
    assert(!pClause.fRoot, "!pClause.fRoot");
    assert(p.nTrailSize == p.nRootSize, "p.nTrailSize == p.nRootSize");
    for  i = 0; i < int32(pClause.nLits); i++  {
        if !(Pr_ManEnqueue( p, lit_neg(pClause.pLits[i]), nil ) != 0) {
            assert(false, "0"); // impossible
            return 0;
        }
    }

    // propagate the assumptions
    pConflict = Pr_ManPropagate( p, p.nRootSize );
    if pConflict == nil {
        assert(false, "0"); // cannot prove
        return 0;
    }

    // construct the proof
    pClause.pProof = (void *)Pr_ManProofTraceOne( p, pConflict, pClause );

    // undo to the root level
    Pr_ManCancelUntil( p, p.nRootSize );

    // add large clauses to the watched lists
    if pClause.nLits > 1 {
        Pr_ManWatchClause( p, pClause, pClause.pLits[0] );
        Pr_ManWatchClause( p, pClause, pClause.pLits[1] );
        return 1;
    }
    assert(pClause.nLits == 1, "pClause.nLits == 1");

    // if the clause proved is unit, add it and propagate
    if !(Pr_ManEnqueue( p, pClause.pLits[0], pClause ) != 0) {
        assert(false, "0"); // impossible
        return 0;
    }

    // propagate the assumption
    pConflict = Pr_ManPropagate( p, p.nRootSize );
    if pConflict != nil {
        // construct the proof
        p.pEmpty.pProof = (void *)Pr_ManProofTraceOne( p, pConflict, p.pEmpty );
        fmt.Printf("Found last conflict after adding unit clause number %d!\n", pClause.Id);
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
func Pr_ManProcessRoots(p *Pr_Man_t) int32 {
    var pClause *Pr_Cls_t;
    var Counter int32;

    // make sure the root clauses are preceeding the learnt clauses
    Counter = 0;
    // inlining of macro Pr_ManForEachClause
    for  pClause = p.pHead; (pClause != nil); pClause = pClause.pNext  {
        assert(int32(pClause.fA)    == (Counter < int32(p.nClausesA)), "int32(pClause.fA)    == (Counter < int32(p.nClausesA))");
        assert(int32(pClause.fRoot) == (Counter < int32(p.nRoots)), "int32(pClause.fRoot) == (Counter < int32(p.nRoots))");
        Counter++;
    }
    assert(p.nClauses == Counter, "p.nClauses == Counter");

    // make sure the last clause if empty
    assert(p.pTail.nLits == 0, "p.pTail.nLits == 0");

    // go through the root unit clauses
    p.nTrailSize = 0;
    // inlining of macro Pr_ManForEachClauseRoot
    for  pClause = p.pHead; pClause != p.pLearnt; pClause = pClause.pNext  {
        // create watcher lists for the root clauses
        if pClause.nLits > 1 {
            Pr_ManWatchClause( p, pClause, pClause.pLits[0] );
            Pr_ManWatchClause( p, pClause, pClause.pLits[1] );
        }
        // empty clause and large clauses
        if pClause.nLits != 1 {
            continue; }
        // unit clause
        assert((lit_check(pClause.pLits[0], p.nVars) != 0), "lit_check(pClause.pLits[0], p.nVars)");
        if !(Pr_ManEnqueue( p, pClause.pLits[0], pClause ) != 0) {
            // detected root level conflict
            fmt.Printf("Pr_ManProcessRoots(): Detected a root-level conflict\n");
            assert(false, "0");
            return 0;
        }
    }

    // propagate the root unit clauses
    pClause = Pr_ManPropagate( p, 0 );
    if pClause != nil {
        // detected root level conflict
        fmt.Printf("Pr_ManProcessRoots(): Detected a root-level conflict\n");
        assert(false, "0");
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
func Pr_ManPrepareInter(p *Pr_Man_t) {
    var uTruths [5]uint32 = [5]uint32{ 0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000 };
    var pClause *Pr_Cls_t;
    var Var, v int32;

    // mark the variable encountered in the clauses of A
    // inlining of macro Pr_ManForEachClauseRoot
    for  pClause = p.pHead; pClause != p.pLearnt; pClause = pClause.pNext  {
        if !pClause.fA {
            break; }
        for  v = 0; v < int32(pClause.nLits); v++  {
            p.pVarTypes[lit_var(pClause.pLits[v])] = 1;
        }
    }

    // check variables that appear in clauses of B
    p.nVarsAB = 0;
    // inlining of macro Pr_ManForEachClauseRoot
    for  pClause = p.pHead; pClause != p.pLearnt; pClause = pClause.pNext  {
        if pClause.fA {
            continue; }
        for  v = 0; v < int32(pClause.nLits); v++  {
            Var = lit_var(pClause.pLits[v]);
            if p.pVarTypes[Var] == 1 { // var of A
            
                // change it into a global variable
                p.nVarsAB++;
                p.pVarTypes[Var] = -1;
            }
        }
    }

    // order global variables
    p.nVarsAB = 0;
    for  v = 0; v < p.nVars; v++  {
        if p.pVarTypes[v] == -1 {
            p.pVarTypes[v] -= p.nVarsAB++; }
    }
fmt.Printf("There are %d global variables.\n", p.nVarsAB);

    // set interpolants for root clauses
    // inlining of macro Pr_ManForEachClauseRoot
    for  pClause = p.pHead; pClause != p.pLearnt; pClause = pClause.pNext  {
        if !pClause.fA { // clause of B
        
            pClause.uTruth = ~0;
            Pr_ManPrintInterOne( p, pClause );
            continue;
        }
        // clause of A
        pClause.uTruth = 0;
        for  v = 0; v < int32(pClause.nLits); v++  {
            Var = lit_var(pClause.pLits[v]);
            if p.pVarTypes[Var] < 0 { // global var
            
                if lit_sign(pClause.pLits[v]) != 0 { // negative var
                    pClause.uTruth |= ~uTruths[ -p.pVarTypes[Var]-1 ];
                } else {
                    pClause.uTruth |= uTruths[ -p.pVarTypes[Var]-1 ]; }
            }
        }
        Pr_ManPrintInterOne( p, pClause );
    }
}

/**Function*************************************************************

  Synopsis    [Records the proof.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManProofWrite(p *Pr_Man_t) int32 {
    var pClause *Pr_Cls_t;
    var RetValue int32  = 1;

    // propagate root level assignments
    Pr_ManProcessRoots( p );

    // prepare the interpolant computation
    if p.nClausesA {
        Pr_ManPrepareInter( p ); }

    // construct proof for each clause
    // start the proof
    if p.fProofWrite {
        p.pManProof = os.OpenFile("proof.cnf_", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666); }
    p.Counter = 0;

    // write the root clauses
    // inlining of macro Pr_ManForEachClauseRoot
    for  pClause = p.pHead; pClause != p.pLearnt; pClause = pClause.pNext  {
        Pr_ManProofWriteOne( p, pClause );
    }

    // consider each learned clause
    // inlining of macro Pr_ManForEachClauseLearnt
    for  pClause = p.pLearnt; (pClause != nil); pClause = pClause.pNext  {
        if !(Pr_ManProofRecordOne( p, pClause ) != 0) {
            RetValue = 0;
            break;
        }
    }

    if p.nClausesA {
        fmt.Printf("Interpolant:  ");
    }


    // stop the proof
    if p.fProofWrite {
        _ = ((FILE *)p.pManProof).Close();
        p.pManProof = nil;    
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Reads clauses from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManProofRead(pFileName string) *Pr_Man_t {
    var p *Pr_Man_t  = nil;
    var pCur string;
    var pBuffer string = nil;
    var pArray *int32  = nil;
    var pFile *os.File;
    var RetValue, Counter, nNumbers, Temp int32;
    var nClauses, nClausesA, nRoots, nVars int32;

    // open the file
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Count not open input file \"%s\".\n", pFileName);
        return nil;
    }

    // read the file
    // inlining of macro ABC_ALLOC
    pBuffer = (char *)((char *) malloc(sizeof(char) * (size_t)((1<<16))));
    for  Counter = 0; fgets( pBuffer, (1<<16), pFile );  {
        if pBuffer[0] == 'c' {
            continue; }
        if pBuffer[0] == 'p' {
            assert(p == nil, "p == nil");
            nClausesA = 0;
            RetValue = sscanf( pBuffer + 1, "%d %d %d %d", &nVars, &nClauses, &nRoots, &nClausesA );
            if RetValue != 3 && RetValue != 4 {
                fmt.Printf("Wrong input file format.\n");
            }
            p = Pr_ManAlloc( nVars );
            // inlining of macro ABC_ALLOC
            pArray = (int *)((char *) malloc(sizeof(char) * (size_t)(sizeof(int) * (nVars + 10))));
            continue;
        }
        // skip empty lines
        for  pCur = pBuffer; *pCur; pCur++  {
            if !(*pCur == ' ' || *pCur == '\t' || *pCur == '\r' || *pCur == '\n') {
                break; }
        }
        if *pCur == 0 {
            continue; }
        // scan the numbers from file
        nNumbers = 0;
        pCur = pBuffer;
        for  *pCur  {
            // skip spaces
            for  ; *pCur && *pCur == ' '; pCur++  {}
            // read next number
            Temp = 0;
            sscanf( pCur, "%d", &Temp );
            if Temp == 0 {
                break; }
            pArray[ nNumbers++ ] = lit_read( Temp );
            // skip non-spaces
            for  ; *pCur && *pCur != ' '; pCur++  {}
        }
        // add the clause
        if !Pr_ManAddClause( p, (unsigned *)pArray, (unsigned *)pArray + nNumbers, Counter < nRoots, Counter < nClausesA ) {
            fmt.Printf("Bad clause number %d.\n", Counter);
            return nil;
        }
        // count the clauses
        Counter++;
    }
    // check the number of clauses
    if Counter != nClauses {
        fmt.Printf("Expected %d clauses but read %d.\n", nClauses, Counter); }

    // finish
    if pArray != nil { pArray = nil; }
    if pBuffer { pBuffer = nil; }
    _ = (pFile).Close();
    return p;
}

/**Function*************************************************************

  Synopsis    [Records the proof.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
int Pr_ManProofCount_rec( Pr_Cls_t * pClause )
{
    Pr_Cls_t * pNext;
    int i, Counter;    
    if ( pClause->fRoot )
        return 0;
    if ( pClause->fVisit )
        return 0;
    pClause->fVisit = 1;
    // count the number of visited clauses
    Counter = 1;
    Vec_PtrForEachEntry( Pr_Cls_t *, pClause->pAntis, pNext, i )
        Counter += Pr_ManProofCount_rec( pNext );
    return Counter;
}
*/

/**Function*************************************************************

  Synopsis    [Records the proof.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pr_ManProofTest(pFileName string) int32 {
    var p *Pr_Man_t;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();

clk = Abc_Clock();
    p = Pr_ManProofRead( pFileName );
p.timeRead = Abc_Clock() - clk;
    if p == nil {
        return 0; }

    Pr_ManProofWrite( p );

    // print stats
/*
    nUsed = Pr_ManProofCount_rec( p->pEmpty );
    printf( "Roots = %d. Learned = %d. Total = %d. Steps = %d.  Ave = %.2f.  Used = %d. Ratio = %.2f. \n", 
        p->nRoots, p->nClauses-p->nRoots, p->nClauses, p->Counter,  
        1.0*(p->Counter-p->nRoots)/(p->nClauses-p->nRoots),
        nUsed, 1.0*nUsed/(p->nClauses-p->nRoots)  );
*/
    fmt.Printf("Vars = %d. Roots = %d. Learned = %d. Resol steps = %d.  Ave = %.2f.  Mem = %.2f MB\n", p.nVars, p.nRoots, p.nClauses-p.nRoots, p.Counter, 1.0*(p.Counter-p.nRoots)/(p.nClauses-p.nRoots), 1.0*Pr_ManMemoryReport(p)/(1<<20));

p.timeTotal = Abc_Clock() - clkTotal;
    Pr_ManFree( p );
    return 1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


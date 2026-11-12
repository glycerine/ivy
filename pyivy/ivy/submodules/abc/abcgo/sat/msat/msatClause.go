package msat

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [msatClause.c]

  PackageName [A C version of SAT solver MINISAT, originally developed 
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of 
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [Procedures working with SAT clauses.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatClause.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

//#include "msatInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Msat_Clause_t_ struct {
    Num           int32;              // unique number of the clause
    fLearned uint32; /* C bitfield: 1 bits; review packing and truncation */  // 1 if the clause is learned
    fMark uint32; /* C bitfield: 1 bits; review packing and truncation */  // used to mark visited clauses during proof recording
    fTypeA uint32; /* C bitfield: 1 bits; review packing and truncation */  // used to mark clauses belonging to A for interpolant computation
    nSize uint32; /* C bitfield: 14 bits; review packing and truncation */  // the number of literals in the clause
    nSizeAlloc uint32; /* C bitfield: 15 bits; review packing and truncation */  // the number of bytes allocated for the clause
    Msat_Lit_t     pData[0];
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates a new clause.]

  Description [Returns FALSE if top-level conflict detected (must be handled); 
  TRUE otherwise. 'pClause_out' may be set to NULL if clause is already 
  satisfied by the top-level assignment.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseCreate(p *Msat_Solver_t, vLits *Msat_IntVec_t, fLearned int32, pClause_out **Msat_Clause_t) int32 {
    var pAssigns *int32  = Msat_SolverReadAssignsArray(p);
    var pvWatched **Msat_ClauseVec_t;
    var pC *Msat_Clause_t;
    var pLits *int32;
    var nLits, i, j int32;
    var nBytes int32;
    var Var Msat_Var_t;
    var Sign int32;

    *pClause_out = nil;

    nLits = Msat_IntVecReadSize(vLits);
    pLits = Msat_IntVecReadArray(vLits);

    if !fLearned {
        var pSeen *int32  = Msat_SolverReadSeenArray( p );
        var nSeenId int32;
        assert(Msat_SolverReadDecisionLevel(p) == 0, "Msat_SolverReadDecisionLevel(p) == 0");
        // sorting literals makes the code trace-equivalent 
        // with to the original C++ solver
        Msat_IntVecSort( vLits, 0 );
        // increment the counter of seen twice
        nSeenId = Msat_SolverIncrementSeenId( p );
        nSeenId = Msat_SolverIncrementSeenId( p );
        // nSeenId - 1 stands for negative
        // nSeenId     stands for positive
        // Remove false literals

        // there is a bug here!!!!
        // when the same var in opposite polarities is given, it drops one polarity!!!

        for  i = j = 0; i < nLits; i++  {
            // get the corresponding variable
            // inlining of macro MSAT_LIT2VAR
            Var  = ((pLits[i])>>1);
            // inlining of macro MSAT_LITSIGN
            Sign = ((pLits[i])&1)     ; // Sign=0 for positive
            // check if we already saw this variable in the this clause
            if pSeen[Var] >= nSeenId - 1 {
                if (pSeen[Var] != nSeenId) == Sign { // the same lit
                    continue; }
                return 1; // two opposite polarity lits -- don't add the clause
            }
            // mark the variable as seen
            pSeen[Var] = nSeenId - !Sign;

            // analize the value of this literal
            if pAssigns[Var] != MSAT_VAR_UNASSIGNED {
                if pAssigns[Var] == pLits[i] {
                    return 1; }  // the clause is always true -- don't add anything
                // the literal has no impact - skip it
                continue;
            }
            // otherwise, add this literal to the clause
            pLits[j++] = pLits[i];
        }
        Msat_IntVecShrink( vLits, j );
        nLits = j;
/*
        // the problem with this code is that performance is very
        // sensitive to the ordering of adjacency lits
        // the best ordering requires fanins first, next fanouts
        // this ordering is more convenient to make from FRAIG

        // create the adjacency information
        if ( nLits > 2 )
        {
            Msat_Var_t VarI, VarJ;
            Msat_IntVec_t * pAdjI, * pAdjJ;

            for ( i = 0; i < nLits; i++ )
            {
                VarI = MSAT_LIT2VAR(pLits[i]);
                pAdjI = (Msat_IntVec_t *)p->vAdjacents->pArray[VarI];

                for ( j = i+1; j < nLits; j++ )
                {
                    VarJ = MSAT_LIT2VAR(pLits[j]);
                    pAdjJ = (Msat_IntVec_t *)p->vAdjacents->pArray[VarJ];

                    Msat_IntVecPushUniqueOrder( pAdjI, VarJ, 1 );
                    Msat_IntVecPushUniqueOrder( pAdjJ, VarI, 1 );
                }
            }
        }
*/
    }
    // 'vLits' is now the (possibly) reduced vector of literals.
    if nLits == 0 {
        return 0; }
    if nLits == 1 {
        return Msat_SolverEnqueue( p, pLits[0], nil ); }

    // Allocate clause:
//    nBytes = sizeof(unsigned)*(nLits + 1 + (int)fLearned);
    nBytes = sizeof(unsigned)*(nLits + 2 + int32(fLearned));
#ifdef USE_SYSTEM_MEMORY_MANAGEMENT
    pC = (Msat_Clause_t *)ABC_ALLOC( char, nBytes );
#else
    pC = (Msat_Clause_t *)Msat_MmStepEntryFetch( Msat_SolverReadMem(p), nBytes );
#endif
    pC.Num        = p.nClauses++;
    pC.fTypeA     = 0;
    pC.fMark      = 0;
    pC.fLearned   = fLearned;
    pC.nSize      = nLits;
    pC.nSizeAlloc = nBytes;
    memcpy( pC.pData, pLits, sizeof(int)*nLits );

    // For learnt clauses only:
    if fLearned {
        var pLevel *int32  = Msat_SolverReadDecisionLevelArray( p );
        var iLevelMax, iLevelCur, iLitMax int32;

        // Put the second watch on the literal with highest decision level:
        iLitMax = 1;
        // inlining of macro MSAT_LIT2VAR
        iLevelMax = pLevel[ ((pLits[1])>>1) ];
        for  i = 2; i < nLits; i++  {
            // inlining of macro MSAT_LIT2VAR
            iLevelCur = pLevel[ ((pLits[i])>>1) ];
            assert(iLevelCur != -1, "iLevelCur != -1");
            if iLevelMax < iLevelCur {
            // this is very strange - shouldn't it be???
            // if ( iLevelMax > iLevelCur )
                iLevelMax = iLevelCur;
iLitMax = i; }
        }
        pC.pData[1]       = pLits[iLitMax];
        pC.pData[iLitMax] = pLits[1];

        // Bumping:
        // (newly learnt clauses should be considered active)
        Msat_ClauseWriteActivity( pC, 0.0 );
        Msat_SolverClaBumpActivity( p, pC ); 
//        if ( nLits < 20 )
        for  i = 0; i < nLits; i++  {
            Msat_SolverVarBumpActivity( p, pLits[i] );
//            Msat_SolverVarBumpActivity( p, pLits[i] );
//            p->pFreq[ MSAT_LIT2VAR(pLits[i]) ]++;
        }
    }

    // Store clause:
    pvWatched = Msat_SolverReadWatchedArray( p );
    // inlining of macro MSAT_LITNOT
    Msat_ClauseVecPush( pvWatched[ ((pC.pData[0])^1) ], pC );
    // inlining of macro MSAT_LITNOT
    Msat_ClauseVecPush( pvWatched[ ((pC.pData[1])^1) ], pC );
    *pClause_out = pC;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Deallocates the clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseFree(p *Msat_Solver_t, pC *Msat_Clause_t, fRemoveWatched int32) {
    if fRemoveWatched != 0 {
        var Lit Msat_Lit_t;
        var pvWatched **Msat_ClauseVec_t;
        pvWatched = Msat_SolverReadWatchedArray( p );
        // inlining of macro MSAT_LITNOT
        Lit = ((pC.pData[0])^1);
        Msat_ClauseRemoveWatch( pvWatched[Lit], pC );
        // inlining of macro MSAT_LITNOT
        Lit = ((pC.pData[1])^1);
        Msat_ClauseRemoveWatch( pvWatched[Lit], pC ); 
    }

#ifdef USE_SYSTEM_MEMORY_MANAGEMENT
    ABC_FREE( pC );
#else
    Msat_MmStepEntryRecycle( Msat_SolverReadMem(p), (char *)pC, pC->nSizeAlloc );
#endif

}

/**Function*************************************************************

  Synopsis    [Access the data field of the clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseReadLearned(pC *Msat_Clause_t) int32 {  return pC.fLearned; }
func Msat_ClauseReadSize(pC *Msat_Clause_t) int32 {  return pC.nSize;    }
func Msat_ClauseReadLits(pC *Msat_Clause_t) *int32 {  return pC.pData;    }
func Msat_ClauseReadMark(pC *Msat_Clause_t) int32 {  return pC.fMark;    }
func Msat_ClauseReadNum(pC *Msat_Clause_t) int32 {  return pC.Num;      }
func Msat_ClauseReadTypeA(pC *Msat_Clause_t) int32 {  return pC.fTypeA;    }

func Msat_ClauseSetMark(pC *Msat_Clause_t, fMark int32) {  pC.fMark = fMark;   }
func Msat_ClauseSetNum(pC *Msat_Clause_t, Num int32) {  pC.Num = Num;       }
func Msat_ClauseSetTypeA(pC *Msat_Clause_t, fTypeA int32) {  pC.fTypeA = fTypeA; }

/**Function*************************************************************

  Synopsis    [Checks whether the learned clause is locked.]

  Description [The clause may be locked if it is the reason of a
  recent conflict. Such clause cannot be removed from the database.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseIsLocked(p *Msat_Solver_t, pC *Msat_Clause_t) int32 {
    var pReasons **Msat_Clause_t  = Msat_SolverReadReasonArray( p );
    // inlining of macro MSAT_LIT2VAR
    return int32( (pReasons[((pC.pData[0])>>1)] == pC));
}

/**Function*************************************************************

  Synopsis    [Reads the activity of the given clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseReadActivity(pC *Msat_Clause_t) float32 {
    var f float32;

    memcpy( &f, pC.pData + pC.nSize, sizeof (f));
    return f;
}

/**Function*************************************************************

  Synopsis    [Sets the activity of the clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseWriteActivity(pC *Msat_Clause_t, Num float32) {
    memcpy( pC.pData + pC.nSize, &Num, sizeof (Num) );
}

/**Function*************************************************************

  Synopsis    [Propages the assignment.]

  Description [The literal that has become true (Lit) is given to this 
  procedure. The array of current variable assignments is given for
  efficiency. The output literal (pLit_out) can be the second watched
  literal (if TRUE is returned) or the conflict literal (if FALSE is 
  returned). This messy interface is used to improve performance.
  This procedure accounts for ~50% of the runtime of the solver.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClausePropagate(pC *Msat_Clause_t, Lit Msat_Lit_t, pAssigns *int32, pLit_out *Msat_Lit_t) int32 {
    // make sure the false literal is pC->pData[1]
    // inlining of macro MSAT_LITNOT
    var LitF Msat_Lit_t  = ((Lit)^1);
    if pC.pData[0] == LitF {
         pC.pData[0] = pC.pData[1], pC.pData[1] = LitF; }
    assert(pC.pData[1] == LitF, "pC.pData[1] == LitF");
    // if the 0-th watch is true, clause is already satisfied
    // inlining of macro MSAT_LIT2VAR
    if pAssigns[((pC.pData[0])>>1)] == pC.pData[0] {
        return 1; } 
    // look for a new watch
    if pC.nSize > 2 {
        var i int32;
        for  i = 2; i < int32(pC.nSize); i++  {
            // inlining of macro MSAT_LIT2VAR
            // inlining of macro MSAT_LITNOT
            if pAssigns[((pC.pData[i])>>1)] != ((pC.pData[i])^1) {
                pC.pData[1] = pC.pData[i], pC.pData[i] = LitF;
                // inlining of macro MSAT_LITNOT
                *pLit_out = ((pC.pData[1])^1);
                return 1; 
            }
        }
    }
    // clause is unit under assignment
    *pLit_out = pC.pData[0];
    return 0;
}

/**Function*************************************************************

  Synopsis    [Simplifies the clause.]

  Description [Assumes everything has been propagated! (esp. watches 
  in clauses are NOT unsatisfied)]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseSimplify(pC *Msat_Clause_t, pAssigns *int32) int32 {
    var Var Msat_Var_t;
    var i, j int32;
    for  i = j = 0; i < int32(pC.nSize); i++  {
        // inlining of macro MSAT_LIT2VAR
        Var = ((pC.pData[i])>>1);
        if pAssigns[Var] == MSAT_VAR_UNASSIGNED {
            pC.pData[j++] = pC.pData[i];
            continue;
        }
        if pAssigns[Var] == pC.pData[i] {
            return 1; }
        // otherwise, the value of the literal is false
        // make sure, this literal is not watched
        assert(i >= 2, "i >= 2");
    }
    // if the size has changed, update it and move activity
    if j < int32(pC.nSize) {
        var Activ float32  = Msat_ClauseReadActivity(pC);
        pC.nSize = j;
        Msat_ClauseWriteActivity(pC, Activ);
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Computes reason of conflict in the given clause.]

  Description [If the literal is unassigned, finds the reason by 
  complementing literals in the given cluase (pC). If the literal is
  assigned, makes sure that this literal is the first one in the clause
  and computes the complement of all other literals in the clause.
  Returns the reason in the given array (vLits_out). If the clause is
  learned, bumps its activity.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseCalcReason(p *Msat_Solver_t, pC *Msat_Clause_t, Lit Msat_Lit_t, vLits_out *Msat_IntVec_t) {
    var i int32;
    // clear the reason
    Msat_IntVecClear( vLits_out );
    assert(Lit == MSAT_LIT_UNASSIGNED || Lit == pC.pData[0], "Lit == MSAT_LIT_UNASSIGNED || Lit == pC.pData[0]");
    for  i = (Lit != MSAT_LIT_UNASSIGNED); i < int32(pC.nSize); i++  {
        // inlining of macro MSAT_LIT2VAR
        // inlining of macro MSAT_LITNOT
        assert(Msat_SolverReadAssignsArray(p)[((pC.pData[i])>>1)] == ((pC.pData[i])^1), "Msat_SolverReadAssignsArray(p)[/*__starting_point_inline_macro:MSAT_LIT2VAR__*/((pC.pData[i])>>1)] == /*__starting_point_inline_macro:MSAT_LITNOT__*/((pC.pData[i])^1)");
        // inlining of macro MSAT_LITNOT
        Msat_IntVecPush( vLits_out, ((pC.pData[i])^1) );
    }
    if pC.fLearned { 
        Msat_SolverClaBumpActivity( p, pC ); }
}

/**Function*************************************************************

  Synopsis    [Removes the given clause from the watched list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseRemoveWatch(vClauses *Msat_ClauseVec_t, pC *Msat_Clause_t) {
    var pClauses **Msat_Clause_t;
    var nClauses, i int32;
    nClauses = Msat_ClauseVecReadSize( vClauses );
    pClauses = Msat_ClauseVecReadArray( vClauses );
    for  i = 0; pClauses[i] != pC; i++  {
        assert(i < nClauses, "i < nClauses");
    }
    for       ; i < nClauses - 1; i++  {
        pClauses[i] = pClauses[i+1];
    }
    Msat_ClauseVecPop( vClauses );
}

/**Function*************************************************************

  Synopsis    [Prints the given clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClausePrint(pC *Msat_Clause_t) {
    var i int32;
    if pC == nil {
        fmt.Printf("NULL pointer");
    } else {
        if pC.fLearned {
            fmt.Printf("Act = %.4f  ", Msat_ClauseReadActivity(pC)); }
        for  i = 0; i < int32(pC.nSize); i++  {
            fmt.Printf(" %s%d", ((pC.pData[i]&1)? "-": ""), pC.pData[i]/2 + 1);
        }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Writes the given clause in a file in DIMACS format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClauseWriteDimacs(pFile *os.File, pC *Msat_Clause_t, fIncrement int32) {
    var i int32;
    for  i = 0; i < int32(pC.nSize); i++  {
        fmt.Fprintf(pFile, "%s%d ", ((pC.pData[i]&1)? "-": ""), pC.pData[i]/2 + int32((fIncrement>0)));
    }
    if fIncrement != 0 {
        fmt.Fprintf(pFile, "0"); }
    fmt.Fprintf(pFile, "\n");
}

/**Function*************************************************************

  Synopsis    [Prints the given clause.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_ClausePrintSymbols(pC *Msat_Clause_t) {
    var i int32;
    if pC == nil {
        fmt.Printf("NULL pointer");
    } else {
//        if ( pC->fLearned )
//            printf( "Act = %.4f  ", Msat_ClauseReadActivity(pC) );
        for  i = 0; i < int32(pC.nSize); i++  {
            // inlining of macro L_lit
            // inlining of macro MSAT_LITSIGN
            // inlining of macro MSAT_LIT2VAR
            fmt.Printf(" "L_LIT, ((pC.pData[i])&1)     ?"-":"", ((pC.pData[i])>>1)+1);
        }
    }
    fmt.Printf("\n");
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


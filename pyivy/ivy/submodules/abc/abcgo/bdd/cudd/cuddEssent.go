package cudd

import "fmt"

import "os"

/**CFile***********************************************************************

  FileName    [cuddEssent.c]

  PackageName [cudd]

  Synopsis    [Functions for the detection of essential variables.]

  Description [External procedures included in this file:
                <ul>
                <li> Cudd_FindEssential()
                <li> Cudd_bddIsVarEssential()
                <li> Cudd_FindTwoLiteralClauses()
                <li> Cudd_ReadIthClause()
                <li> Cudd_PrintTwoLiteralClauses()
                <li> Cudd_tlcInfoFree()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> ddFindEssentialRecur()
                <li> ddFindTwoLiteralClausesRecur()
                <li> computeClauses()
                <li> computeClausesWithUniverse()
                <li> emptyClauseSet()
                <li> sentinelp()
                <li> equalp()
                <li> beforep()
                <li> oneliteralp()
                <li> impliedp()
                <li> bitVectorAlloc()
                <li> bitVectorClear()
                <li> bitVectorFree()
                <li> bitVectorRead()
                <li> bitVectorSet()
                <li> tlcInfoAlloc()
                </ul>]

  Author      [Fabio Somenzi]

  Copyright   [Copyright (c) 1995-2004, Regents of the University of Colorado

  All rights reserved.

  Redistribution and use in source and binary forms, with or without
  modification, are permitted provided that the following conditions
  are met:

  Redistributions of source code must retain the above copyright
  notice, this list of conditions and the following disclaimer.

  Redistributions in binary form must reproduce the above copyright
  notice, this list of conditions and the following disclaimer in the
  documentation and/or other materials provided with the distribution.

  Neither the name of the University of Colorado nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
  "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
  LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
  FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
  COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
  INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
  BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
  LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
  CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
  LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN
  ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
  POSSIBILITY OF SUCH DAMAGE.]

******************************************************************************/

//#include "misc/util/util_hack.h"
//#include "cuddInt.h"

//ABC_NAMESPACE_IMPL_START



/*---------------------------------------------------------------------------*/
/* Constant declarations                                                     */
/*---------------------------------------------------------------------------*/

/* These definitions are for the bit vectors. */
#if SIZEOF_LONG == 8
#define BPL 64
#define LOGBPL 6
#else
#define BPL 32
#define LOGBPL 5
#endif

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/* This structure holds the set of clauses for a node.  Each clause consists
** of two literals.  For one-literal clauses, the second lietral is FALSE.
** Each literal is composed of a variable and a phase.  A variable is a node
** index, and requires sizeof(DdHalfWord) bytes.  The constant literals use
** CUDD_MAXINDEX as variable indicator.  Each phase is a bit: 0 for positive
** phase, and 1 for negative phase.
** Variables and phases are stored separately for the sake of compactness.
** The variables are stored in an array of DdHalfWord's terminated by a
** sentinel (a pair of zeroes).  The phases are stored in a bit vector.
** The cnt field holds, at the end, the number of clauses.
** The clauses of the set are kept sorted.  For each clause, the first literal
** is the one of least index.  So, the clause with literals +2 and -4 is stored
** as (+2,-4).  A one-literal clause with literal +3 is stored as
** (+3,-CUDD_MAXINDEX).  Clauses are sorted in decreasing order as follows:
**      (+5,-7)
**      (+5,+6)
**      (-5,+7)
**      (-4,FALSE)
**      (-4,+8)
**      ...
** That is, one first looks at the variable of the first literal, then at the
** phase of the first litral, then at the variable of the second literal,
** and finally at the phase of the second literal.
*/
type DdTlcInfo struct {
    vars *DdHalfWord;
    phases *long;
    cnt DdHalfWord;
}

/* This structure is for temporary representation of sets of clauses.  It is
** meant to be used in link lists, when the number of clauses is not yet
** known. The encoding of a clause is the same as in DdTlcInfo, though
** the phase information is not stored in a bit array. */
type TlClause struct {
    DdHalfWord v1, v2;
    p1, p2 int16;
    next  *TlClause;
}

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

typedef long BitVector;
   

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddEssent.c,v 1.24 2009/02/21 18:24:10 fabio Exp $";
#endif

var Tolv *BitVector;
var Tolp *BitVector;
var Eolv *BitVector;
var Eolp *BitVector;

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static DdNode * ddFindEssentialRecur (DdManager *dd, DdNode *f);

// static DdTlcInfo * ddFindTwoLiteralClausesRecur (DdManager * dd, DdNode * f, st__table *table);

// static DdTlcInfo * computeClauses (DdTlcInfo *Tres, DdTlcInfo *Eres, DdHalfWord label, int size);

// static DdTlcInfo * computeClausesWithUniverse (DdTlcInfo *Cres, DdHalfWord label, short phase);

// static DdTlcInfo * emptyClauseSet (void);

// static int sentinelp (DdHalfWord var1, DdHalfWord var2);

// static int equalp (DdHalfWord var1a, short phase1a, DdHalfWord var1b, short phase1b, DdHalfWord var2a, short phase2a, DdHalfWord var2b, short phase2b);

// static int beforep (DdHalfWord var1a, short phase1a, DdHalfWord var1b, short phase1b, DdHalfWord var2a, short phase2a, DdHalfWord var2b, short phase2b);

// static int oneliteralp (DdHalfWord var);

// static int impliedp (DdHalfWord var1, short phase1, DdHalfWord var2, short phase2, BitVector *olv, BitVector *olp);

// static BitVector * bitVectorAlloc (int size);

DD_INLINE static void bitVectorClear (BitVector *vector, int size);
// static void bitVectorFree (BitVector *vector);

DD_INLINE static short bitVectorRead (BitVector *vector, int i);
DD_INLINE static void bitVectorSet (BitVector * vector, int i, short val);
// static DdTlcInfo * tlcInfoAlloc (void);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Finds the essential variables of a DD.]

  Description [Returns the cube of the essential variables. A positive
  literal means that the variable must be set to 1 for the function to be
  1. A negative literal means that the variable must be set to 0 for the
  function to be 1. Returns a pointer to the cube BDD if successful;
  NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIsVarEssential]

******************************************************************************/
func Cudd_FindEssential(dd *DdManager, f *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = ddFindEssentialRecur(dd,f);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_FindEssential */


/**Function********************************************************************

  Synopsis    [Determines whether a given variable is essential with a
  given phase in a BDD.]

  Description [Determines whether a given variable is essential with a
  given phase in a BDD. Uses Cudd_bddIteConstant. Returns 1 if phase == 1
  and f-->x_id, or if phase == 0 and f-->x_id'.]

  SideEffects [None]

  SeeAlso     [Cudd_FindEssential]

******************************************************************************/
func Cudd_bddIsVarEssential(manager *DdManager, f *DdNode, id int32, phase int32) int32 {
    var var *DdNode;
    var res int32;

    var = Cudd_bddIthVar(manager, id);

    // inlining of macro Cudd_NotCond
    var = ((DdNode *)((ptrint)(var) ^ (phase == 0)));

    res = Cudd_bddLeq(manager, f, var);

    return(res);

} /* end of Cudd_bddIsVarEssential */


/**Function********************************************************************

  Synopsis    [Finds the two literal clauses of a DD.]

  Description [Returns the one- and two-literal clauses of a DD.
  Returns a pointer to the structure holding the clauses if
  successful; NULL otherwise.  For a constant DD, the empty set of clauses
  is returned.  This is obviously correct for a non-zero constant.  For the
  constant zero, it is based on the assumption that only those clauses
  containing variables in the support of the function are considered.  Since
  the support of a constant function is empty, no clauses are returned.]

  SideEffects [None]

  SeeAlso     [Cudd_FindEssential]

******************************************************************************/
func Cudd_FindTwoLiteralClauses(dd *DdManager, f *DdNode) *DdTlcInfo {
    var res *DdTlcInfo;
    var table *st__table;
    var gen *st__generator;
    var tlc *DdTlcInfo;
    var node *DdNode;
    var size int32  = dd.size;

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {
        res = emptyClauseSet();
        return(res);
    }
    table = st__init_table( st__ptrcmp, st__ptrhash);
    if table == nil { return(nil); }
    Tolv = bitVectorAlloc(size);
    if Tolv == nil {
        st__free_table(table);
        return(nil);
    }
    Tolp = bitVectorAlloc(size);
    if Tolp == nil {
        st__free_table(table);
        bitVectorFree(Tolv);
        return(nil);
    }
    Eolv = bitVectorAlloc(size);
    if Eolv == nil {
        st__free_table(table);
        bitVectorFree(Tolv);
        bitVectorFree(Tolp);
        return(nil);
    }
    Eolp = bitVectorAlloc(size);
    if Eolp == nil {
        st__free_table(table);
        bitVectorFree(Tolv);
        bitVectorFree(Tolp);
        bitVectorFree(Eolv);
        return(nil);
    }

    res = ddFindTwoLiteralClausesRecur(dd,f,table);
    /* Dispose of table contents and free table. */
    // inlining of macro st__foreach_item
    for gen= st__init_gen(table); st__gen(gen,(char **)&node,(char **)&tlc) || ( st__free_gen(gen),0); {
        if node != f {
            Cudd_tlcInfoFree(tlc);
        }
    }
    st__free_table(table);
    bitVectorFree(Tolv);
    bitVectorFree(Tolp);
    bitVectorFree(Eolv);
    bitVectorFree(Eolp);

    if res != nil {
        var i int32;
        for i = 0; !(sentinelp(res.vars[i], res.vars[i+1]) != 0); i += 2 {}
        res.cnt = i >> 1;
    }

    return(res);

} /* end of Cudd_FindTwoLiteralClauses */


/**Function********************************************************************

  Synopsis    [Accesses the i-th clause of a DD.]

  Description [Accesses the i-th clause of a DD given the clause set which
  must be already computed.  Returns 1 if successful; 0 if i is out of range,
  or in case of error.]

  SideEffects [the four components of a clause are returned as side effects.]

  SeeAlso     [Cudd_FindTwoLiteralClauses]

******************************************************************************/
func Cudd_ReadIthClause(tlc *DdTlcInfo, i int32, var1 *DdHalfWord, var2 *DdHalfWord, phase1 *int32, phase2 *int32) int32 {
    if tlc == nil { return(0); }
    if tlc.vars == nil || tlc.phases == nil { return(0); }
    if i < 0 || uint32( i) >= tlc.cnt { return(0); }
    *var1 = tlc.vars[2*i];
    *var2 = tlc.vars[2*i+1];
    *phase1 = int32( bitVectorRead(tlc.phases, 2*i));
    *phase2 = int32( bitVectorRead(tlc.phases, 2*i+1));
    return(1);

} /* end of Cudd_ReadIthClause */


/**Function********************************************************************

  Synopsis    [Prints the two literal clauses of a DD.]

  Description [Prints the one- and two-literal clauses. Returns 1 if
  successful; 0 otherwise.  The argument "names" can be NULL, in which case
  the variable indices are printed.]

  SideEffects [None]

  SeeAlso     [Cudd_FindTwoLiteralClauses]

******************************************************************************/
func Cudd_PrintTwoLiteralClauses(dd *DdManager, f *DdNode, names *string, fp *os.File) int32 {
    var vars *DdHalfWord;
    var phases *BitVector;
    var i int32;
    var res *DdTlcInfo  = Cudd_FindTwoLiteralClauses(dd, f);
    var ifp *os.File  = fp == nil ? dd.out : fp;

    if res == nil { return(0); }
    vars = res.vars;
    phases = res.phases;
    for i = 0; !(sentinelp(vars[i], vars[i+1]) != 0); i += 2 {
        if names != nil {
            if vars[i+1] == CUDD_MAXINDEX {
                (void) fmt.Fprintf(ifp, "%s%s\n", bitVectorRead(phases, i) ? "~" : " ", names[vars[i]]);
            } else {
                (void) fmt.Fprintf(ifp, "%s%s | %s%s\n", bitVectorRead(phases, i) ? "~" : " ", names[vars[i]], bitVectorRead(phases, i+1) ? "~" : " ", names[vars[i+1]]);
            }
        } else {
            if vars[i+1] == CUDD_MAXINDEX {
                (void) fmt.Fprintf(ifp, "%s%d\n", bitVectorRead(phases, i) ? "~" : " ", int32( vars[i]));
            } else {
                (void) fmt.Fprintf(ifp, "%s%d | %s%d\n", bitVectorRead(phases, i) ? "~" : " ", int32( vars[i]), bitVectorRead(phases, i+1) ? "~" : " ", int32( vars[i+1]));
            }
        }
    }
    Cudd_tlcInfoFree(res);

    return(1);

} /* end of Cudd_PrintTwoLiteralClauses */


/**Function********************************************************************

  Synopsis    [Frees a DdTlcInfo Structure.]

  Description [Frees a DdTlcInfo Structure as well as the memory pointed
  by it.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_tlcInfoFree(t *DdTlcInfo) {
    if t.vars != nil { ABC_FREE(t.vars); }
    if t.phases != nil { ABC_FREE(t.phases); }
    ABC_FREE(t);

} /* end of Cudd_tlcInfoFree */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Implements the recursive step of Cudd_FindEssential.]

  Description [Implements the recursive step of Cudd_FindEssential.
  Returns a pointer to the cube BDD if successful; NULL otherwise.]

  SideEffects [None]

******************************************************************************/
func ddFindEssentialRecur(dd *DdManager, f *DdNode) *DdNode {
    var T, E, F *DdNode;
    var essT, essE, res *DdNode;
    var index int32;
    var one, lzero, azero *DdNode;

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    /* If f is constant the set of essential variables is empty. */
    // inlining of macro cuddIsConstant
    if (F).index == CUDD_CONST_INDEX { return(one); }

    res = cuddCacheLookup1(dd,Cudd_FindEssential,f);
    if res != nil {
        return(res);
    }

    // inlining of macro Cudd_Not
    lzero = ((DdNode *)((ptrint)(one) ^ 01));
    // inlining of macro DD_ZERO
    azero = ((dd).zero);
    /* Find cofactors: here f is non-constant. */
    // inlining of macro cuddT
    T = ((F).type.kids.T);
    // inlining of macro cuddE
    E = ((F).type.kids.E);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (f) & 01)) != 0 {
        // inlining of macro Cudd_Not
        T = ((DdNode *)((ptrint)(T) ^ 01)); E = ((DdNode *)((ptrint)(E) ^ 01));
    }

    index = F.index;
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if ((((DdNode *)((ptruint)(T) & ~01))).index == CUDD_CONST_INDEX) && T != lzero && T != azero {
        /* if E is zero, index is essential, otherwise there are no
        ** essentials, because index is not essential and no other variable
        ** can be, since setting index = 1 makes the function constant and
        ** different from 0.
        */
        if E == lzero || E == azero {
            res = dd.vars[index];
        } else {
            res = one;
        }
    } else if T == lzero || T == azero {
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        if (((DdNode *)((ptruint)(E) & ~01))).index == CUDD_CONST_INDEX { /* E cannot be zero here */
            // inlining of macro Cudd_Not
            res = ((DdNode *)((ptrint)(dd.vars[index]) ^ 01));
        } else { /* E == non-constant */
            /* find essentials in the else branch */
            essE = ddFindEssentialRecur(dd,E);
            if essE == nil {
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(essE).ref);

            /* add index to the set with negative phase */
            // inlining of macro Cudd_Not
            res = cuddUniqueInter(dd,index,one,((DdNode *)((ptrint)(essE) ^ 01)));
            if res == nil {
                Cudd_RecursiveDeref(dd,essE);
                return(nil);
            }
            // inlining of macro Cudd_Not
            res = ((DdNode *)((ptrint)(res) ^ 01));
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(essE).ref);
        }
    } else { /* T == non-const */
        if E == lzero || E == azero {
            /* find essentials in the then branch */
            essT = ddFindEssentialRecur(dd,T);
            if essT == nil {
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(essT).ref);

            /* add index to the set with positive phase */
            /* use And because essT may be complemented */
            res = cuddBddAndRecur(dd,dd.vars[index],essT);
            if res == nil {
                Cudd_RecursiveDeref(dd,essT);
                return(nil);
            }
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(essT).ref);
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        } else if !((((DdNode *)((ptruint)(E) & ~01))).index == CUDD_CONST_INDEX) {
            /* if E is a non-zero constant there are no essentials
            ** because T is non-constant.
            */
            essT = ddFindEssentialRecur(dd,T);
            if essT == nil {
                return(nil);
            }
            if essT == one {
                res = one;
            } else {
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(essT).ref);
                essE = ddFindEssentialRecur(dd,E);
                if essE == nil {
                    Cudd_RecursiveDeref(dd,essT);
                    return(nil);
                }
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(essE).ref);

                /* res = intersection(essT, essE) */
                res = cuddBddLiteralSetIntersectionRecur(dd,essT,essE);
                if res == nil {
                    Cudd_RecursiveDeref(dd,essT);
                    Cudd_RecursiveDeref(dd,essE);
                    return(nil);
                }
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(res).ref);
                Cudd_RecursiveDeref(dd,essT);
                Cudd_RecursiveDeref(dd,essE);
                // inlining of macro cuddDeref
                cuddSatDec(Cudd_Regular(res).ref);
            }
        } else {        /* E is a non-zero constant */
            res = one;
        }
    }

    cuddCacheInsert1(dd,Cudd_FindEssential, f, res);
    return(res);

} /* end of ddFindEssentialRecur */


/**Function********************************************************************

  Synopsis    [Implements the recursive step of Cudd_FindTwoLiteralClauses.]

  Description [Implements the recursive step of
  Cudd_FindTwoLiteralClauses.  The DD node is assumed to be not
  constant.  Returns a pointer to a set of clauses if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_FindTwoLiteralClauses]

******************************************************************************/
func ddFindTwoLiteralClausesRecur(dd *DdManager, f *DdNode, table *st__table) *DdTlcInfo {
    var T, E, F *DdNode;
    var one, lzero, azero *DdNode;
    var res, Tres, Eres *DdTlcInfo;
    var index DdHalfWord;

    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));

    // inlining of macro cuddIsConstant
    assert(!((F).index == CUDD_CONST_INDEX), "!/*__starting_point_inline_macro:cuddIsConstant__*/((F).index == CUDD_CONST_INDEX)");

    /* Check computed table.  Separate entries are necessary for
    ** a node and its complement.  We should update the counter here. */
    if st__lookup(table, (char *)f, (char **)&res) {
        return(res);
    }

    /* Easy access to the constants for BDDs and ADDs. */
    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    lzero = ((DdNode *)((ptrint)(one) ^ 01));
    // inlining of macro DD_ZERO
    azero = ((dd).zero);

    /* Find cofactors and variable labeling the top node. */
    // inlining of macro cuddT
    // inlining of macro cuddE
    T = ((F).type.kids.T); E = ((F).type.kids.E);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (f) & 01)) != 0 {
        // inlining of macro Cudd_Not
        T = ((DdNode *)((ptrint)(T) ^ 01)); E = ((DdNode *)((ptrint)(E) ^ 01));
    }
    index = F.index;

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if ((((DdNode *)((ptruint)(T) & ~01))).index == CUDD_CONST_INDEX) && T != lzero && T != azero {
        /* T is a non-zero constant.  If E is zero, then this node's index
        ** is a one-literal clause.  Otherwise, if E is a non-zero
        ** constant, there are no clauses for this node.  Finally,
        ** if E is not constant, we recursively compute its clauses, and then
        ** merge using the empty set for T. */
        if E == lzero || E == azero {
            /* Create the clause (index + 0). */
            res = tlcInfoAlloc();
            if res == nil { return(nil); }
            res.vars = ABC_ALLOC(DdHalfWord,4);
            if res.vars == nil {
                ABC_FREE(res);
                return(nil);
            }
            res.phases = bitVectorAlloc(2);
            if res.phases == nil {
                ABC_FREE(res.vars);
                ABC_FREE(res);
                return(nil);
            }
            res.vars[0] = index;
            res.vars[1] = CUDD_MAXINDEX;
            res.vars[2] = 0;
            res.vars[3] = 0;
            bitVectorSet(res.phases, 0, 0); /* positive phase */
            bitVectorSet(res.phases, 1, 1); /* negative phase */
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        } else if (((DdNode *)((ptruint)(E) & ~01))).index == CUDD_CONST_INDEX {
            /* If E is a non-zero constant, no clauses. */
            res = emptyClauseSet();
        } else {
            /* E is non-constant */
            Tres = emptyClauseSet();
            if Tres == nil { return(nil); }
            Eres = ddFindTwoLiteralClausesRecur(dd, E, table);
            if Eres == nil {
                Cudd_tlcInfoFree(Tres);
                return(nil);
            }
            res = computeClauses(Tres, Eres, index, dd.size);
            Cudd_tlcInfoFree(Tres);
        }
    } else if T == lzero || T == azero {
        /* T is zero.  If E is a non-zero constant, then the
        ** complement of this node's index is a one-literal clause.
        ** Otherwise, if E is not constant, we recursively compute its
        ** clauses, and then merge using the universal set for T. */
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        if (((DdNode *)((ptruint)(E) & ~01))).index == CUDD_CONST_INDEX { /* E cannot be zero here */
            /* Create the clause (!index + 0). */
            res = tlcInfoAlloc();
            if res == nil { return(nil); }
            res.vars = ABC_ALLOC(DdHalfWord,4);
            if res.vars == nil {
                ABC_FREE(res);
                return(nil);
            }
            res.phases = bitVectorAlloc(2);
            if res.phases == nil {
                ABC_FREE(res.vars);
                ABC_FREE(res);
                return(nil);
            }
            res.vars[0] = index;
            res.vars[1] = CUDD_MAXINDEX;
            res.vars[2] = 0;
            res.vars[3] = 0;
            bitVectorSet(res.phases, 0, 1); /* negative phase */
            bitVectorSet(res.phases, 1, 1); /* negative phase */
        } else { /* E == non-constant */
            Eres = ddFindTwoLiteralClausesRecur(dd, E, table);
            if Eres == nil { return(nil); }
            res = computeClausesWithUniverse(Eres, index, 1);
        }
    } else { /* T == non-const */
        Tres = ddFindTwoLiteralClausesRecur(dd, T, table);
        if Tres == nil { return(nil); }
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        if (((DdNode *)((ptruint)(E) & ~01))).index == CUDD_CONST_INDEX {
            if E == lzero || E == azero {
                res = computeClausesWithUniverse(Tres, index, 0);
            } else {
                Eres = emptyClauseSet();
                if Eres == nil { return(nil); }
                res = computeClauses(Tres, Eres, index, dd.size);
                Cudd_tlcInfoFree(Eres);
            }
        } else {
            Eres = ddFindTwoLiteralClausesRecur(dd, E, table);
            if Eres == nil { return(nil); }
            res = computeClauses(Tres, Eres, index, dd.size);
        }
    }

    /* Cache results. */
    if st__add_direct(table, (char *)f, (char *)res) == st__OUT_OF_MEM {
        ABC_FREE(res);
        return(nil);
    }
    return(res);

} /* end of ddFindTwoLiteralClausesRecur */


/**Function********************************************************************

  Synopsis    [Computes the two-literal clauses for a node.]

  Description [Computes the two-literal clauses for a node given the
  clauses for its children and the label of the node.  Returns a
  pointer to a TclInfo structure if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [computeClausesWithUniverse]

******************************************************************************/
static DdTlcInfo *
computeClauses(
  DdTlcInfo *Tres /* list of clauses for T child */,
  DdTlcInfo *Eres /* list of clauses for E child */,
  DdHalfWord label /* variable labeling the current node */,
  int size /* number of variables in the manager */)
{
    var Tcv *DdHalfWord  = Tres.vars; /* variables of clauses for the T child */
    var Tcp *BitVector  = Tres.phases; /* phases of clauses for the T child */
    var Ecv *DdHalfWord  = Eres.vars; /* variables of clauses for the E child */
    var Ecp *BitVector  = Eres.phases; /* phases of clauses for the E child */
    var Vcv *DdHalfWord  = nil; /* pointer to variables of the clauses for v */
    var Vcp *BitVector  = nil; /* pointer to phases of the clauses for v */
    var res *DdTlcInfo  = nil; /* the set of clauses to be returned */
    var pt int32; /* index in the list of clauses of T */
    var pe int32; /* index in the list of clauses of E */
    var cv int32; /* counter of the clauses for this node */
    var iclauses *TlClause  = nil; /* list of inherited clauses */
    var tclauses *TlClause  = nil; /* list of 1-literal clauses of T */
    var eclauses *TlClause  = nil; /* list of 1-literal clauses of E */
    var nclauses *TlClause  = nil; /* list of new (non-inherited) clauses */
    var lnclause *TlClause  = nil; /* pointer to last new clause */
    var newclause *TlClause; /* temporary pointer to new clauses */

    /* Initialize sets of one-literal clauses.  The one-literal clauses
    ** are stored redundantly.  These sets allow constant-time lookup, which
    ** we need when we check for implication of a two-literal clause by a
    ** one-literal clause.  The linked lists allow fast sequential
    ** processing. */
    bitVectorClear(Tolv, size);
    bitVectorClear(Tolp, size);
    bitVectorClear(Eolv, size);
    bitVectorClear(Eolp, size);

    /* Initialize result structure. */
    res = tlcInfoAlloc();
    if res == nil { goto cleanup; }

    /* Scan the two input list.  Extract inherited two-literal clauses
    ** and set aside one-literal clauses from each list.  The incoming lists
    ** are sorted in the order defined by beforep.  The three linked list
    ** produced by this loop are sorted in the reverse order because we
    ** always append to the front of the lists.
    ** The inherited clauses are those clauses (both one- and two-literal)
    ** that are common to both children; and the two-literal clauses of
    ** one child that are implied by a one-literal clause of the other
    ** child. */
    for !(sentinelp(Tcv[pt], Tcv[pt+1]) != 0) || !(sentinelp(Ecv[pe], Ecv[pe+1]) != 0) {
        if equalp(Tcv[pt], bitVectorRead(Tcp, pt),
                   Tcv[pt+1], bitVectorRead(Tcp, pt+1),
                   Ecv[pe], bitVectorRead(Ecp, pe),
                   Ecv[pe+1], bitVectorRead(Ecp, pe+1)) != 0 {
            /* Add clause to inherited list. */
            newclause = new(TlClause);
            if newclause == nil { goto cleanup; }
            newclause.v1 = Tcv[pt];
            newclause.v2 = Tcv[pt+1];
            newclause.p1 = bitVectorRead(Tcp, pt);
            newclause.p2 = bitVectorRead(Tcp, pt+1);
            newclause.next = iclauses;
            iclauses = newclause;
            pt += 2; pe += 2; cv++;
        } else if beforep(Tcv[pt], bitVectorRead(Tcp, pt),
                   Tcv[pt+1], bitVectorRead(Tcp, pt+1),
                   Ecv[pe], bitVectorRead(Ecp, pe),
                   Ecv[pe+1], bitVectorRead(Ecp, pe+1)) != 0 {
            if oneliteralp(Tcv[pt+1]) != 0 {
                /* Add this one-literal clause to the T set. */
                newclause = new(TlClause);
                if newclause == nil { goto cleanup; }
                newclause.v1 = Tcv[pt];
                newclause.v2 = CUDD_MAXINDEX;
                newclause.p1 = bitVectorRead(Tcp, pt);
                newclause.p2 = 1;
                newclause.next = tclauses;
                tclauses = newclause;
                bitVectorSet(Tolv, Tcv[pt], 1);
                bitVectorSet(Tolp, Tcv[pt], bitVectorRead(Tcp, pt));
            } else {
                if impliedp(Tcv[pt], bitVectorRead(Tcp, pt),
                             Tcv[pt+1], bitVectorRead(Tcp, pt+1),
                             Eolv, Eolp) != 0 {
                    /* Add clause to inherited list. */
                    newclause = new(TlClause);
                    if newclause == nil { goto cleanup; }
                    newclause.v1 = Tcv[pt];
                    newclause.v2 = Tcv[pt+1];
                    newclause.p1 = bitVectorRead(Tcp, pt);
                    newclause.p2 = bitVectorRead(Tcp, pt+1);
                    newclause.next = iclauses;
                    iclauses = newclause;
                    cv++;
                }
            }
            pt += 2;
        } else { /* !beforep() */
            if oneliteralp(Ecv[pe+1]) != 0 {
                /* Add this one-literal clause to the E set. */
                newclause = new(TlClause);
                if newclause == nil { goto cleanup; }
                newclause.v1 = Ecv[pe];
                newclause.v2 = CUDD_MAXINDEX;
                newclause.p1 = bitVectorRead(Ecp, pe);
                newclause.p2 = 1;
                newclause.next = eclauses;
                eclauses = newclause;
                bitVectorSet(Eolv, Ecv[pe], 1);
                bitVectorSet(Eolp, Ecv[pe], bitVectorRead(Ecp, pe));
            } else {
                if impliedp(Ecv[pe], bitVectorRead(Ecp, pe),
                             Ecv[pe+1], bitVectorRead(Ecp, pe+1),
                             Tolv, Tolp) != 0 {
                    /* Add clause to inherited list. */
                    newclause = new(TlClause);
                    if newclause == nil { goto cleanup; }
                    newclause.v1 = Ecv[pe];
                    newclause.v2 = Ecv[pe+1];
                    newclause.p1 = bitVectorRead(Ecp, pe);
                    newclause.p2 = bitVectorRead(Ecp, pe+1);
                    newclause.next = iclauses;
                    iclauses = newclause;
                    cv++;
                }
            }
            pe += 2;
        }
    }

    /* Add one-literal clauses for the label variable to the front of
    ** the two lists. */
    newclause = new(TlClause);
    if newclause == nil { goto cleanup; }
    newclause.v1 = label;
    newclause.v2 = CUDD_MAXINDEX;
    newclause.p1 = 0;
    newclause.p2 = 1;
    newclause.next = tclauses;
    tclauses = newclause;
    newclause = new(TlClause);
    if newclause == nil { goto cleanup; }
    newclause.v1 = label;
    newclause.v2 = CUDD_MAXINDEX;
    newclause.p1 = 1;
    newclause.p2 = 1;
    newclause.next = eclauses;
    eclauses = newclause;

    /* Produce the non-inherited clauses.  We preserve the "reverse"
    ** order of the two input lists by appending to the end of the
    ** list.  In this way, iclauses and nclauses are consistent. */
    for tclauses != nil && eclauses != nil {
        if beforep(eclauses.v1, eclauses.p1, eclauses.v2, eclauses.p2,
                    tclauses.v1, tclauses.p1, tclauses.v2, tclauses.p2) != 0 {
            var nextclause *TlClause  = tclauses.next;
            var otherclauses *TlClause  = eclauses;
            for otherclauses != nil {
                if tclauses.v1 != otherclauses.v1 {
                    newclause = new(TlClause);
                    if newclause == nil { goto cleanup; }
                    newclause.v1 = tclauses.v1;
                    newclause.v2 = otherclauses.v1;
                    newclause.p1 = tclauses.p1;
                    newclause.p2 = otherclauses.p1;
                    newclause.next = nil;
                    if nclauses == nil {
                        nclauses = newclause;
                        lnclause = newclause;
                    } else {
                        lnclause.next = newclause;
                        lnclause = newclause;
                    }
                    cv++;
                }
                otherclauses = otherclauses.next;
            }
            ABC_FREE(tclauses);
            tclauses = nextclause;
        } else {
            var nextclause *TlClause  = eclauses.next;
            var otherclauses *TlClause  = tclauses;
            for otherclauses != nil {
                if eclauses.v1 != otherclauses.v1 {
                    newclause = new(TlClause);
                    if newclause == nil { goto cleanup; }
                    newclause.v1 = eclauses.v1;
                    newclause.v2 = otherclauses.v1;
                    newclause.p1 = eclauses.p1;
                    newclause.p2 = otherclauses.p1;
                    newclause.next = nil;
                    if nclauses == nil {
                        nclauses = newclause;
                        lnclause = newclause;
                    } else {
                        lnclause.next = newclause;
                        lnclause = newclause;
                    }
                    cv++;
                }
                otherclauses = otherclauses.next;
            }
            ABC_FREE(eclauses);
            eclauses = nextclause;
        }
    }
    for tclauses != nil {
        var nextclause *TlClause  = tclauses.next;
        ABC_FREE(tclauses);
        tclauses = nextclause;
    }
    for eclauses != nil {
        var nextclause *TlClause  = eclauses.next;
        ABC_FREE(eclauses);
        eclauses = nextclause;
    }

    /* Merge inherited and non-inherited clauses.  Now that we know the
    ** total number, we allocate the arrays, and we fill them bottom-up
    ** to restore the proper ordering. */
    Vcv = ABC_ALLOC(DdHalfWord, 2*(cv+1));
    if Vcv == nil { goto cleanup; }
    if cv > 0 {
        Vcp = bitVectorAlloc(2*cv);
        if Vcp == nil { goto cleanup; }
    } else {
        Vcp = nil;
    }
    res.vars = Vcv;
    res.phases = Vcp;
    /* Add sentinel. */
    Vcv[2*cv] = 0;
    Vcv[2*cv+1] = 0;
    for iclauses != nil || nclauses != nil {
        var nextclause *TlClause;
        cv--;
        if nclauses == nil || (iclauses != nil &&
            (beforep(nclauses.v1, nclauses.p1, nclauses.v2, nclauses.p2,
                    iclauses.v1, iclauses.p1, iclauses.v2, iclauses.p2) != 0)) {
            Vcv[2*cv] = iclauses.v1;
            Vcv[2*cv+1] = iclauses.v2;
            bitVectorSet(Vcp, 2*cv, iclauses.p1);
            bitVectorSet(Vcp, 2*cv+1, iclauses.p2);
            nextclause = iclauses.next;
            ABC_FREE(iclauses);
            iclauses = nextclause;
        } else {
            Vcv[2*cv] = nclauses.v1;
            Vcv[2*cv+1] = nclauses.v2;
            bitVectorSet(Vcp, 2*cv, nclauses.p1);
            bitVectorSet(Vcp, 2*cv+1, nclauses.p2);
            nextclause = nclauses.next;
            ABC_FREE(nclauses);
            nclauses = nextclause;
        }
    }
    assert(cv == 0, "cv == 0");

    return(res);

 cleanup:
    if res != nil { Cudd_tlcInfoFree(res); }
    for iclauses != nil {
        var nextclause *TlClause  = iclauses.next;
        ABC_FREE(iclauses);
        iclauses = nextclause;
    }
    for nclauses != nil {
        var nextclause *TlClause  = nclauses.next;
        ABC_FREE(nclauses);
        nclauses = nextclause;
    }
    for tclauses != nil {
        var nextclause *TlClause  = tclauses.next;
        ABC_FREE(tclauses);
        tclauses = nextclause;
    }
    for eclauses != nil {
        var nextclause *TlClause  = eclauses.next;
        ABC_FREE(eclauses);
        eclauses = nextclause;
    }

    return(nil);

} /* end of computeClauses */


/**Function********************************************************************

  Synopsis    [Computes the two-literal clauses for a node.]

  Description [Computes the two-literal clauses for a node with a zero
  child, given the clauses for its other child and the label of the
  node.  Returns a pointer to a TclInfo structure if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [computeClauses]

******************************************************************************/
static DdTlcInfo *
computeClausesWithUniverse(
  DdTlcInfo *Cres /* list of clauses for child */,
  DdHalfWord label /* variable labeling the current node */,
  short phase /* 0 if E child is zero; 1 if T child is zero */)
{
    var Ccv *DdHalfWord  = Cres.vars; /* variables of clauses for child */
    var Ccp *BitVector  = Cres.phases; /* phases of clauses for child */
    var Vcv *DdHalfWord  = nil; /* pointer to the variables of the clauses for v */
    var Vcp *BitVector  = nil; /* pointer to the phases of the clauses for v */
    var res *DdTlcInfo  = nil; /* the set of clauses to be returned */
    var i int32;

    /* Initialize result. */
    res = tlcInfoAlloc();
    if res == nil { goto cleanup; }
    /* Count entries for new list and allocate accordingly. */
    for i = 0; !(sentinelp(Ccv[i], Ccv[i+1]) != 0); i += 2 {}
    /* At this point, i is twice the number of clauses in the child's
    ** list.  We need four more entries for this node: 2 for the one-literal
    ** clause for the label, and 2 for the sentinel. */
    Vcv = ABC_ALLOC(DdHalfWord,i+4);
    if Vcv == nil { goto cleanup; }
    Vcp = bitVectorAlloc(i+4);
    if Vcp == nil { goto cleanup; }
    res.vars = Vcv;
    res.phases = Vcp;
    /* Copy old list into new. */
    for i = 0; !(sentinelp(Ccv[i], Ccv[i+1]) != 0); i += 2 {
        Vcv[i] = Ccv[i];
        Vcv[i+1] = Ccv[i+1];
        bitVectorSet(Vcp, i, bitVectorRead(Ccp, i));
        bitVectorSet(Vcp, i+1, bitVectorRead(Ccp, i+1));
    }
    /* Add clause corresponding to label. */
    Vcv[i] = label;
    bitVectorSet(Vcp, i, phase);
    i++;
    Vcv[i] = CUDD_MAXINDEX;
    bitVectorSet(Vcp, i, 1);
    i++;
    /* Add sentinel. */
    Vcv[i] = 0;
    Vcv[i+1] = 0;
    bitVectorSet(Vcp, i, 0);
    bitVectorSet(Vcp, i+1, 0);

    return(res);

 cleanup:
    /* Vcp is guaranteed to be NULL here.  Hence, we do not try to free it. */
    if Vcv != nil { ABC_FREE(Vcv); }
    if res != nil { Cudd_tlcInfoFree(res); }

    return(nil);

} /* end of computeClausesWithUniverse */


/**Function********************************************************************

  Synopsis    [Returns an enpty set of clauses.]

  Description [Returns a pointer to an empty set of clauses if
  successful; NULL otherwise.  No bit vector for the phases is
  allocated.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func emptyClauseSet() *DdTlcInfo {
    var eset *DdTlcInfo;

    eset = new(DdTlcInfo);
    if eset == nil { return(nil); }
    eset.vars = ABC_ALLOC(DdHalfWord,2);
    if eset.vars == nil {
        ABC_FREE(eset);
        return(nil);
    }
    /* Sentinel */
    eset.vars[0] = 0;
    eset.vars[1] = 0;
    eset.phases = nil; /* does not matter */
    eset.cnt = 0;
    return(eset);

} /* end of emptyClauseSet */


/**Function********************************************************************

  Synopsis    [Returns true iff the argument is the sentinel clause.]

  Description [Returns true iff the argument is the sentinel clause.
  A sentinel clause has both variables equal to 0.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func sentinelp(var1 DdHalfWord, var2 DdHalfWord) int32 {
    return(var1 == 0 && var2 == 0);

} /* end of sentinelp */


/**Function********************************************************************

  Synopsis    [Returns true iff the two arguments are identical clauses.]

  Description [Returns true iff the two arguments are identical
  clauses.  Since literals are sorted, we only need to compare
  literals in the same position.]

  SideEffects [None]

  SeeAlso     [beforep]

******************************************************************************/
func equalp(var1a DdHalfWord, phase1a int16, var1b DdHalfWord, phase1b int16, var2a DdHalfWord, phase2a int16, var2b DdHalfWord, phase2b int16) int32 {
    return(var1a == var2a && phase1a == phase2a &&
           var1b == var2b && phase1b == phase2b);

} /* end of equalp */


/**Function********************************************************************

  Synopsis    [Returns true iff the first argument precedes the second in
  the clause order.]

  Description [Returns true iff the first argument precedes the second
  in the clause order.  A clause precedes another if its first lieral
  precedes the first literal of the other, or if the first literals
  are the same, and its second literal precedes the second literal of
  the other clause.  A literal precedes another if it has a higher
  index, of if it has the same index, but it has lower phase.  Phase 0
  is the positive phase, and it is lower than Phase 1 (negative
  phase).]

  SideEffects [None]

  SeeAlso     [equalp]

******************************************************************************/
func beforep(var1a DdHalfWord, phase1a int16, var1b DdHalfWord, phase1b int16, var2a DdHalfWord, phase2a int16, var2b DdHalfWord, phase2b int16) int32 {
    return(var1a > var2a || (var1a == var2a &&
           (phase1a < phase2a || (phase1a == phase2a &&
            (var1b > var2b || (var1b == var2b && phase1b < phase2b))))));

} /* end of beforep */


/**Function********************************************************************

  Synopsis    [Returns true iff the argument is a one-literal clause.]

  Description [Returns true iff the argument is a one-literal clause.
  A one-litaral clause has the constant FALSE as second literal.
  Since the constant TRUE is never used, it is sufficient to test for
  a constant.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func oneliteralp(var DdHalfWord) int32 {
    return(var == CUDD_MAXINDEX);

} /* end of oneliteralp */


/**Function********************************************************************

  Synopsis [Returns true iff either literal of a clause is in a set of
  literals.]

  Description [Returns true iff either literal of a clause is in a set
  of literals.  The first four arguments specify the clause.  The
  remaining two arguments specify the literal set.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func impliedp(var1 DdHalfWord, phase1 int16, var2 DdHalfWord, phase2 int16, olv *BitVector, olp *BitVector) int32 {
    return((bitVectorRead(olv, var1) &&
            bitVectorRead(olp, var1) == phase1) ||
           (bitVectorRead(olv, var2) &&
            bitVectorRead(olp, var2) == phase2));

} /* end of impliedp */


/**Function********************************************************************

  Synopsis    [Allocates a bit vector.]

  Description [Allocates a bit vector.  The parameter size gives the
  number of bits.  This procedure allocates enough long's to hold the
  specified number of bits.  Returns a pointer to the allocated vector
  if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [bitVectorClear bitVectorFree]

******************************************************************************/
func bitVectorAlloc(size int32) *BitVector {
    var allocSize int32;
    var vector *BitVector;

    /* Find out how many long's we need.
    ** There are sizeof(long) * 8 bits in a long.
    ** The ceiling of the ratio of two integers m and n is given
    ** by ((n-1)/m)+1.  Putting all this together, we get... */
    allocSize = ((size - 1) / (sizeof(BitVector) * 8)) + 1;
    vector = ABC_ALLOC(BitVector, allocSize);
    if vector == nil { return(nil); }
    /* Clear the whole array. */
    (void) memset(vector, 0, allocSize * sizeof(BitVector));
    return(vector);

} /* end of bitVectorAlloc */


/**Function********************************************************************

  Synopsis    [Clears a bit vector.]

  Description [Clears a bit vector.  The parameter size gives the
  number of bits.]

  SideEffects [None]

  SeeAlso     [bitVectorAlloc]

******************************************************************************/
//DD_INLINE
func bitVectorClear(vector *BitVector, size int32) {
    var allocSize int32;

    /* Find out how many long's we need.
    ** There are sizeof(long) * 8 bits in a long.
    ** The ceiling of the ratio of two integers m and n is given
    ** by ((n-1)/m)+1.  Putting all this together, we get... */
    allocSize = ((size - 1) / (sizeof(BitVector) * 8)) + 1;
    /* Clear the whole array. */
    (void) memset(vector, 0, allocSize * sizeof(BitVector));
    return;

} /* end of bitVectorClear */


/**Function********************************************************************

  Synopsis    [Frees a bit vector.]

  Description [Frees a bit vector.]

  SideEffects [None]

  SeeAlso     [bitVectorAlloc]

******************************************************************************/
func bitVectorFree(vector *BitVector) {
    ABC_FREE(vector);

} /* end of bitVectorFree */


/**Function********************************************************************

  Synopsis    [Returns the i-th entry of a bit vector.]

  Description [Returns the i-th entry of a bit vector.]

  SideEffects [None]

  SeeAlso     [bitVectorSet]

******************************************************************************/
//DD_INLINE
func bitVectorRead(vector *BitVector, i int32) int16 {
    var word, bit int32;
    var result int16;

    if vector == nil { return(int16( 0)); }

    word = i >> LOGBPL;
    bit = i & (BPL - 1);
    result = int16( ((vector[word] >> bit) & int64(1)));
    return(result);

} /* end of bitVectorRead */


/**Function********************************************************************

  Synopsis    [Sets the i-th entry of a bit vector to a value.]

  Description [Sets the i-th entry of a bit vector to a value.]

  SideEffects [None]

  SeeAlso     [bitVectorRead]

******************************************************************************/
//DD_INLINE
func bitVectorSet(vector *BitVector, i int32, val int16) {
    var word, bit int32;

    word = i >> LOGBPL;
    bit = i & (BPL - 1);
    vector[word] &= ~(int64(1) << bit);
    vector[word] |= ((int64( val)) << bit);

} /* end of bitVectorSet */


/**Function********************************************************************

  Synopsis    [Allocates a DdTlcInfo Structure.]

  Description [Returns a pointer to a DdTlcInfo Structure if successful;
  NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_tlcInfoFree]

******************************************************************************/
func tlcInfoAlloc() *DdTlcInfo {
    var res *DdTlcInfo  = new(DdTlcInfo);
    if res == nil { return(nil); }
    res.vars = nil;
    res.phases = nil;
    res.cnt = 0;
    return(res);

} /* end of tlcInfoAlloc */


//ABC_NAMESPACE_IMPL_END


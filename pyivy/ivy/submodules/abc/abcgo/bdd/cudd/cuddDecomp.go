package cudd

import (
	"fmt"
	"math"
)

import "os"

/**CFile***********************************************************************

  FileName    [cuddDecomp.c]

  PackageName [cudd]

  Synopsis    [Functions for BDD decomposition.]

  Description [External procedures included in this file:
                <ul>
                <li> Cudd_bddApproxConjDecomp()
                <li> Cudd_bddApproxDisjDecomp()
                <li> Cudd_bddIterConjDecomp()
                <li> Cudd_bddIterDisjDecomp()
                <li> Cudd_bddGenConjDecomp()
                <li> Cudd_bddGenDisjDecomp()
                <li> Cudd_bddVarConjDecomp()
                <li> Cudd_bddVarDisjDecomp()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> cuddConjunctsAux()
                <li> CreateBotDist()
                <li> BuildConjuncts()
                <li> ConjunctsFree()
                </ul>]

  Author      [Kavita Ravi, Fabio Somenzi]

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
const DEPTH = 5
const THRESHOLD = 10
const NONE = 0
const PAIR_ST = 1
const PAIR_CR = 2
const G_ST = 3
const G_CR = 4
const H_ST = 5
const H_CR = 6
const BOTH_G = 7
const BOTH_H = 8

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/
typedef struct Conjuncts {
    DdNode *g;
    DdNode *h;
}
 Conjuncts;

typedef struct  NodeStat {
    int distance;
    int localRef;
}
 NodeStat;


/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddDecomp.c,v 1.44 2004/08/13 18:04:47 fabio Exp $";
#endif

var one, zero *DdNode;
var lastTimeG long;

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


// Original macro:
// #define FactorsNotStored(factors)  ((int)((long)(factors) & 01))


// Original macro:
// #define FactorsComplement(factors) ((Conjuncts *)((long)(factors) | 01))


// Original macro:
// #define FactorsUncomplement(factors) ((Conjuncts *)((long)(factors) ^ 01))


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static NodeStat * CreateBotDist (DdNode * node, st__table * distanceTable);

// static double CountMinterms (DdNode * node, double max, st__table * mintermTable, FILE *fp);

// static void ConjunctsFree (DdManager * dd, Conjuncts * factors);

// static int PairInTables (DdNode * g, DdNode * h, st__table * ghTable);

// static Conjuncts * CheckTablesCacheAndReturn (DdNode * node, DdNode * g, DdNode * h, st__table * ghTable, st__table * cacheTable);

// static Conjuncts * PickOnePair (DdNode * node, DdNode * g1, DdNode * h1, DdNode * g2, DdNode * h2, st__table * ghTable, st__table * cacheTable);

// static Conjuncts * CheckInTables (DdNode * node, DdNode * g1, DdNode * h1, DdNode * g2, DdNode * h2, st__table * ghTable, st__table * cacheTable, int * outOfMem);

// static Conjuncts * ZeroCase (DdManager * dd, DdNode * node, Conjuncts * factorsNv, st__table * ghTable, st__table * cacheTable, int switched);

// static Conjuncts * BuildConjuncts (DdManager * dd, DdNode * node, st__table * distanceTable, st__table * cacheTable, int approxDistance, int maxLocalRef, st__table * ghTable, st__table * mintermTable);

// static int cuddConjunctsAux (DdManager * dd, DdNode * f, DdNode ** c1, DdNode ** c2);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs two-way conjunctive decomposition of a BDD.]

  Description [Performs two-way conjunctive decomposition of a
  BDD. This procedure owes its name to the use of supersetting to
  obtain an initial factor of the given function. Returns the number
  of conjuncts produced, that is, 2 if successful; 1 if no meaningful
  decomposition was found; 0 otherwise. The conjuncts produced by this
  procedure tend to be imbalanced.]

  SideEffects [The factors are returned in an array as side effects.
  The array is allocated by this function. It is the caller's responsibility
  to free it. On successful completion, the conjuncts are already
  referenced. If the function returns 0, the array for the conjuncts is
  not allocated. If the function returns 1, the only factor equals the
  function to be decomposed.]

  SeeAlso     [Cudd_bddApproxDisjDecomp Cudd_bddIterConjDecomp
  Cudd_bddGenConjDecomp Cudd_bddVarConjDecomp Cudd_RemapOverApprox
  Cudd_bddSqueeze Cudd_bddLICompaction]

******************************************************************************/
int
Cudd_bddApproxConjDecomp(
  DdManager * dd /* manager */,
  DdNode * f /* function to be decomposed */,
  DdNode *** conjuncts /* address of the first factor */)
{
    var superset1, superset2, glocal, hlocal *DdNode;
    var nvars int32  = Cudd_SupportSize(dd,f);

    /* Find a tentative first factor by overapproximation and minimization. */
    superset1 = Cudd_RemapOverApprox(dd,f,nvars,0,1.0);
    if superset1 == nil { return(0); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(superset1).ref);
    superset2 = Cudd_bddSqueeze(dd,f,superset1);
    if superset2 == nil {
        Cudd_RecursiveDeref(dd,superset1);
        return(0);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(superset2).ref);
    Cudd_RecursiveDeref(dd,superset1);

    /* Compute the second factor by minimization. */
    hlocal = Cudd_bddLICompaction(dd,f,superset2);
    if hlocal == nil {
        Cudd_RecursiveDeref(dd,superset2);
        return(0);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(hlocal).ref);

    /* Refine the first factor by minimization. If h turns out to be f, this
    ** step guarantees that g will be 1. */
    glocal = Cudd_bddLICompaction(dd,superset2,hlocal);
    if glocal == nil {
        Cudd_RecursiveDeref(dd,superset2);
        Cudd_RecursiveDeref(dd,hlocal);
        return(0);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(glocal).ref);
    Cudd_RecursiveDeref(dd,superset2);

    // inlining of macro DD_ONE
    if glocal != ((dd).one) {
        // inlining of macro DD_ONE
        if hlocal != ((dd).one) {
            *conjuncts = ABC_ALLOC(DdNode *,2);
            if *conjuncts == nil {
                Cudd_RecursiveDeref(dd,glocal);
                Cudd_RecursiveDeref(dd,hlocal);
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
            (*conjuncts)[0] = glocal;
            (*conjuncts)[1] = hlocal;
            return(2);
        } else {
            Cudd_RecursiveDeref(dd,hlocal);
            *conjuncts = ABC_ALLOC(DdNode *,1);
            if *conjuncts == nil {
                Cudd_RecursiveDeref(dd,glocal);
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
            (*conjuncts)[0] = glocal;
            return(1);
        }
    } else {
        Cudd_RecursiveDeref(dd,glocal);
        *conjuncts = ABC_ALLOC(DdNode *,1);
        if *conjuncts == nil {
            Cudd_RecursiveDeref(dd,hlocal);
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
        (*conjuncts)[0] = hlocal;
        return(1);
    }

} /* end of Cudd_bddApproxConjDecomp */


/**Function********************************************************************

  Synopsis    [Performs two-way disjunctive decomposition of a BDD.]

  Description [Performs two-way disjunctive decomposition of a BDD.
  Returns the number of disjuncts produced, that is, 2 if successful;
  1 if no meaningful decomposition was found; 0 otherwise. The
  disjuncts produced by this procedure tend to be imbalanced.]

  SideEffects [The two disjuncts are returned in an array as side effects.
  The array is allocated by this function. It is the caller's responsibility
  to free it. On successful completion, the disjuncts are already
  referenced. If the function returns 0, the array for the disjuncts is
  not allocated. If the function returns 1, the only factor equals the
  function to be decomposed.]

  SeeAlso     [Cudd_bddApproxConjDecomp Cudd_bddIterDisjDecomp
  Cudd_bddGenDisjDecomp Cudd_bddVarDisjDecomp]

******************************************************************************/
int
Cudd_bddApproxDisjDecomp(
  DdManager * dd /* manager */,
  DdNode * f /* function to be decomposed */,
  DdNode *** disjuncts /* address of the array of the disjuncts */)
{
    var result, i int32;

    // inlining of macro Cudd_Not
    result = Cudd_bddApproxConjDecomp(dd,((DdNode *)((ptrint)(f) ^ 01)),disjuncts);
    for i = 0; i < result; i++ {
        // inlining of macro Cudd_Not
        (*disjuncts)[i] = ((DdNode *)((ptrint)((*disjuncts)[i]) ^ 01));
    }
    return(result);

} /* end of Cudd_bddApproxDisjDecomp */


/**Function********************************************************************

  Synopsis    [Performs two-way conjunctive decomposition of a BDD.]

  Description [Performs two-way conjunctive decomposition of a
  BDD. This procedure owes its name to the iterated use of
  supersetting to obtain a factor of the given function. Returns the
  number of conjuncts produced, that is, 2 if successful; 1 if no
  meaningful decomposition was found; 0 otherwise. The conjuncts
  produced by this procedure tend to be imbalanced.]

  SideEffects [The factors are returned in an array as side effects.
  The array is allocated by this function. It is the caller's responsibility
  to free it. On successful completion, the conjuncts are already
  referenced. If the function returns 0, the array for the conjuncts is
  not allocated. If the function returns 1, the only factor equals the
  function to be decomposed.]

  SeeAlso     [Cudd_bddIterDisjDecomp Cudd_bddApproxConjDecomp
  Cudd_bddGenConjDecomp Cudd_bddVarConjDecomp Cudd_RemapOverApprox
  Cudd_bddSqueeze Cudd_bddLICompaction]

******************************************************************************/
int
Cudd_bddIterConjDecomp(
  DdManager * dd /* manager */,
  DdNode * f /* function to be decomposed */,
  DdNode *** conjuncts /* address of the array of conjuncts */)
{
    var superset1 *DdNode;
    var superset2 *DdNode;
    var old [2]*DdNode;
    var res [2]*DdNode;
    var sizeOld, sizeNew int32;
    var nvars int32  = Cudd_SupportSize(dd,f);

    // inlining of macro DD_ONE
    old[0] = ((dd).one);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(old[0]).ref);
    old[1] = f;
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(old[1]).ref);
    sizeOld = Cudd_SharingSize(old,2);

    for {
        /* Find a tentative first factor by overapproximation and
        ** minimization. */
        superset1 = Cudd_RemapOverApprox(dd,old[1],nvars,0,1.0);
        if superset1 == nil {
            Cudd_RecursiveDeref(dd,old[0]);
            Cudd_RecursiveDeref(dd,old[1]);
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(superset1).ref);
        superset2 = Cudd_bddSqueeze(dd,old[1],superset1);
        if superset2 == nil {
            Cudd_RecursiveDeref(dd,old[0]);
            Cudd_RecursiveDeref(dd,old[1]);
            Cudd_RecursiveDeref(dd,superset1);
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(superset2).ref);
        Cudd_RecursiveDeref(dd,superset1);
        res[0] = Cudd_bddAnd(dd,old[0],superset2);
        if res[0] == nil {
            Cudd_RecursiveDeref(dd,superset2);
            Cudd_RecursiveDeref(dd,old[0]);
            Cudd_RecursiveDeref(dd,old[1]);
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res[0]).ref);
        Cudd_RecursiveDeref(dd,superset2);
        if res[0] == old[0] {
            Cudd_RecursiveDeref(dd,res[0]);
            break;      /* avoid infinite loop */
        }

        /* Compute the second factor by minimization. */
        res[1] = Cudd_bddLICompaction(dd,old[1],res[0]);
        if res[1] == nil {
            Cudd_RecursiveDeref(dd,old[0]);
            Cudd_RecursiveDeref(dd,old[1]);
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res[1]).ref);

        sizeNew = Cudd_SharingSize(res,2);
        if sizeNew <= sizeOld {
            Cudd_RecursiveDeref(dd,old[0]);
            old[0] = res[0];
            Cudd_RecursiveDeref(dd,old[1]);
            old[1] = res[1];
            sizeOld = sizeNew;
        } else {
            Cudd_RecursiveDeref(dd,res[0]);
            Cudd_RecursiveDeref(dd,res[1]);
            break;
        }

    
if !(true) { break }
} 

    /* Refine the first factor by minimization. If h turns out to
    ** be f, this step guarantees that g will be 1. */
    superset1 = Cudd_bddLICompaction(dd,old[0],old[1]);
    if superset1 == nil {
        Cudd_RecursiveDeref(dd,old[0]);
        Cudd_RecursiveDeref(dd,old[1]);
        return(0);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(superset1).ref);
    Cudd_RecursiveDeref(dd,old[0]);
    old[0] = superset1;

    // inlining of macro DD_ONE
    if old[0] != ((dd).one) {
        // inlining of macro DD_ONE
        if old[1] != ((dd).one) {
            *conjuncts = ABC_ALLOC(DdNode *,2);
            if *conjuncts == nil {
                Cudd_RecursiveDeref(dd,old[0]);
                Cudd_RecursiveDeref(dd,old[1]);
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
            (*conjuncts)[0] = old[0];
            (*conjuncts)[1] = old[1];
            return(2);
        } else {
            Cudd_RecursiveDeref(dd,old[1]);
            *conjuncts = ABC_ALLOC(DdNode *,1);
            if *conjuncts == nil {
                Cudd_RecursiveDeref(dd,old[0]);
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
            (*conjuncts)[0] = old[0];
            return(1);
        }
    } else {
        Cudd_RecursiveDeref(dd,old[0]);
        *conjuncts = ABC_ALLOC(DdNode *,1);
        if *conjuncts == nil {
            Cudd_RecursiveDeref(dd,old[1]);
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
        (*conjuncts)[0] = old[1];
        return(1);
    }

} /* end of Cudd_bddIterConjDecomp */


/**Function********************************************************************

  Synopsis    [Performs two-way disjunctive decomposition of a BDD.]

  Description [Performs two-way disjunctive decomposition of a BDD.
  Returns the number of disjuncts produced, that is, 2 if successful;
  1 if no meaningful decomposition was found; 0 otherwise. The
  disjuncts produced by this procedure tend to be imbalanced.]

  SideEffects [The two disjuncts are returned in an array as side effects.
  The array is allocated by this function. It is the caller's responsibility
  to free it. On successful completion, the disjuncts are already
  referenced. If the function returns 0, the array for the disjuncts is
  not allocated. If the function returns 1, the only factor equals the
  function to be decomposed.]

  SeeAlso     [Cudd_bddIterConjDecomp Cudd_bddApproxDisjDecomp
  Cudd_bddGenDisjDecomp Cudd_bddVarDisjDecomp]

******************************************************************************/
int
Cudd_bddIterDisjDecomp(
  DdManager * dd /* manager */,
  DdNode * f /* function to be decomposed */,
  DdNode *** disjuncts /* address of the array of the disjuncts */)
{
    var result, i int32;

    // inlining of macro Cudd_Not
    result = Cudd_bddIterConjDecomp(dd,((DdNode *)((ptrint)(f) ^ 01)),disjuncts);
    for i = 0; i < result; i++ {
        // inlining of macro Cudd_Not
        (*disjuncts)[i] = ((DdNode *)((ptrint)((*disjuncts)[i]) ^ 01));
    }
    return(result);

} /* end of Cudd_bddIterDisjDecomp */


/**Function********************************************************************

  Synopsis    [Performs two-way conjunctive decomposition of a BDD.]

  Description [Performs two-way conjunctive decomposition of a
  BDD. This procedure owes its name to the fact tht it generalizes the
  decomposition based on the cofactors with respect to one
  variable. Returns the number of conjuncts produced, that is, 2 if
  successful; 1 if no meaningful decomposition was found; 0
  otherwise. The conjuncts produced by this procedure tend to be
  balanced.]

  SideEffects [The two factors are returned in an array as side effects.
  The array is allocated by this function. It is the caller's responsibility
  to free it. On successful completion, the conjuncts are already
  referenced. If the function returns 0, the array for the conjuncts is
  not allocated. If the function returns 1, the only factor equals the
  function to be decomposed.]

  SeeAlso     [Cudd_bddGenDisjDecomp Cudd_bddApproxConjDecomp
  Cudd_bddIterConjDecomp Cudd_bddVarConjDecomp]

******************************************************************************/
int
Cudd_bddGenConjDecomp(
  DdManager * dd /* manager */,
  DdNode * f /* function to be decomposed */,
  DdNode *** conjuncts /* address of the array of conjuncts */)
{
    var result int32;
    var glocal, hlocal *DdNode;

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));
    
    for {
        dd.reordered = 0;
        result = cuddConjunctsAux(dd, f, &glocal, &hlocal);
    
if !(dd.reordered == 1) { break }
} 

    if result == 0 {
        return(0);
    }

    if glocal != one {
        if hlocal != one {
            *conjuncts = ABC_ALLOC(DdNode *,2);
            if *conjuncts == nil {
                Cudd_RecursiveDeref(dd,glocal);
                Cudd_RecursiveDeref(dd,hlocal);
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
            (*conjuncts)[0] = glocal;
            (*conjuncts)[1] = hlocal;
            return(2);
        } else {
            Cudd_RecursiveDeref(dd,hlocal);
            *conjuncts = ABC_ALLOC(DdNode *,1);
            if *conjuncts == nil {
                Cudd_RecursiveDeref(dd,glocal);
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
            (*conjuncts)[0] = glocal;
            return(1);
        }
    } else {
        Cudd_RecursiveDeref(dd,glocal);
        *conjuncts = ABC_ALLOC(DdNode *,1);
        if *conjuncts == nil {
            Cudd_RecursiveDeref(dd,hlocal);
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
        (*conjuncts)[0] = hlocal;
        return(1);
    }

} /* end of Cudd_bddGenConjDecomp */


/**Function********************************************************************

  Synopsis    [Performs two-way disjunctive decomposition of a BDD.]

  Description [Performs two-way disjunctive decomposition of a BDD.
  Returns the number of disjuncts produced, that is, 2 if successful;
  1 if no meaningful decomposition was found; 0 otherwise. The
  disjuncts produced by this procedure tend to be balanced.]

  SideEffects [The two disjuncts are returned in an array as side effects.
  The array is allocated by this function. It is the caller's responsibility
  to free it. On successful completion, the disjuncts are already
  referenced. If the function returns 0, the array for the disjuncts is
  not allocated. If the function returns 1, the only factor equals the
  function to be decomposed.]

  SeeAlso     [Cudd_bddGenConjDecomp Cudd_bddApproxDisjDecomp
  Cudd_bddIterDisjDecomp Cudd_bddVarDisjDecomp]

******************************************************************************/
int
Cudd_bddGenDisjDecomp(
  DdManager * dd /* manager */,
  DdNode * f /* function to be decomposed */,
  DdNode *** disjuncts /* address of the array of the disjuncts */)
{
    var result, i int32;

    // inlining of macro Cudd_Not
    result = Cudd_bddGenConjDecomp(dd,((DdNode *)((ptrint)(f) ^ 01)),disjuncts);
    for i = 0; i < result; i++ {
        // inlining of macro Cudd_Not
        (*disjuncts)[i] = ((DdNode *)((ptrint)((*disjuncts)[i]) ^ 01));
    }
    return(result);

} /* end of Cudd_bddGenDisjDecomp */


/**Function********************************************************************

  Synopsis    [Performs two-way conjunctive decomposition of a BDD.]

  Description [Conjunctively decomposes one BDD according to a
  variable.  If <code>f</code> is the function of the BDD and
  <code>x</code> is the variable, the decomposition is
  <code>(f+x)(f+x')</code>.  The variable is chosen so as to balance
  the sizes of the two conjuncts and to keep them small.  Returns the
  number of conjuncts produced, that is, 2 if successful; 1 if no
  meaningful decomposition was found; 0 otherwise.]

  SideEffects [The two factors are returned in an array as side effects.
  The array is allocated by this function. It is the caller's responsibility
  to free it. On successful completion, the conjuncts are already
  referenced. If the function returns 0, the array for the conjuncts is
  not allocated. If the function returns 1, the only factor equals the
  function to be decomposed.]

  SeeAlso     [Cudd_bddVarDisjDecomp Cudd_bddGenConjDecomp
  Cudd_bddApproxConjDecomp Cudd_bddIterConjDecomp]

*****************************************************************************/
int
Cudd_bddVarConjDecomp(
  DdManager * dd /* manager */,
  DdNode * f /* function to be decomposed */,
  DdNode *** conjuncts /* address of the array of conjuncts */)
{
    var best int32;
    var min int32;
    var support, scan, var, glocal, hlocal *DdNode;

    /* Find best cofactoring variable. */
    support = Cudd_Support(dd,f);
    if support == nil { return(0); }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(support) & ~01))).index == CUDD_CONST_INDEX {
        *conjuncts = ABC_ALLOC(DdNode *,1);
        if *conjuncts == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
        (*conjuncts)[0] = f;
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular((*conjuncts)[0]).ref);
        return(1);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(support).ref);
    min = 1000000000;
    best = -1;
    scan = support;
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    for !((((DdNode *)((ptruint)(scan) & ~01))).index == CUDD_CONST_INDEX) {
        var i int32  = scan.index;
        var est1 int32  = Cudd_EstimateCofactor(dd,f,i,1);
        var est0 int32  = Cudd_EstimateCofactor(dd,f,i,0);
        /* Minimize the size of the larger of the two cofactors. */
        var est int32;
        if est1 > est0 {
            est = est1;
        } else {
            est = est0;
        }
        if est < min {
            min = est;
            best = i;
        }
        // inlining of macro cuddT
        scan = ((scan).type.kids.T);
    }
#ifdef DD_DEBUG
    assert(best >= 0 && best < dd->size);
#endif
    Cudd_RecursiveDeref(dd,support);

    var = Cudd_bddIthVar(dd,best);
    glocal = Cudd_bddOr(dd,f,var);
    if glocal == nil {
        return(0);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(glocal).ref);
    // inlining of macro Cudd_Not
    hlocal = Cudd_bddOr(dd,f,((DdNode *)((ptrint)(var) ^ 01)));
    if hlocal == nil {
        Cudd_RecursiveDeref(dd,glocal);
        return(0);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(hlocal).ref);

    // inlining of macro DD_ONE
    if glocal != ((dd).one) {
        // inlining of macro DD_ONE
        if hlocal != ((dd).one) {
            *conjuncts = ABC_ALLOC(DdNode *,2);
            if *conjuncts == nil {
                Cudd_RecursiveDeref(dd,glocal);
                Cudd_RecursiveDeref(dd,hlocal);
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
            (*conjuncts)[0] = glocal;
            (*conjuncts)[1] = hlocal;
            return(2);
        } else {
            Cudd_RecursiveDeref(dd,hlocal);
            *conjuncts = ABC_ALLOC(DdNode *,1);
            if *conjuncts == nil {
                Cudd_RecursiveDeref(dd,glocal);
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
            (*conjuncts)[0] = glocal;
            return(1);
        }
    } else {
        Cudd_RecursiveDeref(dd,glocal);
        *conjuncts = ABC_ALLOC(DdNode *,1);
        if *conjuncts == nil {
            Cudd_RecursiveDeref(dd,hlocal);
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
        (*conjuncts)[0] = hlocal;
        return(1);
    }

} /* end of Cudd_bddVarConjDecomp */


/**Function********************************************************************

  Synopsis    [Performs two-way disjunctive decomposition of a BDD.]

  Description [Performs two-way disjunctive decomposition of a BDD
  according to a variable. If <code>f</code> is the function of the
  BDD and <code>x</code> is the variable, the decomposition is
  <code>f*x + f*x'</code>.  The variable is chosen so as to balance
  the sizes of the two disjuncts and to keep them small.  Returns the
  number of disjuncts produced, that is, 2 if successful; 1 if no
  meaningful decomposition was found; 0 otherwise.]

  SideEffects [The two disjuncts are returned in an array as side effects.
  The array is allocated by this function. It is the caller's responsibility
  to free it. On successful completion, the disjuncts are already
  referenced. If the function returns 0, the array for the disjuncts is
  not allocated. If the function returns 1, the only factor equals the
  function to be decomposed.]

  SeeAlso     [Cudd_bddVarConjDecomp Cudd_bddApproxDisjDecomp
  Cudd_bddIterDisjDecomp Cudd_bddGenDisjDecomp]

******************************************************************************/
int
Cudd_bddVarDisjDecomp(
  DdManager * dd /* manager */,
  DdNode * f /* function to be decomposed */,
  DdNode *** disjuncts /* address of the array of the disjuncts */)
{
    var result, i int32;

    // inlining of macro Cudd_Not
    result = Cudd_bddVarConjDecomp(dd,((DdNode *)((ptrint)(f) ^ 01)),disjuncts);
    for i = 0; i < result; i++ {
        // inlining of macro Cudd_Not
        (*disjuncts)[i] = ((DdNode *)((ptrint)((*disjuncts)[i]) ^ 01));
    }
    return(result);

} /* end of Cudd_bddVarDisjDecomp */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Get longest distance of node from constant.]

  Description [Get longest distance of node from constant. Returns the
  distance of the root from the constant if successful; CUDD_OUT_OF_MEM
  otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func CreateBotDist(node *DdNode, distanceTable *st__table) *NodeStat {
    var N, Nv, Nnv *DdNode;
    var distance, distanceNv, distanceNnv int32;
    var nodeStat, nodeStatNv, nodeStatNnv *NodeStat;

     
                                
                  
     
      
    
    /* Return the entry in the table if found. */
    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));
    if st__lookup(distanceTable, (char *)N, (char **)&nodeStat) {
        nodeStat.localRef++;
        return(nodeStat);
    }

    // inlining of macro cuddT
    Nv = ((N).type.kids.T);
    // inlining of macro cuddE
    Nnv = ((N).type.kids.E);
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Nv = ((DdNode *)((ptrint)(Nv) ^ ((int32( ((ptrint) (node) & 01))))));
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Nnv = ((DdNode *)((ptrint)(Nnv) ^ ((int32( ((ptrint) (node) & 01))))));

    /* Recur on the children. */
    nodeStatNv = CreateBotDist(Nv, distanceTable);
    if nodeStatNv == nil { return(nil); }
    distanceNv = nodeStatNv.distance;

    nodeStatNnv = CreateBotDist(Nnv, distanceTable);
    if nodeStatNnv == nil { return(nil); }
    distanceNnv = nodeStatNnv.distance;
    /* Store max distance from constant; note sometimes this distance
    ** may be to 0.
    */
    if distanceNv > distanceNnv {
        distance = (distanceNv+1);
    } else {
        distance = (distanceNnv + 1);
    }

    nodeStat = new(NodeStat);
    if nodeStat == nil {
        return(0);
    }
    nodeStat.distance = distance;
    nodeStat.localRef = 1;
    
    if st__insert(distanceTable, (char *)N, (char *)nodeStat) ==
        st__OUT_OF_MEM {
        return(0);

    }
    return(nodeStat);

} /* end of CreateBotDist */


/**Function********************************************************************

  Synopsis    [Count the number of minterms of each node ina a BDD and
  store it in a hash table.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func CountMinterms(node *DdNode, max float64, mintermTable *st__table, fp *os.File) float64 {
    var N, Nv, Nnv *DdNode;
    var min, minNv, minNnv float64;
    var dummy *float64;

    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));

    // inlining of macro cuddIsConstant
    if (N).index == CUDD_CONST_INDEX {
        if node == zero {
            return(0);
        } else {
            return(max);
        }
    }

    /* Return the entry in the table if found. */
    if st__lookup(mintermTable, (char *)node, (char **)&dummy) {
        min = *dummy;
        return(min);
    }

    // inlining of macro cuddT
    Nv = ((N).type.kids.T);
    // inlining of macro cuddE
    Nnv = ((N).type.kids.E);
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Nv = ((DdNode *)((ptrint)(Nv) ^ ((int32( ((ptrint) (node) & 01))))));
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Nnv = ((DdNode *)((ptrint)(Nnv) ^ ((int32( ((ptrint) (node) & 01))))));

    /* Recur on the children. */
    minNv = CountMinterms(Nv, max, mintermTable, fp);
    if minNv == -1.0 { return(-1.0); }
    minNnv = CountMinterms(Nnv, max, mintermTable, fp);
    if minNnv == -1.0 { return(-1.0); }
    min = minNv / 2.0 + minNnv / 2.0;
    /* store 
     */

    dummy = new(float64);
    if dummy == nil { return(-1.0); }
    *dummy = min;
    if st__insert(mintermTable, (char *)node, (char *)dummy) == st__OUT_OF_MEM {
        (void) fmt.Fprintf(fp, "st table insert failed\n");
    }
    return(min);

} /* end of CountMinterms */


/**Function********************************************************************

  Synopsis    [Free factors structure]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func ConjunctsFree(dd *DdManager, factors *Conjuncts) {
    Cudd_RecursiveDeref(dd, factors.g);
    Cudd_RecursiveDeref(dd, factors.h);
    ABC_FREE(factors);
    return;

} /* end of ConjunctsFree */


/**Function********************************************************************

  Synopsis    [Check whether the given pair is in the tables.]

  Description [.Check whether the given pair is in the tables.  gTable
  and hTable are combined.
  absence in both is indicated by 0,
  presence in gTable is indicated by 1,
  presence in hTable by 2 and
  presence in both by 3.
  The values returned by this function are PAIR_ST,
  PAIR_CR, G_ST, G_CR, H_ST, H_CR, BOTH_G, BOTH_H, NONE.
  PAIR_ST implies g in gTable and h in hTable
  PAIR_CR implies g in hTable and h in gTable
  G_ST implies g in gTable and h not in any table
  G_CR implies g in hTable and h not in any table
  H_ST implies h in hTable and g not in any table
  H_CR implies h in gTable and g not in any table
  BOTH_G implies both in gTable
  BOTH_H implies both in hTable
  NONE implies none in table; ]

  SideEffects []

  SeeAlso     [CheckTablesCacheAndReturn CheckInTables]

******************************************************************************/
func PairInTables(g *DdNode, h *DdNode, ghTable *st__table) int32 {
    var valueG, valueH, gPresent, hPresent int32;

    valueG = valueH = gPresent = hPresent = 0;
    
    // inlining of macro Cudd_Regular
    gPresent = st__lookup_int(ghTable, (char *)((DdNode *)((ptruint)(g) & ~01)), &valueG);
    // inlining of macro Cudd_Regular
    hPresent = st__lookup_int(ghTable, (char *)((DdNode *)((ptruint)(h) & ~01)), &valueH);

    if !gPresent && !(hPresent != 0) { return(NONE); }

    if !(hPresent != 0) {
        if valueG & 1 { return(G_ST); }
        if valueG & 2 { return(G_CR); }
    }
    if !gPresent {
        if valueH & 1 { return(H_CR); }
        if valueH & 2 { return(H_ST); }
    }
    /* both in tables */
    if (valueG & 1) && (valueH & 2) { return(PAIR_ST); }
    if (valueG & 2) && (valueH & 1) { return(PAIR_CR); }
    
    if valueG & 1 {
        return(BOTH_G);
    } else {
        return(BOTH_H);
    }
    
} /* end of PairInTables */


/**Function********************************************************************

  Synopsis    [Check the tables for the existence of pair and return one
  combination, cache the result.]

  Description [Check the tables for the existence of pair and return
  one combination, cache the result. The assumption is that one of the
  conjuncts is already in the tables.]

  SideEffects [g and h referenced for the cache]

  SeeAlso     [ZeroCase]

******************************************************************************/
func CheckTablesCacheAndReturn(node *DdNode, g *DdNode, h *DdNode, ghTable *st__table, cacheTable *st__table) *Conjuncts {
    var pairValue int32;
    var value int32;
    var factors *Conjuncts;
    
    value = 0;
    /* check tables */
    pairValue = PairInTables(g, h, ghTable);
    assert(pairValue != NONE, "pairValue != NONE");
    /* if both dont exist in table, we know one exists(either g or h).
     * Therefore store the other and proceed
     */
    factors = new(Conjuncts);
    if factors == nil { return(nil); }
    if (pairValue == BOTH_H) || (pairValue == H_ST) {
        if g != one {
            value = 0;
            // inlining of macro Cudd_Regular
            if st__lookup_int(ghTable, (char *)((DdNode *)((ptruint)(g) & ~01)), &value) {
                value |= 1;
            } else {
                value = 1;
            }
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(g) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                return(nil);
            }
        }
        factors.g = g;
        factors.h = h;
    } else  if (pairValue == BOTH_G) || (pairValue == G_ST) {
        if h != one {
            value = 0;
            // inlining of macro Cudd_Regular
            if st__lookup_int(ghTable, (char *)((DdNode *)((ptruint)(h) & ~01)), &value) {
                value |= 2;
            } else {
                value = 2;
            }
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(h) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                return(nil);
            }
        }
        factors.g = g;
        factors.h = h;
    } else if pairValue == H_CR {
        if g != one {
            value = 2;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(g) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                return(nil);
            }
        }
        factors.g = h;
        factors.h = g;
    } else if pairValue == G_CR {
        if h != one {
            value = 1;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(h) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                return(nil);
            }
        }
        factors.g = h;
        factors.h = g;
    } else if pairValue == PAIR_CR {
    /* pair exists in table */
        factors.g = h;
        factors.h = g;
    } else if pairValue == PAIR_ST {
        factors.g = g;
        factors.h = h;
    }
            
    /* cache the result for this node */
    if st__insert(cacheTable, (char *)node, (char *)factors) == st__OUT_OF_MEM {
        ABC_FREE(factors);
        return(nil);
    }

    return(factors);

} /* end of CheckTablesCacheAndReturn */
        
/**Function********************************************************************

  Synopsis    [Check the tables for the existence of pair and return one
  combination, store in cache.]

  Description [Check the tables for the existence of pair and return
  one combination, store in cache. The pair that has more pointers to
  it is picked. An approximation of the number of local pointers is
  made by taking the reference count of the pairs sent. ]

  SideEffects []

  SeeAlso     [ZeroCase BuildConjuncts]

******************************************************************************/
func PickOnePair(node *DdNode, g1 *DdNode, h1 *DdNode, g2 *DdNode, h2 *DdNode, ghTable *st__table, cacheTable *st__table) *Conjuncts {
    var value int32;
    var factors *Conjuncts;
    var oneRef, twoRef int32;
    
    factors = new(Conjuncts);
    if factors == nil { return(nil); }

    /* count the number of pointers to pair 2 */
    if h2 == one {
        // inlining of macro Cudd_Regular
        twoRef = (((DdNode *)((ptruint)(g2) & ~01))).ref;
    } else if g2 == one {
        // inlining of macro Cudd_Regular
        twoRef = (((DdNode *)((ptruint)(h2) & ~01))).ref;
    } else {
        // inlining of macro Cudd_Regular
        twoRef = ((((DdNode *)((ptruint)(g2) & ~01))).ref + (((DdNode *)((ptruint)(h2) & ~01))).ref)/2;
    }

    /* count the number of pointers to pair 1 */
    if h1 == one {
        // inlining of macro Cudd_Regular
        oneRef  = (((DdNode *)((ptruint)(g1) & ~01))).ref;
    } else if g1 == one {
        // inlining of macro Cudd_Regular
        oneRef  = (((DdNode *)((ptruint)(h1) & ~01))).ref;
    } else {
        // inlining of macro Cudd_Regular
        oneRef = ((((DdNode *)((ptruint)(g1) & ~01))).ref + (((DdNode *)((ptruint)(h1) & ~01))).ref)/2;
    }

    /* pick the pair with higher reference count */
    if oneRef >= twoRef {
        factors.g = g1;
        factors.h = h1;
    } else {
        factors.g = g2;
        factors.h = h2;
    }
    
    /*
     * Store computed factors in respective tables to encourage
     * recombination.
     */
    if factors.g != one {
        /* insert g in htable */
        value = 0;
        // inlining of macro Cudd_Regular
        if st__lookup_int(ghTable, (char *)((DdNode *)((ptruint)(factors.g) & ~01)), &value) {
            if value == 2 {
                value |= 1;
                // inlining of macro Cudd_Regular
                if st__insert(ghTable, (char *)((DdNode *)((ptruint)(factors.g) & ~01)),
                              (char *)int64(value)) == st__OUT_OF_MEM {
                    ABC_FREE(factors);
                    return(nil);
                }
            }
        } else {
            value = 1;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(factors.g) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                ABC_FREE(factors);
                return(nil);
            }
        }
    }

    if factors.h != one {
        /* insert h in htable */
        value = 0;
        // inlining of macro Cudd_Regular
        if st__lookup_int(ghTable, (char *)((DdNode *)((ptruint)(factors.h) & ~01)), &value) {
            if value == 1 {
                value |= 2;
                // inlining of macro Cudd_Regular
                if st__insert(ghTable, (char *)((DdNode *)((ptruint)(factors.h) & ~01)),
                              (char *)int64(value)) == st__OUT_OF_MEM {
                    ABC_FREE(factors);
                    return(nil);
                }
            }       
        } else {
            value = 2;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(factors.h) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                ABC_FREE(factors);
                return(nil);
            }
        }
    }
    
    /* Store factors in cache table for later use. */
    if st__insert(cacheTable, (char *)node, (char *)factors) ==
            st__OUT_OF_MEM {
        ABC_FREE(factors);
        return(nil);
    }

    return(factors);

} /* end of PickOnePair */


/**Function********************************************************************

  Synopsis [Check if the two pairs exist in the table, If any of the
  conjuncts do exist, store in the cache and return the corresponding pair.]

  Description [Check if the two pairs exist in the table. If any of
  the conjuncts do exist, store in the cache and return the
  corresponding pair.]

  SideEffects []

  SeeAlso     [ZeroCase BuildConjuncts]

******************************************************************************/
func CheckInTables(node *DdNode, g1 *DdNode, h1 *DdNode, g2 *DdNode, h2 *DdNode, ghTable *st__table, cacheTable *st__table, outOfMem *int32) *Conjuncts {
    var pairValue1, pairValue2 int32;
    var factors *Conjuncts;
    var value int32;
    
    *outOfMem = 0;

    /* check existence of pair in table */
    pairValue1 = PairInTables(g1, h1, ghTable);
    pairValue2 = PairInTables(g2, h2, ghTable);

    /* if none of the 4 exist in the gh tables, return NULL */
    if (pairValue1 == NONE) && (pairValue2 == NONE) {
        return nil;
    }
    
    factors = new(Conjuncts);
    if factors == nil {
        *outOfMem = 1;
        return nil;
    }

    /* pairs that already exist in the table get preference. */
    if pairValue1 == PAIR_ST {
        factors.g = g1;
        factors.h = h1;
    } else if pairValue2 == PAIR_ST {
        factors.g = g2;
        factors.h = h2;
    } else if pairValue1 == PAIR_CR {
        factors.g = h1;
        factors.h = g1;
    } else if pairValue2 == PAIR_CR {
        factors.g = h2;
        factors.h = g2;
    } else if pairValue1 == G_ST {
        /* g exists in the table, h is not found in either table */
        factors.g = g1;
        factors.h = h1;
        if h1 != one {
            value = 2;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(h1) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    } else if pairValue1 == BOTH_G {
        /* g and h are  found in the g table */
        factors.g = g1;
        factors.h = h1;
        if h1 != one {
            value = 3;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(h1) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    } else if pairValue1 == H_ST {
        /* h exists in the table, g is not found in either table */
        factors.g = g1;
        factors.h = h1;
        if g1 != one {
            value = 1;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(g1) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    } else if pairValue1 == BOTH_H {
        /* g and h are  found in the h table */
        factors.g = g1;
        factors.h = h1;
        if g1 != one {
            value = 3;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(g1) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    } else if pairValue2 == G_ST {
        /* g exists in the table, h is not found in either table */
        factors.g = g2;
        factors.h = h2;
        if h2 != one {
            value = 2;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(h2) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    } else if  pairValue2 == BOTH_G {
        /* g and h are  found in the g table */
        factors.g = g2;
        factors.h = h2;
        if h2 != one {
            value = 3;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(h2) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    } else if pairValue2 == H_ST { 
        /* h exists in the table, g is not found in either table */
        factors.g = g2;
        factors.h = h2;
        if g2 != one {
            value = 1;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(g2) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    } else if pairValue2 == BOTH_H {
        /* g and h are  found in the h table */
        factors.g = g2;
        factors.h = h2;
        if g2 != one {
            value = 3;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(g2) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    } else if pairValue1 == G_CR {
        /* g found in h table and h in none */
        factors.g = h1;
        factors.h = g1;
        if h1 != one {
            value = 1;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(h1) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    } else if pairValue1 == H_CR {
        /* h found in g table and g in none */
        factors.g = h1;
        factors.h = g1;
        if g1 != one {
            value = 2;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(g1) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    } else if pairValue2 == G_CR {
        /* g found in h table and h in none */
        factors.g = h2;
        factors.h = g2;
        if h2 != one {
            value = 1;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(h2) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    } else if pairValue2 == H_CR {
        /* h found in g table and g in none */
        factors.g = h2;
        factors.h = g2;
        if g2 != one {
            value = 2;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(g2) & ~01)),
                          (char *)int64(value)) == st__OUT_OF_MEM {
                *outOfMem = 1;
                ABC_FREE(factors);
                return(nil);
            }
        }
    }
    
    /* Store factors in cache table for later use. */
    if st__insert(cacheTable, (char *)node, (char *)factors) ==
            st__OUT_OF_MEM {
        *outOfMem = 1;
        ABC_FREE(factors);
        return(nil);
    }
    return factors;
} /* end of CheckInTables */



/**Function********************************************************************

  Synopsis    [If one child is zero, do explicitly what Restrict does or better]

  Description [If one child is zero, do explicitly what Restrict does or better.
  First separate a variable and its child in the base case. In case of a cube
  times a function, separate the cube and function. As a last resort, look in
  tables.]

  SideEffects [Frees the BDDs in factorsNv. factorsNv itself is not freed
  because it is freed above.]

  SeeAlso     [BuildConjuncts]

******************************************************************************/
func ZeroCase(dd *DdManager, node *DdNode, factorsNv *Conjuncts, ghTable *st__table, cacheTable *st__table, switched int32) *Conjuncts {
    var topid int32;
    var g, h, g1, g2, h1, h2, x, N, G, H, Gv, Gnv *DdNode;
    var Hv, Hnv *DdNode;
    var value int32;
    var outOfMem int32;
    var factors *Conjuncts;
    
    /* get var at this node */
    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));
    topid = N.index;
    x = dd.vars[topid];
    if switched != 0 {
        // inlining of macro Cudd_Not
        x = ((DdNode *)((ptrint)(x) ^ 01));
    } else {
        x = x;
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(x).ref);

    /* Seprate variable and child */
    if factorsNv.g == one {
        Cudd_RecursiveDeref(dd, factorsNv.g);
        factors = new(Conjuncts);
        if factors == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            Cudd_RecursiveDeref(dd, factorsNv.h);
            Cudd_RecursiveDeref(dd, x);
            return(nil);
        }
        factors.g = x;
        factors.h = factorsNv.h;
        /* cache the result*/
        if st__insert(cacheTable, (char *)node, (char *)factors) == st__OUT_OF_MEM {
            dd.errorCode = CUDD_MEMORY_OUT;
            Cudd_RecursiveDeref(dd, factorsNv.h); 
            Cudd_RecursiveDeref(dd, x);
            ABC_FREE(factors);
            return nil;
        }
        
        /* store  x in g table, the other node is already in the table */
        // inlining of macro Cudd_Regular
        if st__lookup_int(ghTable, (char *)((DdNode *)((ptruint)(x) & ~01)), &value) {
            value |= 1;
        } else {
            value = 1;
        }
        // inlining of macro Cudd_Regular
        if st__insert(ghTable, (char *)((DdNode *)((ptruint)(x) & ~01)), (char *)int64(value)) == st__OUT_OF_MEM {
            dd.errorCode = CUDD_MEMORY_OUT;
            return nil;
        }
        return(factors);
    }
    
    /* Seprate variable and child */
    if factorsNv.h == one {
        Cudd_RecursiveDeref(dd, factorsNv.h);
        factors = new(Conjuncts);
        if factors == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            Cudd_RecursiveDeref(dd, factorsNv.g);
            Cudd_RecursiveDeref(dd, x);
            return(nil);
        }
        factors.g = factorsNv.g;
        factors.h = x;
        /* cache the result. */
        if st__insert(cacheTable, (char *)node, (char *)factors) == st__OUT_OF_MEM {
            dd.errorCode = CUDD_MEMORY_OUT;
            Cudd_RecursiveDeref(dd, factorsNv.g);
            Cudd_RecursiveDeref(dd, x);
            ABC_FREE(factors);
            return(nil);
        }
        /* store x in h table,  the other node is already in the table */
        // inlining of macro Cudd_Regular
        if st__lookup_int(ghTable, (char *)((DdNode *)((ptruint)(x) & ~01)), &value) {
            value |= 2;
        } else {
            value = 2;
        }
        // inlining of macro Cudd_Regular
        if st__insert(ghTable, (char *)((DdNode *)((ptruint)(x) & ~01)), (char *)int64(value)) == st__OUT_OF_MEM {
            dd.errorCode = CUDD_MEMORY_OUT;
            return nil;
        }
        return(factors);
    }

    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(factorsNv.g) & ~01));
    // inlining of macro cuddT
    Gv = ((G).type.kids.T);
    // inlining of macro cuddE
    Gnv = ((G).type.kids.E);
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Gv = ((DdNode *)((ptrint)(Gv) ^ ((int32( ((ptrint) (node) & 01))))));
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Gnv = ((DdNode *)((ptrint)(Gnv) ^ ((int32( ((ptrint) (node) & 01))))));
    /* if the child below is a variable */
    if (Gv == zero) || (Gnv == zero) {
        h = factorsNv.h;
        g = cuddBddAndRecur(dd, x, factorsNv.g);
        // inlining of macro cuddRef
        if g != nil {  cuddSatInc(Cudd_Regular(g).ref); }
        Cudd_RecursiveDeref(dd, factorsNv.g); 
        Cudd_RecursiveDeref(dd, x);
        if g == nil {
            Cudd_RecursiveDeref(dd, factorsNv.h); 
            return nil;
        }
        /* CheckTablesCacheAndReturn responsible for allocating
         * factors structure., g,h referenced for cache store  the
         */
        factors = CheckTablesCacheAndReturn(node,
                                            g,
                                            h,
                                            ghTable,
                                            cacheTable);
        if factors == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            Cudd_RecursiveDeref(dd, g);
            Cudd_RecursiveDeref(dd, h);
        }
        return(factors); 
    }

    // inlining of macro Cudd_Regular
    H = ((DdNode *)((ptruint)(factorsNv.h) & ~01));
    // inlining of macro cuddT
    Hv = ((H).type.kids.T);
    // inlining of macro cuddE
    Hnv = ((H).type.kids.E);
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Hv = ((DdNode *)((ptrint)(Hv) ^ ((int32( ((ptrint) (node) & 01))))));
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Hnv = ((DdNode *)((ptrint)(Hnv) ^ ((int32( ((ptrint) (node) & 01))))));
    /* if the child below is a variable */
    if (Hv == zero) || (Hnv == zero) {
        g = factorsNv.g;
        h = cuddBddAndRecur(dd, x, factorsNv.h);
        // inlining of macro cuddRef
        if h!= nil { cuddSatInc(Cudd_Regular(h).ref); }
        Cudd_RecursiveDeref(dd, factorsNv.h);
        Cudd_RecursiveDeref(dd, x);
        if h == nil {
            Cudd_RecursiveDeref(dd, factorsNv.g);
            return nil;
        }
        /* CheckTablesCacheAndReturn responsible for allocating
         * factors structure.g,h referenced for table store 
         */
        factors = CheckTablesCacheAndReturn(node,
                                            g,
                                            h,
                                            ghTable,
                                            cacheTable);
        if factors == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            Cudd_RecursiveDeref(dd, g);
            Cudd_RecursiveDeref(dd, h);
        }
        return(factors); 
    }

    /* build g1 = x*g; h1 = h */
    /* build g2 = g; h2 = x*h */
    Cudd_RecursiveDeref(dd, x);
    h1 = factorsNv.h;
    g1 = cuddBddAndRecur(dd, x, factorsNv.g);
    // inlining of macro cuddRef
    if g1 != nil { cuddSatInc(Cudd_Regular(g1).ref); }
    if g1 == nil {
        Cudd_RecursiveDeref(dd, factorsNv.g); 
        Cudd_RecursiveDeref(dd, factorsNv.h);
        return nil;
    }
    
    g2 = factorsNv.g;
    h2 = cuddBddAndRecur(dd, x, factorsNv.h);
    // inlining of macro cuddRef
    if h2 != nil { cuddSatInc(Cudd_Regular(h2).ref); }
    if h2 == nil {
        Cudd_RecursiveDeref(dd, factorsNv.h);
        Cudd_RecursiveDeref(dd, factorsNv.g);
        return nil;
    }

    /* check whether any pair is in tables */
    factors = CheckInTables(node, g1, h1, g2, h2, ghTable, cacheTable, &outOfMem);
    if outOfMem {
        dd.errorCode = CUDD_MEMORY_OUT;
        Cudd_RecursiveDeref(dd, g1);
        Cudd_RecursiveDeref(dd, h1);
        Cudd_RecursiveDeref(dd, g2);
        Cudd_RecursiveDeref(dd, h2);
        return nil;
    }
    if factors != nil {
        if (factors.g == g1) || (factors.g == h1) {
            Cudd_RecursiveDeref(dd, g2);
            Cudd_RecursiveDeref(dd, h2);
        } else {
            Cudd_RecursiveDeref(dd, g1);
            Cudd_RecursiveDeref(dd, h1);
        }
        return factors;
    }

    /* check for each pair in tables and choose one */
    factors = PickOnePair(node,g1, h1, g2, h2, ghTable, cacheTable);
    if factors == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        Cudd_RecursiveDeref(dd, g1);
        Cudd_RecursiveDeref(dd, h1);
        Cudd_RecursiveDeref(dd, g2);
        Cudd_RecursiveDeref(dd, h2);
    } else {
        /* now free what was created and not used */
        if (factors.g == g1) || (factors.g == h1) {
            Cudd_RecursiveDeref(dd, g2);
            Cudd_RecursiveDeref(dd, h2);
        } else {
            Cudd_RecursiveDeref(dd, g1);
            Cudd_RecursiveDeref(dd, h1);
        }
    }
        
    return(factors);
} /* end of ZeroCase */


/**Function********************************************************************

  Synopsis    [Builds the conjuncts recursively, bottom up.]

  Description [Builds the conjuncts recursively, bottom up. Constants
  are returned as (f, f). The cache is checked for previously computed
  result. The decomposition points are determined by the local
  reference count of this node and the longest distance from the
  constant. At the decomposition point, the factors returned are (f,
  1). Recur on the two children. The order is determined by the
  heavier branch. Combine the factors of the two children and pick the
  one that already occurs in the gh table. Occurence in g is indicated
  by value 1, occurence in h by 2, occurence in both 3.]

  SideEffects []

  SeeAlso     [cuddConjunctsAux]

******************************************************************************/
func BuildConjuncts(dd *DdManager, node *DdNode, distanceTable *st__table, cacheTable *st__table, approxDistance int32, maxLocalRef int32, ghTable *st__table, mintermTable *st__table) *Conjuncts {
    var topid, distance int32;
    var factorsNv *Conjuncts = nil;
    var factorsNnv *Conjuncts = nil;
    var factors *Conjuncts;
    var dummy *Conjuncts;
    var N, Nv, Nnv, temp, g1, g2, h1, h2, topv *DdNode;
    var minNv float64 = 0.0;
    var minNnv float64 = 0.0;
    var doubleDummy *float64;
    var switched int32;
    var outOfMem int32;
    var freeNv, freeNnv, freeTemp int32;
    var nodeStat *NodeStat;
    var value int32;

    /* if f is constant, return (f,f) */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(node) & ~01))).index == CUDD_CONST_INDEX {
        factors = new(Conjuncts);
        if factors == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(nil);
        }
        factors.g = node;
        factors.h = node;
        // inlining of macro FactorsComplement
        return(((Conjuncts *)(int64((factors)) | 01)));
    }

    /* If result (a pair of conjuncts) in cache, return the factors. */
    if st__lookup(cacheTable, (char *)node, (char **)&dummy) {
        factors = dummy;
        return(factors);
    }
    
    /* check distance and local reference count of this node */
    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));
    if ! st__lookup(distanceTable, (char *)N, (char **)&nodeStat) {
        (void) fmt.Fprintf(dd.err, "Not in table, Something wrong\n");
        dd.errorCode = CUDD_INTERNAL_ERROR;
        return(nil);
    }
    distance = nodeStat.distance;

    /* at or below decomposition point, return (f, 1) */
    if ((nodeStat.localRef > maxLocalRef*2/3) &&
         (distance < approxDistance*2/3)) ||
            (distance <= approxDistance/4) {
        factors = new(Conjuncts);
        if factors == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(nil);
        }
        /* alternate assigning (f,1) */
        value = 0;
        // inlining of macro Cudd_Regular
        if st__lookup_int(ghTable, (char *)((DdNode *)((ptruint)(node) & ~01)), &value) {
            if value == 3 {
                if !lastTimeG {
                    factors.g = node;
                    factors.h = one;
                    lastTimeG = 1;
                } else {
                    factors.g = one;
                    factors.h = node;
                    lastTimeG = 0; 
                }
            } else if value == 1 {
                factors.g = node;
                factors.h = one;
            } else {
                factors.g = one;
                factors.h = node;
            }
        } else if !lastTimeG {
            factors.g = node;
            factors.h = one;
            lastTimeG = 1;
            value = 1;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(node) & ~01)), (char *)int64(value)) == st__OUT_OF_MEM {
                dd.errorCode = CUDD_MEMORY_OUT;
                ABC_FREE(factors);
                return nil;
            }
        } else {
            factors.g = one;
            factors.h = node;
            lastTimeG = 0;
            value = 2;
            // inlining of macro Cudd_Regular
            if st__insert(ghTable, (char *)((DdNode *)((ptruint)(node) & ~01)), (char *)int64(value)) == st__OUT_OF_MEM {
                dd.errorCode = CUDD_MEMORY_OUT;
                ABC_FREE(factors);
                return nil;
            }
        }
        // inlining of macro FactorsComplement
        return(((Conjuncts *)(int64((factors)) | 01)));
    }
    
    /* get the children and recur */
    // inlining of macro cuddT
    Nv = ((N).type.kids.T);
    // inlining of macro cuddE
    Nnv = ((N).type.kids.E);
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Nv = ((DdNode *)((ptrint)(Nv) ^ ((int32( ((ptrint) (node) & 01))))));
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    Nnv = ((DdNode *)((ptrint)(Nnv) ^ ((int32( ((ptrint) (node) & 01))))));

    /* Choose which subproblem to solve first based on the number of
     * minterms. We go first where there are more minterms.
     */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Nv) & ~01))).index == CUDD_CONST_INDEX) {
        if ! st__lookup(mintermTable, (char *)Nv, (char **)&doubleDummy) {
            (void) fmt.Fprintf(dd.err, "Not in table: Something wrong\n");
            dd.errorCode = CUDD_INTERNAL_ERROR;
            return(nil);
        }
        minNv = *doubleDummy;
    }
    
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Nnv) & ~01))).index == CUDD_CONST_INDEX) {
        if ! st__lookup(mintermTable, (char *)Nnv, (char **)&doubleDummy) {
            (void) fmt.Fprintf(dd.err, "Not in table: Something wrong\n");
            dd.errorCode = CUDD_INTERNAL_ERROR;
            return(nil);
        }
        minNnv = *doubleDummy;
    }
    
    if minNv < minNnv {
        temp = Nv;
        Nv = Nnv;
        Nnv = temp;
        switched = 1;
    }

    /* build gt, ht recursively */
    if Nv != zero {
        factorsNv = BuildConjuncts(dd, Nv, distanceTable,
                                   cacheTable, approxDistance, maxLocalRef, 
                                   ghTable, mintermTable);
        if factorsNv == nil { return(nil); }
        // inlining of macro FactorsNotStored
        freeNv = (int32((int64((factorsNv)) & 01)));
        if freeNv {
            // inlining of macro FactorsUncomplement
            factorsNv = ((Conjuncts *)(int64((factorsNv)) ^ 01));
        } else {
            factorsNv = factorsNv;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(factorsNv.g).ref);
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(factorsNv.h).ref);
        
        /* Deal with the zero case */
        if Nnv == zero {
            /* is responsible for freeing factorsNv */
            factors = ZeroCase(dd, node, factorsNv, ghTable,
                               cacheTable, switched);
            if freeNv { ABC_FREE(factorsNv); }
            return(factors);
        }
    }

    /* build ge, he recursively */
    if Nnv != zero {
        factorsNnv = BuildConjuncts(dd, Nnv, distanceTable,
                                    cacheTable, approxDistance, maxLocalRef,
                                    ghTable, mintermTable);
        if factorsNnv == nil {
            Cudd_RecursiveDeref(dd, factorsNv.g);
            Cudd_RecursiveDeref(dd, factorsNv.h);
            if freeNv { ABC_FREE(factorsNv); }
            return(nil);
        }
        // inlining of macro FactorsNotStored
        freeNnv = (int32((int64((factorsNnv)) & 01)));
        if freeNnv {
            // inlining of macro FactorsUncomplement
            factorsNnv = ((Conjuncts *)(int64((factorsNnv)) ^ 01));
        } else {
            factorsNnv = factorsNnv;
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(factorsNnv.g).ref);
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(factorsNnv.h).ref);
        
        /* Deal with the zero case */
        if Nv == zero {
            /* is responsible for freeing factorsNv */
            factors = ZeroCase(dd, node, factorsNnv, ghTable,
                               cacheTable, switched);
            if freeNnv { ABC_FREE(factorsNnv); }
            return(factors);
        }
    }

    /* construct the 2 pairs */
    /* g1 = x*gt + x'*ge; h1 = x*ht + x'*he; */
    /* g2 = x*gt + x'*he; h2 = x*ht + x'*ge */
    if switched != 0 {
        factors = factorsNnv;
        factorsNnv = factorsNv;
        factorsNv = factors;
        freeTemp = freeNv;
        freeNv = freeNnv;
        freeNnv = freeTemp;
    }

    /* Build the factors for this node. */
    topid = N.index;
    topv = dd.vars[topid];
    
    g1 = cuddBddIteRecur(dd, topv, factorsNv.g, factorsNnv.g);
    if g1 == nil {
        Cudd_RecursiveDeref(dd, factorsNv.g);
        Cudd_RecursiveDeref(dd, factorsNv.h);
        Cudd_RecursiveDeref(dd, factorsNnv.g);
        Cudd_RecursiveDeref(dd, factorsNnv.h);
        if freeNv { ABC_FREE(factorsNv); }
        if freeNnv { ABC_FREE(factorsNnv); }
        return(nil);
    }

    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(g1).ref);

    h1 = cuddBddIteRecur(dd, topv, factorsNv.h, factorsNnv.h);
    if h1 == nil {
        Cudd_RecursiveDeref(dd, factorsNv.g);
        Cudd_RecursiveDeref(dd, factorsNv.h);
        Cudd_RecursiveDeref(dd, factorsNnv.g);
        Cudd_RecursiveDeref(dd, factorsNnv.h);
        Cudd_RecursiveDeref(dd, g1);
        if freeNv { ABC_FREE(factorsNv); }
        if freeNnv { ABC_FREE(factorsNnv); }
        return(nil);
    }

    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(h1).ref);

    g2 = cuddBddIteRecur(dd, topv, factorsNv.g, factorsNnv.h);
    if g2 == nil {
        Cudd_RecursiveDeref(dd, factorsNv.h);
        Cudd_RecursiveDeref(dd, factorsNv.g);
        Cudd_RecursiveDeref(dd, factorsNnv.g);
        Cudd_RecursiveDeref(dd, factorsNnv.h);
        Cudd_RecursiveDeref(dd, g1);
        Cudd_RecursiveDeref(dd, h1);
        if freeNv { ABC_FREE(factorsNv); }
        if freeNnv { ABC_FREE(factorsNnv); }
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(g2).ref);
    Cudd_RecursiveDeref(dd, factorsNv.g);
    Cudd_RecursiveDeref(dd, factorsNnv.h);

    h2 = cuddBddIteRecur(dd, topv, factorsNv.h, factorsNnv.g);
    if h2 == nil {
        Cudd_RecursiveDeref(dd, factorsNv.g);
        Cudd_RecursiveDeref(dd, factorsNv.h);
        Cudd_RecursiveDeref(dd, factorsNnv.g);
        Cudd_RecursiveDeref(dd, factorsNnv.h);
        Cudd_RecursiveDeref(dd, g1);
        Cudd_RecursiveDeref(dd, h1);
        Cudd_RecursiveDeref(dd, g2);
        if freeNv { ABC_FREE(factorsNv); }
        if freeNnv { ABC_FREE(factorsNnv); }
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(h2).ref);
    Cudd_RecursiveDeref(dd, factorsNv.h);
    Cudd_RecursiveDeref(dd, factorsNnv.g);
    if freeNv { ABC_FREE(factorsNv); }
    if freeNnv { ABC_FREE(factorsNnv); }

    /* check for each pair in tables and choose one */
    factors = CheckInTables(node, g1, h1, g2, h2, ghTable, cacheTable, &outOfMem);
    if outOfMem {
        dd.errorCode = CUDD_MEMORY_OUT;
        Cudd_RecursiveDeref(dd, g1);
        Cudd_RecursiveDeref(dd, h1);
        Cudd_RecursiveDeref(dd, g2);
        Cudd_RecursiveDeref(dd, h2);
        return(nil);
    }
    if factors != nil {
        if (factors.g == g1) || (factors.g == h1) {
            Cudd_RecursiveDeref(dd, g2);
            Cudd_RecursiveDeref(dd, h2);
        } else {
            Cudd_RecursiveDeref(dd, g1);
            Cudd_RecursiveDeref(dd, h1);
        }
        return(factors);
    }

    /* if not in tables, pick one pair */
    factors = PickOnePair(node,g1, h1, g2, h2, ghTable, cacheTable);
    if factors == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        Cudd_RecursiveDeref(dd, g1);
        Cudd_RecursiveDeref(dd, h1);
        Cudd_RecursiveDeref(dd, g2);
        Cudd_RecursiveDeref(dd, h2);
    } else {
        /* now free what was created and not used */
        if (factors.g == g1) || (factors.g == h1) {
            Cudd_RecursiveDeref(dd, g2);
            Cudd_RecursiveDeref(dd, h2);
        } else {
            Cudd_RecursiveDeref(dd, g1);
            Cudd_RecursiveDeref(dd, h1);
        }
    }
        
    return(factors);
    
} /* end of BuildConjuncts */


/**Function********************************************************************

  Synopsis    [Procedure to compute two conjunctive factors of f and place in *c1 and *c2.]

  Description [Procedure to compute two conjunctive factors of f and
  place in *c1 and *c2. Sets up the required data - table of distances
  from the constant and local reference count. Also minterm table. ]

  SideEffects []

  SeeAlso     []

******************************************************************************/
func cuddConjunctsAux(dd *DdManager, f *DdNode, c1 **DdNode, c2 **DdNode) int32 {
    var distanceTable *st__table  = nil;
    var cacheTable *st__table  = nil;
    var mintermTable *st__table  = nil;
    var ghTable *st__table  = nil;
    var stGen *st__generator;
    var key, value string;
    var factors *Conjuncts;
    var distance, approxDistance int32;
    var max, minterms float64;
    var freeFactors int32;
    var nodeStat *NodeStat;
    var maxLocalRef int32;
    
    /* initialize */
    *c1 = nil;
    *c2 = nil;

    /* initialize distances table */
    distanceTable = st__init_table( st__ptrcmp, st__ptrhash);
    if distanceTable == nil { goto outOfMem; }
    
    /* make the entry for the constant */
    nodeStat = new(NodeStat);
    if nodeStat == nil { goto outOfMem; }
    nodeStat.distance = 0;
    nodeStat.localRef = 1;
    if st__insert(distanceTable, (char *)one, (char *)nodeStat) == st__OUT_OF_MEM {
        goto outOfMem;
    }

    /* Count node distances from constant. */
    nodeStat = CreateBotDist(f, distanceTable);
    if nodeStat == nil { goto outOfMem; }

    /* set the distance for the decomposition points */
    if DEPTH < nodeStat.distance {
        approxDistance = nodeStat.distance;
    } else {
        approxDistance = DEPTH;
    }
    distance = nodeStat.distance;

    if distance < approxDistance {
        /* Too small to bother. */
        *c1 = f;
        // inlining of macro DD_ONE
        *c2 = ((dd).one);
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(*c1).ref); cuddSatInc(Cudd_Regular(*c2).ref);
        stGen = st__init_gen(distanceTable);
        if stGen == nil { goto outOfMem; }
        for  st__gen(stGen, (char **)&key, (char **)&value) {
            ABC_FREE(value);
        }
        st__free_gen(stGen); stGen = nil;
        st__free_table(distanceTable);
        return(1);
    }

    /* record the maximum local reference count */
    maxLocalRef = 0;
    stGen = st__init_gen(distanceTable);
    if stGen == nil { goto outOfMem; }
    for  st__gen(stGen, (char **)&key, (char **)&value) {
        nodeStat = (NodeStat *)value;
        if nodeStat.localRef > maxLocalRef {
            maxLocalRef = nodeStat.localRef;
        } else {
            maxLocalRef = maxLocalRef;
        }
    }
    st__free_gen(stGen); stGen = nil;

            
    /* Count minterms for each node. */
    max = math.Pow(float64(2.0), float64(float64(Cudd_SupportSize(dd,f)))); /* potential overflow */
    mintermTable = st__init_table( st__ptrcmp, st__ptrhash);
    if mintermTable == nil { goto outOfMem; }
    minterms = CountMinterms(f, max, mintermTable, dd.err);
    if minterms == -1.0 { goto outOfMem; }
    
    lastTimeG = Cudd_Random() & 1;
    cacheTable = st__init_table( st__ptrcmp, st__ptrhash);
    if cacheTable == nil { goto outOfMem; }
    ghTable = st__init_table( st__ptrcmp, st__ptrhash);
    if ghTable == nil { goto outOfMem; }

    /* Build conjuncts. */
    factors = BuildConjuncts(dd, f, distanceTable, cacheTable,
                             approxDistance, maxLocalRef, ghTable, mintermTable);
    if factors == nil { goto outOfMem; }

    /* free up tables */
    stGen = st__init_gen(distanceTable);
    if stGen == nil { goto outOfMem; }
    for  st__gen(stGen, (char **)&key, (char **)&value) {
        ABC_FREE(value);
    }
    st__free_gen(stGen); stGen = nil;
    st__free_table(distanceTable); distanceTable = nil;
    st__free_table(ghTable); ghTable = nil;
    
    stGen = st__init_gen(mintermTable);
    if stGen == nil { goto outOfMem; }
    for  st__gen(stGen, (char **)&key, (char **)&value) {
        ABC_FREE(value);
    }
    st__free_gen(stGen); stGen = nil;
    st__free_table(mintermTable); mintermTable = nil;

    // inlining of macro FactorsNotStored
    freeFactors = (int32((int64((factors)) & 01)));
    if freeFactors != 0 {
        // inlining of macro FactorsUncomplement
        factors = ((Conjuncts *)(int64((factors)) ^ 01));
    } else {
        factors = factors;
    }
    if factors != nil {
        *c1 = factors.g;
        *c2 = factors.h;
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(*c1).ref);
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(*c2).ref);
        if freeFactors != 0 { ABC_FREE(factors); }
        
         
                                                  
                               
         
                                                  
                               
         
        
                                                    
                                               
         
                                                    
                                               
         
      
    }

    stGen = st__init_gen(cacheTable);
    if stGen == nil { goto outOfMem; }
    for  st__gen(stGen, (char **)&key, (char **)&value) {
        ConjunctsFree(dd, (Conjuncts *)value);
    }
    st__free_gen(stGen); stGen = nil;

    st__free_table(cacheTable); cacheTable = nil;

    return(1);

outOfMem:
    if distanceTable != nil {
        stGen = st__init_gen(distanceTable);
        if stGen == nil { goto outOfMem; }
        for  st__gen(stGen, (char **)&key, (char **)&value) {
            ABC_FREE(value);
        }
        st__free_gen(stGen); stGen = nil;
        st__free_table(distanceTable); distanceTable = nil;
    }
    if mintermTable != nil {
        stGen = st__init_gen(mintermTable);
        if stGen == nil { goto outOfMem; }
        for  st__gen(stGen, (char **)&key, (char **)&value) {
            ABC_FREE(value);
        }
        st__free_gen(stGen); stGen = nil;
        st__free_table(mintermTable); mintermTable = nil;
    }
    if ghTable != nil { st__free_table(ghTable); }
    if cacheTable != nil {
        stGen = st__init_gen(cacheTable);
        if stGen == nil { goto outOfMem; }
        for  st__gen(stGen, (char **)&key, (char **)&value) {
            ConjunctsFree(dd, (Conjuncts *)value);
        }
        st__free_gen(stGen); stGen = nil;
        st__free_table(cacheTable); cacheTable = nil;
    }
    dd.errorCode = CUDD_MEMORY_OUT;
    return(0);

} /* end of cuddConjunctsAux */


//ABC_NAMESPACE_IMPL_END


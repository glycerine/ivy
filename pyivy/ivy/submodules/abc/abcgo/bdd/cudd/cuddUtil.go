package cudd

import (
	"fmt"
	"math"
	"strings"
)

import "os"

/**CFile***********************************************************************

  FileName    [cuddUtil.c]

  PackageName [cudd]

  Synopsis    [Utility functions.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_PrintMinterm()
                <li> Cudd_bddPrintCover()
                <li> Cudd_PrintDebug()
                <li> Cudd_DagSize()
                <li> Cudd_EstimateCofactor()
                <li> Cudd_EstimateCofactorSimple()
                <li> Cudd_SharingSize()
                <li> Cudd_CountMinterm()
                <li> Cudd_EpdCountMinterm()
                <li> Cudd_CountPath()
                <li> Cudd_CountPathsToNonZero()
                <li> Cudd_Support()
                <li> Cudd_SupportIndex()
                <li> Cudd_SupportSize()
                <li> Cudd_VectorSupport()
                <li> Cudd_VectorSupportIndex()
                <li> Cudd_VectorSupportSize()
                <li> Cudd_ClassifySupport()
                <li> Cudd_CountLeaves()
                <li> Cudd_bddPickOneCube()
                <li> Cudd_bddPickOneMinterm()
                <li> Cudd_bddPickArbitraryMinterms()
                <li> Cudd_SubsetWithMaskVars()
                <li> Cudd_FirstCube()
                <li> Cudd_NextCube()
                <li> Cudd_bddComputeCube()
                <li> Cudd_addComputeCube()
                <li> Cudd_FirstNode()
                <li> Cudd_NextNode()
                <li> Cudd_GenFree()
                <li> Cudd_IsGenEmpty()
                <li> Cudd_IndicesToCube()
                <li> Cudd_PrintVersion()
                <li> Cudd_AverageDistance()
                <li> Cudd_Random()
                <li> Cudd_Srandom()
                <li> Cudd_Density()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddP()
                <li> cuddStCountfree()
                <li> cuddCollectNodes()
                <li> cuddNodeArray()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> dp2()
                <li> ddPrintMintermAux()
                <li> ddDagInt()
                <li> ddCountMintermAux()
                <li> ddEpdCountMintermAux()
                <li> ddCountPathAux()
                <li> ddSupportStep()
                <li> ddClearFlag()
                <li> ddLeavesInt()
                <li> ddPickArbitraryMinterms()
                <li> ddPickRepresentativeCube()
                <li> ddEpdFree()
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

/* Random generator constants. */
const MODULUS1 = 2147483563
const LEQA1 = 40014
const LEQQ1 = 53668
const LEQR1 = 12211
const MODULUS2 = 2147483399
const LEQA2 = 40692
const LEQQ2 = 52774
const LEQR2 = 3791
const STAB_SIZE = 64
#define STAB_DIV (1 + (MODULUS1 - 1) / STAB_SIZE)

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddUtil.c,v 1.81 2009/03/08 02:49:02 fabio Exp $";
#endif

var background, zero *DdNode;

var cuddRand long;
var cuddRand2 long;
var shuffleSelect long;
var shuffleTable [STAB_SIZE]long;

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

// Original macro:
// #define bang(f) ((Cudd_IsComplement(f)) ? '!' : ' ')


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int dp2 (DdManager *dd, DdNode *f, st__table *t);

// static void ddPrintMintermAux (DdManager *dd, DdNode *node, int *list);

// static int ddDagInt (DdNode *n);

// static int cuddNodeArrayRecur (DdNode *f, DdNodePtr *table, int index);

// static int cuddEstimateCofactor (DdManager *dd, st__table *table, DdNode * node, int i, int phase, DdNode ** ptr);

// static DdNode * cuddUniqueLookup (DdManager * unique, int  index, DdNode * T, DdNode * E);

// static int cuddEstimateCofactorSimple (DdNode * node, int i);

// static double ddCountMintermAux (DdNode *node, double max, DdHashTable *table);

// static int ddEpdCountMintermAux (DdNode *node, EpDouble *max, EpDouble *epd, st__table *table);

// static double ddCountPathAux (DdNode *node, st__table *table);

// static double ddCountPathsToNonZero (DdNode * N, st__table * table);

// static void ddSupportStep (DdNode *f, int *support);

// static void ddClearFlag (DdNode *f);

// static int ddLeavesInt (DdNode *n);

// static int ddPickArbitraryMinterms (DdManager *dd, DdNode *node, int nvars, int nminterms, char **string);

// static int ddPickRepresentativeCube (DdManager *dd, DdNode *node, double *weight, char *string);

// static      st__retval ddEpdFree (char * key, char * value, char * arg);


/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Prints a disjoint sum of products.]

  Description [Prints a disjoint sum of product cover for the function
  rooted at node. Each product corresponds to a path from node to a
  leaf node different from the logical zero, and different from the
  background value. Uses the package default output file.  Returns 1
  if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_PrintDebug Cudd_bddPrintCover]

******************************************************************************/
func Cudd_PrintMinterm(manager *DdManager, node *DdNode) int32 {
    var i int32;
    var list *int32;

    background = manager.background;
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(manager.one) ^ 01));
    list = ABC_ALLOC(int,manager.size);
    if list == nil {
        manager.errorCode = CUDD_MEMORY_OUT;
        return(0);
    }
    for i = 0; i < manager.size; i++ { list[i] = 2; }
    ddPrintMintermAux(manager,node,list);
    ABC_FREE(list);
    return(1);

} /* end of Cudd_PrintMinterm */


/**Function********************************************************************

  Synopsis    [Prints a sum of prime implicants of a BDD.]

  Description [Prints a sum of product cover for an incompletely
  specified function given by a lower bound and an upper bound.  Each
  product is a prime implicant obtained by expanding the product
  corresponding to a path from node to the constant one.  Uses the
  package default output file.  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_PrintMinterm]

******************************************************************************/
func Cudd_bddPrintCover(dd *DdManager, l *DdNode, u *DdNode) int32 {
    var array *int32;
    var q, result int32;
    var lb *DdNode;
#ifdef DD_DEBUG
    DdNode *cover;
#endif

    array = ABC_ALLOC(int, Cudd_ReadSize(dd));
    if array == nil { return(0); }
    lb = l;
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(lb).ref);
#ifdef DD_DEBUG
    cover = Cudd_ReadLogicZero(dd);
    cuddRef(cover);
#endif
    for lb != Cudd_ReadLogicZero(dd) {
        var implicant, prime, tmp *DdNode;
        var length int32;
        implicant = Cudd_LargestCube(dd,lb,&length);
        if implicant == nil {
            Cudd_RecursiveDeref(dd,lb);
            ABC_FREE(array);
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(implicant).ref);
        prime = Cudd_bddMakePrime(dd,implicant,u);
        if prime == nil {
            Cudd_RecursiveDeref(dd,lb);
            Cudd_RecursiveDeref(dd,implicant);
            ABC_FREE(array);
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(prime).ref);
        Cudd_RecursiveDeref(dd,implicant);
        // inlining of macro Cudd_Not
        tmp = Cudd_bddAnd(dd,lb,((DdNode *)((ptrint)(prime) ^ 01)));
        if tmp == nil {
            Cudd_RecursiveDeref(dd,lb);
            Cudd_RecursiveDeref(dd,prime);
            ABC_FREE(array);
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(tmp).ref);
        Cudd_RecursiveDeref(dd,lb);
        lb = tmp;
        result = Cudd_BddToCubeArray(dd,prime,array);
        if result == 0 {
            Cudd_RecursiveDeref(dd,lb);
            Cudd_RecursiveDeref(dd,prime);
            ABC_FREE(array);
            return(0);
        }
        for q = 0; q < dd.size; q++ {
            switch array[q] {
            case 0:
                (void) fmt.Fprintf(dd.out, "0");
                break;
            case 1:
                (void) fmt.Fprintf(dd.out, "1");
                break;
            case 2:
                (void) fmt.Fprintf(dd.out, "-");
                break;
            default:
                (void) fmt.Fprintf(dd.out, "?");
            }
        }
        (void) fmt.Fprintf(dd.out, " 1\n");
#ifdef DD_DEBUG
        tmp = Cudd_bddOr(dd,prime,cover);
        if (tmp == NULL) {
            Cudd_RecursiveDeref(dd,cover);
            Cudd_RecursiveDeref(dd,lb);
            Cudd_RecursiveDeref(dd,prime);
            ABC_FREE(array);
            return(0);
        }
        cuddRef(tmp);
        Cudd_RecursiveDeref(dd,cover);
        cover = tmp;
#endif
        Cudd_RecursiveDeref(dd,prime);
    }
    (void) fmt.Fprintf(dd.out, "\n");
    Cudd_RecursiveDeref(dd,lb);
    ABC_FREE(array);
#ifdef DD_DEBUG
    if (!Cudd_bddLeq(dd,cover,u) || !Cudd_bddLeq(dd,l,cover)) {
        Cudd_RecursiveDeref(dd,cover);
        return(0);
    }
    Cudd_RecursiveDeref(dd,cover);
#endif
    return(1);

} /* end of Cudd_bddPrintCover */


/**Function********************************************************************

  Synopsis    [Prints to the standard output a DD and its statistics.]

  Description [Prints to the standard output a DD and its statistics.
  The statistics include the number of nodes, the number of leaves, and
  the number of minterms. (The number of minterms is the number of
  assignments to the variables that cause the function to be different
  from the logical zero (for BDDs) and from the background value (for
  ADDs.) The statistics are printed if pr &gt; 0. Specifically:
  <ul>
  <li> pr = 0 : prints nothing
  <li> pr = 1 : prints counts of nodes and minterms
  <li> pr = 2 : prints counts + disjoint sum of product
  <li> pr = 3 : prints counts + list of nodes
  <li> pr &gt; 3 : prints counts + disjoint sum of product + list of nodes
  </ul>
  For the purpose of counting the number of minterms, the function is
  supposed to depend on n variables. Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_DagSize Cudd_CountLeaves Cudd_CountMinterm
  Cudd_PrintMinterm]

******************************************************************************/
func Cudd_PrintDebug(dd *DdManager, f *DdNode, n int32, pr int32) int32 {
    var azero, bzero *DdNode;
    var nodes int32;
    var leaves int32;
    var minterms float64;
    var retval int32  = 1;

    if f == nil {
        (void) fmt.Fprintf(dd.out, ": is the NULL DD\n");
        (void) fflush(dd.out);
        return(0);
    }
    // inlining of macro DD_ZERO
    azero = ((dd).zero);
    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    bzero = ((DdNode *)((ptrint)(((dd).one)) ^ 01));
    if (f == azero || f == bzero) && pr > 0{
       (void) fmt.Fprintf(dd.out, ": is the zero DD\n");
       (void) fflush(dd.out);
       return(1);
    }
    if pr > 0 {
        nodes = Cudd_DagSize(f);
        if nodes == CUDD_OUT_OF_MEM { retval = 0; }
        leaves = Cudd_CountLeaves(f);
        if leaves == CUDD_OUT_OF_MEM { retval = 0; }
        minterms = Cudd_CountMinterm(dd, f, n);
        if minterms == float64(CUDD_OUT_OF_MEM) { retval = 0; }
        (void) fmt.Fprintf(dd.out, ": %d nodes %d leaves %g minterms\n", nodes, leaves, minterms);
        if pr > 2 {
            if !(cuddP(dd, f) != 0) { retval = 0; }
        }
        if pr == 2 || pr > 3 {
            if !(Cudd_PrintMinterm(dd,f) != 0) { retval = 0; }
            (void) fmt.Fprintf(dd.out, "\n");
        }
        (void) fflush(dd.out);
    }
    return(retval);

} /* end of Cudd_PrintDebug */


/**Function********************************************************************

  Synopsis    [Counts the number of nodes in a DD.]

  Description [Counts the number of nodes in a DD. Returns the number
  of nodes in the graph rooted at node.]

  SideEffects [None]

  SeeAlso     [Cudd_SharingSize Cudd_PrintDebug]

******************************************************************************/
func Cudd_DagSize(node *DdNode) int32 {
    var i int32;

    // inlining of macro Cudd_Regular
    i = ddDagInt(((DdNode *)((ptruint)(node) & ~01)));
    // inlining of macro Cudd_Regular
    ddClearFlag(((DdNode *)((ptruint)(node) & ~01)));

    return(i);

} /* end of Cudd_DagSize */


/**Function********************************************************************

  Synopsis    [Estimates the number of nodes in a cofactor of a DD.]

  Description [Estimates the number of nodes in a cofactor of a DD.
  Returns an estimate of the number of nodes in a cofactor of
  the graph rooted at node with respect to the variable whose index is i.
  In case of failure, returns CUDD_OUT_OF_MEM.
  This function uses a refinement of the algorithm of Cabodi et al.
  (ICCAD96). The refinement allows the procedure to account for part
  of the recombination that may occur in the part of the cofactor above
  the cofactoring variable. This procedure does no create any new node.
  It does keep a small table of results; therefore it may run out of memory.
  If this is a concern, one should use Cudd_EstimateCofactorSimple, which
  is faster, does not allocate any memory, but is less accurate.]

  SideEffects [None]

  SeeAlso     [Cudd_DagSize Cudd_EstimateCofactorSimple]

******************************************************************************/
int
Cudd_EstimateCofactor(
  DdManager *dd /* manager */,
  DdNode * f    /* function */,
  int i         /* index of variable */,
  int phase     /* 1: positive; 0: negative */
  )
{
    var val int32;
    var ptr *DdNode;
    var table *st__table;

    table = st__init_table( st__ptrcmp, st__ptrhash);
    if table == nil { return(CUDD_OUT_OF_MEM); }
    // inlining of macro Cudd_Regular
    val = cuddEstimateCofactor(dd,table,((DdNode *)((ptruint)(f) & ~01)),i,phase,&ptr);
    // inlining of macro Cudd_Regular
    ddClearFlag(((DdNode *)((ptruint)(f) & ~01)));
    st__free_table(table);

    return(val);

} /* end of Cudd_EstimateCofactor */


/**Function********************************************************************

  Synopsis    [Estimates the number of nodes in a cofactor of a DD.]

  Description [Estimates the number of nodes in a cofactor of a DD.
  Returns an estimate of the number of nodes in the positive cofactor of
  the graph rooted at node with respect to the variable whose index is i.
  This procedure implements with minor changes the algorithm of Cabodi et al.
  (ICCAD96). It does not allocate any memory, it does not change the
  state of the manager, and it is fast. However, it has been observed to
  overestimate the size of the cofactor by as much as a factor of 2.]

  SideEffects [None]

  SeeAlso     [Cudd_DagSize]

******************************************************************************/
func Cudd_EstimateCofactorSimple(node *DdNode, i int32) int32 {
    var val int32;

    // inlining of macro Cudd_Regular
    val = cuddEstimateCofactorSimple(((DdNode *)((ptruint)(node) & ~01)),i);
    // inlining of macro Cudd_Regular
    ddClearFlag(((DdNode *)((ptruint)(node) & ~01)));

    return(val);

} /* end of Cudd_EstimateCofactorSimple */


/**Function********************************************************************

  Synopsis    [Counts the number of nodes in an array of DDs.]

  Description [Counts the number of nodes in an array of DDs. Shared
  nodes are counted only once.  Returns the total number of nodes.]

  SideEffects [None]

  SeeAlso     [Cudd_DagSize]

******************************************************************************/
func Cudd_SharingSize(nodeArray **DdNode, n int32) int32 {
    var i, j int32;

    i = 0;
    for j = 0; j < n; j++ {
        // inlining of macro Cudd_Regular
        i += ddDagInt(((DdNode *)((ptruint)(nodeArray[j]) & ~01)));
    }
    for j = 0; j < n; j++ {
        // inlining of macro Cudd_Regular
        ddClearFlag(((DdNode *)((ptruint)(nodeArray[j]) & ~01)));
    }
    return(i);

} /* end of Cudd_SharingSize */


/**Function********************************************************************

  Synopsis    [Counts the number of minterms of a DD.]

  Description [Counts the number of minterms of a DD. The function is
  assumed to depend on nvars variables. The minterm count is
  represented as a double, to allow for a larger number of variables.
  Returns the number of minterms of the function rooted at node if
  successful; (double) CUDD_OUT_OF_MEM otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_PrintDebug Cudd_CountPath]

******************************************************************************/
func Cudd_CountMinterm(manager *DdManager, node *DdNode, nvars int32) float64 {
    var max float64;
    var table *DdHashTable;
    var res float64;
    var epsilon CUDD_VALUE_TYPE;

    background = manager.background;
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(manager.one) ^ 01));

    max = math.Pow(float64(2.0), float64(float64(nvars)));
    table = cuddHashTableInit(manager,1,2);
    if table == nil {
        return(float64(CUDD_OUT_OF_MEM));
    }
    epsilon = Cudd_ReadEpsilon(manager);
    Cudd_SetEpsilon(manager,(CUDD_VALUE_TYPE)0.0);
    res = ddCountMintermAux(node,max,table);
    cuddHashTableQuit(table);
    Cudd_SetEpsilon(manager,epsilon);

    return(res);

} /* end of Cudd_CountMinterm */


/**Function********************************************************************

  Synopsis    [Counts the number of paths of a DD.]

  Description [Counts the number of paths of a DD.  Paths to all
  terminal nodes are counted. The path count is represented as a
  double, to allow for a larger number of variables.  Returns the
  number of paths of the function rooted at node if successful;
  (double) CUDD_OUT_OF_MEM otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_CountMinterm]

******************************************************************************/
func Cudd_CountPath(node *DdNode) float64 {

    var table *st__table;
    var i float64;

    table = st__init_table( st__ptrcmp, st__ptrhash);
    if table == nil {
        return(float64(CUDD_OUT_OF_MEM));
    }
    // inlining of macro Cudd_Regular
    i = ddCountPathAux(((DdNode *)((ptruint)(node) & ~01)),table);
    st__foreach(table, cuddStCountfree, nil);
    st__free_table(table);
    return(i);

} /* end of Cudd_CountPath */


/**Function********************************************************************

  Synopsis    [Counts the number of minterms of a DD with extended precision.]

  Description [Counts the number of minterms of a DD with extended precision.
  The function is assumed to depend on nvars variables. The minterm count is
  represented as an EpDouble, to allow any number of variables.
  Returns 0 if successful; CUDD_OUT_OF_MEM otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_PrintDebug Cudd_CountPath]

******************************************************************************/
func Cudd_EpdCountMinterm(manager *DdManager, node *DdNode, nvars int32, epd *EpDouble) int32 {
    var max, tmp EpDouble;
    var table *st__table;
    var status int32;

    background = manager.background;
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(manager.one) ^ 01));

    EpdPow2(nvars, &max);
    table = st__init_table(EpdCmp, st__ptrhash);
    if table == nil {
        EpdMakeZero(epd, 0);
        return(CUDD_OUT_OF_MEM);
    }
    // inlining of macro Cudd_Regular
    status = ddEpdCountMintermAux(((DdNode *)((ptruint)(node) & ~01)),&max,epd,table);
    st__foreach(table, ddEpdFree, nil);
    st__free_table(table);
    if status == CUDD_OUT_OF_MEM {
        EpdMakeZero(epd, 0);
        return(CUDD_OUT_OF_MEM);
    }
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (node) & 01)) != 0 {
        EpdSubtract3(&max, epd, &tmp);
        EpdCopy(&tmp, epd);
    }
    return(0);

} /* end of Cudd_EpdCountMinterm */


/**Function********************************************************************

  Synopsis    [Counts the number of paths to a non-zero terminal of a DD.]

  Description [Counts the number of paths to a non-zero terminal of a
  DD.  The path count is
  represented as a double, to allow for a larger number of variables.
  Returns the number of paths of the function rooted at node.]

  SideEffects [None]

  SeeAlso     [Cudd_CountMinterm Cudd_CountPath]

******************************************************************************/
func Cudd_CountPathsToNonZero(node *DdNode) float64 {

    var table *st__table;
    var i float64;

    table = st__init_table( st__ptrcmp, st__ptrhash);
    if table == nil {
        return(float64(CUDD_OUT_OF_MEM));
    }
    i = ddCountPathsToNonZero(node,table);
    st__foreach(table, cuddStCountfree, nil);
    st__free_table(table);
    return(i);

} /* end of Cudd_CountPathsToNonZero */


/**Function********************************************************************

  Synopsis    [Finds the variables on which a DD depends.]

  Description [Finds the variables on which a DD depends.
  Returns a BDD consisting of the product of the variables if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_VectorSupport Cudd_ClassifySupport]

******************************************************************************/
DdNode *
Cudd_Support(
  DdManager * dd /* manager */,
  DdNode * f /* DD whose support is sought */)
{
    var support *int32;
    var res, tmp, var *DdNode;
    var i, j int32;
    var size int32;

    /* Allocate and initialize support array for ddSupportStep. */
    // inlining of macro ddMax
    
    if (dd.sizeZ) > (dd.size) {
        size = (dd.sizeZ);
    } else {
        size = (dd.size);
    }
    support = ABC_ALLOC(int,size);
    if support == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    for i = 0; i < size; i++ {
        support[i] = 0;
    }

    /* Compute support and clean up markers. */
    // inlining of macro Cudd_Regular
    ddSupportStep(((DdNode *)((ptruint)(f) & ~01)),support);
    // inlining of macro Cudd_Regular
    ddClearFlag(((DdNode *)((ptruint)(f) & ~01)));

    /* Transform support from array to cube. */
    for {
        dd.reordered = 0;
        // inlining of macro DD_ONE
        res = ((dd).one);
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        for j = size - 1; j >= 0; j-- { /* for each level bottom-up */
            if j >= dd.size {
                i = j;
            } else {
                i = dd.invperm[j];
            }
            if support[i] == 1 {
                /* The following call to cuddUniqueInter is guaranteed
                ** not to trigger reordering because the node we look up
                ** already exists. */
                // inlining of macro Cudd_Not
                var = cuddUniqueInter(dd,i,dd.one,((DdNode *)((ptrint)(dd.one) ^ 01)));
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(var).ref);
                tmp = cuddBddAndRecur(dd,res,var);
                if tmp == nil {
                    Cudd_RecursiveDeref(dd,res);
                    Cudd_RecursiveDeref(dd,var);
                    res = nil;
                    break;
                }
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(tmp).ref);
                Cudd_RecursiveDeref(dd,res);
                Cudd_RecursiveDeref(dd,var);
                res = tmp;
            }
        }
    
if !(dd.reordered == 1) { break }
} 

    ABC_FREE(support);
    // inlining of macro cuddDeref
    if res != nil { cuddSatDec(Cudd_Regular(res).ref); }
    return(res);

} /* end of Cudd_Support */


/**Function********************************************************************

  Synopsis    [Finds the variables on which a DD depends.]

  Description [Finds the variables on which a DD depends.  Returns an
  index array of the variables if successful; NULL otherwise.  The
  size of the array equals the number of variables in the manager.
  Each entry of the array is 1 if the corresponding variable is in the
  support of the DD and 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_Support Cudd_VectorSupport Cudd_ClassifySupport]

******************************************************************************/
int *
Cudd_SupportIndex(
  DdManager * dd /* manager */,
  DdNode * f /* DD whose support is sought */)
{
    var support *int32;
    var i int32;
    var size int32;

    /* Allocate and initialize support array for ddSupportStep. */
    // inlining of macro ddMax
    
    if (dd.sizeZ) > (dd.size) {
        size = (dd.sizeZ);
    } else {
        size = (dd.size);
    }
    support = ABC_ALLOC(int,size);
    if support == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    for i = 0; i < size; i++ {
        support[i] = 0;
    }

    /* Compute support and clean up markers. */
    // inlining of macro Cudd_Regular
    ddSupportStep(((DdNode *)((ptruint)(f) & ~01)),support);
    // inlining of macro Cudd_Regular
    ddClearFlag(((DdNode *)((ptruint)(f) & ~01)));

    return(support);

} /* end of Cudd_SupportIndex */


/**Function********************************************************************

  Synopsis    [Counts the variables on which a DD depends.]

  Description [Counts the variables on which a DD depends.
  Returns the number of the variables if successful; CUDD_OUT_OF_MEM
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_Support]

******************************************************************************/
int
Cudd_SupportSize(
  DdManager * dd /* manager */,
  DdNode * f /* DD whose support size is sought */)
{
    var support *int32;
    var i int32;
    var size int32;
    var count int32;

    /* Allocate and initialize support array for ddSupportStep. */
    // inlining of macro ddMax
    
    if (dd.sizeZ) > (dd.size) {
        size = (dd.sizeZ);
    } else {
        size = (dd.size);
    }
    support = ABC_ALLOC(int,size);
    if support == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(CUDD_OUT_OF_MEM);
    }
    for i = 0; i < size; i++ {
        support[i] = 0;
    }

    /* Compute support and clean up markers. */
    // inlining of macro Cudd_Regular
    ddSupportStep(((DdNode *)((ptruint)(f) & ~01)),support);
    // inlining of macro Cudd_Regular
    ddClearFlag(((DdNode *)((ptruint)(f) & ~01)));

    /* Count support variables. */
    count = 0;
    for i = 0; i < size; i++ {
        if support[i] == 1 { count++; }
    }

    ABC_FREE(support);
    return(count);

} /* end of Cudd_SupportSize */


/**Function********************************************************************

  Synopsis    [Finds the variables on which a set of DDs depends.]

  Description [Finds the variables on which a set of DDs depends.
  The set must contain either BDDs and ADDs, or ZDDs.
  Returns a BDD consisting of the product of the variables if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_Support Cudd_ClassifySupport]

******************************************************************************/
DdNode *
Cudd_VectorSupport(
  DdManager * dd /* manager */,
  DdNode ** F /* array of DDs whose support is sought */,
  int  n /* size of the array */)
{
    var support *int32;
    var res, tmp, var *DdNode;
    var i, j int32;
    var size int32;

    /* Allocate and initialize support array for ddSupportStep. */
    // inlining of macro ddMax
    
    if (dd.sizeZ) > (dd.size) {
        size = (dd.sizeZ);
    } else {
        size = (dd.size);
    }
    support = ABC_ALLOC(int,size);
    if support == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    for i = 0; i < size; i++ {
        support[i] = 0;
    }

    /* Compute support and clean up markers. */
    for i = 0; i < n; i++ {
        // inlining of macro Cudd_Regular
        ddSupportStep(((DdNode *)((ptruint)(F[i]) & ~01)),support);
    }
    for i = 0; i < n; i++ {
        // inlining of macro Cudd_Regular
        ddClearFlag(((DdNode *)((ptruint)(F[i]) & ~01)));
    }

    /* Transform support from array to cube. */
    // inlining of macro DD_ONE
    res = ((dd).one);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    for j = size - 1; j >= 0; j-- { /* for each level bottom-up */
        if j >= dd.size {
            i = j;
        } else {
            i = dd.invperm[j];
        }
        if support[i] == 1 {
            // inlining of macro Cudd_Not
            var = cuddUniqueInter(dd,i,dd.one,((DdNode *)((ptrint)(dd.one) ^ 01)));
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(var).ref);
            tmp = Cudd_bddAnd(dd,res,var);
            if tmp == nil {
                Cudd_RecursiveDeref(dd,res);
                Cudd_RecursiveDeref(dd,var);
                ABC_FREE(support);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(tmp).ref);
            Cudd_RecursiveDeref(dd,res);
            Cudd_RecursiveDeref(dd,var);
            res = tmp;
        }
    }

    ABC_FREE(support);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of Cudd_VectorSupport */


/**Function********************************************************************

  Synopsis    [Finds the variables on which a set of DDs depends.]

  Description [Finds the variables on which a set of DDs depends.
  The set must contain either BDDs and ADDs, or ZDDs.
  Returns an index array of the variables if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_SupportIndex Cudd_VectorSupport Cudd_ClassifySupport]

******************************************************************************/
int *
Cudd_VectorSupportIndex(
  DdManager * dd /* manager */,
  DdNode ** F /* array of DDs whose support is sought */,
  int  n /* size of the array */)
{
    var support *int32;
    var i int32;
    var size int32;

    /* Allocate and initialize support array for ddSupportStep. */
    // inlining of macro ddMax
    
    if (dd.sizeZ) > (dd.size) {
        size = (dd.sizeZ);
    } else {
        size = (dd.size);
    }
    support = ABC_ALLOC(int,size);
    if support == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    for i = 0; i < size; i++ {
        support[i] = 0;
    }

    /* Compute support and clean up markers. */
    for i = 0; i < n; i++ {
        // inlining of macro Cudd_Regular
        ddSupportStep(((DdNode *)((ptruint)(F[i]) & ~01)),support);
    }
    for i = 0; i < n; i++ {
        // inlining of macro Cudd_Regular
        ddClearFlag(((DdNode *)((ptruint)(F[i]) & ~01)));
    }

    return(support);

} /* end of Cudd_VectorSupportIndex */


/**Function********************************************************************

  Synopsis    [Counts the variables on which a set of DDs depends.]

  Description [Counts the variables on which a set of DDs depends.
  The set must contain either BDDs and ADDs, or ZDDs.
  Returns the number of the variables if successful; CUDD_OUT_OF_MEM
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_VectorSupport Cudd_SupportSize]

******************************************************************************/
int
Cudd_VectorSupportSize(
  DdManager * dd /* manager */,
  DdNode ** F /* array of DDs whose support is sought */,
  int  n /* size of the array */)
{
    var support *int32;
    var i int32;
    var size int32;
    var count int32;

    /* Allocate and initialize support array for ddSupportStep. */
    // inlining of macro ddMax
    
    if (dd.sizeZ) > (dd.size) {
        size = (dd.sizeZ);
    } else {
        size = (dd.size);
    }
    support = ABC_ALLOC(int,size);
    if support == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(CUDD_OUT_OF_MEM);
    }
    for i = 0; i < size; i++ {
        support[i] = 0;
    }

    /* Compute support and clean up markers. */
    for i = 0; i < n; i++ {
        // inlining of macro Cudd_Regular
        ddSupportStep(((DdNode *)((ptruint)(F[i]) & ~01)),support);
    }
    for i = 0; i < n; i++ {
        // inlining of macro Cudd_Regular
        ddClearFlag(((DdNode *)((ptruint)(F[i]) & ~01)));
    }

    /* Count vriables in support. */
    count = 0;
    for i = 0; i < size; i++ {
        if support[i] == 1 { count++; }
    }

    ABC_FREE(support);
    return(count);

} /* end of Cudd_VectorSupportSize */


/**Function********************************************************************

  Synopsis    [Classifies the variables in the support of two DDs.]

  Description [Classifies the variables in the support of two DDs
  <code>f</code> and <code>g</code>, depending on whther they appear
  in both DDs, only in <code>f</code>, or only in <code>g</code>.
  Returns 1 if successful; 0 otherwise.]

  SideEffects [The cubes of the three classes of variables are
  returned as side effects.]

  SeeAlso     [Cudd_Support Cudd_VectorSupport]

******************************************************************************/
int
Cudd_ClassifySupport(
  DdManager * dd /* manager */,
  DdNode * f /* first DD */,
  DdNode * g /* second DD */,
  DdNode ** common /* cube of shared variables */,
  DdNode ** onlyF /* cube of variables only in f */,
  DdNode ** onlyG /* cube of variables only in g */)
{
    var supportF, supportG *int32;
    var tmp, var *DdNode;
    var i, j int32;
    var size int32;

    /* Allocate and initialize support arrays for ddSupportStep. */
    // inlining of macro ddMax
    
    if (dd.sizeZ) > (dd.size) {
        size = (dd.sizeZ);
    } else {
        size = (dd.size);
    }
    supportF = ABC_ALLOC(int,size);
    if supportF == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(0);
    }
    supportG = ABC_ALLOC(int,size);
    if supportG == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        ABC_FREE(supportF);
        return(0);
    }
    for i = 0; i < size; i++ {
        supportF[i] = 0;
        supportG[i] = 0;
    }

    /* Compute supports and clean up markers. */
    // inlining of macro Cudd_Regular
    ddSupportStep(((DdNode *)((ptruint)(f) & ~01)),supportF);
    // inlining of macro Cudd_Regular
    ddClearFlag(((DdNode *)((ptruint)(f) & ~01)));
    // inlining of macro Cudd_Regular
    ddSupportStep(((DdNode *)((ptruint)(g) & ~01)),supportG);
    // inlining of macro Cudd_Regular
    ddClearFlag(((DdNode *)((ptruint)(g) & ~01)));

    /* Classify variables and create cubes. */
    // inlining of macro DD_ONE
    *common = *onlyF = *onlyG = ((dd).one);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(*common).ref); cuddSatInc(Cudd_Regular(*onlyF).ref); cuddSatInc(Cudd_Regular(*onlyG).ref);
    for j = size - 1; j >= 0; j-- { /* for each level bottom-up */
        if j >= dd.size {
            i = j;
        } else {
            i = dd.invperm[j];
        }
        if supportF[i] == 0 && supportG[i] == 0 { continue; }
        // inlining of macro Cudd_Not
        var = cuddUniqueInter(dd,i,dd.one,((DdNode *)((ptrint)(dd.one) ^ 01)));
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(var).ref);
        if supportG[i] == 0 {
            tmp = Cudd_bddAnd(dd,*onlyF,var);
            if tmp == nil {
                Cudd_RecursiveDeref(dd,*common);
                Cudd_RecursiveDeref(dd,*onlyF);
                Cudd_RecursiveDeref(dd,*onlyG);
                Cudd_RecursiveDeref(dd,var);
                ABC_FREE(supportF); ABC_FREE(supportG);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(tmp).ref);
            Cudd_RecursiveDeref(dd,*onlyF);
            *onlyF = tmp;
        } else if supportF[i] == 0 {
            tmp = Cudd_bddAnd(dd,*onlyG,var);
            if tmp == nil {
                Cudd_RecursiveDeref(dd,*common);
                Cudd_RecursiveDeref(dd,*onlyF);
                Cudd_RecursiveDeref(dd,*onlyG);
                Cudd_RecursiveDeref(dd,var);
                ABC_FREE(supportF); ABC_FREE(supportG);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(tmp).ref);
            Cudd_RecursiveDeref(dd,*onlyG);
            *onlyG = tmp;
        } else {
            tmp = Cudd_bddAnd(dd,*common,var);
            if tmp == nil {
                Cudd_RecursiveDeref(dd,*common);
                Cudd_RecursiveDeref(dd,*onlyF);
                Cudd_RecursiveDeref(dd,*onlyG);
                Cudd_RecursiveDeref(dd,var);
                ABC_FREE(supportF); ABC_FREE(supportG);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(tmp).ref);
            Cudd_RecursiveDeref(dd,*common);
            *common = tmp;
        }
        Cudd_RecursiveDeref(dd,var);
    }

    ABC_FREE(supportF); ABC_FREE(supportG);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(*common).ref); cuddSatDec(Cudd_Regular(*onlyF).ref); cuddSatDec(Cudd_Regular(*onlyG).ref);
    return(1);

} /* end of Cudd_ClassifySupport */


/**Function********************************************************************

  Synopsis    [Counts the number of leaves in a DD.]

  Description [Counts the number of leaves in a DD. Returns the number
  of leaves in the DD rooted at node if successful; CUDD_OUT_OF_MEM
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_PrintDebug]

******************************************************************************/
func Cudd_CountLeaves(node *DdNode) int32 {
    var i int32;

    // inlining of macro Cudd_Regular
    i = ddLeavesInt(((DdNode *)((ptruint)(node) & ~01)));
    // inlining of macro Cudd_Regular
    ddClearFlag(((DdNode *)((ptruint)(node) & ~01)));
    return(i);

} /* end of Cudd_CountLeaves */


/**Function********************************************************************

  Synopsis    [Picks one on-set cube randomly from the given DD.]

  Description [Picks one on-set cube randomly from the given DD. The
  cube is written into an array of characters.  The array must have at
  least as many entries as there are variables. Returns 1 if
  successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddPickOneMinterm]

******************************************************************************/
func Cudd_bddPickOneCube(ddm *DdManager, node *DdNode, string string) int32 {
    var N, T, E *DdNode;
    var one, bzero *DdNode;
    var dir byte;
    var i int32;

    if string == nil || node == nil { return(0); }

    /* The constant 0 function has no on-set cubes. */
    // inlining of macro DD_ONE
    one = ((ddm).one);
    // inlining of macro Cudd_Not
    bzero = ((DdNode *)((ptrint)(one) ^ 01));
    if node == bzero { return(0); }

    for i = 0; i < ddm.size; i++ { string[i] = 2; }

    for {

        if node == one { break; }

        // inlining of macro Cudd_Regular
        N = ((DdNode *)((ptruint)(node) & ~01));

        // inlining of macro cuddT
        // inlining of macro cuddE
        T = ((N).type.kids.T); E = ((N).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (node) & 01)) != 0 {
            // inlining of macro Cudd_Not
            T = ((DdNode *)((ptrint)(T) ^ 01)); E = ((DdNode *)((ptrint)(E) ^ 01));
        }
        if T == bzero {
            string[N.index] = 0;
            node = E;
        } else if E == bzero {
            string[N.index] = 1;
            node = T;
        } else {
            dir = byte( ((Cudd_Random() & 0x2000) >> 13));
            string[N.index] = dir;
            if dir {
                node = T;
            } else {
                node = E;
            }
        }
    }
    return(1);

} /* end of Cudd_bddPickOneCube */


/**Function********************************************************************

  Synopsis    [Picks one on-set minterm randomly from the given DD.]

  Description [Picks one on-set minterm randomly from the given
  DD. The minterm is in terms of <code>vars</code>. The array
  <code>vars</code> should contain at least all variables in the
  support of <code>f</code>; if this condition is not met the minterm
  built by this procedure may not be contained in
  <code>f</code>. Builds a BDD for the minterm and returns a pointer
  to it if successful; NULL otherwise. There are three reasons why the
  procedure may fail:
  <ul>
  <li> It may run out of memory;
  <li> the function <code>f</code> may be the constant 0;
  <li> the minterm may not be contained in <code>f</code>.
  </ul>]

  SideEffects [None]

  SeeAlso     [Cudd_bddPickOneCube]

******************************************************************************/
DdNode *
Cudd_bddPickOneMinterm(
  DdManager * dd /* manager */,
  DdNode * f /* function from which to pick one minterm */,
  DdNode ** vars /* array of variables */,
  int  n /* size of <code>vars</code> */)
{
    var string string;
    var i, size int32;
    var indices *int32;
    var result int32;
    var old, neW *DdNode;

    size = dd.size;
    string = ABC_ALLOC(char, size);
    if string == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    indices = ABC_ALLOC(int,n);
    if indices == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        ABC_FREE(string);
        return(nil);
    }

    for i = 0; i < n; i++ {
        indices[i] = vars[i].index;
    }

    result = Cudd_bddPickOneCube(dd,f,string);
    if result == 0 {
        ABC_FREE(string);
        ABC_FREE(indices);
        return(nil);
    }

    /* Randomize choice for don't cares. */
    for i = 0; i < n; i++ {
        if string[indices[i]] == 2 {
            string[indices[i]] = byte( ((Cudd_Random() & 0x20) >> 5)); }
    }

    /* Build result BDD. */
    old = Cudd_ReadOne(dd);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(old).ref);

    for i = n-1; i >= 0; i-- {
        // inlining of macro Cudd_NotCond
        neW = Cudd_bddAnd(dd,old,((DdNode *)((ptrint)(vars[i]) ^ (string[indices[i]]==0))));
        if neW == nil {
            ABC_FREE(string);
            ABC_FREE(indices);
            Cudd_RecursiveDeref(dd,old);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(neW).ref);
        Cudd_RecursiveDeref(dd,old);
        old = neW;
    }

#ifdef DD_DEBUG
    /* Test. */
    if (Cudd_bddLeq(dd,old,f)) {
        cuddDeref(old);
    } else {
        Cudd_RecursiveDeref(dd,old);
        old = NULL;
    }
#else
    cuddDeref(old);
#endif

    ABC_FREE(string);
    ABC_FREE(indices);
    return(old);

}  /* end of Cudd_bddPickOneMinterm */


/**Function********************************************************************

  Synopsis    [Picks k on-set minterms evenly distributed from given DD.]

  Description [Picks k on-set minterms evenly distributed from given DD.
  The minterms are in terms of <code>vars</code>. The array
  <code>vars</code> should contain at least all variables in the
  support of <code>f</code>; if this condition is not met the minterms
  built by this procedure may not be contained in
  <code>f</code>. Builds an array of BDDs for the minterms and returns a
  pointer to it if successful; NULL otherwise. There are three reasons
  why the procedure may fail:
  <ul>
  <li> It may run out of memory;
  <li> the function <code>f</code> may be the constant 0;
  <li> the minterms may not be contained in <code>f</code>.
  </ul>]

  SideEffects [None]

  SeeAlso     [Cudd_bddPickOneMinterm Cudd_bddPickOneCube]

******************************************************************************/
DdNode **
Cudd_bddPickArbitraryMinterms(
  DdManager * dd /* manager */,
  DdNode * f /* function from which to pick k minterms */,
  DdNode ** vars /* array of variables */,
  int  n /* size of <code>vars</code> */,
  int  k /* number of minterms to find */)
{
    var string *string;
    var i, j, l, size int32;
    var indices *int32;
    var result int32;
    var old **DdNode;
    var neW *DdNode;
    var minterms float64;
    var saveString string;
    var saveFlag int32;
    var savePoint int32 = -1;
    var isSame int32;

    minterms = Cudd_CountMinterm(dd,f,n);
    if float64(k) > minterms {
        return(nil);
    }

    size = dd.size;
    string = ABC_ALLOC(char *, k);
    if string == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    for i = 0; i < k; i++ {
        string[i] = ABC_ALLOC(char, size + 1);
        if string[i] == nil {
            for j = 0; j < i; j++ {
                ABC_FREE(string[i]);
            }
            ABC_FREE(string);
            dd.errorCode = CUDD_MEMORY_OUT;
            return(nil);
        }
        for j = 0; j < size; j++ { string[i][j] = '2'; }
        string[i][size] = '\0';
    }
    indices = ABC_ALLOC(int,n);
    if indices == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        for i = 0; i < k; i++ {
            ABC_FREE(string[i]);
        }
        ABC_FREE(string);
        return(nil);
    }

    for i = 0; i < n; i++ {
        indices[i] = vars[i].index;
    }

    result = ddPickArbitraryMinterms(dd,f,n,k,string);
    if result == 0 {
        for i = 0; i < k; i++ {
            ABC_FREE(string[i]);
        }
        ABC_FREE(string);
        ABC_FREE(indices);
        return(nil);
    }

    old = ABC_ALLOC(DdNode *, k);
    if old == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        for i = 0; i < k; i++ {
            ABC_FREE(string[i]);
        }
        ABC_FREE(string);
        ABC_FREE(indices);
        return(nil);
    }
    saveString = ABC_ALLOC(char, size + 1);
    if saveString == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        for i = 0; i < k; i++ {
            ABC_FREE(string[i]);
        }
        ABC_FREE(string);
        ABC_FREE(indices);
        ABC_FREE(old);
        return(nil);
    }
    saveFlag = 0;

    /* Build result BDD array. */
    for i = 0; i < k; i++ {
        isSame = 0;
        if !(saveFlag != 0) {
            for j = i + 1; j < k; j++ {
                if int32(strings.Compare(string[i], string[j])) == 0 {
                    savePoint = i;
                    saveString = string[i];
                    saveFlag = 1;
                    break;
                }
            }
        } else {
            if int32(strings.Compare(string[i], saveString)) == 0 {
                isSame = 1;
            } else {
                saveFlag = 0;
                for j = i + 1; j < k; j++ {
                    if int32(strings.Compare(string[i], string[j])) == 0 {
                        savePoint = i;
                        saveString = string[i];
                        saveFlag = 1;
                        break;
                    }
                }
            }
        }
        /* Randomize choice for don't cares. */
        for j = 0; j < n; j++ {
            if string[i][indices[j]] == '2' {
                string[i][indices[j]] =
                  byte( ((Cudd_Random() & 0x20) ? '1' : '0')); }
        }

        for isSame != 0 {
            isSame = 0;
            for j = savePoint; j < i; j++ {
                if int32(strings.Compare(string[i], string[j])) == 0 {
                    isSame = 1;
                    break;
                }
            }
            if isSame != 0 {
                strcpy(string[i], saveString);
                /* Randomize choice for don't cares. */
                for j = 0; j < n; j++ {
                    if string[i][indices[j]] == '2' {
                        string[i][indices[j]] =
                          byte( ((Cudd_Random() & 0x20) ? '1' : '0')); }
                }
            }
        }

        old[i] = Cudd_ReadOne(dd);
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(old[i]).ref);

        for j = 0; j < n; j++ {
            if string[i][indices[j]] == '0' {
                // inlining of macro Cudd_Not
                neW = Cudd_bddAnd(dd,old[i],((DdNode *)((ptrint)(vars[j]) ^ 01)));
            } else {
                neW = Cudd_bddAnd(dd,old[i],vars[j]);
            }
            if neW == nil {
                ABC_FREE(saveString);
                for l = 0; l < k; l++ {
                    ABC_FREE(string[l]);
                }
                ABC_FREE(string);
                ABC_FREE(indices);
                for l = 0; l <= i; l++ {
                    Cudd_RecursiveDeref(dd,old[l]);
                }
                ABC_FREE(old);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(neW).ref);
            Cudd_RecursiveDeref(dd,old[i]);
            old[i] = neW;
        }

        /* Test. */
        if !Cudd_bddLeq(dd,old[i],f) {
            ABC_FREE(saveString);
            for l = 0; l < k; l++ {
                ABC_FREE(string[l]);
            }
            ABC_FREE(string);
            ABC_FREE(indices);
            for l = 0; l <= i; l++ {
                Cudd_RecursiveDeref(dd,old[l]);
            }
            ABC_FREE(old);
            return(nil);
        }
    }

    ABC_FREE(saveString);
    for i = 0; i < k; i++ {
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(old[i]).ref);
        ABC_FREE(string[i]);
    }
    ABC_FREE(string);
    ABC_FREE(indices);
    return(old);

}  /* end of Cudd_bddPickArbitraryMinterms */


/**Function********************************************************************

  Synopsis    [Extracts a subset from a BDD.]

  Description [Extracts a subset from a BDD in the following procedure.
  1. Compute the weight for each mask variable by counting the number of
     minterms for both positive and negative cofactors of the BDD with
     respect to each mask variable. (weight = #positive - #negative)
  2. Find a representative cube of the BDD by using the weight. From the
     top variable of the BDD, for each variable, if the weight is greater
     than 0.0, choose THEN branch, othereise ELSE branch, until meeting
     the constant 1.
  3. Quantify out the variables not in maskVars from the representative
     cube and if a variable in maskVars is don't care, replace the
     variable with a constant(1 or 0) depending on the weight.
  4. Make a subset of the BDD by multiplying with the modified cube.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
DdNode *
Cudd_SubsetWithMaskVars(
  DdManager * dd /* manager */,
  DdNode * f /* function from which to pick a cube */,
  DdNode ** vars /* array of variables */,
  int  nvars /* size of <code>vars</code> */,
  DdNode ** maskVars /* array of variables */,
  int  mvars /* size of <code>maskVars</code> */)
{
    var weight *float64;
    var string string;
    var i, size int32;
    var indices, mask *int32;
    var result int32;
    var zero, cube, newCube, subset *DdNode;
    var cof *DdNode;

    var support *DdNode;
    support = Cudd_Support(dd,f);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(support).ref);
    Cudd_RecursiveDeref(dd,support);

    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(dd.one) ^ 01));
    size = dd.size;

    weight = ABC_ALLOC(double,size);
    if weight == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    for i = 0; i < size; i++ {
        weight[i] = 0.0;
    }
    for i = 0; i < mvars; i++ {
        cof = Cudd_Cofactor(dd, f, maskVars[i]);
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(cof).ref);
        weight[i] = Cudd_CountMinterm(dd, cof, nvars);
        Cudd_RecursiveDeref(dd,cof);

        // inlining of macro Cudd_Not
        cof = Cudd_Cofactor(dd, f, ((DdNode *)((ptrint)(maskVars[i]) ^ 01)));
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(cof).ref);
        weight[i] -= Cudd_CountMinterm(dd, cof, nvars);
        Cudd_RecursiveDeref(dd,cof);
    }

    string = ABC_ALLOC(char, size + 1);
    if string == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        ABC_FREE(weight);
        return(nil);
    }
    mask = ABC_ALLOC(int, size);
    if mask == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        ABC_FREE(weight);
        ABC_FREE(string);
        return(nil);
    }
    for i = 0; i < size; i++ {
        string[i] = '2';
        mask[i] = 0;
    }
    string[size] = '\0';
    indices = make([]int32, nvars);
    if indices == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        ABC_FREE(weight);
        ABC_FREE(string);
        ABC_FREE(mask);
        return(nil);
    }
    for i = 0; i < nvars; i++ {
        indices[i] = vars[i].index;
    }

    result = ddPickRepresentativeCube(dd,f,weight,string);
    if result == 0 {
        ABC_FREE(weight);
        ABC_FREE(string);
        ABC_FREE(mask);
        ABC_FREE(indices);
        return(nil);
    }

    cube = Cudd_ReadOne(dd);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(cube).ref);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(Cudd_ReadOne(dd)) ^ 01));
    for i = 0; i < nvars; i++ {
        if string[indices[i]] == '0' {
            // inlining of macro Cudd_Not
            newCube = Cudd_bddIte(dd,cube,((DdNode *)((ptrint)(vars[i]) ^ 01)),zero);
        } else if string[indices[i]] == '1' {
            newCube = Cudd_bddIte(dd,cube,vars[i],zero);
        } else {
            continue; }
        if newCube == nil {
            ABC_FREE(weight);
            ABC_FREE(string);
            ABC_FREE(mask);
            ABC_FREE(indices);
            Cudd_RecursiveDeref(dd,cube);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(newCube).ref);
        Cudd_RecursiveDeref(dd,cube);
        cube = newCube;
    }
    Cudd_RecursiveDeref(dd,cube);

    for i = 0; i < mvars; i++ {
        mask[maskVars[i].index] = 1;
    }
    for i = 0; i < nvars; i++ {
        if mask[indices[i]] {
            if string[indices[i]] == '2' {
                if weight[indices[i]] >= 0.0 {
                    string[indices[i]] = '1';
                } else {
                    string[indices[i]] = '0'; }
            }
        } else {
            string[indices[i]] = '2';
        }
    }

    cube = Cudd_ReadOne(dd);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(cube).ref);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(Cudd_ReadOne(dd)) ^ 01));

    /* Build result BDD. */
    for i = 0; i < nvars; i++ {
        if string[indices[i]] == '0' {
            // inlining of macro Cudd_Not
            newCube = Cudd_bddIte(dd,cube,((DdNode *)((ptrint)(vars[i]) ^ 01)),zero);
        } else if string[indices[i]] == '1' {
            newCube = Cudd_bddIte(dd,cube,vars[i],zero);
        } else {
            continue; }
        if newCube == nil {
            ABC_FREE(weight);
            ABC_FREE(string);
            ABC_FREE(mask);
            ABC_FREE(indices);
            Cudd_RecursiveDeref(dd,cube);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(newCube).ref);
        Cudd_RecursiveDeref(dd,cube);
        cube = newCube;
    }

    subset = Cudd_bddAnd(dd,f,cube);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(subset).ref);
    Cudd_RecursiveDeref(dd,cube);

    /* Test. */
    if Cudd_bddLeq(dd,subset,f) {
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(subset).ref);
    } else {
        Cudd_RecursiveDeref(dd,subset);
        subset = nil;
    }

    ABC_FREE(weight);
    ABC_FREE(string);
    ABC_FREE(mask);
    ABC_FREE(indices);
    return(subset);

} /* end of Cudd_SubsetWithMaskVars */


/**Function********************************************************************

  Synopsis    [Finds the first cube of a decision diagram.]

  Description [Defines an iterator on the onset of a decision diagram
  and finds its first cube. Returns a generator that contains the
  information necessary to continue the enumeration if successful; NULL
  otherwise.<p>
  A cube is represented as an array of literals, which are integers in
  {0, 1, 2}; 0 represents a complemented literal, 1 represents an
  uncomplemented literal, and 2 stands for don't care. The enumeration
  produces a disjoint cover of the function associated with the diagram.
  The size of the array equals the number of variables in the manager at
  the time Cudd_FirstCube is called.<p>
  For each cube, a value is also returned. This value is always 1 for a
  BDD, while it may be different from 1 for an ADD.
  For BDDs, the offset is the set of cubes whose value is the logical zero.
  For ADDs, the offset is the set of cubes whose value is the
  background value. The cubes of the offset are not enumerated.]

  SideEffects [The first cube and its value are returned as side effects.]

  SeeAlso     [Cudd_ForeachCube Cudd_NextCube Cudd_GenFree Cudd_IsGenEmpty
  Cudd_FirstNode]

******************************************************************************/
func Cudd_FirstCube(dd *DdManager, f *DdNode, cube **int32, value *CUDD_VALUE_TYPE) *DdGen {
    var gen *DdGen;
    var top, treg, next, nreg, prev, preg *DdNode;
    var i int32;
    var nvars int32;

    /* Sanity Check. */
    if dd == nil || f == nil { return(nil); }

    /* Allocate generator an initialize it. */
    gen = new(DdGen);
    if gen == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }

    gen.manager = dd;
    gen.genType = CUDD_GEN_CUBES;
    gen.status = CUDD_GEN_EMPTY;
    gen.data.cubes.cube = nil;
    gen.data.cubes.value = DD_ZERO_VAL;
    gen.stack.sp = 0;
    gen.stack.stack = nil;
    gen.node = nil;

    nvars = dd.size;
    gen.data.cubes.cube = make([]int32, nvars);
    if gen.data.cubes.cube == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        ABC_FREE(gen);
        return(nil);
    }
    for i = 0; i < nvars; i++ { gen.data.cubes.cube[i] = 2; }

    /* The maximum stack depth is one plus the number of variables.
    ** because a path may have nodes at all levels, including the
    ** constant level.
    */
    gen.stack.stack = make([]*DdNode, nvars+1);
    if gen.stack.stack == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        ABC_FREE(gen.data.cubes.cube);
        ABC_FREE(gen);
        return(nil);
    }
    for i = 0; i <= nvars; i++ { gen.stack.stack[i] = nil; }

    /* Find the first cube of the onset. */
    gen.stack.stack[gen.stack.sp] = f; gen.stack.sp++;

    for {
        top = gen.stack.stack[gen.stack.sp-1];
        // inlining of macro Cudd_Regular
        treg = ((DdNode *)((ptruint)(top) & ~01));
        // inlining of macro cuddIsConstant
        if !((treg).index == CUDD_CONST_INDEX) {
            /* Take the else branch first. */
            gen.data.cubes.cube[treg.index] = 0;
            // inlining of macro cuddE
            next = ((treg).type.kids.E);
            // inlining of macro Cudd_Not
            if top != treg { next = ((DdNode *)((ptrint)(next) ^ 01)); }
            gen.stack.stack[gen.stack.sp] = next; gen.stack.sp++;
        // inlining of macro Cudd_Not
        // inlining of macro DD_ONE
        } else if top == ((DdNode *)((ptrint)(((dd).one)) ^ 01)) || top == dd.background {
            /* Backtrack */
            for {
                if gen.stack.sp == 1 {
                    /* The current node has no predecessor. */
                    gen.status = CUDD_GEN_EMPTY;
                    gen.stack.sp--;
                    goto done;
                }
                prev = gen.stack.stack[gen.stack.sp-2];
                // inlining of macro Cudd_Regular
                preg = ((DdNode *)((ptruint)(prev) & ~01));
                // inlining of macro cuddT
                nreg = ((preg).type.kids.T);
                // inlining of macro Cudd_Not
                if prev != preg {next = ((DdNode *)((ptrint)(nreg) ^ 01));} else {next = nreg;}
                if next != top { /* follow the then branch next */
                    gen.data.cubes.cube[preg.index] = 1;
                    gen.stack.stack[gen.stack.sp-1] = next;
                    break;
                }
                /* Pop the stack and try again. */
                gen.data.cubes.cube[preg.index] = 2;
                gen.stack.sp--;
                top = gen.stack.stack[gen.stack.sp-1];
                // inlining of macro Cudd_Regular
                treg = ((DdNode *)((ptruint)(top) & ~01));
            }
        } else {
            gen.status = CUDD_GEN_NONEMPTY;
            // inlining of macro cuddV
            gen.data.cubes.value = ((top).type.value);
            goto done;
        }
    }

done:
    *cube = gen.data.cubes.cube;
    *value = gen.data.cubes.value;
    return(gen);

} /* end of Cudd_FirstCube */


/**Function********************************************************************

  Synopsis    [Generates the next cube of a decision diagram onset.]

  Description [Generates the next cube of a decision diagram onset,
  using generator gen. Returns 0 if the enumeration is completed; 1
  otherwise.]

  SideEffects [The cube and its value are returned as side effects. The
  generator is modified.]

  SeeAlso     [Cudd_ForeachCube Cudd_FirstCube Cudd_GenFree Cudd_IsGenEmpty
  Cudd_NextNode]

******************************************************************************/
func Cudd_NextCube(gen *DdGen, cube **int32, value *CUDD_VALUE_TYPE) int32 {
    var top, treg, next, nreg, prev, preg *DdNode;
    var dd *DdManager  = gen.manager;

    /* Backtrack from previously reached terminal node. */
    for {
        if gen.stack.sp == 1 {
            /* The current node has no predecessor. */
            gen.status = CUDD_GEN_EMPTY;
            gen.stack.sp--;
            goto done;
        }
        top = gen.stack.stack[gen.stack.sp-1];
        // inlining of macro Cudd_Regular
        treg = ((DdNode *)((ptruint)(top) & ~01));
        prev = gen.stack.stack[gen.stack.sp-2];
        // inlining of macro Cudd_Regular
        preg = ((DdNode *)((ptruint)(prev) & ~01));
        // inlining of macro cuddT
        nreg = ((preg).type.kids.T);
        // inlining of macro Cudd_Not
        if prev != preg {next = ((DdNode *)((ptrint)(nreg) ^ 01));} else {next = nreg;}
        if next != top { /* follow the then branch next */
            gen.data.cubes.cube[preg.index] = 1;
            gen.stack.stack[gen.stack.sp-1] = next;
            break;
        }
        /* Pop the stack and try again. */
        gen.data.cubes.cube[preg.index] = 2;
        gen.stack.sp--;
    }

    for {
        top = gen.stack.stack[gen.stack.sp-1];
        // inlining of macro Cudd_Regular
        treg = ((DdNode *)((ptruint)(top) & ~01));
        // inlining of macro cuddIsConstant
        if !((treg).index == CUDD_CONST_INDEX) {
            /* Take the else branch first. */
            gen.data.cubes.cube[treg.index] = 0;
            // inlining of macro cuddE
            next = ((treg).type.kids.E);
            // inlining of macro Cudd_Not
            if top != treg { next = ((DdNode *)((ptrint)(next) ^ 01)); }
            gen.stack.stack[gen.stack.sp] = next; gen.stack.sp++;
        // inlining of macro Cudd_Not
        // inlining of macro DD_ONE
        } else if top == ((DdNode *)((ptrint)(((dd).one)) ^ 01)) || top == dd.background {
            /* Backtrack */
            for {
                if gen.stack.sp == 1 {
                    /* The current node has no predecessor. */
                    gen.status = CUDD_GEN_EMPTY;
                    gen.stack.sp--;
                    goto done;
                }
                prev = gen.stack.stack[gen.stack.sp-2];
                // inlining of macro Cudd_Regular
                preg = ((DdNode *)((ptruint)(prev) & ~01));
                // inlining of macro cuddT
                nreg = ((preg).type.kids.T);
                // inlining of macro Cudd_Not
                if prev != preg {next = ((DdNode *)((ptrint)(nreg) ^ 01));} else {next = nreg;}
                if next != top { /* follow the then branch next */
                    gen.data.cubes.cube[preg.index] = 1;
                    gen.stack.stack[gen.stack.sp-1] = next;
                    break;
                }
                /* Pop the stack and try again. */
                gen.data.cubes.cube[preg.index] = 2;
                gen.stack.sp--;
                top = gen.stack.stack[gen.stack.sp-1];
                // inlining of macro Cudd_Regular
                treg = ((DdNode *)((ptruint)(top) & ~01));
            }
        } else {
            gen.status = CUDD_GEN_NONEMPTY;
            // inlining of macro cuddV
            gen.data.cubes.value = ((top).type.value);
            goto done;
        }
    }

done:
    if gen.status == CUDD_GEN_EMPTY { return(0); }
    *cube = gen.data.cubes.cube;
    *value = gen.data.cubes.value;
    return(1);

} /* end of Cudd_NextCube */


/**Function********************************************************************

  Synopsis    [Finds the first prime of a Boolean function.]

  Description [Defines an iterator on a pair of BDDs describing a
  (possibly incompletely specified) Boolean functions and finds the
  first cube of a cover of the function.  Returns a generator
  that contains the information necessary to continue the enumeration
  if successful; NULL otherwise.<p>

  The two argument BDDs are the lower and upper bounds of an interval.
  It is a mistake to call this function with a lower bound that is not
  less than or equal to the upper bound.<p>

  A cube is represented as an array of literals, which are integers in
  {0, 1, 2}; 0 represents a complemented literal, 1 represents an
  uncomplemented literal, and 2 stands for don't care. The enumeration
  produces a prime and irredundant cover of the function associated
  with the two BDDs.  The size of the array equals the number of
  variables in the manager at the time Cudd_FirstCube is called.<p>

  This iterator can only be used on BDDs.]

  SideEffects [The first cube is returned as side effect.]

  SeeAlso     [Cudd_ForeachPrime Cudd_NextPrime Cudd_GenFree Cudd_IsGenEmpty
  Cudd_FirstCube Cudd_FirstNode]

******************************************************************************/
func Cudd_FirstPrime(dd *DdManager, l *DdNode, u *DdNode, cube **int32) *DdGen {
    var gen *DdGen;
    var implicant, prime, tmp *DdNode;
    var length, result int32;

    /* Sanity Check. */
    if dd == nil || l == nil || u == nil { return(nil); }

    /* Allocate generator an initialize it. */
    gen = new(DdGen);
    if gen == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }

    gen.manager = dd;
    gen.genType = CUDD_GEN_PRIMES;
    gen.status = CUDD_GEN_EMPTY;
    gen.data.primes.cube = nil;
    gen.data.primes.ub = u;
    gen.stack.sp = 0;
    gen.stack.stack = nil;
    gen.node = l;
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(l).ref);

    gen.data.primes.cube = make([]int32, dd.size);
    if gen.data.primes.cube == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        ABC_FREE(gen);
        return(nil);
    }

    if gen.node == Cudd_ReadLogicZero(dd) {
        gen.status = CUDD_GEN_EMPTY;
    } else {
        implicant = Cudd_LargestCube(dd,gen.node,&length);
        if implicant == nil {
            Cudd_RecursiveDeref(dd,gen.node);
            ABC_FREE(gen.data.primes.cube);
            ABC_FREE(gen);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(implicant).ref);
        prime = Cudd_bddMakePrime(dd,implicant,gen.data.primes.ub);
        if prime == nil {
            Cudd_RecursiveDeref(dd,gen.node);
            Cudd_RecursiveDeref(dd,implicant);
            ABC_FREE(gen.data.primes.cube);
            ABC_FREE(gen);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(prime).ref);
        Cudd_RecursiveDeref(dd,implicant);
        // inlining of macro Cudd_Not
        tmp = Cudd_bddAnd(dd,gen.node,((DdNode *)((ptrint)(prime) ^ 01)));
        if tmp == nil {
            Cudd_RecursiveDeref(dd,gen.node);
            Cudd_RecursiveDeref(dd,prime);
            ABC_FREE(gen.data.primes.cube);
            ABC_FREE(gen);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(tmp).ref);
        Cudd_RecursiveDeref(dd,gen.node);
        gen.node = tmp;
        result = Cudd_BddToCubeArray(dd,prime,gen.data.primes.cube);
        if result == 0 {
            Cudd_RecursiveDeref(dd,gen.node);
            Cudd_RecursiveDeref(dd,prime);
            ABC_FREE(gen.data.primes.cube);
            ABC_FREE(gen);
            return(nil);
        }
        Cudd_RecursiveDeref(dd,prime);
        gen.status = CUDD_GEN_NONEMPTY;
    }
    *cube = gen.data.primes.cube;
    return(gen);

} /* end of Cudd_FirstPrime */


/**Function********************************************************************

  Synopsis    [Generates the next prime of a Boolean function.]

  Description [Generates the next cube of a Boolean function,
  using generator gen. Returns 0 if the enumeration is completed; 1
  otherwise.]

  SideEffects [The cube and is returned as side effects. The
  generator is modified.]

  SeeAlso     [Cudd_ForeachPrime Cudd_FirstPrime Cudd_GenFree Cudd_IsGenEmpty
  Cudd_NextCube Cudd_NextNode]

******************************************************************************/
func Cudd_NextPrime(gen *DdGen, cube **int32) int32 {
    var implicant, prime, tmp *DdNode;
    var dd *DdManager  = gen.manager;
    var length, result int32;

    if gen.node == Cudd_ReadLogicZero(dd) {
        gen.status = CUDD_GEN_EMPTY;
    } else {
        implicant = Cudd_LargestCube(dd,gen.node,&length);
        if implicant == nil {
            gen.status = CUDD_GEN_EMPTY;
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(implicant).ref);
        prime = Cudd_bddMakePrime(dd,implicant,gen.data.primes.ub);
        if prime == nil {
            Cudd_RecursiveDeref(dd,implicant);
            gen.status = CUDD_GEN_EMPTY;
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(prime).ref);
        Cudd_RecursiveDeref(dd,implicant);
        // inlining of macro Cudd_Not
        tmp = Cudd_bddAnd(dd,gen.node,((DdNode *)((ptrint)(prime) ^ 01)));
        if tmp == nil {
            Cudd_RecursiveDeref(dd,prime);
            gen.status = CUDD_GEN_EMPTY;
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(tmp).ref);
        Cudd_RecursiveDeref(dd,gen.node);
        gen.node = tmp;
        result = Cudd_BddToCubeArray(dd,prime,gen.data.primes.cube);
        if result == 0 {
            Cudd_RecursiveDeref(dd,prime);
            gen.status = CUDD_GEN_EMPTY;
            return(0);
        }
        Cudd_RecursiveDeref(dd,prime);
        gen.status = CUDD_GEN_NONEMPTY;
    }
    if gen.status == CUDD_GEN_EMPTY { return(0); }
    *cube = gen.data.primes.cube;
    return(1);

} /* end of Cudd_NextPrime */


/**Function********************************************************************

  Synopsis    [Computes the cube of an array of BDD variables.]

  Description [Computes the cube of an array of BDD variables. If
  non-null, the phase argument indicates which literal of each
  variable should appear in the cube. If phase\[i\] is nonzero, then the
  positive literal is used. If phase is NULL, the cube is positive unate.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addComputeCube Cudd_IndicesToCube Cudd_CubeArrayToBdd]

******************************************************************************/
func Cudd_bddComputeCube(dd *DdManager, vars **DdNode, phase *int32, n int32) *DdNode {
    var cube *DdNode;
    var fn *DdNode;
    var i int32;

    // inlining of macro DD_ONE
    cube = ((dd).one);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(cube).ref);

    for i = n - 1; i >= 0; i-- {
        if phase == nil || phase[i] != 0 {
            fn = Cudd_bddAnd(dd,vars[i],cube);
        } else {
            // inlining of macro Cudd_Not
            fn = Cudd_bddAnd(dd,((DdNode *)((ptrint)(vars[i]) ^ 01)),cube);
        }
        if fn == nil {
            Cudd_RecursiveDeref(dd,cube);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(fn).ref);
        Cudd_RecursiveDeref(dd,cube);
        cube = fn;
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(cube).ref);

    return(cube);

}  /* end of Cudd_bddComputeCube */


/**Function********************************************************************

  Synopsis    [Computes the cube of an array of ADD variables.]

  Description [Computes the cube of an array of ADD variables.  If
  non-null, the phase argument indicates which literal of each
  variable should appear in the cube. If phase\[i\] is nonzero, then the
  positive literal is used. If phase is NULL, the cube is positive unate.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [none]

  SeeAlso     [Cudd_bddComputeCube]

******************************************************************************/
func Cudd_addComputeCube(dd *DdManager, vars **DdNode, phase *int32, n int32) *DdNode {
    var cube, zero *DdNode;
    var fn *DdNode;
    var i int32;

    // inlining of macro DD_ONE
    cube = ((dd).one);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(cube).ref);
    // inlining of macro DD_ZERO
    zero = ((dd).zero);

    for i = n - 1; i >= 0; i-- {
        if phase == nil || phase[i] != 0 {
            fn = Cudd_addIte(dd,vars[i],cube,zero);
        } else {
            fn = Cudd_addIte(dd,vars[i],zero,cube);
        }
        if fn == nil {
            Cudd_RecursiveDeref(dd,cube);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(fn).ref);
        Cudd_RecursiveDeref(dd,cube);
        cube = fn;
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(cube).ref);

    return(cube);

} /* end of Cudd_addComputeCube */


/**Function********************************************************************

  Synopsis    [Builds the BDD of a cube from a positional array.]

  Description [Builds a cube from a positional array.  The array must
  have one integer entry for each BDD variable.  If the i-th entry is
  1, the variable of index i appears in true form in the cube; If the
  i-th entry is 0, the variable of index i appears complemented in the
  cube; otherwise the variable does not appear in the cube.  Returns a
  pointer to the BDD for the cube if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddComputeCube Cudd_IndicesToCube Cudd_BddToCubeArray]

******************************************************************************/
func Cudd_CubeArrayToBdd(dd *DdManager, array *int32) *DdNode {
    var cube, var, tmp *DdNode;
    var i int32;
    var size int32  = Cudd_ReadSize(dd);

    // inlining of macro DD_ONE
    cube = ((dd).one);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(cube).ref);
    for i = size - 1; i >= 0; i-- {
        if (array[i] & ~1) == 0 {
            var = Cudd_bddIthVar(dd,i);
            // inlining of macro Cudd_NotCond
            tmp = Cudd_bddAnd(dd,cube,((DdNode *)((ptrint)(var) ^ (array[i]==0))));
            if tmp == nil {
                Cudd_RecursiveDeref(dd,cube);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(tmp).ref);
            Cudd_RecursiveDeref(dd,cube);
            cube = tmp;
        }
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(cube).ref);
    return(cube);

} /* end of Cudd_CubeArrayToBdd */


/**Function********************************************************************

  Synopsis    [Builds a positional array from the BDD of a cube.]

  Description [Builds a positional array from the BDD of a cube.
  Array must have one entry for each BDD variable.  The positional
  array has 1 in i-th position if the variable of index i appears in
  true form in the cube; it has 0 in i-th position if the variable of
  index i appears in complemented form in the cube; finally, it has 2
  in i-th position if the variable of index i does not appear in the
  cube.  Returns 1 if successful (the BDD is indeed a cube); 0
  otherwise.]

  SideEffects [The result is in the array passed by reference.]

  SeeAlso     [Cudd_CubeArrayToBdd]

******************************************************************************/
func Cudd_BddToCubeArray(dd *DdManager, cube *DdNode, array *int32) int32 {
    var scan, t, e *DdNode;
    var i int32;
    var size int32  = Cudd_ReadSize(dd);
    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    var zero *DdNode  = ((DdNode *)((ptrint)(((dd).one)) ^ 01));

    for i = size-1; i >= 0; i-- {
        array[i] = 2;
    }
    scan = cube;
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    for !((((DdNode *)((ptruint)(scan) & ~01))).index == CUDD_CONST_INDEX) {
        // inlining of macro Cudd_Regular
        var index int32  = ((DdNode *)((ptruint)(scan) & ~01)).index;
        cuddGetBranches(scan,&t,&e);
        if t == zero {
            array[index] = 0;
            scan = e;
        } else if e == zero {
            array[index] = 1;
            scan = t;
        } else {
            return(0);  /* cube is not a cube */
        }
    }
    if scan == zero {
        return(0);
    } else {
        return(1);
    }

} /* end of Cudd_BddToCubeArray */


/**Function********************************************************************

  Synopsis    [Finds the first node of a decision diagram.]

  Description [Defines an iterator on the nodes of a decision diagram
  and finds its first node. Returns a generator that contains the
  information necessary to continue the enumeration if successful;
  NULL otherwise.  The nodes are enumerated in a reverse topological
  order, so that a node is always preceded in the enumeration by its
  descendants.]

  SideEffects [The first node is returned as a side effect.]

  SeeAlso     [Cudd_ForeachNode Cudd_NextNode Cudd_GenFree Cudd_IsGenEmpty
  Cudd_FirstCube]

******************************************************************************/
func Cudd_FirstNode(dd *DdManager, f *DdNode, node **DdNode) *DdGen {
    var gen *DdGen;
    var size int32;

    /* Sanity Check. */
    if dd == nil || f == nil { return(nil); }

    /* Allocate generator an initialize it. */
    gen = new(DdGen);
    if gen == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }

    gen.manager = dd;
    gen.genType = CUDD_GEN_NODES;
    gen.status = CUDD_GEN_EMPTY;
    gen.stack.sp = 0;
    gen.node = nil;

    /* Collect all the nodes on the generator stack for later perusal. */
    // inlining of macro Cudd_Regular
    gen.stack.stack = cuddNodeArray(((DdNode *)((ptruint)(f) & ~01)), &size);
    if gen.stack.stack == nil {
        ABC_FREE(gen);
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    gen.data.nodes.size = size;

    /* Find the first node. */
    if gen.stack.sp < gen.data.nodes.size {
        gen.status = CUDD_GEN_NONEMPTY;
        gen.node = gen.stack.stack[gen.stack.sp];
        *node = gen.node;
    }

    return(gen);

} /* end of Cudd_FirstNode */


/**Function********************************************************************

  Synopsis    [Finds the next node of a decision diagram.]

  Description [Finds the node of a decision diagram, using generator
  gen. Returns 0 if the enumeration is completed; 1 otherwise.]

  SideEffects [The next node is returned as a side effect.]

  SeeAlso     [Cudd_ForeachNode Cudd_FirstNode Cudd_GenFree Cudd_IsGenEmpty
  Cudd_NextCube]

******************************************************************************/
func Cudd_NextNode(gen *DdGen, node **DdNode) int32 {
    /* Find the next node. */
    gen.stack.sp++;
    if gen.stack.sp < gen.data.nodes.size {
        gen.node = gen.stack.stack[gen.stack.sp];
        *node = gen.node;
        return(1);
    } else {
        gen.status = CUDD_GEN_EMPTY;
        return(0);
    }

} /* end of Cudd_NextNode */


/**Function********************************************************************

  Synopsis    [Frees a CUDD generator.]

  Description [Frees a CUDD generator. Always returns 0, so that it can
  be used in mis-like foreach constructs.]

  SideEffects [None]

  SeeAlso     [Cudd_ForeachCube Cudd_ForeachNode Cudd_FirstCube Cudd_NextCube
  Cudd_FirstNode Cudd_NextNode Cudd_IsGenEmpty]

******************************************************************************/
func Cudd_GenFree(gen *DdGen) int32 {
    if gen == nil { return(0); }
    switch gen.genType {
    case CUDD_GEN_CUBES,
     CUDD_GEN_ZDD_PATHS:
        ABC_FREE(gen.data.cubes.cube);
        ABC_FREE(gen.stack.stack);
        break;
    case CUDD_GEN_PRIMES:
        ABC_FREE(gen.data.primes.cube);
        Cudd_RecursiveDeref(gen.manager,gen.node);
        break;
    case CUDD_GEN_NODES:
        ABC_FREE(gen.stack.stack);
        break;
    default:
        return(0);
    }
    ABC_FREE(gen);
    return(0);

} /* end of Cudd_GenFree */


/**Function********************************************************************

  Synopsis    [Queries the status of a generator.]

  Description [Queries the status of a generator. Returns 1 if the
  generator is empty or NULL; 0 otherswise.]

  SideEffects [None]

  SeeAlso     [Cudd_ForeachCube Cudd_ForeachNode Cudd_FirstCube Cudd_NextCube
  Cudd_FirstNode Cudd_NextNode Cudd_GenFree]

******************************************************************************/
func Cudd_IsGenEmpty(gen *DdGen) int32 {
    if gen == nil { return(1); }
    return(gen.status == CUDD_GEN_EMPTY);

} /* end of Cudd_IsGenEmpty */


/**Function********************************************************************

  Synopsis    [Builds a cube of BDD variables from an array of indices.]

  Description [Builds a cube of BDD variables from an array of indices.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddComputeCube Cudd_CubeArrayToBdd]

******************************************************************************/
func Cudd_IndicesToCube(dd *DdManager, array *int32, n int32) *DdNode {
    var cube, tmp *DdNode;
    var i int32;

    // inlining of macro DD_ONE
    cube = ((dd).one);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(cube).ref);
    for i = n - 1; i >= 0; i-- {
        tmp = Cudd_bddAnd(dd,Cudd_bddIthVar(dd,array[i]),cube);
        if tmp == nil {
            Cudd_RecursiveDeref(dd,cube);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(tmp).ref);
        Cudd_RecursiveDeref(dd,cube);
        cube = tmp;
    }

    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(cube).ref);
    return(cube);

} /* end of Cudd_IndicesToCube */


/**Function********************************************************************

  Synopsis    [Prints the package version number.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_PrintVersion(fp *os.File) {
    (void) fmt.Fprintf(fp, "%s\n", CUDD_VERSION);

} /* end of Cudd_PrintVersion */


/**Function********************************************************************

  Synopsis    [Computes the average distance between adjacent nodes.]

  Description [Computes the average distance between adjacent nodes in
  the manager. Adjacent nodes are node pairs such that the second node
  is the then child, else child, or next node in the collision list.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_AverageDistance(dd *DdManager) float64 {
    var tetotal, nexttotal float64;
    var tesubtotal, nextsubtotal float64;
    var temeasured, nextmeasured float64;
    var i, j int32;
    var slots, nvars int32;
    var diff long;
    var scan *DdNode;
    var nodelist *DdNodePtr;
    var sentinel *DdNode  = &(dd.sentinel);

    nvars = dd.size;
    if nvars == 0 { return(0.0); }

    /* Initialize totals. */
    tetotal = 0.0;
    nexttotal = 0.0;
    temeasured = 0.0;
    nextmeasured = 0.0;

    /* Scan the variable subtables. */
    for i = 0; i < nvars; i++ {
        nodelist = dd.subtables[i].nodelist;
        tesubtotal = 0.0;
        nextsubtotal = 0.0;
        slots = dd.subtables[i].slots;
        for j = 0; j < slots; j++ {
            scan = nodelist[j];
            for scan != sentinel {
                // inlining of macro cuddT
                diff = int64( scan) - int64( ((scan).type.kids.T));
                // inlining of macro ddAbs
                tesubtotal += float64( (((diff)<0) ? -(diff) : (diff)));
                // inlining of macro Cudd_Regular
                // inlining of macro cuddE
                diff = int64( scan) - int64( ((DdNode *)((ptruint)(((scan).type.kids.E)) & ~01)));
                // inlining of macro ddAbs
                tesubtotal += float64( (((diff)<0) ? -(diff) : (diff)));
                temeasured += 2.0;
                if scan.next != sentinel {
                    diff = int64( scan) - int64( scan.next);
                    // inlining of macro ddAbs
                    nextsubtotal += float64( (((diff)<0) ? -(diff) : (diff)));
                    nextmeasured += 1.0;
                }
                scan = scan.next;
            }
        }
        tetotal += tesubtotal;
        nexttotal += nextsubtotal;
    }

    /* Scan the constant table. */
    nodelist = dd.constants.nodelist;
    nextsubtotal = 0.0;
    slots = dd.constants.slots;
    for j = 0; j < slots; j++ {
        scan = nodelist[j];
        for scan != nil {
            if scan.next != nil {
                diff = int64( scan) - int64( scan.next);
                // inlining of macro ddAbs
                nextsubtotal += float64( (((diff)<0) ? -(diff) : (diff)));
                nextmeasured += 1.0;
            }
            scan = scan.next;
        }
    }
    nexttotal += nextsubtotal;

    return((tetotal + nexttotal) / (temeasured + nextmeasured));

} /* end of Cudd_AverageDistance */


/**Function********************************************************************

  Synopsis    [Portable random number generator.]

  Description [Portable number generator based on ran2 from "Numerical
  Recipes in C." It is a long period (> 2 * 10^18) random number generator
  of L'Ecuyer with Bays-Durham shuffle. Returns a long integer uniformly
  distributed between 0 and 2147483561 (inclusive of the endpoint values).
  The random generator can be explicitly initialized by calling
  Cudd_Srandom. If no explicit initialization is performed, then the
  seed 1 is assumed.]

  SideEffects [None]

  SeeAlso     [Cudd_Srandom]

******************************************************************************/
func Cudd_Random() long {
    var i int32;      /* index in the shuffle table */
    long int w; /* work variable */

    /* cuddRand == 0 if the geneartor has not been initialized yet. */
    if cuddRand == 0 { Cudd_Srandom(1); }

    /* Compute cuddRand = (cuddRand * LEQA1) % MODULUS1 avoiding
    ** overflows by Schrage's method.
    */
    w          = cuddRand / LEQQ1;
    cuddRand   = LEQA1 * (cuddRand - w * LEQQ1) - w * LEQR1;
    cuddRand  += (cuddRand < 0) * MODULUS1;

    /* Compute cuddRand2 = (cuddRand2 * LEQA2) % MODULUS2 avoiding
    ** overflows by Schrage's method.
    */
    w          = cuddRand2 / LEQQ2;
    cuddRand2  = LEQA2 * (cuddRand2 - w * LEQQ2) - w * LEQR2;
    cuddRand2 += (cuddRand2 < 0) * MODULUS2;

    /* cuddRand is shuffled with the Bays-Durham algorithm.
    ** shuffleSelect and cuddRand2 are combined to generate the output.
    */

    /* Pick one element from the shuffle table; "i" will be in the range
    ** from 0 to STAB_SIZE-1.
    */
    i = int32( (shuffleSelect / STAB_DIV));
    /* Mix the element of the shuffle table with the current iterate of
    ** the second sub-generator, and replace the chosen element of the
    ** shuffle table with the current iterate of the first sub-generator.
    */
    shuffleSelect   = shuffleTable[i] - cuddRand2;
    shuffleTable[i] = cuddRand;
    shuffleSelect  += (shuffleSelect < 1) * (MODULUS1 - 1);
    /* Since shuffleSelect != 0, and we want to be able to return 0,
    ** here we subtract 1 before returning.
    */
    return(shuffleSelect - 1);

} /* end of Cudd_Random */


/**Function********************************************************************

  Synopsis    [Initializer for the portable random number generator.]

  Description [Initializer for the portable number generator based on
  ran2 in "Numerical Recipes in C." The input is the seed for the
  generator. If it is negative, its absolute value is taken as seed.
  If it is 0, then 1 is taken as seed. The initialized sets up the two
  recurrences used to generate a long-period stream, and sets up the
  shuffle table.]

  SideEffects [None]

  SeeAlso     [Cudd_Random]

******************************************************************************/
func Cudd_Srandom(seed long) {
    var i int32;

    if seed < 0 {       cuddRand = -seed;
    } else if seed == 0 { cuddRand = 1;
    } else {                cuddRand = seed; }
    cuddRand2 = cuddRand;
    /* Load the shuffle table (after 11 warm-ups). */
    for i = 0; i < STAB_SIZE + 11; i++ {
        long int w;
        w = cuddRand / LEQQ1;
        cuddRand = LEQA1 * (cuddRand - w * LEQQ1) - w * LEQR1;
        cuddRand += (cuddRand < 0) * MODULUS1;
        shuffleTable[i % STAB_SIZE] = cuddRand;
    }
    shuffleSelect = shuffleTable[1 % STAB_SIZE];

} /* end of Cudd_Srandom */


/**Function********************************************************************

  Synopsis    [Computes the density of a BDD or ADD.]

  Description [Computes the density of a BDD or ADD. The density is
  the ratio of the number of minterms to the number of nodes. If 0 is
  passed as number of variables, the number of variables existing in
  the manager is used. Returns the density if successful; (double)
  CUDD_OUT_OF_MEM otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_CountMinterm Cudd_DagSize]

******************************************************************************/
double
Cudd_Density(
  DdManager * dd /* manager */,
  DdNode * f /* function whose density is sought */,
  int  nvars /* size of the support of f */)
{
    var minterms float64;
    var nodes int32;
    var density float64;

    if nvars == 0 { nvars = dd.size; }
    minterms = Cudd_CountMinterm(dd,f,nvars);
    if minterms == float64( CUDD_OUT_OF_MEM) { return(minterms); }
    nodes = Cudd_DagSize(f);
    density = minterms / float64( nodes);
    return(density);

} /* end of Cudd_Density */


/**Function********************************************************************

  Synopsis    [Warns that a memory allocation failed.]

  Description [Warns that a memory allocation failed.
  This function can be used as replacement of MMout_of_memory to prevent
  the safe_mem functions of the util package from exiting when malloc
  returns NULL. One possible use is in case of discretionary allocations;
  for instance, the allocation of memory to enlarge the computed table.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
void
Cudd_OutOfMem(
  long size /* size of the allocation that failed */)
{
    (void) fflush(os.Stdout);
    (void) fmt.Fprintf(os.Stderr, "\nunable to allocate %d bytes\n", size);
    return;

} /* end of Cudd_OutOfMem */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Prints a DD to the standard output. One line per node is
  printed.]

  Description [Prints a DD to the standard output. One line per node is
  printed. Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_PrintDebug]

******************************************************************************/
func cuddP(dd *DdManager, f *DdNode) int32 {
    var retval int32;
    var table *st__table  = st__init_table( st__ptrcmp, st__ptrhash);

    if table == nil { return(0); }

    retval = dp2(dd,f,table);
    st__free_table(table);
    (void) fputc('\n',dd.out);
    return(retval);

} /* end of cuddP */


/**Function********************************************************************

  Synopsis [Frees the memory used to store the minterm counts recorded
  in the visited table.]

  Description [Frees the memory used to store the minterm counts
  recorded in the visited table. Returns st__CONTINUE.]

  SideEffects [None]

******************************************************************************/
     func cuddStCountfree(key string, value string, arg string) st__retval {
    var d *float64;

    d = (double *)value;
    ABC_FREE(d);
    return( st__CONTINUE);

} /* end of cuddStCountfree */


/**Function********************************************************************

  Synopsis    [Recursively collects all the nodes of a DD in a symbol
  table.]

  Description [Traverses the DD f and collects all its nodes in a
  symbol table.  f is assumed to be a regular pointer and
  cuddCollectNodes guarantees this assumption in the recursive calls.
  Returns 1 in case of success; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddCollectNodes(f *DdNode, visited *st__table) int32 {
    var T, E *DdNode;
    var retval int32;

#ifdef DD_DEBUG
    assert(!Cudd_IsComplement(f));
#endif

    /* If already visited, nothing to do. */
    // inlining of macro st__is_member
    if  st__lookup(visited,(char *) f,(char **) 0) == 1 {
        return(1); }

    /* Check for abnormal condition that should never happen. */
    if f == nil {
        return(0); }

    /* Mark node as visited. */
    if st__add_direct(visited, (char *) f, nil) == st__OUT_OF_MEM {
        return(0); }

    /* Check terminal case. */
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        return(1); }

    /* Recursive calls. */
    // inlining of macro cuddT
    T = ((f).type.kids.T);
    retval = cuddCollectNodes(T,visited);
    if retval != 1 { return(retval); }
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    E = ((DdNode *)((ptruint)(((f).type.kids.E)) & ~01));
    retval = cuddCollectNodes(E,visited);
    return(retval);

} /* end of cuddCollectNodes */


/**Function********************************************************************

  Synopsis    [Recursively collects all the nodes of a DD in an array.]

  Description [Traverses the DD f and collects all its nodes in an array.
  The caller should free the array returned by cuddNodeArray.
  Returns a pointer to the array of nodes in case of success; NULL
  otherwise.  The nodes are collected in reverse topological order, so
  that a node is always preceded in the array by all its descendants.]

  SideEffects [The number of nodes is returned as a side effect.]

  SeeAlso     [Cudd_FirstNode]

******************************************************************************/
func cuddNodeArray(f *DdNode, n *int32) *DdNodePtr {
    var table *DdNodePtr;
    var size, retval int32;

    // inlining of macro Cudd_Regular
    size = ddDagInt(((DdNode *)((ptruint)(f) & ~01)));
    table = ABC_ALLOC(DdNodePtr, size);
    if table == nil {
        // inlining of macro Cudd_Regular
        ddClearFlag(((DdNode *)((ptruint)(f) & ~01)));
        return(nil);
    }

    retval = cuddNodeArrayRecur(f, table, 0);
    assert(retval == size, "retval == size");

    *n = size;
    return(table);

} /* cuddNodeArray */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of cuddP.]

  Description [Performs the recursive step of cuddP. Returns 1 in case
  of success; 0 otherwise.]

  SideEffects [None]

******************************************************************************/
func dp2(dd *DdManager, f *DdNode, t *st__table) int32 {
    var g, n, N *DdNode;
    var T, E int32;

    if f == nil {
        return(0);
    }
    // inlining of macro Cudd_Regular
    g = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro cuddIsConstant
    if (g).index == CUDD_CONST_INDEX {
#if SIZEOF_VOID_P == 8
        (void) fprintf(dd->out,"ID = %c0x%lx\tvalue = %-9g\n", bang(f),
                (ptruint) g / (ptruint) sizeof(DdNode),cuddV(g));
#else
        (void) fprintf(dd->out,"ID = %c0x%x\tvalue = %-9g\n", bang(f),
                (ptruint) g / (ptruint) sizeof(DdNode),cuddV(g));
#endif
        return(1);
    }
    // inlining of macro st__is_member
    if  st__lookup(t,(char *) g,(char **) 0) == 1 {
        return(1);
    }
    if st__add_direct(t,(char *) g,nil) == st__OUT_OF_MEM {
        return(0); }
#ifdef DD_STATS
#if SIZEOF_VOID_P == 8
    (void) fprintf(dd->out,"ID = %c0x%lx\tindex = %d\tr = %d\t", bang(f),
                (ptruint) g / (ptruint) sizeof(DdNode), g->index, g->ref);
#else
    (void) fprintf(dd->out,"ID = %c0x%x\tindex = %d\tr = %d\t", bang(f),
                (ptruint) g / (ptruint) sizeof(DdNode),g->index,g->ref);
#endif
#else
#if SIZEOF_VOID_P == 8
    (void) fprintf(dd->out,"ID = %c0x%lx\tindex = %u\t", bang(f),
                (ptruint) g / (ptruint) sizeof(DdNode),g->index);
#else
    (void) fprintf(dd->out,"ID = %c0x%x\tindex = %hu\t", bang(f),
                (ptruint) g / (ptruint) sizeof(DdNode),g->index);
#endif
#endif
    // inlining of macro cuddT
    n = ((g).type.kids.T);
    // inlining of macro cuddIsConstant
    if (n).index == CUDD_CONST_INDEX {
        // inlining of macro cuddV
        (void) fmt.Fprintf(dd.out, "T = %-9g\t", ((n).type.value));
        T = 1;
    } else {
#if SIZEOF_VOID_P == 8
        (void) fprintf(dd->out,"T = 0x%lx\t",(ptruint) n / (ptruint) sizeof(DdNode));
#else
        (void) fprintf(dd->out,"T = 0x%x\t",(ptruint) n / (ptruint) sizeof(DdNode));
#endif
        T = 0;
    }

    // inlining of macro cuddE
    n = ((g).type.kids.E);
    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(n) & ~01));
    // inlining of macro cuddIsConstant
    if (N).index == CUDD_CONST_INDEX {
        // inlining of macro bang
        // inlining of macro Cudd_IsComplement
        // inlining of macro cuddV
        (void) fmt.Fprintf(dd.out, "E = %c%-9g\n", (((int32( ((ptrint) (n) & 01)))) ? '!' : ' '), ((N).type.value));
        E = 1;
    } else {
#if SIZEOF_VOID_P == 8
        (void) fprintf(dd->out,"E = %c0x%lx\n", bang(n), (ptruint) N/(ptruint) sizeof(DdNode));
#else
        (void) fprintf(dd->out,"E = %c0x%x\n", bang(n), (ptruint) N/(ptruint) sizeof(DdNode));
#endif
        E = 0;
    }
    if E == 0 {
        if dp2(dd,N,t) == 0 {
            return(0); }
    }
    if T == 0 {
        // inlining of macro cuddT
        if dp2(dd,((g).type.kids.T),t) == 0 {
            return(0); }
    }
    return(1);

} /* end of dp2 */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_PrintMinterm.]

  Description []

  SideEffects [None]

******************************************************************************/
static void
ddPrintMintermAux(
  DdManager * dd /* manager */,
  DdNode * node /* current node */,
  int * list /* current recursion path */)
{
    var N, Nv, Nnv *DdNode;
    var i, v, index int32;

    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));

    // inlining of macro cuddIsConstant
    if (N).index == CUDD_CONST_INDEX {
        /* Terminal case: Print one cube based on the current recursion
        ** path, unless we have reached the background value (ADDs) or
        ** the logical zero (BDDs).
        */
        if node != background && node != zero {
            for i = 0; i < dd.size; i++ {
                v = list[i];
                if v == 0 { (void) fmt.Fprintf(dd.out, "0");
                } else if v == 1 { (void) fmt.Fprintf(dd.out, "1");
                } else { (void) fmt.Fprintf(dd.out, "-"); }
            }
            // inlining of macro cuddV
            (void) fmt.Fprintf(dd.out, " % g\n", ((node).type.value));
        }
    } else {
        // inlining of macro cuddT
        Nv  = ((N).type.kids.T);
        // inlining of macro cuddE
        Nnv = ((N).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (node) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Nv  = ((DdNode *)((ptrint)(Nv) ^ 01));
            // inlining of macro Cudd_Not
            Nnv = ((DdNode *)((ptrint)(Nnv) ^ 01));
        }
        index = N.index;
        list[index] = 0;
        ddPrintMintermAux(dd,Nnv,list);
        list[index] = 1;
        ddPrintMintermAux(dd,Nv,list);
        list[index] = 2;
    }
    return;

} /* end of ddPrintMintermAux */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_DagSize.]

  Description [Performs the recursive step of Cudd_DagSize. Returns the
  number of nodes in the graph rooted at n.]

  SideEffects [None]

******************************************************************************/
func ddDagInt(n *DdNode) int32 {
    var tval, eval int32;

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (n.next) & 01)) != 0 {
        return(0);
    }
    // inlining of macro Cudd_Not
    n.next = ((DdNode *)((ptrint)(n.next) ^ 01));
    // inlining of macro cuddIsConstant
    if (n).index == CUDD_CONST_INDEX {
        return(1);
    }
    // inlining of macro cuddT
    tval = ddDagInt(((n).type.kids.T));
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    eval = ddDagInt(((DdNode *)((ptruint)(((n).type.kids.E)) & ~01)));
    return(1 + tval + eval);

} /* end of ddDagInt */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of cuddNodeArray.]

  Description [Performs the recursive step of cuddNodeArray.  Returns
  an the number of nodes in the DD.  Clear the least significant bit
  of the next field that was used as visited flag by
  cuddNodeArrayRecur when counting the nodes.  node is supposed to be
  regular; the invariant is maintained by this procedure.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddNodeArrayRecur(f *DdNode, table *DdNodePtr, index int32) int32 {
    var tindex, eindex int32;

    // inlining of macro Cudd_IsComplement
    if !((int32( ((ptrint) (f.next) & 01)) != 0)) {
        return(index);
    }
    /* Clear visited flag. */
    // inlining of macro Cudd_Regular
    f.next = ((DdNode *)((ptruint)(f.next) & ~01));
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        table[index] = f;
        return(index + 1);
    }
    // inlining of macro cuddT
    tindex = cuddNodeArrayRecur(((f).type.kids.T), table, index);
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    eindex = cuddNodeArrayRecur(((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)), table, tindex);
    table[eindex] = f;
    return(eindex + 1);

} /* end of cuddNodeArrayRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_CofactorEstimate.]

  Description [Performs the recursive step of Cudd_CofactorEstimate.
  Returns an estimate of the number of nodes in the DD of a
  cofactor of node. Uses the least significant bit of the next field as
  visited flag. node is supposed to be regular; the invariant is maintained
  by this procedure.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddEstimateCofactor(dd *DdManager, table *st__table, node *DdNode, i int32, phase int32, ptr **DdNode) int32 {
    var tval, eval, val int32;
    var ptrT, ptrE *DdNode;

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (node.next) & 01)) != 0 {
        if ! st__lookup(table,(char *)node,(char **)ptr) {
            if st__add_direct(table,(char *)node,(char *)node) ==
                st__OUT_OF_MEM {
                return(CUDD_OUT_OF_MEM); }
            *ptr = node;
        }
        return(0);
    }
    // inlining of macro Cudd_Not
    node.next = ((DdNode *)((ptrint)(node.next) ^ 01));
    // inlining of macro cuddIsConstant
    if (node).index == CUDD_CONST_INDEX {
        *ptr = node;
        if st__add_direct(table,(char *)node,(char *)node) == st__OUT_OF_MEM {
            return(CUDD_OUT_OF_MEM); }
        return(1);
    }
    if int32( node.index) == i {
        if phase == 1 {
            // inlining of macro cuddT
            *ptr = ((node).type.kids.T);
            // inlining of macro cuddT
            val = ddDagInt(((node).type.kids.T));
        } else {
            // inlining of macro cuddE
            *ptr = ((node).type.kids.E);
            // inlining of macro Cudd_Regular
            // inlining of macro cuddE
            val = ddDagInt(((DdNode *)((ptruint)(((node).type.kids.E)) & ~01)));
        }
        if node.ref > 1 {
            if st__add_direct(table,(char *)node,(char *)*ptr) ==
                st__OUT_OF_MEM {
                return(CUDD_OUT_OF_MEM); }
        }
        return(val);
    }
    if dd.perm[node.index] > dd.perm[i] {
        *ptr = node;
        // inlining of macro cuddT
        tval = ddDagInt(((node).type.kids.T));
        // inlining of macro Cudd_Regular
        // inlining of macro cuddE
        eval = ddDagInt(((DdNode *)((ptruint)(((node).type.kids.E)) & ~01)));
        if node.ref > 1 {
            if st__add_direct(table,(char *)node,(char *)node) ==
                st__OUT_OF_MEM {
                return(CUDD_OUT_OF_MEM); }
        }
        val = 1 + tval + eval;
        return(val);
    }
    // inlining of macro cuddT
    tval = cuddEstimateCofactor(dd,table,((node).type.kids.T),i,phase,&ptrT);
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    eval = cuddEstimateCofactor(dd,table,((DdNode *)((ptruint)(((node).type.kids.E)) & ~01)),i,
                                phase,&ptrE);
    // inlining of macro Cudd_NotCond
    // inlining of macro Cudd_IsComplement
    // inlining of macro cuddE
    ptrE = ((DdNode *)((ptrint)(ptrE) ^ ((int32( ((ptrint) (((node).type.kids.E)) & 01))))));
    if ptrT == ptrE {         /* recombination */
        *ptr = ptrT;
        val = tval;
        if node.ref > 1 {
            if st__add_direct(table,(char *)node,(char *)*ptr) ==
                    st__OUT_OF_MEM {
                return(CUDD_OUT_OF_MEM); }
        }
    // inlining of macro cuddT
    // inlining of macro cuddE
    } else if (ptrT != ((node).type.kids.T) || ptrE != ((node).type.kids.E)) &&
               (*ptr = cuddUniqueLookup(dd,node.index,ptrT,ptrE)) != nil {
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) ((*ptr).next) & 01)) != 0 {
            val = 0;
        } else {
            val = 1 + tval + eval;
        }
        if node.ref > 1 {
            if st__add_direct(table,(char *)node,(char *)*ptr) ==
                    st__OUT_OF_MEM {
                return(CUDD_OUT_OF_MEM); }
        }
    } else {
        *ptr = node;
        val = 1 + tval + eval;
    }
    return(val);

} /* end of cuddEstimateCofactor */


/**Function********************************************************************

  Synopsis    [Checks the unique table for the existence of an internal node.]

  Description [Checks the unique table for the existence of an internal
  node. Returns a pointer to the node if it is in the table; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [cuddUniqueInter]

******************************************************************************/
func cuddUniqueLookup(unique *DdManager, index int32, T *DdNode, E *DdNode) *DdNode {
    var posn int32;
    unsigned int level;
    var nodelist *DdNodePtr;
    var looking *DdNode;
    var subtable *DdSubtable;

    if index >= unique.size {
        return(nil);
    }

    level = unique.perm[index];
    subtable = &(unique.subtables[level]);

#ifdef DD_DEBUG
    assert(level < (unsigned) cuddI(unique,T->index));
    assert(level < (unsigned) cuddI(unique,Cudd_Regular(E)->index));
#endif

    posn = ddHash(cuddF2L(T), cuddF2L(E), subtable.shift);
    nodelist = subtable.nodelist;
    looking = nodelist[posn];

    // inlining of macro cuddT
    for T < ((looking).type.kids.T) {
        // inlining of macro Cudd_Regular
        looking = ((DdNode *)((ptruint)(looking.next) & ~01));
    }
    // inlining of macro cuddT
    // inlining of macro cuddE
    for T == ((looking).type.kids.T) && E < ((looking).type.kids.E) {
        // inlining of macro Cudd_Regular
        looking = ((DdNode *)((ptruint)(looking.next) & ~01));
    }
    // inlining of macro cuddT
    // inlining of macro cuddE
    if ((looking).type.kids.T) == T && ((looking).type.kids.E) == E {
        return(looking);
    }

    return(nil);

} /* end of cuddUniqueLookup */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_CofactorEstimateSimple.]

  Description [Performs the recursive step of Cudd_CofactorEstimateSimple.
  Returns an estimate of the number of nodes in the DD of the positive
  cofactor of node. Uses the least significant bit of the next field as
  visited flag. node is supposed to be regular; the invariant is maintained
  by this procedure.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddEstimateCofactorSimple(node *DdNode, i int32) int32 {
    var tval, eval int32;

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (node.next) & 01)) != 0 {
        return(0);
    }
    // inlining of macro Cudd_Not
    node.next = ((DdNode *)((ptrint)(node.next) ^ 01));
    // inlining of macro cuddIsConstant
    if (node).index == CUDD_CONST_INDEX {
        return(1);
    }
    // inlining of macro cuddT
    tval = cuddEstimateCofactorSimple(((node).type.kids.T),i);
    if int32( node.index) == i { return(tval); }
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    eval = cuddEstimateCofactorSimple(((DdNode *)((ptruint)(((node).type.kids.E)) & ~01)),i);
    return(1 + tval + eval);

} /* end of cuddEstimateCofactorSimple */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_CountMinterm.]

  Description [Performs the recursive step of Cudd_CountMinterm.
  It is based on the following identity. Let |f| be the
  number of minterms of f. Then:
  <xmp>
    |f| = (|f0|+|f1|)/2
  </xmp>
  where f0 and f1 are the two cofactors of f.  Does not use the
  identity |f'| = max - |f|, to minimize loss of accuracy due to
  roundoff.  Returns the number of minterms of the function rooted at
  node.]

  SideEffects [None]

******************************************************************************/
func ddCountMintermAux(node *DdNode, max float64, table *DdHashTable) float64 {
    var N, Nt, Ne *DdNode;
    var min, minT, minE float64;
    var res *DdNode;

    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));

    // inlining of macro cuddIsConstant
    if (N).index == CUDD_CONST_INDEX {
        if node == background || node == zero {
            return(0.0);
        } else {
            return(max);
        }
    }
    if N.ref != 1 && (res = cuddHashTableLookup1(table,node)) != nil {
        // inlining of macro cuddV
        min = ((res).type.value);
        if res.ref == 0 {
            table.manager.dead++;
            table.manager.constants.dead++;
        }
        return(min);
    }

    // inlining of macro cuddT
    // inlining of macro cuddE
    Nt = ((N).type.kids.T); Ne = ((N).type.kids.E);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (node) & 01)) != 0 {
        // inlining of macro Cudd_Not
        Nt = ((DdNode *)((ptrint)(Nt) ^ 01)); Ne = ((DdNode *)((ptrint)(Ne) ^ 01));
    }

    minT = ddCountMintermAux(Nt,max,table);
    if minT == float64(CUDD_OUT_OF_MEM) { return(float64(CUDD_OUT_OF_MEM)); }
    minT *= 0.5;
    minE = ddCountMintermAux(Ne,max,table);
    if minE == float64(CUDD_OUT_OF_MEM) { return(float64(CUDD_OUT_OF_MEM)); }
    minE *= 0.5;
    min = minT + minE;

    if N.ref != 1 {
        var fanout ptrint  = (ptrint) N.ref;
        cuddSatDec(fanout);
        res = cuddUniqueConst(table.manager,min);
        if !cuddHashTableInsert1(table,node,res,fanout) {
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res).ref); Cudd_RecursiveDeref(table.manager, res);
            return(float64(CUDD_OUT_OF_MEM));
        }
    }

    return(min);

} /* end of ddCountMintermAux */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_CountPath.]

  Description [Performs the recursive step of Cudd_CountPath.
  It is based on the following identity. Let |f| be the
  number of paths of f. Then:
  <xmp>
    |f| = |f0|+|f1|
  </xmp>
  where f0 and f1 are the two cofactors of f.  Uses the
  identity |f'| = |f|, to improve the utilization of the (local) cache.
  Returns the number of paths of the function rooted at node.]

  SideEffects [None]

******************************************************************************/
func ddCountPathAux(node *DdNode, table *st__table) float64 {

    var Nv, Nnv *DdNode;
    var paths float64;
    var ppaths *float64;
    var paths1 float64;
    var paths2 float64;
    var dummy *float64;


    // inlining of macro cuddIsConstant
    if (node).index == CUDD_CONST_INDEX {
        return(1.0);
    }
    if st__lookup(table, (char *)node, (char **)&dummy) {
        paths = *dummy;
        return(paths);
    }

    // inlining of macro cuddT
    // inlining of macro cuddE
    Nv = ((node).type.kids.T); Nnv = ((node).type.kids.E);

    paths1 = ddCountPathAux(Nv,table);
    if paths1 == float64(CUDD_OUT_OF_MEM) { return(float64(CUDD_OUT_OF_MEM)); }
    // inlining of macro Cudd_Regular
    paths2 = ddCountPathAux(((DdNode *)((ptruint)(Nnv) & ~01)),table);
    if paths2 == float64(CUDD_OUT_OF_MEM) { return(float64(CUDD_OUT_OF_MEM)); }
    paths = paths1 + paths2;

    ppaths = new(float64);
    if ppaths == nil {
        return(float64(CUDD_OUT_OF_MEM));
    }

    *ppaths = paths;

    if st__add_direct(table,(char *)node, (char *)ppaths) == st__OUT_OF_MEM {
        ABC_FREE(ppaths);
        return(float64(CUDD_OUT_OF_MEM));
    }
    return(paths);

} /* end of ddCountPathAux */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_EpdCountMinterm.]

  Description [Performs the recursive step of Cudd_EpdCountMinterm.
  It is based on the following identity. Let |f| be the
  number of minterms of f. Then:
  <xmp>
    |f| = (|f0|+|f1|)/2
  </xmp>
  where f0 and f1 are the two cofactors of f.  Does not use the
  identity |f'| = max - |f|, to minimize loss of accuracy due to
  roundoff.  Returns the number of minterms of the function rooted at
  node.]

  SideEffects [None]

******************************************************************************/
func ddEpdCountMintermAux(node *DdNode, max *EpDouble, epd *EpDouble, table *st__table) int32 {
    var Nt, Ne *DdNode;
    var min *EpDouble;
    var minT EpDouble;
    var minE EpDouble;
    var res *EpDouble;
    var status int32;

    /* node is assumed to be regular */
    // inlining of macro cuddIsConstant
    if (node).index == CUDD_CONST_INDEX {
        if node == background || node == zero {
            EpdMakeZero(epd, 0);
        } else {
            EpdCopy(max, epd);
        }
        return(0);
    }
    if node.ref != 1 && st__lookup(table, (char *)node, (char **)&res) {
        EpdCopy(res, epd);
        return(0);
    }

    // inlining of macro cuddT
    // inlining of macro cuddE
    Nt = ((node).type.kids.T); Ne = ((node).type.kids.E);

    status = ddEpdCountMintermAux(Nt,max,&minT,table);
    if status == CUDD_OUT_OF_MEM { return(CUDD_OUT_OF_MEM); }
    EpdMultiply(&minT, float64(0.5));
    // inlining of macro Cudd_Regular
    status = ddEpdCountMintermAux(((DdNode *)((ptruint)(Ne) & ~01)),max,&minE,table);
    if status == CUDD_OUT_OF_MEM { return(CUDD_OUT_OF_MEM); }
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (Ne) & 01)) != 0 {
        EpdSubtract3(max, &minE, epd);
        EpdCopy(epd, &minE);
    }
    EpdMultiply(&minE, float64(0.5));
    EpdAdd3(&minT, &minE, epd);

    if node.ref > 1 {
        min = EpdAlloc();
        if !(min != nil) {
            return(CUDD_OUT_OF_MEM); }
        EpdCopy(epd, min);
        if st__insert(table, (char *)node, (char *)min) == st__OUT_OF_MEM {
            EpdFree(min);
            return(CUDD_OUT_OF_MEM);
        }
    }

    return(0);

} /* end of ddEpdCountMintermAux */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_CountPathsToNonZero.]

  Description [Performs the recursive step of Cudd_CountPathsToNonZero.
  It is based on the following identity. Let |f| be the
  number of paths of f. Then:
  <xmp>
    |f| = |f0|+|f1|
  </xmp>
  where f0 and f1 are the two cofactors of f.  Returns the number of
  paths of the function rooted at node.]

  SideEffects [None]

******************************************************************************/
func ddCountPathsToNonZero(N *DdNode, table *st__table) float64 {

    var node, Nt, Ne *DdNode;
    var paths float64;
    var ppaths *float64;
    var paths1 float64;
    var paths2 float64;
    var dummy *float64;

    // inlining of macro Cudd_Regular
    node = ((DdNode *)((ptruint)(N) & ~01));
    // inlining of macro cuddIsConstant
    if (node).index == CUDD_CONST_INDEX {
        // inlining of macro Cudd_IsComplement
        // inlining of macro cuddV
        return(float64( !((int32( ((ptrint) (N) & 01))) || ((node).type.value)==DD_ZERO_VAL)));
    }
    if st__lookup(table, (char *)N, (char **)&dummy) {
        paths = *dummy;
        return(paths);
    }

    // inlining of macro cuddT
    // inlining of macro cuddE
    Nt = ((node).type.kids.T); Ne = ((node).type.kids.E);
    if node != N {
        // inlining of macro Cudd_Not
        Nt = ((DdNode *)((ptrint)(Nt) ^ 01)); Ne = ((DdNode *)((ptrint)(Ne) ^ 01));
    }

    paths1 = ddCountPathsToNonZero(Nt,table);
    if paths1 == float64(CUDD_OUT_OF_MEM) { return(float64(CUDD_OUT_OF_MEM)); }
    paths2 = ddCountPathsToNonZero(Ne,table);
    if paths2 == float64(CUDD_OUT_OF_MEM) { return(float64(CUDD_OUT_OF_MEM)); }
    paths = paths1 + paths2;

    ppaths = new(float64);
    if ppaths == nil {
        return(float64(CUDD_OUT_OF_MEM));
    }

    *ppaths = paths;

    if st__add_direct(table,(char *)N, (char *)ppaths) == st__OUT_OF_MEM {
        ABC_FREE(ppaths);
        return(float64(CUDD_OUT_OF_MEM));
    }
    return(paths);

} /* end of ddCountPathsToNonZero */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_Support.]

  Description [Performs the recursive step of Cudd_Support. Performs a
  DFS from f. The support is accumulated in supp as a side effect. Uses
  the LSB of the then pointer as visited flag.]

  SideEffects [None]

  SeeAlso     [ddClearFlag]

******************************************************************************/
func ddSupportStep(f *DdNode, support *int32) {
    // inlining of macro cuddIsConstant
    // inlining of macro Cudd_IsComplement
    if ((f).index == CUDD_CONST_INDEX) || ((int32( ((ptrint) (f.next) & 01)) != 0)) {
        return;
    }

    support[f.index] = 1;
    // inlining of macro cuddT
    ddSupportStep(((f).type.kids.T),support);
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    ddSupportStep(((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)),support);
    /* Mark as visited. */
    // inlining of macro Cudd_Not
    f.next = ((DdNode *)((ptrint)(f.next) ^ 01));
    return;

} /* end of ddSupportStep */


/**Function********************************************************************

  Synopsis    [Performs a DFS from f, clearing the LSB of the next
  pointers.]

  Description []

  SideEffects [None]

  SeeAlso     [ddSupportStep ddDagInt]

******************************************************************************/
func ddClearFlag(f *DdNode) {
    // inlining of macro Cudd_IsComplement
    if !((int32( ((ptrint) (f.next) & 01)) != 0)) {
        return;
    }
    /* Clear visited flag. */
    // inlining of macro Cudd_Regular
    f.next = ((DdNode *)((ptruint)(f.next) & ~01));
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        return;
    }
    // inlining of macro cuddT
    ddClearFlag(((f).type.kids.T));
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    ddClearFlag(((DdNode *)((ptruint)(((f).type.kids.E)) & ~01)));
    return;

} /* end of ddClearFlag */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_CountLeaves.]

  Description [Performs the recursive step of Cudd_CountLeaves. Returns
  the number of leaves in the DD rooted at n.]

  SideEffects [None]

  SeeAlso     [Cudd_CountLeaves]

******************************************************************************/
func ddLeavesInt(n *DdNode) int32 {
    var tval, eval int32;

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (n.next) & 01)) != 0 {
        return(0);
    }
    // inlining of macro Cudd_Not
    n.next = ((DdNode *)((ptrint)(n.next) ^ 01));
    // inlining of macro cuddIsConstant
    if (n).index == CUDD_CONST_INDEX {
        return(1);
    }
    // inlining of macro cuddT
    tval = ddLeavesInt(((n).type.kids.T));
    // inlining of macro Cudd_Regular
    // inlining of macro cuddE
    eval = ddLeavesInt(((DdNode *)((ptruint)(((n).type.kids.E)) & ~01)));
    return(tval + eval);

} /* end of ddLeavesInt */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddPickArbitraryMinterms.]

  Description [Performs the recursive step of Cudd_bddPickArbitraryMinterms.
  Returns 1 if successful; 0 otherwise.]

  SideEffects [none]

  SeeAlso [Cudd_bddPickArbitraryMinterms]

******************************************************************************/
func ddPickArbitraryMinterms(dd *DdManager, node *DdNode, nvars int32, nminterms int32, string *string) int32 {
    var N, T, E *DdNode;
    var one, bzero *DdNode;
    var i, t, result int32;
    var min1, min2 float64;

    if string == nil || node == nil { return(0); }

    /* The constant 0 function has no on-set cubes. */
    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    bzero = ((DdNode *)((ptrint)(one) ^ 01));
    if nminterms == 0 || node == bzero { return(1); }
    if node == one {
        return(1);
    }

    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));
    // inlining of macro cuddT
    // inlining of macro cuddE
    T = ((N).type.kids.T); E = ((N).type.kids.E);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (node) & 01)) != 0 {
        // inlining of macro Cudd_Not
        T = ((DdNode *)((ptrint)(T) ^ 01)); E = ((DdNode *)((ptrint)(E) ^ 01));
    }

    min1 = Cudd_CountMinterm(dd, T, nvars) / 2.0;
    if min1 == float64(CUDD_OUT_OF_MEM) { return(0); }
    min2 = Cudd_CountMinterm(dd, E, nvars) / 2.0;
    if min2 == float64(CUDD_OUT_OF_MEM) { return(0); }

    t = int32((float64(nminterms) * min1 / (min1 + min2) + 0.5));
    for i = 0; i < t; i++ {
        string[i][N.index] = '1';
    }
    for i = t; i < nminterms; i++ {
        string[i][N.index] = '0';
    }

    result = ddPickArbitraryMinterms(dd,T,nvars,t,&string[0]);
    if result == 0 {
        return(0); }
    result = ddPickArbitraryMinterms(dd,E,nvars,nminterms-t,&string[t]);
    return(result);

} /* end of ddPickArbitraryMinterms */


/**Function********************************************************************

  Synopsis    [Finds a representative cube of a BDD.]

  Description [Finds a representative cube of a BDD with the weight of
  each variable. From the top variable, if the weight is greater than or
  equal to 0.0, choose THEN branch unless the child is the constant 0.
  Otherwise, choose ELSE branch unless the child is the constant 0.]

  SideEffects [Cudd_SubsetWithMaskVars Cudd_bddPickOneCube]

******************************************************************************/
func ddPickRepresentativeCube(dd *DdManager, node *DdNode, weight *float64, string string) int32 {
    var N, T, E *DdNode;
    var one, bzero *DdNode;

    if string == nil || node == nil { return(0); }

    /* The constant 0 function has no on-set cubes. */
    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    bzero = ((DdNode *)((ptrint)(one) ^ 01));
    if node == bzero { return(0); }

    // inlining of macro DD_ONE
    if node == ((dd).one) { return(1); }

    for {
        // inlining of macro Cudd_Regular
        N = ((DdNode *)((ptruint)(node) & ~01));
        if N == one {
            break; }
        // inlining of macro cuddT
        T = ((N).type.kids.T);
        // inlining of macro cuddE
        E = ((N).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (node) & 01)) != 0 {
            // inlining of macro Cudd_Not
            T = ((DdNode *)((ptrint)(T) ^ 01));
            // inlining of macro Cudd_Not
            E = ((DdNode *)((ptrint)(E) ^ 01));
        }
        if weight[N.index] >= 0.0 {
            if T == bzero {
                node = E;
                string[N.index] = '0';
            } else {
                node = T;
                string[N.index] = '1';
            }
        } else {
            if E == bzero {
                node = T;
                string[N.index] = '1';
            } else {
                node = E;
                string[N.index] = '0';
            }
        }
    }
    return(1);

} /* end of ddPickRepresentativeCube */


/**Function********************************************************************

  Synopsis [Frees the memory used to store the minterm counts recorded
  in the visited table.]

  Description [Frees the memory used to store the minterm counts
  recorded in the visited table. Returns st__CONTINUE.]

  SideEffects [None]

******************************************************************************/
func ddEpdFree(key string, value string, arg string) st__retval {
    var epd *EpDouble;

    epd = (EpDouble *) value;
    EpdFree(epd);
    return( st__CONTINUE);

} /* end of ddEpdFree */


//ABC_NAMESPACE_IMPL_END


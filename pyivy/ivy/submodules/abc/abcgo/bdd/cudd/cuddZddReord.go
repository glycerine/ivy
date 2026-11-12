package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddZddReord.c]

  PackageName [cudd]

  Synopsis    [Procedures for dynamic variable ordering of ZDDs.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_zddReduceHeap()
                    <li> Cudd_zddShuffleHeap()
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    <li> cuddZddAlignToBdd()
                    <li> cuddZddNextHigh()
                    <li> cuddZddNextLow()
                    <li> cuddZddUniqueCompare()
                    <li> cuddZddSwapInPlace()
                    <li> cuddZddSwapping()
                    <li> cuddZddSifting()
                    </ul>
               Static procedures included in this module:
                    <ul>
                    <li> zddSwapAny()
                    <li> cuddZddSiftingAux()
                    <li> cuddZddSiftingUp()
                    <li> cuddZddSiftingDown()
                    <li> cuddZddSiftingBackward()
                    <li> zddReorderPreprocess()
                    <li> zddReorderPostprocess()
                    <li> zddShuffle()
                    <li> zddSiftUp()
                    </ul>
              ]

  SeeAlso     []

  Author      [Hyong-Kyoon Shin, In-Ho Moon]

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

const DD_MAX_SUBTABLE_SPARSITY = 8
const DD_SHRINK_FACTOR = 2

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
static char rcsid[] DD_UNUSED = "$Id: cuddZddReord.c,v 1.47 2004/08/13 18:04:53 fabio Exp $";
#endif

var zdd_entry *int32;

var zddTotalNumberSwapping int32;

var empty *DdNode;


/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static Move * zddSwapAny (DdManager *table, int x, int y);

// static int cuddZddSiftingAux (DdManager *table, int x, int x_low, int x_high);

// static Move * cuddZddSiftingUp (DdManager *table, int x, int x_low, int initial_size);

// static Move * cuddZddSiftingDown (DdManager *table, int x, int x_high, int initial_size);

// static int cuddZddSiftingBackward (DdManager *table, Move *moves, int size);

// static void zddReorderPreprocess (DdManager *table);

// static int zddReorderPostprocess (DdManager *table);

// static int zddShuffle (DdManager *table, int *permutation);

// static int zddSiftUp (DdManager *table, int x, int xLow);

// static void zddFixTree (DdManager *table, MtrNode *treenode);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Main dynamic reordering routine for ZDDs.]

  Description [Main dynamic reordering routine for ZDDs.
  Calls one of the possible reordering procedures:
  <ul>
  <li>Swapping
  <li>Sifting
  <li>Symmetric Sifting
  </ul>

  For sifting and symmetric sifting it is possible to request reordering
  to convergence.<p>

  The core of all methods is the reordering procedure
  cuddZddSwapInPlace() which swaps two adjacent variables.
  Returns 1 in case of success; 0 otherwise. In the case of symmetric
  sifting (with and without convergence) returns 1 plus the number of
  symmetric variables, in case of success.]

  SideEffects [Changes the variable order for all ZDDs and clears
  the cache.]

******************************************************************************/
int
Cudd_zddReduceHeap(
  DdManager * table /* DD manager */,
  Cudd_ReorderingType heuristic /* method used for reordering */,
  int minsize /* bound below which no reordering occurs */)
{
    var hook *DdHook;
    var result int32;
    unsigned int nextDyn;
#ifdef DD_STATS
    unsigned int initialSize;
    unsigned int finalSize;
#endif
    var localTime long;

    /* Don't reorder if there are too many dead nodes. */
    if table.keysZ - table.deadZ < uint32( minsize) {
        return(1); }

    if heuristic == CUDD_REORDER_SAME {
        heuristic = table.autoMethodZ;
    }
    if heuristic == CUDD_REORDER_NONE {
        return(1);
    }

    /* This call to Cudd_zddReduceHeap does initiate reordering. Therefore
    ** we count it.
    */
    table.reorderings++;
    empty = table.zero;

    localTime = util_cpu_time();

    /* Run the hook functions. */
    hook = table.preReorderingHook;
    for hook != nil {
        var res int32  = (hook.f)(table, "ZDD", (void *)heuristic);
        if res == 0 { return(0); }
        hook = hook.next;
    }

    /* Clear the cache and collect garbage. */
    zddReorderPreprocess(table);
    zddTotalNumberSwapping = 0;

#ifdef DD_STATS
    initialSize = table->keysZ;

    switch(heuristic) {
    case CUDD_REORDER_RANDOM:
    case CUDD_REORDER_RANDOM_PIVOT:
        (void) fprintf(table->out,"#:I_RANDOM  ");
        break;
    case CUDD_REORDER_SIFT:
    case CUDD_REORDER_SIFT_CONVERGE:
    case CUDD_REORDER_SYMM_SIFT:
    case CUDD_REORDER_SYMM_SIFT_CONV:
        (void) fprintf(table->out,"#:I_SIFTING ");
        break;
    case CUDD_REORDER_LINEAR:
    case CUDD_REORDER_LINEAR_CONVERGE:
        (void) fprintf(table->out,"#:I_LINSIFT ");
        break;
    default:
        (void) fprintf(table->err,"Unsupported ZDD reordering method\n");
        return(0);
    }
    (void) fprintf(table->out,"%8d: initial size",initialSize); 
#endif

    result = cuddZddTreeSifting(table,heuristic);

#ifdef DD_STATS
    (void) fprintf(table->out,"\n");
    finalSize = table->keysZ;
    (void) fprintf(table->out,"#:F_REORDER %8d: final size\n",finalSize); 
    (void) fprintf(table->out,"#:T_REORDER %8g: total time (sec)\n",
                   ((double)(util_cpu_time() - localTime)/1000.0)); 
    (void) fprintf(table->out,"#:N_REORDER %8d: total swaps\n",
                   zddTotalNumberSwapping);
#endif

    if result == 0 {
        return(0); }

    if !(zddReorderPostprocess(table) != 0) {
        return(0); }

    if table.realignZ {
        if !cuddBddAlignToZdd(table) {
            return(0); }
    }

    nextDyn = table.keysZ * DD_DYN_RATIO;
    if table.reorderings < 20 || nextDyn > table.nextDyn {
        table.nextDyn = nextDyn;
    } else {
        table.nextDyn += 20; }

    table.reordered = 1;

    /* Run hook functions. */
    hook = table.postReorderingHook;
    for hook != nil {
        var res int32  = (hook.f)(table, "ZDD", (void *)localTime);
        if res == 0 { return(0); }
        hook = hook.next;
    }
    /* Update cumulative reordering time. */
    table.reordTime += util_cpu_time() - localTime;

    return(result);

} /* end of Cudd_zddReduceHeap */


/**Function********************************************************************

  Synopsis    [Reorders ZDD variables according to given permutation.]

  Description [Reorders ZDD variables according to given permutation.
  The i-th entry of the permutation array contains the index of the variable
  that should be brought to the i-th level.  The size of the array should be
  equal or greater to the number of variables currently in use.
  Returns 1 in case of success; 0 otherwise.]

  SideEffects [Changes the ZDD variable order for all diagrams and clears
  the cache.]

  SeeAlso [Cudd_zddReduceHeap]

******************************************************************************/
int
Cudd_zddShuffleHeap(
  DdManager * table /* DD manager */,
  int * permutation /* required variable permutation */)
{

    var result int32;

    empty = table.zero;
    zddReorderPreprocess(table);

    result = zddShuffle(table,permutation);

    if !(zddReorderPostprocess(table) != 0) { return(0); }

    return(result);

} /* end of Cudd_zddShuffleHeap */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Reorders ZDD variables according to the order of the BDD
  variables.]

  Description [Reorders ZDD variables according to the order of the
  BDD variables. This function can be called at the end of BDD
  reordering to insure that the order of the ZDD variables is
  consistent with the order of the BDD variables. The number of ZDD
  variables must be a multiple of the number of BDD variables. Let
  <code>M</code> be the ratio of the two numbers. cuddZddAlignToBdd
  then considers the ZDD variables from <code>M*i</code> to
  <code>(M+1)*i-1</code> as corresponding to BDD variable
  <code>i</code>.  This function should be normally called from
  Cudd_ReduceHeap, which clears the cache.  Returns 1 in case of
  success; 0 otherwise.]

  SideEffects [Changes the ZDD variable order for all diagrams and performs
  garbage collection of the ZDD unique table.]

  SeeAlso [Cudd_zddShuffleHeap Cudd_ReduceHeap]

******************************************************************************/
int
cuddZddAlignToBdd(
  DdManager * table /* DD manager */)
{
    var invpermZ *int32;              /* permutation array */
    var M int32;                      /* ratio of ZDD variables to BDD variables */
    var i, j int32;                    /* loop indices */
    var result int32;                 /* return value */

    /* We assume that a ratio of 0 is OK. */
    if table.sizeZ == 0 {
        return(1); }

    empty = table.zero;
    M = table.sizeZ / table.size;
    /* Check whether the number of ZDD variables is a multiple of the
    ** number of BDD variables.
    */
    if M * table.size != table.sizeZ {
        return(0); }
    /* Create and initialize the inverse permutation array. */
    invpermZ = ABC_ALLOC(int,table.sizeZ);
    if invpermZ == nil {
        table.errorCode = CUDD_MEMORY_OUT;
        return(0);
    }
    for i = 0; i < table.size; i++ {
        var index int32  = table.invperm[i];
        var indexZ int32  = index * M;
        var levelZ int32  = table.permZ[indexZ];
        levelZ = (levelZ / M) * M;
        for j = 0; j < M; j++ {
            invpermZ[M * i + j] = table.invpermZ[levelZ + j];
        }
    }
    /* Eliminate dead nodes. Do not scan the cache again, because we
    ** assume that Cudd_ReduceHeap has already cleared it.
    */
    cuddGarbageCollect(table,0);

    result = zddShuffle(table, invpermZ);
    ABC_FREE(invpermZ);
    /* Fix the ZDD variable group tree. */
    zddFixTree(table,table.treeZ);
    return(result);
    
} /* end of cuddZddAlignToBdd */


/**Function********************************************************************

  Synopsis    [Finds the next subtable with a larger index.]

  Description [Finds the next subtable with a larger index. Returns the
  index.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddNextHigh(table *DdManager, x int32) int32 {
    return(x + 1);

} /* end of cuddZddNextHigh */


/**Function********************************************************************

  Synopsis    [Finds the next subtable with a smaller index.]

  Description [Finds the next subtable with a smaller index. Returns the
  index.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddNextLow(table *DdManager, x int32) int32 {
    return(x - 1);

} /* end of cuddZddNextLow */


/**Function********************************************************************

  Synopsis [Comparison function used by qsort.]

  Description [Comparison function used by qsort to order the
  variables according to the number of keys in the subtables.
  Returns the difference in number of keys between the two
  variables being compared.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddUniqueCompare(ptr_x *int32, ptr_y *int32) int32 {
    return(zdd_entry[*ptr_y] - zdd_entry[*ptr_x]);

} /* end of cuddZddUniqueCompare */


/**Function********************************************************************

  Synopsis    [Swaps two adjacent variables.]

  Description [Swaps two adjacent variables. It assumes that no dead
  nodes are present on entry to this procedure.  The procedure then
  guarantees that no dead nodes will be present when it terminates.
  cuddZddSwapInPlace assumes that x &lt; y.  Returns the number of keys in
  the table if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddSwapInPlace(table *DdManager, x int32, y int32) int32 {
    var xlist, ylist *DdNodePtr;
    var xindex, yindex int32;
    var xslots, yslots int32;
    var xshift, yshift int32;
    var oldxkeys, oldykeys int32;
    var newxkeys, newykeys int32;
    var i int32;
    var posn int32;
    var f, f1, f0, f11, f10, f01, f00 *DdNode;
    var newf1 *DdNode = nil;
    var newf0 *DdNode;
    var next *DdNode;
    var g DdNodePtr;
    var lastP *DdNodePtr;
    var previousP *DdNodePtr;

#ifdef DD_DEBUG
    assert(x < y);
    assert(cuddZddNextHigh(table,x) == y);
    assert(table->subtableZ[x].keys != 0);
    assert(table->subtableZ[y].keys != 0);
    assert(table->subtableZ[x].dead == 0);
    assert(table->subtableZ[y].dead == 0);
#endif

    zddTotalNumberSwapping++;

    /* Get parameters of x subtable. */
    xindex   = table.invpermZ[x];
    xlist    = table.subtableZ[x].nodelist;
    oldxkeys = table.subtableZ[x].keys;
    xslots   = table.subtableZ[x].slots;
    xshift   = table.subtableZ[x].shift;
    newxkeys = 0;

    yindex   = table.invpermZ[y];
    ylist    = table.subtableZ[y].nodelist;
    oldykeys = table.subtableZ[y].keys;
    yslots   = table.subtableZ[y].slots;
    yshift   = table.subtableZ[y].shift;
    newykeys = oldykeys;

    /* The nodes in the x layer that don't depend on y directly
    ** will stay there; the others are put in a chain.
    ** The chain is handled as a FIFO; g points to the beginning and
    ** last points to the end.
    */

    g = nil;
    lastP = &g;
    for i = 0; i < xslots; i++ {
        previousP = &(xlist[i]);
        f = *previousP;
        for f != nil {
            next = f.next;
            // inlining of macro cuddT
            // inlining of macro cuddE
            f1 = ((f).type.kids.T); f0 = ((f).type.kids.E);
            if (f1.index != (DdHalfWord) yindex) &&
                (f0.index != (DdHalfWord) yindex) { /* stays */
                newxkeys++;
                *previousP = f;
                previousP = &(f.next);
            } else {
                f.index = yindex;
                *lastP = f;
                lastP = &(f.next);
            }
            f = next;
        } /* while there are elements in the collision chain */
        *previousP = nil;
    } /* for each slot of the x subtable */
    *lastP = nil;


#ifdef DD_COUNT
    table->swapSteps += oldxkeys - newxkeys;
#endif
    /* Take care of the x nodes that must be re-expressed.
    ** They form a linked list pointed by g. Their index has been
    ** changed to yindex already.
    */
    f = g;
    for f != nil {
        next = f.next;
        /* Find f1, f0, f11, f10, f01, f00. */
        // inlining of macro cuddT
        f1 = ((f).type.kids.T);
        if int32( f1.index) == yindex {
            // inlining of macro cuddT
            // inlining of macro cuddE
            f11 = ((f1).type.kids.T); f10 = ((f1).type.kids.E);
        } else {
            f11 = empty; f10 = f1;
        }
        // inlining of macro cuddE
        f0 = ((f).type.kids.E);
        if int32( f0.index) == yindex {
            // inlining of macro cuddT
            // inlining of macro cuddE
            f01 = ((f0).type.kids.T); f00 = ((f0).type.kids.E);
        } else {
            f01 = empty; f00 = f0;
        }

        /* Decrease ref count of f1. */
        cuddSatDec(f1.ref);
        /* Create the new T child. */
        if f11 == empty {
            if f01 != empty {
                newf1 = f01;
                cuddSatInc(newf1.ref);
            }
            /* else case was already handled when finding nodes
            ** with both children below level y
            */
        } else {
            /* Check xlist for triple (xindex, f11, f01). */
            posn = ddHash(cuddF2L(f11), cuddF2L(f01), xshift);
            /* For each element newf1 in collision list xlist[posn]. */
            newf1 = xlist[posn];
            for newf1 != nil {
                // inlining of macro cuddT
                // inlining of macro cuddE
                if ((newf1).type.kids.T) == f11 && ((newf1).type.kids.E) == f01 {
                    cuddSatInc(newf1.ref);
                    break; /* match */
                }
                newf1 = newf1.next;
            } /* while newf1 */
            if newf1 == nil {        /* no match */
                newf1 = cuddDynamicAllocNode(table);
                if newf1 == nil {
                    goto zddSwapOutOfMem; }
                newf1.index = xindex; newf1.ref = 1;
                // inlining of macro cuddT
                ((newf1).type.kids.T) = f11;
                // inlining of macro cuddE
                ((newf1).type.kids.E) = f01;
                /* Insert newf1 in the collision list xlist[pos];
                ** increase the ref counts of f11 and f01
                */
                newxkeys++;
                newf1.next = xlist[posn];
                xlist[posn] = newf1;
                cuddSatInc(f11.ref);
                cuddSatInc(f01.ref);
            }
        }
        // inlining of macro cuddT
        ((f).type.kids.T) = newf1;

        /* Do the same for f0. */
        /* Decrease ref count of f0. */
        cuddSatDec(f0.ref);
        /* Create the new E child. */
        if f10 == empty {
            newf0 = f00;
            cuddSatInc(newf0.ref);
        } else {
            /* Check xlist for triple (xindex, f10, f00). */
            posn = ddHash(cuddF2L(f10), cuddF2L(f00), xshift);
            /* For each element newf0 in collision list xlist[posn]. */
            newf0 = xlist[posn];
            for newf0 != nil {
                // inlining of macro cuddT
                // inlining of macro cuddE
                if ((newf0).type.kids.T) == f10 && ((newf0).type.kids.E) == f00 {
                    cuddSatInc(newf0.ref);
                    break; /* match */
                }
                newf0 = newf0.next;
            } /* while newf0 */
            if newf0 == nil {        /* no match */
                newf0 = cuddDynamicAllocNode(table);
                if newf0 == nil {
                    goto zddSwapOutOfMem; }
                newf0.index = xindex; newf0.ref = 1;
                // inlining of macro cuddT
                // inlining of macro cuddE
                ((newf0).type.kids.T) = f10; ((newf0).type.kids.E) = f00;
                /* Insert newf0 in the collision list xlist[posn];
                ** increase the ref counts of f10 and f00.
                */
                newxkeys++;
                newf0.next = xlist[posn];
                xlist[posn] = newf0;
                cuddSatInc(f10.ref);
                cuddSatInc(f00.ref);
            }
        }
        // inlining of macro cuddE
        ((f).type.kids.E) = newf0;

        /* Insert the modified f in ylist.
        ** The modified f does not already exists in ylist.
        ** (Because of the uniqueness of the cofactors.)
        */
        posn = ddHash(cuddF2L(newf1), cuddF2L(newf0), yshift);
        newykeys++;
        f.next = ylist[posn];
        ylist[posn] = f;
        f = next;
    } /* while f != NULL */

    /* GC the y layer. */

    /* For each node f in ylist. */
    for i = 0; i < yslots; i++ {
        previousP = &(ylist[i]);
        f = *previousP;
        for f != nil {
            next = f.next;
            if f.ref == 0 {
                cuddSatDec(cuddT(f).ref);
                cuddSatDec(cuddE(f).ref);
                // inlining of macro cuddDeallocNode
                (f).next = (table).nextFree;     (table).nextFree = f;;
                newykeys--;
            } else {
                *previousP = f;
                previousP = &(f.next);
            }
            f = next;
        } /* while f */
        *previousP = nil;
    } /* for i */

    /* Set the appropriate fields in table. */
    table.subtableZ[x].nodelist = ylist;
    table.subtableZ[x].slots    = yslots;
    table.subtableZ[x].shift    = yshift;
    table.subtableZ[x].keys     = newykeys;
    table.subtableZ[x].maxKeys  = yslots * DD_MAX_SUBTABLE_DENSITY;

    table.subtableZ[y].nodelist = xlist;
    table.subtableZ[y].slots    = xslots;
    table.subtableZ[y].shift    = xshift;
    table.subtableZ[y].keys     = newxkeys;
    table.subtableZ[y].maxKeys  = xslots * DD_MAX_SUBTABLE_DENSITY;

    table.permZ[xindex] = y; table.permZ[yindex] = x;
    table.invpermZ[x] = yindex; table.invpermZ[y] = xindex;

    table.keysZ += newxkeys + newykeys - oldxkeys - oldykeys;

    /* Update univ section; univ[x] remains the same. */
    // inlining of macro cuddT
    table.univ[y] = ((table.univ[x]).type.kids.T);

    return (table.keysZ);

zddSwapOutOfMem:
    (void) fmt.Fprintf(table.err, "Error: cuddZddSwapInPlace out of memory\n");

    return (0);

} /* end of cuddZddSwapInPlace */


/**Function********************************************************************

  Synopsis    [Reorders variables by a sequence of (non-adjacent) swaps.]

  Description [Implementation of Plessier's algorithm that reorders
  variables by a sequence of (non-adjacent) swaps.
    <ol>
    <li> Select two variables (RANDOM or HEURISTIC).
    <li> Permute these variables.
    <li> If the nodes have decreased accept the permutation.
    <li> Otherwise reconstruct the original heap.
    <li> Loop.
    </ol>
  Returns 1 in case of success; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddSwapping(table *DdManager, lower int32, upper int32, heuristic Cudd_ReorderingType) int32 {
    var i, j int32;
    var max, keys int32;
    var nvars int32;
    var x, y int32;
    var iterate int32;
    var previousSize int32;
    var moves, move *Move;
    var pivot int32  = -1;
    var modulo int32;
    var result int32;

#ifdef DD_DEBUG
    /* Sanity check */
    assert(lower >= 0 && upper < table->sizeZ && lower <= upper);
#endif

    nvars = upper - lower + 1;
    iterate = nvars;

    for i = 0; i < iterate; i++ {
        if heuristic == CUDD_REORDER_RANDOM_PIVOT {
            /* Find pivot <= id with maximum keys. */
            for max, j = -1, lower; j <= upper; j++ {
                if (keys = table.subtableZ[j].keys) > max {
                    max = keys;
                    pivot = j;
                }
            }

            modulo = upper - pivot;
            if modulo == 0 {
                y = pivot;      /* y = nvars-1 */
            } else {
                /* y = random # from {pivot+1 .. nvars-1} */
                y = pivot + 1 + int32( (Cudd_Random() % modulo));
            }

            modulo = pivot - lower - 1;
            if modulo < 1 {   /* if pivot = 1 or 0 */
                x = lower;
            } else {
                for { /* x = random # from {0 .. pivot-2} */
                    x = int32( Cudd_Random()) % modulo;
                
if !(x == y) { break }
} 
                  /* Is this condition really needed, since x and y
                     are in regions separated by pivot? */
            }
        } else {
            x = int32( (Cudd_Random() % nvars)) + lower;
            for {
                y = int32( (Cudd_Random() % nvars)) + lower;
            
if !(x == y) { break }
} 
        }

        previousSize = table.keysZ;
        moves = zddSwapAny(table, x, y);
        if moves == nil {
            goto cuddZddSwappingOutOfMem; }

        result = cuddZddSiftingBackward(table, moves, previousSize);
        if !(result != 0) {
            goto cuddZddSwappingOutOfMem; }

        for moves != nil {
            move = moves.next;
            // inlining of macro cuddDeallocMove
            ((DdNode *)(moves)).ref = 0;     ((DdNode *)(moves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moves);;
            moves = move;
        }
#ifdef DD_STATS
        if (table->keysZ < (unsigned) previousSize) {
            (void) fprintf(table->out,"-");
        } else if (table->keysZ > (unsigned) previousSize) {
            (void) fprintf(table->out,"+");     /* should never happen */
        } else {
            (void) fprintf(table->out,"=");
        }
        fflush(table->out);
#endif
    }

    return(1);

cuddZddSwappingOutOfMem:
    for moves != nil {
        move = moves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moves)).ref = 0;     ((DdNode *)(moves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moves);;
        moves = move;
    }
    return(0);

} /* end of cuddZddSwapping */


/**Function********************************************************************

  Synopsis    [Implementation of Rudell's sifting algorithm.]

  Description [Implementation of Rudell's sifting algorithm.
  Assumes that no dead nodes are present.
    <ol>
    <li> Order all the variables according to the number of entries
    in each unique table.
    <li> Sift the variable up and down, remembering each time the
    total size of the DD heap.
    <li> Select the best permutation.
    <li> Repeat 3 and 4 for all variables.
    </ol>
  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddSifting(table *DdManager, lower int32, upper int32) int32 {
    var i int32;
    var var *int32;
    var size int32;
    var x int32;
    var result int32;
#ifdef DD_STATS
    int previousSize;
#endif

    size = table.sizeZ;

    /* Find order in which to sift variables. */
    var = nil;
    zdd_entry = ABC_ALLOC(int, size);
    if zdd_entry == nil {
        table.errorCode = CUDD_MEMORY_OUT;
        goto cuddZddSiftingOutOfMem;
    }
    var = ABC_ALLOC(int, size);
    if var == nil {
        table.errorCode = CUDD_MEMORY_OUT;
        goto cuddZddSiftingOutOfMem;
    }

    for i = 0; i < size; i++ {
        x = table.permZ[i];
        zdd_entry[i] = table.subtableZ[x].keys;
        var[i] = i;
    }

    qsort((void *)var, (size_t)size, sizeof(int), (DD_QSFP)cuddZddUniqueCompare);

    /* Now sift. */
    // inlining of macro ddMin
    for i = 0; i < (((size) < (table.siftMaxVar)) ? (size) : (table.siftMaxVar)); i++ {
        if zddTotalNumberSwapping >= table.siftMaxSwap {
            break; }
        x = table.permZ[var[i]];
        if x < lower || x > upper { continue; }
#ifdef DD_STATS
        previousSize = table->keysZ;
#endif
        result = cuddZddSiftingAux(table, x, lower, upper);
        if !(result != 0) {
            goto cuddZddSiftingOutOfMem; }
#ifdef DD_STATS
        if (table->keysZ < (unsigned) previousSize) {
            (void) fprintf(table->out,"-");
        } else if (table->keysZ > (unsigned) previousSize) {
            (void) fprintf(table->out,"+");     /* should never happen */
            (void) fprintf(table->out,"\nSize increased from %d to %d while sifting variable %d\n", previousSize, table->keysZ , var[i]);
        } else {
            (void) fprintf(table->out,"=");
        }
        fflush(table->out);
#endif
    }

    ABC_FREE(var);
    ABC_FREE(zdd_entry);

    return(1);

cuddZddSiftingOutOfMem:

    if zdd_entry != nil { ABC_FREE(zdd_entry); }
    if var != nil { ABC_FREE(var); }

    return(0);

} /* end of cuddZddSifting */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Swaps any two variables.]

  Description [Swaps any two variables. Returns the set of moves.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func zddSwapAny(table *DdManager, x int32, y int32) *Move {
    var move, moves *Move;
    var tmp, size int32;
    var x_ref, y_ref int32;
    var x_next, y_next int32;
    var limit_size int32;

    if x > y {        /* make x precede y */
        tmp = x; x = y; y = tmp;
    }

    x_ref = x; y_ref = y;

    x_next = cuddZddNextHigh(table, x);
    y_next = cuddZddNextLow(table, y);
    moves = nil;
    limit_size = table.keysZ;

    for {
        if x_next == y_next { /* x < x_next = y_next < y */
            size = cuddZddSwapInPlace(table, x, x_next);
            if size == 0 {
                goto zddSwapAnyOutOfMem; }
            move = (Move *) cuddDynamicAllocNode(table);
            if move == nil {
                goto zddSwapAnyOutOfMem; }
            move.x = x;
            move.y = x_next;
            move.size = size;
            move.next = moves;
            moves = move;

            size = cuddZddSwapInPlace(table, y_next, y);
            if size == 0 {
                goto zddSwapAnyOutOfMem; }
            move = (Move *)cuddDynamicAllocNode(table);
            if move == nil {
                goto zddSwapAnyOutOfMem; }
            move.x = y_next;
            move.y = y;
            move.size = size;
            move.next = moves;
            moves = move;

            size = cuddZddSwapInPlace(table, x, x_next);
            if size == 0 {
                goto zddSwapAnyOutOfMem; }
            move = (Move *)cuddDynamicAllocNode(table);
            if move == nil {
                goto zddSwapAnyOutOfMem; }
            move.x = x;
            move.y = x_next;
            move.size = size;
            move.next = moves;
            moves = move;

            tmp = x; x = y; y = tmp;

        } else if x == y_next { /* x = y_next < y = x_next */
            size = cuddZddSwapInPlace(table, x, x_next);
            if size == 0 {
                goto zddSwapAnyOutOfMem; }
            move = (Move *)cuddDynamicAllocNode(table);
            if move == nil {
                goto zddSwapAnyOutOfMem; }
            move.x = x;
            move.y = x_next;
            move.size = size;
            move.next = moves;
            moves = move;

            tmp = x; x = y;  y = tmp;
        } else {
            size = cuddZddSwapInPlace(table, x, x_next);
            if size == 0 {
                goto zddSwapAnyOutOfMem; }
            move = (Move *)cuddDynamicAllocNode(table);
            if move == nil {
                goto zddSwapAnyOutOfMem; }
            move.x = x;
            move.y = x_next;
            move.size = size;
            move.next = moves;
            moves = move;

            size = cuddZddSwapInPlace(table, y_next, y);
            if size == 0 {
                goto zddSwapAnyOutOfMem; }
            move = (Move *)cuddDynamicAllocNode(table);
            if move == nil {
                goto zddSwapAnyOutOfMem; }
            move.x = y_next;
            move.y = y;
            move.size = size;
            move.next = moves;
            moves = move;

            x = x_next; y = y_next;
        }

        x_next = cuddZddNextHigh(table, x);
        y_next = cuddZddNextLow(table, y);
        if x_next > y_ref {
            break; }      /* if x == y_ref */

        if float64( size) > table.maxGrowth * float64( limit_size) {
            break; }
        if size < limit_size {
            limit_size = size; }
    }
    if y_next >= x_ref {
        size = cuddZddSwapInPlace(table, y_next, y);
        if size == 0 {
            goto zddSwapAnyOutOfMem; }
        move = (Move *)cuddDynamicAllocNode(table);
        if move == nil {
            goto zddSwapAnyOutOfMem; }
        move.x = y_next;
        move.y = y;
        move.size = size;
        move.next = moves;
        moves = move;
    }

    return(moves);

zddSwapAnyOutOfMem:
    for moves != nil {
        move = moves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moves)).ref = 0;     ((DdNode *)(moves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moves);;
        moves = move;
    }
    return(nil);

} /* end of zddSwapAny */


/**Function********************************************************************

  Synopsis    [Given xLow <= x <= xHigh moves x up and down between the
  boundaries.]

  Description [Given xLow <= x <= xHigh moves x up and down between the
  boundaries. Finds the best position and does the required changes.
  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddSiftingAux(table *DdManager, x int32, x_low int32, x_high int32) int32 {
    var move *Move;
    var moveUp *Move;        /* list of up move */
    var moveDown *Move;      /* list of down move */

    var initial_size int32;
    var result int32;

    initial_size = table.keysZ;

#ifdef DD_DEBUG
    assert(table->subtableZ[x].keys > 0);
#endif

    moveDown = nil;
    moveUp = nil;

    if x == x_low {
        moveDown = cuddZddSiftingDown(table, x, x_high, initial_size);
        /* after that point x --> x_high */
        if moveDown == nil {
            goto cuddZddSiftingAuxOutOfMem; }
        result = cuddZddSiftingBackward(table, moveDown,
            initial_size);
        /* move backward and stop at best position */
        if !(result != 0) {
            goto cuddZddSiftingAuxOutOfMem; }

    
    } else if x == x_high {
        moveUp = cuddZddSiftingUp(table, x, x_low, initial_size);
        /* after that point x --> x_low */
        if moveUp == nil {
            goto cuddZddSiftingAuxOutOfMem; }
        result = cuddZddSiftingBackward(table, moveUp, initial_size);
        /* move backward and stop at best position */
        if !(result != 0) {
            goto cuddZddSiftingAuxOutOfMem; }
    
    } else if (x - x_low) > (x_high - x) {
        /* must go down first:shorter */
        moveDown = cuddZddSiftingDown(table, x, x_high, initial_size);
        /* after that point x --> x_high */
        if moveDown == nil {
            goto cuddZddSiftingAuxOutOfMem; }
        moveUp = cuddZddSiftingUp(table, moveDown.y, x_low,
            initial_size);
        if moveUp == nil {
            goto cuddZddSiftingAuxOutOfMem; }
        result = cuddZddSiftingBackward(table, moveUp, initial_size);
        /* move backward and stop at best position */
        if !(result != 0) {
            goto cuddZddSiftingAuxOutOfMem; }
    
    } else {
        moveUp = cuddZddSiftingUp(table, x, x_low, initial_size);
        /* after that point x --> x_high */
        if moveUp == nil {
            goto cuddZddSiftingAuxOutOfMem; }
        moveDown = cuddZddSiftingDown(table, moveUp.x, x_high,
            initial_size);
        /* then move up */
        if moveDown == nil {
            goto cuddZddSiftingAuxOutOfMem; }
        result = cuddZddSiftingBackward(table, moveDown,
            initial_size);
        /* move backward and stop at best position */
        if !(result != 0) {
            goto cuddZddSiftingAuxOutOfMem; }
    }

    for moveDown != nil {
        move = moveDown.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moveDown)).ref = 0;     ((DdNode *)(moveDown)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveDown);;
        moveDown = move;
    }
    for moveUp != nil {
        move = moveUp.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moveUp)).ref = 0;     ((DdNode *)(moveUp)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveUp);;
        moveUp = move;
    }

    return(1);

cuddZddSiftingAuxOutOfMem:
    for moveDown != nil {
        move = moveDown.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moveDown)).ref = 0;     ((DdNode *)(moveDown)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveDown);;
        moveDown = move;
    }
    for moveUp != nil {
        move = moveUp.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moveUp)).ref = 0;     ((DdNode *)(moveUp)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveUp);;
        moveUp = move;
    }

    return(0);

} /* end of cuddZddSiftingAux */


/**Function********************************************************************

  Synopsis    [Sifts a variable up.]

  Description [Sifts a variable up. Moves y up until either it reaches
  the bound (x_low) or the size of the ZDD heap increases too much.
  Returns the set of moves in case of success; NULL if memory is full.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddSiftingUp(table *DdManager, x int32, x_low int32, initial_size int32) *Move {
    var moves *Move;
    var move *Move;
    var y int32;
    var size int32;
    var limit_size int32  = initial_size;

    moves = nil;
    y = cuddZddNextLow(table, x);
    for y >= x_low {
        size = cuddZddSwapInPlace(table, y, x);
        if size == 0 {
            goto cuddZddSiftingUpOutOfMem; }
        move = (Move *)cuddDynamicAllocNode(table);
        if move == nil {
            goto cuddZddSiftingUpOutOfMem; }
        move.x = y;
        move.y = x;
        move.size = size;
        move.next = moves;
        moves = move;

        if float64(size) > float64(limit_size) * table.maxGrowth {
            break; }
        if size < limit_size {
            limit_size = size; }

        x = y;
        y = cuddZddNextLow(table, x);
    }
    return(moves);

cuddZddSiftingUpOutOfMem:
    for moves != nil {
        move = moves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moves)).ref = 0;     ((DdNode *)(moves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moves);;
        moves = move;
    }
    return(nil);

} /* end of cuddZddSiftingUp */


/**Function********************************************************************

  Synopsis    [Sifts a variable down.]

  Description [Sifts a variable down. Moves x down until either it
  reaches the bound (x_high) or the size of the ZDD heap increases too
  much. Returns the set of moves in case of success; NULL if memory is
  full.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddSiftingDown(table *DdManager, x int32, x_high int32, initial_size int32) *Move {
    var moves *Move;
    var move *Move;
    var y int32;
    var size int32;
    var limit_size int32  = initial_size;

    moves = nil;
    y = cuddZddNextHigh(table, x);
    for y <= x_high {
        size = cuddZddSwapInPlace(table, x, y);
        if size == 0 {
            goto cuddZddSiftingDownOutOfMem; }
        move = (Move *)cuddDynamicAllocNode(table);
        if move == nil {
            goto cuddZddSiftingDownOutOfMem; }
        move.x = x;
        move.y = y;
        move.size = size;
        move.next = moves;
        moves = move;

        if float64(size) > float64(limit_size) * table.maxGrowth {
            break; }
        if size < limit_size {
            limit_size = size; }

        x = y;
        y = cuddZddNextHigh(table, x);
    }
    return(moves);

cuddZddSiftingDownOutOfMem:
    for moves != nil {
        move = moves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moves)).ref = 0;     ((DdNode *)(moves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moves);;
        moves = move;
    }
    return(nil);

} /* end of cuddZddSiftingDown */


/**Function********************************************************************

  Synopsis    [Given a set of moves, returns the ZDD heap to the position
  giving the minimum size.]

  Description [Given a set of moves, returns the ZDD heap to the
  position giving the minimum size. In case of ties, returns to the
  closest position giving the minimum size. Returns 1 in case of
  success; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddSiftingBackward(table *DdManager, moves *Move, size int32) int32 {
    var i int32;
    var i_best int32;
    var move *Move;
    var res int32;

    /* Find the minimum size among moves. */
    i_best = -1;
    for move, i = moves, 0; move != nil; move, i = move.next, i+1 {
        if move.size < size {
            i_best = i;
            size = move.size;
        }
    }

    for move, i = moves, 0; move != nil; move, i = move.next, i+1 {
        if i == i_best {
            break; }
        res = cuddZddSwapInPlace(table, move.x, move.y);
        if !(res != 0) {
            return(0); }
        if i_best == -1 && res == size {
            break; }
    }

    return(1);

} /* end of cuddZddSiftingBackward */


/**Function********************************************************************

  Synopsis    [Prepares the ZDD heap for dynamic reordering.]

  Description [Prepares the ZDD heap for dynamic reordering. Does
  garbage collection, to guarantee that there are no dead nodes;
  and clears the cache, which is invalidated by dynamic reordering.]

  SideEffects [None]

******************************************************************************/
func zddReorderPreprocess(table *DdManager) {

    /* Clear the cache. */
    cuddCacheFlush(table);

    /* Eliminate dead nodes. Do not scan the cache again. */
    cuddGarbageCollect(table,0);

    return;

} /* end of ddReorderPreprocess */


/**Function********************************************************************

  Synopsis    [Shrinks almost empty ZDD subtables at the end of reordering
  to guarantee that they have a reasonable load factor.]

  Description [Shrinks almost empty subtables at the end of reordering to
  guarantee that they have a reasonable load factor. However, if there many
  nodes are being reclaimed, then no resizing occurs. Returns 1 in case of
  success; 0 otherwise.]

  SideEffects [None]

******************************************************************************/
func zddReorderPostprocess(table *DdManager) int32 {
    var i, j, posn int32;
    var nodelist, oldnodelist *DdNodePtr;
    var node, next *DdNode;
    unsigned int slots, oldslots;
    extern DD_OOMFP MMoutOfMemory;
    DD_OOMFP saveHandler;

#ifdef DD_VERBOSE
    (void) fflush(table->out);
#endif

    /* If we have very many reclaimed nodes, we do not want to shrink
    ** the subtables, because this will lead to more garbage
    ** collections. More garbage collections mean shorter mean life for
    ** nodes with zero reference count; hence lower probability of finding
    ** a result in the cache.
    */
    if table.reclaimed > table.allocated / 2 { return(1); }

    /* Resize subtables. */
    for i = 0; i < table.sizeZ; i++ {
        var shift int32;
        oldslots = table.subtableZ[i].slots;
        if oldslots < table.subtableZ[i].keys * DD_MAX_SUBTABLE_SPARSITY ||
            oldslots <= table.initSlots { continue; }
        oldnodelist = table.subtableZ[i].nodelist;
        slots = oldslots >> 1;
        saveHandler = MMoutOfMemory;
        MMoutOfMemory = Cudd_OutOfMem;
        nodelist = ABC_ALLOC(DdNodePtr, slots);
        MMoutOfMemory = saveHandler;
        if nodelist == nil {
            return(1);
        }
        table.subtableZ[i].nodelist = nodelist;
        table.subtableZ[i].slots = slots;
        table.subtableZ[i].shift++;
        table.subtableZ[i].maxKeys = slots * DD_MAX_SUBTABLE_DENSITY;
#ifdef DD_VERBOSE
        (void) fprintf(table->err,
                       "shrunk layer %d (%d keys) from %d to %d slots\n",
                       i, table->subtableZ[i].keys, oldslots, slots);
#endif

        for j = 0; uint32( j) < slots; j++ {
            nodelist[j] = nil;
        }
        shift = table.subtableZ[i].shift;
        for j = 0; uint32( j) < oldslots; j++ {
            node = oldnodelist[j];
            for node != nil {
                next = node.next;
                posn = ddHash(cuddF2L(cuddT(node)), cuddF2L(cuddE(node)), shift);
                node.next = nodelist[posn];
                nodelist[posn] = node;
                node = next;
            }
        }
        ABC_FREE(oldnodelist);

        table.memused += (slots - oldslots) * sizeof(DdNode *);
        table.slots += slots - oldslots;
        table.minDead = uint32( (table.gcFrac * float64( table.slots)));
        // inlining of macro ddMin
        table.cacheSlack = int32( (((DD_MAX_CACHE_TO_SLOTS_RATIO*table.slots) < (table.maxCacheHard)) ? (DD_MAX_CACHE_TO_SLOTS_RATIO*table.slots) : (table.maxCacheHard))) -
            2 * int32( table.cacheSlots);
    }
    /* We don't look at the constant subtable, because it is not
    ** affected by reordering.
    */

    return(1);

} /* end of zddReorderPostprocess */


/**Function********************************************************************

  Synopsis    [Reorders ZDD variables according to a given permutation.]

  Description [Reorders ZDD variables according to a given permutation.
  The i-th permutation array contains the index of the variable that
  should be brought to the i-th level. zddShuffle assumes that no
  dead nodes are present.  The reordering is achieved by a series of
  upward sifts.  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso []

******************************************************************************/
func zddShuffle(table *DdManager, permutation *int32) int32 {
    var index int32;
    var level int32;
    var position int32;
    var numvars int32;
    var result int32;
#ifdef DD_STATS
    long        localTime;
    int         initialSize;
    int         finalSize;
    int         previousSize;
#endif

    zddTotalNumberSwapping = 0;
#ifdef DD_STATS
    localTime = util_cpu_time();
    initialSize = table->keysZ;
    (void) fprintf(table->out,"#:I_SHUFFLE %8d: initial size\n",
                   initialSize); 
#endif

    numvars = table.sizeZ;

    for level = 0; level < numvars; level++ {
        index = permutation[level];
        position = table.permZ[index];
#ifdef DD_STATS
        previousSize = table->keysZ;
#endif
        result = zddSiftUp(table,position,level);
        if !(result != 0) { return(0); }
#ifdef DD_STATS
        if (table->keysZ < (unsigned) previousSize) {
            (void) fprintf(table->out,"-");
        } else if (table->keysZ > (unsigned) previousSize) {
            (void) fprintf(table->out,"+");     /* should never happen */
        } else {
            (void) fprintf(table->out,"=");
        }
        fflush(table->out);
#endif
    }

#ifdef DD_STATS
    (void) fprintf(table->out,"\n");
    finalSize = table->keysZ;
    (void) fprintf(table->out,"#:F_SHUFFLE %8d: final size\n",finalSize); 
    (void) fprintf(table->out,"#:T_SHUFFLE %8g: total time (sec)\n",
        ((double)(util_cpu_time() - localTime)/1000.0)); 
    (void) fprintf(table->out,"#:N_SHUFFLE %8d: total swaps\n",
                   zddTotalNumberSwapping);
#endif

    return(1);

} /* end of zddShuffle */


/**Function********************************************************************

  Synopsis    [Moves one ZDD variable up.]

  Description [Takes a ZDD variable from position x and sifts it up to
  position xLow;  xLow should be less than or equal to x.
  Returns 1 if successful; 0 otherwise]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func zddSiftUp(table *DdManager, x int32, xLow int32) int32 {
    var y int32;
    var size int32;

    y = cuddZddNextLow(table,x);
    for y >= xLow {
        size = cuddZddSwapInPlace(table,y,x);
        if size == 0 {
            return(0);
        }
        x = y;
        y = cuddZddNextLow(table,x);
    }
    return(1);

} /* end of zddSiftUp */


/**Function********************************************************************

  Synopsis    [Fixes the ZDD variable group tree after a shuffle.]

  Description [Fixes the ZDD variable group tree after a
  shuffle. Assumes that the order of the variables in a terminal node
  has not been changed.]

  SideEffects [Changes the ZDD variable group tree.]

  SeeAlso     []

******************************************************************************/
func zddFixTree(table *DdManager, treenode *MtrNode) {
    if treenode == nil { return; }
    if int32( treenode.index) < table.sizeZ {
        treenode.low = table.permZ[treenode.index];
    } else {
        treenode.low = treenode.index;
    }
    if treenode.child != nil {
        zddFixTree(table, treenode.child);
    }
    if treenode.younger != nil {
        zddFixTree(table, treenode.younger); }
    if treenode.parent != nil && treenode.low < treenode.parent.low {
        treenode.parent.low = treenode.low;
        treenode.parent.index = treenode.index;
    }
    return;

} /* end of zddFixTree */


//ABC_NAMESPACE_IMPL_END



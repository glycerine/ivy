package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddZddLin.c]

  PackageName [cudd]

  Synopsis    [Procedures for dynamic variable ordering of ZDDs.]

  Description [Internal procedures included in this module:
                    <ul>
                    <li> cuddZddLinearSifting()
                    </ul>
               Static procedures included in this module:
                    <ul>
                    <li> cuddZddLinearInPlace()
                    <li> cuddZddLinerAux()
                    <li> cuddZddLinearUp()
                    <li> cuddZddLinearDown()
                    <li> cuddZddLinearBackward()
                    <li> cuddZddUndoMoves()
                    </ul>
              ]

  SeeAlso     [cuddLinear.c cuddZddReord.c]

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

const CUDD_SWAP_MOVE = 0
const CUDD_LINEAR_TRANSFORM_MOVE = 1
const CUDD_INVERSE_TRANSFORM_MOVE = 2

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
static char rcsid[] DD_UNUSED = "$Id: cuddZddLin.c,v 1.14 2004/08/13 18:04:53 fabio Exp $";
#endif

var zdd_entry *int32;
var zddTotalNumberSwapping int32;
var zddTotalNumberLinearTr int32;
var empty *DdNode;


/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int cuddZddLinearInPlace (DdManager * table, int x, int y);

// static int cuddZddLinearAux (DdManager *table, int x, int xLow, int xHigh);

// static Move * cuddZddLinearUp (DdManager *table, int y, int xLow, Move *prevMoves);

// static Move * cuddZddLinearDown (DdManager *table, int x, int xHigh, Move *prevMoves);

// static int cuddZddLinearBackward (DdManager *table, int size, Move *moves);

// static Move* cuddZddUndoMoves (DdManager *table, Move *moves);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/




/**Function********************************************************************

  Synopsis    [Implementation of the linear sifting algorithm for ZDDs.]

  Description [Implementation of the linear sifting algorithm for ZDDs.
  Assumes that no dead nodes are present.
    <ol>
    <li> Order all the variables according to the number of entries
    in each unique table.
    <li> Sift the variable up and down and applies the XOR transformation,
    remembering each time the total size of the DD heap.
    <li> Select the best permutation.
    <li> Repeat 3 and 4 for all variables.
    </ol>
  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddLinearSifting(table *DdManager, lower int32, upper int32) int32 {
    var i int32;
    var var *int32;
    var size int32;
    var x int32;
    var result int32;
#ifdef DD_STATS
    int previousSize;
#endif

    size = table.sizeZ;
    empty = table.zero;

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
        result = cuddZddLinearAux(table, x, lower, upper);
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

} /* end of cuddZddLinearSifting */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Linearly combines two adjacent variables.]

  Description [Linearly combines two adjacent variables. It assumes
  that no dead nodes are present on entry to this procedure.  The
  procedure then guarantees that no dead nodes will be present when it
  terminates.  cuddZddLinearInPlace assumes that x &lt; y.  Returns the
  number of keys in the table if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [cuddZddSwapInPlace cuddLinearInPlace]

******************************************************************************/
func cuddZddLinearInPlace(table *DdManager, x int32, y int32) int32 {
    var xlist, ylist *DdNodePtr;
    var xindex, yindex int32;
    var xslots, yslots int32;
    var xshift, yshift int32;
    var oldxkeys, oldykeys int32;
    var newxkeys, newykeys int32;
    var i int32;
    var posn int32;
    var f, f1, f0, f11, f10, f01, f00 *DdNode;
    var newf1, newf0, g, next, previous *DdNode;
    var special *DdNode;

#ifdef DD_DEBUG
    assert(x < y);
    assert(cuddZddNextHigh(table,x) == y);
    assert(table->subtableZ[x].keys != 0);
    assert(table->subtableZ[y].keys != 0);
    assert(table->subtableZ[x].dead == 0);
    assert(table->subtableZ[y].dead == 0);
#endif

    zddTotalNumberLinearTr++;

    /* Get parameters of x subtable. */
    xindex   = table.invpermZ[x];
    xlist    = table.subtableZ[x].nodelist;
    oldxkeys = table.subtableZ[x].keys;
    xslots   = table.subtableZ[x].slots;
    xshift   = table.subtableZ[x].shift;
    newxkeys = 0;

    /* Get parameters of y subtable. */
    yindex   = table.invpermZ[y];
    ylist    = table.subtableZ[y].nodelist;
    oldykeys = table.subtableZ[y].keys;
    yslots   = table.subtableZ[y].slots;
    yshift   = table.subtableZ[y].shift;
    newykeys = oldykeys;

    /* The nodes in the x layer are put in two chains.  The chain
    ** pointed by g holds the normal nodes. When re-expressed they stay
    ** in the x list. The chain pointed by special holds the elements
    ** that will move to the y list.
    */
    g = special = nil;
    for i = 0; i < xslots; i++ {
        f = xlist[i];
        if f == nil { continue; }
        xlist[i] = nil;
        for f != nil {
            next = f.next;
            // inlining of macro cuddT
            f1 = ((f).type.kids.T);
            /* if (f1->index == yindex) */ cuddSatDec(f1.ref);
            // inlining of macro cuddE
            f0 = ((f).type.kids.E);
            /* if (f0->index == yindex) */ cuddSatDec(f0.ref);
            // inlining of macro cuddE
            if int32( f1.index) == yindex && ((f1).type.kids.E) == empty &&
                int32( f0.index) != yindex {
                f.next = special;
                special = f;
            } else {
                f.next = g;
                g = f;
            }
            f = next;
        } /* while there are elements in the collision chain */
    } /* for each slot of the x subtable */

    /* Mark y nodes with pointers from above x. We mark them by
    **  changing their index to x.
    */
    for i = 0; i < yslots; i++ {
        f = ylist[i];
        for f != nil {
            if f.ref != 0 {
                f.index = xindex;
            }
            f = f.next;
        } /* while there are elements in the collision chain */
    } /* for each slot of the y subtable */

    /* Move special nodes to the y list. */
    f = special;
    for f != nil {
        next = f.next;
        // inlining of macro cuddT
        f1 = ((f).type.kids.T);
        // inlining of macro cuddT
        f11 = ((f1).type.kids.T);
        // inlining of macro cuddT
        ((f).type.kids.T) = f11;
        cuddSatInc(f11.ref);
        // inlining of macro cuddE
        f0 = ((f).type.kids.E);
        cuddSatInc(f0.ref);
        f.index = yindex;
        /* Insert at the beginning of the list so that it will be
        ** found first if there is a duplicate. The duplicate will
        ** eventually be moved or garbage collected. No node
        ** re-expression will add a pointer to it.
        */
        posn = ddHash(cuddF2L(f11), cuddF2L(f0), yshift);
        f.next = ylist[posn];
        ylist[posn] = f;
        newykeys++;
        f = next;
    }

    /* Take care of the remaining x nodes that must be re-expressed.
    ** They form a linked list pointed by g.
    */
    f = g;
    for f != nil {
#ifdef DD_COUNT
        table->swapSteps++;
#endif
        next = f.next;
        /* Find f1, f0, f11, f10, f01, f00. */
        // inlining of macro cuddT
        f1 = ((f).type.kids.T);
        if int32( f1.index) == yindex || int32( f1.index) == xindex {
            // inlining of macro cuddT
            // inlining of macro cuddE
            f11 = ((f1).type.kids.T); f10 = ((f1).type.kids.E);
        } else {
            f11 = empty; f10 = f1;
        }
        // inlining of macro cuddE
        f0 = ((f).type.kids.E);
        if int32( f0.index) == yindex || int32( f0.index) == xindex {
            // inlining of macro cuddT
            // inlining of macro cuddE
            f01 = ((f0).type.kids.T); f00 = ((f0).type.kids.E);
        } else {
            f01 = empty; f00 = f0;
        }
        /* Create the new T child. */
        if f01 == empty {
            newf1 = f10;
            cuddSatInc(newf1.ref);
        } else {
            /* Check ylist for triple (yindex, f01, f10). */
            posn = ddHash(cuddF2L(f01), cuddF2L(f10), yshift);
            /* For each element newf1 in collision list ylist[posn]. */
            newf1 = ylist[posn];
            /* Search the collision chain skipping the marked nodes. */
            for newf1 != nil {
                // inlining of macro cuddT
                // inlining of macro cuddE
                if ((newf1).type.kids.T) == f01 && ((newf1).type.kids.E) == f10 &&
                    int32( newf1.index) == yindex {
                    cuddSatInc(newf1.ref);
                    break; /* match */
                }
                newf1 = newf1.next;
            } /* while newf1 */
            if newf1 == nil {        /* no match */
                newf1 = cuddDynamicAllocNode(table);
                if newf1 == nil {
                    goto zddSwapOutOfMem; }
                newf1.index = yindex; newf1.ref = 1;
                // inlining of macro cuddT
                ((newf1).type.kids.T) = f01;
                // inlining of macro cuddE
                ((newf1).type.kids.E) = f10;
                /* Insert newf1 in the collision list ylist[pos];
                ** increase the ref counts of f01 and f10
                */
                newykeys++;
                newf1.next = ylist[posn];
                ylist[posn] = newf1;
                cuddSatInc(f01.ref);
                cuddSatInc(f10.ref);
            }
        }
        // inlining of macro cuddT
        ((f).type.kids.T) = newf1;

        /* Do the same for f0. */
        /* Create the new E child. */
        if f11 == empty {
            newf0 = f00;
            cuddSatInc(newf0.ref);
        } else {
            /* Check ylist for triple (yindex, f11, f00). */
            posn = ddHash(cuddF2L(f11), cuddF2L(f00), yshift);
            /* For each element newf0 in collision list ylist[posn]. */
            newf0 = ylist[posn];
            for newf0 != nil {
                // inlining of macro cuddT
                // inlining of macro cuddE
                if ((newf0).type.kids.T) == f11 && ((newf0).type.kids.E) == f00 &&
                    int32( newf0.index) == yindex {
                    cuddSatInc(newf0.ref);
                    break; /* match */
                }
                newf0 = newf0.next;
            } /* while newf0 */
            if newf0 == nil {        /* no match */
                newf0 = cuddDynamicAllocNode(table);
                if newf0 == nil {
                    goto zddSwapOutOfMem; }
                newf0.index = yindex; newf0.ref = 1;
                // inlining of macro cuddT
                // inlining of macro cuddE
                ((newf0).type.kids.T) = f11; ((newf0).type.kids.E) = f00;
                /* Insert newf0 in the collision list ylist[posn];
                ** increase the ref counts of f11 and f00.
                */
                newykeys++;
                newf0.next = ylist[posn];
                ylist[posn] = newf0;
                cuddSatInc(f11.ref);
                cuddSatInc(f00.ref);
            }
        }
        // inlining of macro cuddE
        ((f).type.kids.E) = newf0;

        /* Re-insert the modified f in xlist.
        ** The modified f does not already exists in xlist.
        ** (Because of the uniqueness of the cofactors.)
        */
        posn = ddHash(cuddF2L(newf1), cuddF2L(newf0), xshift);
        newxkeys++;
        f.next = xlist[posn];
        xlist[posn] = f;
        f = next;
    } /* while f != NULL */

    /* GC the y layer and move the marked nodes to the x list. */

    /* For each node f in ylist. */
    for i = 0; i < yslots; i++ {
        previous = nil;
        f = ylist[i];
        for f != nil {
            next = f.next;
            if f.ref == 0 {
                cuddSatDec(cuddT(f).ref);
                cuddSatDec(cuddE(f).ref);
                // inlining of macro cuddDeallocNode
                (f).next = (table).nextFree;     (table).nextFree = f;;
                newykeys--;
                if previous == nil {
                    ylist[i] = next;
                } else {
                    previous.next = next; }
            } else if int32( f.index) == xindex { /* move marked node */
                if previous == nil {
                    ylist[i] = next;
                } else {
                    previous.next = next; }
                // inlining of macro cuddT
                f1 = ((f).type.kids.T);
                cuddSatDec(f1.ref);
                /* Check ylist for triple (yindex, f1, empty). */
                posn = ddHash(cuddF2L(f1), cuddF2L(empty), yshift);
                /* For each element newf1 in collision list ylist[posn]. */
                newf1 = ylist[posn];
                for newf1 != nil {
                    // inlining of macro cuddT
                    // inlining of macro cuddE
                    if ((newf1).type.kids.T) == f1 && ((newf1).type.kids.E) == empty &&
                        int32( newf1.index) == yindex {
                        cuddSatInc(newf1.ref);
                        break; /* match */
                    }
                    newf1 = newf1.next;
                } /* while newf1 */
                if newf1 == nil {    /* no match */
                    newf1 = cuddDynamicAllocNode(table);
                    if newf1 == nil {
                        goto zddSwapOutOfMem; }
                    newf1.index = yindex; newf1.ref = 1;
                    // inlining of macro cuddT
                    // inlining of macro cuddE
                    ((newf1).type.kids.T) = f1; ((newf1).type.kids.E) = empty;
                    /* Insert newf1 in the collision list ylist[posn];
                    ** increase the ref counts of f1 and empty.
                    */
                    newykeys++;
                    newf1.next = ylist[posn];
                    ylist[posn] = newf1;
                    if posn == i && previous == nil {
                        previous = newf1; }
                    cuddSatInc(f1.ref);
                    cuddSatInc(empty.ref);
                }
                // inlining of macro cuddT
                ((f).type.kids.T) = newf1;
                // inlining of macro cuddE
                f0 = ((f).type.kids.E);
                /* Insert f in x list. */
                posn = ddHash(cuddF2L(newf1), cuddF2L(f0), xshift);
                newxkeys++;
                newykeys--;
                f.next = xlist[posn];
                xlist[posn] = f;
            } else {
                previous = f;
            }
            f = next;
        } /* while f */
    } /* for i */

    /* Set the appropriate fields in table. */
    table.subtableZ[x].keys     = newxkeys;
    table.subtableZ[y].keys     = newykeys;

    table.keysZ += newxkeys + newykeys - oldxkeys - oldykeys;

    /* Update univ section; univ[x] remains the same. */
    // inlining of macro cuddT
    table.univ[y] = ((table.univ[x]).type.kids.T);

     
                                                        
                                  
                                 
      

    return (table.keysZ);

zddSwapOutOfMem:
    (void) fmt.Fprintf(table.err, "Error: cuddZddSwapInPlace out of memory\n");

    return (0);

} /* end of cuddZddLinearInPlace */


/**Function********************************************************************

  Synopsis    [Given xLow <= x <= xHigh moves x up and down between the
  boundaries.]

  Description [Given xLow <= x <= xHigh moves x up and down between the
  boundaries. Finds the best position and does the required changes.
  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddLinearAux(table *DdManager, x int32, xLow int32, xHigh int32) int32 {
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

    if x == xLow {
        moveDown = cuddZddLinearDown(table, x, xHigh, nil);
        /* At this point x --> xHigh. */
        if moveDown == (Move *) CUDD_OUT_OF_MEM {
            goto cuddZddLinearAuxOutOfMem; }
        /* Move backward and stop at best position. */
        result = cuddZddLinearBackward(table, initial_size, moveDown);
        if !(result != 0) {
            goto cuddZddLinearAuxOutOfMem; }

    } else if x == xHigh {
        moveUp = cuddZddLinearUp(table, x, xLow, nil);
        /* At this point x --> xLow. */
        if moveUp == (Move *) CUDD_OUT_OF_MEM {
            goto cuddZddLinearAuxOutOfMem; }
        /* Move backward and stop at best position. */
        result = cuddZddLinearBackward(table, initial_size, moveUp);
        if !(result != 0) {
            goto cuddZddLinearAuxOutOfMem; }

    } else if (x - xLow) > (xHigh - x) { /* must go down first: shorter */
        moveDown = cuddZddLinearDown(table, x, xHigh, nil);
        /* At this point x --> xHigh. */
        if moveDown == (Move *) CUDD_OUT_OF_MEM {
            goto cuddZddLinearAuxOutOfMem; }
        moveUp = cuddZddUndoMoves(table,moveDown);
#ifdef DD_DEBUG
        assert(moveUp == NULL || moveUp->x == x);
#endif
        moveUp = cuddZddLinearUp(table, x, xLow, moveUp);
        if moveUp == (Move *) CUDD_OUT_OF_MEM {
            goto cuddZddLinearAuxOutOfMem; }
        /* Move backward and stop at best position. */
        result = cuddZddLinearBackward(table, initial_size, moveUp);
        if !(result != 0) {
            goto cuddZddLinearAuxOutOfMem; }

    } else {
        moveUp = cuddZddLinearUp(table, x, xLow, nil);
        /* At this point x --> xHigh. */
        if moveUp == (Move *) CUDD_OUT_OF_MEM {
            goto cuddZddLinearAuxOutOfMem; }
        /* Then move up. */
        moveDown = cuddZddUndoMoves(table,moveUp);
#ifdef DD_DEBUG
        assert(moveDown == NULL || moveDown->y == x);
#endif
        moveDown = cuddZddLinearDown(table, x, xHigh, moveDown);
        if moveDown == (Move *) CUDD_OUT_OF_MEM {
            goto cuddZddLinearAuxOutOfMem; }
        /* Move backward and stop at best position. */
        result = cuddZddLinearBackward(table, initial_size, moveDown);
        if !(result != 0) {
            goto cuddZddLinearAuxOutOfMem; }
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

cuddZddLinearAuxOutOfMem:
    if moveDown != (Move *) CUDD_OUT_OF_MEM {
        for moveDown != nil {
            move = moveDown.next;
            // inlining of macro cuddDeallocMove
            ((DdNode *)(moveDown)).ref = 0;     ((DdNode *)(moveDown)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveDown);;
            moveDown = move;
        }
    }
    if moveUp != (Move *) CUDD_OUT_OF_MEM {
        for moveUp != nil {
            move = moveUp.next;
            // inlining of macro cuddDeallocMove
            ((DdNode *)(moveUp)).ref = 0;     ((DdNode *)(moveUp)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveUp);;
            moveUp = move;
        }
    }

    return(0);

} /* end of cuddZddLinearAux */


/**Function********************************************************************

  Synopsis    [Sifts a variable up applying the XOR transformation.]

  Description [Sifts a variable up applying the XOR
  transformation. Moves y up until either it reaches the bound (xLow)
  or the size of the ZDD heap increases too much.  Returns the set of
  moves in case of success; NULL if memory is full.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddLinearUp(table *DdManager, y int32, xLow int32, prevMoves *Move) *Move {
    var moves *Move;
    var move *Move;
    var x int32;
    var size, newsize int32;
    var limitSize int32;

    moves = prevMoves;
    limitSize = table.keysZ;

    x = cuddZddNextLow(table, y);
    for x >= xLow {
        size = cuddZddSwapInPlace(table, x, y);
        if size == 0 {
            goto cuddZddLinearUpOutOfMem; }
        newsize = cuddZddLinearInPlace(table, x, y);
        if newsize == 0 {
            goto cuddZddLinearUpOutOfMem; }
        move = (Move *) cuddDynamicAllocNode(table);
        if move == nil {
            goto cuddZddLinearUpOutOfMem; }
        move.x = x;
        move.y = y;
        move.next = moves;
        moves = move;
        move.flags = CUDD_SWAP_MOVE;
        if newsize > size {
            /* Undo transformation. The transformation we apply is
            ** its own inverse. Hence, we just apply the transformation
            ** again.
            */
            newsize = cuddZddLinearInPlace(table,x,y);
            if newsize == 0 { goto cuddZddLinearUpOutOfMem; }
#ifdef DD_DEBUG
            if (newsize != size) {
                (void) fprintf(table->err,"Change in size after identity transformation! From %d to %d\n",size,newsize);
            }
#endif
        } else {
            size = newsize;
            move.flags = CUDD_LINEAR_TRANSFORM_MOVE;
        }
        move.size = size;

        if float64(size) > float64(limitSize) * table.maxGrowth {
            break; }
        if size < limitSize {
            limitSize = size; }

        y = x;
        x = cuddZddNextLow(table, y);
    }
    return(moves);

cuddZddLinearUpOutOfMem:
    for moves != nil {
        move = moves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moves)).ref = 0;     ((DdNode *)(moves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moves);;
        moves = move;
    }
    return((Move *) CUDD_OUT_OF_MEM);

} /* end of cuddZddLinearUp */


/**Function********************************************************************

  Synopsis    [Sifts a variable down and applies the XOR transformation.]

  Description [Sifts a variable down. Moves x down until either it
  reaches the bound (xHigh) or the size of the ZDD heap increases too
  much. Returns the set of moves in case of success; NULL if memory is
  full.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddLinearDown(table *DdManager, x int32, xHigh int32, prevMoves *Move) *Move {
    var moves *Move;
    var move *Move;
    var y int32;
    var size, newsize int32;
    var limitSize int32;

    moves = prevMoves;
    limitSize = table.keysZ;

    y = cuddZddNextHigh(table, x);
    for y <= xHigh {
        size = cuddZddSwapInPlace(table, x, y);
        if size == 0 {
            goto cuddZddLinearDownOutOfMem; }
        newsize = cuddZddLinearInPlace(table, x, y);
        if newsize == 0 {
            goto cuddZddLinearDownOutOfMem; }
        move = (Move *) cuddDynamicAllocNode(table);
        if move == nil {
            goto cuddZddLinearDownOutOfMem; }
        move.x = x;
        move.y = y;
        move.next = moves;
        moves = move;
        move.flags = CUDD_SWAP_MOVE;
        if newsize > size {
            /* Undo transformation. The transformation we apply is
            ** its own inverse. Hence, we just apply the transformation
            ** again.
            */
            newsize = cuddZddLinearInPlace(table,x,y);
            if newsize == 0 { goto cuddZddLinearDownOutOfMem; }
            if newsize != size {
                (void) fmt.Fprintf(table.err, "Change in size after identity transformation! From %d to %d\n", size, newsize);
            }
        } else {
            size = newsize;
            move.flags = CUDD_LINEAR_TRANSFORM_MOVE;
        }
        move.size = size;

        if float64(size) > float64(limitSize) * table.maxGrowth {
            break; }
        if size < limitSize {
            limitSize = size; }

        x = y;
        y = cuddZddNextHigh(table, x);
    }
    return(moves);

cuddZddLinearDownOutOfMem:
    for moves != nil {
        move = moves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moves)).ref = 0;     ((DdNode *)(moves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moves);;
        moves = move;
    }
    return((Move *) CUDD_OUT_OF_MEM);

} /* end of cuddZddLinearDown */


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
func cuddZddLinearBackward(table *DdManager, size int32, moves *Move) int32 {
    var move *Move;
    var res int32;

    /* Find the minimum size among moves. */
    for move = moves; move != nil; move = move.next {
        if move.size < size {
            size = move.size;
        }
    }

    for move = moves; move != nil; move = move.next {
        if move.size == size { return(1); }
        if move.flags == CUDD_LINEAR_TRANSFORM_MOVE {
            res = cuddZddLinearInPlace(table,int32(move.x),int32(move.y));
            if !(res != 0) { return(0); }
        }
        res = cuddZddSwapInPlace(table, move.x, move.y);
        if !(res != 0) {
            return(0); }
        if move.flags == CUDD_INVERSE_TRANSFORM_MOVE {
            res = cuddZddLinearInPlace(table,int32(move.x),int32(move.y));
            if !(res != 0) { return(0); }
        }
    }

    return(1);

} /* end of cuddZddLinearBackward */


/**Function********************************************************************

  Synopsis    [Given a set of moves, returns the ZDD heap to the order
  in effect before the moves.]

  Description [Given a set of moves, returns the ZDD heap to the
  order in effect before the moves.  Returns 1 in case of success;
  0 otherwise.]

  SideEffects [None]

******************************************************************************/
func cuddZddUndoMoves(table *DdManager, moves *Move) *Move {
    var invmoves *Move  = nil;
    var move *Move;
    var invmove *Move;
    var size int32;

    for move = moves; move != nil; move = move.next {
        invmove = (Move *) cuddDynamicAllocNode(table);
        if invmove == nil { goto cuddZddUndoMovesOutOfMem; }
        invmove.x = move.x;
        invmove.y = move.y;
        invmove.next = invmoves;
        invmoves = invmove;
        if move.flags == CUDD_SWAP_MOVE {
            invmove.flags = CUDD_SWAP_MOVE;
            size = cuddZddSwapInPlace(table,int32(move.x),int32(move.y));
            if !(size != 0) { goto cuddZddUndoMovesOutOfMem; }
        } else if move.flags == CUDD_LINEAR_TRANSFORM_MOVE {
            invmove.flags = CUDD_INVERSE_TRANSFORM_MOVE;
            size = cuddZddLinearInPlace(table,int32(move.x),int32(move.y));
            if !(size != 0) { goto cuddZddUndoMovesOutOfMem; }
            size = cuddZddSwapInPlace(table,int32(move.x),int32(move.y));
            if !(size != 0) { goto cuddZddUndoMovesOutOfMem; }
        } else { /* must be CUDD_INVERSE_TRANSFORM_MOVE */
#ifdef DD_DEBUG
            (void) fprintf(table->err,"Unforseen event in ddUndoMoves!\n");
#endif
            invmove.flags = CUDD_LINEAR_TRANSFORM_MOVE;
            size = cuddZddSwapInPlace(table,int32(move.x),int32(move.y));
            if !(size != 0) { goto cuddZddUndoMovesOutOfMem; }
            size = cuddZddLinearInPlace(table,int32(move.x),int32(move.y));
            if !(size != 0) { goto cuddZddUndoMovesOutOfMem; }
        }
        invmove.size = size;
    }

    return(invmoves);

cuddZddUndoMovesOutOfMem:
    for invmoves != nil {
        move = invmoves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(invmoves)).ref = 0;     ((DdNode *)(invmoves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(invmoves);;
        invmoves = move;
    }
    return((Move *) CUDD_OUT_OF_MEM);

} /* end of cuddZddUndoMoves */


//ABC_NAMESPACE_IMPL_END



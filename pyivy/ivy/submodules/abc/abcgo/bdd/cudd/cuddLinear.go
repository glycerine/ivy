package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddLinear.c]

  PackageName [cudd]

  Synopsis    [Functions for DD reduction by linear transformations.]

  Description [ Internal procedures included in this module:
                <ul>
                <li> cuddLinearAndSifting()
                <li> cuddLinearInPlace()
                <li> cuddUpdateInteractionMatrix()
                <li> cuddInitLinear()
                <li> cuddResizeLinear()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> ddLinearUniqueCompare()
                <li> ddLinearAndSiftingAux()
                <li> ddLinearAndSiftingUp()
                <li> ddLinearAndSiftingDown()
                <li> ddLinearAndSiftingBackward()
                <li> ddUndoMoves()
                <li> cuddXorLinear()
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

const CUDD_SWAP_MOVE = 0
const CUDD_LINEAR_TRANSFORM_MOVE = 1
const CUDD_INVERSE_TRANSFORM_MOVE = 2
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

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddLinear.c,v 1.28 2009/02/19 16:21:03 fabio Exp $";
#endif

var entry *int32;

#ifdef DD_STATS
extern  int     ddTotalNumberSwapping;
extern  int     ddTotalNISwaps;
static  int     ddTotalNumberLinearTr;
#endif

#ifdef DD_DEBUG
static  int     zero = 0;
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int ddLinearUniqueCompare (int *ptrX, int *ptrY);

// static int ddLinearAndSiftingAux (DdManager *table, int x, int xLow, int xHigh);

// static Move * ddLinearAndSiftingUp (DdManager *table, int y, int xLow, Move *prevMoves);

// static Move * ddLinearAndSiftingDown (DdManager *table, int x, int xHigh, Move *prevMoves);

// static int ddLinearAndSiftingBackward (DdManager *table, int size, Move *moves);

// static Move* ddUndoMoves (DdManager *table, Move *moves);

// static void cuddXorLinear (DdManager *table, int x, int y);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Prints the linear transform matrix.]

  Description [Prints the linear transform matrix. Returns 1 in case of
  success; 0 otherwise.]

  SideEffects [none]

  SeeAlso     []

******************************************************************************/
func Cudd_PrintLinear(table *DdManager) int32 {
    var i, j, k int32;
    var retval int32;
    var nvars int32  = table.linearSize;
    var wordsPerRow int32  = ((nvars - 1) >> LOGBPL) + 1;
    var word long;

    for i = 0; i < nvars; i++ {
        for j = 0; j < wordsPerRow; j++ {
            word = table.linear[i*wordsPerRow + j];
            for k = 0; k < BPL; k++ {
                retval = fmt.Fprintf(table.out, "%d", word & 1);
                if retval == 0 { return(0); }
                word >>= 1;
            }
        }
        retval = fmt.Fprintf(table.out, "\n");
        if retval == 0 { return(0); }
    }
    return(1);

} /* end of Cudd_PrintLinear */


/**Function********************************************************************

  Synopsis    [Reads an entry of the linear transform matrix.]

  Description [Reads an entry of the linear transform matrix.]

  SideEffects [none]

  SeeAlso     []

******************************************************************************/
int
Cudd_ReadLinear(
  DdManager * table /* CUDD manager */,
  int  x /* row index */,
  int  y /* column index */)
{
    var nvars int32  = table.size;
    var wordsPerRow int32  = ((nvars - 1) >> LOGBPL) + 1;
    var word long;
    var bit int32;
    var result int32;

    assert(table.size == table.linearSize, "table.size == table.linearSize");

    word = wordsPerRow * x + (y >> LOGBPL);
    bit  = y & (BPL-1);
    result = int32( ((table.linear[word] >> bit) & 1));
    return(result);

} /* end of Cudd_ReadLinear */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [BDD reduction based on combination of sifting and linear
  transformations.]

  Description [BDD reduction based on combination of sifting and linear
  transformations.  Assumes that no dead nodes are present.
    <ol>
    <li> Order all the variables according to the number of entries
    in each unique table.
    <li> Sift the variable up and down, remembering each time the
    total size of the DD heap. At each position, linear transformation
    of the two adjacent variables is tried and is accepted if it reduces
    the size of the DD.
    <li> Select the best permutation.
    <li> Repeat 3 and 4 for all variables.
    </ol>
  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

******************************************************************************/
func cuddLinearAndSifting(table *DdManager, lower int32, upper int32) int32 {
    var i int32;
    var var *int32;
    var size int32;
    var x int32;
    var result int32;
#ifdef DD_STATS
    int         previousSize;
#endif

#ifdef DD_STATS
    ddTotalNumberLinearTr = 0;
#endif

    size = table.size;

    var = nil;
    entry = nil;
    if table.linear == nil {
        result = cuddInitLinear(table);
        if result == 0 { goto cuddLinearAndSiftingOutOfMem; }
     
                                        
                                         
                                                           
      
    } else if table.size != table.linearSize {
        result = cuddResizeLinear(table);
        if result == 0 { goto cuddLinearAndSiftingOutOfMem; }
     
                                        
                                         
                                                           
      
    }

    /* Find order in which to sift variables. */
    entry = ABC_ALLOC(int,size);
    if entry == nil {
        table.errorCode = CUDD_MEMORY_OUT;
        goto cuddLinearAndSiftingOutOfMem;
    }
    var = ABC_ALLOC(int,size);
    if var == nil {
        table.errorCode = CUDD_MEMORY_OUT;
        goto cuddLinearAndSiftingOutOfMem;
    }

    for i = 0; i < size; i++ {
        x = table.perm[i];
        entry[i] = table.subtables[x].keys;
        var[i] = i;
    }

    qsort((void *)var,(size_t)size,sizeof(int),(DD_QSFP)ddLinearUniqueCompare);

    /* Now sift. */
    // inlining of macro ddMin
    for i = 0; i < (((size) < (table.siftMaxVar)) ? (size) : (table.siftMaxVar)); i++ {
        x = table.perm[var[i]];
        if x < lower || x > upper { continue; }
#ifdef DD_STATS
        previousSize = table->keys - table->isolated;
#endif
        result = ddLinearAndSiftingAux(table,x,lower,upper);
        if !(result != 0) { goto cuddLinearAndSiftingOutOfMem; }
#ifdef DD_STATS
        if (table->keys < (unsigned) previousSize + table->isolated) {
            (void) fprintf(table->out,"-");
        } else if (table->keys > (unsigned) previousSize + table->isolated) {
            (void) fprintf(table->out,"+");     /* should never happen */
            (void) fprintf(table->out,"\nSize increased from %d to %d while sifting variable %d\n", previousSize, table->keys - table->isolated, var[i]);
        } else {
            (void) fprintf(table->out,"=");
        }
        fflush(table->out);
#endif
#ifdef DD_DEBUG
        (void) Cudd_DebugCheck(table);
#endif
    }

    ABC_FREE(var);
    ABC_FREE(entry);

#ifdef DD_STATS
    (void) fprintf(table->out,"\n#:L_LINSIFT %8d: linear trans.",
                   ddTotalNumberLinearTr);
#endif

    return(1);

cuddLinearAndSiftingOutOfMem:

    if entry != nil { ABC_FREE(entry); }
    if var != nil { ABC_FREE(var); }

    return(0);

} /* end of cuddLinearAndSifting */


/**Function********************************************************************

  Synopsis    [Linearly combines two adjacent variables.]

  Description [Linearly combines two adjacent variables. Specifically,
  replaces the top variable with the exclusive nor of the two variables.
  It assumes that no dead nodes are present on entry to this
  procedure.  The procedure then guarantees that no dead nodes will be
  present when it terminates.  cuddLinearInPlace assumes that x &lt;
  y.  Returns the number of keys in the table if successful; 0
  otherwise.]

  SideEffects [The two subtables corrresponding to variables x and y are
  modified. The global counters of the unique table are also affected.]

  SeeAlso     [cuddSwapInPlace]

******************************************************************************/
func cuddLinearInPlace(table *DdManager, x int32, y int32) int32 {
    var xlist, ylist *DdNodePtr;
    var xindex, yindex int32;
    var xslots, yslots int32;
    var xshift, yshift int32;
    var oldxkeys, oldykeys int32;
    var newxkeys, newykeys int32;
    var comple, newcomplement int32;
    var i int32;
    var posn int32;
    var isolated int32;
    var f, f0, f1, f01, f00, f11, f10, newf1, newf0 *DdNode;
    var g *DdNode;
    var next *DdNode;
    var last *DdNode = nil;
    var previousP *DdNodePtr;
    var tmp *DdNode;
    var sentinel *DdNode  = &(table.sentinel);
#ifdef DD_DEBUG
    int    count, idcheck;
#endif

#ifdef DD_DEBUG
    assert(x < y);
    assert(cuddNextHigh(table,x) == y);
    assert(table->subtables[x].keys != 0);
    assert(table->subtables[y].keys != 0);
    assert(table->subtables[x].dead == 0);
    assert(table->subtables[y].dead == 0);
#endif

    xindex = table.invperm[x];
    yindex = table.invperm[y];

    if cuddTestInteract(table,xindex,yindex) {
#ifdef DD_STATS
        ddTotalNumberLinearTr++;
#endif
        /* Get parameters of x subtable. */
        xlist = table.subtables[x].nodelist;
        oldxkeys = table.subtables[x].keys;
        xslots = table.subtables[x].slots;
        xshift = table.subtables[x].shift;

        /* Get parameters of y subtable. */
        ylist = table.subtables[y].nodelist;
        oldykeys = table.subtables[y].keys;
        yslots = table.subtables[y].slots;
        yshift = table.subtables[y].shift;

        newxkeys = 0;
        newykeys = oldykeys;

        /* Check whether the two projection functions involved in this
        ** swap are isolated. At the end, we'll be able to tell how many
        ** isolated projection functions are there by checking only these
        ** two functions again. This is done to eliminate the isolated
        ** projection functions from the node count.
        */
        isolated = - ((table.vars[xindex].ref == 1) +
                     (table.vars[yindex].ref == 1));

        /* The nodes in the x layer are put in a chain.
        ** The chain is handled as a FIFO; g points to the beginning and
        ** last points to the end.
        */
        g = nil;
#ifdef DD_DEBUG
        last = NULL;
#endif
        for i = 0; i < xslots; i++ {
            f = xlist[i];
            if f == sentinel { continue; }
            xlist[i] = sentinel;
            if g == nil {
                g = f;
            } else {
                last.next = f;
            }
            for (next = f.next) != sentinel {
                f = next;
            } /* while there are elements in the collision chain */
            last = f;
        } /* for each slot of the x subtable */
#ifdef DD_DEBUG
        /* last is always assigned in the for loop because there is at
        ** least one key */
        assert(last != NULL);
#endif
        last.next = nil;

#ifdef DD_COUNT
        table->swapSteps += oldxkeys;
#endif
        /* Take care of the x nodes that must be re-expressed.
        ** They form a linked list pointed by g.
        */
        f = g;
        for f != nil {
            next = f.next;
            /* Find f1, f0, f11, f10, f01, f00. */
            // inlining of macro cuddT
            f1 = ((f).type.kids.T);
#ifdef DD_DEBUG
            assert(!(Cudd_IsComplement(f1)));
#endif
            if int32( f1.index) == yindex {
                // inlining of macro cuddT
                // inlining of macro cuddE
                f11 = ((f1).type.kids.T); f10 = ((f1).type.kids.E);
            } else {
                f11 = f10 = f1;
            }
#ifdef DD_DEBUG
            assert(!(Cudd_IsComplement(f11)));
#endif
            // inlining of macro cuddE
            f0 = ((f).type.kids.E);
            // inlining of macro Cudd_IsComplement
            comple = (int32( ((ptrint) (f0) & 01)));
            // inlining of macro Cudd_Regular
            f0 = ((DdNode *)((ptruint)(f0) & ~01));
            if int32( f0.index) == yindex {
                // inlining of macro cuddT
                // inlining of macro cuddE
                f01 = ((f0).type.kids.T); f00 = ((f0).type.kids.E);
            } else {
                f01 = f00 = f0;
            }
            if comple {
                // inlining of macro Cudd_Not
                f01 = ((DdNode *)((ptrint)(f01) ^ 01));
                // inlining of macro Cudd_Not
                f00 = ((DdNode *)((ptrint)(f00) ^ 01));
            }
            /* Decrease ref count of f1. */
            cuddSatDec(f1.ref);
            /* Create the new T child. */
            if f11 == f00 {
                newf1 = f11;
                cuddSatInc(newf1.ref);
            } else {
                /* Check ylist for triple (yindex,f11,f00). */
                posn = ddHash(cuddF2L(f11), cuddF2L(f00), yshift);
                /* For each element newf1 in collision list ylist[posn]. */
                previousP = &(ylist[posn]);
                newf1 = *previousP;
                // inlining of macro cuddT
                for f11 < ((newf1).type.kids.T) {
                    previousP = &(newf1.next);
                    newf1 = *previousP;
                }
                // inlining of macro cuddT
                // inlining of macro cuddE
                for f11 == ((newf1).type.kids.T) && f00 < ((newf1).type.kids.E) {
                    previousP = &(newf1.next);
                    newf1 = *previousP;
                }
                // inlining of macro cuddT
                // inlining of macro cuddE
                if ((newf1).type.kids.T) == f11 && ((newf1).type.kids.E) == f00 {
                    cuddSatInc(newf1.ref);
                } else { /* no match */
                    newf1 = cuddDynamicAllocNode(table);
                    if newf1 == nil {
                        goto cuddLinearOutOfMem; }
                    newf1.index = yindex; newf1.ref = 1;
                    // inlining of macro cuddT
                    ((newf1).type.kids.T) = f11;
                    // inlining of macro cuddE
                    ((newf1).type.kids.E) = f00;
                    /* Insert newf1 in the collision list ylist[posn];
                    ** increase the ref counts of f11 and f00.
                    */
                    newykeys++;
                    newf1.next = *previousP;
                    *previousP = newf1;
                    cuddSatInc(f11.ref);
                    // inlining of macro Cudd_Regular
                    tmp = ((DdNode *)((ptruint)(f00) & ~01));
                    cuddSatInc(tmp.ref);
                }
            }
            // inlining of macro cuddT
            ((f).type.kids.T) = newf1;
#ifdef DD_DEBUG
            assert(!(Cudd_IsComplement(newf1)));
#endif

            /* Do the same for f0, keeping complement dots into account. */
            /* decrease ref count of f0 */
            // inlining of macro Cudd_Regular
            tmp = ((DdNode *)((ptruint)(f0) & ~01));
            cuddSatDec(tmp.ref);
            /* create the new E child */
            if f01 == f10 {
                newf0 = f01;
                // inlining of macro Cudd_Regular
                tmp = ((DdNode *)((ptruint)(newf0) & ~01));
                cuddSatInc(tmp.ref);
            } else {
                /* make sure f01 is regular */
                // inlining of macro Cudd_IsComplement
                newcomplement = (int32( ((ptrint) (f01) & 01)));
                if newcomplement != 0 {
                    // inlining of macro Cudd_Not
                    f01 = ((DdNode *)((ptrint)(f01) ^ 01));
                    // inlining of macro Cudd_Not
                    f10 = ((DdNode *)((ptrint)(f10) ^ 01));
                }
                /* Check ylist for triple (yindex,f01,f10). */
                posn = ddHash(cuddF2L(f01), cuddF2L(f10), yshift);
                /* For each element newf0 in collision list ylist[posn]. */
                previousP = &(ylist[posn]);
                newf0 = *previousP;
                // inlining of macro cuddT
                for f01 < ((newf0).type.kids.T) {
                    previousP = &(newf0.next);
                    newf0 = *previousP;
                }
                // inlining of macro cuddT
                // inlining of macro cuddE
                for f01 == ((newf0).type.kids.T) && f10 < ((newf0).type.kids.E) {
                    previousP = &(newf0.next);
                    newf0 = *previousP;
                }
                // inlining of macro cuddT
                // inlining of macro cuddE
                if ((newf0).type.kids.T) == f01 && ((newf0).type.kids.E) == f10 {
                    cuddSatInc(newf0.ref);
                } else { /* no match */
                    newf0 = cuddDynamicAllocNode(table);
                    if newf0 == nil {
                        goto cuddLinearOutOfMem; }
                    newf0.index = yindex; newf0.ref = 1;
                    // inlining of macro cuddT
                    ((newf0).type.kids.T) = f01;
                    // inlining of macro cuddE
                    ((newf0).type.kids.E) = f10;
                    /* Insert newf0 in the collision list ylist[posn];
                    ** increase the ref counts of f01 and f10.
                    */
                    newykeys++;
                    newf0.next = *previousP;
                    *previousP = newf0;
                    cuddSatInc(f01.ref);
                    // inlining of macro Cudd_Regular
                    tmp = ((DdNode *)((ptruint)(f10) & ~01));
                    cuddSatInc(tmp.ref);
                }
                if newcomplement != 0 {
                    // inlining of macro Cudd_Not
                    newf0 = ((DdNode *)((ptrint)(newf0) ^ 01));
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
            previousP = &(xlist[posn]);
            tmp = *previousP;
            // inlining of macro cuddT
            for newf1 < ((tmp).type.kids.T) {
                previousP = &(tmp.next);
                tmp = *previousP;
            }
            // inlining of macro cuddT
            // inlining of macro cuddE
            for newf1 == ((tmp).type.kids.T) && newf0 < ((tmp).type.kids.E) {
                previousP = &(tmp.next);
                tmp = *previousP;
            }
            f.next = *previousP;
            *previousP = f;
            f = next;
        } /* while f != NULL */

        /* GC the y layer. */

        /* For each node f in ylist. */
        for i = 0; i < yslots; i++ {
            previousP = &(ylist[i]);
            f = *previousP;
            for f != sentinel {
                next = f.next;
                if f.ref == 0 {
                    // inlining of macro cuddT
                    tmp = ((f).type.kids.T);
                    cuddSatDec(tmp.ref);
                    // inlining of macro Cudd_Regular
                    // inlining of macro cuddE
                    tmp = ((DdNode *)((ptruint)(((f).type.kids.E)) & ~01));
                    cuddSatDec(tmp.ref);
                    // inlining of macro cuddDeallocNode
                    (f).next = (table).nextFree;     (table).nextFree = f;;
                    newykeys--;
                } else {
                    *previousP = f;
                    previousP = &(f.next);
                }
                f = next;
            } /* while f */
            *previousP = sentinel;
        } /* for every collision list */

#ifdef DD_DEBUG
#if 0
        (void) fprintf(table->out,"Linearly combining %d and %d\n",x,y);
#endif
        count = 0;
        idcheck = 0;
        for (i = 0; i < yslots; i++) {
            f = ylist[i];
            while (f != sentinel) {
                count++;
                if (f->index != (DdHalfWord) yindex)
                    idcheck++;
                f = f->next;
            }
        }
        if (count != newykeys) {
            fprintf(table->err,"Error in finding newykeys\toldykeys = %d\tnewykeys = %d\tactual = %d\n",oldykeys,newykeys,count);
        }
        if (idcheck != 0)
            fprintf(table->err,"Error in id's of ylist\twrong id's = %d\n",idcheck);
        count = 0;
        idcheck = 0;
        for (i = 0; i < xslots; i++) {
            f = xlist[i];
            while (f != sentinel) {
                count++;
                if (f->index != (DdHalfWord) xindex)
                    idcheck++;
                f = f->next;
            }
        }
        if (count != newxkeys || newxkeys != oldxkeys) {
            fprintf(table->err,"Error in finding newxkeys\toldxkeys = %d \tnewxkeys = %d \tactual = %d\n",oldxkeys,newxkeys,count);
        }
        if (idcheck != 0)
            fprintf(table->err,"Error in id's of xlist\twrong id's = %d\n",idcheck);
#endif

        isolated += (table.vars[xindex].ref == 1) +
                    (table.vars[yindex].ref == 1);
        table.isolated += isolated;

        /* Set the appropriate fields in table. */
        table.subtables[y].keys = newykeys;

        /* Here we should update the linear combination table
        ** to record that x <- x EXNOR y. This is done by complementing
        ** the (x,y) entry of the table.
        */

        table.keys += newykeys - oldykeys;

        cuddXorLinear(table,xindex,yindex);
    }

#ifdef DD_DEBUG
    if (zero) {
        (void) Cudd_DebugCheck(table);
    }
#endif

    return(table.keys - table.isolated);

cuddLinearOutOfMem:
    (void) fmt.Fprintf(table.err, "Error: cuddLinearInPlace out of memory\n");

    return (0);

} /* end of cuddLinearInPlace */


/**Function********************************************************************

  Synopsis    [Updates the interaction matrix.]

  Description []

  SideEffects [none]

  SeeAlso     []

******************************************************************************/
func cuddUpdateInteractionMatrix(table *DdManager, xindex int32, yindex int32) {
    var i int32;
    for i = 0; i < yindex; i++ {
        if i != xindex && cuddTestInteract(table,i,yindex) {
            if i < xindex {
                cuddSetInteract(table,i,xindex);
            } else {
                cuddSetInteract(table,xindex,i);
            }
        }
    }
    for i = yindex+1; i < table.size; i++ {
        if i != xindex && cuddTestInteract(table,yindex,i) {
            if i < xindex {
                cuddSetInteract(table,i,xindex);
            } else {
                cuddSetInteract(table,xindex,i);
            }
        }
    }

} /* end of cuddUpdateInteractionMatrix */


/**Function********************************************************************

  Synopsis    [Initializes the linear transform matrix.]

  Description [Initializes the linear transform matrix.  Returns 1 if
  successful; 0 otherwise.]

  SideEffects [none]

  SeeAlso     []

******************************************************************************/
func cuddInitLinear(table *DdManager) int32 {
    var words int32;
    var wordsPerRow int32;
    var nvars int32;
    var word int32;
    var bit int32;
    var i int32;
    var linear *long;

    nvars = table.size;
    wordsPerRow = ((nvars - 1) >> LOGBPL) + 1;
    words = wordsPerRow * nvars;
    table.linear = linear = ABC_ALLOC(long,words);
    if linear == nil {
        table.errorCode = CUDD_MEMORY_OUT;
        return(0);
    }
    table.memused += words * sizeof(long);
    table.linearSize = nvars;
    for i = 0; i < words; i++ { linear[i] = 0; }
    for i = 0; i < nvars; i++ {
        word = wordsPerRow * i + (i >> LOGBPL);
        bit  = i & (BPL-1);
        linear[word] = 1 << bit;
    }
    return(1);

} /* end of cuddInitLinear */


/**Function********************************************************************

  Synopsis    [Resizes the linear transform matrix.]

  Description [Resizes the linear transform matrix.  Returns 1 if
  successful; 0 otherwise.]

  SideEffects [none]

  SeeAlso     []

******************************************************************************/
func cuddResizeLinear(table *DdManager) int32 {
    var words, oldWords int32;
    var wordsPerRow, oldWordsPerRow int32;
    var nvars, oldNvars int32;
    var word, oldWord int32;
    var bit int32;
    var i, j int32;
    var linear, oldLinear *long;

    oldNvars = table.linearSize;
    oldWordsPerRow = ((oldNvars - 1) >> LOGBPL) + 1;
    oldWords = oldWordsPerRow * oldNvars;
    oldLinear = table.linear;

    nvars = table.size;
    wordsPerRow = ((nvars - 1) >> LOGBPL) + 1;
    words = wordsPerRow * nvars;
    table.linear = linear = ABC_ALLOC(long,words);
    if linear == nil {
        table.errorCode = CUDD_MEMORY_OUT;
        return(0);
    }
    table.memused += (words - oldWords) * sizeof(long);
    for i = 0; i < words; i++ { linear[i] = 0; }

    /* Copy old matrix. */
    for i = 0; i < oldNvars; i++ {
        for j = 0; j < oldWordsPerRow; j++ {
            oldWord = oldWordsPerRow * i + j;
            word = wordsPerRow * i + j;
            linear[word] = oldLinear[oldWord];
        }
    }
    ABC_FREE(oldLinear);

    /* Add elements to the diagonal. */
    for i = oldNvars; i < nvars; i++ {
        word = wordsPerRow * i + (i >> LOGBPL);
        bit  = i & (BPL-1);
        linear[word] = 1 << bit;
    }
    table.linearSize = nvars;

    return(1);

} /* end of cuddResizeLinear */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Comparison function used by qsort.]

  Description [Comparison function used by qsort to order the
  variables according to the number of keys in the subtables.
  Returns the difference in number of keys between the two
  variables being compared.]

  SideEffects [None]

******************************************************************************/
func ddLinearUniqueCompare(ptrX *int32, ptrY *int32) int32 {
     
                                       
                                  
     
      
    return(entry[*ptrY] - entry[*ptrX]);

} /* end of ddLinearUniqueCompare */


/**Function********************************************************************

  Synopsis    [Given xLow <= x <= xHigh moves x up and down between the
  boundaries.]

  Description [Given xLow <= x <= xHigh moves x up and down between the
  boundaries. At each step a linear transformation is tried, and, if it
  decreases the size of the DD, it is accepted. Finds the best position
  and does the required changes.  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

******************************************************************************/
func ddLinearAndSiftingAux(table *DdManager, x int32, xLow int32, xHigh int32) int32 {

    var move *Move;
    var moveUp *Move;                /* list of up moves */
    var moveDown *Move;              /* list of down moves */
    var initialSize int32;
    var result int32;

    initialSize = table.keys - table.isolated;

    moveDown = nil;
    moveUp = nil;

    if x == xLow {
        moveDown = ddLinearAndSiftingDown(table,x,xHigh,nil);
        /* At this point x --> xHigh unless bounding occurred. */
        if moveDown == (Move *) CUDD_OUT_OF_MEM { goto ddLinearAndSiftingAuxOutOfMem; }
        /* Move backward and stop at best position. */
        result = ddLinearAndSiftingBackward(table,initialSize,moveDown);
        if !(result != 0) { goto ddLinearAndSiftingAuxOutOfMem; }

    } else if x == xHigh {
        moveUp = ddLinearAndSiftingUp(table,x,xLow,nil);
        /* At this point x --> xLow unless bounding occurred. */
        if moveUp == (Move *) CUDD_OUT_OF_MEM { goto ddLinearAndSiftingAuxOutOfMem; }
        /* Move backward and stop at best position. */
        result = ddLinearAndSiftingBackward(table,initialSize,moveUp);
        if !(result != 0) { goto ddLinearAndSiftingAuxOutOfMem; }

    } else if (x - xLow) > (xHigh - x) { /* must go down first: shorter */
        moveDown = ddLinearAndSiftingDown(table,x,xHigh,nil);
        /* At this point x --> xHigh unless bounding occurred. */
        if moveDown == (Move *) CUDD_OUT_OF_MEM { goto ddLinearAndSiftingAuxOutOfMem; }
        moveUp = ddUndoMoves(table,moveDown);
#ifdef DD_DEBUG
        assert(moveUp == NULL || moveUp->x == x);
#endif
        moveUp = ddLinearAndSiftingUp(table,x,xLow,moveUp);
        if moveUp == (Move *) CUDD_OUT_OF_MEM { goto ddLinearAndSiftingAuxOutOfMem; }
        /* Move backward and stop at best position. */
        result = ddLinearAndSiftingBackward(table,initialSize,moveUp);
        if !(result != 0) { goto ddLinearAndSiftingAuxOutOfMem; }

    } else { /* must go up first: shorter */
        moveUp = ddLinearAndSiftingUp(table,x,xLow,nil);
        /* At this point x --> xLow unless bounding occurred. */
        if moveUp == (Move *) CUDD_OUT_OF_MEM { goto ddLinearAndSiftingAuxOutOfMem; }
        moveDown = ddUndoMoves(table,moveUp);
#ifdef DD_DEBUG
        assert(moveDown == NULL || moveDown->y == x);
#endif
        moveDown = ddLinearAndSiftingDown(table,x,xHigh,moveDown);
        if moveDown == (Move *) CUDD_OUT_OF_MEM { goto ddLinearAndSiftingAuxOutOfMem; }
        /* Move backward and stop at best position. */
        result = ddLinearAndSiftingBackward(table,initialSize,moveDown);
        if !(result != 0) { goto ddLinearAndSiftingAuxOutOfMem; }
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

ddLinearAndSiftingAuxOutOfMem:
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

} /* end of ddLinearAndSiftingAux */


/**Function********************************************************************

  Synopsis    [Sifts a variable up and applies linear transformations.]

  Description [Sifts a variable up and applies linear transformations.
  Moves y up until either it reaches the bound (xLow) or the size of
  the DD heap increases too much.  Returns the set of moves in case of
  success; NULL if memory is full.]

  SideEffects [None]

******************************************************************************/
func ddLinearAndSiftingUp(table *DdManager, y int32, xLow int32, prevMoves *Move) *Move {
    var moves *Move;
    var move *Move;
    var x int32;
    var size, newsize int32;
    var limitSize int32;
    var xindex, yindex int32;
    var isolated int32;
    var L int32;      /* lower bound on DD size */
#ifdef DD_DEBUG
    int checkL;
    int z;
    int zindex;
#endif

    moves = prevMoves;
    yindex = table.invperm[y];

    /* Initialize the lower bound.
    ** The part of the DD below y will not change.
    ** The part of the DD above y that does not interact with y will not
    ** change. The rest may vanish in the best case, except for
    ** the nodes at level xLow, which will not vanish, regardless.
    */
    limitSize = L = table.keys - table.isolated;
    for x = xLow + 1; x < y; x++ {
        xindex = table.invperm[x];
        if cuddTestInteract(table,xindex,yindex) {
            isolated = table.vars[xindex].ref == 1;
            L -= table.subtables[x].keys - isolated;
        }
    }
    isolated = table.vars[yindex].ref == 1;
    L -= table.subtables[y].keys - isolated;

    x = cuddNextLow(table,y);
    for x >= xLow && L <= limitSize {
        xindex = table.invperm[x];
#ifdef DD_DEBUG
        checkL = table->keys - table->isolated;
        for (z = xLow + 1; z < y; z++) {
            zindex = table->invperm[z];
            if (cuddTestInteract(table,zindex,yindex)) {
                isolated = table->vars[zindex]->ref == 1;
                checkL -= table->subtables[z].keys - isolated;
            }
        }
        isolated = table->vars[yindex]->ref == 1;
        checkL -= table->subtables[y].keys - isolated;
        if (L != checkL) {
            (void) fprintf(table->out, "checkL(%d) != L(%d)\n",checkL,L);
        }
#endif
        size = cuddSwapInPlace(table,x,y);
        if size == 0 { goto ddLinearAndSiftingUpOutOfMem; }
        newsize = cuddLinearInPlace(table,x,y);
        if newsize == 0 { goto ddLinearAndSiftingUpOutOfMem; }
        move = (Move *) cuddDynamicAllocNode(table);
        if move == nil { goto ddLinearAndSiftingUpOutOfMem; }
        move.x = x;
        move.y = y;
        move.next = moves;
        moves = move;
        move.flags = CUDD_SWAP_MOVE;
        if newsize >= size {
            /* Undo transformation. The transformation we apply is
            ** its own inverse. Hence, we just apply the transformation
            ** again.
            */
            newsize = cuddLinearInPlace(table,x,y);
            if newsize == 0 { goto ddLinearAndSiftingUpOutOfMem; }
#ifdef DD_DEBUG
            if (newsize != size) {
                (void) fprintf(table->out,"Change in size after identity transformation! From %d to %d\n",size,newsize);
            }
#endif
        } else if cuddTestInteract(table,xindex,yindex) {
            size = newsize;
            move.flags = CUDD_LINEAR_TRANSFORM_MOVE;
            cuddUpdateInteractionMatrix(table,xindex,yindex);
        }
        move.size = size;
        /* Update the lower bound. */
        if cuddTestInteract(table,xindex,yindex) {
            isolated = table.vars[xindex].ref == 1;
            L += table.subtables[y].keys - isolated;
        }
        if float64( size) > float64( limitSize) * table.maxGrowth { break; }
        if size < limitSize { limitSize = size; }
        y = x;
        x = cuddNextLow(table,y);
    }
    return(moves);

ddLinearAndSiftingUpOutOfMem:
    for moves != nil {
        move = moves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moves)).ref = 0;     ((DdNode *)(moves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moves);;
        moves = move;
    }
    return((Move *) CUDD_OUT_OF_MEM);

} /* end of ddLinearAndSiftingUp */


/**Function********************************************************************

  Synopsis    [Sifts a variable down and applies linear transformations.]

  Description [Sifts a variable down and applies linear
  transformations. Moves x down until either it reaches the bound
  (xHigh) or the size of the DD heap increases too much. Returns the
  set of moves in case of success; NULL if memory is full.]

  SideEffects [None]

******************************************************************************/
func ddLinearAndSiftingDown(table *DdManager, x int32, xHigh int32, prevMoves *Move) *Move {
    var moves *Move;
    var move *Move;
    var y int32;
    var size, newsize int32;
    var R int32;      /* upper bound on node decrease */
    var limitSize int32;
    var xindex, yindex int32;
    var isolated int32;
#ifdef DD_DEBUG
    int         checkR;
    int         z;
    int         zindex;
#endif

    moves = prevMoves;
    /* Initialize R */
    xindex = table.invperm[x];
    limitSize = size = table.keys - table.isolated;
    R = 0;
    for y = xHigh; y > x; y-- {
        yindex = table.invperm[y];
        if cuddTestInteract(table,xindex,yindex) {
            isolated = table.vars[yindex].ref == 1;
            R += table.subtables[y].keys - isolated;
        }
    }

    y = cuddNextHigh(table,x);
    for y <= xHigh && size - R < limitSize {
#ifdef DD_DEBUG
        checkR = 0;
        for (z = xHigh; z > x; z--) {
            zindex = table->invperm[z];
            if (cuddTestInteract(table,xindex,zindex)) {
                isolated = table->vars[zindex]->ref == 1;
                checkR += table->subtables[z].keys - isolated;
            }
        }
        if (R != checkR) {
            (void) fprintf(table->out, "checkR(%d) != R(%d)\n",checkR,R);
        }
#endif
        /* Update upper bound on node decrease. */
        yindex = table.invperm[y];
        if cuddTestInteract(table,xindex,yindex) {
            isolated = table.vars[yindex].ref == 1;
            R -= table.subtables[y].keys - isolated;
        }
        size = cuddSwapInPlace(table,x,y);
        if size == 0 { goto ddLinearAndSiftingDownOutOfMem; }
        newsize = cuddLinearInPlace(table,x,y);
        if newsize == 0 { goto ddLinearAndSiftingDownOutOfMem; }
        move = (Move *) cuddDynamicAllocNode(table);
        if move == nil { goto ddLinearAndSiftingDownOutOfMem; }
        move.x = x;
        move.y = y;
        move.next = moves;
        moves = move;
        move.flags = CUDD_SWAP_MOVE;
        if newsize >= size {
            /* Undo transformation. The transformation we apply is
            ** its own inverse. Hence, we just apply the transformation
            ** again.
            */
            newsize = cuddLinearInPlace(table,x,y);
            if newsize == 0 { goto ddLinearAndSiftingDownOutOfMem; }
            if newsize != size {
                (void) fmt.Fprintf(table.out, "Change in size after identity transformation! From %d to %d\n", size, newsize);
            }
        } else if cuddTestInteract(table,xindex,yindex) {
            size = newsize;
            move.flags = CUDD_LINEAR_TRANSFORM_MOVE;
            cuddUpdateInteractionMatrix(table,xindex,yindex);
        }
        move.size = size;
        if float64( size) > float64( limitSize) * table.maxGrowth { break; }
        if size < limitSize { limitSize = size; }
        x = y;
        y = cuddNextHigh(table,x);
    }
    return(moves);

ddLinearAndSiftingDownOutOfMem:
    for moves != nil {
        move = moves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moves)).ref = 0;     ((DdNode *)(moves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moves);;
        moves = move;
    }
    return((Move *) CUDD_OUT_OF_MEM);

} /* end of ddLinearAndSiftingDown */


/**Function********************************************************************

  Synopsis    [Given a set of moves, returns the DD heap to the order
  giving the minimum size.]

  Description [Given a set of moves, returns the DD heap to the
  position giving the minimum size. In case of ties, returns to the
  closest position giving the minimum size. Returns 1 in case of
  success; 0 otherwise.]

  SideEffects [None]

******************************************************************************/
func ddLinearAndSiftingBackward(table *DdManager, size int32, moves *Move) int32 {
    var move *Move;
    var res int32;

    for move = moves; move != nil; move = move.next {
        if move.size < size {
            size = move.size;
        }
    }

    for move = moves; move != nil; move = move.next {
        if move.size == size { return(1); }
        if move.flags == CUDD_LINEAR_TRANSFORM_MOVE {
            res = cuddLinearInPlace(table,int32(move.x),int32(move.y));
            if !(res != 0) { return(0); }
        }
        res = cuddSwapInPlace(table,int32(move.x),int32(move.y));
        if !(res != 0) { return(0); }
        if move.flags == CUDD_INVERSE_TRANSFORM_MOVE {
            res = cuddLinearInPlace(table,int32(move.x),int32(move.y));
            if !(res != 0) { return(0); }
        }
    }

    return(1);

} /* end of ddLinearAndSiftingBackward */


/**Function********************************************************************

  Synopsis    [Given a set of moves, returns the DD heap to the order
  in effect before the moves.]

  Description [Given a set of moves, returns the DD heap to the
  order in effect before the moves.  Returns 1 in case of success;
  0 otherwise.]

  SideEffects [None]

******************************************************************************/
func ddUndoMoves(table *DdManager, moves *Move) *Move {
    var invmoves *Move  = nil;
    var move *Move;
    var invmove *Move;
    var size int32;

    for move = moves; move != nil; move = move.next {
        invmove = (Move *) cuddDynamicAllocNode(table);
        if invmove == nil { goto ddUndoMovesOutOfMem; }
        invmove.x = move.x;
        invmove.y = move.y;
        invmove.next = invmoves;
        invmoves = invmove;
        if move.flags == CUDD_SWAP_MOVE {
            invmove.flags = CUDD_SWAP_MOVE;
            size = cuddSwapInPlace(table,int32(move.x),int32(move.y));
            if !(size != 0) { goto ddUndoMovesOutOfMem; }
        } else if move.flags == CUDD_LINEAR_TRANSFORM_MOVE {
            invmove.flags = CUDD_INVERSE_TRANSFORM_MOVE;
            size = cuddLinearInPlace(table,int32(move.x),int32(move.y));
            if !(size != 0) { goto ddUndoMovesOutOfMem; }
            size = cuddSwapInPlace(table,int32(move.x),int32(move.y));
            if !(size != 0) { goto ddUndoMovesOutOfMem; }
        } else { /* must be CUDD_INVERSE_TRANSFORM_MOVE */
#ifdef DD_DEBUG
            (void) fprintf(table->err,"Unforseen event in ddUndoMoves!\n");
#endif
            invmove.flags = CUDD_LINEAR_TRANSFORM_MOVE;
            size = cuddSwapInPlace(table,int32(move.x),int32(move.y));
            if !(size != 0) { goto ddUndoMovesOutOfMem; }
            size = cuddLinearInPlace(table,int32(move.x),int32(move.y));
            if !(size != 0) { goto ddUndoMovesOutOfMem; }
        }
        invmove.size = size;
    }

    return(invmoves);

ddUndoMovesOutOfMem:
    for invmoves != nil {
        move = invmoves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(invmoves)).ref = 0;     ((DdNode *)(invmoves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(invmoves);;
        invmoves = move;
    }
    return((Move *) CUDD_OUT_OF_MEM);

} /* end of ddUndoMoves */


/**Function********************************************************************

  Synopsis    [XORs two rows of the linear transform matrix.]

  Description [XORs two rows of the linear transform matrix and replaces
  the first row with the result.]

  SideEffects [none]

  SeeAlso     []

******************************************************************************/
func cuddXorLinear(table *DdManager, x int32, y int32) {
    var i int32;
    var nvars int32  = table.size;
    var wordsPerRow int32  = ((nvars - 1) >> LOGBPL) + 1;
    var xstart int32  = wordsPerRow * x;
    var ystart int32  = wordsPerRow * y;
    var linear *long  = table.linear;

    for i = 0; i < wordsPerRow; i++ {
        linear[xstart+i] ^= linear[ystart+i];
    }

} /* end of cuddXorLinear */


//ABC_NAMESPACE_IMPL_END


package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddSymmetry.c]

  PackageName [cudd]

  Synopsis    [Functions for symmetry-based variable reordering.]

  Description [External procedures included in this file:
                <ul>
                <li> Cudd_SymmProfile()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddSymmCheck()
                <li> cuddSymmSifting()
                <li> cuddSymmSiftingConv()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> ddSymmUniqueCompare()
                <li> ddSymmSiftingAux()
                <li> ddSymmSiftingConvAux()
                <li> ddSymmSiftingUp()
                <li> ddSymmSiftingDown()
                <li> ddSymmGroupMove()
                <li> ddSymmGroupMoveBackward()
                <li> ddSymmSiftingBackward()
                <li> ddSymmSummary()
                </ul>]

  Author      [Shipra Panda, Fabio Somenzi]

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

#define MV_OOM (Move *)1

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
static char rcsid[] DD_UNUSED = "$Id: cuddSymmetry.c,v 1.26 2009/02/19 16:23:54 fabio Exp $";
#endif

var entry *int32;

var ddTotalNumberSwapping int32;
#ifdef DD_STATS
extern  int     ddTotalNISwaps;
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int ddSymmUniqueCompare (int *ptrX, int *ptrY);

// static int ddSymmSiftingAux (DdManager *table, int x, int xLow, int xHigh);

// static int ddSymmSiftingConvAux (DdManager *table, int x, int xLow, int xHigh);

// static Move * ddSymmSiftingUp (DdManager *table, int y, int xLow);

// static Move * ddSymmSiftingDown (DdManager *table, int x, int xHigh);

// static int ddSymmGroupMove (DdManager *table, int x, int y, Move **moves);

// static int ddSymmGroupMoveBackward (DdManager *table, int x, int y);

// static int ddSymmSiftingBackward (DdManager *table, Move *moves, int size);

// static void ddSymmSummary (DdManager *table, int lower, int upper, int *symvars, int *symgroups);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Prints statistics on symmetric variables.]

  Description []

  SideEffects [None]

******************************************************************************/
func Cudd_SymmProfile(table *DdManager, lower int32, upper int32) {
    var i, x, gbot int32;
    var TotalSymm int32;
    var TotalSymmGroups int32;

    for i = lower; i <= upper; i++ {
        if table.subtables[i].next != uint32( i) {
            x = i;
            (void) fmt.Fprintf(table.out, "Group:");
            for {
                (void) fmt.Fprintf(table.out, "  %d", table.invperm[x]);
                TotalSymm++;
                gbot = x;
                x = table.subtables[x].next;
            
if !(x != i) { break }
} 
            TotalSymmGroups++;
#ifdef DD_DEBUG
            assert(table->subtables[gbot].next == (unsigned) i);
#endif
            i = gbot;
            (void) fmt.Fprintf(table.out, "\n");
        }
    }
    (void) fmt.Fprintf(table.out, "Total Symmetric = %d\n", TotalSymm);
    (void) fmt.Fprintf(table.out, "Total Groups = %d\n", TotalSymmGroups);

} /* end of Cudd_SymmProfile */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Checks for symmetry of x and y.]

  Description [Checks for symmetry of x and y. Ignores projection
  functions, unless they are isolated. Returns 1 in case of symmetry; 0
  otherwise.]

  SideEffects [None]

******************************************************************************/
func cuddSymmCheck(table *DdManager, x int32, y int32) int32 {
    var f, f0, f1, f01, f00, f11, f10 *DdNode;
    var comple int32;         /* f0 is complemented */
    var xsymmy int32;         /* x and y may be positively symmetric */
    var xsymmyp int32;        /* x and y may be negatively symmetric */
    var arccount int32;       /* number of arcs from layer x to layer y */
    var TotalRefCount int32;  /* total reference count of layer y minus 1 */
    var yindex int32;
    var i int32;
    var list *DdNodePtr;
    var slots int32;
    var sentinel *DdNode  = &(table.sentinel);
#ifdef DD_DEBUG
    int xindex;
#endif

    /* Checks that x and y are not the projection functions.
    ** For x it is sufficient to check whether there is only one
    ** node; indeed, if there is one node, it is the projection function
    ** and it cannot point to y. Hence, if y isn't just the projection
    ** function, it has one arc coming from a layer different from x.
    */
    if table.subtables[x].keys == 1 {
        return(0);
    }
    yindex = table.invperm[y];
    if table.subtables[y].keys == 1 {
        if table.vars[yindex].ref == 1 {
            return(0); }
    }

    xsymmy = xsymmyp = 1;
    arccount = 0;
    slots = table.subtables[x].slots;
    list = table.subtables[x].nodelist;
    for i = 0; i < slots; i++ {
        f = list[i];
        for f != sentinel {
            /* Find f1, f0, f11, f10, f01, f00. */
            // inlining of macro cuddT
            f1 = ((f).type.kids.T);
            // inlining of macro Cudd_Regular
            // inlining of macro cuddE
            f0 = ((DdNode *)((ptruint)(((f).type.kids.E)) & ~01));
            // inlining of macro Cudd_IsComplement
            // inlining of macro cuddE
            comple = (int32( ((ptrint) (((f).type.kids.E)) & 01)));
            if int32( f1.index) == yindex {
                arccount++;
                // inlining of macro cuddT
                // inlining of macro cuddE
                f11 = ((f1).type.kids.T); f10 = ((f1).type.kids.E);
            } else {
                if int32( f0.index) != yindex {
                    /* If f is an isolated projection function it is
                    ** allowed to bypass layer y.
                    */
                    // inlining of macro DD_ONE
                    if f1 != ((table).one) || f0 != ((table).one) || f.ref != 1 {
                        return(0); } /* f bypasses layer y */
                }
                f11 = f10 = f1;
            }
            if int32( f0.index) == yindex {
                arccount++;
                // inlining of macro cuddT
                // inlining of macro cuddE
                f01 = ((f0).type.kids.T); f00 = ((f0).type.kids.E);
            } else {
                f01 = f00 = f0;
            }
            if comple != 0 {
                // inlining of macro Cudd_Not
                f01 = ((DdNode *)((ptrint)(f01) ^ 01));
                // inlining of macro Cudd_Not
                f00 = ((DdNode *)((ptrint)(f00) ^ 01));
            }

            // inlining of macro DD_ONE
            if f1 != ((table).one) || f0 != ((table).one) || f.ref != 1 {
                xsymmy &= f01 == f10;
                xsymmyp &= f11 == f00;
                if (xsymmy == 0) && (xsymmyp == 0) {
                    return(0); }
            }

            f = f.next;
        } /* while */
    } /* for */

    /* Calculate the total reference counts of y */
    TotalRefCount = -1; /* -1 for projection function */
    slots = table.subtables[y].slots;
    list = table.subtables[y].nodelist;
    for i = 0; i < slots; i++ {
        f = list[i];
        for f != sentinel {
            TotalRefCount += f.ref;
            f = f.next;
        }
    }

#if defined(DD_DEBUG) && defined(DD_VERBOSE)
    if (arccount == TotalRefCount) {
        xindex = table->invperm[x];
        (void) fprintf(table->out,
                       "Found symmetry! x =%d\ty = %d\tPos(%d,%d)\n",
                       xindex,yindex,x,y);
    }
#endif

    return(arccount == TotalRefCount);

} /* end of cuddSymmCheck */


/**Function********************************************************************

  Synopsis    [Symmetric sifting algorithm.]

  Description [Symmetric sifting algorithm.
  Assumes that no dead nodes are present.
    <ol>
    <li> Order all the variables according to the number of entries in
    each unique subtable.
    <li> Sift the variable up and down, remembering each time the total
    size of the DD heap and grouping variables that are symmetric.
    <li> Select the best permutation.
    <li> Repeat 3 and 4 for all variables.
    </ol>
  Returns 1 plus the number of symmetric variables if successful; 0
  otherwise.]

  SideEffects [None]

  SeeAlso     [cuddSymmSiftingConv]

******************************************************************************/
func cuddSymmSifting(table *DdManager, lower int32, upper int32) int32 {
    var i int32;
    var var *int32;
    var size int32;
    var x int32;
    var result int32;
    var symvars int32;
    var symgroups int32;
#ifdef DD_STATS
    int         previousSize;
#endif

    size = table.size;

    /* Find order in which to sift variables. */
    var = nil;
    entry = ABC_ALLOC(int,size);
    if entry == nil {
        table.errorCode = CUDD_MEMORY_OUT;
        goto ddSymmSiftingOutOfMem;
    }
    var = ABC_ALLOC(int,size);
    if var == nil {
        table.errorCode = CUDD_MEMORY_OUT;
        goto ddSymmSiftingOutOfMem;
    }

    for i = 0; i < size; i++ {
        x = table.perm[i];
        entry[i] = table.subtables[x].keys;
        var[i] = i;
    }

    qsort((void *)var,(size_t)size,sizeof(int),(DD_QSFP)ddSymmUniqueCompare);

    /* Initialize the symmetry of each subtable to itself. */
    for i = lower; i <= upper; i++ {
        table.subtables[i].next = i;
    }

    // inlining of macro ddMin
    for i = 0; i < (((size) < (table.siftMaxVar)) ? (size) : (table.siftMaxVar)); i++ {
        if ddTotalNumberSwapping >= table.siftMaxSwap {
            break; }
        // enable timeout during variable reodering - alanmi 2/13/11
        if table.TimeStop && Abc_Clock() > table.TimeStop {
            break; }
        x = table.perm[var[i]];
#ifdef DD_STATS
        previousSize = table->keys - table->isolated;
#endif
        if x < lower || x > upper { continue; }
        if table.subtables[x].next == uint32( x) {
            result = ddSymmSiftingAux(table,x,lower,upper);
            if !(result != 0) { goto ddSymmSiftingOutOfMem; }
#ifdef DD_STATS
            if (table->keys < (unsigned) previousSize + table->isolated) {
                (void) fprintf(table->out,"-");
            } else if (table->keys > (unsigned) previousSize +
                       table->isolated) {
                (void) fprintf(table->out,"+"); /* should never happen */
            } else {
                (void) fprintf(table->out,"=");
            }
            fflush(table->out);
#endif
        }
    }

    ABC_FREE(var);
    ABC_FREE(entry);

    ddSymmSummary(table, lower, upper, &symvars, &symgroups);

#ifdef DD_STATS
    (void) fprintf(table->out, "\n#:S_SIFTING %8d: symmetric variables\n",
                   symvars);
    (void) fprintf(table->out, "#:G_SIFTING %8d: symmetric groups",
                   symgroups);
#endif

    return(1+symvars);

ddSymmSiftingOutOfMem:

    if entry != nil { ABC_FREE(entry); }
    if var != nil { ABC_FREE(var); }

    return(0);

} /* end of cuddSymmSifting */


/**Function********************************************************************

  Synopsis    [Symmetric sifting to convergence algorithm.]

  Description [Symmetric sifting to convergence algorithm.
  Assumes that no dead nodes are present.
    <ol>
    <li> Order all the variables according to the number of entries in
    each unique subtable.
    <li> Sift the variable up and down, remembering each time the total
    size of the DD heap and grouping variables that are symmetric.
    <li> Select the best permutation.
    <li> Repeat 3 and 4 for all variables.
    <li> Repeat 1-4 until no further improvement.
    </ol>
  Returns 1 plus the number of symmetric variables if successful; 0
  otherwise.]

  SideEffects [None]

  SeeAlso     [cuddSymmSifting]

******************************************************************************/
func cuddSymmSiftingConv(table *DdManager, lower int32, upper int32) int32 {
    var i int32;
    var var *int32;
    var size int32;
    var x int32;
    var result int32;
    var symvars int32;
    var symgroups int32;
    var classes int32;
    var initialSize int32;
#ifdef DD_STATS
    int         previousSize;
#endif

    initialSize = table.keys - table.isolated;

    size = table.size;

    /* Find order in which to sift variables. */
    var = nil;
    entry = ABC_ALLOC(int,size);
    if entry == nil {
        table.errorCode = CUDD_MEMORY_OUT;
        goto ddSymmSiftingConvOutOfMem;
    }
    var = ABC_ALLOC(int,size);
    if var == nil {
        table.errorCode = CUDD_MEMORY_OUT;
        goto ddSymmSiftingConvOutOfMem;
    }

    for i = 0; i < size; i++ {
        x = table.perm[i];
        entry[i] = table.subtables[x].keys;
        var[i] = i;
    }

    qsort((void *)var,(size_t)size,sizeof(int),(DD_QSFP)ddSymmUniqueCompare);

    /* Initialize the symmetry of each subtable to itself
    ** for first pass of converging symmetric sifting.
    */
    for i = lower; i <= upper; i++ {
        table.subtables[i].next = i;
    }

    // inlining of macro ddMin
    for i = 0; i < (((table.size) < (table.siftMaxVar)) ? (table.size) : (table.siftMaxVar)); i++ {
        if ddTotalNumberSwapping >= table.siftMaxSwap {
            break; }
        x = table.perm[var[i]];
        if x < lower || x > upper { continue; }
        /* Only sift if not in symmetry group already. */
        if table.subtables[x].next == uint32( x) {
#ifdef DD_STATS
            previousSize = table->keys - table->isolated;
#endif
            result = ddSymmSiftingAux(table,x,lower,upper);
            if !(result != 0) { goto ddSymmSiftingConvOutOfMem; }
#ifdef DD_STATS
            if (table->keys < (unsigned) previousSize + table->isolated) {
                (void) fprintf(table->out,"-");
            } else if (table->keys > (unsigned) previousSize +
                       table->isolated) {
                (void) fprintf(table->out,"+");
            } else {
                (void) fprintf(table->out,"=");
            }
            fflush(table->out);
#endif
        }
    }

    /* Sifting now until convergence. */
    for uint32( initialSize) > table.keys - table.isolated {
        initialSize = table.keys - table.isolated;
#ifdef DD_STATS
        (void) fprintf(table->out,"\n");
#endif
        /* Here we consider only one representative for each symmetry class. */
        for x, classes = lower, 0; x <= upper; x, classes = x+1, classes+1 {
            for uint32( x) < table.subtables[x].next {
                x = table.subtables[x].next;
            }
            /* Here x is the largest index in a group.
            ** Groups consist of adjacent variables.
            ** Hence, the next increment of x will move it to a new group.
            */
            i = table.invperm[x];
            entry[i] = table.subtables[x].keys;
            var[classes] = i;
        }

        qsort((void *)var,(size_t)classes,sizeof(int),(DD_QSFP)ddSymmUniqueCompare);

        /* Now sift. */
        // inlining of macro ddMin
        for i = 0; i < (((classes) < (table.siftMaxVar)) ? (classes) : (table.siftMaxVar)); i++ {
            if ddTotalNumberSwapping >= table.siftMaxSwap {
                break; }
            x = table.perm[var[i]];
            if uint32( x) >= table.subtables[x].next {
#ifdef DD_STATS
                previousSize = table->keys - table->isolated;
#endif
                result = ddSymmSiftingConvAux(table,x,lower,upper);
                if !(result != 0) { goto ddSymmSiftingConvOutOfMem; }
#ifdef DD_STATS
                if (table->keys < (unsigned) previousSize + table->isolated) {
                    (void) fprintf(table->out,"-");
                } else if (table->keys > (unsigned) previousSize +
                           table->isolated) {
                    (void) fprintf(table->out,"+");
                } else {
                    (void) fprintf(table->out,"=");
                }
                fflush(table->out);
#endif
            }
        } /* for */
    }

    ddSymmSummary(table, lower, upper, &symvars, &symgroups);

#ifdef DD_STATS
    (void) fprintf(table->out, "\n#:S_SIFTING %8d: symmetric variables\n",
                   symvars);
    (void) fprintf(table->out, "#:G_SIFTING %8d: symmetric groups",
                   symgroups);
#endif

    ABC_FREE(var);
    ABC_FREE(entry);

    return(1+symvars);

ddSymmSiftingConvOutOfMem:

    if entry != nil { ABC_FREE(entry); }
    if var != nil { ABC_FREE(var); }

    return(0);

} /* end of cuddSymmSiftingConv */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Comparison function used by qsort.]

  Description [Comparison function used by qsort to order the variables
  according to the number of keys in the subtables.
  Returns the difference in number of keys between the two
  variables being compared.]

  SideEffects [None]

******************************************************************************/
func ddSymmUniqueCompare(ptrX *int32, ptrY *int32) int32 {
     
                                       
                                  
     
      
    return(entry[*ptrY] - entry[*ptrX]);

} /* end of ddSymmUniqueCompare */


/**Function********************************************************************

  Synopsis    [Given xLow <= x <= xHigh moves x up and down between the
  boundaries.]

  Description [Given xLow <= x <= xHigh moves x up and down between the
  boundaries. Finds the best position and does the required changes.
  Assumes that x is not part of a symmetry group. Returns 1 if
  successful; 0 otherwise.]

  SideEffects [None]

******************************************************************************/
func ddSymmSiftingAux(table *DdManager, x int32, xLow int32, xHigh int32) int32 {
    var move *Move;
    var moveUp *Move;       /* list of up moves */
    var moveDown *Move;     /* list of down moves */
    var initialSize int32;
    var result int32;
    var i int32;
    var topbot int32;        /* index to either top or bottom of symmetry group */
    var initGroupSize, finalGroupSize int32;


#ifdef DD_DEBUG
    /* check for previously detected symmetry */
    assert(table->subtables[x].next == (unsigned) x);
#endif

    initialSize = table.keys - table.isolated;

    moveDown = nil;
    moveUp = nil;

    if (x - xLow) > (xHigh - x) {
        /* Will go down first, unless x == xHigh:
        ** Look for consecutive symmetries above x.
        */
        for i = x; i > xLow; i-- {
            if !(cuddSymmCheck(table,i-1,i) != 0) {
                break; }
            topbot = table.subtables[i-1].next; /* find top of i-1's group */
            table.subtables[i-1].next = i;
            table.subtables[x].next = topbot; /* x is bottom of group so its */
                                               /* next is top of i-1's group */
            i = topbot + 1; /* add 1 for i--; new i is top of symm group */
        }
    } else {
        /* Will go up first unless x == xlow:
        ** Look for consecutive symmetries below x.
        */
        for i = x; i < xHigh; i++ {
            if !(cuddSymmCheck(table,i,i+1) != 0) {
                break; }
            /* find bottom of i+1's symm group */
            topbot = i + 1;
            for uint32( topbot) < table.subtables[topbot].next {
                topbot = table.subtables[topbot].next;
            }
            table.subtables[topbot].next = table.subtables[i].next;
            table.subtables[i].next = i + 1;
            i = topbot - 1; /* subtract 1 for i++; new i is bottom of group */
        }
    }

    /* Now x may be in the middle of a symmetry group.
    ** Find bottom of x's symm group.
    */
    for uint32( x) < table.subtables[x].next {
        x = table.subtables[x].next;
    }

    if x == xLow { /* Sift down */

#ifdef DD_DEBUG
        /* x must be a singleton */
        assert((unsigned) x == table->subtables[x].next);
#endif
        if x == xHigh { return(1); }      /* just one variable */

        initGroupSize = 1;

        moveDown = ddSymmSiftingDown(table,x,xHigh);
            /* after this point x --> xHigh, unless early term */
        if moveDown == MV_OOM { goto ddSymmSiftingAuxOutOfMem; }
        if moveDown == nil { return(1); }

        x = moveDown.y;
        /* Find bottom of x's group */
        i = x;
        for uint32( i) < table.subtables[i].next {
            i = table.subtables[i].next;
        }
#ifdef DD_DEBUG
        /* x should be the top of the symmetry group and i the bottom */
        assert((unsigned) i >= table->subtables[i].next);
        assert((unsigned) x == table->subtables[i].next);
#endif
        finalGroupSize = i - x + 1;

        if initGroupSize == finalGroupSize {
            /* No new symmetry groups detected, return to best position */
            result = ddSymmSiftingBackward(table,moveDown,initialSize);
        } else {
            initialSize = table.keys - table.isolated;
            moveUp = ddSymmSiftingUp(table,x,xLow);
            result = ddSymmSiftingBackward(table,moveUp,initialSize);
        }
        if !(result != 0) { goto ddSymmSiftingAuxOutOfMem; }

    } else if cuddNextHigh(table,x) > xHigh { /* Sift up */
        /* Find top of x's symm group */
        i = x;                          /* bottom */
        x = table.subtables[x].next;   /* top */

        if x == xLow { return(1); } /* just one big group */

        initGroupSize = i - x + 1;

        moveUp = ddSymmSiftingUp(table,x,xLow);
            /* after this point x --> xLow, unless early term */
        if moveUp == MV_OOM { goto ddSymmSiftingAuxOutOfMem; }
        if moveUp == nil { return(1); }

        x = moveUp.x;
        /* Find top of x's group */
        i = table.subtables[x].next;
#ifdef DD_DEBUG
        /* x should be the bottom of the symmetry group and i the top */
        assert((unsigned) x >= table->subtables[x].next);
        assert((unsigned) i == table->subtables[x].next);
#endif
        finalGroupSize = x - i + 1;

        if initGroupSize == finalGroupSize {
            /* No new symmetry groups detected, return to best position */
            result = ddSymmSiftingBackward(table,moveUp,initialSize);
        } else {
            initialSize = table.keys - table.isolated;
            moveDown = ddSymmSiftingDown(table,x,xHigh);
            result = ddSymmSiftingBackward(table,moveDown,initialSize);
        }
        if !(result != 0) { goto ddSymmSiftingAuxOutOfMem; }

    } else if (x - xLow) > (xHigh - x) { /* must go down first: shorter */

        moveDown = ddSymmSiftingDown(table,x,xHigh);
        /* at this point x == xHigh, unless early term */
        if moveDown == MV_OOM { goto ddSymmSiftingAuxOutOfMem; }

        if moveDown != nil {
            x = moveDown.y;    /* x is top here */
            i = x;
            for uint32( i) < table.subtables[i].next {
                i = table.subtables[i].next;
            }
        } else {
            i = x;
            for uint32( i) < table.subtables[i].next {
                i = table.subtables[i].next;
            }
            x = table.subtables[i].next;
        }
#ifdef DD_DEBUG
        /* x should be the top of the symmetry group and i the bottom */
        assert((unsigned) i >= table->subtables[i].next);
        assert((unsigned) x == table->subtables[i].next);
#endif
        initGroupSize = i - x + 1;

        moveUp = ddSymmSiftingUp(table,x,xLow);
        if moveUp == MV_OOM { goto ddSymmSiftingAuxOutOfMem; }

        if moveUp != nil {
            x = moveUp.x;
            i = table.subtables[x].next;
        } else {
            i = x;
            for uint32( x) < table.subtables[x].next {
                x = table.subtables[x].next;
            }
        }
#ifdef DD_DEBUG
        /* x should be the bottom of the symmetry group and i the top */
        assert((unsigned) x >= table->subtables[x].next);
        assert((unsigned) i == table->subtables[x].next);
#endif
        finalGroupSize = x - i + 1;

        if initGroupSize == finalGroupSize {
            /* No new symmetry groups detected, return to best position */
            result = ddSymmSiftingBackward(table,moveUp,initialSize);
        } else {
            for moveDown != nil {
                move = moveDown.next;
                // inlining of macro cuddDeallocMove
                ((DdNode *)(moveDown)).ref = 0;     ((DdNode *)(moveDown)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveDown);;
                moveDown = move;
            }
            initialSize = table.keys - table.isolated;
            moveDown = ddSymmSiftingDown(table,x,xHigh);
            result = ddSymmSiftingBackward(table,moveDown,initialSize);
        }
        if !(result != 0) { goto ddSymmSiftingAuxOutOfMem; }

    } else { /* moving up first: shorter */
        /* Find top of x's symmetry group */
        x = table.subtables[x].next;

        moveUp = ddSymmSiftingUp(table,x,xLow);
        /* at this point x == xHigh, unless early term */
        if moveUp == MV_OOM { goto ddSymmSiftingAuxOutOfMem; }

        if moveUp != nil {
            x = moveUp.x;
            i = table.subtables[x].next;
        } else {
            for uint32( x) < table.subtables[x].next {
                x = table.subtables[x].next;
            }
            i = table.subtables[x].next;
        }
#ifdef DD_DEBUG
        /* x is bottom of the symmetry group and i is top */
        assert((unsigned) x >= table->subtables[x].next);
        assert((unsigned) i == table->subtables[x].next);
#endif
        initGroupSize = x - i + 1;

        moveDown = ddSymmSiftingDown(table,x,xHigh);
        if moveDown == MV_OOM { goto ddSymmSiftingAuxOutOfMem; }

        if moveDown != nil {
            x = moveDown.y;
            i = x;
            for uint32( i) < table.subtables[i].next {
                i = table.subtables[i].next;
            }
        } else {
            i = x;
            x = table.subtables[x].next;
        }
#ifdef DD_DEBUG
        /* x should be the top of the symmetry group and i the bottom */
        assert((unsigned) i >= table->subtables[i].next);
        assert((unsigned) x == table->subtables[i].next);
#endif
        finalGroupSize = i - x + 1;

        if initGroupSize == finalGroupSize {
            /* No new symmetries detected, go back to best position */
            result = ddSymmSiftingBackward(table,moveDown,initialSize);
        } else {
            for moveUp != nil {
                move = moveUp.next;
                // inlining of macro cuddDeallocMove
                ((DdNode *)(moveUp)).ref = 0;     ((DdNode *)(moveUp)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveUp);;
                moveUp = move;
            }
            initialSize = table.keys - table.isolated;
            moveUp = ddSymmSiftingUp(table,x,xLow);
            result = ddSymmSiftingBackward(table,moveUp,initialSize);
        }
        if !(result != 0) { goto ddSymmSiftingAuxOutOfMem; }
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

ddSymmSiftingAuxOutOfMem:
    if moveDown != MV_OOM {
        for moveDown != nil {
            move = moveDown.next;
            // inlining of macro cuddDeallocMove
            ((DdNode *)(moveDown)).ref = 0;     ((DdNode *)(moveDown)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveDown);;
            moveDown = move;
        }
    }
    if moveUp != MV_OOM {
        for moveUp != nil {
            move = moveUp.next;
            // inlining of macro cuddDeallocMove
            ((DdNode *)(moveUp)).ref = 0;     ((DdNode *)(moveUp)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveUp);;
            moveUp = move;
        }
    }

    return(0);

} /* end of ddSymmSiftingAux */


/**Function********************************************************************

  Synopsis    [Given xLow <= x <= xHigh moves x up and down between the
  boundaries.]

  Description [Given xLow <= x <= xHigh moves x up and down between the
  boundaries. Finds the best position and does the required changes.
  Assumes that x is either an isolated variable, or it is the bottom of
  a symmetry group. All symmetries may not have been found, because of
  exceeded growth limit. Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

******************************************************************************/
func ddSymmSiftingConvAux(table *DdManager, x int32, xLow int32, xHigh int32) int32 {
    var move *Move;
    var moveUp *Move;       /* list of up moves */
    var moveDown *Move;     /* list of down moves */
    var initialSize int32;
    var result int32;
    var i int32;
    var initGroupSize, finalGroupSize int32;


    initialSize = table.keys - table.isolated;

    moveDown = nil;
    moveUp = nil;

    if x == xLow { /* Sift down */
#ifdef DD_DEBUG
        /* x is bottom of symmetry group */
        assert((unsigned) x >= table->subtables[x].next);
#endif
        i = table.subtables[x].next;
        initGroupSize = x - i + 1;

        moveDown = ddSymmSiftingDown(table,x,xHigh);
        /* at this point x == xHigh, unless early term */
        if moveDown == MV_OOM { goto ddSymmSiftingConvAuxOutOfMem; }
        if moveDown == nil { return(1); }

        x = moveDown.y;
        i = x;
        for uint32( i) < table.subtables[i].next {
            i = table.subtables[i].next;
        }
#ifdef DD_DEBUG
        /* x should be the top of the symmetric group and i the bottom */
        assert((unsigned) i >= table->subtables[i].next);
        assert((unsigned) x == table->subtables[i].next);
#endif
        finalGroupSize = i - x + 1;

        if initGroupSize == finalGroupSize {
            /* No new symmetries detected, go back to best position */
            result = ddSymmSiftingBackward(table,moveDown,initialSize);
        } else {
            initialSize = table.keys - table.isolated;
            moveUp = ddSymmSiftingUp(table,x,xLow);
            result = ddSymmSiftingBackward(table,moveUp,initialSize);
        }
        if !(result != 0) { goto ddSymmSiftingConvAuxOutOfMem; }

    } else if cuddNextHigh(table,x) > xHigh { /* Sift up */
        /* Find top of x's symm group */
        for uint32( x) < table.subtables[x].next {
            x = table.subtables[x].next;
        }
        i = x;                          /* bottom */
        x = table.subtables[x].next;   /* top */

        if x == xLow { return(1); }

        initGroupSize = i - x + 1;

        moveUp = ddSymmSiftingUp(table,x,xLow);
            /* at this point x == xLow, unless early term */
        if moveUp == MV_OOM { goto ddSymmSiftingConvAuxOutOfMem; }
        if moveUp == nil { return(1); }

        x = moveUp.x;
        i = table.subtables[x].next;
#ifdef DD_DEBUG
        /* x should be the bottom of the symmetry group and i the top */
        assert((unsigned) x >= table->subtables[x].next);
        assert((unsigned) i == table->subtables[x].next);
#endif
        finalGroupSize = x - i + 1;

        if initGroupSize == finalGroupSize {
            /* No new symmetry groups detected, return to best position */
            result = ddSymmSiftingBackward(table,moveUp,initialSize);
        } else {
            initialSize = table.keys - table.isolated;
            moveDown = ddSymmSiftingDown(table,x,xHigh);
            result = ddSymmSiftingBackward(table,moveDown,initialSize);
        }
        if !(result != 0) {
            goto ddSymmSiftingConvAuxOutOfMem; }

    } else if (x - xLow) > (xHigh - x) { /* must go down first: shorter */
        moveDown = ddSymmSiftingDown(table,x,xHigh);
            /* at this point x == xHigh, unless early term */
        if moveDown == MV_OOM { goto ddSymmSiftingConvAuxOutOfMem; }

        if moveDown != nil {
            x = moveDown.y;
            i = x;
            for uint32( i) < table.subtables[i].next {
                i = table.subtables[i].next;
            }
        } else {
            for uint32( x) < table.subtables[x].next {
                x = table.subtables[x].next;
            }
            i = x;
            x = table.subtables[x].next;
        }
#ifdef DD_DEBUG
        /* x should be the top of the symmetry group and i the bottom */
        assert((unsigned) i >= table->subtables[i].next);
        assert((unsigned) x == table->subtables[i].next);
#endif
        initGroupSize = i - x + 1;

        moveUp = ddSymmSiftingUp(table,x,xLow);
        if moveUp == MV_OOM { goto ddSymmSiftingConvAuxOutOfMem; }

        if moveUp != nil {
            x = moveUp.x;
            i = table.subtables[x].next;
        } else {
            i = x;
            for uint32( x) < table.subtables[x].next {
                x = table.subtables[x].next;
            }
        }
#ifdef DD_DEBUG
        /* x should be the bottom of the symmetry group and i the top */
        assert((unsigned) x >= table->subtables[x].next);
        assert((unsigned) i == table->subtables[x].next);
#endif
        finalGroupSize = x - i + 1;

        if initGroupSize == finalGroupSize {
            /* No new symmetry groups detected, return to best position */
            result = ddSymmSiftingBackward(table,moveUp,initialSize);
        } else {
            for moveDown != nil {
                move = moveDown.next;
                // inlining of macro cuddDeallocMove
                ((DdNode *)(moveDown)).ref = 0;     ((DdNode *)(moveDown)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveDown);;
                moveDown = move;
            }
            initialSize = table.keys - table.isolated;
            moveDown = ddSymmSiftingDown(table,x,xHigh);
            result = ddSymmSiftingBackward(table,moveDown,initialSize);
        }
        if !(result != 0) { goto ddSymmSiftingConvAuxOutOfMem; }

    } else { /* moving up first: shorter */
        /* Find top of x's symmetry group */
        x = table.subtables[x].next;

        moveUp = ddSymmSiftingUp(table,x,xLow);
        /* at this point x == xHigh, unless early term */
        if moveUp == MV_OOM { goto ddSymmSiftingConvAuxOutOfMem; }

        if moveUp != nil {
            x = moveUp.x;
            i = table.subtables[x].next;
        } else {
            i = x;
            for uint32( x) < table.subtables[x].next {
                x = table.subtables[x].next;
            }
        }
#ifdef DD_DEBUG
        /* x is bottom of the symmetry group and i is top */
        assert((unsigned) x >= table->subtables[x].next);
        assert((unsigned) i == table->subtables[x].next);
#endif
        initGroupSize = x - i + 1;

        moveDown = ddSymmSiftingDown(table,x,xHigh);
        if moveDown == MV_OOM { goto ddSymmSiftingConvAuxOutOfMem; }

        if moveDown != nil {
            x = moveDown.y;
            i = x;
            for uint32( i) < table.subtables[i].next {
                i = table.subtables[i].next;
            }
        } else {
            i = x;
            x = table.subtables[x].next;
        }
#ifdef DD_DEBUG
        /* x should be the top of the symmetry group and i the bottom */
        assert((unsigned) i >= table->subtables[i].next);
        assert((unsigned) x == table->subtables[i].next);
#endif
        finalGroupSize = i - x + 1;

        if initGroupSize == finalGroupSize {
            /* No new symmetries detected, go back to best position */
            result = ddSymmSiftingBackward(table,moveDown,initialSize);
        } else {
            for moveUp != nil {
                move = moveUp.next;
                // inlining of macro cuddDeallocMove
                ((DdNode *)(moveUp)).ref = 0;     ((DdNode *)(moveUp)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveUp);;
                moveUp = move;
            }
            initialSize = table.keys - table.isolated;
            moveUp = ddSymmSiftingUp(table,x,xLow);
            result = ddSymmSiftingBackward(table,moveUp,initialSize);
        }
        if !(result != 0) { goto ddSymmSiftingConvAuxOutOfMem; }
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

ddSymmSiftingConvAuxOutOfMem:
    if moveDown != MV_OOM {
        for moveDown != nil {
            move = moveDown.next;
            // inlining of macro cuddDeallocMove
            ((DdNode *)(moveDown)).ref = 0;     ((DdNode *)(moveDown)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveDown);;
            moveDown = move;
        }
    }
    if moveUp != MV_OOM {
        for moveUp != nil {
            move = moveUp.next;
            // inlining of macro cuddDeallocMove
            ((DdNode *)(moveUp)).ref = 0;     ((DdNode *)(moveUp)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moveUp);;
            moveUp = move;
        }
    }

    return(0);

} /* end of ddSymmSiftingConvAux */


/**Function********************************************************************

  Synopsis    [Moves x up until either it reaches the bound (xLow) or
  the size of the DD heap increases too much.]

  Description [Moves x up until either it reaches the bound (xLow) or
  the size of the DD heap increases too much. Assumes that x is the top
  of a symmetry group.  Checks x for symmetry to the adjacent
  variables. If symmetry is found, the symmetry group of x is merged
  with the symmetry group of the other variable. Returns the set of
  moves in case of success; MV_OOM if memory is full.]

  SideEffects [None]

******************************************************************************/
func ddSymmSiftingUp(table *DdManager, y int32, xLow int32) *Move {
    var moves *Move;
    var move *Move;
    var x int32;
    var size int32;
    var i int32;
    var gxtop, gybot int32;
    var limitSize int32;
    var xindex, yindex int32;
    var zindex int32;
    var z int32;
    var isolated int32;
    var L int32;     /* lower bound on DD size */
#ifdef DD_DEBUG
    int  checkL;
#endif


    moves = nil;
    yindex = table.invperm[y];

    /* Initialize the lower bound.
    ** The part of the DD below the bottom of y' group will not change.
    ** The part of the DD above y that does not interact with y will not
    ** change. The rest may vanish in the best case, except for
    ** the nodes at level xLow, which will not vanish, regardless.
    */
    limitSize = L = table.keys - table.isolated;
    gybot = y;
    for uint32( gybot) < table.subtables[gybot].next {
        gybot = table.subtables[gybot].next;
    }
    for z = xLow + 1; z <= gybot; z++ {
        zindex = table.invperm[z];
        if zindex == yindex || cuddTestInteract(table,zindex,yindex) {
            isolated = table.vars[zindex].ref == 1;
            L -= table.subtables[z].keys - isolated;
        }
    }

    x = cuddNextLow(table,y);
    for x >= xLow && L <= limitSize {
#ifdef DD_DEBUG
        gybot = y;
        while ((unsigned) gybot < table->subtables[gybot].next)
            gybot = table->subtables[gybot].next;
        checkL = table->keys - table->isolated;
        for (z = xLow + 1; z <= gybot; z++) {
            zindex = table->invperm[z];
            if (zindex == yindex || cuddTestInteract(table,zindex,yindex)) {
                isolated = table->vars[zindex]->ref == 1;
                checkL -= table->subtables[z].keys - isolated;
            }
        }
        assert(L == checkL);
#endif
        gxtop = table.subtables[x].next;
        if cuddSymmCheck(table,x,y) != 0 {
            /* Symmetry found, attach symm groups */
            table.subtables[x].next = y;
            i = table.subtables[y].next;
            for table.subtables[i].next != uint32( y) {
                i = table.subtables[i].next;
            }
            table.subtables[i].next = gxtop;
        } else if table.subtables[x].next == uint32( x) &&
                   table.subtables[y].next == uint32( y) {
            /* x and y have self symmetry */
            xindex = table.invperm[x];
            size = cuddSwapInPlace(table,x,y);
#ifdef DD_DEBUG
            assert(table->subtables[x].next == (unsigned) x);
            assert(table->subtables[y].next == (unsigned) y);
#endif
            if size == 0 { goto ddSymmSiftingUpOutOfMem; }
            /* Update the lower bound. */
            if cuddTestInteract(table,xindex,yindex) {
                isolated = table.vars[xindex].ref == 1;
                L += table.subtables[y].keys - isolated;
            }
            move = (Move *) cuddDynamicAllocNode(table);
            if move == nil { goto ddSymmSiftingUpOutOfMem; }
            move.x = x;
            move.y = y;
            move.size = size;
            move.next = moves;
            moves = move;
            if float64( size) > float64( limitSize) * table.maxGrowth {
                return(moves); }
            if size < limitSize { limitSize = size; }
        } else { /* Group move */
            size = ddSymmGroupMove(table,x,y,&moves);
            if size == 0 { goto ddSymmSiftingUpOutOfMem; }
            /* Update the lower bound. */
            z = moves.y;
            for {
                zindex = table.invperm[z];
                if cuddTestInteract(table,zindex,yindex) {
                    isolated = table.vars[zindex].ref == 1;
                    L += table.subtables[z].keys - isolated;
                }
                z = table.subtables[z].next;
            
if !(z != int32( moves.y)) { break }
} 
            if float64( size) > float64( limitSize) * table.maxGrowth {
                return(moves); }
            if size < limitSize { limitSize = size; }
        }
        y = gxtop;
        x = cuddNextLow(table,y);
    }

    return(moves);

ddSymmSiftingUpOutOfMem:
    for moves != nil {
        move = moves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moves)).ref = 0;     ((DdNode *)(moves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moves);;
        moves = move;
    }
    return(MV_OOM);

} /* end of ddSymmSiftingUp */


/**Function********************************************************************

  Synopsis    [Moves x down until either it reaches the bound (xHigh) or
  the size of the DD heap increases too much.]

  Description [Moves x down until either it reaches the bound (xHigh)
  or the size of the DD heap increases too much. Assumes that x is the
  bottom of a symmetry group. Checks x for symmetry to the adjacent
  variables. If symmetry is found, the symmetry group of x is merged
  with the symmetry group of the other variable. Returns the set of
  moves in case of success; MV_OOM if memory is full.]

  SideEffects [None]

******************************************************************************/
func ddSymmSiftingDown(table *DdManager, x int32, xHigh int32) *Move {
    var moves *Move;
    var move *Move;
    var y int32;
    var size int32;
    var limitSize int32;
    var gxtop, gybot int32;
    var R int32;     /* upper bound on node decrease */
    var xindex, yindex int32;
    var isolated int32;
    var z int32;
    var zindex int32;
#ifdef DD_DEBUG
    int  checkR;
#endif

    moves = nil;
    /* Initialize R */
    xindex = table.invperm[x];
    gxtop = table.subtables[x].next;
    limitSize = size = table.keys - table.isolated;
    R = 0;
    for z = xHigh; z > gxtop; z-- {
        zindex = table.invperm[z];
        if zindex == xindex || cuddTestInteract(table,xindex,zindex) {
            isolated = table.vars[zindex].ref == 1;
            R += table.subtables[z].keys - isolated;
        }
    }

    y = cuddNextHigh(table,x);
    for y <= xHigh && size - R < limitSize {
#ifdef DD_DEBUG
        gxtop = table->subtables[x].next;
        checkR = 0;
        for (z = xHigh; z > gxtop; z--) {
            zindex = table->invperm[z];
            if (zindex == xindex || cuddTestInteract(table,xindex,zindex)) {
                isolated = table->vars[zindex]->ref == 1;
                checkR += table->subtables[z].keys - isolated;
            }
        }
        assert(R == checkR);
#endif
        gybot = table.subtables[y].next;
        for table.subtables[gybot].next != uint32( y) {
            gybot = table.subtables[gybot].next;
        }
        if cuddSymmCheck(table,x,y) != 0 {
            /* Symmetry found, attach symm groups */
            gxtop = table.subtables[x].next;
            table.subtables[x].next = y;
            table.subtables[gybot].next = gxtop;
        } else if table.subtables[x].next == uint32( x) &&
                   table.subtables[y].next == uint32( y) {
            /* x and y have self symmetry */
            /* Update upper bound on node decrease. */
            yindex = table.invperm[y];
            if cuddTestInteract(table,xindex,yindex) {
                isolated = table.vars[yindex].ref == 1;
                R -= table.subtables[y].keys - isolated;
            }
            size = cuddSwapInPlace(table,x,y);
#ifdef DD_DEBUG
            assert(table->subtables[x].next == (unsigned) x);
            assert(table->subtables[y].next == (unsigned) y);
#endif
            if size == 0 { goto ddSymmSiftingDownOutOfMem; }
            move = (Move *) cuddDynamicAllocNode(table);
            if move == nil { goto ddSymmSiftingDownOutOfMem; }
            move.x = x;
            move.y = y;
            move.size = size;
            move.next = moves;
            moves = move;
            if float64( size) > float64( limitSize) * table.maxGrowth {
                return(moves); }
            if size < limitSize { limitSize = size; }
        } else { /* Group move */
            /* Update upper bound on node decrease: first phase. */
            gxtop = table.subtables[x].next;
            z = gxtop + 1;
            for {
                zindex = table.invperm[z];
                if zindex == xindex || cuddTestInteract(table,xindex,zindex) {
                    isolated = table.vars[zindex].ref == 1;
                    R -= table.subtables[z].keys - isolated;
                }
                z++;
            
if !(z <= gybot) { break }
} 
            size = ddSymmGroupMove(table,x,y,&moves);
            if size == 0 { goto ddSymmSiftingDownOutOfMem; }
            if float64( size) > float64( limitSize) * table.maxGrowth {
                return(moves); }
            if size < limitSize { limitSize = size; }
            /* Update upper bound on node decrease: second phase. */
            gxtop = table.subtables[gybot].next;
            for z = gxtop + 1; z <= gybot; z++ {
                zindex = table.invperm[z];
                if zindex == xindex || cuddTestInteract(table,xindex,zindex) {
                    isolated = table.vars[zindex].ref == 1;
                    R += table.subtables[z].keys - isolated;
                }
            }
        }
        x = gybot;
        y = cuddNextHigh(table,x);
    }

    return(moves);

ddSymmSiftingDownOutOfMem:
    for moves != nil {
        move = moves.next;
        // inlining of macro cuddDeallocMove
        ((DdNode *)(moves)).ref = 0;     ((DdNode *)(moves)).next = (table).nextFree;     (table).nextFree = (DdNode *)(moves);;
        moves = move;
    }
    return(MV_OOM);

} /* end of ddSymmSiftingDown */


/**Function********************************************************************

  Synopsis    [Swaps two groups.]

  Description [Swaps two groups. x is assumed to be the bottom variable
  of the first group. y is assumed to be the top variable of the second
  group.  Updates the list of moves. Returns the number of keys in the
  table if successful; 0 otherwise.]

  SideEffects [None]

******************************************************************************/
func ddSymmGroupMove(table *DdManager, x int32, y int32, moves **Move) int32 {
    var move *Move;
    var size int32  = -1;
    var i, j int32;
    var xtop, xbot, xsize, ytop, ybot, ysize, newxtop int32;
    var swapx int32 = -1;
    var swapy int32 = -1;

#ifdef DD_DEBUG
    assert(x < y);      /* we assume that x < y */
#endif
    /* Find top, bottom, and size for the two groups. */
    xbot = x;
    xtop = table.subtables[x].next;
    xsize = xbot - xtop + 1;
    ybot = y;
    for uint32( ybot) < table.subtables[ybot].next {
        ybot = table.subtables[ybot].next;
    }
    ytop = y;
    ysize = ybot - ytop + 1;

    /* Sift the variables of the second group up through the first group. */
    for i = 1; i <= ysize; i++ {
        for j = 1; j <= xsize; j++ {
            size = cuddSwapInPlace(table,x,y);
            if size == 0 { return(0); }
            swapx = x; swapy = y;
            y = x;
            x = y - 1;
        }
        y = ytop + i;
        x = y - 1;
    }

    /* fix symmetries */
    y = xtop; /* ytop is now where xtop used to be */
    for i = 0; i < ysize-1 ; i++ {
        table.subtables[y].next = y + 1;
        y = y + 1;
    }
    table.subtables[y].next = xtop; /* y is bottom of its group, join */
                                     /* its symmetry to top of its group */
    x = y + 1;
    newxtop = x;
    for i = 0; i < xsize - 1 ; i++ {
        table.subtables[x].next = x + 1;
        x = x + 1;
    }
    table.subtables[x].next = newxtop; /* x is bottom of its group, join */
                                        /* its symmetry to top of its group */
    /* Store group move */
    move = (Move *) cuddDynamicAllocNode(table);
    if move == nil { return(0); }
    move.x = swapx;
    move.y = swapy;
    move.size = size;
    move.next = *moves;
    *moves = move;

    return(size);

} /* end of ddSymmGroupMove */


/**Function********************************************************************

  Synopsis    [Undoes the swap of two groups.]

  Description [Undoes the swap of two groups. x is assumed to be the
  bottom variable of the first group. y is assumed to be the top
  variable of the second group.  Returns the number of keys in the table
  if successful; 0 otherwise.]

  SideEffects [None]

******************************************************************************/
func ddSymmGroupMoveBackward(table *DdManager, x int32, y int32) int32 {
    var size int32  = -1;
    var i, j int32;
    var xtop, xbot, xsize, ytop, ybot, ysize, newxtop int32;

#ifdef DD_DEBUG
    assert(x < y); /* We assume that x < y */
#endif

    /* Find top, bottom, and size for the two groups. */
    xbot = x;
    xtop = table.subtables[x].next;
    xsize = xbot - xtop + 1;
    ybot = y;
    for uint32( ybot) < table.subtables[ybot].next {
        ybot = table.subtables[ybot].next;
    }
    ytop = y;
    ysize = ybot - ytop + 1;

    /* Sift the variables of the second group up through the first group. */
    for i = 1; i <= ysize; i++ {
        for j = 1; j <= xsize; j++ {
            size = cuddSwapInPlace(table,x,y);
            if size == 0 { return(0); }
            y = x;
            x = cuddNextLow(table,y);
        }
        y = ytop + i;
        x = y - 1;
    }

    /* Fix symmetries. */
    y = xtop;
    for i = 0; i < ysize-1 ; i++ {
        table.subtables[y].next = y + 1;
        y = y + 1;
    }
    table.subtables[y].next = xtop; /* y is bottom of its group, join */
                                     /* its symmetry to top of its group */
    x = y + 1;
    newxtop = x;
    for i = 0; i < xsize-1 ; i++ {
        table.subtables[x].next = x + 1;
        x = x + 1;
    }
    table.subtables[x].next = newxtop; /* x is bottom of its group, join */
                                        /* its symmetry to top of its group */

    return(size);

} /* end of ddSymmGroupMoveBackward */


/**Function********************************************************************

  Synopsis    [Given a set of moves, returns the DD heap to the position
  giving the minimum size.]

  Description [Given a set of moves, returns the DD heap to the
  position giving the minimum size. In case of ties, returns to the
  closest position giving the minimum size. Returns 1 in case of
  success; 0 otherwise.]

  SideEffects [None]

******************************************************************************/
func ddSymmSiftingBackward(table *DdManager, moves *Move, size int32) int32 {
    var move *Move;
    var res int32  = -1;

    for move = moves; move != nil; move = move.next {
        if move.size < size {
            size = move.size;
        }
    }

    for move = moves; move != nil; move = move.next {
        if move.size == size { return(1); }
        if table.subtables[move.x].next == move.x && table.subtables[move.y].next == move.y {
            res = cuddSwapInPlace(table,int32(move.x),int32(move.y));
#ifdef DD_DEBUG
            assert(table->subtables[move->x].next == move->x);
            assert(table->subtables[move->y].next == move->y);
#endif
        } else { /* Group move necessary */
            res = ddSymmGroupMoveBackward(table,int32(move.x),int32(move.y));
        }
        if !(res != 0) { return(0); }
    }

    return(1);

} /* end of ddSymmSiftingBackward */


/**Function********************************************************************

  Synopsis    [Counts numbers of symmetric variables and symmetry
  groups.]

  Description []

  SideEffects [None]

******************************************************************************/
func ddSymmSummary(table *DdManager, lower int32, upper int32, symvars *int32, symgroups *int32) {
    var i, x, gbot int32;
    var TotalSymm int32;
    var TotalSymmGroups int32;

    for i = lower; i <= upper; i++ {
        if table.subtables[i].next != uint32( i) {
            TotalSymmGroups++;
            x = i;
            for {
                TotalSymm++;
                gbot = x;
                x = table.subtables[x].next;
            
if !(x != i) { break }
} 
#ifdef DD_DEBUG
            assert(table->subtables[gbot].next == (unsigned) i);
#endif
            i = gbot;
        }
    }
    *symvars = TotalSymm;
    *symgroups = TotalSymmGroups;

    return;

} /* end of ddSymmSummary */


//ABC_NAMESPACE_IMPL_END


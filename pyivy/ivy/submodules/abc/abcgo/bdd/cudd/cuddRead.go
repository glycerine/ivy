package cudd

/**CFile***********************************************************************

  FileName    [cuddRead.c]

  PackageName [cudd]

  Synopsis    [Functions to read in a matrix]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addRead()
                <li> Cudd_bddRead()
                </ul>]

  SeeAlso     [cudd_addHarwell.c]

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
static char rcsid[] DD_UNUSED = "$Id: cuddRead.c,v 1.6 2004/08/13 18:04:50 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Reads in a sparse matrix.]

  Description [Reads in a sparse matrix specified in a simple format.
  The first line of the input contains the numbers of rows and columns.
  The remaining lines contain the elements of the matrix, one per line.
  Given a background value
  (specified by the background field of the manager), only the values
  different from it are explicitly listed.  Each foreground element is
  described by two integers, i.e., the row and column number, and a
  real number, i.e., the value.<p>
  Cudd_addRead produces an ADD that depends on two sets of variables: x
  and y.  The x variables (x\[0\] ... x\[nx-1\]) encode the row index and
  the y variables (y\[0\] ... y\[ny-1\]) encode the column index.
  x\[0\] and y\[0\] are the most significant bits in the indices.
  The variables may already exist or may be created by the function.
  The index of x\[i\] is bx+i*sx, and the index of y\[i\] is by+i*sy.<p>
  On input, nx and ny hold the numbers
  of row and column variables already in existence. On output, they
  hold the numbers of row and column variables actually used by the
  matrix. When Cudd_addRead creates the variable arrays,
  the index of x\[i\] is bx+i*sx, and the index of y\[i\] is by+i*sy.
  When some variables already exist Cudd_addRead expects the indices
  of the existing x variables to be bx+i*sx, and the indices of the
  existing y variables to be by+i*sy.<p>
  m and n are set to the numbers of rows and columns of the
  matrix.  Their values on input are immaterial.
  The ADD for the
  sparse matrix is returned in E, and its reference count is > 0.
  Cudd_addRead returns 1 in case of success; 0 otherwise.]

  SideEffects [nx and ny are set to the numbers of row and column
  variables. m and n are set to the numbers of rows and columns. x and y
  are possibly extended to represent the array of row and column
  variables. Similarly for xn and yn_, which hold on return from
  Cudd_addRead the complements of the row and column variables.]

  SeeAlso     [Cudd_addHarwell Cudd_bddRead]

******************************************************************************/
int
Cudd_addRead(
  FILE * fp /* input file pointer */,
  DdManager * dd /* DD manager */,
  DdNode ** E /* characteristic function of the graph */,
  DdNode *** x /* array of row variables */,
  DdNode *** y /* array of column variables */,
  DdNode *** xn /* array of complemented row variables */,
  DdNode *** yn_ /* array of complemented column variables */,
  int * nx /* number or row variables */,
  int * ny /* number or column variables */,
  int * m /* number of rows */,
  int * n /* number of columns */,
  int  bx /* first index of row variables */,
  int  sx /* step of row variables */,
  int  by /* first index of column variables */,
  int  sy /* step of column variables */)
{
    var one, zero *DdNode;
    var w, neW *DdNode;
    var minterm1 *DdNode;
    var u, v, err, i, nv int32;
    var lnx, lny int32;
    CUDD_VALUE_TYPE val;
    var lx, ly, lxn, lyn **DdNode;

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro DD_ZERO
    zero = ((dd).zero);

    err = fscanf(fp, "%d %d", &u, &v);
    if err == -1 {
        return(0);
    } else if err != 2 {
        return(0);
    }

    *m = u;
    /* Compute the number of x variables. */
    lx = *x; lxn = *xn;
    u--;        /* row and column numbers start from 0 */
    for lnx=0; u > 0; lnx++ {
        u >>= 1;
    }
    /* Here we rely on the fact that REALLOC of a null pointer is
    ** translates to an ALLOC.
    */
    if lnx > *nx {
        *x = lx = ABC_REALLOC(DdNode *, *x, lnx);
        if lx == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
        *xn = lxn =  ABC_REALLOC(DdNode *, *xn, lnx);
        if lxn == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
    }

    *n = v;
    /* Compute the number of y variables. */
    ly = *y; lyn = *yn_;
    v--;        /* row and column numbers start from 0 */
    for lny=0; v > 0; lny++ {
        v >>= 1;
    }
    /* Here we rely on the fact that REALLOC of a null pointer is
    ** translates to an ALLOC.
    */
    if lny > *ny {
        *y = ly = ABC_REALLOC(DdNode *, *y, lny);
        if ly == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
        *yn_ = lyn =  ABC_REALLOC(DdNode *, *yn_, lny);
        if lyn == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
    }

    /* Create all new variables. */
    for i = *nx, nv = bx + (*nx) * sx; i < lnx; i++, nv += sx {
        for {
            dd.reordered = 0;
            lx[i] = cuddUniqueInter(dd, nv, one, zero);
        
if !(dd.reordered == 1) { break }
} 
        if lx[i] == nil { return(0); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(lx[i]).ref);
        for {
            dd.reordered = 0;
            lxn[i] = cuddUniqueInter(dd, nv, zero, one);
        
if !(dd.reordered == 1) { break }
} 
        if lxn[i] == nil { return(0); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(lxn[i]).ref);
    }
    for i = *ny, nv = by + (*ny) * sy; i < lny; i++, nv += sy {
        for {
            dd.reordered = 0;
            ly[i] = cuddUniqueInter(dd, nv, one, zero);
        
if !(dd.reordered == 1) { break }
} 
        if ly[i] == nil { return(0); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(ly[i]).ref);
        for {
            dd.reordered = 0;
            lyn[i] = cuddUniqueInter(dd, nv, zero, one);
        
if !(dd.reordered == 1) { break }
} 
        if lyn[i] == nil { return(0); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(lyn[i]).ref);
    }
    *nx = lnx;
    *ny = lny;

    *E = dd.background; /* this call will never cause reordering */
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(*E).ref);

    for ! feof(fp) {
        err = fscanf(fp, "%d %d %lf", &u, &v, &val);
        if err == -1 {
            break;
        } else if err != 3 {
            return(0);
        } else if u >= *m || v >= *n || u < 0 || v < 0 {
            return(0);
        }
 
        // inlining of macro cuddRef
        minterm1 = one; cuddSatInc(Cudd_Regular(minterm1).ref);

        /* Build minterm1 corresponding to this arc */
        for i = lnx - 1; i>=0; i-- {
            if u & 1 {
                w = Cudd_addApply(dd, Cudd_addTimes, minterm1, lx[i]);
            } else {
                w = Cudd_addApply(dd, Cudd_addTimes, minterm1, lxn[i]);
            }
            if w == nil {
                Cudd_RecursiveDeref(dd, minterm1);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(w).ref);
            Cudd_RecursiveDeref(dd, minterm1);
            minterm1 = w;
            u >>= 1;
        }
        for i = lny - 1; i>=0; i-- {
            if v & 1 {
                w = Cudd_addApply(dd, Cudd_addTimes, minterm1, ly[i]);
            } else {
                w = Cudd_addApply(dd, Cudd_addTimes, minterm1, lyn[i]);
            }
            if w == nil {
                Cudd_RecursiveDeref(dd, minterm1);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(w).ref);
            Cudd_RecursiveDeref(dd, minterm1);
            minterm1 = w;
            v >>= 1;
        }
        /* Create new constant node if necessary.
        ** This call will never cause reordering.
        */
        neW = cuddUniqueConst(dd, val);
        if neW == nil {
            Cudd_RecursiveDeref(dd, minterm1);
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(neW).ref);

        w = Cudd_addIte(dd, minterm1, neW, *E);
        if w == nil {
            Cudd_RecursiveDeref(dd, minterm1);
            Cudd_RecursiveDeref(dd, neW);
            return(0);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(w).ref);
        Cudd_RecursiveDeref(dd, minterm1);
        Cudd_RecursiveDeref(dd, neW);
        Cudd_RecursiveDeref(dd, *E);
        *E = w;
    }
    return(1);

} /* end of Cudd_addRead */


/**Function********************************************************************

  Synopsis    [Reads in a graph (without labels) given as a list of arcs.]

  Description [Reads in a graph (without labels) given as an adjacency
  matrix.  The first line of the input contains the numbers of rows and
  columns of the adjacency matrix. The remaining lines contain the arcs
  of the graph, one per line. Each arc is described by two integers,
  i.e., the row and column number, or the indices of the two endpoints.
  Cudd_bddRead produces a BDD that depends on two sets of variables: x
  and y.  The x variables (x\[0\] ... x\[nx-1\]) encode
  the row index and the y variables (y\[0\] ... y\[ny-1\]) encode the
  column index. x\[0\] and y\[0\] are the most significant bits in the
  indices.
  The variables may already exist or may be created by the function.
  The index of x\[i\] is bx+i*sx, and the index of y\[i\] is by+i*sy.<p>
  On input, nx and ny hold the numbers of row and column variables already
  in existence. On output, they hold the numbers of row and column
  variables actually used by the matrix. When Cudd_bddRead creates the
  variable arrays, the index of x\[i\] is bx+i*sx, and the index of
  y\[i\] is by+i*sy. When some variables already exist, Cudd_bddRead
  expects the indices of the existing x variables to be bx+i*sx, and the
  indices of the existing y variables to be by+i*sy.<p>
  m and n are set to the numbers of rows and columns of the
  matrix.  Their values on input are immaterial.  The BDD for the graph
  is returned in E, and its reference count is > 0. Cudd_bddRead returns
  1 in case of success; 0 otherwise.]

  SideEffects [nx and ny are set to the numbers of row and column
  variables. m and n are set to the numbers of rows and columns. x and y
  are possibly extended to represent the array of row and column
  variables.]

  SeeAlso     [Cudd_addHarwell Cudd_addRead]

******************************************************************************/
int
Cudd_bddRead(
  FILE * fp /* input file pointer */,
  DdManager * dd /* DD manager */,
  DdNode ** E /* characteristic function of the graph */,
  DdNode *** x /* array of row variables */,
  DdNode *** y /* array of column variables */,
  int * nx /* number or row variables */,
  int * ny /* number or column variables */,
  int * m /* number of rows */,
  int * n /* number of columns */,
  int  bx /* first index of row variables */,
  int  sx /* step of row variables */,
  int  by /* first index of column variables */,
  int  sy /* step of column variables */)
{
    var one, zero *DdNode;
    var w *DdNode;
    var minterm1 *DdNode;
    var u, v, err, i, nv int32;
    var lnx, lny int32;
    var lx, ly **DdNode;

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    err = fscanf(fp, "%d %d", &u, &v);
    if err == -1 {
        return(0);
    } else if err != 2 {
        return(0);
    }

    *m = u;
    /* Compute the number of x variables. */
    lx = *x;
    u--;        /* row and column numbers start from 0 */
    for lnx=0; u > 0; lnx++ {
        u >>= 1;
    }
    if lnx > *nx {
        *x = lx = ABC_REALLOC(DdNode *, *x, lnx);
        if lx == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
    }

    *n = v;
    /* Compute the number of y variables. */
    ly = *y;
    v--;        /* row and column numbers start from 0 */
    for lny=0; v > 0; lny++ {
        v >>= 1;
    }
    if lny > *ny {
        *y = ly = ABC_REALLOC(DdNode *, *y, lny);
        if ly == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
    }

    /* Create all new variables. */
    for i = *nx, nv = bx + (*nx) * sx; i < lnx; i++, nv += sx {
        for {
            dd.reordered = 0;
            lx[i] = cuddUniqueInter(dd, nv, one, zero);
        
if !(dd.reordered == 1) { break }
} 
        if lx[i] == nil { return(0); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(lx[i]).ref);
    }
    for i = *ny, nv = by + (*ny) * sy; i < lny; i++, nv += sy {
        for {
            dd.reordered = 0;
            ly[i] = cuddUniqueInter(dd, nv, one, zero);
        
if !(dd.reordered == 1) { break }
} 
        if ly[i] == nil { return(0); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(ly[i]).ref);
    }
    *nx = lnx;
    *ny = lny;

    *E = zero; /* this call will never cause reordering */
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(*E).ref);

    for ! feof(fp) {
        err = fscanf(fp, "%d %d", &u, &v);
        if err == -1 {
            break;
        } else if err != 2 {
            return(0);
        } else if u >= *m || v >= *n || u < 0 || v < 0 {
            return(0);
        }
 
        // inlining of macro cuddRef
        minterm1 = one; cuddSatInc(Cudd_Regular(minterm1).ref);

        /* Build minterm1 corresponding to this arc. */
        for i = lnx - 1; i>=0; i-- {
            if u & 1 {
                w = Cudd_bddAnd(dd, minterm1, lx[i]);
            } else {
                // inlining of macro Cudd_Not
                w = Cudd_bddAnd(dd, minterm1, ((DdNode *)((ptrint)(lx[i]) ^ 01)));
            }
            if w == nil {
                Cudd_RecursiveDeref(dd, minterm1);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(w).ref);
            Cudd_RecursiveDeref(dd,minterm1);
            minterm1 = w;
            u >>= 1;
        }
        for i = lny - 1; i>=0; i-- {
            if v & 1 {
                w = Cudd_bddAnd(dd, minterm1, ly[i]);
            } else {
                // inlining of macro Cudd_Not
                w = Cudd_bddAnd(dd, minterm1, ((DdNode *)((ptrint)(ly[i]) ^ 01)));
            }
            if w == nil {
                Cudd_RecursiveDeref(dd, minterm1);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(w).ref);
            Cudd_RecursiveDeref(dd, minterm1);
            minterm1 = w;
            v >>= 1;
        }

        // inlining of macro Cudd_Not
        w = Cudd_bddAnd(dd, ((DdNode *)((ptrint)(minterm1) ^ 01)), ((DdNode *)((ptrint)(*E) ^ 01)));
        if w == nil {
            Cudd_RecursiveDeref(dd, minterm1);
            return(0);
        }
        // inlining of macro Cudd_Not
        w = ((DdNode *)((ptrint)(w) ^ 01));
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(w).ref);
        Cudd_RecursiveDeref(dd, minterm1);
        Cudd_RecursiveDeref(dd, *E);
        *E = w;
    }
    return(1);

} /* end of Cudd_bddRead */

/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


//ABC_NAMESPACE_IMPL_END



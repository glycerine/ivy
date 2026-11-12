package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddHarwell.c]

  PackageName [cudd]

  Synopsis    [Function to read a matrix in Harwell format.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addHarwell()
                </ul>
        ]

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
static char rcsid[] DD_UNUSED = "$Id: cuddHarwell.c,v 1.9 2004/08/13 18:04:49 fabio Exp $";
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

  Synopsis    [Reads in a matrix in the format of the Harwell-Boeing
  benchmark suite.]

  Description [Reads in a matrix in the format of the Harwell-Boeing
  benchmark suite. The variables are ordered as follows:
  <blockquote>
  x\[0\] y\[0\] x\[1\] y\[1\] ...
  </blockquote>
  0 is the most significant bit.  On input, nx and ny hold the numbers
  of row and column variables already in existence. On output, they
  hold the numbers of row and column variables actually used by the
  matrix.  m and n are set to the numbers of rows and columns of the
  matrix.  Their values on input are immaterial.  Returns 1 on
  success; 0 otherwise. The ADD for the sparse matrix is returned in
  E, and its reference count is > 0.]

  SideEffects [None]

  SeeAlso     [Cudd_addRead Cudd_bddRead]

******************************************************************************/
int
Cudd_addHarwell(
  FILE * fp /* pointer to the input file */,
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
  int  sy /* step of column variables */,
  int  pr /* verbosity level */)
{
    var one, zero *DdNode;
    var w *DdNode;
    var cubex, cubey, minterm1 *DdNode;
    var u, v, err, i, j, nv int32;
    var val float64;
    var lx **DdNode = nil;
    var ly **DdNode = nil;
    var lxn **DdNode = nil;
    var lyn **DdNode = nil;    /* local copies of x, y, xn, yn_ */
    var lnx, lny int32;                       /* local copies of nx and ny */
    var title [73]byte;
    var key [9]byte;
    var mxtype [4]byte;
    var rhstyp [4]byte;
    var totcrd, ptrcrd, indcrd, valcrd, rhscrd, nrow, ncol, nnzero, neltvl, nrhs, nrhsix int32;
    var colptr, rowind *int32;
     
                       
                         
      

    if *nx < 0 || *ny < 0 { return(0); }

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro DD_ZERO
    zero = ((dd).zero);

    /* Read the header */
    err = fscanf(fp, "%72c %8c", title, key);
    if err == -1 {
        return(0);
    } else if err != 2 {
        return(0);
    }
    title[72] = byte( 0);
    key[8] = byte( 0);

    err = fscanf(fp, "%d %d %d %d %d", &totcrd, &ptrcrd, &indcrd,
        &valcrd, &rhscrd);
    if err == -1 {
        return(0);
    } else if err != 5 {
        return(0);
    }

    err = fscanf(fp, "%3s %d %d %d %d", mxtype, &nrow, &ncol,
        &nnzero, &neltvl);
    if err == -1 {
        return(0);
    } else if err != 5 {
        return(0);
    }

    /* Skip FORTRAN formats */
    if rhscrd == 0 {
        err = fscanf(fp, "%*s %*s %*s \n");
    } else {
        err = fscanf(fp, "%*s %*s %*s %*s \n");
    }
    if err == -1 {
        return(0);
    } else if err != 0 {
        return(0);
    }

    /* Print out some stuff if requested to be verbose */
    if pr>0 {
        (void) fprintf(dd.out,"%s: type %s, %d rows, %d columns, %d entries\n", key,
                mxtype, nrow, ncol, nnzero);
        if pr>1 { (void) fprintf(dd.out,"%s\n", title); }
    }

    /* Check matrix type */
    if mxtype[0] != 'R' || mxtype[1] != 'U' || mxtype[2] != 'A' {
        (void) fprintf(dd.err,"%s: Illegal matrix type: %s\n",
                               key, mxtype);
        return(0);
    }
    if neltvl != 0 { return(0); }

    /* Read optional 5-th line */
    if rhscrd != 0 {
        err = fscanf(fp, "%3c %d %d", rhstyp, &nrhs, &nrhsix);
        if err == -1 {
            return(0);
        } else if err != 3 {
            return(0);
        }
        rhstyp[3] = byte( 0);
        if rhstyp[0] != 'F' {
            (void) fprintf(dd.err,
                        "%s: Sparse right-hand side not yet supported\n", key);
            return(0);
        }
        if pr>0 { (void) fmt.Fprintf(dd.out, "%d right-hand side(s)\n", nrhs); }
    } else {
        nrhs = 0;
    }

    /* Compute the number of variables */

    /* row and column numbers start from 0 */
    u = nrow - 1;
    for i=0; u > 0; i++ {
        u >>= 1;
    }
    lnx = i;
    if nrhs == 0 {
        v = ncol - 1;
    } else {
        // inlining of macro ddMax
        v = 2* ((((nrhs) > (ncol)) ? (nrhs) : (ncol)) - 1);
    }
    for i=0; v > 0; i++ {
        v >>= 1;
    }
    lny = i;

    /* Allocate or reallocate arrays for variables as needed */
    if *nx == 0 {
        if lnx > 0 {
            *x = lx = ABC_ALLOC(DdNode *,lnx);
            if lx == nil {
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
            *xn = lxn =  ABC_ALLOC(DdNode *,lnx);
            if lxn == nil {
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
        } else {
            *x = *xn = nil;
        }
    } else if lnx > *nx {
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
    } else {
        lx = *x;
        lxn = *xn;
    }
    if *ny == 0 {
        if lny >0 {
            *y = ly = ABC_ALLOC(DdNode *,lny);
            if ly == nil {
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
            *yn_ = lyn = ABC_ALLOC(DdNode *,lny);
            if lyn == nil {
                dd.errorCode = CUDD_MEMORY_OUT;
                return(0);
            }
        } else {
            *y = *yn_ = nil;
        }
    } else if lny > *ny {
        *y = ly = ABC_REALLOC(DdNode *, *y, lny);
        if ly == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
        *yn_ = lyn = ABC_REALLOC(DdNode *, *yn_, lny);
        if lyn == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
    } else {
        ly = *y;
        lyn = *yn_;
    }

    /* Create new variables as needed */
    for i= *nx,nv=bx+(*nx)*sx; i < lnx; i++,nv+=sx {
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
    for i= *ny,nv=by+(*ny)*sy; i < lny; i++,nv+=sy {
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

    /* Update matrix parameters */
    *nx = lnx;
    *ny = lny;
    *m = nrow;
    if nrhs == 0 {
        *n = ncol;
    } else {
        *n = (1 << (lny - 1)) + nrhs;
    }
    
    /* Read structure data */
    colptr = ABC_ALLOC(int, ncol+1);
    if colptr == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(0);
    }
    rowind = ABC_ALLOC(int, nnzero);
    if rowind == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(0);
    }

    for i=0; i<ncol+1; i++ {
        err = fscanf(fp, " %d ", &u);
        if err == -1{ 
            ABC_FREE(colptr);
            ABC_FREE(rowind);
            return(0);
        } else if err != 1 {
            ABC_FREE(colptr);
            ABC_FREE(rowind);
            return(0);
        }
        colptr[i] = u - 1;
    }
    if colptr[0] != 0 {
        (void) fprintf(dd.err,"%s: Unexpected colptr[0] (%d)\n",
                               key,colptr[0]);
        ABC_FREE(colptr);
        ABC_FREE(rowind);
        return(0);
    }
    for i=0; i<nnzero; i++ {
        err = fscanf(fp, " %d ", &u);
        if err == -1{ 
            ABC_FREE(colptr);
            ABC_FREE(rowind);
            return(0);
        } else if err != 1 {
            ABC_FREE(colptr);
            ABC_FREE(rowind);
            return(0);
        }
        rowind[i] = u - 1;
    }

    // inlining of macro cuddRef
    *E = zero; cuddSatInc(Cudd_Regular(*E).ref);

    for j=0; j<ncol; j++ {
        v = j;
        // inlining of macro cuddRef
        cubey = one; cuddSatInc(Cudd_Regular(cubey).ref);
        for nv = lny - 1; nv>=0; nv-- {
            if v & 1 {
                w = Cudd_addApply(dd, Cudd_addTimes, cubey, ly[nv]);
            } else {
                w = Cudd_addApply(dd, Cudd_addTimes, cubey, lyn[nv]);
            }
            if w == nil {
                Cudd_RecursiveDeref(dd, cubey);
                ABC_FREE(colptr);
                ABC_FREE(rowind);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(w).ref);
            Cudd_RecursiveDeref(dd, cubey);
            cubey = w;
            v >>= 1;
        }
        for i=colptr[j]; i<colptr[j+1]; i++ {
            u = rowind[i];
            err = fscanf(fp, " %lf ", &val);
            if err == -1 || err != 1{ 
                Cudd_RecursiveDeref(dd, cubey);
                ABC_FREE(colptr);
                ABC_FREE(rowind);
                return(0);
            }
            /* Create new Constant node if necessary */
            cubex = cuddUniqueConst(dd, (CUDD_VALUE_TYPE) val);
            if cubex == nil {
                Cudd_RecursiveDeref(dd, cubey);
                ABC_FREE(colptr);
                ABC_FREE(rowind);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(cubex).ref);

            for nv = lnx - 1; nv>=0; nv-- {
                if u & 1 {
                    w = Cudd_addApply(dd, Cudd_addTimes, cubex, lx[nv]);
                } else { 
                    w = Cudd_addApply(dd, Cudd_addTimes, cubex, lxn[nv]);
                }
                if w == nil {
                    Cudd_RecursiveDeref(dd, cubey);
                    Cudd_RecursiveDeref(dd, cubex);
                    ABC_FREE(colptr);
                    ABC_FREE(rowind);
                    return(0);
                }
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(w).ref);
                Cudd_RecursiveDeref(dd, cubex);
                cubex = w;
                u >>= 1;
            }
            minterm1 = Cudd_addApply(dd, Cudd_addTimes, cubey, cubex);
            if minterm1 == nil {
                Cudd_RecursiveDeref(dd, cubey);
                Cudd_RecursiveDeref(dd, cubex);
                ABC_FREE(colptr);
                ABC_FREE(rowind);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(minterm1).ref);
            Cudd_RecursiveDeref(dd, cubex);
            w = Cudd_addApply(dd, Cudd_addPlus, *E, minterm1);
            if w == nil {
                Cudd_RecursiveDeref(dd, cubey);
                ABC_FREE(colptr);
                ABC_FREE(rowind);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(w).ref);
            Cudd_RecursiveDeref(dd, minterm1);
            Cudd_RecursiveDeref(dd, *E);
            *E = w;
        }
        Cudd_RecursiveDeref(dd, cubey);
    }
    ABC_FREE(colptr);
    ABC_FREE(rowind);

    /* Read right-hand sides */
    for j=0; j<nrhs; j++ {
        v = j + (1<< (lny-1));
        // inlining of macro cuddRef
        cubey = one; cuddSatInc(Cudd_Regular(cubey).ref);
        for nv = lny - 1; nv>=0; nv-- {
            if v & 1 {
                w = Cudd_addApply(dd, Cudd_addTimes, cubey, ly[nv]);
            } else {
                w = Cudd_addApply(dd, Cudd_addTimes, cubey, lyn[nv]);
            }
            if w == nil {
                Cudd_RecursiveDeref(dd, cubey);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(w).ref);
            Cudd_RecursiveDeref(dd, cubey);
            cubey = w;
            v >>= 1;
        }
        for i=0; i<nrow; i++ {
            u = i;
            err = fscanf(fp, " %lf ", &val);
            if err == -1 || err != 1{ 
                Cudd_RecursiveDeref(dd, cubey);
                return(0);
            }
            /* Create new Constant node if necessary */
            if val == float64( 0.0) { continue; }
            cubex = cuddUniqueConst(dd, (CUDD_VALUE_TYPE) val);
            if cubex == nil {
                Cudd_RecursiveDeref(dd, cubey);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(cubex).ref);

            for nv = lnx - 1; nv>=0; nv-- {
                if u & 1 {
                   w = Cudd_addApply(dd, Cudd_addTimes, cubex, lx[nv]);
                } else { 
                    w = Cudd_addApply(dd, Cudd_addTimes, cubex, lxn[nv]);
                }
                if w == nil {
                    Cudd_RecursiveDeref(dd, cubey);
                    Cudd_RecursiveDeref(dd, cubex);
                    return(0);
                }
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(w).ref);
                Cudd_RecursiveDeref(dd, cubex);
                cubex = w;
                u >>= 1;
            }
            minterm1 = Cudd_addApply(dd, Cudd_addTimes, cubey, cubex);
            if minterm1 == nil {
                Cudd_RecursiveDeref(dd, cubey);
                Cudd_RecursiveDeref(dd, cubex);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(minterm1).ref);
            Cudd_RecursiveDeref(dd, cubex);
            w = Cudd_addApply(dd, Cudd_addPlus, *E, minterm1);
            if w == nil {
                Cudd_RecursiveDeref(dd, cubey);
                return(0);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(w).ref);
            Cudd_RecursiveDeref(dd, minterm1);
            Cudd_RecursiveDeref(dd, *E);
            *E = w;
        }
        Cudd_RecursiveDeref(dd, cubey);
    }

    return(1);

} /* end of Cudd_addHarwell */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


//ABC_NAMESPACE_IMPL_END



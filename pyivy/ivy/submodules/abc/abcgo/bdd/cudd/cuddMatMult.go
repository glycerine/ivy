package cudd

/**CFile***********************************************************************

  FileName    [cuddMatMult.c]

  PackageName [cudd]

  Synopsis    [Matrix multiplication functions.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addMatrixMultiply()
                <li> Cudd_addTimesPlus()
                <li> Cudd_addTriangle()
                <li> Cudd_addOuterSum()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> addMMRecur()
                <li> addTriangleRecur()
                <li> cuddAddOuterSumRecur()
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
static char rcsid[] DD_UNUSED = "$Id: cuddMatMult.c,v 1.17 2004/08/13 18:04:50 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static DdNode * addMMRecur (DdManager *dd, DdNode *A, DdNode *B, int topP, int *vars);

// static DdNode * addTriangleRecur (DdManager *dd, DdNode *f, DdNode *g, int *vars, DdNode *cube);

// static DdNode * cuddAddOuterSumRecur (DdManager *dd, DdNode *M, DdNode *r, DdNode *c);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis [Calculates the product of two matrices represented as
  ADDs.]

  Description [Calculates the product of two matrices, A and B,
  represented as ADDs. This procedure implements the quasiring multiplication
  algorithm.  A is assumed to depend on variables x (rows) and z
  (columns).  B is assumed to depend on variables z (rows) and y
  (columns).  The product of A and B then depends on x (rows) and y
  (columns).  Only the z variables have to be explicitly identified;
  they are the "summation" variables.  Returns a pointer to the
  result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addTimesPlus Cudd_addTriangle Cudd_bddAndAbstract]

******************************************************************************/
func Cudd_addMatrixMultiply(dd *DdManager, A *DdNode, B *DdNode, z **DdNode, nz int32) *DdNode {
    var i int32;
    var nvars int32;
    var vars *int32;
    var res *DdNode; 

    /* Array vars says what variables are "summation" variables. */
    nvars = dd.size;
    vars = ABC_ALLOC(int,nvars);
    if vars == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    for i = 0; i < nvars; i++ {
        vars[i] = 0;
    }
    for i = 0; i < nz; i++ {
        vars[z[i].index] = 1;
    }

    for {
        dd.reordered = 0;
        res = addMMRecur(dd,A,B,-1,vars);
    
if !(dd.reordered == 1) { break }
} 
    ABC_FREE(vars);
    return(res);

} /* end of Cudd_addMatrixMultiply */


/**Function********************************************************************

  Synopsis    [Calculates the product of two matrices represented as
  ADDs.]

  Description [Calculates the product of two matrices, A and B,
  represented as ADDs, using the CMU matrix by matrix multiplication
  procedure by Clarke et al..  Matrix A has x's as row variables and z's
  as column variables, while matrix B has z's as row variables and y's
  as column variables. Returns the pointer to the result if successful;
  NULL otherwise. The resulting matrix has x's as row variables and y's
  as column variables.]

  SideEffects [None]

  SeeAlso     [Cudd_addMatrixMultiply]

******************************************************************************/
func Cudd_addTimesPlus(dd *DdManager, A *DdNode, B *DdNode, z **DdNode, nz int32) *DdNode {
    var w, cube, tmp, res *DdNode; 
    var i int32;
    tmp = Cudd_addApply(dd,Cudd_addTimes,A,B);
    if tmp == nil { return(nil); }
    Cudd_Ref(tmp);
    // inlining of macro DD_ONE
    Cudd_Ref(cube = ((dd).one));
    for i = nz-1; i >= 0; i-- {
         // inlining of macro DD_ZERO
         w = Cudd_addIte(dd,z[i],cube,((dd).zero));
         if w == nil {
            Cudd_RecursiveDeref(dd,tmp);
            return(nil);
         }
         Cudd_Ref(w);
         Cudd_RecursiveDeref(dd,cube);
         cube = w;
    }
    res = Cudd_addExistAbstract(dd,tmp,cube);
    if res == nil {
        Cudd_RecursiveDeref(dd,tmp);
        Cudd_RecursiveDeref(dd,cube);
        return(nil);
    }
    Cudd_Ref(res);
    Cudd_RecursiveDeref(dd,cube);
    Cudd_RecursiveDeref(dd,tmp);
    Cudd_Deref(res);
    return(res);

} /* end of Cudd_addTimesPlus */


/**Function********************************************************************

  Synopsis    [Performs the triangulation step for the shortest path
  computation.]

  Description [Implements the semiring multiplication algorithm used in
  the triangulation step for the shortest path computation.  f
  is assumed to depend on variables x (rows) and z (columns).  g is
  assumed to depend on variables z (rows) and y (columns).  The product
  of f and g then depends on x (rows) and y (columns).  Only the z
  variables have to be explicitly identified; they are the
  "abstraction" variables.  Returns a pointer to the result if
  successful; NULL otherwise. ]

  SideEffects [None]

  SeeAlso     [Cudd_addMatrixMultiply Cudd_bddAndAbstract]

******************************************************************************/
func Cudd_addTriangle(dd *DdManager, f *DdNode, g *DdNode, z **DdNode, nz int32) *DdNode {
    var i int32;
    var nvars int32;
    var vars *int32;
    var res, cube *DdNode;

    nvars = dd.size;
    vars = ABC_ALLOC(int, nvars);
    if vars == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    for i = 0; i < nvars; i++ { vars[i] = -1; }
    for i = 0; i < nz; i++ { vars[z[i].index] = i; }
    cube = Cudd_addComputeCube(dd, z, nil, nz);
    if cube == nil {
        ABC_FREE(vars);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(cube).ref);

    for {
        dd.reordered = 0;
        res = addTriangleRecur(dd, f, g, vars, cube);
    
if !(dd.reordered == 1) { break }
} 
    // inlining of macro cuddRef
    if res != nil { cuddSatInc(Cudd_Regular(res).ref); }
    Cudd_RecursiveDeref(dd,cube);
    // inlining of macro cuddDeref
    if res != nil { cuddSatDec(Cudd_Regular(res).ref); }
    ABC_FREE(vars);
    return(res);

} /* end of Cudd_addTriangle */


/**Function********************************************************************

  Synopsis    [Takes the minimum of a matrix and the outer sum of two vectors.]

  Description [Takes the pointwise minimum of a matrix and the outer
  sum of two vectors.  This procedure is used in the Floyd-Warshall
  all-pair shortest path algorithm.  Returns a pointer to the result if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_addOuterSum(dd *DdManager, M *DdNode, r *DdNode, c *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddAddOuterSumRecur(dd, M, r, c);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_addOuterSum */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_addMatrixMultiply.]

  Description [Performs the recursive step of Cudd_addMatrixMultiply.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

******************************************************************************/
func addMMRecur(dd *DdManager, A *DdNode, B *DdNode, topP int32, vars *int32) *DdNode {
    DdNode *zero,
           *At,         /* positive cofactor of first operand */
           *Ae,         /* negative cofactor of first operand */
           *Bt,         /* positive cofactor of second operand */
           *Be,         /* negative cofactor of second operand */
           *t,          /* positive cofactor of result */
           *e,          /* negative cofactor of result */
           *scaled,     /* scaled result */
           *add_scale,  /* ADD representing the scaling factor */
           *res;
    var i int32;              /* loop index */
    var scale float64;       /* scaling factor */
    var index int32;          /* index of the top variable */
    CUDD_VALUE_TYPE value;
    unsigned int topA, topB, topV;
    DD_CTFP cacheOp;

    statLine(dd);
    // inlining of macro DD_ZERO
    zero = ((dd).zero);

    if A == zero || B == zero {
        return(zero);
    }

    // inlining of macro cuddIsConstant
    if ((A).index == CUDD_CONST_INDEX) && ((B).index == CUDD_CONST_INDEX) {
        /* Compute the scaling factor. It is 2^k, where k is the
        ** number of summation variables below the current variable.
        ** Indeed, these constants represent blocks of 2^k identical
        ** constant values in both A and B.
        */
        // inlining of macro cuddV
        value = ((A).type.value) * ((B).type.value);
        for i = 0; i < dd.size; i++ {
            if vars[i] {
                if dd.perm[i] > topP {
                    value *= (CUDD_VALUE_TYPE) 2;
                }
            }
        }
        res = cuddUniqueConst(dd, value);
        return(res);
    }

    /* Standardize to increase cache efficiency. Clearly, A*B != B*A
    ** in matrix multiplication. However, which matrix is which is
    ** determined by the variables appearing in the ADDs and not by
    ** which one is passed as first argument.
    */
    if A > B {
        var tmp *DdNode  = A;
        A = B;
        B = tmp;
    }

    // inlining of macro cuddI
    
    if (A.index)==CUDD_CONST_INDEX {
        topA = int32((A.index));
    } else {
        topA = (dd).perm[(A.index)];
    // inlining of macro cuddI
    } 
    if (B.index)==CUDD_CONST_INDEX {
        topB = int32((B.index));
    } else {
        topB = (dd).perm[(B.index)];
    }
    // inlining of macro ddMin
    
    if (topB) < (topA) {
        topV = (topB);
    } else {
        topV = (topA);
    }

    cacheOp = (DD_CTFP) addMMRecur;
    res = cuddCacheLookup2(dd,cacheOp,A,B);
    if res != nil {
        /* If the result is 0, there is no need to normalize.
        ** Otherwise we count the number of z variables between
        ** the current depth and the top of the ADDs. These are
        ** the missing variables that determine the size of the
        ** constant blocks.
        */
        if res == zero { return(res); }
        scale = 1.0;
        for i = 0; i < dd.size; i++ {
            if vars[i] {
                if dd.perm[i] > topP && uint32( dd.perm[i]) < topV {
                    scale *= 2;
                }
            }
        }
        if scale > 1.0 {
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res).ref);
            add_scale = cuddUniqueConst(dd,(CUDD_VALUE_TYPE)scale);
            if add_scale == nil {
                Cudd_RecursiveDeref(dd, res);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(add_scale).ref);
            scaled = cuddAddApplyRecur(dd,Cudd_addTimes,res,add_scale);
            if scaled == nil {
                Cudd_RecursiveDeref(dd, add_scale);
                Cudd_RecursiveDeref(dd, res);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(scaled).ref);
            Cudd_RecursiveDeref(dd, add_scale);
            Cudd_RecursiveDeref(dd, res);
            res = scaled;
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(res).ref);
        }
        return(res);
    }

    /* compute the cofactors */
    if topV == topA {
        // inlining of macro cuddT
        At = ((A).type.kids.T);
        // inlining of macro cuddE
        Ae = ((A).type.kids.E);
    } else {
        At = Ae = A;
    }
    if topV == topB {
        // inlining of macro cuddT
        Bt = ((B).type.kids.T);
        // inlining of macro cuddE
        Be = ((B).type.kids.E);
    } else {
        Bt = Be = B;
    }

    t = addMMRecur(dd, At, Bt, int32(topV), vars);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);
    e = addMMRecur(dd, Ae, Be, int32(topV), vars);
    if e == nil {
        Cudd_RecursiveDeref(dd, t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    index = dd.invperm[topV];
    if vars[index] == 0 {
        /* We have split on either the rows of A or the columns
        ** of B. We just need to connect the two subresults,
        ** which correspond to two submatrices of the result.
        */
        if t == e {
            res = t;
        } else {
            res = cuddUniqueInter(dd,index,t,e);
        }
        if res == nil {
            Cudd_RecursiveDeref(dd, t);
            Cudd_RecursiveDeref(dd, e);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    } else {
        /* we have simultaneously split on the columns of A and
        ** the rows of B. The two subresults must be added.
        */
        res = cuddAddApplyRecur(dd,Cudd_addPlus,t,e);
        if res == nil {
            Cudd_RecursiveDeref(dd, t);
            Cudd_RecursiveDeref(dd, e);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_RecursiveDeref(dd, t);
        Cudd_RecursiveDeref(dd, e);
    }

    cuddCacheInsert2(dd,cacheOp,A,B,res);

    /* We have computed (and stored in the computed table) a minimal
    ** result; that is, a result that assumes no summation variables
    ** between the current depth of the recursion and its top
    ** variable. We now take into account the z variables by properly
    ** scaling the result.
    */
    if res != zero {
        scale = 1.0;
        for i = 0; i < dd.size; i++ {
            if vars[i] {
                if dd.perm[i] > topP && uint32( dd.perm[i]) < topV {
                    scale *= 2;
                }
            }
        }
        if scale > 1.0 {
            add_scale = cuddUniqueConst(dd,(CUDD_VALUE_TYPE)scale);
            if add_scale == nil {
                Cudd_RecursiveDeref(dd, res);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(add_scale).ref);
            scaled = cuddAddApplyRecur(dd,Cudd_addTimes,res,add_scale);
            if scaled == nil {
                Cudd_RecursiveDeref(dd, res);
                Cudd_RecursiveDeref(dd, add_scale);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(scaled).ref);
            Cudd_RecursiveDeref(dd, add_scale);
            Cudd_RecursiveDeref(dd, res);
            res = scaled;
        }
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of addMMRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_addTriangle.]

  Description [Performs the recursive step of Cudd_addTriangle. Returns
  a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

******************************************************************************/
func addTriangleRecur(dd *DdManager, f *DdNode, g *DdNode, vars *int32, cube *DdNode) *DdNode {
    var fv, fvn, gv, gvn, t, e, res *DdNode;
    CUDD_VALUE_TYPE value;
    var top, topf, topg, index int32;

    statLine(dd);
    // inlining of macro DD_PLUS_INFINITY
    if f == ((dd).plusinfinity) || g == ((dd).plusinfinity) {
        // inlining of macro DD_PLUS_INFINITY
        return(((dd).plusinfinity));
    }

    // inlining of macro cuddIsConstant
    if ((f).index == CUDD_CONST_INDEX) && ((g).index == CUDD_CONST_INDEX) {
        // inlining of macro cuddV
        value = ((f).type.value) + ((g).type.value);
        res = cuddUniqueConst(dd, value);
        return(res);
    }
    if f < g {
        var tmp *DdNode  = f;
        f = g;
        g = tmp;
    }

    if f.ref != 1 || g.ref != 1 {
        res = cuddCacheLookup(dd, DD_ADD_TRIANGLE_TAG, f, g, cube);
        if res != nil {
            return(res);
        }
    }

    // inlining of macro cuddI
    
    if (f.index)==CUDD_CONST_INDEX {
        topf = int32((f.index));
    } else {
        topf = (dd).perm[(f.index)];
    // inlining of macro cuddI
    } 
    if (g.index)==CUDD_CONST_INDEX {
        topg = int32((g.index));
    } else {
        topg = (dd).perm[(g.index)];
    }
    // inlining of macro ddMin
    
    if (topg) < (topf) {
        top = (topg);
    } else {
        top = (topf);
    }

    // inlining of macro cuddT
    // inlining of macro cuddE
    if top == topf {fv = ((f).type.kids.T); fvn = ((f).type.kids.E);} else {fv = fvn = f;}
    // inlining of macro cuddT
    // inlining of macro cuddE
    if top == topg {gv = ((g).type.kids.T); gvn = ((g).type.kids.E);} else {gv = gvn = g;}

    t = addTriangleRecur(dd, fv, gv, vars, cube);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);
    e = addTriangleRecur(dd, fvn, gvn, vars, cube);
    if e == nil {
        Cudd_RecursiveDeref(dd, t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    index = dd.invperm[top];
    if vars[index] < 0 {
        if t == e {
            res = t;
        } else {
            res = cuddUniqueInter(dd,index,t,e);
        }
        if res == nil {
            Cudd_RecursiveDeref(dd, t);
            Cudd_RecursiveDeref(dd, e);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    } else {
        res = cuddAddApplyRecur(dd,Cudd_addMinimum,t,e);
        if res == nil {
            Cudd_RecursiveDeref(dd, t);
            Cudd_RecursiveDeref(dd, e);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_RecursiveDeref(dd, t);
        Cudd_RecursiveDeref(dd, e);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref);
    }

    if f.ref != 1 || g.ref != 1 {
        cuddCacheInsert(dd, DD_ADD_TRIANGLE_TAG, f, g, cube, res);
    }

    return(res);

} /* end of addTriangleRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_addOuterSum.]

  Description [Performs the recursive step of Cudd_addOuterSum.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddAddOuterSumRecur(dd *DdManager, M *DdNode, r *DdNode, c *DdNode) *DdNode {
    var P, R, Mt, Me, rt, re, ct, ce, Rt, Re *DdNode;
    var topM, topc, topr int32;
    var v, index int32;

    statLine(dd);
    /* Check special cases. */
    // inlining of macro DD_PLUS_INFINITY
    if r == ((dd).plusinfinity) || c == ((dd).plusinfinity) { return(M); } 

    // inlining of macro cuddIsConstant
    if ((c).index == CUDD_CONST_INDEX) && ((r).index == CUDD_CONST_INDEX) {
        // inlining of macro Cudd_V
        // inlining of macro Cudd_Regular
        R = cuddUniqueConst(dd,((((DdNode *)((ptruint)(c) & ~01))).type.value)+((((DdNode *)((ptruint)(r) & ~01))).type.value));
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(R).ref);
        // inlining of macro cuddIsConstant
        if (M).index == CUDD_CONST_INDEX {
            // inlining of macro cuddV
            if ((R).type.value) <= ((M).type.value) {
                // inlining of macro cuddDeref
                cuddSatDec(Cudd_Regular(R).ref);
                return(R);
            } else {
                Cudd_RecursiveDeref(dd,R);       
                return(M);
            }
        } else {
            P = Cudd_addApply(dd,Cudd_addMinimum,R,M);
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(P).ref);
            Cudd_RecursiveDeref(dd,R);
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(P).ref);
            return(P);
        }
    }

    /* Check the cache. */
    R = cuddCacheLookup(dd,DD_ADD_OUT_SUM_TAG,M,r,c);
    if R != nil { return(R); }

    // inlining of macro cuddI
    
    if (M.index)==CUDD_CONST_INDEX {
        topM = int32((M.index));
    } else {
        topM = (dd).perm[(M.index)];
    // inlining of macro cuddI
    } 
    if (r.index)==CUDD_CONST_INDEX {
        topr = int32((r.index));
    } else {
        topr = (dd).perm[(r.index)];
    }
    // inlining of macro cuddI
    
    if (c.index)==CUDD_CONST_INDEX {
        topc = int32((c.index));
    } else {
        topc = (dd).perm[(c.index)];
    }
    // inlining of macro ddMin
    
    // inlining of macro ddMin
    if ((((topc) < (topr)) ? (topc) : (topr))) < (topM) {
        // inlining of macro ddMin
        v = ((((topc) < (topr)) ? (topc) : (topr)));
    } else {
        v = (topM);
    }

    /* Compute cofactors. */
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topM == v { Mt = ((M).type.kids.T); Me = ((M).type.kids.E); } else { Mt = Me = M; }
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topr == v { rt = ((r).type.kids.T); re = ((r).type.kids.E); } else { rt = re = r; }
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topc == v { ct = ((c).type.kids.T); ce = ((c).type.kids.E); } else { ct = ce = c; }

    /* Recursively solve. */
    Rt = cuddAddOuterSumRecur(dd,Mt,rt,ct);
    if Rt == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(Rt).ref);
    Re = cuddAddOuterSumRecur(dd,Me,re,ce);
    if Re == nil {
        Cudd_RecursiveDeref(dd, Rt);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(Re).ref);
    index = dd.invperm[v];
    if Rt == Re {
        R = Rt;
    } else {
        R = cuddUniqueInter(dd,index,Rt,Re);
    }
    if R == nil {
        Cudd_RecursiveDeref(dd, Rt);
        Cudd_RecursiveDeref(dd, Re);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(Rt).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(Re).ref);

    /* Store the result in the cache. */
    cuddCacheInsert(dd,DD_ADD_OUT_SUM_TAG,M,r,c,R);

    return(R);

} /* end of cuddAddOuterSumRecur */


//ABC_NAMESPACE_IMPL_END


package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddPriority.c]

  PackageName [cudd]

  Synopsis    [Priority functions.]

  Description [External procedures included in this file:
            <ul>
            <li> Cudd_PrioritySelect()
            <li> Cudd_Xgty()
            <li> Cudd_Xeqy()
            <li> Cudd_addXeqy()
            <li> Cudd_Dxygtdxz()
            <li> Cudd_Dxygtdyz()
            <li> Cudd_Inequality()
            <li> Cudd_Disequality()
            <li> Cudd_bddInterval()
            <li> Cudd_CProjection()
            <li> Cudd_addHamming()
            <li> Cudd_MinHammingDist()
            <li> Cudd_bddClosestCube()
            </ul>
        Internal procedures included in this module:
            <ul>
            <li> cuddCProjectionRecur()
            <li> cuddBddClosestCube()
            </ul>
        Static procedures included in this module:
            <ul>
            <li> cuddMinHammingDistRecur()
            <li> separateCube()
            <li> createResult()
            </ul>
            ]

  SeeAlso     []

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

const DD_DEBUG = 1

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
static char rcsid[] DD_UNUSED = "$Id: cuddPriority.c,v 1.33 2009/02/20 02:14:58 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/
// static int cuddMinHammingDistRecur (DdNode * f, int *minterm, DdHashTable * table, int upperBound);

// static DdNode * separateCube (DdManager *dd, DdNode *f, CUDD_VALUE_TYPE *distance);

// static DdNode * createResult (DdManager *dd, unsigned int index, unsigned int phase, DdNode *cube, CUDD_VALUE_TYPE distance);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Selects pairs from R using a priority function.]

  Description [Selects pairs from a relation R(x,y) (given as a BDD)
  in such a way that a given x appears in one pair only. Uses a
  priority function to determine which y should be paired to a given x.
  Cudd_PrioritySelect returns a pointer to
  the selected function if successful; NULL otherwise.
  Three of the arguments--x, y, and z--are vectors of BDD variables.
  The first two are the variables on which R depends. The third vectore
  is a vector of auxiliary variables, used during the computation. This
  vector is optional. If a NULL value is passed instead,
  Cudd_PrioritySelect will create the working variables on the fly.
  The sizes of x and y (and z if it is not NULL) should equal n.
  The priority function Pi can be passed as a BDD, or can be built by
  Cudd_PrioritySelect. If NULL is passed instead of a DdNode *,
  parameter Pifunc is used by Cudd_PrioritySelect to build a BDD for the
  priority function. (Pifunc is a pointer to a C function.) If Pi is not
  NULL, then Pifunc is ignored. Pifunc should have the same interface as
  the standard priority functions (e.g., Cudd_Dxygtdxz).
  Cudd_PrioritySelect and Cudd_CProjection can sometimes be used
  interchangeably. Specifically, calling Cudd_PrioritySelect with
  Cudd_Xgty as Pifunc produces the same result as calling
  Cudd_CProjection with the all-zero minterm as reference minterm.
  However, depending on the application, one or the other may be
  preferable:
  <ul>
  <li> When extracting representatives from an equivalence relation,
  Cudd_CProjection has the advantage of nor requiring the auxiliary
  variables.
  <li> When computing matchings in general bipartite graphs,
  Cudd_PrioritySelect normally obtains better results because it can use
  more powerful matching schemes (e.g., Cudd_Dxygtdxz).
  </ul>
  ]

  SideEffects [If called with z == NULL, will create new variables in
  the manager.]

  SeeAlso     [Cudd_Dxygtdxz Cudd_Dxygtdyz Cudd_Xgty
  Cudd_bddAdjPermuteX Cudd_CProjection]

******************************************************************************/
DdNode *
Cudd_PrioritySelect(
  DdManager * dd /* manager */,
  DdNode * R /* BDD of the relation */,
  DdNode ** x /* array of x variables */,
  DdNode ** y /* array of y variables */,
  DdNode ** z /* array of z variables (optional: may be NULL) */,
  DdNode * Pi /* BDD of the priority function (optional: may be NULL) */,
  int  n /* size of x, y, and z */,
  DD_PRFP Pifunc /* function used to build Pi if it is NULL */)
{
    var res *DdNode  = nil;
    var zcube *DdNode  = nil;
    var Rxz, Q *DdNode;
    var createdZ int32;
    var createdPi int32;
    var i int32;

    /* Create z variables if needed. */
    if z == nil {
        if Pi != nil { return(nil); }
        z = ABC_ALLOC(DdNode *,n);
        if z == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(nil);
        }
        createdZ = 1;
        for i = 0; i < n; i++ {
            if dd.size >= int32( CUDD_MAXINDEX) - 1 { goto endgame; }
            // inlining of macro Cudd_Not
            z[i] = cuddUniqueInter(dd,dd.size,dd.one,((DdNode *)((ptrint)(dd.one) ^ 01)));
            if z[i] == nil { goto endgame; }
        }
    }

    /* Create priority function BDD if needed. */
    if Pi == nil {
        Pi = Pifunc(dd,n,x,y,z);
        if Pi == nil { goto endgame; }
        createdPi = 1;
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(Pi).ref);
    }

    /* Initialize abstraction cube. */
    // inlining of macro DD_ONE
    zcube = ((dd).one);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(zcube).ref);
    for i = n - 1; i >= 0; i-- {
        var tmpp *DdNode;
        tmpp = Cudd_bddAnd(dd,z[i],zcube);
        if tmpp == nil { goto endgame; }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(tmpp).ref);
        Cudd_RecursiveDeref(dd,zcube);
        zcube = tmpp;
    }

    /* Compute subset of (x,y) pairs. */
    Rxz = Cudd_bddSwapVariables(dd,R,y,z,n);
    if Rxz == nil { goto endgame; }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(Rxz).ref);
    Q = Cudd_bddAndAbstract(dd,Rxz,Pi,zcube);
    if Q == nil {
        Cudd_RecursiveDeref(dd,Rxz);
        goto endgame;
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(Q).ref);
    Cudd_RecursiveDeref(dd,Rxz);
    // inlining of macro Cudd_Not
    res = Cudd_bddAnd(dd,R,((DdNode *)((ptrint)(Q) ^ 01)));
    if res == nil {
        Cudd_RecursiveDeref(dd,Q);
        goto endgame;
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref(dd,Q);

endgame:
    if zcube != nil { Cudd_RecursiveDeref(dd,zcube); }
    if createdZ != 0 {
        ABC_FREE(z);
    }
    if createdPi != 0 {
        Cudd_RecursiveDeref(dd,Pi);
    }
    // inlining of macro cuddDeref
    if res != nil { cuddSatDec(Cudd_Regular(res).ref); }
    return(res);

} /* Cudd_PrioritySelect */


/**Function********************************************************************

  Synopsis    [Generates a BDD for the function x &gt; y.]

  Description [This function generates a BDD for the function x &gt; y.
  Both x and y are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\] and
  y\[0\] y\[1\] ...  y\[N-1\], with 0 the most significant bit.
  The BDD is built bottom-up.
  It has 3*N-1 internal nodes, if the variables are ordered as follows:
  x\[0\] y\[0\] x\[1\] y\[1\] ... x\[N-1\] y\[N-1\].
  Argument z is not used by Cudd_Xgty: it is included to make it
  call-compatible to Cudd_Dxygtdxz and Cudd_Dxygtdyz.]

  SideEffects [None]

  SeeAlso     [Cudd_PrioritySelect Cudd_Dxygtdxz Cudd_Dxygtdyz]

******************************************************************************/
DdNode *
Cudd_Xgty(
  DdManager * dd /* DD manager */,
  int  N /* number of x and y variables */,
  DdNode ** z /* array of z variables: unused */,
  DdNode ** x /* array of x variables */,
  DdNode ** y /* array of y variables */)
{
    var u, v, w *DdNode;
    var i int32;

    /* Build bottom part of BDD outside loop. */
    // inlining of macro Cudd_Not
    u = Cudd_bddAnd(dd, x[N-1], ((DdNode *)((ptrint)(y[N-1]) ^ 01)));
    if u == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(u).ref);

    /* Loop to build the rest of the BDD. */
    for i = N-2; i >= 0; i-- {
        // inlining of macro Cudd_Not
        v = Cudd_bddAnd(dd, y[i], ((DdNode *)((ptrint)(u) ^ 01)));
        if v == nil {
            Cudd_RecursiveDeref(dd, u);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(v).ref);
        // inlining of macro Cudd_Not
        w = Cudd_bddAnd(dd, ((DdNode *)((ptrint)(y[i]) ^ 01)), u);
        if w == nil {
            Cudd_RecursiveDeref(dd, u);
            Cudd_RecursiveDeref(dd, v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(w).ref);
        Cudd_RecursiveDeref(dd, u);
        // inlining of macro Cudd_Not
        u = Cudd_bddIte(dd, x[i], ((DdNode *)((ptrint)(v) ^ 01)), w);
        if u == nil {
            Cudd_RecursiveDeref(dd, v);
            Cudd_RecursiveDeref(dd, w);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(u).ref);
        Cudd_RecursiveDeref(dd, v);
        Cudd_RecursiveDeref(dd, w);

    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(u).ref);
    return(u);

} /* end of Cudd_Xgty */


/**Function********************************************************************

  Synopsis    [Generates a BDD for the function x==y.]

  Description [This function generates a BDD for the function x==y.
  Both x and y are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\] and
  y\[0\] y\[1\] ...  y\[N-1\], with 0 the most significant bit.
  The BDD is built bottom-up.
  It has 3*N-1 internal nodes, if the variables are ordered as follows:
  x\[0\] y\[0\] x\[1\] y\[1\] ... x\[N-1\] y\[N-1\]. ]

  SideEffects [None]

  SeeAlso     [Cudd_addXeqy]

******************************************************************************/
DdNode *
Cudd_Xeqy(
  DdManager * dd /* DD manager */,
  int  N /* number of x and y variables */,
  DdNode ** x /* array of x variables */,
  DdNode ** y /* array of y variables */)
{
    var u, v, w *DdNode;
    var i int32;

    /* Build bottom part of BDD outside loop. */
    // inlining of macro Cudd_Not
    u = Cudd_bddIte(dd, x[N-1], y[N-1], ((DdNode *)((ptrint)(y[N-1]) ^ 01)));
    if u == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(u).ref);

    /* Loop to build the rest of the BDD. */
    for i = N-2; i >= 0; i-- {
        v = Cudd_bddAnd(dd, y[i], u);
        if v == nil {
            Cudd_RecursiveDeref(dd, u);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(v).ref);
        // inlining of macro Cudd_Not
        w = Cudd_bddAnd(dd, ((DdNode *)((ptrint)(y[i]) ^ 01)), u);
        if w == nil {
            Cudd_RecursiveDeref(dd, u);
            Cudd_RecursiveDeref(dd, v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(w).ref);
        Cudd_RecursiveDeref(dd, u);
        u = Cudd_bddIte(dd, x[i], v, w);
        if u == nil {
            Cudd_RecursiveDeref(dd, v);
            Cudd_RecursiveDeref(dd, w);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(u).ref);
        Cudd_RecursiveDeref(dd, v);
        Cudd_RecursiveDeref(dd, w);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(u).ref);
    return(u);

} /* end of Cudd_Xeqy */


/**Function********************************************************************

  Synopsis    [Generates an ADD for the function x==y.]

  Description [This function generates an ADD for the function x==y.
  Both x and y are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\] and
  y\[0\] y\[1\] ...  y\[N-1\], with 0 the most significant bit.
  The ADD is built bottom-up.
  It has 3*N-1 internal nodes, if the variables are ordered as follows:
  x\[0\] y\[0\] x\[1\] y\[1\] ... x\[N-1\] y\[N-1\]. ]

  SideEffects [None]

  SeeAlso     [Cudd_Xeqy]

******************************************************************************/
DdNode *
Cudd_addXeqy(
  DdManager * dd /* DD manager */,
  int  N /* number of x and y variables */,
  DdNode ** x /* array of x variables */,
  DdNode ** y /* array of y variables */)
{
    var one, zero *DdNode;
    var u, v, w *DdNode;
    var i int32;

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro DD_ZERO
    zero = ((dd).zero);

    /* Build bottom part of ADD outside loop. */
    v = Cudd_addIte(dd, y[N-1], one, zero);
    if v == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(v).ref);
    w = Cudd_addIte(dd, y[N-1], zero, one);
    if w == nil {
        Cudd_RecursiveDeref(dd, v);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(w).ref);
    u = Cudd_addIte(dd, x[N-1], v, w);
    if u == nil {
        Cudd_RecursiveDeref(dd, v);
        Cudd_RecursiveDeref(dd, w);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(u).ref);
    Cudd_RecursiveDeref(dd, v);
    Cudd_RecursiveDeref(dd, w);

    /* Loop to build the rest of the ADD. */
    for i = N-2; i >= 0; i-- {
        v = Cudd_addIte(dd, y[i], u, zero);
        if v == nil {
            Cudd_RecursiveDeref(dd, u);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(v).ref);
        w = Cudd_addIte(dd, y[i], zero, u);
        if w == nil {
            Cudd_RecursiveDeref(dd, u);
            Cudd_RecursiveDeref(dd, v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(w).ref);
        Cudd_RecursiveDeref(dd, u);
        u = Cudd_addIte(dd, x[i], v, w);
        if w == nil {
            Cudd_RecursiveDeref(dd, v);
            Cudd_RecursiveDeref(dd, w);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(u).ref);
        Cudd_RecursiveDeref(dd, v);
        Cudd_RecursiveDeref(dd, w);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(u).ref);
    return(u);

} /* end of Cudd_addXeqy */


/**Function********************************************************************

  Synopsis    [Generates a BDD for the function d(x,y) &gt; d(x,z).]

  Description [This function generates a BDD for the function d(x,y)
  &gt; d(x,z);
  x, y, and z are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\],
  y\[0\] y\[1\] ...  y\[N-1\], and z\[0\] z\[1\] ...  z\[N-1\],
  with 0 the most significant bit.
  The distance d(x,y) is defined as:
        \sum_{i=0}^{N-1}(|x_i - y_i| \cdot 2^{N-i-1}).
  The BDD is built bottom-up.
  It has 7*N-3 internal nodes, if the variables are ordered as follows:
  x\[0\] y\[0\] z\[0\] x\[1\] y\[1\] z\[1\] ... x\[N-1\] y\[N-1\] z\[N-1\]. ]

  SideEffects [None]

  SeeAlso     [Cudd_PrioritySelect Cudd_Dxygtdyz Cudd_Xgty Cudd_bddAdjPermuteX]

******************************************************************************/
DdNode *
Cudd_Dxygtdxz(
  DdManager * dd /* DD manager */,
  int  N /* number of x, y, and z variables */,
  DdNode ** x /* array of x variables */,
  DdNode ** y /* array of y variables */,
  DdNode ** z /* array of z variables */)
{
    var one, zero *DdNode;
    var z1, z2, z3, z4, y1_, y2, x1 *DdNode;
    var i int32;

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Build bottom part of BDD outside loop. */
    // inlining of macro Cudd_Not
    y1_ = Cudd_bddIte(dd, y[N-1], one, ((DdNode *)((ptrint)(z[N-1]) ^ 01)));
    if y1_ == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(y1_).ref);
    y2 = Cudd_bddIte(dd, y[N-1], z[N-1], one);
    if y2 == nil {
        Cudd_RecursiveDeref(dd, y1_);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(y2).ref);
    x1 = Cudd_bddIte(dd, x[N-1], y1_, y2);
    if x1 == nil {
        Cudd_RecursiveDeref(dd, y1_);
        Cudd_RecursiveDeref(dd, y2);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(x1).ref);
    Cudd_RecursiveDeref(dd, y1_);
    Cudd_RecursiveDeref(dd, y2);

    /* Loop to build the rest of the BDD. */
    for i = N-2; i >= 0; i-- {
        // inlining of macro Cudd_Not
        z1 = Cudd_bddIte(dd, z[i], one, ((DdNode *)((ptrint)(x1) ^ 01)));
        if z1 == nil {
            Cudd_RecursiveDeref(dd, x1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(z1).ref);
        z2 = Cudd_bddIte(dd, z[i], x1, one);
        if z2 == nil {
            Cudd_RecursiveDeref(dd, x1);
            Cudd_RecursiveDeref(dd, z1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(z2).ref);
        z3 = Cudd_bddIte(dd, z[i], one, x1);
        if z3 == nil {
            Cudd_RecursiveDeref(dd, x1);
            Cudd_RecursiveDeref(dd, z1);
            Cudd_RecursiveDeref(dd, z2);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(z3).ref);
        z4 = Cudd_bddIte(dd, z[i], x1, zero);
        if z4 == nil {
            Cudd_RecursiveDeref(dd, x1);
            Cudd_RecursiveDeref(dd, z1);
            Cudd_RecursiveDeref(dd, z2);
            Cudd_RecursiveDeref(dd, z3);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(z4).ref);
        Cudd_RecursiveDeref(dd, x1);
        // inlining of macro Cudd_Not
        y1_ = Cudd_bddIte(dd, y[i], z2, ((DdNode *)((ptrint)(z1) ^ 01)));
        if y1_ == nil {
            Cudd_RecursiveDeref(dd, z1);
            Cudd_RecursiveDeref(dd, z2);
            Cudd_RecursiveDeref(dd, z3);
            Cudd_RecursiveDeref(dd, z4);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(y1_).ref);
        y2 = Cudd_bddIte(dd, y[i], z4, z3);
        if y2 == nil {
            Cudd_RecursiveDeref(dd, z1);
            Cudd_RecursiveDeref(dd, z2);
            Cudd_RecursiveDeref(dd, z3);
            Cudd_RecursiveDeref(dd, z4);
            Cudd_RecursiveDeref(dd, y1_);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(y2).ref);
        Cudd_RecursiveDeref(dd, z1);
        Cudd_RecursiveDeref(dd, z2);
        Cudd_RecursiveDeref(dd, z3);
        Cudd_RecursiveDeref(dd, z4);
        x1 = Cudd_bddIte(dd, x[i], y1_, y2);
        if x1 == nil {
            Cudd_RecursiveDeref(dd, y1_);
            Cudd_RecursiveDeref(dd, y2);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(x1).ref);
        Cudd_RecursiveDeref(dd, y1_);
        Cudd_RecursiveDeref(dd, y2);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(x1).ref);
    // inlining of macro Cudd_Not
    return(((DdNode *)((ptrint)(x1) ^ 01)));

} /* end of Cudd_Dxygtdxz */


/**Function********************************************************************

  Synopsis    [Generates a BDD for the function d(x,y) &gt; d(y,z).]

  Description [This function generates a BDD for the function d(x,y)
  &gt; d(y,z);
  x, y, and z are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\],
  y\[0\] y\[1\] ...  y\[N-1\], and z\[0\] z\[1\] ...  z\[N-1\],
  with 0 the most significant bit.
  The distance d(x,y) is defined as:
        \sum_{i=0}^{N-1}(|x_i - y_i| \cdot 2^{N-i-1}).
  The BDD is built bottom-up.
  It has 7*N-3 internal nodes, if the variables are ordered as follows:
  x\[0\] y\[0\] z\[0\] x\[1\] y\[1\] z\[1\] ... x\[N-1\] y\[N-1\] z\[N-1\]. ]

  SideEffects [None]

  SeeAlso     [Cudd_PrioritySelect Cudd_Dxygtdxz Cudd_Xgty Cudd_bddAdjPermuteX]

******************************************************************************/
DdNode *
Cudd_Dxygtdyz(
  DdManager * dd /* DD manager */,
  int  N /* number of x, y, and z variables */,
  DdNode ** x /* array of x variables */,
  DdNode ** y /* array of y variables */,
  DdNode ** z /* array of z variables */)
{
    var one, zero *DdNode;
    var z1, z2, z3, z4, y1_, y2, x1 *DdNode;
    var i int32;

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Build bottom part of BDD outside loop. */
    y1_ = Cudd_bddIte(dd, y[N-1], one, z[N-1]);
    if y1_ == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(y1_).ref);
    y2 = Cudd_bddIte(dd, y[N-1], z[N-1], zero);
    if y2 == nil {
        Cudd_RecursiveDeref(dd, y1_);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(y2).ref);
    // inlining of macro Cudd_Not
    x1 = Cudd_bddIte(dd, x[N-1], y1_, ((DdNode *)((ptrint)(y2) ^ 01)));
    if x1 == nil {
        Cudd_RecursiveDeref(dd, y1_);
        Cudd_RecursiveDeref(dd, y2);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(x1).ref);
    Cudd_RecursiveDeref(dd, y1_);
    Cudd_RecursiveDeref(dd, y2);

    /* Loop to build the rest of the BDD. */
    for i = N-2; i >= 0; i-- {
        z1 = Cudd_bddIte(dd, z[i], x1, zero);
        if z1 == nil {
            Cudd_RecursiveDeref(dd, x1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(z1).ref);
        z2 = Cudd_bddIte(dd, z[i], x1, one);
        if z2 == nil {
            Cudd_RecursiveDeref(dd, x1);
            Cudd_RecursiveDeref(dd, z1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(z2).ref);
        z3 = Cudd_bddIte(dd, z[i], one, x1);
        if z3 == nil {
            Cudd_RecursiveDeref(dd, x1);
            Cudd_RecursiveDeref(dd, z1);
            Cudd_RecursiveDeref(dd, z2);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(z3).ref);
        // inlining of macro Cudd_Not
        z4 = Cudd_bddIte(dd, z[i], one, ((DdNode *)((ptrint)(x1) ^ 01)));
        if z4 == nil {
            Cudd_RecursiveDeref(dd, x1);
            Cudd_RecursiveDeref(dd, z1);
            Cudd_RecursiveDeref(dd, z2);
            Cudd_RecursiveDeref(dd, z3);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(z4).ref);
        Cudd_RecursiveDeref(dd, x1);
        y1_ = Cudd_bddIte(dd, y[i], z2, z1);
        if y1_ == nil {
            Cudd_RecursiveDeref(dd, z1);
            Cudd_RecursiveDeref(dd, z2);
            Cudd_RecursiveDeref(dd, z3);
            Cudd_RecursiveDeref(dd, z4);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(y1_).ref);
        // inlining of macro Cudd_Not
        y2 = Cudd_bddIte(dd, y[i], z4, ((DdNode *)((ptrint)(z3) ^ 01)));
        if y2 == nil {
            Cudd_RecursiveDeref(dd, z1);
            Cudd_RecursiveDeref(dd, z2);
            Cudd_RecursiveDeref(dd, z3);
            Cudd_RecursiveDeref(dd, z4);
            Cudd_RecursiveDeref(dd, y1_);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(y2).ref);
        Cudd_RecursiveDeref(dd, z1);
        Cudd_RecursiveDeref(dd, z2);
        Cudd_RecursiveDeref(dd, z3);
        Cudd_RecursiveDeref(dd, z4);
        // inlining of macro Cudd_Not
        x1 = Cudd_bddIte(dd, x[i], y1_, ((DdNode *)((ptrint)(y2) ^ 01)));
        if x1 == nil {
            Cudd_RecursiveDeref(dd, y1_);
            Cudd_RecursiveDeref(dd, y2);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(x1).ref);
        Cudd_RecursiveDeref(dd, y1_);
        Cudd_RecursiveDeref(dd, y2);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(x1).ref);
    // inlining of macro Cudd_Not
    return(((DdNode *)((ptrint)(x1) ^ 01)));

} /* end of Cudd_Dxygtdyz */


/**Function********************************************************************

  Synopsis    [Generates a BDD for the function x - y &ge; c.]

  Description [This function generates a BDD for the function x -y &ge; c.
  Both x and y are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\] and
  y\[0\] y\[1\] ...  y\[N-1\], with 0 the most significant bit.
  The BDD is built bottom-up.
  It has a linear number of nodes if the variables are ordered as follows:
  x\[0\] y\[0\] x\[1\] y\[1\] ... x\[N-1\] y\[N-1\].]

  SideEffects [None]

  SeeAlso     [Cudd_Xgty]

******************************************************************************/
DdNode *
Cudd_Inequality(
  DdManager * dd /* DD manager */,
  int  N /* number of x and y variables */,
  int c /* right-hand side constant */,
  DdNode ** x /* array of x variables */,
  DdNode ** y /* array of y variables */)
{
    /* The nodes at level i represent values of the difference that are
    ** multiples of 2^i.  We use variables with names starting with k
    ** to denote the multipliers of 2^i in such multiples. */
    var kTrue int32  = c;
    var kFalse int32  = c - 1;
    /* Mask used to compute the ceiling function.  Since we divide by 2^i,
    ** we want to know whether the dividend is a multiple of 2^i.  If it is,
    ** then ceiling and floor coincide; otherwise, they differ by one. */
    var mask int32  = 1;
    var i int32;

    var f *DdNode  = nil;           /* the eventual result */
    // inlining of macro DD_ONE
    var one *DdNode  = ((dd).one);
    // inlining of macro Cudd_Not
    var zero *DdNode  = ((DdNode *)((ptrint)(one) ^ 01));

    /* Two x-labeled nodes are created at most at each iteration.  They are
    ** stored, along with their k values, in these variables.  At each level,
    ** the old nodes are freed and the new nodes are copied into the old map.
    */
    var map [2]*DdNode ;
    var invalidIndex int32  = 1 << (N-1);
    var index [2]int32 = [2]int32{invalidIndex, invalidIndex};

    /* This should never happen. */
    if N < 0 { return(nil); }

    /* If there are no bits, both operands are 0.  The result depends on c. */
    if N == 0 {
        if c >= 0 { return(one);
        } else { return(zero); }
    }

    /* The maximum or the minimum difference comparing to c can generate the terminal case */
    if (1 << N) - 1 < c { return(zero);
    } else if (-(1 << N) + 1) >= c { return(one); }

    /* Build the result bottom up. */
    for i = 1; i <= N; i++ {
        var kTrueLower, kFalseLower int32;
        var leftChild, middleChild, rightChild int32;
        var g0, g1, fplus, fequal, fminus *DdNode;
        var j int32;
        var newMap [2]*DdNode = [2]*DdNode{nil};
        var newIndex [2]int32;

        kTrueLower = kTrue;
        kFalseLower = kFalse;
        /* kTrue = ceiling((c-1)/2^i) + 1 */
        kTrue = ((c-1) >> i) + ((c & mask) != 1) + 1;
        mask = (mask << 1) | 1;
        /* kFalse = floor(c/2^i) - 1 */
        kFalse = (c >> i) - 1;
        newIndex[0] = invalidIndex;
        newIndex[1] = invalidIndex;

        for j = kFalse + 1; j < kTrue; j++ {
            /* Skip if node is not reachable from top of BDD. */
            if (j >= (1 << (N - i))) || (j <= -(1 << (N -i))) { continue; }

            /* Find f- */
            leftChild = (j << 1) - 1;
            if leftChild >= kTrueLower {
                fminus = one;
            } else if leftChild <= kFalseLower {
                fminus = zero;
            } else {
                assert(leftChild == index[0] || leftChild == index[1], "leftChild == index[0] || leftChild == index[1]");
                if leftChild == index[0] {
                    fminus = map[0];
                } else {
                    fminus = map[1];
                }
            }

            /* Find f= */
            middleChild = j << 1;
            if middleChild >= kTrueLower {
                fequal = one;
            } else if middleChild <= kFalseLower {
                fequal = zero;
            } else {
                assert(middleChild == index[0] || middleChild == index[1], "middleChild == index[0] || middleChild == index[1]");
                if middleChild == index[0] {
                    fequal = map[0];
                } else {
                    fequal = map[1];
                }
            }

            /* Find f+ */
            rightChild = (j << 1) + 1;
            if rightChild >= kTrueLower {
                fplus = one;
            } else if rightChild <= kFalseLower {
                fplus = zero;
            } else {
                assert(rightChild == index[0] || rightChild == index[1], "rightChild == index[0] || rightChild == index[1]");
                if rightChild == index[0] {
                    fplus = map[0];
                } else {
                    fplus = map[1];
                }
            }

            /* Build new nodes. */
            g1 = Cudd_bddIte(dd, y[N - i], fequal, fplus);
            if g1 == nil {
                if index[0] != invalidIndex { Cudd_IterDerefBdd(dd, map[0]); }
                if index[1] != invalidIndex { Cudd_IterDerefBdd(dd, map[1]); }
                if newIndex[0] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[0]); }
                if newIndex[1] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[1]); }
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(g1).ref);
            g0 = Cudd_bddIte(dd, y[N - i], fminus, fequal);
            if g0 == nil {
                Cudd_IterDerefBdd(dd, g1);
                if index[0] != invalidIndex { Cudd_IterDerefBdd(dd, map[0]); }
                if index[1] != invalidIndex { Cudd_IterDerefBdd(dd, map[1]); }
                if newIndex[0] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[0]); }
                if newIndex[1] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[1]); }
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(g0).ref);
            f = Cudd_bddIte(dd, x[N - i], g1, g0);
            if f == nil {
                Cudd_IterDerefBdd(dd, g1);
                Cudd_IterDerefBdd(dd, g0);
                if index[0] != invalidIndex { Cudd_IterDerefBdd(dd, map[0]); }
                if index[1] != invalidIndex { Cudd_IterDerefBdd(dd, map[1]); }
                if newIndex[0] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[0]); }
                if newIndex[1] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[1]); }
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(f).ref);
            Cudd_IterDerefBdd(dd, g1);
            Cudd_IterDerefBdd(dd, g0);

            /* Save newly computed node in map. */
            assert(newIndex[0] == invalidIndex || newIndex[1] == invalidIndex, "newIndex[0] == invalidIndex || newIndex[1] == invalidIndex");
            if newIndex[0] == invalidIndex {
                newIndex[0] = j;
                newMap[0] = f;
            } else {
                newIndex[1] = j;
                newMap[1] = f;
            }
        }

        /* Copy new map to map. */
        if index[0] != invalidIndex { Cudd_IterDerefBdd(dd, map[0]); }
        if index[1] != invalidIndex { Cudd_IterDerefBdd(dd, map[1]); }
        map[0] = newMap[0];
        map[1] = newMap[1];
        index[0] = newIndex[0];
        index[1] = newIndex[1];
    }

    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(f).ref);
    return(f);

} /* end of Cudd_Inequality */


/**Function********************************************************************

  Synopsis    [Generates a BDD for the function x - y != c.]

  Description [This function generates a BDD for the function x -y != c.
  Both x and y are N-bit numbers, x\[0\] x\[1\] ... x\[N-1\] and
  y\[0\] y\[1\] ...  y\[N-1\], with 0 the most significant bit.
  The BDD is built bottom-up.
  It has a linear number of nodes if the variables are ordered as follows:
  x\[0\] y\[0\] x\[1\] y\[1\] ... x\[N-1\] y\[N-1\].]

  SideEffects [None]

  SeeAlso     [Cudd_Xgty]

******************************************************************************/
DdNode *
Cudd_Disequality(
  DdManager * dd /* DD manager */,
  int  N /* number of x and y variables */,
  int c /* right-hand side constant */,
  DdNode ** x /* array of x variables */,
  DdNode ** y /* array of y variables */)
{
    /* The nodes at level i represent values of the difference that are
    ** multiples of 2^i.  We use variables with names starting with k
    ** to denote the multipliers of 2^i in such multiples. */
    var kTrueLb int32  = c + 1;
    var kTrueUb int32  = c - 1;
    var kFalse int32  = c;
    /* Mask used to compute the ceiling function.  Since we divide by 2^i,
    ** we want to know whether the dividend is a multiple of 2^i.  If it is,
    ** then ceiling and floor coincide; otherwise, they differ by one. */
    var mask int32  = 1;
    var i int32;

    var f *DdNode  = nil;           /* the eventual result */
    // inlining of macro DD_ONE
    var one *DdNode  = ((dd).one);
    // inlining of macro Cudd_Not
    var zero *DdNode  = ((DdNode *)((ptrint)(one) ^ 01));

    /* Two x-labeled nodes are created at most at each iteration.  They are
    ** stored, along with their k values, in these variables.  At each level,
    ** the old nodes are freed and the new nodes are copied into the old map.
    */
    var map [2]*DdNode ;
    var invalidIndex int32  = 1 << (N-1);
    var index [2]int32 = [2]int32{invalidIndex, invalidIndex};

    /* This should never happen. */
    if N < 0 { return(nil); }

    /* If there are no bits, both operands are 0.  The result depends on c. */
    if N == 0 {
        if c != 0 { return(one);
        } else { return(zero); }
    }

    /* The maximum or the minimum difference comparing to c can generate the terminal case */
    if (1 << N) - 1 < c || (-(1 << N) + 1) > c { return(one); }

    /* Build the result bottom up. */
    for i = 1; i <= N; i++ {
        var kTrueLbLower, kTrueUbLower int32;
        var leftChild, middleChild, rightChild int32;
        var g0, g1, fplus, fequal, fminus *DdNode;
        var j int32;
        var newMap [2]*DdNode = [2]*DdNode{nil};
        var newIndex [2]int32;

        kTrueLbLower = kTrueLb;
        kTrueUbLower = kTrueUb;
        /* kTrueLb = floor((c-1)/2^i) + 2 */
        kTrueLb = ((c-1) >> i) + 2;
        /* kTrueUb = ceiling((c+1)/2^i) - 2 */
        kTrueUb = ((c+1) >> i) + (((c+2) & mask) != 1) - 2;
        mask = (mask << 1) | 1;
        newIndex[0] = invalidIndex;
        newIndex[1] = invalidIndex;

        for j = kTrueUb + 1; j < kTrueLb; j++ {
            /* Skip if node is not reachable from top of BDD. */
            if (j >= (1 << (N - i))) || (j <= -(1 << (N -i))) { continue; }

            /* Find f- */
            leftChild = (j << 1) - 1;
            if leftChild >= kTrueLbLower || leftChild <= kTrueUbLower {
                fminus = one;
            } else if i == 1 && leftChild == kFalse {
                fminus = zero;
            } else {
                assert(leftChild == index[0] || leftChild == index[1], "leftChild == index[0] || leftChild == index[1]");
                if leftChild == index[0] {
                    fminus = map[0];
                } else {
                    fminus = map[1];
                }
            }

            /* Find f= */
            middleChild = j << 1;
            if middleChild >= kTrueLbLower || middleChild <= kTrueUbLower {
                fequal = one;
            } else if i == 1 && middleChild == kFalse {
                fequal = zero;
            } else {
                assert(middleChild == index[0] || middleChild == index[1], "middleChild == index[0] || middleChild == index[1]");
                if middleChild == index[0] {
                    fequal = map[0];
                } else {
                    fequal = map[1];
                }
            }

            /* Find f+ */
            rightChild = (j << 1) + 1;
            if rightChild >= kTrueLbLower || rightChild <= kTrueUbLower {
                fplus = one;
            } else if i == 1 && rightChild == kFalse {
                fplus = zero;
            } else {
                assert(rightChild == index[0] || rightChild == index[1], "rightChild == index[0] || rightChild == index[1]");
                if rightChild == index[0] {
                    fplus = map[0];
                } else {
                    fplus = map[1];
                }
            }

            /* Build new nodes. */
            g1 = Cudd_bddIte(dd, y[N - i], fequal, fplus);
            if g1 == nil {
                if index[0] != invalidIndex { Cudd_IterDerefBdd(dd, map[0]); }
                if index[1] != invalidIndex { Cudd_IterDerefBdd(dd, map[1]); }
                if newIndex[0] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[0]); }
                if newIndex[1] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[1]); }
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(g1).ref);
            g0 = Cudd_bddIte(dd, y[N - i], fminus, fequal);
            if g0 == nil {
                Cudd_IterDerefBdd(dd, g1);
                if index[0] != invalidIndex { Cudd_IterDerefBdd(dd, map[0]); }
                if index[1] != invalidIndex { Cudd_IterDerefBdd(dd, map[1]); }
                if newIndex[0] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[0]); }
                if newIndex[1] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[1]); }
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(g0).ref);
            f = Cudd_bddIte(dd, x[N - i], g1, g0);
            if f == nil {
                Cudd_IterDerefBdd(dd, g1);
                Cudd_IterDerefBdd(dd, g0);
                if index[0] != invalidIndex { Cudd_IterDerefBdd(dd, map[0]); }
                if index[1] != invalidIndex { Cudd_IterDerefBdd(dd, map[1]); }
                if newIndex[0] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[0]); }
                if newIndex[1] != invalidIndex { Cudd_IterDerefBdd(dd, newMap[1]); }
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(f).ref);
            Cudd_IterDerefBdd(dd, g1);
            Cudd_IterDerefBdd(dd, g0);

            /* Save newly computed node in map. */
            assert(newIndex[0] == invalidIndex || newIndex[1] == invalidIndex, "newIndex[0] == invalidIndex || newIndex[1] == invalidIndex");
            if newIndex[0] == invalidIndex {
                newIndex[0] = j;
                newMap[0] = f;
            } else {
                newIndex[1] = j;
                newMap[1] = f;
            }
        }

        /* Copy new map to map. */
        if index[0] != invalidIndex { Cudd_IterDerefBdd(dd, map[0]); }
        if index[1] != invalidIndex { Cudd_IterDerefBdd(dd, map[1]); }
        map[0] = newMap[0];
        map[1] = newMap[1];
        index[0] = newIndex[0];
        index[1] = newIndex[1];
    }

    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(f).ref);
    return(f);

} /* end of Cudd_Disequality */


/**Function********************************************************************

  Synopsis    [Generates a BDD for the function lowerB &le; x &le; upperB.]

  Description [This function generates a BDD for the function
  lowerB &le; x &le; upperB, where x is an N-bit number,
  x\[0\] x\[1\] ... x\[N-1\], with 0 the most significant bit (important!).
  The number of variables N should be sufficient to represent the bounds;
  otherwise, the bounds are truncated to their N least significant bits.
  Two BDDs are built bottom-up for lowerB &le; x and x &le; upperB, and they
  are finally conjoined.]

  SideEffects [None]

  SeeAlso     [Cudd_Xgty]

******************************************************************************/
DdNode *
Cudd_bddInterval(
  DdManager * dd /* DD manager */,
  int  N /* number of x variables */,
  DdNode ** x /* array of x variables */,
  unsigned int lowerB /* lower bound */,
  unsigned int upperB /* upper bound */)
{
    var one, zero *DdNode;
    var r, rl, ru *DdNode;
    var i int32;

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    rl = one;
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(rl).ref);
    ru = one;
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(ru).ref);

    /* Loop to build the rest of the BDDs. */
    for i = N-1; i >= 0; i-- {
        var vl, vu *DdNode;
        vl = Cudd_bddIte(dd, x[i],
                         lowerB&1 ? rl : one,
                         lowerB&1 ? zero : rl);
        if vl == nil {
            Cudd_IterDerefBdd(dd, rl);
            Cudd_IterDerefBdd(dd, ru);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(vl).ref);
        Cudd_IterDerefBdd(dd, rl);
        rl = vl;
        lowerB >>= 1;
        vu = Cudd_bddIte(dd, x[i],
                         upperB&1 ? ru : zero,
                         upperB&1 ? one : ru);
        if vu == nil {
            Cudd_IterDerefBdd(dd, rl);
            Cudd_IterDerefBdd(dd, ru);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(vu).ref);
        Cudd_IterDerefBdd(dd, ru);
        ru = vu;
        upperB >>= 1;
    }

    /* Conjoin the two bounds. */
    r = Cudd_bddAnd(dd, rl, ru);
    if r == nil {
        Cudd_IterDerefBdd(dd, rl);
        Cudd_IterDerefBdd(dd, ru);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(r).ref);
    Cudd_IterDerefBdd(dd, rl);
    Cudd_IterDerefBdd(dd, ru);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(r).ref);
    return(r);

} /* end of Cudd_bddInterval */


/**Function********************************************************************

  Synopsis    [Computes the compatible projection of R w.r.t. cube Y.]

  Description [Computes the compatible projection of relation R with
  respect to cube Y. Returns a pointer to the c-projection if
  successful; NULL otherwise. For a comparison between Cudd_CProjection
  and Cudd_PrioritySelect, see the documentation of the latter.]

  SideEffects [None]

  SeeAlso     [Cudd_PrioritySelect]

******************************************************************************/
func Cudd_CProjection(dd *DdManager, R *DdNode, Y *DdNode) *DdNode {
    var res *DdNode;
    var support *DdNode;

    if cuddCheckCube(dd,Y) == 0 {
        (void) fmt.Fprintf(dd.err, "Error: The third argument of Cudd_CProjection should be a cube\n");
        dd.errorCode = CUDD_INVALID_ARG;
        return(nil);
    }

    /* Compute the support of Y, which is used by the abstraction step
    ** in cuddCProjectionRecur.
    */
    support = Cudd_Support(dd,Y);
    if support == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(support).ref);

    for {
        dd.reordered = 0;
        res = cuddCProjectionRecur(dd,R,Y,support);
    
if !(dd.reordered == 1) { break }
} 

    if res == nil {
        Cudd_RecursiveDeref(dd,support);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref(dd,support);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);

    return(res);

} /* end of Cudd_CProjection */


/**Function********************************************************************

  Synopsis    [Computes the Hamming distance ADD.]

  Description [Computes the Hamming distance ADD. Returns an ADD that
  gives the Hamming distance between its two arguments if successful;
  NULL otherwise. The two vectors xVars and yVars identify the variables
  that form the two arguments.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_addHamming(dd *DdManager, xVars **DdNode, yVars **DdNode, nVars int32) *DdNode {
    var result, tempBdd *DdNode;
    var tempAdd, temp *DdNode;
    var i int32;

    // inlining of macro DD_ZERO
    result = ((dd).zero);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(result).ref);

    for i = 0; i < nVars; i++ {
        // inlining of macro Cudd_Not
        tempBdd = Cudd_bddIte(dd,xVars[i],((DdNode *)((ptrint)(yVars[i]) ^ 01)),yVars[i]);
        if tempBdd == nil {
            Cudd_RecursiveDeref(dd,result);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(tempBdd).ref);
        tempAdd = Cudd_BddToAdd(dd,tempBdd);
        if tempAdd == nil {
            Cudd_RecursiveDeref(dd,tempBdd);
            Cudd_RecursiveDeref(dd,result);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(tempAdd).ref);
        Cudd_RecursiveDeref(dd,tempBdd);
        temp = Cudd_addApply(dd,Cudd_addPlus,tempAdd,result);
        if temp == nil {
            Cudd_RecursiveDeref(dd,tempAdd);
            Cudd_RecursiveDeref(dd,result);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(temp).ref);
        Cudd_RecursiveDeref(dd,tempAdd);
        Cudd_RecursiveDeref(dd,result);
        result = temp;
    }

    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(result).ref);
    return(result);

} /* end of Cudd_addHamming */


/**Function********************************************************************

  Synopsis    [Returns the minimum Hamming distance between f and minterm.]

  Description [Returns the minimum Hamming distance between the
  minterms of a function f and a reference minterm. The function is
  given as a BDD; the minterm is given as an array of integers, one
  for each variable in the manager.  Returns the minimum distance if
  it is less than the upper bound; the upper bound if the minimum
  distance is at least as large; CUDD_OUT_OF_MEM in case of failure.]

  SideEffects [None]

  SeeAlso     [Cudd_addHamming Cudd_bddClosestCube]

******************************************************************************/
int
Cudd_MinHammingDist(
  DdManager *dd /* DD manager */,
  DdNode *f /* function to examine */,
  int *minterm /* reference minterm */,
  int upperBound /* distance above which an approximate answer is OK */)
{
    var table *DdHashTable;
    var epsilon CUDD_VALUE_TYPE;
    var res int32;

    table = cuddHashTableInit(dd,1,2);
    if table == nil {
        return(CUDD_OUT_OF_MEM);
    }
    epsilon = Cudd_ReadEpsilon(dd);
    Cudd_SetEpsilon(dd,(CUDD_VALUE_TYPE)0.0);
    res = cuddMinHammingDistRecur(f,minterm,table,upperBound);
    cuddHashTableQuit(table);
    Cudd_SetEpsilon(dd,epsilon);

    return(res);

} /* end of Cudd_MinHammingDist */


/**Function********************************************************************

  Synopsis    [Finds a cube of f at minimum Hamming distance from g.]

  Description [Finds a cube of f at minimum Hamming distance from the
  minterms of g.  All the minterms of the cube are at the minimum
  distance.  If the distance is 0, the cube belongs to the
  intersection of f and g.  Returns the cube if successful; NULL
  otherwise.]

  SideEffects [The distance is returned as a side effect.]

  SeeAlso     [Cudd_MinHammingDist]

******************************************************************************/
func Cudd_bddClosestCube(dd *DdManager, f *DdNode, g *DdNode, distance *int32) *DdNode {
    var res, acube *DdNode;
    var rdist CUDD_VALUE_TYPE;

    /* Compute the cube and distance as a single ADD. */
    for {
        dd.reordered = 0;
        res = cuddBddClosestCube(dd,f,g,CUDD_CONST_INDEX + 1.0);
    
if !(dd.reordered == 1) { break }
} 
    if res == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);

    /* Unpack distance and cube. */
    for {
        dd.reordered = 0;
        acube = separateCube(dd, res, &rdist);
    
if !(dd.reordered == 1) { break }
} 
    if acube == nil {
        Cudd_RecursiveDeref(dd, res);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(acube).ref);
    Cudd_RecursiveDeref(dd, res);

    /* Convert cube from ADD to BDD. */
    for {
        dd.reordered = 0;
        res = cuddAddBddDoPattern(dd, acube);
    
if !(dd.reordered == 1) { break }
} 
    if res == nil {
        Cudd_RecursiveDeref(dd, acube);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref(dd, acube);

    *distance = int32( rdist);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of Cudd_bddClosestCube */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_CProjection.]

  Description [Performs the recursive step of Cudd_CProjection. Returns
  the projection if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_CProjection]

******************************************************************************/
func cuddCProjectionRecur(dd *DdManager, R *DdNode, Y *DdNode, Ysupp *DdNode) *DdNode {
    var res, res1, res2, resA *DdNode;
    var r, y, RT, RE, YT, YE, Yrest, Ra, Ran, Gamma, Alpha *DdNode;
    unsigned int topR, topY, top, index = 0;
    // inlining of macro DD_ONE
    var one *DdNode  = ((dd).one);

    statLine(dd);
    if Y == one { return(R); }

               
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    assert(!((((DdNode *)((ptruint)(Y) & ~01))).index == CUDD_CONST_INDEX), "!/*__starting_point_inline_macro:Cudd_IsConstant__*/((/*__starting_point_inline_macro:Cudd_Regular__*/((DdNode *)((ptruint)(Y) & ~01))).index == CUDD_CONST_INDEX)");
      

    // inlining of macro Cudd_Not
    if R == ((DdNode *)((ptrint)(one) ^ 01)) { return(R); }

    res = cuddCacheLookup2(dd, Cudd_CProjection, R, Y);
    if res != nil { return(res); }

    // inlining of macro Cudd_Regular
    r = ((DdNode *)((ptruint)(R) & ~01));
    // inlining of macro cuddI
    
    if (r.index)==CUDD_CONST_INDEX {
        topR = int32((r.index));
    } else {
        topR = (dd).perm[(r.index)];
    }
    // inlining of macro Cudd_Regular
    y = ((DdNode *)((ptruint)(Y) & ~01));
    // inlining of macro cuddI
    
    if (y.index)==CUDD_CONST_INDEX {
        topY = int32((y.index));
    } else {
        topY = (dd).perm[(y.index)];
    }

    // inlining of macro ddMin
    
    if (topY) < (topR) {
        top = (topY);
    } else {
        top = (topR);
    }

    /* Compute the cofactors of R */
    if topR == top {
        index = r.index;
        // inlining of macro cuddT
        RT = ((r).type.kids.T);
        // inlining of macro cuddE
        RE = ((r).type.kids.E);
        if r != R {
            // inlining of macro Cudd_Not
            RT = ((DdNode *)((ptrint)(RT) ^ 01)); RE = ((DdNode *)((ptrint)(RE) ^ 01));
        }
    } else {
        RT = RE = R;
    }

    if topY > top {
        /* Y does not depend on the current top variable.
        ** We just need to compute the results on the two cofactors of R
        ** and make them the children of a node labeled r->index.
        */
        res1 = cuddCProjectionRecur(dd,RT,Y,Ysupp);
        if res1 == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        res2 = cuddCProjectionRecur(dd,RE,Y,Ysupp);
        if res2 == nil {
            Cudd_RecursiveDeref(dd,res1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res2).ref);
        res = cuddBddIteRecur(dd, dd.vars[index], res1, res2);
        if res == nil {
            Cudd_RecursiveDeref(dd,res1);
            Cudd_RecursiveDeref(dd,res2);
            return(nil);
        }
        /* If we have reached this point, res1 and res2 are now
        ** incorporated in res. cuddDeref is therefore sufficient.
        */
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res1).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res2).ref);
    } else {
        /* Compute the cofactors of Y */
        index = y.index;
        // inlining of macro cuddT
        YT = ((y).type.kids.T);
        // inlining of macro cuddE
        YE = ((y).type.kids.E);
        if y != Y {
            // inlining of macro Cudd_Not
            YT = ((DdNode *)((ptrint)(YT) ^ 01)); YE = ((DdNode *)((ptrint)(YE) ^ 01));
        }
        // inlining of macro Cudd_Not
        if YT == ((DdNode *)((ptrint)(one) ^ 01)) {
            // inlining of macro Cudd_Not
            Alpha  = ((DdNode *)((ptrint)(dd.vars[index]) ^ 01));
            Yrest = YE;
            Ra = RE;
            Ran = RT;
        } else {
            Alpha = dd.vars[index];
            Yrest = YT;
            Ra = RT;
            Ran = RE;
        }
        // inlining of macro cuddT
        Gamma = cuddBddExistAbstractRecur(dd,Ra,((Ysupp).type.kids.T));
        if Gamma == nil { return(nil); }
        if Gamma == one {
            // inlining of macro cuddT
            res1 = cuddCProjectionRecur(dd,Ra,Yrest,((Ysupp).type.kids.T));
            if res1 == nil { return(nil); }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res1).ref);
            res = cuddBddAndRecur(dd, Alpha, res1);
            if res == nil {
                Cudd_RecursiveDeref(dd,res1);
                return(nil);
            }
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(res1).ref);
        // inlining of macro Cudd_Not
        } else if Gamma == ((DdNode *)((ptrint)(one) ^ 01)) {
            // inlining of macro cuddT
            res1 = cuddCProjectionRecur(dd,Ran,Yrest,((Ysupp).type.kids.T));
            if res1 == nil { return(nil); }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res1).ref);
            // inlining of macro Cudd_Not
            res = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(Alpha) ^ 01)), res1);
            if res == nil {
                Cudd_RecursiveDeref(dd,res1);
                return(nil);
            }
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(res1).ref);
        } else {
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(Gamma).ref);
            // inlining of macro cuddT
            resA = cuddCProjectionRecur(dd,Ran,Yrest,((Ysupp).type.kids.T));
            if resA == nil {
                Cudd_RecursiveDeref(dd,Gamma);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(resA).ref);
            // inlining of macro Cudd_Not
            res2 = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(Gamma) ^ 01)), resA);
            if res2 == nil {
                Cudd_RecursiveDeref(dd,Gamma);
                Cudd_RecursiveDeref(dd,resA);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res2).ref);
            Cudd_RecursiveDeref(dd,Gamma);
            Cudd_RecursiveDeref(dd,resA);
            // inlining of macro cuddT
            res1 = cuddCProjectionRecur(dd,Ra,Yrest,((Ysupp).type.kids.T));
            if res1 == nil {
                Cudd_RecursiveDeref(dd,res2);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res1).ref);
            res = cuddBddIteRecur(dd, Alpha, res1, res2);
            if res == nil {
                Cudd_RecursiveDeref(dd,res1);
                Cudd_RecursiveDeref(dd,res2);
                return(nil);
            }
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(res1).ref);
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(res2).ref);
        }
    }

    cuddCacheInsert2(dd,Cudd_CProjection,R,Y,res);

    return(res);

} /* end of cuddCProjectionRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddClosestCube.]

  Description [Performs the recursive step of Cudd_bddClosestCube.
  Returns the cube if succesful; NULL otherwise.  The procedure uses a
  four-way recursion to examine all four combinations of cofactors of
  <code>f</code> and <code>g</code> according to the following formula.
  <pre>
    H(f,g) = min(H(ft,gt), H(fe,ge), H(ft,ge)+1, H(fe,gt)+1)
  </pre>
  Bounding is based on the following observations.
  <ul>
  <li> If we already found two points at distance 0, there is no point in
       continuing.  Furthermore,
  <li> If F == not(G) then the best we can hope for is a minimum distance
       of 1.  If we have already found two points at distance 1, there is
       no point in continuing.  (Indeed, H(F,G) == 1 in this case.  We
       have to continue, though, to find the cube.)
  </ul>
  The variable <code>bound</code> is set at the largest value of the distance
  that we are still interested in.  Therefore, we desist when
  <pre>
    (bound == -1) and (F != not(G)) or (bound == 0) and (F == not(G)).
  </pre>
  If we were maximally aggressive in using the bound, we would always
  set the bound to the minimum distance seen thus far minus one.  That
  is, we would maintain the invariant
  <pre>
    bound < minD,
  </pre>
  except at the very beginning, when we have no value for
  <code>minD</code>.<p>

  However, we do not use <code>bound < minD</code> when examining the
  two negative cofactors, because we try to find a large cube at
  minimum distance.  To do so, we try to find a cube in the negative
  cofactors at the same or smaller distance from the cube found in the
  positive cofactors.<p>

  When we compute <code>H(ft,ge)</code> and <code>H(fe,gt)</code> we
  know that we are going to add 1 to the result of the recursive call
  to account for the difference in the splitting variable.  Therefore,
  we decrease the bound correspondingly.<p>

  Another important observation concerns the need of examining all
  four pairs of cofators only when both <code>f</code> and
  <code>g</code> depend on the top variable.<p>

  Suppose <code>gt == ge == g</code>.  (That is, <code>g</code> does
  not depend on the top variable.)  Then
  <pre>
    H(f,g) = min(H(ft,g), H(fe,g), H(ft,g)+1, H(fe,g)+1)
           = min(H(ft,g), H(fe,g)) .
  </pre>
  Therefore, under these circumstances, we skip the two "cross" cases.<p>

  An interesting feature of this function is the scheme used for
  caching the results in the global computed table.  Since we have a
  cube and a distance, we combine them to form an ADD.  The
  combination replaces the zero child of the top node of the cube with
  the negative of the distance.  (The use of the negative is to avoid
  ambiguity with 1.)  The degenerate cases (zero and one) are treated
  specially because the distance is known (0 for one, and infinity for
  zero).]

  SideEffects [None]

  SeeAlso     [Cudd_bddClosestCube]

******************************************************************************/
func cuddBddClosestCube(dd *DdManager, f *DdNode, g *DdNode, bound CUDD_VALUE_TYPE) *DdNode {
    var res, F, G, ft, fe, gt, ge, tt, ee *DdNode;
    var ctt, cee, cte, cet *DdNode;
    var minD, dtt, dee, dte, det CUDD_VALUE_TYPE;
    // inlining of macro DD_ONE
    var one *DdNode  = ((dd).one);
    // inlining of macro Cudd_Not
    var lzero *DdNode  = ((DdNode *)((ptrint)(one) ^ 01));
    // inlining of macro DD_ZERO
    var azero *DdNode  = ((dd).zero);
    unsigned int topf, topg, index;

    statLine(dd);
    // inlining of macro Cudd_Not
    if bound < int32((f == ((DdNode *)((ptrint)(g) ^ 01)))) { return(azero); }
    /* Terminal cases. */
    if g == lzero || f == lzero { return(azero); }
    if f == one && g == one { return(one); }

    /* Check cache. */
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    if F.ref != 1 || G.ref != 1 {
        res = cuddCacheLookup2(dd,(DD_CTFP) Cudd_bddClosestCube, f, g);
        if res != nil { return(res); }
    }

    // inlining of macro cuddI
    
    if (F.index)==CUDD_CONST_INDEX {
        topf = int32((F.index));
    } else {
        topf = (dd).perm[(F.index)];
    }
    // inlining of macro cuddI
    
    if (G.index)==CUDD_CONST_INDEX {
        topg = int32((G.index));
    } else {
        topg = (dd).perm[(G.index)];
    }

    /* Compute cofactors. */
    if topf <= topg {
        index = F.index;
        // inlining of macro cuddT
        ft = ((F).type.kids.T);
        // inlining of macro cuddE
        fe = ((F).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (f) & 01)) != 0 {
            // inlining of macro Cudd_Not
            ft = ((DdNode *)((ptrint)(ft) ^ 01));
            // inlining of macro Cudd_Not
            fe = ((DdNode *)((ptrint)(fe) ^ 01));
        }
    } else {
        index = G.index;
        ft = fe = f;
    }

    if topg <= topf {
        // inlining of macro cuddT
        gt = ((G).type.kids.T);
        // inlining of macro cuddE
        ge = ((G).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (g) & 01)) != 0 {
            // inlining of macro Cudd_Not
            gt = ((DdNode *)((ptrint)(gt) ^ 01));
            // inlining of macro Cudd_Not
            ge = ((DdNode *)((ptrint)(ge) ^ 01));
        }
    } else {
        gt = ge = g;
    }

    tt = cuddBddClosestCube(dd,ft,gt,bound);
    if tt == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(tt).ref);
    ctt = separateCube(dd,tt,&dtt);
    if ctt == nil {
        Cudd_RecursiveDeref(dd, tt);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(ctt).ref);
    Cudd_RecursiveDeref(dd, tt);
    minD = dtt;
    // inlining of macro ddMin
    
    if (minD) < (bound) {
        bound = (minD);
    } else {
        bound = (bound);
    }

    ee = cuddBddClosestCube(dd,fe,ge,bound);
    if ee == nil {
        Cudd_RecursiveDeref(dd, ctt);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(ee).ref);
    cee = separateCube(dd,ee,&dee);
    if cee == nil {
        Cudd_RecursiveDeref(dd, ctt);
        Cudd_RecursiveDeref(dd, ee);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(cee).ref);
    Cudd_RecursiveDeref(dd, ee);
    // inlining of macro ddMin
    
    if (dee) < (dtt) {
        minD = (dee);
    } else {
        minD = (dtt);
    }
    // inlining of macro ddMin
    if minD <= CUDD_CONST_INDEX { 
    if (minD-1) < (bound) {
        bound = (minD-1);
    } else {
        bound = (bound);
    } }

    if minD > 0 && topf == topg {
        var te *DdNode  = cuddBddClosestCube(dd,ft,ge,bound-1);
        if te == nil {
            Cudd_RecursiveDeref(dd, ctt);
            Cudd_RecursiveDeref(dd, cee);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(te).ref);
        cte = separateCube(dd,te,&dte);
        if cte == nil {
            Cudd_RecursiveDeref(dd, ctt);
            Cudd_RecursiveDeref(dd, cee);
            Cudd_RecursiveDeref(dd, te);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(cte).ref);
        Cudd_RecursiveDeref(dd, te);
        dte += 1.0;
        // inlining of macro ddMin
        
        if (dte) < (minD) {
            minD = (dte);
        } else {
            minD = (minD);
        }
    } else {
        cte = azero;
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(cte).ref);
        dte = CUDD_CONST_INDEX + 1.0;
    }
    // inlining of macro ddMin
    if minD <= CUDD_CONST_INDEX { 
    if (minD-1) < (bound) {
        bound = (minD-1);
    } else {
        bound = (bound);
    } }

    if minD > 0 && topf == topg {
        var et *DdNode  = cuddBddClosestCube(dd,fe,gt,bound-1);
        if et == nil {
            Cudd_RecursiveDeref(dd, ctt);
            Cudd_RecursiveDeref(dd, cee);
            Cudd_RecursiveDeref(dd, cte);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(et).ref);
        cet = separateCube(dd,et,&det);
        if cet == nil {
            Cudd_RecursiveDeref(dd, ctt);
            Cudd_RecursiveDeref(dd, cee);
            Cudd_RecursiveDeref(dd, cte);
            Cudd_RecursiveDeref(dd, et);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(cet).ref);
        Cudd_RecursiveDeref(dd, et);
        det += 1.0;
        // inlining of macro ddMin
        
        if (det) < (minD) {
            minD = (det);
        } else {
            minD = (minD);
        }
    } else {
        cet = azero;
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(cet).ref);
        det = CUDD_CONST_INDEX + 1.0;
    }

    if minD == dtt {
        if dtt == dee && ctt == cee {
            res = createResult(dd,CUDD_CONST_INDEX,1,ctt,dtt);
        } else {
            res = createResult(dd,index,1,ctt,dtt);
        }
    } else if minD == dee {
        res = createResult(dd,index,0,cee,dee);
    } else if minD == dte {
               
        assert(topf == topg, "topf == topg");
      
        res = createResult(dd,index,1,cte,dte);
    } else {
               
        assert(topf == topg, "topf == topg");
      
        res = createResult(dd,index,0,cet,det);
    }
    if res == nil {
        Cudd_RecursiveDeref(dd, ctt);
        Cudd_RecursiveDeref(dd, cee);
        Cudd_RecursiveDeref(dd, cte);
        Cudd_RecursiveDeref(dd, cet);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref(dd, ctt);
    Cudd_RecursiveDeref(dd, cee);
    Cudd_RecursiveDeref(dd, cte);
    Cudd_RecursiveDeref(dd, cet);

    /* Only cache results that are different from azero to avoid
    ** storing results that depend on the value of the bound. */
    if (F.ref != 1 || G.ref != 1) && res != azero {
        cuddCacheInsert2(dd,(DD_CTFP) Cudd_bddClosestCube, f, g, res); }

    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of cuddBddClosestCube */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_MinHammingDist.]

  Description [Performs the recursive step of Cudd_MinHammingDist.
  It is based on the following identity. Let H(f) be the
  minimum Hamming distance of the minterms of f from the reference
  minterm. Then:
  <xmp>
    H(f) = min(H(f0)+h0,H(f1)+h1)
  </xmp>
  where f0 and f1 are the two cofactors of f with respect to its top
  variable; h0 is 1 if the minterm assigns 1 to the top variable of f;
  h1 is 1 if the minterm assigns 0 to the top variable of f.
  The upper bound on the distance is used to bound the depth of the
  recursion.
  Returns the minimum distance unless it exceeds the upper bound or
  computation fails.]

  SideEffects [None]

  SeeAlso     [Cudd_MinHammingDist]

******************************************************************************/
func cuddMinHammingDistRecur(f *DdNode, minterm *int32, table *DdHashTable, upperBound int32) int32 {
    var F, Ft, Fe *DdNode;
    var h, hT, hE float64;
    var zero, res *DdNode;
    var dd *DdManager  = table.manager;

    statLine(dd);
    if upperBound == 0 { return(0); }

    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));

    // inlining of macro cuddIsConstant
    if (F).index == CUDD_CONST_INDEX {
        // inlining of macro Cudd_Not
        // inlining of macro DD_ONE
        zero = ((DdNode *)((ptrint)(((dd).one)) ^ 01));
        if f == dd.background || f == zero {
            return(upperBound);
        } else {
            return(0);
        }
    }
    if (res = cuddHashTableLookup1(table,f)) != nil {
        // inlining of macro cuddV
        h = ((res).type.value);
        if res.ref == 0 {
            dd.dead++;
            dd.constants.dead++;
        }
        return(int32( h));
    }

    // inlining of macro cuddT
    // inlining of macro cuddE
    Ft = ((F).type.kids.T); Fe = ((F).type.kids.E);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (f) & 01)) != 0 {
        // inlining of macro Cudd_Not
        Ft = ((DdNode *)((ptrint)(Ft) ^ 01)); Fe = ((DdNode *)((ptrint)(Fe) ^ 01));
    }
    if minterm[F.index] == 0 {
        var temp *DdNode  = Ft;
        Ft = Fe; Fe = temp;
    }

    hT = cuddMinHammingDistRecur(Ft,minterm,table,upperBound);
    if hT == CUDD_OUT_OF_MEM { return(CUDD_OUT_OF_MEM); }
    if hT == 0 {
        hE = upperBound;
    } else {
        hE = cuddMinHammingDistRecur(Fe,minterm,table,upperBound - 1);
        if hE == CUDD_OUT_OF_MEM { return(CUDD_OUT_OF_MEM); }
    }
    // inlining of macro ddMin
    
    if (hE + 1) < (hT) {
        h = (hE + 1);
    } else {
        h = (hT);
    }

    if F.ref != 1 {
        var fanout ptrint  = (ptrint) F.ref;
        cuddSatDec(fanout);
        res = cuddUniqueConst(dd, (CUDD_VALUE_TYPE) h);
        if !cuddHashTableInsert1(table,f,res,fanout) {
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res).ref); Cudd_RecursiveDeref(dd, res);
            return(CUDD_OUT_OF_MEM);
        }
    }

    return(int32( h));

} /* end of cuddMinHammingDistRecur */


/**Function********************************************************************

  Synopsis    [Separates cube from distance.]

  Description [Separates cube from distance.  Returns the cube if
  successful; NULL otherwise.]

  SideEffects [The distance is returned as a side effect.]

  SeeAlso     [cuddBddClosestCube createResult]

******************************************************************************/
func separateCube(dd *DdManager, f *DdNode, distance *CUDD_VALUE_TYPE) *DdNode {
    var cube, t *DdNode;

    /* One and zero are special cases because the distance is implied. */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {
        // inlining of macro DD_ONE
        *distance = (f == ((dd).one)) ? 0.0 :
            (1.0 + (CUDD_VALUE_TYPE) CUDD_CONST_INDEX);
        return(f);
    }

    /* Find out which branch points to the distance and replace the top
    ** node with one pointing to zero instead. */
    // inlining of macro cuddT
    t = ((f).type.kids.T);
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    // inlining of macro cuddV
    if ((((DdNode *)((ptruint)(t) & ~01))).index == CUDD_CONST_INDEX) && ((t).type.value) <= 0 {
               
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        // inlining of macro cuddE
        // inlining of macro DD_ONE
        assert(!((((DdNode *)((ptruint)(((f).type.kids.E)) & ~01))).index == CUDD_CONST_INDEX) || ((f).type.kids.E) == ((dd).one), "!/*__starting_point_inline_macro:Cudd_IsConstant__*/((/*__starting_point_inline_macro:Cudd_Regular__*/((DdNode *)((ptruint)(/*__starting_point_inline_macro:cuddE__*/((f).type.kids.E)) & ~01))).index == CUDD_CONST_INDEX) || /*__starting_point_inline_macro:cuddE__*/((f).type.kids.E) == /*__starting_point_inline_macro:DD_ONE__*/((dd).one)");
      
        // inlining of macro cuddV
        *distance = -((t).type.value);
        // inlining of macro DD_ZERO
        // inlining of macro cuddE
        cube = cuddUniqueInter(dd, f.index, ((dd).zero), ((f).type.kids.E));
    } else {
               
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        // inlining of macro DD_ONE
        assert(!((((DdNode *)((ptruint)(t) & ~01))).index == CUDD_CONST_INDEX) || t == ((dd).one), "!/*__starting_point_inline_macro:Cudd_IsConstant__*/((/*__starting_point_inline_macro:Cudd_Regular__*/((DdNode *)((ptruint)(t) & ~01))).index == CUDD_CONST_INDEX) || t == /*__starting_point_inline_macro:DD_ONE__*/((dd).one)");
      
        // inlining of macro cuddV
        // inlining of macro cuddE
        *distance = -((((f).type.kids.E)).type.value);
        // inlining of macro DD_ZERO
        cube = cuddUniqueInter(dd, f.index, t, ((dd).zero));
    }

    return(cube);

} /* end of separateCube */


/**Function********************************************************************

  Synopsis    [Builds a result for cache storage.]

  Description [Builds a result for cache storage.  Returns a pointer
  to the resulting ADD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [cuddBddClosestCube separateCube]

******************************************************************************/
static DdNode *
createResult(
  DdManager *dd,
  unsigned int index,
  unsigned int phase,
  DdNode *cube,
  CUDD_VALUE_TYPE distance)
{
    var res, constant *DdNode;

    /* Special case.  The cube is either one or zero, and we do not
    ** add any variables.  Hence, the result is also one or zero,
    ** and the distance remains implied by the value of the constant. */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if index == CUDD_CONST_INDEX && ((((DdNode *)((ptruint)(cube) & ~01))).index == CUDD_CONST_INDEX) { return(cube); }

    constant = cuddUniqueConst(dd,-distance);
    if constant == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(constant).ref);

    if index == CUDD_CONST_INDEX {
        /* Replace the top node. */
        // inlining of macro cuddT
        // inlining of macro DD_ZERO
        if ((cube).type.kids.T) == ((dd).zero) {
            // inlining of macro cuddE
            res = cuddUniqueInter(dd,cube.index,constant,((cube).type.kids.E));
        } else {
            // inlining of macro cuddT
            res = cuddUniqueInter(dd,cube.index,((cube).type.kids.T),constant);
        }
    } else {
        /* Add a new top node. */
               
        // inlining of macro cuddI
        assert((((index)==CUDD_CONST_INDEX)?int32((index)):(dd).perm[(index)]) < (((cube.index)==CUDD_CONST_INDEX)?int32((cube.index)):(dd).perm[(cube.index)]), "/*__starting_point_inline_macro:cuddI__*/(((index)==CUDD_CONST_INDEX)?int32((index)):(dd).perm[(index)]) < /*__starting_point_inline_macro:cuddI__*/(((cube.index)==CUDD_CONST_INDEX)?int32((cube.index)):(dd).perm[(cube.index)])");
      
        if phase {
            res = cuddUniqueInter(dd,index,cube,constant);
        } else {
            res = cuddUniqueInter(dd,index,constant,cube);
        }
    }
    if res == nil {
        Cudd_RecursiveDeref(dd, constant);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(constant).ref); /* safe because constant is part of res */

    return(res);

} /* end of createResult */


//ABC_NAMESPACE_IMPL_END


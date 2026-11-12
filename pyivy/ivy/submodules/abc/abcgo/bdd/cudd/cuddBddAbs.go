package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddBddAbs.c]

  PackageName [cudd]

  Synopsis    [Quantification functions for BDDs.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_bddExistAbstract()
                <li> Cudd_bddXorExistAbstract()
                <li> Cudd_bddUnivAbstract()
                <li> Cudd_bddBooleanDiff()
                <li> Cudd_bddVarIsDependent()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddBddExistAbstractRecur()
                <li> cuddBddXorExistAbstractRecur()
                <li> cuddBddBooleanDiffRecur()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> bddCheckPositiveCube()
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
static char rcsid[] DD_UNUSED = "$Id: cuddBddAbs.c,v 1.26 2004/08/13 18:04:46 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int bddCheckPositiveCube (DdManager *manager, DdNode *cube);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Existentially abstracts all the variables in cube from f.]

  Description [Existentially abstracts all the variables in cube from f.
  Returns the abstracted BDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddUnivAbstract Cudd_addExistAbstract]

******************************************************************************/
func Cudd_bddExistAbstract(manager *DdManager, f *DdNode, cube *DdNode) *DdNode {
    var res *DdNode;

    if bddCheckPositiveCube(manager, cube) == 0 {
        (void) fmt.Fprintf(manager.err, "Error: Can only abstract positive cubes\n");
        manager.errorCode = CUDD_INVALID_ARG;
        return(nil);
    }

    for {
        manager.reordered = 0;
        res = cuddBddExistAbstractRecur(manager, f, cube);
    
if !(manager.reordered == 1) { break }
} 

    return(res);

} /* end of Cudd_bddExistAbstract */


/**Function********************************************************************

  Synopsis [Takes the exclusive OR of two BDDs and simultaneously abstracts the
  variables in cube.]

  Description [Takes the exclusive OR of two BDDs and simultaneously abstracts
  the variables in cube. The variables are existentially abstracted.  Returns a
  pointer to the result is successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddUnivAbstract Cudd_bddExistAbstract Cudd_bddAndAbstract]

******************************************************************************/
func Cudd_bddXorExistAbstract(manager *DdManager, f *DdNode, g *DdNode, cube *DdNode) *DdNode {
    var res *DdNode;

    if bddCheckPositiveCube(manager, cube) == 0 {
        (void) fmt.Fprintf(manager.err, "Error: Can only abstract positive cubes\n");
        manager.errorCode = CUDD_INVALID_ARG;
        return(nil);
    }

    for {
        manager.reordered = 0;
        res = cuddBddXorExistAbstractRecur(manager, f, g, cube);
    
if !(manager.reordered == 1) { break }
} 

    return(res);

} /* end of Cudd_bddXorExistAbstract */


/**Function********************************************************************

  Synopsis    [Universally abstracts all the variables in cube from f.]

  Description [Universally abstracts all the variables in cube from f.
  Returns the abstracted BDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddExistAbstract Cudd_addUnivAbstract]

******************************************************************************/
func Cudd_bddUnivAbstract(manager *DdManager, f *DdNode, cube *DdNode) *DdNode {
    var res *DdNode;

    if bddCheckPositiveCube(manager, cube) == 0 {
        (void) fmt.Fprintf(manager.err, "Error: Can only abstract positive cubes\n");
        manager.errorCode = CUDD_INVALID_ARG;
        return(nil);
    }

    for {
        manager.reordered = 0;
        // inlining of macro Cudd_Not
        res = cuddBddExistAbstractRecur(manager, ((DdNode *)((ptrint)(f) ^ 01)), cube);
    
if !(manager.reordered == 1) { break }
} 
    // inlining of macro Cudd_Not
    if res != nil { res = ((DdNode *)((ptrint)(res) ^ 01)); }

    return(res);

} /* end of Cudd_bddUnivAbstract */


/**Function********************************************************************

  Synopsis    [Computes the boolean difference of f with respect to x.]

  Description [Computes the boolean difference of f with respect to the
  variable with index x.  Returns the BDD of the boolean difference if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_bddBooleanDiff(manager *DdManager, f *DdNode, x int32) *DdNode {
    var res, var *DdNode;

    /* If the variable is not currently in the manager, f cannot
    ** depend on it.
    */
    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    if x >= manager.size { return(((DdNode *)((ptrint)(((manager).one)) ^ 01))); }
    var = manager.vars[x];

    for {
        manager.reordered = 0;
        // inlining of macro Cudd_Regular
        res = cuddBddBooleanDiffRecur(manager, ((DdNode *)((ptruint)(f) & ~01)), var);
    
if !(manager.reordered == 1) { break }
} 

    return(res);

} /* end of Cudd_bddBooleanDiff */


/**Function********************************************************************

  Synopsis    [Checks whether a variable is dependent on others in a
  function.]

  Description [Checks whether a variable is dependent on others in a
  function.  Returns 1 if the variable is dependent; 0 otherwise. No
  new nodes are created.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
int
Cudd_bddVarIsDependent(
  DdManager *dd,                /* manager */
  DdNode *f,                    /* function */
  DdNode *var                   /* variable */)
{
    var F, res, zero, ft, fe *DdNode;
    var topf, level uint32;
    DD_CTFP cacheOp;
    var retval int32;

    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    zero = ((DdNode *)((ptrint)(((dd).one)) ^ 01));
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX { return(f == zero); }

    /* From now on f is not constant. */
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    topf = uint32( dd.perm[F.index]);
    level = uint32( dd.perm[var.index]);

    /* Check terminal case. If topf > index of var, f does not depend on var.
    ** Therefore, var is not dependent in f. */
    if topf > level {
        return(0);
    }

    cacheOp = (DD_CTFP) Cudd_bddVarIsDependent;
    res = cuddCacheLookup2(dd,cacheOp,f,var);
    if res != nil {
        return(res != zero);
    }

    /* Compute cofactors. */
    // inlining of macro Cudd_NotCond
    // inlining of macro cuddT
    ft = ((DdNode *)((ptrint)(((F).type.kids.T)) ^ (f != F)));
    // inlining of macro Cudd_NotCond
    // inlining of macro cuddE
    fe = ((DdNode *)((ptrint)(((F).type.kids.E)) ^ (f != F)));

    if topf == level {
        // inlining of macro Cudd_Not
        retval = Cudd_bddLeq(dd,ft,((DdNode *)((ptrint)(fe) ^ 01)));
    } else {
        retval = Cudd_bddVarIsDependent(dd,ft,var) &&
            Cudd_bddVarIsDependent(dd,fe,var);
    }

    // inlining of macro Cudd_NotCond
    cuddCacheInsert2(dd,cacheOp,f,var,((DdNode *)((ptrint)(zero) ^ (retval))));

    return(retval);

} /* Cudd_bddVarIsDependent */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive steps of Cudd_bddExistAbstract.]

  Description [Performs the recursive steps of Cudd_bddExistAbstract.
  Returns the BDD obtained by abstracting the variables
  of cube from f if successful; NULL otherwise. It is also used by
  Cudd_bddUnivAbstract.]

  SideEffects [None]

  SeeAlso     [Cudd_bddExistAbstract Cudd_bddUnivAbstract]

******************************************************************************/
func cuddBddExistAbstractRecur(manager *DdManager, f *DdNode, cube *DdNode) *DdNode {
    var F, T, E, res, res1, res2, one *DdNode;

    statLine(manager);
    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));

    /* Cube is guaranteed to be a cube at this point. */        
    if cube == one || F == one {  
        return(f);
    }
    /* From now on, f and cube are non-constant. */

    /* Abstract a variable that does not appear in f. */
    for manager.perm[F.index] > manager.perm[cube.index] {
        // inlining of macro cuddT
        cube = ((cube).type.kids.T);
        if cube == one { return(f); }
    }

    /* Check the cache. */
    if F.ref != 1 && (res = cuddCacheLookup2(manager, Cudd_bddExistAbstract, f, cube)) != nil {
        return(res);
    }

    /* Compute the cofactors of f. */
    // inlining of macro cuddT
    // inlining of macro cuddE
    T = ((F).type.kids.T); E = ((F).type.kids.E);
    if f != F {
        // inlining of macro Cudd_Not
        T = ((DdNode *)((ptrint)(T) ^ 01)); E = ((DdNode *)((ptrint)(E) ^ 01));
    }

    /* If the two indices are the same, so are their levels. */
    if F.index == cube.index {
        // inlining of macro Cudd_Not
        if T == one || E == one || T == ((DdNode *)((ptrint)(E) ^ 01)) {
            return(one);
        }
        // inlining of macro cuddT
        res1 = cuddBddExistAbstractRecur(manager, T, ((cube).type.kids.T));
        if res1 == nil { return(nil); }
        if res1 == one {
            if F.ref != 1 {
                cuddCacheInsert2(manager, Cudd_bddExistAbstract, f, cube, one); }
            return(one);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        // inlining of macro cuddT
        res2 = cuddBddExistAbstractRecur(manager, E, ((cube).type.kids.T));
        if res2 == nil {
            Cudd_IterDerefBdd(manager,res1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res2).ref);
        // inlining of macro Cudd_Not
        res = cuddBddAndRecur(manager, ((DdNode *)((ptrint)(res1) ^ 01)), ((DdNode *)((ptrint)(res2) ^ 01)));
        if res == nil {
            Cudd_IterDerefBdd(manager, res1);
            Cudd_IterDerefBdd(manager, res2);
            return(nil);
        }
        // inlining of macro Cudd_Not
        res = ((DdNode *)((ptrint)(res) ^ 01));
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_IterDerefBdd(manager, res1);
        Cudd_IterDerefBdd(manager, res2);
        if F.ref != 1 {
            cuddCacheInsert2(manager, Cudd_bddExistAbstract, f, cube, res); }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref);
        return(res);
    } else { /* if (cuddI(manager,F->index) < cuddI(manager,cube->index)) */
        res1 = cuddBddExistAbstractRecur(manager, T, cube);
        if res1 == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        res2 = cuddBddExistAbstractRecur(manager, E, cube);
        if res2 == nil {
            Cudd_IterDerefBdd(manager, res1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res2).ref);
        /* ITE takes care of possible complementation of res1 and of the
        ** case in which res1 == res2. */
        res = cuddBddIteRecur(manager, manager.vars[F.index], res1, res2);
        if res == nil {
            Cudd_IterDerefBdd(manager, res1);
            Cudd_IterDerefBdd(manager, res2);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref); //Added
        Cudd_IterDerefBdd(manager, res1); //cuddDeref(res1);
        Cudd_IterDerefBdd(manager, res2); //cuddDeref(res2);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref); //Added
        if F.ref != 1 {
            cuddCacheInsert2(manager, Cudd_bddExistAbstract, f, cube, res); }
        return(res);
    }       

} /* end of cuddBddExistAbstractRecur */


/**Function********************************************************************

  Synopsis [Takes the exclusive OR of two BDDs and simultaneously abstracts the
  variables in cube.]

  Description [Takes the exclusive OR of two BDDs and simultaneously abstracts
  the variables in cube. The variables are existentially abstracted.  Returns a
  pointer to the result is successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddAndAbstract]

******************************************************************************/
func cuddBddXorExistAbstractRecur(manager *DdManager, f *DdNode, g *DdNode, cube *DdNode) *DdNode {
    var F, fv, fnv, G, gv, gnv *DdNode;
    var one, zero, r, t, e, Cube *DdNode;
    unsigned int topf, topg, topcube, top, index;

    statLine(manager);
    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Terminal cases. */
    if f == g {
        return(zero);
    }
    // inlining of macro Cudd_Not
    if f == ((DdNode *)((ptrint)(g) ^ 01)) {
        return(one);
    }
    if cube == one {
        return(cuddBddXorRecur(manager, f, g));
    }
    if f == one {
        // inlining of macro Cudd_Not
        return(cuddBddExistAbstractRecur(manager, ((DdNode *)((ptrint)(g) ^ 01)), cube));
    }
    if g == one {
        // inlining of macro Cudd_Not
        return(cuddBddExistAbstractRecur(manager, ((DdNode *)((ptrint)(f) ^ 01)), cube));
    }
    if f == zero {
        return(cuddBddExistAbstractRecur(manager, g, cube));
    }
    if g == zero {
        return(cuddBddExistAbstractRecur(manager, f, cube));
    }

    /* At this point f, g, and cube are not constant. */

    // inlining of macro cuddF2L
    // inlining of macro Cudd_Regular
    // inlining of macro Cudd_IsComplement
    if ((((DdNode *)((ptruint)(f) & ~01)).Id << 1) | (int32( ((ptrint) (f) & 01)))) > ((((DdNode *)((ptruint)(g) & ~01)).Id << 1) | (int32( ((ptrint) (g) & 01)))) { /* Try to increase cache efficiency. */
        var tmp *DdNode  = f;
        f = g;
        g = tmp;
    }

    /* Check cache. */
    r = cuddCacheLookup(manager, DD_BDD_XOR_EXIST_ABSTRACT_TAG, f, g, cube);
    if r != nil {
        return(r);
    }

    /* Here we can skip the use of cuddI, because the operands are known
    ** to be non-constant.
    */
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    topf = manager.perm[F.index];
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    topg = manager.perm[G.index];
    // inlining of macro ddMin
    
    if (topg) < (topf) {
        top = (topg);
    } else {
        top = (topf);
    }
    topcube = manager.perm[cube.index];

    if topcube < top {
        // inlining of macro cuddT
        return(cuddBddXorExistAbstractRecur(manager, f, g, ((cube).type.kids.T)));
    }
    /* Now, topcube >= top. */

    if topf == top {
        index = F.index;
        // inlining of macro cuddT
        fv = ((F).type.kids.T);
        // inlining of macro cuddE
        fnv = ((F).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (f) & 01)) != 0 {
            // inlining of macro Cudd_Not
            fv = ((DdNode *)((ptrint)(fv) ^ 01));
            // inlining of macro Cudd_Not
            fnv = ((DdNode *)((ptrint)(fnv) ^ 01));
        }
    } else {
        index = G.index;
        fv = fnv = f;
    }

    if topg == top {
        // inlining of macro cuddT
        gv = ((G).type.kids.T);
        // inlining of macro cuddE
        gnv = ((G).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (g) & 01)) != 0 {
            // inlining of macro Cudd_Not
            gv = ((DdNode *)((ptrint)(gv) ^ 01));
            // inlining of macro Cudd_Not
            gnv = ((DdNode *)((ptrint)(gnv) ^ 01));
        }
    } else {
        gv = gnv = g;
    }

    if topcube == top {
        // inlining of macro cuddT
        Cube = ((cube).type.kids.T);
    } else {
        Cube = cube;
    }

    t = cuddBddXorExistAbstractRecur(manager, fv, gv, Cube);
    if t == nil { return(nil); }

    /* Special case: 1 OR anything = 1. Hence, no need to compute
    ** the else branch if t is 1.
    */
    if t == one && topcube == top {
        cuddCacheInsert(manager, DD_BDD_XOR_EXIST_ABSTRACT_TAG, f, g, cube, one);
        return(one);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = cuddBddXorExistAbstractRecur(manager, fnv, gnv, Cube);
    if e == nil {
        Cudd_IterDerefBdd(manager, t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    if topcube == top {       /* abstract */
        // inlining of macro Cudd_Not
        r = cuddBddAndRecur(manager, ((DdNode *)((ptrint)(t) ^ 01)), ((DdNode *)((ptrint)(e) ^ 01)));
        if r == nil {
            Cudd_IterDerefBdd(manager, t);
            Cudd_IterDerefBdd(manager, e);
            return(nil);
        }
        // inlining of macro Cudd_Not
        r = ((DdNode *)((ptrint)(r) ^ 01));
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(r).ref);
        Cudd_IterDerefBdd(manager, t);
        Cudd_IterDerefBdd(manager, e);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(r).ref);
    } else if t == e {
        r = t;
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
    } else {
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (t) & 01)) != 0 {
            // inlining of macro Cudd_Not
            r = cuddUniqueInter(manager,int32(index),((DdNode *)((ptrint)(t) ^ 01)),((DdNode *)((ptrint)(e) ^ 01)));
            if r == nil {
                Cudd_IterDerefBdd(manager, t);
                Cudd_IterDerefBdd(manager, e);
                return(nil);
            }
            // inlining of macro Cudd_Not
            r = ((DdNode *)((ptrint)(r) ^ 01));
        } else {
            r = cuddUniqueInter(manager,int32(index),t,e);
            if r == nil {
                Cudd_IterDerefBdd(manager, t);
                Cudd_IterDerefBdd(manager, e);
                return(nil);
            }
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
    }
    cuddCacheInsert(manager, DD_BDD_XOR_EXIST_ABSTRACT_TAG, f, g, cube, r);
    return (r);

} /* end of cuddBddXorExistAbstractRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive steps of Cudd_bddBoleanDiff.]

  Description [Performs the recursive steps of Cudd_bddBoleanDiff.
  Returns the BDD obtained by XORing the cofactors of f with respect to
  var if successful; NULL otherwise. Exploits the fact that dF/dx =
  dF'/dx.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddBddBooleanDiffRecur(manager *DdManager, f *DdNode, var *DdNode) *DdNode {
    var T, E, res, res1, res2 *DdNode;

    statLine(manager);
    // inlining of macro cuddI
    if (((f.index)==CUDD_CONST_INDEX)?int32((f.index)):(manager).perm[(f.index)]) > manager.perm[var.index] {
        /* f does not depend on var. */
        // inlining of macro Cudd_Not
        // inlining of macro DD_ONE
        return(((DdNode *)((ptrint)(((manager).one)) ^ 01)));
    }

    /* From now on, f is non-constant. */

    /* If the two indices are the same, so are their levels. */
    if f.index == var.index {
        // inlining of macro cuddT
        // inlining of macro cuddE
        res = cuddBddXorRecur(manager, ((f).type.kids.T), ((f).type.kids.E));
        return(res);
    }

    /* From now on, cuddI(manager,f->index) < cuddI(manager,cube->index). */

    /* Check the cache. */
    res = cuddCacheLookup2(manager, cuddBddBooleanDiffRecur, f, var);
    if res != nil {
        return(res);
    }

    /* Compute the cofactors of f. */
    // inlining of macro cuddT
    // inlining of macro cuddE
    T = ((f).type.kids.T); E = ((f).type.kids.E);

    res1 = cuddBddBooleanDiffRecur(manager, T, var);
    if res1 == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res1).ref);
    // inlining of macro Cudd_Regular
    res2 = cuddBddBooleanDiffRecur(manager, ((DdNode *)((ptruint)(E) & ~01)), var);
    if res2 == nil {
        Cudd_IterDerefBdd(manager, res1);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res2).ref);
    /* ITE takes care of possible complementation of res1 and of the
    ** case in which res1 == res2. */
    res = cuddBddIteRecur(manager, manager.vars[f.index], res1, res2);
    if res == nil {
        Cudd_IterDerefBdd(manager, res1);
        Cudd_IterDerefBdd(manager, res2);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res1).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res2).ref);
    cuddCacheInsert2(manager, cuddBddBooleanDiffRecur, f, var, res);
    return(res);

} /* end of cuddBddBooleanDiffRecur */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis [Checks whether cube is an BDD representing the product of
  positive literals.]

  Description [Returns 1 in case of success; 0 otherwise.]

  SideEffects [None]

******************************************************************************/
func bddCheckPositiveCube(manager *DdManager, cube *DdNode) int32 {
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (cube) & 01)) != 0 { return(0); }
    // inlining of macro DD_ONE
    if cube == ((manager).one) { return(1); }
    // inlining of macro cuddIsConstant
    if (cube).index == CUDD_CONST_INDEX { return(0); }
    // inlining of macro cuddE
    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    if ((cube).type.kids.E) == ((DdNode *)((ptrint)(((manager).one)) ^ 01)) {
        // inlining of macro cuddT
        return(bddCheckPositiveCube(manager, ((cube).type.kids.T)));
    }
    return(0);

} /* end of bddCheckPositiveCube */


//ABC_NAMESPACE_IMPL_END



package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddSat.c]

  PackageName [cudd]

  Synopsis    [Functions for the solution of satisfiability related
  problems.]

  Description [External procedures included in this file:
                <ul>
                <li> Cudd_Eval()
                <li> Cudd_ShortestPath()
                <li> Cudd_LargestCube()
                <li> Cudd_ShortestLength()
                <li> Cudd_Decreasing()
                <li> Cudd_Increasing()
                <li> Cudd_EquivDC()
                <li> Cudd_bddLeqUnless()
                <li> Cudd_EqualSupNorm()
                <li> Cudd_bddMakePrime()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddBddMakePrime()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> freePathPair()
                <li> getShortest()
                <li> getPath()
                <li> getLargest()
                <li> getCube()
                </ul>]

  Author      [Seh-Woong Jeong, Fabio Somenzi]

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

const DD_BIGGY =        1000000

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

typedef struct cuddPathPair {
    int pos;
    int neg;
}
 cuddPathPair;

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddSat.c,v 1.36 2009/03/08 02:49:02 fabio Exp $";
#endif

var one, zero *DdNode;

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

// Original macro:
// #define WEIGHT(weight, col)     ((weight) == NULL ? 1 : weight[col])


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static      st__retval freePathPair (char *key, char *value, char *arg);

// static cuddPathPair getShortest (DdNode *root, int *cost, int *support, st__table *visited);

// static DdNode * getPath (DdManager *manager, st__table *visited, DdNode *f, int *weight, int cost);

// static cuddPathPair getLargest (DdNode *root, st__table *visited);

// static DdNode * getCube (DdManager *manager, st__table *visited, DdNode *f, int cost);


/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Returns the value of a DD for a given variable assignment.]

  Description [Finds the value of a DD for a given variable
  assignment. The variable assignment is passed in an array of int's,
  that should specify a zero or a one for each variable in the support
  of the function. Returns a pointer to a constant node. No new nodes
  are produced.]

  SideEffects [None]

  SeeAlso     [Cudd_bddLeq Cudd_addEvalConst]

******************************************************************************/
func Cudd_Eval(dd *DdManager, f *DdNode, inputs *int32) *DdNode {
    var comple int32;
    var ptr *DdNode;

    // inlining of macro Cudd_IsComplement
    comple = (int32( ((ptrint) (f) & 01)));
    // inlining of macro Cudd_Regular
    ptr = ((DdNode *)((ptruint)(f) & ~01));

    // inlining of macro cuddIsConstant
    for !((ptr).index == CUDD_CONST_INDEX) {
        if inputs[ptr.index] == 1 {
            // inlining of macro cuddT
            ptr = ((ptr).type.kids.T);
        } else {
            // inlining of macro Cudd_IsComplement
            // inlining of macro cuddE
            comple ^= (int32( ((ptrint) (((ptr).type.kids.E)) & 01)));
            // inlining of macro Cudd_Regular
            // inlining of macro cuddE
            ptr = ((DdNode *)((ptruint)(((ptr).type.kids.E)) & ~01));
        }
    }
    // inlining of macro Cudd_NotCond
    return(((DdNode *)((ptrint)(ptr) ^ (comple))));

} /* end of Cudd_Eval */


/**Function********************************************************************

  Synopsis    [Finds a shortest path in a DD.]

  Description [Finds a shortest path in a DD. f is the DD we want to
  get the shortest path for; weight\[i\] is the weight of the THEN arc
  coming from the node whose index is i. If weight is NULL, then unit
  weights are assumed for all THEN arcs. All ELSE arcs have 0 weight.
  If non-NULL, both weight and support should point to arrays with at
  least as many entries as there are variables in the manager.
  Returns the shortest path as the BDD of a cube.]

  SideEffects [support contains on return the true support of f.
  If support is NULL on entry, then Cudd_ShortestPath does not compute
  the true support info. length contains the length of the path.]

  SeeAlso     [Cudd_ShortestLength Cudd_LargestCube]

******************************************************************************/
func Cudd_ShortestPath(manager *DdManager, f *DdNode, weight *int32, support *int32, length *int32) *DdNode {
    var F *DdNode;
    var visited *st__table;
    var sol *DdNode;
    var rootPair *cuddPathPair;
    var complement, cost int32;
    var i int32;

    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro DD_ZERO
    zero = ((manager).zero);

    /* Initialize support. Support does not depend on variable order.
    ** Hence, it does not need to be reinitialized if reordering occurs.
    */
    if support != nil {
      for i = 0; i < manager.size; i++ {
        support[i] = 0;
      }
    }

    // inlining of macro Cudd_Not
    if f == ((DdNode *)((ptrint)(one) ^ 01)) || f == zero {
      *length = DD_BIGGY;
      // inlining of macro Cudd_Not
      return(((DdNode *)((ptrint)(one) ^ 01)));
    }
    /* From this point on, a path exists. */

    for {
        manager.reordered = 0;

        /* Initialize visited table. */
        visited = st__init_table( st__ptrcmp, st__ptrhash);

        /* Now get the length of the shortest path(s) from f to 1. */
        (void) getShortest(f, weight, support, visited);

        // inlining of macro Cudd_IsComplement
        complement = (int32( ((ptrint) (f) & 01)));

        // inlining of macro Cudd_Regular
        F = ((DdNode *)((ptruint)(f) & ~01));

        if ! st__lookup(visited, (char *)F, (char **)&rootPair) { return(nil); }

        if complement != 0 {
          cost = rootPair.neg;
        } else {
          cost = rootPair.pos;
        }

        /* Recover an actual shortest path. */
        sol = getPath(manager,visited,f,weight,cost);

        st__foreach(visited, freePathPair, nil);
        st__free_table(visited);

    
if !(manager.reordered == 1) { break }
} 

    *length = cost;
    return(sol);

} /* end of Cudd_ShortestPath */


/**Function********************************************************************

  Synopsis    [Finds a largest cube in a DD.]

  Description [Finds a largest cube in a DD. f is the DD we want to
  get the largest cube for. The problem is translated into the one of
  finding a shortest path in f, when both THEN and ELSE arcs are assumed to
  have unit length. This yields a largest cube in the disjoint cover
  corresponding to the DD. Therefore, it is not necessarily the largest
  implicant of f.  Returns the largest cube as a BDD.]

  SideEffects [The number of literals of the cube is returned in length.]

  SeeAlso     [Cudd_ShortestPath]

******************************************************************************/
func Cudd_LargestCube(manager *DdManager, f *DdNode, length *int32) *DdNode {
    register    DdNode  *F;
    var visited *st__table;
    var sol *DdNode;
    var rootPair *cuddPathPair;
    var complement, cost int32;

    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro DD_ZERO
    zero = ((manager).zero);

    // inlining of macro Cudd_Not
    if f == ((DdNode *)((ptrint)(one) ^ 01)) || f == zero {
        *length = DD_BIGGY;
        // inlining of macro Cudd_Not
        return(((DdNode *)((ptrint)(one) ^ 01)));
    }
    /* From this point on, a path exists. */

    for {
        manager.reordered = 0;

        /* Initialize visited table. */
        visited = st__init_table( st__ptrcmp, st__ptrhash);

        /* Now get the length of the shortest path(s) from f to 1. */
        (void) getLargest(f, visited);

        // inlining of macro Cudd_IsComplement
        complement = (int32( ((ptrint) (f) & 01)));

        // inlining of macro Cudd_Regular
        F = ((DdNode *)((ptruint)(f) & ~01));

        if ! st__lookup(visited, (char *)F, (char **)&rootPair) { return(nil); }

        if complement != 0 {
          cost = rootPair.neg;
        } else {
          cost = rootPair.pos;
        }

        /* Recover an actual shortest path. */
        sol = getCube(manager,visited,f,cost);

        st__foreach(visited, freePathPair, nil);
        st__free_table(visited);

    
if !(manager.reordered == 1) { break }
} 

    *length = cost;
    return(sol);

} /* end of Cudd_LargestCube */


/**Function********************************************************************

  Synopsis    [Find the length of the shortest path(s) in a DD.]

  Description [Find the length of the shortest path(s) in a DD. f is
  the DD we want to get the shortest path for; weight\[i\] is the
  weight of the THEN edge coming from the node whose index is i. All
  ELSE edges have 0 weight. Returns the length of the shortest
  path(s) if such a path is found; a large number if the function is
  identically 0, and CUDD_OUT_OF_MEM in case of failure.]

  SideEffects [None]

  SeeAlso     [Cudd_ShortestPath]

******************************************************************************/
func Cudd_ShortestLength(manager *DdManager, f *DdNode, weight *int32) int32 {
    register    DdNode  *F;
    var visited *st__table;
    var my_pair *cuddPathPair;
    var complement, cost int32;

    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro DD_ZERO
    zero = ((manager).zero);

    // inlining of macro Cudd_Not
    if f == ((DdNode *)((ptrint)(one) ^ 01)) || f == zero {
        return(DD_BIGGY);
    }

    /* From this point on, a path exists. */
    /* Initialize visited table and support. */
    visited = st__init_table( st__ptrcmp, st__ptrhash);

    /* Now get the length of the shortest path(s) from f to 1. */
    (void) getShortest(f, weight, nil, visited);

    // inlining of macro Cudd_IsComplement
    complement = (int32( ((ptrint) (f) & 01)));

    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));

    if ! st__lookup(visited, (char *)F, (char **)&my_pair) { return(CUDD_OUT_OF_MEM); }
    
    if complement != 0 {
        cost = my_pair.neg;
    } else {
        cost = my_pair.pos;
    }

    st__foreach(visited, freePathPair, nil);
    st__free_table(visited);

    return(cost);

} /* end of Cudd_ShortestLength */


/**Function********************************************************************

  Synopsis    [Determines whether a BDD is negative unate in a
  variable.]

  Description [Determines whether the function represented by BDD f is
  negative unate (monotonic decreasing) in variable i. Returns the
  constant one is f is unate and the (logical) constant zero if it is not.
  This function does not generate any new nodes.]

  SideEffects [None]

  SeeAlso     [Cudd_Increasing]

******************************************************************************/
func Cudd_Decreasing(dd *DdManager, f *DdNode, i int32) *DdNode {
    unsigned int topf, level;
    var F, fv, fvn, res *DdNode;
    DD_CTFP cacheOp;

    statLine(dd);
#ifdef DD_DEBUG
    assert(0 <= i && i < dd->size);
#endif

    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro cuddI
    
    if (F.index)==CUDD_CONST_INDEX {
        topf = int32((F.index));
    } else {
        topf = (dd).perm[(F.index)];
    }

    /* Check terminal case. If topf > i, f does not depend on var.
    ** Therefore, f is unate in i.
    */
    level = uint32( dd.perm[i]);
    if topf > level {
        // inlining of macro DD_ONE
        return(((dd).one));
    }

    /* From now on, f is not constant. */

    /* Check cache. */
    cacheOp = (DD_CTFP) Cudd_Decreasing;
    res = cuddCacheLookup2(dd,cacheOp,f,dd.vars[i]);
    if res != nil {
        return(res);
    }

    /* Compute cofactors. */
    // inlining of macro cuddT
    // inlining of macro cuddE
    fv = ((F).type.kids.T); fvn = ((F).type.kids.E);
    if F != f {
        // inlining of macro Cudd_Not
        fv = ((DdNode *)((ptrint)(fv) ^ 01));
        // inlining of macro Cudd_Not
        fvn = ((DdNode *)((ptrint)(fvn) ^ 01));
    }

    if topf == uint32( level) {
        /* Special case: if fv is regular, fv(1,...,1) = 1;
        ** If in addition fvn is complemented, fvn(1,...,1) = 0.
        ** But then f(1,1,...,1) > f(0,1,...,1). Hence f is not
        ** monotonic decreasing in i.
        */
        // inlining of macro Cudd_IsComplement
        if !((int32( ((ptrint) (fv) & 01)) != 0)) && ((int32( ((ptrint) (fvn) & 01)) != 0)) {
            // inlining of macro Cudd_Not
            // inlining of macro DD_ONE
            return(((DdNode *)((ptrint)(((dd).one)) ^ 01)));
        }
        if Cudd_bddLeq(dd,fv,fvn) {
            // inlining of macro DD_ONE
            res = ((dd).one);
        } else {
            // inlining of macro Cudd_Not
            // inlining of macro DD_ONE
            res = ((DdNode *)((ptrint)(((dd).one)) ^ 01));
        }
    } else {
        res = Cudd_Decreasing(dd,fv,i);
        // inlining of macro DD_ONE
        if res == ((dd).one) {
            res = Cudd_Decreasing(dd,fvn,i);
        }
    }

    cuddCacheInsert2(dd,cacheOp,f,dd.vars[i],res);
    return(res);

} /* end of Cudd_Decreasing */


/**Function********************************************************************

  Synopsis    [Determines whether a BDD is positive unate in a
  variable.]

  Description [Determines whether the function represented by BDD f is
  positive unate (monotonic increasing) in variable i. It is based on
  Cudd_Decreasing and the fact that f is monotonic increasing in i if
  and only if its complement is monotonic decreasing in i.]

  SideEffects [None]

  SeeAlso     [Cudd_Decreasing]

******************************************************************************/
func Cudd_Increasing(dd *DdManager, f *DdNode, i int32) *DdNode {
    // inlining of macro Cudd_Not
    return(Cudd_Decreasing(dd,((DdNode *)((ptrint)(f) ^ 01)),i));

} /* end of Cudd_Increasing */


/**Function********************************************************************

  Synopsis    [Tells whether F and G are identical wherever D is 0.]

  Description [Tells whether F and G are identical wherever D is 0.  F
  and G are either two ADDs or two BDDs.  D is either a 0-1 ADD or a
  BDD.  The function returns 1 if F and G are equivalent, and 0
  otherwise.  No new nodes are created.]

  SideEffects [None]

  SeeAlso     [Cudd_bddLeqUnless]

******************************************************************************/
func Cudd_EquivDC(dd *DdManager, F *DdNode, G *DdNode, D *DdNode) int32 {
    var tmp, One, Gr, Dr *DdNode;
    var Fv, Fvn, Gv, Gvn, Dv, Dvn *DdNode;
    var res int32;
    unsigned int flevel, glevel, dlevel, top;

    // inlining of macro DD_ONE
    One = ((dd).one);

    statLine(dd);
    /* Check terminal cases. */
    if D == One || F == G { return(1); }
    // inlining of macro Cudd_Not
    // inlining of macro DD_ZERO
    if D == ((DdNode *)((ptrint)(One) ^ 01)) || D == ((dd).zero) || F == ((DdNode *)((ptrint)(G) ^ 01)) { return(0); }

    /* From now on, D is non-constant. */

    /* Normalize call to increase cache efficiency. */
    if F > G {
        tmp = F;
        F = G;
        G = tmp;
    }
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (F) & 01)) != 0 {
        // inlining of macro Cudd_Not
        F = ((DdNode *)((ptrint)(F) ^ 01));
        // inlining of macro Cudd_Not
        G = ((DdNode *)((ptrint)(G) ^ 01));
    }

    /* From now on, F is regular. */

    /* Check cache. */
    tmp = cuddCacheLookup(dd,DD_EQUIV_DC_TAG,F,G,D);
    if tmp != nil { return(tmp == One); }

    /* Find splitting variable. */
    // inlining of macro cuddI
    
    if (F.index)==CUDD_CONST_INDEX {
        flevel = int32((F.index));
    } else {
        flevel = (dd).perm[(F.index)];
    }
    // inlining of macro Cudd_Regular
    Gr = ((DdNode *)((ptruint)(G) & ~01));
    // inlining of macro cuddI
    
    if (Gr.index)==CUDD_CONST_INDEX {
        glevel = int32((Gr.index));
    } else {
        glevel = (dd).perm[(Gr.index)];
    }
    // inlining of macro ddMin
    
    if (glevel) < (flevel) {
        top = (glevel);
    } else {
        top = (flevel);
    }
    // inlining of macro Cudd_Regular
    Dr = ((DdNode *)((ptruint)(D) & ~01));
    dlevel = dd.perm[Dr.index];
    // inlining of macro ddMin
    
    if (dlevel) < (top) {
        top = (dlevel);
    } else {
        top = (top);
    }

    /* Compute cofactors. */
    if top == flevel {
        // inlining of macro cuddT
        Fv = ((F).type.kids.T);
        // inlining of macro cuddE
        Fvn = ((F).type.kids.E);
    } else {
        Fv = Fvn = F;
    }
    if top == glevel {
        // inlining of macro cuddT
        Gv = ((Gr).type.kids.T);
        // inlining of macro cuddE
        Gvn = ((Gr).type.kids.E);
        if G != Gr {
            // inlining of macro Cudd_Not
            Gv = ((DdNode *)((ptrint)(Gv) ^ 01));
            // inlining of macro Cudd_Not
            Gvn = ((DdNode *)((ptrint)(Gvn) ^ 01));
        }
    } else {
        Gv = Gvn = G;
    }
    if top == dlevel {
        // inlining of macro cuddT
        Dv = ((Dr).type.kids.T);
        // inlining of macro cuddE
        Dvn = ((Dr).type.kids.E);
        if D != Dr {
            // inlining of macro Cudd_Not
            Dv = ((DdNode *)((ptrint)(Dv) ^ 01));
            // inlining of macro Cudd_Not
            Dvn = ((DdNode *)((ptrint)(Dvn) ^ 01));
        }
    } else {
        Dv = Dvn = D;
    }

    /* Solve recursively. */
    res = Cudd_EquivDC(dd,Fv,Gv,Dv);
    if res != 0 {
        res = Cudd_EquivDC(dd,Fvn,Gvn,Dvn);
    }
    // inlining of macro Cudd_Not
    cuddCacheInsert(dd,DD_EQUIV_DC_TAG,F,G,D,(res) ? One : ((DdNode *)((ptrint)(One) ^ 01)));

    return(res);

} /* end of Cudd_EquivDC */


/**Function********************************************************************

  Synopsis    [Tells whether f is less than of equal to G unless D is 1.]

  Description [Tells whether f is less than of equal to G unless D is
  1.  f, g, and D are BDDs.  The function returns 1 if f is less than
  of equal to G, and 0 otherwise.  No new nodes are created.]

  SideEffects [None]

  SeeAlso     [Cudd_EquivDC Cudd_bddLeq Cudd_bddIteConstant]

******************************************************************************/
func Cudd_bddLeqUnless(dd *DdManager, f *DdNode, g *DdNode, D *DdNode) int32 {
    var tmp, One, F, G *DdNode;
    var Ft, Fe, Gt, Ge, Dt, De *DdNode;
    var res int32;
    unsigned int flevel, glevel, dlevel, top;

    statLine(dd);

    // inlining of macro DD_ONE
    One = ((dd).one);

    /* Check terminal cases. */
    // inlining of macro Cudd_Not
    if f == g || g == One || f == ((DdNode *)((ptrint)(One) ^ 01)) || D == One ||
        // inlining of macro Cudd_Not
        D == f || D == ((DdNode *)((ptrint)(g) ^ 01)) { return(1); }
    /* Check for two-operand cases. */
    // inlining of macro Cudd_Not
    if D == ((DdNode *)((ptrint)(One) ^ 01)) || D == g || D == ((DdNode *)((ptrint)(f) ^ 01)) {
        return(Cudd_bddLeq(dd,f,g)); }
    // inlining of macro Cudd_Not
    if g == ((DdNode *)((ptrint)(One) ^ 01)) || g == ((DdNode *)((ptrint)(f) ^ 01)) { return(Cudd_bddLeq(dd,f,D)); }
    // inlining of macro Cudd_Not
    if f == One { return(Cudd_bddLeq(dd,((DdNode *)((ptrint)(g) ^ 01)),D)); }

    /* From now on, f, g, and D are non-constant, distinct, and
    ** non-complementary. */

    /* Normalize call to increase cache efficiency.  We rely on the
    ** fact that f <= g unless D is equivalent to not(g) <= not(f)
    ** unless D and to f <= D unless g.  We make sure that D is
    ** regular, and that at most one of f and g is complemented.  We also
    ** ensure that when two operands can be swapped, the one with the
    ** lowest address comes first. */

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (D) & 01)) != 0 {
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (g) & 01)) != 0 {
            /* Special case: if f is regular and g is complemented,
            ** f(1,...,1) = 1 > 0 = g(1,...,1).  If D(1,...,1) = 0, return 0.
            */
            // inlining of macro Cudd_IsComplement
            if !((int32( ((ptrint) (f) & 01)) != 0)) { return(0); }
            /* !g <= D unless !f  or  !D <= g unless !f */
            tmp = D;
            // inlining of macro Cudd_Not
            D = ((DdNode *)((ptrint)(f) ^ 01));
            if g < tmp {
                // inlining of macro Cudd_Not
                f = ((DdNode *)((ptrint)(g) ^ 01));
                g = tmp;
            } else {
                // inlining of macro Cudd_Not
                f = ((DdNode *)((ptrint)(tmp) ^ 01));
            }
        } else {
            // inlining of macro Cudd_IsComplement
            if int32( ((ptrint) (f) & 01)) != 0 {
                /* !D <= !f unless g  or  !D <= g unless !f */
                tmp = f;
                // inlining of macro Cudd_Not
                f = ((DdNode *)((ptrint)(D) ^ 01));
                if tmp < g {
                    D = g;
                    // inlining of macro Cudd_Not
                    g = ((DdNode *)((ptrint)(tmp) ^ 01));
                } else {
                    // inlining of macro Cudd_Not
                    D = ((DdNode *)((ptrint)(tmp) ^ 01));
                }
            } else {
                /* f <= D unless g  or  !D <= !f unless g */
                tmp = D;
                D = g;
                if tmp < f {
                    // inlining of macro Cudd_Not
                    g = ((DdNode *)((ptrint)(f) ^ 01));
                    // inlining of macro Cudd_Not
                    f = ((DdNode *)((ptrint)(tmp) ^ 01));
                } else {
                    g = tmp;
                }
            }
        }
    } else {
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (g) & 01)) != 0 {
            // inlining of macro Cudd_IsComplement
            if int32( ((ptrint) (f) & 01)) != 0 {
                /* !g <= !f unless D  or  !g <= D unless !f */
                tmp = f;
                // inlining of macro Cudd_Not
                f = ((DdNode *)((ptrint)(g) ^ 01));
                if D < tmp {
                    g = D;
                    // inlining of macro Cudd_Not
                    D = ((DdNode *)((ptrint)(tmp) ^ 01));
                } else {
                    // inlining of macro Cudd_Not
                    g = ((DdNode *)((ptrint)(tmp) ^ 01));
                }
            } else {
                /* f <= g unless D  or  !g <= !f unless D */
                if g < f {
                    tmp = g;
                    // inlining of macro Cudd_Not
                    g = ((DdNode *)((ptrint)(f) ^ 01));
                    // inlining of macro Cudd_Not
                    f = ((DdNode *)((ptrint)(tmp) ^ 01));
                }
            }
        } else {
            /* f <= g unless D  or  f <= D unless g */
            if D < g {
                tmp = D;
                D = g;
                g = tmp;
            }
        }
    }

    /* From now on, D is regular. */

    /* Check cache. */
    tmp = cuddCacheLookup(dd,DD_BDD_LEQ_UNLESS_TAG,f,g,D);
    if tmp != nil { return(tmp == One); }

    /* Find splitting variable. */
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    flevel = dd.perm[F.index];
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    glevel = dd.perm[G.index];
    // inlining of macro ddMin
    
    if (glevel) < (flevel) {
        top = (glevel);
    } else {
        top = (flevel);
    }
    dlevel = dd.perm[D.index];
    // inlining of macro ddMin
    
    if (dlevel) < (top) {
        top = (dlevel);
    } else {
        top = (top);
    }

    /* Compute cofactors. */
    if top == flevel {
        // inlining of macro cuddT
        Ft = ((F).type.kids.T);
        // inlining of macro cuddE
        Fe = ((F).type.kids.E);
        if F != f {
            // inlining of macro Cudd_Not
            Ft = ((DdNode *)((ptrint)(Ft) ^ 01));
            // inlining of macro Cudd_Not
            Fe = ((DdNode *)((ptrint)(Fe) ^ 01));
        }
    } else {
        Ft = Fe = f;
    }
    if top == glevel {
        // inlining of macro cuddT
        Gt = ((G).type.kids.T);
        // inlining of macro cuddE
        Ge = ((G).type.kids.E);
        if G != g {
            // inlining of macro Cudd_Not
            Gt = ((DdNode *)((ptrint)(Gt) ^ 01));
            // inlining of macro Cudd_Not
            Ge = ((DdNode *)((ptrint)(Ge) ^ 01));
        }
    } else {
        Gt = Ge = g;
    }
    if top == dlevel {
        // inlining of macro cuddT
        Dt = ((D).type.kids.T);
        // inlining of macro cuddE
        De = ((D).type.kids.E);
    } else {
        Dt = De = D;
    }

    /* Solve recursively. */
    res = Cudd_bddLeqUnless(dd,Ft,Gt,Dt);
    if res != 0 {
        res = Cudd_bddLeqUnless(dd,Fe,Ge,De);
    }
    // inlining of macro Cudd_NotCond
    cuddCacheInsert(dd,DD_BDD_LEQ_UNLESS_TAG,f,g,D,((DdNode *)((ptrint)(One) ^ (!res))));

    return(res);

} /* end of Cudd_bddLeqUnless */


/**Function********************************************************************

  Synopsis    [Compares two ADDs for equality within tolerance.]

  Description [Compares two ADDs for equality within tolerance. Two
  ADDs are reported to be equal if the maximum difference between them
  (the sup norm of their difference) is less than or equal to the
  tolerance parameter. Returns 1 if the two ADDs are equal (within
  tolerance); 0 otherwise. If parameter <code>pr</code> is positive
  the first failure is reported to the standard output.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
int
Cudd_EqualSupNorm(
  DdManager * dd /* manager */,
  DdNode * f /* first ADD */,
  DdNode * g /* second ADD */,
  CUDD_VALUE_TYPE  tolerance /* maximum allowed difference */,
  int  pr /* verbosity level */)
{
    var fv, fvn, gv, gvn, r *DdNode;
    unsigned int topf, topg;

    statLine(dd);
    /* Check terminal cases. */
    if f == g { return(1); }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if ((((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX) && ((((DdNode *)((ptruint)(g) & ~01))).index == CUDD_CONST_INDEX) {
        // inlining of macro ddEqualVal
        // inlining of macro ddAbs
        // inlining of macro cuddV
        if ((((((f).type.value))-(((g).type.value)))<0) ? -((((f).type.value))-(((g).type.value))) : ((((f).type.value))-(((g).type.value))))<(tolerance) {
            return(1);
        } else {
            if pr>0 {
                (void) fmt.Fprintf(dd.out, "Offending nodes:\n");
                // inlining of macro cuddV
                (void) fmt.Fprintf(dd.out, "f: address = %p\t value = %40.30f\n", (void *) f, ((f).type.value));
                // inlining of macro cuddV
                (void) fmt.Fprintf(dd.out, "g: address = %p\t value = %40.30f\n", (void *) g, ((g).type.value));
            }
            return(0);
        }
    }

    /* We only insert the result in the cache if the comparison is
    ** successful. Therefore, if we hit we return 1. */
    r = cuddCacheLookup2(dd,(DD_CTFP)Cudd_EqualSupNorm,f,g);
    if r != nil {
        return(1);
    }

    /* Compute the cofactors and solve the recursive subproblems. */
    // inlining of macro cuddI
    
    if (f.index)==CUDD_CONST_INDEX {
        topf = int32((f.index));
    } else {
        topf = (dd).perm[(f.index)];
    }
    // inlining of macro cuddI
    
    if (g.index)==CUDD_CONST_INDEX {
        topg = int32((g.index));
    } else {
        topg = (dd).perm[(g.index)];
    }

    // inlining of macro cuddT
    // inlining of macro cuddE
    if topf <= topg {fv = ((f).type.kids.T); fvn = ((f).type.kids.E);} else {fv = fvn = f;}
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topg <= topf {gv = ((g).type.kids.T); gvn = ((g).type.kids.E);} else {gv = gvn = g;}

    if !Cudd_EqualSupNorm(dd,fv,gv,tolerance,pr) { return(0); }
    if !Cudd_EqualSupNorm(dd,fvn,gvn,tolerance,pr) { return(0); }

    // inlining of macro DD_ONE
    cuddCacheInsert2(dd,(DD_CTFP)Cudd_EqualSupNorm,f,g,((dd).one));

    return(1);

} /* end of Cudd_EqualSupNorm */


/**Function********************************************************************

  Synopsis    [Expands cube to a prime implicant of f.]

  Description [Expands cube to a prime implicant of f. Returns the prime
  if successful; NULL otherwise.  In particular, NULL is returned if cube
  is not a real cube or is not an implicant of f.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
DdNode *
Cudd_bddMakePrime(
  DdManager *dd /* manager */,
  DdNode *cube /* cube to be expanded */,
  DdNode *f /* function of which the cube is to be made a prime */)
{
    var res *DdNode;

    if !Cudd_bddLeq(dd,cube,f) { return(nil); }

    for {
        dd.reordered = 0;
        res = cuddBddMakePrime(dd,cube,f);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddMakePrime */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddMakePrime.]

  Description [Performs the recursive step of Cudd_bddMakePrime.
  Returns the prime if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
DdNode *
cuddBddMakePrime(
  DdManager *dd /* manager */,
  DdNode *cube /* cube to be expanded */,
  DdNode *f /* function of which the cube is to be made a prime */)
{
    var scan *DdNode;
    var t, e *DdNode;
    var res *DdNode  = cube;
    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    var zero *DdNode  = ((DdNode *)((ptrint)(((dd).one)) ^ 01));

    Cudd_Ref(res);
    scan = cube;
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    for !((((DdNode *)((ptruint)(scan) & ~01))).index == CUDD_CONST_INDEX) {
        // inlining of macro Cudd_Regular
        var reg *DdNode  = ((DdNode *)((ptruint)(scan) & ~01));
        var var *DdNode  = dd.vars[reg.index];
        var expanded *DdNode  = Cudd_bddExistAbstract(dd,res,var);
        if expanded == nil {
            return(nil);
        }
        Cudd_Ref(expanded);
        if Cudd_bddLeq(dd,expanded,f) {
            Cudd_RecursiveDeref(dd,res);
            res = expanded;
        } else {
            Cudd_RecursiveDeref(dd,expanded);
        }
        cuddGetBranches(scan,&t,&e);
        if t == zero {
            scan = e;
        } else if e == zero {
            scan = t;
        } else {
            Cudd_RecursiveDeref(dd,res);
            return(nil);       /* cube is not a cube */
        }
    }

    // inlining of macro DD_ONE
    if scan == ((dd).one) {
        Cudd_Deref(res);
        return(res);
    } else {
        Cudd_RecursiveDeref(dd,res);
        return(nil);
    }

} /* end of cuddBddMakePrime */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Frees the entries of the visited symbol table.]

  Description [Frees the entries of the visited symbol table. Returns
  st__CONTINUE.]

  SideEffects [None]

******************************************************************************/
func freePathPair(key string, value string, arg string) st__retval {
    var pair *cuddPathPair;

    pair = (cuddPathPair *) value;
        ABC_FREE(pair);
    return( st__CONTINUE);

} /* end of freePathPair */


/**Function********************************************************************

  Synopsis    [Finds the length of the shortest path(s) in a DD.]

  Description [Finds the length of the shortest path(s) in a DD.
  Uses a local symbol table to store the lengths for each
  node. Only the lengths for the regular nodes are entered in the table,
  because those for the complement nodes are simply obtained by swapping
  the two lenghts.
  Returns a pair of lengths: the length of the shortest path to 1;
  and the length of the shortest path to 0. This is done so as to take
  complement arcs into account.]

  SideEffects [Accumulates the support of the DD in support.]

  SeeAlso     []

******************************************************************************/
func getShortest(root *DdNode, cost *int32, support *int32, visited *st__table) cuddPathPair {
    var my_pair *cuddPathPair;
    var res_pair cuddPathPair;
    var pair_T cuddPathPair;
    var pair_E cuddPathPair;
    var my_root, T, E *DdNode;
    var weight int32;

    // inlining of macro Cudd_Regular
    my_root = ((DdNode *)((ptruint)(root) & ~01));

    if st__lookup(visited, (char *)my_root, (char **)&my_pair) {
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (root) & 01)) != 0 {
            res_pair.pos = my_pair.neg;
            res_pair.neg = my_pair.pos;
        } else {
            res_pair.pos = my_pair.pos;
            res_pair.neg = my_pair.neg;
        }
        return(res_pair);
    }

    /* In the case of a BDD the following test is equivalent to
    ** testing whether the BDD is the constant 1. This formulation,
    ** however, works for ADDs as well, by assuming the usual
    ** dichotomy of 0 and != 0.
    */
    // inlining of macro cuddIsConstant
    if (my_root).index == CUDD_CONST_INDEX {
        if my_root != zero {
            res_pair.pos = 0;
            res_pair.neg = DD_BIGGY;
        } else {
            res_pair.pos = DD_BIGGY;
            res_pair.neg = 0;
        }
    } else {
        // inlining of macro cuddT
        T = ((my_root).type.kids.T);
        // inlining of macro cuddE
        E = ((my_root).type.kids.E);

        pair_T = getShortest(T, cost, support, visited);
        pair_E = getShortest(E, cost, support, visited);
        // inlining of macro WEIGHT
        
        if (cost) == nil {
            weight = 1;
        } else {
            weight = cost[my_root.index];
        }
        // inlining of macro ddMin
        
        if (pair_E.pos) < (pair_T.pos+weight) {
            res_pair.pos = (pair_E.pos);
        } else {
            res_pair.pos = (pair_T.pos+weight);
        }
        // inlining of macro ddMin
        
        if (pair_E.neg) < (pair_T.neg+weight) {
            res_pair.neg = (pair_E.neg);
        } else {
            res_pair.neg = (pair_T.neg+weight);
        }

        /* Update support. */
        if support != nil {
            support[my_root.index] = 1;
        }
    }

    my_pair = new(cuddPathPair);
    if my_pair == nil {
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (root) & 01)) != 0 {
            var tmp int32  = res_pair.pos;
            res_pair.pos = res_pair.neg;
            res_pair.neg = tmp;
        }
        return(res_pair);
    }
    my_pair.pos = res_pair.pos;
    my_pair.neg = res_pair.neg;

    st__insert(visited, (char *)my_root, (char *)my_pair);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (root) & 01)) != 0 {
        res_pair.pos = my_pair.neg;
        res_pair.neg = my_pair.pos;
    } else {
        res_pair.pos = my_pair.pos;
        res_pair.neg = my_pair.neg;
    }
    return(res_pair);

} /* end of getShortest */


/**Function********************************************************************

  Synopsis    [Build a BDD for a shortest path of f.]

  Description [Build a BDD for a shortest path of f.
  Given the minimum length from the root, and the minimum
  lengths for each node (in visited), apply triangulation at each node.
  Of the two children of each node on a shortest path, at least one is
  on a shortest path. In case of ties the procedure chooses the THEN
  children.
  Returns a pointer to the cube BDD representing the path if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func getPath(manager *DdManager, visited *st__table, f *DdNode, weight *int32, cost int32) *DdNode {
    var sol, tmp *DdNode;
    var my_dd, T, E *DdNode;
    var T_pair, E_pair *cuddPathPair;
    var Tcost, Ecost int32;
    var complement int32;

    // inlining of macro Cudd_Regular
    my_dd = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro Cudd_IsComplement
    complement = (int32( ((ptrint) (f) & 01)));

    sol = one;
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(sol).ref);

    // inlining of macro cuddIsConstant
    for !((my_dd).index == CUDD_CONST_INDEX) {
        // inlining of macro WEIGHT
        Tcost = cost - ((weight) == nil ? 1 : weight[my_dd.index]);
        Ecost = cost;

        // inlining of macro cuddT
        T = ((my_dd).type.kids.T);
        // inlining of macro cuddE
        E = ((my_dd).type.kids.E);

        // inlining of macro Cudd_Not
        if complement != 0 {T = ((DdNode *)((ptrint)(T) ^ 01)); E = ((DdNode *)((ptrint)(E) ^ 01));}

        // inlining of macro Cudd_Regular
        st__lookup(visited, (char *)((DdNode *)((ptruint)(T) & ~01)), (char **)&T_pair);
        // inlining of macro Cudd_IsComplement
        if (((int32( ((ptrint) (T) & 01)) != 0)) && T_pair.neg == Tcost) ||
        // inlining of macro Cudd_IsComplement
        (!((int32( ((ptrint) (T) & 01)) != 0)) && T_pair.pos == Tcost) {
            tmp = cuddBddAndRecur(manager,manager.vars[my_dd.index],sol);
            if tmp == nil {
                Cudd_RecursiveDeref(manager,sol);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(tmp).ref);
            Cudd_RecursiveDeref(manager,sol);
            sol = tmp;

            // inlining of macro Cudd_IsComplement
            complement =  (int32( ((ptrint) (T) & 01)));
            // inlining of macro Cudd_Regular
            my_dd = ((DdNode *)((ptruint)(T) & ~01));
            cost = Tcost;
            continue;
        }
        // inlining of macro Cudd_Regular
        st__lookup(visited, (char *)((DdNode *)((ptruint)(E) & ~01)), (char **)&E_pair);
        // inlining of macro Cudd_IsComplement
        if (((int32( ((ptrint) (E) & 01)) != 0)) && E_pair.neg == Ecost) ||
        // inlining of macro Cudd_IsComplement
        (!((int32( ((ptrint) (E) & 01)) != 0)) && E_pair.pos == Ecost) {
            // inlining of macro Cudd_Not
            tmp = cuddBddAndRecur(manager,((DdNode *)((ptrint)(manager.vars[my_dd.index]) ^ 01)),sol);
            if tmp == nil {
                Cudd_RecursiveDeref(manager,sol);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(tmp).ref);
            Cudd_RecursiveDeref(manager,sol);
            sol = tmp;
            // inlining of macro Cudd_IsComplement
            complement = (int32( ((ptrint) (E) & 01)));
            // inlining of macro Cudd_Regular
            my_dd = ((DdNode *)((ptruint)(E) & ~01));
            cost = Ecost;
            continue;
        }
        (void) fmt.Fprintf(manager.err, "We shouldn't be here!!\n");
        manager.errorCode = CUDD_INTERNAL_ERROR;
        return(nil);
    }

    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(sol).ref);
    return(sol);

} /* end of getPath */


/**Function********************************************************************

  Synopsis    [Finds the size of the largest cube(s) in a DD.]

  Description [Finds the size of the largest cube(s) in a DD.
  This problem is translated into finding the shortest paths from a node
  when both THEN and ELSE arcs have unit lengths.
  Uses a local symbol table to store the lengths for each
  node. Only the lengths for the regular nodes are entered in the table,
  because those for the complement nodes are simply obtained by swapping
  the two lenghts.
  Returns a pair of lengths: the length of the shortest path to 1;
  and the length of the shortest path to 0. This is done so as to take
  complement arcs into account.]

  SideEffects [none]

  SeeAlso     []

******************************************************************************/
func getLargest(root *DdNode, visited *st__table) cuddPathPair {
    var my_pair *cuddPathPair;
    var res_pair cuddPathPair;
    var pair_T cuddPathPair;
    var pair_E cuddPathPair;
    var my_root, T, E *DdNode;

    // inlining of macro Cudd_Regular
    my_root = ((DdNode *)((ptruint)(root) & ~01));

    if st__lookup(visited, (char *)my_root, (char **)&my_pair) {
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (root) & 01)) != 0 {
            res_pair.pos = my_pair.neg;
            res_pair.neg = my_pair.pos;
        } else {
            res_pair.pos = my_pair.pos;
            res_pair.neg = my_pair.neg;
        }
        return(res_pair);
    }

    /* In the case of a BDD the following test is equivalent to
    ** testing whether the BDD is the constant 1. This formulation,
    ** however, works for ADDs as well, by assuming the usual
    ** dichotomy of 0 and != 0.
    */
    // inlining of macro cuddIsConstant
    if (my_root).index == CUDD_CONST_INDEX {
        if my_root != zero {
            res_pair.pos = 0;
            res_pair.neg = DD_BIGGY;
        } else {
            res_pair.pos = DD_BIGGY;
            res_pair.neg = 0;
        }
    } else {
        // inlining of macro cuddT
        T = ((my_root).type.kids.T);
        // inlining of macro cuddE
        E = ((my_root).type.kids.E);

        pair_T = getLargest(T, visited);
        pair_E = getLargest(E, visited);
        // inlining of macro ddMin
        res_pair.pos = (((pair_E.pos) < (pair_T.pos)) ? (pair_E.pos) : (pair_T.pos)) + 1;
        // inlining of macro ddMin
        res_pair.neg = (((pair_E.neg) < (pair_T.neg)) ? (pair_E.neg) : (pair_T.neg)) + 1;
    }

    my_pair = new(cuddPathPair);
    if my_pair == nil {      /* simply do not cache this result */
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (root) & 01)) != 0 {
            var tmp int32  = res_pair.pos;
            res_pair.pos = res_pair.neg;
            res_pair.neg = tmp;
        }
        return(res_pair);
    }
    my_pair.pos = res_pair.pos;
    my_pair.neg = res_pair.neg;

    /* Caching may fail without affecting correctness. */
    st__insert(visited, (char *)my_root, (char *)my_pair);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (root) & 01)) != 0 {
        res_pair.pos = my_pair.neg;
        res_pair.neg = my_pair.pos;
    } else {
        res_pair.pos = my_pair.pos;
        res_pair.neg = my_pair.neg;
    }
    return(res_pair);

} /* end of getLargest */


/**Function********************************************************************

  Synopsis    [Build a BDD for a largest cube of f.]

  Description [Build a BDD for a largest cube of f.
  Given the minimum length from the root, and the minimum
  lengths for each node (in visited), apply triangulation at each node.
  Of the two children of each node on a shortest path, at least one is
  on a shortest path. In case of ties the procedure chooses the THEN
  children.
  Returns a pointer to the cube BDD representing the path if
  successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func getCube(manager *DdManager, visited *st__table, f *DdNode, cost int32) *DdNode {
    var sol, tmp *DdNode;
    var my_dd, T, E *DdNode;
    var T_pair, E_pair *cuddPathPair;
    var Tcost, Ecost int32;
    var complement int32;

    // inlining of macro Cudd_Regular
    my_dd = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro Cudd_IsComplement
    complement = (int32( ((ptrint) (f) & 01)));

    sol = one;
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(sol).ref);

    // inlining of macro cuddIsConstant
    for !((my_dd).index == CUDD_CONST_INDEX) {
        Tcost = cost - 1;
        Ecost = cost - 1;

        // inlining of macro cuddT
        T = ((my_dd).type.kids.T);
        // inlining of macro cuddE
        E = ((my_dd).type.kids.E);

        // inlining of macro Cudd_Not
        if complement != 0 {T = ((DdNode *)((ptrint)(T) ^ 01)); E = ((DdNode *)((ptrint)(E) ^ 01));}

        // inlining of macro Cudd_Regular
        if ! st__lookup(visited, (char *)((DdNode *)((ptruint)(T) & ~01)), (char **)&T_pair) { return(nil); }
        // inlining of macro Cudd_IsComplement
        if (((int32( ((ptrint) (T) & 01)) != 0)) && T_pair.neg == Tcost) ||
        // inlining of macro Cudd_IsComplement
        (!((int32( ((ptrint) (T) & 01)) != 0)) && T_pair.pos == Tcost) {
            tmp = cuddBddAndRecur(manager,manager.vars[my_dd.index],sol);
            if tmp == nil {
                Cudd_RecursiveDeref(manager,sol);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(tmp).ref);
            Cudd_RecursiveDeref(manager,sol);
            sol = tmp;

            // inlining of macro Cudd_IsComplement
            complement =  (int32( ((ptrint) (T) & 01)));
            // inlining of macro Cudd_Regular
            my_dd = ((DdNode *)((ptruint)(T) & ~01));
            cost = Tcost;
            continue;
        }
        // inlining of macro Cudd_Regular
        if ! st__lookup(visited, (char *)((DdNode *)((ptruint)(E) & ~01)), (char **)&E_pair) { return(nil); }
        // inlining of macro Cudd_IsComplement
        if (((int32( ((ptrint) (E) & 01)) != 0)) && E_pair.neg == Ecost) ||
        // inlining of macro Cudd_IsComplement
        (!((int32( ((ptrint) (E) & 01)) != 0)) && E_pair.pos == Ecost) {
            // inlining of macro Cudd_Not
            tmp = cuddBddAndRecur(manager,((DdNode *)((ptrint)(manager.vars[my_dd.index]) ^ 01)),sol);
            if tmp == nil {
                Cudd_RecursiveDeref(manager,sol);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(tmp).ref);
            Cudd_RecursiveDeref(manager,sol);
            sol = tmp;
            // inlining of macro Cudd_IsComplement
            complement = (int32( ((ptrint) (E) & 01)));
            // inlining of macro Cudd_Regular
            my_dd = ((DdNode *)((ptruint)(E) & ~01));
            cost = Ecost;
            continue;
        }
        (void) fmt.Fprintf(manager.err, "We shouldn't be here!\n");
        manager.errorCode = CUDD_INTERNAL_ERROR;
        return(nil);
    }

    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(sol).ref);
    return(sol);

} /* end of getCube */


//ABC_NAMESPACE_IMPL_END


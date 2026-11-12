package cudd

/**CFile***********************************************************************

  FileName    [cuddGenCof.c]

  PackageName [cudd]

  Synopsis    [Generalized cofactors for BDDs and ADDs.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_bddConstrain()
                <li> Cudd_bddRestrict()
                <li> Cudd_bddNPAnd()
                <li> Cudd_addConstrain()
                <li> Cudd_bddConstrainDecomp()
                <li> Cudd_addRestrict()
                <li> Cudd_bddCharToVect()
                <li> Cudd_bddLICompaction()
                <li> Cudd_bddSqueeze()
                <li> Cudd_SubsetCompress()
                <li> Cudd_SupersetCompress()
                </ul>
            Internal procedures included in this module:
                <ul>
                <li> cuddBddConstrainRecur()
                <li> cuddBddRestrictRecur()
                <li> cuddBddNPAndRecur()
                <li> cuddAddConstrainRecur()
                <li> cuddAddRestrictRecur()
                <li> cuddBddLICompaction()
                </ul>
            Static procedures included in this module:
                <ul>
                <li> cuddBddConstrainDecomp()
                <li> cuddBddCharToVect()
                <li> cuddBddLICMarkEdges()
                <li> cuddBddLICBuildResult()
                <li> cuddBddSqueeze()
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

/* Codes for edge markings in Cudd_bddLICompaction.  The codes are defined
** so that they can be bitwise ORed to implement the code priority scheme.
*/
const DD_LIC_DC = 0
const DD_LIC_1 =  1
const DD_LIC_0 =  2
const DD_LIC_NL = 3

/*---------------------------------------------------------------------------*/
/* Stucture declarations                                                     */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/

/* Key for the cache used in the edge marking phase. */
typedef struct MarkCacheKey {
    DdNode *f;
    DdNode *c;
}
 MarkCacheKey;

/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddGenCof.c,v 1.38 2005/05/14 17:27:11 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int cuddBddConstrainDecomp (DdManager *dd, DdNode *f, DdNode **decomp);

// static DdNode * cuddBddCharToVect (DdManager *dd, DdNode *f, DdNode *x);

// static int cuddBddLICMarkEdges (DdManager *dd, DdNode *f, DdNode *c, st__table *table, st__table *cache);

// static DdNode * cuddBddLICBuildResult (DdManager *dd, DdNode *f, st__table *cache, st__table *table);

// static int MarkCacheHash (const char *ptr, int modulus);

// static int MarkCacheCompare (const char *ptr1, const char *ptr2);

// static      st__retval MarkCacheCleanUp (char *key, char *value, char *arg);

// static DdNode * cuddBddSqueeze (DdManager *dd, DdNode *l, DdNode *u);


/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Computes f constrain c.]

  Description [Computes f constrain c (f @ c).
  Uses a canonical form: (f' @ c) = ( f @ c)'.  (Note: this is not true
  for c.)  List of special cases:
    <ul>
    <li> f @ 0 = 0
    <li> f @ 1 = f
    <li> 0 @ c = 0
    <li> 1 @ c = 1
    <li> f @ f = 1
    <li> f @ f'= 0
    </ul>
  Returns a pointer to the result if successful; NULL otherwise. Note that if
  F=(f1,...,fn) and reordering takes place while computing F @ c, then the
  image restriction property (Img(F,c) = Img(F @ c)) is lost.]

  SideEffects [None]

  SeeAlso     [Cudd_bddRestrict Cudd_addConstrain]

******************************************************************************/
func Cudd_bddConstrain(dd *DdManager, f *DdNode, c *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddConstrainRecur(dd,f,c);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddConstrain */


/**Function********************************************************************

  Synopsis [BDD restrict according to Coudert and Madre's algorithm
  (ICCAD90).]

  Description [BDD restrict according to Coudert and Madre's algorithm
  (ICCAD90). Returns the restricted BDD if successful; otherwise NULL.
  If application of restrict results in a BDD larger than the input
  BDD, the input BDD is returned.]

  SideEffects [None]

  SeeAlso     [Cudd_bddConstrain Cudd_addRestrict]

******************************************************************************/
func Cudd_bddRestrict(dd *DdManager, f *DdNode, c *DdNode) *DdNode {
    var suppF, suppC, commonSupport *DdNode;
    var cplus, res *DdNode;
    var retval int32;
    var sizeF, sizeRes int32;

    /* Check terminal cases here to avoid computing supports in trivial cases.
    ** This also allows us notto check later for the case c == 0, in which
    ** there is no common support. */
    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    if c == ((DdNode *)((ptrint)(((dd).one)) ^ 01)) { return(((DdNode *)((ptrint)(((dd).one)) ^ 01))); }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX { return(f); }
    // inlining of macro DD_ONE
    if f == c { return(((dd).one)); }
    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    if f == ((DdNode *)((ptrint)(c) ^ 01)) { return(((DdNode *)((ptrint)(((dd).one)) ^ 01))); }

    /* Check if supports intersect. */
    retval = Cudd_ClassifySupport(dd,f,c,&commonSupport,&suppF,&suppC);
    if retval == 0 {
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(commonSupport).ref); cuddSatInc(Cudd_Regular(suppF).ref); cuddSatInc(Cudd_Regular(suppC).ref);
    Cudd_IterDerefBdd(dd,suppF);

    // inlining of macro DD_ONE
    if commonSupport == ((dd).one) {
        Cudd_IterDerefBdd(dd,commonSupport);
        Cudd_IterDerefBdd(dd,suppC);
        return(f);
    }
    Cudd_IterDerefBdd(dd,commonSupport);

    /* Abstract from c the variables that do not appear in f. */
    cplus = Cudd_bddExistAbstract(dd, c, suppC);
    if cplus == nil {
        Cudd_IterDerefBdd(dd,suppC);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(cplus).ref);
    Cudd_IterDerefBdd(dd,suppC);

    for {
        dd.reordered = 0;
        res = cuddBddRestrictRecur(dd, f, cplus);
    
if !(dd.reordered == 1) { break }
} 
    if res == nil {
        Cudd_IterDerefBdd(dd,cplus);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_IterDerefBdd(dd,cplus);
    /* Make restric safe by returning the smaller of the input and the
    ** result. */
    sizeF = Cudd_DagSize(f);
    sizeRes = Cudd_DagSize(res);
    if sizeF <= sizeRes {
        Cudd_IterDerefBdd(dd, res);
        return(f);
    } else {
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref);
        return(res);
    }

} /* end of Cudd_bddRestrict */


/**Function********************************************************************

  Synopsis    [Computes f non-polluting-and g.]

  Description [Computes f non-polluting-and g.  The non-polluting AND
  of f and g is a hybrid of AND and Restrict.  From Restrict, this
  operation takes the idea of existentially quantifying the top
  variable of the second operand if it does not appear in the first.
  Therefore, the variables that appear in the result also appear in f.
  For the rest, the function behaves like AND.  Since the two operands
  play different roles, non-polluting AND is not commutative.

  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddConstrain Cudd_bddRestrict]

******************************************************************************/
func Cudd_bddNPAnd(dd *DdManager, f *DdNode, g *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddNPAndRecur(dd,f,g);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddNPAnd */


/**Function********************************************************************

  Synopsis    [Computes f constrain c for ADDs.]

  Description [Computes f constrain c (f @ c), for f an ADD and c a 0-1
  ADD.  List of special cases:
    <ul>
    <li> F @ 0 = 0
    <li> F @ 1 = F
    <li> 0 @ c = 0
    <li> 1 @ c = 1
    <li> F @ F = 1
    </ul>
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddConstrain]

******************************************************************************/
func Cudd_addConstrain(dd *DdManager, f *DdNode, c *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddAddConstrainRecur(dd,f,c);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_addConstrain */


/**Function********************************************************************

  Synopsis [BDD conjunctive decomposition as in McMillan's CAV96 paper.]

  Description [BDD conjunctive decomposition as in McMillan's CAV96
  paper.  The decomposition is canonical only for a given variable
  order. If canonicity is required, variable ordering must be disabled
  after the decomposition has been computed. Returns an array with one
  entry for each BDD variable in the manager if successful; otherwise
  NULL. The components of the solution have their reference counts
  already incremented (unlike the results of most other functions in
  the package.]

  SideEffects [None]

  SeeAlso     [Cudd_bddConstrain Cudd_bddExistAbstract]

******************************************************************************/
func Cudd_bddConstrainDecomp(dd *DdManager, f *DdNode) **DdNode {
    var decomp **DdNode;
    var res int32;
    var i int32;

    /* Create an initialize decomposition array. */
    decomp = ABC_ALLOC(DdNode *,dd.size);
    if decomp == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    for i = 0; i < dd.size; i++ {
        decomp[i] = nil;
    }
    for {
        dd.reordered = 0;
        /* Clean up the decomposition array in case reordering took place. */
        for i = 0; i < dd.size; i++ {
            if decomp[i] != nil {
                Cudd_IterDerefBdd(dd, decomp[i]);
                decomp[i] = nil;
            }
        }
        res = cuddBddConstrainDecomp(dd,f,decomp);
    
if !(dd.reordered == 1) { break }
} 
    if res == 0 {
        ABC_FREE(decomp);
        return(nil);
    }
    /* Missing components are constant ones. */
    for i = 0; i < dd.size; i++ {
        if decomp[i] == nil {
            // inlining of macro DD_ONE
            decomp[i] = ((dd).one);
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(decomp[i]).ref);
        }
    }
    return(decomp);

} /* end of Cudd_bddConstrainDecomp */


/**Function********************************************************************

  Synopsis [ADD restrict according to Coudert and Madre's algorithm
  (ICCAD90).]

  Description [ADD restrict according to Coudert and Madre's algorithm
  (ICCAD90). Returns the restricted ADD if successful; otherwise NULL.
  If application of restrict results in an ADD larger than the input
  ADD, the input ADD is returned.]

  SideEffects [None]

  SeeAlso     [Cudd_addConstrain Cudd_bddRestrict]

******************************************************************************/
func Cudd_addRestrict(dd *DdManager, f *DdNode, c *DdNode) *DdNode {
    var supp_f, supp_c *DdNode;
    var res, commonSupport *DdNode;
    var intersection int32;
    var sizeF, sizeRes int32;

    /* Check if supports intersect. */
    supp_f = Cudd_Support(dd, f);
    if supp_f == nil {
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(supp_f).ref);
    supp_c = Cudd_Support(dd, c);
    if supp_c == nil {
        Cudd_RecursiveDeref(dd,supp_f);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(supp_c).ref);
    commonSupport = Cudd_bddLiteralSetIntersection(dd, supp_f, supp_c);
    if commonSupport == nil {
        Cudd_RecursiveDeref(dd,supp_f);
        Cudd_RecursiveDeref(dd,supp_c);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(commonSupport).ref);
    Cudd_RecursiveDeref(dd,supp_f);
    Cudd_RecursiveDeref(dd,supp_c);
    // inlining of macro DD_ONE
    intersection = commonSupport != ((dd).one);
    Cudd_RecursiveDeref(dd,commonSupport);

    if intersection != 0 {
        for {
            dd.reordered = 0;
            res = cuddAddRestrictRecur(dd, f, c);
        
if !(dd.reordered == 1) { break }
} 
        sizeF = Cudd_DagSize(f);
        sizeRes = Cudd_DagSize(res);
        if sizeF <= sizeRes {
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res).ref);
            Cudd_RecursiveDeref(dd, res);
            return(f);
        } else {
            return(res);
        }
    } else {
        return(f);
    }

} /* end of Cudd_addRestrict */


/**Function********************************************************************

  Synopsis    [Computes a vector whose image equals a non-zero function.]

  Description [Computes a vector of BDDs whose image equals a non-zero
  function.
  The result depends on the variable order. The i-th component of the vector
  depends only on the first i variables in the order.  Each BDD in the vector
  is not larger than the BDD of the given characteristic function.  This
  function is based on the description of char-to-vect in "Verification of
  Sequential Machines Using Boolean Functional Vectors" by O. Coudert, C.
  Berthet and J. C. Madre.
  Returns a pointer to an array containing the result if successful; NULL
  otherwise. The size of the array equals the number of variables in the
  manager. The components of the solution have their reference counts 
  already incremented (unlike the results of most other functions in 
  the package).]

  SideEffects [None]

  SeeAlso     [Cudd_bddConstrain]

******************************************************************************/
func Cudd_bddCharToVect(dd *DdManager, f *DdNode) **DdNode {
    var i, j int32;
    var vect **DdNode;
    var res *DdNode  = nil;

    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    if f == ((DdNode *)((ptrint)(((dd).one)) ^ 01)) { return(nil); }

    vect = ABC_ALLOC(DdNode *, dd.size);
    if vect == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }

    for {
        dd.reordered = 0;
        for i = 0; i < dd.size; i++ {
            res = cuddBddCharToVect(dd,f,dd.vars[dd.invperm[i]]);
            if res == nil {
                /* Clean up the vector array in case reordering took place. */
                for j = 0; j < i; j++ {
                    Cudd_IterDerefBdd(dd, vect[dd.invperm[j]]);
                }
                break;
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res).ref);
            vect[dd.invperm[i]] = res;
        }
    
if !(dd.reordered == 1) { break }
} 
    if res == nil {
        ABC_FREE(vect);
        return(nil);
    }
    return(vect);

} /* end of Cudd_bddCharToVect */


/**Function********************************************************************

  Synopsis    [Performs safe minimization of a BDD.]

  Description [Performs safe minimization of a BDD. Given the BDD
  <code>f</code> of a function to be minimized and a BDD
  <code>c</code> representing the care set, Cudd_bddLICompaction
  produces the BDD of a function that agrees with <code>f</code>
  wherever <code>c</code> is 1.  Safe minimization means that the size
  of the result is guaranteed not to exceed the size of
  <code>f</code>. This function is based on the DAC97 paper by Hong et
  al..  Returns a pointer to the result if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddRestrict]

******************************************************************************/
DdNode *
Cudd_bddLICompaction(
  DdManager * dd /* manager */,
  DdNode * f /* function to be minimized */,
  DdNode * c /* constraint (care set) */)
{
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddLICompaction(dd,f,c);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddLICompaction */


/**Function********************************************************************

  Synopsis    [Finds a small BDD in a function interval.]

  Description [Finds a small BDD in a function interval. Given BDDs
  <code>l</code> and <code>u</code>, representing the lower bound and
  upper bound of a function interval, Cudd_bddSqueeze produces the BDD
  of a function within the interval with a small BDD.  Returns a
  pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddRestrict Cudd_bddLICompaction]

******************************************************************************/
DdNode *
Cudd_bddSqueeze(
  DdManager * dd /* manager */,
  DdNode * l /* lower bound */,
  DdNode * u /* upper bound */)
{
    var res *DdNode;
    var sizeRes, sizeL, sizeU int32;

    for {
        dd.reordered = 0;
        res = cuddBddSqueeze(dd,l,u);
    
if !(dd.reordered == 1) { break }
} 
    if res == nil { return(nil); }
    /* We now compare the result with the bounds and return the smallest.
    ** We first compare to u, so that in case l == 0 and u == 1, we return
    ** 0 as in other minimization algorithms. */
    sizeRes = Cudd_DagSize(res);
    sizeU = Cudd_DagSize(u);
    if sizeU <= sizeRes {
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_IterDerefBdd(dd,res);
        res = u;
        sizeRes = sizeU;
    }
    sizeL = Cudd_DagSize(l);
    if sizeL <= sizeRes {
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_IterDerefBdd(dd,res);
        res = l;
        sizeRes = sizeL;
    }
    return(res);

} /* end of Cudd_bddSqueeze */


/**Function********************************************************************

  Synopsis    [Finds a small BDD that agrees with <code>f</code> over
  <code>c</code>.]

  Description [Finds a small BDD that agrees with <code>f</code> over
  <code>c</code>.  Returns a pointer to the result if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddRestrict Cudd_bddLICompaction Cudd_bddSqueeze]

******************************************************************************/
func Cudd_bddMinimize(dd *DdManager, f *DdNode, c *DdNode) *DdNode {
    var cplus, res *DdNode;

    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    if c == ((DdNode *)((ptrint)(((dd).one)) ^ 01)) { return(c); }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX { return(f); }
    // inlining of macro DD_ONE
    if f == c { return(((dd).one)); }
    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    if f == ((DdNode *)((ptrint)(c) ^ 01)) { return(((DdNode *)((ptrint)(((dd).one)) ^ 01))); }

    cplus = Cudd_RemapOverApprox(dd,c,0,0,1.0);
    if cplus == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(cplus).ref);
    res = Cudd_bddLICompaction(dd,f,cplus);
    if res == nil {
        Cudd_IterDerefBdd(dd,cplus);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_IterDerefBdd(dd,cplus);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of Cudd_bddMinimize */


/**Function********************************************************************

  Synopsis    [Find a dense subset of BDD <code>f</code>.]

  Description [Finds a dense subset of BDD <code>f</code>. Density is
  the ratio of number of minterms to number of nodes.  Uses several
  techniques in series. It is more expensive than other subsetting
  procedures, but often produces better results. See
  Cudd_SubsetShortPaths for a description of the threshold and nvars
  parameters.  Returns a pointer to the result if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_SubsetRemap Cudd_SubsetShortPaths Cudd_SubsetHeavyBranch
  Cudd_bddSqueeze]

******************************************************************************/
DdNode *
Cudd_SubsetCompress(
  DdManager * dd /* manager */,
  DdNode * f /* BDD whose subset is sought */,
  int  nvars /* number of variables in the support of f */,
  int  threshold /* maximum number of nodes in the subset */)
{
    var res, tmp1, tmp2 *DdNode;

    tmp1 = Cudd_SubsetShortPaths(dd, f, nvars, threshold, 0);
    if tmp1 == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(tmp1).ref);
    tmp2 = Cudd_RemapUnderApprox(dd,tmp1,nvars,0,1.0);
    if tmp2 == nil {
        Cudd_IterDerefBdd(dd,tmp1);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(tmp2).ref);
    Cudd_IterDerefBdd(dd,tmp1);
    res = Cudd_bddSqueeze(dd,tmp2,f);
    if res == nil {
        Cudd_IterDerefBdd(dd,tmp2);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_IterDerefBdd(dd,tmp2);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of Cudd_SubsetCompress */


/**Function********************************************************************

  Synopsis    [Find a dense superset of BDD <code>f</code>.]

  Description [Finds a dense superset of BDD <code>f</code>. Density is
  the ratio of number of minterms to number of nodes.  Uses several
  techniques in series. It is more expensive than other supersetting
  procedures, but often produces better results. See
  Cudd_SupersetShortPaths for a description of the threshold and nvars
  parameters.  Returns a pointer to the result if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_SubsetCompress Cudd_SupersetRemap Cudd_SupersetShortPaths
  Cudd_SupersetHeavyBranch Cudd_bddSqueeze]

******************************************************************************/
DdNode *
Cudd_SupersetCompress(
  DdManager * dd /* manager */,
  DdNode * f /* BDD whose superset is sought */,
  int  nvars /* number of variables in the support of f */,
  int  threshold /* maximum number of nodes in the superset */)
{
    var subset *DdNode;

    // inlining of macro Cudd_Not
    subset = Cudd_SubsetCompress(dd, ((DdNode *)((ptrint)(f) ^ 01)),nvars,threshold);

    // inlining of macro Cudd_NotCond
    return(((DdNode *)((ptrint)(subset) ^ ((subset != nil)))));

} /* end of Cudd_SupersetCompress */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddConstrain.]

  Description [Performs the recursive step of Cudd_bddConstrain.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddConstrain]

******************************************************************************/
func cuddBddConstrainRecur(dd *DdManager, f *DdNode, c *DdNode) *DdNode {
    var Fv, Fnv, Cv, Cnv, t, e, r *DdNode;
    var one, zero *DdNode;
    unsigned int topf, topc;
    var index int32;
    var comple int32;

    statLine(dd);
    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Trivial cases. */
    if c == one {               return(f); }
    if c == zero {              return(zero); }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {     return(f); }
    if f == c {                 return(one); }
    // inlining of macro Cudd_Not
    if f == ((DdNode *)((ptrint)(c) ^ 01)) {       return(zero); }

    /* Make canonical to increase the utilization of the cache. */
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (f) & 01)) != 0 {
        // inlining of macro Cudd_Not
        f = ((DdNode *)((ptrint)(f) ^ 01));
        comple = 1;
    }
    /* Now f is a regular pointer to a non-constant node; c is also
    ** non-constant, but may be complemented.
    */

    /* Check the cache. */
    r = cuddCacheLookup2(dd, Cudd_bddConstrain, f, c);
    if r != nil {
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    }
    
    /* Recursive step. */
    topf = dd.perm[f.index];
    // inlining of macro Cudd_Regular
    topc = dd.perm[((DdNode *)((ptruint)(c) & ~01)).index];
    if topf <= topc {
        index = f.index;
        // inlining of macro cuddT
        // inlining of macro cuddE
        Fv = ((f).type.kids.T); Fnv = ((f).type.kids.E);
    } else {
        // inlining of macro Cudd_Regular
        index = ((DdNode *)((ptruint)(c) & ~01)).index;
        Fv = Fnv = f;
    }
    if topc <= topf {
        // inlining of macro cuddT
        // inlining of macro Cudd_Regular
        // inlining of macro cuddE
        Cv = ((((DdNode *)((ptruint)(c) & ~01))).type.kids.T); Cnv = ((((DdNode *)((ptruint)(c) & ~01))).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (c) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Cv = ((DdNode *)((ptrint)(Cv) ^ 01));
            // inlining of macro Cudd_Not
            Cnv = ((DdNode *)((ptrint)(Cnv) ^ 01));
        }
    } else {
        Cv = Cnv = c;
    }

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Cv) & ~01))).index == CUDD_CONST_INDEX) {
        t = cuddBddConstrainRecur(dd, Fv, Cv);
        if t == nil {
            return(nil); }
    } else if Cv == one {
        t = Fv;
    } else {            /* Cv == zero: return Fnv @ Cnv */
        if Cnv == one {
            r = Fnv;
        } else {
            r = cuddBddConstrainRecur(dd, Fnv, Cnv);
            if r == nil {
                return(nil); }
        }
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Cnv) & ~01))).index == CUDD_CONST_INDEX) {
        e = cuddBddConstrainRecur(dd, Fnv, Cnv);
        if e == nil {
            Cudd_IterDerefBdd(dd, t);
            return(nil);
        }
    } else if Cnv == one {
        e = Fnv;
    } else {            /* Cnv == zero: return Fv @ Cv previously computed */
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(t) ^ (comple))));
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (t) & 01)) != 0 {
        // inlining of macro Cudd_Not
        t = ((DdNode *)((ptrint)(t) ^ 01));
        // inlining of macro Cudd_Not
        e = ((DdNode *)((ptrint)(e) ^ 01));
        if t == e {
            r = t;
        } else {
            r = cuddUniqueInter(dd, index, t, e);
        }
        if r == nil {
            Cudd_IterDerefBdd(dd, e);
            Cudd_IterDerefBdd(dd, t);
            return(nil);
        }
        // inlining of macro Cudd_Not
        r = ((DdNode *)((ptrint)(r) ^ 01));
    } else {
        if t == e {
            r = t;
        } else {
            r = cuddUniqueInter(dd, index, t, e);
        }
        if r == nil {
            Cudd_IterDerefBdd(dd, e);
            Cudd_IterDerefBdd(dd, t);
            return(nil);
        }
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);

    cuddCacheInsert2(dd, Cudd_bddConstrain, f, c, r);
    // inlining of macro Cudd_NotCond
    return(((DdNode *)((ptrint)(r) ^ (comple))));

} /* end of cuddBddConstrainRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddRestrict.]

  Description [Performs the recursive step of Cudd_bddRestrict.
  Returns the restricted BDD if successful; otherwise NULL.]

  SideEffects [None]

  SeeAlso     [Cudd_bddRestrict]

******************************************************************************/
func cuddBddRestrictRecur(dd *DdManager, f *DdNode, c *DdNode) *DdNode {
    var Fv, Fnv, Cv, Cnv, t, e, r, one, zero *DdNode;
    unsigned int topf, topc;
    var index int32;
    var comple int32;

    statLine(dd);
    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Trivial cases */
    if c == one {               return(f); }
    if c == zero {              return(zero); }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {     return(f); }
    if f == c {                 return(one); }
    // inlining of macro Cudd_Not
    if f == ((DdNode *)((ptrint)(c) ^ 01)) {       return(zero); }

    /* Make canonical to increase the utilization of the cache. */
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (f) & 01)) != 0 {
        // inlining of macro Cudd_Not
        f = ((DdNode *)((ptrint)(f) ^ 01));
        comple = 1;
    }
    /* Now f is a regular pointer to a non-constant node; c is also
    ** non-constant, but may be complemented.
    */

    /* Check the cache. */
    r = cuddCacheLookup2(dd, Cudd_bddRestrict, f, c);
    if r != nil {
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    }

    topf = dd.perm[f.index];
    // inlining of macro Cudd_Regular
    topc = dd.perm[((DdNode *)((ptruint)(c) & ~01)).index];

    if topc < topf {  /* abstract top variable from c */
        var d, s1, s2 *DdNode;

        /* Find complements of cofactors of c. */
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (c) & 01)) != 0 {
            // inlining of macro cuddT
            // inlining of macro Cudd_Regular
            s1 = ((((DdNode *)((ptruint)(c) & ~01))).type.kids.T);
            // inlining of macro cuddE
            // inlining of macro Cudd_Regular
            s2 = ((((DdNode *)((ptruint)(c) & ~01))).type.kids.E);
        } else {
            // inlining of macro Cudd_Not
            // inlining of macro cuddT
            s1 = ((DdNode *)((ptrint)(((c).type.kids.T)) ^ 01));
            // inlining of macro Cudd_Not
            // inlining of macro cuddE
            s2 = ((DdNode *)((ptrint)(((c).type.kids.E)) ^ 01));
        }
        /* Take the OR by applying DeMorgan. */
        d = cuddBddAndRecur(dd, s1, s2);
        if d == nil { return(nil); }
        // inlining of macro Cudd_Not
        d = ((DdNode *)((ptrint)(d) ^ 01));
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(d).ref);
        r = cuddBddRestrictRecur(dd, f, d);
        if r == nil {
            Cudd_IterDerefBdd(dd, d);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(r).ref);
        Cudd_IterDerefBdd(dd, d);
        cuddCacheInsert2(dd, Cudd_bddRestrict, f, c, r);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(r).ref);
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    }

    /* Recursive step. Here topf <= topc. */
    index = f.index;
    // inlining of macro cuddT
    // inlining of macro cuddE
    Fv = ((f).type.kids.T); Fnv = ((f).type.kids.E);
    if topc == topf {
        // inlining of macro cuddT
        // inlining of macro Cudd_Regular
        // inlining of macro cuddE
        Cv = ((((DdNode *)((ptruint)(c) & ~01))).type.kids.T); Cnv = ((((DdNode *)((ptruint)(c) & ~01))).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (c) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Cv = ((DdNode *)((ptrint)(Cv) ^ 01));
            // inlining of macro Cudd_Not
            Cnv = ((DdNode *)((ptrint)(Cnv) ^ 01));
        }
    } else {
        Cv = Cnv = c;
    }

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Cv) & ~01))).index == CUDD_CONST_INDEX) {
        t = cuddBddRestrictRecur(dd, Fv, Cv);
        if t == nil { return(nil); }
    } else if Cv == one {
        t = Fv;
    } else {            /* Cv == zero: return(Fnv @ Cnv) */
        if Cnv == one {
            r = Fnv;
        } else {
            r = cuddBddRestrictRecur(dd, Fnv, Cnv);
            if r == nil { return(nil); }
        }
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Cnv) & ~01))).index == CUDD_CONST_INDEX) {
        e = cuddBddRestrictRecur(dd, Fnv, Cnv);
        if e == nil {
            Cudd_IterDerefBdd(dd, t);
            return(nil);
        }
    } else if Cnv == one {
        e = Fnv;
    } else {            /* Cnv == zero: return (Fv @ Cv) previously computed */
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(t) ^ (comple))));
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (t) & 01)) != 0 {
        // inlining of macro Cudd_Not
        t = ((DdNode *)((ptrint)(t) ^ 01));
        // inlining of macro Cudd_Not
        e = ((DdNode *)((ptrint)(e) ^ 01));
        if t == e {
            r = t;
        } else {
            r = cuddUniqueInter(dd, index, t, e);
        }
        if r == nil {
            Cudd_IterDerefBdd(dd, e);
            Cudd_IterDerefBdd(dd, t);
            return(nil);
        }
        // inlining of macro Cudd_Not
        r = ((DdNode *)((ptrint)(r) ^ 01));
    } else {
        if t == e {
            r = t;
        } else {
            r = cuddUniqueInter(dd, index, t, e);
        }
        if r == nil {
            Cudd_IterDerefBdd(dd, e);
            Cudd_IterDerefBdd(dd, t);
            return(nil);
        }
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);

    cuddCacheInsert2(dd, Cudd_bddRestrict, f, c, r);
    // inlining of macro Cudd_NotCond
    return(((DdNode *)((ptrint)(r) ^ (comple))));

} /* end of cuddBddRestrictRecur */


/**Function********************************************************************

  Synopsis [Implements the recursive step of Cudd_bddAnd.]

  Description [Implements the recursive step of Cudd_bddNPAnd.
  Returns a pointer to the result is successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddNPAnd]

******************************************************************************/
func cuddBddNPAndRecur(manager *DdManager, f *DdNode, g *DdNode) *DdNode {
    var F, ft, fe, G, gt, ge *DdNode;
    var one, r, t, e *DdNode;
    unsigned int topf, topg, index;

    statLine(manager);
    // inlining of macro DD_ONE
    one = ((manager).one);

    /* Terminal cases. */
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    if F == G {
        if f == g { return(one);
        // inlining of macro Cudd_Not
        } else { return(((DdNode *)((ptrint)(one) ^ 01))); }
    }
    if G == one {
        if g == one { return(f);
        } else { return(g); }
    }
    if F == one {
        return(f);
    }

    /* At this point f and g are not constant. */
    /* Check cache. */
    if F.ref != 1 || G.ref != 1 {
        r = cuddCacheLookup2(manager, Cudd_bddNPAnd, f, g);
        if r != nil { return(r); }
    }

    /* Here we can skip the use of cuddI, because the operands are known
    ** to be non-constant.
    */
    topf = manager.perm[F.index];
    topg = manager.perm[G.index];

    if topg < topf {  /* abstract top variable from g */
        var d *DdNode;

        /* Find complements of cofactors of g. */
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (g) & 01)) != 0 {
            // inlining of macro cuddT
            gt = ((G).type.kids.T);
            // inlining of macro cuddE
            ge = ((G).type.kids.E);
        } else {
            // inlining of macro Cudd_Not
            // inlining of macro cuddT
            gt = ((DdNode *)((ptrint)(((g).type.kids.T)) ^ 01));
            // inlining of macro Cudd_Not
            // inlining of macro cuddE
            ge = ((DdNode *)((ptrint)(((g).type.kids.E)) ^ 01));
        }
        /* Take the OR by applying DeMorgan. */
        d = cuddBddAndRecur(manager, gt, ge);
        if d == nil { return(nil); }
        // inlining of macro Cudd_Not
        d = ((DdNode *)((ptrint)(d) ^ 01));
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(d).ref);
        r = cuddBddNPAndRecur(manager, f, d);
        if r == nil {
            Cudd_IterDerefBdd(manager, d);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(r).ref);
        Cudd_IterDerefBdd(manager, d);
        cuddCacheInsert2(manager, Cudd_bddNPAnd, f, g, r);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(r).ref);
        return(r);
    }

    /* Compute cofactors. */
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

    if topg == topf {
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

    t = cuddBddAndRecur(manager, ft, gt);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    e = cuddBddAndRecur(manager, fe, ge);
    if e == nil {
        Cudd_IterDerefBdd(manager, t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    if t == e {
        r = t;
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
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    if F.ref != 1 || G.ref != 1 {
        cuddCacheInsert2(manager, Cudd_bddNPAnd, f, g, r); }
    return(r);

} /* end of cuddBddNPAndRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_addConstrain.]

  Description [Performs the recursive step of Cudd_addConstrain.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addConstrain]

******************************************************************************/
func cuddAddConstrainRecur(dd *DdManager, f *DdNode, c *DdNode) *DdNode {
    var Fv, Fnv, Cv, Cnv, t, e, r *DdNode;
    var one, zero *DdNode;
    unsigned int topf, topc;
    var index int32;

    statLine(dd);
    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro DD_ZERO
    zero = ((dd).zero);

    /* Trivial cases. */
    if c == one {               return(f); }
    if c == zero {              return(zero); }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {     return(f); }
    if f == c {                 return(one); }

    /* Now f and c are non-constant. */

    /* Check the cache. */
    r = cuddCacheLookup2(dd, Cudd_addConstrain, f, c);
    if r != nil {
        return(r);
    }
    
    /* Recursive step. */
    topf = dd.perm[f.index];
    topc = dd.perm[c.index];
    if topf <= topc {
        index = f.index;
        // inlining of macro cuddT
        // inlining of macro cuddE
        Fv = ((f).type.kids.T); Fnv = ((f).type.kids.E);
    } else {
        index = c.index;
        Fv = Fnv = f;
    }
    if topc <= topf {
        // inlining of macro cuddT
        // inlining of macro cuddE
        Cv = ((c).type.kids.T); Cnv = ((c).type.kids.E);
    } else {
        Cv = Cnv = c;
    }

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Cv) & ~01))).index == CUDD_CONST_INDEX) {
        t = cuddAddConstrainRecur(dd, Fv, Cv);
        if t == nil {
            return(nil); }
    } else if Cv == one {
        t = Fv;
    } else {            /* Cv == zero: return Fnv @ Cnv */
        if Cnv == one {
            r = Fnv;
        } else {
            r = cuddAddConstrainRecur(dd, Fnv, Cnv);
            if r == nil {
                return(nil); }
        }
        return(r);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Cnv) & ~01))).index == CUDD_CONST_INDEX) {
        e = cuddAddConstrainRecur(dd, Fnv, Cnv);
        if e == nil {
            Cudd_RecursiveDeref(dd, t);
            return(nil);
        }
    } else if Cnv == one {
        e = Fnv;
    } else {            /* Cnv == zero: return Fv @ Cv previously computed */
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        return(t);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    if t == e {
        r = t;
    } else {
        r = cuddUniqueInter(dd, index, t, e);
    }
    if r == nil {
        Cudd_RecursiveDeref(dd, e);
        Cudd_RecursiveDeref(dd, t);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);

    cuddCacheInsert2(dd, Cudd_addConstrain, f, c, r);
    return(r);

} /* end of cuddAddConstrainRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_addRestrict.]

  Description [Performs the recursive step of Cudd_addRestrict.
  Returns the restricted ADD if successful; otherwise NULL.]

  SideEffects [None]

  SeeAlso     [Cudd_addRestrict]

******************************************************************************/
func cuddAddRestrictRecur(dd *DdManager, f *DdNode, c *DdNode) *DdNode {
    var Fv, Fnv, Cv, Cnv, t, e, r, one, zero *DdNode;
    unsigned int topf, topc;
    var index int32;

    statLine(dd);
    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro DD_ZERO
    zero = ((dd).zero);

    /* Trivial cases */
    if c == one {               return(f); }
    if c == zero {              return(zero); }
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX {     return(f); }
    if f == c {                 return(one); }

    /* Now f and c are non-constant. */

    /* Check the cache. */
    r = cuddCacheLookup2(dd, Cudd_addRestrict, f, c);
    if r != nil {
        return(r);
    }

    topf = dd.perm[f.index];
    topc = dd.perm[c.index];

    if topc < topf {  /* abstract top variable from c */
        var d, s1, s2 *DdNode;

        /* Find cofactors of c. */
        // inlining of macro cuddT
        s1 = ((c).type.kids.T);
        // inlining of macro cuddE
        s2 = ((c).type.kids.E);
        /* Take the OR by applying DeMorgan. */
        d = cuddAddApplyRecur(dd, Cudd_addOr, s1, s2);
        if d == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(d).ref);
        r = cuddAddRestrictRecur(dd, f, d);
        if r == nil {
            Cudd_RecursiveDeref(dd, d);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(r).ref);
        Cudd_RecursiveDeref(dd, d);
        cuddCacheInsert2(dd, Cudd_addRestrict, f, c, r);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(r).ref);
        return(r);
    }

    /* Recursive step. Here topf <= topc. */
    index = f.index;
    // inlining of macro cuddT
    // inlining of macro cuddE
    Fv = ((f).type.kids.T); Fnv = ((f).type.kids.E);
    if topc == topf {
        // inlining of macro cuddT
        // inlining of macro cuddE
        Cv = ((c).type.kids.T); Cnv = ((c).type.kids.E);
    } else {
        Cv = Cnv = c;
    }

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Cv) & ~01))).index == CUDD_CONST_INDEX) {
        t = cuddAddRestrictRecur(dd, Fv, Cv);
        if t == nil { return(nil); }
    } else if Cv == one {
        t = Fv;
    } else {            /* Cv == zero: return(Fnv @ Cnv) */
        if Cnv == one {
            r = Fnv;
        } else {
            r = cuddAddRestrictRecur(dd, Fnv, Cnv);
            if r == nil { return(nil); }
        }
        return(r);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Cnv) & ~01))).index == CUDD_CONST_INDEX) {
        e = cuddAddRestrictRecur(dd, Fnv, Cnv);
        if e == nil {
            Cudd_RecursiveDeref(dd, t);
            return(nil);
        }
    } else if Cnv == one {
        e = Fnv;
    } else {            /* Cnv == zero: return (Fv @ Cv) previously computed */
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        return(t);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    if t == e {
        r = t;
    } else {
        r = cuddUniqueInter(dd, index, t, e);
    }
    if r == nil {
        Cudd_RecursiveDeref(dd, e);
        Cudd_RecursiveDeref(dd, t);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);

    cuddCacheInsert2(dd, Cudd_addRestrict, f, c, r);
    return(r);

} /* end of cuddAddRestrictRecur */



/**Function********************************************************************

  Synopsis    [Performs safe minimization of a BDD.]

  Description [Performs safe minimization of a BDD. Given the BDD
  <code>f</code> of a function to be minimized and a BDD
  <code>c</code> representing the care set, Cudd_bddLICompaction
  produces the BDD of a function that agrees with <code>f</code>
  wherever <code>c</code> is 1.  Safe minimization means that the size
  of the result is guaranteed not to exceed the size of
  <code>f</code>. This function is based on the DAC97 paper by Hong et
  al..  Returns a pointer to the result if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddLICompaction]

******************************************************************************/
DdNode *
cuddBddLICompaction(
  DdManager * dd /* manager */,
  DdNode * f /* function to be minimized */,
  DdNode * c /* constraint (care set) */)
{
    var marktable, markcache, buildcache *st__table;
    var res, zero *DdNode;

    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    zero = ((DdNode *)((ptrint)(((dd).one)) ^ 01));
    if c == zero { return(zero); }

    /* We need to use local caches for both steps of this operation.
    ** The results of the edge marking step are only valid as long as the
    ** edge markings themselves are available. However, the edge markings
    ** are lost at the end of one invocation of Cudd_bddLICompaction.
    ** Hence, the cache entries for the edge marking step must be
    ** invalidated at the end of this function.
    ** For the result of the building step we argue as follows. The result
    ** for a node and a given constrain depends on the BDD in which the node
    ** appears. Hence, the same node and constrain may give different results
    ** in successive invocations.
    */
    marktable = st__init_table( st__ptrcmp, st__ptrhash);
    if marktable == nil {
        return(nil);
    }
    markcache = st__init_table(MarkCacheCompare,MarkCacheHash);
    if markcache == nil {
        st__free_table(marktable);
        return(nil);
    }
    if cuddBddLICMarkEdges(dd,f,c,marktable,markcache) == CUDD_OUT_OF_MEM {
        st__foreach(markcache, MarkCacheCleanUp, nil);
        st__free_table(marktable);
        st__free_table(markcache);
        return(nil);
    }
    st__foreach(markcache, MarkCacheCleanUp, nil);
    st__free_table(markcache);
    buildcache = st__init_table( st__ptrcmp, st__ptrhash);
    if buildcache == nil {
        st__free_table(marktable);
        return(nil);
    }
    res = cuddBddLICBuildResult(dd,f,buildcache,marktable);
    st__free_table(buildcache);
    st__free_table(marktable);
    return(res);

} /* end of cuddBddLICompaction */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddConstrainDecomp.]

  Description [Performs the recursive step of Cudd_bddConstrainDecomp.
  Returns f super (i) if successful; otherwise NULL.]

  SideEffects [None]

  SeeAlso     [Cudd_bddConstrainDecomp]

******************************************************************************/
func cuddBddConstrainDecomp(dd *DdManager, f *DdNode, decomp **DdNode) int32 {
    var F, fv, fvn *DdNode;
    var fAbs *DdNode;
    var result *DdNode;
    var ok int32;

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX { return(1); }
    /* Compute complements of cofactors. */
    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));
    // inlining of macro cuddT
    fv = ((F).type.kids.T);
    // inlining of macro cuddE
    fvn = ((F).type.kids.E);
    if F == f {
        // inlining of macro Cudd_Not
        fv = ((DdNode *)((ptrint)(fv) ^ 01));
        // inlining of macro Cudd_Not
        fvn = ((DdNode *)((ptrint)(fvn) ^ 01));
    }
    /* Compute abstraction of top variable. */
    fAbs = cuddBddAndRecur(dd, fv, fvn);
    if fAbs == nil {
        return(0);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(fAbs).ref);
    // inlining of macro Cudd_Not
    fAbs = ((DdNode *)((ptrint)(fAbs) ^ 01));
    /* Recursively find the next abstraction and the components of the
    ** decomposition. */
    ok = cuddBddConstrainDecomp(dd, fAbs, decomp);
    if ok == 0 {
        Cudd_IterDerefBdd(dd,fAbs);
        return(0);
    }
    /* Compute the component of the decomposition corresponding to the
    ** top variable and store it in the decomposition array. */
    result = cuddBddConstrainRecur(dd, f, fAbs);
    if result == nil {
        Cudd_IterDerefBdd(dd,fAbs);
        return(0);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(result).ref);
    decomp[F.index] = result;
    Cudd_IterDerefBdd(dd, fAbs);
    return(1);

} /* end of cuddBddConstrainDecomp */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddCharToVect.]

  Description [Performs the recursive step of Cudd_bddCharToVect.
  This function maintains the invariant that f is non-zero.
  Returns the i-th component of the vector if successful; otherwise NULL.]

  SideEffects [None]

  SeeAlso     [Cudd_bddCharToVect]

******************************************************************************/
func cuddBddCharToVect(dd *DdManager, f *DdNode, x *DdNode) *DdNode {
    unsigned int topf;
    unsigned int level;
    var comple int32;

    var one, zero, res, F, fT, fE, T, E *DdNode;

    statLine(dd);
    /* Check the cache. */
    res = cuddCacheLookup2(dd, cuddBddCharToVect, f, x);
    if res != nil {
        return(res);
    }

    // inlining of macro Cudd_Regular
    F = ((DdNode *)((ptruint)(f) & ~01));

    // inlining of macro cuddI
    
    if (F.index)==CUDD_CONST_INDEX {
        topf = int32((F.index));
    } else {
        topf = (dd).perm[(F.index)];
    }
    level = dd.perm[x.index];

    if topf > level { return(x); }

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    comple = F != f;
    // inlining of macro Cudd_NotCond
    // inlining of macro cuddT
    fT = ((DdNode *)((ptrint)(((F).type.kids.T)) ^ (comple)));
    // inlining of macro Cudd_NotCond
    // inlining of macro cuddE
    fE = ((DdNode *)((ptrint)(((F).type.kids.E)) ^ (comple)));

    if topf == level {
        if fT == zero { return(zero); }
        if fE == zero { return(one); }
        return(x);
    }

    /* Here topf < level. */
    if fT == zero { return(cuddBddCharToVect(dd, fE, x)); }
    if fE == zero { return(cuddBddCharToVect(dd, fT, x)); }

    T = cuddBddCharToVect(dd, fT, x);
    if T == nil {
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);
    E = cuddBddCharToVect(dd, fE, x);
    if E == nil {
        Cudd_IterDerefBdd(dd,T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);
    res = cuddBddIteRecur(dd, dd.vars[F.index], T, E);
    if res == nil {
        Cudd_IterDerefBdd(dd,T);
        Cudd_IterDerefBdd(dd,E);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(T).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(E).ref);
    cuddCacheInsert2(dd, cuddBddCharToVect, f, x, res);
    return(res);

} /* end of cuddBddCharToVect */


/**Function********************************************************************

  Synopsis    [Performs the edge marking step of Cudd_bddLICompaction.]

  Description [Performs the edge marking step of Cudd_bddLICompaction.
  Returns the LUB of the markings of the two outgoing edges of <code>f</code>
  if successful; otherwise CUDD_OUT_OF_MEM.]

  SideEffects [None]

  SeeAlso     [Cudd_bddLICompaction cuddBddLICBuildResult]

******************************************************************************/
func cuddBddLICMarkEdges(dd *DdManager, f *DdNode, c *DdNode, table *st__table, cache *st__table) int32 {
    var Fv, Fnv, Cv, Cnv *DdNode;
    var one, zero *DdNode;
    unsigned int topf, topc;
    var comple int32;
    var resT, resE, res, retval int32;
    var slot *string;
    var key *MarkCacheKey;

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Terminal cases. */
    if c == zero { return(DD_LIC_DC); }
    if f == one {  return(DD_LIC_1); }
    if f == zero { return(DD_LIC_0); }

    /* Make canonical to increase the utilization of the cache. */
    // inlining of macro Cudd_IsComplement
    comple = (int32( ((ptrint) (f) & 01)));
    // inlining of macro Cudd_Regular
    f = ((DdNode *)((ptruint)(f) & ~01));
    /* Now f is a regular pointer to a non-constant node; c may be
    ** constant, or it may be complemented.
    */

    /* Check the cache. */
    key = new(MarkCacheKey);
    if key == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(CUDD_OUT_OF_MEM);
    }
    key.f = f; key.c = c;
    if st__lookup_int(cache, (char *)key, &res) {
        ABC_FREE(key);
        if comple != 0 {
            if res == DD_LIC_0 { res = DD_LIC_1;
            } else if res == DD_LIC_1 { res = DD_LIC_0; }
        }
        return(res);
    }

    /* Recursive step. */
    topf = dd.perm[f.index];
    // inlining of macro cuddI
    
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(c) & ~01)).index)==CUDD_CONST_INDEX {
        // inlining of macro Cudd_Regular
        topc = int32((((DdNode *)((ptruint)(c) & ~01)).index));
    } else {
        // inlining of macro Cudd_Regular
        topc = (dd).perm[(((DdNode *)((ptruint)(c) & ~01)).index)];
    }
    if topf <= topc {
        // inlining of macro cuddT
        // inlining of macro cuddE
        Fv = ((f).type.kids.T); Fnv = ((f).type.kids.E);
    } else {
        Fv = Fnv = f;
    }
    if topc <= topf {
        /* We know that c is not constant because f is not. */
        // inlining of macro cuddT
        // inlining of macro Cudd_Regular
        // inlining of macro cuddE
        Cv = ((((DdNode *)((ptruint)(c) & ~01))).type.kids.T); Cnv = ((((DdNode *)((ptruint)(c) & ~01))).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (c) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Cv = ((DdNode *)((ptrint)(Cv) ^ 01));
            // inlining of macro Cudd_Not
            Cnv = ((DdNode *)((ptrint)(Cnv) ^ 01));
        }
    } else {
        Cv = Cnv = c;
    }

    resT = cuddBddLICMarkEdges(dd, Fv, Cv, table, cache);
    if resT == CUDD_OUT_OF_MEM {
        ABC_FREE(key);
        return(CUDD_OUT_OF_MEM);
    }
    resE = cuddBddLICMarkEdges(dd, Fnv, Cnv, table, cache);
    if resE == CUDD_OUT_OF_MEM {
        ABC_FREE(key);
        return(CUDD_OUT_OF_MEM);
    }

    /* Update edge markings. */
    if topf <= topc {
        retval = st__find_or_add(table, (char *)f, (char ***)&slot);
        if retval == 0 {
            *slot = (char *) (ptrint)((resT << 2) | resE);
        } else if retval == 1 {
            *slot = (char *) (ptrint)(int32(((ptrint) *slot)) | (resT << 2) | resE);
        } else {
            ABC_FREE(key);
            return(CUDD_OUT_OF_MEM);
        }
    }

    /* Cache result. */
    res = resT | resE;
    if st__insert(cache, (char *)key, (char *)(ptrint)res) == st__OUT_OF_MEM {
        ABC_FREE(key);
        return(CUDD_OUT_OF_MEM);
    }

    /* Take into account possible complementation. */
    if comple != 0 {
        if res == DD_LIC_0 { res = DD_LIC_1;
        } else if res == DD_LIC_1 { res = DD_LIC_0; }
    }
    return(res);

} /* end of cuddBddLICMarkEdges */


/**Function********************************************************************

  Synopsis    [Builds the result of Cudd_bddLICompaction.]

  Description [Builds the results of Cudd_bddLICompaction.
  Returns a pointer to the minimized BDD if successful; otherwise NULL.]

  SideEffects [None]

  SeeAlso     [Cudd_bddLICompaction cuddBddLICMarkEdges]

******************************************************************************/
func cuddBddLICBuildResult(dd *DdManager, f *DdNode, cache *st__table, table *st__table) *DdNode {
    var Fv, Fnv, r, t, e *DdNode;
    var one, zero *DdNode;
    var index int32;
    var comple int32;
    var markT, markE, markings int32;

    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX { return(f); }
    /* Make canonical to increase the utilization of the cache. */
    // inlining of macro Cudd_IsComplement
    comple = (int32( ((ptrint) (f) & 01)));
    // inlining of macro Cudd_Regular
    f = ((DdNode *)((ptruint)(f) & ~01));

    /* Check the cache. */
    if st__lookup(cache, (char *)f, (char **)&r) {
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    }

    /* Retrieve the edge markings. */
    if st__lookup_int(table, (char *)f, &markings) == 0 {
        return(nil); }
    markT = markings >> 2;
    markE = markings & 3;

    index = f.index;
    // inlining of macro cuddT
    // inlining of macro cuddE
    Fv = ((f).type.kids.T); Fnv = ((f).type.kids.E);

    if markT == DD_LIC_NL {
        t = cuddBddLICBuildResult(dd,Fv,cache,table);
        if t == nil {
            return(nil);
        }
    } else if markT == DD_LIC_1 {
        t = one;
    } else {
        t = zero;
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);
    if markE == DD_LIC_NL {
        e = cuddBddLICBuildResult(dd,Fnv,cache,table);
        if e == nil {
            Cudd_IterDerefBdd(dd,t);
            return(nil);
        }
    } else if markE == DD_LIC_1 {
        e = one;
    } else {
        e = zero;
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    if markT == DD_LIC_DC && markE != DD_LIC_DC {
        r = e;
    } else if markT != DD_LIC_DC && markE == DD_LIC_DC {
        r = t;
    } else {
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (t) & 01)) != 0 {
            // inlining of macro Cudd_Not
            t = ((DdNode *)((ptrint)(t) ^ 01));
            // inlining of macro Cudd_Not
            e = ((DdNode *)((ptrint)(e) ^ 01));
            if t == e {
                r = t;
            } else {
                r = cuddUniqueInter(dd, index, t, e);
            }
            if r == nil {
                Cudd_IterDerefBdd(dd, e);
                Cudd_IterDerefBdd(dd, t);
                return(nil);
            }
            // inlining of macro Cudd_Not
            r = ((DdNode *)((ptrint)(r) ^ 01));
        } else {
            if t == e {
                r = t;
            } else {
                r = cuddUniqueInter(dd, index, t, e);
            }
            if r == nil {
                Cudd_IterDerefBdd(dd, e);
                Cudd_IterDerefBdd(dd, t);
                return(nil);
            }
        }
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);

    if st__insert(cache, (char *)f, (char *)r) == st__OUT_OF_MEM {
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(r).ref);
        Cudd_IterDerefBdd(dd,r);
        return(nil);
    }

    // inlining of macro Cudd_NotCond
    return(((DdNode *)((ptrint)(r) ^ (comple))));

} /* end of cuddBddLICBuildResult */


/**Function********************************************************************

  Synopsis    [Hash function for the computed table of cuddBddLICMarkEdges.]

  Description [Hash function for the computed table of
  cuddBddLICMarkEdges.  Returns the bucket number.]

  SideEffects [None]

  SeeAlso     [Cudd_bddLICompaction]

******************************************************************************/
func MarkCacheHash(ptr string, modulus int32) int32 {
    var val int32;
    var entry *MarkCacheKey;

    entry = (MarkCacheKey *) ptr;

    val = int32( (ptrint)) entry.f;
    val = val * 997 + int32( (ptrint)) entry.c;

    return ((val < 0) ? -val : val) % modulus;

} /* end of MarkCacheHash */


/**Function********************************************************************

  Synopsis    [Comparison function for the computed table of
  cuddBddLICMarkEdges.]

  Description [Comparison function for the computed table of
  cuddBddLICMarkEdges. Returns 0 if the two nodes of the key are equal; 1
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddLICompaction]

******************************************************************************/
func MarkCacheCompare(ptr1 string, ptr2 string) int32 {
    var entry1, entry2 *MarkCacheKey;

    entry1 = (MarkCacheKey *) ptr1;
    entry2 = (MarkCacheKey *) ptr2;
    
    return((entry1.f != entry2.f) || (entry1.c != entry2.c));

} /* end of MarkCacheCompare */



/**Function********************************************************************

  Synopsis    [Frees memory associated with computed table of
  cuddBddLICMarkEdges.]

  Description [Frees memory associated with computed table of
  cuddBddLICMarkEdges. Returns st__CONTINUE.]

  SideEffects [None]

  SeeAlso     [Cudd_bddLICompaction]

******************************************************************************/
func MarkCacheCleanUp(key string, value string, arg string) st__retval {
    var entry *MarkCacheKey;

    entry = (MarkCacheKey *) key;
    ABC_FREE(entry);
    return st__CONTINUE;

} /* end of MarkCacheCleanUp */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddSqueeze.]

  Description [Performs the recursive step of Cudd_bddSqueeze.  This
  procedure exploits the fact that if we complement and swap the
  bounds of the interval we obtain a valid solution by taking the
  complement of the solution to the original problem. Therefore, we
  can enforce the condition that the upper bound is always regular.
  Returns a pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddSqueeze]

******************************************************************************/
func cuddBddSqueeze(dd *DdManager, l *DdNode, u *DdNode) *DdNode {
    var one, zero, r, lt, le, ut, ue, t, e *DdNode;
     
               
      
    var comple int32;
    unsigned int topu, topl;
    var index int32;

    statLine(dd);
    if l == u {
        return(l);
    }
    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));
    /* The only case when l == zero && u == one is at the top level,
    ** where returning either one or zero is OK. In all other cases
    ** the procedure will detect such a case and will perform
    ** remapping. Therefore the order in which we test l and u at this
    ** point is immaterial. */
    if l == zero { return(l); }
    if u == one {  return(u); }

    /* Make canonical to increase the utilization of the cache. */
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (u) & 01)) != 0 {
        var temp *DdNode;
        // inlining of macro Cudd_Not
        temp = ((DdNode *)((ptrint)(l) ^ 01));
        // inlining of macro Cudd_Not
        l = ((DdNode *)((ptrint)(u) ^ 01));
        u = temp;
        comple = 1;
    }
    /* At this point u is regular and non-constant; l is non-constant, but
    ** may be complemented. */

    /* Here we could check the relative sizes. */

    /* Check the cache. */
    r = cuddCacheLookup2(dd, Cudd_bddSqueeze, l, u);
    if r != nil {
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    }

    /* Recursive step. */
    topu = dd.perm[u.index];
    // inlining of macro Cudd_Regular
    topl = dd.perm[((DdNode *)((ptruint)(l) & ~01)).index];
    if topu <= topl {
        index = u.index;
        // inlining of macro cuddT
        // inlining of macro cuddE
        ut = ((u).type.kids.T); ue = ((u).type.kids.E);
    } else {
        // inlining of macro Cudd_Regular
        index = ((DdNode *)((ptruint)(l) & ~01)).index;
        ut = ue = u;
    }
    if topl <= topu {
        // inlining of macro cuddT
        // inlining of macro Cudd_Regular
        // inlining of macro cuddE
        lt = ((((DdNode *)((ptruint)(l) & ~01))).type.kids.T); le = ((((DdNode *)((ptruint)(l) & ~01))).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (l) & 01)) != 0 {
            // inlining of macro Cudd_Not
            lt = ((DdNode *)((ptrint)(lt) ^ 01));
            // inlining of macro Cudd_Not
            le = ((DdNode *)((ptrint)(le) ^ 01));
        }
    } else {
        lt = le = l;
    }

    /* If one interval is contained in the other, use the smaller
    ** interval. This corresponds to one-sided matching. */
    if (lt == zero || Cudd_bddLeq(dd,lt,le)) &&
        (ut == one  || Cudd_bddLeq(dd,ue,ut)) { /* remap */
        r = cuddBddSqueeze(dd, le, ue);
        if r == nil {
            return(nil); }
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    } else if (le == zero || Cudd_bddLeq(dd,le,lt)) &&
               (ue == one  || Cudd_bddLeq(dd,ut,ue)) { /* remap */
        r = cuddBddSqueeze(dd, lt, ut);
        if r == nil {
            return(nil); }
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    // inlining of macro Cudd_Not
    } else if (le == zero || Cudd_bddLeq(dd,le,((DdNode *)((ptrint)(ut) ^ 01)))) &&
               // inlining of macro Cudd_Not
               (ue == one  || Cudd_bddLeq(dd,((DdNode *)((ptrint)(lt) ^ 01)),ue)) { /* c-remap */
        t = cuddBddSqueeze(dd, lt, ut);
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(t).ref);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (t) & 01)) != 0 {
            // inlining of macro Cudd_Not
            r = cuddUniqueInter(dd, index, ((DdNode *)((ptrint)(t) ^ 01)), t);
            if r == nil {
                Cudd_IterDerefBdd(dd, t);
                return(nil);
            }
            // inlining of macro Cudd_Not
            r = ((DdNode *)((ptrint)(r) ^ 01));
        } else {
            // inlining of macro Cudd_Not
            r = cuddUniqueInter(dd, index, t, ((DdNode *)((ptrint)(t) ^ 01)));
            if r == nil {
                Cudd_IterDerefBdd(dd, t);
                return(nil);
            }
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(t).ref);
        if r == nil {
            return(nil); }
        cuddCacheInsert2(dd, Cudd_bddSqueeze, l, u, r);
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    // inlining of macro Cudd_Not
    } else if (lt == zero || Cudd_bddLeq(dd,lt,((DdNode *)((ptrint)(ue) ^ 01)))) &&
               // inlining of macro Cudd_Not
               (ut == one  || Cudd_bddLeq(dd,((DdNode *)((ptrint)(le) ^ 01)),ut)) { /* c-remap */
        e = cuddBddSqueeze(dd, le, ue);
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(e).ref);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (e) & 01)) != 0 {
            // inlining of macro Cudd_Not
            r = cuddUniqueInter(dd, index, ((DdNode *)((ptrint)(e) ^ 01)), e);
            if r == nil {
                Cudd_IterDerefBdd(dd, e);
                return(nil);
            }
        } else {
            // inlining of macro Cudd_Not
            r = cuddUniqueInter(dd, index, e, ((DdNode *)((ptrint)(e) ^ 01)));
            if r == nil {
                Cudd_IterDerefBdd(dd, e);
                return(nil);
            }
            // inlining of macro Cudd_Not
            r = ((DdNode *)((ptrint)(r) ^ 01));
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(e).ref);
        if r == nil {
            return(nil); }
        cuddCacheInsert2(dd, Cudd_bddSqueeze, l, u, r);
        // inlining of macro Cudd_NotCond
        return(((DdNode *)((ptrint)(r) ^ (comple))));
    }

     
                                                           
                                                                  
                                                        
                                                                   
                    
                                                         
                        
                                       
                       
                         
                    
                                                           
                         
                                     
                         
         
                    
                          
                                        
                         
                                     
                                     
                         
         
                    
                                 
                                 
            
                  
     
      

    t = cuddBddSqueeze(dd, lt, ut);
    if t == nil {
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);
    e = cuddBddSqueeze(dd, le, ue);
    if e == nil {
        Cudd_IterDerefBdd(dd,t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (t) & 01)) != 0 {
        // inlining of macro Cudd_Not
        t = ((DdNode *)((ptrint)(t) ^ 01));
        // inlining of macro Cudd_Not
        e = ((DdNode *)((ptrint)(e) ^ 01));
        if t == e {
            r = t;
        } else {
            r = cuddUniqueInter(dd, index, t, e);
        }
        if r == nil {
            Cudd_IterDerefBdd(dd, e);
            Cudd_IterDerefBdd(dd, t);
            return(nil);
        }
        // inlining of macro Cudd_Not
        r = ((DdNode *)((ptrint)(r) ^ 01));
    } else {
        if t == e {
            r = t;
        } else {
            r = cuddUniqueInter(dd, index, t, e);
        }
        if r == nil {
            Cudd_IterDerefBdd(dd, e);
            Cudd_IterDerefBdd(dd, t);
            return(nil);
        }
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(t).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(e).ref);

     
                                                                          
                                                          
               
                     
                                                  
                                     
                   
                
                                      
         
     
                 
      

    cuddCacheInsert2(dd, Cudd_bddSqueeze, l, u, r);
    // inlining of macro Cudd_NotCond
    return(((DdNode *)((ptrint)(r) ^ (comple))));

} /* end of cuddBddSqueeze */


//ABC_NAMESPACE_IMPL_END


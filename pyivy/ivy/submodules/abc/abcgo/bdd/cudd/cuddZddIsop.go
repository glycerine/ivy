package cudd

import (
	"fmt"
	"os"
)

/**CFile***********************************************************************

  FileName    [cuddZddIsop.c]

  PackageName [cudd]

  Synopsis    [Functions to find irredundant SOP covers as ZDDs from BDDs.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_bddIsop()
                    <li> Cudd_zddIsop()
                    <li> Cudd_MakeBddFromZddCover()
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    <li> cuddBddIsop()
                    <li> cuddZddIsop()
                    <li> cuddMakeBddFromZddCover()
                    </ul>
               Static procedures included in this module:
                    <ul>
                    </ul>
              ]

  SeeAlso     []

  Author      [In-Ho Moon]

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
static char rcsid[] DD_UNUSED = "$Id: cuddZddIsop.c,v 1.20 2009/02/19 16:26:12 fabio Exp $";
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

  Synopsis    [Computes an ISOP in ZDD form from BDDs.]

  Description [Computes an irredundant sum of products (ISOP) in ZDD
  form from BDDs. The two BDDs L and U represent the lower bound and
  the upper bound, respectively, of the function. The ISOP uses two
  ZDD variables for each BDD variable: One for the positive literal,
  and one for the negative literal. These two variables should be
  adjacent in the ZDD order. The two ZDD variables corresponding to
  BDD variable <code>i</code> should have indices <code>2i</code> and
  <code>2i+1</code>.  The result of this procedure depends on the
  variable order. If successful, Cudd_zddIsop returns the BDD for
  the function chosen from the interval. The ZDD representing the
  irredundant cover is returned as a side effect in zdd_I. In case of
  failure, NULL is returned.]

  SideEffects [zdd_I holds the pointer to the ZDD for the ISOP on
  successful return.]

  SeeAlso     [Cudd_bddIsop Cudd_zddVarsFromBddVars]

******************************************************************************/
func Cudd_zddIsop(dd *DdManager, L *DdNode, U *DdNode, zdd_I **DdNode) *DdNode {
    var res *DdNode;
    var autoDynZ int32;

    autoDynZ = dd.autoDynZ;
    dd.autoDynZ = 0;

    for {
        dd.reordered = 0;
        res = cuddZddIsop(dd, L, U, zdd_I);
    
if !(dd.reordered == 1) { break }
} 
    dd.autoDynZ = autoDynZ;
    return(res);

} /* end of Cudd_zddIsop */


/**Function********************************************************************

  Synopsis    [Computes a BDD in the interval between L and U with a
  simple sum-of-produuct cover.]

  Description [Computes a BDD in the interval between L and U with a
  simple sum-of-produuct cover. This procedure is similar to
  Cudd_zddIsop, but it does not return the ZDD for the cover. Returns
  a pointer to the BDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_zddIsop]

******************************************************************************/
func Cudd_bddIsop(dd *DdManager, L *DdNode, U *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddBddIsop(dd, L, U);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_bddIsop */


/**Function********************************************************************

  Synopsis [Converts a ZDD cover to a BDD graph.]

  Description [Converts a ZDD cover to a BDD graph. If successful, it
  returns a BDD node, otherwise it returns NULL.]

  SideEffects []

  SeeAlso     [cuddMakeBddFromZddCover]

******************************************************************************/
func Cudd_MakeBddFromZddCover(dd *DdManager, node *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddMakeBddFromZddCover(dd, node);
    
if !(dd.reordered == 1) { break }
} 
    return(res);
} /* end of Cudd_MakeBddFromZddCover */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_zddIsop.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_zddIsop]

******************************************************************************/
func cuddZddIsop(dd *DdManager, L *DdNode, U *DdNode, zdd_I **DdNode) *DdNode {
    // inlining of macro DD_ONE
    var one *DdNode  = ((dd).one);
    // inlining of macro Cudd_Not
    var zero *DdNode  = ((DdNode *)((ptrint)(one) ^ 01));
    // inlining of macro DD_ONE
    var zdd_one *DdNode  = ((dd).one);
    // inlining of macro DD_ZERO
    var zdd_zero *DdNode  = ((dd).zero);
    var v, top_l, top_u int32;
    var Lsub0, Usub0, Lsub1, Usub1, Ld, Ud *DdNode;
    var Lsuper0, Usuper0, Lsuper1, Usuper1 *DdNode;
    var Isub0, Isub1, Id *DdNode;
    var zdd_Isub0, zdd_Isub1, zdd_Id *DdNode;
    var x *DdNode;
    var term0, term1, sum *DdNode;
    var Lv, Uv, Lnv, Unv *DdNode;
    var r, y, z *DdNode;
    var index int32;
    DD_CTFP     cacheOp;

    statLine(dd);
    if L == zero {
        *zdd_I = zdd_zero;
        return(zero);
    }
    if U == one {
        *zdd_I = zdd_one;
        return(one);
    }

    if U == zero || L == one {
        fmt.Printf("*** ERROR : illegal condition for ISOP (U < L).\n");
        os.Exit(int(1));
    }

    /* Check the cache. We store two results for each recursive call.
    ** One is the BDD, and the other is the ZDD. Both are needed.
    ** Hence we need a double hit in the cache to terminate the
    ** recursion. Clearly, collisions may evict only one of the two
    ** results. */
    cacheOp = (DD_CTFP) cuddZddIsop;
    r = cuddCacheLookup2(dd, cuddBddIsop, L, U);
    if r != nil {
        *zdd_I = cuddCacheLookup2Zdd(dd, cacheOp, L, U);
        if *zdd_I {
            return(r);
        } else {
            /* The BDD result may have been dead. In that case
            ** cuddCacheLookup2 would have called cuddReclaim,
            ** whose effects we now have to undo. */
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(r).ref);
            Cudd_RecursiveDeref(dd, r);
        }
    }

    // inlining of macro Cudd_Regular
    top_l = dd.perm[((DdNode *)((ptruint)(L) & ~01)).index];
    // inlining of macro Cudd_Regular
    top_u = dd.perm[((DdNode *)((ptruint)(U) & ~01)).index];
    // inlining of macro ddMin
    
    if (top_u) < (top_l) {
        v = (top_u);
    } else {
        v = (top_l);
    }

    /* Compute cofactors. */
    if top_l == v {
        // inlining of macro Cudd_Regular
        index = ((DdNode *)((ptruint)(L) & ~01)).index;
        // inlining of macro Cudd_T
        // inlining of macro Cudd_Regular
        Lv = ((((DdNode *)((ptruint)(L) & ~01))).type.kids.T);
        // inlining of macro Cudd_E
        // inlining of macro Cudd_Regular
        Lnv = ((((DdNode *)((ptruint)(L) & ~01))).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (L) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Lv = ((DdNode *)((ptrint)(Lv) ^ 01));
            // inlining of macro Cudd_Not
            Lnv = ((DdNode *)((ptrint)(Lnv) ^ 01));
        }
    
    } else {
        // inlining of macro Cudd_Regular
        index = ((DdNode *)((ptruint)(U) & ~01)).index;
        Lv = Lnv = L;
    }

    if top_u == v {
        // inlining of macro Cudd_T
        // inlining of macro Cudd_Regular
        Uv = ((((DdNode *)((ptruint)(U) & ~01))).type.kids.T);
        // inlining of macro Cudd_E
        // inlining of macro Cudd_Regular
        Unv = ((((DdNode *)((ptruint)(U) & ~01))).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (U) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Uv = ((DdNode *)((ptrint)(Uv) ^ 01));
            // inlining of macro Cudd_Not
            Unv = ((DdNode *)((ptrint)(Unv) ^ 01));
        }
    
    } else {
        Uv = Unv = U;
    }

    // inlining of macro Cudd_Not
    Lsub0 = cuddBddAndRecur(dd, Lnv, ((DdNode *)((ptrint)(Uv) ^ 01)));
    if Lsub0 == nil {
        return(nil); }
    Cudd_Ref(Lsub0);
    Usub0 = Unv;
    // inlining of macro Cudd_Not
    Lsub1 = cuddBddAndRecur(dd, Lv, ((DdNode *)((ptrint)(Unv) ^ 01)));
    if Lsub1 == nil {
        Cudd_RecursiveDeref(dd, Lsub0);
        return(nil);
    }
    Cudd_Ref(Lsub1);
    Usub1 = Uv;

    Isub0 = cuddZddIsop(dd, Lsub0, Usub0, &zdd_Isub0);
    if Isub0 == nil {
        Cudd_RecursiveDeref(dd, Lsub0);
        Cudd_RecursiveDeref(dd, Lsub1);
        return(nil);
    }
    /*
    if ((!cuddIsConstant(Cudd_Regular(Isub0))) &&
        (Cudd_Regular(Isub0)->index != zdd_Isub0->index / 2 ||
        dd->permZ[index * 2] > dd->permZ[zdd_Isub0->index])) {
        printf("*** ERROR : illegal permutation in ZDD. ***\n");
    }
    */
    Cudd_Ref(Isub0);
    Cudd_Ref(zdd_Isub0);
    Isub1 = cuddZddIsop(dd, Lsub1, Usub1, &zdd_Isub1);
    if Isub1 == nil {
        Cudd_RecursiveDeref(dd, Lsub0);
        Cudd_RecursiveDeref(dd, Lsub1);
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
        return(nil);
    }
    /*
    if ((!cuddIsConstant(Cudd_Regular(Isub1))) &&
        (Cudd_Regular(Isub1)->index != zdd_Isub1->index / 2 ||
        dd->permZ[index * 2] > dd->permZ[zdd_Isub1->index])) {
        printf("*** ERROR : illegal permutation in ZDD. ***\n");
    }
    */
    Cudd_Ref(Isub1);
    Cudd_Ref(zdd_Isub1);
    Cudd_RecursiveDeref(dd, Lsub0);
    Cudd_RecursiveDeref(dd, Lsub1);

    // inlining of macro Cudd_Not
    Lsuper0 = cuddBddAndRecur(dd, Lnv, ((DdNode *)((ptrint)(Isub0) ^ 01)));
    if Lsuper0 == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
        return(nil);
    }
    Cudd_Ref(Lsuper0);
    // inlining of macro Cudd_Not
    Lsuper1 = cuddBddAndRecur(dd, Lv, ((DdNode *)((ptrint)(Isub1) ^ 01)));
    if Lsuper1 == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
        Cudd_RecursiveDeref(dd, Lsuper0);
        return(nil);
    }
    Cudd_Ref(Lsuper1);
    Usuper0 = Unv;
    Usuper1 = Uv;

    /* Ld = Lsuper0 + Lsuper1 */
    // inlining of macro Cudd_Not
    Ld = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(Lsuper0) ^ 01)), ((DdNode *)((ptrint)(Lsuper1) ^ 01)));
    if Ld == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
        Cudd_RecursiveDeref(dd, Lsuper0);
        Cudd_RecursiveDeref(dd, Lsuper1);
        return(nil);
    }
    // inlining of macro Cudd_Not
    Ld = ((DdNode *)((ptrint)(Ld) ^ 01));
    Cudd_Ref(Ld);
    /* Ud = Usuper0 * Usuper1 */
    Ud = cuddBddAndRecur(dd, Usuper0, Usuper1);
    if Ud == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
        Cudd_RecursiveDeref(dd, Lsuper0);
        Cudd_RecursiveDeref(dd, Lsuper1);
        Cudd_RecursiveDeref(dd, Ld);
        return(nil);
    }
    Cudd_Ref(Ud);
    Cudd_RecursiveDeref(dd, Lsuper0);
    Cudd_RecursiveDeref(dd, Lsuper1);

    Id = cuddZddIsop(dd, Ld, Ud, &zdd_Id);
    if Id == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
        Cudd_RecursiveDeref(dd, Ld);
        Cudd_RecursiveDeref(dd, Ud);
        return(nil);
    }
    /*
    if ((!cuddIsConstant(Cudd_Regular(Id))) &&
        (Cudd_Regular(Id)->index != zdd_Id->index / 2 ||
        dd->permZ[index * 2] > dd->permZ[zdd_Id->index])) {
        printf("*** ERROR : illegal permutation in ZDD. ***\n");
    }
    */
    Cudd_Ref(Id);
    Cudd_Ref(zdd_Id);
    Cudd_RecursiveDeref(dd, Ld);
    Cudd_RecursiveDeref(dd, Ud);

    x = cuddUniqueInter(dd, index, one, zero);
    if x == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDerefZdd(dd, zdd_Id);
        return(nil);
    }
    Cudd_Ref(x);
    /* term0 = x * Isub0 */
    // inlining of macro Cudd_Not
    term0 = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(x) ^ 01)), Isub0);
    if term0 == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDerefZdd(dd, zdd_Id);
        Cudd_RecursiveDeref(dd, x);
        return(nil);
    }
    Cudd_Ref(term0);
    Cudd_RecursiveDeref(dd, Isub0);
    /* term1 = x * Isub1 */
    term1 = cuddBddAndRecur(dd, x, Isub1);
    if term1 == nil {
        Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDerefZdd(dd, zdd_Id);
        Cudd_RecursiveDeref(dd, x);
        Cudd_RecursiveDeref(dd, term0);
        return(nil);
    }
    Cudd_Ref(term1);
    Cudd_RecursiveDeref(dd, x);
    Cudd_RecursiveDeref(dd, Isub1);
    /* sum = term0 + term1 */
    // inlining of macro Cudd_Not
    sum = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(term0) ^ 01)), ((DdNode *)((ptrint)(term1) ^ 01)));
    if sum == nil {
        Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDerefZdd(dd, zdd_Id);
        Cudd_RecursiveDeref(dd, term0);
        Cudd_RecursiveDeref(dd, term1);
        return(nil);
    }
    // inlining of macro Cudd_Not
    sum = ((DdNode *)((ptrint)(sum) ^ 01));
    Cudd_Ref(sum);
    Cudd_RecursiveDeref(dd, term0);
    Cudd_RecursiveDeref(dd, term1);
    /* r = sum + Id */
    // inlining of macro Cudd_Not
    r = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(sum) ^ 01)), ((DdNode *)((ptrint)(Id) ^ 01)));
    // inlining of macro Cudd_NotCond
    r = ((DdNode *)((ptrint)(r) ^ (r != nil)));
    if r == nil {
        Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
        Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDerefZdd(dd, zdd_Id);
        Cudd_RecursiveDeref(dd, sum);
        return(nil);
    }
    Cudd_Ref(r);
    Cudd_RecursiveDeref(dd, sum);
    Cudd_RecursiveDeref(dd, Id);

    if zdd_Isub0 != zdd_zero {
        z = cuddZddGetNodeIVO(dd, index * 2 + 1, zdd_Isub0, zdd_Id);
        if z == nil {
            Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
            Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
            Cudd_RecursiveDerefZdd(dd, zdd_Id);
            Cudd_RecursiveDeref(dd, r);
            return(nil);
        }
    
    } else {
        z = zdd_Id;
    }
    Cudd_Ref(z);
    if zdd_Isub1 != zdd_zero {
        y = cuddZddGetNodeIVO(dd, index * 2, zdd_Isub1, z);
        if y == nil {
            Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
            Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
            Cudd_RecursiveDerefZdd(dd, zdd_Id);
            Cudd_RecursiveDeref(dd, r);
            Cudd_RecursiveDerefZdd(dd, z);
            return(nil);
        }
    
    } else {
        y = z; }
    Cudd_Ref(y);

    Cudd_RecursiveDerefZdd(dd, zdd_Isub0);
    Cudd_RecursiveDerefZdd(dd, zdd_Isub1);
    Cudd_RecursiveDerefZdd(dd, zdd_Id);
    Cudd_RecursiveDerefZdd(dd, z);

    cuddCacheInsert2(dd, cuddBddIsop, L, U, r);
    cuddCacheInsert2(dd, cacheOp, L, U, y);

    Cudd_Deref(r);
    Cudd_Deref(y);
    *zdd_I = y;
    /*
    if (Cudd_Regular(r)->index != y->index / 2) {
        printf("*** ERROR : mismatch in indices between BDD and ZDD. ***\n");
    }
    */
    return(r);

} /* end of cuddZddIsop */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_bddIsop.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_bddIsop]

******************************************************************************/
func cuddBddIsop(dd *DdManager, L *DdNode, U *DdNode) *DdNode {
    // inlining of macro DD_ONE
    var one *DdNode  = ((dd).one);
    // inlining of macro Cudd_Not
    var zero *DdNode  = ((DdNode *)((ptrint)(one) ^ 01));
    var v, top_l, top_u int32;
    var Lsub0, Usub0, Lsub1, Usub1, Ld, Ud *DdNode;
    var Lsuper0, Usuper0, Lsuper1, Usuper1 *DdNode;
    var Isub0, Isub1, Id *DdNode;
    var x *DdNode;
    var term0, term1, sum *DdNode;
    var Lv, Uv, Lnv, Unv *DdNode;
    var r *DdNode;
    var index int32;

    statLine(dd);
    if L == zero {
        return(zero); }
    if U == one {
        return(one); }

    /* Check cache */
    r = cuddCacheLookup2(dd, cuddBddIsop, L, U);
    if r != nil {
        return(r); }

    // inlining of macro Cudd_Regular
    top_l = dd.perm[((DdNode *)((ptruint)(L) & ~01)).index];
    // inlining of macro Cudd_Regular
    top_u = dd.perm[((DdNode *)((ptruint)(U) & ~01)).index];
    // inlining of macro ddMin
    
    if (top_u) < (top_l) {
        v = (top_u);
    } else {
        v = (top_l);
    }

    /* Compute cofactors */
    if top_l == v {
        // inlining of macro Cudd_Regular
        index = ((DdNode *)((ptruint)(L) & ~01)).index;
        // inlining of macro Cudd_T
        // inlining of macro Cudd_Regular
        Lv = ((((DdNode *)((ptruint)(L) & ~01))).type.kids.T);
        // inlining of macro Cudd_E
        // inlining of macro Cudd_Regular
        Lnv = ((((DdNode *)((ptruint)(L) & ~01))).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (L) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Lv = ((DdNode *)((ptrint)(Lv) ^ 01));
            // inlining of macro Cudd_Not
            Lnv = ((DdNode *)((ptrint)(Lnv) ^ 01));
        }
    
    } else {
        // inlining of macro Cudd_Regular
        index = ((DdNode *)((ptruint)(U) & ~01)).index;
        Lv = Lnv = L;
    }

    if top_u == v {
        // inlining of macro Cudd_T
        // inlining of macro Cudd_Regular
        Uv = ((((DdNode *)((ptruint)(U) & ~01))).type.kids.T);
        // inlining of macro Cudd_E
        // inlining of macro Cudd_Regular
        Unv = ((((DdNode *)((ptruint)(U) & ~01))).type.kids.E);
        // inlining of macro Cudd_IsComplement
        if int32( ((ptrint) (U) & 01)) != 0 {
            // inlining of macro Cudd_Not
            Uv = ((DdNode *)((ptrint)(Uv) ^ 01));
            // inlining of macro Cudd_Not
            Unv = ((DdNode *)((ptrint)(Unv) ^ 01));
        }
    
    } else {
        Uv = Unv = U;
    }

    // inlining of macro Cudd_Not
    Lsub0 = cuddBddAndRecur(dd, Lnv, ((DdNode *)((ptrint)(Uv) ^ 01)));
    if Lsub0 == nil {
        return(nil); }
    Cudd_Ref(Lsub0);
    Usub0 = Unv;
    // inlining of macro Cudd_Not
    Lsub1 = cuddBddAndRecur(dd, Lv, ((DdNode *)((ptrint)(Unv) ^ 01)));
    if Lsub1 == nil {
        Cudd_RecursiveDeref(dd, Lsub0);
        return(nil);
    }
    Cudd_Ref(Lsub1);
    Usub1 = Uv;

    Isub0 = cuddBddIsop(dd, Lsub0, Usub0);
    if Isub0 == nil {
        Cudd_RecursiveDeref(dd, Lsub0);
        Cudd_RecursiveDeref(dd, Lsub1);
        return(nil);
    }
    Cudd_Ref(Isub0);
    Isub1 = cuddBddIsop(dd, Lsub1, Usub1);
    if Isub1 == nil {
        Cudd_RecursiveDeref(dd, Lsub0);
        Cudd_RecursiveDeref(dd, Lsub1);
        Cudd_RecursiveDeref(dd, Isub0);
        return(nil);
    }
    Cudd_Ref(Isub1);
    Cudd_RecursiveDeref(dd, Lsub0);
    Cudd_RecursiveDeref(dd, Lsub1);

    // inlining of macro Cudd_Not
    Lsuper0 = cuddBddAndRecur(dd, Lnv, ((DdNode *)((ptrint)(Isub0) ^ 01)));
    if Lsuper0 == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        return(nil);
    }
    Cudd_Ref(Lsuper0);
    // inlining of macro Cudd_Not
    Lsuper1 = cuddBddAndRecur(dd, Lv, ((DdNode *)((ptrint)(Isub1) ^ 01)));
    if Lsuper1 == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Lsuper0);
        return(nil);
    }
    Cudd_Ref(Lsuper1);
    Usuper0 = Unv;
    Usuper1 = Uv;

    /* Ld = Lsuper0 + Lsuper1 */
    // inlining of macro Cudd_Not
    Ld = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(Lsuper0) ^ 01)), ((DdNode *)((ptrint)(Lsuper1) ^ 01)));
    // inlining of macro Cudd_NotCond
    Ld = ((DdNode *)((ptrint)(Ld) ^ (Ld != nil)));
    if Ld == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Lsuper0);
        Cudd_RecursiveDeref(dd, Lsuper1);
        return(nil);
    }
    Cudd_Ref(Ld);
    Ud = cuddBddAndRecur(dd, Usuper0, Usuper1);
    if Ud == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Lsuper0);
        Cudd_RecursiveDeref(dd, Lsuper1);
        Cudd_RecursiveDeref(dd, Ld);
        return(nil);
    }
    Cudd_Ref(Ud);
    Cudd_RecursiveDeref(dd, Lsuper0);
    Cudd_RecursiveDeref(dd, Lsuper1);

    Id = cuddBddIsop(dd, Ld, Ud);
    if Id == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Ld);
        Cudd_RecursiveDeref(dd, Ud);
        return(nil);
    }
    Cudd_Ref(Id);
    Cudd_RecursiveDeref(dd, Ld);
    Cudd_RecursiveDeref(dd, Ud);

    x = cuddUniqueInter(dd, index, one, zero);
    if x == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Id);
        return(nil);
    }
    Cudd_Ref(x);
    // inlining of macro Cudd_Not
    term0 = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(x) ^ 01)), Isub0);
    if term0 == nil {
        Cudd_RecursiveDeref(dd, Isub0);
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDeref(dd, x);
        return(nil);
    }
    Cudd_Ref(term0);
    Cudd_RecursiveDeref(dd, Isub0);
    term1 = cuddBddAndRecur(dd, x, Isub1);
    if term1 == nil {
        Cudd_RecursiveDeref(dd, Isub1);
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDeref(dd, x);
        Cudd_RecursiveDeref(dd, term0);
        return(nil);
    }
    Cudd_Ref(term1);
    Cudd_RecursiveDeref(dd, x);
    Cudd_RecursiveDeref(dd, Isub1);
    /* sum = term0 + term1 */
    // inlining of macro Cudd_Not
    sum = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(term0) ^ 01)), ((DdNode *)((ptrint)(term1) ^ 01)));
    // inlining of macro Cudd_NotCond
    sum = ((DdNode *)((ptrint)(sum) ^ (sum != nil)));
    if sum == nil {
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDeref(dd, term0);
        Cudd_RecursiveDeref(dd, term1);
        return(nil);
    }
    Cudd_Ref(sum);
    Cudd_RecursiveDeref(dd, term0);
    Cudd_RecursiveDeref(dd, term1);
    /* r = sum + Id */
    // inlining of macro Cudd_Not
    r = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(sum) ^ 01)), ((DdNode *)((ptrint)(Id) ^ 01)));
    // inlining of macro Cudd_NotCond
    r = ((DdNode *)((ptrint)(r) ^ (r != nil)));
    if r == nil {
        Cudd_RecursiveDeref(dd, Id);
        Cudd_RecursiveDeref(dd, sum);
        return(nil);
    }
    Cudd_Ref(r);
    Cudd_RecursiveDeref(dd, sum);
    Cudd_RecursiveDeref(dd, Id);

    cuddCacheInsert2(dd, cuddBddIsop, L, U, r);

    Cudd_Deref(r);
    return(r);

} /* end of cuddBddIsop */


/**Function********************************************************************

  Synopsis [Converts a ZDD cover to a BDD graph.]

  Description [Converts a ZDD cover to a BDD graph. If successful, it
  returns a BDD node, otherwise it returns NULL. It is a recursive
  algorithm as the following. First computes 3 cofactors of a ZDD cover;
  f1, f0 and fd. Second, compute BDDs(b1, b0 and bd) of f1, f0 and fd.
  Third, compute T=b1+bd and E=b0+bd. Fourth, compute ITE(v,T,E) where v
  is the variable which has the index of the top node of the ZDD cover.
  In this case, since the index of v can be larger than either one of T or
  one of E, cuddUniqueInterIVO is called, here IVO stands for
  independent variable ordering.]

  SideEffects []

  SeeAlso     [Cudd_MakeBddFromZddCover]

******************************************************************************/
func cuddMakeBddFromZddCover(dd *DdManager, node *DdNode) *DdNode {
    var neW *DdNode;
    var v int32;
    var f1, f0, fd *DdNode;
    var b1, b0, bd *DdNode;
    var T, E *DdNode;

    statLine(dd);
    if node == dd.one {
        return(dd.one); }
    if node == dd.zero {
        // inlining of macro Cudd_Not
        return(((DdNode *)((ptrint)(dd.one) ^ 01))); }

    /* Check cache */
    neW = cuddCacheLookup1(dd, cuddMakeBddFromZddCover, node);
    if neW != nil {
        return(neW); }

    // inlining of macro Cudd_Regular
    v = ((DdNode *)((ptruint)(node) & ~01)).index;      /* either yi or zi */
    if cuddZddGetCofactors3(dd, node, v, &f1, &f0, &fd) { return(nil); }
    Cudd_Ref(f1);
    Cudd_Ref(f0);
    Cudd_Ref(fd);

    b1 = cuddMakeBddFromZddCover(dd, f1);
    if !b1 {
        Cudd_RecursiveDerefZdd(dd, f1);
        Cudd_RecursiveDerefZdd(dd, f0);
        Cudd_RecursiveDerefZdd(dd, fd);
        return(nil);
    }
    Cudd_Ref(b1);
    b0 = cuddMakeBddFromZddCover(dd, f0);
    if !b0 {
        Cudd_RecursiveDerefZdd(dd, f1);
        Cudd_RecursiveDerefZdd(dd, f0);
        Cudd_RecursiveDerefZdd(dd, fd);
        Cudd_RecursiveDeref(dd, b1);
        return(nil);
    }
    Cudd_Ref(b0);
    Cudd_RecursiveDerefZdd(dd, f1);
    Cudd_RecursiveDerefZdd(dd, f0);
    if fd != dd.zero {
        bd = cuddMakeBddFromZddCover(dd, fd);
        if !(bd != nil) {
            Cudd_RecursiveDerefZdd(dd, fd);
            Cudd_RecursiveDeref(dd, b1);
            Cudd_RecursiveDeref(dd, b0);
            return(nil);
        }
        Cudd_Ref(bd);
        Cudd_RecursiveDerefZdd(dd, fd);

        // inlining of macro Cudd_Not
        T = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(b1) ^ 01)), ((DdNode *)((ptrint)(bd) ^ 01)));
        if !T {
            Cudd_RecursiveDeref(dd, b1);
            Cudd_RecursiveDeref(dd, b0);
            Cudd_RecursiveDeref(dd, bd);
            return(nil);
        }
        // inlining of macro Cudd_NotCond
        T = ((DdNode *)((ptrint)(T) ^ (T != nil)));
        Cudd_Ref(T);
        Cudd_RecursiveDeref(dd, b1);
        // inlining of macro Cudd_Not
        E = cuddBddAndRecur(dd, ((DdNode *)((ptrint)(b0) ^ 01)), ((DdNode *)((ptrint)(bd) ^ 01)));
        if !(E != nil) {
            Cudd_RecursiveDeref(dd, b0);
            Cudd_RecursiveDeref(dd, bd);
            Cudd_RecursiveDeref(dd, T);
            return(nil);
        }
        // inlining of macro Cudd_NotCond
        E = ((DdNode *)((ptrint)(E) ^ (E != nil)));
        Cudd_Ref(E);
        Cudd_RecursiveDeref(dd, b0);
        Cudd_RecursiveDeref(dd, bd);
    
    } else {
        Cudd_RecursiveDerefZdd(dd, fd);
        T = b1;
        E = b0;
    }

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (T) & 01)) != 0 {
        // inlining of macro Cudd_Not
        neW = cuddUniqueInterIVO(dd, v / 2, ((DdNode *)((ptrint)(T) ^ 01)), ((DdNode *)((ptrint)(E) ^ 01)));
        if !(neW != nil) {
            Cudd_RecursiveDeref(dd, T);
            Cudd_RecursiveDeref(dd, E);
            return(nil);
        }
        // inlining of macro Cudd_Not
        neW = ((DdNode *)((ptrint)(neW) ^ 01));
    
    } else {
        neW = cuddUniqueInterIVO(dd, v / 2, T, E);
        if !(neW != nil) {
            Cudd_RecursiveDeref(dd, T);
            Cudd_RecursiveDeref(dd, E);
            return(nil);
        }
    }
    Cudd_Ref(neW);
    Cudd_RecursiveDeref(dd, T);
    Cudd_RecursiveDeref(dd, E);

    cuddCacheInsert1(dd, cuddMakeBddFromZddCover, node, neW);
    Cudd_Deref(neW);
    return(neW);

} /* end of cuddMakeBddFromZddCover */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


//ABC_NAMESPACE_IMPL_END


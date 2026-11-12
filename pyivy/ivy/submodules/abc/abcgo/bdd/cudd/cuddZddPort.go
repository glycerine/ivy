package cudd

/**CFile***********************************************************************

  FileName    [cuddZddPort.c]

  PackageName [cudd]

  Synopsis    [Functions that translate BDDs to ZDDs.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_zddPortFromBdd()
                    <li> Cudd_zddPortToBdd()
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    </ul>
               Static procedures included in this module:
                    <ul>
                    <li> zddPortFromBddStep()
                    <li> zddPortToBddStep()
                    </ul>
              ]

  SeeAlso     []

  Author      [Hyong-kyoon Shin, In-Ho Moon]

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
static char rcsid[] DD_UNUSED = "$Id: cuddZddPort.c,v 1.13 2004/08/13 18:04:53 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static DdNode * zddPortFromBddStep (DdManager *dd, DdNode *B, int expected);

// static DdNode * zddPortToBddStep (DdManager *dd, DdNode *f, int depth);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Converts a BDD into a ZDD.]

  Description [Converts a BDD into a ZDD. This function assumes that
  there is a one-to-one correspondence between the BDD variables and the
  ZDD variables, and that the variable order is the same for both types
  of variables. These conditions are established if the ZDD variables
  are created by one call to Cudd_zddVarsFromBddVars with multiplicity =
  1. Returns a pointer to the resulting ZDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_zddVarsFromBddVars]

******************************************************************************/
func Cudd_zddPortFromBdd(dd *DdManager, B *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = zddPortFromBddStep(dd,B,0);
    
if !(dd.reordered == 1) { break }
} 

    return(res);

} /* end of Cudd_zddPortFromBdd */


/**Function********************************************************************

  Synopsis [Converts a ZDD into a BDD.]

  Description [Converts a ZDD into a BDD. Returns a pointer to the resulting
  ZDD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_zddPortFromBdd]

******************************************************************************/
func Cudd_zddPortToBdd(dd *DdManager, f *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = zddPortToBddStep(dd,f,0);
    
if !(dd.reordered == 1) { break }
} 

    return(res);

} /* end of Cudd_zddPortToBdd */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_zddPortFromBdd.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func zddPortFromBddStep(dd *DdManager, B *DdNode, expected int32) *DdNode {
    var res, prevZdd, t, e *DdNode;
    var Breg, Bt, Be *DdNode;
    var id, level int32;

    statLine(dd);
    /* Terminal cases. */
    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    if B == ((DdNode *)((ptrint)(((dd).one)) ^ 01)) {
        // inlining of macro DD_ZERO
        return(((dd).zero)); }
    // inlining of macro DD_ONE
    if B == ((dd).one) {
        if expected >= dd.sizeZ {
            // inlining of macro DD_ONE
            return(((dd).one));
        } else {
            return(dd.univ[expected]);
        }
    }

    // inlining of macro Cudd_Regular
    Breg = ((DdNode *)((ptruint)(B) & ~01));

    /* Computed table look-up. */
    res = cuddCacheLookup1Zdd(dd,Cudd_zddPortFromBdd,B);
    if res != nil {
        // inlining of macro cuddI
        
        if (Breg.index)==CUDD_CONST_INDEX {
            level = int32((Breg.index));
        } else {
            level = (dd).perm[(Breg.index)];
        }
        /* Adding DC vars. */
        if expected < level {
            /* Add suppressed variables. */
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res).ref);
            for level--; level >= expected; level-- {
                prevZdd = res;
                id = dd.invperm[level];
                res = cuddZddGetNode(dd, id, prevZdd, prevZdd);
                if res == nil {
                    Cudd_RecursiveDerefZdd(dd, prevZdd);
                    return(nil);
                }
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(res).ref);
                Cudd_RecursiveDerefZdd(dd, prevZdd);
            }
            // inlining of macro cuddDeref
            cuddSatDec(Cudd_Regular(res).ref);
        }
        return(res);
    }   /* end of cache look-up */

    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (B) & 01)) != 0 {
        // inlining of macro Cudd_Not
        // inlining of macro cuddT
        Bt = ((DdNode *)((ptrint)(((Breg).type.kids.T)) ^ 01));
        // inlining of macro Cudd_Not
        // inlining of macro cuddE
        Be = ((DdNode *)((ptrint)(((Breg).type.kids.E)) ^ 01));
    } else {
        // inlining of macro cuddT
        Bt = ((Breg).type.kids.T);
        // inlining of macro cuddE
        Be = ((Breg).type.kids.E);
    }

    id = Breg.index;
    // inlining of macro cuddI
    
    if (id)==CUDD_CONST_INDEX {
        level = int32((id));
    } else {
        level = (dd).perm[(id)];
    }
    t = zddPortFromBddStep(dd, Bt, level+1);
    if t == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(t).ref);
    e = zddPortFromBddStep(dd, Be, level+1);
    if e == nil {
        Cudd_RecursiveDerefZdd(dd, t);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(e).ref);
    res = cuddZddGetNode(dd, id, t, e);
    if res == nil {
        Cudd_RecursiveDerefZdd(dd, t);
        Cudd_RecursiveDerefZdd(dd, e);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDerefZdd(dd, t);
    Cudd_RecursiveDerefZdd(dd, e);

    cuddCacheInsert1(dd,Cudd_zddPortFromBdd,B,res);

    for level--; level >= expected; level-- {
        prevZdd = res;
        id = dd.invperm[level];
        res = cuddZddGetNode(dd, id, prevZdd, prevZdd);
        if res == nil {
            Cudd_RecursiveDerefZdd(dd, prevZdd);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_RecursiveDerefZdd(dd, prevZdd);
    }

    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);
    return(res);

} /* end of zddPortFromBddStep */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_zddPortToBdd.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
static DdNode *
zddPortToBddStep(
  DdManager * dd /* manager */,
  DdNode * f /* ZDD to be converted */,
  int  depth /* recursion depth */)
{
    var one, zero, T, E, res, var *DdNode;
    unsigned int index;
    unsigned int level;

    statLine(dd);
    // inlining of macro DD_ONE
    one = ((dd).one);
    // inlining of macro DD_ZERO
    zero = ((dd).zero);
    // inlining of macro Cudd_Not
    if f == zero { return(((DdNode *)((ptrint)(one) ^ 01))); }

    if depth == dd.sizeZ { return(one); }

    index = dd.invpermZ[depth];
    // inlining of macro cuddIZ
    
    if (f.index)==CUDD_CONST_INDEX {
        level = int32((f.index));
    } else {
        level = (dd).permZ[(f.index)];
    }
    // inlining of macro Cudd_Not
    var = cuddUniqueInter(dd,index,one,((DdNode *)((ptrint)(one) ^ 01)));
    if var == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(var).ref);

    if level > uint32( depth) {
        E = zddPortToBddStep(dd,f,depth+1);
        if E == nil {
            Cudd_RecursiveDeref(dd,var);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(E).ref);
        // inlining of macro Cudd_Not
        res = cuddBddIteRecur(dd,var,((DdNode *)((ptrint)(one) ^ 01)),E);
        if res == nil {
            Cudd_RecursiveDeref(dd,var);
            Cudd_RecursiveDeref(dd,E);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_RecursiveDeref(dd,var);
        Cudd_RecursiveDeref(dd,E);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref);
        return(res);
    }

    res = cuddCacheLookup1(dd,Cudd_zddPortToBdd,f);
    if res != nil {
        Cudd_RecursiveDeref(dd,var);
        return(res);
    }

    // inlining of macro cuddT
    T = zddPortToBddStep(dd,((f).type.kids.T),depth+1);
    if T == nil {
        Cudd_RecursiveDeref(dd,var);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);
    // inlining of macro cuddE
    E = zddPortToBddStep(dd,((f).type.kids.E),depth+1);
    if E == nil {
        Cudd_RecursiveDeref(dd,var);
        Cudd_RecursiveDeref(dd,T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);

    res = cuddBddIteRecur(dd,var,T,E);
    if res == nil {
        Cudd_RecursiveDeref(dd,var);
        Cudd_RecursiveDeref(dd,T);
        Cudd_RecursiveDeref(dd,E);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref(dd,var);
    Cudd_RecursiveDeref(dd,T);
    Cudd_RecursiveDeref(dd,E);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);

    cuddCacheInsert1(dd,Cudd_zddPortToBdd,f,res);

    return(res);

} /* end of zddPortToBddStep */


//ABC_NAMESPACE_IMPL_END



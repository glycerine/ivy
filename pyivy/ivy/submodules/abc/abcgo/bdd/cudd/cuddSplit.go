package cudd

import (
	"fmt"
	"math"
)

/**CFile***********************************************************************

  FileName    [cuddSplit.c]

  PackageName [cudd]

  Synopsis    [Returns a subset of minterms from a boolean function.]

  Description [External functions included in this modoule:
                <ul>
                <li> Cudd_SplitSet()
                </ul>
        Internal functions included in this module:
                <ul>
                <li> cuddSplitSetRecur()
                </u>
        Static functions included in this module:
                <ul>
                <li> selectMintermsFromUniverse()
                <li> mintermsFromUniverse()
                <li> bddAnnotateMintermCount()
                </ul> ]

  SeeAlso     []

  Author      [Balakrishna Kumthekar]

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
/* Type declarations                                                         */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Structure declarations                                                    */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static DdNode * selectMintermsFromUniverse (DdManager *manager, int *varSeen, double n);

// static DdNode * mintermsFromUniverse (DdManager *manager, DdNode **vars, int numVars, double n, int index);

// static double bddAnnotateMintermCount (DdManager *manager, DdNode *node, double max, st__table *table);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Returns m minterms from a BDD.]

  Description [Returns <code>m</code> minterms from a BDD whose
  support has <code>n</code> variables at most.  The procedure tries
  to create as few extra nodes as possible. The function represented
  by <code>S</code> depends on at most <code>n</code> of the variables
  in <code>xVars</code>. Returns a BDD with <code>m</code> minterms
  of the on-set of S if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_SplitSet(manager *DdManager, S *DdNode, xVars **DdNode, n int32, m float64) *DdNode {
    var result *DdNode;
    var zero, one *DdNode;
    var max, num float64;
    var mtable *st__table;
    var varSeen *int32;
    var i, index, size int32;

    size = manager.size;
    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Trivial cases. */
    if m == 0.0 {
        return(zero);
    }
    if S == zero {
        return(nil);
    }

    max = math.Pow(float64(2.0), float64(float64(n)));
    if m > max {
        return(nil); }

    for {
        manager.reordered = 0;
        /* varSeen is used to mark the variables that are encountered
        ** while traversing the BDD S.
        */
        varSeen = ABC_ALLOC(int, size);
        if varSeen == nil {
            manager.errorCode = CUDD_MEMORY_OUT;
            return(nil);
        }
        for i = 0; i < size; i++ {
            varSeen[i] = -1;
        }
        for i = 0; i < n; i++ {
            index = (xVars[i]).index;
            varSeen[manager.invperm[index]] = 0;
        }

        if S == one {
            if m == max {
                ABC_FREE(varSeen);
                return(S);
            }
            result = selectMintermsFromUniverse(manager,varSeen,m);
            if result != nil {
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(result).ref); }
            ABC_FREE(varSeen);
        } else {
            mtable = st__init_table( st__ptrcmp, st__ptrhash);
            if mtable == nil {
                (void) fmt.Fprintf(manager.out, "Cudd_SplitSet: out-of-memory.\n");
                ABC_FREE(varSeen);
                manager.errorCode = CUDD_MEMORY_OUT;
                return(nil);
            }
            /* The nodes of BDD S are annotated by the number of minterms
            ** in their onset. The node and the number of minterms in its
            ** onset are stored in mtable.
            */
            num = bddAnnotateMintermCount(manager,S,max,mtable);
            if m == num {
                st__foreach(mtable,cuddStCountfree,NIL(char));
                st__free_table(mtable);
                ABC_FREE(varSeen);
                return(S);
            }
            
            result = cuddSplitSetRecur(manager,mtable,varSeen,S,m,max,0);
            if result != nil {
                // inlining of macro cuddRef
                cuddSatInc(Cudd_Regular(result).ref); }
            st__foreach(mtable,cuddStCountfree,nil);
            st__free_table(mtable);
            ABC_FREE(varSeen);
        }
    
if !(manager.reordered == 1) { break }
} 

    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(result).ref);
    return(result);

} /* end of Cudd_SplitSet */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Implements the recursive step of Cudd_SplitSet.]

  Description [Implements the recursive step of Cudd_SplitSet. The
  procedure recursively traverses the BDD and checks to see if any
  node satisfies the minterm requirements as specified by 'n'. At any
  node X, n is compared to the number of minterms in the onset of X's
  children. If either of the child nodes have exactly n minterms, then
  that node is returned; else, if n is greater than the onset of one
  of the child nodes, that node is retained and the difference in the
  number of minterms is extracted from the other child. In case n
  minterms can be extracted from constant 1, the algorithm returns the
  result with at most log(n) nodes.]

  SideEffects [The array 'varSeen' is updated at every recursive call
  to set the variables traversed by the procedure.]

  SeeAlso     []

******************************************************************************/
func cuddSplitSetRecur(manager *DdManager, mtable *st__table, varSeen *int32, p *DdNode, n float64, max float64, index int32) *DdNode {
    var one, zero, N, Nv *DdNode;
    var Nnv, q, r, v *DdNode;
    var result *DdNode;
    var dummy *float64;
    var numT float64;
    var numE float64;
    var variable, positive int32;
  
    statLine(manager);
    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* If p is constant, extract n minterms from constant 1.  The procedure by
    ** construction guarantees that minterms will not be extracted from
    ** constant 0.
    */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if (((DdNode *)((ptruint)(p) & ~01))).index == CUDD_CONST_INDEX {
        q = selectMintermsFromUniverse(manager,varSeen,n);
        return(q);
    }

    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(p) & ~01));

    /* Set variable as seen. */
    variable = N.index;
    varSeen[manager.invperm[variable]] = -1;

    // inlining of macro cuddT
    Nv = ((N).type.kids.T);
    // inlining of macro cuddE
    Nnv = ((N).type.kids.E);
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (p) & 01)) != 0 {
        // inlining of macro Cudd_Not
        Nv = ((DdNode *)((ptrint)(Nv) ^ 01));
        // inlining of macro Cudd_Not
        Nnv = ((DdNode *)((ptrint)(Nnv) ^ 01));
    }

    /* If both the children of 'p' are constants, extract n minterms from a
    ** constant node.
    */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if ((((DdNode *)((ptruint)(Nv) & ~01))).index == CUDD_CONST_INDEX) && ((((DdNode *)((ptruint)(Nnv) & ~01))).index == CUDD_CONST_INDEX) {
        q = selectMintermsFromUniverse(manager,varSeen,n);
        if q == nil {
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(q).ref);
        r = cuddBddAndRecur(manager,p,q);
        if r == nil {
            Cudd_RecursiveDeref(manager,q);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(r).ref);
        Cudd_RecursiveDeref(manager,q);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(r).ref);
        return(r);
    }
  
    /* Lookup the # of minterms in the onset of the node from the table. */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Nv) & ~01))).index == CUDD_CONST_INDEX) {
        if ! st__lookup(mtable, (char *)Nv, (char **)&dummy) { return(nil); }
        numT = *dummy/(2*(1<<index));
    } else if Nv == one {
        numT = max/(2*(1<<index));
    } else {
        numT = 0;
    }
  
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if !((((DdNode *)((ptruint)(Nnv) & ~01))).index == CUDD_CONST_INDEX) {
        if ! st__lookup(mtable, (char *)Nnv, (char **)&dummy) { return(nil); }
        numE = *dummy/(2*(1<<index));
    } else if Nnv == one {
        numE = max/(2*(1<<index));
    } else {
        numE = 0;
    }

    v = cuddUniqueInter(manager,variable,one,zero);
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(v).ref);

    /* If perfect match. */
    if numT == n {
        q = cuddBddAndRecur(manager,v,Nv);
        if q == nil {
            Cudd_RecursiveDeref(manager,v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(q).ref);
        Cudd_RecursiveDeref(manager,v);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(q).ref);
        return(q);
    }
    if numE == n {
        // inlining of macro Cudd_Not
        q = cuddBddAndRecur(manager,((DdNode *)((ptrint)(v) ^ 01)),Nnv);
        if q == nil {
            Cudd_RecursiveDeref(manager,v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(q).ref);
        Cudd_RecursiveDeref(manager,v);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(q).ref);
        return(q);
    }
    /* If n is greater than numT, extract the difference from the ELSE child
    ** and retain the function represented by the THEN branch.
    */
    if numT < n {
        q = cuddSplitSetRecur(manager,mtable,varSeen,
                              Nnv,(n-numT),max,index+1);
        if q == nil {
            Cudd_RecursiveDeref(manager,v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(q).ref);
        r = cuddBddIteRecur(manager,v,Nv,q);
        if r == nil {
            Cudd_RecursiveDeref(manager,q);
            Cudd_RecursiveDeref(manager,v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(r).ref);
        Cudd_RecursiveDeref(manager,q);
        Cudd_RecursiveDeref(manager,v);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(r).ref);
        return(r);
    }
    /* If n is greater than numE, extract the difference from the THEN child
    ** and retain the function represented by the ELSE branch.
    */
    if numE < n {
        q = cuddSplitSetRecur(manager,mtable,varSeen,
                              Nv, (n-numE),max,index+1);
        if q == nil {
            Cudd_RecursiveDeref(manager,v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(q).ref);
        r = cuddBddIteRecur(manager,v,q,Nnv);
        if r == nil {
            Cudd_RecursiveDeref(manager,q);
            Cudd_RecursiveDeref(manager,v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(r).ref);
        Cudd_RecursiveDeref(manager,q);
        Cudd_RecursiveDeref(manager,v);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(r).ref);    
        return(r);
    }

    /* None of the above cases; (n < numT and n < numE) and either of
    ** the Nv, Nnv or both are not constants. If possible extract the
    ** required minterms the constant branch.
    */
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    if ((((DdNode *)((ptruint)(Nv) & ~01))).index == CUDD_CONST_INDEX) && !((((DdNode *)((ptruint)(Nnv) & ~01))).index == CUDD_CONST_INDEX) {
        q = selectMintermsFromUniverse(manager,varSeen,n);
        if q == nil {
            Cudd_RecursiveDeref(manager,v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(q).ref);
        result = cuddBddAndRecur(manager,v,q);
        if result == nil {
            Cudd_RecursiveDeref(manager,q);
            Cudd_RecursiveDeref(manager,v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(result).ref);
        Cudd_RecursiveDeref(manager,q);
        Cudd_RecursiveDeref(manager,v);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(result).ref);
        return(result);
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    } else if !((((DdNode *)((ptruint)(Nv) & ~01))).index == CUDD_CONST_INDEX) && ((((DdNode *)((ptruint)(Nnv) & ~01))).index == CUDD_CONST_INDEX) {
        q = selectMintermsFromUniverse(manager,varSeen,n);
        if q == nil {
            Cudd_RecursiveDeref(manager,v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(q).ref);
        // inlining of macro Cudd_Not
        result = cuddBddAndRecur(manager,((DdNode *)((ptrint)(v) ^ 01)),q);
        if result == nil {
            Cudd_RecursiveDeref(manager,q);
            Cudd_RecursiveDeref(manager,v);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(result).ref);
        Cudd_RecursiveDeref(manager,q);
        Cudd_RecursiveDeref(manager,v);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(result).ref);
        return(result);
    }

    /* Both Nv and Nnv are not constants. So choose the one which
    ** has fewer minterms in its onset.
    */
    positive = 0;
    if numT < numE {
        q = cuddSplitSetRecur(manager,mtable,varSeen,
                              Nv,n,max,index+1);
        positive = 1;
    } else {
        q = cuddSplitSetRecur(manager,mtable,varSeen,
                              Nnv,n,max,index+1);
    }

    if q == nil {
        Cudd_RecursiveDeref(manager,v);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(q).ref);

    if positive != 0 {
        result = cuddBddAndRecur(manager,v,q);
    } else {
        // inlining of macro Cudd_Not
        result = cuddBddAndRecur(manager,((DdNode *)((ptrint)(v) ^ 01)),q);
    }
    if result == nil {
        Cudd_RecursiveDeref(manager,q);
        Cudd_RecursiveDeref(manager,v);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(result).ref);
    Cudd_RecursiveDeref(manager,q);
    Cudd_RecursiveDeref(manager,v);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(result).ref);

    return(result);

} /* end of cuddSplitSetRecur */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis [This function prepares an array of variables which have not been
  encountered so far when traversing the procedure cuddSplitSetRecur.]

  Description [This function prepares an array of variables which have not been
  encountered so far when traversing the procedure cuddSplitSetRecur. This
  array is then used to extract the required number of minterms from a constant
  1. The algorithm guarantees that the size of BDD will be utmost \log(n).]

  SideEffects [None]

******************************************************************************/
func selectMintermsFromUniverse(manager *DdManager, varSeen *int32, n float64) *DdNode {
    var numVars int32;
    var i, size, j int32;
     var one, zero, result *DdNode;
    var vars **DdNode;

    numVars = 0;
    size = manager.size;
    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    /* Count the number of variables not encountered so far in procedure
    ** cuddSplitSetRecur.
    */
    for i = size-1; i >= 0; i-- {
        if varSeen[i] == 0 {
            numVars++; }
    }
    vars = ABC_ALLOC(DdNode *, numVars);
    if !vars {
        manager.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }

    j = 0;
    for i = size-1; i >= 0; i-- {
        if varSeen[i] == 0 {
            vars[j] = cuddUniqueInter(manager,manager.perm[i],one,zero);
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(vars[j]).ref);
            j++;
        }
    }

    /* Compute a function which has n minterms and depends on at most
    ** numVars variables.
    */
    result = mintermsFromUniverse(manager,vars,numVars,n, 0);
    if result != nil { 
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(result).ref); }

    for i = 0; i < numVars; i++ {
        Cudd_RecursiveDeref(manager,vars[i]);
    }
    ABC_FREE(vars);

    return(result);

} /* end of selectMintermsFromUniverse */


/**Function********************************************************************

  Synopsis    [Recursive procedure to extract n mintems from constant 1.]

  Description [Recursive procedure to extract n mintems from constant 1.]

  SideEffects [None]

******************************************************************************/
func mintermsFromUniverse(manager *DdManager, vars **DdNode, numVars int32, n float64, index int32) *DdNode {
    var one, zero *DdNode;
    var q, result *DdNode;
    var max, max2 float64;
    
    statLine(manager);
    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro Cudd_Not
    zero = ((DdNode *)((ptrint)(one) ^ 01));

    max = math.Pow(float64(2.0), float64(float64(numVars)));
    max2 = max / 2.0;

    if n == max {
        return(one); }
    if n == 0.0 {
        return(zero); }
    /* if n == 2^(numVars-1), return a single variable */
    if n == max2 {
        return vars[index];
    } else if n > max2 {
        /* When n > 2^(numVars-1), a single variable vars[index]
        ** contains 2^(numVars-1) minterms. The rest are extracted
        ** from a constant with 1 less variable.
        */
        q = mintermsFromUniverse(manager,vars,numVars-1,(n-max2),index+1);
        if q == nil {
            return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(q).ref);
        result = cuddBddIteRecur(manager,vars[index],one,q);
    } else {
        /* When n < 2^(numVars-1), a literal of variable vars[index]
        ** is selected. The required n minterms are extracted from a
        ** constant with 1 less variable.
        */
        q = mintermsFromUniverse(manager,vars,numVars-1,n,index+1);
        if q == nil {
            return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(q).ref);
        result = cuddBddAndRecur(manager,vars[index],q);
    }
    
    if result == nil {
        Cudd_RecursiveDeref(manager,q);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(result).ref);
    Cudd_RecursiveDeref(manager,q);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(result).ref);
    return(result);

} /* end of mintermsFromUniverse */


/**Function********************************************************************

  Synopsis    [Annotates every node in the BDD node with its minterm count.]

  Description [Annotates every node in the BDD node with its minterm count.
  In this function, every node and the minterm count represented by it are
  stored in a hash table.]

  SideEffects [Fills up 'table' with the pair <node,minterm_count>.]

******************************************************************************/
func bddAnnotateMintermCount(manager *DdManager, node *DdNode, max float64, table *st__table) float64 {

    var N, Nv, Nnv *DdNode;
    register double min_v,min_nv;
    register double min_N;
    var pmin *float64;
    var dummy *float64;

    statLine(manager);
    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));
    // inlining of macro cuddIsConstant
    if (N).index == CUDD_CONST_INDEX {
        // inlining of macro DD_ONE
        if node == ((manager).one) {
            return(max);
        } else {
            return(0.0);
        }
    }

    if st__lookup(table, (char *)node, (char **)&dummy) {
        return(*dummy);
    }   
  
    // inlining of macro cuddT
    Nv = ((N).type.kids.T);
    // inlining of macro cuddE
    Nnv = ((N).type.kids.E);
    if N != node {
        // inlining of macro Cudd_Not
        Nv = ((DdNode *)((ptrint)(Nv) ^ 01));
        // inlining of macro Cudd_Not
        Nnv = ((DdNode *)((ptrint)(Nnv) ^ 01));
    }

    /* Recur on the two branches. */
    min_v  = bddAnnotateMintermCount(manager,Nv,max,table) / 2.0;
    if min_v == float64(CUDD_OUT_OF_MEM) {
        return (float64(CUDD_OUT_OF_MEM)); }
    min_nv = bddAnnotateMintermCount(manager,Nnv,max,table) / 2.0;
    if min_nv == float64(CUDD_OUT_OF_MEM) {
        return (float64(CUDD_OUT_OF_MEM)); }
    min_N  = min_v + min_nv;

    pmin = new(float64);
    if pmin == nil {
        manager.errorCode = CUDD_MEMORY_OUT;
        return(float64(CUDD_OUT_OF_MEM));
    }
    *pmin = min_N;

    if st__insert(table,(char *)node, (char *)pmin) == st__OUT_OF_MEM {
        ABC_FREE(pmin);
        return(float64(CUDD_OUT_OF_MEM));
    }
    
    return(min_N);

} /* end of bddAnnotateMintermCount */


//ABC_NAMESPACE_IMPL_END


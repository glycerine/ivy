package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddSign.c]

  PackageName [cudd]

  Synopsis    [Computation of signatures.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_CofMinterm();
                    </ul>
                Static procedures included in this module:
                    <ul>
                    <li> ddCofMintermAux()
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
static char rcsid[] DD_UNUSED = "$Id: cuddSign.c,v 1.22 2009/02/20 02:14:58 fabio Exp $";
#endif

var size int32;

#ifdef DD_STATS
static int num_calls;   /* should equal 2n-1 (n is the # of nodes) */
static int table_mem;
#endif


/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static double * ddCofMintermAux (DdManager *dd, DdNode *node, st__table *table);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis [Computes the fraction of minterms in the on-set of all the
  positive cofactors of a BDD or ADD.]

  Description [Computes the fraction of minterms in the on-set of all
  the positive cofactors of DD. Returns the pointer to an array of
  doubles if successful; NULL otherwise. The array has as many
  positions as there are BDD variables in the manager plus one. The
  last position of the array contains the fraction of the minterms in
  the ON-set of the function represented by the BDD or ADD. The other
  positions of the array hold the variable signatures.]

  SideEffects [None]

******************************************************************************/
func Cudd_CofMinterm(dd *DdManager, node *DdNode) *float64 {
    var table *st__table;
    var values *float64;
    var result *float64  = nil;
    var i, firstLevel int32;

#ifdef DD_STATS
    long startTime;
    startTime = util_cpu_time();
    num_calls = 0;
    table_mem = sizeof( st__table);
#endif

    table = st__init_table( st__ptrcmp, st__ptrhash);
    if table == nil {
        (void) fmt.Fprintf(dd.err, "out-of-memory, couldn't measure DD cofactors.\n");
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }
    size = dd.size;
    values = ddCofMintermAux(dd, node, table);
    if values != nil {
        result = ABC_ALLOC(double,size + 1);
        if result != nil {
#ifdef DD_STATS
            table_mem += (size + 1) * sizeof(double);
#endif
            // inlining of macro Cudd_IsConstant
            // inlining of macro Cudd_Regular
            if (((DdNode *)((ptruint)(node) & ~01))).index == CUDD_CONST_INDEX {
                firstLevel = 1;
            } else {
                // inlining of macro cuddI
                
                // inlining of macro Cudd_Regular
                if (((DdNode *)((ptruint)(node) & ~01)).index)==CUDD_CONST_INDEX {
                    // inlining of macro Cudd_Regular
                    firstLevel = int32((((DdNode *)((ptruint)(node) & ~01)).index));
                } else {
                    // inlining of macro Cudd_Regular
                    firstLevel = (dd).perm[(((DdNode *)((ptruint)(node) & ~01)).index)];
                } }
            for i = 0; i < size; i++ {
                // inlining of macro cuddI
                // inlining of macro Cudd_Regular
                if i >= (((((DdNode *)((ptruint)(node) & ~01)).index)==CUDD_CONST_INDEX)?int32((((DdNode *)((ptruint)(node) & ~01)).index)):(dd).perm[(((DdNode *)((ptruint)(node) & ~01)).index)]) {
                    result[dd.invperm[i]] = values[i - firstLevel];
                } else {
                    result[dd.invperm[i]] = values[size - firstLevel];
                }
            }
            result[size] = values[size - firstLevel];
        } else {
            dd.errorCode = CUDD_MEMORY_OUT;
        }
    }

#ifdef DD_STATS
    table_mem += table->num_bins * sizeof( st__table_entry *);
#endif
    // inlining of macro Cudd_Regular
    if ((DdNode *)((ptruint)(node) & ~01)).ref == 1 { ABC_FREE(values); }
    st__foreach(table, cuddStCountfree, nil);
    st__free_table(table);
#ifdef DD_STATS
    (void) fprintf(dd->out,"Number of calls: %d\tTable memory: %d bytes\n",
                  num_calls, table_mem);
    (void) fprintf(dd->out,"Time to compute measures: %s\n",
                  util_print_time(util_cpu_time() - startTime));
#endif
    if result == nil {
        (void) fmt.Fprintf(dd.out, "out-of-memory, couldn't measure DD cofactors.\n");
        dd.errorCode = CUDD_MEMORY_OUT;
    }
    return(result);

} /* end of Cudd_CofMinterm */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Recursive Step for Cudd_CofMinterm function.]

  Description [Traverses the DD node and computes the fraction of
  minterms in the on-set of all positive cofactors simultaneously.
  It allocates an array with two more entries than there are
  variables below the one labeling the node.  One extra entry (the
  first in the array) is for the variable labeling the node. The other
  entry (the last one in the array) holds the fraction of minterms of
  the function rooted at node.  Each other entry holds the value for
  one cofactor. The array is put in a symbol table, to avoid repeated
  computation, and its address is returned by the procedure, for use
  by the caller.  Returns a pointer to the array of cofactor measures.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func ddCofMintermAux(dd *DdManager, node *DdNode, table *st__table) *float64 {
    var N *DdNode;             /* regular version of node */
    var Nv, Nnv *DdNode;
    var values *float64;
    var valuesT, valuesE *float64;
    var i int32;
    var localSize, localSizeT, localSizeE int32;
    var vT, vE float64;

    statLine(dd);
#ifdef DD_STATS
    num_calls++;
#endif

    if st__lookup(table, (char *)node, (char **)&values) {
        return(values);
    }

    // inlining of macro Cudd_Regular
    N = ((DdNode *)((ptruint)(node) & ~01));
    // inlining of macro cuddIsConstant
    if (N).index == CUDD_CONST_INDEX {
        localSize = 1;
    } else {
        // inlining of macro cuddI
        localSize = size - (((N.index)==CUDD_CONST_INDEX)?int32((N.index)):(dd).perm[(N.index)]) + 1;
    }
    values = ABC_ALLOC(double, localSize);
    if values == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(nil);
    }

    // inlining of macro cuddIsConstant
    if (N).index == CUDD_CONST_INDEX {
        // inlining of macro DD_ZERO
        // inlining of macro Cudd_Not
        // inlining of macro DD_ONE
        if node == ((dd).zero) || node == ((DdNode *)((ptrint)(((dd).one)) ^ 01)) {
            values[0] = 0.0;
        } else {
            values[0] = 1.0;
        }
    } else {
        // inlining of macro Cudd_NotCond
        // inlining of macro cuddT
        Nv = ((DdNode *)((ptrint)(((N).type.kids.T)) ^ (N!=node)));
        // inlining of macro Cudd_NotCond
        // inlining of macro cuddE
        Nnv = ((DdNode *)((ptrint)(((N).type.kids.E)) ^ (N!=node)));

        valuesT = ddCofMintermAux(dd, Nv, table);
        if valuesT == nil { return(nil); }
        valuesE = ddCofMintermAux(dd, Nnv, table);
        if valuesE == nil { return(nil); }

        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        if (((DdNode *)((ptruint)(Nv) & ~01))).index == CUDD_CONST_INDEX {
            localSizeT = 1;
        } else {
            // inlining of macro cuddI
            // inlining of macro Cudd_Regular
            localSizeT = size - (((((DdNode *)((ptruint)(Nv) & ~01)).index)==CUDD_CONST_INDEX)?int32((((DdNode *)((ptruint)(Nv) & ~01)).index)):(dd).perm[(((DdNode *)((ptruint)(Nv) & ~01)).index)]) + 1;
        }
        // inlining of macro Cudd_IsConstant
        // inlining of macro Cudd_Regular
        if (((DdNode *)((ptruint)(Nnv) & ~01))).index == CUDD_CONST_INDEX {
            localSizeE = 1;
        } else {
            // inlining of macro cuddI
            // inlining of macro Cudd_Regular
            localSizeE = size - (((((DdNode *)((ptruint)(Nnv) & ~01)).index)==CUDD_CONST_INDEX)?int32((((DdNode *)((ptruint)(Nnv) & ~01)).index)):(dd).perm[(((DdNode *)((ptruint)(Nnv) & ~01)).index)]) + 1;
        }
        values[0] = valuesT[localSizeT - 1];
        for i = 1; i < localSize; i++ {
            // inlining of macro cuddI
            // inlining of macro Cudd_Regular
            if i >= (((((DdNode *)((ptruint)(Nv) & ~01)).index)==CUDD_CONST_INDEX)?int32((((DdNode *)((ptruint)(Nv) & ~01)).index)):(dd).perm[(((DdNode *)((ptruint)(Nv) & ~01)).index)]) - (((N.index)==CUDD_CONST_INDEX)?int32((N.index)):(dd).perm[(N.index)]) {
                // inlining of macro cuddI
                // inlining of macro Cudd_Regular
                vT = valuesT[i - (((((DdNode *)((ptruint)(Nv) & ~01)).index)==CUDD_CONST_INDEX)?int32((((DdNode *)((ptruint)(Nv) & ~01)).index)):(dd).perm[(((DdNode *)((ptruint)(Nv) & ~01)).index)]) +
                            // inlining of macro cuddI
                            (((N.index)==CUDD_CONST_INDEX)?int32((N.index)):(dd).perm[(N.index)])];
            } else {
                vT = valuesT[localSizeT - 1];
            }
            // inlining of macro cuddI
            // inlining of macro Cudd_Regular
            if i >= (((((DdNode *)((ptruint)(Nnv) & ~01)).index)==CUDD_CONST_INDEX)?int32((((DdNode *)((ptruint)(Nnv) & ~01)).index)):(dd).perm[(((DdNode *)((ptruint)(Nnv) & ~01)).index)]) - (((N.index)==CUDD_CONST_INDEX)?int32((N.index)):(dd).perm[(N.index)]) {
                // inlining of macro cuddI
                // inlining of macro Cudd_Regular
                vE = valuesE[i - (((((DdNode *)((ptruint)(Nnv) & ~01)).index)==CUDD_CONST_INDEX)?int32((((DdNode *)((ptruint)(Nnv) & ~01)).index)):(dd).perm[(((DdNode *)((ptruint)(Nnv) & ~01)).index)]) +
                            // inlining of macro cuddI
                            (((N.index)==CUDD_CONST_INDEX)?int32((N.index)):(dd).perm[(N.index)])];
            } else {
                vE = valuesE[localSizeE - 1];
            }
            values[i] = (vT + vE) / 2.0;
        }
        // inlining of macro Cudd_Regular
        if ((DdNode *)((ptruint)(Nv) & ~01)).ref == 1 { ABC_FREE(valuesT); }
        // inlining of macro Cudd_Regular
        if ((DdNode *)((ptruint)(Nnv) & ~01)).ref == 1 { ABC_FREE(valuesE); }
    }

    if N.ref > 1 {
        if st__add_direct(table, (char *) node, (char *) values) == st__OUT_OF_MEM {
            ABC_FREE(values);
            return(nil);
        }
#ifdef DD_STATS
        table_mem += localSize * sizeof(double) + sizeof( st__table_entry);
#endif
    }
    return(values);

} /* end of ddCofMintermAux */


//ABC_NAMESPACE_IMPL_END


package cudd

/**CFile***********************************************************************

  FileName    [cuddZddCount.c]

  PackageName [cudd]

  Synopsis    [Procedures to count the number of minterms of a ZDD.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_zddCount();
                    <li> Cudd_zddCountDouble();
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    </ul>
               Static procedures included in this module:
                    <ul>
                    <li> cuddZddCountStep();
                    <li> cuddZddCountDoubleStep();
                    <li> st__zdd_count_dbl_free()
                    <li> st__zdd_countfree()
                    </ul>
              ]

  SeeAlso     []

  Author      [Hyong-Kyoon Shin, In-Ho Moon]

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
static char rcsid[] DD_UNUSED = "$Id: cuddZddCount.c,v 1.14 2004/08/13 18:04:53 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int cuddZddCountStep (DdNode *P, st__table *table, DdNode *base, DdNode *empty);

// static double cuddZddCountDoubleStep (DdNode *P, st__table *table, DdNode *base, DdNode *empty);

// static      st__retval st__zdd_countfree (char *key, char *value, char *arg);

// static      st__retval st__zdd_count_dbl_free (char *key, char *value, char *arg);


/**AutomaticEnd***************************************************************/

/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Counts the number of minterms in a ZDD.]

  Description [Returns an integer representing the number of minterms
  in a ZDD.]

  SideEffects [None]

  SeeAlso     [Cudd_zddCountDouble]

******************************************************************************/
func Cudd_zddCount(zdd *DdManager, P *DdNode) int32 {
    var table *st__table;
    var res int32;
    var base, empty *DdNode;

    // inlining of macro DD_ONE
    base  = ((zdd).one);
    // inlining of macro DD_ZERO
    empty = ((zdd).zero);
    table = st__init_table( st__ptrcmp, st__ptrhash);
    if table == nil { return(CUDD_OUT_OF_MEM); }
    res = cuddZddCountStep(P, table, base, empty);
    if res == CUDD_OUT_OF_MEM {
        zdd.errorCode = CUDD_MEMORY_OUT;
    }
    st__foreach(table, st__zdd_countfree, NIL(char));
    st__free_table(table);

    return(res);

} /* end of Cudd_zddCount */


/**Function********************************************************************

  Synopsis [Counts the number of minterms of a ZDD.]

  Description [Counts the number of minterms of a ZDD. The result is
  returned as a double. If the procedure runs out of memory, it
  returns (double) CUDD_OUT_OF_MEM. This procedure is used in
  Cudd_zddCountMinterm.]

  SideEffects [None]

  SeeAlso     [Cudd_zddCountMinterm Cudd_zddCount]

******************************************************************************/
func Cudd_zddCountDouble(zdd *DdManager, P *DdNode) float64 {
    var table *st__table;
    var res float64;
    var base, empty *DdNode;

    // inlining of macro DD_ONE
    base  = ((zdd).one);
    // inlining of macro DD_ZERO
    empty = ((zdd).zero);
    table = st__init_table( st__ptrcmp, st__ptrhash);
    if table == nil { return(float64(CUDD_OUT_OF_MEM)); }
    res = cuddZddCountDoubleStep(P, table, base, empty);
    if res == float64(CUDD_OUT_OF_MEM) {
        zdd.errorCode = CUDD_MEMORY_OUT;
    }
    st__foreach(table, st__zdd_count_dbl_free, NIL(char));
    st__free_table(table);

    return(res);

} /* end of Cudd_zddCountDouble */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_zddCount.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddCountStep(P *DdNode, table *st__table, base *DdNode, empty *DdNode) int32 {
    var res int32;
    var dummy *int32;

    if P == empty {
        return(0); }
    if P == base {
        return(1); }

    /* Check cache. */
    if st__lookup(table, (char *)P, (char **)&dummy) {
        res = *dummy;
        return(res);
    }

    // inlining of macro cuddE
    res = cuddZddCountStep(((P).type.kids.E), table, base, empty) +
        // inlining of macro cuddT
        cuddZddCountStep(((P).type.kids.T), table, base, empty);

    dummy = new(int32);
    if dummy == nil {
        return(CUDD_OUT_OF_MEM);
    }
    *dummy = res;
    if st__insert(table, (char *)P, (char *)dummy) == st__OUT_OF_MEM {
        ABC_FREE(dummy);
        return(CUDD_OUT_OF_MEM);
    }

    return(res);

} /* end of cuddZddCountStep */


/**Function********************************************************************

  Synopsis [Performs the recursive step of Cudd_zddCountDouble.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddZddCountDoubleStep(P *DdNode, table *st__table, base *DdNode, empty *DdNode) float64 {
    var res float64;
    var dummy *float64;

    if P == empty {
        return(float64(0.0)); }
    if P == base {
        return(float64(1.0)); }

    /* Check cache */
    if st__lookup(table, (char *)P, (char **)&dummy) {
        res = *dummy;
        return(res);
    }

    // inlining of macro cuddE
    res = cuddZddCountDoubleStep(((P).type.kids.E), table, base, empty) +
        // inlining of macro cuddT
        cuddZddCountDoubleStep(((P).type.kids.T), table, base, empty);

    dummy = new(float64);
    if dummy == nil {
        return(float64(CUDD_OUT_OF_MEM));
    }
    *dummy = res;
    if st__insert(table, (char *)P, (char *)dummy) == st__OUT_OF_MEM {
        ABC_FREE(dummy);
        return(float64(CUDD_OUT_OF_MEM));
    }

    return(res);

} /* end of cuddZddCountDoubleStep */


/**Function********************************************************************

  Synopsis [Frees the memory associated with the computed table of
  Cudd_zddCount.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func st__zdd_countfree(key string, value string, arg string) st__retval {
    var d *int32;

    d = (int *)value;
    ABC_FREE(d);
    return( st__CONTINUE);

} /* end of st__zdd_countfree */


/**Function********************************************************************

  Synopsis [Frees the memory associated with the computed table of
  Cudd_zddCountDouble.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func st__zdd_count_dbl_free(key string, value string, arg string) st__retval {
    var d *float64;

    d = (double *)value;
    ABC_FREE(d);
    return( st__CONTINUE);

} /* end of st__zdd_count_dbl_free */


//ABC_NAMESPACE_IMPL_END


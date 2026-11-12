package cudd

/**CFile***********************************************************************

  FileName    [cuddBddCorr.c]

  PackageName [cudd]

  Synopsis    [Correlation between BDDs.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_bddCorrelation()
                <li> Cudd_bddCorrelationWeights()
                </ul>
            Static procedures included in this module:
                <ul>
                <li> bddCorrelationAux()
                <li> bddCorrelationWeightsAux()
                <li> CorrelCompare()
                <li> CorrelHash()
                <li> CorrelCleanUp()
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

typedef struct hashEntry {
    DdNode *f;
    DdNode *g;
}
 HashEntry;


/*---------------------------------------------------------------------------*/
/* Variable declarations                                                     */
/*---------------------------------------------------------------------------*/

#ifndef lint
static char rcsid[] DD_UNUSED = "$Id: cuddBddCorr.c,v 1.14 2004/08/13 18:04:46 fabio Exp $";
#endif

#ifdef CORREL_STATS
static  int     num_calls;
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static double bddCorrelationAux (DdManager *dd, DdNode *f, DdNode *g, st__table *table);

// static double bddCorrelationWeightsAux (DdManager *dd, DdNode *f, DdNode *g, double *prob, st__table *table);

// static int CorrelCompare (const char *key1, const char *key2);

// static int CorrelHash (const char *key, int modulus);

// static      st__retval CorrelCleanUp (char *key, char *value, char *arg);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Computes the correlation of f and g.]

  Description [Computes the correlation of f and g. If f == g, their
  correlation is 1. If f == g', their correlation is 0.  Returns the
  fraction of minterms in the ON-set of the EXNOR of f and g.  If it
  runs out of memory, returns (double)CUDD_OUT_OF_MEM.]

  SideEffects [None]

  SeeAlso     [Cudd_bddCorrelationWeights]

******************************************************************************/
func Cudd_bddCorrelation(manager *DdManager, f *DdNode, g *DdNode) float64 {

    var table *st__table;
    var correlation float64;

#ifdef CORREL_STATS
    num_calls = 0;
#endif

    table = st__init_table(CorrelCompare,CorrelHash);
    if table == nil { return(float64(CUDD_OUT_OF_MEM)); }
    correlation = bddCorrelationAux(manager,f,g,table);
    st__foreach(table, CorrelCleanUp, NIL(char));
    st__free_table(table);
    return(correlation);

} /* end of Cudd_bddCorrelation */


/**Function********************************************************************

  Synopsis    [Computes the correlation of f and g for given input
  probabilities.]

  Description [Computes the correlation of f and g for given input
  probabilities. On input, prob\[i\] is supposed to contain the
  probability of the i-th input variable to be 1.
  If f == g, their correlation is 1. If f == g', their
  correlation is 0.  Returns the probability that f and g have the same
  value. If it runs out of memory, returns (double)CUDD_OUT_OF_MEM. The
  correlation of f and the constant one gives the probability of f.]

  SideEffects [None]

  SeeAlso     [Cudd_bddCorrelation]

******************************************************************************/
func Cudd_bddCorrelationWeights(manager *DdManager, f *DdNode, g *DdNode, prob *float64) float64 {

    var table *st__table;
    var correlation float64;

#ifdef CORREL_STATS
    num_calls = 0;
#endif

    table = st__init_table(CorrelCompare,CorrelHash);
    if table == nil { return(float64(CUDD_OUT_OF_MEM)); }
    correlation = bddCorrelationWeightsAux(manager,f,g,prob,table);
    st__foreach(table, CorrelCleanUp, NIL(char));
    st__free_table(table);
    return(correlation);

} /* end of Cudd_bddCorrelationWeights */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddCorrelation.]

  Description [Performs the recursive step of Cudd_bddCorrelation.
  Returns the fraction of minterms in the ON-set of the EXNOR of f and
  g.]

  SideEffects [None]

  SeeAlso     [bddCorrelationWeightsAux]

******************************************************************************/
func bddCorrelationAux(dd *DdManager, f *DdNode, g *DdNode, table *st__table) float64 {
    var Fv, Fnv, G, Gv, Gnv *DdNode;
    var min float64;
    var pmin *float64;
    var min1 float64;
    var min2 float64;
    var dummy *float64;
    var entry *HashEntry;
    unsigned int topF, topG;

    statLine(dd);
#ifdef CORREL_STATS
    num_calls++;
#endif

    /* Terminal cases: only work for BDDs. */
    if f == g { return(1.0); }
    // inlining of macro Cudd_Not
    if f == ((DdNode *)((ptrint)(g) ^ 01)) { return(0.0); }

    /* Standardize call using the following properties:
    **     (f EXNOR g)   = (g EXNOR f)
    **     (f' EXNOR g') = (f EXNOR g).
    */
    // inlining of macro cuddF2L
    // inlining of macro Cudd_Regular
    // inlining of macro Cudd_IsComplement
    if ((((DdNode *)((ptruint)(f) & ~01)).Id << 1) | (int32( ((ptrint) (f) & 01)))) > ((((DdNode *)((ptruint)(g) & ~01)).Id << 1) | (int32( ((ptrint) (g) & 01)))) {
        var tmp *DdNode  = f;
        f = g; g = tmp;
    }
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (f) & 01)) != 0 {
        // inlining of macro Cudd_Not
        f = ((DdNode *)((ptrint)(f) ^ 01));
        // inlining of macro Cudd_Not
        g = ((DdNode *)((ptrint)(g) ^ 01));
    }
    /* From now on, f is regular. */
    
    entry = new(HashEntry);
    if entry == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(CUDD_OUT_OF_MEM);
    }
    entry.f = f; entry.g = g;

    /* We do not use the fact that
    ** correlation(f,g') = 1 - correlation(f,g)
    ** to minimize the risk of cancellation.
    */
    if st__lookup(table, (char *)entry, (char **)&dummy) {
        min = *dummy;
        ABC_FREE(entry);
        return(min);
    }

    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    // inlining of macro cuddI
    
    if (f.index)==CUDD_CONST_INDEX {
        topF = int32((f.index));
    } else {
        topF = (dd).perm[(f.index)];
    // inlining of macro cuddI
    } 
    if (G.index)==CUDD_CONST_INDEX {
        topG = int32((G.index));
    } else {
        topG = (dd).perm[(G.index)];
    }
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topF <= topG { Fv = ((f).type.kids.T); Fnv = ((f).type.kids.E); } else { Fv = Fnv = f; }
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topG <= topF { Gv = ((G).type.kids.T); Gnv = ((G).type.kids.E); } else { Gv = Gnv = G; }

    if g != G {
        // inlining of macro Cudd_Not
        Gv = ((DdNode *)((ptrint)(Gv) ^ 01));
        // inlining of macro Cudd_Not
        Gnv = ((DdNode *)((ptrint)(Gnv) ^ 01));
    }

    min1 = bddCorrelationAux(dd, Fv, Gv, table) / 2.0;
    if min1 == float64(CUDD_OUT_OF_MEM) {
        ABC_FREE(entry);
        return(CUDD_OUT_OF_MEM);
    }
    min2 = bddCorrelationAux(dd, Fnv, Gnv, table) / 2.0; 
    if min2 == float64(CUDD_OUT_OF_MEM) {
        ABC_FREE(entry);
        return(CUDD_OUT_OF_MEM);
    }
    min = (min1+min2);
    
    pmin = new(float64);
    if pmin == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(float64(CUDD_OUT_OF_MEM));
    }
    *pmin = min;

    if st__insert(table,(char *)entry, (char *)pmin) == st__OUT_OF_MEM {
        ABC_FREE(entry);
        ABC_FREE(pmin);
        return(float64(CUDD_OUT_OF_MEM));
    }
    return(min);

} /* end of bddCorrelationAux */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_bddCorrelationWeigths.]

  Description []

  SideEffects [None]

  SeeAlso     [bddCorrelationAux]

******************************************************************************/
func bddCorrelationWeightsAux(dd *DdManager, f *DdNode, g *DdNode, prob *float64, table *st__table) float64 {
    var Fv, Fnv, G, Gv, Gnv *DdNode;
    var min float64;
    var pmin *float64;
    var min1 float64;
    var min2 float64;
    var dummy *float64;
    var entry *HashEntry;
    var topF, topG, index int32;

    statLine(dd);
#ifdef CORREL_STATS
    num_calls++;
#endif

    /* Terminal cases: only work for BDDs. */
    if f == g { return(1.0); }
    // inlining of macro Cudd_Not
    if f == ((DdNode *)((ptrint)(g) ^ 01)) { return(0.0); }

    /* Standardize call using the following properties:
    **     (f EXNOR g)   = (g EXNOR f)
    **     (f' EXNOR g') = (f EXNOR g).
    */
    // inlining of macro cuddF2L
    // inlining of macro Cudd_Regular
    // inlining of macro Cudd_IsComplement
    if ((((DdNode *)((ptruint)(f) & ~01)).Id << 1) | (int32( ((ptrint) (f) & 01)))) > ((((DdNode *)((ptruint)(g) & ~01)).Id << 1) | (int32( ((ptrint) (g) & 01)))) {
        var tmp *DdNode  = f;
        f = g; g = tmp;
    }
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (f) & 01)) != 0 {
        // inlining of macro Cudd_Not
        f = ((DdNode *)((ptrint)(f) ^ 01));
        // inlining of macro Cudd_Not
        g = ((DdNode *)((ptrint)(g) ^ 01));
    }
    /* From now on, f is regular. */
    
    entry = new(HashEntry);
    if entry == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(float64(CUDD_OUT_OF_MEM));
    }
    entry.f = f; entry.g = g;

    /* We do not use the fact that
    ** correlation(f,g') = 1 - correlation(f,g)
    ** to minimize the risk of cancellation.
    */
    if st__lookup(table, (char *)entry, (char **)&dummy) {
        min = *dummy;
        ABC_FREE(entry);
        return(min);
    }

    // inlining of macro Cudd_Regular
    G = ((DdNode *)((ptruint)(g) & ~01));
    // inlining of macro cuddI
    
    if (f.index)==CUDD_CONST_INDEX {
        topF = int32((f.index));
    } else {
        topF = (dd).perm[(f.index)];
    // inlining of macro cuddI
    } 
    if (G.index)==CUDD_CONST_INDEX {
        topG = int32((G.index));
    } else {
        topG = (dd).perm[(G.index)];
    }
    if topF <= topG {
        // inlining of macro cuddT
        // inlining of macro cuddE
        Fv = ((f).type.kids.T); Fnv = ((f).type.kids.E);
        index = f.index;
    } else {
        Fv = Fnv = f;
        index = G.index;
    }
    // inlining of macro cuddT
    // inlining of macro cuddE
    if topG <= topF { Gv = ((G).type.kids.T); Gnv = ((G).type.kids.E); } else { Gv = Gnv = G; }

    if g != G {
        // inlining of macro Cudd_Not
        Gv = ((DdNode *)((ptrint)(Gv) ^ 01));
        // inlining of macro Cudd_Not
        Gnv = ((DdNode *)((ptrint)(Gnv) ^ 01));
    }

    min1 = bddCorrelationWeightsAux(dd, Fv, Gv, prob, table) * prob[index];
    if min1 == float64(CUDD_OUT_OF_MEM) {
        ABC_FREE(entry);
        return(float64(CUDD_OUT_OF_MEM));
    }
    min2 = bddCorrelationWeightsAux(dd, Fnv, Gnv, prob, table) * (1.0 - prob[index]); 
    if min2 == float64(CUDD_OUT_OF_MEM) {
        ABC_FREE(entry);
        return(float64(CUDD_OUT_OF_MEM));
    }
    min = (min1+min2);
    
    pmin = new(float64);
    if pmin == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(float64(CUDD_OUT_OF_MEM));
    }
    *pmin = min;

    if st__insert(table,(char *)entry, (char *)pmin) == st__OUT_OF_MEM {
        ABC_FREE(entry);
        ABC_FREE(pmin);
        return(float64(CUDD_OUT_OF_MEM));
    }
    return(min);

} /* end of bddCorrelationWeightsAux */


/**Function********************************************************************

  Synopsis    [Compares two hash table entries.]

  Description [Compares two hash table entries. Returns 0 if they are
  identical; 1 otherwise.]

  SideEffects [None]

******************************************************************************/
func CorrelCompare(key1 string, key2 string) int32 {
    var entry1 *HashEntry;
    var entry2 *HashEntry;

    entry1 = (HashEntry *) key1;
    entry2 = (HashEntry *) key2;
    if entry1.f != entry2.f || entry1.g != entry2.g { return(1); }

    return(0);

} /* end of CorrelCompare */


/**Function********************************************************************

  Synopsis    [Hashes a hash table entry.]

  Description [Hashes a hash table entry. It is patterned after
  st__strhash. Returns a value between 0 and modulus.]

  SideEffects [None]

******************************************************************************/
func CorrelHash(key string, modulus int32) int32 {
    var entry *HashEntry;
    var val int32;

    entry = (HashEntry *) key;
#if SIZEOF_VOID_P == 8 && SIZEOF_INT == 4
    val = ((int) ((long)entry->f))*997 + ((int) ((long)entry->g));
#else
    val = ((int) entry->f)*997 + ((int) entry->g);
#endif

    return ((val < 0) ? -val : val) % modulus;

} /* end of CorrelHash */


/**Function********************************************************************

  Synopsis    [Frees memory associated with hash table.]

  Description [Frees memory associated with hash table. Returns
  st__CONTINUE.]

  SideEffects [None]

******************************************************************************/
func CorrelCleanUp(key string, value string, arg string) st__retval {
    var d *float64;
    var entry *HashEntry;

    entry = (HashEntry *) key;
    ABC_FREE(entry);
    d = (double *)value;
    ABC_FREE(d);
    return st__CONTINUE;

} /* end of CorrelCleanUp */


//ABC_NAMESPACE_IMPL_END



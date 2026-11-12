package cudd

import "math"

/**CFile***********************************************************************

  FileName    [cuddAddApply.c]

  PackageName [cudd]

  Synopsis    [Apply functions for ADDs and their operators.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addApply()
                <li> Cudd_addMonadicApply()
                <li> Cudd_addPlus()
                <li> Cudd_addTimes()
                <li> Cudd_addThreshold()
                <li> Cudd_addSetNZ()
                <li> Cudd_addDivide()
                <li> Cudd_addMinus()
                <li> Cudd_addMinimum()
                <li> Cudd_addMaximum()
                <li> Cudd_addOneZeroMaximum()
                <li> Cudd_addDiff()
                <li> Cudd_addAgreement()
                <li> Cudd_addOr()
                <li> Cudd_addNand()
                <li> Cudd_addNor()
                <li> Cudd_addXor()
                <li> Cudd_addXnor()
                </ul>
            Internal procedures included in this module:
                <ul>
                <li> cuddAddApplyRecur()
                <li> cuddAddMonadicApplyRecur()
                </ul>]

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
static char rcsid[] DD_UNUSED = "$Id: cuddAddApply.c,v 1.18 2009/02/19 16:15:26 fabio Exp $";
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

  Synopsis    [Applies op to the corresponding discriminants of f and g.]

  Description [Applies op to the corresponding discriminants of f and g.
  Returns a pointer to the result if succssful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addMonadicApply Cudd_addPlus Cudd_addTimes
  Cudd_addThreshold Cudd_addSetNZ Cudd_addDivide Cudd_addMinus Cudd_addMinimum
  Cudd_addMaximum Cudd_addOneZeroMaximum Cudd_addDiff Cudd_addAgreement
  Cudd_addOr Cudd_addNand Cudd_addNor Cudd_addXor Cudd_addXnor]

******************************************************************************/
func Cudd_addApply(dd *DdManager, op DD_AOP, f *DdNode, g *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddAddApplyRecur(dd,op,f,g);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_addApply */


/**Function********************************************************************

  Synopsis    [Integer and floating point addition.]

  Description [Integer and floating point addition. Returns NULL if not
  a terminal case; f+g otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addPlus(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var res *DdNode;
    var F, G *DdNode;
    CUDD_VALUE_TYPE value;

    F = *f; G = *g;
    // inlining of macro DD_ZERO
    if F == ((dd).zero) { return(G); }
    // inlining of macro DD_ZERO
    if G == ((dd).zero) { return(F); }
    // inlining of macro cuddIsConstant
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) {
        // inlining of macro cuddV
        value = ((F).type.value)+((G).type.value);
        res = cuddUniqueConst(dd,value);
        return(res);
    }
    if F > G { /* swap f and g */
        *f = G;
        *g = F;
    }
    return(nil);

} /* end of Cudd_addPlus */


/**Function********************************************************************

  Synopsis    [Integer and floating point multiplication.]

  Description [Integer and floating point multiplication. Returns NULL
  if not a terminal case; f * g otherwise.  This function can be used also
  to take the AND of two 0-1 ADDs.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addTimes(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var res *DdNode;
    var F, G *DdNode;
    CUDD_VALUE_TYPE value;

    F = *f; G = *g;
    // inlining of macro DD_ZERO
    if F == ((dd).zero) || G == ((dd).zero) { return(((dd).zero)); }
    // inlining of macro DD_ONE
    if F == ((dd).one) { return(G); }
    // inlining of macro DD_ONE
    if G == ((dd).one) { return(F); }
    // inlining of macro cuddIsConstant
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) {
        // inlining of macro cuddV
        value = ((F).type.value)*((G).type.value);
        res = cuddUniqueConst(dd,value);
        return(res);
    }
    if F > G { /* swap f and g */
        *f = G;
        *g = F;
    }
    return(nil);

} /* end of Cudd_addTimes */


/**Function********************************************************************

  Synopsis    [f if f&gt;=g; 0 if f&lt;g.]

  Description [Threshold operator for Apply (f if f &gt;=g; 0 if f&lt;g).
  Returns NULL if not a terminal case; f op g otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addThreshold(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var F, G *DdNode;

    F = *f; G = *g;
    // inlining of macro DD_PLUS_INFINITY
    if F == G || F == ((dd).plusinfinity) { return(F); }
    // inlining of macro cuddIsConstant
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) {
        // inlining of macro cuddV
        if ((F).type.value) >= ((G).type.value) {
            return(F);
        } else {
            // inlining of macro DD_ZERO
            return(((dd).zero));
        }
    }
    return(nil);

} /* end of Cudd_addThreshold */


/**Function********************************************************************

  Synopsis    [This operator sets f to the value of g wherever g != 0.]

  Description [This operator sets f to the value of g wherever g != 0.
  Returns NULL if not a terminal case; f op g otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addSetNZ(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var F, G *DdNode;

    F = *f; G = *g;
    if F == G { return(F); }
    // inlining of macro DD_ZERO
    if F == ((dd).zero) { return(G); }
    // inlining of macro DD_ZERO
    if G == ((dd).zero) { return(F); }
    // inlining of macro cuddIsConstant
    if (G).index == CUDD_CONST_INDEX { return(G); }
    return(nil);

} /* end of Cudd_addSetNZ */


/**Function********************************************************************

  Synopsis    [Integer and floating point division.]

  Description [Integer and floating point division. Returns NULL if not
  a terminal case; f / g otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addDivide(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var res *DdNode;
    var F, G *DdNode;
    CUDD_VALUE_TYPE value;

    F = *f; G = *g;
    /* We would like to use F == G -> F/G == 1, but F and G may
    ** contain zeroes. */
    // inlining of macro DD_ZERO
    if F == ((dd).zero) { return(((dd).zero)); }
    // inlining of macro DD_ONE
    if G == ((dd).one) { return(F); }
    // inlining of macro cuddIsConstant
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) {
        // inlining of macro cuddV
        value = ((F).type.value)//*__starting_point_inline_macro:cuddV__*/((G)->type.value);
        res = cuddUniqueConst(dd,value);
        return(res);
    }
    return(nil);

} /* end of Cudd_addDivide */


/**Function********************************************************************

  Synopsis    [Integer and floating point subtraction.]

  Description [Integer and floating point subtraction. Returns NULL if
  not a terminal case; f - g otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addMinus(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var res *DdNode;
    var F, G *DdNode;
    CUDD_VALUE_TYPE value;

    F = *f; G = *g;
    // inlining of macro DD_ZERO
    if F == G { return(((dd).zero)); }
    // inlining of macro DD_ZERO
    if F == ((dd).zero) { return(cuddAddNegateRecur(dd,G)); }
    // inlining of macro DD_ZERO
    if G == ((dd).zero) { return(F); }
    // inlining of macro cuddIsConstant
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) {
        // inlining of macro cuddV
        value = ((F).type.value)-((G).type.value);
        res = cuddUniqueConst(dd,value);
        return(res);
    }
    return(nil);

} /* end of Cudd_addMinus */


/**Function********************************************************************

  Synopsis    [Integer and floating point min.]

  Description [Integer and floating point min for Cudd_addApply.
  Returns NULL if not a terminal case; min(f,g) otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addMinimum(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var F, G *DdNode;

    F = *f; G = *g;
    // inlining of macro DD_PLUS_INFINITY
    if F == ((dd).plusinfinity) { return(G); }
    // inlining of macro DD_PLUS_INFINITY
    if G == ((dd).plusinfinity) { return(F); }
    if F == G { return(F); }
     
                                                      
                                              
                                              
      
    // inlining of macro cuddIsConstant
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) {
        // inlining of macro cuddV
        if ((F).type.value) <= ((G).type.value) {
            return(F);
        } else {
            return(G);
        }
    }
    if F > G { /* swap f and g */
        *f = G;
        *g = F;
    }
    return(nil);

} /* end of Cudd_addMinimum */


/**Function********************************************************************

  Synopsis    [Integer and floating point max.]

  Description [Integer and floating point max for Cudd_addApply.
  Returns NULL if not a terminal case; max(f,g) otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addMaximum(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var F, G *DdNode;

    F = *f; G = *g;
    if F == G { return(F); }
    // inlining of macro DD_MINUS_INFINITY
    if F == ((dd).minusinfinity) { return(G); }
    // inlining of macro DD_MINUS_INFINITY
    if G == ((dd).minusinfinity) { return(F); }
     
                                                      
                                             
                                             
      
    // inlining of macro cuddIsConstant
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) {
        // inlining of macro cuddV
        if ((F).type.value) >= ((G).type.value) {
            return(F);
        } else {
            return(G);
        }
    }
    if F > G { /* swap f and g */
        *f = G;
        *g = F;
    }
    return(nil);

} /* end of Cudd_addMaximum */


/**Function********************************************************************

  Synopsis    [Returns 1 if f &gt; g and 0 otherwise.]

  Description [Returns 1 if f &gt; g and 0 otherwise. Used in
  conjunction with Cudd_addApply. Returns NULL if not a terminal
  case.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addOneZeroMaximum(dd *DdManager, f **DdNode, g **DdNode) *DdNode {

    // inlining of macro DD_ZERO
    if *f == *g { return(((dd).zero)); }
    // inlining of macro DD_PLUS_INFINITY
    if *g == ((dd).plusinfinity) {
        // inlining of macro DD_ZERO
        return ((dd).zero); }
    // inlining of macro cuddIsConstant
    if ((*f).index == CUDD_CONST_INDEX) && ((*g).index == CUDD_CONST_INDEX) {
        // inlining of macro cuddV
        if ((*f).type.value) > ((*g).type.value) {
            // inlining of macro DD_ONE
            return(((dd).one));
        } else {
            // inlining of macro DD_ZERO
            return(((dd).zero));
        }
    }

    return(nil);

} /* end of Cudd_addOneZeroMaximum */


/**Function********************************************************************

  Synopsis    [Returns plusinfinity if f=g; returns min(f,g) if f!=g.]

  Description [Returns NULL if not a terminal case; f op g otherwise,
  where f op g is plusinfinity if f=g; min(f,g) if f!=g.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addDiff(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var F, G *DdNode;

    F = *f; G = *g;
    // inlining of macro DD_PLUS_INFINITY
    if F == G { return(((dd).plusinfinity)); }
    // inlining of macro DD_PLUS_INFINITY
    if F == ((dd).plusinfinity) { return(G); }
    // inlining of macro DD_PLUS_INFINITY
    if G == ((dd).plusinfinity) { return(F); }
    // inlining of macro cuddIsConstant
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) {
        // inlining of macro cuddV
        if ((F).type.value) != ((G).type.value) {
            // inlining of macro cuddV
            if ((F).type.value) < ((G).type.value) {
                return(F);
            } else {
                return(G);
            }
        } else {
            // inlining of macro DD_PLUS_INFINITY
            return(((dd).plusinfinity));
        }
    }
    return(nil);

} /* end of Cudd_addDiff */


/**Function********************************************************************

  Synopsis    [f if f==g; background if f!=g.]

  Description [Returns NULL if not a terminal case; f op g otherwise,
  where f op g is f if f==g; background if f!=g.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addAgreement(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var F, G *DdNode;

    F = *f; G = *g;
    if F == G { return(F); }
    if F == dd.background { return(F); }
    if G == dd.background { return(G); }
    // inlining of macro cuddIsConstant
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) { return(dd.background); }
    return(nil);

} /* end of Cudd_addAgreement */


/**Function********************************************************************

  Synopsis    [Disjunction of two 0-1 ADDs.]

  Description [Disjunction of two 0-1 ADDs. Returns NULL
  if not a terminal case; f OR g otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addOr(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var F, G *DdNode;

    F = *f; G = *g;
    // inlining of macro DD_ONE
    if F == ((dd).one) || G == ((dd).one) { return(((dd).one)); }
    // inlining of macro cuddIsConstant
    if (F).index == CUDD_CONST_INDEX { return(G); }
    // inlining of macro cuddIsConstant
    if (G).index == CUDD_CONST_INDEX { return(F); }
    if F == G { return(F); }
    if F > G { /* swap f and g */
        *f = G;
        *g = F;
    }
    return(nil);

} /* end of Cudd_addOr */


/**Function********************************************************************

  Synopsis    [NAND of two 0-1 ADDs.]

  Description [NAND of two 0-1 ADDs. Returns NULL
  if not a terminal case; f NAND g otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addNand(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var F, G *DdNode;

    F = *f; G = *g;
    // inlining of macro DD_ZERO
    // inlining of macro DD_ONE
    if F == ((dd).zero) || G == ((dd).zero) { return(((dd).one)); }
    // inlining of macro cuddIsConstant
    // inlining of macro DD_ZERO
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) { return(((dd).zero)); }
    if F > G { /* swap f and g */
        *f = G;
        *g = F;
    }
    return(nil);

} /* end of Cudd_addNand */


/**Function********************************************************************

  Synopsis    [NOR of two 0-1 ADDs.]

  Description [NOR of two 0-1 ADDs. Returns NULL
  if not a terminal case; f NOR g otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addNor(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var F, G *DdNode;

    F = *f; G = *g;
    // inlining of macro DD_ONE
    // inlining of macro DD_ZERO
    if F == ((dd).one) || G == ((dd).one) { return(((dd).zero)); }
    // inlining of macro cuddIsConstant
    // inlining of macro DD_ONE
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) { return(((dd).one)); }
    if F > G { /* swap f and g */
        *f = G;
        *g = F;
    }
    return(nil);

} /* end of Cudd_addNor */


/**Function********************************************************************

  Synopsis    [XOR of two 0-1 ADDs.]

  Description [XOR of two 0-1 ADDs. Returns NULL
  if not a terminal case; f XOR g otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addXor(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var F, G *DdNode;

    F = *f; G = *g;
    // inlining of macro DD_ZERO
    if F == G { return(((dd).zero)); }
    // inlining of macro DD_ONE
    // inlining of macro DD_ZERO
    if F == ((dd).one) && G == ((dd).zero) { return(((dd).one)); }
    // inlining of macro DD_ONE
    // inlining of macro DD_ZERO
    if G == ((dd).one) && F == ((dd).zero) { return(((dd).one)); }
    // inlining of macro cuddIsConstant
    // inlining of macro DD_ZERO
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) { return(((dd).zero)); }
    if F > G { /* swap f and g */
        *f = G;
        *g = F;
    }
    return(nil);

} /* end of Cudd_addXor */


/**Function********************************************************************

  Synopsis    [XNOR of two 0-1 ADDs.]

  Description [XNOR of two 0-1 ADDs. Returns NULL
  if not a terminal case; f XNOR g otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply]

******************************************************************************/
func Cudd_addXnor(dd *DdManager, f **DdNode, g **DdNode) *DdNode {
    var F, G *DdNode;

    F = *f; G = *g;
    // inlining of macro DD_ONE
    if F == G { return(((dd).one)); }
    // inlining of macro DD_ONE
    if F == ((dd).one) && G == ((dd).one) { return(((dd).one)); }
    // inlining of macro DD_ZERO
    // inlining of macro DD_ONE
    if G == ((dd).zero) && F == ((dd).zero) { return(((dd).one)); }
    // inlining of macro cuddIsConstant
    // inlining of macro DD_ZERO
    if ((F).index == CUDD_CONST_INDEX) && ((G).index == CUDD_CONST_INDEX) { return(((dd).zero)); }
    if F > G { /* swap f and g */
        *f = G;
        *g = F;
    }
    return(nil);

} /* end of Cudd_addXnor */


/**Function********************************************************************

  Synopsis    [Applies op to the discriminants of f.]

  Description [Applies op to the discriminants of f.
  Returns a pointer to the result if succssful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addApply Cudd_addLog]

******************************************************************************/
func Cudd_addMonadicApply(dd *DdManager, op DD_MAOP, f *DdNode) *DdNode {
    var res *DdNode;

    for {
        dd.reordered = 0;
        res = cuddAddMonadicApplyRecur(dd,op,f);
    
if !(dd.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_addMonadicApply */


/**Function********************************************************************

  Synopsis    [Natural logarithm of an ADD.]

  Description [Natural logarithm of an ADDs. Returns NULL
  if not a terminal case; log(f) otherwise.  The discriminants of f must
  be positive double's.]

  SideEffects [None]

  SeeAlso     [Cudd_addMonadicApply]

******************************************************************************/
func Cudd_addLog(dd *DdManager, f *DdNode) *DdNode {
    // inlining of macro cuddIsConstant
    if (f).index == CUDD_CONST_INDEX {
        // inlining of macro cuddV
        CUDD_VALUE_TYPE value = math.Log(float64(((f).type.value)));
        var res *DdNode  = cuddUniqueConst(dd,value);
        return(res);
    }
    return(nil);

} /* end of Cudd_addLog */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_addApply.]

  Description [Performs the recursive step of Cudd_addApply. Returns a
  pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [cuddAddMonadicApplyRecur]

******************************************************************************/
func cuddAddApplyRecur(dd *DdManager, op DD_AOP, f *DdNode, g *DdNode) *DdNode {
    var res, fv, fvn, gv, gvn, T, E *DdNode;
    unsigned int ford, gord;
    unsigned int index;
    DD_CTFP cacheOp;

    /* Check terminal cases. Op may swap f and g to increase the
     * cache hit rate.
     */
    statLine(dd);
    res = (*op)(dd,&f,&g);
    if res != nil { return(res); }

    /* Check cache. */
    cacheOp = (DD_CTFP) op;
    res = cuddCacheLookup2(dd,cacheOp,f,g);
    if res != nil { return(res); }

    /* Recursive step. */
    // inlining of macro cuddI
    
    if (f.index)==CUDD_CONST_INDEX {
        ford = int32((f.index));
    } else {
        ford = (dd).perm[(f.index)];
    }
    // inlining of macro cuddI
    
    if (g.index)==CUDD_CONST_INDEX {
        gord = int32((g.index));
    } else {
        gord = (dd).perm[(g.index)];
    }
    if ford <= gord {
        index = f.index;
        // inlining of macro cuddT
        fv = ((f).type.kids.T);
        // inlining of macro cuddE
        fvn = ((f).type.kids.E);
    } else {
        index = g.index;
        fv = fvn = f;
    }
    if gord <= ford {
        // inlining of macro cuddT
        gv = ((g).type.kids.T);
        // inlining of macro cuddE
        gvn = ((g).type.kids.E);
    } else {
        gv = gvn = g;
    }

    T = cuddAddApplyRecur(dd,op,fv,gv);
    if T == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);

    E = cuddAddApplyRecur(dd,op,fvn,gvn);
    if E == nil {
        Cudd_RecursiveDeref(dd,T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);

    if T == E {
        res = T;
    } else {
        res = cuddUniqueInter(dd,int32(index),T,E);
    }
    if res == nil {
        Cudd_RecursiveDeref(dd, T);
        Cudd_RecursiveDeref(dd, E);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(T).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(E).ref);

    /* Store result. */
    cuddCacheInsert2(dd,cacheOp,f,g,res);

    return(res);

} /* end of cuddAddApplyRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_addMonadicApply.]

  Description [Performs the recursive step of Cudd_addMonadicApply. Returns a
  pointer to the result if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [cuddAddApplyRecur]

******************************************************************************/
func cuddAddMonadicApplyRecur(dd *DdManager, op DD_MAOP, f *DdNode) *DdNode {
    var res, ft, fe, T, E *DdNode;
    unsigned int index;

    /* Check terminal cases. */
    statLine(dd);
    res = (*op)(dd,f);
    if res != nil { return(res); }

    /* Check cache. */
    res = cuddCacheLookup1(dd,op,f);
    if res != nil { return(res); }

    /* Recursive step. */
    index = f.index;
    // inlining of macro cuddT
    ft = ((f).type.kids.T);
    // inlining of macro cuddE
    fe = ((f).type.kids.E);

    T = cuddAddMonadicApplyRecur(dd,op,ft);
    if T == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(T).ref);

    E = cuddAddMonadicApplyRecur(dd,op,fe);
    if E == nil {
        Cudd_RecursiveDeref(dd,T);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(E).ref);

    if T == E {
        res = T;
    } else {
        res = cuddUniqueInter(dd,int32(index),T,E);
    }
    if res == nil {
        Cudd_RecursiveDeref(dd, T);
        Cudd_RecursiveDeref(dd, E);
        return(nil);
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(T).ref);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(E).ref);

    /* Store result. */
    cuddCacheInsert1(dd,op,f,res);

    return(res);

} /* end of cuddAddMonadicApplyRecur */


/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


//ABC_NAMESPACE_IMPL_END


package cudd

import "fmt"

/**CFile***********************************************************************

  FileName    [cuddAddAbs.c]

  PackageName [cudd]

  Synopsis    [Quantification functions for ADDs.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addExistAbstract()
                <li> Cudd_addUnivAbstract()
                <li> Cudd_addOrAbstract()
                </ul>
        Internal procedures included in this module:
                <ul>
                <li> cuddAddExistAbstractRecur()
                <li> cuddAddUnivAbstractRecur()
                <li> cuddAddOrAbstractRecur()
                </ul>
        Static procedures included in this module:
                <ul>
                <li> addCheckPositiveCube()
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
static char rcsid[] DD_UNUSED = "$Id: cuddAddAbs.c,v 1.15 2004/08/13 18:04:45 fabio Exp $";
#endif

var two *DdNode;

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/


/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static int addCheckPositiveCube (DdManager *manager, DdNode *cube);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/

/**Function********************************************************************

  Synopsis    [Existentially Abstracts all the variables in cube from f.]

  Description [Abstracts all the variables in cube from f by summing
  over all possible values taken by the variables. Returns the
  abstracted ADD.]

  SideEffects [None]

  SeeAlso     [Cudd_addUnivAbstract Cudd_bddExistAbstract
  Cudd_addOrAbstract]

******************************************************************************/
func Cudd_addExistAbstract(manager *DdManager, f *DdNode, cube *DdNode) *DdNode {
    var res *DdNode;

    two = cuddUniqueConst(manager,(CUDD_VALUE_TYPE) 2);
    if two == nil { return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(two).ref);

    if addCheckPositiveCube(manager, cube) == 0 {
        (void) fmt.Fprintf(manager.err, "Error: Can only abstract cubes");
        return(nil);
    }

    for {
        manager.reordered = 0;
        res = cuddAddExistAbstractRecur(manager, f, cube);
    
if !(manager.reordered == 1) { break }
} 

    if res == nil {
        Cudd_RecursiveDeref(manager,two);
        return(nil);
    }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(res).ref);
    Cudd_RecursiveDeref(manager,two);
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(res).ref);

    return(res);

} /* end of Cudd_addExistAbstract */


/**Function********************************************************************

  Synopsis    [Universally Abstracts all the variables in cube from f.]

  Description [Abstracts all the variables in cube from f by taking
  the product over all possible values taken by the variable. Returns
  the abstracted ADD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addExistAbstract Cudd_bddUnivAbstract
  Cudd_addOrAbstract]

******************************************************************************/
func Cudd_addUnivAbstract(manager *DdManager, f *DdNode, cube *DdNode) *DdNode {
    var res *DdNode;

    if addCheckPositiveCube(manager, cube) == 0 {
        (void) fmt.Fprintf(manager.err, "Error:  Can only abstract cubes");
        return(nil);
    }

    for {
        manager.reordered = 0;
        res = cuddAddUnivAbstractRecur(manager, f, cube);
    
if !(manager.reordered == 1) { break }
} 

    return(res);

} /* end of Cudd_addUnivAbstract */


/**Function********************************************************************

  Synopsis    [Disjunctively abstracts all the variables in cube from the
  0-1 ADD f.]

  Description [Abstracts all the variables in cube from the 0-1 ADD f
  by taking the disjunction over all possible values taken by the
  variables.  Returns the abstracted ADD if successful; NULL
  otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addUnivAbstract Cudd_addExistAbstract]

******************************************************************************/
func Cudd_addOrAbstract(manager *DdManager, f *DdNode, cube *DdNode) *DdNode {
    var res *DdNode;

    if addCheckPositiveCube(manager, cube) == 0 {
        (void) fmt.Fprintf(manager.err, "Error: Can only abstract cubes");
        return(nil);
    }

    for {
        manager.reordered = 0;
        res = cuddAddOrAbstractRecur(manager, f, cube);
    
if !(manager.reordered == 1) { break }
} 
    return(res);

} /* end of Cudd_addOrAbstract */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_addExistAbstract.]

  Description [Performs the recursive step of Cudd_addExistAbstract.
  Returns the ADD obtained by abstracting the variables of cube from f,
  if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddAddExistAbstractRecur(manager *DdManager, f *DdNode, cube *DdNode) *DdNode {
    var T, E, res, res1, res2, zero *DdNode;

    statLine(manager);
    // inlining of macro DD_ZERO
    zero = ((manager).zero);

    /* Cube is guaranteed to be a cube at this point. */        
    // inlining of macro cuddIsConstant
    if f == zero || ((cube).index == CUDD_CONST_INDEX) {  
        return(f);
    }

    /* Abstract a variable that does not appear in f => multiply by 2. */
    // inlining of macro cuddI
    if (((f.index)==CUDD_CONST_INDEX)?int32((f.index)):(manager).perm[(f.index)]) > (((cube.index)==CUDD_CONST_INDEX)?int32((cube.index)):(manager).perm[(cube.index)]) {
        // inlining of macro cuddT
        res1 = cuddAddExistAbstractRecur(manager, f, ((cube).type.kids.T));
        if res1 == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        /* Use the "internal" procedure to be alerted in case of
        ** dynamic reordering. If dynamic reordering occurs, we
        ** have to abort the entire abstraction.
        */
        res = cuddAddApplyRecur(manager,Cudd_addTimes,res1,two);
        if res == nil {
            Cudd_RecursiveDeref(manager,res1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_RecursiveDeref(manager,res1);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref);
        return(res);
    }

    if (res = cuddCacheLookup2(manager, Cudd_addExistAbstract, f, cube)) != nil {
        return(res);
    }

    // inlining of macro cuddT
    T = ((f).type.kids.T);
    // inlining of macro cuddE
    E = ((f).type.kids.E);

    /* If the two indices are the same, so are their levels. */
    if f.index == cube.index {
        // inlining of macro cuddT
        res1 = cuddAddExistAbstractRecur(manager, T, ((cube).type.kids.T));
        if res1 == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        // inlining of macro cuddT
        res2 = cuddAddExistAbstractRecur(manager, E, ((cube).type.kids.T));
        if res2 == nil {
            Cudd_RecursiveDeref(manager,res1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res2).ref);
        res = cuddAddApplyRecur(manager, Cudd_addPlus, res1, res2);
        if res == nil {
            Cudd_RecursiveDeref(manager,res1);
            Cudd_RecursiveDeref(manager,res2);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_RecursiveDeref(manager,res1);
        Cudd_RecursiveDeref(manager,res2);
        cuddCacheInsert2(manager, Cudd_addExistAbstract, f, cube, res);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref);
        return(res);
    } else { /* if (cuddI(manager,f->index) < cuddI(manager,cube->index)) */
        res1 = cuddAddExistAbstractRecur(manager, T, cube);
        if res1 == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        res2 = cuddAddExistAbstractRecur(manager, E, cube);
        if res2 == nil {
            Cudd_RecursiveDeref(manager,res1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res2).ref);
        if res1 == res2 {
            res = res1;
        } else {
            res = cuddUniqueInter(manager, int32( f.index), res1, res2);
        }
        if res == nil {
            Cudd_RecursiveDeref(manager,res1);
            Cudd_RecursiveDeref(manager,res2);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res1).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res2).ref);
        cuddCacheInsert2(manager, Cudd_addExistAbstract, f, cube, res);
        return(res);
    }       

} /* end of cuddAddExistAbstractRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_addUnivAbstract.]

  Description [Performs the recursive step of Cudd_addUnivAbstract.
  Returns the ADD obtained by abstracting the variables of cube from f,
  if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddAddUnivAbstractRecur(manager *DdManager, f *DdNode, cube *DdNode) *DdNode {
    var T, E, res, res1, res2, one, zero *DdNode;

    statLine(manager);
    // inlining of macro DD_ONE
    one = ((manager).one);
    // inlining of macro DD_ZERO
    zero = ((manager).zero);

    /* Cube is guaranteed to be a cube at this point.
    ** zero and one are the only constatnts c such that c*c=c.
    */
    if f == zero || f == one || cube == one {  
        return(f);
    }

    /* Abstract a variable that does not appear in f. */
    // inlining of macro cuddI
    if (((f.index)==CUDD_CONST_INDEX)?int32((f.index)):(manager).perm[(f.index)]) > (((cube.index)==CUDD_CONST_INDEX)?int32((cube.index)):(manager).perm[(cube.index)]) {
        // inlining of macro cuddT
        res1 = cuddAddUnivAbstractRecur(manager, f, ((cube).type.kids.T));
        if res1 == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        /* Use the "internal" procedure to be alerted in case of
        ** dynamic reordering. If dynamic reordering occurs, we
        ** have to abort the entire abstraction.
        */
        res = cuddAddApplyRecur(manager, Cudd_addTimes, res1, res1);
        if res == nil {
            Cudd_RecursiveDeref(manager,res1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_RecursiveDeref(manager,res1);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref);
        return(res);
    }

    if (res = cuddCacheLookup2(manager, Cudd_addUnivAbstract, f, cube)) != nil {
        return(res);
    }

    // inlining of macro cuddT
    T = ((f).type.kids.T);
    // inlining of macro cuddE
    E = ((f).type.kids.E);

    /* If the two indices are the same, so are their levels. */
    if f.index == cube.index {
        // inlining of macro cuddT
        res1 = cuddAddUnivAbstractRecur(manager, T, ((cube).type.kids.T));
        if res1 == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        // inlining of macro cuddT
        res2 = cuddAddUnivAbstractRecur(manager, E, ((cube).type.kids.T));
        if res2 == nil {
            Cudd_RecursiveDeref(manager,res1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res2).ref);
        res = cuddAddApplyRecur(manager, Cudd_addTimes, res1, res2);
        if res == nil {
            Cudd_RecursiveDeref(manager,res1);
            Cudd_RecursiveDeref(manager,res2);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_RecursiveDeref(manager,res1);
        Cudd_RecursiveDeref(manager,res2);
        cuddCacheInsert2(manager, Cudd_addUnivAbstract, f, cube, res);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref);
        return(res);
    } else { /* if (cuddI(manager,f->index) < cuddI(manager,cube->index)) */
        res1 = cuddAddUnivAbstractRecur(manager, T, cube);
        if res1 == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        res2 = cuddAddUnivAbstractRecur(manager, E, cube);
        if res2 == nil {
            Cudd_RecursiveDeref(manager,res1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res2).ref);
        if res1 == res2 {
            res = res1;
        } else {
            res = cuddUniqueInter(manager, int32( f.index), res1, res2);
        }
        if res == nil {
            Cudd_RecursiveDeref(manager,res1);
            Cudd_RecursiveDeref(manager,res2);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res1).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res2).ref);
        cuddCacheInsert2(manager, Cudd_addUnivAbstract, f, cube, res);
        return(res);
    }

} /* end of cuddAddUnivAbstractRecur */


/**Function********************************************************************

  Synopsis    [Performs the recursive step of Cudd_addOrAbstract.]

  Description [Performs the recursive step of Cudd_addOrAbstract.
  Returns the ADD obtained by abstracting the variables of cube from f,
  if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func cuddAddOrAbstractRecur(manager *DdManager, f *DdNode, cube *DdNode) *DdNode {
    var T, E, res, res1, res2, one *DdNode;

    statLine(manager);
    // inlining of macro DD_ONE
    one = ((manager).one);

    /* Cube is guaranteed to be a cube at this point. */
    // inlining of macro cuddIsConstant
    if ((f).index == CUDD_CONST_INDEX) || cube == one {  
        return(f);
    }

    /* Abstract a variable that does not appear in f. */
    // inlining of macro cuddI
    if (((f.index)==CUDD_CONST_INDEX)?int32((f.index)):(manager).perm[(f.index)]) > (((cube.index)==CUDD_CONST_INDEX)?int32((cube.index)):(manager).perm[(cube.index)]) {
        // inlining of macro cuddT
        res = cuddAddOrAbstractRecur(manager, f, ((cube).type.kids.T));
        return(res);
    }

    if (res = cuddCacheLookup2(manager, Cudd_addOrAbstract, f, cube)) != nil {
        return(res);
    }

    // inlining of macro cuddT
    T = ((f).type.kids.T);
    // inlining of macro cuddE
    E = ((f).type.kids.E);

    /* If the two indices are the same, so are their levels. */
    if f.index == cube.index {
        // inlining of macro cuddT
        res1 = cuddAddOrAbstractRecur(manager, T, ((cube).type.kids.T));
        if res1 == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        if res1 != one {
            // inlining of macro cuddT
            res2 = cuddAddOrAbstractRecur(manager, E, ((cube).type.kids.T));
            if res2 == nil {
                Cudd_RecursiveDeref(manager,res1);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res2).ref);
            res = cuddAddApplyRecur(manager, Cudd_addOr, res1, res2);
            if res == nil {
                Cudd_RecursiveDeref(manager,res1);
                Cudd_RecursiveDeref(manager,res2);
                return(nil);
            }
            // inlining of macro cuddRef
            cuddSatInc(Cudd_Regular(res).ref);
            Cudd_RecursiveDeref(manager,res1);
            Cudd_RecursiveDeref(manager,res2);
        } else {
            res = res1;
        }
        cuddCacheInsert2(manager, Cudd_addOrAbstract, f, cube, res);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res).ref);
        return(res);
    } else { /* if (cuddI(manager,f->index) < cuddI(manager,cube->index)) */
        res1 = cuddAddOrAbstractRecur(manager, T, cube);
        if res1 == nil { return(nil); }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res1).ref);
        res2 = cuddAddOrAbstractRecur(manager, E, cube);
        if res2 == nil {
            Cudd_RecursiveDeref(manager,res1);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res2).ref);
        if res1 == res2 {
            res = res1;
        } else {
            res = cuddUniqueInter(manager, int32( f.index), res1, res2);
        }
        if res == nil {
            Cudd_RecursiveDeref(manager,res1);
            Cudd_RecursiveDeref(manager,res2);
            return(nil);
        }
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res1).ref);
        // inlining of macro cuddDeref
        cuddSatDec(Cudd_Regular(res2).ref);
        cuddCacheInsert2(manager, Cudd_addOrAbstract, f, cube, res);
        return(res);
    }

} /* end of cuddAddOrAbstractRecur */



/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Checks whether cube is an ADD representing the product
  of positive literals.]

  Description [Checks whether cube is an ADD representing the product of
  positive literals. Returns 1 in case of success; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func addCheckPositiveCube(manager *DdManager, cube *DdNode) int32 {
    // inlining of macro Cudd_IsComplement
    if int32( ((ptrint) (cube) & 01)) != 0 { return(0); }
    // inlining of macro DD_ONE
    if cube == ((manager).one) { return(1); }
    // inlining of macro cuddIsConstant
    if (cube).index == CUDD_CONST_INDEX { return(0); }
    // inlining of macro cuddE
    // inlining of macro DD_ZERO
    if ((cube).type.kids.E) == ((manager).zero) {
        // inlining of macro cuddT
        return(addCheckPositiveCube(manager, ((cube).type.kids.T)));
    }
    return(0);

} /* end of addCheckPositiveCube */


//ABC_NAMESPACE_IMPL_END




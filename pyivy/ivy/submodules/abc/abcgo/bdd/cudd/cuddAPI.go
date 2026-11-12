package cudd

import (
	"fmt"
	"math"
	"strings"
)

import "os"

/**CFile***********************************************************************

  FileName    [cuddAPI.c]

  PackageName [cudd]

  Synopsis    [Application interface functions.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_addNewVar()
                <li> Cudd_addNewVarAtLevel()
                <li> Cudd_bddNewVar()
                <li> Cudd_bddNewVarAtLevel()
                <li> Cudd_addIthVar()
                <li> Cudd_bddIthVar()
                <li> Cudd_zddIthVar()
                <li> Cudd_zddVarsFromBddVars()
                <li> Cudd_addConst()
                <li> Cudd_IsNonConstant()
                <li> Cudd_AutodynEnable()
                <li> Cudd_AutodynDisable()
                <li> Cudd_ReorderingStatus()
                <li> Cudd_AutodynEnableZdd()
                <li> Cudd_AutodynDisableZdd()
                <li> Cudd_ReorderingStatusZdd()
                <li> Cudd_zddRealignmentEnabled()
                <li> Cudd_zddRealignEnable()
                <li> Cudd_zddRealignDisable()
                <li> Cudd_bddRealignmentEnabled()
                <li> Cudd_bddRealignEnable()
                <li> Cudd_bddRealignDisable()
                <li> Cudd_ReadOne()
                <li> Cudd_ReadZddOne()
                <li> Cudd_ReadZero()
                <li> Cudd_ReadLogicZero()
                <li> Cudd_ReadPlusInfinity()
                <li> Cudd_ReadMinusInfinity()
                <li> Cudd_ReadBackground()
                <li> Cudd_SetBackground()
                <li> Cudd_ReadCacheSlots()
                <li> Cudd_ReadCacheUsedSlots()
                <li> Cudd_ReadCacheLookUps()
                <li> Cudd_ReadCacheHits()
                <li> Cudd_ReadMinHit()
                <li> Cudd_SetMinHit()
                <li> Cudd_ReadLooseUpTo()
                <li> Cudd_SetLooseUpTo()
                <li> Cudd_ReadMaxCache()
                <li> Cudd_ReadMaxCacheHard()
                <li> Cudd_SetMaxCacheHard()
                <li> Cudd_ReadSize()
                <li> Cudd_ReadSlots()
                <li> Cudd_ReadUsedSlots()
                <li> Cudd_ExpectedUsedSlots()
                <li> Cudd_ReadKeys()
                <li> Cudd_ReadDead()
                <li> Cudd_ReadMinDead()
                <li> Cudd_ReadReorderings()
                <li> Cudd_ReadReorderingTime()
                <li> Cudd_ReadGarbageCollections()
                <li> Cudd_ReadGarbageCollectionTime()
                <li> Cudd_ReadNodesFreed()
                <li> Cudd_ReadNodesDropped()
                <li> Cudd_ReadUniqueLookUps()
                <li> Cudd_ReadUniqueLinks()
                <li> Cudd_ReadSiftMaxVar()
                <li> Cudd_SetSiftMaxVar()
                <li> Cudd_ReadMaxGrowth()
                <li> Cudd_SetMaxGrowth()
                <li> Cudd_ReadMaxGrowthAlternate()
                <li> Cudd_SetMaxGrowthAlternate()
                <li> Cudd_ReadReorderingCycle()
                <li> Cudd_SetReorderingCycle()
                <li> Cudd_ReadTree()
                <li> Cudd_SetTree()
                <li> Cudd_FreeTree()
                <li> Cudd_ReadZddTree()
                <li> Cudd_SetZddTree()
                <li> Cudd_FreeZddTree()
                <li> Cudd_NodeReadIndex()
                <li> Cudd_ReadPerm()
                <li> Cudd_ReadInvPerm()
                <li> Cudd_ReadVars()
                <li> Cudd_ReadEpsilon()
                <li> Cudd_SetEpsilon()
                <li> Cudd_ReadGroupCheck()
                <li> Cudd_SetGroupcheck()
                <li> Cudd_GarbageCollectionEnabled()
                <li> Cudd_EnableGarbageCollection()
                <li> Cudd_DisableGarbageCollection()
                <li> Cudd_DeadAreCounted()
                <li> Cudd_TurnOnCountDead()
                <li> Cudd_TurnOffCountDead()
                <li> Cudd_ReadRecomb()
                <li> Cudd_SetRecomb()
                <li> Cudd_ReadSymmviolation()
                <li> Cudd_SetSymmviolation()
                <li> Cudd_ReadArcviolation()
                <li> Cudd_SetArcviolation()
                <li> Cudd_ReadPopulationSize()
                <li> Cudd_SetPopulationSize()
                <li> Cudd_ReadNumberXovers()
                <li> Cudd_SetNumberXovers()
                <li> Cudd_ReadMemoryInUse()
                <li> Cudd_PrintInfo()
                <li> Cudd_ReadPeakNodeCount()
                <li> Cudd_ReadPeakLiveNodeCount()
                <li> Cudd_ReadNodeCount()
                <li> Cudd_zddReadNodeCount()
                <li> Cudd_AddHook()
                <li> Cudd_RemoveHook()
                <li> Cudd_IsInHook()
                <li> Cudd_StdPreReordHook()
                <li> Cudd_StdPostReordHook()
                <li> Cudd_EnableReorderingReporting()
                <li> Cudd_DisableReorderingReporting()
                <li> Cudd_ReorderingReporting()
                <li> Cudd_ReadErrorCode()
                <li> Cudd_ClearErrorCode()
                <li> Cudd_ReadStdout()
                <li> Cudd_SetStdout()
                <li> Cudd_ReadStderr()
                <li> Cudd_SetStderr()
                <li> Cudd_ReadNextReordering()
                <li> Cudd_SetNextReordering()
                <li> Cudd_ReadSwapSteps()
                <li> Cudd_ReadMaxLive()
                <li> Cudd_SetMaxLive()
                <li> Cudd_ReadMaxMemory()
                <li> Cudd_SetMaxMemory()
                <li> Cudd_bddBindVar()
                <li> Cudd_bddUnbindVar()
                <li> Cudd_bddVarIsBound()
                <li> Cudd_bddSetPiVar()
                <li> Cudd_bddSetPsVar()
                <li> Cudd_bddSetNsVar()
                <li> Cudd_bddIsPiVar()
                <li> Cudd_bddIsPsVar()
                <li> Cudd_bddIsNsVar()
                <li> Cudd_bddSetPairIndex()
                <li> Cudd_bddReadPairIndex()
                <li> Cudd_bddSetVarToBeGrouped()
                <li> Cudd_bddSetVarHardGroup()
                <li> Cudd_bddResetVarToBeGrouped()
                <li> Cudd_bddIsVarToBeGrouped()
                <li> Cudd_bddSetVarToBeUngrouped()
                <li> Cudd_bddIsVarToBeUngrouped()
                <li> Cudd_bddIsVarHardGroup()
                </ul>
              Static procedures included in this module:
                <ul>
                <li> fixVarTree()
                </ul>]

  SeeAlso     []

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
static char rcsid[] DD_UNUSED = "$Id: cuddAPI.c,v 1.59 2009/02/19 16:14:14 fabio Exp $";
#endif

/*---------------------------------------------------------------------------*/
/* Macro declarations                                                        */
/*---------------------------------------------------------------------------*/

/**AutomaticStart*************************************************************/

/*---------------------------------------------------------------------------*/
/* Static function prototypes                                                */
/*---------------------------------------------------------------------------*/

// static void fixVarTree (MtrNode *treenode, int *perm, int size);

// static int addMultiplicityGroups (DdManager *dd, MtrNode *treenode, int multiplicity, char *vmask, char *lmask);


/**AutomaticEnd***************************************************************/


/*---------------------------------------------------------------------------*/
/* Definition of exported functions                                          */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Returns a new ADD variable.]

  Description [Creates a new ADD variable.  The new variable has an
  index equal to the largest previous index plus 1.  Returns a
  pointer to the new variable if successful; NULL otherwise.
  An ADD variable differs from a BDD variable because it points to the
  arithmetic zero, instead of having a complement pointer to 1. ]

  SideEffects [None]

  SeeAlso     [Cudd_bddNewVar Cudd_addIthVar Cudd_addConst
  Cudd_addNewVarAtLevel]

******************************************************************************/
func Cudd_addNewVar(dd *DdManager) *DdNode {
    var res *DdNode;

    if (unsigned int) dd.size >= CUDD_MAXINDEX - 1 { return(nil); }
    for {
        dd.reordered = 0;
        // inlining of macro DD_ONE
        // inlining of macro DD_ZERO
        res = cuddUniqueInter(dd,dd.size,((dd).one),((dd).zero));
    
if !(dd.reordered == 1) { break }
} 

    return(res);

} /* end of Cudd_addNewVar */


/**Function********************************************************************

  Synopsis    [Returns a new ADD variable at a specified level.]

  Description [Creates a new ADD variable.  The new variable has an
  index equal to the largest previous index plus 1 and is positioned at
  the specified level in the order.  Returns a pointer to the new
  variable if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addNewVar Cudd_addIthVar Cudd_bddNewVarAtLevel]

******************************************************************************/
func Cudd_addNewVarAtLevel(dd *DdManager, level int32) *DdNode {
    var res *DdNode;

    if (unsigned int) dd.size >= CUDD_MAXINDEX - 1 { return(nil); }
    if level >= dd.size { return(Cudd_addIthVar(dd,level)); }
    if !cuddInsertSubtables(dd,1,level) { return(nil); }
    for {
        dd.reordered = 0;
        // inlining of macro DD_ONE
        // inlining of macro DD_ZERO
        res = cuddUniqueInter(dd,dd.size - 1,((dd).one),((dd).zero));
    
if !(dd.reordered == 1) { break }
} 

    return(res);

} /* end of Cudd_addNewVarAtLevel */


/**Function********************************************************************

  Synopsis    [Returns a new BDD variable.]

  Description [Creates a new BDD variable.  The new variable has an
  index equal to the largest previous index plus 1.  Returns a
  pointer to the new variable if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addNewVar Cudd_bddIthVar Cudd_bddNewVarAtLevel]

******************************************************************************/
func Cudd_bddNewVar(dd *DdManager) *DdNode {
    var res *DdNode;

    if (unsigned int) dd.size >= CUDD_MAXINDEX - 1 { return(nil); }
    // inlining of macro Cudd_Not
    res = cuddUniqueInter(dd,dd.size,dd.one,((DdNode *)((ptrint)(dd.one) ^ 01)));

    return(res);

} /* end of Cudd_bddNewVar */


/**Function********************************************************************

  Synopsis    [Returns a new BDD variable at a specified level.]

  Description [Creates a new BDD variable.  The new variable has an
  index equal to the largest previous index plus 1 and is positioned at
  the specified level in the order.  Returns a pointer to the new
  variable if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddNewVar Cudd_bddIthVar Cudd_addNewVarAtLevel]

******************************************************************************/
func Cudd_bddNewVarAtLevel(dd *DdManager, level int32) *DdNode {
    var res *DdNode;

    if (unsigned int) dd.size >= CUDD_MAXINDEX - 1 { return(nil); }
    if level >= dd.size { return(Cudd_bddIthVar(dd,level)); }
    if !cuddInsertSubtables(dd,1,level) { return(nil); }
    res = dd.vars[dd.size - 1];

    return(res);

} /* end of Cudd_bddNewVarAtLevel */


/**Function********************************************************************

  Synopsis    [Returns the ADD variable with index i.]

  Description [Retrieves the ADD variable with index i if it already
  exists, or creates a new ADD variable.  Returns a pointer to the
  variable if successful; NULL otherwise.  An ADD variable differs from
  a BDD variable because it points to the arithmetic zero, instead of
  having a complement pointer to 1. ]

  SideEffects [None]

  SeeAlso     [Cudd_addNewVar Cudd_bddIthVar Cudd_addConst
  Cudd_addNewVarAtLevel]

******************************************************************************/
func Cudd_addIthVar(dd *DdManager, i int32) *DdNode {
    var res *DdNode;

    if (unsigned int) i >= CUDD_MAXINDEX - 1 { return(nil); }
    for {
        dd.reordered = 0;
        // inlining of macro DD_ONE
        // inlining of macro DD_ZERO
        res = cuddUniqueInter(dd,i,((dd).one),((dd).zero));
    
if !(dd.reordered == 1) { break }
} 

    return(res);

} /* end of Cudd_addIthVar */


/**Function********************************************************************

  Synopsis    [Returns the BDD variable with index i.]

  Description [Retrieves the BDD variable with index i if it already
  exists, or creates a new BDD variable.  Returns a pointer to the
  variable if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddNewVar Cudd_addIthVar Cudd_bddNewVarAtLevel
  Cudd_ReadVars]

******************************************************************************/
func Cudd_bddIthVar(dd *DdManager, i int32) *DdNode {
    var res *DdNode;

    if (unsigned int) i >= CUDD_MAXINDEX - 1 { return(nil); }
    if i < dd.size {
        res = dd.vars[i];
    } else {
        // inlining of macro Cudd_Not
        res = cuddUniqueInter(dd,i,dd.one,((DdNode *)((ptrint)(dd.one) ^ 01)));
    }

    return(res);

} /* end of Cudd_bddIthVar */


/**Function********************************************************************

  Synopsis    [Returns the ZDD variable with index i.]

  Description [Retrieves the ZDD variable with index i if it already
  exists, or creates a new ZDD variable.  Returns a pointer to the
  variable if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIthVar Cudd_addIthVar]

******************************************************************************/
func Cudd_zddIthVar(dd *DdManager, i int32) *DdNode {
    var res *DdNode;
    var zvar *DdNode;
    var lower *DdNode;
    var j int32;

    if (unsigned int) i >= CUDD_MAXINDEX - 1 { return(nil); }

    /* The i-th variable function has the following structure:
    ** at the level corresponding to index i there is a node whose "then"
    ** child points to the universe, and whose "else" child points to zero.
    ** Above that level there are nodes with identical children.
    */

    /* First we build the node at the level of index i. */
    if i < dd.sizeZ - 1 {
        lower = dd.univ[dd.permZ[i]+1];
    } else {
        // inlining of macro DD_ONE
        lower = ((dd).one);
    }
    for {
        dd.reordered = 0;
        // inlining of macro DD_ZERO
        zvar = cuddUniqueInterZdd(dd, i, lower, ((dd).zero));
    
if !(dd.reordered == 1) { break }
} 

    if zvar == nil {
        return(nil); }
    // inlining of macro cuddRef
    cuddSatInc(Cudd_Regular(zvar).ref);

    /* Now we add the "filler" nodes above the level of index i. */
    for j = dd.permZ[i] - 1; j >= 0; j-- {
        for {
            dd.reordered = 0;
            res = cuddUniqueInterZdd(dd, dd.invpermZ[j], zvar, zvar);
        
if !(dd.reordered == 1) { break }
} 
        if res == nil {
            Cudd_RecursiveDerefZdd(dd,zvar);
            return(nil);
        }
        // inlining of macro cuddRef
        cuddSatInc(Cudd_Regular(res).ref);
        Cudd_RecursiveDerefZdd(dd,zvar);
        zvar = res;
    }
    // inlining of macro cuddDeref
    cuddSatDec(Cudd_Regular(zvar).ref);
    return(zvar);

} /* end of Cudd_zddIthVar */


/**Function********************************************************************

  Synopsis    [Creates one or more ZDD variables for each BDD variable.]

  Description [Creates one or more ZDD variables for each BDD
  variable.  If some ZDD variables already exist, only the missing
  variables are created.  Parameter multiplicity allows the caller to
  control how many variables are created for each BDD variable in
  existence. For instance, if ZDDs are used to represent covers, two
  ZDD variables are required for each BDD variable.  The order of the
  BDD variables is transferred to the ZDD variables. If a variable
  group tree exists for the BDD variables, a corresponding ZDD
  variable group tree is created by expanding the BDD variable
  tree. In any case, the ZDD variables derived from the same BDD
  variable are merged in a ZDD variable group. If a ZDD variable group
  tree exists, it is freed. Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddNewVar Cudd_bddIthVar Cudd_bddNewVarAtLevel]

******************************************************************************/
int
Cudd_zddVarsFromBddVars(
  DdManager * dd /* DD manager */,
  int multiplicity /* how many ZDD variables are created for each BDD variable */)
{
    var res int32;
    var i, j int32;
    var allnew int32;
    var permutation *int32;

    if multiplicity < 1 { return(0); }
    allnew = dd.sizeZ == 0;
    if dd.size * multiplicity > dd.sizeZ {
        res = cuddResizeTableZdd(dd,dd.size * multiplicity - 1);
        if res == 0 { return(0); }
    }
    /* Impose the order of the BDD variables to the ZDD variables. */
    if allnew != 0 {
        for i = 0; i < dd.size; i++ {
            for j = 0; j < multiplicity; j++ {
                dd.permZ[i * multiplicity + j] =
                    dd.perm[i] * multiplicity + j;
                dd.invpermZ[dd.permZ[i * multiplicity + j]] =
                    i * multiplicity + j;
            }
        }
        for i = 0; i < dd.sizeZ; i++ {
            dd.univ[i].index = dd.invpermZ[i];
        }
    } else {
        permutation = ABC_ALLOC(int,dd.sizeZ);
        if permutation == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
        for i = 0; i < dd.size; i++ {
            for j = 0; j < multiplicity; j++ {
                permutation[i * multiplicity + j] =
                    dd.invperm[i] * multiplicity + j;
            }
        }
        for i = dd.size * multiplicity; i < dd.sizeZ; i++ {
            permutation[i] = i;
        }
        res = Cudd_zddShuffleHeap(dd, permutation);
        ABC_FREE(permutation);
        if res == 0 { return(0); }
    }
    /* Copy and expand the variable group tree if it exists. */
    if dd.treeZ != nil {
        Cudd_FreeZddTree(dd);
    }
    if dd.tree != nil {
        dd.treeZ = Mtr_CopyTree(dd.tree, multiplicity);
        if dd.treeZ == nil { return(0); }
    } else if multiplicity > 1 {
        dd.treeZ = Mtr_InitGroupTree(0, dd.sizeZ);
        if dd.treeZ == nil { return(0); }
        dd.treeZ.index = dd.invpermZ[0];
    }
    /* Create groups for the ZDD variables derived from the same BDD variable.
    */
    if multiplicity > 1 {
        var vmask, lmask string;

        vmask = ABC_ALLOC(char, dd.size);
        if vmask == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
        lmask =  ABC_ALLOC(char, dd.size);
        if lmask == nil {
            dd.errorCode = CUDD_MEMORY_OUT;
            return(0);
        }
        for i = 0; i < dd.size; i++ {
            vmask[i] = lmask[i] = 0;
        }
        res = addMultiplicityGroups(dd,dd.treeZ,multiplicity,vmask,lmask);
        ABC_FREE(vmask);
        ABC_FREE(lmask);
        if res == 0 { return(0); }
    }
    return(1);

} /* end of Cudd_zddVarsFromBddVars */


/**Function********************************************************************

  Synopsis    [Returns the ADD for constant c.]

  Description [Retrieves the ADD for constant c if it already
  exists, or creates a new ADD.  Returns a pointer to the
  ADD if successful; NULL otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_addNewVar Cudd_addIthVar]

******************************************************************************/
func Cudd_addConst(dd *DdManager, c CUDD_VALUE_TYPE) *DdNode {
    return(cuddUniqueConst(dd,c));

} /* end of Cudd_addConst */


/**Function********************************************************************

  Synopsis    [Returns 1 if a DD node is not constant.]

  Description [Returns 1 if a DD node is not constant. This function is
  useful to test the results of Cudd_bddIteConstant, Cudd_addIteConstant,
  Cudd_addEvalConst. These results may be a special value signifying
  non-constant. In the other cases the macro Cudd_IsConstant can be used.]

  SideEffects [None]

  SeeAlso     [Cudd_IsConstant Cudd_bddIteConstant Cudd_addIteConstant
  Cudd_addEvalConst]

******************************************************************************/
func Cudd_IsNonConstant(f *DdNode) int32 {
    // inlining of macro Cudd_IsConstant
    // inlining of macro Cudd_Regular
    return(f == DD_NON_CONSTANT || !((((DdNode *)((ptruint)(f) & ~01))).index == CUDD_CONST_INDEX));

} /* end of Cudd_IsNonConstant */


/**Function********************************************************************

  Synopsis    [Enables automatic dynamic reordering of BDDs and ADDs.]

  Description [Enables automatic dynamic reordering of BDDs and
  ADDs. Parameter method is used to determine the method used for
  reordering. If CUDD_REORDER_SAME is passed, the method is
  unchanged.]

  SideEffects [None]

  SeeAlso     [Cudd_AutodynDisable Cudd_ReorderingStatus
  Cudd_AutodynEnableZdd]

******************************************************************************/
func Cudd_AutodynEnable(unique *DdManager, method Cudd_ReorderingType) {
    unique.autoDyn = 1;
    if method != CUDD_REORDER_SAME {
        unique.autoMethod = method;
    }
#ifndef DD_NO_DEATH_ROW
    /* If reordering is enabled, using the death row causes too many
    ** invocations. Hence, we shrink the death row to just one entry.
    */
    cuddClearDeathRow(unique);
    unique->deathRowDepth = 1;
    unique->deadMask = unique->deathRowDepth - 1;
    if ((unsigned) unique->nextDead > unique->deadMask) {
        unique->nextDead = 0;
    }
    unique->deathRow = ABC_REALLOC(DdNodePtr, unique->deathRow,
        unique->deathRowDepth);
#endif
    return;

} /* end of Cudd_AutodynEnable */


/**Function********************************************************************

  Synopsis    [Disables automatic dynamic reordering.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_AutodynEnable Cudd_ReorderingStatus
  Cudd_AutodynDisableZdd]

******************************************************************************/
func Cudd_AutodynDisable(unique *DdManager) {
    unique.autoDyn = 0;
    return;

} /* end of Cudd_AutodynDisable */


/**Function********************************************************************

  Synopsis    [Reports the status of automatic dynamic reordering of BDDs
  and ADDs.]

  Description [Reports the status of automatic dynamic reordering of
  BDDs and ADDs. Parameter method is set to the reordering method
  currently selected. Returns 1 if automatic reordering is enabled; 0
  otherwise.]

  SideEffects [Parameter method is set to the reordering method currently
  selected.]

  SeeAlso     [Cudd_AutodynEnable Cudd_AutodynDisable
  Cudd_ReorderingStatusZdd]

******************************************************************************/
func Cudd_ReorderingStatus(unique *DdManager, method *Cudd_ReorderingType) int32 {
    *method = unique.autoMethod;
    return(unique.autoDyn);

} /* end of Cudd_ReorderingStatus */


/**Function********************************************************************

  Synopsis    [Enables automatic dynamic reordering of ZDDs.]

  Description [Enables automatic dynamic reordering of ZDDs. Parameter
  method is used to determine the method used for reordering ZDDs. If
  CUDD_REORDER_SAME is passed, the method is unchanged.]

  SideEffects [None]

  SeeAlso     [Cudd_AutodynDisableZdd Cudd_ReorderingStatusZdd
  Cudd_AutodynEnable]

******************************************************************************/
func Cudd_AutodynEnableZdd(unique *DdManager, method Cudd_ReorderingType) {
    unique.autoDynZ = 1;
    if method != CUDD_REORDER_SAME {
        unique.autoMethodZ = method;
    }
    return;

} /* end of Cudd_AutodynEnableZdd */


/**Function********************************************************************

  Synopsis    [Disables automatic dynamic reordering of ZDDs.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_AutodynEnableZdd Cudd_ReorderingStatusZdd
  Cudd_AutodynDisable]

******************************************************************************/
func Cudd_AutodynDisableZdd(unique *DdManager) {
    unique.autoDynZ = 0;
    return;

} /* end of Cudd_AutodynDisableZdd */


/**Function********************************************************************

  Synopsis    [Reports the status of automatic dynamic reordering of ZDDs.]

  Description [Reports the status of automatic dynamic reordering of
  ZDDs. Parameter method is set to the ZDD reordering method currently
  selected. Returns 1 if automatic reordering is enabled; 0
  otherwise.]

  SideEffects [Parameter method is set to the ZDD reordering method currently
  selected.]

  SeeAlso     [Cudd_AutodynEnableZdd Cudd_AutodynDisableZdd
  Cudd_ReorderingStatus]

******************************************************************************/
func Cudd_ReorderingStatusZdd(unique *DdManager, method *Cudd_ReorderingType) int32 {
    *method = unique.autoMethodZ;
    return(unique.autoDynZ);

} /* end of Cudd_ReorderingStatusZdd */


/**Function********************************************************************

  Synopsis    [Tells whether the realignment of ZDD order to BDD order is
  enabled.]

  Description [Returns 1 if the realignment of ZDD order to BDD order is
  enabled; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_zddRealignEnable Cudd_zddRealignDisable
  Cudd_bddRealignEnable Cudd_bddRealignDisable]

******************************************************************************/
func Cudd_zddRealignmentEnabled(unique *DdManager) int32 {
    return(unique.realign);

} /* end of Cudd_zddRealignmentEnabled */


/**Function********************************************************************

  Synopsis    [Enables realignment of ZDD order to BDD order.]

  Description [Enables realignment of the ZDD variable order to the
  BDD variable order after the BDDs and ADDs have been reordered.  The
  number of ZDD variables must be a multiple of the number of BDD
  variables for realignment to make sense. If this condition is not met,
  Cudd_ReduceHeap will return 0. Let <code>M</code> be the
  ratio of the two numbers. For the purpose of realignment, the ZDD
  variables from <code>M*i</code> to <code>(M+1)*i-1</code> are
  reagarded as corresponding to BDD variable <code>i</code>. Realignment
  is initially disabled.]

  SideEffects [None]

  SeeAlso     [Cudd_ReduceHeap Cudd_zddRealignDisable
  Cudd_zddRealignmentEnabled Cudd_bddRealignDisable
  Cudd_bddRealignmentEnabled]

******************************************************************************/
func Cudd_zddRealignEnable(unique *DdManager) {
    unique.realign = 1;
    return;

} /* end of Cudd_zddRealignEnable */


/**Function********************************************************************

  Synopsis    [Disables realignment of ZDD order to BDD order.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_zddRealignEnable Cudd_zddRealignmentEnabled
  Cudd_bddRealignEnable Cudd_bddRealignmentEnabled]

******************************************************************************/
func Cudd_zddRealignDisable(unique *DdManager) {
    unique.realign = 0;
    return;

} /* end of Cudd_zddRealignDisable */


/**Function********************************************************************

  Synopsis    [Tells whether the realignment of BDD order to ZDD order is
  enabled.]

  Description [Returns 1 if the realignment of BDD order to ZDD order is
  enabled; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_bddRealignEnable Cudd_bddRealignDisable
  Cudd_zddRealignEnable Cudd_zddRealignDisable]

******************************************************************************/
func Cudd_bddRealignmentEnabled(unique *DdManager) int32 {
    return(unique.realignZ);

} /* end of Cudd_bddRealignmentEnabled */


/**Function********************************************************************

  Synopsis    [Enables realignment of BDD order to ZDD order.]

  Description [Enables realignment of the BDD variable order to the
  ZDD variable order after the ZDDs have been reordered.  The
  number of ZDD variables must be a multiple of the number of BDD
  variables for realignment to make sense. If this condition is not met,
  Cudd_zddReduceHeap will return 0. Let <code>M</code> be the
  ratio of the two numbers. For the purpose of realignment, the ZDD
  variables from <code>M*i</code> to <code>(M+1)*i-1</code> are
  reagarded as corresponding to BDD variable <code>i</code>. Realignment
  is initially disabled.]

  SideEffects [None]

  SeeAlso     [Cudd_zddReduceHeap Cudd_bddRealignDisable
  Cudd_bddRealignmentEnabled Cudd_zddRealignDisable
  Cudd_zddRealignmentEnabled]

******************************************************************************/
func Cudd_bddRealignEnable(unique *DdManager) {
    unique.realignZ = 1;
    return;

} /* end of Cudd_bddRealignEnable */


/**Function********************************************************************

  Synopsis    [Disables realignment of ZDD order to BDD order.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_bddRealignEnable Cudd_bddRealignmentEnabled
  Cudd_zddRealignEnable Cudd_zddRealignmentEnabled]

******************************************************************************/
func Cudd_bddRealignDisable(unique *DdManager) {
    unique.realignZ = 0;
    return;

} /* end of Cudd_bddRealignDisable */


/**Function********************************************************************

  Synopsis    [Returns the one constant of the manager.]

  Description [Returns the one constant of the manager. The one
  constant is common to ADDs and BDDs.]

  SideEffects [None]

  SeeAlso [Cudd_ReadZero Cudd_ReadLogicZero Cudd_ReadZddOne]

******************************************************************************/
func Cudd_ReadOne(dd *DdManager) *DdNode {
    return(dd.one);

} /* end of Cudd_ReadOne */


/**Function********************************************************************

  Synopsis    [Returns the ZDD for the constant 1 function.]

  Description [Returns the ZDD for the constant 1 function.
  The representation of the constant 1 function as a ZDD depends on
  how many variables it (nominally) depends on. The index of the
  topmost variable in the support is given as argument <code>i</code>.]

  SideEffects [None]

  SeeAlso [Cudd_ReadOne]

******************************************************************************/
func Cudd_ReadZddOne(dd *DdManager, i int32) *DdNode {
    if i < 0 {
        return(nil); }
    if i < dd.sizeZ {
        return dd.univ[i];
    } else {
        // inlining of macro DD_ONE
        return ((dd).one);
    }

} /* end of Cudd_ReadZddOne */



/**Function********************************************************************

  Synopsis    [Returns the zero constant of the manager.]

  Description [Returns the zero constant of the manager. The zero
  constant is the arithmetic zero, rather than the logic zero. The
  latter is the complement of the one constant.]

  SideEffects [None]

  SeeAlso [Cudd_ReadOne Cudd_ReadLogicZero]

******************************************************************************/
func Cudd_ReadZero(dd *DdManager) *DdNode {
    // inlining of macro DD_ZERO
    return(((dd).zero));

} /* end of Cudd_ReadZero */


/**Function********************************************************************

  Synopsis    [Returns the logic zero constant of the manager.]

  Description [Returns the zero constant of the manager. The logic zero
  constant is the complement of the one constant, and is distinct from
  the arithmetic zero.]

  SideEffects [None]

  SeeAlso [Cudd_ReadOne Cudd_ReadZero]

******************************************************************************/
func Cudd_ReadLogicZero(dd *DdManager) *DdNode {
    // inlining of macro Cudd_Not
    // inlining of macro DD_ONE
    return(((DdNode *)((ptrint)(((dd).one)) ^ 01)));

} /* end of Cudd_ReadLogicZero */


/**Function********************************************************************

  Synopsis    [Reads the plus-infinity constant from the manager.]

  Description []

  SideEffects [None]

******************************************************************************/
func Cudd_ReadPlusInfinity(dd *DdManager) *DdNode {
    return(dd.plusinfinity);

} /* end of Cudd_ReadPlusInfinity */


/**Function********************************************************************

  Synopsis    [Reads the minus-infinity constant from the manager.]

  Description []

  SideEffects [None]

******************************************************************************/
func Cudd_ReadMinusInfinity(dd *DdManager) *DdNode {
    return(dd.minusinfinity);

} /* end of Cudd_ReadMinusInfinity */


/**Function********************************************************************

  Synopsis    [Reads the background constant of the manager.]

  Description []

  SideEffects [None]

******************************************************************************/
func Cudd_ReadBackground(dd *DdManager) *DdNode {
    return(dd.background);

} /* end of Cudd_ReadBackground */


/**Function********************************************************************

  Synopsis    [Sets the background constant of the manager.]

  Description [Sets the background constant of the manager. It assumes
  that the DdNode pointer bck is already referenced.]

  SideEffects [None]

******************************************************************************/
func Cudd_SetBackground(dd *DdManager, bck *DdNode) {
    dd.background = bck;

} /* end of Cudd_SetBackground */


/**Function********************************************************************

  Synopsis    [Reads the number of slots in the cache.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_ReadCacheUsedSlots]

******************************************************************************/
func Cudd_ReadCacheSlots(dd *DdManager) unsigned int {
    return(dd.cacheSlots);

} /* end of Cudd_ReadCacheSlots */


/**Function********************************************************************

  Synopsis    [Reads the fraction of used slots in the cache.]

  Description [Reads the fraction of used slots in the cache. The unused
  slots are those in which no valid data is stored. Garbage collection,
  variable reordering, and cache resizing may cause used slots to become
  unused.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadCacheSlots]

******************************************************************************/
func Cudd_ReadCacheUsedSlots(dd *DdManager) float64 {
    unsigned long used = 0;
    var slots int32  = dd.cacheSlots;
    DdCache *cache = dd.cache;
    var i int32;

    for i = 0; i < slots; i++ {
        used += cache[i].h != 0;
    }

    return(float64(used) / float64( dd.cacheSlots));

} /* end of Cudd_ReadCacheUsedSlots */


/**Function********************************************************************

  Synopsis    [Returns the number of cache look-ups.]

  Description [Returns the number of cache look-ups.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadCacheHits]

******************************************************************************/
func Cudd_ReadCacheLookUps(dd *DdManager) float64 {
    return(dd.cacheHits + dd.cacheMisses +
           dd.totCachehits + dd.totCacheMisses);

} /* end of Cudd_ReadCacheLookUps */


/**Function********************************************************************

  Synopsis    [Returns the number of cache hits.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_ReadCacheLookUps]

******************************************************************************/
func Cudd_ReadCacheHits(dd *DdManager) float64 {
    return(dd.cacheHits + dd.totCachehits);

} /* end of Cudd_ReadCacheHits */


/**Function********************************************************************

  Synopsis    [Returns the number of recursive calls.]

  Description [Returns the number of recursive calls if the package is
  compiled with DD_COUNT defined.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_ReadRecursiveCalls(dd *DdManager) float64 {
#ifdef DD_COUNT
    return(dd->recursiveCalls);
#else
    return(-1.0);
#endif

} /* end of Cudd_ReadRecursiveCalls */



/**Function********************************************************************

  Synopsis    [Reads the hit rate that causes resizinig of the computed
  table.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_SetMinHit]

******************************************************************************/
func Cudd_ReadMinHit(dd *DdManager) unsigned int {
    /* Internally, the package manipulates the ratio of hits to
    ** misses instead of the ratio of hits to accesses. */
    return((unsigned int) (0.5 + 100 * dd.minHit / (1 + dd.minHit)));

} /* end of Cudd_ReadMinHit */


/**Function********************************************************************

  Synopsis    [Sets the hit rate that causes resizinig of the computed
  table.]

  Description [Sets the minHit parameter of the manager. This
  parameter controls the resizing of the computed table. If the hit
  rate is larger than the specified value, and the cache is not
  already too large, then its size is doubled.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadMinHit]

******************************************************************************/
void
Cudd_SetMinHit(
  DdManager * dd,
  unsigned int hr)
{
    /* Internally, the package manipulates the ratio of hits to
    ** misses instead of the ratio of hits to accesses. */
    dd.minHit = float64( hr) / (100.0 - float64( hr));

} /* end of Cudd_SetMinHit */


/**Function********************************************************************

  Synopsis    [Reads the looseUpTo parameter of the manager.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_SetLooseUpTo Cudd_ReadMinHit Cudd_ReadMinDead]

******************************************************************************/
func Cudd_ReadLooseUpTo(dd *DdManager) unsigned int {
    return(dd.looseUpTo);

} /* end of Cudd_ReadLooseUpTo */


/**Function********************************************************************

  Synopsis    [Sets the looseUpTo parameter of the manager.]

  Description [Sets the looseUpTo parameter of the manager. This
  parameter of the manager controls the threshold beyond which no fast
  growth of the unique table is allowed. The threshold is given as a
  number of slots. If the value passed to this function is 0, the
  function determines a suitable value based on the available memory.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadLooseUpTo Cudd_SetMinHit]

******************************************************************************/
void
Cudd_SetLooseUpTo(
  DdManager * dd,
  unsigned int lut)
{
    if lut == 0 {
        unsigned long datalimit = getSoftDataLimit();
        lut = (unsigned int) (datalimit / (sizeof(DdNode) *
                                           DD_MAX_LOOSE_FRACTION));
    }
    dd.looseUpTo = lut;

} /* end of Cudd_SetLooseUpTo */


/**Function********************************************************************

  Synopsis    [Returns the soft limit for the cache size.]

  Description [Returns the soft limit for the cache size. The soft limit]

  SideEffects [None]

  SeeAlso     [Cudd_ReadMaxCache]

******************************************************************************/
func Cudd_ReadMaxCache(dd *DdManager) unsigned int {
    return(2 * dd.cacheSlots + dd.cacheSlack);

} /* end of Cudd_ReadMaxCache */


/**Function********************************************************************

  Synopsis    [Reads the maxCacheHard parameter of the manager.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_SetMaxCacheHard Cudd_ReadMaxCache]

******************************************************************************/
func Cudd_ReadMaxCacheHard(dd *DdManager) unsigned int {
    return(dd.maxCacheHard);

} /* end of Cudd_ReadMaxCache */


/**Function********************************************************************

  Synopsis    [Sets the maxCacheHard parameter of the manager.]

  Description [Sets the maxCacheHard parameter of the manager. The
  cache cannot grow larger than maxCacheHard entries. This parameter
  allows an application to control the trade-off of memory versus
  speed. If the value passed to this function is 0, the function
  determines a suitable maximum cache size based on the available memory.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadMaxCacheHard Cudd_SetMaxCache]

******************************************************************************/
void
Cudd_SetMaxCacheHard(
  DdManager * dd,
  unsigned int mc)
{
    if mc == 0 {
        unsigned long datalimit = getSoftDataLimit();
        mc = (unsigned int) (datalimit / (sizeof(DdCache) *
                                          DD_MAX_CACHE_FRACTION));
    }
    dd.maxCacheHard = mc;

} /* end of Cudd_SetMaxCacheHard */


/**Function********************************************************************

  Synopsis    [Returns the number of BDD variables in existance.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_ReadZddSize]

******************************************************************************/
func Cudd_ReadSize(dd *DdManager) int32 {
    return(dd.size);

} /* end of Cudd_ReadSize */


/**Function********************************************************************

  Synopsis    [Returns the number of ZDD variables in existance.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_ReadSize]

******************************************************************************/
func Cudd_ReadZddSize(dd *DdManager) int32 {
    return(dd.sizeZ);

} /* end of Cudd_ReadZddSize */


/**Function********************************************************************

  Synopsis    [Returns the total number of slots of the unique table.]

  Description [Returns the total number of slots of the unique table.
  This number ismainly for diagnostic purposes.]

  SideEffects [None]

******************************************************************************/
func Cudd_ReadSlots(dd *DdManager) unsigned int {
    return(dd.slots);

} /* end of Cudd_ReadSlots */


/**Function********************************************************************

  Synopsis    [Reads the fraction of used slots in the unique table.]

  Description [Reads the fraction of used slots in the unique
  table. The unused slots are those in which no valid data is
  stored. Garbage collection, variable reordering, and subtable
  resizing may cause used slots to become unused.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadSlots]

******************************************************************************/
func Cudd_ReadUsedSlots(dd *DdManager) float64 {
    unsigned long used = 0;
    var i, j int32;
    var size int32  = dd.size;
    var nodelist *DdNodePtr;
    var subtable *DdSubtable;
    var node *DdNode;
    var sentinel *DdNode  = &(dd.sentinel);

    /* Scan each BDD/ADD subtable. */
    for i = 0; i < size; i++ {
        subtable = &(dd.subtables[i]);
        nodelist = subtable.nodelist;
        for j = 0; uint32( j) < subtable.slots; j++ {
            node = nodelist[j];
            if node != sentinel {
                used++;
            }
        }
    }

    /* Scan the ZDD subtables. */
    size = dd.sizeZ;

    for i = 0; i < size; i++ {
        subtable = &(dd.subtableZ[i]);
        nodelist = subtable.nodelist;
        for j = 0; uint32( j) < subtable.slots; j++ {
            node = nodelist[j];
            if node != nil {
                used++;
            }
        }
    }

    /* Constant table. */
    subtable = &(dd.constants);
    nodelist = subtable.nodelist;
    for j = 0; uint32( j) < subtable.slots; j++ {
        node = nodelist[j];
        if node != nil {
            used++;
        }
    }

    return(float64(used) / float64( dd.slots));

} /* end of Cudd_ReadUsedSlots */


/**Function********************************************************************

  Synopsis    [Computes the expected fraction of used slots in the unique
  table.]

  Description [Computes the fraction of slots in the unique table that
  should be in use. This expected value is based on the assumption
  that the hash function distributes the keys randomly; it can be
  compared with the result of Cudd_ReadUsedSlots to monitor the
  performance of the unique table hash function.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadSlots Cudd_ReadUsedSlots]

******************************************************************************/
func Cudd_ExpectedUsedSlots(dd *DdManager) float64 {
    var i int32;
    var size int32  = dd.size;
    var subtable *DdSubtable;
    var empty float64  = 0.0;

    /* To each subtable we apply the corollary to Theorem 8.5 (occupancy
    ** distribution) from Sedgewick and Flajolet's Analysis of Algorithms.
    ** The corollary says that for a table with M buckets and a load ratio
    ** of r, the expected number of empty buckets is asymptotically given
    ** by M * exp(-r).
    */

    /* Scan each BDD/ADD subtable. */
    for i = 0; i < size; i++ {
        subtable = &(dd.subtables[i]);
        empty += float64( subtable.slots) *
            math.Exp(float64(-float64( subtable.keys) / float64( subtable.slots)));
    }

    /* Scan the ZDD subtables. */
    size = dd.sizeZ;

    for i = 0; i < size; i++ {
        subtable = &(dd.subtableZ[i]);
        empty += float64( subtable.slots) *
            math.Exp(float64(-float64( subtable.keys) / float64( subtable.slots)));
    }

    /* Constant table. */
    subtable = &(dd.constants);
    empty += float64( subtable.slots) *
        math.Exp(float64(-float64( subtable.keys) / float64( subtable.slots)));

    return(1.0 - empty / float64( dd.slots));

} /* end of Cudd_ExpectedUsedSlots */


/**Function********************************************************************

  Synopsis    [Returns the number of nodes in the unique table.]

  Description [Returns the total number of nodes currently in the unique
  table, including the dead nodes.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadDead]

******************************************************************************/
func Cudd_ReadKeys(dd *DdManager) unsigned int {
    return(dd.keys);

} /* end of Cudd_ReadKeys */


/**Function********************************************************************

  Synopsis    [Returns the number of dead nodes in the unique table.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_ReadKeys]

******************************************************************************/
func Cudd_ReadDead(dd *DdManager) unsigned int {
    return(dd.dead);

} /* end of Cudd_ReadDead */


/**Function********************************************************************

  Synopsis    [Reads the minDead parameter of the manager.]

  Description [Reads the minDead parameter of the manager. The minDead
  parameter is used by the package to decide whether to collect garbage
  or resize a subtable of the unique table when the subtable becomes
  too full. The application can indirectly control the value of minDead
  by setting the looseUpTo parameter.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadDead Cudd_ReadLooseUpTo Cudd_SetLooseUpTo]

******************************************************************************/
func Cudd_ReadMinDead(dd *DdManager) unsigned int {
    return(dd.minDead);

} /* end of Cudd_ReadMinDead */


/**Function********************************************************************

  Synopsis    [Returns the number of times reordering has occurred.]

  Description [Returns the number of times reordering has occurred in the
  manager. The number includes both the calls to Cudd_ReduceHeap from
  the application program and those automatically performed by the
  package. However, calls that do not even initiate reordering are not
  counted. A call may not initiate reordering if there are fewer than
  minsize live nodes in the manager, or if CUDD_REORDER_NONE is specified
  as reordering method. The calls to Cudd_ShuffleHeap are not counted.]

  SideEffects [None]

  SeeAlso [Cudd_ReduceHeap Cudd_ReadReorderingTime]

******************************************************************************/
func Cudd_ReadReorderings(dd *DdManager) int32 {
    return(dd.reorderings);

} /* end of Cudd_ReadReorderings */


/**Function********************************************************************

  Synopsis    [Returns the time spent in reordering.]

  Description [Returns the number of milliseconds spent reordering
  variables since the manager was initialized. The time spent in collecting
  garbage before reordering is included.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadReorderings]

******************************************************************************/
func Cudd_ReadReorderingTime(dd *DdManager) long {
    return(dd.reordTime);

} /* end of Cudd_ReadReorderingTime */


/**Function********************************************************************

  Synopsis    [Returns the number of times garbage collection has occurred.]

  Description [Returns the number of times garbage collection has
  occurred in the manager. The number includes both the calls from
  reordering procedures and those caused by requests to create new
  nodes.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadGarbageCollectionTime]

******************************************************************************/
func Cudd_ReadGarbageCollections(dd *DdManager) int32 {
    return(dd.garbageCollections);

} /* end of Cudd_ReadGarbageCollections */


/**Function********************************************************************

  Synopsis    [Returns the time spent in garbage collection.]

  Description [Returns the number of milliseconds spent doing garbage
  collection since the manager was initialized.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadGarbageCollections]

******************************************************************************/
func Cudd_ReadGarbageCollectionTime(dd *DdManager) long {
    return(dd.GCTime);

} /* end of Cudd_ReadGarbageCollectionTime */


/**Function********************************************************************

  Synopsis    [Returns the number of nodes freed.]

  Description [Returns the number of nodes returned to the free list if the
  keeping of this statistic is enabled; -1 otherwise. This statistic is
  enabled only if the package is compiled with DD_STATS defined.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadNodesDropped]

******************************************************************************/
func Cudd_ReadNodesFreed(dd *DdManager) float64 {
#ifdef DD_STATS
    return(dd->nodesFreed);
#else
    return(-1.0);
#endif

} /* end of Cudd_ReadNodesFreed */


/**Function********************************************************************

  Synopsis    [Returns the number of nodes dropped.]

  Description [Returns the number of nodes killed by dereferencing if the
  keeping of this statistic is enabled; -1 otherwise. This statistic is
  enabled only if the package is compiled with DD_STATS defined.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadNodesFreed]

******************************************************************************/
func Cudd_ReadNodesDropped(dd *DdManager) float64 {
#ifdef DD_STATS
    return(dd->nodesDropped);
#else
    return(-1.0);
#endif

} /* end of Cudd_ReadNodesDropped */


/**Function********************************************************************

  Synopsis    [Returns the number of look-ups in the unique table.]

  Description [Returns the number of look-ups in the unique table if the
  keeping of this statistic is enabled; -1 otherwise. This statistic is
  enabled only if the package is compiled with DD_UNIQUE_PROFILE defined.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadUniqueLinks]

******************************************************************************/
func Cudd_ReadUniqueLookUps(dd *DdManager) float64 {
#ifdef DD_UNIQUE_PROFILE
    return(dd->uniqueLookUps);
#else
    return(-1.0);
#endif

} /* end of Cudd_ReadUniqueLookUps */


/**Function********************************************************************

  Synopsis    [Returns the number of links followed in the unique table.]

  Description [Returns the number of links followed during look-ups in the
  unique table if the keeping of this statistic is enabled; -1 otherwise.
  If an item is found in the first position of its collision list, the
  number of links followed is taken to be 0. If it is in second position,
  the number of links is 1, and so on. This statistic is enabled only if
  the package is compiled with DD_UNIQUE_PROFILE defined.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadUniqueLookUps]

******************************************************************************/
func Cudd_ReadUniqueLinks(dd *DdManager) float64 {
#ifdef DD_UNIQUE_PROFILE
    return(dd->uniqueLinks);
#else
    return(-1.0);
#endif

} /* end of Cudd_ReadUniqueLinks */


/**Function********************************************************************

  Synopsis    [Reads the siftMaxVar parameter of the manager.]

  Description [Reads the siftMaxVar parameter of the manager. This
  parameter gives the maximum number of variables that will be sifted
  for each invocation of sifting.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadSiftMaxSwap Cudd_SetSiftMaxVar]

******************************************************************************/
func Cudd_ReadSiftMaxVar(dd *DdManager) int32 {
    return(dd.siftMaxVar);

} /* end of Cudd_ReadSiftMaxVar */


/**Function********************************************************************

  Synopsis    [Sets the siftMaxVar parameter of the manager.]

  Description [Sets the siftMaxVar parameter of the manager. This
  parameter gives the maximum number of variables that will be sifted
  for each invocation of sifting.]

  SideEffects [None]

  SeeAlso     [Cudd_SetSiftMaxSwap Cudd_ReadSiftMaxVar]

******************************************************************************/
func Cudd_SetSiftMaxVar(dd *DdManager, smv int32) {
    dd.siftMaxVar = smv;

} /* end of Cudd_SetSiftMaxVar */


/**Function********************************************************************

  Synopsis    [Reads the siftMaxSwap parameter of the manager.]

  Description [Reads the siftMaxSwap parameter of the manager. This
  parameter gives the maximum number of swaps that will be attempted
  for each invocation of sifting. The real number of swaps may exceed
  the set limit because the package will always complete the sifting
  of the variable that causes the limit to be reached.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadSiftMaxVar Cudd_SetSiftMaxSwap]

******************************************************************************/
func Cudd_ReadSiftMaxSwap(dd *DdManager) int32 {
    return(dd.siftMaxSwap);

} /* end of Cudd_ReadSiftMaxSwap */


/**Function********************************************************************

  Synopsis    [Sets the siftMaxSwap parameter of the manager.]

  Description [Sets the siftMaxSwap parameter of the manager. This
  parameter gives the maximum number of swaps that will be attempted
  for each invocation of sifting. The real number of swaps may exceed
  the set limit because the package will always complete the sifting
  of the variable that causes the limit to be reached.]

  SideEffects [None]

  SeeAlso     [Cudd_SetSiftMaxVar Cudd_ReadSiftMaxSwap]

******************************************************************************/
func Cudd_SetSiftMaxSwap(dd *DdManager, sms int32) {
    dd.siftMaxSwap = sms;

} /* end of Cudd_SetSiftMaxSwap */


/**Function********************************************************************

  Synopsis    [Reads the maxGrowth parameter of the manager.]

  Description [Reads the maxGrowth parameter of the manager.  This
  parameter determines how much the number of nodes can grow during
  sifting of a variable.  Overall, sifting never increases the size of
  the decision diagrams.  This parameter only refers to intermediate
  results.  A lower value will speed up sifting, possibly at the
  expense of quality.]

  SideEffects [None]

  SeeAlso     [Cudd_SetMaxGrowth Cudd_ReadMaxGrowthAlternate]

******************************************************************************/
func Cudd_ReadMaxGrowth(dd *DdManager) float64 {
    return(dd.maxGrowth);

} /* end of Cudd_ReadMaxGrowth */


/**Function********************************************************************

  Synopsis    [Sets the maxGrowth parameter of the manager.]

  Description [Sets the maxGrowth parameter of the manager.  This
  parameter determines how much the number of nodes can grow during
  sifting of a variable.  Overall, sifting never increases the size of
  the decision diagrams.  This parameter only refers to intermediate
  results.  A lower value will speed up sifting, possibly at the
  expense of quality.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadMaxGrowth Cudd_SetMaxGrowthAlternate]

******************************************************************************/
func Cudd_SetMaxGrowth(dd *DdManager, mg float64) {
    dd.maxGrowth = mg;

} /* end of Cudd_SetMaxGrowth */


/**Function********************************************************************

  Synopsis    [Reads the maxGrowthAlt parameter of the manager.]

  Description [Reads the maxGrowthAlt parameter of the manager.  This
  parameter is analogous to the maxGrowth paramter, and is used every
  given number of reorderings instead of maxGrowth.  The number of
  reorderings is set with Cudd_SetReorderingCycle.  If the number of
  reorderings is 0 (default) maxGrowthAlt is never used.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadMaxGrowth Cudd_SetMaxGrowthAlternate
  Cudd_SetReorderingCycle Cudd_ReadReorderingCycle]

******************************************************************************/
func Cudd_ReadMaxGrowthAlternate(dd *DdManager) float64 {
    return(dd.maxGrowthAlt);

} /* end of Cudd_ReadMaxGrowthAlternate */


/**Function********************************************************************

  Synopsis    [Sets the maxGrowthAlt parameter of the manager.]

  Description [Sets the maxGrowthAlt parameter of the manager.  This
  parameter is analogous to the maxGrowth paramter, and is used every
  given number of reorderings instead of maxGrowth.  The number of
  reorderings is set with Cudd_SetReorderingCycle.  If the number of
  reorderings is 0 (default) maxGrowthAlt is never used.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadMaxGrowthAlternate Cudd_SetMaxGrowth
  Cudd_SetReorderingCycle Cudd_ReadReorderingCycle]

******************************************************************************/
func Cudd_SetMaxGrowthAlternate(dd *DdManager, mg float64) {
    dd.maxGrowthAlt = mg;

} /* end of Cudd_SetMaxGrowthAlternate */


/**Function********************************************************************

  Synopsis    [Reads the reordCycle parameter of the manager.]

  Description [Reads the reordCycle parameter of the manager.  This
  parameter determines how often the alternate threshold on maximum
  growth is used in reordering.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadMaxGrowthAlternate Cudd_SetMaxGrowthAlternate
  Cudd_SetReorderingCycle]

******************************************************************************/
func Cudd_ReadReorderingCycle(dd *DdManager) int32 {
    return(dd.reordCycle);

} /* end of Cudd_ReadReorderingCycle */


/**Function********************************************************************

  Synopsis    [Sets the reordCycle parameter of the manager.]

  Description [Sets the reordCycle parameter of the manager.  This
  parameter determines how often the alternate threshold on maximum
  growth is used in reordering.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadMaxGrowthAlternate Cudd_SetMaxGrowthAlternate
  Cudd_ReadReorderingCycle]

******************************************************************************/
func Cudd_SetReorderingCycle(dd *DdManager, cycle int32) {
    dd.reordCycle = cycle;

} /* end of Cudd_SetReorderingCycle */


/**Function********************************************************************

  Synopsis    [Returns the variable group tree of the manager.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_SetTree Cudd_FreeTree Cudd_ReadZddTree]

******************************************************************************/
func Cudd_ReadTree(dd *DdManager) *MtrNode {
    return(dd.tree);

} /* end of Cudd_ReadTree */


/**Function********************************************************************

  Synopsis    [Sets the variable group tree of the manager.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_FreeTree Cudd_ReadTree Cudd_SetZddTree]

******************************************************************************/
func Cudd_SetTree(dd *DdManager, tree *MtrNode) {
    if dd.tree != nil {
        Mtr_FreeTree(dd.tree);
    }
    dd.tree = tree;
    if tree == nil { return; }

    fixVarTree(tree, dd.perm, dd.size);
    return;

} /* end of Cudd_SetTree */


/**Function********************************************************************

  Synopsis    [Frees the variable group tree of the manager.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_SetTree Cudd_ReadTree Cudd_FreeZddTree]

******************************************************************************/
func Cudd_FreeTree(dd *DdManager) {
    if dd.tree != nil {
        Mtr_FreeTree(dd.tree);
        dd.tree = nil;
    }
    return;

} /* end of Cudd_FreeTree */


/**Function********************************************************************

  Synopsis    [Returns the variable group tree of the manager.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_SetZddTree Cudd_FreeZddTree Cudd_ReadTree]

******************************************************************************/
func Cudd_ReadZddTree(dd *DdManager) *MtrNode {
    return(dd.treeZ);

} /* end of Cudd_ReadZddTree */


/**Function********************************************************************

  Synopsis    [Sets the ZDD variable group tree of the manager.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_FreeZddTree Cudd_ReadZddTree Cudd_SetTree]

******************************************************************************/
func Cudd_SetZddTree(dd *DdManager, tree *MtrNode) {
    if dd.treeZ != nil {
        Mtr_FreeTree(dd.treeZ);
    }
    dd.treeZ = tree;
    if tree == nil { return; }

    fixVarTree(tree, dd.permZ, dd.sizeZ);
    return;

} /* end of Cudd_SetZddTree */


/**Function********************************************************************

  Synopsis    [Frees the variable group tree of the manager.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_SetZddTree Cudd_ReadZddTree Cudd_FreeTree]

******************************************************************************/
func Cudd_FreeZddTree(dd *DdManager) {
    if dd.treeZ != nil {
        Mtr_FreeTree(dd.treeZ);
        dd.treeZ = nil;
    }
    return;

} /* end of Cudd_FreeZddTree */


/**Function********************************************************************

  Synopsis    [Returns the index of the node.]

  Description [Returns the index of the node. The node pointer can be
  either regular or complemented.]

  SideEffects [None]

  SeeAlso [Cudd_ReadIndex]

******************************************************************************/
func Cudd_NodeReadIndex(node *DdNode) unsigned int {
    // inlining of macro Cudd_Regular
    return((unsigned int) ((DdNode *)((ptruint)(node) & ~01)).index);

} /* end of Cudd_NodeReadIndex */


/**Function********************************************************************

  Synopsis    [Returns the current position of the i-th variable in the
  order.]

  Description [Returns the current position of the i-th variable in
  the order. If the index is CUDD_CONST_INDEX, returns
  CUDD_CONST_INDEX; otherwise, if the index is out of bounds returns
  -1.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadInvPerm Cudd_ReadPermZdd]

******************************************************************************/
func Cudd_ReadPerm(dd *DdManager, i int32) int32 {
    if i == CUDD_CONST_INDEX { return(CUDD_CONST_INDEX); }
    if i < 0 || i >= dd.size { return(-1); }
    return(dd.perm[i]);

} /* end of Cudd_ReadPerm */


/**Function********************************************************************

  Synopsis    [Returns the current position of the i-th ZDD variable in the
  order.]

  Description [Returns the current position of the i-th ZDD variable
  in the order. If the index is CUDD_CONST_INDEX, returns
  CUDD_CONST_INDEX; otherwise, if the index is out of bounds returns
  -1.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadInvPermZdd Cudd_ReadPerm]

******************************************************************************/
func Cudd_ReadPermZdd(dd *DdManager, i int32) int32 {
    if i == CUDD_CONST_INDEX { return(CUDD_CONST_INDEX); }
    if i < 0 || i >= dd.sizeZ { return(-1); }
    return(dd.permZ[i]);

} /* end of Cudd_ReadPermZdd */


/**Function********************************************************************

  Synopsis    [Returns the index of the variable currently in the i-th
  position of the order.]

  Description [Returns the index of the variable currently in the i-th
  position of the order. If the index is CUDD_CONST_INDEX, returns
  CUDD_CONST_INDEX; otherwise, if the index is out of bounds returns -1.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadPerm Cudd_ReadInvPermZdd]

******************************************************************************/
func Cudd_ReadInvPerm(dd *DdManager, i int32) int32 {
    if i == CUDD_CONST_INDEX { return(CUDD_CONST_INDEX); }
    if i < 0 || i >= dd.size { return(-1); }
    return(dd.invperm[i]);

} /* end of Cudd_ReadInvPerm */


/**Function********************************************************************

  Synopsis    [Returns the index of the ZDD variable currently in the i-th
  position of the order.]

  Description [Returns the index of the ZDD variable currently in the
  i-th position of the order. If the index is CUDD_CONST_INDEX, returns
  CUDD_CONST_INDEX; otherwise, if the index is out of bounds returns -1.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadPerm Cudd_ReadInvPermZdd]

******************************************************************************/
func Cudd_ReadInvPermZdd(dd *DdManager, i int32) int32 {
    if i == CUDD_CONST_INDEX { return(CUDD_CONST_INDEX); }
    if i < 0 || i >= dd.sizeZ { return(-1); }
    return(dd.invpermZ[i]);

} /* end of Cudd_ReadInvPermZdd */


/**Function********************************************************************

  Synopsis    [Returns the i-th element of the vars array.]

  Description [Returns the i-th element of the vars array if it falls
  within the array bounds; NULL otherwise. If i is the index of an
  existing variable, this function produces the same result as
  Cudd_bddIthVar. However, if the i-th var does not exist yet,
  Cudd_bddIthVar will create it, whereas Cudd_ReadVars will not.]

  SideEffects [None]

  SeeAlso     [Cudd_bddIthVar]

******************************************************************************/
func Cudd_ReadVars(dd *DdManager, i int32) *DdNode {
    if i < 0 || i > dd.size { return(nil); }
    return(dd.vars[i]);

} /* end of Cudd_ReadVars */


/**Function********************************************************************

  Synopsis    [Reads the epsilon parameter of the manager.]

  Description [Reads the epsilon parameter of the manager. The epsilon
  parameter control the comparison between floating point numbers.]

  SideEffects [None]

  SeeAlso     [Cudd_SetEpsilon]

******************************************************************************/
func Cudd_ReadEpsilon(dd *DdManager) CUDD_VALUE_TYPE {
    return(dd.epsilon);

} /* end of Cudd_ReadEpsilon */


/**Function********************************************************************

  Synopsis    [Sets the epsilon parameter of the manager to ep.]

  Description [Sets the epsilon parameter of the manager to ep. The epsilon
  parameter control the comparison between floating point numbers.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadEpsilon]

******************************************************************************/
func Cudd_SetEpsilon(dd *DdManager, ep CUDD_VALUE_TYPE) {
    dd.epsilon = ep;

} /* end of Cudd_SetEpsilon */


/**Function********************************************************************

  Synopsis    [Reads the groupcheck parameter of the manager.]

  Description [Reads the groupcheck parameter of the manager. The
  groupcheck parameter determines the aggregation criterion in group
  sifting.]

  SideEffects [None]

  SeeAlso     [Cudd_SetGroupcheck]

******************************************************************************/
func Cudd_ReadGroupcheck(dd *DdManager) Cudd_AggregationType {
    return(dd.groupcheck);

} /* end of Cudd_ReadGroupCheck */


/**Function********************************************************************

  Synopsis    [Sets the parameter groupcheck of the manager to gc.]

  Description [Sets the parameter groupcheck of the manager to gc. The
  groupcheck parameter determines the aggregation criterion in group
  sifting.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadGroupCheck]

******************************************************************************/
func Cudd_SetGroupcheck(dd *DdManager, gc Cudd_AggregationType) {
    dd.groupcheck = gc;

} /* end of Cudd_SetGroupcheck */


/**Function********************************************************************

  Synopsis    [Tells whether garbage collection is enabled.]

  Description [Returns 1 if garbage collection is enabled; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_EnableGarbageCollection Cudd_DisableGarbageCollection]

******************************************************************************/
func Cudd_GarbageCollectionEnabled(dd *DdManager) int32 {
    return(dd.gcEnabled);

} /* end of Cudd_GarbageCollectionEnabled */


/**Function********************************************************************

  Synopsis    [Enables garbage collection.]

  Description [Enables garbage collection. Garbage collection is
  initially enabled. Therefore it is necessary to call this function
  only if garbage collection has been explicitly disabled.]

  SideEffects [None]

  SeeAlso     [Cudd_DisableGarbageCollection Cudd_GarbageCollectionEnabled]

******************************************************************************/
func Cudd_EnableGarbageCollection(dd *DdManager) {
    dd.gcEnabled = 1;

} /* end of Cudd_EnableGarbageCollection */


/**Function********************************************************************

  Synopsis    [Disables garbage collection.]

  Description [Disables garbage collection. Garbage collection is
  initially enabled. This function may be called to disable it.
  However, garbage collection will still occur when a new node must be
  created and no memory is left, or when garbage collection is required
  for correctness. (E.g., before reordering.)]

  SideEffects [None]

  SeeAlso     [Cudd_EnableGarbageCollection Cudd_GarbageCollectionEnabled]

******************************************************************************/
func Cudd_DisableGarbageCollection(dd *DdManager) {
    dd.gcEnabled = 0;

} /* end of Cudd_DisableGarbageCollection */


/**Function********************************************************************

  Synopsis    [Tells whether dead nodes are counted towards triggering
  reordering.]

  Description [Tells whether dead nodes are counted towards triggering
  reordering. Returns 1 if dead nodes are counted; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_TurnOnCountDead Cudd_TurnOffCountDead]

******************************************************************************/
func Cudd_DeadAreCounted(dd *DdManager) int32 {
    if dd.countDead == 0 {
        return 1;
    } else {
        return 0;
    }

} /* end of Cudd_DeadAreCounted */


/**Function********************************************************************

  Synopsis    [Causes the dead nodes to be counted towards triggering
  reordering.]

  Description [Causes the dead nodes to be counted towards triggering
  reordering. This causes more frequent reorderings. By default dead
  nodes are not counted.]

  SideEffects [Changes the manager.]

  SeeAlso     [Cudd_TurnOffCountDead Cudd_DeadAreCounted]

******************************************************************************/
func Cudd_TurnOnCountDead(dd *DdManager) {
    dd.countDead = 0;

} /* end of Cudd_TurnOnCountDead */


/**Function********************************************************************

  Synopsis    [Causes the dead nodes not to be counted towards triggering
  reordering.]

  Description [Causes the dead nodes not to be counted towards
  triggering reordering. This causes less frequent reorderings. By
  default dead nodes are not counted. Therefore there is no need to
  call this function unless Cudd_TurnOnCountDead has been previously
  called.]

  SideEffects [Changes the manager.]

  SeeAlso     [Cudd_TurnOnCountDead Cudd_DeadAreCounted]

******************************************************************************/
func Cudd_TurnOffCountDead(dd *DdManager) {
    dd.countDead = ~0;

} /* end of Cudd_TurnOffCountDead */


/**Function********************************************************************

  Synopsis    [Returns the current value of the recombination parameter used
  in group sifting.]

  Description [Returns the current value of the recombination
  parameter used in group sifting. A larger (positive) value makes the
  aggregation of variables due to the second difference criterion more
  likely. A smaller (negative) value makes aggregation less likely.]

  SideEffects [None]

  SeeAlso     [Cudd_SetRecomb]

******************************************************************************/
func Cudd_ReadRecomb(dd *DdManager) int32 {
    return(dd.recomb);

} /* end of Cudd_ReadRecomb */


/**Function********************************************************************

  Synopsis    [Sets the value of the recombination parameter used in group
  sifting.]

  Description [Sets the value of the recombination parameter used in
  group sifting. A larger (positive) value makes the aggregation of
  variables due to the second difference criterion more likely. A
  smaller (negative) value makes aggregation less likely. The default
  value is 0.]

  SideEffects [Changes the manager.]

  SeeAlso     [Cudd_ReadRecomb]

******************************************************************************/
func Cudd_SetRecomb(dd *DdManager, recomb int32) {
    dd.recomb = recomb;

} /* end of Cudd_SetRecomb */


/**Function********************************************************************

  Synopsis    [Returns the current value of the symmviolation parameter used
  in group sifting.]

  Description [Returns the current value of the symmviolation
  parameter. This parameter is used in group sifting to decide how
  many violations to the symmetry conditions <code>f10 = f01</code> or
  <code>f11 = f00</code> are tolerable when checking for aggregation
  due to extended symmetry. The value should be between 0 and 100. A
  small value causes fewer variables to be aggregated. The default
  value is 0.]

  SideEffects [None]

  SeeAlso     [Cudd_SetSymmviolation]

******************************************************************************/
func Cudd_ReadSymmviolation(dd *DdManager) int32 {
    return(dd.symmviolation);

} /* end of Cudd_ReadSymmviolation */


/**Function********************************************************************

  Synopsis    [Sets the value of the symmviolation parameter used
  in group sifting.]

  Description [Sets the value of the symmviolation
  parameter. This parameter is used in group sifting to decide how
  many violations to the symmetry conditions <code>f10 = f01</code> or
  <code>f11 = f00</code> are tolerable when checking for aggregation
  due to extended symmetry. The value should be between 0 and 100. A
  small value causes fewer variables to be aggregated. The default
  value is 0.]

  SideEffects [Changes the manager.]

  SeeAlso     [Cudd_ReadSymmviolation]

******************************************************************************/
func Cudd_SetSymmviolation(dd *DdManager, symmviolation int32) {
    dd.symmviolation = symmviolation;

} /* end of Cudd_SetSymmviolation */


/**Function********************************************************************

  Synopsis    [Returns the current value of the arcviolation parameter used
  in group sifting.]

  Description [Returns the current value of the arcviolation
  parameter. This parameter is used in group sifting to decide how
  many arcs into <code>y</code> not coming from <code>x</code> are
  tolerable when checking for aggregation due to extended
  symmetry. The value should be between 0 and 100. A small value
  causes fewer variables to be aggregated. The default value is 0.]

  SideEffects [None]

  SeeAlso     [Cudd_SetArcviolation]

******************************************************************************/
func Cudd_ReadArcviolation(dd *DdManager) int32 {
    return(dd.arcviolation);

} /* end of Cudd_ReadArcviolation */


/**Function********************************************************************

  Synopsis    [Sets the value of the arcviolation parameter used
  in group sifting.]

  Description [Sets the value of the arcviolation
  parameter. This parameter is used in group sifting to decide how
  many arcs into <code>y</code> not coming from <code>x</code> are
  tolerable when checking for aggregation due to extended
  symmetry. The value should be between 0 and 100. A small value
  causes fewer variables to be aggregated. The default value is 0.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadArcviolation]

******************************************************************************/
func Cudd_SetArcviolation(dd *DdManager, arcviolation int32) {
    dd.arcviolation = arcviolation;

} /* end of Cudd_SetArcviolation */


/**Function********************************************************************

  Synopsis    [Reads the current size of the population used by the
  genetic algorithm for reordering.]

  Description [Reads the current size of the population used by the
  genetic algorithm for variable reordering. A larger population size will
  cause the genetic algorithm to take more time, but will generally
  produce better results. The default value is 0, in which case the
  package uses three times the number of variables as population size,
  with a maximum of 120.]

  SideEffects [None]

  SeeAlso     [Cudd_SetPopulationSize]

******************************************************************************/
func Cudd_ReadPopulationSize(dd *DdManager) int32 {
    return(dd.populationSize);

} /* end of Cudd_ReadPopulationSize */


/**Function********************************************************************

  Synopsis    [Sets the size of the population used by the
  genetic algorithm for reordering.]

  Description [Sets the size of the population used by the
  genetic algorithm for variable reordering. A larger population size will
  cause the genetic algorithm to take more time, but will generally
  produce better results. The default value is 0, in which case the
  package uses three times the number of variables as population size,
  with a maximum of 120.]

  SideEffects [Changes the manager.]

  SeeAlso     [Cudd_ReadPopulationSize]

******************************************************************************/
func Cudd_SetPopulationSize(dd *DdManager, populationSize int32) {
    dd.populationSize = populationSize;

} /* end of Cudd_SetPopulationSize */


/**Function********************************************************************

  Synopsis    [Reads the current number of crossovers used by the
  genetic algorithm for reordering.]

  Description [Reads the current number of crossovers used by the
  genetic algorithm for variable reordering. A larger number of crossovers will
  cause the genetic algorithm to take more time, but will generally
  produce better results. The default value is 0, in which case the
  package uses three times the number of variables as number of crossovers,
  with a maximum of 60.]

  SideEffects [None]

  SeeAlso     [Cudd_SetNumberXovers]

******************************************************************************/
func Cudd_ReadNumberXovers(dd *DdManager) int32 {
    return(dd.numberXovers);

} /* end of Cudd_ReadNumberXovers */


/**Function********************************************************************

  Synopsis    [Sets the number of crossovers used by the
  genetic algorithm for reordering.]

  Description [Sets the number of crossovers used by the genetic
  algorithm for variable reordering. A larger number of crossovers
  will cause the genetic algorithm to take more time, but will
  generally produce better results. The default value is 0, in which
  case the package uses three times the number of variables as number
  of crossovers, with a maximum of 60.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadNumberXovers]

******************************************************************************/
func Cudd_SetNumberXovers(dd *DdManager, numberXovers int32) {
    dd.numberXovers = numberXovers;

} /* end of Cudd_SetNumberXovers */

/**Function********************************************************************

  Synopsis    [Returns the memory in use by the manager measured in bytes.]

  Description []

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_ReadMemoryInUse(dd *DdManager) unsigned long {
    return(dd.memused);

} /* end of Cudd_ReadMemoryInUse */


/**Function********************************************************************

  Synopsis    [Prints out statistics and settings for a CUDD manager.]

  Description [Prints out statistics and settings for a CUDD manager.
  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     []

******************************************************************************/
func Cudd_PrintInfo(dd *DdManager, fp *os.File) int32 {
    var retval int32;
    var autoMethod, autoMethodZ Cudd_ReorderingType;

    /* Modifiable parameters. */
    retval = fmt.Fprintf(fp, "**** CUDD modifiable parameters ****\n");
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Hard limit for cache size: %d\n", Cudd_ReadMaxCacheHard(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Cache hit threshold for resizing: %d%%\n", Cudd_ReadMinHit(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Garbage collection enabled: %s\n", Cudd_GarbageCollectionEnabled(dd) ? "yes" : "no");
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Limit for fast unique table growth: %d\n", Cudd_ReadLooseUpTo(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Maximum number of variables sifted per reordering: %d\n", Cudd_ReadSiftMaxVar(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Maximum number of variable swaps per reordering: %d\n", Cudd_ReadSiftMaxSwap(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Maximum growth while sifting a variable: %g\n", Cudd_ReadMaxGrowth(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Dynamic reordering of BDDs enabled: %s\n", Cudd_ReorderingStatus(dd,&autoMethod) ? "yes" : "no");
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Default BDD reordering method: %d\n", int32( autoMethod));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Dynamic reordering of ZDDs enabled: %s\n", Cudd_ReorderingStatusZdd(dd,&autoMethodZ) ? "yes" : "no");
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Default ZDD reordering method: %d\n", int32( autoMethodZ));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Realignment of ZDDs to BDDs enabled: %s\n", Cudd_zddRealignmentEnabled(dd) ? "yes" : "no");
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Realignment of BDDs to ZDDs enabled: %s\n", Cudd_bddRealignmentEnabled(dd) ? "yes" : "no");
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Dead nodes counted in triggering reordering: %s\n", Cudd_DeadAreCounted(dd) ? "yes" : "no");
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Group checking criterion: %d\n", int32( Cudd_ReadGroupcheck(dd)));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Recombination threshold: %d\n", Cudd_ReadRecomb(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Symmetry violation threshold: %d\n", Cudd_ReadSymmviolation(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Arc violation threshold: %d\n", Cudd_ReadArcviolation(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "GA population size: %d\n", Cudd_ReadPopulationSize(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of crossovers for GA: %d\n", Cudd_ReadNumberXovers(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Next reordering threshold: %d\n", Cudd_ReadNextReordering(dd));
    if retval == -1 { return(0); }

    /* Non-modifiable parameters. */
    retval = fmt.Fprintf(fp, "**** CUDD non-modifiable parameters ****\n");
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Memory in use: %d\n", Cudd_ReadMemoryInUse(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Peak number of nodes: %d\n", Cudd_ReadPeakNodeCount(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Peak number of live nodes: %d\n", Cudd_ReadPeakLiveNodeCount(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of BDD variables: %d\n", dd.size);
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of ZDD variables: %d\n", dd.sizeZ);
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of cache entries: %d\n", dd.cacheSlots);
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of cache look-ups: %.0f\n", Cudd_ReadCacheLookUps(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of cache hits: %.0f\n", Cudd_ReadCacheHits(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of cache insertions: %.0f\n", dd.cacheinserts);
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of cache collisions: %.0f\n", dd.cachecollisions);
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of cache deletions: %.0f\n", dd.cachedeletions);
    if retval == -1 { return(0); }
    retval = cuddCacheProfile(dd,fp);
    if retval == 0 { return(0); }
    retval = fmt.Fprintf(fp, "Soft limit for cache size: %d\n", Cudd_ReadMaxCache(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of buckets in unique table: %d\n", dd.slots);
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Used buckets in unique table: %.2f%% (expected %.2f%%)\n", 100.0 * Cudd_ReadUsedSlots(dd), 100.0 * Cudd_ExpectedUsedSlots(dd));
    if retval == -1 { return(0); }
#ifdef DD_UNIQUE_PROFILE
    retval = fprintf(fp,"Unique lookups: %.0f\n", dd->uniqueLookUps);
    if (retval == EOF) return(0);
    retval = fprintf(fp,"Unique links: %.0f (%g per lookup)\n",
            dd->uniqueLinks, dd->uniqueLinks / dd->uniqueLookUps);
    if (retval == EOF) return(0);
#endif
    retval = fmt.Fprintf(fp, "Number of BDD and ADD nodes: %d\n", dd.keys);
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of ZDD nodes: %d\n", dd.keysZ);
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of dead BDD and ADD nodes: %d\n", dd.dead);
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Number of dead ZDD nodes: %d\n", dd.deadZ);
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Total number of nodes allocated: %d\n", int32(dd.allocated));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Total number of nodes reclaimed: %.0f\n", dd.reclaimed);
    if retval == -1 { return(0); }
#ifdef DD_STATS
    retval = fprintf(fp,"Nodes freed: %.0f\n", dd->nodesFreed);
    if (retval == EOF) return(0);
    retval = fprintf(fp,"Nodes dropped: %.0f\n", dd->nodesDropped);
    if (retval == EOF) return(0);
#endif
#ifdef DD_COUNT
    retval = fprintf(fp,"Number of recursive calls: %.0f\n",
                     Cudd_ReadRecursiveCalls(dd));
    if (retval == EOF) return(0);
#endif
    retval = fmt.Fprintf(fp, "Garbage collections so far: %d\n", Cudd_ReadGarbageCollections(dd));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Time for garbage collection: %.2f sec\n", (float64(Cudd_ReadGarbageCollectionTime(dd))/1000.0));
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Reorderings so far: %d\n", dd.reorderings);
    if retval == -1 { return(0); }
    retval = fmt.Fprintf(fp, "Time for reordering: %.2f sec\n", (float64(Cudd_ReadReorderingTime(dd))/1000.0));
    if retval == -1 { return(0); }
#ifdef DD_COUNT
    retval = fprintf(fp,"Node swaps in reordering: %.0f\n",
        Cudd_ReadSwapSteps(dd));
    if (retval == EOF) return(0);
#endif

    return(1);

} /* end of Cudd_PrintInfo */


/**Function********************************************************************

  Synopsis    [Reports the peak number of nodes.]

  Description [Reports the peak number of nodes. This number includes
  node on the free list. At the peak, the number of nodes on the free
  list is guaranteed to be less than DD_MEM_CHUNK.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadNodeCount Cudd_PrintInfo]

******************************************************************************/
func Cudd_ReadPeakNodeCount(dd *DdManager) long {
    var count long;
    var scan *DdNodePtr  = dd.memoryList;

    for scan != nil {
        count += DD_MEM_CHUNK;
        scan = (DdNodePtr *) *scan;
    }
    return(count);

} /* end of Cudd_ReadPeakNodeCount */


/**Function********************************************************************

  Synopsis    [Reports the peak number of live nodes.]

  Description [Reports the peak number of live nodes. This count is kept
  only if CUDD is compiled with DD_STATS defined. If DD_STATS is not
  defined, this function returns -1.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadNodeCount Cudd_PrintInfo Cudd_ReadPeakNodeCount]

******************************************************************************/
func Cudd_ReadPeakLiveNodeCount(dd *DdManager) int32 {
    unsigned int live = dd.keys - dd.dead;

    if live > dd.peakLiveNodes {
        dd.peakLiveNodes = live;
    }
    return(int32(dd.peakLiveNodes));

} /* end of Cudd_ReadPeakLiveNodeCount */


/**Function********************************************************************

  Synopsis    [Reports the number of nodes in BDDs and ADDs.]

  Description [Reports the number of live nodes in BDDs and ADDs. This
  number does not include the isolated projection functions and the
  unused constants. These nodes that are not counted are not part of
  the DDs manipulated by the application.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadPeakNodeCount Cudd_zddReadNodeCount]

******************************************************************************/
func Cudd_ReadNodeCount(dd *DdManager) long {
    var count long;
    var i int32;

#ifndef DD_NO_DEATH_ROW
    cuddClearDeathRow(dd);
#endif

    count = int64( (dd.keys - dd.dead));

    /* Count isolated projection functions. Their number is subtracted
    ** from the node count because they are not part of the BDDs.
    */
    for i=0; i < dd.size; i++ {
        if dd.vars[i].ref == 1 { count--; }
    }
    /* Subtract from the count the unused constants. */
    // inlining of macro DD_ZERO
    if ((dd).zero).ref == 1 { count--; }
    // inlining of macro DD_PLUS_INFINITY
    if ((dd).plusinfinity).ref == 1 { count--; }
    // inlining of macro DD_MINUS_INFINITY
    if ((dd).minusinfinity).ref == 1 { count--; }

    return(count);

} /* end of Cudd_ReadNodeCount */



/**Function********************************************************************

  Synopsis    [Reports the number of nodes in ZDDs.]

  Description [Reports the number of nodes in ZDDs. This
  number always includes the two constants 1 and 0.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadPeakNodeCount Cudd_ReadNodeCount]

******************************************************************************/
func Cudd_zddReadNodeCount(dd *DdManager) long {
    return(int64((dd.keysZ - dd.deadZ + 2)));

} /* end of Cudd_zddReadNodeCount */


/**Function********************************************************************

  Synopsis    [Adds a function to a hook.]

  Description [Adds a function to a hook. A hook is a list of
  application-provided functions called on certain occasions by the
  package. Returns 1 if the function is successfully added; 2 if the
  function was already in the list; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_RemoveHook]

******************************************************************************/
func Cudd_AddHook(dd *DdManager, f DD_HFP, where Cudd_HookType) int32 {
    var hook **DdHook;
    var nextHook *DdHook;
    var newHook *DdHook;

    switch where {
    case CUDD_PRE_GC_HOOK:
        hook = &(dd.preGCHook);
        break;
    case CUDD_POST_GC_HOOK:
        hook = &(dd.postGCHook);
        break;
    case CUDD_PRE_REORDERING_HOOK:
        hook = &(dd.preReorderingHook);
        break;
    case CUDD_POST_REORDERING_HOOK:
        hook = &(dd.postReorderingHook);
        break;
    default:
        return(0);
    }
    /* Scan the list and find whether the function is already there.
    ** If so, just return. */
    nextHook = *hook;
    for nextHook != nil {
        if nextHook.f == f {
            return(2);
        }
        hook = &(nextHook.next);
        nextHook = nextHook.next;
    }
    /* The function was not in the list. Create a new item and append it
    ** to the end of the list. */
    newHook = new(DdHook);
    if newHook == nil {
        dd.errorCode = CUDD_MEMORY_OUT;
        return(0);
    }
    newHook.next = nil;
    newHook.f = f;
    *hook = newHook;
    return(1);

} /* end of Cudd_AddHook */


/**Function********************************************************************

  Synopsis    [Removes a function from a hook.]

  Description [Removes a function from a hook. A hook is a list of
  application-provided functions called on certain occasions by the
  package. Returns 1 if successful; 0 the function was not in the list.]

  SideEffects [None]

  SeeAlso     [Cudd_AddHook]

******************************************************************************/
func Cudd_RemoveHook(dd *DdManager, f DD_HFP, where Cudd_HookType) int32 {
    var hook **DdHook;
    var nextHook *DdHook;

    switch where {
    case CUDD_PRE_GC_HOOK:
        hook = &(dd.preGCHook);
        break;
    case CUDD_POST_GC_HOOK:
        hook = &(dd.postGCHook);
        break;
    case CUDD_PRE_REORDERING_HOOK:
        hook = &(dd.preReorderingHook);
        break;
    case CUDD_POST_REORDERING_HOOK:
        hook = &(dd.postReorderingHook);
        break;
    default:
        return(0);
    }
    nextHook = *hook;
    for nextHook != nil {
        if nextHook.f == f {
            *hook = nextHook.next;
            ABC_FREE(nextHook);
            return(1);
        }
        hook = &(nextHook.next);
        nextHook = nextHook.next;
    }

    return(0);

} /* end of Cudd_RemoveHook */


/**Function********************************************************************

  Synopsis    [Checks whether a function is in a hook.]

  Description [Checks whether a function is in a hook. A hook is a list of
  application-provided functions called on certain occasions by the
  package. Returns 1 if the function is found; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_AddHook Cudd_RemoveHook]

******************************************************************************/
func Cudd_IsInHook(dd *DdManager, f DD_HFP, where Cudd_HookType) int32 {
    var hook *DdHook;

    switch where {
    case CUDD_PRE_GC_HOOK:
        hook = dd.preGCHook;
        break;
    case CUDD_POST_GC_HOOK:
        hook = dd.postGCHook;
        break;
    case CUDD_PRE_REORDERING_HOOK:
        hook = dd.preReorderingHook;
        break;
    case CUDD_POST_REORDERING_HOOK:
        hook = dd.postReorderingHook;
        break;
    default:
        return(0);
    }
    /* Scan the list and find whether the function is already there. */
    for hook != nil {
        if hook.f == f {
            return(1);
        }
        hook = hook.next;
    }
    return(0);

} /* end of Cudd_IsInHook */


/**Function********************************************************************

  Synopsis    [Sample hook function to call before reordering.]

  Description [Sample hook function to call before reordering.
  Prints on the manager's stdout reordering method and initial size.
  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_StdPostReordHook]

******************************************************************************/
func Cudd_StdPreReordHook(dd *DdManager, str string, data any) int32 {
    var method Cudd_ReorderingType  = (Cudd_ReorderingType) (ptruint) data;
    var retval int32;

    retval = fmt.Fprintf(dd.out, "%s reordering with ", str);
    if retval == -1 { return(0); }
    switch method {
    case CUDD_REORDER_SIFT_CONVERGE,
     CUDD_REORDER_SYMM_SIFT_CONV,
     CUDD_REORDER_GROUP_SIFT_CONV,
     CUDD_REORDER_WINDOW2_CONV,
     CUDD_REORDER_WINDOW3_CONV,
     CUDD_REORDER_WINDOW4_CONV,
     CUDD_REORDER_LINEAR_CONVERGE:
        retval = fmt.Fprintf(dd.out, "converging ");
        if retval == -1 { return(0); }
        break;
    default:
        break;
    }
    switch method {
    case CUDD_REORDER_RANDOM,
     CUDD_REORDER_RANDOM_PIVOT:
        retval = fmt.Fprintf(dd.out, "random");
        break;
    case CUDD_REORDER_SIFT,
     CUDD_REORDER_SIFT_CONVERGE:
        retval = fmt.Fprintf(dd.out, "sifting");
        break;
    case CUDD_REORDER_SYMM_SIFT,
     CUDD_REORDER_SYMM_SIFT_CONV:
        retval = fmt.Fprintf(dd.out, "symmetric sifting");
        break;
    case CUDD_REORDER_LAZY_SIFT:
        retval = fmt.Fprintf(dd.out, "lazy sifting");
        break;
    case CUDD_REORDER_GROUP_SIFT,
     CUDD_REORDER_GROUP_SIFT_CONV:
        retval = fmt.Fprintf(dd.out, "group sifting");
        break;
    case CUDD_REORDER_WINDOW2,
     CUDD_REORDER_WINDOW3,
     CUDD_REORDER_WINDOW4,
     CUDD_REORDER_WINDOW2_CONV,
     CUDD_REORDER_WINDOW3_CONV,
     CUDD_REORDER_WINDOW4_CONV:
        retval = fmt.Fprintf(dd.out, "window");
        break;
    case CUDD_REORDER_ANNEALING:
        retval = fmt.Fprintf(dd.out, "annealing");
        break;
    case CUDD_REORDER_GENETIC:
        retval = fmt.Fprintf(dd.out, "genetic");
        break;
    case CUDD_REORDER_LINEAR,
     CUDD_REORDER_LINEAR_CONVERGE:
        retval = fmt.Fprintf(dd.out, "linear sifting");
        break;
    case CUDD_REORDER_EXACT:
        retval = fmt.Fprintf(dd.out, "exact");
        break;
    default:
        return(0);
    }
    if retval == -1 { return(0); }

    retval = fmt.Fprintf(dd.out, ": from %d to ... ", int32(strings.Compare(str, "BDD")) == 0 ?
                         Cudd_ReadNodeCount(dd) : Cudd_zddReadNodeCount(dd));
    if retval == -1 { return(0); }
    fflush(dd.out);
    return(1);

} /* end of Cudd_StdPreReordHook */


/**Function********************************************************************

  Synopsis    [Sample hook function to call after reordering.]

  Description [Sample hook function to call after reordering.
  Prints on the manager's stdout final size and reordering time.
  Returns 1 if successful; 0 otherwise.]

  SideEffects [None]

  SeeAlso     [Cudd_StdPreReordHook]

******************************************************************************/
func Cudd_StdPostReordHook(dd *DdManager, str string, data any) int32 {
    var initialTime long  = int64( data);
    var retval int32;
    var finalTime long  = util_cpu_time();
    var totalTimeSec float64  = float64((finalTime - initialTime)) / 1000.0;

    retval = fmt.Fprintf(dd.out, "%d nodes in %g sec\n", int32(strings.Compare(str, "BDD")) == 0 ?
                         Cudd_ReadNodeCount(dd) : Cudd_zddReadNodeCount(dd), totalTimeSec);
    if retval == -1 { return(0); }
    retval = fflush(dd.out);
    if retval == -1 { return(0); }
    return(1);

} /* end of Cudd_StdPostReordHook */


/**Function********************************************************************

  Synopsis    [Enables reporting of reordering stats.]

  Description [Enables reporting of reordering stats.
  Returns 1 if successful; 0 otherwise.]

  SideEffects [Installs functions in the pre-reordering and post-reordering
  hooks.]

  SeeAlso     [Cudd_DisableReorderingReporting Cudd_ReorderingReporting]

******************************************************************************/
func Cudd_EnableReorderingReporting(dd *DdManager) int32 {
    if !(Cudd_AddHook(dd, Cudd_StdPreReordHook, CUDD_PRE_REORDERING_HOOK) != 0) {
        return(0);
    }
    if !(Cudd_AddHook(dd, Cudd_StdPostReordHook, CUDD_POST_REORDERING_HOOK) != 0) {
        return(0);
    }
    return(1);

} /* end of Cudd_EnableReorderingReporting */


/**Function********************************************************************

  Synopsis    [Disables reporting of reordering stats.]

  Description [Disables reporting of reordering stats.
  Returns 1 if successful; 0 otherwise.]

  SideEffects [Removes functions from the pre-reordering and post-reordering
  hooks.]

  SeeAlso     [Cudd_EnableReorderingReporting Cudd_ReorderingReporting]

******************************************************************************/
func Cudd_DisableReorderingReporting(dd *DdManager) int32 {
    if !(Cudd_RemoveHook(dd, Cudd_StdPreReordHook, CUDD_PRE_REORDERING_HOOK) != 0) {
        return(0);
    }
    if !(Cudd_RemoveHook(dd, Cudd_StdPostReordHook, CUDD_POST_REORDERING_HOOK) != 0) {
        return(0);
    }
    return(1);

} /* end of Cudd_DisableReorderingReporting */


/**Function********************************************************************

  Synopsis    [Returns 1 if reporting of reordering stats is enabled.]

  Description [Returns 1 if reporting of reordering stats is enabled;
  0 otherwise.]

  SideEffects [none]

  SeeAlso     [Cudd_EnableReorderingReporting Cudd_DisableReorderingReporting]

******************************************************************************/
func Cudd_ReorderingReporting(dd *DdManager) int32 {
    return(Cudd_IsInHook(dd, Cudd_StdPreReordHook, CUDD_PRE_REORDERING_HOOK));

} /* end of Cudd_ReorderingReporting */


/**Function********************************************************************

  Synopsis    [Returns the code of the last error.]

  Description [Returns the code of the last error. The error codes are
  defined in cudd.h.]

  SideEffects [None]

  SeeAlso     [Cudd_ClearErrorCode]

******************************************************************************/
func Cudd_ReadErrorCode(dd *DdManager) Cudd_ErrorType {
    return(dd.errorCode);

} /* end of Cudd_ReadErrorCode */


/**Function********************************************************************

  Synopsis    [Clear the error code of a manager.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_ReadErrorCode]

******************************************************************************/
func Cudd_ClearErrorCode(dd *DdManager) {
    dd.errorCode = CUDD_NO_ERROR;

} /* end of Cudd_ClearErrorCode */


/**Function********************************************************************

  Synopsis    [Reads the stdout of a manager.]

  Description [Reads the stdout of a manager. This is the file pointer to
  which messages normally going to stdout are written. It is initialized
  to stdout. Cudd_SetStdout allows the application to redirect it.]

  SideEffects [None]

  SeeAlso     [Cudd_SetStdout Cudd_ReadStderr]

******************************************************************************/
func Cudd_ReadStdout(dd *DdManager) *os.File {
    return(dd.out);

} /* end of Cudd_ReadStdout */


/**Function********************************************************************

  Synopsis    [Sets the stdout of a manager.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_ReadStdout Cudd_SetStderr]

******************************************************************************/
func Cudd_SetStdout(dd *DdManager, fp *os.File) {
    dd.out = fp;

} /* end of Cudd_SetStdout */


/**Function********************************************************************

  Synopsis    [Reads the stderr of a manager.]

  Description [Reads the stderr of a manager. This is the file pointer to
  which messages normally going to stderr are written. It is initialized
  to stderr. Cudd_SetStderr allows the application to redirect it.]

  SideEffects [None]

  SeeAlso     [Cudd_SetStderr Cudd_ReadStdout]

******************************************************************************/
func Cudd_ReadStderr(dd *DdManager) *os.File {
    return(dd.err);

} /* end of Cudd_ReadStderr */


/**Function********************************************************************

  Synopsis    [Sets the stderr of a manager.]

  Description []

  SideEffects [None]

  SeeAlso     [Cudd_ReadStderr Cudd_SetStdout]

******************************************************************************/
func Cudd_SetStderr(dd *DdManager, fp *os.File) {
    dd.err = fp;

} /* end of Cudd_SetStderr */


/**Function********************************************************************

  Synopsis    [Returns the threshold for the next dynamic reordering.]

  Description [Returns the threshold for the next dynamic reordering.
  The threshold is in terms of number of nodes and is in effect only
  if reordering is enabled. The count does not include the dead nodes,
  unless the countDead parameter of the manager has been changed from
  its default setting.]

  SideEffects [None]

  SeeAlso     [Cudd_SetNextReordering]

******************************************************************************/
func Cudd_ReadNextReordering(dd *DdManager) unsigned int {
    return(dd.nextDyn);

} /* end of Cudd_ReadNextReordering */


/**Function********************************************************************

  Synopsis    [Sets the threshold for the next dynamic reordering.]

  Description [Sets the threshold for the next dynamic reordering.
  The threshold is in terms of number of nodes and is in effect only
  if reordering is enabled. The count does not include the dead nodes,
  unless the countDead parameter of the manager has been changed from
  its default setting.]

  SideEffects [None]

  SeeAlso     [Cudd_ReadNextReordering]

******************************************************************************/
void
Cudd_SetNextReordering(
  DdManager *dd,
  unsigned int next)
{
    dd.nextDyn = next;

} /* end of Cudd_SetNextReordering */


/**Function********************************************************************

  Synopsis    [Reads the number of elementary reordering steps.]

  Description []

  SideEffects [none]

  SeeAlso     []

******************************************************************************/
func Cudd_ReadSwapSteps(dd *DdManager) float64 {
#ifdef DD_COUNT
    return(dd->swapSteps);
#else
    return(-1);
#endif

} /* end of Cudd_ReadSwapSteps */


/**Function********************************************************************

  Synopsis    [Reads the maximum allowed number of live nodes.]

  Description [Reads the maximum allowed number of live nodes. When this
  number is exceeded, the package returns NULL.]

  SideEffects [none]

  SeeAlso     [Cudd_SetMaxLive]

******************************************************************************/
func Cudd_ReadMaxLive(dd *DdManager) unsigned int {
    return(dd.maxLive);

} /* end of Cudd_ReadMaxLive */


/**Function********************************************************************

  Synopsis    [Sets the maximum allowed number of live nodes.]

  Description [Sets the maximum allowed number of live nodes. When this
  number is exceeded, the package returns NULL.]

  SideEffects [none]

  SeeAlso     [Cudd_ReadMaxLive]

******************************************************************************/
void
Cudd_SetMaxLive(
  DdManager *dd,
  unsigned int maxLive)
{
    dd.maxLive = maxLive;

} /* end of Cudd_SetMaxLive */


/**Function********************************************************************

  Synopsis    [Reads the maximum allowed memory.]

  Description [Reads the maximum allowed memory. When this
  number is exceeded, the package returns NULL.]

  SideEffects [none]

  SeeAlso     [Cudd_SetMaxMemory]

******************************************************************************/
func Cudd_ReadMaxMemory(dd *DdManager) unsigned long {
    return(dd.maxmemhard);

} /* end of Cudd_ReadMaxMemory */


/**Function********************************************************************

  Synopsis    [Sets the maximum allowed memory.]

  Description [Sets the maximum allowed memory. When this
  number is exceeded, the package returns NULL.]

  SideEffects [none]

  SeeAlso     [Cudd_ReadMaxMemory]

******************************************************************************/
void
Cudd_SetMaxMemory(
  DdManager *dd,
  unsigned long maxMemory)
{
    dd.maxmemhard = maxMemory;

} /* end of Cudd_SetMaxMemory */


/**Function********************************************************************

  Synopsis    [Prevents sifting of a variable.]

  Description [This function sets a flag to prevent sifting of a
  variable.  Returns 1 if successful; 0 otherwise (i.e., invalid
  variable index).]

  SideEffects [Changes the "bindVar" flag in DdSubtable.]

  SeeAlso     [Cudd_bddUnbindVar]

******************************************************************************/
int
Cudd_bddBindVar(
  DdManager *dd /* manager */,
  int index /* variable index */)
{
    if index >= dd.size || index < 0 { return(0); }
    dd.subtables[dd.perm[index]].bindVar = 1;
    return(1);

} /* end of Cudd_bddBindVar */


/**Function********************************************************************

  Synopsis    [Allows the sifting of a variable.]

  Description [This function resets the flag that prevents the sifting
  of a variable. In successive variable reorderings, the variable will
  NOT be skipped, that is, sifted.  Initially all variables can be
  sifted. It is necessary to call this function only to re-enable
  sifting after a call to Cudd_bddBindVar. Returns 1 if successful; 0
  otherwise (i.e., invalid variable index).]

  SideEffects [Changes the "bindVar" flag in DdSubtable.]

  SeeAlso     [Cudd_bddBindVar]

******************************************************************************/
int
Cudd_bddUnbindVar(
  DdManager *dd /* manager */,
  int index /* variable index */)
{
    if index >= dd.size || index < 0 { return(0); }
    dd.subtables[dd.perm[index]].bindVar = 0;
    return(1);

} /* end of Cudd_bddUnbindVar */


/**Function********************************************************************

  Synopsis    [Tells whether a variable can be sifted.]

  Description [This function returns 1 if a variable is enabled for
  sifting.  Initially all variables can be sifted. This function returns
  0 only if there has been a previous call to Cudd_bddBindVar for that
  variable not followed by a call to Cudd_bddUnbindVar. The function returns
  0 also in the case in which the index of the variable is out of bounds.]

  SideEffects [none]

  SeeAlso     [Cudd_bddBindVar Cudd_bddUnbindVar]

******************************************************************************/
int
Cudd_bddVarIsBound(
  DdManager *dd /* manager */,
  int index /* variable index */)
{
    if index >= dd.size || index < 0 { return(0); }
    return(dd.subtables[dd.perm[index]].bindVar);

} /* end of Cudd_bddVarIsBound */


/**Function********************************************************************

  Synopsis    [Sets a variable type to primary input.]

  Description [Sets a variable type to primary input.  The variable type is
  used by lazy sifting.  Returns 1 if successful; 0 otherwise.]

  SideEffects [modifies the manager]

  SeeAlso     [Cudd_bddSetPsVar Cudd_bddSetNsVar Cudd_bddIsPiVar]

******************************************************************************/
int
Cudd_bddSetPiVar(
  DdManager *dd /* manager */,
  int index /* variable index */)
{
    if index >= dd.size || index < 0 { return (0); }
    dd.subtables[dd.perm[index]].varType = CUDD_VAR_PRIMARY_INPUT;
    return(1);

} /* end of Cudd_bddSetPiVar */


/**Function********************************************************************

  Synopsis    [Sets a variable type to present state.]

  Description [Sets a variable type to present state.  The variable type is
  used by lazy sifting.  Returns 1 if successful; 0 otherwise.]

  SideEffects [modifies the manager]

  SeeAlso     [Cudd_bddSetPiVar Cudd_bddSetNsVar Cudd_bddIsPsVar]

******************************************************************************/
int
Cudd_bddSetPsVar(
  DdManager *dd /* manager */,
  int index /* variable index */)
{
    if index >= dd.size || index < 0 { return (0); }
    dd.subtables[dd.perm[index]].varType = CUDD_VAR_PRESENT_STATE;
    return(1);

} /* end of Cudd_bddSetPsVar */


/**Function********************************************************************

  Synopsis    [Sets a variable type to next state.]

  Description [Sets a variable type to next state.  The variable type is
  used by lazy sifting.  Returns 1 if successful; 0 otherwise.]

  SideEffects [modifies the manager]

  SeeAlso     [Cudd_bddSetPiVar Cudd_bddSetPsVar Cudd_bddIsNsVar]

******************************************************************************/
int
Cudd_bddSetNsVar(
  DdManager *dd /* manager */,
  int index /* variable index */)
{
    if index >= dd.size || index < 0 { return (0); }
    dd.subtables[dd.perm[index]].varType = CUDD_VAR_NEXT_STATE;
    return(1);

} /* end of Cudd_bddSetNsVar */


/**Function********************************************************************

  Synopsis    [Checks whether a variable is primary input.]

  Description [Checks whether a variable is primary input.  Returns 1 if
  the variable's type is primary input; 0 if the variable exists but is
  not a primary input; -1 if the variable does not exist.]

  SideEffects [none]

  SeeAlso     [Cudd_bddSetPiVar Cudd_bddIsPsVar Cudd_bddIsNsVar]

******************************************************************************/
int
Cudd_bddIsPiVar(
  DdManager *dd /* manager */,
  int index /* variable index */)
{
    if index >= dd.size || index < 0 { return -1; }
    return (dd.subtables[dd.perm[index]].varType == CUDD_VAR_PRIMARY_INPUT);

} /* end of Cudd_bddIsPiVar */


/**Function********************************************************************

  Synopsis    [Checks whether a variable is present state.]

  Description [Checks whether a variable is present state.  Returns 1 if
  the variable's type is present state; 0 if the variable exists but is
  not a present state; -1 if the variable does not exist.]

  SideEffects [none]

  SeeAlso     [Cudd_bddSetPsVar Cudd_bddIsPiVar Cudd_bddIsNsVar]

******************************************************************************/
func Cudd_bddIsPsVar(dd *DdManager, index int32) int32 {
    if index >= dd.size || index < 0 { return -1; }
    return (dd.subtables[dd.perm[index]].varType == CUDD_VAR_PRESENT_STATE);

} /* end of Cudd_bddIsPsVar */


/**Function********************************************************************

  Synopsis    [Checks whether a variable is next state.]

  Description [Checks whether a variable is next state.  Returns 1 if
  the variable's type is present state; 0 if the variable exists but is
  not a present state; -1 if the variable does not exist.]

  SideEffects [none]

  SeeAlso     [Cudd_bddSetNsVar Cudd_bddIsPiVar Cudd_bddIsPsVar]

******************************************************************************/
func Cudd_bddIsNsVar(dd *DdManager, index int32) int32 {
    if index >= dd.size || index < 0 { return -1; }
    return (dd.subtables[dd.perm[index]].varType == CUDD_VAR_NEXT_STATE);

} /* end of Cudd_bddIsNsVar */


/**Function********************************************************************

  Synopsis    [Sets a corresponding pair index for a given index.]

  Description [Sets a corresponding pair index for a given index.
  These pair indices are present and next state variable.  Returns 1 if
  successful; 0 otherwise.]

  SideEffects [modifies the manager]

  SeeAlso     [Cudd_bddReadPairIndex]

******************************************************************************/
int
Cudd_bddSetPairIndex(
  DdManager *dd /* manager */,
  int index /* variable index */,
  int pairIndex /* corresponding variable index */)
{
    if index >= dd.size || index < 0 { return(0); }
    dd.subtables[dd.perm[index]].pairIndex = pairIndex;
    return(1);

} /* end of Cudd_bddSetPairIndex */


/**Function********************************************************************

  Synopsis    [Reads a corresponding pair index for a given index.]

  Description [Reads a corresponding pair index for a given index.
  These pair indices are present and next state variable.  Returns the
  corresponding variable index if the variable exists; -1 otherwise.]

  SideEffects [modifies the manager]

  SeeAlso     [Cudd_bddSetPairIndex]

******************************************************************************/
func Cudd_bddReadPairIndex(dd *DdManager, index int32) int32 {
    if index >= dd.size || index < 0 { return -1; }
    return dd.subtables[dd.perm[index]].pairIndex;

} /* end of Cudd_bddReadPairIndex */


/**Function********************************************************************

  Synopsis    [Sets a variable to be grouped.]

  Description [Sets a variable to be grouped. This function is used for
  lazy sifting.  Returns 1 if successful; 0 otherwise.]

  SideEffects [modifies the manager]

  SeeAlso     [Cudd_bddSetVarHardGroup Cudd_bddResetVarToBeGrouped]

******************************************************************************/
func Cudd_bddSetVarToBeGrouped(dd *DdManager, index int32) int32 {
    if index >= dd.size || index < 0 { return(0); }
    if dd.subtables[dd.perm[index]].varToBeGrouped <= CUDD_LAZY_SOFT_GROUP {
        dd.subtables[dd.perm[index]].varToBeGrouped = CUDD_LAZY_SOFT_GROUP;
    }
    return(1);

} /* end of Cudd_bddSetVarToBeGrouped */


/**Function********************************************************************

  Synopsis    [Sets a variable to be a hard group.]

  Description [Sets a variable to be a hard group.  This function is used
  for lazy sifting.  Returns 1 if successful; 0 otherwise.]

  SideEffects [modifies the manager]

  SeeAlso     [Cudd_bddSetVarToBeGrouped Cudd_bddResetVarToBeGrouped
  Cudd_bddIsVarHardGroup]

******************************************************************************/
func Cudd_bddSetVarHardGroup(dd *DdManager, index int32) int32 {
    if index >= dd.size || index < 0 { return(0); }
    dd.subtables[dd.perm[index]].varToBeGrouped = CUDD_LAZY_HARD_GROUP;
    return(1);

} /* end of Cudd_bddSetVarHardGrouped */


/**Function********************************************************************

  Synopsis    [Resets a variable not to be grouped.]

  Description [Resets a variable not to be grouped.  This function is
  used for lazy sifting.  Returns 1 if successful; 0 otherwise.]

  SideEffects [modifies the manager]

  SeeAlso     [Cudd_bddSetVarToBeGrouped Cudd_bddSetVarHardGroup]

******************************************************************************/
func Cudd_bddResetVarToBeGrouped(dd *DdManager, index int32) int32 {
    if index >= dd.size || index < 0 { return(0); }
    if dd.subtables[dd.perm[index]].varToBeGrouped <=
        CUDD_LAZY_SOFT_GROUP {
        dd.subtables[dd.perm[index]].varToBeGrouped = CUDD_LAZY_NONE;
    }
    return(1);

} /* end of Cudd_bddResetVarToBeGrouped */


/**Function********************************************************************

  Synopsis    [Checks whether a variable is set to be grouped.]

  Description [Checks whether a variable is set to be grouped. This
  function is used for lazy sifting.]

  SideEffects [none]

  SeeAlso     []

******************************************************************************/
func Cudd_bddIsVarToBeGrouped(dd *DdManager, index int32) int32 {
    if index >= dd.size || index < 0 { return(-1); }
    if dd.subtables[dd.perm[index]].varToBeGrouped == CUDD_LAZY_UNGROUP {
        return(0);
    } else {
        return(dd.subtables[dd.perm[index]].varToBeGrouped); }

} /* end of Cudd_bddIsVarToBeGrouped */


/**Function********************************************************************

  Synopsis    [Sets a variable to be ungrouped.]

  Description [Sets a variable to be ungrouped. This function is used
  for lazy sifting.  Returns 1 if successful; 0 otherwise.]

  SideEffects [modifies the manager]

  SeeAlso     [Cudd_bddIsVarToBeUngrouped]

******************************************************************************/
func Cudd_bddSetVarToBeUngrouped(dd *DdManager, index int32) int32 {
    if index >= dd.size || index < 0 { return(0); }
    dd.subtables[dd.perm[index]].varToBeGrouped = CUDD_LAZY_UNGROUP;
    return(1);

} /* end of Cudd_bddSetVarToBeGrouped */


/**Function********************************************************************

  Synopsis    [Checks whether a variable is set to be ungrouped.]

  Description [Checks whether a variable is set to be ungrouped. This
  function is used for lazy sifting.  Returns 1 if the variable is marked
  to be ungrouped; 0 if the variable exists, but it is not marked to be
  ungrouped; -1 if the variable does not exist.]

  SideEffects [none]

  SeeAlso     [Cudd_bddSetVarToBeUngrouped]

******************************************************************************/
func Cudd_bddIsVarToBeUngrouped(dd *DdManager, index int32) int32 {
    if index >= dd.size || index < 0 { return(-1); }
    return dd.subtables[dd.perm[index]].varToBeGrouped == CUDD_LAZY_UNGROUP;

} /* end of Cudd_bddIsVarToBeGrouped */


/**Function********************************************************************

  Synopsis    [Checks whether a variable is set to be in a hard group.]

  Description [Checks whether a variable is set to be in a hard group.  This
  function is used for lazy sifting.  Returns 1 if the variable is marked
  to be in a hard group; 0 if the variable exists, but it is not marked to be
  in a hard group; -1 if the variable does not exist.]

  SideEffects [none]

  SeeAlso     [Cudd_bddSetVarHardGroup]

******************************************************************************/
func Cudd_bddIsVarHardGroup(dd *DdManager, index int32) int32 {
    if index >= dd.size || index < 0 { return(-1); }
    if dd.subtables[dd.perm[index]].varToBeGrouped == CUDD_LAZY_HARD_GROUP {
        return(1); }
    return(0);

} /* end of Cudd_bddIsVarToBeGrouped */


/*---------------------------------------------------------------------------*/
/* Definition of internal functions                                          */
/*---------------------------------------------------------------------------*/

/*---------------------------------------------------------------------------*/
/* Definition of static functions                                            */
/*---------------------------------------------------------------------------*/


/**Function********************************************************************

  Synopsis    [Fixes a variable group tree.]

  Description []

  SideEffects [Changes the variable group tree.]

  SeeAlso     []

******************************************************************************/
func fixVarTree(treenode *MtrNode, perm *int32, size int32) {
    treenode.index = treenode.low;
    if int32( treenode.index) < size {
        treenode.low = perm[treenode.index];
    } else {
        treenode.low = treenode.index;
    }
    if treenode.child != nil {
        fixVarTree(treenode.child, perm, size); }
    if treenode.younger != nil {
        fixVarTree(treenode.younger, perm, size); }
    return;

} /* end of fixVarTree */


/**Function********************************************************************

  Synopsis    [Adds multiplicity groups to a ZDD variable group tree.]

  Description [Adds multiplicity groups to a ZDD variable group tree.
  Returns 1 if successful; 0 otherwise. This function creates the groups
  for set of ZDD variables (whose cardinality is given by parameter
  multiplicity) that are created for each BDD variable in
  Cudd_zddVarsFromBddVars. The crux of the matter is to determine the index
  each new group. (The index of the first variable in the group.)
  We first build all the groups for the children of a node, and then deal
  with the ZDD variables that are directly attached to the node. The problem
  for these is that the tree itself does not provide information on their
  position inside the group. While we deal with the children of the node,
  therefore, we keep track of all the positions they occupy. The remaining
  positions in the tree can be freely used. Also, we keep track of all the
  variables placed in the children. All the remaining variables are directly
  attached to the group. We can then place any pair of variables not yet
  grouped in any pair of available positions in the node.]

  SideEffects [Changes the variable group tree.]

  SeeAlso     [Cudd_zddVarsFromBddVars]

******************************************************************************/
static int
addMultiplicityGroups(
  DdManager *dd /* manager */,
  MtrNode *treenode /* current tree node */,
  int multiplicity /* how many ZDD vars per BDD var */,
  char *vmask /* variable pairs for which a group has been already built */,
  char *lmask /* levels for which a group has already been built*/)
{
    var startV, stopV, startL int32;
    var i, j int32;
    var auxnode *MtrNode  = treenode;

    for auxnode != nil {
        if auxnode.child != nil {
            addMultiplicityGroups(dd,auxnode.child,multiplicity,vmask,lmask);
        }
        /* Build remaining groups. */
        startV = dd.permZ[auxnode.index] / multiplicity;
        startL = auxnode.low / multiplicity;
        stopV = startV + auxnode.size / multiplicity;
        /* Walk down vmask starting at startV and build missing groups. */
        for i, j = startV, startL; i < stopV; i++ {
            if vmask[i] == 0 {
                var node *MtrNode;
                for lmask[j] == 1 { j++; }
                node = Mtr_MakeGroup(auxnode, j * multiplicity, multiplicity,
                                     MTR_FIXED);
                if node == nil {
                    return(0);
                }
                node.index = dd.invpermZ[i * multiplicity];
                vmask[i] = 1;
                lmask[j] = 1;
            }
        }
        auxnode = auxnode.younger;
    }
    return(1);

} /* end of addMultiplicityGroups */


//ABC_NAMESPACE_IMPL_END


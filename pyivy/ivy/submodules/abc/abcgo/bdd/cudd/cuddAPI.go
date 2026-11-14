package cudd

import core "github.com/glycerine/abcgo/bdd/cudd/core"

import (
	"fmt"
	"time"
)

import "os"
import "github.com/glycerine/abcgo/bdd/mtr"

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

// C source revision identifier omitted.

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

/*
*Function********************************************************************

	Synopsis    [Returns a new ADD variable.]

	Description [Creates a new ADD variable.  The new variable has an
	index equal to the largest previous index plus 1.  Returns a
	pointer to the new variable if successful; NULL otherwise.
	An ADD variable differs from a BDD variable because it points to the
	arithmetic zero, instead of having a complement pointer to 1. ]

	SideEffects [None]

	SeeAlso     [Cudd_bddNewVar Cudd_addIthVar Cudd_addConst
	Cudd_addNewVarAtLevel]

*****************************************************************************
*/
func Cudd_addNewVar(dd *DdManager) *DdNode {
	return core.Cudd_addNewVar(dd)
} // Cudd_addNewVar

/*
*Function********************************************************************

	Synopsis    [Returns a new ADD variable at a specified level.]

	Description [Creates a new ADD variable.  The new variable has an
	index equal to the largest previous index plus 1 and is positioned at
	the specified level in the order.  Returns a pointer to the new
	variable if successful; NULL otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_addNewVar Cudd_addIthVar Cudd_bddNewVarAtLevel]

*****************************************************************************
*/
func Cudd_addNewVarAtLevel(dd *DdManager, level int32) *DdNode {
	return core.Cudd_addNewVarAtLevel(dd, int(level))
} // Cudd_addNewVarAtLevel

/*
*Function********************************************************************

	Synopsis    [Returns a new BDD variable.]

	Description [Creates a new BDD variable.  The new variable has an
	index equal to the largest previous index plus 1.  Returns a
	pointer to the new variable if successful; NULL otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_addNewVar Cudd_bddIthVar Cudd_bddNewVarAtLevel]

*****************************************************************************
*/
func Cudd_bddNewVar(dd *DdManager) *DdNode {
	return core.Cudd_bddNewVar(dd)
} // Cudd_bddNewVar

/*
*Function********************************************************************

	Synopsis    [Returns a new BDD variable at a specified level.]

	Description [Creates a new BDD variable.  The new variable has an
	index equal to the largest previous index plus 1 and is positioned at
	the specified level in the order.  Returns a pointer to the new
	variable if successful; NULL otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_bddNewVar Cudd_bddIthVar Cudd_addNewVarAtLevel]

*****************************************************************************
*/
func Cudd_bddNewVarAtLevel(dd *DdManager, level int32) *DdNode {
	return core.Cudd_bddNewVarAtLevel(dd, int(level))
} // Cudd_bddNewVarAtLevel

/*
*Function********************************************************************

	Synopsis    [Returns the ADD variable with index i.]

	Description [Retrieves the ADD variable with index i if it already
	exists, or creates a new ADD variable.  Returns a pointer to the
	variable if successful; NULL otherwise.  An ADD variable differs from
	a BDD variable because it points to the arithmetic zero, instead of
	having a complement pointer to 1. ]

	SideEffects [None]

	SeeAlso     [Cudd_addNewVar Cudd_bddIthVar Cudd_addConst
	Cudd_addNewVarAtLevel]

*****************************************************************************
*/
func Cudd_addIthVar(dd *DdManager, i int32) *DdNode {
	return core.Cudd_addIthVar(dd, int(i))
} // Cudd_addIthVar

/*
*Function********************************************************************

	Synopsis    [Returns the BDD variable with index i.]

	Description [Retrieves the BDD variable with index i if it already
	exists, or creates a new BDD variable.  Returns a pointer to the
	variable if successful; NULL otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_bddNewVar Cudd_addIthVar Cudd_bddNewVarAtLevel
	Cudd_ReadVars]

*****************************************************************************
*/
func Cudd_bddIthVar(dd *DdManager, i int32) *DdNode {
	return core.Cudd_bddIthVar(dd, int(i))
} // Cudd_bddIthVar

/*
*Function********************************************************************

	Synopsis    [Returns the ZDD variable with index i.]

	Description [Retrieves the ZDD variable with index i if it already
	exists, or creates a new ZDD variable.  Returns a pointer to the
	variable if successful; NULL otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_bddIthVar Cudd_addIthVar]

*****************************************************************************
*/
func Cudd_zddIthVar(dd *DdManager, i int32) *DdNode {
	return core.Cudd_zddIthVar(dd, int(i))
} // Cudd_zddIthVar

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_zddVarsFromBddVars(dd *DdManager, multiplicity int32) int32 {
	return int32(core.Cudd_zddVarsFromBddVars(dd, int(multiplicity)))
} // Cudd_zddVarsFromBddVars

/*
*Function********************************************************************

	Synopsis    [Returns the ADD for constant c.]

	Description [Retrieves the ADD for constant c if it already
	exists, or creates a new ADD.  Returns a pointer to the
	ADD if successful; NULL otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_addNewVar Cudd_addIthVar]

*****************************************************************************
*/
func Cudd_addConst(dd *DdManager, c CUDD_VALUE_TYPE) *DdNode {
	return (cuddUniqueConst(dd, c))

} /* end of Cudd_addConst */

/*
*Function********************************************************************

	Synopsis    [Returns 1 if a DD node is not constant.]

	Description [Returns 1 if a DD node is not constant. This function is
	useful to test the results of Cudd_bddIteConstant, Cudd_addIteConstant,
	Cudd_addEvalConst. These results may be a special value signifying
	non-constant. In the other cases the macro Cudd_IsConstant can be used.]

	SideEffects [None]

	SeeAlso     [Cudd_IsConstant Cudd_bddIteConstant Cudd_addIteConstant
	Cudd_addEvalConst]

*****************************************************************************
*/
func Cudd_IsNonConstant(f *DdNode) int32 {
	if core.Cudd_IsNonConstant(f) {
		return 1
	}
	return 0
} // Cudd_IsNonConstant

/*
*Function********************************************************************

	Synopsis    [Enables automatic dynamic reordering of BDDs and ADDs.]

	Description [Enables automatic dynamic reordering of BDDs and
	ADDs. Parameter method is used to determine the method used for
	reordering. If CUDD_REORDER_SAME is passed, the method is
	unchanged.]

	SideEffects [None]

	SeeAlso     [Cudd_AutodynDisable Cudd_ReorderingStatus
	Cudd_AutodynEnableZdd]

*****************************************************************************
*/
func Cudd_AutodynEnable(dd *DdManager, method Cudd_ReorderingType) {
	core.Cudd_AutodynEnable(dd, method)
} // Cudd_AutodynEnable

/*
*Function********************************************************************

	Synopsis    [Disables automatic dynamic reordering.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_AutodynEnable Cudd_ReorderingStatus
	Cudd_AutodynDisableZdd]

*****************************************************************************
*/
func Cudd_AutodynDisable(dd *DdManager) { core.Cudd_AutodynDisable(dd) } // Cudd_AutodynDisable

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_ReorderingStatus(dd *DdManager, method *Cudd_ReorderingType) int32 {
	m, enabled := core.Cudd_ReorderingStatus(dd)
	if method != nil {
		*method = m
	}
	if enabled {
		return 1
	}
	return 0
} // Cudd_ReorderingStatus

/*
*Function********************************************************************

	Synopsis    [Enables automatic dynamic reordering of ZDDs.]

	Description [Enables automatic dynamic reordering of ZDDs. Parameter
	method is used to determine the method used for reordering ZDDs. If
	CUDD_REORDER_SAME is passed, the method is unchanged.]

	SideEffects [None]

	SeeAlso     [Cudd_AutodynDisableZdd Cudd_ReorderingStatusZdd
	Cudd_AutodynEnable]

*****************************************************************************
*/
func Cudd_AutodynEnableZdd(dd *DdManager, method Cudd_ReorderingType) {
	core.Cudd_AutodynEnableZdd(dd, method)
} // Cudd_AutodynEnableZdd

/*
*Function********************************************************************

	Synopsis    [Disables automatic dynamic reordering of ZDDs.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_AutodynEnableZdd Cudd_ReorderingStatusZdd
	Cudd_AutodynDisable]

*****************************************************************************
*/
func Cudd_AutodynDisableZdd(dd *DdManager) { core.Cudd_AutodynDisableZdd(dd) } // Cudd_AutodynDisableZdd

/*
*Function********************************************************************

	Synopsis    [Reports the status of automatic dynamic reordering of ZDDs.]

	Description [Reports the status of automatic dynamic reordering of
	ZDDs. Parameter method is set to the ZDD reordering method currently
	selected. Returns 1 if automatic reordering is enabled; 0
	otherwise.]

	SideEffects [Parameter method is set to the ZDD reordering method currently
	selected.]

	SeeAlso     [Cudd_AutodynEnableZdd Cudd_AutodynDisableZdd
	Cudd_ReorderingStatus]

*****************************************************************************
*/
func Cudd_ReorderingStatusZdd(dd *DdManager, method *Cudd_ReorderingType) int32 {
	m, enabled := core.Cudd_ReorderingStatusZdd(dd)
	if method != nil {
		*method = m
	}
	if enabled {
		return 1
	}
	return 0
} // Cudd_ReorderingStatusZdd

/*
*Function********************************************************************

	Synopsis    [Tells whether the realignment of ZDD order to BDD order is
	enabled.]

	Description [Returns 1 if the realignment of ZDD order to BDD order is
	enabled; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_zddRealignEnable Cudd_zddRealignDisable
	Cudd_bddRealignEnable Cudd_bddRealignDisable]

*****************************************************************************
*/
func Cudd_zddRealignmentEnabled(dd *DdManager) int32 {
	if core.Cudd_zddRealignmentEnabled(dd) {
		return 1
	}
	return 0
} // Cudd_zddRealignmentEnabled

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_zddRealignEnable(dd *DdManager) { core.Cudd_zddRealignEnable(dd) } // Cudd_zddRealignEnable

/*
*Function********************************************************************

	Synopsis    [Disables realignment of ZDD order to BDD order.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_zddRealignEnable Cudd_zddRealignmentEnabled
	Cudd_bddRealignEnable Cudd_bddRealignmentEnabled]

*****************************************************************************
*/
func Cudd_zddRealignDisable(dd *DdManager) { core.Cudd_zddRealignDisable(dd) } // Cudd_zddRealignDisable

/*
*Function********************************************************************

	Synopsis    [Tells whether the realignment of BDD order to ZDD order is
	enabled.]

	Description [Returns 1 if the realignment of BDD order to ZDD order is
	enabled; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_bddRealignEnable Cudd_bddRealignDisable
	Cudd_zddRealignEnable Cudd_zddRealignDisable]

*****************************************************************************
*/
func Cudd_bddRealignmentEnabled(dd *DdManager) int32 {
	if core.Cudd_bddRealignmentEnabled(dd) {
		return 1
	}
	return 0
} // Cudd_bddRealignmentEnabled

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_bddRealignEnable(dd *DdManager) { core.Cudd_bddRealignEnable(dd) } // Cudd_bddRealignEnable

/*
*Function********************************************************************

	Synopsis    [Disables realignment of ZDD order to BDD order.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_bddRealignEnable Cudd_bddRealignmentEnabled
	Cudd_zddRealignEnable Cudd_zddRealignmentEnabled]

*****************************************************************************
*/
func Cudd_bddRealignDisable(dd *DdManager) { core.Cudd_bddRealignDisable(dd) } // Cudd_bddRealignDisable

/*
*Function********************************************************************

	Synopsis    [Returns the one constant of the manager.]

	Description [Returns the one constant of the manager. The one
	constant is common to ADDs and BDDs.]

	SideEffects [None]

	SeeAlso [Cudd_ReadZero Cudd_ReadLogicZero Cudd_ReadZddOne]

*****************************************************************************
*/
func Cudd_ReadOne(dd *DdManager) *DdNode {
	return core.Cudd_ReadOne(dd)

} /* end of Cudd_ReadOne */

/*
*Function********************************************************************

	Synopsis    [Returns the ZDD for the constant 1 function.]

	Description [Returns the ZDD for the constant 1 function.
	The representation of the constant 1 function as a ZDD depends on
	how many variables it (nominally) depends on. The index of the
	topmost variable in the support is given as argument <code>i</code>.]

	SideEffects [None]

	SeeAlso [Cudd_ReadOne]

*****************************************************************************
*/
func Cudd_ReadZddOne(dd *DdManager, i int32) *DdNode {
	return core.Cudd_ReadZddOne(dd, int(i))

} /* end of Cudd_ReadZddOne */

/*
*Function********************************************************************

	Synopsis    [Returns the zero constant of the manager.]

	Description [Returns the zero constant of the manager. The zero
	constant is the arithmetic zero, rather than the logic zero. The
	latter is the complement of the one constant.]

	SideEffects [None]

	SeeAlso [Cudd_ReadOne Cudd_ReadLogicZero]

*****************************************************************************
*/
func Cudd_ReadZero(dd *DdManager) *DdNode { return core.Cudd_ReadZero(dd) } // Cudd_ReadZero

/*
*Function********************************************************************

	Synopsis    [Returns the logic zero constant of the manager.]

	Description [Returns the zero constant of the manager. The logic zero
	constant is the complement of the one constant, and is distinct from
	the arithmetic zero.]

	SideEffects [None]

	SeeAlso [Cudd_ReadOne Cudd_ReadZero]

*****************************************************************************
*/
func Cudd_ReadLogicZero(dd *DdManager) *DdNode { return core.Cudd_ReadLogicZero(dd) } // Cudd_ReadLogicZero

/*
*Function********************************************************************

	Synopsis    [Reads the plus-infinity constant from the manager.]

	Description []

	SideEffects [None]

*****************************************************************************
*/
func Cudd_ReadPlusInfinity(dd *DdManager) *DdNode { return core.Cudd_ReadPlusInfinity(dd) } // Cudd_ReadPlusInfinity

/*
*Function********************************************************************

	Synopsis    [Reads the minus-infinity constant from the manager.]

	Description []

	SideEffects [None]

*****************************************************************************
*/
func Cudd_ReadMinusInfinity(dd *DdManager) *DdNode { return core.Cudd_ReadMinusInfinity(dd) } // Cudd_ReadMinusInfinity

/*
*Function********************************************************************

	Synopsis    [Reads the background constant of the manager.]

	Description []

	SideEffects [None]

*****************************************************************************
*/
func Cudd_ReadBackground(dd *DdManager) *DdNode { return core.Cudd_ReadBackground(dd) } // Cudd_ReadBackground

/*
*Function********************************************************************

	Synopsis    [Sets the background constant of the manager.]

	Description [Sets the background constant of the manager. It assumes
	that the DdNode pointer bck is already referenced.]

	SideEffects [None]

*****************************************************************************
*/
func Cudd_SetBackground(dd *DdManager, background *DdNode) { core.Cudd_SetBackground(dd, background) } // Cudd_SetBackground

/*
*Function********************************************************************

	Synopsis    [Reads the number of slots in the cache.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_ReadCacheUsedSlots]

*****************************************************************************
*/
func Cudd_ReadCacheSlots(dd *DdManager) uint32 { return core.Cudd_ReadCacheSlots(dd) } // Cudd_ReadCacheSlots

/*
*Function********************************************************************

	Synopsis    [Reads the fraction of used slots in the cache.]

	Description [Reads the fraction of used slots in the cache. The unused
	slots are those in which no valid data is stored. Garbage collection,
	variable reordering, and cache resizing may cause used slots to become
	unused.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadCacheSlots]

*****************************************************************************
*/
func Cudd_ReadCacheUsedSlots(dd *DdManager) float64 { return core.Cudd_ReadCacheUsedSlots(dd) } // Cudd_ReadCacheUsedSlots

/*
*Function********************************************************************

	Synopsis    [Returns the number of cache look-ups.]

	Description [Returns the number of cache look-ups.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadCacheHits]

*****************************************************************************
*/
func Cudd_ReadCacheLookUps(dd *DdManager) float64 { return core.Cudd_ReadCacheLookUps(dd) } // Cudd_ReadCacheLookUps

/*
*Function********************************************************************

	Synopsis    [Returns the number of cache hits.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_ReadCacheLookUps]

*****************************************************************************
*/
func Cudd_ReadCacheHits(dd *DdManager) float64 { return core.Cudd_ReadCacheHits(dd) } // Cudd_ReadCacheHits

/*
*Function********************************************************************

	Synopsis    [Returns the number of recursive calls.]

	Description [Returns the number of recursive calls if the package is
	compiled with DD_COUNT defined.]

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
func Cudd_ReadRecursiveCalls(dd *DdManager) float64 { return -1 } // Cudd_ReadRecursiveCalls

/*
*Function********************************************************************

	Synopsis    [Reads the hit rate that causes resizinig of the computed
	table.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_SetMinHit]

*****************************************************************************
*/
func Cudd_ReadMinHit(dd *DdManager) uint32 { return core.Cudd_ReadMinHit(dd) }

/*
*Function********************************************************************

	Synopsis    [Sets the hit rate that causes resizinig of the computed
	table.]

	Description [Sets the minHit parameter of the manager. This
	parameter controls the resizing of the computed table. If the hit
	rate is larger than the specified value, and the cache is not
	already too large, then its size is doubled.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadMinHit]

*****************************************************************************
*/
func Cudd_SetMinHit(dd *DdManager, hitRate uint32) { core.Cudd_SetMinHit(dd, hitRate) }

/*
*Function********************************************************************

	Synopsis    [Reads the looseUpTo parameter of the manager.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_SetLooseUpTo Cudd_ReadMinHit Cudd_ReadMinDead]

*****************************************************************************
*/
func Cudd_ReadLooseUpTo(dd *DdManager) uint32 { return core.Cudd_ReadLooseUpTo(dd) }

/*
*Function********************************************************************

	Synopsis    [Sets the looseUpTo parameter of the manager.]

	Description [Sets the looseUpTo parameter of the manager. This
	parameter of the manager controls the threshold beyond which no fast
	growth of the unique table is allowed. The threshold is given as a
	number of slots. If the value passed to this function is 0, the
	function determines a suitable value based on the available memory.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadLooseUpTo Cudd_SetMinHit]

*****************************************************************************
*/
func Cudd_SetLooseUpTo(dd *DdManager, value uint32) { core.Cudd_SetLooseUpTo(dd, value) }

/*
*Function********************************************************************

	Synopsis    [Returns the soft limit for the cache size.]

	Description [Returns the soft limit for the cache size. The soft limit]

	SideEffects [None]

	SeeAlso     [Cudd_ReadMaxCache]

*****************************************************************************
*/
func Cudd_ReadMaxCache(dd *DdManager) uint32 { return core.Cudd_ReadMaxCache(dd) }

/*
*Function********************************************************************

	Synopsis    [Reads the maxCacheHard parameter of the manager.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_SetMaxCacheHard Cudd_ReadMaxCache]

*****************************************************************************
*/
func Cudd_ReadMaxCacheHard(dd *DdManager) uint32 { return core.Cudd_ReadMaxCacheHard(dd) }

/*
*Function********************************************************************

	Synopsis    [Sets the maxCacheHard parameter of the manager.]

	Description [Sets the maxCacheHard parameter of the manager. The
	cache cannot grow larger than maxCacheHard entries. This parameter
	allows an application to control the trade-off of memory versus
	speed. If the value passed to this function is 0, the function
	determines a suitable maximum cache size based on the available memory.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadMaxCacheHard Cudd_SetMaxCache]

*****************************************************************************
*/
func Cudd_SetMaxCacheHard(dd *DdManager, value uint32) { core.Cudd_SetMaxCacheHard(dd, value) }

/*
*Function********************************************************************

	Synopsis    [Returns the number of BDD variables in existance.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_ReadZddSize]

*****************************************************************************
*/
func Cudd_ReadSize(dd *DdManager) int32 { return int32(core.Cudd_ReadSize(dd)) }

/*
*Function********************************************************************

	Synopsis    [Returns the number of ZDD variables in existance.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_ReadSize]

*****************************************************************************
*/
func Cudd_ReadZddSize(dd *DdManager) int32 { return int32(core.Cudd_ReadZddSize(dd)) }

/*
*Function********************************************************************

	Synopsis    [Returns the total number of slots of the unique table.]

	Description [Returns the total number of slots of the unique table.
	This number ismainly for diagnostic purposes.]

	SideEffects [None]

*****************************************************************************
*/
func Cudd_ReadSlots(dd *DdManager) uint32 { return core.Cudd_ReadSlots(dd) }

/*
*Function********************************************************************

	Synopsis    [Reads the fraction of used slots in the unique table.]

	Description [Reads the fraction of used slots in the unique
	table. The unused slots are those in which no valid data is
	stored. Garbage collection, variable reordering, and subtable
	resizing may cause used slots to become unused.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadSlots]

*****************************************************************************
*/
func Cudd_ReadUsedSlots(dd *DdManager) float64 { return core.Cudd_ReadUsedSlots(dd) }

/*
*Function********************************************************************

	Synopsis    [Computes the expected fraction of used slots in the unique
	table.]

	Description [Computes the fraction of slots in the unique table that
	should be in use. This expected value is based on the assumption
	that the hash function distributes the keys randomly; it can be
	compared with the result of Cudd_ReadUsedSlots to monitor the
	performance of the unique table hash function.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadSlots Cudd_ReadUsedSlots]

*****************************************************************************
*/
func Cudd_ExpectedUsedSlots(dd *DdManager) float64 { return core.Cudd_ExpectedUsedSlots(dd) }

/*
*Function********************************************************************

	Synopsis    [Returns the number of nodes in the unique table.]

	Description [Returns the total number of nodes currently in the unique
	table, including the dead nodes.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadDead]

*****************************************************************************
*/
func Cudd_ReadKeys(dd *DdManager) uint32 { return core.Cudd_ReadKeys(dd) }

/*
*Function********************************************************************

	Synopsis    [Returns the number of dead nodes in the unique table.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_ReadKeys]

*****************************************************************************
*/
func Cudd_ReadDead(dd *DdManager) uint32 { return core.Cudd_ReadDead(dd) }

/*
*Function********************************************************************

	Synopsis    [Reads the minDead parameter of the manager.]

	Description [Reads the minDead parameter of the manager. The minDead
	parameter is used by the package to decide whether to collect garbage
	or resize a subtable of the unique table when the subtable becomes
	too full. The application can indirectly control the value of minDead
	by setting the looseUpTo parameter.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadDead Cudd_ReadLooseUpTo Cudd_SetLooseUpTo]

*****************************************************************************
*/
func Cudd_ReadMinDead(dd *DdManager) uint32 { return core.Cudd_ReadMinDead(dd) }

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_ReadReorderings(dd *DdManager) int32 { return int32(core.Cudd_ReadReorderings(dd)) } /* end of Cudd_ReadReorderings */

/*
*Function********************************************************************

	Synopsis    [Returns the time spent in reordering.]

	Description [Returns the number of milliseconds spent reordering
	variables since the manager was initialized. The time spent in collecting
	garbage before reordering is included.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadReorderings]

*****************************************************************************
*/
func Cudd_ReadReorderingTime(dd *DdManager) int64 { return core.Cudd_ReadReorderingTime(dd) } /* end of Cudd_ReadReorderingTime */

/*
*Function********************************************************************

	Synopsis    [Returns the number of times garbage collection has occurred.]

	Description [Returns the number of times garbage collection has
	occurred in the manager. The number includes both the calls from
	reordering procedures and those caused by requests to create new
	nodes.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadGarbageCollectionTime]

*****************************************************************************
*/
func Cudd_ReadGarbageCollections(dd *DdManager) uint32 { return core.Cudd_ReadGarbageCollections(dd) }

/*
*Function********************************************************************

	Synopsis    [Returns the time spent in garbage collection.]

	Description [Returns the number of milliseconds spent doing garbage
	collection since the manager was initialized.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadGarbageCollections]

*****************************************************************************
*/
func Cudd_ReadGarbageCollectionTime(dd *DdManager) int64 {
	return core.Cudd_ReadGarbageCollectionTime(dd)
}

/*
*Function********************************************************************

	Synopsis    [Returns the number of nodes freed.]

	Description [Returns the number of nodes returned to the free list if the
	keeping of this statistic is enabled; -1 otherwise. This statistic is
	enabled only if the package is compiled with DD_STATS defined.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadNodesDropped]

*****************************************************************************
*/
func Cudd_ReadNodesFreed(dd *DdManager) float64 { return core.Cudd_ReadNodesFreed(dd) }

/*
*Function********************************************************************

	Synopsis    [Returns the number of nodes dropped.]

	Description [Returns the number of nodes killed by dereferencing if the
	keeping of this statistic is enabled; -1 otherwise. This statistic is
	enabled only if the package is compiled with DD_STATS defined.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadNodesFreed]

*****************************************************************************
*/
func Cudd_ReadNodesDropped(dd *DdManager) float64 { return core.Cudd_ReadNodesDropped(dd) }

/*
*Function********************************************************************

	Synopsis    [Returns the number of look-ups in the unique table.]

	Description [Returns the number of look-ups in the unique table if the
	keeping of this statistic is enabled; -1 otherwise. This statistic is
	enabled only if the package is compiled with DD_UNIQUE_PROFILE defined.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadUniqueLinks]

*****************************************************************************
*/
func Cudd_ReadUniqueLookUps(dd *DdManager) float64 { return core.Cudd_ReadUniqueLookUps(dd) }

/*
*Function********************************************************************

	Synopsis    [Returns the number of links followed in the unique table.]

	Description [Returns the number of links followed during look-ups in the
	unique table if the keeping of this statistic is enabled; -1 otherwise.
	If an item is found in the first position of its collision list, the
	number of links followed is taken to be 0. If it is in second position,
	the number of links is 1, and so on. This statistic is enabled only if
	the package is compiled with DD_UNIQUE_PROFILE defined.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadUniqueLookUps]

*****************************************************************************
*/
func Cudd_ReadUniqueLinks(dd *DdManager) float64 { return core.Cudd_ReadUniqueLinks(dd) }

/*
*Function********************************************************************

	Synopsis    [Reads the siftMaxVar parameter of the manager.]

	Description [Reads the siftMaxVar parameter of the manager. This
	parameter gives the maximum number of variables that will be sifted
	for each invocation of sifting.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadSiftMaxSwap Cudd_SetSiftMaxVar]

*****************************************************************************
*/
func Cudd_ReadSiftMaxVar(dd *DdManager) int32 { return int32(core.Cudd_ReadSiftMaxVar(dd)) } /* end of Cudd_ReadSiftMaxVar */

/*
*Function********************************************************************

	Synopsis    [Sets the siftMaxVar parameter of the manager.]

	Description [Sets the siftMaxVar parameter of the manager. This
	parameter gives the maximum number of variables that will be sifted
	for each invocation of sifting.]

	SideEffects [None]

	SeeAlso     [Cudd_SetSiftMaxSwap Cudd_ReadSiftMaxVar]

*****************************************************************************
*/
func Cudd_SetSiftMaxVar(dd *DdManager, n int32) { core.Cudd_SetSiftMaxVar(dd, int(n)) } /* end of Cudd_SetSiftMaxVar */

/*
*Function********************************************************************

	Synopsis    [Reads the siftMaxSwap parameter of the manager.]

	Description [Reads the siftMaxSwap parameter of the manager. This
	parameter gives the maximum number of swaps that will be attempted
	for each invocation of sifting. The real number of swaps may exceed
	the set limit because the package will always complete the sifting
	of the variable that causes the limit to be reached.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadSiftMaxVar Cudd_SetSiftMaxSwap]

*****************************************************************************
*/
func Cudd_ReadSiftMaxSwap(dd *DdManager) int32 { return int32(core.Cudd_ReadSiftMaxSwap(dd)) } /* end of Cudd_ReadSiftMaxSwap */

/*
*Function********************************************************************

	Synopsis    [Sets the siftMaxSwap parameter of the manager.]

	Description [Sets the siftMaxSwap parameter of the manager. This
	parameter gives the maximum number of swaps that will be attempted
	for each invocation of sifting. The real number of swaps may exceed
	the set limit because the package will always complete the sifting
	of the variable that causes the limit to be reached.]

	SideEffects [None]

	SeeAlso     [Cudd_SetSiftMaxVar Cudd_ReadSiftMaxSwap]

*****************************************************************************
*/
func Cudd_SetSiftMaxSwap(dd *DdManager, n int32) { core.Cudd_SetSiftMaxSwap(dd, int(n)) } /* end of Cudd_SetSiftMaxSwap */

/*
*Function********************************************************************

	Synopsis    [Reads the maxGrowth parameter of the manager.]

	Description [Reads the maxGrowth parameter of the manager.  This
	parameter determines how much the number of nodes can grow during
	sifting of a variable.  Overall, sifting never increases the size of
	the decision diagrams.  This parameter only refers to intermediate
	results.  A lower value will speed up sifting, possibly at the
	expense of quality.]

	SideEffects [None]

	SeeAlso     [Cudd_SetMaxGrowth Cudd_ReadMaxGrowthAlternate]

*****************************************************************************
*/
func Cudd_ReadMaxGrowth(dd *DdManager) float64 { return core.Cudd_ReadMaxGrowth(dd) } /* end of Cudd_ReadMaxGrowth */

/*
*Function********************************************************************

	Synopsis    [Sets the maxGrowth parameter of the manager.]

	Description [Sets the maxGrowth parameter of the manager.  This
	parameter determines how much the number of nodes can grow during
	sifting of a variable.  Overall, sifting never increases the size of
	the decision diagrams.  This parameter only refers to intermediate
	results.  A lower value will speed up sifting, possibly at the
	expense of quality.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadMaxGrowth Cudd_SetMaxGrowthAlternate]

*****************************************************************************
*/
func Cudd_SetMaxGrowth(dd *DdManager, growth float64) { core.Cudd_SetMaxGrowth(dd, growth) } /* end of Cudd_SetMaxGrowth */

/*
*Function********************************************************************

	Synopsis    [Reads the maxGrowthAlt parameter of the manager.]

	Description [Reads the maxGrowthAlt parameter of the manager.  This
	parameter is analogous to the maxGrowth paramter, and is used every
	given number of reorderings instead of maxGrowth.  The number of
	reorderings is set with Cudd_SetReorderingCycle.  If the number of
	reorderings is 0 (default) maxGrowthAlt is never used.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadMaxGrowth Cudd_SetMaxGrowthAlternate
	Cudd_SetReorderingCycle Cudd_ReadReorderingCycle]

*****************************************************************************
*/
func Cudd_ReadMaxGrowthAlternate(dd *DdManager) float64 { return core.Cudd_ReadMaxGrowthAlternate(dd) }

/*
*Function********************************************************************

	Synopsis    [Sets the maxGrowthAlt parameter of the manager.]

	Description [Sets the maxGrowthAlt parameter of the manager.  This
	parameter is analogous to the maxGrowth paramter, and is used every
	given number of reorderings instead of maxGrowth.  The number of
	reorderings is set with Cudd_SetReorderingCycle.  If the number of
	reorderings is 0 (default) maxGrowthAlt is never used.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadMaxGrowthAlternate Cudd_SetMaxGrowth
	Cudd_SetReorderingCycle Cudd_ReadReorderingCycle]

*****************************************************************************
*/
func Cudd_SetMaxGrowthAlternate(dd *DdManager, growth float64) {
	core.Cudd_SetMaxGrowthAlternate(dd, growth)
}

/*
*Function********************************************************************

	Synopsis    [Reads the reordCycle parameter of the manager.]

	Description [Reads the reordCycle parameter of the manager.  This
	parameter determines how often the alternate threshold on maximum
	growth is used in reordering.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadMaxGrowthAlternate Cudd_SetMaxGrowthAlternate
	Cudd_SetReorderingCycle]

*****************************************************************************
*/
func Cudd_ReadReorderingCycle(dd *DdManager) int32 {
	return core.Cudd_ReadReorderingCycle(dd)

} /* end of Cudd_ReadReorderingCycle */

/*
*Function********************************************************************

	Synopsis    [Sets the reordCycle parameter of the manager.]

	Description [Sets the reordCycle parameter of the manager.  This
	parameter determines how often the alternate threshold on maximum
	growth is used in reordering.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadMaxGrowthAlternate Cudd_SetMaxGrowthAlternate
	Cudd_ReadReorderingCycle]

*****************************************************************************
*/
func Cudd_SetReorderingCycle(dd *DdManager, cycle int32) {
	core.Cudd_SetReorderingCycle(dd, cycle)

} /* end of Cudd_SetReorderingCycle */

/*
*Function********************************************************************

	Synopsis    [Returns the variable group tree of the manager.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_SetTree Cudd_FreeTree Cudd_ReadZddTree]

*****************************************************************************
*/
func Cudd_ReadTree(dd *DdManager) *mtr.MtrNode { return core.Cudd_ReadTree(dd) }

/*
*Function********************************************************************

	Synopsis    [Sets the variable group tree of the manager.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_FreeTree Cudd_ReadTree Cudd_SetZddTree]

*****************************************************************************
*/
func Cudd_SetTree(dd *DdManager, tree *mtr.MtrNode) { core.Cudd_SetTree(dd, tree) }

/*
*Function********************************************************************

	Synopsis    [Frees the variable group tree of the manager.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_SetTree Cudd_ReadTree Cudd_FreeZddTree]

*****************************************************************************
*/
func Cudd_FreeTree(dd *DdManager) { core.Cudd_FreeTree(dd) }

/*
*Function********************************************************************

	Synopsis    [Returns the variable group tree of the manager.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_SetZddTree Cudd_FreeZddTree Cudd_ReadTree]

*****************************************************************************
*/
func Cudd_ReadZddTree(dd *DdManager) *mtr.MtrNode { return core.Cudd_ReadZddTree(dd) }

/*
*Function********************************************************************

	Synopsis    [Sets the ZDD variable group tree of the manager.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_FreeZddTree Cudd_ReadZddTree Cudd_SetTree]

*****************************************************************************
*/
func Cudd_SetZddTree(dd *DdManager, tree *mtr.MtrNode) { core.Cudd_SetZddTree(dd, tree) }

/*
*Function********************************************************************

	Synopsis    [Frees the variable group tree of the manager.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_SetZddTree Cudd_ReadZddTree Cudd_FreeTree]

*****************************************************************************
*/
func Cudd_FreeZddTree(dd *DdManager) { core.Cudd_FreeZddTree(dd) }

/*
*Function********************************************************************

	Synopsis    [Returns the index of the node.]

	Description [Returns the index of the node. The node pointer can be
	either regular or complemented.]

	SideEffects [None]

	SeeAlso [Cudd_ReadIndex]

*****************************************************************************
*/
func Cudd_NodeReadIndex(node *DdNode) uint32 { return core.Cudd_NodeReadIndex(node) }

/*
*Function********************************************************************

	Synopsis    [Returns the current position of the i-th variable in the
	order.]

	Description [Returns the current position of the i-th variable in
	the order. If the index is CUDD_CONST_INDEX, returns
	CUDD_CONST_INDEX; otherwise, if the index is out of bounds returns
	-1.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadInvPerm Cudd_ReadPermZdd]

*****************************************************************************
*/
func Cudd_ReadPerm(dd *DdManager, i int32) int32 {
	if uint32(i) == CUDD_CONST_INDEX {
		return int32(CUDD_CONST_INDEX)
	}
	if dd == nil {
		return -1
	}
	return int32(core.Cudd_ReadPerm(dd, int(i)))

} /* end of Cudd_ReadPerm */

/*
*Function********************************************************************

	Synopsis    [Returns the current position of the i-th ZDD variable in the
	order.]

	Description [Returns the current position of the i-th ZDD variable
	in the order. If the index is CUDD_CONST_INDEX, returns
	CUDD_CONST_INDEX; otherwise, if the index is out of bounds returns
	-1.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadInvPermZdd Cudd_ReadPerm]

*****************************************************************************
*/
func Cudd_ReadPermZdd(dd *DdManager, i int32) int32 {
	if uint32(i) == CUDD_CONST_INDEX {
		return int32(CUDD_CONST_INDEX)
	}
	if dd == nil {
		return -1
	}
	return int32(core.Cudd_ReadPermZdd(dd, int(i)))

} /* end of Cudd_ReadPermZdd */

/*
*Function********************************************************************

	Synopsis    [Returns the index of the variable currently in the i-th
	position of the order.]

	Description [Returns the index of the variable currently in the i-th
	position of the order. If the index is CUDD_CONST_INDEX, returns
	CUDD_CONST_INDEX; otherwise, if the index is out of bounds returns -1.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadPerm Cudd_ReadInvPermZdd]

*****************************************************************************
*/
func Cudd_ReadInvPerm(dd *DdManager, i int32) int32 {
	if uint32(i) == CUDD_CONST_INDEX {
		return int32(CUDD_CONST_INDEX)
	}
	if dd == nil {
		return -1
	}
	return int32(core.Cudd_ReadInvPerm(dd, int(i)))

} /* end of Cudd_ReadInvPerm */

/*
*Function********************************************************************

	Synopsis    [Returns the index of the ZDD variable currently in the i-th
	position of the order.]

	Description [Returns the index of the ZDD variable currently in the
	i-th position of the order. If the index is CUDD_CONST_INDEX, returns
	CUDD_CONST_INDEX; otherwise, if the index is out of bounds returns -1.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadPerm Cudd_ReadInvPermZdd]

*****************************************************************************
*/
func Cudd_ReadInvPermZdd(dd *DdManager, i int32) int32 {
	if uint32(i) == CUDD_CONST_INDEX {
		return int32(CUDD_CONST_INDEX)
	}
	if dd == nil {
		return -1
	}
	return int32(core.Cudd_ReadInvPermZdd(dd, int(i)))

} /* end of Cudd_ReadInvPermZdd */

/*
*Function********************************************************************

	Synopsis    [Returns the i-th element of the vars array.]

	Description [Returns the i-th element of the vars array if it falls
	within the array bounds; NULL otherwise. If i is the index of an
	existing variable, this function produces the same result as
	Cudd_bddIthVar. However, if the i-th var does not exist yet,
	Cudd_bddIthVar will create it, whereas Cudd_ReadVars will not.]

	SideEffects [None]

	SeeAlso     [Cudd_bddIthVar]

*****************************************************************************
*/
func Cudd_ReadVars(dd *DdManager, i int32) *DdNode {
	if dd == nil || i < 0 || int(i) >= len(dd.Vars) {
		return nil
	}
	return dd.Vars[i]

} /* end of Cudd_ReadVars */

/*
*Function********************************************************************

	Synopsis    [Reads the epsilon parameter of the manager.]

	Description [Reads the epsilon parameter of the manager. The epsilon
	parameter control the comparison between floating point numbers.]

	SideEffects [None]

	SeeAlso     [Cudd_SetEpsilon]

*****************************************************************************
*/
func Cudd_ReadEpsilon(dd *DdManager) CUDD_VALUE_TYPE {
	return core.Cudd_ReadEpsilon(dd)

} /* end of Cudd_ReadEpsilon */

/*
*Function********************************************************************

	Synopsis    [Sets the epsilon parameter of the manager to ep.]

	Description [Sets the epsilon parameter of the manager to ep. The epsilon
	parameter control the comparison between floating point numbers.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadEpsilon]

*****************************************************************************
*/
func Cudd_SetEpsilon(dd *DdManager, ep CUDD_VALUE_TYPE) {
	core.Cudd_SetEpsilon(dd, ep)

} /* end of Cudd_SetEpsilon */

/*
*Function********************************************************************

	Synopsis    [Reads the groupcheck parameter of the manager.]

	Description [Reads the groupcheck parameter of the manager. The
	groupcheck parameter determines the aggregation criterion in group
	sifting.]

	SideEffects [None]

	SeeAlso     [Cudd_SetGroupcheck]

*****************************************************************************
*/
func Cudd_ReadGroupcheck(dd *DdManager) Cudd_AggregationType {
	return core.Cudd_ReadGroupcheck(dd)

} /* end of Cudd_ReadGroupCheck */

/*
*Function********************************************************************

	Synopsis    [Sets the parameter groupcheck of the manager to gc.]

	Description [Sets the parameter groupcheck of the manager to gc. The
	groupcheck parameter determines the aggregation criterion in group
	sifting.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadGroupCheck]

*****************************************************************************
*/
func Cudd_SetGroupcheck(dd *DdManager, gc Cudd_AggregationType) {
	core.Cudd_SetGroupcheck(dd, gc)

} /* end of Cudd_SetGroupcheck */

/*
*Function********************************************************************

	Synopsis    [Tells whether garbage collection is enabled.]

	Description [Returns 1 if garbage collection is enabled; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_EnableGarbageCollection Cudd_DisableGarbageCollection]

*****************************************************************************
*/
func Cudd_GarbageCollectionEnabled(dd *DdManager) int32 {
	return core.Cudd_GarbageCollectionEnabled(dd)

} /* end of Cudd_GarbageCollectionEnabled */

/*
*Function********************************************************************

	Synopsis    [Enables garbage collection.]

	Description [Enables garbage collection. Garbage collection is
	initially enabled. Therefore it is necessary to call this function
	only if garbage collection has been explicitly disabled.]

	SideEffects [None]

	SeeAlso     [Cudd_DisableGarbageCollection Cudd_GarbageCollectionEnabled]

*****************************************************************************
*/
func Cudd_EnableGarbageCollection(dd *DdManager) {
	core.Cudd_EnableGarbageCollection(dd)

} /* end of Cudd_EnableGarbageCollection */

/*
*Function********************************************************************

	Synopsis    [Disables garbage collection.]

	Description [Disables garbage collection. Garbage collection is
	initially enabled. This function may be called to disable it.
	However, garbage collection will still occur when a new node must be
	created and no memory is left, or when garbage collection is required
	for correctness. (E.g., before reordering.)]

	SideEffects [None]

	SeeAlso     [Cudd_EnableGarbageCollection Cudd_GarbageCollectionEnabled]

*****************************************************************************
*/
func Cudd_DisableGarbageCollection(dd *DdManager) {
	core.Cudd_DisableGarbageCollection(dd)

} /* end of Cudd_DisableGarbageCollection */

/*
*Function********************************************************************

	Synopsis    [Tells whether dead nodes are counted towards triggering
	reordering.]

	Description [Tells whether dead nodes are counted towards triggering
	reordering. Returns 1 if dead nodes are counted; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_TurnOnCountDead Cudd_TurnOffCountDead]

*****************************************************************************
*/
func Cudd_DeadAreCounted(dd *DdManager) int32 {
	return core.Cudd_DeadAreCounted(dd)

} /* end of Cudd_DeadAreCounted */

/*
*Function********************************************************************

	Synopsis    [Causes the dead nodes to be counted towards triggering
	reordering.]

	Description [Causes the dead nodes to be counted towards triggering
	reordering. This causes more frequent reorderings. By default dead
	nodes are not counted.]

	SideEffects [Changes the manager.]

	SeeAlso     [Cudd_TurnOffCountDead Cudd_DeadAreCounted]

*****************************************************************************
*/
func Cudd_TurnOnCountDead(dd *DdManager) {
	core.Cudd_TurnOnCountDead(dd)

} /* end of Cudd_TurnOnCountDead */

/*
*Function********************************************************************

	Synopsis    [Causes the dead nodes not to be counted towards triggering
	reordering.]

	Description [Causes the dead nodes not to be counted towards
	triggering reordering. This causes less frequent reorderings. By
	default dead nodes are not counted. Therefore there is no need to
	call this function unless Cudd_TurnOnCountDead has been previously
	called.]

	SideEffects [Changes the manager.]

	SeeAlso     [Cudd_TurnOnCountDead Cudd_DeadAreCounted]

*****************************************************************************
*/
func Cudd_TurnOffCountDead(dd *DdManager) {
	core.Cudd_TurnOffCountDead(dd)

} /* end of Cudd_TurnOffCountDead */

/*
*Function********************************************************************

	Synopsis    [Returns the current value of the recombination parameter used
	in group sifting.]

	Description [Returns the current value of the recombination
	parameter used in group sifting. A larger (positive) value makes the
	aggregation of variables due to the second difference criterion more
	likely. A smaller (negative) value makes aggregation less likely.]

	SideEffects [None]

	SeeAlso     [Cudd_SetRecomb]

*****************************************************************************
*/
func Cudd_ReadRecomb(dd *DdManager) int32 {
	return core.Cudd_ReadRecomb(dd)

} /* end of Cudd_ReadRecomb */

/*
*Function********************************************************************

	Synopsis    [Sets the value of the recombination parameter used in group
	sifting.]

	Description [Sets the value of the recombination parameter used in
	group sifting. A larger (positive) value makes the aggregation of
	variables due to the second difference criterion more likely. A
	smaller (negative) value makes aggregation less likely. The default
	value is 0.]

	SideEffects [Changes the manager.]

	SeeAlso     [Cudd_ReadRecomb]

*****************************************************************************
*/
func Cudd_SetRecomb(dd *DdManager, recomb int32) {
	core.Cudd_SetRecomb(dd, recomb)

} /* end of Cudd_SetRecomb */

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_ReadSymmviolation(dd *DdManager) int32 {
	return core.Cudd_ReadSymmviolation(dd)

} /* end of Cudd_ReadSymmviolation */

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_SetSymmviolation(dd *DdManager, symmviolation int32) {
	core.Cudd_SetSymmviolation(dd, symmviolation)

} /* end of Cudd_SetSymmviolation */

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_ReadArcviolation(dd *DdManager) int32 {
	return core.Cudd_ReadArcviolation(dd)

} /* end of Cudd_ReadArcviolation */

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_SetArcviolation(dd *DdManager, arcviolation int32) {
	core.Cudd_SetArcviolation(dd, arcviolation)

} /* end of Cudd_SetArcviolation */

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_ReadPopulationSize(dd *DdManager) int32 {
	return core.Cudd_ReadPopulationSize(dd)

} /* end of Cudd_ReadPopulationSize */

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_SetPopulationSize(dd *DdManager, populationSize int32) {
	core.Cudd_SetPopulationSize(dd, populationSize)

} /* end of Cudd_SetPopulationSize */

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_ReadNumberXovers(dd *DdManager) int32 {
	return core.Cudd_ReadNumberXovers(dd)

} /* end of Cudd_ReadNumberXovers */

/*
*Function********************************************************************

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

*****************************************************************************
*/
func Cudd_SetNumberXovers(dd *DdManager, numberXovers int32) {
	core.Cudd_SetNumberXovers(dd, numberXovers)

} /* end of Cudd_SetNumberXovers */

/*
*Function********************************************************************

	Synopsis    [Returns the memory in use by the manager measured in bytes.]

	Description []

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
func Cudd_ReadMemoryInUse(dd *DdManager) uint64 { return core.Cudd_ReadMemoryInUse(dd) }

/*
*Function********************************************************************

	Synopsis    [Prints out statistics and settings for a CUDD manager.]

	Description [Prints out statistics and settings for a CUDD manager.
	Returns 1 if successful; 0 otherwise.]

	SideEffects [None]

	SeeAlso     []

*****************************************************************************
*/
func Cudd_PrintInfo(dd *DdManager, w *os.File) int32 {
	if dd == nil || w == nil {
		return 0
	}
	lines := []string{
		"**** CUDD manager information ****\n",
		fmt.Sprintf("BDD variables: %d\n", Cudd_ReadSize(dd)),
		fmt.Sprintf("ZDD variables: %d\n", Cudd_ReadZddSize(dd)),
		fmt.Sprintf("BDD/ADD/ZDD nodes: %d\n", Cudd_ReadNodeCount(dd)),
		fmt.Sprintf("Cache lookups: %.0f\n", Cudd_ReadCacheLookUps(dd)),
		fmt.Sprintf("Cache hits: %.0f\n", Cudd_ReadCacheHits(dd)),
		fmt.Sprintf("Used unique-table slots: %.4f\n", Cudd_ReadUsedSlots(dd)),
		fmt.Sprintf("Reorderings: %d\n", Cudd_ReadReorderings(dd)),
	}
	for _, line := range lines {
		if _, err := fmt.Fprint(w, line); err != nil {
			return 0
		}
	}
	return 1
}

/*
*Function********************************************************************

	Synopsis    [Reports the peak number of nodes.]

	Description [Reports the peak number of nodes. This number includes
	node on the free list. At the peak, the number of nodes on the free
	list is guaranteed to be less than DD_MEM_CHUNK.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadNodeCount Cudd_PrintInfo]

*****************************************************************************
*/
func Cudd_ReadPeakNodeCount(dd *DdManager) uint64 { return core.Cudd_ReadPeakNodeCount(dd) }

/*
*Function********************************************************************

	Synopsis    [Reports the peak number of live nodes.]

	Description [Reports the peak number of live nodes. This count is kept
	only if CUDD is compiled with DD_STATS defined. If DD_STATS is not
	defined, this function returns -1.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadNodeCount Cudd_PrintInfo Cudd_ReadPeakNodeCount]

*****************************************************************************
*/
func Cudd_ReadPeakLiveNodeCount(dd *DdManager) int32 {
	return int32(core.Cudd_ReadPeakLiveNodeCount(dd))
}

/*
*Function********************************************************************

	Synopsis    [Reports the number of nodes in BDDs and ADDs.]

	Description [Reports the number of live nodes in BDDs and ADDs. This
	number does not include the isolated projection functions and the
	unused constants. These nodes that are not counted are not part of
	the DDs manipulated by the application.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadPeakNodeCount Cudd_zddReadNodeCount]

*****************************************************************************
*/
func Cudd_ReadNodeCount(dd *DdManager) int64 { return int64(core.Cudd_ReadNodeCount(dd)) }

/*
*Function********************************************************************

	Synopsis    [Reports the number of nodes in ZDDs.]

	Description [Reports the number of nodes in ZDDs. This
	number always includes the two constants 1 and 0.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadPeakNodeCount Cudd_ReadNodeCount]

*****************************************************************************
*/
func Cudd_zddReadNodeCount(dd *DdManager) int64 { return int64(core.Cudd_zddReadNodeCount(dd)) }

/*
*Function********************************************************************

	Synopsis    [Adds a function to a hook.]

	Description [Adds a function to a hook. A hook is a list of
	application-provided functions called on certain occasions by the
	package. Returns 1 if the function is successfully added; 2 if the
	function was already in the list; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_RemoveHook]

*****************************************************************************
*/
func Cudd_AddHook(dd *DdManager, f DD_HFP, where Cudd_HookType) int32 {
	return core.Cudd_AddHook(dd, f, where)

} /* end of Cudd_AddHook */

/*
*Function********************************************************************

	Synopsis    [Removes a function from a hook.]

	Description [Removes a function from a hook. A hook is a list of
	application-provided functions called on certain occasions by the
	package. Returns 1 if successful; 0 the function was not in the list.]

	SideEffects [None]

	SeeAlso     [Cudd_AddHook]

*****************************************************************************
*/
func Cudd_RemoveHook(dd *DdManager, f DD_HFP, where Cudd_HookType) int32 {
	return core.Cudd_RemoveHook(dd, f, where)

} /* end of Cudd_RemoveHook */

/*
*Function********************************************************************

	Synopsis    [Checks whether a function is in a hook.]

	Description [Checks whether a function is in a hook. A hook is a list of
	application-provided functions called on certain occasions by the
	package. Returns 1 if the function is found; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_AddHook Cudd_RemoveHook]

*****************************************************************************
*/
func Cudd_IsInHook(dd *DdManager, f DD_HFP, where Cudd_HookType) int32 {
	return core.Cudd_IsInHook(dd, f, where)

} /* end of Cudd_IsInHook */

/*
*Function********************************************************************

	Synopsis    [Sample hook function to call before reordering.]

	Description [Sample hook function to call before reordering.
	Prints on the manager's stdout reordering method and initial size.
	Returns 1 if successful; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_StdPostReordHook]

*****************************************************************************
*/
func Cudd_StdPreReordHook(dd *DdManager, graph string, data any) int {
	method, ok := data.(Cudd_ReorderingType)
	if !ok {
		if value, yes := data.(int32); yes {
			method, ok = Cudd_ReorderingType(value), true
		}
	}
	if !ok {
		return 0
	}
	label := map[Cudd_ReorderingType]string{
		CUDD_REORDER_RANDOM: "random", CUDD_REORDER_RANDOM_PIVOT: "random",
		CUDD_REORDER_SIFT: "sifting", CUDD_REORDER_SIFT_CONVERGE: "converging sifting",
		CUDD_REORDER_SYMM_SIFT: "symmetric sifting", CUDD_REORDER_SYMM_SIFT_CONV: "converging symmetric sifting",
		CUDD_REORDER_GROUP_SIFT: "group sifting", CUDD_REORDER_GROUP_SIFT_CONV: "converging group sifting",
		CUDD_REORDER_LAZY_SIFT: "lazy sifting", CUDD_REORDER_ANNEALING: "annealing",
		CUDD_REORDER_GENETIC: "genetic", CUDD_REORDER_LINEAR: "linear sifting",
		CUDD_REORDER_LINEAR_CONVERGE: "converging linear sifting", CUDD_REORDER_EXACT: "exact",
		CUDD_REORDER_WINDOW2: "window", CUDD_REORDER_WINDOW3: "window", CUDD_REORDER_WINDOW4: "window",
		CUDD_REORDER_WINDOW2_CONV: "converging window", CUDD_REORDER_WINDOW3_CONV: "converging window", CUDD_REORDER_WINDOW4_CONV: "converging window",
	}
	name, exists := label[method]
	if !exists {
		return 0
	}
	count := core.Cudd_ReadNodeCount(dd)
	if graph == "ZDD" {
		count = core.Cudd_zddReadNodeCount(dd)
	}
	if _, err := fmt.Fprintf(core.Cudd_Output(dd), "%s reordering with %s: from %d to ... ", graph, name, count); err != nil {
		return 0
	}
	return 1
}

/*
*Function********************************************************************

	Synopsis    [Sample hook function to call after reordering.]

	Description [Sample hook function to call after reordering.
	Prints on the manager's stdout final size and reordering time.
	Returns 1 if successful; 0 otherwise.]

	SideEffects [None]

	SeeAlso     [Cudd_StdPreReordHook]

*****************************************************************************
*/
func Cudd_StdPostReordHook(dd *DdManager, graph string, data any) int {
	elapsed := time.Duration(0)
	switch start := data.(type) {
	case time.Time:
		elapsed = time.Since(start)
	case int64:
		elapsed = time.Duration(time.Now().UnixMilli()-start) * time.Millisecond
	case uint64:
		elapsed = time.Duration(time.Now().UnixMilli()-int64(start)) * time.Millisecond
	}
	count := core.Cudd_ReadNodeCount(dd)
	if graph == "ZDD" {
		count = core.Cudd_zddReadNodeCount(dd)
	}
	if _, err := fmt.Fprintf(core.Cudd_Output(dd), "%d nodes in %.3f sec\n", count, elapsed.Seconds()); err != nil {
		return 0
	}
	return 1
}

/*
*Function********************************************************************

	Synopsis    [Enables reporting of reordering stats.]

	Description [Enables reporting of reordering stats.
	Returns 1 if successful; 0 otherwise.]

	SideEffects [Installs functions in the pre-reordering and post-reordering
	hooks.]

	SeeAlso     [Cudd_DisableReorderingReporting Cudd_ReorderingReporting]

*****************************************************************************
*/
func Cudd_EnableReorderingReporting(dd *DdManager) int32 {
	if !(Cudd_AddHook(dd, Cudd_StdPreReordHook, CUDD_PRE_REORDERING_HOOK) != 0) {
		return (0)
	}
	if !(Cudd_AddHook(dd, Cudd_StdPostReordHook, CUDD_POST_REORDERING_HOOK) != 0) {
		return (0)
	}
	return (1)

} /* end of Cudd_EnableReorderingReporting */

/*
*Function********************************************************************

	Synopsis    [Disables reporting of reordering stats.]

	Description [Disables reporting of reordering stats.
	Returns 1 if successful; 0 otherwise.]

	SideEffects [Removes functions from the pre-reordering and post-reordering
	hooks.]

	SeeAlso     [Cudd_EnableReorderingReporting Cudd_ReorderingReporting]

*****************************************************************************
*/
func Cudd_DisableReorderingReporting(dd *DdManager) int32 {
	if !(Cudd_RemoveHook(dd, Cudd_StdPreReordHook, CUDD_PRE_REORDERING_HOOK) != 0) {
		return (0)
	}
	if !(Cudd_RemoveHook(dd, Cudd_StdPostReordHook, CUDD_POST_REORDERING_HOOK) != 0) {
		return (0)
	}
	return (1)

} /* end of Cudd_DisableReorderingReporting */

/*
*Function********************************************************************

	Synopsis    [Returns 1 if reporting of reordering stats is enabled.]

	Description [Returns 1 if reporting of reordering stats is enabled;
	0 otherwise.]

	SideEffects [none]

	SeeAlso     [Cudd_EnableReorderingReporting Cudd_DisableReorderingReporting]

*****************************************************************************
*/
func Cudd_ReorderingReporting(dd *DdManager) int32 {
	return (Cudd_IsInHook(dd, Cudd_StdPreReordHook, CUDD_PRE_REORDERING_HOOK))

} /* end of Cudd_ReorderingReporting */

/*
*Function********************************************************************

	Synopsis    [Returns the code of the last error.]

	Description [Returns the code of the last error. The error codes are
	defined in cudd.h.]

	SideEffects [None]

	SeeAlso     [Cudd_ClearErrorCode]

*****************************************************************************
*/
func Cudd_ReadErrorCode(dd *DdManager) Cudd_ErrorType {
	return core.Cudd_ReadErrorCode(dd)

} /* end of Cudd_ReadErrorCode */

/*
*Function********************************************************************

	Synopsis    [Clear the error code of a manager.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_ReadErrorCode]

*****************************************************************************
*/
func Cudd_ClearErrorCode(dd *DdManager) {
	core.Cudd_ClearErrorCode(dd)

} /* end of Cudd_ClearErrorCode */

/*
*Function********************************************************************

	Synopsis    [Reads the stdout of a manager.]

	Description [Reads the stdout of a manager. This is the file pointer to
	which messages normally going to stdout are written. It is initialized
	to stdout. Cudd_SetStdout allows the application to redirect it.]

	SideEffects [None]

	SeeAlso     [Cudd_SetStdout Cudd_ReadStderr]

*****************************************************************************
*/
func Cudd_ReadStdout(dd *DdManager) *os.File {
	file, _ := core.Cudd_Output(dd).(*os.File)
	return file

} /* end of Cudd_ReadStdout */

/*
*Function********************************************************************

	Synopsis    [Sets the stdout of a manager.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_ReadStdout Cudd_SetStderr]

*****************************************************************************
*/
func Cudd_SetStdout(dd *DdManager, fp *os.File) {
	core.Cudd_SetOutput(dd, fp, nil)

} /* end of Cudd_SetStdout */

/*
*Function********************************************************************

	Synopsis    [Reads the stderr of a manager.]

	Description [Reads the stderr of a manager. This is the file pointer to
	which messages normally going to stderr are written. It is initialized
	to stderr. Cudd_SetStderr allows the application to redirect it.]

	SideEffects [None]

	SeeAlso     [Cudd_SetStderr Cudd_ReadStdout]

*****************************************************************************
*/
func Cudd_ReadStderr(dd *DdManager) *os.File {
	file, _ := core.Cudd_ErrorOutput(dd).(*os.File)
	return file

} /* end of Cudd_ReadStderr */

/*
*Function********************************************************************

	Synopsis    [Sets the stderr of a manager.]

	Description []

	SideEffects [None]

	SeeAlso     [Cudd_ReadStderr Cudd_SetStdout]

*****************************************************************************
*/
func Cudd_SetStderr(dd *DdManager, fp *os.File) {
	core.Cudd_SetOutput(dd, nil, fp)

} /* end of Cudd_SetStderr */

/*
*Function********************************************************************

	Synopsis    [Returns the threshold for the next dynamic reordering.]

	Description [Returns the threshold for the next dynamic reordering.
	The threshold is in terms of number of nodes and is in effect only
	if reordering is enabled. The count does not include the dead nodes,
	unless the countDead parameter of the manager has been changed from
	its default setting.]

	SideEffects [None]

	SeeAlso     [Cudd_SetNextReordering]

*****************************************************************************
*/
func Cudd_ReadNextReordering(dd *DdManager) uint32 { return core.Cudd_ReadNextReordering(dd) }

/*
*Function********************************************************************

	Synopsis    [Sets the threshold for the next dynamic reordering.]

	Description [Sets the threshold for the next dynamic reordering.
	The threshold is in terms of number of nodes and is in effect only
	if reordering is enabled. The count does not include the dead nodes,
	unless the countDead parameter of the manager has been changed from
	its default setting.]

	SideEffects [None]

	SeeAlso     [Cudd_ReadNextReordering]

*****************************************************************************
*/
func Cudd_SetNextReordering(dd *DdManager, value uint32) { core.Cudd_SetNextReordering(dd, value) }

/*
*Function********************************************************************

	Synopsis    [Reads the number of elementary reordering steps.]

	Description []

	SideEffects [none]

	SeeAlso     []

*****************************************************************************
*/
func Cudd_ReadSwapSteps(dd *DdManager) float64 { return core.Cudd_ReadSwapSteps(dd) }

/*
*Function********************************************************************

	Synopsis    [Reads the maximum allowed number of live nodes.]

	Description [Reads the maximum allowed number of live nodes. When this
	number is exceeded, the package returns NULL.]

	SideEffects [none]

	SeeAlso     [Cudd_SetMaxLive]

*****************************************************************************
*/
func Cudd_ReadMaxLive(dd *DdManager) uint32 { return core.Cudd_ReadMaxLive(dd) }

/*
*Function********************************************************************

	Synopsis    [Sets the maximum allowed number of live nodes.]

	Description [Sets the maximum allowed number of live nodes. When this
	number is exceeded, the package returns NULL.]

	SideEffects [none]

	SeeAlso     [Cudd_ReadMaxLive]

*****************************************************************************
*/
func Cudd_SetMaxLive(dd *DdManager, value uint32) { core.Cudd_SetMaxLive(dd, value) }

/*
*Function********************************************************************

	Synopsis    [Reads the maximum allowed memory.]

	Description [Reads the maximum allowed memory. When this
	number is exceeded, the package returns NULL.]

	SideEffects [none]

	SeeAlso     [Cudd_SetMaxMemory]

*****************************************************************************
*/
func Cudd_ReadMaxMemory(dd *DdManager) uint64 { return core.Cudd_ReadMaxMemory(dd) }

/*
*Function********************************************************************

	Synopsis    [Sets the maximum allowed memory.]

	Description [Sets the maximum allowed memory. When this
	number is exceeded, the package returns NULL.]

	SideEffects [none]

	SeeAlso     [Cudd_ReadMaxMemory]

*****************************************************************************
*/
func Cudd_SetMaxMemory(dd *DdManager, value uint64) { core.Cudd_SetMaxMemory(dd, value) }

/*
*Function********************************************************************

	Synopsis    [Prevents sifting of a variable.]

	Description [This function sets a flag to prevent sifting of a
	variable.  Returns 1 if successful; 0 otherwise (i.e., invalid
	variable index).]

	SideEffects [Changes the "bindVar" flag in DdSubtable.]

	SeeAlso     [Cudd_bddUnbindVar]

*****************************************************************************
*/
func Cudd_bddBindVar(dd *DdManager, index int32) int32 {
	return int32(core.Cudd_bddBindVar(dd, int(index)))
}

/*
*Function********************************************************************

	Synopsis    [Allows the sifting of a variable.]

	Description [This function resets the flag that prevents the sifting
	of a variable. In successive variable reorderings, the variable will
	NOT be skipped, that is, sifted.  Initially all variables can be
	sifted. It is necessary to call this function only to re-enable
	sifting after a call to Cudd_bddBindVar. Returns 1 if successful; 0
	otherwise (i.e., invalid variable index).]

	SideEffects [Changes the "bindVar" flag in DdSubtable.]

	SeeAlso     [Cudd_bddBindVar]

*****************************************************************************
*/
func Cudd_bddUnbindVar(dd *DdManager, index int32) int32 {
	return int32(core.Cudd_bddUnbindVar(dd, int(index)))
}

/*
*Function********************************************************************

	Synopsis    [Tells whether a variable can be sifted.]

	Description [This function returns 1 if a variable is enabled for
	sifting.  Initially all variables can be sifted. This function returns
	0 only if there has been a previous call to Cudd_bddBindVar for that
	variable not followed by a call to Cudd_bddUnbindVar. The function returns
	0 also in the case in which the index of the variable is out of bounds.]

	SideEffects [none]

	SeeAlso     [Cudd_bddBindVar Cudd_bddUnbindVar]

*****************************************************************************
*/
func Cudd_bddVarIsBound(dd *DdManager, index int32) int32 {
	return int32(core.Cudd_bddVarIsBound(dd, int(index)))
}

/*
*Function********************************************************************

	Synopsis    [Sets a variable type to primary input.]

	Description [Sets a variable type to primary input.  The variable type is
	used by lazy sifting.  Returns 1 if successful; 0 otherwise.]

	SideEffects [modifies the manager]

	SeeAlso     [Cudd_bddSetPsVar Cudd_bddSetNsVar Cudd_bddIsPiVar]

*****************************************************************************
*/
func Cudd_bddSetPiVar(dd *DdManager, index int32) int32 {
	return int32(core.Cudd_bddSetVariableType(dd, int(index), CUDD_VAR_PRIMARY_INPUT))
}

/*
*Function********************************************************************

	Synopsis    [Sets a variable type to present state.]

	Description [Sets a variable type to present state.  The variable type is
	used by lazy sifting.  Returns 1 if successful; 0 otherwise.]

	SideEffects [modifies the manager]

	SeeAlso     [Cudd_bddSetPiVar Cudd_bddSetNsVar Cudd_bddIsPsVar]

*****************************************************************************
*/
func Cudd_bddSetPsVar(dd *DdManager, index int32) int32 {
	return int32(core.Cudd_bddSetVariableType(dd, int(index), CUDD_VAR_PRESENT_STATE))
}

/*
*Function********************************************************************

	Synopsis    [Sets a variable type to next state.]

	Description [Sets a variable type to next state.  The variable type is
	used by lazy sifting.  Returns 1 if successful; 0 otherwise.]

	SideEffects [modifies the manager]

	SeeAlso     [Cudd_bddSetPiVar Cudd_bddSetPsVar Cudd_bddIsNsVar]

*****************************************************************************
*/
func Cudd_bddSetNsVar(dd *DdManager, index int32) int32 {
	return int32(core.Cudd_bddSetVariableType(dd, int(index), CUDD_VAR_NEXT_STATE))
}

/*
*Function********************************************************************

	Synopsis    [Checks whether a variable is primary input.]

	Description [Checks whether a variable is primary input.  Returns 1 if
	the variable's type is primary input; 0 if the variable exists but is
	not a primary input; -1 if the variable does not exist.]

	SideEffects [none]

	SeeAlso     [Cudd_bddSetPiVar Cudd_bddIsPsVar Cudd_bddIsNsVar]

*****************************************************************************
*/
func Cudd_bddIsPiVar(dd *DdManager, index int32) int32 {
	if core.Cudd_bddVariableType(dd, int(index)) < 0 {
		return -1
	}
	if core.Cudd_bddVariableType(dd, int(index)) == int(CUDD_VAR_PRIMARY_INPUT) {
		return 1
	}
	return 0
}

/*
*Function********************************************************************

	Synopsis    [Checks whether a variable is present state.]

	Description [Checks whether a variable is present state.  Returns 1 if
	the variable's type is present state; 0 if the variable exists but is
	not a present state; -1 if the variable does not exist.]

	SideEffects [none]

	SeeAlso     [Cudd_bddSetPsVar Cudd_bddIsPiVar Cudd_bddIsNsVar]

*****************************************************************************
*/
func Cudd_bddIsPsVar(dd *DdManager, index int32) int32 {
	if core.Cudd_bddVariableType(dd, int(index)) < 0 {
		return -1
	}
	if core.Cudd_bddVariableType(dd, int(index)) == int(CUDD_VAR_PRESENT_STATE) {
		return 1
	}
	return 0
}

/*
*Function********************************************************************

	Synopsis    [Checks whether a variable is next state.]

	Description [Checks whether a variable is next state.  Returns 1 if
	the variable's type is present state; 0 if the variable exists but is
	not a present state; -1 if the variable does not exist.]

	SideEffects [none]

	SeeAlso     [Cudd_bddSetNsVar Cudd_bddIsPiVar Cudd_bddIsPsVar]

*****************************************************************************
*/
func Cudd_bddIsNsVar(dd *DdManager, index int32) int32 {
	if core.Cudd_bddVariableType(dd, int(index)) < 0 {
		return -1
	}
	if core.Cudd_bddVariableType(dd, int(index)) == int(CUDD_VAR_NEXT_STATE) {
		return 1
	}
	return 0
}

/*
*Function********************************************************************

	Synopsis    [Sets a corresponding pair index for a given index.]

	Description [Sets a corresponding pair index for a given index.
	These pair indices are present and next state variable.  Returns 1 if
	successful; 0 otherwise.]

	SideEffects [modifies the manager]

	SeeAlso     [Cudd_bddReadPairIndex]

*****************************************************************************
*/
func Cudd_bddSetPairIndex(dd *DdManager, index, pairIndex int32) int32 {
	return int32(core.Cudd_bddSetPairIndex(dd, int(index), int(pairIndex)))
}

/*
*Function********************************************************************

	Synopsis    [Reads a corresponding pair index for a given index.]

	Description [Reads a corresponding pair index for a given index.
	These pair indices are present and next state variable.  Returns the
	corresponding variable index if the variable exists; -1 otherwise.]

	SideEffects [modifies the manager]

	SeeAlso     [Cudd_bddSetPairIndex]

*****************************************************************************
*/
func Cudd_bddReadPairIndex(dd *DdManager, index int32) int32 {
	return int32(core.Cudd_bddReadPairIndex(dd, int(index)))
}

/*
*Function********************************************************************

	Synopsis    [Sets a variable to be grouped.]

	Description [Sets a variable to be grouped. This function is used for
	lazy sifting.  Returns 1 if successful; 0 otherwise.]

	SideEffects [modifies the manager]

	SeeAlso     [Cudd_bddSetVarHardGroup Cudd_bddResetVarToBeGrouped]

*****************************************************************************
*/
func Cudd_bddSetVarToBeGrouped(dd *DdManager, index int32) int32 {
	group := core.Cudd_bddLazyGroup(dd, int(index))
	if group < 0 {
		return 0
	}
	if group <= int(CUDD_LAZY_SOFT_GROUP) {
		return int32(core.Cudd_bddSetLazyGroup(dd, int(index), CUDD_LAZY_SOFT_GROUP))
	}
	return 1
}

/*
*Function********************************************************************

	Synopsis    [Sets a variable to be a hard group.]

	Description [Sets a variable to be a hard group.  This function is used
	for lazy sifting.  Returns 1 if successful; 0 otherwise.]

	SideEffects [modifies the manager]

	SeeAlso     [Cudd_bddSetVarToBeGrouped Cudd_bddResetVarToBeGrouped
	Cudd_bddIsVarHardGroup]

*****************************************************************************
*/
func Cudd_bddSetVarHardGroup(dd *DdManager, index int32) int32 {
	return int32(core.Cudd_bddSetLazyGroup(dd, int(index), CUDD_LAZY_HARD_GROUP))
}

/*
*Function********************************************************************

	Synopsis    [Resets a variable not to be grouped.]

	Description [Resets a variable not to be grouped.  This function is
	used for lazy sifting.  Returns 1 if successful; 0 otherwise.]

	SideEffects [modifies the manager]

	SeeAlso     [Cudd_bddSetVarToBeGrouped Cudd_bddSetVarHardGroup]

*****************************************************************************
*/
func Cudd_bddResetVarToBeGrouped(dd *DdManager, index int32) int32 {
	group := core.Cudd_bddLazyGroup(dd, int(index))
	if group < 0 {
		return 0
	}
	if group <= int(CUDD_LAZY_SOFT_GROUP) {
		return int32(core.Cudd_bddSetLazyGroup(dd, int(index), CUDD_LAZY_NONE))
	}
	return 1
}

/*
*Function********************************************************************

	Synopsis    [Checks whether a variable is set to be grouped.]

	Description [Checks whether a variable is set to be grouped. This
	function is used for lazy sifting.]

	SideEffects [none]

	SeeAlso     []

*****************************************************************************
*/
func Cudd_bddIsVarToBeGrouped(dd *DdManager, index int32) int32 {
	group := core.Cudd_bddLazyGroup(dd, int(index))
	if group < 0 {
		return -1
	}
	if group == int(CUDD_LAZY_UNGROUP) {
		return 0
	}
	return int32(group)
}

/*
*Function********************************************************************

	Synopsis    [Sets a variable to be ungrouped.]

	Description [Sets a variable to be ungrouped. This function is used
	for lazy sifting.  Returns 1 if successful; 0 otherwise.]

	SideEffects [modifies the manager]

	SeeAlso     [Cudd_bddIsVarToBeUngrouped]

*****************************************************************************
*/
func Cudd_bddSetVarToBeUngrouped(dd *DdManager, index int32) int32 {
	return int32(core.Cudd_bddSetLazyGroup(dd, int(index), CUDD_LAZY_UNGROUP))
}

/*
*Function********************************************************************

	Synopsis    [Checks whether a variable is set to be ungrouped.]

	Description [Checks whether a variable is set to be ungrouped. This
	function is used for lazy sifting.  Returns 1 if the variable is marked
	to be ungrouped; 0 if the variable exists, but it is not marked to be
	ungrouped; -1 if the variable does not exist.]

	SideEffects [none]

	SeeAlso     [Cudd_bddSetVarToBeUngrouped]

*****************************************************************************
*/
func Cudd_bddIsVarToBeUngrouped(dd *DdManager, index int32) int32 {
	group := core.Cudd_bddLazyGroup(dd, int(index))
	if group < 0 {
		return -1
	}
	if group == int(CUDD_LAZY_UNGROUP) {
		return 1
	}
	return 0
}

/*
*Function********************************************************************

	Synopsis    [Checks whether a variable is set to be in a hard group.]

	Description [Checks whether a variable is set to be in a hard group.  This
	function is used for lazy sifting.  Returns 1 if the variable is marked
	to be in a hard group; 0 if the variable exists, but it is not marked to be
	in a hard group; -1 if the variable does not exist.]

	SideEffects [none]

	SeeAlso     [Cudd_bddSetVarHardGroup]

*****************************************************************************
*/
func Cudd_bddIsVarHardGroup(dd *DdManager, index int32) int32 {
	group := core.Cudd_bddLazyGroup(dd, int(index))
	if group < 0 {
		return -1
	}
	if group == int(CUDD_LAZY_HARD_GROUP) {
		return 1
	}
	return 0
}

/*
*Function********************************************************************

	Synopsis    [Fixes a variable group tree.]

	Description []

	SideEffects [Changes the variable group tree.]

	SeeAlso     []

*****************************************************************************
*/
func fixVarTree(tree *mtr.MtrNode, permutation []int) {
	for node := tree; node != nil; node = node.Younger {
		index := int(node.Low)
		node.Index = node.Low
		if index < len(permutation) {
			node.Low = uint32(permutation[index])
		}
		if node.Child != nil {
			fixVarTree(node.Child, permutation)
		}
	}
}

/*
*Function********************************************************************

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

*****************************************************************************
*/
func addMultiplicityGroups(dd *DdManager, tree *mtr.MtrNode, multiplicity int, variableMask, levelMask []byte) int32 {
	if dd == nil || tree == nil || multiplicity < 1 {
		return 0
	}
	var visit func(*mtr.MtrNode) int32
	visit = func(node *mtr.MtrNode) int32 {
		for ; node != nil; node = node.Younger {
			if node.Child != nil && visit(node.Child) == 0 {
				return 0
			}
			if int(node.Index) >= len(dd.PermZ) || int(node.Low) < 0 {
				continue
			}
			startVariable := dd.PermZ[node.Index] / multiplicity
			startLevel := int(node.Low) / multiplicity
			stopVariable := startVariable + int(node.Size)/multiplicity
			for variable, level := startVariable, startLevel; variable < stopVariable; variable++ {
				if variable < 0 || variable >= len(variableMask) {
					return 0
				}
				if variableMask[variable] != 0 {
					continue
				}
				for level < len(levelMask) && levelMask[level] != 0 {
					level++
				}
				if level >= len(levelMask) {
					return 0
				}
				group := mtr.Mtr_MakeGroup(node, uint32(level*multiplicity), uint32(multiplicity), mtr.MTR_FIXED)
				if group == nil {
					return 0
				}
				index := variable * multiplicity
				if index < 0 || index >= len(dd.InvPermZ) {
					return 0
				}
				group.Index = uint32(dd.InvPermZ[index])
				variableMask[variable], levelMask[level] = 1, 1
			}
		}
		return 1
	}
	return visit(tree)
}

//ABC_NAMESPACE_IMPL_END

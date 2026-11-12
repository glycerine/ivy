/**CFile***********************************************************************

  FileName    [cuddInit.c]

  PackageName [cudd]

  Synopsis    [Functions to initialize and shut down the DD manager.]

  Description [External procedures included in this module:
                <ul>
                <li> Cudd_Init()
                <li> Cudd_Quit()
                </ul>
               Internal procedures included in this module:
                <ul>
                <li> cuddZddInitUniv()
                <li> cuddZddFreeUniv()
                </ul>
              ]

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

package core

import "github.com/glycerine/abcgo/bdd/mtr"

// Cudd_Init creates BDD variables and the ZDD universe. Go maps replace the
// original raw-memory pools; sizing arguments remain allocation hints.
func Cudd_Init(numVars, numVarsZ, numSlots, cacheSize uint32, maxMemory uint64) *DdManager {
	if numVars >= CUDD_MAXINDEX || numVarsZ >= CUDD_MAXINDEX {
		return nil
	}
	dd := NewBDDManager(int(numVars))
	if dd == nil {
		return nil
	}
	dd.ZddSize = int(numVarsZ)
	dd.PermZ = make([]int, dd.ZddSize)
	dd.InvPermZ = make([]int, dd.ZddSize)
	for i := 0; i < dd.ZddSize; i++ {
		dd.PermZ[i] = i
		dd.InvPermZ[i] = i
	}
	cuddZddInitUniv(dd)
	return dd
}
func Cudd_Quit(dd *DdManager) {
	if dd == nil {
		return
	}
	dd.closed = true
	clear(dd.unique)
	clear(dd.cache)
	clear(dd.zddUnique)
	clear(dd.addUnique)
	clear(dd.addCache)
	clear(dd.constants)
	dd.Vars = nil
	dd.Perm = nil
	dd.InvPerm = nil
	dd.PermZ = nil
	dd.InvPermZ = nil
	dd.Univ = nil
}
func cuddZddInitUniv(dd *DdManager) int {
	if dd == nil || dd.closed {
		return 0
	}
	if dd.ZddSize < 0 || len(dd.PermZ) != dd.ZddSize || len(dd.InvPermZ) != dd.ZddSize {
		return 0
	}
	for level, index := range dd.InvPermZ {
		if index < 0 || index >= dd.ZddSize || dd.PermZ[index] != level {
			return 0
		}
	}
	dd.Univ = make([]*DdNode, dd.ZddSize)
	result := dd.One
	for i := dd.ZddSize - 1; i >= 0; i-- {
		result = dd.zddNode(uint32(dd.InvPermZ[i]), result, result)
		dd.Univ[i] = result
	}
	return 1
}
func cuddZddFreeUniv(dd *DdManager) {
	if dd != nil {
		dd.Univ = nil
	}
}

// InitializeZDDUniverse reconstructs the universal set after a variable-order
// update. It is exported for the remaining CUDD reordering modules' integration.
func InitializeZDDUniverse(dd *DdManager) int { return cuddZddInitUniv(dd) }
func FreeZDDUniverse(dd *DdManager)           { cuddZddFreeUniv(dd) }

// Cudd_zddVarsFromBddVars aligns each BDD variable with multiplicity adjacent
// ZDD variables and rebuilds live ZDD handles when the resulting order changes.
func Cudd_zddVarsFromBddVars(dd *DdManager, multiplicity int) int {
	if dd == nil || dd.closed || multiplicity < 1 || len(dd.Vars) > int(CUDD_MAXINDEX)/multiplicity {
		return 0
	}
	target := len(dd.Vars) * multiplicity
	if dd.ZddSize > target {
		target = dd.ZddSize
	}
	order := make([]int, target)
	for level, index := range dd.InvPerm {
		for offset := 0; offset < multiplicity; offset++ {
			order[level*multiplicity+offset] = index*multiplicity + offset
		}
	}
	for level := len(dd.Vars) * multiplicity; level < target; level++ {
		order[level] = level
	}
	permutation := make([]int, target)
	for level, index := range order {
		permutation[index] = level
	}
	if !reorderZdd(dd, target, permutation, order) {
		return 0
	}
	var newTree *mtr.MtrNode
	if dd.Tree != nil {
		newTree = mtr.Mtr_CopyTree(dd.Tree, multiplicity)
		if newTree == nil {
			return 0
		}
	} else if multiplicity > 1 && target > 0 {
		newTree = mtr.Mtr_InitGroupTree(0, target)
	}
	if newTree != nil {
		newTree.Index = uint32(dd.InvPermZ[0])
		if multiplicity > 1 {
			for index := range dd.Vars {
				low := dd.Perm[index] * multiplicity
				group := mtr.Mtr_MakeGroup(newTree, uint32(low), uint32(multiplicity), mtr.MTR_FIXED)
				if group == nil {
					mtr.Mtr_FreeTree(newTree)
					return 0
				}
				group.Index = uint32(dd.InvPermZ[low])
			}
		}
	}
	mtr.Mtr_FreeTree(dd.TreeZ)
	dd.TreeZ = newTree
	return 1
}

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

package core

// addPositiveCube validates the ADD product-of-positive-literals convention.
func addPositiveCube(dd *DdManager, cube *DdNode) bool {
	if !dd.validAdd(cube) {
		return false
	}
	for cube != dd.One {
		if Cudd_IsConstant(cube) || cube.E != dd.Zero {
			return false
		}
		cube = cube.T
	}
	return true
}

type addAbstractKey struct{ function, cube *DdNode }
type addAbstractMode uint8

const (
	addAbstractSum addAbstractMode = iota
	addAbstractProduct
	addAbstractOr
)

func addAbstract(dd *DdManager, f, cube *DdNode, mode addAbstractMode) *DdNode {
	if !dd.validAdd(f) || !addPositiveCube(dd, cube) {
		return nil
	}
	memo := make(map[addAbstractKey]*DdNode)
	var visit func(*DdNode, *DdNode) *DdNode
	visit = func(f, cube *DdNode) *DdNode {
		if cube == dd.One {
			return f
		}
		if mode == addAbstractSum && f == dd.Zero || mode == addAbstractProduct && (f == dd.Zero || f == dd.One) || mode == addAbstractOr && Cudd_IsConstant(f) {
			return f
		}
		key := addAbstractKey{f, cube}
		if r := memo[key]; r != nil {
			return r
		}
		var result *DdNode
		if dd.level(f) > dd.level(cube) {
			r := visit(f, cube.T)
			switch mode {
			case addAbstractSum:
				result = Cudd_addApply(dd, Cudd_addTimes, r, Cudd_addConst(dd, 2))
			case addAbstractProduct:
				result = Cudd_addApply(dd, Cudd_addTimes, r, r)
			case addAbstractOr:
				result = r
			}
		} else if f.Index == cube.Index {
			t := visit(f.T, cube.T)
			if mode == addAbstractOr && t == dd.One {
				result = t
			} else {
				e := visit(f.E, cube.T)
				switch mode {
				case addAbstractSum:
					result = Cudd_addApply(dd, Cudd_addPlus, t, e)
				case addAbstractProduct:
					result = Cudd_addApply(dd, Cudd_addTimes, t, e)
				case addAbstractOr:
					result = Cudd_addApply(dd, Cudd_addOr, t, e)
				}
			}
		} else {
			result = dd.addNode(f.Index, visit(f.T, cube), visit(f.E, cube))
		}
		memo[key] = result
		return result
	}
	return visit(f, cube)
}

// Cudd_addExistAbstract sums over all assignments of the cube variables,
// including variables absent from the function (which multiply its value by 2).
func Cudd_addExistAbstract(dd *DdManager, f, cube *DdNode) *DdNode {
	return addAbstract(dd, f, cube, addAbstractSum)
}

// Cudd_addUnivAbstract multiplies over all assignments of the cube variables.
func Cudd_addUnivAbstract(dd *DdManager, f, cube *DdNode) *DdNode {
	return addAbstract(dd, f, cube, addAbstractProduct)
}

// Cudd_addOrAbstract disjunctively abstracts a 0/1 ADD.
func Cudd_addOrAbstract(dd *DdManager, f, cube *DdNode) *DdNode {
	return addAbstract(dd, f, cube, addAbstractOr)
}

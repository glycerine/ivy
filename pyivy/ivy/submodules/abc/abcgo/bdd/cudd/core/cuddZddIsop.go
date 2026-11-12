/**CFile***********************************************************************

  FileName    [cuddZddIsop.c]

  PackageName [cudd]

  Synopsis    [Functions to find irredundant SOP covers as ZDDs from BDDs.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_bddIsop()
                    <li> Cudd_zddIsop()
                    <li> Cudd_MakeBddFromZddCover()
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    <li> cuddBddIsop()
                    <li> cuddZddIsop()
                    <li> cuddMakeBddFromZddCover()
                    </ul>
               Static procedures included in this module:
                    <ul>
                    </ul>
              ]

  SeeAlso     []

  Author      [In-Ho Moon]

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

// isopResult carries both results of the original Minato lower/upper-bound
// recursion. A BDD-only invocation never constructs or grows ZDDs.
type isopResult struct {
	function, cover *DdNode
	cubes           int
}

func (dd *DdManager) isop(lower, upper *DdNode, wantCover bool) isopResult {
	return dd.isopBounded(lower, upper, wantCover, -1)
}

func (dd *DdManager) isopBounded(lower, upper *DdNode, wantCover bool, limit int) isopResult {
	memo := map[[2]*DdNode]isopResult{}
	var visit func(*DdNode, *DdNode) isopResult
	visit = func(lower, upper *DdNode) isopResult {
		if lower == dd.LogicZero {
			r := isopResult{function: dd.LogicZero}
			if wantCover {
				r.cover = dd.Zero
			}
			return r
		}
		if upper == dd.One {
			if limit == 0 {
				return isopResult{}
			}
			r := isopResult{function: dd.One, cubes: 1}
			if wantCover {
				r.cover = dd.One
			}
			return r
		}
		key := [2]*DdNode{lower, upper}
		if r, ok := memo[key]; ok {
			return r
		}
		level := min(dd.level(lower), dd.level(upper))
		index := dd.InvPerm[level]
		lt, le := dd.cofactors(lower, level)
		ut, ue := dd.cofactors(upper, level)
		negative := visit(Cudd_bddAnd(dd, le, Cudd_Not(ut)), ue)
		if negative.function == nil {
			return isopResult{}
		}
		positive := visit(Cudd_bddAnd(dd, lt, Cudd_Not(ue)), ut)
		if positive.function == nil {
			return isopResult{}
		}
		remaining := Cudd_bddOr(dd, Cudd_bddAnd(dd, le, Cudd_Not(negative.function)), Cudd_bddAnd(dd, lt, Cudd_Not(positive.function)))
		common := visit(remaining, Cudd_bddAnd(dd, ue, ut))
		if common.function == nil {
			return isopResult{}
		}
		count := 0
		if limit >= 0 {
			if negative.cubes > limit-positive.cubes {
				return isopResult{}
			}
			count = negative.cubes + positive.cubes
			if common.cubes > limit-count {
				return isopResult{}
			}
			count += common.cubes
		}
		chosen := Cudd_bddOr(dd, Cudd_bddIte(dd, dd.Vars[index], positive.function, negative.function), common.function)
		result := isopResult{function: chosen, cubes: count}
		if wantCover {
			cover := common.cover
			if negative.cover != dd.Zero || positive.cover != dd.Zero {
				lastLiteral := uint64(index)*2 + 1
				if lastLiteral >= uint64(CUDD_MAXINDEX-1) || !dd.ensureZddVar(int(lastLiteral)) {
					return isopResult{}
				}
				if !dd.literalPairOrder(index*2, false) {
					return isopResult{}
				}
			}
			// Child covers no longer contain this Boolean variable. Change therefore
			// inserts its literal, and Union provides the original order-independent
			// GetNodeIVO behavior even when BDD and ZDD level orders differ.
			if negative.cover != dd.Zero {
				cover = Cudd_zddUnion(dd, Cudd_zddChange(dd, negative.cover, index*2+1), cover)
			}
			if positive.cover != dd.Zero {
				cover = Cudd_zddUnion(dd, Cudd_zddChange(dd, positive.cover, index*2), cover)
			}
			result.cover = cover
		}
		memo[key] = result
		return result
	}
	return visit(lower, upper)
}
func Cudd_zddIsop(dd *DdManager, lower, upper *DdNode, cover **DdNode) *DdNode {
	if cover == nil || !dd.valid(lower, upper) || !Cudd_bddLeq(dd, lower, upper) {
		return nil
	}
	result := dd.isop(lower, upper, true)
	if result.function != nil {
		*cover = result.cover
	}
	return result.function
}
func Cudd_bddIsop(dd *DdManager, lower, upper *DdNode) *DdNode {
	if !dd.valid(lower, upper) || !Cudd_bddLeq(dd, lower, upper) {
		return nil
	}
	return dd.isop(lower, upper, false).function
}
func Cudd_MakeBddFromZddCover(dd *DdManager, cover *DdNode) *DdNode {
	if !dd.coverDomain(false, cover) {
		return nil
	}
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(cover *DdNode) *DdNode {
		if cover == dd.One {
			return dd.One
		}
		if cover == dd.Zero {
			return dd.LogicZero
		}
		if n := memo[cover]; n != nil {
			return n
		}
		positive, negative, absent := dd.coverCofactors(cover, int(cover.Index))
		p, n, d := visit(positive), visit(negative), visit(absent)
		variable := Cudd_bddIthVar(dd, int(cover.Index)/2)
		result := Cudd_bddIte(dd, variable, Cudd_bddOr(dd, p, d), Cudd_bddOr(dd, n, d))
		memo[cover] = result
		return result
	}
	return visit(cover)
}

// Complement computes logical cover complementation through the original
// cover→BDD→ISOP pipeline, not set-family difference.
func Cudd_zddComplement(dd *DdManager, cover *DdNode) *DdNode {
	function := Cudd_MakeBddFromZddCover(dd, cover)
	if function == nil {
		return nil
	}
	var result *DdNode
	if Cudd_zddIsop(dd, Cudd_Not(function), Cudd_Not(function), &result) == nil {
		return nil
	}
	return result
}

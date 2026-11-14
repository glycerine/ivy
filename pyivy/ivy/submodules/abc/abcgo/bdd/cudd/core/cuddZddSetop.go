/**CFile***********************************************************************

  FileName    [cuddZddSetop.c]

  PackageName [cudd]

  Synopsis    [Set operations on ZDDs.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_zddIte()
                    <li> Cudd_zddUnion()
                    <li> Cudd_zddIntersect()
                    <li> Cudd_zddDiff()
                    <li> Cudd_zddDiffConst()
                    <li> Cudd_zddSubset1()
                    <li> Cudd_zddSubset0()
                    <li> Cudd_zddChange()
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    <li> cuddZddIte()
                    <li> cuddZddUnion()
                    <li> cuddZddIntersect()
                    <li> cuddZddDiff()
                    <li> cuddZddChangeAux()
                    <li> cuddZddSubset1()
                    <li> cuddZddSubset0()
                    </ul>
               Static procedures included in this module:
                    <ul>
                    <li> zdd_subset1_aux()
                    <li> zdd_subset0_aux()
                    <li> zddVarToConst()
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

package core

import "math"

func (dd *DdManager) validZdd(nodes ...*DdNode) bool {
	if dd == nil || dd.closed {
		return false
	}
	for _, n := range nodes {
		if n == nil || n.manager != dd || Cudd_IsComplement(n) || (n != dd.One && n != dd.Zero && n.family != 1) {
			return false
		}
	}
	return true
}

// Cudd_ZddNode returns the canonical ZDD node for an index and children.
func Cudd_ZddNode(dd *DdManager, index int, t, e *DdNode) *DdNode {
	if !dd.validZdd(t, e) || index < 0 || index >= len(dd.PermZ) {
		return nil
	}
	level := func(n *DdNode) int {
		if n == dd.One || n == dd.Zero {
			return len(dd.InvPermZ)
		}
		if int(n.Index) >= len(dd.PermZ) {
			return len(dd.InvPermZ)
		}
		return dd.PermZ[n.Index]
	}
	if dd.PermZ[index] >= level(t) || dd.PermZ[index] >= level(e) {
		return nil
	}
	return dd.zddNode(uint32(index), t, e)
}

func (dd *DdManager) ensureZddVar(index int) bool {
	if !dd.validZdd() || index < 0 || uint64(index) >= uint64(CUDD_MAXINDEX-1) {
		return false
	}
	if index < dd.ZddSize {
		return true
	}
	for i := dd.ZddSize; i <= index; i++ {
		dd.PermZ = append(dd.PermZ, i)
		dd.InvPermZ = append(dd.InvPermZ, i)
	}
	dd.ZddSize = index + 1
	cuddZddInitUniv(dd)
	return true
}
func (dd *DdManager) zddLevel(n *DdNode) int {
	if Cudd_IsConstant(n) {
		return math.MaxInt
	}
	return dd.PermZ[n.Index]
}
func (dd *DdManager) zddCofactors(n *DdNode, level int) (*DdNode, *DdNode) {
	if dd.zddLevel(n) == level {
		return n.T, n.E
	}
	return dd.Zero, n
}
func Cudd_ReadZddSize(dd *DdManager) int { return dd.ZddSize }
func Cudd_ReadZddOne(dd *DdManager, level int) *DdNode {
	if !dd.validZdd() || level < 0 {
		return nil
	}
	if level >= dd.ZddSize {
		return dd.One
	}
	return dd.Univ[level]
}
func Cudd_ReadZero(dd *DdManager) *DdNode { return dd.Zero }
func Cudd_ReadPermZdd(dd *DdManager, index int) int {
	if dd == nil || index < 0 || index >= len(dd.PermZ) {
		return -1
	}
	return dd.PermZ[index]
}
func Cudd_ReadInvPermZdd(dd *DdManager, level int) int {
	if dd == nil || level < 0 || level >= len(dd.InvPermZ) {
		return -1
	}
	return dd.InvPermZ[level]
}
func Cudd_zddIthVar(dd *DdManager, index int) *DdNode {
	if !dd.ensureZddVar(index) {
		return nil
	}
	level := dd.PermZ[index]
	lower := dd.One
	if level+1 < dd.ZddSize {
		lower = dd.Univ[level+1]
	}
	n := dd.zddNode(uint32(index), lower, dd.Zero)
	for i := level - 1; i >= 0; i-- {
		n = dd.zddNode(uint32(dd.InvPermZ[i]), n, n)
	}
	return n
}
func (dd *DdManager) zddBinary(kind byte, f, g *DdNode) *DdNode {
	memo := map[[2]*DdNode]*DdNode{}
	var visit func(*DdNode, *DdNode) *DdNode
	visit = func(f, g *DdNode) *DdNode {
		switch kind {
		case '|':
			if f == dd.Zero {
				return g
			}
			if g == dd.Zero || f == g {
				return f
			}
		case '&':
			if f == dd.Zero || g == dd.Zero {
				return dd.Zero
			}
			if f == g {
				return f
			}
		case '-':
			if f == dd.Zero || f == g {
				return dd.Zero
			}
			if g == dd.Zero {
				return f
			}
		}
		key := [2]*DdNode{f, g}
		if n := memo[key]; n != nil {
			return n
		}
		level := min(dd.zddLevel(f), dd.zddLevel(g))
		ft, fe := dd.zddCofactors(f, level)
		gt, ge := dd.zddCofactors(g, level)
		n := dd.zddNode(uint32(dd.InvPermZ[level]), visit(ft, gt), visit(fe, ge))
		memo[key] = n
		return n
	}
	return visit(f, g)
}
func Cudd_zddUnion(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.validZdd(f, g) {
		return nil
	}
	return dd.zddBinary('|', f, g)
}
func Cudd_zddIntersect(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.validZdd(f, g) {
		return nil
	}
	return dd.zddBinary('&', f, g)
}
func Cudd_zddDiff(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.validZdd(f, g) {
		return nil
	}
	return dd.zddBinary('-', f, g)
}
func Cudd_zddIte(dd *DdManager, f, g, h *DdNode) *DdNode {
	if !dd.validZdd(f, g, h) {
		return nil
	}
	if g == h {
		return g
	}
	return Cudd_zddUnion(dd, Cudd_zddIntersect(dd, f, g), Cudd_zddDiff(dd, h, f))
}

// DiffConst is an inclusion query and does not allocate unique-table nodes.
func Cudd_zddDiffConst(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.validZdd(f, g) {
		return nil
	}
	memo := map[[2]*DdNode]*DdNode{}
	var visit func(*DdNode, *DdNode) *DdNode
	visit = func(f, g *DdNode) *DdNode {
		if f == dd.Zero || f == g {
			return dd.Zero
		}
		if g == dd.Zero {
			return f
		}
		key := [2]*DdNode{f, g}
		if n := memo[key]; n != nil {
			return n
		}
		lf, lg := dd.zddLevel(f), dd.zddLevel(g)
		var n *DdNode
		if lf < lg {
			n = DD_NON_CONSTANT
		} else if lf > lg {
			n = visit(f, g.E)
		} else {
			if visit(f.T, g.T) != dd.Zero {
				n = DD_NON_CONSTANT
			} else {
				n = visit(f.E, g.E)
			}
		}
		memo[key] = n
		return n
	}
	return visit(f, g)
}
func (dd *DdManager) zddSubset(f *DdNode, index int, positive bool) *DdNode {
	if !dd.validZdd(f) || !dd.ensureZddVar(index) {
		return nil
	}
	target := dd.PermZ[index]
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(f *DdNode) *DdNode {
		if f == dd.Zero {
			return f
		}
		level := dd.zddLevel(f)
		if level > target {
			if positive {
				return dd.Zero
			}
			return f
		}
		if level == target {
			if positive {
				return f.T
			}
			return f.E
		}
		if n := memo[f]; n != nil {
			return n
		}
		n := dd.zddNode(f.Index, visit(f.T), visit(f.E))
		memo[f] = n
		return n
	}
	return visit(f)
}
func Cudd_zddSubset1(dd *DdManager, f *DdNode, index int) *DdNode {
	return dd.zddSubset(f, index, true)
}
func Cudd_zddSubset0(dd *DdManager, f *DdNode, index int) *DdNode {
	return dd.zddSubset(f, index, false)
}
func Cudd_zddChange(dd *DdManager, f *DdNode, index int) *DdNode {
	if !dd.validZdd(f) || !dd.ensureZddVar(index) {
		return nil
	}
	target := dd.PermZ[index]
	memo := map[*DdNode]*DdNode{}
	var visit func(*DdNode) *DdNode
	visit = func(f *DdNode) *DdNode {
		if f == dd.Zero {
			return f
		}
		level := dd.zddLevel(f)
		if level > target {
			return dd.zddNode(uint32(index), f, dd.Zero)
		}
		if level == target {
			return dd.zddNode(uint32(index), f.E, f.T)
		}
		if n := memo[f]; n != nil {
			return n
		}
		n := dd.zddNode(f.Index, visit(f.T), visit(f.E))
		memo[f] = n
		return n
	}
	return visit(f)
}

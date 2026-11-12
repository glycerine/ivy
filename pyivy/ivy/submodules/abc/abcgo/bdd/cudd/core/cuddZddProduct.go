/**CFile***********************************************************************

  FileName    [cuddZddFuncs.c]

  PackageName [cudd]

  Synopsis    [Functions to manipulate covers represented as ZDDs.]

  Description [External procedures included in this module:
                    <ul>
                    <li> Cudd_zddProduct();
                    <li> Cudd_zddUnateProduct();
                    <li> Cudd_zddWeakDiv();
                    <li> Cudd_zddWeakDivF();
                    <li> Cudd_zddDivide();
                    <li> Cudd_zddDivideF();
                    <li> Cudd_zddComplement();
                    </ul>
               Internal procedures included in this module:
                    <ul>
                    <li> cuddZddProduct();
                    <li> cuddZddUnateProduct();
                    <li> cuddZddWeakDiv();
                    <li> cuddZddWeakDivF();
                    <li> cuddZddDivide();
                    <li> cuddZddDivideF();
                    <li> cuddZddGetCofactors3()
                    <li> cuddZddGetCofactors2()
                    <li> cuddZddComplement();
                    <li> cuddZddGetPosVarIndex();
                    <li> cuddZddGetNegVarIndex();
                    <li> cuddZddGetPosVarLevel();
                    <li> cuddZddGetNegVarLevel();
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

// pairedCover validates the original cover-domain contract: even/odd literal
// indices occur in adjacent positive-first level pairs and no cube contains
// both polarities. Unate APIs do not impose these paired-literal constraints.
func (dd *DdManager) pairedCover(roots ...*DdNode) bool {
	if !dd.validZdd(roots...) {
		return false
	}
	seen := map[*DdNode]bool{}
	indices := map[int]bool{}
	highest := -1
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		if Cudd_IsConstant(n) || seen[n] {
			return
		}
		seen[n] = true
		indices[int(n.Index)] = true
		highest = max(highest, int(n.Index)|1)
		visit(n.T)
		visit(n.E)
	}
	for _, root := range roots {
		visit(root)
	}
	if highest >= 0 && !dd.ensureZddVar(highest) {
		return false
	}
	for index := range indices {
		positive := index &^ 1
		level := dd.PermZ[positive]
		if level%2 != 0 || dd.PermZ[positive+1] != level+1 {
			return false
		}
	}
	for n := range seen {
		if n.Index%2 == 0 && !Cudd_IsConstant(n.T) && n.T.Index == n.Index+1 {
			return false
		}
	}
	return true
}

// coverCofactors returns the positive, negative and absent-variable parts,
// removing both literal indices from each resulting cube.
func (dd *DdManager) coverCofactors(f *DdNode, index int) (*DdNode, *DdNode, *DdNode) {
	positive, negative := index&^1, index|1
	p := Cudd_zddSubset0(dd, Cudd_zddSubset1(dd, f, positive), negative)
	rest := Cudd_zddSubset0(dd, f, positive)
	n := Cudd_zddSubset1(dd, rest, negative)
	absent := Cudd_zddSubset0(dd, rest, negative)
	return p, n, absent
}
func (dd *DdManager) coverNode(index int, positive, negative, absent *DdNode) *DdNode {
	p, n := index&^1, index|1
	return dd.zddNode(uint32(p), positive, dd.zddNode(uint32(n), negative, absent))
}
func Cudd_zddUnateProduct(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.validZdd(f, g) {
		return nil
	}
	memo := map[[2]*DdNode]*DdNode{}
	var product func(*DdNode, *DdNode) *DdNode
	product = func(f, g *DdNode) *DdNode {
		if f == dd.Zero || g == dd.Zero {
			return dd.Zero
		}
		if f == dd.One {
			return g
		}
		if g == dd.One {
			return f
		}
		key := [2]*DdNode{f, g}
		if n := memo[key]; n != nil {
			return n
		}
		level := min(dd.zddLevel(f), dd.zddLevel(g))
		ft, fe := dd.zddCofactors(f, level)
		gt, ge := dd.zddCofactors(g, level)
		both := Cudd_zddUnion(dd, product(ft, gt), product(ft, ge))
		positive := Cudd_zddUnion(dd, both, product(fe, gt))
		negative := product(fe, ge)
		n := dd.zddNode(uint32(dd.InvPermZ[level]), positive, negative)
		memo[key] = n
		return n
	}
	return product(f, g)
}
func Cudd_zddProduct(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.pairedCover(f, g) {
		return nil
	}
	memo := map[[2]*DdNode]*DdNode{}
	var product func(*DdNode, *DdNode) *DdNode
	product = func(f, g *DdNode) *DdNode {
		if f == dd.Zero || g == dd.Zero {
			return dd.Zero
		}
		if f == dd.One {
			return g
		}
		if g == dd.One {
			return f
		}
		key := [2]*DdNode{f, g}
		if n := memo[key]; n != nil {
			return n
		}
		variable := int(f.Index)
		if dd.zddLevel(g) < dd.zddLevel(f) {
			variable = int(g.Index)
		}
		fp, fn, fd := dd.coverCofactors(f, variable)
		gp, gn, gd := dd.coverCofactors(g, variable)
		absent := product(fd, gd)
		positive := Cudd_zddUnion(dd, Cudd_zddUnion(dd, product(fp, gp), product(fp, gd)), product(fd, gp))
		negative := Cudd_zddUnion(dd, Cudd_zddUnion(dd, product(fn, gn), product(fn, gd)), product(fd, gn))
		n := dd.coverNode(variable, positive, negative, absent)
		memo[key] = n
		return n
	}
	return product(f, g)
}
func Cudd_zddDivide(dd *DdManager, f, g *DdNode) *DdNode {
	if !dd.validZdd(f, g) || g == dd.Zero {
		return nil
	}
	memo := map[[2]*DdNode]*DdNode{}
	var divide func(*DdNode, *DdNode) *DdNode
	divide = func(f, g *DdNode) *DdNode {
		if g == dd.One {
			return f
		}
		if f == dd.Zero || f == dd.One {
			return dd.Zero
		}
		if f == g {
			return dd.One
		}
		key := [2]*DdNode{f, g}
		if n := memo[key]; n != nil {
			return n
		}
		variable := int(g.Index)
		fp, fn := Cudd_zddSubset1(dd, f, variable), Cudd_zddSubset0(dd, f, variable)
		gp, gn := g.T, g.E
		n := divide(fp, gp)
		if n != dd.Zero && gn != dd.Zero {
			n = Cudd_zddIntersect(dd, n, divide(fn, gn))
		}
		memo[key] = n
		return n
	}
	return divide(f, g)
}

// DivideF's original recursive body is identical to Divide except its cache tag.
// Sharing the engine preserves that behavior without a second implementation.
func Cudd_zddDivideF(dd *DdManager, f, g *DdNode) *DdNode { return Cudd_zddDivide(dd, f, g) }
func (dd *DdManager) weakDivide(f, g *DdNode, modified bool) *DdNode {
	if !dd.pairedCover(f, g) || g == dd.Zero {
		return nil
	}
	memo := map[[2]*DdNode]*DdNode{}
	var divide func(*DdNode, *DdNode) *DdNode
	divide = func(f, g *DdNode) *DdNode {
		if g == dd.One {
			return f
		}
		if f == dd.Zero || f == dd.One {
			return dd.Zero
		}
		if f == g {
			return dd.One
		}
		key := [2]*DdNode{f, g}
		if n := memo[key]; n != nil {
			return n
		}
		lf, lg := dd.zddLevel(f), dd.zddLevel(g)
		// The F variant preserves earlier dividend-variable branches separately;
		// the ordinary weak division instead starts at the divisor's top pair.
		if modified && lf/2 < lg/2 {
			fp, fn, fd := dd.coverCofactors(f, int(f.Index))
			n := dd.coverNode(int(f.Index), divide(fp, g), divide(fn, g), divide(fd, g))
			memo[key] = n
			return n
		}
		variable := int(g.Index)
		if modified && lf < lg {
			variable = int(f.Index)
		}
		fp, fn, fd := dd.coverCofactors(f, variable)
		gp, gn, gd := dd.coverCofactors(g, variable)
		var n *DdNode
		for _, branch := range [][2]*DdNode{{fn, gn}, {fp, gp}, {fd, gd}} {
			if branch[1] == dd.Zero {
				continue
			}
			q := divide(branch[0], branch[1])
			if n == nil {
				n = q
			} else {
				n = Cudd_zddIntersect(dd, n, q)
			}
			if n == dd.Zero {
				break
			}
		}
		memo[key] = n
		return n
	}
	return divide(f, g)
}
func Cudd_zddWeakDiv(dd *DdManager, f, g *DdNode) *DdNode  { return dd.weakDivide(f, g, false) }
func Cudd_zddWeakDivF(dd *DdManager, f, g *DdNode) *DdNode { return dd.weakDivide(f, g, true) }

// ZDDCoverCofactors exposes the original three-way cover decomposition for
// subsequent original-module ports, with explicit failure instead of raw output
// pointers. The absent part contains neither literal of the selected pair.
func ZDDCoverCofactors(dd *DdManager, f *DdNode, index int) (positive, negative, absent *DdNode, ok bool) {
	if index < 0 || uint64(index|1) >= uint64(CUDD_MAXINDEX-1) || !dd.pairedCover(f) || !dd.ensureZddVar(index|1) {
		return nil, nil, nil, false
	}
	level := dd.PermZ[index&^1]
	if level%2 != 0 || dd.PermZ[index|1] != level+1 {
		return nil, nil, nil, false
	}
	positive, negative, absent = dd.coverCofactors(f, index)
	return positive, negative, absent, true
}
func ZDDCofactors(dd *DdManager, f *DdNode, index int) (positive, negative *DdNode, ok bool) {
	if !dd.validZdd(f) {
		return nil, nil, false
	}
	positive = Cudd_zddSubset1(dd, f, index)
	if positive == nil {
		return nil, nil, false
	}
	negative = Cudd_zddSubset0(dd, f, index)
	return positive, negative, negative != nil
}

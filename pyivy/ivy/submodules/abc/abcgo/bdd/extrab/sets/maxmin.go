/**CFile***********************************************************************

  FileName    [zMaxMin.c]

  PackageName [extra]

  Synopsis    [Experimental version of some ZDD operators.]

  Description [External procedures included in this module:
				<ul>
				<li> Extra_zddMaximal();
				<li> Extra_zddMinimal();
				<li> Extra_zddMaxUnion();
				<li> Extra_zddMinUnion();
				<li> Extra_zddDotProduct();
				<li> Extra_zddCrossProduct();
				<li> Extra_zddMaxDotProduct();
				</ul>
			Internal procedures included in this module:
				<ul>
				<li> extraZddMaximal();
				<li> extraZddMinimal();
				<li> extraZddMaxUnion();
				<li> extraZddMinUnion();
				<li> extraZddDotProduct();
				<li> extraZddCrossProduct();
				<li> extraZddMaxDotProduct();
				</ul>
			StaTc procedures included in this module:
				<ul>
				</ul>

          DotProduct and MaxDotProduct were introduced
		  by O.Coudert to solve problems arising in two-level planar routing
          See O. Coudert, C.-J. R. Shi. Exact Multi-Layer Topological Planar
		  Routing. Proc. of IEEE Custom Integrated Circuit Conference '96,
		  pp. 179-182.
	      ]

  SeeAlso     []

  Author      [Alan Mishchenko]

  Copyright   []

  ReviSon    [$zMaxMin.c, v.1.2, November 26, 2000, alanmi $]

******************************************************************************/

package sets

import cudd "github.com/glycerine/abcgo/bdd/cudd/core"

type binaryOperation uint8

const (
	maxUnion binaryOperation = iota
	minUnion
	dotProduct
	crossProduct
	maxDotProduct
)

type binaryKey struct {
	kind binaryOperation
	x, y *DdNode
}

func (e *engine) minimal(f *DdNode) *DdNode {
	if f == e.dd.Zero {
		return f
	}
	if emptyBelongs(e.dd, f) {
		return e.dd.One
	}
	if r := e.minimalMemo[f]; r != nil {
		return r
	}
	without, with := e.minimal(f.E), e.minimal(f.T)
	with = e.filter(notSuperset, with, without)
	r := e.compose(f.Index, with, without)
	if e.minimalMemo == nil {
		e.minimalMemo = make(map[*DdNode]*DdNode)
	}
	e.minimalMemo[f] = r
	return r
}
func (e *engine) binary(kind binaryOperation, x, y *DdNode) *DdNode {
	dd := e.dd
	switch kind {
	case maxUnion, minUnion:
		if x == dd.Zero {
			return y
		}
		if y == dd.Zero || x == y {
			return x
		}
		if kind == maxUnion {
			if x == dd.One {
				return y
			}
			if y == dd.One {
				return x
			}
		} else if emptyBelongs(dd, x) || emptyBelongs(dd, y) {
			return dd.One
		}
	default:
		if x == dd.Zero || y == dd.Zero {
			return dd.Zero
		}
		if kind == crossProduct {
			if x == dd.One || y == dd.One {
				return dd.One
			}
		} else {
			if x == dd.One {
				return y
			}
			if y == dd.One {
				return x
			}
		}
	}
	lx, ly := dd.PermZ[x.Index], dd.PermZ[y.Index]
	// Source pointer ordering serves only cache normalization; stable IDs avoid
	// unsafe address arithmetic and preserve commutativity in Go.
	if lx > ly || lx == ly && x.ID > y.ID {
		x, y = y, x
		lx, ly = ly, lx
	}
	key := binaryKey{kind, x, y}
	if r := e.binaryMemo[key]; r != nil {
		return r
	}
	var with, without, result *DdNode
	switch kind {
	case maxUnion, minUnion:
		if lx == ly {
			without = e.binary(kind, x.E, y.E)
			with = e.binary(kind, x.T, y.T)
		} else {
			without = e.binary(kind, x.E, y)
			with = x.T
		}
		if kind == maxUnion {
			without = e.filter(notSubset, without, with)
		} else {
			with = e.filter(notSuperset, with, without)
		}
		result = e.compose(x.Index, with, without)
	case maxDotProduct:
		if lx == ly {
			merged := e.binary(maxUnion, y.E, y.T)
			a := e.binary(kind, x.T, merged)
			b := e.binary(kind, x.E, y.T)
			with = e.binary(maxUnion, a, b)
			without = e.binary(kind, x.E, y.E)
		} else {
			with = e.binary(kind, x.T, y)
			without = e.binary(kind, x.E, y)
		}
		without = e.filter(notSubset, without, with)
		result = e.compose(x.Index, with, without)
	case crossProduct:
		if lx == ly {
			merged := cudd.Cudd_zddUnion(dd, y.E, y.T)
			without = cudd.Cudd_zddUnion(dd, e.binary(kind, x.E, merged), e.binary(kind, x.T, y.E))
			with = e.binary(kind, x.T, y.T)
			result = e.compose(x.Index, with, without)
		} else {
			result = cudd.Cudd_zddUnion(dd, e.binary(kind, x.E, y), e.binary(kind, x.T, y))
		}
	}
	if e.binaryMemo == nil {
		e.binaryMemo = make(map[binaryKey]*DdNode)
	}
	e.binaryMemo[key] = result
	return result
}

// Extra_zddMinimal selects members that contain no distinct member of f.
func Extra_zddMinimal(dd *DdManager, f *DdNode) *DdNode {
	if !valid(dd, f) {
		return nil
	}
	return newEngine(dd).minimal(f)
}
func maxMinApply(dd *DdManager, x, y *DdNode, kind binaryOperation) *DdNode {
	if !valid(dd, x, y) {
		return nil
	}
	if kind == maxDotProduct && (x == dd.Zero || y == dd.Zero) {
		return dd.Zero
	}
	if kind == dotProduct {
		return cudd.Cudd_zddUnateProduct(dd, x, y)
	}
	e := newEngine(dd)
	// The fused source recursions require antichains but its public wrappers do
	// not enforce this. Normalize here so the documented public semantics also
	// hold for general families, including identity/empty-operand shortcuts.
	switch kind {
	case maxUnion, maxDotProduct:
		x, y = e.maximal(x), e.maximal(y)
	case minUnion:
		x, y = e.minimal(x), e.minimal(y)
	}
	return e.binary(kind, x, y)
}

// Extra_zddMaxUnion returns the maximal members of the union.
func Extra_zddMaxUnion(dd *DdManager, x, y *DdNode) *DdNode { return maxMinApply(dd, x, y, maxUnion) }

// Extra_zddMinUnion returns the minimal members of the union.
func Extra_zddMinUnion(dd *DdManager, x, y *DdNode) *DdNode { return maxMinApply(dd, x, y, minUnion) }

// Extra_zddDotProduct forms all pairwise unions of member sets.
func Extra_zddDotProduct(dd *DdManager, x, y *DdNode) *DdNode {
	return maxMinApply(dd, x, y, dotProduct)
}

// Extra_zddCrossProduct forms all pairwise intersections of member sets.
func Extra_zddCrossProduct(dd *DdManager, x, y *DdNode) *DdNode {
	return maxMinApply(dd, x, y, crossProduct)
}

// Extra_zddMaxDotProduct forms maximal pairwise unions using the fused recursion.
func Extra_zddMaxDotProduct(dd *DdManager, x, y *DdNode) *DdNode {
	return maxMinApply(dd, x, y, maxDotProduct)
}

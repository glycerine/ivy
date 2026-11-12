/**CFile***********************************************************************

  FileName    [zSubSet.c]

  PackageName [extra]

  Synopsis    [Experimental version of some ZDD operators.]

  Description [External procedures included in this module:
		    <ul>
		    <li> Extra_zddSubSet();
		    <li> Extra_zddSupSet();
		    <li> Extra_zddNotSubSet();
		    <li> Extra_zddNotSupSet();
		    <li> Extra_zddMaxNotSupSet();
			<li> Extra_zddEmptyBelongs();
			<li> Extra_zddIsOneSubset();
		    </ul>
	       Internal procedures included in this module:
		    <ul>
		    <li> extraZddSubSet();
		    <li> extraZddSupSet();
		    <li> extraZddNotSubSet();
		    <li> extraZddNotSupSet();
		    <li> extraZddMaxNotSupSet();
		    </ul>
	       Static procedures included in this module:
		    <ul>
		    </ul>

          SubSet, SupSet, NotSubSet, NotSupSet were introduced
		  by O.Coudert to solve problems arising in two-level SOP
		  minimization. See O. Coudert, "Two-Level Logic Minimization:
		  An Overview", Integration. Vol. 17, No. 2, pp. 97-140, Oct 1994.
	      ]

  SeeAlso     []

  Author      [Alan Mishchenko]

  Copyright   []

  Revision    [$zSubSet.c, v.1.2, November 16, 2000, alanmi $]

******************************************************************************/

package sets

import cudd "github.com/glycerine/abcgo/bdd/cudd/core"

type DdManager = cudd.DdManager
type DdNode = cudd.DdNode

type relation uint8

const (
	subset relation = iota
	superset
	notSubset
	notSuperset
	maxNotSuperset
)

type relationKey struct {
	kind relation
	x, y *DdNode
}
type engine struct {
	dd          *DdManager
	memo        map[relationKey]*DdNode
	maximalMemo map[*DdNode]*DdNode
}

func valid(dd *DdManager, nodes ...*DdNode) bool {
	if dd == nil {
		return false
	}
	for _, n := range nodes {
		if n == nil || cudd.Cudd_zddUnion(dd, n, dd.Zero) == nil {
			return false
		}
	}
	return true
}
func newEngine(dd *DdManager) *engine {
	return &engine{dd, make(map[relationKey]*DdNode), make(map[*DdNode]*DdNode)}
}
func emptyBelongs(dd *DdManager, f *DdNode) bool {
	for !cudd.Cudd_IsConstant(f) {
		f = f.E
	}
	return f == dd.One
}

// Extra_zddEmptyBelongs reports whether the family contains the empty set.
func Extra_zddEmptyBelongs(dd *DdManager, f *DdNode) bool { return valid(dd, f) && emptyBelongs(dd, f) }

// Extra_zddIsOneSubset follows the actual source: exactly one subset is true;
// the empty family is false (despite the original synopsis saying otherwise).
func Extra_zddIsOneSubset(dd *DdManager, f *DdNode) bool {
	if !valid(dd, f) {
		return false
	}
	for !cudd.Cudd_IsConstant(f) {
		if f.E != dd.Zero {
			return false
		}
		f = f.T
	}
	return f == dd.One
}
func (e *engine) compose(index uint32, with, without *DdNode) *DdNode {
	return cudd.Cudd_zddUnion(e.dd, without, cudd.Cudd_zddChange(e.dd, with, int(index)))
}
func (e *engine) maximal(f *DdNode) *DdNode {
	if cudd.Cudd_IsConstant(f) {
		return f
	}
	if r := e.maximalMemo[f]; r != nil {
		return r
	}
	without, with := e.maximal(f.E), e.maximal(f.T)
	without = e.filter(notSubset, without, with)
	r := e.compose(f.Index, with, without)
	e.maximalMemo[f] = r
	return r
}
func (e *engine) filter(kind relation, x, y *DdNode) *DdNode {
	dd := e.dd
	positive := kind == subset || kind == superset
	sub := kind == subset || kind == notSubset
	if x == y {
		if positive {
			return x
		}
		return dd.Zero
	}
	if y == dd.Zero {
		if positive {
			return dd.Zero
		}
		if kind == maxNotSuperset {
			return e.maximal(x)
		}
		return x
	}
	if x == dd.Zero {
		return dd.Zero
	}
	if sub {
		if x == dd.One {
			if positive {
				return dd.One
			}
			return dd.Zero
		}
		if y == dd.One {
			if positive {
				if emptyBelongs(dd, x) {
					return dd.One
				}
				return dd.Zero
			}
			return cudd.Cudd_zddDiff(dd, x, dd.One)
		}
	} else {
		if emptyBelongs(dd, y) {
			if positive {
				return x
			}
			return dd.Zero
		}
		if x == dd.One {
			if positive {
				return dd.Zero
			}
			return dd.One
		}
	}
	key := relationKey{kind, x, y}
	if r := e.memo[key]; r != nil {
		return r
	}
	lx, ly := dd.PermZ[x.Index], dd.PermZ[y.Index]
	var result *DdNode
	if sub {
		switch {
		case lx < ly:
			without := e.filter(kind, x.E, y)
			if positive {
				result = without
			} else {
				result = e.compose(x.Index, x.T, without)
			}
		case lx == ly:
			merged := cudd.Cudd_zddUnion(dd, y.E, y.T)
			without := e.filter(kind, x.E, merged)
			with := e.filter(kind, x.T, y.T)
			result = e.compose(x.Index, with, without)
		default:
			result = e.filter(kind, x, cudd.Cudd_zddUnion(dd, y.E, y.T))
		}
	} else {
		if lx > ly {
			result = e.filter(kind, x, y.E)
		} else {
			y0, y1 := y, y
			if lx == ly {
				y0 = y.E
				y1 = cudd.Cudd_zddUnion(dd, y.E, y.T)
			}
			without := e.filter(kind, x.E, y0)
			with := e.filter(kind, x.T, y1)
			if kind == maxNotSuperset {
				without = e.filter(notSubset, without, with)
			}
			result = e.compose(x.Index, with, without)
		}
	}
	e.memo[key] = result
	return result
}
func apply(dd *DdManager, x, y *DdNode, kind relation) *DdNode {
	if !valid(dd, x, y) {
		return nil
	}
	return newEngine(dd).filter(kind, x, y)
}

// Extra_zddSubSet selects members of x contained in some member of y.
func Extra_zddSubSet(dd *DdManager, x, y *DdNode) *DdNode { return apply(dd, x, y, subset) }

// Extra_zddSupSet selects members of x containing some member of y.
func Extra_zddSupSet(dd *DdManager, x, y *DdNode) *DdNode    { return apply(dd, x, y, superset) }
func Extra_zddNotSubSet(dd *DdManager, x, y *DdNode) *DdNode { return apply(dd, x, y, notSubset) }
func Extra_zddNotSupSet(dd *DdManager, x, y *DdNode) *DdNode { return apply(dd, x, y, notSuperset) }

// Extra_zddMaxNotSupSet fuses not-superset filtering with maximal-family extraction.
func Extra_zddMaxNotSupSet(dd *DdManager, x, y *DdNode) *DdNode {
	return apply(dd, x, y, maxNotSuperset)
}

// Extra_zddMaximal is the source MaxMin helper used by the fused set filter.
func Extra_zddMaximal(dd *DdManager, f *DdNode) *DdNode {
	if !valid(dd, f) {
		return nil
	}
	return newEngine(dd).maximal(f)
}
